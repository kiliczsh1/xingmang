<template>
  <div class="app-shell">
    <canvas ref="particleCanvas" class="bg-particles"></canvas>
    
    <!-- 折叠状态的顶部栏 -->
    <div class="topbar-collapsed-bar" v-if="!topbarExpanded">
      <button class="topbar-expand-btn" @click="topbarExpanded = true" title="展开导航栏">
        <el-icon><Expand /></el-icon>
      </button>
    </div>

    <!-- 展开状态的顶部栏 -->
    <header class="topbar topbar-expanded" v-else>
      <div class="brand-block">
        <button class="topbar-collapse-btn" @click="topbarExpanded = false">
          <el-icon><Fold /></el-icon>
        </button>
        <div
          class="user-badge-wrapper"
          @click.stop="toggleUserPopup"
        >
          <div class="user-badge">
            <img v-if="avatarUrl" :src="avatarUrl" class="badge-avatar" />
            <span v-else>我</span>
          </div>
          <div
            class="user-popup"
            :class="{ visible: userPopupVisible }"
          >
            <div class="profile-container">
              <div class="user-info">
                <div class="info-item">
                  <label><el-icon class="label-icon"><Monitor /></el-icon>用户 ID:</label>
                  <span>{{ machineId }}</span>
                </div>
                <div class="info-item">
                  <label><el-icon class="label-icon"><User /></el-icon>用户名称:</label>
                  <div class="name-edit-container">
                    <span>{{ userName }}</span>
                  </div>
                </div>
                <div class="info-item" v-if="userBio">
                  <label><el-icon class="label-icon"><EditPen /></el-icon>个性签名:</label>
                  <span class="bio-preview">{{ userBio }}</span>
                </div>
                <button class="profile-edit-trigger" @click.stop="openProfileEdit">
                  <el-icon><Edit /></el-icon>
                  <span>编辑个人资料</span>
                </button>
                <div class="info-item">
                  <label><el-icon class="label-icon"><Timer /></el-icon>登录时间:</label>
                  <span>{{ loginTime }}</span>
                </div>
                <div class="info-item">
                  <label><el-icon class="label-icon"><DataAnalysis /></el-icon>总调用次数:</label>
                  <span>{{ overview?.totalUsageCount || modelCallCount }}</span>
                </div>
                <div class="info-item">
                  <label><el-icon class="label-icon"><List /></el-icon>提示词数量:</label>
                  <span>{{ overview?.promptCount || promptCount }}</span>
                </div>
                <div class="info-item">
                  <label><el-icon class="label-icon"><Collection /></el-icon>模型数量:</label>
                  <span>{{ models.length }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="brand-copy">
          <h1>星芒 AI 写作</h1>
          <p></p>
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
        <button
          class="topbar-action theme-toggle-btn"
          type="button"
          @click="toggleTheme"
          :title="currentThemeType === 'dark' ? '切换为亮色主题' : '切换为暗色主题'"
        >
          <el-icon><component :is="currentThemeType === 'dark' ? 'Sunny' : 'Moon'" /></el-icon>
          <span>{{ currentThemeType === 'dark' ? '亮色' : '暗色' }}</span>
        </button>
        <button
          v-for="item in globalNavItems.filter(i => i.label !== '外观')"
          :key="item.label"
          class="topbar-action"
          type="button"
          @click="handleGlobalNavClick(item.label)"
        >
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.label }}</span>
        </button>
      </nav>
    </header>

    <!-- 外观主题选择弹窗 -->
    <el-dialog
      v-model="themeDialogVisible"
      title="选择主题"
      width="400px"
      :close-on-click-modal="true"
      class="theme-select-dialog"
      append-to-body
    >
      <div class="theme-options">
        <div
          class="theme-option"
          :class="{ active: currentThemeType === 'light' }"
          @click="selectTheme('light')"
        >
          <div class="theme-preview theme-preview-light">
            <div class="preview-sidebar"></div>
            <div class="preview-content"></div>
          </div>
          <div class="theme-label">
            <span class="theme-name">白色主界面</span>
            <el-icon v-if="currentThemeType === 'light'" class="check-icon"><Check /></el-icon>
          </div>
        </div>
        <div
          class="theme-option"
          :class="{ active: currentThemeType === 'dark' }"
          @click="selectTheme('dark')"
        >
          <div class="theme-preview theme-preview-dark">
            <div class="preview-sidebar"></div>
            <div class="preview-content"></div>
          </div>
          <div class="theme-label">
            <span class="theme-name">暗色主界面</span>
            <el-icon v-if="currentThemeType === 'dark'" class="check-icon"><Check /></el-icon>
          </div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="themeDialogVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>

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

    <!-- AI 对话弹窗 -->
    <ChatDialog v-model:visible="chatDialogVisible" />

    <!-- 消息中心 -->
    <MessageCenter v-model:visible="messageCenterVisible" />

    <div class="workspace-shell">
      <aside class="sidebar" :class="currentTheme">
        <div class="sidebar-panel">
          <el-menu
            :default-active="activeMenu"
            router
            class="sidebar-menu"
            :background-color="'transparent'"
            :text-color="menuTextColor"
            :active-text-color="menuActiveColor"
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
        </div>
      </aside>

      <main class="content-shell">
        <section class="content-frame">
          <router-view />
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Reading,
  ChatDotSquare,
  View,
  Star,
  Connection,
  Collection,
  Setting,
  User,
  Brush,
  ChatLineRound,
  School,
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
  Expand,
  Fold,
  ChatLineSquare,
  Camera
} from '@element-plus/icons-vue'
import HistoryCenter from '../components/HistoryCenter.vue'
import ChatDialog from '../components/ChatDialog.vue'
import MessageCenter from '../components/MessageCenter.vue'
import { statsAPI, providerAPI, configAPI } from '@/api'
import type { UsageOverview, ApiModel } from '@/types'

const route = useRoute()

type ThemeType = 'light' | 'dark'

const getInitialTheme = (): ThemeType => {
  const saved = localStorage.getItem('main-theme')
  if (saved === 'light' || saved === 'dark') {
    return saved
  }
  return 'light'
}

const themeDialogVisible = ref(false)
const currentThemeType = ref<ThemeType>(getInitialTheme())
const historyDialogVisible = ref(false)
const chatDialogVisible = ref(false)
const messageCenterVisible = ref(false)
const userPopupVisible = ref(false)
const topbarExpanded = ref(true)
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

const toggleUserPopup = () => {
  userPopupVisible.value = !userPopupVisible.value
}

const handleClickOutside = (e: MouseEvent) => {
  if (!userPopupVisible.value) return
  const badge = document.querySelector('.user-badge-wrapper')
  const popup = document.querySelector('.user-popup')
  if (badge && !badge.contains(e.target as Node) && popup && !popup.contains(e.target as Node)) {
    userPopupVisible.value = false
  }
}

const particleCanvas = ref<HTMLCanvasElement>()
let particleCtx: CanvasRenderingContext2D | null = null
let particleAnimId = 0
let particleArray: { x: number; y: number; vx: number; vy: number; r: number; o: number }[] = []
const PARTICLE_COUNT = 70
const PARTICLE_MAX_DIST = 120

const initParticles = () => {
  const canvas = particleCanvas.value
  if (!canvas) return

  particleCtx = canvas.getContext('2d')
  if (!particleCtx) return

  const resize = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  }
  resize()
  window.addEventListener('resize', resize)

  const isDark = currentThemeType.value === 'dark'
  const dotColor = isDark ? 'rgba(45, 212, 191, 0.5)' : 'rgba(13, 148, 136, 0.28)'
  const lineColorWeak = isDark ? 'rgba(45, 212, 191, 0.04)' : 'rgba(13, 148, 136, 0.03)'
  const lineColorStrong = isDark ? 'rgba(45, 212, 191, 0.12)' : 'rgba(13, 148, 136, 0.08)'

  particleArray = []
  for (let i = 0; i < PARTICLE_COUNT; i++) {
    particleArray.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      vx: (Math.random() - 0.5) * 0.35,
      vy: (Math.random() - 0.5) * 0.35,
      r: 1 + Math.random() * 2,
      o: 0.3 + Math.random() * 0.5
    })
  }

  const draw = () => {
    const ctx = particleCtx
    if (!ctx || !canvas) return
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    for (let i = 0; i < particleArray.length; i++) {
      const p = particleArray[i]
      p.x += p.vx
      p.y += p.vy

      if (p.x < -20) p.x = canvas.width + 20
      if (p.x > canvas.width + 20) p.x = -20
      if (p.y < -20) p.y = canvas.height + 20
      if (p.y > canvas.height + 20) p.y = -20

      ctx.beginPath()
      ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
      ctx.fillStyle = dotColor
      ctx.globalAlpha = p.o
      ctx.fill()

      for (let j = i + 1; j < particleArray.length; j++) {
        const q = particleArray[j]
        const dx = p.x - q.x
        const dy = p.y - q.y
        const dist = Math.sqrt(dx * dx + dy * dy)
        if (dist < PARTICLE_MAX_DIST) {
          const t = 1 - dist / PARTICLE_MAX_DIST
          ctx.beginPath()
          ctx.moveTo(p.x, p.y)
          ctx.lineTo(q.x, q.y)
          ctx.strokeStyle = t > 0.5 ? lineColorStrong : lineColorWeak
          ctx.globalAlpha = t * 0.35
          ctx.lineWidth = 0.4 + t * 0.6
          ctx.stroke()
        }
      }
    }
    ctx.globalAlpha = 1
    particleAnimId = requestAnimationFrame(draw)
  }
  draw()

  watch(currentThemeType, (theme) => {
    cancelAnimationFrame(particleAnimId)
    if (particleCtx) {
      particleCtx.clearRect(0, 0, canvas.width, canvas.height)
    }
    nextTick(() => initParticles())
  })
}

const destroyParticles = () => {
  cancelAnimationFrame(particleAnimId)
  particleArray = []
  particleCtx = null
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
    const res = await statsAPI.getOverview()
    if (res.success && res.data) {
      overview.value = res.data
      promptCount.value = res.data.promptCount || 0
      modelCallCount.value = res.data.totalUsageCount || 0
      localStorage.setItem('promptCount', res.data.promptCount.toString())
      localStorage.setItem('modelCallCount', res.data.totalUsageCount.toString())
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
  nextTick(() => initParticles())
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  destroyParticles()
  document.removeEventListener('click', handleClickOutside)
})

const menuItems = [
  { index: '/books', label: '我的小说', icon: Reading },
  { index: '/prompts', label: '提示词库', icon: ChatDotSquare },
  { index: '/prompt-preview', label: '提示预览', icon: View },
  { index: '/creative', label: '变量抽卡', icon: Star },
  { index: '/workflow', label: '工作流', icon: Connection },
  { index: '/experience-shares', label: '经验分享', icon: Collection },
  { index: '/config', label: 'API 配置', icon: Setting },
  { index: '/profile', label: '个人中心', icon: User }
]

const globalNavItems = [
  { label: '外观', icon: Brush },
  { label: '对话', icon: ChatLineRound },
  { label: '论坛', icon: ChatLineSquare },
  { label: '教程', icon: School },
  { label: '历史', icon: Clock },
  { label: '消息', icon: Message }
]

const routeLabelMap: Record<string, string> = {
  '/books': '我的小说',
  '/prompts': '提示词库',
  '/prompt-preview': '提示预览',
  '/creative': '变量抽卡',
  '/workflow': '工作流',
  '/experience-shares': '经验分享',
  '/config': 'API 配置',
  '/profile': '个人中心'
}

const activeMenu = computed(() => {
  const matched = menuItems.find(item => route.path === item.index || route.path.startsWith(`${item.index}/`))
  return matched?.index || '/books'
})

const currentPageLabel = computed(() => routeLabelMap[activeMenu.value] || '创作空间')

const currentTheme = ref<ThemeType>(getInitialTheme())

const themeColors = {
  light: {
    appBackground: '#f0fdf9',
    appColor: '#1e293b',
    topbarBackground: 'rgba(255, 255, 255, 0.78)',
    topbarBorder: 'rgba(45, 212, 191, 0.18)',
    sidebarBackground: 'rgba(255, 255, 255, 0.72)',
    sidebarBorder: 'rgba(45, 212, 191, 0.12)',
    contentBackground: '#ffffff',
    contentBorder: 'rgba(45, 212, 191, 0.1)',
    menuText: '#5a6c7d',
    menuActive: '#ffffff',
    menuHover: 'rgba(45, 212, 191, 0.08)',
    menuHoverColor: '#2dd4bf'
  },
  dark: {
    appBackground: '#060b14',
    appColor: '#f4f7ff',
    topbarBackground: 'rgba(10, 22, 40, 0.72)',
    topbarBorder: 'rgba(45, 212, 191, 0.18)',
    sidebarBackground: 'rgba(8, 18, 36, 0.65)',
    sidebarBorder: 'rgba(45, 212, 191, 0.12)',
    contentBackground: 'rgba(15, 29, 51, 0.55)',
    contentBorder: 'rgba(45, 212, 191, 0.08)',
    menuText: '#94a3b8',
    menuActive: '#ffffff',
    menuHover: 'rgba(45, 212, 191, 0.08)',
    menuHoverColor: '#2dd4bf'
  }
}

const currentThemeColors = computed(() => themeColors[currentThemeType.value])

const menuTextColor = computed(() => themeColors[currentTheme.value].menuText)
const menuActiveColor = computed(() => themeColors[currentTheme.value].menuActive)

const todayLabel = computed(() => {
  const now = new Date()
  return now.toLocaleDateString('zh-CN', {
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})

const handleGlobalNavClick = (label: string) => {
  if (label === '外观') {
    themeDialogVisible.value = true
  } else if (label === '对话') {
    chatDialogVisible.value = true
  } else if (label === '论坛') {
    window.open('https://kook.vip/sLySCy', '_blank', 'noopener,noreferrer')
  } else if (label === '历史') {
    historyDialogVisible.value = true
  } else if (label === '消息') {
    messageCenterVisible.value = true
  } else {
    ElMessage.info(`"${label}"功能开发中`)
  }
}

const toggleTheme = () => {
  const newTheme = currentThemeType.value === 'dark' ? 'light' : 'dark'
  selectTheme(newTheme)
}

const selectTheme = (theme: ThemeType) => {
  currentThemeType.value = theme
  localStorage.setItem('main-theme', theme)
  applyTheme(theme)
}

const applyTheme = (theme: ThemeType) => {
  const colors = themeColors[theme]
  const root = document.documentElement
  
  // 设置 data-theme 属性用于 CSS 变量覆盖
  if (theme === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else {
    root.removeAttribute('data-theme')
  }
  
  // 设置内联 CSS 变量
  root.style.setProperty('--app-background', colors.appBackground)
  root.style.setProperty('--app-color', colors.appColor)
  root.style.setProperty('--topbar-background', colors.topbarBackground)
  root.style.setProperty('--topbar-border', colors.topbarBorder)
  root.style.setProperty('--sidebar-background', colors.sidebarBackground)
  root.style.setProperty('--sidebar-border', colors.sidebarBorder)
  root.style.setProperty('--content-background', colors.contentBackground)
  root.style.setProperty('--content-border', colors.contentBorder)
  root.style.setProperty('--menu-text', colors.menuText)
  root.style.setProperty('--menu-hover', colors.menuHover)
  root.style.setProperty('--menu-hover-color', colors.menuHoverColor)
}

watch(
  currentThemeType,
  newTheme => {
    applyTheme(newTheme)
  },
  { immediate: true }
)

watch(
  () => route.name,
  (routeName) => {
    if (routeName === 'BookAnalysis' || routeName === 'CharacterLibrary') {
      topbarExpanded.value = false
    }
  },
  { immediate: true }
)
</script>

<style scoped>
.app-shell {
  height: 100%;
  display: flex;
  flex-direction: column;
  background:
    linear-gradient(rgba(45, 212, 191, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(45, 212, 191, 0.025) 1px, transparent 1px),
    radial-gradient(ellipse at 15% 10%, rgba(34, 211, 238, 0.07) 0%, transparent 55%),
    radial-gradient(ellipse at 85% 85%, rgba(16, 185, 129, 0.06) 0%, transparent 55%),
    radial-gradient(ellipse at 50% 50%, rgba(45, 212, 191, 0.03) 0%, transparent 70%),
    var(--app-background, #060b14);
  background-size:
    48px 48px,
    48px 48px,
    100% 100%,
    100% 100%,
    100% 100%,
    100% 100%;
  color: var(--app-color, #2c3e50);
  position: relative;
}

:root:not([data-theme='dark']) .app-shell {
  background:
    radial-gradient(ellipse at 80% 0%, rgba(45, 212, 191, 0.06) 0%, transparent 50%),
    radial-gradient(ellipse at 20% 100%, rgba(34, 211, 238, 0.04) 0%, transparent 50%),
    var(--app-background, #f0fdf9);
}

.topbar {
  height: 64px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 0 16px 0 20px;
  border-bottom: 1px solid var(--topbar-border, rgba(45, 212, 191, 0.18));
  background: var(--topbar-background, rgba(255, 255, 255, 0.78));
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 10;
}

.topbar-collapsed-bar {
  position: fixed;
  top: 12px;
  left: 12px;
  z-index: 100;
}

.topbar-expand-btn {
  width: 40px;
  height: 40px;
  border: 1px solid rgba(45, 212, 191, 0.25);
  border-radius: 12px;
  background: var(--topbar-background, rgba(255, 255, 255, 0.85));
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  color: var(--menu-text, #5a6c7d);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s ease;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.topbar-expand-btn:hover {
  background: rgba(45, 212, 191, 0.15);
  color: #2dd4bf;
  border-color: rgba(45, 212, 191, 0.5);
  box-shadow: 0 6px 20px rgba(45, 212, 191, 0.2);
  transform: translateY(-2px);
}

.topbar-expand-btn .el-icon {
  font-size: 18px;
}

.topbar-expanded .brand-block {
  display: flex;
  align-items: center;
  gap: 14px;
}

.topbar-collapse-btn {
  width: 36px;
  height: 36px;
  border: 1px solid rgba(45, 212, 191, 0.2);
  border-radius: 10px;
  background: transparent;
  color: var(--menu-text, #5a6c7d);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.25s ease;
}

.topbar-collapse-btn:hover {
  background: rgba(45, 212, 191, 0.1);
  color: #2dd4bf;
  border-color: rgba(45, 212, 191, 0.4);
}

.topbar::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(45, 212, 191, 0.15) 15%,
    var(--menu-hover-color, #2dd4bf) 30%,
    rgba(34, 211, 238, 0.7) 50%,
    var(--menu-hover-color, #2dd4bf) 70%,
    rgba(45, 212, 191, 0.15) 85%,
    transparent 100%
  );
  box-shadow:
    0 0 10px rgba(45, 212, 191, 0.3),
    0 0 24px rgba(45, 212, 191, 0.12);
  animation: topbarGlow 4s ease-in-out infinite;
}

@keyframes topbarGlow {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 1; }
}

:root:not([data-theme='dark']) .topbar::after {
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(45, 212, 191, 0.1) 15%,
    #2dd4bf 30%,
    #22d3ee 50%,
    #2dd4bf 70%,
    rgba(45, 212, 191, 0.1) 85%,
    transparent 100%
  );
  box-shadow:
    0 0 8px rgba(45, 212, 191, 0.2),
    0 0 18px rgba(45, 212, 191, 0.08);
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.brand-mark {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
  color: #ffffff;
  font-size: 18px;
  font-weight: 700;
  box-shadow: 0 10px 24px rgba(45, 212, 191, 0.35);
}

.brand-copy h1 {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: 0.06em;
  background: linear-gradient(
    135deg,
    #22d3ee 0%,
    #2dd4bf 25%,
    #6ee7b7 55%,
    #2dd4bf 80%,
    #22d3ee 100%
  );
  background-size: 200% 200%;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 0 8px rgba(45, 212, 191, 0.45));
  animation: brandShimmer 4s ease-in-out infinite;
}

@keyframes brandShimmer {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

:root:not([data-theme='dark']) .brand-copy h1 {
  background: linear-gradient(135deg, #0d9488, #14b8a6, #0d9488);
  background-size: 200% 200%;
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  filter: drop-shadow(0 0 4px rgba(20, 184, 166, 0.25));
}

.topbar-center {
  display: flex;
  justify-content: center;
}

.notice-chip {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 40px;
  padding: 0 16px;
  border-radius: 999px;
  color: #b8ffd7;
  background: rgba(71, 214, 152, 0.1);
  border: 1px solid rgba(71, 214, 152, 0.22);
}

.notice-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #47d698;
  box-shadow: 0 0 12px rgba(71, 214, 152, 0.8);
}

.topbar-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.topbar-action {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 10px;
  background: transparent;
  color: #5a6c7d;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.topbar-action:hover {
  color: #2dd4bf;
  background: rgba(45, 212, 191, 0.06);
  border-color: rgba(45, 212, 191, 0.35);
  box-shadow:
    0 0 12px rgba(45, 212, 191, 0.15),
    0 0 24px rgba(45, 212, 191, 0.06),
    inset 0 0 12px rgba(45, 212, 191, 0.04);
  transform: translateY(-1px);
  animation: navGlowPulse 2s ease-in-out infinite;
}

@keyframes navGlowPulse {
  0%, 100% {
    box-shadow:
      0 0 12px rgba(45, 212, 191, 0.15),
      0 0 24px rgba(45, 212, 191, 0.06),
      inset 0 0 12px rgba(45, 212, 191, 0.04);
  }
  50% {
    box-shadow:
      0 0 20px rgba(34, 211, 238, 0.25),
      0 0 36px rgba(45, 212, 191, 0.12),
      inset 0 0 16px rgba(45, 212, 191, 0.08);
  }
}

:root:not([data-theme='dark']) .topbar-action {
  border-color: rgba(45, 212, 191, 0.08);
}

:root:not([data-theme='dark']) .topbar-action:hover {
  background: rgba(45, 212, 191, 0.06);
  border-color: rgba(45, 212, 191, 0.3);
  box-shadow:
    0 0 12px rgba(20, 184, 166, 0.12),
    inset 0 0 12px rgba(20, 184, 166, 0.03);
}

.theme-toggle-btn {
  color: #5a6c7d;
  border-color: rgba(45, 212, 191, 0.15);
  background: rgba(45, 212, 191, 0.05);
}

.theme-toggle-btn:hover {
  color: #2dd4bf;
  background: rgba(45, 212, 191, 0.12);
  border-color: rgba(45, 212, 191, 0.3);
}

.user-badge-wrapper {
  position: relative;
  display: inline-block;
  z-index: 101;
}

.user-badge {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-left: 4px;
  background: linear-gradient(135deg, #2dd4bf 0%, #14b8a6 50%, #0d9488 100%);
  background-size: 200% 200%;
  color: #ffffff;
  font-weight: 700;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  animation: plasmaPulse 2.5s ease-in-out infinite;
  position: relative;
  overflow: hidden;
  box-shadow:
    0 0 18px rgba(45, 212, 191, 0.45),
    0 0 36px rgba(45, 212, 191, 0.18),
    0 0 64px rgba(45, 212, 191, 0.08),
    inset 0 0 10px rgba(255, 255, 255, 0.12);
}

.badge-avatar {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  object-fit: cover;
}

@keyframes plasmaPulse {
  0%, 100% {
    box-shadow:
      0 0 18px rgba(45, 212, 191, 0.45),
      0 0 36px rgba(45, 212, 191, 0.18),
      0 0 64px rgba(45, 212, 191, 0.08),
      inset 0 0 10px rgba(255, 255, 255, 0.12);
  }
  50% {
    box-shadow:
      0 0 28px rgba(34, 211, 238, 0.55),
      0 0 52px rgba(45, 212, 191, 0.28),
      0 0 90px rgba(45, 212, 191, 0.14),
      inset 0 0 16px rgba(255, 255, 255, 0.2);
  }
}

.user-badge::before {
  content: '';
  position: absolute;
  top: -60%;
  left: -60%;
  width: 220%;
  height: 220%;
  background: radial-gradient(ellipse at 30% 20%, rgba(255, 255, 255, 0.35) 0%, transparent 55%);
  opacity: 0;
  transition: opacity 0.35s ease;
}

.user-badge:hover::before {
  opacity: 1;
}

.user-badge-wrapper:hover .user-badge {
  animation: none;
  transform: scale(1.06);
  box-shadow:
    0 0 28px rgba(34, 211, 238, 0.55),
    0 0 56px rgba(45, 212, 191, 0.28),
    0 0 100px rgba(45, 212, 191, 0.14),
    inset 0 0 16px rgba(255, 255, 255, 0.22);
}

.user-badge-wrapper:hover::after {
  inset: -6px;
  border-width: 2px;
  border-top-color: rgba(45, 212, 191, 0.75);
  border-right-color: rgba(34, 211, 238, 0.5);
  animation-duration: 1.8s;
}

.user-badge:active {
  transform: scale(0.95) !important;
}

@keyframes gradientFlow {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.user-popup {
  position: fixed;
  top: 72px;
  left: 24px;
  width: 300px;
  background: rgba(10, 22, 40, 0.94);
  border: 1px solid rgba(45, 212, 191, 0.15);
  border-radius: 16px;
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.5),
    0 0 0 1px rgba(45, 212, 191, 0.08) inset;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0s 0s;
  z-index: 9999;
  overflow: hidden;
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
}

:root:not([data-theme='dark']) .user-popup {
  background: rgba(255, 255, 255, 0.94);
  border-color: rgba(13, 148, 136, 0.15);
  box-shadow:
    0 20px 50px rgba(0, 0, 0, 0.12),
    0 0 0 1px rgba(255, 255, 255, 0.6) inset;
}

.user-popup.visible {
  opacity: 1;
  pointer-events: auto;
  transition: opacity 0.15s ease;
}

.user-popup::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #22d3ee, #2dd4bf, #6ee7b7, #2dd4bf, #22d3ee);
  background-size: 200% 100%;
  filter: drop-shadow(0 0 6px rgba(45, 212, 191, 0.5));
  animation: gradientShift 3s ease infinite;
}

:root:not([data-theme='dark']) .user-popup::before {
  background: linear-gradient(90deg, #0d9488, #14b8a6, #2dd4bf, #14b8a6, #0d9488);
}

@keyframes gradientShift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

.user-badge-wrapper::after {
  content: '';
  position: absolute;
  inset: -5px;
  border-radius: 50%;
  border: 1.5px solid transparent;
  border-top-color: rgba(45, 212, 191, 0.55);
  border-right-color: rgba(34, 211, 238, 0.28);
  animation: orbitalSpin 3.5s linear infinite;
  pointer-events: none;
}

:root:not([data-theme='dark']) .user-badge-wrapper::after {
  border-top-color: rgba(13, 148, 136, 0.45);
  border-right-color: rgba(20, 184, 166, 0.2);
}

@keyframes orbitalSpin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.user-popup .profile-container {
  padding: 0;
}

.user-popup .user-info {
  display: flex;
  flex-direction: column;
  padding: 14px 18px 16px;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
  margin-bottom: 0;
}

.user-popup .info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 10px;
  margin-bottom: 2px;
  border-radius: 8px;
  transition: all 0.2s ease;
  border-bottom: none;
  background: transparent;
}

.user-popup .info-item:last-child {
  margin-bottom: 0;
  padding-bottom: 8px;
  border-bottom: none;
}

.user-popup .info-item:hover {
  background: rgba(45, 212, 191, 0.06);
  transform: none;
}

.user-popup .info-item label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  font-weight: 600;
  color: #94a3b8;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

:root:not([data-theme='dark']) .user-popup .info-item label {
  color: #64748b;
}

.user-popup .info-item label .label-icon {
  font-size: 13px;
  color: #2dd4bf;
}

.user-popup .info-item:hover .label-icon {
  color: #22d3ee;
  transform: none;
}

.user-popup .info-item label::after {
  content: '';
  flex: 1;
}

.user-popup .info-item span {
  font-size: 12px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.88);
  font-family: 'SF Mono', 'JetBrains Mono', monospace;
  padding: 3px 8px;
  background: rgba(45, 212, 191, 0.07);
  border-radius: 5px;
  border: 1px solid rgba(45, 212, 191, 0.1);
}

:root:not([data-theme='dark']) .user-popup .info-item span {
  color: #1e293b;
  background: rgba(13, 148, 136, 0.06);
  border-color: rgba(13, 148, 136, 0.12);
}

.user-popup .bio-preview {
  max-width: 180px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-edit-trigger {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: calc(100% - 36px);
  margin: 10px 18px 14px;
  padding: 10px 0;
  border: 1px solid rgba(45, 212, 191, 0.2);
  border-radius: 10px;
  background: rgba(45, 212, 191, 0.06);
  color: #2dd4bf;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.profile-edit-trigger:hover {
  background: rgba(45, 212, 191, 0.12);
  border-color: rgba(45, 212, 191, 0.35);
  box-shadow: 0 0 12px rgba(45, 212, 191, 0.1);
}

:root:not([data-theme='dark']) .profile-edit-trigger {
  background: rgba(13, 148, 136, 0.05);
  border-color: rgba(13, 148, 136, 0.15);
  color: #0d9488;
}

:root:not([data-theme='dark']) .profile-edit-trigger:hover {
  background: rgba(13, 148, 136, 0.1);
  border-color: rgba(13, 148, 136, 0.3);
}

.user-popup .name-edit-container {
  display: flex;
  align-items: center;
  gap: 6px;
}

@keyframes slideIn {
  from { opacity: 0; transform: translateX(-6px); }
  to { opacity: 1; transform: translateX(0); }
}

.workspace-shell {
  min-height: 0;
  flex: 1;
  display: grid;
  grid-template-columns: 165px minmax(0, 1fr);
  position: relative;
  z-index: 1;
}

.sidebar {
  padding: 0;
  border-right: 1px solid var(--sidebar-border, rgba(45, 212, 191, 0.12));
  background: var(--sidebar-background, rgba(255, 255, 255, 0.72));
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
}

.sidebar-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.sidebar-hero {
  padding: 18px 16px;
  border-radius: 20px;
  background:
    radial-gradient(circle at top right, rgba(255, 255, 255, 0.6), transparent 24%),
    linear-gradient(135deg, #2dd4bf 0%, #14b8a6 100%);
  box-shadow: 0 18px 38px rgba(20, 184, 166, 0.25);
}

.sidebar-hero-date {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
}

.sidebar-hero-title {
  margin-top: 10px;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.1;
  color: #ffffff;
}

.sidebar-hero-subtitle {
  margin-top: 8px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
}

.sidebar-menu {
  flex: 1;
  border-right: none;
}

.content-shell {
  min-width: 0;
  min-height: 0;
  padding: 0;
  overflow: hidden;
}

.content-frame {
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  border-radius: 0;
  padding: 20px;
  background:
    radial-gradient(ellipse at 100% 0%, rgba(34, 211, 238, 0.03) 0%, transparent 40%),
    radial-gradient(ellipse at 0% 100%, rgba(16, 185, 129, 0.02) 0%, transparent 40%),
    var(--content-background, transparent);
  border: none;
  box-shadow: none;
}

:root[data-theme='dark'] .content-frame {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
}

:deep(.el-menu) {
  border-right: none;
}

:deep(.el-menu-item) {
  height: 48px;
  margin: 6px 8px;
  border-radius: 12px;
  color: var(--menu-text, #5a6c7d);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  position: relative;
}

:deep(.el-menu-item .el-icon) {
  font-size: 18px;
}

:deep(.el-menu-item::after) {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent 0%,
    rgba(45, 212, 191, 0.06) 50%,
    transparent 100%
  );
  transition: left 0.5s ease;
  pointer-events: none;
}

:deep(.el-menu-item:hover) {
  background: var(--menu-hover, rgba(45, 212, 191, 0.06)) !important;
  color: var(--menu-hover-color, #2dd4bf) !important;
}

:deep(.el-menu-item:hover::after) {
  left: 100%;
}

:deep(.el-menu-item.is-active) {
  background: linear-gradient(90deg, rgba(45, 212, 191, 0.14) 0%, rgba(45, 212, 191, 0.04) 100%) !important;
  color: #2dd4bf !important;
  box-shadow: 0 4px 12px rgba(20, 184, 166, 0.15);
}

:deep(.el-menu-item.is-active::before) {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #2dd4bf;
  border-radius: 0 2px 2px 0;
  box-shadow: 0 0 10px rgba(45, 212, 191, 0.5);
}

.sidebar.classic {
  background: var(--sidebar-background, #ffffff);
}

.sidebar.tech {
  background: var(--sidebar-background, #ffffff);
}

/* 主题选择弹窗样式 */
.theme-select-dialog .el-dialog {
  border-radius: 16px;
}

.theme-select-dialog .el-dialog__header {
  padding: 20px 24px;
  border-bottom: 1px solid #e0e6ed;
}

.theme-select-dialog .el-dialog__title {
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
}

.theme-select-dialog .el-dialog__body {
  padding: 24px;
}

.theme-options {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.theme-option {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px;
  border: 2px solid #e0e6ed;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.theme-option:hover {
  border-color: #2dd4bf;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.15);
}

.theme-option.active {
  border-color: #2dd4bf;
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.25);
}

.theme-preview {
  width: 100%;
  height: 120px;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  border: 1px solid #e0e6ed;
}

.theme-preview-light {
  background: #f5f7fa;
}

.theme-preview-light .preview-sidebar {
  width: 30%;
  background: #ffffff;
  border-right: 1px solid #e0e6ed;
}

.theme-preview-light .preview-content {
  width: 70%;
  background: #ffffff;
}

.theme-preview-dark {
  background: #0d1424;
}

.theme-preview-dark .preview-sidebar {
  width: 30%;
  background: rgba(9, 14, 26, 0.94);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
}

.theme-preview-dark .preview-content {
  width: 70%;
  background: rgba(17, 26, 46, 0.96);
}

.theme-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.theme-name {
  font-size: 15px;
  font-weight: 600;
  color: #2c3e50;
}

.check-icon {
  color: #2dd4bf;
  font-size: 20px;
}

.theme-select-dialog .el-dialog__footer {
  padding: 16px 24px;
  border-top: 1px solid #e0e6ed;
}

/* --- Profile Edit Dialog --- */

.profile-edit-dialog :deep(.el-dialog) {
  border-radius: 16px;
}

.profile-edit-dialog :deep(.el-dialog__header) {
  padding: 20px 24px 0;
  border-bottom: none;
}

.profile-edit-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 600;
}

.profile-edit-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
}

.profile-edit-dialog :deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid #e0e6ed;
}

.profile-edit-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-avatar-row {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.form-avatar-row > label {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-color, #1e293b);
}

.avatar-upload {
  position: relative;
  width: 88px;
  height: 88px;
  border-radius: 50%;
  cursor: pointer;
  overflow: hidden;
  border: 2px dashed rgba(45, 212, 191, 0.3);
  transition: all 0.2s ease;
}

.avatar-upload:hover {
  border-color: #2dd4bf;
  box-shadow: 0 0 20px rgba(45, 212, 191, 0.18);
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
  background: rgba(45, 212, 191, 0.06);
  color: #94a3b8;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.avatar-upload:hover .avatar-overlay {
  opacity: 1;
}

.avatar-file-input {
  display: none;
}

.avatar-hint {
  font-size: 12px;
  color: var(--ds-text-tertiary, rgba(0, 0, 0, 0.36));
}

.form-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-item > label {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-color, #1e293b);
}

:root[data-theme='dark'] .profile-edit-dialog :deep(.el-dialog__footer) {
  border-top-color: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .profile-edit-dialog :deep(.el-input__wrapper) {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.12);
}

:root[data-theme='dark'] .profile-edit-dialog :deep(.el-textarea__inner) {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.12);
  color: #f4f7ff;
}

:root[data-theme='dark'] .profile-edit-dialog :deep(.el-textarea__inner::placeholder) {
  color: #6b7280;
}

@media (max-width: 1200px) {
  .topbar-center {
    display: none;
  }

  .topbar-actions span {
    display: none;
  }

  .topbar-action {
    width: 40px;
    padding: 0;
    justify-content: center;
  }
}

@media (max-width: 900px) {
  .workspace-shell {
    grid-template-columns: 150px minmax(0, 1fr);
  }

  .content-shell {
    padding-left: 14px;
  }

  .content-frame {
    padding: 18px;
  }
}

@media (max-width: 768px) {
  .topbar {
    padding: 0 14px 0 16px;
  }

  .workspace-shell {
    grid-template-columns: 1fr;
  }

  .sidebar {
    display: none;
  }

  .content-shell {
    padding: 12px;
  }

  .content-frame {
    border-radius: 18px;
    padding: 16px;
  }
}

.bg-particles {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}
</style>
