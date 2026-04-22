const express = require('express');
const router = express.Router();
const db = require('../database/db');

// 获取所有服务商
router.get('/', (req, res) => {
  try {
    const providers = db.prepare('SELECT * FROM api_providers ORDER BY created_at DESC').all();
    res.json({ success: true, data: providers });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取服务商详情（包含所有模型）
router.get('/:id', (req, res) => {
  try {
    const provider = db.prepare('SELECT * FROM api_providers WHERE id = ?').get(req.params.id);
    if (!provider) {
      return res.status(404).json({ success: false, message: '服务商不存在' });
    }
    
    const models = db.prepare('SELECT * FROM api_models WHERE provider_id = ? ORDER BY created_at DESC').all(req.params.id);
    res.json({ 
      success: true, 
      data: {
        ...provider,
        models
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建服务商
router.post('/', (req, res) => {
  try {
    const { name, provider_type, api_key, api_url } = req.body;
    
    if (!name || !provider_type || !api_url) {
      return res.status(400).json({ success: false, message: '请填写完整信息' });
    }
    
    const stmt = db.prepare('INSERT INTO api_providers (name, provider_type, api_key, api_url) VALUES (?, ?, ?, ?)');
    const result = stmt.run(name, provider_type, api_key, api_url);
    const provider = db.prepare('SELECT * FROM api_providers WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: provider });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新服务商
router.put('/:id', (req, res) => {
  try {
    const { name, provider_type, api_key, api_url } = req.body;
    
    if (!name || !provider_type || !api_url) {
      return res.status(400).json({ success: false, message: '请填写完整信息' });
    }
    
    const stmt = db.prepare('UPDATE api_providers SET name = ?, provider_type = ?, api_key = ?, api_url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(name, provider_type, api_key, api_url, req.params.id);
    const provider = db.prepare('SELECT * FROM api_providers WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: provider });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除服务商（会级联删除所有关联的模型）
router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM api_providers WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

