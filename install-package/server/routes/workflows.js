const express = require('express');
const db = require('../database/db');

const router = express.Router();

router.get('/', (req, res) => {
  try {
    const rows = db.prepare('SELECT * FROM workflows ORDER BY updated_at DESC').all();
    const workflows = rows.map(row => ({
      id: row.id,
      name: row.name,
      description: row.description,
      scene: row.scene ? JSON.parse(row.scene) : null,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }));
    res.json({ success: true, data: workflows });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const row = db.prepare('SELECT * FROM workflows WHERE id = ?').get(req.params.id);
    if (!row) {
      return res.status(404).json({ success: false, message: '工作流不存在' });
    }
    res.json({
      success: true,
      data: {
        id: row.id,
        name: row.name,
        description: row.description,
        scene: row.scene ? JSON.parse(row.scene) : null,
        createdAt: row.created_at,
        updatedAt: row.updated_at
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/', (req, res) => {
  try {
    const { name, description, scene } = req.body;
    if (!name || !name.trim()) {
      return res.status(400).json({ success: false, message: '名称不能为空' });
    }
    const sceneStr = scene ? JSON.stringify(scene) : '{}';
    const result = db.prepare(
      'INSERT INTO workflows (name, description, scene) VALUES (?, ?, ?)'
    ).run(name.trim(), description || '', sceneStr);
    const row = db.prepare('SELECT * FROM workflows WHERE id = ?').get(result.lastInsertRowid);
    res.json({
      success: true,
      data: {
        id: row.id,
        name: row.name,
        description: row.description,
        scene: row.scene ? JSON.parse(row.scene) : null,
        createdAt: row.created_at,
        updatedAt: row.updated_at
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/:id', (req, res) => {
  try {
    const { name, description, scene } = req.body;
    const existing = db.prepare('SELECT * FROM workflows WHERE id = ?').get(req.params.id);
    if (!existing) {
      return res.status(404).json({ success: false, message: '工作流不存在' });
    }
    const nextName = name !== undefined ? name : existing.name;
    const nextDescription = description !== undefined ? description : existing.description;
    const nextScene = scene !== undefined ? JSON.stringify(scene) : existing.scene;
    db.prepare(
      'UPDATE workflows SET name = ?, description = ?, scene = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?'
    ).run(nextName, nextDescription, nextScene, req.params.id);
    const row = db.prepare('SELECT * FROM workflows WHERE id = ?').get(req.params.id);
    res.json({
      success: true,
      data: {
        id: row.id,
        name: row.name,
        description: row.description,
        scene: row.scene ? JSON.parse(row.scene) : null,
        createdAt: row.created_at,
        updatedAt: row.updated_at
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const existing = db.prepare('SELECT * FROM workflows WHERE id = ?').get(req.params.id);
    if (!existing) {
      return res.status(404).json({ success: false, message: '工作流不存在' });
    }
    db.prepare('DELETE FROM workflows WHERE id = ?').run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/node-library/all', (req, res) => {
  try {
    const rows = db.prepare('SELECT * FROM workflow_node_library ORDER BY category, label').all();
    const items = rows.map(row => ({
      id: row.template_id,
      type: row.type,
      variant: row.variant,
      category: row.category,
      label: row.label,
      description: row.description,
      outputs: row.outputs,
      generated: row.generated === 1,
      template: row.template ? JSON.parse(row.template) : null,
      createdAt: row.created_at,
      updatedAt: row.updated_at
    }));
    res.json({ success: true, data: items });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/node-library', (req, res) => {
  try {
    const { id, type, variant, category, label, description, outputs, generated, template } = req.body;
    if (!id || !type || !category || !label || !template) {
      return res.status(400).json({ success: false, message: '缺少必要字段' });
    }
    const templateStr = JSON.stringify(template);
    const outputsStr = outputs || '';
    db.prepare(
      `INSERT INTO workflow_node_library (template_id, type, variant, category, label, description, outputs, generated, template)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
       ON CONFLICT(template_id) DO UPDATE SET
         type = excluded.type,
         variant = excluded.variant,
         category = excluded.category,
         label = excluded.label,
         description = excluded.description,
         outputs = excluded.outputs,
         generated = excluded.generated,
         template = excluded.template,
         updated_at = CURRENT_TIMESTAMP`
    ).run(id, type, variant || null, category, label, description || null, outputsStr, generated ? 1 : 0, templateStr);
    res.json({ success: true, data: { id } });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/node-library/:id', (req, res) => {
  try {
    db.prepare('DELETE FROM workflow_node_library WHERE template_id = ?').run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
