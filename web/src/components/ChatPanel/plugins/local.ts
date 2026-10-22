// 本地插件注册表：从 zip 包安装的插件
//
// 包格式约定（供插件作者参考）：
//   plugin.zip
//   ├── manifest.json   # 必需：{ name, label, version, description?, author?, icon?, color?, tags?, entry? }
//   └── <entry>.js      # 可选：入口代码，通过 window.__installPlugin 暴露 ChatPlugin
//
// 入口代码契约（index.js 模板）：
//   window.__installPlugin = {
//     name: 'my-plugin',                 // 与 manifest.name 一致
//     description: '一句话说明',
//     enabled: true,
//     process: (content, context) => { /* 返回处理后的内容 */ return content }
//   }
//
// 持久化：localStorage `plugin-local`（manifest + 入口代码源码，跨刷新重放）。
// 运行：入口代码在本机执行（new Function 包裹），插件对象是模块级单例，
//       与内置插件同机制参与聊天链；启停复用 plugin-prefs。
import type { ChatPlugin } from './type'
import { getPluginEnabled, savePluginPref } from '@/utils/pluginPrefs'

const STORAGE_KEY = 'plugin-local'

export interface LocalPluginManifest {
  name: string
  label: string
  version: string
  description: string
  author?: string
  icon?: string
  color?: string
  tags?: string[]
  entry?: string
  main?: string // 兼容写法：与 entry 二选一
}

export interface LocalPluginRecord {
  manifest: LocalPluginManifest
  entryCode: string
  installedAt: string
}

// 安全默认图标名（避免 manifest 里的非法图标名破坏渲染）
const SAFE_ICONS = new Set([
  'Box', 'FolderAdd', 'Upload', 'MagicStick', 'ChatDotRound', 'Document', 'Tools',
  'User', 'DataLine', 'Lightning', 'Search', 'EditPen', 'Star', 'Collection', 'Connection'
])

// 模块级单例：插件中心启停/卸载即时生效，聊天链共享同一引用
export const LOCAL_PLUGINS: ChatPlugin[] = []

let loaded = false

/**
 * 执行入口代码并取出导出的 ChatPlugin。
 * 兼容两种写法：
 *  - 契约写法：window.__installPlugin = { … }
 *  - ES 模块写法：export default { … }（自动改写为 window.__installPlugin 赋值）
 * 入口代码的副作用（如 UI 插件挂载按钮）在本次执行时生效。
 */
function runEntryCode(code: string): any {
  const execute = (src: string): any => {
    try {
      ;(window as any).__installPlugin = null
      // eslint-disable-next-line no-new-func
      new Function('window', src)(window)
      return (window as any).__installPlugin
    } catch {
      return null
    }
  }
  let exported = execute(code)
  if (!exported && /\bexport\s+default\s*\{/.test(code)) {
    exported = execute(code.replace(/\bexport\s+default\s+/, 'window.__installPlugin = '))
  }
  return exported
}

/** 入口文件字段：manifest.entry 或 manifest.main（兼容），忽略前导 ./ */
export function resolveEntryName(manifest: LocalPluginManifest): string | null {
  const raw = manifest.entry || manifest.main
  return raw ? String(raw).replace(/^\.\//, '') : null
}

/** 由记录构建可运行 ChatPlugin（入口代码执行失败时降级为 no-op 扩展点） */
function buildPlugin(rec: LocalPluginRecord): ChatPlugin {
  const base: ChatPlugin = {
    name: rec.manifest.name,
    enabled: true,
    description: rec.manifest.description || '',
    process: (content: string) => content,
  }
  if (rec.entryCode.trim()) {
    const exported = runEntryCode(rec.entryCode)
    if (exported && typeof exported.process === 'function') {
      base.name = exported.name || base.name
      base.enabled = exported.enabled !== false
      base.description = exported.description || base.description
      base.process = exported.process
    }
  }
  // 复用 plugin-prefs 的启停覆盖（与内置插件同一套持久化）
  base.enabled = getPluginEnabled(base.name, base.enabled)
  return base
}

function readRecords(): LocalPluginRecord[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function persist(records: LocalPluginRecord[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(records))
  } catch {
    /* localStorage 不可用/超限时静默降级（仅内存生效） */
  }
}

/** 首次进入聊天链前调用：从本地重放已安装插件 */
export function loadLocalPlugins() {
  if (loaded) return
  loaded = true
  for (const rec of readRecords()) {
    LOCAL_PLUGINS.push(buildPlugin(rec))
  }
}

/** 安装本地插件（同名覆盖）；返回构建后的 ChatPlugin，失败返回 null */
export function addLocalPlugin(rec: LocalPluginRecord): ChatPlugin | null {
  const plugin = buildPlugin(rec)
  const idx = LOCAL_PLUGINS.findIndex(p => p.name === rec.manifest.name)
  if (idx >= 0) LOCAL_PLUGINS.splice(idx, 1)
  LOCAL_PLUGINS.push(plugin)
  const records = readRecords().filter(r => r.manifest.name !== rec.manifest.name)
  records.push(rec)
  persist(records)
  return plugin
}

/** 卸载本地插件（从单例 + 本地存储移除，并清理 plugin-prefs 覆盖） */
export function removeLocalPlugin(name: string) {
  const idx = LOCAL_PLUGINS.findIndex(p => p.name === name)
  if (idx >= 0) LOCAL_PLUGINS.splice(idx, 1)
  persist(readRecords().filter(r => r.manifest.name !== name))
  try {
    const prefs = JSON.parse(localStorage.getItem('plugin-prefs') || '{}')
    delete prefs[name]
    localStorage.setItem('plugin-prefs', JSON.stringify(prefs))
  } catch {
    /* 忽略 */
  }
}

/** 本地插件启停（改单例 + 写 plugin-prefs，与内置插件同一持久化） */
export function setLocalPluginEnabled(name: string, enabled: boolean) {
  const plugin = LOCAL_PLUGINS.find(p => p.name === name)
  if (plugin) plugin.enabled = enabled
  savePluginPref(name, enabled)
}

/** 已安装的本地插件记录（含 manifest，供管理页渲染） */
export function getLocalRecords(): LocalPluginRecord[] {
  return readRecords()
}

/** 校验 manifest.icon 是否安全可用，非法则回退 Box */
export function safeIcon(icon?: string): string {
  return icon && SAFE_ICONS.has(icon) ? icon : 'Box'
}
