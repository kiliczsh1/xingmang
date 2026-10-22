<template>
  <div class="merge-panel">
    <header class="merge-header">
      <span class="merge-hint">同一实体的重复抽取可合并成一个主词条，合并可随时解除</span>
      <el-button size="small" text :loading="loading" @click="reload" title="重新查找疑似重复">
        <el-icon><Refresh /></el-icon>
      </el-button>
    </header>

    <el-alert
      v-if="errorMessage"
      class="merge-alert"
      :title="errorMessage"
      type="warning"
      :closable="false"
      show-icon
    />

    <!-- ── 疑似重复：待合并的候选对 ── -->
    <section class="merge-section">
      <div class="merge-section-head">
        <span class="merge-section-title">疑似重复</span>
        <span class="merge-count">{{ visibleCandidates.length }} 组</span>
      </div>

      <div v-if="loading" class="merge-empty">正在查找疑似重复的词条…</div>
      <div v-else-if="visibleCandidates.length === 0" class="merge-empty">
        没有发现疑似重复的词条。名称完全一致或高度相似的实体才会出现在这里。
      </div>

      <ul v-else class="candidate-list">
        <li v-for="c in visibleCandidates" :key="`${c.a.id}-${c.b.id}`" class="candidate">
          <div class="candidate-pair">
            <button class="entity-chip" @click="reveal(c.a.id)" :title="`在图谱中定位「${c.a.name}」`">
              <span class="dot" :style="{ background: typeColor(c.a.type) }"></span>
              <span class="name">{{ c.a.name }}</span>
              <span class="deg">{{ degreeOf(c.a.id) }} 连</span>
            </button>

            <span class="pair-icon">↔</span>
            <span class="score" :title="c.reasons.join(' · ')">{{ Math.round(c.score * 100) }}%</span>

            <button class="entity-chip" @click="reveal(c.b.id)" :title="`在图谱中定位「${c.b.name}」`">
              <span class="dot" :style="{ background: typeColor(c.b.type) }"></span>
              <span class="name">{{ c.b.name }}</span>
              <span class="deg">{{ degreeOf(c.b.id) }} 连</span>
            </button>
          </div>

          <div class="candidate-reasons">
            <span v-for="(r, i) in c.reasons" :key="i" class="reason">{{ r }}</span>
          </div>

          <div class="candidate-actions">
            <el-button
              size="small"
              type="primary"
              plain
              :loading="acting === `${c.a.id}-${c.b.id}`"
              @click="doMerge(c.suggest_primary_id, c)"
            >
              合并
            </el-button>
            <el-button size="small" text @click="doMerge(c.a.id, c)" title="把右侧并入左侧">
              反向
            </el-button>
            <el-button size="small" text @click="dismiss(c)">忽略</el-button>
          </div>
        </li>
      </ul>
    </section>

    <!-- ── 已合并的组：可解除 ── -->
    <section class="merge-section">
      <div class="merge-section-head">
        <span class="merge-section-title">已合并</span>
        <span class="merge-count">{{ aliasGroups.length }} 组</span>
      </div>

      <div v-if="aliasGroups.length === 0" class="merge-empty">
        还没有合并过词条。
      </div>

      <ul v-else class="group-list">
        <li v-for="g in aliasGroups" :key="g.primary.id" class="group">
          <div class="group-main">
            <span class="entity-name">{{ g.primary.name }}</span>
            <span class="group-arrow">⇠</span>
            <span class="alias-list">
              <button
                v-for="a in g.aliases"
                :key="a.id"
                class="alias-chip"
                :title="`解除「${a.name}」的合并`"
                @click="unmerge([a.id])"
              >
                {{ a.name }} ✕
              </button>
            </span>
          </div>
          <el-button size="small" text type="danger" @click="unmergeAllOf(g.primary.id)">
            整组解除
          </el-button>
        </li>
      </ul>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'
import { typeColor } from '@/utils/graphColor'
import type { EntityMergeCandidate } from '@/types'

const props = defineProps<{
  bookId: number
}>()

const emit = defineEmits<{
  /** 让父组件切回图谱并定位到该实体 */
  (e: 'reveal', id: number): void
}>()

const store = useKnowledgeGraphStore()

const loading = ref(false)
const acting = ref<string>('')
const errorMessage = ref('')
const candidates = ref<EntityMergeCandidate[]>([])

/**
 * 本地忽略集合。
 *
 * 后端没有"忽略候选"的接口 —— 这类候选是每次按名称相似度重新算出来的，
 * 记在 localStorage 里只是为了让用户不必反复看到同一组。
 * key 用「两个实体 id 升序拼接」，与顺序无关。
 */
const DISMISS_KEY = 'xingmang:merge-dismissed'
const dismissed = ref<Set<string>>(new Set())

function pairKey(a: number, b: number): string {
  return a < b ? `${a}-${b}` : `${b}-${a}`
}

function loadDismissed() {
  try {
    const raw = localStorage.getItem(DISMISS_KEY)
    dismissed.value = new Set(raw ? (JSON.parse(raw) as string[]) : [])
  } catch {
    dismissed.value = new Set()
  }
}

function saveDismissed() {
  try {
    localStorage.setItem(DISMISS_KEY, JSON.stringify([...dismissed.value]))
  } catch {
    /* 存不了就算了，不影响主流程 */
  }
}

const visibleCandidates = computed(() =>
  candidates.value.filter(c => !dismissed.value.has(pairKey(c.a.id, c.b.id)))
)

/** 已有的别名分组（主词条 → 其下副词条） */
const aliasGroups = computed(() => {
  const groups: Array<{ primary: (typeof store.collapsedEntities)[number]; aliases: ReturnType<typeof store.aliasesOf> }> = []
  for (const entity of store.collapsedEntities) {
    const aliases = store.aliasesOf(entity.id)
    if (aliases.length > 0) groups.push({ primary: entity, aliases })
  }
  // 别名多的排前面 —— 需要清理的组更值得看
  return groups.sort((x, y) => y.aliases.length - x.aliases.length)
})

function degreeOf(id: number): number {
  return store.graphIndex.degree.get(id) || 0
}

async function reload() {
  loading.value = true
  errorMessage.value = ''
  try {
    candidates.value = await store.fetchMergeCandidates()
  } catch (e) {
    errorMessage.value = e instanceof Error ? e.message : '获取候选失败'
    candidates.value = []
  } finally {
    loading.value = false
  }
}

async function doMerge(primaryId: number, c: EntityMergeCandidate) {
  const otherId = primaryId === c.a.id ? c.b.id : c.a.id
  const key = `${c.a.id}-${c.b.id}`
  acting.value = key
  try {
    const res = await store.mergeEntities(primaryId, [otherId])
    if (res.ok) {
      ElMessage.success(res.message)
      // 合并后本地列表立刻去掉这一组，不重拉（重拉会整页闪一下）
      candidates.value = candidates.value.filter(x => pairKey(x.a.id, x.b.id) !== pairKey(c.a.id, c.b.id))
    } else {
      ElMessage.error(res.message)
    }
  } finally {
    acting.value = ''
  }
}

function dismiss(c: EntityMergeCandidate) {
  dismissed.value.add(pairKey(c.a.id, c.b.id))
  saveDismissed()
}

async function unmerge(ids: number[]) {
  const res = await store.unmergeEntities(ids)
  if (res.ok) ElMessage.success(res.message)
  else ElMessage.error(res.message)
}

async function unmergeAllOf(primaryId: number) {
  const res = await store.unmergeEntities([primaryId])
  if (res.ok) ElMessage.success(res.message)
  else ElMessage.error(res.message)
}

function reveal(id: number) {
  emit('reveal', id)
}

onMounted(() => {
  loadDismissed()
  reload()
})

// 换书时重新拉候选
watch(
  () => props.bookId,
  () => {
    candidates.value = []
    reload()
  }
)
</script>

<style scoped>
.merge-panel {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 12px 12px 32px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.merge-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
}

.merge-hint {
  font-size: 11px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.merge-alert {
  margin: 0;
}

.merge-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.merge-section-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.merge-section-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.merge-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.merge-empty {
  padding: 18px 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  text-align: center;
}

/* ── 候选对 ── */
.candidate-list,
.group-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.candidate {
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.candidate-pair {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.entity-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  max-width: 100%;
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  background: transparent;
  font: inherit;
  font-size: 12px;
  color: var(--el-text-color-primary);
  cursor: pointer;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.entity-chip:hover {
  background: var(--el-fill-color-light);
  border-color: var(--el-color-primary-light-5);
}

.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.deg {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.pair-icon {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}

.score {
  font-size: 11px;
  font-variant-numeric: tabular-nums;
  color: var(--el-color-primary);
  cursor: help;
}

.candidate-reasons {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.reason {
  font-size: 11px;
  padding: 1px 7px;
  border-radius: 4px;
  background: var(--el-fill-color-light);
  color: var(--el-text-color-secondary);
}

.candidate-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

/* ── 已合并组 ── */
.group {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 8px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-bg-color);
}

.group-main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  min-width: 0;
}

.entity-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.group-arrow {
  color: var(--el-text-color-placeholder);
  font-size: 12px;
}

.alias-list {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.alias-chip {
  padding: 2px 8px;
  border: 1px dashed var(--el-border-color);
  border-radius: 999px;
  background: transparent;
  font: inherit;
  font-size: 11px;
  color: var(--el-text-color-regular);
  cursor: pointer;
}

.alias-chip:hover {
  border-color: var(--el-color-danger-light-5);
  color: var(--el-color-danger);
}

:root[data-theme='dark'] .candidate,
:root[data-theme='dark'] .group {
  background: var(--el-fill-color-blank);
}
</style>
