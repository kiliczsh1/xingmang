<template>
  <div class="graph-canvas" ref="wrapperRef" :class="{ 'is-link-mode': linkMode }">
    <div ref="containerRef" class="sigma-container"></div>

    <!-- 连线模式的橡皮筋预览层：只画一条临时线，pointer-events 关闭 -->
    <canvas
      v-show="linking"
      ref="linkCanvasRef"
      class="graph-link-overlay"
      :class="{ 'is-linking': linking }"
    ></canvas>

    <!-- 连线模式提示条 -->
    <div v-if="linkMode" class="graph-link-hint">
      <el-icon><Share /></el-icon>
      <span>{{ linkHintText }}</span>
      <el-button link size="small" @click="emit('exit-link')">退出连线（Esc）</el-button>
    </div>

    <!-- 悬停不再弹浮动信息卡：详情统一走居中弹窗（由 hover-entity 事件驱动），避免两处重复显示 -->

    <div v-if="entities.length === 0" class="graph-canvas-empty">
      <slot name="empty" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import Sigma from 'sigma'
import type { EdgeDisplayData, NodeDisplayData } from 'sigma/types'
import { EdgeArrowProgram, EdgeLineProgram, type NodeLabelDrawingFunction } from 'sigma/rendering'
import Graph from 'graphology'
import EdgeCurveProgram, { EdgeCurvedArrowProgram } from '@sigma/edge-curve'
import { NodeBorderProgram } from '@sigma/node-border'
import { Share } from '@element-plus/icons-vue'
import type { GraphEntity, GraphRelation } from '@/types'
import { buildGraphIndex, groupRelationsByPair, pairKey } from '@/services/knowledgeGraphService'
import { computeLayout } from '@/composables/useGraphLayout'
import {
  DEFAULT_FORCE_PARAMS,
  ForceSimulator,
  NODE_RADIUS,
  type ForceParams
} from '@/composables/useForceLayout'
import { MANUAL_EVIDENCE, typeColor } from '@/utils/graphColor'

/** 节点/边在图中的自定义属性（叠加在 sigma 的 DisplayData 之上） */
type NodeAttrs = NodeDisplayData & {
  description?: string
  degree?: number
  /** 业务实体类型（character/location/...），sigma 的 type 字段留给渲染 program 用 */
  entityType?: string
  /** @sigma/node-border 的描边颜色 */
  borderColor?: string
}
type EdgeAttrs = EdgeDisplayData & {
  relation_type?: string
  /** 曲线弯曲比例（@sigma/edge-curve 的 curvature，正负决定弯曲方向） */
  curvature?: number
  /** 人工连线添加的边（evidence==='人工添加'），用绿色标识 */
  manual?: boolean
}

interface LinkRequest {
  sourceId: number
  targetId: number
}

const props = withDefaults(
  defineProps<{
    entities: GraphEntity[]
    relations: GraphRelation[]
    selectedId: number | null
    /** 需要高亮的节点（检索命中 / 搜索命中） */
    highlightedIds?: number[]
    showEdgeLabels?: boolean
    /** 连线模式：在节点间拖出橡皮筋，落点确认后由父组件建关系 */
    linkMode?: boolean
    /** 力度参数：作为新建力模拟器时的初值（运行时改动走 setForceParams） */
    forceParams?: ForceParams
    /**
     * 是否给**所有**边显示方向箭头。
     * 默认关闭：Obsidian 不画箭头，满屏箭头是最大的方向噪声源。
     * 单向关系（只存在 A→B 一条）在默认状态下看不出指向哪边，需要时打开这个开关即可。
     */
    showArrows?: boolean
  }>(),
  {
    highlightedIds: () => [],
    // 默认不显示关系标签：满屏边标签是最大的视觉噪声源（Obsidian 也不显示）
    showEdgeLabels: false,
    linkMode: false,
    forceParams: () => ({ ...DEFAULT_FORCE_PARAMS }),
    showArrows: false
  }
)

const emit = defineEmits<{
  (e: 'select', id: number | null): void
  (e: 'link', request: LinkRequest): void
  (e: 'exit-link'): void
  /**
   * 悬停实体：悬停只负责"看详情"（页面层据此开关弹窗），
   * 与 select（点击，负责高亮连线）刻意分开，两者互不干扰。
   * 带上视口坐标，是为了让详情弹窗能避开鼠标所在的一侧、不挡住正在看的节点。
   */
  (e: 'hover-entity', payload: { id: number | null; clientX: number; clientY: number }): void
}>()

const wrapperRef = ref<HTMLElement>()
const containerRef = ref<HTMLElement>()
const linkCanvasRef = ref<HTMLCanvasElement>()

// ── 连线模式（橡皮筋拉线） ───────────────────────────────
/** overlay 是否可见（正在拖或已按下等待） */
const linking = ref(false)
/** 拉线起点节点 key（graph 内 string key），null 表示尚未按下 */
let linkFromKey: string | null = null
/** 橡皮筋当前端点（容器像素坐标） */
let linkCursor = { x: 0, y: 0 }
/** 悬停命中的目标节点 key（用于高亮提示可放置） */
let linkTargetKey: string | null = null
/** 粘滞态：在起点上按下未拖动即松开，等待第二次点击目标节点（触屏/精确点选友好） */
let linkSticky = false

const linkHintText = computed(() =>
  !linking.value
    ? '连线模式：在起点节点上按住并拖到终点节点松开（也可点起点再点终点）'
    : '已选起点，拖到/点击高亮的目标节点完成连线；点空白处取消'
)

// ── sigma 实例与图（非响应式） ───────────────────────────
let sigma: Sigma<NodeAttrs, EdgeAttrs> | null = null
const graph = new Graph<NodeAttrs, EdgeAttrs>()

// 只存纯数据，不放进响应式，避免每帧刷新
let adjacency = new Map<number, Set<number>>()

// 强调状态的快照，reducer 读取；变化后 sigma.refresh()
let snapSelectedId: number | null = null
let snapHighlighted = new Set<number>()
let hoveredNodeId: string | null = null
/** 正在被拖拽的节点 key —— sigma 核心不含节点拖拽能力，只能自己维护状态 */
let draggedNode: string | null = null

// 主题（sigma 是 canvas 绘制，颜色必须在 JS 里决定）
let themeDark = false

// ── 配色（随主题） ──────────────────────────────────────
function isDark(): boolean {
  return document.documentElement.getAttribute('data-theme') === 'dark'
}

function edgeColor(): string {
  /**
   * 中性灰（对标 Obsidian 的边色），不用偏蓝的灰。
   *
   * 关于浓度：边要比原来淡，但 alpha 不能压太低 ——
   * rgba(154,164,178,0.45) 在白色画布上混合后约等于 rgb(209,214,221)，
   * 已经和背景几乎同色，用户看到的就是"节点之间没有关系线"。
   * 这里取"淡但不牺牲可见"的值（混合后约 rgb(174,174,174)）。
   */
  return themeDark ? 'rgba(138,138,142,0.62)' : 'rgba(150,150,150,0.78)'
}

function manualEdgeColor(): string {
  return themeDark ? '#34d399' : '#10b981'
}

function edgeDimColor(): string {
  return themeDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)'
}

function nodeDimColor(): string {
  return themeDark ? '#1e2634' : '#eef0f3'
}

/**
 * 节点描边色：与画布底色同色，作用是让相邻节点之间有"留白"而不是糊成一坨。
 * 深色主题必须换成深色 —— 原来的固定白描边在暗色下会给每个节点套一个亮圈。
 */
function nodeBorderColor(): string {
  return themeDark ? 'rgba(18,22,30,0.9)' : 'rgba(255,255,255,0.92)'
}

function selectedColor(): string {
  return '#f59e0b'
}

function hoverColor(): string {
  return '#f59e0b'
}

function labelColor(): string {
  return themeDark ? '#e6edf3' : '#1f2329'
}

function edgeLabelColor(): string {
  return themeDark ? '#8b98a9' : '#6b7280'
}

/**
 * 节点标签：裸文字 + 极细描边（对标 Obsidian）。
 *
 * 为什么不再用白底药丸：同时显示几十个节点时，一个个白色圆角块会在画布上排成
 * "白点阵列"，比文字本身更抢眼。用描边代替底板 —— 既能压住复杂背景保证可读，
 * 又不会产生块状遮挡（描边是沿着字形走的，不占额外面积）。
 */
const drawNodeLabel: NodeLabelDrawingFunction = (ctx, data, settings) => {
  const label = data.label
  if (!label) return
  const size = settings.labelSize as number
  // 标签挂在节点上方
  const cx = data.x
  const cy = data.y - (data.size || NODE_RADIUS) - 2

  ctx.font = `${settings.labelWeight} ${size}px ${settings.labelFont}`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.lineWidth = 3
  ctx.lineJoin = 'round'
  ctx.strokeStyle = themeDark ? 'rgba(15,18,26,0.85)' : 'rgba(255,255,255,0.9)'
  ctx.strokeText(label, cx, cy)
  ctx.fillStyle = themeDark ? '#e6edf3' : '#1f2329'
  ctx.fillText(label, cx, cy)
}

// ── 强调集合（与 d3 版一致） ─────────────────────────────
function collectActive(): Set<number> | null {
  if (snapSelectedId !== null) {
    const active = new Set<number>()
    active.add(snapSelectedId)
    for (const n of adjacency.get(snapSelectedId) || []) active.add(n)
    return active
  }
  if (snapHighlighted.size > 0) return snapHighlighted
  return null
}

function nodeReducer(node: string, data: NodeDisplayData): NodeDisplayData {
  const id = Number(node)
  const d = { ...data }
  const active = collectActive()

  // 连线模式：起点/命中目标用强调色，提示可放置
  if (linkFromKey !== null && node === linkFromKey) {
    return { ...d, color: selectedColor(), size: (d.size || 6) * 1.3, zIndex: 4 }
  }
  if (linkTargetKey !== null && node === linkTargetKey && node !== linkFromKey) {
    return { ...d, color: hoverColor(), size: (d.size || 6) * 1.3, zIndex: 4 }
  }

  if (active !== null && !active.has(id)) {
    // 压暗：缩小、置灰、隐藏标签
    return { ...d, color: nodeDimColor(), size: Math.max(3, (d.size || 6) * 0.55), label: '', zIndex: 1 }
  }
  if (hoveredNodeId === node) {
    d.size = (d.size || 6) * 1.5
    d.color = hoverColor()
    d.zIndex = 3
  }
  if (snapSelectedId !== null && snapSelectedId === id) {
    d.size = (d.size || 6) * 1.25
    d.color = selectedColor()
    d.zIndex = 2
  } else if (snapHighlighted.has(id)) {
    d.color = hoverColor()
    d.zIndex = 2
  }
  return d
}

/**
 * 边渲染程序名。
 *
 * ⚠ 直线边**必须**用 sigma 自带的 EdgeLineProgram / EdgeArrowProgram，
 * 绝不能用 @sigma/edge-curve 的曲线程序 + `curvature: 0` 去"画直线"：
 * 那个程序的顶点 shader 里有一句 `float boundingBoxThickness = len * a_curvature;`，
 * 曲率归零时它跟着归零、条带几何塌缩，最终只剩 minEdgeThickness 那么一丁点宽 ——
 * 渲染出来的效果就是"**所有直线关系全部消失，只剩几条弯曲的边**"（踩过一次，很隐蔽）。
 */
const EDGE_TYPE_LINE = 'line'
const EDGE_TYPE_ARROW = 'arrow'
const EDGE_TYPE_CURVE = 'curve'
const EDGE_TYPE_CURVED_ARROW = 'curvedArrow'

/** 把边的 type 升级成"带箭头"的对应版本（直线→直线箭头，曲线→曲线箭头） */
function toArrowType(type: string): string {
  return type === EDGE_TYPE_CURVE || type === EDGE_TYPE_CURVED_ARROW
    ? EDGE_TYPE_CURVED_ARROW
    : EDGE_TYPE_ARROW
}

/**
 * 边渲染的取舍：
 * - 默认的**无箭头**是常态（Obsidian 就不画箭头，满屏箭头是最大的方向噪声源）
 * - **双向关系对**建边时就用带箭头版（两条重叠直线 → 双箭头，能看出互指）
 * - 选中某个节点时，把它相连的边强制切到带箭头版，便于看清方向
 *
 * sigma 是按 display data 的 type 选程序的（见 sigma.esm.js 的 edgesPerPrograms 统计），
 * 所以这里改 type 即可生效，不需要去遍历改图属性。
 */
function edgeReducer(edge: string, data: EdgeDisplayData): EdgeDisplayData {
  const g = graph
  const s = Number(g.source(edge))
  const t = Number(g.target(edge))
  const active = collectActive()

  if (active !== null && (!active.has(s) || !active.has(t))) {
    // 压暗：变淡变细并隐藏标签。type 保持原样，不额外处理
    return { ...data, color: edgeDimColor(), size: 0.4, label: '', zIndex: 1 }
  }
  if (snapSelectedId !== null && (s === snapSelectedId || t === snapSelectedId)) {
    // 强调：染强调色（橙）、明显加粗，并确保显示方向箭头（曲线边保持曲线）
    return { ...data, type: toArrowType(data.type), color: selectedColor(), size: 3.6, zIndex: 2 }
  }
  if (props.showArrows) {
    // 全局方向开关打开：所有边都带箭头。
    // 单向关系（只存在 A→B 一条）平时看不出指向哪边，开了它就能看清。
    const arrowType = toArrowType(data.type)
    return data.type === arrowType ? data : { ...data, type: arrowType }
  }
  // 其余边直接复用原对象：type 由建图时决定（双向对=带箭头、普通=无箭头）。
  // ⚠ 这里绝不能改写 type —— 会把双向关系的箭头一起抹掉。
  return data
}

// ── 建图 ────────────────────────────────────────────────
/**
 * 节点统一大小（不再按度数缩放，避免枢纽节点过大破坏比例）。
 * 与 useForceLayout 的 NODE_RADIUS 同源：sigma 的 size 是**半径**，
 * 而 linkDistance / collide 等力参数都要按「半径的倍数」定标，两边必须一致。
 */
const NODE_SIZE = NODE_RADIUS
/**
 * 边的粗细（图坐标单位）。
 *
 * 为什么是 2 而不是 1：sigma 的 `itemSizesReference` 默认是 'positions'，
 * 边宽 = size × graphToViewportRatio（见 sigma 的 scaleSize()）。图跨度约 1000 个坐标单位
 * 映射到 800px 画布时这个比例只有 0.8 —— 于是 size 1.0 实际只有 **0.8px** 屏幕宽度，
 * 视觉上就是"关系线太细、几乎看不见"。2.0 换算后约 1.6px，与 Obsidian 的观感相当。
 */
const EDGE_SIZE = 2
/** 人工连线的边：在普通边上再加重一档，保持原有的区分度 */
const EDGE_SIZE_MANUAL = 2.6
/**
 * 单边曲率取 0（即直线）。
 *
 * 原先是 0.2 的统一弯曲，但密集区里所有边朝同一侧弯会互相缠绕、反而更乱；
 * Obsidian 的边就是直的 —— 只有同一对节点的多条平行边才需要用弧度区分开。
 */
const EDGE_CURVATURE = 0
/** 同一对节点的多条平行边，按索引扇形展开的步长 */
const PARALLEL_CURVE_STEP = 0.24

/**
 * 计算每条边的 curvature：
 * - 单边：直线（curvature 0）
 * - 平行边：以 0 为基准按索引对称展开（_i - (n-1)/2），左右交替形成扇形
 */
function computeCurvature(groupSize: number, index: number): number {
  if (groupSize <= 1) return EDGE_CURVATURE
  return (index - (groupSize - 1) / 2) * PARALLEL_CURVE_STEP
}

/**
 * 一对节点之间**恰好两条、且互为反向**（A→B 与 B→A）。
 *
 * 这种用两条完全重叠的直线表示：各自的箭头画在自己的终点，
 * 视觉上合起来就是一条**双箭头直线** —— 既干净，又能一眼看出两边互相指向。
 *
 * 其余情况仍走扇形弯曲：如果同向有多条（例如老刘→陈奕 有「追随者」「反目」两条），
 * 画直线会完全重叠、把关系条数直接丢掉。
 */
function isPureBidirectional(group: GraphRelation[]): boolean {
  if (group.length !== 2) return false
  const [first, second] = group
  return first.source_id === second.target_id && first.target_id === second.source_id
}

function rebuild() {
  // 先停掉正在跑的模拟：它的 tick 回调会往图上写坐标，
  // 而接下来这一小段要 clear 并重建整张图（虽然 JS 单线程保证同步块内不会插入 tick，
  // 但显式停掉更稳妥，也避免重建期间白白空转）
  forceSimulator?.stop()
  graph.clear()
  const index = buildGraphIndex(props.entities, props.relations)
  adjacency = index.adjacency

  if (props.entities.length === 0) {
    sigma?.refresh()
    return
  }

  const positions = computeLayout({
    entities: props.entities,
    relations: props.relations,
    width: containerRef.value?.clientWidth || 800,
    height: containerRef.value?.clientHeight || 600
  })

  const degree = index.degree
  for (const e of props.entities) {
    const p = positions.get(e.id) || { x: 0, y: 0 }
    const deg = degree.get(e.id) || 0
    graph.mergeNode(String(e.id), {
      x: p.x,
      y: p.y,
      size: NODE_SIZE,
      label: e.name,
      color: typeColor(e.type),
      // @sigma/node-border：带描边的圆形（type 'border' 匹配 nodeProgramClasses.border）
      type: 'border',
      borderColor: nodeBorderColor(),
      entityType: e.type,
      description: e.description || '',
      degree: deg
    })
  }

  const groups = groupRelationsByPair(props.relations)
  const cursor = new Map<string, number>()
  for (const rel of props.relations) {
    if (!graph.hasNode(String(rel.source_id)) || !graph.hasNode(String(rel.target_id))) continue
    const key = pairKey(rel.source_id, rel.target_id)
    const group = groups.get(key) ?? [rel]
    const groupSize = group.length
    const i = cursor.get(key) || 0
    cursor.set(key, i + 1)

    // 互为反向的两条 → 直线 + 箭头（重叠后呈双箭头）
    const bidirectional = isPureBidirectional(group)
    // 只有同一对节点的**多条平行边**才需要弯曲分开；单条和多条互为反向的走直线
    const parallel = groupSize > 1 && !bidirectional

    graph.mergeEdgeWithKey(String(rel.id), String(rel.source_id), String(rel.target_id), {
      label: rel.relation_type || '',
      // 直线边用 EdgeLineProgram / EdgeArrowProgram 渲染，
      // 不能靠 curvature 0 让曲线程序画直线（原因见上方 EDGE_TYPE_* 的说明）
      type: bidirectional ? EDGE_TYPE_ARROW : parallel ? EDGE_TYPE_CURVE : EDGE_TYPE_LINE,
      curvature: parallel ? computeCurvature(groupSize, i) : 0,
      color: rel.evidence === MANUAL_EVIDENCE ? manualEdgeColor() : edgeColor(),
      size: rel.evidence === MANUAL_EVIDENCE ? EDGE_SIZE_MANUAL : EDGE_SIZE,
      manual: rel.evidence === MANUAL_EVIDENCE
    })
  }

  // 位置从此交给 d3-force：上面的坐标只是种子
  startForceSimulation()
  // 立即取景一次：autoRescale 会把种子范围映射满视野，ratio 1 就是"看全图"
  sigma?.refresh()
  fitView()
}

// ── 力导向布局（d3-force） ──────────────────────────────
let forceSimulator: ForceSimulator | null = null

/** 每跑一帧就把坐标刷到画布；schedule 让同一帧内的多次请求合并成一次渲染 */
function onForceTick() {
  sigma?.refresh({ schedule: true })
}

/**
 * 收敛后取景。
 *
 * ⚠ 这里**不能**用 setCustomBBox(getBBox()) 去"冻结包围盒"（我试过，是个坑）：
 * getBBox() 返回的 nodeExtent 会被 clearNodeIndices() 重置成 {x:[0,1], y:[0,1]}，
 * 要等下一次 process 才重算；而 onTick 用的 refresh({schedule:true}) 是**推迟到下一帧**
 * 才 process 的，d3 的 'end' 事件却在最后一帧 tick 之后同步触发 ——
 * 正好落进这个"已清空、未重算"的窗口，于是把 [0,1] 当全图范围冻住。
 * 而节点实际坐标在 ±600 量级，整张图被映射到视野之外，
 * 表现为「切换图谱几秒后节点和连线一起消失」。
 *
 * 交给 autoRescale 自己算反而是稳的：模拟期间它让图始终填满视野，收敛后也是。
 */
function onForceSettled() {
  fitView()
}

/** 重建并启动模拟。alpha 传小值可用于"重新加热但不打散结构" */
function startForceSimulation(alpha = 1) {
  if (!forceSimulator) {
    forceSimulator = new ForceSimulator(graph, {
      params: props.forceParams,
      onTick: onForceTick,
      onSettled: onForceSettled
    })
  }
  forceSimulator.build()
  forceSimulator.start(alpha)
}

// ── 相机操作（对外接口保持与 d3 版一致） ─────────────────
// 注意：sigma 默认 autoRescale 会把图坐标归一化到 [0,1] 空间，
// 相机的 x/y/ratio 都基于该归一化空间，绝不能用原始图坐标设置相机
function fitView() {
  const s = sigma
  if (!s || graph.order === 0) return
  // 归一化空间下 (0.5, 0.5) + ratio 1 即完整展示全图（含边距）
  s.getCamera().setState({ x: 0.5, y: 0.5, ratio: 1, angle: 0 })
}

function zoomBy(factor: number) {
  sigma?.getCamera().animatedZoom({ factor, duration: 200 })
}

/**
 * 重置布局：重新播种（黄金角螺旋）并重跑模拟。
 * 等价于 Obsidian 的「重置布局」——不是把视图拉回原处，而是让力重新收敛一次。
 */
function resetView() {
  if (graph.order === 0) return
  rebuild()
}

function focusNode(id: number) {
  const s = sigma
  if (!s) return
  const key = String(id)
  if (!graph.hasNode(key)) return
  // 相机坐标必须用归一化后的显示坐标（getNodeDisplayData），不能用原始图坐标
  const display = s.getNodeDisplayData(key)
  if (!display) return
  const camera = s.getCamera()
  camera.setState({ x: display.x, y: display.y, ratio: camera.ratio / 1.5, angle: 0 })
}

// ── 连线模式实现 ────────────────────────────────────────
/** 把 overlay canvas 对齐到容器尺寸（含 DPR） */
function resizeLinkCanvas() {
  const canvas = linkCanvasRef.value
  const container = containerRef.value
  if (!canvas || !container) return
  const dpr = window.devicePixelRatio || 1
  const w = container.clientWidth
  const h = container.clientHeight
  canvas.width = Math.round(w * dpr)
  canvas.height = Math.round(h * dpr)
  canvas.style.width = `${w}px`
  canvas.style.height = `${h}px`
}

/** 绘制橡皮筋：起点固定为 linkFromKey 节点，终点跟随光标 */
function drawLinkOverlay() {
  const canvas = linkCanvasRef.value
  const s = sigma
  if (!canvas || !s || linkFromKey === null || !graph.hasNode(linkFromKey)) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const dpr = window.devicePixelRatio || 1
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  ctx.clearRect(0, 0, canvas.width, canvas.height)

  const display = s.getNodeDisplayData(linkFromKey)
  if (!display) return
  // display 坐标在归一化（framed）空间，必须用 framedGraphToViewport 换算
  const start = s.framedGraphToViewport({ x: display.x, y: display.y })
  const end = linkCursor

  const valid = linkTargetKey !== null && linkTargetKey !== linkFromKey
  ctx.strokeStyle = valid ? hoverColor() : themeDark ? '#8b98a9' : '#9aa4b2'
  ctx.lineWidth = valid ? 2.4 : 1.6
  ctx.setLineDash(valid ? [] : [6, 4])
  ctx.beginPath()
  ctx.moveTo(start.x, start.y)
  ctx.lineTo(end.x, end.y)
  ctx.stroke()
  ctx.setLineDash([])

  // 端点圆点
  ctx.fillStyle = ctx.strokeStyle
  ctx.beginPath()
  ctx.arc(end.x, end.y, valid ? 5 : 3.5, 0, Math.PI * 2)
  ctx.fill()
}

function onLinkMove(e: { x: number; y: number }) {
  if (linkFromKey === null) return
  linkCursor = { x: e.x, y: e.y }
  // 目标节点由 enterNode/leaveNode 维护，这里只跟随光标重画橡皮筋
  drawLinkOverlay()
}

function startLinkingFrom(key: string, x: number, y: number) {
  linkFromKey = key
  linkTargetKey = null
  linkCursor = { x, y }
  linking.value = true
  resizeLinkCanvas()
  drawLinkOverlay()
}

function finishLinking(key: string | null) {
  const from = linkFromKey
  resetLinking()
  if (from !== null && key !== null && key !== from) {
    emit('link', { sourceId: Number(from), targetId: Number(key) })
  }
}

function resetLinking() {
  linkFromKey = null
  linkTargetKey = null
  linkSticky = false
  linking.value = false
  const canvas = linkCanvasRef.value
  const ctx = canvas?.getContext('2d')
  if (canvas && ctx) ctx.clearRect(0, 0, canvas.width, canvas.height)
}

/** 连线模式开关：关闭时清掉进行中的拉线；Esc 退出 */
watch(
  () => props.linkMode,
  on => {
    if (!on) resetLinking()
  }
)

function onLinkKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && props.linkMode) {
    resetLinking()
    emit('exit-link')
  }
}

defineExpose({
  zoomIn: () => zoomBy(1.3),
  zoomOut: () => zoomBy(1 / 1.3),
  fitView,
  resetView,
  focusNode,
  /** 力度参数热更新：改完立刻重新加热模拟，几百毫秒内就能看到形态变化 */
  setForceParams: (params: Partial<ForceParams>) => {
    forceSimulator?.setParams(params)
  },
  /** 读当前力度参数（参数面板初始化用） */
  getForceParams: (): ForceParams => forceSimulator?.getParams() ?? { ...DEFAULT_FORCE_PARAMS }
})

// ── 强调同步 ────────────────────────────────────────────
function syncEmphasis() {
  snapSelectedId = props.selectedId
  snapHighlighted = new Set(props.highlightedIds || [])
  sigma?.refresh()
}

// ── 生命周期 ────────────────────────────────────────────
let themeObserver: MutationObserver | null = null
let resizeObserver: ResizeObserver | null = null

/**
 * 创建 Sigma 并绑定画布事件。容器尚未完成布局（宽/高为 0）时跳过，
 * 由 ResizeObserver 在获得尺寸后重试——destroy-on-close 弹窗（如正文页的图谱按钮）
 * 首次挂载即此场景，直接构造 Sigma 会抛 "Container has no width" 并中断弹窗渲染。
 */
/** 悬停事件的统一出口：把容器内坐标换算成视口坐标一并带出去，供弹窗避让使用 */
function emitHover(id: number | null, coords: { x: number; y: number }) {
  const rect = containerRef.value?.getBoundingClientRect()
  emit('hover-entity', {
    id,
    clientX: rect ? rect.left + coords.x : coords.x,
    clientY: rect ? rect.top + coords.y : coords.y
  })
}

function initSigma() {
  if (sigma) return
  const container = containerRef.value
  if (!container || container.offsetWidth === 0 || container.offsetHeight === 0) return

  sigma = new Sigma(graph, container, {
    renderEdgeLabels: props.showEdgeLabels,
    edgeLabelSize: 10,
    labelSize: 13,
    labelColor: { color: labelColor() },
    edgeLabelColor: { color: edgeLabelColor() },
    nodeReducer,
    edgeReducer,
    minEdgeThickness: 0.3,
    /**
     * 标签只在放大到一定程度后才绘制。
     * 关键：NODE_SIZE 是 7，而 sigma 默认阈值是 6 —— 不改的话所有节点标签会**全量渲染**，
     * 低缩放下就是一片糊。11 相当于"节点显示半径超过 11px 才显示名字"。
     */
    labelRenderedSizeThreshold: 11,
    // 弹窗开场动画期间容器可能瞬间测得 0 尺寸：不抛错，先按 1px 构造，
    // 之后由 ResizeObserver 调 resize() 恢复到真实尺寸
    allowInvalidContainer: true,
    defaultNodeColor: typeColor('character'),
    defaultNodeType: 'border',
    /**
     * ⚠ 必须显式指定。sigma 的 defaultEdgeType 默认是 'line'，
     * 而 'line' 并没有注册进下面的 edgeProgramClasses（只注册了 curve / curvedArrow）。
     * 任何"没有 type 属性的边"都会因此落到 'line' 上，sigma 找不到对应 program 就会抛错
     * 并中断边渲染 —— 表现为只看得见节点、看不见任何连线。
     */
    defaultEdgeType: EDGE_TYPE_LINE,
    // 节点：带描边的圆；边：直线 / 曲线各两套（无箭头 + 带箭头）
    nodeProgramClasses: { border: NodeBorderProgram },
    edgeProgramClasses: {
      [EDGE_TYPE_LINE]: EdgeLineProgram,
      [EDGE_TYPE_ARROW]: EdgeArrowProgram,
      [EDGE_TYPE_CURVE]: EdgeCurveProgram,
      [EDGE_TYPE_CURVED_ARROW]: EdgeCurvedArrowProgram
    },
    // 裸文字标签（无底板）
    defaultDrawNodeLabel: drawNodeLabel
  })

  // 事件
  sigma.on('clickNode', e => {
    e.event.preventSigmaDefault()
    // 连线模式下点击不切换选中，避免和拉线互相干扰
    if (props.linkMode) return
    const id = Number(e.node)
    emit('select', props.selectedId === id ? null : id)
  })
  sigma.on('clickStage', () => {
    if (props.linkMode) return
    emit('select', null)
  })

  // ── 连线模式：按下起拖 / 移动橡皮筋 / 松开落点 ──
  sigma.on('downNode', e => {
    if (!props.linkMode) return
    e.event.preventSigmaDefault()
    if (linking.value && linkSticky) {
      // 点选模式：第二次按下在目标节点上即完成连线
      finishLinking(e.node)
      return
    }
    startLinkingFrom(e.node, e.event.x, e.event.y)
  })
  sigma.on('upNode', e => {
    if (!props.linkMode || linkFromKey === null) return
    if (e.node !== linkFromKey) {
      finishLinking(e.node)
    } else {
      // 按下后没拖动就松开：橡皮筋粘住起点，等下一次点击目标
      linkSticky = true
    }
  })
  sigma.on('upStage', () => {
    if (linkFromKey === null) return
    // 松开在空白处：取消本次拉线
    resetLinking()
  })
  sigma.on('moveBody', e => {
    if (linkFromKey === null) return
    // 拉线期间禁用相机平移
    e.preventSigmaDefault()
    onLinkMove({ x: e.event.x, y: e.event.y })
  })

  sigma.on('enterNode', e => {
    hoveredNodeId = e.node
    // 悬停只负责"看详情"：页面层据此开关弹窗（点击走 select，负责高亮连线）
    emitHover(Number(e.node), e.event)
    // 拉线期间：进入新节点即更新橡皮筋目标（sigma 源码确认拖动中仍派发该事件）
    if (linkFromKey !== null && e.node !== linkFromKey) {
      linkTargetKey = e.node
      drawLinkOverlay()
    }
    sigma?.refresh()
  })
  sigma.on('leaveNode', e => {
    hoveredNodeId = null
    emitHover(null, e.event)
    if (linkFromKey !== null && linkTargetKey === e.node) {
      linkTargetKey = null
      drawLinkOverlay()
    }
    sigma?.refresh()
  })

  // ── 节点拖拽 ────────────────────────────────────────────
  // sigma 核心不含节点拖拽（dist 里没有 draggedNode），必须自行实现。三个要点：
  // ① 用 viewportToGraph 把视口像素换算成图坐标（它会处理 autoRescale 的变换）；
  // ② 按下时冻结包围盒，否则改坐标会让 autoRescale 重新归一化 —— 表现为"拖一个节点、整张图跳"；
  // ③ 拖拽期间保持模拟活跃（wake），邻居才能实时跟弹，而不是拖完才动。
  sigma.on('downNode', e => {
    if (props.linkMode) return
    e.event.preventSigmaDefault()
    draggedNode = e.node
    // 刻意不在这里 setCustomBBox 冻结包围盒：冻结要靠 getBBox()，而它在
    // refresh 的"已清空索引、尚未重新 process"窗口里会返回初始值 {x:[0,1],y:[0,1]}，
    // 冻错一次整张图就会跑到视野外（节点和连线同时消失）。
    // autoRescale 自己算的代价只是拖拽时可能有轻微整体缩放，比"图消失"安全得多。
    forceSimulator?.wake(0.3)
  })

  sigma.getMouseCaptor().on('mousemovebody', e => {
    const s = sigma
    if (!s || draggedNode === null) return
    const pos = s.viewportToGraph(e)
    // 图坐标和 d3 的 fx/fy 都要写：前者让渲染立刻跟上，后者阻止模拟再把节点积分走
    graph.setNodeAttribute(draggedNode, 'x', pos.x)
    graph.setNodeAttribute(draggedNode, 'y', pos.y)
    forceSimulator?.pin(draggedNode, pos.x, pos.y)
    s.refresh({ schedule: true })
    // 拖拽期间禁止画布平移
    e.preventSigmaDefault()
    e.original.preventDefault()
  })

  sigma.getMouseCaptor().on('mouseup', () => endDrag())
  // 兜底：拖到画布外再松手时 captor 可能收不到 mouseup，会让节点被永久钉住
  window.addEventListener('mouseup', endDrag)
}

/** 结束拖拽：释放钉住 + 让 alphaTarget 归零，邻居会带着被拖的节点一起回弹收敛 */
function endDrag() {
  if (draggedNode === null) return
  forceSimulator?.releasePin(draggedNode)
  draggedNode = null
  forceSimulator?.sleep()
}

onMounted(() => {
  const container = containerRef.value
  if (!container) return
  themeDark = isDark()

  initSigma()

  window.addEventListener('keydown', onLinkKeydown)

  syncEmphasis()
  rebuild()

  if (typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => {
      // 容器此前为 0 尺寸时 sigma 尚未创建，拿到实际尺寸后先初始化
      if (!sigma) {
        initSigma()
        return
      }
      // resize() 重新测量容器（含 0 尺寸期间以 1px 构造的实例），refresh() 触发重渲染
      sigma.resize()
      sigma.refresh()
      // 拉线进行中容器尺寸变化时，overlay 需要重新对齐
      if (linking.value) {
        resizeLinkCanvas()
        drawLinkOverlay()
      }
    })
    resizeObserver.observe(container)
  }
  if (typeof MutationObserver !== 'undefined') {
    themeObserver = new MutationObserver(() => {
      const dark = isDark()
      if (dark === themeDark) return
      themeDark = dark
      // 边色与节点描边随主题重建（压暗是 reducer 动态行为，不落盘），人工边保持绿色
      graph.forEachEdge(edge =>
        graph.setEdgeAttribute(
          edge,
          'color',
          graph.getEdgeAttribute(edge, 'manual') ? manualEdgeColor() : edgeColor()
        )
      )
      graph.forEachNode(node => graph.setNodeAttribute(node, 'borderColor', nodeBorderColor()))
      sigma?.setSetting('labelColor', { color: labelColor() })
      sigma?.setSetting('edgeLabelColor', { color: edgeLabelColor() })
      sigma?.refresh()
    })
    themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  }
})

onUnmounted(() => {
  window.removeEventListener('keydown', onLinkKeydown)
  window.removeEventListener('mouseup', endDrag)
  // 先停模拟再销毁 sigma：否则 tick 回调会打到已 kill 的实例上
  forceSimulator?.destroy()
  forceSimulator = null
  draggedNode = null
  sigma?.kill()
  sigma = null
  resizeObserver?.disconnect()
  resizeObserver = null
  themeObserver?.disconnect()
  themeObserver = null
})

watch(
  [() => props.entities, () => props.relations],
  () => rebuild(),
  { deep: false }
)
watch(() => props.selectedId, syncEmphasis)
watch(() => props.highlightedIds, syncEmphasis)
watch(
  () => props.showEdgeLabels,
  v => sigma?.setSetting('renderEdgeLabels', v)
)
// 方向开关只改渲染程序（走 edgeReducer），refresh 一次即可生效
watch(
  () => props.showArrows,
  () => sigma?.refresh()
)
</script>

<style scoped>
.graph-canvas {
  position: relative;
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.sigma-container {
  position: absolute;
  inset: 0;
}

.sigma-container :deep(canvas) {
  outline: none;
}

/* ===== 连线模式 ===== */
/* 橡皮筋预览层：叠在 sigma 画布上，不吃鼠标事件（事件由 sigma 统一处理） */
.graph-link-overlay {
  position: absolute;
  inset: 0;
  z-index: 4;
  pointer-events: none;
}

.is-link-mode .sigma-container {
  cursor: crosshair;
}

/* 顶部提示条 */
.graph-link-hint {
  position: absolute;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 6;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
  border: 1px solid var(--el-color-warning-light-7);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  white-space: nowrap;
}

:root[data-theme='dark'] .graph-link-hint {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
}

/* ===== 空态 ===== */
.graph-canvas-empty {
  position: absolute;
  inset: 0;
  z-index: 3;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--el-bg-color-overlay, #fff);
}

/* ===== 暗色适配 ===== */
:root[data-theme='dark'] .graph-canvas-empty {
  background: rgba(13, 17, 23, 0.85);
}
</style>
