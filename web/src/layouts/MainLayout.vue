<template>
  <div class="app-shell">
    <header class="topbar" v-if="!isHiddenTopbarRoute">
      <div class="brand-block">
        <img class="brand-mark" src="/brand/mark-32.png" width="32" height="32" alt="">
        <div class="brand-copy">
          <h1>星芒 AI 写作</h1>
          <span class="chip">创作工作台</span>
        </div>
      </div>

      <nav class="topbar-actions" aria-label="全局导航">
        <a
          href="https://www.123912.com/s/cc0dTd-GDNGh"
          target="_blank"
          class="topbar-action"
          rel="noopener noreferrer"
        >
          <el-icon><Download /></el-icon>
          <span>最新下载</span>
        </a>
        <div class="theme-dropdown-wrapper" ref="themeDropdownWrapperRef">
          <button
            class="topbar-action theme-toggle-btn"
            type="button"
            @click.stop="toggleThemeDropdown"
            :title="`当前主题：${themeModeLabel}`"
          >
            <el-icon><component :is="themeModeIcon" /></el-icon>
            <span>外观</span>
          </button>
          <div class="theme-dropdown-panel" :class="{ visible: themeDropdownVisible }">
            <div class="theme-panel-header">
              <span class="theme-panel-title">外观</span>
              <span class="theme-panel-sub">主题与皮肤</span>
            </div>
            <div class="theme-panel-section">
              <div class="theme-section-label">主题</div>
              <div class="theme-mode-row">
                <button
                  v-for="opt in themeModeOptions"
                  :key="opt.value"
                  class="theme-mode-chip"
                  :class="{ active: themeMode === opt.value }"
                  type="button"
                  @click="selectThemeMode(opt.value)"
                >
                  <el-icon><component :is="opt.icon" /></el-icon>
                  <span>{{ opt.label }}</span>
                </button>
              </div>
            </div>
            <div class="theme-panel-section">
              <div class="theme-section-label">护眼模式</div>
              <div class="theme-eyecare-row">
                <span class="theme-eyecare-hint">开启后亮色界面偏暖，降低蓝光</span>
                <el-switch
                  v-model="eyeCareEnabled"
                  size="small"
                  @change="toggleEyeCare"
                />
              </div>
            </div>
            <div class="theme-panel-section">
              <div class="theme-section-label">皮肤</div>
              <div class="theme-skin-row">
                <button
                  v-for="skin in skinOptions"
                  :key="skin.value"
                  class="theme-skin-swatch"
                  :class="{ active: skinValue === skin.value }"
                  type="button"
                  :title="skin.label"
                  :style="{ background: skin.color }"
                  @click="selectSkin(skin.value)"
                >
                  <el-icon v-if="skinValue === skin.value" class="theme-skin-check"><Check /></el-icon>
                </button>
              </div>
            </div>
          </div>
        </div>
        <button
          v-for="item in globalNavItems"
          :key="item.label"
          class="topbar-action"
          type="button"
          @click="handleGlobalNavClick(item.label)"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </button>
      </nav>

      <div class="user-badge-wrapper" @click.stop="toggleUserPopup">
        <div class="user-badge">
          <img v-if="avatarUrl" :src="avatarUrl" class="badge-avatar" />
          <span v-else>我</span>
        </div>
        <div class="user-popup" :class="{ visible: userPopupVisible }">
          <div class="profile-card">
            <div class="p-avatar">
              <img v-if="avatarUrl" :src="avatarUrl" />
              <span v-else>我</span>
            </div>
            <div class="p-name">{{ userName }}</div>
            <div class="p-bio">{{ userBio || '这个人很懒，还没有写介绍' }}</div>
          </div>
          <div class="info-list">
            <div class="info-item"><label>机器码</label><span>{{ machineId }}</span></div>
            <div class="info-item"><label>模型调用</label><span>{{ modelCallCount }} 次</span></div>
            <div class="info-item"><label>提示词</label><span>{{ promptCount }} 条</span></div>
            <div class="info-item"><label>登录时间</label><span class="bio-preview">{{ loginTime }}</span></div>
          </div>
          <button class="profile-edit-trigger" type="button" @click="openProfileEdit">编辑个人资料</button>
          <button class="app-exit-trigger" type="button" @click="handleExitApp">退出程序</button>
        </div>
      </div>
    </header>

    <!-- 个人资料编辑弹窗 -->
    <el-dialog v-model="profileEditVisible" title="编辑个人资料" width="520px" class="profile-edit-dialog" append-to-body destroy-on-close>
      <div class="profile-edit-form">
        <div class="form-avatar-row">
          <label>头像</label>
          <div class="avatar-upload" @click="triggerAvatarUpload">
            <img v-if="editAvatarUrl" :src="editAvatarUrl" class="avatar-preview" />
            <div v-else class="avatar-placeholder">
              <el-icon :size="28"><Camera /></el-icon>
            </div>
            <div class="avatar-overlay">
              <el-icon :size="16"><Edit /></el-icon>
            </div>
          </div>
          <input ref="avatarInputRef" type="file" accept="image/*" class="avatar-file-input" @change="handleAvatarUpload" />
          <span class="avatar-hint">点击更换头像</span>
        </div>
        <div class="form-item">
          <label>用户昵称</label>
          <el-input
            v-model="editUserName"
            placeholder="请输入昵称"
            maxlength="50"
            show-word-limit
          />
        </div>
        <div class="form-item">
          <label>个人介绍</label>
          <el-input
            type="textarea"
            v-model="editUserBio"
            placeholder="介绍一下你自己……"
            maxlength="500"
            show-word-limit
            :rows="5"
          />
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="cancelProfileEdit">取消</el-button>
          <el-button type="primary" @click="saveProfile">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 历史记录中心 -->
    <HistoryCenter v-model="historyDialogVisible" />

    <!-- 对话 = 星芒 Agent（大矩形弹窗，与侧栏 /agent 同一份数据） -->
    <AgentDialog v-model:visible="chatDialogVisible" />

    <!-- 消息中心 -->
    <MessageCenter v-model:visible="messageCenterVisible" />

    <!-- 更新日志（简略版 / 详略版） -->
    <ChangelogDialog v-model:visible="changelogVisible" :current-version="appVersion" />

    <div class="workspace-shell">
      <aside class="sidebar">
        <div class="sidebar-panel">
          <el-menu
            :default-active="activeMenu"
            router
            class="sidebar-menu"
          >
            <el-menu-item
              v-for="item in menuItems"
              :key="item.index"
              :index="item.index"
            >
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ item.label }}</span>
            </el-menu-item>
          </el-menu>
          <div class="side-foot">
            <div class="foot-row"><span>模型调用</span><b>{{ modelCallCount }}</b></div>
            <div class="foot-row"><span>提示词</span><b>{{ promptCount }}</b></div>
            <div class="foot-row foot-ver" :title="appVersionTitle">
              <span>版本</span><b>{{ versionLabel }}</b>
            </div>
          </div>
        </div>
      </aside>

      <main class="content-shell">
        <section class="content-frame">
          <router-view />
        </section>
      </main>
    </div>

    <!-- 退出后的全屏提示：后端已停止，页面已失效，必须明确告知用户
         （否则用户会以为界面卡死，继续点击然后看到一堆请求失败） -->
    <div v-if="appExited" class="exit-overlay">
      <div class="exit-card">
        <div class="exit-mark">✓</div>
        <h3>程序已退出</h3>
        <p>后台服务已停止，现在可以关闭此页面。</p>
        <p class="exit-hint">下次使用请重新双击 xingmang.exe</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Reading,
  ChatDotSquare,
  View,
  Star,
  Setting,
  User,
  ChatLineRound,
  Clock,
  Wallet,
  Message,
  Check,
  Download,
  Edit,
  EditPen,
  Grid,
  Timer,
  List,
  DataAnalysis,
  Monitor,
  Box,
  Camera,
  Share,
  MagicStick,
  Connection,
  Document,
  Notebook
} from '@element-plus/icons-vue'
import HistoryCenter from '../components/HistoryCenter.vue'
import MessageCenter from '../components/MessageCenter.vue'
import ChangelogDialog from '../components/ChangelogDialog.vue'
import AgentDialog from '../components/AgentDialog.vue'
import { statsAPI, providerAPI, configAPI, promptAPI, systemAPI } from '@/api'
import type { UsageOverview, ApiModel } from '@/types'

const route = useRoute()

type ThemeMode = 'light' | 'dark' | 'system'
type ThemeType = 'light' | 'dark'

// 这些路径下不显示顶部栏（折叠状态和展开状态都隐藏）
// 包括：使用 FullScreenLayout 的全屏页 + MainLayout 下需要独立空间的拆书库/角色库
const HIDDEN_TOPBAR_ROUTE_PATTERNS = [
  /^\/analysis\/[^/]+(\/|$)/,
  /^\/characters\/[^/]+(\/|$)/,
  /^\/write(\/|$)/,
  /^\/book-analysis-prompts(\/|$)/
]

const isHiddenTopbarRoute = computed(() => {
  return HIDDEN_TOPBAR_ROUTE_PATTERNS.some(pattern => pattern.test(route.path))
})

const getInitialThemeMode = (): ThemeMode => {
  const saved = localStorage.getItem('main-theme-mode')
  if (saved === 'light' || saved === 'dark' || saved === 'system') {
    return saved
  }
  // 兼容旧版本只存 main-theme（light/dark）的用户
  const legacy = localStorage.getItem('main-theme')
  if (legacy === 'light' || legacy === 'dark') {
    return legacy
  }
  return 'light'
}

const getSystemTheme = (): ThemeType => {
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

const resolveTheme = (mode: ThemeMode): ThemeType => {
  return mode === 'system' ? getSystemTheme() : mode
}

const chatDialogVisible = ref(false)
const themeDropdownVisible = ref(false)
const themeMode = ref<ThemeMode>(getInitialThemeMode())
const currentThemeType = ref<ThemeType>(resolveTheme(themeMode.value))
const themeDropdownWrapperRef = ref<HTMLElement>()
const historyDialogVisible = ref(false)
const messageCenterVisible = ref(false)
const changelogVisible = ref(false)
const userPopupVisible = ref(false)
const userName = ref('星芒创作者')
const userBio = ref('')
const machineId = ref('12345678')

const loadUserProfile = () => {
  const stored = localStorage.getItem('userName')
  if (stored) {
    userName.value = stored
  }
  const storedBio = localStorage.getItem('userBio')
  if (storedBio) {
    userBio.value = storedBio
  }
  const storedAvatar = localStorage.getItem('userAvatar')
  if (storedAvatar) {
    avatarUrl.value = storedAvatar
  }
}
const modelCallCount = ref(0)
const promptCount = ref(0)
const overview = ref<UsageOverview | null>(null)
const loginTime = ref('')
const models = ref<ApiModel[]>([])

// 当前后端版本。用于让用户自查「我更新了没有」——
// 前端产物是内嵌在后端 exe 里的，只比对界面看不出版本差异。
const appVersion = ref('')
const appVersionTitle = ref('')

// 后端未上报版本（老版本）时显示「未知」，不能留空让用户以为坏了
const versionLabel = computed(() => appVersion.value || '未知')

// 版本号优先显示打包时指定的 -Version；只有 dev/提交号时退回短 commit，避免显示成串哈希
const formatVersion = (version?: string, commit?: string) => {
  const v = (version || '').trim()
  if (!v) return ''
  if (v === 'dev') {
    const c = (commit || '').trim()
    return c && c !== 'unknown' ? `dev(${c.slice(0, 7)})` : 'dev'
  }
  return v.startsWith('v') ? v : `v${v}`
}

const loadAppVersion = async () => {
  try {
    const health = await systemAPI.getHealth()
    appVersion.value = formatVersion(health?.version, health?.commit)
    appVersionTitle.value = [
      `commit: ${health?.commit || 'unknown'}`,
      `构建时间: ${health?.buildTime || 'unknown'}`
    ].join('\n')
  } catch {
    // 老版本后端没有 version 字段 / 请求失败：保持「未知」，不打扰用户
    appVersion.value = ''
  }
}

// 退出程序：真正的进程退出（不是关页面）
//
// 后端收到请求后会优雅关闭服务并结束进程；响应先于关闭发出，因此这里能拿到结果。
// 退出后必须盖一层提示，否则用户面对的是一个所有请求都失败的界面，会以为程序坏了。
const appExited = ref(false)
const exiting = ref(false)

const handleExitApp = async () => {
  if (exiting.value) return

  try {
    await ElMessageBox.confirm(
      '将停止后台服务并退出程序。关闭此页面不会退出程序，如需彻底停止请使用本功能。',
      '退出程序',
      { confirmButtonText: '退出程序', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return // 用户取消
  }

  exiting.value = true
  try {
    await systemAPI.shutdown()
    appExited.value = true
    ElMessage.success('程序已退出')
  } catch (error: any) {
    // 403 = 从局域网设备访问，后端只允许本机退出
    const status = error?.response?.status
    ElMessage.error(
      status === 403
        ? '仅允许在本机退出程序（当前通过局域网地址访问）'
        : '退出失败，请到任务管理器结束 xingmang.exe'
    )
  } finally {
    exiting.value = false
  }
}

const toggleUserPopup = () => {
  userPopupVisible.value = !userPopupVisible.value
}

const handleClickOutside = (e: MouseEvent) => {
  if (userPopupVisible.value) {
    const badge = document.querySelector('.user-badge-wrapper')
    const popup = document.querySelector('.user-popup')
    if (badge && !badge.contains(e.target as Node) && popup && !popup.contains(e.target as Node)) {
      userPopupVisible.value = false
    }
  }
  if (themeDropdownVisible.value) {
    const wrapper = themeDropdownWrapperRef.value
    if (wrapper && !wrapper.contains(e.target as Node)) {
      themeDropdownVisible.value = false
    }
  }
}

const avatarUrl = ref('')
const profileEditVisible = ref(false)
const editUserName = ref('')
const editUserBio = ref('')
const editAvatarUrl = ref('')
const avatarInputRef = ref<HTMLInputElement>()

const triggerAvatarUpload = () => {
  avatarInputRef.value?.click()
}

const handleAvatarUpload = (e: Event) => {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  if (file.size > 2 * 1024 * 1024) {
    ElMessage.warning('头像大小不能超过 2MB')
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    editAvatarUrl.value = reader.result as string
  }
  reader.readAsDataURL(file)
  input.value = ''
}

const openProfileEdit = () => {
  editUserName.value = userName.value
  editUserBio.value = userBio.value
  editAvatarUrl.value = avatarUrl.value
  userPopupVisible.value = false
  profileEditVisible.value = true
}

const cancelProfileEdit = () => {
  profileEditVisible.value = false
}

const saveProfile = () => {
  const trimmedName = editUserName.value.trim()
  if (!trimmedName) {
    ElMessage.warning('昵称不能为空')
    return
  }
  userName.value = trimmedName
  localStorage.setItem('userName', trimmedName)
  userBio.value = editUserBio.value.trim()
  localStorage.setItem('userBio', userBio.value)
  avatarUrl.value = editAvatarUrl.value
  if (avatarUrl.value) {
    localStorage.setItem('userAvatar', avatarUrl.value)
  } else {
    localStorage.removeItem('userAvatar')
  }
  profileEditVisible.value = false
  ElMessage.success('个人资料已保存')
}

const getMachineId = () => {
  const stored = localStorage.getItem('machineId')
  if (stored) {
    machineId.value = stored.substring(0, 8)
  } else {
    const id = Math.random().toString(36).substring(2, 10).toUpperCase()
    machineId.value = id
    localStorage.setItem('machineId', id)
  }
}

const syncModelCallCount = () => {
  const stored = localStorage.getItem('modelCallCount')
  if (stored) {
    modelCallCount.value = parseInt(stored, 10)
  } else {
    modelCallCount.value = 0
    localStorage.setItem('modelCallCount', '0')
  }
}

const syncPromptCount = () => {
  const stored = localStorage.getItem('promptCount')
  if (stored) {
    promptCount.value = parseInt(stored, 10)
  } else {
    promptCount.value = 0
    localStorage.setItem('promptCount', '0')
  }
}

const loadOverview = async () => {
  try {
    // 与个人中心保持同一数据源：模型调用取 overview.total_usage，提示词数取 promptAPI 实际条数
    const [ov, prompts] = await Promise.all([
      statsAPI.getOverview(),
      promptAPI.getAll(),
    ])
    if (ov.success && ov.data) {
      overview.value = ov.data
      // 后端 overview 返回 total_usage，映射到 totalUsageCount
      const usage = (ov.data as any).totalUsageCount ?? (ov.data as any).total_usage ?? 0
      modelCallCount.value = usage
      localStorage.setItem('modelCallCount', usage.toString())
    }
    if (prompts.success && Array.isArray(prompts.data)) {
      promptCount.value = prompts.data.length
      localStorage.setItem('promptCount', prompts.data.length.toString())
    }
  } catch (error) {
    console.error('加载概览失败:', error)
  }
}

const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data
    }
  } catch (error) {
    console.error('加载模型失败:', error)
  }
}

const updateLoginTime = () => {
  const now = new Date()
  loginTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
  localStorage.setItem('loginTime', loginTime.value)
}

getMachineId()
syncModelCallCount()
syncPromptCount()
loadUserProfile()
onMounted(() => {
  loadOverview()
  fetchModels()
  updateLoginTime()
  loadAppVersion()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const menuItems = [
  { index: '/books', label: '我的小说', icon: Reading },
  { index: '/prompts', label: '提示词库', icon: ChatDotSquare },
  { index: '/prompt-preview', label: '提示预览', icon: View },
  { index: '/creative', label: '变量抽卡', icon: Star },
  { index: '/ai-ranking', label: 'AI扫榜', icon: DataAnalysis },
  { index: '/agent', label: '星芒 Agent', icon: MagicStick },
  { index: '/wiki-graph', label: 'wiki图谱', icon: Share },
  { index: '/workflow', label: '工作流', icon: Connection },
  { index: '/plugins', label: '插件系统', icon: Box },
  { index: '/config', label: 'API 配置', icon: Setting },
  { index: '/profile', label: '个人中心', icon: User }
]

const globalNavItems = [
  { label: '教程', icon: Notebook },
  { label: '对话', icon: ChatLineRound },
  { label: '日志', icon: Document },
  { label: '历史', icon: Clock },
  { label: '消息', icon: Message }
]

const routeLabelMap: Record<string, string> = {
  '/books': '我的小说',
  '/prompts': '提示词库',
  '/prompt-preview': '提示预览',
  '/creative': '变量抽卡',
  '/ai-ranking': 'AI扫榜',
  '/agent': '星芒 Agent',
  '/wiki-graph': 'wiki图谱',
  '/workflow': '工作流',
  '/plugins': '插件系统',
  '/config': 'API 配置',
  '/profile': '个人中心'
}

const activeMenu = computed(() => {
  const matched = menuItems.find(item => route.path === item.index || route.path.startsWith(`${item.index}/`))
  return matched?.index || '/books'
})

const currentPageLabel = computed(() => routeLabelMap[activeMenu.value] || '创作空间')

const themeColors = {
  light: {
    'app-background': '#f4f6f8',
    'app-color': '#1a2230',
    'topbar-background': '#ffffff',
    'topbar-border': '#e6eaf0',
    'sidebar-background': '#ffffff',
    'sidebar-border': '#e6eaf0',
    'content-background': '#f4f6f8',
    'content-border': '#e6eaf0',
    'menu-text': '#5a6a80',
    'menu-active': '#2563eb',
    'menu-hover': '#eef4ff',
    'menu-hover-color': '#2563eb',
    'accent': '#2563eb',
    'accent-soft': '#eef4ff',
    'ink2': '#5a6a80',
    'ink3': '#98a5b8',
    'card': '#ffffff',
    'shadow': '0 1px 2px rgba(30, 36, 51, .04), 0 10px 30px rgba(30, 36, 51, .05)'
  },
  dark: {
    'app-background': '#0d1117',
    'app-color': '#e6edf3',
    'topbar-background': '#161b22',
    'topbar-border': '#21262d',
    'sidebar-background': '#161b22',
    'sidebar-border': '#21262d',
    'content-background': '#0d1117',
    'content-border': '#21262d',
    'menu-text': '#8b98a8',
    'menu-active': '#58a6ff',
    'menu-hover': '#12253f',
    'menu-hover-color': '#58a6ff',
    'accent': '#58a6ff',
    'accent-soft': '#12253f',
    'ink2': '#8b98a8',
    'ink3': '#5f6b7a',
    'card': '#161b22',
    'shadow': '0 1px 2px rgba(0, 0, 0, .3), 0 10px 30px rgba(0, 0, 0, .3)'
  }
}

const todayLabel = computed(() => {
  const now = new Date()
  return now.toLocaleDateString('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})

const handleGlobalNavClick = (label: string) => {
  if (label === '教程') {
    openTutorial()
  } else if (label === '对话') {
    chatDialogVisible.value = true
  } else if (label === '日志') {
    changelogVisible.value = true
  } else if (label === '历史') {
    historyDialogVisible.value = true
  } else if (label === '消息') {
    messageCenterVisible.value = true
  } else {
    ElMessage.info(`"${label}"功能开发中`)
  }
}

/* 【教程】= 新标签页打开教程中心（独立页面，两个窗口互不影响）。
   ⚠ 用 BASE_URL 而不是写死 '/'：打包版 base 变了也不会断。 */
const openTutorial = () => {
  const base = import.meta.env.BASE_URL || '/'
  const url = `${base.replace(/\/$/, '')}/tutorial`
  const win = window.open(url, '_blank', 'noopener')
  if (!win) {
    ElMessage.warning('浏览器拦截了新标签页，请允许弹出窗口后重试')
  }
}

const applyTheme = (theme: ThemeType) => {
  const colors = themeColors[theme]
  const root = document.documentElement

  if (theme === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else {
    root.removeAttribute('data-theme')
  }

  Object.entries(colors).forEach(([key, value]) => {
    root.style.setProperty(`--${key}`, value)
  })

  // 应用皮肤（覆盖强调色相关变量）
  applySkin()
  // 应用护眼模式（亮色下加暖色滤镜）
  applyEyeCare()
}

// ===== 护眼模式：亮色下给界面加暖色滤镜，降低蓝光 =====
type SkinValue = 'blue' | 'purple' | 'green' | 'orange' | 'pink'

const getInitialEyeCare = (): boolean => {
  return localStorage.getItem('main-theme-eyecare') === '1'
}

const getInitialSkin = (): SkinValue => {
  const saved = localStorage.getItem('main-theme-skin')
  if (saved === 'blue' || saved === 'purple' || saved === 'green' || saved === 'orange' || saved === 'pink') {
    return saved
  }
  return 'blue'
}

const eyeCareEnabled = ref<boolean>(getInitialEyeCare())
const skinValue = ref<SkinValue>(getInitialSkin())

// 皮肤强调色：每组含主色 accent 与柔色背景 accent-soft（亮色/暗色各一套）
const skinPalettes: Record<SkinValue, { light: { accent: string; soft: string; menuActive: string; menuHover: string; menuHoverColor: string }; dark: { accent: string; soft: string; menuActive: string; menuHover: string; menuHoverColor: string } }> = {
  blue: {
    light: { accent: '#2563eb', soft: '#eef4ff', menuActive: '#2563eb', menuHover: '#eef4ff', menuHoverColor: '#2563eb' },
    dark: { accent: '#58a6ff', soft: '#12253f', menuActive: '#58a6ff', menuHover: '#12253f', menuHoverColor: '#58a6ff' }
  },
  purple: {
    light: { accent: '#7c3aed', soft: '#f3eefe', menuActive: '#7c3aed', menuHover: '#f3eefe', menuHoverColor: '#7c3aed' },
    dark: { accent: '#a78bff', soft: '#221540', menuActive: '#a78bff', menuHover: '#221540', menuHoverColor: '#a78bff' }
  },
  green: {
    light: { accent: '#059669', soft: '#e6f7f1', menuActive: '#059669', menuHover: '#e6f7f1', menuHoverColor: '#059669' },
    dark: { accent: '#3fb950', soft: '#0f2417', menuActive: '#3fb950', menuHover: '#0f2417', menuHoverColor: '#3fb950' }
  },
  orange: {
    light: { accent: '#ea580c', soft: '#fff0e6', menuActive: '#ea580c', menuHover: '#fff0e6', menuHoverColor: '#ea580c' },
    dark: { accent: '#f0883e', soft: '#2a1a0e', menuActive: '#f0883e', menuHover: '#2a1a0e', menuHoverColor: '#f0883e' }
  },
  pink: {
    light: { accent: '#db2777', soft: '#fde8f1', menuActive: '#db2777', menuHover: '#fde8f1', menuHoverColor: '#db2777' },
    dark: { accent: '#f778ba', soft: '#2e1024', menuActive: '#f778ba', menuHover: '#2e1024', menuHoverColor: '#f778ba' }
  }
}

const skinOptions: { value: SkinValue; label: string; color: string }[] = [
  { value: 'blue', label: '星芒蓝', color: '#2563eb' },
  { value: 'purple', label: '星辉紫', color: '#7c3aed' },
  { value: 'green', label: '青芒绿', color: '#059669' },
  { value: 'orange', label: '暖橙', color: '#ea580c' },
  { value: 'pink', label: '樱花粉', color: '#db2777' }
]

// Element Plus 主色族的派生档位：[变量名, 主色占比, 混合目标]
// 占比与 EP 的 light-N / dark-2 一一对应
const EP_PRIMARY_RAMPS: [string, number, string][] = [
  ['--el-color-primary-light-3', 70, '#ffffff'],
  ['--el-color-primary-light-5', 50, '#ffffff'],
  ['--el-color-primary-light-7', 30, '#ffffff'],
  ['--el-color-primary-light-8', 20, '#ffffff'],
  ['--el-color-primary-light-9', 10, '#ffffff'],
  ['--el-color-primary-dark-2', 80, '#000000']
]

const applySkin = () => {
  const palette = skinPalettes[skinValue.value][currentThemeType.value]
  const root = document.documentElement
  root.style.setProperty('--accent', palette.accent)
  root.style.setProperty('--accent-soft', palette.soft)
  root.style.setProperty('--menu-active', palette.menuActive)
  root.style.setProperty('--menu-hover', palette.menuHover)
  root.style.setProperty('--menu-hover-color', palette.menuHoverColor)

  // ===== Element Plus 主色族一起跟随皮肤 =====
  // 全站有 10 个文件直接消费 --el-color-primary 及其派生档位（el-button / switch / radio /
  // checkbox / tabs / input 聚焦环 / pagination…），只改这一个变量的话 EP 组件会停在默认蓝
  // #409eff。EP 的派生规则（已比对 element-plus 2.11.4 dist/index.css 的实际取值）：
  //   light-N = 主色 (100-N*10)% + 白 N*10%，dark-2 = 主色 80% + 黑 20%
  // 这里把主色字面量直接写进 color-mix，避免依赖使用处的 --accent，保证 6 个档位永远同源。
  root.style.setProperty('--el-color-primary', palette.accent)
  EP_PRIMARY_RAMPS.forEach(([name, primaryPct, mixWith]) => {
    root.style.setProperty(name, `color-mix(in srgb, ${palette.accent} ${primaryPct}%, ${mixWith})`)
  })
}

const applyEyeCare = () => {
  const root = document.documentElement
  // 仅在亮色下生效；暗色下移除滤镜
  if (eyeCareEnabled.value && currentThemeType.value === 'light') {
    root.setAttribute('data-eyecare', '1')
  } else {
    root.removeAttribute('data-eyecare')
  }
}

const toggleEyeCare = (val: boolean) => {
  eyeCareEnabled.value = val
  localStorage.setItem('main-theme-eyecare', val ? '1' : '0')
  applyEyeCare()
}

const selectSkin = (skin: SkinValue) => {
  skinValue.value = skin
  localStorage.setItem('main-theme-skin', skin)
  applySkin()
  // 广播皮肤变化
  window.dispatchEvent(new CustomEvent('xingmang:skin-changed', { detail: { skin } }))
}

// ===== 跟随系统主题：监听系统深浅色变化，system 模式下自动同步 =====
let systemThemeQuery: MediaQueryList | null = null
const onSystemThemeChange = (e: MediaQueryListEvent) => {
  if (themeMode.value !== 'system') return
  const resolved: ThemeType = e.matches ? 'dark' : 'light'
  currentThemeType.value = resolved
  localStorage.setItem('main-theme', resolved)
  applyTheme(resolved)
  window.dispatchEvent(new CustomEvent('xingmang:theme-changed', { detail: { theme: resolved } }))
}

const initSystemThemeListener = () => {
  systemThemeQuery = window.matchMedia('(prefers-color-scheme: dark)')
  systemThemeQuery.addEventListener('change', onSystemThemeChange)
}

initSystemThemeListener()

const themeModeOptions: { value: ThemeMode; label: string; icon: string }[] = [
  { value: 'dark', label: '暗色', icon: 'Moon' },
  { value: 'light', label: '亮色', icon: 'Sunny' },
  { value: 'system', label: '跟随系统', icon: 'Monitor' }
]

const themeModeLabel = computed(() => {
  const opt = themeModeOptions.find(o => o.value === themeMode.value)
  return opt ? opt.label : '亮色'
})

const themeModeIcon = computed(() => {
  return themeModeOptions.find(o => o.value === themeMode.value)?.icon || 'Sunny'
})

const toggleThemeDropdown = () => {
  themeDropdownVisible.value = !themeDropdownVisible.value
}

const selectThemeMode = (mode: ThemeMode) => {
  themeMode.value = mode
  localStorage.setItem('main-theme-mode', mode)
  const resolved = resolveTheme(mode)
  currentThemeType.value = resolved
  localStorage.setItem('main-theme', resolved)
  applyTheme(resolved)
  window.dispatchEvent(new CustomEvent('xingmang:theme-changed', { detail: { theme: resolved } }))
}

// ===== 主题事件桥（供 UI 插件调用，应用保持主题唯一状态源） =====
// 插件只发事件触发切换，等价于点击顶栏主题下拉；不直接改 DOM / localStorage。
const onThemeToggle = () => {
  // 旧插件语义：在亮/暗之间循环切换；跟随系统时切到另一极
  const next: ThemeMode = themeMode.value === 'dark' ? 'light' : 'dark'
  selectThemeMode(next)
}
const onThemeSet = (e: Event) => {
  const theme = (e as CustomEvent<{ theme: string }>).detail?.theme
  if (theme === 'light' || theme === 'dark') selectThemeMode(theme)
}
window.addEventListener('xingmang:theme-toggle', onThemeToggle)
window.addEventListener('xingmang:theme-set', onThemeSet)
onUnmounted(() => {
  window.removeEventListener('xingmang:theme-toggle', onThemeToggle)
  window.removeEventListener('xingmang:theme-set', onThemeSet)
  if (systemThemeQuery) systemThemeQuery.removeEventListener('change', onSystemThemeChange)
})

// 初次挂载时按当前模式应用一次主题
applyTheme(currentThemeType.value)
</script>

<style scoped>
.app-shell {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--app-background, #f4f6f8);
  color: var(--app-color, #1a2230);
  transition: background .3s, color .3s;
}

/* ---------- 顶栏 ---------- */
.topbar {
  height: 58px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 20px;
  background: var(--topbar-background, #ffffff);
  border-bottom: 1px solid var(--topbar-border, #e6eaf0);
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-right: auto;
  min-width: 0;
}

/* 品牌标：直接用烘焙好的 32px 图标。
   不再用"渐变底 + 文字"，也不再叠加投影 —— 图标自带辉光，再加阴影会发脏。 */
.brand-mark {
  width: 32px;
  height: 32px;
  border-radius: 7px;
  flex-shrink: 0;
  display: block;
  object-fit: contain;
}

.brand-copy {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.brand-copy h1 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
  white-space: nowrap;
}

.brand-copy .chip {
  font-size: 10.5px;
  color: var(--accent, #2563eb);
  background: var(--accent-soft, #eef4ff);
  padding: 2px 8px;
  border-radius: 999px;
  white-space: nowrap;
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.topbar-action {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  height: 34px;
  padding: 0 10px;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--ink2, #5a6a80);
  font-size: 12.5px;
  cursor: pointer;
  text-decoration: none;
  transition: background .15s, color .15s;
}

.topbar-action .el-icon {
  font-size: 15px;
}

.topbar-action:hover {
  background: var(--accent-soft, #eef4ff);
  color: var(--accent, #2563eb);
}

.theme-toggle-btn {
  color: var(--accent, #2563eb);
  background: var(--accent-soft, #eef4ff);
}

.user-badge-wrapper {
  position: relative;
  margin-left: 4px;
}

.user-badge {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #ffffff;
  font-size: 14px;
  font-weight: 700;
  display: grid;
  place-items: center;
  cursor: pointer;
  transition: transform .15s;
  box-shadow: 0 4px 10px rgba(37, 99, 235, .25);
}

.user-badge:hover {
  transform: scale(1.05);
}

.badge-avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

/* ---------- 用户弹窗 ---------- */
.user-popup {
  position: fixed;
  top: 64px;
  right: 20px;
  width: 300px;
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 14px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, .14);
  opacity: 0;
  pointer-events: none;
  transform: translateY(-4px);
  transition: opacity .15s ease, transform .15s ease;
  overflow: hidden;
  z-index: 100;
}

.user-popup.visible {
  opacity: 1;
  pointer-events: auto;
  transform: none;
}

.profile-card {
  padding: 18px 20px 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  border-bottom: 1px solid var(--topbar-border, #e6eaf0);
}

.p-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #6366f1);
  color: #ffffff;
  font-size: 20px;
  font-weight: 800;
  display: grid;
  place-items: center;
  overflow: hidden;
}

.p-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.p-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
  margin-top: 10px;
}

.p-bio {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
  margin-top: 4px;
  max-width: 230px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.info-list {
  padding: 10px 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 7px 8px;
  border-radius: 8px;
  font-size: 12px;
}

.info-item label {
  color: var(--ink3, #98a5b8);
  flex-shrink: 0;
}

.info-item span {
  color: var(--app-color, #1a2230);
  font-family: 'SF Mono', 'JetBrains Mono', Consolas, monospace;
  background: var(--accent-soft, #eef4ff);
  padding: 2px 8px;
  border-radius: 6px;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-edit-trigger {
  display: block;
  width: calc(100% - 32px);
  margin: 6px 16px 16px;
  padding: 9px 0;
  border: 0;
  border-radius: 9px;
  background: var(--accent-soft, #eef4ff);
  color: var(--accent, #2563eb);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: filter .15s;
}

.profile-edit-trigger:hover {
  filter: brightness(.95);
}

/* 退出程序：危险操作，用警示色与「编辑个人资料」区分 */
.app-exit-trigger {
  width: calc(100% - 32px);
  margin: 0 16px 16px;
  padding: 9px 0;
  border: 0;
  border-radius: 9px;
  background: #fef2f2;
  color: #dc2626;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: filter .15s;
}

.app-exit-trigger:hover {
  filter: brightness(.95);
}

/* 暗色下 #fef2f2 近乎白底配红字会刺眼，改用低透明红底 + 提亮文字 */
:root[data-theme='dark'] .app-exit-trigger {
  background: rgba(220, 38, 38, .15);
  color: #f87171;
}

/* ---------- 退出后全屏提示 ---------- */
.exit-overlay {
  position: fixed;
  inset: 0;
  z-index: 4000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 23, 32, .72);
  backdrop-filter: blur(2px);
}

.exit-card {
  min-width: 320px;
  max-width: 420px;
  padding: 32px 28px;
  border-radius: 14px;
  background: var(--card, #ffffff);
  text-align: center;
  box-shadow: 0 12px 40px rgba(0, 0, 0, .28);
}

.exit-mark {
  width: 46px;
  height: 46px;
  margin: 0 auto 14px;
  border-radius: 50%;
  background: rgba(16, 185, 129, .14);
  color: #10b981;
  font-size: 24px;
  line-height: 46px;
}

.exit-card h3 {
  margin: 0 0 10px;
  font-size: 17px;
  color: var(--app-color, #1a2230);
}

.exit-card p {
  margin: 0;
  font-size: 13px;
  color: var(--ink2, #5a6a80);
  line-height: 1.7;
}

.exit-card .exit-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--ink3, #98a5b8);
}

/* ---------- 工作区 + 图标轨道侧栏 ---------- */
.workspace-shell {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: stretch;
  position: relative;
}

.sidebar {
  width: 64px;
  flex-shrink: 0;
  min-height: 0;
  background: var(--sidebar-background, #ffffff);
  border-right: 1px solid var(--sidebar-border, #e6eaf0);
  transition: width .22s ease;
  overflow: hidden;
}

.sidebar:hover {
  width: 200px;
}

.sidebar-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px 0;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  background: transparent;
}

.sidebar-menu :deep(.el-menu) {
  border-right: none;
  background: transparent;
}

.sidebar-menu :deep(.el-menu-item) {
  height: 42px;
  margin: 3px 8px;
  border-radius: 10px;
  padding: 0;
  justify-content: center;
  color: var(--menu-text, #5a6a80);
  position: relative;
  transition: background .15s, color .15s;
}

.sidebar-menu :deep(.el-menu-item .el-icon) {
  font-size: 18px;
  margin-right: 0;
  flex-shrink: 0;
}

.sidebar-menu :deep(.el-menu-item span) {
  display: none;
  white-space: nowrap;
  font-size: 13px;
  margin-left: 10px;
}

.sidebar:hover .sidebar-menu :deep(.el-menu-item) {
  justify-content: flex-start;
  padding: 0 12px;
}

.sidebar:hover .sidebar-menu :deep(.el-menu-item span) {
  display: inline;
}

.sidebar-menu :deep(.el-menu-item:hover) {
  background: var(--menu-hover, #eef4ff);
  color: var(--menu-hover-color, #2563eb);
}

.sidebar-menu :deep(.el-menu-item.is-active) {
  background: var(--menu-hover, #eef4ff);
  color: var(--menu-active, #2563eb);
  font-weight: 600;
}

.sidebar-menu :deep(.el-menu-item.is-active::after) {
  content: '';
  position: absolute;
  right: 10px;
  top: 50%;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--menu-active, #2563eb);
  transform: translateY(-50%);
}

.side-foot {
  margin: 0 10px 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: var(--accent-soft, #eef4ff);
  font-size: 11px;
  color: var(--ink2, #5a6a80);
  line-height: 1.9;
  display: none;
}

.sidebar:hover .side-foot {
  display: block;
}

.foot-row {
  display: flex;
  justify-content: space-between;
}

.foot-row b {
  color: var(--accent, #2563eb);
}

.foot-row.foot-ver {
  margin-top: 2px;
}

.foot-ver span {
  flex: none;
  margin-right: 8px;
}

/* 版本号可能是 dev(abc1234) 这类长串，超出即省略，完整信息走 title */
.foot-ver b {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 600;
}

/* ---------- 内容区 ---------- */
.content-shell {
  flex: 1;
  min-width: 0;
  min-height: 0;
  display: flex;
}

.content-frame {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 28px 34px;
  background: var(--content-background, #f4f6f8);
}

/* ---------- 主题切换下拉面板 ---------- */
.theme-dropdown-wrapper {
  position: relative;
  display: inline-flex;
}

.theme-dropdown-panel {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  width: 248px;
  padding: 4px;
  background: var(--card, #ffffff);
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 14px;
  box-shadow: 0 16px 40px rgba(15, 23, 42, .14);
  opacity: 0;
  pointer-events: none;
  transform: translateY(-4px);
  transition: opacity .15s ease, transform .15s ease;
  z-index: 200;
}

.theme-dropdown-panel.visible {
  opacity: 1;
  pointer-events: auto;
  transform: none;
}

.theme-panel-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  padding: 10px 12px 6px;
}

.theme-panel-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
}

.theme-panel-sub {
  font-size: 11px;
  color: var(--ink3, #98a5b8);
}

.theme-panel-section {
  padding: 8px 12px;
}

.theme-panel-section + .theme-panel-section {
  border-top: 1px solid var(--topbar-border, #e6eaf0);
}

.theme-section-label {
  font-size: 11.5px;
  font-weight: 600;
  color: var(--ink2, #5a6a80);
  margin-bottom: 8px;
  letter-spacing: .02em;
}

/* 主题模式三选 */
.theme-mode-row {
  display: flex;
  gap: 8px;
}

.theme-mode-chip {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: 38px;
  padding: 0 8px;
  border: 1px solid var(--topbar-border, #e6eaf0);
  border-radius: 10px;
  background: transparent;
  color: var(--ink2, #5a6a80);
  font-size: 12.5px;
  cursor: pointer;
  transition: border-color .15s, background .15s, color .15s;
}

.theme-mode-chip .el-icon {
  font-size: 15px;
}

.theme-mode-chip:hover {
  border-color: var(--accent, #2563eb);
  color: var(--accent, #2563eb);
}

.theme-mode-chip.active {
  border-color: var(--accent, #2563eb);
  background: var(--accent-soft, #eef4ff);
  color: var(--accent, #2563eb);
  font-weight: 600;
}

/* 护眼模式行 */
.theme-eyecare-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.theme-eyecare-hint {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
  line-height: 1.5;
}

/* 皮肤色板 */
.theme-skin-row {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.theme-skin-swatch {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: 2px solid transparent;
  background: var(--c, #2563eb);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0;
  position: relative;
  transition: transform .15s, box-shadow .15s;
}

.theme-skin-swatch:hover {
  transform: scale(1.08);
}

.theme-skin-swatch.active {
  box-shadow: 0 0 0 2px var(--card, #ffffff), 0 0 0 4px var(--accent, #2563eb);
}

.theme-skin-check {
  color: #ffffff;
  font-size: 14px;
  filter: drop-shadow(0 1px 1px rgba(0, 0, 0, .3));
}

/* ---------- 个人资料编辑弹窗 ---------- */
.profile-edit-dialog :deep(.el-dialog) {
  border-radius: 14px;
}

.profile-edit-form {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.form-avatar-row {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
}

.form-avatar-row > label {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-color, #1a2230);
}

.avatar-upload {
  position: relative;
  width: 84px;
  height: 84px;
  border-radius: 50%;
  cursor: pointer;
  overflow: hidden;
  border: 2px dashed var(--accent, #2563eb);
  background: var(--accent-soft, #eef4ff);
  transition: border-color .2s;
}

.avatar-upload:hover {
  border-color: var(--accent, #2563eb);
}

.avatar-preview {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--ink3, #98a5b8);
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, .4);
  color: #ffffff;
  opacity: 0;
  transition: opacity .2s;
}

.avatar-upload:hover .avatar-overlay {
  opacity: 1;
}

.avatar-file-input {
  display: none;
}

.avatar-hint {
  font-size: 12px;
  color: var(--ink3, #98a5b8);
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-item > label {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-color, #1a2230);
}

/* ---------- 响应式 ---------- */
@media (max-width: 1200px) {
  .topbar-action span {
    display: none;
  }

  .topbar-action {
    padding: 0 9px;
  }
}

@media (max-width: 768px) {
  .topbar {
    padding: 0 12px;
  }

  .brand-copy .chip {
    display: none;
  }

  .sidebar {
    display: none;
  }

  .content-frame {
    padding: 16px;
  }
}
</style>
