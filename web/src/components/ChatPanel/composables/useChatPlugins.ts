// 插件链管理器
import { ref } from 'vue'
import type { ChatPlugin, PluginContext } from '../plugins/type'

export function useChatPlugins() {
  const plugins = ref<ChatPlugin[]>([])

  const register = (plugin: ChatPlugin) => {
    // 避免重复注册同名插件
    if (plugins.value.some(p => p.name === plugin.name)) return
    plugins.value.push(plugin)
  }

  const unregister = (name: string) => {
    plugins.value = plugins.value.filter(p => p.name !== name)
  }

  const getPlugin = (name: string) => {
    return plugins.value.find(p => p.name === name)
  }

  /** 按注册顺序链式处理，返回最终内容 */
  const process = (context: PluginContext): string => {
    let content = context.input
    for (const plugin of plugins.value) {
      if (plugin.enabled) {
        content = plugin.process(content, context)
      }
    }
    return content
  }

  /** 切换插件启用/禁用 */
  const togglePlugin = (name: string, enabled?: boolean) => {
    const plugin = plugins.value.find(p => p.name === name)
    if (plugin) {
      plugin.enabled = enabled ?? !plugin.enabled
    }
  }

  return {
    plugins,
    register,
    unregister,
    getPlugin,
    process,
    togglePlugin,
  }
}
