// 插件启停偏好持久化（localStorage）
// 约定：单 key JSON，只存显式覆盖项，缺省启用。
// 例：{ "prompt-inject": false, "worldbook-inject": true }
// 新增内置插件未写入时 getPluginEnabled 自动回退默认值，无需迁移。
const PREFS_KEY = 'plugin-prefs'

export function loadPluginPrefs(): Record<string, boolean> {
  try {
    const raw = localStorage.getItem(PREFS_KEY)
    if (!raw) return {}
    const parsed = JSON.parse(raw)
    return typeof parsed === 'object' && parsed !== null ? parsed : {}
  } catch {
    return {}
  }
}

/** 读取单个插件启用状态；无显式覆盖时回退 defaultVal */
export function getPluginEnabled(name: string, defaultVal = true): boolean {
  const prefs = loadPluginPrefs()
  return typeof prefs[name] === 'boolean' ? prefs[name] : defaultVal
}

/** 写入单个插件启用状态 */
export function savePluginPref(name: string, enabled: boolean) {
  const prefs = loadPluginPrefs()
  prefs[name] = enabled
  try {
    localStorage.setItem(PREFS_KEY, JSON.stringify(prefs))
  } catch {
    // localStorage 不可用（隐私模式/配额）时静默降级为仅内存生效
  }
}
