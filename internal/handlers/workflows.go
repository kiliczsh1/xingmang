package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xingmang/xingmang-desktop/internal/models"
)

// 工作流工作台的两组接口，契约见 web/src/utils/workflow.ts：
//   /api/workflows/*              条目 CRUD，scene 是整张画布的 JSON 快照
//   /api/workflows/node-library/* 节点模板库，对外键是 template_id
// 注意前端读的是驼峰 createdAt / updatedAt，且 scene 传的是对象（模型里用 JSONText 接）。

const workflowSelect = "SELECT id, name, description, scene, created_at, updated_at FROM workflows"

// respondWorkflow 读回整行再响应，保证 createdAt/updatedAt/scene 都是库里的真实值
// （INSERT/UPDATE 之后直接回显入参的话，前端拿到的 createdAt 会是空串）。
func (h *Handler) respondWorkflow(w http.ResponseWriter, id int) {
	var wf models.Workflow
	err := h.db.QueryRow(workflowSelect+" WHERE id = ?", id).
		Scan(&wf.ID, &wf.Name, &wf.Description, &wf.Scene, &wf.CreatedAt, &wf.UpdatedAt)
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

// GetWorkflows returns all workflows.
func (h *Handler) GetWorkflows(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(workflowSelect + " ORDER BY created_at DESC, id DESC")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := make([]models.Workflow, 0)
	for rows.Next() {
		var wf models.Workflow
		if err := rows.Scan(&wf.ID, &wf.Name, &wf.Description, &wf.Scene, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, wf)
	}
	if err := rows.Err(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, list)
}

// GetWorkflow returns a single workflow by id.
func (h *Handler) GetWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}
	h.respondWorkflow(w, id)
}

// CreateWorkflow creates a new workflow.
func (h *Handler) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	var wf models.Workflow
	if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	if strings.TrimSpace(wf.Name) == "" {
		respondFail(w, 400, "工作流名称不能为空")
		return
	}

	result, err := h.db.Exec("INSERT INTO workflows (name, description, scene) VALUES (?, ?, ?)",
		wf.Name, wf.Description, wf.Scene)
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	id, _ := result.LastInsertId()
	h.respondWorkflow(w, int(id))
}

// UpdateWorkflow updates an existing workflow.
// 前端有两种调用，且都是「局部更新」：
//
//	updateWorkflowItem(id, {name?, description?})   只改标题/描述
//	saveWorkflowScene(id, {scene: {...}})           只改画布
//
// 所以必须按「请求体里出现了哪些键」来拼 UPDATE，不能整行覆盖——
// 否则保存一次画布就会把 name / description 清成空串。
func (h *Handler) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r, "id")
	if !ok {
		respondFail(w, 400, "无效的ID")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}

	sets := []string{"updated_at = CURRENT_TIMESTAMP"}
	args := make([]interface{}, 0, 4)

	if v, ok := raw["name"]; ok {
		var name string
		if err := json.Unmarshal(v, &name); err != nil {
			respondFail(w, 400, "name 字段格式错误")
			return
		}
		if strings.TrimSpace(name) == "" {
			respondFail(w, 400, "工作流名称不能为空")
			return
		}
		sets = append(sets, "name = ?")
		args = append(args, name)
	}
	if v, ok := raw["description"]; ok {
		var description string
		if err := json.Unmarshal(v, &description); err != nil {
			respondFail(w, 400, "description 字段格式错误")
			return
		}
		sets = append(sets, "description = ?")
		args = append(args, description)
	}
	if v, ok := raw["scene"]; ok {
		var scene models.JSONText
		if err := json.Unmarshal(v, &scene); err != nil {
			respondFail(w, 400, "scene 字段格式错误")
			return
		}
		sets = append(sets, "scene = ?")
		args = append(args, scene)
	}

	args = append(args, id)
	if _, err := h.db.Exec("UPDATE workflows SET "+strings.Join(sets, ", ")+" WHERE id = ?", args...); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	h.respondWorkflow(w, id)
}

// DeleteWorkflow deletes a workflow by id.
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
	respondOK(w, "删除成功")
}

// ---------------------------------------------------------------- 节点模板库

const nodeLibrarySelect = "SELECT template_id, type, COALESCE(variant, ''), category, label, " +
	"COALESCE(description, ''), COALESCE(outputs, ''), generated, template FROM workflow_node_library"

// GetWorkflowNodeLibrary 返回全部节点模板（GET /api/workflows/node-library/all）
func (h *Handler) GetWorkflowNodeLibrary(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(nodeLibrarySelect + " ORDER BY category, label")
	if err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	defer rows.Close()

	list := make([]models.WorkflowNodeLibraryItem, 0)
	for rows.Next() {
		var item models.WorkflowNodeLibraryItem
		if err := rows.Scan(&item.ID, &item.Type, &item.Variant, &item.Category, &item.Label,
			&item.Description, &item.Outputs, &item.Generated, &item.Template); err != nil {
			respondFail(w, 500, err.Error())
			return
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, list)
}

// SaveWorkflowNodeLibraryItem 新增或覆盖一个节点模板（POST /api/workflows/node-library）
// 前端 saveWorkflowNodeLibrary 是逐条 POST，所以这里必须按 template_id 做 upsert。
func (h *Handler) SaveWorkflowNodeLibraryItem(w http.ResponseWriter, r *http.Request) {
	var item models.WorkflowNodeLibraryItem
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		respondFail(w, 400, "请求体格式错误: "+err.Error())
		return
	}
	item.ID = strings.TrimSpace(item.ID)
	if item.ID == "" {
		respondFail(w, 400, "缺少节点模板 id")
		return
	}
	if item.Template == "" {
		item.Template = "{}"
	}

	// ON CONFLICT(template_id)：保留 created_at，只刷新 updated_at
	if _, err := h.db.Exec(`
		INSERT INTO workflow_node_library (template_id, type, variant, category, label, description, outputs, generated, template)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(template_id) DO UPDATE SET
			type        = excluded.type,
			variant     = excluded.variant,
			category    = excluded.category,
			label       = excluded.label,
			description = excluded.description,
			outputs     = excluded.outputs,
			generated   = excluded.generated,
			template    = excluded.template,
			updated_at  = CURRENT_TIMESTAMP`,
		item.ID, item.Type, item.Variant, item.Category, item.Label,
		item.Description, item.Outputs, item.Generated, item.Template); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, item.ID)
}

// DeleteWorkflowNodeLibraryItem 按 template_id 删除节点模板
// （DELETE /api/workflows/node-library/{id}；这里的 {id} 是字符串 template_id）
func (h *Handler) DeleteWorkflowNodeLibraryItem(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondFail(w, 400, "无效的模板ID")
		return
	}
	if _, err := h.db.Exec("DELETE FROM workflow_node_library WHERE template_id = ?", id); err != nil {
		respondFail(w, 500, err.Error())
		return
	}
	respondOK(w, "删除成功")
}
