/**
 * theme-toggle · 全局右下角暗亮切换按钮
 * ----------------------------------------------------------
 * 适配星芒应用（Vue3 + Element Plus）的真实主题机制，而非通用 .dark class 约定：
 *   1) 主题状态源  = localStorage['main-theme']（'light' | 'dark'）
 *   2) 样式开关    = <html> 上的 data-theme='dark' 属性（:root[data-theme='dark'] 选择器）
 *   3) 颜色变量    = 在 :root 内联设置的 18 个 CSS 变量（--app-background/--topbar-*/
 *                    --sidebar-*/--content-*/--menu-*/--accent/--accent-soft/--ink2/
 *                    --ink3/--card/--shadow），调色板与 MainLayout.vue themeColors 一致
 * 本插件完整复刻 applyTheme 的三步，点按钮直接驱动 DOM + localStorage；
 * 应用下次 applyTheme 会写回相同值，天然幂等。
 *
 * 注册契约：window.__installPlugin（UI 插件，process 为聊天链 no-op）
 * 入口以 new Function('window', code)(window) 本机执行。
 */
(function (win) {
  'use strict'

  // 同步注册插件对象（加载器执行完入口代码后立即读取）
  win.__installPlugin = {
    name: 'theme-toggle',
    description: '全局右下角圆形按钮，一键切换暗色 / 亮色主题',
    enabled: true,
    // UI 类插件不参与聊天内容处理链，原样返回
    process: function (content, context) {
      return content
    }
  }

  var KEY = 'main-theme'            // 与应用主题状态源一致
  var LEGACY_KEY = 'xingmang-theme' // 旧版本遗留，清理用
  var BTN_ID = 'xingmang-theme-toggle'
  var CSS_ID = 'xingmang-theme-css'

  // 调色板：逐项对应 MainLayout.vue themeColors（light / dark），勿遗漏 / 改名
  var LIGHT = {
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
  }
  var DARK = {
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

  // 注入按钮自身样式：显式配色，两种主题下都清晰可见；暗色变体挂 data-theme 选择器
  var style = document.getElementById(CSS_ID) || document.createElement('style')
  style.id = CSS_ID
  style.textContent =
    '#' + BTN_ID + '{' +
      'position:fixed;right:20px;bottom:20px;width:48px;height:48px;border-radius:50%;' +
      'border:1px solid rgba(127,139,158,.3);background:rgba(31,41,55,.92);color:#fff;' +
      'box-shadow:0 6px 16px rgba(0,0,0,.22);cursor:pointer;z-index:2147483000;' +
      'display:flex;align-items:center;justify-content:center;' +
      'font-size:22px;line-height:1;user-select:none;-webkit-user-select:none;' +
      'transition:background .3s,color .3s,transform .2s;' +
    '}' +
    '#' + BTN_ID + ':hover{transform:scale(1.08)}' +
    'html[data-theme="dark"] #' + BTN_ID + '{' +
      'background:rgba(255,255,255,.94);color:#111827;border-color:rgba(0,0,0,.14);' +
      'box-shadow:0 6px 16px rgba(0,0,0,.5);' +
    '}'
  if (!style.isConnected) document.head.appendChild(style)

  // 清理旧版本遗留的独立主题 key（与应用 main-theme 统一）
  try { win.localStorage.removeItem(LEGACY_KEY) } catch (e) { /* 隐私模式忽略 */ }

  var btn = null

  function isDark() {
    return document.documentElement.hasAttribute('data-theme')
  }

  function refreshIcon() {
    if (!btn) return
    var dark = isDark()
    btn.textContent = dark ? '☀️' : '🌙'
    btn.title = dark ? '切换到亮色' : '切换到暗色'
  }

  /**
   * 核心：复刻 MainLayout.applyTheme ——
   *   data-theme 属性 + :root 内联 CSS 变量 + colorScheme + localStorage['main-theme']
   * 与应用自身主题机制完全一致，下次应用 applyTheme 写回相同值，无冲突。
   */
  function apply(theme, save) {
    var root = document.documentElement
    var palette = theme === 'dark' ? DARK : LIGHT
    if (theme === 'dark') root.setAttribute('data-theme', 'dark')
    else root.removeAttribute('data-theme')
    Object.keys(palette).forEach(function (k) {
      root.style.setProperty('--' + k, palette[k])
    })
    root.style.colorScheme = theme
    if (save !== false) {
      try { win.localStorage.setItem(KEY, theme) } catch (e) { /* 忽略 */ }
    }
    refreshIcon()
  }

  function mount() {
    // 旧版本按钮可能已存在（带着旧的点击逻辑）：移除重建，确保绑定的是新逻辑
    var old = document.getElementById(BTN_ID)
    if (old) old.remove()
    btn = document.createElement('button')
    btn.id = BTN_ID
    btn.type = 'button'
    btn.setAttribute('aria-label', '切换暗色 / 亮色')
    btn.addEventListener('click', function (e) {
      e.stopPropagation()
      apply(isDark() ? 'light' : 'dark')
    })
    document.body.appendChild(btn)
    refreshIcon()
  }

  function init() {
    // 初始化只读当前已生效主题来同步图标，不主动写主题（应用启动时已 applyTheme）
    if (document.body) { mount() }
    else {
      document.addEventListener('DOMContentLoaded', function () {
        mount()
      })
    }
  }
  init()
})(window)
