// Package aitokens 实现「请求时自适应 max_tokens」。
//
// 核心思想（移植自蓝本 xingnovel 的 `tokenLimitPolicy.ts` + `maxTokensCacheService.ts`）：
//
//	**不预设真值，而是从家族上限逐档下探，直到上游接受，并把成功值缓存下来。**
//
// 为什么不能靠一张静态表：各家的输出上限随版本迭代剧烈变化，而且网关型渠道
// （opencode / 枫叶这类）会自造别名，表永远覆盖不全。**表只能当"探测起点"，
// 一旦被当成"真值"写进数据库并展示给用户，就会变成静默的假数据** ——
// 这正是本项目踩过的坑（deepseek 写 16K、glm 写 4K，实际分别是 384K / 128K）。
//
// 起点猜错的代价只是「多一次 400 往返」：上游拒绝后自动降档重试，用户无感。
// 所以家族表粗一点无害，这与"当成真值"有本质区别。
//
// 三条边界：
//   - `max_tokens_mode = manual` → 完全不探测，只发配置里那一个值（用户说了算）；
//   - `auto` → 走探测；缓存命中则直接从上次成功的档位开始，省掉 400；
//   - 只把**明确的 token 超限错误**当作"继续降档"的信号，其余错误原样抛出，
//     避免把鉴权/格式错误也拿去重试（蓝本同样这么做）。
package aitokens

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ladder 是降序档位表：探测时按顺序取"比起点小"的项依次尝试。
//
// ⚠ 档位必须**近似减半**，不能密集。曾经把 2 的幂与十进制整值（131072/128000、
// 65536/64000…）混在一起想"少走一两档"，结果起点 393216 的模型要连降 8 次才到
// 16384，**真实上限 8192 的模型会 8 次全败直接报错** —— 档位越密，能覆盖的跨度越小。
// 现在按减半设计：从最高档降到 4096 只需 6 步，任何起点都能在 maxAttempts 内触底。
var ladder = []int{
	393216, 131072, 65536, 32768, 16384, 8192, 4096,
}

// familyMax 是各家族的**探测起点**（不是真值！）。
//
// 取值依据（均为已核实的官方数字，见 MEMORY.md#max_tokens 体系）：
//
//	deepseek 393216 ← 实抓 DeepSeek 官方 /v1/models 的 max_output_tokens
//	glm      131072 ← 智谱官方「最大输出 128K」
//	claude   128000 ← Anthropic 官方「1M 上下文模型单请求最多 128k 输出」
//	gpt      128000 ← OpenAI 官方
//	gemini    65536 ← Google 官方 output token limit
//	other    128000 ← 保守起点：几乎所有现代模型都接受 128K，超了会自动降档
//
// ⚠ 起点偏高是安全的（会自动降档），起点偏低才是问题（会白白浪费模型能力）。
// 所以这里宁可取大不取小。
var familyMax = map[string]int{
	"deepseek": 393216,
	"glm":      131072,
	"claude":   128000,
	"gpt":      128000,
	"gemini":   65536,
	"other":    128000,
}

// maxAttempts 限制一次请求最多尝试几档。
//
// 防的是"家族起点与真实上限差距极大"时打出一串请求（例如起点 393216、真实 4096
// 要连降 7 档）。8 档足以覆盖实际跨度，再降不到就该报错而不是继续打。
const maxAttempts = 8

// omitTokens 是候选序列里的哨兵：表示**本次不发送 max_tokens 字段**。
//
// 为什么需要它：max_tokens 在 OpenAI 兼容接口里是可选字段，不传 = 由上游按模型
// 自身能力决定，这比我们填一个偏小的值安全得多。项目里历史上散着五个互不相同的
// 兜底值（2000 / 4000 / 4096 / 8192 / 8192），全部无据可依，且偏小时会**静默截断输出**。
//
// Anthropic 格式是必填，调用方遇到 0 时应改用 FamilyStart 兜底
// （见 `ai.BuildRequestBody` 与 `agent.callModel`）。
const omitTokens = 0

// InferFamily 从模型名 / 显示名 / 渠道名推断家族。
// 只认几个明确的品牌词，其余归 other —— 因为结果只用于挑"起点"，粗糙无所谓。
func InferFamily(cfg *models.APIModelWithProvider) string {
	if cfg == nil {
		return "other"
	}
	hay := strings.ToLower(strings.Join([]string{cfg.Model, cfg.Name, cfg.ProviderName}, " "))
	switch {
	case strings.Contains(hay, "deepseek"):
		return "deepseek"
	case strings.Contains(hay, "glm") || strings.Contains(hay, "chatglm"):
		return "glm"
	case strings.Contains(hay, "claude"):
		return "claude"
	case strings.Contains(hay, "gpt") || strings.Contains(hay, "o1") || strings.Contains(hay, "o3"):
		return "gpt"
	case strings.Contains(hay, "gemini"):
		return "gemini"
	}
	return "other"
}

// FamilyStart 返回该模型家族的**探测起点**（不是真值）。
//
// 取值依据见 familyMax 的注释。除了探测起点，它还有第二个用途：
// Anthropic 等**必填** max_tokens 的格式在配置为空时，用它兜底
// —— 这是全项目唯一允许出现的「无配置兜底值」，不再是各处各写一个数。
func FamilyStart(cfg *models.APIModelWithProvider) int {
	if cfg == nil {
		return familyMax["other"]
	}
	if v := familyMax[InferFamily(cfg)]; v > 0 {
		return v
	}
	return familyMax["other"]
}

// startFor 选探测起点：**上游 /models 明确声明过**就用它，否则按家族起点。
//
// ⚠ 只认 source=upstream 的值。本地家族表（table）历史上写过一批严重偏低的值
// （deepseek 16K、glm 4K），拿它当起点会把探测永久压在低档 —— 起点偏低是本方案里
// 唯一"会真的损失能力"的错误，而起点偏高只多一次 400。
func startFor(cfg *models.APIModelWithProvider) int {
	if cfg != nil && cfg.MaxTokensSource == "upstream" && cfg.MaxTokens > 0 {
		return cfg.MaxTokens
	}
	return FamilyStart(cfg)
}

// IsAuto 判断该模型是否处于自适应模式。
//
// ⚠ 取值口径必须同时认两套：**本项目用 `fixed`/`auto`**（界面与历史数据），
// **蓝本用 `manual`/`auto`**。早期照蓝本写成 `!= "manual"`，会把本项目的 `fixed`
// 误判成自适应 —— 那等于用户明明选了固定值、却仍被探测覆盖，是静默错误。
// 判定改为「白名单命中才 auto」，未知值一律当固定值（更保守，不会擅自改动用户意图）。
//
// 空值按 auto 处理：新导入的模型默认就是 auto（后端建表默认值），
// 且 auto 是更安全的一侧 —— 固定值会把可能过时的数字直接发出去。
func IsAuto(cfg *models.APIModelWithProvider) bool {
	if cfg == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(cfg.MaxTokensMode)) {
	case "auto":
		return true
	case "":
		return true
	default:
		// fixed / manual / 其它未知值 → 都按"用户指定了固定值"处理
		return false
	}
}

// CacheKey 是缓存的键：渠道 + 模型串 + 地址指纹。
//
// 为什么**不含模型行 id**：
//   - 导入弹窗里要在"落库之前"探测，那时还没有 id；
//   - 用户删掉模型再重建同一条配置时，应该继承之前的结论，而不是从零重探。
//
// 为什么**含地址与协议**：换中转是最常见的操作，换了地址就是另一条链路了，
// 沿用旧值等于把 A 渠道的上限套到 B 渠道上。
func CacheKey(cfg *models.APIModelWithProvider) string {
	if cfg == nil {
		return ""
	}
	fp := strings.ToLower(strings.TrimSpace(cfg.APIUrl)) + "|" + strings.ToLower(strings.TrimSpace(cfg.APIFormat))
	sum := sha1.Sum([]byte(fp))
	return fmt.Sprintf("%d::%s::%s", cfg.ProviderID, strings.ToLower(strings.TrimSpace(cfg.Model)), hex.EncodeToString(sum[:8]))
}

// Candidates 生成从大到小的 max_tokens 尝试序列。
//
// 返回值里 **0 表示「本次不发送 max_tokens 字段」**（见 omitTokens）：
//   - auto  ：家族起点 → 逐档下探，全部为正数，不会出现 0；
//   - fixed ：只返回配置里那一个值；配置为空则返回 [0]（不发送）。
//
// hint 是调用方显式指定的上限（AIChatRequest.MaxTokens，用于长输出场景）；
// <=0 表示未指定。
func Candidates(cfg *models.APIModelWithProvider, hint int) []int {
	return candidates(cfg, hint, true)
}

// CandidatesFresh 与 Candidates 相同，但**忽略缓存**。
//
// 用于用户显式触发的探测：点了"测试/重测"，就必须真的从高处重新走一遍，
// 否则缓存会把结果按在原地，"再测一次"变成"再看一眼旧值"。
func CandidatesFresh(cfg *models.APIModelWithProvider, hint int) []int {
	return candidates(cfg, hint, false)
}

func candidates(cfg *models.APIModelWithProvider, hint int, useCache bool) []int {
	if cfg == nil {
		return []int{omitTokens}
	}

	if !IsAuto(cfg) {
		v := hint
		if v <= 0 {
			v = cfg.MaxTokens
		}
		if v <= 0 {
			// 用户选了固定值却没填数：没有值可发，就不发。
			// 补一个我们编的默认值，只会在它偏小时**静默截断输出**。
			return []int{omitTokens}
		}
		return []int{v}
	}

	start := startFor(cfg)

	// 缓存优先：上次已成功过的档位直接当起点，省掉必然失败的第一发。
	//
	// 只在缓存值更小时采用 —— 更大的值意味着上游能力提升了，值得重新探。
	// ⚠ 但这会让缓存**只降不升**：一次因上下文占满而失败就把值永久压低，
	// 之后哪怕只发一句话也上不去。所以额外用 ReprobeDue 做 ε-greedy 探索，
	// 偶尔忽略缓存、从家族起点重新来一次。
	if useCache {
		if cached, ok := Cached(CacheKey(cfg)); ok && cached > 0 && cached < start {
			if ReprobeDue(CacheKey(cfg)) {
				log.Printf("[aitokens] %s 触发重探（缓存 %d 已过期或到达重探周期），从家族起点 %d 重新开始",
					cfg.Model, cached, start)
			} else {
				start = cached
			}
		}
	}

	// 调用方显式要求的上限比家族起点更保守时以它为准（与蓝本一致）。
	if hint > 0 && hint < start {
		start = hint
	}

	out := []int{start}
	for _, s := range ladder {
		if s < start {
			out = append(out, s)
		}
	}
	// 起点可能落在两档之间（如 128000），去重并截断，避免重复尝试同一个值。
	seen := make(map[int]bool, len(out))
	uniq := out[:0]
	for _, v := range out {
		if v <= 0 || seen[v] {
			continue
		}
		seen[v] = true
		uniq = append(uniq, v)
		if len(uniq) >= maxAttempts {
			break
		}
	}
	return uniq
}

// outputKeywords 是"单次生成长度超限"的特征词。
var outputKeywords = []string{
	"max tokens",
	"max_tokens",
	"maxtokens",
	"max_completion_tokens",
	"maxcompletiontokens",
	"maxoutputtokens",
	"max output",
	"maximum output",
	"maximum tokens",
	"output token",
	"output is too long",
	"output limit",
}

// contextKeywords 是"上下文 / 提示词长度超限"的特征词。
var contextKeywords = []string{
	"context length",
	"maximum context",
	"max context",
	"context window",
	"prompt is too long",
	"input is too long",
	"input token",
	"too many tokens",
	"request too large",
	"reduce the length",
}

// limitSubjects / limitVerbs 用于中文文案的**复合**判定。
//
// 中文报错里主体词与超限词之间常夹空格或其它字（「token 过长」「上下文长度超出限制」），
// 把它们逐一拼成固定短语会漏 —— 曾经因此漏掉带空格的写法。改为"两者都出现即命中"。
var (
	limitSubjects = []string{"上下文", "token", "输出", "长度", "提示词", "输入"}
	limitVerbs    = []string{"超出", "超过", "过长", "太长", "最大", "上限"}
	// contextSubjects 是中文里指向**输入侧**的主体词，用于区分两种长度超限。
	contextSubjects = []string{"上下文", "提示词", "输入"}
)

// FailureKind 描述一次失败的性质。它决定两件事：
//   - 是否允许降档重试（只有两种长度类失败才允许）；
//   - **是否允许把最终成功的值写进缓存** —— 上下文超限属于"本次输入太长"，
//     换个短输入本来就能用大值，把它沉淀成结论等于把偶发当永久。
type FailureKind int

const (
	// FailureOther 与长度无关：鉴权失败、参数格式错误、上游故障。
	FailureOther FailureKind = iota
	// FailureOutputLimit 单次输出上限被拒 → 降档，成功值**可以**缓存。
	FailureOutputLimit
	// FailureContextLimit 输入+输出超出上下文窗口 → 降档，成功值**不可**缓存。
	FailureContextLimit
)

// String 便于日志阅读。
func (k FailureKind) String() string {
	switch k {
	case FailureOutputLimit:
		return "输出超限"
	case FailureContextLimit:
		return "上下文超限"
	default:
		return "其它错误"
	}
}

// Classify 判定一次失败的性质。
//
// 判据刻意保守：**只有明确的长度类错误才允许降档**。鉴权失败 / 参数格式错误
// 重试多少次都一样，只会把一次失败拖成 N 次。
//
// 上下文侧优先于输出侧：两者都命中时按"不可缓存"处理 —— 不缓存是更安全的一侧
// （宁可多探几次，也不要让一个偶发值永久锁死输出上限）。
func Classify(status int, body string) FailureKind {
	// 只有 4xx 才可能是"参数被拒"。5xx 是上游故障，降档没有意义。
	if status < 400 || status >= 500 {
		return FailureOther
	}
	lower := strings.ToLower(body)

	// 413 Request Entity Too Large：请求体过大，属输入侧。
	if status == 413 {
		return FailureContextLimit
	}

	// ⚠ OpenAI 在"max_tokens 超过模型输出上限"时，给的也是上下文口径的文案。
	// 靠这两个短句区分，否则会把输出超限误判成上下文超限、永远不敢写缓存，
	// 导致每次对话都要重打 6 发降档请求：
	//   max_tokens 过大 → "... you requested 20000 tokens (10 in the messages, 19990 in the completion)"
	//   提示词过长      → "... your messages resulted in 130000 tokens"
	if strings.Contains(lower, "messages resulted in") {
		return FailureContextLimit
	}
	if strings.Contains(lower, "in the completion") || strings.Contains(lower, "you requested") {
		return FailureOutputLimit
	}

	outputHit := containsAny(lower, outputKeywords)
	contextHit := containsAny(lower, contextKeywords)

	// "must be between X and Y" 这类范围校验，只在同时提到 token/max 时才算。
	if strings.Contains(lower, "must be between") &&
		(strings.Contains(lower, "token") || strings.Contains(lower, "max")) {
		outputHit = true
	}
	// exceed + token
	if strings.Contains(lower, "exceed") && strings.Contains(lower, "token") {
		outputHit = true
	}

	// 中文复合判定：主体词与超限词都出现才算（中间允许夹空格或其它字）
	if containsAny(lower, limitSubjects) && containsAny(lower, limitVerbs) {
		if containsAny(lower, contextSubjects) {
			contextHit = true
		} else {
			outputHit = true
		}
	}

	if contextHit {
		return FailureContextLimit
	}
	if outputHit {
		return FailureOutputLimit
	}
	return FailureOther
}

// IsLimitError 判断一次失败是否属于「长度上限被拒」（两种长度类之一）。
// 只有它返回 true 才允许降档重试。
func IsLimitError(status int, body string) bool {
	return Classify(status, body) != FailureOther
}

// ApplyToBody 把候选值写进请求体的 max_tokens 字段。
//
//   - tokens > 0：直接写入；
//   - tokens == omitTokens（0）：OpenAI 兼容格式**删除该键**，让上游按模型自身
//     能力决定；Anthropic 格式该字段必填，此时退回 FamilyStart，绝不发一个空值。
//
// 返回本次真正生效的值（0 表示未发送该字段）——调用方据此决定要不要写缓存：
// 没发送就没有「成功的值」可言，缓存它是无意义的。
func ApplyToBody(body map[string]interface{}, tokens int, cfg *models.APIModelWithProvider) int {
	if tokens > 0 {
		body["max_tokens"] = tokens
		return tokens
	}
	if cfg != nil && (cfg.APIFormat == "anthropic" || cfg.APIFormat == "claude") {
		v := FamilyStart(cfg)
		body["max_tokens"] = v
		return v
	}
	delete(body, "max_tokens")
	return omitTokens
}

// CanCacheAfter 判断「以降档重试最终成功」的值是否可以写入缓存。
// 上下文超限是本次输入太长造成的，与模型能力无关，不能沉淀为结论。
func CanCacheAfter(kind FailureKind) bool {
	return kind == FailureOutputLimit
}

func containsAny(hay string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(hay, n) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// 缓存
// ---------------------------------------------------------------------------

// Cache 记录「某条模型配置实测可行过的 max_tokens」。
//
// 内存为主、SQLite 为持久层：探测结果不需要强一致，重启后重新加载一次即可。
// 全部失败都只记日志不返回错误 —— 它挂在请求热路径上，不能因为缓存写不进去
// 就把一次正常的对话请求弄失败。
// cacheTTL 是探测结果的保鲜期。超过它就从家族起点重探一次。
//
// 为什么需要：缓存只在「更小」时被采用，等于**只降不升** —— 一次因上下文占满而
// 失败会把值永久压低，之后哪怕只发一句话也上不去。TTL 把起点定期抬回去。
const cacheTTL = 7 * 24 * time.Hour

// probeEvery 是 ε-greedy 的重探周期：每 N 次使用缓存的请求，允许一次从家族起点重来。
// 防的是「TTL 未到但上游能力已提升」以及「缓存值本身就是错的」两种情况。
// 代价约 5% 的请求多花一次失败往返，换来的是不会永久锁死。
const probeEvery = 20

// entry 是一条探测缓存。tokens / verifiedAt 落库；probes 仅存内存
// （每次请求都写库代价太高，重启后归零可以接受）。
type entry struct {
	tokens     int
	probes     int
	verifiedAt time.Time
}

type Cache struct {
	db *sql.DB

	mu     sync.RWMutex
	mem    map[string]entry
	loaded bool
}

var (
	defMu    sync.RWMutex
	defCache *Cache
)

// Init 在服务启动时调用一次，注入数据库连接。
// 未调用时所有缓存操作退化为 no-op（仍然可用，只是每次都要重新探测）。
func Init(db *sql.DB) {
	defMu.Lock()
	defer defMu.Unlock()
	defCache = &Cache{db: db, mem: make(map[string]entry)}
}

func current() *Cache {
	defMu.RLock()
	defer defMu.RUnlock()
	return defCache
}

// Cached 读一条缓存。未初始化或未命中时返回 (0, false)。
func Cached(key string) (int, bool) {
	c := current()
	if c == nil || key == "" {
		return 0, false
	}
	e, ok := c.get(key)
	if !ok {
		return 0, false
	}
	return e.tokens, true
}

// Remember 记录一条探测成功的结果。
func Remember(key string, tokens int) {
	c := current()
	if c == nil || key == "" || tokens <= 0 {
		return
	}
	c.put(key, tokens)
}

// ReprobeDue 判断本次是否应该**忽略缓存**、从家族起点重新探测。
//
// ⚠ 有副作用：每次调用都会给该键的请求计数加一（仅内存）。之所以不做成纯函数，
// 是因为「第 N 次重探」本身就需要一次累加，拆成两步反而容易漏调用。
//
// 两个触发条件（任一命中即重探）：
//   - 距离上次被上游接受已超过 cacheTTL；
//   - 累计使用缓存达到 probeEvery 次。
func ReprobeDue(key string) bool {
	c := current()
	if c == nil || key == "" {
		return false
	}
	return c.reprobeDue(key)
}

// All 返回缓存快照，供接口下发给前端展示「实际探测到的上限」。
func All() map[string]int {
	c := current()
	if c == nil {
		return map[string]int{}
	}
	c.ensureLoaded()
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]int, len(c.mem))
	for k, e := range c.mem {
		out[k] = e.tokens
	}
	return out
}

func (c *Cache) ensureLoaded() {
	c.mu.RLock()
	loaded := c.loaded
	c.mu.RUnlock()
	if loaded || c.db == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.loaded {
		return
	}
	rows, err := c.db.Query(`SELECT cache_key, max_tokens, updated_at FROM model_max_tokens`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var k string
			var v int
			var updated sql.NullString
			if err := rows.Scan(&k, &v, &updated); err != nil || v <= 0 {
				continue
			}
			c.mem[k] = entry{tokens: v, verifiedAt: parseSQLiteTime(updated)}
		}
	}
	c.loaded = true
}

// parseSQLiteTime 解析 CURRENT_TIMESTAMP 写入的 "2006-01-02 15:04:05"（UTC）。
//
// 解析不出来时**按"刚刚验证过"处理**而不是"已过期"：前者最多让一个陈旧值多活
// 一个 TTL，后者会让所有读不到时间的行同时触发重探，打出一批没必要的请求。
func parseSQLiteTime(s sql.NullString) time.Time {
	if !s.Valid || strings.TrimSpace(s.String) == "" {
		return time.Now()
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02T15:04:05Z"} {
		if t, err := time.Parse(layout, s.String); err == nil {
			return t
		}
	}
	return time.Now()
}

func (c *Cache) get(key string) (entry, bool) {
	c.ensureLoaded()
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.mem[key]
	return e, ok
}

// reprobeDue 实现 ReprobeDue 的计数与判据（调用方已保证 key 非空）。
func (c *Cache) reprobeDue(key string) bool {
	c.ensureLoaded()
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.mem[key]
	if !ok {
		return false
	}
	e.probes++
	c.mem[key] = e
	if !e.verifiedAt.IsZero() && time.Since(e.verifiedAt) > cacheTTL {
		return true
	}
	return e.probes%probeEvery == 0
}

func (c *Cache) put(key string, tokens int) {
	c.ensureLoaded()
	c.mu.Lock()
	prev, existed := c.mem[key]
	if existed && prev.tokens == tokens {
		// 值没变：只把内存里的验证时间刷新、计数归零，不写库
		// （否则每次请求都要 UPDATE 一次）。
		c.mem[key] = entry{tokens: tokens, verifiedAt: time.Now()}
		c.mu.Unlock()
		return
	}
	c.mem[key] = entry{tokens: tokens, verifiedAt: time.Now()}
	c.mu.Unlock()

	if c.db == nil {
		return
	}
	if _, err := c.db.Exec(`INSERT INTO model_max_tokens (cache_key, max_tokens, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(cache_key) DO UPDATE SET max_tokens = excluded.max_tokens, updated_at = CURRENT_TIMESTAMP`,
		key, tokens); err != nil {
		log.Printf("[aitokens] 写入探测缓存失败（不影响本次请求）: %v", err)
	}
}
