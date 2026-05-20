const express = require('express');
const router = express.Router();
const db = require('../database/db');
const {
  parseJsonSafely,
  buildAuthHeaders,
  requestJson,
  extractResponseContent,
  getErrorMessage,
  recordUsage,
  getModelConfig
} = require('./apiHelpers');

function stripMarkdownCodeFence(text) {
  return text.replace(/```json\s*/gi, '').replace(/```\s*/g, '').trim();
}

router.get('/entities/book/:bookId', (req, res) => {
  try {
    const { bookId } = req.params;
    const entities = db.prepare('SELECT * FROM graph_entities WHERE book_id = ? ORDER BY updated_at DESC').all(Number(bookId));
    entities.forEach(e => {
      e.metadata = parseJsonSafely(e.metadata);
    });
    res.json({ success: true, data: entities });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/entities/:id', (req, res) => {
  try {
    const entity = db.prepare('SELECT * FROM graph_entities WHERE id = ?').get(req.params.id);
    if (!entity) return res.status(404).json({ success: false, message: '实体不存在' });
    entity.metadata = parseJsonSafely(entity.metadata);
    res.json({ success: true, data: entity });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/entities', (req, res) => {
  try {
    const { book_id, name, type = 'character', description = '', metadata = {} } = req.body;
    if (!book_id || !name) return res.status(400).json({ success: false, message: '缺少必填字段' });
    const result = db.prepare(
      'INSERT INTO graph_entities (book_id, name, type, description, metadata) VALUES (?, ?, ?, ?, ?)'
    ).run(book_id, name, type, description, JSON.stringify(metadata));
    const entity = db.prepare('SELECT * FROM graph_entities WHERE id = ?').get(result.lastInsertRowid);
    entity.metadata = parseJsonSafely(entity.metadata);
    res.json({ success: true, data: entity });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/entities/:id', (req, res) => {
  try {
    const { name, type, description, metadata } = req.body;
    const existing = db.prepare('SELECT * FROM graph_entities WHERE id = ?').get(req.params.id);
    if (!existing) return res.status(404).json({ success: false, message: '实体不存在' });
    db.prepare(
      'UPDATE graph_entities SET name = ?, type = ?, description = ?, metadata = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?'
    ).run(
      name ?? existing.name,
      type ?? existing.type,
      description ?? existing.description,
      metadata !== undefined ? JSON.stringify(metadata) : existing.metadata,
      req.params.id
    );
    const entity = db.prepare('SELECT * FROM graph_entities WHERE id = ?').get(req.params.id);
    entity.metadata = parseJsonSafely(entity.metadata);
    res.json({ success: true, data: entity });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/entities/:id', (req, res) => {
  try {
    db.prepare('DELETE FROM graph_relations WHERE source_id = ? OR target_id = ?').run(req.params.id, req.params.id);
    db.prepare('DELETE FROM graph_entities WHERE id = ?').run(req.params.id);
    res.json({ success: true });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/relations/book/:bookId', (req, res) => {
  try {
    const { bookId } = req.params;
    const relations = db.prepare('SELECT * FROM graph_relations WHERE book_id = ? ORDER BY updated_at DESC').all(Number(bookId));
    res.json({ success: true, data: relations });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/relations', (req, res) => {
  try {
    const { book_id, source_id, target_id, relation_type = 'related', description = '', evidence = '', confidence = 1.0 } = req.body;
    if (!book_id || !source_id || !target_id) return res.status(400).json({ success: false, message: '缺少必填字段' });
    const result = db.prepare(
      'INSERT INTO graph_relations (book_id, source_id, target_id, relation_type, description, evidence, confidence) VALUES (?, ?, ?, ?, ?, ?, ?)'
    ).run(book_id, source_id, target_id, relation_type, description, evidence, confidence);
    const relation = db.prepare('SELECT * FROM graph_relations WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: relation });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/relations/:id', (req, res) => {
  try {
    const { relation_type, description, evidence, confidence } = req.body;
    const existing = db.prepare('SELECT * FROM graph_relations WHERE id = ?').get(req.params.id);
    if (!existing) return res.status(404).json({ success: false, message: '关系不存在' });
    db.prepare(
      'UPDATE graph_relations SET relation_type = ?, description = ?, evidence = ?, confidence = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?'
    ).run(
      relation_type ?? existing.relation_type,
      description ?? existing.description,
      evidence ?? existing.evidence,
      confidence ?? existing.confidence,
      req.params.id
    );
    const relation = db.prepare('SELECT * FROM graph_relations WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: relation });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/relations/:id', (req, res) => {
  try {
    db.prepare('DELETE FROM graph_relations WHERE id = ?').run(req.params.id);
    res.json({ success: true });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/graph/:bookId', (req, res) => {
  try {
    const { bookId } = req.params;
    const entities = db.prepare('SELECT * FROM graph_entities WHERE book_id = ?').all(Number(bookId));
    const relations = db.prepare('SELECT * FROM graph_relations WHERE book_id = ?').all(Number(bookId));
    entities.forEach(e => {
      e.metadata = parseJsonSafely(e.metadata);
    });
    res.json({ success: true, data: { entities, relations } });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/graph/:bookId', (req, res) => {
  try {
    const { bookId } = req.params;
    db.prepare('DELETE FROM graph_relations WHERE book_id = ?').run(Number(bookId));
    db.prepare('DELETE FROM graph_entities WHERE book_id = ?').run(Number(bookId));
    res.json({ success: true });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/analyze', async (req, res) => {
  try {
    const { bookId, configId, scope = 'chapters', chapterIds = [], versionId } = req.body;

    if (!bookId) {
      return res.status(400).json({ success: false, message: '缺少 bookId' });
    }
    
    if (!versionId) {
      return res.status(400).json({ success: false, message: '请先选择或创建一个图谱文件' });
    }

    const config = getModelConfig(configId);
    if (!config) {
      return res.status(400).json({ success: false, message: '未找到 API 配置，请先配置 AI 模型' });
    }

    let textContent = '';
    if (chapterIds.length > 0) {
      const placeholders = chapterIds.map(() => '?').join(',');
      const chapters = db.prepare(`SELECT title, content FROM chapters WHERE id IN (${placeholders}) AND book_id = ?`).all(...chapterIds, bookId);
      textContent = chapters.map(c => `【${c.title}】\n${c.content || ''}`).join('\n\n');
    } else {
      const chapters = db.prepare('SELECT title, content FROM chapters WHERE book_id = ?').all(bookId);
      textContent = chapters.map(c => `【${c.title}】\n${c.content || ''}`).join('\n\n');
    }

    if (!textContent.trim()) {
      return res.status(400).json({ success: false, message: '没有可分析的文本内容' });
    }

    const truncatedText = textContent.length > 30000 ? textContent.substring(0, 30000) + '\n...(文本过长已截断)' : textContent;

    const systemPrompt = `你是一个专业的小说知识图谱分析助手。请从给定的小说文本中提取实体和关系，构建知识图谱。

【重要要求】
- 所有输出内容（实体名称、关系类型、描述、证据）必须全部使用中文
- 关系类型字段请使用中文，例如：盟友、敌对、师徒、亲属、位于、属于、参与、拥有、竞争对手、下属、相关
- 所有实体必有关系，严禁出现独立实体
实体类型包括：
- （角色）
- （地点）
- （物件/法宝）
- （势力/组织）
- （重要事件）
- （功法/技能）
- （线索/秘密）

关系类型包括：
- （盟友）
- （敌对）
- （师徒）
- family（亲属）
- located_at（位于）
- belongs_to（属于）
- participates（参与）
- owns（拥有）
- rival（竞争对手）
- subordinate（下属）
- related（相关）

请只返回 JSON，格式如下：
{
  "entities": [
    {
      "name": "实体名称",
      "type": "character",
      "description": "简要描述"
    }
  ],
  "relations": [
    {
      "source": "源实体名称",
      "target": "目标实体名称",
      "relation_type": "盟友",
      "description": "关系描述",
      "evidence": "原文依据"
    }
  ]
}

注意：
- 只返回 JSON，不要包含额外说明
- 提取所有有意义的实体和关系
- 实体名称要准确，保持一致
- 关系描述要简洁明了但要全面
- evidence 尽量引用原文关键语句
- 所有内容必须使用中文`;

    const response = await requestJson(
      config.api_url,
      {
        model: config.model,
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: truncatedText }
        ],
        temperature: 0.3,
        max_tokens: Math.min(Math.max(Number(config.max_tokens) || 2000, 1), 8192)
      },
      buildAuthHeaders(config.api_key)
    );

    const content = extractResponseContent(response.data);
    const parsed = parseJsonSafely(stripMarkdownCodeFence(content));

    if (!parsed || typeof parsed !== 'object') {
      return res.status(500).json({ success: false, message: 'AI 返回的格式不正确，请重试' });
    }

    const entities = Array.isArray(parsed.entities) ? parsed.entities : [];
    const relations = Array.isArray(parsed.relations) ? parsed.relations : [];

    const insertEntity = db.prepare(
      'INSERT INTO graph_entities (book_id, version_id, name, type, description, metadata) VALUES (?, ?, ?, ?, ?, ?)'
    );
    const insertRelation = db.prepare(
      'INSERT INTO graph_relations (book_id, version_id, source_id, target_id, relation_type, description, evidence, confidence) VALUES (?, ?, ?, ?, ?, ?, ?, ?)'
    );

    // 删除该版本的旧数据
    db.prepare('DELETE FROM graph_relations WHERE version_id = ?').run(versionId);
    db.prepare('DELETE FROM graph_entities WHERE version_id = ?').run(versionId);

    const entityMap = new Map();

    const insertAll = db.transaction(() => {
      for (const entity of entities) {
        if (!entity.name) continue;
        const result = insertEntity.run(
          bookId,
          versionId,
          entity.name,
          entity.type || 'character',
          entity.description || '',
          JSON.stringify({})
        );
        entityMap.set(entity.name, Number(result.lastInsertRowid));
      }

      for (const relation of relations) {
        const sourceId = entityMap.get(relation.source);
        const targetId = entityMap.get(relation.target);
        if (!sourceId || !targetId) continue;
        if (sourceId === targetId) continue;
        insertRelation.run(
          bookId,
          versionId,
          sourceId,
          targetId,
          relation.relation_type || 'related',
          relation.description || '',
          relation.evidence || '',
          1.0
        );
      }
    });

    insertAll();

    recordUsage(config.id, config.name, config.provider_name);

    // 查询保存的实体和关系
    const savedEntities = db.prepare('SELECT * FROM graph_entities WHERE version_id = ?').all(versionId);
    const savedRelations = db.prepare('SELECT * FROM graph_relations WHERE version_id = ?').all(versionId);

    // 更新版本记录
    const chapterCount = chapterIds.length || db.prepare('SELECT COUNT(*) as count FROM chapters WHERE book_id = ?').get(bookId).count;
    db.prepare(`
      UPDATE graph_versions 
      SET entity_count = ?, relation_count = ?, chapter_count = ?, scope = ?, description = ?
      WHERE id = ?
    `).run(
      savedEntities.length,
      savedRelations.length,
      chapterCount,
      chapterIds.length > 0 ? 'chapters' : 'all',
      chapterIds.length > 0 ? `基于 ${chapterIds.length} 个章节的分析` : '基于全书的分析',
      versionId
    );

    const savedVersions = db.prepare('SELECT * FROM graph_versions WHERE book_id = ? ORDER BY created_at DESC').all(bookId);

    res.json({
      success: true,
      data: {
        entities: savedEntities,
        relations: savedRelations,
        stats: {
          entityCount: savedEntities.length,
          relationCount: savedRelations.length
        },
        version: {
          id: versionId,
          versions: savedVersions
        }
      }
    });
  } catch (error) {
    console.error('Graph analyze failed:', error.message);
    res.status(500).json({ success: false, message: getErrorMessage(error) });
  }
});

// 获取图谱版本列表
router.get('/versions/:bookId', (req, res) => {
  try {
    const { bookId } = req.params;
    const versions = db.prepare('SELECT * FROM graph_versions WHERE book_id = ? ORDER BY created_at DESC').all(Number(bookId));
    res.json({ success: true, data: versions });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除图谱版本
router.delete('/versions/:bookId/:versionId', (req, res) => {
  try {
    const { bookId, versionId } = req.params;
    
    // 删除该版本的关系
    db.prepare('DELETE FROM graph_relations WHERE version_id = ?').run(Number(versionId));
    
    // 删除该版本的实体
    db.prepare('DELETE FROM graph_entities WHERE version_id = ?').run(Number(versionId));
    
    // 删除版本记录
    db.prepare('DELETE FROM graph_versions WHERE id = ?').run(Number(versionId));
    
    res.json({ success: true });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建空白图谱版本
router.post('/versions', (req, res) => {
  try {
    const { bookId, name } = req.body;
    if (!bookId) return res.status(400).json({ success: false, message: '缺少 bookId' });
    
    const versionName = name || `新图谱_${new Date().toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}`;
    
    const result = db.prepare(`
      INSERT INTO graph_versions (book_id, name, description, scope, chapter_count, entity_count, relation_count)
      VALUES (?, ?, '', 'manual', 0, 0, 0)
    `).run(bookId, versionName);
    
    const version = db.prepare('SELECT * FROM graph_versions WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: version });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 重命名图谱版本
router.put('/versions/:versionId', (req, res) => {
  try {
    const { versionId } = req.params;
    const { name } = req.body;
    
    db.prepare('UPDATE graph_versions SET name = ? WHERE id = ?').run(name, Number(versionId));
    const version = db.prepare('SELECT * FROM graph_versions WHERE id = ?').get(Number(versionId));
    res.json({ success: true, data: version });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取指定版本的图谱数据
router.get('/version/:versionId', (req, res) => {
  try {
    const { versionId } = req.params;
    
    const entities = db.prepare('SELECT * FROM graph_entities WHERE version_id = ?').all(Number(versionId));
    entities.forEach(e => {
      e.metadata = parseJsonSafely(e.metadata);
    });
    
    const relations = db.prepare('SELECT * FROM graph_relations WHERE version_id = ?').all(Number(versionId));
    
    res.json({
      success: true,
      data: {
        entities,
        relations,
        stats: {
          entityCount: entities.length,
          relationCount: relations.length
        }
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
