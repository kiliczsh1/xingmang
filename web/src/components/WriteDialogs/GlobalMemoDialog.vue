<template>
  <el-dialog
    v-model="visible"
    title="全局备忘录"
    width="1200px"
    top="5vh"
    draggable
    overflow
    append-to-body
    :modal="false"
    class="global-memo-dialog"
    destroy-on-close
  >
    <div class="global-memo-container">
      <!-- ===== 主画布：编辑器常驻 ===== -->
      <div class="memo-canvas">
        <div v-if="dialogSelectedMemo" class="content-area">
          <div class="memo-content-header">
            <span
              class="memo-scope-badge"
              :class="isBookMemo(dialogSelectedMemo) ? 'book' : 'global'"
            >
              {{ isBookMemo(dialogSelectedMemo) ? '本书备忘录' : '全局备忘录' }}
            </span>
            <div class="header-actions">
              <button
                type="button"
                class="icon-btn"
                :class="{ pinned: dialogSelectedMemo.is_pinned }"
                title="置顶"
                @click="toggleMemoPin(dialogSelectedMemo)"
              >
                <el-icon><Star /></el-icon>
              </button>
              <button
                type="button"
                class="icon-btn danger"
                title="删除"
                @click="deleteMemo(dialogSelectedMemo)"
              >
                <el-icon><Delete /></el-icon>
              </button>
            </div>
          </div>
          <el-input
            v-model="dialogSelectedMemo.title"
            placeholder="请输入备忘录标题"
            class="memo-title-input"
          />
          <div class="memo-content-meta">
            <span>创建 {{ formatExactTime(dialogSelectedMemo.created_at) }}</span>
            <span>更新 {{ formatExactTime(dialogSelectedMemo.updated_at) }}</span>
            <span>{{ getContentLength(dialogSelectedMemo.content) }} 字</span>
          </div>
          <el-input
            v-model="dialogSelectedMemo.content"
            type="textarea"
            placeholder="在这里输入内容..."
            class="memo-content-input memo-content-input-large"
          />
          <div class="content-footer">
            <select
              v-model="editorMoveFolder"
              class="folder-select"
              title="移动到文件夹"
              @change="moveFromEditor"
            >
              <option v-for="folder in globalMemoFolders" :key="folder" :value="folder">
                {{ folder }}
              </option>
            </select>
            <div class="footer-actions">
              <el-button @click="$emit('jumpToMemo', dialogSelectedMemo)">跳转编辑</el-button>
              <el-button type="primary" @click="saveDialogMemo">保存</el-button>
            </div>
          </div>
        </div>
        <div v-else class="content-empty">
          <el-icon><EditPen /></el-icon>
          <span>选择左侧备忘录，或点击抽屉「＋」新建</span>
        </div>
      </div>

      <!-- ===== 抽屉拉手 ===== -->
      <div v-if="!drawerVisible" class="drawer-tab" @click="drawerVisible = true">
        <span class="drawer-tab-text">备忘录</span>
        <span class="drawer-tab-count">{{ props.memos.length }}</span>
      </div>

      <!-- ===== 抽屉 ===== -->
      <div class="memo-drawer" :class="{ open: drawerVisible }">
        <div class="drawer-head">
          <div class="drawer-title">
            <el-icon><Notebook /></el-icon>
            <span>备忘录</span>
          </div>
          <button type="button" class="icon-btn" title="关闭" @click="drawerVisible = false">
            <el-icon><Close /></el-icon>
          </button>
        </div>

        <div class="drawer-tools">
          <div class="memo-search-box">
            <el-input
              v-model="memoSearchKeyword"
              placeholder="搜索备忘录..."
              clearable
              prefix-icon="Search"
              size="small"
            />
          </div>
          <button type="button" class="new-btn" title="新建备忘录" @click="createMemoFromDrawer">
            <el-icon><Plus /></el-icon>
          </button>
        </div>

        <div class="drawer-tabs">
          <button
            type="button"
            class="d-tab"
            :class="{ active: memoScopeTab === 'all' }"
            @click="memoScopeTab = 'all'"
          >
            全部
          </button>
          <button
            type="button"
            class="d-tab"
            :class="{ active: memoScopeTab === 'global' }"
            @click="memoScopeTab = 'global'"
          >
            全局
          </button>
          <button
            type="button"
            class="d-tab"
            :class="{ active: memoScopeTab === 'book' }"
            @click="memoScopeTab = 'book'"
          >
            本书
          </button>
        </div>

        <!-- 全局文件夹筛选 -->
        <div v-if="memoScopeTab === 'global'" class="folder-chips">
          <button
            type="button"
            class="folder-chip"
            :class="{ active: currentFolderFilter === '全部' }"
            @click="currentFolderFilter = '全部'"
          >
            全部
          </button>
          <button
            v-for="folder in globalMemoFolders"
            :key="folder"
            type="button"
            class="folder-chip"
            :class="{ active: currentFolderFilter === folder }"
            @click="currentFolderFilter = folder"
          >
            {{ folder }}
            <span class="folder-chip-count">{{ getFolderMemoCount(folder) }}</span>
          </button>
          <button
            type="button"
            class="folder-chip add"
            title="新建文件夹"
            @click="handleCreateGlobalMemoFolder"
          >
            <el-icon><Plus /></el-icon>
          </button>
        </div>

        <!-- 批量工具条 -->
        <div v-if="memoBatchMode" class="batch-toolbar">
          <el-checkbox
            :model-value="hasSelectedMemos && drawerMemos.length > 0 && selectedMemoIds.length === drawerMemos.length"
            :indeterminate="hasSelectedMemos && selectedMemoIds.length < drawerMemos.length"
            @change="toggleSelectAll"
          >
            全选
          </el-checkbox>
          <el-button size="small" type="danger" :disabled="!hasSelectedMemos" @click="batchDelete">
            删除 ({{ selectedMemoIds.length }})
          </el-button>
          <el-button size="small" :disabled="!hasSelectedMemos" @click="batchTogglePin">
            置顶 ({{ selectedMemoIds.length }})
          </el-button>
        </div>

        <div class="drawer-list">
          <div v-if="drawerMemos.length === 0" class="drawer-empty">
            <el-icon><Document /></el-icon>
            <span>暂无备忘录</span>
          </div>
          <div v-else>
            <div
              v-for="memo in drawerMemos"
              :key="memo.id"
              :class="['drawer-item', { active: dialogSelectedMemo?.id === memo.id, 'batch-selected': selectedMemoIds.includes(memo.id) }]"
              @click="handleMemoClick(memo)"
            >
              <div v-if="memoBatchMode" class="batch-checkbox">
                <el-checkbox
                  :model-value="selectedMemoIds.includes(memo.id)"
                  @click.stop
                  @change="() => toggleMemoSelection(memo.id)"
                />
              </div>
              <div class="item-content">
                <div class="item-title-row">
                  <span class="item-title">{{ memo.title || '无标题' }}</span>
                  <span v-if="memo.is_pinned && !memoBatchMode" class="pin-star">
                    <el-icon><Star /></el-icon>
                  </span>
                </div>
                <div class="item-snippet">{{ memo.content }}</div>
                <div class="item-meta-row">
                  <span v-if="memo.tags" class="item-tags">
                    <el-tag
                      v-for="tag in getMemoTags(memo.tags)"
                      :key="tag"
                      size="small"
                      class="item-tag"
                    >
                      {{ tag }}
                    </el-tag>
                  </span>
                  <span class="item-meta-time">{{ formatTime(memo.updated_at) }}</span>
                </div>
              </div>
              <div class="item-quick-actions">
                <button
                  v-if="!memoBatchMode"
                  type="button"
                  class="item-icon-btn"
                  title="移动到文件夹"
                  @click.stop="openMoveFolderDialog(memo)"
                >
                  <el-icon><Folder /></el-icon>
                </button>
                <button
                  v-if="!memoBatchMode"
                  type="button"
                  class="item-icon-btn pin"
                  :class="{ on: memo.is_pinned }"
                  title="置顶"
                  @click.stop="toggleMemoPin(memo)"
                >
                  <el-icon><Star /></el-icon>
                </button>
                <button
                  v-if="!memoBatchMode"
                  type="button"
                  class="item-icon-btn danger"
                  title="删除"
                  @click.stop="deleteMemoInList(memo)"
                >
                  <el-icon><Delete /></el-icon>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="drawer-foot">
          <button
            type="button"
            class="batch-toggle"
            :class="{ active: memoBatchMode }"
            @click="toggleBatchMode"
          >
            批量
          </button>
          <button type="button" class="refresh-btn" title="刷新" @click="$emit('changed')">
            <el-icon><Refresh /></el-icon>
          </button>
        </div>
      </div>

      <!-- 对话框缩放手柄 -->
      <div class="dialog-resize-handle" @mousedown="startDialogResize"></div>
    </div>

    <!-- 移动文件夹对话框 -->
    <el-dialog
      v-model="moveFolderDialogVisible"
      title="移动到文件夹"
      width="400px"
      append-to-body
    >
      <div class="move-folder-content">
        <p class="move-folder-hint">选择目标文件夹：</p>
        <div class="folder-list">
          <button
            v-for="folder in globalMemoFolders"
            :key="folder"
            class="folder-option"
            @click="moveToFolder(folder)"
          >
            <el-icon><Folder /></el-icon>
            <span>{{ folder }}</span>
            <span class="folder-option-count">{{ getFolderMemoCount(folder) }}</span>
          </button>
        </div>
      </div>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Notebook, Plus, Delete, Folder, Close, EditPen, Star, Refresh, Document
} from '@element-plus/icons-vue'
import { memoAPI } from '@/api'
import type { Memo } from '@/types'
import { getContentLength, formatTime, formatExactTime } from '@/utils/format'

const props = defineProps<{
  bookId: number
  memos: Memo[]
}>()

const visible = defineModel<boolean>({ default: false })

const emit = defineEmits<{
  changed: []
  jumpToMemo: [memo: Memo]
}>()

// ===== 内部状态 =====
const dialogSelectedMemo = ref<Memo | null>(null)
const memoSearchKeyword = ref('')
const memoBatchMode = ref(false)
const selectedMemoIds = ref<number[]>([])
const dialogSelectedMemoTags = ref<string[]>([])
const globalMemoCustomFolders = ref<string[]>([])
const moveFolderDialogVisible = ref(false)
const memoToMove = ref<Memo | null>(null)

// ===== 抽屉状态 =====
const drawerVisible = ref(true)
const memoScopeTab = ref<'all' | 'global' | 'book'>('all')
const currentFolderFilter = ref('全部')
const editorMoveFolder = ref('默认')

// ===== 工具函数 =====
const getMemoTags = (tagsStr: string) => {
  try {
    const tags = JSON.parse(tagsStr)
    return Array.isArray(tags) ? tags.slice(0, 3) : []
  } catch (e) {
    return []
  }
}

const isBookMemo = (memo: Memo) => {
  return (memo.category || '').includes('本书')
}

const getGlobalMemoFolder = (memo: Memo) => {
  if (isBookMemo(memo)) return ''
  const category = (memo.category || '').trim()
  if (!category || category === '全局') return '默认'
  return category
}

// ===== 计算属性 =====
const filteredMemos = computed(() => {
  if (!memoSearchKeyword.value) {
    return props.memos
  }
  const keyword = memoSearchKeyword.value.toLowerCase()
  return props.memos.filter(memo =>
    memo.title.toLowerCase().includes(keyword) ||
    memo.content.toLowerCase().includes(keyword)
  )
})

const globalMemoList = computed(() => {
  return filteredMemos.value.filter(memo => !isBookMemo(memo))
})

const bookMemoList = computed(() => {
  return filteredMemos.value.filter(memo => isBookMemo(memo))
})

const globalMemoFolders = computed(() => {
  const folderSet = new Set<string>(['默认'])
  globalMemoCustomFolders.value.forEach(folder => folderSet.add(folder))
  globalMemoList.value.forEach(memo => folderSet.add(getGlobalMemoFolder(memo)))
  return Array.from(folderSet)
})

const hasSelectedMemos = computed(() => selectedMemoIds.value.length > 0)

const drawerMemos = computed(() => {
  let list = filteredMemos.value
  if (memoScopeTab.value === 'global') {
    list = list.filter(memo => !isBookMemo(memo))
    if (currentFolderFilter.value !== '全部') {
      list = list.filter(memo => getGlobalMemoFolder(memo) === currentFolderFilter.value)
    }
  } else if (memoScopeTab.value === 'book') {
    list = list.filter(memo => isBookMemo(memo))
  }
  return list
})

// ===== 文件夹持久化 =====
const GLOBAL_MEMO_FOLDER_STORAGE_KEY = 'write-global-memo-folders'

const loadGlobalMemoFolders = () => {
  try {
    const raw = localStorage.getItem(GLOBAL_MEMO_FOLDER_STORAGE_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) {
      globalMemoCustomFolders.value = parsed.filter(folder => typeof folder === 'string' && folder.trim())
    }
  } catch (error) {
    globalMemoCustomFolders.value = []
  }
}

const persistGlobalMemoFolders = () => {
  localStorage.setItem(GLOBAL_MEMO_FOLDER_STORAGE_KEY, JSON.stringify(globalMemoCustomFolders.value))
}

// ===== 文件夹操作 =====
const getMemosByFolder = (folder: string) => {
  return globalMemoList.value.filter(memo => getGlobalMemoFolder(memo) === folder)
}

const getFolderMemoCount = (folder: string) => {
  return getMemosByFolder(folder).length
}

const handleCreateGlobalMemoFolder = async () => {
  try {
    const { value } = await ElMessageBox.prompt('请输入全局备忘录文件夹名称', '新建文件夹', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPlaceholder: '例如：灵感池',
      inputValidator: (input) => {
        const name = input.trim()
        if (!name) return '请输入文件夹名称'
        if (globalMemoFolders.value.includes(name)) return '文件夹已存在'
        return true
      }
    })

    const folderName = value.trim()
    globalMemoCustomFolders.value.push(folderName)
    persistGlobalMemoFolders()
    currentFolderFilter.value = folderName
    memoScopeTab.value = 'global'
    ElMessage.success('文件夹创建成功')
  } catch (error) {
    // 用户取消
  }
}

const deleteMemoFolder = async (folder: string) => {
  try {
    await ElMessageBox.confirm(`确定要删除文件夹"${folder}"吗？该文件夹下的所有备忘录将被移动到"默认"文件夹。`, '删除文件夹', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const memosInFolder = getMemosByFolder(folder)
    for (const memo of memosInFolder) {
      await memoAPI.update(memo.id, {
        category: '默认'
      })
    }

    const index = globalMemoCustomFolders.value.indexOf(folder)
    if (index !== -1) {
      globalMemoCustomFolders.value.splice(index, 1)
    }
    persistGlobalMemoFolders()

    if (currentFolderFilter.value === folder) {
      currentFolderFilter.value = '全部'
    }

    ElMessage.success('文件夹已删除')
    emit('changed')
  } catch (error) {
    // 用户取消
  }
}

// ===== 备忘录 CRUD =====
const createMemoFromDrawer = async () => {
  let scope: 'global' | 'book' = 'global'
  let folder = '默认'
  if (memoScopeTab.value === 'book') {
    scope = 'book'
  } else {
    scope = 'global'
    folder = currentFolderFilter.value !== '全部' ? currentFolderFilter.value : '默认'
  }
  const res = await memoAPI.create({
    title: '新备忘录',
    content: '',
    category: scope === 'book' ? '本书' : folder === '默认' ? '全局' : folder,
    order_num: props.memos.length
  })
  if (res.success && res.data) {
    selectDialogMemo(res.data)
    emit('changed')
  }
}

const selectDialogMemo = (memo: Memo) => {
  if (!isBookMemo(memo)) {
    currentFolderFilter.value = getGlobalMemoFolder(memo)
  }
  editorMoveFolder.value = isBookMemo(memo) ? '默认' : getGlobalMemoFolder(memo)
  dialogSelectedMemo.value = { ...memo }
  if (memo.tags) {
    try {
      const tags = JSON.parse(memo.tags)
      if (Array.isArray(tags)) {
        dialogSelectedMemoTags.value = [...tags]
      } else {
        dialogSelectedMemoTags.value = []
      }
    } catch (e) {
      dialogSelectedMemoTags.value = []
    }
  } else {
    dialogSelectedMemoTags.value = []
  }
}

const handleMemoClick = (memo: Memo) => {
  if (memoBatchMode.value) {
    toggleMemoSelection(memo.id)
  } else {
    selectDialogMemo(memo)
  }
}

const saveDialogMemo = async () => {
  if (!dialogSelectedMemo.value) return

  const res = await memoAPI.update(dialogSelectedMemo.value.id, {
    title: dialogSelectedMemo.value.title,
    content: dialogSelectedMemo.value.content,
    tags: JSON.stringify(dialogSelectedMemoTags.value)
  })

  if (res.success) {
    ElMessage.success('保存成功')
    emit('changed')
  }
}

const deleteMemoInList = async (memo: Memo) => {
  try {
    await ElMessageBox.confirm(`确定要删除备忘录"${memo.title || '无标题'}"吗？`, '删除备忘录', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const res = await memoAPI.delete(memo.id)
    if (res.success) {
      if (dialogSelectedMemo.value?.id === memo.id) {
        dialogSelectedMemo.value = null
      }
      ElMessage.success('备忘录已删除')
      emit('changed')
    }
  } catch (error) {
    // 用户取消
  }
}

const deleteMemo = async (memo: Memo) => {
  try {
    await ElMessageBox.confirm(`确定删除备忘录"${memo.title}"吗？`, '提示', {
      type: 'warning'
    })
    const res = await memoAPI.delete(memo.id)
    if (res.success) {
      if (dialogSelectedMemo.value?.id === memo.id) {
        dialogSelectedMemo.value = null
      }
      ElMessage.success('删除成功')
      emit('changed')
    }
  } catch (error) {
    // 取消删除
  }
}

const toggleMemoPin = async (memo: Memo) => {
  const nextPinned = memo.is_pinned ? 0 : 1
  const res = await memoAPI.update(memo.id, { is_pinned: nextPinned })
  if (res.success) {
    if (dialogSelectedMemo.value?.id === memo.id) {
      dialogSelectedMemo.value.is_pinned = nextPinned
    }
    ElMessage.success(nextPinned ? '已置顶' : '已取消置顶')
    emit('changed')
  }
}

// ===== 批量操作 =====
const toggleBatchMode = () => {
  memoBatchMode.value = !memoBatchMode.value
  if (!memoBatchMode.value) {
    selectedMemoIds.value = []
  }
}

const toggleMemoSelection = (memoId: number) => {
  const index = selectedMemoIds.value.indexOf(memoId)
  if (index > -1) {
    selectedMemoIds.value.splice(index, 1)
  } else {
    selectedMemoIds.value.push(memoId)
  }
}

const toggleSelectAll = (checked: boolean) => {
  if (checked) {
    selectedMemoIds.value = drawerMemos.value.map(m => m.id)
  } else {
    selectedMemoIds.value = []
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定删除选中的 ${selectedMemoIds.value.length} 个备忘录吗？`, '提示', {
      type: 'warning'
    })
    const res = await memoAPI.batch({
      action: 'delete',
      ids: selectedMemoIds.value
    })
    if (res.success) {
      selectedMemoIds.value = []
      memoBatchMode.value = false
      ElMessage.success('批量删除成功')
      emit('changed')
    }
  } catch (error) {
    // 取消删除
  }
}

const batchTogglePin = async () => {
  const res = await memoAPI.batch({
    action: 'toggle_pin',
    ids: selectedMemoIds.value
  })
  if (res.success) {
    selectedMemoIds.value = []
    memoBatchMode.value = false
    ElMessage.success('批量操作成功')
    emit('changed')
  }
}

// ===== 移动备忘录 =====
const openMoveFolderDialog = (memo: Memo) => {
  memoToMove.value = memo
  moveFolderDialogVisible.value = true
}

const moveToFolder = async (folder: string) => {
  if (!memoToMove.value) return

  const res = await memoAPI.update(memoToMove.value.id, {
    title: memoToMove.value.title,
    content: memoToMove.value.content,
    category: folder === '默认' ? '全局' : folder,
    tags: memoToMove.value.tags
  })

  if (res.success) {
    ElMessage.success('备忘录已移动')
    moveFolderDialogVisible.value = false
    memoToMove.value = null
    emit('changed')
  }
}

const moveFromEditor = async () => {
  if (!dialogSelectedMemo.value) return
  const folder = editorMoveFolder.value

  const res = await memoAPI.update(dialogSelectedMemo.value.id, {
    title: dialogSelectedMemo.value.title,
    content: dialogSelectedMemo.value.content,
    category: folder === '默认' ? '全局' : folder,
    tags: dialogSelectedMemo.value.tags
  })

  if (res.success) {
    ElMessage.success('备忘录已移动')
    emit('changed')
  }
}

// ===== 对话框缩放 =====
let isDialogResizing = false
let dialogResizeStartX = 0
let dialogResizeStartY = 0
let dialogResizeStartWidth = 0
let dialogResizeStartHeight = 0
let dialogResizeStartLeft = 0
let dialogResizeStartTop = 0
let dialogResizeStartTranslateX = 0
let dialogResizeStartTranslateY = 0
let memoDialogElement: HTMLElement | null = null
let dialogResizeFrame = 0
let pendingDialogResizeX = 0
let pendingDialogResizeY = 0

const getElementTranslate = (element: HTMLElement) => {
  const transform = window.getComputedStyle(element).transform
  if (!transform || transform === 'none') {
    return { x: 0, y: 0 }
  }
  const matrix = new DOMMatrix(transform)
  return {
    x: matrix.m41,
    y: matrix.m42
  }
}

const stabilizeGlobalMemoDialogPosition = async () => {
  await nextTick()
  const dialog = document.querySelector('.global-memo-dialog') as HTMLElement | null
  if (!dialog) return
  const rect = dialog.getBoundingClientRect()
  dialog.style.position = 'fixed'
  dialog.style.margin = '0'
  dialog.style.left = `${rect.left}px`
  dialog.style.top = `${rect.top}px`
}

const startDialogResize = (e: MouseEvent) => {
  isDialogResizing = true
  dialogResizeStartX = e.clientX
  dialogResizeStartY = e.clientY
  pendingDialogResizeX = e.clientX
  pendingDialogResizeY = e.clientY

  memoDialogElement = (e.currentTarget as HTMLElement | null)?.closest('.el-dialog') as HTMLElement | null
    ?? document.querySelector('.global-memo-dialog')
  if (memoDialogElement) {
    const rect = memoDialogElement.getBoundingClientRect()
    const translate = getElementTranslate(memoDialogElement)
    memoDialogElement.style.position = 'fixed'
    memoDialogElement.style.margin = '0'
    memoDialogElement.style.left = `${rect.left}px`
    memoDialogElement.style.top = `${rect.top}px`
    dialogResizeStartWidth = memoDialogElement.offsetWidth
    dialogResizeStartHeight = memoDialogElement.offsetHeight
    dialogResizeStartLeft = rect.left
    dialogResizeStartTop = rect.top
    dialogResizeStartTranslateX = translate.x
    dialogResizeStartTranslateY = translate.y
    memoDialogElement.style.transition = 'none'
    memoDialogElement.style.willChange = 'width, transform'
  }

  document.addEventListener('mousemove', handleDialogResize)
  document.addEventListener('mouseup', stopDialogResize)
  document.body.style.userSelect = 'none'
  document.body.style.cursor = 'se-resize'
  e.preventDefault()
}

const applyDialogResize = () => {
  if (!isDialogResizing || !memoDialogElement) return

  dialogResizeFrame = 0

  const deltaX = pendingDialogResizeX - dialogResizeStartX
  const deltaY = pendingDialogResizeY - dialogResizeStartY

  const screenWidth = window.innerWidth
  const screenHeight = window.innerHeight

  const minWidth = Math.max(600, screenWidth * 0.1)
  const maxWidth = Math.max(minWidth, screenWidth - Math.max(dialogResizeStartLeft, 0) - 24)
  const minHeight = Math.max(400, screenHeight * 0.1)
  const maxHeight = Math.max(minHeight, screenHeight - Math.max(dialogResizeStartTop, 0) - 24)

  const newWidth = Math.min(maxWidth, Math.max(minWidth, dialogResizeStartWidth + deltaX))
  const newHeight = Math.min(maxHeight, Math.max(minHeight, dialogResizeStartHeight + deltaY))

  memoDialogElement.style.width = newWidth + 'px'
  memoDialogElement.style.height = 'auto'

  const container = memoDialogElement.querySelector('.global-memo-container')
  if (container) {
    (container as HTMLElement).style.height = (newHeight - 120) + 'px'
  }
}

const handleDialogResize = (e: MouseEvent) => {
  if (!isDialogResizing || !memoDialogElement) return

  pendingDialogResizeX = e.clientX
  pendingDialogResizeY = e.clientY

  if (!dialogResizeFrame) {
    dialogResizeFrame = requestAnimationFrame(applyDialogResize)
  }
}

const stopDialogResize = () => {
  isDialogResizing = false
  if (dialogResizeFrame) {
    cancelAnimationFrame(dialogResizeFrame)
    dialogResizeFrame = 0
  }
  if (memoDialogElement) {
    memoDialogElement.style.willChange = ''
    memoDialogElement.style.transition = ''
  }
  document.body.style.userSelect = ''
  document.body.style.cursor = ''
  document.removeEventListener('mousemove', handleDialogResize)
  document.removeEventListener('mouseup', stopDialogResize)
}

// ===== Watchers =====
watch(visible, (val) => {
  if (val) {
    drawerVisible.value = true
    void stabilizeGlobalMemoDialogPosition()
  }
})

watch(globalMemoCustomFolders, () => {
  persistGlobalMemoFolders()
}, { deep: true })

// ===== 初始化 =====
loadGlobalMemoFolders()
</script>

<style scoped>
.global-memo-container {
  display: flex;
  height: 640px;
  min-height: 440px;
  border: 1px solid rgba(228, 213, 178, 0.45);
  border-radius: 14px;
  overflow: hidden;
  background: #fff;
  position: relative;
}

.global-memo-dialog :deep(.el-dialog) {
  margin: 0 !important;
}

/* ===== 主画布 ===== */
.memo-canvas {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: linear-gradient(180deg, #ffffff 0%, #fcfdff 100%);
}

.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px 26px;
}

.memo-content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eef1f5;
}

.memo-scope-badge {
  display: inline-flex;
  align-items: center;
  padding: 5px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.memo-scope-badge.global {
  color: #11a683;
  background: #e8faf4;
}

.memo-scope-badge.book {
  color: #4d86ff;
  background: #edf3ff;
}

.header-actions {
  display: flex;
  gap: 6px;
}

.icon-btn {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 9px;
  background: #f4f6fa;
  color: #556274;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.icon-btn:hover {
  background: #edf3ff;
  color: #3572ff;
}

.icon-btn.danger:hover {
  background: #fff1f2;
  color: #ef4444;
}

.icon-btn.pinned {
  color: #f0a128;
}

.memo-title-input :deep(.el-input__wrapper) {
  padding-left: 0;
  box-shadow: none !important;
  background: transparent;
  border: none;
  font-size: 17px;
  font-weight: 700;
}

.memo-content-meta {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 8px;
  color: #9aa4b2;
  font-size: 12px;
  flex-wrap: wrap;
}

.memo-content-input {
  flex: 1;
  margin-top: 14px;
}

.memo-content-input-large :deep(.el-textarea__inner) {
  height: 100%;
  min-height: 360px;
  padding: 0;
  resize: none;
  line-height: 1.8;
  border: none;
  box-shadow: none;
  font-size: 15px;
  color: #667085;
  background: transparent;
}

.content-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #eef1f5;
}

.folder-select {
  background: #f4f6fa;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 7px 10px;
  font-size: 12px;
  color: #606266;
  outline: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.folder-select:hover {
  border-color: #3572ff;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.content-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #9bb8b5;
}

.content-empty .el-icon {
  font-size: 48px;
}

/* ===== 抽屉拉手 ===== */
.drawer-tab {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  z-index: 8;
  writing-mode: vertical-rl;
  background: #3572ff;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 3px;
  padding: 16px 7px;
  border-radius: 10px 0 0 10px;
  cursor: pointer;
  box-shadow: 0 10px 30px rgba(31, 45, 61, 0.12);
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
}

.drawer-tab:hover {
  filter: brightness(1.1);
}

.drawer-tab-count {
  display: inline-block;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 999px;
  padding: 1px 6px;
  font-size: 10px;
  writing-mode: horizontal-tb;
}

/* ===== 抽屉 ===== */
.memo-drawer {
  position: absolute;
  right: 0;
  top: 0;
  bottom: 0;
  width: 340px;
  background: linear-gradient(180deg, #fffdfa 0%, #ffffff 100%);
  border-left: 1px solid rgba(228, 232, 240, 0.95);
  box-shadow: 0 10px 30px rgba(31, 45, 61, 0.1);
  transform: translateX(100%);
  transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 7;
  display: flex;
  flex-direction: column;
}

.memo-drawer.open {
  transform: translateX(0);
}

.drawer-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px 10px;
}

.drawer-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #243042;
  font-size: 14px;
  font-weight: 700;
  flex: 1;
}

.drawer-title .el-icon {
  color: #4d8dff;
  font-size: 15px;
}

.drawer-tools {
  display: flex;
  gap: 8px;
  padding: 0 14px 10px;
}

.memo-search-box {
  flex: 1;
}

.new-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 10px;
  background: #3572ff;
  color: #fff;
  font-size: 16px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.new-btn:hover {
  filter: brightness(1.08);
}

.drawer-tabs {
  display: flex;
  gap: 4px;
  padding: 0 14px 8px;
}

.d-tab {
  flex: 1;
  text-align: center;
  padding: 7px 0;
  border: none;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #8a94a3;
  background: transparent;
  cursor: pointer;
  transition: all 0.15s ease;
}

.d-tab:hover {
  background: #f5f7fa;
}

.d-tab.active {
  background: #edf3ff;
  color: #3572ff;
}

.folder-chips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  padding: 0 14px 10px;
}

.folder-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  border: 1px solid #e4e7ed;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  color: #8a94a3;
  background: #fff;
  cursor: pointer;
  transition: all 0.2s ease;
}

.folder-chip:hover {
  border-color: #3572ff;
  color: #3572ff;
}

.folder-chip.active {
  background: #3572ff;
  border-color: #3572ff;
  color: #fff;
}

.folder-chip-count {
  font-weight: 400;
  opacity: 0.7;
  font-size: 10px;
}

.folder-chip.add {
  border-style: dashed;
}

.batch-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 14px 10px;
  padding: 10px 12px;
  border-radius: 12px;
  background: #fff7ef;
  border: 1px solid #ffe0bf;
}

.drawer-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 8px 12px;
}

.drawer-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 8px;
  color: #9bb8b5;
}

.drawer-empty .el-icon {
  font-size: 32px;
}

.drawer-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-bottom: 2px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.drawer-item:hover {
  background: #f5f7fa;
}

.drawer-item.active {
  background: #ecf5ff;
}

.drawer-item.active .item-title {
  color: #409eff;
  font-weight: 600;
}

.drawer-item.batch-selected {
  background: #fff7ef;
}

.batch-checkbox {
  margin-right: 4px;
}

.item-content {
  min-width: 0;
  flex: 1;
}

.item-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.item-title {
  display: block;
  font-size: 13px;
  font-weight: 400;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.pin-star {
  color: #f0a128;
  font-size: 12px;
  flex-shrink: 0;
  display: inline-flex;
}

.item-snippet {
  font-size: 11px;
  color: #909399;
  margin-top: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-top: 4px;
}

.item-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  min-width: 0;
  overflow: hidden;
}

.item-tag {
  height: 18px;
  line-height: 18px;
  font-size: 11px;
  background: #f4f4f5;
  color: #909399;
  border: none;
}

.item-meta-time {
  font-size: 10px;
  color: #b2bcc9;
  flex-shrink: 0;
}

.item-quick-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
  flex-shrink: 0;
}

.drawer-item:hover .item-quick-actions,
.drawer-item.active .item-quick-actions {
  opacity: 1;
}

.item-icon-btn {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 7px;
  background: #fff;
  color: #95a0af;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.item-icon-btn:hover {
  background: #edf3ff;
  color: #3572ff;
}

.item-icon-btn.pin.on {
  color: #f0a128;
}

.item-icon-btn.danger:hover {
  background: #fff1f2;
  color: #ef4444;
}

.drawer-foot {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-top: 1px solid #eef1f5;
}

.batch-toggle {
  flex: 1;
  padding: 8px 0;
  border: 1px solid #e4e7ed;
  border-radius: 10px;
  background: #fff;
  color: #8a94a3;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.batch-toggle:hover {
  border-color: #3572ff;
  color: #3572ff;
}

.batch-toggle.active {
  background: #fff7ef;
  border-color: #f28b24;
  color: #f28b24;
}

.refresh-btn {
  width: 34px;
  height: 34px;
  border: none;
  border-radius: 10px;
  background: #f4f6fa;
  color: #556274;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.refresh-btn:hover {
  background: #edf3ff;
  color: #3572ff;
}

/* ===== 移动文件夹对话框 ===== */
.move-folder-content {
  padding: 10px 0;
}

.move-folder-hint {
  margin: 0 0 16px;
  color: #606266;
  font-size: 13px;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.folder-option {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e4e7ed;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.folder-option:hover {
  border-color: #409eff;
  background: #ecf5ff;
  transform: translateX(4px);
}

.folder-option .el-icon {
  color: #e6a23c;
  font-size: 18px;
}

.folder-option span {
  flex: 1;
  color: #606266;
  font-size: 14px;
}

.folder-option-count {
  color: #909399;
  font-size: 12px;
  background: #f4f4f5;
  padding: 4px 8px;
  border-radius: 10px;
  flex-shrink: 0;
}

/* 对话框缩放手柄 */
.dialog-resize-handle {
  position: absolute;
  bottom: 0;
  right: 0;
  width: 24px;
  height: 24px;
  cursor: se-resize;
  z-index: 1000;
}

.dialog-resize-handle::after {
  content: '';
  position: absolute;
  bottom: 4px;
  right: 4px;
  width: 10px;
  height: 10px;
  border-right: 2px solid rgba(0, 201, 167, 0.5);
  border-bottom: 2px solid rgba(0, 201, 167, 0.5);
}

.dialog-resize-handle:hover::after {
  border-right-color: #00c9a7;
  border-bottom-color: #00c9a7;
}

/* ===== 暗色主题 ===== */
:root[data-theme='dark'] .memo-canvas {
  background: linear-gradient(180deg, #161b22 0%, #0f172a 100%);
}

:root[data-theme='dark'] .memo-content-header {
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .memo-title-input :deep(.el-input__inner) {
  color: #e6edf3;
}

:root[data-theme='dark'] .memo-content-meta {
  color: #9ca3af;
}

:root[data-theme='dark'] .memo-scope-badge.global {
  color: #34d399;
  background: rgba(52, 211, 153, 0.15);
}

:root[data-theme='dark'] .memo-scope-badge.book {
  color: #60a5fa;
  background: rgba(96, 165, 250, 0.15);
}

:root[data-theme='dark'] .memo-content-input-large :deep(.el-textarea__inner) {
  color: #c8d6cd;
  background: transparent;
}

:root[data-theme='dark'] .content-footer {
  border-top-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .content-empty {
  color: #4d5768;
}

:root[data-theme='dark'] .icon-btn {
  background: rgba(28, 35, 51, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .icon-btn:hover {
  background: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
}

:root[data-theme='dark'] .icon-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .icon-btn.pinned {
  color: #fbbf24;
}

:root[data-theme='dark'] .folder-select {
  background: rgba(28, 35, 51, 0.6);
  border-color: rgba(42, 50, 66, 0.4);
  color: #c8d6cd;
}

:root[data-theme='dark'] .folder-select:hover {
  border-color: rgba(59, 130, 246, 0.4);
}

:root[data-theme='dark'] .global-memo-container {
  border-color: rgba(42, 50, 66, 0.4);
  background: #161b22;
}

:root[data-theme='dark'] .memo-drawer {
  background: linear-gradient(180deg, #161b22 0%, #0f172a 100%);
  border-left-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .drawer-title {
  color: #e6edf3;
}

:root[data-theme='dark'] .drawer-title .el-icon {
  color: #93c5fd;
}

:root[data-theme='dark'] .d-tab:hover {
  background: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .d-tab.active {
  background: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
}

:root[data-theme='dark'] .folder-chip {
  border-color: rgba(42, 50, 66, 0.4);
  background: rgba(22, 27, 34, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .folder-chip:hover {
  border-color: rgba(59, 130, 246, 0.4);
  color: #93c5fd;
}

:root[data-theme='dark'] .folder-chip.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.4);
  color: #93c5fd;
}

:root[data-theme='dark'] .batch-toolbar {
  background: rgba(242, 139, 36, 0.1);
  border-color: rgba(242, 139, 36, 0.3);
}

:root[data-theme='dark'] .drawer-empty {
  color: #4d5768;
}

:root[data-theme='dark'] .drawer-item:hover {
  background: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .drawer-item.active {
  background: rgba(59, 130, 246, 0.15);
}

:root[data-theme='dark'] .drawer-item.active .item-title {
  color: #93c5fd;
}

:root[data-theme='dark'] .drawer-item.batch-selected {
  background: rgba(242, 139, 36, 0.15);
}

:root[data-theme='dark'] .item-title {
  color: #c8d6cd;
}

:root[data-theme='dark'] .pin-star {
  color: #fbbf24;
}

:root[data-theme='dark'] .item-snippet {
  color: #9ca3af;
}

:root[data-theme='dark'] .item-tag {
  background: rgba(42, 50, 66, 0.4);
  color: #9ca3af;
}

:root[data-theme='dark'] .item-meta-time {
  color: #4d5768;
}

:root[data-theme='dark'] .item-icon-btn {
  background: rgba(28, 35, 51, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .item-icon-btn:hover {
  background: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
}

:root[data-theme='dark'] .item-icon-btn.pin.on {
  color: #fbbf24;
}

:root[data-theme='dark'] .item-icon-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .drawer-foot {
  border-top-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .batch-toggle {
  border-color: rgba(42, 50, 66, 0.4);
  background: rgba(22, 27, 34, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .batch-toggle:hover {
  border-color: rgba(59, 130, 246, 0.4);
  color: #93c5fd;
}

:root[data-theme='dark'] .batch-toggle.active {
  background: rgba(242, 139, 36, 0.1);
  border-color: rgba(242, 139, 36, 0.3);
  color: #fbbf24;
}

:root[data-theme='dark'] .refresh-btn {
  background: rgba(28, 35, 51, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .refresh-btn:hover {
  background: rgba(59, 130, 246, 0.15);
  color: #93c5fd;
}

:root[data-theme='dark'] .folder-option {
  border-color: rgba(42, 50, 66, 0.4);
  background: rgba(22, 27, 34, 0.6);
}

:root[data-theme='dark'] .folder-option:hover {
  border-color: rgba(59, 130, 246, 0.4);
  background: rgba(59, 130, 246, 0.15);
}

:root[data-theme='dark'] .folder-option span {
  color: #c8d6cd;
}

:root[data-theme='dark'] .folder-option-count {
  background: rgba(42, 50, 66, 0.4);
  color: #9ca3af;
}
</style>
