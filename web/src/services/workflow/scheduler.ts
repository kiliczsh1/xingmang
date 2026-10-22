import { buildExecutionGraph, collectLoopBodyIds } from './graphUtils'
import type { NodeExecuteResult, WNode, WorkflowFile } from '@/types/workflow'

/** 并发度配置 */
export interface ConcurrencyLimits {
  /** 同时在跑的节点总数上限 */
  total: number
  /** 同时在跑的 AI 节点上限（生成类请求最贵，单独限流） */
  ai: number
}

export const DEFAULT_LIMITS: ConcurrencyLimits = { total: 4, ai: 2 }

/** 调度器需要主执行器提供的能力（不直接依赖 WorkflowExecutor，避免循环引用） */
export interface SchedulerHost {
  workflow: WorkflowFile
  store: any
  stopped: boolean
  executeNode(node: WNode, opts?: { preventFollowEdges?: boolean }): Promise<NodeExecuteResult>
  log(message: string, type?: 'info' | 'success' | 'warning' | 'error'): void
}

/**
 * 并发调度器：拓扑入度 + 就绪队列 + 限流。
 *
 * 与顺序引擎（executor 里从 start 沿边递归）并存，由执行模式开关选择。
 *
 * 两条关键约定：
 * 1. loop 体内部节点不参与调度 —— 循环体由 LoopExecutor 自己串行跑（会用到回边，
 *    交给调度器会重复执行）。loop 节点本身只跑循环，其后继（loop-done）正常调度。
 * 2. 节点统一用 preventFollowEdges=true 执行，下游推进完全由调度器负责。
 */
export class ConcurrentScheduler {
  host: SchedulerHost
  limits: ConcurrencyLimits

  constructor(host: SchedulerHost, limits: Partial<ConcurrencyLimits> = {}) {
    this.host = host
    this.limits = {
      total: Math.max(1, limits.total ?? DEFAULT_LIMITS.total),
      ai: Math.max(1, limits.ai ?? DEFAULT_LIMITS.ai)
    }
  }

  async run(): Promise<{ completed: number; total: number }> {
    const { nodes, edges } = this.host.workflow
    if (!nodes.length) throw new Error('工作流没有任何节点')

    const excluded = collectLoopBodyIds(nodes, edges)
    const graph = buildExecutionGraph(nodes, edges, excluded)
    const nodeMap = new Map(nodes.map(n => [n.id, n]))
    const typeOf = (id: string) => nodeMap.get(id)?.type
    const labelOf = (id: string) => nodeMap.get(id)?.data?.label || typeOf(id) || id

    if (!graph.nodeIds.length) throw new Error('并发调度没有可执行的节点')

    if (excluded.size) {
      this.host.log(
        `ℹ️ 并发模式：${excluded.size} 个循环体节点由循环节点内部串行执行，不参与并发调度`,
        'info'
      )
    }

    const pending = new Set(graph.nodeIds.filter(id => (graph.inDegree[id] || 0) === 0))
    const done = new Set<string>()
    const running = new Map<string, Promise<void>>()
    const total = graph.nodeIds.length
    let firstError: any = null

    if (!pending.size) {
      // 全是环（画布已禁止无 loop 成环，这里兜底手写 JSON 的情况）
      throw new Error(
        `工作流存在环路，没有可启动的节点（无入度 0 节点）。涉及节点: ${graph.nodeIds.map(labelOf).join('、')}`
      )
    }

    /** 所有会发起模型请求的节点类型（星域 aiModel 与星芒 aiProcess 同属此类） */
    const AI_NODE_TYPES = new Set(['aiProcess', 'aiModel'])
    const aiRunning = () =>
      Array.from(running.keys()).filter(id => AI_NODE_TYPES.has(typeOf(id) as string)).length

    const startNode = (id: string) => {
      pending.delete(id)
      const node = nodeMap.get(id)!
      const p = (async () => {
        // 必须统一 preventFollowEdges: true —— 下游推进完全由调度器的 successors 负责。
        // 若按isLoop 传值，非 loop 节点会拿到 false，executeNode 内部会沿getNextNodes
        // 递归执行下游，而调度器 .then() 又会把同一批下游加进 pending → 下游被执行两次
        // （AI 节点重复计费、输出类节点重复落盘）。isLoop 变量已无用途。
        const res = await this.host.executeNode(node, { preventFollowEdges: true })
        if (res?.aborted) this.host.stopped = true
      })()
        .then(() => {
          running.delete(id)
          done.add(id)
          this.host.store.executionProgress = Math.round((done.size / total) * 100)
          for (const next of graph.successors[id] || []) {
            const left = (graph.inDegree[next] || 0) - 1
            graph.inDegree[next] = left
            if (left === 0 && !done.has(next)) pending.add(next)
          }
        })
        .catch((e: any) => {
          running.delete(id)
          if (!firstError) firstError = e
          throw e
        })
      running.set(id, p)
    }

    while (done.size < total) {
      if (this.host.stopped) {
        this.host.log(`⏹️ 已停止：完成 ${done.size}/${total} 个节点`, 'warning')
        break
      }
      if (firstError) break

      // 填满并发槽位
      let started = true
      while (started) {
        started = false
        if (running.size >= this.limits.total) break
        for (const id of Array.from(pending)) {
          if (running.size >= this.limits.total) break
          if (AI_NODE_TYPES.has(typeOf(id) as string) && aiRunning() >= this.limits.ai) continue
          startNode(id)
          started = true
        }
      }

      if (!running.size) {
        if (firstError) break
        if (!pending.size) {
          // 没有可跑的、也没有在跑的，但还没全部完成 → 死锁
          const stuck = graph.nodeIds.filter(id => !done.has(id))
          const detail = stuck
            .map(id => `${labelOf(id)}（入度 ${graph.inDegree[id] || 0}）`)
            .join('、')
          throw new Error(`检测到死锁：以下节点互相等待，无法继续执行 → ${detail}`)
        }
        throw new Error(
          `调度停滞：并发上限（总 ${this.limits.total} / AI ${this.limits.ai}）过小，或存在未识别的结构问题`
        )
      }

      try {
        await Promise.race(running.values())
      } catch (e: any) {
        // 单个节点失败即整体失败（与顺序引擎一致：下游不再执行）
        this.host.log(`❌ 并发执行中断：${e?.message || e}`, 'error')
        throw e
      }
    }

    if (firstError) throw firstError
    return { completed: done.size, total }
  }
}

export default ConcurrentScheduler
