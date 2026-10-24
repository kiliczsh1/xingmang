<template>
  <div class="wiki-graph-page">
    <div v-if="!selectedBookId" class="empty-tip">
      <el-icon class="empty-icon" :size="56"><Connection /></el-icon>
      <div class="empty-text">请先选择一本书开始构建图谱</div>
      <div class="empty-sub">从下面「选择书籍」下拉里挑一本已有的书</div>
      <div class="empty-actions">
        <el-select
          v-model="selectedBookId"
          placeholder="选择书籍"
          size="default"
          class="book-selector"
          popper-class="kg-select-popper"
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

    <div v-else class="wiki-layout">
      <main class="wiki-main">
        <KnowledgeGraph v-if="selectedBookId" :book-id="selectedBookId" :with-side-panel="false">
          <!-- 页头功能并进画布工具栏：不再单独占一行，画布往上撑满 -->
          <template #toolbar-extra>
            <el-select
              v-model="selectedBookId"
              placeholder="选择书籍"
              size="small"
              class="book-selector"
              popper-class="kg-select-popper"
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

            <!-- 版本：按钮 + 下拉列表（点开选，跟历史对话那种下拉一致） -->
            <el-popover
              v-model:visible="versionMenuOpen"
              placement="bottom-end"
              :width="330"
              trigger="click"
              popper-class="graph-version-pop"
            >
              <template #reference>
                <button type="button" class="tb-btn" :disabled="!selectedBookId" title="切换图谱版本">
                  <el-icon class="tb-btn-icon"><Share /></el-icon>
                  <span class="tb-btn-text">{{ currentVersionName }}</span>
                  <el-icon class="tb-caret"><ArrowDown /></el-icon>
                </button>
              </template>

              <div class="gv-head">
                <span class="gv-title">图谱版本<b>{{ versions.length }}</b></span>
                <div class="gv-head-ops">
                  <el-button size="small" :disabled="!selectedBookId" @click="createBlankVersion">新建</el-button>
                  <el-button size="small" type="primary" plain :disabled="!selectedBookId" @click="saveAsNewVersion">
                    存为新版本
                  </el-button>
                </div>
              </div>

              <div class="gv-list">
                <div
                  v-for="ver in versions"
                  :key="ver.id"
                  class="gv-item"
                  :class="{ active: ver.id === currentVersionId, switching: switchingId === ver.id }"
                  @click="pickVersion(ver.id)"
                >
                  <span class="gv-name" :title="ver.name">{{ ver.name }}</span>
                  <span class="gv-meta">
                    {{ ver.entity_count }} 实体 · {{ ver.relation_count }} 关系 · {{ formatRelative(ver.created_at) }}
                  </span>
                  <span class="gv-ops" @click.stop>
                    <el-icon title="重命名" @click="renameVersion(ver.id)"><Edit /></el-icon>
                    <el-icon title="删除" class="danger" @click="deleteVersion(ver.id)"><Delete /></el-icon>
                  </span>
                </div>
                <div v-if="!versions.length" class="gv-empty">
                  还没有图谱版本，点「新建」从空白开始
                </div>
              </div>
            </el-popover>
          </template>

          <!-- 最右侧：导入 / 导出 -->
          <template #toolbar-tail>
            <el-button size="small" :disabled="!selectedBookId" @click="importVisible = true">
              <el-icon><Upload /></el-icon>
              导入
            </el-button>
            <el-button size="small" :disabled="!currentVersionId" @click="exportCurrent">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
          </template>
        </KnowledgeGraph>
      </main>
    </div>

    <!-- 导入：原侧栏的导入面板，挪进弹窗 -->
    <el-dialog
      v-model="importVisible"
      title="导入图谱"
      width="520px"
      align-center
      append-to-body
      destroy-on-close
      class="xm-dialog"
    >
      <WikiFileImporter :book-id="selectedBookId" @import="onImportFile" />
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  ArrowDown,
  Connection,
  Delete,
  Download,
  Edit,
  Notebook,
  Share,
  Upload
} from '@element-plus/icons-vue'
import { bookAPI } from '@/api'
import type { Book, GraphEntity, GraphRelation } from '@/types'
import KnowledgeGraph from '@/components/KnowledgeGraph.vue'
import WikiFileImporter from '@/components/wiki/WikiFileImporter.vue'
import { useGraphVersionActions } from '@/composables/useGraphVersionActions'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'

const store = useKnowledgeGraphStore()
const {
  switchingId,
  switchVersion,
  renameVersion,
  deleteVersion,
  exportVersion,
  saveAsNewVersion,
  createBlankVersion,
  importFile
} = useGraphVersionActions()

const books = ref<Book[]>([])
const booksLoading = ref(false)
const selectedBookId = ref<number | null>(null)
const versionMenuOpen = ref(false)
const importVisible = ref(false)

const versions = computed(() => store.versions)
const currentVersionId = computed(() => store.currentVersionId)
const currentVersionName = computed(() => {
  const cur = store.versions.find(v => v.id === store.currentVersionId)
  if (cur) return cur.name
  return store.versions.length ? '选择版本' : '暂无版本'
})

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

async function pickVersion(id: number) {
  versionMenuOpen.value = false
  await switchVersion(id)
}

function exportCurrent() {
  if (currentVersionId.value) exportVersion(currentVersionId.value)
}

async function onImportFile(payload: {
  entities: GraphEntity[]
  relations: GraphRelation[]
  name: string
}) {
  importVisible.value = false
  await importFile(payload)
}

function formatRelative(iso: string): string {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  if (isNaN(t)) return iso
  const diff = Date.now() - t
  const min = 60 * 1000
  const hour = 60 * min
  const day = 24 * hour
  if (diff < min) return '刚刚'
  if (diff < hour) return `${Math.floor(diff / min)} 分钟前`
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`
  if (diff < 7 * day) return `${Math.floor(diff / day)} 天前`
  const d = new Date(t)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
</script>

<style scoped>
/* 页头已并入画布工具栏，这里只留画布本身，整块高度全给图谱 */
.wiki-graph-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 0;
  box-sizing: border-box;
  gap: 0;
  overflow: hidden;
}

.book-selector {
  width: 190px;
}

.book-option {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.book-option .el-icon {
  color: #3b82f6;
}

/* 版本切换按钮：等高于 el-button default */
.tb-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 24px;
  max-width: 190px;
  padding: 0 8px;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  background: var(--el-fill-color-blank);
  color: var(--el-text-color-regular);
  font-size: 12px;
  cursor: pointer;
  transition: border-color 0.16s ease, color 0.16s ease;
}
.tb-btn:hover:not(:disabled) {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}
.tb-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}
.tb-btn-icon {
  font-size: 14px;
  flex-shrink: 0;
  color: #3b82f6;
}
.tb-btn-text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.tb-caret {
  font-size: 12px;
  flex-shrink: 0;
  color: var(--el-text-color-secondary);
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
.empty-actions {
  margin-top: 10px;
}

.wiki-layout {
  flex: 1;
  display: flex;
  gap: 12px;
  min-height: 0;
  overflow: hidden;
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
:root[data-theme='dark'] .empty-text {
  color: #e6edf3;
}

:root[data-theme='dark'] .wiki-main {
  background: rgba(22, 27, 34, 0.6);
  border-color: rgba(42, 50, 66, 0.45);
}

:root[data-theme='dark'] .tb-btn {
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.6);
  color: #c9d1d9;
}
</style>

<style>
/* 版本下拉：teleport 到 body，样式必须放非 scoped 块 */
.graph-version-pop {
  padding: 10px 10px 8px;
}

.graph-version-pop .gv-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 0 2px 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.graph-version-pop .gv-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.graph-version-pop .gv-title b {
  margin-left: 6px;
  padding: 0 6px;
  border-radius: 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 11px;
  font-weight: 500;
}

.graph-version-pop .gv-head-ops {
  display: flex;
  align-items: center;
  gap: 6px;
}

.graph-version-pop .gv-list {
  max-height: 320px;
  overflow-y: auto;
  padding-top: 6px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.graph-version-pop .gv-item {
  position: relative;
  display: grid;
  grid-template-columns: 1fr auto;
  grid-template-areas:
    'name ops'
    'meta ops';
  align-items: center;
  gap: 2px 8px;
  padding: 7px 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.14s ease;
}

.graph-version-pop .gv-item:hover {
  background: var(--el-fill-color-light);
}

.graph-version-pop .gv-item.active {
  background: var(--el-color-primary-light-9);
}

.graph-version-pop .gv-item.switching {
  opacity: 0.6;
  pointer-events: none;
}

.graph-version-pop .gv-name {
  grid-area: name;
  font-size: 13px;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.graph-version-pop .gv-item.active .gv-name {
  color: var(--el-color-primary);
  font-weight: 500;
}

.graph-version-pop .gv-meta {
  grid-area: meta;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.graph-version-pop .gv-ops {
  grid-area: ops;
  display: flex;
  align-items: center;
  gap: 6px;
  opacity: 0;
  transition: opacity 0.14s ease;
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.graph-version-pop .gv-item:hover .gv-ops {
  opacity: 1;
}

.graph-version-pop .gv-ops .el-icon:hover {
  color: var(--el-color-primary);
}

.graph-version-pop .gv-ops .danger:hover {
  color: var(--el-color-danger);
}

.graph-version-pop .gv-empty {
  padding: 16px 8px;
  text-align: center;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

/* ============================================================
   暗色：工具栏按钮 / 书籍下拉弹层 / 版本弹层
   EP 内部元素没有本组件的 scope 属性，必须放这个非 scoped 块
   ============================================================ */

/* 工具栏里的普通按钮（显示 / 编辑 / 导入 / 导出）——插槽按钮也算，所以按容器选中 */
:root[data-theme='dark'] .graph-toolbar .el-button:not(.el-button--primary) {
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.9);
  color: #c9d3de;
}
:root[data-theme='dark'] .graph-toolbar .el-button:not(.el-button--primary):hover,
:root[data-theme='dark'] .graph-toolbar .el-button:not(.el-button--primary):focus {
  background: #1c2333;
  border-color: rgba(59, 130, 246, 0.35);
  color: #e6edf3;
}

/* 主操作（AI 分析，plain） */
:root[data-theme='dark'] .graph-toolbar .el-button--primary.is-plain {
  background: rgba(59, 130, 246, 0.14);
  border-color: rgba(59, 130, 246, 0.38);
  color: #79b8ff;
}
:root[data-theme='dark'] .graph-toolbar .el-button--primary.is-plain:hover {
  background: rgba(59, 130, 246, 0.24);
  border-color: rgba(59, 130, 246, 0.55);
  color: #a5d6ff;
}

/* 搜索框：底色已由全局接管，剩内阴影那圈亮边框要自己压掉 */
:root[data-theme='dark'] .graph-toolbar .el-input__wrapper {
  background: #1c2333;
  box-shadow: 0 0 0 1px rgba(42, 50, 66, 0.9) inset;
}
:root[data-theme='dark'] .graph-toolbar .el-input__wrapper:hover {
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.35) inset;
}
:root[data-theme='dark'] .graph-toolbar .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.55) inset;
}
:root[data-theme='dark'] .graph-toolbar .el-input__inner {
  color: #e6edf3;
}

/* 书籍下拉：popper 内层 .el-select-dropdown 自带亮底，全局只压了外层 */
:root[data-theme='dark'] .kg-select-popper {
  --el-bg-color-overlay: #161b22;
  --el-fill-color-light: #1c2333;
  --el-text-color-regular: #c9d3de;
  --el-border-color-light: rgba(42, 50, 66, 0.9);
}
:root[data-theme='dark'] .kg-select-popper .el-select-dropdown {
  background: #161b22;
  border: 1px solid rgba(42, 50, 66, 0.9);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
}
:root[data-theme='dark'] .kg-select-popper .el-select-dropdown__item {
  color: #c9d3de;
}
:root[data-theme='dark'] .kg-select-popper .el-select-dropdown__item.is-hovering,
:root[data-theme='dark'] .kg-select-popper .el-select-dropdown__item:hover {
  background: #1c2333;
  color: #e6edf3;
}
:root[data-theme='dark'] .kg-select-popper .el-select-dropdown__item.is-selected {
  color: #58a6ff;
  font-weight: 600;
}

/* 版本弹层 */
:root[data-theme='dark'] .graph-version-pop {
  --el-bg-color-overlay: #161b22;
  --el-fill-color-light: #1c2333;
  --el-text-color-primary: #e6edf3;
  --el-text-color-secondary: #8b98a8;
  --el-border-color-lighter: rgba(42, 50, 66, 0.9);
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.9);
  color: #c9d3de;
}
:root[data-theme='dark'] .graph-version-pop .el-popper__arrow::before {
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.9);
}
:root[data-theme='dark'] .graph-version-pop .gv-head {
  border-bottom-color: rgba(42, 50, 66, 0.9);
}
:root[data-theme='dark'] .graph-version-pop .gv-title {
  color: #e6edf3;
}
:root[data-theme='dark'] .graph-version-pop .gv-name {
  color: #e6edf3;
}
:root[data-theme='dark'] .graph-version-pop .gv-meta {
  color: #8b98a8;
}
:root[data-theme='dark'] .graph-version-pop .gv-item:hover {
  background: #1c2333;
}
:root[data-theme='dark'] .graph-version-pop .gv-item.active {
  background: rgba(59, 130, 246, 0.14);
}
:root[data-theme='dark'] .graph-version-pop .gv-ops {
  color: #8b98a8;
}
:root[data-theme='dark'] .graph-version-pop .el-button:not(.el-button--primary) {
  background: #1c2333;
  border-color: rgba(42, 50, 66, 0.9);
  color: #c9d3de;
}
:root[data-theme='dark'] .graph-version-pop .el-button--primary.is-plain {
  background: rgba(59, 130, 246, 0.14);
  border-color: rgba(59, 130, 246, 0.38);
  color: #79b8ff;
}
</style>
