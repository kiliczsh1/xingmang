// 提示词「动态字段 + 内容编辑」的共用逻辑（Prompts.vue 与 PackDetail.vue 共用）
//
// 设计取舍：**状态由调用方（视图）持有**，以 ref 注入；本模块只沉淀逻辑。
// 原因是这些状态直接绑在模板上（`ref="contentTextareaRef"`、`v-model="fieldExpanded"` 等），
// 把状态搬进 composable 会连带改动模板绑定，收益小、风险大。
//
// 注入的形参名与视图里原来的变量名**逐字相同**，因此函数体可以原样搬运 ——
// 等价性由 scripts/_verify_extract_equiv.py 逐字符证明。

import { computed, nextTick, type Ref } from 'vue'
import { ElMessage } from 'element-plus'

/** 动态字段配置项 */
export interface FieldConfig {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  optionLabels: string[]
  optionsText: string
  optionLabelsText: string
  description: string
  required: boolean
  defaultValue: string
}

/** 预览测试里的字段填写项 */
export interface PreviewField {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  value: string
}

export function usePromptFields<T extends { content: string }>(deps: {
  formData: Ref<T>
  fieldsConfig: Ref<FieldConfig[]>
  fieldExpanded: Ref<number | null>
  previewFields: Ref<PreviewField[]>
  previewTestVisible: Ref<boolean>
  insertFieldPopoverVisible: Ref<boolean>
  enlargeInsertPopoverVisible: Ref<boolean>
  customFieldName: Ref<string>
  contentTextareaRef: Ref<{ $el: HTMLDivElement } | null>
  formSubcategories: Ref<string[]>
  newSubcategory: Ref<string>
}) {
  const { formData, fieldsConfig, fieldExpanded, previewFields, previewTestVisible, insertFieldPopoverVisible, enlargeInsertPopoverVisible, customFieldName, contentTextareaRef, formSubcategories, newSubcategory } = deps

  // 字段类型标签映射
  const fieldTypeLabel = (type: string): string => {
    const map: Record<string, string> = { text: '单行文本', textarea: '多行文本', select: '下拉选择' }
    return map[type] || type
  }

  // 字段卡片展开/收起
  const toggleFieldExpand = (index: number) => {
    fieldExpanded.value = fieldExpanded.value === index ? null : index
  }

  // 预览并保存字段配置（收起卡片）
  const saveFieldExpand = (index: number) => {
    const field = fieldsConfig.value[index]
    if (!field) return
    if (!field.name.trim()) {
      ElMessage.warning('字段名称不能为空')
      return
    }
    fieldExpanded.value = null
    ElMessage.success(`字段「${field.name}」已保存`)
  }

  // 获取 textarea 原生元素（el-input ref 的 $el 是包裹 div，内部有 textarea）
  function getTextareaEl(refItem: { $el: HTMLDivElement } | null): HTMLTextAreaElement | null {
    if (!refItem?.$el) return null
    return refItem.$el.querySelector('textarea')
  }

  // 在指定 textarea 的光标位置插入 ${字段名}
  function insertFieldToContent(fieldName: string, targetRef?: { $el: HTMLDivElement } | null) {
    const name = (fieldName || '').trim()
    if (!name) return
    const placeholder = `\${${name}}`
    const ta = getTextareaEl(targetRef ?? contentTextareaRef.value) ?? getTextareaEl(contentTextareaRef.value)
  
    if (ta && ta.selectionStart !== undefined && ta.selectionStart !== null) {
      const start = ta.selectionStart
      const end = ta.selectionEnd
      const content = formData.value.content || ''
      formData.value.content = content.slice(0, start) + placeholder + content.slice(end)
      // 恢复焦点与光标位置（插入点之后）
      nextTick(() => {
        ta.focus()
        const pos = start + placeholder.length
        ta.setSelectionRange(pos, pos)
      })
    } else {
      formData.value.content = (formData.value.content || '') + placeholder
    }
    insertFieldPopoverVisible.value = false
    enlargeInsertPopoverVisible.value = false
    customFieldName.value = ''
  }

  // 插入自定义字段
  function insertCustomField(targetRef?: { $el: HTMLDivElement } | null) {
    const name = customFieldName.value.trim()
    if (!name) {
      ElMessage.warning('请输入字段名称')
      return
    }
    if (/\$\{/.test(name) || /\}/.test(name)) {
      ElMessage.warning('字段名称不能包含 ${} 符号')
      return
    }
    insertFieldToContent(name, targetRef)
  }

  // 粘贴导入：读取剪贴板文本填入正文
  const handlePasteImport = async () => {
    try {
      const text = await navigator.clipboard.readText()
      if (!text || !text.trim()) {
        ElMessage.warning('剪贴板中没有可导入的文本')
        return
      }
      // 追加到正文末尾（保留已有内容）
      formData.value.content = formData.value.content
        ? formData.value.content + '\n' + text
        : text
      ElMessage.success('已从剪贴板导入')
    } catch (e) {
      ElMessage.error('读取剪贴板失败，请检查浏览器权限')
    }
  }

  // 打开预览测试：初始化字段表单
  const openPreviewTest = () => {
    previewFields.value = fieldsConfig.value.map(field => ({
      name: field.name,
      label: field.label || field.name,
      type: field.type,
      options: field.options || [],
      value: field.defaultValue || ''
    }))
    previewTestVisible.value = true
  }

  // 复制预览结果（已填字段替换后的完整提示词）
  const copyPreviewResult = async () => {
    const filled = previewFields.value.filter(f => f.value.trim()).map(f => ({
      name: f.name,
      value: f.value.trim()
    }))
    let result = formData.value.content || ''
    filled.forEach(f => {
      result = result.replace(new RegExp(`\\$\\{${f.name}\\}`, 'g'), f.value)
    })
    try {
      await navigator.clipboard.writeText(result)
      ElMessage.success('已复制')
    } catch (e) {
      ElMessage.error('复制失败')
    }
  }

  // 添加小分类到表单
  const addFormSubcategory = () => {
    const value = newSubcategory.value.trim()
    if (value && !formSubcategories.value.includes(value)) {
      formSubcategories.value.push(value)
      newSubcategory.value = ''
    }
  }

  // 从表单中删除小分类
  const removeFormSubcategory = (index: number) => {
    formSubcategories.value.splice(index, 1)
  }

  // 从提示词内容中提取字段名称（幂等：只追加不存在的字段）
  const extractFieldsFromContent = (content: string) => {
    if (!content) {
      ElMessage.info('正文为空，无可提取的字段')
      return
    }
    const fieldRegex = /\$\{([^}]+)\}/g
    const fields = new Set<string>()
    let match
    while ((match = fieldRegex.exec(content)) !== null) {
      if (match && match[1]) {
        fields.add(match[1].trim())
      }
    }
    
    if (fields.size === 0) {
      ElMessage.info('正文中没有找到 ${字段名} 占位符')
      return
    }
    
    // 收集现有字段名称，避免重复添加
    const existingNames = new Set(fieldsConfig.value.map(f => f.name))
    let addedCount = 0
    
    // 为每个新字段创建配置
    fields.forEach(fieldName => {
      if (!existingNames.has(fieldName)) {
        fieldsConfig.value.push({
          name: fieldName,
          label: fieldName,
          type: 'text',
          options: [],
          optionLabels: [],
          optionsText: '',
          optionLabelsText: '',
          description: '',
          required: true,
          defaultValue: ''
        })
        addedCount++
      }
    })
    
    if (addedCount > 0) {
      ElMessage.success(`已提取 ${addedCount} 个字段`)
    } else {
      ElMessage.info('所有字段均已配置')
    }
  }

  // 添加字段
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
      required: true,
      defaultValue: ''
    })
    fieldExpanded.value = fieldsConfig.value.length - 1
  }

  // 删除字段
  const removeField = (index: number) => {
    fieldsConfig.value.splice(index, 1)
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

  // 预览测试：正文分段（未填=占位符橙色，已填=替换绿色）
  const previewSegments = computed(() => {
    const content = formData.value.content || ''
    const regex = /\$\{([^}]+)\}/g
    const segments: Array<{ type: 'text' | 'unfilled' | 'filled'; text: string }> = []
    let lastIndex = 0
    let match
    while ((match = regex.exec(content)) !== null) {
      if (match.index > lastIndex) {
        segments.push({ type: 'text', text: content.slice(lastIndex, match.index) })
      }
      const fieldName = match[1].trim()
      const field = previewFields.value.find(f => f.name === fieldName)
      const value = field?.value?.trim() || ''
      if (value) {
        segments.push({ type: 'filled', text: value })
      } else {
        segments.push({ type: 'unfilled', text: `\${${fieldName}}` })
      }
      lastIndex = match.index + match[0].length
    }
    if (lastIndex < content.length) {
      segments.push({ type: 'text', text: content.slice(lastIndex) })
    }
    return segments
  })

  const unfilledFields = computed(() => {
    const content = formData.value.content || ''
    const regex = /\$\{([^}]+)\}/g
    const names = new Set<string>()
    let match
    while ((match = regex.exec(content)) !== null) {
      if (match && match[1]) names.add(match[1].trim())
    }
    return Array.from(names).filter(name => {
      const field = previewFields.value.find(f => f.name === name)
      return !field?.value?.trim()
    })
  })

  // 配置了但未在模板正文中引用的字段（未引用提示）
  const unusedFields = computed(() => {
    const content = formData.value.content || ''
    const regex = /\$\{([^}]+)\}/g
    const used = new Set<string>()
    let match
    while ((match = regex.exec(content)) !== null) {
      if (match && match[1]) used.add(match[1].trim())
    }
    return fieldsConfig.value.filter(f => f.name && !used.has(f.name)).map(f => f.name)
  })

  return {
    fieldTypeLabel,
    toggleFieldExpand,
    saveFieldExpand,
    getTextareaEl,
    insertFieldToContent,
    insertCustomField,
    handlePasteImport,
    openPreviewTest,
    copyPreviewResult,
    addFormSubcategory,
    removeFormSubcategory,
    extractFieldsFromContent,
    addField,
    removeField,
    validateFieldsConfig,
    previewSegments,
    unfilledFields,
    unusedFields,
  }
}
