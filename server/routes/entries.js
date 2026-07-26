const express = require('express');
const router = express.Router();
const db = require('../database/db');

const CATEGORY_ICONS = {
  character: 'User',
  world_setting: 'Setting',
  location: 'Location',
  item: 'Box',
  faction: 'OfficeBuilding',
  event: 'Clock',
  skill: 'MagicStick',
  clue: 'Search',
  other: 'Document'
};

const CATEGORY_COLORS = {
  character: '#409EFF',
  world_setting: '#67C23A',
  location: '#E6A23C',
  item: '#F56C6C',
  faction: '#909399',
  event: '#9b59b6',
  skill: '#e91e63',
  clue: '#00bcd4',
  other: '#95a5a6'
};

function initDefaultCategories() {
  const existing = db.prepare('SELECT COUNT(*) as count FROM entry_categories').get();
  if (existing.count === 0) {
    const types = ['character', 'world_setting', 'location', 'item', 'faction', 'event', 'skill', 'clue', 'other'];
    const names = {
      character: '人物',
      world_setting: '世界观',
      location: '地点',
      item: '物品',
      faction: '势力',
      event: '事件',
      skill: '技能',
      clue: '线索',
      other: '其他'
    };
    
    const stmt = db.prepare('INSERT INTO entry_categories (name, type, icon, color, sort_order) VALUES (?, ?, ?, ?, ?)');
    types.forEach((type, index) => {
      stmt.run(names[type], type, CATEGORY_ICONS[type], CATEGORY_COLORS[type], index);
    });
    console.log('已初始化默认词条分类');
  }
}

initDefaultCategories();

router.get('/', (req, res) => {
  try {
    const { book_id, category_type, keyword, page = 1, page_size = 12 } = req.query;
    const offset = (parseInt(page) - 1) * parseInt(page_size);
    
    let query = `
      SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color
      FROM entries e
      LEFT JOIN entry_categories ec ON e.category_id = ec.id
      WHERE 1=1
    `;
    const params = [];
    
    if (book_id) {
      query += ' AND (e.book_id = ? OR e.is_public = 1)';
      params.push(parseInt(book_id));
    } else {
      query += ' AND e.is_public = 1';
    }
    
    if (category_type) {
      query += ' AND e.category_type = ?';
      params.push(category_type);
    }
    
    if (keyword) {
      query += ' AND (e.name LIKE ? OR e.description LIKE ? OR e.tags LIKE ?)';
      params.push(`%${keyword}%`, `%${keyword}%`, `%${keyword}%`);
    }
    
    query += ' ORDER BY e.use_count DESC, e.created_at DESC';
    
    const countQuery = query.replace('SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color', 'SELECT COUNT(*) as total');
    const countResult = db.prepare(countQuery).get(...params);
    
    query += ' LIMIT ? OFFSET ?';
    params.push(parseInt(page_size), offset);
    
    const entries = db.prepare(query).all(...params);
    
    const formattedEntries = entries.map(entry => ({
      ...entry,
      tags: JSON.parse(entry.tags || '[]'),
      custom_fields: JSON.parse(entry.custom_fields || '{}'),
      is_public: !!entry.is_public
    }));
    
    const categoriesQuery = `
      SELECT ec.*, COUNT(e.id) as entry_count
      FROM entry_categories ec
      LEFT JOIN entries e ON ec.id = e.category_id AND (e.book_id = ? OR e.is_public = 1)
      GROUP BY ec.id
      ORDER BY ec.sort_order ASC
    `;
    const categories = db.prepare(categoriesQuery).all(book_id ? parseInt(book_id) : null);
    
    res.json({
      success: true,
      data: {
        entries: formattedEntries,
        categories: categories,
        total: countResult.total,
        page: parseInt(page),
        page_size: parseInt(page_size)
      }
    });
  } catch (error) {
    console.error('获取词条列表失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/categories', (req, res) => {
  try {
    const { book_id } = req.query;
    
    const query = `
      SELECT ec.*, COUNT(e.id) as entry_count
      FROM entry_categories ec
      LEFT JOIN entries e ON ec.id = e.category_id AND (e.book_id = ? OR e.is_public = 1)
      GROUP BY ec.id
      ORDER BY ec.sort_order ASC
    `;
    const categories = db.prepare(query).all(book_id ? parseInt(book_id) : null);
    
    res.json({ success: true, data: categories });
  } catch (error) {
    console.error('获取词条分类失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const entry = db.prepare(`
      SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color
      FROM entries e
      LEFT JOIN entry_categories ec ON e.category_id = ec.id
      WHERE e.id = ?
    `).get(parseInt(req.params.id));
    
    if (!entry) {
      return res.status(404).json({ success: false, message: '词条不存在' });
    }
    
    res.json({
      success: true,
      data: {
        ...entry,
        tags: JSON.parse(entry.tags || '[]'),
        custom_fields: JSON.parse(entry.custom_fields || '{}'),
        is_public: !!entry.is_public
      }
    });
  } catch (error) {
    console.error('获取词条详情失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/', (req, res) => {
  try {
    const {
      book_id,
      category_id,
      category_type,
      name,
      avatar,
      tags,
      appearance,
      background,
      personality,
      relationships,
      custom_fields,
      description,
      is_public,
      creator_name,
      creator_avatar
    } = req.body;
    
    const stmt = db.prepare(`
      INSERT INTO entries (
        book_id, category_id, category_type, name, avatar, tags,
        appearance, background, personality, relationships, custom_fields,
        description, is_public, creator_name, creator_avatar
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    
    const result = stmt.run(
      book_id || null,
      category_id,
      category_type || 'other',
      name,
      avatar || null,
      JSON.stringify(tags || []),
      appearance || '',
      background || '',
      personality || null,
      relationships || null,
      JSON.stringify(custom_fields || {}),
      description || '',
      is_public ? 1 : 0,
      creator_name || null,
      creator_avatar || null
    );
    
    const entry = db.prepare(`
      SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color
      FROM entries e
      LEFT JOIN entry_categories ec ON e.category_id = ec.id
      WHERE e.id = ?
    `).get(result.lastInsertRowid);
    
    res.json({
      success: true,
      data: {
        ...entry,
        tags: JSON.parse(entry.tags || '[]'),
        custom_fields: JSON.parse(entry.custom_fields || '{}'),
        is_public: !!entry.is_public
      }
    });
  } catch (error) {
    console.error('创建词条失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/:id', (req, res) => {
  try {
    const {
      category_id,
      category_type,
      name,
      avatar,
      tags,
      appearance,
      background,
      personality,
      relationships,
      custom_fields,
      description,
      is_public
    } = req.body;
    
    const stmt = db.prepare(`
      UPDATE entries SET
        category_id = COALESCE(?, category_id),
        category_type = COALESCE(?, category_type),
        name = COALESCE(?, name),
        avatar = COALESCE(?, avatar),
        tags = COALESCE(?, tags),
        appearance = COALESCE(?, appearance),
        background = COALESCE(?, background),
        personality = COALESCE(?, personality),
        relationships = COALESCE(?, relationships),
        custom_fields = COALESCE(?, custom_fields),
        description = COALESCE(?, description),
        is_public = COALESCE(?, is_public),
        updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `);
    
    stmt.run(
      category_id,
      category_type,
      name,
      avatar,
      tags ? JSON.stringify(tags) : null,
      appearance,
      background,
      personality,
      relationships,
      custom_fields ? JSON.stringify(custom_fields) : null,
      description,
      is_public !== undefined ? (is_public ? 1 : 0) : null,
      parseInt(req.params.id)
    );
    
    const entry = db.prepare(`
      SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color
      FROM entries e
      LEFT JOIN entry_categories ec ON e.category_id = ec.id
      WHERE e.id = ?
    `).get(parseInt(req.params.id));
    
    res.json({
      success: true,
      data: {
        ...entry,
        tags: JSON.parse(entry.tags || '[]'),
        custom_fields: JSON.parse(entry.custom_fields || '{}'),
        is_public: !!entry.is_public
      }
    });
  } catch (error) {
    console.error('更新词条失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM entries WHERE id = ?');
    stmt.run(parseInt(req.params.id));
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    console.error('删除词条失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/:id/add-to-book', (req, res) => {
  try {
    const { bookId } = req.body;
    const entryId = parseInt(req.params.id);
    
    const existingEntry = db.prepare('SELECT * FROM entries WHERE id = ?').get(entryId);
    if (!existingEntry) {
      return res.status(404).json({ success: false, message: '词条不存在' });
    }
    
    const stmt = db.prepare(`
      INSERT INTO entries (
        book_id, category_id, category_type, name, avatar, tags,
        appearance, background, personality, relationships, custom_fields,
        description, is_public, creator_name, creator_avatar
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);
    
    const result = stmt.run(
      bookId,
      existingEntry.category_id,
      existingEntry.category_type,
      existingEntry.name,
      existingEntry.avatar,
      existingEntry.tags,
      existingEntry.appearance,
      existingEntry.background,
      existingEntry.personality,
      existingEntry.relationships,
      existingEntry.custom_fields,
      existingEntry.description,
      0,
      existingEntry.creator_name,
      existingEntry.creator_avatar
    );
    
    db.prepare('UPDATE entries SET use_count = use_count + 1 WHERE id = ?').run(entryId);
    
    const newEntry = db.prepare(`
      SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color
      FROM entries e
      LEFT JOIN entry_categories ec ON e.category_id = ec.id
      WHERE e.id = ?
    `).get(result.lastInsertRowid);
    
    res.json({
      success: true,
      data: {
        ...newEntry,
        tags: JSON.parse(newEntry.tags || '[]'),
        custom_fields: JSON.parse(newEntry.custom_fields || '{}'),
        is_public: !!newEntry.is_public
      }
    });
  } catch (error) {
    console.error('添加词条到小说失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/:id/ai-auto-complete', async (req, res) => {
  try {
    const { configId } = req.body;
    const entryId = parseInt(req.params.id);
    
    const entry = db.prepare('SELECT * FROM entries WHERE id = ?').get(entryId);
    if (!entry) {
      return res.status(404).json({ success: false, message: '词条不存在' });
    }
    
    res.json({
      success: true,
      data: {
        ...entry,
        tags: JSON.parse(entry.tags || '[]'),
        custom_fields: JSON.parse(entry.custom_fields || '{}'),
        is_public: !!entry.is_public,
        description: entry.description + '\n\n（AI 自动补全功能需要配置 AI 服务）'
      }
    });
  } catch (error) {
    console.error('AI 自动补全失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/batch-add-to-book', (req, res) => {
  try {
    const { entryIds, bookId } = req.body;
    
    if (!entryIds || !Array.isArray(entryIds) || entryIds.length === 0) {
      return res.status(400).json({ success: false, message: '请提供有效的词条 ID 列表' });
    }
    
    const results = [];
    
    db.transaction(() => {
      const stmt = db.prepare(`
        INSERT INTO entries (
          book_id, category_id, category_type, name, avatar, tags,
          appearance, background, personality, relationships, custom_fields,
          description, is_public, creator_name, creator_avatar
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
      `);
      
      const updateUseCount = db.prepare('UPDATE entries SET use_count = use_count + 1 WHERE id = ?');
      
      for (const entryId of entryIds) {
        const existingEntry = db.prepare('SELECT * FROM entries WHERE id = ?').get(entryId);
        if (existingEntry) {
          const result = stmt.run(
            bookId,
            existingEntry.category_id,
            existingEntry.category_type,
            existingEntry.name,
            existingEntry.avatar,
            existingEntry.tags,
            existingEntry.appearance,
            existingEntry.background,
            existingEntry.personality,
            existingEntry.relationships,
            existingEntry.custom_fields,
            existingEntry.description,
            0,
            existingEntry.creator_name,
            existingEntry.creator_avatar
          );
          
          updateUseCount.run(entryId);
          
          const newEntry = db.prepare(`
            SELECT e.*, ec.name as category_name, ec.icon as category_icon, ec.color as category_color
            FROM entries e
            LEFT JOIN entry_categories ec ON e.category_id = ec.id
            WHERE e.id = ?
          `).get(result.lastInsertRowid);
          
          results.push({
            ...newEntry,
            tags: JSON.parse(newEntry.tags || '[]'),
            custom_fields: JSON.parse(newEntry.custom_fields || '{}'),
            is_public: !!newEntry.is_public
          });
        }
      }
    })();
    
    res.json({ success: true, data: results, message: `成功添加 ${results.length} 个词条到小说` });
  } catch (error) {
    console.error('批量添加词条失败:', error);
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
