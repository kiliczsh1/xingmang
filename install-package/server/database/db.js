const fs = require('fs');
const path = require('path');

const isPortablePackage = () => {
  if (process.pkg) return true;
  return process.execPath.includes('XingNovel-node');
};

const getRootDir = () => {
  if (isPortablePackage()) {
    return path.dirname(process.execPath);
  }
  return path.join(__dirname, '..', '..');
};

const loadBetterSqlite3 = () => {
  if (process.pkg) {
    const portableModulePath = path.join(
      path.dirname(process.execPath),
      'server',
      'node_modules',
      'better-sqlite3'
    );
    if (fs.existsSync(portableModulePath)) {
      return require(portableModulePath);
    }
    const runtimeModulePath = path.join(
      path.dirname(process.execPath),
      'runtime',
      'node_modules',
      'better-sqlite3'
    );
    if (fs.existsSync(runtimeModulePath)) {
      return require(runtimeModulePath);
    }
    throw new Error(`Packaged runtime dependency not found: better-sqlite3`);
  }

  return require('better-sqlite3');
};
const Database = loadBetterSqlite3();

// 获取数据库路径（兼容 pkg 打包环境和便携式打包环境）
const getDbPath = () => {
  // 如果环境变量指定了 DB_PATH，优先使用
  if (process.env.DB_PATH) {
    return process.env.DB_PATH;
  }
  // 便携式打包环境：使用可执行文件所在目录下的 server/database 子目录
  if (isPortablePackage()) {
    return path.join(getRootDir(), 'server', 'database', 'novel.db');
  }
  // 开发环境使用当前目录
  return path.join(__dirname, 'novel.db');
};

// 创建或打开数据库
const db = new Database(getDbPath(), { verbose: console.log });

// 创建表结构
const initDatabase = () => {
  // 书本表
  db.exec(`
    CREATE TABLE IF NOT EXISTS books (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      title TEXT NOT NULL,
      description TEXT,
      cover TEXT,
      author TEXT,
      category TEXT,
      tags TEXT,
      status TEXT DEFAULT 'draft',
      is_public INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // 章节表
  db.exec(`
    CREATE TABLE IF NOT EXISTS chapters (
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
    )
  `);

  // 提示词表
  db.exec(`
    CREATE TABLE IF NOT EXISTS prompts (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT,
      content TEXT NOT NULL,
      category TEXT,
      order_num INTEGER DEFAULT 0,
      fields TEXT,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // 备忘录表（全局共享）
  db.exec(`
    CREATE TABLE IF NOT EXISTS memos (
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
    )
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS experience_shares (
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
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // 分卷表
  db.exec(`
    CREATE TABLE IF NOT EXISTS volumes (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER NOT NULL,
      title TEXT NOT NULL,
      order_num INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
    )
  `);

  // API 服务商表
  db.exec(`
    CREATE TABLE IF NOT EXISTS api_providers (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      provider_type TEXT NOT NULL,
      api_key TEXT,
      api_url TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // API模型表
  db.exec(`
    CREATE TABLE IF NOT EXISTS api_models (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      provider_id INTEGER NOT NULL,
      name TEXT NOT NULL,
      model TEXT NOT NULL,
      temperature REAL DEFAULT 0.7,
      max_tokens INTEGER DEFAULT 2000,
      is_default INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (provider_id) REFERENCES api_providers(id) ON DELETE CASCADE
    )
  `);

  // 迁移旧的api_configs数据到新表结构
  const migrateOldConfigs = () => {
    try {
      // 检查是否存在旧表
      const tableExists = db.prepare("SELECT name FROM sqlite_master WHERE type='table' AND name='api_configs'").get();
      
      if (tableExists) {
        // 获取旧配置
        const oldConfigs = db.prepare('SELECT * FROM api_configs').all();
        
        if (oldConfigs.length > 0) {
          console.log(`发现${oldConfigs.length}条旧配置，开始迁移...`);
          
          // 为每个旧配置创建服务商和模型
          const insertProvider = db.prepare('INSERT INTO api_providers (name, provider_type, api_key, api_url) VALUES (?, ?, ?, ?)');
          const insertModel = db.prepare('INSERT INTO api_models (provider_id, name, model, temperature, max_tokens, is_default) VALUES (?, ?, ?, ?, ?, ?)');
          
          for (const config of oldConfigs) {
            // 创建服务商
            const providerResult = insertProvider.run(
              `${config.name}-服务商`,
              config.provider,
              config.api_key,
              config.api_url
            );
            
            // 创建模型
            insertModel.run(
              providerResult.lastInsertRowid,
              config.name,
              config.model,
              config.temperature,
              config.max_tokens,
              config.is_default
            );
          }
          
          console.log('配置迁移完成');
          
          // 重命名旧表作为备份
          db.exec('ALTER TABLE api_configs RENAME TO api_configs_backup');
          console.log('旧配置表已重命名为api_configs_backup');
        }
      }
    } catch (error) {
      console.log('配置迁移过程:', error.message);
    }
  };
  
  migrateOldConfigs();

  // 迁移books表，添加新字段
  const migrateBooksTable = () => {
    try {
      const columns = db.prepare("PRAGMA table_info(books)").all();
      const columnNames = columns.map(col => col.name);
      
      const newColumns = [
        { name: 'author', sql: 'ALTER TABLE books ADD COLUMN author TEXT' },
        { name: 'category', sql: 'ALTER TABLE books ADD COLUMN category TEXT' },
        { name: 'tags', sql: 'ALTER TABLE books ADD COLUMN tags TEXT' },
        { name: 'status', sql: 'ALTER TABLE books ADD COLUMN status TEXT DEFAULT "draft"' },
        { name: 'is_public', sql: 'ALTER TABLE books ADD COLUMN is_public INTEGER DEFAULT 0' }
      ];
      
      for (const col of newColumns) {
        if (!columnNames.includes(col.name)) {
          db.exec(col.sql);
          console.log(`已添加books表字段: ${col.name}`);
        }
      }
    } catch (error) {
      console.log('books表迁移过程:', error.message);
    }
  };
  
  migrateBooksTable();

  const migrateChaptersTable = () => {
    try {
      const columns = db.prepare("PRAGMA table_info(chapters)").all();
      const columnNames = columns.map(col => col.name);

      const newColumns = [
        { name: 'summary', sql: "ALTER TABLE chapters ADD COLUMN summary TEXT DEFAULT ''" },
        { name: 'volume_id', sql: 'ALTER TABLE chapters ADD COLUMN volume_id INTEGER' }
      ];

      for (const col of newColumns) {
        if (!columnNames.includes(col.name)) {
          db.exec(col.sql);
          console.log(`chapters琛ㄦ坊鍔犲瓧娈? ${col.name}`);
        }
      }
    } catch (error) {
      console.log('chapters琛ㄨ縼绉昏繃绋?', error.message);
    }
  };

  migrateChaptersTable();

  // 迁移prompts表，添加fields字段
  const migratePromptsTable = () => {
    try {
      const columns = db.prepare("PRAGMA table_info(prompts)").all();
      const columnNames = columns.map(col => col.name);
      
      if (!columnNames.includes('fields')) {
        db.exec('ALTER TABLE prompts ADD COLUMN fields TEXT');
        console.log('已添加prompts表字段: fields');
      }
      if (!columnNames.includes('subcategories')) {
        db.exec('ALTER TABLE prompts ADD COLUMN subcategories TEXT');
        console.log('已添加prompts表字段: subcategories');
      }
      if (!columnNames.includes('description')) {
        db.exec('ALTER TABLE prompts ADD COLUMN description TEXT');
        console.log('已添加prompts表字段: description');
      }
      if (!columnNames.includes('card_type')) {
        db.exec("ALTER TABLE prompts ADD COLUMN card_type TEXT DEFAULT 'normal'");
        console.log('已添加prompts表字段: card_type');
      }
      if (!columnNames.includes('password')) {
        db.exec('ALTER TABLE prompts ADD COLUMN password TEXT');
        console.log('已添加prompts表字段: password');
      }
    } catch (error) {
      console.log('prompts表迁移过程:', error.message);
    }
  };
  
  migratePromptsTable();

  const migrateExperienceSharesTable = () => {
    try {
      const columns = db.prepare("PRAGMA table_info(experience_shares)").all();
      const columnNames = columns.map(col => col.name);

      const newColumns = [
        { name: 'content_render_mode', sql: "ALTER TABLE experience_shares ADD COLUMN content_render_mode TEXT DEFAULT 'markdown'" },
        { name: 'cover_url', sql: 'ALTER TABLE experience_shares ADD COLUMN cover_url TEXT' },
        { name: 'pdf_file_url', sql: 'ALTER TABLE experience_shares ADD COLUMN pdf_file_url TEXT' },
        { name: 'pdf_file_name', sql: 'ALTER TABLE experience_shares ADD COLUMN pdf_file_name TEXT' },
        { name: 'pdf_file_size', sql: 'ALTER TABLE experience_shares ADD COLUMN pdf_file_size INTEGER' },
        { name: 'create_type', sql: "ALTER TABLE experience_shares ADD COLUMN create_type TEXT DEFAULT 'manual'" },
        { name: 'author_id', sql: 'ALTER TABLE experience_shares ADD COLUMN author_id INTEGER' },
        { name: 'author_name', sql: "ALTER TABLE experience_shares ADD COLUMN author_name TEXT DEFAULT '星芒用户'" },
        { name: 'status', sql: "ALTER TABLE experience_shares ADD COLUMN status TEXT DEFAULT 'published'" },
        { name: 'pdf_parse_status', sql: 'ALTER TABLE experience_shares ADD COLUMN pdf_parse_status TEXT' },
        { name: 'pdf_parse_result', sql: 'ALTER TABLE experience_shares ADD COLUMN pdf_parse_result TEXT' },
        { name: 'source_file_name', sql: 'ALTER TABLE experience_shares ADD COLUMN source_file_name TEXT' }
      ];

      for (const col of newColumns) {
        if (!columnNames.includes(col.name)) {
          db.exec(col.sql);
          console.log(`宸叉坊鍔爀xperience_shares琛ㄥ瓧娈? ${col.name}`);
        }
      }
    } catch (error) {
      console.log('experience_shares琛ㄨ縼绉昏繃绋?', error.message);
    }
  };

  migrateExperienceSharesTable();

  // 迁移 memos 表，添加新字段
  const migrateMemosTable = () => {
    try {
      const columns = db.prepare("PRAGMA table_info(memos)").all();
      const columnNames = columns.map(col => col.name);
      
      const newColumns = [
        { name: 'tags', sql: "ALTER TABLE memos ADD COLUMN tags TEXT DEFAULT '[]'" },
        { name: 'is_pinned', sql: 'ALTER TABLE memos ADD COLUMN is_pinned INTEGER DEFAULT 0' },
        { name: 'word_count', sql: 'ALTER TABLE memos ADD COLUMN word_count INTEGER DEFAULT 0' }
      ];
      
      for (const col of newColumns) {
        if (!columnNames.includes(col.name)) {
          db.exec(col.sql);
          console.log(`已添加 memos 表字段：${col.name}`);
        }
      }
    } catch (error) {
      console.log('memos 表迁移过程:', error.message);
    }
  };
  
  migrateMemosTable();

  // 对话表（会话）
  db.exec(`
    CREATE TABLE IF NOT EXISTS conversations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER NOT NULL,
      title TEXT DEFAULT '新对话',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
    )
  `);

  // 对话消息表
  db.exec(`
    CREATE TABLE IF NOT EXISTS conversation_messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      conversation_id INTEGER NOT NULL,
      role TEXT NOT NULL,
      content TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
    )
  `);

  // 生成器表
  db.exec(`
    CREATE TABLE IF NOT EXISTS generators (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT,
      icon TEXT DEFAULT 'Lightning',
      core_prompt TEXT,
      remark TEXT,
      order_num INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  // 角色表
  db.exec(`
    CREATE TABLE IF NOT EXISTS characters (
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
    )
  `);

  // 使用统计表
  db.exec(`
    CREATE TABLE IF NOT EXISTS usage_stats (
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
    )
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS workflows (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      description TEXT,
      scene TEXT NOT NULL,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS workflow_node_library (
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
    )
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS graph_entities (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER NOT NULL,
      name TEXT NOT NULL,
      type TEXT NOT NULL DEFAULT 'character',
      description TEXT DEFAULT '',
      metadata TEXT DEFAULT '{}',
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
    )
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS graph_relations (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER NOT NULL,
      source_id INTEGER NOT NULL,
      target_id INTEGER NOT NULL,
      relation_type TEXT NOT NULL DEFAULT 'related',
      description TEXT DEFAULT '',
      evidence TEXT DEFAULT '',
      confidence REAL DEFAULT 1.0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
      FOREIGN KEY (source_id) REFERENCES graph_entities(id) ON DELETE CASCADE,
      FOREIGN KEY (target_id) REFERENCES graph_entities(id) ON DELETE CASCADE
    )
  `);

  db.exec(`
    CREATE TABLE IF NOT EXISTS graph_versions (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      book_id INTEGER NOT NULL,
      name TEXT NOT NULL,
      description TEXT DEFAULT '',
      scope TEXT DEFAULT 'all',
      chapter_count INTEGER DEFAULT 0,
      entity_count INTEGER DEFAULT 0,
      relation_count INTEGER DEFAULT 0,
      created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
      FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE
    )
  `);

  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_graph_entities_book ON graph_entities(book_id)
  `);
  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_graph_relations_book ON graph_relations(book_id)
  `);
  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_graph_relations_source ON graph_relations(source_id)
  `);
  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_graph_relations_target ON graph_relations(target_id)
  `);
  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_graph_versions_book ON graph_versions(book_id)
  `);

  // 迁移：为 graph_entities 和 graph_relations 添加 version_id 字段
  try {
    const entityColumns = db.prepare("PRAGMA table_info(graph_entities)").all();
    const hasEntityVersionId = entityColumns.some(col => col.name === 'version_id');
    if (!hasEntityVersionId) {
      db.exec(`ALTER TABLE graph_entities ADD COLUMN version_id INTEGER REFERENCES graph_versions(id) ON DELETE CASCADE`);
      db.exec(`CREATE INDEX IF NOT EXISTS idx_graph_entities_version ON graph_entities(version_id)`);
      console.log('已为 graph_entities 添加 version_id 字段');
    }

    const relationColumns = db.prepare("PRAGMA table_info(graph_relations)").all();
    const hasRelationVersionId = relationColumns.some(col => col.name === 'version_id');
    if (!hasRelationVersionId) {
      db.exec(`ALTER TABLE graph_relations ADD COLUMN version_id INTEGER REFERENCES graph_versions(id) ON DELETE CASCADE`);
      db.exec(`CREATE INDEX IF NOT EXISTS idx_graph_relations_version ON graph_relations(version_id)`);
      console.log('已为 graph_relations 添加 version_id 字段');
    }
  } catch (err) {
    console.log('迁移 version_id 字段时出错（可能已存在）:', err.message);
  }

  // 创建索引
  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_usage_stats_date ON usage_stats(date)
  `);
  db.exec(`
    CREATE INDEX IF NOT EXISTS idx_usage_stats_model ON usage_stats(model_id)
  `);

  console.log('数据库初始化完成');
};

// 导入默认提示词
const importDefaultPrompts = () => {
  try {
    // 检查是否已经导入过默认提示词
    const checkPrompt = db.prepare('SELECT id FROM prompts WHERE name = ? AND category = ?');
    const existingPrompt = checkPrompt.get('写作要求', '写作要求');
    
    if (existingPrompt) {
      console.log('默认提示词已存在，跳过导入');
      return;
    }

    console.log('正在导入默认提示词...');
    
    // 默认提示词数据
    const defaultPrompts = [
      {
        name: '写作要求',
        content: '你是一位爆款网文作家，你将写一篇符合以下要求的小说正文：（以下不必输出）\n题材：${题材}\n根据所给细纲及要求，参考对应爆款小说特征，按点生成几个提示词：\n作品类型：\n读者受众：\n角色动机：\n语言风格：\n逻辑：\n读者情感曲线：\n信息密度：\n节奏强度：\n_\n语言要求：口语化，语言精炼自然无 ai 味、具体、生活化、情感化的表达和细节，避免技术感、抽象化或机械化，砍掉虚词、用动词代替名词化表达不堆砌词藻，不无意义描写，根据细纲合理增加环境，细节描写等，非剧情必要、有益少用甚至不用比喻等修辞手法，可适当网络用语\n符号要求使用全角和""\n场景要求（为人物、主旨服务）：\n场景与主旨联动：主题映射、情绪容器、冲突具象化\n角色与场景渗透：感官滤镜、主观投射、互动触发\n动态感官描写：光影流动性、声音层次感\n空间叙事设计：空间压迫感、路径暗示性\n群像场景互动：环境反应链、群体行为模式\n禁止使用任何情感词汇（如悲伤、愤怒），仅用环境和动作描写表现人物心情、心理状态\n人物刻画满足：\n\n动作刻画\n习惯性动作。\n情绪外化：用动作替代情绪描述（摔门表愤怒）\n语言塑造\n个性化表达：符合角色的口头禅、特定语气词。\n节奏差异：语速快慢、句式长短对应性格。\n话中有话：对话隐含潜台词或矛盾心理。\n心理描写：\n矛盾挣扎：内心决策过程（如伸手又缩回）\n瞬间直觉：潜意识念头暴露本性。\n空间叙事设计：镜头焦距切换（特写/远景）\n群体衬托：周围人的态度和行为侧面烘托形象。\n表里反差：外在表现与内在真实对比。\n展示非标签：用具体行为替代性格直述。\n隐去过度解释：借环境/动作暗示心理状态。\n情感：\n禁止使用「感到/觉得/似乎」类陈述句\n用可感知的生理反应、物体状态或动作细节替代抽象情绪词汇\n将情感映射到场景元素中，使环境与人物心理相互映照\n通过动作节奏变化表现情绪层次，展现情感流动过程\n通过反常行为强化张力，制造心理冲击\n用他人反应反衬主角情绪，突出个体心理状态\n情绪调动：\n1. 感官触发：视觉 + 听觉触觉 + 嗅觉→沉浸感\n2. 转化：抽象情绪→具象行为\n3. 节奏控制：\n紧张：短句 + 重复 + 分段\n舒缓：长句 + 环境细节\n4. 反差刺激：铺垫（常态）→ 突转（细节破坏）→ 爆发\n核心原则：不直述情绪用语言/行为/环境/细节暗示\n对话突出人物性格，体现动机、弧光，符合角色公式，占比 30%\n语言节奏感强，不用以下词语：值得注意的是、综上所、述本质上、夜色如墨、无意识摩挲着、无意义数字、挑眉、融入夜色中、低声重复、不易察觉、金属摩擦般的质感、一丝、有些、好像、仿佛等明喻标志词、说、道、铁锈、锈迹斑斑、血腥味、勾起嘴角、坚定、知道、明白。\n结尾设计钩子，不得用总结性，展望性语言如：他明白、知道等，用悬念、钩子结尾\n氛围感=环境渲染 + 情绪基调 + 沉浸体验\n悬疑感=未知 + 心理压迫\n共情=情感共鸣 + 心理投射 + 立场同步\n信息差=信息不对称 + 认知博弈 + 张力构建\n期待感=目标预设 + 悬念堆叠 + 延迟满足\n高冲突=对抗密度 + 利益碰撞 + 高速推进\n爽感=情绪宣泄 + 目标达成 + 正向反馈\n虐感=情感剥夺 + 命运碾压 + 价值拷问\n代入感=视角绑定 + 情境还原 + 感知同步\n角色动机=需求内核 + 外部刺激 + 行为触发器\n角色弧光=初始状态 + 冲突催化 + 转变终点\n伏笔回收=隐性暗示 + 因果闭环 + 情绪爆发\n反转=预期误导 + 证据重组\n情感拉扯=双向矛盾 + 价值对冲 + 延迟释放\n群体=个体差异 + 动态平衡\n压迫感=权力落差 + 行动限制\n冷硬派=细节聚焦 + 情感抑制 + 行动主导\n浸入式=五感渗透 + 心理独白\n戏剧性=通过角色、情境或价值观的对立制造紧张感 + 情节突转或揭示隐藏信息，打破观众预期 + 事件与情感在短时间内密集爆发，强化观演吸引力。\n情感冲击力=角色遭遇极端情境引发观众强烈代入 + 结合视觉或表演的极致表达放大情感效果。\n冲击力 =通过突然的视觉、听觉或叙事刺激产生生理/心理震撼 + 事件结束后仍持续影响观众认知或情绪的状态。\n不定时投放爽点 + 每 3-5 章设置情绪高点 + 悬念 - 释放 - 期待的闭环结构\n通过文笔调节情绪传递效率（如短句加速紧张感）+ 利用矛盾冲突制造情绪漩涡 + 平稳叙事与激烈冲突的交替节奏\n三阶段构建：明确目标 - 即时反馈 - 挑战匹配（越级但合理）\n日常/温馨情节调节叙事张力 +避免情绪过载导致麻木 + 压抑后及时给予正向反馈\n角色行动引发连锁反应 + 最决策符合利益逻辑 + 动态平衡延长悬念\n减少冗余描写维持叙事流速\n精准定位共情点 + 铺垫时长控制情绪强度 + 多情感维度叠加释放 + 根据题材调整虐度/甜度配比\n边际效用递减=减少重复套路 + 用新类型爆点刺激阈值重置\n认知失调=信息差制造 + 预期违背 + 逻辑自洽\n情感账户=储蓄（正向互动）/ 透支（伤害）/ 兑现（高潮）\n社会认同=群体压力/从众效应/权威背书\n损失厌恶=强调失去成本 + 制造稀缺性\n心流体验=技能挑战平衡 + 即时反馈 + 掌控感\n叙事经济学=信息释放节奏 + 情绪价值密度 + 注意力分配\n集体无意识=原型意象 + 文化符号 + 普世情感\n神经叙事=感官刺激序列 + 多巴胺触发点 + 镜像神经元激活',
        category: '写作要求',
        order_num: 0,
        fields: JSON.stringify([
          {
            name: '题材',
            label: '题材',
            type: 'text',
            options: [],
            description: '',
            required: true
          }
        ])
      },
      {
        name: '爆款知乎风，一次 1w，生成停止请使用续写继续写',
        content: '角色：知乎风格万字短篇故事生成专家\n核心任务：\n根据用户提供的核心剧情概述，创作一篇 约 10,000 字 的、完整的、具有鲜明知乎风格的短篇小说正文。\n\n关键指令与多维度执行要求：\n1. 核心风格与基调 (Overall Style & Tone):\n* 严格遵循知乎风格：故事必须充满知乎网文的典型特征——强代入感、快节奏、重情绪、情节驱动、语言直白。\n* 第一人称叙事 (Narrative Voice): 全文必须使用 第一人称"我" 进行叙述，确保读者能迅速代入主角视角。\n* 情绪核心明确 (Emotional Core): 从剧情概述中提炼核心情绪（例如：爽文 - 复仇逆袭、虐文 - 意难平/悲剧、甜宠 - 高糖互动、悬疑 - 紧张探索），并让整个故事紧密围绕这个核心情绪展开和波动。\n\n2. 结构与节奏控制 (Structure & Pacing):\n* 黄金开头原则 (Hook Principle): 即使是长篇幅，开头也必须迅速切入核心冲突或悬念，牢牢抓住读者（可参考之前的"黄金开篇"指令思路，但要自然融入正文）。\n* 经典叙事结构 (Narrative Arc): 遵循"开端（引入冲突）-> 发展（冲突升级与多轮交锋）-> 高潮（核心矛盾总爆发与解决）-> 结局（尘埃落定与情绪收束）"的清晰结构。\n* 快节奏推进 (Fast Pacing): 避免冗长的描写和内心戏。情节推进要快，尤其是在发展阶段，要不断设置小冲突、小高潮、小反转，保持读者的阅读兴趣。用对话和行动主要驱动情节。\n* 张弛有度 (Pacing Variation): 在整体快节奏中，允许在高潮前后或关键情感节点有适当的情绪渲染或短暂舒缓，但整体不拖沓。对于 10,000 字的篇幅，意味着需要多个上升和解决的迷你弧线。\n\n3. 情节与冲突设计 (Plot & Conflict):\n* 情节密度 (Plot Density): 在 10,000 字内填充足够的情节量。围绕主线，可以设计 1-2 条副线（如次要人物关系、主角的次要目标）来丰富故事，但必须服务于主线和核心情绪。\n* 冲突是引擎 (Conflict as Engine): 冲突必须贯穿始终。设计清晰的 主角 vs 反派/困境 的对抗线。反派/困境的行动要有逻辑，能持续给主角制造压力和麻烦。\n* "爽点"/"虐点"/"糖点" 精准投放 (Emotional Payoffs): 根据故事的核心情绪，在关键节点精心设计并明确、有力地展现"爽点"（如打脸、复仇成功、逆袭）、"虐点"（如误会加深、生离死别、悲惨遭遇）或"糖点"（如甜蜜互动、浪漫时刻），给予读者强烈的情绪反馈。这是知乎风格的关键。\n* 反转运用 (Use of Twists): 可在故事中段或高潮部分设计 1-2 个关键反转，增加故事的意外性和吸引力，但反转需合乎情理，有铺垫。\n\n4. 人物塑造 (Characterization):\n* 聚焦主角 (Protagonist Focus): 深度挖掘主角的动机、情感变化和成长（或沉沦）。第一人称视角下，主角的心理活动和感受是重点，但要通过事件和反应来展现，而非大段枯燥的内心独白。\n* 典型化与标签化 (Archetypes & Labels): 配角（如反派、盟友、恋人）可以带有一定的典型标签（如"恶毒女配"、"深情男二"、"极品亲戚"）以便快速建立形象，但需赋予其符合自身立场的行为逻辑。\n* 行动与对话塑造 (Show, Don\'t Just Tell): 通过角色的具体行动、选择和对话来展现其性格、能力和关系，而不是直接描述。\n\n5. 语言风格 (Language Style):\n* 直白简洁 (Direct & Concise): 使用现代、口语化的书面语，避免过于华丽、复杂的词藻和句式。句子偏短，表达清晰。\n* 情绪饱满 (Emotionally Charged): 语言要能准确传达角色的强烈情绪，即使在平淡叙述时，也要蕴含张力。\n* 网络感适度 (Internet Slang - Moderate Use): 可以适度融入一些符合语境和角色身份的网络词汇或表达方式，增加亲切感，但避免滥用和过时。\n\n6. 生成过程要求 (Generation Process):\n* 连贯性优先 (Coherence First): 尽最大努力保持全文的逻辑连贯和情节一致性。\n* 细节填充 (Detail Filling): 基于概述，大胆地、创造性地填充必要的场景描写、动作细节、对话内容，使故事生动具体。\n* 篇幅控制 (Word Count Management): 尽量将总字数控制在 10,000 字左右。如果模',
        category: '写作要求',
        order_num: 1,
        fields: null
      },
      {
        name: '女频写作 - 情感共鸣',
        content: '角色扮演\n你是一位深谙女性心理和阅读爽点的网络小说大神级作者，尤其精通【番茄】和【飞卢】女频的爆款写作风格。你擅长营造强烈的情感共鸣、设计精彩的虐渣打脸情节、塑造极致人设（如马甲女王、霸道总裁、可爱萌宝），并能精准把握快节奏与情绪铺垫的平衡。\n\n任务\n根据我提供的核心【剧情点】，创作一段引人入胜、符合女频读者口味的小说章节或场景片段。\n\n风格与要求 (请根据需要调整括号内的侧重点)\n核心驱动：优先侧重于【 (选择一个或多个，可融合：A. 强烈的情感冲突与共鸣 (如虐恋、冤屈、背叛) / B. 酣畅淋漓的虐渣打脸与逆袭 / C. 女主自身强大魅力/能力的展现 (含马甲、异能、才华) / D. 极致甜宠或守护型浪漫关系) 】。\n开篇/引入：若是章节开头，必须迅速切入核心【情绪点】或【冲突点】，立刻抓住读者注意力，营造【 (选择一个：A. 强烈的代入感 (如感同身受的委屈、愤怒) / B. 对后续剧情的期待 (如复仇、揭秘) / C. 对人物关系的兴趣 (如宿命般的相遇)) 】。\n节奏：整体保持【快节奏】，但要在关键情节（尤其是打脸、情感爆发前）进行适当的【情绪铺垫】，让爽点或泪点更具冲击力。精准设置【悬念或高潮点】（尤其是在段落末尾或模拟广告位前），刺激追读欲望。\n爽点/情绪点:\n(若侧重 A/B - 情感/虐渣): 重点描绘女主的【内心挣扎、委屈、愤怒或决心】，以及【虐渣过程的细节】和反派的【后悔、震惊、凄惨下场】。强调正义得到伸张或冤屈得以洗刷的【解气感】。\n(若侧重 C - 女主强大): 清晰展现女主的【高光时刻】，如【隐藏技能/马甲曝光】时的惊艳四座、【运用智慧/能力解决危机】时的从容不迫。通过【他人（尤其是质疑者）的反应】来放大效果。\n(若侧重 D - 浪漫关系): 细腻描绘【男女主之间的互动】，可以是【霸道总裁式的强势宠溺】、【忠犬式的默默守护】或是【双强之间的极限拉扯】。突出【甜蜜细节】（发糖）或【为爱不顾一切】的牺牲感。\n人物:\n女主：形象需【鲜明有记忆点】，可以开局【美强惨】但【内心坚韧、有仇必报】，或是一开始就【智商在线、能力超群】。重点突出其【情感动机】（爱、恨、守护、责任）。\n男主：人设需【满足女性幻想】，如【颜值天花板、权势滔天、专一深情、洁身自好】（即使是霸总也要有其独特的魅力点）。他通常是女主【最坚实的后盾】或【甜蜜的负担】。\n配角/反派：【工具人属性】可以明显，如【降智的恶毒女配/极品亲戚/渣男】，他们的存在主要为了【制造冲突】和【被女主打脸】。若是【萌宝】，则要突出其【可爱、聪慧、神助攻】的特点。\n运用【标签化人设】（例："宠妻狂魔"、"鉴婊达人"、"马甲大佬"、"小作精"）增强辨识度。\n语言：使用【通俗易懂、感染力强】的语言，【情绪渲染】要到位。多用【对话】推进剧情和展现人物性格。适当运用【心理描写】增强代入感。文风可以是【轻松沙雕】、【正剧爽文】或【虐中带甜】等，但要保持易读性。\n(可选) 特殊元素：若涉及【系统、空间、异能、重生、穿越、年代】等元素，需自然融入，并服务于【女主成长、虐渣或感情线】。\n结尾：章节或段落结尾务必留下【强力钩子（Hook）】，可以是【未解的谜团】、【即将爆发的冲突】、【感情的重大转折】或是【女主身份即将暴露的危机】。\n输出长度：产出大约【 (指定大致字数，例如：1000-1800 字) 】的内容\n输出文笔要求：\n1、语言精简干练、长短句成段，不超过两个句号为一段，优化移动端阅读体验\n2、穿插对话，对话不得少于 30%\n3、强去 AI 味道，禁用词汇="他相信"、"他知道"、"一丝"、"坚定"，"不容置疑的威严"\n4、少用形容词、转折词、副词="非常"、"很"、"虽然"、"但是"，多用动词，提升画面感\n5、禁止使用上帝视角，代入角色沉浸式叙事\n6、禁止出现出现展望性、期望性等结尾，如"迎接挑战"，"他相信会"',
        category: '写作要求',
        order_num: 2,
        fields: null
      },
      {
        name: '爆款写作 -【番茄风】',
        content: '【身份】\n你是头部网文作者，擅长用动作与细节推动叙事，语言克制锋利，信息高密度，阅读顺滑。\n\n【创作目标】\n\n将【写作剧情】扩展为完整正文，篇幅自适应；如信息充足，建议≥3000 中文字符；如信息稀薄，以完整性与可读性优先\n所有叙事、节奏与镜头切换均以剧情点内在逻辑为唯一锚点，不强行制造冲突或钩子\n【叙事规则（剧情点优先，非模板化）】\n\n锚定：所有桥段、对白与细节均服务于剧情点的因果与情绪核，不引入对剧情点无关的大设定\n视角：限当前角色可感知的信息；不越界解释他人内心，不用上帝视角\n因果闭环：优先形成"动机 - 行动 - 反馈"的最小闭环；若剧情点要求悬而未决，保留合理缺口\n补全策略（最小必要）：\n只补"动机/过渡/场景细节/人物微反应/必要线索"\n系统/金手指仅在剧情点出现或隐含时呈现，且每次触发产生"可见变化"（数值/感知/资源/代价）\n设定露出：以"使用场景"带出设定，不做长篇说明\n【对话要求（默认高权重，可随剧情点自适应）】\n\n职能：对白须承载信息、推进关系或矛盾、暴露人物目标/立场，不做空转\n比例：默认对话≥40%，若剧情点信息不足则以信息密度优先\n形态：短句为主，一句一义；打断、停顿、回避式回答可体现潜台词\n人物差异：词汇选择、语气节奏、信息隐瞒程度，体现人物身份与处境\n动作夹对白：呼吸、手部动作、视线移动、器物声，穿插在对白前后，避免口水化堆叠\n【内置大神文笔库（写作习惯与技巧包）】\n\n动词驱动：少抽象形容，多具体动作（例：扣、拧、掀、压、拨、倚、抹、掐、敲、攥、撕）\n五感抓手：声音/温度/触感/湿度/气味/光线/重量（例：门轴轻响、杯壁发凉、空气有潮味、霓虹打在地面）\n镜头语言：近景→特写→环境反打，重点处用停顿或省略，避免空镜头描述\n潜台词构建：答非所问、沉默、迟疑词、重复关键词、被打断的句子，优先胜过"心理解释"\n道具锚点：让一个小物反复入镜，承载情绪或信息（指尖油渍、笔帽啮痕、屏幕裂纹、未关的提示音）\n比喻控量：每段不超过一次，取材于现场可见物，拒绝浮夸比喻\n情绪落地：用生理与环境反馈替代评价词（喉结滚动、气息发紧、鞋底摩擦声变重）\n语言克制：删去能删的修饰与连接词，句群形成节奏感（快 - 慢 - 快）\n【用词与禁用项】\n\n通俗、口语化，拒绝模板化起承转合与口号式宣言\n少形容词与转折词（如"非常/很/虽然/但是"等）\n禁用套话：不得出现"他相信""他知道""一丝""坚定""不容置疑的""嘴角的弧度""难以置信的"等 AI 腔\n术语与专名：完全沿用剧情点中的称呼\n【排版与输出】\n\n移动端友好：优先短段，尽量每段不超过两个句号\n仅输出正文，不复述剧情点，不解释写法，不加小标题\n结尾不做展望式总结；如留钩子，必须源自剧情点的内在因果\n【自检（你用，不输出）】\n\n所有新增信息是否属于"最小必要补全"\n对话是否承担信息并形成推动\n视角是否未越界；术语与时间线是否一致\n动词/五感是否显著多于抽象评判词；禁用词是否规避\n若出现系统/金手指，是否产生可见且有代价的变化\n【开始创作】\n请据上述规则，基于【剧情点】直接生成正文',
        category: '写作要求',
        order_num: 3,
        fields: null
      },
      {
        name: '乐子文风格',
        content: '作者的写作风格鲜明独特，可以用"接地气"来概括。\n他注重写作技巧的实际运用，语言风格直白通俗\n\n偏爱反套路，反转，悬念，打脸，爽点的情节设计\n拥有厚重的习俗人文的文笔，来刻画人的七情六欲\n偶尔穿插一种增强配角记忆点的东西（不需要每张都带，在情感最爆发的时候释放的时候来穿插）\n不要说教夹带私货\n\n这种风格使其作品更易于被大众接受，并引发情感共鸣。\n\n巧妙运用角色信息的性格，彰显对话特点，动作行为等',
        category: '写作风格',
        order_num: 0,
        fields: null
      }
    ];

    // 插入提示词
    const insertStmt = db.prepare(`
      INSERT INTO prompts (name, description, content, category, order_num, fields, subcategories, card_type, password)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);

    for (const prompt of defaultPrompts) {
      try {
        insertStmt.run(
          prompt.name,
          null,
          prompt.content,
          prompt.category,
          prompt.order_num,
          prompt.fields,
          null,
          'normal',
          null
        );
        console.log(`✓ 导入默认提示词：${prompt.name}`);
      } catch (error) {
        console.error(`✗ 导入提示词失败：${prompt.name} - ${error.message}`);
      }
    }

    console.log('默认提示词导入完成');
  } catch (error) {
    console.error('导入默认提示词过程出错:', error.message);
  }
};

// 初始化数据库
initDatabase();

// 导入默认提示词
importDefaultPrompts();

module.exports = db;
