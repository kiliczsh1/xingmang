/**
 * 从大模型输出里抽取知识图谱三元组。
 *
 * 这个文件**刻意不 import 任何项目内模块**：纯函数 + 自带的输入输出类型，
 * 因此可以用 `npx tsc` 单独编译后直接跑 node 验证解析逻辑，不依赖测试框架
 * （项目里没有 vitest）。改这块前请保持零依赖。
 */

/** 项目支持的 7 类实体 */
export const GRAPH_ENTITY_TYPES = [
  'character',
  'location',
  'item',
  'faction',
  'event',
  'skill',
  'clue'
] as const

export interface ExtractedEntity {
  name: string
  type: string
  description: string
}

export interface ExtractedRelation {
  source: string
  target: string
  relation_type: string
  description: string
  evidence: string
}

export interface ExtractedGraph {
  entities: ExtractedEntity[]
  relations: ExtractedRelation[]
}

export interface ExtractionStats {
  entityCount: number
  relationCount: number
  /** 因为端点无法解析而丢弃的关系数 */
  droppedRelations: number
}

/** 累加器：跨章节/跨分块去重用 */
export interface GraphAccumulator {
  entities: Map<string, ExtractedEntity>
  relations: Map<string, ExtractedRelation>
}

/**
 * 归一化名称用于去重。与 services/knowledgeGraphService.ts 的 normalizeLabel 同规则，
 * 这里为了保持本模块零依赖而重复一份。
 */
export function normalizeName(name: unknown): string {
  return String(name ?? '').trim().toLowerCase().replace(/\s+/g, ' ')
}

export function entityKey(name: string, type: string): string {
  return `${normalizeName(name)}|${normalizeName(type)}`
}

export function relationKey(source: string, target: string, relationType: string): string {
  return `${normalizeName(source)}|${normalizeName(target)}|${normalizeName(relationType)}`
}

/** 中文/近义类型描述归一化到 7 类；无法识别时返回原值（图例会归入「其他」） */
export function normalizeEntityType(raw: unknown): string {
  const value = String(raw ?? '').trim()
  if (!value) return 'clue'
  const lower = value.toLowerCase()
  if ((GRAPH_ENTITY_TYPES as readonly string[]).includes(lower)) return lower

  const aliases: Record<string, string> = {
    角色: 'character',
    人物: 'character',
    主角: 'character',
    人物角色: 'character',
    地点: 'location',
    场所: 'location',
    位置: 'location',
    地域: 'location',
    物品: 'item',
    物件: 'item',
    道具: 'item',
    法宝: 'item',
    势力: 'faction',
    组织: 'faction',
    门派: 'faction',
    家族: 'faction',
    事件: 'event',
    剧情: 'event',
    功法: 'skill',
    技能: 'skill',
    武技: 'skill',
    能力: 'skill',
    线索: 'clue',
    伏笔: 'clue',
    概念: 'clue'
  }
  return aliases[value] || value
}

/** 抽取用的 system / user 提示词 */
export function buildExtractionPrompt(
  chapterTitle: string,
  chunkText: string
): { system: string; user: string } {
  const system = [
    '你在为一部中文小说构建知识图谱，需要从章节正文中抽取实体与关系。',
    '',
    '严格按以下要求输出：',
    '1. 只输出一个 JSON 对象，不要解释、不要 markdown 代码块、不要多余文字。',
    '2. 实体类型 type 只能取这几个值：character(角色) / location(地点) / item(物件) / faction(势力) / event(事件) / skill(功法) / clue(线索)。',
    '3. relations 里的 source / target 必须是实体名称，且该名称要出现在你输出的 entities 里。',
    '4. 不要重复输出同一个实体；同名同类型的只保留一条。',
    '5. description 必须是对实体／关系的生动介绍，具体、有信息量，正文能支撑多少就写多少，长度至少超过 20 字；不要空泛套话。',
    '6. relation_type（关系标签）要从下面的标签池里选，选最贴合原文的，并可以按语义替换成更符合世界观、更具体的叫法：',
    '   - 角色相关：师父/徒弟、父亲/母亲、儿子/女儿、兄长/姐姐、弟弟/妹妹、恩人、仇人、救命之恩、恩重如山、青梅竹马、恋人、伴侣、挚友、兄弟、闺蜜、死对头、对手、君臣、上下级、同僚、门徒、追随者、救命恩人、故人',
    '   - 实力与地位：强于、弱于、平手、取胜、碾压、拜师、突破、晋升、夺得、排名',
    '   - 事件与因果：导致、引发、参与、见证、肇因、为其报仇、为其而死、归因于',
    '   - 归属与关系：属于、担任、创立、统领、效忠、部下、同盟、结盟、背叛、反目、离开、加入、雇佣、雇佣关系、暗恋、仰慕、仇恨、嫉妒、信任、怀疑',
    '   关系标签要显得立体形象、贴合剧情，不要用空泛的「有关系」「相关」。',
    '7. evidence 填直接支撑该关系的原文短句（尽可能短、原话），没有就留空字符串。',
    '8. 优先抽取人物、地点、势力、关键物件、事件、功法、线索，忽略景物修辞与无关铺陈。',
    '',
    'JSON 结构（字段名必须完全一致）：',
    '{"entities":[{"name":"","type":"","description":""}],"relations":[{"source":"","target":"","relation_type":"","description":"","evidence":""}]}'
  ].join('\n')

  const user = [`章节标题：${chapterTitle || '（未命名）'}`, '章节正文：', chunkText].join('\n')

  return { system, user }
}

/** 去掉 ```json ... ``` 围栏 */
function stripCodeFence(text: string): string {
  const fenced = text.match(/```(?:json)?\s*([\s\S]*?)```/i)
  return fenced ? fenced[1] : text
}

/** 从第一个 { 开始做括号配平扫描，取最外层的完整 JSON 对象（跳过字符串内的括号） */
function sliceBalancedJson(text: string): string | null {
  const start = text.indexOf('{')
  if (start < 0) return null
  let depth = 0
  let inString = false
  let escaped = false

  for (let i = start; i < text.length; i++) {
    const ch = text[i]
    if (inString) {
      if (escaped) escaped = false
      else if (ch === '\\') escaped = true
      else if (ch === '"') inString = false
      continue
    }
    if (ch === '"') {
      inString = true
      continue
    }
    if (ch === '{') depth++
    else if (ch === '}') {
      depth--
      if (depth === 0) return text.slice(start, i + 1)
    }
  }
  return null
}

/**
 * 修掉模型常见的结构性标点错误：全角逗号/冒号/引号。
 * 只改「字符串字面量之外」的字符——字符串内部的中文标点是正文内容，不能动。
 * 顺带能救「用全角引号当 JSON 分隔符」这种写法：扫描器不把全角引号当字符串起点，
 * 于是它会被替换成半角引号，反而把结构补正了。
 */
function sanitizeStructural(text: string): string {
  let out = ''
  let inString = false
  let escaped = false

  for (const ch of text) {
    if (inString) {
      out += ch
      if (escaped) escaped = false
      else if (ch === '\\') escaped = true
      else if (ch === '"') inString = false
      continue
    }
    if (ch === '"') {
      inString = true
      out += ch
      continue
    }
    if (ch === '，') out += ','
    else if (ch === '：') out += ':'
    else if (ch === '“' || ch === '”') out += '"'
    else if (ch === '\u00a0') out += ' '
    else out += ch
  }
  return out
}

function stripTrailingCommas(text: string): string {
  return text.replace(/,(\s*[}\]])/g, '$1')
}

/** 顶层可能是数组，或把图谱塞在某个键下面，这里把 entities/relations 找出来 */
function resolveGraphShape(value: unknown): { entities: unknown[]; relations: unknown[] } | null {
  const asShape = (candidate: unknown): { entities: unknown[]; relations: unknown[] } | null => {
    if (!candidate || typeof candidate !== 'object' || Array.isArray(candidate)) return null
    const obj = candidate as Record<string, unknown>
    const entities = obj.entities ?? obj.nodes ?? obj.实体
    const relations = obj.relations ?? obj.edges ?? obj.关系
    if (!Array.isArray(entities) && !Array.isArray(relations)) return null
    return {
      entities: Array.isArray(entities) ? entities : [],
      relations: Array.isArray(relations) ? relations : []
    }
  }

  if (Array.isArray(value)) {
    const first = value.find(item => item && typeof item === 'object' && !Array.isArray(item))
    return first ? asShape(first) : null
  }
  const direct = asShape(value)
  if (direct) return direct

  // 往下一层找：{"data": {"entities": [...]}} / {"knowledge_graph": {...}}
  if (value && typeof value === 'object') {
    for (const nested of Object.values(value as Record<string, unknown>)) {
      const found = asShape(nested)
      if (found) return found
    }
  }
  return null
}

function tryParse(text: string): { entities: unknown[]; relations: unknown[] } | null {
  const attempts = [
    text,
    stripTrailingCommas(text),
    sanitizeStructural(text),
    stripTrailingCommas(sanitizeStructural(text))
  ]
  for (const attempt of attempts) {
    try {
      const shape = resolveGraphShape(JSON.parse(attempt))
      if (shape) return shape
    } catch {
      // 换下一种修法
    }
  }
  return null
}

/**
 * 从每个 `{` 位置各做一次括号配平切片。
 * 只取第一个 `{` 是不够的：模型常在 JSON 之前写解释文字，里面可能带花括号
 * （例如「我先说明 {这只是解释} 然后结果：{...}」），那样就会切错。
 */
function balancedSlices(text: string, limit = 12): string[] {
  const slices: string[] = []
  for (let i = 0; i < text.length && slices.length < limit; i++) {
    if (text[i] !== '{') continue
    const slice = sliceBalancedJson(text.slice(i))
    if (slice) slices.push(slice)
  }
  return slices
}

/**
 * 从模型返回的文本里解析出图谱结构。解析不出来返回 null。
 * 依次尝试：剥围栏后的内容 / 各段括号配平切片 / 原文，各自再做标点修正与尾逗号清理。
 */
export function extractJsonPayload(text: unknown): { entities: unknown[]; relations: unknown[] } | null {
  if (typeof text !== 'string' || !text.trim()) return null
  const cleaned = text.replace(/^\uFEFF/, '').trim()

  const candidates: string[] = []
  const fenced = stripCodeFence(cleaned)
  const fencedTrimmed = fenced.trim()

  if (fenced !== cleaned) candidates.push(fencedTrimmed)
  candidates.push(...balancedSlices(fencedTrimmed))
  if (fenced !== cleaned) candidates.push(...balancedSlices(cleaned))
  candidates.push(cleaned)

  for (const candidate of candidates) {
    const parsed = tryParse(candidate)
    if (parsed) return parsed
  }
  return null
}

/**
 * 校验并归一化解析结果。knownNames 用于解析关系端点（通常是前面章节已抽到的实体名）。
 * 端点无法解析的关系会被丢弃并计入 droppedRelations——宁可少一条边，也不要凭空造实体。
 */
export function normalizeExtractedGraph(
  payload: { entities: unknown[]; relations: unknown[] } | null,
  knownNames: string[] = []
): { graph: ExtractedGraph; stats: ExtractionStats } {
  if (!payload) {
    return {
      graph: { entities: [], relations: [] },
      stats: { entityCount: 0, relationCount: 0, droppedRelations: 0 }
    }
  }

  // 1) 实体：名称必填，类型归一化，按 名称|类型 去重
  const entityMap = new Map<string, ExtractedEntity>()
  for (const raw of payload.entities) {
    if (!raw || typeof raw !== 'object') continue
    const obj = raw as Record<string, unknown>
    const name = String(obj.name ?? obj.实体 ?? obj.label ?? '').trim()
    if (!name) continue
    const entity: ExtractedEntity = {
      name,
      type: normalizeEntityType(obj.type ?? obj.类型),
      description: String(obj.description ?? obj.描述 ?? '').trim()
    }
    const key = entityKey(entity.name, entity.type)
    if (!entityMap.has(key)) entityMap.set(key, entity)
  }

  // 2) 端点白名单 = 本次抽到的实体名 + 调用方传入的已知实体名
  const knownLowerNames = new Set<string>()
  for (const entity of entityMap.values()) knownLowerNames.add(normalizeName(entity.name))
  for (const name of knownNames) {
    const normalized = normalizeName(name)
    if (normalized) knownLowerNames.add(normalized)
  }

  // 3) 关系：端点必须在白名单内（否则丢弃并计数），按 源|目标|类型 去重
  // 计数口径：凡是没能进结果的 relation 条目都算 dropped，前端预览才能如实告知
  const relationMap = new Map<string, ExtractedRelation>()
  let droppedRelations = 0
  for (const raw of payload.relations) {
    if (!raw || typeof raw !== 'object') {
      droppedRelations++
      continue
    }
    const obj = raw as Record<string, unknown>
    const source = String(obj.source ?? obj.源 ?? obj.from ?? obj.subject ?? '').trim()
    const target = String(obj.target ?? obj.目标 ?? obj.to ?? obj.object ?? '').trim()

    const unresolvable =
      !source ||
      !target ||
      normalizeName(source) === normalizeName(target) ||
      !knownLowerNames.has(normalizeName(source)) ||
      !knownLowerNames.has(normalizeName(target))
    if (unresolvable) {
      droppedRelations++
      continue
    }

    const relationType =
      String(obj.relation_type ?? obj.relation ?? obj.关系 ?? obj.predicate ?? '').trim() || '关联'
    const key = relationKey(source, target, relationType)
    if (relationMap.has(key)) continue
    relationMap.set(key, {
      source,
      target,
      relation_type: relationType,
      description: String(obj.description ?? obj.描述 ?? '').trim(),
      evidence: String(obj.evidence ?? obj.出处 ?? obj.依据 ?? '').trim()
    })
  }

  const graph: ExtractedGraph = {
    entities: [...entityMap.values()],
    relations: [...relationMap.values()]
  }
  return {
    graph,
    stats: {
      entityCount: graph.entities.length,
      relationCount: graph.relations.length,
      droppedRelations
    }
  }
}

export function createAccumulator(): GraphAccumulator {
  return { entities: new Map(), relations: new Map() }
}

/** 把一批结果并入累加器，返回新增数量（重复的计入 duplicate 不重复添加） */
export function addToAccumulator(
  accumulator: GraphAccumulator,
  graph: ExtractedGraph
): { addedEntities: number; addedRelations: number; duplicateEntities: number; duplicateRelations: number } {
  let addedEntities = 0
  let duplicateEntities = 0
  for (const entity of graph.entities) {
    const key = entityKey(entity.name, entity.type)
    if (accumulator.entities.has(key)) {
      duplicateEntities++
      continue
    }
    accumulator.entities.set(key, entity)
    addedEntities++
  }

  let addedRelations = 0
  let duplicateRelations = 0
  for (const relation of graph.relations) {
    const key = relationKey(relation.source, relation.target, relation.relation_type)
    if (accumulator.relations.has(key)) {
      duplicateRelations++
      continue
    }
    accumulator.relations.set(key, relation)
    addedRelations++
  }

  return { addedEntities, addedRelations, duplicateEntities, duplicateRelations }
}

export function accumulatorToGraph(accumulator: GraphAccumulator): ExtractedGraph {
  return {
    entities: [...accumulator.entities.values()],
    relations: [...accumulator.relations.values()]
  }
}

/** 累加器里已有的实体名称，供下一章解析关系端点时使用 */
export function accumulatorNames(accumulator: GraphAccumulator): string[] {
  return [...accumulator.entities.values()].map(e => e.name)
}

/** 孤岛补边：度数 0 的实体，拿 name/type/description 让模型提议关系 */
export interface OrphanEntity {
  name: string
  type?: string
  description?: string
}

/** 孤岛补边提示词：输出与主抽取一致的 JSON 三元组结构，但只含 relations */
export function buildOrphanLinkPrompt(
  orphans: OrphanEntity[],
  candidates: string[]
): { system: string; user: string } {
  const system = [
    '你在为一部中文小说的知识图谱补全关系：有一批孤立实体（目前没有任何连边），需要根据它们的介绍推断出合理的关系，把图谱补得更连通。',
    '',
    '严格按以下要求输出：',
    '1. 只输出一个 JSON 对象，不要解释、不要 markdown 代码块、不要多余文字。',
    '2. 只输出 relations 字段，不要输出 entities。',
    '3. relations 的 source / target 必须逐字来自【候选实体名】列表，不要改写、不要发明新名字。',
    '4. 宁缺毋滥：只输出有把握的关系，没有把握的不要输出。',
    '5. relation_type 用简短中文标签（如 居住于、属于、师徒），不要加任何前缀（系统会统一加「推断·」）。',
    '6. evidence 固定为空字符串 ""。',
    '7. description 简要说明推断依据。',
    '',
    'JSON 结构（字段名必须完全一致）：{"relations":[{"source":"","target":"","relation_type":"","description":"","evidence":""}]}'
  ].join('\n')

  const user = [
    `候选实体名（source/target 只能从这里选）：${candidates.join('、')}`,
    '',
    '孤立实体（至少一端必须是它们）：',
    ...orphans.map(o => {
      const meta = [o.type, o.description].filter(Boolean).join('：')
      return `- ${o.name}${meta ? `（${meta}）` : ''}`
    })
  ].join('\n')

  return { system, user }
}

/** 长章节按窗口切块，带重叠，避免关系被切断在边界上 */
export function chunkText(text: string, size = 6000, overlap = 500): string[] {
  const content = String(text ?? '')
  if (content.length <= size) return content.trim() ? [content] : []
  const step = Math.max(1, size - overlap)
  const chunks: string[] = []
  for (let start = 0; start < content.length; start += step) {
    chunks.push(content.slice(start, start + size))
    if (start + size >= content.length) break
  }
  return chunks
}
