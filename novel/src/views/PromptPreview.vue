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

    <!-- 排序 Tab -->
    <div class="sort-tab-bar">
      <div class="sort-tabs">
        <span
          v-for="tab in sortTabs"
          :key="tab.value"
          class="sort-tab-item"
          :class="{ active: selectedSort === tab.value }"
          @click="selectedSort = tab.value"
        >
          {{ tab.label }}
        </span>
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
          :data-category="prompt.category"
          @click="handlePreview(prompt)"
        >
          <span class="card-category-badge" :style="{ borderColor: getCategoryColor(prompt.category), color: getCategoryColor(prompt.category) }">
            {{ getCategoryLabel(prompt.category) }}
          </span>

          <div class="card-title-row">
            <h3 class="card-title" :title="prompt.name">{{ prompt.name }}</h3>
          </div>

          <div class="card-author-row">
            <div class="author-avatar" :style="{ background: getAvatarColor(prompt.creator_name || prompt.name) }">
              {{ (prompt.creator_name || '匿').charAt(0) }}
            </div>
            <span class="author-name" :title="prompt.creator_name">{{ prompt.creator_name || '匿名用户' }}</span>
            <span class="author-stats">
              <span class="stat-item">
                <svg class="stat-icon fire-icon" viewBox="0 0 24 24" fill="currentColor"><path d="M12 23c-4.97 0-9-3.58-9-8 0-2.52 1.18-4.78 3-6.26V6c0-.55.45-1 1-1s1 .45 1 1v1.24c1.06-.5 2.25-.79 3.5-.79h.5c.28 0 .5.22.5.5v2c0 .28-.22.5-.5.5H12c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5c0-1.13-.38-2.17-1.01-3.02C15.62 11.55 14 10.45 14 9c0-.55.45-1 1-1s1 .45 1 1c0 .83.67 1.5 1.5 1.5S19 9.83 19 9c0-2.76-2.24-5-5-5-.34 0-.67.03-1 .09V3c0-.55.45-1 1-1s1 .45 1 1v1.74c1.82 1.48 3 3.74 3 6.26 0 4.42-4.03 8-9 8z"/></svg>
                {{ prompt.use_count ?? prompt.order_num ?? 0 }}
              </span>
            </span>
            <span class="card-date">{{ formatDate(prompt.created_at) }}</span>
          </div>

          <div class="card-body-preview" :title="getPromptDescriptionPreview(prompt)">
            {{ getPromptDescriptionPreview(prompt) || '暂无简介' }}
          </div>

          <div class="card-footer-tags">
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
              添加标签
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewDialogVisible"
      width="72vw"
      top="0"
      :show-close="false"
      destroy-on-close
      class="preview-dialog"
    >
      <div class="preview-dialog-content">
        <div v-if="selectedPreviewPrompt" class="preview-dialog-header">
          <div class="preview-title-row">
            <h2 class="preview-title">{{ selectedPreviewPrompt.name }}</h2>
            <button class="preview-close-btn" @click="previewDialogVisible = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M18 6L6 18M6 6l12 12"/></svg>
            </button>
          </div>
          <div class="preview-subtitle-row">
            <span class="preview-type-label">提示词类型：{{ getCategoryLabel(selectedPreviewPrompt.category) }}</span>
          </div>
          <div v-if="(selectedPreviewPrompt.subcategories || []).length > 0" class="preview-tags-row">
            <span
              v-for="(subcat, idx) in (selectedPreviewPrompt.subcategories || [])"
              :key="idx"
              class="preview-tag"
              :style="{
                background: getTagColor(subcat).bg,
                color: getTagColor(subcat).text,
                borderColor: getTagColor(subcat).border
              }"
            >{{ subcat }}</span>
          </div>
          <div class="preview-author-row">
            <div class="preview-author-info">
              <div class="preview-author-avatar" :style="{ background: getAvatarColor(selectedPreviewPrompt.creator_name || selectedPreviewPrompt.name) }">
                {{ (selectedPreviewPrompt.creator_name || '匿').charAt(0) }}
              </div>
              <span class="preview-author-name">{{ selectedPreviewPrompt.creator_name || '匿名用户' }}</span>
              <div v-if="selectedPreviewPrompt.version" class="preview-version-badge">
                {{ selectedPreviewPrompt.version }}
              </div>
            </div>
            <div class="preview-dialog-actions">
              <template v-if="selectedPreviewPrompt.card_type === 'encrypted' && selectedPreviewPrompt.password">
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
            </div>
          </div>
        </div>
        <el-scrollbar class="preview-scrollbar">
          <div v-if="selectedPreviewPrompt" class="preview-body">
            <div class="preview-section">
              <div class="preview-section-label">简介</div>
              <div class="preview-markdown-card">
                <MarkdownRenderer :content="selectedPreviewPrompt.description?.trim() || '无'" />
              </div>
            </div>

            <div class="preview-section">
              <template v-if="selectedPreviewPrompt.card_type === 'encrypted' && selectedPreviewPrompt.password">
                <div class="preview-encrypted-notice">
                  <el-icon :size="32" color="#e6a23c"><Lock /></el-icon>
                  <p class="encrypted-title">私人提示词</p>
                </div>
              </template>
              <template v-else>
                <div class="preview-section-label">提示词内容</div>
                <div class="preview-markdown-card preview-content-card">
                  <MarkdownRenderer :content="selectedPreviewPrompt.content" />
                </div>
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
          <el-form-item label="作者名">
            <el-input v-model="editFormData.creator_name" placeholder="请输入作者名，非必填" />
          </el-form-item>
          <el-form-item label="版本号">
            <el-input v-model="editFormData.version" placeholder="如 v1.0、v2.1.0，非必填" />
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

// 排序选项
const selectedSort = ref('latest')
const sortTabs = [
  { label: '最新', value: 'latest' },
  { label: '最热', value: 'hottest' },
  { label: '名称A-Z', value: 'name_asc' },
]

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
  order_num: 0,
  creator_name: '',
  version: ''
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

const categoryColorMap: Record<string, string> = {}
const categoryLabelMap: Record<string, string> = {}
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
  '#a18cd1', '#fccb90', '#f5576c', '#764ba2', '#00f2fe'
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
const tagColorCache: Record<string, typeof tagColorPalette[0]> = {}

const getTagColor = (tagText: string): typeof tagColorPalette[0] => {
  if (!tagColorCache[tagText]) {
    let hash = 0
    for (let i = 0; i < tagText.length; i++) {
      hash = tagText.charCodeAt(i) + ((hash << 5) - hash)
    }
    tagColorCache[tagText] = tagColorPalette[Math.abs(hash) % tagColorPalette.length]
  }
  return tagColorCache[tagText]
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

  // 排序
  switch (selectedSort.value) {
    case 'latest':
      result = [...result].sort((a, b) => {
        const timeA = a.created_at ? new Date(a.created_at).getTime() : 0
        const timeB = b.created_at ? new Date(b.created_at).getTime() : 0
        return timeB - timeA
      })
      break
    case 'hottest':
      result = [...result].sort((a, b) => (b.order_num || 0) - (a.order_num || 0))
      break
    case 'name_asc':
      result = [...result].sort((a, b) => a.name.localeCompare(b.name, 'zh-CN'))
      break
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
    order_num: prompt.order_num,
    creator_name: prompt.creator_name || '',
    version: prompt.version || ''
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
  min-height: 100vh;
  padding: 24px;
  background: linear-gradient(180deg, #eaf6f5 0%, #d9f0ee 30%, #e8f4f2 60%, #f0f7f6 100%);
  background-attachment: fixed;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid #e8ecef;
  border-radius: 12px;
  padding: 20px 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
  position: relative;
}

.header h2 {
  font-size: 22px;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0;
  position: relative;
  z-index: 1;
}

.header h2::before {
  content: '';
  display: inline-block;
  width: 4px;
  height: 20px;
  background: linear-gradient(180deg, #667eea, #764ba2);
  border-radius: 2px;
  margin-right: 10px;
  vertical-align: middle;
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
  background: #f5f7fa;
  border: 1px solid #e0e4ea;
  box-shadow: none;
  transition: all 0.25s;
}

.search-input :deep(.el-input__wrapper:hover),
.search-input :deep(.el-input__wrapper.is-focus) {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.search-input :deep(.el-input__inner) {
  color: #333;
}

.search-input :deep(.el-input__inner::placeholder) {
  color: #a0aab8;
}

.category-selector {
 margin-bottom: 20px;
 background: rgba(255, 255, 255, 0.85);
 backdrop-filter: blur(12px);
 -webkit-backdrop-filter: blur(12px);
 border: 1px solid rgba(255, 255, 255, 0.6);
 padding: 12px 20px;
 border-radius: 12px;
 box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
 display: flex;
 align-items: center;
 gap: 12px;
 }
 .category-title {
 font-weight: 600;
 color: #555;
 white-space: nowrap;
 flex-shrink: 0;
 font-size: 14px;
 }
 .category-buttons-wrapper {
 flex: 1;
 overflow-x: auto;
 overflow-y: hidden;
 scrollbar-width: thin;
 scrollbar-color: #ddd transparent;
 }
 .category-buttons-wrapper::-webkit-scrollbar {
 height: 4px;
 }
 .category-buttons-wrapper::-webkit-scrollbar-track {
 background: transparent;
 }
 .category-buttons-wrapper::-webkit-scrollbar-thumb {
 background: #ddd;
 border-radius: 2px;
 }
 .category-buttons-wrapper::-webkit-scrollbar-thumb:hover {
 background: #bbb;
 }
 .category-buttons {
 display: flex;
 gap: 8px;
 flex-wrap: nowrap;
 }

.category-buttons .el-button {
  border-radius: 16px;
  font-size: 12px;
  padding: 4px 14px;
  min-width: auto;
  background: #f5f7fa;
  border: 1px solid #e2e8f0;
  color: #555;
  transition: all 0.25s;
}

.category-buttons .el-button:hover {
  background: #eef1f6;
  border-color: #c8d0da;
  color: #333;
  transform: translateY(-1px);
}

.category-buttons .el-button--primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  color: #fff;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.category-buttons .el-button--primary:hover {
  background: linear-gradient(135deg, #7c93f0 0%, #8a5db8 100%);
  box-shadow: 0 4px 14px rgba(102, 126, 234, 0.4);
}

.sort-tab-bar {
 margin-bottom: 16px;
 display: flex;
 align-items: center;
 }
 .sort-tabs {
 display: flex;
 gap: 4px;
 background: rgba(255, 255, 255, 0.7);
 backdrop-filter: blur(8px);
 -webkit-backdrop-filter: blur(8px);
 border-radius: 10px;
 padding: 4px;
 border: 1px solid rgba(255, 255, 255, 0.5);
 }
 .sort-tab-item {
 padding: 6px 18px;
 border-radius: 8px;
 font-size: 13px;
 font-weight: 500;
 color: #666;
 cursor: pointer;
 transition: all 0.25s;
 user-select: none;
 white-space: nowrap;
 }
 .sort-tab-item:hover {
 color: #333;
 background: rgba(0, 0, 0, 0.04);
 }
 .sort-tab-item.active {
 background: #fff;
 color: #667eea;
 font-weight: 600;
 box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
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
 background: #fff;
 border: 1px solid rgba(255, 255, 255, 0.8);
 border-radius: 12px;
 box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06), 0 1px 3px rgba(0, 0, 0, 0.04);
 padding: 0;
 transition: all 0.28s cubic-bezier(0.4, 0, 0.2, 1);
 height: 280px;
 display: flex;
 flex-direction: column;
 position: relative;
 overflow: hidden;
 cursor: pointer;
 }
 .card-category-badge {
 position: absolute;
 top: 12px;
 right: 12px;
 font-size: 11px;
 padding: 3px 10px;
 border-radius: 6px;
 font-weight: 500;
 z-index: 2;
 pointer-events: none;
 background: rgba(255, 255, 255, 0.15);
 backdrop-filter: blur(12px);
 -webkit-backdrop-filter: blur(12px);
 border: 1px solid rgba(255, 255, 255, 0.25);
}
 .prompt-card:hover {
 box-shadow: 0 10px 28px rgba(0, 0, 0, 0.1), 0 2px 8px rgba(0, 0, 0, 0.06);
 transform: translateY(-4px);
 border-color: rgba(255, 255, 255, 1);
 }
 .card-title-row {
 padding: 14px 16px 6px;
 padding-right: 70px;
 }
 .card-title {
 font-size: 15px;
 font-weight: 700;
 color: #1a1a2e;
 margin: 0;
 line-height: 1.4;
 overflow: hidden;
 text-overflow: ellipsis;
 display: -webkit-box;
 -webkit-line-clamp: 2;
 -webkit-box-orient: vertical;
 word-break: break-word;
 }

 .card-style-tag-row {
 padding: 2px 16px 6px;
 }
 .card-style-tag {
 display: inline-block;
 font-size: 11px;
 font-weight: 600;
 color: #b45309;
 background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
 border: 1px solid #fcd34d;
 padding: 2px 10px;
 border-radius: 10px;
 letter-spacing: 0.3px;
 }

 .card-author-row {
 display: flex;
 align-items: center;
 gap: 8px;
 padding: 6px 16px 10px;
 }
 .author-avatar {
 width: 28px;
 height: 28px;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 color: #fff;
 font-size: 12px;
 font-weight: 700;
 flex-shrink: 0;
 box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
 }
 .author-name {
 font-size: 12px;
 color: #555;
 max-width: 72px;
 overflow: hidden;
 text-overflow: ellipsis;
 white-space: nowrap;
 flex-shrink: 0;
 }
 .author-stats {
 display: flex;
 align-items: center;
 gap: 4px;
 margin-left: auto;
 flex-shrink: 0;
 }
 .stat-item {
 display: inline-flex;
 align-items: center;
 gap: 3px;
 font-size: 11px;
 color: #e8590c;
 font-weight: 600;
 background: rgba(232, 89, 12, 0.06);
 padding: 2px 8px;
 border-radius: 10px;
 }
 .stat-icon {
 width: 13px;
 height: 13px;
 }
 .fire-icon {
 color: #e8590c;
 }
 .card-date {
 font-size: 11px;
 color: #bbb;
 flex-shrink: 0;
 }

 .card-body-preview {
 flex: 1;
 padding: 0 16px;
 font-size: 13px;
 color: #666;
 line-height: 1.65;
 overflow: hidden;
 display: -webkit-box;
 -webkit-line-clamp: 4;
 -webkit-box-orient: vertical;
 word-break: break-word;
 min-height: 0;
 }

 .card-footer-tags {
 display: flex;
 flex-wrap: wrap;
 gap: 6px;
 padding: 10px 16px 14px;
 align-items: center;
 border-top: 1px solid #f5f6f8;
 margin-top: auto;
 }
 .subcategory-tag {
 font-size: 11px;
 border-radius: 12px;
 margin: 0;
 background: #f0f2f5;
 border-color: #e1e5ea;
 color: #666;
 transition: all 0.2s;
 max-width: 80px;
 overflow: hidden;
 text-overflow: ellipsis;
 white-space: nowrap;
 }
 .subcategory-tag:hover {
 background: #e4e7ed;
 border-color: #c8d0da;
 transform: scale(1.05);
 }
 .more-tags {
 font-size: 11px;
 color: #999;
 padding: 2px 6px;
 background: #f5f6f8;
 border-radius: 10px;
 }
 .add-subcategory-btn {
 font-size: 11px;
 padding: 0;
 color: #667eea;
 }
 .add-subcategory-btn:hover {
 color: #764ba2;
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
 color: #333;
 }
 .subcategory-count {
 font-size: 12px;
 color: #999;
 background: #f5f6f8;
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
 background: #ddd;
 border-radius: 2px;
 }
 .subcategory-item {
 display: flex;
 align-items: center;
 gap: 10px;
 padding: 8px 12px;
 background: #f8f9fb;
 border-radius: 10px;
 transition: all 0.2s ease;
 border: 1px solid transparent;
 }
 .subcategory-item:hover {
 background: #f0f2f5;
 border-color: #e2e8f0;
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
 border-color: #c8d0da;
 background: #fff;
 }
 .delete-btn {
 opacity: 0.5;
 transition: all 0.2s;
 }
 .subcategory-item:hover .delete-btn {
 opacity: 1;
 }
 .empty-tags {
 text-align: center;
 color: #bbb;
 font-size: 13px;
 padding: 24px 0;
 }
 .add-btn {
 width: 100%;
 border-radius: 10px;
 border: 1px dashed #ddd;
 background: transparent;
 color: #667eea;
 transition: all 0.2s;
 }
 .add-btn:hover {
 border-color: #667eea;
 background: rgba(102, 126, 234, 0.04);
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
 background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
 border: none;
 }
 .save-btn:hover {
 background: linear-gradient(135deg, #7c93f0 0%, #8a5db8 100%);
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
  padding: 24px 28px;
  height: 90vh;
  display: flex;
  flex-direction: column;
}

.preview-dialog :deep(.el-dialog__header) {
  display: none;
}

.preview-dialog :deep(.el-dialog__body) {
  padding: 0;
  height: 88vh;
}

.preview-dialog :deep(.el-dialog__wrapper) {
  height: 100vh !important;
  padding: 0 !important;
}

.preview-dialog :deep(.el-dialog) {
  height: 90vh !important;
  max-height: 90vh !important;
  margin: 0 auto !important;
  transform: translateY(0) !important;
  display: flex !important;
  flex-direction: column !important;
}

.preview-dialog-header {
  margin-bottom: 20px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.preview-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.preview-title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: #1a1a2e;
  line-height: 1.3;
  flex: 1;
  min-width: 0;
}

.preview-close-btn {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border: none;
  background: #f0f0f2;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
  transition: all 0.2s;
}

.preview-close-btn:hover {
  background: #e8e8ea;
  color: #333;
}

.preview-close-btn svg {
  width: 16px;
  height: 16px;
}

.preview-subtitle-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.preview-type-label {
  font-size: 13px;
  color: #888;
}

.preview-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.preview-tag {
  display: inline-block;
  padding: 4px 14px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  border: 1px solid;
  transition: opacity 0.2s;
}

.preview-tag:hover {
  opacity: 0.8;
}

.preview-author-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding-top: 4px;
}

.preview-author-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.preview-author-avatar {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.preview-author-name {
  font-size: 14px;
  color: #444;
  font-weight: 500;
}

.preview-heat-block {
  margin-left: 4px;
}

.heat-number {
  font-size: 22px;
  font-weight: 800;
  color: #f59e0b;
  letter-spacing: -0.02em;
}

.preview-version-badge {
  margin-left: 10px;
  padding: 2px 12px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  background: rgba(102, 126, 234, 0.1);
  color: #667eea;
  border: 1px solid rgba(102, 126, 234, 0.25);
}

.preview-dialog-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.preview-scrollbar {
  flex-grow: 1;
  overflow-y: auto;
  min-height: 0;
}

.preview-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.preview-meta-card {
 padding: 18px 20px;
 border-radius: 14px;
 background: #f8f9fb;
 border: 1px solid #e8ecef;
}

.preview-meta-card h2 {
 margin: 0 0 6px;
 font-size: 22px;
 color: #1a1a2e;
}

.preview-meta-card span {
 font-size: 13px;
 color: #888;
}

.preview-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.preview-section-label {
 font-size: 14px;
 font-weight: 700;
 letter-spacing: 0.04em;
 color: #1a1a2e;
 padding-left: 12px;
 border-left: 3px solid #10b981;
 line-height: 1.4;
}

.preview-markdown-card {
 padding: 16px 18px;
 border-radius: 12px;
 background: #fff;
 border: 1px solid #e8ecef;
 box-shadow: 0 1px 4px rgba(0, 0, 0, 0.03);
}

.preview-content-card {
  margin-top: 0;
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
 color: #333;
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
 border-bottom: 1px solid #eee;
}

.form-label {
 font-weight: 600;
 color: #444;
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
 border: 1px solid #e8ecef;
 border-radius: 10px;
 background: #fafbfc;
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
 color: #667eea;
}

.prompt-guide-popover {
 padding: 8px;
}

.prompt-guide-title {
 font-size: 14px;
 font-weight: 600;
 color: #333;
 margin-bottom: 8px;
}

.prompt-guide-intro {
 font-size: 13px;
 color: #606266;
 margin-bottom: 12px;
}

.prompt-guide-intro code {
 background: #f0f2f5;
 padding: 2px 6px;
 border-radius: 4px;
 color: #667eea;
}

.prompt-guide-example {
 background: #f8f9fb;
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
 background: #edeef1;
 padding: 1px 4px;
 border-radius: 3px;
 color: #667eea;
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
 background: #edeef1;
 color: #667eea;
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

:root[data-theme='dark'] .prompt-preview-container {
  background: linear-gradient(180deg, #0d1424 0%, #111a2e 30%, #131d30 60%, #0f1729 100%);
}

:root[data-theme='dark'] .header {
  background: rgba(26, 35, 50, 0.88);
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .header h2 {
  color: #f4f7ff;
}

:root[data-theme='dark'] .search-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}

:root[data-theme='dark'] .search-input :deep(.el-input__wrapper:hover),
:root[data-theme='dark'] .search-input :deep(.el-input__wrapper.is-focus) {
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
}

:root[data-theme='dark'] .search-input :deep(.el-input__inner) {
  color: #f4f7ff;
}

:root[data-theme='dark'] .search-input :deep(.el-input__inner::placeholder) {
  color: #6b7280;
}

:root[data-theme='dark'] .category-selector {
  background: rgba(26, 35, 50, 0.88);
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .category-title {
  color: #c8d0e0;
}

:root[data-theme='dark'] .category-buttons-wrapper {
  scrollbar-color: rgba(255, 255, 255, 0.15) transparent;
}

:root[data-theme='dark'] .category-buttons-wrapper::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
}

:root[data-theme='dark'] .category-buttons-wrapper::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.25);
}

:root[data-theme='dark'] .category-buttons .el-button {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
  color: #c8d0e0;
}

:root[data-theme='dark'] .category-buttons .el-button:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.18);
  color: #f4f7ff;
}

:root[data-theme='dark'] .category-buttons .el-button--primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  color: #fff;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.35);
}

:root[data-theme='dark'] .category-buttons .el-button--primary:hover {
  background: linear-gradient(135deg, #7c93f0 0%, #8a5db8 100%);
  box-shadow: 0 4px 14px rgba(102, 126, 234, 0.45);
}

:root[data-theme='dark'] .sort-tabs {
  background: rgba(26, 35, 50, 0.7);
  border-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme='dark'] .sort-tab-item {
  color: #8d99af;
}

:root[data-theme='dark'] .sort-tab-item:hover {
  color: #c8d0e0;
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .sort-tab-item.active {
  background: rgba(35, 45, 63, 0.95);
  color: #667eea;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .prompt-card {
  background: #1a2332;
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.25), 0 1px 3px rgba(0, 0, 0, 0.15);
}

:root[data-theme='dark'] .prompt-card:hover {
  box-shadow: 0 10px 28px rgba(0, 0, 0, 0.35), 0 2px 8px rgba(0, 0, 0, 0.2);
  border-color: rgba(102, 126, 234, 0.35);
}

:root[data-theme='dark'] .card-title {
  color: #f4f7ff;
}

:root[data-theme='dark'] .card-style-tag {
  color: #fbbf24;
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.15) 0%, rgba(245, 158, 11, 0.1) 100%);
  border-color: rgba(251, 191, 36, 0.3);
}

:root[data-theme='dark'] .author-name {
  color: #c8d0e0;
}

:root[data-theme='dark'] .stat-item {
  color: #fb923c;
  background: rgba(251, 146, 60, 0.1);
}

:root[data-theme='dark'] .fire-icon {
  color: #fb923c;
}

:root[data-theme='dark'] .card-date {
  color: #6b7280;
}

:root[data-theme='dark'] .card-body-preview {
  color: #a0aab8;
}

:root[data-theme='dark'] .card-footer-tags {
  border-top-color: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .subcategory-tag {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.1);
  color: #a0aab8;
}

:root[data-theme='dark'] .subcategory-tag:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.16);
}

:root[data-theme='dark'] .more-tags {
  color: #6b7280;
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .add-subcategory-btn {
  color: #818cf8;
}

:root[data-theme='dark'] .add-subcategory-btn:hover {
  color: #a78bfa;
}

:root[data-theme='dark'] .card-category-badge {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.12);
}

:root[data-theme='dark'] .preview-version-badge {
  background: rgba(102, 126, 234, 0.15);
  color: #818cf8;
  border-color: rgba(102, 126, 234, 0.3);
}

:root[data-theme='dark'] .preview-section-label {
  color: #f4f7ff;
  border-left-color: #34d399;
}

:root[data-theme='dark'] .preview-markdown-card {
  background: rgba(26, 35, 50, 0.7);
  border-color: rgba(255, 255, 255, 0.08);
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.25);
}

:root[data-theme='dark'] .preview-title {
  color: #f4f7ff;
}

:root[data-theme='dark'] .preview-author-name {
  color: #c8d0e0;
}

:root[data-theme='dark'] .heat-number {
  color: #fbbf24;
}
</style>
