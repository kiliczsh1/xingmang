// 正文 AI 历史记录工具（写入逻辑；UI 在 HistoryDialog 组件）

export type WriteHistoryRole = 'system' | 'user' | 'prompt' | 'assistant'
export type WriteHistoryStatus = 'completed' | 'cancelled' | 'failed'

export interface WriteHistoryMessage {
  role: WriteHistoryRole
  content: string
  timestamp: number
}

export interface WriteHistoryRecord {
  id: string
  bookId: number
  title: string
  source: 'chat' | 'continue' | 'creative2'
  sourceLabel: string
  promptName: string
  promptCount: number
  status: WriteHistoryStatus
  previewContent: string
  timestamp: number
  messages: WriteHistoryMessage[]
}

const WRITE_HISTORY_STORAGE_KEY = 'write-ai-api-history'

export const readHistoryStorage = (): WriteHistoryRecord[] => {
  try {
    const raw = localStorage.getItem(WRITE_HISTORY_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed as WriteHistoryRecord[] : []
  } catch (error) {
    console.error('读取正文历史记录失败:', error)
    return []
  }
}

export const writeHistoryStorage = (records: WriteHistoryRecord[]) => {
  localStorage.setItem(WRITE_HISTORY_STORAGE_KEY, JSON.stringify(records))
}

export const buildHistoryMessages = (options: {
  promptContents?: string[]
  systemContents?: string[]
  userContents?: string[]
  assistantContents?: string[]
}): WriteHistoryMessage[] => {
  const baseTimestamp = Date.now()
  const messages: WriteHistoryMessage[] = []

  options.systemContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'system',
      content,
      timestamp: baseTimestamp + index
    })
  })

  const promptStart = messages.length
  options.promptContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'prompt',
      content,
      timestamp: baseTimestamp + promptStart + index
    })
  })

  const userStart = messages.length
  options.userContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'user',
      content,
      timestamp: baseTimestamp + userStart + index
    })
  })

  const assistantStart = messages.length
  options.assistantContents?.filter(Boolean).forEach((content, index) => {
    messages.push({
      role: 'assistant',
      content,
      timestamp: baseTimestamp + assistantStart + index
    })
  })

  return messages
}

export const saveHistoryRecord = (bookId: number, record: Omit<WriteHistoryRecord, 'id' | 'bookId' | 'timestamp'>): string => {
  const historyRecord: WriteHistoryRecord = {
    ...record,
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    bookId,
    timestamp: Date.now()
  }

  const records = readHistoryStorage()
  records.unshift(historyRecord)
  writeHistoryStorage(records)

  return historyRecord.id
}
