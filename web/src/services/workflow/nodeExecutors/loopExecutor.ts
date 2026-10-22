import type { NodeExecuteResult, NodeLogger, WNode, WorkflowExecutorHost } from '@/types/workflow'

const HISTORY_SEPARATOR = '\n\n--- 上一章分隔 ---\n\n'

/**
 * 循环节点 —— 直译源项目 loopExecutor。
 * 双出口：sourceHandle='loop-body' 走循环体，'loop-done' 走循环完成之后。
 * 注意：源版的 condition 循环并未实现真正的条件判定，只按 maxIterations 跑满；此处保持一致。
 */
export class LoopExecutor {
  host: WorkflowExecutorHost

  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const config = node.data?.config || {}
    const loopType = config.loopType || 'fixed'

    let iterations = 1
    let arraySource: any[] | null = null
    let resolved = false

    if (loopType === 'array') {
      logger.info('📋 循环类型: 数组循环')
      const configured = config.arraySource?.variableName
      if (configured) {
        logger.info(`🔍 使用配置的数组源: ${configured}`)
        let value = context[configured]
        if (typeof value === 'string' && value.trim().startsWith('[')) {
          try {
            const parsed = JSON.parse(value)
            if (Array.isArray(parsed)) {
              context[configured] = parsed
              value = parsed
              logger.info(`🔄 自动将 ${configured} 从JSON字符串解析为数组`)
            }
          } catch (e: any) {
            logger.warning(`⚠️ 解析 ${configured} 失败: ${e?.message || e}`)
          }
        }
        if (Array.isArray(value)) {
          arraySource = value
          iterations = value.length
          resolved = true
          logger.info(`✅ 从 ${configured} 获取循环次数: ${iterations}`)
        } else {
          logger.warning(`⚠️ ${configured} 不是数组，当前值类型: ${typeof value}`)
        }
      }
      if (!resolved) {
        for (const name of ['chapterPlans', 'chapters', 'plans', 'items']) {
          if (context[name] === undefined) continue
          let value = context[name]
          if (typeof value === 'string' && value.trim().startsWith('[')) {
            try {
              const parsed = JSON.parse(value)
              if (Array.isArray(parsed)) {
                context[name] = parsed
                value = parsed
                logger.info(`🔄 自动将 ${name} 从JSON字符串解析为数组`)
              }
            } catch {
              /* 解析失败则继续按原值处理 */
            }
          }
          if (Array.isArray(value) && value.length > 0) {
            arraySource = value
            iterations = value.length
            resolved = true
            logger.info(`🔍 自动检测到数组 ${name}: ${iterations} 个元素`)
            break
          }
        }
      }
      if (!resolved) {
        logger.warning('⚠️ 未找到有效的数组源，将使用固定循环次数 1')
        iterations = config.fixedCount || 1
      }
    } else if (loopType === 'fixed') {
      iterations = config.fixedCount || config.iterations || 1
      logger.info(`📋 循环类型: 固定循环，次数: ${iterations}`)
    } else if (loopType === 'condition') {
      iterations = config.maxIterations || 100
      logger.info(`📋 循环类型: 条件循环，最大次数: ${iterations}`)
    }

    const iterVar = 'currentIteration'
    const itemVar = 'currentItem'
    const keepContext = config.keepContext !== false

    logger.info(`🔄 开始循环 ${iterations} 次`)
    if (loopType === 'array' && arraySource) {
      logger.info(`📦 数组源包含 ${arraySource.length} 个元素`)
      logger.info(`🔖 当前项变量名: ${itemVar}`)
      logger.info(`📊 数组内容预览: ${JSON.stringify(arraySource.slice(0, 2)).substring(0, 300)}...`)
    }
    logger.info(`🔖 循环变量名: ${iterVar}`)

    const results: any[] = []

    const bodyEdge = this.host.workflow.edges.find(e => e.source === node.id && e.sourceHandle === 'loop-body')
    if (!bodyEdge) {
      logger.warning('⚠️ 未找到循环体入口，跳过循环执行')
      return { outputs: { loopResults: [], iterationCount: 0 }, data: [] }
    }
    const bodyNode = this.host.workflow.nodes.find(n => n.id === bodyEdge.target)
    if (!bodyNode) throw new Error(`循环体入口节点不存在: ${bodyEdge.target}`)

    // 断点续跑：源版用 node.data.currentIteration 记进度
    const resumeFrom = Number(node.data.currentIteration || 0)
    const startIndex = resumeFrom > 0 && resumeFrom < iterations ? resumeFrom : 0
    node.data.iterations = iterations
    this.host.updateNodeData(node.id, { iterations })
    if (startIndex > 0) logger.info(`⏭️ 继续执行：从第 ${startIndex + 1}/${iterations} 次开始`)

    for (let i = startIndex; i < iterations; i++) {
      if (this.host.stopped) {
        logger.warning(`⏹️ 工作流已停止，在第 ${i + 1}/${iterations} 次循环时中断`)
        break
      }
      logger.info(`  🔁 循环 ${i + 1}/${iterations}`)

      // 历史上下文（上一章窗口）
      try {
        const windowSize = Number(config.historyWindow || 3)
        const history = Array.isArray(context.chapterHistory) ? context.chapterHistory : []
        const tail = history.slice(-windowSize)
        context.previousChapters = tail
          .map((k: any) => {
            if (typeof k === 'string') return k
            try {
              return JSON.stringify(k, null, 2)
            } catch {
              return String(k)
            }
          })
          .join(HISTORY_SEPARATOR)
        context.previousChapterCount = history.length
        logger.info(`  📚 历史上下文准备完成：已有 ${history.length} 章，注入最近 ${tail.length} 章`)
      } catch (e: any) {
        logger.warning(`  ⚠️ 历史上下文准备失败：${e?.message || e}`)
      }

      context[iterVar] = i + 1
      context.currentIteration = i + 1

      const plans = Array.isArray(context.chapterPlans) ? context.chapterPlans : null
      const plan = plans ? plans[i] : null
      const title = plan
        ? plan.title || plan.name || plan.chapterTitle || plan['标题'] || plan['章节标题'] || plan['题目'] || undefined
        : undefined
      const total = plans?.length || iterations
      if (title) context.currentChapterTitle = title

      node.data.currentIteration = i + 1
      this.host.updateNodeData(node.id, {
        currentIteration: i + 1,
        totalChapters: total,
        currentChapterTitle: title
      })

      if (loopType === 'array' && arraySource) {
        context[itemVar] = arraySource[i]
        logger.info(`  🔖 设置当前项: ${JSON.stringify(arraySource[i]).substring(0, 200)}`)
      }

      try {
        await this.executeLoopBody(bodyNode, node.id, context, logger)
      } catch (e: any) {
        const msg = e?.message || String(e)
        logger.error(`  ❌ 循环第 ${i + 1} 次执行失败: ${msg}`)
        results.push({ iteration: i + 1, error: msg, status: 'failed' })
        continue
      }

      if (!keepContext && i < iterations - 1) {
        // iterVar === 'currentIteration'，只留计算键，避免重复键
        const keep: Record<string, any> = {
          [iterVar]: context[iterVar],
          chapterHistory: context.chapterHistory,
          novelOutline: context.novelOutline,
          characters: context.characters
        }
        Object.keys(context).forEach(k => {
          if (!Object.prototype.hasOwnProperty.call(keep, k)) delete context[k]
        })
        Object.assign(context, keep)
      }

      results.push({ iteration: i + 1 })
      logger.success(`  ✅ 循环第 ${i + 1} 次执行成功`)
    }

    logger.success(`✅ 循环完成，共执行 ${iterations} 次`)
    return { outputs: { loopResults: results, iterationCount: iterations }, data: results }
  }

  /** 沿单一主线走完循环体，遇到回边（指回 loop 节点）或无出边即停 */
  private async executeLoopBody(start: WNode, loopId: string, context: Record<string, any>, logger: NodeLogger) {
    let cur: WNode | undefined = start
    const visited = new Set<string>()
    while (cur) {
      if (visited.has(cur.id)) throw new Error(`循环体内检测到死循环: ${cur.id}`)
      visited.add(cur.id)

      const result = await this.host.executeNode(cur, { preventFollowEdges: true })
      if (this.host.stopped || result?.aborted) break

      const outs = this.host.workflow.edges.filter(e => e.source === cur!.id)
      if (outs.length === 0 || outs.find(e => e.target === loopId)) break
      if (outs.length > 1) logger.warning(`  ⚠️ 循环体节点 ${cur.id} 有多个下游分支，只执行第一个`)

      const nextId = outs[0].target
      const next = this.host.workflow.nodes.find(n => n.id === nextId)
      if (!next) throw new Error(`下游节点不存在: ${nextId}`)
      cur = next
    }
  }
}

export default LoopExecutor
