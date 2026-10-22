import type { NodeExecuteResult, NodeLogger, WNode } from '@/types/workflow'

/** 开始节点：把 config.initialVariables 注入上下文 */
export class StartExecutor {
  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    logger.info('✅ 工作流开始')
    const initial = node.data?.config?.initialVariables
    if (initial && typeof initial === 'object') {
      const keys = Object.keys(initial)
      if (keys.length > 0) {
        Object.assign(context, initial)
        logger.info(`📋 初始化 ${keys.length} 个变量: ${keys.join(', ')}`)
        keys.forEach(k => {
          const v = initial[k]
          const shown = typeof v === 'string' && v.length > 50 ? v.substring(0, 50) + '...' : v
          logger.info(`  - ${k} = ${shown}`)
        })
      } else {
        logger.warning('⚠️ 开始节点未配置初始变量')
      }
    } else {
      logger.warning('⚠️ 开始节点未配置初始变量')
    }
    return { outputs: {}, data: null }
  }
}

export default StartExecutor
