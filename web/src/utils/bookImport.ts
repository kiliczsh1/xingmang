// 作品导入：文件读取 / 编码探测 / 标题清洗 / 章节切分
// 2026-09-21 从 views/Books.vue 抽出（第 2 期瘦身），函数体与原实现逐字等价

import type { Chapter } from '@/types'

export type ParsedImportChapter = Pick<Chapter, 'title' | 'content'>

export const CHAPTER_HEADING_REGEX = /^(?:\s| )*(第\s*[0-9零一二三四五六七八九十百千万两〇○O0-9①-⑳]+?\s*[章节卷回部篇集幕折][^\r\n]{0,30}|(?:序章|序言|前言|楔子|引子|终章|尾声|后记|番外(?:篇|合集)?|大结局)[^\r\n]{0,30}|chapter\s*[0-9ivxlcdm]+[^\r\n]{0,30})\s*$/gim

export const formatFileSize = (size: number) => {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  return (size / (1024 * 1024)).toFixed(1) + ' MB'
}

export const detectTextEncoding = (bytes: Uint8Array) => {
  if (bytes.length >= 3 && bytes[0] === 0xef && bytes[1] === 0xbb && bytes[2] === 0xbf) {
    return 'utf-8'
  }
  if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe) {
    return 'utf-16le'
  }
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    return 'utf-16be'
  }
  return ''
}

export const readTxtFile = async (file: File) => {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  const preferredEncoding = detectTextEncoding(bytes)
  const encodings = [preferredEncoding, 'utf-8', 'gb18030', 'big5'].filter(Boolean)
  let bestText = ''
  let bestScore = Number.POSITIVE_INFINITY
  for (const encoding of encodings) {
    try {
      const text = new TextDecoder(encoding).decode(bytes)
      const invalidCharCount = (text.match(/\uFFFD/g) || []).length
      if (invalidCharCount < bestScore) {
        bestScore = invalidCharCount
        bestText = text
      }
      if (invalidCharCount === 0) {
        break
      }
    } catch (error) {
    }
  }
  return bestText
    .replace(/\r\n?/g, '\n')
    .replace(/\u0000/g, '')
    .replace(/^\uFEFF/, '')
    .trim()
}

export const cleanImportTitle = (title: string, fallbackIndex: number) => {
  const normalizedTitle = title
    .replace(/^[\s ]+|[\s ]+$/g, '')
    .replace(/[：:]\s*$/, '')
    .replace(/\s+/g, ' ')
    .trim()
  return normalizedTitle || `第${fallbackIndex + 1}章`
}

// 「```」「~~~」这类 Markdown 代码围栏出现在导入的小说正文里基本都是复制残留，
// 留着会被渲染成代码块。只删「独占一行的围栏」（允许 ```python 这种语言标记），
// ⚠ 绝不能做全文替换：正文里的「“嗷~~~”」是拟声词，替掉就是削内容。
export const stripCodeFence = (text: string) =>
  text
    .replace(/^[ \t]*(?:```|~~~)[ \t]*[A-Za-z0-9+#._-]*[ \t]*$/gm, '')
    .replace(/\n{3,}/g, '\n\n')
    .trim()

// ⚠⚠ Markdown 里「4 个半角空格 / tab 开头」= 缩进式代码块。小说 txt 常用 4 空格做段首缩进，
// 直接导入会让整段在 Vditor / MarkdownRenderer 里变灰底代码块。统一换成全角空格缩进。
export const normalizeImportIndent = (text: string) =>
  text
    .split('\n')
    .map((line) =>
      line
        .replace(/^[\t ]{4,}/, '\u3000\u3000')
        .replace(/^[\t ]{1,3}(?=\S)/, '')
    )
    .join('\n')

const cleanChapters = (chapters: ParsedImportChapter[]) =>
  chapters
    .map((chapter) => ({ ...chapter, content: stripCodeFence(normalizeImportIndent(chapter.content)) }))
    .filter((chapter) => chapter.content.length > 0)

export const chunkPlainText = (text: string, baseTitle: string) => {
  const paragraphs = text
    .split(/\n{2,}/)
    .map(item => item.trim())
    .filter(Boolean)
  const chunks: ParsedImportChapter[] = []
  let currentChunk = ''
  for (const paragraph of paragraphs) {
    const nextChunk = currentChunk ? `${currentChunk}\n\n${paragraph}` : paragraph
    if (nextChunk.length > 6000 && currentChunk.length >= 1500) {
      chunks.push({ title: `第${chunks.length + 1}章`, content: currentChunk.trim() })
      currentChunk = paragraph
    } else {
      currentChunk = nextChunk
    }
  }
  if (currentChunk.trim()) {
    chunks.push({
      title: chunks.length === 0 ? `${baseTitle} 正文` : `第${chunks.length + 1}章`,
      content: currentChunk.trim()
    })
  }
  if (chunks.length > 1) {
    return chunks.map((chapter, index) => ({
      ...chapter,
      title: `第${index + 1}章`
    }))
  }
  return chunks
}

export const splitTxtIntoChapters = (text: string, baseTitle: string) => {
  const normalizedText = text.replace(/\n{3,}/g, '\n\n').trim()
  if (!normalizedText) {
    return []
  }
  const matches = Array.from(normalizedText.matchAll(CHAPTER_HEADING_REGEX))
  if (matches.length === 0) {
    return cleanChapters(chunkPlainText(normalizedText, baseTitle))
  }
  const chapters: ParsedImportChapter[] = []
  matches.forEach((match, index) => {
    const title = cleanImportTitle(match[0], index)
    const start = (match.index || 0) + match[0].length
    const end = index + 1 < matches.length ? (matches[index + 1].index || normalizedText.length) : normalizedText.length
    const content = normalizedText.slice(start, end).trim()
    if (content) {
      chapters.push({ title, content })
    }
  })
  if (chapters.length > 0) {
    return cleanChapters(chapters)
  }
  return cleanChapters(chunkPlainText(normalizedText, baseTitle))
}
