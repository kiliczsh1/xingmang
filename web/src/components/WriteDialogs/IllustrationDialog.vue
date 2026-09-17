<template>
  <el-dialog
    :model-value="visible"
    title="生成插画"
    width="560px"
    append-to-body
    destroy-on-close
    class="xm-illust-dialog xm-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="xm-il-body">
      <div class="xm-il-field">
        <label class="xm-il-label"><span class="xm-il-req">*</span>画面描述</label>
        <el-input
          v-model="prompt"
          type="textarea"
          :rows="3"
          maxlength="1000"
          show-word-limit
          placeholder="描述想生成的画面：场景、人物、氛围…"
        />
      </div>
      <div class="xm-il-field-row">
        <div class="xm-il-field xm-il-field--half">
          <label class="xm-il-label">AI模型</label>
          <el-select v-model="modelId" size="default">
            <el-option v-for="cfg in apiConfigs" :key="cfg.id" :label="cfg.name" :value="cfg.id" />
          </el-select>
        </div>
        <div class="xm-il-field xm-il-field--half">
          <label class="xm-il-label">尺寸</label>
          <el-select v-model="size" size="default">
            <el-option label="方形 1024×1024" value="1024x1024" />
            <el-option label="竖版 768×1344" value="768x1344" />
            <el-option label="横版 1344×768" value="1344x768" />
          </el-select>
        </div>
      </div>
      <div class="xm-il-field">
        <label class="xm-il-label">画风</label>
        <el-radio-group v-model="artStyle" size="small">
          <el-radio-button value="realistic">写实</el-radio-button>
          <el-radio-button value="anime">动漫</el-radio-button>
          <el-radio-button value="ink">水墨</el-radio-button>
          <el-radio-button value="illustration">插画</el-radio-button>
        </el-radio-group>
      </div>

      <div v-if="imageUrl || generating" class="xm-il-result">
        <div v-if="generating" class="xm-il-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>生成中，约需 10~30 秒…</span>
        </div>
        <img v-else-if="imageUrl" :src="imageUrl" class="xm-il-img" alt="生成结果" />
      </div>
    </div>

    <template #footer>
      <div class="xm-il-footer">
        <div class="xm-il-footer-left">
          <template v-if="imageUrl">
            <el-button size="small" type="warning" @click="insertToChapter">插入正文</el-button>
            <el-button size="small" @click="copyUrl">复制链接</el-button>
            <el-button size="small" @click="openUrl" v-if="imageUrl">原图</el-button>
          </template>
        </div>
        <el-button type="warning" :loading="generating" :disabled="!prompt.trim() || !modelId" @click="generate">
          {{ imageUrl ? '重新生成' : '生成插画' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading } from '@element-plus/icons-vue'
import { aiAPI } from '@/api'
import type { ApiConfig } from '@/types'

defineOptions({ name: 'IllustrationDialog' })

const props = defineProps<{
  visible: boolean
  apiConfigs: ApiConfig[]
  /** 初始提示词（选中文本或本章场景摘要） */
  initialPrompt: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'insert', markdown: string): void
}>()

const prompt = ref('')
const modelId = ref<number>()
const size = ref('1024x1024')
const artStyle = ref('illustration')
const imageUrl = ref('')
const generating = ref(false)
let lastRequest: { prompt: string; size: string; style: string } | null = null

watch(() => props.visible, (val) => {
  if (!val) return
  prompt.value = props.initialPrompt
  imageUrl.value = ''
  generating.value = false
  modelId.value = props.apiConfigs.find(c => c.is_default)?.id ?? props.apiConfigs[0]?.id
})

const generate = async () => {
  if (!prompt.value.trim() || generating.value) return
  generating.value = true
  imageUrl.value = ''
  lastRequest = { prompt: prompt.value.trim(), size: size.value, style: artStyle.value }
  try {
    const res = await aiAPI.generateImage({
      messages: [{ role: 'user', content: prompt.value.trim() }],
      configId: modelId.value,
      size: size.value,
      quality: 'standard',
      style: artStyle.value
    })
    if (res.success && res.data?.url) {
      imageUrl.value = res.data.url
      ElMessage.success('插画已生成')
    } else {
      ElMessage.error(res.message || '生成失败，请检查模型配置')
    }
  } catch (error) {
    ElMessage.error((error as Error)?.message || '生成失败，请检查模型配置')
  } finally {
    generating.value = false
  }
}

const insertToChapter = () => {
  if (!imageUrl.value) return
  const alt = prompt.value.trim().slice(0, 30) || '插画'
  emit('insert', `\n\n![${alt}](${imageUrl.value})\n\n`)
  emit('update:visible', false)
}

const copyUrl = async () => {
  if (!imageUrl.value) return
  try {
    await navigator.clipboard.writeText(imageUrl.value)
    ElMessage.success('已复制图片链接')
  } catch {
    ElMessage.error('复制失败')
  }
}

const openUrl = () => {
  if (imageUrl.value) window.open(imageUrl.value, '_blank')
}
</script>

<style scoped>
.xm-il-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.xm-il-field {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.xm-il-field-row {
  display: flex;
  gap: 12px;
}

.xm-il-field--half {
  flex: 1;
}

.xm-il-label {
  font-size: 12.5px;
  font-weight: 500;
  color: var(--xm-ink, #374151);
}

.xm-il-req {
  color: #e54d42;
  margin-right: 2px;
}

.xm-il-result {
  border: 1px solid var(--xm-line, #eef0f2);
  border-radius: 10px;
  padding: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 180px;
  background: var(--xm-surface-2, #f7f8fa);
}

.xm-il-img {
  max-width: 100%;
  max-height: 320px;
  border-radius: 6px;
}

.xm-il-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12.5px;
  color: var(--xm-ink-3, #8a8f99);
}

.xm-il-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.xm-il-footer-left {
  display: flex;
  gap: 8px;
}
</style>
