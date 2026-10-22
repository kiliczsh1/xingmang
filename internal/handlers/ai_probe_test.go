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

	"github.com/xingmang/xingmang-desktop/internal/database"
)

// fakeUpstream 是一个可控的假上游，用来验证「自适应 max_tokens」的探测行为。
//
// limit 是它愿意接受的 max_tokens 上限；超过就按 errStatus/errBody 拒绝。
// 它记录每一次收到的 max_tokens，测试据此断言探测序列。
type fakeUpstream struct {
	limit    int
	errStatus int
	errBody  string

	mu       sync.Mutex
	received []int
}

func (f *fakeUpstream) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			MaxTokens int `json:"max_tokens"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		f.mu.Lock()
		f.received = append(f.received, body.MaxTokens)
		f.mu.Unlock()

		if f.limit > 0 && body.MaxTokens > f.limit {
			w.WriteHeader(f.errStatus)
			_, _ = w.Write([]byte(f.errBody))
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"ok\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}
}

func (f *fakeUpstream) calls() []int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int(nil), f.received...)
}

// setupProbeTest 建一个只含一条模型的库，指向给定的假上游。
func setupProbeTest(t *testing.T, upstreamURL, modelName, mode string, maxTokens int) (*Handler, int) {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "probe.db"))
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

	res, err = db.Exec(`INSERT INTO api_models
		(provider_id, name, model, config_type, max_tokens, max_tokens_mode, enabled, is_default, temperature, top_p)
		VALUES (?, ?, ?, ?, ?, ?, 1, 1, 0.7, 0.9)`,
		pid, modelName, modelName, "chat", maxTokens, mode)
	if err != nil {
		t.Fatalf("insert model: %v", err)
	}
	mid, _ := res.LastInsertId()

	return NewHandler(db), int(mid)
}

func chatOnce(t *testing.T, h *Handler, modelID int) *httptest.ResponseRecorder {
	t.Helper()
	body := fmt.Sprintf(`{"configId":%d,"messages":[{"role":"user","content":"hi"}]}`, modelID)
	req := httptest.NewRequest("POST", "/api/ai/chat", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.AIChat(rec, req)
	return rec
}

// ⚠ 核心回归：auto 模式下，上游拒绝过大的 max_tokens 时必须**逐档降下来**并最终成功。
//
// 这条同时钉住三件事：起点不被过时的配置值(8192)封顶、降档真的发生、成功值被缓存。
//
// errBody 用的是 OpenAI 真实的"max_tokens 过大"文案（注意它是**上下文口径**的，
// 必须靠 "in the completion" 认出是输出侧问题，否则不会被缓存 —— 见
// aitokens.Classify 的注释）。
func TestAIChat_AutoProbesDownUntilAccepted(t *testing.T) {
	fake := &fakeUpstream{
		limit:     8192,
		errStatus: 400,
		errBody: `{"error":{"message":"This model's maximum context length is 8192 tokens. ` +
			`However, you requested 393216 tokens (5 in the messages, 393211 in the completion)."}}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, mid := setupProbeTest(t, srv.URL, "deepseek-v4-pro", "auto", 8192)
	rec := chatOnce(t, h, mid)

	if rec.Code != 200 {
		t.Fatalf("最终应成功，实得 %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("响应应含上游内容，实得 %q", rec.Body.String())
	}

	got := fake.calls()
	// 起点必须是家族上限（393216），而不是配置里那个过时的 8192
	if len(got) == 0 || got[0] != 393216 {
		t.Fatalf("首次应发家族上限 393216，实得 %v", got)
	}
	// 必须严格递减，且最后一发就是上游能接受的那个值
	for i := 1; i < len(got); i++ {
		if got[i] >= got[i-1] {
			t.Fatalf("探测序列必须递减，实得 %v", got)
		}
	}
	if last := got[len(got)-1]; last != 8192 {
		t.Fatalf("应最终落在 8192，实得 %v", got)
	}
	if len(got) < 2 {
		t.Fatalf("应发生过降档重试，实得只有 %v", got)
	}

	// 成功值必须被记住，供后续请求直接命中
	rows := 0
	if err := h.db.QueryRow(`SELECT count(*) FROM model_max_tokens`).Scan(&rows); err != nil {
		t.Fatalf("查询缓存表: %v", err)
	}
	if rows != 1 {
		t.Fatalf("探测成功应写入 1 条缓存，实得 %d", rows)
	}
	var cached int
	if err := h.db.QueryRow(`SELECT max_tokens FROM model_max_tokens`).Scan(&cached); err != nil {
		t.Fatalf("读缓存: %v", err)
	}
	if cached != 8192 {
		t.Fatalf("缓存值应为 8192，实得 %d", cached)
	}
}

// 缓存命中后应直接从上次成功的档位起步 —— 不再重复打那次必然失败的请求。
func TestAIChat_UsesCacheOnSubsequentCall(t *testing.T) {
	fake := &fakeUpstream{
		limit:     8192,
		errStatus: 400,
		errBody:   `{"error":{"message":"max_tokens must be less than or equal to 8192"}}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, mid := setupProbeTest(t, srv.URL, "deepseek-v4-pro", "auto", 8192)

	first := chatOnce(t, h, mid)
	if first.Code != 200 {
		t.Fatalf("首次应成功，实得 %d", first.Code)
	}
	firstCalls := len(fake.calls())

	second := chatOnce(t, h, mid)
	if second.Code != 200 {
		t.Fatalf("二次应成功，实得 %d", second.Code)
	}
	secondCalls := len(fake.calls()) - firstCalls

	if secondCalls != 1 {
		t.Fatalf("缓存命中后二次请求应只打 1 发，实得 %d 发", secondCalls)
	}
}

// ⚠ 上下文超限（输入侧）即使降档成功也**不得写缓存**。
//
// 那是"本次输入太长"造成的，换句短话本来就能用大值；把它沉淀成结论，
// 等于让一次偶发永久压低这个模型的输出上限 —— 正是「缓存只降不升」的病根。
//
// 代价是这类请求每次都要重新降档（本用例断言 2 发），这是刻意选择的保守侧。
func TestAIChat_ContextLimitNotCached(t *testing.T) {
	fake := &fakeUpstream{
		limit:     8192,
		errStatus: 400,
		// 明确指向输入侧的文案（"your messages resulted in"）
		errBody: `{"error":{"message":"This model's maximum context length is 8192 tokens. ` +
			`However, your messages resulted in 20000 tokens."}}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, mid := setupProbeTest(t, srv.URL, "deepseek-v4-pro", "auto", 8192)
	if rec := chatOnce(t, h, mid); rec.Code != 200 {
		t.Fatalf("应降档后成功，实得 %d", rec.Code)
	}

	rows := 0
	if err := h.db.QueryRow(`SELECT count(*) FROM model_max_tokens`).Scan(&rows); err != nil {
		t.Fatalf("查询缓存表: %v", err)
	}
	if rows != 0 {
		t.Fatalf("上下文超限的降档结果不得写缓存，实得 %d 条", rows)
	}

	// 二次请求必须重新降档（因为没缓存可复用）
	before := len(fake.calls())
	if rec := chatOnce(t, h, mid); rec.Code != 200 {
		t.Fatalf("二次也应成功，实得 %d", rec.Code)
	}
	if after := len(fake.calls()) - before; after < 2 {
		t.Fatalf("无缓存时应重新降档（>=2 发），实得 %d 发", after)
	}
}

// ⚠ 反向用例：非 token 超限的错误**绝不能**降档重试。
// 否则一次鉴权失败会被放大成 7 次无谓请求。
func TestAIChat_NonLimitErrorDoesNotRetry(t *testing.T) {
	fake := &fakeUpstream{
		limit:     100, // 必然触发"超限"分支
		errStatus: 401,
		errBody:   `{"error":{"message":"invalid api key"}}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, mid := setupProbeTest(t, srv.URL, "deepseek-v4-pro", "auto", 8192)
	rec := chatOnce(t, h, mid)

	if len(fake.calls()) != 1 {
		t.Fatalf("鉴权失败不应重试，实得 %d 发: %v", len(fake.calls()), fake.calls())
	}
	if !strings.Contains(rec.Body.String(), "401") {
		t.Fatalf("应把 401 原样报给前端，实得 %q", rec.Body.String())
	}
}

// fixed 模式：不探测，只发配置值，哪怕上游拒绝也不降档。
func TestAIChat_FixedModeSendsSingleValue(t *testing.T) {
	fake := &fakeUpstream{
		limit:     8192,
		errStatus: 400,
		errBody:   `{"error":"maximum context length"}`,
	}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	// fixed + 一个上游接受不了的值 → 必然失败，且只发一发
	h, mid := setupProbeTest(t, srv.URL, "deepseek-v4-pro", "fixed", 99999)
	_ = chatOnce(t, h, mid)

	got := fake.calls()
	if len(got) != 1 || got[0] != 99999 {
		t.Fatalf("fixed 模式应只发一次配置值 99999，实得 %v", got)
	}
}

// 探测成功值不应污染 fixed 模式（fixed 永远听配置的）。
func TestAIChat_FixedModeIgnoresCache(t *testing.T) {
	fake := &fakeUpstream{}
	srv := httptest.NewServer(fake.handler())
	defer srv.Close()

	h, mid := setupProbeTest(t, srv.URL, "deepseek-v4-pro", "fixed", 4321)
	_ = chatOnce(t, h, mid)

	got := fake.calls()
	if len(got) != 1 || got[0] != 4321 {
		t.Fatalf("fixed 应只发配置值 4321，实得 %v", got)
	}
}
