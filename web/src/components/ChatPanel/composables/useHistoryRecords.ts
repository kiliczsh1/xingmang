// 历史记录持久化组合函数（后端版）
//
// 原来只存在浏览器 localStorage 的 `write-history-records-v1` 里：清了缓存或换台机器就没了。
// 现在写进 <书>/memory/history/YYYY-MM/ 下的 .md，数据库只留索引。
//
// 界面结构保持不变 —— 整条记录作为 payload 存进文件的 xm:payload 块，读回来整体还原，
// 所以这里除了"存哪儿"之外，调用方不需要改任何东西。
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useMemoryRecords, payloadOf, parseBackendTime } from '@/composables/useMemoryRecords'
import type { MemoryRow } from '@/api'

export interface WriteHistoryMessage {
  role: 'system' | 'prompt' | 'user' | 'assistant'
  content: string
  timestamp: number
}

export interface WriteHistoryRecord {
  id: string
  title: string
  source: string
  sourceLabel: string
  promptName?: string
  promptCount?: number
  status: 'completed' | 'failed' | 'cancelled'
  previewContent?: string
  messages: WriteHistoryMessage[]
  createdAt: number
}

export function useHistoryRecords(storageKey: string, bookId = 0) {
  const historyRecords = ref<WriteHistoryRecord[]>([])
  const memory = useMemoryRecords('generation', { source: 'chat' })
  /** 本地 id → 服务端索引 id，删除时要用 */
  const serverIds = new Map<string, number>()

  const rowToRecord = (row: MemoryRow): WriteHistoryRecord => {
    const original = payloadOf<WriteHistoryRecord>(row)
    const base: WriteHistoryRecord = original
      ? { ...original }
      : {
          id: row.external_id || String(row.id),
          title: row.title,
          source: row.source || 'chat',
          sourceLabel: row.source || '生成',
          status: (row.status as WriteHistoryRecord['status']) || 'completed',
          previewContent: row.result || '',
          messages: [],
          createdAt: parseBackendTime(row.created_at)
        }
    if (!base.previewContent) base.previewContent = row.result || ''
    serverIds.set(base.id, row.id)
    return base
  }

  const buildMessages = (options: {
    promptContents?: string[]
    systemContents?: string[]
    userContents?: string[]
    assistantContents?: string[]
  }) => {
    const baseTimestamp = Date.now()
    const messages: WriteHistoryMessage[] = []

    options.systemContents?.filter(Boolean).forEach((content, index) => {
      messages.push({ role: 'system', content, timestamp: baseTimestamp + index })
    })

    const promptStart = messages.length
    options.promptContents?.filter(Boolean).forEach((content, index) => {
      messages.push({ role: 'prompt', content, timestamp: baseTimestamp + promptStart + index })
    })

    const userStart = messages.length
    options.userContents?.filter(Boolean).forEach((content, index) => {
      messages.push({ role: 'user', content, timestamp: baseTimestamp + userStart + index })
    })

    const assistantStart = messages.length
    options.assistantContents?.filter(Boolean).forEach((content, index) => {
      messages.push({ role: 'assistant', content, timestamp: baseTimestamp + assistantStart + index })
    })

    return messages
  }

  const loadAll = async () => {
    if (!bookId) {
      historyRecords.value = []
      return
    }
    const rows = await memory.list(bookId, true)
    historyRecords.value = rows.map(rowToRecord)
  }

  const saveRecord = (record: Omit<WriteHistoryRecord, 'id' | 'createdAt'>) => {
    const newRecord: WriteHistoryRecord = {
      ...record,
      id: `hist_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`,
      createdAt: Date.now()
    }
    // 先让界面立刻看到（列表是本地数组驱动的），落盘在后台完成
    historyRecords.value.unshift(newRecord)
    if (historyRecords.value.length > 200) {
      historyRecords.value = historyRecords.value.slice(0, 200)
    }
    if (!bookId) return
    void memory
      .save(bookId, {
        externalId: newRecord.id,
        title: newRecord.title,
        source: newRecord.source || 'chat',
        status: newRecord.status || 'completed',
        prompt: newRecord.promptName || '',
        result: newRecord.previewContent || '',
        readable: newRecord.messages
          .map(m => `### ${m.role}\n\n${m.content}`)
          .join('\n\n'),
        payload: newRecord,
        createdAtMs: newRecord.createdAt
      })
      .then(res => {
        if (!res.ok) {
          // 落盘失败要出声，否则用户以为这条记录存住了
          ElMessage.error(`历史记录保存失败：${res.message || '未知错误'}`)
        } else if (res.row) {
          serverIds.set(newRecord.id, res.row.id)
        }
      })
  }

  const deleteRecord = (id: string) => {
    historyRecords.value = historyRecords.value.filter(r => r.id !== id)
    const sid = serverIds.get(id)
    if (sid) {
      serverIds.delete(id)
      void memory.remove(sid)
    }
  }

  const clearAll = async () => {
    try {
      await ElMessageBox.confirm('确定要清空所有历史记录吗？文件会移入回收站。', '清空历史', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }
    for (const rec of [...historyRecords.value]) {
      const sid = serverIds.get(rec.id)
      if (sid) await memory.remove(sid)
    }
    historyRecords.value = []
    serverIds.clear()
    // 浏览器里属于本书的旧记录一并清掉，免得下次迁移又带回来
    try {
      const legacy = JSON.parse(localStorage.getItem(storageKey) || '[]') as any[]
      localStorage.setItem(
        storageKey,
        JSON.stringify(legacy.filter(r => !bookId || r.bookId === undefined || r.bookId === bookId))
      )
    } catch {
      /* 旧数据坏了就不管 */
    }
    ElMessage.success('历史记录已清空')
  }

  const viewDetail = (id: string) => historyRecords.value.find(r => r.id === id)

  /** 首次进入时把浏览器里的旧记录导入一次（需要用户确认）。 */
  const migrateLegacy = async () => {
    if (!bookId) return
    const flagKey = `memory-migrated-generation-${bookId}-${storageKey}`
    if (localStorage.getItem(flagKey)) return
    let legacy: WriteHistoryRecord[] = []
    try {
      legacy = JSON.parse(localStorage.getItem(storageKey) || '[]')
    } catch {
      legacy = []
    }
    if (legacy.length === 0) return
    try {
      await ElMessageBox.confirm(
        `发现有 ${legacy.length} 条历史记录只存在浏览器里。` +
          String.fromCharCode(10) +
          '导入后会写进作品目录，随作品一起备份，清浏览器缓存也不会丢。',
        '导入旧历史记录',
        { confirmButtonText: '导入', cancelButtonText: '暂不', type: 'info' }
      )
    } catch {
      return
    }
    const result = await memory.importLegacy(bookId, storageKey, (item: WriteHistoryRecord) => ({
      externalId: String(item.id),
      title: item.title || '未命名',
      prompt: item.promptName || '',
      result: item.previewContent || '',
      readable: (item.messages || []).map(m => `### ${m.role}\n\n${m.content}`).join('\n\n'),
      status: item.status || 'completed',
      createdAtMs: item.createdAt
    }))
    localStorage.setItem(flagKey, String(Date.now()))
    await loadAll()
    ElMessage.success(`已导入 ${result.imported} 条历史记录（跳过 ${result.skipped} 条）`)
  }

  // 初始化：读取 + 必要时迁移
  void (async () => {
    await loadAll()
    await migrateLegacy()
  })()

  return {
    historyRecords,
    buildMessages,
    saveRecord,
    loadAll,
    deleteRecord,
    clearAll,
    viewDetail
  }
}
