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

	"github.com/xingmang/xingmang-desktop/internal/ai"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetProviders returns all API providers.
func (h *Handler) GetProviders(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT " + providerColumns + " FROM api_providers ORDER BY created_at DESC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.APIProvider
	for rows.Next() {
		p, err := scanProvider(rows)
		if err != nil {
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

	p, err := scanProvider(h.db.QueryRow("SELECT "+providerColumns+" FROM api_providers WHERE id = ?", id))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "提供商不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// Fetch models for this provider
	rows, err := h.db.Query("SELECT "+apiModelColumns+" FROM api_models WHERE provider_id = ? ORDER BY sort_order ASC", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var models_ []models.APIModel
	for rows.Next() {
		m, err := scanModel(rows)
		if err != nil {
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

// normalizeProviderDefaults 补 provider 的默认值：auth_type 缺省 api_key，并发缺省 1。
func normalizeProviderDefaults(p *models.APIProvider) {
	if p.AuthType == "" {
		p.AuthType = "api_key"
	}
	if p.Concurrency <= 0 {
		p.Concurrency = 1
	}
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
	normalizeProviderDefaults(&p)

	result, err := h.db.Exec(`INSERT INTO api_providers
		(name, provider_type, api_key, api_url, api_format, use_full_url,
		 auth_type, access_key, secret_key, group_id, endpoint_id, concurrency, channel_name, extra_config)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.ProviderType, p.APIKey, p.APIUrl, p.APIFormat, p.UseFullURL,
		p.AuthType, p.AccessKey, p.SecretKey, p.GroupID, p.EndpointID, p.Concurrency, p.ChannelName, p.ExtraConfig)
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
	normalizeProviderDefaults(&p)

	_, err := h.db.Exec(`UPDATE api_providers SET
		name=?, provider_type=?, api_key=?, api_url=?, api_format=?, use_full_url=?,
		auth_type=?, access_key=?, secret_key=?, group_id=?, endpoint_id=?, concurrency=?, channel_name=?, extra_config=?,
		updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		p.Name, p.ProviderType, p.APIKey, p.APIUrl, p.APIFormat, p.UseFullURL,
		p.AuthType, p.AccessKey, p.SecretKey, p.GroupID, p.EndpointID, p.Concurrency, p.ChannelName, p.ExtraConfig, id)
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
	// OwnedBy 上游声明的归属方（OpenAI 兼容接口的 owned_by 字段）；
	// 部分上游不返回，前端「来源」列会显示为 —。
	OwnedBy string `json:"owned_by"`

	// ContextWindow / MaxOutputTokens 是**上游主动给出**的能力值，0 = 该上游没提供。
	//
	// ⚠ 别以为拿不到：官方 API 一直是给的，只是各家字段名不统一，而早期版本只解析
	// id/name/owned_by，把这两个值整个丢掉了 —— 于是前端只能靠本地表猜，
	// 猜出一批严重偏低的错值。实测（2026-10-16 实抓）：
	//   DeepSeek 官方 GET /v1/models    → {"context_window":1048576,"max_output_tokens":393216}
	//   Gemini   官方 GET /v1beta/models → {"inputTokenLimit":1048576,"outputTokenLimit":65536}
	//   Anthropic     GET /v1/models    → {"max_input_tokens":...,"max_tokens":...}
	// 第三方中转多数不透传这两个字段，拉不到就保持 0，由前端本地表兜底。
	ContextWindow   int `json:"context_window,omitempty"`
	MaxOutputTokens int `json:"max_output_tokens,omitempty"`
}

// firstPositive 返回第一个大于 0 的值。各家上下文字段名不统一，按优先级归一到一处。
func firstPositive(vals ...int) int {
	for _, v := range vals {
		if v > 0 {
			return v
		}
	}
	return 0
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
	// 保持原有鉴权方式不变，仅追加 extra_config 自定义头与 OpenCode Go 必需的会话头。
	// 端点判断用 p.APIUrl（url 是派生出的 /models 地址）。
	if p.APIKey != nil && *p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+*p.APIKey)
	}
	headers := map[string]string{}
	ai.ApplyProviderHeaders(headers, p.APIUrl, p.ID, p.ExtraConfig)
	for k, v := range headers {
		req.Header.Set(k, v)
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
			// Gemini 官方就带这两个字段（输入 1,048,576 / 输出 65,536 这类）
			InputTokenLimit  int `json:"inputTokenLimit"`
			OutputTokenLimit int `json:"outputTokenLimit"`
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
		out = append(out, DiscoveredModel{
			ID: id, Name: name,
			ContextWindow:   m.InputTokenLimit,
			MaxOutputTokens: m.OutputTokenLimit,
		})
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
			ID      string `json:"id"`
			Name    string `json:"name"`
			OwnedBy string `json:"owned_by"`

			// token 上限：各家用不同字段名，全部收下来再归一化（缺失时为零值）。
			//   DeepSeek 官方: context_window / max_output_tokens
			//   Anthropic:     max_input_tokens / max_tokens
			//   部分自建/中转:  context_length
			ContextWindow   int `json:"context_window"`
			ContextLength   int `json:"context_length"`
			MaxInputTokens  int `json:"max_input_tokens"`
			MaxOutputTokens int `json:"max_output_tokens"`
			MaxTokens       int `json:"max_tokens"`
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
		out = append(out, DiscoveredModel{
			ID: m.ID, Name: name, OwnedBy: m.OwnedBy,
			ContextWindow:   firstPositive(m.ContextWindow, m.ContextLength, m.MaxInputTokens),
			MaxOutputTokens: firstPositive(m.MaxOutputTokens, m.MaxTokens),
		})
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

	p, err := scanProvider(h.db.QueryRow("SELECT "+providerColumns+" FROM api_providers WHERE id = ?", id))
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
