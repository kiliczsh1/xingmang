const express = require('express');
const { TextDecoder } = require('util');

const router = express.Router();
const db = require('../database/db');
const {
  parseJsonSafely,
  buildAuthHeaders,
  buildApiUrl,
  buildRequestBody,
  buildRequestConfig,
  readResponsePayload,
  requestJson,
  extractResponseContent,
  getErrorMessage,
  recordUsage,
  getModelConfig
} = require('./apiHelpers');

const FANQIE_RANK_URL = 'https://fanqienovel.com/rank';

function createRequestError(message, details = {}) {
  const error = new Error(message);

  if (details.response) {
    error.response = details.response;
  }

  if (details.request) {
    error.request = details.request;
  }

  if (details.config) {
    error.config = details.config;
  }

  if (details.cause) {
    error.cause = details.cause;
  }

  return error;
}

async function requestStream(url, body, headers) {
  const config = buildRequestConfig(url, body, headers);
  let response;

  try {
    response = await fetch(url, config);
  } catch (error) {
    throw createRequestError(error.message, {
      request: true,
      config,
      cause: error
    });
  }

  if (!response.ok) {
    const data = await readResponsePayload(response);
    throw createRequestError(`HTTP ${response.status}`, {
      response: {
        status: response.status,
        statusText: response.statusText,
        data
      },
      config
    });
  }

  return response;
}

function emitSsePayload(res, payload) {
  res.write(`data: ${JSON.stringify(payload)}\n\n`);
}

function emitSseDone(res) {
  res.write('data: [DONE]\n\n');
}

function processStreamLines(lines, res) {
  for (const rawLine of lines) {
    const line = rawLine.trim();

    if (!line || !line.startsWith('data:')) {
      continue;
    }

    const message = rawLine.replace(/^data:\s*/, '');

    if (message === '[DONE]') {
      return 'done';
    }

    try {
      const parsed = JSON.parse(message);

      if (parsed.error) {
        const errorMessage = parsed.error.message || parsed.error.code || JSON.stringify(parsed.error);
        emitSsePayload(res, { error: errorMessage });
        return 'error';
      }

      if (parsed.choices?.[0]?.finish_reason === 'error' || parsed.choices?.[0]?.error) {
        emitSsePayload(res, {
          error: parsed.choices[0].error || 'AI returned an error response'
        });
        return 'error';
      }

      const content = parsed.choices?.[0]?.delta?.content
        || parsed.delta?.content
        || parsed.output?.text
        || '';

      if (content) {
        emitSsePayload(res, { content });
      }
    } catch (error) {
    }
  }

  return 'continue';
}

async function forwardStreamToSse(response, res) {
  if (!response.body) {
    throw new Error('AI stream body is empty');
  }

  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';

  while (true) {
    const { value, done } = await reader.read();

    if (done) {
      break;
    }

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop() || '';

    const status = processStreamLines(lines, res);
    if (status !== 'continue') {
      return status;
    }
  }

  buffer += decoder.decode();

  if (buffer.trim()) {
    return processStreamLines([buffer], res);
  }

  return 'continue';
}

function stripMarkdownCodeFence(text) {
  return text.replace(/```json\s*/gi, '').replace(/```\s*/g, '').trim();
}

function decodeHtmlEntities(text) {
  return String(text || '')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&nbsp;/g, ' ');
}

function stripHtmlTags(text) {
  return decodeHtmlEntities(String(text || '').replace(/<[^>]+>/g, ' ')).replace(/\s+/g, ' ').trim();
}

function normalizeFanqieUrl(url) {
  const value = String(url || '').trim();
  if (!value) return '';
  if (value.startsWith('//')) return `https:${value}`;
  if (value.startsWith('/')) return `https://fanqienovel.com${value}`;
  return value;
}

async function fetchPageHtml(url) {
  const response = await fetch(url, {
    headers: {
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36'
    }
  });
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
  return response.text();
}

function parseRankCategoriesFromHtml(html) {
  const groupRegex = /<div tabindex="0" aria-expanded="false" class="arco-menu-inline-header"><span><span>([^<]+)<\/span><\/span>[\s\S]*?<div class="arco-menu-inline-content"[^>]*>([\s\S]*?)<\/div><\/div>/g;
  const itemRegex = /<a level="2" href="([^"]+)">([^<]+)<\/a>/g;
  const groups = [];
  let groupMatch;

  while ((groupMatch = groupRegex.exec(html)) !== null) {
    const label = stripHtmlTags(groupMatch[1]);
    const content = groupMatch[2];
    const categories = [];
    let itemMatch;

    while ((itemMatch = itemRegex.exec(content)) !== null) {
      const path = itemMatch[1];
      const id = path.split('_').pop() || path;
      categories.push({
        id,
        label: stripHtmlTags(itemMatch[2]),
        path,
        fullUrl: normalizeFanqieUrl(path)
      });
    }

    if (label && categories.length > 0) {
      groups.push({
        key: `group_${groups.length + 1}`,
        label,
        categories
      });
    }
  }

  return groups;
}

function parseRankBooksFromHtml(html) {
  const itemRegex = /<div class="rank-book-item">([\s\S]*?)<\/div><\/div>(?=<div class="rank-book-item">|<\/div><\/div><\/div><\/div><\/div>)/g;
  const rankRegex = /<div class="book-item-index"><h1>(\d+)<\/h1>/;
  const coverRegex = /<img class="book-cover-img\s*" src="([^"]+)" alt="([^"]*)"\/>/;
  const titleRegex = /<div class="title"><a href="([^"]+)"[^>]*>([^<]+)<\/a><\/div>/;
  const authorRegex = /<div class="author">[\s\S]*?<span[^>]*>([^<]+)<\/span>/;
  const books = [];
  let match;

  while ((match = itemRegex.exec(html)) !== null) {
    const block = match[1];
    const rank = Number(block.match(rankRegex)?.[1] || 0);
    const coverMatch = block.match(coverRegex);
    const titleMatch = block.match(titleRegex);
    const authorMatch = block.match(authorRegex);
    const coverUrl = normalizeFanqieUrl(coverMatch?.[1] || '');

    if (!coverUrl) {
      continue;
    }

    books.push({
      rank,
      title: stripHtmlTags(titleMatch?.[2] || coverMatch?.[2] || ''),
      author: stripHtmlTags(authorMatch?.[1] || ''),
      coverUrl,
      bookPath: titleMatch?.[1] || ''
    });
  }

  return books;
}

function extractBookDetailMetadata(html) {
  const imageMatch = html.match(/"image":\s*\[\s*"([^"]+)"\s*\]/);
  const titleMatch = html.match(/<div class="info-name"><h1>([^<]+)<\/h1><\/div>/);
  const authorMatch = html.match(/<span class="author-name-text">([^<]+)<\/span>/);

  return {
    coverUrl: normalizeFanqieUrl(imageMatch?.[1] || ''),
    title: stripHtmlTags(titleMatch?.[1] || ''),
    author: stripHtmlTags(authorMatch?.[1] || '')
  };
}

function isPlaceholderCover(url) {
  return /\/novel-static\//i.test(String(url || ''));
}

async function enrichRankBooks(books) {
  return Promise.all(
    books.map(async (book) => {
      if (!book.bookPath) {
        return book;
      }

      try {
        const detailHtml = await fetchPageHtml(normalizeFanqieUrl(book.bookPath));
        const detail = extractBookDetailMetadata(detailHtml);
        return {
          ...book,
          coverUrl: detail.coverUrl || book.coverUrl,
          title: detail.title || book.title,
          author: detail.author || book.author
        };
      } catch (error) {
        return book;
      }
    })
  );
}

function walkJson(value, visitor) {
  if (Array.isArray(value)) {
    for (const item of value) {
      const result = walkJson(item, visitor);
      if (result) return result;
    }
    return '';
  }

  if (!value || typeof value !== 'object') {
    return '';
  }

  for (const [key, child] of Object.entries(value)) {
    const result = visitor(key, child);
    if (result) return result;
    const nested = walkJson(child, visitor);
    if (nested) return nested;
  }

  return '';
}

function extractImageResult(payload) {
  if (!payload || typeof payload !== 'object') {
    return '';
  }

  const directUrl = walkJson(payload, (key, value) => {
    if (
      ['url', 'image_url', 'fileUri', 'file_uri'].includes(key)
      && typeof value === 'string'
      && value.trim()
    ) {
      return value.trim();
    }
    return '';
  });

  if (directUrl) {
    return directUrl;
  }

  const base64Image = walkJson(payload, (key, value) => {
    if (
      ['b64_json', 'image_base64', 'base64', 'b64'].includes(key)
      && typeof value === 'string'
      && value.trim()
    ) {
      return value.trim();
    }
    return '';
  });

  if (base64Image) {
    return `data:image/png;base64,${base64Image}`;
  }

  const textContent = walkJson(payload, (key, value) => {
    if (typeof value !== 'string' || !value.trim()) {
      return '';
    }
    if (!['content', 'text', 'output_text'].includes(key)) {
      return '';
    }
    return value;
  });

  if (!textContent) {
    return '';
  }

  const imageUrlMatch = textContent.match(/https?:\/\/[^\s)"']+/);
  if (imageUrlMatch) {
    return imageUrlMatch[0];
  }

  try {
    const parsedContent = JSON.parse(stripMarkdownCodeFence(textContent));
    return extractImageResult(parsedContent);
  } catch (error) {
    return '';
  }
}

function summarizeImagePayload(payload) {
  const summary = {
    topLevelType: Array.isArray(payload) ? 'array' : typeof payload,
    topLevelKeys: payload && typeof payload === 'object' && !Array.isArray(payload)
      ? Object.keys(payload).slice(0, 20)
      : [],
    foundKeys: [],
    foundTextSamples: []
  };

  const interestingKeys = new Set([
    'url',
    'image_url',
    'fileUri',
    'file_uri',
    'b64_json',
    'image_base64',
    'base64',
    'b64',
    'content',
    'text',
    'output_text'
  ]);

  walkJson(payload, (key, value) => {
    if (!interestingKeys.has(key)) {
      return '';
    }
    if (!summary.foundKeys.includes(key)) {
      summary.foundKeys.push(key);
    }
    if (
      typeof value === 'string'
      && value.trim()
      && ['content', 'text', 'output_text'].includes(key)
      && summary.foundTextSamples.length < 3
    ) {
      summary.foundTextSamples.push(value.trim().slice(0, 200));
    }
    return '';
  });

  return summary;
}

function buildImageRequestCandidates(apiUrl, requestBody, prompt) {
  const baseUrl = String(apiUrl || '').trim();
  const normalizedPrompt = String(prompt || '').trim();
  const defaultSize = requestBody.size || '600x800';
  const referenceImages = Array.isArray(requestBody.reference_images) ? requestBody.reference_images : [];
  const candidates = [];

  const addCandidate = (url, body, label) => {
    if (!url) return;
    if (candidates.some(candidate => candidate.url === url && JSON.stringify(candidate.body) === JSON.stringify(body))) {
      return;
    }
    candidates.push({ url, body, label });
  };

  addCandidate(baseUrl, requestBody, 'original');

  if (/\/chat\/completions\/?$/i.test(baseUrl)) {
    addCandidate(
      baseUrl.replace(/\/chat\/completions\/?$/i, '/images/generations'),
      {
        model: requestBody.model,
        prompt: normalizedPrompt,
        n: 1,
        size: defaultSize,
        image: referenceImages[0]?.data_base64
      },
      'images-from-chat-endpoint'
    );

    addCandidate(
      baseUrl,
      {
        model: requestBody.model,
        messages: [{
          role: 'user',
          content: [
            { type: 'text', text: normalizedPrompt },
            ...referenceImages.map((image) => ({
              type: 'image_url',
              image_url: {
                url: `data:${image.mime_type || 'image/png'};base64,${image.data_base64}`
              }
            }))
          ]
        }],
        modalities: ['text', 'image']
      },
      'chat-modalities'
    );
  }

  if (/\/images\/generations\/?$/i.test(baseUrl)) {
    addCandidate(
      baseUrl,
      {
        model: requestBody.model,
        prompt: normalizedPrompt,
        n: 1,
        size: defaultSize,
        image: referenceImages[0]?.data_base64
      },
      'images-prompt'
    );

    addCandidate(
      baseUrl.replace(/\/images\/generations\/?$/i, '/chat/completions'),
      {
        model: requestBody.model,
        messages: referenceImages.length > 0
          ? [{
              role: 'user',
              content: [
                { type: 'text', text: normalizedPrompt },
                ...referenceImages.map((image) => ({
                  type: 'image_url',
                  image_url: {
                    url: `data:${image.mime_type || 'image/png'};base64,${image.data_base64}`
                  }
                }))
              ]
            }]
          : requestBody.messages,
        n: 1,
        size: defaultSize,
        quality: requestBody.quality,
        style: requestBody.style
      },
      'chat-from-images-endpoint'
    );
  }

  return candidates;
}

async function requestImageWithFallback(config, requestBody, prompt) {
  const candidates = buildImageRequestCandidates(config.api_url, requestBody, prompt);
  let lastError = null;
  const imageTimeoutMs = 180000;

  for (const candidate of candidates) {
    try {
      const response = await requestJson(
        candidate.url,
        candidate.body,
        buildAuthHeaders(config.api_key),
        imageTimeoutMs
      );

      console.log('==================== Image Generation Response ====================');
      console.log('Provider URL:', candidate.url);
      console.log('Model:', config.model);
      console.log('Attempt:', candidate.label);
      console.log('Payload summary:', JSON.stringify(summarizeImagePayload(response.data), null, 2));
      console.log('==================================================================');

      return response;
    } catch (error) {
      lastError = error;
      console.error('==================== Image Generation Attempt Failed ====================');
      console.error('Attempt:', candidate.label);
      console.error('Request URL:', candidate.url);
      console.error('Request Body:', JSON.stringify(candidate.body, null, 2));
      console.error('Response status:', error.response?.status);
      console.error('Response status text:', error.response?.statusText);
      console.error('Response data:', JSON.stringify(error.response?.data ?? null, null, 2));
      console.error('========================================================================');
    }
  }

  throw lastError || new Error('图片生成请求失败');
}

router.post('/chat', async (req, res) => {
  try {
    const {
      messages = [],
      configId,
      systemPrompts = [],
      relatedContent = []
    } = req.body;

    const config = getModelConfig(configId);
    if (!config) {
      return res.status(400).json({ success: false, message: '未找到 API 配置' });
    }

    const chatMessages = [];
    const systemContents = [];

    console.log('Received systemPrompts:', systemPrompts);

    if (systemPrompts.length > 0) {
      systemContents.push(systemPrompts.join('\n\n'));
    }

    if (relatedContent.length > 0) {
      const contextContent = relatedContent
        .map((item) => `[${item.type}] ${item.title}\n${item.content}`)
        .join('\n\n');
      systemContents.push(`参考内容：\n${contextContent}`);
    }

    if (systemContents.length > 0) {
      chatMessages.push({
        role: 'system',
        content: systemContents.join('\n\n---\n\n')
      });
    }

    chatMessages.push(
      ...messages.map((message) => ({
        role: message.role,
        content: message.content
      }))
    );

    const apiFormat = config.api_format || 'openai';
    const useFullUrl = config.use_full_url === 1;
    const apiUrl = buildApiUrl(config.api_url, apiFormat, useFullUrl);
    const requestBody = buildRequestBody(config, chatMessages, { stream: true });

    console.log('==================== AI Request ====================');
    console.log('URL:', apiUrl);
    console.log('API Format:', apiFormat);
    console.log('Use Full URL:', useFullUrl);
    console.log('Model:', config.model);
    console.log('Temperature:', config.temperature);
    console.log('Max Tokens:', config.max_tokens);
    console.log('Messages Count:', chatMessages.length);
    console.log('Messages:', JSON.stringify(chatMessages, null, 2));
    console.log('===================================================');

    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    try {
      const response = await requestStream(
        apiUrl,
        requestBody,
        buildAuthHeaders(config.api_key, apiFormat)
      );

      const streamStatus = await forwardStreamToSse(response, res);

      if (streamStatus !== 'error') {
        recordUsage(config.id, config.name, config.provider_name);
      }

      emitSseDone(res);
      res.end();
    } catch (error) {
      console.error('==================== AI Request Failed ====================');
      console.error('Error name:', error.name);
      console.error('Error message:', error.message);
      console.error('Request URL:', apiUrl);
      console.error('API Format:', apiFormat);
      console.error('Request Body:', JSON.stringify(requestBody, null, 2));
      console.error('Response status:', error.response?.status);
      console.error('Response status text:', error.response?.statusText);
      console.error('===========================================================');

      emitSsePayload(res, {
        error: getErrorMessage(error)
      });
      emitSseDone(res);
      res.end();
    }
  } catch (error) {
    console.error('==================== Unexpected Error ====================');
    console.error('Error message:', error.message);
    console.error('Error stack:', error.stack);
    console.error('=========================================================');

    if (!res.headersSent) {
      return res.status(500).json({
        success: false,
        message: error.message
      });
    }

    emitSsePayload(res, { error: error.message });
    emitSseDone(res);
    res.end();
  }
});

router.post('/generate-description', async (req, res) => {
  try {
    const { title, promptId } = req.body;

    let promptContent = '你是一个专业的网络小说编辑，请根据书名生成一个吸引人的简介，字数控制在 100-200 字之间。';
    if (promptId) {
      const prompt = db.prepare('SELECT * FROM prompts WHERE id = ?').get(promptId);
      if (prompt) {
        promptContent = prompt.content;
      }
    }

    const config = getModelConfig();
    if (!config) {
      return res.status(400).json({ success: false, message: '未找到 API 配置' });
    }

    const apiFormat = config.api_format || 'openai';
    const useFullUrl = config.use_full_url === 1;
    const apiUrl = buildApiUrl(config.api_url, apiFormat, useFullUrl);
    const messages = [
      { role: 'system', content: promptContent },
      { role: 'user', content: `书名：${title}` }
    ];
    const requestBody = buildRequestBody(config, messages, { stream: false });

    const response = await requestJson(
      apiUrl,
      requestBody,
      buildAuthHeaders(config.api_key, apiFormat)
    );

    recordUsage(config.id, config.name, config.provider_name);

    res.json({
      success: true,
      data: extractResponseContent(response.data)
    });
  } catch (error) {
    console.error('Generate description failed:', error.message);
    res.status(500).json({
      success: false,
      message: getErrorMessage(error)
    });
  }
});

router.post('/recognize-characters', async (req, res) => {
  try {
    const { text, configId, customPrompt } = req.body;

    if (!text || !text.trim()) {
      return res.status(400).json({
        success: false,
        message: '请输入待识别的文本'
      });
    }

    const config = getModelConfig(configId);
    if (!config) {
      return res.status(400).json({
        success: false,
        message: '未找到 API 配置，请先配置 AI 模型'
      });
    }

    const defaultPrompt = `你是一个专业的角色识别助手。请从给定的文本中识别出所有角色，并为每个角色提取以下信息：
1. 角色名称
2. 性别（male / female / unknown / none）
3. 角色性格
4. 角色信息（仅填写剧情中明确给出的信息）

请只返回 JSON，格式如下：
{
  "characters": [
    {
      "name": "角色名称",
      "gender": "male|female|unknown|none",
      "personality": "角色性格描述",
      "info": "角色信息"
    }
  ]
}

注意：
- 只返回 JSON，不要包含额外说明
- 如果无法确定性别，请使用 "unknown"
- 性格和信息保持简洁
- 如果文本中没有明确角色，请返回空数组`;

    const systemPrompt = typeof customPrompt === 'string' && customPrompt.trim()
      ? customPrompt.trim()
      : defaultPrompt;

    const response = await requestJson(
      config.api_url,
      {
        model: config.model,
        messages: [
          { role: 'system', content: systemPrompt },
          { role: 'user', content: text }
        ],
        temperature: 0.7,
        max_tokens: Math.min(Math.max(Number(config.max_tokens) || 2000, 1), 8192)
      },
      buildAuthHeaders(config.api_key)
    );

    const content = extractResponseContent(response.data);
    const parsed = parseJsonSafely(stripMarkdownCodeFence(content));

    if (!parsed || typeof parsed !== 'object' || !Array.isArray(parsed.characters)) {
      return res.status(500).json({
        success: false,
        message: 'AI 返回的格式不正确，请重试'
      });
    }

    recordUsage(config.id, config.name, config.provider_name);

    res.json({
      success: true,
      data: parsed.characters
    });
  } catch (error) {
    console.error('Recognize characters failed:', error.message);
    res.status(500).json({
      success: false,
      message: getErrorMessage(error)
    });
  }
});

router.post('/generate-image', async (req, res) => {
  try {
    const { messages, configId, size, quality, style, reference_images } = req.body;

    let prompt = '';
    if (messages && Array.isArray(messages)) {
      prompt = messages.map(m => m.content).join('\n');
    }

    if (!prompt || !prompt.trim()) {
      return res.status(400).json({
        success: false,
        message: '请输入图片描述提示词'
      });
    }

    const config = getModelConfig(configId);
    if (!config) {
      return res.status(400).json({
        success: false,
        message: '未找到 API 配置'
      });
    }

    const requestBody = {
      model: config.model,
      messages: messages,
      n: 1,
      size: size || '600x800',
      quality: quality || 'standard',
      style: style || 'vivid',
      reference_images: Array.isArray(reference_images) ? reference_images : []
    };

    const response = await requestImageWithFallback(config, requestBody, prompt);

    recordUsage(config.id, config.name, config.provider_name);

    const imageUrl = extractImageResult(response.data);

    if (!imageUrl) {
      return res.status(502).json({
        success: false,
        message: '图片接口返回成功，但未找到可显示的图片地址或 base64 数据'
      });
    }

    res.json({
      success: true,
      data: { url: imageUrl }
    });
  } catch (error) {
    console.error('Generate image failed:', error.message);
    console.error('Generate image response status:', error.response?.status);
    console.error('Generate image response data:', JSON.stringify(error.response?.data ?? null, null, 2));
    res.status(500).json({
      success: false,
      message: getErrorMessage(error)
    });
  }
});

router.get('/rank-reference/categories', async (req, res) => {
  try {
    const html = await fetchPageHtml(FANQIE_RANK_URL);
    const groups = parseRankCategoriesFromHtml(html);
    res.json({ success: true, data: groups });
  } catch (error) {
    res.status(500).json({ success: false, message: getErrorMessage(error, '获取排行榜分类失败') });
  }
});

router.get('/rank-reference/books', async (req, res) => {
  try {
    const requestedPath = String(req.query.path || '').trim();
    if (!/^\/rank\/[0-9_]+$/.test(requestedPath)) {
      return res.status(400).json({ success: false, message: '无效的排行榜路径' });
    }
    const html = await fetchPageHtml(normalizeFanqieUrl(requestedPath));
    const parsedBooks = parseRankBooksFromHtml(html);
    const needDetailEnrichment = parsedBooks.some((book) => isPlaceholderCover(book.coverUrl));
    const books = needDetailEnrichment
      ? await enrichRankBooks(parsedBooks)
      : parsedBooks;
    res.json({ success: true, data: books });
  } catch (error) {
    res.status(500).json({ success: false, message: getErrorMessage(error, '获取排行榜封面失败') });
  }
});

module.exports = router;
