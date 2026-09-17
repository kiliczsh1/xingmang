<template>
  <div class="scan-page">
    <div class="scanline" aria-hidden="true"></div>

    <header class="scan-header">
      <div>
        <div class="logo">AI_SCAN <span class="logo-slash">//</span> 扫榜终端</div>
        <h1>榜 <span class="accent">·</span> 扫描中</h1>
      </div>
      <div class="status" :class="{ demo: !!error }">
        <span class="dot"></span>
        {{ error ? '连接失败' : '番茄数据 · 实时' }}
      </div>
    </header>

    <div class="radar-bar">
      <div class="radar" aria-hidden="true"></div>
      <div class="r-info">
        <div class="r-line">当前目标：<b>{{ activeGroupLabel }} · {{ activeGenreName }}</b> · 命中 {{ filtered.length }} 项</div>
        <div class="r-prog" aria-hidden="true">
          <i v-for="n in 10" :key="n" :class="{ on: n <= progress }"></i>
        </div>
      </div>
      <div class="r-nums">{{ String(filtered.length).padStart(4, '0') }} / {{ String(books.length).padStart(4, '0') }}</div>
    </div>

    <div class="filters">
      <button v-for="(g, gi) in groups" :key="g.title" class="chip" :class="{ on: gi === activeGroupIndex }" @click="switchGroup(gi)">[{{ g.title }}]</button>
      <div class="spacer"></div>
      <div class="search">
        <input v-model.trim="keyword" type="text" placeholder="INPUT 关键词…" />
        <button class="run" @click="keyword = ''">RUN</button>
      </div>
    </div>

    <div v-if="genres.length" class="genre-chips">
      <button v-for="g in genres" :key="g.id" class="gchip" :class="{ on: g.id === activeGenre }" @click="switchGenre(g.id)">
        {{ g.name }}
      </button>
    </div>

    <div class="scope">TARGET：<b>{{ activeGroupLabel }} · {{ activeGenreName }}</b> · 命中 <b class="hit">{{ filtered.length }}</b></div>

    <div class="scan-scroll">
    <main class="grid">
      <template v-if="loading">
        <div v-for="n in skeletons" :key="n" class="hud skel">
          <div class="skel-cover"></div>
          <div class="skel-body">
            <div class="skel-row">
              <div class="skel-line w10"></div>
              <div class="skel-line w16"></div>
            </div>
            <div class="skel-line w70"></div>
            <div class="skel-line w40"></div>
            <div class="skel-line w90"></div>
            <div class="skel-line w60"></div>
            <div class="skel-line w80"></div>
          </div>
        </div>
      </template>

      <template v-else>
        <article v-for="(b, i) in filtered" :key="b.bookId" class="hud">
          <div class="cover">
            <div class="cover-fallback" :style="{ background: coverGradient(b.title) }"><span>{{ (b.title || '?')[0] }}</span></div>
            <img v-if="b.coverUrl" :src="b.coverUrl" :alt="b.title" loading="lazy" referrerpolicy="no-referrer" @error="onCoverError" />
          </div>
          <div class="info">
            <div class="hd">
              <span class="rank" :class="rankCls(i)">{{ String(b.rank || i + 1).padStart(2, '0') }}</span>
              <span class="tag" :class="{ warn: b.status === '已完结' }">{{ b.status || '实时' }}</span>
            </div>
            <div class="name" :title="b.title">{{ b.title }}</div>
            <div class="auth">{{ b.author }} · {{ genreName(b.category) }}</div>
            <div v-if="b.synopsis" class="synopsis">{{ b.synopsis }}</div>
            <div class="data">
              <span v-if="b.readingCount">在读<b>{{ b.readingCount }}</b></span>
              <span v-if="b.wordCount">字数<b>{{ b.wordCount }}</b></span>
              <span>变动<b v-html="diffHtml(b.rankDiff)"></b></span>
            </div>
            <div class="ops">
              <a :href="b.link" target="_blank" rel="noopener">[ 阅读原文 ]</a>
              <button @click="handleAnalyze(b)">AI 拆书</button>
            </div>
          </div>
        </article>

        <div v-if="error && !books.length" class="empty">
          <div class="empty-icon" aria-hidden="true">⚠</div>
          {{ error }}<br />
          <span class="hint">请先启动 Python 桥：<code>python scripts/fanqie_bridge.py</code></span>
        </div>
        <div v-else-if="!filtered.length" class="empty">
          <div class="empty-icon" aria-hidden="true">◉</div>
          NO MATCH · 信号丢失<br />该榜单/题材暂无数据，换个题材或关键词重新扫描
        </div>
      </template>
    </main>

    <div v-if="books.length >= size || page > 0" class="pager">
      <button :disabled="page <= 0" @click="prevPage">‹ 上一页</button>
      <span>第 {{ page + 1 }} 页 · 每页 {{ size }} 本</span>
      <button :disabled="books.length < size" @click="nextPage">下一页 ›</button>
    </div>

    <footer class="scan-footer">
      <span>UPLINK：<span class="ok-blink">●</span> ONLINE · 数据来自番茄榜单（a_bogus + PUA 解码）</span>
      <span>LATENCY 12ms · CACHE HIT 98%</span>
    </footer>
    </div>

    <!-- AI 拆书分析弹窗 -->
    <el-dialog
      v-model="analysisDialogVisible"
      :show-close="false"
      width="780px"
      top="5vh"
      append-to-body
      destroy-on-close
      class="analysis-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <div class="dialog-glyph" aria-hidden="true">✦</div>
          <div class="dialog-title-row">
            <h2 class="dialog-title">
              《{{ currentBook?.title ?? '—' }}》
              <span class="dialog-status" :class="{ done: currentBook?.status === '已完结' }">
                {{ currentBook?.status || '连载中' }}
              </span>
            </h2>
          </div>
          <button class="dialog-close" title="关闭" @click="analysisDialogVisible = false">✕</button>
        </div>
      </template>

      <div class="dialog-book-info">
        <div class="dialog-cover">
          <div class="cover-fallback" :style="{ background: coverGradient(currentBook?.title || '') }">
            <span>{{ (currentBook?.title || '?')[0] }}</span>
          </div>
          <img
            v-if="currentBook?.coverUrl"
            :src="currentBook.coverUrl"
            :alt="currentBook?.title"
            referrerpolicy="no-referrer"
            @error="onCoverError"
          />
        </div>
        <div class="dialog-meta">
          <div class="meta-line">
            <span class="lab">AUTHOR&nbsp;</span>
            <b>{{ currentBook?.author || '—' }}</b>
            <span class="sep">·</span>
            <b>{{ genreName(currentBook?.category || '') }}</b>
          </div>
          <div class="meta-line">
            <span class="lab">RANK&nbsp;</span>
            <b>都市高武 · 男频阅读榜 · 第{{ currentBook?.rank ?? '—' }}名</b>
          </div>
          <div class="meta-line">
            <span class="lab">DATA&nbsp;</span>
            <span v-if="currentBook?.readingCount"><b>{{ currentBook.readingCount }}</b>人在读</span>
            <span v-if="currentBook?.wordCount"> · <b>{{ currentBook.wordCount }}</b>字</span>
          </div>
        </div>
      </div>

      <div class="glass-chips">
        <span class="chip"><b>榜</b>男频阅读榜</span>
        <span class="chip"><b>名次</b>第 {{ currentBook?.rank ?? '—' }} 名</span>
        <span v-if="currentBook?.readingCount" class="chip"><b>在读</b>{{ currentBook.readingCount }}</span>
        <span v-if="currentBook?.wordCount" class="chip"><b>字数</b>{{ currentBook.wordCount }}</span>
      </div>

      <div v-if="currentBook?.synopsis" class="meta-synopsis">{{ currentBook.synopsis }}</div>

      <div class="dialog-analysis-section">
        <h3 class="analysis-title"><span class="orb" aria-hidden="true"></span>AI 深度分析</h3>
        <div class="analysis-actions">
          <button class="action-btn" @click="handleAction('title')">
            <span class="action-icon">✏️</span>AI 标题仿写
          </button>
          <button class="action-btn" @click="handleAction('synopsis')">
            <span class="action-icon">📝</span>AI 简介仿写
          </button>
          <button class="action-btn" @click="handleAction('brainstorm')">
            <span class="action-icon">💡</span>AI 脑洞仿写
          </button>
          <button class="action-btn primary" @click="handleAction('analyze')">
            <span class="action-icon">🔍</span>AI 一键拆书
          </button>
        </div>
      </div>
    </el-dialog>

    <!-- 生成器式 AI 分析弹窗（4 个 AI 按钮共用） -->
    <AiAnalyzeDialog
      v-model="analysisActionVisible"
      :book="currentBook"
      :action-type="analysisAction"
      :category-label="currentBook ? genreName(currentBook.category || '') : ''"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { rankAPI, type RankBook, type RankGroup } from '@/api'
import AiAnalyzeDialog from '@/components/WriteDialogs/AiAnalyzeDialog.vue'

/**
 * AI 扫榜 —— 赛博扫描风（demo-10 方向）
 * 数据：真实番茄榜单。Go 后端转调本机 Python 桥（scripts/fanqie_bridge.py，a_bogus 签名 + PUA 字体解码）。
 * 已删除演示样本：桥不可用时如实显示错误态，不做假数据回退。
 */

const groups = ref<RankGroup[]>([])
const activeGroupIndex = ref(0)
const activeGenre = ref('') // item.id，如 "1_2_1141" = gender_rankMold_categoryId
const books = ref<RankBook[]>([])
const keyword = ref('')
const loading = ref(false)
const error = ref('')
const page = ref(0)
const size = 30
const analysisDialogVisible = ref(false)
const currentBook = ref<RankBook | null>(null)
// 生成器式 AI 分析弹窗（4 个按钮共用，按 actionType 区分）
const analysisAction = ref<'title' | 'synopsis' | 'brainstorm' | 'analyze'>('analyze')
const analysisActionVisible = ref(false)

const activeGroup = computed(() => groups.value[activeGroupIndex.value])
const genres = computed(() => activeGroup.value?.items ?? [])
const activeGroupLabel = computed(() => activeGroup.value?.title ?? '')
const activeGenreName = computed(() => genres.value.find(g => g.id === activeGenre.value)?.name || activeGenre.value || '—')

const filtered = computed(() => {
  let list = books.value
  const kw = keyword.value.trim()
  if (kw) list = list.filter(b => b.title.includes(kw) || b.author.includes(kw))
  return list
})
const progress = computed(() => Math.max(1, Math.round((filtered.value.length / (books.value.length || 1)) * 10)))
// 加载骨架屏数量（最多 6 张占位卡）
const skeletons = computed(() => Array(Math.min(6, size)).fill(0).map((_, i) => i))

const genreName = (categoryId: string) => {
  const g = activeGroup.value?.items.find(i => i.id.split('_')[2] === categoryId)
  return g?.name || categoryId
}
const coverGradient = (title: string) => {
  let h = 0
  for (const c of title || '') h = (h * 31 + c.codePointAt(0)!) % 360
  return `linear-gradient(150deg,hsl(${h},70%,52%),hsl(${(h + 40) % 360},65%,36%))`
}
const rankCls = (i: number) => (i < 1 ? 'r1' : i < 2 ? 'r2' : i < 3 ? 'r3' : '')
const diffHtml = (d: number) => {
  if (d > 0) return `<span class="up">▲${d}</span>`
  if (d < 0) return `<span class="down">▼${-d}</span>`
  return '<span>—</span>'
}

const loadBooks = async () => {
  if (!activeGenre.value) return
  loading.value = true
  error.value = ''
  try {
    const res = await rankAPI.getBooks({ tab: activeGenre.value, page: page.value, size })
    books.value = res.data?.books ?? []
  } catch {
    books.value = []
    error.value = '番茄榜单不可用：后端无法连接 Python 桥'
  } finally {
    loading.value = false
  }
}

const loadListsAndBooks = async () => {
  try {
    const res = await rankAPI.getLists()
    groups.value = res.data?.groups ?? []
  } catch {
    error.value = '番茄榜单不可用：后端无法连接 Python 桥'
    return
  }
  if (!groups.value.length) {
    error.value = '番茄榜单未返回任何分组'
    return
  }
  activeGroupIndex.value = 0
  activeGenre.value = groups.value[0].items[0]?.id || ''
  await loadBooks()
}

const switchGroup = (gi: number) => {
  activeGroupIndex.value = gi
  activeGenre.value = groups.value[gi]?.items[0]?.id || ''
  page.value = 0
  loadBooks()
}
const switchGenre = (id: string) => {
  activeGenre.value = id
  page.value = 0
  loadBooks()
}
const prevPage = () => {
  if (page.value <= 0) return
  page.value--
  loadBooks()
}
const nextPage = () => {
  if (books.value.length < size) return
  page.value++
  loadBooks()
}

const handleAnalyze = (b: RankBook) => {
  currentBook.value = b
  analysisDialogVisible.value = true
}
const handleAction = (type: 'title' | 'synopsis' | 'brainstorm' | 'analyze') => {
  analysisAction.value = type
  analysisActionVisible.value = true
}
const onCoverError = (e: Event) => {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
}

onMounted(loadListsAndBooks)
</script>

<style scoped>
.scan-page {
  --bg: #f2f8fa;
  --panel: #ffffff;
  --panel-2: rgba(240, 248, 251, 0.85);
  --line: rgba(8, 145, 178, 0.2);
  --ink: #12313c;
  --muted: #5f8794;
  --faint: #9cb6c0;
  --cy: #0891b2;
  --mg: #d946ef;
  --ok: #059669;
  --warn: #d97706;
  --gridline: rgba(8, 145, 178, 0.06);
  --scan: rgba(8, 145, 178, 0.08);
  position: relative;
  padding: 4px 6px 12px;
  color: var(--ink);
  font-family: 'SF Mono', 'Consolas', ui-monospace, 'PingFang SC', 'Microsoft YaHei', monospace;
  font-size: 13px;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
:root[data-theme='dark'] .scan-page {
  --bg: #05070d;
  --panel: rgba(12, 20, 34, 0.82);
  --panel-2: rgba(10, 18, 30, 0.7);
  --line: rgba(56, 189, 248, 0.16);
  --ink: #d8f6ff;
  --muted: #6fa8bd;
  --faint: #4c6b7c;
  --cy: #38bdf8;
  --mg: #e879f9;
  --ok: #3b82f6;
  --warn: #fbbf24;
  --gridline: rgba(56, 189, 248, 0.05);
  --scan: rgba(56, 189, 248, 0.07);
}
.scan-page::before {
  content: '';
  position: fixed;
  inset: 0;
  z-index: -1;
  pointer-events: none;
  background-image: linear-gradient(var(--gridline) 1px, transparent 1px),
    linear-gradient(90deg, var(--gridline) 1px, transparent 1px);
  background-size: 40px 40px;
}
.scanline {
  position: fixed;
  left: 0;
  right: 0;
  height: 120px;
  z-index: -1;
  pointer-events: none;
  background: linear-gradient(180deg, transparent, var(--scan), rgba(0, 0, 0, 0), transparent);
  animation: scan 5s linear infinite;
}
@keyframes scan {
  0% { top: -140px; }
  100% { top: 110vh; }
}

.scan-header {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}
.logo {
  font-size: 13px;
  letter-spacing: 0.3em;
  color: var(--muted);
}
.logo-slash {
  color: var(--cy);
  letter-spacing: 0.04em;
}
.scan-header h1 {
  font-size: 30px;
  font-weight: 800;
  letter-spacing: 0.02em;
  margin-top: 2px;
}
.scan-header .accent {
  color: var(--cy);
}
.status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--ok);
  border: 1px solid color-mix(in srgb, var(--ok) 35%, transparent);
  padding: 5px 14px;
  border-radius: 6px;
  background: color-mix(in srgb, var(--ok) 6%, transparent);
}
.status.demo {
  color: var(--warn);
  border-color: color-mix(in srgb, var(--warn) 35%, transparent);
  background: color-mix(in srgb, var(--warn) 7%, transparent);
}
.status .dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 10px currentColor;
  animation: blink 1.2s infinite;
}
@keyframes blink {
  50% { opacity: 0.3; }
}

.radar-bar {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}
.radar {
  position: relative;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  flex: none;
  background: radial-gradient(circle, color-mix(in srgb, var(--cy) 12%, transparent), transparent 65%);
  border: 1px solid color-mix(in srgb, var(--cy) 40%, transparent);
}
.radar::before {
  content: '';
  position: absolute;
  inset: 6px;
  border-radius: 50%;
  border: 1px dashed color-mix(in srgb, var(--cy) 30%, transparent);
}
.radar::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: conic-gradient(from 0deg, color-mix(in srgb, var(--cy) 55%, transparent), transparent 70deg);
  animation: spin 2.2s linear infinite;
  -webkit-mask: radial-gradient(circle, transparent 38%, #000 42%);
  mask: radial-gradient(circle, transparent 38%, #000 42%);
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
.r-info {
  flex: 1;
  min-width: 0;
}
.r-line {
  font-size: 12px;
  color: var(--muted);
}
.r-line b {
  color: var(--cy);
  font-weight: 600;
}
.r-prog {
  display: flex;
  gap: 6px;
  margin-top: 6px;
}
.r-prog i {
  display: block;
  height: 4px;
  flex: 1;
  background: color-mix(in srgb, var(--cy) 15%, transparent);
  border-radius: 2px;
}
.r-prog i.on {
  background: var(--cy);
  box-shadow: 0 0 6px color-mix(in srgb, var(--cy) 80%, transparent);
}
.r-nums {
  font-size: 20px;
  font-weight: 700;
  color: var(--cy);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.06em;
}

.filters {
  flex-shrink: 0;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;
  align-items: center;
}
.chip {
  font-size: 12px;
  letter-spacing: 0.06em;
  padding: 6px 14px;
  color: var(--muted);
  border: 1px solid var(--line);
  background: var(--panel-2);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.chip:hover {
  color: var(--cy);
  border-color: var(--cy);
  box-shadow: inset 0 0 10px color-mix(in srgb, var(--cy) 20%, transparent);
}
.chip.on {
  color: #04141a;
  background: var(--cy);
  border-color: var(--cy);
  font-weight: 700;
  box-shadow: 0 0 14px color-mix(in srgb, var(--cy) 40%, transparent);
}
.spacer {
  flex: 1;
}
.search {
  display: flex;
  border: 1px solid var(--line);
  background: var(--panel-2);
  border-radius: 6px;
  overflow: hidden;
}
.search input {
  border: none;
  outline: none;
  background: none;
  color: var(--ink);
  padding: 7px 12px;
  font-size: 12px;
  width: 160px;
  letter-spacing: 0.04em;
  font-family: inherit;
}
.search .run {
  background: var(--mg);
  color: #16041a;
  font-weight: 700;
  padding: 0 14px;
  font-size: 12px;
  border: none;
  cursor: pointer;
  font-family: inherit;
}

.genre-chips {
  flex-shrink: 0;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
.gchip {
  font-size: 11px;
  letter-spacing: 0.04em;
  padding: 4px 10px;
  color: var(--muted);
  border: 1px solid var(--line);
  background: var(--panel-2);
  border-radius: 999px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.gchip:hover {
  color: var(--cy);
  border-color: var(--cy);
}
.gchip.on {
  color: #04141a;
  background: var(--cy);
  border-color: var(--cy);
  font-weight: 700;
}

.scope {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--muted);
  letter-spacing: 0.08em;
  margin-bottom: 12px;
}
.scope b {
  color: var(--mg);
  font-weight: 600;
}
.scope .hit {
  color: var(--cy);
}

/* 书籍网格（尺寸保持不变，随 scan-scroll 一起滚动） */
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 14px;
}

/* 滚动容器：网格 + 分页器 + 底部状态栏整体滚动，只固定顶端控制区 */
.scan-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 2px 4px 16px;
}
.hud {
  position: relative;
  display: flex;
  gap: 12px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 4px;
  padding: 12px 14px;
  overflow: hidden;
}
.hud::before,
.hud::after {
  content: '';
  position: absolute;
  width: 12px;
  height: 12px;
  border: 2px solid var(--cy);
  opacity: 0.6;
  pointer-events: none;
}
.hud::before {
  top: 5px;
  left: 5px;
  border-right: none;
  border-bottom: none;
}
.hud::after {
  bottom: 5px;
  right: 5px;
  border-left: none;
  border-top: none;
}

/* 加载骨架屏（切换题材时显示占位，加载完无缝换真实数据） */
.skel {
  pointer-events: none;
}
.skel-cover {
  width: 62px;
  height: 88px;
  flex: none;
  border-radius: 6px;
  background: color-mix(in srgb, var(--cy) 10%, transparent);
  animation: shimmer 1.4s ease-in-out infinite;
}
.skel-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 9px;
}
.skel-row {
  display: flex;
  justify-content: space-between;
  gap: 8px;
}
.skel-line {
  height: 11px;
  border-radius: 3px;
  background: color-mix(in srgb, var(--cy) 10%, transparent);
  animation: shimmer 1.4s ease-in-out infinite;
}
.w10 { width: 10%; }
.w16 { width: 16%; }
.w40 { width: 40%; }
.w60 { width: 60%; }
.w70 { width: 70%; }
.w80 { width: 80%; }
.w90 { width: 90%; }
@keyframes shimmer {
  0%, 100% { opacity: 0.45; }
  50% { opacity: 1; }
}

.cover {
  position: relative;
  width: 62px;
  height: 88px;
  flex: none;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.25);
}
.cover-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}
.cover-fallback span {
  font-size: 30px;
  font-weight: 900;
  color: rgba(255, 255, 255, 0.95);
  padding-bottom: 6px;
  text-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
}
.cover img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.hd {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 5px;
}
.rank {
  font-size: 20px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  color: var(--faint);
  letter-spacing: 0.04em;
}
.rank.r1 {
  color: var(--cy);
  text-shadow: 0 0 10px color-mix(in srgb, var(--cy) 80%, transparent), 0 0 26px color-mix(in srgb, var(--cy) 40%, transparent);
}
.rank.r2 {
  color: var(--mg);
  text-shadow: 0 0 10px color-mix(in srgb, var(--mg) 70%, transparent);
}
.rank.r3 {
  color: var(--ok);
  text-shadow: 0 0 10px color-mix(in srgb, var(--ok) 70%, transparent);
}
.tag {
  font-size: 10px;
  letter-spacing: 0.12em;
  color: var(--cy);
  border: 1px solid color-mix(in srgb, var(--cy) 40%, transparent);
  padding: 1px 7px;
  border-radius: 2px;
  background: color-mix(in srgb, var(--cy) 7%, transparent);
  white-space: nowrap;
}
.tag.warn {
  color: var(--warn);
  border-color: color-mix(in srgb, var(--warn) 40%, transparent);
  background: color-mix(in srgb, var(--warn) 7%, transparent);
}
.name {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.auth {
  font-size: 11px;
  color: var(--muted);
  margin: 3px 0 4px;
  letter-spacing: 0.03em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.synopsis {
  font-size: 11px;
  color: var(--muted);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 6px;
}
.data {
  display: flex;
  gap: 14px;
  font-size: 11px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
  margin-bottom: 6px;
}
.data b {
  display: block;
  font-size: 13px;
  color: var(--ink);
  font-weight: 600;
}
.data .up {
  color: var(--ok);
}
.data .down {
  color: var(--warn);
}
.ops {
  margin-top: auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px dashed color-mix(in srgb, var(--cy) 18%, transparent);
}
.ops a {
  font-size: 11.5px;
  color: var(--cy);
  text-decoration: none;
  letter-spacing: 0.05em;
}
.ops a:hover {
  text-shadow: 0 0 8px color-mix(in srgb, var(--cy) 80%, transparent);
}
.ops button {
  font-size: 11.5px;
  color: #04141a;
  background: var(--cy);
  padding: 3px 11px;
  border-radius: 3px;
  border: none;
  font-weight: 700;
  letter-spacing: 0.04em;
  cursor: pointer;
  box-shadow: 0 0 10px color-mix(in srgb, var(--cy) 35%, transparent);
  font-family: inherit;
}

.empty {
  grid-column: 1 / -1;
  text-align: center;
  padding: 70px 0;
  color: var(--muted);
  letter-spacing: 0.1em;
  line-height: 1.8;
}
.empty .hint {
  font-size: 11px;
  color: var(--faint);
  letter-spacing: 0.05em;
}
.empty code {
  color: var(--cy);
  background: color-mix(in srgb, var(--cy) 8%, transparent);
  padding: 1px 6px;
  border-radius: 4px;
}
.empty-icon {
  font-size: 38px;
  margin-bottom: 10px;
  color: var(--cy);
  text-shadow: 0 0 14px color-mix(in srgb, var(--cy) 50%, transparent);
}

.pager {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-top: 20px;
}
.pager button {
  font-size: 12px;
  letter-spacing: 0.08em;
  color: var(--cy);
  border: 1px solid var(--line);
  background: var(--panel-2);
  padding: 6px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-family: inherit;
}
.pager button:hover:not(:disabled) {
  border-color: var(--cy);
  box-shadow: 0 0 10px color-mix(in srgb, var(--cy) 30%, transparent);
}
.pager button:disabled {
  opacity: 0.35;
  cursor: default;
}
.pager span {
  font-size: 12px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.06em;
}

.scan-footer {
  margin-top: 26px;
  font-size: 11px;
  color: var(--faint);
  letter-spacing: 0.14em;
  display: flex;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: 8px;
}
.ok-blink {
  color: var(--ok);
  animation: blink 1.2s infinite;
}

@media (max-width: 768px) {
  .grid {
    grid-template-columns: 1fr;
  }
  .search input {
    width: 110px;
  }
}

/* 动效降级 */
@media (prefers-reduced-motion: reduce) {
  .scanline,
  .radar::after,
  .ok-blink,
  .status .dot,
  .skel-cover,
  .skel-line {
    animation: none !important;
  }
  .scanline {
    display: none;
  }
  .skel-cover,
  .skel-line {
    opacity: 0.6;
  }
}

/* ============ AI 拆书分析弹窗（玻璃浮窗 Glass Holo / demo-4） ============ */
/* 元素级样式（teleport 后弹窗内容仍保留 data-v，scoped 可命中）；面板级/结构规则见文件底部非 scoped <style> 块 */
.dialog-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.dialog-glyph {
  flex: none;
  width: 34px;
  height: 34px;
  border-radius: 12px;
  background: var(--g-grad);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  color: #ffffff;
  box-shadow: 0 6px 18px var(--g-glow);
}
.dialog-title-row {
  flex: 1;
  min-width: 0;
}
.dialog-title {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 0.01em;
  color: var(--g-ink);
  display: flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dialog-status {
  flex: none;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: #b45309;
  border: 1px solid rgba(217, 119, 6, 0.45);
  background: rgba(245, 158, 11, 0.14);
  padding: 2px 9px;
  border-radius: 999px;
  white-space: nowrap;
}
.dialog-status.done {
  color: var(--g-ok);
  border-color: rgba(5, 150, 105, 0.35);
  background: rgba(5, 150, 105, 0.12);
}
.dialog-close {
  flex: none;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: 1px solid var(--g-line);
  background: transparent;
  color: var(--g-faint);
  font-size: 14px;
  line-height: 1;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  display: flex;
  align-items: center;
  justify-content: center;
}
.dialog-close:hover {
  background: var(--g-inner);
  color: var(--g-ink);
}

/* 书籍信息区 */
.dialog-book-info {
  display: flex;
  gap: 18px;
  align-items: center;
}
.dialog-cover {
  position: relative;
  width: 72px;
  height: 104px;
  flex: none;
  border-radius: 14px;
  overflow: hidden;
  box-shadow: 0 12px 30px rgba(10, 25, 47, 0.3), inset 0 1px 0 var(--g-inner);
}
.dialog-cover .cover-fallback {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.dialog-cover .cover-fallback span {
  font-size: 30px;
  font-weight: 900;
  color: rgba(255, 255, 255, 0.95);
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
}
.dialog-cover img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.dialog-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.meta-line {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 12.5px;
  color: var(--g-muted);
  line-height: 1.6;
}
.meta-line .lab {
  flex: none;
  font-size: 10.5px;
  letter-spacing: 1px;
  color: var(--g-faint);
}
.meta-line b {
  color: var(--g-ink);
  font-weight: 700;
}
.meta-line .sep {
  opacity: 0.5;
}

/* 玻璃信息胶囊 */
.glass-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 14px;
  margin-bottom: 14px;
}
.glass-chips .chip {
  font-size: 11.5px;
  font-weight: 600;
  color: var(--g-ink);
  background: var(--g-glass-strong);
  border: 1px solid var(--g-line);
  padding: 5px 12px;
  border-radius: 999px;
  box-shadow: 0 4px 12px rgba(10, 25, 47, 0.08), inset 0 1px 0 var(--g-inner);
}
.glass-chips .chip b {
  color: var(--g-vio);
  margin-right: 5px;
}
.meta-synopsis {
  font-size: 12.5px;
  color: var(--g-muted);
  line-height: 1.75;
  background: var(--g-glass-strong);
  border: 1px solid var(--g-line);
  border-radius: 14px;
  padding: 12px 15px;
  margin-bottom: 16px;
  box-shadow: inset 0 1px 0 var(--g-inner);
}

/* AI 深度分析区 */
.dialog-analysis-section {
  border-top: 1px solid var(--g-line);
  padding-top: 16px;
}
.analysis-title {
  margin: 0 0 12px;
  font-size: 14px;
  font-weight: 800;
  letter-spacing: 0.1em;
  color: var(--g-ink);
  display: flex;
  align-items: center;
  gap: 9px;
}
.analysis-title .orb {
  flex: none;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--g-grad);
  box-shadow: 0 0 12px var(--g-glow);
}
.analysis-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.action-btn {
  flex: 1;
  min-width: 130px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 13px 10px;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.06em;
  color: var(--g-ink);
  background: var(--g-glass-strong);
  border: 1px solid var(--g-line);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.18s;
  font-family: inherit;
  box-shadow: 0 6px 18px rgba(10, 25, 47, 0.1), inset 0 1px 0 var(--g-inner);
}
.action-btn:hover {
  transform: translateY(-3px);
  border-color: var(--g-cy);
  box-shadow: 0 12px 26px rgba(10, 25, 47, 0.16), inset 0 1px 0 var(--g-inner);
}
.action-btn:active {
  transform: translateY(-1px);
}
.action-btn .action-icon {
  font-size: 15px;
  line-height: 1;
}
.action-btn.primary {
  color: #ffffff;
  background: var(--g-grad);
  border-color: transparent;
  box-shadow: 0 12px 30px var(--g-glow);
}
.action-btn.primary:hover {
  filter: brightness(1.08);
}

@media (max-width: 768px) {
  .dialog-book-info {
    flex-direction: column;
    align-items: stretch;
  }
  .dialog-cover {
    width: 72px;
    height: 104px;
  }
  .action-btn {
    min-width: calc(50% - 6px);
  }
}
</style>

<style>
/* ============ AI 拆书分析弹窗（玻璃浮窗 Glass Holo / demo-4）· 面板级/结构规则（非 scoped） ============ */
/* 说明：el-dialog append-to-body 后，面板根元素没有 scope id；scoped 里 :global() 与 ::before / :root[data-theme] 联合会被 Vue 编译器剥掉伪元素与前缀，
   导致面板塌缩、暗色令牌全局生效。故面板级样式放入非 scoped 块，用 .analysis-dialog 精确限定，不影响其它弹窗。 */
.analysis-dialog {
  --g-glass: rgba(255, 255, 255, 0.52);
  --g-glass-strong: rgba(255, 255, 255, 0.68);
  --g-line: rgba(255, 255, 255, 0.65);
  --g-inner: rgba(255, 255, 255, 0.35);
  --g-ink: #1c2a3a;
  --g-muted: #54687c;
  --g-faint: #8ba0b2;
  --g-cy: #0284c7;
  --g-vio: #7c3aed;
  --g-ok: #059669;
  --g-grad: linear-gradient(135deg, #0284c7, #7c3aed);
  --g-glow: rgba(2, 132, 199, 0.35);
  position: relative;
  border-radius: 22px;
  background: var(--g-glass);
  border: 1px solid var(--g-line);
  backdrop-filter: blur(26px) saturate(150%);
  -webkit-backdrop-filter: blur(26px) saturate(150%);
  box-shadow: 0 8px 24px rgba(10, 25, 47, 0.14), 0 42px 90px rgba(10, 25, 47, 0.3),
    inset 0 1px 0 var(--g-inner);
  overflow: hidden;
}
/* 顶部高光线 */
.analysis-dialog::before {
  content: '';
  position: absolute;
  left: 8%;
  right: 8%;
  top: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--g-line), transparent);
  z-index: 2;
}
/* 全息光斑背景（仅此弹窗的遮罩层） */
.el-overlay:has(.analysis-dialog) {
  background-color: rgba(15, 23, 42, 0.45);
  background-image:
    radial-gradient(circle at 15% 20%, rgba(125, 211, 252, 0.3) 0%, transparent 42%),
    radial-gradient(circle at 85% 18%, rgba(167, 139, 250, 0.3) 0%, transparent 42%),
    radial-gradient(circle at 60% 92%, rgba(240, 171, 252, 0.24) 0%, transparent 42%);
}
.analysis-dialog .el-dialog__header {
  margin: 0;
  padding: 18px 22px 14px;
  border-bottom: 1px solid var(--g-line);
}
.analysis-dialog .el-dialog__body {
  padding: 16px 22px 20px;
}
/* 深色模式适配（玻璃霓虹） */
:root[data-theme='dark'] .analysis-dialog {
  --g-glass: rgba(10, 18, 34, 0.52);
  --g-glass-strong: rgba(14, 24, 44, 0.68);
  --g-line: rgba(0, 229, 255, 0.22);
  --g-inner: rgba(0, 229, 255, 0.15);
  --g-ink: #e8eef7;
  --g-muted: #9fb2c8;
  --g-faint: #5d7189;
  --g-cy: #00e5ff;
  --g-vio: #8b7bff;
  --g-ok: #34d399;
  --g-grad: linear-gradient(135deg, #00e5ff 0%, #3b82f6 55%, #7c6bff 100%);
  --g-glow: rgba(0, 229, 255, 0.4);
  border-color: var(--g-line);
}
:root[data-theme='dark'] .analysis-dialog .dialog-status {
  color: #fbbf24;
  border-color: rgba(251, 191, 36, 0.4);
  background: rgba(251, 191, 36, 0.12);
}
:root[data-theme='dark'] .analysis-dialog .dialog-status.done {
  color: var(--g-ok);
  border-color: rgba(52, 211, 153, 0.4);
  background: rgba(52, 211, 153, 0.12);
}
@media (max-width: 768px) {
  .analysis-dialog {
    width: 95vw !important;
  }
}
</style>
