package agent

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/bookstore"
	"github.com/xingmang/xingmang-desktop/internal/database"
)

// M2 的接线验证：AppendEvent → MarkMemoryDirty → 合并写入器 → 磁盘文件。
//
// 光测渲染函数（bookstore 里的 TestAgentSession_*）测不出这条链路，
// 而它恰恰是最容易漏的地方：事件写进账本了，文件却一直是空的。
func TestAgentMemoryWiring(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "agentmem.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO books (title, status, folder_path, tags) VALUES ('接线书','draft','接线书','[]')`); err != nil {
		t.Fatal(err)
	}
	var bookID int
	db.QueryRow(`SELECT id FROM books WHERE title='接线书'`).Scan(&bookID)

	// 与 server.Start 完全相同的接线（用的是同一个包级写入器）
	svc := bookstore.NewService(db)
	bookstore.RegisterMemoryRenderer(bookstore.MemoryKindAgent, svc.MirrorAgentSession)

	m := NewManager(db)
	bid := int64(bookID)
	s, err := m.CreateSession("改写第三章", "book", &bid, nil, PermConfirm)
	if err != nil {
		t.Fatal(err)
	}

	// 造一轮完整事件；中间夹一条 delta，它不应参与落盘
	const deltaText = "DELTA_SHOULD_NOT_APPEAR"
	runID := "run-1"
	if _, err := m.AppendEvent(s.ID, runID, EventRunCreated,
		map[string]string{"run_id": runID, "input": "把打斗改紧凑些"}); err != nil {
		t.Fatal(err)
	}
	m.AppendEvent(s.ID, runID, EventModelDelta, map[string]string{"run_id": runID, "text": deltaText})
	m.AppendEvent(s.ID, runID, EventThinkingDelta, map[string]string{"run_id": runID, "text": deltaText})
	m.AppendEvent(s.ID, runID, EventToolStarted, map[string]any{
		"run_id": runID, "call_id": "c1", "name": "chapter_list",
		"arguments": json.RawMessage(`{"book_id":1}`),
	})
	m.AppendEvent(s.ID, runID, EventToolCompleted, map[string]any{
		"run_id": runID, "call_id": "c1", "name": "chapter_list",
		"result": json.RawMessage(`{"ok":true}`),
	})
	m.AppendEvent(s.ID, runID, EventModelDone, map[string]string{"run_id": runID, "text": "建议合并这两章。"})
	m.AppendEvent(s.ID, runID, EventRunCompleted, map[string]string{"run_id": runID, "status": StatusCompleted})

	if got := bookstore.PendingMemoryWrites(); got == 0 {
		t.Fatal("写事件后应当有待落盘项")
	}
	if _, err := bookstore.DefaultFlusher().FlushAll(); err != nil {
		t.Fatalf("落盘失败: %v", err)
	}

	var rel string
	db.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, s.ID).Scan(&rel)
	if rel == "" {
		t.Fatal("落盘后应写回 file_path")
	}
	abs := filepath.Join(root, "接线书", filepath.FromSlash(rel))
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("会话文件不存在: %v", err)
	}
	text := string(raw)

	for _, want := range []string{
		"kind: agent_session",
		"session_id: " + s.ID,
		"把打斗改紧凑些",
		"建议合并这两章。",
		"chapter_list",
		"```json",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("文件里缺少 %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, deltaText) {
		t.Fatalf("逐 token 增量不应落盘:\n%s", text)
	}
	if n := strings.Count(text, "## 第 "); n != 1 {
		t.Fatalf("应渲染成 1 轮，实得 %d:\n%s", n, text)
	}
	// 轮次时间要来自账本事件的落库时间，不能是零值
	if strings.Contains(text, "0001-01-01") {
		t.Fatalf("轮次时间不该是零值:\n%s", text)
	}
}

// 会话创建即落一份文件，用户不必等第一轮跑完才看见。
func TestAgentSessionFileCreatedOnCreate(t *testing.T) {
	db, err := database.Init(filepath.Join(t.TempDir(), "agentcreate.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root)

	m := NewManager(db)
	s, err := m.CreateSession("空会话", "global", nil, nil, PermConfirm)
	if err != nil {
		t.Fatal(err)
	}
	svc := bookstore.NewService(db)
	if err := svc.MirrorAgentSession(s.ID); err != nil {
		t.Fatal(err)
	}
	var rel string
	db.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, s.ID).Scan(&rel)
	if !strings.HasPrefix(rel, "_memory/agent/") {
		t.Fatalf("全局会话应落到库根 _memory/agent/: %q", rel)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatalf("会话文件未创建: %v", err)
	}
}

// 这两条是"漏接写入点"的回归：改的是库，文件却不动。
//
// TouchSession 更新的是 front-matter 里的 updated_at，DeleteRun 删掉的是轮次的来源事件——
// 两者不标脏时，文件会停在旧内容，而界面上看不出任何异常。
func TestTouchSessionMarksDirty(t *testing.T) {
	db, root := newAgentMemEnv(t, "touch.db")
	m := NewManager(db)
	s := mustCreateSession(t, m, "触点会话")
	svc := bookstore.NewService(db)
	bookstore.RegisterMemoryRenderer(bookstore.MemoryKindAgent, svc.MirrorAgentSession)

	if err := svc.MirrorAgentSession(s.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := bookstore.DefaultFlusher().FlushAll(); err != nil {
		t.Fatal(err)
	}
	if n := bookstore.PendingMemoryWrites(); n != 0 {
		t.Fatalf("落盘后不该还有待写项: %d", n)
	}

	m.TouchSession(s.ID)
	if n := bookstore.PendingMemoryWrites(); n == 0 {
		t.Fatal("TouchSession 改了 updated_at，必须标脏（否则文件里的更新时间永远停在首次落盘）")
	}
	if _, err := bookstore.DefaultFlusher().FlushAll(); err != nil {
		t.Fatal(err)
	}
	var rel string
	db.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, s.ID).Scan(&rel)
	// 全局会话落在库根的 _memory/agent/，不在某本书下面
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	_ = raw
}

func TestDeleteRunMarksDirty(t *testing.T) {
	db, root := newAgentMemEnv(t, "delrun.db")
	m := NewManager(db)
	s := mustCreateSession(t, m, "删轮次会话")
	svc := bookstore.NewService(db)
	bookstore.RegisterMemoryRenderer(bookstore.MemoryKindAgent, svc.MirrorAgentSession)

	// DeleteRun 只允许删已结束的 run，所以先按正常状态建一行
	if _, err := db.Exec(`INSERT INTO agent_runs (id, session_id, status, input)
		VALUES ('run-x', ?, ?, '第一轮的问题')`, s.ID, StatusCompleted); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AppendEvent(s.ID, "run-x", EventRunCreated, map[string]string{"input": "第一轮的问题"}); err != nil {
		t.Fatal(err)
	}
	if _, err := m.AppendEvent(s.ID, "run-x", EventRunCompleted, map[string]string{"status": StatusCompleted}); err != nil {
		t.Fatal(err)
	}
	if _, err := bookstore.DefaultFlusher().FlushAll(); err != nil {
		t.Fatal(err)
	}
	var rel string
	db.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, s.ID).Scan(&rel)
	abs := filepath.Join(root, filepath.FromSlash(rel))
	if raw, _ := os.ReadFile(abs); !strings.Contains(string(raw), "第一轮的问题") {
		t.Fatalf("前置条件不成立，文件里应有这一轮:\n%s", string(raw))
	}

	if err := m.DeleteRun("run-x"); err != nil {
		t.Fatal(err)
	}
	if n := bookstore.PendingMemoryWrites(); n == 0 {
		t.Fatal("删轮次删掉了事件，文件必须重渲染")
	}
	if _, err := bookstore.DefaultFlusher().FlushAll(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(abs)
	if strings.Contains(string(raw), "第一轮的问题") {
		t.Fatalf("已删除的轮次不该还留在文件里（否则按目录恢复会把它找回来）:\n%s", string(raw))
	}
}

// 工具直接写库的路径最容易漏：AI 建的角色卡/图谱节点也必须出现在目录里。
func TestToolWritesReachDisk(t *testing.T) {
	db, root := newAgentMemEnv(t, "tools.db")
	svc := bookstore.NewService(db)
	bookstore.RegisterMemoryRenderer(bookstore.MemoryKindGraph, svc.MirrorGraphByKey)

	var bookID int
	db.QueryRow(`SELECT id FROM books WHERE title='接线书'`).Scan(&bookID)

	m := NewManager(db)
	create := &Proposal{ToolName: "character_create", Arguments: `{"book_id":` + itoa(bookID) + `,"name":"林越","gender":"male","personality":"沉稳","info":"副车长"}`}
	if _, _, err := m.execProposalWrite(create); err != nil {
		t.Fatal(err)
	}
	var rel string
	db.QueryRow(`SELECT COALESCE(file_path,'') FROM characters WHERE book_id=?`, bookID).Scan(&rel)
	if rel == "" {
		t.Fatal("工具建的角色卡没有 file_path —— 目录里看不到它")
	}
	if _, err := os.Stat(filepath.Join(root, "接线书", filepath.FromSlash(rel))); err != nil {
		t.Fatalf("角色卡文件不存在: %v", err)
	}

	ent := &Proposal{ToolName: "graph_entity_create", Arguments: `{"book_id":` + itoa(bookID) + `,"name":"太荒宗","type":"组织","description":"主角出身"}`}
	if _, _, err := m.execProposalWrite(ent); err != nil {
		t.Fatal(err)
	}
	if n := bookstore.PendingMemoryWrites(); n == 0 {
		t.Fatal("工具建图谱实体后必须标脏，否则 AI 改了图谱而文件不动")
	}
	if _, err := bookstore.DefaultFlusher().FlushAll(); err != nil {
		t.Fatal(err)
	}
	matches, _ := filepath.Glob(filepath.Join(root, "接线书", "memory", "graph", "*.md"))
	if len(matches) == 0 {
		t.Fatal("未生成图谱文件")
	}
	raw, _ := os.ReadFile(matches[0])
	if !strings.Contains(string(raw), "太荒宗") {
		t.Fatalf("图谱文件里没有工具新建的实体:\n%s", string(raw))
	}
}

func newAgentMemEnv(t *testing.T, dbName string) (*sql.DB, string) {
	t.Helper()
	db, err := database.Init(filepath.Join(t.TempDir(), dbName))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	root := t.TempDir()
	if _, err := db.Exec(`INSERT INTO app_settings(key, value) VALUES('books_root', ?)`, root); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO books (title, status, folder_path, tags) VALUES ('接线书','draft','接线书','[]')`); err != nil {
		t.Fatal(err)
	}
	// 文件操作需要目录真实存在
	if err := os.MkdirAll(filepath.Join(root, "接线书"), 0o755); err != nil {
		t.Fatal(err)
	}
	return db, root
}

func mustCreateSession(t *testing.T, m *Manager, title string) *Session {
	t.Helper()
	s, err := m.CreateSession(title, "global", nil, nil, PermConfirm)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func itoa(v int) string {
	return strconv.Itoa(v)
}
