package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 工作流画布后端：只做持久化 + 文件落盘，不解析节点语义（执行完全在前端）。
//
// 落盘根目录 = 进程工作目录下的 workflow_outputs/（与源项目 Electron 版同名的目录约定）。
// 每次执行一个子目录 <名称>_<日期>_<时间戳>_<runId>/，内部 nodes/ chapters/ logs/ 三类。

const workflowOutputRootName = "workflow_outputs"

// workflowOutputRoot 返回落盘根目录（绝对路径）。
func workflowOutputRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root := filepath.Join(cwd, workflowOutputRootName)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return abs, nil
}

// sanitizeSegment 去掉文件名/目录名里的非法字符。
func sanitizeSegment(s string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "*", "_", "?", "_", "\"", "_", "<", "_", ">", "_", "|", "_")
	s = replacer.Replace(strings.TrimSpace(s))
	s = strings.Trim(s, ". ")
	if s == "" {
		s = "workflow"
	}
	if len(s) > 80 {
		s = s[:80]
	}
	return s
}

// safeJoin 把客户端给的相对路径拼到 base 下，并保证结果不逃出 base。
// 拒绝绝对路径、盘符、以及任何 ".." 段落。
func safeJoin(base, rel string) (string, error) {
	rel = strings.TrimSpace(strings.ReplaceAll(rel, "\\", "/"))
	rel = strings.TrimLeft(rel, "/")
	if rel == "" {
		return "", fmt.Errorf("路径不能为空")
	}
	if strings.Contains(rel, "\x00") {
		return "", fmt.Errorf("路径非法")
	}
	if filepath.IsAbs(rel) || (len(rel) >= 2 && rel[1] == ':') {
		return "", fmt.Errorf("不接受绝对路径")
	}
	for _, seg := range strings.Split(rel, "/") {
		if seg == ".." {
			return "", fmt.Errorf("路径非法（含 ..）")
		}
	}
	full := filepath.Join(base, filepath.FromSlash(rel))
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if absFull != absBase && !strings.HasPrefix(absFull, absBase+string(os.PathSeparator)) {
		return "", fmt.Errorf("路径越界")
	}
	return absFull, nil
}

// runDir 返回某次执行的落盘目录（workflow_runs.output_dir 存绝对路径）。
func (h *Handler) runDir(runID int) (string, error) {
	var dir string
	err := h.db.QueryRow("SELECT output_dir FROM workflow_runs WHERE id = ?", runID).Scan(&dir)
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", fmt.Errorf("该执行没有落盘目录")
	}
	return dir, nil
}

// ---------- workflows ----------

// GetWorkflows lists all workflows (updated_at DESC) with run counts.
func (h *Handler) GetWorkflows(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT w.id, w.name, w.description, w.category, w.graph, w.config, w.created_at, w.updated_at,
		       (SELECT COUNT(*) FROM workflow_runs r WHERE r.workflow_id = w.id) AS run_count
		FROM workflows w
		ORDER BY w.category ASC, w.updated_at DESC, w.id DESC`)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := []models.Workflow{}
	for rows.Next() {
		var wf models.Workflow
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.Description, &wf.Category, &wf.Graph, &wf.Config, &wf.CreatedAt, &wf.UpdatedAt, &wf.RunCount); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, wf)
	}
	respondOK(w, list)
}

// GetWorkflow returns one workflow by id.
func (h *Handler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	wf, err := h.queryWorkflow(id)
	if err == sql.ErrNoRows {
		respondFail(w, 404, "工作流不存在")
		return
	}
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, wf)
}

func (h *Handler) queryWorkflow(id int) (models.Workflow, error) {
	var wf models.Workflow
	err := h.db.QueryRow(`
		SELECT id, name, description, category, graph, config, created_at, updated_at
		FROM workflows WHERE id = ?`, id).
		Scan(&wf.ID, &wf.Name, &wf.Description, &wf.Category, &wf.Graph, &wf.Config, &wf.CreatedAt, &wf.UpdatedAt)
	return wf, err
}

// CreateWorkflow creates a workflow. graph/config 缺省时落空图。
func (h *Handler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var wf models.Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	wf.Name = strings.TrimSpace(wf.Name)
	if wf.Name == "" {
		wf.Name = "未命名工作流"
	}
	if strings.TrimSpace(wf.Graph) == "" {
		wf.Graph = `{"nodes":[],"edges":[]}`
	}
	if strings.TrimSpace(wf.Config) == "" {
		wf.Config = "{}"
	}

	res, err := h.db.Exec(
		"INSERT INTO workflows (name, description, category, graph, config) VALUES (?, ?, ?, ?, ?)",
		wf.Name, wf.Description, wf.Category, wf.Graph, wf.Config)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := res.LastInsertId()
	out, err := h.queryWorkflow(int(id))
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, out)
}

// UpdateWorkflow 按「请求体出现过的键」拼 SET，避免整行覆盖把未传字段清零。
func (h *Handler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	body, err := readAllBody(r)
	if err != nil {
		respondFail(w, 400, "读取请求体失败")
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	sets := []string{}
	args := []interface{}{}
	for _, key := range []string{"name", "description", "category", "graph", "config"} {
		v, ok := raw[key]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			// graph/config 也允许直接传对象，统一序列化成字符串
			var anyVal interface{}
			if err2 := json.Unmarshal(v, &anyVal); err2 != nil {
				continue
			}
			b, mErr := json.Marshal(anyVal)
			if mErr != nil {
				continue
			}
			s = string(b)
		}
		if key == "name" && strings.TrimSpace(s) == "" {
			s = "未命名工作流"
		}
		sets = append(sets, key+" = ?")
		args = append(args, s)
	}
	if len(sets) == 0 {
		respondFail(w, 400, "没有可更新的字段")
		return
	}
	sets = append(sets, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	if _, err := h.db.Exec("UPDATE workflows SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	out, err := h.queryWorkflow(id)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, out)
}

// DeleteWorkflow 删工作流（级联删执行记录）。
func (h *Handler) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	if _, err := h.db.Exec("DELETE FROM workflows WHERE id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.db.Exec("DELETE FROM workflow_outputs WHERE workflow_id = ?", id)
	respondOK(w, "删除成功")
}

// ---------- workflow runs ----------

// CreateWorkflowRun 开始一次执行：建行 + 建落盘目录，返回带 output_dir 的 run。
func (h *Handler) CreateWorkflowRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WorkflowID int    `json:"workflow_id"`
		OutputDir  string `json:"output_dir"` // 仅作展示（前端可能配了自定义目录）；实际落盘固定在根目录内
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if req.WorkflowID <= 0 {
		respondFail(w, 400, "缺少 workflow_id")
		return
	}

	var name string
	if err := h.db.QueryRow("SELECT name FROM workflows WHERE id = ?", req.WorkflowID).Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			respondFail(w, 404, "工作流不存在")
			return
		}
		respondFail(w, 500, err.Error())
		return
	}

	res, err := h.db.Exec("INSERT INTO workflow_runs (workflow_id, status) VALUES (?, 'running')", req.WorkflowID)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	runID64, _ := res.LastInsertId()
	runID := int(runID64)

	dir := req.OutputDir
	if root, rootErr := workflowOutputRoot(); rootErr == nil {
		sub := fmt.Sprintf("%s_%s_%d", sanitizeSegment(name), time.Now().Format("2006-01-02_150405"), runID)
		candidate := filepath.Join(root, sub)
		if mkErr := os.MkdirAll(filepath.Join(candidate, "nodes"), 0o755); mkErr == nil {
			os.MkdirAll(filepath.Join(candidate, "chapters"), 0o755)
			os.MkdirAll(filepath.Join(candidate, "logs"), 0o755)
			dir = candidate
		}
	}
	if dir != "" {
		h.db.Exec("UPDATE workflow_runs SET output_dir = ? WHERE id = ?", dir, runID)
	}

	var run models.WorkflowRun
	if err := h.db.QueryRow(`
		SELECT id, workflow_id, status, logs, context, COALESCE(error, ''), output_dir, started_at, COALESCE(finished_at, '')
		FROM workflow_runs WHERE id = ?`, runID).
		Scan(&run.ID, &run.WorkflowID, &run.Status, &run.Logs, &run.Context, &run.Error, &run.OutputDir, &run.StartedAt, &run.FinishedAt); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, run)
}

// UpdateWorkflowRun 结束时回写 status / logs / context / error（按出现过的键）。
func (h *Handler) UpdateWorkflowRun(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	body, err := readAllBody(r)
	if err != nil {
		respondFail(w, 400, "读取请求体失败")
		return
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	sets := []string{}
	args := []interface{}{}
	for _, key := range []string{"status", "logs", "context", "error", "output_dir"} {
		v, ok := raw[key]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			var anyVal interface{}
			if err2 := json.Unmarshal(v, &anyVal); err2 != nil {
				continue
			}
			b, mErr := json.Marshal(anyVal)
			if mErr != nil {
				continue
			}
			s = string(b)
		}
		sets = append(sets, key+" = ?")
		args = append(args, s)
	}
	if len(sets) == 0 {
		respondFail(w, 400, "没有可更新的字段")
		return
	}
	// 终态时记 finished_at
	if st, ok := raw["status"]; ok {
		var s string
		if json.Unmarshal(st, &s) == nil && s != "" && s != "running" {
			sets = append(sets, "finished_at = CURRENT_TIMESTAMP")
		}
	}
	args = append(args, id)

	if _, err := h.db.Exec("UPDATE workflow_runs SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "更新成功")
}

// ListWorkflowRuns ?workflow_id=&limit=
func (h *Handler) ListWorkflowRuns(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, workflow_id, status, logs, context, COALESCE(error, ''), output_dir, started_at, COALESCE(finished_at, '') FROM workflow_runs`
	args := []interface{}{}
	if wid := r.URL.Query().Get("workflow_id"); wid != "" {
		query += " WHERE workflow_id = ?"
		args = append(args, wid)
	}
	query += " ORDER BY started_at DESC, id DESC LIMIT ?"
	args = append(args, queryInt(r, "limit", 50))

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := []models.WorkflowRun{}
	for rows.Next() {
		var run models.WorkflowRun
		if err := rows.Scan(&run.ID, &run.WorkflowID, &run.Status, &run.Logs, &run.Context, &run.Error, &run.OutputDir, &run.StartedAt, &run.FinishedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, run)
	}
	respondOK(w, list)
}

// DeleteWorkflowRun 删除执行记录（含落盘目录）。
func (h *Handler) DeleteWorkflowRun(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	if dir, err := h.runDir(id); err == nil && dir != "" {
		if root, rootErr := workflowOutputRoot(); rootErr == nil && strings.HasPrefix(dir, root+string(os.PathSeparator)) {
			os.RemoveAll(dir)
		}
	}
	h.db.Exec("DELETE FROM workflow_outputs WHERE run_id = ?", id)
	if _, err := h.db.Exec("DELETE FROM workflow_runs WHERE id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}

// ---------- workflow outputs（文件落盘）----------

// WorkflowOutputInfo 详情接口 / 写盘接口的请求体。
type workflowWriteRequest struct {
	RunID      int    `json:"run_id"`
	WorkflowID int    `json:"workflow_id"`
	Kind       string `json:"kind"`     // node / chapter / log / final / config
	RelPath    string `json:"rel_path"` // 相对该次执行目录，如 nodes/xx.txt
	Content    string `json:"content"`
	Append     bool   `json:"append"`
	// Encoding: "" 按文本写；"base64" 先解码再按二进制写（图片/视频输出节点用）
	Encoding string `json:"encoding"`
}

// WriteWorkflowOutput 把内容写进某次执行的目录。
func (h *Handler) WriteWorkflowOutput(w http.ResponseWriter, r *http.Request) {
	var req workflowWriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	base, err := h.runDir(req.RunID)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	full, err := safeJoin(base, req.RelPath)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		respondFail(w, 500, err.Error())
		return
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if req.Append {
		flags = os.O_CREATE | os.O_WRONLY | os.O_APPEND
	}
	f, err := os.OpenFile(full, flags, 0o644)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	if strings.EqualFold(req.Encoding, "base64") {
		// 图片/视频等二进制：base64 解码后写字节
		raw := strings.TrimSpace(req.Content)
		if idx := strings.Index(raw, ","); strings.HasPrefix(raw, "data:") && idx > 0 {
			raw = raw[idx+1:]
		}
		decoded, derr := base64.StdEncoding.DecodeString(raw)
		if derr != nil {
			f.Close()
			respondFail(w, 400, "base64 解码失败: "+derr.Error())
			return
		}
		if _, werr := f.Write(decoded); werr != nil {
			f.Close()
			respondFail(w, 500, werr.Error())
			return
		}
	} else if _, err := f.WriteString(req.Content); err != nil {
		f.Close()
		respondFail(w, 500, err.Error())
		return
	}
	f.Close()

	fi, _ := os.Stat(full)
	rel := filepath.ToSlash(req.RelPath)
	kind := req.Kind
	if kind == "" {
		kind = "node"
	}
	// 同一路径重复写就更新索引，不堆行
	var existID int
	if err := h.db.QueryRow("SELECT id FROM workflow_outputs WHERE run_id = ? AND rel_path = ?", req.RunID, rel).Scan(&existID); err == nil && existID > 0 {
		h.db.Exec("UPDATE workflow_outputs SET size = ?, kind = ? WHERE id = ?", fi.Size(), kind, existID)
	} else {
		h.db.Exec("INSERT INTO workflow_outputs (run_id, workflow_id, kind, rel_path, size) VALUES (?, ?, ?, ?, ?)",
			req.RunID, req.WorkflowID, kind, rel, fi.Size())
	}

	respondOK(w, map[string]interface{}{
		"rel_path":   rel,
		"abs_path":   full,
		"size":       fi.Size(),
		"output_dir": base,
	})
}

// MkdirWorkflowOutput 建目录（executor 初始化工作区用）。
func (h *Handler) MkdirWorkflowOutput(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RunID   int    `json:"run_id"`
		RelPath string `json:"rel_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	base, err := h.runDir(req.RunID)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	full, err := safeJoin(base, req.RelPath)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	if err := os.MkdirAll(full, 0o755); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, map[string]string{"abs_path": full})
}

// ListWorkflowOutputs ?run_id=
func (h *Handler) ListWorkflowOutputs(w http.ResponseWriter, r *http.Request) {
	runIDStr := r.URL.Query().Get("run_id")
	query := "SELECT id, run_id, workflow_id, kind, rel_path, size, created_at FROM workflow_outputs"
	args := []interface{}{}
	if runIDStr != "" {
		query += " WHERE run_id = ?"
		args = append(args, runIDStr)
	}
	query += " ORDER BY created_at ASC, id ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := []models.WorkflowOutput{}
	for rows.Next() {
		var o models.WorkflowOutput
		if err := rows.Scan(&o.ID, &o.RunID, &o.WorkflowID, &o.Kind, &o.RelPath, &o.Size, &o.CreatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, o)
	}
	respondOK(w, list)
}

// ReadWorkflowOutput ?run_id=&rel_path= —— 查看落盘内容
func (h *Handler) ReadWorkflowOutput(w http.ResponseWriter, r *http.Request) {
	runID, err := strconv.Atoi(r.URL.Query().Get("run_id"))
	if err != nil || runID <= 0 {
		respondFail(w, 400, "缺少 run_id")
		return
	}
	base, err := h.runDir(runID)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	full, err := safeJoin(base, r.URL.Query().Get("rel_path"))
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	data, err := os.ReadFile(full)
	if err != nil {
		respondFail(w, 404, "文件不存在或不可读")
		return
	}
	respondOK(w, map[string]string{
		"rel_path": filepath.ToSlash(r.URL.Query().Get("rel_path")),
		"content":  string(data),
	})
}

// RunOutputDir 返回某次执行的绝对目录（前端「打开输出目录」用）。
func (h *Handler) RunOutputDir(w http.ResponseWriter, r *http.Request) {
	runID, err := strconv.Atoi(r.URL.Query().Get("run_id"))
	if err != nil || runID <= 0 {
		respondFail(w, 400, "缺少 run_id")
		return
	}
	base, err := h.runDir(runID)
	if err != nil {
		respondFail(w, 400, err.Error())
		return
	}
	root, _ := workflowOutputRoot()
	respondOK(w, map[string]string{"output_dir": base, "root": root})
}
