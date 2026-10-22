package database

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func GetDB() *sql.DB {
	return db
}

// Close 关闭数据库连接。
//
// ⚠ 关闭前先做 WAL checkpoint(TRUNCATE)：WAL 模式下最新写入可能只存在于 -wal 文件里，
// 此时用户若直接复制 xingmang.db 做备份，拿到的是**缺最新数据**的旧库。
// TRUNCATE 会把 WAL 内容全部并入主库并清空 -wal，退出后目录里只剩一个自包含的 .db。
func Close() {
	if db != nil {
		if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			// 不致命：数据仍在 -wal 中，下次启动会自动恢复，只是备份需带上 -wal
			log.Printf("WAL checkpoint 失败（数据仍在 -wal 中，下次启动会自动恢复）: %v", err)
		}
		db.Close()
		db = nil
	}
}

func Init(dbPath string) (*sql.DB, error) {
	var err error
	if dbPath == "" {
		dbPath = "xingmang.db"
	}
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=10000&_txlock=immediate")
	if err != nil {
		return nil, err
	}

	// SQLite file-level locking: limit to 1 open connection to avoid SQLITE_BUSY
	// WAL mode allows concurrent reads, but only 1 writer at a time
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Enable foreign keys
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	statements := []string{
		// books
		`CREATE TABLE IF NOT EXISTS books (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  description TEXT,
  cover TEXT,
  author TEXT,
  category TEXT,
  tags TEXT,
  status TEXT DEFAULT 'draft',
  is_public INTEGER DEFAULT 0,
  target_words INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// chapters
		`CREATE TABLE IF NOT EXISTS chapters (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  content TEXT,
  summary TEXT DEFAULT '',
  order_num INTEGER DEFAULT 0,
  type TEXT DEFAULT 'chapter',
  volume_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  FOREIGN KEY (volume_id) REFERENCES volumes(id) ON DELETE SET NULL
)`,
		// prompts
		`CREATE TABLE IF NOT EXISTS prompts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  content TEXT NOT NULL,
  category TEXT,
  order_num INTEGER DEFAULT 0,
  fields TEXT,
  subcategories TEXT,
  card_type TEXT DEFAULT 'normal',
  password TEXT,
  is_locked INTEGER DEFAULT 0,
  insight TEXT,
  creator_name TEXT,
  version TEXT,
  metadata TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// prompt_insights：卡片使用心得，一条卡片可以有多条（时间线）
		// 新建表所以可以直接带外键（老表补列做不到），删卡片时级联清掉心得。
		`CREATE TABLE IF NOT EXISTS prompt_insights (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  prompt_id INTEGER NOT NULL,
  content TEXT NOT NULL,
  rating INTEGER DEFAULT 0,
  tags TEXT DEFAULT '[]',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (prompt_id) REFERENCES prompts(id) ON DELETE CASCADE
)`,
		// 表在此处新建，索引跟着建是安全的（需要补列的 ALTER 才必须放 migrate 里）
		`CREATE INDEX IF NOT EXISTS idx_prompt_insights_prompt ON prompt_insights(prompt_id)`,
		// memos
		`CREATE TABLE IF NOT EXISTS memos (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  content TEXT,
  category TEXT,
  order_num INTEGER DEFAULT 0,
  tags TEXT DEFAULT '[]',
  is_pinned INTEGER DEFAULT 0,
  word_count INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// experience_shares
		// board：所属吧（tutorial = 教程吧，experience = 经验分享吧）。存量库由
		// migrateExperienceSharesColumns 补列并把老帖刷成 experience。
		`CREATE TABLE IF NOT EXISTS experience_shares (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  summary TEXT,
  content TEXT NOT NULL,
  content_render_mode TEXT DEFAULT 'markdown',
  cover_url TEXT,
  pdf_file_url TEXT,
  pdf_file_name TEXT,
  pdf_file_size INTEGER,
  create_type TEXT DEFAULT 'manual',
  author_id INTEGER,
  author_name TEXT DEFAULT '星芒用户',
  status TEXT DEFAULT 'published',
  pdf_parse_status TEXT,
  pdf_parse_result TEXT,
  source_file_name TEXT,
  version TEXT,
  board TEXT DEFAULT 'experience',
  section_id INTEGER,
  is_featured INTEGER DEFAULT 0,
  is_pinned INTEGER DEFAULT 0,
  like_count INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// experience_sections（经验分享的贴吧式自定义分区；「精华」「全部」是内置入口，不落这张表）
		// 注意：不给 section_id 加外键 —— 存量库是 ALTER 补出来的列，SQLite 不支持给已有列加 REFERENCES。
		// board：分区按吧隔离，两个吧各自最多 15 个，可以同名。
		`CREATE TABLE IF NOT EXISTS experience_sections (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  board TEXT DEFAULT 'experience',
  name TEXT NOT NULL,
  order_num INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// app_meta：一次性动作的标记（如「经验分享的引导帖已种过」）。
		// ⚠ 有了它，引导帖被用户删掉之后不会在下次启动时又冒出来。
		`CREATE TABLE IF NOT EXISTS app_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL DEFAULT '',
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// volumes
		`CREATE TABLE IF NOT EXISTS volumes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  parent_id INTEGER,
  title TEXT NOT NULL,
  order_num INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
)`,
		// api_providers
		// 多模态补齐（2026-10-15）：新增渠道 / 扩展认证 / 并发 / 长尾配置列。
		// 存量库由 migrateAPIProvidersColumns 增量补列，新库在此直接建全。
		`CREATE TABLE IF NOT EXISTS api_providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  provider_type TEXT NOT NULL,
  api_key TEXT,
  api_url TEXT NOT NULL,
  api_format TEXT DEFAULT 'openai',
  use_full_url INTEGER DEFAULT 0,
  auth_type TEXT NOT NULL DEFAULT 'api_key',
  access_key TEXT,
  secret_key TEXT,
  group_id TEXT,
  endpoint_id TEXT,
  concurrency INTEGER DEFAULT 1,
  channel_name TEXT DEFAULT '',
  extra_config TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// api_models
		// 多模态补齐（2026-10-15）：新增 config_type（强类型）与高级参数列。
		// category 保留为兼容字段，新逻辑一律以 config_type 为准。
		`CREATE TABLE IF NOT EXISTS api_models (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  model TEXT NOT NULL,
  temperature REAL DEFAULT 0.7,
  max_tokens INTEGER DEFAULT 0,
  top_p REAL DEFAULT 0.9,
  frequency_penalty REAL DEFAULT 0.0,
  description TEXT DEFAULT '',
  is_default INTEGER DEFAULT 0,
  enabled INTEGER DEFAULT 1,
  sort_order INTEGER DEFAULT 0,
  category TEXT DEFAULT '',
  config_type TEXT NOT NULL DEFAULT 'chat',
  prompt TEXT DEFAULT '',
  prompt_role TEXT DEFAULT 'system',
  top_k INTEGER,
  presence_penalty REAL,
  stop TEXT DEFAULT '',
  max_tokens_mode TEXT DEFAULT 'auto',
  thinking_mode INTEGER DEFAULT 0,
  thinking_effort TEXT DEFAULT 'default',
  generation_config_source TEXT DEFAULT 'manual',
  request_rules TEXT,
  param_groups TEXT DEFAULT '',
  default_params TEXT DEFAULT '',
  is_free INTEGER DEFAULT 0,
  input_ratio REAL,
  output_ratio REAL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (provider_id) REFERENCES api_providers(id) ON DELETE CASCADE
)`,
		// conversations
		`CREATE TABLE IF NOT EXISTS conversations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  title TEXT DEFAULT '新对话',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
)`,
		// conversation_messages
		`CREATE TABLE IF NOT EXISTS conversation_messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  conversation_id INTEGER NOT NULL,
  role TEXT NOT NULL,
  content TEXT NOT NULL,
  display_content TEXT DEFAULT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
)`,
		// generators
		`CREATE TABLE IF NOT EXISTS generators (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  icon TEXT DEFAULT 'Lightning',
  core_prompt TEXT,
  remark TEXT,
  order_num INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// characters
		`CREATE TABLE IF NOT EXISTS characters (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  name TEXT NOT NULL DEFAULT '未命名角色',
  gender TEXT DEFAULT 'unknown',
  personality TEXT,
  info TEXT,
  folder TEXT DEFAULT '全部',
  folders TEXT DEFAULT '[]',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
)`,
		// usage_stats
		`CREATE TABLE IF NOT EXISTS usage_stats (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  date TEXT NOT NULL,
  model_id INTEGER,
  model_name TEXT NOT NULL,
  provider_name TEXT,
  usage_count INTEGER DEFAULT 1,
  total_tokens INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (model_id) REFERENCES api_models(id) ON DELETE SET NULL
)`,
		// graph_versions (must be before graph_entities and graph_relations due to foreign key references)
		`CREATE TABLE IF NOT EXISTS graph_versions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  description TEXT DEFAULT '',
  scope TEXT DEFAULT 'all',
  chapter_count INTEGER DEFAULT 0,
  entity_count INTEGER DEFAULT 0,
  relation_count INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
)`,
		// graph_entities
		// graph_entities
		// primary_id：实体合并系统的主/副词条标记。
		//   NULL  = 主词条（画布上真正显示的节点）
		//   非空  = 副词条，值指向同版本内的主词条 id（关系端点读接口会重定向到主词条）
		// 这里刻意不加 FOREIGN KEY：存量库只能 ALTER TABLE ADD COLUMN，加不了外键，
		// 会造成「新库有外键、老库没有」的行为差异。悬空由 DeleteEntity 显式清理。
		`CREATE TABLE IF NOT EXISTS graph_entities (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  type TEXT NOT NULL DEFAULT 'character',
  description TEXT DEFAULT '',
  metadata TEXT DEFAULT '{}',
  version_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  primary_id INTEGER,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  FOREIGN KEY (version_id) REFERENCES graph_versions(id) ON DELETE CASCADE
)`,
		// graph_relations
		`CREATE TABLE IF NOT EXISTS graph_relations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  source_id INTEGER NOT NULL,
  target_id INTEGER NOT NULL,
  relation_type TEXT NOT NULL DEFAULT 'related',
  description TEXT DEFAULT '',
  evidence TEXT DEFAULT '',
  confidence REAL DEFAULT 1.0,
  version_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  FOREIGN KEY (source_id) REFERENCES graph_entities(id) ON DELETE CASCADE,
  FOREIGN KEY (target_id) REFERENCES graph_entities(id) ON DELETE CASCADE,
  FOREIGN KEY (version_id) REFERENCES graph_versions(id) ON DELETE CASCADE
)`,
		// graph_chapter_refs：实体/关系与章节的关联（AI 分析出处、手工关联）。
		// entity_id / relation_id 二选一，用 CHECK 约束并在两列上都挂外键，
		// 这样删实体、删关系、删版本都能沿 CASCADE 自动清掉关联，不留孤儿。
		`CREATE TABLE IF NOT EXISTS graph_chapter_refs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  version_id INTEGER,
  chapter_id INTEGER NOT NULL,
  entity_id INTEGER,
  relation_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  CHECK ((entity_id IS NOT NULL) <> (relation_id IS NOT NULL)),
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  FOREIGN KEY (entity_id) REFERENCES graph_entities(id) ON DELETE CASCADE,
  FOREIGN KEY (relation_id) REFERENCES graph_relations(id) ON DELETE CASCADE,
  FOREIGN KEY (version_id) REFERENCES graph_versions(id) ON DELETE CASCADE
)`,
		// entry_categories
		`CREATE TABLE IF NOT EXISTS entry_categories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  type TEXT NOT NULL DEFAULT 'other',
  icon TEXT DEFAULT 'Document',
  color TEXT DEFAULT '#409EFF',
  sort_order INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// entries
		`CREATE TABLE IF NOT EXISTS entries (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER,
  category_id INTEGER NOT NULL,
  category_type TEXT NOT NULL DEFAULT 'other',
  name TEXT NOT NULL,
  avatar TEXT,
  tags TEXT DEFAULT '[]',
  appearance TEXT DEFAULT '',
  background TEXT DEFAULT '',
  personality TEXT,
  relationships TEXT,
  custom_fields TEXT DEFAULT '{}',
  description TEXT DEFAULT '',
  is_public INTEGER DEFAULT 0,
  creator_name TEXT,
  creator_avatar TEXT,
  use_count INTEGER DEFAULT 0,
  like_count INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
  FOREIGN KEY (category_id) REFERENCES entry_categories(id) ON DELETE CASCADE
)`,
		// 榜单快照（AI 扫榜）：按天记录榜单名次，用于算出真实的名次变动幅度。
		// 番茄接口自带 rankPosDiff，七猫只给 -1/0/1 的方向，靠这份快照补齐幅度；
		// 起点连方向都没有，完全依赖它。
		//
		// tab 是「榜标识」，可能带筛选条件（如 "male_hotsales|catId=21"），见 rank.BookQuery.ScopeKey。
		// scope 区分快照来源：page=用户实际翻过的那几页；full=整榜翻完（可穷尽的数据源才有，
		// 是「新上榜 / 掉榜」可信判定的前提）。
		// 同一 (源, 榜, 书, 日期, scope) 只有一行，当天重复抓取按最新名次覆盖（见 internal/ranksnap）。
		`CREATE TABLE IF NOT EXISTS rank_snapshots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source TEXT NOT NULL,
  tab TEXT NOT NULL,
  book_id TEXT NOT NULL,
  rank INTEGER NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  author TEXT NOT NULL DEFAULT '',
  heat TEXT NOT NULL DEFAULT '',
  scope TEXT NOT NULL DEFAULT 'page',
  captured_date TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE (source, tab, book_id, captured_date, scope)
)`,
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_rank_snapshots_prev ON rank_snapshots(source, tab, captured_date)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_entities_book ON graph_entities(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_relations_book ON graph_relations(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_relations_source ON graph_relations(source_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_relations_target ON graph_relations(target_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_versions_book ON graph_versions(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_entities_version ON graph_entities(version_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_relations_version ON graph_relations(version_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_chapter_refs_chapter ON graph_chapter_refs(chapter_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_chapter_refs_entity ON graph_chapter_refs(entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_chapter_refs_relation ON graph_chapter_refs(relation_id)`,
		`CREATE INDEX IF NOT EXISTS idx_graph_chapter_refs_version ON graph_chapter_refs(version_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_stats_date ON usage_stats(date)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_stats_model ON usage_stats(model_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_book ON entries(book_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_category ON entries(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_entries_category_type ON entries(category_type)`,

		// ===== 星芒 Agent =====
		// 会话：scope 绑定（global=全库只读 / book=绑定单本书），与普通对话表语义不同故并存
		`CREATE TABLE IF NOT EXISTS agent_sessions (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL DEFAULT '新会话',
  scope_kind TEXT NOT NULL DEFAULT 'global',
  scope_id INTEGER,
  model_id INTEGER,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// run：一次用户输入触发的完整 agent 执行
		`CREATE TABLE IF NOT EXISTS agent_runs (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'running',
  input TEXT NOT NULL,
  error TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (session_id) REFERENCES agent_sessions(id) ON DELETE CASCADE
)`,
		// 账本：append-only，seq 每会话严格递增（UNIQUE 是幂等根基，勿删）。
		// 会话历史即账本：上下文装配从 events 重放，不另设消息表。
		`CREATE TABLE IF NOT EXISTS agent_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  session_id TEXT NOT NULL,
  run_id TEXT,
  seq INTEGER NOT NULL,
  type TEXT NOT NULL,
  payload TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(session_id, seq),
  FOREIGN KEY (session_id) REFERENCES agent_sessions(id) ON DELETE CASCADE
)`,
		// 工具调用记录（审计与回放）
		`CREATE TABLE IF NOT EXISTS agent_tool_calls (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  name TEXT NOT NULL,
  arguments TEXT,
  state TEXT NOT NULL,
  result TEXT,
  error TEXT,
  started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  ended_at DATETIME,
  FOREIGN KEY (run_id) REFERENCES agent_runs(id) ON DELETE CASCADE
)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_runs_session ON agent_runs(session_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_events_run ON agent_events(run_id, seq)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_tool_calls_run ON agent_tool_calls(run_id)`,
		// 写操作提案（方案 B 三档权限）：写工具先建提案，批准后才真正落库。
		// chapter 类提案的 before_text 存原文快照，用于回滚。
		`CREATE TABLE IF NOT EXISTS agent_change_proposals (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  run_id TEXT,
  tool_call_id TEXT,
  tool_name TEXT NOT NULL,
  target TEXT NOT NULL DEFAULT '',
  arguments TEXT NOT NULL DEFAULT '{}',
  risk TEXT NOT NULL DEFAULT 'low',
  summary TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  before_text TEXT,
  after_text TEXT,
  apply_error TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  decided_at DATETIME,
  applied_at DATETIME,
  FOREIGN KEY (session_id) REFERENCES agent_sessions(id) ON DELETE CASCADE
)`,
		`CREATE INDEX IF NOT EXISTS idx_agent_proposals_session ON agent_change_proposals(session_id, created_at)`,

		// ===== 工作流画布（节点式工作流，执行完全在前端，Go 侧只做持久化与文件落盘）=====
		// graph 存完整 {nodes:[],edges:[]} JSON 字符串，Go 不解析节点语义
		`CREATE TABLE IF NOT EXISTS workflows (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT '',
  graph TEXT NOT NULL DEFAULT '{"nodes":[],"edges":[]}',
  config TEXT NOT NULL DEFAULT '{}',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// 一次执行 = 一行；logs 与 context 为 JSON 字符串
		`CREATE TABLE IF NOT EXISTS workflow_runs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  workflow_id INTEGER NOT NULL,
  status TEXT NOT NULL DEFAULT 'running',
  logs TEXT NOT NULL DEFAULT '[]',
  context TEXT NOT NULL DEFAULT '{}',
  error TEXT,
  output_dir TEXT NOT NULL DEFAULT '',
  started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  finished_at DATETIME,
  FOREIGN KEY (workflow_id) REFERENCES workflows(id) ON DELETE CASCADE
)`,
		// 执行落盘文件索引（真实文件在 workflow_outputs/<runId>/ 下）
		`CREATE TABLE IF NOT EXISTS workflow_outputs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id INTEGER NOT NULL DEFAULT 0,
  workflow_id INTEGER NOT NULL DEFAULT 0,
  kind TEXT NOT NULL DEFAULT 'node',
  rel_path TEXT NOT NULL,
  size INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow ON workflow_runs(workflow_id, started_at)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_outputs_run ON workflow_outputs(run_id, created_at)`,
		// 自适应 max_tokens 的探测结果缓存（见 internal/aitokens）。
		//
		// 存的是「这条模型配置实测可行过的最大输出」，不是"真值"：上游能力变了、
		// 或换了渠道，探测会自动覆盖它。key 含 模型行 id + 渠道 id + 模型串，
		// 三者任一变化都会自然失效。
		//
		// 用 ON CONFLICT 覆盖而不是追加历史 —— 这里只关心"最近一次可行值"。
		`CREATE TABLE IF NOT EXISTS model_max_tokens (
  cache_key TEXT PRIMARY KEY,
  max_tokens INTEGER NOT NULL,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,

		// 书内自由文档（大纲 / 灵感）：正文在 outlines/ 与 ideas/ 的 .md 里，这里只做索引。
		`CREATE TABLE IF NOT EXISTS book_docs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  kind TEXT NOT NULL DEFAULT 'outline',
  title TEXT NOT NULL,
  file_path TEXT NOT NULL,
  word_count INTEGER NOT NULL DEFAULT 0,
  pinned INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		`CREATE INDEX IF NOT EXISTS idx_book_docs_book ON book_docs(book_id, kind)`,

		// 应用级设置（key/value）。首个键是 books_root（作品库根目录）。
		`CREATE TABLE IF NOT EXISTS app_settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL DEFAULT '',
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,

		// 记忆记录（拆书 / 生成历史）。正文在 memory/analysis 与 memory/history 的 .md 里，
		// 这张表只做索引 —— 按 docs/books-memory-storage.md v3 的决策，不保存正文副本。
		`CREATE TABLE IF NOT EXISTS memory_records (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  book_id INTEGER NOT NULL,
  kind TEXT NOT NULL,
  external_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'completed',
  file_path TEXT NOT NULL DEFAULT '',
  word_count INTEGER NOT NULL DEFAULT 0,
  content_hash TEXT NOT NULL DEFAULT '',
  missing INTEGER NOT NULL DEFAULT 0,
  parse_error TEXT NOT NULL DEFAULT '',
  synced_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// external_id 去重：重复链接同一个目录时不会产生重复记录。
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_memory_external ON memory_records(book_id, kind, external_id)`,
		`CREATE INDEX IF NOT EXISTS idx_memory_book ON memory_records(book_id, kind, created_at DESC)`,
	}

	for _, stmt := range statements {
		if _, err = db.Exec(stmt); err != nil {
			db.Close()
			db = nil
			return nil, err
		}
	}

	// 迁移：为存量库的 books 表补充缺失列（新库已在 CREATE TABLE 中定义）
	if err = migrateBooksColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 api_providers 表补充渠道 / 扩展认证 / 并发列
	if err = migrateAPIProvidersColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 api_models 表补充 category / config_type / 高级参数列
	if err = migrateAPIModelsColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 graph_entities 表补充 primary_id 列（实体合并的主/副词条）
	if err = migrateGraphEntitiesPrimaryColumn(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 prompts 表补充扩展列（metadata / is_locked / insight）
	if err = migratePromptsColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 rank_snapshots 表补充 scope 列（整榜快照）
	if err = migrateRankSnapshotsColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 experience_shares 表补充分区 / 精华 / 点赞 / 所属吧列
	if err = migrateExperienceSharesColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 experience_sections 表补充所属吧列（分区按吧隔离）
	if err = migrateExperienceSectionsColumns(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 agent_sessions 表补充 permission 列（写入权限三档）
	if err = ensureColumns(db, "agent_sessions", []struct{ name, ddl string }{
		{"permission", "permission TEXT NOT NULL DEFAULT 'confirm'"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 迁移：为存量库的 workflows 表补充 category 列（工作流分类）
	if err = ensureColumns(db, "workflows", []struct{ name, ddl string }{
		{"category", "category TEXT NOT NULL DEFAULT ''"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 作品文件夹化（docs/books-folder-storage.md）：
	// 正文/人物/世界观/大纲/灵感/备忘录全部落成磁盘文件，DB 只保留索引。
	// ⚠ 这些列必须走 ensureColumns；CREATE TABLE IF NOT EXISTS 不会给存量表加列。
	if err = ensureColumns(db, "books", []struct{ name, ddl string }{
		{"folder_path", "folder_path TEXT NOT NULL DEFAULT ''"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}
	if err = ensureColumns(db, "chapters", []struct{ name, ddl string }{
		{"file_path", "file_path TEXT NOT NULL DEFAULT ''"},
		{"word_count", "word_count INTEGER NOT NULL DEFAULT 0"},
		{"content_hash", "content_hash TEXT NOT NULL DEFAULT ''"},
		{"file_mtime", "file_mtime INTEGER NOT NULL DEFAULT 0"},
		{"missing", "missing INTEGER NOT NULL DEFAULT 0"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}
	if err = ensureColumns(db, "characters", []struct{ name, ddl string }{
		{"file_path", "file_path TEXT NOT NULL DEFAULT ''"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}
	if err = ensureColumns(db, "entries", []struct{ name, ddl string }{
		{"file_path", "file_path TEXT NOT NULL DEFAULT ''"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}
	if err = ensureColumns(db, "memos", []struct{ name, ddl string }{
		{"book_id", "book_id INTEGER NOT NULL DEFAULT 0"},
		{"file_path", "file_path TEXT NOT NULL DEFAULT ''"},
	}); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 记忆落盘（docs/books-memory-storage.md）：
	// 对话 / Agent 会话 / 图谱版本也变成"文件是真相"，这三张表只保留索引列。
	// 记忆类型的新增标记统一为：file_path（.md 相对路径）、content_hash（冲突检测）、
	// missing（文件不见了，不删行）、parse_error（文件被改坏的人话原因）、synced_at（最近落盘时间）。
	memoryIndexCols := []struct{ name, ddl string }{
		{"file_path", "file_path TEXT NOT NULL DEFAULT ''"},
		{"content_hash", "content_hash TEXT NOT NULL DEFAULT ''"},
		{"missing", "missing INTEGER NOT NULL DEFAULT 0"},
		{"parse_error", "parse_error TEXT NOT NULL DEFAULT ''"},
		{"synced_at", "synced_at DATETIME"},
	}
	if err = ensureColumns(db, "conversations", memoryIndexCols); err != nil {
		db.Close()
		db = nil
		return nil, err
	}
	if err = ensureColumns(db, "agent_sessions", memoryIndexCols); err != nil {
		db.Close()
		db = nil
		return nil, err
	}
	if err = ensureColumns(db, "graph_versions", memoryIndexCols); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	// 数据迁移：把旧的单条 prompts.insight 搬成 prompt_insights 的第一条
	if err = migratePromptInsightsFromLegacy(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	return db, nil
}

// tableColumns 返回某张表已有的列名集合（PRAGMA table_info）。
func tableColumns(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	existing := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		existing[name] = true
	}
	return existing, rows.Err()
}

// ensureColumns 给存量表补列：SQLite 的 CREATE TABLE IF NOT EXISTS 不会给已有表加列。
//
// ⚠ 新增列的 ALTER TABLE 必须走这里，而不是上面 statements 列表里的 CREATE INDEX——
// 列表在补列**之前**执行，对还没有该列的老库会直接报 "no such column"，导致库初始化失败、服务起不来。
func ensureColumns(db *sql.DB, table string, columns []struct{ name, ddl string }) error {
	existing, err := tableColumns(db, table)
	if err != nil {
		return err
	}
	for _, col := range columns {
		if existing[col.name] {
			continue
		}
		if _, err := db.Exec("ALTER TABLE " + table + " ADD COLUMN " + col.ddl); err != nil {
			return err
		}
	}
	return nil
}

// migratePromptsColumns 确保 prompts 表包含扩展列：
// metadata（灵犀式扩展信息）、is_locked（卡片锁定标记）、insight（使用心得）。
func migratePromptsColumns(db *sql.DB) error {
	return ensureColumns(db, "prompts", []struct{ name, ddl string }{
		{"metadata", "metadata TEXT"},
		{"is_locked", "is_locked INTEGER DEFAULT 0"},
		{"insight", "insight TEXT"},
	})
}

// migrateExperienceSharesColumns 确保 experience_shares 表包含分区/精华/点赞/吧 四列：
// section_id（贴吧式分区，NULL = 未归类）、is_featured（精华）、like_count（主楼赞）、
// board（所属吧，存量帖一律刷成 experience —— 加列那一刻还没有教程吧）。
func migrateExperienceSharesColumns(db *sql.DB) error {
	if err := ensureColumns(db, "experience_shares", []struct{ name, ddl string }{
		{"section_id", "section_id INTEGER"},
		{"is_featured", "is_featured INTEGER DEFAULT 0"},
		{"like_count", "like_count INTEGER DEFAULT 0"},
		{"board", "board TEXT DEFAULT 'experience'"},
		{"is_pinned", "is_pinned INTEGER DEFAULT 0"},
	}); err != nil {
		return err
	}
	// 存量刷平：SQLite 的 ADD COLUMN 带 DEFAULT 时读出来就是默认值，但手改库或
	// 早期插入可能留下 NULL/空串，统一收成 experience（幂等，每次启动跑一遍成本可忽略）。
	if _, err := db.Exec(`UPDATE experience_shares SET board = 'experience' WHERE board IS NULL OR board = ''`); err != nil {
		return err
	}
	// 索引必须排在补列之后，否则老库上 CREATE INDEX 会因 "no such column" 直接把服务卡死
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_experience_shares_section ON experience_shares(section_id)"); err != nil {
		return err
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_experience_shares_board ON experience_shares(board)"); err != nil {
		return err
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_experience_shares_pinned ON experience_shares(board, is_pinned)"); err != nil {
		return err
	}
	return nil
}

// migrateExperienceSectionsColumns 确保 experience_sections 表包含 board 列。
// 分区按吧隔离：教程吧和经验分享吧各自最多 15 个，允许同名。
func migrateExperienceSectionsColumns(db *sql.DB) error {
	if err := ensureColumns(db, "experience_sections", []struct{ name, ddl string }{
		{"board", "board TEXT DEFAULT 'experience'"},
	}); err != nil {
		return err
	}
	if _, err := db.Exec(`UPDATE experience_sections SET board = 'experience' WHERE board IS NULL OR board = ''`); err != nil {
		return err
	}
	if _, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_experience_sections_board ON experience_sections(board)"); err != nil {
		return err
	}
	return nil
}

// migratePromptInsightsFromLegacy 把旧的单条 prompts.insight 纯文本搬成 prompt_insights 里的第一条。
//
// 只处理「该卡片还没有任何心得行」的卡片；搬完立刻把旧列清空 ——
// 否则用户把心得全删掉之后，下次启动又会被这列搬回来（反复复活）。
// 迁移与清空在同一个事务里，避免中途失败造成「插了但没清」或反过来的半状态。
func migratePromptInsightsFromLegacy(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
INSERT INTO prompt_insights (prompt_id, content, rating, tags, created_at, updated_at)
SELECT p.id, TRIM(p.insight), 0, '[]',
       COALESCE(p.updated_at, CURRENT_TIMESTAMP),
       COALESCE(p.updated_at, CURRENT_TIMESTAMP)
FROM prompts p
WHERE p.insight IS NOT NULL
  AND TRIM(p.insight) <> ''
  AND NOT EXISTS (SELECT 1 FROM prompt_insights i WHERE i.prompt_id = p.id)`); err != nil {
		return err
	}

	// 已经是列表形态的卡片，旧列不再需要保留（避免删光心得后被搬回来）
	if _, err := tx.Exec(`
UPDATE prompts SET insight = NULL
WHERE insight IS NOT NULL AND TRIM(insight) <> ''
  AND EXISTS (SELECT 1 FROM prompt_insights i WHERE i.prompt_id = prompts.id)`); err != nil {
		return err
	}

	return tx.Commit()
}

// migrateBooksColumns 确保 books 表包含 target_words 等新列。
func migrateBooksColumns(db *sql.DB) error {
	return ensureColumns(db, "books", []struct{ name, ddl string }{
		{"target_words", "target_words INTEGER"},
	})
}

// migrateAPIModelsColumns 确保 api_models 表包含 category（模型分类，兼容字段）
// 以及多模态补齐新增的 config_type / 高级参数列。
//
// 首次补 config_type 时，把旧的 category 映射回填进 config_type（一次性，见 backfillConfigType）；
// 之后再启动不再回填，避免覆盖用户手动改过的类型。
func migrateAPIModelsColumns(db *sql.DB) error {
	existing, err := tableColumns(db, "api_models")
	if err != nil {
		return err
	}
	firstTimeConfigType := !existing["config_type"]

	if err := ensureColumns(db, "api_models", []struct{ name, ddl string }{
		{"category", "category TEXT DEFAULT ''"},
		{"config_type", "config_type TEXT NOT NULL DEFAULT 'chat'"},
		{"prompt", "prompt TEXT DEFAULT ''"},
		{"prompt_role", "prompt_role TEXT DEFAULT 'system'"},
		{"top_k", "top_k INTEGER"},
		{"presence_penalty", "presence_penalty REAL"},
		{"stop", "stop TEXT DEFAULT ''"},
		{"max_tokens_mode", "max_tokens_mode TEXT DEFAULT 'auto'"},
		// max_tokens_source 记录 max_tokens 这个值的**来源**，让界面能区分
		// 「官方给的」「探测出来的」「我们猜的」「用户填的」。
		// 空串 = 未知来源（存量数据），界面显示「未设置来源」而不是假装确定。
		{"max_tokens_source", "max_tokens_source TEXT DEFAULT ''"},
		{"thinking_mode", "thinking_mode INTEGER DEFAULT 0"},
		{"thinking_effort", "thinking_effort TEXT DEFAULT 'default'"},
		{"generation_config_source", "generation_config_source TEXT DEFAULT 'manual'"},
		{"request_rules", "request_rules TEXT"},
		{"param_groups", "param_groups TEXT DEFAULT ''"},
		{"default_params", "default_params TEXT DEFAULT ''"},
		{"is_free", "is_free INTEGER DEFAULT 0"},
		{"input_ratio", "input_ratio REAL"},
		{"output_ratio", "output_ratio REAL"},
	}); err != nil {
		return err
	}

	if firstTimeConfigType {
		if err := backfillConfigType(db); err != nil {
			return err
		}
	}
	return nil
}

// backfillConfigType 把存量模型的 category 映射进 config_type（只在首次加列时执行一次）。
//
// 映射口径与前端历史取值对齐：image / vision / video / embed 原样保留，
// 其余（chat / reason / multimodal / code / long / 空）一律归 chat。
func backfillConfigType(db *sql.DB) error {
	_, err := db.Exec(`
UPDATE api_models
SET config_type = CASE LOWER(TRIM(category))
    WHEN 'image'  THEN 'image'
    WHEN 'vision' THEN 'vision'
    WHEN 'video'  THEN 'video'
    WHEN 'embed'  THEN 'embed'
    ELSE 'chat'
END
WHERE category IS NOT NULL AND TRIM(category) <> ''`)
	return err
}

// migrateAPIProvidersColumns 确保 api_providers 表包含渠道 / 扩展认证 / 并发列。
//
// 渠道策略采用「复用 provider 作渠道载体」：provider 本就是「端点 + Key」语义，
// channel_name 即渠道名，concurrency 为该渠道并发上限，无需另建 api_channels 表。
func migrateAPIProvidersColumns(db *sql.DB) error {
	return ensureColumns(db, "api_providers", []struct{ name, ddl string }{
		{"auth_type", "auth_type TEXT NOT NULL DEFAULT 'api_key'"},
		{"access_key", "access_key TEXT"},
		{"secret_key", "secret_key TEXT"},
		{"group_id", "group_id TEXT"},
		{"endpoint_id", "endpoint_id TEXT"},
		{"concurrency", "concurrency INTEGER DEFAULT 1"},
		{"channel_name", "channel_name TEXT DEFAULT ''"},
		{"extra_config", "extra_config TEXT"},
	})
}

// migrateGraphEntitiesPrimaryColumn 确保 graph_entities 表包含 primary_id 列
// （实体合并系统的主/副词条标记；NULL = 主词条），并建好它的索引。
func migrateGraphEntitiesPrimaryColumn(db *sql.DB) error {
	if err := ensureColumns(db, "graph_entities", []struct{ name, ddl string }{
		{"primary_id", "primary_id INTEGER"},
	}); err != nil {
		return err
	}
	// 索引必须建在补列之后：老库在补列前建会报 no such column
	_, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_graph_entities_primary ON graph_entities(primary_id)")
	return err
}

// migrateRankSnapshotsColumns 确保 rank_snapshots 表包含 scope 列
// （page=当前页快照 / full=整榜快照），并建好按 scope 查询的索引。
func migrateRankSnapshotsColumns(db *sql.DB) error {
	if err := ensureColumns(db, "rank_snapshots", []struct{ name, ddl string }{
		{"scope", "scope TEXT NOT NULL DEFAULT 'page'"},
	}); err != nil {
		return err
	}
	// 索引必须建在补列之后
	_, err := db.Exec("CREATE INDEX IF NOT EXISTS idx_rank_snapshots_scope ON rank_snapshots(source, tab, scope, captured_date)")
	return err
}
