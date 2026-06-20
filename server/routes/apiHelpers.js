const db = require('../database/db');

function parseJsonSafely(value) {
  if (typeof value !== 'string') {
    return value;
  }

  try {
    return JSON.parse(value);
  } catch (error) {
    return value;
  }
}

function buildAuthHeaders(apiKey, apiFormat = 'openai') {
  const headers = {
    'Content-Type': 'application/json'
  };

  if (apiFormat === 'anthropic' || apiFormat === 'anthropic_compat' || apiFormat === 'claude') {
    headers['x-api-key'] = apiKey;
    if (apiFormat === 'anthropic' || apiFormat === 'claude') {
      headers['anthropic-version'] = '2023-06-01';
    }
  } else if (apiFormat === 'gemini') {
    headers['x-goog-api-key'] = apiKey;
  } else {
    headers['Authorization'] = `Bearer ${apiKey}`;
  }

  return headers;
}

function buildApiUrl(baseUrl, apiFormat = 'openai', useFullUrl = false) {
  const trimmed = baseUrl.trim().replace(/\/+$/, '');
  
  if (useFullUrl) {
    return trimmed;
  }

  if (apiFormat === 'anthropic' || apiFormat === 'claude') {
    if (trimmed.endsWith('/v1/messages')) return trimmed;
    return `${trimmed}/v1/messages`;
  }

  if (apiFormat === 'openai' || apiFormat === 'anthropic_compat' || apiFormat === 'gemini') {
    if (trimmed.endsWith('/chat/completions')) return trimmed;
    if (trimmed.endsWith('/v1')) return `${trimmed}/chat/completions`;
    return `${trimmed}/v1/chat/completions`;
  }

  return trimmed;
}

function buildRequestBody(config, messages, options = {}) {
  const { stream = true, maxTokens } = options;
  const apiFormat = config.api_format || 'openai';

  if (apiFormat === 'anthropic' || apiFormat === 'claude') {
    const systemMessage = messages.find(m => m.role === 'system');
    const userMessages = messages.filter(m => m.role !== 'system');

    return {
      model: config.model,
      max_tokens: maxTokens || Math.min(Math.max(Number(config.max_tokens) || 2000, 1), 8192),
      system: systemMessage?.content || '',
      messages: userMessages.map(m => ({
        role: m.role === 'assistant' ? 'assistant' : 'user',
        content: m.content
      })),
      stream
    };
  }

  // openai, anthropic_compat, gemini 都使用 OpenAI 格式的请求体
  return {
    model: config.model,
    messages,
    temperature: config.temperature,
    max_tokens: maxTokens || Math.min(Math.max(Number(config.max_tokens) || 2000, 1), 8192),
    top_p: config.top_p ?? 0.9,
    frequency_penalty: config.frequency_penalty ?? 0.0,
    stream
  };
}

function buildRequestConfig(url, body, headers) {
  return {
    method: 'POST',
    headers,
    body: JSON.stringify(body)
  };
}

async function readResponsePayload(response) {
  const rawText = await response.text();

  if (!rawText) {
    return null;
  }

  return parseJsonSafely(rawText);
}

async function requestJson(url, body, headers, timeoutMs = 120000) {
  const config = buildRequestConfig(url, body, headers);
  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeoutMs);
  
  config.signal = controller.signal;
  
  let response;

  try {
    response = await fetch(url, config);
    clearTimeout(timeoutId);
  } catch (error) {
    clearTimeout(timeoutId);
    const err = new Error(error.name === 'AbortError' ? '请求超时' : error.message);
    err.request = true;
    err.config = config;
    err.cause = error;
    throw err;
  }

  const data = await readResponsePayload(response);

  if (!response.ok) {
    const err = new Error(`HTTP ${response.status}`);
    err.response = {
      status: response.status,
      statusText: response.statusText,
      data
    };
    err.config = config;
    throw err;
  }

  return { data };
}

function extractResponseContent(data) {
  if (data?.choices?.[0]) {
    return data.choices[0].message?.content || data.choices[0].text || '';
  }

  if (data?.content) {
    return data.content;
  }

  if (data?.output?.text) {
    return data.output.text;
  }

  if (typeof data?.output === 'string') {
    return data.output;
  }

  return typeof data === 'string' ? data : JSON.stringify(data);
}

function getErrorMessage(error, fallbackMessage = '请求失败') {
  if (error.response) {
    const responseData = error.response.data;
    return responseData?.error?.message
      || responseData?.error?.code
      || responseData?.message
      || `API错误(${error.response.status}): ${error.response.statusText}`;
  }

  return error.message || fallbackMessage;
}

function recordUsage(modelId, modelName, providerName, tokens = 0) {
  try {
    const today = new Date().toISOString().split('T')[0];
    const existing = db.prepare(`
      SELECT * FROM usage_stats WHERE date = ? AND model_id = ?
    `).get(today, modelId);

    if (existing) {
      db.prepare(`
        UPDATE usage_stats
        SET usage_count = usage_count + 1,
            total_tokens = total_tokens + ?,
            updated_at = CURRENT_TIMESTAMP
        WHERE date = ? AND model_id = ?
      `).run(tokens, today, modelId);
      return;
    }

    db.prepare(`
      INSERT INTO usage_stats (date, model_id, model_name, provider_name, usage_count, total_tokens)
      VALUES (?, ?, ?, ?, 1, ?)
    `).run(today, modelId, modelName, providerName, tokens);
  } catch (error) {
    console.error('Failed to record usage stats:', error.message);
  }
}

function getModelConfig(configId) {
  const baseQuery = `
    SELECT
      m.*,
      p.api_key,
      p.api_url,
      p.api_format,
      p.use_full_url,
      p.name AS provider_name
    FROM api_models m
    LEFT JOIN api_providers p ON m.provider_id = p.id
    WHERE m.enabled = 1
  `;

  if (configId) {
    return db.prepare(`
      ${baseQuery}
      AND m.id = ?
    `).get(configId);
  }

  return db.prepare(`
    ${baseQuery}
    AND m.is_default = 1
    LIMIT 1
  `).get();
}

module.exports = {
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
};
