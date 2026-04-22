const express = require('express');
const router = express.Router();
const db = require('../database/db');

// 获取书本的所有分卷
router.get('/book/:bookId', (req, res) => {
  try {
    const volumes = db.prepare('SELECT * FROM volumes WHERE book_id = ? ORDER BY order_num ASC').all(req.params.bookId);
    res.json({ success: true, data: volumes });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个分卷
router.get('/:id', (req, res) => {
  try {
    const volume = db.prepare('SELECT * FROM volumes WHERE id = ?').get(req.params.id);
    if (!volume) {
      return res.status(404).json({ success: false, message: '分卷不存在' });
    }
    res.json({ success: true, data: volume });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建分卷
router.post('/', (req, res) => {
  try {
    const { book_id, title, order_num } = req.body;
    const stmt = db.prepare('INSERT INTO volumes (book_id, title, order_num) VALUES (?, ?, ?)');
    const result = stmt.run(book_id, title, order_num || 0);
    const volume = db.prepare('SELECT * FROM volumes WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: volume });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新分卷
router.put('/:id', (req, res) => {
  try {
    const { title, order_num } = req.body;
    const stmt = db.prepare('UPDATE volumes SET title = ?, order_num = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(title, order_num, req.params.id);
    const volume = db.prepare('SELECT * FROM volumes WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: volume });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除分卷
router.delete('/:id', (req, res) => {
  try {
    // 删除分卷时，将该分卷下的章节的 volume_id 设为 NULL
    db.prepare('UPDATE chapters SET volume_id = NULL WHERE volume_id = ?').run(req.params.id);
    const stmt = db.prepare('DELETE FROM volumes WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
