package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetCharactersByBook returns all characters for a given book.
func (h *Handler) GetCharactersByBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT id, book_id, name, gender, personality, info, folder, folders, created_at, updated_at FROM characters WHERE book_id = ? ORDER BY created_at ASC", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Character
	for rows.Next() {
		var c models.Character
		if err := rows.Scan(&c.ID, &c.BookID, &c.Name, &c.Gender, &c.Personality, &c.Info, &c.Folder, &c.Folders, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, c)
	}
	respondOK(w, list)
}

// GetCharacter returns a single character by id.
func (h *Handler) GetCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var c models.Character
	err := h.db.QueryRow("SELECT id, book_id, name, gender, personality, info, folder, folders, created_at, updated_at FROM characters WHERE id = ?", id).
		Scan(&c.ID, &c.BookID, &c.Name, &c.Gender, &c.Personality, &c.Info, &c.Folder, &c.Folders, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "角色不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, c)
}

// CreateCharacter creates a new character.
func (h *Handler) CreateCharacter(w http.ResponseWriter, r *http.Request) {
	var c models.Character
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO characters (book_id, name, gender, personality, info, folder, folders) VALUES (?, ?, ?, ?, ?, ?, ?)",
		c.BookID, c.Name, c.Gender, c.Personality, c.Info, c.Folder, c.Folders)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	c.ID = int(id)
	respondOK(w, c)
}

// UpdateCharacter updates an existing character.
func (h *Handler) UpdateCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var c models.Character
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE characters SET book_id=?, name=?, gender=?, personality=?, info=?, folder=?, folders=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		c.BookID, c.Name, c.Gender, c.Personality, c.Info, c.Folder, c.Folders, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	c.ID = id
	respondOK(w, c)
}

// DeleteCharacter deletes a character by id.
func (h *Handler) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM characters WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// BatchDeleteCharactersRequest is the request body for batch deleting characters.
type BatchDeleteCharactersRequest struct {
	IDs []int `json:"ids"`
}

// BatchDeleteCharacters deletes multiple characters by ids.
func (h *Handler) BatchDeleteCharacters(w http.ResponseWriter, r *http.Request) {
	var req BatchDeleteCharactersRequest
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

	for _, id := range req.IDs {
		if _, err := tx.Exec("DELETE FROM characters WHERE id = ?", id); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// BatchMoveCharactersRequest is the request body for batch moving characters.
type BatchMoveCharactersRequest struct {
	IDs    []int  `json:"ids"`
	Folder string `json:"folder"`
}

// BatchMoveCharacters moves multiple characters to a folder.
func (h *Handler) BatchMoveCharacters(w http.ResponseWriter, r *http.Request) {
	var req BatchMoveCharactersRequest
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

	for _, id := range req.IDs {
		if _, err := tx.Exec("UPDATE characters SET folder=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", req.Folder, id); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "移动成功")
}

// GetFolders returns distinct folders for a book.
func (h *Handler) GetFolders(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	rows, err := h.db.Query("SELECT DISTINCT folder FROM characters WHERE book_id = ?", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var folders []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		folders = append(folders, f)
	}
	respondOK(w, folders)
}
