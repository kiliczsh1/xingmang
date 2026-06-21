<template>
  <div class="entry-library-panel">
    <div class="el-layout">
      <div class="el-left">
        <div class="el-left-top">
          <div class="el-search-bar">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索词条..."
              clearable
              size="default"
              @clear="handleSearch"
              @keyup.enter="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </div>
          <div class="el-categories">
            <div
              class="el-category-item"
              :class="{ active: activeCategories.length === 0 }"
              @click="handleCategoryToggle('')"
            >
              <el-icon class="el-category-icon"><Document /></el-icon>
              <span class="el-category-name">全部</span>
              <span class="el-category-count">{{ totalCount }}</span>
            </div>
            <div
              v-for="cat in categories"
              :key="cat.id"
              class="el-category-item"
              :class="{ active: activeCategories.includes(cat.type) }"
              @click="handleCategoryToggle(cat.type)"
            >
              <el-icon class="el-category-icon" :style="{ color: cat.color }">
                <component :is="getCategoryIcon(cat.type)" />
              </el-icon>
              <span class="el-category-name">{{ cat.name }}</span>
              <span class="el-category-count">{{ cat.entry_count }}</span>
            </div>
          </div>
        </div>
        <div class="el-left-bottom">
          <div class="el-entries-header">
            <span class="el-entries-title">
              {{ activeCategories.length > 0 ? activeCategories.map(c => getCategoryName(c)).join('、') : '全部词条' }}
              <span class="el-entries-total">({{ filteredEntries.length }})</span>
            </span>
            <div class="el-entries-header-right">
              <el-button
                v-if="selectedEntryIds.size > 0"
                size="small"
                type="primary"
                @click="handleBatchAddToBook"
              >
                <el-icon><Plus /></el-icon>
                批量添加到小说 ({{ selectedEntryIds.size }})
              </el-button>
              <el-select
                v-model="sortBy"
                size="small"
                class="sort-select"
                @change="handleSortChange"
              >
                <el-option label="默认排序" value="default" />
                <el-option label="按时间" value="time" />
                <el-option label="按热度" value="hot" />
              </el-select>
            </div>
          </div>
          <div class="el-entries-grid" v-loading="loading && !initialLoading">
            <template v-if="initialLoading">
              <div v-for="i in skeletonCount" :key="'sk-'+i" class="el-entry-card el-skeleton-card">
                <div class="el-card-checkbox"><span class="el-unchecked el-skeleton-box"></span></div>
                <div class="el-skeleton-avatar"></div>
                <div class="el-card-body">
                  <div class="el-skeleton-line el-skeleton-name"></div>
                  <div class="el-skeleton-line el-skeleton-desc"></div>
                  <div class="el-skeleton-line el-skeleton-tag"></div>
                </div>
              </div>
            </template>
            <template v-else-if="filteredEntries.length === 0 && !loading">
              <div class="el-empty">
                <el-icon><Document /></el-icon>
                <p>暂无词条</p>
                <p class="el-empty-hint">去提示词库创建第一个词条吧</p>
              </div>
            </template>
            <template v-else>
              <div
                v-for="entry in filteredEntries"
                :key="entry.id"
                class="el-entry-card"
                :class="{ active: selectedEntry?.id === entry.id, selected: selectedEntryIds.has(entry.id) }"
                @click="handleCardClick(entry, $event)"
              >
                <div class="el-card-checkbox" @click.stop="handleToggleSelect(entry)">
                  <el-icon v-if="selectedEntryIds.has(entry.id)" class="el-checked"><Check /></el-icon>
                  <span v-else class="el-unchecked"></span>
                </div>
                <div class="el-card-avatar">
                  <el-avatar :size="32" shape="square" :src="entry.avatar">
                    <el-icon :size="16"><component :is="getCategoryIcon(entry.category_type)" /></el-icon>
                  </el-avatar>
                </div>
                <div class="el-card-body">
                  <div class="el-card-name" v-html="highlightKeyword(entry.name)"></div>
                  <div class="el-card-desc" v-html="highlightKeyword(entry.description)"></div>
                  <div class="el-card-tags" v-if="entry.tags && entry.tags.length > 0">
                    <span
                      v-for="(tag, idx) in entry.tags.slice(0, 3)"
                      :key="tag"
                      class="el-card-tag"
                      :class="'el-tag-color-' + (idx % 6)"
                      @click.stop="handleTagClick(tag)"
                      :title="'按标签「' + tag + '」筛选'"
                    >{{ tag }}</span>
                    <span v-if="entry.tags.length > 3" class="el-card-tag el-card-tag-more">+{{ entry.tags.length - 3 }}</span>
                  </div>
                </div>
              </div>
            </template>
          </div>
          <div class="el-pagination" v-if="total > pageSize">
            <el-pagination
              v-model:current-page="currentPage"
              :page-size="pageSize"
              :total="total"
              layout="prev, pager, next"
              small
              @current-change="handlePageChange"
            />
          </div>
        </div>
      </div>

      <div class="el-right">
        <template v-if="selectedEntry">
          <div class="el-detail-header">
            <div class="el-detail-title-row">
              <h3 class="el-detail-title">{{ selectedEntry.name }}</h3>
              <el-tag
                size="small"
                :color="getCategoryColor(selectedEntry.category_type)"
                effect="dark"
              >
                {{ selectedEntry.category_name || getCategoryName(selectedEntry.category_type) }}
              </el-tag>
              <span class="el-detail-use-count" v-if="selectedEntry.use_count > 0">
                使用 {{ selectedEntry.use_count }} 次
              </span>
            </div>
            <el-button size="small" text @click="selectedEntry = null">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>

          <div class="el-detail-tabs">
            <div
              class="el-detail-tab"
              :class="{ active: detailTab === 'detail' }"
              @click="detailTab = 'detail'"
            >详情</div>
            <div
              class="el-detail-tab"
              :class="{ active: detailTab === 'relation' }"
              @click="detailTab = 'relation'"
            >关系</div>
          </div>

          <div class="el-detail-content" v-show="detailTab === 'detail'" ref="detailContentRef">
            <div class="el-detail-section" v-if="selectedEntry.appearance">
              <div class="el-section-label">◆ 外貌特征</div>
              <MarkdownRenderer :content="selectedEntry.appearance" />
            </div>
            <div class="el-detail-section" v-if="selectedEntry.background">
              <div class="el-section-label">◆ 背景故事</div>
              <MarkdownRenderer :content="selectedEntry.background" />
            </div>
            <div class="el-detail-section" v-if="selectedEntry.personality">
              <div class="el-section-label">◆ 性格描述</div>
              <MarkdownRenderer :content="selectedEntry.personality" />
            </div>
            <div class="el-detail-section" v-if="selectedEntry.description">
              <div class="el-section-label">◆ 概述</div>
              <MarkdownRenderer :content="selectedEntry.description" />
            </div>
            <div
              v-if="selectedEntry.custom_fields && Object.keys(selectedEntry.custom_fields).length > 0"
              class="el-detail-section"
            >
              <div class="el-section-label">◆ 补充设定</div>
              <div
                v-for="(val, key) in selectedEntry.custom_fields"
                :key="key"
                class="el-custom-field"
              >
                <span class="el-custom-field-key">{{ key }}：</span>
                <span class="el-custom-field-val">{{ val }}</span>
              </div>
            </div>
            <div v-if="!hasDetailContent" class="el-detail-empty">
              <el-empty description="暂无详细信息" :image-size="80" />
            </div>
          </div>

          <div class="el-detail-content" v-show="detailTab === 'relation'">
            <div v-if="selectedEntry.relationships" class="el-detail-section">
              <div class="el-section-label">◆ 关系网络</div>
              <MarkdownRenderer :content="selectedEntry.relationships" />
            </div>
            <div v-else class="el-detail-empty">
              <el-empty description="暂未建立关系" :image-size="80" />
            </div>
          </div>

          <div class="el-detail-actions">
            <el-button
              type="primary"
              :loading="completing"
              @click="handleAiComplete"
              :disabled="!canAiComplete"
            >
              <el-icon><MagicStick /></el-icon>
              {{ completing ? 'AI 正在补全...' : 'AI 智能补全' }}
            </el-button>
            <el-button
              type="success"
              @click="handleAddToBook"
              :disabled="addedEntryIds.has(selectedEntry.id)"
            >
              <el-icon><Plus /></el-icon>
              {{ addedEntryIds.has(selectedEntry.id) ? '已添加到我的小说' : '+ 添加到我的小说' }}
            </el-button>
          </div>

          <div v-if="aiPreviewEntry" class="el-ai-preview">
            <div class="el-ai-preview-header">
              <el-icon><MagicStick /></el-icon>
              <span>AI 补全预览</span>
              <el-button size="small" text @click="aiPreviewEntry = null">
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
            <div class="el-ai-preview-content">
              <div v-if="aiPreviewEntry.appearance" class="el-preview-item">
                <span class="el-preview-label">外貌特征：</span>
                <span>{{ aiPreviewEntry.appearance }}</span>
              </div>
              <div v-if="aiPreviewEntry.background" class="el-preview-item">
                <span class="el-preview-label">背景故事：</span>
                <span>{{ aiPreviewEntry.background }}</span>
              </div>
              <div v-if="aiPreviewEntry.personality" class="el-preview-item">
                <span class="el-preview-label">性格描述：</span>
                <span>{{ aiPreviewEntry.personality }}</span>
              </div>
            </div>
            <div class="el-ai-preview-actions">
              <el-button size="small" type="primary" @click="handleApplyAiPreview">
                <el-icon><Check /></el-icon>
                应用补全结果
              </el-button>
              <el-button size="small" @click="aiPreviewEntry = null">放弃</el-button>
            </div>
          </div>
        </template>
        <div v-else class="el-detail-placeholder">
          <el-empty description="选择左侧词条查看详情" :image-size="100" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { entryAPI } from '@/api'
import type { Entry, EntryCategory } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import { Search, Document, User, Close, Plus, MagicStick, Check, MapLocation, Box, OfficeBuilding, Clock, Location, Setting } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const props = defineProps<{
  bookId: number
  apiConfigs: ApiConfig[]
}>()

import type { ApiConfig } from '@/types'

const loading = ref(false)
const initialLoading = ref(true)
const completing = ref(false)
const searchKeyword = ref('')
const activeCategories = ref<string[]>([])
const sortBy = ref('default')
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)
const totalCount = ref(0)
const categories = ref<EntryCategory[]>([])
const entries = ref<Entry[]>([])
const selectedEntry = ref<Entry | null>(null)
const detailTab = ref<'detail' | 'relation'>('detail')
const addedEntryIds = ref(new Set<number>())
const selectedEntryIds = ref(new Set<number>())
const aiPreviewEntry = ref<Entry | null>(null)
const searchInputRef = ref<InstanceType<typeof import('element-plus')['ElInput']>>()
const detailContentRef = ref<HTMLElement>()

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const categoryTypeMap: Record<string, string> = {
  character: '角色',
  world_setting: '世界观设定',
  location: '地点',
  item: '物品',
  faction: '势力',
  event: '事件',
  skill: '功法',
  clue: '线索',
  other: '其他'
}

const categoryColorMap: Record<string, string> = {
  character: '#5470c6',
  world_setting: '#91cc75',
  location: '#fac858',
  item: '#ee6666',
  faction: '#73c0de',
  event: '#3ba272',
  skill: '#fc8452',
  clue: '#9a60b4',
  other: '#8b8b8b'
}

const categoryIconMap: Record<string, any> = {
  character: User,
  world_setting: Setting,
  location: MapLocation,
  item: Box,
  faction: OfficeBuilding,
  event: Clock,
  skill: MagicStick,
  clue: Search,
  other: Document
}

const skeletonCount = computed(() => pageSize.value)

const filteredEntries = computed(() => {
  let result = entries.value
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    result = result.filter(e =>
      e.name.toLowerCase().includes(kw) ||
      e.description.toLowerCase().includes(kw) ||
      e.tags.some(t => t.toLowerCase().includes(kw))
    )
  }
  if (sortBy.value === 'time') {
    result = [...result].sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
  } else if (sortBy.value === 'hot') {
    result = [...result].sort((a, b) => b.use_count - a.use_count)
  }
  return result
})

const hasDetailContent = computed(() => {
  if (!selectedEntry.value) return false
  const e = selectedEntry.value
  return !!(e.appearance || e.background || e.personality || e.description || (e.custom_fields && Object.keys(e.custom_fields).length > 0))
})

const canAiComplete = computed(() => {
  return props.apiConfigs.length > 0 && !completing.value
})

const allTags = computed(() => {
  const tagSet = new Set<string>()
  entries.value.forEach(e => e.tags.forEach(t => tagSet.add(t)))
  return [...tagSet].sort()
})

function highlightKeyword(text: string): string {
  if (!searchKeyword.value || !text) return text
  const kw = searchKeyword.value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const re = new RegExp(`(${kw})`, 'gi')
  return text.replace(re, '<mark class="el-highlight">$1</mark>')
}

function getCategoryName(type: string): string {
  return categoryTypeMap[type] || type
}

function getCategoryColor(type: string): string {
  return categoryColorMap[type] || '#8b8b8b'
}

function getCategoryIcon(type: string) {
  return categoryIconMap[type] || Document
}

function handleCategoryToggle(type: string) {
  if (type === '') {
    activeCategories.value = []
  } else {
    const idx = activeCategories.value.indexOf(type)
    if (idx >= 0) {
      activeCategories.value.splice(idx, 1)
    } else {
      activeCategories.value.push(type)
    }
  }
  currentPage.value = 1
  fetchEntries()
}

function handleSearch() {
  currentPage.value = 1
  fetchEntries()
}

function handleTagClick(tag: string) {
  searchKeyword.value = tag
  currentPage.value = 1
  fetchEntries()
}

function handlePageChange(page: number) {
  currentPage.value = page
  fetchEntries()
}

function handleSortChange() {
  currentPage.value = 1
}

async function fetchEntries() {
  loading.value = true
  try {
    const res = await entryAPI.getList({
      book_id: props.bookId,
      category_type: activeCategories.value.length > 0 ? activeCategories.value.join(',') : undefined,
      keyword: searchKeyword.value || undefined,
      page: currentPage.value,
      page_size: pageSize.value
    })
    if (res.success && res.data) {
      entries.value = res.data.entries
      total.value = res.data.total
      if (res.data.categories) {
        categories.value = res.data.categories
        totalCount.value = categories.value.reduce((sum, c) => sum + c.entry_count, 0)
      }
    }
  } catch (error) {
    console.error('Fetch entries failed:', error)
  } finally {
    loading.value = false
    initialLoading.value = false
  }
}

function handleCardClick(entry: Entry, event: MouseEvent) {
  const hasSelection = selectedEntryIds.value.size > 0
  if (hasSelection) {
    handleToggleSelect(entry)
    return
  }
  handleSelectEntry(entry)
}

function handleSelectEntry(entry: Entry) {
  selectedEntry.value = entry
  detailTab.value = 'detail'
  nextTick(() => {
    if (detailContentRef.value) {
      detailContentRef.value.scrollTop = 0
    }
  })
}

function handleToggleSelect(entry: Entry) {
  const set = selectedEntryIds.value
  if (set.has(entry.id)) {
    set.delete(entry.id)
  } else {
    set.add(entry.id)
  }
}

function handleClearSelection() {
  selectedEntryIds.value.clear()
}

function handleKeyboard(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
    e.preventDefault()
    const input = document.querySelector('.el-search-bar input') as HTMLInputElement
    if (input) input.focus()
    return
  }
  if (e.key === 'Escape') {
    if (aiPreviewEntry.value) {
      aiPreviewEntry.value = null
      return
    }
    if (selectedEntry.value) {
      selectedEntry.value = null
      return
    }
    if (selectedEntryIds.value.size > 0) {
      selectedEntryIds.value.clear()
      return
    }
  }
}

async function handleAiComplete() {
  if (!selectedEntry.value || completing.value) return
  const configId = props.apiConfigs.length > 0
    ? (props.apiConfigs.find(c => c.is_default)?.id || props.apiConfigs[0].id)
    : undefined
  if (!configId) {
    ElMessage.warning('请先配置 AI 模型')
    return
  }
  aiPreviewEntry.value = null
  completing.value = true
  try {
    const res = await entryAPI.aiAutoComplete(selectedEntry.value.id, configId)
    if (res.success && res.data) {
      aiPreviewEntry.value = res.data
      ElMessage.success('AI 补全完成，请预览结果')
    }
  } catch (error) {
    console.error('AI complete failed:', error)
    ElMessage.error('AI 补全失败')
  } finally {
    completing.value = false
  }
}

function handleApplyAiPreview() {
  if (!aiPreviewEntry.value || !selectedEntry.value) return
  selectedEntry.value = aiPreviewEntry.value
  aiPreviewEntry.value = null
  ElMessage.success('已应用 AI 补全结果')
}

async function handleAddToBook() {
  if (!selectedEntry.value || addedEntryIds.value.has(selectedEntry.value.id)) return
  await addEntriesToBook([selectedEntry.value.id])
}

async function handleBatchAddToBook() {
  if (selectedEntryIds.value.size === 0) return
  const idsToAdd = [...selectedEntryIds.value].filter(id => !addedEntryIds.value.has(id))
  if (idsToAdd.length === 0) {
    ElMessage.info('所选词条均已添加')
    return
  }
  try {
    await ElMessageBox.confirm(
      `确定将 ${idsToAdd.length} 个词条添加到当前小说吗？`,
      '批量添加确认',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'info' }
    )
    await addEntriesToBook(idsToAdd)
    selectedEntryIds.value.clear()
  } catch (error) {
    // 取消操作
  }
}

async function addEntriesToBook(entryIds: number[]) {
  const results: { id: number; ok: boolean }[] = []
  for (const id of entryIds) {
    try {
      const res = await entryAPI.addToBook(id, props.bookId)
      if (res.success) {
        addedEntryIds.value.add(id)
        results.push({ id, ok: true })
      } else {
        results.push({ id, ok: false })
      }
    } catch {
      results.push({ id, ok: false })
    }
  }
  const okCount = results.filter(r => r.ok).length
  const failCount = results.length - okCount
  if (results.length === 1) {
    if (okCount === 1) {
      ElMessage.success('已添加到我的小说')
    } else {
      ElMessage.error('添加失败')
    }
  } else {
    let msg = `成功添加 ${okCount} 个词条`
    if (failCount > 0) msg += `，${failCount} 个失败`
    ElMessage({ message: msg, type: okCount > 0 ? 'success' : 'error' })
  }
}

watch(searchKeyword, () => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    currentPage.value = 1
    fetchEntries()
  }, 300)
})

watch(() => props.bookId, () => {
  currentPage.value = 1
  activeCategories.value = []
  sortBy.value = 'default'
  searchKeyword.value = ''
  selectedEntry.value = null
  aiPreviewEntry.value = null
  addedEntryIds.value.clear()
  selectedEntryIds.value.clear()
  initialLoading.value = true
  fetchEntries()
})

onMounted(() => {
  fetchEntries()
  window.addEventListener('keydown', handleKeyboard)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyboard)
  if (debounceTimer) clearTimeout(debounceTimer)
})

defineExpose({
  fetchEntries
})
</script>

<style scoped>
.entry-library-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.el-layout {
  display: flex;
  height: 100%;
  flex: 1;
  overflow: hidden;
}

.el-left {
  width: 60%;
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--el-border-color-lighter);
  overflow: hidden;
}

.el-left-top {
  flex-shrink: 0;
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.el-search-bar {
  margin-bottom: 12px;
}

.el-categories {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.el-category-item {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid rgba(0, 0, 0, 0.08);
  background: var(--el-bg-color);
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 12px;
  user-select: none;
}

.el-category-item:hover {
  border-color: rgba(0, 0, 0, 0.15);
  background: rgba(0, 0, 0, 0.02);
}

.el-category-item.active {
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: #fff;
}

.el-category-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.el-category-item.active .el-category-icon {
  color: #fff !important;
}

.el-category-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.el-category-name {
  white-space: nowrap;
}

.el-category-count {
  font-size: 10px;
  opacity: 0.7;
}

.el-left-bottom {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.el-entries-header {
  padding: 10px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.el-entries-header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.el-entries-title {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--el-text-color-secondary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.el-entries-total {
  font-weight: 400;
  color: var(--el-text-color-placeholder);
  text-transform: none;
  letter-spacing: normal;
}

.el-empty-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-top: 4px;
}

.el-entries-grid {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  align-content: start;
}

.el-empty {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 16px;
  color: var(--el-text-color-secondary);
}

.el-empty .el-icon {
  font-size: 48px;
  margin-bottom: 8px;
  opacity: 0.5;
}

.el-entry-card {
  display: flex;
  gap: 8px;
  padding: 8px;
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.06);
  background: var(--el-bg-color);
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.el-entry-card:hover {
  border-color: rgba(0, 0, 0, 0.12);
  background: rgba(0, 0, 0, 0.02);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.el-entry-card.active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.el-entry-card.selected {
  border-color: var(--el-color-primary-light-3);
  background: var(--el-color-primary-light-9);
}

.el-card-checkbox {
  flex-shrink: 0;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.el-unchecked {
  width: 16px;
  height: 16px;
  border-radius: 3px;
  border: 2px solid var(--el-border-color);
  display: inline-block;
  transition: all 0.2s ease;
}

.el-entry-card:hover .el-unchecked {
  border-color: var(--el-color-primary-light-3);
}

.el-checked {
  color: var(--el-color-primary);
  font-size: 18px;
}

.el-skeleton-card {
  cursor: default;
  pointer-events: none;
}

.el-skeleton-avatar {
  width: 32px;
  height: 32px;
  border-radius: 4px;
  background: linear-gradient(90deg, var(--el-fill-color-light) 25%, var(--el-fill-color) 50%, var(--el-fill-color-light) 75%);
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.5s ease-in-out infinite;
  flex-shrink: 0;
}

.el-skeleton-line {
  height: 10px;
  border-radius: 5px;
  background: linear-gradient(90deg, var(--el-fill-color-light) 25%, var(--el-fill-color) 50%, var(--el-fill-color-light) 75%);
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.5s ease-in-out infinite;
}

.el-skeleton-name {
  width: 60%;
  height: 14px;
  margin-bottom: 6px;
}

.el-skeleton-desc {
  width: 90%;
  margin-bottom: 4px;
}

.el-skeleton-tag {
  width: 40%;
  height: 8px;
}

.el-skeleton-box {
  opacity: 0.3;
}

@keyframes skeleton-shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

:deep(.el-highlight) {
  background: #fde68a;
  color: #92400e;
  padding: 0 2px;
  border-radius: 2px;
  font-weight: 600;
}

.el-entries-sort {
  flex-shrink: 0;
}

.sort-select {
  width: 110px;
}

.el-card-avatar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
}

.el-card-avatar :deep(.el-avatar) {
  width: 32px !important;
  height: 32px !important;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
}

.el-card-avatar :deep(.el-icon) {
  font-size: 16px !important;
  color: var(--el-text-color-placeholder);
}

.el-card-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 2px;
}

.el-card-creator {
  font-size: 10px;
  color: var(--el-text-color-placeholder);
}

.el-card-use-count {
  font-size: 10px;
  color: var(--el-color-primary);
}

.el-card-body {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.el-card-name {
  font-size: 15px;
  font-weight: 500;
  font-family: 'Noto Serif SC', 'Songti SC', serif;
  letter-spacing: 0.02em;
  color: rgba(0, 0, 0, 0.9);
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.el-card-desc {
  font-size: 11px;
  line-height: 1.4;
  color: var(--el-text-color-secondary);
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 6px;
  word-break: break-word;
  white-space: pre-wrap;
}

.el-card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.el-card-tag {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 80px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.el-card-tag:hover {
  opacity: 0.8;
}

.el-card-tag-more {
  background: rgba(0, 0, 0, 0.06);
  color: var(--el-text-color-placeholder);
  cursor: default;
}

.el-pagination {
  display: flex;
  justify-content: center;
  padding: 8px 0;
  flex-shrink: 0;
}

.el-right {
  width: 40%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--el-fill-color-lighter);
}

.el-detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.el-detail-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.el-detail-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.el-detail-use-count {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  white-space: nowrap;
}

.el-detail-tabs {
  display: flex;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.el-detail-tab {
  flex: 1;
  text-align: center;
  padding: 10px 0;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: all 0.2s ease;
  user-select: none;
}

.el-detail-tab:hover {
  color: var(--el-color-primary);
}

.el-detail-tab.active {
  color: var(--el-color-primary);
  border-bottom-color: var(--el-color-primary);
  font-weight: 600;
}

.el-detail-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
}

.el-detail-section {
  margin-bottom: 16px;
}

.el-section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 8px;
}

.el-section-text {
  font-size: 13px;
  color: var(--el-text-color-regular);
  line-height: 1.7;
  white-space: pre-wrap;
}

.el-custom-field {
  margin-bottom: 4px;
  font-size: 13px;
  line-height: 1.6;
}

.el-custom-field-key {
  color: var(--el-text-color-secondary);
  font-weight: 500;
}

.el-custom-field-val {
  color: var(--el-text-color-regular);
}

.el-detail-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.el-detail-actions {
  padding: 12px 16px;
  border-top: 1px solid var(--el-border-color-lighter);
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.el-detail-placeholder {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.el-ai-preview {
  border-top: 2px solid var(--el-color-primary-light-3);
  background: var(--el-color-primary-light-9);
  padding: 12px 16px;
  flex-shrink: 0;
}

.el-ai-preview-header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-color-primary);
  margin-bottom: 10px;
}

.el-ai-preview-content {
  max-height: 200px;
  overflow-y: auto;
  margin-bottom: 10px;
}

.el-preview-item {
  margin-bottom: 8px;
  font-size: 12px;
  line-height: 1.6;
}

.el-preview-label {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.el-ai-preview-actions {
  display: flex;
  gap: 8px;
}

:root[data-theme='dark'] .el-left {
  border-right-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .el-left-top {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .el-category-item {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  color: #d1d5db;
}

:root[data-theme='dark'] .el-category-item:hover {
  border-color: rgba(94, 234, 212, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .el-category-item.active {
  background: #00c9a7;
  border-color: #00c9a7;
  color: #fff;
}

:root[data-theme='dark'] .el-entries-header {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .el-entries-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .el-entry-card {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  color: #d1d5db;
}

:root[data-theme='dark'] .el-entry-card:hover {
  border-color: rgba(94, 234, 212, 0.4);
}

:root[data-theme='dark'] .el-entry-card.active {
  border-color: #00c9a7;
  background: rgba(0, 201, 167, 0.15);
}

:root[data-theme='dark'] .el-card-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .el-card-desc {
  color: #9ca3af;
}

:root[data-theme='dark'] .el-card-tag {
  background: rgba(0, 201, 167, 0.15);
  color: #5eead4;
}

:root[data-theme='dark'] .el-right {
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .el-detail-header {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .el-detail-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .el-detail-tabs {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .el-detail-tab {
  color: #9ca3af;
}

:root[data-theme='dark'] .el-detail-tab:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .el-detail-tab.active {
  color: #5eead4;
  border-bottom-color: #5eead4;
}

:root[data-theme='dark'] .el-section-label {
  color: #f3f4f6;
}

:root[data-theme='dark'] .el-section-text {
  color: #d1d5db;
}

:root[data-theme='dark'] .el-custom-field-key {
  color: #9ca3af;
}

:root[data-theme='dark'] .el-custom-field-val {
  color: #d1d5db;
}

:root[data-theme='dark'] .el-detail-actions {
  border-top-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .el-empty {
  color: #9ca3af;
}

:root[data-theme='dark'] .el-empty-hint {
  color: #6b7280;
}

:root[data-theme='dark'] .el-card-creator {
  color: #6b7280;
}

:root[data-theme='dark'] .el-detail-use-count {
  color: #6b7280;
}

:root[data-theme='dark'] .el-card-tag-more {
  background: rgba(51, 65, 85, 0.6);
  color: #6b7280;
}

:root[data-theme='dark'] .el-ai-preview {
  border-top-color: rgba(0, 201, 167, 0.3);
  background: rgba(0, 201, 167, 0.08);
}

:root[data-theme='dark'] .el-preview-label {
  color: #f3f4f6;
}

.el-tag-color-0 {
  background: rgba(239, 68, 68, 0.1);
  color: #dc2626;
}

.el-tag-color-1 {
  background: rgba(245, 158, 11, 0.1);
  color: #d97706;
}

.el-tag-color-2 {
  background: rgba(34, 197, 94, 0.1);
  color: #16a34a;
}

.el-tag-color-3 {
  background: rgba(59, 130, 246, 0.1);
  color: #2563eb;
}

.el-tag-color-4 {
  background: rgba(168, 85, 247, 0.1);
  color: #9333ea;
}

.el-tag-color-5 {
  background: rgba(236, 72, 153, 0.1);
  color: #db2777;
}

:root[data-theme='dark'] .el-tag-color-0 {
  background: rgba(239, 68, 68, 0.2);
  color: #fca5a5;
}

:root[data-theme='dark'] .el-tag-color-1 {
  background: rgba(245, 158, 11, 0.2);
  color: #fcd34d;
}

:root[data-theme='dark'] .el-tag-color-2 {
  background: rgba(34, 197, 94, 0.2);
  color: #86efac;
}

:root[data-theme='dark'] .el-tag-color-3 {
  background: rgba(59, 130, 246, 0.2);
  color: #93c5fd;
}

:root[data-theme='dark'] .el-tag-color-4 {
  background: rgba(168, 85, 247, 0.2);
  color: #d8b4fe;
}

:root[data-theme='dark'] .el-tag-color-5 {
  background: rgba(236, 72, 153, 0.2);
  color: #f9a8d4;
}

</style>

<style>
:root[data-theme='dark'] .el-highlight {
  background: #713f12;
  color: #fde68a;
}

:root[data-theme='dark'] .el-skeleton-avatar,
:root[data-theme='dark'] .el-skeleton-line {
  background: linear-gradient(90deg, rgba(71, 85, 105, 0.3) 25%, rgba(71, 85, 105, 0.5) 50%, rgba(71, 85, 105, 0.3) 75%);
  background-size: 200% 100%;
}

.el-entries-grid::-webkit-scrollbar,
.el-detail-content::-webkit-scrollbar {
  width: 6px;
}

.el-entries-grid::-webkit-scrollbar-track,
.el-detail-content::-webkit-scrollbar-track {
  background: transparent;
}

.el-entries-grid::-webkit-scrollbar-thumb,
.el-detail-content::-webkit-scrollbar-thumb {
  background: var(--el-border-color);
  border-radius: 3px;
}

.el-entries-grid::-webkit-scrollbar-thumb:hover,
.el-detail-content::-webkit-scrollbar-thumb:hover {
  background: var(--el-text-color-placeholder);
}
</style>
