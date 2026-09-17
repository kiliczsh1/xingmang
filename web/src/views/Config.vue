<template>
  <div class="config-container">
    <!-- 顶部标题 + 操作 -->
    <div class="header">
      <div class="header-content">
        <h2>API 配置管理</h2>
        <p class="header-subtitle">统一管理服务商、模型分类与导入导出</p>
      </div>
      <div class="header-actions">
        <el-dropdown trigger="click" @command="handleImportCommand">
          <button class="cfg-btn import">
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
          <button class="cfg-btn export">
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
        <button class="cfg-btn primary" @click="handleCreateNewModel">
          <el-icon><Plus /></el-icon>新建模型
        </button>
      </div>
    </div>

    <!-- 状态栏 -->
    <div class="statusbar">
      <span class="sb-item">服务商 <b>{{ providers.length }}</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">模型 <b>{{ models.length }}</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">已启用 <b class="green">{{ enabledCount }}</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">默认模型 <b>{{ defaultModelName }}</b></span>
      <span class="sb-sep"></span>
      <span class="sb-item">分类 <b>{{ categories.length }}</b></span>
      <span class="pill" :class="{ dirty }">{{ dirty ? '● 有未保存修改' : '✓ 全部已保存' }}</span>
    </div>

    <!-- 视图切换 -->
    <div class="view-switch">
      <button class="vs-btn" :class="{ on: activeView === 'models' }" @click="activeView = 'models'">
        <el-icon><Collection /></el-icon>模型工作台
      </button>
      <button class="vs-btn" :class="{ on: activeView === 'providers' }" @click="activeView = 'providers'">
        <el-icon><OfficeBuilding /></el-icon>服务商管理
      </button>
    </div>

    <!-- 模型工作台：左分类库 + 右编辑器 -->
    <div v-show="activeView === 'models'" class="ws">
      <div class="pane left">
        <div class="pane-h">
          <span class="ic"><el-icon><Collection /></el-icon></span>
          模型分类库
          <span class="x">选择分类 → 选择模型</span>
        </div>
        <div class="cats">
          <span class="ctab" :class="{ on: curCat === '' }" @click="curCat = ''">全部<span class="cnt">{{ models.length }}</span></span>
          <span
            v-for="c in categories"
            :key="c.key"
            class="ctab"
            :class="{ on: curCat === c.key }"
            @click="curCat = c.key"
          >{{ c.name }}<span class="cnt">{{ modelsByCat(c.key).length }}</span></span>
        </div>
        <div class="mbody">
          <div
            v-for="m in modelsInCurCat"
            :key="m.id"
            class="mrow"
            :class="{ on: selectedId === m.id }"
            @click="selectModel(m)"
          >
            <span class="st" :style="{ background: catColor(m) }"></span>
            <div class="mrow-main">
              <div class="nm">{{ m.name }}<span v-if="m.is_default === 1" class="star">★</span></div>
              <div class="pv">{{ m.provider_name || '未关联服务商' }} · {{ m.model }}</div>
            </div>
            <span class="en" :class="{ off: m.enabled !== 1 }">{{ m.enabled === 1 ? '启用' : '停用' }}</span>
          </div>
          <div v-if="modelsInCurCat.length === 0" class="empty">
            该分类暂无模型<br /><span class="empty-sub">点击右上角「新建模型」添加</span>
          </div>
        </div>
      </div>

      <div class="pane editor-pane">
        <div class="pane-h">
          <span class="ic"><el-icon><EditPen /></el-icon></span>
          编辑模型
          <span class="x">{{ isNewModel ? '新建配置' : editingModel?.id ? '实时编辑' : '' }}</span>
        </div>
        <div v-if="editingModel" class="editor">
          <!-- 模型头部 -->
          <div class="e-top">
            <span class="e-ic" :style="{ background: catColor(editingModel) }">
              <el-icon :size="22"><component :is="catIcon(editingModel)" /></el-icon>
            </span>
            <div>
              <div class="e-nm">{{ editingModel.name || '未命名模型' }}<span v-if="editingModel.is_default === 1" class="def">默认</span></div>
              <div class="e-pv">{{ providerNameOf(editingModel.provider_id) }} · {{ editingModel.model || '—' }} · {{ fmtTok(editingModel.max_tokens) }} 上下文</div>
              <div class="e-badges">
                <span class="b">{{ catLabel(editingModel) }}</span>
                <span class="b" :class="editingModel.enabled === 1 ? 'b-on' : 'b-off'">{{ editingModel.enabled === 1 ? '已启用' : '已停用' }}</span>
              </div>
            </div>
          </div>
          <!-- 表单 -->
          <div class="form">
            <div class="fg-row">
              <div class="fg">
                <label>模型名称</label>
                <el-input v-model="editingModel.name" placeholder="如 GPT-4o" @input="dirty = true" />
              </div>
              <div class="fg">
                <label>模型标识</label>
                <el-input v-model="editingModel.model" placeholder="如 gpt-4o" @input="dirty = true" />
              </div>
            </div>
            <div class="fg-row">
              <div class="fg">
                <label>所属服务商</label>
                <el-select v-model="editingModel.provider_id" placeholder="选择服务商" popper-class="cfg-select-popper" style="width: 100%" @change="onProviderChange">
                  <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
                </el-select>
              </div>
              <div class="fg">
                <label>模型分类</label>
                <el-select v-model="editingModel.category" placeholder="选择分类" popper-class="cfg-select-popper" style="width: 100%" @change="dirty = true">
                  <el-option label="未分类" value="" />
                  <el-option v-for="c in categories" :key="c.key" :label="c.name" :value="c.key">
                    <span style="display:inline-flex;align-items:center;gap:8px">
                      <span class="cat-dot" :style="{ background: c.color }"></span>{{ c.name }}
                    </span>
                  </el-option>
                </el-select>
              </div>
            </div>

            <div class="fg discover-row">
              <label>获取模型列表</label>
              <div class="discover-inner">
                <el-button :loading="discoveringModels" :disabled="!editingModel.provider_id" @click="handleDiscoverModels">
                  <el-icon><Search /></el-icon>获取模型列表
                </el-button>
                <span v-if="discoveredModels.length > 0" class="d-hint link" @click="modelListDialogVisible = true">
                  已获取 {{ discoveredModels.length }} 个模型 · 点击查看
                </span>
              </div>
            </div>

            <div class="fg slider">
              <div class="slider-row"><label>Temperature</label><span class="v">{{ editingModel.temperature }}</span></div>
              <el-slider v-model="editingModel.temperature" :min="0" :max="2" :step="0.1" @change="dirty = true" />
            </div>
            <div class="fg slider">
              <div class="slider-row"><label>Top P</label><span class="v">{{ editingModel.top_p }}</span></div>
              <el-slider v-model="editingModel.top_p" :min="0" :max="1" :step="0.05" @change="dirty = true" />
            </div>
            <div class="fg slider">
              <div class="slider-row"><label>频率惩罚</label><span class="v">{{ editingModel.frequency_penalty }}</span></div>
              <el-slider v-model="editingModel.frequency_penalty" :min="-2" :max="2" :step="0.1" @change="dirty = true" />
            </div>
            <div class="fg slider">
              <div class="slider-row"><label>最大 Token</label><span class="v">{{ editingModel.max_tokens }}</span></div>
              <el-slider v-model="editingModel.max_tokens" :min="100" :max="maxTokensLimit" :step="100" @change="dirty = true" />
            </div>

            <div class="toggle-grid">
              <div class="tcard">
                <div><div class="t">启用模型</div><div class="d">保存后全局可用</div></div>
                <el-switch v-model="enabledModel" />
              </div>
              <div class="tcard">
                <div><div class="t">设为默认</div><div class="d">作为系统兜底模型</div></div>
                <el-switch v-model="defaultModel" />
              </div>
            </div>

            <div class="fg">
              <label>模型描述</label>
              <el-input v-model="editingModel.description" type="textarea" :rows="2" placeholder="模型用途与能力简介" @input="dirty = true" />
            </div>

            <div class="e-ops">
              <button class="big save" @click="saveEditingModel"><el-icon><Check /></el-icon>保存修改</button>
              <button v-if="!isNewModel" class="big test" :disabled="testing" @click="handleTestModel(editingModel)"><el-icon><Connection /></el-icon>测试连接</button>
              <button v-if="!isNewModel" class="big del" @click="handleDeleteEditingModel"><el-icon><Delete /></el-icon>删除</button>
              <button v-if="isNewModel" class="big cancel" @click="cancelNewModel">取消</button>
            </div>
          </div>
        </div>
        <div v-else class="e-empty">
          <div class="e-empty-ic"><el-icon><EditPen /></el-icon></div>
          <p>在左侧选择模型开始编辑</p>
          <p class="empty-sub">或点击右上角「新建模型」配置新模型</p>
        </div>
      </div>
    </div>

    <!-- 服务商管理 -->
    <div v-show="activeView === 'providers'" class="providers-view">
      <div class="pv-head">
        <div class="pv-title">
          <h3>服务商列表</h3>
          <span class="pv-sub">{{ providers.length }} 家服务商 · 共 {{ models.length }} 个模型</span>
        </div>
        <button class="cfg-btn primary" @click="handleCreateProvider">
          <el-icon><Plus /></el-icon>添加服务商
        </button>
      </div>
      <el-table :data="providers" stripe style="width: 100%">
        <el-table-column label="服务商" min-width="200">
          <template #default="{ row }">
            <div class="pv-prov">
              <span class="pv-lg" :style="{ background: providerColor(row.provider_type) }">{{ row.name.slice(0, 1) }}</span>
              <div class="pv-meta">
                <div class="pv-nm">{{ row.name }}</div>
                <div class="pv-tp">{{ getProviderTypeLabel(row.provider_type) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="api_url" label="API 地址" min-width="240" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="pv-url">{{ row.api_url }}</span>
          </template>
        </el-table-column>
        <el-table-column label="模型" width="80" align="center">
          <template #default="{ row }">
            <span class="pv-cnt">{{ getProviderModelCount(row.id) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="right">
          <template #default="{ row }">
            <button class="cfg-btn sm" @click="handleEditProvider(row)">编辑</button>
            <button class="cfg-btn sm danger" @click="handleDeleteProvider(row)">删除</button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 服务商弹窗 -->
    <el-dialog
      v-model="providerDialogVisible"
      :title="isEditProvider ? '编辑服务商' : '添加服务商'"
      width="620px"
      append-to-body
      class="cfg-dialog provider-add-dialog"
    >
      <div class="pseg">
        <button :class="{ on: providerAddMode === 'preset' }" @click="providerAddMode = 'preset'">模型服务商</button>
        <button :class="{ on: providerAddMode === 'custom' }" @click="providerAddMode = 'custom'">自定义配置</button>
      </div>

      <div v-show="providerAddMode === 'preset'" class="model-add-content">
        <div class="pcard-grid">
          <div
            v-for="c in PROVIDER_CARDS"
            :key="c.key"
            class="pcard"
            :class="{ on: providerFormData.provider_type === c.key }"
            @click="onCardPick(c.key)"
          >
            <span class="pcard-av">{{ c.initial }}</span>
            <span class="pcard-lb">{{ c.label }}</span>
          </div>
        </div>
        <transition name="reveal">
          <div v-show="providerFormData.provider_type" class="pcard-reveal">
            <div class="pcard-chosen">已选择 <b>{{ presetName }}</b> · 默认端点 <b>{{ presetUrl }}</b></div>
            <el-input v-model="providerFormData.name" placeholder="备注名称（已自动带出，可修改）" />
            <el-input v-model="providerFormData.api_key" type="password" placeholder="API 密钥（必填）" show-password style="margin-top: 10px" />
          </div>
        </transition>
      </div>

      <div v-show="providerAddMode === 'custom'" class="model-add-content">
        <el-form :model="providerFormData" label-width="100px">
          <el-alert type="info" :closable="false" show-icon class="custom-config-hint">
            <template #default>请填写兼容 OpenAI API 的服务端点地址，不要以斜杠结尾。/chat/completions 将会被补充到你填写的地址末尾。</template>
          </el-alert>
          <el-form-item label="API 格式" required>
            <el-select v-model="providerFormData.api_format" placeholder="请选择 API 格式" popper-class="cfg-select-popper" style="width: 100%">
              <el-option label="OpenAI Chat Completions 格式" value="openai" />
              <el-option label="Claude Messages 格式" value="claude" />
              <el-option label="Gemini 格式" value="gemini" />
            </el-select>
          </el-form-item>
          <el-form-item label="完整 URL">
            <div class="url-mode-row">
              <el-switch v-model="providerFormData.use_full_url" />
              <span class="form-item-hint">开启后直接使用填写的地址</span>
            </div>
          </el-form-item>
          <el-form-item label="请求地址" required>
            <el-input v-model="providerFormData.api_url" placeholder="e.g. https://api.openai.com/v1" />
          </el-form-item>
          <el-form-item label="服务商名称" required>
            <el-input v-model="providerFormData.name" placeholder="请输入服务商名称" />
          </el-form-item>
          <el-form-item label="API 密钥" required>
            <el-input v-model="providerFormData.api_key" type="password" placeholder="输入 API 密钥" show-password />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <button class="cfg-btn" @click="providerDialogVisible = false">取消</button>
        <button class="cfg-btn primary" @click="handleSubmitProvider">确定</button>
      </template>
    </el-dialog>

    <!-- 导入弹窗 -->
    <el-dialog v-model="importDialogVisible" title="导入API配置" width="600px" :close-on-click-modal="false" append-to-body class="cfg-dialog">
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

    <!-- 获取到的模型列表弹窗 -->
    <el-dialog v-model="modelListDialogVisible" title="获取到的模型" width="640px" append-to-body class="cfg-dialog">
      <div class="model-list-dialog">
        <div class="ml-bar">
          <el-input v-model="mlSearch" placeholder="搜索模型 ID / 名称" clearable>
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
          <span class="ml-count">{{ filteredModels.length }} / {{ discoveredModels.length }}</span>
        </div>
        <div class="ml-tags">
          <span class="ml-tag" :class="{ on: mlTag === '' }" @click="mlTag = ''">全部</span>
          <span
            v-for="t in MODEL_TAGS"
            :key="t.key"
            class="ml-tag"
            :class="{ on: mlTag === t.key }"
            @click="mlTag = mlTag === t.key ? '' : t.key"
          >{{ t.label }}</span>
        </div>
        <div class="ml-body">
          <template v-for="g in groupedModels" :key="g.name">
            <div class="ml-grp-h">{{ g.name }}<span class="cnt">{{ g.items.length }}</span></div>
            <div
              v-for="m in g.items"
              :key="m.id"
              class="ml-item"
              :class="{ on: m.id === selectedModelId }"
              @click="selectedModelId = m.id"
            >
              <span class="ml-radio" :class="{ on: m.id === selectedModelId }"></span>
              <div class="ml-info">
                <div class="ml-id">{{ m.id }}</div>
                <div class="ml-meta">{{ m.name !== m.id ? m.name : '' }}{{ m.max_tokens ? ' · ' + fmtTok(m.max_tokens) + ' tokens' : '' }}</div>
              </div>
            </div>
          </template>
          <div v-if="filteredModels.length === 0" class="ml-empty">没有匹配的模型</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="modelListDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedModelId" @click="applyDiscoveredModel">
          添加到当前模型
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, type Component } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Upload, Download, ArrowDown, Document, UploadFilled, Search,
  Plus, Collection, EditPen, OfficeBuilding, Check, Connection, Delete,
  ChatDotRound, MagicStick, PictureFilled, Cpu
} from '@element-plus/icons-vue'
import { providerAPI, configAPI } from '@/api'
import type { ApiProvider, ApiModel } from '@/types'

/* ---------------- 模型分类定义 ---------------- */
const categories = [
  { key: 'chat', name: '通用对话', color: '#16a34a', icon: ChatDotRound },
  { key: 'reason', name: '深度推理', color: '#7c3aed', icon: MagicStick },
  { key: 'multimodal', name: '多模态', color: '#db2777', icon: PictureFilled },
  { key: 'code', name: '代码生成', color: '#2563eb', icon: Cpu },
  { key: 'long', name: '长文本', color: '#d97706', icon: Document },
  { key: 'embed', name: '嵌入向量', color: '#0891b2', icon: Connection }
]

/** 根据模型标识猜测分类（老数据没有分类时兜底） */
/**
 * 从模型 ID 自动推断分类（融合 Cherry Studio 的 ID 派生思路）：
 * 1. 规范化 ID：剥掉 / 前缀（openrouter 的 deepseek/deepseek-r1）与 :tag 后缀（ollama 的 qwen2.5:7b）
 * 2. 正则精确匹配能力关键词
 * 3. 家族名兜底：取第一个 '-' 前的家族名，映射到能力分类（Cherry 的 deriveModelGroupName 思路）
 */
const guessCategory = (modelId: string): string => {
  const raw = (modelId || '').toLowerCase()
  const base = (raw.split('/').pop() || raw).split(':')[0]

  if (/(reasoner|r1|thinking|o1-|o3-|o4-)/.test(base)) return 'reason'
  if (/(vl|vision|visual|image|gemini|llava|omni)/.test(base)) return 'multimodal'
  if (/(coder|codegeex|starcoder|code-)/.test(base)) return 'code'
  if (/(bge|embed|rerank|text-embedding)/.test(base)) return 'embed'
  if (/(longcontext|long-context|1\.5-pro|max-longcontext)/.test(base)) return 'long'

  // Cherry 式家族派生：取第一个 '-' 前的家族名，能力型家族映射到对应分类
  const family = base.split('-')[0]?.trim()
  if (family && family !== base) {
    const familyMap: Record<string, string> = {
      code: 'code', coder: 'code', codegeex: 'code', starcoder: 'code',
      reason: 'reason', reasoner: 'reason', thinking: 'reason',
      vl: 'multimodal', vision: 'multimodal', visual: 'multimodal', llava: 'multimodal', omni: 'multimodal',
      bge: 'embed', embed: 'embed', embedding: 'embed', rerank: 'embed',
      long: 'long', longcontext: 'long'
    }
    if (familyMap[family]) return familyMap[family]
  }
  return 'chat'
}

const categoryOf = (m: ApiModel): string => m.category || guessCategory(m.model || '')
const catDef = (key: string) => categories.find(c => c.key === key)
const catColor = (m: ApiModel): string => catDef(categoryOf(m))?.color || '#64748b'
const catLabel = (m: ApiModel): string => catDef(categoryOf(m))?.name || '未分类'
const catIcon = (m: ApiModel): Component => catDef(categoryOf(m))?.icon || Document
const modelsByCat = (key: string) => models.value.filter(m => categoryOf(m) === key)

/* ---------------- 基础状态 ---------------- */
const activeView = ref<'models' | 'providers'>('models')
const curCat = ref('')
const providers = ref<ApiProvider[]>([])
const models = ref<ApiModel[]>([])

/* ---------------- 编辑器状态 ---------------- */
const editingModel = ref<ApiModel | null>(null)
const selectedId = ref<number | null>(null)
const isNewModel = ref(false)
const dirty = ref(false)
const testing = ref(false)
const discoveredModels = ref<Array<{ id: string; name: string; max_tokens?: number | null }>>([])
const discoveringModels = ref(false)

/* ---------------- 状态栏计算 ---------------- */
const enabledCount = computed(() => models.value.filter(m => m.enabled === 1).length)
const defaultModelName = computed(() => models.value.find(m => m.is_default === 1)?.name || '未设置')
const modelsInCurCat = computed(() => (curCat.value === '' ? models.value : modelsByCat(curCat.value)))
const maxTokensLimit = computed(() => {
  const m = editingModel.value
  if (!m) return 128000
  const guess = m.model ? resolveMaxTokens(m.model) : 128000
  return Math.max(guess, m.max_tokens || 0, 128000)
})

const enabledModel = computed({
  get: () => editingModel.value?.enabled === 1,
  set: (v: boolean) => {
    if (editingModel.value) {
      editingModel.value.enabled = v ? 1 : 0
      dirty.value = true
    }
  }
})
const defaultModel = computed({
  get: () => editingModel.value?.is_default === 1,
  set: (v: boolean) => {
    if (editingModel.value) {
      editingModel.value.is_default = v ? 1 : 0
      dirty.value = true
    }
  }
})

/* ---------------- 生命周期 ---------------- */
onMounted(async () => {
  await fetchProviders()
  await initializeDefaultProviders()
  await fetchModels()
})

/* ---------------- 服务商逻辑 ---------------- */
const fetchProviders = async () => {
  const res = await providerAPI.getAll()
  if (res.success && res.data) {
    providers.value = res.data
  }
}

const initializeDefaultProviders = async () => {
  if (providers.value.length > 0) return
  try {
    const defaultProviders = [
      { name: 'DeepSeek', provider_type: 'deepseek', api_key: '', api_url: 'https://api.deepseek.com/v1/chat/completions' },
      { name: 'SiliconFlow', provider_type: 'other', api_key: '', api_url: 'https://api.siliconflow.cn/v1/chat/completions' },
      { name: 'bltcy.ai', provider_type: 'other', api_key: '', api_url: 'https://api.bltcy.ai/v1/chat/completions' },
      { name: 'ModelScope', provider_type: 'other', api_key: '', api_url: 'https://api-inference.modelscope.cn/v1/chat/completions' }
    ]
    for (const provider of defaultProviders) {
      const res = await providerAPI.create(provider)
      if (res.success && res.data) {
        ElMessage.success(`已自动添加 ${provider.name} 服务商，请配置 API 密钥后使用`)
      }
    }
    await fetchProviders()
  } catch (error) {
    console.error('初始化默认服务商失败:', error)
  }
}

const getProviderTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    openai: 'OpenAI', azure: 'Azure OpenAI', claude: 'Claude',
    qwen: '通义千问', deepseek: 'DeepSeek', gemini: 'Gemini', ollama: 'Ollama', other: '其他'
  }
  return map[type] || type
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

const handleCreateProvider = () => {
  isEditProvider.value = false
  providerAddMode.value = 'preset'
  providerFormData.value = {
    id: 0, name: '', provider_type: 'openai', api_key: '', api_url: '',
    url_suffix_mode: 'compat', api_format: 'openai', use_full_url: false
  }
  onPresetProviderChange('openai') // 默认选中 OpenAI 卡片，自动带出名称与端点
  providerDialogVisible.value = true
}

const handleEditProvider = (provider: ApiProvider) => {
  isEditProvider.value = true
  providerAddMode.value = 'preset'
  providerFormData.value = {
    id: provider.id, name: provider.name, provider_type: provider.provider_type,
    api_key: provider.api_key, api_url: provider.api_url,
    url_suffix_mode: provider.url_suffix_mode || 'compat',
    api_format: 'openai', use_full_url: provider.use_full_url || false
  }
  providerDialogVisible.value = true
}

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

const normalizeApiUrl = (url: string, useFullUrl: boolean) => {
  const trimmed = url.trim().replace(/\/+$/, '')
  if (!trimmed) return ''
  if (useFullUrl) return trimmed
  if (trimmed.endsWith('/chat/completions')) return trimmed
  if (!trimmed.includes('/v1')) {
    return `${trimmed}/v1/chat/completions`
  }
  return `${trimmed}/chat/completions`
}

const PRESET_PROVIDER_DEFAULTS: Record<string, { name: string; api_url: string }> = {
  openai: { name: 'OpenAI', api_url: 'https://api.openai.com/v1' },
  azure: { name: 'Azure OpenAI', api_url: 'https://YOUR_RESOURCE.openai.azure.com' },
  claude: { name: 'Claude', api_url: 'https://api.anthropic.com/v1' },
  qwen: { name: '通义千问', api_url: 'https://dashscope.aliyuncs.com/compatible-mode/v1' },
  deepseek: { name: 'DeepSeek', api_url: 'https://api.deepseek.com/v1' },
  gemini: { name: 'Gemini', api_url: 'https://generativelanguage.googleapis.com/v1beta' },
  ollama: { name: 'Ollama', api_url: 'http://localhost:11434' },
  siliconflow: { name: '硅基流动 SiliconFlow', api_url: 'https://api.siliconflow.cn/v1' },
  modelscope: { name: 'ModelScope', api_url: 'https://api-inference.modelscope.cn/v1' },
  other: { name: '', api_url: '' }
}

/** 服务商卡片（卡片直选风格，与上面 PREDET 列表一一对应） */
const PROVIDER_CARDS: Array<{ key: string; label: string; initial: string }> = [
  { key: 'openai', label: 'OpenAI', initial: 'O' },
  { key: 'azure', label: 'Azure OpenAI', initial: 'A' },
  { key: 'claude', label: 'Claude', initial: 'C' },
  { key: 'qwen', label: '通义千问', initial: '千' },
  { key: 'deepseek', label: 'DeepSeek', initial: 'D' },
  { key: 'gemini', label: 'Gemini', initial: 'G' },
  { key: 'ollama', label: 'Ollama（本地）', initial: 'Ol' },
  { key: 'siliconflow', label: '硅基流动 SiliconFlow', initial: 'Si' },
  { key: 'modelscope', label: 'ModelScope', initial: 'Mo' },
  { key: 'other', label: '其他', initial: '+' }
]

/** 已选卡片对应的预置名称 / 默认端点（用于选中提示条） */
const presetName = computed(() => PRESET_PROVIDER_DEFAULTS[providerFormData.value.provider_type]?.name || providerFormData.value.name || '—')
const presetUrl = computed(() => providerFormData.value.api_url || PRESET_PROVIDER_DEFAULTS[providerFormData.value.provider_type]?.api_url || '（自定义地址）')

const onPresetProviderChange = (value: string) => {
  const preset = PRESET_PROVIDER_DEFAULTS[value]
  if (!preset) return
  if (!providerFormData.value.name || Object.values(PRESET_PROVIDER_DEFAULTS).some(p => p.name === providerFormData.value.name)) {
    providerFormData.value.name = preset.name
  }
  if (!providerFormData.value.api_url) {
    providerFormData.value.api_url = preset.api_url
  }
}

/** 卡片直选：点卡片即选中服务商，并自动带出预置名称 / 地址（编辑时保留已填内容） */
const onCardPick = (key: string) => {
  providerFormData.value.provider_type = key
  onPresetProviderChange(key)
}

const handleSubmitProvider = async () => {
  if (!providerFormData.value.provider_type) {
    ElMessage.warning('请选择服务商')
    return
  }
  if (!providerFormData.value.api_key) {
    ElMessage.warning('请填写 API 密钥')
    return
  }
  if (!providerFormData.value.name) {
    const preset = PRESET_PROVIDER_DEFAULTS[providerFormData.value.provider_type]
    if (preset) providerFormData.value.name = preset.name
  }
  if (providerAddMode.value === 'custom' && !providerFormData.value.api_url) {
    ElMessage.warning('请填写请求地址')
    return
  }
  providerFormData.value.api_url = normalizeApiUrl(providerFormData.value.api_url, providerFormData.value.use_full_url)
  try {
    if (isEditProvider.value) {
      const res = await providerAPI.update(providerFormData.value.id, providerFormData.value)
      if (res.success) ElMessage.success('更新成功')
    } else {
      const res = await providerAPI.create(providerFormData.value)
      if (res.success) ElMessage.success('创建成功')
    }
    providerDialogVisible.value = false
    await fetchProviders()
    await fetchModels()
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

const fmtTok = (n: number) => (n >= 1000 ? `${(n / 1000).toFixed(0)}K` : `${n}`)

const selectModel = (m: ApiModel) => {
  selectedId.value = m.id
  isNewModel.value = false
  dirty.value = false
  discoveredModels.value = []
  selectedModelId.value = ''
  editingModel.value = {
    ...m,
    category: m.category || guessCategory(m.model),
    temperature: m.temperature ?? 0.7,
    top_p: m.top_p ?? 0.9,
    frequency_penalty: m.frequency_penalty ?? 0,
    max_tokens: m.max_tokens ?? 2000,
    enabled: m.enabled ?? 1,
    is_default: m.is_default ?? 0
  }
}

const handleCreateNewModel = () => {
  if (providers.value.length === 0) {
    ElMessage.warning('请先添加服务商')
    return
  }
  activeView.value = 'models' // 无论当前在哪个区块，点击新建模型都跳转到模型工作台
  isNewModel.value = true
  selectedId.value = null
  dirty.value = false
  discoveredModels.value = []
  selectedModelId.value = ''
  editingModel.value = {
    id: 0,
    provider_id: providers.value[0].id,
    name: '',
    model: '',
    temperature: 0.7,
    max_tokens: 2000,
    top_p: 0.9,
    frequency_penalty: 0,
    description: '',
    is_default: 0,
    enabled: 1,
    category: '',
    sort_order: models.value.length,
    provider_name: providers.value[0].name,
    created_at: '',
    updated_at: ''
  }
}

const cancelNewModel = () => {
  editingModel.value = null
  isNewModel.value = false
  dirty.value = false
}

const onProviderChange = () => {
  dirty.value = true
  discoveredModels.value = []
  selectedModelId.value = ''
  if (editingModel.value) {
    editingModel.value.provider_name = providerNameOf(editingModel.value.provider_id)
  }
}

const saveEditingModel = async () => {
  const m = editingModel.value
  if (!m) return
  if (!m.provider_id) { ElMessage.warning('请选择服务商'); return }
  if (!m.name.trim()) { ElMessage.warning('请填写模型名称'); return }
  if (!m.model.trim()) { ElMessage.warning('请填写模型标识'); return }
  if (!m.category) m.category = guessCategory(m.model)

  const payload = {
    ...m,
    is_default: m.is_default ? 1 : 0,
    enabled: m.enabled ? 1 : 0
  }
  try {
    if (isNewModel.value || !m.id) {
      const res = await configAPI.create(payload)
      if (res.success) ElMessage.success('创建成功')
    } else {
      const res = await configAPI.update(m.id, payload)
      if (res.success) ElMessage.success('保存成功')
    }
    dirty.value = false
    await fetchModels()
    const saved =
      models.value.find(x => x.name === m.name && x.model === m.model && x.provider_id === m.provider_id) ||
      models.value.find(x => x.id === m.id)
    if (saved) selectModel(saved)
    else cancelNewModel()
  } catch (error) {
  }
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
      await fetchModels()
    }
  } catch (error) {
  }
}

const handleTestModel = async (m: ApiModel) => {
  if (!m.id || testing.value) return
  testing.value = true
  try {
    const res = await configAPI.testModel(m.id)
    const d = (res as any).data || {}
    if (res.success && d.success) {
      await ElMessageBox.alert(
        `服务商：${d.providerName || providerNameOf(m.provider_id)}\n耗时：${d.latencyMs} ms\n返回：${(d.preview || '').slice(0, 200)}`,
        `模型测试成功 - ${d.modelName || m.name}`,
        { confirmButtonText: '知道了' }
      )
    } else {
      ElMessage.warning(`${m.name}：${d.message || '连接失败'}`)
    }
  } catch (error) {
    ElMessage.error('测试出错')
  } finally {
    testing.value = false
  }
}

/* ---------------- 发现模型 ---------------- */
const KNOWN_MODEL_MAX_TOKENS: Record<string, number> = {
  'gpt-4': 8192, 'gpt-4-turbo': 128000, 'gpt-4o': 128000, 'gpt-4o-mini': 128000,
  'gpt-4-32k': 32768, 'gpt-3.5-turbo': 4096, 'gpt-3.5-turbo-16k': 16384,
  'claude-3-opus': 200000, 'claude-3-sonnet': 200000, 'claude-3-haiku': 200000,
  'claude-3.5-sonnet': 200000, 'claude-3.5-haiku': 200000,
  'deepseek-chat': 128000, 'deepseek-reasoner': 128000, 'deepseek-r1': 128000,
  'qwen-turbo': 131072, 'qwen-plus': 131072, 'qwen-max': 32768, 'qwen-max-longcontext': 1048576,
  'gemini-1.5-pro': 2097152, 'gemini-1.5-flash': 1048576, 'gemini-2.0-flash': 1048576,
  'gemini-2.5-pro': 1048576, 'gemini-2.5-flash': 1048576, 'gemini-3-pro-preview': 1048576, 'gemini-3-flash-preview': 1048576,
  'kimi-k2.5': 131072, 'glm-4': 131072, 'glm-5': 131072,
  'llama-3-70b': 8192, 'llama-3-8b': 8192, 'mixtral-8x7b': 32768,
  'mistral-large': 131072, 'mistral-small': 32768,
  'command-r': 131072, 'command-r-plus': 131072
}

const resolveMaxTokens = (modelId: string, apiMaxTokens?: number | null) => {
  if (apiMaxTokens != null) return apiMaxTokens
  const lower = modelId.toLowerCase()
  for (const [key, value] of Object.entries(KNOWN_MODEL_MAX_TOKENS)) {
    if (lower.includes(key)) return value
  }
  return 8192
}

/* 获取到的模型列表弹窗 */
const modelListDialogVisible = ref(false)
const mlSearch = ref('')
const selectedModelId = ref('')

/** 能力标签（Cherry 式：全部/推理/视觉/联网/免费/嵌入） */
const MODEL_TAGS = [
  { key: 'reason', label: '推理' },
  { key: 'vision', label: '视觉' },
  { key: 'search', label: '联网' },
  { key: 'free', label: '免费' },
  { key: 'embed', label: '嵌入' }
] as const
const mlTag = ref('')

/**
 * 判断模型是否具备某能力标签（Cherry 式「包含」判断，多能力模型可在多个标签下出现）。
 * 内置主流模型能力表（对齐 Cherry 的 provider-registry 思路）+ ID 关键词兜底。
 */
const modelHasTag = (m: { id: string; name?: string }, tag: string): boolean => {
  const s = (m.id + ' ' + (m.name || '')).toLowerCase()
  switch (tag) {
    case 'free':
      return s.includes('free')
    case 'embed':
      return /(bge|text-embedding|\bembed\b|rerank)/.test(s)
    case 'search':
      return /(search|web-search|websearch|perplexity|sonar)/.test(s)
    case 'vision':
      // 主流视觉模型能力表 + 关键词
      return /(gpt-(4o|4\.1|4\.5)|claude|gemini|qwen.*-vl|glm-\d+v|pixtral|minicpm|internvl|llava|llama-3\.2-vision|llama-4|vision|visual|image)/.test(s)
    case 'reason':
      // 主流推理模型能力表 + 关键词
      return /^o[134](-|$)/.test(s)
        || /^gpt-5\b/.test(s)
        || /deepseek-(r1|reasoner|v3\.1|v4)/.test(s)
        || /claude-(3-7|sonnet-4|opus-4|haiku-4)/.test(s)
        || /gemini-(2\.5|3)/.test(s)
        || /(qwq|qwen-max|glm-(z1|4\.5|5|5\.1)|kimi.*thinking|grok-[34]|llama-4|reasoner|\br1\b|thinking)/.test(s)
  }
  return false
}

const filteredModels = computed(() => {
  const q = mlSearch.value.trim().toLowerCase()
  const tag = mlTag.value
  return discoveredModels.value.filter(m => {
    if (q && !(m.id.toLowerCase().includes(q) || (m.name || '').toLowerCase().includes(q))) return false
    if (tag && !modelHasTag(m, tag)) return false
    return true
  })
})

/** Cherry Studio 式分组名：带 / 前缀的 ID 取厂商段，扁平 ID 取第一个 '-' 前的家族名（deepseek-v4-pro → deepseek） */
const deriveModelGroupName = (modelId: string): string | undefined => {
  const normalizedId = (modelId || '').trim()
  const pathParts = normalizedId.split('/')
  if (pathParts.length > 1) {
    return pathParts[0]?.trim() || undefined
  }
  const familyName = normalizedId.split('-')[0]?.trim()
  return familyName && familyName !== normalizedId ? familyName : undefined
}

/** 按厂商/家族分组展示（Cherry 式），无连字符的 ID 归入服务商名下 */
const groupedModels = computed(() => {
  const fallback = editingModel.value ? providerNameOf(editingModel.value.provider_id) : '其他'
  const groups = new Map<string, typeof filteredModels.value>()
  for (const m of filteredModels.value) {
    const g = deriveModelGroupName(m.id) ?? fallback
    const list = groups.get(g)
    if (list) list.push(m)
    else groups.set(g, [m])
  }
  return [...groups.entries()].map(([name, items]) => ({ name, items }))
})

const handleDiscoverModels = async () => {
  if (!editingModel.value?.provider_id) {
    ElMessage.warning('请先选择服务商')
    return
  }
  discoveringModels.value = true
  discoveredModels.value = []
  selectedModelId.value = ''
  mlSearch.value = ''
  mlTag.value = ''
  try {
    const res = await providerAPI.discoverModels(editingModel.value.provider_id)
    if (res.success && res.data && res.data.length > 0) {
      discoveredModels.value = res.data
      modelListDialogVisible.value = true
      ElMessage.success(`成功获取 ${res.data.length} 个可用模型`)
    } else {
      ElMessage.warning('该服务商未返回可用模型列表')
    }
  } catch (error) {
    ElMessage.error('获取模型列表失败，请检查 API 地址和密钥是否正确')
  } finally {
    discoveringModels.value = false
  }
}

const applyDiscoveredModel = () => {
  if (!selectedModelId.value) return
  const m = editingModel.value
  if (m) {
    m.model = selectedModelId.value
    if (!m.name) m.name = selectedModelId.value
    const selected = discoveredModels.value.find(x => x.id === selectedModelId.value)
    m.max_tokens = resolveMaxTokens(selectedModelId.value, selected?.max_tokens)
    if (!m.category) m.category = guessCategory(selectedModelId.value)
    dirty.value = true
  }
  modelListDialogVisible.value = false
}

/* ---------------- 服务商弹窗状态 ---------------- */
const providerDialogVisible = ref(false)
const isEditProvider = ref(false)
const providerAddMode = ref<'preset' | 'custom'>('preset')
const providerFormData = ref({
  id: 0,
  name: '',
  provider_type: 'openai',
  api_key: '',
  api_url: '',
  url_suffix_mode: 'compat' as 'compat' | 'none',
  api_format: 'openai' as 'openai' | 'claude' | 'gemini',
  use_full_url: false
})

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

.cat-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
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
.cfg-dialog.el-dialog {
  border-radius: 14px;
  overflow: hidden;
}

.cfg-dialog .el-dialog__header {
  padding: 18px 22px 14px;
  margin-right: 0;
  border-bottom: 1px solid var(--cfg-line-soft);
}

.cfg-dialog .el-dialog__title {
  font-size: 16px;
  font-weight: 700;
  color: var(--cfg-ink);
}

.cfg-dialog .el-dialog__headerbtn {
  top: 16px;
  right: 18px;
}

.cfg-dialog .el-dialog__body {
  padding: 20px 22px;
  color: var(--cfg-ink);
}

.cfg-dialog .el-dialog__footer {
  padding: 14px 22px;
  border-top: 1px solid var(--cfg-line-soft);
  background: var(--cfg-panel-2);
  border-radius: 0 0 14px 14px;
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

.cfg-dialog .ml-bar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.cfg-dialog .ml-bar .el-input {
  flex: 1;
}

.cfg-dialog .ml-count {
  font-size: 12px;
  color: #94a3a0;
  flex-shrink: 0;
}

.cfg-dialog .ml-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.cfg-dialog .ml-tag {
  font-size: 12px;
  color: #55635c;
  background: #f4f7f4;
  border: 1px solid #e2e8e2;
  border-radius: 999px;
  padding: 5px 13px;
  cursor: pointer;
  transition: 0.15s;
  user-select: none;
}

.cfg-dialog .ml-tag:hover {
  border-color: #94a3a0;
}

.cfg-dialog .ml-tag.on {
  background: #16a34a;
  border-color: #16a34a;
  color: #fff;
  font-weight: 600;
}

.cfg-dialog .ml-body {
  max-height: 420px;
  overflow-y: auto;
  border: 1px solid #e2e8e2;
  border-radius: 12px;
  padding: 10px;
}

.cfg-dialog .ml-grp-h {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  font-weight: 700;
  color: #55635c;
  padding: 10px 8px 6px;
}

.cfg-dialog .ml-grp-h .cnt {
  font-size: 10.5px;
  color: #94a3a0;
  background: #f4f7f4;
  border-radius: 999px;
  padding: 1px 8px;
}

.cfg-dialog .ml-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 9px;
  cursor: pointer;
  transition: 0.13s;
  border: 1px solid transparent;
}

.cfg-dialog .ml-item:hover {
  background: #f7faf7;
}

.cfg-dialog .ml-item.on {
  background: #ecfdf5;
  border-color: #bbf7d0;
}

.cfg-dialog .ml-radio {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid #d5ded6;
  flex-shrink: 0;
  transition: 0.15s;
}

.cfg-dialog .ml-radio.on {
  border-color: #16a34a;
  background: radial-gradient(circle, #16a34a 0 4px, transparent 5px);
}

.cfg-dialog .ml-info {
  flex: 1;
  min-width: 0;
}

.cfg-dialog .ml-id {
  font-size: 13.5px;
  font-weight: 600;
  font-family: Consolas, monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cfg-dialog .ml-meta {
  font-size: 11.5px;
  color: #94a3a0;
  margin-top: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.cfg-dialog .ml-empty {
  text-align: center;
  color: #94a3a0;
  font-size: 13px;
  padding: 34px;
}

/* 暗色：获取到的模型列表弹窗 */
:root[data-theme='dark'] .cfg-dialog .ml-tag {
  background: #1c2333;
  border-color: rgba(59, 130, 246, 0.16);
  color: #8b98a8;
}

:root[data-theme='dark'] .cfg-dialog .ml-tag:hover {
  border-color: #3b82f6;
}

:root[data-theme='dark'] .cfg-dialog .ml-tag.on {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #fff;
}

:root[data-theme='dark'] .cfg-dialog .ml-body {
  background: #161b22;
  border-color: rgba(59, 130, 246, 0.16);
}

:root[data-theme='dark'] .cfg-dialog .ml-grp-h {
  color: #8b98a8;
}

:root[data-theme='dark'] .cfg-dialog .ml-grp-h .cnt {
  background: #1c2333;
  color: #5f6b7a;
}

:root[data-theme='dark'] .cfg-dialog .ml-item:hover {
  background: #1c2333;
}

:root[data-theme='dark'] .cfg-dialog .ml-item.on {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.30);
}

:root[data-theme='dark'] .cfg-dialog .ml-radio {
  border-color: #2a3444;
}

:root[data-theme='dark'] .cfg-dialog .ml-radio.on {
  border-color: #3b82f6;
  background: radial-gradient(circle, #3b82f6 0 4px, transparent 5px);
}

:root[data-theme='dark'] .cfg-dialog .ml-id {
  color: #e6edf3;
}

:root[data-theme='dark'] .cfg-dialog .ml-meta,
:root[data-theme='dark'] .cfg-dialog .ml-count,
:root[data-theme='dark'] .cfg-dialog .ml-empty {
  color: #5f6b7a;
}
</style>

