/**
 * 模板替换 —— 直译源项目 workflow-engine.js 的 compileTemplate，
 * 并与前端执行器里的 replaceVariables 行为合并成一套（前者是后者的超集）。
 *
 * 支持：
 * - {{#if var}}...{{/if}}，var 为数组时以 length>0 为真，其他值以真值判断；支持嵌套
 * - {{var}}                     普通变量
 * - {{obj.prop}} / {{a.b.c}}    属性访问
 * - 未定义的值保持原样（{{x}} 原样留着，便于用户看出哪里没接上）
 */

/**
 * 只做 `{{word}}` 替换（源执行器 aiProcessExecutor / textExecutor 的原语义）。
 * 变量名支持中文（\w 只有 ASCII，星域占位符常为中文名，如 {{故事大纲}} —— 不加会静默不替换）。
 */
export function replaceVariables(text: string, ctx: Record<string, any>): string {
  return text.replace(/\{\{([\p{L}\p{N}_]+)\}\}/gu, (match, key: string) => {
    if (ctx[key] === undefined) return match
    return stringifyValue(ctx[key])
  })
}

function stringifyValue(value: any): string {
  if (value === null) return 'null'
  if (typeof value === 'object') {
    try {
      return JSON.stringify(value, null, 2)
    } catch {
      return String(value)
    }
  }
  return String(value)
}

/** 完整模板（含 #if 与属性访问） */
export function compileTemplate(template: string, data: Record<string, any>): string {
  if (typeof template !== 'string') return ''
  let result = template

  // 先处理 {{#if}}，避免块内的变量替换把条件标记吃掉
  result = result.replace(/\{\{#if\s+([\w.]+)\}\}([\s\S]*?)\{\{\/if\}\}/g, (_m, key: string, content: string) => {
    const value = readPath(data, key)
    const truthy = Array.isArray(value) ? value.length > 0 : Boolean(value)
    return truthy ? compileTemplate(content, data) : ''
  })

  result = result.replace(/\{\{([^}]+)\}\}/g, (match, expression: string) => {
    const key = expression.trim()
    const value = readPath(data, key)
    if (value === undefined) return match
    return stringifyValue(value)
  })

  return result
}

/** 读取 a.b.c 形式的路径，首段必须是自有键 */
function readPath(data: Record<string, any>, path: string): any {
  if (!data || typeof data !== 'object') return undefined
  const parts = path.split('.')
  let cur: any = data
  for (const part of parts) {
    if (cur === null || cur === undefined) return undefined
    if (typeof cur !== 'object') return undefined
    if (!Object.prototype.hasOwnProperty.call(cur, part)) return undefined
    cur = cur[part]
  }
  return cur
}

export default compileTemplate
