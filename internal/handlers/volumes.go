package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetVolumesByBook returns all volumes for a given book.
func (h *Handler) GetVolumesByBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT id, book_id, parent_id, title, order_num, created_at, updated_at FROM volumes WHERE book_id = ? ORDER BY order_num ASC", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Volume
	for rows.Next() {
		var v models.Volume
		if err := rows.Scan(&v.ID, &v.BookID, &v.ParentID, &v.Title, &v.OrderNum, &v.CreatedAt, &v.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, v)
	}
	respondOK(w, list)
}

// GetVolume returns a single volume by id.
func (h *Handler) GetVolume(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var v models.Volume
	err := h.db.QueryRow("SELECT id, book_id, parent_id, title, order_num, created_at, updated_at FROM volumes WHERE id = ?", id).
		Scan(&v.ID, &v.BookID, &v.ParentID, &v.Title, &v.OrderNum, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "卷不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, v)
}

// CreateVolume creates a new volume.
func (h *Handler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var v models.Volume
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO volumes (book_id, parent_id, title, order_num) VALUES (?, ?, ?, ?)",
		v.BookID, v.ParentID, v.Title, v.OrderNum)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	v.ID = int(id)
	respondOK(w, v)
}

// UpdateVolume updates an existing volume.
func (h *Handler) UpdateVolume(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var v models.Volume
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE volumes SET book_id=?, parent_id=?, title=?, order_num=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		v.BookID, v.ParentID, v.Title, v.OrderNum, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	v.ID = id
	respondOK(w, v)
}

// DeleteVolume deletes a volume by id.
func (h *Handler) DeleteVolume(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM volumes WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
