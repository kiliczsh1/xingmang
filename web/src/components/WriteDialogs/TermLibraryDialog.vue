<template>
  <el-dialog
    :model-value="visible"
    fullscreen
    append-to-body
    destroy-on-close
    class="xm-term-library"
    @update:model-value="emit('update:visible', $event)"
  >
    <template #header>
      <div class="xm-tl-header">
        <span class="xm-tl-title">词条卡</span>
        <span class="xm-tl-sub">{{ worldBookName }}</span>
        <span class="xm-tl-link">教程</span>
        <el-button size="small" type="warning" @click="createDraft">新建</el-button>
      </div>
    </template>

    <div class="xm-tl-body">
      <!-- 左：列表 -->
      <aside class="xm-tl-list">
        <el-input v-model="keyword" size="small" placeholder="搜索词条..." clearable />
        <div class="xm-tl-list-scroll">
          <div v-if="filteredEntries.length === 0" class="xm-tl-empty">暂无词条</div>
          <div
            v-for="e in filteredEntries"
            :key="e.uid"
            class="xm-tl-card"
            :class="{ 'is-active': current?.uid === e.uid, 'is-disabled': e.disable }"
            @click="selectEntry(e)"
          >
            <span class="xm-tl-card-name">{{ e.comment || e.name || '(未命名)' }}</span>
            <span class="xm-tl-card-tag" :class="{ 'is-off': e.disable }">{{ e.disable ? '已停用' : '启用中' }}</span>
          </div>
        </div>
      </aside>

      <!-- 右：详情 -->
      <section class="xm-tl-detail">
        <template v-if="current">
          <div class="xm-tl-detail-head">
            <span class="xm-tl-detail-title">词条详情</span>
            <span class="xm-tl-detail-id">UID: {{ current.uid }}</span>
            <el-button size="small" text @click="current = null">关闭详情</el-button>
          </div>
          <div class="xm-tl-detail-form">
            <div class="xm-tl-field">
              <label class="xm-tl-label"><span class="xm-tl-req">*</span>词条名称</label>
              <el-input v-model="form.comment" maxlength="50" show-word-limit placeholder="例如：青石巷" />
            </div>
            <div class="xm-tl-field">
              <label class="xm-tl-label">关键词</label>
              <p class="xm-tl-hint">正文中出现这些词时，词条内容会注入 AI 上下文（多个用逗号分隔）</p>
              <el-input v-model="keyText" placeholder="例如：青石巷,巷口" />
            </div>
            <div class="xm-tl-field">
              <label class="xm-tl-label">次要关键词</label>
              <el-input v-model="keySecondaryText" placeholder="可选，配合关键词使用（多个用逗号分隔）" />
            </div>
            <div class="xm-tl-field">
              <label class="xm-tl-label">词条内容</label>
              <p class="xm-tl-hint">词条被触发时注入的设定内容</p>
              <el-input v-model="form.content" type="textarea" :rows="10" placeholder="输入与该词条相关的设定、背景等信息..." />
            </div>
            <div class="xm-tl-field xm-tl-field--row">
              <label class="xm-tl-label">启用</label>
              <el-switch v-model="enabled" />
            </div>
          </div>
          <div class="xm-tl-detail-footer">
            <el-button type="warning" @click="saveEntry">保存词条</el-button>
            <el-button type="danger" plain @click="removeEntry">删除词条</el-button>
            <el-button @click="exportEntry">导出词条</el-button>
          </div>
        </template>
        <div v-else class="xm-tl-empty xm-tl-empty--detail">从左侧选择词条，或点击「新建」创建词条卡</div>
      </section>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWorldBookStore } from '@/stores/worldbook'
import type { WorldBookEntry } from '@/types/worldbook'

defineOptions({ name: 'TermLibraryDialog' })

const props = defineProps<{
  visible: boolean
  bookId: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'changed'): void
}>()

const worldBookStore = useWorldBookStore()
const keyword = ref('')
const current = ref<WorldBookEntry | null>(null)
const form = ref<{ comment: string; content: string }>({ comment: '', content: '' })
const keyText = ref('')
const keySecondaryText = ref('')
const enabled = ref(true)
const isNewDraft = ref(false)

const worldBookName = computed(() => `世界书：${worldBookStore.getBookWorldBook(String(props.bookId))?.name || '默认世界书'}`)

const entries = computed(() => worldBookStore.getBookWorldBook(String(props.bookId))?.entries || [])

const filteredEntries = computed(() =>
  entries.value.filter(e => {
    const name = e.comment || e.name || ''
    return !keyword.value.trim() || name.includes(keyword.value.trim())
  })
)

watch(() => props.visible, (val) => {
  if (!val) return
  current.value = null
  keyword.value = ''
})

const selectEntry = (e: WorldBookEntry) => {
  current.value = { ...e }
  isNewDraft.value = false
  form.value = { comment: e.comment || '', content: e.content || '' }
  keyText.value = (e.key || []).join(', ')
  keySecondaryText.value = (e.keysecondary || []).join(', ')
  enabled.value = !e.disable
}

const createDraft = () => {
  current.value = null
  isNewDraft.value = true
  form.value = { comment: '', content: '' }
  keyText.value = ''
  keySecondaryText.value = ''
  enabled.value = true
  // 用一个临时对象进入编辑态
  current.value = { uid: 'draft', name: '', key: [], keysecondary: [], keyword: '', keywords: '', comment: '', content: '', constant: false, selective: false, order: 100, position: 0, disable: false, excludeRecursion: false, preventRecursion: false, delayUntilRecursion: false, probability: 100, useProbability: true, depth: 4, selectLogic: 0, group: '', groupOverride: false, groupWeight: 100, scanDepth: null, caseSensitive: false, matchWholeWords: false, useGroupScoring: false, automationId: '', role: null, vectorized: false, displayIndex: 0, sticky: 0, cooldown: 0, delay: 0 } as WorldBookEntry
}

const saveEntry = async () => {
  if (!current.value) return
  if (!form.value.comment.trim()) {
    ElMessage.warning('请输入词条名称')
    return
  }
  const payload: Partial<WorldBookEntry> = {
    comment: form.value.comment.trim(),
    name: form.value.comment.trim(),
    content: form.value.content,
    key: keyText.value.split(/[,，]/).map(s => s.trim()).filter(Boolean),
    keysecondary: keySecondaryText.value.split(/[,，]/).map(s => s.trim()).filter(Boolean),
    disable: !enabled.value
  }
  if (isNewDraft.value) {
    worldBookStore.addEntry(payload, String(props.bookId))
    ElMessage.success('词条已创建')
  } else {
    worldBookStore.updateEntry(current.value.uid, payload, String(props.bookId))
    ElMessage.success('词条已保存')
  }
  isNewDraft.value = false
  // 重新选中保存后的条目
  const saved = entries.value.find(e => (e.comment || '') === payload.comment)
  if (saved) selectEntry(saved)
  emit('changed')
}

const removeEntry = async () => {
  if (!current.value || isNewDraft.value) {
    current.value = null
    return
  }
  try {
    await ElMessageBox.confirm(`确定删除词条「${current.value.comment || current.value.name}」？`, '删除词条', { type: 'warning' })
  } catch {
    return
  }
  worldBookStore.deleteEntry(current.value.uid, String(props.bookId))
  ElMessage.success('词条已删除')
  current.value = null
  emit('changed')
}

const exportEntry = () => {
  if (!current.value) return
  const blob = new Blob([JSON.stringify(current.value, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `词条卡_${current.value.comment || '未命名'}.json`
  a.click()
  URL.revokeObjectURL(url)
}
</script>

<style scoped>
.xm-tl-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.xm-tl-title {
  font-size: 17px;
  font-weight: 600;
  color: #1f2329;
}

.xm-tl-sub {
  font-size: 12px;
  color: #a3a8b3;
}

.xm-tl-link {
  font-size: 12px;
  color: #8a8f99;
  cursor: pointer;
}

.xm-tl-body {
  display: flex;
  gap: 14px;
  height: calc(100vh - 140px);
}

.xm-tl-list {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-right: 1px solid #eeeeea;
  padding-right: 12px;
}

.xm-tl-list-scroll {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.xm-tl-empty {
  font-size: 12px;
  color: #a3a8b3;
  text-align: center;
  padding: 40px 0;
}

.xm-tl-card {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border: 1px solid #eef0f2;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.12s ease;
}

.xm-tl-card:hover {
  border-color: #d8dcdf;
  background: #fafbfc;
}

.xm-tl-card.is-active {
  border-color: #d97706;
  background: #fff7ed;
}

.xm-tl-card-name {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  font-weight: 500;
  color: #1f2329;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-tl-card-tag {
  flex-shrink: 0;
  font-size: 10.5px;
  color: #0ca678;
  background: #ecfaf3;
  border-radius: 4px;
  padding: 1px 6px;
}

.xm-tl-card-tag.is-off {
  color: #a3a8b3;
  background: #f3f5f7;
}

.xm-tl-detail {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.xm-tl-detail-head {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.xm-tl-detail-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2329;
}

.xm-tl-detail-id {
  font-size: 11px;
  color: #a3a8b3;
  flex: 1;
}

.xm-tl-detail-form {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-right: 4px;
}

.xm-tl-field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.xm-tl-field--row {
  flex-direction: row;
  align-items: center;
  gap: 10px;
}

.xm-tl-label {
  font-size: 12.5px;
  font-weight: 500;
  color: #374151;
}

.xm-tl-req {
  color: #e54d42;
  margin-right: 2px;
}

.xm-tl-hint {
  margin: 0;
  font-size: 11px;
  color: #a3a8b3;
}

.xm-tl-detail-footer {
  display: flex;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid #eeeeea;
}

.xm-tl-empty--detail {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

:root[data-theme='dark'] .xm-tl-title,
:root[data-theme='dark'] .xm-tl-detail-title,
:root[data-theme='dark'] .xm-tl-card-name,
:root[data-theme='dark'] .xm-tl-label {
  color: #e6edf3;
}

:root[data-theme='dark'] .xm-tl-list,
:root[data-theme='dark'] .xm-tl-detail-footer {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-tl-card {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-tl-card:hover {
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .xm-tl-card.is-active {
  background: rgba(217, 119, 6, 0.15);
}

:root[data-theme='dark'] .xm-tl-card-tag {
  background: rgba(12, 166, 120, 0.18);
}

:root[data-theme='dark'] .xm-tl-card-tag.is-off {
  background: rgba(148, 163, 184, 0.15);
}

:root[data-theme='dark'] .xm-tl-hint,
:root[data-theme='dark'] .xm-tl-sub,
:root[data-theme='dark'] .xm-tl-detail-id,
:root[data-theme='dark'] .xm-tl-empty {
  color: #6b7688;
}
</style>
