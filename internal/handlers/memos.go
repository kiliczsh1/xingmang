package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetMemos returns all memos ordered by is_pinned DESC, order_num ASC.
func (h *Handler) GetMemos(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, title, content, category, order_num, tags, is_pinned, word_count, created_at, updated_at FROM memos ORDER BY is_pinned DESC, order_num ASC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Memo
	for rows.Next() {
		var m models.Memo
		if err := rows.Scan(&m.ID, &m.Title, &m.Content, &m.Category, &m.OrderNum, &m.Tags, &m.IsPinned, &m.WordCount, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// SearchMemos searches memos by keyword and category.
func (h *Handler) SearchMemos(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	category := r.URL.Query().Get("category")

	query := "SELECT id, title, content, category, order_num, tags, is_pinned, word_count, created_at, updated_at FROM memos WHERE 1=1"
	args := []interface{}{}

	if keyword != "" {
		query += " AND (title LIKE ? OR content LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" {
		query += " AND category = ?"
		args = append(args, category)
	}
	query += " ORDER BY is_pinned DESC, order_num ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Memo
	for rows.Next() {
		var m models.Memo
		if err := rows.Scan(&m.ID, &m.Title, &m.Content, &m.Category, &m.OrderNum, &m.Tags, &m.IsPinned, &m.WordCount, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// GetMemo returns a single memo by id.
func (h *Handler) GetMemo(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var m models.Memo
	err := h.db.QueryRow("SELECT id, title, content, category, order_num, tags, is_pinned, word_count, created_at, updated_at FROM memos WHERE id = ?", id).
		Scan(&m.ID, &m.Title, &m.Content, &m.Category, &m.OrderNum, &m.Tags, &m.IsPinned, &m.WordCount, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "备忘录不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, m)
}

// CreateMemo creates a new memo.
func (h *Handler) CreateMemo(w http.ResponseWriter, r *http.Request) {
	var m models.Memo
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	m.WordCount = len(m.Content)

	result, err := h.db.Exec("INSERT INTO memos (title, content, category, order_num, tags, is_pinned, word_count) VALUES (?, ?, ?, ?, ?, ?, ?)",
		m.Title, m.Content, m.Category, m.OrderNum, m.Tags, m.IsPinned, m.WordCount)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	m.ID = int(id)
	respondOK(w, m)
}

// UpdateMemo updates an existing memo.
func (h *Handler) UpdateMemo(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var m models.Memo
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	m.WordCount = len(m.Content)

	_, err := h.db.Exec("UPDATE memos SET title=?, content=?, category=?, order_num=?, tags=?, is_pinned=?, word_count=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		m.Title, m.Content, m.Category, m.OrderNum, m.Tags, m.IsPinned, m.WordCount, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	m.ID = id
	respondOK(w, m)
}

// DeleteMemo deletes a memo by id.
func (h *Handler) DeleteMemo(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM memos WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// BatchMemosRequest is the request body for batch memo operations.
type BatchMemosRequest struct {
	Action string `json:"action"` // delete, update, toggle_pin
	IDs    []int  `json:"ids"`
	// Fields for update action
	IsPinned *int `json:"is_pinned,omitempty"`
}

// BatchMemos handles batch operations on memos.
func (h *Handler) BatchMemos(w http.ResponseWriter, r *http.Request) {
	var req BatchMemosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	if len(req.IDs) == 0 {
		respondFail(w, 400, "未提供ID列表")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	switch req.Action {
	case "delete":
		for _, id := range req.IDs {
			if _, err := tx.Exec("DELETE FROM memos WHERE id = ?", id); err != nil {
				respondFail(w, 500, err.Error())
				return
			}
		}
	case "update":
		for _, id := range req.IDs {
			if _, err := tx.Exec("UPDATE memos SET is_pinned=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.IsPinned, id); err != nil {
				respondFail(w, 500, err.Error())
				return
			}
		}
	case "toggle_pin":
		for _, id := range req.IDs {
			if _, err := tx.Exec("UPDATE memos SET is_pinned = CASE WHEN is_pinned = 1 THEN 0 ELSE 1 END, updated_at=CURRENT_TIMESTAMP WHERE id=?", id); err != nil {
				respondFail(w, 500, err.Error())
				return
			}
		}
	default:
		respondFail(w, 400, "不支持的操作: "+req.Action)
		return
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "操作成功")
}
