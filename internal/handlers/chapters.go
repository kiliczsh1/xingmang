package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func (h *Handler) GetChaptersByBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}
	rows, err := h.db.Query("SELECT id, book_id, title, content, summary, order_num, type, volume_id, created_at, updated_at FROM chapters WHERE book_id = ? ORDER BY order_num ASC, id ASC", bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var chapters []models.Chapter
	for rows.Next() {
		var c models.Chapter
		if err := rows.Scan(&c.ID, &c.BookID, &c.Title, &c.Content, &c.Summary, &c.OrderNum, &c.Type, &c.VolumeID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		chapters = append(chapters, c)
	}
	respondOK(w, chapters)
}

func (h *Handler) GetChapter(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var c models.Chapter
	err := h.db.QueryRow("SELECT id, book_id, title, content, summary, order_num, type, volume_id, created_at, updated_at FROM chapters WHERE id = ?", id).
		Scan(&c.ID, &c.BookID, &c.Title, &c.Content, &c.Summary, &c.OrderNum, &c.Type, &c.VolumeID, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "章节不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, c)
}

func (h *Handler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	var c models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	result, err := h.db.Exec("INSERT INTO chapters (book_id, title, content, summary, order_num, type, volume_id) VALUES (?, ?, ?, ?, ?, ?, ?)",
		c.BookID, c.Title, c.Content, c.Summary, c.OrderNum, c.Type, c.VolumeID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := result.LastInsertId()
	c.ID = int(id)
	h.touchBook(c.BookID)
	respondOK(w, c)
}

func (h *Handler) UpdateChapter(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var c models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	// 更新前记录所属书，用于刷新书的 updated_at
	var bookID int
	h.db.QueryRow("SELECT book_id FROM chapters WHERE id=?", id).Scan(&bookID)
	_, err := h.db.Exec("UPDATE chapters SET title=?, content=?, summary=?, order_num=?, type=?, volume_id=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		c.Title, c.Content, c.Summary, c.OrderNum, c.Type, c.VolumeID, id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	c.ID = id
	h.touchBook(bookID)
	respondOK(w, c)
}

func (h *Handler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	var bookID int
	h.db.QueryRow("SELECT book_id FROM chapters WHERE id=?", id).Scan(&bookID)
	_, err := h.db.Exec("DELETE FROM chapters WHERE id = ?", id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.touchBook(bookID)
	respondOK(w, "删除成功")
}

func (h *Handler) ImportChapters(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BookID   int              `json:"bookId"`
		Chapters []models.Chapter `json:"chapters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if req.BookID == 0 {
		respondFail(w, 400, "缺少书籍ID")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	// Get max order_num
	var maxOrder int
	tx.QueryRow("SELECT COALESCE(MAX(order_num), -1) FROM chapters WHERE book_id = ?", req.BookID).Scan(&maxOrder)

	var inserted []models.Chapter
	for i, ch := range req.Chapters {
		orderNum := maxOrder + 1 + i
		result, err := tx.Exec("INSERT INTO chapters (book_id, title, content, summary, order_num, type, volume_id) VALUES (?, ?, ?, ?, ?, 'chapter', NULL)",
			req.BookID, ch.Title, ch.Content, ch.Summary, orderNum)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		id, _ := result.LastInsertId()
		ch.ID = int(id)
		ch.BookID = req.BookID
		ch.OrderNum = orderNum
		inserted = append(inserted, ch)
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.touchBook(req.BookID)
	respondOK(w, map[string]interface{}{"insertedCount": len(inserted), "chapters": inserted})
}

// ImportFile handles file upload for chapter import (supports .txt, placeholder for .docx)
// POST /api/chapters/import-file
// Multipart form: "file" (file), "bookId" (form field)
func (h *Handler) ImportFile(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 20MB)
	if err := r.ParseMultipartForm(20 << 20); err != nil {
		respondFail(w, 400, "无法解析上传文件，文件大小不能超过20MB")
		return
	}

	bookIDStr := r.FormValue("bookId")
	if bookIDStr == "" {
		respondFail(w, 400, "缺少书籍ID")
		return
	}
	bookID, err := strconv.Atoi(bookIDStr)
	if err != nil {
		respondFail(w, 400, "无效的书籍ID")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondFail(w, 400, "请选择要导入的文件")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	var textContent string

	switch ext {
	case ".txt":
		data, err := io.ReadAll(file)
		if err != nil {
			respondFail(w, 500, "读取文件失败")
			return
		}
		textContent = string(data)
	case ".docx":
		// TODO: .docx support requires a dedicated library (e.g. unioffice)
		// For now, return a helpful error
		respondFail(w, 400, "暂不支持 .docx 文件，请将 docx 转为 txt 后再导入")
		return
	default:
		respondFail(w, 400, "不支持的文件类型，仅支持 txt 和 docx")
		return
	}

	chapters := parseChaptersFromText(textContent, header.Filename)
	if len(chapters) == 0 {
		respondFail(w, 400, "文件内容为空或无法识别章节格式")
		return
	}

	// Insert chapters into DB
	tx, err := h.db.Begin()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	var maxOrder int
	tx.QueryRow("SELECT COALESCE(MAX(order_num), -1) FROM chapters WHERE book_id = ?", bookID).Scan(&maxOrder)

	var inserted []models.Chapter
	for i, ch := range chapters {
		orderNum := maxOrder + 1 + i
		result, err := tx.Exec("INSERT INTO chapters (book_id, title, content, summary, order_num, type, volume_id) VALUES (?, ?, ?, '', ?, 'chapter', NULL)",
			bookID, ch.Title, ch.Content, orderNum)
		if err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		id, _ := result.LastInsertId()
		ch.ID = int(id)
		ch.BookID = bookID
		ch.OrderNum = orderNum
		inserted = append(inserted, ch)
	}

	if err := tx.Commit(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.touchBook(bookID)
	respondOK(w, map[string]interface{}{"insertedCount": len(inserted), "chapters": inserted})
}

// parseChaptersFromText splits text content into chapters using common chapter patterns.
// Patterns: "第X章/节/回/卷/集/部/篇" (Chinese), "Chapter X" (English).
// If no pattern matches, the entire text is returned as a single chapter.
func parseChaptersFromText(text, filename string) []models.Chapter {
	// Chinese chapter pattern: 第[一二三四五六七八九十百千零\d]+[章节回卷集部篇][^\n]*
	chinesePattern := `第[一二三四五六七八九十百千零\d]+[章节回卷集部篇][^\n]*`
	// English chapter pattern: Chapter \d+[^\n]*
	englishPattern := `[Cc]hapter\s+\d+[^\n]*`

	pattern := chinesePattern + `|` + englishPattern
	re := regexp.MustCompile(pattern)

	matches := re.FindAllStringIndex(text, -1)
	var chapters []models.Chapter

	if len(matches) > 0 {
		for i, loc := range matches {
			startIdx := loc[0]
			endIdx := len(text)
			if i < len(matches)-1 {
				endIdx = matches[i+1][0]
			}
			title := strings.TrimSpace(text[loc[0]:loc[1]])
			content := strings.TrimSpace(text[startIdx:endIdx])
			if title == "" && content == "" {
				continue
			}
			chapters = append(chapters, models.Chapter{Title: title, Content: content})
		}
	} else {
		// No chapter markers found: treat entire file as one chapter
		baseName := filename
		if idx := strings.LastIndex(filename, "."); idx != -1 {
			baseName = filename[:idx]
		}
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return nil
		}
		chapters = append(chapters, models.Chapter{Title: baseName, Content: trimmed})
	}

	return chapters
}
