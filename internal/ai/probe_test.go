package ai

import (
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func probeCfg(format string) models.APIModelWithProvider {
	return models.APIModelWithProvider{
		APIModel:  models.APIModel{Model: "m1"},
		APIFormat: format,
	}
}

// 探测请求体必须最小化：不带 temperature / top_p / stop。
// 否则"参数不兼容"的 400 会被误判成"长度超限"。
func TestBuildProbeBody_Minimal(t *testing.T) {
	body, family, err := BuildProbeBody(probeCfg("openai"), "hi", 4096)
	if err != nil || family != "openai" {
		t.Fatalf("unexpected: %v %s", err, family)
	}
	if body["max_tokens"] != 4096 {
		t.Fatalf("max_tokens 未写入: %v", body)
	}
	if body["stream"] != false {
		t.Fatal("探测必须非流式")
	}
	for _, k := range []string{"temperature", "top_p", "frequency_penalty", "stop"} {
		if _, ok := body[k]; ok {
			t.Fatalf("探测请求不应带 %s", k)
		}
	}
	msgs, _ := body["messages"].([]map[string]string)
	if len(msgs) != 1 || msgs[0]["content"] != "hi" {
		t.Fatalf("messages 不符合预期: %v", body["messages"])
	}
}

func TestBuildProbeBody_Anthropic(t *testing.T) {
	body, family, err := BuildProbeBody(probeCfg("anthropic"), "hi", 8192)
	if err != nil || family != "anthropic" {
		t.Fatalf("unexpected: %v %s", err, family)
	}
	if body["max_tokens"] != 8192 {
		t.Fatal("Anthropic 必须带 max_tokens")
	}
	if _, ok := body["temperature"]; ok {
		t.Fatal("不应带 temperature")
	}
}

func TestBuildProbeBody_Responses(t *testing.T) {
	body, family, err := BuildProbeBody(probeCfg("responses"), "hi", 16384)
	if err != nil || family != "responses" {
		t.Fatalf("unexpected: %v %s", err, family)
	}
	if body["max_output_tokens"] != 16384 {
		t.Fatalf("responses 应使用 max_output_tokens: %v", body)
	}
	if body["input"] != "hi" {
		t.Fatal("responses 应使用 input")
	}
}

// custom_json 的字段位置由用户规则决定，无法安全构造 —— 必须诚实拒绝。
func TestBuildProbeBody_CustomJSONUnsupported(t *testing.T) {
	if _, _, err := BuildProbeBody(probeCfg("custom_json"), "hi", 1024); err != ErrProbeUnsupported {
		t.Fatalf("应返回 ErrProbeUnsupported，实得 %v", err)
	}
}

func TestParseProbeObservation_OpenAI(t *testing.T) {
	raw := []byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"length"}],
		"usage":{"prompt_tokens":5,"completion_tokens":4096}}`)
	obs, ok := ParseProbeObservation("openai", raw)
	if !ok || !obs.HasUsage || obs.CompletionTokens != 4096 || obs.FinishReason != "length" {
		t.Fatalf("解析错误: %+v ok=%v", obs, ok)
	}
	if !IsTruncated("openai", obs.FinishReason) {
		t.Fatal("length 应判为截断")
	}
}

func TestParseProbeObservation_Anthropic(t *testing.T) {
	raw := []byte(`{"content":[{"type":"text","text":"ok"}],"stop_reason":"max_tokens",
		"usage":{"input_tokens":5,"output_tokens":8192}}`)
	obs, ok := ParseProbeObservation("anthropic", raw)
	if !ok || obs.CompletionTokens != 8192 || obs.FinishReason != "max_tokens" {
		t.Fatalf("解析错误: %+v ok=%v", obs, ok)
	}
	if !IsTruncated("anthropic", obs.FinishReason) {
		t.Fatal("max_tokens 应判为截断")
	}
}

func TestParseProbeObservation_Responses(t *testing.T) {
	raw := []byte(`{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},
		"usage":{"input_tokens":5,"output_tokens":1024}}`)
	obs, ok := ParseProbeObservation("responses", raw)
	if !ok || obs.CompletionTokens != 1024 || obs.FinishReason != "max_output_tokens" {
		t.Fatalf("解析错误: %+v ok=%v", obs, ok)
	}
	if !IsTruncated("responses", obs.FinishReason) {
		t.Fatal("max_output_tokens 应判为截断")
	}
}

// 没有 usage 也没有 finish_reason：不可测量，必须返回 ok=false，
// 不能当成"成功且为 0"。
func TestParseProbeObservation_Unmeasurable(t *testing.T) {
	if _, ok := ParseProbeObservation("openai", []byte(`{"foo":1}`)); ok {
		t.Fatal("无可用字段时应判为不可测量")
	}
	if _, ok := ParseProbeObservation("openai", []byte(`not json`)); ok {
		t.Fatal("非法 JSON 应判为失败")
	}
}
