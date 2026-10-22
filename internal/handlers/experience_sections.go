package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// maxExperienceSections 每个吧的自定义分区上限（不含内置的「精华」与「全部」）
const maxExperienceSections = 15

const experienceSectionSelectCols = `id, COALESCE(board, 'experience') AS board, name, order_num, created_at, updated_at`

func scanExperienceSection(row interface {
	Scan(dest ...interface{}) error
}) (models.ExperienceSection, error) {
	var s models.ExperienceSection
	err := row.Scan(&s.ID, &s.Board, &s.Name, &s.OrderNum, &s.CreatedAt, &s.UpdatedAt)
	s.Board = normalizeBoard(s.Board)
	return s, err
}

// requireBoard 从请求里取并校验 board（query 或 body 用同一套规则）：
// 空 → experience（保持老前端不带参数时的行为）；非法值 → 400。
func requireBoard(w http.ResponseWriter, raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return BoardExperience, true
	case BoardExperience:
		return BoardExperience, true
	case BoardTutorial:
		return BoardTutorial, true
	default:
		respondFail(w, 400, "board 只能是 experience 或 tutorial")
		return "", false
	}
}

// GetExperienceSections 返回某个吧的自定义分区（按 order_num 升序）。
// ?board=tutorial|experience，缺省 experience。
func (h *Handler) GetExperienceSections(w http.ResponseWriter, r *http.Request) {
	board, ok := requireBoard(w, r.URL.Query().Get("board"))
	if !ok {
		return
	}

	rows, err := h.db.Query(`SELECT `+experienceSectionSelectCols+` FROM experience_sections WHERE COALESCE(board, 'experience') = ? ORDER BY order_num ASC, id ASC`, board)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := []models.ExperienceSection{}
	for rows.Next() {
		s, err := scanExperienceSection(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, s)
	}
	respondOK(w, list)
}

// CreateExperienceSection 新建分区：名称非空、同吧内不重名、不超过 15 个。
func (h *Handler) CreateExperienceSection(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string `json:"name"`
		Board string `json:"board"`
	}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	board, ok := requireBoard(w, req.Board)
	if !ok {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		respondFail(w, 400, "分区名称不能为空")
		return
	}
	if runes := []rune(name); len(runes) > 12 {
		respondFail(w, 400, "分区名称最多 12 个字")
		return
	}

	var count int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM experience_sections WHERE COALESCE(board, 'experience') = ?`, board).Scan(&count); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if count >= maxExperienceSections {
		respondFail(w, 400, "每个吧的自定义分区最多 15 个")
		return
	}

	// ⚠ 重名只在同一个吧内算：两个吧都建「入门」必须都成功
	var dup int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM experience_sections WHERE COALESCE(board, 'experience') = ? AND name = ?`, board, name).Scan(&dup); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if dup > 0 {
		respondFail(w, 400, "该吧已存在同名分区")
		return
	}

	var nextOrder int
	if err := h.db.QueryRow(`SELECT COALESCE(MAX(order_num), 0) + 1 FROM experience_sections WHERE COALESCE(board, 'experience') = ?`, board).Scan(&nextOrder); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	res, err := h.db.Exec(`INSERT INTO experience_sections (board, name, order_num) VALUES (?, ?, ?)`, board, name, nextOrder)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := res.LastInsertId()

	s, err := scanExperienceSection(h.db.QueryRow(`SELECT `+experienceSectionSelectCols+` FROM experience_sections WHERE id = ?`, id))
	if err != nil {
		respondOK(w, map[string]interface{}{"id": id, "board": board, "name": name, "order_num": nextOrder})
		return
	}
	respondOK(w, s)
}

// UpdateExperienceSection 重命名 / 改排序（按出现的键更新）。
func (h *Handler) UpdateExperienceSection(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var raw map[string]json.RawMessage
	if !decodeJSONBody(w, r, &raw) {
		return
	}

	sets := []string{}
	args := []interface{}{}

	if v, ok := raw["name"]; ok {
		name := strings.TrimSpace(rawString(v))
		if name == "" {
			respondFail(w, 400, "分区名称不能为空")
			return
		}
		if runes := []rune(name); len(runes) > 12 {
			respondFail(w, 400, "分区名称最多 12 个字")
			return
		}
		// 重名只在同一个吧内算（分区 id 全局唯一，但两个吧允许同名分区）
		var sectionBoard string
		_ = h.db.QueryRow(`SELECT COALESCE(board, 'experience') FROM experience_sections WHERE id = ?`, id).Scan(&sectionBoard)
		var dup int
		if err := h.db.QueryRow(`SELECT COUNT(*) FROM experience_sections WHERE COALESCE(board, 'experience') = ? AND name = ? AND id <> ?`,
			normalizeBoard(sectionBoard), name, id).Scan(&dup); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		if dup > 0 {
			respondFail(w, 400, "该吧已存在同名分区")
			return
		}
		sets = append(sets, "name=?")
		args = append(args, name)
	}
	if v, ok := raw["order_num"]; ok {
		sets = append(sets, "order_num=?")
		args = append(args, rawInt(v))
	}

	if len(sets) == 0 {
		respondFail(w, 400, "没有需要更新的字段")
		return
	}
	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	args = append(args, id)

	if _, err := h.db.Exec("UPDATE experience_sections SET "+strings.Join(sets, ", ")+" WHERE id=?", args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	s, err := scanExperienceSection(h.db.QueryRow(`SELECT `+experienceSectionSelectCols+` FROM experience_sections WHERE id = ?`, id))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "分区不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, s)
}

// ReorderExperienceSections 整体重排（前端拖拽/上移下移后把最终顺序整个提交）。
func (h *Handler) ReorderExperienceSections(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []int `json:"ids"`
	}
	if !decodeJSONBody(w, r, &req) {
		return
	}
	if len(req.IDs) == 0 {
		respondFail(w, 400, "排序内容为空")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	for i, id := range req.IDs {
		if _, err := tx.Exec(`UPDATE experience_sections SET order_num=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, i+1, id); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
	}
	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]interface{}{"affected": len(req.IDs)})
}

// DeleteExperienceSection 删除分区：该分区下的帖子不会被删，只是回落到「未归类」。
func (h *Handler) DeleteExperienceSection(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	// 没有外键可依赖（section_id 是 ALTER 补的列），必须手动解绑，否则帖子会指向一个已不存在的分区 id
	if _, err := tx.Exec(`UPDATE experience_shares SET section_id=NULL WHERE section_id=?`, id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	res, err := tx.Exec(`DELETE FROM experience_sections WHERE id=?`, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		tx.Rollback()
		respondFail(w, 404, "分区不存在")
		return
	}
	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	// 文件帖没有外键可依赖，同样要手动解绑（否则指向已删除的分区 id）
	if h.shareStore != nil {
		h.shareStore.ClearSection(id)
	}
	respondOK(w, "删除成功")
}
