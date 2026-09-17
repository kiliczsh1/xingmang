/**
 * AI 增量分析结果的合并查重：旧图 ⊕ 本次抽取 = 完整新快照。
 * 零副作用纯函数：不发请求、不改 store，输入输出均为普通对象，便于独立验证。
 *
 * 去重规则（与 utils/aiGraphExtraction.ts 保持一致）：
 * - 实体按 normalizeName(name) 查重，不看 type——人名全局唯一，同名异类型以旧实体为准；
 * - 关系按 源|目标|类型 查重，方向敏感（A→B 与 B→A 视为两条）。
 *
 * 端点解析：抽取结果的 source_id/target_id 是抽取数据内的局部 id，
 * 这里统一重映射——指向旧实体的关系直接落在旧 id 上（新实体自动连到旧实体），
 * 指向新实体的关系映射到分配的临时 id（ImportGraph 的 entityIDMap 会再做 old→new 映射）。
 */
import type { GraphEntity, GraphRelation } from '@/types'
import { normalizeName, relationKey } from '@/utils/aiGraphExtraction'

export interface GraphDataSlice {
  entities: GraphEntity[]
  relations: GraphRelation[]
}

export interface MergeStats {
  /** 旧实体数 */
  baseEntities: number
  /** 旧关系数 */
  baseRelations: number
  /** 真正新增的实体数 */
  newEntities: number
  /** 同名复用旧实体的数量 */
  reusedEntities: number
  /** 真正新增的关系数 */
  newRelations: number
  /** 因查重被丢弃的新关系数 */
  duplicateRelations: number
}

export interface MergeResult {
  data: GraphDataSlice
  stats: MergeStats
}

function unionChapters(a: number[] | undefined, b: number[] | undefined): number[] {
  return [...new Set([...(a || []), ...(b || [])])]
}

function pickLongerDescription(oldDesc: string, newDesc: string): string {
  return (newDesc || '').length > (oldDesc || '').length ? newDesc : oldDesc
}

export function mergeGraphData(base: GraphDataSlice, incoming: GraphDataSlice): MergeResult {
  const stats: MergeStats = {
    baseEntities: base.entities.length,
    baseRelations: base.relations.length,
    newEntities: 0,
    reusedEntities: 0,
    newRelations: 0,
    duplicateRelations: 0
  }

  const bookId = base.entities[0]?.book_id ?? incoming.entities[0]?.book_id ?? 0

  // ── 1. 实体合并：按 normalizeName 查重（不看 type） ──────
  // 深拷贝旧实体：后续 chapter_ids/description 的合并不污染入参（零副作用）
  const entities: GraphEntity[] = base.entities.map(entity => ({ ...entity }))
  const baseByName = new Map<string, GraphEntity>()
  for (const entity of entities) {
    const key = normalizeName(entity.name)
    if (!baseByName.has(key)) baseByName.set(key, entity)
  }

  // 抽取局部实体 id -> 合并后 id；旧实体 id 集合单独记录（关系端点可能直接引用）
  const incomingIdMap = new Map<number, number>()
  const baseIds = new Set(base.entities.map(e => e.id))
  const usedNewIds = new Set<number>()
  let nextEntityId = base.entities.reduce((max, e) => Math.max(max, e.id), 0) + 1

  for (const entity of incoming.entities) {
    const key = normalizeName(entity.name)
    const old = baseByName.get(key)
    if (old) {
      // 命中旧实体：chapter_ids 并集 + description 择优，id 不变
      stats.reusedEntities++
      old.chapter_ids = unionChapters(old.chapter_ids, entity.chapter_ids)
      old.description = pickLongerDescription(old.description, entity.description || '')
      if (entity.id) incomingIdMap.set(entity.id, old.id)
      continue
    }
    // 分配新 id：优先沿用抽取侧的临时 id（已从 maxBaseId+1 起），冲突时顺序递增兜底
    let id: number
    if (entity.id && !baseIds.has(entity.id) && !usedNewIds.has(entity.id)) {
      id = entity.id
    } else {
      id = nextEntityId
      while (baseIds.has(id) || usedNewIds.has(id)) id++
      nextEntityId = id + 1
    }
    usedNewIds.add(id)
    if (entity.id) incomingIdMap.set(entity.id, id)
    entities.push({ ...entity, id, book_id: entity.book_id || bookId })
    stats.newEntities++
  }

  // ── 2. 关系合并：按 relationKey 查重（方向敏感） ─────────
  const idToName = new Map(entities.map(e => [e.id, e.name]))
  const relations: GraphRelation[] = base.relations.map(rel => ({ ...rel }))
  const existingKeys = new Set(
    base.relations.map(rel => {
      const s = idToName.get(rel.source_id) ?? String(rel.source_id)
      const t = idToName.get(rel.target_id) ?? String(rel.target_id)
      return relationKey(s, t, rel.relation_type)
    })
  )
  let nextRelationId = base.relations.reduce((max, r) => Math.max(max, r.id), 0) + 1

  for (const relation of incoming.relations) {
    // 端点解析：旧实体 id 直接用；否则查抽取局部 id 的重映射
    const sourceId = baseIds.has(relation.source_id)
      ? relation.source_id
      : incomingIdMap.get(relation.source_id) ?? null
    const targetId = baseIds.has(relation.target_id)
      ? relation.target_id
      : incomingIdMap.get(relation.target_id) ?? null
    if (!sourceId || !targetId) continue
    // 自环（同名实体映射到同一点）与抽取规则一致：丢弃不计
    if (sourceId === targetId) continue

    const key = relationKey(
      idToName.get(sourceId) ?? String(sourceId),
      idToName.get(targetId) ?? String(targetId),
      relation.relation_type
    )
    if (existingKeys.has(key)) {
      // 命中已有关系：丢弃新关系，把新章节 id 并到旧关系
      stats.duplicateRelations++
      const existing = relations.find(rel => {
        const s = idToName.get(rel.source_id) ?? String(rel.source_id)
        const t = idToName.get(rel.target_id) ?? String(rel.target_id)
        return relationKey(s, t, rel.relation_type) === key
      })
      if (existing) existing.chapter_ids = unionChapters(existing.chapter_ids, relation.chapter_ids)
      continue
    }

    relations.push({
      ...relation,
      id: nextRelationId++,
      book_id: relation.book_id || bookId,
      source_id: sourceId,
      target_id: targetId
    })
    existingKeys.add(key)
    stats.newRelations++
  }

  return { data: { entities, relations }, stats }
}
