import api from '@/api'
import type {
  ApiResponse
} from '@/types'
import type {
  WorkflowRecord,
  WorkflowRun,
  WorkflowOutputFile,
  WorkflowFile
} from '@/types/workflow'

/** 后端 graph/config 以 JSON 字符串存储，这里统一解析成结构体 */
const parseJson = <T>(value: unknown, fallback: T): T => {
  if (value && typeof value === 'object') return value as T
  if (typeof value === 'string' && value.trim()) {
    try {
      return JSON.parse(value) as T
    } catch {
      return fallback
    }
  }
  return fallback
}

/** 后端行 → 前端内存态 */
export const toWorkflowFile = (row: WorkflowRecord): WorkflowFile => {
  const graph = parseJson<{ nodes?: any[]; edges?: any[] }>(row.graph, { nodes: [], edges: [] })
  return {
    id: row.id,
    name: row.name,
    description: row.description || '',
    category: row.category || '',
    nodes: Array.isArray(graph.nodes) ? graph.nodes : [],
    edges: Array.isArray(graph.edges) ? graph.edges : [],
    config: parseJson<Record<string, any>>(row.config, {}),
    modified: false,
    created_at: row.created_at,
    updated_at: row.updated_at,
    run_count: row.run_count
  }
}

export const workflowAPI = {
  /** 列表 */
  getAll: () =>
    api.get<any, ApiResponse<WorkflowRecord[]>>('/workflows'),

  getOne: (id: number) =>
    api.get<any, ApiResponse<WorkflowRecord>>(`/workflows/${id}`),

  /**
   * 新建。⚠ 边界兜底：调用方若把本方法包成了事件处理器，name 可能收到 MouseEvent 这类对象，
   * 后端按 string 反序列化会报「请求体格式错误」，所以这里强制归一化成字符串。
   */
  create: (data: { name: string; description?: string; category?: string; graph?: string; config?: string }) =>
    api.post<any, ApiResponse<WorkflowRecord>>('/workflows', {
      name: typeof data.name === 'string' && data.name.trim() ? data.name : '未命名工作流',
      description: typeof data.description === 'string' ? data.description : '',
      category: typeof data.category === 'string' ? data.category : '',
      graph: typeof data.graph === 'string' ? data.graph : '{"nodes":[],"edges":[]}',
      config: typeof data.config === 'string' ? data.config : '{}'
    }),

  /** 保存图：只传 graph，避免覆盖 name/description */
  saveGraph: (id: number, graph: { nodes: unknown[]; edges: unknown[] }) =>
    api.put<any, ApiResponse<WorkflowRecord>>(`/workflows/${id}`, {
      graph: JSON.stringify(graph)
    }),

  update: (id: number, data: Record<string, unknown>) =>
    api.put<any, ApiResponse<WorkflowRecord>>(`/workflows/${id}`, data),

  remove: (id: number) =>
    api.delete<any, ApiResponse>(`/workflows/${id}`),

  // ---- 执行记录 ----
  createRun: (workflowId: number, outputDir?: string) =>
    api.post<any, ApiResponse<WorkflowRun>>('/workflow-runs', {
      workflow_id: workflowId,
      output_dir: outputDir || ''
    }),

  updateRun: (runId: number, data: Partial<Pick<WorkflowRun, 'status' | 'logs' | 'context' | 'error'>>) =>
    api.put<any, ApiResponse>(`/workflow-runs/${runId}`, data),

  listRuns: (workflowId?: number, limit = 50) =>
    api.get<any, ApiResponse<WorkflowRun[]>>('/workflow-runs', {
      params: { workflow_id: workflowId, limit }
    }),

  removeRun: (runId: number) =>
    api.delete<any, ApiResponse>(`/workflow-runs/${runId}`),

  // ---- 落盘文件 ----
  writeOutput: (payload: {
    run_id: number
    workflow_id?: number
    kind?: string
    rel_path: string
    content: string
    append?: boolean
    /** 'base64' 时后端按二进制解码写盘（图片/视频输出节点用） */
    encoding?: string
  }) => api.post<any, ApiResponse<{ rel_path: string; abs_path: string; size: number; output_dir: string }>>('/workflow-outputs', payload),

  mkdir: (runId: number, relPath: string) =>
    api.post<any, ApiResponse<{ abs_path: string }>>('/workflow-outputs/mkdir', {
      run_id: runId,
      rel_path: relPath
    }),

  listOutputs: (runId: number) =>
    api.get<any, ApiResponse<WorkflowOutputFile[]>>('/workflow-outputs', {
      params: { run_id: runId }
    }),

  readOutput: (runId: number, relPath: string) =>
    api.get<any, ApiResponse<{ rel_path: string; content: string }>>('/workflow-outputs/read', {
      params: { run_id: runId, rel_path: relPath }
    }),

  runOutputDir: (runId: number) =>
    api.get<any, ApiResponse<{ output_dir: string; root: string }>>('/workflow-outputs/root', {
      params: { run_id: runId }
    })
}

export default workflowAPI
