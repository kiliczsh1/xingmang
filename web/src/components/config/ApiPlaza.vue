<template>
  <div class="plaza">
    <!-- ============ 列表：上组个人渠道 / 下组平台预制 ============ -->
    <template v-if="!current && !detailChannel">
      <div class="plaza__toolbar">
        <div>
          <h3 class="plaza__title">API 广场</h3>
          <p class="plaza__subtitle">上面是你自己的渠道，下面是平台预制；渠道建好后可直接拉取模型</p>
        </div>
        <div class="plaza__toolbar-ops">
          <el-input v-model="keyword" placeholder="搜索渠道 / 平台" clearable class="plaza__search">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <button type="button" class="cfg-btn primary" @click="openAddChannel">
            <el-icon><Plus /></el-icon>添加渠道
          </button>
        </div>
      </div>

      <div class="plaza__filter-tabs">
        <button
          type="button"
          class="plaza__filter-tab"
          :class="{ 'is-active': tag === '' }"
          @click="tag = ''"
        >全部</button>
        <button
          v-for="t in MODEL_TYPES"
          :key="t.key"
          type="button"
          class="plaza__filter-tab"
          :class="{ 'is-active': tag === t.key }"
          @click="tag = tag === t.key ? '' : t.key"
        >{{ t.name }}</button>
      </div>

      <div class="plaza__list">
        <!-- 上组：个人渠道 -->
        <div class="plaza__group">
          <div class="plaza__group-head">
            <span class="plaza__group-name">个人渠道</span>
            <span class="plaza__group-count">{{ channels.length }}</span>
          </div>

          <button
            v-for="p in channels"
            :key="p.id"
            type="button"
            class="plaza__row"
            @click="detailChannel = p"
          >
            <span class="plaza__avatar" :style="{ background: providerColor(p.provider_type), color: '#fff' }">
              {{ (p.name || '渠').slice(0, 1) }}
            </span>
            <div class="plaza__row-main">
              <div class="plaza__row-name">{{ p.name }}</div>
              <div class="plaza__row-desc">{{ p.api_url }}</div>
              <div class="plaza__row-stats">
                <span class="plaza__stat"><i>模型</i><b>{{ modelCountOf(p.id) }}</b></span>
                <span class="plaza__stat"><i>并发</i><b>{{ p.concurrency || 1 }}</b></span>
                <span class="plaza__stat"><i>协议</i><b>{{ formatLabel(p.api_format || 'openai') }}</b></span>
              </div>
            </div>
            <span class="plaza__row-arrow">›</span>
          </button>

          <div v-if="channels.length === 0" class="config-page__empty-action">
            <p>还没有个人渠道，点右上角「添加渠道」，或从下面的平台预制里引用一个</p>
          </div>
        </div>

        <!-- 下组：平台预制 -->
        <div class="plaza__group">
          <div class="plaza__group-head">
            <span class="plaza__group-name">平台预制</span>
            <span class="plaza__group-count">{{ presets.length }}</span>
          </div>

          <button
            v-for="p in presets"
            :key="p.key"
            type="button"
            class="plaza__row"
            @click="current = p"
          >
            <span class="plaza__avatar" :style="{ background: presetColor(p), color: '#fff' }">{{ p.initial }}</span>
            <div class="plaza__row-main">
              <div class="plaza__row-name">
                {{ p.name }}
                <span v-if="isConnected(p)" class="plaza__row-connected">已接入</span>
              </div>
              <div class="plaza__row-desc">{{ p.desc }}</div>
              <div class="plaza__row-stats">
                <span class="plaza__stat"><i>推荐模型</i><b>{{ p.models.length }}</b></span>
                <span class="plaza__stat"><i>能力</i><b>{{ p.tags.map(t => PRESET_TAG_LABELS[t]).join(' / ') }}</b></span>
                <span v-if="p.noKey" class="plaza__stat"><i>鉴权</i><b>免 Key</b></span>
              </div>
            </div>
            <span class="plaza__row-arrow">›</span>
          </button>

          <div v-if="presets.length === 0" class="config-page__empty-action">
            <p>没有匹配的平台，换个关键词试试</p>
          </div>
        </div>
      </div>
    </template>

    <!-- ============ 个人渠道详情 ============ -->
    <template v-else-if="detailChannel">
      <div class="plaza__detail">
        <button type="button" class="plaza__back" @click="detailChannel = null">← 返回 API 广场</button>

        <div class="plaza__detail-head">
          <span
            class="plaza__avatar plaza__avatar--lg"
            :style="{ background: providerColor(detailChannel.provider_type), color: '#fff' }"
          >{{ (detailChannel.name || '渠').slice(0, 1) }}</span>
          <div class="plaza__detail-info">
            <div class="plaza__detail-text">
              <b>{{ detailChannel.name }}</b>
              <span>{{ detailChannel.channel_name || '个人渠道' }}</span>
              <code>{{ detailChannel.api_url }}</code>
            </div>
            <div class="plaza__detail-stats">
              <span class="plaza__stat-box"><i>协议</i><b>{{ formatLabel(detailChannel.api_format || 'openai') }}</b></span>
              <span class="plaza__stat-box"><i>认证</i><b>{{ authLabel(detailChannel.auth_type || 'api_key') }}</b></span>
              <span class="plaza__stat-box"><i>并发</i><b>{{ detailChannel.concurrency || 1 }}</b></span>
              <span class="plaza__stat-box"><i>模型</i><b>{{ channelModels.length }}</b></span>
            </div>
          </div>
        </div>

        <div class="plaza__models">
          <div v-for="m in channelModels" :key="m.id" class="plaza__model">
            <span
              class="plaza__model-avatar"
              :style="{ background: typeDef(configTypeOf(m))?.color || '#64748b', color: '#fff' }"
            >{{ (typeDef(configTypeOf(m))?.name || '模').slice(0, 1) }}</span>
            <div class="plaza__model-main">
              <div class="plaza__model-name">
                {{ m.name }}
                <code class="plaza__model-id">{{ m.model }}</code>
              </div>
              <div class="plaza__model-desc">{{ m.description || '—' }}</div>
            </div>
            <div class="plaza__model-tags">
              <span class="config-page__api-tag">{{ typeLabel(m) }}</span>
            </div>
          </div>

          <div v-if="channelModels.length === 0" class="config-page__empty-action">
            <p>该渠道下还没有模型。点右下角「添加模型」，或用「从渠道导入」批量拉取</p>
          </div>
        </div>

        <div class="plaza__toolbar">
          <div class="plaza__toolbar-ops">
            <button type="button" class="cfg-btn" @click="openEditChannel">
              <el-icon><Setting /></el-icon>管理渠道
            </button>
            <button type="button" class="cfg-btn danger" @click="removeChannel">
              <el-icon><Delete /></el-icon>删除渠道
            </button>
          </div>
          <button type="button" class="cfg-btn primary" @click="emit('add-model', detailChannel.id)">
            <el-icon><Plus /></el-icon>从渠道导入
          </button>
        </div>
      </div>
    </template>

    <!-- ============ 平台预制详情 ============ -->
    <template v-else-if="current">
      <div class="plaza__detail">
        <button type="button" class="plaza__back" @click="current = null">← 返回 API 广场</button>

        <div class="plaza__detail-head">
          <span
            class="plaza__avatar plaza__avatar--lg"
            :style="{ background: presetColor(current), color: '#fff' }"
          >{{ current.initial }}</span>
          <div class="plaza__detail-info">
            <div class="plaza__detail-text">
              <b>{{ current.name }}</b>
              <span>{{ current.desc }}</span>
              <code>{{ current.api_url }}</code>
            </div>
            <div class="plaza__detail-stats">
              <span class="plaza__stat-box"><i>协议</i><b>{{ formatLabel(current.api_format) }}</b></span>
              <span class="plaza__stat-box"><i>认证</i><b>{{ authLabel(current.auth_type) }}</b></span>
              <span class="plaza__stat-box"><i>模型</i><b>{{ current.models.length }}</b></span>
            </div>
          </div>
        </div>

        <div class="plaza__models">
          <div v-for="m in current.models" :key="m.model" class="plaza__model">
            <span
              class="plaza__model-avatar"
              :style="{ background: typeDef(m.config_type)?.color || '#64748b', color: '#fff' }"
            >{{ (typeDef(m.config_type)?.name || '模').slice(0, 1) }}</span>
            <div class="plaza__model-main">
              <div class="plaza__model-name">
                {{ m.name }}
                <code class="plaza__model-id">{{ m.model }}</code>
              </div>
              <div class="plaza__model-desc">{{ m.description || '—' }}</div>
            </div>
            <div class="plaza__model-tags">
              <span v-if="m.paramGroups?.length" class="config-page__api-tag">
                参数组 {{ m.paramGroups.length }}
              </span>
              <span class="config-page__api-tag">{{ typeDef(m.config_type)?.name || '对话' }}</span>
            </div>
          </div>

          <div v-if="current.models.length === 0" class="config-page__empty-action">
            <p>该平台暂无推荐模型，可先接入服务商，再到「我的 API」手动添加模型</p>
          </div>
        </div>

        <div class="plaza__toolbar">
          <div class="plaza__toolbar-ops">
            <button type="button" class="cfg-btn" @click="openDocs">查看接入文档</button>
          </div>
          <button type="button" class="cfg-btn primary" @click="openQuote(current)">
            <el-icon><Plus /></el-icon>引用模型
          </button>
        </div>
      </div>
    </template>

    <!-- ============ 渠道弹窗 ============ -->
    <ChannelDialog
      v-model:visible="channelDialogVisible"
      :channel="editingChannel"
      :is-new="isNewChannel"
      @saved="onChannelSaved"
      @delete="onChannelDelete"
    />

    <!-- ============ 引用弹窗 ============ -->
    <el-dialog
      v-model="quoteVisible"
      :title="`引用 ${quoteTarget?.name || ''}`"
      width="var(--app-dialog-width-compact)"
      append-to-body
      :close-on-click-modal="false"
      class="cfg-dialog"
    >
      <div v-if="quoteTarget" class="quote">
        <div class="quote__header">
          <span class="quote__label">接入信息</span>
          <p class="quote__hint">确认后会在「个人渠道」里创建该渠道，并按勾选创建模型。</p>
        </div>

        <div class="quote__model-card">
          <span class="plaza__avatar" :style="{ background: presetColor(quoteTarget), color: '#fff' }">{{ quoteTarget.initial }}</span>
          <div class="quote__model-text">
            <b>{{ quoteTarget.name }}</b>
            <code>{{ quoteTarget.api_url }}</code>
          </div>
        </div>

        <div class="quote__field">
          <label class="quote__label">渠道名称 <span class="quote__req">*</span></label>
          <el-input v-model="form.name" placeholder="渠道名称" />
        </div>

        <template v-if="quoteTarget.auth_type === 'ak_sk'">
          <div class="quote__field">
            <label class="quote__label">Access Key <span class="quote__req">*</span></label>
            <el-input v-model="form.access_key" placeholder="Access Key" />
          </div>
          <div class="quote__field">
            <label class="quote__label">Secret Key <span class="quote__req">*</span></label>
            <el-input v-model="form.secret_key" type="password" placeholder="Secret Key" show-password />
          </div>
        </template>
        <div v-else-if="!quoteTarget.noKey" class="quote__field">
          <label class="quote__label">API Key <span class="quote__req">*</span></label>
          <el-input v-model="form.api_key" type="password" placeholder="输入 API 密钥" show-password />
        </div>

        <div v-if="quoteTarget.auth_type === 'group_id'" class="quote__field">
          <label class="quote__label">Group ID</label>
          <el-input v-model="form.group_id" placeholder="Group ID（智谱等）" />
        </div>
        <div v-if="quoteTarget.auth_type === 'endpoint_id'" class="quote__field">
          <label class="quote__label">接入点 ID</label>
          <el-input v-model="form.endpoint_id" placeholder="推理接入点 ID（豆包等）" />
        </div>

        <div class="quote__field">
          <label class="quote__label">渠道备注名</label>
          <el-input v-model="form.channel_name" placeholder="可选，用于区分同类型的多条渠道" />
        </div>

        <div v-if="quoteTarget.models.length" class="quote__field">
          <label class="quote__label">
            引用模型（已选 {{ selectedModels.length }} / {{ quoteTarget.models.length }}）
          </label>
          <div class="quote__models">
            <label v-for="m in quoteTarget.models" :key="m.model" class="quote__model-card quote__model-card--pick">
              <el-checkbox
                :model-value="selectedModels.includes(m.model)"
                @change="toggleModel(m.model)"
              />
              <div class="quote__model-text">
                <b>{{ m.name }}</b>
                <code>{{ m.model }}</code>
              </div>
              <span class="config-page__api-tag">{{ typeDef(m.config_type)?.name || '对话' }}</span>
            </label>
          </div>
        </div>
      </div>

      <template #footer>
        <button class="cfg-btn" @click="quoteVisible = false">取消</button>
        <button class="cfg-btn primary" :disabled="quoting" @click="submitQuote">
          {{ quoting ? '创建中…' : '确定引用' }}
        </button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Setting, Delete } from '@element-plus/icons-vue'
import { providerAPI, configAPI } from '@/api'
import { MODEL_TYPES, typeDef, configTypeOf, typeLabel } from '@/services/modelType'
import { PROVIDER_PRESETS, PRESET_TAG_LABELS, type ProviderPreset } from '@/data/providerPresets'
import ChannelDialog from './ChannelDialog.vue'
import type { ApiConfigType, ApiModel, ApiProvider } from '@/types'

defineOptions({ name: 'ApiPlaza' })

const emit = defineEmits<{
  (e: 'quoted'): void
  (e: 'changed'): void
  /** 请求父层为该渠道打开「添加模型」弹窗 */
  (e: 'add-model', providerId: number): void
}>()

const keyword = ref('')
const tag = ref<ApiConfigType | ''>('')
const current = ref<ProviderPreset | null>(null)
const detailChannel = ref<ApiProvider | null>(null)

const providers = ref<ApiProvider[]>([])
const models = ref<ApiModel[]>([])

onMounted(load)

async function load() {
  try {
    const [pRes, mRes] = await Promise.all([providerAPI.getAll(), configAPI.getAll()])
    if (pRes.success && pRes.data) providers.value = pRes.data
    if (mRes.success && mRes.data) models.value = mRes.data
  } catch {
    // 拉不到就当作空列表，不影响浏览
  }
}

const modelCountOf = (providerId: number) => models.value.filter(m => m.provider_id === providerId).length

/** 渠道配色按 provider_type 取一个稳定色 */
const PROVIDER_COLORS: Record<string, string> = {
  openai: '#16a34a',
  deepseek: '#4f46e5',
  qwen: '#615cef',
  gemini: '#2563eb',
  claude: '#d97706',
  ollama: '#0891b2',
  other: '#64748b'
}
const providerColor = (t?: string) => PROVIDER_COLORS[t || 'other'] || '#64748b'

/** 个人渠道：按关键词 + 能力标签（该渠道下任一模型的类型）过滤 */
const channels = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return providers.value.filter(p => {
    if (q && !`${p.name} ${p.api_url} ${p.channel_name || ''}`.toLowerCase().includes(q)) return false
    if (tag.value) {
      const hit = models.value.some(m => m.provider_id === p.id && configTypeOf(m) === tag.value)
      if (!hit) return false
    }
    return true
  })
})

/** 平台预制：固定目录 */
const presets = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  return PROVIDER_PRESETS.filter(p => {
    if (tag.value && !p.tags.includes(tag.value)) return false
    if (q && !(p.name.toLowerCase().includes(q) || p.desc.toLowerCase().includes(q))) return false
    return true
  })
})

const channelModels = computed(() =>
  detailChannel.value ? models.value.filter(m => m.provider_id === detailChannel.value!.id) : []
)

/** 卡片配色取该平台第一个能力标签的颜色 */
const presetColor = (p: ProviderPreset) => typeDef(p.tags[0])?.color || '#64748b'

/** 是否已接入：按渠道名称或端点匹配 */
const isConnected = (p: ProviderPreset) =>
  providers.value.some(x => x.name === p.name || x.api_url === p.api_url)

const formatLabel = (f: string) =>
  ({
    openai: 'OpenAI Chat',
    responses: 'OpenAI Responses',
    claude: 'Anthropic',
    gemini: 'Gemini',
    custom_json: '自定义 JSON'
  })[f] || f

const authLabel = (a: string) =>
  ({
    api_key: 'API Key',
    ak_sk: 'Access Key + Secret Key',
    group_id: 'API Key + Group ID',
    endpoint_id: '推理接入点 ID'
  })[a] || a

const openDocs = () => {
  if (!current.value) return
  let origin = current.value.api_url
  try {
    origin = new URL(current.value.api_url).origin
  } catch {
    // 本地地址等无法解析时原样展示
  }
  ElMessage.info(`接入文档请参考平台官方说明：${origin}`)
}

/* ---------------- 渠道管理 ---------------- */
const channelDialogVisible = ref(false)
const isNewChannel = ref(true)
const editingChannel = ref<ApiProvider | null>(null)

const openAddChannel = () => {
  isNewChannel.value = true
  editingChannel.value = null
  channelDialogVisible.value = true
}

const openEditChannel = () => {
  isNewChannel.value = false
  editingChannel.value = detailChannel.value
  channelDialogVisible.value = true
}

const onChannelSaved = async () => {
  channelDialogVisible.value = false
  const keepId = detailChannel.value?.id
  await load()
  if (keepId) detailChannel.value = providers.value.find(p => p.id === keepId) || null
  emit('changed')
}

const onChannelDelete = async (id: number) => {
  const target = providers.value.find(p => p.id === id)
  try {
    await ElMessageBox.confirm(
      `确定删除渠道「${target?.name || id}」吗？该渠道下的模型会一并删除。`,
      '删除渠道',
      { type: 'warning' }
    )
  } catch {
    return
  }
  const res: any = await providerAPI.delete(id)
  if (!res?.success) {
    ElMessage.error(res?.message || '删除失败')
    return
  }
  ElMessage.success('渠道已删除')
  channelDialogVisible.value = false
  detailChannel.value = null
  await load()
  emit('changed')
}

/** 详情页里的「删除渠道」按钮 */
const removeChannel = async () => {
  if (detailChannel.value) await onChannelDelete(detailChannel.value.id)
}

/* ---------------- 引用 ---------------- */
const quoteVisible = ref(false)
const quoting = ref(false)
const quoteTarget = ref<ProviderPreset | null>(null)
const selectedModels = ref<string[]>([])
const form = ref({
  name: '',
  api_key: '',
  access_key: '',
  secret_key: '',
  group_id: '',
  endpoint_id: '',
  channel_name: ''
})

const openQuote = (p: ProviderPreset) => {
  quoteTarget.value = p
  selectedModels.value = p.models.map(m => m.model)
  form.value = {
    name: p.name,
    api_key: '',
    access_key: '',
    secret_key: '',
    group_id: '',
    endpoint_id: '',
    channel_name: ''
  }
  quoteVisible.value = true
}

const toggleModel = (model: string) => {
  const i = selectedModels.value.indexOf(model)
  if (i >= 0) selectedModels.value.splice(i, 1)
  else selectedModels.value.push(model)
}

const submitQuote = async () => {
  const p = quoteTarget.value
  if (!p) return
  if (!form.value.name.trim()) {
    ElMessage.warning('请填写渠道名称')
    return
  }
  if (p.auth_type === 'ak_sk') {
    if (!form.value.access_key || !form.value.secret_key) {
      ElMessage.warning('请填写 Access Key 与 Secret Key')
      return
    }
  } else if (!p.noKey && !form.value.api_key) {
    ElMessage.warning('请填写 API Key')
    return
  }

  quoting.value = true
  try {
    const res: any = await providerAPI.create({
      name: form.value.name.trim(),
      provider_type: p.provider_type,
      api_key: form.value.api_key,
      api_url: p.api_url,
      api_format: p.api_format,
      use_full_url: !!p.use_full_url,
      auth_type: p.auth_type,
      access_key: form.value.access_key,
      secret_key: form.value.secret_key,
      group_id: form.value.group_id,
      endpoint_id: form.value.endpoint_id,
      channel_name: form.value.channel_name,
      concurrency: 1
    } as any)

    if (!res?.success || !res?.data?.id) {
      ElMessage.error(res?.message || '创建渠道失败')
      return
    }
    const providerId = res.data.id

    let created = 0
    for (const m of p.models) {
      if (!selectedModels.value.includes(m.model)) continue
      try {
        const r: any = await configAPI.create({
          provider_id: providerId,
          name: m.name,
          model: m.model,
          config_type: m.config_type,
          max_tokens: m.max_tokens ?? 2000,
          temperature: 0.7,
          top_p: 0.9,
          frequency_penalty: 0,
          description: m.description || '',
          param_groups: m.paramGroups ? JSON.stringify(m.paramGroups) : '',
          default_params: m.defaultParams ? JSON.stringify(m.defaultParams) : '',
          enabled: 1,
          is_default: 0
        } as any)
        if (r?.success) created++
      } catch {
        // 单个模型失败不阻塞其余
      }
    }

    ElMessage.success(created > 0 ? `已接入 ${p.name}，并创建 ${created} 个模型` : `已接入 ${p.name}`)
    quoteVisible.value = false
    await load()
    emit('quoted')
  } finally {
    quoting.value = false
  }
}
</script>

<style scoped>
/*
 * 空状态修正。
 *
 * ⚠ 全局的 .config-page__empty-action 带 `margin-top: -16px`（定义在 Config.vue 的非 scoped
 *   样式块里，为「我的 API」卡片网格下方的用法设计，那里需要往上贴）。
 *   ApiPlaza 复用了同一个类名，但它位于 .plaza__models 内部，而后者是
 *   `flex: 1 1 0; min-height: 0; overflow-y: auto` —— 负 margin 把文字顶出容器上边界，
 *   高度 22.7px 的文字只剩底部约 6.7px 可见，也就是「文字被遮挡」。
 *
 *   这里就地覆盖为 0，并把空状态在可用高度里居中，让它看起来是有意为之而不是贴顶的碎字。
 *   断言：文字顶边必须 >= 容器顶边（见 scripts/_probe_plaza_empty.mjs）。
 */
.plaza__models .config-page__empty-action {
  margin-top: 0;
  flex: 1 1 auto;
  min-height: 120px;
}

/* 空状态的说明文字：避免在窄容器里挤成一行被截断 */
.plaza__models .config-page__empty-action > p {
  margin: 0;
  text-align: center;
  line-height: 1.7;
}

/* 广场上下分组（蓝本无此结构，B 项目为「个人渠道 / 平台预制」新增） */
.plaza__group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.plaza__group + .plaza__group {
  margin-top: 6px;
}

.plaza__group-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 2px 2px 0;
}

.plaza__group-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--cfg-ink);
}

.plaza__group-count {
  font-size: 11.5px;
  color: var(--cfg-ink-3);
  background: var(--cfg-panel-2);
  border: 1px solid var(--cfg-line);
  border-radius: 999px;
  padding: 0 8px;
}

.plaza__toolbar-ops {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 蓝本 CSS 未覆盖的少量补充：引用弹窗里的模型多选列表 */
.quote__models {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 220px;
  overflow-y: auto;
  padding-right: 2px;
}

.quote__model-card--pick {
  cursor: pointer;
  user-select: none;
}

.quote__model-card--pick:hover {
  border-color: var(--cfg-accent);
}

.quote__model-card--pick .config-page__api-tag {
  margin-left: auto;
  flex-shrink: 0;
}
</style>
