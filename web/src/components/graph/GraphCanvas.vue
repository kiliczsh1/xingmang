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
import { EdgeArrowProgram, EdgeLineProgram } from 'sigma/rendering'
import Graph from 'graphology'
import EdgeCurveProgram, { EdgeCurvedArrowProgram } from '@sigma/edge-curve'
import { createNodeBorderProgram } from '@sigma/node-border'
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
import { computeNodeSizes } from '@/utils/graphSize'

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
/** 最近一次指针在容器内的位置，供拖拽结束后补发悬停用 */
let lastPointer = { x: 0, y: 0 }

// 主题（sigma 是 canvas 绘制，颜色必须在 JS 里决定）
let themeDark = false

// ── 配色（随主题） ──────────────────────────────────────
function isDark(): boolean {
  return document.documentElement.getAttribute('data-theme') === 'dark'
}

function edgeColor(): string {
  /**
   * 边是**极淡的灰**，不抢节点的视觉重心。
   * 亮色 #CCC / 暗色 #555 —— 这是参照图里的取值。
   * 靠 minEdgeThickness 兜住可见度（细边不会被压没）。
   */
  return themeDark ? '#555' : '#CCC'
}

function manualEdgeColor(): string {
  return themeDark ? '#34d399' : '#10b981'
}

/**
 * 人工边在「聚焦到无关节点」时的颜色。
 *
 * 人工边是用户亲手连的，不能像 AI 抽取的边那样直接藏掉 ——
 * 藏了会让人以为连线丢了。所以只把它压到接近背景的淡绿：
 * 存在感极低，但一眼扫过去还看得见"这里有一条我连的线"。
 */
function manualEdgeDimColor(): string {
  return themeDark ? 'rgba(52,211,153,0.45)' : 'rgba(16,185,129,0.42)'
}

function edgeDimColor(): string {
  return themeDark ? 'rgba(255,255,255,0.08)' : 'rgba(0,0,0,0.08)'
}

/**
 * 节点描边色 = **白色**（不是主色）。
 *
 * 参照图里每个节点都是"主色圆 + 白圈"的贴纸效果：
 * 白圈把节点从背景上"撕"出来，密集区里也不会和邻居糊成一片。
 * 之前用同色描边，等于没有描边 —— 节点看着是平的、边界含糊。
 */
function nodeBorderColor(_nodeColor: string): string {
  return themeDark ? 'rgba(18,22,30,0.95)' : '#ffffff'
}

/**
 * 节点标签绘制：**带半透明底的药丸**（不是裸文字）。
 *
 * 裸文字在密集区会和边、别的标签糊在一起，读不出来；
 * 加一层半透明底就相当于给每个名字垫了张纸，任何背景上都清晰。
 * 参照图用的就是这个做法。
 *
 * 位置在节点**右侧**垂直居中（与 sigma 默认一致），
 * 换成自定义绘制器不会让标签跳位。
 */
function drawPillLabel(
  ctx: CanvasRenderingContext2D,
  data: { x: number; y: number; size: number; label?: string | null },
  settings: { labelSize: number; labelFont: string; labelWeight: string },
  /** 悬停/选中节点给完整名字（只有它一个标签，不怕叠） */
  truncate = true
): void {
  const raw = data.label
  if (!raw) return

  const size = settings.labelSize
  ctx.font = `${settings.labelWeight} ${size}px ${settings.labelFont}`

  /**
   * 超长名字截断到 8 个字。
   *
   * sigma 的网格避让（labelGridCellSize）是按**节点中心**分格的，
   * 不管标签本身多宽 —— 两个相距不到一格的节点都会被画出来，
   * 而它们向右伸的标签就会叠在一起（实测"落难无温迪榜单"压住"失落之地编辑系统"）。
   *
   * 截断是最直接的解法：标签短了，就算位置接近也不会互相压。
   * 完整名字在右侧详情栏里看。
   */
  const MAX_CHARS = 8
  const label = truncate && raw.length > MAX_CHARS ? `${raw.slice(0, MAX_CHARS)}…` : raw

  const textWidth = ctx.measureText(label).width

  const padX = 4
  const padY = 2
  const x = data.x + data.size + 3
  const height = size + padY * 2
  const y = data.y - height / 2
  const w = textWidth + padX * 2

  // 底板：亮色白、暗色深，都带透明度 —— 别像贴了张不干胶
  ctx.fillStyle = themeDark ? 'rgba(22,27,37,0.86)' : 'rgba(255,255,255,0.88)'
  if (typeof ctx.roundRect === 'function') {
    ctx.beginPath()
    ctx.roundRect(x - padX, y, w, height, 3)
    ctx.fill()
  } else {
    ctx.fillRect(x - padX, y, w, height)
  }

  ctx.fillStyle = themeDark ? '#e6edf3' : '#1f2329'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  ctx.fillText(label, x, data.y)
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
 * 选中/悬停节点的标签由 sigma 画在 hovers 层，用的**不是**上面的 defaultDrawNodeLabel，
 * 而是 defaultDrawNodeHover。sigma 3.0.3 的默认实现 `drawDiscNodeHover` 把底板写死成
 * `fillStyle = '#FFF'`（settings.labelBackgroundColor 根本没被读取），
 * 所以暗色下选中节点会露出一块白板 —— 这里直接换成同一套胶囊。
 */
function drawHoverPillLabel(
  ctx: CanvasRenderingContext2D,
  data: { x: number; y: number; size: number; label?: string | null },
  settings: { labelSize: number; labelFont: string; labelWeight: string }
): void {
  drawPillLabel(ctx, data, settings, false)
}

function nodeReducer(node: string, data: NodeDisplayData): NodeAttrs {
  // 显式标成 NodeAttrs：下面要写 borderColor，展开后的推断类型不含这个字段
  const d: NodeAttrs = { ...data }
  const id = Number(node)
  const sel = snapSelectedId

  // ── 选中态：「清场」──
  //
  // 无关节点只把**填充**灰化成 #E2E2E2，描边保留主色 →
  // 视觉上是"空心彩边"（和参照图一致）。
  // 注意不要顺手把描边也灰掉，那样节点会变成一坨没有轮廓的灰饼。
  if (sel !== null && graph.hasNode(String(sel))) {
    const selKey = String(sel)
    if (node === selKey) {
      /**
       * 只有**选中节点**标 highlighted。
       *
       * sigma 会把 highlighted 节点重画到 hovers canvas，而 hovers 的层序
       * 在 labels **之上** → 标得越多，越多节点被提到最上层，把旁边的标签盖住。
       * 邻居只保持彩色（不灰化）就已经能区分出来，不需要 highlighted。
       */
      d.highlighted = true
      d.borderColor = MALING_ACCENT
      return d
    }
    if (!graph.neighbors(selKey).includes(node)) {
      /**
       * 无关节点：只灰化，**不用 forceLabel**。
       *
       * forceLabel 会绕过 labelGridCellSize 的网格避让 —— 一旦给所有节点都标上，
       * 密集区的名字就会直接叠在一起（实测"落难无温迪榜单"和"失落之地编辑系统"完全重合）。
       * 让它们走正常的标签通道，由网格和阈值决定显示哪些，才不会叠。
       */
      d.color = dimFill(data.color)
      return d
    }
    /**
     * 邻居：保持彩色 + **强制显示标签**。
     *
     * 为什么用 forceLabel 而不是 highlighted：
     *   - highlighted 会把节点重画到 hovers canvas（层序在 labels 之上），
     *     邻居一多就会把旁边的标签盖掉；
     *   - forceLabel 只是让该节点跳过 `labelRenderedSizeThreshold` 的尺寸检查，
     *     并且会被并入网格避让的结果里，不动层序、不遮挡。
     *
     * 选中一个节点时，用户最想知道的就是"它连着的都是谁"，所以这批名字必须出来。
     */
    d.forceLabel = true
    return d
  }

  // ── 搜索命中高亮 ──
  if (snapHighlighted.has(id)) {
    d.color = MALING_ACCENT
    d.zIndex = 2
    return d
  }

  // ── 悬停：只放大，不改色（颜色留给选中态用）──
  if (hoveredNodeId === node) {
    d.size = (d.size || 6) * 1.4
    d.zIndex = 3
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
function edgeReducer(edge: string, data: EdgeAttrs): EdgeAttrs {
  const g = graph
  const s = Number(g.source(edge))
  const t = Number(g.target(edge))
  const sel = snapSelectedId

  /**
   * ── 默认态：不画任何连线 ──
   *
   * 未选中时整张图只有节点，没有线 —— 画面立刻干净，
   * 读者先看清"有谁"，点某个节点时才展开"它和谁有关系"。
   *
   * 两个例外（此时必须画线，否则开关等于失效）：
   *   - 打开「显示方向箭头」：用户就是想看箭头指向
   *   - 打开「显示关系标签」：标签挂在线上，线不画就没有意义
   */
  const edgesRequested = props.showArrows || props.showEdgeLabels
  if (sel === null && snapHighlighted.size === 0 && !edgesRequested) {
    return { ...data, hidden: true }
  }

  // ── 选中态：「清场」──
  //
  // 无关边 **hidden = true**（整个消失），不是"变淡"。
  // 选中一个节点后画面里只剩它和它的邻居，其余整个退场 ——
  // 这比"压暗"干净得多，也是参照图观感的来源。
  if (sel !== null && g.hasNode(String(sel))) {
    const ext = g.extremities(edge)
    if (ext.includes(String(sel))) {
      return { ...data, type: toArrowType(data.type), color: MALING_ACCENT, size: 2.6, zIndex: 2 }
    }
    // 人工边永不隐藏，只压到很淡的绿
    if (data.manual) {
      return { ...data, color: manualEdgeDimColor(), size: 1, zIndex: 1 }
    }
    return { ...data, hidden: true }
  }

  // ── 搜索命中：相关边高亮，其余压暗 ──
  if (snapHighlighted.size > 0) {
    if (snapHighlighted.has(s) || snapHighlighted.has(t)) {
      return { ...data, color: MALING_ACCENT, size: 2.6, zIndex: 2 }
    }
    return { ...data, color: edgeDimColor(), size: 0.4, label: '', zIndex: 1 }
  }

  if (props.showArrows) {
    // 全局方向开关打开：所有边都带箭头。
    // 单向关系平时看不出指向哪边，开了它就能看清。
    const arrowType = toArrowType(data.type)
    return data.type === arrowType ? data : { ...data, type: arrowType }
  }
  return data
}

// ── 建图 ────────────────────────────────────────────────
/**
 * 节点描边宽度（**像素**，不是半径的倍数）。
 *
 * ⚠️ @sigma/node-border 的描边宽度是 **program 级**配置，节点属性里的
 * `borderSize` 根本不会被读取（想按节点调粗细必须自己建 program）。
 * 默认是「半径的 10%」，这里显式改成 1.5px 绝对值。
 */
const NODE_BORDER_PX = 3

/**
 * 节点渲染程序：1.5px 同色描边 + 内部填充。
 *
 * 同色描边（描边色 = 节点主色）而不是背景色描边 —— 描边是"垫底"的留白分隔，
 * 同色描边则让节点看起来更饱满厚重，是参照图里的观感。
 */
const NodeBorderProgramPx = createNodeBorderProgram({
  borders: [
    { size: { value: NODE_BORDER_PX, mode: 'pixels' }, color: { attribute: 'borderColor' } },
    { size: { fill: true }, color: { attribute: 'color' } }
  ]
})

/**
 * 把主色淡化到 ratio 比例（与背景混合），用于"无关节点"的填充。
 *
 * 参照图里无关节点不是统一灰，而是**各自的颜色淡化到约 14%** ——
 * 远看是浅浅一圈底色，近看还能分辨出类型（浅蓝=角色、浅绿=地点…）。
 * 用统一灰会把类型信息全抹掉。
 */
function tintColor(hex: string, ratio = 0.14): string {
  const m = /^#?([0-9a-f]{6})$/i.exec(hex.trim())
  if (!m) return hex
  const n = parseInt(m[1], 16)
  const r = (n >> 16) & 255
  const g = (n >> 8) & 255
  const b = n & 255
  // 暗色主题往深色混，亮色往白混 —— 否则淡色节点在深底上会发光
  const base = themeDark ? 18 : 255
  const mix = (c: number) => Math.round(base + (c - base) * ratio)
  return `rgb(${mix(r)},${mix(g)},${mix(b)})`
}

/** 聚焦时无关节点的填充（淡化的主色，不是统一灰） */
function dimFill(color: string): string {
  return tintColor(color, themeDark ? 0.16 : 0.13)
}
/** 聚焦时的强调色：选中节点描边 + 相关边 */
const MALING_ACCENT = '#F57F17'

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
const EDGE_SIZE_MANUAL = 1.8
/**
 * 单边曲率 0.2 —— 柔和弧线，不是直线。
 *
 * curvature 是**相对边长**的值：长边弯得明显、短边几乎直。
 * 0.2 在 200 节点规模下不会缠绕（0.35 以上才会打结）。
 *
 * ⚠ 绝不能把这里的值设成 0：@sigma/edge-curve 的顶点 shader 里
 * `float boundingBoxThickness = len * a_curvature;`，曲率归零时条带几何塌缩，
 * 最终只剩 minEdgeThickness 那一丁点宽 —— 表现为"所有关系线全部消失"。
 * 要画直线必须换回 EdgeLineProgram，不能靠 curvature 0。
 */
const EDGE_CURVATURE = 0.2
/** 同一对节点的多条平行边，按索引扇形展开的步长 */
const PARALLEL_CURVE_STEP = 0.24

/**
 * 计算每条边的 curvature。
 *
 * 曲率是**有向**的：正负决定弯向哪一侧。若所有边都用同一个正值，
 * 同一个节点发出的边会全部朝同一侧弯、糊成一把"扫把"（实测很明显）。
 *
 * 所以符号按「一对节点的 id 之和」的奇偶交替：
 * 相邻的边弯向相反方向，从同一个节点发散出去时会自然撑成扇形。
 */
function computeCurvature(groupSize: number, index: number, flip: number): number {
  const base = EDGE_CURVATURE
  if (groupSize <= 1) return base * flip
  return base * flip + (index - (groupSize - 1) / 2) * PARALLEL_CURVE_STEP
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
  // 新图重建后重新取景，旧的"用户已调整视角"标记作废
  cameraMovedByUser = false
  // 先停掉正在跑的模拟：它的 tick 回调会往图上写坐标，
  // 而接下来这一小段要 clear 并重建整张图（虽然 JS 单线程保证同步块内不会插入 tick，
  // 但显式停掉更稳妥，也避免重建期间白白空转）
  forceSimulator?.stop()
  draggedNode = null
  graph.clear()
  // 重建必须解冻：拖拽中切换图谱时若留着旧包围盒，新图会按旧范围归一化 ——
  // 整张图被映射到视野外。放在 clear 之后，空图分支也能覆盖到。
  unfreezeBBox()
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
  // 节点半径按度数缩放（√ 曲线，自适应本图最大度数）
  const sizeById = computeNodeSizes(degree)
  for (const e of props.entities) {
    const p = positions.get(e.id) || { x: 0, y: 0 }
    const deg = degree.get(e.id) || 0
    const color = typeColor(e.type)
    graph.mergeNode(String(e.id), {
      x: p.x,
      y: p.y,
      // 节点大小**按度数缩放**：枢纽（主角/核心势力）明显大，
      // 配角小 —— 层次感就来自这里。固定大小会让整张图没有主次。
      size: sizeById.get(e.id) ?? NODE_RADIUS,
      label: e.name,
      color,
      // @sigma/node-border：带描边的圆形（type 'border' 匹配 nodeProgramClasses.border）
      type: 'border',
      // 同色描边：描边色 = 节点主色
      borderColor: nodeBorderColor(color),
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
      // 所有边统一走曲线程序（柔和弧线）。
      // 双向对也走曲线的箭头版：正反两条会弯成对称的两条弧，方向反而更清楚。
      type: bidirectional ? EDGE_TYPE_CURVED_ARROW : EDGE_TYPE_CURVE,
      curvature: computeCurvature(
        groupSize,
        i,
        (rel.source_id + rel.target_id) % 2 === 0 ? 1 : -1
      ),
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
 *
 * ⚠ 但取景**不能无条件做**：收敛发生在加载后 6~9 秒（alphaDecay 0.0228 → 约 300 帧），
 * 而用户往往在这之前就已经滚轮放大看某个簇了。无条件 fitView 会在那一刻把相机
 * 硬拉回 ratio 1，表现为「放大后停留几秒，视图自己重置了」。
 * 因此只有用户没动过相机时才自动取景 —— 一旦他调过视角，收敛就只管布局，不碰相机。
 */
function onForceSettled() {
  if (cameraMovedByUser) return
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

type BBox = { x: [number, number]; y: [number, number] }

/**
 * 用户是否亲手调过视角（滚轮缩放 / 拖动平移 / 工具栏缩放 / 聚焦节点）。
 * 为 true 时，力模拟收敛不再自动 fitView —— 否则加载后 6~9 秒那一下会把用户刚调好的视角抢走。
 */
let cameraMovedByUser = false
function markCameraMoved() {
  cameraMovedByUser = true
}

/**
 * 自己算节点包围盒（与 sigma 内部 graphExtent() 的算法一致：纯节点 x/y 的 min/max）。
 *
 * ⚠ 不能用 sigma.getBBox() 代替：process 一开始会 clearNodeIndices() 把 nodeExtent
 * 重置成初始值 {x:[0,1], y:[0,1]}，而 refresh({schedule:true}) 是把 process 推迟到下一帧的 ——
 * 在「已清空、尚未重算」的窗口里读 getBBox() 拿到的就是那个初始值，
 * 拿去冻结包围盒会把整张图映射到视野之外（节点和连线一起消失）。
 * 直接遍历 graph 不受渲染时序影响。
 */
function computeNodeBBox(): BBox | null {
  if (graph.order === 0) return null
  let xMin = Infinity
  let xMax = -Infinity
  let yMin = Infinity
  let yMax = -Infinity
  graph.forEachNode((_node, attr) => {
    const x = Number(attr.x)
    const y = Number(attr.y)
    if (!Number.isFinite(x) || !Number.isFinite(y)) return
    if (x < xMin) xMin = x
    if (x > xMax) xMax = x
    if (y < yMin) yMin = y
    if (y > yMax) yMax = y
  })
  if (xMin === Infinity) return null
  return { x: [xMin, xMax], y: [yMin, yMax] }
}

/**
 * 拖拽期间冻结包围盒。
 *
 * 为什么必须冻结：开了 autoRescale 后，sigma 每次 process 都会用当前 nodeExtent 重算
 * 归一化映射（见 sigma 源码 normalizationFunction = createNormalizationFunction(customBBox || nodeExtent)）。
 * 拖拽链路是 mousemove → viewportToGraph（用**上一帧**的映射把像素换算成图坐标）
 * → 写入节点坐标 → 下一帧 process 发现 bbox 变大了 → 映射变宽 → 同样的鼠标位移
 * 被换算成更大的图坐标 → 节点又被放到更远 → bbox 再变大 ……
 * 这是**正反馈**，几十帧内就发散：表现为「轻轻一拖就飞出十万八千里，
 * 原来那一大片节点反而被压成一个点」（被拖的节点撑爆了包围盒，其余节点在归一化后
 * 只能挤在 [0,1] 里的一小块）。冻结之后映射恒定，拖多少像素就是多少，1:1 跟手。
 *
 * 取值刻意不加 padding：按下瞬间的视觉零跳变比「留余量」更重要；
 * 而且鼠标始终在视口内，节点不会真正跑出视野。
 */
function freezeBBox() {
  const s = sigma
  if (!s) return
  const bbox = computeNodeBBox()
  if (bbox) s.setCustomBBox(bbox)
}

/** 松手后把包围盒交回 autoRescale 自动计算（相机不动，避免额外跳变） */
function unfreezeBBox() {
  const s = sigma
  if (!s || !s.getCustomBBox()) return
  s.setCustomBBox(null)
  s.refresh()
}

function fitView() {
  const s = sigma
  if (!s || graph.order === 0) return
  // 归一化空间下 (0.5, 0.5) + ratio 1 即完整展示全图（含边距）
  s.getCamera().setState({ x: 0.5, y: 0.5, ratio: 1, angle: 0 })
}

function zoomBy(factor: number) {
  markCameraMoved()
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
  markCameraMoved()
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
  lastPointer = { x: coords.x, y: coords.y }
  /**
   * 拖拽节点期间一律不发悬停。
   * 拖动时指针会横穿一堆节点，逐 hit 地开关详情弹窗会疯狂闪烁，还挡住正在拖的那个节点。
   * （enterNode/leaveNode 也做了同样的拦截，这里兜底防止别的调用点绕过。）
   */
  if (draggedNode !== null) return
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
    labelSize: 12,
    labelColor: { color: labelColor() },
    edgeLabelColor: { color: edgeLabelColor() },
    nodeReducer,
    edgeReducer,
    minEdgeThickness: 1.7,
    /**
     * 标签只在放大到一定程度后才绘制（节点屏幕半径超过该值才显示名字）。
     *
     * 定 7：节点半径范围是 9~18，取 7 意味着**几乎所有节点在默认视图下都有名字**，
     * 同时又保住了网格避让（labelGridCellSize）的防重叠能力 ——
     * 这是"名字多"和"名字不叠"之间唯一能同时成立的做法。
     *
     * 早期给所有节点硬标 forceLabel 能做到"全都有名字"，但那会绕过网格，
     * 密集区直接叠成一团，反而更读不了。
     */
    labelRenderedSizeThreshold: 7,
    /**
     * 标签网格避让：画布按 60×60 的格子切，同一格内只画优先级最高的标签，
     * 其余自动隐藏。密集区不压字的关键，比调字号有效得多。
     */
    labelGridCellSize: 60,
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
    defaultEdgeType: EDGE_TYPE_CURVE,
    // 节点：带描边的圆；边：直线 / 曲线各两套（无箭头 + 带箭头）
    nodeProgramClasses: { border: NodeBorderProgramPx },
    // 标签加半透明底，密集区也能读
    defaultDrawNodeLabel: drawPillLabel as never,
    // 选中/悬停节点走同一套胶囊（默认实现是写死的白底，暗色下会露白板）
    defaultDrawNodeHover: drawHoverPillLabel as never,
    edgeProgramClasses: {
      [EDGE_TYPE_LINE]: EdgeLineProgram,
      [EDGE_TYPE_ARROW]: EdgeArrowProgram,
      [EDGE_TYPE_CURVE]: EdgeCurveProgram,
      [EDGE_TYPE_CURVED_ARROW]: EdgeCurvedArrowProgram
    },
    // 裸文字标签（无底板）
  })

  // 事件
  // 用户亲手改视角（滚轮缩放 / 双击放大）→ 力模拟收敛时不再自动取景
  sigma.on('wheelStage', markCameraMoved)
  sigma.on('doubleClickStage', markCameraMoved)

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
    // 拖拽期间不做悬停切换：路过节点既不弹详情，也不该把高亮从正在拖的节点上抢走
    if (draggedNode !== null) {
      lastPointer = { x: e.event.x, y: e.event.y }
      return
    }
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
    // 同上：拖拽期间保持悬停状态不变（仍是被拖的那个节点）
    if (draggedNode !== null) return
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
  // ② 按下时冻结包围盒（freezeBBox，用自己算的 bbox，不碰 getBBox()），
  //    否则 autoRescale 每帧重算归一化 → 拖得越远跳得越狠，几帧就发散；
  // ③ 拖拽期间保持模拟活跃（wake），邻居才能实时跟弹，而不是拖完才动。
  sigma.on('downNode', e => {
    if (props.linkMode) return
    e.event.preventSigmaDefault()
    // 按下即把详情弹窗收掉（此刻 draggedNode 还是 null，事件能正常发出去）
    emitHover(null, e.event)
    draggedNode = e.node
    freezeBBox()
    forceSimulator?.wake(0.3)
  })

  sigma.getMouseCaptor().on('mousemovebody', e => {
    const s = sigma
    if (!s) return
    // 拖节点或拖画布平移，都算用户自己安排过视角
    markCameraMoved()
    if (draggedNode === null) return
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

/** 结束拖拽：解冻包围盒 + 释放钉住 + 让 alphaTarget 归零，邻居会带着被拖的节点一起回弹收敛 */
function endDrag() {
  if (draggedNode === null) return
  unfreezeBBox()
  forceSimulator?.releasePin(draggedNode)
  draggedNode = null
  forceSimulator?.sleep()
  // 松手后补发一次悬停：指针若正好停在节点上，详情弹窗能立刻回来
  // （拖拽期间被吞掉的 enterNode 不会再补一次，不补的话要挪一下鼠标才恢复）
  if (hoveredNodeId !== null) emitHover(Number(hoveredNodeId), lastPointer)
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
watch(() => props.selectedId, id => {
  syncEmphasis()
  /**
   * 选中后**冻结布局**。
   *
   * 选中态下用户是在"读关系"——节点再多飘一两个像素都会干扰判断，
   * 尤其是聚焦后只剩一个中心节点加一圈邻居，画面必须是静止的。
   *
   * 用 stop() 而不是 sleep()：sleep 只是把 alphaTarget 归零，
   * 模拟仍会继续 tick 到自然冷却；stop 是直接停掉 tick 循环，一步到位。
   *
   * 为什么取消选中后不主动重启：位置早就收敛了，重启只会让刚看过的
   * 结构轻轻抖一下（alpha 从 0 重新升温）—— 那才是真正的干扰。
   * 需要重新布局时用户会点「重置布局」，那条路径会 rebuild + start。
   */
  if (id !== null) forceSimulator?.stop()
})
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
