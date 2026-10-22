package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetAllConfigs returns all api_models with provider info.
func (h *Handler) GetAllConfigs(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT " + apiModelColumnsWithProvider + `
		FROM api_models m
		JOIN api_providers p ON m.provider_id = p.id
		ORDER BY m.sort_order ASC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.APIModelWithProvider
	for rows.Next() {
		m, err := scanModelWithProvider(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// GetDefaultConfig returns the default enabled api_model with provider info.
func (h *Handler) GetDefaultConfig(w http.ResponseWriter, r *http.Request) {
	m, err := scanModelWithProvider(h.db.QueryRow("SELECT " + apiModelColumnsWithProvider + `
		FROM api_models m
		JOIN api_providers p ON m.provider_id = p.id
		WHERE m.is_default = 1 AND m.enabled = 1
		ORDER BY m.sort_order ASC
		LIMIT 1`))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "未找到默认配置")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, m)
}

// GetConfigsByProvider returns api_models for a specific provider.
func (h *Handler) GetConfigsByProvider(w http.ResponseWriter, r *http.Request) {
	providerID, ok := idParam(r, "providerId")
	if !ok {
		respondFail(w, 400, "无效的提供商ID")
		return
	}

	rows, err := h.db.Query("SELECT "+apiModelColumnsWithProvider+`
		FROM api_models m
		JOIN api_providers p ON m.provider_id = p.id
		WHERE m.provider_id = ?
		ORDER BY m.sort_order ASC`, providerID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.APIModelWithProvider
	for rows.Next() {
		m, err := scanModelWithProvider(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// deriveConfigTypeFromCategory 由旧 category 推 config_type（前端未显式给类型时的兜底）。
func deriveConfigTypeFromCategory(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "image", "vision", "video", "embed":
		return strings.ToLower(strings.TrimSpace(category))
	default:
		return "chat"
	}
}

// normalizeModelDefaults 补模型默认值：config_type 缺省由 category 推，其余高级参数给稳定默认。
func normalizeModelDefaults(m *models.APIModel) {
	if m.ConfigType == "" {
		m.ConfigType = deriveConfigTypeFromCategory(m.Category)
	}
	if m.PromptRole == "" {
		m.PromptRole = "system"
	}
	if m.MaxTokensMode == "" {
		m.MaxTokensMode = "auto"
	}
	if m.ThinkingEffort == "" {
		m.ThinkingEffort = "default"
	}
	if m.GenerationConfigSource == "" {
		m.GenerationConfigSource = "manual"
	}
}

// CreateConfig creates a a new api_model config.
func (h *Handler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var m models.APIModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	normalizeModelDefaults(&m)

	result, err := h.db.Exec(`INSERT INTO api_models
		(provider_id, name, model, temperature, max_tokens, top_p, frequency_penalty, description,
		 is_default, enabled, sort_order, category,
		 config_type, prompt, prompt_role, top_k, presence_penalty, stop, max_tokens_mode,
		 max_tokens_source,
		 thinking_mode, thinking_effort, generation_config_source, request_rules,
		 param_groups, default_params, is_free, input_ratio, output_ratio)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ProviderID, m.Name, m.Model, m.Temperature, m.MaxTokens, m.TopP, m.FrequencyPenalty, m.Description,
		m.IsDefault, m.Enabled, m.SortOrder, m.Category,
		m.ConfigType, m.Prompt, m.PromptRole, m.TopK, m.PresencePenalty, m.Stop, m.MaxTokensMode,
		m.MaxTokensSource,
		m.ThinkingMode, m.ThinkingEffort, m.GenerationConfigSource, m.RequestRules,
		m.ParamGroups, m.DefaultParams, m.IsFree, m.InputRatio, m.OutputRatio)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	m.ID = int(id)
	respondOK(w, m)
}

// UpdateConfig updates an existing api_model config.
func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var m models.APIModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	normalizeModelDefaults(&m)

	_, err := h.db.Exec(`UPDATE api_models SET
		provider_id=?, name=?, model=?, temperature=?, max_tokens=?, top_p=?, frequency_penalty=?, description=?,
		is_default=?, enabled=?, sort_order=?, category=?,
		config_type=?, prompt=?, prompt_role=?, top_k=?, presence_penalty=?, stop=?, max_tokens_mode=?,
		max_tokens_source=?,
		thinking_mode=?, thinking_effort=?, generation_config_source=?, request_rules=?,
		param_groups=?, default_params=?, is_free=?, input_ratio=?, output_ratio=?,
		updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		m.ProviderID, m.Name, m.Model, m.Temperature, m.MaxTokens, m.TopP, m.FrequencyPenalty, m.Description,
		m.IsDefault, m.Enabled, m.SortOrder, m.Category,
		m.ConfigType, m.Prompt, m.PromptRole, m.TopK, m.PresencePenalty, m.Stop, m.MaxTokensMode,
		m.MaxTokensSource,
		m.ThinkingMode, m.ThinkingEffort, m.GenerationConfigSource, m.RequestRules,
		m.ParamGroups, m.DefaultParams, m.IsFree, m.InputRatio, m.OutputRatio,
		id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	m.ID = id
	respondOK(w, m)
}

// DeleteConfig deletes an api_model config by id.
func (h *Handler) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM api_models WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// ToggleModel toggles the enabled field of an api_model.
func (h *Handler) ToggleModel(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("UPDATE api_models SET enabled = CASE WHEN enabled = 1 THEN 0 ELSE 1 END, updated_at=CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "切换成功")
}

// ReorderModelsRequest is the request body for reordering models.
type ReorderModelsRequest struct {
	Orders []struct {
		ID        int `json:"id"`
		SortOrder int `json:"sort_order"`
	} `json:"orders"`
}

// ReorderModels updates the sort_order for all given models.
func (h *Handler) ReorderModels(w http.ResponseWriter, r *http.Request) {
	var req ReorderModelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	for _, item := range req.Orders {
		if _, err := tx.Exec("UPDATE api_models SET sort_order = ?, updated_at=CURRENT_TIMESTAMP WHERE id = ?", item.SortOrder, item.ID); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "排序成功")
}

// TestModelResponse is the response for testing model connectivity.
type TestModelResponse struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
	ModelID      int         `json:"modelId"`
	ModelName    string      `json:"modelName"`
	ProviderName string      `json:"providerName"`
	LatencyMs    int64       `json:"latencyMs"`
	Preview      string      `json:"preview"`
	Usage        *TokenUsage `json:"usage,omitempty"`
}

// TokenUsage 是测试请求的 token 用量（取自 OpenAI 兼容响应的 usage 字段）。
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// parseTestResponse 从响应体里提取 AI 回复正文与 token 用量。
// 测试请求按 OpenAI 兼容格式发，但渠道协议可能是 claude / gemini，响应结构不同；
// 解析不出来就退回原始响应（截断），保证弹窗里总有内容可看。
func parseTestResponse(raw []byte) (string, *TokenUsage) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage *TokenUsage `json:"usage"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return truncateText(string(raw), 400), nil
	}
	text := ""
	if len(payload.Choices) > 0 {
		text = payload.Choices[0].Message.Content
	}
	if strings.TrimSpace(text) == "" {
		text = truncateText(string(raw), 400)
	}
	return text, payload.Usage
}

// truncateText 按「字符」截断（不是字节），避免切断多字节 UTF-8。
func truncateText(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit]) + "…"
}

// TestModel tests API connectivity for a specific model config.
func (h *Handler) TestModel(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	m, err := scanModelWithProvider(h.db.QueryRow("SELECT "+apiModelColumnsWithProvider+`
		FROM api_models m
		JOIN api_providers p ON m.provider_id = p.id
		WHERE m.id = ?`, id))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "模型配置不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// Build the chat completion URL（与真实聊天路径一致，正确处理 /chat/completions 尾缀）
	chatURL := ai.BuildApiUrl(m.APIUrl, m.APIFormat, m.UseFullURL == 1)

	// Build a simple test request body
	body := `{"model":"` + m.Model + `","messages":[{"role":"user","content":"hi"}],"max_tokens":5}`
	req, err := http.NewRequest("POST", chatURL, strings.NewReader(body))
	if err != nil {
		respondOK(w, TestModelResponse{Success: false, Message: "创建请求失败: " + err.Error(), ModelID: m.ID, ModelName: m.Name, ProviderName: m.ProviderName})
		return
	}

	headers := ai.BuildAuthHeadersForModel(m.APIKey, m.APIFormat, m.AuthType, m.AccessKey, m.SecretKey, m.GroupID)
	ai.ApplyProviderHeaders(headers, m.APIUrl, m.ProviderID, m.ExtraConfig)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		respondOK(w, TestModelResponse{Success: false, Message: "连接失败: " + err.Error(), ModelID: m.ID, ModelName: m.Name, ProviderName: m.ProviderName, LatencyMs: latency})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		// 读全量（限 64KB）后解析出 AI 回复与 token 用量，而不是把原始 JSON 直接丢给用户
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		text, usage := parseTestResponse(raw)
		respondOK(w, TestModelResponse{
			Success: true, Message: "连接成功",
			ModelID: m.ID, ModelName: m.Name, ProviderName: m.ProviderName,
			LatencyMs: latency, Preview: text, Usage: usage,
		})
	} else {
		respondOK(w, TestModelResponse{
			Success: false, Message: "API 返回状态码: " + resp.Status,
			ModelID: m.ID, ModelName: m.Name, ProviderName: m.ProviderName,
			LatencyMs: latency,
		})
	}
}
