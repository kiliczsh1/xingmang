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
          @change="onBookChange"
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
      <!-- 左侧多 Tab 侧边栏（参考项目风格 + 星芒主题） -->
      <aside class="wiki-sidebar" :class="{ collapsed: sidebarCollapsed }">
        <div class="sidebar-collapse-handle" @click="sidebarCollapsed = !sidebarCollapsed">
          <el-icon :class="{ 'rotate-180': sidebarCollapsed }"><ArrowLeft /></el-icon>
        </div>
        <div class="sidebar-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="sidebar-tab"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            <el-icon><component :is="tab.icon" /></el-icon>
            <span>{{ tab.label }}</span>
            <span v-if="tab.badge" class="tab-badge">{{ tab.badge }}</span>
          </button>
        </div>
        <div class="sidebar-content">
          <GraphVersionList
            v-show="activeTab === 'versions'"
            :versions="versions"
            :current-version-id="currentVersionId"
            :book-id="selectedBookId"
            :switching-id="switchingId"
            @switch-version="handleSwitchVersion"
            @rename-version="handleRenameVersion"
            @delete-version="handleDeleteVersion"
            @export-version="handleExportVersion"
            @create-version="handleSaveAsNewVersion"
          />
          <WikiFileImporter
            v-show="activeTab === 'import'"
            :book-id="selectedBookId"
            @import="handleImportFile"
          />
        </div>
      </aside>

      <!-- 主区域：知识图谱 -->
      <main class="wiki-main">
        <KnowledgeGraph
          v-if="selectedBookId"
          :key="selectedBookId"
          ref="knowledgeGraphRef"
          :book-id="selectedBookId"
          :api-configs="apiConfigs"
          :chapters="chapters"
        />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Notebook, Connection, Share, UploadFilled, ArrowLeft
} from '@element-plus/icons-vue'
import { bookAPI, chapterAPI, configAPI, knowledgeGraphAPI } from '@/api'
import type { Book, Chapter, ApiConfig, GraphVersion, GraphEntity, GraphRelation } from '@/types'
import KnowledgeGraph from '@/components/KnowledgeGraph.vue'
import GraphVersionList from '@/components/wiki/GraphVersionList.vue'
import WikiFileImporter from '@/components/wiki/WikiFileImporter.vue'

const books = ref<Book[]>([])
const booksLoading = ref(false)
const selectedBookId = ref<number | null>(null)
const apiConfigs = ref<ApiConfig[]>([])
const chapters = ref<Chapter[]>([])

const activeTab = ref<'versions' | 'import'>('versions')
const sidebarCollapsed = ref(false)

// 版本管理
const versions = ref<GraphVersion[]>([])
const currentVersionId = ref<number | null>(null)
const switchingId = ref<number | null>(null)

const knowledgeGraphRef = ref<InstanceType<typeof KnowledgeGraph> | null>(null)

const tabs = computed(() => [
  { key: 'versions' as const, label: '图谱列表', icon: Share, badge: versions.value.length || '' },
  { key: 'import' as const, label: '导入文件', icon: UploadFilled, badge: '' }
])

onMounted(async () => {
  await loadBooks()
  await loadApiConfigs()
})

watch(selectedBookId, async (val, old) => {
  if (val && val !== old) {
    await loadChapters(val)
    await loadVersions(val)
  } else if (!val) {
    chapters.value = []
    versions.value = []
    currentVersionId.value = null
  }
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

async function loadApiConfigs() {
  try {
    const res: any = await configAPI.getAll()
    if (res && res.success && Array.isArray(res.data)) {
      apiConfigs.value = res.data
    }
  } catch (e) {
    console.error('加载 API 配置失败:', e)
  }
}

async function loadChapters(bookId: number) {
  try {
    const res: any = await chapterAPI.getByBook(bookId)
    if (res && res.success && Array.isArray(res.data)) {
      chapters.value = res.data
    } else {
      chapters.value = []
    }
  } catch (e) {
    console.error('加载章节失败:', e)
    chapters.value = []
  }
}

async function loadVersions(bookId: number) {
  try {
    const res: any = await knowledgeGraphAPI.getVersions(bookId)
    if (res && res.success && Array.isArray(res.data)) {
      versions.value = res.data
      // 默认激活第一个版本
      if (!currentVersionId.value && versions.value.length > 0) {
        currentVersionId.value = versions.value[0].id
      }
    } else {
      versions.value = []
    }
  } catch (e) {
    console.error('加载图谱版本失败:', e)
    versions.value = []
  }
}

function onBookChange(_val: number) {
  currentVersionId.value = null
}

// ── 版本操作 ──────────────────────────────────────────

async function handleSwitchVersion(versionId: number) {
  if (!selectedBookId.value) return
  if (versionId === currentVersionId.value) return
  switchingId.value = versionId
  try {
    const ok = await knowledgeGraphRef.value?.loadVersionData(versionId)
    if (ok) {
      currentVersionId.value = versionId
      ElMessage.success(`已切换到：${versions.value.find(v => v.id === versionId)?.name || ''}`)
    } else {
      ElMessage.error('该版本数据读取失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '切换失败')
  } finally {
    switchingId.value = null
  }
}

async function handleRenameVersion(versionId: number) {
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  try {
    const { value: name } = await ElMessageBox.prompt('重命名图谱版本', '重命名', {
      inputPlaceholder: '输入新名称',
      inputValue: ver.name,
      confirmButtonText: '保存',
      cancelButtonText: '取消'
    })
    const trimmed = (name || '').trim()
    if (!trimmed || trimmed === ver.name) return
    const res: any = await knowledgeGraphAPI.renameVersion(versionId, trimmed)
    if (res && res.success) {
      ElMessage.success('已重命名')
      if (selectedBookId.value) await loadVersions(selectedBookId.value)
    } else {
      ElMessage.error('重命名失败')
    }
  } catch (e: any) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e?.message || '操作失败')
  }
}

async function handleDeleteVersion(versionId: number) {
  if (!selectedBookId.value) return
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  try {
    await ElMessageBox.confirm(
      `确认删除版本「${ver.name}」？该操作不可恢复。`,
      '删除版本',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    const res: any = await knowledgeGraphAPI.deleteVersion(selectedBookId.value, versionId)
    if (res && res.success) {
      ElMessage.success('已删除')
      if (currentVersionId.value === versionId) {
        currentVersionId.value = null
      }
      await loadVersions(selectedBookId.value)
    } else {
      ElMessage.error('删除失败')
    }
  } catch (e: any) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e?.message || '操作失败')
  }
}

async function handleExportVersion(versionId: number) {
  if (!selectedBookId.value) return
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  try {
    const res: any = await knowledgeGraphAPI.getVersionData(versionId)
    if (!res || !res.success || !res.data) {
      ElMessage.error('无法读取版本数据')
      return
    }
    const data = res.data
    const payload = {
      name: ver.name,
      description: ver.description,
      scope: ver.scope,
      bookId: selectedBookId.value,
      exportedAt: Date.now(),
      entities: data.entities || [],
      relations: data.relations || []
    }
    const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `知识图谱_${ver.name}_${Date.now()}.json`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('已导出')
  } catch (e: any) {
    ElMessage.error(e?.message || '导出失败')
  }
}

async function handleSaveAsNewVersion() {
  if (!selectedBookId.value) return
  if (!knowledgeGraphRef.value) {
    ElMessage.warning('请先打开图谱')
    return
  }
  // 触发 KnowledgeGraph 内部的保存逻辑
  await knowledgeGraphRef.value.saveAsNewVersion()
  if (selectedBookId.value) await loadVersions(selectedBookId.value)
}

async function handleImportFile(payload: { entities: GraphEntity[]; relations: GraphRelation[]; name: string }) {
  if (!selectedBookId.value) return
  try {
    // 1) 创建版本
    const createRes: any = await knowledgeGraphAPI.createVersion(selectedBookId.value, payload.name)
    if (!createRes || !createRes.success || !createRes.data) {
      ElMessage.error('创建版本失败')
      return
    }
    const newVer = createRes.data
    // 2) 写入图谱数据
    const importRes: any = await knowledgeGraphAPI.importGraph(selectedBookId.value, {
      entities: payload.entities,
      relations: payload.relations,
      versionId: newVer.id,
      name: newVer.name
    } as any)
    if (importRes && importRes.success) {
      ElMessage.success(`已导入为新版本：${newVer.name}`)
      currentVersionId.value = newVer.id
      await loadVersions(selectedBookId.value)
      // 触发主区刷新（通过重新加载图谱）
      if (knowledgeGraphRef.value) {
        await knowledgeGraphRef.value.fetchGraphData()
      }
    } else {
      ElMessage.error('写入数据失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '导入失败')
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
  color: #2dd4bf;
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
  color: rgba(45, 212, 191, 0.4);
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

/* ── 侧边栏 ─────────────────────────── */
.wiki-sidebar {
  width: 320px;
  min-width: 320px;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.7);
  border: 1px solid rgba(45, 212, 191, 0.15);
  border-radius: 10px;
  overflow: hidden;
  backdrop-filter: blur(8px);
  position: relative;
  transition: width 0.25s ease, min-width 0.25s ease;
}

.wiki-sidebar.collapsed {
  width: 48px !important;
  min-width: 48px !important;
}

.wiki-sidebar.collapsed .sidebar-tabs,
.wiki-sidebar.collapsed .sidebar-content {
  opacity: 0;
  pointer-events: none;
}

.sidebar-collapse-handle {
  position: absolute;
  top: 50%;
  right: -10px;
  transform: translateY(-50%);
  background: #fff;
  border: 1px solid rgba(45, 212, 191, 0.3);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  transition: all 0.2s ease;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

.sidebar-collapse-handle:hover {
  background: #2dd4bf;
  border-color: #2dd4bf;
  color: #fff;
}

.sidebar-collapse-handle .el-icon {
  font-size: 12px;
  transition: transform 0.25s ease;
}

.sidebar-collapse-handle .el-icon.rotate-180 {
  transform: rotate(180deg);
}

.sidebar-tabs {
  display: flex;
  border-bottom: 1px solid rgba(45, 212, 191, 0.15);
  background: rgba(45, 212, 191, 0.04);
  flex-shrink: 0;
  transition: opacity 0.2s ease;
}

.sidebar-tab {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 10px 8px;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  transition: all 0.18s ease;
  position: relative;
}

.sidebar-tab:hover {
  color: #2dd4bf;
  background: rgba(45, 212, 191, 0.06);
}

.sidebar-tab.active {
  color: #2dd4bf;
  border-bottom-color: #2dd4bf;
  background: rgba(45, 212, 191, 0.08);
  font-weight: 500;
}

.sidebar-tab .el-icon {
  font-size: 14px;
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 16px;
  padding: 0 5px;
  margin-left: 2px;
  background: rgba(45, 212, 191, 0.18);
  color: #2dd4bf;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 500;
  line-height: 1;
}

.sidebar-content {
  flex: 1;
  padding: 12px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
  transition: opacity 0.2s ease;
}

/* ── 主区域 ─────────────────────────── */
.wiki-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(45, 212, 191, 0.15);
  border-radius: 10px;
  overflow: hidden;
  min-width: 0;
}

.wiki-main > :deep(.knowledge-graph-panel) {
  height: 100%;
}

/* ── 暗色主题适配 ─────────────────────────── */
:root[data-theme='dark'] .header-left h1 {
  color: #f3f4f6;
}

:root[data-theme='dark'] .header-left .subtitle {
  color: #9ca3af;
}

:root[data-theme='dark'] .empty-text {
  color: #e5e7eb;
}

:root[data-theme='dark'] .wiki-sidebar {
  background: rgba(30, 41, 59, 0.7);
  border-color: rgba(71, 85, 105, 0.45);
}

:root[data-theme='dark'] .sidebar-collapse-handle {
  background: #1e293b;
  border-color: rgba(94, 234, 212, 0.35);
  color: #e5e7eb;
}

:root[data-theme='dark'] .sidebar-collapse-handle:hover {
  background: #2dd4bf;
  color: #0f172a;
}

:root[data-theme='dark'] .sidebar-tabs {
  background: rgba(15, 23, 42, 0.5);
  border-bottom-color: rgba(71, 85, 105, 0.45);
}

:root[data-theme='dark'] .sidebar-tab {
  color: #9ca3af;
}

:root[data-theme='dark'] .sidebar-tab:hover {
  color: #5eead4;
  background: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .sidebar-tab.active {
  color: #5eead4;
  border-bottom-color: #2dd4bf;
  background: rgba(45, 212, 191, 0.12);
}

:root[data-theme='dark'] .tab-badge {
  background: rgba(45, 212, 191, 0.22);
  color: #5eead4;
}

:root[data-theme='dark'] .wiki-main {
  background: rgba(30, 41, 59, 0.6);
  border-color: rgba(71, 85, 105, 0.45);
}
</style>
