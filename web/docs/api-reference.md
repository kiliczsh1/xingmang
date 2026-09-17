# 星芒 AI 写作 · 前端 API 参考（全项目）

> 版本：v1（2026-08-10）· 适用代码：`web/src/api/index.ts`（17 个模块 142 个方法）、`web/src/api/backup.ts`
> 插件系统专用接口见 [plugins-api.md](./plugins-api.md)（§7，26 个方法）

---

## 0. 全局约定

### axios 实例与响应包裹

- `baseURL: '/api'`，默认超时 `180000ms`（知识图谱分析 / AI 生成类接口单独传 `timeout` 放大）。
- **响应拦截器**：成功时返回 `response.data`（即 `ApiResponse<T>` 整体）；错误时自动 `ElMessage.error` 弹提示，调用方可通过请求配置 `{ silent: true }` 抑制并自行回退（如 AI 扫榜失败回退演示数据）。

```ts
interface ApiResponse<T> {
  success: boolean
  data: T
  message?: string
}
```

### 特殊约定

| 事项 | 说明 |
|---|---|
| `promptAPI` | `fields` / `subcategories` 后端以 JSON 字符串存储，API 层统一 `parseJsonArray` 解析为数组，消费方无需再处理 |
| `backupAPI` | 独立于 `index.ts`，用原始 axios（不经拦截器），导出走 `responseType: 'blob'` 触发浏览器下载 |
| `rankAPI` | 全部 `silent: true`，失败由调用方回退演示数据 |
| 分页参数 | `rankAPI` 用 `page/size`；`entryAPI` 用 `page/page_size`；`pluginAPI` 用 `page/pageSize`（各自保持历史命名） |

---

## 1. 页面 → API 模块映射

| 路由 / 页面 | 使用的 API 模块 |
|---|---|
| `/ai-ranking` AiRanking | `rankAPI` |
| `/analysis/:bookId` BookAnalysis | `chapterAPI` `configAPI` `promptAPI` |
| `/books` Books | `aiAPI` `chapterAPI` `configAPI` `promptAPI` |
| `/config` Config | `configAPI` `providerAPI` |
| `/creative` Creative | `bookAPI` `chapterAPI` `characterAPI` `configAPI` `generatorAPI` `memoAPI` `promptAPI` |
| `/experience-shares/:id` 详情 | `experienceShareAPI` |
| `/experience-shares` 列表 | `experienceShareAPI` |
| `/prompts/pack/:name` PackDetail | `promptAPI` |
| `/profile` Profile | `bookAPI` `generatorAPI` `promptAPI` `statsAPI` |
| `/prompt-preview` PromptPreview | `promptAPI` |
| `/prompts` Prompts | `promptAPI` |
| `/wiki-graph` WikiGraph | `bookAPI` `chapterAPI` `configAPI` `knowledgeGraphAPI` |
| `/workflow/:id` WorkflowWorkbench | `configAPI` `promptAPI` |
| `/write/:bookId` Write | `chapterAPI` `characterAPI` `configAPI` `conversationAPI` `memoAPI` `promptAPI` `volumeAPI` |
| `/plugins` 插件中心 | 本地管理（见 plugins-api.md）；接入后端后走 `pluginAPI` |
| 组件：ChatDialog / ChatMain | `bookAPI` `configAPI` `conversationAPI` / `promptAPI` |
| 组件：EntryLibrary / KnowledgeGraph | `entryAPI` / `knowledgeGraphAPI` |
| 布局：MainLayout | `configAPI` `statsAPI` |

> `backupAPI`（备份导入导出）当前仅在 `src/api/backup.ts` 定义，尚未接入页面。

---

## 2. 书本与章节

### `bookAPI`（5）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll | GET `/books` | — | `Book[]` |
| getOne | GET `/books/:id` | — | `Book` |
| create | POST `/books` | `Partial<Book>` | `Book` |
| update | PUT `/books/:id` | `Partial<Book>` | `Book` |
| delete | DELETE `/books/:id` | — | `ApiResponse` |

### `chapterAPI`（7）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getByBook | GET `/chapters/book/:bookId` | — | `Chapter[]` |
| getOne | GET `/chapters/:id` | — | `Chapter` |
| create | POST `/chapters` | `Partial<Chapter>` | `Chapter` |
| importBook | POST `/chapters/import-book` | `{ bookId, chapters[] }` | `{ insertedCount, chapters }` |
| importFile | POST `/chapters/import-file` | `{ bookId, file: { name,size,data_base64 } }` | `{ insertedCount, chapters }` |
| update | PUT `/chapters/:id` | `Partial<Chapter>` | `Chapter` |
| delete | DELETE `/chapters/:id` | — | `ApiResponse` |

### `volumeAPI`（5）· 分卷

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getByBook | GET `/volumes/book/:bookId` | — | `Volume[]` |
| getOne / create / update | GET/POST/PUT `/volumes(/:id)` | `Partial<Volume>` | `Volume` |
| delete | DELETE `/volumes/:id` | — | `ApiResponse` |

---

## 3. 提示词与卡包

### `promptAPI`（5）· 注意字段已自动解析为数组

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll | GET `/prompts` | — | `Prompt[]`（`fields`/`subcategories` 已解析） |
| getOne | GET `/prompts/:id` | — | `Prompt` |
| create | POST `/prompts` | `Partial<Prompt>` | `Prompt` |
| update | PUT `/prompts/:id` | `Partial<Prompt>` | `Prompt` |
| delete | DELETE `/prompts/:id` | — | `ApiResponse` |

---

## 4. 备忘录

### `memoAPI`（7）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll / getOne | GET `/memos(/:id)` | — | `Memo[]` / `Memo` |
| create / update | POST `/memos` / PUT `/memos/:id` | `Partial<Memo>` | `Memo` |
| delete | DELETE `/memos/:id` | — | `ApiResponse` |
| search | GET `/memos/search` | `keyword` / `category` | `Memo[]` |
| batch | POST `/memos/batch` | `{ action, ids[], data? }` | `ApiResponse` |

---

## 5. 经验分享

### `experienceShareAPI`（8）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll | GET `/experience-shares` | — | `ExperienceShare[]` |
| getOne / create / update / delete | 标准 CRUD `/experience-shares(/:id)` | — | `ExperienceShare` |
| importPdf | POST `/experience-shares/import-pdf` | `{ pdf_file: { name,size,data_base64 } }` | `Partial<ExperienceShare>` |
| exportAll | GET `/experience-shares/export` | — | `ExperienceShare[]` |
| importAll | POST `/experience-shares/import` | `{ cards[] }` | `ApiResponse` |

---

## 6. API 服务商与模型配置

### `providerAPI`（6）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll / getOne / create / update / delete | 标准 CRUD `/providers(/:id)` | — | `ApiProvider` |
| discoverModels | GET `/providers/:id/models` | — | `{ id, name, max_tokens? }[]` |

### `configAPI`（9）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll | GET `/config` | — | `ApiModel[]` |
| getDefault | GET `/config/default` | — | `ApiModel` |
| getByProvider | GET `/config/provider/:providerId` | — | `ApiModel[]` |
| create / update / delete | 标准 CRUD `/config(/:id)` | — | `ApiModel` |
| toggleModel | PUT `/config/:id/toggle` | — | `{ id, enabled }` |
| reorderModels | PUT `/config/reorder/all` | `{ ids[] }` | `ApiResponse` |
| testModel | POST `/config/:id/test` | — | `{ modelId, modelName, providerName?, latencyMs, preview }` |

---

## 7. AI 能力

### `aiAPI`（6）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| chat | POST `/ai/chat` | `{ messages, configId?, systemPrompts?, relatedContent? }` | `string` |
| generateDescription | POST `/ai/generate-description` | `{ title, promptId? }` | `string` |
| generateImage | POST `/ai/generate-image` | `{ messages, configId?, size?, quality?, style?, reference_images? }`（可传 AbortSignal） | `{ url }` |
| getRankReferenceCategories | GET `/ai/rank-reference/categories` | — | `{ key, label, categories[] }[]` |
| getRankReferenceBooks | GET `/ai/rank-reference/books` | `{ path }` | `{ rank, title, author, coverUrl, bookPath }[]` |
| recognizeCharacters | POST `/ai/recognize-characters` | `{ text, configId?, customPrompt? }` | `any[]` |

### `rankAPI`（2）· 番茄榜单（经 Go 后端转调 Python 桥）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getLists | GET `/ai/rank/lists` | —（`silent`） | `{ groups: RankGroup[] }` |
| getBooks | GET `/ai/rank/books` | `tab`（如 `1_2_1141`）/ `page` / `size`（`silent`） | `{ books: RankBook[] }` |

> `RankBook` / `RankGroup` 内联定义于 `src/api/index.ts`（仅被 API 消费的响应形状）。

---

## 8. 对话（创作侧聊天）

### `conversationAPI`（11）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getAll / list | GET `/conversations` | — | `any[]` |
| getByBook | GET `/conversations/book/:bookId` | — | `any[]` |
| create | POST `/conversations` | `{ book_id, title? }` | `any` |
| update | PUT `/conversations/:id` | `{ title }` | `any` |
| delete | DELETE `/conversations/:id` | — | `ApiResponse` |
| getMessages | GET `/conversations/:id/messages` | — | `any[]` |
| saveMessage | POST `/conversations/:id/messages` | `{ role, content, displayContent? }` | `any` |
| clearMessages | DELETE `/conversations/:id/messages` | — | `ApiResponse` |
| deleteMessage | DELETE `/conversations/:conversationId/messages/:messageId` | — | `ApiResponse` |
| updateMessage | PUT `/conversations/:conversationId/messages/:messageId` | `{ content, displayContent? }` | `any` |

---

## 9. 生成器、角色、知识图谱

### `generatorAPI`（5）· 变量抽卡生成器

| 方法 | 路径 | 响应 |
|---|---|---|
| getAll | GET `/generators` | `Generator[]` |
| getOne / create / update / delete | 标准 CRUD `/generators(/:id)` | `Generator` |

### `characterAPI`（8）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getByBook | GET `/characters/book/:bookId` | — | `Character[]` |
| getOne / create / update / delete | 标准 CRUD `/characters(/:id)` | — | `Character` |
| batchDelete | POST `/characters/batch-delete` | `{ ids[] }` | `ApiResponse` |
| batchMove | POST `/characters/batch-move` | `{ ids[], folder }` | `ApiResponse` |
| getFolders | GET `/characters/folders/:bookId` | — | `string[]` |

### `knowledgeGraphAPI`（19）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getGraph | GET `/knowledge-graph/graph/:bookId` | — | `KnowledgeGraphData` |
| getEntities | GET `/knowledge-graph/entities/book/:bookId` | — | `GraphEntity[]` |
| getEntity / createEntity / updateEntity / deleteEntity | 标准 CRUD `/knowledge-graph/entities(/:id)` | — | `GraphEntity` |
| getRelations | GET `/knowledge-graph/relations/book/:bookId` | — | `GraphRelation[]` |
| createRelation / updateRelation / deleteRelation | 标准 CRUD `/knowledge-graph/relations(/:id)` | — | `GraphRelation` |
| analyze | POST `/knowledge-graph/analyze` | `{ bookId, configId?, scope?, chapterIds?, versionId? }`（`timeout 300s`） | `KnowledgeGraphData` |
| aiSupplement | POST `/knowledge-graph/ai-supplement` | `{ bookId, configId, entityName, entityType, entityDescription, hint?, chapterIds? }`（`timeout 120s`） | `{ content }` |
| clearGraph | DELETE `/knowledge-graph/graph/:bookId` | — | `ApiResponse` |
| importGraph | POST `/knowledge-graph/graph/import` | `{ bookId, ...data }` | `KnowledgeGraphData` |
| getVersions / createVersion / getVersionData / renameVersion / deleteVersion | `/knowledge-graph/versions*` | — | `GraphVersion` 系列 |

---

## 10. 词条库（Entry）

### `entryAPI`（8）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getList | GET `/entries` | `book_id` / `category_type` / `keyword` / `page` / `page_size` | `EntryListResponse` |
| getOne / create / update / delete | 标准 CRUD `/entries(/:id)` | — | `Entry` |
| addToBook | POST `/entries/:entryId/add-to-book` | `{ bookId }` | `Entry` |
| getCategories | GET `/entries/categories` | `book_id?` | `EntryCategory[]` |
| aiAutoComplete | POST `/entries/:entryId/ai-auto-complete` | `{ configId }`（`timeout 120s`） | `Entry` |

---

## 11. 统计

### `statsAPI`（5）

| 方法 | 路径 | 请求 | 响应 |
|---|---|---|---|
| getOverview | GET `/stats/overview` | — | `UsageOverview` |
| getDaily | GET `/stats/daily` | `startDate` / `endDate` | `DailyUsage[]` |
| getModelStats | GET `/stats/model-stats` | — | `ModelStats[]` |
| getMonthly | GET `/stats/monthly` | — | `MonthlyStats[]` |
| record | POST `/stats/record` | `{ modelId, modelName, providerName, tokens? }` | `ApiResponse` |

---

## 12. 备份导入导出

### `backupAPI`（`src/api/backup.ts`，3+ 方法）· 原始 axios，不经统一拦截器

| 方法 | 路径 | 说明 |
|---|---|---|
| downloadBackup | GET `/backup/export?mask=0\|1` | `responseType: 'blob'`，解析 `Content-Disposition` 文件名并触发下载 |
| preview | POST `/backup/preview` | 传入备份 payload，返回 `BackupPreview`（版本/统计/模块清单） |
| importBackup | POST `/backup/import` | `{ payload, mode: 'overwrite'\|'merge', modules[] }`（`timeout 600s`），返回 `BackupImportResult` |

> 当前仅在 `backup.ts` 定义，尚未接入任何页面。

---

## 13. 插件系统

`pluginAPI`（26 个方法，7 组）已落地 `src/api/index.ts`，完整端点表见 [plugins-api.md §8](./plugins-api.md#8-后端-rest-api-契约)。插件运行时契约、本地 zip 安装见同一文档。

---

## 附：调用约定速查

```ts
import { bookAPI, configAPI, statsAPI } from '@/api'

const res = await bookAPI.getAll()          // res: ApiResponse<Book[]>
if (res.success) { /* res.data */ }

await rankAPI.getBooks({ tab: '1_2_1141' }) // silent: 失败自行回退
await configAPI.testModel(3)                // 长时间接口, 已有 timeout
```
