package ai

import "testing"

// TestBuildAuthHeadersForModel 覆盖「认证方式 + 协议」的全部组合。
//
// 背景：主聊天链路（agent/loop.go、handlers/ai.go）此前调用的是只认协议的
// BuildAuthHeaders，导致 ak_sk / group_id 认证的渠道鉴权头是错的。
// 本测试把修复后的正确行为固化下来。
func TestBuildAuthHeadersForModel(t *testing.T) {
	key := "sk-test"
	ak := "AK123"
	sk := "SK456"
	gid := "grp-1"

	cases := []struct {
		name      string
		apiKey    *string
		apiFormat string
		authType  string
		access    *string
		secret    *string
		group     *string
		want      map[string]string
		absent    []string
	}{
		{
			name:   "api_key + openai → Bearer",
			apiKey: &key, apiFormat: "openai", authType: "api_key",
			want: map[string]string{"Authorization": "Bearer sk-test"},
		},
		{
			name:   "api_key + claude → x-api-key + anthropic-version",
			apiKey: &key, apiFormat: "claude", authType: "api_key",
			want:   map[string]string{"x-api-key": "sk-test", "anthropic-version": "2023-06-01"},
			absent: []string{"Authorization"},
		},
		{
			name:   "api_key + gemini → x-goog-api-key",
			apiKey: &key, apiFormat: "gemini", authType: "api_key",
			want:   map[string]string{"x-goog-api-key": "sk-test"},
			absent: []string{"Authorization"},
		},
		{
			name:   "endpoint_id 与 api_key 同构（endpoint 只是 URL 变量）",
			apiKey: &key, apiFormat: "openai", authType: "endpoint_id",
			want: map[string]string{"Authorization": "Bearer sk-test"},
		},
		{
			name:   "ak_sk → 密钥对，不发 Bearer",
			apiKey: &key, apiFormat: "openai", authType: "ak_sk", access: &ak, secret: &sk,
			want:   map[string]string{"X-Access-Key": "AK123", "X-Secret-Key": "SK456"},
			absent: []string{"Authorization"},
		},
		{
			name:   "group_id → Bearer + X-Group-Id",
			apiKey: &key, apiFormat: "openai", authType: "group_id", group: &gid,
			want: map[string]string{"Authorization": "Bearer sk-test", "X-Group-Id": "grp-1"},
		},
		{
			name:   "authType 大小写与空格容错",
			apiKey: &key, apiFormat: "openai", authType: "  AK_SK  ", access: &ak, secret: &sk,
			want:   map[string]string{"X-Access-Key": "AK123", "X-Secret-Key": "SK456"},
			absent: []string{"Authorization"},
		},
		{
			name:   "apiKey 为 nil 不 panic（原实现 *config.APIKey 会崩）",
			apiKey: nil, apiFormat: "openai", authType: "api_key",
			want: map[string]string{"Authorization": "Bearer "},
		},
	}

	for _, c := range cases {
		got := BuildAuthHeadersForModel(c.apiKey, c.apiFormat, c.authType, c.access, c.secret, c.group)
		if got["Content-Type"] != "application/json" {
			t.Fatalf("%s: 缺少 Content-Type", c.name)
		}
		for k, want := range c.want {
			if got[k] != want {
				t.Fatalf("%s: %s = %q，期望 %q", c.name, k, got[k], want)
			}
		}
		for _, k := range c.absent {
			if _, ok := got[k]; ok {
				t.Fatalf("%s: 不应出现 %s（%v）", c.name, k, got)
			}
		}
	}
}
