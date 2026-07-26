// 正则过滤组合函数
// 从 ChatDialog.vue 行 294-543 搬迁的逻辑，统一供 ChatPanel 和 ChatDialog 复用
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'

export type RegexRule = {
  id: string
  name: string
  pattern: string
  flags: string
  replacement: string
  enabled: boolean
  affectsActual?: boolean
  order: number
  lastError?: string
}

const generateId = () => {
  const g = globalThis as any
  if (g.crypto?.randomUUID) return g.crypto.randomUUID()
  return `r_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`
}

const THINK_OPEN = '\u003Cthink\u003E'
const THINK_CLOSE = '\u003C/think\u003E'
const ANALYSIS_OPEN = '\u003Canalysis\u003E'
const ANALYSIS_CLOSE = '\u003C/analysis\u003E'

const getDefaultRegexRules = (): RegexRule[] => {
  const make = (partial: Omit<RegexRule, 'id' | 'order'>, order: number): RegexRule => ({
    id: generateId(),
    order,
    ...partial,
  })
  return [
    make({
      name: '移除 think 标签',
      pattern: THINK_OPEN + '[\\s\\S]*?' + THINK_CLOSE,
      flags: 'g',
      replacement: '',
      enabled: true,
    }, 10),
    make({
      name: '移除 analysis 标签',
      pattern: ANALYSIS_OPEN + '[\\s\\S]*?' + ANALYSIS_CLOSE,
      flags: 'g',
      replacement: '',
      enabled: false,
    }, 20),
  ]
}

export function useRegexFilter(storageKey: string) {
  const displayRegexEnabled = ref(true)
  const streamGuardEnabled = ref(true)
  const copyUsesFiltered = ref(false)
  const regexRules = ref<RegexRule[]>([])
  const regexTestInput = ref('')
  const regexRulesJson = ref('')

  const normalizeOrder = () => {
    regexRules.value.forEach((r, idx) => {
      r.order = idx + 1
    })
  }

  const addRegexRule = () => {
    regexRules.value.push({
      id: generateId(),
      name: '新规则',
      pattern: '',
      flags: 'g',
      replacement: '',
      enabled: true,
      order: regexRules.value.length + 1,
    })
    normalizeOrder()
  }

  const removeRule = (id: string) => {
    const idx = regexRules.value.findIndex(r => r.id === id)
    if (idx >= 0) {
      regexRules.value.splice(idx, 1)
      normalizeOrder()
    }
  }

  const moveRule = (index: number, direction: -1 | 1) => {
    const nextIndex = index + direction
    if (nextIndex < 0 || nextIndex >= regexRules.value.length) return
    const moved = regexRules.value.splice(index, 1)[0]
    regexRules.value.splice(nextIndex, 0, moved)
    normalizeOrder()
  }

  const resetRegexRules = () => {
    regexRules.value = getDefaultRegexRules()
    normalizeOrder()
  }

  const exportRegexRules = () => {
    regexRulesJson.value = JSON.stringify(
      regexRules.value.map(r => ({
        id: r.id,
        name: r.name,
        pattern: r.pattern,
        flags: r.flags,
        replacement: r.replacement,
        enabled: r.enabled,
        affectsActual: r.affectsActual,
        order: r.order,
      })),
      null,
      2
    )
  }

  const importRegexRules = () => {
    try {
      const parsed = JSON.parse(regexRulesJson.value)
      if (!Array.isArray(parsed)) {
        ElMessage.error('导入失败：JSON 必须是数组')
        return
      }
      const nextRules: RegexRule[] = parsed.map((item: any, idx: number) => ({
        id: typeof item.id === 'string' && item.id ? item.id : generateId(),
        name: typeof item.name === 'string' ? item.name : `规则 ${idx + 1}`,
        pattern: typeof item.pattern === 'string' ? item.pattern : '',
        flags: typeof item.flags === 'string' ? item.flags : 'g',
        replacement: typeof item.replacement === 'string' ? item.replacement : '',
        enabled: typeof item.enabled === 'boolean' ? item.enabled : true,
        affectsActual: typeof item.affectsActual === 'boolean' ? item.affectsActual : false,
        order: typeof item.order === 'number' ? item.order : idx + 1,
      }))
      regexRules.value = nextRules.sort((a, b) => a.order - b.order)
      normalizeOrder()
      ElMessage.success('导入成功')
    } catch (e: any) {
      ElMessage.error('导入失败：' + (e?.message || 'JSON 解析错误'))
    }
  }

  const copyRegexRulesJson = async () => {
    if (!regexRulesJson.value) exportRegexRules()
    try {
      await navigator.clipboard.writeText(regexRulesJson.value)
      ElMessage.success('JSON 已复制')
    } catch {
      ElMessage.error('复制失败')
    }
  }

  // ---- 过滤管道 ----

  const applyStreamingGuard = (text: string) => {
    if (!streamGuardEnabled.value) return text
    const guards = [
      { open: THINK_OPEN, close: THINK_CLOSE },
      { open: ANALYSIS_OPEN, close: ANALYSIS_CLOSE },
    ]

    const cutPositions: number[] = []
    for (const g of guards) {
      const openIndex = text.lastIndexOf(g.open)
      if (openIndex === -1) continue
      const closeIndex = text.indexOf(g.close, openIndex + g.open.length)
      if (closeIndex === -1) {
        cutPositions.push(openIndex)
      }
    }

    if (cutPositions.length === 0) return text
    return text.slice(0, Math.min(...cutPositions))
  }

  const getActiveRules = (onlyAffectsActual = false) => {
    return regexRules.value
      .filter(r => {
        if (!r.enabled || !r.pattern.trim()) return false
        if (onlyAffectsActual) return r.affectsActual === true
        return true
      })
      .slice()
      .sort((a, b) => a.order - b.order)
  }

  const applyRegexPipeline = (text: string, onlyAffectsActual = false) => {
    if (!displayRegexEnabled.value && !onlyAffectsActual) return text
    let out = text

    const activeRules = getActiveRules(onlyAffectsActual)
    for (const rule of activeRules) {
      rule.lastError = undefined
      if (rule.pattern.length > 2000 || rule.replacement.length > 5000) {
        rule.lastError = '规则过长，已跳过'
        continue
      }
      try {
        const flags = rule.flags?.trim() || 'g'
        const re = new RegExp(rule.pattern, flags)
        out = out.replace(re, rule.replacement)
      } catch (e: any) {
        rule.lastError = e?.message || '无效正则'
      }
    }
    return out
  }

  const computeAssistantDisplay = (raw: string) => {
    const guarded = applyStreamingGuard(raw || '')
    return applyRegexPipeline(guarded)
  }

  const applyActualPipeline = (text: string) => {
    return applyRegexPipeline(text, true)
  }

  const regexTestOutput = computed(() => {
    return computeAssistantDisplay(regexTestInput.value)
  })

  // ---- 持久化 ----

  const loadRegexSettings = () => {
    try {
      const raw = localStorage.getItem(storageKey)
      if (!raw) {
        resetRegexRules()
        return
      }
      const parsed = JSON.parse(raw)
      displayRegexEnabled.value = parsed?.displayRegexEnabled !== false
      streamGuardEnabled.value = parsed?.streamGuardEnabled !== false
      copyUsesFiltered.value = parsed?.copyUsesFiltered === true
      if (Array.isArray(parsed?.regexRules)) {
        regexRules.value = parsed.regexRules.map((item: any, idx: number) => ({
          id: typeof item.id === 'string' && item.id ? item.id : generateId(),
          name: typeof item.name === 'string' ? item.name : `规则 ${idx + 1}`,
          pattern: typeof item.pattern === 'string' ? item.pattern : '',
          flags: typeof item.flags === 'string' ? item.flags : 'g',
          replacement: typeof item.replacement === 'string' ? item.replacement : '',
          enabled: typeof item.enabled === 'boolean' ? item.enabled : true,
          affectsActual: typeof item.affectsActual === 'boolean' ? item.affectsActual : false,
          order: typeof item.order === 'number' ? item.order : idx + 1,
        })).sort((a: RegexRule, b: RegexRule) => a.order - b.order)
        normalizeOrder()
      } else {
        resetRegexRules()
      }
    } catch {
      resetRegexRules()
    }
  }

  const saveRegexSettings = () => {
    const payload = {
      displayRegexEnabled: displayRegexEnabled.value,
      streamGuardEnabled: streamGuardEnabled.value,
      copyUsesFiltered: copyUsesFiltered.value,
      regexRules: regexRules.value.map(r => ({
        id: r.id,
        name: r.name,
        pattern: r.pattern,
        flags: r.flags,
        replacement: r.replacement,
        enabled: r.enabled,
        affectsActual: r.affectsActual,
        order: r.order,
      })),
    }
    localStorage.setItem(storageKey, JSON.stringify(payload))
  }

  watch([displayRegexEnabled, streamGuardEnabled, copyUsesFiltered], () => {
    saveRegexSettings()
  }, { deep: false })

  watch(regexRules, () => {
    saveRegexSettings()
  }, { deep: true })

  loadRegexSettings()

  return {
    displayRegexEnabled,
    streamGuardEnabled,
    copyUsesFiltered,
    regexRules,
    regexTestInput,
    regexTestOutput,
    regexRulesJson,
    addRegexRule,
    removeRule,
    moveRule,
    resetRegexRules,
    exportRegexRules,
    importRegexRules,
    copyRegexRulesJson,
    applyStreamingGuard,
    applyRegexPipeline,
    applyActualPipeline,
    computeAssistantDisplay,
    loadRegexSettings,
    saveRegexSettings,
  }
}