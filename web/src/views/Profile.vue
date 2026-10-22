<template>
  <div class="profile-container" ref="rootRef" v-loading="loading">
    <!-- 创作仪表盘：作品数据总览 -->
    <header class="dash-brand">
      <div class="brand-left">
        <div class="brand-title">
          <img class="dash-mark" src="/brand/mark-48.png" width="24" height="24" alt="">
          <h2>创作仪表盘</h2>
        </div>
        <div class="sub">作品全景 · 一屏掌握</div>
      </div>
      <!-- 版本自查：前端产物内嵌在后端 exe 里，只比界面看不出是否已更新 -->
      <div class="brand-ver" :title="appVersionTitle">版本 {{ versionLabel }}</div>
    </header>
    <div class="stats">
      <div class="stat s1">
        <div class="dot">📚</div>
        <span class="s-label">作品总数</span>
        <span class="s-value">{{ statTotal }} <small>部</small></span>
        <div class="bar"><i :style="{ width: pctOf('all') + '%' }"></i></div>
        <span class="s-foot">全部作品</span>
      </div>
      <div class="stat s2">
        <div class="dot">✍️</div>
        <span class="s-label">创作中</span>
        <span class="s-value">{{ statDraft }} <small>部</small></span>
        <div class="bar"><i :style="{ width: pctOf('draft') + '%' }"></i></div>
        <span class="s-foot">草稿 · 待续写</span>
      </div>
      <div class="stat s3">
        <div class="dot">📄</div>
        <span class="s-label">已发布</span>
        <span class="s-value">{{ statPublished }} <small>部</small></span>
        <div class="bar"><i :style="{ width: pctOf('published') + '%' }"></i></div>
        <span class="s-foot">对外可见</span>
      </div>
      <div class="stat s4">
        <div class="dot">🏁</div>
        <span class="s-label">已完结</span>
        <span class="s-value">{{ statCompleted }} <small>部</small></span>
        <div class="bar"><i :style="{ width: pctOf('completed') + '%' }"></i></div>
        <span class="s-foot">完整的故事</span>
      </div>
    </div>

    <!-- 顶栏：标题 + 三组标签 -->
    <div class="topbar">
      <div>
        <h2 class="page-title">个人中心</h2>
        <div class="sub">三组标签 · 概览 / 时间轴故事 / 热力网格 <span v-if="isDemo" class="demo-tag">演示数据</span></div>
      </div>
      <div class="seg">
        <button :class="{ active: tab === 'ov' }" @click="setTab('ov')">🗂 概览</button>
        <button :class="{ active: tab === 'timeline' }" @click="setTab('timeline')">📖 时间轴故事</button>
        <button :class="{ active: tab === 'heatmap' }" @click="setTab('heatmap')">🔥 热力网格</button>
      </div>
    </div>

    <!-- ============ 标签页 1：概览 ============ -->
    <div v-show="tab === 'ov'" class="view">
      <div class="ov-grid">
        <div class="ov-card" v-for="c in statCards" :key="c.key">
          <div class="o-ico" :style="{ background: c.bg }">{{ c.icon }}</div>
          <div class="o-val">{{ c.value.toLocaleString() }}</div>
          <div class="o-lbl">{{ c.label }}</div>
          <div class="o-delta" v-if="c.key === 'usage' && usageDelta" :class="usageDelta.up ? 'up' : 'down'">{{ usageDelta.text }}</div>
        </div>
      </div>

      <div class="ov-bottom">
        <div class="ov-panel">
          <h3>🏆 模型排行</h3>
          <div v-if="modelStats.length" class="ov-rows">
            <div class="ov-row" v-for="(m, i) in modelStats.slice(0, 5)" :key="m.model_id">
              <span class="n">{{ m.model_name }}</span>
              <div class="tr"><div class="fl" :data-w="pct(m.total_usage, maxModelUsage)" :style="{ background: modelColor(i) }"></div></div>
              <span class="v">{{ m.total_usage.toLocaleString() }}</span>
            </div>
          </div>
          <div v-else class="empty">暂无数据</div>
        </div>
        <div class="ov-panel">
          <h3>📈 月度趋势</h3>
          <div v-if="monthlyStats.length" class="ov-rows">
            <div class="ov-row" v-for="m in monthlyStats" :key="m.month">
              <span class="n">{{ m.month }}月</span>
              <div class="tr"><div class="fl" :data-w="pct(m.total_usage, maxMonthlyUsage)"></div></div>
              <span class="v">{{ m.total_usage.toLocaleString() }}</span>
            </div>
          </div>
          <div v-else class="empty">暂无数据</div>
        </div>
      </div>

      <!-- 领奖台 -->
      <div class="podium-card" v-if="podiumOrder.length">
        <h3>🏆 领奖台</h3>
        <div class="podium">
          <div class="slot" :class="slotCls(m.rank)" v-for="m in podiumOrder" :key="m.model_id">
            <div class="crown" v-if="m.rank === 1">👑</div>
            <div class="avatar">{{ modelIcon(m) }}</div>
            <div class="p-name">{{ m.model_name }}</div>
            <div class="p-provider">{{ m.provider_name }}</div>
            <div class="p-count">{{ m.total_usage.toLocaleString() }} 次</div>
            <div class="step" :class="stepCls(m.rank)" :data-h="podiumH(m.rank)">{{ m.rank }}</div>
          </div>
        </div>
        <div class="rest" v-if="restModels.length">
          <div class="rest-row" v-for="m in restModels" :key="m.model_id">
            <div class="idx">{{ m.rank }}</div>
            <div class="name">{{ m.model_name }}<div class="provider">{{ m.provider_name }}</div></div>
            <div class="bar"><div class="fill" :data-w="pct(m.total_usage, maxModelUsage)"></div></div>
            <div class="cnt">{{ m.total_usage.toLocaleString() }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- ============ 标签页 2：时间轴故事 ============ -->
    <div v-show="tab === 'timeline'" class="view">
      <div class="prologue">
        <div class="kicker">MY USAGE STORY</div>
        <h2>从 0 到 {{ totalUsage.toLocaleString() }} 次调用</h2>
        <p>滚动翻阅你的使用历程 —— 每一段都是与 AI 协作的足迹。</p>
      </div>
      <div v-if="timelineData.length" class="timeline">
        <div class="tl-item" :class="{ milestone: m.milestone }" v-for="m in timelineData" :key="m.month">
          <div class="tl-dot"></div>
          <div class="tl-card">
            <div class="tl-month">{{ m.month }}月</div>
            <div class="tl-count">{{ m.total_usage.toLocaleString() }}<small> 次</small></div>
            <div class="tl-bar"><div class="fill" :data-w="pct(m.total_usage, maxMonthlyUsage)"></div></div>
            <div class="tl-note"><span class="tag" v-if="m.milestone">★ 里程碑</span>{{ m.note }}</div>
          </div>
        </div>
      </div>
      <div v-else class="empty" style="padding: 60px 0;">暂无月度数据</div>
    </div>

    <!-- ============ 标签页 3：热力网格 ============ -->
    <div v-show="tab === 'heatmap'" class="view" :class="{ on: heatOn }">
      <div class="card">
        <div class="card-head">
          <h2>📅 近 20 周使用分布</h2>
          <div class="stat">过去一年共 <b>{{ totalUsage.toLocaleString() }}</b> 次调用</div>
        </div>
        <div class="heat-wrap" v-if="heatCols.length">
          <div class="heat">
            <div class="heat-col" v-for="(col, wi) in heatCols" :key="wi">
              <div class="cell" :class="'l' + c.lvl" v-for="(c, di) in col" :key="di" :style="{ transitionDelay: (wi * 7 + di) * 2.5 + 'ms' }">
                <span class="tip">{{ c.date }} · {{ c.cnt }} 次</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty" style="padding: 40px 0;">暂无数据</div>
        <div class="heat-meta" v-if="heatCols.length">
          <span>周一</span>
          <div class="legend">少 <span class="sw"></span><span class="sw l1"></span><span class="sw l2"></span><span class="sw l3"></span><span class="sw l4"></span> 多</div>
        </div>
      </div>
      <div class="insights" v-if="heatCols.length">
        <div class="insight">
          <div class="i-ico">🔥</div>
          <div class="i-title">最活跃时段</div>
          <div class="i-text">{{ insights.busy }}</div>
        </div>
        <div class="insight">
          <div class="i-ico">😴</div>
          <div class="i-title">低谷时段</div>
          <div class="i-text">{{ insights.quiet }}</div>
        </div>
        <div class="insight">
          <div class="i-ico">📈</div>
          <div class="i-title">增长趋势</div>
          <div class="i-text">{{ insights.trend }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { statsAPI, promptAPI, generatorAPI, bookAPI, systemAPI } from '@/api'
import type { UsageOverview, ModelStats, MonthlyStats, Book } from '@/types'

const overview = ref<UsageOverview | null>(null)
const modelStats = ref<ModelStats[]>([])
const monthlyStats = ref<MonthlyStats[]>([])
const bookList = ref<Book[]>([])
const loading = ref(false)
const isDemo = ref(false)

// 当前后端版本（与 MainLayout 侧边栏同一数据源，供用户自查是否已更新）
const appVersion = ref('')
const appVersionTitle = ref('')
const versionLabel = computed(() => appVersion.value || '未知')

const loadAppVersion = async () => {
  try {
    const health = await systemAPI.getHealth()
    const v = (health?.version || '').trim()
    const c = (health?.commit || '').trim()
    if (!v) {
      appVersion.value = ''
    } else if (v === 'dev') {
      appVersion.value = c && c !== 'unknown' ? `dev(${c.slice(0, 7)})` : 'dev'
    } else {
      appVersion.value = v.startsWith('v') ? v : `v${v}`
    }
    appVersionTitle.value = [
      `commit: ${c || 'unknown'}`,
      `构建时间: ${health?.buildTime || 'unknown'}`
    ].join('\n')
  } catch {
    appVersion.value = ''
  }
}

const tab = ref<'ov' | 'timeline' | 'heatmap'>('ov')
const heatOn = ref(false)
const rootRef = ref<HTMLElement | null>(null)

let io: IntersectionObserver | null = null

/* ---------- 数据归一化：后端实际返回 usage_count / 无 model_id，映射为前端类型 ---------- */
const normalizeModelStats = (list: any[]): ModelStats[] =>
  (list || []).map((m, i) => ({
    model_id: m.model_id ?? i,
    model_name: m.model_name || '未知模型',
    provider_name: m.provider_name || '未知',
    total_usage: m.total_usage ?? m.usage_count ?? 0,
    total_tokens: m.total_tokens ?? 0,
  }))

const monthLabel = (m: string) => {
  const s = String(m)
  const last = s.split('-').pop() || s
  return last.replace(/^0/, '')
}

const normalizeMonthly = (list: any[]): MonthlyStats[] =>
  (list || []).map(m => ({
    month: monthLabel(m.month),
    total_usage: m.total_usage ?? m.usage_count ?? 0,
    total_tokens: m.total_tokens ?? 0,
    model_count: m.model_count ?? 0,
  }))

const normalizeOverview = (o: any): UsageOverview => ({
  promptCount: o.promptCount ?? 0,
  generatorCount: o.generatorCount ?? 0,
  bookCount: o.bookCount ?? 0,
  totalUsageCount: o.totalUsageCount ?? o.total_usage ?? 0,
  totalTokens: o.totalTokens ?? o.total_tokens ?? 0,
})

/* ---------- 概览数据 ---------- */
const totalUsage = computed(() => overview.value?.totalUsageCount || 0)

/* ---------- 创作仪表盘：作品统计 ---------- */
const statTotal = computed(() => bookList.value.length)
const statDraft = computed(() => bookList.value.filter(b => (b.status || 'draft') === 'draft').length)
const statPublished = computed(() => bookList.value.filter(b => b.status === 'published').length)
const statCompleted = computed(() => bookList.value.filter(b => b.status === 'completed').length)

const pctOf = (key: string) => {
  if (!statTotal.value) return 0
  const n = key === 'all'
    ? statTotal.value
    : key === 'draft'
      ? statDraft.value
      : key === 'published'
        ? statPublished.value
        : statCompleted.value
  return Math.round((n / statTotal.value) * 100)
}

const statCards = computed(() => [
  { key: 'prompts', icon: '💬', value: overview.value?.promptCount || 0, label: '提示词数量', bg: 'linear-gradient(135deg,#3b82f6,#6366f1)' },
  { key: 'generators', icon: '✨', value: overview.value?.generatorCount || 0, label: '生成器应用', bg: 'linear-gradient(135deg,#6366f1,#8b5cf6)' },
  { key: 'books', icon: '📚', value: overview.value?.bookCount || 0, label: '书本数量', bg: 'linear-gradient(135deg,#38bdf8,#3b82f6)' },
  { key: 'usage', icon: '📊', value: totalUsage.value, label: '总调用次数', bg: 'linear-gradient(135deg,#8b5cf6,#6366f1)' },
])

const usageDelta = computed(() => {
  const ms = monthlyStats.value
  if (ms.length < 2) return null
  const last = ms[ms.length - 1]
  const prev = ms[ms.length - 2]
  if (!prev.total_usage) return null
  const p = Math.round(((last.total_usage - prev.total_usage) / prev.total_usage) * 100)
  return { up: p >= 0, text: `${p >= 0 ? '▲' : '▼'} ${Math.abs(p)}% 较上月` }
})

const maxModelUsage = computed(() => modelStats.value[0]?.total_usage || 1)
const maxMonthlyUsage = computed(() => {
  if (!monthlyStats.value.length) return 1
  return Math.max(...monthlyStats.value.map(m => m.total_usage))
})

const modelColors = [
  'linear-gradient(90deg,#3b82f6,#6366f1)',
  'linear-gradient(90deg,#6366f1,#8b5cf6)',
  'linear-gradient(90deg,#38bdf8,#3b82f6)',
  'linear-gradient(90deg,#8b5cf6,#6366f1)',
  'linear-gradient(90deg,#a5b4fc,#818cf8)',
]
const modelColor = (i: number) => modelColors[i % modelColors.length]

const modelIcon = (m: ModelStats) => {
  const n = m.model_name.toLowerCase()
  if (n.includes('gpt')) return '🤖'
  if (n.includes('claude')) return '🟠'
  if (n.includes('deepseek')) return '🐋'
  if (n.includes('qwen')) return '🌊'
  if (n.includes('glm')) return '🧬'
  if (n.includes('kimi')) return '🌙'
  return '🧠'
}

/* ---------- 领奖台 ---------- */
const podiumOrder = computed(() => {
  const top = modelStats.value.slice(0, 3).map((m, i) => ({ ...m, rank: i + 1 }))
  const byRank = (r: number) => top.find(x => x.rank === r)
  return [byRank(2), byRank(3), byRank(1)].filter(Boolean) as (ModelStats & { rank: number })[]
})
const restModels = computed(() => modelStats.value.slice(3).map((m, i) => ({ ...m, rank: i + 4 })))
const slotCls = (rank: number) => (rank === 1 ? 'first' : rank === 2 ? 'second' : 'third')
const stepCls = (rank: number) => (rank === 1 ? 's1' : rank === 2 ? 's2' : 's3')
// 台阶高度用固定像素（百分比在 auto 高度的父级下无法解析，会塌缩成内容高度）
const podiumH = (rank: number) => (rank === 1 ? 180 : rank === 2 ? 135 : 90)

/* ---------- 时间轴 ---------- */
const timelineData = computed(() => {
  const ms = monthlyStats.value
  let cum = 0
  let firstOver1000 = true
  return ms.map(m => {
    const prevCum = cum
    cum += m.total_usage
    const cross1000 = prevCum < 1000 && cum >= 1000
    const singleOver1000 = firstOver1000 && m.total_usage > 1000
    if (m.total_usage > 1000) firstOver1000 = false
    const milestone = cross1000 || singleOver1000
    const note = `Token 用量 ${formatTokens(m.total_tokens)}${m.model_count ? ` · 使用 ${m.model_count} 种模型` : ''}`
    return { ...m, milestone, note }
  })
})

/* ---------- 热力图（由月度数据摊分生成日分布） ---------- */
const heatCols = computed(() => {
  const total = monthlyStats.value.reduce((s, m) => s + m.total_usage, 0)
  if (!total) return [] as { lvl: number; cnt: number; date: string }[][]
  const weeks = 20
  const days = weeks * 7
  // 种子随机，保证每次渲染一致
  let seed = 20260809
  const rand = () => {
    seed = (seed * 1664525 + 1013904223) % 4294967296
    return seed / 4294967296
  }
  const raw: number[] = []
  for (let i = 0; i < days; i++) {
    const weekend = i % 7 >= 5
    let v = rand()
    if (weekend) v *= 0.3
    else v = v * 0.6 + 0.4
    raw.push(v * (0.4 + 0.6 * (i / days))) // 越早越少，模拟增长
  }
  const sum = raw.reduce((s, v) => s + v, 0)
  const scale = total / sum
  const today = new Date()
  const cols: { lvl: number; cnt: number; date: string }[][] = []
  for (let w = 0; w < weeks; w++) {
    const col: { lvl: number; cnt: number; date: string }[] = []
    for (let d = 0; d < 7; d++) {
      const idx = w * 7 + d
      const cnt = Math.round(raw[idx] * scale)
      const dt = new Date(today)
      dt.setDate(dt.getDate() - (days - 1 - idx))
      const date = `${dt.getFullYear()}-${String(dt.getMonth() + 1).padStart(2, '0')}-${String(dt.getDate()).padStart(2, '0')}`
      col.push({ lvl: 0, cnt, date })
    }
    cols.push(col)
  }
  const maxCnt = Math.max(...cols.flat().map(c => c.cnt))
  cols.forEach(col => col.forEach(c => {
    const r = c.cnt / maxCnt
    c.lvl = r > 0.8 ? 4 : r > 0.55 ? 3 : r > 0.3 ? 2 : r > 0.1 ? 1 : 0
  }))
  return cols
})

const insights = computed(() => {
  const cols = heatCols.value
  if (!cols.length) return { busy: '暂无数据', quiet: '暂无数据', trend: '暂无数据' }
  const byDow = Array.from({ length: 7 }, () => ({ sum: 0, n: 0 }))
  cols.forEach(col => col.forEach((c, d) => { byDow[d].sum += c.cnt; byDow[d].n++ }))
  const avg = byDow.map(x => x.sum / x.n)
  const busy = avg.indexOf(Math.max(...avg))
  const quiet = avg.indexOf(Math.min(...avg))
  const names = ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
  const recent = cols.slice(-4).flat().reduce((s, c) => s + c.cnt, 0)
  const prev = cols.slice(-8, -4).flat().reduce((s, c) => s + c.cnt, 0)
  const p = prev ? Math.round(((recent - prev) / prev) * 100) : 0
  return {
    busy: `${names[busy]}使用最密集，平均每天 ${Math.round(avg[busy])} 次`,
    quiet: `${names[quiet]}使用最少，平均每天 ${Math.round(avg[quiet])} 次`,
    trend: `近 4 周${p >= 0 ? '上升' : '回落'} ${Math.abs(p)}%${p >= 0 ? '，习惯正在养成' : '，可以安排新任务'}`,
  }
})

/* ---------- 工具 ---------- */
const pct = (v: number, max: number) => (max ? Math.round((v / max) * 100) : 0)

const formatTokens = (tokens: number) => {
  if (tokens >= 1000000) return (tokens / 1000000).toFixed(2) + 'M'
  if (tokens >= 1000) return (tokens / 1000).toFixed(1) + 'K'
  return tokens.toString()
}

/* ---------- 标签切换与动画 ---------- */
const q = (sel: string) => Array.from(rootRef.value?.querySelectorAll(sel) || []) as HTMLElement[]

function setTab(t: 'ov' | 'timeline' | 'heatmap') {
  tab.value = t
  if (t === 'ov') nextTick(activateOv)
  if (t === 'timeline') nextTick(activateTimeline)
  if (t === 'heatmap') nextTick(activateHeatmap)
}

function activateOv() {
  requestAnimationFrame(() => requestAnimationFrame(() => {
    q('.ov-row .fl').forEach(f => { f.style.width = f.dataset.w + '%' })
    q('.podium .step').forEach(s => { s.style.height = s.dataset.h + 'px' })
    q('.rest-row .fill').forEach(f => { f.style.width = f.dataset.w + '%' })
  }))
}

function activateTimeline() {
  const items = q('.tl-item')
  items.forEach(el => {
    if (el.getBoundingClientRect().top < window.innerHeight * 0.9) {
      el.classList.add('visible')
      el.querySelectorAll('.fill').forEach(f => { (f as HTMLElement).style.width = (f as HTMLElement).dataset.w + '%' })
    }
  })
  if (io) io.disconnect()
  io = new IntersectionObserver(entries => {
    entries.forEach(e => {
      if (e.isIntersecting) {
        e.target.classList.add('visible')
        e.target.querySelectorAll('.fill').forEach(f => { (f as HTMLElement).style.width = (f as HTMLElement).dataset.w + '%' })
        io?.unobserve(e.target)
      }
    })
  }, { threshold: 0.25 })
  items.forEach(el => io?.observe(el))
}

function activateHeatmap() {
  heatOn.value = false
  nextTick(() => { heatOn.value = true })
}

/* ---------- 加载 ---------- */
onMounted(async () => {
  loading.value = true
  loadAppVersion()
  try {
    const [ov, ms, mm, prompts, gens, booksRes] = await Promise.all([
      statsAPI.getOverview(),
      statsAPI.getModelStats(),
      statsAPI.getMonthly(),
      promptAPI.getAll(),
      generatorAPI.getAll(),
      bookAPI.getAll(),
    ])
    // 概览：后端只返回 total_usage 等，提示词/生成器/书本数量用真实 API 统计
    bookList.value = booksRes.data ?? []
    overview.value = {
      ...normalizeOverview(ov.data),
      promptCount: prompts.data?.length ?? 0,
      generatorCount: gens.data?.length ?? 0,
      bookCount: bookList.value.length,
    }
    modelStats.value = normalizeModelStats(ms.data ?? [])
    monthlyStats.value = normalizeMonthly(mm.data ?? [])
    // 使用记录为空 → 演示数据兜底（仅开发环境，便于 UI 评估）
    // 生产构建时整个块被 tree-shaking 移除，模拟数据不会进入产物
    if (import.meta.env.DEV && (!modelStats.value.length || !monthlyStats.value.length)) {
      const DEMO_MODEL_STATS: ModelStats[] = [
        { model_id: 1, model_name: 'GPT-4o', provider_name: 'OpenAI', total_usage: 3120, total_tokens: 5600000 },
        { model_id: 2, model_name: 'Claude 3.5 Sonnet', provider_name: 'Anthropic', total_usage: 2480, total_tokens: 4100000 },
        { model_id: 3, model_name: 'DeepSeek-V3', provider_name: 'DeepSeek', total_usage: 1560, total_tokens: 2900000 },
        { model_id: 4, model_name: 'Qwen2.5', provider_name: '阿里云', total_usage: 890, total_tokens: 1500000 },
        { model_id: 5, model_name: 'GLM-4', provider_name: '智谱', total_usage: 592, total_tokens: 900000 },
        { model_id: 6, model_name: 'Kimi', provider_name: '月之暗面', total_usage: 0, total_tokens: 0 },
      ]
      const DEMO_MONTHLY: MonthlyStats[] = [
        { month: '1', total_usage: 320, total_tokens: 1200000, model_count: 3 },
        { month: '2', total_usage: 480, total_tokens: 1800000, model_count: 4 },
        { month: '3', total_usage: 760, total_tokens: 2900000, model_count: 5 },
        { month: '4', total_usage: 1050, total_tokens: 4100000, model_count: 6 },
        { month: '5', total_usage: 1380, total_tokens: 5600000, model_count: 6 },
        { month: '6', total_usage: 1520, total_tokens: 6300000, model_count: 7 },
        { month: '7', total_usage: 1650, total_tokens: 7100000, model_count: 7 },
        { month: '8', total_usage: 1482, total_tokens: 6800000, model_count: 7 },
      ]
      isDemo.value = true
      modelStats.value = DEMO_MODEL_STATS
      monthlyStats.value = DEMO_MONTHLY
      overview.value = { ...overview.value, totalUsageCount: 8642, totalTokens: 36000000 }
    }
  } catch (error) {
    console.error('加载统计失败:', error)
  } finally {
    loading.value = false
    nextTick(() => {
      activateOv()
      if (tab.value === 'timeline') activateTimeline()
    })
  }
})
</script>

<style scoped>
/* 统一走 MainLayout 注入的主题变量：亮/暗两套自动跟随 */
.profile-container {
  padding: 20px;
  margin: 0;
  box-sizing: border-box;
}

/* ---------- 创作仪表盘 ---------- */
.dash-brand {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.brand-ver {
  flex: none;
  padding: 4px 10px;
  border-radius: 999px;
  background: var(--accent-soft, #eef4ff);
  color: var(--accent, #2563eb);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: .3px;
  white-space: nowrap;
}

.dash-brand h2 {
  margin: 0;
  font-size: 23px;
  font-weight: 800;
  letter-spacing: .5px;
  color: var(--app-color, #1a2230);
}

/* 页眉落款：24px 品牌标，与标题同一基线。
   尺寸取 24 是规范里的"文档/页眉"档，不抢标题的视觉重量。 */
.brand-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.dash-mark {
  width: 24px;
  height: 24px;
  border-radius: 5px;
  display: block;
  flex: none;
}

.dash-brand .sub {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
  margin-top: 5px;
}

.stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
  margin-bottom: 22px;
}

.stat {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 16px;
  padding: 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  position: relative;
  overflow: hidden;
  transition: transform .18s ease, box-shadow .18s ease;
}

.stat:hover {
  transform: translateY(-3px);
  box-shadow: 0 10px 26px rgba(60, 80, 130, .1);
}

.stat .s-label {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
  font-weight: 600;
}

.stat .s-value {
  font-size: 28px;
  font-weight: 900;
  font-variant-numeric: tabular-nums;
  color: var(--app-color, #1a2230);
}

.stat .s-value small {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
  font-weight: 600;
}

.stat .s-foot {
  font-size: 11px;
  color: #16a365;
  font-weight: 700;
}

.stat .bar {
  height: 6px;
  border-radius: 3px;
  background: var(--accent-soft, #eef4ff);
  overflow: hidden;
}

.stat .bar i {
  display: block;
  height: 100%;
  border-radius: 3px;
  transition: width .4s ease;
}

.stat .dot {
  position: absolute;
  top: 16px;
  right: 18px;
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 15px;
}

.stat.s1 .dot { background: linear-gradient(135deg, #6c8cff, #9c6cff); }
.stat.s1 .bar i { background: linear-gradient(90deg, #6c8cff, #9c6cff); }
.stat.s2 .dot { background: linear-gradient(135deg, #ffa63d, #ff7d5c); }
.stat.s2 .bar i { background: linear-gradient(90deg, #ffa63d, #ff7d5c); }
.stat.s3 .dot { background: linear-gradient(135deg, #2ac79a, #54d4d4); }
.stat.s3 .bar i { background: linear-gradient(90deg, #2ac79a, #54d4d4); }
.stat.s4 .dot { background: linear-gradient(135deg, #ff6c9c, #ff6cd0); }
.stat.s4 .bar i { background: linear-gradient(90deg, #ff6c9c, #ff6cd0); }

/* ---------- 顶栏 ---------- */
.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 12px;
}

.page-title {
  font-size: 22px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
}

.topbar .sub {
  font-size: 13px;
  color: var(--ink3, #98a5b8);
  margin-top: 4px;
}

.demo-tag {
  display: inline-block;
  font-size: 11px;
  font-weight: 600;
  color: #f59e0b;
  background: rgba(245, 158, 11, .12);
  padding: 1px 8px;
  border-radius: 10px;
  margin-left: 4px;
  vertical-align: 1px;
}

.seg {
  display: inline-flex;
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 14px;
  padding: 4px;
  gap: 4px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
}

.seg button {
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  color: var(--ink2, #5a6a80);
  padding: 8px 16px;
  border-radius: 10px;
  cursor: pointer;
  transition: all .25s;
  white-space: nowrap;
}

.seg button.active {
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #fff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, .3);
}

/* ---------- 视图切换 ---------- */
.view {
  animation: fade .4s ease;
}

@keyframes fade {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* ---------- 概览：统计卡 ---------- */
.ov-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.ov-card {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 16px;
  padding: 18px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
  transition: transform .25s;
}

.ov-card:hover {
  transform: translateY(-3px);
}

.ov-card .o-ico {
  width: 40px;
  height: 40px;
  border-radius: 11px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #fff;
  margin-bottom: 12px;
}

.ov-card .o-val {
  font-size: 26px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  color: var(--app-color, #1a2230);
}

.ov-card .o-lbl {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
  margin-top: 3px;
}

.ov-card .o-delta {
  font-size: 11px;
  margin-top: 8px;
  font-weight: 700;
}

.ov-card .o-delta.up { color: #10b981; }
.ov-card .o-delta.down { color: #f43f5e; }

/* ---------- 概览：双栏面板 ---------- */
.ov-bottom {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-top: 14px;
}

.ov-panel {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 16px;
  padding: 18px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
}

.ov-panel h3 {
  font-size: 14px;
  margin-bottom: 12px;
  color: var(--app-color, #1a2230);
}

.ov-row {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 0;
  font-size: 13px;
}

.ov-row .n {
  width: 120px;
  color: var(--ink2, #5a6a80);
  flex-shrink: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ov-row .tr {
  flex: 1;
  height: 8px;
  border-radius: 5px;
  background: var(--accent-soft, #eef4ff);
  overflow: hidden;
}

.ov-row .fl {
  height: 100%;
  width: 0;
  border-radius: 5px;
  transition: width .8s cubic-bezier(.4, 0, .2, 1);
}

.ov-row .v {
  width: 56px;
  text-align: right;
  color: var(--ink2, #5a6a80);
  font-variant-numeric: tabular-nums;
}

/* ---------- 领奖台 ---------- */
.podium-card {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 16px;
  padding: 18px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
  margin-top: 14px;
}

.podium-card h3 {
  font-size: 14px;
  margin-bottom: 12px;
  color: var(--app-color, #1a2230);
}

.podium {
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 14px;
  height: 300px;
  padding: 0 10px;
}

.podium .slot {
  flex: 1;
  max-width: 220px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.podium .avatar {
  width: 52px;
  height: 52px;
  border-radius: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  color: #fff;
  margin-bottom: 10px;
  box-shadow: 0 8px 20px rgba(0, 0, 0, .15);
}

.podium .p-name {
  font-size: 13px;
  font-weight: 700;
  text-align: center;
  color: var(--app-color, #1a2230);
}

.podium .p-provider {
  font-size: 11px;
  color: var(--ink3, #98a5b8);
  margin-top: 2px;
}

.podium .p-count {
  font-size: 12px;
  font-weight: 800;
  color: var(--accent, #2563eb);
  margin-top: 4px;
  font-variant-numeric: tabular-nums;
}

.podium .step {
  width: 100%;
  border-radius: 12px 12px 0 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 10px;
  font-size: 22px;
  font-weight: 800;
  color: rgba(255, 255, 255, .95);
  transition: height .8s cubic-bezier(.4, 0, .2, 1);
}

.podium .step.s1 { height: 0; background: linear-gradient(135deg, #ffd700, #fbbf24); }
.podium .step.s2 { height: 0; background: linear-gradient(135deg, #c0c0c0, #e2e8f0); }
.podium .step.s3 { height: 0; background: linear-gradient(135deg, #cd7f32, #daa06d); }

.slot.first .avatar {
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  box-shadow: 0 0 0 4px rgba(99, 102, 241, .25), 0 0 30px rgba(99, 102, 241, .45);
  animation: glow 2.4s ease-in-out infinite;
}

@keyframes glow {
  0%, 100% { box-shadow: 0 0 0 4px rgba(99, 102, 241, .25), 0 0 24px rgba(99, 102, 241, .4); }
  50% { box-shadow: 0 0 0 6px rgba(99, 102, 241, .35), 0 0 44px rgba(99, 102, 241, .6); }
}

.slot.first .crown {
  font-size: 22px;
  margin-bottom: 2px;
  animation: bob 2.4s ease-in-out infinite;
}

@keyframes bob {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}

.slot.second .avatar { background: linear-gradient(135deg, #94a3b8, #cbd5e1); }
.slot.third .avatar { background: linear-gradient(135deg, #b45309, #d97706); }

.rest {
  margin-top: 18px;
  border-top: 1px solid var(--topbar-border, #e6eaf0);
  padding-top: 8px;
}

.rest-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 6px;
  border-radius: 10px;
  transition: background .2s;
}

.rest-row:hover {
  background: var(--accent-soft, #eef4ff);
}

.rest-row .idx {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: var(--accent-soft, #eef4ff);
  color: var(--ink2, #5a6a80);
  font-size: 12px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.rest-row .name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: var(--app-color, #1a2230);
}

.rest-row .provider {
  font-size: 11px;
  color: var(--ink3, #98a5b8);
}

.rest-row .cnt {
  font-size: 13px;
  font-weight: 700;
  color: var(--ink2, #5a6a80);
  font-variant-numeric: tabular-nums;
}

.rest-row .bar {
  width: 90px;
  height: 6px;
  border-radius: 4px;
  background: var(--accent-soft, #eef4ff);
  overflow: hidden;
}

.rest-row .bar .fill {
  height: 100%;
  width: 0;
  border-radius: 4px;
  background: linear-gradient(90deg, #a5b4fc, #818cf8);
  transition: width .8s;
}

/* ---------- 时间轴 ---------- */
.prologue {
  text-align: center;
  padding: 20px 0 6px;
}

.prologue .kicker {
  font-size: 12px;
  letter-spacing: .25em;
  color: var(--accent, #2563eb);
  font-weight: 700;
}

.prologue h2 {
  font-size: 24px;
  font-weight: 800;
  margin: 10px 0 8px;
  color: var(--app-color, #1a2230);
}

.prologue p {
  font-size: 14px;
  color: var(--ink2, #5a6a80);
  line-height: 1.7;
}

.timeline {
  position: relative;
  margin-top: 32px;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 50%;
  top: 0;
  bottom: 0;
  width: 2px;
  background: linear-gradient(180deg, var(--accent, #2563eb), transparent);
  transform: translateX(-50%);
  opacity: .35;
}

.tl-item {
  position: relative;
  width: 50%;
  padding: 0 34px 40px;
}

.tl-item:nth-child(odd) { left: 0; text-align: right; }
.tl-item:nth-child(even) { left: 50%; text-align: left; }

.tl-dot {
  position: absolute;
  top: 6px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: var(--card, #ffffff);
  border: 3px solid var(--accent, #2563eb);
  box-shadow: 0 0 0 5px var(--accent-soft, #eef4ff);
}

.tl-item:nth-child(odd) .tl-dot { right: -8px; }
.tl-item:nth-child(even) .tl-dot { left: -8px; }

.tl-item.milestone .tl-dot {
  border-color: #f59e0b;
  box-shadow: 0 0 0 5px rgba(245, 158, 11, .18);
}

.tl-item.milestone .tl-dot::after {
  content: '★';
  position: absolute;
  top: -9px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 13px;
  color: #f59e0b;
}

.tl-card {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 14px;
  padding: 16px 18px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
  display: inline-block;
  min-width: 200px;
  max-width: 100%;
  opacity: 0;
  transform: translateY(24px);
  transition: opacity .6s, transform .6s cubic-bezier(.4, 0, .2, 1);
}

.tl-item.visible .tl-card {
  opacity: 1;
  transform: translateY(0);
}

.tl-month {
  font-size: 13px;
  font-weight: 800;
  color: var(--accent, #2563eb);
}

.tl-item.milestone .tl-month { color: #f59e0b; }

.tl-count {
  font-size: 30px;
  font-weight: 800;
  margin: 4px 0 2px;
  font-variant-numeric: tabular-nums;
  color: var(--app-color, #1a2230);
}

.tl-count small {
  font-size: 13px;
  color: var(--ink3, #98a5b8);
  font-weight: 600;
}

.tl-bar {
  height: 6px;
  border-radius: 4px;
  background: var(--accent-soft, #eef4ff);
  overflow: hidden;
  margin-top: 8px;
}

.tl-bar .fill {
  height: 100%;
  width: 0;
  border-radius: 4px;
  background: linear-gradient(90deg, #3b82f6, #8b5cf6);
  transition: width 1s cubic-bezier(.4, 0, .2, 1) .2s;
}

.tl-note {
  font-size: 12px;
  color: var(--ink2, #5a6a80);
  margin-top: 8px;
}

.tl-note .tag {
  display: inline-block;
  background: rgba(245, 158, 11, .12);
  color: #f59e0b;
  font-weight: 700;
  padding: 1px 8px;
  border-radius: 10px;
  font-size: 11px;
  margin-right: 4px;
}

/* ---------- 热力图 ---------- */
.card {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 16px;
  padding: 24px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
}

.card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 18px;
  flex-wrap: wrap;
  gap: 8px;
}

.card-head h2 {
  font-size: 16px;
  font-weight: 800;
  color: var(--app-color, #1a2230);
}

.card-head .stat {
  font-size: 13px;
  color: var(--ink2, #5a6a80);
}

.card-head .stat b {
  color: var(--accent, #2563eb);
  font-size: 16px;
}

.heat-wrap {
  overflow-x: auto;
  padding-bottom: 6px;
}

.heat {
  display: flex;
  gap: 3px;
}

.heat-col {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.cell {
  width: 13px;
  height: 13px;
  border-radius: 3px;
  background: var(--accent-soft, #eef4ff);
  cursor: pointer;
  position: relative;
  opacity: 0;
  transform: scale(.4);
  transition: opacity .3s, transform .3s;
}

.view.on .cell {
  opacity: 1;
  transform: scale(1);
}

.cell:hover {
  transform: scale(1.35) !important;
  z-index: 2;
}

.cell.l1 { background: rgba(59, 130, 246, .25); }
.cell.l2 { background: rgba(59, 130, 246, .5); }
.cell.l3 { background: rgba(99, 102, 241, .75); }
.cell.l4 { background: linear-gradient(135deg, #3b82f6, #8b5cf6); }

.cell .tip {
  position: absolute;
  bottom: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%) translateY(4px);
  background: var(--app-color, #1a2230);
  color: var(--card, #ffffff);
  font-size: 11px;
  padding: 5px 9px;
  border-radius: 7px;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity .2s, transform .2s;
  z-index: 9;
}

.cell:hover .tip {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}

.heat-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 14px;
  font-size: 12px;
  color: var(--ink3, #98a5b8);
}

.legend {
  display: flex;
  align-items: center;
  gap: 4px;
}

.legend .sw {
  width: 11px;
  height: 11px;
  border-radius: 3px;
  background: var(--accent-soft, #eef4ff);
}

.legend .sw.l1 { background: rgba(59, 130, 246, .25); }
.legend .sw.l2 { background: rgba(59, 130, 246, .5); }
.legend .sw.l3 { background: rgba(99, 102, 241, .75); }
.legend .sw.l4 { background: linear-gradient(135deg, #3b82f6, #8b5cf6); }

.insights {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  margin-top: 20px;
}

.insight {
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 14px;
  padding: 16px;
  box-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04));
}

.insight .i-ico { font-size: 20px; }
.insight .i-title {
  font-size: 13px;
  font-weight: 700;
  margin: 8px 0 4px;
  color: var(--app-color, #1a2230);
}
.insight .i-text {
  font-size: 12px;
  color: var(--ink2, #5a6a80);
  line-height: 1.6;
}
.insight .i-text b { color: var(--accent, #2563eb); }

/* ---------- 空状态 ---------- */
.empty {
  text-align: center;
  color: var(--ink3, #98a5b8);
  font-size: 13px;
}
</style>
