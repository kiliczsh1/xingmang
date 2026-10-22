<template>
  <div class="scan-page">
    <div class="scanline" aria-hidden="true"></div>

    <header class="scan-header">
      <div class="header-left">
        <div class="logo">AI_SCAN <span class="logo-slash">//</span> 扫榜终端</div>
        <h1>榜 <span class="accent">·</span> 扫描中</h1>
      </div>
      <!-- 数据源切换器居中：三栏栅格里中间列是 auto，两侧各 minmax(0,1fr) -->
      <div v-if="sources.length" class="source-switch" role="group" aria-label="榜单数据源">
        <button
          v-for="s in sources"
          :key="s.id"
          type="button"
          class="src"
          :class="{ on: s.id === source }"
          :title="s.desc"
          :disabled="loading"
          @click="switchSource(s.id)"
        >
          {{ s.name }}
        </button>
      </div>
      <div class="header-right">
        <div class="status" :class="{ demo: !!error }">
          <span class="dot"></span>
          {{ error ? '连接失败' : (sourceName || '数据源') + ' · 实时' }}
        </div>
        <button
          type="button"
          class="gear"
          :class="{ on: !!savedDefault }"
          :title="savedDefault ? `默认常用榜单：${savedDefaultPath}（点击修改）` : '设置默认常用榜单'"
          aria-label="设置默认常用榜单"
          @click="openSettings"
        >⚙</button>
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
      <!-- 掉榜：只有能抓完整榜的数据源才有（起点）。抓取中 / 待基准 都要如实说，不能显示 0 -->
      <button
        v-if="sourceFullCapture"
        type="button"
        class="chip drops-toggle"
        :class="{ on: showDrops }"
        :title="dropsHint"
        @click="showDrops = !showDrops"
      >[掉榜 {{ dropsBadge }}]</button>
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

    <!-- 第三层起：榜单自带的筛选维度（起点 0~2 个；番茄/七猫没有这一层）。
         维度内容全部由后端接口给出，前端不硬编码——起点的题材 ID 是分性别的。 -->
    <div v-for="dim in dims" :key="dim.key" class="dim-row">
      <span class="dim-label">{{ dim.name }}</span>
      <div class="dim-chips">
        <button
          v-for="o in dim.options"
          :key="o.id"
          type="button"
          class="dchip"
          :class="{ on: filters[dim.key] === o.id }"
          :disabled="loading"
          @click="switchFilter(dim.key, o.id)"
        >{{ o.name }}</button>
      </div>
    </div>

    <div class="scope">
      TARGET：<b>{{ activeGroupLabel }} · {{ activeGenreName }}</b>
      <span v-if="scopeFilters"> · {{ scopeFilters }}</span>
      · 命中 <b class="hit">{{ filtered.length }}</b>
    </div>

    <div class="scan-scroll">
    <!-- 掉榜视图：与「上一份整榜」相比掉出榜单的书。用名次列表而不是卡片，因为后端只给得到这些字段 -->
    <div v-if="showDrops" class="drop-list">
      <div class="drop-hint">{{ dropsHint }}</div>
      <div v-for="d in drops.drops" :key="d.bookId" class="drop-row">
        <span class="drop-rank">{{ String(d.rank || 0).padStart(3, '0') }}</span>
        <span class="drop-title" :title="d.title">{{ d.title }}</span>
        <span class="drop-auth">{{ d.author }}</span>
      </div>
      <div v-if="drops.ready && drops.baseDate && !drops.drops.length" class="empty">
        <div class="empty-icon" aria-hidden="true">◉</div>
        没有掉榜 · 与 {{ drops.baseDate }} 相比榜内没有掉出的书
      </div>
    </div>

    <main v-else class="grid">
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
          <div
            class="cover"
            :class="{ zoomable: !!b.coverUrl && !coverFailed.has(b.bookId) }"
            :role="b.coverUrl ? 'button' : undefined"
            :tabindex="b.coverUrl ? 0 : undefined"
            :aria-label="b.coverUrl ? `放大《${b.title}》封面` : undefined"
            :title="b.coverUrl ? '点击放大封面' : undefined"
            @click="onCoverClick(b)"
            @keydown.enter.prevent="onCoverClick(b)"
            @keydown.space.prevent="onCoverClick(b)"
          >
            <div class="cover-fallback" :style="{ background: coverGradient(b.title) }"><span>{{ (b.title || '?')[0] }}</span></div>
            <img v-if="b.coverUrl" :src="b.coverUrl" :alt="b.title" loading="lazy" referrerpolicy="no-referrer" @error="onCoverError($event, b)" />
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
              <span v-if="b.readingCount">{{ b.metricLabel || '在读' }}<b>{{ b.readingCount }}</b></span>
              <span v-if="b.wordCount">字数<b>{{ b.wordCount }}</b></span>
              <span>变动<b v-html="diffHtml(b.rankDiff, b.rankDiffMode, b.isNew)"></b></span>
              <span v-if="(b.daysOnList || 0) > 1">在榜<b>{{ b.daysOnList }}天</b></span>
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
          <span class="hint">{{ sourceDesc || '数据源未就绪' }} · 若持续失败请检查后端网络与榜单接口</span>
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
      <span>UPLINK：<span class="ok-blink">●</span> ONLINE · {{ sourceDesc || '数据源未就绪' }}</span>
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
            <b>{{ currentBook ? genreName(currentBook.category || '') : '—' }} · {{ activeGenreName }} · 第{{ currentBook?.rank ?? '—' }}名</b>
          </div>
          <div class="meta-line">
            <span class="lab">DATA&nbsp;</span>
            <span v-if="currentBook?.readingCount"><b>{{ currentBook.readingCount }}</b>人在读</span>
            <span v-if="currentBook?.wordCount"> · <b>{{ currentBook.wordCount }}</b>字</span>
          </div>
        </div>
      </div>

      <div class="glass-chips">
        <span class="chip"><b>榜</b>{{ activeGroupLabel }} · {{ activeGenreName }}</span>
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
      :source="source"
    />

    <!-- 封面放大预览（点卡片封面 / 回车 / 空格触发；Esc 关闭，← → 切换） -->
    <Teleport to="body">
      <div
        v-if="previewVisible"
        class="xm-cover-viewer"
        role="dialog"
        aria-modal="true"
        aria-label="封面放大预览"
        @click.self="closePreview"
      >
        <button class="cv-close" type="button" title="关闭（Esc）" aria-label="关闭预览" @click="closePreview">✕</button>
        <button
          class="cv-arrow prev"
          type="button"
          title="上一张（←）"
          aria-label="上一张"
          :disabled="previewList.length < 2"
          @click.stop="stepPreview(-1)"
        >‹</button>
        <figure class="cv-figure" @click.stop>
          <div class="cv-frame">
            <div v-if="previewLoading" class="cv-skel" aria-hidden="true"></div>
            <img
              v-if="previewUrl && !previewFailed"
              :src="previewUrl"
              :alt="previewBook?.title"
              referrerpolicy="no-referrer"
              @load="previewLoading = false"
              @error="onPreviewError"
            />
            <div v-if="previewFailed" class="cv-fail">封面加载失败</div>
          </div>
          <figcaption class="cv-cap">
            <b>{{ previewBook?.title || '—' }}</b>
            <span v-if="previewBook">
              {{ previewBook.author || '—' }} · {{ genreName(previewBook.category || '') }} · 第 {{ previewBook.rank ?? '—' }} 名
            </span>
            <span class="cv-count">{{ previewIndex + 1 }} / {{ previewList.length }}</span>
          </figcaption>
        </figure>
        <button
          class="cv-arrow next"
          type="button"
          title="下一张（→）"
          aria-label="下一张"
          :disabled="previewList.length < 2"
          @click.stop="stepPreview(1)"
        >›</button>
      </div>
    </Teleport>

    <!-- 设置弹窗：设定打开 AI 扫榜时默认进入的榜单。
         选完数据源会现拉该源的目录，所以分组/榜单/筛选都是真实可选值，不是写死的。 -->
    <Teleport to="body">
      <div
        v-if="settingsVisible"
        class="xm-scan-settings"
        role="dialog"
        aria-modal="true"
        aria-label="设置默认常用榜单"
        @click.self="closeSettings"
      >
        <div class="ss-panel" @click.stop>
          <div class="ss-head">
            <b>设置</b>
            <button class="ss-x" type="button" title="关闭（Esc）" aria-label="关闭设置" @click="closeSettings">✕</button>
          </div>

          <div class="ss-body">
            <div class="ss-label">默认常用榜单</div>
            <div class="ss-hint">打开 AI 扫榜时自动进入下面这个组合</div>

            <div class="ss-row">
              <span class="ss-key">数据源</span>
              <div class="ss-chips">
                <button
                  v-for="s in sources"
                  :key="s.id"
                  type="button"
                  class="sc"
                  :class="{ on: s.id === cfgSource }"
                  @click="switchCfgSource(s.id)"
                >{{ s.name }}</button>
              </div>
            </div>

            <div v-if="cfgLoading" class="ss-loading">目录加载中…</div>
            <div v-else-if="cfgError" class="ss-err">{{ cfgError }}</div>
            <template v-else>
              <div v-if="cfgGroups.length" class="ss-row">
                <span class="ss-key">分组</span>
                <div class="ss-chips">
                  <button
                    v-for="(g, gi) in cfgGroups"
                    :key="g.title"
                    type="button"
                    class="sc"
                    :class="{ on: gi === cfgGroupIndex }"
                    @click="switchCfgGroup(gi)"
                  >{{ g.title }}</button>
                </div>
              </div>

              <div v-if="cfgItems.length" class="ss-row">
                <span class="ss-key">榜单</span>
                <div class="ss-chips">
                  <button
                    v-for="it in cfgItems"
                    :key="it.id"
                    type="button"
                    class="sc"
                    :class="{ on: it.id === cfgTab }"
                    @click="switchCfgTab(it.id)"
                  >{{ it.name }}</button>
                </div>
              </div>

              <div v-for="d in cfgDims" :key="d.key" class="ss-row">
                <span class="ss-key">{{ d.name }}</span>
                <div class="ss-chips">
                  <button
                    v-for="o in d.options"
                    :key="o.id"
                    type="button"
                    class="sc"
                    :class="{ on: cfgFilters[d.key] === o.id }"
                    @click="switchCfgFilter(d.key, o.id)"
                  >{{ o.name }}</button>
                </div>
              </div>
            </template>

            <div class="ss-preview">
              <span class="ss-key">将进入</span>
              <b>{{ cfgPath || '—' }}</b>
            </div>

            <div v-if="savedDefault" class="ss-current">
              当前已设：{{ savedDefaultPath }}
            </div>
            <div v-else class="ss-current muted">未设置 · 打开时默认进入番茄小说</div>
          </div>

          <div class="ss-foot">
            <button class="ss-btn ghost" type="button" :disabled="!savedDefault" @click="clearDefault">清除默认</button>
            <div class="ss-spacer"></div>
            <button class="ss-btn" type="button" :disabled="!cfgTab" @click="saveDefault">保存为默认</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { rankAPI, type RankBook, type RankDrops, type RankFilterDim, type RankGroup, type RankSource } from '@/api'
import AiAnalyzeDialog from '@/components/WriteDialogs/AiAnalyzeDialog.vue'

/**
 * AI 扫榜 —— 赛博扫描风（demo-10 方向）
 * 数据：真实榜单，数据源在**页头居中**的切换器里切换（列表由后端 /ai/rank/sources 注册表驱动）。
 *   - 番茄小说：Go 后端内置实现（internal/fanqie，a_bogus 签名 + PUA 字体解码）
 *   - 七猫小说：书库分类接口（internal/qimao，题材维度）
 *   - 起点中文网：移动端公开接口（internal/qidian，无字体反爬），是唯一能抓完整榜的，
 *     因此独有「掉榜」视图与可信的「新上榜」判定
 *   - 飞卢小说：GB18030 静态页（internal/feilu）
 * 已删除演示样本：数据源不可用时如实显示错误态，不做假数据回退。
 */

// 数据源本身**不做记忆**（切换不会持久化），但允许在设置里显式指定一个
// 「默认常用榜单」——那是用户主动选定的组合，打开时按它进入；没设置就回到后端默认（番茄）。
const DEFAULT_KEY = 'xingmang.scan.default'

/** 默认常用榜单：一个完整组合（源 + 榜单 + 筛选），不是单个数据源。 */
interface ScanDefault {
  source: string
  groupIndex: number
  tab: string
  filters: Record<string, string>
  /** 保存时的人读描述（存下来是为了不用拉目录也能在按钮上显示它） */
  label: string
}

// 打开时待应用的默认组合。loadSources 里读出来，等目录拉回后再由 applyPresetOrDefaults 落地
let pendingPreset: ScanDefault | null = null

const savedDefault = ref<ScanDefault | null>(readStoredDefault())
// 设置弹窗里正在编辑的组合
const settingsVisible = ref(false)
const cfgSource = ref('')
const cfgGroups = ref<RankGroup[]>([])
const cfgGroupIndex = ref(0)
const cfgTab = ref('')
const cfgFilters = ref<Record<string, string>>({})
const cfgLoading = ref(false)
const cfgError = ref('')

const sources = ref<RankSource[]>([])
const source = ref('')
const groups = ref<RankGroup[]>([])
const activeGroupIndex = ref(0)
const activeGenre = ref('') // 榜单标识，格式由各数据源自定（番茄 "1_2_1141"；七猫 "0_1_1"；起点 "male_hotsales"）
// 榜单内的筛选取值（起点：题材/月份/周期）。没有第三层的源恒为空对象
const filters = ref<Record<string, string>>({})
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
// 掉榜（只有能抓完整榜的源才有）
const drops = ref<RankDrops>(emptyDrops())
const showDrops = ref(false)

function emptyDrops(): RankDrops {
  return { ready: false, baseDate: '', drops: [] }
}

// 请求序号：切换数据源 / 榜单 / 筛选 / 翻页都会作废此前未返回的请求，避免旧响应覆盖新视图
let reqSeq = 0

const sourceName = computed(() => sources.value.find(s => s.id === source.value)?.name || '')
const sourceDesc = computed(() => sources.value.find(s => s.id === source.value)?.desc || '')
const sourceFullCapture = computed(() => sources.value.find(s => s.id === source.value)?.fullCapture === true)

const activeGroup = computed(() => groups.value[activeGroupIndex.value])
const genres = computed(() => activeGroup.value?.items ?? [])
const activeGroupLabel = computed(() => activeGroup.value?.title ?? '')
const activeItem = computed(() => genres.value.find(g => g.id === activeGenre.value))
const activeGenreName = computed(() => activeItem.value?.name || activeGenre.value || '—')
// 当前榜单自带的筛选维度（起点 0~2 个，番茄/七猫恒为空）
const dims = computed<RankFilterDim[]>(() => activeItem.value?.filters ?? [])

// 拼给后端的 filters 参数：catId:21,yearmonth:202608
const filterQuery = computed(() =>
  dims.value
    .map(d => (filters.value[d.key] ? `${d.key}:${filters.value[d.key]}` : ''))
    .filter(Boolean)
    .join(',')
)
const scopeFilters = computed(() =>
  dims.value
    .map(d => d.options.find(o => o.id === filters.value[d.key])?.name)
    .filter(Boolean)
    .join(' · ')
)

const dropsBadge = computed(() => {
  if (!drops.value.ready) return '抓取中'
  if (!drops.value.baseDate) return '待基准'
  return String(drops.value.drops.length)
})
const dropsHint = computed(() => {
  if (!drops.value.ready) return '今天的整榜还在后台抓取，抓完才能判定掉榜'
  if (!drops.value.baseDate) return '今天第一次整榜，明天起才有对比基准'
  return `与 ${drops.value.baseDate} 的整榜相比掉出榜单的书，前面的数字是它当时的最后名次`
})

const filtered = computed(() => {
  let list = books.value
  const kw = keyword.value.trim()
  if (kw) list = list.filter(b => b.title.includes(kw) || b.author.includes(kw))
  return list
})
const progress = computed(() => Math.max(1, Math.round((filtered.value.length / (books.value.length || 1)) * 10)))
// 加载骨架屏数量（最多 6 张占位卡）
const skeletons = computed(() => Array(Math.min(6, size)).fill(0).map((_, i) => i))

// 题材名：番茄的 category 是 categoryId（需查分组里的题材列表）；
// 七猫的 category 本身就是题材名（category2_name），查不到时原样返回。
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
// 变动：新上榜优先；step 口径带幅度（番茄原生 / 七猫由后端快照补齐）;
// dir 口径只有方向（七猫首次抓取、还没攒到快照时），此时带数字就是假数据
const diffHtml = (d: number, mode?: string, isNew?: boolean) => {
  if (isNew) return '<span class="new">NEW</span>'
  const dirOnly = mode === 'dir'
  if (d > 0) return `<span class="up">▲${dirOnly ? '' : d}</span>`
  if (d < 0) return `<span class="down">▼${dirOnly ? '' : -d}</span>`
  return '<span>—</span>'
}

// 切换榜单时把筛选维度重置为接口给的缺省值（月份这类有真实默认值，不能想当然取第一个）
const applyDefaultFilters = () => {
  const next: Record<string, string> = {}
  for (const d of dims.value) {
    next[d.key] = d.defaultId || d.options[0]?.id || ''
  }
  filters.value = next
}

const loadDrops = async (seq: number) => {
  if (!sourceFullCapture.value || !activeGenre.value) {
    drops.value = emptyDrops()
    return
  }
  try {
    const res = await rankAPI.getDrops({
      source: source.value,
      tab: activeGenre.value,
      filters: filterQuery.value || undefined
    })
    if (seq !== reqSeq) return
    drops.value = res.data ?? emptyDrops()
  } catch {
    if (seq === reqSeq) drops.value = emptyDrops()
  }
}

const loadBooks = async () => {
  if (!activeGenre.value) return
  const seq = ++reqSeq
  loading.value = true
  error.value = ''
  previewVisible.value = false // 换榜/翻页后旧预览的图集已过期，直接关掉
  try {
    const res = await rankAPI.getBooks({
      source: source.value,
      tab: activeGenre.value,
      filters: filterQuery.value || undefined,
      page: page.value,
      size
    })
    if (seq !== reqSeq) return
    books.value = res.data?.books ?? []
  } catch {
    if (seq !== reqSeq) return
    books.value = []
    error.value = `${sourceName.value || '榜单'}不可用：后端无法获取数据`
  } finally {
    if (seq === reqSeq) loading.value = false
  }
  await loadDrops(seq)
}

const loadListsAndBooks = async () => {
  const seq = ++reqSeq
  try {
    const res = await rankAPI.getLists(source.value)
    if (seq !== reqSeq) return
    groups.value = res.data?.groups ?? []
  } catch {
    if (seq !== reqSeq) return
    groups.value = []
    books.value = []
    error.value = `${sourceName.value || '榜单'}不可用：后端无法获取数据`
    return
  }
  if (!groups.value.length) {
    books.value = []
    error.value = `${sourceName.value || '榜单'}未返回任何分组`
    return
  }
  page.value = 0
  applyPresetOrDefaults()
  await loadBooks()
}

// 落地打开时的组合：有默认常用榜单就按它走，否则第一个分组的第一项 + 接口缺省筛选
const applyPresetOrDefaults = () => {
  const p = pendingPreset
  pendingPreset = null

  if (p) {
    // tab 可能因为接口改版而不在目录里：先按 tab 找它属于哪个分组，
    // 找不到再退回记录的分组下标，仍越界则回到第一组 —— 不因为一条失效记录就白屏
    let gi = groups.value.findIndex(g => (g.items ?? []).some(i => i.id === p.tab))
    if (gi < 0) gi = p.groupIndex >= 0 && p.groupIndex < groups.value.length ? p.groupIndex : 0
    activeGroupIndex.value = gi

    const items = groups.value[gi]?.items ?? []
    activeGenre.value = items.find(i => i.id === p.tab)?.id || items[0]?.id || ''

    // 筛选逐维度对齐：保存的取值在当前维度里还存在才用它，否则用接口的缺省值
    // （起点月份的缺省是「当月」，不是第一个选项，不能用 options[0] 顶替）
    const next: Record<string, string> = {}
    for (const d of dims.value) {
      const saved = p.filters?.[d.key]
      next[d.key] = saved && d.options.some(o => o.id === saved) ? saved : d.defaultId || d.options[0]?.id || ''
    }
    filters.value = next
    return
  }

  activeGroupIndex.value = 0
  activeGenre.value = groups.value[0].items[0]?.id || ''
  applyDefaultFilters()
}

// 拉取可用数据源，并选中默认值：设置过「默认常用榜单」就用它，否则用后端默认（番茄小说）
const loadSources = async () => {
  let list: RankSource[] = []
  let defaultId = ''
  try {
    const res = await rankAPI.getSources()
    list = res.data?.sources ?? []
    defaultId = res.data?.default || ''
  } catch {
    error.value = '榜单数据源不可用：后端未连接'
    return
  }
  if (!list.length) {
    error.value = '榜单数据源未注册'
    return
  }
  sources.value = list

  // 默认组合指向的源若已下线（注册表里没有），整条记录作废，退回后端默认
  const stored = savedDefault.value
  const preset = stored && list.some(s => s.id === stored.source) ? stored : null
  if (!preset && stored) {
    localStorage.removeItem(DEFAULT_KEY)
    savedDefault.value = null
  }
  pendingPreset = preset
  source.value = preset?.source || list.find(s => s.id === defaultId)?.id || list[0].id
  await loadListsAndBooks()
}

const switchSource = (id: string) => {
  if (id === source.value) return
  source.value = id
  keyword.value = ''
  groups.value = []
  books.value = []
  filters.value = {}
  drops.value = emptyDrops()
  showDrops.value = false
  activeGenre.value = ''
  error.value = ''
  loadListsAndBooks()
}

const switchGroup = (gi: number) => {
  activeGroupIndex.value = gi
  activeGenre.value = groups.value[gi]?.items[0]?.id || ''
  page.value = 0
  showDrops.value = false
  applyDefaultFilters()
  loadBooks()
}
const switchGenre = (id: string) => {
  activeGenre.value = id
  page.value = 0
  showDrops.value = false
  applyDefaultFilters()
  loadBooks()
}
// 切筛选维度：同一维度再点当前值不重复请求
const switchFilter = (key: string, id: string) => {
  if (filters.value[key] === id) return
  filters.value = { ...filters.value, [key]: id }
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

/* ─── 设置：默认常用榜单 ───
   只存一个组合（源 + 榜单 + 筛选），存在 localStorage。
   不存「上次浏览的位置」——那是隐式记忆，会让人以为看到的还是同一批数据；
   这里的默认是用户主动设的，没设就回到后端默认（番茄）。 */

function readStoredDefault(): ScanDefault | null {
  try {
    const raw = localStorage.getItem(DEFAULT_KEY)
    if (!raw) return null
    const p = JSON.parse(raw) as Partial<ScanDefault>
    if (!p || typeof p.source !== 'string' || !p.source || typeof p.tab !== 'string' || !p.tab) return null
    return {
      source: p.source,
      groupIndex: typeof p.groupIndex === 'number' && p.groupIndex >= 0 ? p.groupIndex : 0,
      tab: p.tab,
      filters: p.filters && typeof p.filters === 'object' ? p.filters : {},
      label: typeof p.label === 'string' ? p.label : ''
    }
  } catch {
    return null
  }
}

const savedDefaultPath = computed(() => {
  const p = savedDefault.value
  if (!p) return ''
  if (p.label) return p.label
  const name = sources.value.find(s => s.id === p.source)?.name || p.source
  return `${name} · ${p.tab}`
})

const cfgItems = computed(() => cfgGroups.value[cfgGroupIndex.value]?.items ?? [])
const cfgItem = computed(() => cfgItems.value.find(i => i.id === cfgTab.value))
const cfgDims = computed<RankFilterDim[]>(() => cfgItem.value?.filters ?? [])
const cfgPath = computed(() => {
  const parts = [
    sources.value.find(s => s.id === cfgSource.value)?.name || cfgSource.value,
    cfgGroups.value[cfgGroupIndex.value]?.title || '',
    cfgItem.value?.name || ''
  ]
  for (const d of cfgDims.value) {
    const name = d.options.find(o => o.id === cfgFilters.value[d.key])?.name
    if (name) parts.push(name)
  }
  return parts.filter(Boolean).join(' · ')
})

// 筛选重置同样以接口给的 defaultId 为准（起点月份的缺省是「当月」，不是第一个选项）
const resetCfgFilters = () => {
  const next: Record<string, string> = {}
  for (const d of cfgDims.value) next[d.key] = d.defaultId || d.options[0]?.id || ''
  cfgFilters.value = next
}

const loadCfgLists = async () => {
  const src = cfgSource.value
  if (!src) return
  cfgLoading.value = true
  cfgError.value = ''
  try {
    const res = await rankAPI.getLists(src)
    if (src !== cfgSource.value) return // 已经切到别的源，丢弃这次结果
    const list = res.data?.groups ?? []
    cfgGroups.value = list
    if (!list.length) {
      cfgTab.value = ''
      cfgError.value = '该数据源未返回任何分组'
      return
    }
    cfgGroupIndex.value = 0
    cfgTab.value = list[0].items[0]?.id || ''
    resetCfgFilters()
  } catch {
    if (src !== cfgSource.value) return
    cfgGroups.value = []
    cfgTab.value = ''
    cfgError.value = '目录拉取失败：后端无法获取该数据源'
  } finally {
    if (src === cfgSource.value) cfgLoading.value = false
  }
}

const openSettings = () => {
  settingsVisible.value = true
  // 已经编辑过就保留，不重拉；首次打开以「当前正在看的榜单」为起点，省得从头点
  if (!cfgSource.value) {
    cfgSource.value = source.value
    loadCfgLists()
  }
}

const closeSettings = () => {
  settingsVisible.value = false
}

const switchCfgSource = (id: string) => {
  if (id === cfgSource.value) return
  cfgSource.value = id
  cfgGroups.value = []
  cfgTab.value = ''
  cfgFilters.value = {}
  loadCfgLists()
}

const switchCfgGroup = (gi: number) => {
  cfgGroupIndex.value = gi
  cfgTab.value = cfgGroups.value[gi]?.items[0]?.id || ''
  resetCfgFilters()
}

const switchCfgTab = (id: string) => {
  cfgTab.value = id
  resetCfgFilters()
}

const switchCfgFilter = (key: string, id: string) => {
  cfgFilters.value = { ...cfgFilters.value, [key]: id }
}

const saveDefault = () => {
  if (!cfgTab.value) return
  const p: ScanDefault = {
    source: cfgSource.value,
    groupIndex: cfgGroupIndex.value,
    tab: cfgTab.value,
    filters: { ...cfgFilters.value },
    label: cfgPath.value
  }
  localStorage.setItem(DEFAULT_KEY, JSON.stringify(p))
  savedDefault.value = p
  settingsVisible.value = false
}

const clearDefault = () => {
  localStorage.removeItem(DEFAULT_KEY)
  savedDefault.value = null
}

const handleAnalyze = (b: RankBook) => {
  currentBook.value = b
  analysisDialogVisible.value = true
}
const handleAction = (type: 'title' | 'synopsis' | 'brainstorm' | 'analyze') => {
  analysisAction.value = type
  analysisActionVisible.value = true
}
const onCoverError = (e: Event, b?: RankBook | null) => {
  const img = e.target as HTMLImageElement
  img.style.display = 'none'
  if (b) coverFailed.value.add(b.bookId)
}

/* ─── 封面放大预览 ───
   可点性只排除「已确认加载失败」的封面（那时卡片上是渐变兜底色块，放大没意义）。
   不能加「必须已加载完」的条件：慢网络下图片还在下载时点封面会毫无反应，
   是典型的 dead click。失败状态由 @error 事件维护，而不是靠 naturalWidth 反推。 */
const coverFailed = ref<Set<string>>(new Set())
const previewVisible = ref(false)
const previewIndex = ref(0)
const previewList = ref<RankBook[]>([])
const previewLoading = ref(false)
const previewFailed = ref(false)

const previewBook = computed(() => previewList.value[previewIndex.value] ?? null)
const previewUrl = computed(() => previewBook.value?.coverUrl || '')

const onCoverClick = (b: RankBook) => {
  if (!b.coverUrl || coverFailed.value.has(b.bookId)) return

  // 预览范围为当前筛选结果里所有有封面的卡片，方便直接左右翻看
  const list = filtered.value.filter(x => x.coverUrl && !coverFailed.value.has(x.bookId))
  const idx = list.findIndex(x => x.bookId === b.bookId)
  if (idx < 0) return

  previewList.value = list
  previewIndex.value = idx
  previewLoading.value = true
  previewFailed.value = false
  previewVisible.value = true
}

const closePreview = () => {
  previewVisible.value = false
}

const stepPreview = (delta: number) => {
  const total = previewList.value.length
  if (total < 2) return
  previewIndex.value = (previewIndex.value + delta + total) % total
  previewLoading.value = true
  previewFailed.value = false
}

const onPreviewError = () => {
  previewLoading.value = false
  previewFailed.value = true
}

// Esc 关闭 / ← → 翻页。只在弹层打开时挂行为，避免影响其它页面的键盘操作。
const onPreviewKey = (e: KeyboardEvent) => {
  // 设置弹窗优先：它盖在最上层，此时 Esc 关设置而不是关预览
  if (settingsVisible.value) {
    if (e.key === 'Escape') closeSettings()
    return
  }
  if (!previewVisible.value) return
  if (e.key === 'Escape') closePreview()
  else if (e.key === 'ArrowLeft') stepPreview(-1)
  else if (e.key === 'ArrowRight') stepPreview(1)
}

onMounted(() => {
  window.addEventListener('keydown', onPreviewKey)
  loadSources()
})
onUnmounted(() => window.removeEventListener('keydown', onPreviewKey))
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

/* 三栏栅格：两侧各 minmax(0,1fr) 等宽、中间 auto ——
   切换器不管两侧内容多宽都落在真正的居中位置（flex 的 space-between 做不到） */
.scan-header {
  flex-shrink: 0;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 14px;
  margin-bottom: 14px;
}
.header-left {
  min-width: 0;
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
.header-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  min-width: 0;
}
/* 设置按钮（齿轮）。常态给足对比度：纯图标控件在浅底上太淡会看不见 */
.gear {
  flex: none;
  width: 32px;
  height: 30px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  line-height: 1;
  color: var(--muted);
  background: var(--panel-2);
  border: 1px solid var(--line);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.gear:hover {
  color: var(--cy);
  border-color: var(--cy);
}
/* 设过默认榜单时常亮，提示「这里有个已生效的配置」 */
.gear.on {
  color: var(--cy);
  border-color: color-mix(in srgb, var(--cy) 55%, transparent);
  background: color-mix(in srgb, var(--cy) 10%, transparent);
}
/* 数据源切换器（页头居中）。与 .chip 同族观感，但做成连体分段控件 */
.source-switch {
  display: flex;
  border: 1px solid var(--line);
  background: var(--panel-2);
  border-radius: 6px;
  overflow: hidden;
}
.source-switch .src {
  font-size: 12px;
  letter-spacing: 0.06em;
  padding: 6px 14px;
  color: var(--muted);
  background: none;
  border: none;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.source-switch .src + .src {
  border-left: 1px solid var(--line);
}
.source-switch .src:hover:not(:disabled) {
  color: var(--cy);
}
.source-switch .src.on {
  color: #04141a;
  background: var(--cy);
  font-weight: 700;
}
.source-switch .src:disabled {
  opacity: 0.5;
  cursor: default;
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

/* 第三层：榜单自带的筛选维度（起点）。一行一个维度，维度数量由接口决定 */
.dim-row {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}
.dim-label {
  flex: none;
  font-size: 11px;
  letter-spacing: 0.14em;
  color: var(--faint);
}
.dim-chips {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}
.dchip {
  font-size: 11px;
  letter-spacing: 0.04em;
  padding: 3px 9px;
  color: var(--muted);
  border: 1px solid var(--line);
  background: var(--panel-2);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.dchip:hover:not(:disabled) {
  color: var(--cy);
  border-color: var(--cy);
}
.dchip.on {
  color: #04141a;
  background: var(--cy);
  border-color: var(--cy);
  font-weight: 700;
}
.dchip:disabled {
  opacity: 0.5;
  cursor: default;
}

/* 掉榜视图 */
.drop-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.drop-hint {
  font-size: 12px;
  color: var(--muted);
  letter-spacing: 0.06em;
  padding: 2px 2px 12px;
}
.drop-row {
  display: flex;
  align-items: baseline;
  gap: 14px;
  padding: 9px 14px;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: 4px;
}
.drop-rank {
  flex: none;
  font-size: 13px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.04em;
  color: var(--warn);
}
.drop-title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: var(--ink);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.drop-auth {
  flex: none;
  font-size: 11.5px;
  color: var(--muted);
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
/* 有封面时可点放大：cursor + hover 叠一层「+」，避免用户不知道能点 */
.cover.zoomable {
  cursor: zoom-in;
}
.cover.zoomable::after {
  content: '+';
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  font-weight: 700;
  color: #ffffff;
  background: rgba(6, 18, 26, 0.5);
  opacity: 0;
  transition: opacity 0.15s;
  pointer-events: none;
}
.cover.zoomable:hover::after,
.cover.zoomable:focus-visible::after {
  opacity: 1;
}
.cover.zoomable:focus-visible {
  outline: 2px solid var(--cy);
  outline-offset: 2px;
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
  row-gap: 4px;
  flex-wrap: wrap;
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
/* 新上榜：用强调色与 ▲/▼ 区分开，避免和"上升N名"混淆 */
.data .new {
  color: var(--mg);
  letter-spacing: 0.08em;
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
  /* 窄屏放不下三栏，退回堆叠（居中在这里没有意义，居中列会跟两侧挤在一起） */
  .scan-header {
    grid-template-columns: 1fr;
    justify-items: start;
    gap: 10px;
  }
  .header-right {
    justify-content: flex-start;
  }
  .scan-header h1 {
    font-size: 24px;
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

/* ============ 封面放大预览 ============ */
/* 放在非 scoped 块：内容 Teleport 到 body，而 --cy / --panel 这些令牌挂在 .scan-page 上，
   查看器拿不到；用 :root[data-theme] 联写又会被 scoped 编译器剥掉前缀（同 .analysis-dialog 的坑）。
   遮罩层亮/暗两套都压成深色，所以内部文字统一用浅色，不需要再做主题分支。 */
.xm-cover-viewer {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5vh 6vw;
  background: rgba(8, 14, 24, 0.82);
  backdrop-filter: blur(8px) saturate(130%);
  -webkit-backdrop-filter: blur(8px) saturate(130%);
  font-family: 'SF Mono', 'Consolas', ui-monospace, 'PingFang SC', 'Microsoft YaHei', monospace;
}
.xm-cover-viewer .cv-figure {
  margin: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
}
/* 固定画框：图未加载时也有稳定尺寸，不会跳版；图片一律 contain 居中 */
.xm-cover-viewer .cv-frame {
  position: relative;
  width: min(420px, 80vw);
  height: min(560px, 66vh);
  border: 1px solid rgba(56, 189, 248, 0.3);
  border-radius: 10px;
  background: rgba(10, 20, 32, 0.72);
  box-shadow: 0 28px 80px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}
.xm-cover-viewer .cv-frame img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.xm-cover-viewer .cv-skel {
  position: absolute;
  inset: 0;
  background: rgba(56, 189, 248, 0.12);
  animation: cv-shimmer 1.4s ease-in-out infinite;
}
@keyframes cv-shimmer {
  0%, 100% { opacity: 0.4; }
  50% { opacity: 1; }
}
.xm-cover-viewer .cv-fail {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #9db6c6;
  letter-spacing: 0.06em;
}
.xm-cover-viewer .cv-cap {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  text-align: center;
  color: #e6f3fb;
  font-size: 13px;
  line-height: 1.6;
  max-width: min(520px, 84vw);
}
.xm-cover-viewer .cv-cap b {
  font-size: 15px;
  font-weight: 700;
}
.xm-cover-viewer .cv-cap span {
  color: #9db6c6;
  font-size: 12px;
}
.xm-cover-viewer .cv-count {
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.1em;
}
.xm-cover-viewer .cv-close,
.xm-cover-viewer .cv-arrow {
  position: absolute;
  border: 1px solid rgba(56, 189, 248, 0.35);
  background: rgba(10, 18, 30, 0.62);
  color: #cfe9f5;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  font-family: inherit;
}
.xm-cover-viewer .cv-close {
  top: 16px;
  right: 18px;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  font-size: 14px;
  line-height: 1;
}
.xm-cover-viewer .cv-arrow {
  top: 50%;
  transform: translateY(-50%);
  width: 46px;
  height: 70px;
  border-radius: 8px;
  font-size: 26px;
  line-height: 1;
}
.xm-cover-viewer .cv-arrow.prev { left: 18px; }
.xm-cover-viewer .cv-arrow.next { right: 18px; }
.xm-cover-viewer .cv-close:hover,
.xm-cover-viewer .cv-arrow:hover:not(:disabled) {
  border-color: #38bdf8;
  color: #ffffff;
  box-shadow: 0 0 14px rgba(56, 189, 248, 0.35);
}
.xm-cover-viewer .cv-arrow:disabled {
  opacity: 0.28;
  cursor: default;
}
.xm-cover-viewer .cv-close:focus-visible,
.xm-cover-viewer .cv-arrow:focus-visible {
  outline: 2px solid #38bdf8;
  outline-offset: 2px;
}
@media (max-width: 768px) {
  .xm-cover-viewer {
    padding: 4vh 3vw;
  }
  .xm-cover-viewer .cv-arrow {
    width: 38px;
    height: 56px;
    font-size: 22px;
  }
  .xm-cover-viewer .cv-arrow.prev { left: 8px; }
  .xm-cover-viewer .cv-arrow.next { right: 8px; }
}
@media (prefers-reduced-motion: reduce) {
  .xm-cover-viewer .cv-skel {
    animation: none;
    opacity: 0.6;
  }
}

/* ============ 默认常用榜单设置弹窗 ============ */
/* 与封面预览同一套做法：非 scoped 块（内容 Teleport 到 body，拿不到 .scan-page 上的令牌），
   遮罩压成深色，内部文字统一浅色，因此不需要再做亮/暗主题分支。 */
.xm-scan-settings {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5vh 6vw;
  background: rgba(8, 14, 24, 0.82);
  backdrop-filter: blur(8px) saturate(130%);
  -webkit-backdrop-filter: blur(8px) saturate(130%);
  font-family: 'SF Mono', 'Consolas', ui-monospace, 'PingFang SC', 'Microsoft YaHei', monospace;
}
.xm-scan-settings .ss-panel {
  display: flex;
  flex-direction: column;
  width: min(620px, 92vw);
  max-height: min(76vh, 720px);
  border: 1px solid rgba(56, 189, 248, 0.3);
  border-radius: 10px;
  background: rgba(10, 20, 32, 0.94);
  box-shadow: 0 28px 80px rgba(0, 0, 0, 0.5);
  overflow: hidden;
}
.xm-scan-settings .ss-head {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(56, 189, 248, 0.18);
  color: #e6f3fb;
  font-size: 14px;
  letter-spacing: 0.08em;
}
.xm-scan-settings .ss-x {
  border: 1px solid rgba(56, 189, 248, 0.35);
  background: rgba(10, 18, 30, 0.62);
  color: #cfe9f5;
  width: 26px;
  height: 26px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  line-height: 1;
  font-family: inherit;
  transition: all 0.15s;
}
.xm-scan-settings .ss-x:hover {
  background: #38bdf8;
  border-color: #38bdf8;
  color: #ffffff;
}
.xm-scan-settings .ss-body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 14px 16px;
}
.xm-scan-settings .ss-label {
  color: #e6f3fb;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.06em;
}
.xm-scan-settings .ss-hint {
  margin-top: 4px;
  color: #8aa6b6;
  font-size: 12px;
  line-height: 1.6;
}
.xm-scan-settings .ss-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: 12px;
}
.xm-scan-settings .ss-key {
  flex: none;
  width: 48px;
  padding-top: 5px;
  color: #8aa6b6;
  font-size: 12px;
  letter-spacing: 0.06em;
}
.xm-scan-settings .ss-chips {
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.xm-scan-settings .sc {
  padding: 4px 11px;
  font-size: 12px;
  color: #b9d3e0;
  background: rgba(18, 32, 48, 0.8);
  border: 1px solid rgba(56, 189, 248, 0.2);
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
}
.xm-scan-settings .sc:hover {
  border-color: rgba(56, 189, 248, 0.6);
  color: #d9f1fc;
}
.xm-scan-settings .sc.on {
  background: #38bdf8;
  border-color: #38bdf8;
  color: #04141a;
  font-weight: 700;
}
.xm-scan-settings .ss-loading,
.xm-scan-settings .ss-err {
  margin-top: 14px;
  font-size: 12px;
  color: #8aa6b6;
}
.xm-scan-settings .ss-err {
  color: #fbbf24;
}
.xm-scan-settings .ss-preview {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-top: 16px;
  padding: 10px 12px;
  border: 1px solid rgba(56, 189, 248, 0.22);
  border-radius: 8px;
  background: rgba(18, 32, 48, 0.6);
}
.xm-scan-settings .ss-preview b {
  color: #7dd3fc;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
}
.xm-scan-settings .ss-current {
  margin-top: 10px;
  font-size: 12px;
  color: #9db6c6;
  line-height: 1.6;
  word-break: break-word;
}
.xm-scan-settings .ss-current.muted {
  color: #6f8b9b;
}
.xm-scan-settings .ss-foot {
  flex: none;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-top: 1px solid rgba(56, 189, 248, 0.18);
}
.xm-scan-settings .ss-spacer {
  flex: 1 1 auto;
}
.xm-scan-settings .ss-btn {
  padding: 7px 18px;
  font-size: 12px;
  letter-spacing: 0.06em;
  color: #04141a;
  background: #38bdf8;
  border: 1px solid #38bdf8;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 700;
  font-family: inherit;
  transition: all 0.15s;
}
.xm-scan-settings .ss-btn:hover:not(:disabled) {
  box-shadow: 0 0 14px rgba(56, 189, 248, 0.35);
}
.xm-scan-settings .ss-btn.ghost {
  color: #b9d3e0;
  background: rgba(18, 32, 48, 0.8);
  border-color: rgba(56, 189, 248, 0.25);
  font-weight: 400;
}
.xm-scan-settings .ss-btn:disabled {
  opacity: 0.4;
  cursor: default;
}
.xm-scan-settings .ss-btn:focus-visible,
.xm-scan-settings .ss-x:focus-visible,
.xm-scan-settings .sc:focus-visible {
  outline: 2px solid #38bdf8;
  outline-offset: 2px;
}
@media (max-width: 768px) {
  .xm-scan-settings {
    padding: 4vh 3vw;
  }
  .xm-scan-settings .ss-panel {
    width: 96vw;
    max-height: 86vh;
  }
}
</style>
