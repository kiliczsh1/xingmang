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
  word_count?: number
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
    optionLabels?: string[]
    description: string
    required: boolean
    defaultValue?: string
  }>
  subcategories?: string[]
  metadata?: Record<string, any>
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
  version?: string | null
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
  url_suffix_mode: 'compat' | 'none'
  api_format?: 'openai' | 'claude' | 'gemini'
  use_full_url?: boolean
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
  top_p: number
  frequency_penalty: number
  description: string
  is_default: number
  enabled: number
  sort_order: number
  category?: string
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
  displayContent?: string
  created_at?: string
  timestamp?: number
}

// 关联内容类型
export interface RelatedContent {
  type: 'book' | 'chapter' | 'memo' | 'character' | 'worldbook'
  id: number | string
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
  tutorial_url?: string
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
  parent_id?: number | null
  title: string
  order_num: number
  created_at: string
  updated_at: string
  folder_name?: string
}

export type GraphEntityType = 'character' | 'location' | 'item' | 'faction' | 'event' | 'skill' | 'clue'

export interface GraphEntity {
  id: number
  book_id: number
  name: string
  type: GraphEntityType
  description: string
  metadata: Record<string, any>
  version_id?: number | null
  /**
   * 实体合并系统的主/副词条标记：
   *   null/undefined = 主词条（画布上真正渲染的节点）
   *   数字           = 副词条（别名），值指向同版本内主词条的 id
   */
  primary_id?: number | null
  /** 仅导入时使用：声明该实体出自哪些章节，落库到 graph_chapter_refs */
  chapter_ids?: number[]
  created_at: string
  updated_at: string
}

/** 疑似重复的实体候选对，由服务端启发式发现并给出可解释理由，供人工合并检测确认 */
export interface EntityMergeCandidate {
  a: GraphEntity
  b: GraphEntity
  score: number
  reasons: string[]
  /** 建议的合并方向：把 suggest_duplicate_id 并入 suggest_primary_id */
  suggest_primary_id: number
  suggest_duplicate_id: number
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
  version_id?: number | null
  /** 仅导入时使用：声明该关系出自哪些章节 */
  chapter_ids?: number[]
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

/** 图谱筛选条件；types 为空数组表示不限制类型 */
export interface GraphFilters {
  types: GraphEntityType[]
}

/** 章节与图谱的关联，服务端按章节分组返回 */
export interface GraphChapterRef {
  chapter_id: number
  entity_ids: number[]
  relation_ids: number[]
}

export type EntryCategoryType = 'character' | 'world_setting' | 'location' | 'item' | 'faction' | 'event' | 'skill' | 'clue' | 'other'

export interface EntryCategory {
  id: number
  name: string
  type: EntryCategoryType
  icon: string
  color: string
  sort_order: number
  entry_count: number
  created_at: string
  updated_at: string
}

export interface Entry {
  id: number
  book_id: number
  category_id: number
  category_type: EntryCategoryType
  category_name?: string
  name: string
  avatar?: string
  tags: string[]
  appearance: string
  background: string
  personality?: string
  relationships?: string
  custom_fields?: Record<string, string>
  description: string
  is_public: boolean
  creator_name?: string
  creator_avatar?: string
  use_count: number
  like_count: number
  created_at: string
  updated_at: string
}

export interface EntryListResponse {
  entries: Entry[]
  categories: EntryCategory[]
  total: number
  page: number
  page_size: number
}

// ---------- 插件系统 ----------

// 插件来源：内置 / 市场
export type PluginSource = 'builtin' | 'marketplace'

// 插件管理面契约：Plugins.vue 展示、plugin-prefs 持久化、未来 /plugins REST API 共用
export interface PluginManifest {
  name: string        // 唯一 ID，与 ChatPlugin.name 一致（如 prompt-inject）
  label: string       // 可读中文名
  description: string
  enabled: boolean
  source: PluginSource
  icon: string        // Element Plus 图标组件名，component :is 渲染
  color: string       // 卡片头像品牌色
  version?: string    // 预留
  author?: string     // 预留
}
