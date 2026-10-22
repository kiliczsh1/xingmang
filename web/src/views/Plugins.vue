<template>
  <div class="plugins-container">
    <!-- 顶部标题 + 视图切换 -->
    <div class="header">
      <div class="header-content">
        <h2>插件中心</h2>
        <p class="header-subtitle">内置插件统一管理 · 支持从本地 .zip 安装自研插件</p>
      </div>
      <div class="header-actions">
        <button class="doc-btn" type="button" @click="openApiDoc">
          <el-icon><component :is="'Document'" /></el-icon>查看 API 文档
        </button>
        <div class="view-switch">
          <button class="vs-btn" :class="{ on: view === 'market' }" @click="view = 'market'">
            <el-icon><component :is="'Shop'" /></el-icon>插件市场
          </button>
          <button class="vs-btn" :class="{ on: view === 'diag' }" @click="view = 'diag'">
            <el-icon><component :is="'DataAnalysis'" /></el-icon>运行诊断
          </button>
        </div>
      </div>
    </div>

    <!-- 统计条 -->
    <div class="statusbar">
      <span class="sb-item">内置插件 <b class="blue">{{ builtinCount }}</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">本地安装 <b class="teal">{{ localCount }}</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">第三方插件 <b class="purple">0</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">运行中 <b>{{ runningCount }}</b></span>
      <span class="pill">内置=随客户端 · 本地=zip 导入 · 第三方=待开放</span>
    </div>

    <!-- ============ 市场视图 ============ -->
    <div v-show="view === 'market'" class="market">
      <!-- 搜索 + 联想 -->
      <div class="search-box">
        <el-input
          v-model="query"
          placeholder="搜索插件…（试试输入「世界」）"
          clearable
          class="search-input"
          @input="hl = -1"
          @keydown="onKey"
        >
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <div v-if="suggestions.length && query" class="suggest">
          <div
            v-for="(s, i) in suggestions"
            :key="s.n"
            class="suggest-item"
            :class="{ hl: hl === i }"
            @mousedown.prevent="setQuery(s.n)"
          >
            <span v-html="highlight(s.n, query)"></span>
            <span class="s-src" :class="s.src">{{ srcLabel[s.src] }}</span>
            <span class="s-cat">{{ s.c }}</span>
          </div>
        </div>
      </div>

      <!-- 来源分段：内置 / 本地 / 第三方 -->
      <div class="srcbar">
        <button v-for="s in SRC_OPTIONS" :key="s.key" :class="{ on: src === s.key }" @click="src = s.key">
          {{ s.label }}<span class="n">{{ srcCount(s.key) }}</span>
        </button>
      </div>

      <!-- 搜索历史 -->
      <div v-if="history.length" class="hist">
        <span class="h-label">最近搜索</span>
        <span v-for="h in history" :key="h" class="h-chip" @click="setQuery(h)">{{ h }}</span>
      </div>

      <!-- 标签云 -->
      <div class="tags">
        <span
          v-for="t in allTags"
          :key="t"
          class="tag"
          :class="{ on: tag === t }"
          @click="tag = tag === t ? '' : t"
        >{{ t }}<span class="n">{{ tagCount(t) }}</span></span>
      </div>

      <!-- 分类 Tab -->
      <div class="cat-tabs">
        <button
          v-for="c in cats"
          :key="c"
          class="cat-tab"
          :class="{ on: cat === c }"
          @click="cat = c"
        >{{ c }}<span class="n">{{ catCount(c) }}</span></button>
      </div>

      <!-- 插件卡片网格（C2：≤6 字段；hover 渐进显形 C3） -->
      <div class="plugin-grid">
        <div v-for="p in filteredList" :key="p.src + p.n" class="plugin-card" @click="openDrawer(p)">
          <div class="card-top">
            <span class="p-avatar" :style="{ background: p.color + '1a', color: p.color }">
              <el-icon :size="18"><component :is="p.icon" /></el-icon>
            </span>
            <div class="p-id">
              <span class="p-label">{{ p.n }}</span>
              <span class="p-name">{{ p.c }}</span>
            </div>
            <span class="src-badge" :class="p.src">{{ srcLabel[p.src] }}</span>
          </div>
          <p class="p-desc">{{ p.d }}</p>
          <div class="p-foot">
            <span class="p-state" :class="{ on: p.on }">
              <i class="dot"></i>{{ p.on ? '运行中' : '已停用' }}
            </span>
            <button class="c-btn" @click.stop="openDrawer(p)">管理</button>
          </div>
        </div>
      </div>

      <!-- 空态（C4） -->
      <div v-if="!filteredList.length" class="empty">
        <template v-if="src === 'market'">
          <div class="empty-icon"><el-icon :size="40"><Box /></el-icon></div>
          <p>第三方插件市场开发中</p>
          <p class="empty-sub">后续将开放安装与分享。现在可以先从本地 .zip 安装自研插件。</p>
          <div class="empty-actions">
            <button class="empty-btn primary" @click="triggerPick">安装本地插件</button>
            <a class="empty-link" href="/plugin-demos/" target="_blank" rel="noopener noreferrer">查看演示合集 →</a>
          </div>
        </template>
        <template v-else>
          <div class="empty-icon"><el-icon :size="40"><Box /></el-icon></div>
          <p>没有找到匹配的插件</p>
          <p class="empty-sub">换个关键词，或清除筛选后再试</p>
          <button class="empty-btn" @click="resetFilters">清除筛选</button>
        </template>
      </div>

      <!-- 本地安装 + 第三方市场入口（C4） -->
      <div v-if="src !== 'market'" class="third-empty">
        <span class="te-icon"><el-icon :size="22"><FolderAdd /></el-icon></span>
        <div class="te-body">
          <div class="te-title">安装本地插件</div>
          <p class="te-desc">从本地 .zip 插件包安装（解析 manifest.json + 加载运行代码）。第三方市场后续开放，可先安装自研插件。</p>
        </div>
        <button class="te-btn" @click="triggerPick">选择插件包</button>
        <a class="te-link" href="/plugin-demos/" target="_blank" rel="noopener noreferrer">演示合集</a>
      </div>

      <input ref="fileInput" type="file" accept=".zip,.7z" class="file-input" @change="onPickFile" />
    </div>

    <!-- ============ 诊断视图（透明度 / 可探索性） ============ -->
    <div v-show="view === 'diag'" class="diag">
      <div class="health">
        <div class="ring"><span>74%</span></div>
        <div class="h-copy">
          <div class="h-title">健康度良好</div>
          <div class="h-sub">内置 {{ builtinCount }} · 本地 {{ localCount }} · 运行中 {{ runningCount }}</div>
          <div class="h-sub">最近事件：{{ logs[0]?.msg || '无' }}</div>
        </div>
        <div class="h-score"><b>B+</b><br><small>运行评分</small></div>
      </div>

      <div class="diag-tools">
        <div class="cat-tabs">
          <button
            v-for="l in LVL_FILTERS"
            :key="l.key"
            class="cat-tab"
            :class="{ on: level === l.key }"
            @click="level = l.key"
          >{{ l.label }}<span class="n">{{ lvCount(l.key) }}</span></button>
        </div>
        <button class="copy-btn" @click="copyDiag">⧉ 复制诊断信息</button>
      </div>

      <div class="logs">
        <div v-for="(x, i) in filteredLogs" :key="i" class="log" :class="{ open: openLog === i }">
          <div class="log-head" :class="{ clickable: !!x.stack }" @click="openLog = openLog === i ? -1 : i">
            <span class="lv" :class="x.lv"></span>
            <span class="msg">{{ x.msg }}</span>
            <span class="time">{{ x.time }}</span>
            <span v-if="x.stack" class="expand">▸</span>
          </div>
          <div v-if="x.stack" class="log-body">
            {{ x.msg }}
            <div class="stack">{{ x.stack }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 详情抽屉（渐进披露 · B2 单层） -->
    <el-drawer
      v-model="drawerVisible"
      :size="400"
      :with-header="false"
      :append-to-body="false"
      class="plugin-drawer"
    >
      <div v-if="cur" class="d-wrap">
        <div class="d-head">
          <span class="d-ico" :style="{ background: cur.color + '1a', color: cur.color }">
            <el-icon :size="25"><component :is="cur.icon" /></el-icon>
          </span>
          <div class="d-id">
            <div class="d-name">{{ cur.n }}</div>
            <div class="d-code">{{ curCode }}</div>
          </div>
          <span class="d-src" :class="cur.src">{{ srcLabel[cur.src] }}</span>
        </div>

        <div class="d-body">
          <p class="d-desc">{{ cur.d }}</p>
          <el-collapse v-model="activeSec" class="d-collapse">
            <el-collapse-item title="功能简介" name="func">
              <div v-for="(kv, i) in cur.func" :key="i" class="kv-row"><span>{{ kv[0] }}</span><span>{{ kv[1] }}</span></div>
            </el-collapse-item>
            <el-collapse-item title="权限声明" name="perm">
              <div v-for="(kv, i) in cur.perm" :key="i" class="kv-row"><span>{{ kv[0] }}</span><span>{{ kv[1] }}</span></div>
            </el-collapse-item>
            <el-collapse-item title="版本历史" name="hist">
              <div v-for="(kv, i) in cur.hist" :key="i" class="kv-row"><span>{{ kv[0] }}</span><span>{{ kv[1] }}</span></div>
            </el-collapse-item>
          </el-collapse>
        </div>

        <div class="d-foot">
          <button class="d-primary" @click="drawerPrimary">{{ drawerPrimaryLabel }}</button>
          <button class="d-secondary" :class="{ danger: cur.src === 'local' }" @click="drawerSecondary">
            {{ drawerSecondaryLabel }}
          </button>
        </div>
      </div>
    </el-drawer>

    <!-- API 文档弹窗（Markdown 渲染 + 下载链接） -->
    <el-dialog
      v-model="apiDocVisible"
      title="插件系统 API 参考"
      width="760px"
      top="5vh"
      class="api-doc-dialog"
      :append-to-body="true"
    >
      <div class="api-doc-toolbar">
        <span class="api-doc-meta">Markdown 版 · 随客户端更新</span>
        <a class="api-doc-download" :href="API_DOC_URL" download="plugins-api.md">
          <el-icon><component :is="'Download'" /></el-icon>下载 .md
        </a>
      </div>
      <div v-loading="apiDocLoading" class="api-doc-body">
        <MarkdownRenderer v-if="apiDocContent" :content="apiDocContent" />
        <el-empty v-else-if="!apiDocLoading" description="文档加载失败，请稍后重试" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import JSZip from 'jszip'
import { Search, Box, FolderAdd } from '@element-plus/icons-vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import { BUILTIN_PLUGINS, BUILTIN_META, setBuiltinEnabled } from '@/components/ChatPanel/plugins'
import { LOCAL_PLUGINS, loadLocalPlugins, addLocalPlugin, removeLocalPlugin, setLocalPluginEnabled, getLocalRecords, safeIcon } from '@/components/ChatPanel/plugins'
import { getPluginEnabled } from '@/utils/pluginPrefs'
import type { ChatPlugin } from '@/components/ChatPanel/plugins/type'

/* ========== 类型 ========== */
type Source = 'builtin' | 'local' | 'market'
interface HubItem {
  pname: string          // 真实插件名（name）
  n: string              // 展示名（label）
  src: Source
  c: string
  d: string
  r: number
  color: string
  icon: string
  t: string[]
  kw: string[]
  on: boolean
  func: [string, string][]
  perm: [string, string][]
  hist: [string, string][]
}
interface LogEntry { lv: 'err' | 'warn' | 'info' | 'ok'; msg: string; time: string; stack?: string }

/* ========== 常量 ========== */
const srcLabel: Record<Source, string> = { builtin: '内置', local: '本地', market: '第三方' }
const SRC_OPTIONS = [
  { key: 'all', label: '全部' },
  { key: 'builtin', label: '内置' },
  { key: 'local', label: '本地' },
  { key: 'market', label: '第三方' }
] as const
const LVL_FILTERS = [
  { key: 'all', label: '全部' },
  { key: 'err', label: '错误' },
  { key: 'warn', label: '警告' },
  { key: 'info', label: '信息' }
] as const

/* ========== 内置插件元数据（仅供展示） ========== */
const BUILTIN_DETAIL: Record<string, { func: [string, string][]; perm: [string, string][]; hist: [string, string][] }> = {
  'prompt-inject': {
    func: [['拼装顺序', '提示词 → 关联内容 → 用户输入'], ['关联内容', '章节 / 备忘录'], ['空内容', '自动跳过']],
    perm: [['读取', '选中提示词与关联内容'], ['修改', '仅用户消息（发送前）'], ['网络', '无']],
    hist: [['1.1.2', '支持跨书本提示词模板'], ['1.0.0', '重构为插件链第一环']]
  },
  'worldbook-inject': {
    func: [['匹配方式', '关键词包含匹配（忽略大小写）'], ['注入位置', '用户输入之前'], ['空命中', '不注入，原样放行']],
    perm: [['读取', '当前书本世界书词条'], ['修改', '仅用户消息（发送前）'], ['网络', '无']],
    hist: [['1.0.1', '优化多关键词命中排序'], ['1.0.0', '首次发布']]
  },
  'related-content-inject': {
    func: [['定位', '扩展点（当前 no-op）'], ['关联内容', '已由提示词注入拼装'], ['预留', '未来不同注入策略']],
    perm: [['读取', '关联章节 / 备忘录'], ['修改', '仅用户消息'], ['网络', '无']],
    hist: [['1.0.0', '作为扩展点注册']]
  },
  'regex-filter': {
    func: [['处理阶段', 'AI 响应显示前（显示层）'], ['规则来源', '用户自建正则'], ['扩展点', '聊天链上 no-op']],
    perm: [['读取', 'AI 响应内容'], ['修改', '仅显示内容（不入库）'], ['网络', '无']],
    hist: [['0.9.0', '作为扩展点注册']]
  }
}
const BUILTIN_TAGS: Record<string, string[]> = {
  'prompt-inject': ['注入', '提示词'],
  'worldbook-inject': ['注入', '世界书'],
  'related-content-inject': ['注入', '关联'],
  'regex-filter': ['正则', '过滤']
}
const BUILTIN_KW: Record<string, string[]> = {
  'prompt-inject': ['提示词', '注入', '拼装', '上下文'],
  'worldbook-inject': ['世界', '注入', '词条', '角色'],
  'related-content-inject': ['关联', '注入', '章节', '备忘录'],
  'regex-filter': ['正则', '过滤', '规则', '清洗']
}

/* ========== 状态 ========== */
const view = ref<'market' | 'diag'>('market')
const src = ref<'all' | Source>('all')
const cat = ref('全部')
const tag = ref('')
const query = ref('')
const hl = ref(-1)
const history = ref<string[]>([])
const level = ref<'all' | 'err' | 'warn' | 'info'>('all')
const openLog = ref(-1)
const drawerVisible = ref(false)
const curName = ref('')
const activeSec = ref(['func'])
const fileInput = ref<HTMLInputElement>()
const installing = ref(false)

/* ========== API 文档弹窗（Markdown 渲染 + 下载） ========== */
const API_DOC_URL = '/docs/plugins-api.md'
const apiDocVisible = ref(false)
const apiDocContent = ref('')
const apiDocLoading = ref(false)
const openApiDoc = async () => {
  apiDocVisible.value = true
  if (apiDocContent.value) return
  apiDocLoading.value = true
  try {
    const res = await fetch(API_DOC_URL)
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    apiDocContent.value = await res.text()
  } catch (err: any) {
    apiDocContent.value = ''
    ElMessage.error(`API 文档加载失败：${err?.message || '未知错误'}`)
  } finally {
    apiDocLoading.value = false
  }
}

/* ========== 内置条目（真实来源，静态） ========== */
const builtinItems: HubItem[] = BUILTIN_PLUGINS.map((p: ChatPlugin) => {
  const meta = BUILTIN_META[p.name] || { label: p.name, icon: 'Box', color: '#64748b' }
  const detail = BUILTIN_DETAIL[p.name] || { func: [], perm: [], hist: [] }
  return {
    pname: p.name,
    n: meta.label,
    src: 'builtin',
    c: '内容处理',
    d: p.description,
    r: 5,
    color: meta.color,
    icon: meta.icon,
    t: BUILTIN_TAGS[p.name] || [],
    kw: BUILTIN_KW[p.name] || [],
    on: getPluginEnabled(p.name, p.enabled),
    func: detail.func,
    perm: detail.perm,
    hist: detail.hist,
  }
})

/* ========== 本地条目（zip 导入，随安装/卸载刷新） ========== */
const localItems = ref<HubItem[]>([])
const refreshLocal = () => {
  const records = getLocalRecords()
  const metaByName = new Map(records.map(r => [r.manifest.name, r.manifest]))
  localItems.value = LOCAL_PLUGINS.map(p => {
    const m = metaByName.get(p.name)
    const tags = m?.tags || []
    return {
      pname: p.name,
      n: m?.label || p.name,
      src: 'local',
      c: '本地安装',
      d: p.description,
      r: 5,
      color: m?.color || '#0d9488',
      icon: safeIcon(m?.icon),
      t: tags,
      kw: tags,
      on: p.enabled,
      func: [['来源', `本地 zip 安装`], ['版本', m?.version || '?'], ['作者', m?.author || '未知'], ['入口', m?.entry || '无（扩展点）']],
      perm: [['读取', '本地文件（zip 包）'], ['运行', '本机执行插件代码'], ['网络', '由插件自身决定']],
      hist: [[m?.version || '?', '本地安装']],
    }
  })
}
refreshLocal()

/* ========== 计算属性 ========== */
const items = computed(() => [...builtinItems, ...localItems.value])
const builtinCount = computed(() => builtinItems.length)
const localCount = computed(() => localItems.value.length)
const runningCount = computed(() => items.value.filter(i => i.on).length)
const cats = computed(() => ['全部', ...new Set(items.value.map(i => i.c))])
const allTags = computed(() => [...new Set(items.value.flatMap(i => i.t))])

const filteredList = computed(() => {
  const q = query.value.trim().toLowerCase()
  return items.value.filter(p =>
    (src.value === 'all' || p.src === src.value) &&
    (cat.value === '全部' || p.c === cat.value) &&
    (!tag.value || p.t.includes(tag.value)) &&
    (!q || p.n.toLowerCase().includes(q) || p.kw.some(k => k.includes(q)))
  )
})
const suggestions = computed(() => {
  const q = query.value.trim()
  if (!q) return []
  return items.value.filter(p => p.n.includes(q) || p.kw.some(k => k.includes(q))).slice(0, 5)
})
const cur = computed(() => items.value.find(i => i.pname === curName.value))
const curCode = computed(() => (cur.value?.pname || '').replace(/\s/g, '-').toLowerCase())

const drawerPrimaryLabel = computed(() => {
  if (!cur.value) return ''
  return cur.value.on ? '停用插件' : '启用插件'
})
const drawerSecondaryLabel = computed(() => (cur.value?.src === 'local' ? '卸载' : '取消'))

const filteredLogs = computed(() => {
  return logs.value.filter(x =>
    level.value === 'all' ||
    level.value === 'err' && x.lv === 'err' ||
    level.value === 'warn' && x.lv === 'warn' ||
    level.value === 'info' && (x.lv === 'info' || x.lv === 'ok')
  )
})

/* ========== 计数 ========== */
const srcCount = (key: string) =>
  key === 'all' ? items.value.length : items.value.filter(i => i.src === key).length
const catCount = (c: string) =>
  c === '全部' ? items.value.length : items.value.filter(i => i.c === c).length
const tagCount = (t: string) => items.value.filter(i => i.t.includes(t)).length
const lvCount = (l: string) => {
  if (l === 'all') return logs.value.length
  if (l === 'info') return logs.value.filter(x => x.lv === 'info' || x.lv === 'ok').length
  return logs.value.filter(x => x.lv === l).length
}

/* ========== 抽屉操作 ========== */
const openDrawer = (p: HubItem) => {
  curName.value = p.pname
  activeSec.value = ['func']
  drawerVisible.value = true
}

const drawerPrimary = () => {
  const p = cur.value
  if (!p) return
  p.on = !p.on
  if (p.src === 'builtin') {
    // 真实联动：写回插件单例 + plugin-prefs → 同 SPA 会话内聊天链实时生效
    const bp = BUILTIN_PLUGINS.find(b => BUILTIN_META[b.name]?.label === p.n)
    if (bp) setBuiltinEnabled(bp.name, p.on)
  } else if (p.src === 'local') {
    setLocalPluginEnabled(p.pname, p.on)
  }
  log(p.on ? 'info' : 'warn', `「${p.n}」已${p.on ? '启用' : '停用'}（${p.on ? '运行中' : '已停用'}）`)
  ElMessage.success(`「${p.n}」已${p.on ? '启用' : '停用'}`)
}

const drawerSecondary = () => {
  const p = cur.value
  if (!p) return
  if (p.src === 'local') {
    removeLocalPlugin(p.pname)
    refreshLocal()
    log('warn', `本地插件「${p.n}」已卸载`)
    ElMessage.success(`已卸载「${p.n}」`)
    drawerVisible.value = false
  } else {
    drawerVisible.value = false
  }
}

/* ========== 本地安装（zip / 7z） ========== */
const triggerPick = () => fileInput.value?.click()

const onPickFile = async (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = '' // 允许重复选择同一文件
  if (!file || installing.value) return

  const lower = file.name.toLowerCase()
  if (lower.endsWith('.7z')) {
    ElMessage.warning('7z 暂不支持在线解析，请用 7-Zip 将插件包导出为 zip 后导入')
    return
  }
  if (!lower.endsWith('.zip')) {
    ElMessage.warning('请选择 .zip 或 .7z 插件包')
    return
  }

  installing.value = true
  try {
    const zip = await JSZip.loadAsync(await file.arrayBuffer())
    const manifestFile = zip.file('manifest.json')
    if (!manifestFile) {
      ElMessage.error('插件包缺少 manifest.json（包格式见 API 文档）')
      return
    }
    const manifest = JSON.parse(await manifestFile.async('string'))
    if (!manifest?.name || !manifest?.label || !manifest?.version) {
      ElMessage.error('manifest.json 字段缺失：name / label / version 为必填')
      return
    }

    // 加载入口代码（可选）：entry 或 main 字段均支持，忽略前导 ./
    let entryCode = ''
    const entryName = manifest.entry || manifest.main
    if (entryName) {
      const cleanName = String(entryName).replace(/^\.\//, '')
      const entryFile = zip.file(cleanName)
      if (entryFile) {
        entryCode = await entryFile.async('string')
      } else {
        ElMessage.warning(`入口文件 ${entryName} 不存在，插件将以扩展点注册（无运行代码）`)
      }
    }

    // 含运行代码：本机执行需确认
    if (entryCode.trim()) {
      try {
        await ElMessageBox.confirm(
          '该插件包含运行代码，安装后将在本机执行。请确保插件包来源可信。',
          '安装本地插件',
          { confirmButtonText: '安装', cancelButtonText: '取消', type: 'warning' }
        )
      } catch {
        return // 用户取消
      }
    }

    const plugin = addLocalPlugin({
      manifest: {
        name: String(manifest.name),
        label: String(manifest.label),
        version: String(manifest.version),
        description: String(manifest.description || ''),
        author: manifest.author ? String(manifest.author) : undefined,
        icon: safeIcon(manifest.icon),
        color: typeof manifest.color === 'string' ? manifest.color : undefined,
        tags: Array.isArray(manifest.tags) ? manifest.tags.map(String) : undefined,
        entry: manifest.entry ? String(manifest.entry) : undefined,
        main: manifest.main ? String(manifest.main) : undefined,
      },
      entryCode,
      installedAt: new Date().toISOString(),
    })
    if (!plugin) {
      ElMessage.error('插件注册失败')
      return
    }
    refreshLocal()
    log('info', `已从本地安装插件「${manifest.label}」${entryCode.trim() ? '（含运行代码）' : '（仅扩展点）'}`)
    ElMessage.success(`「${manifest.label}」已安装`)
  } catch (err: any) {
    ElMessage.error(`插件包解析失败：${err?.message || '未知错误'}`)
  } finally {
    installing.value = false
  }
}

/* ========== 搜索 ========== */
const setQuery = (v: string) => {
  query.value = v
  hl.value = -1
  if (!history.value.includes(v)) {
    history.value.unshift(v)
    if (history.value.length > 4) history.value.pop()
  }
}

const onKey = (e: KeyboardEvent) => {
  const n = suggestions.value.length
  if (!n) return
  if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    e.preventDefault()
    hl.value = (hl.value + (e.key === 'ArrowDown' ? 1 : -1) + n) % n
  } else if (e.key === 'Enter' && hl.value >= 0) {
    setQuery(suggestions.value[hl.value].n)
  } else if (e.key === 'Escape') {
    query.value = ''
    hl.value = -1
  }
}

const highlight = (s: string, q: string) => {
  const i = s.indexOf(q)
  if (i < 0) return s
  return s.slice(0, i) + `<b>${q}</b>` + s.slice(i + q.length)
}

const resetFilters = () => {
  query.value = ''
  tag.value = ''
  cat.value = '全部'
  src.value = 'all'
}

/* ========== 诊断 ========== */
const logs = ref<LogEntry[]>([])
const now = () => new Date().toTimeString().slice(0, 8)
const log = (lv: LogEntry['lv'], msg: string, stack?: string) => {
  logs.value.unshift({ lv, msg, time: now(), stack })
  if (logs.value.length > 40) logs.value.pop()
}

const copyDiag = async () => {
  const text = [
    '=== 星芒插件诊断信息 ===',
    `时间: ${new Date().toLocaleString('zh-CN')}`,
    ...items.value.map(p => `[${srcLabel[p.src]}] ${p.on ? '运行' : '停用'} ${p.n}`),
    '',
    ...logs.value.map(x => `[${x.lv.toUpperCase()}] ${x.time} ${x.msg}`)
  ].join('\n')
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('诊断信息已复制')
  } catch {
    ElMessage.warning('复制失败，请手动选择')
  }
}

/* ========== 初始化 ========== */
onMounted(() => {
  loadLocalPlugins()
  refreshLocal()
  log('ok', `插件链自检通过：内置 ${builtinCount.value} 个 + 本地 ${localCount.value} 个`)
  const overrides = (() => {
    try {
      const prefs = JSON.parse(localStorage.getItem('plugin-prefs') || '{}')
      return Object.keys(prefs).filter(k => typeof prefs[k] === 'boolean').length
    } catch {
      return 0
    }
  })()
  log('info', `已从本地恢复插件偏好 plugin-prefs（${overrides} 项覆盖）`)
})
</script>

<style scoped>
.plugins-container {
  padding: 24px;
  /* 消费全局主题变量（MainLayout 注入） */
  --pg-ink: var(--app-color, #1a2230);
  --pg-ink-2: var(--ink2, #5a6a80);
  --pg-ink-3: var(--ink3, #98a5b8);
  --pg-line: var(--topbar-border, #e6eaf0);
  --pg-card: var(--card, #ffffff);
  --pg-accent: var(--accent, #2563eb);
  --pg-accent-soft: var(--accent-soft, #eef4ff);
  --pg-third: #7c3aed;
  --pg-third-bg: #f5f3ff;
  --pg-local: #0d9488;
  --pg-local-bg: #f0fdfa;
  --pg-shadow: var(--shadow, 0 1px 2px rgba(30, 36, 51, .04), 0 10px 30px rgba(30, 36, 51, .05));
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 14px;
}
.header-content h2 { margin: 0 0 6px 0; font-size: 24px; color: var(--pg-ink); }
.header-subtitle { margin: 0; color: var(--pg-ink-2); font-size: 14px; }

.header-actions { display: flex; align-items: center; gap: 10px; }
.doc-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 38px;
  padding: 0 14px;
  border: 1px solid var(--pg-line);
  border-radius: 10px;
  background: var(--pg-card);
  color: var(--pg-ink-2);
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: .15s;
  box-shadow: var(--pg-shadow);
}
.doc-btn:hover { border-color: var(--pg-accent); color: var(--pg-accent); }
.doc-btn .el-icon { font-size: 15px; }

.view-switch {
  display: flex;
  gap: 4px;
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 11px;
  padding: 4px;
  box-shadow: var(--pg-shadow);
}
.vs-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: 0;
  background: transparent;
  color: var(--pg-ink-2);
  border-radius: 8px;
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: .15s;
}
.vs-btn.on { background: var(--pg-accent); color: #fff; }

/* 状态栏 */
.statusbar {
  display: flex;
  gap: 18px;
  align-items: center;
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 12px;
  padding: 12px 18px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--pg-ink-2);
  flex-wrap: wrap;
  box-shadow: var(--pg-shadow);
}
.statusbar .sb-item b { color: var(--pg-ink); }
.statusbar .sb-item b.blue { color: var(--pg-accent); }
.statusbar .sb-item b.teal { color: var(--pg-local); }
.statusbar .sb-item b.purple { color: var(--pg-third); }
.statusbar .sb-sep { width: 1px; height: 18px; background: var(--pg-line); }
.statusbar .pill {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--pg-accent-soft);
  color: var(--pg-accent);
  border-radius: 999px;
  padding: 4px 12px;
  font-weight: 600;
  font-size: 12px;
}

/* ---------- 市场视图 ---------- */
.search-box { position: relative; margin-bottom: 12px; }
.search-input :deep(.el-input__wrapper) {
  border-radius: 12px;
  box-shadow: 0 0 0 2px var(--pg-line) inset;
  height: 46px;
  background: var(--pg-card);
}
.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px var(--pg-accent) inset;
}
.search-input :deep(.el-input__inner) { color: var(--pg-ink); }

.suggest {
  position: absolute;
  top: 54px;
  left: 0;
  right: 0;
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 12px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, .14);
  overflow: hidden;
  z-index: 30;
}
.suggest-item {
  padding: 11px 16px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  gap: 10px;
  align-items: center;
  color: var(--pg-ink);
}
.suggest-item:hover, .suggest-item.hl { background: var(--pg-accent-soft); color: var(--pg-accent); }
.suggest-item b { color: var(--pg-accent); }
.s-src { margin-left: 8px; font-size: 10.5px; font-weight: 600; padding: 1px 7px; border-radius: 999px; flex-shrink: 0; }
.s-src.builtin { background: var(--pg-accent-soft); color: var(--pg-accent); }
.s-src.local { background: var(--pg-local-bg); color: var(--pg-local); }
.s-src.market { background: var(--pg-third-bg); color: var(--pg-third); }
.s-cat { margin-left: auto; font-size: 11px; color: var(--pg-ink-3); }

/* 来源分段（内置 / 本地 / 第三方） */
.srcbar {
  display: flex;
  gap: 4px;
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 10px;
  padding: 4px;
  width: fit-content;
  margin-bottom: 12px;
  box-shadow: var(--pg-shadow);
}
.srcbar button {
  border: 0;
  background: transparent;
  color: var(--pg-ink-2);
  border-radius: 7px;
  padding: 6px 16px;
  font-size: 12.5px;
  font-weight: 600;
  cursor: pointer;
  transition: .15s;
}
.srcbar button.on { background: var(--pg-accent); color: #fff; }
.srcbar .n { font-size: 11px; opacity: .75; margin-left: 4px; }

/* 历史 */
.hist { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin-bottom: 12px; }
.h-label { font-size: 12px; color: var(--pg-ink-3); }
.h-chip {
  padding: 4px 11px;
  border-radius: 8px;
  background: var(--pg-accent-soft);
  color: var(--pg-accent);
  font-size: 12px;
  cursor: pointer;
}
.h-chip:hover { filter: brightness(.95); }

/* 标签云 */
.tags { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 12px; }
.tag {
  padding: 5px 12px;
  border-radius: 999px;
  border: 1px solid var(--pg-line);
  background: var(--pg-card);
  color: var(--pg-ink-2);
  font-size: 12.5px;
  cursor: pointer;
  transition: .15s;
}
.tag:hover { border-color: var(--pg-accent); color: var(--pg-accent); }
.tag.on { background: var(--pg-accent); border-color: var(--pg-accent); color: #fff; }
.tag .n { font-size: 11px; opacity: .7; }

/* 分类 Tab */
.cat-tabs { display: flex; gap: 6px; margin-bottom: 16px; flex-wrap: wrap; }
.cat-tab {
  padding: 7px 16px;
  border-radius: 999px;
  border: 1px solid var(--pg-line);
  background: var(--pg-card);
  color: var(--pg-ink-2);
  font-size: 13px;
  cursor: pointer;
  transition: .15s;
}
.cat-tab:hover { border-color: var(--pg-accent); color: var(--pg-accent); }
.cat-tab.on { background: var(--pg-accent); border-color: var(--pg-accent); color: #fff; font-weight: 600; }
.cat-tab .n { font-size: 11px; opacity: .7; margin-left: 4px; }

/* 卡片网格（C2：≤6 字段） */
.plugin-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 14px;
}
.plugin-card {
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 14px;
  padding: 15px;
  display: flex;
  flex-direction: column;
  gap: 9px;
  cursor: pointer;
  transition: .18s;
  box-shadow: var(--pg-shadow);
}
.plugin-card:hover { transform: translateY(-3px); border-color: color-mix(in srgb, var(--pg-accent) 35%, var(--pg-line)); }
.card-top { display: flex; align-items: center; gap: 10px; }
.p-avatar {
  width: 38px;
  height: 38px;
  border-radius: 11px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.p-id { min-width: 0; }
.p-label { font-size: 14px; font-weight: 700; color: var(--pg-ink); display: block; }
.p-name { font-size: 11px; color: var(--pg-ink-3); }
.src-badge {
  font-size: 10.5px;
  font-weight: 700;
  padding: 1.5px 8px;
  border-radius: 999px;
  flex-shrink: 0;
  margin-left: auto;
}
.src-badge.builtin { background: var(--pg-accent-soft); color: var(--pg-accent); }
.src-badge.local { background: var(--pg-local-bg); color: var(--pg-local); }
.src-badge.market { background: var(--pg-third-bg); color: var(--pg-third); }
.p-desc {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--pg-ink-2);
  flex: 1;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.p-foot { display: flex; align-items: center; justify-content: space-between; }
.p-state { display: inline-flex; align-items: center; gap: 6px; font-size: 11.5px; font-weight: 600; color: var(--pg-ink-3); }
.p-state .dot { width: 7px; height: 7px; border-radius: 50%; background: var(--pg-ink-3); }
.p-state.on { color: #16a34a; }
.p-state.on .dot { background: #16a34a; }
/* 操作按钮 hover 渐进显形（C3） */
.c-btn {
  opacity: 0;
  border: 0;
  background: var(--pg-accent);
  color: #fff;
  border-radius: 8px;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: .15s;
  pointer-events: none;
}
.plugin-card:hover .c-btn { opacity: 1; pointer-events: auto; }
.c-btn:hover { filter: brightness(.92); }

/* 空态（C4） */
.empty { text-align: center; padding: 52px 20px; color: var(--pg-ink-3); }
.empty-icon {
  width: 84px;
  height: 84px;
  border-radius: 20px;
  background: var(--pg-accent-soft);
  color: var(--pg-accent);
  display: grid;
  place-items: center;
  margin: 0 auto 14px;
}
.empty p { font-size: 13px; line-height: 1.8; }
.empty-sub { font-size: 12px; }
.empty-actions { display: flex; gap: 10px; justify-content: center; margin-top: 16px; flex-wrap: wrap; }
.empty-btn {
  border: 1px solid var(--pg-accent);
  color: var(--pg-accent);
  background: transparent;
  border-radius: 9px;
  padding: 8px 18px;
  font-size: 13px;
  cursor: pointer;
}
.empty-btn.primary {
  background: var(--pg-accent);
  color: #fff;
  font-weight: 600;
}
.empty-btn.primary:hover { filter: brightness(.92); }
.empty-link {
  display: inline-flex;
  align-items: center;
  font-size: 13px;
  font-weight: 600;
  color: var(--pg-accent);
  text-decoration: none;
  padding: 8px 0;
}
.empty-link:hover { text-decoration: underline; }

/* 本地安装 + 第三方市场入口（C4） */
.third-empty {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-top: 18px;
  background: var(--pg-card);
  border: 1px dashed var(--pg-line);
  border-radius: 12px;
  padding: 16px 18px;
  flex-wrap: wrap;
}
.te-icon {
  width: 42px;
  height: 42px;
  border-radius: 11px;
  background: var(--pg-local-bg);
  color: var(--pg-local);
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.te-body { flex: 1; min-width: 200px; }
.te-title { font-size: 14px; font-weight: 700; color: var(--pg-ink); }
.te-desc { margin: 3px 0 0; font-size: 12.5px; color: var(--pg-ink-2); line-height: 1.7; }
.te-btn {
  flex-shrink: 0;
  border: 0;
  background: var(--pg-local);
  color: #fff;
  border-radius: 9px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
}
.te-btn:hover { filter: brightness(.92); }
.te-link {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--pg-accent);
  text-decoration: none;
}
.te-link:hover { text-decoration: underline; }
.file-input { display: none; }

/* ---------- 诊断视图 ---------- */
.health {
  display: flex;
  align-items: center;
  gap: 16px;
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 12px;
  padding: 15px 18px;
  margin-bottom: 14px;
  box-shadow: var(--pg-shadow);
}
.ring {
  position: relative;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  background: conic-gradient(#16a34a 0 74%, var(--pg-line) 74% 100%);
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.ring::after { content: ""; width: 38px; height: 38px; border-radius: 50%; background: var(--pg-card); }
.ring span { position: absolute; font-size: 11px; font-weight: 800; color: var(--pg-ink); }
.h-copy { flex: 1; min-width: 0; }
.h-title { font-size: 14px; font-weight: 800; color: var(--pg-ink); }
.h-sub { font-size: 12px; color: var(--pg-ink-2); line-height: 1.7; }
.h-score { text-align: center; flex-shrink: 0; color: var(--pg-ink-3); }
.h-score b { font-size: 20px; color: #16a34a; }
.diag-tools {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 8px;
}
.diag-tools .cat-tabs { margin-bottom: 0; }
.copy-btn {
  border: 1px solid var(--pg-line);
  background: var(--pg-card);
  color: var(--pg-ink-2);
  border-radius: 9px;
  padding: 7px 14px;
  font-size: 12.5px;
  cursor: pointer;
}
.copy-btn:hover { border-color: var(--pg-accent); color: var(--pg-accent); }

.log { background: var(--pg-card); border: 1px solid var(--pg-line); border-radius: 10px; margin-bottom: 8px; overflow: hidden; }
.log-head { display: flex; align-items: center; gap: 12px; padding: 11px 14px; }
.log-head.clickable { cursor: pointer; }
.log-head.clickable:hover { background: var(--pg-accent-soft); }
.lv { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.lv.err { background: #dc2626; }
.lv.warn { background: #d97706; }
.lv.info { background: #0ea5e9; }
.lv.ok { background: #16a34a; }
.log .msg { font-size: 12.5px; color: var(--pg-ink); flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.log .time { font-size: 11px; color: var(--pg-ink-3); font-family: monospace; flex-shrink: 0; }
.log .expand { color: var(--pg-ink-3); font-size: 11px; transition: transform .2s; flex-shrink: 0; }
.log.open .expand { transform: rotate(90deg); }
.log-body {
  display: none;
  padding: 4px 14px 13px;
  font-size: 11.5px;
  color: var(--pg-ink-2);
  line-height: 1.8;
  font-family: Consolas, monospace;
  background: var(--pg-accent-soft);
  border-top: 1px dashed var(--pg-line);
}
.log.open .log-body { display: block; }
.log-body .stack {
  margin-top: 7px;
  background: var(--pg-card);
  border: 1px solid var(--pg-line);
  border-radius: 8px;
  padding: 9px 11px;
  color: var(--pg-ink-3);
  white-space: pre-wrap;
  word-break: break-all;
}
</style>

<!-- 抽屉样式：el-drawer 使用 teleport 到 body，scoped 不生效，单独全局作用域 -->
<style>
.plugin-drawer .el-drawer {
  background: var(--card, #ffffff);
  --el-drawer-bg-color: var(--card, #ffffff);
}
.plugin-drawer .d-wrap { display: flex; flex-direction: column; height: 100%; }
.plugin-drawer .d-head {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px 22px 14px;
  border-bottom: 1px solid var(--topbar-border, #e6eaf0);
}
.plugin-drawer .d-ico {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  flex-shrink: 0;
}
.plugin-drawer .d-id { min-width: 0; }
.plugin-drawer .d-name { font-size: 17px; font-weight: 800; color: var(--app-color, #1a2230); }
.plugin-drawer .d-code { font-size: 11px; color: var(--ink3, #98a5b8); font-family: monospace; margin-top: 2px; }
.plugin-drawer .d-src {
  margin-left: auto;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 11px;
  border-radius: 999px;
  flex-shrink: 0;
}
.plugin-drawer .d-src.builtin { background: var(--accent-soft, #eef4ff); color: var(--accent, #2563eb); }
.plugin-drawer .d-src.local { background: #f0fdfa; color: #0d9488; }
.plugin-drawer .d-src.market { background: #f5f3ff; color: #7c3aed; }
.plugin-drawer .d-body { flex: 1; overflow-y: auto; padding: 16px 22px; }
.plugin-drawer .d-desc { font-size: 13px; color: var(--ink2, #5a6a80); line-height: 1.8; margin-bottom: 14px; }
.plugin-drawer .d-collapse { border: 0; }
.plugin-drawer .d-collapse .el-collapse-item__header {
  background: var(--accent-soft, #eef4ff);
  border-radius: 10px;
  padding: 0 14px;
  height: auto;
  min-height: 38px;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--app-color, #1a2230);
  border-bottom: 0;
  margin-bottom: 9px;
}
.plugin-drawer .d-collapse .el-collapse-item__content {
  font-size: 12px;
  color: var(--ink2, #5a6a80);
  padding: 2px 4px 12px;
  line-height: 1.85;
}
.plugin-drawer .kv-row {
  display: flex;
  justify-content: space-between;
  padding: 3.5px 0;
  border-bottom: 1px dashed var(--topbar-border, #e6eaf0);
}
.plugin-drawer .kv-row:last-child { border-bottom: 0; }
.plugin-drawer .kv-row span:first-child { color: var(--ink3, #98a5b8); flex-shrink: 0; margin-right: 12px; }
.plugin-drawer .kv-row span:last-child { text-align: right; }
.plugin-drawer .d-foot {
  border-top: 1px solid var(--topbar-border, #e6eaf0);
  padding: 15px 22px;
  display: flex;
  gap: 10px;
}
.plugin-drawer .d-primary {
  flex: 1;
  border: 0;
  background: var(--accent, #2563eb);
  color: #fff;
  border-radius: 10px;
  padding: 12px 0;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}
.plugin-drawer .d-primary:hover { filter: brightness(.92); }
.plugin-drawer .d-secondary {
  border: 1px solid var(--topbar-border, #e6eaf0);
  background: transparent;
  color: var(--ink2, #5a6a80);
  border-radius: 10px;
  padding: 0 18px;
  cursor: pointer;
}
.plugin-drawer .d-secondary.danger {
  color: #dc2626;
  border-color: #fecaca;
  background: #fef2f2;
}
.plugin-drawer .d-secondary.danger:hover { background: #fee2e2; }

/* API 文档弹窗（el-dialog teleport 到 body，scoped 不生效，全局作用域） */
.api-doc-dialog .el-dialog__header {
  border-bottom: 1px solid var(--topbar-border, #e6eaf0);
  margin-right: 0;
  padding: 16px 20px;
}
.api-doc-dialog .el-dialog__title {
  font-size: 16px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
}
.api-doc-dialog .el-dialog__body { padding: 14px 20px 20px; }
.api-doc-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}
.api-doc-meta { font-size: 12px; color: var(--ink3, #98a5b8); }
.api-doc-download {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 13px;
  font-weight: 600;
  color: var(--accent, #2563eb);
  text-decoration: none;
}
.api-doc-download:hover { text-decoration: underline; }
.api-doc-body {
  max-height: 62vh;
  overflow-y: auto;
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 10px;
  padding: 16px 18px;
  background: var(--card, #ffffff);
}
</style>
