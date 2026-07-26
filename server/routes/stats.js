const express = require('express');
const router = express.Router();
const db = require('../database/db');

router.get('/overview', (req, res) => {
  try {
    const promptResult = db.prepare('SELECT COUNT(*) as count FROM prompts').get();
    const generatorResult = db.prepare('SELECT COUNT(*) as count FROM generators').get();
    const bookResult = db.prepare('SELECT COUNT(*) as count FROM books').get();
    const usageResult = db.prepare('SELECT COALESCE(SUM(usage_count), 0) as count FROM usage_stats').get();
    const tokensResult = db.prepare('SELECT COALESCE(SUM(total_tokens), 0) as count FROM usage_stats').get();

    const promptCount = promptResult ? promptResult.count : 0;
    const generatorCount = generatorResult ? generatorResult.count : 0;
    const bookCount = bookResult ? bookResult.count : 0;
    const totalUsageCount = usageResult ? usageResult.count : 0;
    const totalTokens = tokensResult ? tokensResult.count : 0;

    res.json({
      success: true,
      data: {
        promptCount,
        generatorCount,
        bookCount,
        totalUsageCount,
        totalTokens
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/daily', (req, res) => {
  try {
    const { startDate, endDate } = req.query;
    const today = new Date().toISOString().split('T')[0];
    const start = startDate || today;
    const end = endDate || today;

    const stats = db.prepare(`
      SELECT 
        date,
        model_id,
        model_name,
        provider_name,
        SUM(usage_count) as usage_count,
        SUM(total_tokens) as total_tokens
      FROM usage_stats
      WHERE date BETWEEN ? AND ?
      GROUP BY date, model_id
      ORDER BY date DESC, usage_count DESC
    `).all(start, end);

    res.json({
      success: true,
      data: stats
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/model-stats', (req, res) => {
  try {
    const stats = db.prepare(`
      SELECT 
        model_id,
        model_name,
        provider_name,
        SUM(usage_count) as total_usage,
        SUM(total_tokens) as total_tokens
      FROM usage_stats
      GROUP BY model_id
      ORDER BY total_usage DESC
    `).all();

    res.json({
      success: true,
      data: stats
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/monthly', (req, res) => {
  try {
    const stats = db.prepare(`
      SELECT 
        strftime('%Y-%m', date) as month,
        SUM(usage_count) as total_usage,
        SUM(total_tokens) as total_tokens,
        COUNT(DISTINCT model_id) as model_count
      FROM usage_stats
      GROUP BY strftime('%Y-%m', date)
      ORDER BY month DESC
      LIMIT 12
    `).all();

    res.json({
      success: true,
      data: stats
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/record', (req, res) => {
  try {
    const { modelId, modelName, providerName, tokens } = req.body;
    const today = new Date().toISOString().split('T')[0];

    const existing = db.prepare(`
      SELECT * FROM usage_stats WHERE date = ? AND model_id = ?
    `).get(today, modelId);

    if (existing) {
      db.prepare(`
        UPDATE usage_stats 
        SET usage_count = usage_count + 1, 
            total_tokens = total_tokens + ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE date = ? AND model_id = ?
      `).run(tokens || 0, today, modelId);
    } else {
      db.prepare(`
        INSERT INTO usage_stats (date, model_id, model_name, provider_name, usage_count, total_tokens)
        VALUES (?, ?, ?, ?, 1, ?)
      `).run(today, modelId, modelName, providerName, tokens || 0);
    }

    res.json({ success: true, message: '统计记录成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
