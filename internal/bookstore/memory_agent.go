package bookstore

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Agent 会话的文件存储（docs/books-memory-storage.md §5.2）。
//
// 形态（所有标记平铺，不嵌套；同一轮用 run 属性关联）：
//
//	---
//	kind: agent_session
//	session_id: a2215299…
//	book_id: 22
//	title: 改写第三章
//	scope: book
//	permission: confirm
//	runs: 3
//	---
//
//	<!-- xm:session -->
//
//	## 第 1 轮 · 22:40 · 完成
//
//	<!-- xm:run id=r1 at="2026-10-18 22:40:10" status=completed -->
//
//	### 用户
//
//	<!-- xm:input run=r1 -->
//
//	把第三章打斗改紧凑些
//
//	<!-- xm:input-end -->
//
//	### 助手
//
//	<!-- xm:answer run=r1 -->
//
//	（最终回答；用户改写过就用改写版）
//
//	<!-- xm:answer-end -->
//
//	### 工具调用
//
//	<!-- xm:tool run=r1 call=call_1 name=chapter_list state=completed -->
//
//	```json
//	{"book_id": 22}
//	```
//
//	<!-- xm:tool-end -->
//
// 两个刻意的取舍：
//  1. **逐 token 增量（model.delta / thinking.delta）不落盘**：那是打字机过程数据，
//     拼接结果与 model.completed 完全一致，按 token 存档只会让文件膨胀到几十 MB。
//  2. **工具参数在 md 里遮蔽、在库里保持原文**：md 是"可能被分享"的视图，
//     库里的 agent_events / agent_tool_calls 是运行期账本 —— 改了就查不出问题。

const (
	// agentToolArgDisplayChars 是 md 里单条工具参数的最大展示长度。
	agentToolArgDisplayChars = 2000
	// agentInputDisplayChars 是 md 里用户输入的最大展示长度（超长输入只留头部）。
	agentInputDisplayChars = 20000
)

// AgentToolCall 是一轮里的一次工具调用。
type AgentToolCall struct {
	CallID    string
	Name      string
	State     string // completed / error
	Args      string // 原始 JSON（库里是原文；渲染时遮蔽）
	Result    string // 原始 JSON 或错误文本
	ArgsShown string // md 里实际出现的（已遮蔽、可能截断）
	ArgsTrunc bool
}

// AgentRun 是账本里的一轮对话。
type AgentRun struct {
	RunID     string
	Input     string
	Answer    string // 最终回答（answer.edited 优先）
	At        time.Time
	Status    string
	Error     string
	Tools     []AgentToolCall
	Proposals []AgentProposal
}

// AgentProposal 是一次写入提案的终态。
type AgentProposal struct {
	ProposalID string
	Tool       string
	Target     string
	Risk       string
	Summary    string
	Status     string
	ApplyError string
}

// AgentDoc 是一份 Agent 会话文件解析后的内容。
type AgentDoc struct {
	SessionID  string
	BookID     int
	Title      string
	Scope      string
	Permission string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Runs       []AgentRun
	Extra      [][2]string
	RelPath    string
}

// agentFrontKeys 是程序负责的 front-matter 键。
var agentFrontKeys = []string{
	"kind", "session_id", "book_id", "title", "scope", "permission", "runs", "created_at", "updated_at",
}

// ---------------- 路径 ----------------

// AgentFilePath 返回会话文件的绝对路径与相对书目录的路径。
//
// 全局会话（scope=global，没有 book_id）落到库根的 `_memory/agent/`，
// 这样"把作品库整个拷走"能带上所有会话，不只是绑定书的那部分。
func (s *Service) AgentFilePath(sessionID string) (string, string, error) {
	var title, scope, rel string
	var scopeID sql.NullInt64
	var createdRaw string
	err := s.DB.QueryRow(`SELECT COALESCE(title,''), COALESCE(scope_kind,'global'), scope_id,
		COALESCE(file_path,''), COALESCE(created_at,'') FROM agent_sessions WHERE id=?`, sessionID).
		Scan(&title, &scope, &scopeID, &rel, &createdRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", fmt.Errorf("会话不存在")
		}
		return "", "", err
	}

	at := ParseDBTime(createdRaw)
	if at.IsZero() {
		at = time.Now()
	}

	bookID := 0
	if scope == "book" && scopeID.Valid {
		bookID = int(scopeID.Int64)
	}
	if bookID > 0 {
		if strings.TrimSpace(rel) == "" {
			rel, err = MemoryRelPathOf(MemoryKindAgent, at, title, sessionID)
			if err != nil {
				return "", "", err
			}
			if _, err := s.DB.Exec(`UPDATE agent_sessions SET file_path=? WHERE id=?`, rel, sessionID); err != nil {
				return "", "", err
			}
		}
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

	// 全局会话：相对库根的 _memory/agent/
	if strings.TrimSpace(rel) == "" {
		sub, err := MemoryKindDir(MemoryKindAgent)
		if err != nil {
			return "", "", err
		}
		rel = filepath.ToSlash(filepath.Join("_memory", sub,
			MemoryMainNameOf(at, title, MemoryKindAgent, sessionID)+".md"))
		if _, err := s.DB.Exec(`UPDATE agent_sessions SET file_path=? WHERE id=?`, rel, sessionID); err != nil {
			return "", "", err
		}
	}
	root, err := s.Root()
	if err != nil {
		return "", "", err
	}
	abs, err := JoinWithin(root, rel)
	if err != nil {
		return "", "", err
	}
	if err := ensureDir(filepath.Dir(abs)); err != nil {
		return "", "", err
	}
	return abs, rel, nil
}

// MemoryShortCodeOf 是字符串主键版的短码：取 sha256 前 6 位。
//
// 数字主键走 MemoryShortCode，会话 id（32 位十六进制）走这里；
// 两者规则一致（sha256 前 6 位），只是入参形态不同。
func MemoryShortCodeOf(kind, id string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(kind) + "#" + id))
	return hex.EncodeToString(sum[:])[:6]
}

// ---------------- 渲染 ----------------

// RenderAgentDoc 把会话渲染成文件内容。
func RenderAgentDoc(d *AgentDoc) string {
	known := [][2]string{
		{"kind", memoryFileKind(MemoryKindAgent)},
		{"session_id", d.SessionID},
		{"book_id", fmt.Sprintf("%d", d.BookID)},
		{"title", d.Title},
		{"scope", d.Scope},
		{"permission", d.Permission},
		{"runs", fmt.Sprintf("%d", len(d.Runs))},
		{"created_at", FormatMemTime(d.CreatedAt)},
		{"updated_at", FormatMemTime(d.UpdatedAt)},
	}
	var b strings.Builder
	b.WriteString(RenderMemoryFrontMatter(known, d.Extra))
	b.WriteString("\n")
	b.WriteString(RenderMarkerLine("session", nil))
	b.WriteString("\n\n")

	title := strings.TrimSpace(d.Title)
	if title == "" {
		title = "新会话"
	}
	b.WriteString("# " + title + "\n\n")
	b.WriteString("<!-- 自动生成，请勿手改标记；正文可以随便改。 -->\n\n")
	b.WriteString("<!-- 工具参数里的密钥已遮蔽；逐 token 的流式增量不存档。 -->\n\n")

	for i, run := range d.Runs {
		status := run.Status
		if status == "" {
			status = "completed"
		}
		b.WriteString(fmt.Sprintf("## 第 %d 轮 · %s · %s\n\n", i+1, run.At.Format("15:04"), agentStatusLabel(status)))
		b.WriteString(RenderMarkerLine("run", [][2]string{
			{"id", run.RunID},
			{"at", FormatMemTime(run.At)},
			{"status", status},
		}))
		b.WriteString("\n\n")
		if run.Error != "" {
			b.WriteString("> 失败原因：" + run.Error + "\n\n")
		}

		if strings.TrimSpace(run.Input) != "" {
			b.WriteString("### 用户\n\n")
			b.WriteString(RenderPairedMarker("input", [][2]string{{"run", run.RunID}},
				TruncateRunesHelp(run.Input, agentInputDisplayChars)))
		}
		if strings.TrimSpace(run.Answer) != "" {
			b.WriteString("### 助手\n\n")
			b.WriteString(RenderPairedMarker("answer", [][2]string{{"run", run.RunID}}, run.Answer))
		}
		if len(run.Tools) > 0 {
			b.WriteString("### 工具调用\n\n")
			for _, tc := range run.Tools {
				attrs := [][2]string{
					{"run", run.RunID},
					{"call", tc.CallID},
					{"name", tc.Name},
					{"state", tc.State},
				}
				body := ""
				if strings.TrimSpace(tc.ArgsShown) != "" {
					body = "```json\n" + tc.ArgsShown + "\n```"
				}
				b.WriteString(RenderPairedMarker("tool", attrs, body))
				if strings.TrimSpace(tc.Result) != "" {
					b.WriteString(RenderPairedMarker("tool-result",
						[][2]string{{"call", tc.CallID}},
						"```json\n"+TruncateRunesHelp(tc.Result, agentToolArgDisplayChars)+"\n```"))
				}
			}
		}
		if len(run.Proposals) > 0 {
			// 标题只写一次：循环里写会出现"每有一个提案就多一个 ### 写入提案"
			b.WriteString("### 写入提案\n\n")
		}
		for _, p := range run.Proposals {
			attrs := [][2]string{
				{"run", run.RunID},
				{"id", p.ProposalID},
				{"tool", p.Tool},
				{"status", p.Status},
			}
			if p.Target != "" {
				attrs = append(attrs, [2]string{"target", p.Target})
			}
			if p.Risk != "" {
				attrs = append(attrs, [2]string{"risk", p.Risk})
			}
			// 提案是自闭合标记 + 一份人读的要点清单：
			// 标记供程序还原，清单供人快速看懂"它想改什么、有没有通过"。
			b.WriteString(RenderMarkerLine("proposal", attrs))
			b.WriteString("\n\n")
			b.WriteString("- 工具：`" + p.Tool + "`\n")
			b.WriteString("- 状态：" + agentProposalLabel(p.Status) + "\n")
			if p.Target != "" {
				b.WriteString("- 目标：" + p.Target + "\n")
			}
			if p.Risk != "" {
				b.WriteString("- 风险：" + p.Risk + "\n")
			}
			if strings.TrimSpace(p.Summary) != "" {
				b.WriteString("- 摘要：" + p.Summary + "\n")
			}
			if p.ApplyError != "" {
				b.WriteString("- 应用失败：" + p.ApplyError + "\n")
			}
			b.WriteString("\n")
		}
	}
	return strings.TrimRight(b.String(), "\n") + "\n"
}

func agentStatusLabel(status string) string {
	switch status {
	case "completed":
		return "完成"
	case "failed":
		return "失败"
	case "cancelled":
		return "已取消"
	case "running":
		return "进行中"
	default:
		return status
	}
}

func agentProposalLabel(status string) string {
	switch status {
	case "pending":
		return "待确认"
	case "applied":
		return "已应用"
	case "rejected":
		return "已拒绝"
	case "rolled_back":
		return "已回滚"
	case "apply_failed":
		return "应用失败"
	default:
		return status
	}
}

// ---------------- 解析 ----------------

// ParseAgentDoc 解析 Agent 会话文件。
func ParseAgentDoc(text string) (*AgentDoc, error) {
	fm, err := ParseMemoryFrontMatter(text)
	if err != nil {
		return nil, err
	}
	pairs := fm.Pairs()
	vals := KnownFields(pairs, agentFrontKeys...)
	d := &AgentDoc{
		Extra:      ExtraFields(pairs, agentFrontKeys...),
		SessionID:  vals["session_id"],
		Title:      vals["title"],
		Scope:      vals["scope"],
		Permission: vals["permission"],
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

	// 先收 run 元数据（自闭合），再按 run 属性把各块归位
	runIdx := map[string]int{}
	for _, mk := range FindMarkers(markers, "run") {
		id := mk.Attr("id")
		if id == "" {
			continue
		}
		runIdx[id] = len(d.Runs)
		d.Runs = append(d.Runs, AgentRun{
			RunID:  id,
			At:     ParseMemTime(mk.Attr("at")),
			Status: mk.Attr("status"),
		})
	}

	ensureRun := func(runID string) int {
		if i, ok := runIdx[runID]; ok {
			return i
		}
		d.Runs = append(d.Runs, AgentRun{RunID: runID})
		runIdx[runID] = len(d.Runs) - 1
		return runIdx[runID]
	}

	for _, mk := range markers {
		runID := mk.Attr("run")
		switch mk.Name {
		case "input":
			i := ensureRun(runID)
			d.Runs[i].Input = mk.Body
		case "answer":
			i := ensureRun(runID)
			d.Runs[i].Answer = mk.Body
		case "tool":
			i := ensureRun(runID)
			d.Runs[i].Tools = append(d.Runs[i].Tools, AgentToolCall{
				CallID:    mk.Attr("call"),
				Name:      mk.Attr("name"),
				State:     mk.Attr("state"),
				ArgsShown: fencedJSON(mk.Body),
			})
		case "tool-result":
			// 结果块用 call 关联到上面那次调用
			callID := mk.Attr("call")
			for ri := range d.Runs {
				for ti := range d.Runs[ri].Tools {
					if d.Runs[ri].Tools[ti].CallID == callID {
						d.Runs[ri].Tools[ti].Result = fencedJSON(mk.Body)
					}
				}
			}
		case "proposal":
			i := ensureRun(runID)
			d.Runs[i].Proposals = append(d.Runs[i].Proposals, AgentProposal{
				ProposalID: mk.Attr("id"),
				Tool:       mk.Attr("tool"),
				Target:     mk.Attr("target"),
				Risk:       mk.Attr("risk"),
				Status:     mk.Attr("status"),
			})
		}
	}
	return d, nil
}

// fencedJSON 从 ```json 围栏里取出内容；没有围栏就返回原样。
func fencedJSON(body string) string {
	s := strings.TrimSpace(body)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return s
	}
	lines = lines[1:]
	if n := len(lines); n > 0 && strings.HasPrefix(strings.TrimSpace(lines[n-1]), "```") {
		lines = lines[:n-1]
	}
	return strings.Join(lines, "\n")
}

func strconvAtoiSafe(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n, err
}

// ---------------- 从账本读 + 渲染 ----------------

// ledgerEvent 是渲染会话时用到的最小账本条目（与 agent.LedgerEvent 同构）。
type ledgerEvent struct {
	Seq       int64
	Type      string
	RunID     string
	Payload   json.RawMessage
	CreatedAt string
}

// BuildAgentDocFromLedger 把账本事件折成一份会话文档。
//
// 这是"从库里渲染出 md"的唯一入口：Agent 的写入点只调 MarkMemoryDirty，
// 真正的渲染在这里做，保证内容与账本一致。
func (s *Service) BuildAgentDocFromLedger(sessionID string) (*AgentDoc, error) {
	var d AgentDoc
	var scopeID sql.NullInt64
	var createdRaw, updatedRaw string
	err := s.DB.QueryRow(`SELECT id, COALESCE(title,''), COALESCE(scope_kind,'global'), scope_id,
		COALESCE(permission,'confirm'), COALESCE(created_at,''), COALESCE(updated_at,'')
		FROM agent_sessions WHERE id=?`, sessionID).
		Scan(&d.SessionID, &d.Title, &d.Scope, &scopeID, &d.Permission, &createdRaw, &updatedRaw)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("会话不存在")
		}
		return nil, err
	}
	if d.Scope == "book" && scopeID.Valid {
		d.BookID = int(scopeID.Int64)
	}
	d.CreatedAt = ParseDBTime(createdRaw)
	if d.CreatedAt.IsZero() {
		d.CreatedAt = time.Now()
	}
	d.UpdatedAt = ParseDBTime(updatedRaw)

	events, err := s.agentLedgerEvents(sessionID)
	if err != nil {
		return nil, err
	}

	runIdx := map[string]int{}
	// 每轮的时间取该轮**首个事件**的落库时间：账本事件本身不带时间戳，
	// 而"什么时候开始的"只能从事件行的 created_at 推出来。
	ensureRun := func(runID, createdAt string) int {
		if i, ok := runIdx[runID]; ok {
			return i
		}
		d.Runs = append(d.Runs, AgentRun{
			RunID:  runID,
			Status: "completed",
			At:     ParseDBTime(createdAt),
		})
		runIdx[runID] = len(d.Runs) - 1
		return runIdx[runID]
	}

	for _, ev := range events {
		switch ev.Type {
		case "run.created":
			var p struct {
				Input string `json:"input"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			d.Runs[i].Input = p.Input
		case "run.completed", "run.status":
			// 终态有两个来源：loop 结束时记 run.completed，
			// setRunStatus（取消/失败/进行中）记 run.status。两者都要认。
			var p struct {
				Status string `json:"status"`
				Error  string `json:"error"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			if p.Status != "" {
				d.Runs[i].Status = p.Status
			}
			if p.Error != "" {
				d.Runs[i].Error = p.Error
			}
		case "model.completed":
			var p struct {
				Text string `json:"text"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			d.Runs[i].Answer = p.Text
		case "answer.edited":
			var p struct {
				Text string `json:"text"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			// 用户改过的版本才是"最终回答"
			d.Runs[i].Answer = p.Text
		case "tool.started":
			var p struct {
				CallID    string          `json:"call_id"`
				Name      string          `json:"name"`
				Arguments json.RawMessage `json:"arguments"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			args := strings.TrimSpace(string(p.Arguments))
			if args == "" {
				args = "{}"
			}
			shown, truncated := redactAndTruncate(args, agentToolArgDisplayChars)
			d.Runs[i].Tools = append(d.Runs[i].Tools, AgentToolCall{
				CallID: p.CallID, Name: p.Name, State: "completed",
				Args: args, ArgsShown: shown, ArgsTrunc: truncated,
			})
		case "tool.completed":
			var p struct {
				CallID string          `json:"call_id"`
				Result json.RawMessage `json:"result"`
				Error  string          `json:"error"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			for ti := range d.Runs[i].Tools {
				tc := &d.Runs[i].Tools[ti]
				if tc.CallID != p.CallID {
					continue
				}
				if p.Error != "" {
					tc.State = "error"
					tc.Result = p.Error
				} else {
					tc.Result = strings.TrimSpace(string(p.Result))
				}
			}
		case "proposal.created":
			var p struct {
				ProposalID string `json:"proposal_id"`
				Tool       string `json:"tool"`
				Target     string `json:"target"`
				Risk       string `json:"risk"`
				Summary    string `json:"summary"`
				Status     string `json:"status"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			d.Runs[i].Proposals = append(d.Runs[i].Proposals, AgentProposal{
				ProposalID: p.ProposalID, Tool: p.Tool, Target: p.Target,
				Risk: p.Risk, Summary: p.Summary, Status: p.Status,
			})
		case "proposal.decided":
			var p struct {
				ProposalID string `json:"proposal_id"`
				Status     string `json:"status"`
				ApplyError string `json:"apply_error"`
			}
			if json.Unmarshal(ev.Payload, &p) != nil {
				continue
			}
			i := ensureRun(ev.RunID, ev.CreatedAt)
			for pi := range d.Runs[i].Proposals {
				if d.Runs[i].Proposals[pi].ProposalID == p.ProposalID {
					d.Runs[i].Proposals[pi].Status = p.Status
					d.Runs[i].Proposals[pi].ApplyError = p.ApplyError
				}
			}
		}
	}
	return &d, nil
}

// redactAndTruncate 遮蔽密钥并按字符截断，返回展示文本与是否截断过。
func redactAndTruncate(raw string, max int) (string, bool) {
	redacted := RedactSecrets(raw)
	if max <= 0 || utf8Len(redacted) <= max {
		return redacted, false
	}
	return string([]rune(redacted)[:max]) + "…", true
}

func utf8Len(s string) int { return len([]rune(s)) }

// agentLedgerEvents 读某会话的账本事件（按 seq 升序）。
func (s *Service) agentLedgerEvents(sessionID string) ([]ledgerEvent, error) {
	rows, err := s.DB.Query(`SELECT seq, type, COALESCE(run_id,''), payload, COALESCE(created_at,'')
		FROM agent_events WHERE session_id=? ORDER BY seq ASC`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ledgerEvent
	for rows.Next() {
		var e ledgerEvent
		var payload string
		if err := rows.Scan(&e.Seq, &e.Type, &e.RunID, &payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = json.RawMessage(payload)
		list = append(list, e)
	}
	return list, rows.Err()
}

// ---------------- 落盘 ----------------

// MirrorAgentSession 把会话渲染落盘（flusher 与创建时都用它）。
func (s *Service) MirrorAgentSession(sessionID string) error {
	doc, err := s.BuildAgentDocFromLedger(sessionID)
	if err != nil {
		// 会话已删除：不做任何事（删除路径已经清过文件）
		if strings.Contains(err.Error(), "会话不存在") {
			return nil
		}
		return err
	}
	abs, rel, err := s.AgentFilePath(sessionID)
	if err != nil {
		return err
	}

	// 文件是真相，程序不覆盖被外部改过的文件。
	// 这里只拦"程序自己盖掉用户改动"这一种情况；用户把标记改坏了属于他自己承担，
	// 但至少要让他知道这次没写进去。
	var recorded string
	s.DB.QueryRow(`SELECT COALESCE(content_hash,'') FROM agent_sessions WHERE id=?`, sessionID).Scan(&recorded)
	if changed, _, herr := FileEditedExternally(abs, recorded); herr == nil && changed {
		s.DB.Exec(`UPDATE agent_sessions SET parse_error=? WHERE id=?`, ExternalEditReason, sessionID)
		return nil
	}

	// 改名跟随标题，与章节、图谱一致（短码稳定，所以身份不变）
	if wantRel, ok := ReconcileMemoryFileName(rel, doc.Title); ok {
		if newAbs, newRel, renamed := RenameMemoryFile(abs, rel, wantRel); renamed {
			abs, rel = newAbs, newRel
			s.DB.Exec(`UPDATE agent_sessions SET file_path=? WHERE id=?`, rel, sessionID)
		}
	}
	doc.RelPath = rel

	// 保留用户自己加进 front-matter 的键（已有文件才谈得上保留）
	if raw, rerr := os.ReadFile(abs); rerr == nil {
		if old, perr := ParseMemoryFrontMatter(string(raw)); perr == nil {
			doc.Extra = ExtraFields(old.Pairs(), agentFrontKeys...)
		}
	}
	meta, err := WriteTextFile(abs, RenderAgentDoc(doc))
	if err != nil {
		return err
	}
	s.DB.Exec(`UPDATE agent_sessions SET file_path=?, content_hash=?, missing=0, parse_error='',
		synced_at=CURRENT_TIMESTAMP WHERE id=?`, rel, meta.Hash, sessionID)
	return nil
}

// TrashAgentFile 把会话文件移入回收站。
func (s *Service) TrashAgentFile(sessionID string) error {
	var scope string
	var scopeID sql.NullInt64
	var rel string
	if err := s.DB.QueryRow(`SELECT COALESCE(scope_kind,'global'), scope_id, COALESCE(file_path,'')
		FROM agent_sessions WHERE id=?`, sessionID).Scan(&scope, &scopeID, &rel); err != nil {
		return nil
	}
	if strings.TrimSpace(rel) == "" {
		return nil
	}
	if scope == "book" && scopeID.Valid && scopeID.Int64 > 0 {
		return s.trashBookFile(int(scopeID.Int64), rel)
	}
	// 全局会话：库根下，走同一套回收逻辑（借用一个不存在的书 id 不成立，这里直接处理）
	root, err := s.Root()
	if err != nil {
		return err
	}
	src, err := JoinWithin(root, rel)
	if err != nil {
		return err
	}
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	if err := moveToRecycleBin(src); err == nil {
		return nil
	}
	trashDir := filepath.Join(root, "_trash", "_memory", time.Now().Format("20060102-150405"))
	if err := ensureDir(trashDir); err != nil {
		return err
	}
	return os.Rename(src, filepath.Join(trashDir, filepath.Base(src)))
}
