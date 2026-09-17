// 阅读预览的正文清洗与章头解析
// 从 PhonePreview.vue 内联正则里抽出来，便于与编辑器共用、单独测试

/** 把编辑器产出的富文本 / markdown 片段洗成纯段落数组（预览只显示纯文本） */
export function toPlainParagraphs(content: string): string[] {
  const plain = (content || '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/<[^>]*>/g, '\n')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
  return plain
    .split(/\n+/)
    .map((s) => s.trim())
    .filter(Boolean)
}

/** 正文字数（去除所有空白，与中文写作的「字数」口径一致） */
export function countWords(content: string): number {
  return (content || '').replace(/\s/g, '').length
}

/** 预览用 HTML 转义（正文来自编辑器，拼 v-html 前再兜一层） */
export function escapeHtml(text: string): string {
  return (text || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

export interface ChapterHead {
  /** 章号，如「第一章」；无法识别时为空串 */
  no: string
  /** 章名，如「雪落长安」 */
  name: string
}

/**
 * 章节标题拆成「章号 + 章名」，用于版式化章头。
 * 先按中文/阿拉伯数字章号匹配，匹配不到再退回「首个空格」切分。
 */
export function splitChapterTitle(title: string): ChapterHead {
  const t = (title || '').trim()
  if (!t) return { no: '', name: '正文' }
  const m = t.match(/^(第\s*[零一二三四五六七八九十百千两0-9]+\s*[章节回卷篇][^\s·]*)\s*[·\s]*(.*)$/)
  if (m) return { no: m[1].replace(/\s+/g, ''), name: (m[2] || '').trim() }
  const sp = t.indexOf(' ')
  if (sp > 0) return { no: t.slice(0, sp), name: t.slice(sp + 1).trim() }
  return { no: '', name: t }
}
