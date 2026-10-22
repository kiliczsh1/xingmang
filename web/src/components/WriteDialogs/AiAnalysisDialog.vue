<template>
  <el-dialog
    :model-value="visible"
    width="640px"
    top="6vh"
    append-to-body
    destroy-on-close
    class="xm-analysis-dialog xm-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <template #header>
      <div class="xm-an-header">
        <h3 class="xm-an-title">拆书</h3>
        <span class="xm-an-link" @click="historyOpen = !historyOpen">历史</span>
        <span class="xm-an-link">教程</span>
      </div>
    </template>

    <div class="xm-an-body">
      <!-- 拆书历史 -->
      <div v-if="historyOpen" class="xm-an-history">
        <div class="xm-an-history-head">
          <span>最近拆书结果（{{ history?.length ?? 0 }}）</span>
          <el-button size="small" text @click="historyOpen = false">收起</el-button>
        </div>
        <div v-if="!(history || []).length" class="xm-an-history-empty">暂无拆书历史</div>
        <div v-for="(h, i) in history || []" :key="i" class="xm-an-history-row">
          <div class="xm-an-history-info">
            <span class="xm-an-history-time">{{ h.time }}</span>
            <span class="xm-an-history-preview">{{ h.content.slice(0, 60).replace(/\s+/g, ' ') }}…</span>
          </div>
          <el-button size="small" text type="primary" @click="copyHistory(h.content)">复制</el-button>
        </div>
      </div>

      <!-- 章节选择 -->
      <div class="xm-an-field">
        <div class="xm-an-chapter-quick">
          <el-button size="small" @click="pickFirst(5)">前5章</el-button>
          <el-button size="small" @click="pickFirst(10)">前10章</el-button>
          <el-button size="small" @click="addMore(10)">增选10章</el-button>
          <el-button size="small" @click="selectedIds = []">重置</el-button>
          <span class="xm-an-cap">最多选择50个章节</span>
        </div>
        <div class="xm-an-chapter-list">
          <el-checkbox
            v-for="c in chaptersList"
            :key="c.id"
            :model-value="selectedIds.includes(c.id)"
            :disabled="!selectedIds.includes(c.id) && selectedIds.length >= 50"
            class="xm-an-chapter-item"
            @change="toggleChapter(c.id)"
          >
            <span class="xm-an-chapter-name">{{ c.title }}</span>
            <span class="xm-an-chapter-count">{{ c.word_count || 0 }}字</span>
          </el-checkbox>
          <div v-if="!chaptersList.length" class="xm-an-empty">暂无章节</div>
        </div>
      </div>

      <!-- 拆书要求 -->
      <div class="xm-an-field">
        <label class="xm-an-label"><span class="xm-an-req">*</span>拆书要求</label>
        <el-radio-group v-model="requirementMode" size="small">
          <el-radio-button value="preset">快捷选项</el-radio-button>
          <el-radio-button value="custom">自定义</el-radio-button>
        </el-radio-group>
        <el-select v-if="requirementMode === 'preset'" v-model="presetIndex" class="xm-an-preset-select">
          <el-option v-for="(p, i) in tool?.presets || []" :key="i" :label="p.name" :value="i" />
        </el-select>
        <p v-if="requirementMode === 'preset' && currentPreset?.intro" class="xm-an-preset-intro">{{ currentPreset.intro }}</p>
        <el-input
          v-if="requirementMode === 'custom'"
          v-model="customRequirement"
          type="textarea"
          :rows="3"
          placeholder="例如：拆解每章的核心剧情、节奏、爽点分布…"
        />
      </div>

      <!-- 补充信息 -->
      <div class="xm-an-field">
        <label class="xm-an-label">补充信息</label>
        <el-input
          v-model="extra"
          type="textarea"
          :rows="2"
          placeholder="可选：添加任何有助于拆书的补充信息…"
          maxlength="500"
          show-word-limit
        />
      </div>

      <!-- 模型 -->
      <div class="xm-an-field">
        <label class="xm-an-label">AI模型</label>
        <el-select v-model="modelId">
          <el-option v-for="cfg in apiConfigs" :key="cfg.id" :label="cfg.name" :value="cfg.id" />
        </el-select>
      </div>

      <!-- 拆分方式 -->
      <div class="xm-an-field">
        <div class="xm-an-split-row">
          <span class="xm-an-label">已选{{ selectedIds.length }}章{{ selectedIds.length ? '' : ',请选择章节' }}</span>
          <el-radio-group v-model="splitMode" size="small">
            <el-radio-button value="split">分章拆</el-radio-button>
            <el-radio-button value="merge">合并拆</el-radio-button>
          </el-radio-group>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="xm-an-footer">
        <span class="xm-an-hint">内容由AI生成,仅供参考和借鉴</span>
        <el-button type="warning" @click="start">{{ tool?.startLabel || '开始拆书' }}</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { ApiConfig, Chapter } from '@/types'
import type { AiToolMeta } from '@/constants/aiTools'

defineOptions({ name: 'AiAnalysisDialog' })

export interface AnalysisHistoryItem {
  time: string
  content: string
}

const props = defineProps<{
  visible: boolean
  tool: AiToolMeta | null
  chaptersList: Chapter[]
  apiConfigs: ApiConfig[]
  history?: AnalysisHistoryItem[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'start', payload: {
    requirement: string
    extra: string
    chapterIds: number[]
    splitMode: 'split' | 'merge'
  }): void
}>()

const selectedIds = ref<number[]>([])
const requirementMode = ref<'preset' | 'custom'>('preset')
const presetIndex = ref(0)
const customRequirement = ref('')
const extra = ref('')
const modelId = ref<number>()
const splitMode = ref<'split' | 'merge'>('split')
const historyOpen = ref(false)

const copyHistory = async (content: string) => {
  try {
    await navigator.clipboard.writeText(content)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const currentPreset = computed(() => props.tool?.presets?.[presetIndex.value])

watch(() => props.visible, (val) => {
  if (!val) return
  selectedIds.value = []
  requirementMode.value = 'preset'
  presetIndex.value = 0
  customRequirement.value = ''
  extra.value = ''
  splitMode.value = 'split'
  modelId.value = props.apiConfigs.find(c => c.is_default)?.id ?? props.apiConfigs[0]?.id
})

const pickFirst = (n: number) => {
  selectedIds.value = props.chaptersList.slice(0, n).map(c => c.id)
}

const addMore = (n: number) => {
  const rest = props.chaptersList.filter(c => !selectedIds.value.includes(c.id))
  selectedIds.value = [...selectedIds.value, ...rest.slice(0, n).map(c => c.id)].slice(0, 50)
}

const toggleChapter = (id: number) => {
  if (selectedIds.value.includes(id)) {
    selectedIds.value = selectedIds.value.filter(x => x !== id)
  } else if (selectedIds.value.length < 50) {
    selectedIds.value = [...selectedIds.value, id]
  }
}

const start = () => {
  const requirement = requirementMode.value === 'custom'
    ? customRequirement.value.trim()
    : currentPreset.value?.content?.trim() || props.tool?.defaultRequirement || ''
  emit('start', {
    requirement,
    extra: extra.value.trim(),
    chapterIds: selectedIds.value,
    splitMode: splitMode.value
  })
  emit('update:visible', false)
}
</script>

<style scoped>
.xm-an-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.xm-an-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--xm-ink, #1f2329);
}

.xm-an-link {
  font-size: 12px;
  color: var(--xm-accent, #4f7cf7);
  text-decoration: underline;
  cursor: pointer;
}

.xm-an-history {
  border: 1px solid var(--xm-line, #eef0f2);
  border-radius: 10px;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.xm-an-history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 12.5px;
  font-weight: 600;
  color: var(--xm-ink, #374151);
}

.xm-an-history-empty {
  font-size: 12px;
  color: var(--xm-ink-3, #a3a8b3);
  text-align: center;
  padding: 12px 0;
}

.xm-an-history-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 0;
  border-top: 1px dashed var(--xm-line, #eef0f2);
}

.xm-an-history-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.xm-an-history-time {
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
  font-variant-numeric: tabular-nums;
}

.xm-an-history-preview {
  font-size: 12px;
  color: var(--xm-ink-2, #5a6270);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-an-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.xm-an-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.xm-an-chapter-quick {
  display: flex;
  align-items: center;
  gap: 8px;
}

.xm-an-cap {
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
  margin-left: auto;
}

.xm-an-chapter-list {
  max-height: 220px;
  overflow: auto;
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: 8px;
  padding: 6px 10px;
  display: flex;
  flex-direction: column;
}

.xm-an-chapter-item {
  height: 28px;
  margin-right: 0;
}

.xm-an-chapter-name {
  font-size: 12.5px;
  color: var(--xm-ink, #374151);
}

.xm-an-chapter-count {
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
  margin-left: 6px;
}

.xm-an-empty {
  font-size: 12px;
  color: var(--xm-ink-3, #a3a8b3);
  text-align: center;
  padding: 16px 0;
}

.xm-an-label {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--xm-ink, #374151);
}

.xm-an-req {
  color: #e54d42;
  margin-right: 2px;
}

.xm-an-preset-select {
  width: 100%;
}

.xm-an-preset-intro {
  margin: 0;
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
}

.xm-an-split-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.xm-an-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.xm-an-hint {
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
}

/* 暗色适配全部交给 write-theme.css 令牌翻转（--xm-ink/--xm-ink-2/--xm-ink-3/--xm-line），原手工暗色补丁已删除 */

</style>
