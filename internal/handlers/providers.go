package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetProviders returns all API providers.
func (h *Handler) GetProviders(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, name, provider_type, api_key, api_url, api_format, use_full_url, created_at, updated_at FROM api_providers ORDER BY created_at DESC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.APIProvider
	for rows.Next() {
		var p models.APIProvider
		if err := rows.Scan(&p.ID, &p.Name, &p.ProviderType, &p.APIKey, &p.APIUrl, &p.APIFormat, &p.UseFullURL, &p.CreatedAt, &p.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, p)
	}
	respondOK(w, list)
}

// GetProvider returns a single provider by id, including its models.
func (h *Handler) GetProvider(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var p models.APIProvider
	err := h.db.QueryRow("SELECT id, name, provider_type, api_key, api_url, api_format, use_full_url, created_at, updated_at FROM api_providers WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.ProviderType, &p.APIKey, &p.APIUrl, &p.APIFormat, &p.UseFullURL, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "提供商不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// Fetch models for this provider
	rows, err := h.db.Query("SELECT id, provider_id, name, model, temperature, max_tokens, top_p, frequency_penalty, description, is_default, enabled, sort_order, category, created_at, updated_at FROM api_models WHERE provider_id = ? ORDER BY sort_order ASC", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var models_ []models.APIModel
	for rows.Next() {
		var m models.APIModel
		if err := rows.Scan(&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty, &m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		models_ = append(models_, m)
	}

	type ProviderWithModels struct {
		models.APIProvider
		Models []models.APIModel `json:"models"`
	}

	respondOK(w, ProviderWithModels{APIProvider: p, Models: models_})
}

// CreateProvider creates a new API provider.
func (h *Handler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var p models.APIProvider
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	if p.APIFormat == "" {
		p.APIFormat = "openai"
	}

	result, err := h.db.Exec("INSERT INTO api_providers (name, provider_type, api_key, api_url, api_format, use_full_url) VALUES (?, ?, ?, ?, ?, ?)",
		p.Name, p.ProviderType, p.APIKey, p.APIUrl, p.APIFormat, p.UseFullURL)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	p.ID = int(id)
	respondOK(w, p)
}

// UpdateProvider updates an existing API provider.
func (h *Handler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var p models.APIProvider
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE api_providers SET name=?, provider_type=?, api_key=?, api_url=?, api_format=?, use_full_url=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		p.Name, p.ProviderType, p.APIKey, p.APIUrl, p.APIFormat, p.UseFullURL, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	p.ID = id
	respondOK(w, p)
}

// DeleteProvider deletes an API provider by id (cascades to models).
func (h *Handler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM api_providers WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// DiscoveredModel is a single model returned by model discovery.
type DiscoveredModel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// modelFetcher 按服务商类型拉取模型列表（Strategy Registry：第一个 match 生效）。
type modelFetcher struct {
	match func(p models.APIProvider) bool
	fetch func(p models.APIProvider) ([]DiscoveredModel, error)
}

// fetchProviderJSON 发起 GET 请求并读取响应体，统一处理超时、鉴权与状态码。
func fetchProviderJSON(p models.APIProvider, url string, extraHeaders map[string]string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	if p.APIKey != nil && *p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+*p.APIKey)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取模型列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("提供商返回状态码 %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// isOllamaProvider 判断是否为本地 Ollama：显式选择 ollama 类型，或本地默认端口 11434。
func isOllamaProvider(p models.APIProvider) bool {
	if p.ProviderType == "ollama" {
		return true
	}
	u, err := url.Parse(p.APIUrl)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	return (host == "localhost" || host == "127.0.0.1") && port == "11434"
}

// isGeminiProvider 判断是否为 Gemini：显式选择 gemini 类型，或官方域名。
func isGeminiProvider(p models.APIProvider) bool {
	if p.ProviderType == "gemini" {
		return true
	}
	u, err := url.Parse(p.APIUrl)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return strings.Contains(host, "googleapis") || strings.Contains(host, "generativelanguage")
}

// fetchOllamaModels 从本地 Ollama 的 /api/tags 拉取模型。
func fetchOllamaModels(p models.APIProvider) ([]DiscoveredModel, error) {
	base := strings.TrimRight(strings.TrimSpace(p.APIUrl), "/")
	base, _ = strings.CutSuffix(base, "/v1")
	base, _ = strings.CutSuffix(base, "/api")

	body, err := fetchProviderJSON(p, base+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}
	var out []DiscoveredModel
	for _, m := range resp.Models {
		out = append(out, DiscoveredModel{ID: m.Name, Name: m.Name})
	}
	return out, nil
}

// fetchGeminiModels 从 Gemini 的 /v1beta/models 拉取模型（key 走 x-goog-api-key header）。
func fetchGeminiModels(p models.APIProvider) ([]DiscoveredModel, error) {
	base := strings.TrimRight(strings.TrimSpace(p.APIUrl), "/")
	base, _ = strings.CutSuffix(base, "/v1beta")
	base, _ = strings.CutSuffix(base, "/v1")

	headers := map[string]string{}
	if p.APIKey != nil && *p.APIKey != "" {
		headers["x-goog-api-key"] = *p.APIKey
	}
	body, err := fetchProviderJSON(p, base+"/v1beta/models", headers)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Models []struct {
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}
	var out []DiscoveredModel
	for _, m := range resp.Models {
		id := strings.TrimPrefix(m.Name, "models/")
		name := m.DisplayName
		if name == "" {
			name = id
		}
		out = append(out, DiscoveredModel{ID: id, Name: name})
	}
	return out, nil
}

// fetchOpenAICompatibleModels 从 OpenAI 兼容的 /models 拉取模型（兜底）。
func fetchOpenAICompatibleModels(p models.APIProvider) ([]DiscoveredModel, error) {
	base := strings.TrimRight(strings.TrimSpace(p.APIUrl), "/")
	// api_url 通常以 /chat/completions（OpenAI 兼容）或 /v1/messages（Claude）结尾，
	// 先剥掉路径尾缀再拼 /models，例如 https://api.xxx.com/v1/chat/completions → https://api.xxx.com/v1/models
	base, _ = strings.CutSuffix(base, "/chat/completions")
	base, _ = strings.CutSuffix(base, "/v1/messages")

	body, err := fetchProviderJSON(p, base+"/models", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}
	var out []DiscoveredModel
	for _, m := range resp.Data {
		name := m.Name
		if name == "" {
			name = m.ID
		}
		out = append(out, DiscoveredModel{ID: m.ID, Name: name})
	}
	return out, nil
}

// modelFetchers 注册表：顺序即优先级，第一个 match 生效，OpenAI 兼容兜底必须最后。
var modelFetchers = []modelFetcher{
	{match: isOllamaProvider, fetch: fetchOllamaModels},
	{match: isGeminiProvider, fetch: fetchGeminiModels},
	{match: func(models.APIProvider) bool { return true }, fetch: fetchOpenAICompatibleModels},
}

// DiscoverModels fetches available models from a provider's API.
func (h *Handler) DiscoverModels(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var p models.APIProvider
	err := h.db.QueryRow("SELECT id, name, provider_type, api_key, api_url, api_format, use_full_url, created_at, updated_at FROM api_providers WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.ProviderType, &p.APIKey, &p.APIUrl, &p.APIFormat, &p.UseFullURL, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "提供商不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	for _, f := range modelFetchers {
		if f.match(p) {
			models_, err := f.fetch(p)
			if err != nil {
				respondFail(w, 502, err.Error())
				return
			}
			respondOK(w, models_)
			return
		}
	}
	respondFail(w, 500, "没有匹配的模型拉取器")
}
