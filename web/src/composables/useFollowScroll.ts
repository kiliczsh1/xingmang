// 手机端阅读预览 · 正文追踪（跟随正文）
//
// 双向同步：编辑器滚动 → 预览跟随；预览翻页/滚动 → 编辑器跟随。
// 编辑器（富文本组件）的滚动事件拿不到，只能定时读它的滚动比例；但做了三件事让轮询无副作用：
//   1. 只在「跟随正文」打开且面板可见时轮询（150ms，隐藏时完全停掉）；
//   2. 比例没变化就直接 return，不做任何 DOM 写入；
//   3. 双向都带时间戳锁，避免互相打架（原来的 40ms 布尔锁换成时间窗）。
import { onBeforeUnmount, ref, type Ref } from 'vue'

export interface FollowScrollOptions {
  /** 读取编辑器当前滚动比例 0–1（整章） */
  getScrollRatio?: () => number | null
  /** 把编辑器滚到指定比例 */
  scrollToRatio?: (ratio: number) => void
  /** 预览是否可见（不可见时停止轮询） */
  visible: () => boolean
  /** 预览当前的章内进度 0–1 */
  getReaderRatio: () => number
  /** 把预览落到指定章内进度 */
  setReaderRatio: (ratio: number) => void
  /** 是否开启跟随 */
  enabled: () => boolean
}

const POLL_MS = 150
/** 程序性滚动后的静默窗口：这段时间内忽略对侧的回传，避免来回抖动 */
const LOCK_MS = 90

export function useFollowScroll(options: FollowScrollOptions) {
  /** 编辑器 → 预览 的挂起目标（比例变化才排队） */
  const lastEditorRatio = ref<number | null>(null)
  let timer: number | undefined
  let lockReader = 0
  let lockEditor = 0

  const now = () => performance.now()

  /** 预览侧位置变化后调用（翻页 / 滚动 / 换章） */
  function syncEditorFromReader() {
    if (!options.enabled()) return
    if (now() - lockEditor < LOCK_MS) return
    const set = options.scrollToRatio
    if (!set) return
    lockReader = now()
    set(options.getReaderRatio())
  }

  function poll() {
    if (!options.enabled() || !options.visible()) {
      lastEditorRatio.value = null
      return
    }
    if (now() - lockReader < LOCK_MS) return
    const ratio = options.getScrollRatio?.()
    if (ratio == null || Number.isNaN(ratio)) return
    // 比例没变就别动 DOM（编辑器静止时这里就直接返回）
    if (lastEditorRatio.value != null && Math.abs(ratio - lastEditorRatio.value) < 0.002) return
    lastEditorRatio.value = ratio
    options.setReaderRatio(ratio)
  }

  function start() {
    stop()
    timer = window.setInterval(poll, POLL_MS)
  }

  function stop() {
    if (timer) window.clearInterval(timer)
    timer = undefined
    lastEditorRatio.value = null
  }

  /** 跟随开关或可见性变化时调用 */
  function refresh(force = false) {
    if (options.enabled() && options.visible()) {
      if (force) {
        // 刚打开跟随时，先把预览对齐到编辑器当前位置
        const ratio = options.getScrollRatio?.()
        if (ratio != null) {
          lockReader = now()
          lastEditorRatio.value = ratio
          options.setReaderRatio(ratio)
        }
      }
      start()
    } else {
      stop()
    }
  }

  onBeforeUnmount(stop)

  return { refresh, start, stop, syncEditorFromReader, isPolling: () => timer !== undefined }
}

export type FollowScrollApi = ReturnType<typeof useFollowScroll>
