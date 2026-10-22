// 手机端阅读预览 · 状态与主题（暖木 · 书房夜灯皮肤）
// 排版参数、主题色板、翻页模式等全部在这里；写入 localStorage，跨会话记住
import { computed, reactive, watch } from 'vue'

/* ---------------- 常量表 ---------------- */

export type FontKey = 'system' | 'song' | 'hei' | 'kai' | 'fang'
export type TurnMode = 'scroll' | 'cover' | 'slide' | 'fade' | 'none'
export type Level = 'compact' | 'std' | 'loose'
export type MarginKey = 'narrow' | 'std' | 'wide'
export type AlignKey = 'justify' | 'left'
export type ZonesKey = 'page' | 'chapter'

export interface FontOption {
  label: string
  stack: string
}

export const READER_FONTS: Record<FontKey, FontOption> = {
  system: {
    label: '默认',
    stack: `-apple-system,"PingFang SC","Microsoft YaHei","Segoe UI",sans-serif`
  },
  song: { label: '宋体', stack: `Georgia,"Songti SC","SimSun",serif` },
  hei: { label: '黑体', stack: `"PingFang SC","Microsoft YaHei","Helvetica Neue",sans-serif` },
  kai: { label: '楷体', stack: `"Kaiti SC","KaiTi","STKaiti",serif` },
  fang: { label: '仿宋', stack: `"FangSong","STFangsong","FangSong_GB2312",serif` }
}

export const READER_LINE_HEIGHT: Record<Level, number> = { compact: 1.6, std: 1.9, loose: 2.25 }
export const READER_PARA_GAP: Record<Level, string> = { compact: '0.45em', std: '0.9em', loose: '1.5em' }
/** 左右边距占屏宽比例（窄 / 标准 / 宽） */
export const READER_MARGIN: Record<MarginKey, number> = { narrow: 0.052, std: 0.074, wide: 0.105 }

export interface ModeOption {
  value: TurnMode
  label: string
}

/** 上下滑动（滚动）放在第一位，也是默认值 */
export const READER_MODES: ModeOption[] = [
  { value: 'scroll', label: '上下滑动' },
  { value: 'cover', label: '覆盖' },
  { value: 'slide', label: '平移' },
  { value: 'fade', label: '淡入' },
  { value: 'none', label: '无动画' }
]

/** 翻页动画时长（毫秒）；none 由状态机直接瞬切 */
export const READER_TURN_DURATION: Record<TurnMode, number> = {
  scroll: 0,
  cover: 260,
  slide: 280,
  fade: 240,
  none: 0
}

export interface ReaderPalette {
  name: string
  bg: string
  fg: string
  sub: string
  accent: string
  line: string
  /** 夜间色板：供「夜间」按钮一键切换与皮肤钩子 */
  night?: boolean
  /** 纸面纹理覆盖层（CSS background 值） */
  texture?: string
}

const PAPER_TEXTURE =
  'repeating-linear-gradient(45deg,rgba(120,90,50,.045) 0 1px,transparent 1px 3px),' +
  'repeating-linear-gradient(-45deg,rgba(120,90,50,.035) 0 1px,transparent 1px 3px)'

/** 暖木 · 书房夜灯：5 套主题色板（含 1 套夜间） */
export const READER_PALETTES: ReaderPalette[] = [
  { name: '暖黄', bg: '#f6e8cf', fg: '#463320', sub: '#9c7f5c', accent: '#8c5a1e', line: '#e9d8b8', texture: PAPER_TEXTURE },
  { name: '米白', bg: '#fbf6ec', fg: '#3f382c', sub: '#9a8f7c', accent: '#9a6b2f', line: '#efe6d7', texture: PAPER_TEXTURE },
  { name: '墨绿', bg: '#e8ece0', fg: '#2d3a2b', sub: '#7f8c77', accent: '#4a6b3a', line: '#d9e0d1', texture: PAPER_TEXTURE },
  { name: '酒红', bg: '#f4e6e0', fg: '#442a24', sub: '#a08477', accent: '#8f3b2e', line: '#ecd9d2', texture: PAPER_TEXTURE },
  {
    name: '深棕夜',
    bg: '#1d1611',
    fg: '#c7b191',
    sub: '#7a6750',
    accent: '#d9a34e',
    line: '#2b2118',
    night: true,
    texture: 'repeating-linear-gradient(45deg,rgba(200,160,100,.035) 0 1px,transparent 1px 3px)'
  }
]

/* ---------------- 状态 ---------------- */

export interface ReaderSettings {
  mode: TurnMode
  fontSize: number
  font: FontKey
  /** 行距档位 */
  lineHeight: Level
  /** 段距档位 */
  paraGap: Level
  margin: MarginKey
  indent: boolean
  align: AlignKey
  bold: boolean
  /** 主题色板下标 */
  palette: number
  /** 亮度 28–100 */
  bright: number
  /** 页眉页脚常显 */
  chromeAlways: boolean
  /** 菜单 6 秒自动隐藏 */
  autoHide: boolean
  /** 跟随正文 */
  follow: boolean
  /** 点击分区：翻页 / 换章 */
  zones: ZonesKey
}

/** 默认：上下滑动 + 楷体 + 松行距 + 暖黄纸（书房夜灯的默认阅读档） */
export const READER_DEFAULTS: ReaderSettings = {
  mode: 'scroll',
  fontSize: 16,
  font: 'kai',
  lineHeight: 'loose',
  paraGap: 'std',
  margin: 'wide',
  indent: true,
  align: 'justify',
  bold: false,
  palette: 0,
  bright: 100,
  chromeAlways: false,
  autoHide: false,
  follow: false,
  zones: 'page'
}

const STORAGE_KEY = 'xm-reader-settings-v1'

function loadSettings(): ReaderSettings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...READER_DEFAULTS }
    const saved = JSON.parse(raw) as Partial<ReaderSettings>
    return { ...READER_DEFAULTS, ...saved }
  } catch {
    return { ...READER_DEFAULTS }
  }
}

/* 模块级单例：面板与设置抽屉共享同一份状态，不必层层透传 props */
const settings = reactive<ReaderSettings>(loadSettings())
const palette = computed(() => READER_PALETTES[settings.palette] ?? READER_PALETTES[0])
const fontStack = computed(() => (READER_FONTS[settings.font] ?? READER_FONTS.system).stack)
const lineHeight = computed(() => String(READER_LINE_HEIGHT[settings.lineHeight]))
const paraGap = computed(() => READER_PARA_GAP[settings.paraGap])
const modeLabel = computed(() => READER_MODES.find((m) => m.value === settings.mode)?.label ?? '上下滑动')

let saveTimer: number | undefined
function save() {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(settings))
  } catch {
    /* 隐私模式 / 配额满：忽略，不影响阅读 */
  }
}
function saveSoon() {
  if (saveTimer) window.clearTimeout(saveTimer)
  saveTimer = window.setTimeout(save, 400)
}
watch(settings, saveSoon)

/** 把当前设置写进 CSS 变量（挂在面板根元素上，随面板卸载自动失效） */
function applyTokens(el: HTMLElement | null | undefined) {
  if (!el) return
  const p = palette.value
  const set = (k: string, v: string) => el.style.setProperty(k, v)
  set('--rx-fs', `${settings.fontSize}px`)
  set('--rx-font', fontStack.value)
  set('--rx-lh', lineHeight.value)
  set('--rx-para-gap', paraGap.value)
  set('--rx-indent', settings.indent ? '2em' : '0em')
  set('--rx-align', settings.align)
  set('--rx-fw', settings.bold ? '600' : '400')
  set('--rx-margin', String(READER_MARGIN[settings.margin]))
  set('--pg-bg', p.bg)
  set('--pg-fg', p.fg)
  set('--pg-sub', p.sub)
  set('--pg-accent', p.accent)
  set('--pg-line', p.line)
  set('--pg-texture', p.texture ?? 'none')
  el.dataset.night = p.night ? '1' : '0'
}

/** 夜间 / 日间一键切换（在昼间色板与夜间色板之间来回） */
function toggleNight() {
  const isNight = !!palette.value.night
  const idx = READER_PALETTES.findIndex((p) => !!p.night === !isNight)
  if (idx >= 0) settings.palette = idx
}

function reset() {
  Object.assign(settings, READER_DEFAULTS)
}

/**
 * 阅读设置：响应式状态 + 持久化 + 主题/排印令牌。
 * 令牌写在传入的元素上（预览面板根节点），不碰全局主题（遵守 C5：页面不得劫持全局主题）。
 */
export function useReaderSettings() {
  return {
    settings,
    palette,
    paletteList: READER_PALETTES,
    fontOptions: Object.entries(READER_FONTS).map(([value, o]) => ({ value: value as FontKey, ...o })),
    modeOptions: READER_MODES,
    modeLabel,
    applyTokens,
    toggleNight,
    reset
  }
}

export type ReaderSettingsApi = ReturnType<typeof useReaderSettings>
