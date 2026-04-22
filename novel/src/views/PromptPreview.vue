<template>
  <div class="prompt-preview-container">
    <div class="header">
      <h2>提示词预览</h2>
      <div class="search-box">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索提示词"
          clearable
          prefix-icon="Search"
          class="search-input"
        />
      </div>
    </div>

    <!-- 顶部长方体模块 - 卡包分类选择 -->
  <div class="category-selector">
    <span class="category-title">分类：</span>
    <div class="category-buttons-wrapper">
      <div class="category-buttons">
        <el-button
          v-for="category in ['all', ...categoryList.map(c => c.name)]"
          :key="category"
          :type="selectedCategory === category ? 'primary' : ''"
          size="small"
          @click="selectedCategory = category"
        >
          {{ category === 'all' ? '全部' : category }}
        </el-button>
      </div>
    </div>
  </div>

    <!-- 提示词列表 -->
    <div class="prompts-list">
      <div v-if="filteredPrompts.length === 0" class="empty-hint">
        <el-empty description="暂无提示词" />
      </div>
      <div v-else class="prompts-grid">
        <div
          v-for="prompt in filteredPrompts"
          :key="prompt.id"
          class="prompt-card"
          @click="handlePreview(prompt)"
        >
          <div class="prompt-card-section prompt-card-name" :title="prompt.name">
            <span class="section-label">提示词名称</span>
            <h3 class="section-content">{{ prompt.name }}</h3>
          </div>
          <div class="prompt-card-section prompt-card-date" :title="formatDate(prompt.created_at)">
            <span class="section-label">创建日期</span>
            <span class="section-content">{{ formatDate(prompt.created_at) }}</span>
          </div>
          <div class="prompt-card-section prompt-card-desc" :title="getPromptDescriptionPreview(prompt)">
            <span class="section-label">简介内容</span>
            <div class="section-content">{{ getPromptDescriptionPreview(prompt) || '暂无简介' }}</div>
          </div>
          <div class="prompt-card-section prompt-card-tags">
            <span class="section-label">标签内容</span>
            <div class="section-content tags-wrapper">
              <el-tag
                v-for="(subcat, index) in (prompt.subcategories || []).slice(0, 3)"
                :key="index"
                size="small"
                class="subcategory-tag"
              >
                {{ subcat }}
              </el-tag>
              <span v-if="(prompt.subcategories || []).length > 3" class="more-tags">+{{ (prompt.subcategories || []).length - 3 }}</span>
              <el-button
                v-if="(prompt.subcategories || []).length === 0"
                type="primary"
                link
                size="small"
                @click.stop="openSubcategoryEditor(prompt)"
                class="add-subcategory-btn"
              >
                <el-icon><Plus /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      width="56vw"
      :show-close="false"
      destroy-on-close
      class="preview-dialog"
    >
      <div class="preview-dialog-content">
        <div class="preview-dialog-header">
          <div class="preview-dialog-actions">
            <template v-if="selectedPreviewPrompt?.card_type === 'encrypted' && selectedPreviewPrompt?.password">
              <el-popover
                :visible="editPasswordPopoverVisible"
                placement="bottom"
                :width="280"
                trigger="click"
              >
                <template #reference>
                  <el-button type="primary" @click="editPasswordPopoverVisible = true">
                    <el-icon><Edit /></el-icon>
                    编辑提示词
                  </el-button>
                </template>
                <div class="edit-password-popover">
                  <p class="popover-hint">请输入密码编辑</p>
                  <el-input
                    v-model="editPasswordInput"
                    type="password"
                    placeholder="请输入密码"
                    show-password
                    @keyup.enter="verifyEditPassword"
                  />
                  <div class="popover-actions">
                    <el-button size="small" @click="editPasswordPopoverVisible = false">取消</el-button>
                    <el-button type="primary" size="small" @click="verifyEditPassword">确认</el-button>
                  </div>
                </div>
              </el-popover>
            </template>
            <template v-else>
              <el-button type="primary" @click="editPreviewPrompt">
                <el-icon><Edit /></el-icon>
                编辑提示词
              </el-button>
            </template>
            <el-button type="primary" link @click="copyPreviewContent">
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
            <el-button link @click="previewDialogVisible = false">
              <el-icon><Close /></el-icon>
              关闭
            </el-button>
          </div>
        </div>
        <el-scrollbar class="preview-scrollbar">
          <div v-if="selectedPreviewPrompt" class="preview-body">
            <div class="preview-meta-card">
              <h2>{{ selectedPreviewPrompt.name }}</h2>
              <span>{{ formatDate(selectedPreviewPrompt.created_at) }}</span>
            </div>

            <div v-if="selectedPreviewPrompt.description?.trim()" class="preview-section">
              <div class="preview-section-label">简介</div>
              <div class="preview-markdown-card">
                <MarkdownRenderer :content="selectedPreviewPrompt.description || ''" />
              </div>
            </div>

            <div class="preview-section preview-section-collapse">
              <template v-if="selectedPreviewPrompt.card_type === 'encrypted' && selectedPreviewPrompt.password">
                <div class="preview-encrypted-notice">
                  <el-icon :size="32" color="#e6a23c"><Lock /></el-icon>
                  <p class="encrypted-title">私人提示词</p>
                </div>
              </template>
              <template v-else>
                <el-collapse>
                  <el-collapse-item name="content">
                    <template #title>
                      <div class="preview-collapse-title">
                        <span>提示词内容</span>
                        <small>默认折叠，点击展开</small>
                      </div>
                    </template>
                    <div class="preview-markdown-card preview-content-card">
                      <MarkdownRenderer :content="selectedPreviewPrompt.content" />
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </template>
            </div>
          </div>
        </el-scrollbar>
      </div>
    </el-dialog>

    <!-- 小分类编辑弹窗 -->
  <el-dialog
    v-model="subcategoryDialogVisible"
    title=""
    width="380px"
    destroy-on-close
    class="subcategory-dialog"
  >
    <div class="subcategory-editor">
      <div class="subcategory-header">
        <span class="subcategory-title">编辑标签</span>
        <span class="subcategory-count">{{ editingSubcategories.length }} 个标签</span>
      </div>
      <div class="subcategory-list">
        <transition-group name="tag-list">
          <div
            v-for="(_, index) in editingSubcategories"
            :key="index"
            class="subcategory-item"
          >
            <el-input
              v-model="editingSubcategories[index]"
              size="small"
              placeholder="输入标签名称"
              class="tag-input"
            />
            <el-button
              type="danger"
              link
              size="small"
              @click="removeSubcategory(index)"
              class="delete-btn"
            >
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </transition-group>
        <div v-if="editingSubcategories.length === 0" class="empty-tags">
          暂无标签，点击下方添加
        </div>
      </div>
      <el-button
        type="primary"
        size="small"
        @click="addSubcategory"
        class="add-btn"
        plain
      >
        <el-icon><Plus /></el-icon>
        添加标签
      </el-button>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="subcategoryDialogVisible = false" class="cancel-btn">取消</el-button>
        <el-button type="primary" @click="saveSubcategories" class="save-btn">保存</el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 编辑提示词弹窗 -->
  <el-dialog
    v-model="editDialogVisible"
    title="编辑提示词"
    :width="editDialogWidthMap[editDialogSize]"
    top="6vh"
    append-to-body
    :close-on-click-modal="false"
    class="edit-prompt-dialog"
  >
    <template #header>
      <div class="edit-dialog-header">
        <span>编辑提示词</span>
        <div class="size-controls">
          <el-button-group>
            <el-button 
              :type="editDialogSize === 'small' ? 'primary' : ''" 
              size="small"
              @click="editDialogSize = 'small'"
            >
              小
            </el-button>
            <el-button 
              :type="editDialogSize === 'medium' ? 'primary' : ''" 
              size="small"
              @click="editDialogSize = 'medium'"
            >
              中
            </el-button>
            <el-button 
              :type="editDialogSize === 'large' ? 'primary' : ''" 
              size="small"
              @click="editDialogSize = 'large'"
            >
              大
            </el-button>
          </el-button-group>
        </div>
      </div>
    </template>
    <div class="edit-dialog-content" :style="{ height: editDialogHeightMap[editDialogSize] }">
      <div class="edit-left-panel">
        <el-form :model="editFormData" label-width="80px">
          <el-form-item label="名称" required>
            <el-input v-model="editFormData.name" placeholder="请输入名称" />
          </el-form-item>
          <el-form-item label="分类名称" required>
            <el-select
              v-model="editFormData.category"
              placeholder="请选择卡包"
              class="field-input full-width-input"
            >
              <el-option
                v-for="category in editCategoryOptions"
                :key="category"
                :label="category"
                :value="category"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="简介">
            <SplitRichTextEditor
              v-model="editFormData.description"
              placeholder="用于介绍这个提示词，非必填"
              class="description-rich-editor"
            />
          </el-form-item>
          <el-form-item required>
            <template #label>
              <div class="form-item-label-with-guide">
                <span>内容</span>
                <el-popover
                  placement="right-start"
                  :width="360"
                  trigger="click"
                  popper-class="prompt-guide-popper"
                >
                  <template #reference>
                    <el-button size="small" text class="prompt-guide-trigger">
                      <el-icon><InfoFilled /></el-icon>
                      模板说明
                    </el-button>
                  </template>
                  <div class="prompt-guide-popover">
                    <div class="prompt-guide-title">把说明写进模板，把变量留给字段</div>
                    <p class="prompt-guide-intro">
                      正文里直接写完整提示词，把需要用户补充的内容写成 <code>${字段名称}</code>。
                    </p>
                    <div class="prompt-guide-example">
                      <span class="prompt-guide-example-label">示例</span>
                      <p>帮我生成 <code>${数量}</code> 个 <code>${小说类型}</code> 类型、适合 <code>${平台}</code> 平台的小说书名。</p>
                    </div>
                    <div class="prompt-guide-tips">
                      <div class="prompt-guide-tip">
                        <span class="prompt-guide-tip-index">1</span>
                        <span>先按正常语气写完整句子，再把可变内容替换成字段。</span>
                      </div>
                      <div class="prompt-guide-tip">
                        <span class="prompt-guide-tip-index">2</span>
                        <span>长文本内容建议单独成段，避免和短字段混在一句里。</span>
                      </div>
                      <div class="prompt-guide-tip">
                        <span class="prompt-guide-tip-index">3</span>
                        <span>字段名称尽量直接，比如"数量""平台""风格要求"。</span>
                      </div>
                    </div>
                  </div>
                </el-popover>
              </div>
            </template>
            <el-input
              v-model="editFormData.content"
              type="textarea"
              :rows="15"
              placeholder="请输入提示词内容（作为AI的system层指令）"
            />
          </el-form-item>
          <div class="panel-header edit-panel-header">
            <label class="form-label">字段配置</label>
            <el-button type="primary" size="small" @click="addEditField">
              <el-icon><Plus /></el-icon>
              添加字段
            </el-button>
          </div>
          <div class="field-config-section">
            <el-alert
              title="字段配置说明"
              type="info"
              :closable="false"
              show-icon
            >
              <template #default>
                <div class="format-description">
                  <p><strong>⚙️ 配置字段信息</strong></p>
                  <p>为每个 <code>${字段名称}</code> 设置详细信息，决定用户看到什么样的表单</p>
                </div>
              </template>
            </el-alert>
            
            <div v-if="editFieldsConfig.length === 0" class="empty-fields">
              <p class="empty-hint">暂无字段配置，点击右上角"添加字段"开始配置</p>
            </div>
            
            <div v-else class="fields-list">
              <div v-for="(field, index) in editFieldsConfig" :key="index" class="field-item">
                <div class="field-header">
                  <div class="field-name-container">
                    <el-tag 
                      :type="field.required ? 'danger' : 'info'" 
                      size="small"
                      class="required-tag"
                    >
                      {{ field.required ? '必选' : '选填' }}
                    </el-tag>
                    <el-input
                      v-model="field.name"
                      class="field-name-input"
                      placeholder="字段名称"
                      @change="updateEditFieldName(index, field.name)"
                    />
                    <el-button
                      type="primary"
                      link
                      size="small"
                      @click="copyFieldName(field.name)"
                      title="复制占位符"
                    >
                      <el-icon><CopyDocument /></el-icon>
                    </el-button>
                  </div>
                  <div class="field-actions">
                    <el-switch
                      v-model="field.required"
                      active-text="必选"
                      inactive-text="选填"
                      :active-action-icon="Check"
                      inline-prompt
                      style="--el-switch-on-color: #f56c6c; --el-switch-off-color: #909399;"
                    />
                    <el-button 
                      type="primary" 
                      link 
                      size="small" 
                      @click="moveEditField(index, index - 1)"
                      :disabled="index === 0"
                      title="上移"
                    >
                      <el-icon><ArrowUp /></el-icon>
                    </el-button>
                    <el-button 
                      type="primary" 
                      link 
                      size="small" 
                      @click="moveEditField(index, index + 1)"
                      :disabled="index === editFieldsConfig.length - 1"
                      title="下移"
                    >
                      <el-icon><ArrowDown /></el-icon>
                    </el-button>
                    <el-button type="danger" link size="small" @click="removeEditField(index)" title="删除">
                      <el-icon><Delete /></el-icon>
                    </el-button>
                  </div>
                </div>
                <div class="field-config-form">
                  <el-input
                    v-model="field.label"
                    placeholder="字段显示名称"
                    class="field-input"
                  />
                  <el-select
                    v-model="field.type"
                    placeholder="字段类型"
                    class="field-input"
                  >
                    <el-option label="单行文本" value="text" />
                    <el-option label="多行文本" value="textarea" />
                    <el-option label="下拉选择" value="select" />
                  </el-select>
                  <div v-if="field.type === 'select'" class="field-options">
                    <label>选项（每行一个）：</label>
                    <el-input
                      v-model="field.optionsText"
                      type="textarea"
                      :rows="3"
                      placeholder="请输入选项，每行一个"
                      @change="updateEditFieldOptions(index)"
                      class="field-input"
                    />
                  </div>
                  <el-input
                    v-model="field.description"
                    type="textarea"
                    :rows="2"
                    placeholder="字段说明"
                    class="field-input"
                  />
                </div>
              </div>
            </div>
          </div>
        </el-form>
      </div>
    </div>
    <template #footer>
      <div class="edit-dialog-footer">
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSubmit">保存</el-button>
      </div>
    </template>
  </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { CopyDocument, Edit, Plus, Delete, Close, ArrowUp, ArrowDown, Check, InfoFilled, Lock } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

// 提示词列表
const prompts = ref<Prompt[]>([])

// 选中的分类
const selectedCategory = ref('all')

// 搜索关键词
const searchKeyword = ref('')

// 预览弹窗相关状态
const previewDialogVisible = ref(false)
const selectedPreviewPrompt = ref<Prompt | null>(null)
const editPasswordPopoverVisible = ref(false)
const editPasswordInput = ref('')

// 密码哈希函数
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

const verifyEditPassword = async () => {
  if (!selectedPreviewPrompt.value) return
  const isValid = await verifyPassword(editPasswordInput.value, selectedPreviewPrompt.value.password || '')
  if (isValid) {
    editPasswordPopoverVisible.value = false
    editPasswordInput.value = ''
    editPreviewPrompt()
  } else {
    ElMessage.error('密码错误')
  }
}

// 小分类编辑弹窗相关状态
const subcategoryDialogVisible = ref(false)
const editingPrompt = ref<Prompt | null>(null)
const editingSubcategories = ref<string[]>([])

// 编辑弹窗相关状态
const editDialogVisible = ref(false)
const editDialogSize = ref<'small' | 'medium' | 'large'>('medium')
const editDialogWidthMap = {
  small: '900px',
  medium: '1200px',
  large: '1500px'
}
const editDialogHeightMap = {
  small: '500px',
  medium: '600px',
  large: '700px'
}
const editFormData = ref({
  id: 0,
  name: '',
  description: '',
  content: '',
  category: '默认',
  order_num: 0
})
const editFieldsConfig = ref<Array<{
  name: string;
  label: string;
  type: 'text' | 'textarea' | 'select';
  options: string[];
  optionsText: string;
  description: string;
  required: boolean;
}>>([])
const editFormSubcategories = ref<string[]>([])
const DEFAULT_CATEGORIES = ['默认', '写作要求', '写作风格']

const editCategoryOptions = computed(() => {
  const names = new Set<string>(DEFAULT_CATEGORIES)
  if (editFormData.value.category) {
    names.add(editFormData.value.category)
  }
  customCategories.value.forEach(name => {
    if (name && name !== '未分类') {
      names.add(name)
    }
  })
  categoryList.value.forEach(category => {
    if (category.name && category.name !== '未分类') {
      names.add(category.name)
    }
  })
  return Array.from(names)
})

const getPromptDescriptionPreview = (prompt: Prompt) => {
  const description = stripHtml(prompt.description || '').trim()
  if (!description) return ''
  return `${description.slice(0, 100)}${description.length > 100 ? '...' : ''}`
}

const stripHtml = (value: string) => {
  if (!value) return ''
  if (typeof window === 'undefined') {
    return value.replace(/<[^>]+>/g, ' ')
  }
  const temp = document.createElement('div')
  temp.innerHTML = value
  return temp.textContent || temp.innerText || ''
}

// 打开小分类编辑器
const openSubcategoryEditor = (prompt: Prompt) => {
  editingPrompt.value = prompt
  editingSubcategories.value = [...(prompt.subcategories || [])]
  subcategoryDialogVisible.value = true
}

// 添加小分类
const addSubcategory = () => {
  editingSubcategories.value.push('')
}

// 删除小分类
const removeSubcategory = (index: number) => {
  editingSubcategories.value.splice(index, 1)
}

// 保存小分类
const saveSubcategories = async () => {
  if (!editingPrompt.value) return
  
  // 过滤掉空的小分类
  const filteredSubcategories = editingSubcategories.value.filter(subcat => subcat.trim())
  
  try {
    await promptAPI.update(editingPrompt.value.id, {
      ...editingPrompt.value,
      subcategories: filteredSubcategories
    })
    
    // 更新本地数据
    const promptIndex = prompts.value.findIndex(p => p.id === editingPrompt.value!.id)
    if (promptIndex !== -1) {
      const existingPrompt = prompts.value[promptIndex]
      if (existingPrompt) {
        prompts.value[promptIndex] = {
          ...existingPrompt,
          subcategories: filteredSubcategories
        }
      }
    }
    
    ElMessage.success('标签保存成功')
    subcategoryDialogVisible.value = false
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 从localStorage加载卡包列表
const loadCategoriesFromStorage = (): string[] => {
  const stored = localStorage.getItem('prompt_categories')
  return stored ? JSON.parse(stored) : []
}

// 卡包列表（从localStorage加载）
const customCategories = ref<string[]>(loadCategoriesFromStorage())

// 按分类分组的提示词列表
const categoryList = computed(() => {
  const groups: Record<string, Prompt[]> = {}
  
  // 先确保所有自定义卡包都存在
  customCategories.value.forEach(cat => {
    if (!groups[cat]) {
      groups[cat] = []
    }
  })
  
  // 添加提示词到对应分组
  prompts.value.forEach(prompt => {
    const category = prompt.category || '默认'
    if (!groups[category]) {
      groups[category] = []
    }
    groups[category].push(prompt)
  })
  
  // 确保默认分类存在
  if (!groups['默认']) {
    groups['默认'] = []
  }
  
  return Object.entries(groups).map(([name, prompts]) => ({
    name,
    prompts
  }))
})

// 过滤后的提示词列表
const filteredPrompts = computed(() => {
  let result = prompts.value
  
  // 按分类过滤
  if (selectedCategory.value !== 'all') {
    result = result.filter(prompt => prompt.category === selectedCategory.value)
  }
  
  // 按搜索关键词过滤
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    result = result.filter(prompt => 
      prompt.name.toLowerCase().includes(keyword) || 
      prompt.content.toLowerCase().includes(keyword) ||
      (prompt.description || '').toLowerCase().includes(keyword)
    )
  }
  
  return result
})

// 获取标签类型
const getTagType = (category: string): any => {
  const typeMap: Record<string, any> = {
    '角色': 'success',
    '剧情': 'warning',
    '对话': 'primary',
    '场景': 'info'
  }
  return typeMap[category] || 'info'
}

// 格式化日期
const formatDate = (dateStr: string | undefined) => {
  if (!dateStr) return '未知时间'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  })
}

// 加载提示词
const fetchPrompts = async () => {
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
}

// 预览提示词
const handlePreview = (prompt: Prompt) => {
  selectedPreviewPrompt.value = prompt
  previewDialogVisible.value = true
}

const editPreviewPrompt = () => {
  if (!selectedPreviewPrompt.value) return

  const prompt = selectedPreviewPrompt.value
  editFormData.value = {
    id: prompt.id,
    name: prompt.name,
    description: prompt.description || '',
    content: prompt.content,
    category: prompt.category,
    order_num: prompt.order_num
  }
  
  editFieldsConfig.value = []
  
  if (prompt.fields && prompt.fields.length > 0) {
    editFieldsConfig.value = prompt.fields.map(field => ({
      name: field.name,
      label: field.label,
      type: field.type,
      options: field.options || [],
      optionsText: (field.options || []).join('\n'),
      description: field.description || '',
      required: field.required !== undefined ? field.required : true
    }))
  } else {
    const fieldNames = extractFieldNames(prompt.content)
    editFieldsConfig.value = fieldNames.map(name => ({
      name,
      label: name,
      type: 'text' as const,
      options: [],
      optionsText: '',
      description: '',
      required: true
    }))
  }
  
  editFormSubcategories.value = [...(prompt.subcategories || [])]
  previewDialogVisible.value = false
  editDialogVisible.value = true
}

const extractFieldNames = (content: string): string[] => {
  const regex = /\$\{([^}]+)\}/g
  const names: string[] = []
  let match
  while ((match = regex.exec(content)) !== null) {
    if (!names.includes(match[1])) {
      names.push(match[1])
    }
  }
  return names
}

const addEditField = () => {
  editFieldsConfig.value.push({
    name: '',
    label: '',
    type: 'text',
    options: [],
    optionsText: '',
    description: '',
    required: true
  })
}

const removeEditField = (index: number) => {
  editFieldsConfig.value.splice(index, 1)
}

const moveEditField = (fromIndex: number, toIndex: number) => {
  if (toIndex < 0 || toIndex >= editFieldsConfig.value.length) return
  const field = editFieldsConfig.value[fromIndex]
  editFieldsConfig.value.splice(fromIndex, 1)
  editFieldsConfig.value.splice(toIndex, 0, field)
}

const updateEditFieldOptions = (index: number) => {
  const field = editFieldsConfig.value[index]
  if (field) {
    field.options = field.optionsText.split('\n').map(opt => opt.trim()).filter(opt => opt)
  }
}

const updateEditFieldName = (index: number, newName: string) => {
  const oldName = editFieldsConfig.value[index]?.name
  if (oldName && newName && oldName !== newName) {
    editFormData.value.content = editFormData.value.content.replace(
      new RegExp(`\\$\\{${oldName}\\}`, 'g'),
      `\${${newName}}`
    )
  }
}

const copyFieldName = (name: string) => {
  navigator.clipboard.writeText(`\${${name}}`)
  ElMessage.success('已复制字段占位符')
}

const validateEditFieldsConfig = (): string | null => {
  for (const field of editFieldsConfig.value) {
    if (!field.name.trim()) {
      return '字段名称不能为空'
    }
    if (!field.label.trim()) {
      return `字段"${field.name}"的显示名称不能为空`
    }
    if (field.type === 'select' && field.options.length === 0) {
      return `字段"${field.name}"是下拉选择类型，请至少添加一个选项`
    }
  }
  return null
}

const handleEditSubmit = async () => {
  if (!editFormData.value.name || !editFormData.value.category || !editFormData.value.content) {
    ElMessage.warning('请填写完整信息')
    return
  }

  const validationError = validateEditFieldsConfig()
  if (validationError) {
    ElMessage.error(validationError)
    return
  }

  try {
    const dataToSave = {
      ...editFormData.value,
      fields: editFieldsConfig.value.map(field => ({
        name: field.name,
        label: field.label,
        type: field.type,
        options: field.options,
        description: field.description,
        required: field.required
      })),
      subcategories: editFormSubcategories.value
    }
    
    const res = await promptAPI.update(editFormData.value.id, dataToSave)
    if (res.success) {
      const promptIndex = prompts.value.findIndex(p => p.id === editFormData.value.id)
      if (promptIndex !== -1) {
        prompts.value[promptIndex] = {
          ...prompts.value[promptIndex],
          ...dataToSave
        } as Prompt
      }
      ElMessage.success('保存成功')
      editDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

// 复制预览内容到剪贴板
const copyPreviewContent = async () => {
  try {
    if (!selectedPreviewPrompt.value) return
    const description = selectedPreviewPrompt.value.description?.trim()
    const textToCopy = [
      selectedPreviewPrompt.value.name,
      formatDate(selectedPreviewPrompt.value.created_at),
      description ? `简介\n${stripHtml(description)}` : '',
      `提示词内容\n${selectedPreviewPrompt.value.content}`
    ].filter(Boolean).join('\n\n')
    await navigator.clipboard.writeText(textToCopy)
    ElMessage.success('内容已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

onMounted(async () => {
  await fetchPrompts()
})
</script>

<style scoped>
.prompt-preview-container {
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.08) 0%, rgba(5, 150, 145, 0.12) 100%);
  border: 1px solid rgba(8, 198, 190, 0.2);
  border-radius: 16px;
  padding: 20px 24px;
  box-shadow: 0 4px 20px rgba(8, 198, 190, 0.1);
  position: relative;
  overflow: hidden;
}

.header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, #08c6be, #059691, #08c6be);
  border-radius: 16px 16px 0 0;
}

.header::after {
  content: '';
  position: absolute;
  top: 15px;
  right: 150px;
  width: 100px;
  height: 100px;
  background: radial-gradient(circle, rgba(8, 198, 190, 0.1) 0%, transparent 70%);
  border-radius: 50%;
  pointer-events: none;
}

.header h2 {
  font-size: 24px;
  font-weight: 600;
  color: #059691;
  margin: 0;
  text-shadow: 0 1px 3px rgba(8, 198, 190, 0.3);
  position: relative;
  z-index: 1;
}

.search-box {
  width: 300px;
  position: relative;
  z-index: 1;
}

.search-input {
  width: 100%;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.8);
  border: 1px solid rgba(8, 198, 190, 0.25);
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.1);
  transition: all 0.3s;
}

.search-input :deep(.el-input__wrapper:hover),
.search-input :deep(.el-input__wrapper.is-focus) {
  border-color: rgba(8, 198, 190, 0.5);
  box-shadow: 0 4px 16px rgba(8, 198, 190, 0.2);
}

.search-input :deep(.el-input__inner) {
  color: #059691;
}

.search-input :deep(.el-input__inner::placeholder) {
  color: rgba(5, 150, 145, 0.5);
}

.category-selector {
 margin-bottom: 20px;
 background: linear-gradient(135deg, rgba(8, 198, 190, 0.08) 0%, rgba(5, 150, 145, 0.12) 100%);
 border: 1px solid rgba(8, 198, 190, 0.2);
 padding: 12px 20px;
 border-radius: 16px;
 box-shadow: 0 4px 20px rgba(8, 198, 190, 0.1);
 position: relative;
 overflow: hidden;
 display: flex;
 align-items: center;
 gap: 12px;
 }
 .category-selector::before {
 content: '';
 position: absolute;
 top: 0;
 left: 0;
 right: 0;
 height: 3px;
 background: linear-gradient(90deg, #08c6be, #059691, #08c6be);
 border-radius: 16px 16px 0 0;
 }
 .category-title {
 font-weight: 600;
 color: #059691;
 white-space: nowrap;
 text-shadow: 0 1px 2px rgba(8, 198, 190, 0.2);
 flex-shrink: 0;
 }
 .category-buttons-wrapper {
 flex: 1;
 overflow-x: auto;
 overflow-y: hidden;
 scrollbar-width: thin;
 scrollbar-color: rgba(8, 198, 190, 0.3) transparent;
 }
 .category-buttons-wrapper::-webkit-scrollbar {
 height: 6px;
 }
 .category-buttons-wrapper::-webkit-scrollbar-track {
 background: transparent;
 }
 .category-buttons-wrapper::-webkit-scrollbar-thumb {
 background: rgba(8, 198, 190, 0.3);
 border-radius: 3px;
 }
 .category-buttons-wrapper::-webkit-scrollbar-thumb:hover {
 background: rgba(8, 198, 190, 0.5);
 }
 .category-buttons {
 display: flex;
 gap: 8px;
 flex-wrap: nowrap;
 }

.category-buttons .el-button {
  border-radius: 16px;
  font-size: 12px;
  padding: 4px 12px;
  min-width: auto;
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.1) 0%, rgba(5, 150, 145, 0.15) 100%);
  border: 1px solid rgba(8, 198, 190, 0.25);
  color: #059691;
  transition: all 0.3s;
}

.category-buttons .el-button:hover {
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.2) 0%, rgba(5, 150, 145, 0.25) 100%);
  border-color: rgba(8, 198, 190, 0.5);
  color: #08c6be;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(8, 198, 190, 0.2);
}

.category-buttons .el-button--primary {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border-color: #08c6be;
  color: #fff;
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.3);
}

.category-buttons .el-button--primary:hover {
  background: linear-gradient(135deg, #0dd9d1 0%, #08c6be 100%);
  border-color: #0dd9d1;
  box-shadow: 0 4px 16px rgba(8, 198, 190, 0.4);
}

.prompts-list {
  min-height: 400px;
}

.prompts-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 16px;
  width: 100%;
}

@media (max-width: 1400px) {
  .prompts-grid {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 1100px) {
  .prompts-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 800px) {
  .prompts-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.prompt-card {
 background: linear-gradient(135deg, rgba(8, 198, 190, 0.08) 0%, rgba(5, 150, 145, 0.15) 100%);
 border: 1px solid rgba(8, 198, 190, 0.25);
 border-radius: 16px;
 box-shadow: 0 4px 20px rgba(8, 198, 190, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.1);
 padding: 0;
 transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
 height: 280px;
 display: flex;
 flex-direction: column;
 position: relative;
 overflow: hidden;
 cursor: pointer;
 }
 .prompt-card::before {
 content: '';
 position: absolute;
 top: 0;
 left: 0;
 right: 0;
 height: 3px;
 background: linear-gradient(90deg, #08c6be, #059691, #08c6be);
 border-radius: 16px 16px 0 0;
 }
 .prompt-card::after {
 content: '';
 position: absolute;
 top: 10px;
 right: 10px;
 width: 60px;
 height: 60px;
 background: radial-gradient(circle, rgba(8, 198, 190, 0.15) 0%, transparent 70%);
 border-radius: 50%;
 pointer-events: none;
 }
 .prompt-card:hover {
 box-shadow: 0 8px 32px rgba(8, 198, 190, 0.25), inset 0 1px 0 rgba(255, 255, 255, 0.15);
 transform: translateY(-4px);
 border-color: rgba(8, 198, 190, 0.5);
 }
 .prompt-card:hover::after {
 background: radial-gradient(circle, rgba(8, 198, 190, 0.3) 0%, transparent 70%);
 }
 .prompt-card-section {
 display: flex;
 flex-direction: column;
 justify-content: center;
 padding: 10px 16px;
 position: relative;
 }
 .prompt-card-section:not(:last-child)::after {
 content: '';
 position: absolute;
 bottom: 0;
 left: 50%;
 transform: translateX(-50%);
 width: 60%;
 height: 1px;
 background: linear-gradient(90deg, transparent, rgba(8, 198, 190, 0.12), transparent);
 }
 .prompt-card-name {
 flex: 1;
 min-height: 0;
 }
 .prompt-card-date {
 flex: 1;
 min-height: 0;
 }
 .prompt-card-desc {
 flex: 2;
 min-height: 0;
 }
 .prompt-card-tags {
 flex: 1;
 min-height: 0;
 }
 .section-label {
 display: none;
 }
 .section-content {
 font-size: 14px;
 color: #059691;
 line-height: 1.4;
 overflow: hidden;
 text-overflow: ellipsis;
 white-space: nowrap;
 }
 .prompt-card-name .section-content {
 font-size: 15px;
 font-weight: 600;
 margin: 0;
 }
 .prompt-card-desc .section-content {
 white-space: normal;
 display: -webkit-box;
 -webkit-line-clamp: 3;
 -webkit-box-orient: vertical;
 word-break: break-word;
 }
 .tags-wrapper {
 display: flex;
 flex-wrap: wrap;
 gap: 6px;
 align-items: center;
 }
 .subcategory-tag {
 font-size: 11px;
 border-radius: 12px;
 margin: 0;
 background: linear-gradient(135deg, rgba(8, 198, 190, 0.15) 0%, rgba(5, 150, 145, 0.2) 100%);
 border-color: rgba(8, 198, 190, 0.3);
 color: #059691;
 transition: all 0.2s;
 max-width: 80px;
 overflow: hidden;
 text-overflow: ellipsis;
 white-space: nowrap;
 }
 .subcategory-tag:hover {
 background: linear-gradient(135deg, rgba(8, 198, 190, 0.25) 0%, rgba(5, 150, 145, 0.3) 100%);
 border-color: rgba(8, 198, 190, 0.5);
 transform: scale(1.05);
 }
 .more-tags {
 font-size: 11px;
 color: #5a8a89;
 padding: 2px 6px;
 background: rgba(8, 198, 190, 0.1);
 border-radius: 10px;
 }
 .add-subcategory-btn {
 font-size: 11px;
 padding: 0;
 color: #059691;
 }
 .add-subcategory-btn:hover {
 color: #08c6be;
 text-shadow: 0 0 6px rgba(8, 198, 190, 0.5);
 }

.subcategory-dialog :deep(.el-dialog__header) {
 display: none;
 }
 .subcategory-dialog :deep(.el-dialog__body) {
 padding: 24px 28px 16px;
 }
 .subcategory-dialog :deep(.el-dialog__footer) {
 padding: 0 28px 20px;
 }
 .subcategory-editor {
 display: flex;
 flex-direction: column;
 gap: 16px;
 }
 .subcategory-header {
 display: flex;
 justify-content: space-between;
 align-items: center;
 padding-bottom: 12px;
 border-bottom: 1px solid rgba(8, 198, 190, 0.1);
 }
 .subcategory-title {
 font-size: 16px;
 font-weight: 600;
 color: #059691;
 }
 .subcategory-count {
 font-size: 12px;
 color: #8ba6a5;
 background: rgba(8, 198, 190, 0.08);
 padding: 4px 10px;
 border-radius: 12px;
 }
 .subcategory-list {
 display: flex;
 flex-direction: column;
 gap: 10px;
 max-height: 280px;
 overflow-y: auto;
 padding-right: 4px;
 }
 .subcategory-list::-webkit-scrollbar {
 width: 4px;
 }
 .subcategory-list::-webkit-scrollbar-track {
 background: transparent;
 }
 .subcategory-list::-webkit-scrollbar-thumb {
 background: rgba(8, 198, 190, 0.2);
 border-radius: 2px;
 }
 .subcategory-item {
 display: flex;
 align-items: center;
 gap: 10px;
 padding: 8px 12px;
 background: rgba(8, 198, 190, 0.04);
 border-radius: 10px;
 transition: all 0.2s ease;
 }
 .subcategory-item:hover {
 background: rgba(8, 198, 190, 0.08);
 }
 .tag-input {
 flex: 1;
 }
 .tag-input :deep(.el-input__wrapper) {
 background: transparent;
 box-shadow: none;
 border: 1px solid transparent;
 border-radius: 8px;
 transition: all 0.2s;
 }
 .tag-input :deep(.el-input__wrapper:hover),
 .tag-input :deep(.el-input__wrapper.is-focus) {
 border-color: rgba(8, 198, 190, 0.3);
 background: rgba(255, 255, 255, 0.6);
 }
 .delete-btn {
 opacity: 0.5;
 transition: all 0.2s;
 }
 .subcategory-item:hover .delete-btn {
 opacity: 1;
 }
 .delete-btn:hover {
 color: #f56c6c !important;
 }
 .empty-tags {
 text-align: center;
 color: #a0b5b4;
 font-size: 13px;
 padding: 24px 0;
 }
 .add-btn {
 width: 100%;
 border-radius: 10px;
 border: 1px dashed rgba(8, 198, 190, 0.3);
 background: transparent;
 color: #059691;
 transition: all 0.2s;
 }
 .add-btn:hover {
 border-color: rgba(8, 198, 190, 0.5);
 background: rgba(8, 198, 190, 0.05);
 }
 .dialog-footer {
 display: flex;
 justify-content: flex-end;
 gap: 10px;
 }
 .cancel-btn {
 border-radius: 8px;
 padding: 8px 20px;
 }
 .save-btn {
 border-radius: 8px;
 padding: 8px 20px;
 background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
 border: none;
 }
 .save-btn:hover {
 background: linear-gradient(135deg, #0dd9d1 0%, #08c6be 100%);
 }
 .tag-list-enter-active,
 .tag-list-leave-active {
 transition: all 0.3s ease;
 }
 .tag-list-enter-from,
 .tag-list-leave-to {
 opacity: 0;
 transform: translateX(-20px);
 }

.empty-hint {
  margin-top: 50px;
  text-align: center;
}

.preview-dialog-content {
  padding: 20px;
  max-height: 600px;
  display: flex;
  flex-direction: column;
}

.preview-dialog :deep(.el-dialog__header) {
  display: none;
}

.preview-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.preview-dialog-header {
  margin-bottom: 15px;
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
}

.preview-dialog-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-scrollbar {
  flex-grow: 1;
  overflow-y: auto;
  max-height: 500px;
}

.preview-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-meta-card {
  padding: 18px 20px;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.1) 0%, rgba(5, 150, 145, 0.16) 100%);
  border: 1px solid rgba(8, 198, 190, 0.18);
}

.preview-meta-card h2 {
  margin: 0 0 6px;
  font-size: 22px;
  color: #055f5b;
}

.preview-meta-card span {
  font-size: 13px;
  color: #4f7b79;
}

.preview-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.preview-section-label {
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: #0f766e;
}

.preview-markdown-card {
  padding: 16px 18px;
  border-radius: 16px;
  background: #ffffff;
  border: 1px solid rgba(8, 198, 190, 0.14);
  box-shadow: 0 8px 24px rgba(8, 198, 190, 0.06);
}

.preview-section-collapse :deep(.el-collapse) {
  border-top: none;
  border-bottom: none;
}

.preview-section-collapse :deep(.el-collapse-item__wrap) {
  border-bottom: none;
}

.preview-section-collapse :deep(.el-collapse-item__header) {
  height: auto;
  line-height: 1.4;
  padding: 14px 16px;
  border-radius: 16px;
  background: rgba(248, 250, 252, 0.9);
  border: 1px solid rgba(8, 198, 190, 0.14);
  color: #055f5b;
}

.preview-section-collapse :deep(.el-collapse-item__content) {
  padding-bottom: 0;
}

.preview-collapse-title {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.preview-collapse-title span {
  font-size: 14px;
  font-weight: 700;
}

.preview-collapse-title small {
  font-size: 12px;
  color: #6b7280;
}

.preview-content-card {
  margin-top: 12px;
}

.edit-prompt-dialog :deep(.el-dialog__body) {
  padding: 20px;
  overflow-y: auto;
}

.edit-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.edit-dialog-header span {
  font-size: 18px;
  font-weight: 600;
  color: #059691;
}

.edit-dialog-content {
  overflow-y: auto;
}

.edit-left-panel {
  padding-right: 10px;
}

.edit-panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 16px 0 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(8, 198, 190, 0.15);
}

.form-label {
  font-weight: 600;
  color: #059691;
}

.field-config-section {
  margin-top: 8px;
}

.empty-fields {
  padding: 20px;
  text-align: center;
  color: #909399;
}

.fields-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 12px;
}

.field-item {
  padding: 12px;
  border: 1px solid rgba(8, 198, 190, 0.2);
  border-radius: 12px;
  background: rgba(8, 198, 190, 0.04);
}

.field-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.field-name-container {
  display: flex;
  align-items: center;
  gap: 8px;
}

.required-tag {
  font-size: 11px;
}

.field-name-input {
  width: 150px;
}

.field-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.field-config-form {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.field-input {
  width: 100%;
}

.full-width-input {
  width: 100%;
}

.field-options {
  margin-top: 8px;
}

.field-options label {
  font-size: 12px;
  color: #606266;
  margin-bottom: 4px;
  display: block;
}

.form-item-label-with-guide {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prompt-guide-trigger {
  padding: 0 4px;
  color: #059691;
}

.prompt-guide-popover {
  padding: 8px;
}

.prompt-guide-title {
  font-size: 14px;
  font-weight: 600;
  color: #059691;
  margin-bottom: 8px;
}

.prompt-guide-intro {
  font-size: 13px;
  color: #606266;
  margin-bottom: 12px;
}

.prompt-guide-intro code {
  background: rgba(8, 198, 190, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  color: #059691;
}

.prompt-guide-example {
  background: rgba(8, 198, 190, 0.06);
  padding: 10px 12px;
  border-radius: 8px;
  margin-bottom: 12px;
}

.prompt-guide-example-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
  display: block;
}

.prompt-guide-example p {
  font-size: 13px;
  color: #303133;
  margin: 0;
}

.prompt-guide-example code {
  background: rgba(8, 198, 190, 0.15);
  padding: 1px 4px;
  border-radius: 3px;
  color: #059691;
}

.prompt-guide-tips {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.prompt-guide-tip {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 12px;
  color: #606266;
}

.prompt-guide-tip-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  background: rgba(8, 198, 190, 0.15);
  color: #059691;
  border-radius: 50%;
  font-size: 11px;
  flex-shrink: 0;
}

.edit-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.description-rich-editor {
  width: 100%;
}

.preview-encrypted-notice {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  background: linear-gradient(135deg, rgba(230, 162, 60, 0.08) 0%, rgba(230, 162, 60, 0.04) 100%);
  border: 1px dashed #e6a23c;
  border-radius: 12px;
  text-align: center;
}

.encrypted-title {
  font-size: 18px;
  font-weight: 600;
  color: #e6a23c;
  margin: 12px 0 0;
}

.edit-password-popover {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.popover-hint {
  font-size: 13px;
  color: #606266;
  margin: 0;
}

.popover-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
