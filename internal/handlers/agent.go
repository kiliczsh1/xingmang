package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/agent"
	"github.com/xingmang/xingmang-desktop/internal/bookstore"
)

// ===== 星芒 Agent 端点 =====
// 路由注册见 server.go；事件流格式：event 不单独发，data 里带 seq/type/payload。
// 前端用 fetch + ReadableStream 解析（不能用 EventSource：不支持断线传 after_seq 之外还要保活）。

// AgentModels 返回启用的模型列表（前端会话选择器用），并标记是否兼容 Agent。
func (h *Handler) AgentModels(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT m.id, m.name, m.model, m.is_default, COALESCE(p.name,''), COALESCE(p.api_format,'openai')
		FROM api_models m LEFT JOIN api_providers p ON m.provider_id = p.id
		WHERE m.enabled = 1 ORDER BY m.is_default DESC, m.sort_order ASC, m.id ASC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()
	type agentModel struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Model       string `json:"model"`
		IsDefault   int    `json:"is_default"`
		Provider    string `json:"provider_name"`
		Format      string `json:"api_format"`
		AgentUsable bool   `json:"agent_usable"`
	}
	list := []agentModel{}
	for rows.Next() {
		var m agentModel
		if err := rows.Scan(&m.ID, &m.Name, &m.Model, &m.IsDefault, &m.Provider, &m.Format); err != nil {
			continue
		}
		switch m.Format {
		case "", "openai", "anthropic_compat":
			m.AgentUsable = true
		}
		list = append(list, m)
	}
	respondOK(w, list)
}

// AgentSessionsList GET /api/agent/sessions
func (h *Handler) AgentSessionsList(w http.ResponseWriter, r *http.Request) {
	list, err := h.agent.ListSessions()
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if list == nil {
		list = []agent.Session{}
	}
	respondOK(w, list)
}

// AgentSessionCreate POST /api/agent/sessions {title, scope_kind, scope_id, model_id}
func (h *Handler) AgentSessionCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title      string `json:"title"`
		ScopeKind  string `json:"scope_kind"`
		ScopeID    *int64 `json:"scope_id"`
		ModelID    *int64 `json:"model_id"`
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	s, err := h.agent.CreateSession(req.Title, req.ScopeKind, req.ScopeID, req.ModelID, req.Permission)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	// 立刻落一份会话文件：让用户在资源管理器里马上能看到这个会话，
	// 而不是等第一轮跑完才出现。建不出来不影响会话可用（账本才是运行期依据）。
	if err := bookstore.NewService(h.db).MirrorAgentSession(s.ID); err != nil {
		log.Printf("[agent] 会话文件初始化失败 session=%s: %v", s.ID, err)
	}
	respondOK(w, s)
}

// AgentSessionGet GET /api/agent/sessions/{id}
func (h *Handler) AgentSessionGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	s, err := h.agent.GetSession(id)
	if err != nil {
		respondFail(w, 404, err.Error())
		return
	}
	respondOK(w, s)
}

// AgentSessionDelete DELETE /api/agent/sessions/{id}
func (h *Handler) AgentSessionDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// 先把会话文件移入回收站，再删行：反过来的话文件就成了孤儿，用户再也找不到。
	if err := bookstore.NewService(h.db).TrashAgentFile(id); err != nil {
		respondFail(w, 500, "移入回收站失败: "+err.Error())
		return
	}
	if err := h.agent.DeleteSession(id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]bool{"ok": true})
}

// AgentRunStart POST /api/agent/sessions/{id}/runs {input, model_id?}
func (h *Handler) AgentRunStart(w http.ResponseWriter, r *http.Request) {
	sid := r.PathValue("id")
	var req struct {
		Input   string `json:"input"`
		ModelID *int64 `json:"model_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if req.Input == "" {
		respondFail(w, 400, "input 不能为空")
		return
	}
	s, err := h.agent.GetSession(sid)
	if err != nil {
		respondFail(w, 404, err.Error())
		return
	}
	runID, err := h.agent.StartRun(s, req.Input, req.ModelID)
	if err != nil {
		respondFail(w, 409, err.Error())
		return
	}
	respondOK(w, map[string]string{"run_id": runID})
}

// AgentSessionPatch PATCH /api/agent/sessions/{id} {permission}
func (h *Handler) AgentSessionPatch(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Permission string `json:"permission"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if err := h.agent.UpdateSessionPermission(id, req.Permission); err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	s, err := h.agent.GetSession(id)
	if err != nil {
		respondFail(w, 404, err.Error())
		return
	}
	respondOK(w, s)
}

// AgentRunCancel POST /api/agent/runs/{id}/cancel
func (h *Handler) AgentRunCancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !h.agent.Cancel(id) {
		respondFail(w, 404, "运行不存在或已结束")
		return
	}
	respondOK(w, map[string]bool{"ok": true})
}

// AgentRunDelete DELETE /api/agent/runs/{id}
// 删除一整条对话（用户消息 + 回复 + 工具调用），账本 seq 由后端重排保持连续。
func (h *Handler) AgentRunDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.agent.DeleteRun(id); err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, map[string]bool{"ok": true})
}

// AgentRunEditAnswer PUT /api/agent/runs/{id}/answer {text}
// 修改该段对话的助手回答；后续对话按修改后的内容作为上下文。
func (h *Handler) AgentRunEditAnswer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Text) == "" {
		respondFail(w, 400, "回答内容不能为空")
		return
	}
	if err := h.agent.EditRunAnswer(id, req.Text); err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, map[string]bool{"ok": true})
}

// AgentProposalApprove POST /api/agent/proposals/{id}/approve —— 批准并写入。
func (h *Handler) AgentProposalApprove(w http.ResponseWriter, r *http.Request) {
	h.decideProposal(w, r, true)
}

// AgentProposalReject POST /api/agent/proposals/{id}/reject
func (h *Handler) AgentProposalReject(w http.ResponseWriter, r *http.Request) {
	h.decideProposal(w, r, false)
}

func (h *Handler) decideProposal(w http.ResponseWriter, r *http.Request, approve bool) {
	id := r.PathValue("id")
	p, err := h.agent.GetProposal(id)
	if err != nil {
		respondFail(w, 404, err.Error())
		return
	}
	if approve {
		// 试运行模式禁止写入
		if s, serr := h.agent.GetSession(p.SessionID); serr == nil && s.Permission == agent.PermObserve {
			respondFail(w, 400, "当前会话是试运行模式，不会写入任何数据；请先切换写入权限")
			return
		}
		result, err := h.agent.ApplyProposal(p)
		if err != nil {
			respondFail(w, 400, "写入失败: "+err.Error())
			return
		}
		respondOK(w, map[string]any{"status": agent.ProposalApplied, "result": result})
		return
	}
	if err := h.agent.RejectProposal(id); err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, map[string]any{"status": agent.ProposalRejected})
}

// AgentProposalRollback POST /api/agent/proposals/{id}/rollback —— 用原文快照还原章节。
func (h *Handler) AgentProposalRollback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.agent.RollbackProposal(id); err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	respondOK(w, map[string]bool{"ok": true})
}

// AgentStream GET /api/agent/sessions/{id}/stream?after_seq=0
// 服务端轮询账本推 SSE；客户端断开后按 after_seq 续传，事件不丢不重。
func (h *Handler) AgentStream(w http.ResponseWriter, r *http.Request) {
	sid := r.PathValue("id")
	after := int64(queryInt(r, "after_seq", 0))

	if _, err := h.agent.GetSession(sid); err != nil {
		respondFail(w, 404, err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		respondFail(w, 500, "当前连接不支持流式响应")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	fmt.Fprintf(w, "retry: 3000\n\n")
	flusher.Flush()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case <-ticker.C:
			events, err := h.agent.EventsAfter(sid, after)
			if err != nil {
				continue
			}
			for _, ev := range events {
				data, _ := json.Marshal(ev)
				fmt.Fprintf(w, "data: %s\n\n", data)
				after = ev.Seq
			}
			if len(events) > 0 {
				flusher.Flush()
			}
		}
	}
}

// AgentBookChoices GET /api/agent/books —— 新建会话时可绑定的书籍选择列表。
func (h *Handler) AgentBookChoices(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT id, title FROM books ORDER BY updated_at DESC LIMIT 200")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()
	type choice struct {
		ID    int    `json:"id"`
		Title string `json:"title"`
	}
	list := []choice{}
	for rows.Next() {
		var c choice
		if rows.Scan(&c.ID, &c.Title) == nil {
			list = append(list, c)
		}
	}
	respondOK(w, list)
}
