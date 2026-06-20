const express = require('express');
const router = express.Router();
const bookStorage = require('../services/bookStorage');

// 获取书本的所有分卷
router.get('/book/:bookId', (req, res) => {
  try {
    const volumes = bookStorage.listVolumes(Number(req.params.bookId));
    if (!volumes) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }
    res.json({ success: true, data: volumes });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个分卷
router.get('/:id', (req, res) => {
  try {
    const volume = bookStorage.getVolume(Number(req.params.id));
    if (!volume) {
      return res.status(404).json({ success: false, message: '分卷不存在' });
    }
    res.json({ success: true, data: volume });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建分卷
router.post('/', (req, res) => {
  try {
    const volume = bookStorage.createVolume(req.body || {});
    if (!volume) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }
    res.json({ success: true, data: volume });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新分卷
router.put('/:id', (req, res) => {
  try {
    const volume = bookStorage.updateVolume(Number(req.params.id), req.body || {});
    if (!volume) {
      return res.status(404).json({ success: false, message: '分卷不存在' });
    }
    res.json({ success: true, data: volume });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除分卷
router.delete('/:id', (req, res) => {
  try {
    const success = bookStorage.deleteVolume(Number(req.params.id));
    if (!success) {
      return res.status(404).json({ success: false, message: '分卷不存在' });
    }
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
