package handlers

import (
	"net/http"
	"os"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/rank"
)

// rankBridge 复用同一个榜单数据源客户端（含内存缓存）。
// 默认使用 Go 原生内置实现；显式设置 FANQIE_BRIDGE_URL（外部 Python 桥地址）
// 时才走外部桥（不可用时自动回退内置）。
var rankBridge = rank.NewBridgeClient(os.Getenv("FANQIE_BRIDGE_URL"))

// GetRankLists 返回番茄榜单分组（tab + 题材）。
// 分组结构由桥返回，如 [{title:"男频阅读榜", items:[{id:"1_2_1141", name:"西方奇幻"}]}]。
func (h *Handler) GetRankLists(w http.ResponseWriter, r *http.Request) {
	groups, err := rankBridge.Rankings()
	if err != nil {
		respondFail(w, http.StatusBadGateway, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"groups": groups})
}

// GetRankBooks 返回指定榜单的书。tab 参数格式：gender_rankMold_categoryId（如 "1_2_1141"）。
// 桥不可用时返回 502，前端如实显示错误态（无演示数据回退）。
func (h *Handler) GetRankBooks(w http.ResponseWriter, r *http.Request) {
	tab := r.URL.Query().Get("tab")
	if tab == "" {
		tab = "1_2_1141"
	}
	gender, rankMold, categoryID := "1", "2", "1141"
	if parts := strings.Split(tab, "_"); len(parts) >= 3 {
		gender, rankMold, categoryID = parts[0], parts[1], parts[2]
	}
	page := queryInt(r, "page", 0)
	size := queryInt(r, "size", 30)
	if size > 30 {
		size = 30
	}
	offset := page * size

	books, err := rankBridge.Books(gender, rankMold, categoryID, offset, size)
	if err != nil {
		respondFail(w, http.StatusBadGateway, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"books": books})
}
