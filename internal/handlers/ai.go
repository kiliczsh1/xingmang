package handlers

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// getModelConfig gets model config with provider info
func (h *Handler) getModelConfig(configID *int) (*models.APIModelWithProvider, error) {
	var c models.APIModelWithProvider
	var query string
	var args []interface{}

	baseQuery := `SELECT m.id, m.provider_id, m.name, m.model, m.temperature, m.max_tokens, m.top_p, m.frequency_penalty, m.description, m.is_default, m.enabled, m.sort_order, m.category, m.created_at, m.updated_at, p.api_key, p.api_url, p.api_format, p.use_full_url, p.name FROM api_models m LEFT JOIN api_providers p ON m.provider_id = p.id WHERE m.enabled = 1`

	if configID != nil && *configID > 0 {
		query = baseQuery + " AND m.id = ?"
		args = append(args, *configID)
	} else {
		query = baseQuery + " AND m.is_default = 1 LIMIT 1"
	}

	err := h.db.QueryRow(query, args...).Scan(
		&c.ID, &c.ProviderID, &c.Name, &c.Model, &c.Temperature, &c.MaxTokens,
		&c.TopP, &c.FrequencyPenalty, &c.Description, &c.IsDefault, &c.Enabled, &c.SortOrder, &c.Category,
		&c.CreatedAt, &c.UpdatedAt, &c.APIKey, &c.APIUrl, &c.APIFormat, &c.UseFullURL, &c.ProviderName,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("未找到 API 配置")
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// recordUsage records usage statistics
func (h *Handler) recordUsage(modelID int, modelName, providerName string) {
	today := time.Now().Format("2006-01-02")
	var exists int
	h.db.QueryRow("SELECT 1 FROM usage_stats WHERE date = ? AND model_id = ?", today, modelID).Scan(&exists)
	if exists == 1 {
		h.db.Exec("UPDATE usage_stats SET usage_count = usage_count + 1, updated_at = CURRENT_TIMESTAMP WHERE date = ? AND model_id = ?", today, modelID)
	} else {
		h.db.Exec("INSERT INTO usage_stats (date, model_id, model_name, provider_name, usage_count) VALUES (?, ?, ?, ?, 1)", today, modelID, modelName, providerName)
	}
}


// writeSSEError 以合法 JSON 推送 SSE 错误事件（错误文本可能含引号，必须走 json.Marshal 转义）
func writeSSEError(w http.ResponseWriter, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (h *Handler) AIChat(w http.ResponseWriter, r *http.Request) {
	var req models.AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	config, err := h.getModelConfig(req.ConfigID)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	// Build messages
	var chatMessages []models.ChatMessage
	var systemContents []string

	if len(req.SystemPrompts) > 0 {
		systemContents = append(systemContents, strings.Join(req.SystemPrompts, "\n\n"))
	}
	if len(req.RelatedContent) > 0 {
		var parts []string
		for _, rc := range req.RelatedContent {
			parts = append(parts, fmt.Sprintf("[%s] %s\n%s", rc.Type, rc.Title, rc.Content))
		}
		systemContents = append(systemContents, "参考内容：\n"+strings.Join(parts, "\n\n"))
	}
	if len(systemContents) > 0 {
		chatMessages = append(chatMessages, models.ChatMessage{
			Role:    "system",
			Content: strings.Join(systemContents, "\n\n---\n\n"),
		})
	}
	chatMessages = append(chatMessages, req.Messages...)

	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	useFullUrl := config.UseFullURL == 1
	apiUrl := ai.BuildApiUrl(config.APIUrl, apiFormat, useFullUrl)
	requestBody := ai.BuildRequestBody(*config, chatMessages, true)
	headers := ai.BuildAuthHeaders(*config.APIKey, apiFormat)

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Make streaming request
	bodyBytes, _ := json.Marshal(requestBody)
	client := &http.Client{Timeout: 180 * time.Second}
	httpReq, _ := http.NewRequest("POST", apiUrl, strings.NewReader(string(bodyBytes)))
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		writeSSEError(w, err.Error())
		fmt.Fprintf(w, "data: [DONE]\n\n")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		writeSSEError(w, fmt.Sprintf("API错误(%d): %s", resp.StatusCode, string(body)))
		fmt.Fprintf(w, "data: [DONE]\n\n")
		return
	}

	// Forward stream
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		message := strings.TrimPrefix(line, "data:")
		message = strings.TrimSpace(message)
		if message == "[DONE]" {
			h.recordUsage(config.ID, config.Name, config.ProviderName)
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			return
		}
		// Parse and extract content
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(message), &parsed); err == nil {
			if errMsg, ok := parsed["error"]; ok {
				errData, _ := json.Marshal(map[string]interface{}{"error": errMsg})
				fmt.Fprintf(w, "data: %s\n\n", errData)
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				continue
			}
			if choices, ok := parsed["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if delta, ok := choice["delta"].(map[string]interface{}); ok {
						if content, ok := delta["content"].(string); ok && content != "" {
							data, _ := json.Marshal(map[string]string{"content": content})
							fmt.Fprintf(w, "data: %s\n\n", data)
						}
					}
				}
			}
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		writeSSEError(w, fmt.Sprintf("流读取错误: %s", err.Error()))
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

func (h *Handler) GenerateDescription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string `json:"title"`
		PromptID *int   `json:"promptId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	promptContent := "你是一个专业的网络小说编辑，请根据书名生成一个吸引人的简介，字数控制在 100-200 字之间。"
	if req.PromptID != nil {
		var p models.Prompt
		h.db.QueryRow("SELECT content FROM prompts WHERE id = ?", *req.PromptID).Scan(&p.Content)
		if p.Content != "" {
			promptContent = p.Content
		}
	}

	config, err := h.getModelConfig(nil)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}

	apiFormat := config.APIFormat
	if apiFormat == "" {
		apiFormat = "openai"
	}
	apiUrl := ai.BuildApiUrl(config.APIUrl, apiFormat, config.UseFullURL == 1)
	messages := []models.ChatMessage{
		{Role: "system", Content: promptContent},
		{Role: "user", Content: "书名：" + req.Title},
	}
	requestBody := ai.BuildRequestBody(*config, messages, false)
	headers := ai.BuildAuthHeaders(*config.APIKey, apiFormat)

	result, err := ai.RequestJSON(apiUrl, requestBody, headers, 120*time.Second)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	h.recordUsage(config.ID, config.Name, config.ProviderName)
	respondOK(w, ai.ExtractResponseContent(result))
}

// GenerateImage handles image generation requests (placeholder).
func (h *Handler) GenerateImage(w http.ResponseWriter, r *http.Request) {
	respondFail(w, 501, "图片生成功能尚未实现")
}

// RecognizeCharacters handles character recognition requests (placeholder).
func (h *Handler) RecognizeCharacters(w http.ResponseWriter, r *http.Request) {
	respondFail(w, 501, "角色识别功能尚未实现")
}

// GetRankReferenceCategories returns available rank reference categories.
func (h *Handler) GetRankReferenceCategories(w http.ResponseWriter, r *http.Request) {
	categories := []map[string]string{
		{"key": "genre", "label": "题材类型"},
		{"key": "setting", "label": "世界观设定"},
		{"key": "power", "label": "力量体系"},
		{"key": "faction", "label": "势力阵营"},
		{"key": "role", "label": "角色定位"},
		{"key": "relationship", "label": "关系类型"},
	}
	respondOK(w, categories)
}

// GetRankReferenceBooks returns rank reference books for a given category.
func (h *Handler) GetRankReferenceBooks(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	bookIDStr := r.URL.Query().Get("bookId")

	var chapters []map[string]interface{}
	query := "SELECT id, title, content FROM chapters"
	var args []interface{}

	if bookIDStr != "" {
		query += " WHERE book_id = ?"
		args = append(args, bookIDStr)
	}
	query += " ORDER BY order_num ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondOK(w, []interface{}{})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title, content string
		if err := rows.Scan(&id, &title, &content); err != nil {
			continue
		}
		chapters = append(chapters, map[string]interface{}{
			"id":      id,
			"title":   title,
			"content": content,
		})
	}

	_ = category
	respondOK(w, chapters)
}
