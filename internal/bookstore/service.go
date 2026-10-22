package bookstore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/appdir"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ErrFileConflict 表示磁盘文件在本次编辑期间被外部改过。
//
// 保存时必须让用户选择"载入磁盘版 / 保留当前"，不能静默覆盖 ——
// 这是"文件为真相"模式下唯一不可省略的防线。
var ErrFileConflict = errors.New("file_conflict")

// Service 把"DB 索引 + 磁盘文件"合成一个入口。
//
// 所有会改正文/人物/条目的代码（HTTP handler、Agent 工具、回滚）都必须走这里，
// 否则会出现"库改了、文件没改"的裂缝 —— 而文件是唯一真相，下次读取就会把改动吞掉。
type Service struct {
	DB *sql.DB
}

func NewService(db *sql.DB) *Service { return &Service{DB: db} }

// 书根下的固定子目录。空目录也建好，用户打开资源管理器就能看到结构。
// memory 是记忆层（对话/Agent/图谱/拆书/生成历史）的根，见 docs/books-memory-storage.md。
var bookSubdirs = []string{"covers", "gallery", "characters", "world", "outlines", "ideas", "memos", "assets", memoryRootDir}

// bookSubSubdirs 是"子目录里还要再建一层"的固定结构。
var bookSubSubdirs = map[string][]string{
	memoryRootDir: MemorySubdirs,
}

// ensureBookDirs 建齐书目录下的固定结构（含 memory/ 的五个子目录）。
func ensureBookDirs(dir string) error {
	for _, sub := range bookSubdirs {
		if err := ensureDir(filepath.Join(dir, sub)); err != nil {
			return fmt.Errorf("创建子目录 %s 失败: %w", sub, err)
		}
		for _, inner := range bookSubSubdirs[sub] {
			if err := ensureDir(filepath.Join(dir, sub, inner)); err != nil {
				return fmt.Errorf("创建子目录 %s/%s 失败: %w", sub, inner, err)
			}
		}
	}
	return nil
}

// ensureDir 是 MkdirAll 的统一入口（本项目只建目录、不建文件）。
func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// Root 返回作品库根目录（绝对路径），并确保它存在。
//
// 优先级：app_settings.books_root → <运行目录>/books。
func (s *Service) Root() (string, error) {
	var v string
	err := s.DB.QueryRow(`SELECT value FROM app_settings WHERE key='books_root'`).Scan(&v)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	root := strings.TrimSpace(v)
	if root == "" {
		root = filepath.Join(appdir.Root(), "books")
	}
	if !filepath.IsAbs(root) {
		root = filepath.Join(appdir.Root(), root)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("创建作品库目录失败: %w", err)
	}
	return root, nil
}

// BookDir 返回某本书的绝对目录。
func (s *Service) BookDir(bookID int) (string, error) {
	var rel string
	if err := s.DB.QueryRow(`SELECT folder_path FROM books WHERE id=?`, bookID).Scan(&rel); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("书籍不存在")
		}
		return "", err
	}
	if strings.TrimSpace(rel) == "" {
		return "", fmt.Errorf("书籍目录未初始化（folder_path 为空）")
	}
	root, err := s.Root()
	if err != nil {
		return "", err
	}
	return JoinWithin(root, rel)
}

// CreateBook 建目录 + 写 book.json + 插库，三者要么都成功，要么都回滚。
func (s *Service) CreateBook(b *models.Book) error {
	root, err := s.Root()
	if err != nil {
		return err
	}
	dirName := uniqueDirName(root, SanitizeName(b.Title))
	dir := filepath.Join(root, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建书目录失败: %w", err)
	}
	if err := ensureBookDirs(dir); err != nil {
		os.RemoveAll(dir)
		return err
	}

	res, err := s.DB.Exec(`INSERT INTO books
		(title, description, cover, author, category, tags, status, is_public, target_words, folder_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.Title, b.Description, b.Cover, b.Author, b.Category, b.Tags.ToDBValue(), b.Status, b.IsPublic, b.TargetWords, dirName)
	if err != nil {
		os.RemoveAll(dir)
		return err
	}
	id, _ := res.LastInsertId()
	b.ID = int(id)
	b.FolderPath = dirName

	meta := bookJSON{
		Schema:    1,
		ID:        int(id),
		Title:     b.Title,
		Author:    b.Author,
		Status:    b.Status,
		Tags:      b.Tags,
		Target:    b.TargetWords,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	if err := writeBookJSON(filepath.Join(dir, "book.json"), meta); err != nil {
		s.DB.Exec(`DELETE FROM books WHERE id=?`, id)
		os.RemoveAll(dir)
		return fmt.Errorf("写入 book.json 失败: %w", err)
	}
	return nil
}

type bookJSON struct {
	Schema    int               `json:"schema"`
	ID        int               `json:"id"`
	Title     string            `json:"title"`
	Author    string            `json:"author"`
	Status    string            `json:"status"`
	Tags      models.StringList `json:"tags"`
	Target    int               `json:"target_words"`
	CreatedAt string            `json:"created_at"`
}

func writeBookJSON(path string, meta bookJSON) error {
	raw, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(path, append(raw, '\n'))
}

// RenameBook 同步改书名与目录名。
//
// 顺序固定：先动磁盘，成功后才改库。反过来一旦 rename 失败，
// 就会出现"库里新名、磁盘旧名"的裂缝，比失败本身更难修。
func (s *Service) RenameBook(bookID int, newTitle string) (string, error) {
	var oldTitle, oldRel string
	if err := s.DB.QueryRow(`SELECT title, folder_path FROM books WHERE id=?`, bookID).Scan(&oldTitle, &oldRel); err != nil {
		return "", err
	}
	root, err := s.Root()
	if err != nil {
		return "", err
	}
	newRel := uniqueDirName(root, SanitizeName(newTitle))
	if oldRel != "" && newRel == oldRel {
		newRel = oldRel
	} else if oldRel != "" {
		oldDir, err := JoinWithin(root, oldRel)
		if err != nil {
			return "", err
		}
		newDir := filepath.Join(root, newRel)
		if err := renameWithRetry(oldDir, newDir); err != nil {
			return "", fmt.Errorf("重命名目录失败（可能被其他程序占用）: %w", err)
		}
	}
	if _, err := s.DB.Exec(`UPDATE books SET title=?, folder_path=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		newTitle, newRel, bookID); err != nil {
		return "", err
	}
	return newRel, nil
}

func renameWithRetry(oldPath, newPath string) error {
	delays := []time.Duration{0, 50 * time.Millisecond, 200 * time.Millisecond}
	var lastErr error
	for _, d := range delays {
		if d > 0 {
			time.Sleep(d)
		}
		if err := os.Rename(oldPath, newPath); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	return lastErr
}

func uniqueDirName(root, base string) string {
	if base == "" {
		base = "未命名"
	}
	candidate := base
	for i := 2; i < 1000; i++ {
		if _, err := os.Lstat(filepath.Join(root, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s (%d)", base, i)
	}
	return base
}

// nextOrder 返回某个容器（书根或某卷）里的下一个章节序号。
func (s *Service) nextOrder(bookID int, volumeID *int) int {
	vid := 0
	if volumeID != nil {
		vid = *volumeID
	}
	var maxOrder int
	s.DB.QueryRow(`SELECT COALESCE(MAX(order_num),0) FROM chapters
		WHERE book_id=? AND COALESCE(volume_id,0)=?`, bookID, vid).Scan(&maxOrder)
	return maxOrder + 1
}

// CreateChapter 写文件 + 插库。文件写失败则整条操作失败，不产生"只有库没有文件"的章节。
func (s *Service) CreateChapter(c *models.Chapter) error {
	dir, prefix, err := s.chapterDir(c.BookID, c.VolumeID)
	if err != nil {
		return err
	}
	order := c.OrderNum
	if order <= 0 {
		order = s.nextOrder(c.BookID, c.VolumeID)
	}
	name := UniqueName(dir, ChapterFileName(order, c.Title, ".md"))
	text := composeChapterText("", c.Content, c.Summary)
	meta, err := WriteTextFile(filepath.Join(dir, name), text)
	if err != nil {
		return err
	}
	rel := prefix + name

	res, err := s.DB.Exec(`INSERT INTO chapters
		(book_id, title, content, summary, order_num, type, volume_id, file_path, word_count, content_hash, file_mtime)
		VALUES (?, ?, '', ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.BookID, c.Title, c.Summary, order, defaultType(c.Type), c.VolumeID,
		rel, WordCount(c.Content), meta.Hash, meta.MTime)
	if err != nil {
		os.Remove(filepath.Join(dir, name))
		return err
	}
	id, _ := res.LastInsertId()
	c.ID = int(id)
	c.OrderNum = order
	c.FilePath = rel
	return nil
}

// UpdateChapter 覆盖写正文；标题变化或换卷时移动文件。
//
// 正文写入走 front-matter 补丁：文件里用户加的自定义字段不会被抹掉。
func (s *Service) UpdateChapter(c *models.Chapter) error {
	var bookID, order int
	var oldTitle, oldRel, summary, oldHash string
	var oldVolumeID sql.NullInt64
	err := s.DB.QueryRow(`SELECT book_id, title, order_num, COALESCE(summary,''), COALESCE(file_path,''), COALESCE(content_hash,''), volume_id
		FROM chapters WHERE id=?`, c.ID).Scan(&bookID, &oldTitle, &order, &summary, &oldRel, &oldHash, &oldVolumeID)
	if err != nil {
		return err
	}

	// 冲突检测：索引里记的 hash 是"上次读/写时的内容"。
	// 现在磁盘上的 hash 与它不同，说明这期间有外部编辑器改过文件。
	if oldRel != "" && oldHash != "" {
		var oldDirForCheck string
		var oldVolForCheck *int
		if oldVolumeID.Valid {
			v := int(oldVolumeID.Int64)
			oldVolForCheck = &v
		}
		if d, _, derr := s.chapterDir(bookID, oldVolForCheck); derr == nil {
			oldDirForCheck = d
		}
		if oldDirForCheck != "" {
			if raw, rerr := os.ReadFile(filepath.Join(oldDirForCheck, filepath.FromSlash(oldRel))); rerr == nil {
				if HashText(string(raw)) != oldHash {
					return ErrFileConflict
				}
			}
		}
	}

	// 换卷：落到目标容器末尾（不重排源容器，允许序号空洞）
	newVol := 0
	if c.VolumeID != nil {
		newVol = *c.VolumeID
	}
	if newVol != int(oldVolumeID.Int64) || (!oldVolumeID.Valid && newVol != 0) {
		order = s.nextOrder(bookID, c.VolumeID)
	}
	if c.OrderNum > 0 {
		order = c.OrderNum
	}

	dir, prefix, err := s.chapterDir(bookID, c.VolumeID)
	if err != nil {
		return err
	}
	ext := filepath.Ext(oldRel)
	if ext == "" {
		ext = ".md"
	}
	desired := ChapterFileName(order, c.Title, ext)
	newRel := prefix + desired

	// 旧文件可能在上一个卷目录里：必须用旧 volume 解析旧目录，否则换卷时找不到原文
	// （会丢掉 front-matter 自定义字段，旧文件也删不掉）。
	oldFull := ""
	if oldRel != "" {
		var oldVolPtr *int
		if oldVolumeID.Valid {
			v := int(oldVolumeID.Int64)
			oldVolPtr = &v
		}
		oldDir, _, derr := s.chapterDir(bookID, oldVolPtr)
		if derr != nil {
			return derr
		}
		oldFull = filepath.Join(oldDir, filepath.FromSlash(oldRel))
	}
	inPlace := oldRel == newRel && oldRel != ""
	if !inPlace {
		// 目标目录里可能已有同名文件；重名时加后缀（同目录原地写不走这里）
		name := UniqueName(dir, desired)
		newRel = prefix + name
	}

	existing := ""
	if oldFull != "" {
		if raw, err := os.ReadFile(oldFull); err == nil {
			existing = string(raw)
		}
	}
	text := composeChapterText(existing, c.Content, c.Summary)
	meta, err := WriteTextFile(filepath.Join(dir, filepath.FromSlash(strings.TrimPrefix(newRel, prefix))), text)
	if err != nil {
		return err
	}
	if oldFull != "" && filepath.FromSlash(oldRel) != filepath.FromSlash(newRel) {
		_ = os.Remove(oldFull)
	}

	if _, err := s.DB.Exec(`UPDATE chapters SET
		title=?, content='', summary=?, order_num=?, type=?, volume_id=?,
		file_path=?, word_count=?, content_hash=?, file_mtime=?, missing=0, updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		c.Title, c.Summary, order, defaultType(c.Type), c.VolumeID,
		newRel, WordCount(c.Content), meta.Hash, meta.MTime, c.ID); err != nil {
		return err
	}
	c.BookID = bookID
	c.OrderNum = order
	c.FilePath = newRel
	return nil
}

// composeChapterText 把正文与摘要合成一个文件内容。
//
// 从已有文件出发（而不是从零拼），是为了保留 front-matter 里用户自己加的字段。
func composeChapterText(existing, content, summary string) string {
	fm := ParseFrontMatter(existing)
	fm.Body = NormalizeText(content)
	fm = fm.WithField("summary", summary)
	return fm.Render()
}

// ChapterText 读取章节正文（去掉 front-matter，编辑器只该看到正文）。
func (s *Service) ChapterText(chapterID int) (string, error) {
	var bookID int
	var rel, content, summary string
	err := s.DB.QueryRow(`SELECT book_id, COALESCE(file_path,''), COALESCE(content,''), COALESCE(summary,'')
		FROM chapters WHERE id=?`, chapterID).Scan(&bookID, &rel, &content, &summary)
	if err != nil {
		return "", err
	}
	if rel == "" {
		// 存量/异常行：还没有文件，退回库里的正文（只读兼容，不会写回）
		return content, nil
	}
	dir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	text, meta, err := ReadTextFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		return "", err
	}
	fm := ParseFrontMatter(text)
	// 文件里的 summary 是权威；与库不一致时顺手同步，避免"两边各说一套"
	if fmSummary := fm.Get("summary"); fmSummary != summary {
		s.DB.Exec(`UPDATE chapters SET summary=? WHERE id=?`, fmSummary, chapterID)
	}
	s.DB.Exec(`UPDATE chapters SET word_count=?, content_hash=?, file_mtime=?, missing=0 WHERE id=?`,
		WordCount(fm.Body), meta.Hash, meta.MTime, chapterID)
	return fm.Body, nil
}

// SetChapterText 只改正文（Agent 追加/替换/回滚走这里），保留摘要与自定义字段。
func (s *Service) SetChapterText(chapterID int, content string) error {
	var bookID, order int
	var title, rel, summary, typ string
	var volumeID sql.NullInt64
	err := s.DB.QueryRow(`SELECT book_id, title, order_num, COALESCE(file_path,''), COALESCE(summary,''), COALESCE(type,'chapter'), volume_id
		FROM chapters WHERE id=?`, chapterID).Scan(&bookID, &title, &order, &rel, &summary, &typ, &volumeID)
	if err != nil {
		return err
	}
	dir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	existing := ""
	if rel != "" {
		if raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			existing = string(raw)
		}
	}
	text := composeChapterText(existing, content, summary)
	target := filepath.Join(dir, filepath.FromSlash(rel))
	if rel == "" {
		// 异常行（没有文件）：补一个，避免正文只存在内存里
		rel = UniqueName(dir, ChapterFileName(order, title, ".md"))
		target = filepath.Join(dir, rel)
	}
	meta, err := WriteTextFile(target, text)
	if err != nil {
		return err
	}
	_, err = s.DB.Exec(`UPDATE chapters SET content='', file_path=?, word_count=?, content_hash=?, file_mtime=?,
		missing=0, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		rel, WordCount(content), meta.Hash, meta.MTime, chapterID)
	return err
}

// ArchiveBook 把整本书的目录移到 <root>/_归档/<书名>_<时间>/。
//
// 删除作品时的默认动作：文件还在，只是移出书架；比直接删目录安全得多。
func (s *Service) ArchiveBook(bookID int) error {
	var rel, title string
	if err := s.DB.QueryRow(`SELECT COALESCE(folder_path,''), title FROM books WHERE id=?`, bookID).Scan(&rel, &title); err != nil {
		return err
	}
	if rel == "" {
		return nil
	}
	root, err := s.Root()
	if err != nil {
		return err
	}
	src, err := JoinWithin(root, rel)
	if err != nil {
		return err
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	archiveRoot := filepath.Join(root, "_归档")
	if err := os.MkdirAll(archiveRoot, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(archiveRoot, fmt.Sprintf("%s_%s", SanitizeName(title), time.Now().Format("20060102-150405")))
	return renameWithRetry(src, dst)
}

// TrashChapterFile 把章节文件移到 <root>/_trash/<book>/<时间>/ 下。
//
// 先做"移动到回收目录"，等 F2 再接系统回收站；无论哪种，都不能直接 os.Remove——
// 那样误删不可恢复。
// TrashChapterFile 把章节文件移到回收站（章节与记忆共用同一套落点）。
func (s *Service) TrashChapterFile(bookID int, rel string) error {
	return s.trashBookFile(bookID, rel)
}

// trashBookFile 把书目录内的某个文件移入系统回收站；不可用时回退到库内 _trash/<书>/<时间>/。
func (s *Service) trashBookFile(bookID int, rel string) error {
	if strings.TrimSpace(rel) == "" {
		return nil
	}
	dir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	src := filepath.Join(dir, filepath.FromSlash(rel))
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	// 优先系统回收站（用户可还原）；不可用时才退到库内 _trash。
	if err := moveToRecycleBin(src); err == nil {
		return nil
	}
	root, err := s.Root()
	if err != nil {
		return err
	}
	var bookRel string
	s.DB.QueryRow(`SELECT folder_path FROM books WHERE id=?`, bookID).Scan(&bookRel)
	trashDir := filepath.Join(root, "_trash", SanitizeName(bookRel), time.Now().Format("20060102-150405"))
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return err
	}
	return os.Rename(src, filepath.Join(trashDir, filepath.Base(src)))
}

func defaultType(t string) string {
	if strings.TrimSpace(t) == "" {
		return "chapter"
	}
	return t
}

// SwitchRoot 切换作品库根目录，并把现有内容一起搬过去。
//
// 只改指针不搬文件的话，所有 folder_path 会瞬间指向不存在的目录 ——
// 书架看起来"作品全没了"，比切换失败更糟。所以这里做 move，失败整体回滚。
func (s *Service) SwitchRoot(newRoot string) error {
	newRoot = strings.TrimSpace(newRoot)
	if newRoot == "" {
		return fmt.Errorf("目录不能为空")
	}
	abs, err := filepath.Abs(newRoot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	// 可写性检查：建一个临时文件再删，避免切过去才发现没权限
	probe := filepath.Join(abs, ".xm-write-test")
	if err := os.WriteFile(probe, []byte("ok"), 0o644); err != nil {
		return fmt.Errorf("目标目录不可写: %w", err)
	}
	os.Remove(probe)

	oldRoot, err := s.Root()
	if err != nil {
		return err
	}
	if filepath.Clean(oldRoot) == filepath.Clean(abs) {
		return nil
	}

	entries, err := os.ReadDir(oldRoot)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	var moved []string
	rollback := func() {
		for i := len(moved) - 1; i >= 0; i-- {
			_ = moveDir(filepath.Join(abs, moved[i]), filepath.Join(oldRoot, moved[i]))
		}
	}
	for _, e := range entries {
		name := e.Name()
		if name == "_trash" || strings.HasPrefix(name, ".") {
			continue
		}
		src := filepath.Join(oldRoot, name)
		dst := filepath.Join(abs, name)
		if _, err := os.Lstat(dst); err == nil {
			rollback()
			return fmt.Errorf("目标目录里已存在同名的 %q，请换一个目录或先处理冲突", name)
		}
		if err := moveDir(src, dst); err != nil {
			rollback()
			return fmt.Errorf("移动 %q 失败: %w", name, err)
		}
		moved = append(moved, name)
	}

	if _, err := s.DB.Exec(`INSERT INTO app_settings(key, value, updated_at) VALUES('books_root', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`, abs); err != nil {
		rollback()
		return err
	}
	return nil
}

// moveDir 先尝试 rename（同盘瞬间完成），失败再复制+删除（跨盘）。
func moveDir(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyTree(src, dst); err != nil {
		os.RemoveAll(dst)
		return err
	}
	return os.RemoveAll(src)
}
