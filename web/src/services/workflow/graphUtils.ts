import type { NodeType, WEdge, WNode } from '@/types/workflow'

/** 连线合法性校验结果 */
export interface ConnectionCheck {
  ok: boolean
  reason?: string
}

/** 邻接表（按边方向 source → target） */
const buildAdjacency = (edges: WEdge[]) => {
  const adj = new Map<string, string[]>()
  for (const e of edges) {
    if (!adj.has(e.source)) adj.set(e.source, [])
    adj.get(e.source)!.push(e.target)
  }
  return adj
}

/**
 * BFS 找 from → to 的一条路径（含首尾）。
 * 只用于连线校验的即时判断，边数量在画布规模下（几十条）没有性能问题。
 */
export const findPath = (edges: WEdge[], from: string, to: string): string[] | null => {
  if (from === to) return [from]
  const adj = buildAdjacency(edges)
  const prev = new Map<string, string | null>([[from, null]])
  const queue: string[] = [from]
  while (queue.length) {
    const cur = queue.shift()!
    for (const next of adj.get(cur) || []) {
      if (prev.has(next)) continue
      prev.set(next, cur)
      if (next === to) {
        const path: string[] = []
        let node: string | null = to
        while (node) {
          path.unshift(node)
          node = prev.get(node) ?? null
        }
        return path
      }
      queue.push(next)
    }
  }
  return null
}

/**
 * 新增 source → target 这条边是否会成环。
 *
 * 判据：若 target 已经能沿边走到 source，则新边会闭环。
 *
 * 循环节点例外：星芒的 loop 靠回边表达（loop-body 子图末尾接回 loop 节点，
 * 见「章节流水线」模板 ctx-1 → loop-1），这在 DAG 意义上就是环，必须放行。
 * 因此只要可达路径上出现过 loop 节点，就认为是合法的循环回边。
 */
export const wouldCreateCycle = (
  edges: WEdge[],
  nodes: WNode[],
  source: string,
  target: string
): boolean => {
  if (source === target) return true
  const path = findPath(edges, target, source)
  if (!path) return false
  const typeOf = new Map<string, NodeType>()
  nodes.forEach(n => typeOf.set(n.id, n.type))
  return !path.some(id => typeOf.get(id) === 'loop')
}

/** 同向、同端口句柄的连线视为重复；ignoreId 用于排除「自己」 */
export const isDuplicateEdge = (
  edges: WEdge[],
  conn: {
    id?: string | null
    source?: string | null
    target?: string | null
    sourceHandle?: string | null
    targetHandle?: string | null
  },
  ignoreId?: string | null
): boolean => {
  // 只在明确给了 id 时才做「排除自己」；否则没有 id 的边会被一并跳过
  const skipId = ignoreId ?? null
  return edges.some(
    e =>
      (skipId === null || e.id !== skipId) &&
      e.source === conn.source &&
      e.target === conn.target &&
      (e.sourceHandle ?? null) === (conn.sourceHandle ?? null) &&
      (e.targetHandle ?? null) === (conn.targetHandle ?? null)
  )
}

/**
 * 画布连线总校验：自连 / 重复 / 成环。
 *
 * ⚠ Vue Flow 不只是「新建连线时」问这个函数：`v-model:edges` 每次变更都会走
 * `setEdges → createGraphEdges`，把边表里**每一条已有的边**（带 id）再送回来复核，
 * 判 false 就直接丢弃该边（EDGE_INVALID）。所以：
 *
 * - 带 `id` 的连接对象 = 已有边复核 → 直接放行（查重会匹配到它自己）。
 *   新连线在 `onConnect` 里是「先校验、后分配 id」，所以不会误放行真正的重复连线。
 * - 调用方必须把 Vue Flow 给的对象原样透传，别自己挑字段重建，否则 id 会丢。
 */
export const validateConnection = (
  edges: WEdge[],
  nodes: WNode[],
  conn: {
    id?: string | null
    source?: string | null
    target?: string | null
    sourceHandle?: string | null
    targetHandle?: string | null
  }
): ConnectionCheck => {
  if (!conn) return { ok: false, reason: '连接对象为空' }
  if (conn.id) return { ok: true }
  if (!conn.source || !conn.target) return { ok: false, reason: '连接端点不完整' }
  if (conn.source === conn.target) return { ok: false, reason: '不能把节点连到自己' }
  if (isDuplicateEdge(edges, conn)) return { ok: false, reason: '这两个端口已经有连线了' }
  if (wouldCreateCycle(edges, nodes, conn.source, conn.target)) {
    return { ok: false, reason: '这条连线会造成环路（循环回边请通过循环节点连接）' }
  }
  return { ok: true }
}

// ---------------------------------------------------------------- 执行期图算法

/** 反向可达：所有能（沿边）到达 targetId 的节点 id，不含 targetId 自身 */
export const collectAncestorIds = (edges: WEdge[], targetId: string): Set<string> => {
  const reverse = new Map<string, string[]>()
  for (const e of edges) {
    if (!reverse.has(e.target)) reverse.set(e.target, [])
    reverse.get(e.target)!.push(e.source)
  }
  const seen = new Set<string>()
  const queue = [...(reverse.get(targetId) || [])]
  while (queue.length) {
    const cur = queue.shift()!
    if (seen.has(cur) || cur === targetId) continue
    seen.add(cur)
    for (const prev of reverse.get(cur) || []) if (!seen.has(prev)) queue.push(prev)
  }
  return seen
}

/** targetId 自身 + 全部下游 id */
export const collectDescendantIds = (edges: WEdge[], id: string): Set<string> => {
  const adj = buildAdjacency(edges)
  const seen = new Set<string>([id])
  const queue = [id]
  while (queue.length) {
    const cur = queue.shift()!
    for (const next of adj.get(cur) || []) {
      if (seen.has(next)) continue
      seen.add(next)
      queue.push(next)
    }
  }
  return seen
}

/**
 * 能到达 targetId 的节点，按拓扑序返回（上游在前）。
 * 目标节点自身不在结果里。单节点测试运行按这个顺序补齐上游。
 */
export const upstreamInTopoOrder = (nodes: WNode[], edges: WEdge[], targetId: string): WNode[] => {
  const ancestors = collectAncestorIds(edges, targetId)
  if (!ancestors.size) return []
  const indeg = new Map<string, number>()
  ancestors.forEach(id => indeg.set(id, 0))
  const adj = new Map<string, string[]>()
  for (const e of edges) {
    if (!ancestors.has(e.source) || !ancestors.has(e.target)) continue
    indeg.set(e.target, (indeg.get(e.target) || 0) + 1)
    if (!adj.has(e.source)) adj.set(e.source, [])
    adj.get(e.source)!.push(e.target)
  }
  const nodeMap = new Map(nodes.map(n => [n.id, n]))
  const queue = Array.from(ancestors).filter(id => (indeg.get(id) || 0) === 0)
  const out: WNode[] = []
  while (queue.length) {
    const id = queue.shift()!
    const node = nodeMap.get(id)
    if (node) out.push(node)
    for (const next of adj.get(id) || []) {
      const left = (indeg.get(next) || 0) - 1
      indeg.set(next, left)
      if (left === 0) queue.push(next)
    }
  }
  return out
}

/**
 * loop 体内部节点集合（并发调度的排除项）。
 *
 * 循环体由 LoopExecutor 自己沿单主线串行跑完（见 loopExecutor.executeLoopBody），
 * 所以这些节点不能同时交给并发调度器，否则会被执行两次。
 * 判定：从每个 loop 节点的 loop-body 出边出发可达的节点，遇到 loop 节点自身即停（回边）。
 */
export const collectLoopBodyIds = (nodes: WNode[], edges: WEdge[]): Set<string> => {
  const loopIds = nodes.filter(n => n.type === 'loop').map(n => n.id)
  const result = new Set<string>()
  if (!loopIds.length) return result
  const adj = buildAdjacency(edges)
  for (const loopId of loopIds) {
    const seeds = edges.filter(e => e.source === loopId && e.sourceHandle === 'loop-body').map(e => e.target)
    const queue = [...seeds]
    while (queue.length) {
      const cur = queue.shift()!
      if (cur === loopId || result.has(cur)) continue
      result.add(cur)
      for (const next of adj.get(cur) || []) if (next !== loopId && !result.has(next)) queue.push(next)
    }
  }
  return result
}

/** 并发调度用的执行图（loop 体内部节点已被剔除，回边/进入回边的边随之丢弃） */
export interface ExecutionGraph {
  /** 参与调度的节点 id */
  nodeIds: string[]
  /** 入度（只统计调度图内的边） */
  inDegree: Record<string, number>
  /** 后继表 */
  successors: Record<string, string[]>
  /** 被排除的 loop 体内部节点 */
  excluded: Set<string>
}

export const buildExecutionGraph = (nodes: WNode[], edges: WEdge[], excluded: Set<string>): ExecutionGraph => {
  const nodeIds = nodes.filter(n => !excluded.has(n.id)).map(n => n.id)
  const inSet = new Set(nodeIds)
  const inDegree: Record<string, number> = {}
  const successors: Record<string, string[]> = {}
  nodeIds.forEach(id => {
    inDegree[id] = 0
    successors[id] = []
  })
  for (const e of edges) {
    // 只保留两端都在调度图内的边：loop-body 边与回边会因此被丢弃
    if (!inSet.has(e.source) || !inSet.has(e.target)) continue
    if (e.source === e.target) continue
    successors[e.source].push(e.target)
    inDegree[e.target]++
  }
  return { nodeIds, inDegree, successors, excluded }
}

