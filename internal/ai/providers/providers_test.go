package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func strPtr(s string) *string { return &s }

func TestExtractByPath(t *testing.T) {
	root := map[string]interface{}{
		"data": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"url": "u1"},
				map[string]interface{}{"url": "u2"},
			},
			"task_id": "t1",
		},
	}
	if got := firstString(extractByPath(root, "data.task_id")); got != "t1" {
		t.Fatalf("data.task_id = %q, want t1", got)
	}
	nodes := extractByPath(root, "data.items[].url")
	if len(nodes) != 2 || nodes[0] != "u1" || nodes[1] != "u2" {
		t.Fatalf("data.items[].url = %v, want [u1 u2]", nodes)
	}
	if got := extractByPath(root, "data.nope"); len(got) != 0 {
		t.Fatalf("missing path should be empty, got %v", got)
	}
}

func TestNormalizeImagePayload(t *testing.T) {
	cases := []map[string]interface{}{
		{"data": []interface{}{map[string]interface{}{"b64_json": "AAA"}}},
		{"data": []interface{}{map[string]interface{}{"url": "https://x/1.png"}}},
		{"images": []interface{}{"https://x/2.png", map[string]interface{}{"url": "https://x/3.png"}}},
		{"url": "https://x/4.png"},
	}
	for i, raw := range cases {
		got := NormalizeImagePayload(raw)
		if len(got) == 0 {
			t.Fatalf("case %d: expected at least one image", i)
		}
	}
}

func TestApplyVarsDeep(t *testing.T) {
	in := map[string]interface{}{
		"prompt": "{{prompt}}",
		"nested": []interface{}{"{{model}}"},
		"keep":   42,
	}
	out := applyVarsDeep(in, map[string]string{"prompt": "cat", "model": "m1"}).(map[string]interface{})
	if out["prompt"] != "cat" {
		t.Fatalf("prompt = %v", out["prompt"])
	}
	if out["nested"].([]interface{})[0] != "m1" {
		t.Fatalf("nested = %v", out["nested"])
	}
	if out["keep"] != 42 {
		t.Fatalf("keep = %v", out["keep"])
	}
}

func TestResolveImagePicksCustomThenFallback(t *testing.T) {
	custom := &models.APIModelWithProvider{}
	custom.APIFormat = "custom_json"
	if a := ResolveImage(custom); a.Name() != "custom_json_image" {
		t.Fatalf("custom_json should pick custom_json_image, got %s", a.Name())
	}

	plain := &models.APIModelWithProvider{}
	plain.APIFormat = "openai"
	if a := ResolveImage(plain); a.Name() != "openai_image" {
		t.Fatalf("openai should fall back to openai_image, got %s", a.Name())
	}
}

func TestResolveVideoNoMatchReturnsNil(t *testing.T) {
	plain := &models.APIModelWithProvider{}
	plain.APIFormat = "openai"
	if a := ResolveVideo(plain); a != nil {
		t.Fatalf("expected nil video adapter, got %s", a.Name())
	}
}

// TestCustomJSONImageEndToEnd 用一个本地 mock 服务端验证自定义 JSON 图片链路。
func TestCustomJSONImageEndToEnd(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"url":"https://cdn.example.com/a.png"}]}`)
	}))
	defer srv.Close()

	rules := `{"submit":{"url":"/v1/img","body":{"prompt":"{{prompt}}","model":"{{model}}"}},"response":{"images":"data[].url"}}`
	m := &models.APIModelWithProvider{}
	m.APIUrl = srv.URL
	m.APIModel.Model = "m1"
	m.APIFormat = "custom_json"
	m.APIModel.RequestRules = strPtr(rules)

	adapter := ResolveImage(m)
	res, err := adapter.Generate(context.Background(), m, ImageRequest{Prompt: "cat"})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if len(res.Images) != 1 || res.Images[0].URL != "https://cdn.example.com/a.png" {
		t.Fatalf("unexpected images: %+v", res.Images)
	}
}

// TestCustomJSONVideoEndToEnd 验证「提交 + 轮询」两段式视频链路。
func TestCustomJSONVideoEndToEnd(t *testing.T) {
	polls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/submit":
			fmt.Fprint(w, `{"data":{"task_id":"t1"}}`)
		case "/poll/t1":
			polls++
			if polls < 2 {
				fmt.Fprint(w, `{"data":{"status":"running"}}`)
			} else {
				fmt.Fprint(w, `{"data":{"status":"succeeded","video_url":"https://cdn.example.com/v.mp4"}}`)
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	rules := `{"submit":{"url":"/submit"},"poll":{"url":"/poll/{{task_id}}","interval_ms":10,"timeout_ms":5000},"response":{"task_id":"data.task_id","video_url":"data.video_url","status":"data.status","done_values":["succeeded"]}}`
	m := &models.APIModelWithProvider{}
	m.APIUrl = srv.URL
	m.APIModel.Model = "v1"
	m.APIFormat = "custom_json"
	m.APIModel.RequestRules = strPtr(rules)

	adapter := ResolveVideo(m)
	if adapter == nil {
		t.Fatal("expected a video adapter")
	}
	res, err := adapter.Generate(context.Background(), m, VideoRequest{Prompt: "a cat"})
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}
	if res.URL != "https://cdn.example.com/v.mp4" || res.TaskID != "t1" {
		t.Fatalf("unexpected result: %+v", res)
	}
}
