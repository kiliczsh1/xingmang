# 🔌 星芒写作 — 插件 API 完全参考手册

> 版本：v1.0 | 更新：2026-07-11

---

## 1. 插件结构

一个插件就是一个文件夹，放在 `data/user-plugins/` 下：

```
data/user-plugins/my-plugin/
├── plugin.json    ← 清单（必须）
├── index.js        ← 后端代码（可选）
└── frontend.js     ← 前端代码（可选）
```

### 1.1 plugin.json

```json
{
  "name": "my-plugin",
  "version": "1.0.0",
  "description": "简短描述",
  "author": { "name": "你的名字" },
  "hooks": {
    "ui.books.mounted": "onBooksPageReady"
  },
  "permissions": ["ui.style", "ui.books.read"],
  "interface": {
    "displayName": "我的插件",
    "shortDescription": "一句话说明",
    "category": "UI",
    "capabilities": ["Theme", "Interactive"],
    "brandColor": "#f59e0b"
  }
}
```

### 1.2 index.js（后端）

```js
module.exports = {
  install(manager) {
    manager.on('book.created', (data) => {
      console.log('新书:', data.book.title)
    })
  }
}
```

### 1.3 frontend.js（前端）

```js
(function(plugin, manifest) {
  'use strict'
  plugin.on('ui.books.mounted', function(data) {
    // data.container → .books-container DOM 元素
    console.log('我的小说页面已加载')
  })
})(/* plugin 参数由系统传入 */)
```

---

## 2. 前端钩子（完整列表）

> 在 frontend.js 中用 `plugin.on('钩子名', callback)` 监听

### 2.1 📚 我的小说（Books）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.books.mounted` | 进入页面 | `{ container: HTMLElement, bookList: Book[] }` |
| `ui.book.created` | 创建作品 | `{ book: Book }` |
| `ui.book.updated` | 编辑作品 | `{ id: number, title: string }` |
| `ui.book.deleted` | 删除作品 | `{ id: number, title: string }` |

### 2.2 ✍️ 写作（Write）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.write.mounted` | 进入写作页面 | `{ bookId: number }` |

### 2.3 🏆 AI扫榜（AIRank）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.airank.mounted` | 进入页面 | — |

### 2.4 💬 提示词库（Prompts）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.prompts.mounted` | 进入页面 | — |
| `ui.promptpreview.mounted` | 进入提示预览 | — |

### 2.5 ⭐ 变量抽卡（Creative）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.creative.mounted` | 进入页面 | — |

### 2.6 🔗 工作流（Workflow）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.workflow.mounted` | 进入工作流列表 | — |
| `ui.workflow.workbench.mounted` | 进入工作流编辑器 | `{ workflowId: number }` |

### 2.7 🕸️ wiki图谱（WikiGraph）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.wikigraph.mounted` | 进入页面 | — |

### 2.8 📖 拆书库（BookAnalysis）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.analysis.mounted` | 进入页面 | `{ bookId: string }` |

### 2.9 👤 角色库（CharacterLibrary）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.characters.mounted` | 进入页面 | `{ bookId: string }` |

### 2.10 📝 经验分享（ExperienceShares）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.experienceshares.mounted` | 进入列表页 | — |

### 2.11 ⚙️ API配置（Config）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.config.mounted` | 进入页面 | — |

### 2.12 👨‍💻 个人中心（Profile）

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.profile.mounted` | 进入页面 | — |

### 2.13 🎨 主题

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `ui.theme.change` | 切换亮色/暗色 | `{ theme: 'light' \| 'dark' }` |

---

## 3. 后端钩子（完整列表）

> 在 index.js 中用 `manager.on('钩子名', callback)` 监听

### 3.1 📚 Books

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `book.created` | 创建作品 | `{ book: { id, title, description, ... } }` |
| `book.updated` | 更新作品 | `{ id, data, book }` |
| `book.deleted` | 删除作品 | `{ id }` |

### 3.2 📄 Chapters

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `chapter.created` | 创建章节 | `{ chapter }` |
| `chapter.updated` | 更新章节 | `{ id, chapter }` |
| `chapter.deleted` | 删除章节 | `{ id }` |

### 3.3 💬 Prompts

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `prompt.created` | 创建提示词 | `{ prompt }` |
| `prompt.updated` | 更新提示词 | `{ id, prompt }` |
| `prompt.deleted` | 删除提示词 | `{ id }` |

### 3.4 👤 Characters

| 钩子 | 触发时机 | 数据 |
|------|---------|------|
| `character.created` | 创建角色 | `{ character }` |
| `character.updated` | 更新角色 | `{ id, character }` |
| `character.deleted` | 删除角色 | `{ id }` |

---

## 4. 前端 API（plugin 对象的方法）

| 方法 | 说明 |
|------|------|
| `plugin.on(name, callback, pluginName)` | 注册钩子 |
| `plugin.trigger(name, data)` | 手动触发钩子 |
| `plugin.injectStyle(pluginName, css, styleId?)` | 注入 CSS |
| `plugin.removeStyle(pluginName, styleId?)` | 移除 CSS |
| `plugin.registerDomElement(pluginName, el)` | 注册 DOM（禁用时自动清理） |
| `plugin.isEnabled(pluginName)` | 检查是否启用 |
| `plugin.getLoadedPlugins()` | 获取已加载插件列表 |

## 5. 后端 API（manager 对象的方法）

| 方法 | 说明 |
|------|------|
| `manager.on(name, callback, pluginName)` | 注册钩子 |
| `manager.trigger(name, data)` | 触发钩子 |
| `manager.getLoadedPlugins()` | 获取已加载插件列表 |

---

## 6. 系统 CSS 变量

```css
/* 色调海拔（暗色主题下） */
--surf-0: #060d17    --surf-3: #13243e
--surf-1: #0a1424    --surf-4: #192d4c
--surf-2: #0e1c31    --surf-5: #1f3659

/* 语义化变量 */
--bg-primary    --card-bg       --dialog-bg
--bg-secondary  --card-hover-bg --dropdown-bg
--sidebar-background  --content-background

/* 品牌色 */
--menu-hover-color: #2dd4bf
```

> 完整变量见 `novel/src/styles/dark-theme.css`

---

## 7. 完整示例

### 7.1 plugin.json
```json
{
  "name": "export-txt",
  "version": "1.0.0",
  "description": "在作品卡片上加一键导出 TXT 按钮",
  "author": { "name": "你的名字" },
  "permissions": ["ui.books.read", "ui.style"],
  "interface": {
    "displayName": "导出 TXT",
    "shortDescription": "作品卡片增加 TXT 导出按钮",
    "category": "Utility",
    "capabilities": ["Export"],
    "brandColor": "#10b981"
  }
}
```

### 7.2 index.js
```js
module.exports = {
  install(manager) {
    manager.on('book.created', (data) => {
      console.log('[export-txt] 新书:', data.book.title)
    }, 'export-txt')
  }
}
```

### 7.3 frontend.js
```js
(function(plugin) {
  'use strict'
  plugin.on('ui.books.mounted', function(data) {
    document.querySelectorAll('.book-card').forEach(function(card) {
      if (card.querySelector('.plugin-export-btn')) return
      var btn = document.createElement('button')
      btn.className = 'plugin-export-btn'
      btn.textContent = '📄 TXT'
      btn.onclick = function(e) {
        e.stopPropagation()
        alert('导出功能！')
      }
      var actions = card.querySelector('.card-actions')
      if (actions) {
        actions.insertBefore(btn, actions.firstChild)
        plugin.registerDomElement('export-txt', btn)
      }
    })
  })
})
```
