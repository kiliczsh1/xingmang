<template>
  <div class="config-page">
    <div class="config-page__content">
      <div class="config-page__scroll">
        <div class="config-page__inner">
    <!-- 顶部：标题 + 导航 + 导入导出 -->
    <div class="config-page__header">
      <div class="config-page__heading">
        <h2>模型服务</h2>
        <p>「我的 API」管理已接入的自定义模型；「API 广场」从支持的平台一键引用新模型。</p>
      </div>

      <div class="config-page__nav-tabs">
        <button
          type="button"
          class="config-page__nav-tab"
          :class="{ 'is-active': activeNav === 'mine' }"
          @click="activeNav = 'mine'"
        >
          <el-icon :size="14"><Collection /></el-icon>
          我的 API
          <span class="config-page__nav-count">{{ models.length }}</span>
        </button>
        <button
          type="button"
          class="config-page__nav-tab"
          :class="{ 'is-active': activeNav === 'plaza' }"
          @click="activeNav = 'plaza'"
        >
          <el-icon :size="14"><Shop /></el-icon>
          API 广场
        </button>
      </div>

      <div class="config-page__header-actions">
        <el-dropdown trigger="click" @command="handleImportCommand">
          <button class="cfg-btn">
            <el-icon><Upload /></el-icon>导入
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="importConfig">
                <el-icon><Upload /></el-icon>
                导入配置文件
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown trigger="click" @command="handleExportCommand">
          <button class="cfg-btn">
            <el-icon><Download /></el-icon>导出
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="exportAll">
                <el-icon><Document /></el-icon>
                导出全部配置
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <button class="cfg-btn primary" @click="handleCreateNewChannel">
          <el-icon><Plus /></el-icon>创建自定义渠道
        </button>
      </div>
    </div>

      <!-- ============ 我的 API ============ -->
      <div v-if="activeNav === 'mine'" class="config-page__mine">
        <div class="config-page__mine-toolbar">
          <div class="config-page__type-tabs">
            <button
              type="button"
              class="config-page__type-tab"
              :class="{ 'is-active': curCat === '' }"
              @click="curCat = ''"
            >
              全部
              <span class="config-page__type-count">{{ models.length }}</span>
            </button>
            <button
              v-for="t in MODEL_TYPES"
              :key="t.key"
              type="button"
              class="config-page__type-tab"
              :class="{ 'is-active': curCat === t.key }"
              @click="curCat = t.key"
            >
              {{ t.name }}
              <span class="config-page__type-count">{{ modelsByType(t.key).length }}</span>
            </button>
          </div>

          <div class="config-page__mine-ops">
            <el-input v-model="mineSearch" placeholder="搜索 API 名" clearable class="config-page__mine-search">
              <template #prefix><el-icon><Search /></el-icon></template>
            </el-input>
            <el-select v-model="mineSort" class="config-page__mine-sort" style="width: 150px">
              <el-option label="创建时间 · 倒序" value="created_desc" />
              <el-option label="创建时间 · 升序" value="created_asc" />
              <el-option label="名称 · A-Z" value="name_asc" />
              <el-option label="名称 · Z-A" value="name_desc" />
            </el-select>
          </div>
        </div>

        <div class="config-page__scroll">
          <div class="config-page__api-list">
            <div v-for="g in apiGroups" :key="g.key" class="config-page__api-group">
              <div class="config-page__group-head">
                <span class="config-page__group-dot" :style="{ background: g.color }"></span>
                <span class="config-page__group-name">{{ g.name }}</span>
                <span
                  class="config-page__group-count"
                  :class="{ 'is-empty': g.totalModels === 0 }"
                >{{ g.totalModels }} 个模型</span>
                <code v-if="g.baseUrl" class="config-page__group-tag">{{ g.baseUrl }}</code>
                <span v-if="g.concurrency > 1" class="config-page__group-conc">并发 {{ g.concurrency }}</span>
                <div class="config-page__group-actions">
                  <button
                    type="button"
                    class="cfg-btn sm"
                    :disabled="discoveringModels"
                    @click="handlePullModels(g)"
                  >
                    <el-icon><Search /></el-icon>{{ discoveringModels ? '拉取中' : '从渠道导入' }}
                  </button>
                  <button type="button" class="cfg-btn sm danger-outline" @click="handleDeleteGroupProvider(g)">删除渠道</button>
                </div>
              </div>

              <div class="config-page__card-grid">
                <div v-for="m in g.items" :key="m.id" class="config-page__api-card">
                  <div class="config-page__api-card-head">
                    <span class="config-page__api-model" :title="m.model">{{ m.name }}</span>
                    <span class="config-page__api-type" :style="{ background: typeColor(m) }">{{ typeLabel(m) }}</span>
                  </div>
                  <div class="config-page__api-tags">
                    <span v-if="m.is_default === 1" class="config-page__api-tag config-page__api-tag--success">默认</span>
                    <span v-if="m.enabled !== 1" class="config-page__api-tag">已停用</span>
                    <span class="config-page__api-tag">{{ m.model }}</span>
                  </div>
                  <div class="config-page__api-card-foot">
                    <button type="button" class="cfg-btn sm" :disabled="testing" @click="handleTestModel(m)">测试</button>
                    <button type="button" class="cfg-btn sm" @click="selectModel(m)">编辑</button>
                    <button type="button" class="cfg-btn sm danger" @click="handleDeleteModel(m)">删除</button>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="apiGroups.length === 0" class="config-page__empty-action">
              <p>还没有任何模型，先去 API 广场引用，或建一个自定义渠道再拉取模型</p>
              <div>
                <button type="button" class="config-page__empty-link" @click="handleCreateNewChannel">创建渠道</button>
                <button
                  type="button"
                  class="config-page__empty-link config-page__empty-link--ghost"
                  @click="activeNav = 'plaza'"
                >去 API 广场逛逛</button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ============ API 广场（内联视图） ============ -->
      <ApiPlaza
        v-else-if="activeNav === 'plaza'"
        @quoted="handleQuoted"
        @changed="handlePlazaChanged"
        @add-model="handleAddModelForChannel"
      />
        </div>
      </div>
    </div>

    <!-- ============ 自定义 API 弹窗（模型编辑） ============ -->
    <CapDialog
      v-model:visible="editDialogVisible"
      v-model:dirty="dirty"
      :model="editingModel"
      :is-new="isNewModel"
      :testing="testing"
      @saved="handleCapSaved"
      @test="handleTestModel"
      @delete="handleDeleteEditingModel"
      @cancel="cancelNewModel"
      @closed="onEditDialogClosed"
    />

    <!-- ============ 导入弹窗 ============ -->
    <el-dialog v-model="importDialogVisible" title="导入API配置" width="var(--app-dialog-width-form)" :close-on-click-modal="false" append-to-body class="cfg-dialog">
      <div class="import-content">
        <div class="import-upload">
          <el-upload
            ref="uploadRef"
            drag
            accept=".json"
            :auto-upload="false"
            :show-file-list="false"
            :on-change="handleFileChange"
          >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">将配置文件拖到此处，或<em>点击上传</em></div>
            <template #tip>
              <div class="el-upload__tip">只能上传 JSON 格式的配置文件</div>
            </template>
          </el-upload>
        </div>
        <div v-if="importPreview" class="import-preview">
          <el-alert title="配置预览" type="info" :closable="false" show-icon>
            <template #default>
              <div class="preview-info">
                <p><strong>服务商数量：</strong>{{ importPreview.providers?.length || 0 }} 个</p>
                <p><strong>模型数量：</strong>{{ importPreview.models?.length || 0 }} 个</p>
                <p><strong>导出时间：</strong>{{ importPreview.exportTime || '未知' }}</p>
              </div>
            </template>
          </el-alert>
          <div class="import-options">
            <el-checkbox v-model="importOptions.overrideExisting">覆盖同名配置（不勾选则跳过同名配置）</el-checkbox>
          </div>
        </div>
      </div>
      <template #footer>
        <button class="cfg-btn" @click="importDialogVisible = false">取消</button>
        <button class="cfg-btn primary" :disabled="!importPreview" @click="executeImport">导入</button>
      </template>
    </el-dialog>

    <!-- ============ 远程模型（从渠道拉取） ============ -->
    <el-dialog v-model="modelListDialogVisible" title="远程模型" width="var(--app-dialog-width-content)" append-to-body class="cfg-dialog">
      <div class="ml-toolbar">
        <el-input
          v-model="mlSearch"
          class="ml-toolbar__search"
          placeholder="搜索模型标识"
          clearable
        />
        <div class="ml-types">
          <button
            type="button"
            class="ml-type-btn"
            :class="{ on: mlTypeFilter === '' }"
            @click="mlTypeFilter = ''"
          >全部 <b>{{ discoveredModels.length }}</b></button>
          <button
            v-for="t in MODEL_TYPES"
            :key="t.key"
            type="button"
            class="ml-type-btn"
            :class="{ on: mlTypeFilter === t.key }"
            :style="mlTypeFilter === t.key ? { borderColor: t.color, color: t.color } : undefined"
            @click="mlTypeFilter = mlTypeFilter === t.key ? '' : t.key"
          >{{ t.name }} <b>{{ typeCounts[t.key] || 0 }}</b></button>
        </div>
      </div>
      <el-table
        ref="remoteModelsTableRef"
        :data="filteredModels"
        stripe
        height="420"
        row-key="id"
        @selection-change="onRemoteSelectionChange"
      >
        <el-table-column type="selection" width="46" />
        <el-table-column prop="id" label="模型标识" min-width="300" show-overflow-tooltip />
        <el-table-column label="类型" width="116">
          <template #default="{ row }">
            <el-select v-model="mlTypeMap[row.id]" size="small" class="ml-type-select">
              <el-option v-for="t in MODEL_TYPES" :key="t.key" :label="t.name" :value="t.key" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="来源" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.owned_by || '—' }}</template>
        </el-table-column>
        <el-table-column label="探测" width="176">
          <template #default="{ row }">
            <span
              v-if="probeStates[row.id]?.state === 'running'"
              class="probe-cell probe-running"
            ><i class="probe-spin" />{{ probeStates[row.id].text }}</span>
            <span
              v-else-if="probeStates[row.id]?.state === 'ok'"
              class="probe-cell probe-ok"
              :title="probeStates[row.id].detail || ''"
            >{{ probeStates[row.id].text }}</span>
            <span
              v-else-if="probeStates[row.id]?.state === 'fail'"
              class="probe-cell probe-fail"
              :title="probeStates[row.id].detail || ''"
            >{{ probeStates[row.id].text }}</span>
            <span v-else class="probe-cell probe-idle">—</span>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <div class="remote-foot">
          <span class="remote-foot__hint">
            <template v-if="probing">
              <i class="probe-spin" />正在逐条测试 {{ probeDone }}/{{ probeTotal }}，每个约 2–8 秒
            </template>
            <template v-else-if="needsProbe">
              已选 {{ selectedModelIds.length }} 个，先「测试选中」才能导入
            </template>
            <template v-else-if="selectionTested && selectedModelIds.length === 1">
              已测完：点「导入选中」
            </template>
            <template v-else-if="selectionTested && selectedModelIds.length >= 2">
              已选 {{ selectedModelIds.length }} 个且已测完：点「全部导入」一起导入
            </template>
            <template v-else-if="pulledProviderName">渠道：{{ pulledProviderName }}</template>
          </span>
          <div class="remote-foot__ops">
            <!-- 关闭只保留右上角的 X：底部再放一个属于重复入口 -->
            <el-button
              :disabled="selectedModelIds.length === 0 || probing || importingModels"
              :loading="probing"
              @click="probeSelectedModels"
            >测试选中</el-button>
            <!-- 单选：选完并测完才允许导入 -->
            <el-button
              type="primary"
              :disabled="!canImportOne"
              :loading="importingModels"
              @click="importSelectedModels"
            >导入选中</el-button>
            <!-- 多选：同样的动作，但只导入这 N 个，不是导入渠道全部 -->
            <el-button
              :disabled="!canImportMany"
              :loading="importingModels"
              @click="importSelectedBulk"
            >全部导入</el-button>
          </div>
        </div>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, h, type VNode } from 'vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import {
  Upload, Download, ArrowDown, Document, UploadFilled, Search,
  Plus, Collection, EditPen, Check, Connection, Delete,
  ChatDotRound, MagicStick, PictureFilled, Cpu, Setting, Shop
} from '@element-plus/icons-vue'
import { providerAPI, configAPI } from '@/api'
import type { ApiProvider, ApiModel, ApiConfigType } from '@/types'
import { MODEL_TYPES, configTypeOf, typeLabel, typeColor, filterByTypes, guessConfigType } from '@/services/modelType'
import ApiPlaza from '@/components/config/ApiPlaza.vue'
import CapDialog from '@/components/config/CapDialog.vue'
import {
  guessCategory,
  fmtTok,
  resolveMaxTokens,
  resolveMaxOutput
} from '@/services/modelMeta'

/* ---------------- 模型类型（决定该模型被哪条链路使用） ----------------
 * 与上面的「能力分类」是两回事：
 * - 类型（config_type）由后端强约束，决定对话 / 生图 / 生视频 / 图片分析 / 向量
 * - 分类（category）只是展示用的能力标签，由模型 ID 推断
 *
 * 类型定义与判定口径统一放在 @/services/modelType，供封面 / 插图 / 工作流共用。
 */
const modelsByType = (key: string) => filterByTypes(models.value, key as ApiConfigType)

/* ---------------- 基础状态 ---------------- */
/** 顶层导航：我的 API / API 广场 / 服务商 */
const activeNav = ref<'mine' | 'plaza'>('mine')
/** 自定义 API（模型编辑）弹窗 */
const editDialogVisible = ref(false)
/** 我的 API：搜索与排序 */
const mineSearch = ref('')
const mineSort = ref<'created_desc' | 'created_asc' | 'name_asc' | 'name_desc'>('created_desc')
const curCat = ref('')
const providers = ref<ApiProvider[]>([])
const models = ref<ApiModel[]>([])

/* ---------------- 编辑器状态 ---------------- */
const editingModel = ref<ApiModel | null>(null)
const selectedId = ref<number | null>(null)
const isNewModel = ref(false)
const dirty = ref(false)
const testing = ref(false)
const discoveredModels = ref<
  Array<{
    id: string
    name: string
    owned_by?: string
    /** 上游主动给出的上下文窗口；缺失（中转不透传）时为空 */
    context_window?: number
    /** 上游主动给出的单次输出上限；缺失时为空，此时才回落到本地表 */
    max_output_tokens?: number
  }>
>([])
const discoveringModels = ref(false)
/** 最近一次「从渠道导入」拉取的是哪个服务商（无编辑态时用它建模型） */
const pulledProviderId = ref<number | null>(null)

/* ---------------- 状态栏计算 ---------------- */
const enabledCount = computed(() => models.value.filter(m => m.enabled === 1).length)

/* ---------------- 生命周期 ---------------- */
onMounted(async () => {
  await fetchProviders()
  await fetchModels()
})

/* ---------------- 服务商逻辑 ---------------- */
// 不再自动创建默认服务商。
//
// 历史行为：首次进入（providers 为空）时自动插入 DeepSeek / SiliconFlow /
// bltcy.ai / ModelScope 四条**空 Key** 服务商，并逐条弹「已自动添加…」。
// 这四条既不能直接使用（没有 Key），又会被当成"我的渠道"参与分组与计数，
// 用户还得手动删掉。空状态本来就该是空的 —— 引导放在广场的「添加渠道」入口里。
const fetchProviders = async () => {
  const res = await providerAPI.getAll()
  if (res.success && res.data) {
    providers.value = res.data
  }
}

/** 服务商类型 → logo 品牌色 */
const providerColor = (type: string) => {
  const map: Record<string, string> = {
    openai: '#10a37f', azure: '#0078d4', claude: '#d97757',
    qwen: '#615ced', deepseek: '#4d6bfe', gemini: '#4285f4',
    ollama: '#2e2e2e', siliconflow: '#3b82f6', modelscope: '#6236ff'
  }
  return map[type] || '#64748b'
}

const getProviderModelCount = (providerId: number) => models.value.filter(m => m.provider_id === providerId).length
const providerNameOf = (providerId: number) => providers.value.find(p => p.id === providerId)?.name || '未知服务商'

/**
 * 渠道是否「配置完整」——决定它要不要出现在「我的 API」里。
 *
 * 口径（2026-10-16 用户明确）：**出现条件不是「有模型」，而是「有密钥 + 有可用 URL」**。
 * 以前的分组是从 models 反推渠道，导致「渠道已配好但还没加模型」时整条渠道消失，
 * 用户看不到自己刚建好的渠道，只能再去 API 广场找。
 *
 * 认证方式不同，「密钥」落在不同字段：
 *   api_key      → api_key
 *   ak_sk        → access_key + secret_key
 *   group_id     → access_key + group_id      （火山方舟那类）
 *   endpoint_id  → access_key + endpoint_id
 */
const isChannelConfigured = (p: ApiProvider): boolean => {
  if (!(p.api_url || '').trim()) return false
  const auth = p.auth_type || 'api_key'
  const has = (v?: string) => !!(v || '').trim()
  switch (auth) {
    case 'ak_sk':
      return has(p.access_key) && has(p.secret_key)
    case 'group_id':
      return has(p.access_key) && has(p.group_id)
    case 'endpoint_id':
      return has(p.access_key) && has(p.endpoint_id)
    // ollama 这类本地服务可以不填 key，此时 URL 本身就是「可用」的判据
    case 'api_key':
    default:
      return has(p.api_key) || has(p.access_key)
  }
}

/**
 * 我的 API：按渠道（provider）分组的卡片流。
 *
 * 分组主体是**渠道**而不是模型：只要渠道满足 isChannelConfigured 就出现，
 * 哪怕一个模型都没有（此时显示空状态 + 添加入口）。
 * 受类型 tab（curCat）+ 搜索（mineSearch）+ 排序（mineSort）三重约束。
 */
const apiGroups = computed(() => {
  const kw = mineSearch.value.trim().toLowerCase()

  const groups = providers.value
    .filter(p => isChannelConfigured(p))
    .filter(p => {
      if (!kw) return true
      return `${p.name} ${p.api_url || ''} ${p.channel_name || ''}`.toLowerCase().includes(kw)
    })
    .map(p => {
      const all = models.value
        .filter(m => m.provider_id === p.id)
        .sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
      // 类型 tab 是模型维度的筛选：命中则只留该类型；
      // ⚠ 渠道本身没有模型的渠道（all.length === 0）在选了类型 tab 时保留，
      //   否则「还没加模型的新渠道」会在切到某个类型页签后凭空消失，用户又找不到它了。
      const items = curCat.value ? all.filter(m => configTypeOf(m) === curCat.value) : all
      return {
        key: p.id,
        name: p.name,
        color: providerColor(p.provider_type),
        baseUrl: p.api_url || '',
        concurrency: p.concurrency || 1,
        provider: p,
        /** 该渠道的全部模型数（不受类型 tab 影响），用于「N 个模型」这类口径展示 */
        totalModels: all.length,
        items,
        /** 渠道下没有任何模型：显示空状态与添加入口 */
        isEmpty: all.length === 0
      }
    })
    // 类型 tab 下，把「该类型一个都没有、但渠道里有别的模型」的渠道剔掉；
    // isEmpty 的渠道始终保留（见上）。
    .filter(g => !curCat.value || g.isEmpty || g.items.length > 0)

  const dir = mineSort.value
  groups.sort((a, b) => {
    if (dir === 'name_asc') return a.name.localeCompare(b.name, 'zh')
    if (dir === 'name_desc') return b.name.localeCompare(a.name, 'zh')
    const ai = a.provider?.id || 0
    const bi = b.provider?.id || 0
    return dir === 'created_asc' ? ai - bi : bi - ai
  })
  return groups
})

const handleDeleteProvider = async (provider: ApiProvider) => {
  try {
    const modelCount = getProviderModelCount(provider.id)
    let confirmMsg = `确定删除服务商"${provider.name}"吗？`
    if (modelCount > 0) {
      confirmMsg += `\n注意：该服务商下有 ${modelCount} 个模型，删除后这些模型也会被删除！`
    }
    await ElMessageBox.confirm(confirmMsg, '提示', { type: 'warning' })
    const res = await providerAPI.delete(provider.id)
    if (res.success) {
      ElMessage.success('删除成功')
      await fetchProviders()
      await fetchModels()
    }
  } catch (error) {
  }
}

/* ---------------- 模型逻辑 ---------------- */
const fetchModels = async () => {
  const res = await configAPI.getAll()
  if (res.success && res.data) {
    models.value = res.data
  }
}

/** API 广场引用完成后刷新服务商与模型 */
const handleQuoted = async () => {
  await fetchProviders()
  await fetchModels()
}

const selectModel = (m: ApiModel) => {
  selectedId.value = m.id
  isNewModel.value = false
  dirty.value = false
  discoveredModels.value = []
  selectedModelIds.value = []
  editDialogVisible.value = true
  editingModel.value = {
    ...m,
    category: m.category || guessCategory(m.model),
    temperature: m.temperature ?? 0.7,
    top_p: m.top_p ?? 0.9,
    frequency_penalty: m.frequency_penalty ?? 0,
    max_tokens: m.max_tokens ?? 2000,
    enabled: m.enabled ?? 1,
    is_default: m.is_default ?? 0,
    config_type: configTypeOf(m) as ApiConfigType,
    prompt: m.prompt ?? '',
    prompt_role: m.prompt_role || 'system',
    top_k: m.top_k ?? null,
    presence_penalty: m.presence_penalty ?? null,
    stop: m.stop ?? '',
    max_tokens_mode: m.max_tokens_mode || 'auto',
    thinking_mode: m.thinking_mode ?? 0,
    thinking_effort: m.thinking_effort || 'default',
    generation_config_source: m.generation_config_source || 'manual',
    request_rules: m.request_rules ?? null,
    is_free: m.is_free ?? 0,
    input_ratio: m.input_ratio ?? null,
    output_ratio: m.output_ratio ?? null
  }
}

/**
 * 打开「创建自定义渠道」弹窗。
 *
 * 模型不在这里建 —— 渠道建好后用分组头的「从渠道导入」批量拉取。
 * 这样「建渠道」只有一处入口（原先这里和 API 广场的渠道弹窗都能建，重复了）。
 */
const openNewChannelDialog = () => {
  activeNav.value = 'mine'
  isNewModel.value = true
  selectedId.value = null
  dirty.value = false
  editingModel.value = null
  editDialogVisible.value = true
}

const handleCreateNewChannel = () => openNewChannelDialog()

/** 广场渠道详情里点「从渠道导入」：拉取该渠道的模型列表 */
const handleAddModelForChannel = (providerId: number) => {
  const p = providers.value.find(x => x.id === providerId)
  if (!p) {
    ElMessage.warning('渠道不存在')
    return
  }
  return pullModelsForProvider(p)
}

/** 广场里改过渠道（新增/编辑/删除）后刷新本页数据 */
const handlePlazaChanged = async () => {
  await fetchProviders()
  await fetchModels()
}

const cancelNewModel = () => {
  editingModel.value = null
  isNewModel.value = false
  dirty.value = false
  editDialogVisible.value = false
}

/** 弹窗关闭后清理临时状态（避免下次打开仍显示「创建自定义 API」） */
const onEditDialogClosed = () => {
  isNewModel.value = false
  dirty.value = false
}

const handleDeleteEditingModel = async () => {
  const m = editingModel.value
  if (!m || !m.id) return
  try {
    await ElMessageBox.confirm(`确定删除模型"${m.name}"吗？`, '提示', { type: 'warning' })
    const res = await configAPI.delete(m.id)
    if (res.success) {
      ElMessage.success('删除成功')
      editingModel.value = null
      selectedId.value = null
      editDialogVisible.value = false
      await fetchModels()
    }
  } catch (error) {
  }
}

/** 卡片上的删除：直接删指定模型（不必先打开弹窗） */
const handleDeleteModel = async (m: ApiModel) => {
  if (!m?.id) return
  try {
    await ElMessageBox.confirm(`确定删除模型"${m.name}"吗？`, '提示', { type: 'warning' })
    const res = await configAPI.delete(m.id)
    if (res.success) {
      ElMessage.success('删除成功')
      if (editingModel.value?.id === m.id) editingModel.value = null
      await fetchModels()
    }
  } catch (error) {
  }
}

/** 分组头「从渠道导入」：拉取该渠道的模型列表并打开选择弹窗 */
/** 拉取指定渠道的模型列表并打开导入弹窗（分组头与广场详情共用） */
const pullModelsForProvider = async (p: ApiProvider) => {
  discoveringModels.value = true
  discoveredModels.value = []
  selectedModelIds.value = []
  // 每次打开都是新一批模型，旧的探测结果不能留在这里冒充结论。
  probeStates.value = {}
  probeDone.value = 0
  probeTotal.value = 0
  mlSearch.value = ''
  try {
    const res = await providerAPI.discoverModels(p.id)
    if (res.success && res.data && res.data.length > 0) {
      discoveredModels.value = res.data
      pulledProviderId.value = p.id
      syncMlTypeMap(res.data)
      modelListDialogVisible.value = true
      ElMessage.success(`成功获取 ${res.data.length} 个可用模型`)
    } else {
      ElMessage.warning('该渠道未返回可用模型列表')
    }
  } catch (error) {
    ElMessage.error('拉取模型列表失败，请检查 API 地址和密钥是否正确')
  } finally {
    discoveringModels.value = false
  }
}

const handlePullModels = (g: { provider: ApiProvider | null }) => {
  if (!g.provider) {
    ElMessage.warning('该渠道信息缺失，无法拉取')
    return
  }
  return pullModelsForProvider(g.provider)
}

/** 分组头「删除渠道」：删服务商（级联删其下模型） */
const handleDeleteGroupProvider = async (g: { provider: ApiProvider | null }) => {
  if (!g.provider) return
  await handleDeleteProvider(g.provider)
}

/**
 * 测试成功的结构化结果（照搬蓝本）：
 * 模型 / 提供商 / 延迟 / Token 用量 / AI 回复。
 * 用 h() 构造 VNode 而非 HTML 字符串 —— 模型名和回复正文都来自上游，
 * 拼 HTML 会有注入风险，VNode 会自动转义。
 */
const buildTestResult = (d: any, m: ApiModel): VNode => {
  const rows: VNode[] = [
    h('p', [h('strong', '模型: '), d.modelName || m.name]),
    h('p', [h('strong', '提供商: '), d.providerName || providerNameOf(m.provider_id)]),
    h('p', [h('strong', '延迟: '), `${d.latencyMs}ms`])
  ]
  if (d.usage) {
    rows.push(h('p', [h('strong', 'Token用量:')]))
    rows.push(h('ul', { style: { paddingLeft: '20px' } }, [
      h('li', `输入: ${d.usage.prompt_tokens}`),
      h('li', `输出: ${d.usage.completion_tokens}`),
      h('li', `总计: ${d.usage.total_tokens}`)
    ]))
  }
  rows.push(h('p', [h('strong', 'AI回复:')]))
  rows.push(h('div', {
    style: {
      background: 'var(--el-fill-color-light)',
      color: 'var(--el-text-color-primary)',
      padding: '10px',
      borderRadius: '4px',
      whiteSpace: 'pre-wrap'
    }
  }, d.preview || '(空)'))
  return h('div', {
    style: {
      maxHeight: '300px',
      overflowY: 'auto',
      overflowWrap: 'anywhere',
      // B 项目全局 reset 掉了 p / ul 的默认间距（* { margin: 0; padding: 0 }），
      // 不自己补的话几行会挤成一坨 —— 蓝本没这问题是因为它没 reset
      display: 'flex',
      flexDirection: 'column',
      gap: '6px'
    }
  }, rows)
}

const handleTestModel = async (m: ApiModel) => {
  if (!m.id || testing.value) return
  testing.value = true
  try {
    const res = await configAPI.testModel(m.id)
    const d = (res as any).data || {}
    if (res.success && d.success) {
      await ElMessageBox.alert(buildTestResult(d, m), '模型测试成功', { confirmButtonText: '确定' })
    } else {
      ElNotification.error({ message: `测试失败: ${d.message || '未知错误'}`, duration: 5000 })
    }
  } catch (error) {
    ElNotification.error({ message: '测试失败: 网络错误', duration: 5000 })
  } finally {
    testing.value = false
  }
}

/* ---------------- 发现模型 ---------------- */

/* 获取到的模型列表弹窗 */
const modelListDialogVisible = ref(false)
const mlSearch = ref('')
/** 模型列表弹窗：类型筛选（'' = 全部） */
const mlTypeFilter = ref<ApiConfigType | ''>('')
/** 模型列表弹窗：每个模型的类型。拉取时按模型 ID 推断，用户可在表格里改 */
const mlTypeMap = ref<Record<string, ApiConfigType>>({})
/** 模型列表弹窗：多选（支持「导入选中 / 全部导入」） */
const selectedModelIds = ref<string[]>([])
/** 批量导入进行中 */
const importingModels = ref(false)

/* ---------------- 导入弹窗内的单模型探测 ----------------
 * 用户主动点「测试选中」才发请求：不勾选不测、不点不测。
 * 结果只写探测缓存（可填的 max_tokens），**不改模型的生产配置** ——
 * 把大值写进 max_tokens 会让每条请求都按它预扣费，比探测本身贵得多。
 */
type ProbeRowState = { state: 'running' | 'ok' | 'fail'; text: string; detail?: string }
const probeStates = ref<Record<string, ProbeRowState>>({})
const probing = ref(false)
const probeDone = ref(0)
const probeTotal = ref(0)

/* 导入门槛：必须先测过才能导入。
 * 这不是流程洁癖 —— 没测过的模型导进来，用户对它的上限一无所知，
 * 而"测试"是唯一零成本拿到这个信息的时机。 */
const selectionTested = computed(
  () =>
    selectedModelIds.value.length > 0 &&
    selectedModelIds.value.every(id => {
      const st = probeStates.value[id]?.state
      return st === 'ok' || st === 'fail'
    })
)

/** 有勾选但还没测完 —— 底部提示用 */
const needsProbe = computed(
  () => selectedModelIds.value.length > 0 && !selectionTested.value && !probing.value
)

/** 单选可导入 */
const canImportOne = computed(
  () =>
    selectedModelIds.value.length === 1 &&
    selectionTested.value &&
    !probing.value &&
    !importingModels.value
)

/** 多选可导入（≥2）：动作相同，只是数量门槛不同 */
const canImportMany = computed(
  () =>
    selectedModelIds.value.length >= 2 &&
    selectionTested.value &&
    !probing.value &&
    !importingModels.value
)

const PROBE_FAIL_TEXT: Record<string, string> = {
  auth: '鉴权失败',
  quota: '额度不足以预扣',
  notfound: '模型不存在',
  ratelimit: '触发限流',
  unsupported: '不支持探测',
  network: '连接失败',
  upstream: '上游故障',
  limit_out: '全部档位被拒',
  limit_ctx: '上下文超限',
  error: '测试失败'
}

/** 逐条测试选中模型：串行，避免同一渠道被并发打爆 */
const probeSelectedModels = async () => {
  const ids = selectedModelIds.value.slice()
  if (ids.length === 0) return
  const pid = pulledProviderId.value || editingModel.value?.provider_id
  if (!pid) {
    ElMessage.warning('请先选择要测试的渠道')
    return
  }

  probing.value = true
  probeDone.value = 0
  probeTotal.value = ids.length
  let ok = 0
  let fail = 0
  let skip = 0
  // 渠道级故障（鉴权/模型不存在/连不上）一旦出现，剩余模型全部跳过 ——
  // 否则 20 个模型会各打一发注定失败的请求，纯属浪费时间并可能触发风控。
  let channelDown = ''

  try {
    for (const id of ids) {
      const type = mlTypeMap.value[id] || guessConfigType(id)
      if (channelDown) {
        probeStates.value[id] = { state: 'fail', text: `跳过（${channelDown}）` }
        skip++
        probeDone.value++
        continue
      }
      // 非对话模型的 max_tokens 语义不同（图片尺寸 / 向量维度），后端也会拒；
      // 这里先跳过，省一次必失败的网络往返。
      if (type !== 'chat') {
        probeStates.value[id] = { state: 'fail', text: '跳过（非对话）', detail: '该类型不支持上限探测' }
        skip++
        probeDone.value++
        continue
      }

      probeStates.value[id] = { state: 'running', text: '测试中' }
      const found = discoveredModels.value.find(x => x.id === id)
      try {
        const res: any = await configAPI.probe({
          provider_id: pid,
          model: id,
          config_type: type,
          kind: 'cap',
          declared_max: found?.max_output_tokens || 0
        })
        const d = res?.data
        if (d?.reachable) {
          ok++
          probeStates.value[id] = {
            state: 'ok',
            // 「可填」而不是「已验证」：200 只证明这个值没被拒，
            // 不证明模型真能生成这么多。真实输出上限要靠长输出校准。
            text: d.accepted > 0 ? `可填 ${fmtTok(d.accepted)}` : '连通正常',
            detail: `${d.latency_ms}ms · ${d.steps?.length || 0} 发`
          }
        } else {
          fail++
          probeStates.value[id] = {
            state: 'fail',
            text: PROBE_FAIL_TEXT[d?.conclusion] || '测试失败',
            detail: d?.message || ''
          }
          if (['auth', 'notfound', 'network'].includes(d?.conclusion)) {
            channelDown = PROBE_FAIL_TEXT[d.conclusion] || '渠道不可用'
          }
        }
      } catch (e: any) {
        fail++
        probeStates.value[id] = {
          state: 'fail',
          text: '请求失败',
          detail: e?.response?.data?.message || e?.message || ''
        }
      }
      probeDone.value++
    }
  } finally {
    probing.value = false
  }

  const msg = `测试完成：可用 ${ok}${fail ? `，失败 ${fail}` : ''}${skip ? `，跳过 ${skip}` : ''}`
  if (fail) ElMessage.warning(msg)
  else ElMessage.success(msg)
}

/** 各类模型数量（筛选按钮上的角标） */
const typeCounts = computed(() => {
  const c: Record<string, number> = {}
  for (const m of discoveredModels.value) {
    const t = mlTypeMap.value[m.id] || 'chat'
    c[t] = (c[t] || 0) + 1
  }
  return c
})

/** 拉取完成后初始化每行的类型，并重置筛选 */
const syncMlTypeMap = (list: Array<{ id: string }>) => {
  mlTypeMap.value = Object.fromEntries(list.map(m => [m.id, guessConfigType(m.id)]))
  mlTypeFilter.value = ''
}

const filteredModels = computed(() => {
  let list = discoveredModels.value
  if (mlTypeFilter.value) {
    list = list.filter(m => (mlTypeMap.value[m.id] || 'chat') === mlTypeFilter.value)
  }
  const q = mlSearch.value.trim().toLowerCase()
  if (!q) return list
  return list.filter(
    m => m.id.toLowerCase().includes(q) || (m.name || '').toLowerCase().includes(q)
  )
})

/** 远程模型表格：多选同步到 selectedModelIds（表格自身维护选中态） */
const remoteModelsTableRef = ref()
const onRemoteSelectionChange = (rows: Array<{ id: string }>) => {
  selectedModelIds.value = rows.map(r => r.id)
}

/** 底部左侧显示的渠道名 */
const pulledProviderName = computed(() => {
  const pid = pulledProviderId.value
  return pid ? providers.value.find(p => p.id === pid)?.name || '' : ''
})

/** 弹窗保存成功（保存逻辑已移入 CapDialog） */
/** 弹窗保存成功：建渠道 → 刷新渠道；改模型 → 刷新模型 */
const handleCapSaved = async (kind: 'channel' | 'model') => {
  if (kind === 'channel') await fetchProviders()
  else await fetchModels()
  editDialogVisible.value = false
}

/** 批量建模型：把给定的模型 ID 创建到目标渠道下（已存在的跳过） */
const importModels = async (ids: string[]) => {
  const pid = pulledProviderId.value || editingModel.value?.provider_id || null
  if (!pid) {
    ElMessage.warning('请先选择要导入到哪个渠道')
    return
  }
  if (ids.length === 0) return

  const existing = new Set(
    models.value.filter(m => m.provider_id === pid).map(m => String(m.model || '').trim())
  )
  let created = 0
  let skipped = 0
  let failed = 0

  importingModels.value = true
  try {
    for (const id of ids) {
      if (existing.has(id)) {
        skipped++
        continue
      }
      const found = discoveredModels.value.find(x => x.id === id)
      // 单次输出上限（不是上下文窗口）。
      // 优先用上游给的值：官方 API 会在 /models 里直接返回 max_output_tokens
      // （DeepSeek 官方实测 deepseek-flash → 393216）。中转多数不透传，此时才查本地表；
      // 表也没命中就**留 0（未设置）**，由后端决定不发送该字段 —— 不再编一个兜底数字。
      const upstreamOut = found?.max_output_tokens
      const local = resolveMaxOutput(id)
      const maxTokens = upstreamOut || local.value
      const maxTokensSource = upstreamOut ? 'upstream' : local.source
      try {
        const res: any = await configAPI.create({
          provider_id: pid,
          name: id,
          model: id,
          config_type: mlTypeMap.value[id] || guessConfigType(id),
          max_tokens: maxTokens,
          max_tokens_source: maxTokensSource,
          temperature: 0.7,
          top_p: 0.9,
          frequency_penalty: 0,
          enabled: 1,
          is_default: 0
        } as any)
        if (res?.success) {
          created++
          existing.add(id)
        } else {
          failed++
        }
      } catch (error) {
        failed++
      }
    }
    await fetchModels()

    const msg = `导入完成：成功 ${created}，跳过 ${skipped}${failed ? `，失败 ${failed}` : ''}`
    if (failed) ElMessage.warning(msg)
    else ElMessage.success(msg)
    modelListDialogVisible.value = false
  } finally {
    importingModels.value = false
  }
}

/** 单选：导入当前勾选的这一个 */
const importSelectedModels = () => importModels(selectedModelIds.value.slice())

/**
 * 多选：「全部导入」= 把这 N 个**选中的**模型一起导入。
 *
 * 旧的同名按钮是"导入渠道里的全部模型"，与「导入选中」语义重叠又危险
 * （用户以为在导 2 个，实际把 40 个都导进来了）。现在两者都只作用于勾选项，
 * 区别只在数量门槛：1 个走「导入选中」，≥2 个走「全部导入」。
 */
const importSelectedBulk = () => importModels(selectedModelIds.value.slice())

/* ---------------- 导入导出 ---------------- */
const importDialogVisible = ref(false)
const importPreview = ref<{
  version: string
  exportTime: string
  providers: Partial<ApiProvider>[]
  models: Partial<ApiModel>[]
} | null>(null)
const importOptions = ref({ overrideExisting: false })

const handleImportCommand = (command: string) => {
  if (command === 'importConfig') openImportDialog()
}

const handleExportCommand = (command: string) => {
  if (command === 'exportAll') executeExportAll()
}

const openImportDialog = () => {
  importPreview.value = null
  importOptions.value.overrideExisting = false
  importDialogVisible.value = true
}

const handleFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const data = JSON.parse(content)
      if (data.type === 'api-config' && (data.providers || data.models)) {
        importPreview.value = {
          version: data.version || '1.0',
          exportTime: data.exportTime || '未知',
          providers: data.providers || [],
          models: data.models || []
        }
        ElMessage.success(`成功读取配置文件：${data.providers?.length || 0} 个服务商，${data.models?.length || 0} 个模型`)
      } else {
        ElMessage.error('文件格式不正确，请上传API配置文件')
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的JSON文件')
    }
  }
  reader.readAsText(file.raw)
}

const executeImport = async () => {
  if (!importPreview.value) return
  const { providers: providersToImport, models: modelsToImport } = importPreview.value
  const overrideExisting = importOptions.value.overrideExisting
  let providerSuccessCount = 0
  let providerSkipCount = 0
  let modelSuccessCount = 0
  let modelSkipCount = 0
  const providerNameToId: Record<string, number> = {}

  for (const provider of providersToImport) {
    if (!provider.name) continue
    const existingProvider = providers.value.find(p => p.name === provider.name)
    if (existingProvider) {
      if (overrideExisting) {
        try {
          const res = await providerAPI.update(existingProvider.id, {
            name: provider.name,
            provider_type: provider.provider_type || 'openai',
            api_key: provider.api_key || '',
            api_url: provider.api_url || ''
          })
          if (res.success) {
            providerNameToId[provider.name] = existingProvider.id
            providerSuccessCount++
          }
        } catch (error) {
          providerSkipCount++
        }
      } else {
        providerNameToId[provider.name] = existingProvider.id
        providerSkipCount++
      }
    } else {
      try {
        const res = await providerAPI.create({
          name: provider.name,
          provider_type: provider.provider_type || 'openai',
          api_key: provider.api_key || '',
          api_url: provider.api_url || ''
        })
        if (res.success && res.data) {
          providerNameToId[provider.name] = res.data.id
          providerSuccessCount++
        }
      } catch (error) {
        providerSkipCount++
      }
    }
  }

  await fetchProviders()

  for (const model of modelsToImport) {
    if (!model.name || !model.model) continue
    let providerId = model.provider_id
    const providerName = (model as any).provider_name
    if (providerName && providerNameToId[providerName]) {
      providerId = providerNameToId[providerName]
    }
    if (!providerId) {
      modelSkipCount++
      continue
    }
    const existingModel = models.value.find(m => m.name === model.name && m.provider_id === providerId)
    if (existingModel) {
      if (overrideExisting) {
        try {
          const res = await configAPI.update(existingModel.id, {
            provider_id: providerId,
            name: model.name,
            model: model.model,
            temperature: model.temperature || 0.7,
            max_tokens: model.max_tokens || 2000,
            top_p: model.top_p ?? 0.9,
            frequency_penalty: model.frequency_penalty ?? 0.0,
            description: model.description ?? '',
            is_default: model.is_default ? 1 : 0,
            enabled: model.enabled != null ? model.enabled : 1,
            category: model.category || '',
            sort_order: model.sort_order ?? 0
          })
          if (res.success) modelSuccessCount++
        } catch (error) {
          modelSkipCount++
        }
      } else {
        modelSkipCount++
      }
    } else {
      try {
        const res = await configAPI.create({
          provider_id: providerId,
          name: model.name,
          model: model.model,
          temperature: model.temperature || 0.7,
          max_tokens: model.max_tokens || 2000,
          top_p: model.top_p ?? 0.9,
          frequency_penalty: model.frequency_penalty ?? 0.0,
          description: model.description ?? '',
          is_default: model.is_default ? 1 : 0,
          enabled: model.enabled != null ? model.enabled : 1,
          category: model.category || '',
          sort_order: model.sort_order ?? 0
        })
        if (res.success) modelSuccessCount++
      } catch (error) {
        modelSkipCount++
      }
    }
  }

  await fetchModels()
  ElMessage.success(`导入完成：服务商 ${providerSuccessCount} 个成功${providerSkipCount > 0 ? `，${providerSkipCount} 个跳过` : ''}；模型 ${modelSuccessCount} 个成功${modelSkipCount > 0 ? `，${modelSkipCount} 个跳过` : ''}`)
  importDialogVisible.value = false
}

const executeExportAll = () => {
  const exportData = {
    type: 'api-config',
    version: '1.0',
    exportTime: new Date().toLocaleString('zh-CN'),
    providers: providers.value.map(p => ({
      name: p.name,
      provider_type: p.provider_type,
      api_key: p.api_key,
      api_url: p.api_url
    })),
    models: models.value.map(m => ({
      name: m.name,
      model: m.model,
      provider_id: m.provider_id,
      provider_name: m.provider_name,
      temperature: m.temperature,
      max_tokens: m.max_tokens,
      top_p: m.top_p,
      frequency_penalty: m.frequency_penalty,
      description: m.description,
      is_default: m.is_default,
      enabled: m.enabled,
      category: m.category,
      sort_order: m.sort_order
    }))
  }

  const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `API配置_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success(`成功导出 ${providers.value.length} 个服务商和 ${models.value.length} 个模型`)
}
</script>

<style scoped>
.config-container {
  padding: 24px;
  /* 主题变量统一定义在全局 :root（含暗色覆盖），此处不再声明，
     否则 .config-container 自身的声明会遮蔽 :root[data-theme='dark'] 的暗色变量 */
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 14px;
}

.header-content h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
}

.header-subtitle {
  margin: 0;
  color: var(--cfg-ink-2);
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

/* 顶部操作按钮统一主题色：导入/导出 = 浅主题底 + 主题描边；新建模型 = 主题实底 */
.header-actions :deep(.el-button) {
  border-radius: 9px;
  font-weight: 500;
}

.header-actions :deep(.el-button--success),
.header-actions :deep(.el-button--warning) {
  --el-button-bg-color: var(--cfg-accent-bg);
  --el-button-border-color: var(--cfg-accent-line);
  --el-button-text-color: var(--cfg-accent-deep);
  --el-button-hover-bg-color: var(--cfg-accent-bg);
  --el-button-hover-border-color: var(--cfg-accent);
  --el-button-hover-text-color: var(--cfg-accent);
  --el-button-active-bg-color: var(--cfg-accent-bg);
  --el-button-active-border-color: var(--cfg-accent);
}

.header-actions :deep(.el-button--primary) {
  --el-button-bg-color: var(--cfg-accent);
  --el-button-border-color: var(--cfg-accent);
  --el-button-text-color: #ffffff;
  --el-button-hover-bg-color: var(--cfg-accent-hover);
  --el-button-hover-border-color: var(--cfg-accent-hover);
  --el-button-active-bg-color: var(--cfg-accent-hover);
  --el-button-active-border-color: var(--cfg-accent-hover);
}

/* 状态栏 */
.statusbar {
  display: flex;
  gap: 18px;
  align-items: center;
  background: var(--cfg-panel);
  border: 1px solid var(--cfg-line);
  border-radius: 12px;
  padding: 12px 18px;
  margin-bottom: 16px;
  font-size: 13px;
  color: var(--cfg-ink-2);
  flex-wrap: wrap;
}

.statusbar .sb-item b {
  color: var(--cfg-ink);
}

.statusbar .sb-item b.green {
  color: var(--cfg-accent);
}

.statusbar .sb-sep {
  width: 1px;
  height: 18px;
  background: var(--cfg-line);
}

.statusbar .pill {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  border-radius: 999px;
  padding: 4px 12px;
  font-weight: 600;
  font-size: 12px;
}

.statusbar .pill.dirty {
  background: var(--cfg-warn-bg);
  color: var(--cfg-warn-ink);
}

/* 视图切换 */
.view-switch {
  display: flex;
  gap: 4px;
  background: var(--cfg-switch-bg);
  border-radius: 10px;
  padding: 4px;
  width: fit-content;
  margin-bottom: 16px;
}

.vs-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: none;
  background: transparent;
  border-radius: 7px;
  padding: 8px 16px;
  font-size: 13.5px;
  color: var(--cfg-ink-2);
  cursor: pointer;
  font-family: inherit;
  transition: 0.15s;
}

.vs-btn svg {
  width: 15px;
  height: 15px;
}

.vs-btn:hover {
  color: var(--cfg-ink);
}

.vs-btn.on {
  background: var(--cfg-panel);
  color: var(--cfg-accent);
  font-weight: 600;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

/* 双栏工作台 */
.ws {
  display: grid;
  grid-template-columns: 380px 1fr;
  gap: 18px;
  align-items: start;
}

@media (max-width: 1000px) {
  .ws {
    grid-template-columns: 1fr;
  }
}

.pane {
  background: var(--cfg-panel);
  border: 1px solid var(--cfg-line);
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.03);
}

.pane-h {
  display: flex;
  align-items: center;
  gap: 9px;
  padding: 14px 18px;
  border-bottom: 1px solid var(--cfg-line-soft);
  font-size: 13px;
  font-weight: 700;
  color: var(--cfg-ink-2);
}

.pane-h .ic {
  width: 26px;
  height: 26px;
  border-radius: 8px;
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent);
  display: flex;
  align-items: center;
  justify-content: center;
}

.pane-h .ic svg {
  width: 14px;
  height: 14px;
}

.pane-h .x {
  margin-left: auto;
  font-size: 11px;
  color: var(--cfg-ink-3);
  font-weight: 400;
}

/* 左栏 */
.left {
  display: flex;
  flex-direction: column;
  max-height: 700px;
}

.cats {
  display: flex;
  gap: 6px;
  padding: 12px 14px;
  flex-wrap: wrap;
  border-bottom: 1px solid var(--cfg-line-soft);
}

.ctab {
  font-size: 11.5px;
  color: var(--cfg-ink-2);
  background: var(--cfg-panel-2);
  border: 1px solid var(--cfg-line);
  border-radius: 999px;
  padding: 5px 12px;
  cursor: pointer;
  transition: 0.15s;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  user-select: none;
}

.ctab .cnt {
  font-size: 10px;
  color: var(--cfg-ink-3);
}

.ctab:hover {
  border-color: var(--cfg-ink-3);
}

.ctab.on {
  background: var(--cfg-accent);
  border-color: var(--cfg-accent);
  color: #fff;
}

.ctab.on .cnt {
  color: rgba(255, 255, 255, 0.75);
}

.mbody {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.mrow {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  border: 1px solid transparent;
  transition: 0.13s;
}

.mrow:hover {
  background: var(--cfg-hover);
}

.mrow.on {
  background: var(--cfg-accent-bg);
  border-color: var(--cfg-accent-line);
}

.mrow .st {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  flex-shrink: 0;
}

.mrow-main {
  flex: 1;
  min-width: 0;
}

.mrow .nm {
  font-size: 13.5px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mrow .nm .star {
  color: var(--cfg-star);
  margin-left: 3px;
}

.mrow .pv {
  font-size: 11px;
  color: var(--cfg-ink-3);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.mrow .en {
  margin-left: auto;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  font-weight: 600;
  flex-shrink: 0;
}

.mrow .en.off {
  background: var(--cfg-panel-2);
  color: var(--cfg-ink-3);
}

.empty {
  text-align: center;
  color: var(--cfg-ink-3);
  font-size: 13px;
  padding: 44px 10px;
  line-height: 2;
}

.empty-sub {
  font-size: 11.5px;
  color: var(--cfg-ink-3);
}

/* 右栏编辑器 */
.editor {
  padding: 22px 24px;
}

.e-top {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--cfg-line-soft);
}

.e-ic {
  width: 52px;
  height: 52px;
  border-radius: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.e-nm {
  font-size: 19px;
  font-weight: 800;
  display: flex;
  align-items: center;
  gap: 9px;
}

.e-nm .def {
  font-size: 10px;
  background: var(--cfg-default-bg);
  color: var(--cfg-default-ink);
  border-radius: 999px;
  padding: 2px 9px;
}

.e-pv {
  font-size: 12.5px;
  color: var(--cfg-ink-3);
  margin-top: 4px;
}

.e-badges {
  display: flex;
  gap: 7px;
  margin-top: 10px;
}

.e-badges .b {
  font-size: 11px;
  background: var(--cfg-panel-2);
  color: var(--cfg-ink-2);
  border: 1px solid var(--cfg-line);
  border-radius: 999px;
  padding: 3px 11px;
}

.e-badges .b.b-on {
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  border-color: var(--cfg-accent-line);
}

.e-badges .b.b-off {
  background: var(--cfg-danger-bg);
  color: var(--cfg-danger);
  border-color: var(--cfg-danger-line);
}

.form {
  padding-top: 6px;
}

.fg-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

@media (max-width: 560px) {
  .fg-row {
    grid-template-columns: 1fr;
  }
}

.fg {
  margin-top: 16px;
}

.fg label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  color: var(--cfg-ink-2);
  margin-bottom: 7px;
}

.fg-hint {
  margin-top: 6px;
  font-size: 11.5px;
  line-height: 1.5;
  color: var(--cfg-ink-3);
}

/* 高级参数折叠区 */
.adv {
  margin-top: 18px;
  border: 1px solid var(--cfg-line);
  border-radius: 12px;
  background: var(--cfg-panel-2);
  overflow: hidden;
}

.adv-h {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 11px 14px;
  border: none;
  background: transparent;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  color: var(--cfg-ink-2);
  cursor: pointer;
}

.adv-h:hover {
  color: var(--cfg-ink);
}

.adv-body {
  padding: 2px 14px 16px;
  border-top: 1px solid var(--cfg-line-soft);
}

.discover-row .discover-inner {
  display: flex;
  align-items: center;
  gap: 12px;
}

.d-hint {
  font-size: 12px;
  color: var(--cfg-accent-deep);
}

.d-hint.link {
  cursor: pointer;
  text-decoration: underline;
  text-underline-offset: 3px;
}

.d-hint.link:hover {
  color: var(--cfg-accent);
}

.slider-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 2px;
}

.slider-row label {
  margin-bottom: 0;
}

.slider-row .v {
  font-size: 12.5px;
  color: var(--cfg-accent);
  font-family: Consolas, monospace;
  font-weight: 600;
}

/* 滑块主色适配主题 */
.fg.slider :deep(.el-slider) {
  --el-slider-main-bg-color: var(--cfg-accent);
  --el-slider-runway-bg-color: var(--cfg-slider-track);
  margin-top: 4px;
}

/* 开关颜色随主题变量 */
.tcard :deep(.el-switch) {
  --el-switch-on-color: var(--cfg-accent);
  --el-switch-off-color: var(--cfg-switch-off);
}

.toggle-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 18px;
}

@media (max-width: 520px) {
  .toggle-grid {
    grid-template-columns: 1fr;
  }
}

.tcard {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border: 1px solid var(--cfg-line);
  border-radius: 11px;
  padding: 12px 14px;
}

.tcard .t {
  font-size: 13px;
  font-weight: 600;
}

.tcard .d {
  font-size: 11px;
  color: var(--cfg-ink-3);
  margin-top: 2px;
}

.e-ops {
  display: flex;
  gap: 10px;
  margin-top: 22px;
  flex-wrap: wrap;
}

.big {
  flex: 1;
  min-width: 110px;
  padding: 11px;
  border-radius: 10px;
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  border: 1px solid transparent;
  text-align: center;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: 0.15s;
}

.big svg {
  width: 14px;
  height: 14px;
}

.big.save {
  background: var(--cfg-accent);
  color: #fff;
}

.big.save:hover {
  background: var(--cfg-accent-hover);
}

.big.test {
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  border-color: var(--cfg-accent-line);
}

.big.test:hover {
  background: var(--cfg-accent-bg);
}

.big.del {
  background: var(--cfg-danger-bg);
  color: var(--cfg-danger);
  border-color: var(--cfg-danger-line);
}

.big.del:hover {
  background: var(--cfg-danger-hover);
}

.big.cancel {
  background: var(--cfg-panel-2);
  color: var(--cfg-ink-2);
  border-color: var(--cfg-line);
}

.e-empty {
  text-align: center;
  color: var(--cfg-ink-3);
  padding: 90px 20px;
  line-height: 2;
}

.e-empty-ic {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: var(--cfg-panel-2);
  color: var(--cfg-ink-3);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.e-empty-ic svg {
  width: 26px;
  height: 26px;
}

/* 服务商视图 */
.providers-view {
  background: var(--cfg-panel);
  border: 1px solid var(--cfg-line);
  border-radius: 16px;
  padding: 20px;
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.tab-header h3 {
  margin: 0;
  font-size: 17px;
}

.model-add-content {
  padding-top: 8px;
}

.custom-config-hint {
  margin-bottom: 16px;
}

.form-item-hint {
  font-size: 12px;
  color: var(--cfg-ink-3);
  margin-top: 4px;
  line-height: 1.5;
}

.url-mode-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 导入 */
.import-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.import-upload {
  width: 100%;
}

.import-preview {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-info {
  line-height: 1.8;
}

.preview-info p {
  margin: 4px 0;
}

.import-options {
  padding: 12px 16px;
  background: var(--cfg-panel-2);
  border-radius: 8px;
}

@media (max-width: 768px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
  }
}

/* 暗色下编辑器内的 Element Plus 控件（scoped 可达） */
:root[data-theme='dark'] .config-container .editor :deep(.el-select__wrapper) {
  background: var(--cfg-panel-2);
  box-shadow: 0 0 0 1px var(--cfg-line) inset;
}

:root[data-theme='dark'] .config-container .editor :deep(.el-select__wrapper.is-focused),
:root[data-theme='dark'] .config-container .editor :deep(.el-select__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--cfg-accent) inset;
}

:root[data-theme='dark'] .config-container .editor :deep(.el-input__wrapper),
:root[data-theme='dark'] .config-container .editor :deep(.el-textarea__inner) {
  background: var(--cfg-panel-2);
}

/* 暗色下阴影增强（浅色 rgba 黑阴影在暗底上不可见） */
:root[data-theme='dark'] .config-container .pane,
:root[data-theme='dark'] .config-container .providers-view {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
}

:root[data-theme='dark'] .config-container .vs-btn.on {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

/* ================= 服务商列表 ================= */
.providers-view {
  background: var(--cfg-panel);
  border: 1px solid var(--cfg-line);
  border-radius: 16px;
  padding: 20px;
}

.pv-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.pv-title {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.pv-title h3 {
  margin: 0;
  font-size: 17px;
  color: var(--cfg-ink);
}

.pv-sub {
  font-size: 12px;
  color: var(--cfg-ink-3);
}

.pv-prov {
  display: flex;
  align-items: center;
  gap: 11px;
}

.pv-lg {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 800;
  font-size: 14px;
  flex-shrink: 0;
}

.pv-meta {
  min-width: 0;
}

.pv-nm {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--cfg-ink);
}

.pv-tp {
  font-size: 11px;
  color: var(--cfg-ink-3);
}

.pv-url {
  font-family: Consolas, monospace;
  font-size: 12px;
  color: var(--cfg-ink-2);
}

.pv-cnt {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  padding: 2px 9px;
  border-radius: 999px;
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  font-size: 12px;
  font-weight: 700;
}

/* el-table 主题化（对齐工作台风格） */
.providers-view :deep(.el-table) {
  --el-table-border-color: var(--cfg-line-soft);
  --el-table-header-bg-color: var(--cfg-panel-2);
  --el-table-header-text-color: var(--cfg-ink-2);
  --el-table-text-color: var(--cfg-ink);
  --el-table-row-hover-bg-color: var(--cfg-hover);
  --el-table-bg-color: var(--cfg-panel);
  --el-table-tr-bg-color: var(--cfg-panel);
  border-radius: 12px;
  overflow: hidden;
}

.providers-view :deep(.el-table th.el-table__cell) {
  font-size: 12px;
  font-weight: 600;
}

.providers-view :deep(.el-table--striped .el-table__row--striped td) {
  background: var(--cfg-panel-2);
}

.providers-view :deep(.el-table__inner-wrapper::before) {
  display: none;
}
</style>

<style>
/* ============ 主题变量（:root 级，页面与弹窗共用） ============ */
:root {
  --cfg-panel: #ffffff;
  --cfg-panel-2: #f4f7f4;
  --cfg-hover: #f7faf7;
  --cfg-line: #e2e8e2;
  --cfg-line-soft: #edf2ed;
  --cfg-ink: #22303c;
  --cfg-ink-2: #55635c;
  --cfg-ink-3: #94a3a0;
  --cfg-accent: #16a34a;
  --cfg-accent-deep: #059669;
  --cfg-accent-hover: #15803d;
  --cfg-accent-bg: #ecfdf5;
  --cfg-accent-line: #bbf7d0;
  --cfg-switch-bg: #eef1ef;
  --cfg-switch-off: #d5ded6;
  --cfg-slider-track: #e2e8e2;
  --cfg-star: #eab308;
  --cfg-default-bg: #fef3c7;
  --cfg-default-ink: #b45309;
  --cfg-warn-bg: #fff7ed;
  --cfg-warn-ink: #c2410c;
  --cfg-danger: #dc2626;
  --cfg-danger-bg: #fef2f2;
  --cfg-danger-hover: #fee2e2;
  --cfg-danger-line: #fecaca;
}

:root[data-theme='dark'] {
  --cfg-panel: #161b22;
  --cfg-panel-2: #1c2333;
  --cfg-hover: #1c2333;
  --cfg-line: rgba(59, 130, 246, 0.16);
  --cfg-line-soft: rgba(59, 130, 246, 0.10);
  --cfg-ink: #e6edf3;
  --cfg-ink-2: #8b98a8;
  --cfg-ink-3: #5f6b7a;
  --cfg-accent: #3b82f6;
  --cfg-accent-deep: #60a5fa;
  --cfg-accent-hover: #2563eb;
  --cfg-accent-bg: rgba(59, 130, 246, 0.12);
  --cfg-accent-line: rgba(59, 130, 246, 0.30);
  --cfg-switch-bg: #21262d;
  --cfg-switch-off: #2a3444;
  --cfg-slider-track: #2a3444;
  --cfg-star: #fbbf24;
  --cfg-default-bg: rgba(251, 191, 36, 0.15);
  --cfg-default-ink: #fbbf24;
  --cfg-warn-bg: rgba(251, 146, 60, 0.15);
  --cfg-warn-ink: #fb923c;
  --cfg-danger: #f87171;
  --cfg-danger-bg: rgba(248, 113, 113, 0.12);
  --cfg-danger-hover: rgba(248, 113, 113, 0.22);
  --cfg-danger-line: rgba(248, 113, 113, 0.30);
}

/* ============ 弹窗整体风格（统一升级） ============ */
/* 「固定弹窗」的关键：整窗 flex 纵向排布 + max-height 限制在视口内。
   header / footer 不参与滚动（flex-shrink: 0），只有 body 内部滚动。
   B 项目原先没有这层，内容一长弹窗就撑出视口，标题和按钮跟着滚走。
   尺寸变量定义在 src/style.css（必须全局，组件卸载时变量会跟着消失）。 */
.cfg-dialog.el-dialog {
  /* Element 默认 --el-dialog-margin-top: 15vh（弹窗吊在半空、且极易溢出视口），
     蓝本改成 gutter。上下留白都取 gutter，配合 max-height 正好贴合视口。 */
  --el-dialog-margin-top: var(--app-dialog-gutter);
  display: flex;
  flex-direction: column;
  margin-bottom: var(--app-dialog-gutter);
  max-width: var(--app-dialog-available-width);
  max-height: var(--app-dialog-available-height);
  padding: 0;
  border: 1px solid var(--cfg-line-soft);
  border-radius: 8px;
  overflow: hidden;
}

.cfg-dialog .el-dialog__header {
  flex-shrink: 0;
  margin: 0;
  padding: 12px 16px;
  border-bottom: 1px solid var(--cfg-line-soft);
}

.cfg-dialog .el-dialog__title {
  font-size: 14px;
  font-weight: 600;
  color: var(--cfg-ink);
}

/* 关闭按钮铺满 header 右上角（44px = header 高度），
   否则 Element 默认的 54px 尺寸会压到 body 上 */
.cfg-dialog .el-dialog__headerbtn {
  top: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
}

.cfg-dialog .el-dialog__body {
  flex: 1 1 auto;
  min-height: 0;
  padding: 14px 16px;
  overflow: auto;
  color: var(--cfg-ink);
}

.cfg-dialog .el-dialog__footer {
  flex-shrink: 0;
  padding: 12px 16px;
  border-top: 1px solid var(--cfg-line-soft);
}

.cfg-dialog .el-form-item__label {
  color: var(--cfg-ink-2);
  font-weight: 600;
}

.cfg-dialog .el-form-item {
  margin-bottom: 18px;
}

.cfg-dialog .el-switch {
  --el-switch-on-color: var(--cfg-accent);
  --el-switch-off-color: var(--cfg-switch-off);
}

.cfg-dialog .el-input__wrapper,
.cfg-dialog .el-textarea__inner {
  border-radius: 9px;
}

/* 服务商弹窗 · 卡片直选（demo-06 风格） */
.cfg-dialog .pseg {
  display: flex;
  background: var(--cfg-panel-2);
  border-radius: 10px;
  padding: 4px;
  gap: 4px;
  margin-bottom: 18px;
}

.cfg-dialog .pseg button {
  flex: 1;
  border: none;
  background: transparent;
  padding: 8px 0;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--cfg-ink-2);
  cursor: pointer;
  font-family: inherit;
  transition: 0.18s;
}

.cfg-dialog .pseg button:hover {
  color: var(--cfg-accent-deep);
}

.cfg-dialog .pseg button.on {
  background: var(--cfg-panel);
  color: var(--cfg-accent-deep);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.cfg-dialog .pcard-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
}

.cfg-dialog .pcard {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 12px 4px 10px;
  border: 1.5px solid var(--cfg-line);
  border-radius: 12px;
  cursor: pointer;
  transition: 0.15s;
  background: var(--cfg-panel);
  user-select: none;
}

.cfg-dialog .pcard:hover {
  border-color: var(--cfg-accent-line);
  transform: translateY(-1px);
}

.cfg-dialog .pcard.on {
  border-color: var(--cfg-accent);
  background: var(--cfg-accent-bg);
  box-shadow: 0 0 0 3px var(--cfg-accent-line);
}

.cfg-dialog .pcard-av {
  width: 30px;
  height: 30px;
  border-radius: 9px;
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  font-weight: 800;
  font-size: 12.5px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: 0.15s;
}

.cfg-dialog .pcard.on .pcard-av {
  background: var(--cfg-accent);
  color: #fff;
}

.cfg-dialog .pcard-lb {
  font-size: 11px;
  color: var(--cfg-ink-2);
  font-weight: 600;
  text-align: center;
  line-height: 1.3;
}

.cfg-dialog .pcard.on .pcard-lb {
  color: var(--cfg-accent-deep);
}

.cfg-dialog .pcard-reveal {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid var(--cfg-line);
}

.cfg-dialog .pcard-chosen {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--cfg-ink-2);
  background: var(--cfg-accent-bg);
  border: 1px solid var(--cfg-accent-line);
  border-radius: 9px;
  padding: 8px 12px;
  margin-bottom: 12px;
}

.cfg-dialog .pcard-chosen b {
  color: var(--cfg-accent-deep);
  font-weight: 600;
}

.reveal-enter-active,
.reveal-leave-active {
  transition: 0.22s ease;
  overflow: hidden;
}

.reveal-enter-from,
.reveal-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}

/* 弹窗内控件统一主题色 */
.cfg-dialog .el-button--primary {
  --el-button-bg-color: var(--cfg-accent);
  --el-button-border-color: var(--cfg-accent);
  --el-button-text-color: #ffffff;
  --el-button-hover-bg-color: var(--cfg-accent-hover);
  --el-button-hover-border-color: var(--cfg-accent-hover);
  --el-button-active-bg-color: var(--cfg-accent-hover);
  --el-button-active-border-color: var(--cfg-accent-hover);
}

.cfg-dialog .el-alert--info {
  background: var(--cfg-accent-bg);
  border-color: var(--cfg-accent-line);
}

.cfg-dialog .el-alert--info .el-alert__title,
.cfg-dialog .el-alert__description {
  color: var(--cfg-accent-deep);
}

.cfg-dialog .el-switch {
  --el-switch-on-color: var(--cfg-accent);
}

.cfg-dialog .el-select__wrapper.is-focused,
.cfg-dialog .el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px var(--cfg-accent) inset;
}

.cfg-dialog .el-form-item__label {
  color: var(--cfg-ink-2);
}

.cfg-dialog .el-checkbox__label {
  color: var(--cfg-ink-2);
}

/* 下拉弹层 / 弹窗内 Element Plus 控件的暗色覆盖（teleport 到 body，scoped 不可达） */
:root[data-theme='dark'] .cfg-select-popper {
  --el-bg-color-overlay: #161b22;
  --el-bg-color: #161b22;
  --el-fill-color-blank: #161b22;
  --el-fill-color-light: #1c2333;
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #8b98a8;
  --el-border-color-light: rgba(59, 130, 246, 0.16);
}

:root[data-theme='dark'] .cfg-select-popper .el-select-dropdown {
  border: 1px solid rgba(59, 130, 246, 0.16);
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.5);
}

:root[data-theme='dark'] .cfg-select-popper .el-select-dropdown__item {
  color: #8b98a8;
}

:root[data-theme='dark'] .cfg-select-popper .el-select-dropdown__item:hover {
  background: #1c2333;
  color: #e6edf3;
}

:root[data-theme='dark'] .cfg-select-popper .el-select-dropdown__item.is-selected {
  color: #3b82f6;
}

/* 导入 / 服务商弹窗 */
:root[data-theme='dark'] .cfg-dialog .preview-info p {
  color: #8b98a8;
}

:root[data-theme='dark'] .cfg-dialog .el-upload-dragger {
  background: #1c2333;
  border-color: rgba(59, 130, 246, 0.16);
}

:root[data-theme='dark'] .cfg-dialog .el-upload-dragger:hover {
  border-color: #3b82f6;
}

:root[data-theme='dark'] .cfg-dialog .el-upload__text {
  color: #8b98a8;
}

:root[data-theme='dark'] .cfg-dialog .el-upload__tip {
  color: #5f6b7a;
}

/* 获取到的模型列表弹窗（浅色默认） */
.cfg-dialog .model-list-dialog {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* 暗色：获取到的模型列表弹窗 */
</style>

<style>
/* ============ 复刻自 xingnovel 0.3.0 的 API 设置页样式 ============
   来源 web/css/Config-BoY9XKxz.css，已剥离 scoped 的 [data-v-*] 属性。
   整段保持非 scoped：原样式有 50 条规则把 data-v 挂在祖先选择器上以命中
   Element 组件内部（如 .cap-code .el-textarea__inner），若交给 scoped 编译，
   Vue 会把 data-v 挪到末尾选择器上导致这些规则失效。
   类名均带 config-page__ / plaza__ / cap- / quote__ 前缀，全局污染可忽略。 */
/* 复刻页所需的变量别名：映射到 B 项目主题变量，跟随亮/暗主题自动切换 */
:root {
  --app-bg-color: var(--cfg-panel);
  --app-border-color: var(--cfg-line);
  --app-border-color-light: var(--cfg-line-soft);
  --app-color-danger: var(--cfg-danger);
  --app-color-primary: var(--cfg-accent);
  --app-color-primary-contrast: #ffffff;
  --app-fill-color: var(--cfg-panel-2);
  --app-fill-color-dark: var(--cfg-hover);
  --app-fill-color-light: var(--cfg-panel-2);
  --app-text-color-primary: var(--cfg-ink);
  --app-text-color-regular: var(--cfg-ink-2);
  --app-text-color-secondary: var(--cfg-ink-2);
  --workbench-page-gutter: 24px;
  --workbench-selected: var(--cfg-accent-bg);
}

.plaza {
  display: flex;
  flex-direction: column;
  gap: 14px;
  flex: 1 1 0;
  min-height: 0
}

.plaza__toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap
}

.plaza__title {
  margin: 0 0 2px;
  font-size: 16px;
  font-weight: 700;
  color: var(--app-text-color-primary)
}

.plaza__subtitle {
  margin: 0;
  font-size: 12px;
  color: var(--app-text-color-secondary)
}

.plaza__search {
  width: 260px
}

.plaza__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  flex: 1 1 0;
  min-height: 0
}

.plaza__row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 14px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 12px;
  background: var(--app-bg-color);
  cursor: pointer;
  text-align: left;
  transition: all .15s ease
}

.plaza__row:hover {
  border-color: var(--app-color-primary);
  background: var(--app-fill-color-light)
}

.plaza__avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  flex-shrink: 0;
  border-radius: 10px;
  color: #fff;
  font-size: 16px;
  font-weight: 700
}

.plaza__avatar--lg {
  width: 48px;
  height: 48px;
  font-size: 20px;
  border-radius: 12px
}

.plaza__row-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1 1 0;
  min-width: 0
}

.plaza__row-name {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--app-text-color-primary)
}

.plaza__row-connected {
  display: inline-flex;
  align-items: center;
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  color: #0d9488;
  background: #0d94881a
}

.plaza__row-desc {
  font-size: 12px;
  color: var(--app-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap
}

.plaza__row-stats {
  display: flex;
  gap: 18px;
  flex-shrink: 0
}

.plaza__stat,
.plaza__stat-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
  min-width: 44px
}

.plaza__stat b,
.plaza__stat-box b {
  font-size: 15px;
  color: var(--app-text-color-primary)
}

.plaza__stat i,
.plaza__stat-box i {
  font-style: normal;
  font-size: 11px;
  color: var(--app-text-color-secondary)
}

.plaza__row-arrow {
  color: var(--app-text-color-secondary);
  flex-shrink: 0
}

.plaza__detail {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1 1 0;
  min-height: 0
}

.plaza__back {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  width: fit-content;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 13px;
  color: var(--app-text-color-secondary);
  cursor: pointer
}

.plaza__back:hover {
  color: var(--app-color-primary)
}

.plaza__detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap
}

.plaza__detail-info {
  display: flex;
  gap: 12px;
  min-width: 0
}

.plaza__detail-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0
}

.plaza__docs {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  width: fit-content;
  font-size: 12px;
  color: var(--app-color-primary);
  text-decoration: none
}

.plaza__docs:hover {
  text-decoration: underline
}

.plaza__detail-stats {
  display: flex;
  gap: 10px;
  flex-shrink: 0
}

.plaza__stat-box {
  justify-content: center;
  min-width: 64px;
  padding: 8px 12px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 10px;
  background: var(--app-fill-color-light)
}

.plaza__filter-tabs {
  display: flex;
  gap: 6px;
  flex-wrap: wrap
}

.plaza__filter-tab {
  padding: 4px 14px;
  font-size: 12px;
  border-radius: 999px;
  border: 1px solid var(--app-border-color);
  background: var(--app-fill-color);
  color: var(--app-text-color-regular);
  cursor: pointer;
  transition: all .15s ease
}

.plaza__filter-tab:hover {
  border-color: var(--app-color-primary);
  color: var(--app-text-color-primary)
}

.plaza__filter-tab.is-active {
  background: var(--app-color-primary);
  border-color: var(--app-color-primary);
  color: var(--app-color-primary-contrast)
}

.plaza__models {
  display: flex;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  flex: 1 1 0;
  min-height: 0
}

.plaza__model {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 12px;
  background: var(--app-bg-color)
}

.plaza__model-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 700
}

.plaza__model-main {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1 1 0;
  min-width: 0
}

.plaza__model-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--app-text-color-primary)
}

.plaza__model-id {
  font-size: 11px;
  font-weight: 400;
  color: var(--app-text-color-secondary);
  background: var(--app-fill-color);
  padding: 0 6px;
  border-radius: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 260px
}

.plaza__model-desc {
  font-size: 12px;
  color: var(--app-text-color-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap
}

.plaza__model-tags {
  display: flex;
  gap: 6px;
  flex-shrink: 0
}

.plaza__model-action {
  flex-shrink: 0
}

.quote__header {
  display: flex;
  flex-direction: column;
  gap: 2px
}

.quote {
  display: flex;
  flex-direction: column;
  gap: 14px
}

.quote__model-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 10px;
  background: var(--app-fill-color-light)
}

.quote__model-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0
}

.quote__model-text b {
  font-size: 14px;
  color: var(--app-text-color-primary)
}

.quote__model-text code {
  font-size: 12px;
  color: var(--app-text-color-secondary)
}

.quote__field {
  display: flex;
  flex-direction: column;
  gap: 4px
}

.quote__label {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-text-color-primary)
}

.quote__req {
  color: var(--app-color-danger)
}

.quote__hint {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--app-text-color-secondary)
}

@media(max-width:860px) {
  .plaza__row {
    flex-wrap: wrap
  }

  .plaza__row-stats {
    width: 100%;
    justify-content: flex-start
  }
}

.cap-header {
  display: flex;
  flex-direction: column;
  gap: 2px
}

.cap-header__sub {
  font-weight: 400
}

.cap-form {
  display: flex;
  flex-direction: column;
  gap: 12px
}

.cap-label {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  font-size: 13px;
  font-weight: 600;
  color: var(--app-text-color-primary);
  line-height: 1.4;
  margin-bottom: 4px
}

.cap-req {
  color: var(--app-color-danger);
  font-style: normal
}

.cap-hint {
  margin: 4px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--app-text-color-secondary)
}

.cap-error {
  margin: 4px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--app-color-danger)
}

.cap-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px
}

.cap-field {
  display: flex;
  flex-direction: column;
  min-width: 0
}

.cap-field--full {
  grid-column: 1/-1
}

.cap-inline {
  display: flex;
  align-items: center;
  gap: 8px
}

.cap-inline>.el-select {
  flex: 1 1 0;
  min-width: 0
}

.cap-switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 32px
}

.cap-panel {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 12px;
  background: var(--app-bg-color);
  min-width: 0
}

.cap-panel__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px
}

.cap-panel__title {
  min-width: 0
}

.cap-panel__title h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  color: var(--app-text-color-primary)
}

.cap-panel__title p {
  margin: 2px 0 0;
  font-size: 12px;
  color: var(--app-text-color-secondary)
}

.cap-panel__actions {
  display: flex;
  gap: 6px;
  flex-shrink: 0
}

.cap-panel__empty {
  padding: 18px 0;
  text-align: center;
  border: 1px dashed var(--app-border-color);
  border-radius: 10px
}

.cap-panel__foot {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px
}

.cap-param-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 10px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 10px;
  background: var(--app-fill-color-light)
}

.cap-param-group__head {
  display: flex;
  align-items: center;
  gap: 8px
}

.cap-param-group__name {
  flex: 1 1 0;
  min-width: 0
}

.cap-param-group__count {
  flex-shrink: 0;
  white-space: nowrap
}

.cap-param-options {
  display: flex;
  flex-direction: column;
  gap: 6px
}

.cap-param-option {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 24px;
  gap: 6px;
  align-items: center
}

.cap-param-option--head {
  font-size: 11px;
  color: var(--app-text-color-secondary);
  padding: 0 2px
}

.cap-param-option--head span:last-child {
  width: 24px
}

.cap-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  flex-shrink: 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--app-text-color-secondary);
  cursor: pointer;
  transition: all .15s ease
}

.cap-icon-btn:hover {
  background: var(--app-fill-color-dark, rgba(0, 0, 0, .06));
  color: var(--app-text-color-primary)
}

.cap-icon-btn--danger:hover {
  color: var(--app-color-danger)
}

.cap-code .el-textarea__inner {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6
}

/* 高级参数：常驻展开（无折叠），左栏滑动条 / 右栏其余控件 */
.cap-advanced {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 12px;
  background: var(--app-bg-color);
  min-width: 0
}

.cap-advanced__head {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0
}

.cap-advanced__head .el-icon {
  flex-shrink: 0;
  font-size: 14px;
  color: var(--app-text-color-secondary)
}

.cap-advanced__head h4 {
  flex-shrink: 0;
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  color: var(--app-text-color-primary)
}

.cap-advanced__tip {
  min-width: 0;
  margin-left: 2px;
  font-size: 11.5px;
  font-weight: 400;
  color: var(--app-text-color-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis
}

.cap-advanced__cols {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  align-items: stretch
}

.cap-col {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px 12px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 10px;
  background: var(--app-fill-color-light);
  min-width: 0
}

/* 列内说明文字收一档，别把卡片撑高 */
.cap-col .cap-hint {
  margin-top: 3px;
  font-size: 11.5px
}

.cap-col__title {
  flex-shrink: 0;
  font-size: 11.5px;
  font-weight: 700;
  letter-spacing: .04em;
  color: var(--app-text-color-secondary)
}

/* 滑动条列：每条压成「名称 + 轨道 + 数值」单行，纵向撑满左栏与右栏底边对齐 */
.cap-sliders {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  justify-content: space-between;
  gap: 6px;
  min-width: 0
}

.cap-slider {
  display: grid;
  grid-template-columns: 92px minmax(0, 1fr);
  align-items: center;
  gap: 8px;
  min-width: 0
}

.cap-slider>.cap-label {
  margin-bottom: 0
}

.cap-slider>.el-slider {
  min-width: 0
}

.cap-slider .el-slider {
  height: 32px
}

/* 滑条主色跟弹窗交互色（亮色绿 / 暗色蓝），与同区的开关保持一致；
   拖柄从 20px 收到 16px，避免在 32px 行里显笨重 */
.cap-advanced .el-slider {
  --el-slider-main-bg-color: var(--cfg-accent);
  --el-slider-height: 5px;
  --el-slider-button-size: 16px;
  --el-slider-button-wrapper-size: 28px;
  --el-slider-button-wrapper-offset: -11px
}

.cap-slider .el-slider__runway.show-input {
  margin-right: 10px
}

.cap-slider .el-slider__input,
.cap-slider .el-slider__input .el-input-number {
  width: 78px
}

/* 右栏：短字段两两并排，长文本 / 全宽项用 .cap-field--full 跨两列 */
.cap-pairs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px 12px;
  align-items: start;
  min-width: 0
}

.cap-preset-row {
  display: flex;
  align-items: center;
  gap: 8px
}

.cap-preset-row>.el-select {
  flex: 1 1 0;
  min-width: 0
}

.cap-hint-inline {
  font-weight: 400;
  font-size: 11px;
  color: var(--app-text-color-secondary)
}

/* 「最大 Token」的来源徽标。
   同一个数字可能来自官方文档、别人的默认配置、或某个下午有人随手写的一个
   "看起来挺安全"的数 —— 它们长得完全一样。必须把来源显式标出来。 */
.cap-src {
  margin-left: 6px;
  padding: 0 5px;
  border: 1px solid currentColor;
  border-radius: 4px;
  font-weight: 400;
  font-size: 11px;
  line-height: 16px;
  opacity: .85
}

.cap-src--upstream { color: #1d9e75 }

.cap-src--table,
.cap-src--catalog { color: #ba7517 }

.cap-src--manual { color: #378add }

.cap-src--probed { color: #7f77dd }

.cap-src--unknown { color: var(--app-text-color-secondary) }

.cap-protocol-option {
  display: flex;
  flex-direction: column;
  padding: 2px 0;
  line-height: 1.4
}

.cap-protocol-hint {
  font-size: 12px;
  color: var(--app-text-color-secondary)
}

.config-page {
  min-width: 0;
  min-height: 0;
  height: 100%;
  background: transparent;
  overflow: hidden
}

.config-page__content,
.config-page__scroll {
  min-width: 0
}

.config-page__scroll {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  overflow-x: hidden
}

.config-page__inner {
  display: flex;
  flex: 1 1 0;
  flex-direction: column;
  min-height: 0;
  padding: 16px var(--workbench-page-gutter)
}

.config-page__header {
  flex: 0 0 auto;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--app-border-color-light)
}

.config-page__heading {
  flex: 1 1 260px;
  min-width: 0
}

.config-page__heading h2 {
  font-size: 18px;
  font-weight: 600
}

.config-page__heading p {
  font-size: 12px;
  max-width: 480px
}

.config-page__nav-tabs {
  display: inline-flex;
  gap: 4px;
  padding: 3px;
  border-radius: 6px;
  background: var(--app-fill-color-light);
  width: fit-content
}

.config-page__nav-tab {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 18px;
  font-size: 13px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: var(--app-text-color-regular);
  cursor: pointer;
  transition: all .15s ease
}

.config-page__nav-tab:hover {
  color: var(--app-text-color-primary)
}

.config-page__nav-tab.is-active {
  background: var(--workbench-selected);
  color: var(--app-text-color-primary);
  font-weight: 600;
  box-shadow: none
}

.config-page__mine {
  display: flex;
  flex-direction: column;
  flex: 1 1 0;
  min-height: 0
}

.config-page__mine-head,
.config-page__mine-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap
}

.config-page__type-tabs {
  display: inline-flex;
  gap: 4px;
  padding: 3px;
  border-radius: 5px;
  background: var(--app-fill-color-light)
}

.config-page__type-tab {
  padding: 4px 16px;
  font-size: 12px;
  line-height: 1.5;
  border: none;
  border-radius: 5px;
  background: transparent;
  color: var(--app-text-color-regular);
  cursor: pointer;
  transition: all .15s ease
}

.config-page__type-tab:hover {
  color: var(--app-text-color-primary)
}

.config-page__type-tab.is-active {
  background: var(--workbench-selected);
  color: var(--app-text-color-primary);
  font-weight: 600;
  box-shadow: none
}

.config-page__mine-search {
  width: 220px
}

.config-page__relay-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 10px 14px;
  border: 1px solid color-mix(in srgb, var(--app-color-primary) 28%, transparent);
  border-radius: 8px;
  background: color-mix(in srgb, var(--app-color-primary) 6%, var(--app-bg-color))
}

.config-page__relay-info {
  flex: 1 1 0;
  min-width: 0
}

.config-page__api-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  flex: 1 1 0;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 4px
}

.config-page__api-group {
  display: flex;
  flex-direction: column;
  gap: 8px
}

.config-page__group-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  min-width: 0;
  padding: 2px 2px 0
}

.config-page__group-dot {
  width: 8px;
  height: 8px;
  border-radius: 5px;
  flex-shrink: 0
}

.config-page__group-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-text-color-primary)
}

.config-page__group-count {
  font-size: 11px;
  color: var(--app-text-color-secondary);
  padding: 1px 8px;
  border-radius: 5px;
  background: var(--app-fill-color-light)
}

.config-page__group-conc {
  font-size: 11px;
  color: var(--app-text-color-secondary)
}

.config-page__group-tag {
  font-size: 10px;
  font-weight: 600;
  padding: 1px 8px;
  border-radius: 5px;
  color: var(--app-color-primary);
  background: color-mix(in srgb, var(--app-color-primary) 10%, transparent)
}

.config-page__group-actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: auto
}

.config-page__card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 240px), 1fr));
  gap: 10px
}

/*
 * 模型数为 0 时把计数标成中性态，避免看上去像出错了。
 * 无模型渠道不再显示空状态块 —— 右上角的「从渠道导入」已经是拉取入口，
 * 再给一套「添加模型」按钮属于重复（用户 2026-10-16 明确要求只保留「从渠道导入」）。
 */
.config-page__group-count.is-empty {
  color: var(--app-text-color-secondary);
  opacity: .75;
}

.config-page__api-card {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border: 1px solid var(--app-border-color-light);
  border-radius: 8px;
  background: var(--app-bg-color);
  min-width: 0;
  transition: border-color .15s ease, box-shadow .15s ease
}

.config-page__api-card:hover {
  border-color: color-mix(in srgb, var(--app-color-primary) 45%, var(--app-border-color-light));
  box-shadow: 0 2px 8px #0000000f
}

.config-page__api-card.is-server {
  background: color-mix(in srgb, var(--app-color-primary) 3%, var(--app-bg-color))
}

.config-page__api-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0
}

.config-page__api-type {
  flex-shrink: 0;
  font-size: 10px;
  font-weight: 600;
  padding: 1px 8px;
  border-radius: 5px
}

.config-page__api-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap
}

.config-page__api-tag {
  font-size: 10px;
  padding: 1px 8px;
  border-radius: 5px;
  color: var(--app-color-primary);
  background: color-mix(in srgb, var(--app-color-primary) 10%, transparent)
}

.config-page__api-tag--success {
  color: #16a34a;
  background: #16a34a1a
}

.config-page__api-card .config-page__api-model {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap
}

.config-page__api-card-foot {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: auto;
  padding-top: 4px
}

.config-page__empty-action {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  margin-top: -16px
}

.config-page__empty-link {
  border: none;
  background: transparent;
  padding: 0;
  margin-top: 8px;
  font-size: 13px;
  color: var(--app-color-primary);
  cursor: pointer
}

.config-page__empty-link:hover {
  text-decoration: underline
}

.config-page__empty-link--ghost {
  color: var(--app-text-color-secondary)
}

.config-page__nav-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 17px;
  padding: 0 5px;
  margin-left: 2px;
  border-radius: 5px;
  font-size: 11px;
  font-weight: 600;
  color: var(--app-text-color-secondary);
  background: var(--app-fill-color);
  transition: all .15s ease
}

.config-page__nav-tab.is-active .config-page__nav-count {
  color: var(--app-color-primary);
  background: color-mix(in srgb, var(--app-color-primary) 12%, transparent)
}

.config-page__type-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 17px;
  height: 16px;
  padding: 0 5px;
  margin-left: 4px;
  border-radius: 5px;
  font-size: 10px;
  font-weight: 600;
  color: var(--app-text-color-secondary);
  background: var(--app-fill-color);
  transition: all .15s ease
}

.config-page__type-tab.is-active .config-page__type-count {
  color: var(--app-color-primary);
  background: color-mix(in srgb, var(--app-color-primary) 12%, transparent)
}

@media(max-width:720px) {
  .config-page__inner {
    padding: 12px
  }

  .config-page__card-grid {
    grid-template-columns: 1fr
  }
}

.config-page__mine-toolbar>.flex {
  min-width: 0;
  flex-wrap: wrap
}

@media(max-width:640px) {
  .config-page__mine-toolbar>.flex {
    width: 100%
  }

  .config-page__mine-search {
    flex: 1 1 140px;
    min-width: 0;
    width: auto
  }

  .config-page__type-tabs {
    flex-wrap: wrap
  }
}
</style>

<style>
/* ============ UI 搬迁补充 ============
   蓝本 CSS 里没有、但 B 项目兼容所需的少量类：
   - header-actions：B 项目的「导入 / 导出 / 创建自定义 API」按钮组
   - mine-sort：排序下拉（蓝本无排序功能）
   - providers：服务商管理视图容器（蓝本没有独立服务商页） */
.config-page__header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

.config-page__mine-sort {
  flex: none;
}

.config-page__providers {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding-right: 2px;
}

/* 远程模型弹窗底部（蓝本用工具类 flex items-center justify-between gap-3 等） */
.remote-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.remote-foot__hint {
  font-size: 12px;
  color: var(--cfg-ink-3);
}

.remote-foot__ops {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

/* 单模型探测：行内状态。
   转圈只用 CSS 画，不给每行挂 el-loading 遮罩 —— 几十行同时遮罩会闪。 */
.probe-cell {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  line-height: 1;
  white-space: nowrap;
}

.probe-idle {
  color: var(--cfg-ink-3);
}

.probe-running,
.probe-ok {
  color: var(--cfg-accent);
}

.probe-ok {
  font-variant-numeric: tabular-nums;
}

.probe-fail {
  color: var(--cfg-danger);
}

.probe-spin {
  display: inline-block;
  width: 11px;
  height: 11px;
  margin-right: 5px;
  border: 1.5px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: probe-spin 0.8s linear infinite;
}

@keyframes probe-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .probe-spin {
    animation: none;
    opacity: 0.55;
  }
}

/* 远程模型弹窗：搜索 + 类型筛选（对话 / 图片 / 视频 / 视觉 / 嵌入） */
.ml-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.ml-toolbar__search {
  flex: 1 1 200px;
  min-width: 160px;
}

.ml-types {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.ml-type-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 999px;
  border: 1px solid var(--cfg-line);
  background: var(--cfg-panel);
  color: var(--cfg-ink-2);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
  transition: 0.15s;
}

.ml-type-btn:hover {
  border-color: var(--cfg-accent-line);
  color: var(--cfg-accent-deep);
}

.ml-type-btn.on {
  border-color: var(--cfg-accent);
  background: var(--cfg-accent-bg);
  color: var(--cfg-accent-deep);
  font-weight: 600;
}

.ml-type-btn b {
  font-weight: 600;
  opacity: 0.72;
}

.ml-type-select {
  width: 100%;
}

/* ============ 高度链路修复 ============
   蓝本在这些容器上还叠了 Tailwind 工具类（flex / h-full / flex-1 / flex-col /
   min-w-0 / overflow-y-auto），B 项目没有工具类层。若不把它们固化进类本身，
   高度链路会断裂，flex:1 1 0 的列表在自动高度容器里塌成 0 高 ——
   表现就是「有数据但界面什么都不显示」。 */
.config-page {
  display: flex;
  height: 100%;
}

.config-page__content {
  flex: 1 1 0;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.config-page__scroll {
  flex: 1 1 0;
  min-height: 0;
  overflow-y: auto;
}

/* 蓝本此处是 mb-3 */
.config-page__mine-toolbar {
  margin-bottom: 12px;
}

/* B 项目的 .content-frame 已有 28px 内边距，这里去掉蓝本 __inner 的 16px 避免双重留白 */
.config-page__inner {
  padding: 0 0 16px;
}

/* 蓝本此处是工具类 "flex items-center gap-2"（搜索 + 排序紧挨着，不撑开） */
.config-page__mine-ops {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>

<style>
/* ============ 跨组件通用类（提升为全局） ============
   这些类原先定义在 Config.vue 的 <style scoped> 里，但被 ApiPlaza / ChannelDialog /
   CapDialog 使用。Vue 的 scoped 会给选择器追加 [data-v-xxx]，只对 Config.vue 模板内的
   元素生效 —— 其他组件里的这些元素因此**完全没有样式**，表现就是「按钮是毛坯」。
   故提升为全局：命名均带 cfg- / cap- 前缀或为局部语义，污染风险可忽略。 */

/* ================= 统一按钮 ================= */
.cfg-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 9px 16px;
  border-radius: 10px;
  border: 1px solid var(--cfg-line);
  background: var(--cfg-panel);
  color: var(--cfg-ink-2);
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  cursor: pointer;
  transition: 0.15s;
  white-space: nowrap;
}

.cfg-btn svg {
  width: 14px;
  height: 14px;
}

.cfg-btn:hover {
  border-color: var(--cfg-accent);
  color: var(--cfg-accent);
}

.cfg-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.cfg-btn.primary {
  background: var(--cfg-accent);
  border-color: var(--cfg-accent);
  color: #fff;
}

.cfg-btn.primary:hover {
  background: var(--cfg-accent-hover);
  color: #fff;
}

.cfg-btn.import {
  color: var(--cfg-accent-deep);
  background: var(--cfg-accent-bg);
  border-color: var(--cfg-accent-line);
}

.cfg-btn.import:hover {
  color: var(--cfg-accent);
}

.cfg-btn.export {
  color: var(--cfg-warn-ink);
  background: var(--cfg-warn-bg);
  border-color: var(--cfg-warn-ink);
  opacity: 0.9;
}

.cfg-btn.export:hover {
  opacity: 1;
  color: var(--cfg-warn-ink);
}

.cfg-btn.sm {
  padding: 5px 12px;
  font-size: 12px;
  border-radius: 8px;
  font-weight: 500;
}

.cfg-btn.danger {
  color: var(--cfg-danger);
}

.cfg-btn.danger:hover {
  border-color: var(--cfg-danger);
  background: var(--cfg-danger-bg);
  color: var(--cfg-danger);
}

/*
 * 渠道级「删除渠道」：破坏性操作，需要和旁边的常规按钮拉开区分度。
 * 旧样式是灰底 + 红字（跟「从渠道导入」长得几乎一样，容易误点），
 * 改为描边危险按钮：浅红底 + 红边 + 红字，hover 加深。
 *
 * ⚠ 文字色不直接用 --cfg-danger（#dc2626）：它压在 --cfg-danger-bg（#fef2f2）
 *   上只有 4.41:1，hover 时掉到 3.95，都不达 4.5。这里用加深一档的 #b91c1c
 *   （常态 5.91 / hover 5.30）。**不能改 --cfg-danger 本身**：弹窗里的删除按钮
 *   与 --app-color-danger 都在用同一个变量，改全局会连带影响别处。
 *   暗色下 --cfg-danger 是 #f87171 压在深红底上已有 5.89，继续走变量即可。
 */
.cfg-btn.danger-outline {
  background: var(--cfg-danger-bg);
  border-color: var(--cfg-danger-line);
  color: #b91c1c;
}

.cfg-btn.danger-outline:hover {
  background: var(--cfg-danger-hover);
  border-color: var(--cfg-danger);
  color: #b91c1c;
}

:root[data-theme='dark'] .cfg-btn.danger-outline,
:root[data-theme='dark'] .cfg-btn.danger-outline:hover {
  color: var(--cfg-danger);
}

.cat-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}

.cap-switch-row .t {
  font-size: 13px;
  font-weight: 600;
  color: var(--cfg-ink);
}

.cap-switch-row .d {
  font-size: 11px;
  color: var(--cfg-ink-3);
  margin-top: 2px;
}
</style>
