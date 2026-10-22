import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse, Book, Chapter, Prompt, PromptInsight, Memo, ApiProvider, ApiModel, ChatMessage, RelatedContent, Generator, Character, UsageOverview, DailyUsage, ModelStats, MonthlyStats, Volume, ExperienceShare, ExperienceSection, ExperienceExportPayload, ExperienceImportResult, BoardKind, GraphEntity, GraphRelation, GraphChapterRef, KnowledgeGraphData, GraphVersion, Entry, EntryCategory, EntryListResponse, EntityMergeCandidate, PluginManifest } from '@/types'

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
  // mode：archive=归档到 _归档（默认），trash=移入系统回收站
  delete: (id: number, mode: 'archive' | 'trash' = 'archive') =>
    api.delete<any, ApiResponse>(`/books/${id}`, { params: { mode } }),
  reveal: (id: number) => api.post<any, ApiResponse>(`/books/${id}/reveal`),
  // 导入已有目录（服务端路径）；导出走浏览器直接下载
  importFolder: (path: string) => api.post<any, ApiResponse<Book>>('/books/import-folder', { path }),
  // 旧版本导出的作品 JSON：原样把文件内容作为请求体
  importBookJson: (raw: string, title: string) =>
    api.post<any, ApiResponse<{ book: Book; chapters: number }>>('/books/import-json', raw, {
      params: { title },
      headers: { 'Content-Type': 'application/json' }
    }),
  exportUrl: (id: number) => `/api/books/${id}/export`,
  gallery: (id: number) =>
    api.get<any, ApiResponse<Array<{ name: string; mtime: number; current: boolean }>>>(`/books/${id}/gallery`),
  activateCover: (id: number, name: string) =>
    api.post<any, ApiResponse>(`/books/${id}/gallery/activate`, { name }),
  coverUrl: (id: number, v?: number) => `/api/books/${id}/cover${v ? `?v=${v}` : ''}`,
  galleryUrl: (id: number, name: string) => `/api/books/${id}/gallery/${encodeURIComponent(name)}`,
  addGallery: (id: number, image: string) => api.post<any, ApiResponse<{ name: string }>>(`/books/${id}/gallery`, { image }),
  deleteGallery: (id: number, name: string) =>
    api.delete<any, ApiResponse>(`/books/${id}/gallery/${encodeURIComponent(name)}`)
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
    // 锁定标记统一为 0/1，兼容后端返回 0 / 1 / true / false
    prompt.is_locked = prompt.is_locked ? 1 : 0
    // 心得聚合为只读字段，统一成数字，避免 null 参与计算
    prompt.insight_count = Number(prompt.insight_count) || 0
    prompt.insight_avg = Number(prompt.insight_avg) || 0
  }
  return prompt
}

// 心得的 tags 同样是 JSON 字符串（如 "[\"长篇\"]"），在 API 层统一解析成数组
const normalizeInsight = (item: PromptInsight) => {
  if (item) {
    item.tags = parseJsonArray(item.tags).map((t) => String(t))
    item.rating = Number(item.rating) || 0
  }
  return item
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

// 卡片使用心得 API（一张卡片多条，时间线）
// ⚠ 更新/删除走 /prompt-insights/{id} 而不是 /prompts/insights/{id}：
// 后者会和 /prompts/{id}/insights 在 Go 1.22 的 ServeMux 里注册冲突（直接 panic）。
export const promptInsightAPI = {
  /** 全量（扁平，带 prompt_id），供导出等批量场景 */
  getAll: async () => {
    const res = await api.get<any, ApiResponse<PromptInsight[]>>('/prompt-insights')
    if (res.success && Array.isArray(res.data)) res.data.forEach(normalizeInsight)
    return res
  },
  getForPrompt: async (promptId: number) => {
    const res = await api.get<any, ApiResponse<PromptInsight[]>>(`/prompts/${promptId}/insights`)
    if (res.success && Array.isArray(res.data)) res.data.forEach(normalizeInsight)
    return res
  },
  create: async (promptId: number, data: { content: string; rating?: number; tags?: string[] }) => {
    const res = await api.post<any, ApiResponse<PromptInsight>>(`/prompts/${promptId}/insights`, data)
    if (res.success && res.data) normalizeInsight(res.data)
    return res
  },
  update: async (id: number, data: { content?: string; rating?: number; tags?: string[] }) => {
    const res = await api.put<any, ApiResponse<PromptInsight>>(`/prompt-insights/${id}`, data)
    if (res.success && res.data) normalizeInsight(res.data)
    return res
  },
  delete: (id: number) => api.delete<any, ApiResponse>(`/prompt-insights/${id}`)
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
  /** board 缺省 experience（后端口径），老调用不带参仍兼容 */
  getAll: (params?: { board?: BoardKind }) =>
    api.get<any, ApiResponse<ExperienceShare[]>>('/experience-shares', { params }),
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
  /**
   * 导出全部帖子（两个吧全量），返回 export_version 2 的完整对象。
   * ⚠ 附件文件本体（封面图 / PDF）不在 JSON 里，别在界面上把它说成「备份」。
   */
  exportAll: () => api.get<any, ExperienceExportPayload>('/experience-shares/export'),
  /**
   * 导入帖子。请求体可以是裸数组，也可以是 cards / shares / data 任意外壳 ——
   * 解析交给后端（它认四种形态），前端只负责把 JSON.parse 的结果原样传过去。
   */
  importAll: (payload: unknown) =>
    api.post<any, ApiResponse<ExperienceImportResult>>('/experience-shares/import', payload, { silent: true }),
  /** 主楼点赞 / 取消赞 */
  like: (id: number, delta: 1 | -1) =>
    api.post<any, ApiResponse<{ id: number; like_count: number }>>(`/experience-shares/${id}/like`, { delta }),
  /** 加精 / 取消加精 */
  setFeatured: (id: number, is_featured: 0 | 1) =>
    api.post<any, ApiResponse<{ id: number; is_featured: number }>>(`/experience-shares/${id}/feature`, { is_featured }),
  /** 置顶 / 取消置顶（每个吧最多 3 条，超了后端返回 400） */
  setPinned: (id: number, is_pinned: 0 | 1) =>
    api.post<any, ApiResponse<{ id: number; is_pinned: number }>>(`/experience-shares/${id}/pin`, { is_pinned }, { silent: true }),
  /** 移动到分区（section_id = null 表示移出到未归类） */
  setSection: (id: number, section_id: number | null) =>
    api.post<any, ApiResponse<{ id: number; section_id: number | null }>>(`/experience-shares/${id}/section`, { section_id }),
  /** 批量移动分区 */
  batchSetSection: (ids: number[], section_id: number | null) =>
    api.post<any, ApiResponse<{ affected: number }>>('/experience-shares/batch-section', { ids, section_id })
}

/** 正文插图上传（base64 → 后端落盘 data/editor-images/，返回站内 URL） */
export const imageAPI = {
  upload: (payload: { name: string; mime?: string; data_base64: string }) =>
    api.post<any, ApiResponse<{ url: string; name: string }>>('/images', payload, { silent: true })
}

export const experienceSectionAPI = {
  /** board 缺省 experience（后端口径）；分区按吧隔离，两个吧各自最多 15 个、可同名 */
  getAll: (params?: { board?: BoardKind }) =>
    api.get<any, ApiResponse<ExperienceSection[]>>('/experience-sections', { params }),
  create: (name: string, board?: BoardKind) =>
    api.post<any, ApiResponse<ExperienceSection>>('/experience-sections', { name, board }),
  rename: (id: number, name: string) => api.put<any, ApiResponse<ExperienceSection>>(`/experience-sections/${id}`, { name }),
  /** 整体重排：传最终顺序的 id 数组 */
  reorder: (ids: number[]) => api.post<any, ApiResponse<{ affected: number }>>('/experience-sections/reorder', { ids }),
  remove: (id: number) => api.delete<any, ApiResponse>(`/experience-sections/${id}`)
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
  }>>(`/config/${id}/test`),
  /**
   * 单模型探测：连通 + 递减探测「可填的 max_tokens」。
   *
   * 支持尚未落库的模型（导入弹窗里 provider_id + model），所以不挂在 /config/{id} 下。
   * `accepted` 的含义是「这个值没被拒」，**不是**「模型能生成这么多」。
   * silent：失败要在表格行里显示，不要弹全局 toast。
   */
  probe: (data: {
    provider_id: number
    model: string
    config_type?: string
    kind?: 'connect' | 'cap'
    declared_max?: number
    model_id?: number
  }) => api.post<any, ApiResponse<{
    model: string
    provider_id: number
    reachable: boolean
    accepted: number
    conclusion: string
    message: string
    latency_ms: number
    steps: Array<{
      tokens: number
      status: number
      kind: string
      completion_tokens: number
      finish_reason: string
      note?: string
    }>
  }>>('/ai/probe', data, { silent: true })
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
      /** 带上作品 id，生成的图会落进 books/<书名>/image/，而不是共享缓存目录 */
      bookId?: number
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

// AI 扫榜相关 API（榜单数据源：番茄 / 七猫，经 Go 后端的 internal/rank 注册表聚合）
export const rankAPI = {
  // 可用数据源（前端右上角切换器与页脚来源说明都由它驱动）
  getSources: () =>
    api.get<any, ApiResponse<{ sources: RankSource[]; default: string }>>('/ai/rank/sources', { silent: true }),
  // source 缺省或非法时后端回退默认数据源
  getLists: (source?: string) =>
    api.get<any, ApiResponse<{ source: string; groups: RankGroup[] }>>('/ai/rank/lists', {
      params: { source },
      silent: true
    }),
  // tab 的格式由各数据源自定：
  //   番茄 gender_rankMold_categoryId，如 "1_2_1141"（男频·阅读榜·西方奇幻）
  //   七猫 gender_rankType_dayType，如 "0_1_1"（男生·大热榜·日榜）
  //   起点 gender_type，如 "male_hotsales"
  // filters 是该榜内的一层筛选（起点用），格式 "catId:21,yearmonth:202608"
  getBooks: (params: { source?: string; tab: string; filters?: string; page?: number; size?: number }) =>
    api.get<any, ApiResponse<{ source: string; books: RankBook[] }>>('/ai/rank/books', { params, silent: true }),
  // 掉榜名单：只有能抓完整榜的数据源（起点）才有意义，其它源 ready 恒为 false
  getDrops: (params: { source?: string; tab: string; filters?: string }) =>
    api.get<any, ApiResponse<RankDrops>>('/ai/rank/drops', { params, silent: true }),
  // 在线书籍前 N 章目录（可选带正文）：AI 拆书弹窗「关联章节」用。
  // 仅实现了章节能力的数据源可用（目前番茄）；其他源后端返回 501
  getBookChapters: (params: { source?: string; bookId: string; limit?: number; content?: number; maxChars?: number }) =>
    api.get<any, ApiResponse<{ source: string; bookId: string; chapters: RankChapterItem[] }>>('/ai/rank/book-chapters', {
      params,
      silent: true
    })
}

// 榜单数据源元信息
export interface RankSource {
  id: string
  name: string
  desc: string
  /** 是否能把整榜一次抓完（决定掉榜视图是否可用） */
  fullCapture?: boolean
}

/** 榜单内一层筛选维度的一个取值 */
export interface RankFilterOption {
  id: string
  name: string
}

/** 榜单内的一层筛选维度（起点：题材 / 月份 / 周期）。番茄与七猫没有这一层 */
export interface RankFilterDim {
  key: string
  name: string
  options: RankFilterOption[]
  defaultId: string
}

export interface RankGroupItem {
  id: string
  name: string
  filters?: RankFilterDim[]
}

/** 掉榜的一本书 */
export interface RankDrop {
  bookId: string
  title: string
  author: string
  rank: number
}

export interface RankDrops {
  /** 今天是否已抓完整榜。没有它就不能判定掉榜，界面应显示"抓取中"而不是 0 */
  ready: boolean
  /** 对比基准日（上一份整榜的日期），为空表示今天第一次整榜 */
  baseDate: string
  drops: RankDrop[]
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
  /** readingCount 的数值口径：在读 / 热度 / 收藏；缺省按「在读」处理 */
  metricLabel?: string
  /** rankDiff 的口径：step = 真实名次差（番茄原生 / 七猫经快照补齐）；dir = 仅方向，只有 ±1 的意义；缺省按 step */
  rankDiffMode?: string
  /** 新上榜：上一份榜单快照里没有这本（由后端 rank_snapshots 判定） */
  isNew?: boolean
  /** 在快照里出现过的天数（含今天）。只有第 2 天起才说明问题 */
  daysOnList?: number
}

/** 在线书籍的一章（AI 拆书弹窗「关联章节」）。content 仅在请求带正文时返回 */
export interface RankChapterItem {
  id: string
  index: number
  title: string
  isFree: boolean
  content?: string
}

// 榜单分组（第一级），items 是该组下的榜单（第二级）
// 番茄：组 = 男频/女频阅读榜，items = 题材；七猫：组 = 男生/女生榜，items = 榜单类型；
// 起点：组 = 男频/女频，items = 榜单类型，且每个榜单自带 filters（题材/月份/周期）
export interface RankGroup {
  title: string
  items: RankGroupItem[]
}

// 对话相关 API
export const conversationAPI = {
  getAll: () => api.get<any, ApiResponse<any[]>>('/conversations'),
  list: () => api.get<any, ApiResponse<any[]>>('/conversations'),
  getByBook: (bookId: number) => api.get<any, ApiResponse<any[]>>(`/conversations/book/${bookId}`),
  create: (data: { book_id: number; title?: string }) => api.post<any, ApiResponse<any>>('/conversations', data),
  update: (id: number, data: { title: string }) => api.put<any, ApiResponse<any>>(`/conversations/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/conversations/${id}`),
  getMessages: (id: number, opts?: { silent?: boolean }) =>
    api.get<any, ApiResponse<any[]>>(`/conversations/${id}/messages`, opts as any),
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

// 星芒 Agent（独立板块）
// 事件流不走 axios：SSE 用 fetch + ReadableStream 手写解析（见 views/Agent.vue）
export const agentAPI = {
  models: () => api.get<any, ApiResponse<any[]>>('/agent/models'),
  books: () => api.get<any, ApiResponse<{ id: number; title: string }[]>>('/agent/books'),
  sessions: () => api.get<any, ApiResponse<any[]>>('/agent/sessions'),
  createSession: (data: { title?: string; scope_kind: 'global' | 'book'; scope_id?: number | null; model_id?: number | null; permission?: 'observe' | 'confirm' | 'auto' }) =>
    api.post<any, ApiResponse<any>>('/agent/sessions', data),
  patchSession: (id: string, data: { permission?: 'observe' | 'confirm' | 'auto' }) =>
    api.patch<any, ApiResponse<any>>(`/agent/sessions/${id}`, data),
  deleteSession: (id: string) => api.delete<any, ApiResponse>(`/agent/sessions/${id}`),
  startRun: (sessionId: string, data: { input: string; model_id?: number }) =>
    api.post<any, ApiResponse<{ run_id: string }>>(`/agent/sessions/${sessionId}/runs`, data),
  cancelRun: (runId: string) => api.post<any, ApiResponse>(`/agent/runs/${runId}/cancel`),
  deleteRun: (runId: string) => api.delete<any, ApiResponse>(`/agent/runs/${runId}`),
  editRunAnswer: (runId: string, text: string) =>
    api.put<any, ApiResponse>(`/agent/runs/${runId}/answer`, { text }),
  approveProposal: (id: string) => api.post<any, ApiResponse<{ status: string; result?: string }>>(`/agent/proposals/${id}/approve`),
  rejectProposal: (id: string) => api.post<any, ApiResponse<{ status: string }>>(`/agent/proposals/${id}/reject`),
  rollbackProposal: (id: string) => api.post<any, ApiResponse>(`/agent/proposals/${id}/rollback`)
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

// 系统信息
// 与后端启动时的「单实例探测 + 版本握手」共用同一个接口：
// 启动新 exe 时会先读这里的 version，判断是否在复用旧实例。
export interface HealthInfo {
  status: string
  message: string
  version: string
  commit: string
  buildTime: string
}

export const systemAPI = {
  // silent：版本探测失败不该弹全局错误提示（老版本后端没有这些字段也算正常）
  getHealth: () => api.get<any, HealthInfo>('/health', { silent: true }),

  // 退出程序：后端会优雅关闭服务并结束进程。
  // 仅本机来源可用（后端校验 RemoteAddr / Origin），从局域网访问时点它会返回 403。
  shutdown: () => api.post<any, { success: boolean; message: string }>('/shutdown', undefined, { silent: true })
}

export default api

// 链接目录的报告（"接管"与"重建索引"共用）
export interface LinkReport {
  root: string
  books: number
  created?: string[]
  existing?: string[]
  details?: Record<string, number>
  errors?: string[]
}

// 应用设置（本机路径）
export const settingsAPI = {
  getBooksRoot: () => api.get<any, ApiResponse<{ path: string }>>('/settings/books-root'),
  setBooksRoot: (path: string) => api.put<any, ApiResponse<{ path: string }>>('/settings/books-root', { path }),
  /** 弹系统文件夹选择框；cancelled=true 表示用户点了取消 */
  pickFolder: () => api.post<any, ApiResponse<{ path: string; cancelled: boolean }>>('/settings/pick-folder'),
  /** 扫描目标目录：里面有哪些作品、各自是否已在库中。只读，不改任何东西 */
  previewRoot: (path: string) => api.post<any, ApiResponse<LinkReport>>('/settings/books-root/preview', { path }),
  /** 接管：登记已有作品，不搬移、不复制、不删除任何文件 */
  linkRoot: (path: string) => api.put<any, ApiResponse<LinkReport>>('/settings/books-root/link', { path })
}

// 记忆记录：拆书（analysis）与生成历史（generation）
// 这两类原先只存在浏览器 localStorage 里，现在正文落在 <书>/memory/ 下的 .md，
// 数据库只留索引。full=1 表示"顺带把正文与 payload 一起返回"。
export interface MemoryRow {
  id: number
  book_id: number
  kind: 'analysis' | 'generation'
  external_id: string
  title: string
  source: string
  status: string
  file_path: string
  word_count: number
  missing: number
  parse_error?: string
  created_at: string
  updated_at: string
  /** 来自 JOIN books：历史中心要按作品筛选 */
  book_title?: string
  /** 以下字段仅在 full=1 时出现 */
  result?: string
  readable?: string
  payload?: string
  prompt?: string
  chapters?: number[]
}

export interface MemorySaveInput {
  kind: 'analysis' | 'generation'
  external_id?: string
  title: string
  source: string
  status?: string
  prompt?: string
  chapters?: number[]
  result: string
  readable?: string
  /** 原始记录 JSON 文本（记录原本的结构，无损） */
  payloadText?: string
  created_at_ms?: number
}

export const memoryAPI = {
  list: (bookId: number, kind?: string, full = false) =>
    api.get<any, ApiResponse<MemoryRow[]>>(`/books/${bookId}/memories`, {
      params: { kind, full: full ? 1 : undefined }
    }),
  create: (bookId: number, body: MemorySaveInput) =>
    api.post<any, ApiResponse<MemoryRow>>(`/books/${bookId}/memories`, body),
  get: (id: number) => api.get<any, ApiResponse<MemoryRow>>(`/memories/${id}`),
  update: (id: number, body: MemorySaveInput) =>
    api.put<any, ApiResponse<MemoryRow>>(`/memories/${id}`, body),
  remove: (id: number) => api.delete<any, ApiResponse>(`/memories/${id}`),
  /** 跨作品列出记录（历史中心用） */
  listAll: (kind?: string, full = true) =>
    api.get<any, ApiResponse<MemoryRow[]>>('/memories', {
      params: { kind, full: full ? 1 : undefined }
    }),
  summary: (bookId: number) => api.get<any, ApiResponse<any>>(`/books/${bookId}/memory/summary`)
}

/**
 * 对当前作品库做一次链接重建。
 *
 * 用途：手工把作品目录拷进作品库之后，让它们出现在书架里 ——
 * 不必走"导入"（那会再复制一份）。
 */
export const rebuildBookIndex = () =>
  api.post<any, ApiResponse<LinkReport>>('/books/rebuild-index')
