<template>
  <el-dialog
    :model-value="visible"
    title="修改系统提示词"
    width="600px"
    destroy-on-close
    modal-class="prompt-edit-dialog-modal"
    append-to-body
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="prompt-edit-content">
      <div class="form-item">
        <label class="form-label">提示词名称</label>
        <el-input
          v-model="editingPrompt.name"
          placeholder="请输入提示词名称"
          class="full-width"
        />
      </div>
      <div class="form-item">
        <label class="form-label">提示词内容</label>
        <div class="prompt-format-hint">
          <el-alert
            title="提示词格式说明"
            type="info"
            :closable="false"
            show-icon
          >
            <template #default>
              <div class="format-description">
                <p><strong>当前使用模板式格式（方案3）：</strong></p>
                <p>角色：[生成器名称]</p>
                <p>[此处为您编辑的内容]</p>
                <p>具体要求：[选择的提示词内容]</p>
                <p class="format-tip">💡 提示：请在此输入核心角色设定和基础能力描述</p>
              </div>
            </template>
          </el-alert>

          <el-alert
            title="模板功能说明"
            type="warning"
            :closable="false"
            show-icon
            style="margin-top: 12px;"
          >
            <template #default>
              <div class="format-description">
                <p><strong>📝 写模板，留空位</strong></p>
                <p>在提示词正文中，用 <code>${字段名称}</code> 标记需要用户填写的地方</p>
                <p class="format-tip">示例：生成书名提示词</p>
                <p class="format-example">帮我生成${数量}个${小说类型}类型${平台}平台的小说书名</p>
                <p class="format-tip">📋 要求：</p>
                <ul style="margin: 8px 0; padding-left: 20px;">
                  <li>合理放置字段位置</li>
                  <li>长文本字段建议独立成段</li>
                  <li>字段名称要清晰明确</li>
                </ul>
              </div>
            </template>
          </el-alert>
        </div>
        <el-input
          v-model="editingPrompt.content"
          type="textarea"
          :rows="8"
          placeholder="请输入系统提示词内容...&#10;例如：你是一个专业的角色生成器助手，擅长创建丰富立体的虚构角色。你会根据用户的需求，生成包含性格、背景、外貌等详细信息的角色设定。"
          resize="none"
          class="full-width"
        />
      </div>
      <div class="form-item">
        <label class="form-label">提示词分类</label>
        <el-input
          v-model="editingPrompt.category"
          placeholder="请输入提示词分类"
          class="full-width"
        />
      </div>

      <!-- 字段配置 -->
      <div class="form-item">
        <label class="form-label">字段配置</label>
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

          <div v-if="fieldsConfig.length === 0" class="empty-fields">
            <el-button type="primary" @click="addField">添加字段</el-button>
            <p class="empty-hint">暂无字段配置，点击添加字段开始配置</p>
          </div>

          <div v-else class="fields-list">
            <div v-for="(field, index) in fieldsConfig" :key="index" class="field-item">
              <div class="field-header">
                <span class="field-name">${{ field.name }}</span>
                <div class="field-actions">
                  <el-button
                    type="primary"
                    link
                    size="small"
                    @click="moveField(index, index - 1)"
                    :disabled="index === 0"
                  >
                    上移
                  </el-button>
                  <el-button
                    type="primary"
                    link
                    size="small"
                    @click="moveField(index, index + 1)"
                    :disabled="index === fieldsConfig.length - 1"
                  >
                    下移
                  </el-button>
                  <el-button type="primary" link size="small" @click="addField">添加字段</el-button>
                  <el-button type="danger" link size="small" @click="removeField(index)">删除</el-button>
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
                    @change="updateFieldOptions(index)"
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
      </div>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="emit('update:visible', false)">取消</el-button>
        <el-button type="primary" @click="savePrompt">保存修改</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

interface EditingPrompt {
  id: number
  name: string
  content: string
  category: string
}

interface FieldConfig {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  optionLabels?: string[]
  optionsText: string
  description: string
}

const props = defineProps<{
  visible: boolean
  initialPrompt: EditingPrompt | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'save', prompt: EditingPrompt): void
}>()

const editingPrompt = ref<EditingPrompt>({ id: 0, name: '', content: '', category: '' })
const fieldsConfig = ref<FieldConfig[]>([])

// 每次打开时用父组件传入的初始值初始化
watch(
  () => props.visible,
  (v) => {
    if (v && props.initialPrompt) {
      editingPrompt.value = { ...props.initialPrompt }
      fieldsConfig.value = []
      extractFieldsFromContent(props.initialPrompt.content)
    }
  }
)

// 监听提示词内容变化，自动更新字段配置
watch(() => editingPrompt.value.content, (newContent) => {
  if (newContent) {
    const fieldRegex = /\$\{([^}]+)\}/g
    const newFields = new Set<string>()
    let match
    while ((match = fieldRegex.exec(newContent)) !== null) {
      if (match && match[1]) {
        newFields.add(match[1].trim())
      }
    }

    const existingFieldNames = new Set(fieldsConfig.value.map(field => field.name))

    newFields.forEach(fieldName => {
      if (!existingFieldNames.has(fieldName)) {
        fieldsConfig.value.push({
          name: fieldName,
          label: fieldName,
          type: 'text',
          options: [],
          optionsText: '',
          description: ''
        })
      }
    })

    fieldsConfig.value = fieldsConfig.value.filter(field => newFields.has(field.name))
  } else {
    fieldsConfig.value = []
  }
})

// 从提示词内容中提取字段名称
const extractFieldsFromContent = (content: string) => {
  const fieldRegex = /\$\{([^}]+)\}/g
  const fields = new Set<string>()
  let match
  while ((match = fieldRegex.exec(content)) !== null) {
    if (match && match[1]) {
      fields.add(match[1].trim())
    }
  }

  // 为每个字段创建配置
  fields.forEach(fieldName => {
    fieldsConfig.value.push({
      name: fieldName,
      label: fieldName,
      type: 'text',
      options: [],
      optionsText: '',
      description: ''
    })
  })
}

// 添加字段
const addField = () => {
  fieldsConfig.value.push({
    name: `字段${fieldsConfig.value.length + 1}`,
    label: `字段${fieldsConfig.value.length + 1}`,
    type: 'text',
    options: [],
    optionsText: '',
    description: ''
  })
}

// 删除字段
const removeField = (index: number) => {
  fieldsConfig.value.splice(index, 1)
}

// 更新字段选项
const updateFieldOptions = (index: number) => {
  const field = fieldsConfig.value[index]
  if (field) {
    const optionsText = field.optionsText
    field.options = optionsText.split('\n').filter((option: string) => option.trim())
  }
}

// 验证字段配置
const validateFieldsConfig = () => {
  // 检查是否所有字段都有名称
  for (const field of fieldsConfig.value) {
    if (!field.name.trim()) {
      return '所有字段必须有名称'
    }

    // 检查下拉选择类型的字段是否有选项
    if (field.type === 'select' && field.options.length === 0) {
      return `字段"${field.name}"是下拉选择类型，必须添加选项`
    }
  }

  // 检查字段名称是否唯一
  const fieldNames = fieldsConfig.value.map(field => field.name.trim())
  const uniqueNames = new Set(fieldNames)
  if (fieldNames.length !== uniqueNames.size) {
    return '字段名称必须唯一'
  }

  return ''
}

// 调整字段顺序
const moveField = (fromIndex: number, toIndex: number) => {
  if (fromIndex < 0 || fromIndex >= fieldsConfig.value.length || toIndex < 0 || toIndex >= fieldsConfig.value.length) return

  const [movedField] = fieldsConfig.value.splice(fromIndex, 1)
  if (movedField) {
    fieldsConfig.value.splice(toIndex, 0, movedField)
  }
}

// 保存提示词修改
const savePrompt = () => {
  const validationError = validateFieldsConfig()
  if (validationError) {
    ElMessage.error(validationError)
    return
  }

  emit('save', { ...editingPrompt.value })
}
</script>
