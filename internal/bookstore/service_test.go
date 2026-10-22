package bookstore

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/database"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// setupService 建一个临时库，并把 books_root 指到临时目录，避免写到仓库里。
func setupService(t *testing.T) (*Service, string) {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "svc.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	return NewService(db), root
}

func TestCreateBook_MakesFolderAndJSON(t *testing.T) {
	svc, root := setupService(t)

	b := &models.Book{Title: "太荒神诀", Status: "draft", Tags: models.StringList{"玄幻"}}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	if b.ID == 0 || b.FolderPath != "太荒神诀" {
		t.Fatalf("unexpected book: %+v", b)
	}

	dir := filepath.Join(root, "太荒神诀")
	for _, sub := range bookSubdirs {
		if st, err := os.Stat(filepath.Join(dir, sub)); err != nil || !st.IsDir() {
			t.Fatalf("子目录 %s 未创建", sub)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "book.json"))
	if err != nil {
		t.Fatalf("book.json 未写入: %v", err)
	}
	if !strings.Contains(string(raw), `"title": "太荒神诀"`) {
		t.Fatalf("book.json 内容不对: %s", raw)
	}

	// 同名书应得到 (2) 后缀，而不是覆盖
	b2 := &models.Book{Title: "太荒神诀", Status: "draft"}
	if err := svc.CreateBook(b2); err != nil {
		t.Fatal(err)
	}
	if b2.FolderPath != "太荒神诀 (2)" {
		t.Fatalf("重名目录应加后缀，实得 %q", b2.FolderPath)
	}
}

func TestCreateChapter_WritesFileAndIndex(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "测试书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}

	c := &models.Chapter{BookID: b.ID, Title: "第一章 落雪", Content: "正文第一段\n第二段", Summary: "主角登场"}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}
	if c.FilePath != "001 第一章 落雪.md" {
		t.Fatalf("文件名不对: %q", c.FilePath)
	}
	raw, err := os.ReadFile(filepath.Join(root, "测试书", c.FilePath))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(raw), "---\nsummary: 主角登场\n---") {
		t.Fatalf("front-matter 未写入: %q", raw)
	}
	if !strings.Contains(string(raw), "正文第一段") {
		t.Fatalf("正文未写入: %q", raw)
	}

	text, err := svc.ChapterText(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if text != "正文第一段\n第二段" {
		t.Fatalf("读回正文不对: %q", text)
	}
}

func TestUpdateChapter_RenamesAndKeepsCustomFields(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	c := &models.Chapter{BookID: b.ID, Title: "旧标题", Content: "旧正文", Summary: "旧摘要"}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}

	// 手工塞一个自定义字段，模拟用户在 Obsidian 里加的
	path := filepath.Join(root, "书", c.FilePath)
	raw, _ := os.ReadFile(path)
	custom := strings.Replace(string(raw), "---\nsummary: 旧摘要", "---\nmood: 紧张\nsummary: 旧摘要", 1)
	if err := os.WriteFile(path, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	// 外部改完文件后，应用要先"读一次"把索引 hash 对齐（真实流程也是如此），
	// 否则下面的保存会被冲突检测拦下 —— 这是预期行为，不是 bug。
	if _, err := svc.ChapterText(c.ID); err != nil {
		t.Fatal(err)
	}

	c.Title = "新标题"
	c.Content = "新正文"
	c.Summary = "新摘要"
	if err := svc.UpdateChapter(c); err != nil {
		t.Fatal(err)
	}
	if c.FilePath != "001 新标题.md" {
		t.Fatalf("标题变化应重命名文件，实得 %q", c.FilePath)
	}
	if _, err := os.Stat(filepath.Join(root, "书", "001 旧标题.md")); !os.IsNotExist(err) {
		t.Fatal("旧文件应被删除")
	}
	got, _ := os.ReadFile(filepath.Join(root, "书", c.FilePath))
	if !strings.Contains(string(got), "mood: 紧张") {
		t.Fatalf("自定义字段被抹掉了: %s", got)
	}
	if !strings.Contains(string(got), "summary: 新摘要") || !strings.Contains(string(got), "新正文") {
		t.Fatalf("内容未更新: %s", got)
	}
}

func TestSetChapterText_PreservesSummary(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	c := &models.Chapter{BookID: b.ID, Title: "章", Content: "原文", Summary: "摘要"}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}

	if err := svc.SetChapterText(c.ID, "追加后的正文"); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(root, "书", c.FilePath))
	if !strings.Contains(string(raw), "summary: 摘要") {
		t.Fatalf("摘要应保留: %s", raw)
	}
	if !strings.Contains(string(raw), "追加后的正文") {
		t.Fatalf("正文未更新: %s", raw)
	}
}

func TestRenameBook_RenamesDir(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "旧书名", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	rel, err := svc.RenameBook(b.ID, "新书名")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "新书名" {
		t.Fatalf("got %q", rel)
	}
	if _, err := os.Stat(filepath.Join(root, "新书名", "book.json")); err != nil {
		t.Fatalf("新目录缺少 book.json: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "旧书名")); !os.IsNotExist(err) {
		t.Fatal("旧目录应已改名")
	}
}

func TestTrashChapterFile(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	c := &models.Chapter{BookID: b.ID, Title: "待删", Content: "x"}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}

	if err := svc.TrashChapterFile(b.ID, c.FilePath); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "书", c.FilePath)); !os.IsNotExist(err) {
		t.Fatal("原文件应已移走")
	}
	// Windows 上走系统回收站（不在 _trash）；其他平台应回退到 _trash。
	if runtime.GOOS != "windows" {
		var found bool
		filepath.Walk(filepath.Join(root, "_trash"), func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, "001 待删.md") {
				found = true
			}
			return nil
		})
		if !found {
			t.Fatal("回收目录里找不到文件")
		}
	}
}

func TestSwitchRootMovesBooks(t *testing.T) {
	svc, oldRoot := setupService(t)
	b := &models.Book{Title: "切换书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	newRoot := t.TempDir()
	if err := svc.SwitchRoot(newRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(newRoot, "切换书", "book.json")); err != nil {
		t.Fatalf("作品未搬到新目录: %v", err)
	}
	if _, err := os.Stat(filepath.Join(oldRoot, "切换书")); !os.IsNotExist(err) {
		t.Fatal("旧目录里应已清空")
	}
	dir, err := svc.BookDir(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dir, newRoot) {
		t.Fatalf("BookDir 仍指向旧根: %s", dir)
	}
}
