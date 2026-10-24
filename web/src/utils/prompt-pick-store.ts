/**
 * 「更多提示词」选择器的本地状态：收藏 / 使用次数 / 每个工具最近一次的选择。
 *
 * 对齐云端版的收藏与最近选择行为：
 * - 收藏 = 云端版的「已收藏」页签；
 * - 使用次数 = 本地没有云端热度，用它给「推荐」页签排序；
 * - 最近选择 = 按工具 + 书籍维度记住上次选择，避免每次打开弹窗都要重选。
 */

const FAV_KEY = 'xm_prompt_more_favs_v1'
const USAGE_KEY = 'xm_prompt_more_usage_v1'
const LAST_KEY = 'xm_prompt_last_pick_v1'

function readJson<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    if (!raw) return fallback
    const parsed = JSON.parse(raw)
    return parsed ?? fallback
  } catch {
    return fallback
  }
}

function writeJson(key: string, value: unknown) {
  try {
    localStorage.setItem(key, JSON.stringify(value))
  } catch {
    /* localStorage 不可用时忽略 */
  }
}

/* ---------------- 收藏 ---------------- */

export function readFavIds(): number[] {
  const list = readJson<unknown>(FAV_KEY, [])
  return Array.isArray(list) ? list.filter((x): x is number => typeof x === 'number') : []
}

export function writeFavIds(ids: number[]) {
  writeJson(FAV_KEY, Array.from(new Set(ids)))
}

/* ---------------- 使用次数 ---------------- */

export function readUsage(): Record<string, number> {
  const map = readJson<Record<string, unknown>>(USAGE_KEY, {})
  const out: Record<string, number> = {}
  if (map && typeof map === 'object') {
    for (const [k, v] of Object.entries(map)) {
      if (typeof v === 'number' && Number.isFinite(v)) out[k] = v
    }
  }
  return out
}

export function bumpUsage(promptId: number) {
  const map = readUsage()
  const key = String(promptId)
  map[key] = (map[key] || 0) + 1
  writeJson(USAGE_KEY, map)
}

/* ---------------- 每个工具最近一次的选择 ---------------- */

type LastPickMap = Record<string, number>

function readLastPicks(): LastPickMap {
  const map = readJson<LastPickMap>(LAST_KEY, {})
  return map && typeof map === 'object' ? map : {}
}

/** 取某个工具（或卡包）上次选中的提示词 id */
export function readLastPick(scope: string): number | null {
  const map = readLastPicks()
  const v = map[scope]
  return typeof v === 'number' ? v : null
}

/** 记住某个工具（或卡包）本次选中的提示词 id */
export function saveLastPick(scope: string, promptId: number | null) {
  const map = readLastPicks()
  if (promptId === null) delete map[scope]
  else map[scope] = promptId
  writeJson(LAST_KEY, map)
}

/* ---------------- 审稿/去痕的「固定提示词」列表 ---------------- */

const PINS_KEY = 'xm_prompt_pins_v1'

type PinMap = Record<string, number[]>

/** 取某个工具固定的提示词 id 列表（保持顺序） */
export function readPinned(scope: string): number[] {
  const map = readJson<PinMap>(PINS_KEY, {})
  const list = map && typeof map === 'object' ? map[scope] : null
  return Array.isArray(list) ? list.filter((x): x is number => typeof x === 'number') : []
}

/** 覆盖写入某个工具固定的提示词 id 列表 */
export function writePinned(scope: string, ids: number[]) {
  const map = readJson<PinMap>(PINS_KEY, {})
  const next: PinMap = map && typeof map === 'object' ? { ...map } : {}
  next[scope] = Array.from(new Set(ids))
  writeJson(PINS_KEY, next)
}
