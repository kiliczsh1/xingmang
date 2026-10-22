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

// M4 端到端：拆书 / 生成历史经过接口后必须落到 memory/ 下的真实文件，
// 而且读回来以文件为准。
//
// 这一类记录原来**只存在浏览器 localStorage 里**：既没有库行，也没有文件。
// 所以这里重点验三件事：点是点、文件真写了、读回来还是那条记录。

type recordResp struct {
	Success bool `json:"success"`
	Data    struct {
		ID       int    `json:"id"`
		Kind     string `json:"kind"`
		Title    string `json:"title"`
		FilePath string `json:"file_path"`
		Result   string `json:"result"`
		Payload  string `json:"payload"`
	} `json:"data"`
	Message string `json:"message"`
}

func newMemoryTestHandler(t *testing.T) (*Handler, string, int) {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), "memory.db"))
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
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"拆书书","status":"draft"}`)))
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

func TestMemoryRecordsGoThroughFile(t *testing.T) {
	h, root, bookID := newMemoryTestHandler(t)

	// 1) 新建：正文落盘
	body := `{"kind":"analysis","external_id":"1729262400000","title":"黄金三章拆解","source":"拆书",
		"status":"completed","prompt":"黄金三章拆解","chapters":[1,2,3],
		"result":"第一处问题是节奏。","readable":"### system · 提示词\n\n你是资深编辑。",
		"payload":{"id":1729262400000,"messages":[{"role":"user","content":"拆一下"}]},
		"created_at_ms":1729262400000}`
	rec := httptest.NewRecorder()
	req := setPathValue(
		httptest.NewRequest("POST", "/api/books/"+itoaHandler(bookID)+"/memories", strings.NewReader(body)),
		"id", itoaHandler(bookID))
	h.CreateMemoryRecord(rec, req)

	var created recordResp
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || !created.Success {
		t.Fatalf("新建记录失败: %s", rec.Body.String())
	}
	if created.Data.FilePath == "" || !strings.HasPrefix(created.Data.FilePath, "memory/analysis/") {
		t.Fatalf("拆书应落在 memory/analysis/: %q", created.Data.FilePath)
	}
	abs := filepath.Join(root, "拆书书", filepath.FromSlash(created.Data.FilePath))
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("文件没写出来: %v", err)
	}
	if !strings.Contains(string(raw), "第一处问题是节奏。") {
		t.Fatalf("文件里没有结果正文:\n%s", string(raw))
	}
	if !strings.Contains(string(raw), "```json") {
		t.Fatalf("文件里没有 payload 块:\n%s", string(raw))
	}

	// 2) 列表：full=1 时带正文与 payload，界面才能还原原记录
	listRec := httptest.NewRecorder()
	listReq := setPathValue(
		httptest.NewRequest("GET", "/api/books/"+itoaHandler(bookID)+"/memories?kind=analysis&full=1", nil),
		"id", itoaHandler(bookID))
	h.GetBookMemories(listRec, listReq)
	if !strings.Contains(listRec.Body.String(), "第一处问题是节奏。") {
		t.Fatalf("带正文的列表没返回结果: %s", listRec.Body.String())
	}
	if !strings.Contains(listRec.Body.String(), "\\\"messages\\\"") {
		t.Fatalf("带正文的列表没返回 payload: %s", listRec.Body.String())
	}

	// 3) 单条：读的是文件
	getRec := httptest.NewRecorder()
	getReq := setPathValue(httptest.NewRequest("GET", "/api/memories/1", nil), "id", itoaHandler(created.Data.ID))
	h.GetMemoryRecord(getRec, getReq)
	var got recordResp
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil || !got.Success {
		t.Fatalf("读单条失败: %s", getRec.Body.String())
	}
	if !strings.Contains(got.Data.Payload, "messages") {
		t.Fatalf("payload 没还原: %q", got.Data.Payload)
	}

	// 4) 外部改过文件后再保存 → 409，且不覆盖
	if err := os.WriteFile(abs, []byte(strings.Replace(string(raw), "第一处问题是节奏。", "我在外部改的", 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	updRec := httptest.NewRecorder()
	updReq := setPathValue(
		httptest.NewRequest("PUT", "/api/memories/1",
			strings.NewReader(`{"kind":"analysis","title":"黄金三章拆解","source":"拆书","result":"界面里的版本"}`)),
		"id", itoaHandler(created.Data.ID))
	h.UpdateMemoryRecord(updRec, updReq)
	if updRec.Code != 409 {
		t.Fatalf("外部改过之后保存应返回 409，实得 %d: %s", updRec.Code, updRec.Body.String())
	}
	after, _ := os.ReadFile(abs)
	if !strings.Contains(string(after), "我在外部改的") {
		t.Fatalf("外部改动被覆盖了:\n%s", string(after))
	}

	// 5) 删除：文件移走、索引消失
	delRec := httptest.NewRecorder()
	delReq := setPathValue(httptest.NewRequest("DELETE", "/api/memories/1", nil), "id", itoaHandler(created.Data.ID))
	h.DeleteMemoryRecord(delRec, delReq)
	if !strings.Contains(delRec.Body.String(), `"success":true`) {
		t.Fatalf("删除失败: %s", delRec.Body.String())
	}
	if _, err := os.Stat(abs); err == nil {
		t.Fatal("文件应已移入回收站")
	}
}

func TestMemorySummaryAndGlobalList(t *testing.T) {
	h, _, bookID := newMemoryTestHandler(t)

	body := `{"kind":"generation","external_id":"h1","title":"续写","source":"continuation",
		"result":"他把车票递了过去。","created_at_ms":1729262400000}`
	rec := httptest.NewRecorder()
	req := setPathValue(
		httptest.NewRequest("POST", "/api/books/"+itoaHandler(bookID)+"/memories", strings.NewReader(body)),
		"id", itoaHandler(bookID))
	h.CreateMemoryRecord(rec, req)
	if !strings.Contains(rec.Body.String(), `"success":true`) {
		t.Fatalf("新建失败: %s", rec.Body.String())
	}

	// 概览：应数出 1 条 generation
	sumRec := httptest.NewRecorder()
	sumReq := setPathValue(
		httptest.NewRequest("GET", "/api/books/"+itoaHandler(bookID)+"/memory/summary", nil),
		"id", itoaHandler(bookID))
	h.GetBookMemorySummary(sumRec, sumReq)
	if !strings.Contains(sumRec.Body.String(), "generation") {
		t.Fatalf("概览里应有 generation 计数: %s", sumRec.Body.String())
	}
	if !strings.Contains(sumRec.Body.String(), "external_edit_reason") {
		t.Fatalf("概览应带上外部改动的说明，供界面提示: %s", sumRec.Body.String())
	}

	// 跨作品列表：历史中心要用
	allRec := httptest.NewRecorder()
	allReq := httptest.NewRequest("GET", "/api/memories?kind=generation&full=1", nil)
	h.GetAllMemories(allRec, allReq)
	if !strings.Contains(allRec.Body.String(), "续写") {
		t.Fatalf("跨作品列表没返回记录: %s", allRec.Body.String())
	}
	if !strings.Contains(allRec.Body.String(), "拆书书") {
		t.Fatalf("跨作品列表应带上书名，供按作品筛选: %s", allRec.Body.String())
	}
}

func itoaHandler(v int) string { return strconv.Itoa(v) }
