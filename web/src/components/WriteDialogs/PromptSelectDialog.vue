<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="600px"
    destroy-on-close
    class="prompt-select-dialog xm-dialog"
    align-center
    append-to-body
    center
    :modal="modal"
  >
    <div class="prompt-select-content">
      <div class="prompt-select-toolbar">
        <div class="prompt-select-search-row">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索提示词..."
            prefix-icon="Search"
            clearable
            class="search-input"
          />
          <div class="prompt-select-count">共 {{ filteredPrompts.length }} 条</div>
        </div>
        <div v-if="showCategories" class="category-tabs-wrapper">
          <el-button
            class="category-scroll-btn"
            size="small"
            circle
            @click="scrollCategory('left')"
            :disabled="!canScrollLeft"
          >
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <div class="category-tabs" ref="categoryTabsRef" @scroll="handleCategoryScroll">
            <div class="category-tabs-inner">
              <div
                v-for="category in categories"
                :key="category"
                class="category-tab"
                :class="{ active: selectedCategory === category || (category === '全部' && selectedCategory === 'all') }"
                @click="selectedCategory = category === '全部' ? 'all' : category"
              >
                {{ category }}
              </div>
            </div>
          </div>
          <el-button
            class="category-scroll-btn"
            size="small"
            circle
            @click="scrollCategory('right')"
            :disabled="!canScrollRight"
          >
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </div>
      </div>
      <div class="prompt-select-list">
        <div
          v-for="prompt in filteredPrompts"
          :key="prompt.id"
          class="prompt-select-item"
          :class="{ selected: isPromptSelected(prompt.id) }"
          @click="togglePrompt(prompt)"
        >
          <el-checkbox
            v-if="mode === 'multi'"
            :model-value="isPromptSelected(prompt.id)"
            @click.stop
            @change="togglePrompt(prompt)"
          />
          <el-radio
            v-else
            :model-value="isPromptSelected(prompt.id)"
            @click.stop
          />
          <div class="prompt-select-info">
            <div class="prompt-select-name">{{ prompt.name }}</div>
            <div class="prompt-select-desc">{{ prompt.content?.slice(0, 80) }}{{ prompt.content?.length > 80 ? '...' : '' }}</div>
          </div>
          <el-tag size="small" type="info" class="prompt-category-tag">{{ prompt.category }}</el-tag>
        </div>
        <div v-if="filteredPrompts.length === 0" class="empty-prompts">
          <el-icon :size="32"><Document /></el-icon>
          <p>暂无匹配的提示词</p>
        </div>
      </div>
    </div>
    <template v-if="showFooter" #footer>
      <el-button v-if="mode === 'single'" @click="handleClear">清除选择</el-button>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleConfirm">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Prompt } from '@/types'

const props = withDefaults(defineProps<{
  title?: string
  prompts: Prompt[]
  mode?: 'multi' | 'single'
  modal?: boolean
  categoryFilter?: string
  showCategories?: boolean
  showFooter?: boolean
}>(), {
  title: '选择提示词',
  mode: 'multi',
  modal: false,
  categoryFilter: '',
  showCategories: true,
  showFooter: false,
})

const emit = defineEmits<{
  confirm: []
}>()

const visible = defineModel<boolean>('visible')
const modelValue = defineModel<number[] | number>('modelValue')

const searchKeyword = ref('')
const selectedCategory = ref('all')
const categoryTabsRef = ref<HTMLElement>()
const canScrollLeft = ref(false)
const canScrollRight = ref(false)

// 计算分类列表
const categories = computed(() => {
  const categorySet = new Set<string>(['全部'])
  const source = props.categoryFilter
    ? props.prompts.filter(p => p.category === props.categoryFilter)
    : props.prompts
  source.forEach(prompt => {
    if (prompt.category) {
      categorySet.add(prompt.category)
    }
  })
  return Array.from(categorySet)
})

// 过滤后的提示词
const filteredPrompts = computed(() => {
  let result = props.prompts

  // 按分类预过滤（用于 creative2 弹窗）
  if (props.categoryFilter) {
    result = result.filter(prompt => prompt.category === props.categoryFilter)
  }

  // 按分类过滤（分类标签选择）
  if (!props.categoryFilter && selectedCategory.value !== 'all') {
    result = result.filter(prompt => prompt.category === selectedCategory.value)
  }

  // 按关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(prompt =>
      prompt.name.toLowerCase().includes(keyword) ||
      prompt.content.toLowerCase().includes(keyword)
    )
  }

  return result
})

// 判断是否选中
const isPromptSelected = (promptId: number): boolean => {
  if (props.mode === 'multi') {
    return (modelValue.value as number[])?.includes(promptId) ?? false
  }
  return modelValue.value === promptId
}

// 切换选中状态
const togglePrompt = (prompt: Prompt) => {
  if (props.mode === 'multi') {
    const ids = [...((modelValue.value as number[]) || [])]
    const index = ids.indexOf(prompt.id)
    if (index > -1) {
      ids.splice(index, 1)
    } else {
      ids.push(prompt.id)
    }
    modelValue.value = ids
  } else {
    // single mode: toggle off if already selected
    modelValue.value = modelValue.value === prompt.id ? 0 : prompt.id
  }
}

// 分类滚动
const scrollCategory = (direction: 'left' | 'right') => {
  if (categoryTabsRef.value) {
    const scrollAmount = 150
    if (direction === 'left') {
      categoryTabsRef.value.scrollLeft -= scrollAmount
    } else {
      categoryTabsRef.value.scrollLeft += scrollAmount
    }
  }
}

// 检查分类标签是否可以滚动
const checkCategoryScroll = () => {
  if (categoryTabsRef.value) {
    canScrollLeft.value = categoryTabsRef.value.scrollLeft > 0
    canScrollRight.value = categoryTabsRef.value.scrollLeft < (categoryTabsRef.value.scrollWidth - categoryTabsRef.value.clientWidth)
  }
}

const handleCategoryScroll = () => {
  checkCategoryScroll()
}

const handleConfirm = () => {
  visible.value = false
  emit('confirm')
}

const handleClear = () => {
  modelValue.value = 0
  visible.value = false
}

// 弹窗打开时重置搜索和分类
watch(visible, (val) => {
  if (val) {
    searchKeyword.value = ''
    selectedCategory.value = 'all'
  }
})
</script>

<style scoped>
.prompt-select-dialog :deep(.el-dialog) {
  position: fixed;
  margin: 0 !important;
  max-height: calc(100vh - 48px);
  overflow: hidden;
  border: 1px solid var(--xm-line, #e5e7eb);
  border-radius: var(--xm-radius, 16px);
  background: var(--xm-surface, #fff);
  box-shadow: var(--xm-shadow-dialog, 0 4px 24px rgba(0, 0, 0, 0.08));
}

:deep(.prompt-select-dialog-modal) {
  background: var(--xm-veil, rgba(0, 0, 0, 0.3)) !important;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.prompt-select-dialog :deep(.el-dialog__header) {
  margin-right: 0;
  padding: 16px 20px;
  border-bottom: 1px solid var(--xm-surface-2, #f3f4f6);
}

.prompt-select-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 600;
  color: var(--xm-ink, #1f2937);
}

.prompt-select-dialog :deep(.el-dialog__headerbtn) {
  top: 16px;
  right: 16px;
}

.prompt-select-dialog :deep(.el-dialog__close) {
  color: var(--xm-ink-3, #9ca3af);
}

.prompt-select-dialog :deep(.el-dialog__close:hover) {
  color: var(--xm-ink-2, #6b7280);
}

.prompt-select-dialog :deep(.el-dialog__body) {
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 140px);
  overflow: hidden;
  padding: 16px 20px;
}

.prompt-select-dialog :deep(.el-dialog__footer) {
  padding: 12px 20px 16px;
  border-top: 1px solid var(--xm-surface-2, #f3f4f6);
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button) {
  min-width: 80px;
  border-radius: 6px;
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--default) {
  border-color: var(--xm-line, #e5e7eb);
  background: var(--xm-surface, #fff);
  color: var(--xm-ink-2, #4b5563);
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--default:hover) {
  border-color: var(--xm-line-strong, #d1d5db);
  background: var(--xm-surface-2, #f9fafb);
  color: var(--xm-ink, #374151);
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--primary) {
  border-color: var(--xm-accent, #00c9a7);
  background: var(--xm-accent, #00c9a7);
}

.prompt-select-dialog :deep(.el-dialog__footer .el-button--primary:hover) {
  background: #00b896;
  border-color: #00b896;
}

.prompt-select-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
}

.prompt-select-toolbar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--xm-line, #e5e7eb);
  background: var(--xm-surface-2, #f9fafb);
}

.prompt-select-search-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.prompt-select-count {
  flex-shrink: 0;
  padding: 6px 10px;
  border-radius: 4px;
  background: var(--xm-surface, #fff);
  border: 1px solid var(--xm-line, #e5e7eb);
  color: var(--xm-ink-2, #6b7280);
  font-size: 12px;
  white-space: nowrap;
}

.search-input {
  flex: 1;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 6px;
  border: 1px solid var(--xm-line, #e5e7eb);
  background: var(--xm-surface, #fff);
  box-shadow: none;
  transition: all 0.2s var(--xm-ease, ease);
}

.search-input :deep(.el-input__wrapper:hover) {
  border-color: var(--xm-line-strong, #d1d5db);
}

.search-input :deep(.el-input__wrapper.is-focus) {
  border-color: var(--xm-accent, #00c9a7);
  box-shadow: 0 0 0 2px var(--xm-accent-soft, rgba(0, 201, 167, 0.1));
}

.search-input :deep(.el-input__inner) {
  color: var(--xm-ink, #1f2937);
}

.search-input :deep(.el-input__inner::placeholder) {
  color: var(--xm-ink-3, #9ca3af);
}

.search-input :deep(.el-input__prefix-inner) {
  color: var(--xm-ink-3, #9ca3af);
}

.category-tabs-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px;
  border-radius: 8px;
  border: 1px solid var(--xm-line, #e5e7eb);
  background: var(--xm-surface, #fff);
}

.category-tabs {
  flex: 1;
  padding: 2px;
  overflow-x: auto;
  overflow-y: hidden;
  white-space: nowrap;
  scrollbar-width: thin;
  scrollbar-color: var(--xm-line-strong, #d1d5db) transparent;
}

.category-tabs::-webkit-scrollbar {
  height: 4px;
}

.category-tabs::-webkit-scrollbar-track {
  background: transparent;
}

.category-tabs::-webkit-scrollbar-thumb {
  background: var(--xm-line-strong, #d1d5db);
  border-radius: 4px;
}

.category-tabs::-webkit-scrollbar-thumb:hover {
  background: var(--xm-ink-3, #9ca3af);
}

.category-tabs-inner {
  display: flex;
  gap: 6px;
}

.category-tab {
  flex-shrink: 0;
  padding: 6px 14px;
  background: var(--xm-surface-2, #f9fafb);
  border: 1px solid var(--xm-line, #e5e7eb);
  border-radius: 6px;
  color: var(--xm-ink-2, #4b5563);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s var(--xm-ease, ease);
  white-space: nowrap;
}

.category-tab:hover {
  background: var(--xm-surface-2, #f3f4f6);
  border-color: var(--xm-line-strong, #d1d5db);
}

.category-tab.active {
  background: var(--xm-accent, #00c9a7);
  color: var(--xm-on-accent, #fff);
  border-color: var(--xm-accent, #00c9a7);
}

.category-scroll-btn {
  flex-shrink: 0;
  border: 1px solid var(--xm-line, #e5e7eb) !important;
  background: var(--xm-surface, #fff) !important;
  color: var(--xm-ink-2, #6b7280) !important;
}

.category-scroll-btn:hover:not(:disabled) {
  background: var(--xm-surface-2, #f9fafb) !important;
  border-color: var(--xm-line-strong, #d1d5db) !important;
}

.category-scroll-btn:disabled {
  opacity: 0.5;
}

.prompt-select-list {
  flex: 1;
  min-height: 0;
  max-height: calc(100vh - 280px);
  overflow-y: auto;
  padding: 2px;
}

.prompt-select-list::-webkit-scrollbar {
  width: 4px;
}

.prompt-select-list::-webkit-scrollbar-track {
  background: transparent;
}

.prompt-select-list::-webkit-scrollbar-thumb {
  background: var(--xm-line-strong, #d1d5db);
  border-radius: 4px;
}

.prompt-select-list::-webkit-scrollbar-thumb:hover {
  background: var(--xm-ink-3, #9ca3af);
}

.prompt-select-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 8px;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid var(--xm-line, #e5e7eb);
  background: var(--xm-surface, #fff);
  cursor: pointer;
  transition: all 0.2s var(--xm-ease, ease);
}

.prompt-select-item:hover {
  border-color: var(--xm-line-strong, #d1d5db);
  background: var(--xm-surface-2, #f9fafb);
}

.prompt-select-item.selected {
  border-color: var(--xm-accent, #00c9a7);
  background: var(--xm-accent-soft, #f0fdfa);
}

.prompt-select-item:last-child {
  margin-bottom: 0;
}

.prompt-select-info {
  flex: 1;
  min-width: 0;
}

.prompt-select-name {
  margin-bottom: 4px;
  color: var(--xm-ink, #1f2937);
  font-size: 14px;
  font-weight: 500;
  line-height: 1.4;
}

.prompt-select-desc {
  color: var(--xm-ink-2, #6b7280);
  font-size: 13px;
  line-height: 1.5;
}

.prompt-category-tag {
  flex-shrink: 0;
  border-radius: 4px;
  border: 1px solid var(--xm-line, #e5e7eb) !important;
  background: var(--xm-surface-2, #f9fafb) !important;
  color: var(--xm-ink-2, #6b7280) !important;
}

.empty-prompts {
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 10px;
  border: 1px dashed var(--xm-line-strong, #d1d5db);
  border-radius: 8px;
  background: var(--xm-surface-2, #f9fafb);
  color: var(--xm-ink-3, #9ca3af);
}

.empty-prompts .el-icon {
  color: var(--xm-line-strong, #d1d5db);
}

.empty-prompts p {
  margin: 0;
  font-size: 14px;
}

@media (max-width: 768px) {
  .prompt-select-dialog :deep(.el-dialog) {
    width: calc(100vw - 24px) !important;
    max-height: calc(100vh - 24px);
    border-radius: var(--xm-radius, 16px);
  }

  .prompt-select-dialog :deep(.el-dialog__header) {
    padding: 18px 18px 12px;
  }

  .prompt-select-dialog :deep(.el-dialog__body) {
    padding: 14px 18px 18px;
  }

  .prompt-select-dialog :deep(.el-dialog__footer) {
    padding: 12px 18px 18px;
  }

  .prompt-select-search-row {
    flex-direction: column;
    align-items: stretch;
  }

  .prompt-select-count {
    align-self: flex-start;
  }

  .prompt-select-item {
    padding: 12px 14px;
  }

  .prompt-category-tag {
    display: none;
  }
}

/* 暗色适配全部交给 write-theme.css 令牌翻转：
   面底 #1f2937/#111827 → --xm-surface/--xm-surface-2，边框 #374151/#4b5563 → --xm-line/--xm-line-strong，
   文字 #e6edf3/#9ca3af/#c8d6cd/#4d5768 → --xm-ink/--xm-ink-2/--xm-ink-3，
   选中 #3b82f6/rgba(59,130,246,0.1) → --xm-accent/--xm-accent-soft，
   原手工暗色补丁与令牌翻转等价，已整体删除 */

</style>
