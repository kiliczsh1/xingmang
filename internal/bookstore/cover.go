package bookstore

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// GalleryItem 是封面画廊里的一张图。
type GalleryItem struct {
	Name    string `json:"name"`
	MTime   int64  `json:"mtime"`
	Current bool   `json:"current"`
}

var coverExts = []string{".png", ".jpg", ".jpeg", ".webp", ".gif"}

func coverExtFromMime(mime string) string {
	switch strings.ToLower(strings.TrimSpace(mime)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/png"
	}
}

// decodeDataURL 接受 data URL 或裸 base64。
func decodeDataURL(s string) ([]byte, string, error) {
	mime := "image/png"
	if strings.HasPrefix(s, "data:") {
		comma := strings.Index(s, ",")
		if comma < 0 {
			return nil, "", fmt.Errorf("图片数据格式错误")
		}
		meta := s[len("data:"):comma]
		if i := strings.Index(meta, ";"); i >= 0 {
			mime = meta[:i]
		}
		s = s[comma+1:]
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, "", fmt.Errorf("图片 base64 解码失败: %w", err)
	}
	return raw, mime, nil
}

func hash8(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:8]
}

// findCover 返回当前封面文件路径（没有则空）。
func findCover(bookDir string) string {
	for _, ext := range coverExts {
		p := filepath.Join(bookDir, "covers", "cover"+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// SetCover 写入当前封面（data URL）；旧封面先归档到 gallery/，不覆盖丢失。
//
// 内容相同则直接返回（避免每次保存都往画廊堆一份）。
func (s *Service) SetCover(bookID int, dataURL string) error {
	raw, mime, err := decodeDataURL(dataURL)
	if err != nil {
		return err
	}
	return s.SetCoverBytes(bookID, raw, coverExtFromMime(mime))
}

// SetCoverBytes 把已经解码好的图片设为当前封面。
//
// 单独开一个入口，是因为封面来源不止"前端上传的 data URL"一种：
// AI 生成的图已经落在本地文件里（返回的是站内 URL），那条路径也得走**同一套**
// "旧图归档到 gallery/ → 写新图 → 清空 books.cover"的逻辑。
// 之前只认 data:，于是生成图的 URL 被当成字符串存进了 books.cover，
// 旧封面既没进画廊也没被替换 —— 这就是"换封面后画廊里少一张"的根因。
func (s *Service) SetCoverBytes(bookID int, raw []byte, ext string) error {
	if len(raw) == 0 {
		return fmt.Errorf("图片内容为空")
	}
	ext = strings.ToLower(ext)
	if !isCoverExt(ext) {
		ext = ".png"
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	newHash := hash8(raw)

	if cur := findCover(bookDir); cur != "" {
		if old, err := os.ReadFile(cur); err == nil {
			if hash8(old) == newHash {
				return nil // 同一张图，不重复存
			}
			if err := s.archiveCover(bookDir, cur, old); err != nil {
				return err
			}
			_ = os.Remove(cur)
		}
	}
	// 旧封面可能是别的扩展名：一并清掉，避免同时存在两张"当前封面"
	for _, e := range coverExts {
		if e != ext {
			_ = os.Remove(filepath.Join(bookDir, "covers", "cover"+e))
		}
	}

	if err := os.MkdirAll(filepath.Join(bookDir, "covers"), 0o755); err != nil {
		return err
	}
	if err := WriteFileAtomic(filepath.Join(bookDir, "covers", "cover"+ext), raw); err != nil {
		return err
	}
	// 数据库不再存 base64，只留空；读取走 CoverFile。
	_, err = s.DB.Exec(`UPDATE books SET cover='', updated_at=CURRENT_TIMESTAMP WHERE id=?`, bookID)
	return err
}

// galleryFindByContent 在画廊里找内容完全相同的那张图，返回文件名（没有则空串）。
//
// ⚠ 只能按内容判重，不能按文件名后缀：迁移进来的历史生成图保留的是缓存目录的
// 命名（<纳秒时间戳>_<随机hex>.png），名字里**没有内容指纹**，
// 靠 `-<hash8>` 后缀匹配会漏，同一个文件就会在画廊里出现两遍。
func galleryFindByContent(dir, wantHash string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if e.IsDir() || !isCoverExt(strings.ToLower(filepath.Ext(e.Name()))) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if hash8(data) == wantHash {
			return e.Name()
		}
	}
	return ""
}

// archiveCover 把一张旧封面放进画廊，按时间戳+内容指纹命名。
//
// 画廊里已经有同一张（内容相同）就不再堆一份：AI 生成图本来就先进画廊，
// 再被设为封面时若又归档一次，同一个文件会在画廊里出现两遍。
func (s *Service) archiveCover(bookDir, path string, raw []byte) error {
	gallery := filepath.Join(bookDir, "gallery")
	if err := os.MkdirAll(gallery, 0o755); err != nil {
		return err
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		ext = ".png"
	}
	h := hash8(raw)
	if galleryFindByContent(gallery, h) != "" {
		return nil // 已存在同一张
	}
	name := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102-150405"), h, ext)
	dst := filepath.Join(gallery, name)
	if _, err := os.Stat(dst); err == nil {
		return nil // 已存在同一张
	}
	return os.WriteFile(dst, raw, 0o644)
}

// CoverFile 返回当前封面的绝对路径与 MIME。
func (s *Service) CoverFile(bookID int) (string, string, error) {
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", "", err
	}
	p := findCover(bookDir)
	if p == "" {
		return "", "", os.ErrNotExist
	}
	return p, mimeFromExt(filepath.Ext(p)), nil
}

// ListGallery 返回画廊文件（按文件名倒序，即最新在前）。
func (s *Service) ListGallery(bookID int) ([]GalleryItem, error) {
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return nil, err
	}
	cur := filepath.Base(findCover(bookDir))
	return listImageDir(filepath.Join(bookDir, "gallery"), cur)
}

// listImageDir 列出一个图片目录里的图片文件（按文件名倒序）。
// current 用于标记"当前封面"（只有 gallery 需要）。
func listImageDir(dir, current string) ([]GalleryItem, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []GalleryItem{}, nil
		}
		return nil, err
	}
	out := []GalleryItem{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !isCoverExt(strings.ToLower(filepath.Ext(e.Name()))) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, GalleryItem{Name: e.Name(), MTime: info.ModTime().Unix(), Current: e.Name() == current})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

func isCoverExt(ext string) bool {
	for _, e := range coverExts {
		if e == ext {
			return true
		}
	}
	return false
}

// ActivateGallery 把画廊里的一张设为当前封面，当前这张回到画廊（交换）。
func (s *Service) ActivateGallery(bookID int, name string) error {
	if strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("非法的画廊文件名")
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	src := filepath.Join(bookDir, "gallery", name)
	raw, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("画廊里找不到这张图")
	}
	if cur := findCover(bookDir); cur != "" {
		if old, err := os.ReadFile(cur); err == nil {
			if err := s.archiveCover(bookDir, cur, old); err != nil {
				return err
			}
		}
	}
	ext := strings.ToLower(filepath.Ext(name))
	if !isCoverExt(ext) {
		ext = ".png"
	}
	if err := WriteFileAtomic(filepath.Join(bookDir, "covers", "cover"+ext), raw); err != nil {
		return err
	}
	_ = os.Remove(src)
	// 旧封面可能是别的扩展名，避免同时存在两张"当前封面"
	for _, e := range coverExts {
		if e != ext {
			_ = os.Remove(filepath.Join(bookDir, "covers", "cover"+e))
		}
	}
	_, err = s.DB.Exec(`UPDATE books SET cover='', updated_at=CURRENT_TIMESTAMP WHERE id=?`, bookID)
	return err
}

// DeleteGallery 删除画廊里的单张（进系统回收站，可还原）。
func (s *Service) DeleteGallery(bookID int, name string) error {
	if strings.ContainsAny(name, `/\`) || name != filepath.Base(name) {
		return fmt.Errorf("非法的画廊文件名")
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return err
	}
	src := filepath.Join(bookDir, "gallery", name)
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("画廊里找不到这张图")
	}
	if err := moveToRecycleBin(src); err == nil {
		return nil
	}
	trash := filepath.Join(bookDir, "_trash")
	if err := os.MkdirAll(trash, 0o755); err != nil {
		return err
	}
	return os.Rename(src, filepath.Join(trash, name))
}

// TrashBook 把整本书目录送进系统回收站；不可用时回退到库内 _trash。
func (s *Service) TrashBook(bookID int) error {
	rel, title, err := s.bookRel(bookID)
	if err != nil {
		return err
	}
	if rel == "" {
		return nil
	}
	root, err := s.Root()
	if err != nil {
		return err
	}
	dir, err := JoinWithin(root, rel)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	if err := moveToRecycleBin(dir); err == nil {
		return nil
	}
	trash := filepath.Join(root, "_trash")
	if err := os.MkdirAll(trash, 0o755); err != nil {
		return err
	}
	dst := filepath.Join(trash, fmt.Sprintf("%s_%s", SanitizeName(title), time.Now().Format("20060102-150405")))
	return renameWithRetry(dir, dst)
}

func (s *Service) bookRel(bookID int) (string, string, error) {
	var rel, title string
	err := s.DB.QueryRow(`SELECT COALESCE(folder_path,''), title FROM books WHERE id=?`, bookID).Scan(&rel, &title)
	return rel, title, err
}

// AddGallery 把一张图直接放进画廊（不改变当前封面）。
//
// 与 SetCover 的区别：SetCover 是"设为封面 + 旧图归档"，这里只是往候选池里加一张。
func (s *Service) AddGallery(bookID int, dataURL string) (string, error) {
	raw, mime, err := decodeDataURL(dataURL)
	if err != nil {
		return "", err
	}
	return s.AddGalleryBytes(bookID, raw, coverExtFromMime(mime))
}

// AddGalleryBytes 把已经解码好的图放进画廊，返回文件名。
//
// AI 生成的原图走的就是这条路：生成图**直接进 gallery/**，和"被顶掉的封面"共用一个池子
// —— 前端画廊里一起展示，用户在一处挑图就够了。
//
// 命名与 archiveCover 保持一致（<时间戳>-<内容指纹><扩展名>），
// 这样画廊里新旧图按文件名排序就是按时间排序，不会出现"生成图全排在归档图后面"。
// 同一张图不重复入库（重复点生成拿到同样的字节时不会堆一串）。
func (s *Service) AddGalleryBytes(bookID int, raw []byte, ext string) (string, error) {
	if len(raw) == 0 {
		return "", fmt.Errorf("图片内容为空")
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	gallery := filepath.Join(bookDir, "gallery")
	if err := os.MkdirAll(gallery, 0o755); err != nil {
		return "", err
	}
	ext = strings.ToLower(ext)
	if !isCoverExt(ext) {
		ext = ".png"
	}
	h := hash8(raw)
	if existing := galleryFindByContent(gallery, h); existing != "" {
		return existing, nil // 池子里已有同一张，直接复用它的文件名
	}
	name := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102-150405"), h, ext)
	dst := filepath.Join(gallery, name)
	if _, err := os.Stat(dst); err == nil {
		return name, nil
	}
	if err := WriteFileAtomic(dst, raw); err != nil {
		return "", err
	}
	return name, nil
}
