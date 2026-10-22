package bookstore

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// ScanResult 是一次目录对账的结果。
type ScanResult struct {
	AddedChapters   int `json:"added_chapters"`
	MissingChapters int `json:"missing_chapters"`
	AddedCharacters int `json:"added_characters"`
	AddedEntries    int `json:"added_entries"`
	AddedDocs       int `json:"added_docs"`
	AddedMemos      int `json:"added_memos"`
}

// ScanBook 用磁盘内容补索引：
//   - 磁盘上有、库里没有 → 新增行（外部新建的文件能出现在界面里）；
//   - 库里有、磁盘上没有 → missing=1（**不删行**，保住 id 与引用）。
//
// 不处理"改名"：改名在文件系统里就是删+增，这里按新文件新增、旧行标 missing，
// 让用户自己决定是保留旧摘要还是删掉。
func (s *Service) ScanBook(bookID int) (ScanResult, error) {
	var res ScanResult
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return res, err
	}

	// 卷目录 → volume_id，供章节归属
	volMap := map[string]int{}
	if rows, err := s.DB.Query(`SELECT id FROM volumes WHERE book_id=?`, bookID); err == nil {
		var ids []int
		for rows.Next() {
			var id int
			if rows.Scan(&id) == nil {
				ids = append(ids, id)
			}
		}
		rows.Close()
		for _, id := range ids {
			if p, err := s.volumePath(bookID, id); err == nil {
				volMap[p] = id
			}
		}
	}

	seenChapters := map[string]bool{}
	seenChar := map[string]bool{}
	seenEntry := map[string]bool{}
	seenDoc := map[string]bool{}
	seenMemo := map[string]bool{}

	err = filepath.WalkDir(bookDir, func(path string, d fs.DirEntry, werr error) error {
		if werr != nil {
			return nil
		}
		rel, rerr := filepath.Rel(bookDir, path)
		if rerr != nil || rel == "." {
			return nil
		}
		relSlash := filepath.ToSlash(rel)
		base := d.Name()
		if d.IsDir() {
			// 跳过隐藏目录与回收目录（系统目录要进去扫，不能跳过）
			if strings.HasPrefix(base, ".") || base == "_trash" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(base))
		if ext != ".md" && ext != ".txt" {
			return nil
		}

		first := strings.SplitN(relSlash, "/", 2)[0]
		switch first {
		case "covers", "gallery", "assets":
			return nil
		case "memory":
			// 记忆目录（对话/Agent/图谱/拆书/生成历史）自成体系，由 memory 对账单独处理，
			// 绝不能当成章节 —— 否则记一次对话就会在目录树里多出一"章"。
			return nil
		case "characters":
			seenChar[relSlash] = true
			if s.entityExists("character", bookID, relSlash) {
				return nil
			}
			group := strings.TrimSuffix(strings.TrimPrefix(relSlash, "characters/"), "/"+base)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			if s.insertScannedCharacter(bookID, name, group, relSlash) {
				res.AddedCharacters++
			}
			return nil
		case "world":
			seenEntry[relSlash] = true
			if s.entityExists("entry", bookID, relSlash) {
				return nil
			}
			rest := strings.TrimPrefix(relSlash, "world/")
			category := ""
			if i := strings.LastIndex(rest, "/"); i > 0 {
				category = rest[:i]
			}
			name := strings.TrimSuffix(base, filepath.Ext(base))
			if s.insertScannedEntry(bookID, name, category, relSlash) {
				res.AddedEntries++
			}
			return nil
		case "outlines", "ideas":
			kind := first[:len(first)-1] // outline / idea
			seenDoc[relSlash] = true
			if s.docExists(bookID, relSlash) {
				return nil
			}
			name := strings.TrimSuffix(base, filepath.Ext(base))
			if s.insertScannedDoc(bookID, kind, name, relSlash) {
				res.AddedDocs++
			}
			return nil
		case "memos":
			seenMemo[relSlash] = true
			if s.entityExists("memo", bookID, relSlash) {
				return nil
			}
			name := strings.TrimSuffix(base, filepath.Ext(base))
			if s.insertScannedMemo(bookID, name, relSlash) {
				res.AddedMemos++
			}
			return nil
		}

		// 其余一律按章节处理（书根或卷目录）
		seenChapters[relSlash] = true
		if s.chapterExists(bookID, relSlash) {
			s.DB.Exec(`UPDATE chapters SET missing=0 WHERE book_id=? AND file_path=?`, bookID, relSlash)
			return nil
		}
		order, title, ok := ParseOrderedName(base)
		if !ok {
			order = 0
			title = strings.TrimSuffix(base, filepath.Ext(base))
		}
		dirRel := filepath.ToSlash(filepath.Dir(relSlash))
		var volID *int
		if id, ok := volMap[dirRel]; ok {
			v := id
			volID = &v
		}
		text, meta, rerr := ReadTextFile(path)
		if rerr != nil {
			return nil
		}
		body := BodyOf(text)
		if _, err := s.DB.Exec(`INSERT INTO chapters
			(book_id, title, content, summary, order_num, type, volume_id, file_path, word_count, content_hash, file_mtime)
			VALUES (?, ?, '', ?, ?, 'chapter', ?, ?, ?, ?, ?)`,
			bookID, title, ParseFrontMatter(text).Get("summary"), order, volID, relSlash,
			WordCount(body), meta.Hash, meta.MTime); err == nil {
			res.AddedChapters++
		}
		return nil
	})
	if err != nil {
		return res, err
	}

	// 库里有、磁盘没有 → 标 missing（不删）
	rows, err := s.DB.Query(`SELECT id, COALESCE(file_path,'') FROM chapters WHERE book_id=? AND file_path<>''`, bookID)
	if err == nil {
		type miss struct {
			id  int
			rel string
		}
		var list []miss
		for rows.Next() {
			var id int
			var rel string
			if rows.Scan(&id, &rel) == nil && !seenChapters[rel] {
				list = append(list, miss{id, rel})
			}
		}
		rows.Close()
		for _, m := range list {
			s.DB.Exec(`UPDATE chapters SET missing=1 WHERE id=?`, m.id)
			res.MissingChapters++
		}
	}
	return res, nil
}

func (s *Service) chapterExists(bookID int, rel string) bool {
	var n int
	s.DB.QueryRow(`SELECT COUNT(*) FROM chapters WHERE book_id=? AND file_path=?`, bookID, rel).Scan(&n)
	return n > 0
}

func (s *Service) entityExists(kind string, bookID int, rel string) bool {
	tbl := map[string]string{"character": "characters", "entry": "entries", "memo": "memos"}[kind]
	if tbl == "" {
		return false
	}
	var n int
	if kind == "memo" {
		s.DB.QueryRow(`SELECT COUNT(*) FROM memos WHERE book_id=? AND file_path=?`, bookID, rel).Scan(&n)
	} else {
		s.DB.QueryRow(`SELECT COUNT(*) FROM `+tbl+` WHERE book_id=? AND file_path=?`, bookID, rel).Scan(&n)
	}
	return n > 0
}

func (s *Service) docExists(bookID int, rel string) bool {
	var n int
	s.DB.QueryRow(`SELECT COUNT(*) FROM book_docs WHERE book_id=? AND file_path=?`, bookID, rel).Scan(&n)
	return n > 0
}

func (s *Service) insertScannedCharacter(bookID int, name, group, rel string) bool {
	_, err := s.DB.Exec(`INSERT INTO characters (book_id, name, gender, personality, info, folder, folders, file_path)
		VALUES (?, ?, 'unknown', '', '', ?, '[]', ?)`, bookID, name, group, rel)
	return err == nil
}

func (s *Service) insertScannedEntry(bookID int, name, category, rel string) bool {
	_, err := s.DB.Exec(`INSERT INTO entries (book_id, category_id, category_type, name, tags, description, file_path)
		VALUES (?, 0, ?, ?, '[]', '', ?)`, bookID, category, name, rel)
	return err == nil
}

func (s *Service) insertScannedDoc(bookID int, kind, title, rel string) bool {
	word := 0
	if _, body, err := s.ReadEntity(bookID, kind, rel); err == nil {
		word = WordCount(body)
	}
	_, err := s.DB.Exec(`INSERT INTO book_docs (book_id, kind, title, file_path, word_count) VALUES (?, ?, ?, ?, ?)`,
		bookID, kind, title, rel, word)
	return err == nil
}

func (s *Service) insertScannedMemo(bookID int, title, rel string) bool {
	word := 0
	if _, body, err := s.ReadEntity(bookID, "memo", rel); err == nil {
		word = WordCount(body)
	}
	_, err := s.DB.Exec(`INSERT INTO memos (book_id, title, content, category, tags, word_count, file_path)
		VALUES (?, ?, '', '', '[]', ?, ?)`, bookID, title, word, rel)
	return err == nil
}
