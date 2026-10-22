package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/aitokens"
	"github.com/xingmang/xingmang-desktop/internal/database"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// probeUpstream 是探测接口的假上游：按 limit 决定接受还是拒绝，
// 并记录收到的 max_tokens 序列（用来验证起点与降档行为）。
type probeUpstream struct {
	limit  int
	status int
	body   string

	mu       sync.Mutex
	received []int
}

func (p *probeUpstream) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			MaxTokens   int      `json:"max_tokens"`
			MaxOutput   int      `json:"max_output_tokens"`
			Temperature *float64 `json:"temperature"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		mt := body.MaxTokens
		if mt == 0 {
			mt = body.MaxOutput
		}

		p.mu.Lock()
		p.received = append(p.received, mt)
		p.mu.Unlock()

		if p.limit > 0 && mt > p.limit {
			w.WriteHeader(p.status)
			_, _ = w.Write([]byte(p.body))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}],` +
			`"usage":{"prompt_tokens":2,"completion_tokens":2}}`))
	}
}

func (p *probeUpstream) calls() []int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]int(nil), p.received...)
}

// setupProbeHandler 建库并插入一条渠道，返回 handler 与渠道 id。
func setupProbeHandler(t *testing.T, upstreamURL string) (*Handler, int) {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "probe_api.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	res, err := db.Exec(`INSERT INTO api_providers
		(name, provider_type, api_key, api_url, api_format, use_full_url, auth_type)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"probe", "other", "k", upstreamURL, "openai", 0, "api_key")
	if err != nil {
		t.Fatalf("insert provider: %v", err)
	}
	pid, _ := res.LastInsertId()
	aitokens.Init(db)
	return NewHandler(db), int(pid)
}

func callProbe(t *testing.T, h *Handler, payload map[string]interface{}) ProbeResponse {
	t.Helper()
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/ai/probe", strings.NewReader(string(raw)))
	rec := httptest.NewRecorder()
	h.ProbeModel(rec, req)

	var env struct {
		Success bool          `json:"success"`
		Data    ProbeResponse `json:"data"`
		Message string        `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("解析响应失败: %v / %s", err, rec.Body.String())
	}
	if !env.Success {
		t.Fatalf("接口返回失败: %s", env.Message)
	}
	return env.Data
}

// 声明值 100000、真实只接受 8192：必须从声明值起步、逐档降到 8192，
// 并把 8192 记成 accepted（不是"已验证"）。
func TestProbeModel_CapDowngradesAndAccepts(t *testing.T) {
	fake := &probeUpstream{
		limit:  8192,
		status: 400,
		body: `{"error":{"message":"This model's maximum context length is 8192 tokens. ` +
			`However, you requested 100000 tokens (5 in the messages, 99995 in the completion)."}}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, pid := setupProbeHandler(t, srv.URL)
	got := callProbe(t, h, map[string]interface{}{
		"provider_id": pid, "model": "deepseek-v4-pro",
		"config_type": "chat", "kind": "cap", "declared_max": 100000,
	})

	if !got.Reachable || got.Accepted != 8192 {
		t.Fatalf("应接受 8192，实得 reachable=%v accepted=%d msg=%s", got.Reachable, got.Accepted, got.Message)
	}
	if got.Conclusion != "ok" {
		t.Fatalf("结论应为 ok，实得 %s", got.Conclusion)
	}
	calls := fake.calls()
	if len(calls) < 2 || calls[0] != 100000 {
		t.Fatalf("应从声明值 100000 起步并降档，实得 %v", calls)
	}
	// 成功值要能被生产请求复用（同 key）。
	key := aitokens.CacheKey(&models.APIModelWithProvider{
		APIModel:  models.APIModel{ProviderID: pid, Model: "deepseek-v4-pro"},
		APIUrl:    srv.URL,
		APIFormat: "openai",
	})
	if v, ok := aitokens.Cached(key); !ok || v != 8192 {
		t.Fatalf("成功值未写入缓存: %d %v", v, ok)
	}
}

// 额度不足（预扣失败）不能降档：否则会把"余额不够"记成"模型不支持大值"。
func TestProbeModel_QuotaDoesNotDowngrade(t *testing.T) {
	fake := &probeUpstream{
		limit:  1,
		status: 403,
		body:   `{"error":{"message":"insufficient quota: pre-consumed amount exceeds balance"}}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, pid := setupProbeHandler(t, srv.URL)
	got := callProbe(t, h, map[string]interface{}{
		"provider_id": pid, "model": "glm-5.3",
		"config_type": "chat", "kind": "cap", "declared_max": 131072,
	})

	if got.Conclusion != "quota" {
		t.Fatalf("应判为 quota，实得 %s (%s)", got.Conclusion, got.Message)
	}
	if got.Accepted != 0 || got.Reachable {
		t.Fatalf("额度失败不应判为可用: %+v", got)
	}
	if n := len(fake.calls()); n != 1 {
		t.Fatalf("额度失败不得降档重试，实际打了 %d 发", n)
	}
}

// 非对话模型不探测：语义不同，且白名单靠行为/类型而不是模型名。
func TestProbeModel_NonChatUnsupported(t *testing.T) {
	fake := &probeUpstream{limit: 0}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, pid := setupProbeHandler(t, srv.URL)
	got := callProbe(t, h, map[string]interface{}{
		"provider_id": pid, "model": "gpt-image-1",
		"config_type": "image", "kind": "cap",
	})

	if got.Conclusion != "unsupported" {
		t.Fatalf("应判 unsupported，实得 %s", got.Conclusion)
	}
	if n := len(fake.calls()); n != 0 {
		t.Fatalf("不应发请求，实际 %d 发", n)
	}
}

// connect 只验连通，不产生 accepted —— 界面不能据此显示"可填 N"。
func TestProbeModel_ConnectOnly(t *testing.T) {
	fake := &probeUpstream{limit: 0}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, pid := setupProbeHandler(t, srv.URL)
	got := callProbe(t, h, map[string]interface{}{
		"provider_id": pid, "model": "gpt-4.1",
		"config_type": "chat", "kind": "connect",
	})

	if !got.Reachable || got.Accepted != 0 {
		t.Fatalf("连通测试不应给出上限: %+v", got)
	}
	if calls := fake.calls(); len(calls) != 1 || calls[0] != probeConnectTokens {
		t.Fatalf("连通测试应只打 1 发 %d，实得 %v", probeConnectTokens, calls)
	}
}

// 探测请求不得继承生产参数（temperature 等）。
func TestProbeModel_BodyIsMinimal(t *testing.T) {
	var seen map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&seen)
		fmt.Fprint(w, `{"choices":[{"message":{"content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1}}`)
	}))
	defer srv.Close()

	h, pid := setupProbeHandler(t, srv.URL)
	callProbe(t, h, map[string]interface{}{
		"provider_id": pid, "model": "m", "config_type": "chat", "kind": "connect",
	})

	for _, k := range []string{"temperature", "top_p", "frequency_penalty", "stop"} {
		if _, ok := seen[k]; ok {
			t.Fatalf("探测请求不应带 %s: %v", k, seen)
		}
	}
}

// 中转常把"模型不存在"包成 400（实测："Upstream request failed: Model is unavailable."）。
// 必须认成 notfound，前端才会把剩余模型整体跳过，而不是逐个打一遍。
func TestProbeStepKind_ModelUnavailable(t *testing.T) {
	body := `{"error":{"type":"server_error","message":"Upstream request failed: Model is unavailable."}}`
	if got := probeStepKind(400, body); got != "notfound" {
		t.Fatalf("应判 notfound，实得 %s", got)
	}
	if got := probeStepKind(404, `{}`); got != "notfound" {
		t.Fatalf("404 应判 notfound，实得 %s", got)
	}
}
