import workflowAPI from './workflowAPI'
import { isAbortLike, AiProcessExecutor } from './nodeExecutors/aiProcessExecutor'
import { StartExecutor } from './nodeExecutors/startExecutor'
import { TextExecutor } from './nodeExecutors/textExecutor'
import { MergeExecutor } from './nodeExecutors/mergeExecutor'
import { LoopExecutor } from './nodeExecutors/loopExecutor'
import { ContextExecutor } from './nodeExecutors/contextExecutor'
import { EndExecutor } from './nodeExecutors/endExecutor'
import {
  CommentExecutor,
  FileInputExecutor,
  FileOutputExecutor,
  TextOutputExecutor,
  ImageAnalyzeExecutor,
  ImageGenerationExecutor,
  TransferExecutor
} from './nodeExecutors/extraExecutors'
import {
  AiModelExecutor,
  ImageOutputExecutor,
  ImageToVideoExecutor,
  PromptNodeExecutor,
  TextInputExecutor,
  TextProcessorExecutor,
  VideoGenerationExecutor,
  VideoOutputExecutor
} from './nodeExecutors/starExecutors'
import { collectDescendantIds, upstreamInTopoOrder } from './graphUtils'
import { ConcurrentScheduler, DEFAULT_LIMITS, type ConcurrencyLimits } from './scheduler'
import type {
  NodeExecuteResult,
  NodeExecutionRecord,
  NodeExecutor,
  NodeLogger,
  NodeStatus,
  WNode,
  WorkflowFile
} from '@/types/workflow'

/** 执行引擎模式：顺序（沿边递归，默认） / 并发（拓扑就绪队列） */
export type ExecutionMode = 'sequential' | 'concurrent'

/**
 * 工作流主执行器 —— 直译源项目 executor-C2yGHSMY.js 的 WorkflowExecutor（F 类）。
 *
 * 本地化差异：
 * - `window.electronAPI.file.*` 全部换成 Go 的 /api/workflow-outputs（落盘被限制在数据目录内）
 * - 落盘根目录由后端在建 run 时确定，前端只传相对路径
 * - 停止按钮在 `stop()` 里额外 abort 正在进行的 SSE 请求
 *
 * 二期新增：
 * - mode='concurrent' 时改由 ConcurrentScheduler 调度（顺序引擎原样保留）
 * - 断点挂起 / 继续 / 从指定节点重跑 / 单节点测试运行
 */
export class WorkflowExecutor {
  workflow: WorkflowFile
  store: any
  context: Record<string, any> = {}
  stopped = false
  /** 本次执行的落盘目录（绝对路径，由后端返回）；为空表示不落盘 */
  workspaceDir: string | null = null
  chapterCounter = 0

  /** 执行引擎模式 */
  mode: ExecutionMode = 'sequential'
  /** 并发度限制（仅并发模式生效） */
  limits: Partial<ConcurrencyLimits> = {}

  workflowId: number | null
  runId: number | null
  abortController: AbortController | null = null

  /**
   * 断点挂起队列。并发模式下可能同时命中多个断点，
   * 只用一个 resolve 会让后到的节点永久悬挂，所以排队逐个放行。
   */
  private pauseQueue: Array<{
    nodeId: string
    label: string
    reason?: 'breakpoint' | 'stopped'
    resolve: (ok: boolean) => void
  }> = []
  private activePause: { nodeId: string; label: string; reason?: string } | null = null
  /** 已放行过一次的断点（避免「继续」后在同一节点原地再次暂停） */
  private skipBreakpointOnce: string | null = null
  /**
   * 被「停止」按钮单独中止的节点（星域：node.properties.stopGeneration）。
   * 与整体 stop() 的区别：只中止该节点本次生成，随后工作流转入分步暂停而非整体失败。
   */
  private nodeStopFlags = new Set<string>()
  /** 每个节点自己的 AbortController（星域：currentAbortController），与全局停止联动 */
  private nodeAborts = new Map<string, AbortController>()
  /** 节点级停止后强制挂起一次，等用户点「继续」再往下走 */
  private forcePauseOnce: string | null = null
  /** 单节点测试模式：不建工作目录、不落盘 */
  testMode = false

  private aiExecutor: AiProcessExecutor
  private startExecutor = new StartExecutor()
  private textExecutor = new TextExecutor()
  private mergeExecutor = new MergeExecutor()
  private loopExecutor: LoopExecutor
  private contextExecutor: ContextExecutor
  private endExecutor = new EndExecutor()
  private commentExecutor = new CommentExecutor()
  private fileInputExecutor = new FileInputExecutor()
  private textOutputExecutor: TextOutputExecutor
  private transferExecutor: TransferExecutor
  private fileOutputExecutor: FileOutputExecutor
  private imageGenerationExecutor: ImageGenerationExecutor
  private imageAnalyzeExecutor: ImageAnalyzeExecutor
  // 星域移植节点（与上面星芒原生节点并列）
  private textInputExecutor = new TextInputExecutor()
  private aiModelExecutor: AiModelExecutor
  private textProcessorExecutor = new TextProcessorExecutor()
  private promptNodeExecutor = new PromptNodeExecutor()
  private videoGenerationExecutor: VideoGenerationExecutor
  private imageToVideoExecutor: ImageToVideoExecutor
  private imageOutputExecutor: ImageOutputExecutor
  private videoOutputExecutor: VideoOutputExecutor

  constructor(
    workflow: WorkflowFile,
    store: any,
    opts: {
      workflowId?: number | null
      runId?: number | null
      outputDir?: string | null
      mode?: ExecutionMode
      limits?: Partial<ConcurrencyLimits>
    } = {}
  ) {
    this.workflow = workflow
    this.store = store
    this.workflowId = opts.workflowId ?? workflow.id ?? null
    this.runId = opts.runId ?? null
    this.workspaceDir = opts.outputDir || null
    this.mode = opts.mode || 'sequential'
    this.limits = opts.limits || {}
    this.aiExecutor = new AiProcessExecutor(store)
    this.loopExecutor = new LoopExecutor(this)
    this.contextExecutor = new ContextExecutor(this)
    this.fileOutputExecutor = new FileOutputExecutor(this)
    this.textOutputExecutor = new TextOutputExecutor(this)
    this.imageGenerationExecutor = new ImageGenerationExecutor(this)
    this.imageAnalyzeExecutor = new ImageAnalyzeExecutor(this)
    this.aiModelExecutor = new AiModelExecutor(this)
    this.transferExecutor = new TransferExecutor(this)
    this.videoGenerationExecutor = new VideoGenerationExecutor(this)
    this.imageToVideoExecutor = new ImageToVideoExecutor(this)
    this.imageOutputExecutor = new ImageOutputExecutor(this)
    this.videoOutputExecutor = new VideoOutputExecutor(this)
  }

  // ---------------------------------------------------------------- 工作目录

  /** 建目录结构 + 落一份 workflow_config.json（源版靠 Electron file API，这里走后端） */
  async initializeWorkspaceDirectory() {
    if (!this.runId) {
      this.store.addExecutionLog({
        type: 'warning',
        message: '⚠️ 本次执行没有落盘目录（工作流可能尚未保存），输出文件功能将被禁用',
        timestamp: new Date().toISOString()
      })
      this.workspaceDir = null
      return
    }
    try {
      const endNode = this.workflow.nodes.find(n => n.type === 'end')
      let configuredDir = endNode?.data?.config?.outputDir || ''
      if (!configuredDir) configuredDir = this.workflow.config?.outputDir || ''
      if (configuredDir) {
        this.store.addExecutionLog({
          type: 'info',
          message: `ℹ️ 结束节点配置的输出目录：${configuredDir}（实际落盘固定在数据目录的 workflow_outputs/ 内）`,
          timestamp: new Date().toISOString()
        })
      }
      for (const sub of ['nodes', 'chapters', 'logs']) {
        await workflowAPI.mkdir(this.runId, sub)
      }
      await this.writeFile('workflow_config.json', JSON.stringify(this.workflow, null, 2), 'config')
      this.store.addExecutionLog({
        type: 'success',
        message: `📁 工作目录已就绪: ${this.workspaceDir}`,
        timestamp: new Date().toISOString()
      })
    } catch (e: any) {
      this.workspaceDir = null
      this.store.addExecutionLog({
        type: 'error',
        message: `❌ 创建工作目录失败: ${e?.message || e}`,
        timestamp: new Date().toISOString()
      })
    }
  }

  /** 落盘到本次执行目录（fileOutput 节点也用这个） */
  async writeFile(relPath: string, content: string, kind: string, append = false, encoding?: string) {
    if (!this.runId) return
    await workflowAPI.writeOutput({
      run_id: this.runId,
      workflow_id: this.workflowId ?? 0,
      kind,
      rel_path: relPath,
      content,
      append,
      encoding
    })
  }

  private safeName(raw: string) {
    return String(raw || 'node').replace(/[/\\:*?"<>|]/g, '_')
  }

  /** 每节点输出存一份 txt（nodes/ 下） */
  async saveNodeOutput(node: WNode, data: unknown) {
    if (!this.workspaceDir || !this.runId) return
    try {
      const label = this.safeName(node.data.label || node.type)
      const filename = `${label}_${Date.now()}.txt`
      let text = ''
      text += `节点: ${node.data.label || node.type}\n`
      text += `类型: ${node.type}\n`
      text += `时间: ${new Date().toLocaleString()}\n`
      text += `${'='.repeat(50)}\n\n`
      if (typeof data === 'string') text += data
      else if (typeof data === 'object') text += JSON.stringify(data, null, 2)
      else text += String(data)

      await this.writeFile(`nodes/${filename}`, text, 'node')
      this.store.addExecutionLog({
        type: 'info',
        message: `💾 节点输出已保存: ${filename}`,
        timestamp: new Date().toISOString()
      })
    } catch (e) {
      console.error('保存节点输出失败:', e)
    }
  }

  /** 章节文件（chapters/000_标题.md） */
  async saveChapter(title: string | null, content: string, index: number | null = null) {
    if (!this.workspaceDir || !this.runId) return
    try {
      this.chapterCounter++
      const num = index || this.chapterCounter
      const name = title || `第${num}章`
      const safe = this.safeName(name).trim()
      const filename = `${String(num).padStart(3, '0')}_${safe}.md`
      let text = `# ${name}\n\n`
      text += `生成时间: ${new Date().toLocaleString()}\n\n`
      text += `---\n\n`
      text += content
      await this.writeFile(`chapters/${filename}`, text, 'chapter')
      this.store.addExecutionLog({
        type: 'success',
        message: `📖 章节已保存: ${filename}`,
        timestamp: new Date().toISOString()
      })
    } catch (e) {
      console.error('保存章节失败:', e)
    }
  }

  async saveExecutionLogs() {
    if (!this.workspaceDir || !this.runId) return
    try {
      const list = this.store.executionLogs || []
      let text = '工作流执行日志\n'
      text += `工作流: ${this.workflow.name || '未命名'}\n`
      text += `开始时间: ${new Date().toLocaleString()}\n`
      text += `${'='.repeat(80)}\n\n`
      list.forEach((l: any) => {
        const time = l.timestamp || l.time || ''
        const type = l.type || l.level || 'info'
        const icon = type === 'error' ? '❌' : type === 'success' ? '✅' : type === 'warning' ? '⚠️' : 'ℹ️'
        text += `[${time}] ${icon} ${l.message}\n`
      })
      await this.writeFile('logs/execution_log.txt', text, 'log')
    } catch (e) {
      console.error('保存执行日志失败:', e)
    }
  }

  /** 组装最终输出（按结束节点 includeOutputs 或智能挑选） */
  private buildFinalText(): string {
    const endNode = this.workflow.nodes.find(n => n.type === 'end')
    const config = endNode?.data?.config || {}
    const separator = config.separator || '\n\n'
    let text = ''

    if (Array.isArray(config.includeOutputs) && config.includeOutputs.length > 0) {
      for (const name of config.includeOutputs as string[]) {
        const value = this.context[name]
        if (value === null || value === undefined) continue
        text += this.renderFinalValue(value, separator)
        text += separator
      }
    } else {
      for (const name of ['chapterHistory', 'allChapters', 'finalOutput', 'lastOutput', 'lastAIOutput', 'aiResult']) {
        const value = this.context[name]
        if (value === null || value === undefined) continue
        text += this.renderFinalValue(value, separator)
        break
      }
    }
    return text
  }

  private renderFinalValue(value: any, separator: string): string {
    if (Array.isArray(value)) {
      const CHUNK = 50
      const parts: string[] = []
      for (let i = 0; i < value.length; i += CHUNK) {
        const piece = value
          .slice(i, i + CHUNK)
          .map(v => {
            if (typeof v === 'object' && v !== null) {
              if (v.title && v.content) return `# ${v.title}\n\n${v.content}`
              if (v.title && v.summary) return `## ${v.title}\n${v.summary}`
              return JSON.stringify(v, null, 2)
            }
            return String(v)
          })
          .join(separator)
        parts.push(piece)
      }
      return parts.join(separator)
    }
    if (typeof value === 'object' && value !== null) return JSON.stringify(value, null, 2)
    return String(value)
  }

  async saveFinalOutput() {
    if (!this.workspaceDir || !this.runId) return
    try {
      const endNode = this.workflow.nodes.find(n => n.type === 'end')
      const config = endNode?.data?.config || {}
      const text = this.buildFinalText()
      if (!text.trim()) return
      const filename = `final_output.${config.outputFormat || 'md'}`
      let body = `# ${this.workflow.name || '工作流输出'}\n\n`
      body += `生成时间: ${new Date().toLocaleString()}\n\n`
      body += `---\n\n`
      body += text
      await this.writeFile(filename, body, 'final')
      this.store.addExecutionLog({
        type: 'success',
        message: `📄 最终输出已保存: ${filename}`,
        timestamp: new Date().toISOString()
      })
    } catch (e) {
      console.error('保存最终输出失败:', e)
    }
  }

  /** 没有工作目录时的兜底（源版弹保存对话框；本地化后只提示） */
  async saveOutputToFile() {
    const endNode = this.workflow.nodes.find(n => n.type === 'end')
    if (!endNode || !endNode.data?.config) return
    this.store.addExecutionLog({
      type: 'warning',
      message: '⚠️ 没有落盘目录，最终输出仅在页面内可用（保存工作流后重跑即可落盘）',
      timestamp: new Date().toISOString()
    })
  }

  // ---------------------------------------------------------------- 主流程

  async execute(): Promise<{ success: boolean; stopped: boolean; context: Record<string, any> }> {
    try {
      this.stopped = false
      this.abortController = new AbortController()
      this.store.setExecuting(true)
      this.store.clearExecutionLogs()
      this.store.resetAllNodeStatus()
      this.store.addExecutionLog({
        type: 'info',
        message: `🚀 开始执行工作流: ${this.workflow.name || '未命名工作流'}`,
        timestamp: new Date().toISOString()
      })
      await this.initializeWorkspaceDirectory()

      if (this.mode === 'concurrent') {
        // 并发模式：从全图入度 0 的节点起步，不强制要求 start 节点
        const scheduler = new ConcurrentScheduler(this, this.limits)
        const { completed, total } = await scheduler.run()
        this.log(`⚡ 并发执行结束：完成 ${completed}/${total} 个节点`, 'info')
      } else {
        const startNode = this.workflow.nodes.find(n => n.type === 'start')
        if (!startNode) throw new Error('未找到开始节点')
        await this.executeNode(startNode)
      }
      await this.logFinalOutput()
      if (this.workspaceDir) await this.saveFinalOutput()
      else await this.saveOutputToFile()

      this.store.addExecutionLog({
        type: 'success',
        message: this.stopped ? '⏹️ 工作流已停止' : '✅ 工作流执行完成',
        timestamp: new Date().toISOString()
      })
      return { success: true, stopped: this.stopped, context: this.context }
    } catch (e: any) {
      const msg = this.formatErrorMessage(e)
      this.store.addExecutionLog({
        type: 'error',
        message: `❌ 工作流执行失败: ${msg}`,
        timestamp: new Date().toISOString()
      })
      throw e
    } finally {
      await this.saveExecutionLogs()
      this.store.setExecuting(false)
    }
  }

  async executeNode(node: WNode, opts: { preventFollowEdges?: boolean } = {}): Promise<NodeExecuteResult> {
    const preventFollowEdges = opts.preventFollowEdges === true

    if (this.stopped) {
      this.store.addExecutionLog({
        type: 'warning',
        message: `⏹️ 已停止，跳过节点: ${node.data?.label || node.type}`,
        timestamp: new Date().toISOString()
      })
      return { outputs: {}, data: null, aborted: true }
    }

    // 上一轮的节点级停止标记在这里消费掉：停止针对的是上一次生成，重新执行时应当放行
    this.nodeStopFlags.delete(node.id)

    // 断点 / 分步：命中的话在此挂起，等用户「继续」；停止按钮会让它返回 aborted
    const proceed = await this.waitForBreakpoint(node)
    if (!proceed) {
      return { outputs: {}, data: null, aborted: true }
    }

    this.store.updateNodeStatus(node.id, 'running', undefined, undefined, { startedAt: Date.now() })
    if (!node.data.executionHistory) node.data.executionHistory = []
    const startedAt = Date.now()

    const record: NodeExecutionRecord = {
      id: Date.now(),
      timestamp: new Date().toISOString(),
      status: 'running',
      inputs: {},
      outputs: {},
      progress: null,
      error: null
    }

    this.store.addExecutionLog({
      type: 'info',
      message: `📍 执行节点: ${node.data.label || node.type}`,
      timestamp: new Date().toISOString()
    })

    let succeeded = false
    try {
      const executor = this.getNodeExecutor(node.type, node.id)
      if (!executor) throw new Error(`未知的节点类型: ${node.type}`)
      // 流式日志按节点分桶：并发模式下多个 AI 节点同时输出，共用一个条目会互相覆盖
      if (node.type === 'aiProcess') this.aiExecutor.streamKey = node.id

      await this.logNodeInput(node)
      record.inputs = this.captureRelevantContext(node)

      const result = await executor.execute(node, this.context, this.createLogger(node), record)

      await this.logNodeOutput(node, result)
      record.outputs = result?.outputs || {}
      ;(record as any).data = result?.data
      record.status = 'success'

      if (result?.outputs) Object.assign(this.context, result.outputs)
      if (result?.data !== undefined) await this.saveNodeOutput(node, result.data)

      // AI 节点的字符串输出按章节落盘（源版按 currentIteration / chapterNum 命名）
      if (node.type === 'aiProcess' && result?.data && typeof result.data === 'string') {
        const iter = this.context.currentIteration || this.context.chapterNum
        const title = this.context.currentChapterTitle || (iter ? `第${iter}章` : null)
        if (iter || title) await this.saveChapter(title, result.data, iter || null)
      }

      node.data.executionHistory.unshift(record)
      if (node.data.executionHistory.length > 100) {
        node.data.executionHistory = node.data.executionHistory.slice(0, 100)
      }
      this.store.updateNodeStatus(node.id, 'success', undefined, undefined, {
        durationMs: Date.now() - startedAt,
        finishedAt: Date.now()
      })
      succeeded = true

      if (node.type === 'end' || preventFollowEdges) return result

      if (node.type === 'loop') {
        for (const next of this.getNextNodes(node.id, 'loop-done')) {
          await this.executeNode(next)
        }
        return result
      }

      for (const next of this.getNextNodes(node.id)) {
        try {
          await this.executeNode(next)
        } catch (e: any) {
          const msg = this.formatErrorMessage(e)
          this.store.addExecutionLog({
            type: 'error',
            message: `🔻 下游节点失败（来自 ${node.data?.label || node.type} → ${next.data?.label || next.type}）: ${msg}`,
            timestamp: new Date().toISOString()
          })
          throw e
        }
      }
      return result
    } catch (e: any) {
      const msg = this.formatErrorMessage(e)
      const aborted = isAbortLike(msg)
      record.status = aborted ? 'cancelled' : 'error'
      record.error = msg
      node.data.executionHistory.unshift(record)
      if (!succeeded) {
        this.store.updateNodeStatus(node.id, (aborted ? 'idle' : 'error') as NodeStatus, undefined, undefined, {
          durationMs: Date.now() - startedAt,
          finishedAt: Date.now()
        })
      }
      this.store.addExecutionLog({
        type: aborted ? 'warning' : 'error',
        message: `${aborted ? '⏹️ 已停止于节点' : '❌ 节点执行失败'}: [${node.data?.label || node.type}] ${msg}`,
        timestamp: new Date().toISOString()
      })
      if (aborted) {
        // 节点级停止（点了「停止」）：不结束整个工作流，强制在此挂起一次，等用户「继续」
        if (this.nodeStopFlags.has(node.id) && !this.stopped) {
          this.nodeStopFlags.delete(node.id)
          this.forcePauseOnce = node.id
          this.store.addExecutionLog({
            type: 'warning',
            message: `⏸️ 节点 ${node.data?.label || node.id} 已停止生成，工作流分步暂停`,
            timestamp: new Date().toISOString()
          })
          await this.waitForBreakpoint(node)
        }
        return { outputs: {}, data: null, aborted: true }
      }
      throw e
    }
  }

  // ---------------------------------------------------------------- 断点 / 重跑 / 单节点测试

  /**
   * 命中节点断点时挂起，返回 true 表示可以继续执行，false 表示被停止。
   *
   * 三种触发来源（与星域一致）：
   * - 右键菜单设的断点（store.breakpoints）
   * - 星域「分步」按钮写的 config.controlState === 'paused'
   * - 节点级「停止」之后强制挂起一次
   */
  private async waitForBreakpoint(node: WNode): Promise<boolean> {
    if (this.stopped) return false
    const bp = this.store.breakpoints
    const byBreakpoint = !!bp && typeof bp.has === 'function' && bp.has(node.id)
    const byControlState = node.data?.config?.controlState === 'paused'
    const byStopFlag = this.forcePauseOnce === node.id
    if (!byBreakpoint && !byControlState && !byStopFlag) return true
    if (this.skipBreakpointOnce === node.id) {
      this.skipBreakpointOnce = null
      this.forcePauseOnce = null
      return true
    }
    const label = node.data?.label || node.type
    const reason: 'breakpoint' | 'stopped' = byStopFlag ? 'stopped' : 'breakpoint'
    return await new Promise<boolean>(resolve => {
      this.pauseQueue.push({ nodeId: node.id, label, reason, resolve })
      this.activateNextPause()
    })
  }

  /** 把队首断点设为「当前挂起点」（写入 store 供面板显示） */
  private activateNextPause() {
    if (this.activePause || !this.pauseQueue.length) return
    const head = this.pauseQueue[0]
    this.activePause = { nodeId: head.nodeId, label: head.label, reason: head.reason }
    this.store.pauseExecution({
      nodeId: head.nodeId,
      label: head.label,
      context: this.snapshotContext(),
      reason: head.reason
    })
    const extra = this.pauseQueue.length > 1 ? `（另有 ${this.pauseQueue.length - 1} 个断点排队）` : ''
    const reasonText = head.reason === 'stopped' ? '节点生成已停止' : '命中断点'
    this.store.addExecutionLog({
      type: 'warning',
      message: `⏸️ ${reasonText}，已暂停于节点: ${head.label}${extra}`,
      timestamp: new Date().toISOString()
    })
  }

  /**
   * 取该节点的 AbortController（与全局停止联动）。
   * 星域把 controller 存在节点属性上，这里按节点 id 管理，语义一致但更干净。
   */
  private getNodeAbort(nodeId: string): AbortController {
    let ctrl = this.nodeAborts.get(nodeId)
    // 上一轮被「停止」中止过的 controller 已 aborted，直接复用会让重跑立刻中止 → 重建
    if (ctrl && ctrl.signal.aborted) {
      this.nodeAborts.delete(nodeId)
      ctrl = undefined
    }
    if (!ctrl) {
      ctrl = new AbortController()
      this.nodeAborts.set(nodeId, ctrl)
      const globalSignal = this.abortController?.signal
      if (globalSignal) {
        if (globalSignal.aborted) ctrl.abort()
        else globalSignal.addEventListener('abort', () => ctrl?.abort(), { once: true })
      }
    }
    return ctrl
  }

  /**
   * 节点级「停止」（星域 stopAIGeneration）：只中止该节点本次生成，
   * 随后工作流转入分步暂停（等用户点「继续」），不是整体停止。
   */
  stopNode(nodeId: string) {
    this.nodeStopFlags.add(nodeId)
    const ctrl = this.nodeAborts.get(nodeId)
    if (ctrl) {
      try {
        ctrl.abort()
      } catch {
        /* 已结束 */
      }
    }
    const node = this.workflow.nodes.find(n => n.id === nodeId)
    const label = node?.data?.label || nodeId
    if (node?.data?.config) node.data.config.generatedStoppedAt = Date.now()
    // 星域口径：已生成的内容尾部追加「[用户已停止生成]」
    const d: any = node?.data
    if (d) {
      const cur = typeof d.streamText === 'string' ? d.streamText : ''
      if (cur && !cur.endsWith('[用户已停止生成]')) {
        this.store.updateNodeData(nodeId, { streamText: `${cur}\n\n[用户已停止生成]` })
      }
    }
    this.store.updateNodeStatus(nodeId, 'idle')
    this.store.addExecutionLog({
      type: 'warning',
      message: `⏹️ 已停止节点生成: ${label}（工作流将转入分步暂停）`,
      timestamp: new Date().toISOString()
    })
  }

  /** 清掉节点级停止标记（重跑 / 重复执行前调用，否则请求会被立刻中止） */
  clearNodeStopFlag(nodeId: string) {
    this.nodeStopFlags.delete(nodeId)
    this.nodeAborts.delete(nodeId)
  }

  /** UI「继续」按钮调用：放行当前断点，若还有排队的断点则接着挂起 */
  resume() {
    const head = this.pauseQueue.shift()
    const active = this.activePause
    if (!head || !active) return
    this.activePause = null
    this.skipBreakpointOnce = active.nodeId
    this.store.resumeExecution()
    this.store.addExecutionLog({
      type: 'info',
      message: `▶️ 从断点继续: ${active.label}`,
      timestamp: new Date().toISOString()
    })
    head.resolve(true)
    this.activateNextPause()
  }

  /** 当前上下文快照（截断后给挂起面板显示） */
  snapshotContext(limit = 30) {
    const out: Record<string, any> = {}
    Object.keys(this.context)
      .slice(0, limit)
      .forEach(k => {
        out[k] = this.truncateForDisplay(this.context[k], 300)
      })
    return out
  }

  /** 重置单个节点的运行期状态（重跑前调用，避免旧结果残留） */
  private resetNodeRunState(id: string) {
    this.store.clearNodeRuntime?.(id)
    const node = this.workflow.nodes.find(n => n.id === id)
    if (!node?.data) return
    node.data.status = 'idle'
    node.data.result = null
    node.data.error = null
    node.data.durationMs = undefined
    node.data.finishedAt = undefined
    node.data.startedAt = undefined
  }

  /**
   * 单节点测试运行：补齐上游（已成功的节点直接复用缓存）后只跑目标节点。
   * 不建工作目录、不落盘，仅把结果写进节点状态与日志。
   */
  async runSingleNode(nodeId: string): Promise<void> {
    const target = this.workflow.nodes.find(n => n.id === nodeId)
    if (!target) throw new Error(`节点不存在: ${nodeId}`)
    this.stopped = false
    this.testMode = true
    this.abortController = new AbortController()
    this.store.setExecuting(true)
    this.store.addExecutionLog({
      type: 'info',
      message: `🧪 单节点测试: ${target.data?.label || target.type}（上游已有的结果会复用，测试运行不落盘）`,
      timestamp: new Date().toISOString()
    })
    try {
      const upstream = upstreamInTopoOrder(this.workflow.nodes, this.workflow.edges, nodeId)
      for (const node of upstream) {
        if (this.stopped) break
        if (node.data?.status === 'success') {
          this.store.addExecutionLog({
            type: 'info',
            message: `♻️ 复用上游缓存，跳过: ${node.data?.label || node.type}`,
            timestamp: new Date().toISOString()
          })
          continue
        }
        this.resetNodeRunState(node.id)
        await this.executeNode(node, { preventFollowEdges: true })
      }
      if (!this.stopped) {
        this.resetNodeRunState(nodeId)
        await this.executeNode(target, { preventFollowEdges: true })
      }
      this.store.addExecutionLog({
        type: this.stopped ? 'warning' : 'success',
        message: this.stopped ? '⏹️ 单节点测试已停止' : '✅ 单节点测试完成',
        timestamp: new Date().toISOString()
      })
    } finally {
      this.testMode = false
      this.store.setExecuting(false)
    }
  }

  /**
   * 从指定节点重跑：清除该节点及其全部下游状态后从该节点继续（沿边递归）。
   * 复用的是同一实例里的上下文变量池，因此需要上一次完整执行留下的结果。
   */
  async rerunFromNode(nodeId: string): Promise<void> {
    const node = this.workflow.nodes.find(n => n.id === nodeId)
    if (!node) throw new Error(`节点不存在: ${nodeId}`)
    const ids = collectDescendantIds(this.workflow.edges, nodeId)
    ids.forEach(id => this.resetNodeRunState(id))
    this.stopped = false
    this.abortController = new AbortController()
    this.store.setExecuting(true)
    this.store.addExecutionLog({
      type: 'warning',
      message: `🔁 从节点重跑: ${node.data?.label || node.type}（已清除 ${ids.size} 个下游节点状态）`,
      timestamp: new Date().toISOString()
    })
    if (Object.keys(this.context).length === 0) {
      this.store.addExecutionLog({
        type: 'warning',
        message: '⚠️ 当前上下文为空（本次会话还没有成功执行过上游），若该节点依赖上游变量会失败；可改用「单节点测试运行」补齐上游',
        timestamp: new Date().toISOString()
      })
    }
    try {
      await this.executeNode(node)
    } finally {
      this.store.setExecuting(false)
    }
  }

  /** 供调度器使用的日志入口 */
  log(message: string, type: 'info' | 'success' | 'warning' | 'error' = 'info') {
    this.store.addExecutionLog({ type, message, timestamp: new Date().toISOString() })
  }

  // ---------------------------------------------------------------- 辅助

  captureRelevantContext(node: WNode): Record<string, any> {
    const out: Record<string, any> = {}
    if (node.type === 'aiProcess') {
      const config = node.data.config || {}
      if (config.userPrompt || config.prompt) {
        const matches = String(config.userPrompt || config.prompt).matchAll(/\{\{(\w+)\}\}/g)
        for (const m of matches) {
          const key = m[1]
          if (this.context[key] !== undefined) out[key] = this.truncateForDisplay(this.context[key])
        }
      }
    } else if (node.type === 'merge') {
      const config = node.data.config || {}
      if (Array.isArray(config.inputVariables)) {
        config.inputVariables.forEach((name: string) => {
          if (this.context[name] !== undefined) out[name] = this.truncateForDisplay(this.context[name])
        })
      }
    }
    for (const key of ['lastOutput', 'lastAIOutput', 'currentIteration', 'chapterNum']) {
      if (this.context[key] !== undefined && !out[key]) out[key] = this.truncateForDisplay(this.context[key])
    }
    return out
  }

  truncateForDisplay(value: any, limit = 200): any {
    if (typeof value === 'string') return value.length > limit ? value.substring(0, limit) + '...' : value
    if (Array.isArray(value)) {
      try {
        const s = JSON.stringify(value.slice(0, 2), null, 2)
        if (s.length <= limit) return value.length > 2 ? `${s.slice(0, -2)}...] (共${value.length}项)` : s
        return `[数组: ${value.length} 项]`
      } catch {
        return `[数组: ${value.length} 项]`
      }
    }
    if (typeof value === 'object' && value !== null) {
      try {
        const s = JSON.stringify(value, null, 2)
        return s.length <= limit ? s : s.substring(0, limit) + `... [对象: ${Object.keys(value).length} 个键]`
      } catch {
        return `[对象: ${Object.keys(value).length} 个键]`
      }
    }
    return value
  }

  getNextNodes(source: string, sourceHandle: string | null = null): WNode[] {
    let edges = this.workflow.edges.filter(e => e.source === source)
    if (sourceHandle) edges = edges.filter(e => e.sourceHandle === sourceHandle)
    return edges.map(e => {
      const node = this.workflow.nodes.find(n => n.id === e.target)
      if (!node) throw new Error(`未找到节点: ${e.target}`)
      return node
    })
  }

  getNodeExecutor(type: string, nodeId = ''): NodeExecutor | null {
    switch (type) {
      case 'start':
        return this.startExecutor
      case 'aiProcess':
        this.aiExecutor.abortSignal = this.getNodeAbort(nodeId).signal
        return this.aiExecutor
      case 'comment':
        return this.commentExecutor
      case 'transfer':
        return this.transferExecutor
      case 'fileOutput':
        return this.fileOutputExecutor
      case 'fileInput':
        return this.fileInputExecutor
      case 'textOutput':
        return this.textOutputExecutor
      case 'imageGeneration':
        this.imageGenerationExecutor.abortSignal = this.abortController?.signal
        return this.imageGenerationExecutor
      case 'imageAnalyze':
        this.imageAnalyzeExecutor.abortSignal = this.abortController?.signal
        return this.imageAnalyzeExecutor
      // ---- 星域移植节点 ----
      case 'textInput':
        return this.textInputExecutor
      case 'aiModel':
        this.aiModelExecutor.abortSignal = this.getNodeAbort(nodeId).signal
        return this.aiModelExecutor
      case 'textProcessor':
        return this.textProcessorExecutor
      case 'promptNode':
        return this.promptNodeExecutor
      case 'videoGeneration':
        this.videoGenerationExecutor.abortSignal = this.abortController?.signal
        return this.videoGenerationExecutor
      case 'imageToVideo':
        this.imageToVideoExecutor.abortSignal = this.abortController?.signal
        return this.imageToVideoExecutor
      case 'imageOutput':
        return this.imageOutputExecutor
      case 'videoOutput':
        return this.videoOutputExecutor
      case 'text':
        return this.textExecutor
      case 'merge':
        return this.mergeExecutor
      case 'loop':
        return this.loopExecutor
      case 'context':
        return this.contextExecutor
      case 'end':
        return this.endExecutor
      default:
        return null
    }
  }

  createLogger(_node: WNode): NodeLogger {
    const push = (type: any) => (message: string) =>
      this.store.addExecutionLog({ type, message: `  ${message}`, timestamp: new Date().toISOString() })
    return {
      info: push('info'),
      success: push('success'),
      warning: push('warning'),
      error: push('error')
    }
  }

  async logNodeInput(node: WNode) {
    const config = node.data?.config
    if (!config) return
    if (node.type === 'text' && config.text) {
      const text = String(config.text)
      this.store.addExecutionLog({
        type: 'info',
        message: `  📥 输入文本: ${text.substring(0, 100)}${text.length > 100 ? '...' : ''}`,
        timestamp: new Date().toISOString()
      })
    }
  }

  async logNodeOutput(_node: WNode, result: NodeExecuteResult | null | undefined) {
    if (!result || !result.outputs) return
    const entries = Object.entries(result.outputs)
    if (entries.length === 0) return
    for (const [key, value] of entries) {
      this.store.addExecutionLog({
        type: 'info',
        message: `  📤 输出 [${key}]: ${this.formatValue(value)}`,
        timestamp: new Date().toISOString()
      })
    }
  }

  formatValue(value: any): string {
    if (value === null || value === undefined) return '(空)'
    if (typeof value === 'object') {
      try {
        const s = JSON.stringify(value, null, 2)
        return s.length <= 500 ? s : s.substring(0, 500) + `... (共 ${s.length} 字符)`
      } catch {
        return `[${Array.isArray(value) ? '数组' : '对象'}: 无法序列化]`
      }
    }
    const s = String(value)
    return s.length <= 100 ? s : s.substring(0, 100) + `... (共 ${s.length} 字符)`
  }

  formatErrorMessage(e: any): string {
    if (!e) return '未知错误'
    if (typeof e === 'string') return e
    if (e instanceof Error) return e.message || e.toString()
    if (typeof e === 'object') {
      if (e.message) return e.message
      if (e.error) return this.formatErrorMessage(e.error)
      if (e.msg) return e.msg
      if (e.detail) return e.detail
      if (e.description) return e.description
      try {
        return JSON.stringify(e)
      } catch {
        return String(e)
      }
    }
    return String(e)
  }

  async logFinalOutput() {
    this.store.addExecutionLog({ type: 'info', message: '📊 最终输出:', timestamp: new Date().toISOString() })
    const keys = Object.keys(this.context)
    if (keys.length === 0) {
      this.store.addExecutionLog({ type: 'info', message: '  (无输出)', timestamp: new Date().toISOString() })
      return
    }
    for (const key of keys) {
      this.store.addExecutionLog({
        type: 'info',
        message: `  • ${key}: ${this.formatValue(this.context[key])}`,
        timestamp: new Date().toISOString()
      })
    }
  }

  updateNodeData(id: string, patch: Record<string, any>) {
    this.store.updateNodeData(id, patch)
  }

  stop() {
    this.stopped = true
    try {
      this.abortController?.abort()
    } catch {
      /* 已结束 */
    }
    // 节点级 controller 一并中止，避免残留请求
    this.nodeAborts.forEach(c => {
      try {
        c.abort()
      } catch {
        /* 已结束 */
      }
    })
    this.nodeAborts.clear()
    this.nodeStopFlags.clear()
    this.forcePauseOnce = null
    // 挂在断点上时必须放行所有排队的 Promise，否则执行器永久悬挂
    const queued = this.pauseQueue
    this.pauseQueue = []
    this.activePause = null
    queued.forEach(p => p.resolve(false))
    this.store.addExecutionLog({
      type: 'warning',
      message: '⏹️ 停止执行工作流',
      timestamp: new Date().toISOString()
    })
  }
}

export default WorkflowExecutor
