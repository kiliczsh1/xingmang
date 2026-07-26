// 提示词 + 关联内容注入插件
// 对应 Write.vue 行 7414-7449 的提示词+关联内容拼装逻辑
import type { ChatPlugin, PluginContext } from './type'

export const promptInjectPlugin: ChatPlugin = {
  name: 'prompt-inject',
  enabled: true,
  description: '将选中的提示词和关联内容注入到用户消息前面',
  process: (content: string, context: PluginContext) => {
    const parts: string[] = []

    // 提示词
    const promptContents = context.selectedPrompts
      .map((p: any) => p.content)
      .filter(Boolean)
      .join('\n\n')
    if (promptContents) {
      parts.push(`提示词：\n${promptContents}`)
    }

    // 关联内容
    const relatedContentSummary = context.relatedContent.length > 0
      ? `关联内容：\n${context.relatedContent
          .map(item => `${item.type === 'chapter' ? '章节' : '备忘录'}：${item.title}\n${item.content}`)
          .join('\n\n')}`
      : ''
    if (relatedContentSummary) {
      parts.push(relatedContentSummary)
    }

    // 用户输入
    if (content.trim()) {
      parts.push(content.trim())
    }

    return parts.join('\n\n')
  },
}
