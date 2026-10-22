<template>
  <div v-if="modelValue" class="workflow-ai-generator-overlay" @click.self="close">
    <div class="workflow-ai-generator-dialog">
      <div class="dialog-header">
        <div class="header-left">
          <i class="fas fa-wand-magic-sparkles" />
          <span>AI 生成工作流</span>
        </div>
        <div class="header-right">
          <el-button text @click="close"><i class="fas fa-times" /></el-button>
        </div>
      </div>

      <div class="dialog-body">
        <div v-if="!messages.length" class="welcome-screen">
          <div class="welcome-icon"><i class="fas fa-wand-magic-sparkles" /></div>
          <div class="welcome-title">描述你想要的工作流</div>
          <div class="welcome-desc">
            例如「把一章大纲扩写成正文，循环 3 次，把结果合并成一个 Markdown 文件」。
            生成后可在右侧对比并应用。
          </div>
          <div class="quick-examples">
            <div class="examples-tags">
              <el-tag
                v-for="ex in QUICK_EXAMPLES"
                :key="ex"
                class="example-tag"
                effect="plain"
                @click="requirement = ex"
              >
                {{ ex }}
              </el-tag>
            </div>
          </div>
        </div>

        <div v-else class="messages-list">
          <div v-for="(m, i) in messages" :key="i" class="message-item" :class="`message-${m.role}`">
            <div class="message-avatar">
              <i :class="m.role === 'assistant' ? 'fas fa-robot' : 'fas fa-user'" />
            </div>
            <div class="message-body">
              <div class="message-header">
                <span class="message-role">{{ m.role === 'assistant' ? '工作流助手' : '我' }}</span>
              </div>
              <pre v-if="m.isJson" class="code-block">{{ m.content }}</pre>
              <div v-else class="message-content">{{ m.content }}</div>
            </div>
          </div>
          <div v-if="generating" class="message-item message-assistant">
            <div class="message-body">
              <div class="loading-indicator">
                <span class="loading-dots"><span class="dots"><i /><i /><i /></span></span>
                生成中...
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="dialog-footer">
        <div class="action-bar">
          <div class="action-left">
            <el-select v-model="configId" size="small" placeholder="选择模型" style="width: 220px">
              <el-option-group v-for="g in groups" :key="g.provider" :label="g.provider">
                <el-option v-for="m in g.items" :key="m.id" :label="m.label" :value="m.id" />
              </el-option-group>
            </el-select>
          </div>
          <div class="input-area">
            <el-input
              v-model="requirement"
              type="textarea"
              :rows="2"
              placeholder="描述你想要的工作流，例如：循环 3 章改稿并合并输出"
              @keydown.enter.exact.prevent="send"
            />
          </div>
          <div class="action-right">
            <el-button size="small" @click="close">取消</el-button>
            <el-button size="small" type="primary" :loading="generating" :disabled="!requirement.trim() || !configId" @click="send">
              <i class="fas fa-paper-plane" /> 生成
            </el-button>
          </div>
        </div>
        <div v-if="generatedJson" class="apply-bar">
          <span class="apply-tip">已生成工作流 JSON，可应用当前修改（覆盖画布）</span>
          <div class="apply-actions">
            <el-button size="small" type="success" :disabled="generating" @click="$emit('apply', generatedJson)">
              <i class="fas fa-check" /> 应用修改
            </el-button>
            <el-button size="small" :disabled="generating" @click="discard">取消</el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { chatStream } from '@/services/workflow/aiBridge'
import { loadModels, modelDisplayName } from '@/services/workflow/modelResolve'
import type { ApiModel } from '@/types'

const props = defineProps<{ modelValue: boolean; existing?: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'apply', json: string): void }>()

const requirement = ref('')
const configId = ref<number | undefined>()
const generating = ref(false)
const generatedJson = ref('')
const messages = ref<Array<{ role: 'user' | 'assistant'; content: string; isJson?: boolean }>>([])
const models = ref<ApiModel[]>([])

const QUICK_EXAMPLES = [
  '把一章大纲扩写成 500 字正文，循环处理 chapterPlans 里的每一章',
  '开始 → 文本模板 → 合并 → 结束，输出一个 Markdown 文件',
  'AI 生成 3 个标题，解析成数组，再逐个扩写'
]

/** 源版 AI 生成器的系统提示词（节选自 index-C01Kt0qb.js，节点字段按星芒执行器口径收敛） */
const SYSTEM_PROMPT = `你是工作流生成助手，根据用户需求生成或修改工作流 JSON 配置。

# 核心概念
工作流通过节点（nodes）和连接（edges）组成，节点间通过变量传递数据，下游节点使用 {{变量名}} 引用上游输出。

# 节点类型与配置
## start - 开始节点
data.config: { "initialVariables": { "变量名": "值" } }
输出：initialVariables 里的所有变量
## aiProcess - AI 处理节点
data.config: { "modelId": "", "configId": 0, "systemPrompt": "可直接写文本", "userPrompt": "提示词，用 {{变量名}} 引用", "contextVariables": [{"variableName":"上游变量"}], "contextMaxItems": 3, "temperature": 0.7, "maxTokens": 4000, "outputVariable": "自定义输出变量名（可选）" }
输出：aiResult、lastAIOutput，配置了 outputVariable 时额外输出该变量
## text - 文本节点
data.config: { "text": "文本内容，支持 {{变量}} 与 {{#if 变量}}...{{/if}}", "outputVariable": "textResult" }
输出：text、lastOutput
## merge - 合并节点
data.config: { "mergeType": "concat|array|object", "inputVariables": ["变量名"], "separator": "\\n\\n" }
输出：merged、lastOutput
## loop - 循环节点
data.config: { "loopType": "array|fixed|condition", "arraySource": {"variableName":"数组变量"}, "fixedCount": 5, "maxIterations": 100, "historyWindow": 3, "keepContext": true }
输出：loopResults、iterationCount、currentIteration、currentItem
循环体从 sourceHandle 为 "loop-body" 的连线进入；循环完成后走 "loop-done"。
## context - 上下文节点
data.config: { "action": "save|append|parseJson|pick|load|clear|saveChapter", "variableName": "目标变量", "inputVariables": [{"variableName":"源变量"}], "maxHistory": 20, "includeTitle": true }
## end - 结束节点
data.config: { "outputFormat": "md|txt|json", "includeOutputs": ["要汇总的变量"], "separator": "\\n\\n---\\n\\n", "autoSave": true }

# 节点结构规范
每个节点必须包含：id（如 "start-1"）、type、position {x,y}、data { label, config }
每条连接必须包含：id（如 "e1"）、source、target、sourceHandle（loop 出口必须写 "loop-body" 或 "loop-done"，其他节点写 "default"）

# 输出要求
只输出一个 JSON 对象，形如 {"name":"...","description":"...","config":{"globalVariables":{}},"nodes":[...],"edges":[...]}，不要任何解释文字或代码块标记。`

const groups = computed(() => {
  const map = new Map<string, Array<{ id: number; label: string }>>()
  models.value
    .filter(m => m.enabled !== 0)
    .forEach(m => {
      const provider = m.provider_name || '未分组'
      if (!map.has(provider)) map.set(provider, [])
      map.get(provider)!.push({ id: m.id, label: modelDisplayName(m) })
    })
  return Array.from(map.entries()).map(([provider, items]) => ({ provider, items }))
})

onMounted(async () => {
  try {
    models.value = await loadModels()
    if (!configId.value) {
      const def = models.value.find(m => m.enabled !== 0 && m.is_default === 1) || models.value.find(m => m.enabled !== 0)
      configId.value = def?.id
    }
  } catch {
    models.value = []
  }
})

const close = () => emit('update:modelValue', false)

const extractJson = (text: string): string | null => {
  const fence = text.match(/```json\s*([\s\S]*?)\s*```/i)
  if (fence) return fence[1].trim()
  const brace = text.indexOf('{')
  const last = text.lastIndexOf('}')
  if (brace !== -1 && last > brace) return text.substring(brace, last + 1)
  return null
}

const send = async () => {
  const req = requirement.value.trim()
  if (!req || generating.value) return
  if (!configId.value) {
    ElMessage.warning('请先选择模型')
    return
  }
  messages.value.push({ role: 'user', content: req })
  requirement.value = ''
  generating.value = true
  generatedJson.value = ''

  const userMsg = props.existing
    ? `当前工作流 JSON：\n${props.existing}\n\n需求：${req}`
    : req

  let acc = ''
  try {
    await chatStream({
      configId: configId.value,
      messages: [
        { role: 'system', content: SYSTEM_PROMPT },
        { role: 'user', content: userMsg }
      ],
      maxTokens: 8000,
      onDelta: full => {
        acc = full
        const last = messages.value[messages.value.length - 1]
        if (last && last.role === 'assistant') last.content = full
        else messages.value.push({ role: 'assistant', content: full })
      }
    })
    const json = extractJson(acc)
    if (!json) {
      messages.value.push({ role: 'assistant', content: '⚠️ 没有解析到 JSON，请换个说法再试一次。' })
    } else {
      try {
        JSON.parse(json)
        generatedJson.value = json
        const last = messages.value[messages.value.length - 1]
        if (last && last.role === 'assistant') {
          last.content = json
          last.isJson = true
        }
      } catch {
        messages.value.push({ role: 'assistant', content: '⚠️ 生成的 JSON 不合法，已放弃。' })
      }
    }
  } catch (e: any) {
    messages.value.push({ role: 'assistant', content: `❌ 生成失败：${e?.message || e}` })
  } finally {
    generating.value = false
  }
}

const discard = () => {
  generatedJson.value = ''
  messages.value.push({ role: 'assistant', content: '已取消本次修改。' })
}
</script>

<style scoped>
/* 源版 .workflow-ai-generator-overlay / -dialog（scope 095d5caa）*/
.workflow-ai-generator-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10000;
  backdrop-filter: blur(8px);
  animation: wf-fade-in 0.2s ease;
}
.workflow-ai-generator-dialog {
  background: var(--wf-el-bg, #1a1a1a);
  border: 1px solid var(--wf-el-border, #333);
  border-radius: 16px;
  width: 90%;
  max-width: 1200px;
  height: 85vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  animation: wf-slide-up 0.3s ease;
  overflow: hidden;
}
@keyframes wf-fade-in {
  from {
    opacity: 0;
  }
}
@keyframes wf-slide-up {
  from {
    transform: translateY(24px);
    opacity: 0;
  }
}
.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--wf-el-border, #333);
  flex-shrink: 0;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--wf-el-text-primary, #e5eaf3);
}
.header-left i {
  color: var(--el-color-primary, #409eff);
}
.dialog-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}
.welcome-screen {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  text-align: center;
}
.welcome-icon i {
  font-size: 48px;
  color: var(--el-color-primary, #409eff);
  margin-bottom: 16px;
}
.welcome-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--wf-el-text-primary, #e5eaf3);
  margin-bottom: 10px;
}
.welcome-desc {
  font-size: 13px;
  color: var(--wf-el-text-secondary, #a3a6ad);
  max-width: 520px;
  line-height: 1.7;
  margin-bottom: 20px;
}
.quick-examples {
  max-width: 640px;
}
.examples-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}
.example-tag {
  cursor: pointer;
}
.messages-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 900px;
  margin: 0 auto;
}
.message-item {
  display: flex;
  gap: 12px;
}
.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--wf-el-fill, #303030);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  color: var(--el-color-primary, #409eff);
}
.message-body {
  flex: 1;
  min-width: 0;
}
.message-header {
  margin-bottom: 4px;
}
.message-role {
  font-size: 12px;
  color: var(--wf-el-text-secondary, #a3a6ad);
}
.message-content {
  font-size: 13px;
  line-height: 1.7;
  color: var(--wf-el-text-regular, #cfd3dc);
  white-space: pre-wrap;
  word-break: break-word;
}
.code-block {
  margin: 0;
  padding: 12px;
  background: var(--wf-el-bg-page, #1a1a1a);
  border: 1px solid var(--wf-el-border, #333);
  border-radius: 8px;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--wf-el-text-regular, #cfd3dc);
  overflow: auto;
  max-height: 320px;
  white-space: pre-wrap;
  word-break: break-word;
}
.loading-indicator {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--wf-el-text-secondary, #a3a6ad);
}
.loading-dots .dots {
  display: inline-flex;
  gap: 4px;
}
.loading-dots .dots i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--el-color-primary, #409eff);
  animation: wf-dot 1.2s infinite ease-in-out;
}
.loading-dots .dots i:nth-child(2) {
  animation-delay: 0.2s;
}
.loading-dots .dots i:nth-child(3) {
  animation-delay: 0.4s;
}
@keyframes wf-dot {
  0%,
  60%,
  100% {
    opacity: 0.3;
    transform: translateY(0);
  }
  30% {
    opacity: 1;
    transform: translateY(-3px);
  }
}
.dialog-footer {
  flex-shrink: 0;
  border-top: 1px solid var(--wf-el-border, #333);
  padding: 12px 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.action-bar {
  display: flex;
  align-items: flex-end;
  gap: 12px;
}
.action-left {
  flex-shrink: 0;
}
.input-area {
  flex: 1;
  min-width: 0;
}
.action-right {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}
.apply-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding-top: 10px;
  border-top: 1px dashed var(--wf-el-border, #333);
}
.apply-tip {
  font-size: 12px;
  color: var(--wf-el-text-secondary, #a3a6ad);
}
.apply-actions {
  display: flex;
  gap: 8px;
}
</style>
