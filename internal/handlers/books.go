package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// bookSelectCols 查询书，并实时聚合章节字数与章数（供前端卡片/仪表盘展示）。
// word_count 用 SUM(LENGTH(content))：SQLite 的 LENGTH 对文本按字符计数，中文按字计。
const bookSelectCols = `
SELECT b.id, b.title, b.description, b.cover, b.author, b.category, b.tags, b.status, b.is_public, b.target_words, b.created_at, b.updated_at,
       COALESCE((SELECT COUNT(*) FROM chapters ch WHERE ch.book_id = b.id AND ch.type = 'chapter'), 0) AS chapter_count,
       COALESCE((SELECT SUM(LENGTH(ch.content)) FROM chapters ch WHERE ch.book_id = b.id AND ch.type = 'chapter'), 0) AS word_count
FROM books b`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanBook 按 bookSelectCols 的列顺序把一行解析为 Book（tags 文本 → 数组）。
func scanBook(row rowScanner) (models.Book, error) {
	var b models.Book
	var tagsStr string
	var tw sql.NullInt64 // target_words 可能为 NULL（未设置目标）
	err := row.Scan(&b.ID, &b.Title, &b.Description, &b.Cover, &b.Author, &b.Category, &tagsStr, &b.Status, &b.IsPublic, &tw, &b.CreatedAt, &b.UpdatedAt, &b.ChapterCount, &b.WordCount)
	if err != nil {
		return b, err
	}
	if tw.Valid {
		b.TargetWords = int(tw.Int64)
	}
	b.Tags = parseTags(tagsStr)
	return b, nil
}

// parseTags 把 DB 中的 tags 文本解析为数组，兼容 JSON 数组字符串与逗号分隔文本。
func parseTags(tagsStr string) models.StringList {
	tagsStr = strings.TrimSpace(tagsStr)
	if tagsStr == "" || tagsStr == "null" {
		return models.StringList{}
	}
	var list models.StringList
	if err := json.Unmarshal([]byte(tagsStr), &list); err == nil {
		return list
	}
	parts := strings.Split(tagsStr, ",")
	cleaned := make(models.StringList, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			cleaned = append(cleaned, t)
		}
	}
	return cleaned
}

func (h *Handler) GetBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(bookSelectCols + " ORDER BY b.updated_at DESC, b.id DESC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var books []models.Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		books = append(books, b)
	}
	respondOK(w, books)
}

func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	b, err := scanBook(h.db.QueryRow(bookSelectCols+" WHERE b.id = ?", id))
	if err == sql.ErrNoRows {
		respondFail(w, 404, "书本不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, b)
}

func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	var b models.Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	result, err := h.db.Exec("INSERT INTO books (title, description, cover, author, category, tags, status, is_public, target_words) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		b.Title, b.Description, b.Cover, b.Author, b.Category, b.Tags.ToDBValue(), b.Status, b.IsPublic, b.TargetWords)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := result.LastInsertId()
	b.ID = int(id)
	if b.Tags == nil {
		b.Tags = models.StringList{}
	}
	respondOK(w, b)
}

func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if len(payload) == 0 {
		respondFail(w, 400, "没有可更新的字段")
		return
	}

	// 仅更新请求中出现的字段，避免把未提交的字段覆盖为空值
	sets := []string{}
	args := []interface{}{}
	setString := func(col string, raw json.RawMessage) {
		var s string
		if err := json.Unmarshal(raw, &s); err == nil {
			sets = append(sets, col+"=?")
			args = append(args, s)
		}
	}
	if v, ok := payload["title"]; ok {
		setString("title", v)
	}
	if v, ok := payload["description"]; ok {
		setString("description", v)
	}
	if v, ok := payload["cover"]; ok {
		setString("cover", v)
	}
	if v, ok := payload["author"]; ok {
		setString("author", v)
	}
	if v, ok := payload["category"]; ok {
		setString("category", v)
	}
	if v, ok := payload["status"]; ok {
		setString("status", v)
	}
	if v, ok := payload["tags"]; ok {
		var list models.StringList
		if err := json.Unmarshal(v, &list); err == nil {
			sets = append(sets, "tags=?")
			args = append(args, list.ToDBValue())
		}
	}
	if v, ok := payload["is_public"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err == nil {
			val := 0
			if b {
				val = 1
			}
			sets = append(sets, "is_public=?")
			args = append(args, val)
		} else {
			var n int
			if err := json.Unmarshal(v, &n); err == nil {
				sets = append(sets, "is_public=?")
				args = append(args, n)
			}
		}
	}
	if v, ok := payload["target_words"]; ok {
		var n int
		if err := json.Unmarshal(v, &n); err == nil {
			sets = append(sets, "target_words=?")
			args = append(args, n)
		} else if string(v) == "null" {
			sets = append(sets, "target_words=NULL")
		}
	}
	if len(sets) == 0 {
		respondFail(w, 400, "没有可更新的字段")
		return
	}
	sets = append(sets, "updated_at=CURRENT_TIMESTAMP")
	args = append(args, id)
	query := "UPDATE books SET " + strings.Join(sets, ", ") + " WHERE id=?"
	if _, err := h.db.Exec(query, args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	b, err := scanBook(h.db.QueryRow(bookSelectCols+" WHERE b.id = ?", id))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, b)
}

func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	_, err := h.db.Exec("DELETE FROM books WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
