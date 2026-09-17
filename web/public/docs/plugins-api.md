# 星芒 AI · 前端 API 参考

> 全项目 17 模块 · 142 方法 ｜ 插件 26 方法 ｜ 组件 API ｜ 版本 v1.3 · 2026-08-10
> 覆盖全项目 API（book / chapter / prompt / memo / … / backup）+ 插件系统专项 + 常用组件接口

---

## 1. 全项目 API 参考

> `src/api/index.ts` · 17 个模块 142 个方法 + `backup.ts`

### 1.0 全局约定

**axios 实例**：`baseURL: '/api'`，默认超时 `180000ms`（知识图谱分析 / AI 生成类接口单独放大）。

**响应拦截器**：成功返回 `response.data`（即 `ApiResponse<T>` 整体）；错误自动 `ElMessage.error`，调用方可通过 `{ silent: true }` 抑制并自行回退。

```ts
interface ApiResponse<T> {
  success: boolean
  data: T
  message?: string
}
```

| 特殊约定 | 说明 |
|---|---|
| promptAPI | `fields` / `subcategories` 后端存 JSON 字符串，API 层统一 `parseJsonArray` 解析为数组 |
| backupAPI | 独立于 index.ts，原始 axios（不经拦截器），导出走 `responseType: 'blob'` 触发下载 |
| rankAPI | 全部 `silent: true`，失败由调用方回退演示数据 |
| 分页参数 | rankAPI 用 `page/size`；entryAPI 用 `page/page_size`；pluginAPI 用 `page/pageSize` |

#### 页面 → API 模块映射

| 路由 / 页面 | 使用的 API 模块 |
|---|---|
| `/ai-ranking` | rankAPI |
| `/analysis/:bookId` | chapterAPI configAPI promptAPI |
| `/books` | aiAPI chapterAPI configAPI promptAPI |
| `/config` | configAPI providerAPI |
| `/creative` | bookAPI chapterAPI characterAPI configAPI generatorAPI memoAPI promptAPI |
| `/experience-shares(/:id)` | experienceShareAPI |
| `/prompts` 系列 | promptAPI |
| `/profile` | bookAPI generatorAPI promptAPI statsAPI |
| `/wiki-graph` | bookAPI chapterAPI configAPI knowledgeGraphAPI |
| `/workflow/:id` | configAPI promptAPI |
| `/write/:bookId` | chapterAPI characterAPI configAPI conversationAPI memoAPI promptAPI volumeAPI |
| `/plugins` | 本地管理 · 插件专项见 Part 2 |
| 组件：ChatDialog / ChatMain / EntryLibrary / KnowledgeGraph / MainLayout | bookAPI configAPI conversationAPI / promptAPI / entryAPI / knowledgeGraphAPI / configAPI statsAPI |

> **backupAPI**（备份导入导出）当前仅在 backup.ts 定义，尚未接入页面。

### 1.1 书本 / 章节 / 分卷

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| bookAPI | getAll | `GET /books` | — | `Book[]` |
| bookAPI | getOne / create / update | `/books(/:id)` | `Partial<Book>` | `Book` |
| bookAPI | delete | `DELETE /books/:id` | — | `ApiResponse` |

> bookAPI 共 5 个方法（getAll / getOne / create / update / delete）

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| chapterAPI | getByBook | `GET /chapters/book/:bookId` | — | `Chapter[]` |
| chapterAPI | getOne / create / update / delete | `/chapters(/:id)` | `Partial<Chapter>` | `Chapter` |
| chapterAPI | importBook | `POST /chapters/import-book` | `{ bookId, chapters[] }` | `{ insertedCount, chapters }` |
| chapterAPI | importFile | `POST /chapters/import-file` | `{ bookId, file: { name, size, data_base64 } }` | `{ insertedCount, chapters }` |

> chapterAPI 共 7 个方法

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| volumeAPI | getByBook | `GET /volumes/book/:bookId` | — | `Volume[]` |
| volumeAPI | getOne / create / update | `/volumes(/:id)` | `Partial<Volume>` | `Volume` |
| volumeAPI | delete | `DELETE /volumes/:id` | — | `ApiResponse` |

### 1.2 提示词 / 备忘录 / 经验分享

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| promptAPI（5） | getAll | `GET /prompts` | — | `Prompt[]`（字段已解析） |
| promptAPI（5） | getOne / create / update / delete | `/prompts(/:id)` | `Partial<Prompt>` | `Prompt` |

> 消费方无需处理 `fields` / `subcategories` 字符串

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| memoAPI（7） | getAll / getOne | `GET /memos(/:id)` | — | `Memo[]` / `Memo` |
| memoAPI（7） | create / update | `/memos(/:id)` | `Partial<Memo>` | `Memo` |
| memoAPI（7） | delete | `DELETE /memos/:id` | — | `ApiResponse` |
| memoAPI（7） | search | `GET /memos/search` | `keyword` / `category` | `Memo[]` |
| memoAPI（7） | batch | `POST /memos/batch` | `{ action, ids[], data? }` | `ApiResponse` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| experienceShareAPI（8） | getAll | `GET /experience-shares` | — | `ExperienceShare[]` |
| experienceShareAPI（8） | getOne / create / update / delete | `/experience-shares(/:id)` | — | `ExperienceShare` |
| experienceShareAPI（8） | importPdf | `POST /experience-shares/import-pdf` | `{ pdf_file: { name, size, data_base64 } }` | `Partial<ExperienceShare>` |
| experienceShareAPI（8） | exportAll | `GET /experience-shares/export` | — | `ExperienceShare[]` |
| experienceShareAPI（8） | importAll | `POST /experience-shares/import` | `{ cards[] }` | `ApiResponse` |

### 1.3 服务商与模型配置

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| providerAPI（6） | getAll / getOne / create / update / delete | `/providers(/:id)` | — | `ApiProvider` |
| providerAPI（6） | discoverModels | `GET /providers/:id/models` | — | `{ id, name, max_tokens? }[]` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| configAPI（9） | getAll | `GET /config` | — | `ApiModel[]` |
| configAPI（9） | getDefault | `GET /config/default` | — | `ApiModel` |
| configAPI（9） | getByProvider | `GET /config/provider/:providerId` | — | `ApiModel[]` |
| configAPI（9） | create / update / delete | `/config(/:id)` | — | `ApiModel` |
| configAPI（9） | toggleModel | `PUT /config/:id/toggle` | — | `{ id, enabled }` |
| configAPI（9） | reorderModels | `PUT /config/reorder/all` | `{ ids[] }` | `ApiResponse` |
| configAPI（9） | testModel | `POST /config/:id/test` | — | `{ modelId, modelName, providerName?, latencyMs, preview }` |

### 1.4 AI 能力与扫榜

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| aiAPI（6） | chat | `POST /ai/chat` | `{ messages, configId?, systemPrompts?, relatedContent? }` | `string` |
| aiAPI（6） | generateDescription | `POST /ai/generate-description` | `{ title, promptId? }` | `string` |
| aiAPI（6） | generateImage | `POST /ai/generate-image` | `{ messages, configId?, size?, quality?, style?, reference_images? }` | `{ url }` |
| aiAPI（6） | getRankReferenceCategories | `GET /ai/rank-reference/categories` | — | `{ key, label, categories[] }[]` |
| aiAPI（6） | getRankReferenceBooks | `GET /ai/rank-reference/books` | `{ path }` | `{ rank, title, author, coverUrl, bookPath }[]` |
| aiAPI（6） | recognizeCharacters | `POST /ai/recognize-characters` | `{ text, configId?, customPrompt? }` | `any[]` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| rankAPI（2） | getLists | `GET /ai/rank/lists` | —（silent） | `{ groups: RankGroup[] }` |
| rankAPI（2） | getBooks | `GET /ai/rank/books` | `tab` / `page` / `size`（silent） | `{ books: RankBook[] }` |

> **RankBook / RankGroup** 内联定义于 `src/api/index.ts`（仅被 API 消费的响应形状）。番茄榜单经 Go 后端转调 Python 桥。

### 1.5 对话（创作侧聊天）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll / list | `GET /conversations` | — | `any[]` |
| getByBook | `GET /conversations/book/:bookId` | — | `any[]` |
| create | `POST /conversations` | `{ book_id, title? }` | `any` |
| update / delete | `/conversations/:id` | `{ title }` / — | `any` / `ApiResponse` |
| getMessages | `GET /conversations/:id/messages` | — | `any[]` |
| saveMessage | `POST /conversations/:id/messages` | `{ role, content, displayContent? }` | `any` |
| clearMessages | `DELETE /conversations/:id/messages` | — | `ApiResponse` |
| deleteMessage | `DELETE /conversations/:conversationId/messages/:messageId` | — | `ApiResponse` |
| updateMessage | `PUT /conversations/:conversationId/messages/:messageId` | `{ content, displayContent? }` | `any` |

> conversationAPI 共 11 个方法

### 1.6 生成器 / 角色 / 知识图谱

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| generatorAPI（5） | getAll | `GET /generators` | — | `Generator[]` |
| generatorAPI（5） | getOne / create / update / delete | `/generators(/:id)` | — | `Generator` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| characterAPI（8） | getByBook | `GET /characters/book/:bookId` | — | `Character[]` |
| characterAPI（8） | getOne / create / update / delete | `/characters(/:id)` | — | `Character` |
| characterAPI（8） | batchDelete / batchMove | `POST /characters/batch-delete` / `batch-move` | `{ ids[] }` / `{ ids[], folder }` | `ApiResponse` |
| characterAPI（8） | getFolders | `GET /characters/folders/:bookId` | — | `string[]` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| knowledgeGraphAPI（19） | getGraph / clearGraph / importGraph | `/knowledge-graph/graph/:bookId` | — | `KnowledgeGraphData` / `ApiResponse` |
| knowledgeGraphAPI（19） | getEntities / getEntity · 实体 CRUD | `/knowledge-graph/entities*` | — | `GraphEntity` |
| knowledgeGraphAPI（19） | getRelations / 关系 CRUD | `/knowledge-graph/relations*` | — | `GraphRelation` |
| knowledgeGraphAPI（19） | analyze | `POST /knowledge-graph/analyze` | `{ bookId, configId?, scope?, chapterIds?, versionId? }`（300s） | `KnowledgeGraphData` |
| knowledgeGraphAPI（19） | aiSupplement | `POST /knowledge-graph/ai-supplement` | `{ bookId, configId, entityName, entityType, entityDescription, hint?, chapterIds? }`（120s） | `{ content }` |
| knowledgeGraphAPI（19） | 版本系列（5 个） | `/knowledge-graph/versions*` | — | `GraphVersion` |

### 1.7 词条库 / 统计 / 备份

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| entryAPI（8） | getList | `GET /entries` | `book_id` / `category_type` / `keyword` / `page` / `page_size` | `EntryListResponse` |
| entryAPI（8） | getOne / create / update / delete | `/entries(/:id)` | — | `Entry` |
| entryAPI（8） | addToBook / getCategories | `POST /entries/:id/add-to-book` · `GET /entries/categories` | `{ bookId }` / `book_id?` | `Entry` / `EntryCategory[]` |
| entryAPI（8） | aiAutoComplete | `POST /entries/:id/ai-auto-complete` | `{ configId }`（120s） | `Entry` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| statsAPI（5） | getOverview | `GET /stats/overview` | — | `UsageOverview` |
| statsAPI（5） | getDaily | `GET /stats/daily` | `startDate` / `endDate` | `DailyUsage[]` |
| statsAPI（5） | getModelStats / getMonthly | `GET /stats/model-stats` · `/stats/monthly` | — | `ModelStats[]` / `MonthlyStats[]` |
| statsAPI（5） | record | `POST /stats/record` | `{ modelId, modelName, providerName, tokens? }` | `ApiResponse` |

| 模块 | 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|---|
| backupAPI（backup.ts） | downloadBackup | `GET /backup/export?mask=0\|1` | blob 下载，解析 Content-Disposition 文件名 | `{ filename, size }` |
| backupAPI（backup.ts） | preview | `POST /backup/preview` | 备份 payload | `BackupPreview` |
| backupAPI（backup.ts） | importBackup | `POST /backup/import` | `{ payload, mode: overwrite\|merge, modules[] }`（600s） | `BackupImportResult` |

> **backupAPI** 走原始 axios（不经统一拦截器），当前尚未接入页面。

---

## 2. 插件系统 API

> pluginAPI 26 方法 · 运行时契约 · 插件中心 Hub

### 2.0 概述与核心区分

插件机制定位为**聊天消息发送前的内容处理链**：已启用插件按注册顺序依次处理消息内容，前一个插件的输出是后一个插件的输入。

```
用户输入 + 关联内容 ──▶ [插件 1] ──▶ [插件 2] ──▶ … ──▶ 最终发送内容
```

| 维度 | 内置 `builtin` | 本地 `local` | 第三方 `market` |
|---|---|---|---|
| 来源 | 随客户端编译期附带 | 从本地 .zip 包导入 | 市场（本期**待开放**，空态引导） |
| 是否可安装 / 卸载 | 否（预装） | 可安装、可卸载 | 待开放（预留 REST API） |
| 运行代码 | 编译进包 | zip 内入口代码，本机执行 | — |
| 启停影响 | 模块单例 → 聊天链实时生效 | 模块单例 → 聊天链实时生效 | — |
| 持久化 | `plugin-prefs` | `plugin-local` + `plugin-prefs` | 接入后端后由 REST 承载 |

### 2.1 核心接口

#### ChatPlugin（运行时执行契约）· `plugins/type.ts`

```ts
export interface ChatPlugin {
  name: string          // 唯一 ID，kebab-case 英文，如 'prompt-inject'
  enabled: boolean      // 是否参与处理链
  description: string   // 一句话中文说明（管理页展示）
  process: (content: string, context: PluginContext) => string
}
```

#### PluginContext（处理时上下文）

```ts
export interface PluginContext {
  input: string          // 用户原始输入（未处理）
  relatedContent: Array<{ type: string; id: number | string; title: string; content: string }>
  selectedPrompts: any[] // 用户选中的提示词
  worldBookLinked: boolean
  worldBookEntries?: any[]
  [key: string]: any     // 允许扩展
}
```

#### PluginManifest（管理面展示契约）· `types/index.ts`

```ts
export type PluginSource = 'builtin' | 'marketplace'
export interface PluginManifest {
  name: string; label: string; description: string; enabled: boolean
  source: PluginSource; icon: string; color: string
  version?: string; author?: string   // 预留
}
```

#### HubItem（插件中心内部条目）· `Plugins.vue`

```ts
interface HubItem {
  n: string; src: 'builtin'; c: string              // 名称 / 来源 / 分类
  d: string; r: number                              // 描述 / 评分
  color: string; icon: string
  t: string[]; kw: string[]                         // 标签 / 搜索关键词
  on: boolean                                       // 已启用（内置恒为已安装）
  func: [string, string][]; perm: [string, string][]; hist: [string, string][]
}
```

### 2.2 注册表 API（useChatPlugins）

工厂函数，每次调用创建**会话级**注册表实例（生命周期 = 一次 Write 页会话）。

```ts
const chat = useChatPanel({ bookId })
chat.pluginChain.register(plugin)   // 注册（同名自动去重）
chat.pluginChain.unregister(name)   // 注销
chat.pluginChain.getPlugin(name)    // 查询
chat.pluginChain.togglePlugin(name, enabled?)  // 切换启用状态
const result = chat.pluginChain.process(context) // 链式处理
```

| 方法 | 行为 |
|---|---|
| register | 已存在同名插件时静默跳过 |
| unregister / getPlugin | 从链中移除 / 按名查询 |
| togglePlugin | 省略 enabled 时取反 |
| process | 按注册顺序执行所有 enabled 的插件，每次实时读取 enabled（非注册时快照） |

### 2.3 内置插件清单（注册顺序 = 处理顺序）

| 顺序 | name | 可读名 | 说明 |
|---|---|---|---|
| 1 | `prompt-inject` | 提示词注入 | 核心拼装：提示词 + 关联内容 → 用户消息前 |
| 2 | `worldbook-inject` | 世界书注入 | 命中才注入世界书词条 |
| 3 | `related-content-inject` | 关联内容注入 | 扩展点，当前 no-op |
| 4 | `regex-filter` | 正则过滤 | 显示层处理，链上 no-op |

> **最终格式：** 提示词 → 世界书 → 关联内容 → 用户输入。

### 2.4 插件中心 Hub（Plugins.vue）

```
┌─ 插件市场（默认）─────────────────────────────┐
　搜索(联想补全) · 来源分段(全部/内置/第三方) · 历史 · 标签云
　分类 Tab · 内置插件卡片网格(内置徽章)
　第三方空态区(市场开发中 → 演示合集) · 点卡片 → 详情抽屉
└─────────────────────────────────────────┘
┌─ 运行诊断 ──────────────────────────────────┐
　健康度总览 · 严重度过滤 · 日志流(可展开堆栈) · 复制诊断信息
└─────────────────────────────────────────┘
```

- **内置条目**：来自 `BUILTIN_PLUGINS` + `BUILTIN_META`（真实数据），`on` 经 `getPluginEnabled` 读 `plugin-prefs`；搜索标签/关键词来自 `BUILTIN_TAGS` / `BUILTIN_KW`。
- **第三方**：本期不展示数据，来源分段切「第三方」时显示空态引导（C4），跳转 `/plugin-demos/`。
- **操作写回**：内置启停 → `setBuiltinEnabled()`（改模块单例 + 写 `plugin-prefs`）→ 同 SPA 会话内聊天链实时生效。

#### localStorage 约定

| key | 内容 | 语义 |
|---|---|---|
| `plugin-prefs` | `{ "prompt-inject": false }` | 内置 + 本地插件启停覆盖，缺省启用，只存显式覆盖 |
| `plugin-local` | `[{ manifest, entryCode, installedAt }]` | 本地安装的插件包（manifest + 入口代码源码），跨刷新重放 |

### 2.5 本地插件安装（zip 导入）

插件中心「安装本地插件」支持从本地 `.zip` 包安装自研插件。实现：`plugins/local.ts`。

#### 包格式

```
plugin.zip
├── manifest.json   # 必需：{ name, label, version, description?, author?, icon?, color?, tags?, entry? }
└── <entry>.js      # 可选：入口代码
```

#### 入口代码契约（entry 文件）

```js
window.__installPlugin = {
  name: 'example-suffix',
  description: '一句话说明',
  enabled: true,
  process: function (content, context) {
    return content // 返回处理后的内容
  }
}
```

- **执行方式**：`new Function('window', code)(window)`，本机执行；安装含运行代码的包前有确认弹窗。
- **ES 模块兼容**：入口文件若写成 `export default { … }`（ESM 风格），加载器自动改写为 `window.__installPlugin` 赋值再执行；`main` / `entry` 字段均可指定入口（忽略前导 `./`）。
- **副作用加载时机**：UI 类插件（如主题按钮）在应用启动时挂载（`App.vue` 调 `loadLocalPlugins()`）。
- **降级**：执行失败 / 无入口 → 注册为 no-op 扩展点（占位，不阻断安装）。
- **运行**：本地插件是模块级单例（`LOCAL_PLUGINS`），ChatMain 注册内置后追加注册，参与聊天链。
- **API**：`loadLocalPlugins` / `addLocalPlugin` / `removeLocalPlugin` / `setLocalPluginEnabled` / `getLocalRecords` / `safeIcon` / `resolveEntryName`。
- **限制**：7z 暂不支持在线解析（JSZip 仅支持 zip），请转 zip 后导入。示例包：`docs/examples/example-plugin.zip`。

#### UI 类插件（全局 UI 集成）

UI 类插件**不参与聊天内容处理链**（`process` 原样返回），入口代码的副作用在应用启动时挂载，可向页面注入全局 UI（如右下角按钮）。**不要**在插件里重新实现应用已有功能（如主题切换）——应通过事件桥调用应用自身逻辑，应用保持唯一状态源。

**主题事件桥**（`MainLayout.vue` 提供）：

| 事件（window） | 触发方 | 行为 |
|---|---|---|
| `xingmang:theme-toggle` | UI 插件 | 触发应用主题切换，**等价于点击顶栏「外观」开关** |
| `xingmang:theme-set` | UI 插件 | 设置指定主题，`detail: { theme: 'light' \| 'dark' }` |
| `xingmang:theme-changed` | 应用 | 主题变化后广播，`detail: { theme }`，UI 插件据此同步自身状态 |

主题状态源：`localStorage['main-theme']`（`'light' \| 'dark'`）；样式开关：`<html>` 上的 `data-theme='dark'` 属性（`:root[data-theme='dark']` 选择器）。

**示例：theme-toggle（暗亮切换快捷按钮）**——右下角圆形按钮，点击派发 `xingmang:theme-toggle`，监听 `xingmang:theme-changed` 同步图标：

```js
window.__installPlugin = {
  name: 'theme-toggle',
  description: '全局右下角圆形按钮，一键切换暗色 / 亮色主题',
  enabled: true,
  process: function (content) { return content } // UI 插件不处理内容
}

// 点击按钮 → 触发应用主题切换（等价顶栏外观开关）
window.dispatchEvent(new CustomEvent('xingmang:theme-toggle'))

// 应用切换后广播 → 刷新按钮图标
window.addEventListener('xingmang:theme-changed', function (e) {
  var dark = document.documentElement.hasAttribute('data-theme')
  btn.textContent = dark ? '☀️' : '🌙'
})
```

> 完整源码见 `docs/examples/theme-toggle/`，打包 zip 见 `docs/examples/theme-toggle.zip`。

### 2.6 pluginAPI（26 方法 · 7 组 · 后端待实现）

> **前端客户端已落地** `src/api/index.ts`；后端按此契约实现后，Plugins.vue 第三方空态区可无缝替换为真实数据。

| 组 | 方法 | 路径 | 响应 |
|---|---|---|---|
| A. 市场发现 | getMarketplace | `GET /plugins/marketplace` | `{ list, total }` |
| A. 市场发现 | getMarketCategories / getMarketTags | `/plugins/marketplace/categories` · `/tags` | 分类 / 标签云 |
| A. 市场发现 | getRecommended | `GET /plugins/marketplace/recommended` | 精选推荐 |
| A. 市场发现 | getMarketItem | `GET /plugins/marketplace/:id` | 市场详情 |
| A. 市场发现 | getReviews / addReview | `GET\|POST /plugins/marketplace/:id/reviews` | 评价 |
| B. 已安装管理 | getInstalled / getOne | `GET /plugins(/:id)` | `PluginManifest` |
| B. 已安装管理 | install / uninstall | `POST /plugins` · `DELETE /plugins/:id` | `PluginManifest` / `ApiResponse` |
| B. 已安装管理 | toggle | `PUT /plugins/:id/toggle` | `PluginManifest` |
| B. 已安装管理 | batch | `POST /plugins/batch` | `{ action, ids[] }` |
| B. 已安装管理 | detect | `GET /plugins/detect` | `{ report }` |
| C. 更新 | getUpdates | `GET /plugins/updates` | `PluginUpdateInfo[]` |
| C. 更新 | updateOne / updateAll | `PUT /plugins/:id/update` · `/plugins/updates` | `PluginManifest` / `{ updated }` |
| C. 更新 | getChangelog | `GET /plugins/:id/changelog` | 版本历史 |
| D. 依赖冲突 | getDependencies | `GET /plugins/:id/dependencies` | `PluginDependency[]` |
| D. 依赖冲突 | resolveConflicts | `POST /plugins/:id/dependencies/resolve` | `{ fixed[], ignored[] }` |
| E. 配置 | getSettings / saveSettings | `GET\|PUT /plugins/:id/settings` | `{ schema, values }` |
| F. 诊断 | getDiagnostics | `GET /plugins/diagnostics` | `PluginDiagnostics` |
| F. 诊断 | getLogs / getLogDetail | `GET /plugins/logs(/:logId)` | `PluginLog` |
| G. 统计 | getStats | `GET /plugins/stats` | `{ installs, activeCount, topPlugins[] }` |

### 2.7 如何编写一个新插件

**两种路径**：内置插件（编译进包，改前端代码）或本地插件（zip 包，无需改前端）。本地包步骤：

1. **编写包**：`manifest.json` + 入口 `index.js`（契约见 2.5），参考 `docs/examples/example-plugin/`。
2. **打包 zip**：确保 `manifest.json` 在 zip 根目录。
3. **导入**：插件中心 →「安装本地插件」→ 选择 .zip（含运行代码时确认执行）。
4. **验证**：出现在列表（本地徽章）、启停生效、聊天链可观察到处理效果。

内置插件步骤：

1. **定义对象**：在 `plugins/` 新建文件，实现 `ChatPlugin`。
2. **实现 process**：返回处理后的内容，不处理时原样返回。
3. **加入清单**：`BUILTIN_PLUGINS` 追加（顺序即处理顺序）+ `BUILTIN_META` 补充元数据。
4. **展示元数据**：`Plugins.vue` 的 `BUILTIN_DETAIL` / `BUILTIN_TAGS` / `BUILTIN_KW`。
5. **持久化**：无需改动，缺省启用，覆盖自动进 `plugin-prefs`。
6. **测试**：Write 页会话验证插件链输出；`/plugins` 验证展示/搜索/抽屉/启停联动。

### 2.8 已知边界

- regex-filter 是显示层扩展点，链上 process 为 no-op；related-content-inject 为占位扩展点。
- **本地插件代码在本机执行**（`new Function`），安装前有确认弹窗；仅导入可信来源的包。
- 7z 暂不支持在线解析，请转 zip 后导入。
- 第三方市场待开放，页面显示空态引导（C4）；接入后端后替换为 pluginAPI 渲染。
- 内置/本地插件启停联动依赖模块单例 + localStorage；多标签页不同步为已知限制。
- 插件中心仅一层详情抽屉（B2 合规），安装确认用 ElMessageBox（仍在深度 2 内）。
- 演示合集：`/plugin-demos/`（7 个 HCI 交互 Demo）。

---

## 3. 组件 API（前端组件接口）

> 前端组件通过 **props / emits / `defineExpose`** 暴露接口，供父组件与页面调用。本节收录常用组件。

### 3.1 SplitRichTextEditor（富文本编辑器）

封装 Vditor 的写作编辑器组件（`src/components/SplitRichTextEditor.vue`），用于章节正文 / 简介 / 提示词等富文本编辑。使用页：`Write.vue` / `Prompts.vue` / `PromptPreview.vue` / `PackDetail.vue`。

#### Props

| prop | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `modelValue` | `string` | — | 编辑器内容（v-model 双向绑定，Markdown 格式） |
| `placeholder` | `string` | `'请输入简介内容'` | 空内容占位提示 |
| `disabled` | `boolean` | `false` | 只读禁用 |
| `fontFamily` | `string` | `'Microsoft YaHei'` | 编辑区字体 |
| `fontSize` | `number` | `16` | 编辑区字号（px） |

#### Emits

| 事件 | 载荷 | 说明 |
|---|---|---|
| `update:modelValue` | `string` | 内容变化（v-model） |
| `blur` | — | 失焦（Write.vue 用于自动保存章节） |

#### 暴露方法（defineExpose）

| 方法 | 签名 | 说明 |
|---|---|---|
| `insertPlainText` | `(text: string) => void` | 光标处插入纯文本 |
| `wrapSelectionWithHtml` | `(html: string) => void` | 将 HTML 转 Markdown 后插入光标处 |
| `insertContent` | `(html: string) => void` | `wrapSelectionWithHtml` 别名 |
| `insertValue` | `(text: string) => void` | 光标处插入文本（Vditor `insertValue`） |
| `getValue` | `() => string` | 获取当前内容（Markdown） |
| `getHtml` | `() => string` | 获取当前内容（HTML） |
| `setValue` | `(value: string) => void` | 外部写入内容（走同步逻辑，避免回显误判） |
| `clear` | `() => void` | 清空内容 |
| `getSelectedText` | `() => string` | 获取选中文本 |
| `getSelection` | `() => string` | `getSelectedText` 别名 |
| `getCursorPosition` | `() => object` | 光标位置（Vditor `getCursorPosition`） |
| `focusEditor` | `() => void` | 聚焦编辑器 |
| `focus` | `() => void` | `focusEditor` 别名 |
| `disabled` | `(val: boolean) => void` | 切换只读状态 |
| `enable` | `() => void` | 解除只读 |
| `vditor` | `() => Vditor \| null` | 返回原始 Vditor 实例（可调用任意 Vditor API） |

> 示例：`chapterEditorRef.value?.insertPlainText('\n\n— · —\n\n')`（Write.vue 插入分隔线）。

### 3.2 MainLayout（主布局 · 顶栏按钮）

应用外壳布局（`src/layouts/MainLayout.vue`），常驻挂载（隐藏顶栏路由除外）。顶栏按钮是全局导航入口，动作由 `handleGlobalNavClick(label)` 统一分发。

#### 顶栏按钮

| 按钮 | 图标 | 触发动作 |
|---|---|---|
| 最新下载 | `Download` | 外部链接 `https://www.123912.com/s/cc0dTd-GDNGh` |
| 主题切换 | `Sunny` / `Moon` | `toggleTheme()`（亮↔暗，见下方主题事件桥） |
| 对话 | `ChatLineRound` | 打开对话弹窗（`chatDialogVisible = true`） |
| 教程 | `School` | `ElMessage.info('功能开发中')` |
| 历史 | `Clock` | 打开历史弹窗（`historyDialogVisible = true`） |
| 消息 | `Message` | 打开消息中心（`messageCenterVisible = true`） |

> **外观 / 论坛** 在 `globalNavItems` 中定义但被顶栏 `v-for` 过滤（`i.label !== '外观' && i.label !== '论坛'`）；`handleGlobalNavClick` 仍支持：外观 → 主题选择弹窗（`themeDialogVisible = true`），论坛 → 外部链接 `https://kook.vip/sLySCy`。

#### 主题事件桥（主题切换按钮的 API）

| 事件（window） | 触发方 | 行为 |
|---|---|---|
| `xingmang:theme-toggle` | 外部 / UI 插件 | 触发 `toggleTheme()`，等价点击顶栏主题按钮 |
| `xingmang:theme-set` | 外部 / UI 插件 | 设置指定主题，`detail: { theme: 'light' \| 'dark' }` |
| `xingmang:theme-changed` | 应用 | 主题变化后广播，`detail: { theme }`，供外部同步状态 |

> 主题状态源：`localStorage['main-theme']`；样式开关：`<html>` 的 `data-theme='dark'` 属性。UI 插件用法见 §2.5「UI 类插件」。
