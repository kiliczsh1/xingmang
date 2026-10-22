// 手机端阅读预览 · 分页与翻页内核
// 分页用 CSS 多列切栏：列宽 = 屏宽 - 左右边距，片间空白 = 2×边距 ⇒ 每个溢出的列就是一页。
// 双轨（A/B）各持一份相同内容，翻页时把相邻页摆到入轨再做位移动画。
import { nextTick, reactive, ref, type Ref } from 'vue'
import { READER_TURN_DURATION, type TurnMode } from './useReaderSettings'

export interface PagerTracks {
  screen: Ref<HTMLElement | null>
  /** 两轨的容器（.rx-pg），动画直接改它的 transform */
  pageEls: [Ref<HTMLElement | null>, Ref<HTMLElement | null>]
  /** 两轨的多列容器（.rx-flow），分页靠它的 translateX 定位 */
  flowEls: [Ref<HTMLElement | null>, Ref<HTMLElement | null>]
  scrollBox: Ref<HTMLElement | null>
}

export interface PagerOptions {
  tracks: PagerTracks
  scrollMode: () => boolean
  duration: (mode: TurnMode) => number
  reducedMotion: () => boolean
  /** 越过章边界时通知父组件换章（父组件通过 nav 事件驱动，换章后调用 onChapterChanged） */
  onBoundary: (dir: number) => void
  /** 章内位置变化（用于刷新页眉页脚 / 进度） */
  onProgress?: () => void
}

const clamp = (v: number, a: number, b: number) => Math.min(b, Math.max(a, v))

export function useReaderPager(options: PagerOptions) {
  const { tracks } = options
  const geom = reactive({ W: 0, padX: 0, padY: 0, colW: 0, gap: 0, slot: 1 })
  const pages = ref(1)
  const page = ref(0)
  /** 当前轨下标：0 = A，1 = B */
  const activeTrack = ref(0)
  /** 每轨当前停在哪一页（模板里绑成 translateX） */
  const trackPage: [Ref<number>, Ref<number>] = [ref(0), ref(0)]
  /** 滚动模式的章内比例 0–1 */
  const scrollRatio = ref(0)

  /** 翻页前的落点提示：跨越边界后新章要停在首页还是末页 */
  const pendingEdge = ref<'first' | 'last'>('first')
  /** 作废飞行中的翻页回调 */
  let turnToken = 0
  let remeasureQueued = false
  let resizeObserver: ResizeObserver | null = null

  const el = (i: number) => tracks.pageEls[i].value
  const flowEl = (i: number) => tracks.flowEls[i].value

  /* ---------------- 几何与分页 ---------------- */

  /** 写回边距：必须比量列宽更早，否则会读到上一次的边距，页数就算少了 */
  function applyMetrics() {
    const screen = tracks.screen.value
    if (!screen) return
    const W = screen.clientWidth
    const cs = getComputedStyle(screen)
    const marginRatio = parseFloat(cs.getPropertyValue('--rx-margin')) || 0.074
    const safeT = parseFloat(cs.getPropertyValue('--rx-safeT')) || 0
    const padX = Math.round(W * marginRatio)
    screen.style.setProperty('--rx-padX', `${padX}px`)
    screen.style.setProperty('--rx-padY', `${Math.round(padX * 0.78) + safeT}px`)
  }

  function measure() {
    applyMetrics()
    const screen = tracks.screen.value
    if (!screen) return

    if (options.scrollMode()) {
      const box = tracks.scrollBox.value
      const W = box ? box.clientWidth : screen.clientWidth
      const cs = box ? getComputedStyle(box) : null
      geom.W = W
      geom.padX = cs ? parseFloat(cs.paddingLeft) || 0 : 0
      geom.padY = cs ? parseFloat(cs.paddingTop) || 0 : 0
      geom.colW = Math.max(48, W)
      geom.gap = 0
      geom.slot = Math.max(1, W)
      pages.value = 1
      return
    }

    const pg = el(activeTrack.value)
    const flow = flowEl(activeTrack.value)
    if (!pg || !flow) return
    const cs = getComputedStyle(pg)
    const padX = parseFloat(cs.paddingLeft) || 0
    const padY = parseFloat(cs.paddingTop) || 0
    const W = pg.clientWidth
    const colW = Math.max(48, W - 2 * padX)
    const gap = 2 * padX
    geom.W = W
    geom.padX = padX
    geom.padY = padY
    geom.colW = colW
    geom.gap = gap
    geom.slot = colW + gap
    tracks.flowEls.forEach((f) => {
      if (!f.value) return
      f.value.style.columnWidth = `${colW}px`
      f.value.style.columnGap = `${gap}px`
    })
    pages.value = pagesOf(flow)
  }

  function pagesOf(flow: HTMLElement) {
    const sw = flow.scrollWidth
    if (sw <= geom.colW + 1) return 1
    return Math.max(1, Math.ceil((sw - geom.colW - 1) / geom.slot) + 1)
  }

  /** 内容或尺寸变化后补一次重排：字体回退、换章、缩放都可能让首帧列几何偏小 */
  function scheduleRemeasure() {
    if (remeasureQueued) return
    remeasureQueued = true
    requestAnimationFrame(() => {
      remeasureQueued = false
      const r = chapterRatio()
      measure()
      setRatio(r)
    })
  }

  /* ---------------- 位置 ---------------- */

  function place() {
    trackPage[0].value = page.value
    trackPage[1].value = page.value
  }

  function setPage(p: number) {
    page.value = clamp(p, 0, pages.value - 1)
    place()
    options.onProgress?.()
  }

  /** 章内进度 0–1（滚动模式看滚动条，分页模式看页序） */
  function chapterRatio() {
    if (options.scrollMode()) return clamp(scrollRatio.value, 0, 1)
    return pages.value <= 1 ? 0 : page.value / (pages.value - 1)
  }

  /** 按章内进度落位 */
  function setRatio(r: number) {
    const v = clamp(r, 0, 1)
    if (options.scrollMode()) {
      const box = tracks.scrollBox.value
      if (box) {
        scrollRatio.value = v
        box.scrollTop = (box.scrollHeight - box.clientHeight) * v
      }
    } else {
      setPage(Math.round(v * (pages.value - 1)))
    }
    options.onProgress?.()
  }

  /** 手动拖动滚动容器时同步比例 */
  function onScroll() {
    const box = tracks.scrollBox.value
    if (!box) return
    const max = box.scrollHeight - box.clientHeight
    scrollRatio.value = max > 0 ? clamp(box.scrollTop / max, 0, 1) : 0
    options.onProgress?.()
  }

  /** 设置变化后重排（默认保住当前阅读位置） */
  function relayout(keep = true) {
    const r = chapterRatio()
    measure()
    setRatio(keep ? r : 0)
    scheduleRemeasure()
  }

  /** 章节内容变化后调用：重新分页并按落点定位 */
  function onChapterChanged(edge: 'first' | 'last' | 'keep' = 'first') {
    turnToken++
    resetStyles()
    activeTrack.value = 0
    nextTick(() => {
      measure()
      if (edge === 'keep') setRatio(chapterRatio())
      else setRatio(edge === 'last' ? 1 : 0)
      scheduleRemeasure()
    })
  }

  /* ---------------- 翻页动画 ---------------- */

  function resetStyles() {
    ;[0, 1].forEach((i) => {
      const node = el(i)
      if (!node) return
      node.style.transition = 'none'
      node.style.transform = ''
      node.style.opacity = ''
      node.style.zIndex = ''
      node.style.clipPath = ''
    })
  }

  function turn(dir: number) {
    if (!dir) return
    if (options.scrollMode()) {
      // 上下滑动模式：左右点击 / 方向键直接换章
      pendingEdge.value = dir > 0 ? 'first' : 'last'
      options.onBoundary(dir)
      return
    }
    const to = page.value + dir
    if (to < 0) {
      pendingEdge.value = 'last'
      options.onBoundary(-1)
      return
    }
    if (to > pages.value - 1) {
      pendingEdge.value = 'first'
      options.onBoundary(1)
      return
    }
    const mode = currentMode()
    if (mode === 'none' || options.reducedMotion()) {
      setPage(to)
      return
    }
    animateTurn(mode, dir, to)
  }

  function currentMode(): TurnMode {
    return currentModeHint
  }

  /** 父组件每次翻页前写入当前模式 */
  let currentModeHint: TurnMode = 'scroll'
  function setMode(mode: TurnMode) {
    currentModeHint = mode
  }

  function animateTurn(mode: TurnMode, dir: number, to: number) {
    const dur = Math.max(0, options.duration(mode))
    const outIdx = activeTrack.value
    const incIdx = 1 - outIdx
    const outEl = el(outIdx)
    const incEl = el(incIdx)
    if (!outEl || !incEl) {
      setPage(to)
      return
    }
    trackPage[incIdx].value = to
    trackPage[outIdx].value = page.value
    resetStyles()

    const ease = 'cubic-bezier(.22,.61,.36,1)'
    const t = `transform ${dur}ms ${ease}, opacity ${dur}ms ${ease}`
    const token = ++turnToken
    const done = () => {
      if (token === turnToken) finishTurn(to)
    }

    if (mode === 'fade') {
      incEl.style.zIndex = '2'
      incEl.style.opacity = '0'
      void incEl.getBoundingClientRect()
      incEl.style.transition = t
      incEl.style.opacity = '1'
    } else if (mode === 'cover') {
      if (dir > 0) {
        incEl.style.zIndex = '2'
        incEl.style.transform = 'translateX(100%)'
        void incEl.getBoundingClientRect()
        incEl.style.transition = t
        incEl.style.transform = 'translateX(0)'
      } else {
        outEl.style.zIndex = '2'
        void outEl.getBoundingClientRect()
        outEl.style.transition = t
        outEl.style.transform = 'translateX(100%)'
      }
    } else {
      // slide：两轨同时平移
      incEl.style.transform = `translateX(${dir > 0 ? 100 : -100}%)`
      void incEl.getBoundingClientRect()
      incEl.style.transition = t
      outEl.style.transition = t
      incEl.style.transform = 'translateX(0)'
      outEl.style.transform = `translateX(${dir > 0 ? -100 : 100}%)`
    }

    // 收尾走 transitionend，定时兜底（后台标签页会节流定时器）
    incEl.addEventListener('transitionend', function onEnd(e) {
      if (e.target !== incEl || (mode === 'fade' ? e.propertyName !== 'opacity' : e.propertyName !== 'transform')) return
      incEl.removeEventListener('transitionend', onEnd)
      done()
    })
    window.setTimeout(done, dur + 80)
  }

  function finishTurn(to: number) {
    turnToken++
    page.value = clamp(to, 0, pages.value - 1)
    activeTrack.value = 1 - activeTrack.value
    resetStyles()
    place()
    options.onProgress?.()
  }

  /* ---------------- 生命周期 ---------------- */

  function start() {
    stop()
    if (window.ResizeObserver) {
      let lw = 0
      let lh = 0
      resizeObserver = new ResizeObserver(() => {
        const screen = tracks.screen.value
        if (!screen) return
        const W = screen.clientWidth
        const H = screen.clientHeight
        if (!W || (W === lw && H === lh)) return
        lw = W
        lh = H
        relayout(true)
      })
      if (tracks.screen.value) resizeObserver.observe(tracks.screen.value)
    }
    // 保留上次阅读位置（重开面板时不至于回到首页）
    measure()
    setRatio(chapterRatio())
    scheduleRemeasure()
  }

  function stop() {
    resizeObserver?.disconnect()
    resizeObserver = null
    turnToken++
  }

  return {
    geom,
    pages,
    page,
    trackPage,
    scrollRatio,
    start,
    stop,
    measure,
    relayout,
    scheduleRemeasure,
    onChapterChanged,
    pendingEdge,
    setMode,
    setPage,
    setRatio,
    onScroll,
    chapterRatio,
    turn,
    /** 供皮肤扩展（后续接仿真/打字机等自定义翻页） */
    finishTurn
  }
}

export type ReaderPager = ReturnType<typeof useReaderPager>

/** 与 READER_TURN_DURATION 对齐的取值助手（父组件注入，保持 pager 无状态依赖） */
export const defaultDuration = (mode: TurnMode) => READER_TURN_DURATION[mode] ?? 240
