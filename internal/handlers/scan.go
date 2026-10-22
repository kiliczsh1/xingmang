package handlers

import (
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 本文件集中放 api_providers / api_models 的列清单与扫描辅助，
// 避免同一份列顺序在 providers.go / config.go / media.go 里重复维护、改一处漏一处。

// providerColumns 是 api_providers 的完整列清单（顺序须与 scanProvider 一致）。
// 可空列一律 COALESCE 兜底，避免 NULL 扫进 Go 的 string 报错。
const providerColumns = `id, name, provider_type, api_key, api_url, api_format, use_full_url,
	COALESCE(auth_type, 'api_key') AS auth_type, access_key, secret_key, group_id, endpoint_id,
	COALESCE(concurrency, 1) AS concurrency, COALESCE(channel_name, '') AS channel_name, extra_config,
	created_at, updated_at`

// apiModelColumns 是 api_models 的完整列清单（不带表别名，顺序须与 scanModel 一致）。
const apiModelColumns = `id, provider_id, name, model, temperature, max_tokens, top_p, frequency_penalty,
	description, is_default, enabled, sort_order, COALESCE(category, '') AS category,
	COALESCE(config_type, 'chat') AS config_type, COALESCE(prompt, '') AS prompt,
	COALESCE(prompt_role, 'system') AS prompt_role, top_k, presence_penalty,
	COALESCE(stop, '') AS stop, COALESCE(max_tokens_mode, 'auto') AS max_tokens_mode,
	COALESCE(thinking_mode, 0) AS thinking_mode, COALESCE(thinking_effort, 'default') AS thinking_effort,
	COALESCE(generation_config_source, 'manual') AS generation_config_source, request_rules,
	COALESCE(param_groups, '') AS param_groups, COALESCE(default_params, '') AS default_params,
	COALESCE(is_free, 0) AS is_free, input_ratio, output_ratio, created_at, updated_at`

// apiModelColumnsWithProvider 是 api_models JOIN api_providers 的列清单，
// 别名固定为 m / p（顺序须与 scanModelWithProvider 一致）。
const apiModelColumnsWithProvider = `m.id, m.provider_id, m.name, m.model, m.temperature, m.max_tokens, m.top_p, m.frequency_penalty,
	m.description, m.is_default, m.enabled, m.sort_order, COALESCE(m.category, '') AS category,
	COALESCE(m.config_type, 'chat') AS config_type, COALESCE(m.prompt, '') AS prompt,
	COALESCE(m.prompt_role, 'system') AS prompt_role, m.top_k, m.presence_penalty,
	COALESCE(m.stop, '') AS stop, COALESCE(m.max_tokens_mode, 'auto') AS max_tokens_mode,
	COALESCE(m.thinking_mode, 0) AS thinking_mode, COALESCE(m.thinking_effort, 'default') AS thinking_effort,
	COALESCE(m.generation_config_source, 'manual') AS generation_config_source, m.request_rules,
	COALESCE(m.param_groups, '') AS param_groups, COALESCE(m.default_params, '') AS default_params,
	COALESCE(m.is_free, 0) AS is_free, m.input_ratio, m.output_ratio, m.created_at, m.updated_at,
	p.api_key, p.api_url, p.api_format, p.use_full_url, p.name AS provider_name,
	COALESCE(p.auth_type, 'api_key') AS auth_type, p.access_key, p.secret_key, p.group_id, p.endpoint_id,
	COALESCE(p.concurrency, 1) AS concurrency, COALESCE(p.channel_name, '') AS channel_name, p.extra_config`

// scanProvider 按 providerColumns 顺序扫描一行 api_providers。
// （rowScanner 接口复用 books.go 中的定义。）
func scanProvider(s rowScanner) (models.APIProvider, error) {
	var p models.APIProvider
	err := s.Scan(
		&p.ID, &p.Name, &p.ProviderType, &p.APIKey, &p.APIUrl, &p.APIFormat, &p.UseFullURL,
		&p.AuthType, &p.AccessKey, &p.SecretKey, &p.GroupID, &p.EndpointID, &p.Concurrency, &p.ChannelName, &p.ExtraConfig,
		&p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

// scanModel 按 apiModelColumns 顺序扫描一行 api_models。
func scanModel(s rowScanner) (models.APIModel, error) {
	var m models.APIModel
	err := s.Scan(
		&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty,
		&m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category,
		&m.ConfigType, &m.Prompt, &m.PromptRole, &m.TopK, &m.PresencePenalty, &m.Stop, &m.MaxTokensMode,
		&m.ThinkingMode, &m.ThinkingEffort, &m.GenerationConfigSource, &m.RequestRules,
		&m.ParamGroups, &m.DefaultParams, &m.IsFree,
		&m.InputRatio, &m.OutputRatio, &m.CreatedAt, &m.UpdatedAt,
	)
	return m, err
}

// scanModelWithProvider 按 apiModelColumnsWithProvider 顺序扫描一行 join 结果。
func scanModelWithProvider(s rowScanner) (models.APIModelWithProvider, error) {
	var m models.APIModelWithProvider
	err := s.Scan(
		&m.ID, &m.ProviderID, &m.Name, &m.Model, &m.Temperature, &m.MaxTokens, &m.TopP, &m.FrequencyPenalty,
		&m.Description, &m.IsDefault, &m.Enabled, &m.SortOrder, &m.Category,
		&m.ConfigType, &m.Prompt, &m.PromptRole, &m.TopK, &m.PresencePenalty, &m.Stop, &m.MaxTokensMode,
		&m.ThinkingMode, &m.ThinkingEffort, &m.GenerationConfigSource, &m.RequestRules,
		&m.ParamGroups, &m.DefaultParams, &m.IsFree,
		&m.InputRatio, &m.OutputRatio, &m.CreatedAt, &m.UpdatedAt,
		&m.APIKey, &m.APIUrl, &m.APIFormat, &m.UseFullURL, &m.ProviderName,
		&m.AuthType, &m.AccessKey, &m.SecretKey, &m.GroupID, &m.EndpointID, &m.Concurrency, &m.ChannelName, &m.ExtraConfig,
	)
	return m, err
}
