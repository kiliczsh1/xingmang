/**
 * 知识图谱的图谱学计算：索引、邻域、筛选、同名匹配。
 * 纯函数，不碰网络与响应式状态——传输层见 api/index.ts，状态编排见 stores/knowledgeGraph.ts。
 */
import type { GraphChapterRef, GraphEntity, GraphFilters, GraphRelation } from '@/types'

/** 归一化实体名，用于同名判断与搜索匹配 */
export function normalizeLabel(label: string): string {
  return (label || '').trim().toLowerCase().replace(/\s+/g, ' ')
}

export interface GraphIndex {
  entityById: Map<number, GraphEntity>
  /** 实体 id -> 连接数（平行边各计一次） */
  degree: Map<number, number>
  /** 实体 id -> 邻居实体 id 集合 */
  adjacency: Map<number, Set<number>>
  /** 实体 id -> 相关关系（有向，含入边与出边） */
  relationsByEntity: Map<number, GraphRelation[]>
}

export function buildGraphIndex(
  entities: GraphEntity[],
  relations: GraphRelation[]
): GraphIndex {
  const entityById = new Map<number, GraphEntity>()
  const degree = new Map<number, number>()
  const adjacency = new Map<number, Set<number>>()
  const relationsByEntity = new Map<number, GraphRelation[]>()

  for (const e of entities) {
    entityById.set(e.id, e)
    degree.set(e.id, 0)
    adjacency.set(e.id, new Set())
    relationsByEntity.set(e.id, [])
  }

  for (const rel of relations) {
    // 端点不在当前图谱内时跳过，否则度数会被幽灵边抬高
    if (!entityById.has(rel.source_id) || !entityById.has(rel.target_id)) continue
    degree.set(rel.source_id, (degree.get(rel.source_id) || 0) + 1)
    degree.set(rel.target_id, (degree.get(rel.target_id) || 0) + 1)
    adjacency.get(rel.source_id)!.add(rel.target_id)
    adjacency.get(rel.target_id)!.add(rel.source_id)
    relationsByEntity.get(rel.source_id)!.push(rel)
    relationsByEntity.get(rel.target_id)!.push(rel)
  }

  return { entityById, degree, adjacency, relationsByEntity }
}

/** 无向端点键，把 A→B 与 B→A 归为同一对 */
export function pairKey(a: number, b: number): string {
  return a <= b ? `${a}|${b}` : `${b}|${a}`
}

/** 按无向端点分组，渲染时用组内序号错开曲率，避免平行边完全重叠 */
export function groupRelationsByPair(relations: GraphRelation[]): Map<string, GraphRelation[]> {
  const groups = new Map<string, GraphRelation[]>()
  for (const rel of relations) {
    const key = pairKey(rel.source_id, rel.target_id)
    const list = groups.get(key)
    if (list) list.push(rel)
    else groups.set(key, [rel])
  }
  return groups
}

/** 以 seeds 为起点 BFS，返回 depth 跳内的节点 id（含 seeds 自身） */
export function extractNeighborhood(
  adjacency: Map<number, Set<number>>,
  seeds: number[],
  depth: number,
  maxNodes = 200
): Set<number> {
  const visited = new Set<number>()
  let frontier = seeds.filter(id => adjacency.has(id))
  for (const id of frontier) visited.add(id)

  for (let d = 0; d < depth && frontier.length > 0; d++) {
    const next: number[] = []
    for (const id of frontier) {
      for (const nb of adjacency.get(id) || []) {
        if (visited.has(nb)) continue
        visited.add(nb)
        next.push(nb)
        if (visited.size >= maxNodes) return visited
      }
    }
    frontier = next
  }
  return visited
}

/**
 * 按类型筛选，并丢弃端点缺失的关系。
 * types 为空数组时只做端点清理（全量显示）。
 */
export function filterGraph(
  entities: GraphEntity[],
  relations: GraphRelation[],
  filters: GraphFilters
): { entities: GraphEntity[]; relations: GraphRelation[] } {
  const typeSet = new Set(filters.types)
  const kept = typeSet.size > 0 ? entities.filter(e => typeSet.has(e.type)) : entities
  const keptIds = new Set(kept.map(e => e.id))
  return {
    entities: kept,
    relations: relations.filter(r => keptIds.has(r.source_id) && keptIds.has(r.target_id))
  }
}

/** 按名称匹配实体，精确命中排在前面（供搜索框定位） */
export function matchEntities(entities: GraphEntity[], keyword: string): GraphEntity[] {
  const kw = normalizeLabel(keyword)
  if (!kw) return []
  const exact: GraphEntity[] = []
  const partial: GraphEntity[] = []
  for (const e of entities) {
    const name = normalizeLabel(e.name)
    if (name === kw) exact.push(e)
    else if (name.includes(kw)) partial.push(e)
  }
  return [...exact, ...partial]
}

/** 各类型实体计数，用于图例 */
export function countByType(entities: GraphEntity[]): Map<string, number> {
  const counts = new Map<string, number>()
  for (const e of entities) {
    counts.set(e.type, (counts.get(e.type) || 0) + 1)
  }
  return counts
}

export interface ChapterCoverage {
  /** 章节 -> 该章关联的实体 id */
  entityIdsByChapter: Map<number, Set<number>>
  /** 章节 -> 该章关联的关系 id */
  relationIdsByChapter: Map<number, Set<number>>
  /** 已有关联数据的章节 */
  coveredChapterIds: Set<number>
  /** 实体 -> 其所属章节（详情面板「出现章节」） */
  chapterIdsByEntity: Map<number, number[]>
  /** 关系 -> 其所属章节 */
  chapterIdsByRelation: Map<number, number[]>
}

/** 把服务端按章节分组的关联摊成双向索引：既支持按章节过滤，也支持反查实体出处 */
export function buildChapterCoverage(refs: GraphChapterRef[]): ChapterCoverage {
  const entityIdsByChapter = new Map<number, Set<number>>()
  const relationIdsByChapter = new Map<number, Set<number>>()
  const coveredChapterIds = new Set<number>()
  const chapterIdsByEntity = new Map<number, number[]>()
  const chapterIdsByRelation = new Map<number, number[]>()

  const push = (map: Map<number, number[]>, key: number, value: number) => {
    const list = map.get(key)
    if (!list) map.set(key, [value])
    else if (!list.includes(value)) list.push(value)
  }

  for (const ref of refs) {
    const chapterId = ref.chapter_id
    if (!chapterId) continue
    const entityIds = ref.entity_ids || []
    const relationIds = ref.relation_ids || []

    if (entityIds.length > 0 || relationIds.length > 0) coveredChapterIds.add(chapterId)

    if (entityIds.length > 0) {
      const set = entityIdsByChapter.get(chapterId) || new Set<number>()
      for (const id of entityIds) {
        set.add(id)
        push(chapterIdsByEntity, id, chapterId)
      }
      entityIdsByChapter.set(chapterId, set)
    }
    if (relationIds.length > 0) {
      const set = relationIdsByChapter.get(chapterId) || new Set<number>()
      for (const id of relationIds) {
        set.add(id)
        push(chapterIdsByRelation, id, chapterId)
      }
      relationIdsByChapter.set(chapterId, set)
    }
  }

  return { entityIdsByChapter, relationIdsByChapter, coveredChapterIds, chapterIdsByEntity, chapterIdsByRelation }
}

/** 某章关联的实体 + 关系总数，用于章节列表计数 */
export function chapterRefCount(coverage: ChapterCoverage, chapterId: number): number {
  return (
    (coverage.entityIdsByChapter.get(chapterId)?.size || 0) +
    (coverage.relationIdsByChapter.get(chapterId)?.size || 0)
  )
}

// ── 实体合并：主词条 / 副词条 ──────────────────────────────

export interface AliasCollapse {
  /** 主词条：画布上真正渲染的节点（副词条已剔除） */
  entities: GraphEntity[]
  /** 关系端点已重定向到主词条，并去重、丢弃自环 */
  relations: GraphRelation[]
  /** 副词条 id -> 主词条 id（顺链归到根） */
  aliasOf: Map<number, number>
  /** 主词条 id -> 它的副词条列表 */
  aliasesByPrimary: Map<number, GraphEntity[]>
}

/**
 * 把「副词条」折进它的主词条：
 *   - 副词条不单独出节点（否则图上会出现两个几乎重合的点）
 *   - 指向副词条的关系端点改指主词条；折完与已有边重合的去重，折成自环的丢弃
 *   - chapter_refs 的归并在 store 里另做（见 normalizedChapterRefs）
 *
 * 零副作用纯函数，便于独立验证。合并只改 primary_id，所以这个折叠纯属展示层的事：
 * 解除合并（primary_id 置空）后一切自动还原，不丢数据。
 */
export function collapseAliases(entities: GraphEntity[], relations: GraphRelation[]): AliasCollapse {
  const direct = new Map<number, number>()
  const byId = new Map<number, GraphEntity>()
  for (const e of entities) {
    byId.set(e.id, e)
    const pid = e.primary_id
    if (typeof pid === 'number' && pid > 0 && pid !== e.id) direct.set(e.id, pid)
  }

  // 顺链归到根，并容忍脏数据里的环
  const aliasOf = new Map<number, number>()
  for (const id of direct.keys()) {
    let cur = id
    for (let step = 0; step < 16; step++) {
      const next = direct.get(cur)
      if (next === undefined || next === cur) break
      cur = next
    }
    // 根不存在（悬空引用）就不折：宁可多显示一个节点，也不要凭空吞掉数据
    if (cur !== id && byId.has(cur)) aliasOf.set(id, cur)
  }

  const aliasesByPrimary = new Map<number, GraphEntity[]>()
  for (const [aliasId, rootId] of aliasOf) {
    const alias = byId.get(aliasId)
    if (!alias) continue
    const list = aliasesByPrimary.get(rootId)
    if (list) list.push(alias)
    else aliasesByPrimary.set(rootId, [alias])
  }

  if (aliasOf.size === 0) {
    return { entities, relations, aliasOf, aliasesByPrimary }
  }

  const collapsed = entities.filter(e => !aliasOf.has(e.id))

  const seen = new Set<string>()
  const mergedRelations: GraphRelation[] = []
  for (const rel of relations) {
    const source = aliasOf.get(rel.source_id) ?? rel.source_id
    const target = aliasOf.get(rel.target_id) ?? rel.target_id
    if (source === target) continue
    const key = `${source}|${target}|${rel.relation_type}`
    if (seen.has(key)) continue
    seen.add(key)
    mergedRelations.push(
      source === rel.source_id && target === rel.target_id
        ? rel
        : { ...rel, source_id: source, target_id: target }
    )
  }
  return { entities: collapsed, relations: mergedRelations, aliasOf, aliasesByPrimary }
}
