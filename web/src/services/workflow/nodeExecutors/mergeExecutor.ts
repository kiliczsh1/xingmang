import type { NodeExecuteResult, NodeLogger, WNode } from '@/types/workflow'

/**
 * 合并节点 —— 直译源项目 mergeExecutor。
 * mergeType: concat（默认）/ array / object
 */
export class MergeExecutor {
  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const config = node.data?.config || {}
    const mergeType = config.mergeType || 'concat'
    const inputs: any[] = []

    if (Array.isArray(config.inputVariables) && config.inputVariables.length > 0) {
      logger.info(`📥 收集指定的 ${config.inputVariables.length} 个输入变量`)
      config.inputVariables.forEach((name: string) => {
        const v = context[name]
        if (v !== null && v !== undefined) {
          if (Array.isArray(v)) {
            logger.info(`  • ${name}: 数组 (${v.length} 项)`)
            inputs.push(...v)
          } else {
            logger.info(`  • ${name}: ${typeof v === 'string' ? v.length + ' 字符' : typeof v}`)
            inputs.push(v)
          }
        } else {
          logger.warning(`  • ${name}: 不存在或为空`)
        }
      })
    } else {
      logger.warning('⚠️ 未配置 inputVariables，使用智能推断（不推荐）')
      for (const name of ['lastOutput', 'lastAIOutput', 'aiResult', 'merged', 'text']) {
        const v = context[name]
        if (v !== null && v !== undefined) {
          inputs.push(v)
          logger.info(`  • 自动选择: ${name}`)
          break
        }
      }
    }

    if (inputs.length === 0) {
      logger.error('❌ 没有找到可合并的输入！')
      logger.info('💡 提示：请在合并节点配置中指定 inputVariables')
      logger.info(
        `💡 可用的上下文变量: ${Object.keys(context)
          .filter(k => context[k] !== null && context[k] !== undefined)
          .join(', ')}`
      )
      return { outputs: { merged: '', lastOutput: '' }, data: '' }
    }

    logger.info(`🔗 开始合并 ${inputs.length} 个输入 (类型: ${mergeType})`)

    const separator = config.separator || '\n\n'
    let merged: any
    try {
      switch (mergeType) {
        case 'concat': {
          if (inputs.length > 100) logger.warning(`⚠️ 输入数量较多 (${inputs.length}个)，使用分块合并`)
          const CHUNK = 50
          const chunks: string[] = []
          for (let i = 0; i < inputs.length; i += CHUNK) {
            const piece = inputs
              .slice(i, i + CHUNK)
              .map(v => String(v))
              .join(separator)
            chunks.push(piece)
            if (piece.length > 10 * 1024 * 1024) {
              logger.warning(`⚠️ 分块 ${Math.floor(i / CHUNK) + 1} 较大: ${Math.round(piece.length / 1024 / 1024)}MB`)
            }
          }
          merged = chunks.join(separator)
          if (merged.length > 200 * 1024 * 1024) {
            throw new Error(`合并内容过大 (${Math.round(merged.length / 1024 / 1024)}MB)，超出安全限制`)
          }
          break
        }
        case 'array':
          merged = inputs
          logger.info(`  数组合并完成，共 ${inputs.length} 项`)
          break
        case 'object': {
          const objs = inputs.filter(v => typeof v === 'object' && v !== null && !Array.isArray(v))
          merged = Object.assign({}, ...objs)
          logger.info(`  对象合并完成，共 ${Object.keys(merged).length} 个键`)
          break
        }
        default:
          merged = inputs.map(v => String(v)).join('\n\n')
      }

      const size = typeof merged === 'string' ? merged.length : JSON.stringify(merged).length
      const human =
        size > 1024 * 1024 ? `${(size / 1024 / 1024).toFixed(2)} MB` : size > 1024 ? `${(size / 1024).toFixed(2)} KB` : `${size} 字符`
      logger.success(`✅ 合并完成 (${human})`)
    } catch (e: any) {
      logger.error(`❌ 合并失败: ${e?.message || e}`)
      logger.info('💡 建议：检查输入数据大小，可能需要：')
      logger.info('   1. 减少循环次数')
      logger.info('   2. 在循环中只保存必要的历史')
      logger.info('   3. 使用 Context 节点的 maxHistory 限制历史长度')
      throw e
    }

    return { outputs: { merged, lastOutput: merged }, data: merged }
  }
}

export default MergeExecutor
