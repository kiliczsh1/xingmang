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

// M1 端到端：对话消息经过 handler 写入后必须落到 memory/chats/*.md，
// 而且**读回来也走文件**（`conversation_messages` 不再参与）。
//
// 这一条覆盖的是"接口层面真的换了存储"，单测只测 service 是测不出漏改 handler 的。

type convResp struct {
	Success bool `json:"success"`
	Data    struct {
		ID       int    `json:"id"`
		FilePath string `json:"file_path"`
	} `json:"data"`
	Message string `json:"message"`
}

type msgListResp struct {
	Success bool `json:"success"`
	Data    []struct {
		ID             int     `json:"id"`
		Role           string  `json:"role"`
		Content        string  `json:"content"`
		DisplayContent *string `json:"display_content"`
	} `json:"data"`
	Message string `json:"message"`
}

func newConvTestHandler(t *testing.T) (*Handler, string, int) {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "conv.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)

	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"对话书","status":"draft"}`)))
	var b struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil || b.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	return h, root, b.Data.ID
}

func TestConversationsGoThroughFile(t *testing.T) {
	h, root, bookID := newConvTestHandler(t)

	rec := httptest.NewRecorder()
	h.CreateConversation(rec, httptest.NewRequest("POST", "/api/conversations",
		strings.NewReader(`{"book_id":`+strconv.Itoa(bookID)+`,"title":"列车开场"}`)))
	var conv convResp
	if err := json.Unmarshal(rec.Body.Bytes(), &conv); err != nil || !conv.Success {
		t.Fatalf("建对话失败: %s", rec.Body.String())
	}
	convID := conv.Data.ID
	if !strings.HasPrefix(conv.Data.FilePath, "memory/chats/") {
		t.Fatalf("建对话应立即返回 file_path: %q", conv.Data.FilePath)
	}
	abs := filepath.Join(root, "对话书", filepath.FromSlash(conv.Data.FilePath))
	if _, err := os.Stat(abs); err != nil {
		t.Fatalf("对话文件未创建: %v", err)
	}

	// 写入两条消息（第二条带驼峰 displayContent —— 前端历史上就是这么发的）
	rec = httptest.NewRecorder()
	h.SaveMessage(rec, setPathValue(httptest.NewRequest("POST", "/api/conversations/1/messages",
		strings.NewReader(`{"role":"user","content":"原始提问","displayContent":"用户看到的问题"}`)), "id", strconv.Itoa(convID)))
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("写消息失败: %s", rec.Body.String())
	}
	rec = httptest.NewRecorder()
	h.SaveMessage(rec, setPathValue(httptest.NewRequest("POST", "/api/conversations/1/messages",
		strings.NewReader(`{"role":"assistant","content":"模型的回答"}`)), "id", strconv.Itoa(convID)))
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("写第二条消息失败: %s", rec.Body.String())
	}

	// 库里不应留下任何消息行：消息只在文件里
	var dbCount int
	h.db.QueryRow(`SELECT COUNT(*) FROM conversation_messages WHERE conversation_id=?`, convID).Scan(&dbCount)
	if dbCount != 0 {
		t.Fatalf("conversation_messages 不应再写入，实得 %d 行", dbCount)
	}

	// 文件内容校验
	raw, _ := os.ReadFile(abs)
	text := string(raw)
	for _, want := range []string{"kind: chat", "<!-- xm:msg ", "原始提问", "模型的回答", "用户看到的问题"} {
		if !strings.Contains(text, want) {
			t.Fatalf("文件里缺少 %q:\n%s", want, text)
		}
	}

	// 读回来：正文与 display 都要对
	rec = httptest.NewRecorder()
	h.GetMessages(rec, setPathValue(httptest.NewRequest("GET", "/api/conversations/1/messages", nil), "id", strconv.Itoa(convID)))
	var list msgListResp
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil || !list.Success {
		t.Fatalf("读消息失败: %s", rec.Body.String())
	}
	if len(list.Data) != 2 {
		t.Fatalf("应有 2 条消息，实得 %d: %s", len(list.Data), rec.Body.String())
	}
	if list.Data[0].Content != "原始提问" {
		t.Fatalf("第 1 条正文不对: %q", list.Data[0].Content)
	}
	if list.Data[0].DisplayContent == nil || *list.Data[0].DisplayContent != "用户看到的问题" {
		t.Fatalf("displayContent 没被接住（前端发的驼峰字段）: %+v", list.Data[0].DisplayContent)
	}

	// 外部改文件 → 接口读回来就是改后的内容
	edited := strings.Replace(text, "模型的回答", "我在外部改成了这句", 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	h.GetMessages(rec, setPathValue(httptest.NewRequest("GET", "/api/conversations/1/messages", nil), "id", strconv.Itoa(convID)))
	list = msgListResp{}
	json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list.Data) != 2 || list.Data[1].Content != "我在外部改成了这句" {
		t.Fatalf("手改文件未生效: %s", rec.Body.String())
	}

	// 标记被改坏 → 409，且不改写文件
	broken := strings.Replace(edited, "<!-- xm:msg-end -->", "", 1)
	os.WriteFile(abs, []byte(broken), 0o644)
	rec = httptest.NewRecorder()
	h.GetMessages(rec, setPathValue(httptest.NewRequest("GET", "/api/conversations/1/messages", nil), "id", strconv.Itoa(convID)))
	if rec.Code != 409 {
		t.Fatalf("文件损坏应返回 409，实得 %d: %s", rec.Code, rec.Body.String())
	}
	after, _ := os.ReadFile(abs)
	if string(after) != broken {
		t.Fatal("解析失败时不得改写用户的文件")
	}
	var parseErr string
	h.db.QueryRow(`SELECT parse_error FROM conversations WHERE id=?`, convID).Scan(&parseErr)
	if parseErr == "" {
		t.Fatal("索引里应记下解析失败原因")
	}

	// 改回来后恢复正常
	os.WriteFile(abs, []byte(edited), 0o644)
	rec = httptest.NewRecorder()
	h.GetMessages(rec, setPathValue(httptest.NewRequest("GET", "/api/conversations/1/messages", nil), "id", strconv.Itoa(convID)))
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("改回来后应恢复正常: %s", rec.Body.String())
	}
	h.db.QueryRow(`SELECT parse_error FROM conversations WHERE id=?`, convID).Scan(&parseErr)
	if parseErr != "" {
		t.Fatalf("恢复正常后 parse_error 应清空: %q", parseErr)
	}

	// 删除对话 → 文件进回收站
	rec = httptest.NewRecorder()
	h.DeleteConversation(rec, setPathValue(httptest.NewRequest("DELETE", "/api/conversations/1", nil), "id", strconv.Itoa(convID)))
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("删除失败: %s", rec.Body.String())
	}
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatal("删除对话后原位置不应还有文件")
	}
}
