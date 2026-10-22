package agent

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/models"
)

// ToolDef 是暴露给模型的工具定义（OpenAI function 格式）。
type ToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// openAITool 是请求体里 tools 数组的元素形态。
type openAITool struct {
	Type     string  `json:"type"`
	Function ToolDef `json:"function"`
}

func toolParamString(desc string) map[string]any {
	return map[string]any{"type": "string", "description": desc}
}

func toolParamInt(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

// 写工具风险分级：low = 自动档可直写；high = 任何档位都要提案-审批（正文类）
var writeToolRisk = map[string]string{
	"chapter_append":        "high",
	"chapter_replace":       "high",
	"book_create":           "low",
	"character_create":      "low",
	"graph_entity_create":   "low",
	"graph_relation_create": "low",
}

// toolDefs 返回暴露给模型的全量工具定义（第一版全量注入，工具多了再改延迟加载）。
func toolDefs() []openAITool {
	write := []openAITool{
		{Type: "function", Function: ToolDef{
			Name:        "chapter_append",
			Description: "把一段正文追加到章节末尾（用于续写、补写结尾）。会生成提案，用户批准后才真正写入。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chapter_id": toolParamInt("章节 id"),
					"content":    toolParamString("要追加的正文内容"),
				},
				"required": []string{"chapter_id", "content"},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "chapter_replace",
			Description: "把章节正文中的一段原文替换为新内容（用于改写、润色）。find 必须与正文逐字一致。会生成提案，用户批准后才写入。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chapter_id":   toolParamInt("章节 id"),
					"find":         toolParamString("要被替换的原文片段，必须逐字一致"),
					"replace_with": toolParamString("替换后的新内容"),
				},
				"required": []string{"chapter_id", "find", "replace_with"},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "book_create",
			Description: "新建一本书（书名 + 可选简介/分类）。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"title":       toolParamString("书名"),
					"description": toolParamString("简介，可省略"),
					"category":    toolParamString("分类，可省略"),
				},
				"required": []string{"title"},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "character_create",
			Description: "为某本书新建一张角色卡。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id":     toolParamInt("书籍 id；绑定单本书的会话可省略"),
					"name":        toolParamString("角色名"),
					"gender":      toolParamString("性别：male/female/unknown"),
					"personality": toolParamString("性格描述"),
					"info":        toolParamString("角色信息"),
				},
				"required": []string{"name"},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "graph_entity_create",
			Description: "向知识图谱添加一个实体（人物/地点/物品/设定等）。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id":     toolParamInt("书籍 id；绑定单本书的会话可省略"),
					"name":        toolParamString("实体名称"),
					"type":        toolParamString("实体类型，如 character/location/item/faction/setting"),
					"description": toolParamString("实体描述"),
				},
				"required": []string{"name", "type"},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "graph_relation_create",
			Description: "向知识图谱添加一条关系（两个实体之间）。实体可用 id 或精确名称指定。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id":       toolParamInt("书籍 id；绑定单本书的会话可省略"),
					"source_id":     toolParamInt("起始实体 id（与 source_name 二选一）"),
					"source_name":   toolParamString("起始实体名称（精确匹配）"),
					"target_id":     toolParamInt("目标实体 id（与 target_name 二选一）"),
					"target_name":   toolParamString("目标实体名称（精确匹配）"),
					"relation_type": toolParamString("关系类型，如 related/parent/lover/enemy"),
					"description":   toolParamString("关系描述"),
				},
			},
		}},
	}
	return append(readToolDefs(), write...)
}

func readToolDefs() []openAITool {
	return []openAITool{
		{Type: "function", Function: ToolDef{
			Name:        "book_list",
			Description: "列出库里的全部书籍（含 id、书名、分类、状态、章节总数）。全局作用域下用它找书；绑定单本书的会话看不到此工具。",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "book_read",
			Description: "读取一本书的元信息、卷结构与章节目录（含每章摘要开头）。写正文前先读它了解整体结构。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id": toolParamInt("书籍 id；绑定单本书的会话可省略"),
				},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "chapter_list",
			Description: "列出某本书的全部章节（id、标题、序号、字数、摘要开头），不返回正文。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id": toolParamInt("书籍 id；绑定单本书的会话可省略"),
				},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "chapter_read",
			Description: "读取章节正文。长章节请用 offset/limit 分段读（按字符计），返回值里的 total_chars 是总字数。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"chapter_id": toolParamInt("章节 id"),
					"offset":     toolParamInt("起始字符位置，默认 0"),
					"limit":      toolParamInt("本次读取的字符数，默认 6000，最大 12000"),
				},
				"required": []string{"chapter_id"},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "character_list",
			Description: "列出某本书的角色卡（姓名、性别、性格、信息摘要）。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id": toolParamInt("书籍 id；绑定单本书的会话可省略"),
				},
			},
		}},
		{Type: "function", Function: ToolDef{
			Name:        "graph_query",
			Description: "查询知识图谱。按关键词搜实体（人名/地名/设定词），返回实体与它们之间的关系；不带关键词返回前 30 个实体。",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"book_id": toolParamInt("书籍 id；绑定单本书的会话可省略"),
					"keyword": toolParamString("实体名称关键词，可省略"),
				},
			},
		}},
	}
}

// toolContext 是工具执行时的作用域环境。
type toolContext struct {
	m         *Manager
	scopeKind string
	scopeID   int64 // book 作用域下强制使用；global 下为 0
	session   *Session
	runID     string
	callID    string
}

// str 取字符串参数（nil 安全）。
func str(args map[string]any, key string) string {
	s, _ := args[key].(string)
	return strings.TrimSpace(s)
}

// resolveBookID 解析 book_id 参数并做作用域闸门：
// book 作用域一律强制为本会话的书（模型传别的 id 也无效）；
// global 作用域要求模型显式给 id。
func (t *toolContext) resolveBookID(args map[string]any) (int64, error) {
	if t.scopeKind == "book" {
		return t.scopeID, nil
	}
	v, ok := toInt(args["book_id"])
	if !ok || v <= 0 {
		return 0, fmt.Errorf("缺少 book_id 参数（全局作用域必须显式指定书籍）")
	}
	return v, nil
}

func toInt(v any) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	}
	return 0, false
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// execTool 执行一个工具调用，返回 JSON 可序列化结果。
// 任何错误都作为字符串返回给模型（让模型有机会纠正参数），只有作用域拒绝才向上抛。
func (t *toolContext) execTool(name string, args map[string]any) (any, error) {
	if _, isWrite := writeToolRisk[name]; isWrite {
		return t.execWrite(name, writeToolRisk[name], args)
	}
	switch name {
	case "book_list":
		if t.scopeKind == "book" {
			return nil, fmt.Errorf("当前会话已绑定单本书，没有 book_list 工具")
		}
		return t.bookList()
	case "book_read":
		id, err := t.resolveBookID(args)
		if err != nil {
			return nil, err
		}
		return t.bookRead(id)
	case "chapter_list":
		id, err := t.resolveBookID(args)
		if err != nil {
			return nil, err
		}
		return t.chapterList(id)
	case "chapter_read":
		id, err := t.resolveBookID(args)
		if err != nil {
			return nil, err
		}
		return t.chapterRead(args, id)
	case "character_list":
		id, err := t.resolveBookID(args)
		if err != nil {
			return nil, err
		}
		return t.characterList(id)
	case "graph_query":
		id, err := t.resolveBookID(args)
		if err != nil {
			return nil, err
		}
		kw, _ := args["keyword"].(string)
		return t.graphQuery(id, kw)
	default:
		return nil, fmt.Errorf("未知工具: %s", name)
	}
}

func (t *toolContext) bookList() (any, error) {
	rows, err := t.m.db.Query(`SELECT b.id, b.title, COALESCE(b.category,''), COALESCE(b.status,''), COALESCE(b.description,''),
		(SELECT COUNT(*) FROM chapters c WHERE c.book_id = b.id)
		FROM books b ORDER BY b.updated_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []any{}
	for rows.Next() {
		var id int64
		var title, category, status, desc string
		var chs int
		if err := rows.Scan(&id, &title, &category, &status, &desc, &chs); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"book_id": id, "title": title, "category": category,
			"status": status, "chapter_count": chs, "description": truncate(desc, 80),
		})
	}
	return map[string]any{"books": list}, rows.Err()
}

func (t *toolContext) bookRead(bookID int64) (any, error) {
	var id int64
	var title, desc, category, status string
	err := t.m.db.QueryRow("SELECT id, title, COALESCE(description,''), COALESCE(category,''), COALESCE(status,'') FROM books WHERE id=?", bookID).
		Scan(&id, &title, &desc, &category, &status)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("书不存在（id=%d）", bookID)
	}
	if err != nil {
		return nil, err
	}
	result := map[string]any{"book_id": id, "title": title, "description": desc, "category": category, "status": status}

	// 卷结构
	vols := []any{}
	vrows, err := t.m.db.Query("SELECT id, title, order_num FROM volumes WHERE book_id=? ORDER BY order_num ASC", bookID)
	if err == nil {
		defer vrows.Close()
		for vrows.Next() {
			var vid int64
			var vtitle string
			var order int
			if vrows.Scan(&vid, &vtitle, &order) == nil {
				vols = append(vols, map[string]any{"volume_id": vid, "title": vtitle, "order": order})
			}
		}
	}
	result["volumes"] = vols

	// 章节目录（含摘要开头，给模型足够上下文判断该读哪章）
	chs := []any{}
	crows, err := t.m.db.Query(`SELECT id, title, order_num, COALESCE(summary,''), COALESCE(volume_id,0), length(COALESCE(content,''))
		FROM chapters WHERE book_id=? ORDER BY order_num ASC LIMIT 300`, bookID)
	if err == nil {
		defer crows.Close()
		for crows.Next() {
			var cid, vid int64
			var ctitle, summary string
			var order, clen int
			if crows.Scan(&cid, &ctitle, &order, &summary, &vid, &clen) == nil {
				chs = append(chs, map[string]any{
					"chapter_id": cid, "title": ctitle, "order": order,
					"summary_head": truncate(summary, 100), "char_count": clen, "volume_id": vid,
				})
			}
		}
	}
	result["chapters"] = chs
	return result, nil
}

func (t *toolContext) chapterList(bookID int64) (any, error) {
	rows, err := t.m.db.Query(`SELECT id, title, order_num, COALESCE(summary,''), length(COALESCE(content,''))
		FROM chapters WHERE book_id=? ORDER BY order_num ASC LIMIT 500`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []any{}
	for rows.Next() {
		var id int64
		var title, summary string
		var order, clen int
		if err := rows.Scan(&id, &title, &order, &summary, &clen); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"chapter_id": id, "title": title, "order": order,
			"summary_head": truncate(summary, 120), "char_count": clen,
		})
	}
	return map[string]any{"chapters": list}, rows.Err()
}

func (t *toolContext) chapterRead(args map[string]any, bookID int64) (any, error) {
	chID, ok := toInt(args["chapter_id"])
	if !ok || chID <= 0 {
		return nil, fmt.Errorf("缺少 chapter_id 参数")
	}
	var title, content string
	var order int
	err := t.m.db.QueryRow(`SELECT title, COALESCE(content,''), order_num FROM chapters WHERE id=? AND book_id=?`, chID, bookID).
		Scan(&title, &content, &order)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("章节不存在或不属于当前作用域的书（chapter_id=%d）", chID)
	}
	if err != nil {
		return nil, err
	}
	runes := []rune(content)
	offset := 0
	limit := 6000
	if v, ok := toInt(args["offset"]); ok && v > 0 {
		offset = int(v)
	}
	if v, ok := toInt(args["limit"]); ok && v > 0 {
		limit = int(v)
		if limit > 12000 {
			limit = 12000
		}
	}
	if offset >= len(runes) {
		return map[string]any{"chapter_id": chID, "title": title, "total_chars": len(runes),
			"offset": offset, "content": "", "note": "offset 超出正文长度"}, nil
	}
	end := offset + limit
	if end > len(runes) {
		end = len(runes)
	}
	return map[string]any{
		"chapter_id": chID, "title": title, "order": order,
		"total_chars": len(runes), "offset": offset, "chars_returned": end - offset,
		"content": string(runes[offset:end]),
	}, nil
}

func (t *toolContext) characterList(bookID int64) (any, error) {
	rows, err := t.m.db.Query(`SELECT name, COALESCE(gender,''), COALESCE(personality,''), COALESCE(info,'')
		FROM characters WHERE book_id=? ORDER BY id ASC LIMIT 200`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []any{}
	for rows.Next() {
		var name, gender, personality, info string
		if err := rows.Scan(&name, &gender, &personality, &info); err != nil {
			continue
		}
		list = append(list, map[string]any{
			"name": name, "gender": gender,
			"personality": truncate(personality, 200), "info": truncate(info, 300),
		})
	}
	return map[string]any{"characters": list}, rows.Err()
}

func (t *toolContext) graphQuery(bookID int64, keyword string) (any, error) {
	var rows *sql.Rows
	var err error
	kw := strings.TrimSpace(keyword)
	if kw == "" {
		rows, err = t.m.db.Query(`SELECT id, name, type, COALESCE(description,'') FROM graph_entities
			WHERE book_id=? AND (primary_id IS NULL) ORDER BY id ASC LIMIT 30`, bookID)
	} else {
		like := "%" + kw + "%"
		rows, err = t.m.db.Query(`SELECT id, name, type, COALESCE(description,'') FROM graph_entities
			WHERE book_id=? AND (primary_id IS NULL) AND (name LIKE ? OR type LIKE ? OR description LIKE ?)
			ORDER BY id ASC LIMIT 30`, bookID, like, like, like)
	}
	if err != nil {
		return nil, err
	}
	entities := []any{}
	ids := []int64{}
	for rows.Next() {
		var id int64
		var name, typ, desc string
		if err := rows.Scan(&id, &name, &typ, &desc); err != nil {
			continue
		}
		ids = append(ids, id)
		entities = append(entities, map[string]any{"entity_id": id, "name": name, "type": typ, "description": truncate(desc, 200)})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	relations := []any{}
	if len(ids) > 0 {
		ph := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
		args := make([]any, 0, len(ids))
		for _, id := range ids {
			args = append(args, id)
		}
		rrows, err := t.m.db.Query(`SELECT r.id, r.source_id, es.name, r.target_id, et.name, r.relation_type, COALESCE(r.description,'')
			FROM graph_relations r
			JOIN graph_entities es ON es.id = r.source_id
			JOIN graph_entities et ON et.id = r.target_id
			WHERE r.source_id IN (`+ph+`) OR r.target_id IN (`+ph+`) LIMIT 60`,
			append(args, args...)...)
		if err == nil {
			defer rrows.Close()
			for rrows.Next() {
				var rid, sid, tid int64
				var sname, tname, rtype, desc string
				if rrows.Scan(&rid, &sid, &sname, &tid, &tname, &rtype, &desc) == nil {
					relations = append(relations, map[string]any{
						"source": sname, "target": tname, "relation_type": rtype, "description": truncate(desc, 120),
					})
				}
			}
		}
	}
	return map[string]any{"entities": entities, "relations": relations,
		"note": "实体搜索基于名称/类型/描述关键词匹配，未命中不代表实体不存在"}, nil
}

// ===== 写工具：提案生成与执行 =====

// execWrite 处理全部写工具：校验参数 → 生成摘要 → 按会话权限决定
// 「直接执行（auto + 低风险）」还是「生成提案等待批准」。
func (t *toolContext) execWrite(name, risk string, args map[string]any) (any, error) {
	// 作用域注入：book 会话强制锁书；global 要求显式 book_id（book_create 除外）
	switch name {
	case "character_create", "graph_entity_create", "graph_relation_create", "chapter_append", "chapter_replace":
		id, err := t.resolveBookID(args)
		if err != nil {
			return nil, err
		}
		args["book_id"] = id
	}

	// 必填校验 + 摘要 / 目标
	var target, summary string
	switch name {
	case "chapter_append":
		content := str(args, "content")
		if content == "" {
			return nil, fmt.Errorf("content 不能为空")
		}
		title, err := t.chapterTitle(args)
		if err != nil {
			return nil, err
		}
		target = fmt.Sprintf("chapter:%v", args["chapter_id"])
		summary = fmt.Sprintf("追加正文 · %s（+%d 字）", title, len([]rune(content)))
	case "chapter_replace":
		find := str(args, "find")
		if find == "" || str(args, "replace_with") == "" {
			return nil, fmt.Errorf("find 与 replace_with 不能为空")
		}
		title, err := t.chapterTitle(args)
		if err != nil {
			return nil, err
		}
		target = fmt.Sprintf("chapter:%v", args["chapter_id"])
		head := []rune(find)
		if len(head) > 12 {
			head = head[:12]
		}
		summary = fmt.Sprintf("替换正文 · %s「%s…」", title, string(head))
	case "book_create":
		title := str(args, "title")
		if title == "" {
			return nil, fmt.Errorf("title 不能为空")
		}
		target = "book"
		summary = "新建书籍 · 《" + title + "》"
	case "character_create":
		if str(args, "name") == "" {
			return nil, fmt.Errorf("name 不能为空")
		}
		target = fmt.Sprintf("character:book:%v", args["book_id"])
		summary = "新建角色卡 · " + str(args, "name")
	case "graph_entity_create":
		if str(args, "name") == "" || str(args, "type") == "" {
			return nil, fmt.Errorf("name 与 type 不能为空")
		}
		target = fmt.Sprintf("graph-entity:book:%v", args["book_id"])
		summary = "图谱实体 · " + str(args, "name")
	case "graph_relation_create":
		if str(args, "source_name") == "" && args["source_id"] == nil {
			return nil, fmt.Errorf("source_name 与 source_id 至少给一个")
		}
		if str(args, "target_name") == "" && args["target_id"] == nil {
			return nil, fmt.Errorf("target_name 与 target_id 至少给一个")
		}
		target = fmt.Sprintf("graph-relation:book:%v", args["book_id"])
		summary = fmt.Sprintf("图谱关系 · %s →%s %s",
			relationEnd(args, "source"), str(args, "relation_type"), relationEnd(args, "target"))
	default:
		return nil, fmt.Errorf("未知写工具: %s", name)
	}

	perm := t.session.Permission
	if perm == "" {
		perm = PermConfirm
	}
	argsJSON, _ := json.Marshal(args)

	p, err := t.m.CreateProposal(t.session, t.runID, t.callID, name, target, string(argsJSON), risk, summary)
	if err != nil {
		return nil, err
	}

	// 自动档 + 低风险：直接落库
	if perm == PermAuto && risk == "low" {
		result, err := t.m.ApplyProposal(p)
		if err != nil {
			return nil, fmt.Errorf("自动写入失败: %s", err.Error())
		}
		return map[string]any{"applied": true, "summary": result}, nil
	}

	msg := "已生成写入提案（等待用户批准，批准后才会写入）。不要重复提交同一提案，如实告知用户即可。"
	if perm == PermObserve {
		msg = "已生成写入提案（当前为试运行模式，不会写入任何数据）。"
	}
	return map[string]any{"proposal_id": p.ID, "status": p.Status, "summary": summary, "note": msg}, nil
}

func relationEnd(args map[string]any, side string) string {
	if v, ok := args[side+"_name"].(string); ok && v != "" {
		return v
	}
	return fmt.Sprintf("#%v", args[side+"_id"])
}

// chapterTitle 校验章节存在且属于作用域书，返回展示标题。
func (t *toolContext) chapterTitle(args map[string]any) (string, error) {
	chID, ok := toInt(args["chapter_id"])
	if !ok || chID <= 0 {
		return "", fmt.Errorf("缺少 chapter_id 参数")
	}
	bookID, _ := toInt(args["book_id"])
	var title string
	err := t.m.db.QueryRow("SELECT title FROM chapters WHERE id=? AND book_id=?", chID, bookID).Scan(&title)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("章节不存在或不属于当前作用域的书（chapter_id=%d）", chID)
	}
	return title, err
}

// ApplyProposal 真正执行一个提案（approve 或 auto 低风险直写时调用）。
func (m *Manager) ApplyProposal(p *Proposal) (string, error) {
	result, before, err := m.execProposalWrite(p)
	if err != nil {
		m.db.Exec("UPDATE agent_change_proposals SET status=?, apply_error=?, decided_at=CURRENT_TIMESTAMP WHERE id=?",
			ProposalApplyFailed, err.Error(), p.ID)
		m.AppendEvent(p.SessionID, p.RunID, EventProposalDecided, map[string]any{
			"proposal_id": p.ID, "tool": p.ToolName, "target": p.Target, "risk": p.Risk,
			"summary": p.Summary, "status": ProposalApplyFailed, "apply_error": err.Error(),
		})
		return "", err
	}
	if before != nil {
		m.db.Exec(`UPDATE agent_change_proposals SET status=?, before_text=?, applied_at=CURRENT_TIMESTAMP, decided_at=CURRENT_TIMESTAMP WHERE id=?`,
			ProposalApplied, *before, p.ID)
	} else {
		m.db.Exec(`UPDATE agent_change_proposals SET status=?, applied_at=CURRENT_TIMESTAMP, decided_at=CURRENT_TIMESTAMP WHERE id=?`,
			ProposalApplied, p.ID)
	}
	m.AppendEvent(p.SessionID, p.RunID, EventProposalDecided, map[string]any{
		"proposal_id": p.ID, "tool": p.ToolName, "target": p.Target, "risk": p.Risk,
		"summary": p.Summary, "status": ProposalApplied, "result": result,
	})
	return result, nil
}

// execProposalWrite 在业务表上执行写操作；chapter 类返回原文快照用于回滚。
func (m *Manager) execProposalWrite(p *Proposal) (string, *string, error) {
	var ac argsCarrier
	if err := json.Unmarshal([]byte(p.Arguments), &ac); err != nil {
		return "", nil, fmt.Errorf("提案参数损坏")
	}
	switch p.ToolName {
	case "book_create":
		// 走 bookstore：建目录 + book.json + 插库，Agent 建的书同样是一个真实文件夹。
		b := &models.Book{
			Title:       str(ac, "title"),
			Description: str(ac, "description"),
			Category:    str(ac, "category"),
			Status:      "draft",
			Tags:        models.StringList{},
		}
		if err := bookstore.NewService(m.db).CreateBook(b); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("已创建书籍《%s》（id=%d）", b.Title, b.ID), nil, nil

	case "character_create":
		bookID, _ := toInt(ac["book_id"])
		if bookID <= 0 {
			return "", nil, fmt.Errorf("缺少 book_id")
		}
		gender := str(ac, "gender")
		if gender == "" {
			gender = "unknown"
		}
		personality := str(ac, "personality")
		info := str(ac, "info")
		card := &models.Character{
			BookID:      int(bookID),
			Name:        str(ac, "name"),
			Gender:      gender,
			Personality: &personality,
			Info:        &info,
		}
		res, err := m.db.Exec("INSERT INTO characters (book_id, name, gender, personality, info) VALUES (?,?,?,?,?)",
			card.BookID, card.Name, card.Gender, personality, info)
		if err != nil {
			return "", nil, err
		}
		id, _ := res.LastInsertId()
		card.ID = int(id)
		// 人物是文件类实体：只在库里插一行会让它不出现在书目录里，
		// 于是"把目录拷到新机器"就会少一张角色卡。写文件失败即回滚这一行。
		rel, syncErr := bookstore.NewService(m.db).SyncCharacter(card, "")
		if syncErr != nil {
			m.db.Exec("DELETE FROM characters WHERE id=?", id)
			return "", nil, fmt.Errorf("写入角色卡文件失败: %w", syncErr)
		}
		if _, err := m.db.Exec("UPDATE characters SET file_path=? WHERE id=?", rel, id); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("已创建角色卡「%s」（id=%d）", card.Name, id), nil, nil

	case "graph_entity_create":
		bookID, _ := toInt(ac["book_id"])
		if bookID <= 0 {
			return "", nil, fmt.Errorf("缺少 book_id")
		}
		res, err := m.db.Exec("INSERT INTO graph_entities (book_id, name, type, description) VALUES (?,?,?,?)",
			bookID, str(ac, "name"), str(ac, "type"), str(ac, "description"))
		if err != nil {
			return "", nil, err
		}
		id, _ := res.LastInsertId()
		// 工具建的是"未分版本"实体（version_id 为空），落到该书的 0-未分版本.md。
		// 不标脏的话，AI 改了图谱而文件纹丝不动 —— 下次按目录恢复就把这次改动丢了。
		bookstore.MarkMemoryDirty(bookstore.MemoryKindGraph, bookstore.GraphDirtyKey(0, int(bookID)))
		return fmt.Sprintf("已在图谱添加实体「%s」（id=%d）", str(ac, "name"), id), nil, nil

	case "graph_relation_create":
		bookID, _ := toInt(ac["book_id"])
		if bookID <= 0 {
			return "", nil, fmt.Errorf("缺少 book_id")
		}
		srcID, err := m.resolveEntityID(bookID, ac, "source")
		if err != nil {
			return "", nil, err
		}
		tgtID, err := m.resolveEntityID(bookID, ac, "target")
		if err != nil {
			return "", nil, err
		}
		rtype := str(ac, "relation_type")
		if rtype == "" {
			rtype = "related"
		}
		_, err = m.db.Exec("INSERT INTO graph_relations (book_id, source_id, target_id, relation_type, description) VALUES (?,?,?,?,?)",
			bookID, srcID, tgtID, rtype, str(ac, "description"))
		if err != nil {
			return "", nil, err
		}
		bookstore.MarkMemoryDirty(bookstore.MemoryKindGraph, bookstore.GraphDirtyKey(0, int(bookID)))
		return fmt.Sprintf("已添加图谱关系 %s —%s→ %s", relationEnd(ac, "source"), rtype, relationEnd(ac, "target")), nil, nil

	case "chapter_append":
		chID, _ := toInt(ac["chapter_id"])
		bookID, _ := toInt(ac["book_id"])
		var title string
		err := m.db.QueryRow("SELECT title FROM chapters WHERE id=? AND book_id=?", chID, bookID).Scan(&title)
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("章节不存在或不属于作用域书（chapter_id=%d）", chID)
		}
		if err != nil {
			return "", nil, err
		}
		svc := bookstore.NewService(m.db)
		old, err := svc.ChapterText(int(chID))
		if err != nil {
			return "", nil, err
		}
		newContent := str(ac, "content")
		if strings.TrimSpace(old) != "" {
			newContent = old + "\n" + newContent
		}
		// 正文是文件，必须走 SetChapterText 落盘，否则下次读取会把这次追加吞掉。
		if err := svc.SetChapterText(int(chID), newContent); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("已向「%s」追加 %d 字", title, len([]rune(str(ac, "content")))), &old, nil

	case "chapter_replace":
		chID, _ := toInt(ac["chapter_id"])
		bookID, _ := toInt(ac["book_id"])
		find := str(ac, "find")
		replaceWith := str(ac, "replace_with")
		var title string
		err := m.db.QueryRow("SELECT title FROM chapters WHERE id=? AND book_id=?", chID, bookID).Scan(&title)
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("章节不存在或不属于作用域书（chapter_id=%d）", chID)
		}
		if err != nil {
			return "", nil, err
		}
		svc := bookstore.NewService(m.db)
		old, err := svc.ChapterText(int(chID))
		if err != nil {
			return "", nil, err
		}
		if !strings.Contains(old, find) {
			return "", nil, fmt.Errorf("正文中未找到要替换的片段（可能已被修改），已拒绝写入")
		}
		newContent := strings.Replace(old, find, replaceWith, 1)
		if err := svc.SetChapterText(int(chID), newContent); err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("已替换「%s」中的片段", title), &old, nil

	default:
		return "", nil, fmt.Errorf("未知写工具: %s", p.ToolName)
	}
}

// resolveEntityID 按 id 或精确名称定位实体。
func (m *Manager) resolveEntityID(bookID int64, ac argsCarrier, side string) (int64, error) {
	if id, ok := toInt(ac[side+"_id"]); ok && id > 0 {
		return id, nil
	}
	name := str(ac, side+"_name")
	if name == "" {
		return 0, fmt.Errorf("%s 未指定", side)
	}
	var id int64
	err := m.db.QueryRow("SELECT id FROM graph_entities WHERE book_id=? AND name=? ORDER BY id ASC LIMIT 1", bookID, name).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("图谱中不存在实体「%s」，请先用 graph_entity_create 创建", name)
	}
	return id, err
}
