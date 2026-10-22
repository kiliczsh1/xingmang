import type { WorldBookEntry } from '@/types/worldbook'

export interface MatchResult {
  entry: WorldBookEntry
  matchedKeywords: string[]
}

export function matchWorldBookEntries(
  text: string,
  entries: WorldBookEntry[],
  scanDepth: number = 5
): MatchResult[] {
  const results: MatchResult[] = []
  const textToScan = getTextToScan(text, scanDepth)

  for (const entry of entries) {
    if (entry.disable) continue

    if (entry.constant) {
      results.push({
        entry,
        matchedKeywords: ['[常驻]']
      })
      continue
    }

    const matchedKeywords: string[] = []
    
    for (const keyword of entry.key) {
      if (keywordMatches(keyword, textToScan, entry)) {
        if (entry.keysecondary.length === 0) {
          matchedKeywords.push(keyword)
        } else {
          const secondaryMatched = entry.keysecondary.some(secondaryKeyword =>
            keywordMatches(secondaryKeyword, textToScan, entry)
          )
          if (secondaryMatched) {
            matchedKeywords.push(keyword)
          }
        }
      }
    }

    if (matchedKeywords.length > 0) {
      if (entry.useProbability && entry.probability < 100) {
        if (Math.random() * 100 > entry.probability) {
          continue
        }
      }

      results.push({
        entry,
        matchedKeywords
      })
    }
  }

  results.sort((a, b) => a.entry.order - b.entry.order)

  return results
}

function getTextToScan(text: string, depth: number): string {
  const lines = text.split('\n')
  const relevantLines = lines.slice(-depth)
  return relevantLines.join('\n')
}

function keywordMatches(
  keyword: string,
  text: string,
  entry: WorldBookEntry
): boolean {
  if (!keyword.trim()) return false

  let searchKeyword = keyword
  let searchText = text

  if (!entry.caseSensitive) {
    searchKeyword = keyword.toLowerCase()
    searchText = text.toLowerCase()
  }

  if (entry.matchWholeWords) {
    const regex = new RegExp(`\\b${escapeRegex(searchKeyword)}\\b`, entry.caseSensitive ? 'g' : 'gi')
    return regex.test(searchText)
  } else {
    return searchText.includes(searchKeyword)
  }
}

function escapeRegex(str: string): string {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

export function injectWorldBookContent(
  basePrompt: string,
  matchResults: MatchResult[]
): string {
  const injections: { position: number; content: string; order: number }[] = []

  for (const result of matchResults) {
    const { entry } = result
    injections.push({
      position: entry.position,
      content: entry.content,
      order: entry.order
    })
  }

  injections.sort((a, b) => a.order - b.order)

  const beforePrompt: string[] = []
  const afterPrompt: string[] = []
  const topOfWorldInfo: string[] = []
  const bottomOfWorldInfo: string[] = []

  for (const injection of injections) {
    switch (injection.position) {
      case 0:
        beforePrompt.push(injection.content)
        break
      case 1:
      case 2:
      case 3:
      case 4:
      case 5:
      case 6:
        afterPrompt.push(injection.content)
        break
      case 7:
        topOfWorldInfo.push(injection.content)
        break
      case 8:
      default:
        bottomOfWorldInfo.push(injection.content)
        break
    }
  }

  let finalPrompt = basePrompt

  if (topOfWorldInfo.length > 0) {
    const worldInfoSection = topOfWorldInfo.join('\n\n')
    finalPrompt = worldInfoSection + '\n\n' + finalPrompt
  }

  if (bottomOfWorldInfo.length > 0) {
    const worldInfoSection = bottomOfWorldInfo.join('\n\n')
    finalPrompt = finalPrompt + '\n\n' + worldInfoSection
  }

  if (beforePrompt.length > 0) {
    finalPrompt = beforePrompt.join('\n\n') + '\n\n' + finalPrompt
  }

  if (afterPrompt.length > 0) {
    finalPrompt = finalPrompt + '\n\n' + afterPrompt.join('\n\n')
  }

  return finalPrompt
}

export function processWorldBookForAI(
  prompt: string,
  entries: WorldBookEntry[],
  scanDepth: number = 5
): string {
  const matchResults = matchWorldBookEntries(prompt, entries, scanDepth)
  
  if (matchResults.length === 0) {
    return prompt
  }

  return injectWorldBookContent(prompt, matchResults)
}

export function estimateTokens(text: string): number {
  const chineseChars = (text.match(/[\u4e00-\u9fa5]/g) || []).length
  const englishWords = text.replace(/[\u4e00-\u9fa5]/g, ' ').split(/\s+/).filter(w => w.length > 0).length
  const otherChars = text.replace(/[\u4e00-\u9fa5]/g, '').replace(/\s+/g, '').length
  
  return Math.ceil(chineseChars * 1.5 + englishWords * 1.3 + otherChars * 0.5)
}

export function validateWorldBookEntry(entry: WorldBookEntry): string[] {
  const errors: string[] = []

  if (!entry.constant && entry.key.length === 0) {
    errors.push('非常驻条目必须至少有一个触发关键词')
  }

  if (!entry.content.trim()) {
    errors.push('条目内容不能为空')
  }

  if (entry.probability < 0 || entry.probability > 100) {
    errors.push('触发概率必须在 0-100 之间')
  }

  if (entry.depth < 1 || entry.depth > 999) {
    errors.push('扫描深度必须在 1-999 之间')
  }

  return errors
}
