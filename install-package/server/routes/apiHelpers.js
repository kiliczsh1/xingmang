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

function buildAuthHeaders(apiKey) {
  return {
    Authorization: `Bearer ${apiKey}`,
    'Content-Type': 'application/json'
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

async function requestJson(url, body, headers) {
  const config = buildRequestConfig(url, body, headers);
  let response;

  try {
    response = await fetch(url, config);
  } catch (error) {
    const err = new Error(error.message);
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
      p.name AS provider_name
    FROM api_models m
    LEFT JOIN api_providers p ON m.provider_id = p.id
  `;

  if (configId) {
    return db.prepare(`
      ${baseQuery}
      WHERE m.id = ?
    `).get(configId);
  }

  return db.prepare(`
    ${baseQuery}
    WHERE m.is_default = 1
    LIMIT 1
  `).get();
}

module.exports = {
  parseJsonSafely,
  buildAuthHeaders,
  buildRequestConfig,
  readResponsePayload,
  requestJson,
  extractResponseContent,
  getErrorMessage,
  recordUsage,
  getModelConfig
};
