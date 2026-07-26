const express = require('express');
const router = express.Router();
const db = require('../database/db');

router.get('/', (req, res) => {
  try {
    const generators = db.prepare('SELECT * FROM generators ORDER BY order_num ASC').all();
    res.json({ success: true, data: generators });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const generator = db.prepare('SELECT * FROM generators WHERE id = ?').get(req.params.id);
    if (!generator) {
      return res.status(404).json({ success: false, message: '生成器不存在' });
    }
    res.json({ success: true, data: generator });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/', (req, res) => {
  try {
    const { name, description, icon, core_prompt, remark, order_num } = req.body;
    const stmt = db.prepare('INSERT INTO generators (name, description, icon, core_prompt, remark, order_num) VALUES (?, ?, ?, ?, ?, ?)');
    const result = stmt.run(name, description || '', icon || 'Lightning', core_prompt || '', remark || '', order_num || 0);
    const generator = db.prepare('SELECT * FROM generators WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: generator });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/:id', (req, res) => {
  try {
    const { name, description, icon, core_prompt, remark, order_num } = req.body;
    const stmt = db.prepare('UPDATE generators SET name = ?, description = ?, icon = ?, core_prompt = ?, remark = ?, order_num = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(name, description, icon, core_prompt, remark, order_num, req.params.id);
    const generator = db.prepare('SELECT * FROM generators WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: generator });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM generators WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
