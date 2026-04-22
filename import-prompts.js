const fs = require('fs');
const path = require('path');

console.log('=== 网文助手 - 提示词批量导入工具 ===\n');

// 从 server 目录加载 better-sqlite3
const Database = require('./server/node_modules/better-sqlite3');

// 读取 JSON 文件 - 支持多个可能的位置
const possiblePaths = [
  path.join(__dirname, '提示词整体备份_2026-4-7.json'),
  'd:\\Downloads\\提示词整体备份_2026-4-7.json',
  path.join(__dirname, '..', 'Downloads', '提示词整体备份_2026-4-7.json')
];

let jsonPath = null;
for (const p of possiblePaths) {
  if (fs.existsSync(p)) {
    jsonPath = p;
    break;
  }
}

if (!jsonPath) {
  console.error('错误：找不到 JSON 文件');
  console.log('请在以下位置之一放置文件 "提示词整体备份_2026-4-7.json":');
  possiblePaths.forEach(p => console.log('  -', p));
  process.exit(1);
}

console.log('读取 JSON 文件:', jsonPath);
const jsonData = JSON.parse(fs.readFileSync(jsonPath, 'utf-8'));

// 获取数据库路径 - 支持多个可能的位置
const possibleDbPaths = [
  path.join(__dirname, 'server', 'database', 'novel.db'),
  path.join(__dirname, '..', 'server', 'database', 'novel.db'),
  'd:\\bimoxinghe\\XingNovel\\server\\database\\novel.db'
];

let dbPath = null;
for (const p of possibleDbPaths) {
  if (fs.existsSync(p)) {
    dbPath = p;
    break;
  }
}

if (!dbPath) {
  console.error('错误：数据库文件不存在');
  console.log('数据库可能位置:');
  possibleDbPaths.forEach(p => console.log('  -', p));
  console.log('\n请先运行网文助手系统，确保数据库已初始化');
  process.exit(1);
}

console.log('数据库路径:', dbPath);

const db = new Database(dbPath);

// 确保表结构
const ensureTableSchema = () => {
  const columns = db.prepare("PRAGMA table_info(prompts)").all();
  const columnNames = columns.map(col => col.name);
  
  const newColumns = [
    { name: 'fields', sql: 'ALTER TABLE prompts ADD COLUMN fields TEXT' },
    { name: 'subcategories', sql: 'ALTER TABLE prompts ADD COLUMN subcategories TEXT' },
    { name: 'description', sql: 'ALTER TABLE prompts ADD COLUMN description TEXT' },
    { name: 'card_type', sql: "ALTER TABLE prompts ADD COLUMN card_type TEXT DEFAULT 'normal'" },
    { name: 'password', sql: 'ALTER TABLE prompts ADD COLUMN password TEXT' }
  ];
  
  for (const col of newColumns) {
    if (!columnNames.includes(col.name)) {
      db.exec(col.sql);
      console.log(`✓ 添加字段：${col.name}`);
    }
  }
};

try {
  ensureTableSchema();
  
  // 准备 SQL 语句
  const insertStmt = db.prepare(`
    INSERT INTO prompts (
      name, description, content, category, order_num, fields, subcategories, card_type, password
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
  `);
  
  const updateStmt = db.prepare(`
    UPDATE prompts SET 
      description = ?, content = ?, order_num = ?, fields = ?, subcategories = ?, 
      card_type = ?, password = ?, updated_at = CURRENT_TIMESTAMP
    WHERE name = ? AND category = ?
  `);
  
  const checkStmt = db.prepare('SELECT id FROM prompts WHERE name = ? AND category = ?');
  
  let imported = 0;
  let updated = 0;
  let skipped = 0;
  
  // 处理提示词
  if (jsonData.data && jsonData.data.packs) {
    const packs = jsonData.data.packs;
    
    for (const [category, prompts] of Object.entries(packs)) {
      if (!Array.isArray(prompts) || prompts.length === 0) {
        console.log(`\n跳过空分类：${category}`);
        continue;
      }
      
      console.log(`\n📁 分类：${category} (${prompts.length} 个提示词)`);
      
      for (const prompt of prompts) {
        try {
          const exists = checkStmt.get(prompt.name, category);
          
          const fields = prompt.fields ? JSON.stringify(prompt.fields) : null;
          const subcategories = (prompt.subcategories && prompt.subcategories.length > 0) 
            ? JSON.stringify(prompt.subcategories) : null;
          
          if (exists) {
            updateStmt.run(
              prompt.description || null,
              prompt.content,
              prompt.order_num || 0,
              fields,
              subcategories,
              prompt.card_type || 'normal',
              prompt.password || null,
              prompt.name,
              category
            );
            updated++;
            console.log(`  ✏️  更新：${prompt.name}`);
          } else {
            insertStmt.run(
              prompt.name,
              prompt.description || null,
              prompt.content,
              category,
              prompt.order_num || 0,
              fields,
              subcategories,
              prompt.card_type || 'normal',
              prompt.password || null
            );
            imported++;
            console.log(`  ✅ 导入：${prompt.name}`);
          }
        } catch (error) {
          console.error(`  ❌ 失败：${prompt.name} - ${error.message}`);
          skipped++;
        }
      }
    }
  }
  
  console.log('\n' + '='.repeat(50));
  console.log('📊 导入统计:');
  console.log(`  新导入：${imported} 个`);
  console.log(`  更  新：${updated} 个`);
  console.log(`  跳  过：${skipped} 个`);
  console.log(`  总  计：${imported + updated + skipped} 个`);
  console.log('='.repeat(50));
  console.log('\n✅ 提示词导入完成！请重启网文助手系统查看效果。');
  
} catch (error) {
  console.error('\n❌ 导入失败:', error.message);
  console.error(error.stack);
  process.exit(1);
} finally {
  db.close();
}
