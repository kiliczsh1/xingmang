<template>
  <!-- AI 拆书分析 · 生成器式弹窗（照搬 Creative.vue 的 generator-dialog 骨架与 Cyber 皮肤） -->
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
          <span class="title-text">{{ meta.label }} ·《{{ book?.title || '—' }}》</span>
        </div>
        <div class="gen-dialog-header-right">
          <span v-if="book?.status" class="book-status">{{ book.status }}</span>
          <el-button class="gen-close-btn" type="primary" link @click="visible = false">
            <el-icon><Close /></el-icon>
          </el-button>
        </div>
      </div>
    </template>

    <!-- ============ 配置视图 ============ -->
    <div v-if="dialogStep === 'config'" class="dialog-content">
      <!-- 左栏：固定提示词（初始为空，从卡包选取后固定，最多 12 条，超出滚动） -->
      <div class="prompt-template-area">
        <div class="template-header">
          <span class="area-label">固定提示词</span>
          <span class="area-hint">卡包：{{ packName }} · 最多 12 条</span>
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
            <button class="template-item-remove" type="button" :aria-label="`移除 ${p.name}`" @click.stop="removePinned(p.id)">✕</button>
          </div>
        </div>
        <div v-else class="empty-template">
          <template v-if="promptPool.length">尚未固定提示词，点击下方「更多提示词」从「{{ packName }}」卡包选取（默认使用内置话术）</template>
          <template v-else>「{{ packName }}」卡包暂无提示词，可在提示词库该卡包中添加；生成将使用内置话术</template>
        </div>
        <div class="template-footer">
          <el-button class="more-prompt-btn" size="large" @click="openPromptSelect">更多提示词（从卡包选取）</el-button>
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
          <div class="form-block-label">补充信息</div>
          <el-input v-model="additionalInfo" type="textarea" :rows="3" maxlength="500" placeholder="可选，补充你想让 AI 关注的要点" />
        </div>

        <div class="form-block">
          <div class="form-block-label">生成素材</div>
          <pre class="material-preview">{{ userPromptPreview }}</pre>
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
            <el-button class="gen-generate-btn" type="success" round :disabled="!canGenerate" @click="generateContent">生成</el-button>
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

  <!-- 选卡弹窗：从卡包（分类）多选提示词并固定 -->
  <PromptSelectDialog
    v-model:visible="promptSelectVisible"
    v-model:model-value="pinSelection"
    :prompts="promptPool"
    :title="`选择提示词 · ${packName}`"
    mode="multi"
    modal
    :show-categories="false"
    :show-footer="true"
    @confirm="applyPinned"
  />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, Close, Loading } from '@element-plus/icons-vue'
import { configAPI, promptAPI } from '@/api'
import type { RankBook } from '@/api'
import type { ApiModel, Prompt } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import PromptSelectDialog from '@/components/WriteDialogs/PromptSelectDialog.vue'

type ActionType = 'title' | 'synopsis' | 'brainstorm' | 'analyze'

const props = defineProps<{
  modelValue: boolean
  book: RankBook | null
  actionType: ActionType
  categoryLabel?: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()
const visible = computed({ get: () => props.modelValue, set: v => emit('update:modelValue', v) })

/* 各动作的元信息与内置默认 system 话术 */
const ACTION_META: Record<ActionType, { label: string; system: string }> = {
  title: {
    label: 'AI 标题仿写',
    system:
      '你是一位网文金牌策划。根据给出书的书名、作者、题材、状态、字数与简介，产出 10 个更具吸引力、符合题材调性的备选书名。每个书名一行，并简要说明它的卖点。',
  },
  synopsis: {
    label: 'AI 简介仿写',
    system: '你是一位网文简介编辑。根据书名、题材、状态、字数与简介，重写一段更抓人、更完整、能勾起阅读欲的作品简介，控制在 150 字以内，分点突出核心设定与爽点。',
  },
  brainstorm: {
    label: 'AI 脑洞仿写',
    system: '你是一位网文脑洞策划。基于题材和现有简介，头脑风暴若干高概念设定、核心冲突或反转剧情，给出 6-8 个方向，每个用 2-3 句说明其看点与可行性。',
  },
  analyze: {
    label: 'AI 一键拆书',
    system: '你是一位资深网文编辑与拆书专家。根据书名、题材、状态、字数与简介，拆解这本书的定位、核心卖点、节奏结构、人物亮点、爽点分布与可借鉴之处，给出可落地的分析结论。',
  },
}

const dialogStep = ref<'config' | 'result'>('config')
const models = ref<ApiModel[]>([])
const selectedModel = ref<number | null>(null)
const allPrompts = ref<Prompt[]>([])
const pinnedPrompts = ref<Prompt[]>([])
const promptSelectVisible = ref(false)
const pinSelection = ref<number[]>([])
const selectedPromptId = ref<number | null>(null)
const promptMode = ref<'quick' | 'custom'>('quick')
const customPrompt = ref('')
const additionalInfo = ref('')
const output = ref('')
const generating = ref(false)
const abortController = ref<AbortController | null>(null)
const isLoading = ref(false)

const meta = computed(() => ACTION_META[props.actionType] ?? ACTION_META.analyze)

/* 动作 → 对应卡包（卡包 = prompt.category 分组，卡包名即分类名） */
const ACTION_PACK_MAP: Record<ActionType, string> = {
  title: '标题仿写',
  synopsis: '简介仿写',
  brainstorm: '脑洞仿写',
  analyze: '一键拆书',
}
const packName = computed(() => ACTION_PACK_MAP[props.actionType])

/* 选卡弹窗只显示当前卡包内的提示词 */
const promptPool = computed(() => allPrompts.value.filter(p => (p.category || '') === packName.value))

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
  return p?.name || meta.value.label + '（默认话术）'
})

const systemPrompt = computed(() => {
  if (promptMode.value === 'custom' && customPrompt.value.trim()) return customPrompt.value
  const p = pinnedPrompts.value.find(pp => pp.id === selectedPromptId.value)
  if (p?.content) return p.content
  return meta.value.system
})

const userPromptPreview = computed(() => buildUserPrompt())

const canGenerate = computed(() => !!selectedModel.value && (promptMode.value === 'custom' ? !!customPrompt.value.trim() : true))

/* 素材拼接：榜单书无本地章节，用 RankBook 元数据 + 简介 */
function buildUserPrompt(): string {
  const b = props.book
  if (!b) return ''
  const genre = props.categoryLabel || b.category || ''
  const rows = [
    `书名：《${b.title}》`,
    `作者：${b.author || '—'}`,
    `题材：${genre}`,
    `状态：${b.status || '—'}`,
    `字数：${b.wordCount || '—'}`,
    `在读：${b.readingCount || '—'}`,
    `榜单名次：第 ${b.rank ?? '—'} 名`,
    `简介：${b.synopsis || '（暂无简介）'}`,
  ]
  if (additionalInfo.value.trim()) rows.push(`补充信息：${additionalInfo.value.trim()}`)
  return rows.join('\n')
}

function selectPrompt(id: number) {
  selectedPromptId.value = id
  promptMode.value = 'quick'
}

function showPromptDetail(p: Prompt) {
  ElMessageBox.alert(p.content, p.name, { confirmButtonText: '知道了', type: 'info' })
}

function openPromptSelect() {
  pinSelection.value = pinnedPrompts.value.map(p => p.id)
  promptSelectVisible.value = true
}

function applyPinned() {
  const byId = new Map<number, Prompt>()
  for (const p of allPrompts.value) byId.set(p.id, p)
  const next: Prompt[] = []
  for (const id of pinSelection.value) {
    const p = byId.get(id)
    if (p && !next.some(n => n.id === id)) next.push(p)
  }
  pinnedPrompts.value = next.slice(0, 12)
  if (selectedPromptId.value === null || !pinnedPrompts.value.some(p => p.id === selectedPromptId.value)) {
    selectedPromptId.value = pinnedPrompts.value[0]?.id ?? null
  }
}

function removePinned(id: number) {
  pinnedPrompts.value = pinnedPrompts.value.filter(p => p.id !== id)
  if (selectedPromptId.value === id) {
    selectedPromptId.value = pinnedPrompts.value[0]?.id ?? null
  }
}

/* 打开弹窗时复位并加载模型 / 提示词 */
watch(
  () => props.modelValue,
  async v => {
    if (!v) return
    ensurePackRegistered(packName.value)
    dialogStep.value = 'config'
    output.value = ''
    additionalInfo.value = ''
    generating.value = false
    abortController.value = null
    selectedPromptId.value = null
    promptMode.value = 'quick'
    customPrompt.value = meta.value.system
    selectedModel.value = null
    isLoading.value = true
    try {
      const [modelRes, promptRes] = await Promise.all([configAPI.getAll(), promptAPI.getAll()])
      models.value = (modelRes.data ?? []).filter(m => m.enabled !== 0)
      const def = models.value.find(m => m.is_default === 1) || models.value[0]
      selectedModel.value = def?.id ?? null
      allPrompts.value = promptRes.data ?? []
      pinnedPrompts.value = []
      selectedPromptId.value = null
      promptSelectVisible.value = false
    } catch {
      /* 模型/提示词加载失败不阻塞，仍可用默认话术与自定义 */
    } finally {
      isLoading.value = false
    }
  }
)

async function generateContent() {
  const b = props.book
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
  const relatedContent = b?.synopsis ? [{ type: 'book' as const, id: b.bookId, title: b.title, content: b.synopsis }] : []
  try {
    const res = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        messages: [{ role: 'user', content: buildUserPrompt() }],
        configId: modelData.id,
        systemPrompts: [systemPrompt.value],
        relatedContent,
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

function backToConfig() {
  dialogStep.value = 'config'
}
</script>

<style>
/* ============ AI 分析弹窗（生成器式，照搬 Creative generator-dialog 的 Cyber 皮肤） ============ */
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
.book-status {
  font-size: 11px;
  font-weight: 700;
  color: var(--cyber-ok);
  background: rgba(22, 163, 74, 0.1);
  border: 1px solid rgba(22, 163, 74, 0.35);
  padding: 2px 9px;
  border-radius: 999px;
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
.material-preview {
  font-family: var(--cyber-mono);
  font-size: 12px;
  line-height: 1.7;
  color: var(--cyber-dim);
  background: var(--cyber-panel2);
  border: 1px solid var(--cyber-line);
  border-radius: 8px;
  padding: 10px 12px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 180px;
  overflow-y: auto;
  margin: 0;
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
  color: #aab3c5;
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
