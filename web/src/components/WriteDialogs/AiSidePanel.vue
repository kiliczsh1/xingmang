<template>
  <div v-if="visible" class="xm-ai-side">
    <!-- 头部：标题 + 模型 + 关闭 -->
    <div class="xm-ai-side__header">
      <span class="xm-ai-side__title">{{ tool?.title }}</span>
      <span class="xm-ai-side__model">模型: {{ modelName }}</span>
      <button type="button" class="xm-ai-side__close" @click="close">
        <el-icon><Close /></el-icon>
      </button>
    </div>

    <div class="xm-ai-side__body">
      <!-- ===== 配置态 ===== -->
      <template v-if="state === 'config'">
        <div class="xm-ai-side__adv">
          <div class="xm-ai-side__adv-head">
            <span class="xm-ai-side__adv-title">高级功能</span>
            <el-switch v-model="advanced" size="small" />
          </div>
          <p class="xm-ai-side__adv-hint">通过提供角色、词条、关联知识库等元数据，能够有效提高 AI 创作内容的质量和相关性</p>
        </div>
        <el-divider class="xm-ai-side__divider" />

        <div class="xm-ai-side__field">
          <label class="xm-ai-side__label">AI模型</label>
          <el-select v-model="modelId" size="default" class="xm-ai-side__model-select">
            <el-option v-for="cfg in apiConfigs" :key="cfg.id" :label="cfg.name" :value="cfg.id" />
          </el-select>
        </div>

        <!-- 高级功能展开区（对齐星月：角色卡 / 角色关系 / 词条卡） -->
        <template v-if="advanced">
          <div class="xm-ai-side__field">
            <label class="xm-ai-side__label">本章角色卡</label>
            <p class="xm-ai-side__field-hint">创建角色卡提升100%生成效果。本章不出场角色不要选择</p>
            <el-select
              v-model="advancedCharacterIds"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="请选择角色…"
            >
              <el-option v-for="c in availableCharacters" :key="c.id" :label="c.name" :value="c.id" />
            </el-select>
          </div>

          <div class="xm-ai-side__field">
            <label class="xm-ai-side__label">角色关系</label>
            <p class="xm-ai-side__field-hint">可写临时/非核心角色，或者补充人物关系</p>
            <el-input
              v-model="characterRelation"
              type="textarea"
              :rows="2"
              placeholder="例如：陈默是老槐树守夜人的孙子…"
            />
          </div>

          <div class="xm-ai-side__field">
            <label class="xm-ai-side__label">本章词条卡</label>
            <p class="xm-ai-side__field-hint">本章未使用词条不要选择，词条太多可能会严重降低AI生成质量</p>
            <el-select
              v-model="advancedTerms"
              multiple
              filterable
              collapse-tags
              collapse-tags-tooltip
              placeholder="请选择词条…"
            >
              <el-option v-for="t in terms" :key="t" :label="t" :value="t" />
            </el-select>
          </div>
        </template>

        <div v-if="sourceLabel" class="xm-ai-side__field">
          <label class="xm-ai-side__label"><span class="xm-ai-side__req">*</span>{{ sourceLabel }}</label>
          <el-input
            :model-value="sourceText"
            type="textarea"
            :rows="4"
            readonly
            resize="vertical"
          />
          <span class="xm-ai-side__count">{{ sourceLength }}</span>
        </div>

        <div class="xm-ai-side__field">
          <label class="xm-ai-side__label"><span class="xm-ai-side__req">*</span>{{ tool?.requirementLabel }}</label>
          <el-radio-group v-model="requirementMode" size="small" class="xm-ai-side__radio-group">
            <el-radio-button value="preset">快捷选项</el-radio-button>
            <el-radio-button value="custom">自定义</el-radio-button>
            <el-radio-button value="more">更多</el-radio-button>
          </el-radio-group>

          <template v-if="requirementMode !== 'custom'">
            <el-select
              v-model="presetIndex"
              class="xm-ai-side__preset-select"
              placeholder="请选择预设"
            >
              <el-option
                v-for="(p, i) in tool?.presets || []"
                :key="i"
                :label="p.name"
                :value="i"
              />
            </el-select>
            <p v-if="currentPreset?.intro" class="xm-ai-side__preset-intro">
              {{ currentPreset.intro }}
              <span class="xm-ai-side__intro-link" @click="introVisible = !introVisible">{{ introVisible ? '收起介绍' : '查看介绍' }}</span>
            </p>
            <p v-if="introVisible && currentPreset" class="xm-ai-side__preset-detail">{{ currentPreset.content }}</p>
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
      </template>

      <!-- ===== 生成中 ===== -->
      <template v-else-if="state === 'streaming'">
        <div class="xm-ai-side__result-wrap">
          <textarea class="xm-ai-side__result-text" :value="result" readonly></textarea>
          <span class="xm-ai-side__count xm-ai-side__count--result">{{ resultLength }}</span>
        </div>
      </template>

      <!-- ===== 结果态 ===== -->
      <template v-else-if="state === 'result'">
        <div class="xm-ai-side__result-wrap">
          <textarea v-model="result" class="xm-ai-side__result-text" rows="14"></textarea>
          <span class="xm-ai-side__count xm-ai-side__count--result">{{ resultLength }}</span>
        </div>
      </template>

      <!-- ===== 继续追问 ===== -->
      <template v-else-if="state === 'followup'">
        <div class="xm-ai-side__result-wrap">
          <textarea class="xm-ai-side__result-text" :value="result" readonly></textarea>
        </div>
        <div class="xm-ai-side__field">
          <label class="xm-ai-side__label">继续追问</label>
          <el-input
            v-model="followupText"
            type="textarea"
            :rows="3"
            placeholder="告诉AI需要怎样调整上次的结果…"
          />
        </div>
      </template>
    </div>

    <!-- 底部 -->
    <div class="xm-ai-side__footer">
      <template v-if="state === 'config'">
        <span class="xm-ai-side__disclaimer">以上内容均由AI生成,仅供参考和借鉴</span>
        <button type="button" class="xm-ai-side__btn xm-ai-side__btn--start" @click="start">
          {{ tool?.startLabel || '生成' }}
        </button>
      </template>

      <template v-else-if="state === 'streaming'">
        <button type="button" class="xm-ai-side__btn xm-ai-side__btn--thinking" disabled>深度思考中…</button>
        <button type="button" class="xm-ai-side__btn xm-ai-side__btn--stop" @click="stop">停止生成</button>
      </template>

      <template v-else-if="state === 'result'">
        <span class="xm-ai-side__disclaimer">以上内容均由AI生成,仅供参考和借鉴</span>
        <div class="xm-ai-side__result-actions">
          <button type="button" class="xm-ai-side__btn xm-ai-side__btn--info" @click="state = 'followup'">继续追问</button>
          <button type="button" class="xm-ai-side__btn" :disabled="historyIndex <= 0" @click="goPrev">
            <el-icon><ArrowLeft /></el-icon>上一步
          </button>
          <button type="button" class="xm-ai-side__btn xm-ai-side__btn--danger" @click="regenerate">重新生成</button>
          <button type="button" class="xm-ai-side__btn" @click="copyResult">复制</button>
          <button type="button" class="xm-ai-side__btn xm-ai-side__btn--primary" @click="compareVisible = true">对比</button>
          <button type="button" class="xm-ai-side__btn xm-ai-side__btn--info" @click="exportResult">导出至作品</button>
          <button type="button" class="xm-ai-side__btn xm-ai-side__btn--success" @click="applyResult">替换 / 插入</button>
        </div>
      </template>

      <template v-else-if="state === 'followup'">
        <button type="button" class="xm-ai-side__btn" @click="state = 'result'">返回结果</button>
        <button type="button" class="xm-ai-side__btn xm-ai-side__btn--start" :disabled="!followupText.trim()" @click="sendFollowup">发送追问</button>
      </template>
    </div>

    <!-- 对比弹窗 -->
    <el-dialog v-model="compareVisible" title="文本对照与微调工具" width="820px" append-to-body class="xm-dialog">
      <p class="xm-ai-side__compare-hint">文本 A 为原始内容，文本 B 为 AI 结果；确认后可应用 B 覆盖 A。</p>
      <div class="xm-ai-side__compare">
        <div class="xm-ai-side__compare-col">
          <span class="xm-ai-side__compare-tag">A · 原文</span>
          <el-input :model-value="sourceText" type="textarea" :rows="14" readonly />
        </div>
        <div class="xm-ai-side__compare-col">
          <span class="xm-ai-side__compare-tag">B · AI结果</span>
          <el-input v-model="result" type="textarea" :rows="14" />
        </div>
      </div>
      <template #footer>
        <el-button @click="compareVisible = false">关闭</el-button>
        <el-button type="primary" @click="applyResult(); compareVisible = false">应用 B 到正文</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Close } from '@element-plus/icons-vue'
import type { ApiConfig, Chapter, Character } from '@/types'
import type { AiToolMeta } from '@/constants/aiTools'

defineOptions({ name: 'AiSidePanel' })

const props = defineProps<{
  apiConfigs: ApiConfig[]
  availableCharacters?: Character[]
  terms?: string[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'apply', payload: { applyMode: string; text: string }): void
  (e: 'export', payload: { title: string; text: string }): void
  (e: 'finish', payload: { toolId: string; text: string }): void
}>()

const visible = ref(false)
const state = ref<'config' | 'streaming' | 'result' | 'followup'>('config')
const tool = ref<AiToolMeta | null>(null)
const sourceText = ref('')
const sourceTitle = ref('')
const chaptersContext = ref('')
const advanced = ref(false)
const modelId = ref<number>()
const requirementMode = ref<'preset' | 'custom' | 'more'>('preset')
const presetIndex = ref(0)
const customRequirement = ref('')
const introVisible = ref(false)
const result = ref('')
const followupText = ref('')
const history = ref<string[]>([])
const historyIndex = ref(-1)
const compareVisible = ref(false)
const streaming = ref(false)
let abortController: AbortController | null = null
// 上一次请求的完整参数（重新生成/追问用）
let lastRequest: { userMessage: string } | null = null
// 弹窗直启时固定的要求/附加信息与多章节请求
const fixedRequirement = ref('')
const fixedExtra = ref('')
const multiRequests = ref<Array<{ label: string; userMessage: string }>>([])
// 高级功能：本章角色卡 / 角色关系 / 本章词条卡
const advancedCharacterIds = ref<number[]>([])
const characterRelation = ref('')
const advancedTerms = ref<string[]>([])

const modelName = computed(() =>
  props.apiConfigs.find(c => c.id === modelId.value)?.name || '未选择模型'
)
const sourceLabel = computed(() => tool.value?.sourceLabel || '')
const sourceLength = computed(() => [...sourceText.value].length)
const resultLength = computed(() => [...result.value].length)
const currentPreset = computed(() => tool.value?.presets?.[presetIndex.value])

const selectedConfig = computed(() =>
  props.apiConfigs.find(c => c.id === modelId.value) ||
  props.apiConfigs.find(c => c.is_default) ||
  props.apiConfigs[0]
)

const effectiveRequirement = () => {
  if (fixedRequirement.value) return fixedRequirement.value
  if (requirementMode.value === 'custom') return customRequirement.value.trim()
  return currentPreset.value?.content?.trim() || tool.value?.defaultRequirement || ''
}

const buildUserMessage = (overrideRequirement?: string): string => {
  const parts: string[] = []
  if (sourceText.value.trim()) {
    if (tool.value?.requireSelection) {
      parts.push(`【选中的内容】\n${sourceText.value}`)
    } else {
      parts.push(`【章节：${sourceTitle.value || '当前章节'}】\n${sourceText.value}`)
    }
  }
  if (chaptersContext.value) parts.push(chaptersContext.value)
  if (advanced.value) {
    const advNames = advancedCharacterIds.value
      .map(id => props.availableCharacters?.find(c => c.id === id)?.name)
      .filter((n): n is string => !!n)
    if (advNames.length) parts.push(`【本章角色】\n${advNames.join('、')}`)
    if (characterRelation.value.trim()) parts.push(`【角色关系】\n${characterRelation.value.trim()}`)
    if (advancedTerms.value.length) parts.push(`【本章词条】\n${advancedTerms.value.join('、')}`)
  }
  parts.push(`【要求】\n${overrideRequirement ?? effectiveRequirement() ?? tool.value?.defaultRequirement ?? ''}`)
  if (fixedExtra.value) parts.push(`【${tool.value?.extraField?.label || '补充'}】\n${fixedExtra.value}`)
  if (lastRequest && state.value === 'followup') {
    parts.push(`【上一次生成结果】\n${history.value[history.value.length - 1] || result.value}`)
  }
  return parts.join('\n\n')
}

// 供父组件打开面板（配置态或直接生成）
const open = (options: {
  tool: AiToolMeta
  sourceText?: string
  sourceTitle?: string
  chaptersContext?: string
  requirement?: string
  extra?: string
  configId?: number
  autoStart?: boolean
  multi?: Array<{ label: string; userMessage: string }>
  presetResult?: string
}) => {
  tool.value = options.tool
  sourceText.value = options.sourceText || ''
  sourceTitle.value = options.sourceTitle || ''
  chaptersContext.value = options.chaptersContext || ''
  fixedRequirement.value = options.requirement || ''
  fixedExtra.value = options.extra || ''
  multiRequests.value = options.multi || []
  visible.value = true
  state.value = 'config'
  requirementMode.value = 'preset'
  presetIndex.value = 0
  customRequirement.value = ''
  introVisible.value = false
  result.value = ''
  followupText.value = ''
  history.value = []
  historyIndex.value = -1
  lastRequest = null
  advanced.value = false
  advancedCharacterIds.value = []
  characterRelation.value = ''
  advancedTerms.value = []
  modelId.value = options.configId
    ?? props.apiConfigs.find(c => c.is_default)?.id
    ?? props.apiConfigs[0]?.id
  if (options.presetResult) {
    // 查看上次结果：直接进入结果态
    result.value = options.presetResult
    history.value = [options.presetResult]
    historyIndex.value = 0
    state.value = 'result'
  } else if (options.autoStart) {
    start()
  }
}

const close = () => {
  if (streaming.value) stop()
  visible.value = false
  emit('close')
}

const requestStream = async (userMessage: string, signal: AbortSignal): Promise<void> => {
  const config = selectedConfig.value
  if (!config) throw new Error('请先在设置中配置并启用 AI 模型')
  const response = await fetch('/api/ai/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      configId: config.id,
      systemPrompts: [tool.value?.systemPrompt || ''],
      messages: [{ role: 'user', content: userMessage }],
      temperature: config.temperature ?? 0.7,
      top_p: config.top_p ?? 0.9
    }),
    signal
  })
  if (!response.ok || !response.body) {
    throw new Error(`AI 接口请求失败（HTTP ${response.status}）`)
  }
  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''
    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed.startsWith('data:')) continue
      const payload = trimmed.slice(5).trim()
      if (!payload || payload === '[DONE]') continue
      let parsed: { content?: string; error?: string }
      try {
        parsed = JSON.parse(payload)
      } catch {
        if (payload.includes('"error"')) {
          const msg = payload
            .replace(/^\s*\{"error"\s*:\s*"/, '')
            .replace(/"?\s*\}?\s*$/, '')
            .replace(/\\"/g, '"')
          throw new Error(msg)
        }
        continue
      }
      if (parsed.error) throw new Error(parsed.error)
      if (typeof parsed.content === 'string' && parsed.content) {
        result.value += parsed.content
      }
    }
  }
}

const run = async (userMessage: string) => {
  streaming.value = true
  state.value = 'streaming'
  result.value = ''
  abortController = new AbortController()
  try {
    await requestStream(userMessage, abortController.signal)
    if (result.value) {
      history.value = history.value.slice(0, historyIndex.value + 1)
      history.value.push(result.value)
      historyIndex.value = history.value.length - 1
      state.value = 'result'
      if (tool.value) emit('finish', { toolId: tool.value.id, text: result.value })
    } else {
      ElMessage.warning('未获取到生成内容')
      state.value = 'config'
    }
  } catch (error) {
    if ((error as Error)?.name === 'AbortError') {
      ElMessage.info('已停止生成')
      if (result.value) {
        state.value = 'result'
      } else {
        state.value = 'config'
      }
    } else {
      ElMessage.error((error as Error)?.message || '生成失败，请检查模型配置')
      state.value = 'config'
    }
  } finally {
    streaming.value = false
    abortController = null
  }
}

const start = () => {
  if (!tool.value || streaming.value) return
  if (multiRequests.value.length) {
    // 拆书分章模式：逐章请求，流式拼接
    const queue = multiRequests.value
    lastRequest = { userMessage: queue[0].userMessage }
    runMulti(queue)
    return
  }
  const userMessage = buildUserMessage()
  lastRequest = { userMessage }
  run(userMessage)
}

const runMulti = async (queue: Array<{ label: string; userMessage: string }>) => {
  streaming.value = true
  state.value = 'streaming'
  result.value = ''
  abortController = new AbortController()
  try {
    for (let i = 0; i < queue.length; i++) {
      if (i > 0) result.value += '\n\n'
      result.value += `【${queue[i].label}】\n`
      await requestStream(queue[i].userMessage, abortController.signal)
    }
    history.value = [result.value]
    historyIndex.value = 0
    state.value = 'result'
    if (tool.value) emit('finish', { toolId: tool.value.id, text: result.value })
    ElMessage.success('生成完成')
  } catch (error) {
    if ((error as Error)?.name === 'AbortError') {
      ElMessage.info('已停止生成')
      if (result.value) { history.value = [result.value]; historyIndex.value = 0; state.value = 'result' }
      else state.value = 'config'
    } else {
      ElMessage.error((error as Error)?.message || '生成失败，请检查模型配置')
      state.value = 'config'
    }
  } finally {
    streaming.value = false
    abortController = null
  }
}

const regenerate = () => {
  if (!lastRequest) return
  run(lastRequest.userMessage)
}

const goPrev = () => {
  if (historyIndex.value > 0) {
    historyIndex.value -= 1
    result.value = history.value[historyIndex.value]
  }
}

const sendFollowup = () => {
  if (!followupText.value.trim() || !tool.value) return
  const prev = history.value[history.value.length - 1] || result.value
  const parts: string[] = []
  if (sourceText.value.trim()) parts.push(`【原始内容】\n${sourceText.value}`)
  parts.push(`【上一次生成结果】\n${prev}`)
  parts.push(`【调整要求】\n${followupText.value.trim()}`)
  const userMessage = parts.join('\n\n')
  lastRequest = { userMessage }
  followupText.value = ''
  run(userMessage)
}

const stop = () => {
  abortController?.abort()
  abortController = null
  streaming.value = false
}

const copyResult = async () => {
  if (!result.value) return
  try {
    await navigator.clipboard.writeText(result.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const applyResult = () => {
  if (!result.value || !tool.value) return
  emit('apply', { applyMode: tool.value.applyMode, text: result.value })
}

const exportResult = async () => {
  if (!result.value || !tool.value) return
  try {
    const { value } = await ElMessageBox.prompt(
      '将本次结果保存为作品中的一个新章节',
      '导出至作品',
      {
        confirmButtonText: '导出',
        cancelButtonText: '取消',
        inputValue: `${tool.value.title}结果`,
        inputPlaceholder: '请输入新章节标题',
        inputValidator: (v: string) => (v.trim() ? true : '标题不能为空')
      }
    )
    emit('export', { title: value.trim(), text: result.value })
  } catch {
    // 取消
  }
}

defineExpose({ open })
</script>

<style scoped>
.xm-ai-side {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--xm-surface, #ffffff);
  border-left: 1px solid var(--xm-line, #e4e2dd);
}

.xm-ai-side__header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid #eeeeea;
  flex-shrink: 0;
}

.xm-ai-side__title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: #1f2329;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-ai-side__model {
  font-size: 12px;
  color: #8a8f99;
  white-space: nowrap;
}

.xm-ai-side__close {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: 5px;
  color: #8a8f99;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.xm-ai-side__close:hover {
  background: rgba(0, 0, 0, 0.06);
  color: #374151;
}

.xm-ai-side__body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 12px;
}

.xm-ai-side__adv-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.xm-ai-side__adv-title {
  font-size: 13px;
  font-weight: 600;
  color: #374151;
}

.xm-ai-side__adv-hint {
  margin: 6px 0 0;
  font-size: 11px;
  line-height: 1.6;
  color: #a3a8b3;
}

.xm-ai-side__divider {
  margin: 10px 0;
}

.xm-ai-side__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 14px;
  position: relative;
}

.xm-ai-side__label {
  font-size: 12.5px;
  font-weight: 500;
  color: #374151;
}

.xm-ai-side__field-hint {
  margin: 0;
  font-size: 11px;
  color: #a3a8b3;
  line-height: 1.6;
}

.xm-ai-side__req {
  color: #e54d42;
  margin-right: 2px;
}

.xm-ai-side__radio-group {
  display: flex;
}

.xm-ai-side__radio-group :deep(.el-radio-button) {
  flex: 1;
}

.xm-ai-side__radio-group :deep(.el-radio-button__inner) {
  width: 100%;
}

.xm-ai-side__preset-intro {
  margin: 0;
  font-size: 11px;
  color: #a3a8b3;
  line-height: 1.6;
}

.xm-ai-side__preset-detail {
  margin: 0;
  font-size: 11px;
  color: #6b7280;
  line-height: 1.7;
  background: #f7f8fa;
  border-radius: 6px;
  padding: 6px 8px;
}

.xm-ai-side__intro-link {
  color: #4f7cf7;
  text-decoration: underline;
  cursor: pointer;
  margin-left: 4px;
}

.xm-ai-side__count {
  position: absolute;
  right: 6px;
  bottom: -14px;
  font-size: 11px;
  color: #a3a8b3;
}

.xm-ai-side__count--result {
  position: static;
  align-self: flex-end;
}

.xm-ai-side__result-wrap {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.xm-ai-side__result-text {
  width: 100%;
  min-height: 300px;
  resize: none;
  border: 1px solid #e4e2dd;
  border-radius: 8px;
  padding: 8px 10px;
  font-size: 13px;
  line-height: 1.7;
  color: #1f2329;
  background: #fafaf8;
  font-family: inherit;
}

.xm-ai-side__result-text:focus {
  outline: none;
  border-color: rgba(0, 201, 167, 0.45);
}

.xm-ai-side__footer {
  flex-shrink: 0;
  border-top: 1px solid #eeeeea;
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.xm-ai-side__disclaimer {
  font-size: 11px;
  color: #a3a8b3;
  text-align: center;
}

.xm-ai-side__result-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  justify-content: flex-end;
}

.xm-ai-side__btn {
  height: 30px;
  padding: 0 12px;
  border-radius: 6px;
  border: 1px solid #e4e2dd;
  background: #ffffff;
  color: #5a6270;
  font-size: 12.5px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  transition: all 0.12s ease;
}

.xm-ai-side__btn:hover:not(:disabled) {
  border-color: #c9c6bf;
  color: #1f2329;
}

.xm-ai-side__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.xm-ai-side__btn--start {
  background: linear-gradient(135deg, #f59e0b, #f97316);
  border-color: transparent;
  color: #ffffff;
  font-weight: 600;
  justify-content: center;
}

.xm-ai-side__btn--start:hover:not(:disabled) {
  filter: brightness(1.05);
  color: #ffffff;
}

.xm-ai-side__btn--thinking {
  background: #fff7ed;
  border-color: #fdba74;
  color: #d97706;
  justify-content: center;
}

.xm-ai-side__btn--stop {
  background: #fef2f2;
  border-color: #fca5a5;
  color: #dc2626;
  justify-content: center;
}

.xm-ai-side__btn--primary {
  background: #eef4ff;
  border-color: #b6ccf9;
  color: #4f7cf7;
}

.xm-ai-side__btn--info {
  background: #f0f4ff;
  border-color: #c3d0fb;
  color: #4f6ef7;
}

.xm-ai-side__btn--success {
  background: #ecfaf3;
  border-color: #a7e8c6;
  color: #0ca678;
}

.xm-ai-side__btn--danger {
  background: #fef2f2;
  border-color: #fca5a5;
  color: #dc2626;
}

.xm-ai-side__compare-hint {
  margin: 0 0 10px;
  font-size: 12px;
  color: #8a8f99;
}

.xm-ai-side__compare {
  display: flex;
  gap: 12px;
}

.xm-ai-side__compare-col {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.xm-ai-side__compare-tag {
  font-size: 12px;
  font-weight: 600;
  color: #374151;
}

:root[data-theme='dark'] .xm-ai-side {
  background: rgba(22, 27, 34, 0.92);
  border-left-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-ai-side__title,
:root[data-theme='dark'] .xm-ai-side__adv-title,
:root[data-theme='dark'] .xm-ai-side__label,
:root[data-theme='dark'] .xm-ai-side__result-text {
  color: #e6edf3;
}

:root[data-theme='dark'] .xm-ai-side__header,
:root[data-theme='dark'] .xm-ai-side__footer {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-ai-side__adv-hint,
:root[data-theme='dark'] .xm-ai-side__count,
:root[data-theme='dark'] .xm-ai-side__disclaimer,
:root[data-theme='dark'] .xm-ai-side__preset-intro,
:root[data-theme='dark'] .xm-ai-side__field-hint {
  color: #6b7688;
}

:root[data-theme='dark'] .xm-ai-side__compare-tag {
  color: #c9d3de;
}

:root[data-theme='dark'] .xm-ai-side__compare-hint {
  color: #6b7688;
}

:root[data-theme='dark'] .xm-ai-side__result-text {
  background: rgba(28, 35, 51, 0.5);
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-ai-side__btn {
  background: rgba(28, 35, 51, 0.6);
  border-color: #3f4a5c;
  color: #cbd5e1;
}

:root[data-theme='dark'] .xm-ai-side__preset-detail {
  background: rgba(28, 35, 51, 0.6);
  color: #8b98a8;
}

:root[data-theme='dark'] .xm-ai-side__close:hover {
  background: rgba(255, 255, 255, 0.08);
}
</style>
