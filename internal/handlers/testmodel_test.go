package handlers

import (
	"strings"
	"testing"
)

// TestParseTestResponse 覆盖「模型测试成功」弹窗依赖的响应解析：
// 要从上游响应里安全地取出 AI 回复正文与 token 用量，取不到时退回原始内容。
func TestParseTestResponse(t *testing.T) {
	t.Run("标准 OpenAI 响应", func(t *testing.T) {
		raw := []byte(`{"choices":[{"message":{"role":"assistant","content":"你好！"}}],` +
			`"usage":{"prompt_tokens":8,"completion_tokens":3,"total_tokens":11}}`)
		text, usage := parseTestResponse(raw)
		if text != "你好！" {
			t.Fatalf("text = %q，期望 你好！", text)
		}
		if usage == nil {
			t.Fatal("usage 应为非 nil")
		}
		if usage.PromptTokens != 8 || usage.CompletionTokens != 3 || usage.TotalTokens != 11 {
			t.Fatalf("usage = %+v", usage)
		}
	})

	t.Run("无 usage 字段", func(t *testing.T) {
		raw := []byte(`{"choices":[{"message":{"content":"hi"}}]}`)
		text, usage := parseTestResponse(raw)
		if text != "hi" {
			t.Fatalf("text = %q", text)
		}
		if usage != nil {
			t.Fatalf("usage 应为 nil，实为 %+v", usage)
		}
	})

	t.Run("choices 为空时退回原始响应", func(t *testing.T) {
		raw := []byte(`{"choices":[],"error":{"message":"boom"}}`)
		text, usage := parseTestResponse(raw)
		if !strings.Contains(text, "boom") {
			t.Fatalf("应退回原始响应，实为 %q", text)
		}
		if usage != nil {
			t.Fatalf("usage 应为 nil")
		}
	})

	t.Run("content 为空时退回原始响应", func(t *testing.T) {
		raw := []byte(`{"choices":[{"message":{"content":""}}]}`)
		text, _ := parseTestResponse(raw)
		if !strings.Contains(text, "choices") {
			t.Fatalf("应退回原始响应，实为 %q", text)
		}
	})

	t.Run("非 JSON 时退回原始响应", func(t *testing.T) {
		raw := []byte(`<html>502 Bad Gateway</html>`)
		text, usage := parseTestResponse(raw)
		if text != "<html>502 Bad Gateway</html>" {
			t.Fatalf("text = %q", text)
		}
		if usage != nil {
			t.Fatalf("usage 应为 nil")
		}
	})

	t.Run("按字符截断，不切坏多字节", func(t *testing.T) {
		raw := []byte(strings.Repeat("中", 500))
		text, _ := parseTestResponse(raw)
		if len([]rune(text)) != 401 {
			t.Fatalf("rune 数应为 401（400 字 + 省略号），实为 %d", len([]rune(text)))
		}
		if !strings.HasSuffix(text, "…") {
			t.Fatal("应带省略号")
		}
		// 按字节截断会把「中」切成半个字，产生 U+FFFD
		if strings.ContainsRune(text, '\uFFFD') {
			t.Fatal("截断产生了非法字符（说明是按字节切的）")
		}
	})
}
