/**
 * AI 图谱抽取的编排层：逐章调用现有的 /api/ai/chat（SSE），
 * 把流式文本交给 utils/aiGraphExtraction 解析成三元组，最后汇总成可导入的图谱数据。
 *
 * 走的是项目里已验证的 AI 通道（同 AiAnalyzeDialog / AiSidePanel）：
 * { messages, systemPrompts, configId } -> data: {"content":"..."} -> data: [DONE]
 */
import type { GraphEntity, GraphRelation } from '@/types'
import {
  accumulatorNames,
  accumulatorToGraph,
  addToAccumulator,
  buildExtractionPrompt,
  buildOrphanLinkPrompt,
  chunkText,
  createAccumulator,
  entityKey,
  extractJsonPayload,
  normalizeExtractedGraph,
  normalizeName,
  relationKey,
  type GraphAccumulator
} from '@/utils/aiGraphExtraction'
import { mergeGraphData } from '@/services/graphMergeService'

export interface AnalyzeChapter {
  id: number
  title: string
  content: string
}

export interface AnalyzeProgress {
  /** 已完成的章节数 */
  done: number
  total: number
  currentTitle: string
  entityCount: number
  relationCount: number
}

export interface AnalyzeStats {
  entityCount: number
  relationCount: number
  /** 端点无法解析等原因被丢弃的关系数 */
  droppedRelations: number
  duplicateEntities: number
  duplicateRelations: number
  /** 与当前图谱已有实体同名而复用旧实体的数量（不生成新实体行） */
  reusedWithBase: number
  /** 请求失败的文本块数（其余章节照常继续） */
  failedChunks: number
  /** 请求成功但解析不出图谱结构的文本块数（模型没按格式输出，或 JSON 被截断） */
  parseFailures: number
  /** 请求成功但正文为空、思考内容非空的文本块数（输出预算被思考耗尽） */
  emptyChunks: number
  chapterCount: number
  /** 用户中途取消：结果为取消前已抽取的部分数据，不视为失败 */
  cancelled: boolean
}

export interface AnalyzeResult {
  entities: GraphEntity[]
  relations: GraphRelation[]
  stats: AnalyzeStats
}

/** 当前图谱已有实体快照：抽取时注入名字，让指向旧实体的关系能解析成功 */
export interface AnalyzeBaseEntity {
  id: number
  name: string
  type: string
}

export interface AnalyzeOptions {
  bookId: number
  chapters: AnalyzeChapter[]
  configId: number
  /** 拆分方式：split 分章拆（逐章串行请求）；merge 合并拆（所选章节拼成一份正文一次跑完，默认） */
  mode?: AnalyzeMode
  /** 累加合并：当前版本的已有实体，AI 抽取时其名字参与关系端点白名单 */
  baseEntities?: AnalyzeBaseEntity[]
  signal?: AbortSignal
  onProgress?: (progress: AnalyzeProgress) => void
}

export type AnalyzeMode = 'split' | 'merge'

interface ChatPayload {
  messages: Array<{ role: string; content: string }>
  systemPrompts: string[]
  configId: number
}

/** streamChat 的结果：正文与思考分开返回，正文为空时能据此给出可诊断的错误 */
interface ChatStreamResult {
  content: string
  reasoning: string
}

/**
 * 调用 /api/ai/chat，累积 SSE 文本流后整段返回。
 *
 * ⚠ 不传 maxTokens：输出上限交给模型配置（auto 模式走后端探测）。
 * 早期这里写死 8192，对**所有**模型生效 —— 带思考的模型把预算大量花在 reasoning 上，
 * 正文 JSON 写到一半就被截断，解析失败后表现为「抽不到实体」。
 * 更糟的是这个硬编码值会被后端探测当成「用户显式上限」写进 model_max_tokens 缓存，
 * 之后即使不传也一直按 8192 起步。
 */
async function streamChat(payload: ChatPayload, signal?: AbortSignal): Promise<ChatStreamResult> {
  const res = await fetch('/api/ai/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
    signal
  })
  if (!res.ok) {
    throw new Error(`AI 请求失败（HTTP ${res.status}）`)
  }
  if (!res.body) {
    throw new Error('AI 未返回数据流')
  }

  const reader = res.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  let content = ''
  let reasoning = ''

  for (;;) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })

    const lines = buffer.split('\n')
    // 最后一段可能是半行，留到下次拼接
    buffer = lines.pop() ?? ''

    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed.startsWith('data:')) continue
      const data = trimmed.slice(5).trim()
      if (!data || data === '[DONE]') continue

      let parsed: any
      try {
        parsed = JSON.parse(data)
      } catch {
        continue
      }
      if (parsed?.error) throw new Error(String(parsed.error))
      if (typeof parsed?.content === 'string' && parsed.content) {
        content += parsed.content
      }
      if (typeof parsed?.reasoning === 'string' && parsed.reasoning) {
        reasoning += parsed.reasoning
      }
    }
  }

  return { content, reasoning }
}

function isAbort(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

/** 一次 AI 请求对应的正文块；chapterIds 用于把抽到的实体/关系挂到来源章节 */
interface AnalyzeJob {
  title: string
  content: string
  chapterIds: number[]
}

/** 长文切块参数，与 utils/aiGraphExtraction.chunkText 保持一致 */
const CHUNK_SIZE = 4000
const CHUNK_OVERLAP = 400

/**
 * 按模式把所选章节拆成若干待请求的分析任务：
 * - 分章拆（split）：每章各自按窗口切块，逐块请求
 * - 合并拆（merge）：把所有章节拼成一份正文再切块，一次跑完，跨章关系更完整
 */
function buildJobs(chapters: AnalyzeChapter[], mode: AnalyzeMode): AnalyzeJob[] {
  if (mode === 'split') {
    const jobs: AnalyzeJob[] = []
    for (const chapter of chapters) {
      for (const content of chunkText(chapter.content, CHUNK_SIZE, CHUNK_OVERLAP)) {
        jobs.push({ title: chapter.title, content, chapterIds: [chapter.id] })
      }
    }
    return jobs
  }

  const merged = chapters.filter(c => c.content && c.content.trim())
  if (merged.length === 0) return []
  const title = merged.length === 1 ? merged[0].title : '全书合并分析'

  // 记录每个章节在合并正文里的字符区间，用于推算每块覆盖了哪些章节
  const parts: Array<{ chapterId: number; text: string; start: number; end: number }> = []
  let offset = 0
  for (const chapter of merged) {
    parts.push({ chapterId: chapter.id, text: chapter.content, start: offset, end: offset + chapter.content.length })
    offset += chapter.content.length
  }
  const full = parts.map(p => p.text).join('')
  const step = Math.max(1, CHUNK_SIZE - CHUNK_OVERLAP)

  const jobs: AnalyzeJob[] = []
  for (let start = 0; start < full.length; start += step) {
    const content = full.slice(start, start + CHUNK_SIZE)
    if (!content.trim()) continue
    const chunkEnd = Math.min(full.length, start + CHUNK_SIZE)
    const chapterIds = [
      ...new Set(
        parts.filter(p => p.end > start && p.start < chunkEnd).map(p => p.chapterId)
      )
    ]
    jobs.push({ title, content, chapterIds })
    if (start + CHUNK_SIZE >= full.length) break
  }
  return jobs
}

/**
 * 分析所选章节。单块失败只计数并继续；
 * 取消（AbortSignal）不再抛错丢弃成果，而是停止后续请求，
 * 把取消前已抽取的部分结果照常组装返回（stats.cancelled = true），避免算力浪费。
 * 按 mode 决定逐章（split）还是合并成一份正文一次跑完（merge）。
 */
export async function analyzeChapters(options: AnalyzeOptions): Promise<AnalyzeResult> {
  const { bookId, chapters, configId, signal, onProgress, mode = 'merge' } = options

  // 累加合并：已有实体名并入关系端点白名单，指向旧实体的关系不再被丢弃
  const baseEntities = options.baseEntities || []
  const baseNames = [
    ...new Set(baseEntities.map(e => normalizeName(e.name)).filter(Boolean))
  ]

  const accumulator: GraphAccumulator = createAccumulator()
  const chapterIdsByEntity = new Map<string, Set<number>>()
  const chapterIdsByRelation = new Map<string, Set<number>>()

  let droppedRelations = 0
  let duplicateEntities = 0
  let duplicateRelations = 0
  let failedChunks = 0
  /** 请求成功、但返回文本里解析不出图谱结构（模型没按格式输出，或 JSON 被截断） */
  let parseFailures = 0
  /** 请求成功、正文为空但思考内容非空：输出预算被思考吃光，正文没生成 */
  let emptyChunks = 0
  let cancelled = false

  const jobs = buildJobs(chapters, mode)
  const total = jobs.length
  let done = 0

  const trackChapter = (map: Map<string, Set<number>>, key: string, chapterIds: number[]) => {
    for (const chapterId of chapterIds) {
      const set = map.get(key)
      if (set) set.add(chapterId)
      else map.set(key, new Set([chapterId]))
    }
  }

  for (const job of jobs) {
    if (signal?.aborted) {
      cancelled = true
      break
    }

    const { system, user } = buildExtractionPrompt(job.title, job.content)
    let text = ''
    let reasoningChars = 0
    try {
      const streamed = await streamChat(
        {
          messages: [{ role: 'user', content: user }],
          systemPrompts: [system],
          configId
        },
        signal
      )
      text = streamed.content
      reasoningChars = streamed.reasoning.length
    } catch (error) {
      // 取消：当前块成果作废，保留之前所有块的结果并停止后续请求
      if (isAbort(error)) {
        cancelled = true
        break
      }
      failedChunks++
      done++
      onProgress?.({
        done,
        total,
        currentTitle: job.title,
        entityCount: accumulator.entities.size,
        relationCount: accumulator.relations.size
      })
      continue
    }

    // 端点解析要用上前序已抽到的实体名与已有图谱实体名，
    // 否则跨块/跨章关系、指向旧实体的关系会被当成未声明端点丢掉
    const payload = extractJsonPayload(text)
    if (!payload) {
      // 请求成功但解析不出图谱：以前这里静默算作「抽到 0 条」，界面上只看到
      // 「实体 0」而没有任何线索，查不出是截断还是模型没按格式输出。
      // 现在单独计数并区分成因，让失败可见。
      if (!text.trim()) {
        if (reasoningChars > 0) {
          // 输出预算全花在思考上，正文一个 token 都没剩 —— 典型是上限被压太小
          emptyChunks++
        } else {
          failedChunks++
        }
      } else {
        parseFailures++
      }
      done++
      onProgress?.({
        done,
        total,
        currentTitle: job.title,
        entityCount: accumulator.entities.size,
        relationCount: accumulator.relations.size
      })
      continue
    }

    const { graph, stats } = normalizeExtractedGraph(payload, [
      ...accumulatorNames(accumulator),
      ...baseNames
    ])
    droppedRelations += stats.droppedRelations

    const added = addToAccumulator(accumulator, graph)
    duplicateEntities += added.duplicateEntities
    duplicateRelations += added.duplicateRelations

    for (const entity of graph.entities) {
      trackChapter(chapterIdsByEntity, entityKey(entity.name, entity.type), job.chapterIds)
    }
    for (const relation of graph.relations) {
      trackChapter(
        chapterIdsByRelation,
        relationKey(relation.source, relation.target, relation.relation_type),
        job.chapterIds
      )
    }

    done++
    onProgress?.({
      done,
      total,
      currentTitle: job.title,
      entityCount: accumulator.entities.size,
      relationCount: accumulator.relations.size
    })
  }

  return assemble(bookId, accumulator, chapterIdsByEntity, chapterIdsByRelation, baseEntities, {
    droppedRelations,
    duplicateEntities,
    duplicateRelations,
    reusedWithBase: 0,
    failedChunks,
    parseFailures,
    emptyChunks,
    chapterCount: chapters.length,
    cancelled
  })
}

/** 把累加结果落成可导入的形状：实体拿到本地 id，关系按名称解析成 id */
function assemble(
  bookId: number,
  accumulator: GraphAccumulator,
  chapterIdsByEntity: Map<string, Set<number>>,
  chapterIdsByRelation: Map<string, Set<number>>,
  baseEntities: AnalyzeBaseEntity[],
  partial: Omit<AnalyzeStats, 'entityCount' | 'relationCount'>
): AnalyzeResult {
  const collected = accumulatorToGraph(accumulator)

  // 累加合并：base 实体名先装入映射，与旧实体同名的抽取条目复用旧 id，
  // 不生成新实体行；新实体临时 id 从 maxBaseId+1 起，与 mergeGraphData 的分配一致
  const baseIdByName = new Map<string, number>()
  const baseNameSet = new Set<string>()
  let maxBaseId = 0
  for (const base of baseEntities) {
    const key = normalizeName(base.name)
    if (!key) continue
    baseNameSet.add(key)
    if (!baseIdByName.has(key)) baseIdByName.set(key, base.id)
    maxBaseId = Math.max(maxBaseId, base.id || 0)
  }

  const nameToId = new Map(baseIdByName)
  const entities: GraphEntity[] = []
  let reusedWithBase = 0
  let nextId = maxBaseId + 1
  for (const entity of collected.entities) {
    const normalized = normalizeName(entity.name)
    if (baseNameSet.has(normalized)) {
      // 与旧实体同名：不生成新实体行，关系端点解析到旧 id
      reusedWithBase++
      continue
    }
    const id = nextId++
    if (!nameToId.has(normalized)) nameToId.set(normalized, id)
    entities.push({
      id,
      book_id: bookId,
      name: entity.name,
      // 库里允许任意 type 字符串（图例会把未知类型归入「其他」），
      // 前端的 GraphEntityType 联合只是 UI 假设，这里在边界上收敛一次
      type: entity.type as GraphEntity['type'],
      description: entity.description,
      metadata: {},
      chapter_ids: [...(chapterIdsByEntity.get(entityKey(entity.name, entity.type)) || [])],
      created_at: '',
      updated_at: ''
    })
  }

  const stats: AnalyzeStats = { ...partial, entityCount: 0, relationCount: 0 }
  const relations: GraphRelation[] = []
  for (const relation of collected.relations) {
    const sourceId = nameToId.get(normalizeName(relation.source))
    const targetId = nameToId.get(normalizeName(relation.target))
    if (!sourceId || !targetId) {
      // 同名异类型导致解析不到时兜底计数，不让它静默消失
      stats.droppedRelations++
      continue
    }
    relations.push({
      id: relations.length + 1,
      book_id: bookId,
      source_id: sourceId,
      target_id: targetId,
      relation_type: relation.relation_type,
      description: relation.description,
      evidence: relation.evidence,
      confidence: 1,
      chapter_ids: [
        ...(chapterIdsByRelation.get(
          relationKey(relation.source, relation.target, relation.relation_type)
        ) || [])
      ],
      created_at: '',
      updated_at: ''
    })
  }

  stats.entityCount = entities.length
  stats.relationCount = relations.length
  stats.reusedWithBase = reusedWithBase
  return { entities, relations, stats }
}

// ── 孤岛补边（二期）─────────────────────────────────────
export interface OrphanLinkOptions {
  bookId: number
  configId: number
  /** 当前图谱数据：度数 0 的实体作为补边对象，全部实体名作为端点候选 */
  base: { entities: GraphEntity[]; relations: GraphRelation[] }
  signal?: AbortSignal
}

export interface OrphanLinkStats {
  /** 孤立节点数 */
  orphanCount: number
  /** 提议并入的推断关系数 */
  proposed: number
  /** 解析失败被丢弃的提议数 */
  dropped: number
}

export interface OrphanLinkResult {
  /** 合并推断关系后的完整图谱数据（直接可用于保存） */
  data: { entities: GraphEntity[]; relations: GraphRelation[] }
  stats: OrphanLinkStats
}

/**
 * 为孤立节点（度数 0）提议补边：找孤岛 → 一次 AI 请求 →
 * normalizeExtractedGraph（knownNames=全部实体名）→ mergeGraphData 并入。
 * 提议关系 confidence 置 0.5、evidence 留空、relation_type 带「推断·」前缀，UI 上可识别；
 * AI 请求失败会直接抛错，由调用方静默跳过，不阻塞主流程。
 */
export async function linkOrphanEntities(options: OrphanLinkOptions): Promise<OrphanLinkResult> {
  const { bookId, base, signal } = options
  const entities = base.entities
  const empty: OrphanLinkResult = {
    data: { entities: [...entities], relations: [...base.relations] },
    stats: { orphanCount: 0, proposed: 0, dropped: 0 }
  }

  // 找孤岛：度数 0（端点不在图内的关系不计度数）
  const validIds = new Set(entities.map(e => e.id))
  const degree = new Map<number, number>()
  for (const rel of base.relations) {
    if (!validIds.has(rel.source_id) || !validIds.has(rel.target_id)) continue
    degree.set(rel.source_id, (degree.get(rel.source_id) || 0) + 1)
    degree.set(rel.target_id, (degree.get(rel.target_id) || 0) + 1)
  }
  const orphans = entities.filter(e => (degree.get(e.id) || 0) === 0)
  if (entities.length === 0 || orphans.length === 0) return empty

  const { system, user } = buildOrphanLinkPrompt(
    orphans.map(e => ({ name: e.name, type: e.type, description: e.description })),
    entities.map(e => e.name)
  )
  // 同样不传 maxTokens：补边一次要提议几十条关系，写死小值会截断 JSON，
  // 截断后 normalize 拿到的关系不完整，看起来就像「补边没效果」。
  const { content: text } = await streamChat(
    {
      messages: [{ role: 'user', content: user }],
      systemPrompts: [system],
      configId: options.configId
    },
    signal
  )

  // 端点白名单 = 全部实体名，端点解析不上的提议会被丢弃（宁缺毋滥）
  const { graph, stats } = normalizeExtractedGraph(extractJsonPayload(text), entities.map(e => e.name))

  const idByName = new Map<string, number>()
  for (const e of entities) {
    const key = normalizeName(e.name)
    if (!idByName.has(key)) idByName.set(key, e.id)
  }

  const proposed: GraphRelation[] = []
  for (const rel of graph.relations) {
    const sourceId = idByName.get(normalizeName(rel.source))
    const targetId = idByName.get(normalizeName(rel.target))
    if (!sourceId || !targetId || sourceId === targetId) continue
    proposed.push({
      // id 由 mergeGraphData 分配，只需请求内占位
      id: 0,
      book_id: bookId,
      source_id: sourceId,
      target_id: targetId,
      relation_type: `推断·${rel.relation_type}`,
      description: rel.description,
      evidence: '',
      confidence: 0.5,
      chapter_ids: [],
      created_at: '',
      updated_at: ''
    })
  }

  const { data, stats: mergeStats } = mergeGraphData(
    { entities, relations: base.relations },
    { entities: [], relations: proposed }
  )
  return {
    data,
    stats: { orphanCount: orphans.length, proposed: mergeStats.newRelations, dropped: stats.droppedRelations }
  }
}
