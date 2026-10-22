package bookstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// 知识图谱的文件存储（docs/books-memory-storage.md §5.3）。
//
// 形态：实体与关系各成一块，全部平铺；元数据与证据用独立块按 id 关联。
//
//	---
//	kind: graph
//	version_id: 3
//	book_id: 22
//	name: 初版
//	---
//
//	<!-- xm:graph v=1 -->
//
//	## 人物
//
//	### 林越
//
//	<!-- xm:e id=1 type=人物 chapters="1,3" -->
//
//	副车长。第三章里第一次显露身手。
//
//	<!-- xm:e-end -->
//
//	<!-- xm:emeta id=1 -->
//
//	```json
//	{"aliases":["越哥"]}
//	```
//
//	<!-- xm:emeta-end -->
//
//	## 关系
//
//	### 林越 → 老周 · 师徒
//
//	<!-- xm:r id=1 src=1 dst=2 type=师徒 conf=0.8 chapters="5" -->
//
//	旧相识
//
//	<!-- xm:r-end -->
//
// 三处刻意的取舍：
//  1. **描述进块体、元数据进 `xm:emeta` 块**：元数据是任意 JSON，塞进标记属性会又长又容易踩转义；
//     而描述是给人读的正文，留在块体里最自然。
//  2. **证据单独一块**：证据常是成段的原文引用，含换行，属性放不下。
//  3. **关系端点用实体 id**：改标题不会断链，这也是"合并/拆分"能保持一致的前提。

// GraphEntityFile 是文件里的一个实体。
type GraphEntityFile struct {
	ID          int
	Name        string
	Type        string
	Description string
	Metadata    string // 原始 JSON 文本
	PrimaryID   int    // 0 表示主词条
	ChapterIDs  []int
}

// GraphRelationFile 是文件里的一条关系。
type GraphRelationFile struct {
	ID           int
	SourceID     int
	TargetID     int
	RelationType string
	Description  string
	Evidence     string
	Confidence   float64
	ChapterIDs   []int
}

// GraphDoc 是一份图谱版本文件解析后的内容。
type GraphDoc struct {
	VersionID int
	BookID    int
	Name      string
	Scope     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []GraphEntityFile
	Relations []GraphRelationFile
	Extra     [][2]string
	RelPath   string
}

// graphFrontKeys 是程序负责的 front-matter 键。
var graphFrontKeys = []string{
	"kind", "version_id", "book_id", "name", "scope",
	"chapter_count", "entity_count", "relation_count", "created_at", "updated_at",
}

// ---------------- 路径 ----------------

// GraphFilePath 返回图谱版本文件的绝对路径与相对书目录的路径。
func (s *Service) GraphFilePath(versionID int) (string, string, error) {
	var bookID int
	var name, rel, createdRaw string
	err := s.DB.QueryRow(`SELECT book_id, COALESCE(name,''), COALESCE(file_path,''), COALESCE(created_at,'')
		FROM graph_versions WHERE id=?`, versionID).Scan(&bookID, &name, &rel, &createdRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("图谱版本不存在")
		}
		return "", "", err
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(rel) == "" {
		rel, err = s.defaultGraphRelPath(bookID, versionID, name, createdRaw)
		if err != nil {
			return "", "", err
		}
		if _, err := s.DB.Exec(`UPDATE graph_versions SET file_path=? WHERE id=?`, rel, versionID); err != nil {
			return "", "", err
		}
	}
	abs, err := JoinWithin(bookDir, rel)
	if err != nil {
		return "", "", err
	}
	if err := ensureDir(filepath.Dir(abs)); err != nil {
		return "", "", err
	}
	return abs, rel, nil
}

// defaultGraphRelPath 生成 `memory/graph/<版本号>-<版本名>.md`。
//
// 版本数少、且用户会用"版本名"来认它，所以这里用可读名而不是短码；
// 重名时加序号，保证不与已有文件撞。
func (s *Service) defaultGraphRelPath(bookID, versionID int, name, createdRaw string) (string, error) {
	dir, err := MemoryRelDir(MemoryKindGraph, ParseDBTime(createdRaw))
	if err != nil {
		return "", err
	}
	base := fmt.Sprintf("%d-%s.md", versionID, SanitizeName(TruncateRunes(name, memMaxTitleRunes)))
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	absDir, err := JoinWithin(bookDir, dir)
	if err != nil {
		return "", err
	}
	if err := ensureDir(absDir); err != nil {
		return "", err
	}
	name = UniqueName(absDir, base)
	return filepath.ToSlash(filepath.Join(dir, name)), nil
}

// ---------------- 渲染 ----------------

// RenderGraphDoc 把图谱版本渲染成文件内容。
func RenderGraphDoc(d *GraphDoc) string {
	known := [][2]string{
		{"kind", MemoryKindGraph},
		{"version_id", strconv.Itoa(d.VersionID)},
		{"book_id", strconv.Itoa(d.BookID)},
		{"name", d.Name},
		{"scope", d.Scope},
		{"chapter_count", strconv.Itoa(countGraphChapters(d))},
		{"entity_count", strconv.Itoa(len(d.Entities))},
		{"relation_count", strconv.Itoa(len(d.Relations))},
		{"created_at", FormatMemTime(d.CreatedAt)},
		{"updated_at", FormatMemTime(d.UpdatedAt)},
	}
	var b strings.Builder
	b.WriteString(RenderMemoryFrontMatter(known, d.Extra))
	b.WriteString("\n")
	b.WriteString(RenderMarkerLine("graph", [][2]string{{"v", "1"}}))
	b.WriteString("\n\n")

	name := strings.TrimSpace(d.Name)
	if name == "" {
		name = "未命名版本"
	}
	b.WriteString(fmt.Sprintf("# %s（%d 实体 / %d 关系）\n\n", name, len(d.Entities), len(d.Relations)))
	b.WriteString("<!-- 自动生成，请勿手改标记；正文可以随便改。 -->\n\n")

	// 按类型分组：这是人浏览图谱最自然的方式，也便于对照界面上的图例。
	byType := map[string][]GraphEntityFile{}
	var types []string
	for _, e := range d.Entities {
		t := strings.TrimSpace(e.Type)
		if t == "" {
			t = "未分类"
		}
		if _, ok := byType[t]; !ok {
			types = append(types, t)
		}
		byType[t] = append(byType[t], e)
	}
	sort.Strings(types)
	for _, t := range types {
		list := byType[t]
		b.WriteString(fmt.Sprintf("## %s（%d）\n\n", t, len(list)))
		for _, e := range list {
			title := e.Name
			if title == "" {
				title = fmt.Sprintf("实体 #%d", e.ID)
			}
			if e.PrimaryID > 0 {
				title += fmt.Sprintf("（已合并到 #%d）", e.PrimaryID)
			}
			b.WriteString("### " + title + "\n\n")
			attrs := [][2]string{{"id", strconv.Itoa(e.ID)}, {"type", e.Type}}
			if e.PrimaryID > 0 {
				attrs = append(attrs, [2]string{"primary", strconv.Itoa(e.PrimaryID)})
			}
			if len(e.ChapterIDs) > 0 {
				attrs = append(attrs, [2]string{"chapters", FormatIntListAttr(e.ChapterIDs)})
			}
			b.WriteString(RenderPairedMarker("e", attrs, e.Description))
			meta := strings.TrimSpace(e.Metadata)
			if meta != "" && meta != "{}" && meta != "null" {
				b.WriteString(RenderPairedMarker("emeta", [][2]string{{"id", strconv.Itoa(e.ID)}},
					"```json\n"+meta+"\n```"))
			}
		}
	}

	if len(d.Relations) > 0 {
		b.WriteString("## 关系\n\n")
		names := map[int]string{}
		for _, e := range d.Entities {
			names[e.ID] = e.Name
		}
		label := func(id int) string {
			if n := strings.TrimSpace(names[id]); n != "" {
				return n
			}
			return fmt.Sprintf("#%d", id)
		}
		for _, r := range d.Relations {
			relType := strings.TrimSpace(r.RelationType)
			if relType == "" {
				relType = "关联"
			}
			b.WriteString(fmt.Sprintf("### %s → %s · %s\n\n", label(r.SourceID), label(r.TargetID), relType))
			attrs := [][2]string{
				{"id", strconv.Itoa(r.ID)},
				{"src", strconv.Itoa(r.SourceID)},
				{"dst", strconv.Itoa(r.TargetID)},
				{"type", r.RelationType},
			}
			if r.Confidence != 0 {
				attrs = append(attrs, [2]string{"conf", strconv.FormatFloat(r.Confidence, 'g', -1, 64)})
			}
			if len(r.ChapterIDs) > 0 {
				attrs = append(attrs, [2]string{"chapters", FormatIntListAttr(r.ChapterIDs)})
			}
			b.WriteString(RenderPairedMarker("r", attrs, r.Description))
			if strings.TrimSpace(r.Evidence) != "" {
				b.WriteString(RenderPairedMarker("revidence", [][2]string{{"id", strconv.Itoa(r.ID)}}, r.Evidence))
			}
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func countGraphChapters(d *GraphDoc) int {
	seen := map[int]bool{}
	for _, e := range d.Entities {
		for _, c := range e.ChapterIDs {
			seen[c] = true
		}
	}
	for _, r := range d.Relations {
		for _, c := range r.ChapterIDs {
			seen[c] = true
		}
	}
	return len(seen)
}

// ---------------- 解析 ----------------

// ParseGraphDoc 解析图谱版本文件。
func ParseGraphDoc(text string) (*GraphDoc, error) {
	fm, err := ParseMemoryFrontMatter(text)
	if err != nil {
		return nil, err
	}
	pairs := fm.Pairs()
	vals := KnownFields(pairs, graphFrontKeys...)
	d := &GraphDoc{
		Extra:   ExtraFields(pairs, graphFrontKeys...),
		Name:    vals["name"],
		Scope:   vals["scope"],
		RelPath: "",
	}
	if n, err := strconvAtoiSafe(vals["version_id"]); err == nil {
		d.VersionID = n
	}
	if n, err := strconvAtoiSafe(vals["book_id"]); err == nil {
		d.BookID = n
	}
	d.CreatedAt = ParseMemTime(vals["created_at"])
	d.UpdatedAt = ParseMemTime(vals["updated_at"])

	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		return nil, err
	}

	// 元数据与证据都按 id 关联到实体/关系（标记平铺，没有嵌套）
	metas := map[int]string{}
	for _, mk := range FindMarkers(markers, "emeta") {
		metas[mk.IntAttr("id")] = fencedJSON(mk.Body)
	}
	evidences := map[int]string{}
	for _, mk := range FindMarkers(markers, "revidence") {
		evidences[mk.IntAttr("id")] = mk.Body
	}

	for _, mk := range FindMarkers(markers, "e") {
		e := GraphEntityFile{
			ID:          mk.IntAttr("id"),
			Type:        mk.Attr("type"),
			Description: mk.Body,
			Metadata:    metas[mk.IntAttr("id")],
			PrimaryID:   mk.IntAttr("primary"),
			ChapterIDs:  ParseIntList(mk.Attr("chapters")),
		}
		// 名字在装饰性的 `###` 行里，而装饰行不参与解析 —— 所以名字只能从别处来。
		// 这里用"标记前的最近一个标题"来补，保持文件可读的同时不牺牲还原能力。
		e.Name = nearestHeadingBefore(text, mk.Line, 3)
		d.Entities = append(d.Entities, e)
	}
	for _, mk := range FindMarkers(markers, "r") {
		conf, _ := strconv.ParseFloat(strings.TrimSpace(mk.Attr("conf")), 64)
		d.Relations = append(d.Relations, GraphRelationFile{
			ID:           mk.IntAttr("id"),
			SourceID:     mk.IntAttr("src"),
			TargetID:     mk.IntAttr("dst"),
			RelationType: mk.Attr("type"),
			Description:  mk.Body,
			Evidence:     evidences[mk.IntAttr("id")],
			Confidence:   conf,
			ChapterIDs:   ParseIntList(mk.Attr("chapters")),
		})
	}
	return d, nil
}

// nearestHeadingBefore 找某个标记之前最近的 `###` 标题文本（去掉"（已合并到 #N）"后缀）。
//
// 这是"让装饰行也承担一点信息"的折中：实体名写在 `###` 标题里最自然，
// 但装饰行本来是不解析的。与其把名字塞进又长又难读的属性，不如在解析时向上找最近的标题。
func nearestHeadingBefore(text string, line, level int) string {
	if line <= 1 {
		return ""
	}
	lines := strings.Split(NormalizeText(text), "\n")
	prefix := strings.Repeat("#", level) + " "
	for i := line - 2; i >= 0 && i < len(lines); i-- {
		raw := strings.TrimSpace(lines[i])
		if strings.HasPrefix(raw, prefix) && !strings.HasPrefix(raw, prefix+"#") {
			name := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
			if i := strings.LastIndex(name, "（已合并到 #"); i > 0 {
				name = strings.TrimSpace(name[:i])
			}
			return name
		}
		if strings.HasPrefix(raw, "# ") || strings.HasPrefix(raw, "## ") {
			return ""
		}
	}
	return ""
}

// ---------------- 从库渲染 ----------------

// MirrorGraphVersion 从库渲染某个版本并落盘。
func (s *Service) MirrorGraphVersion(versionID int) error {
	var d GraphDoc
	var createdRaw, updatedRaw string
	err := s.DB.QueryRow(`SELECT id, book_id, COALESCE(name,''), COALESCE(scope,''),
		COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM graph_versions WHERE id=?`, versionID).
		Scan(&d.VersionID, &d.BookID, &d.Name, &d.Scope, &createdRaw, &updatedRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // 版本已删除
		}
		return err
	}
	d.CreatedAt = ParseDBTime(createdRaw)
	d.UpdatedAt = ParseDBTime(updatedRaw)

	// 章节引用一次性取出，按实体/关系分组
	entityChapters := map[int][]int{}
	relationChapters := map[int][]int{}
	rows, err := s.DB.Query(`SELECT COALESCE(entity_id,0), COALESCE(relation_id,0), chapter_id
		FROM graph_chapter_refs WHERE version_id=? ORDER BY chapter_id ASC`, versionID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var eid, rid, cid int
		if err := rows.Scan(&eid, &rid, &cid); err != nil {
			rows.Close()
			return fmt.Errorf("读取章节引用失败（%v）—— 不写出残缺文件", err)
		}
		if eid > 0 {
			entityChapters[eid] = append(entityChapters[eid], cid)
		}
		if rid > 0 {
			relationChapters[rid] = append(relationChapters[rid], cid)
		}
	}
	rows.Close()

	erows, err := s.DB.Query(`SELECT id, COALESCE(name,''), COALESCE(type,''), COALESCE(description,''),
		COALESCE(metadata,'{}'), COALESCE(primary_id,0)
		FROM graph_entities WHERE version_id=? ORDER BY id ASC`, versionID)
	if err != nil {
		return err
	}
	for erows.Next() {
		var e GraphEntityFile
		if err := erows.Scan(&e.ID, &e.Name, &e.Type, &e.Description, &e.Metadata, &e.PrimaryID); err != nil {
			erows.Close()
			return fmt.Errorf("读取图谱实体失败（%v）—— 不写出残缺文件", err)
		}
		e.ChapterIDs = entityChapters[e.ID]
		d.Entities = append(d.Entities, e)
	}
	erows.Close()

	rrows, err := s.DB.Query(`SELECT id, source_id, target_id, COALESCE(relation_type,''),
		COALESCE(description,''), COALESCE(evidence,''), COALESCE(confidence,0)
		FROM graph_relations WHERE version_id=? ORDER BY id ASC`, versionID)
	if err != nil {
		return err
	}
	for rrows.Next() {
		var r GraphRelationFile
		if err := rrows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.Description, &r.Evidence, &r.Confidence); err != nil {
			rrows.Close()
			return fmt.Errorf("读取图谱关系失败（%v）—— 不写出残缺文件", err)
		}
		r.ChapterIDs = relationChapters[r.ID]
		d.Relations = append(d.Relations, r)
	}
	rrows.Close()

	abs, rel, err := s.GraphFilePath(versionID)
	if err != nil {
		return err
	}

	// 文件是真相，程序不覆盖被外部改过的文件 —— 与对话、Agent 会话同一规则。
	var recorded string
	s.DB.QueryRow(`SELECT COALESCE(content_hash,'') FROM graph_versions WHERE id=?`, versionID).Scan(&recorded)
	if changed, _, herr := FileEditedExternally(abs, recorded); herr == nil && changed {
		s.DB.Exec(`UPDATE graph_versions SET parse_error=? WHERE id=?`, ExternalEditReason, versionID)
		return nil
	}

	// 版本改名后把文件也改名，让"文件名 = 版本名"这个直觉始终成立。
	//
	// 不做的话会出现**不确定的行为**：如果改名发生在首次落盘之前，文件名会是新的；
	// 如果发生在之后，文件名会停在旧的。同一个操作因为时序不同而结果不同，是最难解释的一类问题。
	if bookDir, derr := s.BookDir(d.BookID); derr == nil {
		want := fmt.Sprintf("%d-%s.md", versionID, SanitizeName(TruncateRunes(d.Name, memMaxTitleRunes)))
		cur := filepath.Base(filepath.FromSlash(rel))
		if cur != want {
			newRel := filepath.ToSlash(filepath.Join(filepath.Dir(filepath.FromSlash(rel)), want))
			if newAbs, jerr := JoinWithin(bookDir, newRel); jerr == nil {
				if _, serr := os.Stat(newAbs); os.IsNotExist(serr) {
					if rerr := renameWithRetry(abs, newAbs); rerr == nil {
						abs, rel = newAbs, newRel
						s.DB.Exec(`UPDATE graph_versions SET file_path=? WHERE id=?`, rel, versionID)
					}
				}
			}
		}
	}
	d.RelPath = rel
	if raw, rerr := os.ReadFile(abs); rerr == nil {
		if old, perr := ParseMemoryFrontMatter(string(raw)); perr == nil {
			d.Extra = ExtraFields(old.Pairs(), graphFrontKeys...)
		}
	}
	meta, err := WriteTextFile(abs, RenderGraphDoc(&d))
	if err != nil {
		return err
	}
	s.DB.Exec(`UPDATE graph_versions SET file_path=?, content_hash=?, missing=0, parse_error='',
		synced_at=CURRENT_TIMESTAMP WHERE id=?`, rel, meta.Hash, versionID)
	return nil
}

// TrashGraphFile 把版本文件移入回收站。
func (s *Service) TrashGraphFile(versionID int) error {
	var bookID int
	var rel string
	if err := s.DB.QueryRow(`SELECT book_id, COALESCE(file_path,'') FROM graph_versions WHERE id=?`, versionID).
		Scan(&bookID, &rel); err != nil {
		return nil
	}
	if strings.TrimSpace(rel) == "" {
		return nil
	}
	return s.trashBookFile(bookID, rel)
}

// GraphDocToEntities 把解析结果转成可直接喂给 ImportGraph 的结构。
//
// 这是"导出与导入同一套格式"的落点：文件里的 id 是**文件内的 id**，
// 导入时由 ImportGraph 重新分配，端点通过 src/dst 的对应关系重定向。
func GraphDocToEntities(d *GraphDoc) ([]json.RawMessage, error) {
	type entityJSON struct {
		ID          int             `json:"id"`
		Name        string          `json:"name"`
		Type        string          `json:"type"`
		Description string          `json:"description"`
		Metadata    json.RawMessage `json:"metadata,omitempty"`
		PrimaryID   *int            `json:"primary_id,omitempty"`
		ChapterIDs  []int           `json:"chapter_ids,omitempty"`
	}
	type relationJSON struct {
		ID           int     `json:"id"`
		SourceID     int     `json:"source_id"`
		TargetID     int     `json:"target_id"`
		RelationType string  `json:"relation_type"`
		Description  string  `json:"description"`
		Evidence     string  `json:"evidence"`
		Confidence   float64 `json:"confidence"`
		ChapterIDs   []int   `json:"chapter_ids,omitempty"`
	}
	var out []json.RawMessage
	for _, e := range d.Entities {
		obj := entityJSON{ID: e.ID, Name: e.Name, Type: e.Type, Description: e.Description, ChapterIDs: e.ChapterIDs}
		if m := strings.TrimSpace(e.Metadata); m != "" {
			obj.Metadata = json.RawMessage(m)
		}
		if e.PrimaryID > 0 {
			p := e.PrimaryID
			obj.PrimaryID = &p
		}
		raw, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		out = append(out, raw)
	}
	for _, r := range d.Relations {
		raw, err := json.Marshal(relationJSON{
			ID: r.ID, SourceID: r.SourceID, TargetID: r.TargetID, RelationType: r.RelationType,
			Description: r.Description, Evidence: r.Evidence, Confidence: r.Confidence,
			ChapterIDs: r.ChapterIDs,
		})
		if err != nil {
			return nil, err
		}
		out = append(out, raw)
	}
	return out, nil
}

// ---------------- 未分版本的实体（历史数据） ----------------

// unversionedGraphName 是"没有版本号"的实体/关系所用的固定文件名。
//
// 前端图谱页专门有 legacyData 回退逻辑处理 version_id 为空的行，
// 说明这类数据是真实存在的。既然文件是真相，就不能把它们排除在外，
// 否则"指向目录就能恢复"会静默少一块。
const unversionedGraphName = "0-未分版本.md"

// GraphDirtyKey 生成 flusher 键：有版本用版本 id，没有版本用 "book:<书 id>"。
func GraphDirtyKey(versionID, bookID int) string {
	if versionID > 0 {
		return strconv.Itoa(versionID)
	}
	return "book:" + strconv.Itoa(bookID)
}

// MirrorGraphByKey 是注册给合并写入器的入口，兼容两种 key。
func (s *Service) MirrorGraphByKey(key string) error {
	if rest, ok := strings.CutPrefix(key, "book:"); ok {
		bookID, err := strconv.Atoi(strings.TrimSpace(rest))
		if err != nil {
			return nil
		}
		return s.MirrorUnversionedGraph(bookID)
	}
	versionID, err := strconv.Atoi(strings.TrimSpace(key))
	if err != nil || versionID <= 0 {
		return nil
	}
	return s.MirrorGraphVersion(versionID)
}

// UnversionedGraphPath 返回未分版本图谱文件的路径（固定名，可重复定位）。
func (s *Service) UnversionedGraphPath(bookID int) (string, string, error) {
	dir, err := MemoryRelDir(MemoryKindGraph, time.Time{})
	if err != nil {
		return "", "", err
	}
	rel := filepath.ToSlash(filepath.Join(dir, unversionedGraphName))
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", "", err
	}
	abs, err := JoinWithin(bookDir, rel)
	if err != nil {
		return "", "", err
	}
	if err := ensureDir(filepath.Dir(abs)); err != nil {
		return "", "", err
	}
	return abs, rel, nil
}

// MirrorUnversionedGraph 渲染某本书里"没有版本号"的实体与关系。
func (s *Service) MirrorUnversionedGraph(bookID int) error {
	var d GraphDoc
	var createdRaw string
	if err := s.DB.QueryRow(`SELECT COALESCE(MIN(created_at),'') FROM graph_entities WHERE book_id=? AND version_id IS NULL`, bookID).
		Scan(&createdRaw); err != nil {
		return err
	}
	d.BookID = bookID
	d.Name = "未分版本"
	d.Scope = "无版本（历史数据）"
	d.CreatedAt = ParseDBTime(createdRaw)
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	d.UpdatedAt = time.Now()

	entityChapters := map[int][]int{}
	relationChapters := map[int][]int{}
	rows, err := s.DB.Query(`SELECT COALESCE(entity_id,0), COALESCE(relation_id,0), chapter_id
		FROM graph_chapter_refs WHERE book_id=? AND version_id IS NULL ORDER BY chapter_id ASC`, bookID)
	if err != nil {
		return err
	}
	for rows.Next() {
		var eid, rid, cid int
		if err := rows.Scan(&eid, &rid, &cid); err != nil {
			rows.Close()
			return fmt.Errorf("读取章节引用失败（%v）—— 不写出残缺文件", err)
		}
		if eid > 0 {
			entityChapters[eid] = append(entityChapters[eid], cid)
		}
		if rid > 0 {
			relationChapters[rid] = append(relationChapters[rid], cid)
		}
	}
	rows.Close()

	erows, err := s.DB.Query(`SELECT id, COALESCE(name,''), COALESCE(type,''), COALESCE(description,''),
		COALESCE(metadata,'{}'), COALESCE(primary_id,0)
		FROM graph_entities WHERE book_id=? AND version_id IS NULL ORDER BY id ASC`, bookID)
	if err != nil {
		return err
	}
	for erows.Next() {
		var e GraphEntityFile
		if err := erows.Scan(&e.ID, &e.Name, &e.Type, &e.Description, &e.Metadata, &e.PrimaryID); err != nil {
			erows.Close()
			return fmt.Errorf("读取图谱实体失败（%v）—— 不写出残缺文件", err)
		}
		e.ChapterIDs = entityChapters[e.ID]
		d.Entities = append(d.Entities, e)
	}
	erows.Close()

	rrows, err := s.DB.Query(`SELECT id, source_id, target_id, COALESCE(relation_type,''),
		COALESCE(description,''), COALESCE(evidence,''), COALESCE(confidence,0)
		FROM graph_relations WHERE book_id=? AND version_id IS NULL ORDER BY id ASC`, bookID)
	if err != nil {
		return err
	}
	for rrows.Next() {
		var r GraphRelationFile
		if err := rrows.Scan(&r.ID, &r.SourceID, &r.TargetID, &r.RelationType, &r.Description, &r.Evidence, &r.Confidence); err != nil {
			rrows.Close()
			return fmt.Errorf("读取图谱关系失败（%v）—— 不写出残缺文件", err)
		}
		r.ChapterIDs = relationChapters[r.ID]
		d.Relations = append(d.Relations, r)
	}
	rrows.Close()

	abs, rel, err := s.UnversionedGraphPath(bookID)
	if err != nil {
		return err
	}
	// 一条都没有：把可能存在的旧文件收进回收站，避免留下会误导人的空壳
	if len(d.Entities) == 0 && len(d.Relations) == 0 {
		if _, statErr := os.Stat(abs); statErr == nil {
			return s.trashBookFile(bookID, rel)
		}
		return nil
	}
	d.RelPath = rel
	if _, err := WriteTextFile(abs, RenderGraphDoc(&d)); err != nil {
		return err
	}
	return nil
}
