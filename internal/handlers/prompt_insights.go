package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

const insightColumns = "id, prompt_id, content, rating, tags, created_at, updated_at"

func scanInsight(row interface {
	Scan(dest ...interface{}) error
}) (models.PromptInsight, error) {
	var i models.PromptInsight
	err := row.Scan(&i.ID, &i.PromptID, &i.Content, &i.Rating, &i.Tags, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

// normalizeTags 把请求里的 tags 规整成 JSON 数组字符串。
// 兼容三种写法：数组（原样）、"长篇,有效" / "#长篇 #有效" 这类分隔字符串、空值。
func normalizeTags(raw json.RawMessage) string {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return "[]"
	}
	if strings.HasPrefix(trimmed, "[") {
		var arr []string
		if json.Unmarshal(raw, &arr) == nil {
			if arr == nil {
				arr = []string{}
			}
			out, err := json.Marshal(arr)
			if err == nil {
				return string(out)
			}
		}
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		parts := strings.FieldsFunc(s, func(r rune) bool {
			return r == ',' || r == '，' || r == '#' || r == ' ' || r == '\t' || r == '/' || r == '、'
		})
		cleaned := make([]string, 0, len(parts))
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				cleaned = append(cleaned, p)
			}
		}
		out, err := json.Marshal(cleaned)
		if err == nil {
			return string(out)
		}
	}
	return "[]"
}

func clampRating(v int) int {
	if v < 0 {
		return 0
	}
	if v > 5 {
		return 5
	}
	return v
}

// GetAllPromptInsights 返回全部心得（扁平列表，带 prompt_id），供导出等批量场景使用。
// GET /api/prompt-insights
func (h *Handler) GetAllPromptInsights(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT " + insightColumns + " FROM prompt_insights ORDER BY prompt_id ASC, created_at DESC, id DESC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := []models.PromptInsight{}
	for rows.Next() {
		item, err := scanInsight(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, item)
	}
	respondOK(w, list)
}

// GetPromptInsights 返回某张卡片的心得，时间倒序。
// GET /api/prompts/{id}/insights
func (h *Handler) GetPromptInsights(w http.ResponseWriter, r *http.Request) {
	promptID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	rows, err := h.db.Query("SELECT "+insightColumns+" FROM prompt_insights WHERE prompt_id = ? ORDER BY created_at DESC, id DESC", promptID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := []models.PromptInsight{}
	for rows.Next() {
		item, err := scanInsight(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, item)
	}
	respondOK(w, list)
}

// CreatePromptInsight 给某张卡片追加一条心得。
// POST /api/prompts/{id}/insights
func (h *Handler) CreatePromptInsight(w http.ResponseWriter, r *http.Request) {
	promptID, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	// 卡片必须存在：外键开着，否则会直接抛 constraint 错误；这里给个更友好的提示
	var exists int
	if err := h.db.QueryRow("SELECT COUNT(1) FROM prompts WHERE id = ?", promptID).Scan(&exists); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if exists == 0 {
		respondFail(w, 404, "提示词不存在")
		return
	}

	var req struct {
		Content string          `json:"content"`
		Rating  int             `json:"rating"`
		Tags    json.RawMessage `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		respondFail(w, 400, "心得内容不能为空")
		return
	}

	res, err := h.db.Exec(
		"INSERT INTO prompt_insights (prompt_id, content, rating, tags) VALUES (?, ?, ?, ?)",
		promptID, content, clampRating(req.Rating), normalizeTags(req.Tags),
	)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := res.LastInsertId()

	item, err := scanInsight(h.db.QueryRow("SELECT "+insightColumns+" FROM prompt_insights WHERE id = ?", id))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, item)
}

// UpdatePromptInsight 修改一条心得（content / rating / tags 都是可选的，没带就保留原值）。
// PUT /api/prompt-insights/{id}
func (h *Handler) UpdatePromptInsight(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	body, err := readAllBody(r)
	if err != nil {
		respondFail(w, 400, "请求体读取失败: "+err.Error())
		return
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(body, &keys); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	var req struct {
		Content string          `json:"content"`
		Rating  int             `json:"rating"`
		Tags    json.RawMessage `json:"tags"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	sets := []string{}
	args := []interface{}{}
	if _, has := keys["content"]; has {
		content := strings.TrimSpace(req.Content)
		if content == "" {
			respondFail(w, 400, "心得内容不能为空")
			return
		}
		sets = append(sets, "content=?")
		args = append(args, content)
	}
	if _, has := keys["rating"]; has {
		sets = append(sets, "rating=?")
		args = append(args, clampRating(req.Rating))
	}
	if _, has := keys["tags"]; has {
		sets = append(sets, "tags=?")
		args = append(args, normalizeTags(req.Tags))
	}
	if len(sets) == 0 {
		respondFail(w, 400, "没有需要更新的字段")
		return
	}
	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	args = append(args, id)

	res, err := h.db.Exec("UPDATE prompt_insights SET "+strings.Join(sets, ", ")+" WHERE id=?", args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// 内容没变时 RowsAffected 也可能是 0，这里不当作错误，只确保记录存在
		var exists int
		if err := h.db.QueryRow("SELECT COUNT(1) FROM prompt_insights WHERE id = ?", id).Scan(&exists); err != nil || exists == 0 {
			respondFail(w, 404, "心得不存在")
			return
		}
	}

	item, err := scanInsight(h.db.QueryRow("SELECT "+insightColumns+" FROM prompt_insights WHERE id = ?", id))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "心得不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, item)
}

// DeletePromptInsight 删除一条心得。
// DELETE /api/prompt-insights/{id}
func (h *Handler) DeletePromptInsight(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	res, err := h.db.Exec("DELETE FROM prompt_insights WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		respondFail(w, 404, "心得不存在")
		return
	}
	respondOK(w, "删除成功")
}
