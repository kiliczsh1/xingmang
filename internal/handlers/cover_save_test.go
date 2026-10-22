package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/ai/providers"
	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/database"
)

// AI 生成的封面返回的是**站内 URL**（图已在 data/generated-images 里），不是 data URL。
// 之前保存时只认 `data:`，于是这个 URL 被当成普通字符串写进了 books.cover：
//   - covers/cover.png 还是旧图（没被替换）；
//   - 旧图也从来没进 gallery/（表现就是"画廊里少一张"）。
//
// 这条用例锁住修复后的行为：走同一套归档逻辑。
const tinyPNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

func TestUpdateBookCoverFromGeneratedURL(t *testing.T) {
	// generatedImageDir 是相对路径，切到临时目录避免污染仓库
	t.Chdir(t.TempDir())

	db, err := database.Init(filepath.Join(t.TempDir(), "cover.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)

	// 建书
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"封面书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID
	bookDir := filepath.Join(root, "封面书")

	// 1) 先用上传的 data URL 设一张封面（作为"旧封面"）
	oldPNG := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 1, 2, 3, 4}
	updRec := httptest.NewRecorder()
	body := `{"cover":"data:image/png;base64,` + base64.StdEncoding.EncodeToString(oldPNG) + `"}`
	h.UpdateBook(updRec, setPathValue(httptest.NewRequest("PUT", "/api/books/1", strings.NewReader(body)), "id", itoaHandler(bookID)))
	if updRec.Code != 200 {
		t.Fatalf("设置旧封面失败: %d %s", updRec.Code, updRec.Body.String())
	}
	cur := filepath.Join(bookDir, "covers", "cover.png")
	if got, err := os.ReadFile(cur); err != nil || string(got) != string(oldPNG) {
		t.Fatalf("旧封面没写进 covers/: err=%v", err)
	}

	// 2) 造一个"AI 生成"的图，用站内 URL 保存成新封面
	genDir := filepath.Join("data", "generated-images")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newPNG, err := base64.StdEncoding.DecodeString(tinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	genName := "gen-cover-1.png"
	if err := os.WriteFile(filepath.Join(genDir, genName), newPNG, 0o644); err != nil {
		t.Fatal(err)
	}

	genRec := httptest.NewRecorder()
	genBody := `{"cover":"` + generatedImageURLPrefix + genName + `"}`
	h.UpdateBook(genRec, setPathValue(httptest.NewRequest("PUT", "/api/books/1", strings.NewReader(genBody)), "id", itoaHandler(bookID)))
	if genRec.Code != 200 {
		t.Fatalf("用生成图设封面失败: %d %s", genRec.Code, genRec.Body.String())
	}

	// 3) 新封面就是那张生成图
	now, err := os.ReadFile(cur)
	if err != nil {
		t.Fatalf("新封面不存在: %v", err)
	}
	if string(now) != string(newPNG) {
		t.Fatal("covers/cover.png 不是新生成的图")
	}

	// 4) 旧封面必须进了画廊（这正是原来缺的那一张）
	gallery, err := os.ReadDir(filepath.Join(bookDir, "gallery"))
	if err != nil {
		t.Fatalf("画廊目录不存在: %v", err)
	}
	var oldFound bool
	for _, e := range gallery {
		raw, rerr := os.ReadFile(filepath.Join(bookDir, "gallery", e.Name()))
		if rerr == nil && string(raw) == string(oldPNG) {
			oldFound = true
		}
	}
	if !oldFound {
		names := []string{}
		for _, e := range gallery {
			names = append(names, e.Name())
		}
		t.Fatalf("旧封面没有进画廊，里面只有: %v", names)
	}

	// 5) books.cover 必须清空（正文只留文件，不再存 URL / base64）
	var cover string
	db.QueryRow(`SELECT COALESCE(cover,'') FROM books WHERE id=?`, bookID).Scan(&cover)
	if cover != "" {
		t.Fatalf("books.cover 应清空，实得 %q", cover)
	}

	// 6) 画廊接口能列出这张（前端读的就是它）
	listRec := httptest.NewRecorder()
	h.GetBookGallery(listRec, setPathValue(httptest.NewRequest("GET", "/api/books/1/gallery", nil), "id", itoaHandler(bookID)))
	if listRec.Code != 200 || !strings.Contains(listRec.Body.String(), ".png") {
		t.Fatalf("画廊列表不对: %d %s", listRec.Code, listRec.Body.String())
	}
}

// 用同一个 URL 再保存一次：内容相同不该往画廊重复堆。
func TestUpdateBookCoverFromGeneratedURLIsIdempotent(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "cover2.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"幂等书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID

	genDir := filepath.Join("data", "generated-images")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatal(err)
	}
	png, _ := base64.StdEncoding.DecodeString(tinyPNG)
	if err := os.WriteFile(filepath.Join(genDir, "same.png"), png, 0o644); err != nil {
		t.Fatal(err)
	}
	body := `{"cover":"` + generatedImageURLPrefix + `same.png"}`
	for i := 0; i < 2; i++ {
		r := httptest.NewRecorder()
		h.UpdateBook(r, setPathValue(httptest.NewRequest("PUT", "/api/books/1", strings.NewReader(body)), "id", itoaHandler(bookID)))
		if r.Code != 200 {
			t.Fatalf("第 %d 次保存失败: %s", i+1, r.Body.String())
		}
	}
	entries, err := os.ReadDir(filepath.Join(root, "幂等书", "gallery"))
	if err == nil && len(entries) > 0 {
		t.Fatalf("同一张图不该进画廊，实得 %d 张", len(entries))
	}
}

// 历史遗留：早期版本把 AI 生成封面的站内 URL 直接写进了 books.cover，
// 封面因此"住"在 data/generated-images 里，作品文件夹内的 gallery/ 永远是空的。
// HealLegacyCovers 要在启动时把这类行搬成真正的文件。
func TestHealLegacyCovers(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "heal.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"遗留书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID
	bookDir := filepath.Join(root, "遗留书")

	// 目录里已经有一张"当前封面文件"（旧版本保存过的）
	oldPNG := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 9, 9, 9, 9}
	if err := os.MkdirAll(filepath.Join(bookDir, "covers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bookDir, "covers", "cover.png"), oldPNG, 0o644); err != nil {
		t.Fatal(err)
	}

	// 造一张生成图，并把它的**绝对地址**写进 books.cover（模拟旧版本 + 带 host 的形态）
	genDir := filepath.Join("data", "generated-images")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newPNG, _ := base64.StdEncoding.DecodeString(tinyPNG)
	if err := os.WriteFile(filepath.Join(genDir, "legacy.png"), newPNG, 0o644); err != nil {
		t.Fatal(err)
	}
	legacy := "http://localhost:19825" + generatedImageURLPrefix + "legacy.png"
	if _, err := db.Exec(`UPDATE books SET cover=? WHERE id=?`, legacy, bookID); err != nil {
		t.Fatal(err)
	}

	HealLegacyCovers(db)

	// 1) 生成图成为当前封面文件
	cur, err := os.ReadFile(filepath.Join(bookDir, "covers", "cover.png"))
	if err != nil || string(cur) != string(newPNG) {
		t.Fatalf("生成图没有落成 covers/cover.png: err=%v", err)
	}
	// 2) 原来那张封面进了画廊
	gallery, err := os.ReadDir(filepath.Join(bookDir, "gallery"))
	if err != nil {
		t.Fatalf("画廊目录不存在: %v", err)
	}
	var oldFound bool
	for _, e := range gallery {
		if raw, rerr := os.ReadFile(filepath.Join(bookDir, "gallery", e.Name())); rerr == nil && string(raw) == string(oldPNG) {
			oldFound = true
		}
	}
	if !oldFound {
		t.Fatal("迁移后旧封面没有进画廊")
	}
	// 3) books.cover 清空，读取统一走文件通道
	var cover string
	db.QueryRow(`SELECT COALESCE(cover,'') FROM books WHERE id=?`, bookID).Scan(&cover)
	if cover != "" {
		t.Fatalf("迁移后 books.cover 应清空，实得 %q", cover)
	}
	// 4) 幂等：再跑一次不报错、不改变结果
	HealLegacyCovers(db)
	if again, _ := os.ReadFile(filepath.Join(bookDir, "covers", "cover.png")); string(again) != string(newPNG) {
		t.Fatal("重复迁移改变了封面内容")
	}
}

// 生成图已被缓存清理时：保留 cover 原值，不要凭空把封面抹成占位图。
func TestHealLegacyCoversKeepsValueWhenSourceMissing(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "heal2.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"缺图书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	missing := generatedImageURLPrefix + "never-existed.png"
	if _, err := db.Exec(`UPDATE books SET cover=? WHERE id=?`, missing, created.Data.ID); err != nil {
		t.Fatal(err)
	}

	HealLegacyCovers(db)

	var cover string
	db.QueryRow(`SELECT COALESCE(cover,'') FROM books WHERE id=?`, created.Data.ID).Scan(&cover)
	if cover != missing {
		t.Fatalf("源文件不存在时不该动 cover，实得 %q", cover)
	}
}

// AI 生成图直接进 books/<书名>/gallery/（与被顶掉的封面同一个池子），
// 设为封面时复制到 covers/；画廊里始终留着这张图，随时能再点回来。
func TestGeneratedImageLandsInGalleryAndPromotes(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "pool.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"池子书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID
	bookDir := filepath.Join(root, "池子书")
	svc := bookstore.NewService(db)

	a, _ := base64.StdEncoding.DecodeString(tinyPNG)
	b := append(append([]byte{}, a...), 'A')

	// 模拟两次 AI 生成：直接进画廊
	nameA, err := svc.AddGalleryBytes(bookID, a, ".png")
	if err != nil {
		t.Fatalf("生成图入库失败: %v", err)
	}
	nameB, err := svc.AddGalleryBytes(bookID, b, ".png")
	if err != nil {
		t.Fatalf("生成图入库失败: %v", err)
	}
	if again, _ := svc.AddGalleryBytes(bookID, a, ".png"); again != nameA {
		t.Fatalf("同一张图不该在画廊里重复入库: %s vs %s", again, nameA)
	}
	if raw, err := os.ReadFile(filepath.Join(bookDir, "gallery", nameA)); err != nil || string(raw) != string(a) {
		t.Fatalf("生成图没落进 gallery/: err=%v", err)
	}

	put := func(url string) {
		r := httptest.NewRecorder()
		body := `{"cover":"` + url + `"}`
		h.UpdateBook(r, setPathValue(httptest.NewRequest("PUT", "/api/books/1", strings.NewReader(body)), "id", itoaHandler(bookID)))
		if r.Code != 200 {
			t.Fatalf("设封面失败: %d %s", r.Code, r.Body.String())
		}
	}
	put(bookGalleryURLPrefix(bookID) + nameA)
	put(bookGalleryURLPrefix(bookID) + nameB)

	if cur, err := os.ReadFile(filepath.Join(bookDir, "covers", "cover.png")); err != nil || string(cur) != string(b) {
		t.Fatalf("当前封面应为 B: err=%v", err)
	}
	var oldInGallery bool
	if entries, err := os.ReadDir(filepath.Join(bookDir, "gallery")); err == nil {
		for _, e := range entries {
			if raw, rerr := os.ReadFile(filepath.Join(bookDir, "gallery", e.Name())); rerr == nil && string(raw) == string(a) {
				oldInGallery = true
			}
		}
	}
	if !oldInGallery {
		t.Fatal("被顶掉的 A 没有留在 gallery/")
	}
	var cover string
	db.QueryRow(`SELECT COALESCE(cover,'') FROM books WHERE id=?`, bookID).Scan(&cover)
	if cover != "" {
		t.Fatalf("books.cover 应清空，实得 %q", cover)
	}
	// 画廊里两张都在（AI 生成图不会被"用掉就消失"），且没有因为归档再堆一份 A
	items, err := svc.ListGallery(bookID)
	if err != nil || len(items) != 2 {
		t.Fatalf("gallery/ 应有 2 张（A、B），实得 %d (err=%v)", len(items), err)
	}
}

// 缓存目录里的生成图按归属搬进作品 gallery/；认不出归属的原地不动。
func TestMigrateGeneratedImagesByOwnership(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "mig.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"迁移书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID
	bookDir := filepath.Join(root, "迁移书")

	// 先让这本书有一张封面（内容与缓存里的一张图相同 → 用来判归属）
	cover, _ := base64.StdEncoding.DecodeString(tinyPNG)
	r := httptest.NewRecorder()
	body := `{"cover":"data:image/png;base64,` + tinyPNG + `"}`
	h.UpdateBook(r, setPathValue(httptest.NewRequest("PUT", "/api/books/1", strings.NewReader(body)), "id", itoaHandler(bookID)))
	if r.Code != 200 {
		t.Fatalf("设封面失败: %s", r.Body.String())
	}

	genDir := filepath.Join("data", "generated-images")
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(genDir, "owned.png"), cover, 0o644); err != nil {
		t.Fatal(err)
	}
	stray := []byte{0x89, 'P', 'N', 'G', 1, 1, 1, 1, 1}
	if err := os.WriteFile(filepath.Join(genDir, "stray.png"), stray, 0o644); err != nil {
		t.Fatal(err)
	}

	MigrateGeneratedImages(db)

	if raw, err := os.ReadFile(filepath.Join(bookDir, "gallery", "owned.png")); err != nil || string(raw) != string(cover) {
		t.Fatalf("认得出的图应进作品 gallery/: err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(genDir, "owned.png")); !os.IsNotExist(err) {
		t.Fatal("缓存目录里的原文件应被搬走")
	}
	// 认不出归属的原地不动，不另建目录
	if raw, err := os.ReadFile(filepath.Join(genDir, "stray.png")); err != nil || string(raw) != string(stray) {
		t.Fatalf("认不出的图应留在缓存目录: err=%v", err)
	}
	if entries, err := os.ReadDir(root); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "_orphan") {
				t.Fatalf("不该再产生兜底目录: %s", e.Name())
			}
		}
	}

	// 幂等：再跑一次不该报错、也不该改变结果
	MigrateGeneratedImages(db)
	if _, err := os.Stat(filepath.Join(bookDir, "gallery", "owned.png")); err != nil {
		t.Fatalf("重复迁移后文件不见了: %v", err)
	}
}

// 封面弹窗里"AI 生成"那条路：带 bookId 时图直接落进该作品的 gallery/，
// 返回的站内地址也指向画廊 —— 前端画廊里能立刻看到，不用等保存封面。
func TestPersistGeneratedImageIntoBookGallery(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "sink.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"生成书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID

	svc := bookstore.NewService(db)
	url := persistGeneratedImage(
		providers.GeneratedImage{B64JSON: tinyPNG},
		galleryImageSink{svc: svc, bookID: bookID},
	)
	if !strings.HasPrefix(url, bookGalleryURLPrefix(bookID)) {
		t.Fatalf("返回地址应指向作品画廊，实得 %q", url)
	}
	name := strings.TrimPrefix(url, bookGalleryURLPrefix(bookID))
	want, _ := base64.StdEncoding.DecodeString(tinyPNG)
	if raw, err := os.ReadFile(filepath.Join(root, "生成书", "gallery", name)); err != nil || string(raw) != string(want) {
		t.Fatalf("生成图没落进 gallery/: err=%v", err)
	}
	// 画廊接口立刻能看到它
	listRec := httptest.NewRecorder()
	h.GetBookGallery(listRec, setPathValue(httptest.NewRequest("GET", "/api/books/1/gallery", nil), "id", itoaHandler(bookID)))
	if !strings.Contains(listRec.Body.String(), name) {
		t.Fatalf("画廊列表里没有刚生成的图: %s", listRec.Body.String())
	}
	// 不带 bookId 时仍落共享缓存目录（工作流等无作品上下文的调用）
	shared := persistGeneratedImage(providers.GeneratedImage{B64JSON: tinyPNG}, globalImageSink)
	if !strings.HasPrefix(shared, generatedImageURLPrefix) {
		t.Fatalf("无作品上下文时应落缓存目录，实得 %q", shared)
	}
}

// 迁移进来的历史生成图保留的是缓存命名（名字里没有内容指纹）。
// 归档判重必须按**内容**，按文件名后缀匹配会漏 —— 同一个文件会在画廊里出现两遍。
func TestArchiveCoverDedupesByContentNotName(t *testing.T) {
	t.Chdir(t.TempDir())
	db, err := database.Init(filepath.Join(t.TempDir(), "dedupe.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(db)
	rec := httptest.NewRecorder()
	h.CreateBook(rec, httptest.NewRequest("POST", "/api/books", strings.NewReader(`{"title":"判重书","status":"draft"}`)))
	var created struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.Data.ID == 0 {
		t.Fatalf("建书失败: %s", rec.Body.String())
	}
	bookID := created.Data.ID
	bookDir := filepath.Join(root, "判重书")
	svc := bookstore.NewService(db)

	x, _ := base64.StdEncoding.DecodeString(tinyPNG)
	y := append(append([]byte{}, x...), 'Y')

	// 手动放一个"缓存命名"的画廊文件（名字里没有内容指纹）
	gallery := filepath.Join(bookDir, "gallery")
	if err := os.MkdirAll(gallery, 0o755); err != nil {
		t.Fatal(err)
	}
	cacheName := "1792348904115792100_c02d0e835768d7fa.png"
	if err := os.WriteFile(filepath.Join(gallery, cacheName), x, 0o644); err != nil {
		t.Fatal(err)
	}

	// 先设 X 为封面，再设 Y —— 这一步会把 X 归档，而画廊里已经有 X 了
	if err := svc.SetCoverBytes(bookID, x, ".png"); err != nil {
		t.Fatalf("设 X 为封面失败: %v", err)
	}
	if err := svc.SetCoverBytes(bookID, y, ".png"); err != nil {
		t.Fatalf("设 Y 为封面失败: %v", err)
	}

	items, err := svc.ListGallery(bookID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		names := []string{}
		for _, it := range items {
			names = append(names, it.Name)
		}
		t.Fatalf("画廊里应只有 1 张（同一内容不重复归档），实得 %d: %v", len(items), names)
	}
	if items[0].Name != cacheName {
		t.Fatalf("应保留原来那张，实得 %s", items[0].Name)
	}
	// 当前封面是 Y
	if cur, err := os.ReadFile(filepath.Join(bookDir, "covers", "cover.png")); err != nil || string(cur) != string(y) {
		t.Fatalf("当前封面应为 Y: err=%v", err)
	}
}
