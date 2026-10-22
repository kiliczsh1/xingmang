<template>
  <div class="tutorial-shell">
    <!-- 顶栏：与主站顶栏同一视觉语言（高度、边框、头像区）。
         ⚠ 不放「返回主站」：教程中心本来就是新标签页，用户自己切标签就行。 -->
    <header class="tb-top">
      <div class="tb-brand">
        <span class="tb-mark">教</span>
        <div class="tb-copy">
          <h1>教程中心</h1>
          <span class="tb-sub">独立页面 · 与原窗口互不影响</span>
        </div>
      </div>

      <div class="tb-right">
        <button class="tb-theme" type="button" :title="`当前主题：${themeLabel}`" @click="toggleTheme">
          <el-icon><component :is="isDark ? Moon : Sunny" /></el-icon>
        </button>
        <div class="tb-user" :title="userName">
          <img v-if="avatarUrl" :src="avatarUrl" alt="" />
          <span v-else>我</span>
        </div>
      </div>
    </header>

    <div class="tb-body">
      <!-- 贴吧式左栏：顶部两项入口 → 分组标题 → 吧列表 -->
      <aside class="tb-side" :class="{ collapsed }">
        <nav class="tb-nav">
          <router-link :to="BOARD_HOME_PATH" class="tb-navitem" :class="{ on: isHomeActive }">
            <i class="tb-navicon"><el-icon><House /></el-icon></i>
            <span class="tb-label">首页</span>
          </router-link>
          <router-link :to="TUTORIAL_QUICKREF_PATH" class="tb-navitem" :class="{ on: isQuickRefActive }">
            <i class="tb-navicon"><el-icon><Reading /></el-icon></i>
            <span class="tb-label">功能速查表</span>
          </router-link>

          <div class="tb-group-title tb-label">全部吧</div>

          <router-link
            v-for="b in BOARD_ORDER"
            :key="b"
            :to="tutorialBoardPath(b)"
            class="tb-baritem"
            :class="{ on: isBoardActive(b) }"
            :title="BOARD_CONF[b].desc"
          >
            <i class="tb-badge" :class="`badge-${b}`">{{ BOARD_CONF[b].short }}</i>
            <span class="tb-label">{{ BOARD_CONF[b].name }}</span>
          </router-link>
        </nav>
      </aside>

      <!-- 折叠按钮：压在左右分界线上，收/展都点它（对应贴吧那枚圆形按钮） -->
      <button
        class="tb-fold"
        :class="{ collapsed }"
        type="button"
        :title="collapsed ? '展开导航栏' : '折叠导航栏'"
        @click="toggleCollapsed"
      >
        <el-icon><component :is="collapsed ? Expand : Fold" /></el-icon>
      </button>

      <main class="tb-main">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { Expand, Fold, House, Moon, Reading, Sunny } from '@element-plus/icons-vue'
import {
  BOARD_CONF,
  BOARD_HOME_PATH,
  BOARD_ORDER,
  TUTORIAL_QUICKREF_PATH,
  tutorialBoardPath
} from '@/constants/boards'
import type { BoardKind } from '@/types'

const route = useRoute()

/* ---------- 主题 ----------
   ⚠ 教程中心是独立路由 + 独立布局，MainLayout.applyTheme() 不参与渲染，
   --accent / data-theme 都没人挂。这里自己做三件事：
     1. CSS 里给每个变量写 var(--x, 兜底)（见 style，不能自引用）；
     2. onMounted 按 localStorage['main-theme-mode'] 补挂 data-theme="dark"；
     3. 顶栏给切换按钮，改同一份 localStorage，跟主站保持一致。 */
const themeMode = ref<'light' | 'dark'>('light')
const isDark = computed(() => themeMode.value === 'dark')
const themeLabel = computed(() => (isDark.value ? '暗色' : '亮色'))

const resolveInitialTheme = (): 'light' | 'dark' => {
  const saved = localStorage.getItem('main-theme-mode')
  const legacy = localStorage.getItem('main-theme')
  const mode = saved === 'light' || saved === 'dark' || saved === 'system'
    ? saved
    : (legacy === 'dark' ? 'dark' : 'light')
  if (mode === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  }
  return mode
}

const applyTheme = (mode: 'light' | 'dark') => {
  const root = document.documentElement
  if (mode === 'dark') {
    root.setAttribute('data-theme', 'dark')
  } else {
    root.removeAttribute('data-theme')
  }
}

const toggleTheme = () => {
  themeMode.value = isDark.value ? 'light' : 'dark'
  localStorage.setItem('main-theme-mode', themeMode.value)
  localStorage.removeItem('main-theme')
  applyTheme(themeMode.value)
}

const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
const handleSystemThemeChange = () => {
  if (localStorage.getItem('main-theme-mode') !== 'system') return
  themeMode.value = mediaQuery.matches ? 'dark' : 'light'
  applyTheme(themeMode.value)
}

/* ---------- 左栏 ---------- */
const COLLAPSE_KEY = 'tutorial-sidebar-collapsed'
const collapsed = ref(localStorage.getItem(COLLAPSE_KEY) === '1')

const toggleCollapsed = () => {
  collapsed.value = !collapsed.value
  localStorage.setItem(COLLAPSE_KEY, collapsed.value ? '1' : '0')
}

const currentBoard = computed<BoardKind | null>(() => {
  if (!route.path.startsWith('/tutorial/board/')) return null
  const value = route.params.board
  const raw = Array.isArray(value) ? value[0] : value
  return raw === 'tutorial' ? 'tutorial' : 'experience'
})

const isBoardActive = (board: BoardKind) => currentBoard.value === board
const isHomeActive = computed(() => route.path === BOARD_HOME_PATH)
const isQuickRefActive = computed(() => route.path.startsWith(TUTORIAL_QUICKREF_PATH))

/* ---------- 用户信息（与主站读同一份 localStorage） ---------- */
const userName = ref('星芒用户')
const avatarUrl = ref('')

const loadUserProfile = () => {
  try {
    const raw = localStorage.getItem('userProfile')
    if (!raw) return
    const p = JSON.parse(raw)
    if (p?.name) userName.value = String(p.name)
    if (p?.avatar) avatarUrl.value = String(p.avatar)
  } catch {
    /* 坏了就用默认值，不影响页面 */
  }
}

/* 主站改了主题/资料，本标签页跟着更新（同源 localStorage 不触发 storage 事件时靠 focus 兜底） */
const handleFocus = () => {
  const saved = localStorage.getItem('main-theme-mode')
  if ((saved === 'light' || saved === 'dark') && saved !== themeMode.value) {
    themeMode.value = saved
    applyTheme(saved)
  }
  loadUserProfile()
}

onMounted(() => {
  const initial = resolveInitialTheme()
  themeMode.value = initial
  applyTheme(initial)
  loadUserProfile()
  mediaQuery.addEventListener('change', handleSystemThemeChange)
  window.addEventListener('focus', handleFocus)
})

onUnmounted(() => {
  mediaQuery.removeEventListener('change', handleSystemThemeChange)
  window.removeEventListener('focus', handleFocus)
})
</script>

<style scoped>
/* ---------- 变量兜底 ----------
   ⚠ 全部写 var(--x, 兜底)：本布局可能在没有 MainLayout 的情况下直接渲染。
   暗色值统一在 :global([data-theme='dark']) 那段覆盖，不写死在本块里。 */
.tutorial-shell {
  --tb-accent: var(--accent, #2563eb);
  --tb-accent-soft: var(--accent-soft, #eef4ff);
  --tb-bg: var(--content-background, #f4f6f8);
  --tb-card: var(--card, #ffffff);
  --tb-border: var(--content-border, #e6eaf0);
  --tb-text: var(--app-color, #1a2230);
  --tb-text-2: var(--ink2, #5a6a80);
  --tb-text-3: var(--ink3, #98a5b8);
  --tb-hover: #f2f3f5;
  --tb-top-h: 56px;
  --tb-side-w: 220px;
  --tb-side-w-collapsed: 64px;

  /* ⚠ 全局 style.css 里 html/body/#app 是 height:100% + overflow:hidden，
     所以**整页永远不会滚**。滚动必须由内容区自己承担，否则速查表下半截看不到。 */
  height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: var(--tb-bg);
  color: var(--tb-text);
}

/* ---------- 顶栏 ---------- */
.tb-top {
  height: var(--tb-top-h);
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 18px;
  background: var(--tb-card);
  border-bottom: 1px solid var(--tb-border);
  position: relative;
  z-index: 20;
}

.tb-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.tb-mark {
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  background: var(--tb-accent);
  color: #fff;
  font-size: 15px;
  font-weight: 600;
}

.tb-copy {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.tb-copy h1 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}

.tb-sub {
  font-size: 12px;
  color: var(--tb-text-3);
  white-space: nowrap;
}

.tb-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.tb-theme {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  border-radius: 8px;
  border: 1px solid var(--tb-border);
  background: transparent;
  color: var(--tb-text-2);
  cursor: pointer;
  transition: background .15s, color .15s, border-color .15s;
}

.tb-theme:hover {
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
  border-color: var(--tb-accent);
}

.tb-user {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  overflow: hidden;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--tb-accent-soft);
  color: var(--tb-accent);
  font-size: 13px;
  flex-shrink: 0;
}

.tb-user img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* ---------- 主体 ---------- */
.tb-body {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: stretch;
  position: relative;
}

/* ---------- 左栏 ---------- */
.tb-side {
  width: var(--tb-side-w);
  flex-shrink: 0;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 10px 10px 16px;
  background: var(--tb-card);
  border-right: 1px solid var(--tb-border);
  transition: width .2s ease;
}

.tb-side.collapsed {
  width: var(--tb-side-w-collapsed);
  padding-left: 8px;
  padding-right: 8px;
}

.tb-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

/* 顶部入口（首页 / 功能速查表）：圆形线性图标 + 常规字重，像贴吧那样 */
.tb-navitem {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 44px;
  padding: 0 10px;
  border-radius: 8px;
  color: var(--tb-text);
  font-size: 15px;
  text-decoration: none;
  white-space: nowrap;
  transition: background .15s, color .15s;
}

.tb-navitem:hover {
  background: var(--tb-hover);
}

.tb-navicon {
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-style: normal;
  font-size: 21px;
  color: var(--tb-text-2);
}

.tb-navitem.on .tb-navicon,
.tb-navitem.on .tb-label {
  color: var(--tb-accent);
}

.tb-navitem.on .tb-label {
  font-weight: 600;
}

/* 分组标题：贴吧用一行灰色小字，不加竖条 */
.tb-group-title {
  margin: 14px 0 6px;
  padding-left: 10px;
  font-size: 13px;
  color: var(--tb-text-3);
  line-height: 20px;
}

/* 吧条目：圆角方块头像 + 吧名 */
.tb-baritem {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 52px;
  padding: 0 10px;
  border-radius: 8px;
  color: var(--tb-text);
  font-size: 15px;
  text-decoration: none;
  white-space: nowrap;
  transition: background .15s;
}

.tb-baritem:hover {
  background: var(--tb-hover);
}

.tb-baritem.on {
  background: var(--tb-hover);
}

.tb-baritem.on .tb-label {
  color: var(--tb-accent);
  font-weight: 600;
}

.tb-badge {
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 10px;
  font-size: 15px;
  font-style: normal;
  font-weight: 600;
  color: #fff;
}

.badge-tutorial {
  background: linear-gradient(135deg, #2563eb, #0ea5e9);
}

.badge-experience {
  background: linear-gradient(135deg, #14b8a6, #0d9488);
}

.tb-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 收起态：只留图标并居中，文字隐藏 */
.tb-side.collapsed .tb-label {
  display: none;
}

.tb-side.collapsed .tb-navitem,
.tb-side.collapsed .tb-baritem {
  justify-content: center;
  padding: 0;
  gap: 0;
}

.tb-side.collapsed .tb-group-title {
  display: none;
}

/* ---------- 折叠按钮：压在分界线上 ---------- */
.tb-fold {
  position: absolute;
  top: 50%;
  left: calc(var(--tb-side-w) - 12px);
  transform: translateY(-50%);
  width: 24px;
  height: 24px;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  border: 1px solid var(--tb-border);
  background: var(--tb-card);
  color: var(--tb-text-3);
  font-size: 12px;
  cursor: pointer;
  z-index: 5;
  box-shadow: 0 1px 4px rgba(16, 24, 40, .12);
  transition: left .2s ease, color .15s, border-color .15s;
}

.tb-fold.collapsed {
  left: calc(var(--tb-side-w-collapsed) - 12px);
}

.tb-fold:hover {
  color: var(--tb-accent);
  border-color: var(--tb-accent);
}

/* ---------- 内容区（滚动容器） ---------- */
.tb-main {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
}

@media (max-width: 1100px) {
  .tutorial-shell {
    --tb-side-w: 180px;
  }
}
</style>

<style>
/* ---------- 暗色覆盖 ----------
   ⚠ 必须放**非 scoped** 全局块：data-theme 挂在 <html> 上，
   scoped 选择器编译出的 [data-v] 属性在 html 上不成立，写 scoped 会静默失效。 */
[data-theme='dark'] .tutorial-shell {
  --tb-bg: #0f141a;
  --tb-card: #161b22;
  --tb-border: #21262d;
  --tb-text: #e6edf3;
  --tb-text-2: #9aa7b6;
  --tb-text-3: #6e7a89;
  --tb-hover: #1f2937;
}

[data-theme='dark'] .tb-fold {
  box-shadow: 0 1px 4px rgba(0, 0, 0, .5);
}
</style>
