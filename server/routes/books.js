const express = require('express');
const router = express.Router();
const bookStorage = require('../services/bookStorage');

// 获取所有书本
router.get('/', (req, res) => {
  try {
    res.json({ success: true, data: bookStorage.loadAllBooks() });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个书本
router.get('/:id', (req, res) => {
  try {
    const meta = bookStorage.loadBookMetaById(Number(req.params.id));
    if (!meta) {
      return res.status(404).json({ success: false, message: '书本不存在' });
    }
    res.json({ success: true, data: { ...meta.book, base_path: undefined } });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建书本
router.post('/', (req, res) => {
  try {
    const created = bookStorage.createBook(req.body || {});
    res.json({ success: true, data: created });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新书本
router.put('/:id', (req, res) => {
  try {
    const updated = bookStorage.updateBook(Number(req.params.id), req.body || {});
    if (!updated) {
      return res.status(404).json({ success: false, message: '书本不存在' });
    }
    res.json({ success: true, data: updated });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除书本
router.delete('/:id', (req, res) => {
  try {
    bookStorage.deleteBook(Number(req.params.id));
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
