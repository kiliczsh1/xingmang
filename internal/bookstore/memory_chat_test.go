package bookstore

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// M1 的核心验证：文件是真相、手改生效、改坏要能发现、并发不丢消息。

func newChatBook(t *testing.T) (*Service, string, *models.Book, int) {
	t.Helper()
	svc, root := setupService(t)
	b := &models.Book{Title: "对话落盘", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	res, err := svc.DB.Exec(`INSERT INTO conversations (book_id, title) VALUES (?, ?)`, b.ID, "列车开场")
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return svc, root, b, int(id)
}

func TestChatFile_CreateAndAppend(t *testing.T) {
	svc, root, _, convID := newChatBook(t)

	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	if !strings.HasPrefix(rel, "memory/chats/") || !strings.HasSuffix(rel, ".md") {
		t.Fatalf("对话文件应落在 memory/chats/ 下: %q", rel)
	}
	abs := filepath.Join(root, "对话落盘", filepath.FromSlash(rel))
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("文件未创建: %v", err)
	}

	first, err := svc.AppendChatMessage(convID, "user", "主角上车那段太平了", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendChatMessage(convID, "assistant", "可以加一点压迫感", `带"引号"的展示版`); err != nil {
		t.Fatal(err)
	}

	doc, err := svc.LoadChat(convID)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Messages) != 2 {
		t.Fatalf("应有 2 条消息，实得 %d", len(doc.Messages))
	}
	if doc.Messages[0].ID != first.ID || doc.Messages[0].Role != "user" {
		t.Fatalf("第 1 条消息不对: %+v", doc.Messages[0])
	}
	if doc.Messages[1].Display != `带"引号"的展示版` {
		t.Fatalf("display 未往返: %q", doc.Messages[1].Display)
	}

	// 文件里应该是两个 msg 块 + 一个 display 块
	raw, _ := os.ReadFile(abs)
	if n := strings.Count(string(raw), "<!-- xm:msg "); n != 2 {
		t.Fatalf("文件里应有 2 个 msg 标记，实得 %d", n)
	}
	if n := strings.Count(string(raw), "<!-- xm:display "); n != 1 {
		t.Fatalf("文件里应有 1 个 display 标记，实得 %d", n)
	}
}

func TestChatFile_ManualEditTakesEffect(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendChatMessage(convID, "user", "原文", ""); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	abs := filepath.Join(root, "对话落盘", filepath.FromSlash(rel))

	// 模拟用户在 VSCode / 记事本里改正文
	raw, _ := os.ReadFile(abs)
	edited := strings.Replace(string(raw), "原文", "我在外部改过的内容", 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	doc, err := svc.LoadChat(convID)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Messages[0].Content != "我在外部改过的内容" {
		t.Fatalf("外部改动没生效: %q", doc.Messages[0].Content)
	}

	// 再追加一条，外部改动不能被覆盖掉
	if _, err := svc.AppendChatMessage(convID, "assistant", "追加", ""); err != nil {
		t.Fatal(err)
	}
	doc, _ = svc.LoadChat(convID)
	if len(doc.Messages) != 2 || doc.Messages[0].Content != "我在外部改过的内容" {
		t.Fatalf("追加后外部改动丢了: %+v", doc.Messages)
	}
}

func TestChatFile_BrokenMarkerIsReported(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendChatMessage(convID, "user", "正文", ""); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	abs := filepath.Join(root, "对话落盘", filepath.FromSlash(rel))

	raw, _ := os.ReadFile(abs)
	broken := strings.Replace(string(raw), "<!-- xm:msg-end -->", "", 1)
	if err := os.WriteFile(abs, []byte(broken), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.LoadChat(convID); err == nil {
		t.Fatal("缺 xm:msg-end 应当报错")
	}
	var parseErr string
	svc.DB.QueryRow(`SELECT parse_error FROM conversations WHERE id=?`, convID).Scan(&parseErr)
	if !strings.Contains(parseErr, "msg-end") {
		t.Fatalf("索引里应记下失败原因: %q", parseErr)
	}

	// 关键：解析失败时**不能覆盖用户的文件**
	after, _ := os.ReadFile(abs)
	if string(after) != broken {
		t.Fatal("解析失败时不该改写文件")
	}
	// 追加消息也必须失败，而不是把坏文件重写成只剩新消息
	if _, err := svc.AppendChatMessage(convID, "user", "新的", ""); err == nil {
		t.Fatal("文件损坏时不应允许写入")
	}
	after2, _ := os.ReadFile(abs)
	if string(after2) != broken {
		t.Fatal("失败路径也不该改写文件")
	}
}

func TestChatFile_UpdateDeleteClear(t *testing.T) {
	svc, _, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	m1, _ := svc.AppendChatMessage(convID, "user", "一", "")
	m2, _ := svc.AppendChatMessage(convID, "assistant", "二", "")

	if err := svc.UpdateChatMessage(convID, m1.ID, "一（改）", ""); err != nil {
		t.Fatal(err)
	}
	doc, _ := svc.LoadChat(convID)
	if doc.Messages[0].Content != "一（改）" {
		t.Fatalf("更新未生效: %q", doc.Messages[0].Content)
	}

	if err := svc.DeleteChatMessage(convID, m2.ID); err != nil {
		t.Fatal(err)
	}
	doc, _ = svc.LoadChat(convID)
	if len(doc.Messages) != 1 {
		t.Fatalf("删除后应剩 1 条，实得 %d", len(doc.Messages))
	}

	if err := svc.ClearChatMessages(convID); err != nil {
		t.Fatal(err)
	}
	doc, _ = svc.LoadChat(convID)
	if len(doc.Messages) != 0 {
		t.Fatalf("清空后应无消息，实得 %d", len(doc.Messages))
	}

	// 清空后再追加，id 从 1 重新开始（文件里没有别的消息）
	m3, err := svc.AppendChatMessage(convID, "user", "三", "")
	if err != nil {
		t.Fatal(err)
	}
	if m3.ID != 1 {
		t.Fatalf("清空后第一条消息 id 应为 1，实得 %d", m3.ID)
	}
}

// 改标题要连同文件一起改：文件名里的标题段始终等于当前标题（与章节、图谱一致）。
//
// 时间前缀与短码保持不动，所以"身份"是稳定的；用户改标题不会让对话变成一个新文件。
func TestChatFile_RenameFollowsFileName(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	var before string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&before)
	if before == "" {
		t.Fatal("建文件时应写回 file_path")
	}
	if err := svc.RenameChat(convID, "新标题"); err != nil {
		t.Fatal(err)
	}
	var after string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&after)
	if filepath.Base(filepath.FromSlash(after)) != "new" && !strings.Contains(after, "新标题") {
		t.Fatalf("文件名应跟随新标题: %q", after)
	}
	if after == before {
		t.Fatalf("跑题了：这条用例要求改名后文件名变化")
	}
	if filepath.Dir(after) != filepath.Dir(before) {
		t.Fatalf("只该改文件名，不该换目录: %q → %q", before, after)
	}
	// 时间前缀与短码不变
	tailOld := strings.Split(strings.TrimSuffix(filepath.Base(before), ".md"), "-")
	tailNew := strings.Split(strings.TrimSuffix(filepath.Base(after), ".md"), "-")
	if tailOld[0]+"-"+tailOld[1] != tailNew[0]+"-"+tailNew[1] {
		t.Fatalf("时间前缀不该变: %q → %q", before, after)
	}
	if tailOld[len(tailOld)-1] != tailNew[len(tailNew)-1] {
		t.Fatalf("短码是稳定标识，不该变: %q → %q", before, after)
	}
	// 旧文件应已不在，新文件在
	if _, err := os.Stat(filepath.Join(root, "对话落盘", filepath.FromSlash(before))); !os.IsNotExist(err) {
		t.Fatal("旧文件名的文件不该还留着")
	}
	if _, err := os.Stat(filepath.Join(root, "对话落盘", filepath.FromSlash(after))); err != nil {
		t.Fatalf("新文件名的文件不存在: %v", err)
	}
	doc, err := svc.LoadChat(convID)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Title != "新标题" {
		t.Fatalf("标题未写入 front-matter: %q", doc.Title)
	}
}

func TestChatFile_CustomFrontMatterKept(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	abs := filepath.Join(root, "对话落盘", filepath.FromSlash(rel))

	raw, _ := os.ReadFile(abs)
	custom := strings.Replace(string(raw), "---\n\n", "我的标签: 手写\n---\n\n", 1)
	// 上面的替换可能不命中，直接手工插到第一个 --- 之后更稳
	if custom == string(raw) {
		lines := strings.SplitN(string(raw), "\n", 2)
		custom = lines[0] + "\n我的标签: 手写\n" + lines[1]
	}
	if err := os.WriteFile(abs, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.AppendChatMessage(convID, "user", "触发重写", ""); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(abs)
	if !strings.Contains(string(after), "我的标签: 手写") {
		t.Fatalf("重写时应保留用户自定义字段:\n%s", string(after))
	}
}

func TestChatFile_ConcurrentAppend(t *testing.T) {
	svc, _, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errs := make([]error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, err := svc.AppendChatMessage(convID, "user", "并发消息", "")
			errs[n] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("第 %d 条并发追加失败: %v", i, err)
		}
	}
	doc, err := svc.LoadChat(convID)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Messages) != 10 {
		t.Fatalf("并发追加应一条不漏，实得 %d", len(doc.Messages))
	}
	ids := map[int]bool{}
	for _, m := range doc.Messages {
		if ids[m.ID] {
			t.Fatalf("消息 id 重复: %d", m.ID)
		}
		ids[m.ID] = true
	}
}

func TestChatFile_Trash(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	abs := filepath.Join(root, "对话落盘", filepath.FromSlash(rel))

	if err := svc.TrashChatFile(convID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatal("回收后原位置不应还有文件")
	}
}

// 首次访问老对话（file_path 为空）时应定下文件名并写回索引，之后再访问不变。
func TestChatFile_BackfillPathIsStable(t *testing.T) {
	svc, _, _, convID := newChatBook(t)
	svc.DB.Exec(`UPDATE conversations SET file_path='' WHERE id=?`, convID)

	_, first, err := svc.ChatFilePath(convID)
	if err != nil {
		t.Fatal(err)
	}
	if first == "" {
		t.Fatal("应生成 file_path")
	}
	_, second, err := svc.ChatFilePath(convID)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("file_path 应稳定: %q → %q", first, second)
	}
	var stored string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&stored)
	if stored != first {
		t.Fatalf("应写回索引: %q ≠ %q", stored, first)
	}
}

// 数据库时间戳是 UTC（CURRENT_TIMESTAMP），文件里写的是本地时间。
// 不做换算的话，同一文件里会出现"created_at 比消息时间早 8 小时"的怪相。
func TestChatFile_TimeIsLocal(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	// 造一条明确写成 UTC 的记录（模拟 SQLite CURRENT_TIMESTAMP）
	svc.DB.Exec(`UPDATE conversations SET created_at=?, updated_at=? WHERE id=?`,
		"2026-10-18 15:48:01", "2026-10-18 15:48:01", convID)

	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	// 文件名里的时间应是本地时间（UTC+8 → 23:48），而不是 15:48
	if !strings.Contains(rel, "20261018-2348") {
		t.Fatalf("文件名应按本地时间编码，实得 %q", rel)
	}
	abs := filepath.Join(root, "对话落盘", filepath.FromSlash(rel))
	raw, _ := os.ReadFile(abs)
	if !strings.Contains(string(raw), "created_at: \"2026-10-18 23:48:01\"") {
		t.Fatalf("front-matter 的 created_at 应是本地时间:\n%s", string(raw))
	}

	// 往返：读回来的时间应与写入的一致（不因解析偏好而漂移）
	doc, err := svc.LoadChat(convID)
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatMemTime(doc.CreatedAt); got != "2026-10-18 23:48:01" {
		t.Fatalf("时间往返不一致: %s", got)
	}
}

// 成对标记结尾要留空行，否则紧邻的 `## 助手` 标题会贴住 `-end`。
func TestChatFile_BlankLineBetweenBlocks(t *testing.T) {
	svc, root, _, convID := newChatBook(t)
	if err := svc.CreateChatFile(convID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendChatMessage(convID, "user", "第一条", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendChatMessage(convID, "assistant", "第二条", ""); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE id=?`, convID).Scan(&rel)
	raw, _ := os.ReadFile(filepath.Join(root, "对话落盘", filepath.FromSlash(rel)))
	if strings.Contains(string(raw), "<!-- xm:msg-end -->\n##") {
		t.Fatalf("结束标记与下一个标题之间应有空行:\n%s", string(raw))
	}
}
