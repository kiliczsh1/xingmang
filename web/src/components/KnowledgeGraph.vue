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
    />

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
          @hover-entity="onHoverEntity"
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

    <!--
      实体详情浮层：贴着鼠标（也就是贴着正在悬停的那个节点）出现，
      鼠标移到别的实体上就跟着挪过去，靠近视口边缘时自动翻向另一侧。
      这里刻意不用 el-dialog / el-drawer —— 它们做的是"居中/贴边"这类固定定位，
      而悬停预览要的是"跟着目标走"，只能自己算坐标。
    -->
    <div
      v-if="detailVisible"
      class="graph-detail-float"
      :style="{ left: `${detailPos.x}px`, top: `${detailPos.y}px` }"
      @mouseenter="onDetailEnter"
      @mouseleave="onDetailLeave"
    >
      <GraphDetailPanel
        :entity="detailEntity"
        :degree="detailDegree"
        :neighbors="detailNeighbors"
        :focus-depth="store.focusDepth"
        @close="closeDetail"
        @select="store.selectEntity"
        @set-focus-depth="handleFocusDepth"
      />
    </div>

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
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { Collection, Connection, InfoFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import GraphCanvas from '@/components/graph/GraphCanvas.vue'
import GraphToolbar from '@/components/graph/GraphToolbar.vue'
import GraphLegend from '@/components/graph/GraphLegend.vue'
import GraphDetailPanel, { type GraphNeighbor } from '@/components/graph/GraphDetailPanel.vue'
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

// ── 详情弹窗：由「悬停」驱动，与「点击选中」解耦 ─────────────
//
// 悬停 = 看详情，点击 = 只高亮该节点的连线（画布侧 nodeReducer/edgeReducer 负责压暗其余）。
// 两者用的是两套状态，所以悬停不会改变画布高亮、点击也不会弹窗。

/** 当前悬停的实体 id —— 弹窗显示的是它，不是 selectedEntityId */
const hoveredEntityId = ref<number | null>(null)
const detailVisible = ref(false)

/**
 * 关窗延迟。
 * 鼠标从节点移向弹窗的途中必然要经过画布空白区（会触发 leaveNode），
 * 立刻关掉的话弹窗根本够不着，所以留一点缓冲。
 */
let closeTimer: ReturnType<typeof setTimeout> | null = null

function clearCloseTimer() {
  if (closeTimer !== null) {
    clearTimeout(closeTimer)
    closeTimer = null
  }
}

function closeDetail() {
  clearCloseTimer()
  detailVisible.value = false
  hoveredEntityId.value = null
}

/** 浮层尺寸：宽度固定，高度取上限，两者都用于边缘避让计算 */
const DETAIL_WIDTH = 420
const DETAIL_MAX_HEIGHT = 460

/** 浮层左上角坐标（视口坐标系，元素用 position: fixed 定位） */
const detailPos = ref({ x: 0, y: 0 })

/**
 * 让浮层出现在光标旁边：默认落在右下 20px 处；
 * 右侧/下方放不下就翻到反方向，仍放不下再贴边 —— 保证整块始终可见。
 */
function placeDetail(clientX: number, clientY: number) {
  const vw = window.innerWidth
  const vh = window.innerHeight
  const width = Math.min(DETAIL_WIDTH, vw - 24)
  const height = Math.min(DETAIL_MAX_HEIGHT, vh - 24)

  let x = clientX + 20
  if (x + width > vw - 12) x = clientX - width - 20

  let y = clientY + 16
  if (y + height > vh - 12) y = clientY - height - 16

  detailPos.value = {
    x: Math.max(12, Math.min(x, vw - width - 12)),
    y: Math.max(12, Math.min(y, vh - height - 12))
  }
}

function onHoverEntity(payload: { id: number | null; clientX: number; clientY: number }) {
  clearCloseTimer()
  if (payload.id !== null) {
    hoveredEntityId.value = payload.id
    detailVisible.value = true
    // 每次都重新定位：鼠标移到哪个实体，浮层就跟到哪个实体旁边
    placeDetail(payload.clientX, payload.clientY)
    return
  }
  closeTimer = setTimeout(closeDetail, 260)
}

/** 鼠标移入弹窗：取消待关闭（用户正要看详情） */
function onDetailEnter() {
  clearCloseTimer()
}

/** 鼠标移出弹窗：直接关（此刻鼠标已经不在任何节点上） */
function onDetailLeave() {
  closeDetail()
}

/** 弹窗里展示的实体 / 度数 / 邻居，全部基于悬停实体计算 */
const detailEntity = computed(() => {
  const id = hoveredEntityId.value
  return id === null ? null : store.graphIndex.entityById.get(id) || null
})

const detailDegree = computed(() => {
  const id = hoveredEntityId.value
  return id === null ? 0 : store.graphIndex.degree.get(id) || 0
})

/**
 * 弹窗展示的邻居列表。
 * 结构与 store 的 selectedNeighbors 一致（relation + 对端实体 + 方向），
 * 只是把"选中实体"换成了"悬停实体" —— 详情面板靠 relation/direction 渲染关系标签与箭头。
 */
const detailNeighbors = computed<GraphNeighbor[]>(() => {
  const id = hoveredEntityId.value
  if (id === null) return []
  return store.collapsedRelations
    .filter(relation => relation.source_id === id || relation.target_id === id)
    .map(relation => {
      const otherId = relation.source_id === id ? relation.target_id : relation.source_id
      return {
        relation,
        entity: store.graphIndex.entityById.get(otherId) || null,
        direction: (relation.source_id === id ? 'out' : 'in') as 'out' | 'in'
      }
    })
})

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

/**
 * 邻域聚焦以「选中实体」为种子。悬停预览时用户往往还没点击过，
 * 所以点聚焦档位时顺手把当前悬停的实体选中，否则聚焦没有起点、不生效。
 */
function handleFocusDepth(depth: number | null) {
  if (depth !== null && hoveredEntityId.value !== null) {
    store.selectEntity(hoveredEntityId.value)
  }
  store.setFocusDepth(depth)
}

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

// 卸载时清掉待关闭计时器，避免回调打到已销毁的组件上
onUnmounted(() => {
  clearCloseTimer()
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
 * 实体详情浮层。
 * position: fixed + 视口坐标（坐标由 placeDetail() 算好、已做边缘避让），
 * 这样它才能"贴着悬停的节点出现"并跟随移动 ——
 * el-dialog / el-drawer 做不到，它们本质是固定的居中/贴边定位。
 */
.graph-detail-float {
  position: fixed;
  z-index: 2200;
  display: flex;
  flex-direction: column;
  width: min(420px, 92vw);
  max-height: 70vh;
  overflow-y: auto;
  background: var(--el-bg-color-overlay, #fff);
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  box-shadow: var(--el-box-shadow);
  animation: detail-float-in 0.12s ease-out;
}

@keyframes detail-float-in {
  from {
    opacity: 0;
    transform: translateY(4px);
  }
  to {
    opacity: 1;
    transform: none;
  }
}

.graph-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.graph-empty p {
  margin: 0;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.graph-empty .empty-sub {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
  color: var(--el-color-info-light-5);
}

:root[data-theme='dark'] .graph-notice {
  background: rgba(245, 158, 11, 0.1);
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .graph-empty p {
  color: #9ca3af;
}

:root[data-theme='dark'] .empty-icon {
  color: #4d5768;
}
</style>
