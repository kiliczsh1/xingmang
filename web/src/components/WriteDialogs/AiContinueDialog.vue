<template>
  <el-dialog
    :model-value="visible"
    :title="tool?.title"
    width="680px"
    top="6vh"
    append-to-body
    destroy-on-close
    class="xm-continue-dialog xm-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <template #header>
      <div class="xm-cd-header">
        <h3 class="xm-cd-title">{{ tool?.title }}</h3>
        <span class="xm-cd-tutorial">教程</span>
      </div>
    </template>

    <div class="xm-cd-body">
      <!-- 续写要求 -->
      <div class="xm-cd-field">
        <label class="xm-cd-label"><span class="xm-cd-req">*</span>{{ tool?.requirementLabel }}</label>
        <el-radio-group v-model="requirementMode" size="small">
          <el-radio-button value="preset">快捷选项</el-radio-button>
          <el-radio-button value="custom">自定义</el-radio-button>
          <el-radio-button value="more">更多</el-radio-button>
        </el-radio-group>

        <template v-if="requirementMode !== 'custom'">
          <el-select v-model="presetIndex" class="xm-cd-preset-select" placeholder="请选择预设">
            <el-option
              v-for="(p, i) in activePresets"
              :key="i"
              :label="p.name"
              :value="i"
            />
          </el-select>
          <p v-if="currentPreset?.intro" class="xm-cd-preset-intro">
            {{ currentPreset.intro }}
            <span class="xm-cd-intro-link" @click="introVisible = !introVisible">{{ introVisible ? '收起介绍' : '查看介绍' }}</span>
          </p>
          <p v-if="introVisible && currentPreset" class="xm-cd-preset-detail">{{ currentPreset.content }}</p>
        </template>

        <el-input
          v-if="requirementMode === 'custom'"
          v-model="customRequirement"
          type="textarea"
          :rows="3"
          :placeholder="tool?.requirementPlaceholder"
          maxlength="5000"
        />
      </div>

      <!-- 后续剧情 -->
      <div v-if="tool?.extraField" class="xm-cd-field">
        <label class="xm-cd-label">{{ tool.extraField.label }}</label>
        <el-input
          v-model="extra"
          type="textarea"
          :rows="3"
          :placeholder="tool.extraField.placeholder"
          :maxlength="tool.extraField.maxLen"
          show-word-limit
        />
      </div>

      <!-- 提及角色 / 词条 -->
      <div class="xm-cd-field-row">
        <div class="xm-cd-field xm-cd-field--half">
          <label class="xm-cd-label">提及角色</label>
          <el-select v-model="mentionCharacters" multiple filterable placeholder="请选择角色…">
            <el-option v-for="c in availableCharacters" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </div>
        <div class="xm-cd-field xm-cd-field--half">
          <label class="xm-cd-label">提及词条</label>
          <el-select v-model="mentionTerms" multiple filterable placeholder="请选择词条…">
            <el-option v-for="t in terms" :key="t" :label="t" :value="t" />
          </el-select>
        </div>
      </div>

      <!-- 故事脉络 / 关联知识库 -->
      <div class="xm-cd-field">
        <label class="xm-cd-label">故事脉络 · 关联知识库</label>
        <div class="xm-cd-context">
          <div class="xm-cd-context-row">
            <span class="xm-cd-context-name">关联章节</span>
            <el-radio-group v-model="chapterMode" size="small">
              <el-radio-button value="none">未选择</el-radio-button>
              <el-radio-button value="recent3">最近3章</el-radio-button>
              <el-radio-button value="recent5">最近5章</el-radio-button>
              <el-radio-button value="custom">选择章节</el-radio-button>
            </el-radio-group>
          </div>
          <p class="xm-cd-context-hint">可以选择关联章节以提供上下文参考，不选择也可以直接生成内容。</p>
          <el-select
            v-if="chapterMode === 'custom'"
            v-model="customChapterIds"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            placeholder="未选择任何章节"
            class="xm-cd-context-select"
          >
            <el-option v-for="c in chaptersList" :key="c.id" :label="c.title" :value="c.id" />
          </el-select>

          <div class="xm-cd-context-row">
            <span class="xm-cd-context-name">关联备忘录</span>
            <el-select
              v-model="memoIds"
              multiple
              filterable
              collapse-tags
              placeholder="未选择任何备忘录"
              class="xm-cd-context-select"
            >
              <el-option v-for="m in memos" :key="m.id" :label="m.title" :value="m.id" />
            </el-select>
          </div>
          <p class="xm-cd-context-hint">可以选择关联备忘录中的内容以提供上下文参考，不选择也可以直接生成内容。</p>

          <div class="xm-cd-context-row">
            <span class="xm-cd-context-name">关联知识卡</span>
            <span class="xm-cd-context-empty">暂无知识卡</span>
          </div>

          <p class="xm-cd-cost">已选关联字数：{{ contextLength }}（仅预估参考，不包含输出消耗）</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="xm-cd-footer">
        <span class="xm-cd-hint">提示：高频使用可开启快捷键唤醒，秒级生成故事</span>
        <div class="xm-cd-actions">
          <el-button size="small" @click="showLast">查看上次结果</el-button>
          <el-button size="small" type="warning" @click="start">{{ tool?.startLabel || '开始生成' }}</el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { ApiConfig, Chapter, Character, Memo, Prompt } from '@/types'
import type { AiToolMeta } from '@/constants/aiTools'

defineOptions({ name: 'AiContinueDialog' })

const props = defineProps<{
  visible: boolean
  tool: AiToolMeta | null
  chaptersList: Chapter[]
  memos: Memo[]
  availableCharacters: Character[]
  terms: string[]
  prompts: Prompt[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'start', payload: {
    requirement: string
    extra: string
    contextChapterIds: number[]
    memoIds: number[]
    characterNames: string[]
    termNames: string[]
  }): void
  (e: 'show-last'): void
}>()

const requirementMode = ref<'preset' | 'custom' | 'more'>('preset')
const presetIndex = ref(0)
const customRequirement = ref('')
const introVisible = ref(false)
const extra = ref('')
const mentionCharacters = ref<number[]>([])
const mentionTerms = ref<string[]>([])
const chapterMode = ref<'none' | 'recent3' | 'recent5' | 'custom'>('none')
const customChapterIds = ref<number[]>([])
const memoIds = ref<number[]>([])

// 快捷选项 = 工具内置预设；更多 = 本地提示词库
const activePresets = computed(() => {
  if (requirementMode.value === 'more') {
    return props.prompts
      .filter(p => !!p.content)
      .slice(0, 50)
      .map(p => ({ name: p.name, content: p.content, intro: p.description || '' }))
  }
  return props.tool?.presets || []
})

const currentPreset = computed(() => activePresets.value[presetIndex.value])

const contextChapterIds = computed(() => {
  if (chapterMode.value === 'recent3') return props.chaptersList.slice(-3).map(c => c.id)
  if (chapterMode.value === 'recent5') return props.chaptersList.slice(-5).map(c => c.id)
  if (chapterMode.value === 'custom') return customChapterIds.value
  return []
})

const contextLength = computed(() => {
  let total = 0
  for (const id of contextChapterIds.value) {
    const c = props.chaptersList.find(x => x.id === id)
    if (c?.content) total += c.content.length
  }
  for (const id of memoIds.value) {
    const m = props.memos.find(x => x.id === id)
    if (m?.content) total += m.content.length
  }
  return total
})

watch(() => props.visible, (val) => {
  if (!val) return
  requirementMode.value = 'preset'
  presetIndex.value = 0
  customRequirement.value = ''
  introVisible.value = false
  extra.value = ''
  mentionCharacters.value = []
  mentionTerms.value = []
  chapterMode.value = 'none'
  customChapterIds.value = []
  memoIds.value = []
})

watch(requirementMode, () => { presetIndex.value = 0 })

const effectiveRequirement = () => {
  if (requirementMode.value === 'custom') return customRequirement.value.trim()
  return currentPreset.value?.content?.trim() || props.tool?.defaultRequirement || ''
}

const start = () => {
  emit('start', {
    requirement: effectiveRequirement(),
    extra: extra.value.trim(),
    contextChapterIds: contextChapterIds.value,
    memoIds: memoIds.value,
    characterNames: mentionCharacters.value
      .map(id => props.availableCharacters.find(c => c.id === id)?.name)
      .filter((n): n is string => !!n),
    termNames: mentionTerms.value
  })
  emit('update:visible', false)
}

const showLast = () => {
  emit('show-last')
  emit('update:visible', false)
}
</script>

<style scoped>
.xm-cd-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.xm-cd-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--xm-ink, #1f2329);
}

.xm-cd-tutorial {
  font-size: 12px;
  color: var(--xm-accent, #4f7cf7);
  text-decoration: underline;
  cursor: pointer;
}

.xm-cd-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.xm-cd-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.xm-cd-field-row {
  display: flex;
  gap: 12px;
}

.xm-cd-field--half {
  flex: 1;
}

.xm-cd-label {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--xm-ink, #374151);
}

.xm-cd-req {
  color: #e54d42;
  margin-right: 2px;
}

.xm-cd-preset-select {
  width: 100%;
}

.xm-cd-preset-intro {
  margin: 0;
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
  line-height: 1.6;
}

.xm-cd-preset-detail {
  margin: 0;
  font-size: 11px;
  color: var(--xm-ink-2, #6b7280);
  line-height: 1.7;
  background: var(--xm-surface-2, #f7f8fa);
  border-radius: 6px;
  padding: 6px 8px;
}

.xm-cd-intro-link {
  color: var(--xm-accent, #4f7cf7);
  text-decoration: underline;
  cursor: pointer;
  margin-left: 4px;
}

.xm-cd-context {
  display: flex;
  flex-direction: column;
  gap: 8px;
  background: var(--xm-surface-2, #f7f8fa);
  border-radius: 8px;
  padding: 10px 12px;
}

.xm-cd-context-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.xm-cd-context-name {
  flex-shrink: 0;
  width: 64px;
  font-size: 12.5px;
  color: var(--xm-ink, #374151);
  font-weight: 500;
}

.xm-cd-context-select {
  width: 100%;
}

.xm-cd-context-empty {
  font-size: 12px;
  color: var(--xm-ink-3, #a3a8b3);
}

.xm-cd-context-hint {
  margin: 0;
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
}

.xm-cd-cost {
  margin: 0;
  font-size: 11px;
  color: var(--xm-ink-3, #8a8f99);
}

.xm-cd-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.xm-cd-hint {
  font-size: 11px;
  color: var(--xm-ink-3, #a3a8b3);
}

.xm-cd-actions {
  display: flex;
  gap: 8px;
}

/* 暗色适配全部交给 write-theme.css 令牌翻转（--xm-ink/--xm-ink-3/--xm-surface-2），原手工暗色补丁已删除 */

</style>
