// 插件聚合层：内置插件清单 + 管理面映射 + 与聊天链的持久化联动
// 插件定义唯一来源是各 plugins/*.ts 的模块单例对象；本模块只做聚合与派生，
// 页面展示、启停持久化、聊天链状态都从这里读写，保证一处定义、多方一致。
import type { ChatPlugin } from './type'
import type { PluginManifest, PluginSource } from '@/types'
import { getPluginEnabled, savePluginPref } from '@/utils/pluginPrefs'
import { promptInjectPlugin } from './promptInjectPlugin'
import { worldBookInjectPlugin } from './worldBookInjectPlugin'
import { relatedContentPlugin } from './relatedContentPlugin'
import { regexFilterPlugin } from './regexFilterPlugin'
export { LOCAL_PLUGINS, loadLocalPlugins, addLocalPlugin, removeLocalPlugin, setLocalPluginEnabled, getLocalRecords, safeIcon } from './local'

/** 内置插件清单（注册顺序 = 聊天链处理顺序） */
export const BUILTIN_PLUGINS: ChatPlugin[] = [
  promptInjectPlugin,
  worldBookInjectPlugin,
  relatedContentPlugin,
  regexFilterPlugin,
]

/** 内置插件展示元信息：可读中文名 / 图标组件名 / 品牌色 */
export const BUILTIN_META: Record<string, Pick<PluginManifest, 'label' | 'icon' | 'color'>> = {
  'prompt-inject': { label: '提示词注入', icon: 'ChatDotRound', color: '#2563eb' },
  'worldbook-inject': { label: '世界书注入', icon: 'MagicStick', color: '#7c3aed' },
  'related-content-inject': { label: '关联内容注入', icon: 'Document', color: '#16a34a' },
  'regex-filter': { label: '正则过滤', icon: 'Tools', color: '#d97706' },
}

const DEFAULT_SOURCE: PluginSource = 'builtin'

/** 运行时 ChatPlugin → 管理面 PluginManifest（enabled 优先读本地持久化覆盖） */
export function toManifest(plugin: ChatPlugin): PluginManifest {
  const meta = BUILTIN_META[plugin.name] || { label: plugin.name, icon: 'Box', color: '#64748b' }
  return {
    name: plugin.name,
    label: meta.label,
    description: plugin.description,
    enabled: getPluginEnabled(plugin.name, plugin.enabled),
    source: DEFAULT_SOURCE,
    icon: meta.icon,
    color: meta.color,
  }
}

/**
 * 把本地持久化的启停偏好重放到插件单例上。
 * 在聊天链 register 之前调用，使"插件页的开关状态"在 SPA 会话内传播到聊天链。
 */
export function applyStoredPluginPrefs() {
  for (const plugin of BUILTIN_PLUGINS) {
    plugin.enabled = getPluginEnabled(plugin.name, plugin.enabled)
  }
}

/**
 * 单点写入：改插件单例 enabled + 写 localStorage，保证页面状态与聊天链行为永远一致。
 * Plugins.vue 的开关一律走这里。
 */
export function setBuiltinEnabled(name: string, enabled: boolean) {
  const plugin = BUILTIN_PLUGINS.find(p => p.name === name)
  if (plugin) {
    plugin.enabled = enabled
  }
  savePluginPref(name, enabled)
}
