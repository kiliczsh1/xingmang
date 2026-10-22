package bookstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 拆书与生成历史原先只在浏览器 localStorage 里：没有库行、没有文件。
// 这一组用例锁住"写下去、读回来、原样还给你"这三件事。

func newRecordBook(t *testing.T) (*Service, string, int) {
	t.Helper()
	svc, root := setupService(t)
	b := &models.Book{Title: "记录书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	return svc, root, b.ID
}

// 拆书：payload 里的 messages / fieldValues 这类结构化字段必须一字不差地回来。
func TestRecord_AnalysisRoundTrip(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	payload := map[string]any{
		"id": 1729262400000,
		"messages": []map[string]string{
			{"role": "user", "content": "拆一下这三章"},
			{"role": "assistant", "content": "第一处问题是节奏。"},
		},
		"fieldValues":      map[string]string{"重点": "节奏", "角度": "读者"},
		"relatedContext":   []map[string]string{{"type": "chapter", "title": "第一章 上车"}},
		"selectedChapters": []int{1, 2, 3},
	}
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind:       MemoryKindAnalysis,
		ExternalID: "1729262400000",
		Title:      "黄金三章拆解",
		Source:     "拆书",
		Status:     "completed",
		Prompt:     "黄金三章拆解",
		Chapters:   []int{1, 2, 3},
		Result:     "第一处问题是节奏。\n第二处是人物的动机不够。",
		Readable:   "### system · 提示词\n\n你是资深编辑。\n\n### user · 素材\n\n（第一章正文）",
		Payload:    JSONPayloadOf(payload),
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.FilePath == "" {
		t.Fatal("应写回 file_path")
	}
	if !strings.HasPrefix(rec.FilePath, "memory/analysis/") {
		t.Fatalf("拆书应落在 memory/analysis/: %q", rec.FilePath)
	}

	abs := filepath.Join(root, "记录书", filepath.FromSlash(rec.FilePath))
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("记录文件不存在: %v", err)
	}
	text := string(raw)
	for _, want := range []string{
		"kind: analysis",
		"external_id:",
		"title:",
		"## 结果",
		"## 输入与提示词",
		"第一处问题是节奏。",
		"你是资深编辑。",
		"```json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("文件里缺少 %q:\n%s", want, text)
		}
	}

	// 读回来：payload 必须能原样还原
	doc, err := svc.LoadRecord(rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doc.Result, "第二处是人物的动机不够") {
		t.Fatalf("结果正文丢失: %q", doc.Result)
	}
	var back map[string]any
	if err := json.Unmarshal([]byte(doc.Payload), &back); err != nil {
		t.Fatalf("payload 不是合法 JSON: %v\n%s", err, doc.Payload)
	}
	if back["fieldValues"].(map[string]any)["重点"] != "节奏" {
		t.Fatalf("结构化字段没还原: %v", back)
	}
	if len(back["messages"].([]any)) != 2 {
		t.Fatalf("消息数组没还原: %v", back["messages"])
	}
}

// 生成历史：按月分目录，source 区分续写/改写/创作。
func TestRecord_GenerationGoesToMonthDir(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind:   MemoryKindGeneration,
		Title:  "续写",
		Source: "continuation",
		Status: "completed",
		Result: "他把车票递了过去。",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rec.FilePath, "memory/history/") {
		t.Fatalf("生成历史应在 memory/history/: %q", rec.FilePath)
	}
	// 目录名必须是 YYYY-MM
	dir := filepath.Base(filepath.Dir(filepath.FromSlash(rec.FilePath)))
	if len(dir) != 7 || dir[4] != '-' {
		t.Fatalf("应按月分目录，实得 %q", dir)
	}
	if _, err := os.Stat(filepath.Join(root, "记录书", filepath.FromSlash(rec.FilePath))); err != nil {
		t.Fatalf("文件未生成: %v", err)
	}
}

// 改标题要连文件名一起改（与章节、记忆其他三类同一规则）。
func TestRecord_RenameFollowsFileName(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "初稿拆解", Source: "拆书", Result: "结果",
	})
	if err != nil {
		t.Fatal(err)
	}
	before := rec.FilePath
	upd, err := svc.UpdateRecord(rec.ID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "终稿拆解", Source: "拆书", Result: "结果 2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if upd.FilePath == before {
		t.Fatalf("文件名应跟随标题: %q", before)
	}
	if !strings.Contains(upd.FilePath, "终稿拆解") {
		t.Fatalf("文件名里应是新标题: %q", upd.FilePath)
	}
	if _, err := os.Stat(filepath.Join(root, "记录书", filepath.FromSlash(before))); !os.IsNotExist(err) {
		t.Fatal("旧文件名的文件不该还留着")
	}
	if _, err := os.Stat(filepath.Join(root, "记录书", filepath.FromSlash(upd.FilePath))); err != nil {
		t.Fatalf("新文件不存在: %v", err)
	}
}

// 文件被外部改过时，保存要被拒（返回 MemoryRecordConflict），且不覆盖。
//
// 与后台渲染不同，这里的写入来自用户点「保存」，静默跳过等于骗他。
func TestRecord_ExternalEditBlocksSave(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindGeneration, Title: "续写", Source: "continuation", Result: "原文",
	})
	if err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(root, "记录书", filepath.FromSlash(rec.FilePath))
	raw, _ := os.ReadFile(abs)
	if err := os.WriteFile(abs, []byte(strings.Replace(string(raw), "原文", "我在外部改的", 1)), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.UpdateRecord(rec.ID, MemoryRecordInput{
		Kind: MemoryKindGeneration, Title: "续写", Source: "continuation", Result: "界面里的新版本",
	}); err != MemoryRecordConflict {
		t.Fatalf("应返回 MemoryRecordConflict，实得 %v", err)
	}
	after, _ := os.ReadFile(abs)
	if !strings.Contains(string(after), "我在外部改的") {
		t.Fatalf("外部改动被覆盖了:\n%s", string(after))
	}
	var pe string
	svc.DB.QueryRow(`SELECT COALESCE(parse_error,'') FROM memory_records WHERE id=?`, rec.ID).Scan(&pe)
	if !strings.Contains(pe, "未覆盖") {
		t.Fatalf("必须留下原因，实得 %q", pe)
	}
}

// 自定义 front-matter 键在重写时要保住。
func TestRecord_KeepsCustomFrontMatter(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "带自定义键", Source: "拆书", Result: "结果",
	})
	if err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(root, "记录书", filepath.FromSlash(rec.FilePath))
	raw, _ := os.ReadFile(abs)
	edited := strings.Replace(string(raw), "---\n\n", "my_note: 我自己加的\n---\n\n", 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	// 让索引哈希与磁盘对齐（真实流程里会先读一次）
	if _, err := svc.LoadRecord(rec.ID); err != nil {
		t.Fatal(err)
	}
	// 把 index 的 hash 更新成当前内容
	cur, _ := os.ReadFile(abs)
	svc.DB.Exec(`UPDATE memory_records SET content_hash=? WHERE id=?`, HashText(string(cur)), rec.ID)

	if _, err := svc.UpdateRecord(rec.ID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "带自定义键", Source: "拆书", Result: "换了个结果",
	}); err != nil {
		t.Fatal(err)
	}
	out, _ := os.ReadFile(abs)
	if !strings.Contains(string(out), "my_note") {
		t.Fatalf("自定义键被抹掉了:\n%s", string(out))
	}
	if !strings.Contains(string(out), "换了个结果") {
		t.Fatalf("新结果没写进去:\n%s", string(out))
	}
}

// 类型只认 analysis / generation —— 写错类别会让文件跑到错误目录里。
func TestRecord_RejectsUnknownKind(t *testing.T) {
	svc, _, bookID := newRecordBook(t)
	if _, err := svc.CreateRecord(bookID, MemoryRecordInput{Kind: "whatever", Title: "x"}); err == nil {
		t.Fatal("未知类型应被拒绝")
	}
}

// 删除：文件进回收站，索引行消失。
func TestRecord_TrashRemovesIndexAndFile(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "待删", Source: "拆书", Result: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(root, "记录书", filepath.FromSlash(rec.FilePath))
	if err := svc.TrashRecord(rec.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(abs); err == nil {
		t.Fatal("文件应已移走")
	}
	var n int
	svc.DB.QueryRow(`SELECT COUNT(*) FROM memory_records WHERE id=?`, rec.ID).Scan(&n)
	if n != 0 {
		t.Fatal("索引行应已删除")
	}
}

// (书, 类型, external_id) 去重：重复导入同一份旧记录不会产生重复。
func TestRecord_ExternalIDDedup(t *testing.T) {
	svc, _, bookID := newRecordBook(t)
	first, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, ExternalID: "abc", Title: "旧记录", Source: "拆书", Result: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, ExternalID: "abc", Title: "旧记录", Source: "拆书", Result: "x",
	}); err == nil {
		t.Fatal("同一 external_id 再插应被唯一索引挡住")
	}
	got, err := svc.FindRecordByExternal(bookID, MemoryKindAnalysis, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != first.ID {
		t.Fatalf("查到的应是同一条: %d vs %d", got.ID, first.ID)
	}
}

// 块体以换行结尾，紧接着写 `## 标题` 会被 Markdown 当成上一段的懒延续，
// 渲染出来像被吞进了上一块。标题前必须有空行。
func TestRecord_BlankLineBeforeHeadings(t *testing.T) {
	svc, root, bookID := newRecordBook(t)
	rec, err := svc.CreateRecord(bookID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "空行", Source: "拆书",
		Result: "结果正文", Readable: "输入正文", Payload: `{"a":1}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "记录书", filepath.FromSlash(rec.FilePath)))
	text := string(raw)
	for _, bad := range []string{"<!-- xm:result-end -->\n\n\n##", "<!-- xm:input-end -->\n\n\n<!--"} {
		if strings.Contains(text, bad) {
			t.Fatalf("不该有三连换行: %q\n%s", bad, text)
		}
	}
	for _, good := range []string{"<!-- xm:result-end -->\n\n## 输入与提示词"} {
		if !strings.Contains(text, good) {
			t.Fatalf("标题前应恰好一个空行，找不到 %q:\n%s", good, text)
		}
	}
	if strings.HasSuffix(text, "\n\n") {
		t.Fatal("文件不该以多余空行结尾")
	}
}
