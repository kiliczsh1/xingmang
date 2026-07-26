<template>
  <div class="split-rich-text-editor" :class="{ 'is-disabled': disabled }">
    <div class="toolbar" role="toolbar" aria-label="简介编辑工具栏">
      <el-tooltip
        v-for="action in toolbarActions"
        :key="action.key"
        :content="action.title"
        placement="top"
        :show-after="150"
      >
        <button
          type="button"
          class="toolbar-button"
          :aria-label="action.title"
          :disabled="disabled"
          @mousedown.prevent
          @click="handleAction(action)"
        >
          <component :is="action.icon" v-if="action.icon" class="toolbar-icon" />
          <span v-else class="toolbar-glyph" :class="action.glyphClass">{{ action.label }}</span>
        </button>
      </el-tooltip>
    </div>

    <div class="editor-shell">
      <div class="editor-meta">
        <span class="editor-title">所见即所得编辑</span>
        <span class="editor-subtitle">输入、排版和渲染都在同一块区域完成</span>
      </div>
      <div
        ref="editorRef"
        class="editor-surface"
        :class="{ 'is-empty': isEmpty }"
        :contenteditable="!disabled"
        :data-placeholder="placeholder"
        :spellcheck="false"
        @focus="handleFocus"
        @blur="handleBlur"
        @input="handleInput"
        @keyup="saveSelection"
        @mouseup="saveSelection"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, markRaw, nextTick, onMounted, ref, watch } from 'vue'
import {
  Back,
  Brush,
  ChatLineSquare,
  Delete,
  Grid,
  Link,
  List,
  Operation,
  Picture,
  RefreshLeft,
  RefreshRight,
  Right,
  Tickets
} from '@element-plus/icons-vue'

type ToolbarAction = {
  key: string
  label: string
  title: string
  kind: 'command' | 'action'
  command?: string
  value?: string
  icon?: object
  glyphClass?: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
}>(), {
  placeholder: '请输入简介内容',
  disabled: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'blur'): void
}>()

const editorRef = ref<HTMLDivElement | null>(null)
const isFocused = ref(false)
const savedRange = ref<Range | null>(null)

const toolbarActions: ToolbarAction[] = [
  { key: 'bold', label: 'B', title: '粗体', kind: 'command', command: 'bold' },
  { key: 'italic', label: 'I', title: '斜体', kind: 'command', command: 'italic' },
  { key: 'h1', label: 'H1', title: '一级标题', kind: 'action' },
  { key: 'h2', label: 'H2', title: '二级标题', kind: 'action' },
  { key: 'underline', label: 'U', title: '下划线', kind: 'command', command: 'underline' },
  { key: 'strike', label: 'S', title: '删除线', kind: 'command', command: 'strikeThrough', glyphClass: 'is-strike' },
  { key: 'quote', label: '', title: '引用', kind: 'action', icon: markRaw(ChatLineSquare) },
  { key: 'sup', label: 'x2', title: '上标', kind: 'command', command: 'superscript' },
  { key: 'sub', label: 'x_', title: '下标', kind: 'command', command: 'subscript' },
  { key: 'left', label: '', title: '左对齐', kind: 'command', command: 'justifyLeft', icon: markRaw(Back) },
  { key: 'center', label: '', title: '居中', kind: 'command', command: 'justifyCenter', icon: markRaw(Operation) },
  { key: 'right', label: '', title: '右对齐', kind: 'command', command: 'justifyRight', icon: markRaw(Right) },
  { key: 'ul', label: '', title: '无序列表', kind: 'command', command: 'insertUnorderedList', icon: markRaw(List) },
  { key: 'ol', label: '', title: '有序列表', kind: 'command', command: 'insertOrderedList', icon: markRaw(Tickets) },
  { key: 'link', label: '', title: '插入链接', kind: 'action', icon: markRaw(Link) },
  { key: 'image', label: '', title: '插入图片', kind: 'action', icon: markRaw(Picture) },
  { key: 'code', label: '</>', title: '代码', kind: 'action' },
  { key: 'table', label: '', title: '插入表格', kind: 'action', icon: markRaw(Grid) },
  { key: 'undo', label: '', title: '撤销', kind: 'command', command: 'undo', icon: markRaw(RefreshLeft) },
  { key: 'redo', label: '', title: '重做', kind: 'command', command: 'redo', icon: markRaw(RefreshRight) },
  { key: 'delete', label: '', title: '删除选中内容', kind: 'action', icon: markRaw(Delete) },
  { key: 'clear', label: '', title: '清除格式', kind: 'action', icon: markRaw(Brush) }
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

const isEmpty = computed(() => !normalizeEditorHtml(editorRef.value?.innerHTML || props.modelValue || ''))

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

const insertHeading = async (level: 'h1' | 'h2') => {
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

  const text = window.prompt('请输入链接文本')?.trim() || url
  await wrapSelectionWithHtml(
    `<a href="${escapeHtml(url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(text)}</a>`
  )
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
  const body = Array.from({ length: Math.max(safeRows - 1, 1) }, () =>
    `<tr>${Array.from({ length: safeCols }, () => '<td>内容</td>').join('')}</tr>`
  ).join('')

  await wrapSelectionWithHtml(
    `<table><thead><tr>${header}</tr></thead><tbody>${body}</tbody></table><p><br></p>`
  )
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

const clearFormatting = async () => {
  await focusEditor()
  document.execCommand('removeFormat')
  saveSelection()
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
    case 'quote':
      await insertQuote()
      break
    case 'link':
      await insertLink()
      break
    case 'image':
      await insertImage()
      break
    case 'code':
      await insertCode()
      break
    case 'table':
      await insertTable()
      break
    case 'delete':
      await deleteContent()
      break
    case 'clear':
      await clearFormatting()
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
  emit('blur')
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
.split-rich-text-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid var(--el-border-color);
  border-radius: 16px;
  background: #fff;
  overflow: hidden;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.split-rich-text-editor:focus-within {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px rgba(64, 158, 255, 0.15);
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 12px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.16);
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
}

.toolbar-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 42px;
  height: 34px;
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

.toolbar-icon {
  width: 15px;
  height: 15px;
}

.toolbar-glyph {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
}

.toolbar-glyph.is-strike {
  text-decoration: line-through;
}

.editor-shell {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.editor-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-bottom: 1px solid rgba(148, 163, 184, 0.12);
  background: rgba(248, 250, 252, 0.76);
}

.editor-title {
  font-size: 13px;
  font-weight: 700;
  color: #1f2937;
}

.editor-subtitle {
  font-size: 12px;
  color: #64748b;
}

.editor-surface {
  flex: 1;
  min-height: 240px;
  padding: 18px;
  color: #1f2937;
  font-size: 14px;
  line-height: 1.75;
  outline: none;
  overflow-y: auto;
  word-break: break-word;
}

.editor-surface.is-empty::before {
  content: attr(data-placeholder);
  color: #94a3b8;
}

.editor-surface :deep(h1),
.editor-surface :deep(h2),
.editor-surface :deep(p),
.editor-surface :deep(ul),
.editor-surface :deep(ol),
.editor-surface :deep(blockquote),
.editor-surface :deep(pre),
.editor-surface :deep(table) {
  margin: 0 0 12px;
}

.editor-surface :deep(h1) {
  font-size: 24px;
}

.editor-surface :deep(h2) {
  font-size: 20px;
}

.editor-surface :deep(ul),
.editor-surface :deep(ol) {
  padding-left: 22px;
}

.editor-surface :deep(blockquote) {
  padding-left: 12px;
  border-left: 3px solid rgba(64, 158, 255, 0.35);
  color: #475569;
}

.editor-surface :deep(pre) {
  padding: 12px 14px;
  background: #f8fafc;
  border-radius: 12px;
  overflow-x: auto;
}

.editor-surface :deep(code) {
  padding: 2px 6px;
  border-radius: 6px;
  background: rgba(15, 23, 42, 0.06);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.editor-surface :deep(pre code) {
  padding: 0;
  background: transparent;
}

.editor-surface :deep(a) {
  color: #2563eb;
  text-decoration: none;
}

.editor-surface :deep(a:hover) {
  text-decoration: underline;
}

.editor-surface :deep(img) {
  max-width: 100%;
  border-radius: 12px;
}

.editor-surface :deep(table) {
  width: 100%;
  border-collapse: collapse;
}

.editor-surface :deep(th),
.editor-surface :deep(td) {
  padding: 8px 10px;
  border: 1px solid #dbe4f0;
}

.editor-surface :deep(th) {
  background: #f8fafc;
}

.is-disabled {
  background: #f8fafc;
}

.is-disabled .editor-surface {
  color: #94a3b8;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .toolbar {
    gap: 6px;
  }

  .editor-meta {
    align-items: flex-start;
    flex-direction: column;
  }
}
</style>
