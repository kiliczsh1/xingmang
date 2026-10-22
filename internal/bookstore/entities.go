package bookstore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 实体落盘：人物 / 世界观 / 大纲 / 灵感 / 备忘录。
//
// 正文进文件、结构化字段进 front-matter，文件名即标题。
// 目录：
//
//	characters/<分组>/<姓名>.md     分组可嵌套（folder 里的 "/" 视为层级）
//	world/<分类>/<条目名>.md
//	outlines/<标题>.md, ideas/<标题>.md
//	memos/<标题>.md（书内）；_memos/<标题>.md（bookID=0 的全局便签）
var entityKinds = map[string]string{
	"character": "characters",
	"entry":     "world",
	"outline":   "outlines",
	"idea":      "ideas",
	"memo":      "memos",
}

// entityRoot 返回实体目录。
func (s *Service) entityRoot(bookID int, kind string) (string, error) {
	sub, ok := entityKinds[kind]
	if !ok {
		return "", fmt.Errorf("未知实体类型: %s", kind)
	}
	if kind == "memo" && bookID == 0 {
		root, err := s.Root()
		if err != nil {
			return "", err
		}
		dir := filepath.Join(root, "_memos")
		return dir, os.MkdirAll(dir, 0o755)
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(bookDir, sub)
	return dir, os.MkdirAll(dir, 0o755)
}

// entityRel 计算实体的相对路径（相对书目录；全局备忘录相对 books_root）。
func entityRel(kind, group, name string) string {
	sub := entityKinds[kind]
	parts := []string{sub}
	if group != "" {
		for _, seg := range strings.Split(group, "/") {
			if s := SanitizeName(seg); s != "" && s != "未命名" {
				parts = append(parts, s)
			}
		}
	}
	parts = append(parts, SanitizeName(name)+".md")
	return filepath.ToSlash(filepath.Join(parts...))
}

// SaveEntity 写实体文件；oldRel 非空时视为改名/换分组，写完删旧文件。
//
// 从旧文件出发做 front-matter 补丁，保留用户自己加的字段。
func (s *Service) SaveEntity(bookID int, kind, name, group, body string, fields [][2]string, oldRel string) (string, error) {
	if strings.TrimSpace(name) == "" {
		name = "未命名"
	}
	rootDir, err := s.entityRoot(bookID, kind)
	if err != nil {
		return "", err
	}
	rel := entityRel(kind, group, name)
	// 去掉 kind 目录后的部分用于定位到 rootDir
	relInRoot := strings.TrimPrefix(rel, entityKinds[kind]+"/")
	full := filepath.Join(rootDir, filepath.FromSlash(relInRoot))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	// 新建或改名时若目标已存在（同名实体），加后缀而不是覆盖别人的文件
	if rel != oldRel {
		if _, err := os.Stat(full); err == nil {
			dir := filepath.Dir(full)
			base := UniqueName(dir, filepath.Base(full))
			full = filepath.Join(dir, base)
			rel = filepath.ToSlash(filepath.Join(filepath.Dir(relInRoot), base))
			if filepath.Dir(relInRoot) == "." {
				rel = entityKinds[kind] + "/" + base
			} else {
				rel = entityKinds[kind] + "/" + rel
			}
		}
	}

	existing := ""
	var oldFull string
	if oldRel != "" {
		base, err := s.entityBaseDir(bookID, kind)
		if err != nil {
			return "", err
		}
		oldFull = filepath.Join(base, filepath.FromSlash(strings.TrimPrefix(oldRel, entityKinds[kind]+"/")))
		if raw, err := os.ReadFile(oldFull); err == nil {
			existing = string(raw)
		}
	}

	fm := ParseFrontMatter(existing)
	fm.Body = NormalizeText(body)
	for _, kv := range fields {
		fm = fm.WithField(kv[0], kv[1])
	}
	text := fm.Render()
	if _, err := WriteTextFile(full, text); err != nil {
		return "", err
	}
	if oldFull != "" && filepath.Clean(oldFull) != filepath.Clean(full) {
		_ = os.Remove(oldFull)
	}
	return rel, nil
}

// entityBaseDir 返回实体目录的绝对路径（与 entityRoot 一致，但不创建）。
func (s *Service) entityBaseDir(bookID int, kind string) (string, error) {
	sub := entityKinds[kind]
	if kind == "memo" && bookID == 0 {
		root, err := s.Root()
		if err != nil {
			return "", err
		}
		return filepath.Join(root, "_memos"), nil
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	return filepath.Join(bookDir, sub), nil
}

// ReadEntity 读取实体正文与字段。
func (s *Service) ReadEntity(bookID int, kind, rel string) (map[string]string, string, error) {
	base, err := s.entityBaseDir(bookID, kind)
	if err != nil {
		return nil, "", err
	}
	full := filepath.Join(base, filepath.FromSlash(strings.TrimPrefix(rel, entityKinds[kind]+"/")))
	text, _, err := ReadTextFile(full)
	if err != nil {
		return nil, "", err
	}
	fm := ParseFrontMatter(text)
	out := map[string]string{}
	for _, line := range fm.Lines {
		if i := strings.Index(line, ":"); i > 0 {
			out[strings.TrimSpace(line[:i])] = strings.Trim(strings.TrimSpace(line[i+1:]), `"'`)
		}
	}
	return out, fm.Body, nil
}

// TrashEntity 删除实体文件（优先系统回收站）。
func (s *Service) TrashEntity(bookID int, kind, rel string) error {
	if rel == "" {
		return nil
	}
	base, err := s.entityBaseDir(bookID, kind)
	if err != nil {
		return err
	}
	full := filepath.Join(base, filepath.FromSlash(strings.TrimPrefix(rel, entityKinds[kind]+"/")))
	if _, err := os.Stat(full); os.IsNotExist(err) {
		return nil
	}
	if err := moveToRecycleBin(full); err == nil {
		return nil
	}
	trash := filepath.Join(base, "_trash")
	if err := os.MkdirAll(trash, 0o755); err != nil {
		return err
	}
	return os.Rename(full, filepath.Join(trash, UniqueName(trash, filepath.Base(full))))
}

// SyncCharacter 把人物同步到磁盘，返回相对书目录的路径。
func (s *Service) SyncCharacter(c *models.Character, oldRel string) (string, error) {
	fields := [][2]string{
		{"gender", c.Gender},
		{"personality", ps(c.Personality)},
		{"folders", c.Folders},
	}
	group := c.Folder
	if group == "" && c.Folders != "" {
		// Folders 是 JSON 数组字符串，这里只取第一个作为目录（其余留在 front-matter）
		group = firstJSONString(c.Folders)
	}
	return s.SaveEntity(c.BookID, "character", c.Name, group, ps(c.Info), fields, oldRel)
}

// SyncEntry 把世界观条目同步到磁盘。
func (s *Service) SyncEntry(e *models.Entry, category string, oldRel string) (string, error) {
	fields := [][2]string{
		{"category", category},
		{"tags", e.Tags},
		{"appearance", e.Appearance},
		{"background", e.Background},
		{"personality", ps(e.Personality)},
		{"relationships", ps(e.Relationships)},
		{"fields_json", e.CustomFields},
	}
	bookID := 0
	if e.BookID != nil {
		bookID = *e.BookID
	}
	return s.SaveEntity(bookID, "entry", e.Name, category, e.Description, fields, oldRel)
}

// SyncMemo 把备忘录同步到磁盘（bookID=0 即全局便签）。
func (s *Service) SyncMemo(m *models.Memo, bookID int, oldRel string) (string, error) {
	fields := [][2]string{
		{"category", m.Category},
		{"tags", m.Tags},
		{"pinned", fmt.Sprintf("%d", m.IsPinned)},
	}
	return s.SaveEntity(bookID, "memo", m.Title, "", m.Content, fields, oldRel)
}

func ps(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func firstJSONString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(parts[0]), `"'`)
}

// ---------------- 书内自由文档：大纲 / 灵感 ----------------

// CreateDoc 写文件 + 插索引。
func (s *Service) CreateDoc(bookID int, kind, title, content string) (*models.BookDoc, error) {
	if kind != "outline" && kind != "idea" {
		return nil, fmt.Errorf("未知文档类型: %s", kind)
	}
	if strings.TrimSpace(title) == "" {
		title = "未命名"
	}
	rel, err := s.SaveEntity(bookID, kind, title, "", content, nil, "")
	if err != nil {
		return nil, err
	}
	res, err := s.DB.Exec(`INSERT INTO book_docs (book_id, kind, title, file_path, word_count) VALUES (?, ?, ?, ?, ?)`,
		bookID, kind, title, rel, WordCount(content))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &models.BookDoc{ID: int(id), BookID: bookID, Kind: kind, Title: title, FilePath: rel, Content: content, WordCount: WordCount(content)}, nil
}

// UpdateDoc 改写正文/标题；标题变化会重命名文件。
func (s *Service) UpdateDoc(id int, title, content string) (*models.BookDoc, error) {
	var doc models.BookDoc
	err := s.DB.QueryRow(`SELECT book_id, kind, title, COALESCE(file_path,'') FROM book_docs WHERE id=?`, id).
		Scan(&doc.BookID, &doc.Kind, &doc.Title, &doc.FilePath)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(title) == "" {
		title = doc.Title
	}
	rel, err := s.SaveEntity(doc.BookID, doc.Kind, title, "", content, nil, doc.FilePath)
	if err != nil {
		return nil, err
	}
	_, err = s.DB.Exec(`UPDATE book_docs SET title=?, file_path=?, word_count=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		title, rel, WordCount(content), id)
	if err != nil {
		return nil, err
	}
	doc.ID = id
	doc.Title = title
	doc.FilePath = rel
	doc.Content = content
	doc.WordCount = WordCount(content)
	return &doc, nil
}

// ListDocs 列出某类文档，正文从文件读。
func (s *Service) ListDocs(bookID int, kind string) ([]models.BookDoc, error) {
	rows, err := s.DB.Query(`SELECT id, book_id, kind, title, COALESCE(file_path,''), word_count, pinned, created_at, updated_at
		FROM book_docs WHERE book_id=? AND kind=? ORDER BY pinned DESC, updated_at DESC`, bookID, kind)
	if err != nil {
		return nil, err
	}
	var out []models.BookDoc
	for rows.Next() {
		var d models.BookDoc
		if err := rows.Scan(&d.ID, &d.BookID, &d.Kind, &d.Title, &d.FilePath, &d.WordCount, &d.Pinned, &d.CreatedAt, &d.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, d)
	}
	rows.Close()
	for i := range out {
		if _, body, err := s.ReadEntity(out[i].BookID, out[i].Kind, out[i].FilePath); err == nil {
			out[i].Content = body
		}
	}
	return out, nil
}

// DeleteDoc 删除文档：文件进回收站，再删索引。
func (s *Service) DeleteDoc(id int) error {
	var bookID int
	var kind, rel string
	if err := s.DB.QueryRow(`SELECT book_id, kind, COALESCE(file_path,'') FROM book_docs WHERE id=?`, id).
		Scan(&bookID, &kind, &rel); err != nil {
		return err
	}
	if err := s.TrashEntity(bookID, kind, rel); err != nil {
		return err
	}
	_, err := s.DB.Exec(`DELETE FROM book_docs WHERE id=?`, id)
	return err
}
