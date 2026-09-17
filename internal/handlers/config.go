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
	rows, err := h.db.Query(`
		SELECT m.id, m.provider_id, m.name, m.model, m.temperature, m.max_tokens, m.top_p, m.frequency_penalty,
			m.description, m.is_default, m.enabled, m.sort_order, m.category, m.created_at, m.updated_at,
			p.api_key, p.api_url, p.api_format, p.use_full_url, p.name as provider_name
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
		var m models.APIModelWithProvider
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty,
			&m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category, &m.CreatedAt, &m.UpdatedAt,
			&m.APIKey, &m.APIUrl, &m.APIFormat, &m.UseFullURL, &m.ProviderName); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// GetDefaultConfig returns the default enabled api_model with provider info.
func (h *Handler) GetDefaultConfig(w http.ResponseWriter, r *http.Request) {
	var m models.APIModelWithProvider
	err := h.db.QueryRow(`
		SELECT m.id, m.provider_id, m.name, m.model, m.temperature, m.max_tokens, m.top_p, m.frequency_penalty,
			m.description, m.is_default, m.enabled, m.sort_order, m.category, m.created_at, m.updated_at,
			p.api_key, p.api_url, p.api_format, p.use_full_url, p.name as provider_name
		FROM api_models m
		JOIN api_providers p ON m.provider_id = p.id
		WHERE m.is_default = 1 AND m.enabled = 1
		ORDER BY m.sort_order ASC
		LIMIT 1`).Scan(&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty,
		&m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category, &m.CreatedAt, &m.UpdatedAt,
		&m.APIKey, &m.APIUrl, &m.APIFormat, &m.UseFullURL, &m.ProviderName)
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

	rows, err := h.db.Query(`
		SELECT m.id, m.provider_id, m.name, m.model, m.temperature, m.max_tokens, m.top_p, m.frequency_penalty,
			m.description, m.is_default, m.enabled, m.sort_order, m.category, m.created_at, m.updated_at,
			p.api_key, p.api_url, p.api_format, p.use_full_url, p.name as provider_name
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
		var m models.APIModelWithProvider
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty,
			&m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category, &m.CreatedAt, &m.UpdatedAt,
			&m.APIKey, &m.APIUrl, &m.APIFormat, &m.UseFullURL, &m.ProviderName); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// CreateConfig creates a a new api_model config.
func (h *Handler) CreateConfig(w http.ResponseWriter, r *http.Request) {
	var m models.APIModel
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO api_models (provider_id, name, model, temperature, max_tokens, top_p, frequency_penalty, description, is_default, enabled, sort_order, category) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		m.ProviderID, m.Name, m.Model, m.Temperature, m.MaxTokens, m.TopP, m.FrequencyPenalty, m.Description, m.IsDefault, m.Enabled, m.SortOrder, m.Category)
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

	_, err := h.db.Exec("UPDATE api_models SET provider_id=?, name=?, model=?, temperature=?, max_tokens=?, top_p=?, frequency_penalty=?, description=?, is_default=?, enabled=?, sort_order=?, category=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		m.ProviderID, m.Name, m.Model, m.Temperature, m.MaxTokens, m.TopP, m.FrequencyPenalty, m.Description, m.IsDefault, m.Enabled, m.SortOrder, m.Category, id)
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
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ModelID   int    `json:"modelId"`
	ModelName string `json:"modelName"`
	LatencyMs int64  `json:"latencyMs"`
	Preview   string `json:"preview"`
}

// TestModel tests API connectivity for a specific model config.
func (h *Handler) TestModel(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var m models.APIModelWithProvider
	err := h.db.QueryRow(`
		SELECT m.id, m.provider_id, m.name, m.model, m.temperature, m.max_tokens, m.top_p, m.frequency_penalty,
			m.description, m.is_default, m.enabled, m.sort_order, m.category, m.created_at, m.updated_at,
			p.api_key, p.api_url, p.api_format, p.use_full_url, p.name as provider_name
		FROM api_models m
		JOIN api_providers p ON m.provider_id = p.id
		WHERE m.id = ?`, id).Scan(&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty,
		&m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category, &m.CreatedAt, &m.UpdatedAt,
		&m.APIKey, &m.APIUrl, &m.APIFormat, &m.UseFullURL, &m.ProviderName)
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
		respondOK(w, TestModelResponse{Success: false, Message: "创建请求失败: " + err.Error()})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if m.APIKey != nil && *m.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+*m.APIKey)
	}

	start := time.Now()
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		respondOK(w, TestModelResponse{Success: false, Message: "连接失败: " + err.Error(), ModelID: m.ID, ModelName: m.Name, LatencyMs: latency})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		preview := ""
		if b, err := io.ReadAll(io.LimitReader(resp.Body, 400)); err == nil {
			preview = string(b)
		}
		respondOK(w, TestModelResponse{Success: true, Message: "连接成功", ModelID: m.ID, ModelName: m.Name, LatencyMs: latency, Preview: preview})
	} else {
		respondOK(w, TestModelResponse{Success: false, Message: "API 返回状态码: " + resp.Status, ModelID: m.ID, ModelName: m.Name, LatencyMs: latency})
	}
}
