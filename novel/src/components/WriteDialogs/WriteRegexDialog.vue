<template>
  <el-dialog
    v-model="visible"
    title="正文对话正则过滤"
    width="760px"
    append-to-body
  >
    <div class="regex-dialog-body">
      <div class="regex-dialog-row">
        <el-switch v-model="writeDisplayRegexEnabled" active-text="启用过滤" inactive-text="关闭过滤" />
        <el-switch v-model="writeStreamGuardEnabled" active-text="流式防闪" inactive-text="关闭防闪" />
        <el-switch v-model="writeCopyUsesFiltered" active-text="复制过滤后" inactive-text="复制原文" />
        <el-button size="small" @click="addWriteRegexRule">新增规则</el-button>
        <el-button size="small" @click="resetWriteRegexRules">重置默认</el-button>
      </div>

      <div v-if="writeRegexRules.length === 0" class="regex-empty">
        <el-empty description="暂无规则" :image-size="80" />
      </div>

      <div v-else class="regex-rule-list">
        <div v-for="(rule, idx) in writeRegexRules" :key="rule.id" class="regex-rule-item">
          <div class="regex-rule-header">
            <el-switch v-model="rule.enabled" />
            <el-switch v-model="rule.affectsActual" active-text="影响实际" inactive-text="仅视觉" />
            <el-input v-model="rule.name" size="small" placeholder="规则名称" class="regex-rule-name" />
            <el-input v-model="rule.flags" size="small" placeholder="flags" class="regex-rule-flags" />
            <el-button size="small" :disabled="idx === 0" @click="moveWriteRegexRule(idx, -1)">上移</el-button>
            <el-button size="small" :disabled="idx === writeRegexRules.length - 1" @click="moveWriteRegexRule(idx, 1)">下移</el-button>
            <el-button size="small" type="danger" @click="removeWriteRegexRule(rule.id)">删除</el-button>
          </div>
          <div class="regex-rule-fields">
            <el-input
              v-model="rule.pattern"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 6 }"
              placeholder="pattern（正则表达式）"
            />
            <el-input
              v-model="rule.replacement"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 6 }"
              placeholder="replacement（替换内容，可用 $1/$2 捕获组）"
            />
          </div>
          <div v-if="rule.lastError" class="regex-rule-error">
            {{ rule.lastError }}
          </div>
        </div>
      </div>

      <div class="regex-test">
        <div class="regex-test-title">快速测试</div>
        <el-input
          v-model="writeRegexTestInput"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 10 }"
          placeholder="输入一段文本，查看过滤结果"
        />
        <div class="regex-test-title">过滤结果</div>
        <el-input
          :model-value="writeRegexTestOutput"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 10 }"
          readonly
        />
      </div>

      <div class="regex-import">
        <div class="regex-test-title">酒馆 JSON 导入</div>
        <div
          :class="['regex-dropzone', { dragging: writeRegexDragActive }]"
          @dragover.prevent="handleWriteRegexDragOver"
          @dragleave.prevent="handleWriteRegexDragLeave"
          @drop.prevent="handleWriteRegexFileDrop"
        >
          <el-icon class="regex-dropzone-icon"><Upload /></el-icon>
          <div class="regex-dropzone-title">拖拽 JSON 文件到这里</div>
          <div class="regex-dropzone-desc">支持 SillyTavern 单对象或数组格式</div>
        </div>
        <div class="regex-dialog-row">
          <span v-if="writeRegexImportHint" class="regex-import-hint">{{ writeRegexImportHint }}</span>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'

interface RegexRule {
  id: string
  name: string
  pattern: string
  flags: string
  replacement: string
  enabled: boolean
  affectsActual: boolean
  order: number
  lastError?: string
}

const visible = defineModel<boolean>('visible')
const writeDisplayRegexEnabled = defineModel<boolean>('writeDisplayRegexEnabled')
const writeStreamGuardEnabled = defineModel<boolean>('writeStreamGuardEnabled')
const writeCopyUsesFiltered = defineModel<boolean>('writeCopyUsesFiltered')
const writeRegexRules = defineModel<RegexRule[]>('writeRegexRules')

const props = defineProps<{
  computeDisplayContent: (rawContent: string) => string
  createDefaultRules: () => RegexRule[]
}>()

const writeRegexTestInput = ref('')
const writeRegexImportHint = ref('')
const writeRegexDragActive = ref(false)

const createRegexRuleId = () => {
  const g = globalThis as any
  if (g.crypto?.randomUUID) return g.crypto.randomUUID()
  return `write_regex_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`
}

const normalizeWriteRegexRulesOrder = () => {
  writeRegexRules.value?.forEach((rule, index) => {
    rule.order = index + 1
  })
}

const addWriteRegexRule = () => {
  if (!writeRegexRules.value) return
  writeRegexRules.value.push({
    id: createRegexRuleId(),
    name: '新规则',
    pattern: '',
    flags: 'g',
    replacement: '',
    enabled: true,
    affectsActual: false,
    order: writeRegexRules.value.length + 1
  })
  normalizeWriteRegexRulesOrder()
}

const removeWriteRegexRule = (id: string) => {
  if (!writeRegexRules.value) return
  const index = writeRegexRules.value.findIndex(rule => rule.id === id)
  if (index >= 0) {
    writeRegexRules.value.splice(index, 1)
    normalizeWriteRegexRulesOrder()
  }
}

const moveWriteRegexRule = (index: number, direction: -1 | 1) => {
  if (!writeRegexRules.value) return
  const nextIndex = index + direction
  if (nextIndex < 0 || nextIndex >= writeRegexRules.value.length) return
  const moved = writeRegexRules.value.splice(index, 1)[0]
  writeRegexRules.value.splice(nextIndex, 0, moved)
  normalizeWriteRegexRulesOrder()
}

const resetWriteRegexRules = () => {
  if (!writeRegexRules.value) return
  writeRegexRules.value = props.createDefaultRules()
  normalizeWriteRegexRulesOrder()
}

const parseRegexLiteral = (input: string) => {
  const trimmed = input.trim()
  if (!trimmed) {
    return { pattern: '', flags: 'g' }
  }

  if (trimmed.startsWith('/')) {
    let escaped = false
    let endIndex = -1
    for (let i = trimmed.length - 1; i > 0; i--) {
      const char = trimmed[i]
      if (char === '/' && !escaped) {
        endIndex = i
        break
      }
      escaped = char === '\\' ? !escaped : false
    }

    if (endIndex > 0) {
      return {
        pattern: trimmed.slice(1, endIndex),
        flags: trimmed.slice(endIndex + 1) || 'g'
      }
    }
  }

  return { pattern: trimmed, flags: 'g' }
}

const mapSillyTavernRegexRule = (item: any, index: number): RegexRule => {
  const source = typeof item?.findRegex === 'string' ? item.findRegex : ''
  const parsed = parseRegexLiteral(source)
  return {
    id: typeof item?.id === 'string' && item.id ? item.id : createRegexRuleId(),
    name: typeof item?.scriptName === 'string' && item.scriptName
      ? item.scriptName
      : typeof item?.name === 'string' && item.name
        ? item.name
        : `酒馆规则 ${index + 1}`,
    pattern: parsed.pattern,
    flags: parsed.flags,
    replacement: typeof item?.replaceString === 'string'
      ? item.replaceString
      : typeof item?.replacement === 'string'
        ? item.replacement
        : '',
    enabled: typeof item?.disabled === 'boolean'
      ? !item.disabled
      : typeof item?.enabled === 'boolean'
        ? item.enabled
        : true,
    affectsActual: typeof item?.affectsActual === 'boolean' ? item.affectsActual : false,
    order: writeRegexRules.value ? writeRegexRules.value.length + index + 1 : index + 1
  }
}

const importSillyTavernRegexJson = (rawInput: string) => {
  const raw = rawInput.trim()
  if (!raw) {
    ElMessage.warning('请先提供酒馆 JSON 内容')
    return
  }

  try {
    const parsed = JSON.parse(raw)
    const sourceList = Array.isArray(parsed) ? parsed : [parsed]
    if (sourceList.length === 0) {
      ElMessage.warning('未检测到可导入的规则')
      return
    }

    const importedRules = sourceList
      .filter(item => item && typeof item === 'object')
      .map((item, index) => mapSillyTavernRegexRule(item, index))
      .filter(rule => rule.pattern.trim())

    if (importedRules.length === 0) {
      ElMessage.warning('导入失败：未解析出有效正则规则')
      return
    }

    writeRegexRules.value?.push(...importedRules)
    normalizeWriteRegexRulesOrder()

    if (importedRules.length === 1) {
      writeRegexTestInput.value = writeRegexTestInput.value || '她咬着唇，指尖微微发白，迟迟没有说话。'
    }

    writeRegexImportHint.value = `已导入 ${importedRules.length} 条规则`
    ElMessage.success(`已导入 ${importedRules.length} 条酒馆规则`)
  } catch (error: any) {
    writeRegexImportHint.value = ''
    ElMessage.error(`导入失败：${error?.message || 'JSON 解析错误'}`)
  }
}

const handleWriteRegexDragOver = () => {
  writeRegexDragActive.value = true
}

const handleWriteRegexDragLeave = () => {
  writeRegexDragActive.value = false
}

const handleWriteRegexFileDrop = async (event: DragEvent) => {
  writeRegexDragActive.value = false
  const file = event.dataTransfer?.files?.[0]

  if (!file) {
    ElMessage.warning('未检测到文件')
    return
  }

  const isJsonFile = file.type === 'application/json' || file.name.toLowerCase().endsWith('.json')
  if (!isJsonFile) {
    ElMessage.warning('请拖入 JSON 文件')
    return
  }

  try {
    const text = await file.text()
    importSillyTavernRegexJson(text)
    if (writeRegexImportHint.value) {
      writeRegexImportHint.value = `${writeRegexImportHint.value}：${file.name}`
    }
  } catch (error: any) {
    writeRegexImportHint.value = ''
    ElMessage.error(`读取文件失败：${error?.message || '未知错误'}`)
  }
}

const writeRegexTestOutput = computed(() => {
  return props.computeDisplayContent(writeRegexTestInput.value)
})
</script>

<style scoped>
.regex-dialog-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.regex-dialog-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.regex-empty {
  padding: 20px 0;
}

.regex-rule-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 42vh;
  overflow-y: auto;
}

.regex-rule-item {
  padding: 14px;
  border-radius: 12px;
  border: 1px solid rgba(0, 169, 137, 0.18);
  background: rgba(255, 255, 255, 0.88);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.regex-rule-header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.regex-rule-name {
  flex: 1;
  min-width: 180px;
}

.regex-rule-flags {
  width: 100px;
}

.regex-rule-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.regex-rule-error {
  color: #d93025;
  font-size: 13px;
  line-height: 1.5;
}

.regex-test {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.regex-import {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.regex-dropzone {
  min-height: 120px;
  border: 2px dashed rgba(0, 169, 137, 0.28);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.72);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  text-align: center;
  color: #275f54;
  transition: all 0.2s ease;
}

.regex-dropzone.dragging {
  border-color: #00a989;
  background: rgba(0, 201, 167, 0.1);
  box-shadow: 0 0 0 3px rgba(0, 201, 167, 0.08);
}

.regex-dropzone-icon {
  font-size: 24px;
  color: #00a989;
}

.regex-dropzone-title {
  font-size: 15px;
  font-weight: 600;
}

.regex-dropzone-desc {
  font-size: 13px;
  color: #5b7b74;
}

.regex-test-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f3b35;
}

.regex-import-hint {
  font-size: 13px;
  color: #208068;
}

@media (max-width: 768px) {
  .regex-rule-fields {
    grid-template-columns: 1fr;
  }
}
</style>
