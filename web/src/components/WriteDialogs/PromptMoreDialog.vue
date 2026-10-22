<template>
  <!-- 「更多提示词」：结构/观感对齐星月写作的快捷指令选择器（标题沿用当前需求的 label） -->
  <el-dialog
    :model-value="visible"
    width="820px"
    align-center
    append-to-body
    destroy-on-close
    :show-close="false"
    :close-on-click-modal="true"
    class="xm-dialog prompt-more-dialog"
    modal-class="prompt-more-modal"
    @update:model-value="onVisibleChange"
  >
    <template #header>
      <div class="pm-head">
        <span class="pm-title">{{ title }}</span>
        <button type="button" class="pm-close" aria-label="关闭" @click="close">
          <el-icon><Close /></el-icon>
        </button>
      </div>
    </template>

    <div class="pm-body">
      <!-- 页签：已收藏 / 已创建 / 推荐 -->
      <div class="pm-tabs" role="tablist">
        <button
          v-for="t in tabs"
          :key="t.key"
          type="button"
          role="tab"
          class="pm-tab"
          :class="{ active: tab === t.key }"
          :aria-selected="tab === t.key"
          @click="switchTab(t.key)"
        >
          {{ t.label }}
        </button>
      </div>

      <!-- 搜索：胶囊输入 + 右侧「搜索」按钮 -->
      <div class="pm-search">
        <el-icon class="pm-search-icon"><Search /></el-icon>
        <input
          v-model="keywordDraft"
          class="pm-search-input"
          type="text"
          placeholder="搜索全部提示词..."
          @keydown.enter="applySearch"
        />
        <button type="button" class="pm-search-btn" @click="applySearch">搜索</button>
      </div>

      <!-- 分类筛选（可收起） -->
      <div class="pm-packs">
        <button type="button" class="pm-packs-head" :aria-expanded="packsOpen" @click="packsOpen = !packsOpen">
          <el-icon class="pm-packs-icon"><Filter /></el-icon>
          <span>分类筛选</span>
          <el-icon class="pm-packs-chevron" :class="{ open: packsOpen }"><ArrowDown /></el-icon>
        </button>
        <Transition name="pm-collapse">
          <div v-show="packsOpen" class="pm-packs-list">
            <button
              type="button"
              class="pm-chip"
              :class="{ active: activePack === ALL_PACKS }"
              @click="setPack(ALL_PACKS)"
            >全部</button>
            <button
              v-for="p in packList"
              :key="p.name"
              type="button"
              class="pm-chip"
              :class="{ active: activePack === p.name }"
              @click="setPack(p.name)"
            >{{ p.name }}</button>
          </div>
        </Transition>
      </div>

      <!-- 排序页签：最热 / 月榜 / 精选 / 最新 / 搜索 -->
      <div class="pm-sorts" role="tablist">
        <button
          v-for="s in SORTS"
          :key="s.key"
          type="button"
          role="tab"
          class="pm-sort"
          :class="{ active: sort === s.key }"
          @click="sort = s.key"
        >{{ s.label }}</button>
      </div>

      <!-- 卡片列表 -->
      <div :key="listKey" class="pm-list">
        <div v-if="list.length === 0" class="pm-empty">
          <p class="pm-empty-title">{{ emptyTitle }}</p>
          <p class="pm-empty-hint">{{ emptyHint }}</p>
        </div>
        <article
          v-for="(p, i) in pagedList"
          :key="p.id"
          class="pm-card"
          :class="{ picked: p.id === selectedId }"
          :style="{ '--i': i }"
        >
          <div class="pm-card-head">
            <h4 class="pm-card-name" :title="p.name">{{ p.name }}</h4>
            <div class="pm-card-ops">
              <button
                type="button"
                class="pm-fav"
                :class="{ on: isFav(p.id) }"
                :aria-label="isFav(p.id) ? '取消收藏' : '收藏'"
                :title="isFav(p.id) ? '取消收藏' : '收藏'"
                @click="toggleFav(p.id)"
              >
                <el-icon><StarFilled v-if="isFav(p.id)" /><Star v-else /></el-icon>
              </button>
              <button type="button" class="pm-use-btn" @click="pick(p)">
                <el-icon><CaretRight /></el-icon>使用
              </button>
              <button type="button" class="pm-del-btn" @click="removePrompt(p)">
                <el-icon><Remove /></el-icon>删除
              </button>
            </div>
          </div>

          <div class="pm-card-body">
            <div class="pm-avatar" :class="avatarTone(p)">{{ packInitial(p) }}</div>
            <div class="pm-card-info">
              <div class="pm-author">
                <span class="pm-author-name">{{ p.category || '未分类' }}</span>
                <el-icon class="pm-author-badge"><StarFilled /></el-icon>
              </div>
              <span class="pm-tag">{{ cardTag(p) }}</span>
              <div class="pm-metrics">
                <svg class="pm-fire" viewBox="0 0 24 24" aria-hidden="true">
                  <path
                    fill="currentColor"
                    d="M13.6 2c.3 3-1 4.4-2.6 6-1.9 1.9-3.9 3.8-3.9 7A6.9 6.9 0 0 0 14 22a6.9 6.9 0 0 0 6.9-7c0-5-3.3-8.3-7.3-13Zm-1 17.4A3.3 3.3 0 0 1 9.3 16c0-1.6.9-2.6 1.9-3.6.7-.7 1.5-1.4 2-2.5.9 2.1 2.4 3.5 2.4 6.1a3.3 3.3 0 0 1-3 3.4Z"
                  />
                </svg>
                <span class="pm-metric-num">{{ usageText(p) }}</span>
                <span class="pm-date">{{ dateText(p) }}</span>
              </div>
            </div>
          </div>
        </article>
      </div>

      <!-- 加载更多 -->
      <div class="pm-footer">
        <button v-if="hasMore" type="button" class="pm-more-btn" @click="page++">加载更多提示词</button>
        <span v-else class="pm-more-done">已显示全部提示词</span>
      </div>
    </div>

    <!-- 详情 -->
    <el-dialog
      v-model="detailVisible"
      :title="detailPrompt?.name || '提示词详情'"
      width="640px"
      align-center
      append-to-body
      class="xm-dialog prompt-more-detail"
    >
      <div v-if="detailPrompt" class="pm-detail">
        <div class="pm-detail-meta">
          <span class="pm-tag">{{ detailPrompt.category || '未分类' }}</span>
          <span class="pm-date">{{ charCount(detailPrompt) }} 字</span>
        </div>
        <p v-if="detailPrompt.description" class="pm-detail-desc">{{ detailPrompt.description }}</p>
        <pre class="pm-detail-content">{{ detailPrompt.content || '（这条提示词还没有正文）' }}</pre>
      </div>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
        <el-button type="primary" class="pm-use-primary" @click="useFromDetail">使用此提示词</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, CaretRight, Close, Filter, Remove, Search, Star, StarFilled } from '@element-plus/icons-vue'
import type { Prompt } from '@/types'
import { promptAPI } from '@/api'
import { bumpUsage, readFavIds, readLastPick, readUsage, writeFavIds } from '@/utils/prompt-pick-store'

defineOptions({ name: 'PromptMoreDialog' })

const props = withDefaults(
  defineProps<{
    visible: boolean
    /** 本地提示词库全量（promptAPI.getAll） */
    prompts: Prompt[]
    /** 弹窗标题：对齐星月，用当前需求的 label（如「扩写要求」） */
    title?: string
    /** 持久化作用域（工具 id / 卡包），用于记住上次选择 */
    scope?: string
    selectedId?: number | null
  }>(),
  { title: '快捷选项', scope: 'default', selectedId: null }
)

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'select', prompt: Prompt): void
  (e: 'deleted', id: number): void
}>()

type TabKey = 'fav' | 'mine' | 'recommend'
type SortKey = 'hot' | 'month' | 'featured' | 'latest' | 'search'

const ALL_PACKS = '__all__'
const PAGE_SIZE = 8

const SORTS: Array<{ key: SortKey; label: string }> = [
  { key: 'hot', label: '最热' },
  { key: 'month', label: '月榜' },
  { key: 'featured', label: '精选' },
  { key: 'latest', label: '最新' },
  { key: 'search', label: '搜索' }
]

const tab = ref<TabKey>('recommend')
const sort = ref<SortKey>('hot')
const keywordDraft = ref('')
const keyword = ref('')
const activePack = ref<string>(ALL_PACKS)
const packsOpen = ref(true)
const page = ref(1)
const favIds = ref<number[]>(readFavIds())
const usage = ref<Record<string, number>>({})
const detailVisible = ref(false)
const detailPrompt = ref<Prompt | null>(null)
/** 本地删除后即时从列表移除（父组件下次打开会重新拉取） */
const removedIds = ref<number[]>([])

const isFav = (id: number) => favIds.value.includes(id)
const visiblePrompts = computed(() => props.prompts.filter(p => !removedIds.value.includes(p.id)))

function charCount(p: Prompt): number {
  return [...(p.content || '')].length
}

function packInitial(p: Prompt): string {
  return (p.category || '未分类').trim().slice(0, 1) || '·'
}

function avatarTone(p: Prompt): string {
  const tones = ['tone-a', 'tone-b', 'tone-c', 'tone-d', 'tone-e']
  const key = p.category || p.name || ''
  let h = 0
  for (let i = 0; i < key.length; i++) h = (h * 31 + key.charCodeAt(i)) % tones.length
  return tones[h]
}

function usageText(p: Prompt): string {
  const n = usage.value[String(p.id)] || 0
  return n >= 1000 ? `${(n / 1000).toFixed(1)}K` : String(n)
}

function dateText(p: Prompt): string {
  const raw = p.updated_at || p.created_at || ''
  if (!raw) return '本地'
  return raw.slice(0, 10)
}

function cardTag(p: Prompt): string {
  if (isFav(p.id)) return '已收藏'
  if ((p.insight_count || 0) > 0) return '有心得的'
  return '本地'
}

/* 页签集合 */
const favPrompts = computed(() => visiblePrompts.value.filter(p => isFav(p.id)))

const tabs = computed(() => [
  { key: 'fav' as TabKey, label: '已收藏' },
  { key: 'mine' as TabKey, label: '已创建' },
  { key: 'recommend' as TabKey, label: '推荐' }
])

const baseList = computed<Prompt[]>(() => (tab.value === 'fav' ? favPrompts.value : visiblePrompts.value))

/* 分类筛选 */
const packList = computed(() => {
  const map = new Map<string, number>()
  for (const p of baseList.value) {
    const key = p.category || '未分类'
    map.set(key, (map.get(key) || 0) + 1)
  }
  return Array.from(map.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name, 'zh-Hans-CN'))
})

/* 搜索 + 分类 + 排序 */
const list = computed<Prompt[]>(() => {
  const kw = keyword.value.trim().toLowerCase()
  let rows = baseList.value.filter(p => {
    if (activePack.value !== ALL_PACKS && (p.category || '未分类') !== activePack.value) return false
    if (!kw) return true
    return [p.name, p.content, p.description, p.category]
      .filter(Boolean)
      .some(v => String(v).toLowerCase().includes(kw))
  })

  const score = (p: Prompt) => usage.value[String(p.id)] || 0
  const time = (p: Prompt) => {
    const raw = p.updated_at || p.created_at || ''
    const t = Date.parse(raw.replace(' ', 'T'))
    return Number.isFinite(t) ? t : 0
  }
  const monthAgo = Date.now() - 30 * 24 * 3600 * 1000

  rows = [...rows]
  if (sort.value === 'hot') rows.sort((a, b) => score(b) - score(a) || time(b) - time(a))
  else if (sort.value === 'month') rows = rows.filter(p => time(p) >= monthAgo).sort((a, b) => time(b) - time(a))
  else if (sort.value === 'featured') {
    rows.sort((a, b) => {
      const fa = isFav(a.id) ? 1 : 0
      const fb = isFav(b.id) ? 1 : 0
      if (fa !== fb) return fb - fa
      const da = a.description ? 1 : 0
      const db = b.description ? 1 : 0
      return db - da || score(b) - score(a)
    })
  } else if (sort.value === 'latest') rows.sort((a, b) => time(b) - time(a) || b.id - a.id)

  return rows
})

const pagedList = computed(() => list.value.slice(0, page.value * PAGE_SIZE))
const hasMore = computed(() => pagedList.value.length < list.value.length)
const listKey = computed(() => `${tab.value}|${sort.value}|${activePack.value}|${keyword.value}`)

const emptyTitle = computed(() => {
  if (tab.value === 'fav') return '还没有收藏的提示词'
  if (sort.value === 'month') return '本月还没有更新过的提示词'
  if (keyword.value.trim()) return '没有找到匹配的提示词'
  return '当前分类暂无提示词'
})

const emptyHint = computed(() => {
  if (tab.value === 'fav') return '点卡片右侧的星标即可收藏，收藏后在这里快速取用。'
  if (sort.value === 'month') return '切到「最新」看全部，或先在提示词库里更新一条。'
  if (keyword.value.trim()) return '换个关键词试试，或清空搜索查看全部。'
  if (activePack.value !== ALL_PACKS) return '切到「全部」分类，或去提示词库给这个卡包添加内容。'
  return '到提示词库新建提示词后，这里会自动出现。'
})

function switchTab(key: TabKey) {
  if (tab.value === key) return
  tab.value = key
  page.value = 1
}

function setPack(name: string) {
  if (activePack.value === name) return
  activePack.value = name
  page.value = 1
}

function applySearch() {
  keyword.value = keywordDraft.value.trim()
  page.value = 1
  if (keyword.value) sort.value = 'search'
}

watch(keywordDraft, v => {
  // 输入即清空搜索高亮状态，回车/点按钮才真正过滤（与星月一致）
  if (!v.trim() && keyword.value) {
    keyword.value = ''
    page.value = 1
  }
})

watch([tab, sort, activePack], () => {
  page.value = 1
})

watch(packList, packs => {
  if (activePack.value !== ALL_PACKS && !packs.some(p => p.name === activePack.value)) {
    activePack.value = ALL_PACKS
  }
})

function toggleFav(id: number) {
  const next = isFav(id) ? favIds.value.filter(x => x !== id) : [...favIds.value, id]
  favIds.value = next
  writeFavIds(next)
}

function openDetail(p: Prompt) {
  detailPrompt.value = p
  detailVisible.value = true
}

function pick(p: Prompt) {
  bumpUsage(p.id)
  usage.value = { ...usage.value }
  emit('select', p)
  emit('update:visible', false)
}

function useFromDetail() {
  const p = detailPrompt.value
  detailVisible.value = false
  if (p) pick(p)
}

async function removePrompt(p: Prompt) {
  try {
    await ElMessageBox.confirm(`删除后无法恢复，确定删除「${p.name}」？`, '删除提示词', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    const res = await promptAPI.delete(p.id)
    if (res && res.success === false) throw new Error(res.message || '删除失败')
    removedIds.value.push(p.id)
    emit('deleted', p.id)
    ElMessage.success('已删除')
  } catch (e) {
    ElMessage.error((e as Error)?.message || '删除失败')
  }
}

function close() {
  emit('update:visible', false)
}

function onVisibleChange(v: boolean) {
  emit('update:visible', v)
}

/* 打开时刷新收藏/使用统计，复位筛选 */
watch(
  () => props.visible,
  v => {
    if (!v) return
    favIds.value = readFavIds()
    usage.value = readUsage()
    removedIds.value = []
    keywordDraft.value = ''
    keyword.value = ''
    activePack.value = ALL_PACKS
    packsOpen.value = true
    page.value = 1
    tab.value = 'recommend'
    sort.value = 'hot'
    const lastId = readLastPick(props.scope)
    const last = lastId === null ? null : props.prompts.find(p => p.id === lastId)
    if (last && last.category) activePack.value = last.category
  }
)
</script>

<style scoped>
/* ============ 弹窗内部（对齐星月：单列大卡片 + 薄荷绿主色） ============ */
.pm-body {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 0;
  height: 100%;
}

.pm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.pm-title {
  font-size: 15px;
  font-weight: 500;
  color: var(--pm-ink);
  letter-spacing: 0.2px;
}
.pm-close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--pm-dim);
  cursor: pointer;
  transition: background 0.16s ease, color 0.16s ease;
}
.pm-close:hover {
  background: var(--pm-surface-2);
  color: var(--pm-ink);
}

/* ---------- 一级页签 ---------- */
.pm-tabs {
  display: flex;
  align-items: center;
  gap: 26px;
  border-bottom: 1px solid transparent;
}
.pm-tab {
  position: relative;
  padding: 6px 0 9px;
  border: none;
  background: transparent;
  color: var(--pm-dim);
  font-size: 14px;
  cursor: pointer;
  transition: color 0.18s ease;
}
.pm-tab:hover {
  color: var(--pm-ink);
}
.pm-tab.active {
  color: var(--pm-accent);
  font-weight: 500;
}
.pm-tab.active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 2px;
  border-radius: 2px;
  background: var(--pm-accent);
  animation: pm-ink-in 0.24s cubic-bezier(0.16, 1, 0.3, 1) both;
}

/* ---------- 搜索 ---------- */
.pm-search {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 40px;
  padding: 0 0 0 12px;
  border: 1px solid var(--pm-line-strong);
  border-radius: 999px;
  background: var(--pm-surface-1);
  transition: border-color 0.16s ease;
}
.pm-search:focus-within {
  border-color: var(--pm-accent);
}
.pm-search-icon {
  flex: none;
  color: var(--pm-dim);
  font-size: 15px;
}
.pm-search-input {
  flex: 1;
  min-width: 0;
  height: 100%;
  border: none;
  outline: none;
  background: transparent;
  color: var(--pm-ink);
  font-size: 13px;
}
.pm-search-input::placeholder {
  color: var(--pm-dim);
}
.pm-search-btn {
  flex: none;
  height: 38px;
  padding: 0 26px;
  margin-right: 1px;
  border: none;
  border-radius: 999px;
  background: var(--pm-accent);
  color: var(--pm-accent-ink);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: filter 0.16s ease, transform 0.16s ease;
}
.pm-search-btn:hover {
  filter: brightness(1.08);
}
.pm-search-btn:active {
  transform: scale(0.98);
}

/* ---------- 分类筛选 ---------- */
.pm-packs {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pm-packs-head {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  align-self: flex-start;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--pm-dim);
  font-size: 13px;
  cursor: pointer;
}
.pm-packs-head:hover {
  color: var(--pm-ink);
}
.pm-packs-icon {
  font-size: 13px;
}
.pm-packs-chevron {
  font-size: 12px;
  transition: transform 0.2s ease;
}
.pm-packs-chevron.open {
  transform: rotate(180deg);
}
.pm-packs-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  max-height: 108px;
  overflow-y: auto;
  padding-right: 4px;
}
.pm-packs-list::-webkit-scrollbar {
  width: 5px;
}
.pm-packs-list::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: var(--pm-line-strong);
}
.pm-chip {
  flex: none;
}
.pm-chip {
  padding: 5px 14px;
  border: 1px solid var(--pm-line);
  border-radius: 999px;
  background: var(--pm-surface-2);
  color: var(--pm-dim);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.16s ease;
}
.pm-chip:hover {
  color: var(--pm-ink);
}
.pm-chip.active {
  background: var(--pm-accent);
  border-color: var(--pm-accent);
  color: var(--pm-accent-ink);
  font-weight: 500;
}

/* 折叠动画 */
.pm-collapse-enter-active,
.pm-collapse-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.pm-collapse-enter-from,
.pm-collapse-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ---------- 排序页签 ---------- */
.pm-sorts {
  display: flex;
  align-items: center;
  gap: 26px;
  padding: 2px 0 6px;
  border-bottom: 1px solid var(--pm-line);
}
.pm-sort {
  position: relative;
  padding: 2px 0 8px;
  border: none;
  background: transparent;
  color: var(--pm-dim);
  font-size: 13px;
  cursor: pointer;
  transition: color 0.16s ease;
}
.pm-sort:hover {
  color: var(--pm-ink);
}
.pm-sort.active {
  color: var(--pm-accent);
  font-weight: 500;
}
.pm-sort.active::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: -1px;
  height: 2px;
  border-radius: 2px;
  background: var(--pm-accent);
}

/* ---------- 卡片列表（单列） ---------- */
.pm-list {
  flex: 1;
  min-height: 0;
  height: min(50vh, 430px);
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 0 4px 12px 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.pm-list::-webkit-scrollbar {
  width: 5px;
}
.pm-list::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: var(--pm-line-strong);
}
.pm-list::-webkit-scrollbar-track {
  background: transparent;
}

.pm-card {
  padding: 16px 18px;
  border: 1px solid var(--pm-line);
  border-radius: 12px;
  background: var(--pm-surface-1);
  animation: pm-card-in 0.28s cubic-bezier(0.16, 1, 0.3, 1) both;
  animation-delay: calc(var(--i, 0) * 26ms);
  transition: border-color 0.16s ease, background 0.16s ease;
}
.pm-card:hover {
  border-color: var(--pm-line-strong);
  background: var(--pm-surface-2);
}
.pm-card.picked {
  border-color: var(--pm-accent);
}
.pm-card-head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}
.pm-card-name {
  flex: 1;
  min-width: 0;
  margin: 0;
  color: var(--pm-ink);
  font-size: 15px;
  font-weight: 500;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.pm-card-ops {
  flex: none;
  display: flex;
  align-items: center;
  gap: 8px;
}
.pm-use-btn,
.pm-del-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 28px;
  padding: 0 14px;
  border-radius: 999px;
  font-size: 13px;
  cursor: pointer;
  transition: filter 0.16s ease, background 0.16s ease;
}
.pm-use-btn {
  border: none;
  background: var(--pm-accent);
  color: var(--pm-accent-ink);
}
.pm-use-btn:hover {
  filter: brightness(1.08);
}
.pm-del-btn {
  border: 1px solid var(--pm-danger-line);
  background: transparent;
  color: var(--pm-danger);
}
.pm-del-btn:hover {
  background: var(--pm-danger-soft);
}

.pm-card-body {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 14px;
}
.pm-avatar {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  color: #fff;
  font-size: 19px;
  font-weight: 600;
  user-select: none;
}
.pm-avatar.tone-a { background: linear-gradient(135deg, #f7b955, #f08c3a); }
.pm-avatar.tone-b { background: linear-gradient(135deg, #6ec6ff, #4a8ff0); }
.pm-avatar.tone-c { background: linear-gradient(135deg, #8be0b6, #46b58a); }
.pm-avatar.tone-d { background: linear-gradient(135deg, #c8a2ff, #8d6cf0); }
.pm-avatar.tone-e { background: linear-gradient(135deg, #ffa8a8, #ef7070); }

.pm-card-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.pm-author {
  display: flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
}
.pm-author-name {
  color: var(--pm-ink);
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pm-author-badge {
  flex: none;
  color: #f2b544;
  font-size: 12px;
}
.pm-tag {
  align-self: flex-start;
  padding: 1px 7px;
  border-radius: 4px;
  background: var(--pm-surface-3);
  color: var(--pm-dim);
  font-size: 11px;
  line-height: 17px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pm-metrics {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--pm-dim);
  font-size: 12px;
}
.pm-fire {
  width: 14px;
  height: 14px;
  flex: none;
  color: #f2704a;
}
.pm-metric-num {
  color: var(--pm-accent);
}
.pm-date {
  color: var(--pm-dim);
}

.pm-fav {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--pm-dim);
  cursor: pointer;
  transition: color 0.16s ease, background 0.16s ease;
}
.pm-fav:hover {
  background: var(--pm-surface-3);
  color: #f2b544;
}
.pm-fav.on {
  color: #f2b544;
}

/* ---------- 空态 ---------- */
.pm-empty {
  padding: 48px 12px;
  text-align: center;
}
.pm-empty-title {
  margin: 0 0 6px;
  color: var(--pm-ink);
  font-size: 13px;
  font-weight: 500;
}
.pm-empty-hint {
  margin: 0;
  color: var(--pm-dim);
  font-size: 12px;
  line-height: 1.6;
}

/* ---------- 底部 ---------- */
.pm-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  padding-top: 2px;
}
.pm-more-btn {
  padding: 5px 18px;
  border: 1px solid var(--pm-line);
  border-radius: 999px;
  background: transparent;
  color: var(--pm-dim);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.16s ease;
}
.pm-more-btn:hover {
  border-color: var(--pm-accent);
  color: var(--pm-accent);
}
.pm-more-done {
  color: var(--pm-dim);
  font-size: 12px;
}

/* ---------- 详情 ---------- */
.pm-detail-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.pm-detail-desc {
  margin: 0 0 10px;
  color: var(--pm-dim);
  font-size: 13px;
  line-height: 1.6;
}
.pm-detail-content {
  margin: 0;
  max-height: 46vh;
  overflow-y: auto;
  padding: 12px;
  border: 1px solid var(--pm-line);
  border-radius: 10px;
  background: var(--pm-surface-2);
  color: var(--pm-ink);
  font-family: inherit;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

@keyframes pm-card-in {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pm-ink-in {
  from {
    transform: scaleX(0.2);
    opacity: 0.4;
  }
  to {
    transform: scaleX(1);
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .pm-card,
  .pm-tab.active::after {
    animation: none;
  }
}
</style>

<style>
/*
 * 非 scoped：弹窗根节点与 el-dialog 内部元素不带本组件的 scope 属性。
 * 配色按星月实际截图取值（暗色薄荷绿），亮色下换成可读的深绿。
 */
.prompt-more-dialog {
  --pm-surface-1: #ffffff;
  --pm-surface-2: #f4f5f6;
  --pm-surface-3: #eceef0;
  --pm-line: #e6e8ea;
  --pm-line-strong: #d4d8db;
  --pm-ink: #22272b;
  --pm-dim: #8b9298;
  --pm-accent: #17a882;
  --pm-accent-ink: #ffffff;
  --pm-danger: #e5534b;
  --pm-danger-line: rgba(229, 83, 75, 0.45);
  --pm-danger-soft: rgba(229, 83, 75, 0.1);

  max-width: calc(100vw - 48px);
  border-radius: 14px;
}

.prompt-more-dialog .el-dialog__header {
  padding: 16px 20px 8px;
  border-bottom: none;
}

.prompt-more-dialog .el-dialog__body {
  display: flex;
  flex-direction: column;
  padding: 0 20px 14px;
  height: min(760px, 86dvh);
  overflow: hidden;
}

/* 暗色：对齐截图的近黑卡片 + 薄荷绿 */
:root[data-theme='dark'] .prompt-more-dialog {
  --pm-surface-1: #1f2124;
  --pm-surface-2: #26292c;
  --pm-surface-3: #2e3236;
  --pm-line: #2c3034;
  --pm-line-strong: #3a4045;
  --pm-ink: #eef0f2;
  --pm-dim: #9aa1a8;
  --pm-accent: #5fe3b4;
  --pm-accent-ink: #0f1a16;
  --pm-danger: #ff6b6b;
  --pm-danger-line: rgba(255, 107, 107, 0.42);
  --pm-danger-soft: rgba(255, 107, 107, 0.12);
}

:root[data-theme='dark'] .prompt-more-dialog .el-dialog__header {
  background: transparent;
}

/* 主按钮跟随薄荷绿 */
.prompt-more-dialog .el-button--primary,
.prompt-more-detail .el-button--primary {
  background: var(--pm-accent, #17a882);
  border-color: var(--pm-accent, #17a882);
  color: var(--pm-accent-ink, #fff);
}

/* 入场：面板缩放淡入（与 xm-pop-in 叠加时以遮罩渐显为主） */
.prompt-more-modal {
  animation: pm-mask-in 0.18s ease both;
}

@keyframes pm-mask-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}
</style>
