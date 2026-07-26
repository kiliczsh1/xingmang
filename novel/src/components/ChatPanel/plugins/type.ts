// 插件接口定义

export interface PluginContext {
  input: string
  relatedContent: Array<{ type: string; id: number | string; title: string; content: string }>
  selectedPrompts: any[]
  worldBookLinked: boolean
  worldBookEntries?: any[]
  [key: string]: any
}

export interface ChatPlugin {
  name: string
  enabled: boolean
  description: string
  process: (content: string, context: PluginContext) => string
}
