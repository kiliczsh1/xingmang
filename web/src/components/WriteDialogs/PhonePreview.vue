<!-- 手机端阅读预览（暖木 · 书房夜灯皮肤）
     形态：可拖拽浮窗 → 一键全屏；屏内是一套完整的阅读器
     交互：上下滑动（默认）/ 覆盖 / 平移 / 淡入 / 无动画 · 点击三分区 · 自动隐藏页眉页脚
          阅读菜单 · 目录 · 界面设置（排版 / 页面 / 主题）· 亮度 · 缎带书签 · 跟随正文
     对外契约与旧版一致（props / emits 未变），因此 Write.vue 无需改动
-->
<template>
  <Teleport to="body">
    <div v-if="zoomed" class="rx-zoom-veil" @click="toggleZoom(false)" />
    <div
      v-if="visible"
      ref="rootRef"
      class="xm-reader"
      :class="{ 'is-scrolling': isScrollMode, 'is-zoomed': zoomed }"
      :style="{
        left: pos.x + 'px',
        top: pos.y + 'px',
        '--dev-w': devW + 'px',
        '--scr-h': scrH + 'px'
      }"
    >
      <!-- 面板条（可拖拽） -->
      <div
        ref="barRef"
        class="rx-panel-bar"
        :class="{ 'is-grabbing': dragging }"
        @pointerdown="onBarDown"
      >
        <span class="rx-bar-title">📱 手机端阅读预览</span>
        <span class="rx-bar-sp" />
        <button
          type="button"
          class="rx-bar-btn"
          :class="{ 'is-on': settings.follow }"
          title="开启后与正文滚动互相同步"
          @pointerdown.stop
          @click="toggleFollow"
        >
          跟随正文
        </button>
        <button
          type="button"
          class="rx-bar-btn"
          :title="zoomed ? '还原为浮窗（F）' : '全屏阅读（F）'"
          @pointerdown.stop
          @click="toggleZoom()"
        >
          {{ zoomed ? '◱' : '⤢' }}
        </button>
        <button type="button" class="rx-bar-btn is-x" title="退出预览" @pointerdown.stop @click="emit('close')">
          ✕
        </button>
      </div>

      <!-- 机身 + 屏幕 -->
      <div ref="deviceRef" class="rx-device is-right">
        <span class="rx-key k1" /><span class="rx-key k2" /><span class="rx-key k3" />
        <div ref="screenRef" class="rx-screen" :data-night="palette.night ? '1' : '0'">
          <div class="rx-island" />
          <div class="rx-gloss" />

          <!-- 页眉：章节 / 时间 / 电量（自动隐藏） -->
          <header class="rx-chrome rx-chrome-top" :class="{ 'is-on': chromeOn }">
            <span class="ct-l">‹ 返回</span>
            <span class="ct-m">{{ chromeTitle }}</span>
            <span class="ct-r">
              <span>{{ clock }}</span>
              <span class="rx-bat" :style="{ '--bat': battery }" />
            </span>
          </header>

          <!-- 阅读区 -->
          <div ref="readerRef" class="rx-reader" @click="onReaderTap">
            <div class="rx-mode rx-mode-paged">
              <div ref="pgARef" class="rx-pg">
                <div ref="flowARef" class="rx-flow" :style="flowStyleA" v-html="contentHtml" />
              </div>
              <div ref="pgBRef" class="rx-pg">
                <div ref="flowBRef" class="rx-flow" :style="flowStyleB" v-html="contentHtml" />
              </div>
            </div>
            <div class="rx-mode rx-mode-scroll">
              <div ref="scrollRef" class="rx-sflow" @scroll="pager.onScroll" v-html="contentHtml" />
            </div>
          </div>

          <!-- 屏内叠加层 -->
          <div class="rx-mask" :style="{ opacity: maskOpacity }" />
          <div class="rx-zones" :class="{ 'is-on': zonesOn }">
            <i>{{ zoneLeft }}</i><i>菜单</i><i>{{ zoneRight }}</i>
          </div>
          <div class="rx-pill" :class="{ 'is-on': pillOn }">{{ pillText }}</div>
          <div
            ref="ribbonRef"
            class="rx-ribbon"
            :class="{ 'is-dragging': ribbonDragging }"
            :style="{ top: ribbonTop + 'px' }"
            title="拖动缎带跳章"
            @pointerdown="onRibbonDown"
            @pointermove="onRibbonMove"
            @pointerup="onRibbonUp"
            @pointercancel="onRibbonUp"
          />

          <!-- 页脚：进度 / 页码 / 估余 -->
          <footer class="rx-chrome rx-chrome-bottom" :class="{ 'is-on': chromeOn }">
            <span>全书 {{ Math.round(bookRatio * 100) }}%</span>
            <span>{{ footerMid }}</span>
            <span>估余 {{ remainMinutes }} 分钟</span>
          </footer>

          <!-- 阅读菜单：上栏 -->
          <div class="rx-menu rx-menu-top" :class="{ 'is-on': menuOpen }">
            <div class="mt-row">
              <button type="button" class="rx-mbtn back" title="收起菜单 / 退出全屏" @click="onMenuBack">‹</button>
              <span class="m-book">{{ bookTitle }}</span>
              <button type="button" class="rx-mbtn" @click="openSheet('toc')">目录</button>
              <button type="button" class="rx-mbtn" @click="openSheet('set')">界面</button>
            </div>
          </div>

          <!-- 阅读菜单：下栏 -->
          <div class="rx-menu rx-menu-bottom" :class="{ 'is-on': menuOpen }">
            <div class="rx-bright" :class="{ 'is-on': brightOpen }">
              <span>☀</span>
              <input
                class="rx-range"
                type="range"
                min="28"
                max="100"
                :value="settings.bright"
                @input="onBrightInput"
              />
              <span>◐</span>
            </div>
            <div class="rx-seek">
              <button type="button" class="sk-btn" @click="seekStep(-1)">上一章</button>
              <div
                ref="seekRef"
                class="rx-sk-track"
                @pointerdown="onSeekDown"
                @pointermove="onSeekMove"
                @pointerup="onSeekUp"
                @pointercancel="onSeekUp"
              >
                <div class="rx-sk-fill" :style="{ width: seekPercent + '%' }" />
                <div class="rx-sk-knob" :style="{ left: seekPercent + '%' }" />
                <span class="rx-sk-tip" :class="{ 'is-on': seekDragging }" :style="{ left: seekPercent + '%' }">
                  {{ seekTip }}
                </span>
              </div>
              <button type="button" class="sk-btn" @click="seekStep(1)">下一章</button>
            </div>
            <div class="rx-nav">
              <button type="button" @click="openSheet('toc')"><b>☰</b>目录</button>
              <button type="button" @click="openSheet('set')"><b>◐</b>界面</button>
              <button type="button" @click="brightOpen = !brightOpen"><b>☀</b>亮度</button>
              <button type="button" :class="{ 'is-on': !!palette.night }" @click="onToggleNight">
                <b>☾</b>夜间
              </button>
              <button type="button" :class="{ 'is-on': settings.follow }" @click="toggleFollow"><b>⇅</b>跟随</button>
            </div>
          </div>

          <!-- 目录 / 界面设置 -->
          <ReaderSheets
            :open="sheet"
            :chapter-index="props.chapterIndex"
            :total-chapters="props.totalChapters"
            @close="closeSheets"
            @jump="onTocJump"
          />

          <div class="rx-veil" :class="{ 'is-on': !!sheet }" @click="closeSheets" />
          <div class="rx-toast" :class="{ 'is-on': !!toastText }">{{ toastText }}</div>
          <div class="rx-home" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import ReaderSheets from './PhonePreview/ReaderSheets.vue'
import { READER_TURN_DURATION, useReaderSettings } from '@/composables/useReaderSettings'
import { useReaderPager } from '@/composables/useReaderPager'
import { useFollowScroll } from '@/composables/useFollowScroll'
import { countWords, escapeHtml, splitChapterTitle, toPlainParagraphs } from '@/utils/readerText'
import '@/styles/reader-theme.css'

defineOptions({ name: 'PhonePreview' })

const props = defineProps<{
  visible: boolean
  title: string
  content: string
  chapterIndex: number
  totalChapters: number
  /** 获取编辑器滚动百分比 0-1（跟随正文用） */
  getScrollRatio?: () => number | null
  /** 编辑器滚动到指定百分比（跟随正文反向同步） */
  scrollToRatio?: (ratio: number) => void
}>()

const emit = defineEmits<{
  (e: 'close'): void
  /** dir 是相对当前章的偏移量（±1 换一章，目录可直接给绝对偏移） */
  (e: 'nav', dir: number): void
}>()

const { settings, palette, applyTokens, toggleNight, modeLabel } = useReaderSettings()

/* ---------------- 内容 ---------------- */
const paragraphs = computed(() => toPlainParagraphs(props.content))
const head = computed(() => splitChapterTitle(props.title))
const currentWords = computed(() => countWords(props.content))
const isScrollMode = computed(() => settings.mode === 'scroll')

const contentHtml = computed(() => {
  if (!paragraphs.value.length) {
    return `<div class="rx-empty"><span class="ico">◌</span><p>本章还没有正文<br>在写作页继续写，预览会同步</p></div>`
  }
  const no = head.value.no ? `<span class="ch-no">${escapeHtml(head.value.no)}</span>` : ''
  const title = `<h2 class="rx-ch-head">${no}<span class="ch-t">${escapeHtml(head.value.name)}</span></h2>`
  return title + paragraphs.value.map((p) => `<p>${escapeHtml(p)}</p>`).join('')
})

const bookTitle = computed(() => head.value.name || props.title || '阅读预览')
/** Write.vue 用 findIndex 传 chapterIndex，找不到时是 -1，显示层兜一层 */
const chapterNo = computed(() => Math.max(0, props.chapterIndex) + 1)
const chromeTitle = computed(
  () => `${bookTitle.value} · 第${chapterNo.value}/${Math.max(1, props.totalChapters)}章`
)

/* ---------------- 分页与翻页内核 ---------------- */
const rootRef = ref<HTMLElement | null>(null)
const barRef = ref<HTMLElement | null>(null)
const deviceRef = ref<HTMLElement | null>(null)
const screenRef = ref<HTMLElement | null>(null)
const readerRef = ref<HTMLElement | null>(null)
const pgARef = ref<HTMLElement | null>(null)
const pgBRef = ref<HTMLElement | null>(null)
const flowARef = ref<HTMLElement | null>(null)
const flowBRef = ref<HTMLElement | null>(null)
const scrollRef = ref<HTMLElement | null>(null)

const reducedMotion = ref(false)
if (typeof window !== 'undefined' && window.matchMedia) {
  const mq = window.matchMedia('(prefers-reduced-motion: reduce)')
  reducedMotion.value = mq.matches
  mq.addEventListener?.('change', (e) => (reducedMotion.value = e.matches))
}

const pager = useReaderPager({
  tracks: {
    screen: screenRef,
    pageEls: [pgARef, pgBRef],
    flowEls: [flowARef, flowBRef],
    scrollBox: scrollRef
  },
  scrollMode: () => settings.mode === 'scroll',
  duration: (mode) => READER_TURN_DURATION[mode] ?? 240,
  reducedMotion: () => reducedMotion.value,
  onBoundary: (dir) => requestChapter(dir),
  onProgress: () => syncAll()
})

const flowStyleA = computed(() => ({
  transform: `translateX(${-pager.trackPage[0].value * pager.geom.slot}px)`
}))
const flowStyleB = computed(() => ({
  transform: `translateX(${-pager.trackPage[1].value * pager.geom.slot}px)`
}))

/* ---------------- 跟随正文 ---------------- */
const follow = useFollowScroll({
  getScrollRatio: () => props.getScrollRatio?.() ?? null,
  scrollToRatio: (r) => props.scrollToRatio?.(r),
  visible: () => props.visible,
  getReaderRatio: () => pager.chapterRatio(),
  setReaderRatio: (r) => pager.setRatio(r),
  enabled: () => settings.follow
})

function syncAll() {
  follow.syncEditorFromReader()
}

/* ---------------- 面板尺寸 / 拖拽 / 全屏 ---------------- */
const PANEL_W = 336
const PANEL_H = 604
const devW = ref(PANEL_W)
const scrH = ref(PANEL_H)
const pos = ref({ x: 0, y: 84 })
const zoomed = ref(false)
const dragging = ref(false)
let dragOffset = { x: 0, y: 0 }

function panelSize() {
  if (!zoomed.value) return { w: PANEL_W, h: PANEL_H }
  const w = Math.min(430, window.innerWidth - 40)
  const h = Math.max(420, Math.min(880, window.innerHeight - 110))
  return { w, h }
}

function placePanel(x: number, y: number, w: number, h: number) {
  pos.value = { x, y }
  devW.value = w
  scrH.value = h
}

function initPosition() {
  const sz = panelSize()
  placePanel(Math.max(8, window.innerWidth - sz.w - 28), 84, sz.w, sz.h)
}

function toggleZoom(force?: boolean) {
  const first = rootRef.value?.getBoundingClientRect()
  const next = force === undefined ? !zoomed.value : force
  if (next === zoomed.value) return
  zoomed.value = next
  const sz = panelSize()
  const x = Math.max(8, (window.innerWidth - sz.w) / 2)
  const y = Math.max(10, (window.innerHeight - sz.h - 44) / 2)
  placePanel(x, y, sz.w, sz.h)
  nextTick(() => {
    const node = rootRef.value
    const last = node?.getBoundingClientRect()
    if (node && first && last && !reducedMotion.value && last.width) {
      const sx = first.width / last.width
      const sy = first.height / last.height
      const dx = first.left - last.left
      const dy = first.top - last.top
      node.animate(
        [
          { transformOrigin: 'top left', transform: `translate(${dx}px, ${dy}px) scale(${sx}, ${sy})` },
          { transformOrigin: 'top left', transform: 'none' }
        ],
        { duration: 420, easing: 'cubic-bezier(.22,.61,.36,1)' }
      )
    }
    // 尺寸变了要重算法列几何（等动画收尾，避免中途量到中间态）
    window.setTimeout(() => pager.relayout(true), reducedMotion.value ? 0 : 430)
  })
  showToast(next ? '全屏阅读 · Esc 还原' : '已还原为浮窗')
}

function onBarDown(e: PointerEvent) {
  if (zoomed.value || e.button !== 0) return
  const rect = rootRef.value?.getBoundingClientRect()
  if (!rect) return
  dragging.value = true
  dragOffset = { x: e.clientX - rect.left, y: e.clientY - rect.top }
  barRef.value?.setPointerCapture?.(e.pointerId)
  window.addEventListener('pointermove', onBarMove)
  window.addEventListener('pointerup', onBarUp)
}

function onBarMove(e: PointerEvent) {
  if (!dragging.value) return
  const w = rootRef.value?.offsetWidth ?? PANEL_W
  const x = Math.min(Math.max(e.clientX - dragOffset.x, 6), window.innerWidth - w - 6)
  const y = Math.min(Math.max(e.clientY - dragOffset.y, 6), window.innerHeight - 70)
  pos.value = { x, y }
}

function onBarUp() {
  dragging.value = false
  window.removeEventListener('pointermove', onBarMove)
  window.removeEventListener('pointerup', onBarUp)
}

/* ---------------- 菜单 / 抽屉 / 提示 ---------------- */
const menuOpen = ref(false)
const sheet = ref<'toc' | 'set' | null>(null)
const brightOpen = ref(false)
const zonesOn = ref(false)
const toastText = ref('')
let toastTimer: number | undefined
let autoTimer: number | undefined

function showToast(msg: string, ms = 1600) {
  toastText.value = msg
  if (toastTimer) window.clearTimeout(toastTimer)
  toastTimer = window.setTimeout(() => (toastText.value = ''), ms)
}

function setMenu(open: boolean) {
  menuOpen.value = open
  if (!open) brightOpen.value = false
}

function openSheet(which: 'toc' | 'set') {
  sheet.value = which
  setMenu(false)
}

function closeSheets() {
  sheet.value = null
}

function onMenuBack() {
  if (zoomed.value) toggleZoom(false)
  else setMenu(false)
}

function onToggleNight() {
  toggleNight()
  showToast(palette.value.night ? '夜间模式' : `日间模式 · ${palette.value.name}`)
}

function toggleFollow() {
  settings.follow = !settings.follow
  showToast(settings.follow ? '已开启跟随正文' : '已关闭跟随正文')
}

watch(menuOpen, (open) => {
  if (autoTimer) window.clearTimeout(autoTimer)
  if (open && settings.autoHide) autoTimer = window.setTimeout(() => setMenu(false), 6000)
})

watch(
  () => settings.zones,
  () => {
    zonesOn.value = true
    window.setTimeout(() => (zonesOn.value = false), 2600)
  }
)

/* ---------------- 屏内信息 ---------------- */
const chromeOn = computed(() => settings.chromeAlways || menuOpen.value)
const pillOn = computed(() => !chromeOn.value)
const maskOpacity = computed(() => ((100 - settings.bright) / 100) * 0.78)
const zoneLeft = computed(() => {
  if (settings.zones === 'chapter') return '上一章'
  return isScrollMode.value ? '上下滑动' : '上一页'
})
const zoneRight = computed(() => {
  if (settings.zones === 'chapter') return '下一章'
  return isScrollMode.value ? '上下滑动' : '下一页'
})

const bookRatio = computed(() => {
  const total = Math.max(1, props.totalChapters)
  return Math.min(1, Math.max(0, (Math.max(0, props.chapterIndex) + pager.chapterRatio()) / total))
})
const footerMid = computed(() =>
  isScrollMode.value
    ? `第${chapterNo.value}章 · ${Math.round(pager.chapterRatio() * 100)}%`
    : `第${chapterNo.value}章 · ${pager.page.value + 1}/${pager.pages.value} 页`
)
const pillText = computed(() => {
  const p = `全书 ${Math.round(bookRatio.value * 100)}%`
  return isScrollMode.value ? p : `${p} · ${pager.page.value + 1}/${pager.pages.value} 页`
})
const remainMinutes = computed(() => {
  const left = Math.max(0, Math.round(currentWords.value * (1 - pager.chapterRatio())))
  return Math.max(1, Math.round(left / 320))
})

const clock = ref('')
const battery = ref('72%')
function tickClock() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  clock.value = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  const h = d.getHours()
  battery.value = h > 20 || h < 6 ? '34%' : '72%'
}

/* ---------------- 阅读区交互 ---------------- */
function requestChapter(dir: number) {
  const atStart = dir < 0 && props.chapterIndex <= 0
  const atEnd = dir > 0 && props.chapterIndex >= props.totalChapters - 1
  if (atStart || atEnd) {
    showToast(atStart ? '已经是第一章' : '已经是最后一章')
    return
  }
  emit('nav', dir)
}

function turnPage(dir: number) {
  pager.setMode(settings.mode)
  pager.turn(dir)
}

/**
 * 左右两侧的点击 / 方向键：
 * - 分页模式：翻页（「点击区=换章」时改为换章）
 * - 上下滑动模式：没有「页」可翻，默认不做任何事 —— 避免点到空白处就跳章；
 *   只有用户把「点击区」显式设成「换章」才换，且换章始终有明确的菜单 / 目录 / 滑杆入口
 */
function canSideAct() {
  return settings.zones === 'chapter' || !isScrollMode.value
}

function sideAction(dir: number) {
  if (settings.zones === 'chapter') {
    requestChapter(dir)
    return
  }
  if (isScrollMode.value) return
  turnPage(dir)
}

function onReaderTap(e: MouseEvent) {
  if (sheet.value) return
  const rect = readerRef.value?.getBoundingClientRect()
  if (!rect) return
  const x = (e.clientX - rect.left) / rect.width
  if (x >= 0.34 && x <= 0.66) {
    setMenu(!menuOpen.value)
    return
  }
  sideAction(x < 0.34 ? -1 : 1)
}

/* ---------------- 章节滑杆 ---------------- */
const seekRef = ref<HTMLElement | null>(null)
const seekDragging = ref(false)
const seekPercent = ref(0)
const seekTip = ref('')

function seekRatio(index: number) {
  const total = Math.max(1, props.totalChapters)
  return total <= 1 ? 0 : index / (total - 1)
}

function updateSeekVisual(index: number) {
  const p = index / Math.max(1, props.totalChapters - 1)
  seekPercent.value = Math.min(100, Math.max(0, p * 100))
}

function seekIndexAt(clientX: number) {
  const rect = seekRef.value?.getBoundingClientRect()
  if (!rect) return props.chapterIndex
  const p = Math.min(1, Math.max(0, (clientX - rect.left) / rect.width))
  const idx = Math.round(p * Math.max(0, props.totalChapters - 1))
  updateSeekVisual(idx)
  seekTip.value = `第 ${idx + 1} 章`
  return idx
}

function onSeekDown(e: PointerEvent) {
  seekDragging.value = true
  seekRef.value?.setPointerCapture?.(e.pointerId)
  seekIndexAt(e.clientX)
}

function onSeekMove(e: PointerEvent) {
  if (seekDragging.value) seekIndexAt(e.clientX)
}

function onSeekUp(e: PointerEvent) {
  if (!seekDragging.value) return
  seekDragging.value = false
  const idx = seekIndexAt(e.clientX)
  const delta = idx - props.chapterIndex
  if (delta) emit('nav', delta)
}

function seekStep(dir: number) {
  requestChapter(dir)
}

/* ---------------- 缎带书签（可拖拽跳章） ---------------- */
const ribbonRef = ref<HTMLElement | null>(null)
const ribbonDragging = ref(false)
const ribbonTop = ref(24)

function updateRibbon() {
  const h = screenRef.value?.clientHeight ?? PANEL_H
  ribbonTop.value = Math.round(24 + bookRatio.value * Math.max(40, h - 78))
}

function onRibbonDown(e: PointerEvent) {
  ribbonDragging.value = true
  ribbonRef.value?.setPointerCapture?.(e.pointerId)
  e.preventDefault()
}

function onRibbonMove(e: PointerEvent) {
  if (!ribbonDragging.value) return
  const rect = screenRef.value?.getBoundingClientRect()
  if (!rect) return
  const p = Math.min(1, Math.max(0, (e.clientY - rect.top - 24) / Math.max(40, rect.height - 78)))
  ribbonTop.value = Math.round(24 + p * Math.max(40, rect.height - 78))
}

function onRibbonUp(e: PointerEvent) {
  if (!ribbonDragging.value) return
  ribbonDragging.value = false
  const rect = screenRef.value?.getBoundingClientRect()
  if (!rect) return
  const p = Math.min(1, Math.max(0, (e.clientY - rect.top - 24) / Math.max(40, rect.height - 78)))
  const idx = Math.round(p * Math.max(0, props.totalChapters - 1))
  const delta = idx - props.chapterIndex
  if (delta) {
    emit('nav', delta)
    showToast(`书签 · 第 ${idx + 1} 章`)
  }
  updateRibbon()
}

/* ---------------- 目录跳章 ---------------- */
function onTocJump(delta: number) {
  closeSheets()
  if (delta) emit('nav', delta)
}

/* ---------------- 亮度 ---------------- */
function onBrightInput(e: Event) {
  const v = Number((e.target as HTMLInputElement).value)
  settings.bright = Number.isFinite(v) ? v : settings.bright
}

/* ---------------- 键盘（只保留不与编辑器冲突的键） ---------------- */
function onKeydown(e: KeyboardEvent) {
  if (!props.visible) return
  const t = e.target as HTMLElement | null
  if (t && (t.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(t.tagName))) return
  const k = e.key.toLowerCase()
  if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
    if (!canSideAct()) return
    e.preventDefault()
    sideAction(e.key === 'ArrowLeft' ? -1 : 1)
  } else if (e.key === 'Escape') {
    if (sheet.value) closeSheets()
    else if (menuOpen.value) setMenu(false)
    else if (zoomed.value) toggleZoom(false)
  } else if (k === 'f' && !e.metaKey && !e.ctrlKey && !e.altKey) {
    toggleZoom()
  }
}

/* ---------------- 生命周期与联动 ---------------- */
watch(contentHtml, () => {
  nextTick(() => pager.relayout(true))
})

// 换章：按翻页方向决定落在首页还是末页（放在 contentHtml 之后，保证由它收尾）
watch(
  () => props.chapterIndex,
  () => {
    nextTick(() => {
      pager.onChapterChanged(pager.pendingEdge.value === 'last' ? 'last' : 'first')
      pager.pendingEdge.value = 'first'
      updateRibbon()
    })
  }
)

watch(
  () => [
    settings.fontSize,
    settings.font,
    settings.lineHeight,
    settings.paraGap,
    settings.indent,
    settings.align,
    settings.bold,
    settings.margin,
    settings.palette,
    settings.mode
  ],
  () => {
    applyTokens(rootRef.value)
    nextTick(() => pager.relayout(true))
  }
)

watch([() => settings.follow, () => props.visible], () => follow.refresh(true))
watch(bookRatio, () => {
  updateRibbon()
  seekPercent.value = Math.min(100, Math.max(0, seekRatio(props.chapterIndex) * 100))
})

watch(
  () => props.visible,
  (v) => {
    if (v) {
      nextTick(() => {
        applyTokens(rootRef.value)
        pager.start()
        initPosition()
        updateRibbon()
        updateSeekVisual(props.chapterIndex)
        tickClock()
        follow.refresh(true)
        window.addEventListener('keydown', onKeydown)
      })
    } else {
      setMenu(false)
      closeSheets()
      follow.stop()
      pager.stop()
      window.removeEventListener('keydown', onKeydown)
    }
  },
  { immediate: true }
)

function onResize() {
  if (zoomed.value) {
    const sz = panelSize()
    placePanel(Math.max(8, (window.innerWidth - sz.w) / 2), Math.max(10, (window.innerHeight - sz.h - 44) / 2), sz.w, sz.h)
  } else {
    const w = rootRef.value?.offsetWidth ?? PANEL_W
    pos.value = {
      x: Math.min(Math.max(6, pos.value.x), Math.max(6, window.innerWidth - w - 6)),
      y: Math.min(Math.max(6, pos.value.y), Math.max(6, window.innerHeight - 70))
    }
  }
  updateRibbon()
}

let clockTimer: number | undefined

onMounted(() => {
  window.addEventListener('resize', onResize)
  clockTimer = window.setInterval(tickClock, 20000)
  tickClock()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', onResize)
  window.removeEventListener('keydown', onKeydown)
  onBarUp()
  if (clockTimer) window.clearInterval(clockTimer)
  if (toastTimer) window.clearTimeout(toastTimer)
  if (autoTimer) window.clearTimeout(autoTimer)
  pager.stop()
})

defineExpose({ modeLabel })
</script>
