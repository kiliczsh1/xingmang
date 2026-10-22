// 拆书 / 生成历史的统一读写：文件为准，数据库只做索引。
//
// 这两类记录原先只存在浏览器 localStorage 里 —— 既没有库行，也没有文件，
// 清了缓存或换台机器就没了。现在正文落在 <书>/memory/analysis/ 与
// <书>/memory/history/YYYY-MM/ 下的 .md 文件里。
//
// 关键点：界面上的记录结构（messages / fieldValues / relatedContext …）各式各样，
// 我们不把它拆成字段，而是整段 JSON 存进文件的 xm:payload 块，读回来再整体还原。
// 这样前端加字段不需要服务端配合，也不会静默丢字段。
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { memoryAPI, type MemoryRow, type MemorySaveInput } from '@/api'

export type MemoryKind = 'analysis' | 'generation'

/** 后端返回的 created_at 是本地时间 "YYYY-MM-DD HH:mm:ss"，直接给 Date 解析不可靠。 */
export function parseBackendTime(s: string): number {
  if (!s) return 0
  const normalized = s.includes('T') ? s : s.replace(' ', 'T')
  const t = Date.parse(normalized)
  return Number.isFinite(t) ? t : 0
}

/** 从索引行里取出原始记录对象；payload 坏了就返回 null（界面退化为只用索引字段）。 */
export function payloadOf<T>(row: MemoryRow | undefined | null): T | null {
  if (!row?.payload) return null
  try {
    return JSON.parse(row.payload) as T
  } catch {
    return null
  }
}

export interface UseMemoryRecordsOptions {
  /** 新建记录时的 source 标记（续写/改写/创作/拆书…），列表与文件里都会带上 */
  source?: string
  /** 记录标题的兜底来源：payload 里没有 title 时用它 */
  fallbackTitle?: string
}

export function useMemoryRecords(kind: MemoryKind, options: UseMemoryRecordsOptions = {}) {
  const rows = ref<MemoryRow[]>([])
  const loading = ref(false)
  const lastError = ref('')

  /** 读某本书的记录。full=true 时连正文与 payload 一起拿回来，界面才能还原原记录。 */
  const list = async (bookId: number, full = true) => {
    if (!bookId) {
      rows.value = []
      return []
    }
    loading.value = true
    lastError.value = ''
    try {
      const res: any = await memoryAPI.list(bookId, kind, full)
      rows.value = res?.success ? (res.data || []) : []
      return rows.value
    } catch (e: any) {
      lastError.value = e?.response?.data?.message || e?.message || '读取失败'
      rows.value = []
      return []
    } finally {
      loading.value = false
    }
  }

  /** 新建或覆盖一条记录。existingId 有值表示更新同一条（文件会被覆盖）。 */
  const save = async (
    bookId: number,
    record: {
      externalId?: string
      title: string
      source?: string
      status?: string
      prompt?: string
      chapters?: number[]
      result: string
      readable?: string
      payload?: unknown
      createdAtMs?: number
      existingId?: number
    }
  ) => {
    const body: MemorySaveInput = {
      kind,
      external_id: record.externalId || '',
      title: record.title || options.fallbackTitle || '未命名',
      source: record.source || options.source || '',
      status: record.status || 'completed',
      prompt: record.prompt || '',
      chapters: record.chapters || [],
      result: record.result || '',
      readable: record.readable || '',
      payloadText: record.payload === undefined ? '' : JSON.stringify(record.payload),
      created_at_ms: record.createdAtMs || 0
    }
    try {
      const res: any = record.existingId
        ? await memoryAPI.update(record.existingId, body)
        : await memoryAPI.create(bookId, body)
      if (!res?.success) {
        return { ok: false, message: res?.message || '保存失败' }
      }
      return { ok: true, row: res.data as MemoryRow }
    } catch (e: any) {
      // 409：文件被外部改过，程序不覆盖 —— 如实告诉用户，别让他以为保存成功了
      if (e?.response?.status === 409) {
        return { ok: false, conflict: true, message: e.response.data?.message || '文件已被外部修改' }
      }
      return { ok: false, message: e?.response?.data?.message || e?.message || '保存失败' }
    }
  }

  /** 删除：文件进回收站，索引行消失。 */
  const remove = async (id: number) => {
    try {
      const res: any = await memoryAPI.remove(id)
      if (res?.success) {
        rows.value = rows.value.filter(r => r.id !== id)
        return true
      }
      return false
    } catch (e: any) {
      ElMessage.error(e?.response?.data?.message || '删除失败')
      return false
    }
  }

  /**
   * 把浏览器里的旧记录一次性导入。
   *
   * 按 (书, 类型, 原 id) 去重，所以点两次不会产生重复。
   * 不做静默迁移 —— 同一台浏览器可能被多本书共用，必须让用户确认过再动。
   */
  const importLegacy = async (
    bookId: number,
    storageKey: string,
    mapFn: (item: any) => {
      externalId: string
      title: string
      result: string
      readable?: string
      prompt?: string
      chapters?: number[]
      status?: string
      createdAtMs?: number
    }
  ) => {
    let raw: any[] = []
    try {
      const stored = localStorage.getItem(storageKey)
      const parsed = stored ? JSON.parse(stored) : []
      raw = Array.isArray(parsed) ? parsed : []
    } catch {
      raw = []
    }
    if (raw.length === 0) return { imported: 0, skipped: 0 }

    // 已有的 external_id：去重靠它，也靠数据库的唯一索引兜底
    await list(bookId, false)
    const existing = new Set(rows.value.map(r => r.external_id).filter(Boolean))

    let imported = 0
    let skipped = 0
    for (const item of raw) {
      const mapped = mapFn(item)
      if (!mapped?.externalId || existing.has(mapped.externalId)) {
        skipped++
        continue
      }
      const saved = await save(bookId, {
        ...mapped,
        payload: item,
        source: options.source
      })
      if (saved.ok) {
        imported++
        existing.add(mapped.externalId)
      } else {
        skipped++
      }
    }
    await list(bookId, true)
    return { imported, skipped }
  }

  return { rows, loading, lastError, list, save, remove, importLegacy }
}
