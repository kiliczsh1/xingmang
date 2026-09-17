package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

// BuildRequestBody constructs the request body for AI API
func BuildRequestBody(config models.APIModelWithProvider, messages []models.ChatMessage, stream bool) map[string]interface{} {
	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	maxTokens := config.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2000
	}
	if maxTokens > 8192 {
		maxTokens = 8192
	}

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
		return map[string]interface{}{
			"model":      config.Model,
			"max_tokens": maxTokens,
			"system":     systemContent,
			"messages":   userMessages,
			"stream":     stream,
		}
	}

	// OpenAI compatible format
	return map[string]interface{}{
		"model":             config.Model,
		"messages":          messages,
		"temperature":       config.Temperature,
		"max_tokens":        maxTokens,
		"top_p":             config.TopP,
		"frequency_penalty": config.FrequencyPenalty,
		"stream":            stream,
	}
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
