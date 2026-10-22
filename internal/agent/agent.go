// Package agent 实现星芒 Agent：Go 原生的最小 agent 内核。
//
// 设计要点（对照 docs/agent-migration-plan.md）：
//   - 账本即历史：会话消息不单独存表，上下文装配从 agent_events 按 seq 重放。
//   - seq 每会话严格递增，UNIQUE(session_id, seq) 是断线续传幂等的根基。
//   - 第一版只读工具（书/章节/角色/图谱），写操作走后续版本的提案-审批。
//   - 只支持 OpenAI 兼容格式（api_format 为空/openai/anthropic_compat）的 function calling。
package agent

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// 事件类型（SSE 的 event 名与账本 type 字段一致）
const (
	EventRunCreated      = "run.created"
	EventRunStatus       = "run.status"
	EventModelDelta      = "model.delta"
	EventThinkingDelta   = "thinking.delta"
	EventModelDone       = "model.completed"
	EventAnswerEdited    = "answer.edited"
	EventToolStarted     = "tool.started"
	EventToolCompleted   = "tool.completed"
	EventRunCompleted    = "run.completed"
	EventProposalCreated = "proposal.created"
	EventProposalDecided = "proposal.decided"
)

const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
	StatusFailed    = "failed"
)

// 写入权限三档（方案 B）
const (
	PermObserve = "observe" // 试运行：只出提案，永不落库
	PermConfirm = "confirm" // 每次确认：全部写入走提案-审批
	PermAuto    = "auto"    // 自动执行：低风险直写，高风险仍走审批
)

// 提案状态
const (
	ProposalPending     = "pending"
	ProposalApplied     = "applied"
	ProposalRejected    = "rejected"
	ProposalRolledBack  = "rolled_back"
	ProposalApplyFailed = "apply_failed"
)

// maxSteps 单个 run 的模型调用轮数上限（模型→工具→模型记一轮），
// 防止模型反复调工具不停机（执行预算的第一道闸，token 预算后续版本再补）。
const maxSteps = 12

// Manager 管理 agent 会话与运行中的 run。
type Manager struct {
	db *sql.DB

	appendMu sync.Mutex // 保证同会话 seq 分配的顺序性

	mu        sync.Mutex
	cancels   map[string]context.CancelFunc // runID -> cancel
	busy      map[string]bool               // sessionID -> 是否有 run 在跑
	sessionOf map[string]string             // runID -> sessionID
}

// NewManager 创建 Manager，并把上次进程中断遗留的 running 孤儿 run 标记为失败
// （崩溃恢复：孤儿 run 不允许静默丢失）。
func NewManager(db *sql.DB) *Manager {
	m := &Manager{
		db:        db,
		cancels:   map[string]context.CancelFunc{},
		busy:      map[string]bool{},
		sessionOf: map[string]string{},
	}
	db.Exec(`UPDATE agent_runs SET status='failed', error='服务重启导致执行中断', updated_at=CURRENT_TIMESTAMP WHERE status IN ('running')`)
	return m
}

// newID 生成 32 位十六进制随机 ID（不引入 uuid 依赖）。
func newID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// AppendEvent 追加一条账本事件，返回分配的 seq。
// seq = 同会话 MAX(seq)+1；撞 UNIQUE 即重试（理论上 appendMu 内不会发生，防御性保留）。
func (m *Manager) AppendEvent(sessionID, runID, evType string, payload any) (int64, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	m.appendMu.Lock()
	defer m.appendMu.Unlock()
	for attempt := 0; attempt < 3; attempt++ {
		var seq int64
		if err := m.db.QueryRow(
			"SELECT COALESCE(MAX(seq),0)+1 FROM agent_events WHERE session_id=?", sessionID).Scan(&seq); err != nil {
			return 0, err
		}
		_, err := m.db.Exec(
			"INSERT INTO agent_events (session_id, run_id, seq, type, payload) VALUES (?,?,?,?,?)",
			sessionID, runID, seq, evType, string(p))
		if err == nil {
			markSessionMemoryDirty(sessionID, evType)
			return seq, nil
		}
		if attempt == 2 {
			return 0, fmt.Errorf("写入账本失败: %w", err)
		}
	}
	return 0, fmt.Errorf("unreachable")
}

// markSessionMemoryDirty 把会话标脏，等后台合并写入器渲染落盘。
//
// 逐 token 的 model.delta / thinking.delta 不参与：它们不落盘（见 bookstore/memory_agent.go），
// 每来一个 token 都标一次脏会让写入器一直在队列里空转。
func markSessionMemoryDirty(sessionID, evType string) {
	if sessionID == "" {
		return
	}
	switch evType {
	case EventModelDelta, EventThinkingDelta:
		return
	}
	bookstore.MarkMemoryDirty(bookstore.MemoryKindAgent, sessionID)
}

// setRunStatus 更新 run 状态并同步推送 run.status 事件。
func (m *Manager) setRunStatus(sessionID, runID, status, errMsg string) {
	if errMsg != "" {
		m.db.Exec("UPDATE agent_runs SET status=?, error=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", status, errMsg, runID)
	} else {
		m.db.Exec("UPDATE agent_runs SET status=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", status, runID)
	}
	payload := map[string]string{"run_id": runID, "status": status}
	if errMsg != "" {
		payload["error"] = errMsg
	}
	m.AppendEvent(sessionID, runID, EventRunStatus, payload)
}

// LedgerEvent 是流式接口吐出的账本条目。
type LedgerEvent struct {
	Seq       int64           `json:"seq"`
	Type      string          `json:"type"`
	RunID     string          `json:"run_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt string          `json:"created_at"`
}

// EventsAfter 返回某会话 seq 大于 after 的事件（按 seq 升序）。
func (m *Manager) EventsAfter(sessionID string, after int64) ([]LedgerEvent, error) {
	rows, err := m.db.Query(
		"SELECT seq, type, COALESCE(run_id,''), payload, COALESCE(created_at,'') FROM agent_events WHERE session_id=? AND seq>? ORDER BY seq ASC",
		sessionID, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []LedgerEvent
	for rows.Next() {
		var e LedgerEvent
		// ⚠ payload 不能直接 Scan 进 json.RawMessage：modernc 驱动把 TEXT 返回为
		// string，database/sql 不支持 string → *json.RawMessage（会静默失败）。先收 string。
		var payload string
		if err := rows.Scan(&e.Seq, &e.Type, &e.RunID, &payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = json.RawMessage(payload)
		list = append(list, e)
	}
	return list, rows.Err()
}

// Session 是 agent 会话的存储形态。
type Session struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	ScopeKind  string `json:"scope_kind"`
	ScopeID    *int64 `json:"scope_id"`
	ModelID    *int64 `json:"model_id"`
	Permission string `json:"permission"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// GetSession 按 ID 取会话。
func (m *Manager) GetSession(id string) (*Session, error) {
	var s Session
	err := m.db.QueryRow(
		"SELECT id, title, scope_kind, scope_id, model_id, COALESCE(permission,'confirm'), created_at, updated_at FROM agent_sessions WHERE id=?", id).
		Scan(&s.ID, &s.Title, &s.ScopeKind, &s.ScopeID, &s.ModelID, &s.Permission, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("会话不存在")
	}
	return &s, err
}

// UpdateSessionPermission 修改会话写入权限档位。
func (m *Manager) UpdateSessionPermission(id, permission string) error {
	if permission != PermObserve && permission != PermConfirm && permission != PermAuto {
		return fmt.Errorf("permission 只能是 observe / confirm / auto")
	}
	res, err := m.db.Exec("UPDATE agent_sessions SET permission=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", permission, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("会话不存在")
	}
	// 权限写在文件 front-matter 里，改了就要重渲染
	markSessionMemoryDirty(id, "session.permission")
	return nil
}

// CreateSession 建会话；scope 为 book 时校验书存在。
func (m *Manager) CreateSession(title, scopeKind string, scopeID, modelID *int64, permission string) (*Session, error) {
	if scopeKind != "global" && scopeKind != "book" {
		return nil, fmt.Errorf("scope_kind 只能是 global 或 book")
	}
	if permission == "" {
		permission = PermConfirm
	}
	if permission != PermObserve && permission != PermConfirm && permission != PermAuto {
		return nil, fmt.Errorf("permission 只能是 observe / confirm / auto")
	}
	if scopeKind == "book" {
		if scopeID == nil || *scopeID <= 0 {
			return nil, fmt.Errorf("book 作用域必须指定 scope_id")
		}
		var exists int
		if err := m.db.QueryRow("SELECT 1 FROM books WHERE id=?", *scopeID).Scan(&exists); err != nil {
			return nil, fmt.Errorf("绑定的书不存在（id=%d）", *scopeID)
		}
	}
	s := &Session{ID: newID(), Title: title, ScopeKind: scopeKind, ScopeID: scopeID, ModelID: modelID, Permission: permission}
	if s.Title == "" {
		s.Title = "新会话"
	}
	_, err := m.db.Exec(
		"INSERT INTO agent_sessions (id, title, scope_kind, scope_id, model_id, permission) VALUES (?,?,?,?,?,?)",
		s.ID, s.Title, s.ScopeKind, s.ScopeID, s.ModelID, s.Permission)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListSessions 返回全部会话（按更新时间倒序）。
func (m *Manager) ListSessions() ([]Session, error) {
	rows, err := m.db.Query("SELECT id, title, scope_kind, scope_id, model_id, COALESCE(permission,'confirm'), created_at, updated_at FROM agent_sessions ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.Title, &s.ScopeKind, &s.ScopeID, &s.ModelID, &s.Permission, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

// DeleteSession 删除会话（事件/run/工具调用沿外键级联清理）。
func (m *Manager) DeleteSession(id string) error {
	_, err := m.db.Exec("DELETE FROM agent_sessions WHERE id=?", id)
	return err
}

// TouchSession 刷新 updated_at（新消息时保持会话列表排序）。
//
// updated_at 会写进记忆文件的 front-matter，所以这里必须标脏：
// 否则文件里的"最后更新"会永远停在首次落盘那一刻，看起来像这个会话再没动过。
func (m *Manager) TouchSession(id string) {
	if _, err := m.db.Exec("UPDATE agent_sessions SET updated_at=CURRENT_TIMESTAMP WHERE id=?", id); err != nil {
		return
	}
	markSessionMemoryDirty(id, "session.touch")
}

// EditRunAnswer 修改某段对话的助手回答。
//
// 账本是 append-only，所以不覆写 model.completed，而是追加一条 answer.edited 事件；
// buildMessages 重放时会用编辑后的文本替换该段的回答，因此**后续对话按修改后的内容作为上下文**。
func (m *Manager) EditRunAnswer(runID, text string) error {
	if text == "" {
		return fmt.Errorf("回答内容不能为空")
	}
	var sessionID string
	err := m.db.QueryRow("SELECT session_id FROM agent_runs WHERE id=?", runID).Scan(&sessionID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("对话不存在或已删除")
	}
	if err != nil {
		return err
	}
	if _, err := m.AppendEvent(sessionID, runID, EventAnswerEdited, map[string]string{"run_id": runID, "text": text}); err != nil {
		return err
	}
	m.TouchSession(sessionID)
	return nil
}

// ===== 写入提案 =====

// Proposal 是一次写入操作的提案记录。
type Proposal struct {
	ID         string  `json:"id"`
	SessionID  string  `json:"session_id"`
	RunID      string  `json:"run_id"`
	ToolCallID string  `json:"tool_call_id"`
	ToolName   string  `json:"tool_name"`
	Target     string  `json:"target"`
	Arguments  string  `json:"arguments"`
	Risk       string  `json:"risk"`
	Summary    string  `json:"summary"`
	Status     string  `json:"status"`
	BeforeText *string `json:"before_text"`
	AfterText  *string `json:"after_text"`
	ApplyError string  `json:"apply_error"`
	CreatedAt  string  `json:"created_at"`
	DecidedAt  string  `json:"decided_at"`
	AppliedAt  string  `json:"applied_at"`
}

// CreateProposal 落一条提案并广播 proposal.created 事件。
func (m *Manager) CreateProposal(s *Session, runID, toolCallID, toolName, target, argsJSON, risk, summary string) (*Proposal, error) {
	p := &Proposal{
		ID: newID(), SessionID: s.ID, RunID: runID, ToolCallID: toolCallID,
		ToolName: toolName, Target: target, Arguments: argsJSON, Risk: risk, Summary: summary,
		Status: ProposalPending,
	}
	_, err := m.db.Exec(`INSERT INTO agent_change_proposals
		(id, session_id, run_id, tool_call_id, tool_name, target, arguments, risk, summary, status)
		VALUES (?,?,?,?,?,?,?,?,?,?)`,
		p.ID, p.SessionID, p.RunID, p.ToolCallID, p.ToolName, p.Target, p.Arguments, p.Risk, p.Summary, p.Status)
	if err != nil {
		return nil, err
	}
	m.AppendEvent(s.ID, runID, EventProposalCreated, map[string]any{
		"proposal_id": p.ID, "tool": p.ToolName, "target": p.Target,
		"risk": p.Risk, "summary": p.Summary, "status": p.Status, "arguments": json.RawMessage(argsJSON),
	})
	return p, nil
}

// getProposal 按 ID 取提案。
func (m *Manager) getProposal(id string) (*Proposal, error) {
	var p Proposal
	var before, after sql.NullString
	err := m.db.QueryRow(`SELECT id, session_id, COALESCE(run_id,''), COALESCE(tool_call_id,''), tool_name, target,
		arguments, risk, summary, status, before_text, after_text, COALESCE(apply_error,''),
		COALESCE(created_at,''), COALESCE(decided_at,''), COALESCE(applied_at,'')
		FROM agent_change_proposals WHERE id=?`, id).
		Scan(&p.ID, &p.SessionID, &p.RunID, &p.ToolCallID, &p.ToolName, &p.Target,
			&p.Arguments, &p.Risk, &p.Summary, &p.Status, &before, &after, &p.ApplyError,
			&p.CreatedAt, &p.DecidedAt, &p.AppliedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("提案不存在")
	}
	if err != nil {
		return nil, err
	}
	if before.Valid {
		p.BeforeText = &before.String
	}
	if after.Valid {
		p.AfterText = &after.String
	}
	return &p, nil
}

// decideProposal 终态迁移 + 事件广播（approve/reject/rollback 共用）。
func (m *Manager) decideProposal(p *Proposal, status, applyErr string) {
	if applyErr != "" {
		m.db.Exec(`UPDATE agent_change_proposals SET status=?, apply_error=?, decided_at=CURRENT_TIMESTAMP WHERE id=?`,
			status, applyErr, p.ID)
	} else {
		m.db.Exec(`UPDATE agent_change_proposals SET status=?, apply_error=NULL, decided_at=CURRENT_TIMESTAMP WHERE id=?`,
			status, p.ID)
	}
	payload := map[string]any{
		"proposal_id": p.ID, "tool": p.ToolName, "target": p.Target,
		"risk": p.Risk, "summary": p.Summary, "status": status,
	}
	if applyErr != "" {
		payload["apply_error"] = applyErr
	}
	m.AppendEvent(p.SessionID, p.RunID, EventProposalDecided, payload)
}

// GetProposal 按 ID 取提案（handler 用）。
func (m *Manager) GetProposal(id string) (*Proposal, error) {
	return m.getProposal(id)
}

// RejectProposal 拒绝提案。
func (m *Manager) RejectProposal(id string) error {
	p, err := m.getProposal(id)
	if err != nil {
		return err
	}
	if p.Status != ProposalPending {
		return fmt.Errorf("提案当前状态是 %s，不能拒绝", p.Status)
	}
	m.decideProposal(p, ProposalRejected, "")
	return nil
}

// RollbackProposal 回滚已应用的章节类提案（用 before_text 快照还原）。
func (m *Manager) RollbackProposal(id string) error {
	p, err := m.getProposal(id)
	if err != nil {
		return err
	}
	if p.Status != ProposalApplied {
		return fmt.Errorf("只有已写入的提案才能回滚（当前 %s）", p.Status)
	}
	if p.BeforeText == nil {
		return fmt.Errorf("该提案没有原文快照，无法回滚")
	}
	switch p.ToolName {
	case "chapter_append", "chapter_replace":
		var ac argsCarrier
		if err := json.Unmarshal([]byte(p.Arguments), &ac); err != nil {
			return fmt.Errorf("提案参数损坏")
		}
		chID, _ := toInt(ac["chapter_id"])
		if chID <= 0 {
			return fmt.Errorf("提案参数缺少 chapter_id")
		}
		// 回滚同样是"写正文"：必须落文件，不能只改数据库。
		if err := bookstore.NewService(m.db).SetChapterText(int(chID), *p.BeforeText); err != nil {
			return fmt.Errorf("回滚失败（章节可能已被删除）: %w", err)
		}
		m.decideProposal(p, ProposalRolledBack, "")
		return nil
	default:
		return fmt.Errorf("该提案类型不支持回滚")
	}
}

// argsCarrier 是提案 arguments 的通用载体。
type argsCarrier map[string]any

// DeleteRun 删除一个 run（用户消息 + 助手回复 + 工具调用 + 状态事件），
// 并在同一事务内把该会话后续事件的 seq 前移，保持 seq 严格连续。
//
// 安全性：同一会话同时只有一个 run，一个 run 的事件在账本上是连续区间，
// 因此删除后整体前移不会撞 UNIQUE(session_id, seq)（目标区间已全部腾空）。
// 前端收到 seq 空洞会全量重拉，删除后由前端 fullReload 重新对齐游标。
func (m *Manager) DeleteRun(runID string) error {
	var sessionID, status string
	err := m.db.QueryRow("SELECT session_id, status FROM agent_runs WHERE id=?", runID).Scan(&sessionID, &status)
	if err == sql.ErrNoRows {
		return fmt.Errorf("运行不存在或已删除")
	}
	if err != nil {
		return err
	}
	if status == StatusRunning {
		return fmt.Errorf("该对话正在执行中，请先停止再删除")
	}

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var minSeq, maxSeq sql.NullInt64
	if err := tx.QueryRow(
		"SELECT MIN(seq), MAX(seq) FROM agent_events WHERE session_id=? AND run_id=?", sessionID, runID,
	).Scan(&minSeq, &maxSeq); err != nil {
		return err
	}

	if minSeq.Valid {
		count := maxSeq.Int64 - minSeq.Int64 + 1
		if _, err := tx.Exec("DELETE FROM agent_events WHERE session_id=? AND run_id=?", sessionID, runID); err != nil {
			return err
		}
		if _, err := tx.Exec(
			"UPDATE agent_events SET seq=seq-? WHERE session_id=? AND seq>?", count, sessionID, maxSeq.Int64,
		); err != nil {
			return err
		}
	}

	// 删 run 本体（agent_tool_calls 靠外键级联清理）
	if _, err := tx.Exec("DELETE FROM agent_runs WHERE id=?", runID); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	// 事件被删了，文件里的这一轮还在 —— 不重渲染的话，从目录恢复会把删掉的对话找回来。
	markSessionMemoryDirty(sessionID, "run.delete")
	return nil
}

// touchSessionIfDefault 用首条输入给默认标题命名，只改一次。
func (m *Manager) touchSessionIfDefault(s *Session, input string) {
	if s != nil && s.Title == "新会话" {
		runes := []rune(input)
		if len(runes) > 24 {
			runes = runes[:24]
		}
		title := string(runes)
		m.db.Exec("UPDATE agent_sessions SET title=?, updated_at=CURRENT_TIMESTAMP WHERE id=?", title, s.ID)
		// 标题写在文件 front-matter 与文件名里，改了要重渲染
		markSessionMemoryDirty(s.ID, "session.title")
	} else {
		m.TouchSession(s.ID)
	}
}
