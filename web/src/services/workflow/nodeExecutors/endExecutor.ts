import type { NodeExecuteResult, NodeLogger, WNode } from '@/types/workflow'

/** 结束节点：把当前上下文整体作为输出（源版行为） */
export class EndExecutor {
  async execute(_node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    logger.success('🏁 到达结束节点')
    const snapshot = { ...context }
    return { outputs: snapshot, data: snapshot }
  }
}

export default EndExecutor
