<template>
  <el-dialog
    v-model="rankReferenceDialogVisible"
    title="排行榜参考图"
    width="880px"
    append-to-body
    class="rank-reference-dialog"
  >
    <div class="rank-reference-layout">
      <div class="rank-reference-groups">
        <button
          v-for="group in rankReferenceGroups"
          :key="group.key"
          type="button"
          class="rank-reference-group-btn"
          :class="{ active: activeRankGroupKey === group.key }"
          @click="selectRankGroup(group.key)"
        >
          {{ group.label }}
        </button>
      </div>
      <div class="rank-reference-content">
        <div v-if="activeRankGroup" class="rank-reference-categories">
          <button
            v-for="category in activeRankGroup.categories"
            :key="category.id"
            type="button"
            class="rank-reference-category-btn"
            :class="{ active: activeRankCategoryPath === category.path }"
            @click="selectRankCategory(category.path)"
          >
            {{ category.label }}
          </button>
        </div>
        <div v-if="rankReferenceLoading" class="rank-reference-state">正在加载排行榜封面...</div>
        <div v-else-if="rankReferenceBooks.length === 0" class="rank-reference-state">暂无可用封面</div>
        <div v-else class="rank-reference-grid">
          <button
            v-for="book in rankReferenceBooks"
            :key="`${activeRankCategoryPath}_${book.rank}_${book.coverUrl}`"
            type="button"
            class="rank-reference-card"
            @click="useRankReferenceBook(book)"
          >
            <img :src="book.coverUrl" :alt="book.title || '排行榜封面'" class="rank-reference-cover" />
            <div class="rank-reference-book-meta">
              <span class="rank-reference-rank">#{{ String(book.rank).padStart(2, '0') }}</span>
              <span class="rank-reference-title">{{ book.title || '未命名作品' }}</span>
              <span class="rank-reference-author">{{ book.author || '未知作者' }}</span>
            </div>
          </button>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
// 排行榜参考图弹窗（分组 → 分类 → 选图）
// 2026-09-21 第 3 期从 views/Books.vue 拆出；选中结果通过 select 事件交回父组件
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { aiAPI } from '@/api'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'select', book: { rank: number; title: string; author: string; coverUrl: string }): void
}>()

const rankReferenceDialogVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v)
})

const rankReferenceLoading = ref(false)

const rankReferenceGroups = ref<Array<{
  key: string
  label: string
  categories: Array<{
    id: string
    label: string
    path: string
    fullUrl: string
  }>
}>>([])

const activeRankGroupKey = ref('')

const activeRankCategoryPath = ref('')

const rankReferenceBooks = ref<Array<{
  rank: number
  title: string
  author: string
  coverUrl: string
  bookPath: string
}>>([])

const loadRankReferenceGroups = async () => {
  const res = await aiAPI.getRankReferenceCategories()
  if (!res.success || !res.data) {
    throw new Error(res.message || '加载排行榜分类失败')
  }
  rankReferenceGroups.value = res.data
  if (!activeRankGroupKey.value && res.data.length > 0) {
    activeRankGroupKey.value = res.data[0].key
  }
}

const loadRankReferenceBooks = async (path: string) => {
  rankReferenceLoading.value = true
  try {
    const res = await aiAPI.getRankReferenceBooks(path)
    if (!res.success || !res.data) {
      throw new Error(res.message || '加载排行榜封面失败')
    }
    rankReferenceBooks.value = res.data
  } finally {
    rankReferenceLoading.value = false
  }
}

const selectRankGroup = async (groupKey: string) => {
  activeRankGroupKey.value = groupKey
  const group = rankReferenceGroups.value.find(item => item.key === groupKey)
  if (!group || group.categories.length === 0) {
    activeRankCategoryPath.value = ''
    rankReferenceBooks.value = []
    return
  }
  await selectRankCategory(group.categories[0].path)
}

const selectRankCategory = async (path: string) => {
  activeRankCategoryPath.value = path
  await loadRankReferenceBooks(path)
}

const activeRankGroup = computed(() =>
  rankReferenceGroups.value.find(group => group.key === activeRankGroupKey.value) || null
)

// 模板调用名保持不变；真正的落库动作由父组件接 select 事件处理
const useRankReferenceBook = (book: { rank: number; title: string; author: string; coverUrl: string }) => {
  emit('select', book)
}

// 打开时按需加载（原 openRankReferenceDialog 的加载逻辑，去掉了设 visible 那一步）
async function onOpen() {
  if (rankReferenceGroups.value.length === 0) {
    try {
      rankReferenceLoading.value = true
      await loadRankReferenceGroups()
      if (activeRankGroup.value?.categories?.[0]?.path) {
        await selectRankCategory(activeRankGroup.value.categories[0].path)
      }
    } catch (error: any) {
      ElMessage.error(error.message || '加载排行榜参考图失败')
    } finally {
      rankReferenceLoading.value = false
    }
    return
  }

  if (!activeRankCategoryPath.value && activeRankGroup.value?.categories?.[0]?.path) {
    try {
      await selectRankCategory(activeRankGroup.value.categories[0].path)
    } catch (error: any) {
      ElMessage.error(error.message || '加载排行榜参考图失败')
    }
  }
}
watch(() => props.visible, (v) => {
  if (v) onOpen()
})
</script>

<style scoped>
.rank-reference-layout {
  display: grid;
  grid-template-columns: 116px 1fr;
  gap: 10px;
  height: 100%;
  min-height: 0;
}

.rank-reference-groups {
  display: flex;
  flex-direction: column;
  gap: 6px;
  border-right: 1px solid #ebeef5;
  padding-right: 8px;
  min-height: 0;
  overflow-y: auto;
}

.rank-reference-group-btn,
.rank-reference-category-btn,
.rank-reference-card {
  border: none;
  background: transparent;
  cursor: pointer;
}

.rank-reference-group-btn {
  text-align: left;
  padding: 8px 10px;
  border-radius: 8px;
  color: #606266;
  font-size: 12px;
}

.rank-reference-group-btn.active {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 600;
}

.rank-reference-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.rank-reference-categories {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  flex-shrink: 0;
  max-height: 72px;
  overflow-y: auto;
  align-content: flex-start;
}

.rank-reference-category-btn {
  padding: 4px 10px;
  border-radius: 999px;
  background: #f5f7fa;
  color: #606266;
  font-size: 12px;
}

.rank-reference-category-btn.active {
  background: var(--el-color-primary);
  color: #fff;
}

.rank-reference-state {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.rank-reference-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 10px;
  align-content: start;
}

.rank-reference-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  text-align: left;
  transition: border-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
  min-height: 214px;
  height: 214px;
  overflow: hidden;
}

.rank-reference-card:hover {
  border-color: var(--el-color-primary);
  transform: translateY(-2px);
  box-shadow: 0 8px 20px color-mix(in srgb, var(--el-color-primary) 12%, transparent);
}

.rank-reference-cover {
  width: 100%;
  height: 124px;
  object-fit: cover;
  border-radius: 8px;
  background: #f5f7fa;
  flex-shrink: 0;
}

.rank-reference-book-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
  overflow: hidden;
}

.rank-reference-rank {
  font-size: 11px;
  color: #f56c6c;
  font-weight: 600;
}

.rank-reference-title {
  font-size: 12px;
  color: #303133;
  line-height: 1.45;
  min-height: 34px;
  max-height: 34px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rank-reference-author {
  font-size: 11px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-reference-dialog :deep(.el-dialog) {
  width: min(880px, calc(100vw - 32px)) !important;
  max-width: calc(100vw - 32px);
  height: min(360px, calc(100vh - 32px));
  max-height: calc(100vh - 32px);
  margin: 0 auto !important;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.rank-reference-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
}

.rank-reference-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  padding: 12px;
  overflow: hidden;
}

.rank-reference-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 16px;
  box-sizing: border-box;
}
</style>
