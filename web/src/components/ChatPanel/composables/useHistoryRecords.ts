// 历史记录持久化组合函数
// 从 Write.vue 行 4893-5013 搬迁
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

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

export function useHistoryRecords(storageKey: string) {
  const historyRecords = ref<WriteHistoryRecord[]>([])

  const readStorage = (): WriteHistoryRecord[] => {
    try {
      const raw = localStorage.getItem(storageKey)
      const parsed = raw ? JSON.parse(raw) : []
      return Array.isArray(parsed) ? parsed : []
    } catch (error) {
      console.error('读取历史记录失败:', error)
      return []
    }
  }

  const writeStorage = (records: WriteHistoryRecord[]) => {
    localStorage.setItem(storageKey, JSON.stringify(records))
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

  const saveRecord = (record: Omit<WriteHistoryRecord, 'id' | 'createdAt'>) => {
    const newRecord: WriteHistoryRecord = {
      ...record,
      id: `hist_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`,
      createdAt: Date.now(),
    }
    historyRecords.value.unshift(newRecord)
    // 保留最近 200 条
    if (historyRecords.value.length > 200) {
      historyRecords.value = historyRecords.value.slice(0, 200)
    }
    writeStorage(historyRecords.value)
  }

  const loadAll = () => {
    historyRecords.value = readStorage()
  }

  const deleteRecord = (id: string) => {
    historyRecords.value = historyRecords.value.filter(r => r.id !== id)
    writeStorage(historyRecords.value)
  }

  const clearAll = async () => {
    try {
      await ElMessageBox.confirm('确定要清空所有历史记录吗？', '清空历史', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      })
      historyRecords.value = []
      writeStorage([])
      ElMessage.success('历史记录已清空')
    } catch {
      // 用户取消
    }
  }

  const viewDetail = (id: string) => {
    return historyRecords.value.find(r => r.id === id)
  }

  // 初始化
  loadAll()

  return {
    historyRecords,
    buildMessages,
    saveRecord,
    loadAll,
    deleteRecord,
    clearAll,
    viewDetail,
  }
}
