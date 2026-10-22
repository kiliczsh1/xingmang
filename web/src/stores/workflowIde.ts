import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

/**
 * IDE 工作台的文件/标签页状态 —— 复刻源项目 IDELayout 所依赖的文件仓库语义
 * （openFiles / activeFileId / recentFiles / viewMode / 字号字体设置）。
 *
 * 本地化：源版文件是磁盘上的真实文件（Electron file API），星芒这里的工作区 =
 * sqlite 里的 workflows 表，每个工作流映射为一个 `<名称>.workflow.json` 文件节点。
 */
export interface IdeFile {
  id: number
  label: string
  /** 展示用路径，如 工作流/番茄改稿.workflow.json */
  fullPath: string
  language: 'json'
  isWorkflowFile: true
  /** workflow JSON 文本（源码模式的编辑对象） */
  content: string
  /** 'editor' = Monaco 源码；'visual' = Vue Flow 画布 */
  viewMode: 'editor' | 'visual'
  dirty: boolean
  loadedAt: number
}

const RECENT_KEY = 'xingmang.workflow.recent'

export const useWorkflowIdeStore = defineStore('workflowIde', () => {
  const openFiles = ref<IdeFile[]>([])
  const activeFileId = ref<number | null>(null)
  const recentFiles = ref<Array<{ id: number; name: string; fullPath: string; lastOpened: number }>>(
    JSON.parse(localStorage.getItem(RECENT_KEY) || '[]')
  )
  /** 左栏宽度 / 右栏宽度（源版默认 260 / 480） */
  const leftSidebarWidth = ref(260)
  const rightSidebarWidth = ref(480)
  const leftSidebarVisible = ref(true)
  const logsPanelVisible = ref(true)
  const editorFontSize = ref(14)
  const editorFontFamily = ref("Consolas, 'Courier New', monospace")

  const activeFile = computed<IdeFile | null>(() => {
    if (!activeFileId.value || openFiles.value.length === 0) return null
    return openFiles.value.find(f => f.id === activeFileId.value) || null
  })

  const isWorkflowFileActive = computed(() => {
    const f = activeFile.value
    return !!f && (f.isWorkflowFile === true || f.label.endsWith('.workflow.json'))
  })

  const persistRecent = () => {
    localStorage.setItem(RECENT_KEY, JSON.stringify(recentFiles.value.slice(0, 12)))
  }

  const pushRecent = (file: { id: number; name: string; fullPath: string }) => {
    recentFiles.value = [
      { id: file.id, name: file.name, fullPath: file.fullPath, lastOpened: Date.now() },
      ...recentFiles.value.filter(r => r.id !== file.id)
    ].slice(0, 12)
    persistRecent()
  }

  const removeRecent = (id: number) => {
    recentFiles.value = recentFiles.value.filter(r => r.id !== id)
    persistRecent()
  }

  const openFile = (file: IdeFile) => {
    if (!openFiles.value.find(f => f.id === file.id)) openFiles.value.push(file)
    activeFileId.value = file.id
    pushRecent({ id: file.id, name: file.label, fullPath: file.fullPath })
  }

  const closeFile = (id: number) => {
    const idx = openFiles.value.findIndex(f => f.id === id)
    if (idx === -1) return
    openFiles.value.splice(idx, 1)
    if (activeFileId.value === id) {
      const next = openFiles.value[idx] || openFiles.value[idx - 1]
      activeFileId.value = next ? next.id : null
    }
  }

  const updateFileContent = (id: number, content: string) => {
    const f = openFiles.value.find(x => x.id === id)
    if (f) {
      f.content = content
      f.dirty = true
    }
  }

  const touchFile = (id: number) => {
    const f = openFiles.value.find(x => x.id === id)
    if (f) f.dirty = true
  }

  const markSaved = (id: number) => {
    const f = openFiles.value.find(x => x.id === id)
    if (f) f.dirty = false
  }

  const setViewMode = (id: number, mode: 'editor' | 'visual') => {
    const f = openFiles.value.find(x => x.id === id)
    if (f) f.viewMode = mode
  }

  return {
    openFiles,
    activeFileId,
    activeFile,
    isWorkflowFileActive,
    recentFiles,
    leftSidebarWidth,
    rightSidebarWidth,
    leftSidebarVisible,
    logsPanelVisible,
    editorFontSize,
    editorFontFamily,
    openFile,
    closeFile,
    updateFileContent,
    touchFile,
    markSaved,
    setViewMode,
    pushRecent,
    removeRecent
  }
})
