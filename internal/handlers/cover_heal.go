package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// cachedImageSource 描述一个"图存在共享缓存目录里"的站内地址族。
//
// 这些目录（data/generated-images、data/editor-images）是**暂存区**，不是作品的
// 最终归属：它们跟 xingmang.db 同级，属于进程级缓存，会被清理、也不会随作品目录
// 一起被备份/搬迁。所以一旦某张图要"成为某个作品的封面"，就必须复制进
// books/<书名>/covers/，让它跟着作品走。
var cachedImageSources = []struct{ dir, prefix string }{
	{generatedImageDir, generatedImageURLPrefix},
	{editorImageDir, editorImageURLPrefix},
}

// bookGalleryURLPrefix 返回某作品画廊的站内地址前缀。
func bookGalleryURLPrefix(bookID int) string {
	return fmt.Sprintf("/api/books/%d/gallery/", bookID)
}

// asLocalMediaURL 把可能的绝对地址（http://host:19825/api/generated-images/x.png）
// 归一成站内路径，让"前缀判断"不受 URL 来源影响。
//
// 前端拿到的 URL 形态不总是纯路径：换过基址、走过代理、或从别的页面复制过来的
// 地址都会带 host。只按字符串前缀匹配的话，这类地址会被误判成"不是本站图片"，
// 于是又被原样存进 books.cover —— 正是这次要修的病根。
func asLocalMediaURL(s string) string {
	if !strings.Contains(s, "://") {
		return s
	}
	u, err := url.Parse(s)
	if err != nil || u.Path == "" {
		return s
	}
	if u.RawQuery != "" {
		return u.Path + "?" + u.RawQuery
	}
	return u.Path
}

// parseBookGalleryURL 解析 /api/books/{id}/gallery/{name} 形态的站内地址。
// 注意别把 /api/books/{id}/gallery/activate 这类动作地址认成图片。
func parseBookGalleryURL(local string) (int, string, bool) {
	const marker = "/gallery/"
	if !strings.HasPrefix(local, "/api/books/") {
		return 0, "", false
	}
	rest := strings.TrimPrefix(local, "/api/books/")
	i := strings.Index(rest, marker)
	if i <= 0 {
		return 0, "", false
	}
	id, err := strconv.Atoi(rest[:i])
	if err != nil || id <= 0 {
		return 0, "", false
	}
	name := rest[i+len(marker):]
	if j := strings.IndexAny(name, "?#"); j >= 0 {
		name = name[:j]
	}
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return 0, "", false
	}
	return id, name, true
}

// isSiteImageURL 判断是不是本程序已经在管的站内图片地址
// （共享缓存目录、正文插图，或某作品的画廊图）。
func isSiteImageURL(u string) bool {
	local := asLocalMediaURL(u)
	if strings.HasPrefix(local, generatedImageURLPrefix) || strings.HasPrefix(local, editorImageURLPrefix) {
		return true
	}
	_, _, ok := parseBookGalleryURL(local)
	return ok
}

// promoteCoverFromLocalURL 把一张站内图片设为当前封面（旧封面自动归档进 gallery/）。
//
// 返回 handled=false 表示这个地址不属于任何"图片来源"（例如已经指向
// /api/books/{id}/cover 这类"当前封面"地址，或站外地址），调用方应按原样处理。
func promoteCoverFromLocalURL(svc *bookstore.Service, bookID int, coverURL string) (bool, error) {
	local := asLocalMediaURL(coverURL)

	// 0) 已经是"当前封面"的地址：前端从画廊点选后会把预览设成它，
	//    再点一次「保存封面」时不该把 URL 又写回 books.cover。
	//    这里认下来但不做任何事，保证 books.cover 始终为空、读取只走文件通道。
	if local == fmt.Sprintf("/api/books/%d/cover", bookID) {
		return true, nil
	}

	// 1) 作品自己画廊里的图（AI 生成图与历史封面都在这里）
	if srcID, name, ok := parseBookGalleryURL(local); ok {
		if srcID != bookID {
			return true, fmt.Errorf("这张图属于另一个作品")
		}
		bookDir, err := svc.BookDir(bookID)
		if err != nil {
			return true, err
		}
		path := filepath.Join(bookDir, "gallery", name)
		data, err := os.ReadFile(path)
		if err != nil {
			return true, fmt.Errorf("画廊里找不到这张图")
		}
		return true, svc.SetCoverBytes(bookID, data, filepath.Ext(path))
	}

	// 2) 共享缓存目录里的图（历史路径 / 没有作品上下文的生成）
	for _, src := range cachedImageSources {
		if !strings.HasPrefix(local, src.prefix) {
			continue
		}
		path, _, err := localMediaPath(src.dir, src.prefix, local)
		if err != nil {
			return true, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return true, err
		}
		return true, svc.SetCoverBytes(bookID, data, filepath.Ext(path))
	}
	return false, nil
}

// HealLegacyCovers 把"住在缓存目录里的封面"搬进作品文件夹。
//
// 背景：早期版本的 UpdateBook 只认 data: 开头的封面。而 AI 生成封面返回的是
// 站内 URL（/api/generated-images/xxx.png），于是这个地址被当成普通字符串直接
// 写进了 books.cover —— 封面数据其实还躺在共享缓存目录里，既不在
// books/<书名>/covers/，换封面时也没有"上一张文件"可以归档，gallery/ 永远是空的。
// 表现就是：作品列表有封面，但打开封面弹窗画廊里一张都没有；缓存目录一旦被清理，封面直接裂图。
//
// 这里做一次幂等搬迁：
//   - data: URL        → SetCover（解码后落到 covers/）
//   - 站内图片 URL      → 读原文件 → SetCoverBytes（旧封面自动归档进 gallery/）
//   - 其它形态          → 不动（已经是文件通道地址，或站外地址）
//
// 搬迁成功后 books.cover 会被置空，读取统一走 /api/books/{id}/cover 文件通道。
//
// ⚠ 源文件已经不存在时**不**清空 cover：缓存目录可能被清理过。宁可保留一个还能
// 解释的旧值，也不要凭空把用户的封面抹成占位图。
func HealLegacyCovers(db *sql.DB) {
	rows, err := db.Query(`SELECT id, COALESCE(cover,'') FROM books WHERE TRIM(COALESCE(cover,'')) <> ''`)
	if err != nil {
		log.Printf("封面迁移：查询失败，跳过: %v", err)
		return
	}
	type legacyCover struct {
		id    int
		cover string
	}
	var list []legacyCover
	for rows.Next() {
		var r legacyCover
		if err := rows.Scan(&r.id, &r.cover); err == nil {
			list = append(list, r)
		}
	}
	_ = rows.Close()
	if len(list) == 0 {
		return
	}

	svc := bookstore.NewService(db)
	moved := 0
	for _, r := range list {
		cover := strings.TrimSpace(r.cover)
		if cover == "" {
			continue
		}

		var err error
		handled := false
		switch {
		case strings.HasPrefix(cover, "data:"):
			handled, err = true, svc.SetCover(r.id, cover)
		default:
			handled, err = promoteCoverFromLocalURL(svc, r.id, cover)
		}
		if !handled {
			continue // 没有可搬的东西
		}
		if err != nil {
			log.Printf("封面迁移：书 %d 未搬迁（%v）", r.id, err)
			continue
		}
		moved++
	}
	if moved > 0 {
		log.Printf("封面迁移：%d 张历史封面已从缓存目录搬进作品文件夹", moved)
	}
}

// MigrateGeneratedImages 把共享缓存目录里的 AI 生成图按归属搬进作品画廊。
//
// 背景：早期"每次生成"都统一落在 data/generated-images（与 xingmang.db 同级的
// 进程级缓存目录）。这些图其实只服务于某一次封面生成，留在缓存里等于
// "作品带不走自己的素材"：备份作品拿不到图，删作品也清不掉。
//
// 归属判定用**内容指纹**，不靠文件名或时间（文件名里没有作品信息，时间也不可靠）：
// 与某作品当前封面（covers/cover.*）或画廊（gallery/*）内容相同的，就归该作品的 gallery/。
//
// 认不出归属的**原地不动**（只记一行日志）：宁可让用户在缓存目录里看到一张图，
// 也不要把它塞进某个作品 —— 塞错了比不搬更难收拾。
// 幂等：搬完缓存目录就空了，下次启动无事可做。
func MigrateGeneratedImages(db *sql.DB) {
	entries, err := os.ReadDir(generatedImageDir)
	if err != nil {
		return // 目录不存在 = 没有可搬的
	}
	svc := bookstore.NewService(db)

	// 先建"内容指纹 → 作品 id"索引：拿封面文件与画廊文件当归属证据
	owner := map[string]int{}
	rows, err := db.Query(`SELECT id FROM books WHERE TRIM(COALESCE(folder_path,'')) <> ''`)
	if err != nil {
		log.Printf("生成图迁移：查询作品失败，跳过: %v", err)
		return
	}
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	_ = rows.Close()

	for _, id := range ids {
		bookDir, err := svc.BookDir(id)
		if err != nil {
			continue
		}
		if p, _, err := svc.CoverFile(id); err == nil {
			if h, err := fileHash(p); err == nil {
				owner[h] = id
			}
		}
		items, err := svc.ListGallery(id)
		if err != nil {
			continue
		}
		for _, it := range items {
			if h, err := fileHash(filepath.Join(bookDir, "gallery", it.Name)); err == nil {
				owner[h] = id
			}
		}
	}

	moved, unknown := 0, 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		src := filepath.Join(generatedImageDir, e.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			continue
		}
		id, ok := owner[contentHash(data)]
		if !ok {
			unknown++
			continue
		}
		bookDir, err := svc.BookDir(id)
		if err != nil {
			log.Printf("生成图迁移：%s 找不到归属作品的目录（%v）", e.Name(), err)
			continue
		}
		if err := moveFile(src, filepath.Join(bookDir, "gallery", e.Name())); err != nil {
			log.Printf("生成图迁移：%s 搬入作品 %d 失败（%v）", e.Name(), id, err)
			continue
		}
		moved++
	}
	if moved > 0 {
		log.Printf("生成图迁移：%d 张已从缓存目录搬进作品 gallery/", moved)
	}
	if unknown > 0 {
		log.Printf("生成图迁移：%d 张认不出归属，保留在 %s/", unknown, generatedImageDir)
	}
}

// contentHash 与 bookstore 内部 hash8 同口径（sha256 前 8 位十六进制）。
func contentHash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:8]
}

func fileHash(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return contentHash(data), nil
}

// moveFile 移动文件；跨盘（作品库可能在别的盘）时退化为复制 + 删除。
// 目标已存在视为成功（幂等：重复迁移不报错、不覆盖）。
func moveFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(dst); err == nil {
		return os.Remove(src)
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	return os.Remove(src)
}
