/**
 * 图谱初始坐标（纯函数：输入节点/关系与画布尺寸，输出 id -> {x, y}，不碰 DOM 与模拟器）。
 *
 * 当前挂载的是 computeSeedLayout —— 黄金角螺旋、无类型偏置。
 * 最终位置由 d3-force 连续模拟决定，见 useForceLayout.ts（对标 Obsidian 图谱观感）。
 *
 * computeTypeRingLayout 是上一版「七层嵌套环 + FA2 轻度修饰」的实现，语义分层清晰
 * （角色 → 地点 → 势力 → 事件 → 物件 → 功法 → 线索，空类型跳环），
 * 现已不再挂到 computeLayout 上，但**完整保留**：
 * 需要退回旧观感时，把 computeLayout 改回转调它即可（力模拟需同时停用）。
 */
import type { GraphEntity, GraphEntityType, GraphRelation } from '@/types'
import { buildGraphIndex } from '@/services/knowledgeGraphService'

export interface NodePosition {
  id: number
  x: number
  y: number
}

export interface LayoutContext {
  entities: GraphEntity[]
  relations: GraphRelation[]
  width: number
  height: number
}

/** 画布边缘留白，避免节点贴边 */
const PADDING = 48
/** 节点在环上的最小间距（px），用于按周长自适应半径 */
const MIN_NODE_GAP = 26
/** 相邻环的最小径向间距（px），保证「层」能被肉眼区分 */
const MIN_RING_GAP = 48
/** 最内环基础半径（px） */
const INNER_R = 70

/** 环序：顺序即嵌套层级，从内到外 */
const TYPE_RING_ORDER: readonly GraphEntityType[] = [
  'character',
  'location',
  'faction',
  'event',
  'item',
  'skill',
  'clue'
]

function usableRadius(width: number, height: number): number {
  return Math.max(60, Math.min(width, height) / 2 - PADDING)
}

/** 环内排序：度数降序 → 同度数按名称，保证布局稳定可复现 */
function sortRingNodes(entities: GraphEntity[], degree: Map<number, number>): GraphEntity[] {
  return [...entities].sort((a, b) => {
    const diff = (degree.get(b.id) || 0) - (degree.get(a.id) || 0)
    if (diff !== 0) return diff
    return a.name.localeCompare(b.name, 'zh-Hans-CN')
  })
}

/**
 * 七层嵌套环布局：
 * 1. 按固定类型顺序分环，空类型跳过（层数收缩，仍是嵌套结构）；
 * 2. 自适应半径：环半径不小于「该环节点按最小间距排满一圈」所需，也不小于外环半径 + 最小环间距；
 *    整体超出画布可用半径时等比收缩（保持嵌套，只是更紧凑）；
 * 3. 环内节点沿环均布，相邻环错开半个角步长（蜂窝式错位）。
 */
export function computeTypeRingLayout(ctx: LayoutContext): Map<number, NodePosition> {
  const positions = new Map<number, NodePosition>()
  const index = buildGraphIndex(ctx.entities, ctx.relations)

  // 分环：空类型跳过不占环；类型不在 7 类内的历史数据兜底挂最外环
  const rings: GraphEntity[][] = []
  for (const type of TYPE_RING_ORDER) {
    const group = ctx.entities.filter(e => e.type === type)
    if (group.length > 0) rings.push(sortRingNodes(group, index.degree))
  }
  const knownTypes = new Set<string>(TYPE_RING_ORDER)
  const others = ctx.entities.filter(e => !knownTypes.has(e.type))
  if (others.length > 0) rings.push(sortRingNodes(others, index.degree))

  if (rings.length === 0) return positions

  const cx = ctx.width / 2
  const cy = ctx.height / 2
  const maxR = usableRadius(ctx.width, ctx.height)

  // 定半径：从内到外逐环分配
  const radii: number[] = []
  let prev = INNER_R
  for (const ring of rings) {
    const need = (ring.length * MIN_NODE_GAP) / (2 * Math.PI)
    const r = Math.max(need, prev + MIN_RING_GAP)
    radii.push(r)
    prev = r
  }
  const outer = radii[radii.length - 1]
  const scale = outer > maxR ? maxR / outer : 1

  rings.forEach((ring, k) => {
    const r = radii[k] * scale
    const n = ring.length
    const step = (2 * Math.PI) / n
    // 相邻环错开半个角步长
    const offset = k % 2 === 1 ? step / 2 : 0
    ring.forEach((entity, i) => {
      const angle = step * i + offset - Math.PI / 2
      positions.set(entity.id, {
        id: entity.id,
        x: cx + r * Math.cos(angle),
        y: cy + r * Math.sin(angle)
      })
    })
  })

  return positions
}

/**
 * 黄金角螺旋种子（Vogel 模型）：面积均分 → 半径按 sqrt 递增，配合黄金角
 * 得到圆盘内近似均匀、且两两不重合的分布。
 *
 * 为什么是螺旋而不是类型环：力导向的最终位置由力平衡决定，种子只负责两件事 ——
 * ① 别让节点从同一个坐标出发（重合会让 d3 算出 NaN）；
 * ② 保证可复现。按 id 升序 + 无类型参与，就不会把类型结构偷偷带进收敛结果里。
 * 枢纽节点也不再特意居中，避免引入"中心—边缘"的结构性偏置。
 *
 * 坐标以**原点为中心**输出（forceX(0)/forceY(0) 直接就是"拉向中心"）。
 */
export function computeSeedLayout(ctx: LayoutContext): Map<number, NodePosition> {
  const positions = new Map<number, NodePosition>()
  const total = ctx.entities.length
  if (total === 0) return positions

  const radius = usableRadius(ctx.width, ctx.height)
  const goldenAngle = Math.PI * (3 - Math.sqrt(5))
  const ordered = [...ctx.entities].sort((a, b) => a.id - b.id)

  ordered.forEach((entity, i) => {
    const r = radius * Math.sqrt((i + 0.5) / total)
    const angle = i * goldenAngle
    positions.set(entity.id, {
      id: entity.id,
      x: r * Math.cos(angle),
      y: r * Math.sin(angle)
    })
  })

  return positions
}

/**
 * 对外签名不变：GraphCanvas.rebuild() 无需改调用。
 * 换布局只改这一行 —— 改回 computeTypeRingLayout 即恢复旧的嵌套环观感。
 */
export function computeLayout(ctx: LayoutContext): Map<number, NodePosition> {
  return computeSeedLayout(ctx)
}
