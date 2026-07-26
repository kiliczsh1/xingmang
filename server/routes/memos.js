const express = require('express');
const router = express.Router();
const db = require('../database/db');

// 获取所有备忘录
router.get('/', (req, res) => {
  try {
    const memos = db.prepare('SELECT * FROM memos ORDER BY is_pinned DESC, order_num ASC').all();
    res.json({ success: true, data: memos });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 搜索备忘录
router.get('/search', (req, res) => {
  try {
    const { keyword, category } = req.query;
    let query = 'SELECT * FROM memos WHERE 1=1';
    const params = [];
    
    if (keyword) {
      query += ' AND (title LIKE ? OR content LIKE ?)';
      params.push(`%${keyword}%`, `%${keyword}%`);
    }
    if (category) {
      query += ' AND category = ?';
      params.push(category);
    }
    
    query += ' ORDER BY is_pinned DESC, order_num ASC';
    const memos = db.prepare(query).all(...params);
    res.json({ success: true, data: memos });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个备忘录
router.get('/:id', (req, res) => {
  try {
    const memo = db.prepare('SELECT * FROM memos WHERE id = ?').get(req.params.id);
    if (!memo) {
      return res.status(404).json({ success: false, message: '备忘录不存在' });
    }
    res.json({ success: true, data: memo });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建备忘录
router.post('/', (req, res) => {
  try {
    const { title, content, category, order_num } = req.body;
    const wordCount = content ? content.length : 0;
    const stmt = db.prepare('INSERT INTO memos (title, content, category, order_num, word_count) VALUES (?, ?, ?, ?, ?)');
    const result = stmt.run(title, content || '', category || '默认', order_num || 0, wordCount);
    const memo = db.prepare('SELECT * FROM memos WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: memo });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新备忘录
router.put('/:id', (req, res) => {
  try {
    const { title, content, category, order_num } = req.body;
    const wordCount = content ? content.length : 0;
    const stmt = db.prepare('UPDATE memos SET title = ?, content = ?, category = ?, order_num = ?, word_count = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(title, content, category, order_num, wordCount, req.params.id);
    const memo = db.prepare('SELECT * FROM memos WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: memo });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除备忘录
router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM memos WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 批量操作
router.post('/batch', (req, res) => {
  try {
    const { action, ids, data } = req.body;
    
    if (!ids || !Array.isArray(ids) || ids.length === 0) {
      return res.status(400).json({ success: false, message: '请提供有效的 ID 列表' });
    }
    
    db.transaction(() => {
      if (action === 'delete') {
        const deleteStmt = db.prepare('DELETE FROM memos WHERE id = ?');
        ids.forEach(id => deleteStmt.run(id));
      } else if (action === 'update') {
        const { title, content, category, order_num, tags, is_pinned } = data || {};
        const wordCount = content ? content.length : 0;
        const updateStmt = db.prepare(`
          UPDATE memos SET 
            title = COALESCE(?, title),
            content = COALESCE(?, content),
            category = COALESCE(?, category),
            order_num = COALESCE(?, order_num),
            word_count = ?,
            tags = COALESCE(?, tags),
            is_pinned = COALESCE(?, is_pinned),
            updated_at = CURRENT_TIMESTAMP
          WHERE id = ?
        `);
        ids.forEach(id => {
          updateStmt.run(title, content, category, order_num, wordCount, tags, is_pinned, id);
        });
      } else if (action === 'toggle_pin') {
        const toggleStmt = db.prepare('UPDATE memos SET is_pinned = 1 - is_pinned, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
        ids.forEach(id => toggleStmt.run(id));
      }
    })();
    
    res.json({ success: true, message: '批量操作成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

