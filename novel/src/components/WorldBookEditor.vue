<template>
  <div class="worldbook-editor">
    <div class="worldbook-header">
      <div class="worldbook-stats">
        <span class="stat-item">
          <el-icon><Document /></el-icon>
          共 {{ store.globalEntryCount }} 条条目
        </span>
        <span class="stat-item">
          <el-icon><CircleCheck /></el-icon>
          {{ enabledEntriesCount }} 条已启用
        </span>
      </div>
      <div class="worldbook-actions">
        <el-button size="small" @click="handleImport">
          <el-icon><Upload /></el-icon>
          导入
        </el-button>
        <el-button size="small" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
        <el-button type="primary" size="small" @click="createNew">
          <el-icon><Plus /></el-icon>
          新建条目
        </el-button>
      </div>
    </div>

    <div class="worldbook-search">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索关键词、标题或内容..."
        clearable
        :prefix-icon="Search"
      />
    </div>

    <div class="worldbook-list" v-if="filteredEntries.length > 0">
      <div
        v-for="entry in filteredEntries"
        :key="entry.uid"
        :class="['worldbook-item', { disabled: entry.disable }]"
      >
        <div class="item-row">
          <div class="item-status">
            <el-switch
              :model-value="!entry.disable"
              size="small"
              @click.stop
              @change="(val: boolean) => handleToggleDisabled(entry, !val)"
            />
          </div>
          <div class="item-info">
            <div class="item-name">{{ entry.comment || '未命名条目' }}</div>
            <div class="item-keys">
              <el-tag v-if="entry.constant" size="small" type="success">常驻</el-tag>
              <el-tag v-if="entry.disable" size="small" type="danger">禁用</el-tag>
            </div>
          </div>
          <div class="item-actions">
            <el-button size="small" @click="editEntry(entry)" circle>
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(entry)" circle>
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <div class="worldbook-empty" v-else>
      <el-empty :description="searchKeyword ? '未找到匹配的条目' : '暂无世界书条目，点击【新建条目】开始'" />
    </div>

    <el-dialog
      v-model="editorVisible"
      :title="editingId ? '编辑条目' : '新建条目'"
      width="800px"
      top="3vh"
      append-to-body
      destroy-on-close
      :close-on-click-modal="false"
    >
      <div class="entry-editor">
        <div class="editor-field">
          <label class="field-label">标题/备注</label>
          <el-input v-model="form.comment" placeholder="输入条目标题或备注" />
        </div>

        <div class="editor-field">
          <label class="field-label">内容</label>
          <el-input
            v-model="form.content"
            type="textarea"
            :rows="8"
            placeholder="触发后注入的内容，支持 Markdown 格式"
          />
        </div>

        <div class="editor-row">
          <div class="editor-field half">
            <label class="field-label">插入位置</label>
            <el-select v-model="form.position" placeholder="选择插入位置">
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
            <el-input-number v-model="form.order" :min="0" :max="1000" />
          </div>
        </div>

        <div class="editor-row">
          <div class="editor-field half">
            <label class="field-label">扫描深度</label>
            <el-input-number v-model="form.depth" :min="1" :max="999" />
          </div>
        </div>

        <div class="editor-field">
          <label class="field-label">选项</label>
          <div class="checkbox-group">
            <el-checkbox v-model="form.constant">常驻激活</el-checkbox>
            <el-checkbox v-model="form.disable">禁用</el-checkbox>
            <el-checkbox v-model="form.caseSensitive">区分大小写</el-checkbox>
            <el-checkbox v-model="form.matchWholeWords">全字匹配</el-checkbox>
            <el-checkbox v-model="form.excludeRecursion">排除递归</el-checkbox>
            <el-checkbox v-model="form.preventRecursion">阻止递归</el-checkbox>
          </div>
        </div>

        <div class="editor-field">
          <label class="field-label">分组</label>
          <el-input v-model="form.group" placeholder="输入分组名称，同组条目可设置互斥" />
        </div>
      </div>

      <template #footer>
        <el-button @click="editorVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">
          {{ editingId ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <input
      ref="fileInput"
      type="file"
      accept=".json"
      style="display: none"
      @change="handleFileSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document,
  CircleCheck,
  Plus,
  Edit,
  Delete,
  Upload,
  Download,
  Search
} from '@element-plus/icons-vue'
import { useWorldBookStore } from '@/stores/worldbook'
import { createDefaultWorldBookEntry } from '@/types/worldbook'
import type { WorldBookEntry } from '@/types/worldbook'

const store = useWorldBookStore()

const searchKeyword = ref('')
const editorVisible = ref(false)
const editingId = ref<string | null>(null)
const fileInput = ref<HTMLInputElement>()

const form = ref(createDefaultWorldBookEntry())

const enabledEntriesCount = computed(() => 
  store.globalWorldBook.entries.filter(e => !e.disable).length
)

const filteredEntries = computed(() => {
  if (!searchKeyword.value.trim()) {
    return store.globalWorldBook.entries
  }

  const keyword = searchKeyword.value.toLowerCase()
  return store.globalWorldBook.entries.filter(entry => {
    return (
      entry.comment.toLowerCase().includes(keyword) ||
      entry.key.some(k => k.toLowerCase().includes(keyword)) ||
      entry.content.toLowerCase().includes(keyword)
    )
  })
})

function createNew() {
  editingId.value = null
  form.value = createDefaultWorldBookEntry()
  editorVisible.value = true
}

function editEntry(entry: WorldBookEntry) {
  editingId.value = entry.uid
  form.value = { ...entry }
  editorVisible.value = true
}

async function handleSave() {
  if (!form.value.comment.trim()) {
    ElMessage.warning('请输入条目标题')
    return
  }

  if (editingId.value) {
    store.updateEntry(editingId.value, form.value)
    ElMessage.success('条目已更新')
  } else {
    store.addEntry(form.value)
    ElMessage.success('条目已创建')
  }

  editorVisible.value = false
}

async function handleToggleDisabled(entry: WorldBookEntry, disable: boolean) {
  store.updateEntry(entry.uid, { disable })
}

async function handleDelete(entry: WorldBookEntry) {
  try {
    await ElMessageBox.confirm(
      `确定要删除条目"${entry.comment || '未命名条目'}"吗？`,
      '删除确认',
      { type: 'warning' }
    )
    store.deleteEntry(entry.uid)
    ElMessage.success('条目已删除')
  } catch {
    // 取消
  }
}

function handleImport() {
  fileInput.value?.click()
}

async function handleFileSelect(event: Event) {
  const file = (event.target as HTMLInputElement).files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text)
    
    const success = store.importWorldBookData(data)
    if (success) {
      ElMessage.success('世界书导入成功')
    } else {
      ElMessage.error('世界书格式不正确')
    }
  } catch (error) {
    ElMessage.error('导入失败：文件格式错误')
  }

  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

function handleExport() {
  const data = store.exportWorldBookData()
  if (!data) {
    ElMessage.warning('没有可导出的条目')
    return
  }

  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `worldbook-${Date.now()}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success('世界书已导出')
}
</script>

<style scoped>
.worldbook-editor {
  min-height: 400px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.worldbook-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.worldbook-stats {
  display: flex;
  gap: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.worldbook-actions {
  display: flex;
  gap: 8px;
}

.worldbook-search {
  margin-bottom: 8px;
}

.worldbook-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex: 1;
  overflow-y: auto;
}

.worldbook-item {
  background: var(--el-fill-color-light);
  border-radius: 8px;
  overflow: hidden;
  transition: all 0.2s;
}

.worldbook-item.disabled {
  opacity: 0.5;
}

.item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
}

.item-status {
  flex-shrink: 0;
}

.item-info {
  flex: 1;
  min-width: 0;
}

.item-name {
  font-weight: 600;
  font-size: 14px;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.item-keys {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.item-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.worldbook-empty {
  padding: 60px 0;
}

.entry-editor {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.editor-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.editor-field.half {
  flex: 1;
}

.field-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.editor-row {
  display: flex;
  gap: 24px;
}

.checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}
</style>
