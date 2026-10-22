package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// GetBookCover 直接返回封面文件（不再走 base64 内联，列表页省流量）。
// GET /api/books/{id}/cover
func (h *Handler) GetBookCover(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	path, mime, err := bookstore.NewService(h.db).CoverFile(id)
	if err != nil {
		// 404 也要显式禁缓存：否则浏览器会把"没有封面"缓存住，保存后仍显示占位
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "no-cache")
	// ServeContent 会带上 Last-Modified 与 Content-Length，浏览器可条件请求
	http.ServeContent(w, r, filepath.Base(path), st.ModTime(), f)
}

// GetBookGallery 返回封面画廊列表。
// GET /api/books/{id}/gallery
func (h *Handler) GetBookGallery(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	items, err := bookstore.NewService(h.db).ListGallery(id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	// 前端只需要文件名与时间，URL 由前端拼 /api/books/{id}/gallery/{name}
	respondOK(w, items)
}

// ActivateBookCover 把画廊里的一张设为当前封面。
// POST /api/books/{id}/gallery/activate  { "name": "20261018-203012-a1b2c3.png" }
func (h *Handler) ActivateBookCover(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		respondFail(w, 400, "缺少画廊文件名")
		return
	}
	if err := bookstore.NewService(h.db).ActivateGallery(id, req.Name); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "已设为封面")
}

// GetGalleryImage 返回画廊里的单张图片。
// GET /api/books/{id}/gallery/{name}
func (h *Handler) GetGalleryImage(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	name := r.PathValue("name")
	if name == "" || name != filepath.Base(name) {
		respondFail(w, 400, "非法的文件名")
		return
	}
	dir, err := bookstore.NewService(h.db).BookDir(id)
	if err != nil {
		respondFail(w, 404, err.Error())
		return
	}
	path := filepath.Join(dir, "gallery", name)
	f, err := os.Open(path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeContent(w, r, name, st.ModTime(), f)
}

// DeleteGalleryImage 删除画廊里的一张（进回收站）。
// DELETE /api/books/{id}/gallery/{name}
func (h *Handler) DeleteGalleryImage(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	name := r.PathValue("name")
	if name == "" || name != filepath.Base(name) {
		respondFail(w, 400, "非法的文件名")
		return
	}
	if err := bookstore.NewService(h.db).DeleteGallery(id, name); err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, "已移入回收站")
}

// RevealBook 在系统文件管理器里定位书目录。
// POST /api/books/{id}/reveal
func (h *Handler) RevealBook(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	dir, err := bookstore.NewService(h.db).BookDir(id)
	if err != nil {
		respondFail(w, 404, err.Error())
		return
	}
	if _, err := os.Stat(dir); err != nil {
		respondFail(w, 404, "作品目录不存在")
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", "/select,"+dir)
	case "darwin":
		cmd = exec.Command("open", "-R", dir)
	default:
		cmd = exec.Command("xdg-open", dir)
	}
	// 打开失败不影响业务：返回 success 但把错误信息带回去
	if err := cmd.Start(); err != nil {
		respondFail(w, 500, fmt.Sprintf("打开文件管理器失败: %v", err))
		return
	}
	respondOK(w, "已打开")
}

// RescanBook 手动触发目录对账（外部改过文件后点一下）。
// POST /api/books/{id}/rescan
func (h *Handler) RescanBook(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	res, err := bookstore.NewService(h.db).ScanBook(id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, res)
}

// ImportFolder 把已有目录导入为作品。
// POST /api/books/import-folder  { "path": "D:/vault/我的小说" }
func (h *Handler) ImportFolder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		respondFail(w, 400, "请提供目录路径")
		return
	}
	book, err := bookstore.NewService(h.db).ImportFolder(req.Path)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, book)
}

// ExportBook 把作品目录打包成 zip 下载。
// GET /api/books/{id}/export
func (h *Handler) ExportBook(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var title string
	if err := h.db.QueryRow(`SELECT title FROM books WHERE id=?`, id).Scan(&title); err != nil {
		respondFail(w, 404, "作品不存在")
		return
	}
	name := title + ".zip"
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"book.zip\"; filename*=UTF-8''"+url.PathEscape(name))
	if err := bookstore.NewService(h.db).ExportZip(id, w); err != nil {
		// 头已经发出去了，只能记日志；前端会收到一个不完整的 zip
		log.Printf("[books] 导出失败 book=%d: %v", id, err)
	}
}

// ImportBookJSON 导入旧版本导出的作品 JSON（兼容多种结构）。
// POST /api/books/import-json?title=文件名
func (h *Handler) ImportBookJSON(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, 32<<20))
	if err != nil || len(raw) == 0 {
		respondFail(w, 400, "请提供 JSON 文件内容")
		return
	}
	book, chapters, err := bookstore.NewService(h.db).ImportBookJSON(raw, r.URL.Query().Get("title"))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"book": book, "chapters": chapters})
}

// AddGalleryImage 往画廊里加一张图（不改变当前封面）。
// POST /api/books/{id}/gallery  { "image": "data:image/png;base64,..." }
func (h *Handler) AddGalleryImage(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var req struct {
		Image string `json:"image"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Image == "" {
		respondFail(w, 400, "缺少图片数据")
		return
	}
	name, err := bookstore.NewService(h.db).AddGallery(id, req.Image)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]string{"name": name})
}
