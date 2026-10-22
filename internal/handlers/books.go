package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// bookSelectCols 查询书，并实时聚合章节字数与章数（供前端卡片/仪表盘展示）。
// word_count 用 SUM(LENGTH(content))：SQLite 的 LENGTH 对文本按字符计数，中文按字计。
const bookSelectCols = `
SELECT b.id, b.title, b.description, b.cover, b.author, b.category, b.tags, b.status, b.is_public, b.target_words,
       b.folder_path, b.created_at, b.updated_at,
       COALESCE((SELECT COUNT(*) FROM chapters ch WHERE ch.book_id = b.id AND ch.type = 'chapter'), 0) AS chapter_count,
       COALESCE((SELECT SUM(CASE WHEN ch.file_path <> '' THEN ch.word_count ELSE LENGTH(ch.content) END)
                 FROM chapters ch WHERE ch.book_id = b.id AND ch.type = 'chapter'), 0) AS word_count
FROM books b`

type rowScanner interface {
	Scan(dest ...interface{}) error
}

// scanBook 按 bookSelectCols 的列顺序把一行解析为 Book（tags 文本 → 数组）。
//
// ⚠ books 的 description / cover / author / category / tags / status 都可能是 NULL：
// 库里的存量数据（例如 Agent 早期用 book_create 建的书只写了 title/description/category）
// 就是 NULL。直接扫进 string 会报 "converting NULL to string is unsupported"，
// 让整个书列表接口 500 —— 所以这里一律用 sql.NullXxx 接。
func scanBook(row rowScanner) (models.Book, error) {
	var b models.Book
	var description, cover, author, category, tagsStr, status sql.NullString
	var isPublic sql.NullInt64
	var tw sql.NullInt64 // target_words 可能为 NULL（未设置目标）
	err := row.Scan(&b.ID, &b.Title, &description, &cover, &author, &category, &tagsStr, &status,
		&isPublic, &tw, &b.FolderPath, &b.CreatedAt, &b.UpdatedAt, &b.ChapterCount, &b.WordCount)
	if err != nil {
		return b, err
	}
	b.Description = description.String
	b.Cover = cover.String
	b.Author = author.String
	b.Category = category.String
	b.Status = status.String
	if isPublic.Valid {
		b.IsPublic = int(isPublic.Int64)
	}
	if tw.Valid {
		b.TargetWords = int(tw.Int64)
	}
	b.Tags = parseTags(tagsStr.String)
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
	if strings.TrimSpace(b.Title) == "" {
		respondFail(w, 400, "书名不能为空")
		return
	}
	// 建目录 + 写 book.json + 插库，三件事由 bookstore 保证原子性。
	// 从这里开始，"书"不再只是一行记录，而是一个真实目录。
	if err := bookstore.NewService(h.db).CreateBook(&b); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	created, err := scanBook(h.db.QueryRow(bookSelectCols+" WHERE b.id = ?", b.ID))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, created)
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

	// 书名要同步改目录名：先动磁盘，失败就整体失败（避免库里新名、磁盘旧名）。
	if raw, ok := payload["title"]; ok {
		var title string
		if err := json.Unmarshal(raw, &title); err == nil && strings.TrimSpace(title) != "" {
			if _, err := bookstore.NewService(h.db).RenameBook(id, title); err != nil {
				respondFail(w, 409, err.Error())
				return
			}
		}
		delete(payload, "title")
	}

	// 封面：前端可能提交两种东西 ——
	//   1. 上传图片得到的 data URL；
	//   2. AI 生成后返回的站内 URL（图先落在 data/generated-images，属于暂存区）。
	// 两种都必须**复制进作品文件夹**（covers/ 下的文件），并把旧封面归档到 gallery/。
	// 只认 data: 的话，生成图会被当成普通字符串存进 books.cover —— 封面就"住"在了
	// 缓存目录里，换封面时也没有旧图可归档，表现就是"画廊永远是空的、缓存一清就裂图"。
	if raw, ok := payload["cover"]; ok {
		var cover string
		if err := json.Unmarshal(raw, &cover); err == nil {
			cover = strings.TrimSpace(cover)
			svc := bookstore.NewService(h.db)
			if strings.HasPrefix(cover, "data:") {
				if err := svc.SetCover(id, cover); err != nil {
					respondFail(w, 500, "保存封面失败: "+err.Error())
					return
				}
				payload["cover"] = json.RawMessage(`""`)
			} else if handled, err := promoteCoverFromLocalURL(svc, id, cover); handled {
				if err != nil {
					respondFail(w, 500, "保存封面失败: "+err.Error())
					return
				}
				payload["cover"] = json.RawMessage(`""`)
			}
			// 其他形态（比如已经指向 /api/books/{id}/cover）不处理：那是"当前封面"本身
		}
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
	// 删除前先把整本目录安排好：归档（留在作品库）或回收站（可还原）。
	// 两种都不会直接删文件夹。
	mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
	svc := bookstore.NewService(h.db)
	var err error
	if mode == "trash" {
		err = svc.TrashBook(id)
	} else {
		err = svc.ArchiveBook(id)
	}
	if err != nil {
		respondFail(w, 500, "处理作品目录失败: "+err.Error())
		return
	}
	if _, err := h.db.Exec("DELETE FROM books WHERE id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
