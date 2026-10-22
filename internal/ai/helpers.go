package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/aitokens"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// BuildAuthHeaders builds authentication headers based on API format
func BuildAuthHeaders(apiKey string, apiFormat string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	switch apiFormat {
	case "anthropic", "anthropic_compat", "claude":
		headers["x-api-key"] = apiKey
		if apiFormat == "anthropic" || apiFormat == "claude" {
			headers["anthropic-version"] = "2023-06-01"
		}
	case "gemini":
		headers["x-goog-api-key"] = apiKey
	default:
		headers["Authorization"] = "Bearer " + apiKey
	}
	return headers
}

// derefStr 安全解引用可空字符串。
func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

// BuildAuthHeadersForModel 按「认证方式 + 协议」构造鉴权头。
//
// authType 取值（默认 api_key）：
//   - api_key    ：按 apiFormat 走默认头（Bearer / x-api-key / x-goog-api-key）
//   - endpoint_id：与 api_key 相同（endpoint 只是 URL 变量，鉴权仍是 Key）
//   - ak_sk      ：Access Key + Secret Key（固定头；火山方舟/Vertex 的完整签名
//     需各平台单独实现，由供应商适配器覆盖本函数结果）
//   - group_id   ：API Key + X-Group-Id（智谱等）
func BuildAuthHeadersForModel(apiKey *string, apiFormat, authType string, accessKey, secretKey, groupID *string) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	key := derefStr(apiKey)

	switch strings.ToLower(strings.TrimSpace(authType)) {
	case "ak_sk":
		if ak := derefStr(accessKey); ak != "" {
			headers["X-Access-Key"] = ak
		}
		if sk := derefStr(secretKey); sk != "" {
			headers["X-Secret-Key"] = sk
		}
		return headers
	case "group_id":
		if key != "" {
			headers["Authorization"] = "Bearer " + key
		}
		if g := derefStr(groupID); g != "" {
			headers["X-Group-Id"] = g
		}
		return headers
	}

	// api_key / endpoint_id / 未知：按协议走默认头
	for k, v := range BuildAuthHeaders(key, apiFormat) {
		headers[k] = v
	}
	return headers
}

// BuildAuthHeadersForProvider 是 BuildAuthHeadersForModel 的 provider 便捷入口。
func BuildAuthHeadersForProvider(p models.APIProvider) map[string]string {
	return BuildAuthHeadersForModel(p.APIKey, p.APIFormat, p.AuthType, p.AccessKey, p.SecretKey, p.GroupID)
}

/* ---------------- 自定义请求头 / 上游必需的会话头 ---------------- */

// ExtraHeadersConfig 是 extra_config 字段的约定结构。
//
// 例：{"headers":{"x-opencode-session":"my-session","X-Custom":"v"}}
type ExtraHeadersConfig struct {
	Headers map[string]string `json:"headers"`
}

// OpenCodeSessionHeaderName 是 OpenCode Go（opencode.ai/zen/go）要求的会话亲和头。
const OpenCodeSessionHeaderName = "x-opencode-session"

// ParseExtraHeaders 从 extra_config（JSON 文本）解析自定义请求头。
// 解析失败或为空时返回空 map —— extra_config 允许存放其它用途的配置，不因此报错。
func ParseExtraHeaders(extraConfig *string) map[string]string {
	out := map[string]string{}
	raw := derefStr(extraConfig)
	if raw == "" {
		return out
	}
	var cfg ExtraHeadersConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return out
	}
	for k, v := range cfg.Headers {
		if k = strings.TrimSpace(k); k != "" {
			out[k] = v
		}
	}
	return out
}

// isOpenCodeGoEndpoint 判断是否为 OpenCode Go 套餐端点。
func isOpenCodeGoEndpoint(apiURL string) bool {
	return strings.Contains(strings.ToLower(apiURL), "opencode.ai/zen/go")
}

// openCodeSessionID 生成稳定的会话 ID。
//
// OpenCode Go 自 2026-09-06 起要求 x-opencode-session，缺失的请求会被拒绝。
// 该值不透明（上游不解析），只要求「同一会话稳定、不同会话唯一」。
// B 项目是本地单用户应用，请求链路里没有对话 ID 可传，
// 因此按「渠道」维度取稳定 ID —— 与 operit（operit-<configId>）等客户端做法一致。
// 若将来链路能拿到对话 ID，应替换为真实会话标识。
func openCodeSessionID(providerID int) string {
	return fmt.Sprintf("xingmang-%d", providerID)
}

// ApplyProviderHeaders 把 extra_config 里的自定义头、以及 OpenCode Go 必需的会话头
// 合并进 headers。**已存在的键不覆盖**，因此用户在 extra_config 里显式配置的值优先。
func ApplyProviderHeaders(headers map[string]string, apiURL string, providerID int, extraConfig *string) {
	for k, v := range ParseExtraHeaders(extraConfig) {
		if _, exists := headers[k]; !exists {
			headers[k] = v
		}
	}
	if isOpenCodeGoEndpoint(apiURL) {
		if _, exists := headers[OpenCodeSessionHeaderName]; !exists {
			headers[OpenCodeSessionHeaderName] = openCodeSessionID(providerID)
		}
	}
}

// BuildApiUrl constructs the full API URL
func BuildApiUrl(baseUrl string, apiFormat string, useFullUrl bool) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseUrl), "/")
	if useFullUrl {
		return trimmed
	}
	switch apiFormat {
	case "anthropic", "claude":
		if strings.HasSuffix(trimmed, "/v1/messages") {
			return trimmed
		}
		return trimmed + "/v1/messages"
	case "responses":
		// OpenAI Responses 原生接口（/v1/responses）
		if strings.HasSuffix(trimmed, "/responses") {
			return trimmed
		}
		if strings.HasSuffix(trimmed, "/v1") {
			return trimmed + "/responses"
		}
		return trimmed + "/v1/responses"
	case "custom_json":
		// 路径完全由模型的 request_rules 指定，这里原样返回 Base URL
		return trimmed
	case "openai", "anthropic_compat", "gemini":
		if strings.HasSuffix(trimmed, "/chat/completions") {
			return trimmed
		}
		if strings.HasSuffix(trimmed, "/v1") {
			return trimmed + "/chat/completions"
		}
		return trimmed + "/v1/chat/completions"
	}
	return trimmed
}

// BuildImageApiUrl constructs the image generation endpoint (OpenAI-compatible).
func BuildImageApiUrl(baseUrl string, apiFormat string, useFullUrl bool) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseUrl), "/")
	if useFullUrl {
		return trimmed
	}
	if strings.EqualFold(strings.TrimSpace(apiFormat), "custom_json") {
		// 自定义 JSON 供应商的提交路径由 request_rules 指定
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/images/generations") {
		return trimmed
	}
	if strings.HasSuffix(trimmed, "/v1") {
		return trimmed + "/images/generations"
	}
	return trimmed + "/v1/images/generations"
}

// BuildRequestBody constructs the request body for AI API
func BuildRequestBody(config models.APIModelWithProvider, messages []models.ChatMessage, stream bool) map[string]interface{} {
	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	// max_tokens 按模型配置原样发送，**不再兜底**。
	//
	// 历史上这里先压到 8192，后来改成兜底 2000；项目里同类兜底值共有五个
	// （2000 / 4000 / 4096 / 8192），没有一个有据可依，而且偏小时会**静默截断输出**。
	//
	// 现在：值为空时**不发送这个字段** —— 让上游按模型自身能力决定，比我们编一个
	// 数字安全得多。唯一例外是 Anthropic 格式，它的 max_tokens 是必填字段，
	// 此时用 aitokens.FamilyStart 兜底（全项目唯一允许出现的无配置兜底值）。
	//
	// ⚠ 若上游不接受过大的值会返回 400：fixed 模式下调小该模型的配置即可；
	// auto 模式无需干预 —— 探测会自己降档（见 handlers/ai.go 与 agent/loop.go 的降档循环）。
	maxTokens := config.MaxTokens

	if apiFormat == "anthropic" || apiFormat == "claude" {
		var systemContent string
		var userMessages []map[string]string
		for _, m := range messages {
			if m.Role == "system" {
				systemContent = m.Content
			} else {
				role := m.Role
				if role != "assistant" {
					role = "user"
				}
				userMessages = append(userMessages, map[string]string{"role": role, "content": m.Content})
			}
		}
		if maxTokens <= 0 {
			// Anthropic 的 max_tokens 必填，漏了会被上游直接拒。
			maxTokens = aitokens.FamilyStart(&config)
		}
		return map[string]interface{}{
			"model":      config.Model,
			"max_tokens": maxTokens,
			"system":     systemContent,
			"messages":   userMessages,
			"stream":     stream,
		}
	}

	// OpenAI compatible format：max_tokens 是可选字段，为空就不带这个键。
	body := map[string]interface{}{
		"model":             config.Model,
		"messages":          messages,
		"temperature":       config.Temperature,
		"top_p":             config.TopP,
		"frequency_penalty": config.FrequencyPenalty,
		"stream":            stream,
	}
	if maxTokens > 0 {
		body["max_tokens"] = maxTokens
	}
	return body
}

// RequestJSON makes a JSON request and returns the response
func RequestJSON(url string, body interface{}, headers map[string]string, timeout time.Duration) (map[string]interface{}, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", string(respBody))
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %v", resp.StatusCode, result)
	}

	return result, nil
}

// RequestJSONGet 发起 GET 并解析 JSON 响应（供任务轮询使用）。
func RequestJSONGet(url string, headers map[string]string, timeout time.Duration) (map[string]interface{}, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", string(respBody))
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (%d): %v", resp.StatusCode, result)
	}
	return result, nil
}

// ExtractResponseContent extracts content from AI response
func ExtractResponseContent(data map[string]interface{}) string {
	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := msg["content"].(string); ok {
					return content
				}
			}
			if text, ok := choice["text"].(string); ok {
				return text
			}
		}
	}
	if content, ok := data["content"].(string); ok {
		return content
	}
	if output, ok := data["output"].(map[string]interface{}); ok {
		if text, ok := output["text"].(string); ok {
			return text
		}
	}
	if output, ok := data["output"].(string); ok {
		return output
	}
	b, _ := json.Marshal(data)
	return string(b)
}
