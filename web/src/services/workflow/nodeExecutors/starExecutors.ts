import { resolveConfigId } from '../modelResolve'
import { chatStream, AbortedError } from '../aiBridge'
import { isAbortLike } from './aiProcessExecutor'
import { replaceVariables } from '../template'
import { generateVideo, imageToVideo } from '../mediaAPI'
import { decryptContent } from '@/utils/promptCrypto'
import type { NodeExecuteResult, NodeLogger, WNode, WorkflowExecutorHost } from '@/types/workflow'

/**
 * 星域移植节点的执行器。
 *
 * 口径约定：**配置字段、默认值、操作语义逐条对齐星域 workflow.js 的原实现**
 * （getDefaultNodeProperties / executeXxxNode），不改名、不合并、不替换。
 *
 * 唯一保留的适配是「取值来源」：星域用 `nodeResults + getNodeInputs()` 从上游连接直接取结果对象，
 * 星芒是 context 变量池模型，因此每个节点用 sourceVariable / inputVariable 指定从哪个变量取值。
 * 这类字段在 nodeCatalog 里已标注「星芒适配」。
 */

/** 上下文里取到的值转文本（对象走 JSON） */
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
 * 星域·文本输入节点（text-input / novel-prompt 共用同一个执行函数）。
 *
 * 星域原实现返回 `{ type:'text', content: placeholder, originalText: text, placeholder }` ——
 * 输出的是**占位符名**，真实文本由下游 AI 节点通过 placeholderMap 换回。
 * 星芒等价实现：把文本同时挂到占位符名这个变量上，`{{占位符}}` 直接可替换。
 */
export class TextInputExecutor {
  async execute(node: WNode, _context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const text = String(cfg.text ?? '')
    const placeholder = String(cfg.placeholder || '').trim()
    const selected: string[] = Array.isArray(cfg.selectedPrompts) ? cfg.selectedPrompts.map((x: any) => String(x)) : []
    // 星域：提示词列表里勾选的条目一并带出
    const content = [text, ...selected].filter(s => s && s.trim()).join('\n\n')
    const key = !placeholder || placeholder === '未设占位符' ? 'textInput' : placeholder

    const outputs: Record<string, any> = {
      [key]: content,
      originalText: content,
      placeholder: key,
      textInput: key,
      lastOutput: content
    }
    logger.success(`✅ 文本输入就绪（占位符 {{${key}}}，${content.length} 字符）`)
    return { outputs, data: content }
  }
}

/**
 * 星域·AI 模型节点。
 *
 * 星域原字段：model / prompt / encryptedPrompt / isEncrypted / temperature / maxTokens。
 * - 没有 systemPrompt 字段：星域的完整提示词是「提示词 + 上游输入」一条 user 消息
 * - 没有 promptIds 字段：提示词库由「提示词节点」承担，靠连线拼进上游输入
 * - 加密模式：明文提示词加密后存 encryptedPrompt，执行时解密使用
 */
export class AiModelExecutor {
  host: WorkflowExecutorHost
  abortSignal?: AbortSignal

  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}

    // ---- 模型：星域存模型名（model），星芒按模型名反查 configId ----
    if (!cfg.model && !cfg.configId) throw new Error('❌ 未设置 AI 模型！请在节点属性中选择模型')
    const configId = await resolveConfigId({
      configId: cfg.configId,
      modelConfig: cfg.model ? ({ modelName: String(cfg.model) } as any) : undefined
    } as any)

    // ---- 提示词：加密模式走密文解密 ----
    let prompt = String(cfg.prompt || '')
    if (cfg.isEncrypted) {
      const cipher = String(cfg.encryptedPrompt || '').trim()
      if (!cipher) throw new Error('❌ 加密模式下没有密文提示词')
      const password = String(cfg.promptPassword || '')
      if (!password) throw new Error('❌ 加密提示词需要口令才能解密：请在节点属性中填写加密口令')
      const plain = await decryptContent(cipher, password)
      if (plain === null) throw new Error('❌ 提示词解密失败（口令错误或密文损坏）')
      prompt = plain
      logger.info('🔓 已解密提示词')
    }
    if (!prompt.trim()) throw new Error('❌ 未设置提示词！')

    // ---- 完整提示词 = 提示词 + 上游输入（星域：fullPrompt = prompt + inputText） ----
    const inputVar = String(cfg.inputVariable || 'lastOutput')
    const inputText = pickText(context[inputVar])
    let fullPrompt = replaceVariables(prompt, context)
    if (inputText.trim()) fullPrompt = `${fullPrompt}\n${inputText}`
    logger.info(`📋 完整提示词（${fullPrompt.length} 字符）：${fullPrompt.slice(0, 120)}`)

    const messages = [{ role: 'user' as const, content: fullPrompt }]

    // ---- 内嵌流式预览（星域 AI 节点自带）----
    let streaming = false
    let lastRendered = ''
    let lastSync = 0
    const key = node.id
    const renderStream = (full: string) => {
      const text = `🤖 AI 生成内容:\n${full}`
      if (!streaming) {
        this.host.store.beginStreamLog(key)
        streaming = true
        this.host.updateNodeData(node.id, { streamText: full })
      }
      this.host.store.updateStreamLog(text, key)
      const now = Date.now()
      if (now - lastSync > 200) {
        lastSync = now
        this.host.updateNodeData(node.id, { streamText: full })
      }
      lastRendered = text
    }

    try {
      const result = await chatStream({
        configId,
        messages,
        temperature: Number(cfg.temperature ?? 0.7),
        maxTokens: cfg.maxTokens ? Number(cfg.maxTokens) : undefined,
        signal: this.abortSignal,
        onDelta: renderStream
      })
      if (streaming) {
        this.host.store.updateStreamLog(`${lastRendered}\n\n✅ 生成完成 (共 ${result.content.length} 字符)`, key)
        this.host.store.endStreamLog(key)
      }
      this.host.updateNodeData(node.id, { streamText: result.content, result: result.content })
      logger.success(`✅ AI 生成完成（${result.content.length} 字符）`)

      const outVar = String(cfg.outputVariable || '').trim() || 'aiOutput'
      const outputs: Record<string, any> = {
        aiResult: result.content,
        lastAIOutput: result.content,
        lastOutput: result.content
      }
      if (outVar !== 'aiResult') outputs[outVar] = result.content
      return { outputs, data: result.content }
    } catch (e: any) {
      if (streaming) this.host.store.endStreamLog(key)
      const msg = e?.message || String(e)
      if (isAbortLike(msg) || e instanceof AbortedError) {
        logger.warning('⏹️ AI 生成已被停止（用户中止）')
        return { outputs: {}, data: null, aborted: true }
      }
      logger.error(`❌ AI 生成失败: ${msg}`)
      throw new Error(`AI 模型节点失败: ${msg}`)
    }
  }
}

/**
 * 星域·文本处理节点。
 *
 * 星域原实现只落地了 format / uppercase / lowercase 三种操作，
 * UI 上另外 4 项（clean / extract / replace / split / merge / transform）走 default 分支原样输出。
 * 此处保持同样行为，并对未实现的操作给出明确日志，不擅自补逻辑。
 */
export class TextProcessorExecutor {
  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const operation = String(cfg.operation || 'format')
    const source = String(cfg.sourceVariable || 'lastOutput')
    const content = pickText(context[source])

    logger.info(`🛠️ 文本处理：${operation}`)
    let out = content
    switch (operation) {
      case 'format':
        out = content.trim().replace(/\n\s*\n/g, '\n\n')
        break
      case 'uppercase':
        out = content.toUpperCase()
        break
      case 'lowercase':
        out = content.toLowerCase()
        break
      default:
        logger.warning(`⚠️ 操作 "${operation}" 在星域原版中未实现处理逻辑，按原文输出`)
        break
    }

    const outputs: Record<string, any> = { textProcessed: out, lastOutput: out }
    logger.success(`✅ 文本处理完成（${operation}，${content.length} → ${out.length} 字符）`)
    return { outputs, data: out }
  }
}

/**
 * 星域·提示词节点（novel-prompt）。
 *
 * 与文本输入节点共用执行逻辑（星域原版就是 `case 'novel-prompt': return executeTextInputNode(node)`），
 * 区别只在默认属性：novel-prompt 的 promptList 预置全局提示词。
 */
export class PromptNodeExecutor extends TextInputExecutor {}

/**
 * 星域·文生视频节点。
 * 字段：prompt / model / size / duration / fps / quality / with_audio / style / apiProvider
 * 完整提示词 = prompt + 上游输入 + style（星域 L6528：`${fullPrompt}, ${style}`）
 */
export class VideoGenerationExecutor {
  host: WorkflowExecutorHost
  abortSignal?: AbortSignal

  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    let prompt = replaceVariables(String(cfg.prompt || ''), context)
    const inputVar = String(cfg.inputVariable || 'lastOutput')
    const upstream = pickText(context[inputVar])
    if (upstream.trim()) prompt = `${prompt}\n${upstream}`
    const style = String(cfg.style || '').trim()
    if (style) prompt = `${prompt}, ${style}`
    prompt = prompt.trim()
    if (!prompt) throw new Error('❌ 未设置视频提示词')

    ;(node.data.config as any).analysisStatus = '正在生成视频...'
    logger.info(`🎬 提交文生视频任务：${prompt.slice(0, 80)}`)
    logger.info(`⚙️ 模型 ${cfg.model || '未设置'} · ${cfg.size || '1920x1080'} · ${cfg.duration || 5}s · ${cfg.fps || 30}fps`)
    const res = await generateVideo({
      configId: cfg.configId || undefined,
      model: cfg.model || undefined,
      prompt,
      duration: Number(cfg.duration || 5),
      fps: Number(cfg.fps || 30),
      size: cfg.size,
      quality: cfg.quality,
      withAudio: cfg.with_audio,
      signal: this.abortSignal
    })

    const url = res?.url || ''
    this.host.updateNodeData(node.id, { videoUrl: url, result: url })
    ;(node.data.config as any).analysisStatus = '生成完成'
    const outputs: Record<string, any> = { videoUrl: url, videos: url ? [url] : [], lastOutput: url }
    logger.success('✅ 视频生成完成')
    return { outputs, data: url }
  }

}

/**
 * 星域·图生视频节点。
 * 字段：prompt / model / width / height / duration / fps / style / apiProvider
 * 图片来源：星域从上游图片输出节点取（星芒用 sourceVariable 指定变量）
 */
export class ImageToVideoExecutor {
  host: WorkflowExecutorHost
  abortSignal?: AbortSignal

  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const sourceVar = String(cfg.sourceVariable || 'imageB64')
    // 星域：属性面板「输入图片」导入的图片优先，否则取上游变量
    let rawText = String(cfg.inputImage || '').trim()
    if (!rawText) rawText = pickText(context[sourceVar])
    if (!rawText.trim()) {
      throw new Error(`❌ 没有可用的图片：既未导入图片，上游变量 "${sourceVar}" 也为空`)
    }
    const isUrl = rawText.startsWith('http') || rawText.startsWith('data:')
    let prompt = replaceVariables(String(cfg.prompt || ''), context)
    const style = String(cfg.style || '').trim()
    if (style) prompt = `${prompt}, ${style}`

    ;(node.data.config as any).analysisStatus = '正在生成视频...'
    logger.info(`🎞️ 提交图生视频任务（图片源：${cfg.inputImage ? '节点内导入' : sourceVar}）`)
    const res = await imageToVideo({
      configId: cfg.configId || undefined,
      model: cfg.model || undefined,
      imageBase64: isUrl ? undefined : rawText,
      imageUrl: isUrl ? rawText : undefined,
      prompt: prompt.trim() || undefined,
      duration: Number(cfg.duration || 5),
      fps: Number(cfg.fps || 30),
      signal: this.abortSignal
    })

    const url = res?.url || ''
    this.host.updateNodeData(node.id, { videoUrl: url, result: url })
    ;(node.data.config as any).analysisStatus = '生成完成'
    const outputs: Record<string, any> = { videoUrl: url, videos: url ? [url] : [], lastOutput: url }
    logger.success('✅ 图生视频完成')
    return { outputs, data: url }
  }
}

/**
 * 星域·图片输出节点。
 * 字段：filename / format（png / jpg / webp）。
 * 多张图片时按星域规则命名 `${filename}_${i+1}_${j+1}`。
 */
export class ImageOutputExecutor {
  host: WorkflowExecutorHost
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const filename = String(cfg.filename || 'generated_image').replace(/[/\\:*?"<>|]/g, '_')
    const format = String(cfg.format || 'png')
    const source = String(cfg.sourceVariable || 'images')
    const raw = context[source]

    // 兼容：images 数组 / imageB64 / 图片 URL
    let list: Array<{ b64?: string; url?: string }> = []
    if (Array.isArray(raw)) list = raw.filter(Boolean)
    else if (typeof raw === 'string' && raw.trim()) list = [{ url: raw.startsWith('http') || raw.startsWith('data:') ? raw : undefined, b64: raw.startsWith('http') || raw.startsWith('data:') ? undefined : raw }]
    if (!list.length) throw new Error('图片输出节点没有接收到图片数据')

    const saved: string[] = []
    for (let i = 0; i < list.length; i++) {
      const item = list[i]
      const multi = list.length > 1
      const finalName = multi ? `${filename}_${i + 1}_1.${format}` : `${filename}.${format}`
      if (item.b64) {
        const relPath = `nodes/${finalName}`
        if (this.host.writeFile) await this.host.writeFile(relPath, item.b64, 'node', false, 'base64')
        saved.push(relPath)
      } else {
        logger.warning(`⚠️ 第 ${i + 1} 张图片只有 URL（${item.url}），未落盘`)
        saved.push(item.url || '')
      }
    }
    this.host.updateNodeData(node.id, { images: list, result: `已保存 ${saved.length} 张` })
    ;(node.data.config as any).savedImages = saved
    logger.success(`✅ 图片输出完成（${saved.length} 张，格式 ${format}）`)
    return { outputs: { files: saved, lastOutput: saved[0] || '' }, data: saved.join('\n') }
  }
}

/**
 * 星域·视频输出节点。
 * 字段：filename / format（mp4 / avi / mov）。
 */
export class VideoOutputExecutor {
  host: WorkflowExecutorHost
  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const cfg = node.data?.config || {}
    const filename = String(cfg.filename || 'generated_video').replace(/[/\\:*?"<>|]/g, '_')
    const format = String(cfg.format || 'mp4')
    const source = String(cfg.sourceVariable || 'videoUrl')
    const raw = context[source]

    const list: string[] = Array.isArray(raw) ? raw.map((x: any) => String(x)) : raw ? [String(raw)] : []
    if (!list.length) throw new Error('视频输出节点没有接收到视频数据')

    const saved: string[] = []
    for (let i = 0; i < list.length; i++) {
      const data = list[i]
      const finalName = list.length > 1 ? `${filename}_${i + 1}_1.${format}` : `${filename}.${format}`
      // base64 落盘；仅 URL 时如实提示
      if (!data.startsWith('http')) {
        const relPath = `nodes/${finalName}`
        if (this.host.writeFile) await this.host.writeFile(relPath, data, 'node', false, 'base64')
        saved.push(relPath)
      } else {
        logger.warning(`⚠️ 视频为外链（${data}），未落盘`)
        saved.push(data)
      }
    }
    this.host.updateNodeData(node.id, { videoUrl: list[0], result: `已保存 ${saved.length} 个视频` })
    ;(node.data.config as any).savedVideos = saved
    logger.success(`✅ 视频输出完成（${saved.length} 个，格式 ${format}）`)
    return { outputs: { files: saved, videoUrl: list[0], lastOutput: list[0] }, data: saved.join('\n') }
  }
}
