<template>
  <div class="wiki-graph-page">
    <div class="page-header">
      <div class="header-left">
        <h1>wiki 图谱</h1>
        <p class="subtitle">管理多份图谱版本，从章节或导入文件构建你的知识世界</p>
      </div>
      <div class="header-right">
        <el-select
          v-model="selectedBookId"
          placeholder="选择书籍"
          size="default"
          class="book-selector"
          :loading="booksLoading"
        >
          <el-option
            v-for="book in books"
            :key="book.id"
            :label="book.title"
            :value="book.id"
          >
            <span class="book-option">
              <el-icon><Notebook /></el-icon>
              <span>{{ book.title }}</span>
            </span>
          </el-option>
        </el-select>
      </div>
    </div>

    <div v-if="!selectedBookId" class="empty-tip">
      <el-icon class="empty-icon" :size="56"><Connection /></el-icon>
      <div class="empty-text">请先选择一本书开始构建图谱</div>
      <div class="empty-sub">从右上角的下拉框中选择你已有的书籍</div>
    </div>

    <div v-else class="wiki-layout">
      <!-- 侧栏（章节 / 版本 / 导入）与弹窗共用同一组件，避免两处交互不一致 -->
      <GraphSidePanel class="wiki-side-panel" initial-tab="versions" />

      <main class="wiki-main">
        <KnowledgeGraph v-if="selectedBookId" :book-id="selectedBookId" :with-side-panel="false" />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Connection, Notebook } from '@element-plus/icons-vue'
import { bookAPI } from '@/api'
import type { Book } from '@/types'
import KnowledgeGraph from '@/components/KnowledgeGraph.vue'
import GraphSidePanel from '@/components/graph/GraphSidePanel.vue'

const books = ref<Book[]>([])
const booksLoading = ref(false)
const selectedBookId = ref<number | null>(null)

onMounted(async () => {
  await loadBooks()
})

async function loadBooks() {
  booksLoading.value = true
  try {
    const res: any = await bookAPI.getAll()
    if (res && res.success && Array.isArray(res.data)) {
      books.value = res.data
      // 默认选中第一本
      if (!selectedBookId.value && res.data.length > 0) {
        selectedBookId.value = res.data[0].id
      }
    } else {
      books.value = []
    }
  } catch (e) {
    console.error('加载书籍失败:', e)
  } finally {
    booksLoading.value = false
  }
}
</script>

<style scoped>
.wiki-graph-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px 24px;
  box-sizing: border-box;
  gap: 16px;
  overflow: hidden;
}

.page-header {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 12px;
  flex-shrink: 0;
}

.header-left h1 {
  margin: 0;
  font-size: 22px;
  color: var(--page-title-color, #1e293b);
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-left .subtitle {
  margin: 4px 0 0;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.book-selector {
  width: 240px;
}

.book-option {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.book-option .el-icon {
  color: #3b82f6;
}

.empty-tip {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--el-text-color-secondary);
}

.empty-icon {
  color: rgba(59, 130, 246, 0.4);
  margin-bottom: 8px;
}

.empty-text {
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.empty-sub {
  font-size: 12px;
}

.wiki-layout {
  flex: 1;
  display: flex;
  gap: 12px;
  min-height: 0;
  overflow: hidden;
}

/* 侧栏在页面上用卡片外观，组件本身保持通用样式 */
.wiki-side-panel {
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(59, 130, 246, 0.15);
  border-right: 1px solid rgba(59, 130, 246, 0.15);
  border-radius: 10px;
  backdrop-filter: blur(8px);
}

/* ── 主区域 ─────────────────────────── */
.wiki-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(59, 130, 246, 0.15);
  border-radius: 10px;
  overflow: hidden;
  min-width: 0;
}

.wiki-main > :deep(.knowledge-graph-panel) {
  height: 100%;
}

/* ── 暗色主题适配 ─────────────────────────── */
:root[data-theme='dark'] .header-left h1 {
  color: #e6edf3;
}

:root[data-theme='dark'] .header-left .subtitle {
  color: #9ca3af;
}

:root[data-theme='dark'] .empty-text {
  color: #e6edf3;
}

:root[data-theme='dark'] .wiki-side-panel {
  background: rgba(22, 27, 34, 0.7);
  border-color: rgba(42, 50, 66, 0.45);
}

:root[data-theme='dark'] .wiki-main {
  background: rgba(22, 27, 34, 0.6);
  border-color: rgba(42, 50, 66, 0.45);
}
</style>
