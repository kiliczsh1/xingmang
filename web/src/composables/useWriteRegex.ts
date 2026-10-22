// 正文 AI 输出正则过滤工具（流式输出清洗 / 显示内容过滤）
import { ref } from 'vue'

export type RegexRule = {
  id: string
  name: string
  pattern: string
  flags: string
  replacement: string
  enabled: boolean
  affectsActual: boolean
  order: number
  lastError?: string
}

const WRITE_REGEX_SETTINGS_STORAGE_KEY = 'write-chat-regex-settings-v1'

const createRegexRuleId = () => {
  const g = globalThis as any
  if (g.crypto?.randomUUID) return g.crypto.randomUUID()
  return `write_regex_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`
}

const createDefaultWriteRegexRules = (): RegexRule[] => {
  const make = (partial: Omit<RegexRule, 'id' | 'order'>, order: number): RegexRule => ({
    id: createRegexRuleId(),
    order,
    ...partial
  })

  return [
    make({
      name: '移除  thinking… response',
      pattern: ' thinking[\\s\\S]*?<\\/think>',
      flags: 'g',
      replacement: '',
      enabled: true,
      affectsActual: false
    }, 10),
    make({
      name: '移除 <analysis>…</analysis>',
      pattern: '<analysis>[\\s\\S]*?<\\/analysis>',
      flags: 'g',
      replacement: '',
      enabled: false,
      affectsActual: false
    }, 20)
  ]
}

export function useWriteRegex() {
  const writeDisplayRegexEnabled = ref(true)
  const writeStreamGuardEnabled = ref(true)
  const writeCopyUsesFiltered = ref(false)
  const writeRegexRules = ref<RegexRule[]>([])

  const normalizeWriteRegexRulesOrder = () => {
    writeRegexRules.value.forEach((rule, index) => {
      rule.order = index + 1
    })
  }

  const getActiveWriteRegexRules = (onlyAffectsActual = false) => {
    return writeRegexRules.value
      .filter(rule => rule.enabled && rule.pattern.trim() && (onlyAffectsActual ? rule.affectsActual : true))
      .slice()
      .sort((a, b) => a.order - b.order)
  }

  const applyWriteStreamingGuard = (text: string) => {
    if (!writeStreamGuardEnabled.value) return text

    const guards: Array<{ open: string; close: string }> = [
      { open: ' thinking', close: ' response' },
      { open: '<analysis>', close: '</analysis>' }
    ]

    const cutPositions: number[] = []
    for (const guard of guards) {
      const openIndex = text.lastIndexOf(guard.open)
      if (openIndex === -1) continue
      const closeIndex = text.indexOf(guard.close, openIndex + guard.open.length)
      if (closeIndex === -1) {
        cutPositions.push(openIndex)
      }
    }

    if (cutPositions.length === 0) return text
    return text.slice(0, Math.min(...cutPositions))
  }

  const applyWriteRegexPipeline = (text: string, onlyAffectsActual = false) => {
    if (!writeDisplayRegexEnabled.value && !onlyAffectsActual) return text

    let output = text
    const activeRules = getActiveWriteRegexRules(onlyAffectsActual)

    for (const rule of activeRules) {
      rule.lastError = undefined
      if (rule.pattern.length > 2000 || rule.replacement.length > 5000) {
        rule.lastError = '规则过长，已跳过'
        continue
      }

      try {
        const flags = rule.flags?.trim() || 'g'
        const regex = new RegExp(rule.pattern, flags)
        output = output.replace(regex, rule.replacement)
      } catch (error: any) {
        rule.lastError = error?.message || '无效正则'
      }
    }

    return output
  }

  const computeWriteAssistantDisplayContent = (rawContent: string) => {
    const guardedContent = applyWriteStreamingGuard(rawContent || '')
    return applyWriteRegexPipeline(guardedContent)
  }

  const loadWriteRegexSettings = () => {
    try {
      const raw = localStorage.getItem(WRITE_REGEX_SETTINGS_STORAGE_KEY)
      if (!raw) {
        writeRegexRules.value = createDefaultWriteRegexRules()
        normalizeWriteRegexRulesOrder()
        return
      }

      const parsed = JSON.parse(raw)
      writeDisplayRegexEnabled.value = parsed?.writeDisplayRegexEnabled !== false
      writeStreamGuardEnabled.value = parsed?.writeStreamGuardEnabled !== false
      writeCopyUsesFiltered.value = parsed?.writeCopyUsesFiltered === true

      if (Array.isArray(parsed?.writeRegexRules)) {
        writeRegexRules.value = parsed.writeRegexRules
          .map((item: any, index: number) => ({
            id: typeof item.id === 'string' && item.id ? item.id : createRegexRuleId(),
            name: typeof item.name === 'string' ? item.name : `规则 ${index + 1}`,
            pattern: typeof item.pattern === 'string' ? item.pattern : '',
            flags: typeof item.flags === 'string' ? item.flags : 'g',
            replacement: typeof item.replacement === 'string' ? item.replacement : '',
            enabled: typeof item.enabled === 'boolean' ? item.enabled : true,
            order: typeof item.order === 'number' ? item.order : index + 1
          }))
          .sort((a: RegexRule, b: RegexRule) => a.order - b.order)
        normalizeWriteRegexRulesOrder()
      } else {
        writeRegexRules.value = createDefaultWriteRegexRules()
        normalizeWriteRegexRulesOrder()
      }
    } catch {
      writeRegexRules.value = createDefaultWriteRegexRules()
      normalizeWriteRegexRulesOrder()
    }
  }

  const saveWriteRegexSettings = () => {
    const payload = {
      writeDisplayRegexEnabled: writeDisplayRegexEnabled.value,
      writeStreamGuardEnabled: writeStreamGuardEnabled.value,
      writeCopyUsesFiltered: writeCopyUsesFiltered.value,
      writeRegexRules: writeRegexRules.value.map(rule => ({
        id: rule.id,
        name: rule.name,
        pattern: rule.pattern,
        flags: rule.flags,
        replacement: rule.replacement,
        enabled: rule.enabled,
        order: rule.order
      }))
    }

    localStorage.setItem(WRITE_REGEX_SETTINGS_STORAGE_KEY, JSON.stringify(payload))
  }

  return {
    writeDisplayRegexEnabled,
    writeStreamGuardEnabled,
    writeCopyUsesFiltered,
    writeRegexRules,
    getActiveWriteRegexRules,
    applyWriteStreamingGuard,
    applyWriteRegexPipeline,
    computeWriteAssistantDisplayContent,
    loadWriteRegexSettings,
    saveWriteRegexSettings
  }
}
