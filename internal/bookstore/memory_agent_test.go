package bookstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// M2 的核心验证：账本 → Markdown 的渲染是否正确、
// 密钥是否被遮蔽、解析回来是否一致、全局会话是否落到库根。

// seedAgentSession 造一个会话 + 账本事件。
func seedAgentSession(t *testing.T, svc *Service, scopeKind string, bookID int, events []struct {
	runID string
	typ   string
	body  map[string]any
}) string {
	t.Helper()
	id := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + string(rune('a'+len(events)%26))
	var scopeID any
	if scopeKind == "book" {
		scopeID = bookID
	}
	if _, err := svc.DB.Exec(`INSERT INTO agent_sessions (id, title, scope_kind, scope_id, permission)
		VALUES (?,?,?,?,?)`, id, "改写第三章", scopeKind, scopeID, "confirm"); err != nil {
		t.Fatal(err)
	}
	for i, ev := range events {
		payload, _ := json.Marshal(ev.body)
		if _, err := svc.DB.Exec(`INSERT INTO agent_events (session_id, run_id, seq, type, payload)
			VALUES (?,?,?,?,?)`, id, ev.runID, i+1, ev.typ, string(payload)); err != nil {
			t.Fatal(err)
		}
	}
	return id
}

func TestAgentSession_RenderFromLedger(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "Agent书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}

	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "把第三章打斗改紧凑些"}},
		{"r1", "tool.started", map[string]any{"call_id": "c1", "name": "chapter_list",
			"arguments": json.RawMessage(`{"book_id":22}`)}},
		{"r1", "tool.completed", map[string]any{"call_id": "c1", "name": "chapter_list",
			"result": json.RawMessage(`{"chapters":[1,2]}`)}},
		{"r1", "model.completed", map[string]any{"text": "我看了目录，建议先合并这两章。"}},
		{"r1", "run.completed", map[string]any{"status": "completed"}},
	})

	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	if !strings.HasPrefix(rel, "memory/agent/") {
		t.Fatalf("书内会话应落在 memory/agent/: %q", rel)
	}
	abs := filepath.Join(root, "Agent书", filepath.FromSlash(rel))
	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("会话文件未生成: %v", err)
	}
	text := string(raw)

	for _, want := range []string{
		"kind: agent_session",
		"<!-- xm:session -->",
		`<!-- xm:run id=r1`,
		"把第三章打斗改紧凑些",
		"我看了目录，建议先合并这两章。",
		`name=chapter_list`,
		"## 第 1 轮",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("文件里缺少 %q:\n%s", want, text)
		}
	}

	// 关键：逐 token 增量不落盘
	var deltaCount int
	svc.DB.QueryRow(`SELECT COUNT(*) FROM agent_events WHERE type IN ('model.delta','thinking.delta')`).Scan(&deltaCount)
	if strings.Contains(text, "model.delta") || strings.Contains(text, "thinking.delta") {
		t.Fatal("流式增量不应出现在文件里")
	}

	// 解析回来要能对上
	doc, err := ParseAgentDoc(text)
	if err != nil {
		t.Fatalf("解析自己生成的文件失败: %v", err)
	}
	if doc.SessionID != sessionID || doc.BookID != b.ID || doc.Scope != "book" {
		t.Fatalf("front-matter 解析不对: %+v", doc)
	}
	if len(doc.Runs) != 1 {
		t.Fatalf("应有 1 轮，实得 %d", len(doc.Runs))
	}
	run := doc.Runs[0]
	if run.Input != "把第三章打斗改紧凑些" || run.Status != "completed" {
		t.Fatalf("轮次解析不对: %+v", run)
	}
	if len(run.Tools) != 1 || run.Tools[0].Name != "chapter_list" || run.Tools[0].State != "completed" {
		t.Fatalf("工具调用解析不对: %+v", run.Tools)
	}
	if !strings.Contains(run.Tools[0].ArgsShown, "book_id") {
		t.Fatalf("工具参数未还原: %q", run.Tools[0].ArgsShown)
	}
}

// 工具参数里的密钥必须被遮蔽，但充值账本本身不能被改动。
func TestAgentSession_RedactsSecretsInFileOnly(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "脱敏书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}

	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	secret := "sk-abcdef1234567890"
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "调一下接口"}},
		{"r1", "tool.started", map[string]any{"call_id": "c1", "name": "some_tool",
			"arguments": json.RawMessage(`{"api_key":"` + secret + `","note":"Bearer ` + secret + `","book_id":7}`)}},
		{"r1", "run.completed", map[string]any{"status": "completed"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	raw, _ := os.ReadFile(filepath.Join(root, "脱敏书", filepath.FromSlash(rel)))
	text := string(raw)

	if strings.Contains(text, secret) {
		t.Fatalf("密钥不能出现在 md 里:\n%s", text)
	}
	if !strings.Contains(text, `"api_key":"***"`) {
		t.Fatalf("敏感字段应被遮蔽为 ***:\n%s", text)
	}
	if !strings.Contains(text, `"note":"***"`) {
		t.Fatalf("值里长得像令牌的字符串也要遮掉:\n%s", text)
	}
	// 非敏感字段要保留
	if !strings.Contains(text, `"book_id":7`) {
		t.Fatalf("非敏感字段不该被动:\n%s", text)
	}
	// 运行期账本保持原文：改了就查不出问题
	var payload string
	svc.DB.QueryRow(`SELECT payload FROM agent_events WHERE type='tool.started'`).Scan(&payload)
	if !strings.Contains(payload, secret) {
		t.Fatal("账本里的原始参数不应被改写")
	}
}

// 全局会话（没有绑定书）落到库根的 _memory/agent/。
func TestAgentSession_GlobalGoesToRoot(t *testing.T) {
	svc, root := setupService(t)

	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "global", 0, []ev{
		{"r1", "run.created", map[string]any{"input": "全局会话"}},
		{"r1", "run.completed", map[string]any{"status": "completed"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	if !strings.HasPrefix(rel, "_memory/agent/") {
		t.Fatalf("全局会话应落在 _memory/agent/，实得 %q", rel)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
		t.Fatalf("全局会话文件不存在: %v", err)
	}
}

// 用户改过的回答才是"最终回答"。
func TestAgentSession_EditedAnswerWins(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "改写书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "问一句"}},
		{"r1", "model.completed", map[string]any{"text": "原始回答"}},
		{"r1", "answer.edited", map[string]any{"text": "我改过的回答"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	raw, _ := os.ReadFile(filepath.Join(root, "改写书", filepath.FromSlash(rel)))
	if !strings.Contains(string(raw), "我改过的回答") {
		t.Fatalf("应使用编辑后的回答:\n%s", string(raw))
	}
	if strings.Contains(string(raw), "原始回答") {
		t.Fatalf("编辑后不应再出现原始回答:\n%s", string(raw))
	}
}

// 失败轮次要带上原因。
func TestAgentSession_FailedRunKeepsError(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "失败书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "跑一下"}},
		{"r1", "run.completed", map[string]any{"status": "failed", "error": "模型返回 401"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	raw, _ := os.ReadFile(filepath.Join(root, "失败书", filepath.FromSlash(rel)))
	text := string(raw)
	if !strings.Contains(text, "模型返回 401") || !strings.Contains(text, "失败") {
		t.Fatalf("失败轮次应写明原因:\n%s", text)
	}
}

// 删除会话：文件进回收站，不留孤儿。
func TestAgentSession_Trash(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "删除书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "x"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	abs := filepath.Join(root, "删除书", filepath.FromSlash(rel))

	if err := svc.TrashAgentFile(sessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(abs); !os.IsNotExist(err) {
		t.Fatal("回收后原位置不应还有文件")
	}
}

// 用户自定义的 front-matter 键在重渲染时保留。
func TestAgentSession_KeepsCustomFrontMatter(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "自定义书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "x"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	abs := filepath.Join(root, "自定义书", filepath.FromSlash(rel))

	raw, _ := os.ReadFile(abs)
	lines := strings.SplitN(string(raw), "\n", 2)
	custom := lines[0] + "\n我的备注: 手写的\n" + lines[1]
	if err := os.WriteFile(abs, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(abs)
	if !strings.Contains(string(after), "我的备注: 手写的") {
		t.Fatalf("重渲染时应保留自定义字段:\n%s", string(after))
	}
}

// 多个提案时标题只应出现一次，且文件不应以成串空行结尾。
func TestAgentSession_ProposalHeadingOnce(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "提案书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	type ev = struct {
		runID string
		typ   string
		body  map[string]any
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []ev{
		{"r1", "run.created", map[string]any{"input": "改两处"}},
		{"r1", "proposal.created", map[string]any{"proposal_id": "p1", "tool": "chapter_append",
			"target": "第1章", "risk": "low", "summary": "追加", "status": "pending"}},
		{"r1", "proposal.created", map[string]any{"proposal_id": "p2", "tool": "chapter_replace",
			"target": "第2章", "risk": "medium", "summary": "替换", "status": "pending"}},
		{"r1", "run.completed", map[string]any{"status": "completed"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT file_path FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	raw, _ := os.ReadFile(filepath.Join(root, "提案书", filepath.FromSlash(rel)))
	text := string(raw)

	if n := strings.Count(text, "### 写入提案"); n != 1 {
		t.Fatalf("提案标题应只出现一次，实得 %d:\n%s", n, text)
	}
	if n := strings.Count(text, "<!-- xm:proposal "); n != 2 {
		t.Fatalf("应有两个 proposal 标记，实得 %d", n)
	}
	if strings.HasSuffix(text, "\n\n\n") {
		t.Fatalf("文件不应以成串空行结尾")
	}
	// 解析回来两个提案都在
	doc, err := ParseAgentDoc(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Runs) != 1 || len(doc.Runs[0].Proposals) != 2 {
		t.Fatalf("提案解析不对: %+v", doc.Runs)
	}
	if doc.Runs[0].Proposals[1].Tool != "chapter_replace" {
		t.Fatalf("第二个提案内容不对: %+v", doc.Runs[0].Proposals[1])
	}
}

// 文件是真相：程序不覆盖被外部改过的会话文件，只如实记下原因。
//
// 不拦的话，用户在编辑器里改完，下一次落盘就静默盖掉 —— 他会以为改动生效了。
func TestAgentSession_ExternalEditNotOverwritten(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "外部改书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []struct {
		runID string
		typ   string
		body  map[string]any
	}{
		{"r1", "run.created", map[string]any{"input": "原始问题"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var rel string
	svc.DB.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	abs := filepath.Join(root, "外部改书", filepath.FromSlash(rel))

	raw, err := os.ReadFile(abs)
	if err != nil {
		t.Fatal(err)
	}
	edited := strings.Replace(string(raw), "原始问题", "我在外部改成了这句", 1)
	if err := os.WriteFile(abs, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(abs)
	if !strings.Contains(string(after), "我在外部改成了这句") {
		t.Fatalf("外部改动被覆盖了:\n%s", string(after))
	}
	var pe string
	svc.DB.QueryRow(`SELECT COALESCE(parse_error,'') FROM agent_sessions WHERE id=?`, sessionID).Scan(&pe)
	if !strings.Contains(pe, "未覆盖") {
		t.Fatalf("跳过写入必须留下原因，实得 %q", pe)
	}
}

// 改标题后文件名要跟着改（与章节、图谱一致），短码与时间前缀不动。
func TestAgentSession_RenameFollowsFileName(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "改名书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []struct {
		runID string
		typ   string
		body  map[string]any
	}{
		{"r1", "run.created", map[string]any{"input": "问题"}},
	})
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var before string
	svc.DB.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, sessionID).Scan(&before)

	svc.DB.Exec(`UPDATE agent_sessions SET title='改写第五章' WHERE id=?`, sessionID)
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	var after string
	svc.DB.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, sessionID).Scan(&after)
	if !strings.Contains(after, "改写第五章") {
		t.Fatalf("文件名应跟随新标题: %q → %q", before, after)
	}
	if _, err := os.Stat(filepath.Join(root, "改名书", filepath.FromSlash(after))); err != nil {
		t.Fatalf("改名后的文件不存在: %v", err)
	}
}

// 引擎改名叫"被外部改过"，程序不该再写它，直到用户处理。
func TestAgentSession_ExternalEditBlocksFurtherWrites(t *testing.T) {
	svc, root := setupService(t)
	b := &models.Book{Title: "拦截书", Status: "draft"}
	if err := svc.CreateBook(b); err != nil {
		t.Fatal(err)
	}
	sessionID := seedAgentSession(t, svc, "book", b.ID, []struct {
		runID string
		typ   string
		body  map[string]any
	}{
		{"r1", "run.created", map[string]any{"input": "A"}},
	})
	svc.MirrorAgentSession(sessionID)
	var rel string
	svc.DB.QueryRow(`SELECT COALESCE(file_path,'') FROM agent_sessions WHERE id=?`, sessionID).Scan(&rel)
	abs := filepath.Join(root, "拦截书", filepath.FromSlash(rel))
	os.WriteFile(abs, []byte("用户把整个文件写成了别的东西"), 0o644)

	// 又来了新事件，但文件仍不该被覆盖
	svc.DB.Exec(`INSERT INTO agent_events (session_id, run_id, seq, type, payload)
		VALUES (?, 'r1', 99, 'model.completed', '{"text":"新回答"}')`, sessionID)
	if err := svc.MirrorAgentSession(sessionID); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(abs)
	if !strings.Contains(string(raw), "用户把整个文件写成了别的东西") {
		t.Fatalf("文件被覆盖了:\n%s", string(raw))
	}
}
