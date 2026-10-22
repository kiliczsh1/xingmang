package bookstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// M3 的核心验证：图谱写进文件后能不能原样还原（含元数据、章节引用、
// 合并标记与证据），以及历史遗留的"未分版本"数据会不会被丢掉。

type graphSeed struct {
	versionID int
	entities  []struct {
		id       int
		name     string
		typ      string
		desc     string
		meta     string
		primary  int
		chapters []int
	}
	relations []struct {
		id       int
		src, dst int
		typ      string
		desc     string
		evidence string
		conf     float64
		chapters []int
	}
}

func seedGraph(t *testing.T, svc *Service, bookID int, versionName string, versioned bool, seed graphSeed) int {
	t.Helper()
	vid := seed.versionID
	var versionArg any
	if versioned {
		versionArg = vid
	}
	for _, e := range seed.entities {
		var primaryArg any
		if e.primary > 0 {
			primaryArg = e.primary
		}
		if _, err := svc.DB.Exec(`INSERT INTO graph_entities
			(id, book_id, name, type, description, metadata, version_id, primary_id)
			VALUES (?,?,?,?,?,?,?,?)`, e.id, bookID, e.name, e.typ, e.desc, e.meta, versionArg, primaryArg); err != nil {
			t.Fatal(err)
		}
		for _, cid := range e.chapters {
			eid := e.id
			if _, err := svc.DB.Exec(`INSERT INTO graph_chapter_refs (book_id, version_id, chapter_id, entity_id)
				VALUES (?,?,?,?)`, bookID, versionArg, cid, eid); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, r := range seed.relations {
		if _, err := svc.DB.Exec(`INSERT INTO graph_relations
			(id, book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id)
			VALUES (?,?,?,?,?,?,?,?,?)`,
			r.id, bookID, r.src, r.dst, r.typ, r.desc, r.evidence, r.conf, versionArg); err != nil {
			t.Fatal(err)
		}
		for _, cid := range r.chapters {
			rid := r.id
			if _, err := svc.DB.Exec(`INSERT INTO graph_chapter_refs (book_id, version_id, chapter_id, relation_id)
				VALUES (?,?,?,?)`, bookID, versionArg, cid, rid); err != nil {
				t.Fatal(err)
			}
		}
	}
	return vid
}

func newGraphBook(t *testing.T) (*Service, string, int) {
	t.Helper()
	svc, root := setupService(t)
	b := &models.Book{Title: "图谱书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	return svc, root, b.ID
}

func TestGraphVersion_RoundTrip(t *testing.T) {
	svc, root, bookID := newGraphBook(t)

	if _, err := svc.DB.Exec(`INSERT INTO graph_versions (id, book_id, name, scope) VALUES (7, ?, '初版', 'all')`, bookID); err != nil {
		t.Fatal(err)
	}
	var seed graphSeed
	seed.versionID = 7
	seed.entities = append(seed.entities,
		struct {
			id       int
			name     string
			typ      string
			desc     string
			meta     string
			primary  int
			chapters []int
		}{1, "林越", "人物", "副车长。第三章第一次显露身手。", `{"aliases":["越哥"]}`, 0, []int{1, 3}},
		struct {
			id       int
			name     string
			typ      string
			desc     string
			meta     string
			primary  int
			chapters []int
		}{2, "老周", "人物", "车长。", "{}", 0, []int{1}},
		struct {
			id       int
			name     string
			typ      string
			desc     string
			meta     string
			primary  int
			chapters []int
		}{3, "越哥", "人物", "林越的别名条目。", "{}", 1, nil},
		struct {
			id       int
			name     string
			typ      string
			desc     string
			meta     string
			primary  int
			chapters []int
		}{4, "太荒宗", "组织", "主角出身的宗门。", "{}", 0, nil},
	)
	seed.relations = append(seed.relations, struct {
		id       int
		src, dst int
		typ      string
		desc     string
		evidence string
		conf     float64
		chapters []int
	}{1, 1, 2, "师徒", "旧相识", "第三章“两人换了个眼神”", 0.8, []int{5}})

	seedGraph(t, svc, bookID, "初版", true, seed)

	if err := svc.MirrorGraphVersion(7); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM graph_versions WHERE id=7`).Scan(&rel)
	if !strings.HasPrefix(rel, "memory/graph/") {
		t.Fatalf("图谱应落在 memory/graph/: %q", rel)
	}
	raw, err := os.ReadFile(filepath.Join(root, "图谱书", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("图谱文件未生成: %v", err)
	}
	text := string(raw)

	// 可读性：按类型分组、关系有标题
	for _, want := range []string{"## 人物（3）", "## 组织（1）", "### 林越", "## 关系", "林越 → 老周 · 师徒"} {
		if !strings.Contains(text, want) {
			t.Fatalf("文件里缺少 %q:\n%s", want, text)
		}
	}
	// 合并标记与元数据不能丢
	if !strings.Contains(text, "primary=1") {
		t.Fatalf("合并标记 primary 未写出:\n%s", text)
	}
	if !strings.Contains(text, `{"aliases":["越哥"]}`) {
		t.Fatalf("实体元数据未写出:\n%s", text)
	}
	if !strings.Contains(text, "chapters=1,3") {
		t.Fatalf("章节引用未写出:\n%s", text)
	}
	if !strings.Contains(text, "两人换了个眼神") {
		t.Fatalf("关系证据未写出:\n%s", text)
	}

	// 解析回来逐字段比对
	doc, err := ParseGraphDoc(text)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if doc.VersionID != 7 || doc.BookID != bookID || doc.Name != "初版" {
		t.Fatalf("front-matter 不对: %+v", doc)
	}
	if len(doc.Entities) != 4 || len(doc.Relations) != 1 {
		t.Fatalf("实体/关系数量不对: %d/%d", len(doc.Entities), len(doc.Relations))
	}
	var linyue *GraphEntityFile
	for i := range doc.Entities {
		if doc.Entities[i].ID == 1 {
			linyue = &doc.Entities[i]
		}
	}
	if linyue == nil {
		t.Fatal("没找到林越")
	}
	if linyue.Name != "林越" {
		t.Fatalf("实体名应从装饰标题还原，实得 %q", linyue.Name)
	}
	if linyue.Type != "人物" || !strings.Contains(linyue.Description, "副车长") {
		t.Fatalf("实体字段不对: %+v", linyue)
	}
	if strings.TrimSpace(linyue.Metadata) != `{"aliases":["越哥"]}` {
		t.Fatalf("元数据往返不一致: %q", linyue.Metadata)
	}
	if len(linyue.ChapterIDs) != 2 || linyue.ChapterIDs[1] != 3 {
		t.Fatalf("章节引用往返不一致: %+v", linyue.ChapterIDs)
	}
	for _, e := range doc.Entities {
		if e.ID == 3 && e.PrimaryID != 1 {
			t.Fatalf("合并标记往返不一致: %+v", e)
		}
	}
	r := doc.Relations[0]
	if r.SourceID != 1 || r.TargetID != 2 || r.RelationType != "师徒" || r.Confidence != 0.8 {
		t.Fatalf("关系字段不对: %+v", r)
	}
	if !strings.Contains(r.Evidence, "两人换了个眼神") || len(r.ChapterIDs) != 1 {
		t.Fatalf("关系证据/章节不对: %+v", r)
	}
}

// 名称里的"（已合并到 #N）"是装饰，解析时要去掉，否则重渲染会越叠越长。
func TestGraphVersion_MergedMarkerNotInName(t *testing.T) {
	svc, root, bookID := newGraphBook(t)
	if _, err := svc.DB.Exec(`INSERT INTO graph_versions (id, book_id, name, scope) VALUES (1, ?, '初版', 'all')`, bookID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`INSERT INTO graph_entities (id, book_id, name, type, version_id, primary_id)
		VALUES (1, ?, '林越', '人物', 1, NULL), (2, ?, '越哥', '人物', 1, 1)`, bookID, bookID); err != nil {
		t.Fatal(err)
	}
	if err := svc.MirrorGraphVersion(1); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM graph_versions WHERE id=1`).Scan(&rel)
	raw, _ := os.ReadFile(filepath.Join(root, "图谱书", filepath.FromSlash(rel)))

	doc, err := ParseGraphDoc(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range doc.Entities {
		if strings.Contains(e.Name, "已合并") {
			t.Fatalf("实体名里不该含装饰后缀: %q", e.Name)
		}
	}
	// 再渲染一次，名字不应继续叠加后缀
	again := RenderGraphDoc(doc)
	if strings.Count(again, "（已合并到 #1）") != 1 {
		t.Fatalf("合并后缀应恰好出现一次:\n%s", again)
	}
}

// 未分版本的历史数据也要落盘，否则"指向目录就能恢复"会少一块。
func TestGraphUnversioned_LandsInOwnFile(t *testing.T) {
	svc, root, bookID := newGraphBook(t)
	type ev = struct {
		id       int
		name     string
		typ      string
		desc     string
		meta     string
		primary  int
		chapters []int
	}
	var seed graphSeed
	seed.entities = append(seed.entities, ev{11, "旧实体", "设定", "没有版本号的历史数据", "{}", 0, []int{2}})
	seedGraph(t, svc, bookID, "", false, seed)

	if err := svc.MirrorUnversionedGraph(bookID); err != nil {
		t.Fatal(err)
	}
	abs := filepath.Join(root, "图谱书", "memory", "graph", unversionedGraphName)
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("未分版本文件未生成: %v", err)
	}
	if !strings.Contains(string(raw), "旧实体") {
		t.Fatalf("内容不对:\n%s", string(raw))
	}

	// 数据被清空后，这份文件要收进回收站，而不是留个会误导人的空壳
	if _, err := svc.DB.Exec(`DELETE FROM graph_entities WHERE book_id=?`, bookID); err != nil {
		t.Fatal(err)
	}
	if err := svc.MirrorUnversionedGraph(bookID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatal("没有未分版本数据时不应留着空文件")
	}
}

// 从文件还原成可供 ImportGraph 使用的结构。
func TestGraphDoc_FeedsImport(t *testing.T) {
	doc := &GraphDoc{
		VersionID: 1, BookID: 2, Name: "初版",
		Entities: []GraphEntityFile{
			{ID: 1, Name: "林越", Type: "人物", Description: "副车长", Metadata: `{"a":1}`, ChapterIDs: []int{1}},
			{ID: 2, Name: "越哥", Type: "人物", PrimaryID: 1},
		},
		Relations: []GraphRelationFile{
			{ID: 1, SourceID: 1, TargetID: 2, RelationType: "别名", Confidence: 0.5},
		},
	}
	raws, err := GraphDocToEntities(doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(raws) != 3 {
		t.Fatalf("应导出 3 条（2 实体 + 1 关系），实得 %d", len(raws))
	}
	var first map[string]any
	if err := json.Unmarshal(raws[0], &first); err != nil {
		t.Fatal(err)
	}
	if first["name"] != "林越" || first["description"] != "副车长" {
		t.Fatalf("导出内容不对: %+v", first)
	}
	if _, ok := first["chapter_ids"]; !ok {
		t.Fatalf("章节引用应一并导出: %+v", first)
	}
	var third map[string]any
	json.Unmarshal(raws[2], &third)
	if third["relation_type"] != "别名" {
		t.Fatalf("关系导出不对: %+v", third)
	}
	// 副词条要带上 primary_id，ImportGraph 靠它回填主词条
	var second map[string]any
	json.Unmarshal(raws[1], &second)
	if second["primary_id"] != float64(1) {
		t.Fatalf("primary_id 未导出: %+v", second)
	}
}

// 版本改名后文件名必须跟着改，而且不论"改名发生在首次落盘之前还是之后"结果都一样。
func TestGraphVersion_RenameFollowsFileName(t *testing.T) {
	svc, root, bookID := newGraphBook(t)
	if _, err := svc.DB.Exec(`INSERT INTO graph_versions (id, book_id, name, scope) VALUES (1, ?, '初版', 'all')`, bookID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`INSERT INTO graph_entities (id, book_id, name, type, version_id)
		VALUES (1, ?, '林越', '人物', 1)`, bookID); err != nil {
		t.Fatal(err)
	}

	// 情形 A：先落盘，再改名
	if err := svc.MirrorGraphVersion(1); err != nil {
		t.Fatal(err)
	}
	var first string
	svc.DB.QueryRow(`SELECT file_path FROM graph_versions WHERE id=1`).Scan(&first)
	if filepath.Base(filepath.FromSlash(first)) != "1-初版.md" {
		t.Fatalf("首次文件名不对: %q", first)
	}
	svc.DB.Exec(`UPDATE graph_versions SET name='初版 重制' WHERE id=1`)
	if err := svc.MirrorGraphVersion(1); err != nil {
		t.Fatal(err)
	}
	var second string
	svc.DB.QueryRow(`SELECT file_path FROM graph_versions WHERE id=1`).Scan(&second)
	if filepath.Base(filepath.FromSlash(second)) != "1-初版 重制.md" {
		t.Fatalf("改名后文件名应跟随，实得 %q", second)
	}
	if _, err := os.Stat(filepath.Join(root, "图谱书", filepath.FromSlash(second))); err != nil {
		t.Fatalf("改名后的文件不存在: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "图谱书", filepath.FromSlash(first))); !os.IsNotExist(err) {
		t.Fatal("旧文件名不该还留着")
	}

	// 情形 B：先改名（还从未落盘），再落盘 —— 结果应与情形 A 一致
	if _, err := svc.DB.Exec(`INSERT INTO graph_versions (id, book_id, name, scope) VALUES (2, ?, '草稿版', 'all')`, bookID); err != nil {
		t.Fatal(err)
	}
	svc.DB.Exec(`UPDATE graph_versions SET name='终版' WHERE id=2`)
	if err := svc.MirrorGraphVersion(2); err != nil {
		t.Fatal(err)
	}
	var third string
	svc.DB.QueryRow(`SELECT file_path FROM graph_versions WHERE id=2`).Scan(&third)
	if filepath.Base(filepath.FromSlash(third)) != "2-终版.md" {
		t.Fatalf("未落盘就先改名时，文件名应直接用新名，实得 %q", third)
	}
}

// 图谱文件被外部改过时，程序不覆盖，只记原因（与对话、Agent 会话同一规则）。
func TestGraphVersion_ExternalEditNotOverwritten(t *testing.T) {
	svc, root, bookID := newGraphBook(t)
	if _, err := svc.DB.Exec(`INSERT INTO graph_versions (id, book_id, name, scope) VALUES (1, ?, '初版', 'all')`, bookID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DB.Exec(`INSERT INTO graph_entities (id, book_id, name, type, version_id)
		VALUES (1, ?, '林越', '人物', 1)`, bookID); err != nil {
		t.Fatal(err)
	}
	if err := svc.MirrorGraphVersion(1); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT COALESCE(file_path,'') FROM graph_versions WHERE id=1`).Scan(&rel)
	abs := filepath.Join(root, "图谱书", filepath.FromSlash(rel))

	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), "## 人物", "## 我在外部加的段落", 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	// 库里再加一个实体，触发重渲染
	if _, err := svc.DB.Exec(`INSERT INTO graph_entities (id, book_id, name, type, version_id)
		VALUES (2, ?, '老周', '人物', 1)`, bookID); err != nil {
		t.Fatal(err)
	}
	if err := svc.MirrorGraphVersion(1); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(abs)
	if !strings.Contains(string(after), "我在外部加的段落") {
		t.Fatalf("外部改动被覆盖了:\n%s", string(after))
	}
	if strings.Contains(string(after), "老周") {
		t.Fatalf("既然跳过写入，就不该有库里新增的内容:\n%s", string(after))
	}
	var pe string
	svc.DB.QueryRow(`SELECT COALESCE(parse_error,'') FROM graph_versions WHERE id=1`).Scan(&pe)
	if !strings.Contains(pe, "未覆盖") {
		t.Fatalf("跳过写入必须留下原因，实得 %q", pe)
	}
}
