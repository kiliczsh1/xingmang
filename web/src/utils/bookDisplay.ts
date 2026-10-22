// Books 列表卡片 / 信息弹窗 / 完结弹窗 共用的展示函数
// 2026-09-21 第 3 期拆弹窗时建立（此前只服务单视图，故一直留在视图内）

import type { Book } from '@/types'

export const DEFAULT_TARGET_WORDS = 100000

export const targetWordsOf = (book: Book) => book.target_words || DEFAULT_TARGET_WORDS

export const bookProgress = (book: Book) => {
  const wc = book.word_count || 0
  const tw = targetWordsOf(book)
  return Math.min(100, Math.round((wc / tw) * 100))
}

export const formatWordCount = (n: number) => {
  if (n >= 10000) return (n / 10000).toFixed(1) + ' 万'
  return n.toLocaleString()
}

export const getStatusText = (status?: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    completed: '已完结'
  }
  return statusMap[status || 'draft'] || '草稿'
}
