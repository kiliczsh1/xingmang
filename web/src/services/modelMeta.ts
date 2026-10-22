import type { Component } from 'vue'
import {
  ChatDotRound,
  MagicStick,
  PictureFilled,
  Cpu,
  Document,
  Connection
} from '@element-plus/icons-vue'
import type { ApiModel } from '@/types'

/**
 * 模型「能力分类」的共享定义与推断。
 *
 * 注意与 `@/services/modelType` 的区别：
 * - **类型（config_type）**：后端强约束，决定模型被哪条链路使用（对话 / 生图 / 视频 / 视觉 / 向量）
 * - **能力分类（category）**：仅用于展示的能力标签，老数据可能没有，由模型 ID 推断
 *
 * 这里放的是后者，以及上下文长度的推断表 —— 原先都写在 Config.vue 里，
 * 拆分 CapDialog 后需要被两个文件共用，因此提取到此处。
 */

export const categories = [
  { key: 'chat', name: '通用对话', color: '#16a34a', icon: ChatDotRound },
  { key: 'reason', name: '深度推理', color: '#7c3aed', icon: MagicStick },
  { key: 'multimodal', name: '多模态', color: '#db2777', icon: PictureFilled },
  { key: 'code', name: '代码生成', color: '#2563eb', icon: Cpu },
  { key: 'long', name: '长文本', color: '#d97706', icon: Document },
  { key: 'embed', name: '嵌入向量', color: '#0891b2', icon: Connection }
]

/**
 * 从模型 ID 自动推断分类：
 * 1. 规范化 ID：剥掉 `/` 前缀（openrouter 的 deepseek/deepseek-r1）与 `:tag` 后缀（ollama 的 qwen2.5:7b）
 * 2. 正则精确匹配能力关键词
 * 3. 家族名兜底：取第一个 `-` 前的家族名，映射到能力分类
 */
export const guessCategory = (modelId: string): string => {
  const raw = (modelId || '').toLowerCase()
  const base = (raw.split('/').pop() || raw).split(':')[0]

  if (/(reasoner|r1|thinking|o1-|o3-|o4-)/.test(base)) return 'reason'
  if (/(vl|vision|visual|image|gemini|llava|omni)/.test(base)) return 'multimodal'
  if (/(coder|codegeex|starcoder|code-)/.test(base)) return 'code'
  if (/(bge|embed|rerank|text-embedding)/.test(base)) return 'embed'
  if (/(longcontext|long-context|1\.5-pro|max-longcontext)/.test(base)) return 'long'

  const family = base.split('-')[0]?.trim()
  if (family && family !== base) {
    const familyMap: Record<string, string> = {
      code: 'code', coder: 'code', codegeex: 'code', starcoder: 'code',
      reason: 'reason', reasoner: 'reason', thinking: 'reason',
      vl: 'multimodal', vision: 'multimodal', visual: 'multimodal', llava: 'multimodal', omni: 'multimodal',
      bge: 'embed', embed: 'embed', embedding: 'embed', rerank: 'embed',
      long: 'long', longcontext: 'long'
    }
    if (familyMap[family]) return familyMap[family]
  }
  return 'chat'
}

export const categoryOf = (m: Pick<ApiModel, 'category' | 'model'>): string =>
  m.category || guessCategory(m.model || '')

export const catDef = (key: string) => categories.find(c => c.key === key)

export const catColor = (m: Pick<ApiModel, 'category' | 'model'>): string =>
  catDef(categoryOf(m))?.color || '#64748b'

export const catLabel = (m: Pick<ApiModel, 'category' | 'model'>): string =>
  catDef(categoryOf(m))?.name || '未分类'

export const catIcon = (m: Pick<ApiModel, 'category' | 'model'>): Component =>
  catDef(categoryOf(m))?.icon || Document

/* ---------------- token 上限推断 ----------------
   两张表，语义完全不同，别再混用：
   - KNOWN_MODEL_MAX_TOKENS = **上下文窗口总量**（输入+输出），只用于「最大 Token」滑条的上限参考
   - KNOWN_MODEL_MAX_OUTPUT = **单次输出上限**，这才是请求里 max_tokens 的真实含义
   历史坑：早先只有前者，导入模型时却直接拿它写进了 max_tokens —— 于是弹窗上显示的是
   「上下文窗口」当「输出上限」用。两者差一个数量级。 */

/** 上下文窗口（总量），单位 token */
export const KNOWN_MODEL_MAX_TOKENS: Record<string, number> = {
  'gpt-4': 8192, 'gpt-4-turbo': 128000, 'gpt-4o': 128000, 'gpt-4o-mini': 128000,
  'gpt-4-32k': 32768, 'gpt-3.5-turbo': 4096, 'gpt-3.5-turbo-16k': 16384,
  'claude-3-opus': 200000, 'claude-3-sonnet': 200000, 'claude-3-haiku': 200000,
  'claude-3.5-sonnet': 200000, 'claude-3.5-haiku': 200000,
  'deepseek-chat': 128000, 'deepseek-reasoner': 128000, 'deepseek-r1': 128000,
  // v4 / v4.1 一代：实抓官方 /v1/models 的 context_window = 1048576
  'deepseek': 1048576,
  'qwen-turbo': 131072, 'qwen-plus': 131072, 'qwen-max': 32768, 'qwen-max-longcontext': 1048576,
  'gemini-1.5-pro': 2097152, 'gemini-1.5-flash': 1048576, 'gemini-2.0-flash': 1048576,
  'gemini-2.5-pro': 1048576, 'gemini-2.5-flash': 1048576,
  'gemini-3-pro-preview': 1048576, 'gemini-3-flash-preview': 1048576,
  'kimi-k2.5': 131072, 'glm-4': 131072, 'glm-5': 131072,
  'llama-3-70b': 8192, 'llama-3-8b': 8192, 'mixtral-8x7b': 32768,
  'mistral-large': 131072, 'mistral-small': 32768,
  'command-r': 131072, 'command-r-plus': 131072
}

/**
 * 单次输出上限的**本地兜底表**（请求里 max_tokens 的真实语义）。
 *
 * ⚠ 只在**上游没给**的时候才用。优先走上游：后端 `DiscoveredModel.MaxOutputTokens`
 *   会从 `/models` 解析 `max_output_tokens`（DeepSeek 官方实测返回 393216 / 384K），
 *   前端导入时拿到就直接写进模型；第三方中转多数不透传该字段，才落到这张表。
 *
 * ⚠⚠ **表里的数字必须来自官方文档或实抓，禁止估。** 曾经照老模型印象拍了一批"保守估值"，
 *   结果 deepseek 写成 16K（实际 384K）、glm 写成 4K（实际 128K）、claude-opus-4 写成 32K（实际 128K）
 *   —— 系统性偏低，越新的模型错得越多。查不到的家族**别往这里加**，让它走 8192 兜底、
 *   在界面里手工改；填一个看起来合理的假数比留空危害大得多。
 *
 * 来源标注：
 *   deepseek     实抓 GET https://api.deepseek.com/v1/models → max_output_tokens=393216
 *   gemini-3     官方 output token limit 65,536
 *   glm-5.x      智谱官方「最大输出 Tokens 为 128K」
 *   claude 4/5   官方「1M 上下文模型单请求最多 128k 输出」，Opus 4.6 默认 64k / 上限 128k
 *   gpt-5        官方 128,000；gpt-4o 16,384
 */
export const KNOWN_MODEL_MAX_OUTPUT: Record<string, number> = {
  // DeepSeek（v4 / v4.1 一代统一 384K）
  'deepseek': 393216,
  // Google
  'gemini-2.5': 65536,
  'gemini-3': 65536,
  'gemini-2': 8192,
  'gemini-1.5': 8192,
  // 智谱
  'glm-5.3': 131072, 'glm-5': 131072,
  'glm-4': 4096,
  // Anthropic
  'claude-opus-4': 128000, 'claude-sonnet-4': 128000,
  'claude-opus-5': 128000, 'claude-sonnet-5': 128000,
  'claude-3.5': 8192, 'claude-3-5': 8192,
  'claude-3-': 4096,
  'claude': 128000,
  // OpenAI
  'gpt-5': 128000,
  'gpt-4o': 16384,
  'gpt-4': 8192,
  'gpt-3.5': 4096
}

/**
 * 规范化模型 ID，两张表共用：
 * 1. 小写
 * 2. 剥掉 `[...]` / `【...】` 这类**展示用装饰前缀**（用户给模型起的别名，如 `[JY]deepseek-v4-flash-0731`）
 * 3. 剥掉 `org/` 前缀（openrouter 的 `deepseek/deepseek-r1`）
 * 4. 剥掉 `:tag` 后缀（ollama 的 `qwen2.5:7b`）
 */
const normalizeModelId = (modelId: string) => {
  let s = (modelId || '').toLowerCase().trim()
  s = s.split('/').pop() || s
  s = s.split(':')[0]
  s = s.replace(/^[【\[（(][^】\]）)]*[】\]）)]\s*/, '')
  return s.trim()
}

/**
 * 在表里按家族名查值。
 *
 * ⚠ **必须按 key 长度降序尝试**：`includes` 匹配下 `gpt-4o-mini` 也含 `gpt-4`，
 * 按对象声明顺序遍历会先命中短 key `gpt-4` 拿到 8192，把 `gpt-4o` 的 16384 顶掉。
 * 长 key 优先才能让更具体的条目生效。
 */
const lookupByFamily = (table: Record<string, number>, modelId: string, fallback: number) => {
  const id = normalizeModelId(modelId)
  if (!id) return fallback
  const keys = Object.keys(table).sort((a, b) => b.length - a.length)
  for (const k of keys) {
    if (id.includes(k)) return table[k]
  }
  return fallback
}

/**
 * 推断模型**上下文窗口**（供「最大 Token」滑条的取值上限用）。
 * 上游给了就用上游的（`DiscoveredModel.ContextWindow`），否则查表，再否则 8192。
 */
export const resolveMaxTokens = (modelId: string, apiMaxTokens?: number | null) =>
  apiMaxTokens != null ? apiMaxTokens : lookupByFamily(KNOWN_MODEL_MAX_TOKENS, modelId, 8192)

/** 单次输出上限的来源。空串 = 未知 / 未设置 —— 界面据此显示「未设置」而不是假装确定。 */
export type MaxTokensSource = 'upstream' | 'catalog' | 'table' | 'manual' | 'probed' | ''

/**
 * **本地兜底**推断单次输出上限 —— 只在导入时拿不到上游 `max_output_tokens` 时才调用。
 * 上游优先，见 `Config.vue#importModels`。
 *
 * ⚠ 表未命中时返回 **0（未设置）**，不再兜 8192。
 * 兜一个"看起来合理"的数字是最隐蔽的错误：它和真值形状完全一样，用户无从分辨，
 * 而且偏小时会**静默截断输出**。0 会被后端解释为「本次不发送 max_tokens」，
 * 交给上游按模型自身能力决定（见 `internal/aitokens` 的 omitTokens）。
 */
export const resolveMaxOutput = (modelId: string): { value: number; source: MaxTokensSource } => {
  const id = normalizeModelId(modelId)
  if (!id) return { value: 0, source: '' }
  const keys = Object.keys(KNOWN_MODEL_MAX_OUTPUT).sort((a, b) => b.length - a.length)
  for (const k of keys) {
    if (id.includes(k)) return { value: KNOWN_MODEL_MAX_OUTPUT[k], source: 'table' }
  }
  return { value: 0, source: '' }
}

/** 把 token 数格式化成人类可读（如 128K）——与 Config.vue 原实现保持一致 */
export const fmtTok = (n: number) => (n >= 1000 ? `${(n / 1000).toFixed(0)}K` : `${n}`)
