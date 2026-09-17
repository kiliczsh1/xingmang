// 世界书注入插件
// 对应 Write.vue 行 7425-7435 的世界书注入逻辑
import type { ChatPlugin, PluginContext } from './type'

/**
 * 从世界书条目中匹配关键词
 * 从 Write.vue 的 matchWorldBookEntries 搬迁
 */
function matchWorldBookEntries(input: string, entries: any[]): any[] {
  if (!input || !entries || entries.length === 0) return []
  const lowerInput = input.toLowerCase()
  return entries.filter(entry => {
    if (entry.disable) return false
    const keywords = entry.keywords || entry.keyword || ''
    if (!keywords) return false
    const keywordList = keywords.split(',').map((k: string) => k.trim().toLowerCase()).filter(Boolean)
    return keywordList.some((kw: string) => lowerInput.includes(kw))
  })
}

/**
 * 拼接世界书内容
 * 从 Write.vue 的 injectWorldBookContent 搬迁
 */
function injectWorldBookContent(prefix: string, entries: any[]): string {
  if (!entries || entries.length === 0) return prefix || ''
  const content = entries.map(entry => {
    const name = entry.name || entry.comment || ''
    const content = entry.content || ''
    return name ? `${name}：${content}` : content
  }).join('\n\n')
  return prefix ? `${prefix}\n\n${content}` : content
}

export const worldBookInjectPlugin: ChatPlugin = {
  name: 'worldbook-inject',
  enabled: true,
  description: '根据用户输入匹配世界书词条并注入',
  process: (content: string, context: PluginContext) => {
    if (!context.worldBookLinked) return content

    const bookEntries = context.worldBookEntries || []
    if (bookEntries.length === 0) return content

    const matchResults = matchWorldBookEntries(context.input, bookEntries)
    if (matchResults.length === 0) return content

    const injection = injectWorldBookContent('', matchResults).trim()
    if (!injection) return content

    // 将世界书内容插入到已有内容之后、用户原始输入之前
    // 保持与旧 sendMessage 一致的格式：提示词 → 世界书 → 关联内容 → 用户输入
    // 但此插件在世界书环节只注入世界书部分，由 promptInjectPlugin 负责最终拼装
    // 这里返回 "世界书：\n{injection}\n\n{content}" 以便后续插件链处理
    return `世界书：\n${injection}\n\n${content}`
  },
}
