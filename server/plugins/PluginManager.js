/**
 * 后端插件管理器 V3
 *
 * 核心变化：
 * - 插件存放在用户数据目录 data/user-plugins/，与源码分离
 * - 用户可以自由增删插件目录，删掉即卸载
 * - 系统零插件时也能正常运行
 *
 * 原理：
 * - 插件调用 .on('事件名', 回调) 来注册钩子
 * - 你的代码调用 .trigger('事件名', 数据) 来触发钩子
 */

const fs = require('fs')
const path = require('path')

const c = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  cyan: '\x1b[36m',
  yellow: '\x1b[33m',
  gray: '\x1b[90m',
}

class PluginManager {
  constructor() {
    /** @type {Map<string, Array<{ pluginName: string, callback: Function }>>} */
    this.hooks = new Map()
    /** @type {Map<string, object>} */
    this.plugins = new Map()
    /** @type {string} 用户插件目录 */
    this.pluginsRoot = ''
  }

  /**
   * 设置用户插件目录并扫描加载
   * @param {string} rootDir - 应用根目录（便携版=exe所在目录，开发版=项目根目录）
   */
  init(rootDir) {
    this.pluginsRoot = path.join(rootDir, 'data', 'user-plugins')

    // 确保目录存在
    fs.mkdirSync(this.pluginsRoot, { recursive: true })

    // 首次启动：如果目录为空，创建演示插件
    const entries = fs.readdirSync(this.pluginsRoot).filter(e => !e.startsWith('.'))
    if (entries.length === 0) {
      this._createDemoPlugin()
    }

    // 扫描并加载
    this._scanAndLoad()
  }

  /**
   * 首次启动时创建一个演示插件，让用户能看到效果
   * 用户可以随时在插件中心卸载它
   */
  _createDemoPlugin() {
    const demoDir = path.join(this.pluginsRoot, 'skin-example')
    fs.mkdirSync(demoDir, { recursive: true })

    // plugin.json
    fs.writeFileSync(path.join(demoDir, 'plugin.json'), JSON.stringify({
      name: 'skin-example',
      version: '1.0.0',
      description: '示例皮肤插件 - 为作品管理页添加自定义皮肤切换功能',
      author: { name: 'XingNovel Demo' },
      hooks: {
        'ui.books.mounted': 'injectSkinButton',
        'ui.theme.change': 'onThemeChange'
      },
      permissions: ['ui.style', 'ui.books.read'],
      interface: {
        displayName: '皮肤切换示例',
        shortDescription: '在「我的小说」页注入换肤按钮，提供 3 种卡片皮肤（默认/暖色/赛博朋克）',
        category: 'UI',
        capabilities: ['Theme', 'Interactive'],
        brandColor: '#f59e0b'
      }
    }, null, 2))

    // index.js
    fs.writeFileSync(path.join(demoDir, 'index.js'), [
      'module.exports = {',
      '  install(manager) {',
      "    manager.on('book.created', (data) => {",
      "      console.log('🎨 [skin-example] 📖 新作品:', data.book?.title)",
      "    }, 'skin-example')",
      "    manager.on('book.deleted', (data) => {",
      "      console.log('🎨 [skin-example] 🗑️ 作品已删除:', data.id)",
      "    }, 'skin-example')",
      "    console.log('🎨 [skin-example] 后端已就绪')",
      '  }',
      '}'
    ].join('\n'))

    // frontend.js（从之前的实现复制）
    fs.writeFileSync(path.join(demoDir, 'frontend.js'), [
      '(function(plugin, manifest){',
      "  'use strict'",
      '  var PLUGIN=manifest.name',
      "  var skin=localStorage.getItem('xingnovel_skin')||'default'",
      '  var btn=null',
      '  var SKINS={',
      "    default:{name:'默认风格',icon:'🎨',css:''},",
      "    warm:{name:'暖色皮肤',icon:'🌅',css:",
      "      '.books-container .book-card{background:linear-gradient(135deg,#fffbeb 0%,#fef3c7 100%)!important;border-color:#fcd34d!important;border-radius:18px!important}',",
      "      +'.books-container .book-card:hover{box-shadow:0 8px 30px rgba(245,158,11,0.2)!important;border-color:#f59e0b!important}',",
      "      +'.books-container .book-title{color:#92400e!important}',",
      "      +'.books-container .header h2{color:#d97706!important}',",
      "      +'.books-container .cover-placeholder{background:linear-gradient(160deg,#f59e0b 0%,#d97706 100%)!important}',",
      "      +'.books-container .progress-fill{background:linear-gradient(90deg,#f59e0b 0%,#fbbf24 100%)!important}',",
      "    },",
      "    cyberpunk:{name:'赛博朋克',icon:'🤖',css:",
      "      '.books-container .book-card{background:linear-gradient(135deg,#0a0a1a 0%,#1a0a2e 100%)!important;border:1px solid #00ff88!important;box-shadow:0 0 15px rgba(0,255,136,0.2)!important}',",
      "      +'.books-container .book-card:hover{box-shadow:0 0 30px rgba(0,255,136,0.4),0 0 60px rgba(255,0,255,0.2)!important;border-color:#ff00ff!important}',",
      "      +'.books-container .book-title{color:#00ff88!important;text-shadow:0 0 10px rgba(0,255,136,0.5)!important}',",
      "      +'.books-container .header h2{color:#ff00ff!important;text-shadow:0 0 15px rgba(255,0,255,0.6)!important}',",
      "      +'.books-container .cover-placeholder{background:linear-gradient(160deg,#ff00ff 0%,#00ff88 100%)!important}',",
      "      +'.books-container .progress-fill{background:linear-gradient(90deg,#ff00ff 0%,#00ff88 100%)!important}',",
      "      +'.books-container .act-btn{background:rgba(0,255,136,0.1)!important;border-color:rgba(0,255,136,0.3)!important;color:#00ff88!important}',",
      "      +'.books-container .act-btn:hover{background:#00ff88!important;color:#000!important}',",
      "    }",
      '  }',
      '  function applySkin(name){',
      '    var s=SKINS[name];if(!s)return',
      "    plugin.removeStyle(PLUGIN,'skin-css')",
      "    if(s.css)plugin.injectStyle(PLUGIN,s.css,'skin-css')",
      "    skin=name;localStorage.setItem('xingnovel_skin',name);updateBtn()",
      '  }',
      '  function updateBtn(){',
      '    if(!btn)return;var names=[\'default\',\'warm\',\'cyberpunk\']',
      '    var cur=names.indexOf(skin);var n=SKINS[names[(cur+1)%3]]',
      "    btn.innerHTML=n.icon+' '+n.name",
      '  }',
      '  function injectBtn(data){',
      "    if(document.getElementById('plugin-skin-btn'))return",
      "    var el=data.container;if(!el)return setTimeout(function(){var c=document.querySelector('.books-container');if(c)injectBtn({container:c})},500)",
      "    var ha=el.querySelector('.header-actions');if(!ha)return",
      "    btn=document.createElement('button');btn.id='plugin-skin-btn'",
      "    updateBtn();btn.title='插件提供：点击切换作品卡片皮肤'",
      '    btn.onclick=function(){var names=[\'default\',\'warm\',\'cyberpunk\'];var cur=names.indexOf(skin);applySkin(names[(cur+1)%3])}',
      "    btn.style.cssText='display:inline-flex;align-items:center;gap:6px;height:32px;padding:0 14px;border:2px solid #f59e0b;border-radius:8px;background:linear-gradient(135deg,rgba(245,158,11,0.1),rgba(251,191,36,0.05));color:#d97706;font-size:13px;font-weight:600;cursor:pointer;transition:all 0.3s ease;white-space:nowrap'",
      '    btn.onmouseenter=function(){btn.style.background=\'linear-gradient(135deg,#f59e0b,#fbbf24)\';btn.style.color=\'#fff\';btn.style.boxShadow=\'0 4px 15px rgba(245,158,11,0.4)\'}',
      "    btn.onmouseleave=function(){btn.style.background='linear-gradient(135deg,rgba(245,158,11,0.1),rgba(251,191,36,0.05))';btn.style.color='#d97706';btn.style.boxShadow='none'}",
      "    ha.insertBefore(btn,ha.firstChild);if(plugin.registerDomElement)plugin.registerDomElement(PLUGIN,btn)",
      "    if(skin!=='default')applySkin(skin)",
      '  }',
      "  plugin.on('ui.books.mounted',injectBtn,PLUGIN)",
      "  console.log('🎨 ['+PLUGIN+'] 前端已就绪')",
      '})'
    ].join('\n'))

    console.log(`${c.cyan}📦 [PluginManager]${c.reset} 首次启动，已创建演示插件: skin-example`)
    console.log(`${c.gray}   💡 可在「插件中心」中卸载${c.reset}`)
  }

  /**
   * 插件调用这个来注册钩子
   */
  on(hookName, callback, pluginName = 'unknown') {
    if (!this.hooks.has(hookName)) {
      this.hooks.set(hookName, [])
    }
    this.hooks.get(hookName).push({ pluginName, callback })
    console.log(`${c.cyan}🔌 [PluginManager]${c.reset} 钩子已注册: ${c.green}${hookName}${c.reset} ← ${c.gray}${pluginName}${c.reset}`)
  }

  /**
   * 触发钩子 —— 异步，错误不影响主流程
   */
  async trigger(hookName, data) {
    const listeners = this.hooks.get(hookName) || []
    if (listeners.length === 0) return

    console.log(`${c.yellow}⚡ [PluginManager]${c.reset} 触发: ${c.green}${hookName}${c.reset} (${listeners.length} 个监听者)`)

    for (const { pluginName, callback } of listeners) {
      try {
        await callback(data)
      } catch (err) {
        console.error(`${c.yellow}⚠️ [PluginManager]${c.reset} "${pluginName}" ${hookName} 失败: ${err.message}`)
      }
    }
  }

  /**
   * 加载单个插件目录
   */
  loadPlugin(pluginDir) {
    const dirName = path.basename(pluginDir)
    const manifestPath = path.join(pluginDir, 'plugin.json')
    if (!fs.existsSync(manifestPath)) {
      console.warn(`${c.yellow}⚠️ [PluginManager]${c.reset} 跳过 ${dirName}: 没有 plugin.json`)
      return false
    }

    const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
    const mainPath = path.join(pluginDir, 'index.js')

    if (!fs.existsSync(mainPath)) {
      console.warn(`${c.yellow}⚠️ [PluginManager]${c.reset} "${manifest.name}" 没有 index.js，跳过`)
      return false
    }

    try {
      // 清除 require 缓存，以便插件更新后重新加载
      delete require.cache[require.resolve(mainPath)]
      const pluginMain = require(mainPath)
      if (typeof pluginMain.install === 'function') {
        pluginMain.install(this)
      }
      this.plugins.set(manifest.name, {
        manifest,
        dir: pluginDir,
        dirName,
        loadedAt: new Date().toISOString()
      })
      console.log(`${c.green}✅ [PluginManager]${c.reset} 已加载: ${c.cyan}${manifest.name}${c.reset} v${manifest.version}`)
      return true
    } catch (err) {
      console.error(`${c.yellow}⚠️ [PluginManager]${c.reset} "${manifest.name}" 加载失败: ${err.message}`)
      return false
    }
  }

  /**
   * 卸载插件（清除钩子 + 清除缓存）
   */
  unloadPlugin(pluginName) {
    // 清除该插件的所有钩子
    for (const [hookName, listeners] of this.hooks.entries()) {
      this.hooks.set(hookName, listeners.filter(l => l.pluginName !== pluginName))
    }
    this.plugins.delete(pluginName)
    console.log(`${c.gray}🗑️ [PluginManager]${c.reset} 已卸载: ${pluginName}`)
  }

  /**
   * 扫描并加载所有插件
   */
  _scanAndLoad() {
    if (!fs.existsSync(this.pluginsRoot)) return

    const entries = fs.readdirSync(this.pluginsRoot, { withFileTypes: true })
    const pluginDirs = entries.filter(e => e.isDirectory())

    if (pluginDirs.length === 0) {
      console.log(`${c.gray}📁 [PluginManager]${c.reset} 用户插件目录为空 (${this.pluginsRoot})`)
      return
    }

    console.log(`${c.cyan}📦 [PluginManager]${c.reset} 发现 ${pluginDirs.length} 个用户插件，开始加载...`)
    let loaded = 0
    for (const entry of pluginDirs) {
      if (this.loadPlugin(path.join(this.pluginsRoot, entry.name))) {
        loaded++
      }
    }
    console.log(`${c.green}📦 [PluginManager]${c.reset} 共加载 ${loaded}/${pluginDirs.length} 个插件`)
  }

  /**
   * 重新扫描目录（安装/卸载后调用）
   */
  rescan() {
    // 清空现有钩子和插件
    this.hooks.clear()
    this.plugins.clear()
    this._scanAndLoad()
  }

  /**
   * 获取已加载的插件列表
   */
  getLoadedPlugins() {
    return Array.from(this.plugins.entries()).map(([name, info]) => ({
      name,
      version: info.manifest.version,
      description: info.manifest.description,
      author: info.manifest.author,
      dirName: info.dirName,
      loadedAt: info.loadedAt
    }))
  }

  /**
   * 获取用户插件目录路径
   */
  getPluginsRoot() {
    return this.pluginsRoot
  }
}

module.exports = new PluginManager()
