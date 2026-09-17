package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetAllExperienceShares returns all experience shares.
func (h *Handler) GetAllExperienceShares(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
		create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version, created_at, updated_at
		FROM experience_shares ORDER BY created_at DESC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.ExperienceShare
	for rows.Next() {
		var e models.ExperienceShare
		if err := rows.Scan(&e.ID, &e.Title, &e.Summary, &e.Content, &e.ContentRenderMode, &e.CoverURL, &e.PdfFileURL, &e.PdfFileName, &e.PdfFileSize,
			&e.CreateType, &e.AuthorID, &e.AuthorName, &e.Status, &e.PdfParseStatus, &e.PdfParseResult, &e.SourceFileName, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, e)
	}
	respondOK(w, list)
}

// GetExperienceShare returns a single experience share by id.
func (h *Handler) GetExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var e models.ExperienceShare
	err := h.db.QueryRow(`SELECT id, title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
		create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version, created_at, updated_at
		FROM experience_shares WHERE id = ?`, id).
		Scan(&e.ID, &e.Title, &e.Summary, &e.Content, &e.ContentRenderMode, &e.CoverURL, &e.PdfFileURL, &e.PdfFileName, &e.PdfFileSize,
			&e.CreateType, &e.AuthorID, &e.AuthorName, &e.Status, &e.PdfParseStatus, &e.PdfParseResult, &e.SourceFileName, &e.Version, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "经验分享不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, e)
}

// CreateExperienceShare creates a new experience share.
func (h *Handler) CreateExperienceShare(w http.ResponseWriter, r *http.Request) {
	var e models.ExperienceShare
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec(`INSERT INTO experience_shares (title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
		create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Title, e.Summary, e.Content, e.ContentRenderMode, e.CoverURL, e.PdfFileURL, e.PdfFileName, e.PdfFileSize,
		e.CreateType, e.AuthorID, e.AuthorName, e.Status, e.PdfParseStatus, e.PdfParseResult, e.SourceFileName, e.Version)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	e.ID = int(id)
	respondOK(w, e)
}

// UpdateExperienceShare updates an existing experience share.
func (h *Handler) UpdateExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var e models.ExperienceShare
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec(`UPDATE experience_shares SET title=?, summary=?, content=?, content_render_mode=?, cover_url=?, pdf_file_url=?, pdf_file_name=?, pdf_file_size=?,
		create_type=?, author_id=?, author_name=?, status=?, pdf_parse_status=?, pdf_parse_result=?, source_file_name=?, version=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		e.Title, e.Summary, e.Content, e.ContentRenderMode, e.CoverURL, e.PdfFileURL, e.PdfFileName, e.PdfFileSize,
		e.CreateType, e.AuthorID, e.AuthorName, e.Status, e.PdfParseStatus, e.PdfParseResult, e.SourceFileName, e.Version, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	e.ID = id
	respondOK(w, e)
}

// DeleteExperienceShare deletes an experience share by id.
func (h *Handler) DeleteExperienceShare(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM experience_shares WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// ImportPdf handles PDF file import for experience shares.
func (h *Handler) ImportPdf(w http.ResponseWriter, r *http.Request) {
	respondFail(w, 501, "PDF 导入功能尚未实现，请使用其他格式")
}

// ExportAll exports all experience shares.
func (h *Handler) ExportAll(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
		create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version, created_at, updated_at
		FROM experience_shares ORDER BY created_at DESC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.ExperienceShare
	for rows.Next() {
		var e models.ExperienceShare
		if err := rows.Scan(&e.ID, &e.Title, &e.Summary, &e.Content, &e.ContentRenderMode, &e.CoverURL, &e.PdfFileURL, &e.PdfFileName, &e.PdfFileSize,
			&e.CreateType, &e.AuthorID, &e.AuthorName, &e.Status, &e.PdfParseStatus, &e.PdfParseResult, &e.SourceFileName, &e.Version, &e.CreatedAt, &e.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, e)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=experience-shares.json")
	json.NewEncoder(w).Encode(models.APIResponse{Success: true, Data: list})
}

// ImportAll imports experience shares from JSON.
func (h *Handler) ImportAll(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Shares []models.ExperienceShare `json:"shares"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	inserted := 0
	for _, e := range req.Shares {
		_, err := tx.Exec(`INSERT INTO experience_shares (title, summary, content, content_render_mode, cover_url, pdf_file_url, pdf_file_name, pdf_file_size,
			create_type, author_id, author_name, status, pdf_parse_status, pdf_parse_result, source_file_name, version)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			e.Title, e.Summary, e.Content, e.ContentRenderMode, e.CoverURL, e.PdfFileURL, e.PdfFileName, e.PdfFileSize,
			e.CreateType, e.AuthorID, e.AuthorName, e.Status, e.PdfParseStatus, e.PdfParseResult, e.SourceFileName, e.Version)
		if err != nil {
			continue
		}
		inserted++
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"insertedCount": inserted})
}
