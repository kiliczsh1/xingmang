<template>
  <div
    ref="canvasRef"
    class="workflow-canvas"
    :class="{
      'wf-connecting': connecting,
      'wf-connect-fail': connectFailed,
      'wf-drag-over': dragOver
    }"
    tabindex="0"
    @keydown="onKeydown"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <VueFlow
      v-model:nodes="nodes"
      v-model:edges="edges"
      :node-types="nodeTypes"
      :default-viewport="{ x: 0, y: 0, zoom: 0.5 }"
      :min-zoom="0.2"
      :max-zoom="4"
      :delete-key-code="null"
      :default-edge-options="defaultEdgeOptions"
      :only-render-visible-elements="true"
      :connect-on-click="false"
      :is-valid-connection="isValidConnection"
      @init="onInit"
      @connect="onConnect"
      @connect-start="onConnectStart"
      @connect-end="onConnectEnd"
      @node-click="onNodeClick"
      @node-context-menu="onNodeContextMenu"
      @edge-click="onEdgeClick"
      @edge-context-menu="onEdgeContextMenu"
      @pane-click="onPaneClick"
      @move="onVfMove"
    >
      <template #node-start="p"><StartNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-aiProcess="p"><AiProcessNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-end="p"><EndNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-text="p"><TextNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-merge="p"><MergeNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-loop="p"><LoopNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-context="p"><ContextNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-comment="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-transfer="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-fileOutput="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-imageGeneration="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-imageAnalyze="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-fileInput="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-textOutput="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-textInput="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-aiModel="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-textProcessor="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-promptNode="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-videoGeneration="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-imageToVideo="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-imageOutput="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>
      <template #node-videoOutput="p"><GenericNode v-bind="p" :data="withRuntime(p)" /></template>

      <!-- 拖线预览：虚线流动 + 落点圆点 -->
      <template #connection-line="{ sourceX, sourceY, targetX, targetY }">
        <svg class="wf-conn-line">
          <path
            class="wf-conn-path"
            :d="`M${sourceX},${sourceY} C ${sourceX + 60},${sourceY} ${targetX - 60},${targetY} ${targetX},${targetY}`"
          />
          <circle class="wf-conn-dot" :cx="targetX" :cy="targetY" r="4" />
        </svg>
      </template>

      <Background :pattern-color="'#aaa'" :gap="16" />
      <Controls />
      <MiniMap pannable zoomable />
    </VueFlow>

    <EdgeContextMenu
      :visible="edgeMenu.visible"
      :x="edgeMenu.x"
      :y="edgeMenu.y"
      :title="edgeMenu.title"
      @delete="deleteEdgeFromMenu"
      @close="closeEdgeMenu"
    />

    <NodeContextMenu
      :visible="nodeMenu.visible"
      :x="nodeMenu.x"
      :y="nodeMenu.y"
      :node-label="nodeMenu.label"
      :is-breakpoint="nodeMenu.isBreakpoint"
      @toggle-breakpoint="toggleBreakpointFromMenu"
      @test-run="runNodeAction('test-run-node')"
      @rerun="runNodeAction('rerun-node')"
      @delete="deleteNodeFromMenu"
      @close="closeNodeMenu"
    />
  </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, ref, watch } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { VueFlow, type Edge, type Node } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import { MiniMap } from '@vue-flow/minimap'

import StartNode from './nodes/StartNode.vue'
import AiProcessNode from './nodes/AiProcessNode.vue'
import EndNode from './nodes/EndNode.vue'
import TextNode from './nodes/TextNode.vue'
import MergeNode from './nodes/MergeNode.vue'
import LoopNode from './nodes/LoopNode.vue'
import ContextNode from './nodes/ContextNode.vue'
import GenericNode from './nodes/GenericNode.vue'
import EdgeContextMenu from './EdgeContextMenu.vue'
import NodeContextMenu from './NodeContextMenu.vue'
import { validateConnection } from '@/services/workflow/graphUtils'
import { NODE_DND_MIME, createNodeData } from '@/services/workflow/nodeCatalog'
import { useWorkflowStore } from '@/stores/workflow'
import type { NodeType, WEdge, WNode } from '@/types/workflow'

const props = defineProps<{ initialNodes: WNode[]; initialEdges: WEdge[] }>()
const emit = defineEmits<{
  (e: 'node-click', node: WNode): void
  (e: 'edge-click', edge: WEdge): void
  (e: 'nodes-update', nodes: WNode[]): void
  (e: 'edges-update', edges: WEdge[]): void
  (e: 'delete-node', id: string): void
  (e: 'delete-edge', id: string): void
  (e: 'zoom-change', zoom: number): void
  (e: 'node-added', id: string): void
  (e: 'test-run-node', id: string): void
  (e: 'rerun-node', id: string): void
}>()

const store = useWorkflowStore()

/**
 * 把运行时状态（执行状态 / 结果 / 循环进度 / 耗时）叠加到节点 data 上。
 * Vue Flow 内部会自行解析节点对象，直接改 store 里那份 data 驱动不了画布重渲染，
 * 所以执行期数据统一从 store.nodeRuntime 叠加，这里读它即建立渲染依赖。
 */
const withRuntime = (p: any) => {
  const id = String(p?.id ?? p?.node?.id ?? '')
  const rt = store.nodeRuntime[id]
  const base = p?.data || {}
  const merged = rt ? { ...base, ...rt } : { ...base }
  // 断点标记由 store 单独维护，节点卡片据此显示红点
  merged.breakpoint = !!store.breakpoints?.has?.(id)
  return merged
}

const nodeTypes: any = {
  start: 'start',
  aiProcess: 'aiProcess',
  end: 'end',
  text: 'text',
  merge: 'merge',
  loop: 'loop',
  context: 'context',
  comment: 'comment',
  transfer: 'transfer',
  fileOutput: 'fileOutput',
  imageGeneration: 'imageGeneration',
  imageAnalyze: 'imageAnalyze',
  fileInput: 'fileInput',
  textOutput: 'textOutput',
  textInput: 'textInput',
  aiModel: 'aiModel',
  textProcessor: 'textProcessor',
  promptNode: 'promptNode',
  videoGeneration: 'videoGeneration',
  imageToVideo: 'imageToVideo',
  imageOutput: 'imageOutput',
  videoOutput: 'videoOutput'
}

/** 源版默认边：灰色 + 箭头，选中高亮 #409EFF */
const defaultEdgeOptions: any = {
  type: 'default',
  animated: false,
  style: { stroke: '#b1b1b7', strokeWidth: 2 },
  markerEnd: { type: 'arrowclosed', color: '#b1b1b7' }
}
const DEFAULT_EDGE_STYLE = { stroke: '#b1b1b7', strokeWidth: 2 }
const DEFAULT_EDGE_MARKER = { type: 'arrowclosed', color: '#b1b1b7' }
const SELECTED_EDGE_STYLE = { stroke: '#409EFF', strokeWidth: 3 }
const SELECTED_EDGE_MARKER = { type: 'arrowclosed', color: '#409EFF' }

const nodes = ref<any[]>([])
const edges = ref<any[]>([])
const canvasRef = ref<HTMLElement | null>(null)
const selectedNodeId = ref<string | null>(null)
const selectedEdgeId = ref<string | null>(null)

/** 正在拖线（输入端口高亮依据） */
const connecting = ref(false)
/** 拖线被拒（防环/重复）的红色反馈 */
const connectFailed = ref(false)
/** 节点库拖拽悬停画布 */
const dragOver = ref(false)

const edgeMenu = ref<{ visible: boolean; x: number; y: number; edgeId: string | null; title: string }>({
  visible: false,
  x: 0,
  y: 0,
  edgeId: null,
  title: ''
})

const nodeMenu = ref<{
  visible: boolean
  x: number
  y: number
  nodeId: string | null
  label: string
  isBreakpoint: boolean
}>({ visible: false, x: 0, y: 0, nodeId: null, label: '', isBreakpoint: false })

// ⚠ 这里**不能**深拷贝：执行器改的是 store 里的节点 data，画布必须持有同一批对象，
// 否则节点状态色 / 循环进度永远不会跟着变（源版同样依赖这份共享引用）。
watch(
  () => props.initialNodes,
  v => {
    nodes.value = v || []
  },
  { immediate: true }
)
watch(
  () => props.initialEdges,
  v => {
    edges.value = (v || []).map((e: any) => (e.type ? e : { ...e, type: 'default' }))
  },
  { immediate: true }
)

let vf: any = null
let pendingZoom: number | null = null

/**
 * 把视口定格到指定缩放：⚠ VueFlow 1.48 的 fitView 只吃 padding/includeHiddenNodes/duration/nodes，
 * 没有 zoom 选项，「先居中再 zoomTo 固定比例」才是指定缩放的正确写法。
 * 画布尚未初始化时先记下来，onInit 后补做（挂载与 init 之间有一帧空档）。
 */
const applyZoom = async (zoom: number) => {
  if (!vf) {
    pendingZoom = zoom
    return
  }
  try {
    await nextTick()
    // 等 VueFlow 量完节点尺寸，否则 fitView 找不到可框选节点、不会居中
    await new Promise(r => requestAnimationFrame(() => requestAnimationFrame(() => r(null))))
    await vf.fitView?.({ padding: 0.25 })
    await vf.zoomTo?.(zoom, { duration: 0 })
    const actual = vf.getViewport?.()?.zoom
    emit('zoom-change', typeof actual === 'number' ? actual : zoom)
  } catch {
    /* 画布尚未就绪 */
  }
}

const onInit = (instance: any) => {
  vf = instance
  if (pendingZoom !== null) {
    const z = pendingZoom
    pendingZoom = null
    void applyZoom(z)
  }
}

/** 视口移动/缩放即上报（页面角标显示当前缩放百分比） */
const onVfMove = (e: any) => {
  const z = e?.viewport?.zoom
  if (typeof z === 'number' && Number.isFinite(z)) emit('zoom-change', z)
}
const fitView = (opts?: any) => {
  try {
    vf?.fitView?.(opts)
  } catch {
    /* 画布尚未就绪 */
  }
}

// ⚠ 画布对 initialNodes/initialEdges 做了深拷贝，必须挂载后立刻把这两份数组交回上层（store），
// 否则执行器改的是 store 里的旧数组，画布上的状态色/循环进度永远不动。
onMounted(() => {
  emit('nodes-update', nodes.value as WNode[])
  emit('edges-update', edges.value as WEdge[])
})

const resetEdgeStyles = () => {
  edges.value.forEach(e => {
    if (e.id === selectedEdgeId.value) {
      e.style = { ...DEFAULT_EDGE_STYLE }
      e.markerEnd = { ...DEFAULT_EDGE_MARKER }
    }
  })
}

/**
 * 连线校验（Vue Flow 每次拖线都会问这里）：
 * 自连 / 重复端口 / 成环一律拒绝；loop 回边由 graphUtils 单独放行。
 *
 * ⚠ 必须把 Vue Flow 给的连接对象**原样透传**（尤其是 id）：
 * Vue Flow 不只在拖线时问，v-model:edges 每次变更还会把边表里每条已有边（带 id）
 * 送回来复核，判 false 就直接丢弃该边。这里若只挑 source/target/handle 四个字段，
 * id 会丢失，导致「已有边」在查重时匹配到自己 → 刚连好的线被自己删掉。
 */
const isValidConnection = (conn: any) =>
  validateConnection(edges.value as WEdge[], nodes.value as WNode[], conn || {}).ok

let connectSucceeded = false
let connectStartPoint: { x: number; y: number } | null = null

const onConnectStart = (params: any) => {
  connecting.value = true
  connectSucceeded = false
  const ev = params?.event as MouseEvent | undefined
  connectStartPoint = ev ? { x: ev.clientX, y: ev.clientY } : null
}

const onConnectEnd = (ev?: MouseEvent) => {
  connecting.value = false
  // 只有指针真的拖动过（区分「点一下端口就松开」）才亮失败反馈
  const moved =
    !!connectStartPoint &&
    !!ev &&
    Math.hypot(ev.clientX - connectStartPoint.x, ev.clientY - connectStartPoint.y) > 6
  if (!connectSucceeded && moved) {
    connectFailed.value = true
    window.setTimeout(() => (connectFailed.value = false), 450)
  }
  connectStartPoint = null
  connectSucceeded = false
}

const onConnect = (params: any) => {
  const check = validateConnection(edges.value as WEdge[], nodes.value as WNode[], params)
  if (!check.ok) {
    connectFailed.value = true
    window.setTimeout(() => (connectFailed.value = false), 450)
    if (check.reason) ElMessage.warning(check.reason)
    return
  }
  connectSucceeded = true
  const id = `e${params.source}-${params.target}-${edges.value.length}`
  edges.value.push({
    id,
    source: params.source,
    target: params.target,
    sourceHandle: params.sourceHandle,
    targetHandle: params.targetHandle,
    type: 'default',
    animated: false,
    style: { stroke: '#67c23a', strokeWidth: 3 },
    markerEnd: { type: 'arrowclosed', color: '#67c23a' }
  })
  emit('edges-update', edges.value as WEdge[])
  // 成功反馈：新边先绿一下，随后回到默认灰
  window.setTimeout(() => {
    const hit = edges.value.find(e => e.id === id)
    if (!hit) return
    hit.style = { ...DEFAULT_EDGE_STYLE }
    hit.markerEnd = { ...DEFAULT_EDGE_MARKER }
  }, 700)
}

const onNodeClick = ({ node }: { node: Node }) => {
  closeEdgeMenu()
  closeNodeMenu()
  if (selectedEdgeId.value) {
    resetEdgeStyles()
    selectedEdgeId.value = null
  }
  selectedNodeId.value = node.id
  emit('node-click', node as unknown as WNode)
}

const onEdgeClick = ({ edge }: { edge: Edge }) => {
  if (selectedEdgeId.value) {
    const prev = edges.value.find(e => e.id === selectedEdgeId.value)
    if (prev) {
      prev.style = { ...DEFAULT_EDGE_STYLE }
      prev.markerEnd = { ...DEFAULT_EDGE_MARKER }
    }
  }
  selectedEdgeId.value = edge.id
  selectedNodeId.value = null
  const hit = edges.value.find(e => e.id === edge.id)
  if (hit) {
    hit.style = { ...SELECTED_EDGE_STYLE }
    hit.markerEnd = { ...SELECTED_EDGE_MARKER }
  }
  emit('edge-click', edge as unknown as WEdge)
}

// ---------------- 连线右键菜单 ----------------
const edgeLabelOf = (edge: any) => {
  const name = (n: any) => n?.data?.label || n?.type || '节点'
  const from = nodes.value.find(n => n.id === edge?.source)
  const to = nodes.value.find(n => n.id === edge?.target)
  const handle = edge?.sourceHandle ? `（${edge.sourceHandle}）` : ''
  return `${name(from)}${handle} → ${name(to)}`
}

const onEdgeContextMenu = ({ event, edge }: any) => {
  event?.preventDefault?.()
  const rect = canvasRef.value?.getBoundingClientRect()
  edgeMenu.value = {
    visible: true,
    x: (event?.clientX ?? 0) - (rect?.left ?? 0),
    y: (event?.clientY ?? 0) - (rect?.top ?? 0),
    edgeId: edge?.id ?? null,
    title: edgeLabelOf(edge)
  }
}

const closeEdgeMenu = () => {
  if (edgeMenu.value.visible) edgeMenu.value.visible = false
}

const deleteEdgeFromMenu = () => {
  const id = edgeMenu.value.edgeId
  closeEdgeMenu()
  if (!id) return
  edges.value = edges.value.filter(e => e.id !== id)
  if (selectedEdgeId.value === id) selectedEdgeId.value = null
  emit('delete-edge', id)
  emit('edges-update', edges.value as WEdge[])
  ElMessage.success('连接线已删除')
}

// ---------------- 节点库拖入 ----------------
const onDragOver = (e: DragEvent) => {
  if (!e.dataTransfer?.types?.includes(NODE_DND_MIME)) return
  e.preventDefault()
  e.dataTransfer.dropEffect = 'copy'
  dragOver.value = true
}
const onDragLeave = () => {
  dragOver.value = false
}
const onDrop = (e: DragEvent) => {
  dragOver.value = false
  const type = e.dataTransfer?.getData(NODE_DND_MIME)
  if (!type) return
  e.preventDefault()
  addNodeAt(type as NodeType, toFlowCoordinate(e.clientX, e.clientY))
}

/** 屏幕坐标 → 画布坐标（Vue Flow 1.48 用 screenToFlowCoordinate，老版为 project） */
const toFlowCoordinate = (clientX: number, clientY: number) => {
  try {
    if (vf?.screenToFlowCoordinate) return vf.screenToFlowCoordinate({ x: clientX, y: clientY })
    if (vf?.project) return vf.project({ x: clientX, y: clientY })
  } catch {
    /* 画布尚未就绪：退回原点 */
  }
  return { x: 0, y: 0 }
}

/** 新增节点（节点库拖入 / 双击添加共用） */
const addNodeAt = (type: NodeType, position: { x: number; y: number }) => {
  const id = `${type}-${Date.now()}`
  nodes.value.push({
    id,
    type,
    position: { x: Math.round(position.x), y: Math.round(position.y) },
    data: createNodeData(type)
  })
  emit('nodes-update', nodes.value as WNode[])
  emit('node-added', id)
  return id
}

// ---------------- 节点右键菜单（断点 / 单节点测试 / 从此节点重跑 / 删除） ----------------
const onNodeContextMenu = ({ event, node }: any) => {
  event?.preventDefault?.()
  closeEdgeMenu()
  const rect = canvasRef.value?.getBoundingClientRect()
  const id = String(node?.id ?? '')
  if (!id) return
  selectedNodeId.value = id
  nodeMenu.value = {
    visible: true,
    x: (event?.clientX ?? 0) - (rect?.left ?? 0),
    y: (event?.clientY ?? 0) - (rect?.top ?? 0),
    nodeId: id,
    label: node?.data?.label || node?.type || '节点',
    isBreakpoint: !!store.breakpoints?.has?.(id)
  }
}

const closeNodeMenu = () => {
  if (nodeMenu.value.visible) nodeMenu.value.visible = false
}

const onPaneClick = () => {
  closeEdgeMenu()
  closeNodeMenu()
}

const toggleBreakpointFromMenu = () => {
  const id = nodeMenu.value.nodeId
  closeNodeMenu()
  if (!id) return
  store.toggleBreakpoint(id)
  ElMessage.success(store.breakpoints.has(id) ? '已设为断点：执行到该节点会暂停' : '已取消断点')
}

const runNodeAction = (kind: 'test-run-node' | 'rerun-node') => {
  const id = nodeMenu.value.nodeId
  closeNodeMenu()
  if (!id) return
  if (kind === 'test-run-node') emit('test-run-node', id)
  else emit('rerun-node', id)
}

const deleteNodeFromMenu = () => {
  const id = nodeMenu.value.nodeId
  closeNodeMenu()
  if (id) void requestDeleteNode(id)
}

/** 节点删除（键盘 Delete 与右键菜单共用） */
const requestDeleteNode = async (id: string) => {
  const node = nodes.value.find(n => n.id === id)
  if (!node) return
  try {
    await ElMessageBox.confirm(`确定要删除节点"${node.data.label || node.type}"吗？`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  nodes.value = nodes.value.filter(n => n.id !== id)
  edges.value = edges.value.filter(ed => ed.source !== id && ed.target !== id)
  if (selectedNodeId.value === id) selectedNodeId.value = null
  emit('nodes-update', nodes.value as WNode[])
  emit('edges-update', edges.value as WEdge[])
  emit('delete-node', id)
  ElMessage.success('节点已删除')
}

/** 源版：Delete / Backspace 删除选中的节点或连线（带确认） */
const onKeydown = async (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    closeEdgeMenu()
    closeNodeMenu()
    return
  }
  if (e.key !== 'Delete' && e.key !== 'Backspace') return
  closeEdgeMenu()
  closeNodeMenu()
  e.preventDefault()
  if (selectedNodeId.value) {
    await requestDeleteNode(selectedNodeId.value)
  } else if (selectedEdgeId.value) {
    const edge = edges.value.find(ed => ed.id === selectedEdgeId.value)
    if (!edge) return
    edges.value = edges.value.filter(ed => ed.id !== selectedEdgeId.value)
    selectedEdgeId.value = null
    emit('delete-edge', edge.id)
    emit('edges-update', edges.value as WEdge[])
    ElMessage.success('连接线已删除')
  }
}

defineExpose({
  nodes,
  edges,
  addNode: (node: any) => {
    nodes.value.push(node)
  },
  addNodeAt,
  removeNode: (id: string) => {
    nodes.value = nodes.value.filter(n => n.id !== id)
    edges.value = edges.value.filter(e => e.source !== id && e.target !== id)
  },
  updateNode: (id: string, patch: Record<string, any>) => {
    const node = nodes.value.find(n => n.id === id)
    if (node) Object.assign(node.data, patch)
  },
  fitView,
  applyInitialZoom: applyZoom
})

// 图结构变化即上报（IDE 层据此写回 store / JSON 内容）
watch(nodes, () => emit('nodes-update', nodes.value as WNode[]), { deep: true })
watch(edges, () => emit('edges-update', edges.value as WEdge[]), { deep: true })
</script>

<style>
/* 源版：画布铺满，页面底色 */
.workflow-canvas {
  width: 100%;
  height: 100%;
  background: var(--wf-el-bg-page, #f2f3f5);
  outline: none;
  /* 连线右键菜单按此定位（EdgeContextMenu 用 offsetParent 换算边界） */
  position: relative;
}
.wf-dark .workflow-canvas {
  background: #1a1a1a;
}
.vue-flow__node {
  font-family: inherit;
}
.vue-flow__handle {
  width: 10px;
  height: 10px;
  background: var(--el-color-primary, #409eff);
  border: 2px solid #fff;
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}
.vue-flow__handle-left {
  left: -6px;
}
.vue-flow__handle-right {
  right: -6px;
}
/* 拖线中：接入端口放大高亮，提示可落点 */
.wf-connecting .vue-flow__handle-left {
  transform: scale(1.7);
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.25);
}
.wf-connecting .vue-flow__handle {
  cursor: crosshair;
}
/* 拖线被拒：端口红色闪一下 */
.wf-connect-fail .vue-flow__handle-left {
  background: var(--el-color-danger, #f56c6c);
  box-shadow: 0 0 0 5px rgba(245, 108, 108, 0.35);
}
/* 节点库拖入时画布描边 */
.workflow-canvas.wf-drag-over {
  box-shadow: inset 0 0 0 2px var(--el-color-primary, #409eff);
}
/* 拖线预览 */
.wf-conn-line {
  overflow: visible;
  pointer-events: none;
}
.wf-conn-path {
  fill: none;
  stroke: var(--el-color-primary, #409eff);
  stroke-width: 2;
  stroke-dasharray: 6 4;
  animation: wf-dash-flow 0.6s linear infinite;
}
.wf-conn-dot {
  fill: var(--el-color-primary, #409eff);
  opacity: 0.85;
}
@keyframes wf-dash-flow {
  to {
    stroke-dashoffset: -20;
  }
}
@media (prefers-reduced-motion: reduce) {
  .wf-conn-path {
    animation: none;
  }
}
/* 暗色下 VueFlow 自带控件/小地图/连线标签的默认浅色底 */
.wf-dark .vue-flow__controls {
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.5);
}
.wf-dark .vue-flow__controls-button {
  background: #2a2a2a;
  border-bottom: 1px solid #414243;
  color: #cfd3dc;
  fill: #cfd3dc;
}
.wf-dark .vue-flow__controls-button:hover {
  background: #303030;
}
.wf-dark .vue-flow__minimap {
  background: #2a2a2a;
  border: 1px solid #414243;
}
.wf-dark .vue-flow__minimap-mask {
  fill: rgba(255, 255, 255, 0.08);
}
.vue-flow__attribution {
  display: none;
}
</style>
