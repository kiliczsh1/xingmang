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
      <div class="global-memo-sidebar" :class="{ collapsed: memoSidebarCollapsed }">
        <div class="memo-workspace-header">
          <div class="memo-workspace-title">
            <el-icon><Notebook /></el-icon>
            <span>备忘录</span>
          </div>
        </div>
        <div v-if="!memoSidebarCollapsed" class="sidebar-toolbar">
          <button
            type="button"
            class="sidebar-action sidebar-action-primary"
            @click="handleCreateGlobalMemoFolder"
          >
            <el-icon><Folder /></el-icon>
            <span>新建文件夹</span>
          </button>
          <button
            type="button"
            class="sidebar-action sidebar-action-accent"
            @click="handleCreateMemoInDialog('global')"
          >
            <el-icon><Plus /></el-icon>
            <span>新建</span>
          </button>
          <button
            type="button"
            class="sidebar-icon-btn"
            :class="{ active: memoBatchMode }"
            @click="toggleBatchMode"
          >
            <span>批量</span>
          </button>
          <button type="button" class="sidebar-icon-btn" @click="$emit('changed')">
            <el-icon><Refresh /></el-icon>
          </button>
          <button type="button" class="sidebar-icon-btn" @click="memoSidebarCollapsed = true">
            <el-icon><ArrowLeft /></el-icon>
          </button>
        </div>
        <div v-else class="sidebar-collapsed-rail">
          <button type="button" class="sidebar-icon-btn" @click="memoSidebarCollapsed = false">
            <el-icon><ArrowRight /></el-icon>
          </button>
        </div>
        <div v-if="!memoSidebarCollapsed && memoBatchMode" class="batch-toolbar">
          <el-checkbox
            :model-value="hasSelectedMemos && filteredMemos.length > 0 && selectedMemoIds.length === filteredMemos.length"
            :indeterminate="hasSelectedMemos && selectedMemoIds.length < filteredMemos.length"
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
        <div v-if="!memoSidebarCollapsed" class="memo-search-box">
          <el-input
            v-model="memoSearchKeyword"
            placeholder="搜索备忘录..."
            clearable
            prefix-icon="Search"
            size="small"
          />
        </div>
        <div v-if="!memoSidebarCollapsed" class="sidebar-list">
          <div v-if="filteredMemos.length === 0" class="sidebar-empty">
            <el-icon><Document /></el-icon>
            <span>暂无备忘录</span>
          </div>
          <template v-else>
            <div class="memo-section-card">
              <div class="memo-tree-root">
                <div
                  v-for="folder in globalMemoFolders"
                  :key="folder"
                  class="memo-tree-folder"
                >
                  <button
                    type="button"
                    class="memo-tree-folder-header"
                    :class="{ active: currentGlobalMemoFolder === folder }"
                    @click="toggleMemoFolder(folder)"
                  >
                    <div class="folder-header-content">
                      <el-icon class="folder-arrow">
                        <ArrowRight v-if="!folderExpandedStates[folder]" />
                        <ArrowDown v-else />
                      </el-icon>
                      <el-icon class="folder-icon"><Folder /></el-icon>
                      <span class="folder-name">{{ folder }}</span>
                      <span class="folder-count">{{ getFolderMemoCount(folder) }}</span>
                      <div class="folder-actions" @click.stop>
                        <button
                          type="button"
                          class="folder-action-btn"
                          title="新建备忘录"
                          @click="createMemoInFolder(folder)"
                        >
                          <el-icon><Plus /></el-icon>
                        </button>
                        <button
                          v-if="folder !== '默认'"
                          type="button"
                          class="folder-action-btn danger"
                          title="删除文件夹"
                          @click="deleteMemoFolder(folder)"
                        >
                          <el-icon><Delete /></el-icon>
                        </button>
                      </div>
                    </div>
                  </button>
                  <div
                    v-show="folderExpandedStates[folder]"
                    class="memo-tree-children"
                  >
                    <div
                      v-for="memo in getMemosByFolder(folder)"
                      :key="memo.id"
                      :class="['memo-tree-item', { active: dialogSelectedMemo?.id === memo.id, 'batch-selected': selectedMemoIds.includes(memo.id) }]"
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
                          <div class="item-title-actions">
                            <button
                              type="button"
                              class="item-title-btn"
                              title="移动到文件夹"
                              @click.stop="openMoveFolderDialog(memo)"
                            >
                              <el-icon><Folder /></el-icon>
                            </button>
                            <button
                              type="button"
                              class="item-title-btn danger"
                              title="删除备忘录"
                              @click.stop="deleteMemoInList(memo)"
                            >
                              <el-icon><Delete /></el-icon>
                            </button>
                          </div>
                        </div>
                        <div class="item-tags-row" v-if="memo.tags">
                          <el-tag
                            v-for="tag in getMemoTags(memo.tags)"
                            :key="tag"
                            size="small"
                            class="item-tag"
                          >
                            {{ tag }}
                          </el-tag>
                        </div>
                        <div class="item-meta-row">
                          <span class="item-meta"></span>
                          <span class="item-meta-time"></span>
                        </div>
                      </div>
                      <div class="item-quick-actions">
                        <button
                          v-if="memo.is_pinned && !memoBatchMode"
                          type="button"
                          class="item-icon-btn pin"
                          @click.stop="toggleMemoPin(memo)"
                        >
                          <el-icon><Star /></el-icon>
                        </button>
                        <button
                          v-if="!memoBatchMode"
                          type="button"
                          class="item-icon-btn danger"
                          @click.stop="deleteMemo(memo)"
                        >
                          <el-icon><Delete /></el-icon>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="memo-section-card">
              <div class="memo-tree-root">
                <div class="memo-tree-folder">
                  <button
                    type="button"
                    class="memo-tree-folder-header"
                    :class="{ active: bookMemoExpanded }"
                    @click="bookMemoExpanded = !bookMemoExpanded"
                  >
                    <div class="folder-header-content">
                      <el-icon class="folder-arrow">
                        <ArrowRight v-if="!bookMemoExpanded" />
                        <ArrowDown v-else />
                      </el-icon>
                      <el-icon class="folder-icon"><Folder /></el-icon>
                      <span class="folder-name">本书备忘录</span>
                      <span class="folder-count">{{ bookMemoList.length }}</span>
                    </div>
                  </button>
                  <div
                    v-show="bookMemoExpanded"
                    class="memo-tree-children"
                  >
                    <div
                      v-for="memo in bookMemoList"
                      :key="memo.id"
                      :class="['memo-tree-item', { active: dialogSelectedMemo?.id === memo.id, 'batch-selected': selectedMemoIds.includes(memo.id) }]"
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
                          <button
                            type="button"
                            class="item-move-btn"
                            title="移动到文件夹"
                            @click.stop="openMoveFolderDialog(memo)"
                          >
                            <el-icon><Folder /></el-icon>
                          </button>
                        </div>
                        <div class="item-tags-row" v-if="memo.tags">
                          <el-tag
                            v-for="tag in getMemoTags(memo.tags)"
                            :key="tag"
                            size="small"
                            class="item-tag"
                          >
                            {{ tag }}
                          </el-tag>
                        </div>
                        <div class="item-meta-row">
                          <span class="item-meta">{{ getContentLength(memo.content) }} 字</span>
                          <span class="item-meta-time">{{ formatTime(memo.updated_at) }}</span>
                        </div>
                      </div>
                      <div class="item-quick-actions">
                        <button
                          v-if="memo.is_pinned && !memoBatchMode"
                          type="button"
                          class="item-icon-btn pin"
                          @click.stop="toggleMemoPin(memo)"
                        >
                          <el-icon><Star /></el-icon>
                        </button>
                        <button
                          v-if="!memoBatchMode"
                          type="button"
                          class="item-icon-btn danger"
                          @click.stop="deleteMemo(memo)"
                        >
                          <el-icon><Delete /></el-icon>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
      <div class="global-memo-content">
        <div v-if="dialogSelectedMemo" class="content-area">
          <div class="memo-content-header">
            <div class="memo-content-header-main">
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
            </div>
            <div class="memo-content-header-side">
              <span
                class="memo-scope-badge"
                :class="isBookMemo(dialogSelectedMemo) ? 'book' : 'global'"
              >
                {{ isBookMemo(dialogSelectedMemo) ? '本书备忘录' : '全局备忘录' }}
              </span>
            </div>
          </div>
          <el-input
            v-model="dialogSelectedMemo.content"
            type="textarea"
            placeholder="在这里输入内容..."
            class="memo-content-input memo-content-input-large"
          />
          <div class="content-footer minimalist">
            <div class="footer-actions">
              <el-button @click="$emit('jumpToMemo', dialogSelectedMemo)">跳转编辑</el-button>
              <el-button type="primary" @click="saveDialogMemo">保存</el-button>
            </div>
          </div>
        </div>
        <div v-else class="content-empty">
          <el-icon><EditPen /></el-icon>
          <span>请选择或新建备忘录</span>
        </div>
      </div>
      <div
        class="dialog-resize-handle"
        @mousedown="startDialogResize"
      ></div>
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
  Notebook, Plus, Delete, Folder,
  ArrowRight, ArrowLeft, ArrowDown, EditPen,
  Star, Refresh, Document
} from '@element-plus/icons-vue'
import { memoAPI } from '@/api'
import type { Memo } from '@/types'

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
const memoSidebarCollapsed = ref(false)
const bookMemoExpanded = ref(true)
const memoCreateScope = ref<'global' | 'book'>('book')
const currentGlobalMemoFolder = ref('默认')
const globalMemoCustomFolders = ref<string[]>([])
const folderExpandedStates = ref<Record<string, boolean>>({
  '默认': true
})
const moveFolderDialogVisible = ref(false)
const memoToMove = ref<Memo | null>(null)

// ===== 工具函数 =====
const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`

  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatExactTime = (timeStr: string) => {
  if (!timeStr) return '--'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getContentLength = (content: string) => {
  if (!content) return 0
  const plainText = content
    .replace(/<[^>]*>/g, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`[^`]*`/g, '')
    .replace(/\s+/g, '')
  return [...plainText].length
}

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
const toggleMemoFolder = (folder: string) => {
  if (currentGlobalMemoFolder.value !== folder) {
    currentGlobalMemoFolder.value = folder
    folderExpandedStates.value[folder] = true
  } else {
    folderExpandedStates.value[folder] = !folderExpandedStates.value[folder]
  }
}

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
    currentGlobalMemoFolder.value = folderName
    folderExpandedStates.value[folderName] = true
    memoCreateScope.value = 'global'
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

    if (currentGlobalMemoFolder.value === folder) {
      currentGlobalMemoFolder.value = '默认'
    }
    delete folderExpandedStates.value[folder]

    ElMessage.success('文件夹已删除')
    emit('changed')
  } catch (error) {
    // 用户取消
  }
}

// ===== 备忘录 CRUD =====
const handleCreateMemoInDialog = async (scope?: 'global' | 'book') => {
  const targetScope = scope || memoCreateScope.value
  memoCreateScope.value = targetScope
  const res = await memoAPI.create({
    title: '新备忘录',
    content: '',
    category: targetScope === 'book' ? '本书' : currentGlobalMemoFolder.value,
    order_num: props.memos.length
  })
  if (res.success && res.data) {
    selectDialogMemo(res.data)
    emit('changed')
  }
}

const createMemoInFolder = async (folder: string) => {
  const res = await memoAPI.create({
    title: '新备忘录',
    content: '',
    category: folder === '默认' ? '全局' : folder,
    order_num: props.memos.length
  })
  if (res.success && res.data) {
    currentGlobalMemoFolder.value = folder
    folderExpandedStates.value[folder] = true
    ElMessage.success('备忘录创建成功')
    emit('changed')
  }
}

const selectDialogMemo = (memo: Memo) => {
  memoCreateScope.value = isBookMemo(memo) ? 'book' : 'global'
  if (!isBookMemo(memo)) {
    currentGlobalMemoFolder.value = getGlobalMemoFolder(memo)
  }
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
    selectedMemoIds.value = filteredMemos.value.map(m => m.id)
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

.global-memo-sidebar {
  width: 280px;
  border-right: 1px solid rgba(228, 232, 240, 0.95);
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #fffdfa 0%, #ffffff 100%);
  transition: width 0.25s ease;
}

.global-memo-sidebar.collapsed {
  width: 56px;
}

.memo-workspace-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px 8px;
}

.memo-workspace-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #243042;
  font-size: 13px;
  font-weight: 700;
}

.memo-workspace-title .el-icon {
  color: #4d8dff;
  font-size: 15px;
}

.sidebar-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px 10px;
}

.sidebar-action,
.sidebar-icon-btn {
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar-action {
  height: 40px;
  padding: 0 12px;
  border-radius: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
}

.sidebar-action-primary {
  background: #fff4e6;
  color: #f28b24;
}

.sidebar-action-accent {
  background: #edf3ff;
  color: #3572ff;
}

.sidebar-icon-btn {
  width: 32px;
  height: 32px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f4f6fa;
  color: #556274;
  font-size: 12px;
  font-weight: 600;
}

.sidebar-icon-btn.active,
.sidebar-action:hover,
.sidebar-icon-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(31, 45, 61, 0.08);
}

.sidebar-collapsed-rail {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 12px 0;
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

.memo-search-box {
  padding: 0 14px 10px;
}

.memo-section-card {
  margin-bottom: 0;
}

.memo-tree-root {
  padding: 4px;
}

.memo-tree-folder {
  margin-bottom: 2px;
}

.memo-tree-folder-header {
  width: 100%;
  padding: 8px 12px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  border-radius: 6px;
  transition: all 0.2s ease;
  text-align: left;
}

.memo-tree-folder-header:hover {
  background: #f5f7fa;
}

.memo-tree-folder-header.active {
  background: #ecf5ff;
}

.memo-tree-folder-header.active .folder-name {
  color: #409eff;
  font-weight: 600;
}

.folder-header-content {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}

.folder-arrow {
  color: #909399;
  font-size: 12px;
  transition: transform 0.2s ease;
  flex-shrink: 0;
}

.folder-icon {
  color: #e6a23c;
  font-size: 16px;
  flex-shrink: 0;
}

.folder-name {
  flex: 1;
  color: #606266;
  font-size: 13px;
  text-align: left;
}

.folder-count {
  color: #909399;
  font-size: 11px;
  background: #f4f4f5;
  padding: 2px 6px;
  border-radius: 8px;
  flex-shrink: 0;
}

.folder-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
  flex-shrink: 0;
}

.memo-tree-folder-header:hover .folder-actions {
  opacity: 1;
}

.folder-action-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  background: transparent;
  color: #909399;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.folder-action-btn:hover {
  background: #f5f7fa;
  color: #409eff;
}

.folder-action-btn.danger:hover {
  background: #fef0f0;
  color: #f56c6c;
}

.item-title-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.memo-tree-item:hover .item-title-actions {
  opacity: 1;
}

.item-title-btn {
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  background: transparent;
  color: #909399;
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.item-title-btn:hover {
  background: #f5f7fa;
  color: #409eff;
}

.item-title-btn.danger:hover {
  background: #fef0f0;
  color: #f56c6c;
}

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

.folder-option.active {
  border-color: #409eff;
  background: #ecf5ff;
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

.memo-tree-children {
  padding-left: 20px;
}

.memo-tree-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  margin: 2px 0;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.memo-tree-item:hover {
  background: #f5f7fa;
}

.memo-tree-item.active {
  background: #ecf5ff;
}

.memo-tree-item.active .item-title {
  color: #409eff;
  font-weight: 600;
}

.memo-tree-item.batch-selected {
  background: #fff7ef;
}

.batch-checkbox {
  margin-right: 12px;
}

.sidebar-item.batch-selected {
  background: #fff3e7;
}

.sidebar-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 14px 14px;
}

.sidebar-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 8px;
  color: #9bb8b5;
}

.sidebar-empty .el-icon {
  font-size: 32px;
}

.sidebar-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 2px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar-item:hover {
  background: #f5f7fa;
}

.sidebar-item.active {
  background: #ecf5ff;
  color: #409eff;
}

.sidebar-item.active .item-title {
  color: #409eff;
  font-weight: 600;
}

.item-content {
  min-width: 0;
  flex: 1;
}

.item-title-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.item-title {
  display: block;
  font-size: 13px;
  font-weight: 400;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.memo-scope-tag {
  flex-shrink: 0;
  padding: 0 6px;
  border-radius: 4px;
  font-size: 11px;
  line-height: 18px;
  background: transparent;
  border: 1px solid;
}

.memo-scope-tag.global {
  color: #67c23a;
  border-color: #67c23a;
}

.memo-scope-tag.book {
  color: #409eff;
  border-color: #409eff;
}

.item-meta {
  display: block;
  font-size: 12px;
  color: #909399;
}

.item-meta-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 4px;
}

.item-tags-row {
  display: flex;
  gap: 4px;
  margin-top: 4px;
  flex-wrap: wrap;
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
}

.item-quick-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.sidebar-item:hover .item-quick-actions,
.sidebar-item.active .item-quick-actions {
  opacity: 1;
}

.item-icon-btn {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 999px;
  background: #fff;
  color: #95a0af;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.item-icon-btn.pin {
  color: #f0a128;
}

.item-icon-btn.danger:hover {
  background: #fff1f2;
  color: #ef4444;
}

.global-memo-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: linear-gradient(180deg, #ffffff 0%, #fcfdff 100%);
}

.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 18px 22px;
}

.memo-content-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eef1f5;
}

.memo-content-header-main {
  flex: 1;
  min-width: 0;
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

.memo-content-header-side {
  flex-shrink: 0;
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
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #eef1f5;
}

.content-footer.minimalist {
  justify-content: flex-end;
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

/* memo 对话框暗色主题 */
:root[data-theme='dark'] .memo-content-input-large :deep(.el-textarea__inner) {
  color: #d1d5db;
  background: transparent;
}

:root[data-theme='dark'] .content-footer {
  border-top-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .content-empty {
  color: #6b7280;
}

:root[data-theme='dark'] .global-memo-container {
  border-color: rgba(71, 85, 105, 0.4);
  background: #1e293b;
}

:root[data-theme='dark'] .global-memo-sidebar {
  border-right-color: rgba(71, 85, 105, 0.4);
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
}

:root[data-theme='dark'] .memo-workspace-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .memo-workspace-title .el-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-action-primary {
  background: rgba(242, 139, 36, 0.15);
  color: #fbbf24;
}

:root[data-theme='dark'] .sidebar-action-accent {
  background: rgba(94, 234, 212, 0.15);
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-icon-btn {
  background: rgba(51, 65, 85, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .sidebar-icon-btn:hover {
  background: rgba(71, 85, 105, 0.6);
}

:root[data-theme='dark'] .batch-toolbar {
  background: rgba(242, 139, 36, 0.1);
  border-color: rgba(242, 139, 36, 0.3);
}

:root[data-theme='dark'] .memo-tree-folder-header:hover {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .memo-tree-folder-header.active {
  background: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .memo-tree-folder-header.active .folder-name {
  color: #5eead4;
}

:root[data-theme='dark'] .folder-arrow,
:root[data-theme='dark'] .folder-count {
  color: #9ca3af;
}

:root[data-theme='dark'] .folder-name {
  color: #d1d5db;
}

:root[data-theme='dark'] .folder-count {
  background: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .folder-action-btn:hover {
  background: rgba(71, 85, 105, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .folder-action-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .item-title-btn:hover {
  background: rgba(71, 85, 105, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .item-title-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .folder-option {
  border-color: rgba(71, 85, 105, 0.4);
  background: rgba(30, 41, 59, 0.6);
}

:root[data-theme='dark'] .folder-option:hover,
:root[data-theme='dark'] .folder-option.active {
  border-color: rgba(94, 234, 212, 0.4);
  background: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .folder-option span {
  color: #d1d5db;
}

:root[data-theme='dark'] .folder-option-count {
  background: rgba(71, 85, 105, 0.4);
  color: #9ca3af;
}

:root[data-theme='dark'] .memo-tree-item:hover {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .memo-tree-item.active {
  background: rgba(94, 234, 212, 0.15);
}

:root[data-theme='dark'] .memo-tree-item.active .item-title {
  color: #5eead4;
}

:root[data-theme='dark'] .memo-tree-item.batch-selected {
  background: rgba(242, 139, 36, 0.15);
}

:root[data-theme='dark'] .sidebar-empty {
  color: #6b7280;
}

:root[data-theme='dark'] .sidebar-item:hover {
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .sidebar-item.active {
  background: rgba(94, 234, 212, 0.15);
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-item.active .item-title {
  color: #5eead4;
}

:root[data-theme='dark'] .sidebar-item.batch-selected {
  background: rgba(242, 139, 36, 0.15);
}

:root[data-theme='dark'] .item-title {
  color: #d1d5db;
}

:root[data-theme='dark'] .item-meta {
  color: #9ca3af;
}

:root[data-theme='dark'] .item-tag {
  background: rgba(71, 85, 105, 0.4);
  color: #9ca3af;
}

:root[data-theme='dark'] .item-meta-time {
  color: #6b7280;
}

:root[data-theme='dark'] .item-icon-btn {
  background: rgba(51, 65, 85, 0.6);
  color: #9ca3af;
}

:root[data-theme='dark'] .item-icon-btn.pin {
  color: #fbbf24;
}

:root[data-theme='dark'] .item-icon-btn.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #f87171;
}

:root[data-theme='dark'] .global-memo-content {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
}

:root[data-theme='dark'] .memo-content-header {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .memo-title-input :deep(.el-input__inner) {
  color: #f3f4f6;
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
</style>
