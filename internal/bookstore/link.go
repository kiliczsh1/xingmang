package bookstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 链接目录：把已有作品目录**登记**进索引，不搬移、不复制、不删除任何文件
// （docs/books-memory-storage.md §11.3 / §11.4）。
//
// 这是"下一次更新直接链接目录、不需要备份导入"的落点。它要能做到：
// 换机器 / 重装后，把作品库指向原目录，全部数据（正文 + 四类记忆）就回来了。
//
// 幂等是硬要求：重复点同一个目录不能产生重复数据。做法是每一步都用
// "库里已有的自然键"去比对 —— 书的目录路径、章节/实体的文件路径、
// 会话/记录的外部 id —— 命中就更新，不命中才插入。

// LinkReport 是一次链接重建的结果，用于如实告诉用户"回来了什么"。
type LinkReport struct {
	Root     string         `json:"root"`
	Books    int            `json:"books"`
	Created  []string       `json:"created"`  // 新登记的作品
	Existing []string       `json:"existing"` // 已在库中的作品
	Details  map[string]int `json:"details"`  // 各表新增行数
	Errors   []string       `json:"errors"`   // 逐项失败的原因（不中断整本书）
}

func (r *LinkReport) add(kind string, n int) {
	if r.Details == nil {
		r.Details = map[string]int{}
	}
	if n > 0 {
		r.Details[kind] += n
	}
}

func (r *LinkReport) errf(format string, args ...any) {
	if r.Errors == nil {
		r.Errors = make([]string, 0, 4)
	}
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

// PreviewRoot 扫描目标目录，返回里面有哪些作品、各自是否已在库中。
//
// 只读，不写库、不写盘。前端据此让用户选"接管 / 搬移 / 取消"。
func (s *Service) PreviewRoot(dir string) (*LinkReport, error) {
	abs, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("目录不存在或无法访问: %w", err)
	}
	if !st.IsDir() {
		return nil, fmt.Errorf("这不是一个目录")
	}
	rep := &LinkReport{Root: abs}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		dirPath := filepath.Join(abs, e.Name())
		if !looksLikeBookDir(dirPath) {
			continue
		}
		title := e.Name()
		if meta, err := readBookJSONAt(dirPath); err == nil && strings.TrimSpace(meta.Title) != "" {
			title = meta.Title
		}
		if s.bookIDByFolder(e.Name()) > 0 {
			rep.Existing = append(rep.Existing, title)
		} else {
			rep.Created = append(rep.Created, title)
		}
		rep.Books++
	}
	return rep, nil
}

// LinkRoot 把作品库指向 dir，并把里面的作品登记进索引。
//
// **不移动任何文件**：只写一行 app_settings，然后逐个目录做链接重建。
// 与 SwitchRoot（搬移）并列，由前端让用户二选一。
func (s *Service) LinkRoot(dir string) (*LinkReport, error) {
	abs, err := filepath.Abs(strings.TrimSpace(dir))
	if err != nil {
		return nil, err
	}
	// 先校验目录（存在、是目录），失败就不要动指针
	if _, err := s.PreviewRoot(abs); err != nil {
		return nil, err
	}

	// 先切换指针：重建过程中读文件要用新根目录
	if _, err := s.DB.Exec(`INSERT INTO app_settings(key, value, updated_at) VALUES('books_root', ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP`, abs); err != nil {
		return nil, err
	}
	// 用一份新的报告：PreviewRoot 只是校验，它的计数不能带进来，
	// 否则"发现 N 部作品"会被累加两次。
	rep := &LinkReport{Root: abs}
	if err := s.LinkAll(rep); err != nil {
		return rep, err
	}
	return rep, nil
}

// LinkAll 对作品库里的每个目录做一次链接重建。
func (s *Service) LinkAll(rep *LinkReport) error {
	if rep == nil {
		rep = &LinkReport{}
	}
	root, err := s.Root()
	if err != nil {
		return err
	}
	rep.Root = root
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		dirPath := filepath.Join(root, e.Name())
		if !looksLikeBookDir(dirPath) {
			continue
		}
		bookID, err := s.LinkBookDir(dirPath, rep)
		if err != nil {
			// 单本失败只记账，不中断其余作品
			rep.errf("%s: %v", e.Name(), err)
			continue
		}
		if bookID > 0 {
			_ = bookID
		}
	}
	return nil
}

// LinkBookDir 登记（或重建索引）单个作品目录，返回库里的书 id。
//
// 顺序与 §11.4 一致：book.json → 卷与章节 → 实体目录 → 四类记忆。
// 任一步失败只记录原因，已经建好的部分保留。
func (s *Service) LinkBookDir(dirPath string, rep *LinkReport) (int, error) {
	if rep == nil {
		rep = &LinkReport{}
	}
	root, err := s.Root()
	if err != nil {
		return 0, err
	}
	folder, err := filepath.Rel(root, dirPath)
	if err != nil {
		return 0, err
	}
	folder = filepath.ToSlash(folder)
	if folder == "." || strings.HasPrefix(folder, "..") {
		return 0, fmt.Errorf("目录不在作品库内")
	}

	// 1) book.json → books 行（同 folder_path 复用现有 id）
	meta, metaErr := readBookJSONAt(dirPath)
	if metaErr != nil {
		meta = bookJSON{Title: filepath.Base(dirPath)}
	}
	bookID := s.bookIDByFolder(folder)
	created := false
	if bookID == 0 {
		res, err := s.DB.Exec(`INSERT INTO books (title, description, author, category, tags, status, is_public, target_words, folder_path)
			VALUES (?,?,?,?,?,?,0,?,?)`,
			firstName(meta.Title, filepath.Base(dirPath)), "", meta.Author, "", meta.Tags.ToDBValue(),
			firstName(meta.Status, "draft"), meta.Target, folder)
		if err != nil {
			return 0, err
		}
		id64, _ := res.LastInsertId()
		bookID = int(id64)
		created = true
		rep.add("books", 1)
	} else {
		// 已登记：书名以目录里的 book.json 为准（用户可能在外面改过）
		if _, err := s.DB.Exec(`UPDATE books SET title=?, author=?, tags=?, target_words=?, folder_path=? WHERE id=?`,
			firstName(meta.Title, filepath.Base(dirPath)), meta.Author, meta.Tags.ToDBValue(), meta.Target, folder, bookID); err != nil {
			return bookID, err
		}
	}
	if created {
		rep.Created = append(rep.Created, firstName(meta.Title, filepath.Base(dirPath)))
	} else {
		rep.Existing = append(rep.Existing, firstName(meta.Title, filepath.Base(dirPath)))
	}
	rep.Books++

	// 2) 章节 / 人物 / 世界观 / 文档 / 备忘录：复用已有的目录扫描
	if res, err := s.ScanBook(bookID); err != nil {
		rep.errf("%s 扫描章节失败: %v", folder, err)
	} else {
		rep.add("chapters", res.AddedChapters)
		rep.add("characters", res.AddedCharacters)
		rep.add("entries", res.AddedEntries)
		rep.add("docs", res.AddedDocs)
		rep.add("memos", res.AddedMemos)
	}

	// 3) 记忆：对话 / Agent 会话 / 图谱 / 拆书与生成历史
	s.linkBookMemoryFiles(bookID, dirPath, rep)
	// 全局 Agent 会话单独在库根
	s.linkRootMemoryFiles(root, rep)
	return bookID, nil
}

// linkBookMemoryFiles 从 <书>/memory/ 重建四类记忆的索引。
func (s *Service) linkBookMemoryFiles(bookID int, bookDir string, rep *LinkReport) {
	memRoot := filepath.Join(bookDir, memoryRootDir)
	if _, err := os.Stat(memRoot); err != nil {
		return
	}
	s.linkChats(bookID, filepath.Join(memRoot, "chats"), rep)
	s.linkAgentSessions(bookID, filepath.Join(memRoot, "agent"), rep)
	s.linkGraphVersions(bookID, filepath.Join(memRoot, "graph"), rep)
	s.linkRecords(bookID, filepath.Join(memRoot, "analysis"), MemoryKindAnalysis, rep)
	s.linkRecordsTree(bookID, filepath.Join(memRoot, "history"), MemoryKindGeneration, rep)
}

// linkRootMemoryFiles 处理库根的 _memory/agent/（全局会话）。
func (s *Service) linkRootMemoryFiles(root string, rep *LinkReport) {
	s.linkAgentSessions(0, filepath.Join(root, "_memory", "agent"), rep)
}

// ---------------- 逐类重建 ----------------

func (s *Service) linkChats(bookID int, dir string, rep *LinkReport) {
	files := listMarkdownFiles(dir, false)
	for _, abs := range files {
		text, _, err := ReadTextFile(abs)
		if err != nil {
			continue
		}
		doc, perr := ParseChatDoc(text)
		if perr != nil {
			rep.errf("对话 %s 解析失败: %v", filepath.Base(abs), perr)
			continue
		}
		rel, _ := relFromBook(s, bookID, abs)
		if rel == "" {
			continue
		}
		convID := doc.ConversationID
		var existID int
		// 先按对话 id 找；id 对不上时按文件路径找（文件被复制到别的机器上时 id 可能已变）
		if convID > 0 {
			s.DB.QueryRow(`SELECT id FROM conversations WHERE id=?`, convID).Scan(&existID)
		}
		if existID == 0 {
			s.DB.QueryRow(`SELECT id FROM conversations WHERE file_path=?`, rel).Scan(&existID)
		}
		title := firstName(doc.Title, strings.TrimSuffix(filepath.Base(abs), ".md"))
		if existID == 0 {
			res, err := s.DB.Exec(`INSERT INTO conversations (book_id, title, created_at, updated_at)
				VALUES (?,?,?,CURRENT_TIMESTAMP)`,
				bookID, title, FormatMemTime(firstNonZeroTime(doc.CreatedAt, time.Now())))
			if err != nil {
				rep.errf("登记对话 %s 失败: %v", filepath.Base(abs), err)
				continue
			}
			id64, _ := res.LastInsertId()
			convID = int(id64)
			rep.add("conversations", 1)
		} else {
			convID = existID
		}
		// 路径与哈希按文件为准回写（同时清掉可能残留的 missing/parse_error）
		raw, _ := os.ReadFile(abs)
		s.DB.Exec(`UPDATE conversations SET book_id=?, title=?, file_path=?, content_hash=?, missing=0, parse_error='',
			synced_at=CURRENT_TIMESTAMP WHERE id=?`, bookID, title, rel, HashText(string(raw)), convID)
	}
}

func (s *Service) linkAgentSessions(bookID int, dir string, rep *LinkReport) {
	files := listMarkdownFiles(dir, false)
	for _, abs := range files {
		text, _, err := ReadTextFile(abs)
		if err != nil {
			continue
		}
		doc, perr := ParseAgentDoc(text)
		if perr != nil {
			rep.errf("会话 %s 解析失败: %v", filepath.Base(abs), perr)
			continue
		}
		sessionID := strings.TrimSpace(doc.SessionID)
		if sessionID == "" {
			rep.errf("会话 %s 缺少 session_id，无法登记", filepath.Base(abs))
			continue
		}
		scope := firstName(doc.Scope, "global")
		var scopeID any
		if scope == "book" && bookID > 0 {
			scopeID = bookID
		} else if bookID <= 0 {
			scope = "global"
		}
		var exist string
		s.DB.QueryRow(`SELECT id FROM agent_sessions WHERE id=?`, sessionID).Scan(&exist)
		if exist == "" {
			if _, err := s.DB.Exec(`INSERT INTO agent_sessions (id, title, scope_kind, scope_id, model_id, permission, created_at, updated_at)
				VALUES (?,?,?,?,?,?,?,CURRENT_TIMESTAMP)`,
				sessionID, firstName(doc.Title, "导入的会话"), scope, scopeID, "", firstName(doc.Permission, "confirm"),
				FormatMemTime(firstNonZeroTime(doc.CreatedAt, time.Now()))); err != nil {
				rep.errf("登记会话 %s 失败: %v", filepath.Base(abs), err)
				continue
			}
			rep.add("agent_sessions", 1)
		}
		rel, _ := relFromRoot(s, abs)
		raw, _ := os.ReadFile(abs)
		s.DB.Exec(`UPDATE agent_sessions SET title=?, scope_kind=?, scope_id=?, permission=?,
			file_path=?, content_hash=?, missing=0, parse_error='', synced_at=CURRENT_TIMESTAMP WHERE id=?`,
			firstName(doc.Title, "导入的会话"), scope, scopeID, firstName(doc.Permission, "confirm"),
			rel, HashText(string(raw)), sessionID)

		// 轮次与账本：只有库里完全没有这个会话的事件时才重建，
		// 否则会把用户正在用的会话账本弄成两份。
		s.restoreAgentRuns(sessionID, doc, rel)
	}
}

// restoreAgentRuns 从 Markdown 重建轮次与账本事件（仅在库中无事件时）。
func (s *Service) restoreAgentRuns(sessionID string, doc *AgentDoc, rel string) {
	var n int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM agent_events WHERE session_id=?`, sessionID).Scan(&n); err != nil || n > 0 {
		return
	}
	seq := int64(0)
	for i, run := range doc.Runs {
		runID := strings.TrimSpace(run.RunID)
		if runID == "" {
			runID = fmt.Sprintf("linked-%d-%d", i+1, time.Now().UnixNano()%1e6)
		}
		status := firstName(run.Status, "completed")
		if _, err := s.DB.Exec(`INSERT OR REPLACE INTO agent_runs (id, session_id, status, input, error, created_at, updated_at)
			VALUES (?,?,?,?,?,?,CURRENT_TIMESTAMP)`,
			runID, sessionID, status, run.Input, run.Error,
			FormatMemTime(firstNonZeroTime(run.At, time.Now()))); err != nil {
			continue
		}
		emit := func(evType string, payload any) {
			seq++
			raw, _ := json.Marshal(payload)
			s.DB.Exec(`INSERT INTO agent_events (session_id, run_id, seq, type, payload) VALUES (?,?,?,?,?)`,
				sessionID, runID, seq, evType, string(raw))
		}
		emit("run.created", map[string]any{"run_id": runID, "input": run.Input})
		for _, tool := range run.Tools {
			emit("tool.started", map[string]any{
				"run_id": runID, "call_id": tool.CallID, "name": tool.Name,
				"arguments": json.RawMessage(orEmptyJSON(tool.Args)),
			})
			state := "completed"
			if tool.State != "" {
				state = tool.State
			}
			emit("tool.completed", map[string]any{
				"run_id": runID, "call_id": tool.CallID, "name": tool.Name,
				"result": json.RawMessage(orEmptyJSON(tool.Result)), "state": state,
			})
		}
		for _, p := range run.Proposals {
			emit("proposal.created", map[string]any{
				"proposal_id": p.ProposalID, "tool": p.Tool, "target": p.Target,
				"risk": p.Risk, "summary": p.Summary, "status": p.Status, "arguments": json.RawMessage("{}"),
			})
		}
		if strings.TrimSpace(run.Answer) != "" {
			emit("model.completed", map[string]any{"run_id": runID, "text": run.Answer})
		}
		emit("run.completed", map[string]any{"run_id": runID, "status": status, "error": run.Error})
	}
}

func (s *Service) linkGraphVersions(bookID int, dir string, rep *LinkReport) {
	files := listMarkdownFiles(dir, false)
	for _, abs := range files {
		text, _, err := ReadTextFile(abs)
		if err != nil {
			continue
		}
		doc, perr := ParseGraphDoc(text)
		if perr != nil {
			rep.errf("图谱 %s 解析失败: %v", filepath.Base(abs), perr)
			continue
		}
		rel, _ := relFromBook(s, bookID, abs)
		if rel == "" {
			continue
		}
		// "0-未分版本.md" 是未分版本实体（version_id 为空）的容器，
		// 它对应的是**没有版本行**的数据，不该建出一条 graph_versions。
		unversioned := doc.VersionID <= 0
		versionID := doc.VersionID
		if !unversioned {
			var exist int
			s.DB.QueryRow(`SELECT id FROM graph_versions WHERE id=?`, versionID).Scan(&exist)
			if exist == 0 {
				s.DB.QueryRow(`SELECT id FROM graph_versions WHERE book_id=? AND file_path=?`, bookID, rel).Scan(&exist)
			}
			name := firstName(doc.Name, strings.TrimSuffix(filepath.Base(abs), ".md"))
			if exist == 0 {
				res, err := s.DB.Exec(`INSERT INTO graph_versions (book_id, name, description, scope, chapter_count, entity_count, relation_count, file_path)
					VALUES (?,?,?,?,0,0,0,?)`,
					bookID, name, "", firstName(doc.Scope, "all"), rel)
				if err != nil {
					rep.errf("登记图谱版本 %s 失败: %v", filepath.Base(abs), err)
					continue
				}
				id64, _ := res.LastInsertId()
				versionID = int(id64)
				rep.add("graph_versions", 1)
			} else {
				versionID = exist
				s.DB.Exec(`UPDATE graph_versions SET name=?, scope=?, file_path=?, missing=0, parse_error='' WHERE id=?`,
					name, firstName(doc.Scope, "all"), rel, versionID)
			}
		}
		s.linkGraphBody(bookID, versionID, unversioned, doc, rep)
	}
}

// linkGraphBody 重建某份图谱文件里的实体、关系与章节引用。
func (s *Service) linkGraphBody(bookID, versionID int, unversioned bool, doc *GraphDoc, rep *LinkReport) {
	// 合并标记（primary）指向的实体可能还没落库（文件里的 id → 库里的 id 要映射），
	// 先暂存，等全部实体建完再补写。局部变量即可：这是单次同步调用。
	var pendingPrimary [][2]int
	var verArg any
	if !unversioned {
		verArg = versionID
	}
	// 实体 id → 库里的 id（按 (书, 版本, 名称) 匹配，避免 id 冲突）
	idMap := map[int]int{}
	for _, e := range doc.Entities {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			continue
		}
		var existID int
		if e.ID > 0 {
			// 文件里的 id 若已在库中且属于同一本书，直接复用
			var bookOf int
			if err := s.DB.QueryRow(`SELECT book_id FROM graph_entities WHERE id=?`, e.ID).Scan(&bookOf); err == nil && bookOf == bookID {
				existID = e.ID
			}
		}
		if existID == 0 {
			s.DB.QueryRow(`SELECT id FROM graph_entities WHERE book_id=? AND name=? AND COALESCE(version_id,0)=COALESCE(?,0)`,
				bookID, name, verArg).Scan(&existID)
		}
		if existID == 0 {
			res, err := s.DB.Exec(`INSERT INTO graph_entities (book_id, name, type, description, metadata, version_id, primary_id)
				VALUES (?,?,?,?,?,?,NULL)`,
				bookID, name, firstName(e.Type, "其他"), e.Description, orEmptyJSON(e.Metadata), verArg)
			if err != nil {
				rep.errf("登记图谱实体 %s 失败: %v", name, err)
				continue
			}
			id64, _ := res.LastInsertId()
			existID = int(id64)
			rep.add("graph_entities", 1)
		} else {
			s.DB.Exec(`UPDATE graph_entities SET type=?, description=?, metadata=? WHERE id=?`,
				firstName(e.Type, "其他"), e.Description, orEmptyJSON(e.Metadata), existID)
		}
		if e.ID > 0 {
			idMap[e.ID] = existID
		}
		// 合并标记（primary）：等全部实体落库后再统一写，避免引用到还没建的 id
		if e.PrimaryID > 0 {
			if target, ok := idMap[e.PrimaryID]; ok {
				s.DB.Exec(`UPDATE graph_entities SET primary_id=? WHERE id=?`, target, existID)
			} else {
				pendingPrimary = append(pendingPrimary, [2]int{existID, e.PrimaryID})
			}
		}
	}
	// 补写跨批次的主词条引用
	for _, pair := range pendingPrimary {
		if target, ok := idMap[pair[1]]; ok {
			s.DB.Exec(`UPDATE graph_entities SET primary_id=? WHERE id=?`, target, pair[0])
		}
	}

	for _, r := range doc.Relations {
		src, ok1 := idMap[r.SourceID]
		dst, ok2 := idMap[r.TargetID]
		if !ok1 || !ok2 {
			// 端点没解析出来就不建这条关系 —— 建了也是断的
			continue
		}
		var existID int
		if r.ID > 0 {
			var bookOf int
			if err := s.DB.QueryRow(`SELECT book_id FROM graph_relations WHERE id=?`, r.ID).Scan(&bookOf); err == nil && bookOf == bookID {
				existID = r.ID
			}
		}
		if existID == 0 {
			s.DB.QueryRow(`SELECT id FROM graph_relations WHERE book_id=? AND source_id=? AND target_id=? AND relation_type=?`,
				bookID, src, dst, firstName(r.RelationType, "related")).Scan(&existID)
		}
		if existID == 0 {
			res, err := s.DB.Exec(`INSERT INTO graph_relations (book_id, source_id, target_id, relation_type, description, evidence, confidence, version_id)
				VALUES (?,?,?,?,?,?,?,?)`,
				bookID, src, dst, firstName(r.RelationType, "related"), r.Description, r.Evidence, r.Confidence, verArg)
			if err != nil {
				rep.errf("登记图谱关系失败: %v", err)
				continue
			}
			id64, _ := res.LastInsertId()
			existID = int(id64)
			rep.add("graph_relations", 1)
		} else {
			s.DB.Exec(`UPDATE graph_relations SET description=?, evidence=?, confidence=? WHERE id=?`,
				r.Description, r.Evidence, r.Confidence, existID)
		}
		s.linkGraphChapters(bookID, versionID, unversioned, existID, 0, r.ChapterIDs, rep)
	}
	// 实体上的章节引用
	for _, e := range doc.Entities {
		eid, ok := idMap[e.ID]
		if !ok {
			continue
		}
		s.linkGraphChapters(bookID, versionID, unversioned, 0, eid, e.ChapterIDs, rep)
	}
}

// linkGraphChapters 重建章节引用。
//
// 章节 id 允许悬空（§15.2 第 3 条）：章节被删掉后，历史仍记录当时关联了哪几章，
// 这是事实，不该因为引用不到就丢掉。所以这里**不校验章节是否存在**。
func (s *Service) linkGraphChapters(bookID, versionID int, unversioned bool, relationID, entityID int, chapterIDs []int, rep *LinkReport) {
	if len(chapterIDs) == 0 {
		return
	}
	col, id := "entity_id", entityID
	if relationID > 0 {
		col, id = "relation_id", relationID
	}
	if id <= 0 {
		return
	}
	var verArg any
	if !unversioned {
		verArg = versionID
	}
	for _, cid := range chapterIDs {
		if cid <= 0 {
			continue
		}
		var n int
		s.DB.QueryRow(`SELECT COUNT(*) FROM graph_chapter_refs WHERE book_id=? AND chapter_id=? AND `+col+`=?`,
			bookID, cid, id).Scan(&n)
		if n > 0 {
			continue
		}
		if _, err := s.DB.Exec(`INSERT INTO graph_chapter_refs (book_id, version_id, chapter_id, `+col+`)
			VALUES (?,?,?,?)`, bookID, verArg, cid, id); err != nil {
			rep.errf("登记章节引用失败: %v", err)
			continue
		}
		rep.add("graph_chapter_refs", 1)
	}
}

func (s *Service) linkRecords(bookID int, dir, kind string, rep *LinkReport) {
	for _, abs := range listMarkdownFiles(dir, false) {
		s.linkOneRecord(bookID, abs, kind, rep)
	}
}

// linkRecordsTree 递归处理按月分目录的生成历史。
func (s *Service) linkRecordsTree(bookID int, dir, kind string, rep *LinkReport) {
	for _, abs := range listMarkdownFiles(dir, true) {
		s.linkOneRecord(bookID, abs, kind, rep)
	}
}

func (s *Service) linkOneRecord(bookID int, abs, kind string, rep *LinkReport) {
	text, _, err := ReadTextFile(abs)
	if err != nil {
		return
	}
	doc, perr := ParseRecordDoc(text)
	if perr != nil {
		rep.errf("记录 %s 解析失败: %v", filepath.Base(abs), perr)
		return
	}
	rel, _ := relFromBook(s, bookID, abs)
	if rel == "" {
		return
	}
	externalID := strings.TrimSpace(doc.ExternalID)
	if externalID == "" {
		// 没有 external_id 就用文件路径兜底，保证重复链接不产生重复行
		externalID = "file:" + rel
	}
	title := firstName(doc.Title, strings.TrimSuffix(filepath.Base(abs), ".md"))
	var existID int
	s.DB.QueryRow(`SELECT id FROM memory_records WHERE book_id=? AND kind=? AND external_id=?`, bookID, kind, externalID).Scan(&existID)
	if existID == 0 {
		res, err := s.DB.Exec(`INSERT INTO memory_records (book_id, kind, external_id, title, source, status, file_path, created_at)
			VALUES (?,?,?,?,?,?,?,?)`,
			bookID, kind, externalID, title, doc.Source, firstName(doc.Status, "completed"), rel,
			FormatMemTime(firstNonZeroTime(doc.CreatedAt, time.Now())))
		if err != nil {
			rep.errf("登记记录 %s 失败: %v", filepath.Base(abs), err)
			return
		}
		id64, _ := res.LastInsertId()
		existID = int(id64)
		rep.add(kind, 1)
	}
	raw, _ := os.ReadFile(abs)
	s.DB.Exec(`UPDATE memory_records SET title=?, source=?, status=?, file_path=?, content_hash=?, word_count=?,
		missing=0, parse_error='', synced_at=CURRENT_TIMESTAMP WHERE id=?`,
		title, doc.Source, firstName(doc.Status, "completed"), rel, HashText(string(raw)), WordCount(doc.Result), existID)
}

// ---------------- 辅助 ----------------

func (s *Service) bookIDByFolder(folder string) int {
	var id int
	s.DB.QueryRow(`SELECT id FROM books WHERE folder_path=?`, filepath.ToSlash(folder)).Scan(&id)
	return id
}

// looksLikeBookDir 判断一个目录是不是作品目录。
//
// 只要有 book.json、章节文件或 memory/ 三者之一就算 —— 用户手工建的目录
// 可能还没有 book.json，但已经在里面写了章节，那也是他的作品。
func looksLikeBookDir(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "book.json")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, memoryRootDir)); err == nil {
		return true
	}
	for _, sub := range []string{"characters", "world", "outlines", "ideas", "covers"} {
		if st, err := os.Stat(filepath.Join(dir, sub)); err == nil && st.IsDir() {
			return true
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".md", ".txt":
			return true
		}
	}
	return false
}

func readBookJSONAt(dir string) (bookJSON, error) {
	var meta bookJSON
	raw, err := os.ReadFile(filepath.Join(dir, "book.json"))
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		return meta, err
	}
	return meta, nil
}

// listMarkdownFiles 列出目录下的记录文件；recursive=true 时递归（按月分目录用）。
func listMarkdownFiles(dir string, recursive bool) []string {
	var out []string
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		return out
	}
	if recursive {
		_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(d.Name()), ".md") {
				out = append(out, path)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return out
		}
		for _, e := range entries {
			if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".md") {
				continue
			}
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out
}

// relFromRoot 求文件相对作品库根的路径（全局会话用，它落在库根 _memory/ 下）。
func relFromRoot(s *Service, abs string) (string, error) {
	root, err := s.Root()
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("文件不在作品库内")
	}
	return rel, nil
}

// relFromBook 求文件相对书目录的路径（用于写回索引）。
func relFromBook(s *Service, bookID int, abs string) (string, error) {
	bookDir, err := s.BookDir(bookID)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(bookDir, abs)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(rel)
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("文件不在书目录内")
	}
	return rel, nil
}

func firstName(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func firstNonZeroTime(v, fallback time.Time) time.Time {
	if v.IsZero() {
		return fallback
	}
	return v
}

func orEmptyJSON(s string) string {
	if strings.TrimSpace(s) == "" {
		return "{}"
	}
	return s
}
