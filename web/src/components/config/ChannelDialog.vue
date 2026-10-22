<template>
  <el-dialog
    :model-value="visible"
    :title="isNew ? '添加渠道' : '管理渠道'"
    width="var(--app-dialog-width-form)"
    append-to-body
    :close-on-click-modal="false"
    class="cfg-dialog"
    @update:model-value="emit('update:visible', $event)"
    @closed="clearErrors"
  >
    <div class="cap-form">
      <header class="cap-header">
        <h3>{{ isNew ? '添加渠道' : '管理渠道' }}</h3>
        <p class="cap-header__sub">
          渠道 = 端点 + 鉴权 + 并发。一个渠道下可以挂多个模型，模型的请求会按渠道配置发出。
        </p>
      </header>

      <div class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">渠道名称 <span class="cap-req">*</span></label>
          <el-input v-model="form.name" placeholder="如：火山引擎" clearable />
          <p v-if="errors.name" class="cap-error">{{ errors.name }}</p>
        </div>
        <div class="cap-field">
          <label class="cap-label">渠道备注名</label>
          <el-input v-model="form.channel_name" placeholder="可选，用于区分同类型的多条渠道" clearable />
        </div>
      </div>

      <div class="cap-field cap-field--full">
        <label class="cap-label">请求地址 <span class="cap-req">*</span></label>
        <el-input v-model="form.api_url" placeholder="e.g. https://api.openai.com/v1" clearable />
        <p class="cap-hint">兼容 OpenAI 的端点不要以斜杠结尾；开启「完整 URL」后则按填写的地址原样请求。</p>
        <p v-if="errors.api_url" class="cap-error">{{ errors.api_url }}</p>
      </div>

      <div class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">请求协议 <span class="cap-req">*</span></label>
          <el-select v-model="form.api_format" style="width: 100%">
            <el-option v-for="o in PROTOCOL_OPTIONS" :key="o.value" :label="o.label" :value="o.value">
              <div class="cap-protocol-option">
                <span>{{ o.label }}</span>
                <span class="cap-protocol-hint">{{ o.hint }}</span>
              </div>
            </el-option>
          </el-select>
        </div>
        <div class="cap-field">
          <label class="cap-label">完整 URL</label>
          <div class="cap-inline">
            <el-switch v-model="form.use_full_url" />
            <span class="cap-hint-inline">开启后直接使用填写的地址</span>
          </div>
        </div>
      </div>

      <div class="cap-field cap-field--full">
        <label class="cap-label">认证方式</label>
        <el-select v-model="form.auth_type" style="width: 100%">
          <el-option label="API Key（默认）" value="api_key" />
          <el-option label="Access Key + Secret Key" value="ak_sk" />
          <el-option label="API Key + Group ID" value="group_id" />
          <el-option label="推理接入点 ID" value="endpoint_id" />
        </el-select>
      </div>

      <div class="cap-grid">
        <template v-if="form.auth_type === 'ak_sk'">
          <div class="cap-field">
            <label class="cap-label">Access Key <span class="cap-req">*</span></label>
            <el-input v-model="form.access_key" placeholder="Access Key" clearable />
          </div>
          <div class="cap-field">
            <label class="cap-label">Secret Key <span class="cap-req">*</span></label>
            <el-input v-model="form.secret_key" type="password" show-password placeholder="Secret Key" />
          </div>
        </template>
        <div v-else class="cap-field">
          <label class="cap-label">API Key <span v-if="!isLocal" class="cap-req">*</span></label>
          <el-input
            v-model="form.api_key"
            type="password"
            show-password
            :placeholder="isNew ? '输入 API 密钥' : '留空则不修改'"
          />
          <p class="cap-hint">用于该渠道的请求鉴权，保存后不可查看。</p>
        </div>

        <div class="cap-field">
          <label class="cap-label">并发数</label>
          <el-input-number v-model="form.concurrency" :min="1" :max="64" controls-position="right" style="width: 140px" />
          <p class="cap-hint">该渠道同时进行的任务上限。</p>
        </div>
      </div>

      <div v-if="form.auth_type === 'group_id'" class="cap-field">
        <label class="cap-label">Group ID</label>
        <el-input v-model="form.group_id" placeholder="Group ID（智谱等）" clearable />
      </div>
      <div v-if="form.auth_type === 'endpoint_id'" class="cap-field">
        <label class="cap-label">接入点 ID</label>
        <el-input v-model="form.endpoint_id" placeholder="推理接入点 ID（豆包等）" clearable />
      </div>

      <div class="cap-field cap-field--full">
        <label class="cap-label">自定义请求头（JSON）</label>
        <el-input
          v-model="form.extra_config"
          type="textarea"
          :rows="3"
          spellcheck="false"
          placeholder='{"headers":{"x-opencode-session":"my-session"}}'
          @input="clearErrors"
        />
        <p class="cap-hint">
          该渠道所有请求都会带上这里的头。留空即不额外发送。
          <template v-if="isOpenCodeGo">
            <br />当前地址是 <b>OpenCode Go</b>，已自动补 <code>x-opencode-session</code>（按渠道生成稳定 ID），
            通常无需手动配置；如需固定值可在上方覆盖。
          </template>
        </p>
        <p v-if="errors.extra_config" class="cap-error">{{ errors.extra_config }}</p>
      </div>

      <p v-if="errors.api_key" class="cap-error">{{ errors.api_key }}</p>
      <p v-if="errors.form" class="cap-error">{{ errors.form }}</p>
    </div>

    <template #footer>
      <div class="cap-panel__foot">
        <el-button v-if="!isNew" @click="emit('delete', form.id)">删除渠道</el-button>
        <div class="cap-panel__actions">
          <el-button @click="emit('update:visible', false)">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSave">
            {{ isNew ? '创建' : '保存' }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { providerAPI } from '@/api'
import type { ApiProvider } from '@/types'

defineOptions({ name: 'ChannelDialog' })

const props = defineProps<{
  visible: boolean
  /** 编辑中的渠道；新建时为 null */
  channel: ApiProvider | null
  isNew: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
  (e: 'delete', id: number): void
}>()

const PROTOCOL_OPTIONS = [
  { value: 'openai', label: 'OpenAI Chat Completions', hint: 'OpenAI 兼容 /v1/chat/completions，适合 DeepSeek、Qwen、硅基等兼容接口。' },
  { value: 'responses', label: 'OpenAI Responses', hint: 'OpenAI /v1/responses，支持 Responses API 的 input/reasoning 等原生字段。' },
  { value: 'claude', label: 'Anthropic Messages', hint: 'Claude /v1/messages 原生格式。' },
  { value: 'gemini', label: 'Gemini GenerateContent', hint: 'Google Gemini generateContent 原生格式。' },
  { value: 'custom_json', label: '自定义 JSON', hint: '由模型的「通用请求规则」驱动，适合图片 / 视频等非标准协议的上游。' }
]

const form = reactive({
  id: 0,
  name: '',
  channel_name: '',
  api_url: '',
  api_format: 'openai',
  use_full_url: false,
  auth_type: 'api_key',
  api_key: '',
  access_key: '',
  secret_key: '',
  group_id: '',
  endpoint_id: '',
  concurrency: 1,
  extra_config: ''
})

const errors = reactive<Record<string, string>>({})
const saving = ref(false)

const isLocal = computed(() => /localhost|127\.0\.0\.1/.test(form.api_url || ''))

/** OpenCode Go 端点：后端会自动补 x-opencode-session，这里只做提示 */
const isOpenCodeGo = computed(() => /opencode\.ai\/zen\/go/i.test(form.api_url || ''))

const clearErrors = () => {
  for (const k of Object.keys(errors)) delete errors[k]
}

/** 打开时按传入渠道初始化；编辑态不回填密钥（后端也不返回明文） */
watch(
  () => props.visible,
  v => {
    if (!v) return
    clearErrors()
    const c = props.channel
    if (props.isNew || !c) {
      Object.assign(form, {
        id: 0, name: '', channel_name: '', api_url: '', api_format: 'openai',
        use_full_url: false, auth_type: 'api_key', api_key: '',
        access_key: '', secret_key: '', group_id: '', endpoint_id: '', concurrency: 1,
        extra_config: ''
      })
      return
    }
    Object.assign(form, {
      id: c.id,
      name: c.name || '',
      channel_name: c.channel_name || '',
      api_url: c.api_url || '',
      api_format: c.api_format || 'openai',
      use_full_url: !!c.use_full_url,
      auth_type: c.auth_type || 'api_key',
      api_key: '',
      access_key: c.access_key || '',
      secret_key: c.secret_key || '',
      group_id: c.group_id || '',
      endpoint_id: c.endpoint_id || '',
      concurrency: c.concurrency || 1,
      extra_config: c.extra_config || ''
    })
  }
)

const validate = (): boolean => {
  clearErrors()
  if (!form.name.trim()) errors.name = '渠道名称不能为空'
  const url = form.api_url.trim()
  if (!url) errors.api_url = '请求地址不能为空'
  else if (!/^https?:\/\//i.test(url)) errors.api_url = '请求地址需以 http:// 或 https:// 开头'

  if (form.auth_type === 'ak_sk') {
    if (!form.access_key.trim() || !form.secret_key.trim()) {
      errors.api_key = '请填写 Access Key 与 Secret Key'
    }
  } else if (props.isNew && !isLocal.value && !form.api_key.trim()) {
    errors.api_key = 'API Key 不能为空'
  }

  // 自定义请求头必须是合法 JSON，否则后端解析会静默忽略、用户却以为生效了
  const raw = form.extra_config.trim()
  if (raw) {
    try {
      const parsed = JSON.parse(raw)
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        errors.extra_config = '自定义请求头需为 JSON 对象'
      } else if (parsed.headers !== undefined) {
        const h = parsed.headers
        if (typeof h !== 'object' || h === null || Array.isArray(h)) {
          errors.extra_config = 'headers 需为 JSON 对象'
        }
      }
    } catch {
      errors.extra_config = '自定义请求头不是合法 JSON'
    }
  }

  const bad = Object.keys(errors).length > 0
  if (bad) errors.form = '请先修正上方标红的字段'
  return !bad
}

const handleSave = async () => {
  if (!validate()) return
  saving.value = true
  try {
    const payload: Record<string, unknown> = {
      name: form.name.trim(),
      channel_name: form.channel_name.trim(),
      api_url: form.api_url.trim(),
      api_format: form.api_format,
      use_full_url: form.use_full_url,
      auth_type: form.auth_type,
      access_key: form.access_key,
      secret_key: form.secret_key,
      group_id: form.group_id,
      endpoint_id: form.endpoint_id,
      concurrency: form.concurrency,
      extra_config: form.extra_config.trim()
    }
    // 编辑态留空表示不改密钥
    if (form.api_key.trim()) payload.api_key = form.api_key.trim()

    const res: any = props.isNew
      ? await providerAPI.create({ ...payload, provider_type: 'other' } as any)
      : await providerAPI.update(form.id, payload as any)

    if (!res?.success) {
      ElMessage.error(res?.message || '保存失败')
      return
    }
    ElMessage.success(props.isNew ? '渠道已创建' : '渠道已更新')
    emit('saved')
  } catch (error) {
    ElMessage.error((error as Error)?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>
