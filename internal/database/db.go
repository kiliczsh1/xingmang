package database

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func GetDB() *sql.DB {
	return db
}

func Close() {
	if db != nil {
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
  creator_name TEXT,
  version TEXT,
  metadata TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
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
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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
		`CREATE TABLE IF NOT EXISTS api_providers (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  provider_type TEXT NOT NULL,
  api_key TEXT,
  api_url TEXT NOT NULL,
  api_format TEXT DEFAULT 'openai',
  use_full_url INTEGER DEFAULT 0,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// api_models
		`CREATE TABLE IF NOT EXISTS api_models (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  provider_id INTEGER NOT NULL,
  name TEXT NOT NULL,
  model TEXT NOT NULL,
  temperature REAL DEFAULT 0.7,
  max_tokens INTEGER DEFAULT 2000,
  top_p REAL DEFAULT 0.9,
  frequency_penalty REAL DEFAULT 0.0,
  description TEXT DEFAULT '',
  is_default INTEGER DEFAULT 0,
  enabled INTEGER DEFAULT 1,
  sort_order INTEGER DEFAULT 0,
  category TEXT DEFAULT '',
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
		// workflows（工作流工作台）
		`CREATE TABLE IF NOT EXISTS workflows (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  description TEXT,
  scene TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// workflow_node_library（工作流节点模板库；对外键是 template_id，不是自增主键）
		`CREATE TABLE IF NOT EXISTS workflow_node_library (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  template_id TEXT NOT NULL UNIQUE,
  type TEXT NOT NULL,
  variant TEXT,
  category TEXT NOT NULL,
  label TEXT NOT NULL,
  description TEXT,
  outputs TEXT,
  generated INTEGER DEFAULT 0,
  template TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`,
		// Indexes
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

	// 迁移：为存量库的 api_models 表补充 category 列（模型分类）
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

	// 迁移：为存量库的 prompts 表补充 metadata 列（灵犀式扩展信息）
	if err = migratePromptsMetadataColumn(db); err != nil {
		db.Close()
		db = nil
		return nil, err
	}

	return db, nil
}

// migrateBooksColumns 确保 books 表包含 target_words 等新列（SQLite 的 CREATE IF NOT EXISTS 不会给已有表加列）。
func migrateBooksColumns(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(books)")
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	rows.Close()

	columns := []struct {
		name string
		ddl  string
	}{
		{"target_words", "target_words INTEGER"},
	}
	for _, col := range columns {
		if existing[col.name] {
			continue
		}
		if _, err := db.Exec("ALTER TABLE books ADD COLUMN " + col.ddl); err != nil {
			return err
		}
	}
	return nil
}

// migrateAPIModelsColumns 确保 api_models 表包含 category 列（模型分类）。
func migrateAPIModelsColumns(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(api_models)")
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	rows.Close()

	columns := []struct {
		name string
		ddl  string
	}{
		{"category", "category TEXT DEFAULT ''"},
	}
	for _, col := range columns {
		if existing[col.name] {
			continue
		}
		if _, err := db.Exec("ALTER TABLE api_models ADD COLUMN " + col.ddl); err != nil {
			return err
		}
	}
	return nil
}

// migrateGraphEntitiesPrimaryColumn 确保 graph_entities 表包含 primary_id 列
// （实体合并系统的主/副词条标记；NULL = 主词条），并建好它的索引。
//
// 索引刻意放在这里而不是上面的 statements 列表里：列表里的语句在补列**之前**执行，
// 对还没有该列的老库会直接报 "no such column: primary_id"，导致库初始化失败、服务起不来。
func migrateGraphEntitiesPrimaryColumn(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(graph_entities)")
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	rows.Close()

	if !existing["primary_id"] {
		if _, err = db.Exec("ALTER TABLE graph_entities ADD COLUMN primary_id INTEGER"); err != nil {
			return err
		}
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_graph_entities_primary ON graph_entities(primary_id)")
	return err
}

// migratePromptsMetadataColumn 确保 prompts 表包含 metadata 列（灵犀式扩展信息）。
func migratePromptsMetadataColumn(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(prompts)")
	if err != nil {
		return err
	}
	existing := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		existing[name] = true
	}
	rows.Close()

	if existing["metadata"] {
		return nil
	}
	_, err = db.Exec("ALTER TABLE prompts ADD COLUMN metadata TEXT")
	return err
}
