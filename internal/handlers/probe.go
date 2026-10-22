package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/aitokens"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 探测请求用的 max_tokens。
//
// connect 只验证"这条链路能不能通"，取一个几乎所有模型都接受的极小值；
// 它会真的生成 16 个 token，成本可以忽略，而且能顺带验证 usage 字段是否可用。
const probeConnectTokens = 16

// 探测用的提示词。
//
// cap 阶段**故意只说 "hi"**：这一步要验证的是"上游接不接受这个 max_tokens"，
// 不是"能不能生成这么长"。短输入让每次尝试只花几个 token，也让 400 更容易出现。
// 长输出校准是另一个动作，必须由用户显式触发并确认费用。
const probePrompt = "hi"

// ProbeRequest 是 POST /api/ai/probe 的入参。
//
// 支持两种定位方式：
//   - provider_id + model：导入弹窗里、模型尚未落库时使用；
//   - model_id：模型已存在时使用（行内测试），渠道/模型串以数据库为准。
type ProbeRequest struct {
	ProviderID  int    `json:"provider_id"`
	Model       string `json:"model"`
	ConfigType  string `json:"config_type"`
	Kind        string `json:"kind"` // connect | cap
	DeclaredMax int    `json:"declared_max"`
	ModelID     int    `json:"model_id"`
}

// ProbeStep 是一次尝试的记录，逐条回给前端展示"为什么是这个数"。
type ProbeStep struct {
	Tokens     int    `json:"tokens"`
	Status     int    `json:"status"`
	Kind       string `json:"kind"` // ok|limit_out|limit_ctx|quota|auth|notfound|ratelimit|upstream|network|other
	Completion int    `json:"completion_tokens"`
	Finish     string `json:"finish_reason"`
	Note       string `json:"note,omitempty"`
}

// ProbeResponse 是探测结果。
//
// ⚠ Accepted 的含义是"这个值**没有被拒**"，不是"模型能生成这么多"。
// 界面必须按这个口径显示（"可填 65536"），不能写成"已验证"。
type ProbeResponse struct {
	Model      string      `json:"model"`
	ProviderID int         `json:"provider_id"`
	Reachable  bool        `json:"reachable"`
	Accepted   int         `json:"accepted"`
	Conclusion string      `json:"conclusion"` // ok|auth|quota|notfound|ratelimit|unsupported|network|error
	Message    string      `json:"message"`
	LatencyMs  int64       `json:"latency_ms"`
	Steps      []ProbeStep `json:"steps"`
}

// ProbeModel 探测一条渠道 × 模型组合。
//
// 设计约束（与 docs/max-tokens-探测方案-v2.md 对应）：
//   - 请求体最小化，不继承生产参数；
//   - 只把"明确的长度类 4xx"当作继续降档的信号；
//   - 额度不足（预扣失败）**不降档**，否则会把"余额不够"记成"模型不支持"；
//   - 任何失败都只返回结论，不影响主流程。
func (h *Handler) ProbeModel(w http.ResponseWriter, r *http.Request) {
	var req ProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	cfg, err := h.probeTargetConfig(req)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	url := ai.BuildApiUrl(cfg.APIUrl, cfg.APIFormat, cfg.UseFullURL == 1)
	headers := ai.BuildAuthHeadersForModel(cfg.APIKey, cfg.APIFormat, cfg.AuthType, cfg.AccessKey, cfg.SecretKey, cfg.GroupID)
	ai.ApplyProviderHeaders(headers, cfg.APIUrl, cfg.ProviderID, cfg.ExtraConfig)

	resp := ProbeResponse{Model: cfg.Model, ProviderID: cfg.ProviderID, Steps: []ProbeStep{}}

	kind := strings.ToLower(strings.TrimSpace(req.Kind))
	if kind == "" {
		kind = "cap"
	}

	// 非对话模型的 max_tokens 语义不同（图像尺寸、向量维度…），不探测。
	// 用用户/推断出的 config_type 判断，而不是靠模型名白名单。
	if kind == "cap" && req.ConfigType != "" && req.ConfigType != "chat" {
		resp.Conclusion = "unsupported"
		resp.Message = "非对话模型不支持上限探测"
		respondOK(w, resp)
		return
	}

	var targets []int
	if kind == "connect" {
		targets = []int{probeConnectTokens}
	} else {
		// Fresh：用户点了按钮，就必须真的从高处重走一遍，不能被缓存按在原地。
		targets = aitokens.CandidatesFresh(cfg, 0)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	key := aitokens.CacheKey(cfg)
	start := time.Now()

	for i, mt := range targets {
		body, family, err := ai.BuildProbeBody(*cfg, probePrompt, mt)
		if err != nil {
			resp.Conclusion = "unsupported"
			resp.Message = err.Error()
			respondOK(w, resp)
			return
		}
		payload, _ := json.Marshal(body)
		httpReq, err := http.NewRequest("POST", url, bytes.NewReader(payload))
		if err != nil {
			resp.Conclusion = "error"
			resp.Message = "创建请求失败: " + err.Error()
			respondOK(w, resp)
			return
		}
		for k, v := range headers {
			httpReq.Header.Set(k, v)
		}

		httpResp, err := client.Do(httpReq)
		if err != nil {
			resp.Steps = append(resp.Steps, ProbeStep{Tokens: mt, Kind: "network", Note: err.Error()})
			resp.Conclusion = "network"
			resp.Message = "连接失败: " + err.Error()
			resp.LatencyMs = time.Since(start).Milliseconds()
			respondOK(w, resp)
			return
		}
		raw, _ := io.ReadAll(io.LimitReader(httpResp.Body, 128*1024))
		httpResp.Body.Close()

		if httpResp.StatusCode < 400 {
			obs, measurable := ai.ParseProbeObservation(family, raw)
			step := ProbeStep{
				Tokens: mt, Status: httpResp.StatusCode, Kind: "ok",
				Completion: obs.CompletionTokens, Finish: obs.FinishReason,
			}
			if !measurable {
				step.Note = "响应里没有 usage / finish_reason，无法据此判定"
			}
			resp.Steps = append(resp.Steps, step)
			resp.Reachable = true
			resp.Conclusion = "ok"
			resp.Message = "连通正常"
			if kind == "cap" && mt > 0 {
				// 只记"被接受"，不记"能生成"。真正的输出上限要靠长输出校准。
				resp.Accepted = mt
				aitokens.Remember(key, mt)
			}
			break
		}

		stepKind := probeStepKind(httpResp.StatusCode, string(raw))
		resp.Steps = append(resp.Steps, ProbeStep{
			Tokens: mt, Status: httpResp.StatusCode, Kind: stepKind,
			Note: truncateText(string(raw), 180),
		})

		// 只有长度类失败允许继续降档。额度/鉴权/参数错误重试多少次都一样。
		if (stepKind == "limit_out" || stepKind == "limit_ctx") && i < len(targets)-1 {
			log.Printf("[probe] %s max_tokens=%d 被拒(%s)，降档重试", cfg.Model, mt, stepKind)
			continue
		}

		resp.Conclusion = stepKind
		resp.Message = probeStepMessage(stepKind, httpResp.StatusCode)
		break
	}

	resp.LatencyMs = time.Since(start).Milliseconds()
	respondOK(w, resp)
}

// probeTargetConfig 按 model_id 或 provider_id+model 组装探测用的配置。
func (h *Handler) probeTargetConfig(req ProbeRequest) (*models.APIModelWithProvider, error) {
	if req.ModelID > 0 {
		m, err := scanModelWithProvider(h.db.QueryRow("SELECT "+apiModelColumnsWithProvider+`
			FROM api_models m
			JOIN api_providers p ON m.provider_id = p.id
			WHERE m.id = ?`, req.ModelID))
		if err != nil {
			return nil, fmt.Errorf("模型配置不存在")
		}
		req.ProviderID = m.ProviderID
		req.Model = m.Model
		if req.ConfigType == "" {
			req.ConfigType = m.ConfigType
		}
		// 已落库的模型：声明值直接从它自己的来源字段判断。
		if m.MaxTokensSource == "upstream" && m.MaxTokens > 0 {
			req.DeclaredMax = m.MaxTokens
		}
	}

	if req.ProviderID <= 0 || strings.TrimSpace(req.Model) == "" {
		return nil, fmt.Errorf("缺少 provider_id 或 model")
	}

	p, err := scanProvider(h.db.QueryRow("SELECT "+providerColumns+" FROM api_providers WHERE id = ?", req.ProviderID))
	if err != nil {
		return nil, fmt.Errorf("提供商不存在")
	}

	cfg := &models.APIModelWithProvider{
		APIModel: models.APIModel{
			ProviderID:    p.ID,
			Model:         strings.TrimSpace(req.Model),
			MaxTokensMode: "auto",
			ConfigType:    req.ConfigType,
		},
		APIKey:       p.APIKey,
		APIUrl:       p.APIUrl,
		APIFormat:    p.APIFormat,
		UseFullURL:   p.UseFullURL,
		ProviderName: p.Name,
		AuthType:     p.AuthType,
		AccessKey:    p.AccessKey,
		SecretKey:    p.SecretKey,
		GroupID:      p.GroupID,
		EndpointID:   p.EndpointID,
		Concurrency:  p.Concurrency,
		ChannelName:  p.ChannelName,
		ExtraConfig:  p.ExtraConfig,
	}
	// 上游 /models 声明的输出上限，作为探测起点（比按模型名猜家族更准）。
	if req.DeclaredMax > 0 {
		cfg.MaxTokens = req.DeclaredMax
		cfg.MaxTokensSource = "upstream"
	}
	return cfg, nil
}

// probeStepKind 把一次 4xx 归到可操作的类别。
//
// 与 aitokens.Classify 的分工：那边只回答"能不能降档"，这里还要回答"界面该说什么"。
// 额度不足必须单独拎出来 —— 它最常见的触发条件恰恰是**max_tokens 填得太大**，
// 误判成"模型不支持"会让用户把有能力的大值永久设小。
func probeStepKind(status int, body string) string {
	if isQuotaError(status, body) {
		return "quota"
	}
	// 中转常把"模型不存在"包成 400/500 而不是 404（实测：某渠道返回
	// 400 "Upstream request failed: Model is unavailable."）。
	// 不认出来的话，前端会把它当成普通失败继续打下一个模型，
	// 用户看到的也只是一句含糊的"测试失败"。
	if isModelMissing(body) {
		return "notfound"
	}
	switch status {
	case 401:
		return "auth"
	case 403:
		return "auth"
	case 404:
		return "notfound"
	case 413:
		return "limit_ctx"
	case 429:
		return "ratelimit"
	}
	switch aitokens.Classify(status, body) {
	case aitokens.FailureOutputLimit:
		return "limit_out"
	case aitokens.FailureContextLimit:
		return "limit_ctx"
	}
	if status >= 500 {
		return "upstream"
	}
	return "other"
}

// isModelMissing 识别"该模型在这条渠道上不可用"。
//
// 只匹配语义短语，不匹配渠道名/模型名 —— 换一家中转、换一个模型，判定依然成立。
func isModelMissing(body string) bool {
	lower := strings.ToLower(body)
	for _, kw := range []string{
		"model is unavailable", "model not found", "model_not_found",
		"no such model", "model does not exist", "model not available",
		"模型不存在", "模型不可用", "未开通该模型",
	} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// isQuotaError 识别"余额/预扣不足"，避免把它当成模型能力问题。
func isQuotaError(status int, body string) bool {
	if status != 402 && status != 403 && status != 400 {
		return false
	}
	lower := strings.ToLower(body)
	for _, kw := range []string{
		"insufficient quota", "insufficient_quota", "quota exceeded",
		"exceeded your current quota", "insufficient balance", "insufficient_balance",
		"billing hard limit", "not enough quota", "余额不足", "额度不足", "预扣",
	} {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func probeStepMessage(kind string, status int) string {
	if status == 413 {
		return "请求体过大（413）：多为网关体积限制，不代表模型输出上限"
	}
	switch kind {
	case "quota":
		return "额度不足以预扣该 max_tokens（不是模型不支持，降低取值或充值后重试）"
	case "auth":
		return fmt.Sprintf("鉴权失败（HTTP %d），检查 Key 与认证方式", status)
	case "notfound":
		return "模型不存在或该渠道未开通此模型"
	case "ratelimit":
		return "触发限流（429），稍后再试"
	case "upstream":
		return fmt.Sprintf("上游故障（HTTP %d）", status)
	case "limit_out":
		return "输出上限被拒，已降到最低档仍失败"
	case "limit_ctx":
		return "上下文超限（本次输入过长），不代表模型输出上限"
	default:
		return fmt.Sprintf("请求被拒（HTTP %d），多为参数不兼容", status)
	}
}
