// 书本类型
export interface Book {
  id: number
  title: string
  description: string
  cover?: string
  author?: string
  category?: string
  tags?: string[]
  status?: 'draft' | 'published' | 'completed'
  word_count?: number
  target_words?: number
  chapter_count?: number
  is_public?: boolean
  created_at: string
  updated_at: string
  published_at?: string
  last_read_at?: string
  metadata?: Record<string, any>
}

// 章节类型
export interface Chapter {
  id: number
  book_id: number
  title: string
  content: string
  summary?: string
  order_num: number
  type: 'chapter' | 'memo'
  volume_id?: number
  created_at: string
  updated_at: string
}

// 提示词类型
export interface Prompt {
  id: number
  name: string
  description?: string
  content: string
  category: string
  order_num: number
  card_type?: 'normal' | 'encrypted'
  password?: string | null
  fields?: Array<{
    name: string
    label: string
    type: 'text' | 'textarea' | 'select'
    options: string[]
    description: string
    required: boolean
  }>
  subcategories?: string[]
  created_at?: string
  updated_at?: string
  creator_name?: string
  creator_avatar?: string
  version?: string
  use_count?: number
  like_count?: number
}

// 备忘录类型
export interface Memo {
  id: number
  title: string
  content: string
  category: string
  order_num: number
  tags?: string
  is_pinned?: number
  word_count?: number
  created_at: string
  updated_at: string
}

export interface ExperienceShare {
  id: number
  title: string
  summary: string
  content: string
  content_render_mode?: 'markdown' | 'html'
  cover_url?: string | null
  pdf_file_url?: string | null
  pdf_file_name?: string | null
  pdf_file_size?: number
  create_type: 'manual' | 'pdf_import'
  author_id?: number | null
  author_name?: string | null
  created_at: string
  updated_at: string
  status?: string | null
  pdf_parse_status?: 'success' | 'empty' | 'failed' | string | null
  pdf_parse_result?: string | null
  source_file_name?: string | null
}

// API服务商类型
export interface ApiProvider {
  id: number
  name: string
  provider_type: string
  api_key: string
  api_url: string
  created_at: string
  updated_at: string
  models?: ApiModel[]
}

// API模型类型
export interface ApiModel {
  id: number
  provider_id: number
  name: string
  model: string
  temperature: number
  max_tokens: number
  is_default: number
  created_at: string
  updated_at: string
  // 关联的服务商信息
  provider_name?: string
  provider_type?: string
  api_key?: string
  api_url?: string
  base_url?: string
  model_name?: string
}

// 兼容旧的ApiConfig类型（用于渐进迁移）
export type ApiConfig = ApiModel

// 消息类型
export interface ChatMessage {
  id?: number
  role: 'system' | 'user' | 'assistant'
  content: string
  created_at?: string
  timestamp?: number
}

// 关联内容类型
export interface RelatedContent {
  type: 'book' | 'chapter' | 'memo' | 'character'
  id: number
  title: string
  content: string
}

// 对话类型
export interface Conversation {
  id: number
  book_id: number
  title: string
  message_count?: number
  created_at: string
  updated_at: string
}

// 对话消息类型
export interface ConversationMessage {
  id: number
  conversation_id: number
  role: 'system' | 'user' | 'assistant'
  content: string
  created_at: string
}

// API响应类型
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
}

// 生成器类型
export interface Generator {
  id: number
  name: string
  description: string
  icon: string
  core_prompt: string
  remark?: string
  order_num: number
  created_at: string
  updated_at: string
}

// 角色类型
export interface Character {
  id: number
  book_id: number
  name: string
  gender: 'male' | 'female' | 'unknown' | 'none' | string
  personality: string
  info: string
  folder: string
  folders: string[]
  created_at: string
  updated_at: string
}

// 使用统计概览
export interface UsageOverview {
  promptCount: number
  generatorCount: number
  bookCount: number
  totalUsageCount: number
  totalTokens: number
}

// 每日使用统计
export interface DailyUsage {
  date: string
  model_id: number
  model_name: string
  provider_name: string
  usage_count: number
  total_tokens: number
}

// 模型使用统计
export interface ModelStats {
  model_id: number
  model_name: string
  provider_name: string
  total_usage: number
  total_tokens: number
}

// 月度使用统计
export interface MonthlyStats {
  month: string
  total_usage: number
  total_tokens: number
  model_count: number
}

// 分卷类型
export interface Volume {
  id: number
  book_id: number
  title: string
  order_num: number
  created_at: string
  updated_at: string
}

export type GraphEntityType = 'character' | 'location' | 'item' | 'faction' | 'event' | 'skill' | 'clue'

export interface GraphEntity {
  id: number
  book_id: number
  name: string
  type: GraphEntityType
  description: string
  metadata: Record<string, any>
  created_at: string
  updated_at: string
}

export interface GraphRelation {
  id: number
  book_id: number
  source_id: number
  target_id: number
  relation_type: string
  description: string
  evidence: string
  confidence: number
  created_at: string
  updated_at: string
}

export interface KnowledgeGraphData {
  entities: GraphEntity[]
  relations: GraphRelation[]
  stats?: {
    entityCount: number
    relationCount: number
  }
  version?: {
    id: number
    name: string
    versions: GraphVersion[]
  }
}

export interface GraphVersion {
  id: number
  book_id: number
  name: string
  description: string
  scope: 'all' | 'chapters'
  chapter_count: number
  entity_count: number
  relation_count: number
  created_at: string
}
