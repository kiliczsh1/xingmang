// Package fanqie 提供番茄小说榜单数据源（a_bogus 签名 + PUA 字体解码的 Go 原生实现）。
//
// 由 scripts/fanqie_bridge.py + fanqie-rank-mcp（api.py / a_bogus.py / pua_map.py）
// 整体移植而来，运行无需 Python 环境。对外输出 JSON 与 Python 桥的
// /rankings、/ranking 接口逐字段一致，作为 external 桥不可用时的内置回退。
package fanqie

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// RankingsJSON 返回与 Python 桥 GET /rankings 相同的 JSON：{"groups": [...]}。
func RankingsJSON() ([]byte, error) {
	groups, err := fetchLeaderboards()
	if err != nil {
		// 解析失败时用静态兜底（对齐 Python 桥行为）
		groups = fallbackLeaderboards()
	}
	return json.Marshal(map[string]interface{}{"groups": groups})
}

// BooksJSON 返回与 Python 桥 GET /ranking 相同的 JSON：{"books": [...]}。
// 参数对齐 Python 桥的解析逻辑（gender/rankMold 仅在解析失败时用默认值，
// 注意 strconv.Atoi("0") 返回 (0, nil)——gender=0 女频是合法值，不可改写）。
func BooksJSON(gender, rankMold, categoryID string, offset, limit int) ([]byte, error) {
	g := 1
	if n, err := strconv.Atoi(strings.TrimSpace(gender)); err == nil {
		g = n
	}
	rm := 2
	if n, err := strconv.Atoi(strings.TrimSpace(rankMold)); err == nil {
		rm = n
	}
	if categoryID == "" {
		categoryID = "1141"
	}
	if offset < 0 {
		offset = 0
	}
	switch {
	case limit < 1:
		limit = 30
	case limit > 30:
		limit = 30
	}

	books, err := fetchBooks(strconv.Itoa(g), strconv.Itoa(rm), categoryID, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("番茄桥不可用: %w", err)
	}
	return json.Marshal(map[string]interface{}{"books": books})
}
