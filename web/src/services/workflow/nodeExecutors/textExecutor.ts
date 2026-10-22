import type { NodeExecuteResult, NodeLogger, WNode } from '@/types/workflow'
import { compileTemplate } from '../template'

/** 文本节点：静态文本 + {{变量}} 替换 */
export class TextExecutor {
  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const config = node.data?.config || {}
    const text = config.text
    if (!text) throw new Error('未设置文本内容')
    const out = compileTemplate(String(text), context)
    logger.info(`📝 处理文本 (${out.length} 字符)`)
    return { outputs: { text: out, lastOutput: out }, data: out }
  }
}

export default TextExecutor
