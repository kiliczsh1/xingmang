// 工作流画布类型定义
// 字段以上游参考工程反压缩产物 workflow-DxJ6T8H-.js / executor-C2yGHSMY.js 的实际用法为准。

/**
 * 节点类型。
 *
 * ⚠ 命名口径：星芒原生节点与星域移植节点**并列共存**，不合并同类项 ——
 * 例如星芒的 aiProcess（AI 处理）与星域的 aiModel（AI 模型）是两个独立节点。
 */
export type NodeType =
  | 'start'
  | 'aiProcess'
  | 'text'
  | 'merge'
  | 'loop'
  | 'context'
  | 'end'
  /** 注释（不参与执行，仅画布标注） */
  | 'comment'
  /** 中转：聚合上游数据后转发 */
  | 'transfer'
  /* ---- 以下为星域移植节点（与上面星芒原生节点并列，不合并） ---- */
  /** 星域·文本输入：手动文本 + 占位符名 */
  | 'textInput'
  /** 星域·AI 模型：独立于 aiProcess 的 AI 节点 */
  | 'aiModel'
  /** 星域·文本处理：format / uppercase / lowercase */
  | 'textProcessor'
  /** 星域·提示词节点：携带提示词库内容供下游注入 */
  | 'promptNode'
  /** 星域·图片输出：保存并预览图片 */
  | 'imageOutput'
  /** 星域·视频输出：保存并预览视频 */
  | 'videoOutput'
  /** 星域·文生视频 */
  | 'videoGeneration'
  /** 星域·图生视频 */
  | 'imageToVideo'
  /** 文件输入：读取本地文本文件作为变量 */
  | 'fileInput'
  /** 文件输出：把上游内容落到本次执行的工作目录 */
  | 'fileOutput'
  /** 文本输出：按格式写出文件，并可在节点里预览 */
  | 'textOutput'
  /** 文生图 */
  | 'imageGeneration'
  /** 图片分析（多模态，SSE 流式） */
  | 'imageAnalyze'

/** 节点运行状态（画布配色依据） */
export type NodeStatus = 'idle' | 'running' | 'success' | 'error' | 'cancelled'

/** 日志级别 */
export type LogLevel = 'info' | 'success' | 'warning' | 'error'

/** 执行日志条目；stream 为 true 表示流式条目（原地更新 message） */
export interface LogEntry {
  id?: number
  timestamp: string
  time?: string
  message: string
  type: LogLevel
  level?: LogLevel
  /** 流式日志占位标记 */
  stream?: boolean
  /** 流式是否仍在进行 */
  isStreaming?: boolean
}

/** 变量状态（executionStates） */
export interface ExecutionState {
  name: string
  status: string
  timestamp: number
}

/** 单节点的一次执行历史（上限 100 条） */
export interface NodeExecutionRecord {
  id: number
  timestamp: string
  status: 'running' | 'success' | 'error' | 'cancelled'
  inputs: Record<string, unknown>
  outputs: Record<string, unknown>
  progress: unknown
  error: string | null
  data?: unknown
}

/** 模型快照：选中模型时把 provider 信息一并存进节点 config，执行时优先用快照 */
export interface ModelSnapshot {
  configId?: number
  providerId?: number
  providerName: string
  baseURL: string
  apiFormat?: string
  modelName: string
  apiKey: string
}

/** 上下文变量（新格式为对象数组） */
export interface ContextVariableRef {
  variableName: string
}

/**
 * 节点配置。各节点类型共用一个宽松结构，字段名与源项目一致：
 * - start:   initialVariables: Record<string, unknown>
 * - aiProcess: modelId / configId / modelConfig / systemPrompt / systemPromptId /
 *              userPrompt / prompt / contextVariables / contextMaxItems / maxTokens / temperature
 * - text:    text
 * - merge:   mergeType('concat'|'array'|'object') / inputVariables / separator
 * - loop:    loopType('array'|'fixed'|'condition') / arraySource{variableName} / fixedCount /
 *            iterations / maxIterations / historyWindow / keepContext
 * - context: action / variableName / sourceVariable(s) / inputVariables / arrayVariable /
 *            indexVariable / maxHistory / includeTitle / chapterIndex / chapterPlans / content / outputDir
 * - end:     outputDir / outputFileName / outputFormat / separator / includeOutputs / autoSave
 */
export interface NodeConfig {
  [key: string]: any
}

/** 画布节点 data */
export interface WNodeData {
  label: string
  status: NodeStatus
  result: unknown
  error: string | null
  config: NodeConfig
  executionHistory?: NodeExecutionRecord[]
  iterations?: number
  currentIteration?: number
  totalChapters?: number
  currentChapterTitle?: string
  /** 本次执行该节点的开始时间戳（执行期由 store 叠加，不落盘） */
  startedAt?: number
  /** 本次执行该节点的耗时毫秒（执行期由 store 叠加，不落盘） */
  durationMs?: number
  /** 本次执行该节点的结束时间戳（执行期由 store 叠加，不落盘） */
  finishedAt?: number
  /** 是否被设为断点（由 store.breakpoints 叠加，不落盘） */
  breakpoint?: boolean
  /** AI 节点内嵌预览的流式文本（执行期由执行器叠加，不落盘） */
  streamText?: string
  /** 文本输出节点的预览内容（执行期叠加，不落盘） */
  preview?: string
  /** 预览是否被截断 */
  previewTruncated?: boolean
  /** 文本输出/文件输出写出的相对路径 */
  outputFile?: string
  /** 文生图节点生成的图片（执行期叠加，不落盘） */
  images?: Array<{ b64?: string; url?: string; revised_prompt?: string }>
  /** 视频输出/生成节点的视频地址（执行期叠加，不落盘） */
  videoUrl?: string
}

/** 画布节点 */
export interface WNode {
  id: string
  type: NodeType
  position: { x: number; y: number }
  data: WNodeData
  width?: number
  height?: number
  selected?: boolean
}

/** 画布连线；loop 节点的 sourceHandle 为 'loop-body' / 'loop-done' */
export interface WEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string | null
  targetHandle?: string | null
  type?: string
  animated?: boolean
  label?: string
}

/** 工作流全局配置 */
export interface WorkflowConfig {
  outputDir?: string
  [key: string]: any
}

/** 工作流（前端内存态） */
export interface WorkflowFile {
  id?: number
  name: string
  description?: string
  /** 分类名；空串表示「未分类」 */
  category?: string
  nodes: WNode[]
  edges: WEdge[]
  config: WorkflowConfig
  /** 与后端最后一次同步相比是否有改动 */
  modified?: boolean
  created_at?: string
  updated_at?: string
  run_count?: number
}

/** 后端 workflows 表行（graph/config 是 JSON 字符串） */
export interface WorkflowRecord {
  id: number
  name: string
  description: string
  category?: string
  graph: string
  config: string
  created_at: string
  updated_at: string
  run_count?: number
}

/** 一次执行记录 */
export interface WorkflowRun {
  id: number
  workflow_id: number
  status: 'running' | 'success' | 'error' | 'stopped' | string
  logs: string
  context: string
  error: string
  output_dir: string
  started_at: string
  finished_at: string
}

/** 落盘文件索引行 */
export interface WorkflowOutputFile {
  id: number
  run_id: number
  workflow_id: number
  kind: string
  rel_path: string
  size: number
  created_at: string
}

/** 执行器单节点返回 */
export interface NodeExecuteResult {
  outputs: Record<string, unknown>
  data: unknown
  aborted?: boolean
}

/** 节点执行器接口 */
export interface NodeExecutor {
  execute(
    node: WNode,
    context: Record<string, any>,
    logger: NodeLogger,
    record?: NodeExecutionRecord
  ): Promise<NodeExecuteResult>
}

/** 执行器注入给节点的日志器 */
export interface NodeLogger {
  info(message: string): void
  success(message: string): void
  warning(message: string): void
  error(message: string): void
}

/** 节点元信息（画布 / 属性面板共用） */
export interface NodeTypeMeta {
  type: NodeType
  label: string
  color: string
  description: string
  /** 是否有输入口 */
  hasTarget: boolean
  /** 输出口 id 列表；空数组表示无输出（end） */
  sources: string[]
}

/**
 * 节点执行器能拿到的执行器能力（loop / context 需要）。
 * 单独抽出来是为了避免 nodeExecutors 与 executor.ts 形成运行时循环依赖。
 */
export interface WorkflowExecutorHost {
  workflow: WorkflowFile
  context: Record<string, any>
  stopped: boolean
  store: any
  /** 本次执行的后端 run id；为 null 表示不落盘 */
  runId?: number | null
  /** 本次执行的落盘目录（后端返回的绝对路径）；为 null 表示不落盘 */
  workspaceDir?: string | null
  /** 主执行器：执行单个节点 */
  executeNode(node: WNode, opts?: { preventFollowEdges?: boolean }): Promise<NodeExecuteResult>
  /** 回写节点 data */
  updateNodeData(id: string, patch: Record<string, any>): void
  /** 落盘一个章节文件（workspaceDir 存在时） */
  saveChapter(title: string, content: string, index?: number | null): Promise<void>
  /** 落盘到本次执行目录下（relPath 相对根）；encoding='base64' 时按二进制写 */
  writeFile?(relPath: string, content: string, kind: string, append?: boolean, encoding?: string): Promise<void>
}
