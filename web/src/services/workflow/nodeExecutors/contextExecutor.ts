import type { NodeExecuteResult, NodeLogger, WNode, WorkflowExecutorHost } from '@/types/workflow'

/**
 * 上下文节点 —— 直译源项目 contextExecutor。
 * action: save / append / parseJson / pick / load / clear / saveChapter
 * 兼容旧字段别名（sourceVariables / sourceVariable / outputVariable / …）。
 */
export class ContextExecutor {
  host: WorkflowExecutorHost

  constructor(host: WorkflowExecutorHost) {
    this.host = host
  }

  async execute(node: WNode, context: Record<string, any>, logger: NodeLogger): Promise<NodeExecuteResult> {
    const config = node.data?.config || {}

    const normalizeAction = (raw: any): string => {
      const s = String(raw || '').trim().toLowerCase()
      if (!s || ['save', 'savevariable', 'store', 'save_var', 'savevar'].includes(s)) return 'save'
      if (['append', 'save_to_history', 'appendhistory', 'addhistory', 'pushhistory', 'add', 'push'].includes(s)) return 'append'
      if (['parsejson', 'parse_json', 'json_parse', 'json'].includes(s)) return 'parseJson'
      if (['pick', 'select', 'select_from_array', 'choose'].includes(s)) return 'pick'
      if (['load', 'get', 'fetch'].includes(s)) return 'load'
      if (['clear', 'reset', 'remove', 'delete'].includes(s)) return 'clear'
      if (['savechapter', 'save_chapter', 'savechapterfile', 'save_chapter_file'].includes(s)) return 'saveChapter'
      return raw || 'save'
    }
    const pickName = (c: any) =>
      c.variableName || c.outputVariable || c.variable || c.var || c.name || c.targetVariable
    const pickSource = (c: any) =>
      c.sourceVariable || c.source || c.from || (Array.isArray(c.sourceVariables) ? c.sourceVariables[0] : undefined)

    config.action = normalizeAction(config.action)
    if (!config.variableName) config.variableName = pickName(config)
    if (!config.sourceVariable && !config.sourceVariables) {
      const s = pickSource(config)
      if (s) config.sourceVariable = s
    }

    const action = config.action || 'save'
    const variableName = config.variableName

    if (['save', 'append', 'parseJson', 'pick', 'load', 'clear'].includes(action) && !variableName) {
      throw new Error('未设置变量名')
    }

    logger.info(`💾 上下文操作: ${action} - ${variableName}`)

    /** 收集源变量名：新格式 inputVariables[{variableName}] → 旧格式 sourceVariables → sourceVariable → lastOutput */
    const collectSources = (): string[] => {
      if (Array.isArray(config.inputVariables) && config.inputVariables.length > 0) {
        const names = config.inputVariables.map((r: any) => r.variableName).filter(Boolean)
        logger.info(`✅ 使用新格式输入变量: [${names.join(', ')}]`)
        return names
      }
      if (Array.isArray(config.sourceVariables) && config.sourceVariables.length > 0) {
        logger.info(`⚠️ 使用旧格式源变量: [${config.sourceVariables.join(', ')}]`)
        return [...config.sourceVariables]
      }
      if (config.sourceVariable) return [config.sourceVariable]
      return ['lastOutput']
    }

    switch (action) {
      case 'save': {
        const sources = collectSources()
        if (sources.length === 1) {
          const src = sources[0]
          if (context[src] !== undefined) {
            context[variableName] = context[src]
            logger.success(`✅ 已保存 ${src} 到 ${variableName}`)
          } else {
            logger.warning(`⚠️ 源变量 ${src} 不存在，跳过`)
          }
        } else {
          const bag: Record<string, any> = {}
          let count = 0
          for (const s of sources) {
            if (context[s] !== undefined) {
              bag[s] = context[s]
              count++
            }
          }
          context[variableName] = bag
          logger.success(`✅ 已保存 ${count}/${sources.length} 个源变量到 ${variableName}`)
        }
        break
      }

      case 'parseJson': {
        const sources = collectSources()
        const src = sources[0]
        const raw = context[src]
        if (typeof raw !== 'string') {
          logger.warning(`⚠️ 源变量 ${src} 不是字符串，当前类型: ${typeof raw}，无法解析`)
          break
        }
        try {
          let text = raw.trim()
          const fence = text.match(/```json\s*([\s\S]*?)\s*```/i)
          if (fence) {
            text = fence[1].trim()
          } else {
            const braceIdx = text.indexOf('{')
            const bracketIdx = text.indexOf('[')
            let startIdx = -1
            let endIdx = -1
            if (braceIdx !== -1 && (bracketIdx === -1 || braceIdx < bracketIdx)) {
              startIdx = braceIdx
              endIdx = text.lastIndexOf('}')
            } else if (bracketIdx !== -1) {
              startIdx = bracketIdx
              endIdx = text.lastIndexOf(']')
            }
            if (startIdx !== -1 && endIdx !== -1 && endIdx > startIdx) {
              text = text.substring(startIdx, endIdx + 1)
            }
          }
          let parsed: any = JSON.parse(text)
          if (parsed && parsed.chapters && Array.isArray(parsed.chapters)) {
            parsed = parsed.chapters
          } else if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
            for (const k of ['items', 'list', 'data', 'results']) {
              if (Array.isArray(parsed[k])) {
                parsed = parsed[k]
                break
              }
            }
          }
          context[variableName] = parsed
          const n = Array.isArray(parsed) ? parsed.length : parsed ? Object.keys(parsed).length : 0
          logger.success(`✅ 已解析 JSON 到 ${variableName}（${Array.isArray(parsed) ? '数组' : '对象'}，${n}）`)
        } catch (e: any) {
          logger.warning(`⚠️ JSON 解析失败: ${e?.message || e}`)
        }
        break
      }

      case 'pick': {
        const arrayVar = config.arrayVariable || 'chapterPlans'
        const indexVar = config.indexVariable || 'currentIteration'
        const arr = context[arrayVar]
        const idx = Number(context[indexVar] || 1) - 1
        if (!Array.isArray(arr)) {
          logger.warning(`⚠️ 变量 ${arrayVar} 不是数组，无法选择`)
          break
        }
        if (idx < 0 || idx >= arr.length) {
          logger.warning(`⚠️ 索引超界：${indexVar}=${idx + 1}，数组长度=${arr.length}`)
          break
        }
        const item = arr[idx]
        context[variableName] = item
        const title = item && (item.title || item.name || item.chapterTitle)
        if (title) context.currentChapterTitle = title
        logger.success(`✅ 已从 ${arrayVar}[${idx}] 选择到 ${variableName}${title ? `（${title}）` : ''}`)
        break
      }

      case 'load':
        if (context[variableName] !== undefined) logger.success(`✅ 变量 ${variableName} 可用`)
        else logger.warning(`⚠️ 变量 ${variableName} 不存在`)
        break

      case 'clear':
        if (context[variableName] !== undefined) {
          delete context[variableName]
          logger.success(`✅ 已清除 ${variableName}`)
        }
        break

      case 'saveChapter': {
        let index: any = context[config.chapterIndex || 'currentIteration']
        if (!index) index = context.chapterNum
        if (!index) index = 1
        const plans = context[config.chapterPlans || 'chapterPlans'] || []
        const content = context[config.content || 'aiResult'] || ''
        const plan = Array.isArray(plans) ? plans[index - 1] : null
        const title =
          (plan && (plan.title || plan.name || plan.chapterTitle || plan['标题'] || plan['章节标题'] || plan['题目'])) ||
          `第${index}章`
        try {
          await this.host.saveChapter(title, content, index)
          logger.success(`✅ 已保存章节: ${String(index).padStart(3, '0')}_${String(title).replace(/[\\/:*?"<>|]/g, '_').trim()}.md`)
        } catch (e: any) {
          logger.error(`❌ 保存章节失败: ${e?.message || e}`)
        }
        break
      }

      case 'append': {
        const sources = collectSources()
        if (!context[variableName]) context[variableName] = []
        if (!Array.isArray(context[variableName])) {
          logger.warning(`⚠️ ${variableName} 不是数组，无法追加`)
          break
        }
        let value: any = null
        if (sources.length === 1) {
          value = context[sources[0]]
        } else {
          const bag: Record<string, any> = {}
          for (const s of sources) if (context[s] !== undefined) bag[s] = context[s]
          value = bag
        }
        if (config.includeTitle && context.chapterPlans && context.currentIteration !== undefined) {
          const idx = context.currentIteration - 1
          const plans = context.chapterPlans
          if (Array.isArray(plans) && plans[idx]) {
            const plan = plans[idx]
            const title = plan.title || plan['标题'] || plan.name || `第${context.currentIteration}章`
            if (typeof value === 'string') {
              value = { title, content: value, chapter: context.currentIteration }
              logger.info(`📖 已添加章节标题: ${title}`)
            }
          }
        }
        if (value !== undefined) {
          context[variableName].push(value)
          const maxHistory = config.maxHistory || 10
          if (context[variableName].length > maxHistory) {
            context[variableName] = context[variableName].slice(-maxHistory)
          }
          logger.success(`✅ 已追加到 ${variableName} (${context[variableName].length} 条)`)
        } else {
          logger.warning('⚠️ 源变量不存在，未追加')
        }
        break
      }

      default:
        logger.warning(`⚠️ 未知的上下文操作: ${action}`)
    }

    return { outputs: { [variableName]: context[variableName] }, data: context[variableName] }
  }
}

export default ContextExecutor
