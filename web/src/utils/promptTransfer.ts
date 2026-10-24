// 导入 / 导出的纯机制层：导出编码、心得搬运、判重键集合、文件读取。
// 被 PromptIODialogs.vue（弹窗组件）与 Prompts.vue（整体备份导出）共用。
import { ref } from 'vue'
import { promptInsightAPI } from '@/api'
import type { Prompt } from '@/types'
import { buildPromptKey } from '@/utils/promptDedupMerge'
import type { DedupMode } from '@/utils/promptDedupMerge'

export const EXPORT_MAGIC = 'XNP1'
export const encodeExportData = (data: string): string => {
  const base64 = btoa(unescape(encodeURIComponent(data)))
  const shifted = base64.split('').map(c => String.fromCharCode(c.charCodeAt(0) + 3)).join('')
  return EXPORT_MAGIC + shifted
}

export const decodeExportData = (data: string): string => {
  if (!data.startsWith(EXPORT_MAGIC)) {
    return data
  }
  const shifted = data.slice(EXPORT_MAGIC.length).split('').map(c => String.fromCharCode(c.charCodeAt(0) - 3)).join('')
  return decodeURIComponent(escape(atob(shifted)))
}

// 从现有提示词列表构建判重键集合
export const buildExistingKeySet = (list: Prompt[], mode: DedupMode = 'titleCategory'): Set<string> => {
  const keys = new Set<string>()
  list.forEach(p => {
    const key = buildPromptKey(p.name, p.category, mode)
    if (key) keys.add(key)
  })
  return keys
}

// ============ 心得的导出 / 导入搬运 ============
// 心得是独立资源（一张卡片多条），导出文件里以 `insights` 映射承载，
// 键用 buildPromptKey(name, category) —— 导入时新建的卡片 id 会变，只能靠「名字 + 分类」对上。
export const importedInsights = ref<Record<string, Array<{ content: string; rating: number; tags: string[]; created_at?: string }>>>({})

/** 从导出数据里收集心得：新格式读顶层 insights；旧格式读卡片上的单条 insight 字符串 */
export const registerImportedInsights = (data: any) => {
  if (!data || typeof data !== 'object') return
  const map: Record<string, any[]> = { ...importedInsights.value }

  // 整体备份把 insights 放在 data 里，卡包 / 独立卡片导出放在顶层，两处都读
  const insightSources = [data.insights, data.data?.insights]
  insightSources.forEach(source => {
    if (!source || typeof source !== 'object' || Array.isArray(source)) return
    Object.entries(source).forEach(([key, list]) => {
      if (key && Array.isArray(list) && list.length > 0) map[key] = list as any[]
    })
  })

  const cards: any[] = []
  if (data.prompt) cards.push(data.prompt)
  if (Array.isArray(data.prompts)) cards.push(...data.prompts)
  const collectPacks = (packs: any) => {
    if (!packs || typeof packs !== 'object') return
    Object.values(packs).forEach((arr: any) => { if (Array.isArray(arr)) cards.push(...arr) })
  }
  collectPacks(data.packs)
  collectPacks(data.data?.packs)
  if (Array.isArray(data.data?.standalone)) cards.push(...data.data.standalone)

  cards.forEach(card => {
    if (!card) return
    const key = buildPromptKey(card.name, card.category)
    if (!key || (map[key] && map[key].length > 0)) return
    const text = String(card.insight || '').trim()
    if (text) map[key] = [{ content: text, rating: 0, tags: [], created_at: card.updated_at || card.created_at }]
  })

  importedInsights.value = map
}

/** 取走某张卡片待恢复的心得（取走即删，避免同名卡片重复回挂） */
export const takeImportedInsights = (name: string, category: string) => {
  const map = importedInsights.value
  const key = buildPromptKey(name, category)
  if (key && map[key]?.length) {
    const list = map[key]
    delete map[key]
    return list
  }
  const fallback = Object.keys(map).find(k => k.split('||')[0] === (name || '').trim())
  if (fallback && map[fallback]?.length) {
    const list = map[fallback]
    delete map[fallback]
    return list
  }
  return []
}

/** 卡片新建成功后，把导出文件里带的心得挂回新卡片 */
export const restoreInsightsFor = async (name: string, category: string, newPromptId?: number) => {
  if (!newPromptId) return
  const entries = takeImportedInsights(name, category)
  for (const entry of entries) {
    const content = String(entry?.content || '').trim()
    if (!content) continue
    try {
      await promptInsightAPI.create(newPromptId, {
        content,
        rating: Number(entry.rating) || 0,
        tags: Array.isArray(entry.tags) ? entry.tags.map(String) : []
      })
    } catch {
      // 单条心得失败不影响整体导入
    }
  }
}

// ============ 定级 / 心得笔记的导出 / 导入搬运 ============
// 这两份数据不进库，存在浏览器 localStorage（PackDetail 的「天地玄黄定级」与「藤蔓笔记」共用同一个 key，
// 以卡片 id 为键）。导出时必须改成「名字 + 分类」的键，因为导入后新卡片的 id 全变了。
export const RATING_NOTES_STORAGE_KEY = 'pack_detail_rating_notes'

type RatingNotesStore = {
  ratings: Record<string, { effect: number; quality: number; universal: number; bonus: number }>
  notes: Record<string, Array<{ id: string; content: string; created_at: string; updated_at: string }>>
}

const readRatingNotesStore = (): RatingNotesStore => {
  try {
    const raw = localStorage.getItem(RATING_NOTES_STORAGE_KEY)
    if (!raw) return { ratings: {}, notes: {} }
    const parsed = JSON.parse(raw)
    return {
      ratings: parsed && typeof parsed.ratings === 'object' && parsed.ratings ? parsed.ratings : {},
      notes: parsed && typeof parsed.notes === 'object' && parsed.notes ? parsed.notes : {}
    }
  } catch {
    return { ratings: {}, notes: {} }
  }
}

/** 导出前收集定级 + 心得笔记，按卡片键（名字 + 分类）分组 */
export const collectRatingNotesByKey = (prompts: Prompt[]) => {
  const store = readRatingNotesStore()
  const ratings: Record<string, any> = {}
  const notes: Record<string, any[]> = {}
  prompts.forEach(p => {
    const key = buildPromptKey(p.name, p.category)
    if (!key) return
    const rating = store.ratings[String(p.id)]
    if (rating) ratings[key] = rating
    const list = store.notes[String(p.id)]
    if (Array.isArray(list) && list.length > 0) notes[key] = list
  })
  return { ratings, notes }
}

// 导入侧：先登记文件里的数据，卡片建好后再按新 id 回写
export const importedRatings = ref<Record<string, any>>({})
export const importedNotes = ref<Record<string, any[]>>({})
let pendingRatingNotes: Record<string, { rating?: any; notes?: any[] }> = {}

export const resetImportedRatingNotes = () => {
  importedRatings.value = {}
  importedNotes.value = {}
  pendingRatingNotes = {}
}

/** 读导入文件里的 extras（整体备份放 data.extras，单卡/卡包导出放顶层） */
export const registerImportedRatingNotes = (data: any) => {
  const sources = [data?.extras, data?.data?.extras]
  sources.forEach(source => {
    if (!source || typeof source !== 'object') return
    if (source.ratings && typeof source.ratings === 'object') {
      importedRatings.value = { ...importedRatings.value, ...source.ratings }
    }
    if (source.notes && typeof source.notes === 'object') {
      importedNotes.value = { ...importedNotes.value, ...source.notes }
    }
  })
}

/** 卡片创建成功后调用：把定级/笔记挂到新 id 上，等 flush 一次性写回 localStorage */
export const restoreRatingNotesFor = (name: string, category: string, newPromptId?: number) => {
  if (!newPromptId) return
  const key = buildPromptKey(name, category)
  if (!key) return
  const rating = importedRatings.value[key]
  const notes = importedNotes.value[key]
  if (!rating && !(Array.isArray(notes) && notes.length > 0)) return
  pendingRatingNotes[String(newPromptId)] = { rating, notes }
  delete importedRatings.value[key]
  delete importedNotes.value[key]
}

/** 一次性合并写回 localStorage（整体恢复会建很多卡片，逐张写太慢） */
export const flushRestoredRatingNotes = () => {
  const ids = Object.keys(pendingRatingNotes)
  if (ids.length === 0) return
  const store = readRatingNotesStore()
  ids.forEach(id => {
    const entry = pendingRatingNotes[id]
    if (entry.rating) store.ratings[id] = entry.rating
    if (Array.isArray(entry.notes) && entry.notes.length > 0) store.notes[id] = entry.notes
  })
  try {
    localStorage.setItem(RATING_NOTES_STORAGE_KEY, JSON.stringify(store))
  } catch {
    // 存储满等异常忽略
  }
  pendingRatingNotes = {}
}

/** 导出前收集全部心得，按卡片键分组 */
export const collectInsightsByKey = async (prompts: Prompt[]) => {
  const grouped: Record<string, any[]> = {}
  try {
    const res = await promptInsightAPI.getAll()
    const all = res.success && Array.isArray(res.data) ? res.data : []
    if (all.length === 0) return grouped
    const byId = new Map<number, Prompt>()
    prompts.forEach(p => byId.set(p.id, p))
    all.forEach(item => {
      const owner = byId.get(item.prompt_id)
      if (!owner) return
      const key = buildPromptKey(owner.name, owner.category)
      if (!key) return
      if (!grouped[key]) grouped[key] = []
      grouped[key].push({
        content: item.content,
        rating: item.rating,
        tags: item.tags || [],
        created_at: item.created_at
      })
    })
  } catch {
    // 心得收集失败不阻断主流程，导出文件里就没有 insights 字段
  }
  return grouped
}
