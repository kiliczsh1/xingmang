const express = require('express');
const router = express.Router();
const db = require('../database/db');

// 获取所有对话
router.get('/', (req, res) => {
  try {
    const conversations = db.prepare(`
      SELECT c.*, 
        (SELECT COUNT(*) FROM conversation_messages WHERE conversation_id = c.id) as message_count
      FROM conversations c 
      ORDER BY c.updated_at DESC
    `).all();
    res.json({ success: true, data: conversations });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取书本的所有对话
router.get('/book/:bookId', (req, res) => {
  try {
    const conversations = db.prepare(`
      SELECT c.*, 
        (SELECT COUNT(*) FROM conversation_messages WHERE conversation_id = c.id) as message_count
      FROM conversations c 
      WHERE c.book_id = ? 
      ORDER BY c.updated_at DESC
    `).all(req.params.bookId);
    res.json({ success: true, data: conversations });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建对话
router.post('/', (req, res) => {
  try {
    const { book_id, title } = req.body;
    const stmt = db.prepare('INSERT INTO conversations (book_id, title) VALUES (?, ?)');
    const result = stmt.run(book_id, title || '新对话');
    const conversation = db.prepare('SELECT * FROM conversations WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: conversation });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新对话标题
router.put('/:id', (req, res) => {
  try {
    const { title } = req.body;
    const stmt = db.prepare('UPDATE conversations SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?');
    stmt.run(title, req.params.id);
    const conversation = db.prepare('SELECT * FROM conversations WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: conversation });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除对话
router.delete('/:id', (req, res) => {
  try {
    const stmt = db.prepare('DELETE FROM conversations WHERE id = ?');
    stmt.run(req.params.id);
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取对话消息
router.get('/:id/messages', (req, res) => {
  try {
    const messages = db.prepare(`
      SELECT * FROM conversation_messages 
      WHERE conversation_id = ? 
      ORDER BY created_at ASC
    `).all(req.params.id);
    res.json({ success: true, data: messages });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 保存消息
router.post('/:id/messages', (req, res) => {
  try {
    const { role, content } = req.body;
    const stmt = db.prepare(`
      INSERT INTO conversation_messages (conversation_id, role, content) 
      VALUES (?, ?, ?)
    `);
    const result = stmt.run(req.params.id, role, content);
    
    // 更新对话的更新时间
    db.prepare('UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?').run(req.params.id);
    
    const message = db.prepare('SELECT * FROM conversation_messages WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: message });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 清空对话消息
router.delete('/:id/messages', (req, res) => {
  try {
    db.prepare('DELETE FROM conversation_messages WHERE conversation_id = ?').run(req.params.id);
    res.json({ success: true, message: '清空成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除单条消息
router.delete('/:conversationId/messages/:messageId', (req, res) => {
  try {
    const { conversationId, messageId } = req.params;
    db.prepare('DELETE FROM conversation_messages WHERE id = ? AND conversation_id = ?').run(messageId, conversationId);
    
    // 更新对话的更新时间
    db.prepare('UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?').run(conversationId);
    
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新单条消息
router.put('/:conversationId/messages/:messageId', (req, res) => {
  try {
    const { conversationId, messageId } = req.params;
    const { content } = req.body;
    
    db.prepare('UPDATE conversation_messages SET content = ? WHERE id = ? AND conversation_id = ?')
      .run(content, messageId, conversationId);
    
    // 更新对话的更新时间
    db.prepare('UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?').run(conversationId);
    
    const message = db.prepare('SELECT * FROM conversation_messages WHERE id = ?').get(messageId);
    res.json({ success: true, data: message });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

