package handlers

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/database"
)

func pstr(s string) *string { return &s }

// TestAPIProviderColumnsRoundTrip 验证 providerColumns 常量与真实表结构一致，且新列能写能读。
func TestAPIProviderColumnsRoundTrip(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	res, err := db.Exec(`INSERT INTO api_providers
		(name, provider_type, api_key, api_url, api_format, use_full_url,
		 auth_type, access_key, secret_key, group_id, endpoint_id, concurrency, channel_name, extra_config)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"t", "other", "k", "http://x", "openai", 0,
		"ak_sk", "AK", "SK", "GID", "EP", 4, "渠道A", `{"a":1}`)
	if err != nil {
		t.Fatalf("insert provider: %v", err)
	}
	id, _ := res.LastInsertId()

	got, err := scanProvider(db.QueryRow("SELECT "+providerColumns+" FROM api_providers WHERE id = ?", id))
	if err != nil {
		t.Fatalf("scan provider: %v", err)
	}
	if got.AuthType != "ak_sk" || got.Concurrency != 4 || got.ChannelName != "渠道A" {
		t.Fatalf("unexpected provider: %+v", got)
	}
	if got.AccessKey == nil || *got.AccessKey != "AK" {
		t.Fatalf("access_key not persisted: %+v", got.AccessKey)
	}
}

// TestAPIModelColumnsRoundTrip 验证 apiModelColumnsWithProvider 与表结构一致，且新字段能写能读。
func TestAPIModelColumnsRoundTrip(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	res, err := db.Exec(`INSERT INTO api_providers (name, provider_type, api_url) VALUES (?, ?, ?)`, "p", "other", "http://x")
	if err != nil {
		t.Fatalf("insert provider: %v", err)
	}
	pid, _ := res.LastInsertId()

	res, err = db.Exec(`INSERT INTO api_models
		(provider_id, name, model, temperature, max_tokens, top_p, frequency_penalty, description,
		 is_default, enabled, sort_order, category,
		 config_type, prompt, prompt_role, top_k, presence_penalty, stop, max_tokens_mode,
		 thinking_mode, thinking_effort, generation_config_source, request_rules,
		 param_groups, default_params, is_free, input_ratio, output_ratio)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pid, "封面模型", "gpt-image-1", 0.7, 2000, 0.9, 0.0, "d",
		1, 1, 0, "image",
		"image", "sys", "system", 40, 0.1, `["stop"]`, "auto",
		1, "high", "manual", `{"submit":{"url":"/x"}}`,
		`[{"name":"duration","type":"number","options":[{"label":"5s","value":5}]}]`,
		`{"duration":5,"resolution":"720p"}`, 0, 1.0, 2.0)
	if err != nil {
		t.Fatalf("insert model: %v", err)
	}
	mid, _ := res.LastInsertId()

	got, err := scanModelWithProvider(db.QueryRow("SELECT "+apiModelColumnsWithProvider+`
		FROM api_models m JOIN api_providers p ON m.provider_id = p.id WHERE m.id = ?`, mid))
	if err != nil {
		t.Fatalf("scan model: %v", err)
	}
	if got.ConfigType != "image" {
		t.Fatalf("config_type = %q, want image", got.ConfigType)
	}
	if got.TopK == nil || *got.TopK != 40 {
		t.Fatalf("top_k not persisted: %v", got.TopK)
	}
	if got.ThinkingMode != 1 || got.ThinkingEffort != "high" {
		t.Fatalf("thinking fields wrong: %+v", got)
	}
	if got.RequestRules == nil || *got.RequestRules == "" {
		t.Fatalf("request_rules not persisted")
	}
	if got.ParamGroups == "" || !strings.Contains(got.ParamGroups, "duration") {
		t.Fatalf("param_groups not persisted: %q", got.ParamGroups)
	}
	if got.DefaultParams == "" || !strings.Contains(got.DefaultParams, "720p") {
		t.Fatalf("default_params not persisted: %q", got.DefaultParams)
	}
	if got.ProviderName != "p" {
		t.Fatalf("provider_name = %q", got.ProviderName)
	}
	if got.APIModel.MaxTokensMode != "auto" || got.APIModel.PromptRole != "system" {
		t.Fatalf("mode/role wrong: %+v", got.APIModel)
	}
}
