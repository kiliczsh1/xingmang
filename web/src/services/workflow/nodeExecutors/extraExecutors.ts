import { replaceVariables } from '../template'
import { analyzeImage, generateImage } from '../mediaAPI'
import { resolveConfigId } from '../modelResolve'
import type { NodeExecuteResult, NodeLogger, WNode, WorkflowExecutorHost } from '@/types/workflow'

/**
 * 星域移植节点里需要宿主能力（落盘 / 回写节点 data）的执行器。
 *
 * 口径约定同 starExecutors.ts：**字段名、默认值、行为逐条对齐星域原实现**，
 * 仅「取值来源」按星芒的 context 变量模型适配（字段已在 nodeCatalog 标注）。
 */

const pickText = (v: any): string => {
  if (v === undefined || v === null) return ''
  if (typeof v === 'string') return v
  try {
    return JSON.stringify(v, null, 2)
  } catch {
    return String(v)
  }
}

/**
 * 循环体内的迭代序号（星域 loopIndex，0-based）。
 * 星芒由 LoopExecutor 写入 context.currentIteration（1-based），非循环场景为 0/undefined → 返回 null。
 */
const loopIndexOf = (context: Record<string, any>): number | null => {
  const n = Number(context?.currentIteration)
  return Number.isFinite(n) && n > 0 ? n - 1 : null
}

/** 序号后缀（星域：loopIndex 存在时为 `_${loopIndex+1}`） */
/** 已有的分析文本（重跑时先展示上一轮结果） */
const priorAnalysisText = (cfg: Record<string, any>): string => String(cfg?.analysisResult?.content || '')

const withIndex = (filename: string, index: number | null | undefined): string => {
  if (typeof index !== 'number' || index < 0) return filename
  const dot = filename.lastIndexOf('.')
  return dot > 0 ? `${filename.slice(0, dot)}_${index + 1}${filename.slice(dot)}` : `${filename}_${index + 1}`
}

/** 注释：不参与流程，仅回带文本（星域原实现返回 {success, output: text, type:'comment'}） */
export class CommentExecutor {
  async execute(node: WNode): Promise<NodeExecuteResult> {
    const text = String(node.data?.config?.text || '')
    return { outputs: { comment: text }, data: text }
  }
}

/**
 * 星域·中转节点。
 * 字段：receivedData / placeholder / autoForward / displayMode（text | json）。
 * 星域把下游要取的数据挂在**占位符名**上，此处同时挂到该变量名以支持 `{{占位符}}`。
 */
export class TransferExecutor {
  host: WorkflowExecutorHost
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const placeholder = String(cfg.placeholder || '').trim()
    const inputVar = String(cfg.inputVariable || 'lastOutput')
    let received = pickText(context[inputVar])

    // 星域：无上游数据时用占位符兜底
    if (!received.trim()) {
      received = cfg.placeholder || '等待上游数据...'
      logger.warning('⚠️ 上游无数据，使用占位符值')
    }

    let outputData = received
    if (cfg.displayMode === 'json') {
      try {
        outputData = JSON.stringify(JSON.parse(received), null, 2)
      } catch {
        outputData = JSON.stringify({ data: received }, null, 2)
      }
    }

    const outputs: Record<string, any> = { transferred: outputData, receivedData: outputData }
    if (placeholder && placeholder !== '未设占位符') outputs[placeholder] = outputData
    // autoForward：关掉时不把数据继续往下传
    if (cfg.autoForward !== false) outputs.lastOutput = outputData

    // 星域：接收到的数据写回节点属性，供属性面板「接收的数据」直接查看
    ;(node.data.config as any).receivedData = outputData
    this.host.updateNodeData(node.id, { result: outputData })
    logger.success(`✅ 中转完成（${outputData.length} 字符，${cfg.displayMode || 'text'} 模式）`)
    return { outputs, data: outputData }
  }
}

/**
 * 星域·文件输出节点。
 * 字段：path / filename。
 * 星域把 path 当输出目录（默认 ./exports/）；星芒的落盘被限制在本次执行目录内，
 * 因此把 path 的目录部分映射为执行目录下的子目录。
 */
export class FileOutputExecutor {
  host: WorkflowExecutorHost
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const source = String(cfg.sourceVariable || 'lastOutput')
    const content = pickText(context[source])
    if (!content.trim()) {
      logger.warning(`⚠️ 变量 "${source}" 为空，未写出文件`)
      return { outputs: {}, data: '' }
    }

    const dir = String(cfg.path || './exports/').replace(/\\/g, '/').replace(/^\.\/+/, '').replace(/^\/+/, '').replace(/\.\.+/g, '')
    const filename = withIndex(String(cfg.filename || 'output.txt').replace(/[/\\:*?"<>|]/g, '_'), loopIndexOf(context))
    const relPath = `${dir ? dir.replace(/\/?$/, '/') : ''}${filename}`

    if (!this.host.runId || !this.host.writeFile) {
      logger.warning('⚠️ 本次执行没有落盘目录（工作流可能尚未保存），文件未写出')
      return { outputs: {}, data: '（未落盘）' }
    }
    await this.host.writeFile(relPath, content, 'node')
    logger.success(`💾 已写出 ${relPath}（${content.length} 字符）`)
    this.host.updateNodeData(node.id, { result: content, outputFile: relPath })
    return { outputs: { outputFile: relPath, fileContent: content, lastOutput: content }, data: relPath }
  }
}

/**
 * 星域·文本输出节点。
 * 字段：format（text | markdown | html）/ filename。
 * 星域原实现按格式加工内容（markdown 加 `# 工作流输出` 标题、html 包成文档），不做 Markdown 渲染。
 */
export class TextOutputExecutor {
  host: WorkflowExecutorHost
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const format = String(cfg.format || 'text')
    const source = String(cfg.sourceVariable || 'lastOutput')
    let content = pickText(context[source])
    if (!content.trim()) {
      logger.warning(`⚠️ 变量 "${source}" 为空，未写出文件`)
      return { outputs: {}, data: '' }
    }

    switch (format) {
      case 'markdown':
        content = `# 工作流输出\n\n${content}`
        break
      case 'html':
        content = `<!DOCTYPE html>\n<html>\n<head><title>工作流输出</title></head>\n<body>\n<pre>${content}</pre>\n</body>\n</html>`
        break
      default:
        break
    }

    const filename = withIndex(
      String(cfg.filename || 'output.txt').replace(/[/\\:*?"<>|]/g, '_'),
      loopIndexOf(context)
    )
    if (!this.host.runId || !this.host.writeFile) {
      logger.warning('⚠️ 本次执行没有落盘目录（工作流可能尚未保存），文件未写出')
      this.host.updateNodeData(node.id, { result: content })
      return { outputs: {}, data: '（未落盘）' }
    }
    const relPath = `nodes/${filename}`
    await this.host.writeFile(relPath, content, 'node')
    logger.success(`💾 已写出 ${relPath}（${content.length} 字符，格式 ${format}）`)
    ;(node.data.config as any).outputContent = content
    this.host.updateNodeData(node.id, { result: content, outputFile: relPath, preview: content.slice(0, 2000) })
    return { outputs: { outputFile: relPath, fileContent: content, textOutput: content, lastOutput: content }, data: relPath }
  }
}

/**
 * 星域·文件输入节点。
 * 字段：fileType（text/json/csv/xml/markdown）/ path / encoding（utf-8/gbk/ascii）/ readMode（full/lines/chunks）。
 * 星域优先使用「导入的文件内容」（前端选文件读进 fileContent），否则走 /api/read-file。
 */
export class FileInputExecutor {
  async execute(node: WNode, _context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const fileType = String(cfg.fileType || 'text')
    const encoding = String(cfg.encoding || 'utf-8')
    const readMode = String(cfg.readMode || 'full')

    // 星域：优先用导入的文件内容
    if (cfg.fileContent) {
      const content = String(cfg.fileContent)
      logger.success(`📄 使用导入的文件内容（${cfg.fileName || '未命名'}，${content.length} 字符）`)
      return this.buildResult(cfg, content, fileType)
    }

    const path = String(cfg.path || '').trim()
    if (!path) throw new Error('❌ 未指定文件路径或导入文件')

    logger.info(`📄 读取文件：${path}（${fileType} / ${encoding} / ${readMode}）`)
    let resp: Response
    try {
      resp = await fetch('/api/files/read-text', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, file_type: fileType, encoding, read_mode: readMode })
      })
    } catch (e: any) {
      throw new Error(`读取文件请求失败: ${e?.message || e}`)
    }
    if (!resp.ok) {
      let msg = `HTTP ${resp.status}`
      try {
        const j = await resp.json()
        msg = j?.message || msg
      } catch {
        /* 忽略 */
      }
      throw new Error(`读取文件失败: ${msg}`)
    }
    const json: any = await resp.json().catch(() => null)
    const data = json?.data || {}
    const content = String(data.content ?? '')
    if (data.truncated) logger.warning(`⚠️ 文件较大（${data.size} 字节），只读取了前 ${content.length} 字符`)
    if (data.encoding && String(data.encoding).toLowerCase() !== encoding) {
      logger.info(`ℹ️ 实际识别编码：${data.encoding}`)
    }
    logger.success(`✅ 文件读取成功（${content.length} 字符）`)
    return this.buildResult(cfg, content, fileType)
  }

  private buildResult(cfg: Record<string, any>, content: string, fileType: string): NodeExecuteResult {
    const outVar = String(cfg.outputVariable || '').trim() || 'fileContent'
    const outputs: Record<string, any> = { fileContent: content, fileType, lastOutput: content }
    if (outVar !== 'fileContent') outputs[outVar] = content
    return { outputs, data: content }
  }
}

/**
 * 星域·文生图节点。
 * 字段：prompt / model / width / height / num_images / style / apiProvider。
 * 完整提示词 = prompt + 上游输入 + style（星域 L6528）。
 */
export class ImageGenerationExecutor {
  host: WorkflowExecutorHost
  abortSignal?: AbortSignal
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const inputVar = String(cfg.inputVariable || 'lastOutput')
    const upstream = pickText(context[inputVar])
    let prompt = replaceVariables(String(cfg.prompt || ''), context)
    if (upstream.trim()) prompt = `${prompt}\n${upstream}`
    const style = String(cfg.style || '').trim()
    if (style) prompt = `${prompt}, ${style}`
    prompt = prompt.trim()
    if (!prompt) throw new Error('❌ 未设置图片描述')

    const width = Number(cfg.width || 1024)
    const height = Number(cfg.height || 1024)
    const numImages = Number(cfg.num_images || 1)
    logger.info(`🎨 开始生成图片：${prompt.slice(0, 80)}`)
    logger.info(`📐 模型 ${cfg.model || '默认'} · ${width}x${height} · ${numImages} 张 · 提供方 ${cfg.apiProvider || 'custom'}`)

    const configId = await resolveConfigId({
      configId: cfg.configId,
      modelConfig: cfg.model ? ({ modelName: String(cfg.model) } as any) : undefined
    } as any).catch(() => 0)

    const images = await generateImage({
      configId: configId || undefined,
      model: cfg.model || undefined,
      prompt,
      size: `${width}x${height}`,
      count: numImages,
      signal: this.abortSignal
    })

    const first = images[0] || {}
    this.host.updateNodeData(node.id, { images, result: `已生成 ${images.length} 张` })
    logger.success(`✅ 图片生成完成（${images.length} 张）`)
    return {
      outputs: {
        images,
        imageB64: first.b64 || '',
        imageUrl: first.url || '',
        lastOutput: first.b64 || first.url || ''
      },
      data: `已生成 ${images.length} 张图片（${width}x${height}）`
    }
  }
}

/**
 * 星域·图片分析节点。
 * 字段：prompt / model / inputImage / inputImageName / apiProvider。
 * inputImage 为星域前端导入的图片（dataURL），星芒侧用 sourceVariable 指定上游图片变量。
 */
export class ImageAnalyzeExecutor {
  host: WorkflowExecutorHost
  abortSignal?: AbortSignal
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  /** 星域：分析状态/结果写回节点属性，供面板「分析状态 / 分析结果预览」显示 */
  private setAnalysis(node: WNode, status: string, content?: string, complete?: boolean) {
    const cfg = node.data.config as any
    cfg.analysisStatus = status
    if (content !== undefined) cfg.analysisResult = { content, isComplete: complete !== false }
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const sourceVar = String(cfg.sourceVariable || 'imageB64').trim() || 'imageB64'

    // 图片来源：节点内导入的图片 > 配置的 URL > 上游变量
    let imageUrl = ''
    let imageBase64 = ''
    const imported = String(cfg.inputImage || '').trim()
    if (imported) {
      if (imported.startsWith('http') || imported.startsWith('data:')) imageUrl = imported
      else imageBase64 = imported
    }
    if (!imageUrl && !imageBase64) {
      const fromCtx = context[sourceVar]
      const text = typeof fromCtx === 'string' ? fromCtx.trim() : ''
      if (text) {
        if (text.startsWith('http') || text.startsWith('data:')) imageUrl = text
        else imageBase64 = text
      }
    }
    if (!imageUrl && !imageBase64) {
      throw new Error(`❌ 没有可分析的图片：上游变量 "${sourceVar}" 为空，且未导入图片`)
    }

    const prompt = replaceVariables(String(cfg.prompt || ''), context).trim() || '请描述这张图片的内容。'
    this.setAnalysis(node, '正在分析...', priorAnalysisText(cfg), false)
    logger.info(`🔍 开始分析图片（来源：${imageUrl ? 'URL' : 'base64'}）`)

    const configId = await resolveConfigId({
      configId: cfg.configId,
      modelConfig: cfg.model ? ({ modelName: String(cfg.model) } as any) : undefined
    } as any).catch(() => 0)

    let streaming = false
    const key = node.id
    let text = ''
    try {
      text = await analyzeImage({
        configId: configId || undefined,
        model: cfg.model || undefined,
        imageBase64: imageBase64 || undefined,
        imageUrl: imageUrl || undefined,
        prompt,
        signal: this.abortSignal,
        onDelta: full => {
          if (!streaming) {
            this.host.store.beginStreamLog(key)
            streaming = true
          }
          this.host.store.updateStreamLog(`🖼️ 图片分析:\n${full}`, key)
          this.setAnalysis(node, '正在分析...', full, false)
        }
      })
    } finally {
      if (streaming) this.host.store.endStreamLog(key)
    }

    this.host.updateNodeData(node.id, { result: text })
    this.setAnalysis(node, '分析完成', text, true)
    logger.success(`✅ 图片分析完成（${text.length} 字符）`)
    return { outputs: { imageAnalysis: text, lastAIOutput: text, lastOutput: text }, data: text }
  }
}
