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
)

// 拆书与生成历史的文件存储（docs/books-memory-storage.md §5.4 / §5.5）。
//
// 形态（单文件 Markdown，机器可解析）：
//
//	---
//	kind: analysis
//	book_id: 22
//	external_id: "1729262400000"
//	title: 黄金三章拆解
//	source: 拆书
//	status: completed
//	prompt: 黄金三章拆解
//	chapters: 1,2,3
//	created_at: "2026-10-18 22:45:00"
//	updated_at: "2026-10-18 22:45:00"
//	---
//
//	<!-- xm:record v=1 -->
//
//	# 黄金三章拆解
//
//	## 结果
//
//	<!-- xm:result -->
//
//	（主输出，人读）
//
//	<!-- xm:result-end -->
//
//	## 输入与提示词
//
//	<!-- xm:input -->
//
//	（可读展开）
//
//	<!-- xm:input-end -->
//
//	<!-- xm:payload -->
//
//	```json
//	{…原始记录 JSON…}
//	```
//
//	<!-- xm:payload-end -->
//
// 为什么要有 xm:payload：拆书、续写、改写、创作四类历史的字段结构各不相同
// （messages / fieldValues / relatedContext / selectedPrompts …）。
// 把原记录 JSON 原样存一份，比给每种结构写一套 front-matter 映射更可靠 ——
// 后者每加一个字段就要改一次渲染与解析，且漏字段是静默的。
// 前面那些可读段落是给人看的，程序读的是 payload；payload 被删掉时退化为
// "front-matter + 结果正文"，仍能显示，只是丢掉了结构化细节。

// MemoryRecordConflict 表示文件被程序之外改过，本次写入被拒绝。
//
// 记录（拆书/生成历史）的写操作来自界面上的明确动作，所以这里返回错误让界面如实提示，
// 而不是像后台渲染那样静默跳过 —— 用户点了保存却没保存上，必须当场知道。
var MemoryRecordConflict = errors.New("memory_record_conflict")

// recordFrontKeys 是程序负责的 front-matter 键，其余视为用户自定义字段保留。
var recordFrontKeys = []string{
	"kind", "book_id", "external_id", "title", "source", "status",
	"prompt", "chapters", "created_at", "updated_at",
}

// MemoryRecord 是 memory_records 一行的索引视图（不含正文）。
type MemoryRecord struct {
	ID         int    `json:"id"`
	BookID     int    `json:"book_id"`
	Kind       string `json:"kind"`
	ExternalID string `json:"external_id"`
	Title      string `json:"title"`
	Source     string `json:"source"`
	Status     string `json:"status"`
	FilePath   string `json:"file_path"`
	WordCount  int    `json:"word_count"`
	Missing    int    `json:"missing"`
	ParseError string `json:"parse_error,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`

	// BookTitle 来自 JOIN books：历史中心要按作品筛选，而记录文件里不一定写了书名。
	BookTitle string `json:"book_title,omitempty"`

	// 以下三项仅在 full=1 的列表里出现（要读文件才有）。
	Result   string `json:"result,omitempty"`
	Readable string `json:"readable,omitempty"`
	Payload  string `json:"payload,omitempty"`
	Prompt   string `json:"prompt,omitempty"`
	Chapters []int  `json:"chapters,omitempty"`
}

// MemoryRecordInput 是新建 / 更新时提交的内容。
type MemoryRecordInput struct {
	BookID     int
	Kind       string
	ExternalID string
	Title      string
	Source     string
	Status     string
	Prompt     string
	Chapters   []int
	CreatedAt  time.Time
	Result     string // 主输出（人读）
	Readable   string // 「输入与提示词」的可读展开
	Payload    string // 原始记录 JSON（无损）
}

// MemoryRecordDoc 是解析出来的完整记录（含正文与原始 JSON）。
type MemoryRecordDoc struct {
	MemoryRecordInput
	RelPath string
	Extra   [][2]string // 用户自定义 front-matter 键
}

// ---------------- 路径 ----------------

// RecordPath 返回某条记录的文件路径；file_path 为空时按创建时间与标题现算并写回索引。
func (s *Service) RecordPath(recordID int) (string, string, error) {
	var bookID int
	var kind, title, rel, createdRaw string
	err := s.DB.QueryRow(`SELECT book_id, kind, COALESCE(title,''), COALESCE(file_path,''), COALESCE(created_at,'')
		FROM memory_records WHERE id=?`, recordID).Scan(&bookID, &kind, &title, &rel, &createdRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("记录不存在")
		}
		return "", "", err
	}
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(rel) == "" {
		at := ParseDBTime(createdRaw)
		if at.IsZero() {
			at = time.Now()
		}
		rel, err = MemoryRelPath(kind, at, title, recordID)
		if err != nil {
			return "", "", err
		}
		if _, err := s.DB.Exec(`UPDATE memory_records SET file_path=? WHERE id=?`, rel, recordID); err != nil {
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

// ---------------- 渲染与解析 ----------------

// RenderRecordDoc 把记录渲染成文件内容。
func RenderRecordDoc(d *MemoryRecordDoc) string {
	var b strings.Builder
	b.WriteString("---\n")
	pairs := [][2]string{
		{"kind", memoryFileKind(d.Kind)},
		{"book_id", itoaStr(d.BookID)},
	}
	if d.ExternalID != "" {
		pairs = append(pairs, [2]string{"external_id", d.ExternalID})
	}
	pairs = append(pairs,
		[2]string{"title", d.Title},
		[2]string{"source", d.Source},
		[2]string{"status", d.Status},
	)
	if d.Prompt != "" {
		pairs = append(pairs, [2]string{"prompt", d.Prompt})
	}
	if len(d.Chapters) > 0 {
		pairs = append(pairs, [2]string{"chapters", FormatIntListAttr(d.Chapters)})
	}
	at := d.CreatedAt
	if at.IsZero() {
		at = time.Now()
	}
	pairs = append(pairs,
		[2]string{"created_at", FormatMemTime(at)},
		[2]string{"updated_at", FormatMemTime(time.Now())},
	)
	for _, kv := range pairs {
		b.WriteString(kv[0])
		b.WriteString(": ")
		b.WriteString(QuoteFMValue(kv[1]))
		b.WriteString("\n")
	}
	for _, kv := range d.Extra {
		if kv[0] == "" || kv[1] == "" {
			continue
		}
		b.WriteString(kv[0])
		b.WriteString(": ")
		b.WriteString(QuoteFMValue(kv[1]))
		b.WriteString("\n")
	}
	b.WriteString("---\n\n")

	b.WriteString(RenderMarkerLine("record", [][2]string{{"v", "1"}}))
	b.WriteString("\n\n# ")
	b.WriteString(strings.TrimSpace(d.Title))
	b.WriteString("\n")

	appendHeading(&b, "## 结果")
	b.WriteString(RenderPairedMarker("result", nil, d.Result))

	appendHeading(&b, "## 输入与提示词")
	b.WriteString(RenderPairedMarker("input", nil, d.Readable))

	if strings.TrimSpace(d.Payload) != "" {
		body := "```json\n" + strings.TrimSpace(d.Payload) + "\n```"
		appendHeading(&b, "## 原始记录")
		b.WriteString(RenderPairedMarker("payload", nil, body))
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

// ParseRecordDoc 解析记录文件。
//
// payload 缺失/损坏不算失败：退化成"front-matter + 结果正文"，能显示就先给用户看到。
func ParseRecordDoc(text string) (*MemoryRecordDoc, error) {
	fm, err := ParseMemoryFrontMatter(text)
	if err != nil {
		return nil, err
	}
	markers, err := ParseMemoryMarkers(text)
	if err != nil {
		return nil, err
	}
	pairs := fm.Pairs()
	d := &MemoryRecordDoc{
		MemoryRecordInput: MemoryRecordInput{
			BookID:     intFromStr(fm.Get("book_id")),
			Kind:       fm.Get("kind"),
			ExternalID: fm.Get("external_id"),
			Title:      fm.Get("title"),
			Source:     fm.Get("source"),
			Status:     fm.Get("status"),
			Prompt:     fm.Get("prompt"),
			Chapters:   ParseIntList(fm.Get("chapters")),
		},
		Extra: ExtraFields(pairs, recordFrontKeys...),
	}
	if at := ParseMemTime(fm.Get("created_at")); !at.IsZero() {
		d.CreatedAt = at
	}
	if m, ok := FindMarker(markers, "result"); ok {
		d.Result = m.Body
	}
	if m, ok := FindMarker(markers, "input"); ok {
		d.Readable = m.Body
	}
	if m, ok := FindMarker(markers, "payload"); ok {
		d.Payload = extractJSONFence(m.Body)
	}
	return d, nil
}

// extractJSONFence 从 ```json 代码块里取出原始 JSON。
func extractJSONFence(body string) string {
	body = NormalizeText(body)
	start := strings.Index(body, "```json")
	if start < 0 {
		return ""
	}
	rest := body[start+len("```json"):]
	end := strings.Index(rest, "```")
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(rest[:end])
}

// ---------------- 业务操作 ----------------

// CreateRecord 写文件 + 建索引；写文件失败则不留下索引行。
func (s *Service) CreateRecord(bookID int, in MemoryRecordInput) (*MemoryRecord, error) {
	if err := validRecordKind(in.Kind); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Title) == "" {
		in.Title = "未命名"
	}
	if strings.TrimSpace(in.Status) == "" {
		in.Status = "completed"
	}
	if in.CreatedAt.IsZero() {
		in.CreatedAt = time.Now()
	}
	// 先插索引拿到 id：文件名里的短码要用它
	res, err := s.DB.Exec(`INSERT INTO memory_records (book_id, kind, external_id, title, source, status, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,CURRENT_TIMESTAMP)`,
		bookID, in.Kind, in.ExternalID, in.Title, in.Source, in.Status,
		in.CreatedAt.Format(memTimeLayout))
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	id := int(id64)

	rec, err := s.writeRecordFile(id, bookID, in, "")
	if err != nil {
		s.DB.Exec(`DELETE FROM memory_records WHERE id=?`, id)
		return nil, err
	}
	return rec, nil
}

// UpdateRecord 覆盖记录内容；改标题会连文件一起改名。
func (s *Service) UpdateRecord(id int, in MemoryRecordInput) (*MemoryRecord, error) {
	var bookID int
	var kind, oldRel, oldHash string
	err := s.DB.QueryRow(`SELECT book_id, kind, COALESCE(file_path,''), COALESCE(content_hash,'')
		FROM memory_records WHERE id=?`, id).Scan(&bookID, &kind, &oldRel, &oldHash)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Kind) == "" {
		in.Kind = kind
	}
	if err := validRecordKind(in.Kind); err != nil {
		return nil, err
	}

	// 文件被程序之外改过就不覆盖 —— 与对话、Agent 会话、图谱同一条规则。
	// 区别在于这里把冲突报出来：用户点的是"保存"，静默跳过等于骗他。
	if abs, _, perr := s.RecordPath(id); perr == nil {
		if changed, _, herr := FileEditedExternally(abs, oldHash); herr == nil && changed {
			s.DB.Exec(`UPDATE memory_records SET parse_error=? WHERE id=?`, ExternalEditReason, id)
			return nil, MemoryRecordConflict
		}
	}

	if in.CreatedAt.IsZero() {
		var createdRaw string
		s.DB.QueryRow(`SELECT COALESCE(created_at,'') FROM memory_records WHERE id=?`, id).Scan(&createdRaw)
		in.CreatedAt = ParseDBTime(createdRaw)
		if in.CreatedAt.IsZero() {
			in.CreatedAt = time.Now()
		}
	}
	rec, err := s.writeRecordFile(id, bookID, in, oldRel)
	if err != nil {
		return nil, err
	}
	_, err = s.DB.Exec(`UPDATE memory_records SET kind=?, external_id=?, title=?, source=?, status=?,
		created_at=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		in.Kind, in.ExternalID, in.Title, in.Source, in.Status,
		in.CreatedAt.Format(memTimeLayout), id)
	if err != nil {
		return nil, err
	}
	return rec, nil
}

// writeRecordFile 把内容写到文件，并回填索引里的路径、哈希、字数。
func (s *Service) writeRecordFile(id, bookID int, in MemoryRecordInput, oldRel string) (*MemoryRecord, error) {
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return nil, err
	}
	rel := oldRel
	if strings.TrimSpace(rel) == "" {
		rel, err = MemoryRelPath(in.Kind, in.CreatedAt, in.Title, id)
		if err != nil {
			return nil, err
		}
	} else if want, changed := ReconcileMemoryFileName(rel, in.Title); changed {
		rel = want
	}
	abs, err := JoinWithin(bookDir, rel)
	if err != nil {
		return nil, err
	}
	if err := ensureDir(filepath.Dir(abs)); err != nil {
		return nil, err
	}

	in.BookID = bookID
	doc := &MemoryRecordDoc{
		MemoryRecordInput: in,
		RelPath:           rel,
	}
	// 保留用户自己加进 front-matter 的键
	if raw, rerr := os.ReadFile(abs); rerr == nil {
		if old, perr := ParseMemoryFrontMatter(string(raw)); perr == nil {
			doc.Extra = ExtraFields(old.Pairs(), recordFrontKeys...)
		}
	}
	meta, err := WriteTextFile(abs, RenderRecordDoc(doc))
	if err != nil {
		return nil, err
	}

	// 改名后旧文件要清掉（路径变了，旧名留着就是孤儿）
	if oldRel != "" && filepath.ToSlash(oldRel) != filepath.ToSlash(rel) {
		if oldAbs, jerr := JoinWithin(bookDir, oldRel); jerr == nil {
			_ = os.Remove(oldAbs)
		}
	}
	if _, err := s.DB.Exec(`UPDATE memory_records SET file_path=?, content_hash=?, word_count=?,
		missing=0, parse_error='', synced_at=CURRENT_TIMESTAMP WHERE id=?`,
		rel, meta.Hash, WordCount(in.Result), id); err != nil {
		return nil, err
	}
	return s.getRecord(id)
}

// LoadRecord 读单条记录：正文以文件为准，解析成功顺手回写索引。
func (s *Service) LoadRecord(id int) (*MemoryRecordDoc, error) {
	abs, rel, err := s.RecordPath(id)
	if err != nil {
		return nil, err
	}
	text, _, err := ReadTextFile(abs)
	if err != nil {
		s.DB.Exec(`UPDATE memory_records SET missing=1 WHERE id=?`, id)
		return nil, fmt.Errorf("记录文件不存在: %w", err)
	}
	doc, err := ParseRecordDoc(text)
	if err != nil {
		// 解析失败不覆盖、不猜，如实记录原因
		s.DB.Exec(`UPDATE memory_records SET parse_error=? WHERE id=?`, MemoryParseReason(err), id)
		return nil, err
	}
	doc.RelPath = rel
	s.DB.Exec(`UPDATE memory_records SET missing=0, parse_error='' WHERE id=?`, id)
	return doc, nil
}

// ListRecords 列出某类记录的索引（不读文件：列表只显示标题与时间）。
func (s *Service) ListRecords(bookID int, kind string) ([]MemoryRecord, error) {
	query := `SELECT r.id, r.book_id, r.kind, COALESCE(r.external_id,''), COALESCE(r.title,''), COALESCE(r.source,''),
		COALESCE(r.status,''), COALESCE(r.file_path,''), r.word_count, r.missing, COALESCE(r.parse_error,''),
		COALESCE(r.created_at,''), COALESCE(r.updated_at,''), COALESCE(b.title,'')
		FROM memory_records r LEFT JOIN books b ON b.id = r.book_id WHERE r.book_id=?`
	args := []any{bookID}
	if strings.TrimSpace(kind) != "" {
		query += ` AND r.kind=?`
		args = append(args, kind)
	}
	query += ` ORDER BY r.created_at DESC, r.id DESC`

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MemoryRecord
	for rows.Next() {
		var r MemoryRecord
		if err := rows.Scan(&r.ID, &r.BookID, &r.Kind, &r.ExternalID, &r.Title, &r.Source, &r.Status,
			&r.FilePath, &r.WordCount, &r.Missing, &r.ParseError, &r.CreatedAt, &r.UpdatedAt, &r.BookTitle); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// FindRecordByExternal 按 (书, 类型, 外部 id) 查记录，用于旧数据迁移去重。
func (s *Service) FindRecordByExternal(bookID int, kind, externalID string) (*MemoryRecord, error) {
	var id int
	err := s.DB.QueryRow(`SELECT id FROM memory_records WHERE book_id=? AND kind=? AND external_id=?`,
		bookID, kind, externalID).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.getRecord(id)
}

// TrashRecord 删除记录：文件先进回收站，再删索引。
func (s *Service) TrashRecord(id int) error {
	var bookID int
	var rel string
	err := s.DB.QueryRow(`SELECT book_id, COALESCE(file_path,'') FROM memory_records WHERE id=?`, id).
		Scan(&bookID, &rel)
	if err != nil {
		return err
	}
	if strings.TrimSpace(rel) != "" {
		if err := s.trashBookFile(bookID, rel); err != nil {
			return err
		}
	}
	_, err = s.DB.Exec(`DELETE FROM memory_records WHERE id=?`, id)
	return err
}

// CountRecordsByKind 统计某本书各类型的记录数，供「记忆」概览展示。
func (s *Service) CountRecordsByKind(bookID int) (map[string]int, error) {
	rows, err := s.DB.Query(`SELECT kind, COUNT(*) FROM memory_records WHERE book_id=? GROUP BY kind`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			return nil, err
		}
		out[k] = n
	}
	return out, rows.Err()
}

func (s *Service) getRecord(id int) (*MemoryRecord, error) {
	var r MemoryRecord
	err := s.DB.QueryRow(`SELECT id, book_id, kind, COALESCE(external_id,''), COALESCE(title,''),
		COALESCE(source,''), COALESCE(status,''), COALESCE(file_path,''), word_count, missing,
		COALESCE(parse_error,''), COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM memory_records WHERE id=?`, id).
		Scan(&r.ID, &r.BookID, &r.Kind, &r.ExternalID, &r.Title, &r.Source, &r.Status,
			&r.FilePath, &r.WordCount, &r.Missing, &r.ParseError, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func validRecordKind(kind string) error {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case MemoryKindAnalysis, MemoryKindGeneration:
		return nil
	}
	return fmt.Errorf("记录类型只能是 %s 或 %s，实得 %q", MemoryKindAnalysis, MemoryKindGeneration, kind)
}

func intFromStr(s string) int {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, `"`)
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return n
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func itoaStr(v int) string { return fmt.Sprintf("%d", v) }

// appendHeading 写一个标题，并保证它前面**恰好**有一个空行。
//
// 成对标记渲染出来时结尾自带空行，标题若直接跟在后面，Markdown 会把它当成
// 上一段的懒延续 —— 在 Obsidian 里看起来像被吞进了上一块。所以先把尾部空行
// 全部剥掉、再明确补一个：无论调用方之前写了多少换行，结果都稳定。
func appendHeading(b *strings.Builder, heading string) {
	trimmed := strings.TrimRight(b.String(), "\n")
	b.Reset()
	b.WriteString(trimmed)
	b.WriteString("\n\n")
	b.WriteString(heading)
	// 标题后也留一个空行，否则标记行会紧贴在标题上
	b.WriteString("\n\n")
}

// JSONPayloadOf 把任意结构序列化成 payload 文本（原始记录 JSON 落到 xm:payload 块里）。
func JSONPayloadOf(v any) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(raw)
}

// MemoryRecordFile 只读文件并解析，不回写索引。
//
// 列表要还原界面上的记录结构（payload 在文件里），所以需要批量读文件；
// 用 LoadRecord 会为每一行都写一次 DB，这里给一个纯读版本。
func (s *Service) MemoryRecordFile(id int) (*MemoryRecordDoc, error) {
	abs, rel, err := s.RecordPath(id)
	if err != nil {
		return nil, err
	}
	text, _, err := ReadTextFile(abs)
	if err != nil {
		return nil, err
	}
	doc, err := ParseRecordDoc(text)
	if err != nil {
		return nil, err
	}
	doc.RelPath = rel
	return doc, nil
}

// ListAllRecords 列出全部作品的记录（历史中心要跨作品聚合）。
func (s *Service) ListAllRecords(kind string) ([]MemoryRecord, error) {
	query := `SELECT r.id, r.book_id, r.kind, COALESCE(r.external_id,''), COALESCE(r.title,''), COALESCE(r.source,''),
		COALESCE(r.status,''), COALESCE(r.file_path,''), r.word_count, r.missing, COALESCE(r.parse_error,''),
		COALESCE(r.created_at,''), COALESCE(r.updated_at,''), COALESCE(b.title,'')
		FROM memory_records r LEFT JOIN books b ON b.id = r.book_id`
	var args []any
	if strings.TrimSpace(kind) != "" {
		query += ` WHERE r.kind=?`
		args = append(args, kind)
	}
	query += ` ORDER BY r.created_at DESC, r.id DESC`

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MemoryRecord
	for rows.Next() {
		var r MemoryRecord
		if err := rows.Scan(&r.ID, &r.BookID, &r.Kind, &r.ExternalID, &r.Title, &r.Source, &r.Status,
			&r.FilePath, &r.WordCount, &r.Missing, &r.ParseError, &r.CreatedAt, &r.UpdatedAt, &r.BookTitle); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
