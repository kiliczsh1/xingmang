package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// GetBookDocs 列出书内文档（outline / idea）。
// GET /api/books/{id}/docs?kind=outline
func (h *Handler) GetBookDocs(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "outline"
	}
	list, err := bookstore.NewService(h.db).ListDocs(id, kind)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, list)
}

// CreateBookDoc 新建大纲/灵感文档。
// POST /api/books/{id}/docs  { kind, title, content }
func (h *Handler) CreateBookDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var req struct {
		Kind    string `json:"kind"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	doc, err := bookstore.NewService(h.db).CreateDoc(id, req.Kind, req.Title, req.Content)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, doc)
}

// UpdateBookDoc 改写文档（正文/标题）。
// PUT /api/docs/{id}
func (h *Handler) UpdateBookDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	doc, err := bookstore.NewService(h.db).UpdateDoc(id, req.Title, req.Content)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, doc)
}

// DeleteBookDoc 删除文档（文件进回收站）。
// DELETE /api/docs/{id}
func (h *Handler) DeleteBookDoc(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	if err := bookstore.NewService(h.db).DeleteDoc(id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
