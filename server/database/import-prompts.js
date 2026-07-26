const fs = require('fs');
const path = require('path');
const Database = require('better-sqlite3');

// 获取数据库路径
const getDbPath = () => {
  if (process.env.DB_PATH) {
    return process.env.DB_PATH;
  }
  return path.join(__dirname, 'novel.db');
};

// 读取 JSON 文件
const readJsonFile = () => {
  const jsonPath = path.join(__dirname, '..', '..', '..', 'Downloads', '提示词整体备份_2026-4-7.json');
  
  if (!fs.existsSync(jsonPath)) {
    console.error('错误：找不到 JSON 文件');
    console.log('请确保文件存在于:', jsonPath);
    return null;
  }
  
  const content = fs.readFileSync(jsonPath, 'utf-8');
  return JSON.parse(content);
};

// 确保 prompts 表有所有必需的字段
const ensureTableSchema = (db) => {
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
      console.log(`已添加字段：${col.name}`);
    }
  }
};

// 导入提示词
const importPrompts = () => {
  console.log('=== 开始导入提示词 ===\n');
  
  const jsonData = readJsonFile();
  if (!jsonData) {
    return;
  }
  
  const db = new Database(getDbPath());
  
  try {
    // 确保表结构正确
    ensureTableSchema(db);
    
    // 准备插入语句
    const insertStmt = db.prepare(`
      INSERT INTO prompts (
        name, 
        description, 
        content, 
        category, 
        order_num, 
        fields, 
        subcategories,
        card_type,
        password
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    
    // 准备更新语句（用于存在同名同分类的提示词）
    const updateStmt = db.prepare(`
      UPDATE prompts SET 
        description = ?,
        content = ?,
        order_num = ?,
        fields = ?,
        subcategories = ?,
        card_type = ?,
        password = ?,
        updated_at = CURRENT_TIMESTAMP
      WHERE name = ? AND category = ?
    `);
    
    // 检查提示词是否存在的查询
    const checkStmt = db.prepare(`
      SELECT id FROM prompts WHERE name = ? AND category = ?
    `);
    
    let importedCount = 0;
    let updatedCount = 0;
    let skippedCount = 0;
    
    // 处理 packs 中的提示词
    if (jsonData.data && jsonData.data.packs) {
      const packs = jsonData.data.packs;
      
      for (const [categoryName, prompts] of Object.entries(packs)) {
        if (!Array.isArray(prompts) || prompts.length === 0) {
          console.log(`跳过空分类：${categoryName}`);
          continue;
        }
        
        console.log(`\n处理分类：${categoryName} (${prompts.length} 个提示词)`);
        
        for (const prompt of prompts) {
          try {
            const name = prompt.name;
            const content = prompt.content;
            const orderNum = prompt.order_num || 0;
            const fields = prompt.fields ? JSON.stringify(prompt.fields) : null;
            const subcategories = prompt.subcategories && prompt.subcategories.length > 0 
              ? JSON.stringify(prompt.subcategories) 
              : null;
            const cardType = prompt.card_type || 'normal';
            const password = prompt.password || null;
            const description = prompt.description || null;
            
            // 检查是否已存在
            const existing = checkStmt.get(name, categoryName);
            
            if (existing) {
              // 更新现有提示词
              updateStmt.run(
                description,
                content,
                orderNum,
                fields,
                subcategories,
                cardType,
                password,
                name,
                categoryName
              );
              updatedCount++;
              console.log(`  ✓ 更新：${name}`);
            } else {
              // 插入新提示词
              insertStmt.run(
                name,
                description,
                content,
                categoryName,
                orderNum,
                fields,
                subcategories,
                cardType,
                password
              );
              importedCount++;
              console.log(`  + 导入：${name}`);
            }
          } catch (error) {
            console.error(`  ✗ 失败：${prompt.name} - ${error.message}`);
            skippedCount++;
          }
        }
      }
    }
    
    console.log('\n=== 导入完成 ===');
    console.log(`新导入：${importedCount} 个提示词`);
    console.log(`更新：${updatedCount} 个提示词`);
    console.log(`跳过/失败：${skippedCount} 个提示词`);
    console.log(`总计：${importedCount + updatedCount + skippedCount} 个提示词`);
    
  } catch (error) {
    console.error('\n导入过程发生错误:', error.message);
    console.error(error.stack);
  } finally {
    db.close();
    console.log('\n数据库连接已关闭');
  }
};

// 执行导入
importPrompts();
