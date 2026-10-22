package aitokens

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/database"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

func model(name, mode string, maxTokens int) *models.APIModelWithProvider {
	return &models.APIModelWithProvider{
		APIModel: models.APIModel{
			ID:            1,
			ProviderID:    7,
			Name:          name,
			Model:         name,
			MaxTokens:     maxTokens,
			MaxTokensMode: mode,
		},
		ProviderName: "test",
	}
}

// ---------------------------------------------------------------------------
// 家族推断
// ---------------------------------------------------------------------------

func TestInferFamily(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"deepseek-v4.1-flash", "deepseek"},
		{"[JY]deepseek-v4-pro-0813", "deepseek"},
		{"glm-5.3-flash", "glm"},
		{"chatglm-4", "glm"},
		{"claude-opus-4-6", "claude"},
		{"[5]claude-opus-4-6-thinking", "claude"},
		{"gpt-5.6-luna", "gpt"},
		{"gemini-3-flash-preview", "gemini"},
		{"beam-free", "other"},
		{"omen-alpha", "other"},
		{"longcat-2.5-preview-free", "other"},
	}
	for _, c := range cases {
		if got := InferFamily(model(c.model, "auto", 0)); got != c.want {
			t.Errorf("InferFamily(%q) = %q, want %q", c.model, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// 候选序列
// ---------------------------------------------------------------------------

// ⚠ 关键口径：本项目用 `fixed`/`auto`，蓝本用 `manual`/`auto` —— 两套都必须认。
// 早期照蓝本写成 `!= "manual"`，会把本项目的 `fixed` 误判成自适应，
// 等于用户选了固定值仍被探测覆盖。
func TestIsAuto_BothVocabularies(t *testing.T) {
	cases := []struct {
		mode string
		want bool
	}{
		{"auto", true},
		{"AUTO", true},
		{" auto ", true},
		{"", true},        // 未设置 → 按 auto（与建表默认值一致）
		{"fixed", false},  // ← 本项目口径
		{"manual", false}, // ← 蓝本口径
		{"FIXED", false},
		{"whatever", false}, // 未知值按固定值处理，不擅自改动用户意图
	}
	for _, c := range cases {
		if got := IsAuto(model("m", c.mode, 100)); got != c.want {
			t.Errorf("IsAuto(mode=%q) = %v, want %v", c.mode, got, c.want)
		}
	}
}

// 固定值模式（fixed / manual）必须只返回一个值 —— 用户指定了上限就不该被探测覆盖。
func TestCandidates_FixedModeSingleValue(t *testing.T) {
	for _, mode := range []string{"fixed", "manual"} {
		got := Candidates(model("whatever", mode, 7777), 0)
		if len(got) != 1 || got[0] != 7777 {
			t.Fatalf("%s 模式应只返回配置值，实得 %v", mode, got)
		}
	}
}

// ⚠ 固定值模式但配置为空：返回 [omitTokens]，即**不发送 max_tokens 字段**。
//
// 不再编一个兜底数字。历史上这里散着 2000 / 4000 / 4096 三个互不相同的兜底值，
// 没有一个有据可依，而且偏小时会**静默截断输出** —— 不传让上游自己决定更安全。
func TestCandidates_FixedModeEmptyOmitsField(t *testing.T) {
	for _, mode := range []string{"fixed", "manual", "whatever"} {
		got := Candidates(model("whatever", mode, 0), 0)
		if len(got) != 1 || got[0] != omitTokens {
			t.Fatalf("%s 空配置应返回 [omitTokens]，实得 %v", mode, got)
		}
	}
}

// nil 配置同样只返回"不发送"，不能 panic。
func TestCandidates_NilConfigOmits(t *testing.T) {
	got := Candidates(nil, 0)
	if len(got) != 1 || got[0] != omitTokens {
		t.Fatalf("nil 配置应返回 [omitTokens]，实得 %v", got)
	}
}

// auto 模式：起点取家族上限，随后严格递减。
func TestCandidates_AutoStartsAtFamilyMax(t *testing.T) {
	got := Candidates(model("deepseek-v4-pro", "auto", 8192), 0)
	if len(got) == 0 {
		t.Fatal("候选不应为空")
	}
	if got[0] != familyMax["deepseek"] {
		t.Fatalf("deepseek 起点应为 %d，实得 %d", familyMax["deepseek"], got[0])
	}
	for i := 1; i < len(got); i++ {
		if got[i] >= got[i-1] {
			t.Fatalf("候选必须严格递减，位置 %d: %v", i, got)
		}
	}
	// 必须覆盖到 8192 —— 起点很高、真实上限很低的模型要靠它降下来
	if !contains(got, 8192) {
		t.Fatalf("候选应包含 8192，实得 %v", got)
	}
}

// ⚠ 这条是本次改动的核心回归：auto 模式下**不能**把配置里那个可能过时的
// max_tokens 当作上限。历史 bug 就是表里写死 8192（真实 384K），
// 若拿它当上限就永远探不到真实能力。
func TestCandidates_AutoIgnoresStaleConfiguredValue(t *testing.T) {
	got := Candidates(model("deepseek-v4.1-flash", "auto", 8192), 0)
	if got[0] == 8192 {
		t.Fatalf("auto 模式不应被过时的配置值(8192)封顶，实得起点 %d", got[0])
	}
	if got[0] != familyMax["deepseek"] {
		t.Fatalf("起点应为家族上限 %d，实得 %d", familyMax["deepseek"], got[0])
	}
}

// 空 mode 视同 auto（新导入的模型默认 auto，且 auto 是更安全的一侧）。
func TestCandidates_EmptyModeTreatedAsAuto(t *testing.T) {
	got := Candidates(model("glm-5.3-flash", "", 4096), 0)
	if got[0] != familyMax["glm"] {
		t.Fatalf("空 mode 应按 auto 处理，起点应为 %d，实得 %d", familyMax["glm"], got[0])
	}
}

// 调用方显式指定的上限更保守时，以它为准（长输出场景由调用方决定）。
func TestCandidates_HintCapsStart(t *testing.T) {
	got := Candidates(model("deepseek-v4-pro", "auto", 8192), 20000)
	if got[0] != 20000 {
		t.Fatalf("显式 hint 应作为起点，实得 %d", got[0])
	}
	// hint 比家族上限还大时不被抬高（避免调用方误传导致必然失败）
	got = Candidates(model("deepseek-v4-pro", "auto", 0), 999999)
	if got[0] != familyMax["deepseek"] {
		t.Fatalf("过大的 hint 应被家族上限压制，实得 %d", got[0])
	}
}

// 档位数量必须有上限，否则"起点与真实上限差距极大"时会打出一串请求。
func TestCandidates_BoundedAttempts(t *testing.T) {
	got := Candidates(model("deepseek-v4-pro", "auto", 0), 0)
	if len(got) > maxAttempts {
		t.Fatalf("候选数应 <= %d，实得 %d: %v", maxAttempts, len(got), got)
	}
}

// ---------------------------------------------------------------------------
// 错误判定
// ---------------------------------------------------------------------------

func TestIsLimitError(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   bool
	}{
		{"413 无歧义", 413, "anything", true},
		{"400 英文 token 超限", 400, `{"error":{"message":"This model's maximum context length is 8192 tokens"}}`, true},
		{"400 max_tokens 超范围", 400, `{"error":"max_tokens must be between 1 and 8192"}`, true},
		{"400 中文上下文超限", 400, `{"error":"上下文长度超出限制"}`, true},
		{"400 中文 token 过长", 400, `{"error":"token 过长"}`, true},
		{"400 输出过长", 400, `{"error":"output is too long"}`, true},
		{"400 提示过长", 400, `{"error":"prompt is too long"}`, true},
		// ⚠ 下面这些必须为 false：重试没有意义，只会把一次失败拖成 N 次
		{"401 鉴权失败", 401, `{"error":"invalid api key"}`, false},
		{"403 无权限", 403, `{"error":"forbidden"}`, false},
		{"400 参数缺失", 400, `{"error":"model is required"}`, false},
		{"400 格式错误", 400, `{"error":"invalid request body"}`, false},
		{"429 限流", 429, `{"error":"rate limit exceeded"}`, false},
		{"500 上游故障", 500, `{"error":"internal server error"}`, false},
		{"503 不可用", 503, `{"error":"service unavailable"}`, false},
		{"200 成功", 200, `{}`, false},
	}
	for _, c := range cases {
		if got := IsLimitError(c.status, c.body); got != c.want {
			t.Errorf("%s: IsLimitError(%d, %q) = %v, want %v", c.name, c.status, c.body, got, c.want)
		}
	}
}

// 5xx 即使文案里带 token 也不该降档 —— 那是上游故障，降档解决不了。
func TestIsLimitError_ServerErrorNotRetried(t *testing.T) {
	if IsLimitError(502, "upstream token service unavailable") {
		t.Fatal("5xx 不应被判为可重试的 token 超限")
	}
}

// ---------------------------------------------------------------------------
// 缓存
// ---------------------------------------------------------------------------

func TestCache_MemoryRoundTrip(t *testing.T) {
	Init(nil) // 无数据库：退化为纯内存
	key := "1::7::deepseek-v4-pro"
	if _, ok := Cached(key); ok {
		t.Fatal("初始不应命中")
	}
	Remember(key, 131072)
	if v, ok := Cached(key); !ok || v != 131072 {
		t.Fatalf("应命中 131072，实得 (%d, %v)", v, ok)
	}
	// 非法值不入缓存
	Remember("k2", 0)
	if _, ok := Cached("k2"); ok {
		t.Fatal("0 不应被缓存")
	}
}

// 缓存命中时起点直接取缓存值 —— 省掉一次必然失败的 400 往返。
func TestCandidates_CacheLowersStart(t *testing.T) {
	Init(nil)
	m := model("deepseek-v4-pro", "auto", 0)
	Remember(CacheKey(m), 32000)

	got := Candidates(m, 0)
	if got[0] != 32000 {
		t.Fatalf("缓存命中应以 32000 为起点，实得 %d", got[0])
	}
}

// 缓存值比家族上限还大时不采用 —— 那说明上游能力提升了，值得重新探一遍。
func TestCandidates_CacheAboveFamilyMaxIgnored(t *testing.T) {
	Init(nil)
	m := model("gemini-3-flash", "auto", 0)
	Remember(CacheKey(m), 999999)

	got := Candidates(m, 0)
	if got[0] != familyMax["gemini"] {
		t.Fatalf("超出家族上限的缓存值不应作起点，实得 %d", got[0])
	}
}

// 缓存键必须随 模型串 / 渠道 变化 —— 否则换渠道后会复用错误的探测结果。
func TestCacheKey_ChangesWithIdentity(t *testing.T) {
	a := model("deepseek-v4-pro", "auto", 0)
	b := model("deepseek-v4-pro", "auto", 0)
	b.ProviderID = 99
	c := model("glm-5.3", "auto", 0)

	if CacheKey(a) == CacheKey(b) {
		t.Fatal("渠道不同时缓存键应不同")
	}
	if CacheKey(a) == CacheKey(c) {
		t.Fatal("模型不同时缓存键应不同")
	}
}

// 落库往返：写入后新建实例应能读回（模拟重启后重新加载）。
func TestCache_PersistsToDatabase(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	Init(db)
	key := "1::7::deepseek-v4-pro"
	Remember(key, 393216)

	// 重新 Init 一个新的 Cache 实例，强制从库里加载
	Init(db)
	if v, ok := Cached(key); !ok || v != 393216 {
		t.Fatalf("重启后应能从库里读回 393216，实得 (%d, %v)", v, ok)
	}

	// 覆盖写：只保留最近一次可行值
	Remember(key, 131072)
	Init(db)
	if v, ok := Cached(key); !ok || v != 131072 {
		t.Fatalf("应被覆盖为 131072，实得 (%d, %v)", v, ok)
	}

	// 表结构与 All() 快照
	all := All()
	if all[key] != 131072 {
		t.Fatalf("All() 快照不正确: %v", all)
	}
}

// 未调用 Init 时所有操作都必须安全（no-op），不能 panic。
// 它挂在请求热路径上，测试或早期启动阶段可能尚未初始化。
func TestCache_UninitializedIsSafe(t *testing.T) {
	defMu.Lock()
	saved := defCache
	defCache = nil
	defMu.Unlock()
	defer func() {
		defMu.Lock()
		defCache = saved
		defMu.Unlock()
	}()

	if _, ok := Cached("x"); ok {
		t.Fatal("未初始化不应命中")
	}
	Remember("x", 100) // 不应 panic
	if len(All()) != 0 {
		t.Fatal("未初始化时 All() 应为空")
	}
}

// ---------------------------------------------------------------------------
// 失败分类：输出超限 vs 上下文超限
// ---------------------------------------------------------------------------

// ⚠ 这两类必须分开。输出超限说明"模型能力就这样"，可以沉淀成缓存；
// 上下文超限是"本次输入太长"，换句短话本来就能用大值 —— 缓存它等于把偶发当永久。
func TestClassify_OutputVsContext(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   FailureKind
	}{
		// —— 输出侧 ——
		{"max_tokens 超范围", 400, `{"error":"max_tokens must be between 1 and 8192"}`, FailureOutputLimit},
		{"输出过长", 400, `{"error":"output is too long"}`, FailureOutputLimit},
		{"Gemini maxOutputTokens", 400,
			`{"error":{"status":"INVALID_ARGUMENT","message":"maxOutputTokens must be <= 65536"}}`, FailureOutputLimit},
		// ⚠ OpenAI 的 max_tokens 过大用的是**上下文口径**的文案，必须靠
		// "in the completion" 认出来，否则会永远不敢写缓存。
		{"OpenAI max_tokens 过大", 400,
			`{"error":{"message":"This model's maximum context length is 8192 tokens. However, you requested 20000 tokens (10 in the messages, 19990 in the completion)."}}`,
			FailureOutputLimit},

		// —— 输入侧 ——
		{"提示词过长", 400, `{"error":"prompt is too long: 200000 tokens > 128000 maximum"}`, FailureContextLimit},
		{"messages 结果超窗", 400,
			`{"error":{"message":"This model's maximum context length is 128000 tokens. However, your messages resulted in 130000 tokens."}}`,
			FailureContextLimit},
		{"413 请求体过大", 413, `{"error":"payload too large"}`, FailureContextLimit},
		{"中文上下文超限", 400, `{"error":"上下文长度超出限制"}`, FailureContextLimit},

		// —— 与长度无关：不允许降档，也不允许写缓存 ——
		{"401 鉴权", 401, `{"error":"invalid api key"}`, FailureOther},
		{"400 参数缺失", 400, `{"error":"model is required"}`, FailureOther},
		{"400 格式错误", 400, `{"error":"invalid request body"}`, FailureOther},
		{"429 限流", 429, `{"error":"rate limit exceeded"}`, FailureOther},
		{"500 上游故障", 500, `{"error":"internal server error"}`, FailureOther},
		{"502 文案带 token 也是故障", 502, "upstream token service unavailable", FailureOther},
		{"200 成功", 200, `{}`, FailureOther},
	}
	for _, c := range cases {
		if got := Classify(c.status, c.body); got != c.want {
			t.Errorf("%s: Classify(%d) = %v, want %v", c.name, c.status, got, c.want)
		}
	}
}

// 只有输出超限允许写缓存。
func TestCanCacheAfter(t *testing.T) {
	if !CanCacheAfter(FailureOutputLimit) {
		t.Fatal("输出超限应允许写缓存")
	}
	if CanCacheAfter(FailureContextLimit) {
		t.Fatal("上下文超限不得写缓存（那是本次输入太长，不是模型能力）")
	}
	if CanCacheAfter(FailureOther) {
		t.Fatal("无关错误不得写缓存")
	}
}

// ---------------------------------------------------------------------------
// 请求体写入
// ---------------------------------------------------------------------------

// 值为 omitTokens 时：OpenAI 格式删除该键，Anthropic 格式退回家族起点（必填）。
func TestApplyToBody(t *testing.T) {
	cfg := model("deepseek-v4-pro", "auto", 0)
	cfg.APIFormat = "openai"

	body := map[string]interface{}{"model": "x", "max_tokens": 12345}
	if got := ApplyToBody(body, omitTokens, cfg); got != omitTokens {
		t.Fatalf("openai 格式遇 0 应返回 %d，实得 %d", omitTokens, got)
	}
	if _, exists := body["max_tokens"]; exists {
		t.Fatal("openai 格式遇 0 应删除 max_tokens 键")
	}

	// Anthropic 必填，不能删，要退回 FamilyStart
	cfg.APIFormat = "anthropic"
	body = map[string]interface{}{"model": "x"}
	got := ApplyToBody(body, omitTokens, cfg)
	if got != familyMax["deepseek"] {
		t.Fatalf("anthropic 遇 0 应用家族起点 %d 兜底，实得 %d", familyMax["deepseek"], got)
	}
	if body["max_tokens"] != familyMax["deepseek"] {
		t.Fatalf("anthropic 请求体应写入兜底值，实得 %v", body["max_tokens"])
	}

	// 正常值原样写入
	body = map[string]interface{}{}
	if got := ApplyToBody(body, 65536, cfg); got != 65536 || body["max_tokens"] != 65536 {
		t.Fatalf("正常值应原样写入，实得 %d / %v", got, body["max_tokens"])
	}
}

// ---------------------------------------------------------------------------
// 缓存衰减（ε-greedy）
// ---------------------------------------------------------------------------

// ⚠ 修正「缓存只降不升」：TTL 过期后必须忽略缓存、从家族起点重探。
// 否则一次因上下文占满而失败，会把值永久压低 —— 之后哪怕只发一句话也上不去。
func TestReprobeDue_AfterTTL(t *testing.T) {
	Init(nil)
	m := model("deepseek-v4-pro", "auto", 0)
	key := CacheKey(m)
	Remember(key, 32000)

	if got := Candidates(m, 0); got[0] != 32000 {
		t.Fatalf("未过期应以缓存为起点，实得 %d", got[0])
	}

	// 把验证时间改成已过期
	c := current()
	c.mu.Lock()
	e := c.mem[key]
	e.verifiedAt = time.Now().Add(-cacheTTL - time.Hour)
	c.mem[key] = e
	c.mu.Unlock()

	if got := Candidates(m, 0); got[0] != familyMax["deepseek"] {
		t.Fatalf("TTL 过期后应回到家族起点 %d，实得 %d", familyMax["deepseek"], got[0])
	}
}

// 每 probeEvery 次使用缓存，允许一次重探 —— 防「TTL 未到但上游能力已提升」。
func TestReprobeDue_EveryN(t *testing.T) {
	Init(nil)
	m := model("deepseek-v4-pro", "auto", 0)
	key := CacheKey(m)
	Remember(key, 32000)

	for i := 1; i < probeEvery; i++ {
		if got := Candidates(m, 0); got[0] != 32000 {
			t.Fatalf("第 %d 次应以缓存为起点，实得 %d", i, got[0])
		}
	}
	if got := Candidates(m, 0); got[0] != familyMax["deepseek"] {
		t.Fatalf("第 %d 次应触发重探回到家族起点，实得 %d", probeEvery, got[0])
	}
}

// 没有缓存时不谈"重探"，也不该报 true。
func TestReprobeDue_NoCache(t *testing.T) {
	Init(nil)
	if ReprobeDue("no-such-key") {
		t.Fatal("无缓存时不应判为需要重探")
	}
	if ReprobeDue("") {
		t.Fatal("空键不应判为需要重探")
	}
}

func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// 换地址 = 换链路。缓存键必须随 api_url / api_format 变化，
// 否则用户改渠道后会把 A 站的上限套到 B 站上。
func TestCacheKey_ChangesWithEndpoint(t *testing.T) {
	a := model("glm-5.3", "auto", 0)
	a.APIUrl = "https://a.example.com/v1"
	a.APIFormat = "openai"
	b := model("glm-5.3", "auto", 0)
	b.APIUrl = "https://b.example.com/v1"
	b.APIFormat = "openai"
	c := model("glm-5.3", "auto", 0)
	c.APIUrl = "https://a.example.com/v1"
	c.APIFormat = "anthropic"

	if CacheKey(a) == CacheKey(b) {
		t.Fatal("地址不同时缓存键应不同")
	}
	if CacheKey(a) == CacheKey(c) {
		t.Fatal("协议不同时缓存键应不同")
	}
}

// 上游 /models 声明过输出上限时，以它作为探测起点 —— 比按模型名猜家族更准。
func TestCandidates_UpstreamDeclaredIsStart(t *testing.T) {
	Init(nil)
	m := model("glm-5.3", "auto", 200000)
	m.MaxTokensSource = "upstream"

	if got := Candidates(m, 0)[0]; got != 200000 {
		t.Fatalf("应以声明的 200000 起步，实得 %d", got)
	}
}

// 本地表（table）的值不能当起点：历史上写过 4K/16K 这类严重偏低的值，
// 拿它当起点会把探测永久压在低档，反而探不出真实能力。
func TestCandidates_TableSourceNotUsedAsStart(t *testing.T) {
	Init(nil)
	m := model("glm-5.3", "auto", 4096)
	m.MaxTokensSource = "table"

	if got := Candidates(m, 0)[0]; got != familyMax["glm"] {
		t.Fatalf("table 来源不应作为起点，应回到家族起点 %d，实得 %d", familyMax["glm"], got)
	}
}

// 用户点"重测"必须真的从高处重走：Fresh 版本忽略缓存。
func TestCandidatesFresh_IgnoresCache(t *testing.T) {
	Init(nil)
	m := model("deepseek-v4-pro", "auto", 0)
	m.APIUrl = "https://x.example.com/v1"
	Remember(CacheKey(m), 8192)

	if got := Candidates(m, 0)[0]; got != 8192 {
		t.Fatalf("普通模式应命中缓存 8192，实得 %d", got)
	}
	if got := CandidatesFresh(m, 0)[0]; got != familyMax["deepseek"] {
		t.Fatalf("Fresh 模式应忽略缓存，从 %d 起步，实得 %d", familyMax["deepseek"], got)
	}
}
