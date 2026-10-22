// 提示词查重 / 卡包合并（DedupMergeDialogs.vue）与 Prompts.vue（导入防重、卡片时间）共用的纯声明。
// ⚠ 只是搬运，不是跨视图统一：PackDetail.vue / PromptPreview.vue 里的同名实现是各自的，别顺手合并。
import type { Prompt } from '@/types'

export type DedupMode = 'titleCategory' | 'title'

export type DedupGroup = {
  key: string
  name: string
  category?: string
  items: Prompt[]
  keepId: number
}

export type MergeMode = 'new' | 'absorb'

// 判重键：标题（可选+分类）拼接，trim 后比较（导入防重与查重弹窗共用）
export const buildPromptKey = (name: string, category: string | undefined, mode: DedupMode = 'titleCategory'): string => {
  const title = (name || '').trim()
  if (mode === 'title') return title
  return `${title}||${(category || '').trim()}`
}

// 格式化日期
export const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '未知时间'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}
