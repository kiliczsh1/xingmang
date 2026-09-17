// ChatApiBridge.ts —— 通信桥接层
// 将后端 /api/ai/chat (自定义 SSE) 包装成 @ai-sdk/vue v4 的 HttpChatTransport 兼容接口

import { HttpChatTransport, type UIMessage, type UIMessageChunk } from 'ai'

/**
 * 自定义 Transport：继承 HttpChatTransport，重写：
 * 1. prepareSendMessagesRequest —— 将 AI SDK 消息格式转为后端期望的格式
 * 2. processResponseStream —— 将后端自定义 SSE 流解析为 AI SDK UIMessageChunk 流
 * 3. sendMessages —— 拦截 HTTP 错误，解析后端真实错误信息
 */
export class CustomChatTransport extends HttpChatTransport<UIMessage> {
  private _configId?: number
  private temperature?: number
  private topP?: number
  private _api: string

  /** 获取当前 configId（外部只读） */
  get configId(): number | undefined { return this._configId }

  constructor(options: {
    configId?: number
    temperature?: number
    topP?: number
  }) {
    super({
      api: '/api/ai/chat',
      body: {},
    })
    this._api = '/api/ai/chat'
    this._configId = options.configId
    this.temperature = options.temperature
    this.topP = options.topP
  }

  /** 更新配置（模型切换时调用） */
  updateConfig(options: { configId?: number; temperature?: number; topP?: number }) {
    if (options.configId !== undefined) this._configId = options.configId
    if (options.temperature !== undefined) this.temperature = options.temperature
    if (options.topP !== undefined) this.topP = options.topP
  }

  /**
   * 重写 sendMessages：拦截 HTTP 非200响应，解析后端真实错误信息
   */
  override async sendMessages(
    options: Parameters<HttpChatTransport<UIMessage>['sendMessages']>[0]
  ): Promise<ReadableStream<UIMessageChunk>> {
    // 构造请求体
    const req = this.prepareSendMessagesRequest({
      id: options.chatId,
      messages: options.messages,
      messageId: options.messageId,
      trigger: options.trigger,
    } as Parameters<NonNullable<typeof this.prepareSendMessagesRequest>>[0])

    const body = {
      ...(req?.body || {}),
      ...(options.body || {}),
    }

    // 发起 fetch
    const response = await fetch(this._api, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(options.headers || {}),
      },
      body: JSON.stringify(body),
      signal: options.abortSignal,
    })

    // ===== 关键：非200时解析后端真实错误 =====
    if (!response.ok) {
      let errorMessage = `HTTP ${response.status}`
      try {
        const errorBody = await response.json()
        if (errorBody?.message) {
          errorMessage = errorBody.message
        } else if (errorBody?.error) {
          errorMessage = typeof errorBody.error === 'string' ? errorBody.error : JSON.stringify(errorBody.error)
        }
      } catch {
        // 无法解析 JSON 时使用默认消息
      }
      throw new Error(errorMessage)
    }

    // 检查响应体
    if (!response.body) {
      throw new Error('后端返回空响应')
    }

    // 交由 processResponseStream 处理 SSE 流
    return this.processResponseStream(response.body)
  }

  /**
   * 将 AI SDK 的消息列表转换为后端期望的请求体
   * 后端期望: { messages: ChatMessage[], configId, temperature, top_p }
   */
  protected override prepareSendMessagesRequest = ({
    messages,
  }: Parameters<NonNullable<ConstructorParameters<typeof HttpChatTransport>[0]['prepareSendMessagesRequest']>>[0]) => {
    const backendMessages = messages.map(m => ({
      role: m.role,
      content: m.parts
        ?.filter((p: any) => p.type === 'text')
        .map((p: any) => p.text)
        .join('')
        // fallback: 兼容通过 content 赋值（而非 text+parts）的旧消息
        || (m as any).content
        || '',
    }))

    return {
      body: {
        messages: backendMessages,
        configId: this.configId,
        temperature: this.temperature ?? 0.7,
        top_p: this.topP ?? 0.9,
      },
    }
  }

  /**
   * 解析后端自定义 SSE 流为 AI SDK UIMessageChunk 流
   * 后端格式: data: {"content": "xxx"} 或 data: [DONE]
   * 输出格式: AI SDK v7 标准 —— text-start → text-delta × N → text-end
   */
  protected override processResponseStream(
    stream: ReadableStream<Uint8Array<ArrayBufferLike>>
  ): ReadableStream<UIMessageChunk> {
    const reader = stream.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    // 生成一个本次流式响应的唯一 ID
    const streamId = `stream_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`
    let started = false

    return new ReadableStream<UIMessageChunk>({
      async pull(controller) {
        try {
          while (true) {
            const { done, value } = await reader.read()
            if (done) {
              // 处理 buffer 中可能剩余的未解析行
              if (buffer.trim()) {
                const chunks = parseSseBufferV7(buffer, streamId)
                for (const c of chunks) {
                  if (!started && c.type === 'text-delta') {
                    controller.enqueue({ type: 'text-start', id: streamId })
                    started = true
                  }
                  controller.enqueue(c)
                }
              }
              // 确保发送 text-end
              if (started) {
                controller.enqueue({ type: 'text-end', id: streamId })
              }
              controller.close()
              return
            }

            buffer += decoder.decode(value, { stream: true })
            const lines = buffer.split('\n')
            buffer = lines.pop() || ''

            for (const line of lines) {
              const trimmed = line.trim()
              if (!trimmed || !trimmed.startsWith('data: ')) continue
              const data = trimmed.slice(6)
              if (data === '[DONE]') continue

              try {
                const parsed = JSON.parse(data)
                if (parsed.content) {
                  if (!started) {
                    controller.enqueue({ type: 'text-start', id: streamId })
                    started = true
                  }
                  controller.enqueue({
                    type: 'text-delta',
                    delta: parsed.content,
                    id: streamId,
                  })
                }
                if (parsed.error) {
                  controller.enqueue({
                    type: 'error',
                    errorText: parsed.error,
                  })
                }
              } catch {
                // JSON 解析失败，跳过
              }
            }
          }
        } catch (error) {
          // 确保流被正确关闭
          if (started) {
            controller.enqueue({ type: 'text-end', id: streamId })
          }
          controller.error(error)
        }
      },
      cancel() {
        reader.cancel()
      },
    })
  }
}

/** 解析 buffer 中剩余的 SSE 行（v7 格式输出） */
function parseSseBufferV7(buffer: string, streamId: string): UIMessageChunk[] {
  const chunks: UIMessageChunk[] = []
  const lines = buffer.split('\n')
  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || !trimmed.startsWith('data: ')) continue
    const data = trimmed.slice(6)
    if (data === '[DONE]') continue
    try {
      const parsed = JSON.parse(data)
      if (parsed.content) {
        chunks.push({
          type: 'text-delta',
          delta: parsed.content,
          id: streamId,
        })
      }
    } catch {
      // skip
    }
  }
  return chunks
}

/**
 * 创建 Transport 实例的工厂函数
 */
export function createChatTransport(options: {
  configId?: number
  temperature?: number
  topP?: number
}) {
  return new CustomChatTransport(options)
}
