/**
 * theme-toggle · 全局右下角暗亮切换按钮（顶栏主题开关的快捷入口）
 * ----------------------------------------------------------
 * 本插件不重新实现主题逻辑，只做一件事：把右下角按钮变成
 * 顶栏「外观」主题开关的快捷入口 —— 点击时向应用派发
 * `xingmang:theme-toggle` 事件，由应用（MainLayout.vue 主题事件桥）
 * 执行自身的 toggleTheme()（写 localStorage['main-theme']、
 * 增删 <html> 的 data-theme='dark' 属性、设置 :root CSS 变量）。
 *
 * 应用切换主题后会广播 `xingmang:theme-changed`，插件据此同步按钮图标。
 * 主题状态源始终是应用：localStorage['main-theme'] + <html> data-theme 属性。
 *
 * 注册契约：window.__installPlugin（UI 插件，process 为聊天链 no-op）
 * 入口以 new Function('window', code)(window) 本机执行。
 */
(function (win) {
  'use strict'

  // 同步注册插件对象（加载器执行完入口代码后立即读取）
  win.__installPlugin = {
    name: 'theme-toggle',
    description: '全局右下角圆形按钮，一键切换暗色 / 亮色主题（顶栏主题开关的快捷入口）',
    enabled: true,
    // UI 类插件不参与聊天内容处理链，原样返回
    process: function (content, context) {
      return content
    }
  }

  var BTN_ID = 'xingmang-theme-toggle'
  var CSS_ID = 'xingmang-theme-css'
  var LEGACY_KEY = 'xingmang-theme' // 旧版本遗留，清理用

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

  // 清理旧版本遗留的独立主题 key（主题状态源已统一为应用的 main-theme）
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
      // 快捷入口：触发应用自身的主题切换（等价于点击顶栏外观开关）
      win.dispatchEvent(new CustomEvent('xingmang:theme-toggle'))
    })
    document.body.appendChild(btn)
    refreshIcon()
  }

  // 应用主题变化（顶栏开关 / 本按钮 / 其他入口）后同步按钮图标
  win.addEventListener('xingmang:theme-changed', function () {
    refreshIcon()
  })

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
