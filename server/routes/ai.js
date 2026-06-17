const express = require('express');
const { TextDecoder } = require('util');

const router = express.Router();
const db = require('../database/db');
const {
  parseJsonSafely,
  buildAuthHeaders,
  buildRequestConfig,
  readResponsePayload,
  requestJson,
  extractResponseContent,
  getErrorMessage,
  recordUsage,
  getModelConfig
} = require('./apiHelpers');

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

    const maxTokens = Math.min(Math.max(Number(config.max_tokens) || 2000, 1), 8192);

    const requestBody = {
      model: config.model,
      messages: chatMessages,
      temperature: config.temperature,
      max_tokens: maxTokens,
      stream: true
    };

    console.log('==================== AI Request ====================');
    console.log('URL:', config.api_url);
    console.log('Model:', config.model);
    console.log('Temperature:', config.temperature);
    console.log('Max Tokens (original):', config.max_tokens);
    console.log('Max Tokens (clamped):', maxTokens);
    console.log('Messages Count:', chatMessages.length);
    console.log('Messages:', JSON.stringify(chatMessages, null, 2));
    console.log('===================================================');

    res.setHeader('Content-Type', 'text/event-stream');
    res.setHeader('Cache-Control', 'no-cache');
    res.setHeader('Connection', 'keep-alive');

    try {
      const response = await requestStream(
        config.api_url,
        requestBody,
        buildAuthHeaders(config.api_key)
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
      console.error('Request URL:', config.api_url);
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

    const response = await requestJson(
      config.api_url,
      {
        model: config.model,
        messages: [
          { role: 'system', content: promptContent },
          { role: 'user', content: `书名：${title}` }
        ],
        temperature: config.temperature,
        max_tokens: Math.min(Math.max(Number(config.max_tokens) || 2000, 1), 8192)
      },
      buildAuthHeaders(config.api_key)
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

module.exports = router;
