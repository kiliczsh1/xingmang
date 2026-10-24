<template>
<!-- 批量删除对话框 -->
<el-dialog
  v-model="batchDeleteDialogVisible"
  title="批量删除提示词"
  width="800px"
  :close-on-click-modal="false"
  class="batch-delete-dialog"
  append-to-body
>
  <div class="batch-delete-content">
    <div class="batch-delete-hint">
      <el-alert
        :title="'正在删除卡包' + currentCategory + '中的提示词'"
        type="warning"
        :closable="false"
        show-icon
      >
        勾选需要删除的卡片，然后点击删除按钮。此操作不可恢复。
      </el-alert>
    </div>
    <div class="batch-delete-select-all">
      <el-checkbox
        v-model="selectAllBatchDelete"
        @change="handleSelectAllBatchDelete"
      >
        全选 ({{ selectedBatchDeleteCards.length }}/{{ currentCategoryPrompts.length }})
      </el-checkbox>
    </div>
    <div v-if="currentCategoryPrompts.length > 0" class="batch-delete-cards-grid">
      <div
        v-for="prompt in currentCategoryPrompts"
        :key="prompt.id"
        class="batch-delete-card-item"
        :class="{ selected: selectedBatchDeleteCards.includes(prompt.id) }"
        @click="toggleBatchDeleteCard(prompt.id)"
      >
        <el-checkbox
          :model-value="selectedBatchDeleteCards.includes(prompt.id)"
          @click.stop
          @change="toggleBatchDeleteCard(prompt.id)"
        />
        <div class="batch-delete-card-content">
          <div class="batch-delete-card-name">{{ prompt.name }}</div>
          <div class="batch-delete-card-preview">
            <template v-if="prompt.card_type === 'encrypted'">
              <el-icon><Lock /></el-icon> 已加密
            </template>
            <template v-else>
              {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
            </template>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="batch-delete-empty">
      <el-empty description="该卡包暂无提示词" />
    </div>
  </div>
  <template #footer>
    <el-button @click="batchDeleteDialogVisible = false">取消</el-button>
    <el-button
      type="danger"
      :disabled="selectedBatchDeleteCards.length === 0"
      @click="executeBatchDelete"
    >
      删除 ({{ selectedBatchDeleteCards.length }})
    </el-button>
  </template>
</el-dialog>

<!-- 独立卡片批量删除对话框 -->
<el-dialog
  v-model="standaloneBatchDeleteDialogVisible"
  title="批量删除独立卡片"
  width="800px"
  :close-on-click-modal="false"
  class="batch-delete-dialog"
  append-to-body
>
  <div class="batch-delete-content">
    <div class="batch-delete-hint">
      <el-alert
        title="正在删除独立卡片"
        type="warning"
        :closable="false"
        show-icon
      >
        勾选需要删除的独立卡片，然后点击删除按钮。此操作不可恢复。
      </el-alert>
    </div>
    <div class="batch-delete-select-all">
      <el-checkbox
        v-model="selectAllStandaloneBatchDelete"
        @change="handleSelectAllStandaloneBatchDelete"
      >
        全选 ({{ selectedStandaloneBatchDeleteCards.length }}/{{ combinedUncategorizedPrompts.length }})
      </el-checkbox>
    </div>
    <div v-if="combinedUncategorizedPrompts.length > 0" class="batch-delete-cards-grid">
      <div
        v-for="prompt in combinedUncategorizedPrompts"
        :key="prompt.id"
        class="batch-delete-card-item"
        :class="{ selected: selectedStandaloneBatchDeleteCards.includes(prompt.id) }"
        @click="toggleStandaloneBatchDeleteCard(prompt.id)"
      >
        <el-checkbox
          :model-value="selectedStandaloneBatchDeleteCards.includes(prompt.id)"
          @click.stop
          @change="toggleStandaloneBatchDeleteCard(prompt.id)"
        />
        <div class="batch-delete-card-content">
          <div class="batch-delete-card-name">
            {{ prompt.name }}
            <el-tag v-if="prompt.category === '拆书-未分类'" size="small" type="warning" style="margin-left: 6px;">拆书</el-tag>
          </div>
          <div class="batch-delete-card-preview">
            <template v-if="prompt.card_type === 'encrypted'">
              <el-icon><Lock /></el-icon> 已加密
            </template>
            <template v-else>
              {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
            </template>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="batch-delete-empty">
      <el-empty description="暂无独立卡片" />
    </div>
  </div>
  <template #footer>
    <el-button @click="standaloneBatchDeleteDialogVisible = false">取消</el-button>
    <el-button
      type="danger"
      :disabled="selectedStandaloneBatchDeleteCards.length === 0"
      @click="executeStandaloneBatchDelete"
    >
      删除 ({{ selectedStandaloneBatchDeleteCards.length }})
    </el-button>
  </template>
</el-dialog>

<!-- 全库批量删除对话框（跨卡包，从快捷中心 → 管理 进入） -->
<el-dialog
  v-model="allBatchDialogVisible"
  title="批量删除提示词"
  width="880px"
  :close-on-click-modal="false"
  class="batch-delete-dialog"
  append-to-body
>
  <div class="batch-delete-content">
    <div class="batch-delete-hint">
      <el-alert
        title="跨卡包删除提示词"
        type="error"
        :closable="false"
        show-icon
      >
        可跨卡包、独立卡片、拆书卡包勾选，删除前会再次确认。此操作不可恢复，卡包本身不会被删除。
      </el-alert>
    </div>
    <div class="all-batch-filter">
      <el-input
        v-model="allBatchKeyword"
        placeholder="按名称或内容筛选…"
        clearable
        :prefix-icon="Search"
        class="all-batch-search"
      />
      <el-select v-model="allBatchPack" placeholder="全部来源" clearable class="all-batch-pack">
        <el-option label="全部来源" value="" />
        <el-option
          v-for="pack in allBatchPacks"
          :key="pack.name"
          :label="`${pack.name}（${pack.count}）`"
          :value="pack.name"
        />
      </el-select>
    </div>
    <div class="batch-delete-select-all">
      <el-checkbox
        :model-value="allBatchAllSelected"
        :indeterminate="allBatchPartSelected"
        @change="handleSelectAllAllBatch"
      >
        全选当前筛选 ({{ allBatchVisibleSelectedCount }}/{{ allBatchFilteredPrompts.length }})
      </el-checkbox>
      <span v-if="selectedAllBatchCards.length > 0" class="all-batch-total">
        已选 {{ selectedAllBatchCards.length }} 张
        <button class="all-batch-clear" @click="clearAllBatchSelection">清空</button>
      </span>
    </div>
    <div v-if="allBatchFilteredPrompts.length > 0" class="batch-delete-cards-grid">
      <div
        v-for="prompt in allBatchFilteredPrompts"
        :key="prompt.id"
        class="batch-delete-card-item"
        :class="{ selected: selectedAllBatchCards.includes(prompt.id) }"
        @click="toggleAllBatchCard(prompt.id)"
      >
        <el-checkbox
          :model-value="selectedAllBatchCards.includes(prompt.id)"
          @click.stop
          @change="toggleAllBatchCard(prompt.id)"
        />
        <div class="batch-delete-card-content">
          <div class="batch-delete-card-name">
            {{ prompt.name }}
            <el-icon v-if="prompt.card_type === 'encrypted'" class="all-batch-lock"><Lock /></el-icon>
          </div>
          <div class="all-batch-pack-tag">{{ prompt.category || '默认' }}</div>
          <div class="batch-delete-card-preview">
            <template v-if="prompt.card_type === 'encrypted'">
              <el-icon><Lock /></el-icon> 已加密
            </template>
            <template v-else>
              {{ (prompt.content || '').slice(0, 30) }}{{ (prompt.content || '').length > 30 ? '...' : '' }}
            </template>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="batch-delete-empty">
      <el-empty :description="allPrompts.length === 0 ? '暂无提示词' : '没有匹配的提示词'" />
    </div>
  </div>
  <template #footer>
    <el-button @click="allBatchDialogVisible = false">取消</el-button>
    <el-button
      type="danger"
      :disabled="selectedAllBatchCards.length === 0"
      :loading="allBatchDeleting"
      @click="executeAllBatchDelete"
    >
      删除 ({{ selectedAllBatchCards.length }})
    </el-button>
  </template>
</el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'

const props = defineProps<{
  categoryList: { name: string; prompts: Prompt[] }[]
  combinedUncategorizedPrompts: Prompt[]
  /** 全部提示词（不分组），供跨卡包批量删除使用 */
  allPrompts: Prompt[]
}>()

const emit = defineEmits<{ (e: 'deleted'): void }>()

// 批量删除相关状态
const batchDeleteDialogVisible = ref(false)
const currentCategory = ref<string>('')
const selectedBatchDeleteCards = ref<number[]>([])
const selectAllBatchDelete = ref(false)

// 独立卡片批量删除相关状态
const standaloneBatchDeleteDialogVisible = ref(false)
const selectedStandaloneBatchDeleteCards = ref<number[]>([])
const selectAllStandaloneBatchDelete = ref(false)

const getCategoryPromptsByName = (categoryName: string) => {
  if (!categoryName) return []
  const category = props.categoryList.find(cat => cat.name === categoryName)
  return category ? category.prompts : []
}

// 当前卡包的提示词列表（用于批量删除）
const currentCategoryPrompts = computed(() => {
  return getCategoryPromptsByName(currentCategory.value)
})

// 全选/取消全选批量删除卡片
const handleSelectAllBatchDelete = (val: boolean) => {
  if (val) {
    selectedBatchDeleteCards.value = currentCategoryPrompts.value.map(p => p.id)
  } else {
    selectedBatchDeleteCards.value = []
  }
}

// 切换单个批量删除卡片选择
const toggleBatchDeleteCard = (id: number) => {
  const index = selectedBatchDeleteCards.value.indexOf(id)
  if (index > -1) {
    selectedBatchDeleteCards.value.splice(index, 1)
  } else {
    selectedBatchDeleteCards.value.push(id)
  }
  selectAllBatchDelete.value = selectedBatchDeleteCards.value.length === currentCategoryPrompts.value.length
}

// 执行批量删除
const executeBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedBatchDeleteCards.value.length} 张卡片吗？此操作不可恢复。`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 批量删除选中的卡片
    const deletePromises = selectedBatchDeleteCards.value.map(id => promptAPI.delete(id))
    await Promise.all(deletePromises)
    
    ElMessage.success(`成功删除 ${selectedBatchDeleteCards.value.length} 张卡片`)
    batchDeleteDialogVisible.value = false
    emit('deleted')
  } catch (error) {
    // 取消删除
  }
}

// 全选/取消全选独立卡片批量删除
const handleSelectAllStandaloneBatchDelete = (val: boolean) => {
  if (val) {
    selectedStandaloneBatchDeleteCards.value = props.combinedUncategorizedPrompts.map(p => p.id)
  } else {
    selectedStandaloneBatchDeleteCards.value = []
  }
}

// 切换单个独立卡片批量删除选择
const toggleStandaloneBatchDeleteCard = (id: number) => {
  const index = selectedStandaloneBatchDeleteCards.value.indexOf(id)
  if (index > -1) {
    selectedStandaloneBatchDeleteCards.value.splice(index, 1)
  } else {
    selectedStandaloneBatchDeleteCards.value.push(id)
  }
  selectAllStandaloneBatchDelete.value = selectedStandaloneBatchDeleteCards.value.length === props.combinedUncategorizedPrompts.length
}

// 执行独立卡片批量删除
const executeStandaloneBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${selectedStandaloneBatchDeleteCards.value.length} 张独立卡片吗？此操作不可恢复。`,
      '批量删除独立卡片',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    const deletePromises = selectedStandaloneBatchDeleteCards.value.map(id => promptAPI.delete(id))
    await Promise.all(deletePromises)
    
    ElMessage.success(`成功删除 ${selectedStandaloneBatchDeleteCards.value.length} 张独立卡片`)
    standaloneBatchDeleteDialogVisible.value = false
    emit('deleted')
  } catch (error) {
    // 取消删除
  }
}

// 打开批量删除对话框
const open = (categoryName: string) => {
  currentCategory.value = categoryName
  selectedBatchDeleteCards.value = []
  selectAllBatchDelete.value = false
  batchDeleteDialogVisible.value = true
}

// 打开独立卡片批量删除对话框
const openStandalone = () => {
  selectedStandaloneBatchDeleteCards.value = []
  selectAllStandaloneBatchDelete.value = false
  standaloneBatchDeleteDialogVisible.value = true
}

// ===== 全库批量删除（跨卡包） =====
const allBatchDialogVisible = ref(false)
const allBatchKeyword = ref('')
const allBatchPack = ref('')
const selectedAllBatchCards = ref<number[]>([])
const allBatchDeleting = ref(false)

// 占位卡片不是真实数据，一律不展示、不可删
const allBatchSource = computed(() => props.allPrompts.filter(p => p.name !== '__category_placeholder__'))

// 来源分组（按数量降序），用于下拉筛选
const allBatchPacks = computed(() => {
  const map = new Map<string, number>()
  allBatchSource.value.forEach(p => {
    const c = p.category || '默认'
    map.set(c, (map.get(c) || 0) + 1)
  })
  return [...map.entries()]
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count)
})

const allBatchFilteredPrompts = computed(() => {
  const kw = allBatchKeyword.value.trim().toLowerCase()
  const pack = allBatchPack.value
  return allBatchSource.value.filter(p => {
    const c = p.category || '默认'
    if (pack && c !== pack) return false
    if (!kw) return true
    return (p.name || '').toLowerCase().includes(kw) || (p.content || '').toLowerCase().includes(kw)
  })
})

const allBatchVisibleSelectedCount = computed(
  () => allBatchFilteredPrompts.value.filter(p => selectedAllBatchCards.value.includes(p.id)).length
)
const allBatchAllSelected = computed(
  () => allBatchFilteredPrompts.value.length > 0 &&
    allBatchVisibleSelectedCount.value === allBatchFilteredPrompts.value.length
)
const allBatchPartSelected = computed(
  () => allBatchVisibleSelectedCount.value > 0 && !allBatchAllSelected.value
)

const toggleAllBatchCard = (id: number) => {
  const index = selectedAllBatchCards.value.indexOf(id)
  if (index > -1) {
    selectedAllBatchCards.value.splice(index, 1)
  } else {
    selectedAllBatchCards.value.push(id)
  }
}

// 全选只作用于当前筛选结果，不会把筛选外的卡片带进来
const handleSelectAllAllBatch = (val: boolean | string | number) => {
  const ids = allBatchFilteredPrompts.value.map(p => p.id)
  if (val) {
    const set = new Set(selectedAllBatchCards.value)
    ids.forEach(id => set.add(id))
    selectedAllBatchCards.value = [...set]
  } else {
    const drop = new Set(ids)
    selectedAllBatchCards.value = selectedAllBatchCards.value.filter(id => !drop.has(id))
  }
}

const clearAllBatchSelection = () => {
  selectedAllBatchCards.value = []
}

const executeAllBatchDelete = async () => {
  const ids = [...selectedAllBatchCards.value]
  if (ids.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${ids.length} 张提示词吗？删除后不可恢复。`,
      '二次确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
        confirmButtonClass: 'el-button--danger'
      }
    )
  } catch {
    return
  }

  allBatchDeleting.value = true
  let ok = 0
  let failed = 0
  try {
    // 分批并发，避免一次性打满连接
    const CHUNK = 8
    for (let i = 0; i < ids.length; i += CHUNK) {
      const slice = ids.slice(i, i + CHUNK)
      const results = await Promise.all(
        slice.map(id => promptAPI.delete(id).catch(() => null))
      )
      results.forEach(r => {
        if (r && r.success) ok++
        else failed++
      })
    }
  } finally {
    allBatchDeleting.value = false
  }

  const done = new Set(ids)
  selectedAllBatchCards.value = selectedAllBatchCards.value.filter(id => !done.has(id))

  if (ok > 0) ElMessage.success(`成功删除 ${ok} 张提示词${failed ? `，${failed} 张失败` : ''}`)
  else ElMessage.error(`删除失败${failed ? `（${failed} 张）` : ''}`)

  allBatchDialogVisible.value = false
  emit('deleted')
}

// 打开全库批量删除对话框
const openAll = () => {
  selectedAllBatchCards.value = []
  allBatchKeyword.value = ''
  allBatchPack.value = ''
  allBatchDialogVisible.value = true
}

defineExpose({ open, openStandalone, openAll })
</script>

<style scoped>
/* 批量删除弹窗样式 */
.batch-delete-content {
  max-height: 500px;
  overflow-y: auto;
}

.batch-delete-hint {
  margin-bottom: 16px;
}

.batch-delete-select-all {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  padding: 8px 0;
  border-bottom: 1px solid #e8e8e8;
}

/* ---- 全库批量删除：筛选栏 / 来源标签 ---- */
.all-batch-filter {
  display: flex;
  gap: 10px;
  margin-bottom: 4px;
}

.all-batch-search {
  flex: 1;
  min-width: 0;
}

.all-batch-pack {
  width: 220px;
  flex: 0 0 auto;
}

.all-batch-total {
  font-size: 12px;
  color: #8c8c8c;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.all-batch-clear {
  border: none;
  background: transparent;
  color: #409eff;
  font-size: 12px;
  cursor: pointer;
  padding: 0;
}

.all-batch-clear:hover {
  text-decoration: underline;
}

.all-batch-pack-tag {
  display: inline-block;
  max-width: 100%;
  margin-bottom: 4px;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 10px;
  line-height: 1.5;
  color: #7c3aed;
  background: rgba(124, 58, 237, 0.1);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.all-batch-lock {
  margin-left: 4px;
  font-size: 12px;
  color: #8c8c8c;
  vertical-align: -1px;
}

/* ---- 深色主题 ---- */
:root[data-theme='dark'] .batch-delete-select-all {
  border-color: #2e3650;
}
:root[data-theme='dark'] .batch-delete-card-item {
  background: #1e2434;
  border-color: #2e3650;
}
:root[data-theme='dark'] .batch-delete-card-item.selected {
  background: rgba(239, 68, 68, 0.12);
  border-color: #ef4444;
}
:root[data-theme='dark'] .batch-delete-card-name {
  color: #e5e7eb;
}
:root[data-theme='dark'] .batch-delete-card-preview {
  color: #9ca3af;
}
:root[data-theme='dark'] .all-batch-total {
  color: #9ca3af;
}
:root[data-theme='dark'] .all-batch-pack-tag {
  color: #c4b5fd;
  background: rgba(139, 92, 246, 0.18);
}

.batch-delete-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.batch-delete-card-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.batch-delete-card-item:hover {
  border-color: #ff4d4f;
  box-shadow: 0 2px 8px rgba(255, 77, 79, 0.15);
}

.batch-delete-card-item.selected {
  border-color: #ff4d4f;
  background: #fff1f0;
}

.batch-delete-card-content {
  flex: 1;
  min-width: 0;
}

.batch-delete-card-name {
  font-weight: 600;
  font-size: 13px;
  color: #262626;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.batch-delete-card-preview {
  font-size: 11px;
  color: #8c8c8c;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.batch-delete-empty {
  padding: 40px;
}
</style>
