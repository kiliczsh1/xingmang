const express = require('express');
const router = express.Router();
const db = require('../database/db');
const {
  buildAuthHeaders,
  requestJson,
  extractResponseContent,
  getErrorMessage
} = require('./apiHelpers');

router.get('/', (req, res) => {
  try {
    const models = db.prepare(`
      SELECT 
        m.*,
        p.name as provider_name,
        p.provider_type,
        p.api_key,
        p.api_url
      FROM api_models m
      LEFT JOIN api_providers p ON m.provider_id = p.id
      ORDER BY m.created_at DESC
    `).all();
    res.json({ success: true, data: models });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/default', (req, res) => {
  try {
    const model = db.prepare(`
      SELECT 
        m.*,
        p.name as provider_name,
        p.provider_type,
        p.api_key,
        p.api_url
      FROM api_models m
      LEFT JOIN api_providers p ON m.provider_id = p.id
      WHERE m.is_default = 1
      LIMIT 1
    `).get();
    res.json({ success: true, data: model });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/provider/:providerId', (req, res) => {
  try {
    const models = db.prepare('SELECT * FROM api_models WHERE provider_id = ? ORDER BY created_at DESC').all(req.params.providerId);
    res.json({ success: true, data: models });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/:id/test', async (req, res) => {
  try {
    const modelConfig = db.prepare(`
      SELECT
        m.*,
        p.name as provider_name,
        p.provider_type,
        p.api_key,
        p.api_url
      FROM api_models m
      LEFT JOIN api_providers p ON m.provider_id = p.id
      WHERE m.id = ?
    `).get(req.params.id);

    if (!modelConfig) {
      return res.status(404).json({ success: false, message: '未找到模型配置' });
    }

    if (!modelConfig.api_url || !modelConfig.api_key) {
      return res.status(400).json({ success: false, message: '请先完善服务商的 API 地址和 API 密钥' });
    }

    const startTime = Date.now();
    const response = await requestJson(
      modelConfig.api_url,
      {
        model: modelConfig.model,
        messages: [
          { role: 'system', content: '你是一个接口连通性测试助手，请仅回复"测试成功"。' },
          { role: 'user', content: '请回复测试成功' }
        ],
        temperature: 0,
        max_tokens: Math.min(Number(modelConfig.max_tokens) || 2000, 64)
      },
      buildAuthHeaders(modelConfig.api_key)
    );

    const content = String(extractResponseContent(response.data) || '').trim();

    res.json({
      success: true,
      data: {
        modelId: modelConfig.id,
        modelName: modelConfig.name,
        providerName: modelConfig.provider_name,
        latencyMs: Date.now() - startTime,
        preview: content.slice(0, 120) || '模型已成功返回结果'
      }
    });
  } catch (error) {
    res.status(500).json({
      success: false,
      message: getErrorMessage(error)
    });
  }
});

router.post('/', (req, res) => {
  try {
    const { provider_id, name, model, temperature, max_tokens, is_default } = req.body;
    
    if (!provider_id || !name || !model) {
      return res.status(400).json({ success: false, message: '请填写完整信息' });
    }
    
    if (is_default) {
      db.prepare('UPDATE api_models SET is_default = 0').run();
    }
    
    const stmt = db.prepare('INSERT INTO api_models (provider_id, name, model, temperature, max_tokens, is_default) VALUES (?, ?, ?, ?, ?, ?)');
    const result = stmt.run(provider_id, name, model, temperature || 0.7, max_tokens || 2000, is_default ? 1 : 0);
    
    const modelWithProvider = db.prepare(`
      SELECT 
        m.*,
        p.name as provider_name,
        p.provider_type,
        p.api_key,
        p.api_url
      FROM api_models m
      LEFT JOIN api_providers p ON m.provider_id = p.id
      WHERE m.id = ?
    `).get(result.lastInsertRowid);
    
    res.json({ success: true, data: modelWithProvider });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/:id', (req, res) => {
  try {
    const { provider_id, name, model, temperature, max_tokens, is_default } = req.body;
    
    if (!provider_id || !name || !model) {
      return res.status(400).json({ success: false, message: '请填写完整信息' });
    }
    
    if (is_default) {
      db.prepare('UPDATE api_models SET is_default = 0').run();
    }
    
    const stmt = db.prepare('UPDATE api_models SET provider_id = ?, name = ?, model = ?, temperature = ?, max_tokens = ?, is_default = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(provider_id, name, model, temperature, max_tokens, is_default ? 1 : 0, req.params.id);
    
    const modelWithProvider = db.prepare(`
      SELECT 
        m.*,
        p.name as provider_name,
        p.provider_type,
        p.api_key,
        p.api_url
      FROM api_models m
      LEFT JOIN api_providers p ON m.provider_id = p.id
      WHERE m.id = ?
    `).get(req.params.id);
    
    res.json({ success: true, data: modelWithProvider });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM api_models WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
