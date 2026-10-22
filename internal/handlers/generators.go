package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// GetGenerators returns all generators.
func (h *Handler) GetGenerators(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, name, description, icon, core_prompt, remark, order_num, created_at, updated_at FROM generators ORDER BY order_num ASC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var list []models.Generator
	for rows.Next() {
		var g models.Generator
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.Icon, &g.CorePrompt, &g.Remark, &g.OrderNum, &g.CreatedAt, &g.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, g)
	}
	respondOK(w, list)
}

// GetGenerator returns a single generator by id.
func (h *Handler) GetGenerator(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var g models.Generator
	err := h.db.QueryRow("SELECT id, name, description, icon, core_prompt, remark, order_num, created_at, updated_at FROM generators WHERE id = ?", id).
		Scan(&g.ID, &g.Name, &g.Description, &g.Icon, &g.CorePrompt, &g.Remark, &g.OrderNum, &g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "生成器不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, g)
}

// CreateGenerator creates a new generator.
func (h *Handler) CreateGenerator(w http.ResponseWriter, r *http.Request) {
	var g models.Generator
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	result, err := h.db.Exec("INSERT INTO generators (name, description, icon, core_prompt, remark, order_num) VALUES (?, ?, ?, ?, ?, ?)",
		g.Name, g.Description, g.Icon, g.CorePrompt, g.Remark, g.OrderNum)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	id, _ := result.LastInsertId()
	g.ID = int(id)
	respondOK(w, g)
}

// UpdateGenerator updates an existing generator.
func (h *Handler) UpdateGenerator(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var g models.Generator
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	_, err := h.db.Exec("UPDATE generators SET name=?, description=?, icon=?, core_prompt=?, remark=?, order_num=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		g.Name, g.Description, g.Icon, g.CorePrompt, g.Remark, g.OrderNum, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	g.ID = id
	respondOK(w, g)
}

// DeleteGenerator deletes a generator by id.
func (h *Handler) DeleteGenerator(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	_, err := h.db.Exec("DELETE FROM generators WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
