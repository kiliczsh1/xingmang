/**
 * 媒体接口适配器：对应 Go 侧 internal/handlers/media.go
 *
 * - POST /api/media/image        文生图（一次性 JSON）
 * - POST /api/media/analyze      图片分析（text/event-stream）
 * - POST /api/media/video        文生视频（后端待接入具体服务商）
 * - POST /api/media/image-to-video
 *
 * 与 aiBridge 一致，直接用 fetch 而不是 axios：
 * 执行期错误由工作流日志消化，不要再叠加全局 ElMessage 弹窗。
 */
import { AbortedError } from './aiBridge'

export interface MediaImage {
  b64?: string
  url?: string
  revised_prompt?: string
}

export interface GenerateImageOptions {
  configId?: number | null
  /** 星域节点用模型名（model）指定模型；后端按名字解析 */
  model?: string
  prompt: string
  size?: string
  count?: number
  negativePrompt?: string
  signal?: AbortSignal
}

const readError = async (resp: Response): Promise<string> => {
  let msg = `HTTP ${resp.status}`
  try {
    const j = await resp.json()
    msg = j?.message || j?.error || msg
  } catch {
    /* 忽略解析失败 */
  }
  return msg
}

/** 文生图：返回归一化后的图片列表（b64 或 url） */
export async function generateImage(opts: GenerateImageOptions): Promise<MediaImage[]> {
  const body: Record<string, unknown> = {
    prompt: opts.prompt,
    size: opts.size || '1024x1024',
    count: Math.max(1, Math.min(8, opts.count || 1))
  }
  if (opts.configId) body.config_id = opts.configId
  if (opts.model) body.model = opts.model
  if (opts.negativePrompt) body.negative_prompt = opts.negativePrompt

  let resp: Response
  try {
    resp = await fetch('/api/media/image', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: opts.signal
    })
  } catch (e: any) {
    if (e?.name === 'AbortError') throw new AbortedError()
    throw new Error(`生图请求失败: ${e?.message || e}`)
  }

  if (!resp.ok) throw new Error(await readError(resp))
  const json: any = await resp.json().catch(() => null)
  const images = json?.data?.images
  if (!Array.isArray(images) || !images.length) throw new Error('生图接口没有返回图片数据')
  return images as MediaImage[]
}

export interface AnalyzeImageOptions {
  configId?: number | null
  /** 星域节点用模型名（model）指定模型；后端按名字解析 */
  model?: string
  imageBase64?: string
  imageUrl?: string
  prompt?: string
  detail?: string
  onDelta?: (fullText: string) => void
  signal?: AbortSignal
}

/** 图片分析（SSE 流式），返回完整文本 */
export async function analyzeImage(opts: AnalyzeImageOptions): Promise<string> {
  const body: Record<string, unknown> = {}
  if (opts.configId) body.config_id = opts.configId
  if (opts.model) body.model = opts.model
  if (opts.imageBase64) body.image_base64 = opts.imageBase64
  if (opts.imageUrl) body.image_url = opts.imageUrl
  if (opts.prompt) body.prompt = opts.prompt
  if (opts.detail) body.detail = opts.detail

  let resp: Response
  try {
    resp = await fetch('/api/media/analyze', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: opts.signal
    })
  } catch (e: any) {
    if (e?.name === 'AbortError') throw new AbortedError()
    throw new Error(`图片分析请求失败: ${e?.message || e}`)
  }

  if (!resp.ok) throw new Error(await readError(resp))
  if (!resp.body) throw new Error('图片分析响应没有流式内容')

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
          return content
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
    throw e
  } finally {
    try {
      reader.releaseLock()
    } catch {
      /* 已释放 */
    }
  }

  if (errorMessage) throw new Error(errorMessage)
  return content
}

export interface GenerateVideoOptions {
  configId?: number | null
  /** 星域节点用模型名（model）指定模型 */
  model?: string
  prompt: string
  /** 星域字段：quality（质量优先/速度优先） */
  quality?: string
  /** 星域字段：with_audio */
  withAudio?: boolean
  duration?: number
  fps?: number
  size?: string
  signal?: AbortSignal
}

export interface ImageToVideoOptions {
  configId?: number | null
  /** 星域节点用模型名（model）指定模型 */
  model?: string
  /** 星域字段：fps */
  fps?: number
  imageBase64?: string
  imageUrl?: string
  prompt?: string
  duration?: number
  signal?: AbortSignal
}

/**
 * 视频类接口：Go 侧目前只保留契约并返回 501（各家服务商协议不同，需要按选定的服务商实现提交+轮询）。
 * 这里如实把后端的错误抛出来，不做假成功。
 */
async function postVideo(path: string, body: Record<string, unknown>, signal?: AbortSignal) {
  let resp: Response
  try {
    resp = await fetch(path, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal
    })
  } catch (e: any) {
    if (e?.name === 'AbortError') throw new AbortedError()
    throw new Error(`视频请求失败: ${e?.message || e}`)
  }
  if (!resp.ok) throw new Error(await readError(resp))
  const json: any = await resp.json().catch(() => null)
  return (json?.data || {}) as { url?: string; b64?: string; task_id?: string; status?: string }
}

export async function generateVideo(opts: GenerateVideoOptions) {
  const body: Record<string, unknown> = { prompt: opts.prompt, duration: opts.duration || 5, fps: opts.fps || 24 }
  if (opts.configId) body.config_id = opts.configId
  if (opts.model) body.model = opts.model
  if (opts.size) body.size = opts.size
  if (opts.quality) body.quality = opts.quality
  if (typeof opts.withAudio === 'boolean') body.with_audio = opts.withAudio
  return postVideo('/api/media/video', body, opts.signal)
}

export async function imageToVideo(opts: ImageToVideoOptions) {
  const body: Record<string, unknown> = { duration: opts.duration || 5, fps: opts.fps || 30 }
  if (opts.configId) body.config_id = opts.configId
  if (opts.model) body.model = opts.model
  if (opts.imageBase64) body.image_base64 = opts.imageBase64
  if (opts.imageUrl) body.image_url = opts.imageUrl
  if (opts.prompt) body.prompt = opts.prompt
  return postVideo('/api/media/image-to-video', body, opts.signal)
}
