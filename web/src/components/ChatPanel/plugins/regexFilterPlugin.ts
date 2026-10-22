// 正则过滤插件（对 AI 响应内容的后处理）
// 注意：正则过滤在 useChatPanel 中作为显示层处理，不作为消息发送前的插件
// 此插件保留为扩展点
import type { ChatPlugin, PluginContext } from './type'

export const regexFilterPlugin: ChatPlugin = {
  name: 'regex-filter',
  enabled: true,
  description: '对 AI 响应内容应用正则过滤（保留扩展点）',
  process: (content: string, _context: PluginContext) => {
    // 正则过滤在 useChatPanel 的显示层处理，此处不做处理
    return content
  },
}
