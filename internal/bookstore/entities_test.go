package bookstore

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

func TestSyncCharacterAndEntry(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "实体书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}

	info := "剑修，寡言。"
	personality := "冷静"
	c := &models.Character{BookID: b.ID, Name: "林越", Gender: "男", Personality: &personality, Info: &info, Folder: "主角"}
	rel, err := svc.SyncCharacter(c, "")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "characters/主角/林越.md" {
		t.Fatalf("人物路径不对: %q", rel)
	}
	raw, err := os.ReadFile(filepath.Join(root, "实体书", rel))
	if err != nil {
		t.Fatal(err)
	}
	if !containsAll(string(raw), "gender: 男", "personality: 冷静", "剑修") {
		t.Fatalf("人物文件内容不对: %s", raw)
	}

	// 改名 + 换分组：旧文件应消失
	c.Name = "林越之"
	c.Folder = "配角"
	rel2, err := svc.SyncCharacter(c, rel)
	if err != nil {
		t.Fatal(err)
	}
	if rel2 != "characters/配角/林越之.md" {
		t.Fatalf("改名后路径不对: %q", rel2)
	}
	if _, err := os.Stat(filepath.Join(root, "实体书", rel)); !os.IsNotExist(err) {
		t.Fatal("旧人物文件应删除")
	}

	// 世界观
	category := "修炼体系"
	e := &models.Entry{BookID: &b.ID, Name: "灵气等级", Description: "共九层", CategoryType: category, Tags: `["境界"]`}
	erel, err := svc.SyncEntry(e, category, "")
	if err != nil {
		t.Fatal(err)
	}
	if erel != "world/修炼体系/灵气等级.md" {
		t.Fatalf("条目路径不对: %q", erel)
	}
}

func TestMemoGlobalAndDocs(t *testing.T) {
	svc, root := setupService(t)

	m := &models.Memo{Title: "随手记", Content: "灵感一句", Category: "灵感"}
	rel, err := svc.SyncMemo(m, 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "memos/随手记.md" {
		t.Fatalf("备忘录路径不对: %q", rel)
	}
	if _, err := os.Stat(filepath.Join(root, "_memos", "随手记.md")); err != nil {
		t.Fatalf("全局备忘录应落在 _memos/: %v", err)
	}

	b := &models.Book{Title: "文档书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	doc, err := svc.CreateDoc(b.ID, "outline", "总纲", "第一幕：出发")
	if err != nil {
		t.Fatal(err)
	}
	if doc.FilePath != "outlines/总纲.md" {
		t.Fatalf("大纲路径不对: %q", doc.FilePath)
	}
	list, err := svc.ListDocs(b.ID, "outline")
	if err != nil || len(list) != 1 || list[0].Content != "第一幕：出发" {
		t.Fatalf("列表/正文读取不对: %+v %v", list, err)
	}
	if _, err := svc.UpdateDoc(doc.ID, "总纲 v2", "改过的正文"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "文档书", "outlines", "总纲 v2.md")); err != nil {
		t.Fatalf("改名后文件不存在: %v", err)
	}
	if err := svc.DeleteDoc(doc.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "文档书", "outlines", "总纲 v2.md")); !os.IsNotExist(err) {
		t.Fatal("删除后文件应移走")
	}
}

func containsAll(s string, subs ...string) bool {
	for _, sub := range subs {
		if !contains(s, sub) {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestScanBook_AddsAndMarksMissing(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "扫描书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, "扫描书")

	// 外部新增一章 + 一个人物
	if err := os.WriteFile(filepath.Join(dir, "002 外部新增.md"), []byte("---\nsummary: 外部写的\n---\n\n正文"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(dir, "characters", "路人"), 0o755)
	os.WriteFile(filepath.Join(dir, "characters", "路人", "甲.md"), []byte("路人甲"), 0o644)

	res, err := svc.ScanBook(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.AddedChapters != 1 || res.AddedCharacters != 1 {
		t.Fatalf("对账结果不对: %+v", res)
	}
	var path, summary string
	var order int
	svc.DB.QueryRow(`SELECT file_path, summary, order_num FROM chapters WHERE book_id=?`, b.ID).Scan(&path, &summary, &order)
	if path != "002 外部新增.md" || order != 2 || summary != "外部写的" {
		t.Fatalf("章节索引不对: %q %d %q", path, order, summary)
	}

	// 外部删除 → missing=1，但不删行
	os.Remove(filepath.Join(dir, "002 外部新增.md"))
	res, err = svc.ScanBook(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	if res.MissingChapters != 1 {
		t.Fatalf("应标记 1 章缺失: %+v", res)
	}
	var missing int
	svc.DB.QueryRow(`SELECT missing FROM chapters WHERE book_id=?`, b.ID).Scan(&missing)
	if missing != 1 {
		t.Fatal("missing 未置 1")
	}
}

func TestUpdateChapter_Conflict(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "冲突书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	c := &models.Chapter{BookID: b.ID, Title: "章", Content: "原文"}
	if err := svc.CreateChapter(c); err != nil {
		t.Fatal(err)
	}
	// 读一次，索引里的 hash = 当前文件
	if _, err := svc.ChapterText(c.ID); err != nil {
		t.Fatal(err)
	}
	// 模拟 Obsidian 外部修改
	path := filepath.Join(root, "冲突书", c.FilePath)
	if err := os.WriteFile(path, []byte("外部改过"), 0o644); err != nil {
		t.Fatal(err)
	}
	c.Content = "编辑器里的版本"
	if err := svc.UpdateChapter(c); err != ErrFileConflict {
		t.Fatalf("应返回 ErrFileConflict，实得 %v", err)
	}
}

func TestImportFolderAndExportZip(t *testing.T) {
	svc, root := setupService(t)

	src := t.TempDir()
	os.WriteFile(filepath.Join(src, "book.json"), []byte(`{"title":"外部作品","author":"某人","status":"draft"}`), 0o644)
	os.WriteFile(filepath.Join(src, "001 序章.md"), []byte("序章正文"), 0o644)
	os.MkdirAll(filepath.Join(src, "characters"), 0o755)
	os.WriteFile(filepath.Join(src, "characters", "甲.md"), []byte("角色甲"), 0o644)

	b, err := svc.ImportFolder(src)
	if err != nil {
		t.Fatal(err)
	}
	if b.Title != "外部作品" {
		t.Fatalf("书名应取 book.json: %q", b.Title)
	}
	if _, err := os.Stat(filepath.Join(root, "外部作品", "001 序章.md")); err != nil {
		t.Fatalf("文件未复制进库: %v", err)
	}
	var chapters int
	svc.DB.QueryRow(`SELECT COUNT(*) FROM chapters WHERE book_id=?`, b.ID).Scan(&chapters)
	if chapters != 1 {
		t.Fatalf("应索引 1 章，实得 %d", chapters)
	}
	// 原目录必须还在（导入是复制，不是移动）
	if _, err := os.Stat(filepath.Join(src, "001 序章.md")); err != nil {
		t.Fatal("原目录被移走了")
	}

	var buf bytes.Buffer
	if err := svc.ExportZip(b.ID, &buf); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("导出的不是合法 zip: %v", err)
	}
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	if !names["外部作品/001 序章.md"] || !names["外部作品/book.json"] {
		t.Fatalf("zip 条目不对: %v", names)
	}
}

func TestImportBookJSON_OldShapes(t *testing.T) {
	svc, root := setupService(t)

	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"book包装", `{"book":{"title":"旧书A","author":"甲"},"chapters":[{"title":"序章","content":"一"},{"title":"第一章","content":"二"}]}`, "旧书A"},
		{"平铺", `{"title":"旧书B","chapters":[{"name":"章一","text":"正文"}]}`, "旧书B"},
		{"纯数组", `[{"title":"章一","content":"x"}]`, "按文件名"},
	}
	for _, c := range cases {
		b, n, err := svc.ImportBookJSON([]byte(c.raw), c.want+".json")
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if b.Title != c.want {
			t.Fatalf("%s 书名应为 %q，实得 %q", c.name, c.want, b.Title)
		}
		if n == 0 {
			t.Fatalf("%s 未导入章节", c.name)
		}
		// 每章都必须是真实文件
		entries, _ := os.ReadDir(filepath.Join(root, b.FolderPath))
		var md int
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
				md++
			}
		}
		if md != n {
			t.Fatalf("%s: 文件数 %d 与章节数 %d 不一致", c.name, md, n)
		}
	}
}
