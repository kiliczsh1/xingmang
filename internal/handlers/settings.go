package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// GetBooksRoot 返回当前作品库目录。
// GET /api/settings/books-root
func (h *Handler) GetBooksRoot(w http.ResponseWriter, r *http.Request) {
	root, err := bookstore.NewService(h.db).Root()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]string{"path": root})
}

// SetBooksRoot 切换作品库目录，并把现有作品一起搬过去。
// PUT /api/settings/books-root  { "path": "D:/作品库" }
func (h *Handler) SetBooksRoot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		respondFail(w, 400, "请提供目录路径")
		return
	}
	svc := bookstore.NewService(h.db)
	if err := svc.SwitchRoot(req.Path); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	root, _ := svc.Root()
	respondOK(w, map[string]string{"path": root})
}

// PreviewBooksRoot 扫描目标目录，返回"里面有哪些作品、各自是否已在库中"。
//
// 只读，不写库、不写盘。前端据此让用户选「接管 / 搬移 / 取消」。
// POST /api/settings/books-root/preview  { "path": "D:/作品库" }
func (h *Handler) PreviewBooksRoot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		respondFail(w, 400, "请提供目录路径")
		return
	}
	rep, err := bookstore.NewService(h.db).PreviewRoot(req.Path)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, rep)
}

// LinkBooksRoot 接管目标目录：登记已有作品，**不搬移、不复制、不删除任何文件**。
//
// 这是"下一次更新直接链接目录、不需要备份导入"的落点。
// PUT /api/settings/books-root?mode=link  { "path": "D:/作品库" }
func (h *Handler) LinkBooksRoot(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Path) == "" {
		respondFail(w, 400, "请提供目录路径")
		return
	}
	rep, err := bookstore.NewService(h.db).LinkRoot(req.Path)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, rep)
}

// RebuildBookIndex 对当前作品库做一次链接重建（不换根目录）。
//
// 用途：手工把目录拷进作品库之后，让它们出现在书架里。
// POST /api/books/rebuild-index
func (h *Handler) RebuildBookIndex(w http.ResponseWriter, r *http.Request) {
	svc := bookstore.NewService(h.db)
	rep := &bookstore.LinkReport{}
	if err := svc.LinkAll(rep); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, rep)
}

// PickFolder 弹出系统文件夹选择框（本机应用，浏览器拿不到真实路径）。
// POST /api/settings/pick-folder
func (h *Handler) PickFolder(w http.ResponseWriter, r *http.Request) {
	path, err := bookstore.PickFolder("选择作品库目录")
	if errors.Is(err, bookstore.ErrPickCancelled) {
		respondOK(w, map[string]interface{}{"cancelled": true, "path": ""})
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"cancelled": false, "path": path})
}
