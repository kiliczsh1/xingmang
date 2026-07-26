const express = require('express');
const router = express.Router();
const db = require('../database/db');

router.get('/book/:bookId', (req, res) => {
  try {
    const characters = db.prepare('SELECT * FROM characters WHERE book_id = ? ORDER BY updated_at DESC').all(req.params.bookId);
    const formattedCharacters = characters.map(char => ({
      ...char,
      folders: char.folders ? JSON.parse(char.folders) : []
    }));
    res.json({ success: true, data: formattedCharacters });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/folders/:bookId', (req, res) => {
  try {
    const characters = db.prepare('SELECT DISTINCT folder FROM characters WHERE book_id = ?').all(req.params.bookId);
    const folders = characters.map(c => c.folder).filter(f => f);
    res.json({ success: true, data: folders });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const character = db.prepare('SELECT * FROM characters WHERE id = ?').get(req.params.id);
    if (!character) {
      return res.status(404).json({ success: false, message: '角色不存在' });
    }
    const formattedCharacter = {
      ...character,
      folders: character.folders ? JSON.parse(character.folders) : []
    };
    res.json({ success: true, data: formattedCharacter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/', (req, res) => {
  try {
    const { book_id, name, gender, personality, info, folder, folders } = req.body;
    const stmt = db.prepare(`
      INSERT INTO characters (book_id, name, gender, personality, info, folder, folders)
      VALUES (?, ?, ?, ?, ?, ?, ?)
    `);
    const result = stmt.run(
      book_id,
      name || '未命名角色',
      gender || 'unknown',
      personality || '',
      info || '',
      folder || '全部',
      folders ? JSON.stringify(folders) : '[]'
    );
    const character = db.prepare('SELECT * FROM characters WHERE id = ?').get(result.lastInsertRowid);
    const formattedCharacter = {
      ...character,
      folders: character.folders ? JSON.parse(character.folders) : []
    };
    res.json({ success: true, data: formattedCharacter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/:id', (req, res) => {
  try {
    const { name, gender, personality, info, folder, folders } = req.body;
    const stmt = db.prepare(`
      UPDATE characters SET
        name = ?,
        gender = ?,
        personality = ?,
        info = ?,
        folder = ?,
        folders = ?,
        updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `);
    stmt.run(
      name,
      gender,
      personality,
      info,
      folder,
      folders ? JSON.stringify(folders) : '[]',
      req.params.id
    );
    const character = db.prepare('SELECT * FROM characters WHERE id = ?').get(req.params.id);
    const formattedCharacter = {
      ...character,
      folders: character.folders ? JSON.parse(character.folders) : []
    };
    res.json({ success: true, data: formattedCharacter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM characters WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/batch-delete', (req, res) => {
  try {
    const { ids } = req.body;
    if (!ids || !Array.isArray(ids) || ids.length === 0) {
      return res.status(400).json({ success: false, message: '请提供要删除的角色ID' });
    }
    const placeholders = ids.map(() => '?').join(',');
    const stmt = db.prepare(`DELETE FROM characters WHERE id IN (${placeholders})`);
    stmt.run(...ids);
    res.json({ success: true, message: '批量删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/batch-move', (req, res) => {
  try {
    const { ids, folder } = req.body;
    if (!ids || !Array.isArray(ids) || ids.length === 0) {
      return res.status(400).json({ success: false, message: '请提供要移动的角色ID' });
    }
    const stmt = db.prepare('UPDATE characters SET folder = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    for (const id of ids) {
      stmt.run(folder, id);
    }
    res.json({ success: true, message: '批量移动成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
