import type { ApiModel, ApiConfigType } from '@/types'

/**
 * 模型类型（config_type）的共享定义与判定。
 *
 * 后端用 config_type 决定一个模型被哪条链路使用（对话 / 生图 / 生视频 / 图片分析 / 向量），
 * 前端多处都要按类型过滤（封面、插图、工作流媒体节点、API 设置页），因此统一放这里，
 * 避免各处各写一份推断逻辑导致口径漂移。
 */

export interface ModelTypeDef {
  key: ApiConfigType
  name: string
  color: string
  desc: string
}

export const MODEL_TYPES: ModelTypeDef[] = [
  { key: 'chat', name: '对话', color: '#16a34a', desc: '通用对话 / 写作' },
  { key: 'image', name: '图片', color: '#db2777', desc: '封面 / 插图生成' },
  { key: 'video', name: '视频', color: '#7c3aed', desc: '文生视频 / 图生视频' },
  { key: 'vision', name: '视觉', color: '#2563eb', desc: '图片理解 / 分析' },
  { key: 'embed', name: '嵌入', color: '#0891b2', desc: '向量 / 检索' }
]

const VALID_TYPES: ApiConfigType[] = ['chat', 'image', 'video', 'vision', 'embed']

/**
 * 取模型的类型：优先后端 config_type；老数据没有就由 category 兜底推断。
 * 注意 category 里的 chat / reason / multimodal 等能力标签一律归为对话。
 */
export const configTypeOf = (m: Pick<ApiModel, 'config_type' | 'category'>): ApiConfigType => {
  const t = (m?.config_type || '').trim().toLowerCase() as ApiConfigType
  if (VALID_TYPES.includes(t)) return t

  const legacy = (m?.category || '').trim().toLowerCase()
  if (legacy === 'image' || legacy === 'vision' || legacy === 'video' || legacy === 'embed') {
    return legacy
  }
  return 'chat'
}

export const typeDef = (key: string) => MODEL_TYPES.find(t => t.key === key)

export const typeLabel = (m: Pick<ApiModel, 'config_type' | 'category'>): string =>
  typeDef(configTypeOf(m))?.name || '对话'

export const typeColor = (m: Pick<ApiModel, 'config_type' | 'category'>): string =>
  typeDef(configTypeOf(m))?.color || '#64748b'

/**
 * 从模型 ID 推断 config_type（拉取远程模型时用）。
 *
 * `/models` 只返回 id / owned_by，没有类型信息。若不推断，导入时一律按 chat 落库，
 * 图片 / 视频 / 嵌入模型就会挂到错误的链路上（表现为「导入后根本不能用」）。
 *
 * 判定顺序要紧：先判最专有的（embed → video → image），最后才是 vision，
 * 否则 `gpt-image-1` 会被 vision 分支的关键词抢走。
 *
 * ⚠ 与 `guessCategory`（modelMeta.ts，展示用的能力标签）口径不同，不要互相调用：
 * 那个分 chat/reason/multimodal/code/long/embed，这个分 config_type 五种。
 *
 * 通用多模态（gpt-4o / gemini）归「对话」—— 它们主要用途仍是对话，
 * 需要拿来做图片理解时，用户可在弹窗里手动改成「视觉」。
 */
export const guessConfigType = (modelId: string): ApiConfigType => {
  const raw = (modelId || '').toLowerCase()
  const base = (raw.split('/').pop() || raw).split(':')[0]

  if (/(embed|embedding|bge-|gte-|rerank|voyage|jina-embed|text-embedding)/.test(base)) {
    return 'embed'
  }
  if (/(video|sora|veo|kling|seedance|runway|pika|hailuo|cogvideo|ltx-|wanx|wan2|mochi|vidu)/.test(base)) {
    return 'video'
  }
  if (/(dall-e|dalle|gpt-image|imagen|flux|stable-diffusion|sdxl|sd3|midjourney|seedream|kolors|nano-banana|ideogram|recraft|photon|qwen-image|wan-image)/.test(base)) {
    return 'image'
  }
  if (/(vision|visual|llava|internvl|pixtral|minicpm-v|glm-4v|(^|[-_])vl([-_]|$))/.test(base)) {
    return 'vision'
  }
  return 'chat'
}

/** 过滤出指定类型（可多选）的模型；不传类型则原样返回 */
export const filterByTypes = <T extends Pick<ApiModel, 'config_type' | 'category'>>(
  list: T[],
  types?: ApiConfigType | ApiConfigType[]
): T[] => {
  if (!types || (Array.isArray(types) && types.length === 0)) return list
  const wanted = new Set(Array.isArray(types) ? types : [types])
  return list.filter(m => wanted.has(configTypeOf(m)))
}
