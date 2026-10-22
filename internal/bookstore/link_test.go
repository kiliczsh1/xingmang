package bookstore

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 链接目录是"下一次更新直接指向目录就能用"的落点：换机器 / 重装之后，
// 把作品库指向原目录，全部数据（正文 + 四类记忆）都该回来。
//
// 这一组用例的重点不是"能读"，而是**幂等**：重复点同一个目录不能产生重复数据。

// 造一个"别人的作品库"：直接在磁盘上写文件，不经过任何接口。
func buildExternalBook(t *testing.T, root, title string) string {
	t.Helper()
	dir := filepath.Join(root, title)
	must(t, os.MkdirAll(filepath.Join(dir, "memory", "chats"), 0o755))
	must(t, os.MkdirAll(filepath.Join(dir, "memory", "graph"), 0o755))
	must(t, os.MkdirAll(filepath.Join(dir, "memory", "analysis"), 0o755))
	must(t, os.MkdirAll(filepath.Join(dir, "memory", "history", "2026-10"), 0o755))
	must(t, os.MkdirAll(filepath.Join(dir, "characters"), 0o755))

	// book.json
	must(t, os.WriteFile(filepath.Join(dir, "book.json"),
		[]byte(`{"schema":1,"title":"`+title+`","author":"某人","status":"draft","tags":["测试"],"target_words":100000}`), 0o644))
	// 章节
	must(t, os.WriteFile(filepath.Join(dir, "001 序章.md"), []byte("序章正文"), 0o644))
	must(t, os.WriteFile(filepath.Join(dir, "002 上车.md"), []byte("上车正文"), 0o644))
	// 人物
	must(t, os.WriteFile(filepath.Join(dir, "characters", "林越.md"), []byte("副车长"), 0o644))

	// 对话
	chat := strings.Join([]string{
		"---", "kind: chat", "conversation_id: 101", "book_id: 0", "title: 列车开场",
		`created_at: "2026-10-18 22:40:00"`, `updated_at: "2026-10-18 22:41:00"`, "---", "",
		"<!-- xm:chat v=1 -->", "", "# 列车开场", "",
		"## 用户 · 22:40", "", `<!-- xm:msg id=1 role=user at="2026-10-18 22:40:12" -->`, "",
		"主角上车那段太平了。", "", "<!-- xm:msg-end -->", "",
		"## 助手 · 22:41", "", `<!-- xm:msg id=2 role=assistant at="2026-10-18 22:41:05" -->`, "",
		"可以先把车长的视线写成一种判定的动作。", "", "<!-- xm:msg-end -->", "",
	}, "\n")
	must(t, os.WriteFile(filepath.Join(dir, "memory", "chats", "20261018-2240-列车开场-aaaaaa.md"), []byte(chat), 0o644))

	// 图谱（一个版本，两个实体，一条关系）
	graph := strings.Join([]string{
		"---", "kind: graph", "version_id: 7", "book_id: 0", "name: 初版", "scope: all",
		"chapter_count: 2", "entity_count: 2", "relation_count: 1",
		`created_at: "2026-10-18 22:45:00"`, `updated_at: "2026-10-18 22:45:00"`, "---", "",
		"<!-- xm:graph v=1 -->", "", "# 初版", "", "## 人物", "",
		"### 林越", "", "<!-- xm:e id=1 type=人物 chapters=1,2 -->", "", "副车长。", "",
		"<!-- xm:e-end -->", "", "### 老周", "", "<!-- xm:e id=2 type=人物 chapters=1 -->", "",
		"车长。", "", "<!-- xm:e-end -->", "", "## 关系", "",
		"### 林越 → 老周 · 师徒", "", "<!-- xm:r id=1 src=1 dst=2 type=师徒 conf=0.8 chapters=2 -->", "",
		"旧相识", "", "<!-- xm:r-end -->", "",
	}, "\n")
	must(t, os.WriteFile(filepath.Join(dir, "memory", "graph", "7-初版.md"), []byte(graph), 0o644))

	// 拆书
	analysis := strings.Join([]string{
		"---", "kind: analysis", "book_id: 0", "external_id: a1", "title: 黄金三章拆解",
		"source: 拆书", "status: completed", "chapters: 1,2",
		`created_at: "2026-10-18 22:45:00"`, `updated_at: "2026-10-18 22:45:00"`, "---", "",
		"<!-- xm:record v=1 -->", "", "# 黄金三章拆解", "", "## 结果", "", "<!-- xm:result -->", "",
		"第一处问题是节奏。", "", "<!-- xm:result-end -->", "",
		"<!-- xm:payload -->", "", "```json", `{"id":"a1","fieldValues":{"重点":"节奏"}}`, "```", "",
		"<!-- xm:payload-end -->", "",
	}, "\n")
	must(t, os.WriteFile(filepath.Join(dir, "memory", "analysis", "20261018-2245-黄金三章拆解-bbbbbb.md"), []byte(analysis), 0o644))

	// 生成历史（按月分目录）
	hist := strings.Join([]string{
		"---", "kind: generation", "book_id: 0", "external_id: h1", "title: 续写",
		"source: continuation", "status: completed",
		`created_at: "2026-10-18 23:10:00"`, `updated_at: "2026-10-18 23:10:00"`, "---", "",
		"<!-- xm:record v=1 -->", "", "# 续写", "", "## 结果", "", "<!-- xm:result -->", "",
		"他把车票递了过去。", "", "<!-- xm:result-end -->", "",
	}, "\n")
	must(t, os.WriteFile(filepath.Join(dir, "memory", "history", "2026-10", "20261018-2310-续写-cccccc.md"), []byte(hist), 0o644))
	return dir
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// 空库 + 指向一个"外来"作品库：链接之后全部数据都该出现。
func TestLinkRoot_RebuildsEverything(t *testing.T) {
	svc, _ := setupService(t)
	vault := t.TempDir()
	buildExternalBook(t, vault, "外部作品")

	rep, err := svc.LinkRoot(vault)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Books != 1 {
		t.Fatalf("应发现 1 部作品，实得 %d", rep.Books)
	}

	var bookID int
	svc.DB.QueryRow(`SELECT id FROM books WHERE folder_path='外部作品'`).Scan(&bookID)
	if bookID == 0 {
		t.Fatal("书没登记进来")
	}
	// book.json 的元数据要读进来
	var title, author, tags string
	var target int
	svc.DB.QueryRow(`SELECT title, COALESCE(author,''), COALESCE(tags,''), target_words FROM books WHERE id=?`, bookID).
		Scan(&title, &author, &tags, &target)
	if title != "外部作品" || author != "某人" || target != 100000 {
		t.Fatalf("book.json 元数据没读进来: %q %q %q %d", title, author, tags, target)
	}

	counts := map[string]int{}
	for _, q := range []struct {
		key   string
		query string
	}{
		{"chapters", `SELECT COUNT(*) FROM chapters WHERE book_id=?`},
		{"characters", `SELECT COUNT(*) FROM characters WHERE book_id=?`},
		{"conversations", `SELECT COUNT(*) FROM conversations WHERE book_id=?`},
		{"graph_versions", `SELECT COUNT(*) FROM graph_versions WHERE book_id=?`},
		{"graph_entities", `SELECT COUNT(*) FROM graph_entities WHERE book_id=?`},
		{"graph_relations", `SELECT COUNT(*) FROM graph_relations WHERE book_id=?`},
		{"graph_refs", `SELECT COUNT(*) FROM graph_chapter_refs WHERE book_id=?`},
		{"analysis", `SELECT COUNT(*) FROM memory_records WHERE book_id=? AND kind='analysis'`},
		{"generation", `SELECT COUNT(*) FROM memory_records WHERE book_id=? AND kind='generation'`},
	} {
		var n int
		svc.DB.QueryRow(q.query, bookID).Scan(&n)
		counts[q.key] = n
	}
	want := map[string]int{
		"chapters": 2, "characters": 1, "conversations": 1,
		"graph_versions": 1, "graph_entities": 2, "graph_relations": 1, "graph_refs": 4,
		"analysis": 1, "generation": 1,
	}
	for k, v := range want {
		if counts[k] != v {
			t.Fatalf("%s 应为 %d，实得 %d", k, v, counts[k])
		}
	}
	// 关系端点必须真的连上（断的关系等于没重建）
	var src, dst int
	svc.DB.QueryRow(`SELECT source_id, target_id FROM graph_relations WHERE book_id=?`, bookID).Scan(&src, &dst)
	if src == 0 || dst == 0 || src == dst {
		t.Fatalf("关系端点没接上: %d → %d", src, dst)
	}

	// 文件路径要写回索引（否则下次保存会另起一个新文件）
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM conversations WHERE book_id=?`, bookID).Scan(&rel)
	if rel != "memory/chats/20261018-2240-列车开场-aaaaaa.md" {
		t.Fatalf("对话路径没写回: %q", rel)
	}
}

// **幂等**：同一个目录链接两次，数据不能翻倍。
func TestLinkRoot_Idempotent(t *testing.T) {
	svc, _ := setupService(t)
	vault := t.TempDir()
	buildExternalBook(t, vault, "外部作品")

	if _, err := svc.LinkRoot(vault); err != nil {
		t.Fatal(err)
	}
	var bookID int
	svc.DB.QueryRow(`SELECT id FROM books WHERE folder_path='外部作品'`).Scan(&bookID)

	snapshot := func() map[string]int {
		out := map[string]int{}
		for _, q := range []struct{ key, query string }{
			{"books", `SELECT COUNT(*) FROM books`},
			{"chapters", `SELECT COUNT(*) FROM chapters`},
			{"characters", `SELECT COUNT(*) FROM characters`},
			{"conversations", `SELECT COUNT(*) FROM conversations`},
			{"graph_versions", `SELECT COUNT(*) FROM graph_versions`},
			{"graph_entities", `SELECT COUNT(*) FROM graph_entities`},
			{"graph_relations", `SELECT COUNT(*) FROM graph_relations`},
			{"graph_refs", `SELECT COUNT(*) FROM graph_chapter_refs`},
			{"records", `SELECT COUNT(*) FROM memory_records`},
		} {
			var n int
			svc.DB.QueryRow(q.query).Scan(&n)
			out[q.key] = n
		}
		return out
	}
	before := snapshot()

	// 第二次链接：应识别为"已登记"，只更新不新增
	rep, err := svc.LinkRoot(vault)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Created) != 0 {
		t.Fatalf("第二次不该再有新登记的作品: %v", rep.Created)
	}
	after := snapshot()
	for k, v := range before {
		if after[k] != v {
			t.Fatalf("%s 翻倍了: %d → %d", k, v, after[k])
		}
	}
	// 书的 id 必须保持不变（会话/图谱都按它引用）
	var again int
	svc.DB.QueryRow(`SELECT id FROM books WHERE folder_path='外部作品'`).Scan(&again)
	if again != bookID {
		t.Fatalf("书的 id 变了: %d → %d", bookID, again)
	}
}

// 预览只读：不写库、不改目录。
func TestPreviewRoot_IsReadOnly(t *testing.T) {
	svc, _ := setupService(t)
	vault := t.TempDir()
	buildExternalBook(t, vault, "外部作品")

	rep, err := svc.PreviewRoot(vault)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Created) != 1 || rep.Created[0] != "外部作品" {
		t.Fatalf("预览结果不对: %+v", rep)
	}
	var books int
	svc.DB.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&books)
	if books != 0 {
		t.Fatal("预览不该写库")
	}
	// 文件也不该被动
	if _, err := os.Stat(filepath.Join(vault, "外部作品", "book.json")); err != nil {
		t.Fatal("预览不该动目录")
	}
}

// 手工拷进作品库的目录（不走"切换目录"）也要能被登记。
func TestRebuildIndex_PicksUpCopiedDirs(t *testing.T) {
	svc, root := setupService(t)
	// 直接往当前作品库里放一个目录，模拟"用资源管理器拷进来"
	buildExternalBook(t, root, "拷进来的书")

	rep := &LinkReport{}
	if err := svc.LinkAll(rep); err != nil {
		t.Fatal(err)
	}
	var n int
	svc.DB.QueryRow(`SELECT COUNT(*) FROM books WHERE folder_path='拷进来的书'`).Scan(&n)
	if n != 1 {
		t.Fatalf("拷进来的目录没被登记: %d", n)
	}
	if rep.Details["chapters"] != 2 {
		t.Fatalf("章节应重建 2 章，实得 %v", rep.Details)
	}
}

// 链接之后，从界面保存应该写回**原来那个文件**，而不是另起一个。
func TestLinkRoot_SaveGoesBackToSameFile(t *testing.T) {
	svc, _ := setupService(t)
	vault := t.TempDir()
	buildExternalBook(t, vault, "外部作品")
	if _, err := svc.LinkRoot(vault); err != nil {
		t.Fatal(err)
	}
	var bookID, convID int
	var rel string
	svc.DB.QueryRow(`SELECT id FROM books WHERE folder_path='外部作品'`).Scan(&bookID)
	svc.DB.QueryRow(`SELECT id, file_path FROM conversations WHERE book_id=?`, bookID).Scan(&convID, &rel)
	if convID == 0 {
		t.Fatal("对话没登记")
	}

	before, _ := os.ReadDir(filepath.Join(vault, "外部作品", "memory", "chats"))
	if _, err := svc.AppendChatMessage(convID, "user", "追加的一条", ""); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadDir(filepath.Join(vault, "外部作品", "memory", "chats"))
	if len(after) != len(before) {
		t.Fatalf("追加消息不该新建文件: %d → %d", len(before), len(after))
	}
	raw, _ := os.ReadFile(filepath.Join(vault, "外部作品", filepath.FromSlash(rel)))
	if !strings.Contains(string(raw), "追加的一条") {
		t.Fatalf("消息没写进原文件:\n%s", string(raw))
	}
}

// book.json 改名之后重新链接，书架的标题要跟着走。
func TestLinkRoot_ReflectsBookJSONRename(t *testing.T) {
	svc, _ := setupService(t)
	vault := t.TempDir()
	dir := buildExternalBook(t, vault, "旧名字")
	if _, err := svc.LinkRoot(vault); err != nil {
		t.Fatal(err)
	}
	// 在外部改 book.json（模拟用户在另一台机器上改了书名）
	must(t, os.WriteFile(filepath.Join(dir, "book.json"),
		[]byte(`{"schema":1,"title":"新名字","author":"某人","status":"draft"}`), 0o644))
	if _, err := svc.LinkRoot(vault); err != nil {
		t.Fatal(err)
	}
	var title string
	svc.DB.QueryRow(`SELECT title FROM books WHERE folder_path='旧名字'`).Scan(&title)
	if title != "新名字" {
		t.Fatalf("书名应以 book.json 为准，实得 %q", title)
	}
}

// 导出必须包含 memory/ —— 否则"拷走 zip 就等于拷走作品"不成立。
func TestExportZip_IncludesMemory(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "导出书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	rec, err := svc.CreateRecord(b.ID, MemoryRecordInput{
		Kind: MemoryKindAnalysis, Title: "拆解", Source: "拆书", Result: "结论",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rec.FilePath, "memory/analysis/") {
		t.Fatalf("前置条件不成立: %q", rec.FilePath)
	}

	var buf bytes.Buffer
	if err := svc.ExportZip(b.ID, &buf); err != nil {
		t.Fatal(err)
	}
	// zip 里应能找到 memory/ 下的记录文件
	if !bytes.Contains(buf.Bytes(), []byte("memory/analysis/")) {
		t.Fatal("导出的 zip 里没有 memory/")
	}
	_ = root
	_ = json.Valid
}
