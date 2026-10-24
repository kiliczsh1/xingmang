<template>
  <div class="knowledge-graph-panel">
    <GraphToolbar
      :entity-count="store.entities.length"
      :relation-count="store.relations.length"
      :show-edge-labels="showEdgeLabels"
      :show-arrows="showArrows"
      :loading="store.loading"
      :search-keyword="store.searchKeyword"
      :match-count="store.searchMatches.length"
      :match-cursor="searchCursor"
      @update:search-keyword="handleSearchInput"
      @toggle-edge-labels="showEdgeLabels = !showEdgeLabels"
      @toggle-arrows="showArrows = !showArrows"
      @zoom-in="canvasRef?.zoomIn()"
      @zoom-out="canvasRef?.zoomOut()"
      @fit-view="canvasRef?.fitView()"
      @reset-view="canvasRef?.resetView()"
      @refresh="handleRefresh"
      @search-next="focusNextMatch"
      @analyze="analyzeVisible = true"
      :link-mode="linkMode"
      @toggle-link-mode="linkMode = !linkMode"
      @open-force="forceDrawerVisible = true"
    >
      <!-- 透传页面级控件（书籍/版本）与尾操作（导入导出），并进工具栏同一行 -->
      <template v-if="$slots['toolbar-extra']" #extra>
        <slot name="toolbar-extra" />
      </template>
      <template v-if="$slots['toolbar-tail']" #tail>
        <slot name="toolbar-tail" />
      </template>
    </GraphToolbar>

    <div v-if="store.legacyData" class="graph-notice">
      <el-icon><InfoFilled /></el-icon>
      <span>当前展示的是未按版本归档的历史数据；在侧栏「新建」保存为新版本后即可分版本管理</span>
    </div>

    <div v-if="store.selectedChapterId !== null" class="graph-notice is-chapter">
      <el-icon><Collection /></el-icon>
      <span>
        正在按章节过滤：{{ selectedChapterLabel }}
      </span>
      <el-button link size="small" @click="store.selectChapter(null)">取消过滤</el-button>
    </div>

    <div class="graph-body">
      <GraphSidePanel v-if="withSidePanel" />

      <div class="graph-main">
        <GraphCanvas
          ref="canvasRef"
          :entities="store.visibleEntities"
          :relations="store.visibleRelations"
          :selected-id="store.selectedEntityId"
          :highlighted-ids="highlightedIds"
          :show-edge-labels="showEdgeLabels"
          :link-mode="linkMode"
          :force-params="forceParams"
          :show-arrows="showArrows"
          @select="store.selectEntity"
          @link="handleLink"
          @exit-link="linkMode = false"
        >
          <template #empty>
            <div class="graph-empty">
              <el-icon class="empty-icon"><Connection /></el-icon>
              <p v-if="store.isEmpty">暂无图谱数据</p>
              <p v-else>当前筛选或聚焦条件下没有可显示的实体</p>
              <p v-if="store.isEmpty" class="empty-sub">
                可用「AI 分析」从章节抽取，或在左栏「导入」里导入 JSON
              </p>
            </div>
          </template>
        </GraphCanvas>

        <GraphLegend
          v-if="!store.isEmpty"
          :type-counts="store.typeCounts"
          :active-types="store.filters.types"
          :entities="store.entities"
          @toggle-type="store.toggleTypeFilter"
          @clear-types="store.clearTypeFilter"
        />

        <!--
          实体详情：**浮在画布右侧**的栏（absolute 覆盖，不改画布尺寸）。
          点节点才出现；关掉靠面板上的关闭按钮，或点画布空白处取消选中。

          为什么不用 flex 挤压：画布宽度一变，sigma 就按新宽度重新映射相机，
          整张图往左跳一下 —— 同一缩放下选中前后看着是两个视图位置。
          改成覆盖后画布尺寸恒定，选中前后视图完全一致。
        -->
        <aside v-if="detailVisible" class="graph-detail-aside">
          <!--
            侧栏顶部的 tab：详情 / 关联实体 / 词条合并。
            合并是"全局整理"的动作，不需要先选中某个实体，
            所以它单独占一个 tab；关联实体条数多，也从详情里拆出来单独一页。
          -->
          <div class="aside-tabs">
            <button
              class="aside-tab"
              :class="{ on: asideTab === 'detail' }"
              @click="asideTab = 'detail'"
            >
              详情
            </button>
            <button
              class="aside-tab"
              :class="{ on: asideTab === 'relations' }"
              @click="asideTab = 'relations'"
            >
              关联实体<span v-if="detailNeighbors.length" class="aside-tab-n">{{ detailNeighbors.length }}</span>
            </button>
            <button
              class="aside-tab"
              :class="{ on: asideTab === 'merge' }"
              @click="asideTab = 'merge'"
            >
              词条合并
            </button>
          </div>

          <GraphDetailPanel
            v-if="asideTab === 'detail'"
            :entity="detailEntity"
            :degree="detailDegree"
            :neighbors="detailNeighbors"
            section="info"
            @select="store.selectEntity"
          />
          <GraphDetailPanel
            v-else-if="asideTab === 'relations'"
            :entity="detailEntity"
            :degree="detailDegree"
            :neighbors="detailNeighbors"
            section="relations"
            @select="store.selectEntity"
          />
          <GraphMergePanel
            v-else
            :book-id="bookId"
            @reveal="handleRevealFromMerge"
          />
        </aside>
      </div>

    </div>

    <!-- 力度设置：右侧抽屉，平时不占画布空间 -->
    <el-drawer
      v-model="forceDrawerVisible"
      title="力度"
      direction="rtl"
      size="320px"
      append-to-body
    >
      <GraphForcePanel
        :params="forceParams"
        @update="handleForceUpdate"
        @reset="handleForceReset"
      />
    </el-drawer>

    <GraphAnalyzeDialog v-model:visible="analyzeVisible" @saved="handleAnalyzed" />

    <GraphRelationDialog
      v-model:visible="relationVisible"
      :source="linkEndpoints.source"
      :target="linkEndpoints.target"
      :existing-types="existingRelationTypes"
      :duplicate="linkDuplicate"
      :saving="relationSaving"
      @submit="handleRelationSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { Close, Collection, Connection, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import GraphCanvas from '@/components/graph/GraphCanvas.vue'
import GraphToolbar from '@/components/graph/GraphToolbar.vue'
import GraphLegend from '@/components/graph/GraphLegend.vue'
import GraphDetailPanel, { type GraphNeighbor } from '@/components/graph/GraphDetailPanel.vue'
import GraphMergePanel from '@/components/graph/GraphMergePanel.vue'
import GraphForcePanel from '@/components/graph/GraphForcePanel.vue'
import GraphSidePanel from '@/components/graph/GraphSidePanel.vue'
import GraphAnalyzeDialog from '@/components/graph/GraphAnalyzeDialog.vue'
import GraphRelationDialog from '@/components/graph/GraphRelationDialog.vue'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'
import { DEFAULT_FORCE_PARAMS, type ForceParams } from '@/composables/useForceLayout'
import type { GraphEntity, GraphRelation } from '@/types'

const props = withDefaults(
  defineProps<{
    bookId: number
    /** WikiGraph 页面自己已经有侧栏，避免出现两个 */
    withSidePanel?: boolean
  }>(),
  { withSidePanel: true }
)

const store = useKnowledgeGraphStore()
const canvasRef = ref<InstanceType<typeof GraphCanvas> | null>(null)

// 默认关闭关系标签：与 GraphCanvas 的默认值保持一致（视觉降噪）
const showEdgeLabels = ref(false)
// 默认不显示方向箭头（Obsidian 观感）；单向关系看不出指向时再打开
const showArrows = ref(false)
const searchCursor = ref(0)
const analyzeVisible = ref(false)

/**
 * 力度参数是受控的：状态留在这里，画布只负责应用。
 * 这样面板读数不会因为画布重建（切换版本/筛选）而被重置。
 */
const forceParams = ref<ForceParams>({ ...DEFAULT_FORCE_PARAMS })

function handleForceUpdate(patch: Partial<ForceParams>) {
  forceParams.value = { ...forceParams.value, ...patch }
  canvasRef.value?.setForceParams(patch)
}

function handleForceReset() {
  forceParams.value = { ...DEFAULT_FORCE_PARAMS }
  canvasRef.value?.setForceParams({ ...DEFAULT_FORCE_PARAMS })
}

/** 力度抽屉开关 */
const forceDrawerVisible = ref(false)

// ── 实体详情：由「点击选中」驱动 ─────────────────────────
//
// 早期是「悬停出浮层、点击只高亮连线」，两套状态分开。
// 现在统一成一套：点节点 = 选中（画布清场聚焦）+ 右侧栏出详情。
// 这样做的好处是不用再纠结浮层遮住节点、鼠标够不着卡片这些事。

/** 侧栏当前 tab：详情 / 关联实体 / 词条合并 */
const asideTab = ref<'detail' | 'relations' | 'merge'>('detail')

/** 详情栏是否可见 —— 直接跟随选中状态 */
const detailVisible = computed(() => store.selectedEntityId !== null)

/** 当前展示的实体（= 选中实体） */
const detailEntity = computed(() => store.selectedEntity)

/**
 * 在合并面板里点某个实体 → 切回详情 tab 并选中它。
 * 用户此刻想确认"这两个到底是不是同一个东西"，必须能看到它的详情和图。
 */
function handleRevealFromMerge(id: number) {
  asideTab.value = 'detail'
  store.selectEntity(id)
}

/** 详情栏展示的度数（按选中实体算） */
const detailDegree = computed(() => {
  const id = store.selectedEntityId
  return id === null ? 0 : store.graphIndex.degree.get(id) || 0
})

/**
 * 详情栏的邻居列表。
 *
 * 结构直接复用 store.selectedNeighbors（relation + 对端实体 + 方向），
 * 面板靠 relation / direction 渲染关系标签与箭头。
 */
const detailNeighbors = computed<GraphNeighbor[]>(() => store.selectedNeighbors)

/** 当前过滤的章节标题，用于顶部提示 */
const selectedChapterLabel = computed(() => {
  const id = store.selectedChapterId
  if (id === null) return ''
  const index = store.chapters.findIndex(c => c.id === id)
  if (index < 0) return `章节 #${id}`
  return store.chapters[index].title || `第${index + 1}章`
})

/** 搜索命中与 D 阶段的检索命中都走高亮通道 */
const highlightedIds = computed(() => {
  const ids = new Set(store.highlightedIds)
  for (const entity of store.searchMatches) ids.add(entity.id)
  return [...ids]
})

function focusMatch(index: number) {
  const matches = store.searchMatches
  if (matches.length === 0) return
  const cursor = ((index % matches.length) + matches.length) % matches.length
  searchCursor.value = cursor
  store.selectEntity(matches[cursor].id)
  canvasRef.value?.focusNode(matches[cursor].id)
}

function handleSearchInput(keyword: string) {
  store.setSearchKeyword(keyword)
  searchCursor.value = 0
  focusMatch(0)
}

function focusNextMatch() {
  focusMatch(searchCursor.value + 1)
}

async function handleRefresh() {
  await store.refresh()
  if (store.entities.length > 0) {
    ElMessage.success('刷新成功')
  } else {
    ElMessage.warning('暂无图谱数据')
  }
}

/** AI 分析存为新版本后，版本与章节关联都变了，提示一下当前状态 */
async function handleAnalyzed(versionName: string) {
  ElMessage.success(`已切换到新版本：${versionName}`)
}

// ── 人工连线 ────────────────────────────────────────────
const linkMode = ref(false)
const relationVisible = ref(false)
const relationSaving = ref(false)
const linkEndpoints = ref<{ source: GraphEntity | null; target: GraphEntity | null }>({
  source: null,
  target: null
})

/** 当前图谱已用的关系类型（去重、排除人工边前缀），供表单快捷选择 */
const existingRelationTypes = computed(() => {
  const types = new Set<string>()
  for (const rel of store.relations) {
    const t = (rel.relation_type || '').trim()
    if (t) types.add(t)
  }
  return [...types].slice(0, 12)
})

/** 两端之间已有的关系，用于提示会产生平行边 */
const linkDuplicate = computed<GraphRelation | null>(() => {
  const { source, target } = linkEndpoints.value
  if (!source || !target) return null
  return (
    store.relations.find(
      r => r.source_id === source.id && r.target_id === target.id
    ) || null
  )
})

function handleLink(request: { sourceId: number; targetId: number }) {
  const source = store.entities.find(e => e.id === request.sourceId) || null
  const target = store.entities.find(e => e.id === request.targetId) || null
  if (!source || !target) {
    ElMessage.warning('连线端点已不在当前图谱中')
    return
  }
  linkEndpoints.value = { source, target }
  relationVisible.value = true
}

async function handleRelationSubmit(payload: {
  sourceId: number
  targetId: number
  relationType: string
  description: string
}) {
  relationSaving.value = true
  try {
    const res = await store.createManualRelation(payload)
    if (res.ok) {
      ElMessage.success(res.message)
      relationVisible.value = false
    } else {
      ElMessage.error(res.message)
    }
  } finally {
    relationSaving.value = false
  }
}

onMounted(() => {
  store.loadBook(props.bookId)
})

watch(
  () => props.bookId,
  id => {
    if (id) store.loadBook(id)
  }
)
</script>

<style scoped>
.knowledge-graph-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.graph-notice {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 16px;
  font-size: 12px;
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.graph-notice.is-chapter {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.graph-body {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}

.graph-main {
  position: relative; /* 毛玻璃图例 absolute 定位的锚点 */
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.graph-main :deep(.graph-canvas) {
  flex: 1;
  min-height: 0;
}

/**
 * 实体详情：右侧栏。
 *
 * 从"贴鼠标的浮层"改成常驻右栏，画布让出一块宽度。
 * 好处：不遮节点、鼠标不用横穿画布、滚动不影响画布缩放。
 */
.graph-detail-aside {
  /*
   * 覆盖式（absolute）而不是 flex 挤压。
   * 用 flex 会改画布宽度 → sigma 按新宽度重映射相机 → 整张图跳位。
   * 覆盖则画布尺寸恒定，选中前后视图完全一致。
   */
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  z-index: 5;
  width: 340px;
  display: flex;
  flex-direction: column;
  overflow: hidden;          /* 滚动交给内部各 tab */
  border-left: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color-overlay, #fff);
  box-shadow: -8px 0 24px rgba(15, 23, 42, 0.08);
  animation: detail-aside-in 0.16s ease-out;
}

/* 侧栏顶栏：两个 tab + 关闭 */
.aside-tabs {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.aside-tab {
  appearance: none;
  border: none;
  background: transparent;
  padding: 4px 10px;
  border-radius: 6px;
  font: inherit;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  transition: background 0.15s ease, color 0.15s ease;
}

.aside-tab:hover {
  background: var(--el-fill-color-light);
  color: var(--el-text-color-primary);
}

.aside-tab.on {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 500;
}
.aside-tab-n {
  margin-left: 4px;
  padding: 0 5px;
  border-radius: 8px;
  background: var(--el-fill-color);
  font-size: 10.5px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
}
.aside-tab.on .aside-tab-n {
  background: rgba(59, 130, 246, 0.2);
  color: inherit;
}
:root[data-theme='dark'] .aside-tab-n {
  background: #1c2333;
  color: #8b98a8;
}

:root[data-theme='dark'] .aside-tabs {
  border-bottom-color: rgba(42, 50, 66, 0.6);
}

:root[data-theme='dark'] .aside-tab:hover {
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .aside-tab.on {
  background: color-mix(in srgb, var(--el-color-primary) 22%, transparent);
  color: var(--el-color-primary-light-3);
}

@keyframes detail-aside-in {
  from {
    opacity: 0;
    transform: translateX(8px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

/* 窄屏（图谱弹窗里）自动收窄，避免把画布挤没 */
@media (max-width: 1280px) {
  .graph-detail-aside {
    width: 300px;
  }
}

:root[data-theme='dark'] .graph-detail-aside {
  /* ⚠ 不能用 var(--el-bg-color-overlay)：项目只引了 EP 亮色变量，这个变量在暗色下仍是白色 */
  background: #161b22;
  border-left-color: rgba(42, 50, 66, 0.9);
  box-shadow: -8px 0 24px rgba(0, 0, 0, 0.45);
  color: #e6edf3;
}

/* 尊重减少动效偏好 */
@media (prefers-reduced-motion: reduce) {
  .graph-detail-aside {
    animation: none;
  }
}

</style>
