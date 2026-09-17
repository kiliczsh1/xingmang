/**
 * 知识图谱的单一数据源。
 *
 * 版本语义：图书馆里有 graph_versions 时，视图只展示「当前版本」的数据；
 * 没有版本时回退到整书数据（历史数据 version_id 为空的情况，见 legacyData）。
 * WikiGraph.vue 侧栏与 KnowledgeGraph.vue 主体都只读写这里，不再互相透传状态。
 */
import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { chapterAPI, knowledgeGraphAPI } from '@/api'
import { MANUAL_EVIDENCE } from '@/utils/graphColor'
import {
  buildChapterCoverage,
  buildGraphIndex,
  collapseAliases,
  countByType,
  extractNeighborhood,
  filterGraph,
  matchEntities
} from '@/services/knowledgeGraphService'
import type {
  Chapter,
  EntityMergeCandidate,
  GraphChapterRef,
  GraphEntity,
  GraphEntityType,
  GraphFilters,
  GraphRelation,
  GraphVersion
} from '@/types'

interface GraphData {
  entities: GraphEntity[]
  relations: GraphRelation[]
}

export const useKnowledgeGraphStore = defineStore('knowledgeGraph', () => {
  const bookId = ref<number | null>(null)
  const versions = ref<GraphVersion[]>([])
  const currentVersionId = ref<number | null>(null)
  const entities = ref<GraphEntity[]>([])
  const relations = ref<GraphRelation[]>([])
  const loading = ref(false)
  /** 当前展示的是未按版本落库的历史数据，界面需给出提示 */
  const legacyData = ref(false)

  const filters = ref<GraphFilters>({ types: [] })
  const selectedEntityId = ref<number | null>(null)
  /** 邻域聚焦跳数；null 表示显示全图 */
  const focusDepth = ref<number | null>(null)
  const searchKeyword = ref('')
  /** 预留：D 阶段检索命中的节点高亮 */
  const highlightedIds = ref<number[]>([])

  /** 本书章节（懒加载，左栏章节列表与 AI 分析共用，无需父组件传 props） */
  const chapters = ref<Chapter[]>([])
  const chaptersBookId = ref<number | null>(null)
  /** 章节关联原始数据（服务端按章节分组） */
  const chapterRefs = ref<GraphChapterRef[]>([])
  /** 当前按章节过滤；null 表示不按章节过滤 */
  const selectedChapterId = ref<number | null>(null)

  // ── 派生数据 ────────────────────────────────────────────

  // 先做「实体合并」折叠：副词条不出节点，其关系端点重定向到主词条。
  // 位置刻意放在最前面 —— 后面的类型筛选 / 章节过滤 / 邻域聚焦 / 度数统计
  // 全部基于折叠后的图，否则会出现「节点没了，边还挂在已消失的 id 上」。
  // 折叠本身是 services 层的纯函数，可独立验证。
  const aliasCollapse = computed(() => collapseAliases(entities.value, relations.value))

  /** 副词条 id -> 主词条 id */
  const primaryOfAlias = computed(() => aliasCollapse.value.aliasOf)

  /** 主词条 id -> 它的副词条列表 */
  const aliasesByPrimary = computed(() => aliasCollapse.value.aliasesByPrimary)

  /** 折叠后的实体 / 关系：画布、计数、搜索都基于它们 */
  const collapsedEntities = computed(() => aliasCollapse.value.entities)
  const collapsedRelations = computed(() => aliasCollapse.value.relations)

  /** 章节关联按主词条重新归并：副词条的出处算到主词条头上 */
  const normalizedChapterRefs = computed<GraphChapterRef[]>(() => {
    const aliasMap = primaryOfAlias.value
    if (aliasMap.size === 0) return chapterRefs.value
    const groups = new Map<number, { entityIds: Set<number>; relationIds: Set<number> }>()
    for (const ref of chapterRefs.value) {
      let group = groups.get(ref.chapter_id)
      if (!group) {
        group = { entityIds: new Set(), relationIds: new Set() }
        groups.set(ref.chapter_id, group)
      }
      for (const id of ref.entity_ids || []) group.entityIds.add(aliasMap.get(id) ?? id)
      for (const id of ref.relation_ids || []) group.relationIds.add(id)
    }
    return [...groups.entries()].map(([chapterId, group]) => ({
      chapter_id: chapterId,
      entity_ids: [...group.entityIds],
      relation_ids: [...group.relationIds]
    }))
  })

  const chapterCoverage = computed(() => buildChapterCoverage(normalizedChapterRefs.value))

  /** 全量索引（不受筛选影响），详情面板的度数/关联关系用它 */
  const graphIndex = computed(() => buildGraphIndex(collapsedEntities.value, collapsedRelations.value))

  const filteredGraph = computed(() => filterGraph(collapsedEntities.value, collapsedRelations.value, filters.value))

  /**
   * 章节过滤：保留该章关联的实体与关系，并把关系的两端一并带上，
   * 否则边会被后续的端点清理逻辑丢掉，看起来像关系没关联上。
   */
  const chapterFiltered = computed<GraphData>(() => {
    const base = filteredGraph.value
    const chapterId = selectedChapterId.value
    if (chapterId === null) return base

    const coverage = chapterCoverage.value
    const entityIds = coverage.entityIdsByChapter.get(chapterId)
    const relationIds = coverage.relationIdsByChapter.get(chapterId)
    if (!entityIds && !relationIds) {
      // 该章确实没有关联数据时给空图，避免让人以为过滤没生效
      return { entities: [], relations: [] }
    }

    const relations = relationIds ? base.relations.filter(r => relationIds.has(r.id)) : []
    const keepEntityIds = new Set<number>(entityIds || [])
    for (const rel of relations) {
      keepEntityIds.add(rel.source_id)
      keepEntityIds.add(rel.target_id)
    }
    return {
      entities: base.entities.filter(e => keepEntityIds.has(e.id)),
      relations
    }
  })

  /** 最终渲染的图：类型筛选 → 章节过滤 → 邻域聚焦，并清理端点缺失的关系 */
  const visibleGraph = computed<GraphData>(() => {
    const base = chapterFiltered.value
    const seed = selectedEntityId.value
    if (focusDepth.value === null || seed === null) return base
    // 选中节点被类型筛掉时忽略聚焦，否则会出现空图
    if (!base.entities.some(e => e.id === seed)) return base
    const index = buildGraphIndex(base.entities, base.relations)
    const ids = extractNeighborhood(index.adjacency, [seed], focusDepth.value)
    return {
      entities: base.entities.filter(e => ids.has(e.id)),
      relations: base.relations.filter(r => ids.has(r.source_id) && ids.has(r.target_id))
    }
  })

  const visibleEntities = computed(() => visibleGraph.value.entities)
  const visibleRelations = computed(() => visibleGraph.value.relations)
  const visibleIndex = computed(() => buildGraphIndex(visibleEntities.value, visibleRelations.value))

  const selectedEntity = computed<GraphEntity | null>(() => {
    const id = selectedEntityId.value
    return id === null ? null : graphIndex.value.entityById.get(id) || null
  })

  const selectedRelations = computed<GraphRelation[]>(() => {
    const id = selectedEntityId.value
    return id === null ? [] : graphIndex.value.relationsByEntity.get(id) || []
  })

  /** 详情面板的关联实体列表：关系 + 对端实体 + 方向 */
  const selectedNeighbors = computed(() => {
    const id = selectedEntityId.value
    if (id === null) return []
    return selectedRelations.value.map(relation => {
      const otherId = relation.source_id === id ? relation.target_id : relation.source_id
      return {
        relation,
        entity: graphIndex.value.entityById.get(otherId) || null,
        direction: (relation.source_id === id ? 'out' : 'in') as 'out' | 'in'
      }
    })
  })

  const typeCounts = computed(() => countByType(collapsedEntities.value))

  /** 搜索：主词条名与它的副词条名都参与匹配，命中后仍返回主词条 */
  const searchMatches = computed(() => {
    const aliasGroups = aliasesByPrimary.value
    const haystack = collapsedEntities.value.map(e => {
      const aliasNames = (aliasGroups.get(e.id) || []).map(a => a.name)
      return aliasNames.length ? { ...e, name: `${e.name} ${aliasNames.join(' ')}` } : e
    })
    const hitIds = new Set(matchEntities(haystack, searchKeyword.value).map(e => e.id))
    return collapsedEntities.value.filter(e => hitIds.has(e.id))
  })
  const isEmpty = computed(() => collapsedEntities.value.length === 0)

  /** 某个主词条名下的副词条（别名）；详情面板与人工合并检测用 */
  function aliasesOf(entityId: number): GraphEntity[] {
    return aliasesByPrimary.value.get(entityId) || []
  }

  // ── 加载 ────────────────────────────────────────────────
  async function fetchVersions(id: number): Promise<GraphVersion[]> {
    const res: any = await knowledgeGraphAPI.getVersions(id)
    versions.value = res?.success && Array.isArray(res.data) ? res.data : []
    return versions.value
  }

  async function fetchUnversioned(id: number): Promise<GraphData> {
    const res: any = await knowledgeGraphAPI.getGraph(id)
    const data = res?.success ? res.data : null
    return {
      entities: (data?.entities || []) as GraphEntity[],
      relations: (data?.relations || []) as GraphRelation[]
    }
  }

  /** 本书章节列表（懒加载并缓存；换书后重新取） */
  async function loadChapters(): Promise<Chapter[]> {
    const id = bookId.value
    if (id === null) return []
    if (chaptersBookId.value === id && chapters.value.length > 0) return chapters.value
    try {
      const res: any = await chapterAPI.getByBook(id)
      chapters.value = res?.success && Array.isArray(res.data) ? res.data : []
      chaptersBookId.value = id
    } catch (e) {
      console.error('加载章节列表失败:', e)
      chapters.value = []
    }
    return chapters.value
  }

  /** 章节关联（按当前版本） */
  async function loadChapterRefs(): Promise<void> {
    const id = bookId.value
    if (id === null) {
      chapterRefs.value = []
      return
    }
    try {
      const res: any = await knowledgeGraphAPI.getChapterRefs(id, currentVersionId.value ?? undefined)
      chapterRefs.value = res?.success && Array.isArray(res.data) ? res.data : []
    } catch (e) {
      console.error('加载章节关联失败:', e)
      chapterRefs.value = []
    }
  }

  function applyData(data: GraphData) {
    entities.value = data.entities
    relations.value = data.relations
    highlightedIds.value = []
  }

  function clearFocus() {
    selectedEntityId.value = null
    focusDepth.value = null
  }

  /** 切换版本：只加载该版本的实体与关系；返回是否加载成功 */
  async function loadVersion(versionId: number): Promise<boolean> {
    const id = bookId.value
    if (id === null) return false
    loading.value = true
    let ok = true
    try {
      const res: any = await knowledgeGraphAPI.getGraph(id, versionId)
      const data = res?.success ? res.data : null
      const versionEntities: GraphEntity[] = data?.entities || []
      const versionRelations: GraphRelation[] = data?.relations || []

      // 直接以该版本的实体/关系为准。空版本就显示空图，
      // 不要再回退去拉「整书数据」——那个接口不带 versionId 会返回全书所有版本的数据，
      // 导致切到空版本时仍显示旧版本的图谱（见「切换图谱失败、总是显示以前的代码」）。
      applyData({ entities: versionEntities, relations: versionRelations })
      legacyData.value = false
      currentVersionId.value = versionId
    } catch (e) {
      console.error('加载图谱版本失败:', e)
      ok = false
      applyData({ entities: [], relations: [] })
    } finally {
      loading.value = false
      clearFocus()
      // 章节关联是按版本存的，换版本必须重新取，否则过滤结果会对不上
      selectedChapterId.value = null
      await loadChapterRefs()
    }
    return ok
  }

  /** 加载整书数据（无版本时的回退路径） */
  async function loadBook(id: number) {
    // 换书时先清空，避免请求失败后残留上一本的图谱
    if (bookId.value !== id) {
      versions.value = []
      currentVersionId.value = null
      entities.value = []
      relations.value = []
      legacyData.value = false
      highlightedIds.value = []
      chapters.value = []
      chaptersBookId.value = null
      chapterRefs.value = []
      selectedChapterId.value = null
    }
    bookId.value = id
    loading.value = true
    clearFocus()
    searchKeyword.value = ''
    try {
      // 章节列表先备好，左栏章节 tab 与 AI 分析都要用
      await loadChapters()
      const list = await fetchVersions(id)
      if (list.length > 0) {
        loading.value = false
        const target = currentVersionId.value && list.some(v => v.id === currentVersionId.value)
          ? currentVersionId.value
          : list[0].id
        await loadVersion(target)
        return
      }
      currentVersionId.value = null
      applyData(await fetchUnversioned(id))
      legacyData.value = false
      await loadChapterRefs()
    } catch (e) {
      console.error('加载知识图谱失败:', e)
      applyData({ entities: [], relations: [] })
    } finally {
      loading.value = false
    }
  }

  /** 刷新：重取版本列表 + 当前版本数据 */
  async function refresh() {
    const id = bookId.value
    if (id === null) return
    const versionId = currentVersionId.value
    await Promise.all([fetchVersions(id), loadChapters()])
    if (versionId !== null && versions.value.some(v => v.id === versionId)) {
      await loadVersion(versionId)
    } else if (versions.value.length > 0) {
      // 版本还在，就绝不能落到「整书」回退：那个接口会返回全部版本的数据，
      // 同名实体被叠在一起渲染（看着像两份图谱叠着画）。
      // 当前版本失效时（如刚删掉正在看的版本）默认挑最新的一份。
      await loadVersion(versions.value[0].id)
    } else {
      loading.value = true
      try {
        applyData(await fetchUnversioned(id))
        currentVersionId.value = null
        legacyData.value = false
        await loadChapterRefs()
      } finally {
        loading.value = false
        clearFocus()
      }
    }
  }

  // ── 版本操作 ────────────────────────────────────────────
  /** 把当前图谱存为新版本；名称由调用方提供（视图层负责弹窗） */
  async function saveAsNewVersion(name: string): Promise<{ ok: boolean; message: string }> {
    const id = bookId.value
    if (id === null) return { ok: false, message: '未选择书籍' }
    if (entities.value.length === 0) return { ok: false, message: '当前图谱为空，无需保存为版本' }

    const createRes: any = await knowledgeGraphAPI.createVersion(id, name)
    if (!createRes?.success || !createRes.data) return { ok: false, message: '创建版本失败' }
    const newVersion: GraphVersion = createRes.data

    // 章节关联是按实体/关系 id 存的，新版本会重新分配 id，
    // 所以必须把现有出处一起带上，否则复制出来的版本会丢掉章节关联。
    const coverage = chapterCoverage.value
    const importRes: any = await knowledgeGraphAPI.importGraph(id, {
      entities: entities.value.map(e => ({
        ...e,
        chapter_ids: coverage.chapterIdsByEntity.get(e.id) || []
      })),
      relations: relations.value.map(r => ({
        ...r,
        chapter_ids: coverage.chapterIdsByRelation.get(r.id) || []
      })),
      versionId: newVersion.id
    })
    if (!importRes?.success) return { ok: false, message: '写入图谱数据失败' }

    await fetchVersions(id)
    currentVersionId.value = newVersion.id
    legacyData.value = false
    await loadChapterRefs()
    return { ok: true, message: `已保存为新版本：${newVersion.name}` }
  }

  /** 把外部解析出的图谱数据导入为新版本，并切换到它 */
  async function importAsNewVersion(
    name: string,
    data: GraphData
  ): Promise<{ ok: boolean; message: string }> {
    const id = bookId.value
    if (id === null) return { ok: false, message: '未选择书籍' }

    const createRes: any = await knowledgeGraphAPI.createVersion(id, name)
    if (!createRes?.success || !createRes.data) return { ok: false, message: '创建版本失败' }
    const newVersion: GraphVersion = createRes.data

    const importRes: any = await knowledgeGraphAPI.importGraph(id, {
      entities: data.entities,
      relations: data.relations,
      versionId: newVersion.id
    })
    if (!importRes?.success) return { ok: false, message: '写入图谱数据失败' }

    await fetchVersions(id)
    currentVersionId.value = newVersion.id
    await loadVersion(newVersion.id)
    return { ok: true, message: `已导入为新版本：${newVersion.name}` }
  }

  /** 创建一个空白新版本并切换到它，从白纸开始构建（不复制当前图谱） */
  async function createBlankVersion(name: string): Promise<{ ok: boolean; message: string }> {
    const id = bookId.value
    if (id === null) return { ok: false, message: '未选择书籍' }

    const createRes: any = await knowledgeGraphAPI.createVersion(id, name)
    if (!createRes?.success || !createRes.data) return { ok: false, message: '创建版本失败' }
    const newVersion: GraphVersion = createRes.data

    await fetchVersions(id)
    currentVersionId.value = newVersion.id
    legacyData.value = false
    entities.value = []
    relations.value = []
    chapterRefs.value = []
    selectedChapterId.value = null
    clearFocus()
    return { ok: true, message: `已创建空白版本：${newVersion.name}` }
  }

  async function renameVersion(versionId: number, name: string): Promise<boolean> {
    const res: any = await knowledgeGraphAPI.renameVersion(versionId, name)
    if (!res?.success) return false
    if (bookId.value !== null) await fetchVersions(bookId.value)
    return true
  }

  async function removeVersion(versionId: number): Promise<boolean> {
    const id = bookId.value
    if (id === null) return false
    const res: any = await knowledgeGraphAPI.deleteVersion(id, versionId)
    if (!res?.success) return false
    if (currentVersionId.value === versionId) currentVersionId.value = null
    await loadBook(id)
    return true
  }

  /** 读取某版本的原始数据（导出用） */
  async function fetchVersionData(versionId: number): Promise<GraphData | null> {
    const res: any = await knowledgeGraphAPI.getVersionData(versionId)
    if (!res?.success || !res.data) return null
    return {
      entities: (res.data.entities || []) as GraphEntity[],
      relations: (res.data.relations || []) as GraphRelation[]
    }
  }

  // ── 交互 ────────────────────────────────────────────────
  function selectEntity(id: number | null) {
    selectedEntityId.value = id
    if (id === null) focusDepth.value = null
  }

  function toggleTypeFilter(type: GraphEntityType) {
    const set = new Set(filters.value.types)
    if (set.has(type)) set.delete(type)
    else set.add(type)
    filters.value = { types: [...set] }
  }

  function setTypeFilter(types: GraphEntityType[]) {
    filters.value = { types: [...types] }
  }

  function clearTypeFilter() {
    filters.value = { types: [] }
  }

  function setSearchKeyword(keyword: string) {
    searchKeyword.value = keyword
  }

  function setFocusDepth(depth: number | null) {
    focusDepth.value = depth
  }

  /** 选中/取消某章作为过滤条件；null 表示不按章节过滤 */
  function selectChapter(chapterId: number | null) {
    selectedChapterId.value = chapterId
    // 邻域聚焦与章节过滤叠加会互相抵消，切换章节时清掉聚焦
    focusDepth.value = null
    // 选中实体被章节过滤挡住时顺手清掉，避免画面与详情面板不一致
    const selected = selectedEntityId.value
    if (chapterId !== null && selected !== null) {
      if (!chapterFiltered.value.entities.some(e => e.id === selected)) {
        selectedEntityId.value = null
      }
    }
  }

  /** 实体所属章节（详情面板「出现章节」） */
  function entityChapterIds(entityId: number): number[] {
    return chapterCoverage.value.chapterIdsByEntity.get(entityId) || []
  }

  /** 关系所属章节 */
  function relationChapterIds(relationId: number): number[] {
    return chapterCoverage.value.chapterIdsByRelation.get(relationId) || []
  }

  /** 整体替换某实体关联的章节 */
  async function setEntityChapters(entityId: number, chapterIds: number[]): Promise<boolean> {
    const res: any = await knowledgeGraphAPI.setEntityChapters(entityId, chapterIds)
    if (!res?.success) return false
    await loadChapterRefs()
    return true
  }

  /** 整体替换某关系关联的章节 */
  async function setRelationChapters(relationId: number, chapterIds: number[]): Promise<boolean> {
    const res: any = await knowledgeGraphAPI.setRelationChapters(relationId, chapterIds)
    if (!res?.success) return false
    await loadChapterRefs()
    return true
  }

  // ── 实体合并（主词条 / 副词条）───────────────────────────
  /**
   * 把 memberIds 全部并入 primaryId：它们变成副词条，不再单独出节点，
   * 其关系端点由前端折到主词条上。服务端只改 primary_id，不动关系与章节关联，
   * 所以 unmergeEntities 能完全还原。
   */
  async function mergeEntities(primaryId: number, memberIds: number[]): Promise<{ ok: boolean; message: string }> {
    const id = bookId.value
    if (id === null) return { ok: false, message: '未选择书籍' }
    if (!memberIds.length) return { ok: false, message: '请先选择要并入的节点' }
    if (memberIds.includes(primaryId)) return { ok: false, message: '不能把节点并入它自己' }
    try {
      const res: any = await knowledgeGraphAPI.mergeEntities({
        bookId: id,
        versionId: currentVersionId.value ?? undefined,
        primaryId,
        memberIds
      })
      if (!res?.success) return { ok: false, message: res?.message || '合并失败' }
      await refresh()
      const merged = res.data?.merged ?? 0
      return { ok: true, message: `已把 ${merged} 个节点合并为副词条` }
    } catch (e: any) {
      return { ok: false, message: e?.message || '合并失败' }
    }
  }

  /** 解除合并：传主词条则解散整组，传副词条则只解绑它自己 */
  async function unmergeEntities(entityIds: number[]): Promise<{ ok: boolean; message: string }> {
    if (!entityIds.length) return { ok: false, message: '没有要解除的节点' }
    try {
      const res: any = await knowledgeGraphAPI.unmergeEntities(entityIds)
      if (!res?.success) return { ok: false, message: res?.message || '解除合并失败' }
      await refresh()
      return { ok: true, message: `已解除 ${res.data?.released ?? 0} 个副词条` }
    } catch (e: any) {
      return { ok: false, message: e?.message || '解除合并失败' }
    }
  }

  /** 疑似重复实体候选对（供后续的人工合并检测界面逐对确认） */
  async function fetchMergeCandidates(): Promise<EntityMergeCandidate[]> {
    const id = bookId.value
    if (id === null) return []
    try {
      const res: any = await knowledgeGraphAPI.getMergeCandidates(id, currentVersionId.value ?? undefined)
      return res?.success && Array.isArray(res.data) ? res.data : []
    } catch {
      return []
    }
  }

  /**
   * 人工连线：在当前版本（无版本则整书数据）新增一条关系。
   * evidence 固定「人工添加」，图上以绿色边标识；写完重新拉当前数据即可看到新边。
   */
  async function createManualRelation(payload: {
    sourceId: number
    targetId: number
    relationType: string
    description?: string
  }): Promise<{ ok: boolean; message: string }> {
    const id = bookId.value
    if (id === null) return { ok: false, message: '未选择书籍' }
    if (payload.sourceId === payload.targetId) return { ok: false, message: '不能连接自身' }

    try {
      const res: any = await knowledgeGraphAPI.createRelation({
        book_id: id,
        source_id: payload.sourceId,
        target_id: payload.targetId,
        relation_type: payload.relationType,
        description: payload.description || '',
        evidence: MANUAL_EVIDENCE,
        confidence: 1,
        version_id: currentVersionId.value
      })
      if (!res?.success) return { ok: false, message: res?.message || '添加关系失败' }

      // 直接改当前版本：只重取数据，不走 loadVersion（它会清空聚焦与章节过滤，
      // 连一条线就重置视图太打扰）
      const graphRes: any = await knowledgeGraphAPI.getGraph(id, currentVersionId.value ?? undefined)
      const data = graphRes?.success ? graphRes.data : null
      applyData({
        entities: (data?.entities || []) as GraphEntity[],
        relations: (data?.relations || []) as GraphRelation[]
      })
      return { ok: true, message: '已添加人工关系' }
    } catch (e: any) {
      return { ok: false, message: e?.message || '添加关系失败' }
    }
  }

  function reset() {
    bookId.value = null
    versions.value = []
    currentVersionId.value = null
    entities.value = []
    relations.value = []
    legacyData.value = false
    filters.value = { types: [] }
    selectedEntityId.value = null
    focusDepth.value = null
    searchKeyword.value = ''
    highlightedIds.value = []
    chapters.value = []
    chaptersBookId.value = null
    chapterRefs.value = []
    selectedChapterId.value = null
  }

  return {
    // state
    bookId,
    versions,
    currentVersionId,
    entities,
    relations,
    loading,
    legacyData,
    filters,
    selectedEntityId,
    focusDepth,
    searchKeyword,
    highlightedIds,
    chapters,
    chapterRefs,
    selectedChapterId,
    // derived
    graphIndex,
    chapterCoverage,
    chapterFiltered,
    visibleEntities,
    visibleRelations,
    visibleIndex,
    selectedEntity,
    selectedRelations,
    selectedNeighbors,
    typeCounts,
    searchMatches,
    isEmpty,
    // 实体合并（主词条 / 副词条）
    collapsedEntities,
    collapsedRelations,
    primaryOfAlias,
    aliasesByPrimary,
    aliasesOf,
    // actions
    loadBook,
    loadVersion,
    loadChapters,
    loadChapterRefs,
    refresh,
    saveAsNewVersion,
    importAsNewVersion,
    createBlankVersion,
    renameVersion,
    removeVersion,
    fetchVersionData,
    selectEntity,
    selectChapter,
    entityChapterIds,
    relationChapterIds,
    setEntityChapters,
    setRelationChapters,
    mergeEntities,
    unmergeEntities,
    fetchMergeCandidates,
    createManualRelation,
    toggleTypeFilter,
    setTypeFilter,
    clearTypeFilter,
    setSearchKeyword,
    setFocusDepth,
    reset
  }
})
