const express = require('express');
const router = express.Router();
const db = require('../database/db');

// 获取书本的所有章节
router.get('/book/:bookId', (req, res) => {
  try {
    const chapters = db.prepare('SELECT * FROM chapters WHERE book_id = ? ORDER BY order_num ASC').all(req.params.bookId);
    res.json({ success: true, data: chapters });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个章节
router.post('/import-book', (req, res) => {
  try {
    const { bookId, chapters } = req.body;

    if (!bookId) {
      return res.status(400).json({ success: false, message: '缺少书籍ID' });
    }

    if (!Array.isArray(chapters) || chapters.length === 0) {
      return res.status(400).json({ success: false, message: '没有可导入的章节内容' });
    }

    const book = db.prepare('SELECT id FROM books WHERE id = ?').get(bookId);
    if (!book) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }

    const getMaxOrderStmt = db.prepare('SELECT COALESCE(MAX(order_num), -1) AS maxOrder FROM chapters WHERE book_id = ?');
    const insertStmt = db.prepare('INSERT INTO chapters (book_id, title, content, order_num, type, volume_id) VALUES (?, ?, ?, ?, ?, ?)');
    const getChapterStmt = db.prepare('SELECT * FROM chapters WHERE id = ?');
    const updateBookStmt = db.prepare('UPDATE books SET updated_at = CURRENT_TIMESTAMP WHERE id = ?');

    const importTransaction = db.transaction((payload) => {
      const maxOrderResult = getMaxOrderStmt.get(bookId);
      const baseOrder = typeof maxOrderResult?.maxOrder === 'number' ? maxOrderResult.maxOrder : -1;
      const insertedChapters = [];

      payload.forEach((chapter, index) => {
        const title = String(chapter?.title || '').trim() || `第${index + 1}章`;
        const content = String(chapter?.content || '').trim();

        if (!content) {
          return;
        }

        const result = insertStmt.run(bookId, title, content, baseOrder + insertedChapters.length + 1, 'chapter', null);
        insertedChapters.push(getChapterStmt.get(result.lastInsertRowid));
      });

      updateBookStmt.run(bookId);
      return insertedChapters;
    });

    const inserted = importTransaction(chapters);
    res.json({
      success: true,
      data: {
        insertedCount: inserted.length,
        chapters: inserted
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const chapter = db.prepare('SELECT * FROM chapters WHERE id = ?').get(req.params.id);
    if (!chapter) {
      return res.status(404).json({ success: false, message: '章节不存在' });
    }
    res.json({ success: true, data: chapter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建章节
router.post('/', (req, res) => {
  try {
    const { book_id, title, content, summary, order_num, type, volume_id } = req.body;
    const stmt = db.prepare('INSERT INTO chapters (book_id, title, content, summary, order_num, type, volume_id) VALUES (?, ?, ?, ?, ?, ?, ?)');
    const result = stmt.run(book_id, title, content || '', summary || '', order_num || 0, type || 'chapter', volume_id || null);
    const chapter = db.prepare('SELECT * FROM chapters WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: chapter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新章节
router.put('/:id', (req, res) => {
  try {
    const { title, content, summary, order_num } = req.body;
    const existing = db.prepare('SELECT * FROM chapters WHERE id = ?').get(req.params.id);

    if (!existing) {
      return res.status(404).json({ success: false, message: '章节不存在' });
    }

    const stmt = db.prepare(`
      UPDATE chapters
      SET title = ?, content = ?, summary = ?, order_num = ?, updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `);
    stmt.run(
      title ?? existing.title,
      content ?? existing.content,
      summary ?? existing.summary ?? '',
      order_num ?? existing.order_num,
      req.params.id
    );
    const chapter = db.prepare('SELECT * FROM chapters WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: chapter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除章节
router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM chapters WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

