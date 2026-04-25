<template>
  <div class="book-analysis-prompts-page">
    <div class="page-header">
      <div class="header-main">
        <el-button class="back-button" @click="goBack"><el-icon><ArrowLeft /></el-icon>返回</el-button>
        <div>
          <h1>拆书库提示词</h1>
          <p>管理拆书分析使用的提示词卡片、分类、字段配置与本地备份文件。</p>
        </div>
      </div>
      <div class="header-actions">
        <el-dropdown @command="handleExportCommand">
          <el-button><el-icon><Download /></el-icon>导出</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="backup">完整备份</el-dropdown-item>
              <el-dropdown-item command="standalone">导出未分类提示词</el-dropdown-item>
              <el-dropdown-item command="pack">导出已分类提示词</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-dropdown @command="handleImportCommand">
          <el-button><el-icon><Upload /></el-icon>导入</el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="backup">恢复完整备份</el-dropdown-item>
              <el-dropdown-item command="standalone">导入独立卡片</el-dropdown-item>
              <el-dropdown-item command="pack">导入卡包</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button @click="resetStorage"><el-icon><RefreshLeft /></el-icon>重置数据</el-button>
        <el-button type="success" @click="openCategoryDialog"><el-icon><FolderAdd /></el-icon>新建分类</el-button>
        <el-button type="primary" @click="openPromptDialog()"><el-icon><Plus /></el-icon>新建提示词</el-button>
      </div>
    </div>

    <div class="summary-grid">
      <div class="summary-card"><span class="summary-label">提示词总数</span><strong>{{ prompts.length }}</strong></div>
      <div class="summary-card"><span class="summary-label">默认分类</span><strong>{{ defaultPrompts.length }}</strong></div>
      <div class="summary-card"><span class="summary-label">自定义分类</span><strong>{{ customCategoryList.length }}</strong></div>
      <div class="summary-card"><span class="summary-label">未分类</span><strong>{{ uncategorizedPrompts.length }}</strong></div>
    </div>

    <div class="section-title">
      <el-icon><FolderOpened /></el-icon>
      <span>卡包</span>
      <span class="section-count">{{ categorySections.length }}</span>
    </div>

    <div class="packs-wrapper">
      <div
        v-for="section in categorySections"
        :key="section.name"
        class="card-pack-vertical"
        @dragover.prevent
        @drop.prevent="handleDrop(section.name)"
      >
        <div class="pack-header-vertical">
          <div class="pack-header-left">
            <el-icon class="pack-icon-vertical"><FolderOpened /></el-icon>
            <span class="pack-name-vertical">{{ section.name }}</span>
            <span class="pack-count-badge">{{ section.prompts.length }}</span>
          </div>
          <div class="pack-header-actions">
            <el-button type="primary" size="small" plain @click="openPromptDialog(section.name)">新增</el-button>
            <el-button
              v-if="section.name !== DEFAULT_CATEGORY"
              type="danger"
              size="small"
              plain
              @click="deleteCategory(section.name)"
            >
              删除
            </el-button>
          </div>
        </div>

        <div class="cards-scroll-area">
          <div
            v-for="prompt in section.prompts"
            :key="prompt.id"
            class="pack-card"
            draggable="true"
            @dragstart="handleDragStart(prompt.id)"
            @dragend="handleDragEnd"
            @click="openPromptDialog(prompt.category, prompt)"
          >
            <div class="pack-card-edge"></div>
            <div class="pack-card-header">
              <span class="pack-card-name">{{ prompt.name }}</span>
              <el-tag size="small" type="primary">{{ prompt.category }}</el-tag>
            </div>
            <div class="pack-card-preview">{{ prompt.content.slice(0, 54) }}{{ prompt.content.length > 54 ? '...' : '' }}</div>
            <div class="pack-card-hover-info">
              <span class="pack-card-hover-label">卡面详情</span>
              <p>{{ prompt.description || prompt.content.slice(0, 96) }}{{ (prompt.description || prompt.content).length > 96 ? '...' : '' }}</p>
            </div>
            <div class="pack-card-footer">
              <span class="pack-card-time">{{ formatDate(prompt.updated_at || prompt.created_at) }}</span>
              <div class="pack-card-actions" @click.stop>
                <el-button link type="primary" @click="openPreview(prompt)">预览</el-button>
                <el-button link type="danger" @click="deletePrompt(prompt)">删除</el-button>
              </div>
            </div>
          </div>

          <div v-if="section.prompts.length === 0" class="empty-pack-hint" @click="openPromptDialog(section.name)">
            <el-icon><Plus /></el-icon>
            <span>点击添加提示词</span>
          </div>
        </div>

        <div class="pack-footer-vertical">
          <el-button type="primary" link @click="openPromptDialog(section.name)">新增提示词</el-button>
          <el-button v-if="section.name !== DEFAULT_CATEGORY" type="danger" link @click="deleteCategory(section.name)">删除分类</el-button>
        </div>
      </div>
    </div>

    <div class="section-title standalone-title">
      <el-icon><Document /></el-icon>
      <span>未分类提示词</span>
      <span class="section-count">{{ uncategorizedPrompts.length }}</span>
    </div>

    <div
      v-if="uncategorizedPrompts.length > 0"
      class="standalone-cards-area"
      @dragover.prevent
      @drop.prevent="handleDrop(UNCATEGORIZED_CATEGORY)"
    >
      <div class="standalone-cards-grid">
        <div
          v-for="prompt in uncategorizedPrompts"
          :key="prompt.id"
          class="standalone-card"
          draggable="true"
          @dragstart="handleDragStart(prompt.id)"
          @dragend="handleDragEnd"
          @click="openPromptDialog(UNCATEGORIZED_CATEGORY, prompt)"
        >
          <div class="standalone-card-edge"></div>
          <div class="standalone-card-header">
            <span class="standalone-card-name">{{ prompt.name }}</span>
            <el-tag size="small" type="info">{{ UNCATEGORIZED_CATEGORY }}</el-tag>
          </div>
          <div class="standalone-card-content">
            <div class="standalone-card-preview">{{ prompt.content.slice(0, 80) }}{{ prompt.content.length > 80 ? '...' : '' }}</div>
          </div>
          <div class="standalone-card-hover-info">
            <span class="standalone-card-hover-label">快速预览</span>
            <p>{{ prompt.description || prompt.content.slice(0, 120) }}{{ (prompt.description || prompt.content).length > 120 ? '...' : '' }}</p>
          </div>
          <div class="standalone-card-footer">
            <span class="standalone-card-time">{{ formatDate(prompt.updated_at || prompt.created_at) }}</span>
            <div class="standalone-card-actions" @click.stop>
              <el-button link type="primary" @click="openPreview(prompt)">预览</el-button>
              <el-button link type="danger" @click="deletePrompt(prompt)">删除</el-button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div
      v-else
      class="empty-standalone-hint"
      @click="openPromptDialog(UNCATEGORIZED_CATEGORY)"
      @dragover.prevent
      @drop.prevent="handleDrop(UNCATEGORIZED_CATEGORY)"
    >
      <el-icon><Plus /></el-icon>
      <span>暂无未分类提示词，点击创建或拖拽到这里</span>
    </div>

    <el-dialog v-model="categoryDialogVisible" title="新建分类" width="420px" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item label="分类名称"><el-input v-model="categoryName" placeholder="请输入分类名称" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="categoryDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createCategory">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="promptDialogVisible"
      :title="editingPromptId === null ? '新建提示词' : '编辑提示词'"
      :width="dialogWidthMap[dialogSize]"
      top="6vh"
      append-to-body
      :close-on-click-modal="false"
      destroy-on-close
      class="prompt-dialog"
    >
      <template #header>
        <div class="dialog-header">
          <span>{{ editingPromptId === null ? '新建提示词' : '编辑提示词' }}</span>
          <div class="size-controls">
            <el-button-group>
              <el-button :type="dialogSize === 'small' ? 'primary' : ''" size="small" @click="dialogSize = 'small'">小</el-button>
              <el-button :type="dialogSize === 'medium' ? 'primary' : ''" size="small" @click="dialogSize = 'medium'">中</el-button>
              <el-button :type="dialogSize === 'large' ? 'primary' : ''" size="small" @click="dialogSize = 'large'">大</el-button>
            </el-button-group>
          </div>
        </div>
      </template>
      <div class="dialog-content" :style="{ height: dialogHeightMap[dialogSize] }">
        <div class="left-panel">
          <el-form label-position="top">
            <el-form-item label="名称" required>
              <el-input v-model="form.name" placeholder="例如：章节剧情拆解" />
            </el-form-item>
            <el-form-item label="分类" required>
              <el-select v-model="form.category" placeholder="请选择分类" class="full-width field-input">
                <el-option v-for="item in categoryOptions" :key="item" :label="item" :value="item" />
              </el-select>
            </el-form-item>
            <el-form-item label="简介">
              <SplitRichTextEditor
                v-model="form.description"
                placeholder="用于介绍这个提示词，非必填"
                class="description-rich-editor"
              />
            </el-form-item>
            <el-form-item required>
              <template #label>
                <div class="form-item-label-with-guide">
                  <span>内容</span>
                  <el-popover placement="right-start" :width="360" trigger="click">
                    <template #reference>
                      <el-button size="small" text class="prompt-guide-trigger">
                        <el-icon><InfoFilled /></el-icon>
                        模板说明
                      </el-button>
                    </template>
                    <div class="prompt-guide-popover">
                      <div class="prompt-guide-title">先写完整提示，再把变量换成占位符</div>
                      <p class="prompt-guide-intro">
                        正文里直接写完整提示词，把需要用户补充的内容写成 <code v-pre>{{字段名称}}</code>。
                      </p>
                      <div class="prompt-guide-example">
                        <span class="prompt-guide-example-label">示例</span>
                        <p v-pre>请从 {{章节范围}} 中提取 {{分析维度}}，并输出适合 {{用途}} 的拆书结果。</p>
                      </div>
                    </div>
                  </el-popover>
                </div>
              </template>
              <el-input
                v-model="form.content"
                type="textarea"
                :rows="15"
                placeholder="请输入提示词内容"
              />
            </el-form-item>

            <div class="panel-header prompt-panel-header">
              <label class="form-label">字段配置</label>
              <div class="field-panel-actions">
                <el-button type="primary" plain size="small" @click="extractFieldsFromContent">
                  识别占位符
                </el-button>
                <el-button type="primary" size="small" @click="addField">
                  <el-icon><Plus /></el-icon>
                  添加字段
                </el-button>
              </div>
            </div>

            <div class="field-config-section">
              <el-alert title="字段配置说明" type="info" :closable="false" show-icon>
                <template #default>
                  <div class="format-description">
                    <p><strong>配置字段信息</strong></p>
                    <p>为每个 <code v-pre>{{字段名称}}</code> 设置表单展示方式，决定用户如何填写。</p>
                  </div>
                </template>
              </el-alert>

              <div v-if="form.fields.length === 0" class="empty-fields">
                <p class="empty-hint">暂无字段配置，点击上方按钮添加或从内容里自动识别。</p>
              </div>

              <div v-else class="field-list">
                <div v-for="(field, index) in form.fields" :key="`${field.name}-${index}`" class="field-item">
                  <div class="field-header">
                    <div class="field-name-container">
                      <el-tag :type="field.required ? 'danger' : 'info'" size="small" class="required-tag">
                        {{ field.required ? '必填' : '选填' }}
                      </el-tag>
                      <el-input
                        v-model="field.name"
                        class="field-name-input"
                        placeholder="字段名称"
                        @focus="rememberFieldName(index, field.name)"
                        @change="updateFieldName(index, field.name)"
                      />
                      <el-button type="primary" link size="small" @click="copyFieldName(field.name)" title="复制占位符">
                        <el-icon><CopyDocument /></el-icon>
                      </el-button>
                    </div>
                    <div class="field-actions">
                      <el-switch
                        v-model="field.required"
                        active-text="必填"
                        inactive-text="选填"
                        :active-action-icon="Check"
                        :inactive-action-icon="Close"
                        inline-prompt
                      />
                      <el-button type="primary" link size="small" :disabled="index === 0" @click="moveField(index, index - 1)" title="上移">
                        <el-icon><ArrowUp /></el-icon>
                      </el-button>
                      <el-button type="primary" link size="small" :disabled="index === form.fields.length - 1" @click="moveField(index, index + 1)" title="下移">
                        <el-icon><ArrowDown /></el-icon>
                      </el-button>
                      <el-button type="danger" link size="small" @click="removeField(index)" title="删除">
                        <el-icon><Delete /></el-icon>
                      </el-button>
                    </div>
                  </div>

                  <div class="field-config-form">
                    <el-input v-model="field.label" placeholder="字段显示名称" class="field-input" />
                    <el-select v-model="field.type" placeholder="字段类型" class="field-input">
                      <el-option label="单行文本" value="text" />
                      <el-option label="多行文本" value="textarea" />
                      <el-option label="下拉选择" value="select" />
                    </el-select>
                    <div v-if="field.type === 'select'" class="field-options">
                      <label>选项（每行一个）</label>
                      <el-input
                        :model-value="getFieldOptionsText(field)"
                        type="textarea"
                        :rows="3"
                        placeholder="请输入选项，每行一个"
                        class="field-input"
                        @update:model-value="updateFieldOptions(index, $event)"
                      />
                    </div>
                    <el-input v-model="field.description" type="textarea" :rows="2" placeholder="字段说明" class="field-input" />
                  </div>
                </div>
              </div>
            </div>
          </el-form>
        </div>

        <div class="right-panel">
          <el-form label-position="top">
            <el-form-item label="排序">
              <el-input-number v-model="form.order_num" :min="0" class="full-width" />
            </el-form-item>
            <el-form-item label="标签">
              <div class="subcategory-form">
                <div class="subcategory-tags-row">
                  <el-tag
                    v-for="(subcat, index) in formSubcategories"
                    :key="`${subcat}-${index}`"
                    closable
                    size="small"
                    class="subcategory-tag"
                    @close="removeFormSubcategory(index)"
                  >
                    {{ subcat }}
                  </el-tag>
                </div>
                <div class="subcategory-input-row">
                  <el-input
                    v-model="newSubcategory"
                    placeholder="输入标签后按回车添加"
                    size="small"
                    class="subcategory-input"
                    @keyup.enter="addFormSubcategory"
                  />
                  <el-button size="small" type="primary" @click="addFormSubcategory">
                    <el-icon><Plus /></el-icon>
                  </el-button>
                </div>
              </div>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <el-button @click="promptDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="savePrompt">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="previewDialogVisible" title="提示词预览" width="760px">
      <template v-if="previewPrompt">
        <div class="preview-header-row">
          <div><h3>{{ previewPrompt.name }}</h3><p>{{ previewPrompt.description || '暂无描述。' }}</p></div>
          <el-tag>{{ previewPrompt.category }}</el-tag>
        </div>
        <div v-if="previewPrompt.subcategories?.length" class="tag-list">
          <el-tag v-for="item in previewPrompt.subcategories" :key="item" size="small" effect="plain">{{ item }}</el-tag>
        </div>
        <div v-if="previewPrompt.fields?.length" class="preview-fields">
          <h4>字段配置</h4>
          <ul>
            <li v-for="field in previewPrompt.fields" :key="field.name">{{ field.label || field.name }} / {{ formatFieldType(field.type) }} / {{ field.required ? '必填' : '选填' }}</li>
          </ul>
        </div>
        <pre class="preview-content">{{ previewPrompt.content }}</pre>
      </template>
    </el-dialog>

    <input ref="fileInputRef" class="hidden-input" type="file" accept=".json" @change="handleFileImport" />
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ArrowLeft,
  ArrowDown,
  ArrowUp,
  Check,
  Close,
  CopyDocument,
  Delete,
  Document,
  Download,
  FolderAdd,
  FolderOpened,
  InfoFilled,
  Plus,
  RefreshLeft,
  Upload,
} from '@element-plus/icons-vue'
import type { Prompt } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'

const DEFAULT_CATEGORY = '默认'
const UNCATEGORIZED_CATEGORY = '未分类'
const LOCAL_STORAGE_PREFIX = 'book-analysis-prompts'

type PromptField = NonNullable<Prompt['fields']>[number]
type ImportMode = 'backup' | 'standalone' | 'pack'

interface PromptFormState {
  name: string
  description: string
  content: string
  category: string
  order_num: number
  subcategoriesText: string
  fields: PromptField[]
}

const router = useRouter()
const prompts = ref<Prompt[]>([])
const customCategories = ref<string[]>([])
const draggedPromptId = ref<number | null>(null)
const categoryDialogVisible = ref(false)
const categoryName = ref('')
const promptDialogVisible = ref(false)
const previewDialogVisible = ref(false)
const previewPrompt = ref<Prompt | null>(null)
const editingPromptId = ref<number | null>(null)
const fileInputRef = ref<HTMLInputElement | null>(null)
const importMode = ref<ImportMode>('backup')
const dialogSize = ref<'small' | 'medium' | 'large'>('medium')
const newSubcategory = ref('')
const previousFieldNames = ref<Record<number, string>>({})

const dialogWidthMap = {
  small: '900px',
  medium: '1200px',
  large: '1500px',
} as const

const dialogHeightMap = {
  small: '500px',
  medium: '600px',
  large: '700px',
} as const

const createDefaultForm = (category = DEFAULT_CATEGORY): PromptFormState => ({
  name: '',
  description: '',
  content: '',
  category,
  order_num: 0,
  subcategoriesText: '',
  fields: [],
})

const form = reactive<PromptFormState>(createDefaultForm())
const formSubcategories = ref<string[]>([])

const sortPrompts = (left: Prompt, right: Prompt) => {
  if (left.order_num !== right.order_num) {
    return left.order_num - right.order_num
  }

  return left.id - right.id
}

const now = () => new Date().toISOString()

const normalizeCategory = (value?: string) => {
  const trimmed = value?.trim() ?? ''
  if (!trimmed) return UNCATEGORIZED_CATEGORY
  if (trimmed === DEFAULT_CATEGORY) return DEFAULT_CATEGORY
  if (trimmed === UNCATEGORIZED_CATEGORY) return UNCATEGORIZED_CATEGORY
  return trimmed
}

const normalizeField = (field: Partial<PromptField>): PromptField => ({
  name: String(field.name ?? '').trim(),
  label: String(field.label ?? field.name ?? '').trim(),
  type: field.type === 'textarea' || field.type === 'select' ? field.type : 'text',
  options: Array.isArray(field.options)
    ? field.options.map(item => String(item).trim()).filter(Boolean)
    : [],
  description: String(field.description ?? '').trim(),
  required: Boolean(field.required),
})

const normalizePrompt = (prompt: Partial<Prompt>, index: number): Prompt => ({
  id: Number(prompt.id) || index + 1,
  name: String(prompt.name ?? '').trim() || `未命名提示词 ${index + 1}`,
  description: String(prompt.description ?? '').trim(),
  content: String(prompt.content ?? '').trim(),
  category: normalizeCategory(prompt.category),
  order_num: Number.isFinite(prompt.order_num) ? Number(prompt.order_num) : index,
  fields: Array.isArray(prompt.fields) ? prompt.fields.map(normalizeField).filter(field => field.name) : [],
  subcategories: Array.isArray(prompt.subcategories)
    ? prompt.subcategories.map(item => String(item).trim()).filter(Boolean)
    : [],
  created_at: prompt.created_at || now(),
  updated_at: prompt.updated_at || prompt.created_at || now(),
})

const normalizeCategoryList = (value: unknown) => {
  if (!Array.isArray(value)) return []

  return Array.from(
    new Set(
      value
        .map(item => String(item).trim())
        .filter(item => item && item !== DEFAULT_CATEGORY && item !== UNCATEGORIZED_CATEGORY),
    ),
  ).sort((left, right) => left.localeCompare(right, 'zh-CN'))
}

const loadStoredPrompts = () => {
  const raw = localStorage.getItem(`${LOCAL_STORAGE_PREFIX}_prompts`)
  if (!raw) {
    prompts.value = []
    return
  }

  try {
    const parsed = JSON.parse(raw)
    prompts.value = Array.isArray(parsed) ? parsed.map(normalizePrompt).sort(sortPrompts) : []
  } catch {
    prompts.value = []
    ElMessage.warning('本地提示词数据损坏，已按空数据处理。')
  }
}

const loadStoredCategories = () => {
  const raw = localStorage.getItem(`${LOCAL_STORAGE_PREFIX}_categories`)
  if (!raw) {
    customCategories.value = []
    return
  }

  try {
    customCategories.value = normalizeCategoryList(JSON.parse(raw))
  } catch {
    customCategories.value = []
  }
}

const savePrompts = () => {
  localStorage.setItem(`${LOCAL_STORAGE_PREFIX}_prompts`, JSON.stringify(prompts.value))
}

const saveCategories = () => {
  localStorage.setItem(`${LOCAL_STORAGE_PREFIX}_categories`, JSON.stringify(customCategories.value))
}

loadStoredPrompts()
loadStoredCategories()

watch(prompts, savePrompts, { deep: true })
watch(customCategories, saveCategories, { deep: true })

const customCategoryList = computed(() => {
  const merged = new Set(customCategories.value)
  prompts.value.forEach(prompt => {
    const category = normalizeCategory(prompt.category)
    if (category !== DEFAULT_CATEGORY && category !== UNCATEGORIZED_CATEGORY) {
      merged.add(category)
    }
  })

  return Array.from(merged).sort((left, right) => left.localeCompare(right, 'zh-CN'))
})

const defaultPrompts = computed(() =>
  prompts.value.filter(prompt => normalizeCategory(prompt.category) === DEFAULT_CATEGORY).sort(sortPrompts),
)

const categorySections = computed(() =>
  [DEFAULT_CATEGORY, ...customCategoryList.value].map(name => ({
    name,
    prompts: prompts.value.filter(prompt => normalizeCategory(prompt.category) === name).sort(sortPrompts),
  })),
)

const uncategorizedPrompts = computed(() =>
  prompts.value.filter(prompt => normalizeCategory(prompt.category) === UNCATEGORIZED_CATEGORY).sort(sortPrompts),
)

const categoryOptions = computed(() => [DEFAULT_CATEGORY, ...customCategoryList.value, UNCATEGORIZED_CATEGORY])

const applyForm = (nextState: PromptFormState) => {
  form.name = nextState.name
  form.description = nextState.description
  form.content = nextState.content
  form.category = nextState.category
  form.order_num = nextState.order_num
  form.subcategoriesText = nextState.subcategoriesText
  form.fields = nextState.fields.map(field => normalizeField(field))
  formSubcategories.value = parseSubcategories(nextState.subcategoriesText)
  newSubcategory.value = ''
}

const nextPromptId = () => prompts.value.reduce((max, item) => Math.max(max, item.id), 0) + 1

const goBack = () => router.back()

const openCategoryDialog = () => {
  categoryName.value = ''
  categoryDialogVisible.value = true
}

const createCategory = () => {
  const name = categoryName.value.trim()
  if (!name) {
    ElMessage.warning('请输入分类名称。')
    return
  }

  if (name === DEFAULT_CATEGORY || name === UNCATEGORIZED_CATEGORY) {
    ElMessage.warning('该名称为系统保留分类。')
    return
  }

  if (categoryOptions.value.includes(name)) {
    ElMessage.warning('分类已存在。')
    return
  }

  customCategories.value = [...customCategoryList.value, name]
  categoryDialogVisible.value = false
  ElMessage.success('分类已创建。')
}

const openPromptDialog = (category = DEFAULT_CATEGORY, prompt?: Prompt) => {
  if (!prompt) {
    editingPromptId.value = null
    applyForm(createDefaultForm(category))
  } else {
    editingPromptId.value = prompt.id
    applyForm({
      name: prompt.name,
      description: prompt.description ?? '',
      content: prompt.content,
      category: normalizeCategory(prompt.category),
      order_num: prompt.order_num,
      subcategoriesText: (prompt.subcategories ?? []).join('\n'),
      fields: (prompt.fields ?? []).map(field => normalizeField(field)),
    })
  }

  promptDialogVisible.value = true
}

const addField = () => {
  form.fields.push({
    name: `field_${form.fields.length + 1}`,
    label: `字段 ${form.fields.length + 1}`,
    type: 'text',
    options: [],
    description: '',
    required: false,
  })
}

const removeField = (index: number) => {
  form.fields.splice(index, 1)
}

const moveField = (fromIndex: number, toIndex: number) => {
  if (fromIndex < 0 || toIndex < 0 || fromIndex >= form.fields.length || toIndex >= form.fields.length) {
    return
  }

  const [movedField] = form.fields.splice(fromIndex, 1)
  if (movedField) {
    form.fields.splice(toIndex, 0, movedField)
  }
}

const updateFieldOptions = (index: number, value: string) => {
  const field = form.fields[index]
  if (!field) return

  field.options = value.split(/[\n,]/).map(item => item.trim()).filter(Boolean)
}

const extractFieldsFromContent = () => {
  const matches = Array.from(form.content.matchAll(/\{\{([^{}]+)\}\}/g))
  if (matches.length === 0) {
    ElMessage.info('内容中没有检测到占位符。')
    return
  }

  const existing = new Set(form.fields.map(field => field.name))
  matches.forEach(match => {
    const fieldName = match[1]?.trim()
    if (fieldName && !existing.has(fieldName)) {
      existing.add(fieldName)
      form.fields.push({
        name: fieldName,
        label: fieldName,
        type: 'text',
        options: [],
        description: '',
        required: false,
      })
    }
  })

  ElMessage.success('已将占位符同步到字段配置。')
}

const parseSubcategories = (value: string) =>
  Array.from(new Set(value.split(/[\n,]/).map(item => item.trim()).filter(Boolean)))

const syncSubcategoriesText = () => {
  form.subcategoriesText = formSubcategories.value.join('\n')
}

const addFormSubcategory = () => {
  const value = newSubcategory.value.trim()
  if (!value) return
  if (formSubcategories.value.includes(value)) {
    ElMessage.warning('标签已存在。')
    return
  }

  formSubcategories.value.push(value)
  syncSubcategoriesText()
  newSubcategory.value = ''
}

const removeFormSubcategory = (index: number) => {
  formSubcategories.value.splice(index, 1)
  syncSubcategoriesText()
}

const rememberFieldName = (index: number, value: string) => {
  previousFieldNames.value[index] = value
}

const updateFieldName = (index: number, newName: string) => {
  const field = form.fields[index]
  if (!field) return

  const trimmedName = newName.trim()
  if (!trimmedName) {
    ElMessage.warning('字段名称不能为空。')
    return
  }

  const oldName = previousFieldNames.value[index] ?? field.name
  const existingNames = form.fields
    .filter((_, i) => i !== index)
    .map(item => item.name.trim())

  if (existingNames.includes(trimmedName)) {
    ElMessage.warning('字段名称不能重复。')
    field.name = oldName
    return
  }

  field.name = trimmedName
  if (!field.label.trim() || field.label === oldName) {
    field.label = trimmedName
  }

  if (oldName && oldName !== trimmedName) {
    form.content = form.content.split(`{{${oldName}}}`).join(`{{${trimmedName}}}`)
  }

  previousFieldNames.value[index] = trimmedName
}

const copyFieldName = async (fieldName: string) => {
  if (!fieldName.trim()) {
    ElMessage.warning('字段名称不能为空。')
    return
  }

  try {
    await navigator.clipboard.writeText(`{{${fieldName}}}`)
    ElMessage.success('占位符已复制。')
  } catch {
    ElMessage.error('复制失败。')
  }
}

const getFieldOptionsText = (field: PromptField) => field.options.join('\n')

const savePrompt = () => {
  const name = form.name.trim()
  const content = form.content.trim()
  const category = normalizeCategory(form.category)

  if (!name || !content) {
    ElMessage.warning('名称和内容不能为空。')
    return
  }

  const fields = form.fields.map(field => normalizeField(field)).filter(field => field.name)
  const fieldNames = fields.map(field => field.name)
  if (new Set(fieldNames).size !== fieldNames.length) {
    ElMessage.warning('字段名不能重复。')
    return
  }

  const existingPrompt = prompts.value.find(item => item.id === editingPromptId.value)
  const payload: Prompt = {
    id: editingPromptId.value ?? nextPromptId(),
    name,
    description: form.description.trim(),
    content,
    category,
    order_num: form.order_num,
    fields,
    subcategories: [...formSubcategories.value],
    created_at: existingPrompt?.created_at || now(),
    updated_at: now(),
  }

  if (editingPromptId.value === null) {
    prompts.value = [...prompts.value, payload].sort(sortPrompts)
  } else {
    prompts.value = prompts.value.map(item => (item.id === payload.id ? payload : item)).sort(sortPrompts)
  }

  if (category !== DEFAULT_CATEGORY && category !== UNCATEGORIZED_CATEGORY) {
    customCategories.value = normalizeCategoryList([...customCategoryList.value, category])
  }

  promptDialogVisible.value = false
  ElMessage.success('提示词已保存。')
}

const deletePrompt = async (prompt: Prompt) => {
  try {
    await ElMessageBox.confirm(`确认删除提示词“${prompt.name}”吗？`, '删除提示词', { type: 'warning' })
    prompts.value = prompts.value.filter(item => item.id !== prompt.id)
    ElMessage.success('提示词已删除。')
  } catch {
    return
  }
}

const deleteCategory = async (category: string) => {
  const relatedPrompts = prompts.value.filter(prompt => normalizeCategory(prompt.category) === category)

  try {
    if (relatedPrompts.length > 0) {
      await ElMessageBox.confirm(
        `分类“${category}”下仍有 ${relatedPrompts.length} 条提示词。删除后这些提示词将移动到“${UNCATEGORIZED_CATEGORY}”，是否继续？`,
        '删除分类',
        { type: 'warning' },
      )

      prompts.value = prompts.value.map(prompt =>
        normalizeCategory(prompt.category) === category
          ? { ...prompt, category: UNCATEGORIZED_CATEGORY, updated_at: now() }
          : prompt,
      )
    } else {
      await ElMessageBox.confirm(`确认删除分类“${category}”吗？`, '删除分类', { type: 'warning' })
    }

    customCategories.value = customCategoryList.value.filter(item => item !== category)
    ElMessage.success('分类已删除。')
  } catch {
    return
  }
}

const handleDragStart = (promptId: number) => {
  draggedPromptId.value = promptId
}

const handleDragEnd = () => {
  draggedPromptId.value = null
}

const handleDrop = (category: string) => {
  if (draggedPromptId.value === null) return

  prompts.value = prompts.value.map(prompt =>
    prompt.id === draggedPromptId.value
      ? { ...prompt, category: normalizeCategory(category), updated_at: now() }
      : prompt,
  )
  draggedPromptId.value = null
}

const openPreview = (prompt: Prompt) => {
  previewPrompt.value = prompt
  previewDialogVisible.value = true
}

const downloadJson = (filename: string, payload: unknown) => {
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

const handleExportCommand = (command: 'backup' | 'standalone' | 'pack') => {
  if (command === 'backup') {
    downloadJson('拆书提示词-完整备份.json', {
      type: 'book-analysis-prompts-backup',
      version: 1,
      exportTime: now(),
      categories: customCategoryList.value,
      prompts: prompts.value,
    })
    return
  }

  if (command === 'standalone') {
    downloadJson('拆书提示词-未分类.json', {
      type: 'standalone-cards',
      prompts: uncategorizedPrompts.value,
    })
    return
  }

  downloadJson('拆书提示词-分类卡包.json', {
    type: 'card-pack',
    packName: 'book-analysis-prompts',
    categories: customCategoryList.value,
    prompts: prompts.value.filter(prompt => normalizeCategory(prompt.category) !== UNCATEGORIZED_CATEGORY),
  })
}

const handleImportCommand = (command: ImportMode) => {
  importMode.value = command
  fileInputRef.value?.click()
}

const appendImportedPrompts = (incomingPrompts: Prompt[], incomingCategories: string[]) => {
  let currentId = nextPromptId()
  const appended = incomingPrompts.map(prompt => ({
    ...prompt,
    id: currentId++,
    created_at: prompt.created_at || now(),
    updated_at: now(),
  }))

  prompts.value = [...prompts.value, ...appended].sort(sortPrompts)
  customCategories.value = normalizeCategoryList([...customCategoryList.value, ...incomingCategories])
}

const importBackup = (data: { categories?: unknown; prompts?: unknown }) => {
  prompts.value = Array.isArray(data.prompts) ? data.prompts.map(normalizePrompt).sort(sortPrompts) : []
  customCategories.value = normalizeCategoryList(data.categories)
  ElMessage.success('备份已恢复。')
}

const resolveImportedPayload = (data: any, mode: ImportMode) => {
  if (mode === 'backup' && data?.type === 'book-analysis-prompts-backup') {
    return {
      replace: true,
      prompts: Array.isArray(data.prompts) ? data.prompts.map(normalizePrompt) : [],
      categories: normalizeCategoryList(data.categories),
    }
  }

  if (data?.type === 'standalone-cards') {
    return {
      replace: false,
      prompts: (Array.isArray(data.prompts) ? data.prompts : []).map((item: Partial<Prompt>, index: number) =>
        normalizePrompt({ ...item, category: UNCATEGORIZED_CATEGORY }, index),
      ),
      categories: [] as string[],
    }
  }

  if (data?.type === 'card-pack') {
    const fallbackCategory =
      typeof data.packName === 'string' && data.packName.trim() && data.packName !== 'book-analysis-prompts'
        ? data.packName.trim()
        : DEFAULT_CATEGORY

    return {
      replace: false,
      prompts: (Array.isArray(data.prompts) ? data.prompts : []).map((item: Partial<Prompt>, index: number) =>
        normalizePrompt({ ...item, category: item.category || fallbackCategory }, index),
      ),
      categories: normalizeCategoryList([...(Array.isArray(data.categories) ? data.categories : []), fallbackCategory]),
    }
  }

  if (Array.isArray(data)) {
    const fallbackCategory = mode === 'standalone' ? UNCATEGORIZED_CATEGORY : DEFAULT_CATEGORY
    return {
      replace: false,
      prompts: data.map((item, index) => normalizePrompt({ ...item, category: item.category || fallbackCategory }, index)),
      categories: [] as string[],
    }
  }

  if (Array.isArray(data?.prompts)) {
    return {
      replace: false,
      prompts: data.prompts.map((item: Partial<Prompt>, index: number) => normalizePrompt(item, index)),
      categories: normalizeCategoryList(data.categories),
    }
  }

  throw new Error('不支持的导入文件格式。')
}

const handleFileImport = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const parsed = JSON.parse(text)
    const payload = resolveImportedPayload(parsed, importMode.value)

    if (payload.replace) {
      importBackup({ prompts: payload.prompts, categories: payload.categories })
    } else {
      appendImportedPrompts(payload.prompts, payload.categories)
      ElMessage.success('导入完成。')
    }
  } catch (error) {
    console.error(error)
    ElMessage.error('导入失败，请检查 JSON 文件内容。')
  } finally {
    input.value = ''
  }
}

const resetStorage = async () => {
  try {
    await ElMessageBox.confirm(
      '确认只重置当前页面的本地数据吗？这会清空拆书提示词库的 localStorage 数据。',
      '重置数据',
      { type: 'warning' },
    )
    prompts.value = []
    customCategories.value = []
    ElMessage.success('页面数据已清空。')
  } catch {
    return
  }
}

const formatDate = (value?: string) => {
  if (!value) return '刚刚'

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value

  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const formatFieldType = (type: PromptField['type']) => {
  if (type === 'textarea') return '多行文本'
  if (type === 'select') return '下拉选择'
  return '单行文本'
}
</script>

<style scoped>
.book-analysis-prompts-page {
  height: 100%;
  min-height: 100%;
  padding: 24px;
  background: #f6f8fb;
  overflow-y: auto;
  overflow-x: hidden;
  box-sizing: border-box;
}

.page-header,
.summary-card {
  background: #fff;
  border: 1px solid #e7ecf3;
  border-radius: 20px;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.06);
}

.page-header {
  padding: 24px;
  margin-bottom: 20px;
}

.header-main,
.prompt-meta,
.prompt-actions,
.field-panel-header,
.preview-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.header-main h1,
.preview-header-row h3,
.preview-fields h4 {
  margin: 0;
}

.header-main p,
.preview-header-row p {
  margin: 6px 0 0;
  color: #64748b;
}

.header-actions,
.field-panel-actions,
.tag-list {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.summary-grid,
.dialog-grid {
  display: grid;
  gap: 16px;
}

.summary-grid {
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  margin-bottom: 20px;
}

.summary-card {
  padding: 18px 20px;
}

.summary-label {
  display: block;
  margin-bottom: 8px;
  color: #64748b;
}

.summary-card strong {
  font-size: 28px;
  color: #0f172a;
}

.preview-content {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.7;
  color: #334155;
  background: #f8fafc;
  border-radius: 14px;
  padding: 14px;
  margin: 14px 0;
}

.prompt-meta {
  font-size: 12px;
  color: #94a3b8;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 10px 0 16px;
  padding: 0 4px;
  color: #1f2937;
  font-weight: 700;
}

.standalone-title {
  margin-top: 30px;
}

.section-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  background: #e0f2fe;
  color: #0369a1;
  font-size: 13px;
}

.packs-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 20px;
  align-items: flex-start;
}

.card-pack-vertical {
  width: 240px;
  height: 360px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-radius: 16px;
  border: 2px solid #bfdbfe;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(59, 130, 246, 0.12);
}

.pack-header-vertical {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px 14px 14px;
  background: linear-gradient(135deg, #2563eb 0%, #1d4ed8 100%);
}

.pack-header-left {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
}

.pack-icon-vertical {
  font-size: 22px;
  color: #fff;
}

.pack-name-vertical {
  color: #fff;
  font-size: 15px;
  font-weight: 700;
}

.pack-count-badge {
  position: absolute;
  top: 12px;
  right: 12px;
  min-width: 26px;
  padding: 2px 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  color: #1d4ed8;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
}

.pack-header-actions,
.pack-card-actions,
.pack-footer-vertical,
.standalone-card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.cards-scroll-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: 18px 14px 12px;
  overflow-y: auto;
  perspective: 1200px;
}

.pack-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 140px;
  margin-top: -46px;
  padding: 16px 14px 12px;
  border-radius: 16px;
  border: 1px solid rgba(147, 197, 253, 0.72);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(239, 246, 255, 0.98) 100%);
  cursor: grab;
  box-shadow:
    0 10px 20px rgba(37, 99, 235, 0.10),
    0 2px 0 rgba(255, 255, 255, 0.85) inset;
  transition:
    transform 220ms ease,
    box-shadow 220ms ease,
    border-color 220ms ease,
    background 220ms ease;
  transform-origin: center bottom;
  overflow: visible;
  isolation: isolate;
}

.pack-card:first-child {
  margin-top: 0;
}

.pack-card:nth-child(odd) {
  transform: rotate(-1.1deg);
}

.pack-card:nth-child(even) {
  transform: rotate(1deg);
}

.pack-card::before,
.pack-card::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  z-index: -2;
  pointer-events: none;
}

.pack-card::before {
  inset: 7px 6px -7px;
  background: rgba(219, 234, 254, 0.9);
  border: 1px solid rgba(147, 197, 253, 0.52);
  box-shadow: 0 8px 16px rgba(30, 64, 175, 0.08);
}

.pack-card::after {
  inset: 13px 12px -12px;
  background: rgba(191, 219, 254, 0.72);
  border: 1px solid rgba(147, 197, 253, 0.32);
  z-index: -3;
}

.pack-card-edge,
.standalone-card-edge {
  position: absolute;
  top: 10px;
  right: 12px;
  width: 54px;
  height: 10px;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.25), rgba(148, 163, 184, 0.6), rgba(255, 255, 255, 0.25));
  opacity: 0.45;
  pointer-events: none;
}

.pack-card:hover,
.standalone-card:hover {
  z-index: 8;
}

.pack-card:hover {
  transform: translateY(-18px) scale(1.02) rotate(0deg);
  border-color: rgba(59, 130, 246, 0.82);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 1) 0%, rgba(239, 246, 255, 1) 100%);
  box-shadow:
    0 22px 40px rgba(30, 64, 175, 0.18),
    0 8px 18px rgba(59, 130, 246, 0.16),
    0 2px 0 rgba(255, 255, 255, 0.95) inset;
}

.pack-card-header,
.pack-card-footer,
.standalone-card-header,
.standalone-card-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.pack-card-name,
.standalone-card-name {
  font-weight: 700;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pack-card-preview,
.standalone-card-preview {
  color: #64748b;
  font-size: 12px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.pack-card-preview {
  -webkit-line-clamp: 3;
  line-clamp: 3;
  min-height: 58px;
}

.standalone-card-preview {
  -webkit-line-clamp: 4;
  line-clamp: 4;
}

.pack-card-hover-info,
.standalone-card-hover-info {
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: calc(100% - 14px);
  padding: 10px 12px;
  border-radius: 14px;
  color: #eff6ff;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.96), rgba(30, 41, 59, 0.92));
  box-shadow: 0 18px 30px rgba(15, 23, 42, 0.26);
  opacity: 0;
  transform: translateY(10px) scale(0.96);
  transition: opacity 180ms ease, transform 180ms ease;
  pointer-events: none;
}

.pack-card:hover .pack-card-hover-info,
.standalone-card:hover .standalone-card-hover-info {
  opacity: 1;
  transform: translateY(0) scale(1);
}

.pack-card-hover-label,
.standalone-card-hover-label {
  display: inline-flex;
  margin-bottom: 6px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #93c5fd;
}

.pack-card-hover-info p,
.standalone-card-hover-info p {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
}

.pack-card-time,
.standalone-card-time {
  font-size: 11px;
  color: #94a3b8;
}

.empty-pack-hint,
.empty-standalone-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: #3b82f6;
  text-align: center;
  cursor: pointer;
}

.empty-pack-hint {
  min-height: 120px;
  border: 1px dashed #93c5fd;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.8);
}

.pack-footer-vertical {
  justify-content: space-between;
  padding: 10px 12px 12px;
  border-top: 1px solid #dbeafe;
  background: rgba(255, 255, 255, 0.9);
}

.standalone-cards-area {
  padding: 20px;
  border: 2px dashed #fdba74;
  border-radius: 16px;
  background: linear-gradient(180deg, #fff7ed 0%, #ffffff 100%);
}

.standalone-cards-grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 16px;
}

.standalone-card {
  position: relative;
  display: flex;
  flex-direction: column;
  min-height: 210px;
  padding: 16px 14px 14px;
  border-radius: 16px;
  border: 1px solid rgba(253, 186, 116, 0.72);
  background: linear-gradient(180deg, #ffffff 0%, #fff7ed 100%);
  cursor: grab;
  box-shadow:
    0 10px 20px rgba(249, 115, 22, 0.12),
    0 2px 0 rgba(255, 255, 255, 0.8) inset;
  transition:
    transform 220ms ease,
    box-shadow 220ms ease,
    border-color 220ms ease;
  overflow: visible;
}

.standalone-card::before {
  content: "";
  position: absolute;
  inset: 8px 10px -8px;
  border-radius: inherit;
  background: rgba(255, 237, 213, 0.72);
  border: 1px solid rgba(253, 186, 116, 0.4);
  z-index: -1;
  pointer-events: none;
}

.standalone-card-content {
  flex: 1;
  margin: 12px 0;
}

.standalone-card:hover {
  transform: translateY(-10px) scale(1.015);
  border-color: rgba(249, 115, 22, 0.78);
  box-shadow:
    0 18px 34px rgba(194, 65, 12, 0.16),
    0 8px 18px rgba(251, 146, 60, 0.14);
}

.empty-standalone-hint {
  min-height: 180px;
  padding: 32px 20px;
  border: 2px dashed #fdba74;
  border-radius: 16px;
  background: linear-gradient(180deg, #fff7ed 0%, #ffffff 100%);
}

.dialog-grid {
  grid-template-columns: minmax(0, 1.3fr) minmax(280px, 0.9fr);
}

.dialog-side {
  display: grid;
  gap: 16px;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.dialog-header span {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.size-controls {
  display: flex;
  align-items: center;
}

.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
  padding-right: 4px;
}

.left-panel {
  flex: none;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: visible;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.right-panel {
  flex: none;
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding-top: 4px;
  overflow: visible;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid #f0f2f5;
}

.panel-header .form-label {
  margin-bottom: 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.field-config-section {
  flex: none;
  overflow: visible;
}

.field-panel {
  border: 1px solid #e7ecf3;
  border-radius: 16px;
  padding: 16px;
}

.field-empty {
  color: #64748b;
  font-size: 13px;
}

.field-list,
.preview-fields ul {
  display: grid;
  gap: 12px;
}

.empty-fields {
  padding: 24px 12px;
  text-align: center;
  color: #909399;
}

.empty-hint {
  margin: 0;
}

.field-item {
  display: grid;
  gap: 10px;
  padding: 12px;
  border-radius: 14px;
  background: #f8fafc;
}

.field-header,
.field-name-container,
.field-actions,
.subcategory-input-row,
.subcategory-tags-row,
.form-item-label-with-guide {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.field-header {
  justify-content: space-between;
}

.field-name-container {
  flex: 1;
}

.field-name-input {
  flex: 1;
  min-width: 180px;
}

.field-config-form {
  display: grid;
  gap: 10px;
}

.field-input {
  width: 100%;
}

.field-options label {
  display: block;
  margin-bottom: 8px;
  color: #606266;
  font-size: 13px;
}

.description-rich-editor {
  width: 100%;
}

.format-description p,
.prompt-guide-popover p {
  margin: 0 0 8px;
  line-height: 1.6;
}

.prompt-guide-title {
  font-weight: 600;
  margin-bottom: 10px;
}

.prompt-guide-example {
  padding: 10px 12px;
  border-radius: 10px;
  background: #f8fafc;
}

.prompt-guide-example-label {
  display: inline-block;
  margin-bottom: 6px;
  font-size: 12px;
  color: #909399;
}

.subcategory-form {
  display: grid;
  gap: 10px;
}

.subcategory-tag {
  margin-right: 4px;
}

.subcategory-input {
  flex: 1;
  min-width: 180px;
}

.full-width {
  width: 100%;
}

.hidden-input {
  display: none;
}

@media (max-width: 960px) {
  .book-analysis-prompts-page {
    padding: 16px;
  }

  .header-main,
  .page-header,
  .dialog-grid {
    display: grid;
  }

  .dialog-header,
  .panel-header,
  .field-header {
    display: grid;
  }

  .packs-wrapper {
    display: grid;
    grid-template-columns: 1fr;
  }

  .card-pack-vertical {
    width: 100%;
    height: 340px;
  }

  .standalone-cards-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .dialog-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .standalone-cards-grid {
    grid-template-columns: 1fr;
  }
}

/* 暗色主题适配 */
:root[data-theme='dark'] .book-analysis-prompts-page {
  background: #0f172a;
}

:root[data-theme='dark'] .page-header,
:root[data-theme='dark'] .summary-card {
  background: #1e293b;
  border-color: rgba(71, 85, 105, 0.4);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .header-main h1 {
  color: #f3f4f6;
}

:root[data-theme='dark'] .header-main p {
  color: #9ca3af;
}

:root[data-theme='dark'] .summary-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .summary-card strong {
  color: #f3f4f6;
}

:root[data-theme='dark'] .preview-content {
  color: #d1d5db;
  background: rgba(51, 65, 85, 0.6);
}

:root[data-theme='dark'] .prompt-meta {
  color: #6b7280;
}

:root[data-theme='dark'] .card-pack-vertical {
  border-color: rgba(59, 130, 246, 0.4);
  background: linear-gradient(180deg, rgba(30, 58, 138, 0.3) 0%, #1e293b 100%);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .pack-card {
  border-color: rgba(71, 85, 105, 0.4);
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.98) 0%, rgba(30, 41, 59, 0.95) 100%);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
  transform: none !important;
}

:root[data-theme='dark'] .pack-card:nth-child(odd),
:root[data-theme='dark'] .pack-card:nth-child(even) {
  transform: none !important;
}

:root[data-theme='dark'] .pack-card::before,
:root[data-theme='dark'] .pack-card::after {
  display: none;
}

:root[data-theme='dark'] .pack-card:hover {
  border-color: rgba(59, 130, 246, 0.6);
  background: linear-gradient(180deg, rgba(30, 41, 59, 1) 0%, rgba(30, 58, 138, 0.3) 100%);
  transform: translateY(-4px) !important;
}

:root[data-theme='dark'] .pack-card-name,
:root[data-theme='dark'] .standalone-card-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .pack-card-preview,
:root[data-theme='dark'] .standalone-card-preview {
  color: #9ca3af;
}

:root[data-theme='dark'] .pack-card-time,
:root[data-theme='dark'] .standalone-card-time {
  color: #6b7280;
}

:root[data-theme='dark'] .pack-footer-vertical {
  border-top-color: rgba(71, 85, 105, 0.4);
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .standalone-cards-area {
  border-color: rgba(249, 115, 22, 0.4);
  background: linear-gradient(180deg, rgba(124, 45, 18, 0.2) 0%, #1e293b 100%);
}

:root[data-theme='dark'] .standalone-card {
  border-color: rgba(71, 85, 105, 0.4);
  background: linear-gradient(180deg, #1e293b 0%, rgba(30, 41, 59, 0.95) 100%);
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.2);
}

:root[data-theme='dark'] .standalone-card::before {
  display: none;
}

:root[data-theme='dark'] .standalone-card:hover {
  border-color: rgba(249, 115, 22, 0.5);
}

:root[data-theme='dark'] .empty-pack-hint,
:root[data-theme='dark'] .empty-standalone-hint {
  color: #60a5fa;
  border-color: rgba(59, 130, 246, 0.4);
  background: rgba(30, 41, 59, 0.6);
}

:root[data-theme='dark'] .dialog-header span {
  color: #f3f4f6;
}

:root[data-theme='dark'] .pack-card-hover-info,
:root[data-theme='dark'] .standalone-card-hover-info {
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(30, 41, 59, 0.95));
}
</style>
