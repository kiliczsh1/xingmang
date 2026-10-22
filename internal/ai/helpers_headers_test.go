package ai

import "testing"

func TestParseExtraHeaders(t *testing.T) {
	cases := []struct {
		name string
		raw  *string
		want int
	}{
		{"nil", nil, 0},
		{"empty", pstr(""), 0},
		{"invalid json", pstr("{not json"), 0},
		{"no headers key", pstr(`{"other":1}`), 0},
		{"empty headers", pstr(`{"headers":{}}`), 0},
		{"one header", pstr(`{"headers":{"x-opencode-session":"mine"}}`), 1},
		{"blank key dropped", pstr(`{"headers":{"  ":"v","ok":"1"}}`), 1},
	}
	for _, c := range cases {
		if got := ParseExtraHeaders(c.raw); len(got) != c.want {
			t.Fatalf("%s: 得到 %d 个头，期望 %d（%v）", c.name, len(got), c.want, got)
		}
	}
}

// TestApplyProviderHeadersOpenCodeGo 覆盖本次修复的核心场景：
// OpenCode Go 端点必须带上 x-opencode-session，否则上游拒绝请求。
func TestApplyProviderHeadersOpenCodeGo(t *testing.T) {
	// 1) Go 端点：自动注入会话头，值稳定且不透明
	h := map[string]string{}
	ApplyProviderHeaders(h, "https://opencode.ai/zen/go/v1/chat/completions", 15, nil)
	got, ok := h[OpenCodeSessionHeaderName]
	if !ok || got == "" {
		t.Fatalf("Go 端点未注入 %s：%v", OpenCodeSessionHeaderName, h)
	}
	if got != "xingmang-15" {
		t.Fatalf("会话 ID = %q，期望 xingmang-15", got)
	}

	// 2) 同一渠道两次调用必须一致（上游要求「同一会话稳定」）
	h2 := map[string]string{}
	ApplyProviderHeaders(h2, "https://opencode.ai/zen/go/v1/chat/completions", 15, nil)
	if h2[OpenCodeSessionHeaderName] != got {
		t.Fatalf("同一渠道会话 ID 不稳定：%q vs %q", h2[OpenCodeSessionHeaderName], got)
	}

	// 3) 不同渠道必须不同（上游要求「不同会话唯一」）
	h3 := map[string]string{}
	ApplyProviderHeaders(h3, "https://opencode.ai/zen/go/v1/chat/completions", 16, nil)
	if h3[OpenCodeSessionHeaderName] == got {
		t.Fatalf("不同渠道会话 ID 撞车：%q", got)
	}

	// 4) 非 Go 端点不得注入（避免给无关上游发多余头）
	h4 := map[string]string{}
	ApplyProviderHeaders(h4, "https://api.deepseek.com/v1", 15, nil)
	if _, ok := h4[OpenCodeSessionHeaderName]; ok {
		t.Fatalf("非 Go 端点被误注入会话头：%v", h4)
	}
}

// TestApplyProviderHeadersUserOverride 用户显式配置优先，不被自动值覆盖。
func TestApplyProviderHeadersUserOverride(t *testing.T) {
	cfg := `{"headers":{"x-opencode-session":"my-own-session","X-Custom":"v1"}}`
	h := map[string]string{}
	ApplyProviderHeaders(h, "https://opencode.ai/zen/go/v1/chat/completions", 15, &cfg)

	if h[OpenCodeSessionHeaderName] != "my-own-session" {
		t.Fatalf("用户配置被覆盖：%q", h[OpenCodeSessionHeaderName])
	}
	if h["X-Custom"] != "v1" {
		t.Fatalf("自定义头未注入：%v", h)
	}

	// 已有键（如 Authorization）不被 extra_config 覆盖
	h2 := map[string]string{"Authorization": "Bearer real-key"}
	ApplyProviderHeaders(h2, "https://api.example.com/v1", 1, &cfg)
	if h2["Authorization"] != "Bearer real-key" {
		t.Fatalf("鉴权头被 extra_config 覆盖：%q", h2["Authorization"])
	}
}

func pstr(s string) *string { return &s }
