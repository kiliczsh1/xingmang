package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/database"
)

// 作品文件夹化的端到端冒烟：建书 → 建章 → 读列表 → 改标题 → 读书架字数。
// 这条覆盖 bookSelectCols 的列顺序（scanBook 对位置极其敏感）。
func TestBooksFolderLifecycle(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "books.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)

	// 1. 建书
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"端到端测试","status":"draft"}`))
	h.CreateBook(rec, req)
	var created struct {
		Success bool `json:"success"`
		Data    struct {
			ID         int    `json:"id"`
			FolderPath string `json:"folder_path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || !created.Success {
		t.Fatalf("创建失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID
	if created.Data.FolderPath != "端到端测试" {
		t.Fatalf("folder_path 不对: %q", created.Data.FolderPath)
	}
	if _, err := os.Stat(filepath.Join(root, "端到端测试", "book.json")); err != nil {
		t.Fatalf("book.json 缺失: %v", err)
	}

	// 2. 建章
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/chapters",
		strings.NewReader(`{"book_id":`+strconv.Itoa(bookID)+`,"title":"第一章","content":"你好世界","summary":"开头"}`))
	h.CreateChapter(rec, req)
	var chResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID       int    `json:"id"`
			FilePath string `json:"file_path"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &chResp); err != nil || !chResp.Success {
		t.Fatalf("建章失败: %s", rec.Body.String())
	}
	chapterID := chResp.Data.ID
	if chResp.Data.FilePath != "001 第一章.md" {
		t.Fatalf("file_path 不对: %q", chResp.Data.FilePath)
	}

	// 3. 列表：正文来自文件
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/chapters/book/"+strconv.Itoa(bookID), nil)
	req.SetPathValue("bookId", strconv.Itoa(bookID))
	h.GetChaptersByBook(rec, req)
	var list struct {
		Success bool `json:"success"`
		Data    []struct {
			ID      int    `json:"id"`
			Content string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || !list.Success || len(list.Data) != 1 {
		t.Fatalf("列表失败: %s", rec.Body.String())
	}
	if list.Data[0].Content != "你好世界" {
		t.Fatalf("正文应来自文件，实得 %q", list.Data[0].Content)
	}

	// 4. 改标题 + 正文
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/api/chapters/"+strconv.Itoa(chapterID),
		strings.NewReader(`{"title":"第一章 改名","content":"改过的正文","summary":"开头"}`))
	req.SetPathValue("id", strconv.Itoa(chapterID))
	h.UpdateChapter(rec, req)
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("更新失败: %s", rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(root, "端到端测试", "001 第一章 改名.md")); err != nil {
		t.Fatalf("改名后的文件不存在: %v", err)
	}

	// 5. 书架：字数按字符数聚合
	rec = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/books", nil)
	h.GetBooks(rec, req)
	var books struct {
		Success bool `json:"success"`
		Data    []struct {
			WordCount    int `json:"word_count"`
			ChapterCount int `json:"chapter_count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &books); err != nil || !books.Success || len(books.Data) != 1 {
		t.Fatalf("书架失败: %s", rec.Body.String())
	}
	if books.Data[0].ChapterCount != 1 || books.Data[0].WordCount != 5 {
		t.Fatalf("字数/章数不对: %+v", books.Data[0])
	}
}
