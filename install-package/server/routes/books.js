const express = require('express');
const router = express.Router();
const db = require('../database/db');

// 获取所有书本
router.get('/', (req, res) => {
  try {
    const books = db.prepare('SELECT * FROM books ORDER BY updated_at DESC').all();
    const formattedBooks = books.map(book => ({
      ...book,
      tags: book.tags ? JSON.parse(book.tags) : [],
      is_public: book.is_public === 1
    }));
    res.json({ success: true, data: formattedBooks });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个书本
router.get('/:id', (req, res) => {
  try {
    const book = db.prepare('SELECT * FROM books WHERE id = ?').get(req.params.id);
    if (!book) {
      return res.status(404).json({ success: false, message: '书本不存在' });
    }
    const formattedBook = {
      ...book,
      tags: book.tags ? JSON.parse(book.tags) : [],
      is_public: book.is_public === 1
    };
    res.json({ success: true, data: formattedBook });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建书本
router.post('/', (req, res) => {
  try {
    const { title, description, cover, author, category, tags, status, is_public } = req.body;
    const stmt = db.prepare(`
      INSERT INTO books (title, description, cover, author, category, tags, status, is_public)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `);
    const result = stmt.run(
      title,
      description || '',
      cover || '',
      author || '',
      category || '',
      tags ? JSON.stringify(tags) : '',
      status || 'draft',
      is_public ? 1 : 0
    );
    const book = db.prepare('SELECT * FROM books WHERE id = ?').get(result.lastInsertRowid);
    const formattedBook = {
      ...book,
      tags: book.tags ? JSON.parse(book.tags) : [],
      is_public: book.is_public === 1
    };
    res.json({ success: true, data: formattedBook });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新书本
router.put('/:id', (req, res) => {
  try {
    const { title, description, cover, author, category, tags, status, is_public } = req.body;
    const existingBook = db.prepare('SELECT * FROM books WHERE id = ?').get(req.params.id);
    if (!existingBook) {
      return res.status(404).json({ success: false, message: '书本不存在' });
    }
    const stmt = db.prepare(`
      UPDATE books SET
        title = ?,
        description = ?,
        cover = ?,
        author = ?,
        category = ?,
        tags = ?,
        status = ?,
        is_public = ?,
        updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `);
    stmt.run(
      title ?? existingBook.title,
      description ?? existingBook.description ?? '',
      cover ?? existingBook.cover ?? '',
      author ?? existingBook.author ?? '',
      category ?? existingBook.category ?? '',
      tags != null ? JSON.stringify(tags) : (existingBook.tags ?? ''),
      status ?? existingBook.status ?? 'draft',
      is_public != null ? (is_public ? 1 : 0) : (existingBook.is_public ? 1 : 0),
      req.params.id
    );
    const book = db.prepare('SELECT * FROM books WHERE id = ?').get(req.params.id);
    const formattedBook = {
      ...book,
      tags: book.tags ? JSON.parse(book.tags) : [],
      is_public: book.is_public === 1
    };
    res.json({ success: true, data: formattedBook });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除书本
router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM books WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

