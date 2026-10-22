package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

func (h *Handler) GetChaptersByBook(w http.ResponseWriter, r *http.Request) {
	bookID, ok := idParam(r, "bookId")
	if !ok {
		respondFail(w, 400, "无效的书籍ID")
		return
	}
	// 打开一本书时顺手对账：外部新增/删除的文件在这里进入索引，
	// 用户不需要知道"重新扫描"这个动作的存在。
	if _, err := bookstore.NewService(h.db).ScanBook(bookID); err != nil {
		log.Printf("[books] 对账失败 book=%d: %v", bookID, err)
	}

	rows, err := h.db.Query(`SELECT id, book_id, title, content, summary, order_num, type, volume_id,
		COALESCE(file_path,''), created_at, updated_at
		FROM chapters WHERE book_id = ? ORDER BY order_num ASC, id ASC`, bookID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	var chapters []models.Chapter
	for rows.Next() {
		var c models.Chapter
		if err := rows.Scan(&c.ID, &c.BookID, &c.Title, &c.Content, &c.Summary, &c.OrderNum, &c.Type, &c.VolumeID,
			&c.FilePath, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		chapters = append(chapters, c)
	}
	// 必须先关掉 rows 再查库：SQLite 单连接下，边遍历边查询会自己把自己锁死。
	rows.Close()

	// 一次拿到书目录，后面逐章读文件；读不到就标 missing，不让整个列表 500。
	dir, dirErr := bookstore.NewService(h.db).BookDir(bookID)
	for i := range chapters {
		c := &chapters[i]
		if c.FilePath == "" || dirErr != nil {
			continue
		}
		text, _, err := bookstore.ReadTextFile(filepath.Join(dir, filepath.FromSlash(c.FilePath)))
		if err != nil {
			c.Missing = 1
			c.Content = ""
			continue
		}
		fm := bookstore.ParseFrontMatter(text)
		c.Content = fm.Body
		if sm := fm.Get("summary"); sm != "" {
			c.Summary = sm
		}
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
	err := h.db.QueryRow(`SELECT id, book_id, title, content, summary, order_num, type, volume_id,
		COALESCE(file_path,''), created_at, updated_at FROM chapters WHERE id = ?`, id).
		Scan(&c.ID, &c.BookID, &c.Title, &c.Content, &c.Summary, &c.OrderNum, &c.Type, &c.VolumeID,
			&c.FilePath, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "章节不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	// 文件是唯一真相：正文与摘要从文件读（同时顺手校正库里的字数/哈希）。
	if text, err := bookstore.NewService(h.db).ChapterText(id); err == nil {
		c.Content = text
	}
	respondOK(w, c)
}

func (h *Handler) CreateChapter(w http.ResponseWriter, r *http.Request) {
	var c models.Chapter
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	// 写文件 + 插索引，由 bookstore 保证顺序：文件写失败就不建索引。
	if err := bookstore.NewService(h.db).CreateChapter(&c); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
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
	c.ID = id
	// 正文写盘（标题变化时顺带重命名文件），再更新索引。
	if err := bookstore.NewService(h.db).UpdateChapter(&c); err != nil {
		if errors.Is(err, bookstore.ErrFileConflict) {
			// 409：让前端弹"载入磁盘版 / 保留当前 / 取消"，绝不静默覆盖
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"code":    "file_conflict",
				"message": "磁盘上的文件已被外部修改，请选择要保留的版本",
			})
			return
		}
		respondFail(w, 500, err.Error())
		return
	}
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
	var filePath string
	h.db.QueryRow("SELECT book_id, COALESCE(file_path,'') FROM chapters WHERE id=?", id).Scan(&bookID, &filePath)
	// 先归档文件再删索引：直接删文件等于永久丢失。
	if filePath != "" {
		if err := bookstore.NewService(h.db).TrashChapterFile(bookID, filePath); err != nil {
			respondFail(w, 500, "移入回收目录失败: "+err.Error())
			return
		}
	}
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

	// 逐章写文件 + 建索引；任一章失败即中止，并告知是第几章。
	svc := bookstore.NewService(h.db)
	var inserted []models.Chapter
	for i := range req.Chapters {
		ch := req.Chapters[i]
		ch.BookID = req.BookID
		ch.Type = "chapter"
		ch.OrderNum = 0 // 由 service 按当前最大序号顺延
		if err := svc.CreateChapter(&ch); err != nil {
			respondFail(w, 500, "写入第 "+strconv.Itoa(i+1)+" 章失败: "+err.Error())
			return
		}
		inserted = append(inserted, ch)
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

	// 逐章写文件 + 建索引（.txt 导入同样落盘，不塞进数据库）
	svc := bookstore.NewService(h.db)
	var inserted []models.Chapter
	for i := range chapters {
		ch := chapters[i]
		ch.BookID = bookID
		ch.Type = "chapter"
		ch.OrderNum = 0
		if err := svc.CreateChapter(&ch); err != nil {
			respondFail(w, 500, "写入第 "+strconv.Itoa(i+1)+" 章失败: "+err.Error())
			return
		}
		inserted = append(inserted, ch)
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
