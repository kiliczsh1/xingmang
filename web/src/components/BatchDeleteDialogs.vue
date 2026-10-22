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
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'

const props = defineProps<{
  categoryList: { name: string; prompts: Prompt[] }[]
  combinedUncategorizedPrompts: Prompt[]
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

defineExpose({ open, openStandalone })
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
  margin-bottom: 16px;
  padding: 8px 0;
  border-bottom: 1px solid #e8e8e8;
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
