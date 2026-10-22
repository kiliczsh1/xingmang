package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 列表/详情查询在提示词字段后附带两个心得聚合值（count / 评分均值），
// 让卡片徽标不必为每张卡片再发一次请求。rating = 0 表示未评分，不计入均值。
// ⚠ 列顺序必须和 scanPrompt 的 Scan 顺序严格一致。
const promptSelect = `SELECT
  p.id, p.name, p.description, p.content, p.category, p.order_num, p.fields, p.subcategories,
  p.card_type, p.password, p.is_locked, p.insight, p.creator_name, p.version, p.metadata,
  p.created_at, p.updated_at,
  (SELECT COUNT(1) FROM prompt_insights i WHERE i.prompt_id = p.id) AS insight_count,
  (SELECT AVG(i.rating) FROM prompt_insights i WHERE i.prompt_id = p.id AND i.rating > 0) AS insight_avg
FROM prompts p`

func scanPrompt(row interface {
	Scan(dest ...interface{}) error
}) (models.Prompt, error) {
	var p models.Prompt
	var avg sql.NullFloat64
	err := row.Scan(&p.ID, &p.Name, &p.Description, &p.Content, &p.Category, &p.OrderNum, &p.Fields, &p.Subcategories, &p.CardType, &p.Password, &p.IsLocked, &p.Insight, &p.CreatorName, &p.Version, &p.Metadata, &p.CreatedAt, &p.UpdatedAt, &p.InsightCount, &avg)
	if err != nil {
		return p, err
	}
	if p.CardType == "" {
		p.CardType = "normal"
	}
	if avg.Valid {
		p.InsightAvg = avg.Float64
	}
	return p, nil
}

// refreshPromptInsightStats 只重读两个聚合值，用于更新后的响应回显
// （否则 UpdatePrompt 回显的对象会带 insight_count=0，前端合并时徽标会闪一下 0）。
func (h *Handler) refreshPromptInsightStats(p *models.Prompt) {
	var avg sql.NullFloat64
	if err := h.db.QueryRow("SELECT (SELECT COUNT(1) FROM prompt_insights i WHERE i.prompt_id = ?) AS c, (SELECT AVG(i.rating) FROM prompt_insights i WHERE i.prompt_id = ? AND i.rating > 0) AS a", p.ID, p.ID).
		Scan(&p.InsightCount, &avg); err == nil && avg.Valid {
		p.InsightAvg = avg.Float64
	}
}

func (h *Handler) GetPrompts(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(promptSelect + " ORDER BY p.order_num ASC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var prompts []models.Prompt
	for rows.Next() {
		p, err := scanPrompt(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
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
	p, err := scanPrompt(h.db.QueryRow(promptSelect+" WHERE p.id = ?", id))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "提示词不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, p)
}

func (h *Handler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	var p models.Prompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	result, err := h.db.Exec("INSERT INTO prompts (name, description, content, category, order_num, fields, subcategories, card_type, password, is_locked, insight, creator_name, version, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.Description, p.Content, p.Category, p.OrderNum, p.Fields, p.Subcategories, p.CardType, p.Password, p.IsLocked, p.Insight, p.CreatorName, p.Version, p.Metadata)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		respondFail(w, 400, "请求体读取失败: "+err.Error())
		return
	}
	// 先按原始键集合判断请求体是否携带了可选字段
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(body, &keys); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	var p models.Prompt
	if err := json.Unmarshal(body, &p); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	// is_locked / insight 是可选列：请求体未携带该键时保留库中现值。
	// 否则「编辑保存」「锁定切换」这类只关心部分字段的调用会把另一个字段意外清零。
	base := "UPDATE prompts SET name=?, description=?, content=?, category=?, order_num=?, fields=?, subcategories=?, card_type=?, password=?, creator_name=?, version=?, metadata=?"
	args := []interface{}{p.Name, p.Description, p.Content, p.Category, p.OrderNum, p.Fields, p.Subcategories, p.CardType, p.Password, p.CreatorName, p.Version, p.Metadata}
	optional := []struct {
		key string
		val interface{}
	}{
		{"is_locked", p.IsLocked},
		{"insight", p.Insight},
	}
	for _, opt := range optional {
		if _, ok := keys[opt.key]; ok {
			base += ", " + opt.key + "=?"
			args = append(args, opt.val)
		}
	}
	base += ", updated_at=CURRENT_TIMESTAMP WHERE id=?"
	args = append(args, id)

	if _, err := h.db.Exec(base, args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	p.ID = id
	h.refreshPromptInsightStats(&p)
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
