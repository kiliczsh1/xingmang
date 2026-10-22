<template>
  <div class="creative-panel">
    <div class="creative-header">
      <span class="creative-title">AI 写作</span>
      <div class="creative-header-actions">
        <el-switch
          v-model="creative2AdvancedMode"
          size="small"
          inline-prompt
          active-text="高级"
          inactive-text="简洁"
          style="--el-switch-on-color: #00c9a7; --el-switch-off-color: #909399"
        />
        <el-button
          size="small"
          @click="emit('close')"
          circle
          title="收起"
        >
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
    </div>

        <!-- 节点画布：素材 → 风格 → 生成 流水线 -->
    <div class="canvas-wrap">
      <div class="canvas" ref="canvasRef">
        <svg class="canvas-lines" ref="linesRef"></svg>
        <div class="canvas-toolbar">
          <span class="canvas-mode">{{ creative2AdvancedMode ? '复杂模式' : '简洁模式' }} · {{ canvasNodes.length }} 节点</span>
          <button class="reset-layout" @click="resetLayout" title="恢复默认布局">↺ 重置布局</button>
        </div>
        <div
          v-for="n in canvasNodes"
          :key="n.id"
          class="canvas-node"
          :class="{ done: nodeDone(n), gen: n.id === 'gen', dragging: draggingNode === n.id }"
          :data-id="n.id"
          :style="{ left: pos(n.id).x + '%', top: pos(n.id).y + 'px' }"
          @mousedown="onNodeMouseDown(n.id, $event)"
          @click="onNodeClick(n)"
        >
          <div class="node-head">
            <span class="node-dot"></span>
            <span class="node-label">{{ n.label }}</span>
            <span class="node-tag">{{ n.tag }}</span>
          </div>
          <div class="node-val" :class="{ empty: !nodeDone(n) }">{{ nodeSummary(n) }}</div>
          <span v-if="n.id !== 'gen'" class="node-edit">{{ nodeDone(n) ? '✎ 编辑' : '＋ 添加' }}</span>
        </div>
      </div>
    </div>
  </div>

  <!-- 开始生成按钮 - 固定在右下角 -->
  <div class="creative-generate-btn-wrapper">
    <el-button
      type="primary"
      :loading="creative2Generating"
      @click="handleCreative2Generate"
      class="creative-generate-btn"
    >
      <el-icon><Lightning /></el-icon>
      开始生成
    </el-button>
  </div>

  <!-- AI 写作生成结果弹窗 -->
  <el-dialog
    v-model="creative2ResultDialogVisible"
    :title="creative2WaitingForResponse ? 'AI 写作生成中...' : (creative2Generating ? 'AI 写作生成中' : 'AI 写作生成结果')"
    width="800px"
    align-center
    append-to-body
    class="creative2-result-dialog xm-dialog"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="!creative2Generating"
  >
    <div class="creative2-result-content">
      <!-- 等待AI响应时的克制状态提示 -->
      <div v-if="creative2WaitingForResponse" class="generating-animation">
        <el-icon class="generating-spinner is-loading"><Loading /></el-icon>
        <div class="generating-text">正在生成中，请稍候...</div>
      </div>

      <!-- 流式输出内容区域 -->
      <div v-else-if="creative2Generating && creative2Result" class="streaming-content">
        <div class="result-body">
          <el-scrollbar ref="streamingScrollbarRef" max-height="50vh">
            <MarkdownRenderer :content="creative2Result" />
          </el-scrollbar>
        </div>
      </div>

      <!-- 生成完成内容区域 -->
      <div v-else-if="creative2Result" class="result-complete">
        <div class="result-body">
          <el-scrollbar max-height="50vh">
            <MarkdownRenderer :content="creative2Result" />
          </el-scrollbar>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-else class="empty-result">
        <el-empty description="暂无生成内容" />
      </div>
    </div>
    <template #footer>
      <div class="creative2-result-footer">
        <el-button
          v-if="!creative2Generating"
          @click="creative2ResultDialogVisible = false"
        >
          关闭
        </el-button>
        <el-button
          v-if="creative2Generating"
          type="danger"
          @click="stopCreative2Generation"
        >
          <el-icon><CircleClose /></el-icon>
          停止生成
        </el-button>
        <el-button
          v-if="!creative2Generating"
          type="primary"
          @click="applyCreative2Result"
          :disabled="!hasChapter || !creative2Result"
        >
          <el-icon><Download /></el-icon>
          应用到当前章节
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 抽卡区提示词选择弹窗 -->
  <PromptSelectDialog
    v-model:visible="creative2PromptDialogVisible"
    v-model:model-value="creative2SelectedPrompts"
    :prompts="prompts"
    title="写作风格"
    mode="multi"
    category-filter="写作风格"
    :show-categories="false"
    :show-footer="true"
  />

  <!-- 抽卡区第二个提示词选择弹窗 -->
  <PromptSelectDialog
    v-model:visible="creative2SecondPromptDialogVisible"
    v-model:model-value="creative2SecondPromptId"
    :prompts="prompts"
    title="写作要求"
    mode="single"
    category-filter="写作要求"
    :show-categories="false"
    :show-footer="true"
  />

  <!-- 选择提示词1介绍弹窗 -->
  <el-dialog
    v-model="creative2PromptIntroDialogVisible"
    title="提示词介绍"
    width="600px"
    destroy-on-close
    class="prompt-intro-dialog xm-dialog"
    append-to-body
  >
    <div class="prompt-intro-content" v-if="creative2PromptInfo">
      <div class="prompt-intro-header">
        <h3 class="prompt-intro-name">{{ creative2PromptInfo.name }}</h3>
        <el-tag size="small" type="info">{{ creative2PromptInfo.category }}</el-tag>
      </div>
      <div class="prompt-intro-section">
        <div class="prompt-intro-label">简介</div>
        <div class="prompt-intro-description" v-html="creative2PromptInfo.description || '暂无简介'"></div>
      </div>
      <div class="prompt-intro-section">
        <div class="prompt-intro-label" @click="togglePromptContentCollapse">
          <span>提示词内容</span>
          <el-icon :class="['collapse-icon', { collapsed: isPromptContentCollapsed }]"><ArrowRight /></el-icon>
        </div>
        <el-collapse-transition>
          <div v-show="!isPromptContentCollapsed" class="prompt-intro-content-text">
            {{ creative2PromptInfo.content }}
          </div>
        </el-collapse-transition>
      </div>
    </div>
    <template #footer>
      <el-button type="primary" @click="creative2PromptIntroDialogVisible = false">确定</el-button>
    </template>
  </el-dialog>

  <!-- 选择提示词2介绍弹窗 -->
  <el-dialog
    v-model="creative2SecondPromptIntroDialogVisible"
    title="提示词介绍"
    width="600px"
    destroy-on-close
    class="prompt-intro-dialog xm-dialog"
    append-to-body
  >
    <div class="prompt-intro-content" v-if="creative2SecondPromptInfo">
      <div class="prompt-intro-header">
        <h3 class="prompt-intro-name">{{ creative2SecondPromptInfo.name }}</h3>
        <el-tag size="small" type="info">{{ creative2SecondPromptInfo.category }}</el-tag>
      </div>
      <div class="prompt-intro-section">
        <div class="prompt-intro-label">简介</div>
        <div class="prompt-intro-description" v-html="creative2SecondPromptInfo.description || '暂无简介'"></div>
      </div>
      <div class="prompt-intro-section">
        <div class="prompt-intro-label" @click="togglePromptContentCollapse">
          <span>提示词内容</span>
          <el-icon :class="['collapse-icon', { collapsed: isPromptContentCollapsed }]"><ArrowRight /></el-icon>
        </div>
        <el-collapse-transition>
          <div v-show="!isPromptContentCollapsed" class="prompt-intro-content-text">
            {{ creative2SecondPromptInfo.content }}
          </div>
        </el-collapse-transition>
      </div>
    </div>
    <template #footer>
      <el-button type="primary" @click="creative2SecondPromptIntroDialogVisible = false">确定</el-button>
    </template>
  </el-dialog>

  <!-- 模型选择弹窗 -->
  <ModelSelectDialog
    v-model:visible="modelSelectDialogVisible"
    v-model:temperature="modelTemperature"
    v-model:top-p="modelTopP"
    :api-configs="apiConfigs"
    :selected-config-id="creative2ConfigId"
    model-select-mode="creative2"
    @confirm="handleModelSelectConfirm"
  />

  <CreativeFullscreenDialog
    v-model:visible="creativeFullscreenDialogVisible"
    v-model:content="creativeFullscreenContent"
    title="编辑 - 写作内容"
    @save="saveCreativeFullscreenEditor"
    @cancel="cancelCreativeFullscreenEditor"
  />

  <!-- 关联素材弹窗（节点画布 - 关联素材节点） -->
  <el-dialog
    v-model="creative2RelateDialogVisible"
    title="关联素材"
    width="420px"
    append-to-body
    class="creative2-relate-dialog xm-dialog"
  >
    <div class="relate-dialog-body">
      <div class="relate-field">
        <div class="relate-label">关联角色卡（可选）</div>
        <el-select
          v-model="creative2SelectedCharacterIds"
          multiple
          collapse-tags
          collapse-tags-tooltip
          clearable
          filterable
          placeholder="选择角色卡作为参考"
          style="width: 100%;"
        >
          <el-option
            v-for="character in availableCharacters"
            :key="character.id"
            :label="character.name"
            :value="character.id"
          />
        </el-select>
        <div class="relate-tip">提供角色设定作为上下文参考。</div>
      </div>
      <div class="relate-field">
        <div class="relate-label">关联章节（可选）</div>
        <el-select
          v-model="creative2SelectedChapterId"
          clearable
          filterable
          placeholder="选择章节内容作为参考"
          style="width: 100%;"
        >
          <el-option
            v-for="chapter in chaptersList"
            :key="chapter.id"
            :label="chapter.title"
            :value="chapter.id"
          />
        </el-select>
        <div class="relate-tip">提供章节内容作为上下文参考。</div>
      </div>
      <div class="relate-field">
        <div class="relate-label">关联备忘录（可选）</div>
        <el-select
          v-model="creative2SelectedMemoId"
          clearable
          filterable
          placeholder="选择备忘录内容作为参考"
          style="width: 100%;"
        >
          <el-option
            v-for="memo in memos"
            :key="memo.id"
            :label="memo.title"
            :value="memo.id"
          />
        </el-select>
        <div class="relate-tip">提供备忘录内容作为上下文参考。</div>
      </div>
    </div>
    <template #footer>
      <el-button type="primary" @click="creative2RelateDialogVisible = false">完成</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import PromptSelectDialog from '@/components/WriteDialogs/PromptSelectDialog.vue'
import ModelSelectDialog from '@/components/WriteDialogs/ModelSelectDialog.vue'
import CreativeFullscreenDialog from '@/components/WriteDialogs/CreativeFullscreenDialog.vue'
import { conversationAPI } from '@/api'
import type { ApiConfig, Chapter, Character, Memo, Prompt } from '@/types'
import { saveHistoryRecord, buildHistoryMessages } from '@/composables/useWriteHistory'

const props = defineProps<{
  bookId: number
  apiConfigs: ApiConfig[]
  prompts: Prompt[]
  chaptersList: Chapter[]
  memos: Memo[]
  availableCharacters: Character[]
  hasChapter: boolean
}>()

const emit = defineEmits<{
  close: []
  'apply-result': [content: string]
}>()

// ===== 模型选择 =====
const modelSelectDialogVisible = ref(false)
const modelTemperature = ref(0.7)
const modelTopP = ref(0.9)

const handleModelSelectConfirm = (payload: { configId: number; temperature: number; topP: number; mode: 'chat' | 'creative2' }) => {
  creative2ConfigId.value = payload.configId
}

// ===== 抽卡区状态 =====
const creative2ConfigId = ref<number>()
const creative2ModelName = computed(() => {
  const config = props.apiConfigs.find(c => c.id === creative2ConfigId.value)
  return config?.name || ''
})

const openCreative2ModelSelectDialog = () => {
  // 如果当前没有选中任何模型，自动选中默认模型或第一个可用模型
  if (!creative2ConfigId.value) {
    const defaultConfig = props.apiConfigs.find(c => c.is_default)
    const firstConfig = props.apiConfigs[0]
    creative2ConfigId.value = defaultConfig?.id ?? firstConfig?.id
  }
  modelSelectDialogVisible.value = true
}

const creative2SelectedPrompts = ref<number[]>([])
const creative2Generating = ref(false)
const creative2Result = ref('')
const creative2ResultDialogVisible = ref(false)
const creative2AbortController = ref<AbortController | null>(null)
const creative2WaitingForResponse = ref(false)
const streamingScrollbarRef = ref<any>(null)
const creative2AdvancedMode = ref(false)
const creative2PromptDialogVisible = ref(false)
const creative2SecondPromptDialogVisible = ref(false)
const creative2RelateDialogVisible = ref(false)
const creative2SecondPromptId = ref<number>(0)
const creative2UseFixedPrompt = ref(false)
const creative2FixedPromptId = ref<number>(0)
const creative2StoryBackground = ref('')
const creative2CharacterRelations = ref('')
const creative2ChapterPlot = ref('')
const creative2AdditionalInfo = ref('')
const creative2SelectedChapterId = ref<number>()
const creative2SelectedMemoId = ref<number>()
const creative2SelectedCharacterIds = ref<number[]>([])

// ===== 抽卡区字段配置 =====
const creative2FieldValues = ref<Record<string, string>>({})

const creative2AllFields = computed(() => {
  const fields: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }> = []

  const promptsToUse = creative2UseFixedPrompt.value && creative2FixedPromptId.value !== 0
    ? [creative2FixedPromptId.value]
    : creative2SelectedPrompts.value

  promptsToUse.forEach(promptId => {
    const prompt = props.prompts.find(p => p.id === promptId)
    if (prompt && prompt.fields && Array.isArray(prompt.fields)) {
      prompt.fields.forEach((field: any) => {
        if (!fields.find(f => f.name === field.name)) {
          fields.push({
            name: field.name,
            label: field.label,
            type: field.type || 'text',
            options: field.options || [],
            optionLabels: field.optionLabels || [],
            description: field.description || '',
            required: field.required !== false
          })
        }
      })
    }
  })

  return fields
})

// ===== 节点画布 =====
interface CanvasNode {
  id: string
  label: string
  tag: string
  type: 'model' | 'text' | 'style' | 'relate' | 'gen'
  advanced: boolean
}

const canvasRef = ref<HTMLElement | null>(null)
const linesRef = ref<SVGSVGElement | null>(null)
const draggingNode = ref<string | null>(null)

const canvasNodeDefs: CanvasNode[] = [
  { id: 'model',  label: 'AI 模型',  tag: '必选', type: 'model',  advanced: false },
  { id: 'style',  label: '写作风格', tag: '必选', type: 'style',  advanced: false },
  { id: 'gen',    label: '生成本章', tag: '输出', type: 'gen',    advanced: false },
  { id: 'bg',     label: '故事背景', tag: '可选', type: 'text',   advanced: true },
  { id: 'rel',    label: '角色关系', tag: '可选', type: 'text',   advanced: true },
  { id: 'plot',   label: '本章剧情', tag: '可选', type: 'text',   advanced: true },
  { id: 'relate', label: '关联素材', tag: '可选', type: 'relate', advanced: true },
  { id: 'req',    label: '写作要求', tag: '可选', type: 'style',  advanced: true },
  { id: 'extra',  label: '补充信息', tag: '可选', type: 'text',   advanced: true },
]

const canvasNodes = computed(() => canvasNodeDefs.filter(n => !n.advanced || creative2AdvancedMode.value))

// 默认布局：x 为百分比（适配 260-760px 可拖拽宽度），y 为像素
const DEFAULT_POS: Record<string, Record<string, { x: number; y: number }>> = {
  simple: {
    model: { x: 5, y: 18 },
    style: { x: 55, y: 18 },
    gen:   { x: 30, y: 200 },
  },
  advanced: {
    model:  { x: 5, y: 18 },
    bg:     { x: 5, y: 96 },
    rel:    { x: 5, y: 174 },
    plot:   { x: 5, y: 252 },
    relate: { x: 5, y: 330 },
    style:  { x: 55, y: 96 },
    req:    { x: 55, y: 174 },
    extra:  { x: 55, y: 252 },
    gen:    { x: 55, y: 330 },
  },
}

const nodePositions = ref<Record<string, { x: number; y: number }>>({})

function pos(id: string): { x: number; y: number } {
  const mode = creative2AdvancedMode.value ? 'advanced' : 'simple'
  return nodePositions.value[id] ?? DEFAULT_POS[mode][id] ?? { x: 5, y: 18 }
}

function resetLayout() {
  nodePositions.value = {}
  nextTick(drawLines)
}

// ===== 节点摘要与完成态 =====
function nodeSummary(n: CanvasNode): string {
  switch (n.id) {
    case 'model': return creative2ModelName.value || '点击选择模型'
    case 'style': return creative2PromptInfo.value?.name || '点击选择风格'
    case 'req': return creative2SecondPromptInfo.value?.name || '点击选择要求'
    case 'relate': {
      const count = creative2SelectedCharacterIds.value.length
        + (creative2SelectedChapterId.value ? 1 : 0)
        + (creative2SelectedMemoId.value ? 1 : 0)
      return count ? `${count} 项已关联` : '点击关联素材'
    }
    case 'gen': return '配置完成后生成'
    case 'bg': return creative2StoryBackground.value || '点击填写背景'
    case 'rel': return creative2CharacterRelations.value || '点击填写关系'
    case 'plot': return creative2ChapterPlot.value || '点击填写剧情'
    case 'extra': return creative2AdditionalInfo.value || '点击填写补充'
    default: return ''
  }
}

function nodeDone(n: CanvasNode): boolean {
  switch (n.id) {
    case 'model': return !!creative2ConfigId.value
    case 'style': return !!creative2PromptInfo.value
    case 'req': return !!creative2SecondPromptInfo.value
    case 'relate': return creative2SelectedCharacterIds.value.length > 0
      || !!creative2SelectedChapterId.value
      || !!creative2SelectedMemoId.value
    case 'gen': return false
    case 'bg': return !!creative2StoryBackground.value
    case 'rel': return !!creative2CharacterRelations.value
    case 'plot': return !!creative2ChapterPlot.value
    case 'extra': return !!creative2AdditionalInfo.value
    default: return false
  }
}

// ===== 节点点击 =====
function onNodeClick(n: CanvasNode) {
  if (didDrag) { didDrag = false; return }
  switch (n.id) {
    case 'model': openCreative2ModelSelectDialog(); break
    case 'style': creative2PromptDialogVisible.value = true; break
    case 'req': creative2SecondPromptDialogVisible.value = true; break
    case 'relate': creative2RelateDialogVisible.value = true; break
    case 'gen': handleCreative2Generate(); break
    case 'bg': openCreativeFullscreenEditor('storyBackground', creative2StoryBackground.value); break
    case 'rel': openCreativeFullscreenEditor('characterRelations', creative2CharacterRelations.value); break
    case 'plot': openCreativeFullscreenEditor('chapterPlot', creative2ChapterPlot.value); break
    case 'extra': openCreativeFullscreenEditor('additionalInfo', creative2AdditionalInfo.value); break
  }
}

// ===== 拖拽 =====
let didDrag = false
let dragState: { id: string; startX: number; startY: number; px: number; py: number; cw: number } | null = null

function onNodeMouseDown(id: string, e: MouseEvent) {
  if (id === 'gen') return
  if (e.target instanceof HTMLElement && e.target.closest('.node-edit')) return
  const canvas = canvasRef.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const p = pos(id)
  dragState = { id, startX: e.clientX, startY: e.clientY, px: p.x, py: p.y, cw: rect.width }
  draggingNode.value = id
  didDrag = false
  window.addEventListener('mousemove', onDragMove)
  window.addEventListener('mouseup', onDragEnd)
  e.preventDefault()
}

function onDragMove(e: MouseEvent) {
  const d = dragState
  const canvas = canvasRef.value
  if (!d || !canvas) return
  if (Math.abs(e.clientX - d.startX) + Math.abs(e.clientY - d.startY) > 5) didDrag = true
  const rect = canvas.getBoundingClientRect()
  const dxPct = (e.clientX - d.startX) / d.cw * 100
  const dyPx = e.clientY - d.startY
  const x = Math.max(0, Math.min(60, d.px + dxPct))
  const y = Math.max(0, Math.min(rect.height - 70, d.py + dyPx))
  nodePositions.value[d.id] = { x, y }
  drawLines()
}

function onDragEnd() {
  dragState = null
  draggingNode.value = null
  window.removeEventListener('mousemove', onDragMove)
  window.removeEventListener('mouseup', onDragEnd)
  // 延迟重置 didDrag，让紧随其后的 click 事件被吞掉（拖完不误触编辑）
  setTimeout(() => { didDrag = false }, 0)
}

// ===== 连线 =====
function drawLines() {
  const canvas = canvasRef.value
  const svg = linesRef.value
  if (!canvas || !svg) return
  const cRect = canvas.getBoundingClientRect()
  const genEl = canvas.querySelector('.canvas-node[data-id="gen"]') as HTMLElement | null
  if (!genEl) { svg.innerHTML = ''; return }
  const paths = canvasNodes.value
    .filter(n => n.id !== 'gen')
    .map(n => {
      const el = canvas.querySelector(`.canvas-node[data-id="${n.id}"]`) as HTMLElement | null
      if (!el) return ''
      const r1 = el.getBoundingClientRect()
      const r2 = genEl.getBoundingClientRect()
      const x1 = r1.left - cRect.left + r1.width
      const y1 = r1.top - cRect.top + r1.height / 2
      const x2 = r2.left - cRect.left
      const y2 = r2.top - cRect.top + r2.height / 2
      const mx = (x1 + x2) / 2
      return `<path d="M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}"/>`
    })
    .join('')
  svg.innerHTML = paths
}

// 监听画布尺寸变化（右侧栏宽度可拖拽 260-760px）
let resizeObserver: ResizeObserver | null = null
onMounted(() => {
  resizeObserver = new ResizeObserver(() => drawLines())
  if (canvasRef.value) resizeObserver.observe(canvasRef.value)
  nextTick(drawLines)
})
onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  window.removeEventListener('mousemove', onDragMove)
  window.removeEventListener('mouseup', onDragEnd)
})

// 模式切换：gen 节点归位 + 重绘
watch(creative2AdvancedMode, () => {
  delete nodePositions.value['gen']
  nextTick(drawLines)
})

// 数据变化后重绘连线（节点高度可能变化）
watch(
  [creative2ConfigId, creative2SelectedPrompts, creative2SecondPromptId,
   creative2StoryBackground, creative2CharacterRelations, creative2ChapterPlot,
   creative2AdditionalInfo, creative2SelectedCharacterIds, creative2SelectedChapterId,
   creative2SelectedMemoId],
  () => nextTick(drawLines)
)

// ===== 提示词介绍 =====
const creative2PromptInfo = computed(() => {
  if (creative2SelectedPrompts.value.length > 0) {
    return props.prompts.find(p => p.id === creative2SelectedPrompts.value[0])
  }
  return null
})

const creative2SecondPromptInfo = computed(() => {
  if (creative2SecondPromptId.value) {
    return props.prompts.find(p => p.id === creative2SecondPromptId.value)
  }
  return null
})

const creative2PromptIntroDialogVisible = ref(false)
const creative2SecondPromptIntroDialogVisible = ref(false)
const isPromptContentCollapsed = ref(true)

const showCreative2PromptIntro = () => {
  creative2PromptIntroDialogVisible.value = true
  isPromptContentCollapsed.value = true
}

const showCreative2SecondPromptIntro = () => {
  creative2SecondPromptIntroDialogVisible.value = true
  isPromptContentCollapsed.value = true
}

const togglePromptContentCollapse = () => {
  isPromptContentCollapsed.value = !isPromptContentCollapsed.value
}

// 监听提示词变化，清空字段值
watch([creative2SelectedPrompts, creative2UseFixedPrompt, creative2FixedPromptId], () => {
  creative2FieldValues.value = {}
}, { deep: true })

// ===== 生成 =====
const handleCreative2Generate = async () => {
  if (!creative2ConfigId.value) {
    ElMessage.warning('请选择AI模型')
    return
  }

  let promptsToUse: number[] = []

  if (creative2UseFixedPrompt.value && creative2FixedPromptId.value !== 0) {
    promptsToUse = [creative2FixedPromptId.value]
  } else if (creative2SelectedPrompts.value.length > 0) {
    promptsToUse = creative2SelectedPrompts.value
  }

  creative2Generating.value = true
  creative2WaitingForResponse.value = true
  creative2Result.value = ''
  creative2ResultDialogVisible.value = true
  creative2AbortController.value = new AbortController()

  try {
    const selectedPromptObjects = props.prompts.filter(p =>
      promptsToUse.includes(p.id)
    )

    const secondPromptObj = creative2SecondPromptId.value
      ? props.prompts.find(p => p.id === creative2SecondPromptId.value)
      : null

    // 将字段值代入提示词模板
    let systemPrompts = selectedPromptObjects.map(p => {
      let content = p.content
      if (creative2AllFields.value.length > 0) {
        creative2AllFields.value.forEach(field => {
          const value = creative2FieldValues.value[field.name] || ''
          const regex = new RegExp(`\\$\\{${field.name}\\}`, 'g')
          content = content.replace(regex, value)
        })
      }
      return content
    })

    // 如果有第二个提示词，也添加到系统提示词
    if (secondPromptObj) {
      systemPrompts.push(secondPromptObj.content)
    }

    const config = props.apiConfigs.find(c => c.id === creative2ConfigId.value)
    if (!config) {
      ElMessage.error('未找到 API 配置')
      return
    }

    // 构建用户消息内容
    let userMessageContent = ''

    // 添加故事背景
    if (creative2StoryBackground.value.trim()) {
      userMessageContent += `【故事背景】\n${creative2StoryBackground.value.trim()}\n\n`
    }

    // 添加角色关系
    if (creative2CharacterRelations.value.trim()) {
      userMessageContent += `【角色关系】\n${creative2CharacterRelations.value.trim()}\n\n`
    }

    // 添加本章剧情
    if (creative2ChapterPlot.value.trim()) {
      userMessageContent += `【本章剧情】\n${creative2ChapterPlot.value.trim()}\n\n`
    }

    // 添加补充信息
    if (creative2AdditionalInfo.value.trim()) {
      userMessageContent += `【补充信息】\n${creative2AdditionalInfo.value.trim()}\n\n`
    }

    // 添加关联角色卡内容
    if (creative2SelectedCharacterIds.value.length > 0) {
      const selectedCharacters = props.availableCharacters.filter(c =>
        creative2SelectedCharacterIds.value.includes(c.id)
      )
      if (selectedCharacters.length > 0) {
        userMessageContent += `【关联角色卡】（仅供参考）\n`
        selectedCharacters.forEach(character => {
          userMessageContent += `【${character.name}】\n${character.info || '（无描述）'}\n`
        })
        userMessageContent += '\n'
      }
    }

    // 添加关联章节内容
    if (creative2SelectedChapterId.value) {
      const chapter = props.chaptersList.find(c => c.id === creative2SelectedChapterId.value)
      if (chapter) {
        userMessageContent += `【关联章节】（仅供参考）\n${chapter.title}\n${chapter.content || '（无内容）'}\n\n`
      }
    }

    // 添加关联备忘录内容
    if (creative2SelectedMemoId.value) {
      const memo = props.memos.find(m => m.id === creative2SelectedMemoId.value)
      if (memo) {
        userMessageContent += `【关联备忘录】（仅供参考）\n${memo.title}\n${memo.content || '（无内容）'}\n\n`
      }
    }

    // 保存到对话记录
    const convRes = await conversationAPI.create({
      book_id: props.bookId,
      title: `创意生成 - ${selectedPromptObjects.map(p => p.name).join(', ')}${secondPromptObj ? ` + ${secondPromptObj.name}` : ''}`
    })

    if (convRes.success && convRes.data) {
      const convId = convRes.data.id

      // 保存系统提示词（如果有）
      if (systemPrompts.length > 0) {
        await conversationAPI.saveMessage(convId, {
          role: 'system',
          content: systemPrompts.join('\n\n---\n\n')
        })
      }

      // 保存用户消息
      await conversationAPI.saveMessage(convId, {
        role: 'user',
        content: userMessageContent
      })
    }

    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        configId: creative2ConfigId.value,
        systemPrompts: systemPrompts,
        messages: [{
          role: 'user',
          content: userMessageContent
        }],
        temperature: modelTemperature.value,
        top_p: modelTopP.value
      }),
      signal: creative2AbortController.value.signal
    })

    if (!response.ok) {
      throw new Error(`API请求失败: ${response.status}`)
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    if (reader) {
      let isFirstChunk = true
      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value)
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6)
            if (data === '[DONE]') continue

            try {
              const parsed = JSON.parse(data)
              if (parsed.error) {
                throw new Error(parsed.error)
              }
              if (parsed.content) {
                creative2Result.value += parsed.content
                if (isFirstChunk) {
                  creative2WaitingForResponse.value = false
                  isFirstChunk = false
                }
                await new Promise(resolve => setTimeout(resolve, 0))
                if (streamingScrollbarRef.value) {
                  const scrollbar = streamingScrollbarRef.value
                  if (scrollbar.wrapRef) {
                    scrollbar.wrapRef.scrollTop = scrollbar.wrapRef.scrollHeight
                  }
                }
              }
            } catch (e: any) {
              if (e.message && !e.message.includes('JSON')) {
                throw e
              }
            }
          }
        }
      }
    }

    ElMessage.success('生成完成')

    // 保存历史记录
    saveHistoryRecord(props.bookId, {
      title: `创意生成 - ${selectedPromptObjects.map(p => p.name).join('、') || '未命名任务'}`,
      source: 'creative2',
      sourceLabel: '正文创意生成',
      promptName: selectedPromptObjects.map(p => p.name).join('、'),
      promptCount: systemPrompts.length,
      status: 'completed',
      previewContent: (creative2Result.value || userMessageContent).substring(0, 140),
      messages: buildHistoryMessages({
        promptContents: systemPrompts,
        userContents: [userMessageContent],
        assistantContents: creative2Result.value ? [creative2Result.value] : []
      })
    })
  } catch (error: any) {
    if (error.name === 'AbortError') {
      ElMessage.info('已停止生成')
      saveHistoryRecord(props.bookId, {
        title: `创意生成 - ${props.prompts.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、') || '未命名任务'}`,
        source: 'creative2',
        sourceLabel: '正文创意生成',
        promptName: props.prompts.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、'),
        promptCount: props.prompts.filter(p => promptsToUse.includes(p.id)).length,
        status: 'cancelled',
        previewContent: (creative2Result.value || '已停止生成').substring(0, 140),
        messages: buildHistoryMessages({
          promptContents: props.prompts
            .filter(p => promptsToUse.includes(p.id))
            .map(p => p.content),
          userContents: [''],
          assistantContents: creative2Result.value ? [creative2Result.value] : []
        })
      })
    } else {
      console.error('创意生成错误:', error)
      saveHistoryRecord(props.bookId, {
        title: `创意生成 - ${props.prompts.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、') || '未命名任务'}`,
        source: 'creative2',
        sourceLabel: '正文创意生成',
        promptName: props.prompts.filter(p => promptsToUse.includes(p.id)).map(p => p.name).join('、'),
        promptCount: props.prompts.filter(p => promptsToUse.includes(p.id)).length,
        status: 'failed',
        previewContent: (creative2Result.value || '生成失败').substring(0, 140),
        messages: buildHistoryMessages({
          promptContents: props.prompts
            .filter(p => promptsToUse.includes(p.id))
            .map(p => p.content),
          userContents: [''],
          assistantContents: creative2Result.value ? [creative2Result.value] : []
        })
      })
      ElMessage.error(error.message || '生成失败')
    }
  } finally {
    creative2Generating.value = false
    creative2WaitingForResponse.value = false
    creative2AbortController.value = null
  }
}

const stopCreative2Generation = () => {
  if (creative2AbortController.value) {
    creative2AbortController.value.abort()
  }
}

const applyCreative2Result = () => {
  if (!props.hasChapter || !creative2Result.value) {
    ElMessage.warning('请选择章节或没有可应用的内容')
    return
  }
  emit('apply-result', creative2Result.value)
  creative2ResultDialogVisible.value = false
}

// ===== 全屏编辑 =====
const creativeFullscreenDialogVisible = ref(false)
const creativeFullscreenContent = ref('')
const creativeFullscreenField = ref('')

const openCreativeFullscreenEditor = (field: string, content: string) => {
  creativeFullscreenField.value = field
  creativeFullscreenContent.value = content
  creativeFullscreenDialogVisible.value = true
}

const saveCreativeFullscreenEditor = () => {
  switch (creativeFullscreenField.value) {
    case 'storyBackground':
      creative2StoryBackground.value = creativeFullscreenContent.value
      break
    case 'characterRelations':
      creative2CharacterRelations.value = creativeFullscreenContent.value
      break
    case 'chapterPlot':
      creative2ChapterPlot.value = creativeFullscreenContent.value
      break
    case 'additionalInfo':
      creative2AdditionalInfo.value = creativeFullscreenContent.value
      break
  }
  creativeFullscreenDialogVisible.value = false
  creativeFullscreenField.value = ''
  creativeFullscreenContent.value = ''
}

const cancelCreativeFullscreenEditor = () => {
  creativeFullscreenDialogVisible.value = false
  creativeFullscreenField.value = ''
  creativeFullscreenContent.value = ''
}
</script>


<style scoped>


/* ========== AI 抽卡区样式 ========== */
.creative-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: linear-gradient(180deg, rgba(252, 255, 255, 0.7) 0%, rgba(244, 251, 250, 0.78) 100%);
}


.creative-header {
  padding: 10px 14px;
  border-bottom: 1px solid rgba(8, 198, 190, 0.16);
  display: flex;
  align-items: center;
  gap: 8px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.9) 0%, rgba(242, 251, 250, 0.88) 100%);
}


.creative-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}


.creative-title {
  flex: 1;
  font-weight: 700;
  font-size: 15px;
  color: #0f766e;
}


.creative-content {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px 50px 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}


.creative-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
}


.fixed-height-section .prompt-picker-selected {
  flex: 1;
  min-width: 0;
}


.fixed-height-section .prompt-picker-brief {
  flex-shrink: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  line-clamp: 1;
  -webkit-box-orient: vertical;
  color: #4b5563;
  font-size: 12px;
  line-height: 1.4;
}


.fixed-height-section .prompt-picker-brief :deep(*) {
  color: #4b5563;
  font-size: 12px;
}

.fixed-height-section .prompt-picker-brief :deep(strong),
.fixed-height-section .prompt-picker-brief :deep(b) {
  font-weight: 600;
  color: #1f2937;
}


.fixed-height-section .prompt-picker-brief :deep(em),
.fixed-height-section .prompt-picker-brief :deep(i) {
  font-style: italic;
}


.fixed-height-section .prompt-picker-brief :deep(br) {
  line-height: 1.5;
}


.section-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  color: #134e4a;
  font-size: 13px;
  margin-bottom: 1px;
}


.section-label .el-icon {
  font-size: 14px;
  color: #08c6be;
}


.context-tip {
  font-size: 12px;
  color: #62807c;
  line-height: 1.5;
}


.creative-section .el-select .el-input__wrapper {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(8, 198, 190, 0.2);
  padding: 6px 10px;
}


.creative-section .el-select .el-input__wrapper {
  border-radius: 6px;
}


.creative-section .el-select .el-input__wrapper:hover {
  border-color: rgba(8, 198, 190, 0.4);
}


.creative-section .el-select .el-input__wrapper:focus-within {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}


.creative-section .el-textarea__inner {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(8, 198, 190, 0.2);
  padding: 6px 10px;
  font-size: 13px;
  border-radius: 6px;
}


.creative-section .el-textarea__inner:hover {
  border-color: rgba(8, 198, 190, 0.4);
}


.creative-section .el-textarea__inner:focus {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}


.creative-section .el-input__wrapper {
  background: rgba(255, 255, 255, 0.9);
  border-color: rgba(8, 198, 190, 0.2);
  padding: 6px 10px;
  border-radius: 6px;
}


.creative-section .el-input__wrapper:hover {
  border-color: rgba(8, 198, 190, 0.4);
}


.creative-section .el-input__wrapper:focus-within {
  border-color: #08c6be;
  box-shadow: 0 0 0 2px rgba(8, 198, 190, 0.1);
}


.creative-section .el-button--primary {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: transparent;
  box-shadow: 0 4px 12px rgba(8, 198, 190, 0.15);
  font-size: 13px;
  padding: 8px 16px;
  border-radius: 6px;
}


.creative-section .el-button--primary:hover {
  background: linear-gradient(135deg, #09b4ad 0%, #047a75 100%);
  border-color: transparent;
}


.creative-section .el-button--primary.is-disabled {
  background: rgba(8, 198, 190, 0.5);
  border-color: transparent;
}


.creative-content::-webkit-scrollbar {
  width: 6px;
}


.creative-content::-webkit-scrollbar-track {
  background: transparent;
}


.creative-content::-webkit-scrollbar-thumb {
  background: rgba(8, 198, 190, 0.2);
  border-radius: 3px;
}


.creative-content::-webkit-scrollbar-thumb:hover {
  background: rgba(8, 198, 190, 0.35);
}


/* 生成按钮固定定位 */
.creative-panel {
  position: relative;
}


.creative-generate-btn-wrapper {
  position: absolute;
  bottom: 16px;
  right: 16px;
  z-index: 10;
}


.creative-generate-btn {
  box-shadow: 0 8px 20px rgba(8, 198, 190, 0.25);
  padding: 10px 24px;
  font-size: 14px;
  font-weight: 600;
}


.creative-generate-btn:hover {
  box-shadow: 0 10px 24px rgba(8, 198, 190, 0.35);
}


/* creative-section 暗色主题 */
:root[data-theme='dark'] .creative-section {
  background: transparent;
  border-color: transparent;
  box-shadow: none;
}


:root[data-theme='dark'] .section-label {
  color: #e6edf3;
}


:root[data-theme='dark'] .section-label .el-icon {
  color: #93c5fd;
}


:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief {
  color: #9ca3af;
}


:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief :deep(*) {
  color: #9ca3af;
}


:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief :deep(strong),
:root[data-theme='dark'] .fixed-height-section .prompt-picker-brief :deep(b) {
  color: #e6edf3;
}


:root[data-theme='dark'] .creative-section .el-select .el-input__wrapper {
  background: rgba(28, 35, 51, 0.6);
  border-color: rgba(42, 50, 66, 0.4);
}


:root[data-theme='dark'] .creative-section .el-select .el-input__wrapper:hover {
  border-color: rgba(59, 130, 246, 0.4);
}


:root[data-theme='dark'] .creative-section .el-select .el-input__wrapper:focus-within {
  border-color: #3b82f6;
}


:root[data-theme='dark'] .creative-section .el-textarea__inner {
  background: rgba(28, 35, 51, 0.6);
  border-color: rgba(42, 50, 66, 0.4);
  color: #e6edf3;
}


:root[data-theme='dark'] .creative-section .el-textarea__inner:hover {
  border-color: rgba(59, 130, 246, 0.4);
}


:root[data-theme='dark'] .creative-section .el-textarea__inner:focus {
  border-color: #3b82f6;
}


:root[data-theme='dark'] .creative-section .el-input__wrapper {
  background: rgba(28, 35, 51, 0.6);
  border-color: rgba(42, 50, 66, 0.4);
}


:root[data-theme='dark'] .creative-section .el-input__wrapper:hover {
  border-color: rgba(59, 130, 246, 0.4);
}


:root[data-theme='dark'] .creative-section .el-input__wrapper:focus-within {
  border-color: #3b82f6;
}


:root[data-theme='dark'] .creative-section .el-input__inner {
  color: #e6edf3;
}


:root[data-theme='dark'] .prompt-intro-content h3 {
  color: #e6edf3;
}


.prompt-intro-content h3 {
  color: #1f2937;
  margin-bottom: 8px;
}


.prompt-intro-dialog :deep(.el-dialog) {
  max-height: 72vh;
  display: flex;
  flex-direction: column;
}


.prompt-intro-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
}


.prompt-intro-content {
  max-height: 52vh;
  overflow-y: auto;
  padding-right: 6px;
}


.prompt-intro-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
}


.prompt-intro-header {
  display: flex;
  align-items: center;
  gap: 10px;
}


.prompt-intro-name {
  margin: 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}


.prompt-intro-section {
  display: flex;
  flex-direction: column;
  gap: 6px;
}


.prompt-intro-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  color: #374151;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 0;
  transition: color 0.2s;
}


.prompt-intro-label:hover {
  color: #08c6be;
}


.collapse-icon {
  transition: transform 0.3s;
  font-size: 14px;
  color: #08c6be;
}


.collapse-icon.collapsed {
  transform: rotate(0deg);
}


.collapse-icon:not(.collapsed) {
  transform: rotate(90deg);
}


.prompt-intro-description {
  color: #4b5563;
  font-size: 13px;
  line-height: 1.5;
  padding: 8px 10px;
  background: rgba(8, 198, 190, 0.05);
  border-radius: 6px;
  border-left: 2px solid #08c6be;
}


.prompt-intro-description :deep(*) {
  color: #4b5563;
  font-size: 13px;
}


.prompt-intro-content-text {
  color: #374151;
  font-size: 13px;
  line-height: 1.6;
  padding: 10px 12px;
  background: #f9fafb;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
  white-space: pre-wrap;
  max-height: 300px;
  overflow-y: auto;
}


.model-select-trigger {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}


.model-select-trigger:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.25);
}


.creative-model-trigger {
  width: 100%;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.12);
}


.creative-model-trigger:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}


.model-select-name {
  font-size: 13px;
  color: #e5e7eb;
}


.model-select-arrow {
  font-size: 12px;
  color: #9ca3af;
}


.input-container {
  flex: 1;
  position: relative;
}


.input-container .fullscreen-btn {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 24px;
  height: 24px;
  padding: 0;
  border: none;
  background: rgba(255, 255, 255, 0.95);
  color: #6b7280;
  z-index: 10;
  transition: all 0.2s ease;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}


.input-container .fullscreen-btn:hover {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
  box-shadow: 0 2px 6px rgba(16, 185, 129, 0.2);
}


.creative-section .input-container {
  position: relative;
}


.creative-section .input-container :deep(.el-input__wrapper) {
  padding-right: 32px;
}


.creative-section .input-container :deep(.el-textarea__inner) {
  padding-right: 32px;
}


/* ========== AI 写作生成结果弹窗样式 ========== */
.creative2-result-dialog :deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
  min-height: 400px;
}


.creative2-result-content {
  min-height: 300px;
  display: flex;
  flex-direction: column;
}


.streaming-content {
  min-height: 300px;
  display: flex;
  flex-direction: column;
}


.generating-animation {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  min-height: 300px;
}


.generating-spinner {
  font-size: 40px;
  color: #00c9a7;
  margin-bottom: 20px;
}


.generating-text {
  font-size: 14px;
  color: #909399;
  font-weight: 400;
}


.result-body {
  flex: 1;
  overflow: hidden;
}


.result-complete {
  min-height: 300px;
  display: flex;
  flex-direction: column;
}


.empty-result {
  min-height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
}


.creative2-result-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 8px 0;
}


.prompt-picker-selected {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(8, 198, 190, 0.18);
  cursor: pointer;
  transition: all 0.25s ease;
}


.prompt-picker-selected:hover {
  background: rgba(255, 255, 255, 0.7);
  border-color: rgba(8, 198, 190, 0.28);
}


.prompt-picker-name {
  flex: 1;
  min-width: 0;
  color: #134e4a;
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}


.prompt-picker-arrow {
  color: #08c6be;
  font-size: 16px;
}


.prompt-picker-brief {
  margin-top: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #6b9b97;
  font-size: 12px;
  line-height: 1.4;
}


:root[data-theme='dark'] .model-select-trigger {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}


:root[data-theme='dark'] .model-select-trigger:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.2);
}


:root[data-theme='dark'] .creative-model-trigger {
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.1);
}


:root[data-theme='dark'] .creative-model-trigger:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.16);
}


/* ========== 节点画布样式 ========== */
.canvas-wrap {
  flex: 1;
  position: relative;
  min-height: 0;
}

.canvas {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background:
    radial-gradient(circle at 1px 1px, rgba(8, 198, 190, 0.12) 1px, transparent 0) 0 0/22px 22px,
    linear-gradient(180deg, rgba(252, 255, 255, 0.7) 0%, rgba(244, 251, 250, 0.78) 100%);
}

.canvas-lines {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.canvas-lines :deep(path) {
  fill: none;
  stroke: rgba(8, 198, 190, 0.45);
  stroke-width: 2;
  stroke-dasharray: 5 4;
  animation: canvas-dash 1.2s linear infinite;
}

@keyframes canvas-dash {
  to { stroke-dashoffset: -18; }
}

.canvas-toolbar {
  position: absolute;
  top: 10px;
  right: 12px;
  z-index: 3;
  display: flex;
  align-items: center;
  gap: 8px;
}

.canvas-mode {
  font-size: 10.5px;
  color: #62807c;
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid rgba(8, 198, 190, 0.2);
  border-radius: 999px;
  padding: 3px 10px;
  white-space: nowrap;
}

.reset-layout {
  font-size: 10.5px;
  color: #0f766e;
  background: #fff;
  border: 1px solid rgba(8, 198, 190, 0.3);
  border-radius: 999px;
  padding: 3px 10px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.2s;
}

.reset-layout:hover {
  background: #e0f5f2;
  border-color: #08c6be;
}

.canvas-node {
  position: absolute;
  z-index: 2;
  width: 40%;
  max-width: 200px;
  min-width: 100px;
  background: #fff;
  border: 1.5px solid rgba(8, 198, 190, 0.25);
  border-radius: 12px;
  padding: 10px 12px;
  cursor: grab;
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.06);
  transition: box-shadow 0.2s;
  user-select: none;
}

.canvas-node:hover {
  box-shadow: 0 8px 22px rgba(8, 198, 190, 0.14);
}

.canvas-node.dragging {
  cursor: grabbing;
  box-shadow: 0 12px 30px rgba(8, 198, 190, 0.2);
  z-index: 5;
}

.canvas-node.done {
  border-color: rgba(8, 198, 190, 0.5);
  background: linear-gradient(180deg, #fff 0%, #f7fcfb 100%);
}

.canvas-node.gen {
  border-color: transparent;
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  color: #fff;
  box-shadow: 0 8px 24px rgba(8, 198, 190, 0.35);
  cursor: pointer;
}

.node-head {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 12px;
  font-weight: 700;
  color: #134e4a;
}

.node-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #08c6be;
  flex-shrink: 0;
}

.node-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.node-tag {
  margin-left: auto;
  font-size: 10px;
  color: #94a3b8;
  font-weight: 400;
  flex-shrink: 0;
}

.node-val {
  margin-top: 6px;
  font-size: 11.5px;
  color: #62807c;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
}

.node-val.empty {
  color: #cbd5e1;
}

.node-edit {
  margin-top: 7px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: #08c6be;
  font-weight: 600;
  cursor: pointer;
}

.node-edit:hover {
  text-decoration: underline;
}

.canvas-node.gen .node-head {
  color: #fff;
}

.canvas-node.gen .node-dot {
  background: rgba(255, 255, 255, 0.9);
}

.canvas-node.gen .node-val {
  color: rgba(255, 255, 255, 0.85);
}

.canvas-node.gen .node-tag {
  color: rgba(255, 255, 255, 0.7);
}

/* 关联素材弹窗 */
.relate-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.relate-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.relate-label {
  font-size: 12.5px;
  font-weight: 600;
  color: #134e4a;
}

.relate-tip {
  font-size: 11.5px;
  color: #94a3b8;
  line-height: 1.5;
}

/* 节点画布深色模式 */
:root[data-theme='dark'] .canvas {
  background:
    radial-gradient(circle at 1px 1px, rgba(59, 130, 246, 0.12) 1px, transparent 0) 0 0/22px 22px,
    linear-gradient(180deg, rgba(22, 27, 34, 0.6) 0%, rgba(17, 21, 27, 0.7) 100%);
}

:root[data-theme='dark'] .canvas-lines :deep(path) {
  stroke: rgba(59, 130, 246, 0.4);
}

:root[data-theme='dark'] .canvas-node {
  background: rgba(28, 35, 51, 0.9);
  border-color: rgba(42, 50, 66, 0.5);
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .canvas-node.done {
  border-color: rgba(59, 130, 246, 0.5);
  background: linear-gradient(180deg, rgba(28, 35, 51, 0.9) 0%, rgba(22, 27, 34, 0.9) 100%);
}

:root[data-theme='dark'] .canvas-node.gen {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.35);
}

:root[data-theme='dark'] .node-head {
  color: #e6edf3;
}

:root[data-theme='dark'] .node-dot {
  background: #3b82f6;
}

:root[data-theme='dark'] .node-tag {
  color: #9ca3af;
}

:root[data-theme='dark'] .node-val {
  color: #9ca3af;
}

:root[data-theme='dark'] .node-val.empty {
  color: #4b5563;
}

:root[data-theme='dark'] .node-edit {
  color: #93c5fd;
}

:root[data-theme='dark'] .canvas-mode {
  color: #9ca3af;
  background: rgba(28, 35, 51, 0.8);
  border-color: rgba(42, 50, 66, 0.5);
}

:root[data-theme='dark'] .reset-layout {
  color: #93c5fd;
  background: rgba(28, 35, 51, 0.8);
  border-color: rgba(42, 50, 66, 0.5);
}

:root[data-theme='dark'] .reset-layout:hover {
  background: rgba(59, 130, 246, 0.15);
  border-color: #3b82f6;
}

:root[data-theme='dark'] .relate-label {
  color: #e6edf3;
}

:root[data-theme='dark'] .relate-tip {
  color: #9ca3af;
}
</style>
