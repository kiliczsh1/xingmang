package handlers

// 章节本地缓存的单元测试：落盘/命中/TTL 过期/路径清洗。

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func withTempCacheDir(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

func TestChapterCacheRoundTrip(t *testing.T) {
	withTempCacheDir(t)

	data := map[string]interface{}{
		"source":   "qimao",
		"bookId":   "1834789",
		"chapters": []map[string]interface{}{{"id": "1", "title": "第一章", "content": "正文"}},
	}
	saveChapterCache("qimao", "1834789", 10, true, data)

	got := loadChapterCache("qimao", "1834789", 10, true)
	if got == nil {
		t.Fatal("24h 内应命中缓存")
	}
	raw, _ := json.Marshal(got["chapters"])
	if string(raw) == "" || raw == nil {
		t.Fatal("缓存数据缺失 chapters")
	}

	// 参数不同不命中
	if loadChapterCache("qimao", "1834789", 20, true) != nil {
		t.Fatal("limit 不同不应命中")
	}
	if loadChapterCache("qimao", "1834789", 10, false) != nil {
		t.Fatal("content 不同不应命中")
	}
	if loadChapterCache("fanqie", "1834789", 10, true) != nil {
		t.Fatal("source 不同不应命中")
	}

	// 缓存文件确实落在 chapter-cache/<source>/ 下
	if _, err := os.Stat(filepath.Join("chapter-cache", "qimao", "1834789_10_1.json")); err != nil {
		t.Fatalf("缓存文件未落盘: %v", err)
	}
}

func TestChapterCacheExpired(t *testing.T) {
	withTempCacheDir(t)

	saveChapterCache("fanqie", "7143038691944959011", 10, true, map[string]interface{}{"ok": 1})

	// 把 cachedAt 改成 25 小时前
	path := chapterCachePath("fanqie", "7143038691944959011", 10, true)
	raw, _ := os.ReadFile(path)
	var env chapterCacheEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	env.CachedAt = time.Now().Add(-25 * time.Hour)
	fresh, _ := json.Marshal(env)
	if err := os.WriteFile(path, fresh, chapterFilePerm); err != nil {
		t.Fatal(err)
	}

	if loadChapterCache("fanqie", "7143038691944959011", 10, true) != nil {
		t.Fatal("超过 24h 不应命中")
	}
}

func TestChapterCacheSanitize(t *testing.T) {
	withTempCacheDir(t)

	// 恶意/异常 bookId 不应逃出缓存目录
	saveChapterCache("../evil", "a/b\\c:d", 10, true, map[string]interface{}{"ok": 1})
	path := chapterCachePath("../evil", "a/b\\c:d", 10, true)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("清洗后的路径应可写: %v", err)
	}
	if !filepath.IsLocal(path) {
		t.Fatalf("路径应保持在缓存目录内: %s", path)
	}
}
