import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse, Book, Chapter, Prompt, Memo, ApiProvider, ApiModel, ChatMessage, RelatedContent, Generator, Character, UsageOverview, DailyUsage, ModelStats, MonthlyStats, Volume, ExperienceShare, GraphEntity, GraphRelation, KnowledgeGraphData, GraphVersion } from '@/types'

const api = axios.create({
  baseURL: '/api',
  timeout: 60000
})

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
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
  update: (id: number, data: Partial<Chapter>) => api.put<any, ApiResponse<Chapter>>(`/chapters/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/chapters/${id}`)
}

// 提示词相关API
export const promptAPI = {
  getAll: () => api.get<any, ApiResponse<Prompt[]>>('/prompts'),
  getOne: (id: number) => api.get<any, ApiResponse<Prompt>>(`/prompts/${id}`),
  create: (data: Partial<Prompt>) => api.post<any, ApiResponse<Prompt>>('/prompts', data),
  update: (id: number, data: Partial<Prompt>) => api.put<any, ApiResponse<Prompt>>(`/prompts/${id}`, data),
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
  delete: (id: number) => api.delete<any, ApiResponse>(`/providers/${id}`)
}

// API模型配置相关API
export const configAPI = {
  getAll: () => api.get<any, ApiResponse<ApiModel[]>>('/config'),
  getDefault: () => api.get<any, ApiResponse<ApiModel>>('/config/default'),
  getByProvider: (providerId: number) => api.get<any, ApiResponse<ApiModel[]>>(`/config/provider/${providerId}`),
  create: (data: Partial<ApiModel>) => api.post<any, ApiResponse<ApiModel>>('/config', data),
  update: (id: number, data: Partial<ApiModel>) => api.put<any, ApiResponse<ApiModel>>(`/config/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/config/${id}`),
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
  }) => api.post<any, ApiResponse<string>>('/ai/chat', data),
  generateDescription: (data: { title: string; promptId?: number }) => 
    api.post<any, ApiResponse<string>>('/ai/generate-description', data),
  recognizeCharacters: (data: { text: string; configId?: number; customPrompt?: string }) => 
    api.post<any, ApiResponse<any[]>>('/ai/recognize-characters', data)
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
  saveMessage: (id: number, data: { role: string; content: string }) => 
    api.post<any, ApiResponse<any>>(`/conversations/${id}/messages`, data),
  clearMessages: (id: number) => api.delete<any, ApiResponse>(`/conversations/${id}/messages`),
  deleteMessage: (conversationId: number, messageId: number) => 
    api.delete<any, ApiResponse>(`/conversations/${conversationId}/messages/${messageId}`),
  updateMessage: (conversationId: number, messageId: number, data: { content: string }) => 
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
  getGraph: (bookId: number) => api.get<any, ApiResponse<KnowledgeGraphData>>(`/knowledge-graph/graph/${bookId}`),
  getEntities: (bookId: number) => api.get<any, ApiResponse<GraphEntity[]>>(`/knowledge-graph/entities/book/${bookId}`),
  getEntity: (id: number) => api.get<any, ApiResponse<GraphEntity>>(`/knowledge-graph/entities/${id}`),
  createEntity: (data: Partial<GraphEntity>) => api.post<any, ApiResponse<GraphEntity>>('/knowledge-graph/entities', data),
  updateEntity: (id: number, data: Partial<GraphEntity>) => api.put<any, ApiResponse<GraphEntity>>(`/knowledge-graph/entities/${id}`, data),
  deleteEntity: (id: number) => api.delete<any, ApiResponse>(`/knowledge-graph/entities/${id}`),
  getRelations: (bookId: number) => api.get<any, ApiResponse<GraphRelation[]>>(`/knowledge-graph/relations/book/${bookId}`),
  createRelation: (data: Partial<GraphRelation>) => api.post<any, ApiResponse<GraphRelation>>('/knowledge-graph/relations', data),
  updateRelation: (id: number, data: Partial<GraphRelation>) => api.put<any, ApiResponse<GraphRelation>>(`/knowledge-graph/relations/${id}`, data),
  deleteRelation: (id: number) => api.delete<any, ApiResponse>(`/knowledge-graph/relations/${id}`),
  analyze: (data: { bookId: number; configId?: number; scope?: 'all' | 'chapters'; chapterIds?: number[]; versionId: number }) =>
    api.post<any, ApiResponse<KnowledgeGraphData>>('/knowledge-graph/analyze', data, { timeout: 300000 }),
  clearGraph: (bookId: number) => api.delete<any, ApiResponse>(`/knowledge-graph/graph/${bookId}`),
  getVersions: (bookId: number) => api.get<any, ApiResponse<GraphVersion[]>>(`/knowledge-graph/versions/${bookId}`),
  createVersion: (bookId: number, name?: string) => api.post<any, ApiResponse<GraphVersion>>('/knowledge-graph/versions', { bookId, name }),
  getVersionData: (versionId: number) => api.get<any, ApiResponse<KnowledgeGraphData>>(`/knowledge-graph/version/${versionId}`),
  renameVersion: (versionId: number, name: string) => api.put<any, ApiResponse<GraphVersion>>(`/knowledge-graph/versions/${versionId}`, { name }),
  deleteVersion: (bookId: number, versionId: number) => api.delete<any, ApiResponse>(`/knowledge-graph/versions/${bookId}/${versionId}`)
}

export default api
