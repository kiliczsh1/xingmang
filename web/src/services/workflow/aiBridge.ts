import type { ChatMessage } from '@/types'

/**
 * SSE 适配器：把源项目的 window.electronAPI.ai.chatStream 换成本机后端的
 * POST /api/ai/chat（text/event-stream）。
 *
 * 后端事件格式（见 internal/handlers/ai.go）：
 *   data: {"content":"…"}      增量
 *   data: {"error":"…"}        错误
 *   data: [DONE]               结束
 *
 * 本地化：后端按 configId（api_models.id）选模型；无 configId 时回落到默认模型。
 */

export interface ChatStreamOptions {
  configId?: number
  messages: ChatMessage[]
  temperature?: number
  maxTokens?: number
  onDelta?: (text: string) => void
  signal?: AbortSignal
}

export interface ChatStreamResult {
  content: string
}

/** 被用户中止时抛出的错误，执行器据此标记 cancelled */
export class AbortedError extends Error {
  constructor(message = '工作流已停止') {
    super(message)
    this.name = 'AbortedError'
  }
}

export async function chatStream(opts: ChatStreamOptions): Promise<ChatStreamResult> {
  const body: Record<string, unknown> = { messages: opts.messages }
  if (opts.configId && opts.configId > 0) body.configId = opts.configId
  if (opts.maxTokens && opts.maxTokens > 0) body.maxTokens = Math.floor(opts.maxTokens)
  // temperature 由后端从模型配置读取，这里仅作提示（/api/ai/chat 当前不接收该字段）

  let resp: Response
  try {
    resp = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: opts.signal
    })
  } catch (e: any) {
    if (e?.name === 'AbortError') throw new AbortedError()
    throw new Error(`AI 请求失败: ${e?.message || e}`)
  }

  if (!resp.ok) {
    let msg = `HTTP ${resp.status}`
    try {
      const j = await resp.json()
      msg = j?.message || j?.error || msg
    } catch {
      /* 忽略解析失败 */
    }
    throw new Error(msg)
  }
  if (!resp.body) throw new Error('AI 响应没有流式内容')

  const reader = resp.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''
  let content = ''
  let errorMessage = ''

  try {
    // eslint-disable-next-line no-constant-condition
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      buffer += decoder.decode(value, { stream: true })

      let idx: number
      while ((idx = buffer.indexOf('\n')) !== -1) {
        const line = buffer.slice(0, idx).trim()
        buffer = buffer.slice(idx + 1)
        if (!line.startsWith('data:')) continue
        const payload = line.slice(5).trim()
        if (!payload) continue
        if (payload === '[DONE]') {
          if (errorMessage) throw new Error(errorMessage)
          return { content }
        }
        let parsed: any
        try {
          parsed = JSON.parse(payload)
        } catch {
          continue
        }
        if (parsed?.error) {
          errorMessage = String(parsed.error)
          continue
        }
        if (typeof parsed?.content === 'string' && parsed.content) {
          content += parsed.content
          opts.onDelta?.(content)
        }
      }
    }
  } catch (e: any) {
    if (e?.name === 'AbortError') throw new AbortedError()
    if (e instanceof AbortedError) throw e
    throw e
  } finally {
    try {
      reader.releaseLock()
    } catch {
      /* 已释放 */
    }
  }

  if (errorMessage) throw new Error(errorMessage)
  return { content }
}

export default chatStream
