<template>
  <div class="pack-detail-container">
    <div class="pack-detail-header">
      <div class="pack-detail-title-block">
        <button class="back-btn" @click="goBack" title="返回">
          <el-icon><ArrowLeft /></el-icon>
        </button>
        <h2 class="pack-detail-title">{{ packName }}</h2>
        <span class="pack-detail-count">{{ filteredPrompts.length }} 张卡片</span>
      </div>
      <div class="pack-detail-search">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索提示词..."
          clearable
          prefix-icon="Search"
          class="pack-search-input"
        />
        <el-button type="primary" @click="openCreateInCurrentPack">
          <el-icon><Plus /></el-icon>
          新增提示词
        </el-button>
      </div>
    </div>

    <!-- 排序 Tab -->
    <div class="pack-sort-tab-bar">
      <div class="pack-sort-tabs">
        <span
          v-for="tab in sortTabs"
          :key="tab.value"
          class="pack-sort-tab-item"
          :class="{ active: selectedSort === tab.value }"
          @click="selectedSort = tab.value"
        >
          <el-icon v-if="tab.icon" class="pack-sort-tab-icon">
            <component :is="tab.icon" />
          </el-icon>
          {{ tab.label }}
        </span>
      </div>
    </div>

    <!-- 提示词卡片列表 -->
    <div v-if="filteredPrompts.length === 0" class="pack-empty-hint">
      <el-empty :description="searchKeyword ? '没有匹配的提示词' : '该卡包暂无可预览的卡片'" />
    </div>
    <div v-else class="pack-prompts-grid">
      <div
        v-for="prompt in filteredPrompts"
        :key="prompt.id"
        class="pack-prompt-card"
        :class="{ encrypted: prompt.card_type === 'encrypted' }"
        @click="handleEdit(prompt)"
      >
        <span
          class="pack-card-category-badge"
          :style="{ borderColor: getCategoryColor(prompt.category), color: getCategoryColor(prompt.category) }"
        >
          {{ getCategoryLabel(prompt.category) }}
        </span>

        <div class="pack-card-title-row">
          <h3 class="pack-card-title" :title="prompt.name">
            <el-icon v-if="prompt.card_type === 'encrypted'" class="pack-card-lock"><Lock /></el-icon>
            {{ prompt.name }}
          </h3>
        </div>

        <div class="pack-card-author-row">
          <div class="pack-author-avatar" :style="{ background: getAvatarColor(prompt.creator_name || prompt.name) }">
            {{ (prompt.creator_name || '匿').charAt(0) }}
          </div>
          <span class="pack-author-name" :title="prompt.creator_name">
            {{ prompt.creator_name || '匿名用户' }}
          </span>
          <span class="pack-card-stats">
            <el-icon class="pack-stat-icon"><Star /></el-icon>
            {{ prompt.use_count ?? prompt.order_num ?? 0 }}
          </span>
          <span class="pack-card-date">{{ formatDate(prompt.created_at) }}</span>
        </div>

        <div class="pack-card-body-preview" :title="getPromptDescriptionPreview(prompt)">
          {{ getPromptDescriptionPreview(prompt) || '暂无简介' }}
        </div>

        <div v-if="(prompt.subcategories || []).length > 0" class="pack-card-footer-tags">
          <span
            v-for="(subcat, index) in (prompt.subcategories || []).slice(0, 3)"
            :key="index"
            class="pack-subcategory-tag"
            :style="{
              background: getTagColor(subcat).bg,
              color: getTagColor(subcat).text,
              borderColor: getTagColor(subcat).border
            }"
          >
            {{ subcat }}
          </span>
          <span
            v-if="(prompt.subcategories || []).length > 3"
            class="pack-more-tags"
          >+{{ (prompt.subcategories || []).length - 3 }}</span>
        </div>
        <div v-else class="pack-card-footer-tags pack-card-footer-tags--empty"></div>
      </div>
    </div>

    <!-- 编辑提示词弹窗 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="isCreating ? '新增提示词' : '编辑提示词'"
      width="900px"
      top="6vh"
      :close-on-click-modal="false"
      destroy-on-close
      class="pack-edit-dialog"
    >
      <div class="pack-edit-layout">
        <!-- 左侧：表单字段 (2/3) -->
        <div class="pack-edit-left">
          <el-form :model="formData" label-position="top" class="pack-edit-form">
            <el-form-item label="名称" required>
              <el-input
                v-model="formData.name"
                placeholder="请输入名称"
                maxlength="100"
                show-word-limit
              />
            </el-form-item>

            <el-form-item label="分类" required>
              <el-select
                v-model="formData.category"
                placeholder="请选择卡包"
                style="width: 100%"
              >
                <el-option
                  v-for="cat in categoryOptions"
                  :key="cat"
                  :label="cat"
                  :value="cat"
                />
              </el-select>
            </el-form-item>

            <el-form-item label="作者">
              <el-input v-model="formData.creator_name" placeholder="请输入作者名，非必填" />
            </el-form-item>

            <el-form-item label="版本">
              <el-input v-model="formData.version" placeholder="如 v1.0、v2.1.0，非必填" />
            </el-form-item>

            <el-form-item label="卡片类型">
              <div class="pack-edit-card-type">
                <el-radio-group v-model="formData.card_type" @change="handleCardTypeChange">
                  <el-radio value="normal">
                    <el-icon><Document /></el-icon>
                    普通卡片
                  </el-radio>
                  <el-radio value="encrypted">
                    <el-icon><Lock /></el-icon>
                    加密卡片
                  </el-radio>
                </el-radio-group>
                <div v-if="formData.card_type === 'encrypted'" class="pack-edit-card-type-actions">
                  <el-button
                    v-if="!formData.password"
                    type="warning"
                    size="small"
                    @click="openSetPasswordDialog"
                  >
                    <el-icon><Lock /></el-icon>
                    设置密码
                  </el-button>
                  <el-button
                    v-else
                    type="success"
                    size="small"
                    @click="openSetPasswordDialog"
                  >
                    <el-icon><Lock /></el-icon>
                    修改密码
                  </el-button>
                  <el-button
                    v-if="formData.password"
                    type="danger"
                    size="small"
                    link
                    @click="removeCardPassword"
                  >
                    移除密码
                  </el-button>
                </div>
                <div v-if="formData.card_type === 'encrypted' && formData.password" class="pack-edit-card-type-status">
                  <el-tag type="warning" size="small">
                    <el-icon><Lock /></el-icon>
                    已加密
                  </el-tag>
                </div>
                <div v-if="formData.card_type === 'encrypted' && !formData.password" class="pack-edit-card-type-status">
                  <el-tag type="info" size="small">未设置密码（保存前需设置密码）</el-tag>
                </div>
              </div>
            </el-form-item>

            <el-form-item label="简介">
              <SplitRichTextEditor
                v-model="formData.description"
                placeholder="用于介绍这个提示词，非必填"
                class="description-rich-editor"
              />
            </el-form-item>

            <el-form-item label="内容" required>
              <el-input
                v-model="formData.content"
                type="textarea"
                :rows="12"
                placeholder="请输入提示词内容（作为AI的system层指令），使用 ${字段名} 引用变量"
                class="pack-edit-content-input"
              />
            </el-form-item>

            <div class="pack-edit-section">
              <div class="pack-edit-section-head">
                <span class="pack-edit-section-title">标签</span>
                <span class="pack-edit-section-tip">用于卡包内多维度筛选</span>
              </div>
              <div class="pack-edit-tags">
                <el-tag
                  v-for="(tag, i) in formSubcategories"
                  :key="i"
                  size="small"
                  closable
                  class="pack-edit-tag"
                  @close="removeFormSubcategory(i)"
                >
                  {{ tag }}
                </el-tag>
                <el-input
                  v-model="newSubcategory"
                  size="small"
                  placeholder="输入后回车添加"
                  class="pack-edit-tag-input"
                  @keyup.enter="addFormSubcategory"
                />
                <el-button size="small" type="primary" plain @click="addFormSubcategory">
                  <el-icon><Plus /></el-icon>
                </el-button>
              </div>
            </div>
          </el-form>
        </div>

        <!-- 右侧：字段配置 (1/3) -->
        <div class="pack-edit-right">
          <div class="pack-edit-right-head">
            <span class="pack-edit-right-title">字段配置</span>
            <el-button type="primary" size="small" plain @click="addField">
              <el-icon><Plus /></el-icon>
              添加
            </el-button>
          </div>
          <div class="pack-edit-right-tip">对应内容里的 <code>${字段名}</code></div>
          <div class="pack-edit-right-body">
            <div v-if="fieldsConfig.length === 0" class="pack-edit-empty">暂无字段</div>
            <div v-else class="pack-edit-fields">
              <div v-for="(field, index) in fieldsConfig" :key="index" class="pack-edit-field">
                <div class="pack-edit-field-head">
                  <el-tag :type="field.required ? 'danger' : 'info'" size="small" effect="light">
                    {{ field.required ? '必填' : '选填' }}
                  </el-tag>
                  <el-input
                    v-model="field.name"
                    size="small"
                    class="pack-edit-field-name"
                    placeholder="字段名"
                    @change="updateFieldName(index, field.name)"
                  />
                  <el-button size="small" link @click="copyFieldName(field.name)" title="复制">
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                  <el-switch
                    v-model="field.required"
                    size="small"
                    inline-prompt
                    active-text="必"
                    inactive-text="选"
                    style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #c0c4cc;"
                  />
                  <el-button size="small" type="danger" link @click="removeField(index)" title="删除">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <div class="pack-edit-field-body">
                  <el-input v-model="field.label" size="small" placeholder="显示名" />
                  <el-select v-model="field.type" size="small" placeholder="类型" style="width: 100%">
                    <el-option label="单行文本" value="text" />
                    <el-option label="多行文本" value="textarea" />
                    <el-option label="下拉选择" value="select" />
                  </el-select>
                  <el-input
                    v-model="field.description"
                    size="small"
                    placeholder="说明（可选）"
                  />
                </div>
                <div class="pack-edit-field-order">
                  <el-button size="small" link :disabled="index === 0" @click="moveField(index, index - 1)" title="上移">
                    <el-icon><ArrowUp /></el-icon>
                  </el-button>
                  <el-button size="small" link :disabled="index === fieldsConfig.length - 1" @click="moveField(index, index + 1)" title="下移">
                    <el-icon><ArrowDown /></el-icon>
                  </el-button>
                </div>
                <div v-if="field.type === 'select'" class="pack-edit-field-options">
                  <el-input
                    v-model="field.optionsText"
                    type="textarea"
                    :rows="2"
                    size="small"
                    placeholder="选项（每行一个）"
                    @change="updateFieldOptions(index)"
                  />
                  <el-input
                    v-model="field.optionLabelsText"
                    type="textarea"
                    :rows="2"
                    size="small"
                    placeholder="展示名（可选）"
                    @change="updateFieldOptionLabels(index)"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSaving" @click="submitEdit">{{ isCreating ? '创建' : '保存' }}</el-button>
      </template>
    </el-dialog>

    <!-- 解锁密码弹窗 -->
    <el-dialog
      v-model="unlockPasswordDialogVisible"
      title="输入密码以解锁卡片"
      width="400px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="unlockPasswordForm" label-position="top">
        <el-form-item label="密码" required>
          <el-input
            v-model="unlockPasswordForm.password"
            type="password"
            placeholder="请输入卡片密码"
            show-password
            @keyup.enter="unlockAndEdit"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="unlockPasswordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="unlockAndEdit">解锁并编辑</el-button>
      </template>
    </el-dialog>

    <!-- 设置密码弹窗 -->
    <el-dialog
      v-model="passwordDialogVisible"
      title="设置卡片密码"
      width="400px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-form :model="passwordForm" label-position="top">
        <el-form-item label="密码" required>
          <el-input
            v-model="passwordForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" required>
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="setPassword">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, Plus, Search, Star, Lock, CopyDocument, ArrowUp, ArrowDown, Delete, Document } from '@element-plus/icons-vue'

const DEFAULT_CATEGORIES = ['默认', '写作要求', '写作风格', '续写', '脑洞', '书名', '简介', '大纲', '细纲', '黄金开篇', '金手指', '名字', '人设', '世界观', '卡包']
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

const route = useRoute()
const router = useRouter()

const packName = computed(() => decodeURIComponent(String(route.params.name || '')))

const prompts = ref<Prompt[]>([])
const searchKeyword = ref('')
const selectedSort = ref('latest')

// 编辑弹窗相关
const editDialogVisible = ref(false)
const editSaving = ref(false)
const isCreating = ref(false)
const formData = ref({
  id: 0,
  name: '',
  description: '',
  content: '',
  category: '',
  order_num: 0,
  card_type: 'normal' as 'normal' | 'encrypted',
  password: null as string | null,
  creator_name: '',
  version: ''
})

// 字段配置
type FieldConfig = {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  optionLabels: string[]
  optionsText: string
  optionLabelsText: string
  description: string
  required: boolean
}
const fieldsConfig = ref<FieldConfig[]>([])

// 标签
const formSubcategories = ref<string[]>([])
const newSubcategory = ref('')

// 分类选项
const categoryOptions = computed(() => {
  const names = new Set<string>(DEFAULT_CATEGORIES)
  if (formData.value.category) {
    names.add(formData.value.category)
  }
  prompts.value.forEach((p) => {
    if (p.category) names.add(p.category)
  })
  return Array.from(names)
})

// 加密/解密常量
const ENCRYPT_MARKER = 'XNC1:'
const hashPassword = async (password: string): Promise<string> => {
  const encoder = new TextEncoder()
  const data = encoder.encode(password + '__xingnovel_salt__')
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  return hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
}
const verifyPassword = async (inputPassword: string, storedHash: string): Promise<boolean> => {
  const inputHash = await hashPassword(inputPassword)
  return inputHash === storedHash
}
const bufferToBase64 = (buffer: ArrayBuffer): string => {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  bytes.forEach(b => (binary += String.fromCharCode(b)))
  return btoa(binary)
}
const base64ToBuffer = (base64: string): ArrayBuffer => {
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}
const decryptContent = async (wrapped: string, password: string): Promise<string | null> => {
  if (!wrapped.startsWith(ENCRYPT_MARKER)) {
    return wrapped
  }
  try {
    const enc = new TextEncoder()
    const dec = new TextDecoder()
    const combined = new Uint8Array(base64ToBuffer(wrapped.slice(ENCRYPT_MARKER.length)))
    const salt = combined.slice(0, 16)
    const iv = combined.slice(16, 28)
    const ciphertext = combined.slice(28)
    const keyMaterial = await crypto.subtle.importKey('raw', enc.encode(password), 'PBKDF2', false, ['deriveKey'])
    const key = await crypto.subtle.deriveKey(
      { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['decrypt']
    )
    const plaintext = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext)
    return dec.decode(plaintext)
  } catch {
    return null
  }
}

// 解锁弹窗
const unlockPasswordDialogVisible = ref(false)
const unlockPasswordForm = ref({ password: '' })
let unlockingPrompt: Prompt | null = null

// 密码弹窗
const passwordDialogVisible = ref(false)
const passwordForm = ref({
  password: '',
  confirmPassword: ''
})

const handleCardTypeChange = (value: string) => {
  if (value === 'normal') {
    formData.value.password = null
  }
}

const openSetPasswordDialog = () => {
  passwordForm.value = {
    password: '',
    confirmPassword: ''
  }
  passwordDialogVisible.value = true
}

const removeCardPassword = () => {
  formData.value.password = null
  ElMessage.success('密码已移除')
}

const setPassword = () => {
  if (!passwordForm.value.password.trim()) {
    ElMessage.warning('请输入密码')
    return
  }
  if (passwordForm.value.password !== passwordForm.value.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  formData.value.password = passwordForm.value.password
  passwordDialogVisible.value = false
  ElMessage.success('密码设置成功')
}

const sortTabs: Array<{ label: string; value: string; icon?: any }> = [
  { label: '最新', value: 'latest' },
  { label: '最热', value: 'hottest' },
  { label: '名称A-Z', value: 'name_asc' },
]

const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
}

const packPrompts = computed(() => {
  return prompts.value.filter((p) => (p.category || '默认') === packName.value)
})

const filteredPrompts = computed(() => {
  let result = packPrompts.value
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (keyword) {
    result = result.filter(
      (p) =>
        p.name.toLowerCase().includes(keyword) ||
        p.content.toLowerCase().includes(keyword) ||
        (p.description || '').toLowerCase().includes(keyword)
    )
  }
  const list = [...result]
  switch (selectedSort.value) {
    case 'latest':
      list.sort((a, b) => {
        const ta = a.created_at ? new Date(a.created_at).getTime() : 0
        const tb = b.created_at ? new Date(b.created_at).getTime() : 0
        return tb - ta
      })
      break
    case 'hottest':
      list.sort((a, b) => (b.order_num || 0) - (a.order_num || 0))
      break
    case 'name_asc':
      list.sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
      break
  }
  return list
})

const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '未知时间'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })
}

const stripHtml = (value: string) => {
  if (!value) return ''
  if (typeof window === 'undefined') return value.replace(/<[^>]+>/g, ' ')
  const temp = document.createElement('div')
  temp.innerHTML = value
  return temp.textContent || temp.innerText || ''
}

const getPromptDescriptionPreview = (prompt: Prompt) => {
  const description = stripHtml(prompt.description || '').trim()
  if (!description) return ''
  return `${description.slice(0, 100)}${description.length > 100 ? '...' : ''}`
}

const colorPalette = [
  'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
  'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
  'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
  'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
  'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)',
  'linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%)',
  'linear-gradient(135deg, #f5576c 0%, #ff8a5c 100%)',
  'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
]

const categoryColorMap: Record<string, string> = {}
const categoryLabelMap: Record<string, string> = {}
let colorIndex = 0
const getCategoryColor = (category: string): string => {
  if (!categoryColorMap[category]) {
    categoryColorMap[category] = colorPalette[colorIndex % colorPalette.length]
    colorIndex++
  }
  return categoryColorMap[category]
}
const getCategoryLabel = (category: string): string => {
  if (!categoryLabelMap[category]) {
    categoryLabelMap[category] = (category || '').slice(0, 2)
  }
  return categoryLabelMap[category]
}

const avatarColors = [
  '#667eea', '#f093fb', '#4facfe', '#43e97b', '#fa709a',
  '#a18cd1', '#fccb90', '#f5576c', '#764ba2', '#00f2fe',
]
const avatarColorCache: Record<string, string> = {}
const getAvatarColor = (name: string): string => {
  if (!avatarColorCache[name]) {
    let hash = 0
    for (let i = 0; i < name.length; i++) {
      hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    avatarColorCache[name] = avatarColors[Math.abs(hash) % avatarColors.length]
  }
  return avatarColorCache[name]
}

const tagColorPalette = [
  { bg: '#fef3c7', text: '#d97706', border: '#fcd34d' },
  { bg: '#dbeafe', text: '#2563eb', border: '#93c5fd' },
  { bg: '#dcfce7', text: '#16a34a', border: '#86efac' },
  { bg: '#fce7f3', text: '#db2777', border: '#f9a8d4' },
  { bg: '#f3e8ff', text: '#9333ea', border: '#c4b5fd' },
  { bg: '#fff7ed', text: '#ea580c', border: '#fdba74' },
  { bg: '#ecfeff', text: '#0891b2', border: '#67e8f9' },
  { bg: '#fef2f2', text: '#dc2626', border: '#fca5a5' },
  { bg: '#f5f3ff', text: '#6d28d9', border: '#c4b5fd' },
  { bg: '#ecfdf5', text: '#059669', border: '#6ee7b7' },
]
const tagColorCache: Record<string, { bg: string; text: string; border: string }> = {}
const getTagColor = (tagText: string) => {
  if (!tagColorCache[tagText]) {
    let hash = 0
    for (let i = 0; i < tagText.length; i++) {
      hash = tagText.charCodeAt(i) + ((hash << 5) - hash)
    }
    tagColorCache[tagText] = tagColorPalette[Math.abs(hash) % tagColorPalette.length]
  }
  return tagColorCache[tagText]
}

const goBack = () => {
  router.push({ name: 'Prompts' })
}

const openCreateInCurrentPack = () => {
  isCreating.value = true
  formData.value = {
    id: 0,
    name: '',
    description: '',
    content: '',
    category: packName.value,
    order_num: 0,
    card_type: 'normal',
    password: null,
    creator_name: '',
    version: ''
  }
  fieldsConfig.value = []
  formSubcategories.value = []
  newSubcategory.value = ''
  editDialogVisible.value = true
}

const handleEdit = (prompt: Prompt) => {
  if (prompt.card_type === 'encrypted' && prompt.content.startsWith(ENCRYPT_MARKER)) {
    unlockingPrompt = prompt
    unlockPasswordForm.value = { password: '' }
    unlockPasswordDialogVisible.value = true
    return
  }
  openEditDialog(prompt)
}

const unlockAndEdit = async () => {
  if (!unlockingPrompt) return
  const password = unlockPasswordForm.value.password.trim()
  if (!password) {
    ElMessage.warning('请输入密码')
    return
  }
  const isValid = await verifyPassword(password, unlockingPrompt.password || '')
  if (!isValid) {
    ElMessage.error('密码错误')
    return
  }
  const decrypted = await decryptContent(unlockingPrompt.content, password)
  if (decrypted === null) {
    ElMessage.error('解密失败，内容已损坏或密码不正确')
    return
  }
  unlockPasswordDialogVisible.value = false
  openEditDialog({ ...unlockingPrompt, content: decrypted })
}

const openEditDialog = (prompt: Prompt) => {
  isCreating.value = false
  formData.value = {
    id: prompt.id,
    name: prompt.name,
    description: prompt.description || '',
    content: prompt.content,
    category: prompt.category,
    order_num: prompt.order_num,
    card_type: prompt.card_type || 'normal',
    password: prompt.password || null,
    creator_name: prompt.creator_name || '',
    version: prompt.version || ''
  }
  // 字段配置
  fieldsConfig.value = []
  if (prompt.fields && prompt.fields.length > 0) {
    fieldsConfig.value = prompt.fields.map((field) => ({
      name: field.name,
      label: field.label,
      type: field.type,
      options: field.options || [],
      optionLabels: field.optionLabels || [],
      optionsText: (field.options || []).join('\n'),
      optionLabelsText: (field.optionLabels || []).join('\n'),
      description: field.description || '',
      required: field.required !== undefined ? field.required : true
    }))
  } else {
    extractFieldsFromContent(prompt.content)
  }
  // 标签
  formSubcategories.value = [...(prompt.subcategories || [])]
  newSubcategory.value = ''

  editDialogVisible.value = true
}

const submitEdit = async () => {
  if (!formData.value.name.trim()) {
    ElMessage.warning('请输入名称')
    return
  }
  if (!formData.value.category.trim()) {
    ElMessage.warning('请输入卡包分类')
    return
  }
  if (!formData.value.content.trim()) {
    ElMessage.warning('请输入提示词内容')
    return
  }
  // 校验字段
  for (const field of fieldsConfig.value) {
    if (!field.name.trim()) {
      ElMessage.error('所有字段必须有名称')
      return
    }
    if (field.type === 'select' && field.options.length === 0) {
      ElMessage.error(`字段"${field.name}"是下拉类型，必须添加选项`)
      return
    }
  }
  const fieldNames = fieldsConfig.value.map((f) => f.name.trim())
  if (new Set(fieldNames).size !== fieldNames.length) {
    ElMessage.error('字段名称不能重复')
    return
  }

  editSaving.value = true
  try {
    const payload = {
      name: formData.value.name,
      description: formData.value.description,
      content: formData.value.content,
      category: formData.value.category,
      order_num: formData.value.order_num,
      creator_name: formData.value.creator_name,
      version: formData.value.version,
      card_type: formData.value.card_type,
      password: formData.value.password,
      fields: fieldsConfig.value.map((field) => ({
        name: field.name,
        label: field.label,
        type: field.type,
        options: field.options,
        optionLabels: field.optionLabels,
        description: field.description,
        required: field.required
      })),
      subcategories: formSubcategories.value
    }
    
    let res
    if (isCreating.value) {
      res = await promptAPI.create(payload)
    } else {
      res = await promptAPI.update(formData.value.id, payload)
    }
    
    if (res.success) {
      ElMessage.success(isCreating.value ? '创建成功' : '保存成功')
      editDialogVisible.value = false
      await fetchPrompts()
    } else {
      ElMessage.error(res.message || (isCreating.value ? '创建失败' : '保存失败'))
    }
  } catch (e: any) {
    ElMessage.error(e?.message || (isCreating.value ? '创建失败' : '保存失败'))
  } finally {
    editSaving.value = false
  }
}

// 从内容中提取 ${字段名} 占位符
const extractFieldsFromContent = (content: string) => {
  const fieldRegex = /\$\{([^}]+)\}/g
  const fields = new Set<string>()
  let match
  while ((match = fieldRegex.exec(content)) !== null) {
    if (match && match[1]) fields.add(match[1].trim())
  }
  fields.forEach((fieldName) => {
    fieldsConfig.value.push({
      name: fieldName,
      label: fieldName,
      type: 'text',
      options: [],
      optionLabels: [],
      optionsText: '',
      optionLabelsText: '',
      description: '',
      required: true
    })
  })
}

const addField = () => {
  fieldsConfig.value.push({
    name: `字段${fieldsConfig.value.length + 1}`,
    label: `字段${fieldsConfig.value.length + 1}`,
    type: 'text',
    options: [],
    optionLabels: [],
    optionsText: '',
    optionLabelsText: '',
    description: '',
    required: true
  })
}

const removeField = (index: number) => {
  fieldsConfig.value.splice(index, 1)
}

const moveField = (fromIndex: number, toIndex: number) => {
  if (fromIndex < 0 || fromIndex >= fieldsConfig.value.length) return
  if (toIndex < 0 || toIndex >= fieldsConfig.value.length) return
  const [moved] = fieldsConfig.value.splice(fromIndex, 1)
  if (moved) fieldsConfig.value.splice(toIndex, 0, moved)
}

const updateFieldName = (index: number, newName: string) => {
  if (!newName.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  const field = fieldsConfig.value[index]
  if (!field) return
  const oldName = field.name
  const exists = fieldsConfig.value.some((f, i) => i !== index && f.name === newName)
  if (exists) {
    ElMessage.warning('字段名称不能重复')
    return
  }
  if (oldName && oldName !== newName) {
    formData.value.content = formData.value.content.replace(
      new RegExp(`\\$\\{${oldName}\\}`, 'g'),
      `$\{${newName}\}`
    )
  }
}

const copyFieldName = (fieldName: string) => {
  if (!fieldName.trim()) {
    ElMessage.warning('字段名称不能为空')
    return
  }
  const placeholder = `$\{${fieldName}\}`
  navigator.clipboard.writeText(placeholder).then(
    () => ElMessage.success('占位符已复制'),
    () => ElMessage.error('复制失败')
  )
}

const updateFieldOptions = (index: number) => {
  const field = fieldsConfig.value[index]
  if (!field) return
  field.options = field.optionsText
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)
  field.optionLabels = field.optionLabels.slice(0, field.options.length)
  field.optionLabelsText = field.optionLabels.join('\n')
}

const updateFieldOptionLabels = (index: number) => {
  const field = fieldsConfig.value[index]
  if (!field) return
  const labels = field.optionLabelsText
    .split('\n')
    .map((s) => s.trim())
    .slice(0, field.options.length)
  while (labels.length > 0 && !labels[labels.length - 1]) labels.pop()
  field.optionLabels = labels
  field.optionLabelsText = labels.join('\n')
}

const addFormSubcategory = () => {
  const value = newSubcategory.value.trim()
  if (value && !formSubcategories.value.includes(value)) {
    formSubcategories.value.push(value)
    newSubcategory.value = ''
  }
}

const removeFormSubcategory = (index: number) => {
  formSubcategories.value.splice(index, 1)
}

watch(packName, () => {
  // 切换卡包时无需重新请求数据，但清空搜索关键字
  searchKeyword.value = ''
})

onMounted(() => {
  fetchPrompts()
})
</script>

<style scoped>
.pack-detail-container {
  padding: 24px 28px;
  background: #f5f7fa;
  min-height: 100%;
  box-sizing: border-box;
}

.pack-detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}

.pack-detail-title-block {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.back-btn {
  width: 36px;
  height: 36px;
  border: 1px solid #e2e8f0;
  background: #ffffff;
  color: #475569;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 16px;
}

.back-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  color: #1e293b;
}

.pack-detail-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pack-detail-count {
  font-size: 12px;
  font-weight: 500;
  color: #64748b;
  background: #e2e8f0;
  padding: 2px 10px;
  border-radius: 999px;
}

.pack-detail-search {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  justify-content: flex-end;
  min-width: 320px;
}

.pack-search-input {
  width: 280px;
}

/* 排序 Tab */
.pack-sort-tab-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #ffffff;
  border-radius: 14px;
  padding: 8px 14px;
  margin-bottom: 18px;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  border: 1px solid #eef0f3;
}

.pack-sort-tabs {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.pack-sort-tab-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 14px;
  border-radius: 999px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.pack-sort-tab-item:hover {
  background: #f1f5f9;
  color: #1e293b;
}

.pack-sort-tab-item.active {
  background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
  color: #ffffff;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
}

.pack-sort-tab-icon {
  font-size: 14px;
}

/* 卡片网格 */
.pack-prompts-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

@media (max-width: 1280px) {
  .pack-prompts-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 960px) {
  .pack-prompts-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.pack-prompt-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px 18px 14px;
  background: #ffffff;
  border-radius: 14px;
  border: 1px solid #eef0f3;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
  cursor: pointer;
  transition: all 0.22s cubic-bezier(0.4, 0, 0.2, 1);
  min-height: 168px;
}

.pack-prompt-card:hover {
  border-color: #c7d2fe;
  box-shadow: 0 10px 24px rgba(59, 130, 246, 0.12);
  transform: translateY(-2px);
}

.pack-prompt-card.encrypted {
  background: linear-gradient(180deg, #fffbeb 0%, #ffffff 100%);
}

.pack-card-category-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 6px;
  border: 1px solid transparent;
  background: #ffffff;
}

.pack-card-title-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-right: 56px;
}

.pack-card-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 6px;
}

.pack-card-lock {
  color: #d97706;
  font-size: 14px;
}

.pack-card-author-row {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #64748b;
}

.pack-author-avatar {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  color: #ffffff;
  font-size: 12px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.pack-author-name {
  max-width: 90px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #475569;
}

.pack-card-stats {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #f59e0b;
  font-weight: 500;
}

.pack-stat-icon {
  font-size: 12px;
}

.pack-card-date {
  margin-left: auto;
  color: #94a3b8;
  font-size: 12px;
}

.pack-card-body-preview {
  font-size: 13px;
  color: #475569;
  line-height: 1.55;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  min-height: 60px;
}

.pack-card-footer-tags {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: auto;
}

.pack-card-footer-tags--empty {
  min-height: 22px;
}

/* 编辑弹窗布局 */
.pack-edit-dialog :deep(.el-dialog__body) {
  padding: 12px 16px 4px;
}
.pack-edit-layout {
  display: flex;
  gap: 16px;
  height: 70vh;
}
.pack-edit-left {
  flex: 2;
  min-width: 0;
  overflow-y: auto;
  padding-right: 8px;
}
.pack-edit-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  border-left: 1px solid #e5e7eb;
  padding-left: 16px;
}
.pack-edit-right-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}
.pack-edit-right-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}
.pack-edit-right-tip {
  font-size: 11px;
  color: #94a3b8;
  margin-bottom: 8px;
}
.pack-edit-right-tip code {
  background: #eef2f7;
  color: #4f46e5;
  padding: 0 4px;
  border-radius: 3px;
}
.pack-edit-right-body {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
}
.pack-edit-form :deep(.el-form-item) {
  margin-bottom: 12px;
}
.pack-edit-form :deep(.el-form-item__label) {
  font-size: 13px;
  color: #1f2937;
  font-weight: 600;
  padding-bottom: 3px;
  line-height: 1.4;
}
.pack-edit-content-input :deep(.el-textarea__inner) {
  padding: 10px 12px;
  font-size: 13px;
  line-height: 1.6;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
.description-rich-editor {
  width: 100%;
  height: 320px;
}
.description-rich-editor :deep(.split-rich-text-editor) {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #ffffff;
  box-shadow: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}
.description-rich-editor :deep(.split-rich-text-editor:focus-within) {
  border-color: #d1d5db;
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.04),
    0 4px 12px rgba(0, 0, 0, 0.05);
}
.description-rich-editor :deep(.toolbar) {
  background: #f8fafc;
  border-bottom: 1px solid #e5e7eb;
  padding: 8px 12px;
  gap: 4px;
}
.description-rich-editor :deep(.toolbar-button) {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  color: #374151;
  min-width: 28px;
  height: 24px;
  padding: 0 6px;
  border-radius: 6px;
  transition: all 0.15s ease;
}
.description-rich-editor :deep(.toolbar-button:hover:not(:disabled)) {
  border-color: #d1d5db;
  background: #f3f4f6;
  color: #1f2937;
  transform: none;
  box-shadow: none;
}
.description-rich-editor :deep(.toolbar-button:disabled) {
  opacity: 0.5;
}
.description-rich-editor :deep(.toolbar-icon) {
  width: 12px;
  height: 12px;
}
.description-rich-editor :deep(.toolbar-glyph) {
  font-size: 11px;
  font-weight: 500;
}
.description-rich-editor :deep(.editor-surface) {
  padding: 16px 20px;
  color: #1f2937;
  font-size: 14px;
  line-height: 1.7;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}
.description-rich-editor :deep(.editor-surface.is-empty::before) {
  color: #9ca3af;
  font-size: 14px;
}
.pack-edit-card-type {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.pack-edit-card-type :deep(.el-radio-group) {
  display: flex;
  gap: 12px;
}
.pack-edit-card-type :deep(.el-radio) {
  display: flex;
  align-items: center;
  gap: 4px;
}
.pack-edit-card-type-actions {
  display: flex;
  gap: 6px;
  align-items: center;
}
.pack-edit-card-type-status {
  margin-top: 2px;
}
.pack-edit-section {
  margin-top: 10px;
  border: 1px dashed #e2e8f0;
  border-radius: 6px;
  padding: 6px 8px 8px;
  background: #f8fafc;
}
.pack-edit-section-head {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
}
.pack-edit-section-title {
  font-size: 12px;
  font-weight: 600;
  color: #1e293b;
}
.pack-edit-section-tip {
  font-size: 11px;
  color: #94a3b8;
  flex: 1;
}
.pack-edit-section-tip code {
  background: #eef2f7;
  color: #4f46e5;
  padding: 0 3px;
  border-radius: 3px;
  font-size: 11px;
}
.pack-edit-empty {
  font-size: 12px;
  color: #94a3b8;
  text-align: center;
  padding: 20px 0;
}
.pack-edit-fields {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.pack-edit-field {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.pack-edit-field-head {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}
.pack-edit-field-name {
  flex: 1;
  min-width: 80px;
}
.pack-edit-field-name :deep(.el-input__wrapper) {
  padding: 1px 6px;
}
.pack-edit-field-body {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.pack-edit-field-body :deep(.el-input__wrapper),
.pack-edit-field-body :deep(.el-select__wrapper) {
  padding: 1px 6px;
}
.pack-edit-field-order {
  display: flex;
  gap: 4px;
}
.pack-edit-field-options {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 4px;
}
.pack-edit-field-options :deep(.el-textarea__inner) {
  padding: 4px 6px;
  font-size: 12px;
  min-height: 44px;
}
.pack-edit-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.pack-edit-tag-input {
  width: 150px;
}
.pack-edit-tag-input :deep(.el-input__wrapper) {
  padding: 1px 6px;
}

.pack-subcategory-tag {
  font-size: 11px;
  font-weight: 500;
  padding: 2px 8px;
  border-radius: 6px;
  border: 1px solid transparent;
  white-space: nowrap;
}

.pack-more-tags {
  font-size: 11px;
  color: #94a3b8;
}

.pack-empty-hint {
  background: #ffffff;
  border-radius: 14px;
  padding: 48px 0;
  border: 1px dashed #cbd5e1;
}

/* 暗色主题 */
:root[data-theme='dark'] .pack-detail-container {
  background: #0f172a;
}

:root[data-theme='dark'] .back-btn {
  background: #1e293b;
  border-color: #334155;
  color: #cbd5e1;
}

:root[data-theme='dark'] .back-btn:hover {
  background: #334155;
  border-color: #475569;
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-detail-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-detail-count {
  background: #334155;
  color: #cbd5e1;
}

:root[data-theme='dark'] .pack-sort-tab-bar {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .pack-sort-tab-item {
  color: #cbd5e1;
}

:root[data-theme='dark'] .pack-sort-tab-item:hover {
  background: #334155;
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-prompt-card {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .pack-prompt-card:hover {
  border-color: #818cf8;
  box-shadow: 0 10px 24px rgba(99, 102, 241, 0.2);
}

:root[data-theme='dark'] .pack-prompt-card.encrypted {
  background: linear-gradient(180deg, rgba(180, 83, 9, 0.15) 0%, #1e293b 100%);
}

:root[data-theme='dark'] .pack-card-category-badge {
  background: #1e293b;
}

:root[data-theme='dark'] .pack-card-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .pack-author-name {
  color: #cbd5e1;
}

:root[data-theme='dark'] .pack-card-date {
  color: #64748b;
}

:root[data-theme='dark'] .pack-card-body-preview {
  color: #94a3b8;
}

:root[data-theme='dark'] .pack-empty-hint {
  background: #1e293b;
  border-color: #334155;
}
</style>
