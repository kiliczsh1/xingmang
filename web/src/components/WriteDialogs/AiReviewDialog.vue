<template>
  <!-- AI审稿 / AI去痕 · 生成器式弹窗（逐行克隆 AiAnalyzeDialog = Creative generator-dialog 骨架 + Cyber 皮肤） -->
  <el-dialog
    v-model="visible"
    width="calc(100vw - 128px)"
    align-center
    destroy-on-close
    class="generator-dialog xm-dialog"
    modal-class="generator-dialog-modal"
    :show-close="false"
    :close-on-click-modal="false"
    append-to-body
  >
    <template #header>
      <div class="gen-dialog-header">
        <div class="gen-dialog-title">
          <el-icon class="gen-title-icon"><MagicStick /></el-icon>
          <span class="title-text">{{ tool?.label }} ·《{{ sourceTitle || '当前章节' }}》</span>
        </div>
        <div class="gen-dialog-header-right">
          <el-button class="gen-close-btn" type="primary" link @click="visible = false">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </template>

    <!-- ============ 配置视图 ============ -->
    <div v-if="dialogStep === 'config'" class="dialog-content">
      <!-- 左栏：固定提示词（初始为工具预设，可从卡包选取后固定，最多 12 条，超出滚动） -->
      <div class="prompt-template-area">
        <div class="template-header">
          <span class="area-label">固定提示词</span>
          <span class="area-hint">已固定 {{ pinnedPrompts.length }} / 12</span>
        </div>
        <div class="template-list" v-if="pinnedPrompts.length">
          <div
            v-for="p in pinnedPrompts"
            :key="p.id"
            class="template-item"
            :class="{ active: selectedPromptId === p.id }"
            @click="selectPrompt(p.id)"
          >
            <el-button class="template-item-detail" size="small" @click.stop="showPromptDetail(p)">详情</el-button>
            <span class="template-item-name" :title="p.name">{{ p.name }}</span>
            <span class="template-item-pack" :title="p.category">{{ p.category || '未分类' }}</span>
            <button class="template-item-remove" type="button" :aria-label="`移除 ${p.name}`" @click.stop="removePinned(p.id)">✕</button>
          </div>
        </div>
        <div v-else class="empty-template">
          <template v-if="promptPool.length">尚未固定提示词，点下方「更多提示词」从本地卡包选取（默认使用内置话术）</template>
          <template v-else>提示词库里还没有可用提示词，可先去提示词库添加；生成将使用内置话术</template>
        </div>
        <div class="template-footer">
          <el-button class="more-prompt-btn" size="large" @click="openPromptSelect">更多提示词</el-button>
        </div>
      </div>

      <!-- 右栏：表单 -->
      <div class="form-area">
        <div class="form-block">
          <div class="form-block-label">AI 模型</div>
          <el-select v-model="selectedModel" placeholder="选择模型" style="width: 100%">
            <el-option v-for="m in models" :key="m.id" :label="modelLabel(m)" :value="m.id" />
          </el-select>
        </div>

        <div class="form-block">
          <div class="form-block-label">提示词</div>
          <el-radio-group v-model="promptMode" size="small" class="prompt-mode">
            <el-radio-button value="quick">快速</el-radio-button>
            <el-radio-button value="custom">自定义</el-radio-button>
          </el-radio-group>
          <el-input
            v-if="promptMode === 'custom'"
            v-model="customPrompt"
            type="textarea"
            :rows="5"
            placeholder="输入自定义提示词（system）"
          />
          <div v-else class="quick-prompt">
            <div class="quick-prompt-name">{{ selectedPromptName }}</div>
            <div class="quick-prompt-tip">可在左侧切换提示词，或切到「自定义」编辑</div>
          </div>
        </div>

        <div class="form-block">
          <div class="form-block-label">关联章节</div>
          <el-select v-model="selectedChapterIds" multiple filterable placeholder="可选，选择要一并参考的章节" style="width: 100%">
            <el-option v-for="c in chaptersList" :key="c.id" :label="c.title" :value="c.id" />
          </el-select>
        </div>

        <div class="form-block">
          <div class="form-block-label">关联角色卡</div>
          <el-select v-model="selectedCharacterIds" multiple filterable placeholder="可选，选择相关角色" style="width: 100%">
            <el-option v-for="c in characters" :key="c.id" :label="c.name" :value="c.id" />
          </el-select>
        </div>

        <div class="form-block">
          <div class="form-block-label">补充内容</div>
          <el-input v-model="additionalInfo" type="textarea" :rows="3" :maxlength="tool?.extraField?.maxLen || 500" placeholder="可选，补充你想让 AI 关注的信息" />
        </div>
      </div>
    </div>

    <!-- ============ 结果视图 ============ -->
    <div v-else class="dialog-content result-step">
      <div class="result-toolbar">
        <div class="result-title">
          <el-icon><MagicStick /></el-icon>
          <span>生成结果</span>
        </div>
        <div class="result-actions">
          <el-button v-if="generating" type="danger" link @click="stopGenerate">停止</el-button>
          <el-button v-if="output" type="primary" link @click="copyOutput">复制</el-button>
          <el-button v-if="canApply && output && !generating" type="primary" link @click="applyToEditor">应用到正文</el-button>
          <el-button v-if="!generating" type="primary" link @click="backToConfig">返回配置</el-button>
        </div>
      </div>
      <div class="result-scrollbar">
        <div v-if="generating && !output" class="generating-hint">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>正在生成中…</span>
        </div>
        <MarkdownRenderer v-else :content="output || '暂无生成内容'" />
      </div>
    </div>

    <template #footer>
      <div class="gen-dialog-footer">
        <div class="gen-footer-disclaimer">以上内容均由 AI 生成，仅供参考和借鉴</div>
        <div class="gen-footer-actions">
          <template v-if="dialogStep === 'config'">
            <el-button class="gen-generate-btn" type="success" round :disabled="!canGenerate" @click="generateContent">{{ tool?.startLabel || '生成' }}</el-button>
          </template>
          <template v-else>
            <el-button v-if="generating" class="gen-generate-btn stop" round @click="stopGenerate">停止生成</el-button>
            <template v-else>
              <el-button size="large" round @click="backToConfig">返回配置</el-button>
              <el-button class="gen-generate-btn" type="success" round @click="generateContent">重新生成</el-button>
            </template>
          </template>
        </div>
      </div>
    </template>
  </el-dialog>

  <!-- 更多提示词：从本地卡包挑一条固定进左栏 -->
  <PromptMoreDialog
    v-model:visible="promptSelectVisible"
    :prompts="toolPrompts"
    :title="tool?.requirementLabel || '更多提示词'"
    :scope="tool?.id || 'review'"
    :selected-id="selectedPromptId"
    add-mode
    :added-ids="pinnedPrompts.map(p => p.id)"
    @add="pinFromMore"
    @remove="removePinned"
    @deleted="onPromptDeleted"
  />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, Close, Loading } from '@element-plus/icons-vue'
import { configAPI, promptAPI } from '@/api'
import type { ApiModel, Prompt, Chapter, Character } from '@/types'
import type { AiToolMeta } from '@/constants/aiTools'
import { promptsForTool, toolCoreWords } from '@/constants/aiTools'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import PromptMoreDialog from '@/components/WriteDialogs/PromptMoreDialog.vue'
import { bumpUsage, readPinned, writePinned } from '@/utils/prompt-pick-store'

defineOptions({ name: 'AiReviewDialog' })

const props = defineProps<{
  visible: boolean
  tool: AiToolMeta | null
  /** 素材文本（本章全文或选中文本，无字数上限） */
  sourceText: string
  sourceTitle: string
  /** 当前章节 id（避免关联章节与素材重复） */
  currentChapterId: number | null
  /** 关联章节可选列表 */
  chaptersList: Chapter[]
  /** 关联角色卡可选列表 */
  characters: Character[]
}>()
const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'finish', payload: { toolId: string; text: string }): void
  (e: 'apply', payload: { applyMode: string; text: string }): void
}>()
const visible = computed({ get: () => props.visible, set: v => emit('update:visible', v) })

/* 工具 → 对应卡包：不再写死卡包名，改用工具核心词与卡包名双向匹配（「世界观生成」↔「世界观」） */
const dialogStep = ref<'config' | 'result'>('config')
const models = ref<ApiModel[]>([])
const selectedModel = ref<number | null>(null)
const allPrompts = ref<Prompt[]>([])
const pinnedPrompts = ref<Prompt[]>([])
const promptSelectVisible = ref(false)
const selectedPromptId = ref<number | null>(null)
const promptMode = ref<'quick' | 'custom'>('quick')
const customPrompt = ref('')
const additionalInfo = ref('')
// 关联章节 / 关联角色卡（作为 relatedContent 参考内容发送）
const selectedChapterIds = ref<number[]>([])
const selectedCharacterIds = ref<number[]>([])
const output = ref('')
const generating = ref(false)
const abortController = ref<AbortController | null>(null)

/** 该工具对应的提示词卡包内容（按核心词匹配；一个都没命中时回退全量） */
const toolPrompts = computed(() => promptsForTool(allPrompts.value, props.tool))

/* 选卡弹窗只显示当前工具对应卡包内的提示词 */
const promptPool = computed(() => toolPrompts.value)

/* 卡包不存在则登记创建（与 Prompts.vue「创建卡包」同一机制：写入 prompt_categories，幂等） */
function ensurePackRegistered(name: string) {
  try {
    const list: string[] = JSON.parse(localStorage.getItem('prompt_categories') || '[]')
    if (Array.isArray(list) && !list.includes(name)) {
      list.push(name)
      localStorage.setItem('prompt_categories', JSON.stringify(list))
    }
  } catch {
    /* localStorage 不可用时忽略 */
  }
}

const modelLabel = (m: ApiModel) => (m.provider_name ? `${m.provider_name} · ${m.name}` : m.name)

const selectedPromptName = computed(() => {
  const p = pinnedPrompts.value.find(pp => pp.id === selectedPromptId.value)
  return p?.name || props.tool?.label + '（默认话术）'
})

/* 三级回退：自定义 → 选中的固定提示词 → 工具内置 system 话术 */
const systemPrompt = computed(() => {
  if (promptMode.value === 'custom' && customPrompt.value.trim()) return customPrompt.value
  const p = pinnedPrompts.value.find(pp => pp.id === selectedPromptId.value)
  if (p?.content) return p.content
  return props.tool?.systemPrompt || ''
})

const canGenerate = computed(() => !!selectedModel.value && (promptMode.value === 'custom' ? !!customPrompt.value.trim() : true))

const canApply = computed(() => ['replace', 'insert', 'append'].includes(props.tool?.applyMode || ''))

/* 素材拼接：章节全文（或选中文本）+ 补充内容，无字数上限 */
function buildUserPrompt(): string {
  const parts: string[] = []
  if (props.sourceText.trim()) parts.push(`【章节：${props.sourceTitle || '当前章节'}】\n${props.sourceText}`)
  if (additionalInfo.value.trim()) parts.push(`补充内容：${additionalInfo.value.trim()}`)
  return parts.join('\n\n')
}

/* 关联章节 / 角色卡 → relatedContent（后端拼进 system 的「参考内容」） */
function buildRelatedContent(): Array<{ type: string; title: string; content: string }> {
  const out: Array<{ type: string; title: string; content: string }> = []
  for (const id of selectedChapterIds.value) {
    if (props.currentChapterId !== null && id === props.currentChapterId) continue
    const chapter = props.chaptersList.find(c => c.id === id)
    if (!chapter) continue
    out.push({
      type: 'chapter',
      title: chapter.title,
      content: (chapter.content || '').replace(/<[^>]*>/g, '').trim()
    })
  }
  for (const id of selectedCharacterIds.value) {
    const c = props.characters.find(x => x.id === id)
    if (!c) continue
    const info = [c.personality, c.info].filter(Boolean).join('\n')
    out.push({ type: 'character', title: c.name, content: info || '（暂无设定）' })
  }
  return out
}

function selectPrompt(id: number) {
  selectedPromptId.value = id
  promptMode.value = 'quick'
}

function showPromptDetail(p: Prompt) {
  ElMessageBox.alert(p.content, p.name, { confirmButtonText: '知道了', type: 'info' })
}

function openPromptSelect() {
  promptSelectVisible.value = true
}

/* 从「更多提示词」选一条：固定进左栏（去重、上限 12 条）并立即选中，同时持久化 */
function pinFromMore(p: Prompt) {
  const next = [p, ...pinnedPrompts.value.filter(x => x.id !== p.id)].slice(0, 12)
  pinnedPrompts.value = next
  selectedPromptId.value = p.id
  promptMode.value = 'quick'
  persistPins()
  bumpUsage(p.id)
}

function removePinned(id: number) {
  pinnedPrompts.value = pinnedPrompts.value.filter(p => p.id !== id)
  if (selectedPromptId.value === id) {
    selectedPromptId.value = pinnedPrompts.value[0]?.id ?? null
  }
  persistPins()
}

/** 固定列表落本地：下次打开直接恢复，不用重新到卡包里挑 */
function persistPins() {
  writePinned(props.tool?.id || 'review', pinnedPrompts.value.map(p => p.id))
}

/* 「更多提示词」里删掉了一条：同步从固定列表移除 */
function onPromptDeleted(id: number) {
  allPrompts.value = allPrompts.value.filter(p => p.id !== id)
  removePinned(id)
}

/* 打开弹窗时恢复上次固定的提示词，并按 id 回填最新内容（卡包/正文可能已改） */
function restorePinned(toolId: string) {
  const ids = readPinned(toolId)
  const byId = new Map<number, Prompt>()
  for (const p of allPrompts.value) byId.set(p.id, p)
  const restored: Prompt[] = []
  for (const id of ids) {
    const p = byId.get(id)
    if (p && !restored.some(x => x.id === id)) restored.push(p)
  }
  pinnedPrompts.value = restored.slice(0, 12)
  selectedPromptId.value = pinnedPrompts.value[0]?.id ?? null
}

/* 打开弹窗时复位并加载模型 / 提示词（预设预置进固定列表，默认选中第一条） */
watch(
  () => props.visible,
  async v => {
    if (!v) {
      if (generating.value) stopGenerate()
      return
    }
    ensurePackRegistered(toolCoreWords(props.tool)[0] || props.tool?.label || '通用')
    dialogStep.value = 'config'
    output.value = ''
    additionalInfo.value = ''
    generating.value = false
    abortController.value = null
    promptMode.value = 'quick'
    customPrompt.value = ''
    selectedModel.value = null
    // 左栏恢复上次固定的提示词（记住选择，不用每次重新到卡包里挑）
    pinnedPrompts.value = []
    selectedPromptId.value = null
    promptSelectVisible.value = false
    selectedChapterIds.value = []
    selectedCharacterIds.value = []
    try {
      const [modelRes, promptRes] = await Promise.all([configAPI.getAll(), promptAPI.getAll()])
      models.value = (modelRes.data ?? []).filter(m => m.enabled !== 0)
      const def = models.value.find(m => m.is_default === 1) || models.value[0]
      selectedModel.value = def?.id ?? null
      allPrompts.value = promptRes.data ?? []
      restorePinned(props.tool?.id || 'review')
    } catch {
      /* 模型/提示词加载失败不阻塞，仍可用默认话术与自定义 */
    }
  }
)

async function generateContent() {
  const modelData = models.value.find(m => m.id === selectedModel.value)
  if (!modelData) {
    ElMessage.warning('请先选择 AI 模型')
    return
  }
  generating.value = true
  dialogStep.value = 'result'
  output.value = ''
  const controller = new AbortController()
  abortController.value = controller
  try {
    const res = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        messages: [{ role: 'user', content: buildUserPrompt() }],
        configId: modelData.id,
        systemPrompts: [systemPrompt.value],
        relatedContent: buildRelatedContent(),
      }),
      signal: controller.signal,
    })
    if (!res.ok) throw new Error(`请求失败（${res.status}）`)
    if (!res.body) throw new Error('响应无流')
    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let acc = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      const text = decoder.decode(value, { stream: true })
      for (const line of text.split('\n')) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6)
        if (data === '[DONE]') continue
        try {
          const json = JSON.parse(data)
          if (json.error) throw new Error(json.error)
          if (json.content) {
            acc += json.content
            output.value = acc
          }
        } catch (e) {
          if (e instanceof Error) throw e
        }
      }
    }
    if (output.value && props.tool?.id) {
      emit('finish', { toolId: props.tool.id, text: output.value })
    }
    ElMessage.success('生成成功')
  } catch (e) {
    const err = e as { name?: string; message?: string }
    if (err?.name === 'AbortError') ElMessage.info('已停止生成')
    else ElMessage.error('生成失败：' + (err?.message || '未知错误'))
  } finally {
    generating.value = false
    abortController.value = null
  }
}

function stopGenerate() {
  abortController.value?.abort()
  generating.value = false
}

function copyOutput() {
  if (!output.value) return
  navigator.clipboard?.writeText(output.value).then(
    () => ElMessage.success('已复制到剪贴板'),
    () => ElMessage.warning('复制失败'),
  )
}

function applyToEditor() {
  if (!output.value) return
  emit('apply', { applyMode: props.tool?.applyMode || 'copy', text: output.value })
  emit('update:visible', false)
}

function backToConfig() {
  dialogStep.value = 'config'
}
</script>

<style>
/* ============ AI 审稿/去痕弹窗（生成器式，照搬 Creative generator-dialog 的 Cyber 皮肤） ============ */
/* 非 scoped：append-to-body 面板根元素无 scope id；并自带暗色覆盖（Creative 主面板暗色不跟随，这里修复） */
.generator-dialog-modal {
  --cyber-bg: #f8fafc;
  --cyber-bg-soft: #fff;
  --cyber-panel: #fff;
  --cyber-panel2: #f1f5f9;
  --cyber-line: rgba(15, 23, 42, 0.08);
  --cyber-line-strong: rgba(15, 23, 42, 0.18);
  --cyber-cyan: #55759b;
  --cyber-purple: #5b7c99;
  --cyber-blue: #44648c;
  --cyber-ok: #16a34a;
  --cyber-danger: #ef4444;
  --cyber-warn: #d97706;
  --cyber-txt: #27313e;
  --cyber-dim: #64748b;
  --cyber-faint: #94a3b8;
  --cyber-acc-soft: rgba(85, 117, 155, 0.12);
  --cyber-mono: 'JetBrains Mono', Consolas, 'Courier New', monospace;
}

/* 定位兜底：面板高度按视口算（100vh - 52px），靠 EP 的「overlay fixed + align-center flex」居中时，
   只要外层容器带了 transform / contain / backdrop-filter（写作页编辑器就有 backdrop-filter），
   fixed 的包含块就会被改成那个容器，面板随之掉到文档下方，只能滚动找回来。
   这里显式钉住 overlay 满屏 + margin:auto 居中：空间够就居中，不够时 auto 归零，从顶部开始正常滚动。 */
.generator-dialog-modal {
  position: fixed !important;
  inset: 0 !important;
  padding: 26px 0 !important;
  z-index: 2000;
}

.generator-dialog-modal .el-overlay-dialog {
  position: static !important;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  overflow: auto;
}

.generator-dialog-modal .el-dialog {
  margin: auto !important;
}

.generator-dialog-modal .el-dialog {
  width: calc(100vw - 128px);
  max-width: 1152px;
  height: calc(100vh - 52px);
  max-height: calc(100vh - 52px);
  background: var(--cyber-bg);
  border: 1px solid #e2e8f0;
  border-radius: 20px;
  overflow: hidden;
  padding: 0;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.6) inset,
    rgba(55, 65, 81, 0.12) 0 8px 24px 0,
    rgba(55, 65, 81, 0.04) 0 2px 6px 0;
  display: flex;
  flex-direction: column;
}

.generator-dialog-modal .el-dialog__header {
  margin: 0;
  padding: 0;
  flex-shrink: 0;
}
.generator-dialog-modal .el-dialog__body {
  padding: 0;
  flex: 1;
  min-height: 0;
  display: flex;
}
.generator-dialog-modal .el-dialog__footer {
  padding: 0;
  flex-shrink: 0;
}

.gen-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 22px 12px;
  gap: 12px;
}
.gen-dialog-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 700;
  color: var(--cyber-txt);
  min-width: 0;
}
.gen-title-icon {
  color: var(--cyber-cyan);
  font-size: 18px;
  flex: none;
}
.title-text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.gen-dialog-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
.gen-close-btn {
  --el-color-primary: var(--cyber-dim);
  color: var(--cyber-dim) !important;
  font-size: 14px !important;
}
.gen-close-btn:hover {
  color: var(--cyber-cyan) !important;
}

/* 双栏 */
.dialog-content {
  display: flex;
  height: 100%;
  min-height: 0;
}
.prompt-template-area {
  width: 445px;
  flex-shrink: 0;
  padding: 8px 10px;
  border-right: 1px solid var(--cyber-line);
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.template-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  padding: 4px 8px 10px;
}
.area-label {
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.5px;
  color: var(--cyber-txt);
}
.area-hint {
  font-size: 11px;
  color: var(--cyber-faint);
}
.template-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.template-item {
  position: relative;
  height: 36px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 12px 0 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s var(--xm-ease, ease);
}
.template-item:hover {
  background: var(--cyber-acc-soft);
}
.template-item.active {
  background: var(--cyber-acc-soft);
}
.template-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 6px;
  bottom: 6px;
  width: 3px;
  background: var(--cyber-cyan);
  border-radius: 0;
}
.template-item-detail {
  height: 22px;
  padding: 0 10px !important;
  font-size: 12px;
  color: #f0a020 !important;
  background: #f8fafc !important;
  border: none !important;
  border-radius: 22px !important;
  box-shadow: none !important;
  flex: none;
}
.template-item-detail:hover {
  color: #fff !important;
  background: #f0a020 !important;
}
.template-item-name {
  font-size: 13px;
  color: var(--cyber-dim);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}
.template-item-pack {
  flex: none;
  max-width: 84px;
  padding: 0 7px;
  border-radius: 999px;
  background: var(--cyber-panel2);
  color: var(--cyber-faint);
  font-size: 11px;
  line-height: 18px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.template-item.active .template-item-name {
  color: var(--cyber-txt);
  font-weight: 600;
}
.template-item-remove {
  margin-left: auto;
  flex: none;
  color: var(--cyber-faint);
  background: none;
  border: none;
  cursor: pointer;
  font-size: 13px;
  line-height: 1;
  opacity: 0.65;
  padding: 2px 5px;
  border-radius: 6px;
  transition: color 0.2s var(--xm-ease, ease), opacity 0.2s var(--xm-ease, ease);
}
.template-item-remove:hover {
  color: var(--cyber-danger);
  opacity: 1;
}
.empty-template {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--cyber-faint);
  text-align: center;
  padding: 16px;
}
.template-footer {
  flex-shrink: 0;
  padding: 8px 4px 2px;
}
.more-prompt-btn {
  width: 100%;
  height: 34px;
  border-radius: 3px;
  background: linear-gradient(135deg, #55759b 0%, #44648c 100%) !important;
  color: #fff !important;
  border: none !important;
}

/* 右栏 */
.form-area {
  width: 668px;
  flex-shrink: 0;
  padding: 8px 14px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  overflow-y: auto;
}
.form-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-block-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--cyber-txt);
  letter-spacing: 0.3px;
}
.prompt-mode {
  margin-bottom: 2px;
}
.quick-prompt {
  border: 1px dashed var(--cyber-line-strong);
  border-radius: 8px;
  padding: 10px 12px;
}
.quick-prompt-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--cyber-txt);
}
.quick-prompt-tip {
  font-size: 11px;
  color: var(--cyber-faint);
  margin-top: 4px;
}

/* 结果视图 */
.result-step {
  flex-direction: column;
}
.result-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  border-bottom: 1px solid var(--cyber-line);
}
.result-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 700;
  color: var(--cyber-txt);
}
.result-title .el-icon {
  color: var(--cyber-cyan);
}
.result-actions {
  display: flex;
  gap: 4px;
}
.result-actions .el-button.is-link {
  color: var(--cyber-dim);
}
.result-actions .el-button.is-link:hover {
  color: var(--cyber-cyan);
}
.result-scrollbar {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 18px 24px 28px;
}
.result-scrollbar .markdown-body {
  color: var(--cyber-txt);
  line-height: 1.75;
  font-size: 14px;
}
.result-scrollbar .markdown-body a {
  color: var(--cyber-cyan);
}
.generating-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--cyber-faint);
  font-size: 13px;
  padding: 12px 0;
}

/* 底部 */
.gen-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 22px 16px;
  border-top: 1px solid var(--cyber-line);
}
.gen-footer-disclaimer {
  font-size: 12px;
  color: var(--cyber-dim);
}
.gen-footer-actions {
  display: flex;
  gap: 10px;
}
.gen-generate-btn {
  min-width: 120px;
  border-radius: 34px;
  background: linear-gradient(135deg, #55759b 0%, #44648c 100%) !important;
  color: #fff !important;
  box-shadow: 0 2px 8px rgba(85, 117, 155, 0.3);
  font-weight: 500;
}
.gen-generate-btn:not(.is-disabled):hover {
  filter: brightness(1.06);
}
.gen-generate-btn.stop {
  background: var(--cyber-danger) !important;
}

/* 暗色霓虹（修复「照搬」Creative 时主面板暗色不跟随的问题） */
:root[data-theme='dark'] .generator-dialog-modal {
  --cyber-bg: #0a0f1e;
  --cyber-bg-soft: #0e1628;
  --cyber-panel: #0e1628;
  --cyber-panel2: #111b2f;
  --cyber-line: rgba(0, 229, 255, 0.14);
  --cyber-line-strong: rgba(0, 229, 255, 0.26);
  --cyber-cyan: #00e5ff;
  --cyber-purple: #8b7bff;
  --cyber-blue: #3b82f6;
  --cyber-txt: #e6e9f0;
  --cyber-dim: #aab3c5;
  --cyber-faint: #5a6170;
  --cyber-acc-soft: rgba(0, 229, 255, 0.12);
}
:root[data-theme='dark'] .generator-dialog-modal .el-dialog {
  background: #0a0f1e;
  border-color: rgba(124, 107, 255, 0.12);
  box-shadow:
    0 0 0 1px rgba(124, 107, 255, 0.12),
    0 0 32px rgba(0, 229, 255, 0.1),
    0 24px 80px -16px rgba(0, 0, 0, 0.85);
}
:root[data-theme='dark'] .generator-dialog-modal .gen-dialog-header {
  background: rgba(8, 11, 20, 0.72);
  border-bottom: 1px solid rgba(0, 229, 255, 0.14);
}
:root[data-theme='dark'] .generator-dialog-modal .gen-dialog-title,
:root[data-theme='dark'] .generator-dialog-modal .form-block-label,
:root[data-theme='dark'] .generator-dialog-modal .result-title {
  color: #e6e9f0;
}
:root[data-theme='dark'] .generator-dialog-modal .gen-title-icon,
:root[data-theme='dark'] .generator-dialog-modal .result-title .el-icon {
  color: #00e5ff;
}
:root[data-theme='dark'] .generator-dialog-modal .area-label {
  color: #aab3c5;
}
:root[data-theme='dark'] .generator-dialog-modal .template-item:hover {
  background: rgba(0, 229, 255, 0.05);
}
:root[data-theme='dark'] .generator-dialog-modal .template-item.active {
  background: rgba(0, 229, 255, 0.14);
}
:root[data-theme='dark'] .generator-dialog-modal .template-item.active::before {
  background: #00e5ff;
}
:root[data-theme='dark'] .generator-dialog-modal .template-item-name {
  color: #8b93a3;
}
:root[data-theme='dark'] .generator-dialog-modal .template-item.active .template-item-name {
  color: #00e5ff;
}
:root[data-theme='dark'] .generator-dialog-modal .template-item-detail {
  color: #f0a020 !important;
  background: rgba(240, 160, 32, 0.12) !important;
}
:root[data-theme='dark'] .generator-dialog-modal .template-item-remove:hover {
  color: #ff4d6d;
}
:root[data-theme='dark'] .generator-dialog-modal .more-prompt-btn {
  background: linear-gradient(135deg, #00e5ff 0%, #3b82f6 55%, #7c6bff 100%) !important;
}
:root[data-theme='dark'] .generator-dialog-modal .material-preview {
  background: rgba(10, 15, 30, 0.6);
  border-color: rgba(0, 229, 255, 0.14);
  color: #aab3c5;
}
:root[data-theme='dark'] .generator-dialog-modal .gen-dialog-footer {
  background: rgba(8, 11, 20, 0.8);
  border-top-color: rgba(0, 229, 255, 0.14);
}
:root[data-theme='dark'] .generator-dialog-modal .gen-footer-disclaimer {
  color: #8b93a3;
}
:root[data-theme='dark'] .generator-dialog-modal .gen-generate-btn {
  background: linear-gradient(135deg, #00e5ff 0%, #3b82f6 55%, #7c6bff 100%) !important;
  color: #04101a !important;
  border-radius: 8px;
  box-shadow: 0 0 18px rgba(0, 229, 255, 0.35);
  font-weight: 600;
  letter-spacing: 2px;
  font-family: var(--cyber-mono);
}
:root[data-theme='dark'] .generator-dialog-modal .gen-generate-btn:not(.is-disabled):hover {
  background: linear-gradient(135deg, #33ebff 0%, #5b93ff 55%, #9786ff 100%) !important;
  box-shadow: 0 0 26px rgba(0, 229, 255, 0.55);
}
:root[data-theme='dark'] .generator-dialog-modal .gen-generate-btn.stop {
  background: #ff4d6d !important;
}
:root[data-theme='dark'] .generator-dialog-modal .result-toolbar {
  background: rgba(8, 11, 20, 0.5);
  border-bottom-color: rgba(0, 229, 255, 0.14);
}
:root[data-theme='dark'] .generator-dialog-modal .result-actions .el-button.is-link {
  color: #8b93a3;
}
:root[data-theme='dark'] .generator-dialog-modal .result-actions .el-button.is-link:hover {
  color: #00e5ff;
}
:root[data-theme='dark'] .generator-dialog-modal .result-scrollbar .markdown-body {
  color: #c9d2e3;
}
:root[data-theme='dark'] .generator-dialog-modal .result-scrollbar .markdown-body a {
  color: #00e5ff;
}
:root[data-theme='dark'] .generator-dialog-modal .generating-hint {
  color: #8b93a3;
}

@media (max-width: 900px) {
  .generator-dialog-modal .el-dialog {
    width: calc(100vw - 32px);
  }
  .dialog-content {
    flex-direction: column;
  }
  .prompt-template-area {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--cyber-line);
    max-height: 40%;
  }
  .form-area {
    width: 100%;
  }
}

/* ---------- 兜底：write-theme.css 的 .xm-dialog 全局基类在本文件样式之后注入， ----------
   同特异性规则会覆盖本弹窗的零内边距骨架与链接按钮透明底；用更高特异性锁住，仅抵消类名追加 xm-dialog 带来的冲突 ---------- */
.generator-dialog-modal .el-dialog.xm-dialog .el-dialog__header {
  padding: 0;
  border-bottom: none;
}
.generator-dialog-modal .el-dialog.xm-dialog .el-dialog__body {
  padding: 0;
}
.generator-dialog-modal .el-dialog.xm-dialog .el-dialog__footer {
  padding: 0;
  border-top: none;
}
.generator-dialog-modal .el-dialog.xm-dialog .gen-close-btn,
.generator-dialog-modal .el-dialog.xm-dialog .result-actions .el-button.is-link {
  background: transparent;
  border-color: transparent;
}
</style>
