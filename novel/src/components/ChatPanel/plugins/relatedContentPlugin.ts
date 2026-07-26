// 关联内容注入插件
// 将关联的章节/备忘录内容拼接到消息中
import type { ChatPlugin, PluginContext } from './type'

export const relatedContentPlugin: ChatPlugin = {
  name: 'related-content-inject',
  enabled: true,
  description: '将关联的章节/备忘录内容注入到用户消息中',
  process: (content: string, context: PluginContext) => {
    // 关联内容已由 promptInjectPlugin 处理
    // 此插件保留为扩展点，用于未来可能的不同注入策略
    return content
  },
}
