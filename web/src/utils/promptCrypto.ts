// 提示词卡片内容加密（Prompts / PackDetail / PromptPreview 三个视图共用）
//
// ⚠ 这是**落库格式**的一部分，改动会导致已加密卡片无法解密：
//   - 密文格式 = ENCRYPT_MARKER + base64(salt[16] + iv[12] + AES-GCM ciphertext)
//   - 密钥 = PBKDF2(password, salt, 100000, SHA-256) → AES-GCM 256
// 追加新格式请新增 marker 并保留旧分支，不要原地改参数。

/** 加密内容前缀标记，用于识别「已加密」的卡片内容 */
export const ENCRYPT_MARKER = 'XNC1:'

/** 口令 → SHA-256 十六进制摘要（加固定盐；仅用于比对，不是安全强度保证） */
export const hashPassword = async (password: string): Promise<string> => {
  const encoder = new TextEncoder()
  const data = encoder.encode(password + '__xingnovel_salt__')
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}

/** 校验输入口令与存储的摘要是否一致 */
export const verifyPassword = async (inputPassword: string, storedHash: string): Promise<boolean> => {
  const inputHash = await hashPassword(inputPassword)
  return inputHash === storedHash
}

/** ArrayBuffer → base64（避免对超长二进制用 String.fromCharCode.apply 爆栈，逐字节拼接） */
export const bufferToBase64 = (buffer: ArrayBuffer): string => {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  bytes.forEach(b => binary += String.fromCharCode(b))
  return btoa(binary)
}

/** base64 → ArrayBuffer */
export const base64ToBuffer = (base64: string): ArrayBuffer => {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

/** 加密明文，返回带 ENCRYPT_MARKER 前缀的字符串 */
export const encryptContent = async (plaintext: string, password: string): Promise<string> => {
  const enc = new TextEncoder()
  const salt = crypto.getRandomValues(new Uint8Array(16))
  const keyMaterial = await crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, ['deriveKey'])
  const key = await crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
    keyMaterial,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt']
  )
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const ciphertext = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, enc.encode(plaintext))
  const combined = new Uint8Array(16 + 12 + ciphertext.byteLength)
  combined.set(salt, 0)
  combined.set(iv, 16)
  combined.set(new Uint8Array(ciphertext), 28)
  return ENCRYPT_MARKER + bufferToBase64(combined.buffer)
}

/**
 * 解密内容。
 * 无 ENCRYPT_MARKER 前缀时**原样返回**（明文卡片走这条路径，属正常情况）；
 * 解密失败（口令错 / 数据损坏）返回 null。
 */
export const decryptContent = async (wrapped: string, password: string): Promise<string | null> => {
  if (!wrapped.startsWith(ENCRYPT_MARKER)) {
    return wrapped
  }
  try {
    const enc = new TextEncoder()
    const dec = new TextDecoder()
    const combined = new Uint8Array(base64ToBuffer(wrapped.slice(ENCRYPT_MARKER.length)))
    const salt = combined.slice(0, 16)
    const iv = combined.slice(16, 28)
    const ciphertext = combined.slice(28)
    const keyMaterial = await crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, ['deriveKey'])
    const key = await crypto.subtle.deriveKey(
      { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['decrypt']
    )
    const plaintext = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext)
    return dec.decode(plaintext)
  } catch {
    return null
  }
}
