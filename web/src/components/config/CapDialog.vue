<template>
  <el-dialog
    :model-value="visible"
    :title="isNew ? '创建自定义渠道' : '编辑自定义 API'"
    :width="dialogWidth"
    append-to-body
    :close-on-click-modal="false"
    class="cfg-dialog"
    @update:model-value="emit('update:visible', $event)"
    @closed="emit('closed')"
  >
    <!-- ============ 新建：渠道表单 ============
         渠道 = 接入地址 + 鉴权。模型不在这里建 —— 建完渠道后用分组头的
         「从渠道导入」把模型批量拉进来。 -->
    <div v-if="isNew" class="cap-form">
      <header class="cap-header">
        <p class="cap-header__sub">
          渠道 = 接入地址 + 鉴权。建好后回到列表，用分组头的「从渠道导入」把模型拉进来。
        </p>
      </header>

      <!-- 从预设快速填充（只填渠道侧字段） -->
      <div class="cap-field cap-field--full">
        <label class="cap-label">
          从预设快速填充
          <span class="cap-hint-inline">（可选，仅填地址与协议，可继续修改）</span>
        </label>
        <div class="cap-preset-row">
          <el-select
            :model-value="presetKey"
            clearable
            filterable
            placeholder="选择平台预设（OpenAI、DeepSeek、豆包、Seedance…）"
            style="width: 100%"
            @update:model-value="applyPreset"
          >
            <el-option
              v-for="p in PROVIDER_PRESETS"
              :key="p.key"
              :label="p.name"
              :value="p.key"
            />
          </el-select>
        </div>
      </div>

      <div class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">显示名称</label>
          <el-input
            v-model="channelForm.channel_name"
            placeholder="可选，用于区分同类型的多条渠道"
            clearable
            @input="markDirty"
          />
        </div>
        <div class="cap-field">
          <label class="cap-label">渠道名称 <span class="cap-req">*</span></label>
          <el-input v-model="channelForm.name" placeholder="如：火山引擎" @input="markDirty" />
          <p v-if="errors.channelName" class="cap-error">{{ errors.channelName }}</p>
        </div>
      </div>

      <div class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">Base URL <span class="cap-req">*</span></label>
          <el-input
            v-model="channelForm.api_url"
            placeholder="https://api.example.com"
            clearable
            @input="markDirty"
          />
          <p v-if="errors.baseUrl" class="cap-error">{{ errors.baseUrl }}</p>
        </div>
        <div class="cap-field">
          <label class="cap-label">请求协议 <span class="cap-req">*</span></label>
          <el-select v-model="channelForm.api_format" style="width: 100%" @change="markDirty">
            <el-option v-for="o in PROTOCOL_OPTIONS" :key="o.value" :label="o.label" :value="o.value">
              <div class="cap-protocol-option">
                <span>{{ o.label }}</span>
                <span class="cap-protocol-hint">{{ o.hint }}</span>
              </div>
            </el-option>
          </el-select>
        </div>
      </div>

      <div class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">认证方式</label>
          <el-select v-model="channelForm.auth_type" style="width: 100%" @change="markDirty">
            <el-option label="API Key" value="api_key" />
            <el-option label="Access Key + Secret Key" value="ak_sk" />
            <el-option label="Group ID" value="group_id" />
            <el-option label="接入点 ID" value="endpoint_id" />
          </el-select>
        </div>
        <div class="cap-field">
          <label class="cap-label">并发数</label>
          <el-input-number
            v-model="channelForm.concurrency"
            :min="1"
            :max="99"
            controls-position="right"
            style="width: 140px"
            @change="markDirty"
          />
        </div>
      </div>

      <!-- 鉴权字段：按认证方式切换 -->
      <div v-if="channelForm.auth_type === 'ak_sk'" class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">Access Key <span class="cap-req">*</span></label>
          <el-input v-model="channelForm.access_key" placeholder="Access Key" clearable @input="markDirty" />
        </div>
        <div class="cap-field">
          <label class="cap-label">Secret Key <span class="cap-req">*</span></label>
          <el-input
            v-model="channelForm.secret_key"
            type="password"
            show-password
            placeholder="Secret Key"
            @input="markDirty"
          />
        </div>
      </div>

      <div v-else-if="channelForm.auth_type === 'group_id'" class="cap-field cap-field--full">
        <label class="cap-label">Group ID <span class="cap-req">*</span></label>
        <el-input v-model="channelForm.group_id" placeholder="Group ID（智谱等）" clearable @input="markDirty" />
      </div>

      <div v-else-if="channelForm.auth_type === 'endpoint_id'" class="cap-field cap-field--full">
        <label class="cap-label">接入点 ID <span class="cap-req">*</span></label>
        <el-input v-model="channelForm.endpoint_id" placeholder="推理接入点 ID（豆包等）" clearable @input="markDirty" />
      </div>

      <div v-else class="cap-field cap-field--full">
        <label class="cap-label">API Key <span v-if="!isLocalChannel" class="cap-req">*</span></label>
        <el-input
          v-model="channelForm.api_key"
          type="password"
          show-password
          placeholder="输入 API 密钥"
          @input="markDirty"
        />
        <p class="cap-hint">本地服务（localhost）可留空。</p>
        <p v-if="errors.apiKey" class="cap-error">{{ errors.apiKey }}</p>
      </div>

      <p v-if="errors.form" class="cap-error">{{ errors.form }}</p>
    </div>

    <!-- ============ 编辑：模型表单 ============
         只改模型本身。接入地址 / 密钥 / 协议在渠道上维护（广场 → 管理渠道）。 -->
    <div v-else-if="model" class="cap-form">
      <header class="cap-header">
        <p class="cap-header__sub">
          只改模型本身。接入地址、密钥与协议在渠道上维护（API 广场 → 管理渠道）。
        </p>
      </header>

      <div class="cap-grid">
        <div class="cap-field">
          <label class="cap-label">显示名称</label>
          <el-input v-model="model.name" placeholder="留空时使用模型 ID（展示在模型选择器中）" clearable @input="markDirty" />
        </div>
        <div class="cap-field">
          <label class="cap-label">模型 ID <span class="cap-req">*</span></label>
          <el-input
            v-model="model.model"
            :placeholder="typePlaceholder"
            clearable
            @input="onModelIdInput"
          />
          <p v-if="errors.model" class="cap-error">{{ errors.model }}</p>
        </div>
      </div>

      <!-- 参数组（仅图片 / 视频） -->
      <ParamGroupEditor v-if="showParamGroups" v-model="paramGroups" />

      <!-- 通用请求规则（视频 / 自定义 JSON 协议） -->
      <section v-if="showRequestRules" class="cap-panel">
        <header class="cap-panel__head">
          <div class="cap-panel__title">
            <h4>通用请求规则</h4>
            <p>必须包含 submit 与 poll，URL 为相对 Base URL 的路径</p>
          </div>
        </header>
        <el-input
          v-model="model.request_rules"
          type="textarea"
          :rows="10"
          class="cap-code"
          spellcheck="false"
          placeholder='{"submit":{"url":"/v1/videos","method":"POST"},"poll":{"url":"/v1/videos/{id}","method":"GET"},"response":{"task_id":"id","video_url":"data.video_url","status":"status"}}'
          @input="markDirty"
        />
      </section>

      <!-- 高级参数：默认全展开，不再折叠。
           左栏集中所有滑动条参数（拖动即改），右栏放下拉框 / 开关 / 系统提示词。 -->
      <div class="cap-advanced">
        <header class="cap-advanced__head">
          <el-icon><Setting /></el-icon>
          <h4>高级参数</h4>
          <span class="cap-advanced__tip">左栏采样参数拖动即改，右栏为模式与提示词</span>
        </header>

        <div class="cap-advanced__cols">
          <!-- 左栏：滑动条 -->
          <div class="cap-col">
            <div class="cap-col__title">采样参数</div>
            <div class="cap-sliders">
              <!-- 自适应模式下这个值**不参与请求**：由后端在请求时逐档探测（见 internal/aitokens）。
                   所以这里禁用并说明，而不是让用户以为拖它就能改输出长度。 -->
              <div class="cap-slider">
                <label class="cap-label">
                  最大 Token
                  <span class="cap-src" :class="`cap-src--${maxTokensSourceKind}`">{{ maxTokensSourceLabel }}</span>
                  <span v-if="isAutoTokensMode" class="cap-hint-inline">（自适应，由请求时探测决定）</span>
                </label>
                <el-slider
                  v-model="sliderMaxTokens"
                  :min="0"
                  :max="maxTokensLimit"
                  :step="100"
                  :disabled="isAutoTokensMode"
                  show-input
                  :show-input-controls="false"
                  @change="markDirty"
                />
                <div v-if="!isAutoTokensMode && !sliderMaxTokens" class="cap-hint-inline">
                  未设置：请求不带 max_tokens，由上游按模型自身能力决定
                </div>
              </div>
              <div class="cap-slider">
                <label class="cap-label">Temperature</label>
                <el-slider
                  v-model="sliderTemperature"
                  :min="0"
                  :max="2"
                  :step="0.1"
                  show-input
                  :show-input-controls="false"
                  @change="markDirty"
                />
              </div>
              <div class="cap-slider">
                <label class="cap-label">Top P</label>
                <el-slider
                  v-model="sliderTopP"
                  :min="0"
                  :max="1"
                  :step="0.01"
                  show-input
                  :show-input-controls="false"
                  @change="markDirty"
                />
              </div>
              <div class="cap-slider">
                <label class="cap-label">Top K</label>
                <el-slider
                  v-model="sliderTopK"
                  :min="0"
                  :max="500"
                  :step="1"
                  show-input
                  :show-input-controls="false"
                  @change="markDirty"
                />
              </div>
              <div class="cap-slider">
                <label class="cap-label">频率惩罚</label>
                <el-slider
                  v-model="sliderFrequencyPenalty"
                  :min="-2"
                  :max="2"
                  :step="0.1"
                  show-input
                  :show-input-controls="false"
                  @change="markDirty"
                />
              </div>
              <div class="cap-slider">
                <label class="cap-label">存在惩罚</label>
                <el-slider
                  v-model="sliderPresencePenalty"
                  :min="-2"
                  :max="2"
                  :step="0.1"
                  show-input
                  :show-input-controls="false"
                  @change="markDirty"
                />
              </div>
            </div>
          </div>

          <!-- 右栏：下拉框 / 开关 / 文本框 -->
          <div class="cap-col">
            <div class="cap-col__title">模式与提示词</div>

            <!-- 短字段两两并排，长文本整行，压缩右栏纵向高度 -->
            <div class="cap-pairs">
              <div class="cap-field">
                <label class="cap-label">最大 Token 模式</label>
                <el-select v-model="model.max_tokens_mode" popper-class="cfg-select-popper" style="width: 100%" @change="markDirty">
                  <el-option label="自适应（按模型能力）" value="auto" />
                  <el-option label="固定值" value="fixed" />
                </el-select>
                <p class="cap-hint">
                  {{ isAutoTokensMode
                    ? '请求时从该模型家族的上限逐档下探，被拒就降一档，直到上游接受。左侧数值不参与请求。'
                    : '始终发送左侧那个固定值；上游不接受时会直接报错，不会自动调整。' }}
                </p>
              </div>

              <div class="cap-field">
                <label class="cap-label">停止序列</label>
                <el-input v-model="model.stop" placeholder='JSON 数组，如 ["\n\n"]' @input="markDirty" />
              </div>

              <div class="cap-field cap-field--full">
                <label class="cap-label">
                  思维链
                  <span class="cap-hint-inline">（关闭时请求不附带思维链参数）</span>
                </label>
                <div class="cap-inline">
                  <el-switch v-model="thinkingModeOn" />
                  <el-select v-model="model.thinking_effort" :disabled="!thinkingModeOn" popper-class="cfg-select-popper" style="width: 100%" @change="markDirty">
                    <el-option label="默认" value="default" />
                    <el-option label="低" value="low" />
                    <el-option label="中" value="medium" />
                    <el-option label="高" value="high" />
                  </el-select>
                </div>
              </div>

              <div class="cap-field">
                <label class="cap-label">提示词角色</label>
                <el-select v-model="model.prompt_role" popper-class="cfg-select-popper" style="width: 100%" @change="markDirty">
                  <el-option label="system" value="system" />
                  <el-option label="developer" value="developer" />
                  <el-option label="user" value="user" />
                </el-select>
              </div>

              <div class="cap-field">
                <label class="cap-label">能力分类</label>
                <el-select v-model="model.category" placeholder="可选，仅用于展示" popper-class="cfg-select-popper" style="width: 100%" @change="markDirty">
                  <el-option label="未分类" value="" />
                  <el-option v-for="c in categories" :key="c.key" :label="c.name" :value="c.key">
                    <span style="display:inline-flex;align-items:center;gap:8px">
                      <span class="cat-dot" :style="{ background: c.color }"></span>{{ c.name }}
                    </span>
                  </el-option>
                </el-select>
              </div>

              <div class="cap-field">
                <label class="cap-label">系统提示词</label>
                <el-input v-model="model.prompt" type="textarea" :rows="2" placeholder="该模型专属的系统提示词（可选）" @input="markDirty" />
              </div>

              <div class="cap-field">
                <label class="cap-label">模型描述</label>
                <el-input v-model="model.description" type="textarea" :rows="2" placeholder="模型用途与能力简介" @input="markDirty" />
              </div>
            </div>

            <div class="cap-pairs">
              <div class="cap-switch-row">
                <div>
                  <div class="t">启用模型</div>
                  <div class="d">保存后全局可用</div>
                </div>
                <el-switch v-model="enabledModel" />
              </div>
              <div class="cap-switch-row">
                <div>
                  <div class="t">设为默认</div>
                  <div class="d">作为系统兜底模型</div>
                </div>
                <el-switch v-model="defaultModel" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <p v-if="errors.form" class="cap-error">{{ errors.form }}</p>
    </div>

    <template #footer>
      <div class="cap-panel__foot">
        <el-button v-if="!isNew" :loading="testing" @click="model && emit('test', model)">测试连接</el-button>
        <div class="cap-panel__actions">
          <el-button v-if="!isNew" @click="emit('delete')">删除</el-button>
          <el-button @click="emit('cancel')">取消</el-button>
          <el-button type="primary" :loading="saving" @click="handleSave">
            {{ isNew ? '创建渠道' : '保存' }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Setting } from '@element-plus/icons-vue'
import { configAPI, providerAPI } from '@/api'
import { configTypeOf } from '@/services/modelType'
import { categories, guessCategory, resolveMaxTokens } from '@/services/modelMeta'
import { PROVIDER_PRESETS, type ParamGroup } from '@/data/providerPresets'
import ParamGroupEditor from './ParamGroupEditor.vue'
import type { ApiModel } from '@/types'

defineOptions({ name: 'CapDialog' })

/**
 * 双形态弹窗：
 * - isNew=true  → 「创建自定义渠道」，只建渠道（provider），不建模型
 * - isNew=false → 「编辑自定义 API」，只改模型参数
 *
 * 拆分原因：蓝本搬迁后，「建渠道」在两处重复（这里 + API 广场的渠道弹窗）。
 * 现在这里只管建渠道，模型通过分组头的「从渠道导入」批量创建。
 */
const props = defineProps<{
  visible: boolean
  /** 编辑中的模型对象；与父层共享同一引用，字段改动直接生效。新建渠道时为 null */
  model: ApiModel | null
  isNew: boolean
  dirty: boolean
  testing: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'update:dirty', value: boolean): void
  (e: 'saved', kind: 'channel' | 'model'): void
  (e: 'test', model: ApiModel): void
  (e: 'delete'): void
  (e: 'cancel'): void
  (e: 'closed'): void
}>()

/* ---------------- 弹窗宽度 ---------------- */
/**
 * 新建（渠道表单）用 form 档（560-800px）；编辑按模型类型分档：
 * video 用 workspace（960-1440px，参数组 + 请求规则要横向铺开），其余用 content（720-1120px）。
 */
const dialogWidth = computed(() => {
  if (props.isNew) return 'var(--app-dialog-width-form)'
  return configTypeOf(props.model || { category: '' }) === 'video'
    ? 'var(--app-dialog-width-workspace)'
    : 'var(--app-dialog-width-content)'
})

/* ---------------- 协议 ---------------- */
const PROTOCOL_OPTIONS = [
  { value: 'openai', label: 'OpenAI Chat Completions', hint: 'OpenAI 兼容 /v1/chat/completions，适合 DeepSeek、Qwen、硅基等兼容接口。' },
  { value: 'responses', label: 'OpenAI Responses', hint: 'OpenAI /v1/responses，支持 Responses API 的 input/reasoning 等原生字段。' },
  { value: 'claude', label: 'Anthropic Messages', hint: 'Claude /v1/messages 原生格式，thinking 使用 Messages API 的 thinking 对象。' },
  { value: 'gemini', label: 'Gemini GenerateContent', hint: 'Google Gemini generateContent 原生格式，thinking 使用 thinkingConfig。' },
  { value: 'custom_json', label: '自定义 JSON', hint: '由模型的「通用请求规则」驱动，适合图片 / 视频等非标准协议的上游。' }
]

/* ---------------- 渠道表单（新建形态） ---------------- */
const channelForm = reactive({
  name: '',
  channel_name: '',
  api_url: '',
  api_format: 'openai',
  auth_type: 'api_key',
  api_key: '',
  access_key: '',
  secret_key: '',
  group_id: '',
  endpoint_id: '',
  concurrency: 1
})

const isLocalChannel = computed(() => /localhost|127\.0\.0\.1/.test(channelForm.api_url || ''))

const resetChannelForm = () => {
  Object.assign(channelForm, {
    name: '', channel_name: '', api_url: '', api_format: 'openai',
    auth_type: 'api_key', api_key: '', access_key: '', secret_key: '',
    group_id: '', endpoint_id: '', concurrency: 1
  })
  presetKey.value = ''
}

/* 从预设快速填充：只填渠道侧字段（地址 / 协议 / 认证方式） */
const presetKey = ref('')

const applyPreset = (key: string | undefined) => {
  presetKey.value = key || ''
  if (!key) return
  const p = PROVIDER_PRESETS.find(x => x.key === key)
  if (!p) return
  if (!channelForm.name) channelForm.name = p.name
  channelForm.api_url = p.api_url
  channelForm.api_format = p.api_format
  channelForm.auth_type = p.auth_type || 'api_key'
  markDirty()
}

/* ---------------- 模型侧（编辑形态） ---------------- */
const TYPE_PLACEHOLDER: Record<string, string> = {
  chat: '请选择或输入对话模型标识',
  image: '请选择或输入图像模型标识（如 gpt-image、seedream）',
  video: '请选择或输入视频模型标识（如 veo、seedance）',
  vision: '请选择或输入视觉模型标识',
  embed: '请选择或输入嵌入模型标识（如 bge-m3）'
}

const typePlaceholder = computed(
  () => TYPE_PLACEHOLDER[configTypeOf(props.model || { category: '' })] || '输入模型标识'
)

/** 参数组：与 model.param_groups 的 JSON 文本双向同步 */
const paramGroups = computed<ParamGroup[]>({
  get: () => {
    const raw = props.model?.param_groups
    if (!raw) return []
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  },
  set: (v: ParamGroup[]) => {
    if (!props.model) return
    props.model.param_groups = v.length ? JSON.stringify(v) : ''
    markDirty()
  }
})

/** 参数组：仅图片 / 视频模型 */
const showParamGroups = computed(() => {
  const t = configTypeOf(props.model || { category: '' })
  return t === 'image' || t === 'video'
})

/** 请求规则：视频类型，或所属渠道是自定义 JSON 协议 */
const showRequestRules = computed(
  () => configTypeOf(props.model || { category: '' }) === 'video' || props.model?.api_format === 'custom_json'
)

/**
 * 是否处于自适应 max_tokens 模式。
 *
 * ⚠ 口径：本项目存 `auto` / `fixed`，蓝本存 `auto` / `manual` —— 除 `fixed`/`manual`
 * 外的值（含空）都按自适应处理，与后端 `internal/aitokens.IsAuto` 保持一致。
 * 两边口径必须同步，否则界面禁用状态与实际请求行为会对不上。
 */
const isAutoTokensMode = computed(() => {
  const mode = String(props.model?.max_tokens_mode || '').trim().toLowerCase()
  return mode !== 'fixed' && mode !== 'manual'
})

/**
 * 「最大 Token」的来源徽标。
 *
 * 为什么不直接显示数字就完事：同一个数字可能来自官方文档、别人的默认配置、
 * 或某个下午有人随手写的一个"看起来挺安全"的数 —— 它们长得完全一样。
 * 来源未知时明说「来源未知」，不假装确定。
 */
const MAX_TOKENS_SOURCE_LABEL: Record<string, string> = {
  upstream: '来自上游',
  catalog: '来自目录',
  table: '家族估算',
  manual: '手动填写',
  probed: '探测得到'
}

const maxTokensSourceKind = computed(() => {
  const src = String(props.model?.max_tokens_source || '').trim().toLowerCase()
  return MAX_TOKENS_SOURCE_LABEL[src] ? src : 'unknown'
})

const maxTokensSourceLabel = computed(() =>
  maxTokensSourceKind.value === 'unknown'
    ? ((props.model?.max_tokens ?? 0) > 0 ? '来源未知' : '未设置')
    : MAX_TOKENS_SOURCE_LABEL[maxTokensSourceKind.value]
)

/* ---------------- 高级参数滑动条 ----------------
   一律走 computed：模型上这些值可能是 null（未设置），el-slider 只接受数字。
   getter 用兜底值渲染，用户没动过就不写回，避免把 null 悄悄变成 0。 */
const sliderMaxTokens = computed({
  // 0 = 未设置（请求不带该字段）。不再兜 2000 —— 那会凭空造出一个假值，
  // 而且偏小时会静默截断输出（见 internal/aitokens 的 omitTokens）。
  get: () => props.model?.max_tokens ?? 0,
  set: v => {
    if (!props.model) return
    props.model.max_tokens = v
    // 用户手动改了值，来源就变成「手动填写」——不能继续标着「来自上游」。
    props.model.max_tokens_source = 'manual'
  }
})

const sliderTemperature = computed({
  get: () => props.model?.temperature ?? 0.7,
  set: v => {
    if (props.model) props.model.temperature = v
  }
})

const sliderTopP = computed({
  get: () => props.model?.top_p ?? 0.9,
  set: v => {
    if (props.model) props.model.top_p = v
  }
})

const sliderTopK = computed({
  get: () => props.model?.top_k ?? 0,
  set: v => {
    if (props.model) props.model.top_k = v
  }
})

const sliderFrequencyPenalty = computed({
  get: () => props.model?.frequency_penalty ?? 0,
  set: v => {
    if (props.model) props.model.frequency_penalty = v
  }
})

const sliderPresencePenalty = computed({
  get: () => props.model?.presence_penalty ?? 0,
  set: v => {
    if (props.model) props.model.presence_penalty = v
  }
})

/* ---------------- 其它 ---------------- */
const saving = ref(false)
const errors = reactive<Record<string, string>>({})

const markDirty = () => emit('update:dirty', true)

const clearErrors = () => {
  for (const k of Object.keys(errors)) delete errors[k]
}

/** 模型 ID 变化时同步推断能力分类（仅当尚未设置） */
const onModelIdInput = () => {
  const m = props.model
  if (m && !m.category) m.category = guessCategory(m.model)
  markDirty()
}

const maxTokensLimit = computed(() => {
  const m = props.model
  if (!m) return 128000
  const guess = m.model ? resolveMaxTokens(m.model) : 128000
  return Math.max(guess, m.max_tokens || 0, 128000)
})

const enabledModel = computed({
  get: () => props.model?.enabled === 1,
  set: (v: boolean) => {
    if (props.model) {
      props.model.enabled = v ? 1 : 0
      markDirty()
    }
  }
})

const defaultModel = computed({
  get: () => props.model?.is_default === 1,
  set: (v: boolean) => {
    if (props.model) {
      props.model.is_default = v ? 1 : 0
      markDirty()
    }
  }
})

const thinkingModeOn = computed({
  get: () => props.model?.thinking_mode === 1,
  set: (v: boolean) => {
    if (props.model) {
      props.model.thinking_mode = v ? 1 : 0
      markDirty()
    }
  }
})

/** 打开弹窗时重置临时状态 */
watch(
  () => props.visible,
  v => {
    if (!v) return
    clearErrors()
    if (props.isNew) resetChannelForm()
  }
)

/* ---------------- 校验 ---------------- */
const validate = (): boolean => {
  clearErrors()

  if (props.isNew) {
    const url = channelForm.api_url.trim()
    if (!channelForm.name.trim()) errors.channelName = '渠道名称不能为空'
    if (!url) errors.baseUrl = 'Base URL 不能为空'
    else if (!/^https?:\/\//i.test(url)) errors.baseUrl = 'Base URL 需以 http:// 或 https:// 开头'

    switch (channelForm.auth_type) {
      case 'ak_sk':
        if (!channelForm.access_key.trim() || !channelForm.secret_key.trim()) {
          errors.apiKey = '请填写 Access Key 与 Secret Key'
        }
        break
      case 'group_id':
        if (!channelForm.group_id.trim()) errors.apiKey = 'Group ID 不能为空'
        break
      case 'endpoint_id':
        if (!channelForm.endpoint_id.trim()) errors.apiKey = '接入点 ID 不能为空'
        break
      default:
        if (!isLocalChannel.value && !channelForm.api_key.trim()) errors.apiKey = 'API Key 不能为空'
    }
  } else {
    const m = props.model
    if (!m) return false
    if (!m.model.trim()) errors.model = '模型 ID 不能为空'
  }

  const bad = Object.keys(errors).length > 0
  if (bad) errors.form = '请先修正上方标红的字段'
  return !bad
}

/* ---------------- 保存 ---------------- */
const handleSave = async () => {
  if (!validate()) return
  saving.value = true
  try {
    if (props.isNew) {
      const res: any = await providerAPI.create({
        name: channelForm.name.trim(),
        channel_name: channelForm.channel_name.trim(),
        provider_type: 'other',
        api_url: channelForm.api_url.trim(),
        api_format: channelForm.api_format,
        use_full_url: false,
        auth_type: channelForm.auth_type,
        api_key: channelForm.api_key,
        access_key: channelForm.access_key,
        secret_key: channelForm.secret_key,
        group_id: channelForm.group_id,
        endpoint_id: channelForm.endpoint_id,
        concurrency: channelForm.concurrency
      } as any)
      if (!res?.success) {
        ElMessage.error(res?.message || '创建渠道失败')
        return
      }
      ElMessage.success('渠道创建成功，可用「从渠道导入」拉取模型')
      emit('update:dirty', false)
      emit('saved', 'channel')
      return
    }

    const m = props.model
    if (!m) return
    const payload = {
      ...m,
      is_default: m.is_default ? 1 : 0,
      enabled: m.enabled ? 1 : 0
    }
    const res: any = await configAPI.update(m.id, payload)
    if (!res?.success) {
      ElMessage.error(res?.message || '保存失败')
      return
    }
    ElMessage.success('保存成功')
    emit('update:dirty', false)
    emit('saved', 'model')
  } catch (error) {
    ElMessage.error((error as Error)?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>
