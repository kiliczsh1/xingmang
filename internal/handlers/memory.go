package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// 拆书与生成历史的接口（docs/books-memory-storage.md §11.5）。
//
// 这两类记录原先**只存在浏览器 localStorage 里**：既没有数据库行，也没有文件。
// 清了浏览器缓存就没了，换台机器也带不走。现在正文落在 `memory/analysis/` 与
// `memory/history/YYYY-MM/`，数据库只留索引。
//
// 记录（拆书/生成历史）与对话、Agent 会话的不同点：这里的写操作都来自界面上的明确动作，
// 所以外部改动冲突会**返回错误**而不是静默跳过 —— 用户点了保存却没保存上，必须当场知道。

// maxFullRecords 是"带正文的列表"一次最多读多少条文件。
//
// 读文件比只读索引慢得多，给个上限防止历史攒到几百条时列表变卡。
// 超出时列表照常返回，只是不带正文 —— 点开某一条仍会按需读它那一个文件。
const maxFullRecords = 200

// recordRequest 是新建 / 更新记录的请求体。
//
// payload 是原始记录 JSON（界面上那条记录的完整结构）。之所以整段收下：
// 四类历史（拆书 / 续写改写 / 创作 / 插件）字段各不相同，在服务端逐一映射
// 既容易漏字段，又会在前端加字段时静默丢数据。
type recordRequest struct {
	Kind       string          `json:"kind"`
	ExternalID string          `json:"external_id"`
	Title      string          `json:"title"`
	Source     string          `json:"source"`
	Status     string          `json:"status"`
	Prompt     string          `json:"prompt"`
	Chapters   []int           `json:"chapters"`
	Result     string          `json:"result"`
	Readable   string          `json:"readable"`
	Payload    json.RawMessage `json:"payload"`
	CreatedAt  int64           `json:"created_at_ms"` // 毫秒时间戳（前端 Date.now()）
}

func (r *recordRequest) toInput() bookstore.MemoryRecordInput {
	in := bookstore.MemoryRecordInput{
		Kind:       strings.TrimSpace(r.Kind),
		ExternalID: strings.TrimSpace(r.ExternalID),
		Title:      strings.TrimSpace(r.Title),
		Source:     strings.TrimSpace(r.Source),
		Status:     strings.TrimSpace(r.Status),
		Prompt:     strings.TrimSpace(r.Prompt),
		Chapters:   r.Chapters,
		Result:     r.Result,
		Readable:   r.Readable,
	}
	if len(r.Payload) > 0 {
		in.Payload = string(r.Payload)
	}
	if r.CreatedAt > 0 {
		in.CreatedAt = time.UnixMilli(r.CreatedAt)
	}
	return in
}

// recordIDParam 取路径里的记录 id。
//
// 复用 idParam 而不是自己读 PathValue：那个函数带 context 回退，
// 走 dispatch 分发（非 Go 1.22 路由）时也能取到值。
func recordIDParam(r *http.Request) (int, bool) { return idParam(r, "id") }

// GetBookMemories 列出某本书的记录（拆书 / 生成历史）。
// GET /api/books/{id}/memories?kind=analysis|generation
func (h *Handler) GetBookMemories(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	if kind != "" && kind != bookstore.MemoryKindAnalysis && kind != bookstore.MemoryKindGeneration {
		respondFail(w, 400, "kind 只能是 analysis 或 generation")
		return
	}
	svc := bookstore.NewService(h.db)
	list, err := svc.ListRecords(bookID, kind)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []bookstore.MemoryRecord{}
	}

	// full=1 时把正文与 payload 一起带上：界面要还原"那条记录原本长什么样"，
	// 而记录的结构在文件的 xm:payload 块里。默认不带，列表页就不必读文件。
	if r.URL.Query().Get("full") == "1" && len(list) <= maxFullRecords {
		for i := range list {
			doc, derr := svc.MemoryRecordFile(list[i].ID)
			if derr != nil {
				continue
			}
			list[i].Result = doc.Result
			list[i].Readable = doc.Readable
			list[i].Payload = doc.Payload
			list[i].Prompt = doc.Prompt
			list[i].Chapters = doc.Chapters
		}
	}
	respondOK(w, list)
}

// GetAllMemories 列出**全部作品**的记录，供历史中心跨作品聚合。
// GET /api/memories?kind=analysis|generation&full=1
func (h *Handler) GetAllMemories(w http.ResponseWriter, r *http.Request) {
	kind := strings.TrimSpace(r.URL.Query().Get("kind"))
	if kind != "" && kind != bookstore.MemoryKindAnalysis && kind != bookstore.MemoryKindGeneration {
		respondFail(w, 400, "kind 只能是 analysis 或 generation")
		return
	}
	svc := bookstore.NewService(h.db)
	list, err := svc.ListAllRecords(kind)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []bookstore.MemoryRecord{}
	}
	if r.URL.Query().Get("full") == "1" && len(list) <= maxFullRecords {
		for i := range list {
			doc, derr := svc.MemoryRecordFile(list[i].ID)
			if derr != nil {
				continue
			}
			list[i].Result = doc.Result
			list[i].Readable = doc.Readable
			list[i].Payload = doc.Payload
			list[i].Prompt = doc.Prompt
			list[i].Chapters = doc.Chapters
		}
	}
	respondOK(w, list)
}

// CreateMemoryRecord 新建记录（写文件 + 建索引）。
// POST /api/books/{id}/memories
func (h *Handler) CreateMemoryRecord(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var req recordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	rec, err := bookstore.NewService(h.db).CreateRecord(bookID, req.toInput())
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, rec)
}

// GetMemoryRecord 读单条记录：正文以文件为准。
// GET /api/memories/{id}
func (h *Handler) GetMemoryRecord(w http.ResponseWriter, r *http.Request) {
	id, ok := recordIDParam(r)
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	doc, err := bookstore.NewService(h.db).LoadRecord(id)
	if err != nil {
		// 解析失败要如实报出原因，别让界面以为是"没有这条"
		respondFail(w, 422, err.Error())
		return
	}
	respondOK(w, map[string]any{
		"id":          id,
		"kind":        doc.Kind,
		"external_id": doc.ExternalID,
		"title":       doc.Title,
		"source":      doc.Source,
		"status":      doc.Status,
		"prompt":      doc.Prompt,
		"chapters":    doc.Chapters,
		"result":      doc.Result,
		"readable":    doc.Readable,
		"payload":     doc.Payload,
		"file_path":   doc.RelPath,
		"created_at":  bookstore.FormatMemTime(doc.CreatedAt),
	})
}

// UpdateMemoryRecord 覆盖记录内容。
// PUT /api/memories/{id}
func (h *Handler) UpdateMemoryRecord(w http.ResponseWriter, r *http.Request) {
	id, ok := recordIDParam(r)
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var req recordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	rec, err := bookstore.NewService(h.db).UpdateRecord(id, req.toInput())
	if err != nil {
		if errors.Is(err, bookstore.MemoryRecordConflict) {
			// 409：文件被外部改过。不覆盖，也不假装成功。
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success": false,
				"code":    "file_conflict",
				"message": bookstore.ExternalEditReason,
			})
			return
		}
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, rec)
}

// DeleteMemoryRecord 删除记录：文件进回收站。
// DELETE /api/memories/{id}
func (h *Handler) DeleteMemoryRecord(w http.ResponseWriter, r *http.Request) {
	id, ok := recordIDParam(r)
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	if err := bookstore.NewService(h.db).TrashRecord(id); err != nil {
		respondFail(w, 500, "删除失败: "+err.Error())
		return
	}
	respondOK(w, "已移入回收站")
}

// GetBookMemorySummary 返回「记忆」概览：各类条数、缺失、解析失败、待写入。
// GET /api/books/{id}/memory/summary
func (h *Handler) GetBookMemorySummary(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	svc := bookstore.NewService(h.db)
	counts := map[string]int{}
	add := func(kind, countQuery string, args ...any) {
		var n int
		if err := h.db.QueryRow(countQuery, args...).Scan(&n); err == nil {
			counts[kind] = n
		}
	}
	add("chats", `SELECT COUNT(*) FROM conversations WHERE book_id=?`, bookID)
	add("agents", `SELECT COUNT(*) FROM agent_sessions WHERE scope_kind='book' AND scope_id=?`, bookID)
	add("graph_versions", `SELECT COUNT(*) FROM graph_versions WHERE book_id=?`, bookID)
	if byKind, err := svc.CountRecordsByKind(bookID); err == nil {
		for k, v := range byKind {
			counts[k] = v
		}
	}

	var missing, parseErrors int
	missing += countScalar(h.db, `SELECT COUNT(*) FROM chapters WHERE book_id=? AND missing=1`, bookID)
	missing += countScalar(h.db, `SELECT COUNT(*) FROM conversations WHERE book_id=? AND missing=1`, bookID)
	missing += countScalar(h.db, `SELECT COUNT(*) FROM agent_sessions WHERE scope_kind='book' AND scope_id=? AND missing=1`, bookID)
	missing += countScalar(h.db, `SELECT COUNT(*) FROM graph_versions WHERE book_id=? AND missing=1`, bookID)
	missing += countScalar(h.db, `SELECT COUNT(*) FROM memory_records WHERE book_id=? AND missing=1`, bookID)

	parseErrors += countScalar(h.db, `SELECT COUNT(*) FROM conversations WHERE book_id=? AND parse_error<>''`, bookID)
	parseErrors += countScalar(h.db, `SELECT COUNT(*) FROM agent_sessions WHERE scope_kind='book' AND scope_id=? AND parse_error<>''`, bookID)
	parseErrors += countScalar(h.db, `SELECT COUNT(*) FROM graph_versions WHERE book_id=? AND parse_error<>''`, bookID)
	parseErrors += countScalar(h.db, `SELECT COUNT(*) FROM memory_records WHERE book_id=? AND parse_error<>''`, bookID)

	respondOK(w, map[string]any{
		"counts":               counts,
		"missing":              missing,
		"parse_errors":         parseErrors,
		"pending_writes":       bookstore.PendingMemoryWrites(),
		"write_failures":       bookstore.MemoryWriteFailures(),
		"external_edit_reason": bookstore.ExternalEditReason,
	})
}

// countScalar 查一个计数；查不到按 0 算（统计失败不该让整个概览接口 500）。
func countScalar(db *sql.DB, query string, args ...any) int {
	var n int
	if err := db.QueryRow(query, args...).Scan(&n); err != nil {
		return 0
	}
	return n
}
