import type { BoardKind } from '@/types'

/**
 * 吧配置的唯一真源。
 *
 * 左栏、列表页、详情页都从这里取，不要各写一份 —— 否则改吧名要改好几处，
 * 漏一处就会出现「教程吧里写着经验分享吧」。
 *
 * ⚠ 两个吧是**固定枚举**，不建吧目录表、不做自建吧。
 * 新增吧 = 加一个 key + 后端 normalizeBoard 放行 + 路由不变（靠 :board 参数）。
 */
export const BOARD_CONF: Record<BoardKind, { name: string; short: string; slug: BoardKind; desc: string }> = {
  tutorial: {
    name: '教程吧',
    short: '教',
    slug: 'tutorial',
    desc: '功能怎么用、参数怎么填，按模块整理的说明书'
  },
  experience: {
    name: '经验分享吧',
    short: '经',
    slug: 'experience',
    desc: '踩过的坑、跑通的方法论，用户之间的交流区'
  }
}

/** 左栏展示顺序：教程在上（说明书优先），经验分享在下（交流区）。 */
export const BOARD_ORDER: BoardKind[] = ['tutorial', 'experience']

/** 把任意路由参数收成合法吧名；非法值回落 experience（不 404，用户会手搓地址）。 */
export function normalizeBoard(value: unknown): BoardKind {
  return value === 'tutorial' ? 'tutorial' : 'experience'
}

/** 教程中心的路由前缀，顶栏按钮和返回按钮都用它拼，避免写死字符串。 */
export const TUTORIAL_BASE = '/tutorial'

/** 「首页」入口 = 默认吧（经验分享吧）。左栏顶上那一项用它，不另开首页组件。 */
export const BOARD_HOME_PATH = `${TUTORIAL_BASE}/board/experience`

export const tutorialBoardPath = (board: BoardKind) => `${TUTORIAL_BASE}/board/${board}`
export const tutorialPostPath = (id: number | string) => `${TUTORIAL_BASE}/post/${id}`
export const TUTORIAL_QUICKREF_PATH = `${TUTORIAL_BASE}/quickref`
