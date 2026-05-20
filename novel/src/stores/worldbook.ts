import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { WorldBook, WorldBookEntry } from '@/types/worldbook'
import { createDefaultWorldBookEntry, generateUID } from '@/types/worldbook'

const STORAGE_KEY_GLOBAL = 'worldbook-global'
const STORAGE_KEY_BOOKS = 'worldbook-books'
const STORAGE_KEY_CURRENT = 'worldbook-current-names'
const DEFAULT_WB_NAME = '默认世界书'

export const useWorldBookStore = defineStore('worldbook', () => {
  const globalWorldBook = ref<WorldBook>({
    entries: [],
    name: '全局世界书',
    source: 'global'
  })

  // bookId -> WorldBook[] (多个世界书按名称区分)
  const bookWorldBooks = ref<Map<string, WorldBook[]>>(new Map())
  // bookId -> 当前选中世界书名称
  const currentWorldBookNames = ref<Map<string, string>>(new Map())

  const enabled = ref(true)

  const allEntries = computed(() => {
    const entries: WorldBookEntry[] = [...globalWorldBook.value.entries]
    bookWorldBooks.value.forEach(books => {
      books.forEach(book => {
        entries.push(...book.entries)
      })
    })
    return entries
  })

  const enabledEntries = computed(() => 
    allEntries.value.filter(entry => !entry.disable)
  )

  const globalEntryCount = computed(() => globalWorldBook.value.entries.length)

  // --- 工具函数 ---
  const ensureCurrentName = (bookId: string): string => {
    if (!currentWorldBookNames.value.has(bookId)) {
      currentWorldBookNames.value.set(bookId, DEFAULT_WB_NAME)
      saveCurrentNames()
    }
    return currentWorldBookNames.value.get(bookId) || DEFAULT_WB_NAME
  }

  const markDirty = (bookId: string) => {
    const books = bookWorldBooks.value.get(bookId)
    if (books) {
      bookWorldBooks.value.set(bookId, [...books])
    }
  }

  const ensureWorldBook = (bookId: string, name?: string): WorldBook => {
    const wbName = name || ensureCurrentName(bookId)
    let books = bookWorldBooks.value.get(bookId)
    if (!books) {
      books = []
      bookWorldBooks.value.set(bookId, books)
    }
    let wb = books.find(b => b.name === wbName)
    if (!wb) {
      wb = {
        entries: [],
        name: wbName,
        source: 'book',
        bookId
      }
      books.push(wb)
      bookWorldBooks.value.set(bookId, [...books])
    }
    return wb
  }

  const saveCurrentNames = () => {
    try {
      const obj: Record<string, string> = {}
      currentWorldBookNames.value.forEach((v, k) => { obj[k] = v })
      localStorage.setItem(STORAGE_KEY_CURRENT, JSON.stringify(obj))
    } catch {}
  }

  // --- 持久化 ---

  const loadFromStorage = () => {
    try {
      // 加载全局世界书
      const globalData = localStorage.getItem(STORAGE_KEY_GLOBAL)
      if (globalData) {
        const parsed = JSON.parse(globalData)
        globalWorldBook.value = { ...globalWorldBook.value, ...parsed }
      }

      // 加载书籍世界书（兼容旧格式）
      const booksData = localStorage.getItem(STORAGE_KEY_BOOKS)
      if (booksData) {
        const parsed = JSON.parse(booksData)
        const newMap = new Map<string, WorldBook[]>()
        for (const [key, value] of Object.entries(parsed)) {
          // 旧格式：value 是单个 WorldBook 对象 → 迁移为数组
          if (Array.isArray(value)) {
            newMap.set(key, value as WorldBook[])
          } else {
            // 旧格式迁移
            const old = value as WorldBook
            if (typeof old === 'object' && old.entries) {
              if (!old.name) old.name = DEFAULT_WB_NAME
              newMap.set(key, [old])
            }
          }
        }
        bookWorldBooks.value = newMap
      }

      // 加载当前选中世界书名称
      const currentData = localStorage.getItem(STORAGE_KEY_CURRENT)
      if (currentData) {
        currentWorldBookNames.value = new Map(Object.entries(JSON.parse(currentData)))
      }

      const enabledData = localStorage.getItem('worldbook-enabled')
      if (enabledData !== null) {
        enabled.value = JSON.parse(enabledData)
      }
    } catch (error) {
      console.error('加载世界书数据失败:', error)
    }
  }

  const saveToStorage = () => {
    try {
      localStorage.setItem(STORAGE_KEY_GLOBAL, JSON.stringify(globalWorldBook.value))
      
      const booksObj: Record<string, WorldBook[]> = {}
      bookWorldBooks.value.forEach((value, key) => {
        booksObj[key] = value
      })
      localStorage.setItem(STORAGE_KEY_BOOKS, JSON.stringify(booksObj))

      saveCurrentNames()
      
      localStorage.setItem('worldbook-enabled', JSON.stringify(enabled.value))
    } catch (error) {
      console.error('保存世界书数据失败:', error)
    }
  }

  // --- 世界书管理 ---

  const getBookWorldBookList = (bookId: string): string[] => {
    const books = bookWorldBooks.value.get(bookId)
    return books ? books.map(b => b.name) : []
  }

  const getCurrentWorldBookName = (bookId: string): string => {
    return ensureCurrentName(bookId)
  }

  const setCurrentWorldBookName = (bookId: string, name: string) => {
    currentWorldBookNames.value.set(bookId, name)
    saveCurrentNames()
  }

  const getBookWorldBook = (bookId: string, name?: string): WorldBook | undefined => {
    const books = bookWorldBooks.value.get(bookId)
    if (!books) return undefined
    return books.find(b => b.name === (name || ensureCurrentName(bookId)))
  }

  const createWorldBook = (bookId: string, name: string): WorldBook => {
    const books = bookWorldBooks.value.get(bookId) || []
    const existing = books.find(b => b.name === name)
    if (existing) return existing
    const wb: WorldBook = { entries: [], name, source: 'book', bookId }
    books.push(wb)
    bookWorldBooks.value.set(bookId, [...books])
    setCurrentWorldBookName(bookId, name)
    saveToStorage()
    return wb
  }

  const renameWorldBook = (bookId: string, oldName: string, newName: string): boolean => {
    const books = bookWorldBooks.value.get(bookId)
    if (!books) return false
    if (books.find(b => b.name === newName)) return false
    const wb = books.find(b => b.name === oldName)
    if (!wb) return false
    wb.name = newName
    // 触发 Map 响应式更新
    bookWorldBooks.value.set(bookId, [...books])
    if (currentWorldBookNames.value.get(bookId) === oldName) {
      setCurrentWorldBookName(bookId, newName)
    }
    saveToStorage()
    return true
  }

  const deleteWorldBook = (bookId: string, name: string): boolean => {
    const books = bookWorldBooks.value.get(bookId)
    if (!books) return false
    const idx = books.findIndex(b => b.name === name)
    if (idx === -1) return false
    books.splice(idx, 1)
    if (books.length === 0) {
      bookWorldBooks.value.delete(bookId)
    } else {
      bookWorldBooks.value.set(bookId, [...books])
    }
    if (currentWorldBookNames.value.get(bookId) === name) {
      const remaining = getBookWorldBookList(bookId)
      setCurrentWorldBookName(bookId, remaining[0] || DEFAULT_WB_NAME)
    }
    saveToStorage()
    return true
  }

  // --- 条目操作 ---

  const addEntry = (entry?: Partial<WorldBookEntry>, bookId?: string, worldBookName?: string): WorldBookEntry => {
    const newEntry = {
      ...createDefaultWorldBookEntry(),
      ...entry,
      uid: generateUID()
    }

    if (bookId) {
      const wb = ensureWorldBook(bookId, worldBookName)
      wb.entries.push(newEntry)
      markDirty(bookId)
    } else {
      globalWorldBook.value.entries.push(newEntry)
    }

    saveToStorage()
    return newEntry
  }

  const updateEntry = (uid: string, updates: Partial<WorldBookEntry>, bookId?: string, worldBookName?: string) => {
    if (bookId) {
      const wb = getBookWorldBook(bookId, worldBookName)
      if (!wb) return
      const index = wb.entries.findIndex(e => e.uid === uid)
      if (index !== -1) {
        wb.entries[index] = { ...wb.entries[index], ...updates }
        markDirty(bookId)
        saveToStorage()
      }
    } else {
      const index = globalWorldBook.value.entries.findIndex(e => e.uid === uid)
      if (index !== -1) {
        globalWorldBook.value.entries[index] = { ...globalWorldBook.value.entries[index], ...updates }
        saveToStorage()
      }
    }
  }

  const deleteEntry = (uid: string, bookId?: string, worldBookName?: string) => {
    if (bookId) {
      const wb = getBookWorldBook(bookId, worldBookName)
      if (!wb) return
      const index = wb.entries.findIndex(e => e.uid === uid)
      if (index !== -1) {
        wb.entries.splice(index, 1)
        markDirty(bookId)
        saveToStorage()
      }
    } else {
      const index = globalWorldBook.value.entries.findIndex(e => e.uid === uid)
      if (index !== -1) {
        globalWorldBook.value.entries.splice(index, 1)
        saveToStorage()
      }
    }
  }

  const getEntry = (uid: string, bookId?: string, worldBookName?: string): WorldBookEntry | undefined => {
    if (bookId) {
      const wb = getBookWorldBook(bookId, worldBookName)
      return wb?.entries.find(e => e.uid === uid)
    }
    return globalWorldBook.value.entries.find(e => e.uid === uid)
  }

  const toggleEnabled = (value: boolean) => {
    enabled.value = value
    saveToStorage()
  }

  const importWorldBookData = (data: any, bookId?: string, worldBookName?: string): boolean => {
    try {
      let entriesInput: any = null

      if (data.entries) {
        entriesInput = data.entries
      } else if (Array.isArray(data)) {
        entriesInput = data
      }

      if (!entriesInput) {
        return false
      }

      let entriesArray: any[]
      if (Array.isArray(entriesInput)) {
        entriesArray = entriesInput
      } else if (typeof entriesInput === 'object') {
        entriesArray = Object.values(entriesInput).filter((entry: any) => entry && typeof entry === 'object')
      } else {
        return false
      }

      if (!Array.isArray(entriesArray) || entriesArray.length === 0) {
        return false
      }

      const entries: WorldBookEntry[] = entriesArray.map((entry: any) => ({
        ...createDefaultWorldBookEntry(),
        ...entry,
        uid: entry.uid || generateUID()
      }))

      if (bookId) {
        const wbName = worldBookName || data.name || ensureCurrentName(bookId)
        // 如果导入数据带有名称且与现有不同，则创建新世界书
        if (data.name && data.name !== (worldBookName || ensureCurrentName(bookId))) {
          const existing = getBookWorldBook(bookId, data.name)
          if (existing) {
            existing.entries.push(...entries)
          } else {
            createWorldBook(bookId, data.name)
            const wb = getBookWorldBook(bookId, data.name)
            if (wb) wb.entries.push(...entries)
          }
        } else {
          const wb = ensureWorldBook(bookId, wbName)
          wb.entries.push(...entries)
        }
        markDirty(bookId)
      } else {
        globalWorldBook.value.entries.push(...entries)
      }

      saveToStorage()
      return true
    } catch (error) {
      console.error('导入世界书失败:', error)
      return false
    }
  }

  const exportWorldBookData = (bookId?: string, format: 'default' | 'sillytavern' = 'default', worldBookName?: string) => {
    const targetBook = bookId 
      ? getBookWorldBook(bookId, worldBookName)
      : globalWorldBook.value

    if (!targetBook) return null

    if (format === 'sillytavern') {
      const entriesMap: Record<string, WorldBookEntry> = {}
      for (const entry of targetBook.entries) {
        entriesMap[entry.uid] = entry
      }
      return {
        entries: entriesMap,
        name: targetBook.name,
        exportDate: new Date().toISOString(),
        version: '1.0.0'
      }
    }

    return {
      entries: targetBook.entries,
      name: targetBook.name,
      exportDate: new Date().toISOString(),
      version: '1.0.0'
    }
  }

  const clearWorldBook = (bookId?: string, worldBookName?: string) => {
    if (bookId) {
      const wb = getBookWorldBook(bookId, worldBookName)
      if (wb) wb.entries = []
    } else {
      globalWorldBook.value.entries = []
    }
    saveToStorage()
  }

  loadFromStorage()

  return {
    globalWorldBook,
    bookWorldBooks,
    currentWorldBookNames,
    enabled,
    allEntries,
    enabledEntries,
    globalEntryCount,
    // 世界书管理
    getBookWorldBookList,
    getCurrentWorldBookName,
    setCurrentWorldBookName,
    getBookWorldBook,
    createWorldBook,
    renameWorldBook,
    deleteWorldBook,
    // 条目操作
    addEntry,
    updateEntry,
    deleteEntry,
    getEntry,
    toggleEnabled,
    importWorldBookData,
    exportWorldBookData,
    clearWorldBook,
    loadFromStorage,
    saveToStorage
  }
})
