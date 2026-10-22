package handlers

// 在线章节本地缓存：AI 拆书弹窗「关联章节」用。
//
// 机制：以 数据源/书号 为粒度把整份聚合结果（章节目录 + 正文）落盘到
// chapter-cache/<source>/<bookId>_<limit>_<content>.json，24 小时内命中
// 直接回缓存、不再抓站点；过期或抓取失败不影响原逻辑。
//
// 正文抓取（10 章串行 HTTP）是弹窗打开时的耗时大头，这是缓存的价值点；
// 番茄/七猫/后续任何实现 ChapterLister 的源都自动覆盖，前端无感。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	chapterCacheDir  = "chapter-cache"
	chapterCacheTTL  = 24 * time.Hour
	chapterCachePerm = 0o755
	chapterFilePerm  = 0o644
)

// chapterCacheEnvelope 是缓存文件结构：data 原样存接口响应体。
type chapterCacheEnvelope struct {
	CachedAt time.Time              `json:"cachedAt"`
	Source   string                 `json:"source"`
	BookID   string                 `json:"bookId"`
	Data     map[string]interface{} `json:"data"`
}

var reCacheUnsafe = regexp.MustCompile(`[^0-9A-Za-z_-]`)

// cacheSanitize 兜底清洗路径段（source/bookId 正常是安全字符，防意外）。
func cacheSanitize(s string) string {
	return reCacheUnsafe.ReplaceAllString(strings.TrimSpace(s), "_")
}

func chapterCachePath(source, bookID string, limit int, content bool) string {
	c := 0
	if content {
		c = 1
	}
	name := cacheSanitize(bookID) + "_" + strconv.Itoa(limit) + "_" + strconv.Itoa(c) + ".json"
	return filepath.Join(chapterCacheDir, cacheSanitize(source), name)
}

// loadChapterCache 读取缓存；命中（存在且未过 24h）返回响应体，否则 nil。
func loadChapterCache(source, bookID string, limit int, content bool) map[string]interface{} {
	path := chapterCachePath(source, bookID, limit, content)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var env chapterCacheEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil
	}
	if time.Since(env.CachedAt) > chapterCacheTTL {
		return nil
	}
	if env.Source != source || env.BookID != bookID {
		return nil
	}
	return env.Data
}

// saveChapterCache 抓取成功后原子落盘（tmp + rename），失败静默——缓存永远不能影响主流程。
func saveChapterCache(source, bookID string, limit int, content bool, data map[string]interface{}) {
	raw, err := json.Marshal(chapterCacheEnvelope{
		CachedAt: time.Now(),
		Source:   source,
		BookID:   bookID,
		Data:     data,
	})
	if err != nil {
		return
	}
	path := chapterCachePath(source, bookID, limit, content)
	if err := os.MkdirAll(filepath.Dir(path), chapterCachePerm); err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, chapterFilePerm); err != nil {
		return
	}
	_ = os.Rename(tmp, path) // Go 在 Windows 上以 REPLACE_EXISTING 语义 rename
}
