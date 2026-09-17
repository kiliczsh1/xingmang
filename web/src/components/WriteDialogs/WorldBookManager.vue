<template>
  <!-- 世界书管理对话框 -->
  <el-dialog
    v-model="visible"
    :title="`世界书管理 - ${bookTitle || '当前书籍'}`"
    width="900px"
    top="3vh"
    align-center
    append-to-body
    class="worldbook-dialog xm-dialog"
  >
    <div class="worldbook-manager">
      <div class="worldbook-toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="worldbookSearchKeyword"
            placeholder="搜索条目..."
            clearable
            :prefix-icon="Search"
            style="width: 300px"
          />
          <div class="worldbook-stats-inline">
            <span>共 {{ worldBookEntryCount }} 条</span>
            <span>已启用 {{ enabledWorldBookCount }} 条</span>
          </div>
        </div>
        <div class="toolbar-right">
          <el-button size="small" @click="openCreateGroupDialog">
            <el-icon><Folder /></el-icon>
            新建分组
          </el-button>
          <el-dropdown @command="handleImportCommand" trigger="click">
            <el-button size="small">
              <el-icon><Upload /></el-icon>
              导入
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="current">
                  <el-icon><Plus /></el-icon>
                  导入到当前世界书
                </el-dropdown-item>
                <el-dropdown-item command="new">
                  <el-icon><FolderAdd /></el-icon>
                  导入为新世界书
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-dropdown @command="handleExportFormat" trigger="click">
            <el-button size="small">
              <el-icon><Download /></el-icon>
              导出
              <el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="default">
                  <el-icon><Document /></el-icon>
                  导出为本项目格式
                </el-dropdown-item>
                <el-dropdown-item command="sillytavern">
                  <el-icon><Connection /></el-icon>
                  导出为酒馆格式
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button type="primary" size="small" @click="createWorldBookEntry">
            <el-icon><Plus /></el-icon>
            新建条目
          </el-button>
        </div>
      </div>

      <!-- 世界书选择器 -->
      <div class="worldbook-selector">
        <el-select
          v-model="currentWorldBookName"
          placeholder="选择世界书"
          size="small"
          style="width: 220px"
          @change="collapsedGroups.clear()"
        >
          <el-option
            v-for="name in worldBookList"
            :key="name"
            :label="name"
            :value="name"
          />
        </el-select>
        <el-button size="small" @click="openCreateWorldBookDialog">
          <el-icon><Plus /></el-icon>
          新建
        </el-button>
        <el-button size="small" @click="openRenameWorldBookDialog" :disabled="currentWorldBookName === '默认世界书'">
          <el-icon><Edit /></el-icon>
          重命名
        </el-button>
        <el-button size="small" type="danger" @click="handleDeleteWorldBook">
          <el-icon><Delete /></el-icon>
          删除
        </el-button>
      </div>

      <div class="worldbook-list-container">
        <div class="worldbook-grouped-list">
          <div v-if="worldBookGroups.length === 0" class="worldbook-empty-state">
            <el-empty description="暂无世界书条目">
              <el-button type="primary" @click="createWorldBookEntry">创建第一个条目</el-button>
            </el-empty>
          </div>
          <template v-else>
            <div
              v-for="[groupName, entries] in worldBookGroups"
              :key="groupName"
              class="wb-group-section"
            >
              <div class="wb-group-header" @click="toggleGroupCollapse(groupName)">
                <el-icon class="group-collapse-icon">
                  <component :is="collapsedGroups.has(groupName) ? 'ArrowRight' : 'ArrowDown'" />
                </el-icon>
                <span class="wb-group-name">{{ groupName || '默认' }}</span>
                <span class="wb-group-count">{{ entries.length }}条</span>
                <div class="wb-group-actions" @click.stop>
                  <el-button
                    v-if="groupName"
                    size="small"
                    text
                    @click="openRenameGroupDialog(groupName)"
                  >
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button
                    v-if="groupName"
                    size="small"
                    text
                    type="danger"
                    @click="deleteGroup(groupName)"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
              <div v-show="!collapsedGroups.has(groupName)" class="wb-group-body">
                <div class="wb-entry-grid">
                  <div
                    v-for="entry in entries"
                    :key="entry.uid"
                    :class="['wb-entry-card', { disabled: entry.disable }]"
                  >
                    <div class="wb-entry-top">
                      <el-switch
                        :model-value="!entry.disable"
                        size="small"
                        @click.stop
                        @change="(val: boolean) => toggleWorldBookEntry(entry, val)"
                      />
                      <div class="wb-entry-tags">
                        <el-tag v-if="entry.constant" size="small" type="success">常驻</el-tag>
                        <el-tag v-if="entry.disable" size="small" type="danger">禁用</el-tag>
                      </div>
                    </div>
                    <div class="wb-entry-title">{{ entry.comment || '未命名条目' }}</div>
                    <div class="wb-entry-actions">
                      <el-button size="small" @click="editWorldBookEntry(entry)" circle>
                        <el-icon><Edit /></el-icon>
                      </el-button>
                      <el-button size="small" type="danger" @click="deleteWorldBookEntry(entry)" circle>
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </div>
      </div>
    </div>
  </el-dialog>

  <!-- 分组管理弹窗 -->
  <el-dialog
    v-model="showGroupDialog"
    :title="groupDialogMode === 'create' ? '新建分组' : '重命名分组'"
    width="400px"
    align-center
    append-to-body
    destroy-on-close
    class="xm-dialog"
  >
    <el-input
      v-model="groupDialogName"
      placeholder="输入分组名称"
      @keydown.enter="confirmGroupDialog"
    />
    <template #footer>
      <el-button @click="showGroupDialog = false">取消</el-button>
      <el-button type="primary" @click="confirmGroupDialog">确定</el-button>
    </template>
  </el-dialog>

  <!-- 世界书创建/重命名弹窗 -->
  <el-dialog
    v-model="showWorldBookDialog"
    :title="worldBookDialogMode === 'create' ? '新建世界书' : '重命名世界书'"
    width="400px"
    align-center
    append-to-body
    destroy-on-close
    class="xm-dialog"
  >
    <el-input
      v-model="worldBookDialogName"
      placeholder="输入世界书名称"
      @keydown.enter="confirmWorldBookDialog"
    />
    <template #footer>
      <el-button @click="showWorldBookDialog = false">取消</el-button>
      <el-button type="primary" @click="confirmWorldBookDialog">确定</el-button>
    </template>
  </el-dialog>

  <!-- 世界书条目编辑对话框 -->
  <el-dialog
    v-model="worldBookEditorVisible"
    :title="editingWorldBookEntry ? '编辑条目' : '新建条目'"
    width="800px"
    top="3vh"
    append-to-body
    destroy-on-close
    class="xm-dialog"
    :close-on-click-modal="false"
  >
    <div class="worldbook-editor-form">
      <div class="editor-field">
        <label class="field-label">标题/备注</label>
        <el-input v-model="worldBookForm.comment" placeholder="输入条目标题或备注" />
      </div>

      <div class="editor-field">
        <label class="field-label">内容</label>
        <el-input
          v-model="worldBookForm.content"
          type="textarea"
          :rows="8"
          placeholder="触发后注入的内容，支持 Markdown 格式"
        />
      </div>

      <div class="editor-row">
        <div class="editor-field half">
          <label class="field-label">插入位置</label>
          <el-select v-model="worldBookForm.position" placeholder="选择插入位置">
            <el-option :value="0" label="在提示词前" />
            <el-option :value="1" label="在提示词后" />
            <el-option :value="2" label="在作者注释后" />
            <el-option :value="3" label="在角色描述后" />
            <el-option :value="7" label="世界书顶部" />
            <el-option :value="8" label="世界书底部" />
          </el-select>
        </div>

        <div class="editor-field half">
          <label class="field-label">优先级</label>
          <el-input-number v-model="worldBookForm.order" :min="0" :max="1000" />
        </div>
      </div>

      <div class="editor-row">
        <div class="editor-field half">
          <label class="field-label">扫描深度</label>
          <el-input-number v-model="worldBookForm.depth" :min="1" :max="999" />
        </div>
      </div>

      <div class="editor-field">
        <label class="field-label">选项</label>
        <div class="checkbox-group">
          <label class="field-label">状态</label>
          <el-radio-group v-model="worldBookEntryStatus" size="small">
            <el-radio value="normal">启用</el-radio>
            <el-radio value="constant">常驻激活</el-radio>
            <el-radio value="disable">禁用</el-radio>
          </el-radio-group>
        </div>
      </div>

      <div class="editor-field">
        <label class="field-label">分组</label>
        <el-select
          v-model="worldBookForm.group"
          filterable
          allow-create
          clearable
          placeholder="选择或输入分组名称"
          style="width: 100%;"
        >
          <el-option
            v-for="g in availableGroups"
            :key="g"
            :label="g || '默认'"
            :value="g"
          />
        </el-select>
      </div>

    </div>

    <template #footer>
      <el-button @click="worldBookEditorVisible = false">取消</el-button>
      <el-button type="primary" @click="saveWorldBookEntry">
        {{ editingWorldBookEntry ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>

  <input
    ref="worldBookFileInput"
    type="file"
    accept=".json"
    style="display: none"
    @change="handleWorldBookFileSelect"
  />
  <input
    ref="importAsNewFileInput"
    type="file"
    accept=".json"
    style="display: none"
    @change="handleWorldBookImportAsNew"
  />
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { useWorldBookStore } from '@/stores/worldbook'
import type { WorldBookEntry } from '@/types/worldbook'
import { createDefaultWorldBookEntry } from '@/types/worldbook'

const props = defineProps<{
  bookId: number
  bookTitle?: string
}>()

const visible = defineModel<boolean>('visible')

const worldBookStore = useWorldBookStore()
const worldbookSearchKeyword = ref('')
const worldBookEditorVisible = ref(false)
const editingWorldBookEntry = ref<WorldBookEntry | null>(null)
const worldBookFileInput = ref<HTMLInputElement>()
const importAsNewFileInput = ref<HTMLInputElement>()

// 世界书选择器
const currentWorldBookName = computed({
  get: () => worldBookStore.getCurrentWorldBookName(String(props.bookId)),
  set: (name: string) => {
    worldBookStore.setCurrentWorldBookName(String(props.bookId), name)
  }
})
const worldBookList = computed(() => worldBookStore.getBookWorldBookList(String(props.bookId)))
const showWorldBookDialog = ref(false)
const worldBookDialogMode = ref<'create' | 'rename'>('create')
const worldBookDialogName = ref('')

const openCreateWorldBookDialog = () => {
  worldBookDialogMode.value = 'create'
  worldBookDialogName.value = ''
  showWorldBookDialog.value = true
}

const openRenameWorldBookDialog = () => {
  worldBookDialogMode.value = 'rename'
  worldBookDialogName.value = currentWorldBookName.value
  showWorldBookDialog.value = true
}

const confirmWorldBookDialog = () => {
  const name = worldBookDialogName.value.trim()
  if (!name) return ElMessage.warning('名称不能为空')
  if (worldBookDialogMode.value === 'create') {
    worldBookStore.createWorldBook(String(props.bookId), name)
    ElMessage.success('世界书已创建')
  } else {
    if (name === currentWorldBookName.value) {
      showWorldBookDialog.value = false
      return
    }
    if (!worldBookStore.renameWorldBook(String(props.bookId), currentWorldBookName.value, name)) {
      return ElMessage.warning('名称已存在或操作失败')
    }
    ElMessage.success('已重命名')
  }
  showWorldBookDialog.value = false
}

const handleDeleteWorldBook = async () => {
  const name = currentWorldBookName.value
  try {
    await ElMessageBox.confirm(`确定删除世界书「${name}」及其中所有条目吗？`, '删除确认', { type: 'warning' })
    worldBookStore.deleteWorldBook(String(props.bookId), name)
    ElMessage.success('已删除')
  } catch {}
}

const worldBookForm = ref(createDefaultWorldBookEntry())

// 词条状态：normal/constant/disable 三选一
const worldBookEntryStatus = computed({
  get: () => worldBookForm.value.disable ? 'disable' : worldBookForm.value.constant ? 'constant' : 'normal',
  set: (val: string) => {
    worldBookForm.value.constant = val === 'constant'
    worldBookForm.value.disable = val === 'disable'
  }
})

const currentBookWorldBook = computed(() => {
  return worldBookStore.getBookWorldBook(String(props.bookId), currentWorldBookName.value)
})

const filteredWorldBookEntries = computed(() => {
  const entries = currentBookWorldBook.value?.entries || []

  if (!worldbookSearchKeyword.value.trim()) {
    return entries
  }

  const keyword = worldbookSearchKeyword.value.toLowerCase()
  return entries.filter(entry => {
    return (
      entry.comment.toLowerCase().includes(keyword) ||
      entry.key.some(k => k.toLowerCase().includes(keyword)) ||
      entry.content.toLowerCase().includes(keyword)
    )
  })
})

// 分组
const collapsedGroups = ref<Set<string>>(new Set())

const worldBookGroups = computed(() => {
  const entries = filteredWorldBookEntries.value
  const groupMap = new Map<string, WorldBookEntry[]>()

  entries.forEach(entry => {
    const g = entry.group?.trim() || ''
    if (!groupMap.has(g)) groupMap.set(g, [])
    groupMap.get(g)!.push(entry)
  })

  // 有名称的分组在前 → 默认（空分组名）在最后
  return Array.from(groupMap.entries()).sort((a, b) => {
    if (!a[0] && !b[0]) return 0
    if (!a[0]) return 1
    if (!b[0]) return -1
    return a[0].localeCompare(b[0])
  })
})

// 编辑器分组下拉可选值
const availableGroups = computed(() => {
  const names = new Set<string>()
  currentBookWorldBook.value?.entries.forEach(e => {
    const g = e.group?.trim()
    if (g) names.add(g)
  })
  return Array.from(names).sort((a, b) => a.localeCompare(b))
})

const toggleGroupCollapse = (group: string) => {
  if (collapsedGroups.value.has(group)) {
    collapsedGroups.value.delete(group)
  } else {
    collapsedGroups.value.add(group)
  }
}

// 分组管理
const showGroupDialog = ref(false)
const groupDialogMode = ref<'create' | 'rename'>('create')
const groupDialogName = ref('')
const groupDialogOldName = ref('')

const openCreateGroupDialog = () => {
  groupDialogMode.value = 'create'
  groupDialogName.value = ''
  groupDialogOldName.value = ''
  showGroupDialog.value = true
}

const openRenameGroupDialog = (oldName: string) => {
  groupDialogMode.value = 'rename'
  groupDialogName.value = oldName
  groupDialogOldName.value = oldName
  showGroupDialog.value = true
}

const confirmGroupDialog = async () => {
  const name = groupDialogName.value.trim()
  if (!name) {
    ElMessage.warning('请输入分组名称')
    return
  }
  if (groupDialogMode.value === 'rename' && name === groupDialogOldName.value) {
    showGroupDialog.value = false
    return
  }

  if (groupDialogMode.value === 'create') {
    showGroupDialog.value = false
    ElMessage.success(`分组"${name}"已创建，新建条目时可选择该分组`)
  } else {
    const oldName = groupDialogOldName.value
    const entries = currentBookWorldBook.value?.entries.filter(e => (e.group?.trim() || '') === oldName) || []
    for (const entry of entries) {
      await worldBookStore.updateEntry(entry.uid, { group: name }, String(props.bookId), currentWorldBookName.value)
    }
    showGroupDialog.value = false
    ElMessage.success(`分组已重命名为"${name}"`)
  }
}

const deleteGroup = async (groupName: string) => {
  try {
    const entries = currentBookWorldBook.value?.entries.filter(e => (e.group?.trim() || '') === groupName) || []
    await ElMessageBox.confirm(
      `确定删除分组"${groupName}"吗？该分组下的 ${entries.length} 个条目将移至默认分组。`,
      '删除分组',
      { type: 'warning' }
    )
    for (const entry of entries) {
      await worldBookStore.updateEntry(entry.uid, { group: '' }, String(props.bookId))
    }
    ElMessage.success('分组已删除')
  } catch {
    // 取消
  }
}

const worldBookEntryCount = computed(() => {
  return currentBookWorldBook.value?.entries.length || 0
})

const enabledWorldBookCount = computed(() => {
  const entries = currentBookWorldBook.value?.entries || []
  return entries.filter(e => !e.disable).length
})

const toggleWorldBookEntry = async (entry: WorldBookEntry, enabled: boolean) => {
  entry.disable = !enabled
  await worldBookStore.updateEntry(entry.uid, { disable: !enabled }, String(props.bookId))
}

const createWorldBookEntry = () => {
  editingWorldBookEntry.value = null
  worldBookForm.value = createDefaultWorldBookEntry()
  worldBookEditorVisible.value = true
}

const editWorldBookEntry = (entry: WorldBookEntry) => {
  editingWorldBookEntry.value = entry
  worldBookForm.value = { ...entry }
  worldBookEditorVisible.value = true
}

const saveWorldBookEntry = async () => {
  if (!worldBookForm.value.comment.trim()) {
    ElMessage.warning('请输入条目标题')
    return
  }

  if (editingWorldBookEntry.value) {
    await worldBookStore.updateEntry(editingWorldBookEntry.value.uid, worldBookForm.value, String(props.bookId))
    ElMessage.success('条目已更新')
  } else {
    await worldBookStore.addEntry(worldBookForm.value, String(props.bookId))
    ElMessage.success('条目已创建')
  }

  worldBookEditorVisible.value = false
}

const deleteWorldBookEntry = async (entry: WorldBookEntry) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除条目"${entry.comment || '未命名条目'}"吗？`,
      '删除确认',
      { type: 'warning' }
    )
    await worldBookStore.deleteEntry(entry.uid, String(props.bookId))
    ElMessage.success('条目已删除')
  } catch {
    // 取消
  }
}

const handleImportCommand = (command: string) => {
  if (command === 'current') {
    worldBookFileInput.value?.click()
  } else if (command === 'new') {
    importAsNewFileInput.value?.click()
  }
}

const handleWorldBookFileSelect = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)

    const success = worldBookStore.importWorldBookData(data, String(props.bookId))
    if (success) {
      ElMessage.success('世界书导入成功')
    } else {
      ElMessage.error('世界书格式不正确')
    }
  } catch (error) {
    ElMessage.error('导入失败：文件格式错误')
  }

  if (worldBookFileInput.value) {
    worldBookFileInput.value.value = ''
  }
}

const handleWorldBookImportAsNew = async (event: Event) => {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)

    const fileName = file.name.replace(/\.(json|wb)$/i, '')

    const existingList = worldBookStore.getBookWorldBookList(String(props.bookId))
    if (existingList.includes(fileName)) {
      try {
        await ElMessageBox.confirm(
          `世界书「${fileName}」已存在，是否合并导入？`,
          '名称冲突',
          { type: 'warning' }
        )
      } catch {
        if (importAsNewFileInput.value) {
          importAsNewFileInput.value.value = ''
        }
        return
      }
    }

    data.name = fileName

    const success = worldBookStore.importWorldBookData(data, String(props.bookId))
    if (success) {
      ElMessage.success(`已导入为新世界书「${fileName}」`)
    } else {
      ElMessage.error('世界书格式不正确')
    }
  } catch (error) {
    ElMessage.error('导入失败：文件格式错误')
  }

  if (importAsNewFileInput.value) {
    importAsNewFileInput.value.value = ''
  }
}

const exportWorldBook = (format: 'default' | 'sillytavern' = 'default') => {
  const data = worldBookStore.exportWorldBookData(String(props.bookId), format)
  if (!data || !data.entries || (Array.isArray(data.entries) ? data.entries.length === 0 : Object.keys(data.entries).length === 0)) {
    ElMessage.warning('没有可导出的条目')
    return
  }

  const formatLabel = format === 'sillytavern' ? '酒馆' : '默认'
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `worldbook-${props.bookTitle || 'book'}-${formatLabel}-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)

  ElMessage.success(`世界书已导出（${formatLabel}格式）`)
}

const handleExportFormat = (format: string) => {
  exportWorldBook(format as 'default' | 'sillytavern')
}
</script>


<style scoped>


.field-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #164e49;
  font-size: 14px;
}


.worldbook-dialog :deep(.el-dialog__body) {
  padding: 0;
}


.worldbook-manager {
  display: flex;
  flex-direction: column;
  height: 520px;
  max-height: 70vh;
}


.worldbook-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}


.worldbook-list-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}


.worldbook-grouped-list {
  flex: 1;
  overflow-y: auto;
  padding: 1px 8px 0;
  min-height: 0;
}


.worldbook-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}


.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
  align-items: center;
}


.worldbook-selector {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}


.worldbook-empty-state {
  padding: 60px 20px;
  text-align: center;
}


/* 分组网格布局 */
.worldbook-list-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}


.worldbook-grouped-list {
  flex: 1;
  overflow-y: auto;
  padding: 1px 8px 0;
  min-height: 0;
}


.wb-group-section {
  margin-bottom: 1px;
}


.wb-group-header {
  display: flex;
  align-items: center;
  gap: 3px;
  padding: 2px 6px;
  background: var(--el-fill-color-light);
  border-radius: 3px;
  cursor: pointer;
  user-select: none;
  margin-bottom: 1px;
}


.wb-group-header:hover {
  background: var(--el-fill-color);
}


.group-collapse-icon {
  flex-shrink: 0;
  font-size: 11px;
}


.wb-group-name {
  font-weight: 600;
  font-size: 11px;
  color: var(--el-text-color-primary);
}


.wb-group-count {
  font-size: 10px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  margin-left: auto;
}


.wb-group-actions {
  display: flex;
  gap: 1px;
  flex-shrink: 0;
}


.wb-entry-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2px;
  min-width: 0;
}


.wb-entry-card {
  padding: 3px 5px;
  background: var(--el-fill-color-light);
  border-radius: 3px;
  border: 1px solid transparent;
  transition: all 0.15s;
  min-width: 0;
  overflow: hidden;
}


.wb-entry-card:hover {
  background: var(--el-fill-color);
}


.wb-entry-card.disabled {
  opacity: 0.5;
}


.wb-entry-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 2px;
}


.wb-entry-tags {
  display: flex;
  gap: 1px;
  flex-shrink: 0;
}


.wb-entry-title {
  font-weight: 600;
  font-size: 11px;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 2px;
}


.wb-entry-actions {
  display: flex;
  gap: 1px;
  justify-content: flex-end;
}


.worldbook-editor-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
