package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func (h *Handler) GetPrompts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, name, description, content, category, order_num, fields, subcategories, card_type, password, creator_name, version, metadata, created_at, updated_at FROM prompts ORDER BY order_num ASC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		var p models.Prompt
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Content, &p.Category, &p.OrderNum, &p.Fields, &p.Subcategories, &p.CardType, &p.Password, &p.CreatorName, &p.Version, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if p.CardType == "" {
			p.CardType = "normal"
		}
		prompts = append(prompts, p)
	}
	respondOK(w, prompts)
}

func (h *Handler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var p models.Prompt
	err := h.db.QueryRow("SELECT id, name, description, content, category, order_num, fields, subcategories, card_type, password, creator_name, version, metadata, created_at, updated_at FROM prompts WHERE id = ?", id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Content, &p.Category, &p.OrderNum, &p.Fields, &p.Subcategories, &p.CardType, &p.Password, &p.CreatorName, &p.Version, &p.Metadata, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "提示词不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if p.CardType == "" {
		p.CardType = "normal"
	}
	respondOK(w, p)
}

func (h *Handler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	var p models.Prompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	result, err := h.db.Exec("INSERT INTO prompts (name, description, content, category, order_num, fields, subcategories, card_type, password, creator_name, version, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Description, p.Content, p.Category, p.OrderNum, p.Fields, p.Subcategories, p.CardType, p.Password, p.CreatorName, p.Version, p.Metadata)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := result.LastInsertId()
	p.ID = int(id)
	respondOK(w, p)
}

func (h *Handler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var p models.Prompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	_, err := h.db.Exec("UPDATE prompts SET name=?, description=?, content=?, category=?, order_num=?, fields=?, subcategories=?, card_type=?, password=?, creator_name=?, version=?, metadata=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		p.Name, p.Description, p.Content, p.Category, p.OrderNum, p.Fields, p.Subcategories, p.CardType, p.Password, p.CreatorName, p.Version, p.Metadata, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	p.ID = id
	respondOK(w, p)
}

func (h *Handler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	_, err := h.db.Exec("DELETE FROM prompts WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
