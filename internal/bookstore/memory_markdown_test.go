package bookstore

import (
	"strings"
	"testing"
)

// 本文件覆盖 docs/books-memory-storage.md §6.4 的 10 条契约测试。
// 解析器是整套记忆落盘的地基：这里出问题，后面五类文件全是坏的。

// 1. 正常往返：渲染 → 解析 → 逐字段相等。
func TestMemoryMarkdown_RoundTrip(t *testing.T) {
	known := [][2]string{
		{"kind", "chat"},
		{"conversation_id", "12"},
		{"book_id", "22"},
		{"title", "列车开场怎么写"},
		{"created_at", "2026-10-18 22:40:03"},
	}
	body := RenderMemoryFrontMatter(known, nil) +
		RenderMarkerLine("chat", [][2]string{{"v", "1"}}) + "\n\n" +
		"# 列车开场怎么写\n\n" +
		RenderPairedMarker("msg", [][2]string{
			{"id", "31"}, {"role", "user"}, {"at", "2026-10-18 22:40:12"},
		}, "主角上车那段太平了。") + "\n" +
		RenderPairedMarker("msg", [][2]string{
			{"id", "32"}, {"role", "assistant"}, {"at", "2026-10-18 22:41:05"},
		}, "可以先把\"车长\"的视线写成一种判定的动作……")

	fm, err := ParseMemoryFrontMatter(body)
	if err != nil {
		t.Fatalf("front-matter 解析失败: %v", err)
	}
	got := KnownFields(fm.Pairs(), "kind", "conversation_id", "book_id", "title", "created_at")
	if got["kind"] != "chat" || got["conversation_id"] != "12" || got["title"] != "列车开场怎么写" {
		t.Fatalf("front-matter 往返不一致: %+v", got)
	}

	markers, err := ParseMemoryMarkers(body)
	if err != nil {
		t.Fatalf("标记解析失败: %v", err)
	}
	if len(markers) != 3 {
		t.Fatalf("应有 3 个标记（chat + 2 个 msg），实得 %d", len(markers))
	}
	if markers[0].Name != "chat" || markers[0].Paired {
		t.Fatalf("chat 应识别为自闭合标记: %+v", markers[0])
	}
	msgs := FindMarkers(markers, "msg")
	if len(msgs) != 2 {
		t.Fatalf("应有 2 条消息，实得 %d", len(msgs))
	}
	if msgs[0].IntAttr("id") != 31 || msgs[0].Attr("role") != "user" {
		t.Fatalf("第 1 条消息属性不对: %+v", msgs[0])
	}
	if msgs[0].Body != "主角上车那段太平了。" {
		t.Fatalf("第 1 条消息正文不对: %q", msgs[0].Body)
	}
	if !strings.Contains(msgs[1].Body, "判定的动作") {
		t.Fatalf("第 2 条消息正文不对: %q", msgs[1].Body)
	}
}

// 2. 缺 -end → 报错（不静默当正文）。
func TestMemoryMarkdown_MissingEnd(t *testing.T) {
	text := "---\nkind: chat\n---\n\n" +
		"<!-- xm:msg id=1 -->\n\n正文没闭合\n"
	_, err := ParseMemoryMarkers(text)
	if err == nil {
		t.Fatal("缺 xm:msg-end 应当报错")
	}
	if !strings.Contains(MemoryParseReason(err), "msg-end") {
		t.Fatalf("原因里应指出缺哪个结束标记: %v", err)
	}
}

// 3. 标记大小写混写仍能识别。
func TestMemoryMarkdown_CaseInsensitive(t *testing.T) {
	text := "---\nkind: chat\n---\n\n" +
		"<!-- XM:MSG ID=7 ROLE=USER -->\n\n大小写测试\n\n<!-- XM:MSG-END -->\n"
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		t.Fatalf("大小写混写不应报错: %v", err)
	}
	if len(markers) != 1 || markers[0].IntAttr("id") != 7 || markers[0].Body != "大小写测试" {
		t.Fatalf("大小写混写解析结果不对: %+v", markers)
	}
}

// 4. 块体里缩进的“伪标记”不受影响（只认第 0 列）。
func TestMemoryMarkdown_IndentedFakeMarker(t *testing.T) {
	text := "---\nkind: chat\n---\n\n" +
		"<!-- xm:input -->\n\n" +
		"这里举例说明标记长什么样：\n\n" +
		"    <!-- xm:msg id=999 -->\n\n" +
		"上面这行是缩进的，不算标记。\n\n" +
		"<!-- xm:input-end -->\n"
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if len(markers) != 1 || markers[0].Name != "input" {
		t.Fatalf("缩进的伪标记被误当成真标记: %+v", markers)
	}
	if !strings.Contains(markers[0].Body, "<!-- xm:msg id=999 -->") {
		t.Fatalf("缩进伪标记应原样留在块体里: %q", markers[0].Body)
	}
}

// 5. 块体含代码围栏、表格、连续连字符 → 原样往返。
func TestMemoryMarkdown_BodyKeepsVerbatim(t *testing.T) {
	raw := strings.Join([]string{
		"| 起点 | 关系 | 终点 |",
		"| --- | --- | --- |",
		"| 林越 | 师徒 | 老周 |",
		"",
		"```json",
		`{"note": "-- 两个连字符 --"}`,
		"```",
		"",
		"结尾",
	}, "\n")
	body := RenderPairedMarker("input", nil, raw)
	text := "---\nkind: agent_session\n---\n\n" + body
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		t.Fatalf("不应报错: %v", err)
	}
	if len(markers) != 1 {
		t.Fatalf("应有 1 个标记，实得 %d", len(markers))
	}
	if markers[0].Body != raw {
		t.Fatalf("块体应原样保留。\n得到: %q\n期望: %q", markers[0].Body, raw)
	}
}

// 6. 属性含空格、引号、连续连字符 → 转义往返一致。
func TestMemoryMarkdown_AttrEscaping(t *testing.T) {
	cases := []string{
		"2026-10-18 22:40:12",
		`带"引号"的值`,
		"含 -- 连续连字符",
		`反斜杠 \ 与 " 混合 -- 一起`,
		"",
	}
	for _, want := range cases {
		line := RenderMarkerLine("msg", [][2]string{{"at", want}})
		if strings.Contains(line, "--") && strings.Count(line, "--") > 1 {
			// 只允许结尾的 " -->"，正文里的 -- 必须被转义掉
		}
		head, ok := parseMarkerLine(line)
		if !ok {
			t.Fatalf("渲染出的标记自己解析不了: %s", line)
		}
		if got := head.Attr("at"); got != want {
			t.Fatalf("属性往返不一致: 期望 %q，得到 %q（行 %s）", want, got, line)
		}
		// 标记行里不允许出现未转义的 "--"（除了结尾的 " -->"）
		inner := strings.TrimSuffix(strings.TrimPrefix(line, memMarkerOpen), memMarkerClose)
		if strings.Contains(inner, "--") {
			t.Fatalf("属性值里的 -- 没被转义: %s", line)
		}
	}
}

// 7. front-matter 缺围栏 / 未闭合 → 报错并给出原因。
func TestMemoryMarkdown_FrontMatterErrors(t *testing.T) {
	if _, err := ParseMemoryMarkers("# 没有 front-matter\n\n正文"); err == nil {
		t.Fatal("缺 front-matter 应报错")
	} else if !strings.Contains(MemoryParseReason(err), "front-matter") {
		t.Fatalf("原因应提到 front-matter: %v", err)
	}
	if _, err := ParseMemoryMarkers("---\nkind: chat\n\n没有闭合"); err == nil {
		t.Fatal("front-matter 未闭合应报错")
	}
}

// 8. 用户自定义的 front-matter 键在写回时不丢，且不覆盖程序字段。
func TestMemoryMarkdown_CustomFieldsKept(t *testing.T) {
	text := "---\nkind: chat\ntitle: 旧标题\n我的标签: 手写\ndraft: true\n---\n\n正文"
	fm, err := ParseMemoryFrontMatter(text)
	if err != nil {
		t.Fatal(err)
	}
	existing := fm.Pairs()
	extra := ExtraFields(existing, "kind", "title")
	if len(extra) != 2 {
		t.Fatalf("应保留 2 个自定义键，实得 %d: %+v", len(extra), extra)
	}
	out := RenderMemoryFrontMatter([][2]string{
		{"kind", "chat"},
		{"title", "新标题"},
	}, extra)
	if !strings.Contains(out, "我的标签: 手写") || !strings.Contains(out, "draft: true") {
		t.Fatalf("自定义键丢失:\n%s", out)
	}
	if strings.Contains(out, "旧标题") {
		t.Fatalf("程序字段应被覆盖:\n%s", out)
	}
}

// 9. 中文、emoji、特殊符号都能正确往返。
func TestMemoryMarkdown_Unicode(t *testing.T) {
	title := "列车 🚄 副车长 · 「引号」"
	val := "中文与 emoji 🎉 一起"
	body := RenderPairedMarker("input", [][2]string{{"note", val}}, title)
	text := "---\nkind: chat\n---\n\n" + body
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		t.Fatal(err)
	}
	if markers[0].Attr("note") != val {
		t.Fatalf("属性里的 emoji 往返失败: %q", markers[0].Attr("note"))
	}
	if markers[0].Body != title {
		t.Fatalf("块体里的中文往返失败: %q", markers[0].Body)
	}
}

// 10. 极端输入不 panic：空文件、只有 front-matter、超长单行。
func TestMemoryMarkdown_EdgeCases(t *testing.T) {
	if _, err := ParseMemoryMarkers(""); err == nil {
		t.Fatal("空文件应报错（缺 front-matter）")
	}
	if _, err := ParseMemoryMarkers("---\nkind: chat\n---\n"); err != nil {
		t.Fatalf("只有 front-matter 不应报错: %v", err)
	}
	long := strings.Repeat("很长的正文", 5000)
	text := "---\nkind: chat\n---\n\n" + RenderPairedMarker("input", nil, long)
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		t.Fatal(err)
	}
	if markers[0].Body != long {
		t.Fatal("超长单行应原样保留")
	}
	// 缺少 --- 的半截文件
	if _, err := ParseMemoryMarkers("---\nkind: chat"); err == nil {
		t.Fatal("半截 front-matter 应报错")
	}
}

// 额外：多余结束标记、真嵌套都要报错，且报错要指向真正的原因。
func TestMemoryMarkdown_StructuralErrors(t *testing.T) {
	stray := "---\nkind: chat\n---\n\n<!-- xm:msg-end -->\n"
	if _, err := ParseMemoryMarkers(stray); err == nil {
		t.Fatal("多余的结束标记应报错")
	}

	// 最常见的情形：误删了上一个 -end。
	// 报错必须指明"缺少 msg-end"，而不是含糊地说"嵌套"——
	// 否则用户会去改没坏的那个标记。
	missingEnd := "---\nkind: chat\n---\n\n" +
		"<!-- xm:msg id=1 -->\n\n正文\n\n" +
		"<!-- xm:display id=1 -->\n\n展示版\n\n" +
		"<!-- xm:display-end -->\n"
	_, err := ParseMemoryMarkers(missingEnd)
	if err == nil {
		t.Fatal("缺 msg-end 应报错")
	}
	reason := MemoryParseReason(err)
	if !strings.Contains(reason, "缺少 xm:msg-end") {
		t.Fatalf("原因应明确指出缺哪个结束标记，实得: %s", reason)
	}
	if strings.HasPrefix(reason, "标记嵌套") {
		t.Fatalf("不应把'误删 -end'说成嵌套: %s", reason)
	}
}

// 额外：IntList / 时间 / 截断等基础工具。
func TestMemoryMarkdown_Helpers(t *testing.T) {
	if got := FormatIntList([]int{1, 3, 5}); got != "[1, 3, 5]" {
		t.Fatalf("FormatIntList: %s", got)
	}
	if got := FormatIntList(nil); got != "[]" {
		t.Fatalf("空列表应渲染成 []: %s", got)
	}
	if got := ParseIntList("[1, 3, 5]"); len(got) != 3 || got[2] != 5 {
		t.Fatalf("ParseIntList: %+v", got)
	}
	if got := ParseIntList("1,2,3"); len(got) != 3 {
		t.Fatalf("ParseIntList 应容忍无方括号: %+v", got)
	}
	if got := ParseIntList(""); got != nil {
		t.Fatalf("空值应返回 nil: %+v", got)
	}
	if got := QuoteFMValue("a: b"); got != `"a: b"` {
		t.Fatalf("含冒号的值应加引号: %s", got)
	}
	// front-matter 不需要转义连字符，但必须能无损往返
	for _, raw := range []string{"含 -- 的值", `带"引号"`, "前后有空格的值", "值为空的情况"} {
		if got := unquoteFMValue(QuoteFMValue(raw)); got != raw {
			t.Fatalf("front-matter 值往返失败: 期望 %q，得到 %q", raw, got)
		}
	}
	if got := TruncateRunes("中文标题很长很长", 4); got != "中文标题" {
		t.Fatalf("按字符截断失败: %s", got)
	}
}

// 行号必须是"整份文件的行号"：用户是对着编辑器改的，
// 差一个 front-matter 的高度就会把人指到错误的位置。
func TestMemoryMarkdown_LineNumbersAreAbsolute(t *testing.T) {
	text := strings.Join([]string{
		"---",                  // 1
		"kind: chat",           // 2
		"title: 行号",            // 3
		"---",                  // 4
		"",                     // 5
		"<!-- xm:chat v=1 -->", // 6
		"",                     // 7
		"## 用户",                // 8
		"",                     // 9
		"<!-- xm:msg id=1 -->", // 10
		"",                     // 11
		"正文",                   // 12
		"",                     // 13
		"<!-- xm:msg-end -->",  // 14
	}, "\n")
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		t.Fatal(err)
	}
	msg, ok := FindMarker(markers, "msg")
	if !ok {
		t.Fatal("没找到 msg 标记")
	}
	if msg.Line != 10 {
		t.Fatalf("行号应为整份文件的第 10 行，实得 %d", msg.Line)
	}

	// 报错信息里的行号也要是整份文件的行号
	broken := strings.Replace(text, "<!-- xm:msg id=1 -->", "<!-- xm:msg id=1 -->", 1)
	broken = strings.Replace(broken, "<!-- xm:msg-end -->", "", 1)
	_, err = ParseMemoryMarkers(broken)
	if err == nil {
		t.Fatal("应报错")
	}
	reason := MemoryParseReason(err)
	if !strings.Contains(reason, "第 10 行") {
		t.Fatalf("报错应指向文件第 10 行，实得: %s", reason)
	}
}
