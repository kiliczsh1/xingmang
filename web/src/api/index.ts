import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse, Book, Chapter, Prompt, Memo, ApiProvider, ApiModel, ChatMessage, RelatedContent, Generator, Character, UsageOverview, DailyUsage, ModelStats, MonthlyStats, Volume, ExperienceShare, GraphEntity, GraphRelation, GraphChapterRef, KnowledgeGraphData, GraphVersion, Entry, EntryCategory, EntryListResponse, EntityMergeCandidate, PluginManifest } from '@/types'

// 扩充 axios 配置：silent 用于抑制全局错误弹窗（见响应拦截器）
declare module 'axios' {
  export interface AxiosRequestConfig {
    silent?: boolean
  }
}

const api = axios.create({
  baseURL: '/api',
  timeout: 180000
})

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    // 调用方可传 { silent: true } 抑制全局错误弹窗，自行处理回退（如 AI 扫榜实时接口失败回退演示数据）
    if (error.config?.silent) {
      return Promise.reject(error)
    }
    const message = error.response?.data?.message || error.message || '请求失败'
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// 书本相关API
export const bookAPI = {
  getAll: () => api.get<any, ApiResponse<Book[]>>('/books'),
  getOne: (id: number) => api.get<any, ApiResponse<Book>>(`/books/${id}`),
  create: (data: Partial<Book>) => api.post<any, ApiResponse<Book>>('/books', data),
  update: (id: number, data: Partial<Book>) => api.put<any, ApiResponse<Book>>(`/books/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/books/${id}`)
}

// 章节相关API
export const chapterAPI = {
  getByBook: (bookId: number) => api.get<any, ApiResponse<Chapter[]>>(`/chapters/book/${bookId}`),
  getOne: (id: number) => api.get<any, ApiResponse<Chapter>>(`/chapters/${id}`),
  create: (data: Partial<Chapter>) => api.post<any, ApiResponse<Chapter>>('/chapters', data),
  importBook: (data: {
    bookId: number
    chapters: Array<Pick<Chapter, 'title' | 'content'>>
  }) => api.post<any, ApiResponse<{ insertedCount: number; chapters: Chapter[] }>>('/chapters/import-book', data),
  importFile: (data: {
    bookId: number
    file: { name: string; size: number; data_base64: string }
  }) => api.post<any, ApiResponse<{ insertedCount: number; chapters: Chapter[] }>>('/chapters/import-file', data),
  update: (id: number, data: Partial<Chapter>) => api.put<any, ApiResponse<Chapter>>(`/chapters/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/chapters/${id}`)
}

// 提示词相关API
// 后端将 fields / subcategories 以 JSON 字符串存储并原样返回（如 "[]"），
// 前端类型按数组使用，若直接遍历字符串会出现 "[", "]" 之类的垃圾标签。
// 因此在 API 层统一把字符串解析为数组，所有消费方（Prompts/PackDetail/PromptPreview/Write/Creative 等）一起修复。
const parseJsonArray = (value: unknown): any[] => {
  if (Array.isArray(value)) return value
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

// 解析 metadata（后端以 JSON 字符串存储，如 "{\"work_type\":\"novel\",...}"）
const parseJsonObject = (value: unknown): Record<string, any> => {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value as Record<string, any>
  if (typeof value === 'string') {
    try {
      const parsed = JSON.parse(value)
      return parsed && typeof parsed === 'object' ? parsed : {}
    } catch {
      return {}
    }
  }
  return {}
}

// 图谱 metadata 同样是 JSON 字符串（后端 GraphEntity.Metadata 为 string 列），
// 在 API 层统一解析，避免各消费方重复判断字符串/对象。
const normalizeGraphEntity = (entity: GraphEntity) => {
  if (entity) entity.metadata = parseJsonObject(entity.metadata)
  return entity
}

// 后端空数据时 entities/relations 可能为 null，这里统一归一化并顺带解析 metadata
const normalizeGraphData = (data: Partial<KnowledgeGraphData> | null | undefined): KnowledgeGraphData => {
  const entities = Array.isArray(data?.entities) ? data.entities.map(normalizeGraphEntity) : []
  const relations = Array.isArray(data?.relations) ? data.relations : []
  return { ...data, entities, relations }
}

const normalizePrompt = (prompt: Prompt) => {
  if (prompt) {
    prompt.fields = parseJsonArray(prompt.fields)
    prompt.subcategories = parseJsonArray(prompt.subcategories)
    prompt.metadata = parseJsonObject(prompt.metadata)
  }
  return prompt
}

export const promptAPI = {
  getAll: async () => {
    const res = await api.get<any, ApiResponse<Prompt[]>>('/prompts')
    if (res.success && Array.isArray(res.data)) {
      res.data.forEach(normalizePrompt)
    }
    return res
  },
  getOne: async (id: number) => {
    const res = await api.get<any, ApiResponse<Prompt>>(`/prompts/${id}`)
    if (res.success && res.data) normalizePrompt(res.data)
    return res
  },
  create: async (data: Partial<Prompt>) => {
    const res = await api.post<any, ApiResponse<Prompt>>('/prompts', data)
    if (res.success && res.data) normalizePrompt(res.data)
    return res
  },
  update: async (id: number, data: Partial<Prompt>) => {
    const res = await api.put<any, ApiResponse<Prompt>>(`/prompts/${id}`, data)
    if (res.success && res.data) normalizePrompt(res.data)
    return res
  },
  delete: (id: number) => api.delete<any, ApiResponse>(`/prompts/${id}`)
}

// 备忘录相关 API
export const memoAPI = {
  getAll: () => api.get<any, ApiResponse<Memo[]>>('/memos'),
  getOne: (id: number) => api.get<any, ApiResponse<Memo>>(`/memos/${id}`),
  create: (data: Partial<Memo>) => api.post<any, ApiResponse<Memo>>('/memos', data),
  update: (id: number, data: Partial<Memo>) => api.put<any, ApiResponse<Memo>>(`/memos/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/memos/${id}`),
  search: (params?: { keyword?: string; category?: string }) => 
    api.get<any, ApiResponse<Memo[]>>('/memos/search', { params }),
  batch: (data: { action: string; ids: number[]; data?: any }) => 
    api.post<any, ApiResponse>('/memos/batch', data)
}

export const experienceShareAPI = {
  getAll: () => api.get<any, ApiResponse<ExperienceShare[]>>('/experience-shares'),
  getOne: (id: number) => api.get<any, ApiResponse<ExperienceShare>>(`/experience-shares/${id}`),
  create: (data: any) => api.post<any, ApiResponse<ExperienceShare>>('/experience-shares', data),
  update: (id: number, data: any) => api.put<any, ApiResponse<ExperienceShare>>(`/experience-shares/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/experience-shares/${id}`),
  importPdf: (data: {
    pdf_file: {
      name: string
      size: number
      data_base64: string
    }
  }) => api.post<any, ApiResponse<Partial<ExperienceShare>>>('/experience-shares/import-pdf', data),
  exportAll: () => api.get<any, ApiResponse<ExperienceShare[]>>('/experience-shares/export'),
  importAll: (data: { cards: any[] }) => api.post<any, ApiResponse>('/experience-shares/import', data)
}

// API服务商相关API
export const providerAPI = {
  getAll: () => api.get<any, ApiResponse<ApiProvider[]>>('/providers'),
  getOne: (id: number) => api.get<any, ApiResponse<ApiProvider>>(`/providers/${id}`),
  create: (data: Partial<ApiProvider>) => api.post<any, ApiResponse<ApiProvider>>('/providers', data),
  update: (id: number, data: Partial<ApiProvider>) => api.put<any, ApiResponse<ApiProvider>>(`/providers/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/providers/${id}`),
  discoverModels: (id: number) => api.get<any, ApiResponse<Array<{ id: string; name: string; max_tokens?: number | null }>>>(`/providers/${id}/models`)
}

// API模型配置相关API
export const configAPI = {
  getAll: () => api.get<any, ApiResponse<ApiModel[]>>('/config'),
  getDefault: () => api.get<any, ApiResponse<ApiModel>>('/config/default'),
  getByProvider: (providerId: number) => api.get<any, ApiResponse<ApiModel[]>>(`/config/provider/${providerId}`),
  create: (data: Partial<ApiModel>) => api.post<any, ApiResponse<ApiModel>>('/config', data),
  update: (id: number, data: Partial<ApiModel>) => api.put<any, ApiResponse<ApiModel>>(`/config/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/config/${id}`),
  toggleModel: (id: number) => api.put<any, ApiResponse<{ id: number; enabled: number }>>(`/config/${id}/toggle`),
  reorderModels: (ids: number[]) => api.put<any, ApiResponse>('/config/reorder/all', { ids }),
  testModel: (id: number) => api.post<any, ApiResponse<{
    modelId: number
    modelName: string
    providerName?: string
    latencyMs: number
    preview: string
  }>>(`/config/${id}/test`)
}

// AI相关API
export const aiAPI = {
  chat: (data: {
    messages: ChatMessage[]
    configId?: number
    systemPrompts?: string[]
    relatedContent?: RelatedContent[]
    maxTokens?: number
  }) => api.post<any, ApiResponse<string>>('/ai/chat', data),
  generateDescription: (data: { title: string; promptId?: number }) => 
    api.post<any, ApiResponse<string>>('/ai/generate-description', data),
  generateImage: (
    data: {
      messages: ChatMessage[]
      configId?: number
      size?: string
      quality?: string
      style?: string
      reference_images?: Array<{
        name: string
        size: number
        mime_type: string
        data_base64: string
        source_url?: string
      }>
    },
    options?: { signal?: AbortSignal }
  ) => api.post<any, ApiResponse<{ url: string }>>('/ai/generate-image', data, options),
  getRankReferenceCategories: () =>
    api.get<any, ApiResponse<Array<{
      key: string
      label: string
      categories: Array<{
        id: string
        label: string
        path: string
        fullUrl: string
      }>
    }>>>('/ai/rank-reference/categories'),
  getRankReferenceBooks: (path: string) =>
    api.get<any, ApiResponse<Array<{
      rank: number
      title: string
      author: string
      coverUrl: string
      bookPath: string
    }>>>('/ai/rank-reference/books', { params: { path } }),
  recognizeCharacters: (data: { text: string; configId?: number; customPrompt?: string }) =>
    api.post<any, ApiResponse<any[]>>('/ai/recognize-characters', data)
}

// AI 扫榜相关 API（番茄榜单数据，经 Go 后端转调 Python 桥）
export const rankAPI = {
  getLists: () =>
    api.get<any, ApiResponse<{ groups: RankGroup[] }>>('/ai/rank/lists', { silent: true }),
  // tab 格式：gender_rankMold_categoryId，如 "1_2_1141"（男频·阅读榜·西方奇幻）
  getBooks: (params: { tab: string; page?: number; size?: number }) =>
    api.get<any, ApiResponse<{ books: RankBook[] }>>('/ai/rank/books', { params, silent: true })
}

// AI 扫榜数据结构（与后端 internal/rank 的 JSON 一致）
export interface RankBook {
  rank: number
  title: string
  author: string
  category: string
  status: string
  readingCount: string
  wordCount: string
  coverUrl: string
  bookId: string
  link: string
  rankDiff: number
  synopsis: string
}

// 榜单分组：一组 = 一个榜单（如男频阅读榜），items = 该榜题材列表
export interface RankGroup {
  title: string
  items: Array<{ id: string; name: string }>
}

// 对话相关 API
export const conversationAPI = {
  getAll: () => api.get<any, ApiResponse<any[]>>('/conversations'),
  list: () => api.get<any, ApiResponse<any[]>>('/conversations'),
  getByBook: (bookId: number) => api.get<any, ApiResponse<any[]>>(`/conversations/book/${bookId}`),
  create: (data: { book_id: number; title?: string }) => api.post<any, ApiResponse<any>>('/conversations', data),
  update: (id: number, data: { title: string }) => api.put<any, ApiResponse<any>>(`/conversations/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/conversations/${id}`),
  getMessages: (id: number) => api.get<any, ApiResponse<any[]>>(`/conversations/${id}/messages`),
  saveMessage: (id: number, data: { role: string; content: string; displayContent?: string }) =>
    api.post<any, ApiResponse<any>>(`/conversations/${id}/messages`, data),
  clearMessages: (id: number) => api.delete<any, ApiResponse>(`/conversations/${id}/messages`),
  deleteMessage: (conversationId: number, messageId: number) => 
    api.delete<any, ApiResponse>(`/conversations/${conversationId}/messages/${messageId}`),
  updateMessage: (conversationId: number, messageId: number, data: { content: string; displayContent?: string }) => 
    api.put<any, ApiResponse<any>>(`/conversations/${conversationId}/messages/${messageId}`, data)
}

// 生成器相关API
export const generatorAPI = {
  getAll: () => api.get<any, ApiResponse<Generator[]>>('/generators'),
  getOne: (id: number) => api.get<any, ApiResponse<Generator>>(`/generators/${id}`),
  create: (data: Partial<Generator>) => api.post<any, ApiResponse<Generator>>('/generators', data),
  update: (id: number, data: Partial<Generator>) => api.put<any, ApiResponse<Generator>>(`/generators/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/generators/${id}`)
}

// 角色相关API
export const characterAPI = {
  getByBook: (bookId: number) => api.get<any, ApiResponse<Character[]>>(`/characters/book/${bookId}`),
  getOne: (id: number) => api.get<any, ApiResponse<Character>>(`/characters/${id}`),
  create: (data: Partial<Character>) => api.post<any, ApiResponse<Character>>('/characters', data),
  update: (id: number, data: Partial<Character>) => api.put<any, ApiResponse<Character>>(`/characters/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/characters/${id}`),
  batchDelete: (ids: number[]) => api.post<any, ApiResponse>('/characters/batch-delete', { ids }),
  batchMove: (ids: number[], folder: string) => api.post<any, ApiResponse>('/characters/batch-move', { ids, folder }),
  getFolders: (bookId: number) => api.get<any, ApiResponse<string[]>>(`/characters/folders/${bookId}`)
}

export const statsAPI = {
  getOverview: () => api.get<any, ApiResponse<UsageOverview>>('/stats/overview'),
  getDaily: (startDate?: string, endDate?: string) => 
    api.get<any, ApiResponse<DailyUsage[]>>('/stats/daily', { params: { startDate, endDate } }),
  getModelStats: () => api.get<any, ApiResponse<ModelStats[]>>('/stats/model-stats'),
  getMonthly: () => api.get<any, ApiResponse<MonthlyStats[]>>('/stats/monthly'),
  record: (data: { modelId: number; modelName: string; providerName: string; tokens?: number }) => 
    api.post<any, ApiResponse>('/stats/record', data)
}

// 分卷相关 API
export const volumeAPI = {
  getByBook: (bookId: number) => api.get<any, ApiResponse<Volume[]>>(`/volumes/book/${bookId}`),
  getOne: (id: number) => api.get<any, ApiResponse<Volume>>(`/volumes/${id}`),
  create: (data: Partial<Volume>) => api.post<any, ApiResponse<Volume>>('/volumes', data),
  update: (id: number, data: Partial<Volume>) => api.put<any, ApiResponse<Volume>>(`/volumes/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/volumes/${id}`)
}

export const knowledgeGraphAPI = {
  // 不带 versionId 时返回该书全部实体/关系（后端兼容行为）
  getGraph: async (bookId: number, versionId?: number) => {
    const res = await api.get<any, ApiResponse<KnowledgeGraphData>>(`/knowledge-graph/graph/${bookId}`, {
      params: versionId ? { versionId } : undefined
    })
    if (res.success && res.data) res.data = normalizeGraphData(res.data)
    return res
  },
  getEntities: async (bookId: number) => {
    const res = await api.get<any, ApiResponse<GraphEntity[]>>(`/knowledge-graph/entities/book/${bookId}`)
    if (res.success && Array.isArray(res.data)) res.data.forEach(normalizeGraphEntity)
    return res
  },
  getEntity: async (id: number) => {
    const res = await api.get<any, ApiResponse<GraphEntity>>(`/knowledge-graph/entities/${id}`)
    if (res.success && res.data) normalizeGraphEntity(res.data)
    return res
  },
  createEntity: (data: Partial<GraphEntity>) => api.post<any, ApiResponse<GraphEntity>>('/knowledge-graph/entities', data),
  updateEntity: (id: number, data: Partial<GraphEntity>) => api.put<any, ApiResponse<GraphEntity>>(`/knowledge-graph/entities/${id}`, data),
  deleteEntity: (id: number) => api.delete<any, ApiResponse>(`/knowledge-graph/entities/${id}`),
  getRelations: (bookId: number) => api.get<any, ApiResponse<GraphRelation[]>>(`/knowledge-graph/relations/book/${bookId}`),
  createRelation: (data: Partial<GraphRelation>) => api.post<any, ApiResponse<GraphRelation>>('/knowledge-graph/relations', data),
  updateRelation: (id: number, data: Partial<GraphRelation>) => api.put<any, ApiResponse<GraphRelation>>(`/knowledge-graph/relations/${id}`, data),
  deleteRelation: (id: number) => api.delete<any, ApiResponse>(`/knowledge-graph/relations/${id}`),
  // 章节关联：按章节分组的实体/关系 id，用于左栏覆盖标记与按章节过滤
  getChapterRefs: (bookId: number, versionId?: number) =>
    api.get<any, ApiResponse<GraphChapterRef[]>>(`/knowledge-graph/chapter-refs/${bookId}`, {
      params: versionId ? { versionId } : undefined
    }),
  setEntityChapters: (id: number, chapterIds: number[]) =>
    api.put<any, ApiResponse<{ chapterRefCount: number }>>(`/knowledge-graph/entities/${id}/chapters`, { chapterIds }),
  setRelationChapters: (id: number, chapterIds: number[]) =>
    api.put<any, ApiResponse<{ chapterRefCount: number }>>(`/knowledge-graph/relations/${id}/chapters`, { chapterIds }),
  // ── 实体合并系统（主词条 / 副词条）─────────────────────────
  /** 把 memberIds 全部并入 primaryId：它们变成副词条，不出节点，关系端点由前端折到主词条上 */
  mergeEntities: (data: { bookId: number; versionId?: number; primaryId: number; memberIds: number[] }) =>
    api.post<any, ApiResponse<{ primaryId: number; merged: number; reparented: number; skipped: number }>>(
      '/knowledge-graph/entities/merge',
      data
    ),
  /** 解除合并：传主词条则解散整组，传副词条则只解绑它自己 */
  unmergeEntities: (entityIds: number[]) =>
    api.post<any, ApiResponse<{ released: number }>>('/knowledge-graph/entities/unmerge', { entityIds }),
  /** 疑似重复实体候选对，供人工合并检测确认 */
  getMergeCandidates: (bookId: number, versionId?: number) =>
    api.get<any, ApiResponse<EntityMergeCandidate[]>>(`/knowledge-graph/merge-candidates/${bookId}`, {
      params: versionId ? { versionId } : undefined
    }),
  analyze: (data: { bookId: number; configId?: number; scope?: 'all' | 'chapters'; chapterIds?: number[]; versionId?: number }) =>
    api.post<any, ApiResponse<KnowledgeGraphData>>('/knowledge-graph/analyze', data, { timeout: 300000 }),
  aiSupplement: (data: { bookId: number; configId: number; entityName: string; entityType: string; entityDescription: string; hint?: string; chapterIds?: number[] }) =>
    api.post<any, ApiResponse<{ content: string }>>('/knowledge-graph/ai-supplement', data, { timeout: 120000 }),
  clearGraph: (bookId: number) => api.delete<any, ApiResponse>(`/knowledge-graph/graph/${bookId}`),
  // 带 versionId 时整批写入该版本，后端会同时回写该版本的实体/关系计数
  importGraph: (bookId: number, data: { entities: GraphEntity[]; relations: GraphRelation[]; versionId?: number }) =>
    api.post<any, ApiResponse<{ entityCount: number; relationCount: number }>>('/knowledge-graph/graph/import', { bookId, ...data }),
  getVersions: (bookId: number) => api.get<any, ApiResponse<GraphVersion[]>>(`/knowledge-graph/versions/${bookId}`),
  createVersion: (bookId: number, name?: string) => api.post<any, ApiResponse<GraphVersion>>('/knowledge-graph/versions', { bookId, name }),
  getVersionData: async (versionId: number) => {
    const res = await api.get<any, ApiResponse<KnowledgeGraphData>>(`/knowledge-graph/version/${versionId}`)
    if (res.success && res.data) res.data = normalizeGraphData(res.data)
    return res
  },
  renameVersion: (versionId: number, name: string) => api.put<any, ApiResponse<GraphVersion>>(`/knowledge-graph/versions/${versionId}`, { name }),
  deleteVersion: (bookId: number, versionId: number) => api.delete<any, ApiResponse>(`/knowledge-graph/versions/${bookId}/${versionId}`)
}

export const entryAPI = {
  getList: (params?: { book_id: number; category_type?: string; keyword?: string; page?: number; page_size?: number }) =>
    api.get<any, ApiResponse<EntryListResponse>>('/entries', { params }),
  getOne: (id: number) => api.get<any, ApiResponse<Entry>>(`/entries/${id}`),
  create: (data: Partial<Entry>) => api.post<any, ApiResponse<Entry>>('/entries', data),
  update: (id: number, data: Partial<Entry>) => api.put<any, ApiResponse<Entry>>(`/entries/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/entries/${id}`),
  addToBook: (entryId: number, bookId: number) => api.post<any, ApiResponse<Entry>>(`/entries/${entryId}/add-to-book`, { bookId }),
  getCategories: (bookId?: number) => api.get<any, ApiResponse<EntryCategory[]>>('/entries/categories', { params: { book_id: bookId } }),
  aiAutoComplete: (entryId: number, configId: number) => api.post<any, ApiResponse<Entry>>(`/entries/${entryId}/ai-auto-complete`, { configId }, { timeout: 120000 })
}

// ===================================================================
// 插件系统 API（预留 · 后端待实现）
// 覆盖插件全生命周期：市场发现 → 安装 → 管理 → 更新 → 依赖冲突 →
// 配置 → 诊断 → 评价统计。当前前端仅内置插件本地管理（Plugins.vue），
// 接入后端后按此契约实现即可无缝替换第三方空态区。
// ===================================================================

// 插件分类（市场筛选用）
export interface PluginCategory {
  key: string
  label: string
  count: number
}

// 市场条目：比 PluginManifest 多市场侧字段
export interface PluginMarketItem extends PluginManifest {
  installCount: number
  rating: number
  reviewCount: number
  updatedAt: string
  tags: string[]
}

// 可更新项
export interface PluginUpdateInfo {
  pluginId: string
  currentVersion: string
  latestVersion: string
  size: string
  changelog: string[]
}

// 依赖项
export interface PluginDependency {
  name: string
  version: string
  satisfied: boolean
  status: 'ok' | 'missing' | 'conflict'
  conflictWith?: string
}

// 检测报告项
export interface PluginDetectItem {
  pluginId: string
  status: 'ok' | 'warn' | 'err'
  detail: string
}

// 诊断总览
export interface PluginDiagnostics {
  healthScore: number
  total: number
  running: number
  errorCount: number
  warnCount: number
}

// 诊断日志
export interface PluginLog {
  level: 'err' | 'warn' | 'info' | 'ok'
  message: string
  pluginId?: string
  time: string
  stack?: string
}

// 评价
export interface PluginReview {
  id: number
  author: string
  rating: number
  content: string
  createdAt: string
}

export const pluginAPI = {
  // ---- A. 市场发现 ----
  getMarketplace: (params?: { category?: string; keyword?: string; tag?: string; sort?: 'hot' | 'new' | 'rating'; page?: number; pageSize?: number }) =>
    api.get<any, ApiResponse<{ list: PluginMarketItem[]; total: number }>>('/plugins/marketplace', { params }),
  getMarketCategories: () => api.get<any, ApiResponse<PluginCategory[]>>('/plugins/marketplace/categories'),
  getMarketTags: () => api.get<any, ApiResponse<Array<{ name: string; count: number }>>>('/plugins/marketplace/tags'),
  getRecommended: () => api.get<any, ApiResponse<PluginMarketItem[]>>('/plugins/marketplace/recommended'),
  getMarketItem: (id: string) => api.get<any, ApiResponse<PluginMarketItem>>(`/plugins/marketplace/${id}`),
  getReviews: (id: string, params?: { page?: number }) =>
    api.get<any, ApiResponse<{ list: PluginReview[]; total: number }>>(`/plugins/marketplace/${id}/reviews`, { params }),
  addReview: (id: string, data: { rating: number; content: string }) =>
    api.post<any, ApiResponse<PluginReview>>(`/plugins/marketplace/${id}/reviews`, data),

  // ---- B. 已安装管理 ----
  getInstalled: () => api.get<any, ApiResponse<PluginManifest[]>>('/plugins'),
  getOne: (id: string) => api.get<any, ApiResponse<PluginManifest>>(`/plugins/${id}`),
  install: (data: { name: string; version?: string }) => api.post<any, ApiResponse<PluginManifest>>('/plugins', data),
  uninstall: (id: string) => api.delete<any, ApiResponse>(`/plugins/${id}`),
  toggle: (id: string, enabled: boolean) => api.put<any, ApiResponse<PluginManifest>>(`/plugins/${id}/toggle`, { enabled }),
  batch: (data: { action: 'enable' | 'disable' | 'uninstall'; ids: string[] }) =>
    api.post<any, ApiResponse>('/plugins/batch', data),
  detect: () => api.get<any, ApiResponse<{ report: PluginDetectItem[] }>>('/plugins/detect'),

  // ---- C. 更新 ----
  getUpdates: () => api.get<any, ApiResponse<PluginUpdateInfo[]>>('/plugins/updates'),
  updateOne: (id: string) => api.put<any, ApiResponse<PluginManifest>>(`/plugins/${id}/update`),
  updateAll: () => api.put<any, ApiResponse<{ updated: number }>>('/plugins/updates'),
  getChangelog: (id: string) => api.get<any, ApiResponse<Array<{ version: string; date: string; notes: string[] }>>>(`/plugins/${id}/changelog`),

  // ---- D. 依赖与冲突 ----
  getDependencies: (id: string) => api.get<any, ApiResponse<PluginDependency[]>>(`/plugins/${id}/dependencies`),
  resolveConflicts: (id: string, data?: { strategy?: 'auto' | 'manual'; choices?: Record<string, string> }) =>
    api.post<any, ApiResponse<{ fixed: string[]; ignored: string[] }>>(`/plugins/${id}/dependencies/resolve`, data),

  // ---- E. 配置 ----
  getSettings: (id: string) => api.get<any, ApiResponse<{ schema: any; values: Record<string, any> }>>(`/plugins/${id}/settings`),
  saveSettings: (id: string, data: Record<string, any>) => api.put<any, ApiResponse>(`/plugins/${id}/settings`, data),

  // ---- F. 诊断 ----
  getDiagnostics: () => api.get<any, ApiResponse<PluginDiagnostics>>('/plugins/diagnostics'),
  getLogs: (params?: { level?: 'all' | 'err' | 'warn' | 'info'; limit?: number }) =>
    api.get<any, ApiResponse<PluginLog[]>>('/plugins/logs', { params }),
  getLogDetail: (logId: string) => api.get<any, ApiResponse<PluginLog>>(`/plugins/logs/${logId}`),

  // ---- G. 评价与统计 ----
  getStats: () => api.get<any, ApiResponse<{ installs: number; activeCount: number; topPlugins: Array<{ name: string; installs: number }> }>>('/plugins/stats')
}

export default api
