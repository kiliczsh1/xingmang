const express = require('express');
const router = express.Router();
const db = require('../database/db');
const bookStorage = require('../services/bookStorage');
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

router.post('/graph/import', (req, res) => {
  try {
    const { bookId, entities, relations } = req.body;
    if (!bookId || !Array.isArray(entities)) {
      return res.status(400).json({ success: false, message: '缺少必要参数' });
    }

    const numBookId = Number(bookId);

    const doImport = db.transaction(() => {
      db.prepare('DELETE FROM graph_relations WHERE book_id = ?').run(numBookId);
      db.prepare('DELETE FROM graph_entities WHERE book_id = ?').run(numBookId);

      const insertEntity = db.prepare(
        'INSERT INTO graph_entities (book_id, name, type, description, metadata) VALUES (?, ?, ?, ?, ?)'
      );
      const insertRelation = db.prepare(
        'INSERT INTO graph_relations (book_id, source_id, target_id, relation_type, description, evidence, confidence) VALUES (?, ?, ?, ?, ?, ?, ?)'
      );

      const oldToNew = new Map();
      for (const e of entities) {
        const result = insertEntity.run(
          numBookId,
          e.name || '',
          e.type || 'character',
          e.description || '',
          JSON.stringify(typeof e.metadata === 'object' ? e.metadata : {})
        );
        oldToNew.set(Number(e.id), Number(result.lastInsertRowid));
      }

      for (const r of relations || []) {
        const newSourceId = oldToNew.get(Number(r.source_id));
        const newTargetId = oldToNew.get(Number(r.target_id));
        if (!newSourceId || !newTargetId) continue;
        insertRelation.run(
          numBookId,
          newSourceId,
          newTargetId,
          r.relation_type || '相关',
          r.description || '',
          r.evidence || '',
          typeof r.confidence === 'number' ? r.confidence : 1.0
        );
      }
    });

    doImport();

    const savedEntities = db.prepare('SELECT * FROM graph_entities WHERE book_id = ?').all(numBookId);
    const savedRelations = db.prepare('SELECT * FROM graph_relations WHERE book_id = ?').all(numBookId);
    savedEntities.forEach(e => {
      e.metadata = parseJsonSafely(e.metadata);
    });

    res.json({ success: true, data: { entities: savedEntities, relations: savedRelations } });
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

    const config = getModelConfig(configId);
    if (!config) {
      return res.status(400).json({ success: false, message: '未找到 API 配置，请先配置 AI 模型' });
    }

    let actualVersionId = versionId;

    if (!actualVersionId) {
      const existingVersions = db.prepare('SELECT * FROM graph_versions WHERE book_id = ? ORDER BY created_at DESC LIMIT 1').all(bookId);
      if (existingVersions.length > 0) {
        actualVersionId = existingVersions[0].id;
      } else {
        const createResult = db.prepare(`
          INSERT INTO graph_versions (book_id, name, description, scope, chapter_count, entity_count, relation_count)
          VALUES (?, '默认图谱', '', 'auto', 0, 0, 0)
        `).run(bookId);
        actualVersionId = createResult.lastInsertRowid;
      }
    }

    const chapters = bookStorage.getChaptersForGraph(Number(bookId), chapterIds);
    let textContent = chapters.map(c => `【${c.title}】\n${c.content || ''}`).join('\n\n');

    if (!textContent.trim()) {
      return res.status(400).json({ success: false, message: '没有可分析的文本内容' });
    }

    const truncatedText = textContent.length > 30000 ? textContent.substring(0, 30000) + '\n...(文本过长已截断)' : textContent;

    const systemPrompt = `你是一个专业的小说知识图谱分析助手。你的任务是从给定的小说文本中提取实体和关系，构建以角色为中心的辐射状知识图谱。

【核心原则】
1. 所有输出必须使用中文（包括实体名称、关系类型、描述、证据等字段）
2. 每个提取的实体必须至少有一个关系，禁止出现孤立实体
3. 实体名称要准确一致，同一人物/地点在不同章节中名称必须统一
4. 优先提取重要信息，避免过度细碎的次要内容

【⭐ 核心要求：以角色为中心的关系网络】
图谱结构应该呈现"以角色为核心，向外辐射"的层次：

第一层（核心层）：角色（character）
- 优先提取主要角色（主角、重要配角、反派）
- 重点建立角色之间的关系网络

第二层（关联层）：与角色直接相关的实体
- 地点：角色活动的主要场所、居住地
- 势力：角色所属或对立的组织门派
- 物件：角色拥有的重要法宝武器
- 功法：角色修炼的技能功法

第三层（外围层）：间接相关的事件和线索
- 事件：角色参与的重要情节
- 线索：与角色命运相关的秘密伏笔

【关系提取策略】
1. 角色间关系优先：首先确保每个主要角色至少有2-3条与其他角色的关系
2. 从角色出发的关系：source 尽量使用角色名称，target 使用其他类型实体
3. 避免非角色实体间的直接关系：地点A → 地点B 应该改为 角色→地点A 和 角色→地点B
4. 关系方向性：主动方作为 source（如"张三拥有宝剑"，而非"宝剑属于张三"）

【实体类型定义】
请严格按照以下7种类型分类，type 字段使用英文标识。
【⭐ 重要】每个实体必须写出详尽描述，不要只写一句话，要写清楚该实体的关键信息：

1. "character" - 角色：小说中的人物、NPC、主角配角等【最重要】
   示例：张三、李四、王五、神秘老者
   提取建议：优先提取有对话、行动、情感描写的人物
   描述要求：姓名、身份地位、性格特征、核心能力、行为动机、与其他角色的关系概要

2. "location" - 地点：场景、建筑、区域、世界等
   示例：天剑宗、秘境、京城、修炼洞府
   提取建议：只提取角色实际活动的重要场所
   描述要求：环境特征、地理方位、功能用途、发生在此地的关键事件、所属势力

3. "item" - 物件：法宝、武器、丹药、道具、宝物等
   示例：诛仙剑、九转还魂丹、储物戒指
   提取建议：只提取被角色拥有或使用的关键物品
   描述要求：外观形态、功能能力、来源出处、当前持有者、在剧情中的作用

4. "faction" - 势力：组织、门派、家族、国家、联盟等
   示例：魔道联盟、皇室、李家
   提取建议：只提取角色所属或敌对的组织
   描述要求：规模实力、组织宗旨或目标、核心成员、与其他势力的关系、在故事中的地位

5. "event" - 事件：重大情节、历史事件、战斗、仪式等
   示例：正魔大战、宗门大比、飞升大典
   提取建议：只提取角色直接参与的转折性事件
   描述要求：事件起因、关键经过、最终结果或影响、参与的主要角色

6. "skill" - 功法：技能、功法、武技、法术、秘术等
   示例：九阳神功、御剑术、焚天诀
   提取建议：只提取角色修炼的核心技能
   描述要求：功法等级威力、修炼条件或门槛、传承来源、掌握该功法的角色、在战斗中的表现

7. "clue" - 线索：秘密、谜团、伏笔、隐藏信息等
   示例：身世之谜、上古遗迹位置、叛徒身份
   提取建议：只提取影响角色命运的关键线索
   描述要求：线索来源、指向的目标或人物、重要性程度、与哪些角色相关

【关系类型定义】
relation_type 字段必须使用以下中文关键词之一：

人际关系类（最高优先级）：
- "师徒" - 师父与徒弟的关系
- "亲属" - 血亲、姻亲等家庭关系
- "盟友" - 合作伙伴、同盟
- "敌对" - 对手、敌人、仇人
- "竞争对手" - 竞争关系但不一定是敌人
- "下属" - 上下级从属关系
- "朋友" - 友情关系

空间归属类：
- "位于" - 角色位于某地（推荐：角色 → 地点）
- "属于" - 归属关系（推荐：角色属于势力）

行为参与类：
- "拥有" - 角色拥有某物（推荐：角色 → 物件/功法）
- "参与" - 角色参与事件（推荐：角色 → 事件）
- "相关" - 其他关联关系（兜底类型）

特殊关系：
- "传承" - 师徒间的功法传承
- "守护" - 保护关系

【输出格式要求】
只返回严格的 JSON 格式，不要包含任何额外说明文字：

{
  "entities": [
    {
      "name": "实体名称",
      "type": "character|location|item|faction|event|skill|clue",
      "description": "详细描述该实体的特征、背景、作用或意义，至少50个字以上，角色需包含身份、性格、能力、动机；地点需包含环境特色、功能作用；物件需包含外观、能力、来源；势力需包含规模、宗旨、成员构成；事件需包含起因、经过、影响；功法需包含威力、修炼条件、传承背景；线索需包含来源、指向、重要性"
    }
  ],
  "relations": [
    {
      "source": "源实体名称（必须在entities中存在）",
      "target": "目标实体名称（必须在entities中存在）",
      "relation_type": "上述关系类型之一",
      "description": "关系的具体说明（30字以内）",
      "evidence": "原文中的关键语句作为证据"
    }
  ]
}

【质量标准】
- 实体数量比例：角色占40-50%，其他类型共占50-60%
- 关系分布：角色间关系≥40%，角色到其他实体的关系≥50%
- 提取总量：根据文本长度合理提取，一般6-12个实体，10-20条关系
- 连接度：每个角色至少连接2-3个其他实体（角色或非角色）
- 去重原则：同一对实体间相同类型的关系只保留最重要的一条
- 证据引用：evidence 必须是原文的真实语句，不要编造
- 【⭐ 描述详实度】：每个实体的 description 至少50字，要写出该实体的多个维度信息（身份、特征、能力、关联等），不允许只写一句简短描述就敷衍了事`;

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

    const existingEntities = db.prepare('SELECT * FROM graph_entities WHERE version_id = ?').all(actualVersionId);
    const existingRelations = db.prepare('SELECT * FROM graph_relations WHERE version_id = ?').all(actualVersionId);

    const entityMapByName = new Map();
    existingEntities.forEach(e => {
      const key = `${e.name}_${e.type}`;
      entityMapByName.set(key, e);
    });

    const entityNameToIdMap = new Map();
    existingEntities.forEach(e => {
      entityNameToIdMap.set(e.name, e.id);
    });

    const relationKeySet = new Set();
    existingRelations.forEach(r => {
      const key = `${r.source_id}_${r.target_id}_${r.relation_type}`;
      relationKeySet.add(key);
    });

    let newEntityCount = 0;
    let newRelationCount = 0;

    const insertAll = db.transaction(() => {
      for (const entity of entities) {
        if (!entity.name) continue;
        const entityType = entity.type || 'character';
        const key = `${entity.name}_${entityType}`;

        if (entityMapByName.has(key)) {
          continue;
        }

        const result = insertEntity.run(
          bookId,
          actualVersionId,
          entity.name,
          entityType,
          entity.description || '',
          JSON.stringify({})
        );
        entityNameToIdMap.set(entity.name, Number(result.lastInsertRowid));
        newEntityCount++;
      }

      for (const relation of relations) {
        const sourceId = entityNameToIdMap.get(relation.source);
        const targetId = entityNameToIdMap.get(relation.target);
        if (!sourceId || !targetId) continue;
        if (sourceId === targetId) continue;

        const relKey = `${sourceId}_${targetId}_${relation.relation_type || 'related'}`;
        if (relationKeySet.has(relKey)) {
          continue;
        }

        insertRelation.run(
          bookId,
          actualVersionId,
          sourceId,
          targetId,
          relation.relation_type || '相关',
          relation.description || '',
          relation.evidence || '',
          1.0
        );
        newRelationCount++;
      }
    });

    insertAll();

    recordUsage(config.id, config.name, config.provider_name);

    const savedEntities = db.prepare('SELECT * FROM graph_entities WHERE version_id = ?').all(actualVersionId);
    const savedRelations = db.prepare('SELECT * FROM graph_relations WHERE version_id = ?').all(actualVersionId);

    const chapterCount = chapterIds.length || chapters.length;
    db.prepare(`
      UPDATE graph_versions
      SET entity_count = ?, relation_count = ?, chapter_count = ?, scope = ?, description = ?
      WHERE id = ?
    `).run(
      savedEntities.length,
      savedRelations.length,
      chapterCount,
      chapterIds.length > 0 ? 'chapters' : 'all',
      `累计 ${savedEntities.length} 实体，${savedRelations.length} 关系（本次新增 ${newEntityCount} 实体，${newRelationCount} 关系）`,
      actualVersionId
    );

    const savedVersions = db.prepare('SELECT * FROM graph_versions WHERE book_id = ? ORDER BY created_at DESC').all(bookId);

    res.json({
      success: true,
      data: {
        entities: savedEntities,
        relations: savedRelations,
        stats: {
          entityCount: savedEntities.length,
          relationCount: savedRelations.length,
          newEntityCount,
          newRelationCount
        },
        version: {
          id: actualVersionId,
          versions: savedVersions
        }
      }
    });
  } catch (error) {
    console.error('Graph analyze failed:', error.message);
    res.status(500).json({ success: false, message: getErrorMessage(error) });
  }
});

// AI 补充描述
router.post('/ai-supplement', async (req, res) => {
  try {
    const { bookId, configId, entityName, entityType, entityDescription, hint, chapterIds } = req.body;

    if (!bookId || !configId || !entityName) {
      return res.status(400).json({ success: false, message: '缺少必要参数' });
    }

    const config = getModelConfig(configId);
    if (!config) {
      return res.status(400).json({ success: false, message: '未找到 API 配置' });
    }

    let textContent = '';
    if (chapterIds && chapterIds.length > 0) {
      const chapters = bookStorage.getChaptersForGraph(Number(bookId), chapterIds);
      textContent = chapters.map(c => `【${c.title}】\n${c.content || ''}`).join('\n\n');
    }

    const typeLabelMap = {
      character: '角色', location: '地点', item: '物件',
      faction: '势力', event: '事件', skill: '功法', clue: '线索'
    };
    const typeLabel = typeLabelMap[entityType] || entityType;

    const systemPrompt = `你是一个专业的网文写作辅助助手。你的任务是根据用户提供的文本内容，对指定的小说${typeLabel}（名称：「${entityName}」）进行详细补充描述。

要求：
1. 结合原文中的相关情节，写出该${typeLabel}的详细背景、特征、作用等信息
2. 语言风格应贴合网文语境，内容详实丰富，至少100字
3. ${hint ? `请着重从以下角度补充：${hint}` : '请从多维度进行补充'}
4. 直接输出纯文本描述内容，不要包含任何格式标记、前缀说明`;

    const userMessages = [
      { role: 'system', content: systemPrompt },
      { role: 'user', content: `请对以下${typeLabel}进行补充描述：\n\n当前描述：${entityDescription || '无'}` }
    ];

    if (textContent) {
      userMessages.push({ role: 'user', content: `以下是相关章节内容：\n\n${textContent.substring(0, 15000)}` });
    }

    const response = await requestJson(
      config.api_url,
      {
        model: config.model,
        messages: userMessages,
        temperature: 0.7,
        max_tokens: 1500
      },
      buildAuthHeaders(config.api_key)
    );

    const content = extractResponseContent(response.data);
    recordUsage(config.id, config.name, config.provider_name);

    res.json({ success: true, data: { content } });
  } catch (error) {
    console.error('AI supplement failed:', error.message);
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
