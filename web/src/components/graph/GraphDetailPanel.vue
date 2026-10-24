<template>
  <aside v-if="entity" class="graph-detail">
    <div class="detail-header">
      <div class="detail-title-row">
        <span class="detail-name" :title="entity.name">{{ entity.name }}</span>
      </div>      <div class="detail-meta">
        <span class="type-chip" :style="{ background: typeColor(entity.type) }">
          {{ typeLabel(entity.type) }}
        </span>
        <span class="detail-stat">连接 {{ degree }}</span>
        <span class="detail-stat">关联 {{ neighbors.length }}</span>
      </div>
    </div>

    <div class="detail-body">
      <template v-if="showInfo">
        <section v-if="entity.description" class="detail-section">
          <div class="section-label">描述</div>
          <MarkdownRenderer class="detail-desc" :content="entity.description" />
        </section>

        <section v-if="metadataEntries.length > 0" class="detail-section">
          <div class="section-label">元数据</div>
          <div class="meta-table">
            <div v-for="[key, value] in metadataEntries" :key="key" class="meta-row">
              <span class="meta-key">{{ key }}</span>
              <span class="meta-value">{{ value }}</span>
            </div>
          </div>
        </section>

        <section class="detail-section">
          <div class="section-label section-label-row">
            <span>出现章节</span>
            <el-button v-if="!editingChapters" link size="small" @click="startEditChapters">编辑</el-button>
            <span v-else class="section-actions">
              <el-button link size="small" @click="editingChapters = false">取消</el-button>
              <el-button link size="small" type="primary" :loading="savingChapters" @click="saveChapters">保存</el-button>
            </span>
          </div>

          <template v-if="!editingChapters">
            <div v-if="entityChapterLinks.length === 0" class="detail-empty">
              暂无章节关联。可用「AI 分析」抽取，或点「编辑」手工关联。
            </div>
            <div v-else class="chapter-tags">
              <button
                v-for="link in entityChapterLinks"
                :key="link.id"
                class="chapter-tag"
                :class="{ active: store.selectedChapterId === link.id }"
                :title="`只看这一章：${link.label}`"
                @click="store.selectChapter(store.selectedChapterId === link.id ? null : link.id)"
              >
                {{ link.label }}
              </button>
            </div>
          </template>

          <el-select-v2
            v-else
            v-model="draftChapterIds"
            size="small"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            :options="chapterOptions"
            placeholder="搜索并选择章节"
            style="width: 100%"
          />
        </section>
      </template>

      <section v-if="showRelations" class="detail-section">
        <div v-if="section === 'all'" class="section-label">关联实体</div>
        <div v-else class="section-label section-label-row">
          <span>关联实体</span>
          <span class="detail-stat">{{ neighbors.length }} 条</span>
        </div>
        <div v-if="neighbors.length === 0" class="detail-empty">暂无关联关系</div>
        <div v-else class="neighbor-list">
          <button
            v-for="(item, index) in neighbors"
            :key="`${item.relation.id}-${index}`"
            class="neighbor-row"
            :disabled="!item.entity"
            :title="relationTitle(item.relation)"
            @click="item.entity && emit('select', item.entity.id)"
          >
            <span class="neighbor-dir">{{ item.direction === 'out' ? '→' : '←' }}</span>
            <span class="neighbor-relation">{{ item.relation.relation_type || '关联' }}</span>
            <span class="neighbor-name">{{ item.entity?.name || '（已删除）' }}</span>
            <span class="neighbor-confidence" :title="`置信度 ${item.relation.confidence}`">
              {{ Math.round((item.relation.confidence ?? 1) * 100) }}%
            </span>
          </button>
        </div>
      </section>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import type { GraphEntity, GraphRelation } from '@/types'
import { typeColor, typeLabel } from '@/utils/graphColor'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'

export interface GraphNeighbor {
  relation: GraphRelation
  entity: GraphEntity | null
  direction: 'out' | 'in'
}

const props = defineProps<{
  entity: GraphEntity | null
  degree: number
  neighbors: GraphNeighbor[]
  /**
   * 面板拆成两个 tab 后由侧栏控制只渲染一半：
   * info = 描述/元数据/出现章节，relations = 关联实体，all = 老行为（整块）
   */
  section?: 'all' | 'info' | 'relations'
}>()

const showInfo = computed(() => props.section !== 'relations')
const showRelations = computed(() => props.section !== 'info')

const emit = defineEmits<{
  (e: 'select', id: number): void
}>()

const store = useKnowledgeGraphStore()

const editingChapters = ref(false)
const savingChapters = ref(false)
const draftChapterIds = ref<number[]>([])

function chapterLabel(chapterId: number): string {
  const index = store.chapters.findIndex(c => c.id === chapterId)
  if (index < 0) return `章节 #${chapterId}`
  const chapter = store.chapters[index]
  return chapter.title || `第${index + 1}章`
}

/** 当前实体关联的章节，点击可直接按该章过滤 */
const entityChapterLinks = computed(() => {
  const entity = props.entity
  if (!entity) return []
  return store.entityChapterIds(entity.id).map(id => ({ id, label: chapterLabel(id) }))
})

const chapterOptions = computed(() =>
  store.chapters.map((chapter, index) => ({
    label: `${String(index + 1).padStart(4, '0')} ${chapter.title || `第${index + 1}章`}`,
    value: chapter.id
  }))
)

function startEditChapters() {
  const entity = props.entity
  if (!entity) return
  draftChapterIds.value = store.entityChapterIds(entity.id)
  editingChapters.value = true
}

async function saveChapters() {
  const entity = props.entity
  if (!entity) return
  savingChapters.value = true
  try {
    const ok = await store.setEntityChapters(entity.id, draftChapterIds.value)
    if (ok) {
      ElMessage.success('已更新章节关联')
      editingChapters.value = false
    } else {
      ElMessage.error('更新章节关联失败')
    }
  } finally {
    savingChapters.value = false
  }
}

// 切换到另一个实体时退出编辑态，避免把上一章的草稿存到新实体上
watch(
  () => props.entity?.id,
  () => {
    editingChapters.value = false
    draftChapterIds.value = []
  }
)

/** metadata 已在 API 层解析为对象；只展示标量值，嵌套结构不铺开 */
const metadataEntries = computed<Array<[string, string]>>(() => {
  const meta = props.entity?.metadata
  if (!meta || typeof meta !== 'object') return []
  return Object.entries(meta)
    .filter(([, value]) => value !== null && value !== undefined && typeof value !== 'object')
    .map(([key, value]) => [key, String(value)])
})

/** 关系的描述、出处与所属章节做成悬浮提示——它们都属于单条关系，不铺在实体面板上 */
function relationTitle(relation: GraphRelation): string {
  const parts: string[] = []
  if (relation.description) parts.push(relation.description)
  if (relation.evidence) parts.push(`出处：${relation.evidence}`)
  const chapterIds = store.relationChapterIds(relation.id)
  if (chapterIds.length > 0) parts.push(`章节：${chapterIds.map(chapterLabel).join('、')}`)
  return parts.join('\n')
}
</script>

<style scoped>
.graph-detail {
  /* 现在挂在居中弹窗里：宽度跟随弹窗，不再固定 300px，也不需要左边框 */
  width: 100%;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.detail-header {
  padding: 12px 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.detail-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.detail-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.type-chip {
  padding: 1px 8px;
  border-radius: 9px;
  font-size: 11px;
  color: #fff;
}

.detail-stat {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.detail-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 14px;
}

.detail-section + .detail-section {
  margin-top: 16px;
}

.section-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.section-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.section-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.chapter-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.chapter-tag {
  max-width: 100%;
  padding: 2px 8px;
  border: 1px solid var(--el-border-color);
  border-radius: 10px;
  background: transparent;
  font-size: 11px;
  color: var(--el-text-color-regular);
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.chapter-tag:hover {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}

.chapter-tag.active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
}

.detail-desc {
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
  word-break: break-word;
}

.meta-table {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.meta-row {
  display: flex;
  gap: 8px;
  font-size: 12px;
}

.meta-key {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.meta-value {
  color: var(--el-text-color-regular);
  word-break: break-word;
}

.detail-empty {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.neighbor-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.neighbor-row {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 5px 6px;
  border: none;
  border-radius: 5px;
  background: transparent;
  font-size: 12px;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;
}

.neighbor-row:hover:not(:disabled) {
  background: var(--el-fill-color-light);
}

.neighbor-row:disabled {
  cursor: default;
  opacity: 0.6;
}

.neighbor-dir {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.neighbor-relation {
  color: var(--el-color-primary);
  flex-shrink: 0;
}

.neighbor-name {
  color: var(--el-text-color-primary);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.neighbor-confidence {
  color: var(--el-text-color-placeholder);
  font-size: 11px;
  flex-shrink: 0;
}

:root[data-theme='dark'] .graph-detail {
  border-left-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .detail-header {
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .detail-name,
:root[data-theme='dark'] .neighbor-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .neighbor-row:hover:not(:disabled) {
  background: rgba(59, 130, 246, 0.1);
}

/* EP 的 --el-* 文本/边框变量在暗色下仍是亮色主题的值，这里逐个落字面量 */
:root[data-theme='dark'] .graph-detail,
:root[data-theme='dark'] .graph-detail .el-button.is-link {
  color: #c9d3de;
}
:root[data-theme='dark'] .chapter-tag {
  border-color: rgba(42, 50, 66, 0.9);
  color: #c9d3de;
}
:root[data-theme='dark'] .chapter-tag:hover {
  border-color: rgba(59, 130, 246, 0.4);
  color: #e6edf3;
}
:root[data-theme='dark'] .detail-label,
:root[data-theme='dark'] .neighbor-meta {
  color: #8b98a8;
}
</style>
