const express = require('express');
const router = express.Router();
const bookStorage = require('../services/bookStorage');

// 导入文件章节（支持 txt / docx）
router.post('/import-file', async (req, res) => {
  try {
    const { bookId, file } = req.body;
    if (!bookId) {
      return res.status(400).json({ success: false, message: '缺少书籍ID' });
    }
    if (!file || !file.name || !file.data_base64) {
      return res.status(400).json({ success: false, message: '请选择要导入的文件' });
    }

    const MAX_SIZE = 20 * 1024 * 1024;
    const buf = Buffer.from(file.data_base64, 'base64');
    if (buf.length > MAX_SIZE) {
      return res.status(400).json({ success: false, message: '文件大小不能超过 20MB' });
    }

    const ext = file.name.split('.').pop().toLowerCase();
    let textContent = '';

    if (ext === 'txt') {
      textContent = buf.toString('utf-8');
    } else if (ext === 'docx') {
      try {
        const mammoth = require('mammoth');
        const result = await mammoth.extractRawText({ buffer: buf });
        textContent = result.value;
      } catch (e) {
        // mammoth 未安装时尝试用 JSZip 手动解析
        try {
          const JSZip = require('jszip');
          const zip = await JSZip.loadAsync(buf);
          const docXml = await zip.file('word/document.xml').async('string');
          textContent = docXml.replace(/<w:p[^>]*>[\s\S]*?<\/w:p>/g, (match) => {
            return match.replace(/<[^>]+>/g, '').replace(/\s+/g, ' ').trim() + '\n';
          }).trim();
        } catch (e2) {
          return res.status(400).json({ success: false, message: '解析 docx 失败，请安装 mammoth 或 jszip 依赖' });
        }
      }
    } else {
      return res.status(400).json({ success: false, message: '不支持的文件类型，仅支持 txt 和 docx' });
    }

    // 按 "第N章" 模式拆分章节
    const chapterPattern = /第[一二三四五六七八九十百千零\d]+[章节回卷集部篇][^\n]*/g;
    const matches = [...textContent.matchAll(chapterPattern)];

    let chapters = [];
    if (matches.length > 0) {
      for (let i = 0; i < matches.length; i++) {
        const startIdx = matches[i].index;
        const endIdx = i < matches.length - 1 ? matches[i + 1].index : textContent.length;
        const title = matches[i][0].trim();
        const content = textContent.slice(startIdx, endIdx).trim();
        chapters.push({ title, content });
      }
    } else {
      // 未匹配到章节标题，将全文作为单个章节
      chapters.push({
        title: file.name.replace(/\.[^.]+$/, ''),
        content: textContent
      });
    }

    if (chapters.length === 0 || chapters.every(c => !c.content)) {
      return res.status(400).json({ success: false, message: '文件内容为空或无法识别章节格式' });
    }

    const bookMeta = bookStorage.loadBookMetaById(Number(bookId));
    if (!bookMeta) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }

    const inserted = bookStorage.importChapters(Number(bookId), chapters);
    res.json({
      success: true,
      data: { insertedCount: inserted.length, chapters: inserted }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取书本的所有章节
router.get('/book/:bookId', (req, res) => {
  try {
    const chapters = bookStorage.listChapters(Number(req.params.bookId));
    if (!chapters) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }
    res.json({ success: true, data: chapters });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 获取单个章节
router.post('/import-book', (req, res) => {
  try {
    const { bookId, chapters } = req.body;

    if (!bookId) {
      return res.status(400).json({ success: false, message: '缺少书籍ID' });
    }

    if (!Array.isArray(chapters) || chapters.length === 0) {
      return res.status(400).json({ success: false, message: '没有可导入的章节内容' });
    }

    const bookMeta = bookStorage.loadBookMetaById(Number(bookId));
    if (!bookMeta) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }
    const inserted = bookStorage.importChapters(Number(bookId), chapters);
    res.json({
      success: true,
      data: {
        insertedCount: inserted.length,
        chapters: inserted
      }
    });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const loaded = bookStorage.getChapterMeta(Number(req.params.id));
    if (!loaded) {
      return res.status(404).json({ success: false, message: '章节不存在' });
    }
    res.json({ success: true, data: loaded.chapter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 创建章节
router.post('/', (req, res) => {
  try {
    const chapter = bookStorage.createChapter(req.body || {});
    if (!chapter) {
      return res.status(404).json({ success: false, message: '书籍不存在' });
    }
    res.json({ success: true, data: chapter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 更新章节
router.put('/:id', (req, res) => {
  try {
    const chapter = bookStorage.updateChapter(Number(req.params.id), req.body || {});
    if (!chapter) {
      return res.status(404).json({ success: false, message: '章节不存在' });
    }
    res.json({ success: true, data: chapter });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

// 删除章节
router.delete('/:id', (req, res) => {
  try {
    const success = bookStorage.deleteChapter(Number(req.params.id));
    if (!success) {
      return res.status(404).json({ success: false, message: '章节不存在' });
    }
    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;

