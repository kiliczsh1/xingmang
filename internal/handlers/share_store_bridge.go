package handlers

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 经验分享的「数据库 + 文件」双读桥接。
//
// 老帖仍在 experience_shares 表里；新建的帖子一律写到 data/experience-shares/。
// 判断归属的规则：**文件存在且库里没有同 id 记录** 才当文件帖 ——
// 万一将来把文件帖导回数据库，以库为准，避免同一篇出现两份。

// 两个吧的取值。除 tutorial 之外的任何输入一律当 experience（见 normalizeBoard）。
const (
	BoardExperience = models.BoardExperience
	BoardTutorial   = models.BoardTutorial
)

// normalizeBoard 委托给 models.NormalizeBoard（单一真源，sharestore 也用它）。
//
// ⚠ 唯一需要兜底的场景是**用户从旧版本导出的文件**（当时还没有 board 字段，读出来是空串），
// 以及手改过的帖子 JSON。本机存量数据在迁移时已经刷平，不靠这里。
func normalizeBoard(b string) string {
	return models.NormalizeBoard(b)
}

// parseBoardParam 解析列表接口的 ?board= 查询参数。
// 缺省（不带参数）→ experience，保持老前端不带参数时的行为；非法值直接 400，
// 免得前端拼错参数却看到「另一份数据」还以为数据丢了。
func parseBoardParam(r *http.Request) (string, bool) {
	raw := strings.TrimSpace(r.URL.Query().Get("board"))
	switch strings.ToLower(raw) {
	case "":
		return BoardExperience, true
	case BoardExperience:
		return BoardExperience, true
	case BoardTutorial:
		return BoardTutorial, true
	default:
		return "", false
	}
}

// filterByBoard 按吧过滤（board 为空表示不过滤，导出备份用）。
func filterByBoard(list []models.ExperienceShare, board string) []models.ExperienceShare {
	if board == "" {
		return list
	}
	out := make([]models.ExperienceShare, 0, len(list))
	for _, e := range list {
		if normalizeBoard(e.Board) == board {
			out = append(out, e)
		}
	}
	return out
}

// shareInStore 该 id 是否为文件帖。
func (h *Handler) shareInStore(id int) bool {
	if h.shareStore == nil || !h.shareStore.Has(id) {
		return false
	}
	var exists int
	if err := h.db.QueryRow(`SELECT 1 FROM experience_shares WHERE id = ?`, id).Scan(&exists); err == nil {
		return false // 库里也有 → 以库为准
	}
	return true
}

// storeShareByID 读文件帖。
func (h *Handler) storeShareByID(id int) (models.ExperienceShare, bool) {
	if h.shareStore == nil {
		return models.ExperienceShare{}, false
	}
	e, ok := h.shareStore.Get(id)
	if !ok {
		return e, false
	}
	e.Board = normalizeBoard(e.Board)
	return e, true
}

// mergeShares 合并库内帖与文件帖，置顶优先、其次按创建时间倒序（列表页口径）。
// ⚠ 两端 created_at 都是「2006-01-02 15:04:05」UTC 字符串，可直接比较；
// 同秒时用 id 兜底，保证顺序稳定。
func mergeShares(dbList, fileList []models.ExperienceShare) []models.ExperienceShare {
	merged := make([]models.ExperienceShare, 0, len(dbList)+len(fileList))
	merged = append(merged, dbList...)
	merged = append(merged, fileList...)
	sort.SliceStable(merged, func(i, j int) bool {
		// 置顶帖永远在前（两边的置顶各自参与排序，不区分来源）
		pi, pj := merged[i].IsPinned == 1, merged[j].IsPinned == 1
		if pi != pj {
			return pi
		}
		if merged[i].CreatedAt != merged[j].CreatedAt {
			return merged[i].CreatedAt > merged[j].CreatedAt
		}
		return merged[i].ID > merged[j].ID
	})
	return merged
}

// dbShareList 读库内帖（created_at 倒序）。board 为空表示不过滤。
func (h *Handler) dbShareList(board string) ([]models.ExperienceShare, error) {
	query := `SELECT ` + experienceShareSelectCols + ` FROM experience_shares`
	var args []interface{}
	if board != "" {
		query += ` WHERE board = ?`
		args = append(args, board)
	}
	query += ` ORDER BY COALESCE(is_pinned, 0) DESC, created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ExperienceShare
	for rows.Next() {
		e, err := scanExperienceShare(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

// fileShareList 读全部文件帖（board 为空表示不过滤）。
func (h *Handler) fileShareList(board string) []models.ExperienceShare {
	if h.shareStore == nil {
		return nil
	}
	list := h.shareStore.List()
	for i := range list {
		list[i].Board = normalizeBoard(list[i].Board)
	}
	return filterByBoard(list, board)
}

// shareListByBoard 列表 / 导出统一入口：库与文件两边**各自按 board 过滤后再 merge**。
// ⚠ 先合并再过滤等于多扫一遍全量，而且一旦某边漏了 board 字段就会混吧。
func (h *Handler) shareListByBoard(board string) ([]models.ExperienceShare, error) {
	dbList, err := h.dbShareList(board)
	if err != nil {
		return nil, err
	}
	return mergeShares(dbList, h.fileShareList(board)), nil
}

// patchWithCover 处理封面：上传型封面（cover_image）先落盘换成站内地址，
// remove_cover=true 则置空并删掉旧图；结果合并进要写库/写文件的字段集合。
// 返回的 map 可直接给数据库 SET 或文件存储的 Update 用。
func (h *Handler) resolveCoverPatch(raw map[string]json.RawMessage, oldCover *string) (map[string]json.RawMessage, error) {
	patch := make(map[string]json.RawMessage, len(raw)+1)
	for k, v := range raw {
		patch[k] = v
	}

	cover, err := saveCoverFromRaw(raw)
	if err != nil {
		return nil, err
	}
	if cover != "" {
		if oldCover != nil {
			deleteExperienceCoverByURL(*oldCover)
		}
		encoded, _ := json.Marshal(cover)
		patch["cover_url"] = encoded
		return patch, nil
	}
	if v, ok := raw["remove_cover"]; ok && rawBool(v) {
		if oldCover != nil {
			deleteExperienceCoverByURL(*oldCover)
		}
		patch["cover_url"] = json.RawMessage("null")
	}
	return patch, nil
}
