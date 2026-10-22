import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import router from './router'
import './style.css'
import './styles/dark-theme.css'

const app = createApp(App)

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(createPinia())
app.use(router)
app.use(ElementPlus)

app.mount('#app')

/**
 * 移除首屏骨架（index.html 里的 #xm-boot）。
 *
 * ⚠ 关键点：不能只等 router.isReady() —— 路由导航失败（例如 chunk 404）
 * 时它同样会 resolve，此时撤掉骨架、内容又没渲染出来，等于把白屏换了个地方发生。
 * 因此这里以「真实界面确实渲染出来了」为准：轮询等待 #app 出现可见内容，
 * 拿到才撤骨架；始终没等到就保留骨架（宁可持续显示加载态，也不给用户一片空白）。
 */
function dismissBootSkeleton() {
  const bootEl = document.getElementById('xm-boot')
  if (!bootEl || !bootEl.parentNode) return

  bootEl.classList.add('xm-boot-out')
  let removed = false
  const done = () => {
    if (removed) return
    removed = true
    bootEl.remove()
  }
  bootEl.addEventListener('transitionend', done, { once: true })
  // transitionend 在元素被隐藏 / 降级动画时可能不触发，兜底清掉
  setTimeout(done, 500)
}

/** 真实界面是否已渲染出可交互内容（排除只有空壳、或仅有插件浮窗的情况）。 */
function isAppRendered(): boolean {
  const app = document.getElementById('app')
  if (!app) return false
  // 任一种布局外壳挂出来即可：MainLayout / TutorialLayout / FullScreenLayout。
  // ⚠ 只认 .app-shell 会漏掉不带它的布局（教程中心就踩过），表现为骨架永远不撤、整页卡在加载态。
  if (app.querySelector('.app-shell, .tutorial-shell, .fullscreen-layout')) return true
  // 兜底：新增布局忘了用上述类名时，只要已有可见内容也照样撤骨架
  const first = app.firstElementChild as HTMLElement | null
  return !!first && first.offsetHeight > 40 && (app.textContent || '').trim().length > 0
}

/** 轮询等待界面渲染完成；超时返回 false（不撤骨架）。 */
function waitForAppRendered(timeoutMs = 8000): Promise<boolean> {
  return new Promise(resolve => {
    const deadline = Date.now() + timeoutMs
    const tick = () => {
      if (isAppRendered()) {
        // 双 rAF：确保这一帧真的绘制到屏幕，再撤骨架，避免中间露出空白
        requestAnimationFrame(() => requestAnimationFrame(() => resolve(true)))
        return
      }
      if (Date.now() > deadline) {
        resolve(false)
        return
      }
      setTimeout(tick, 60)
    }
    tick()
  })
}

router
  .isReady()
  .then(() => waitForAppRendered())
  .then(rendered => {
    if (rendered) dismissBootSkeleton()
    // 未渲染成功则保留骨架：用户看到的是「正在加载」，而不是白屏。
    // 真正的失败会由 Vite 的 vite:preloadError 自愈逻辑重载，或由错误面板呈现。
  })
  .catch(() => {
    // 路由初始化异常同样不撤骨架，保留加载态并交给控制台暴露问题
  })
