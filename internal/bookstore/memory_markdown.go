package bookstore

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// 记忆文件的 Markdown 契约（docs/books-memory-storage.md §5、§6）。
//
// 目标：一个文件同时被人读和机器读。
//   - front-matter：结构化小字段；
//   - HTML 注释标记：机器用来定位结构，人在渲染视图里看不见；
//   - 标记之间的块体：原样保留的正文，不需要任何转义；
//   - 装饰内容（## 标题、表格、说明文字）：程序完全不解析，用户随便改。
//
// 约定（与本文件实现一一对应）：
//  1. 标记必须从**第 0 列**开始。缩进过的“像标记的行”按普通正文处理，
//     这样用户在正文里举例写 <!-- xm:msg --> 不会破坏解析。
//  2. 标记**平铺，不允许嵌套**。需要关联的两块用同一个 id 属性关联。
//  3. 只有属性值需要转义：\ → \\、" → \"、-- → -\-。
//  4. 缺少配对的 -end → 解析失败（而非静默当成普通文本）。
const (
	memMarkerOpen       = "<!-- xm:"
	memMarkerClose      = "-->"
	memTimeLayout       = "2006-01-02 15:04:05"
	memNamePrefix       = "xm:"
	memMaxTitleRunes    = 40
	memFileNameTimeForm = "20060102-1504"
)

// memorySelfClosing 是不需要配对 -end 的标记名。
//
// 其余标记一律视为成对出现，缺 -end 直接报错 —— 这是"改坏了要能发现"的前提。
//
// xm:run 是自闭合的**元数据**标记（记录这一轮的时间与状态），它**不包裹**其它块；
// 同一轮的输入/回答/工具调用用 `run=<id>` 属性与之关联。这样所有标记都是平铺的，
// 也就不会出现"嵌套"这种需要额外解析逻辑的结构。
var memorySelfClosing = map[string]bool{
	"chat":     true,
	"session":  true,
	"record":   true,
	"graph":    true,
	"notice":   true,
	"run":      true,
	"proposal": true,
}

// MemoryParseError 是记忆文件的解析错误，Reason 是可以直接展示给用户的原因。
type MemoryParseError struct{ Reason string }

func (e *MemoryParseError) Error() string { return e.Reason }

func memErr(format string, a ...any) error {
	return &MemoryParseError{Reason: fmt.Sprintf(format, a...)}
}

// MemoryParseReason 返回给人看的失败原因；没有错误时返回空串。
func MemoryParseReason(err error) string {
	if err == nil {
		return ""
	}
	if pe, ok := err.(*MemoryParseError); ok {
		return pe.Reason
	}
	return err.Error()
}

// ---------------- 属性转义 ----------------

// EncodeAttr 把任意文本编码成可以安全放进标记属性的形式。
//
// HTML 注释里唯一禁止的序列是 "--"，所以把它拆成 "-\-"；
// 反斜杠与双引号按常规转义。
func EncodeAttr(v string) string {
	var b strings.Builder
	runes := []rune(v)
	for i := 0; i < len(runes); i++ {
		switch r := runes[i]; r {
		case '\\':
			b.WriteString(`\\`)
		case '"':
			b.WriteString(`\"`)
		case '-':
			if i+1 < len(runes) && runes[i+1] == '-' {
				b.WriteString(`-\-`) // 转义第二个连字符
				i++
			} else {
				b.WriteRune('-')
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// DecodeAttr 还原 EncodeAttr 的结果。单遍扫描，避免多轮替换互相干扰。
func DecodeAttr(v string) string {
	var b strings.Builder
	escaped := false
	for _, r := range v {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// formatAttrValue 在编码的基础上决定要不要加引号。
func formatAttrValue(v string) string {
	enc := EncodeAttr(v)
	if enc == "" {
		return `""`
	}
	if strings.ContainsAny(enc, " \t") {
		return `"` + enc + `"`
	}
	return enc
}

// ---------------- 标记 ----------------

// MemoryMarker 是记忆文件里的一个标记（及其块体）。
type MemoryMarker struct {
	Name   string
	Attrs  map[string]string
	Keys   []string // 属性出现顺序，便于稳定渲染
	Body   string   // 成对标记之间的原文；自闭合标记为空
	Paired bool
	Line   int // 起始行号（1 起），用于报错定位
}

// Attr 取属性值，不存在返回空串。
func (m MemoryMarker) Attr(key string) string { return m.Attrs[key] }

// IntAttr 取整型属性，解析失败返回 0。
func (m MemoryMarker) IntAttr(key string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(m.Attrs[key]))
	return n
}

// BoolAttr 取布尔属性：值为 "1" / "true" / "yes" / "on" 时视为真。
func (m MemoryMarker) BoolAttr(key string) bool {
	switch strings.ToLower(strings.TrimSpace(m.Attrs[key])) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// parseMarkerLine 判断一行是否是 xm 标记，并解析出名字与属性。
//
// 只认第 0 列开始的标记；缩进过的按普通正文处理（见文件头注释）。
// 前缀与名字大小写不敏感（`<!-- XM:MSG -->` 与 `<!-- xm:msg -->` 等价）。
func parseMarkerLine(line string) (MemoryMarker, bool) {
	s := strings.TrimRight(line, " \t")
	if len(s) < len(memMarkerOpen)+len(memMarkerClose) {
		return MemoryMarker{}, false
	}
	if !strings.EqualFold(s[:len(memMarkerOpen)], memMarkerOpen) || !strings.HasSuffix(s, memMarkerClose) {
		return MemoryMarker{}, false
	}
	inner := strings.TrimSpace(s[len(memMarkerOpen) : len(s)-len(memMarkerClose)])
	if inner == "" {
		return MemoryMarker{}, false
	}
	// 名字
	i := 0
	for i < len(inner) && !isMarkerSpace(inner[i]) {
		i++
	}
	name := inner[:i]
	if name == "" {
		return MemoryMarker{}, false
	}
	m := MemoryMarker{Name: strings.ToLower(name), Attrs: map[string]string{}}
	attrs, err := parseAttrList(inner[i:])
	if err != nil {
		// 属性坏了不能当成普通文本：否则用户会以为标记生效了。
		m.Name = strings.ToLower(name)
		m.Attrs = map[string]string{"__parse_error": MemoryParseReason(err)}
		return m, true
	}
	for _, kv := range attrs {
		if _, exists := m.Attrs[kv[0]]; !exists {
			m.Keys = append(m.Keys, kv[0])
		}
		m.Attrs[kv[0]] = kv[1]
	}
	return m, true
}

func isMarkerSpace(c byte) bool { return c == ' ' || c == '\t' }

// parseAttrList 解析 "k=v k2=\"v 2\"" 形式的属性串。
func parseAttrList(s string) ([][2]string, error) {
	var out [][2]string
	i := 0
	for i < len(s) {
		for i < len(s) && isMarkerSpace(s[i]) {
			i++
		}
		if i >= len(s) {
			break
		}
		start := i
		for i < len(s) && s[i] != '=' && !isMarkerSpace(s[i]) {
			i++
		}
		key := s[start:i]
		if key == "" {
			return nil, memErr("标记属性名为空")
		}
		// 属性名统一小写：用户可能写成 ID=7，读的时候要能取到 id。
		key = strings.ToLower(key)
		if i >= len(s) || s[i] != '=' {
			return nil, memErr("标记属性 %q 缺少等号", key)
		}
		i++ // 跳过 '='
		var val string
		if i < len(s) && s[i] == '"' {
			i++
			var b strings.Builder
			closed := false
			for i < len(s) {
				if s[i] == '\\' && i+1 < len(s) {
					b.WriteByte(s[i])
					b.WriteByte(s[i+1])
					i += 2
					continue
				}
				if s[i] == '"' {
					i++
					closed = true
					break
				}
				b.WriteByte(s[i])
				i++
			}
			if !closed {
				return nil, memErr("标记属性 %q 的引号未闭合", key)
			}
			val = DecodeAttr(b.String())
		} else {
			start = i
			for i < len(s) && !isMarkerSpace(s[i]) {
				i++
			}
			val = DecodeAttr(s[start:i])
		}
		out = append(out, [2]string{key, val})
	}
	return out, nil
}

// ParseMemoryMarkers 解析整个文件：先严格读 front-matter，再扫描标记。
//
// 失败即返回错误，调用方把原因写进 parse_error —— 不猜、不降级、不覆盖文件。
//
// 标记上的 Line 是**整份文件的行号**（1 起），不是正文内的相对行号：
// 报错时用户是对着编辑器里的行号去改的，差一个 front-matter 的高度就会把人指错地方。
func ParseMemoryMarkers(text string) ([]MemoryMarker, error) {
	fm, err := ParseMemoryFrontMatter(text)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(fm.Body, "\n")
	offset := memoryBodyLineOffset(text, fm)
	var out []MemoryMarker
	for i := 0; i < len(lines); i++ {
		head, ok := parseMarkerLine(lines[i])
		if !ok {
			continue
		}
		if reason := head.Attrs["__parse_error"]; reason != "" {
			return nil, memErr("%s（第 %d 行）", reason, offset+i)
		}
		if strings.HasSuffix(head.Name, "-end") {
			return nil, memErr("出现多余的结束标记 xm:%s（第 %d 行）", head.Name, offset+i)
		}
		if memorySelfClosing[head.Name] {
			head.Line = offset + i
			out = append(out, head)
			continue
		}

		endName := head.Name + "-end"
		var body []string
		closed := false
		j := i + 1
		// 见到另一个"开始标记"时不立刻报"嵌套"：最常见的原因其实是**误删了上一个 -end**。
		// 报错要指向真正的原因，否则用户会去改没坏的地方。
		orphan := ""
		orphanLine := 0
		for ; j < len(lines); j++ {
			next, ok := parseMarkerLine(lines[j])
			if !ok {
				body = append(body, lines[j])
				continue
			}
			if strings.EqualFold(next.Name, endName) {
				closed = true
				break
			}
			orphan, orphanLine = next.Name, offset+j
			break
		}
		if !closed {
			if orphan != "" {
				return nil, memErr("xm:%s（第 %d 行）缺少 xm:%s-end —— 第 %d 行出现 xm:%s，"+
					"多半是上一个结束标记被删掉了", head.Name, offset+i, head.Name, orphanLine, orphan)
			}
			return nil, memErr("xm:%s 缺少 xm:%s-end（起始于第 %d 行）", head.Name, head.Name, offset+i)
		}
		head.Paired = true
		head.Line = offset + i
		head.Body = trimOneBlankLine(strings.Join(body, "\n"))
		out = append(out, head)
		i = j
	}
	return out, nil
}

// memoryBodyLineOffset 返回"正文第 0 行"在整份文件里的**行号**（1 起）。
//
// 用它做 `offset + 正文内下标` 就能得到整份文件的行号。
// 解析 front-matter 时会跳过一个空行，偏移量因此与围栏位置差一，必须按同一套规则重算。
func memoryBodyLineOffset(text string, fm FrontMatter) int {
	text = NormalizeText(text)
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t") != "---" {
		return 0
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t") != "---" {
			continue
		}
		j := i + 1
		if j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		_ = fm
		return j + 1 // lines[j] 的 1 基行号
	}
	return 0
}

// trimOneBlankLine 去掉块体首尾各一个换行（规范 §5.0 规则 8）。
func trimOneBlankLine(s string) string {
	s = strings.TrimPrefix(s, "\n")
	s = strings.TrimSuffix(s, "\n")
	return s
}

// FindMarker 取第一个指定名字的标记。
func FindMarker(list []MemoryMarker, name string) (MemoryMarker, bool) {
	name = strings.ToLower(name)
	for _, m := range list {
		if m.Name == name {
			return m, true
		}
	}
	return MemoryMarker{}, false
}

// FindMarkers 取所有指定名字的标记。
func FindMarkers(list []MemoryMarker, name string) []MemoryMarker {
	name = strings.ToLower(name)
	var out []MemoryMarker
	for _, m := range list {
		if m.Name == name {
			out = append(out, m)
		}
	}
	return out
}

// MarkerBody 取指定名字标记的块体；不存在返回空串与 false。
func MarkerBody(list []MemoryMarker, name string) (string, bool) {
	if m, ok := FindMarker(list, name); ok {
		return m.Body, true
	}
	return "", false
}

// RenderMarkerLine 渲染单行标记（自闭合）。
func RenderMarkerLine(name string, attrs [][2]string) string {
	var b strings.Builder
	b.WriteString(memMarkerOpen)
	b.WriteString(name)
	for _, kv := range attrs {
		if kv[0] == "" {
			continue
		}
		b.WriteString(" ")
		b.WriteString(kv[0])
		b.WriteString("=")
		b.WriteString(formatAttrValue(kv[1]))
	}
	b.WriteString(" ")
	b.WriteString(memMarkerClose)
	return b.String()
}

// RedactSecrets 遮蔽工具参数里的敏感字段。
//
// 依据是**键名**而不是值：键名能稳定表达"这是什么"，而值可能是任何形状。
// 同时兜一层"值里出现 Bearer/sk- 令牌"的情形，避免把密钥塞进普通字符串字段。
//
// ⚠ 只用于 md 视图的展示；数据库里的原始账本（agent_events / agent_tool_calls）不动 ——
// 那是运行期事实，改了就查不出问题。
func RedactSecrets(raw string) string {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return redactTokenPatterns(raw)
	}
	return mustJSONRedacted(v)
}

func mustJSONRedacted(v any) string {
	b, err := json.Marshal(redactValue(v))
	if err != nil {
		return ""
	}
	return redactTokenPatterns(string(b))
}

func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if isSecretKey(k) {
				out[k] = "***"
				continue
			}
			out[k] = redactValue(val)
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			out = append(out, redactValue(item))
		}
		return out
	default:
		// 字符串字段里的令牌也要遮掉
		if s, ok := v.(string); ok {
			return redactTokenPatterns(s)
		}
		return v
	}
}

// isSecretKey 判断键名是否属于"必须遮蔽"的敏感信息。
func isSecretKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	k = strings.NewReplacer("-", "", "_", "", " ", "").Replace(k)
	for _, frag := range []string{"apikey", "token", "secret", "password", "passwd", "authorization", "credential", "privatekey"} {
		if strings.Contains(k, frag) {
			return true
		}
	}
	return false
}

var tokenPattern = regexp.MustCompile(`(?i)\b(Bearer\s+[A-Za-z0-9._\-]{8,}|sk-[A-Za-z0-9._\-]{8,})`)

// redactTokenPatterns 兜底遮蔽"长得像令牌"的字符串。
func redactTokenPatterns(s string) string {
	return tokenPattern.ReplaceAllString(s, "***")
}

// TruncateRunesHelp 是给上层用的显式截断提示（超出时补省略号）。
func TruncateRunesHelp(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max]) + "…"
}

// RenderPairedMarker 渲染成对标记：起始行、空行、块体、空行、结束行。
//
// 结尾统一补一个空行：否则紧跟在后面的装饰行（比如下一个 `## 助手` 标题）
// 会与 `-end` 贴在一起，渲染出来像是被吞进了上一块。
func RenderPairedMarker(name string, attrs [][2]string, body string) string {
	var b strings.Builder
	b.WriteString(RenderMarkerLine(name, attrs))
	b.WriteString("\n\n")
	if inner := strings.Trim(body, "\n"); inner != "" {
		b.WriteString(inner)
		b.WriteString("\n\n")
	}
	b.WriteString(RenderMarkerLine(name+"-end", nil))
	b.WriteString("\n\n")
	return b.String()
}

// ---------------- front-matter ----------------

// ParseMemoryFrontMatter 严格解析 front-matter。
//
// 与章节用的 ParseFrontMatter 的区别：章节允许"没有 front-matter"（用户可能从别处拷来纯文本），
// 记忆文件是程序自己写的，缺围栏等于文件被改坏，必须报错。
func ParseMemoryFrontMatter(text string) (FrontMatter, error) {
	text = NormalizeText(text)
	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t") != "---" {
		return FrontMatter{Body: text}, memErr("缺少 front-matter（文件第一行必须是 ---）")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t") == "---" {
			body := lines[i+1:]
			if len(body) > 0 && strings.TrimSpace(body[0]) == "" {
				body = body[1:]
			}
			return FrontMatter{
				Has:   true,
				Lines: append([]string(nil), lines[1:i]...),
				Body:  strings.Join(body, "\n"),
			}, nil
		}
	}
	return FrontMatter{Body: text}, memErr("front-matter 未闭合（找不到结束的 ---）")
}

// Pairs 按出现顺序返回 front-matter 的键值对，值已去引号。
func (f FrontMatter) Pairs() [][2]string {
	var out [][2]string
	for _, raw := range f.Lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		out = append(out, [2]string{key, unquoteFMValue(val)})
	}
	return out
}

func unquoteFMValue(v string) string {
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			inner := v[1 : len(v)-1]
			inner = strings.ReplaceAll(inner, `\"`, `"`)
			inner = strings.ReplaceAll(inner, `\\`, `\`)
			inner = strings.ReplaceAll(inner, `\'`, `'`)
			return inner
		}
	}
	return v
}

// QuoteFMValue 生成可安全写进单行 front-matter 的值。
//
// front-matter 一行一个键，所以换行先归一为空格；含特殊字符时加双引号。
func QuoteFMValue(v string) string {
	v = strings.ReplaceAll(NormalizeText(v), "\n", " ")
	v = strings.TrimRight(v, " \t")
	if v == "" {
		return `""`
	}
	if needQuoteFM(v) {
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `"`, `\"`)
		return `"` + v + `"`
	}
	return v
}

func needQuoteFM(v string) bool {
	if strings.TrimSpace(v) != v {
		return true
	}
	if strings.ContainsAny(v, `:#"'\\`) {
		return true
	}
	switch v[0] {
	case '[', '{', '&', '*', '!', '|', '>', '%', '@', '`':
		return true
	}
	return false
}

// RenderMemoryFrontMatter 渲染 front-matter。
//
// known 是程序负责的字段（顺序固定）；extra 是用户自己加的键，原样保留在最后，
// 且不会覆盖程序字段 —— 用户的自定义字段永远不会被抹掉。
func RenderMemoryFrontMatter(known [][2]string, extra [][2]string) string {
	var b strings.Builder
	b.WriteString("---\n")
	seen := make(map[string]bool, len(known)+len(extra))
	for _, kv := range known {
		if kv[0] == "" {
			continue
		}
		seen[kv[0]] = true
		b.WriteString(kv[0])
		b.WriteString(": ")
		b.WriteString(QuoteFMValue(kv[1]))
		b.WriteString("\n")
	}
	for _, kv := range extra {
		if kv[0] == "" || seen[kv[0]] {
			continue
		}
		seen[kv[0]] = true
		b.WriteString(kv[0])
		b.WriteString(": ")
		b.WriteString(QuoteFMValue(kv[1]))
		b.WriteString("\n")
	}
	b.WriteString("---\n")
	return b.String()
}

// ExtraFields 从已有 front-matter 里挑出不属于 knownKeys 的字段。
func ExtraFields(existing [][2]string, knownKeys ...string) [][2]string {
	known := make(map[string]bool, len(knownKeys))
	for _, k := range knownKeys {
		known[k] = true
	}
	var out [][2]string
	for _, kv := range existing {
		if kv[0] == "" || known[kv[0]] {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// KnownFields 按给定顺序从已有 field 里取值，缺失的用空串补位。
func KnownFields(existing [][2]string, keys ...string) map[string]string {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = ""
	}
	for _, kv := range existing {
		if _, ok := out[kv[0]]; ok {
			out[kv[0]] = kv[1]
		}
	}
	return out
}

// ---------------- 时间与列表 ----------------

// FormatMemTime 记忆文件里统一的时间格式（本地时间，与 SQLite CURRENT_TIMESTAMP 一致）。
func FormatMemTime(t time.Time) string { return t.Format(memTimeLayout) }

// ParseMemTime 解析记忆文件里的时间；失败返回零值。
func ParseMemTime(s string) time.Time {
	t, err := time.ParseInLocation(memTimeLayout, strings.TrimSpace(s), time.Local)
	if err != nil {
		return time.Time{}
	}
	return t
}

// ParseDBTime 解析数据库里的时间戳，并换算成本地时间。
//
// SQLite 的 CURRENT_TIMESTAMP 存的是 **UTC**，而记忆文件写的是本地时间。
// 不做换算的话，同一个文件里会出现"created_at 15:48（其实是 UTC 被当成本地读）"
// 与"消息 at 23:48（真本地）"相差 8 小时的诡异现象。
func ParseDBTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	layouts := []string{
		memTimeLayout,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t.Local()
		}
	}
	return time.Time{}
}

// FormatIntList 把整数列表渲染成 [1, 2, 3]；空列表渲染成 [].
func FormatIntList(xs []int) string {
	if len(xs) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(xs))
	for _, x := range xs {
		parts = append(parts, strconv.Itoa(x))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// FormatIntListAttr 是标记属性里的紧凑写法：`1,3`（不带方括号、不带空格）。
//
// 不带空格就无需加引号，标记行也更短；解析侧 ParseIntList 两种写法都认。
func FormatIntListAttr(xs []int) string {
	if len(xs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(xs))
	for _, x := range xs {
		parts = append(parts, strconv.Itoa(x))
	}
	return strings.Join(parts, ",")
}

// ParseIntList 解析 [1, 2, 3] / 1,2,3 / 空值，忽略非法项。
func ParseIntList(s string) []int {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// TruncateRunes 按字符截断，用于文件名（不按字节，避免截断中文）。
func TruncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// unknownFieldReason 统一给"未知 kind"之类的错误，避免各处文案不一致。
func unknownKindErr(kind string) error {
	return memErr("未知的记忆类型: %s", kind)
}

// memoryFileKind 是写进 front-matter 的 kind。
//
// 与内部的 flusher 键（MemoryKindAgent="agent"）区分开：文件名/目录用短键，
// 文件头的 kind 用能自解释的完整名，别人打开文件一眼能看懂这是什么。
func memoryFileKind(kind string) string {
	if kind == MemoryKindAgent {
		return "agent_session"
	}
	return kind
}
