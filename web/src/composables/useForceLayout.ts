/**
 * d3-force 力导向布局：把 graphology 图变成一张"活的"力模拟（对标 Obsidian 图谱观感）。
 *
 * 参数刻度直接沿用 Obsidian 的四个滑块，不另造一套心智模型：
 *   图谱向心力 centerForce 0–1 · 节点间排斥力 repelForce 0–20
 *   相连节点间吸引力 linkForce 0–1 · 连线长度 linkDistance 30–500
 *
 * ⚠ 量纲问题：sigma 的 autoRescale 会把图坐标归一化到 [0,1]，**绝对像素没有意义**，
 * 守恒的只有比例。所以 linkDistance / collide 等一律按「节点半径的倍数」定义。
 *
 * ⚠ 系数问题：Obsidian 的滑块值不能直接喂给 d3（两者"单位力度"差一个量级）。
 * CENTER_TO_STRENGTH / REPEL_TO_STRENGTH / LINK_DISTANCE_RATIO 都是标定出来的，
 * 改动前请先跑 scripts/calibrate_force_params.mjs 重新标定。
 */
import {
  forceCollide,
  forceLink,
  forceManyBody,
  forceSimulation,
  forceX,
  forceY
} from 'd3-force'
import type { Simulation, SimulationLinkDatum, SimulationNodeDatum } from 'd3-force'

/** 节点半径（图坐标单位），与 GraphCanvas 的 NODE_SIZE 保持一致 */
export const NODE_RADIUS = 7

export interface ForceParams {
  /** 图谱向心力 0–1：软弹簧拉向画布中心，只做兜底，调高会把图压成圆饼 */
  centerForce: number
  /** 节点间的排斥力 0–20：全局两两作用，调高会抹平簇结构 */
  repelForce: number
  /** 相连节点间的吸引力 0–1：按度数归一化，枢纽节点不会被撕碎 */
  linkForce: number
  /** 连线长度 30–500：弹簧自然长度，是留白比例的主因 */
  linkDistance: number
}

/**
 * 连线长度的基准倍数（相对节点半径）。
 *
 * 标定依据（scripts/calibrate_force_params.mjs，样本 book 16 v10 的 60 实体 / 63 关系）：
 *   比值 20 → 边交叉 22 次，gapRatio 4.6
 *   比值 40 → 边交叉 198 次（每条边平均与 3 条边相交，结构明显变乱），gapRatio 4.6
 * 「边交叉数」是清爽与否的直接量度，而 20 倍正好是交叉数开始暴涨前的拐点，
 * 且留白（gapRatio）与 40 倍完全相同 —— 即"白得一样多，但线不打架"。
 *
 * 注意这与 Obsidian 的 250px ÷ 半径 6px ≈ 40 倍不同：Obsidian 的图动辄几百个节点，
 * 40 倍才撑得开；60 节点规模下用 40 倍，边会横穿整张图。
 */
const LINK_DISTANCE_RATIO = 20

/** 出厂默认：前三项沿用 Obsidian 实测默认值，第四项按标定结论换算 */
export const DEFAULT_FORCE_PARAMS: ForceParams = {
  centerForce: 0.52,
  repelForce: 10,
  linkForce: 1,
  linkDistance: NODE_RADIUS * LINK_DISTANCE_RATIO
}

/** 各滑块的取值范围（与 Obsidian 一致，供 UI 复用） */
export const FORCE_PARAM_RANGES = {
  centerForce: { min: 0, max: 1, step: 0.01 },
  repelForce: { min: 0, max: 20, step: 0.1 },
  linkForce: { min: 0, max: 1, step: 0.01 },
  linkDistance: { min: 30, max: 500, step: 5 }
} as const

/**
 * 内部量纲换算：把 Obsidian 刻度映射到 d3 的实际量纲。
 * 标定结论（scripts/calibrate_force_params.mjs 的三张扫描表）：
 *   - centerStrength 0.104（= 0.52 × 0.2）：图半径从无约束的 953 收到 412，间距比 4.6，且没压扁
 *   - charge -300（= 10 × 30）：间距比 4.6（节点空隙 ≈ 3.3 倍节点直径），零重叠
 * 两者都落在"再拧一点就明显变差"之前的舒适区里。
 */
const CENTER_TO_STRENGTH = 0.2
const REPEL_TO_STRENGTH = 30

/** 模拟器只需要 graphology 图的一小部分能力，用结构化接口避免泛型耦合 */
export interface GraphLike {
  readonly order: number
  forEachNode(callback: (id: string) => void): void
  forEachEdge(
    callback: (edge: string, attributes: unknown, source: string, target: string) => void
  ): void
  getNodeAttribute(id: string, name: string): unknown
  setNodeAttribute(id: string, name: string, value: unknown): void
  degree(id: string): number
}

interface SimNode extends SimulationNodeDatum {
  id: string
}

interface SimLink extends SimulationLinkDatum<SimNode> {
  source: string | SimNode
  target: string | SimNode
}

export interface ForceSimulatorOptions {
  params?: Partial<ForceParams>
  /** 每次迭代后回调（用于把坐标刷到渲染层），调用方应做按帧合并 */
  onTick: () => void
  /** 收敛后回调（用于冻结包围盒、适配视图） */
  onSettled?: () => void
}

function asNumber(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) ? value : 0
}

function endpointId(endpoint: string | SimNode): string {
  return typeof endpoint === 'string' ? endpoint : endpoint.id
}

/**
 * 力模拟器。生命周期与 GraphCanvas 的 rebuild 对齐：
 *   build()（读当前坐标建模拟） → start()（跑） → …tick… → onSettled()
 * 拖拽期间用 pin()/releasePin() 钉住节点，用 wake()/sleep() 控制活跃度。
 */
export class ForceSimulator {
  private readonly graph: GraphLike
  private simulation: Simulation<SimNode, SimLink> | null = null
  private nodes: SimNode[] = []
  private links: SimLink[] = []
  private nodeById = new Map<string, SimNode>()
  private params: ForceParams
  private readonly onTick: () => void
  private readonly onSettled: () => void

  // 注意：项目开了 erasableSyntaxOnly，不能写构造函数参数属性（private readonly graph: ...）
  constructor(graph: GraphLike, options: ForceSimulatorOptions) {
    this.graph = graph
    this.params = { ...DEFAULT_FORCE_PARAMS, ...options.params }
    this.onTick = options.onTick
    this.onSettled = options.onSettled ?? (() => {})
  }

  getParams(): ForceParams {
    return { ...this.params }
  }

  /** 参数热更新：不重建图，只换力度并重新加热，让改动立刻可见 */
  setParams(next: Partial<ForceParams>): void {
    this.params = { ...this.params, ...next }
    if (!this.simulation) return
    this.applyForces()
    // 用中等 alpha 重新加热：够快看到变化，又不会把已收敛的结构彻底打散
    this.simulation.alpha(0.5).restart()
  }

  /** 读取图当前坐标作为初值，构建模拟（调用方应先把种子坐标写进图里） */
  build(): void {
    this.destroy()

    const nodes: SimNode[] = []
    const nodeById = new Map<string, SimNode>()

    this.graph.forEachNode(id => {
      const node: SimNode = {
        id,
        x: asNumber(this.graph.getNodeAttribute(id, 'x')),
        y: asNumber(this.graph.getNodeAttribute(id, 'y'))
      }
      nodes.push(node)
      nodeById.set(id, node)
    })

    const links: SimLink[] = []
    this.graph.forEachEdge((_edge, _attributes, source, target) => {
      // 自环对布局没有意义，d3 会算出 NaN
      if (source === target) return
      links.push({ source, target })
    })

    this.nodes = nodes
    this.nodeById = nodeById
    this.links = links

    this.simulation = forceSimulation<SimNode, SimLink>(nodes)
      .force(
        'link',
        forceLink<SimNode, SimLink>(links)
          .id(node => node.id)
          .distance(this.params.linkDistance)
          .strength(link => this.linkStrength(link))
      )
      .force(
        'charge',
        forceManyBody<SimNode>()
          .strength(-this.params.repelForce * REPEL_TO_STRENGTH)
          .theta(0.9)
      )
      .force('x', forceX<SimNode>(0).strength(this.params.centerForce * CENTER_TO_STRENGTH))
      .force('y', forceY<SimNode>(0).strength(this.params.centerForce * CENTER_TO_STRENGTH))
      .force('collide', forceCollide<SimNode>(NODE_RADIUS * 1.8).strength(0.9))
      // 高阻尼 + d3 默认 alpha 衰减：收敛后几乎完全静止，"静"本身就是清爽的一半
      .velocityDecay(0.6)
      .alphaDecay(0.0228)
      .on('tick', () => this.writeBack())
      .on('end', () => this.onSettled())
  }

  /** 与 d3 默认一致：按两端度数的较小值归一化，枢纽节点的边不会合力把它扯碎 */
  private linkStrength(link: SimLink): number {
    const source = endpointId(link.source)
    const target = endpointId(link.target)
    const minDegree = Math.min(this.graph.degree(source), this.graph.degree(target))
    return this.params.linkForce / Math.max(1, minDegree)
  }

  /** 把力度对象的参数换成当前值（热更新用） */
  private applyForces(): void {
    const simulation = this.simulation
    if (!simulation) return

    const link = simulation.force('link') as ReturnType<typeof forceLink<SimNode, SimLink>>
    link.distance(this.params.linkDistance).strength(entry => this.linkStrength(entry))

    const charge = simulation.force('charge') as ReturnType<typeof forceManyBody<SimNode>>
    charge.strength(-this.params.repelForce * REPEL_TO_STRENGTH)

    const strength = this.params.centerForce * CENTER_TO_STRENGTH
    ;(simulation.force('x') as ReturnType<typeof forceX<SimNode>>).strength(strength)
    ;(simulation.force('y') as ReturnType<typeof forceY<SimNode>>).strength(strength)
  }

  private writeBack(): void {
    const graph = this.graph
    const nodes = this.nodes
    for (let i = 0; i < nodes.length; i++) {
      const node = nodes[i]
      graph.setNodeAttribute(node.id, 'x', node.x)
      graph.setNodeAttribute(node.id, 'y', node.y)
    }
    this.onTick()
  }

  start(alpha = 1): void {
    this.simulation?.alpha(alpha).restart()
  }

  stop(): void {
    this.simulation?.stop()
  }

  /** 拖拽期间保持活跃：alphaTarget 不为 0，模拟不会自行冷却到停机 */
  wake(alpha = 0.3): void {
    this.simulation?.alphaTarget(alpha).restart()
  }

  /** 松手后释放 alphaTarget，让模拟自然降温收敛 */
  sleep(): void {
    this.simulation?.alphaTarget(0)
  }

  /** 钉住节点（拖拽中）：fx/fy 一旦非空，d3 就不再积分该节点的位置 */
  pin(id: string, x: number, y: number): void {
    const node = this.nodeById.get(id)
    if (!node) return
    node.fx = x
    node.fy = y
  }

  /** 释放钉住：节点重新参与积分，会被邻居"拉回"平衡位置 */
  releasePin(id: string): void {
    const node = this.nodeById.get(id)
    if (!node) return
    node.fx = null
    node.fy = null
  }

  destroy(): void {
    if (this.simulation) {
      this.simulation.on('tick', null).on('end', null)
      this.simulation.stop()
    }
    this.simulation = null
    this.nodes = []
    this.links = []
    this.nodeById = new Map()
  }
}
