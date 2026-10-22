import { promptAPI } from '@/api'
import type { ChatMessage } from '@/types'
import type { NodeExecuteResult, NodeLogger, WNode } from '@/types/workflow'
import { compileTemplate, replaceVariables } from '../template'
import { resolveConfigId } from '../modelResolve'
import { AbortedError, chatStream } from '../aiBridge'

/**
 * AI 处理节点 —— 直译源项目 aiProcessExecutor。
 * 差异（本地化）：模型解析走 configId（星芒 /api/ai/chat 的入参），
 * 系统提示词支持直接文本 + 星芒提示词库 id（替代源版 AES 加密提示词）。
 */
export class AiProcessExecutor {
  store: any

  constructor(store: any) {
    this.store = store
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger, record?: any): Promise<NodeExecuteResult> {
    const config = node.data?.config || {}

    if (!config.configId && !config.modelId) {
      throw new Error('❌ 未设置 AI 模型！请在节点属性中选择模型。')
    }
    let userPrompt: string = config.userPrompt || config.prompt || ''
    if (!userPrompt) {
      throw new Error('❌ 未设置用户提示词！请在节点属性中配置提示词。')
    }

    // ---- 系统提示词 ----
    let systemPrompt = typeof config.systemPrompt === 'string' ? config.systemPrompt : ''
    if (!systemPrompt && config.systemPromptId) {
      try {
        const loaded = await this.loadSystemPrompt(config.systemPromptId)
        if (loaded) {
          systemPrompt = loaded
          logger.info(`📚 已从提示词库载入系统提示词 #${config.systemPromptId}`)
        } else {
          logger.warning(`⚠️ 未找到系统提示词 ID: ${config.systemPromptId}`)
        }
      } catch (e: any) {
        logger.warning(`⚠️ 加载系统提示词失败: ${e?.message || e}`)
      }
    }

    // ---- 上下文变量拼成 system 消息 ----
    const contextMessages: ChatMessage[] = []
    try {
      let names: any = config.contextVariables
      if (Array.isArray(names) && names.length > 0 && typeof names[0] === 'object' && names[0]?.variableName) {
        names = names.map((c: any) => c.variableName).filter(Boolean)
        logger.info(`✅ 使用新格式上下文变量: [${names.join(', ')}]`)
      } else if (typeof names === 'string') {
        names = names.split(',').map((c: string) => c.trim()).filter(Boolean)
        logger.info(`⚠️ 使用旧格式上下文变量: [${names.join(', ')}]`)
      }

      const maxItems = Number(config.contextMaxItems || 3)
      const toText = (v: any): string => {
        if (v === null || v === undefined) return ''
        if (typeof v === 'string') return v
        try {
          return JSON.stringify(v, null, 2)
        } catch {
          return String(v)
        }
      }

      if (Array.isArray(names) && names.length > 0) {
        for (const name of names as string[]) {
          const value = context[name]
          if (value === undefined) {
            logger.warning(`⚠️ 上下文变量 "${name}" 在当前上下文中不存在`)
            continue
          }
          if (Array.isArray(value)) {
            const tail = value.slice(-maxItems)
            logger.info(`📚 加载数组变量 "${name}": 最近 ${tail.length} 条`)
            tail.forEach((item, i) => {
              contextMessages.push({ role: 'system', content: `【${name}#${i + 1}】\n${toText(item)}` })
            })
          } else {
            logger.info(`📄 加载变量 "${name}": ${typeof value === 'string' ? value.substring(0, 50) + '...' : typeof value}`)
            contextMessages.push({ role: 'system', content: `【${name}】\n${toText(value)}` })
          }
        }
      } else {
        logger.info('ℹ️ 未配置上下文变量')
      }
    } catch (e: any) {
      logger.warning(`⚠️ 构建上下文消息失败: ${e?.message || e}`)
    }

    // ---- 模板替换 ----
    if (systemPrompt) systemPrompt = compileTemplate(systemPrompt, context)
    const finalUserPrompt = replaceVariables(userPrompt, context)
    logger.info(`📋 系统提示词: ${systemPrompt ? systemPrompt.substring(0, 100) + (systemPrompt.length > 100 ? '...' : '') : '(无)'}`)
    logger.info(`📋 用户提示词: ${finalUserPrompt.length > 100 ? finalUserPrompt.substring(0, 100) + '...' : finalUserPrompt}`)

    // ---- 模型 ----
    const configId = await resolveConfigId(config)
    const modelLabel = config?.modelConfig?.modelName || config.modelId || `#${configId}`
    logger.info(`🤖 使用模型: ${config?.modelConfig?.providerName ? config.modelConfig.providerName + ' / ' : ''}${modelLabel}`)

    const messages: ChatMessage[] = []
    if (systemPrompt) messages.push({ role: 'system', content: systemPrompt })
    if (contextMessages.length > 0) messages.push(...contextMessages)
    messages.push({ role: 'user', content: finalUserPrompt })

    const maxTokens = Number(config.maxTokens || 2000)

    let streaming = false
    let lastRendered = ''
    const renderStreamLog = (content: string) => {
      const text = `🤖 AI 生成内容:\n${content}`
      if (!streaming) {
        this.store.beginStreamLog(this.streamKey)
        streaming = true
      }
      this.store.updateStreamLog(text, this.streamKey)
      lastRendered = text
    }

    try {
      const result = await chatStream({
        configId,
        messages,
        temperature: Number(config.temperature || 0.7),
        maxTokens,
        signal: this.abortSignal,
        onDelta: (full: string) => renderStreamLog(full)
      })

      if (streaming) {
        this.store.updateStreamLog(`${lastRendered}\n\n✅ 生成完成 (共 ${result.content.length} 字符)`, this.streamKey)
        this.store.endStreamLog(this.streamKey)
      }
      logger.success(`✅ AI 生成完成 (共 ${result.content.length} 字符)`)
      this.store.updateNodeData(node.id, { result: result.content })

      if (record) record.progress = 100
      return this.buildResult(config, result.content)
    } catch (e: any) {
      if (streaming) {
        this.store.endStreamLog(this.streamKey)
        const list = this.store.executionLogs || []
        const idx = list.findIndex((l: any) => l.isStreaming || (l.stream && l.message === lastRendered))
        if (idx !== -1) {
          list[idx].type = 'warning'
          list[idx].level = 'warning'
          list[idx].isStreaming = false
        }
      }
      const msg = this.formatErrorMessage(e)
      if (isAbortLike(msg) || e instanceof AbortedError) {
        logger.warning('⏹️ AI 生成已被停止（用户中止）')
        return { outputs: {}, data: null, aborted: true }
      }
      logger.error(`❌ AI 生成失败: ${msg}`)
      throw new Error(`AI 处理失败: ${msg}`)
    }
  }

  /** 由主执行器在执行前注入（供停止按钮联动） */
  abortSignal?: AbortSignal

  /** 流式日志分桶 key：并发模式下按节点 id 区分，避免多个 AI 节点互相覆盖 */
  streamKey = 'default'

  private buildResult(config: any, content: string): NodeExecuteResult {
    const outputs: Record<string, unknown> = { aiResult: content, lastAIOutput: content }
    const outputVar = typeof config.outputVariable === 'string' ? config.outputVariable.trim() : ''
    if (outputVar && outputVar !== 'aiResult') outputs[outputVar] = content
    return { outputs, data: content }
  }

  /** 星芒提示词库取系统提示词（替代源版 electronAPI.store 的 promptLibrary） */
  private async loadSystemPrompt(id: number | string): Promise<string> {
    const numId = Number(id)
    if (!numId || Number.isNaN(numId)) return ''
    try {
      const res: any = await promptAPI.getOne(numId)
      return res?.data?.content || ''
    } catch (e) {
      console.error('加载系统提示词失败:', e)
      return ''
    }
  }

  private formatErrorMessage(e: any): string {
    if (!e) return '未知错误'
    if (typeof e === 'string') return e
    if (e instanceof Error) return e.message || e.toString()
    if (typeof e === 'object') {
      return e.message || e.error || e.msg || e.detail || e.description || JSON.stringify(e)
    }
    return String(e)
  }
}

export function isAbortLike(msg: string): boolean {
  const s = String(msg || '').toLowerCase()
  return (
    s.includes('已停止') ||
    s.includes('stop') ||
    s.includes('abort') ||
    s.includes('canceled') ||
    s.includes('cancelled')
  )
}

export default AiProcessExecutor
