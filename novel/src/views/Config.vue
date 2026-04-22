<template>
  <div class="config-container">
    <div class="header">
      <div class="header-content">
        <h2>API配置管理</h2>
        <p class="header-subtitle">统一管理服务商、模型配置与导入导出</p>
      </div>
      <div class="header-actions">
        <el-dropdown trigger="click" @command="handleImportCommand">
          <el-button type="success" size="default">
            <el-icon><Upload /></el-icon>
            导入
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
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
          <el-button type="warning" size="default">
            <el-icon><Download /></el-icon>
            导出
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="exportAll">
                <el-icon><Document /></el-icon>
                导出全部配置
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div class="overview-grid">
      <div class="overview-card">
        <span class="overview-label">服务商数量</span>
        <strong class="overview-value">{{ providers.length }}</strong>
      </div>
      <div class="overview-card">
        <span class="overview-label">模型数量</span>
        <strong class="overview-value">{{ models.length }}</strong>
      </div>
      <div class="overview-card">
        <span class="overview-label">默认模型</span>
        <strong class="overview-value">{{ defaultModelCount }}</strong>
      </div>
    </div>

    <el-tabs v-model="activeTab" type="border-card" class="config-tabs">
      <el-tab-pane label="服务商管理" name="providers">
        <div class="tab-header">
          <h3>服务商列表</h3>
          <el-button type="primary" @click="handleCreateProvider">
            <el-icon><Plus /></el-icon>
            添加服务商
          </el-button>
        </div>
        <el-table :data="providers" stripe style="width: 100%">
          <el-table-column prop="name" label="服务商名称" width="200" />
          <el-table-column prop="provider_type" label="服务商类型" width="150">
            <template #default="{ row }">
              <el-tag>{{ getProviderTypeLabel(row.provider_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="api_url" label="API地址" show-overflow-tooltip />
          <el-table-column label="模型数量" width="100">
            <template #default="{ row }">
              <el-tag type="info">{{ getProviderModelCount(row.id) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" @click="handleEditProvider(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteProvider(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="模型管理" name="models">
        <div class="tab-header">
          <h3>模型列表</h3>
          <el-button type="primary" @click="handleCreateModel">
            <el-icon><Plus /></el-icon>
            添加模型
          </el-button>
        </div>
        <el-table :data="models" stripe style="width: 100%">
          <el-table-column prop="name" label="模型名称" width="180" />
          <el-table-column prop="provider_name" label="服务商" width="150">
            <template #default="{ row }">
              <el-tag>{{ row.provider_name }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="model" label="模型标识" width="180" />
          <el-table-column prop="temperature" label="温度" width="80" />
          <el-table-column prop="max_tokens" label="最大Token" width="120" />
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag v-if="row.is_default" type="success">是</el-tag>
              <span v-else>否</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button
                size="small"
                type="success"
                :loading="testingModelId === row.id"
                @click="handleTestModel(row)"
              >
                测试
              </el-button>
              <el-button size="small" type="primary" @click="handleEditModel(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteModel(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-dialog
      v-model="providerDialogVisible"
      :title="isEditProvider ? '编辑服务商' : '添加服务商'"
      width="600px"
    >
      <el-form :model="providerFormData" label-width="100px">
        <el-form-item label="服务商名称" required>
          <el-input v-model="providerFormData.name" placeholder="请输入服务商名称" />
        </el-form-item>
        <el-form-item label="服务商类型" required>
          <el-select v-model="providerFormData.provider_type" placeholder="请选择服务商类型" style="width: 100%">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Azure OpenAI" value="azure" />
            <el-option label="Claude" value="claude" />
            <el-option label="通义千问" value="qwen" />
            <el-option label="DeepSeek" value="deepseek" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="API地址" required>
          <el-input v-model="providerFormData.api_url" placeholder="例如: https://api.openai.com/v1/chat/completions" />
        </el-form-item>
        <el-form-item label="API密钥" required>
          <el-input v-model="providerFormData.api_key" type="password" placeholder="请输入API密钥" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="providerDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitProvider">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="modelDialogVisible"
      :title="isEditModel ? '编辑模型' : '添加模型'"
      width="600px"
    >
      <el-form :model="modelFormData" label-width="100px">
        <el-form-item label="选择服务商" required>
          <el-select v-model="modelFormData.provider_id" placeholder="请选择服务商" style="width: 100%">
            <el-option
              v-for="provider in providers"
              :key="provider.id"
              :label="provider.name"
              :value="provider.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模型名称" required>
          <el-input v-model="modelFormData.name" placeholder="请输入模型名称，如 GPT-4" />
        </el-form-item>
        <el-form-item label="模型标识" required>
          <el-input v-model="modelFormData.model" placeholder="例如: gpt-4, claude-3-opus-20240229" />
        </el-form-item>
        <el-form-item label="温度">
          <el-slider v-model="modelFormData.temperature" :min="0" :max="2" :step="0.1" />
          <span style="margin-left: 10px;">{{ modelFormData.temperature }}</span>
        </el-form-item>
        <el-form-item label="最大Token">
          <el-input-number v-model="modelFormData.max_tokens" :min="100" :max="128000" :step="100" style="width: 100%" />
        </el-form-item>
        <el-form-item label="设为默认">
          <el-switch v-model="modelFormData.is_default" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitModel">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="importDialogVisible"
      title="导入API配置"
      width="600px"
      :close-on-click-modal="false"
    >
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
            <div class="el-upload__text">
              将配置文件拖到此处，或<em>点击上传</em>
            </div>
            <template #tip>
              <div class="el-upload__tip">
                只能上传 JSON 格式的配置文件
              </div>
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
            <el-checkbox v-model="importOptions.overrideExisting">
              覆盖同名配置（不勾选则跳过同名配置）
            </el-checkbox>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!importPreview" @click="executeImport">
          导入
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Download, ArrowDown, Document, UploadFilled } from '@element-plus/icons-vue'
import { providerAPI, configAPI } from '@/api'
import type { ApiProvider, ApiModel } from '@/types'

const activeTab = ref('providers')

const providers = ref<ApiProvider[]>([])
const providerDialogVisible = ref(false)
const isEditProvider = ref(false)
const providerFormData = ref({
  id: 0,
  name: '',
  provider_type: 'openai',
  api_key: '',
  api_url: ''
})

const models = ref<ApiModel[]>([])
const modelDialogVisible = ref(false)
const isEditModel = ref(false)
const testingModelId = ref<number | null>(null)
const modelFormData = ref({
  id: 0,
  provider_id: 0,
  name: '',
  model: '',
  temperature: 0.7,
  max_tokens: 2000,
  is_default: false
})

const providerTypeCount = computed(() => new Set(providers.value.map(provider => provider.provider_type)).size)
const defaultModelCount = computed(() => models.value.filter(model => model.is_default === 1).length)

onMounted(async () => {
  await fetchProviders()
  await initializeDefaultProviders()
  await fetchModels()
})

const fetchProviders = async () => {
  const res = await providerAPI.getAll()
  if (res.success && res.data) {
    providers.value = res.data
  }
}

const initializeDefaultProviders = async () => {
  if (providers.value.length > 0) {
    return
  }
  
  try {
    const defaultProviders = [
      {
        name: 'DeepSeek',
        provider_type: 'deepseek',
        api_key: '',
        api_url: 'https://api.deepseek.com/v1/chat/completions'
      },
      {
        name: 'SiliconFlow',
        provider_type: 'other',
        api_key: '',
        api_url: 'https://api.siliconflow.cn/v1/chat/completions'
      },
      {
        name: 'bltcy.ai',
        provider_type: 'other',
        api_key: '',
        api_url: 'https://api.bltcy.ai/v1/chat/completions'
      },
      {
        name: 'ModelScope',
        provider_type: 'other',
        api_key: '',
        api_url: 'https://api-inference.modelscope.cn/v1/chat/completions'
      }
    ]
    
    let deepSeekProviderId: number | null = null
    let siliconFlowProviderId: number | null = null
    let modelScopeProviderId: number | null = null
    let bltcyProviderId: number | null = null
    
    for (const provider of defaultProviders) {
      const res = await providerAPI.create(provider)
      if (res.success && res.data) {
        ElMessage.success(`已自动添加 ${provider.name} 服务商，请配置 API 密钥后使用`)
        
        if (provider.name === 'DeepSeek') {
          deepSeekProviderId = res.data.id
          
          const deepSeekModels = [
            {
              provider_id: res.data.id,
              name: 'DeepSeek Chat',
              model: 'deepseek-chat',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 1
            },
            {
              provider_id: res.data.id,
              name: 'DeepSeek Reasoner',
              model: 'deepseek-reasoner',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            }
          ]
          
          for (const model of deepSeekModels) {
            await configAPI.create(model)
          }
          
          ElMessage.success('已为 DeepSeek 添加默认模型：deepseek-chat, deepseek-reasoner')
        }
        
        if (provider.name === 'SiliconFlow') {
          siliconFlowProviderId = res.data.id
          
          const siliconFlowModels = [
            {
              provider_id: res.data.id,
              name: 'GLM-5',
              model: 'Pro/zai-org/GLM-5',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 1
            },
            {
              provider_id: res.data.id,
              name: 'DeepSeek-R1',
              model: 'deepseek-ai/DeepSeek-R1',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            }
          ]
          
          for (const model of siliconFlowModels) {
            await configAPI.create(model)
          }
          
          ElMessage.success('已为 SiliconFlow 添加默认模型：GLM-5, DeepSeek-R1')
        }
        
        if (provider.name === 'bltcy.ai') {
          bltcyProviderId = res.data.id
          
          const bltcyModels = [
            {
              provider_id: res.data.id,
              name: 'Gemini 3 Pro Preview',
              model: 'gemini-3-pro-preview',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 1
            },
            {
              provider_id: res.data.id,
              name: 'Gemini 2.5 Pro',
              model: 'gemini-2.5-pro',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            },
            {
              provider_id: res.data.id,
              name: 'Gemini 3 Flash Preview',
              model: 'gemini-3-flash-preview',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            },
            {
              provider_id: res.data.id,
              name: 'Gemini 2.5 Flash Thinking',
              model: 'gemini-2.5-flash-thinking',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            },
            {
              provider_id: res.data.id,
              name: 'Gemini 2.5 Pro Thinking',
              model: 'gemini-2.5-pro-thinking-*',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            },
            {
              provider_id: res.data.id,
              name: 'DeepSeek R1',
              model: 'deepseek-r1-250528',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            }
          ]
          
          for (const model of bltcyModels) {
            await configAPI.create(model)
          }
          
          ElMessage.success('已为 bltcy.ai 添加默认模型：Gemini 3 Pro Preview, Gemini 2.5 Pro, Gemini 3 Flash Preview, Gemini 2.5 Flash Thinking, Gemini 2.5 Pro Thinking')
        }
        
        if (provider.name === 'ModelScope') {
          modelScopeProviderId = res.data.id
          
          const modelScopeModels = [
            {
              provider_id: res.data.id,
              name: 'Kimi-K2.5',
              model: 'moonshotai/Kimi-K2.5',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 1
            },
            {
              provider_id: res.data.id,
              name: 'GLM-5',
              model: 'ZhipuAI/GLM-5',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            },
            {
              provider_id: res.data.id,
              name: 'GLM-5.1',
              model: 'ZhipuAI/GLM-5.1',
              temperature: 0.7,
              max_tokens: 12800,
              is_default: 0
            }
          ]
          
          for (const model of modelScopeModels) {
            await configAPI.create(model)
          }
          
          ElMessage.success('已为 ModelScope 添加默认模型：Kimi-K2.5, GLM-5, GLM-5.1')
        }
      }
    }
    
    await fetchProviders()
    await fetchModels()
  } catch (error) {
    console.error('初始化默认服务商失败:', error)
  }
}

const fetchModels = async () => {
  const res = await configAPI.getAll()
  if (res.success && res.data) {
    models.value = res.data
  }
}

const getProviderTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    openai: 'OpenAI',
    azure: 'Azure OpenAI',
    claude: 'Claude',
    qwen: '通义千问',
    deepseek: 'DeepSeek',
    other: '其他'
  }
  return map[type] || type
}

const getProviderModelCount = (providerId: number) => {
  return models.value.filter((m: ApiModel) => m.provider_id === providerId).length
}

const handleCreateProvider = () => {
  isEditProvider.value = false
  providerFormData.value = {
    id: 0,
    name: '',
    provider_type: 'openai',
    api_key: '',
    api_url: ''
  }
  providerDialogVisible.value = true
}

const handleEditProvider = (provider: ApiProvider) => {
  isEditProvider.value = true
  providerFormData.value = {
    id: provider.id,
    name: provider.name,
    provider_type: provider.provider_type,
    api_key: provider.api_key,
    api_url: provider.api_url
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

const handleSubmitProvider = async () => {
  if (!providerFormData.value.name || !providerFormData.value.provider_type || !providerFormData.value.api_url) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    if (isEditProvider.value) {
      const res = await providerAPI.update(providerFormData.value.id, providerFormData.value)
      if (res.success) {
        ElMessage.success('更新成功')
      }
    } else {
      const res = await providerAPI.create(providerFormData.value)
      if (res.success) {
        ElMessage.success('创建成功')
      }
    }
    providerDialogVisible.value = false
    await fetchProviders()
    await fetchModels()
  } catch (error) {
  }
}

const handleCreateModel = () => {
  if (providers.value.length === 0) {
    ElMessage.warning('请先添加服务商')
    return
  }
  isEditModel.value = false
  modelFormData.value = {
    id: 0,
    provider_id: providers.value[0]?.id || 0,
    name: '',
    model: '',
    temperature: 0.7,
    max_tokens: 2000,
    is_default: false
  }
  modelDialogVisible.value = true
}

const handleEditModel = (model: ApiModel) => {
  isEditModel.value = true
  modelFormData.value = {
    id: model.id,
    provider_id: model.provider_id,
    name: model.name,
    model: model.model,
    temperature: model.temperature,
    max_tokens: model.max_tokens,
    is_default: model.is_default === 1
  }
  modelDialogVisible.value = true
}

const handleDeleteModel = async (model: ApiModel) => {
  try {
    await ElMessageBox.confirm(`确定删除模型"${model.name}"吗？`, '提示', { type: 'warning' })
    const res = await configAPI.delete(model.id)
    if (res.success) {
      ElMessage.success('删除成功')
      await fetchModels()
    }
  } catch (error) {
  }
}

const handleTestModel = async (model: ApiModel) => {
  if (testingModelId.value === model.id) return
  try {
    testingModelId.value = model.id
    const res = await configAPI.testModel(model.id)
    if (res.success && res.data) {
      await ElMessageBox.alert(
        `服务商：${res.data.providerName || model.provider_name || '未命名服务商'}\n耗时：${res.data.latencyMs} ms\n返回：${res.data.preview}`,
        `模型测试成功 - ${res.data.modelName || model.name}`,
        { confirmButtonText: '知道了' }
      )
    }
  } catch (error) {
  } finally {
    testingModelId.value = null
  }
}

const handleSubmitModel = async () => {
  if (!modelFormData.value.provider_id || !modelFormData.value.name || !modelFormData.value.model) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    const submitData = {
      ...modelFormData.value,
      is_default: modelFormData.value.is_default ? 1 : 0
    }
    if (isEditModel.value) {
      const res = await configAPI.update(modelFormData.value.id, submitData)
      if (res.success) {
        ElMessage.success('更新成功')
      }
    } else {
      const res = await configAPI.create(submitData)
      if (res.success) {
        ElMessage.success('创建成功')
      }
    }
    modelDialogVisible.value = false
    await fetchModels()
  } catch (error) {
  }
}

const importDialogVisible = ref(false)
const importPreview = ref<{
  version: string
  exportTime: string
  providers: Partial<ApiProvider>[]
  models: Partial<ApiModel>[]
} | null>(null)
const importOptions = ref({ overrideExisting: false })

const handleImportCommand = (command: string) => {
  if (command === 'importConfig') {
    openImportDialog()
  }
}

const handleExportCommand = (command: string) => {
  if (command === 'exportAll') {
    executeExportAll()
  }
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
            is_default: model.is_default ? 1 : 0
          })
          if (res.success) {
            modelSuccessCount++
          }
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
          is_default: model.is_default ? 1 : 0
        })
        if (res.success) {
          modelSuccessCount++
        }
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
      is_default: m.is_default
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
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header-content h2 {
  margin: 0 0 8px 0;
  font-size: 24px;
}

.header-subtitle {
  margin: 0;
  color: #666;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.overview-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.overview-card {
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: center;
}

.overview-label {
  display: block;
  color: #666;
  font-size: 14px;
  margin-bottom: 8px;
}

.overview-value {
  display: block;
  font-size: 28px;
  color: #333;
}

.config-tabs {
  background: var(--tabs-content-bg, #fff);
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.tab-header h3 {
  margin: 0;
  font-size: 18px;
}

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
  background: #f5f7fa;
  border-radius: 8px;
}

@media (max-width: 768px) {
  .overview-grid {
    grid-template-columns: 1fr;
  }

  .header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }
}
</style>