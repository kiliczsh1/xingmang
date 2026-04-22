const express = require('express');
const router = express.Router();
const db = require('../database/db');

const ensureDescriptionColumn = () => {
  const columns = db.prepare("PRAGMA table_info(prompts)").all();
  const hasDescription = columns.some(column => column.name === 'description');
  if (!hasDescription) {
    db.exec('ALTER TABLE prompts ADD COLUMN description TEXT');
  }
  const hasCardType = columns.some(column => column.name === 'card_type');
  if (!hasCardType) {
    db.exec("ALTER TABLE prompts ADD COLUMN card_type TEXT DEFAULT 'normal'");
  }
  const hasPassword = columns.some(column => column.name === 'password');
  if (!hasPassword) {
    db.exec('ALTER TABLE prompts ADD COLUMN password TEXT');
  }
};

// 获取所有提示词
router.get('/', (req, res) => {
  try {
    ensureDescriptionColumn();
    const prompts = db.prepare('SELECT * FROM prompts ORDER BY order_num ASC').all();
    // 解析每个提示词的fields和subcategories字段
    prompts.forEach(prompt => {
      if (prompt.fields) {
        try {
          prompt.fields = JSON.parse(prompt.fields);
        } catch (error) {
          prompt.fields = null;
        }
      }
      if (prompt.subcategories) {
        try {
          prompt.subcategories = JSON.parse(prompt.subcategories);
        } catch (error) {
          prompt.subcategories = null;
        }
      }
      if (!prompt.card_type) {
        prompt.card_type = 'normal';
      }
    });
    res.json({ success: true, data: prompts });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个提示词
router.get('/:id', (req, res) => {
  try {
    ensureDescriptionColumn();
    const prompt = db.prepare('SELECT * FROM prompts WHERE id = ?').get(req.params.id);
    if (!prompt) {
      return res.status(404).json({ success: false, message: '提示词不存在' });
    }
    // 解析fields字段
    if (prompt.fields) {
      try {
        prompt.fields = JSON.parse(prompt.fields);
      } catch (error) {
        prompt.fields = null;
      }
    }
    // 解析subcategories
    if (prompt.subcategories) {
      try {
        prompt.subcategories = JSON.parse(prompt.subcategories);
      } catch (error) {
        prompt.subcategories = null;
      }
    }
    if (!prompt.card_type) {
      prompt.card_type = 'normal';
    }
    res.json({ success: true, data: prompt });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建提示词
router.post('/', (req, res) => {
  try {
    ensureDescriptionColumn();
    const { name, description, content, category, order_num, fields, subcategories, card_type, password } = req.body;
    const stmt = db.prepare('INSERT INTO prompts (name, description, content, category, order_num, fields, subcategories, card_type, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)');
    const result = stmt.run(name, description || null, content, category || '默认', order_num || 0, fields ? JSON.stringify(fields) : null, subcategories ? JSON.stringify(subcategories) : null, card_type || 'normal', password || null);
    const prompt = db.prepare('SELECT * FROM prompts WHERE id = ?').get(result.lastInsertRowid);
    // 解析fields字段
    if (prompt.fields) {
      prompt.fields = JSON.parse(prompt.fields);
    }
    // 解析subcategories字段
    if (prompt.subcategories) {
      try {
        prompt.subcategories = JSON.parse(prompt.subcategories);
      } catch (error) {
        prompt.subcategories = null;
      }
    }
    if (!prompt.card_type) {
      prompt.card_type = 'normal';
    }
    res.json({ success: true, data: prompt });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新提示词
router.put('/:id', (req, res) => {
  try {
    ensureDescriptionColumn();
    const { name, description, content, category, order_num, fields, subcategories, card_type, password } = req.body;
    const stmt = db.prepare('UPDATE prompts SET name = ?, description = ?, content = ?, category = ?, order_num = ?, fields = ?, subcategories = ?, card_type = ?, password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(name, description || null, content, category, order_num, fields ? JSON.stringify(fields) : null, subcategories ? JSON.stringify(subcategories) : null, card_type || 'normal', password !== undefined ? password : null, req.params.id);
    const prompt = db.prepare('SELECT * FROM prompts WHERE id = ?').get(req.params.id);
    // 解析fields字段
    if (prompt.fields) {
      try {
        prompt.fields = JSON.parse(prompt.fields);
      } catch (error) {
        prompt.fields = null;
      }
    }
    // 解析subcategories字段
    if (prompt.subcategories) {
      try {
        prompt.subcategories = JSON.parse(prompt.subcategories);
      } catch (error) {
        prompt.subcategories = null;
      }
    }
    if (!prompt.card_type) {
      prompt.card_type = 'normal';
    }
    res.json({ success: true, data: prompt });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除提示词
router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM prompts WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

