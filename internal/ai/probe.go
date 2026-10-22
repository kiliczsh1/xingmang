package ai

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ErrProbeUnsupported 表示该协议的上限字段位置由用户自定义规则决定，
// 我们没有可靠办法构造"最小探测请求"。
//
// 例如 custom_json：请求体是用户在 request_rules 里描述的，连哪个键是
// 输出上限都不知道，硬试只会得到一堆无关的报错。诚实标"不支持"，不猜。
var ErrProbeUnsupported = errors.New("该协议不支持自动探测")

// ProbeFamily 把 api_format 归一到四种探测形态。
//
// 这一步只按**协议**分类，不看渠道/模型名 —— 这是整个探测链路保持通用的前提。
// gemini / anthropic_compat 走的是 OpenAI 兼容端点（见 BuildApiUrl），归入 openai。
func ProbeFamily(apiFormat string) string {
	switch strings.ToLower(strings.TrimSpace(apiFormat)) {
	case "anthropic", "claude":
		return "anthropic"
	case "responses":
		return "responses"
	case "custom_json":
		return "custom"
	default:
		return "openai"
	}
}

// BuildProbeBody 构造**最小化**探测请求体。
//
// 只带四样东西：model、一条消息、上限字段、stream=false。
//
// 为什么必须最小化：生产请求体会带上 temperature / top_p / stop / response_format，
// 而不少模型（推理类、o 系列、部分代理）会因为这些参数直接 400。那些报错里往往
// 也出现 "max_tokens" 字样，一旦被当成"长度超限"，就会一路降档到最低然后报错 ——
// 一个与模型能力无关的参数问题，被误诊成模型上限问题。
func BuildProbeBody(cfg models.APIModelWithProvider, prompt string, maxTokens int) (map[string]interface{}, string, error) {
	family := ProbeFamily(cfg.APIFormat)
	messages := []map[string]string{{"role": "user", "content": prompt}}

	switch family {
	case "custom":
		return nil, family, ErrProbeUnsupported

	case "anthropic":
		// Anthropic 的 max_tokens 是必填；system 可省。
		body := map[string]interface{}{
			"model":    cfg.Model,
			"messages": messages,
			"stream":   false,
		}
		if maxTokens > 0 {
			body["max_tokens"] = maxTokens
		}
		return body, family, nil

	case "responses":
		body := map[string]interface{}{
			"model":  cfg.Model,
			"input":  prompt,
			"stream": false,
		}
		if maxTokens > 0 {
			body["max_output_tokens"] = maxTokens
		}
		return body, family, nil

	default: // openai 兼容
		body := map[string]interface{}{
			"model":    cfg.Model,
			"messages": messages,
			"stream":   false,
		}
		if maxTokens > 0 {
			body["max_tokens"] = maxTokens
		}
		return body, family, nil
	}
}

// ProbeObservation 是一次响应里能拿到的全部事实。
//
// 判定层只认这里的字段：正文长度不算数（CJK 的字符/token 比浮动、推理模型的
// reasoning 也计入 completion），所以没 usage 就是"不可测量"。
type ProbeObservation struct {
	PromptTokens     int
	CompletionTokens int
	FinishReason     string
	HasUsage         bool
	Text             string
}

// ParseProbeObservation 按协议解析探测响应。
//
// 解析失败返回 ok=false —— 调用方必须把它当成"没有结论"，而不是"成功但为 0"。
func ParseProbeObservation(family string, raw []byte) (ProbeObservation, bool) {
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ProbeObservation{}, false
	}

	obs := ProbeObservation{}
	usage, _ := payload["usage"].(map[string]interface{})

	switch family {
	case "anthropic":
		obs.FinishReason = str(payload["stop_reason"])
		obs.Text = anthropicText(payload)
		if usage != nil {
			obs.PromptTokens = intOf(usage["input_tokens"])
			obs.CompletionTokens = intOf(usage["output_tokens"])
			obs.HasUsage = true
		}

	case "responses":
		obs.FinishReason = str(payload["status"])
		if d, ok := payload["incomplete_details"].(map[string]interface{}); ok {
			// reason = "max_output_tokens" 才代表被上限截断
			obs.FinishReason = str(d["reason"])
		}
		obs.Text = str(payload["output_text"])
		if usage != nil {
			obs.PromptTokens = firstInt(usage["input_tokens"], usage["prompt_tokens"])
			obs.CompletionTokens = firstInt(usage["output_tokens"], usage["completion_tokens"])
			obs.HasUsage = true
		}

	default: // openai 兼容
		if choices, ok := payload["choices"].([]interface{}); ok && len(choices) > 0 {
			if c, ok := choices[0].(map[string]interface{}); ok {
				obs.FinishReason = str(c["finish_reason"])
				if msg, ok := c["message"].(map[string]interface{}); ok {
					obs.Text = str(msg["content"])
				}
			}
		}
		if usage != nil {
			obs.PromptTokens = firstInt(usage["prompt_tokens"], usage["input_tokens"])
			obs.CompletionTokens = firstInt(usage["completion_tokens"], usage["output_tokens"])
			obs.HasUsage = true
		}
	}

	// 没有任何可判定字段时也不宣称成功：宁可让上层判"不可测量"。
	if obs.FinishReason == "" && !obs.HasUsage {
		return obs, false
	}
	return obs, true
}

// IsTruncated 判断本次是否**被输出上限截断**（而不是模型自己说完）。
//
// 三家的表达方式不同，但语义一致：
//   - OpenAI 兼容：finish_reason == "length"
//   - Anthropic：stop_reason == "max_tokens"
//   - Responses：incomplete_details.reason == "max_output_tokens"
func IsTruncated(family, finishReason string) bool {
	switch family {
	case "anthropic":
		return finishReason == "max_tokens"
	case "responses":
		return finishReason == "max_output_tokens"
	default:
		return finishReason == "length"
	}
}

func anthropicText(payload map[string]interface{}) string {
	blocks, ok := payload["content"].([]interface{})
	if !ok {
		return ""
	}
	var b strings.Builder
	for _, item := range blocks {
		if m, ok := item.(map[string]interface{}); ok && str(m["type"]) == "text" {
			b.WriteString(str(m["text"]))
		}
	}
	return b.String()
}

func str(v interface{}) string {
	s, _ := v.(string)
	return s
}

func intOf(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func firstInt(vals ...interface{}) int {
	for _, v := range vals {
		if n := intOf(v); n > 0 {
			return n
		}
	}
	return 0
}
