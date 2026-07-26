<template>
  <div class="rich-text-editor" :class="{ 'is-disabled': disabled }">
    <div class="toolbar">
      <div v-for="(group, groupIndex) in toolbarGroups" :key="groupIndex" class="toolbar-group">
        <button
          v-for="action in group"
          :key="action.key"
          type="button"
          class="toolbar-button"
          :title="action.title"
          :disabled="disabled"
          @mousedown.prevent
          @click="handleAction(action)"
        >
          {{ action.label }}
        </button>
      </div>
    </div>
    <div
      ref="editorRef"
      class="editor"
      :class="{ 'is-empty': isEmpty }"
      :contenteditable="!disabled"
      :data-placeholder="placeholder"
      @focus="handleFocus"
      @blur="handleBlur"
      @input="handleInput"
      @keyup="saveSelection"
      @mouseup="saveSelection"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'

type ToolbarAction = {
  key: string
  label: string
  title: string
  kind: 'command' | 'action'
  command?: string
  value?: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
}>(), {
  placeholder: '请输入内容',
  disabled: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const editorRef = ref<HTMLDivElement | null>(null)
const isFocused = ref(false)
const savedRange = ref<Range | null>(null)

const toolbarGroups: ToolbarAction[][] = [
  [
    { key: 'bold', label: 'B', title: '粗体', kind: 'command', command: 'bold' },
    { key: 'italic', label: 'I', title: '斜体', kind: 'command', command: 'italic' },
    { key: 'underline', label: 'U', title: '下划线', kind: 'command', command: 'underline' },
    { key: 'strike', label: 'S', title: '删除线', kind: 'command', command: 'strikeThrough' },
    { key: 'sup', label: 'x²', title: '上标', kind: 'command', command: 'superscript' },
    { key: 'sub', label: 'x₂', title: '下标', kind: 'command', command: 'subscript' }
  ],
  [
    { key: 'h1', label: 'H1', title: '一级标题', kind: 'action' },
    { key: 'h2', label: 'H2', title: '二级标题', kind: 'action' },
    { key: 'h3', label: 'H3', title: '三级标题', kind: 'action' },
    { key: 'quote', label: '“”', title: '引用', kind: 'action' },
    { key: 'code', label: '</>', title: '代码', kind: 'action' }
  ],
  [
    { key: 'left', label: '左', title: '左对齐', kind: 'command', command: 'justifyLeft' },
    { key: 'center', label: '中', title: '居中', kind: 'command', command: 'justifyCenter' },
    { key: 'right', label: '右', title: '右对齐', kind: 'command', command: 'justifyRight' },
    { key: 'ul', label: '• 列', title: '无序列表', kind: 'command', command: 'insertUnorderedList' },
    { key: 'ol', label: '1. 列', title: '有序列表', kind: 'command', command: 'insertOrderedList' }
  ],
  [
    { key: 'link', label: '链接', title: '插入链接', kind: 'action' },
    { key: 'image', label: '图片', title: '插入图片', kind: 'action' },
    { key: 'table', label: '表格', title: '插入表格', kind: 'action' }
  ],
  [
    { key: 'undo', label: '撤销', title: '撤销', kind: 'command', command: 'undo' },
    { key: 'redo', label: '重做', title: '重做', kind: 'command', command: 'redo' },
    { key: 'clear', label: '清格', title: '清除格式', kind: 'command', command: 'removeFormat' },
    { key: 'delete', label: '删除', title: '删除所选或清空', kind: 'action' }
  ]
]

const hasHtmlTag = (value: string) => /<\/?[a-z][\s\S]*>/i.test(value)

const escapeHtml = (value: string) =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')

const normalizeEditorHtml = (value: string) => {
  const trimmed = value
    .replace(/<(p|div)>(<br>|&nbsp;|\s)*<\/(p|div)>/gi, '')
    .replace(/&nbsp;/gi, ' ')
    .trim()

  return trimmed === '<br>' ? '' : trimmed
}

const toEditorHtml = (value: string) => {
  if (!value) return ''
  if (hasHtmlTag(value)) return value
  return escapeHtml(value).replace(/\n/g, '<br>')
}

const syncEditorContent = (value: string) => {
  if (!editorRef.value) return
  const nextHtml = toEditorHtml(value)
  if (editorRef.value.innerHTML !== nextHtml) {
    editorRef.value.innerHTML = nextHtml
  }
}

const isEmpty = computed(() => !normalizeEditorHtml(editorRef.value?.innerHTML || ''))

const emitContent = () => {
  if (!editorRef.value) return
  emit('update:modelValue', normalizeEditorHtml(editorRef.value.innerHTML))
}

const saveSelection = () => {
  const selection = window.getSelection()
  if (!selection || selection.rangeCount === 0 || !editorRef.value) return

  const range = selection.getRangeAt(0)
  if (editorRef.value.contains(range.commonAncestorContainer)) {
    savedRange.value = range.cloneRange()
  }
}

const restoreSelection = () => {
  const selection = window.getSelection()
  if (!selection || !savedRange.value) return

  selection.removeAllRanges()
  selection.addRange(savedRange.value)
}

const focusEditor = async () => {
  await nextTick()
  editorRef.value?.focus()
  restoreSelection()
}

const runCommand = async (command: string, value?: string) => {
  if (props.disabled) return
  await focusEditor()
  document.execCommand(command, false, value)
  saveSelection()
  emitContent()
}

const wrapSelectionWithHtml = async (html: string) => {
  if (props.disabled) return
  await focusEditor()
  document.execCommand('insertHTML', false, html)
  saveSelection()
  emitContent()
}

const getSelectedText = () => {
  const selection = window.getSelection()
  return selection ? selection.toString() : ''
}

const insertHeading = async (level: 'h1' | 'h2' | 'h3') => {
  await runCommand('formatBlock', `<${level}>`)
}

const insertQuote = async () => {
  await runCommand('formatBlock', '<blockquote>')
}

const insertCode = async () => {
  const selectedText = getSelectedText().trim()
  if (selectedText) {
    const content = escapeHtml(selectedText)
    const html = selectedText.includes('\n')
      ? `<pre><code>${content}</code></pre>`
      : `<code>${content}</code>`
    await wrapSelectionWithHtml(html)
    return
  }

  const code = window.prompt('请输入代码内容')?.trim()
  if (!code) return

  const html = code.includes('\n')
    ? `<pre><code>${escapeHtml(code)}</code></pre>`
    : `<code>${escapeHtml(code)}</code>`
  await wrapSelectionWithHtml(html)
}

const insertLink = async () => {
  const url = window.prompt('请输入链接地址')?.trim()
  if (!url) return

  const selectedText = getSelectedText().trim()
  if (selectedText) {
    await runCommand('createLink', url)
    return
  }

  const text = window.prompt('请输入链接文字')?.trim() || url
  await wrapSelectionWithHtml(`<a href="${escapeHtml(url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(text)}</a>`)
}

const insertImage = async () => {
  const url = window.prompt('请输入图片地址')?.trim()
  if (!url) return

  const alt = window.prompt('请输入图片说明（可选）')?.trim() || 'image'
  await wrapSelectionWithHtml(`<img src="${escapeHtml(url)}" alt="${escapeHtml(alt)}" />`)
}

const insertTable = async () => {
  const rows = Number.parseInt(window.prompt('表格行数', '2') || '2', 10)
  const cols = Number.parseInt(window.prompt('表格列数', '3') || '3', 10)

  const safeRows = Number.isFinite(rows) && rows > 0 ? Math.min(rows, 10) : 2
  const safeCols = Number.isFinite(cols) && cols > 0 ? Math.min(cols, 8) : 3

  const header = Array.from({ length: safeCols }, (_, index) => `<th>标题${index + 1}</th>`).join('')
  const body = Array.from({ length: Math.max(safeRows - 1, 1) }, () => `<tr>${Array.from({ length: safeCols }, () => '<td>内容</td>').join('')}</tr>`).join('')
  await wrapSelectionWithHtml(`<table><thead><tr>${header}</tr></thead><tbody>${body}</tbody></table><p><br></p>`)
}

const deleteContent = async () => {
  await focusEditor()
  const selection = window.getSelection()
  if (selection && !selection.isCollapsed) {
    document.execCommand('delete')
  } else if (editorRef.value) {
    editorRef.value.innerHTML = ''
  }
  savedRange.value = null
  emitContent()
}

const handleAction = async (action: ToolbarAction) => {
  if (props.disabled) return

  if (action.kind === 'command' && action.command) {
    await runCommand(action.command, action.value)
    return
  }

  switch (action.key) {
    case 'h1':
      await insertHeading('h1')
      break
    case 'h2':
      await insertHeading('h2')
      break
    case 'h3':
      await insertHeading('h3')
      break
    case 'quote':
      await insertQuote()
      break
    case 'code':
      await insertCode()
      break
    case 'link':
      await insertLink()
      break
    case 'image':
      await insertImage()
      break
    case 'table':
      await insertTable()
      break
    case 'delete':
      await deleteContent()
      break
    default:
      break
  }
}

const handleInput = () => {
  saveSelection()
  emitContent()
}

const handleFocus = () => {
  isFocused.value = true
  saveSelection()
}

const handleBlur = () => {
  isFocused.value = false
  saveSelection()
  emitContent()
}

watch(
  () => props.modelValue,
  (value) => {
    if (!isFocused.value) {
      syncEditorContent(value || '')
    }
  }
)

onMounted(() => {
  syncEditorContent(props.modelValue || '')
})
</script>

<style scoped>
.rich-text-editor {
  border: 1px solid var(--el-border-color);
  border-radius: 16px;
  background: #fff;
  overflow: hidden;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.rich-text-editor:focus-within {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px rgba(64, 158, 255, 0.15);
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.16);
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
}

.toolbar-group {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding-right: 10px;
  margin-right: 2px;
  border-right: 1px solid rgba(148, 163, 184, 0.18);
}

.toolbar-group:last-child {
  padding-right: 0;
  margin-right: 0;
  border-right: none;
}

.toolbar-button {
  min-width: 38px;
  height: 32px;
  padding: 0 10px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  background: #fff;
  color: #334155;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
  cursor: pointer;
  transition: all 0.2s ease;
}

.toolbar-button:hover:not(:disabled) {
  border-color: rgba(64, 158, 255, 0.35);
  color: #2563eb;
  transform: translateY(-1px);
}

.toolbar-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor {
  min-height: 180px;
  padding: 16px;
  font-size: 14px;
  line-height: 1.75;
  color: #1f2937;
  outline: none;
  word-break: break-word;
}

.editor.is-empty::before {
  content: attr(data-placeholder);
  color: #9ca3af;
}

.editor :deep(h1),
.editor :deep(h2),
.editor :deep(h3),
.editor :deep(p),
.editor :deep(ul),
.editor :deep(ol),
.editor :deep(blockquote),
.editor :deep(pre),
.editor :deep(table) {
  margin: 0 0 12px;
}

.editor :deep(h1) {
  font-size: 24px;
}

.editor :deep(h2) {
  font-size: 20px;
}

.editor :deep(h3) {
  font-size: 17px;
}

.editor :deep(ul),
.editor :deep(ol) {
  padding-left: 22px;
}

.editor :deep(blockquote) {
  padding-left: 12px;
  border-left: 3px solid rgba(64, 158, 255, 0.35);
  color: #475569;
}

.editor :deep(pre) {
  padding: 12px 14px;
  background: #f8fafc;
  border-radius: 12px;
  overflow-x: auto;
}

.editor :deep(code) {
  padding: 2px 6px;
  border-radius: 6px;
  background: rgba(15, 23, 42, 0.06);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.editor :deep(pre code) {
  padding: 0;
  background: transparent;
}

.editor :deep(img) {
  max-width: 100%;
  border-radius: 12px;
}

.editor :deep(table) {
  width: 100%;
  border-collapse: collapse;
}

.editor :deep(th),
.editor :deep(td) {
  padding: 8px 10px;
  border: 1px solid #dbe4f0;
}

.editor :deep(th) {
  background: #f8fafc;
}

.is-disabled {
  background: #f8fafc;
}

.is-disabled .editor {
  color: #94a3b8;
}

@media (max-width: 768px) {
  .toolbar {
    gap: 8px;
  }

  .toolbar-group {
    width: 100%;
    padding-right: 0;
    margin-right: 0;
    border-right: none;
  }
}
</style>
