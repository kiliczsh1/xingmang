const express = require('express');
const fs = require('fs');
const path = require('path');
const { randomUUID } = require('crypto');
const db = require('../database/db');

const router = express.Router();

const MAX_PDF_SIZE = 20 * 1024 * 1024;
const PDF_URL_PREFIX = '/api/uploads/experience-shares/';
let pdfJsLoaderPromise = null;

const loadPdfJs = async () => {
  if (!pdfJsLoaderPromise) {
    pdfJsLoaderPromise = import('pdfjs-dist/legacy/build/pdf.mjs');
  }

  return pdfJsLoaderPromise;
};

const getUploadsRoot = () => {
  if (process.pkg) {
    return path.join(path.dirname(process.execPath), 'uploads');
  }

  return path.join(__dirname, '..', 'uploads');
};

const getExperienceShareUploadDir = () => {
  const dir = path.join(getUploadsRoot(), 'experience-shares');
  fs.mkdirSync(dir, { recursive: true });
  return dir;
};

const sanitizeFileName = (fileName = 'experience.pdf') => {
  const ext = path.extname(fileName) || '.pdf';
  const baseName = path.basename(fileName, ext).replace(/[^\w\u4e00-\u9fa5.-]+/g, '_').slice(0, 80) || 'experience';
  return `${baseName}${ext.toLowerCase()}`;
};

const getBaseName = (fileName = '') => path.basename(fileName, path.extname(fileName)).trim() || '未命名经验';

const decodeBase64Pdf = (value = '') => {
  const cleanBase64 = String(value).replace(/^data:application\/pdf;base64,/i, '');
  return Buffer.from(cleanBase64, 'base64');
};

const stripRichText = (value = '') =>
  String(value)
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/gi, ' ')
    .replace(/&amp;/gi, '&')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/gi, "'")
    .replace(/\s+/g, ' ')
    .trim();

const hasMeaningfulContent = (value = '') => stripRichText(value).length > 0;
const canUsePdfAsContent = (createType, hasPdfAttachment) => createType === 'pdf_import' && Boolean(hasPdfAttachment);

const normalizeImportedText = (text = '') =>
  String(text)
    .replace(/\r\n/g, '\n')
    .replace(/[ \t]+\n/g, '\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim();

const buildSummary = (value = '', maxLength = 140) => {
  const text = stripRichText(value);
  if (!text) return '';
  return text.length > maxLength ? `${text.slice(0, maxLength).trim()}...` : text;
};

const normalizeShare = row => ({
  ...row,
  pdf_file_size: row.pdf_file_size ? Number(row.pdf_file_size) : 0,
  author_id: row.author_id ?? null
});

const validatePdfPayload = pdfFile => {
  if (!pdfFile || typeof pdfFile !== 'object') {
    throw new Error('请先选择 PDF 文件');
  }

  const fileName = String(pdfFile.name || '').trim();
  if (!fileName || !/\.pdf$/i.test(fileName)) {
    throw new Error('仅支持 PDF 格式文件');
  }

  const fileSize = Number(pdfFile.size || 0);
  if (!Number.isFinite(fileSize) || fileSize <= 0) {
    throw new Error('PDF 文件大小无效');
  }

  if (fileSize > MAX_PDF_SIZE) {
    throw new Error(`PDF 文件大小不能超过 ${Math.floor(MAX_PDF_SIZE / 1024 / 1024)}MB`);
  }

  if (!pdfFile.data_base64 || typeof pdfFile.data_base64 !== 'string') {
    throw new Error('PDF 文件内容不能为空');
  }
};

const getAbsolutePathFromUrl = fileUrl => {
  if (!fileUrl || typeof fileUrl !== 'string' || !fileUrl.startsWith(PDF_URL_PREFIX)) {
    return null;
  }

  const relativeName = fileUrl.slice(PDF_URL_PREFIX.length);
  if (!relativeName) return null;
  return path.join(getExperienceShareUploadDir(), path.basename(relativeName));
};

const removeUploadedPdf = fileUrl => {
  const targetPath = getAbsolutePathFromUrl(fileUrl);
  if (targetPath && fs.existsSync(targetPath)) {
    fs.unlinkSync(targetPath);
  }
};

const savePdfAttachment = pdfFile => {
  validatePdfPayload(pdfFile);

  const buffer = decodeBase64Pdf(pdfFile.data_base64);
  const safeOriginalName = sanitizeFileName(pdfFile.name);
  const storedFileName = `${Date.now()}-${randomUUID()}-${safeOriginalName}`;
  const uploadDir = getExperienceShareUploadDir();
  const absolutePath = path.join(uploadDir, storedFileName);

  fs.writeFileSync(absolutePath, buffer);

  return {
    pdf_file_url: `${PDF_URL_PREFIX}${storedFileName}`,
    pdf_file_name: pdfFile.name,
    pdf_file_size: Number(pdfFile.size || buffer.length)
  };
};

const extractTextFromPageContent = textContent => {
  if (!textContent || !Array.isArray(textContent.items)) {
    return '';
  }

  return textContent.items
    .map(item => {
      if (!item || typeof item.str !== 'string') {
        return '';
      }

      return item.hasEOL ? `${item.str}\n` : `${item.str} `;
    })
    .join('');
};

const extractPdfDraft = async pdfFile => {
  validatePdfPayload(pdfFile);

  const pdfjs = await loadPdfJs();
  const loadingTask = pdfjs.getDocument({
    data: new Uint8Array(decodeBase64Pdf(pdfFile.data_base64))
  });

  try {
    const pdfDocument = await loadingTask.promise;
    let infoResult = null;
    const pageTexts = [];

    try {
      infoResult = await pdfDocument.getMetadata().catch(() => null);

      for (let pageIndex = 1; pageIndex <= pdfDocument.numPages; pageIndex += 1) {
        const page = await pdfDocument.getPage(pageIndex);

        try {
          const textContent = await page.getTextContent();
          const pageText = normalizeImportedText(extractTextFromPageContent(textContent));

          if (pageText) {
            pageTexts.push(pageText);
          }
        } finally {
          if (typeof page.cleanup === 'function') {
            page.cleanup();
          }
        }
      }
    } finally {
      await pdfDocument.destroy();
    }

    const extractedText = normalizeImportedText(pageTexts.join('\n\n'));
    const fileNameTitle = getBaseName(pdfFile.name);
    const title = fileNameTitle;
    const totalPages = Number(infoResult?.info?.Pages || 0) || pageTexts.length;
    const summary = buildSummary(extractedText, 180) || `来自 ${pdfFile.name} 的 PDF 内容摘要`;
    const parseStatus = extractedText ? 'success' : 'empty';
    const parseResult = extractedText
      ? `已解析 ${totalPages || '多'} 页 PDF，自动生成标题与简介，正文保留原始 PDF 预览`
      : 'PDF 已导入，未强制解析正文内容，将直接保留原始 PDF 预览';

    return {
      title,
      summary,
      content: '',
      create_type: 'pdf_import',
      pdf_parse_status: parseStatus,
      pdf_parse_result: parseResult,
      source_file_name: pdfFile.name
    };
  } finally {
    await loadingTask.destroy();
  }
};

router.get('/', (req, res) => {
  try {
    const rows = db
      .prepare('SELECT * FROM experience_shares ORDER BY datetime(created_at) DESC, id DESC')
      .all()
      .map(normalizeShare);

    res.json({ success: true, data: rows });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.get('/:id', (req, res) => {
  try {
    const row = db.prepare('SELECT * FROM experience_shares WHERE id = ?').get(req.params.id);

    if (!row) {
      return res.status(404).json({ success: false, message: '经验卡片不存在' });
    }

    res.json({ success: true, data: normalizeShare(row) });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.post('/import-pdf', async (req, res) => {
  try {
    const draft = await extractPdfDraft(req.body?.pdf_file);
    res.json({ success: true, data: draft });
  } catch (error) {
    res.status(400).json({ success: false, message: `导入失败，请重试或改用手动创建。${error.message ? ` ${error.message}` : ''}`.trim() });
  }
});

router.post('/', (req, res) => {
  try {
    const {
      title,
      summary,
      content,
      cover_url,
      create_type,
      author_id,
      author_name,
      status,
      pdf_parse_status,
      pdf_parse_result,
      source_file_name,
      pdf_file
    } = req.body || {};

    if (!String(title || '').trim()) {
      return res.status(400).json({ success: false, message: '标题不能为空' });
    }

    const nextCreateType = create_type === 'pdf_import' ? 'pdf_import' : 'manual';

    if (!hasMeaningfulContent(content) && !canUsePdfAsContent(nextCreateType, pdf_file)) {
      return res.status(400).json({ success: false, message: '正文不能为空，请补充正文后再发布' });
    }

    let pdfMeta = {
      pdf_file_url: null,
      pdf_file_name: null,
      pdf_file_size: 0
    };

    if (pdf_file) {
      pdfMeta = savePdfAttachment(pdf_file);
    }

    const result = db.prepare(`
      INSERT INTO experience_shares (
        title,
        summary,
        content,
        cover_url,
        pdf_file_url,
        pdf_file_name,
        pdf_file_size,
        create_type,
        author_id,
        author_name,
        status,
        pdf_parse_status,
        pdf_parse_result,
        source_file_name
      ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `).run(
      String(title).trim(),
      String(summary || '').trim() || null,
      String(content || '').trim(),
      String(cover_url || '').trim() || null,
      pdfMeta.pdf_file_url,
      pdfMeta.pdf_file_name,
      pdfMeta.pdf_file_size,
      nextCreateType,
      author_id || null,
      String(author_name || '').trim() || '星芒用户',
      String(status || '').trim() || 'published',
      String(pdf_parse_status || '').trim() || null,
      String(pdf_parse_result || '').trim() || null,
      String(source_file_name || pdfMeta.pdf_file_name || '').trim() || null
    );

    const row = db.prepare('SELECT * FROM experience_shares WHERE id = ?').get(result.lastInsertRowid);
    res.json({ success: true, data: normalizeShare(row) });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.put('/:id', (req, res) => {
  try {
    const existing = db.prepare('SELECT * FROM experience_shares WHERE id = ?').get(req.params.id);

    if (!existing) {
      return res.status(404).json({ success: false, message: '经验卡片不存在' });
    }

    const {
      title,
      summary,
      content,
      cover_url,
      create_type,
      author_id,
      author_name,
      status,
      pdf_parse_status,
      pdf_parse_result,
      source_file_name,
      pdf_file,
      remove_pdf
    } = req.body || {};

    if (!String(title || '').trim()) {
      return res.status(400).json({ success: false, message: '标题不能为空' });
    }

    const nextCreateType = create_type === 'pdf_import' ? 'pdf_import' : 'manual';
    const willHavePdfAttachment = Boolean(pdf_file || (!remove_pdf && existing.pdf_file_url));

    if (!hasMeaningfulContent(content) && !canUsePdfAsContent(nextCreateType, willHavePdfAttachment)) {
      return res.status(400).json({ success: false, message: '正文不能为空，请补充正文后再发布' });
    }

    let pdfMeta = {
      pdf_file_url: existing.pdf_file_url,
      pdf_file_name: existing.pdf_file_name,
      pdf_file_size: existing.pdf_file_size
    };

    if (remove_pdf) {
      removeUploadedPdf(existing.pdf_file_url);
      pdfMeta = {
        pdf_file_url: null,
        pdf_file_name: null,
        pdf_file_size: 0
      };
    }

    if (pdf_file) {
      const nextPdfMeta = savePdfAttachment(pdf_file);
      removeUploadedPdf(existing.pdf_file_url);
      pdfMeta = nextPdfMeta;
    }

    db.prepare(`
      UPDATE experience_shares
      SET
        title = ?,
        summary = ?,
        content = ?,
        cover_url = ?,
        pdf_file_url = ?,
        pdf_file_name = ?,
        pdf_file_size = ?,
        create_type = ?,
        author_id = ?,
        author_name = ?,
        status = ?,
        pdf_parse_status = ?,
        pdf_parse_result = ?,
        source_file_name = ?,
        updated_at = CURRENT_TIMESTAMP
      WHERE id = ?
    `).run(
      String(title).trim(),
      String(summary || '').trim() || null,
      String(content || '').trim(),
      String(cover_url || '').trim() || null,
      pdfMeta.pdf_file_url,
      pdfMeta.pdf_file_name,
      Number(pdfMeta.pdf_file_size || 0),
      nextCreateType,
      author_id || null,
      String(author_name || existing.author_name || '').trim() || '星芒用户',
      String(status || existing.status || '').trim() || 'published',
      String(pdf_parse_status || existing.pdf_parse_status || '').trim() || null,
      String(pdf_parse_result || existing.pdf_parse_result || '').trim() || null,
      String(source_file_name || pdfMeta.pdf_file_name || existing.source_file_name || '').trim() || null,
      req.params.id
    );

    const row = db.prepare('SELECT * FROM experience_shares WHERE id = ?').get(req.params.id);
    res.json({ success: true, data: normalizeShare(row) });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

router.delete('/:id', (req, res) => {
  try {
    const existing = db.prepare('SELECT * FROM experience_shares WHERE id = ?').get(req.params.id);

    if (!existing) {
      return res.status(404).json({ success: false, message: '经验卡片不存在' });
    }

    removeUploadedPdf(existing.pdf_file_url);
    db.prepare('DELETE FROM experience_shares WHERE id = ?').run(req.params.id);

    res.json({ success: true, message: '删除成功' });
  } catch (error) {
    res.status(500).json({ success: false, message: error.message });
  }
});

module.exports = router;
