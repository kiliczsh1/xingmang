import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import type {
  ExecutionState,
  LogEntry,
  LogLevel,
  NodeStatus,
  WNode,
  WNodeData,
  WorkflowFile
} from '@/types/workflow'

/**
 * 工作流 store —— 从源项目（魔力创作 2.1.0）workflow-DxJ6T8H-.js 逐函数直译。
 * 保留 canResume / pauseExecution / hasCheckpoint 等字段（源版即空壳，不接 UI）。
 * 本地化新增：currentRunId（关联后端 workflow_runs 行）。
 */
export const useWorkflowStore = defineStore('workflow', () => {
  // ---- state ----
  const workflows = ref<WorkflowFile[]>([])
  const currentWorkflow = ref<WorkflowFile | null>(null)
  const selectedNodeId = ref<string | null>(null)
  const isExecuting = ref(false)
  const executionLogs = ref<LogEntry[]>([])
  const executionStates = ref<ExecutionState[]>([])
  const executionResult = ref<{ success: boolean; error?: string; [k: string]: any } | null>(null)
  const executionProgress = ref(0)
  const isPaused = ref(false)
  const hasCheckpoint = ref(false)
  const checkpointInfo = ref<any>(null)
  const showVariablesDialog = ref(false)
  const variablesForm = ref<Record<string, any>>({})
  const showGenerateDialog = ref(false)
  const generateRequirement = ref('')
  const isGenerating = ref(false)
  /** 当前执行对应的后端 run id */
  const currentRunId = ref<number | null>(null)
  /**
   * 日志里正在流式写入的条目 id，按 key 区分。
   * 并发模式下多个 AI 节点同时流式输出，用同一个 id 会互相覆盖，所以按节点分桶。
   */
  const streamLogIds = ref<Record<string, number>>({})
  /** 断点集合（节点 id） */
  const breakpoints = ref<Set<string>>(new Set())
  /** 并发模式限制（仅会话内有效） */
  const concurrencyLimits = ref<{ total: number; ai: number }>({ total: 4, ai: 2 })
  /**
   * 节点运行时状态叠加层（按节点 id 索引）。
   * ⚠ 不能只写 node.data.status：Vue Flow 自定义节点拿到的是它自己解析出的节点对象，
   * 直接改 store 里那份 data 不保证驱动画布重渲染（实测状态写进去了但卡片不变色）。
   * 所以执行期的状态/结果/循环进度统一落在这里，由画布在渲染时叠加到节点 data 上。
   */
  const nodeRuntime = ref<Record<string, Record<string, any>>>({})

  // ---- getters ----
  const hasWorkflows = computed(() => workflows.value.length > 0)
  const currentWorkflowName = computed(() => currentWorkflow.value?.name || '')
  /** 源版语义：有检查点且当前没在执行（空壳字段，保留） */
  const canResume = computed(() => hasCheckpoint.value && !isExecuting.value)
  const selectedNode = computed<WNode | null>(() => {
    if (!currentWorkflow.value || !selectedNodeId.value || !currentWorkflow.value.nodes) return null
    return currentWorkflow.value.nodes.find(n => n.id === selectedNodeId.value) || null
  })
  const hasUnsavedChanges = computed(() => currentWorkflow.value?.modified || false)

  // ---- 时间戳工具 ----
  const nowStamp = () =>
    new Date().toLocaleTimeString('zh-CN', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3
    } as Intl.DateTimeFormatOptions)

  // ---- 日志 ----
  function addLog(message: string, type: LogLevel = 'info') {
    const t = nowStamp()
    executionLogs.value.push({ timestamp: t, time: t, message, type, level: type, id: Date.now() + Math.random() })
  }

  function beginStreamLog(key = 'default') {
    const id = Date.now() + Math.random()
    streamLogIds.value = { ...streamLogIds.value, [key]: id }
    executionLogs.value.push({
      timestamp: nowStamp(),
      message: '',
      type: 'info',
      id,
      stream: true,
      isStreaming: true
    })
  }

  function updateStreamLog(message: string, key = 'default') {
    if (!streamLogIds.value[key]) beginStreamLog(key)
    const sid = streamLogIds.value[key]
    const idx = executionLogs.value.findIndex(l => l.id === sid)
    if (idx !== -1) executionLogs.value[idx] = { ...executionLogs.value[idx], message }
  }

  function endStreamLog(key = 'default') {
    const sid = streamLogIds.value[key]
    if (sid !== undefined) {
      const idx = executionLogs.value.findIndex(l => l.id === sid)
      if (idx !== -1) executionLogs.value[idx] = { ...executionLogs.value[idx], isStreaming: false }
    }
    const next = { ...streamLogIds.value }
    delete next[key]
    streamLogIds.value = next
  }

  function clearLogs() {
    executionLogs.value = []
    executionStates.value = []
    streamLogIds.value = {}
  }

  function clearStates() {
    executionStates.value = []
    executionResult.value = null
  }

  function updateStateStatus(name: string, status: string) {
    const idx = executionStates.value.findIndex(s => s.name === name)
    if (idx >= 0) {
      executionStates.value[idx].status = status
      executionStates.value[idx].timestamp = Date.now()
    } else {
      executionStates.value.push({ name, status, timestamp: Date.now() })
    }
  }

  // ---- 工作流增删改 ----
  function setCurrentWorkflow(wf: WorkflowFile | null) {
    nodeRuntime.value = {}
    currentWorkflow.value = wf
  }

  function addWorkflow(wf: WorkflowFile) {
    if (!workflows.value.find(w => w.id === wf.id)) workflows.value.push(wf)
  }

  function updateWorkflow(id: number, patch: Partial<WorkflowFile>) {
    const idx = workflows.value.findIndex(w => w.id === id)
    if (idx >= 0) workflows.value[idx] = { ...workflows.value[idx], ...patch }
  }

  function removeWorkflow(id: number) {
    const idx = workflows.value.findIndex(w => w.id === id)
    if (idx >= 0) workflows.value.splice(idx, 1)
  }

  // ---- 执行状态机（与源版一一对应）----
  function startExecution() {
    isExecuting.value = true
    isPaused.value = false
    clearLogs()
    executionResult.value = null
    executionProgress.value = 0
  }

  function completeExecution(result: any) {
    isExecuting.value = false
    isPaused.value = false
    hasCheckpoint.value = false
    checkpointInfo.value = null
    executionResult.value = result
    executionProgress.value = 100
  }

  function pauseExecution(info: any) {
    isExecuting.value = false
    isPaused.value = true
    hasCheckpoint.value = true
    checkpointInfo.value = info
  }

  /** 从断点继续：恢复执行态，保留 checkpoint 供 UI 显示来源节点 */
  function resumeExecution() {
    isExecuting.value = true
    isPaused.value = false
    checkpointInfo.value = null
  }

  // ---- 断点 ----
  function toggleBreakpoint(id: string) {
    const next = new Set(breakpoints.value)
    if (next.has(id)) next.delete(id)
    else next.add(id)
    breakpoints.value = next
  }

  /** 显式设置（星域「分步」按钮切换时用，避免与右键菜单的 toggle 语义互相干扰） */
  function setBreakpoint(id: string, on: boolean) {
    const next = new Set(breakpoints.value)
    if (on) next.add(id)
    else next.delete(id)
    breakpoints.value = next
  }

  const hasBreakpoint = (id: string) => breakpoints.value.has(id)

  function clearBreakpoints() {
    breakpoints.value = new Set()
  }

  /** 清掉单个节点的运行期叠加（重跑该节点前调用） */
  function clearNodeRuntime(id: string) {
    if (!nodeRuntime.value[id]) return
    const next = { ...nodeRuntime.value }
    delete next[id]
    nodeRuntime.value = next
  }

  function handleExecutionError(err: Error) {
    isExecuting.value = false
    executionResult.value = { success: false, error: err.message }
    addLog(`❌ 执行失败: ${err.message}`, 'error')
  }

  function resetExecution() {
    isExecuting.value = false
    isPaused.value = false
    hasCheckpoint.value = false
    checkpointInfo.value = null
    executionResult.value = null
    executionProgress.value = 0
    clearLogs()
  }

  // ---- 节点 ----
  function selectNode(id: string | null) {
    selectedNodeId.value = id
  }

  function updateNodeStatus(
    id: string,
    status: NodeStatus,
    result: unknown = null,
    error: string | null = null,
    extra: Record<string, any> | null = null
  ) {
    const patch: Record<string, any> = { status }
    if (result !== null) patch.result = result
    if (error !== null) patch.error = error
    // 执行期附加信息（耗时 startedAt / durationMs 等）
    if (extra) Object.assign(patch, extra)
    nodeRuntime.value[id] = { ...(nodeRuntime.value[id] || {}), ...patch }

    const node = currentWorkflow.value?.nodes?.find(n => n.id === id)
    if (!node) return
    if (!node.data) node.data = {} as WNodeData
    Object.assign(node.data, patch)
  }

  function updateNodeData(id: string, patch: Record<string, any>) {
    nodeRuntime.value[id] = { ...(nodeRuntime.value[id] || {}), ...patch }
    const node = currentWorkflow.value?.nodes?.find(n => n.id === id)
    if (!node) return
    if (!node.data) node.data = {} as WNodeData
    Object.assign(node.data, patch)
  }

  function resetAllNodeStatus() {
    nodeRuntime.value = {}
    if (!currentWorkflow.value?.nodes) return
    currentWorkflow.value.nodes.forEach(n => {
      if (!n.data) n.data = {} as WNodeData
      n.data.status = 'idle'
      n.data.result = null
      n.data.error = null
      // 上一轮的耗时不能留到本轮（节点卡片会显示过期数字）
      n.data.durationMs = undefined
    })
  }

  // ---- 日志别名（执行器用）----
  function addExecutionLog(entry: { message: string; level?: LogLevel; type?: LogLevel }) {
    addLog(entry.message, entry.level || entry.type || 'info')
  }

  function clearExecutionLogs() {
    clearLogs()
  }

  function setExecuting(v: boolean) {
    isExecuting.value = v
  }

  // ---- 结果 ----
  function getFinalOutput(): Record<string, unknown> | null {
    if (!currentWorkflow.value) return null
    const out: Record<string, unknown> = {}
    const nodes = currentWorkflow.value.nodes
    if (nodes && Array.isArray(nodes)) {
      nodes.forEach(n => {
        if (n.data?.status === 'success' && n.data.result) {
          out[n.data.label || n.type] = n.data.result
        }
      })
    }
    if (executionResult.value && typeof executionResult.value === 'object') {
      Object.assign(out, executionResult.value)
    }
    return out
  }

  async function copyOutputToClipboard(): Promise<{ success: boolean; message?: string; error?: string }> {
    const output = getFinalOutput()
    if (!output || Object.keys(output).length === 0) return { success: false, error: '没有可复制的内容' }
    try {
      let text = '工作流执行结果\n\n'
      text += '==================\n\n'
      for (const [key, value] of Object.entries(output)) {
        text += `${key}:\n`
        text += `${typeof value === 'string' ? value : JSON.stringify(value, null, 2)}\n\n`
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text)
        return { success: true, message: '已复制到剪贴板' }
      }
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      document.body.appendChild(ta)
      ta.select()
      const ok = document.execCommand('copy')
      document.body.removeChild(ta)
      return ok ? { success: true, message: '已复制到剪贴板' } : { success: false, error: '复制失败' }
    } catch (e: any) {
      return { success: false, error: e?.message || '复制失败' }
    }
  }

  return {
    // state
    workflows,
    currentWorkflow,
    selectedNodeId,
    selectedNode,
    isExecuting,
    executionLogs,
    executionStates,
    executionResult,
    executionProgress,
    isPaused,
    hasCheckpoint,
    checkpointInfo,
    showVariablesDialog,
    variablesForm,
    showGenerateDialog,
    generateRequirement,
    isGenerating,
    currentRunId,
    nodeRuntime,
    streamLogIds,
    breakpoints,
    setBreakpoint,
    concurrencyLimits,
    // getters
    hasUnsavedChanges,
    hasWorkflows,
    currentWorkflowName,
    canResume,
    // actions
    addLog,
    clearLogs,
    clearStates,
    updateStateStatus,
    setCurrentWorkflow,
    startExecution,
    completeExecution,
    pauseExecution,
    resumeExecution,
    toggleBreakpoint,
    hasBreakpoint,
    clearBreakpoints,
    clearNodeRuntime,
    handleExecutionError,
    resetExecution,
    addWorkflow,
    updateWorkflow,
    removeWorkflow,
    beginStreamLog,
    updateStreamLog,
    endStreamLog,
    selectNode,
    updateNodeStatus,
    updateNodeData,
    resetAllNodeStatus,
    addExecutionLog,
    clearExecutionLogs,
    setExecuting,
    getFinalOutput,
    copyOutputToClipboard
  }
})
