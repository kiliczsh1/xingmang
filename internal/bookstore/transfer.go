package bookstore

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ErrPickCancelled 表示用户在系统对话框里点了取消。
var ErrPickCancelled = errors.New("已取消")

// ImportFolder 把已有目录导入为作品。
//
// 目录在作品库内 → 原地登记；在库外 → 复制进来（不移动用户的原目录，避免误操作丢数据）。
// 有 book.json 就用它的元数据，没有就用目录名当书名。
func (s *Service) ImportFolder(src string) (*models.Book, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return nil, fmt.Errorf("请提供目录路径")
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("目录不存在: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("不是目录: %s", abs)
	}
	root, err := s.Root()
	if err != nil {
		return nil, err
	}

	// 是否已经在作品库里
	var rel string
	if r, err := filepath.Rel(root, abs); err == nil && r != "." && !strings.HasPrefix(r, "..") {
		rel = filepath.ToSlash(r)
	}

	meta := readBookJSON(abs)
	title := meta.Title
	if strings.TrimSpace(title) == "" {
		title = filepath.Base(abs)
	}

	// 库外目录先复制到库里
	if rel == "" {
		dstName := uniqueDirName(root, SanitizeName(title))
		dst := filepath.Join(root, dstName)
		if err := copyTree(abs, dst); err != nil {
			os.RemoveAll(dst)
			return nil, fmt.Errorf("复制目录失败: %w", err)
		}
		rel = dstName
	}

	b := &models.Book{
		Title:       title,
		Author:      meta.Author,
		Status:      meta.Status,
		Tags:        meta.Tags,
		TargetWords: meta.Target,
		FolderPath:  rel,
	}
	if b.Status == "" {
		b.Status = "draft"
	}
	res, err := s.DB.Exec(`INSERT INTO books (title, description, author, category, tags, status, is_public, target_words, folder_path)
		VALUES (?, '', ?, '', ?, ?, 0, ?, ?)`,
		b.Title, b.Author, b.Tags.ToDBValue(), b.Status, b.TargetWords, rel)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	b.ID = int(id)

	// 补上/更新 book.json，让目录自描述
	meta.Schema = 1
	meta.ID = int(id)
	meta.Title = b.Title
	_ = writeBookJSON(filepath.Join(root, filepath.FromSlash(rel), "book.json"), meta)

	if _, err := s.ScanBook(b.ID); err != nil {
		return b, fmt.Errorf("导入成功但索引失败: %w", err)
	}
	return b, nil
}

func readBookJSON(dir string) bookJSON {
	var meta bookJSON
	raw, err := os.ReadFile(filepath.Join(dir, "book.json"))
	if err != nil {
		return meta
	}
	_ = json.Unmarshal(raw, &meta)
	return meta
}

// copyTree 递归复制；跳过 .git 与 _trash，避免把版本库/回收站搬进来。
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "_trash" {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		info, err := d.Info()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(filepath.Join(dst, rel), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}

// ExportZip 把整本书目录打成 zip 流。
//
// 条目名以书目录名为根（`太荒神诀/001 序章.md`），解压出来就是一个完整作品目录。
func (s *Service) ExportZip(bookID int, w io.Writer) error {
	dir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	var title string
	s.DB.QueryRow(`SELECT title FROM books WHERE id=?`, bookID).Scan(&title)
	if strings.TrimSpace(title) == "" {
		title = filepath.Base(dir)
	}

	zw := zip.NewWriter(w)
	defer zw.Close()

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil || rel == "." {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "_trash" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = title + "/" + filepath.ToSlash(rel)
		hdr.Method = zip.Deflate
		fw, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(fw, f)
		return err
	})
	return err
}

// ImportBookJSON 兼容旧版本导出的作品 JSON。
//
// 旧导出的结构在不同版本里不完全一致，这里按"能认出多少用多少"解析：
//   - { "book": {...}, "chapters": [...] }
//   - { "title": "...", "chapters": [...] }
//   - [ {title, content}, ... ]（纯章节数组，书名用参数兜底）
//
// 每章都会写成真实的 `NNN 标题.md`，不是塞回数据库。
func (s *Service) ImportBookJSON(raw []byte, fallbackTitle string) (*models.Book, int, error) {
	var root interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, 0, fmt.Errorf("不是合法的 JSON: %w", err)
	}

	meta := map[string]interface{}{}
	var chapterList []interface{}
	switch v := root.(type) {
	case map[string]interface{}:
		meta = v
		if b, ok := v["book"].(map[string]interface{}); ok {
			meta = b
			if cl, ok := v["chapters"].([]interface{}); ok {
				chapterList = cl
			}
		}
		if chapterList == nil {
			if cl, ok := meta["chapters"].([]interface{}); ok {
				chapterList = cl
			}
		}
	case []interface{}:
		chapterList = v
	}

	title := firstString(meta, "title", "name", "book_title", "bookTitle")
	if strings.TrimSpace(title) == "" {
		title = strings.TrimSuffix(fallbackTitle, filepath.Ext(fallbackTitle))
	}
	if strings.TrimSpace(title) == "" {
		title = "导入作品"
	}

	b := &models.Book{
		Title:       title,
		Description: firstString(meta, "description", "desc", "intro"),
		Author:      firstString(meta, "author"),
		Status:      firstString(meta, "status"),
	}
	if b.Status == "" {
		b.Status = "draft"
	}
	if tags, ok := meta["tags"].([]interface{}); ok {
		for _, t := range tags {
			if s, ok := t.(string); ok && strings.TrimSpace(s) != "" {
				b.Tags = append(b.Tags, s)
			}
		}
	}

	if err := s.CreateBook(b); err != nil {
		return nil, 0, err
	}

	count := 0
	for i, item := range chapterList {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ct := firstString(m, "title", "name", "chapter_title")
		if strings.TrimSpace(ct) == "" {
			ct = fmt.Sprintf("第 %d 章", i+1)
		}
		body := firstString(m, "content", "text", "body", "chapter_content")
		c := &models.Chapter{
			BookID:  b.ID,
			Title:   ct,
			Content: body,
			Summary: firstString(m, "summary", "outline"),
		}
		if err := s.CreateChapter(c); err != nil {
			return b, count, fmt.Errorf("写入第 %d 章失败: %w", i+1, err)
		}
		count++
	}

	if _, err := s.ScanBook(b.ID); err != nil {
		return b, count, nil
	}
	return b, count, nil
}

func firstString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}
