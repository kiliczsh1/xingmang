export const WorldBookEntryPosition = {
  BEFORE_PROMPT: 0,
  AFTER_PROMPT: 1,
  AFTER_AUTHOR_NOTE: 2,
  AFTER_CHARACTER: 3,
  BEFORE_EXAMPLES: 4,
  AFTER_EXAMPLES: 5,
  AT_DEPTH: 6,
  TOP_OF_WORLD_INFO: 7,
  BOTTOM_OF_WORLD_INFO: 8
} as const

export type WorldBookEntryPosition = typeof WorldBookEntryPosition[keyof typeof WorldBookEntryPosition]

export const WorldBookEntryType = {
  NORMAL: 'normal',
  CONSTANT: 'constant',
  VECTORIZED: 'vectorized'
} as const

export type WorldBookEntryType = typeof WorldBookEntryType[keyof typeof WorldBookEntryType]

export interface WorldBookEntry {
  uid: string
  key: string[]
  keysecondary: string[]
  comment: string
  content: string
  constant: boolean
  selective: boolean
  order: number
  position: WorldBookEntryPosition
  disable: boolean
  excludeRecursion: boolean
  preventRecursion: boolean
  delayUntilRecursion: boolean
  probability: number
  useProbability: boolean
  depth: number
  selectLogic: number
  group: string
  groupOverride: boolean
  groupWeight: number
  scanDepth: number | null
  caseSensitive: boolean
  matchWholeWords: boolean
  useGroupScoring: boolean
  automationId: string
  role: number | null
  vectorized: boolean
  displayIndex: number
  sticky: number
  cooldown: number
  delay: number
}

export interface WorldBook {
  entries: WorldBookEntry[]
  name: string
  source: 'global' | 'book'
  bookId?: string
}

export function createDefaultWorldBookEntry(): WorldBookEntry {
  return {
    uid: generateUID(),
    key: [],
    keysecondary: [],
    comment: '',
    content: '',
    constant: false,
    selective: false,
    order: 100,
    position: WorldBookEntryPosition.BOTTOM_OF_WORLD_INFO,
    disable: false,
    excludeRecursion: false,
    preventRecursion: false,
    delayUntilRecursion: false,
    probability: 100,
    useProbability: false,
    depth: 4,
    selectLogic: 0,
    group: '',
    groupOverride: false,
    groupWeight: 100,
    scanDepth: null,
    caseSensitive: false,
    matchWholeWords: false,
    useGroupScoring: false,
    automationId: '',
    role: null,
    vectorized: false,
    displayIndex: 0,
    sticky: 0,
    cooldown: 0,
    delay: 0
  }
}

export function generateUID(): string {
  return Date.now().toString(36) + Math.random().toString(36).substring(2)
}

export interface WorldBookExport {
  entries: WorldBookEntry[]
  name: string
  exportDate: string
  version: string
}

export function exportWorldBook(worldBook: WorldBook): WorldBookExport {
  return {
    entries: worldBook.entries,
    name: worldBook.name,
    exportDate: new Date().toISOString(),
    version: '1.0.0'
  }
}

export function importWorldBook(data: any): WorldBook | null {
  try {
    if (!data.entries || !Array.isArray(data.entries)) {
      return null
    }

    const entries: WorldBookEntry[] = data.entries.map((entry: any) => ({
      ...createDefaultWorldBookEntry(),
      ...entry,
      uid: entry.uid || generateUID()
    }))

    return {
      entries,
      name: data.name || '导入的世界书',
      source: 'global'
    }
  } catch (error) {
    console.error('导入世界书失败:', error)
    return null
  }
}
