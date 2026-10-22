package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetEntries returns entries with optional filters and pagination.
func (h *Handler) GetEntries(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("book_id")
	categoryType := r.URL.Query().Get("category_type")
	keyword := r.URL.Query().Get("keyword")
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `SELECT id, book_id, category_id, category_type, name, avatar, tags, appearance, background,
		personality, relationships, custom_fields, description, is_public, creator_name, creator_avatar,
		use_count, like_count, created_at, updated_at
		FROM entries WHERE 1=1`
	countQuery := "SELECT COUNT(*) FROM entries WHERE 1=1"
	args := []interface{}{}

	if bookID != "" {
		query += " AND book_id = ?"
		countQuery += " AND book_id = ?"
		args = append(args, bookID)
	}
	if categoryType != "" {
		query += " AND category_type = ?"
		countQuery += " AND category_type = ?"
		args = append(args, categoryType)
	}
	if keyword != "" {
		query += " AND (name LIKE ? OR description LIKE ?)"
		countQuery += " AND (name LIKE ? OR description LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// Get total count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	err := h.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	query += " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Entry
	for rows.Next() {
		var e models.Entry
		if err := rows.Scan(&e.ID, &e.BookID, &e.CategoryID, &e.CategoryType, &e.Name, &e.Avatar, &e.Tags, &e.Appearance, &e.Background,
			&e.Personality, &e.Relationships, &e.CustomFields, &e.Description, &e.IsPublic, &e.CreatorName, &e.CreatorAvatar,
			&e.UseCount, &e.LikeCount, &e.CreatedAt, &e.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, e)
	}

	type PaginatedResult struct {
		Items    []models.Entry `json:"items"`
		Total    int            `json:"total"`
		Page     int            `json:"page"`
		PageSize int            `json:"page_size"`
	}

	respondOK(w, PaginatedResult{Items: list, Total: total, Page: page, PageSize: pageSize})
}

// GetEntry returns a single entry by id.
func (h *Handler) GetEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var e models.Entry
	err := h.db.QueryRow(`SELECT id, book_id, category_id, category_type, name, avatar, tags, appearance, background,
		personality, relationships, custom_fields, description, is_public, creator_name, creator_avatar,
		use_count, like_count, created_at, updated_at
		FROM entries WHERE id = ?`, id).
		Scan(&e.ID, &e.BookID, &e.CategoryID, &e.CategoryType, &e.Name, &e.Avatar, &e.Tags, &e.Appearance, &e.Background,
			&e.Personality, &e.Relationships, &e.CustomFields, &e.Description, &e.IsPublic, &e.CreatorName, &e.CreatorAvatar,
			&e.UseCount, &e.LikeCount, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "条目不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, e)
}

// CreateEntry creates a new entry.
func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	var e models.Entry
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec(`INSERT INTO entries (book_id, category_id, category_type, name, avatar, tags, appearance, background,
		personality, relationships, custom_fields, description, is_public, creator_name, creator_avatar, use_count, like_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.BookID, e.CategoryID, e.CategoryType, e.Name, e.Avatar, e.Tags, e.Appearance, e.Background,
		e.Personality, e.Relationships, e.CustomFields, e.Description, e.IsPublic, e.CreatorName, e.CreatorAvatar, e.UseCount, e.LikeCount)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	e.ID = int(id)
	if rel, err := syncEntryFile(h, &e, ""); err != nil {
		h.db.Exec("DELETE FROM entries WHERE id=?", id)
		respondFail(w, 500, "写入条目文件失败: "+err.Error())
		return
	} else {
		h.db.Exec("UPDATE entries SET file_path=? WHERE id=?", rel, id)
	}
	respondOK(w, e)
}

// UpdateEntry updates an existing entry.
func (h *Handler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var e models.Entry
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec(`UPDATE entries SET book_id=?, category_id=?, category_type=?, name=?, avatar=?, tags=?, appearance=?, background=?,
		personality=?, relationships=?, custom_fields=?, description=?, is_public=?, creator_name=?, creator_avatar=?, use_count=?, like_count=?,
		updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		e.BookID, e.CategoryID, e.CategoryType, e.Name, e.Avatar, e.Tags, e.Appearance, e.Background,
		e.Personality, e.Relationships, e.CustomFields, e.Description, e.IsPublic, e.CreatorName, e.CreatorAvatar, e.UseCount, e.LikeCount, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	e.ID = id
	var oldRel string
	var oldBookID sql.NullInt64
	h.db.QueryRow("SELECT COALESCE(file_path,''), book_id FROM entries WHERE id=?", id).Scan(&oldRel, &oldBookID)
	if e.BookID == nil && oldBookID.Valid {
		v := int(oldBookID.Int64)
		e.BookID = &v
	}
	rel, err := syncEntryFile(h, &e, oldRel)
	if err != nil {
		respondFail(w, 500, "写入条目文件失败: "+err.Error())
		return
	}
	h.db.Exec("UPDATE entries SET file_path=? WHERE id=?", rel, id)
	respondOK(w, e)
}

// DeleteEntry deletes an entry by id.
func (h *Handler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var bookID sql.NullInt64
	var rel string
	h.db.QueryRow("SELECT book_id, COALESCE(file_path,'') FROM entries WHERE id=?", id).Scan(&bookID, &rel)
	bid := 0
	if bookID.Valid {
		bid = int(bookID.Int64)
	}
	if err := bookstore.NewService(h.db).TrashEntity(bid, "entry", rel); err != nil {
		respondFail(w, 500, "移入回收站失败: "+err.Error())
		return
	}
	if _, err := h.db.Exec("DELETE FROM entries WHERE id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// AddToBookRequest is the request body for associating an entry with a book.
type AddToBookRequest struct {
	EntryID int `json:"entry_id"`
	BookID  int `json:"book_id"`
}

// AddToBook associates an entry with a book.
// syncEntryFile 把条目写进 world/<分类>/<名称>.md。
func syncEntryFile(h *Handler, e *models.Entry, oldRel string) (string, error) {
	cat := strings.TrimSpace(e.CategoryType)
	if cat == "" && e.CategoryID > 0 {
		h.db.QueryRow("SELECT name FROM entry_categories WHERE id=?", e.CategoryID).Scan(&cat)
	}
	return bookstore.NewService(h.db).SyncEntry(e, cat, oldRel)
}

func (h *Handler) AddToBook(w http.ResponseWriter, r *http.Request) {
	var req AddToBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE entries SET book_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.BookID, req.EntryID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "关联成功")
}

// GetCategories returns entry categories with optional book_id filter.
func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("book_id")

	var rows *sql.Rows
	var err error

	if bookID != "" {
		rows, err = h.db.Query(`SELECT DISTINCT c.id, c.name, c.type, c.icon, c.color, c.sort_order, c.created_at, c.updated_at
			FROM entry_categories c
			JOIN entries e ON e.category_id = c.id
			WHERE e.book_id = ?
			ORDER BY c.sort_order ASC`, bookID)
	} else {
		rows, err = h.db.Query("SELECT id, name, type, icon, color, sort_order, created_at, updated_at FROM entry_categories ORDER BY sort_order ASC")
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.EntryCategory
	for rows.Next() {
		var c models.EntryCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &c.Icon, &c.Color, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, c)
	}
	respondOK(w, list)
}

// AIAutoComplete is a placeholder for AI auto-complete functionality.
func (h *Handler) AIAutoComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
		Context string `json:"context"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	respondFail(w, 501, "AI 自动补全功能尚未实现")
}
