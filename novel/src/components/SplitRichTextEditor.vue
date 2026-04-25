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
  { key: 'selectAll', label: '全选', title: '全选', kind: 'action' },
  { key: 'replace', label: '替换', title: '查找和替换', kind: 'action' },
  { key: 'bold', label: 'B', title: '粗体', kind: 'command', command: 'bold' },
  { key: 'italic', label: 'I', title: '斜体', kind: 'command', command: 'italic' },
  { key: 'h1', label: 'H1', title: '一级标题', kind: 'action' },
  { key: 'h2', label: 'H2', title: '二级标题', kind: 'action' },
  { key: 'underline', label: 'U', title: '下划线', kind: 'command', command: 'underline' },
  { key: 'strike', label: 'S', title: '删除线', kind: 'command', command: 'strikeThrough', glyphClass: 'is-strike' },
  { key: 'quote', label: '', title: '引用', kind: 'action', icon: markRaw(ChatLineSquare) },
  { key: 'ul', label: '', title: '无序列表', kind: 'command', command: 'insertUnorderedList', icon: markRaw(List) },
  { key: 'ol', label: '', title: '有序列表', kind: 'command', command: 'insertOrderedList', icon: markRaw(Tickets) },
  { key: 'link', label: '', title: '插入链接', kind: 'action', icon: markRaw(Link) },
  { key: 'image', label: '', title: '插入图片', kind: 'action', icon: markRaw(Picture) },
  { key: 'code', label: '</>', title: '代码', kind: 'action' },
  { key: 'table', label: '', title: '插入表格', kind: 'action', icon: markRaw(Grid) },
  { key: 'undo', label: '', title: '撤销', kind: 'command', command: 'undo', icon: markRaw(RefreshLeft) },
  { key: 'redo', label: '', title: '重做', kind: 'command', command: 'redo', icon: markRaw(RefreshRight) },
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

const insertPlainText = async (text: string) => {
  if (props.disabled) return
  await focusEditor()
  document.execCommand('insertText', false, text)
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

const selectAll = async () => {
  await focusEditor()
  if (editorRef.value) {
    const range = document.createRange()
    range.selectNodeContents(editorRef.value)
    const selection = window.getSelection()
    if (selection) {
      selection.removeAllRanges()
      selection.addRange(range)
    }
  }
}

const replaceText = async () => {
  await focusEditor()
  if (!editorRef.value) return

  const selectedText = getSelectedText()
  const findText = window.prompt('请输入要查找的内容：', selectedText || '')
  if (!findText) return

  const replaceWith = window.prompt('请输入要替换为的内容：', '') ?? ''
  
  const html = editorRef.value.innerHTML
  const escapedFind = escapeHtml(findText)
  const escapedReplace = escapeHtml(replaceWith)
  
  const regex = new RegExp(escapedFind.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g')
  const newHtml = html.replace(regex, escapedReplace)
  
  if (newHtml !== html) {
    editorRef.value.innerHTML = newHtml
    savedRange.value = null
    emitContent()
    
    nextTick(() => {
      window.alert(`已替换 ${countMatches(html, escapedFind)} 处内容`)
    })
  } else {
    window.alert(`未找到"${findText}"`)
  }
}

const countMatches = (html: string, text: string): number => {
  const regex = new RegExp(text.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), 'g')
  const matches = html.match(regex)
  return matches ? matches.length : 0
}

const handleAction = async (action: ToolbarAction) => {
  if (props.disabled) return

  if (action.kind === 'command' && action.command) {
    await runCommand(action.command, action.value)
    return
  }

  switch (action.key) {
    case 'selectAll':
      await selectAll()
      break
    case 'replace':
      await replaceText()
      break
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

defineExpose({
  wrapSelectionWithHtml,
  focusEditor,
  getSelectedText,
  insertContent: wrapSelectionWithHtml,
  insertPlainText
})

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
  border: 1px solid rgba(0, 201, 167, 0.16);
  border-radius: 22px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(247, 255, 252, 0.98) 100%);
  overflow: hidden;
  box-shadow:
    0 20px 40px rgba(0, 136, 110, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 0.72);
  transition: border-color 0.25s ease, box-shadow 0.25s ease, transform 0.25s ease;
}

.split-rich-text-editor:focus-within {
  border-color: rgba(0, 168, 150, 0.9);
  box-shadow:
    0 24px 56px rgba(0, 136, 110, 0.14),
    0 0 0 4px rgba(0, 201, 167, 0.12);
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px 12px 7px;
  border-bottom: 1px solid rgba(0, 201, 167, 0.1);
  background:
    linear-gradient(180deg, rgba(247, 255, 252, 0.98) 0%, rgba(240, 253, 249, 0.96) 100%);
}

.toolbar-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 30px;
  height: 26px;
  padding: 0 7px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.55);
  color: #436a63;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  cursor: pointer;
  transition: all 0.2s ease;
}

.toolbar-button:hover:not(:disabled) {
  border-color: rgba(0, 201, 167, 0.28);
  background: rgba(255, 255, 255, 0.94);
  color: #00a187;
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(0, 168, 150, 0.12);
}

.toolbar-button:disabled {
  opacity: 0.42;
  cursor: not-allowed;
}

.toolbar-icon {
  width: 13px;
  height: 13px;
}

.toolbar-glyph {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  font-size: 11px;
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
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(247, 255, 252, 0.98) 100%);
}

.editor-surface {
  flex: 1;
  min-height: 240px;
  width: 100%;
  padding: 34px 44px 56px;
  color: #29443f;
  font-size: 16px;
  line-height: 2;
  letter-spacing: 0.01em;
  outline: none;
  overflow-y: auto;
  word-break: break-word;
  box-sizing: border-box;
}

.editor-surface.is-empty::before {
  content: attr(data-placeholder);
  color: #92b8b0;
}

.editor-surface :deep(h1),
.editor-surface :deep(h2),
.editor-surface :deep(p),
.editor-surface :deep(ul),
.editor-surface :deep(ol),
.editor-surface :deep(blockquote),
.editor-surface :deep(pre),
.editor-surface :deep(table) {
  margin: 0 0 18px;
}

.editor-surface :deep(h1) {
  margin-top: 8px;
  font-size: 32px;
  line-height: 1.35;
  font-weight: 500;
  color: #23423d;
  letter-spacing: 0.02em;
}

.editor-surface :deep(h2) {
  font-size: 24px;
  line-height: 1.45;
  font-weight: 600;
  color: #2d544d;
}

.editor-surface :deep(p) {
  min-height: 2em;
}

.editor-surface :deep(ul),
.editor-surface :deep(ol) {
  padding-left: 26px;
}

.editor-surface :deep(blockquote) {
  padding: 4px 0 4px 18px;
  border-left: 3px solid rgba(0, 201, 167, 0.58);
  color: #46716a;
  background: rgba(233, 250, 246, 0.7);
  border-radius: 0 12px 12px 0;
}

.editor-surface :deep(pre) {
  padding: 14px 16px;
  background: rgba(0, 136, 110, 0.06);
  border-radius: 14px;
  overflow-x: auto;
}

.editor-surface :deep(code) {
  padding: 2px 6px;
  border-radius: 6px;
  background: rgba(0, 136, 110, 0.09);
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}

.editor-surface :deep(pre code) {
  padding: 0;
  background: transparent;
}

.editor-surface :deep(a) {
  color: #00967c;
  text-decoration: none;
}

.editor-surface :deep(a:hover) {
  text-decoration: underline;
}

.editor-surface :deep(img) {
  max-width: 100%;
  border-radius: 14px;
}

.editor-surface :deep(table) {
  width: 100%;
  border-collapse: collapse;
}

.editor-surface :deep(th),
.editor-surface :deep(td) {
  padding: 8px 10px;
  border: 1px solid #cfe8e2;
}

.editor-surface :deep(th) {
  background: rgba(238, 250, 246, 0.9);
}

.is-disabled {
  background: rgba(245, 252, 250, 0.96);
}

.is-disabled .editor-surface {
  color: #99b7b1;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .toolbar {
    gap: 4px;
    padding: 8px;
  }

  .editor-surface {
    padding: 24px 20px 36px;
    font-size: 15px;
  }
}

/* 暗色主题适配 */
:root[data-theme='dark'] .split-rich-text-editor {
  background: transparent !important;
}

:root[data-theme='dark'] .toolbar {
  background: rgba(30, 41, 59, 0.95) !important;
  border-bottom-color: rgba(71, 85, 105, 0.5) !important;
}

:root[data-theme='dark'] .toolbar-button {
  background: rgba(51, 65, 85, 0.6) !important;
  color: #c8d0e0 !important;
  border-color: rgba(71, 85, 105, 0.4) !important;
}

:root[data-theme='dark'] .toolbar-button:hover:not(:disabled) {
  background: rgba(71, 85, 105, 0.8) !important;
  color: #5eead4 !important;
  border-color: rgba(94, 234, 212, 0.3) !important;
}

:root[data-theme='dark'] .editor-shell {
  background: transparent !important;
}

:root[data-theme='dark'] .editor-surface {
  color: #f4f7ff !important;
}

:root[data-theme='dark'] .editor-surface.is-empty::before {
  color: #6b7280 !important;
}

:root[data-theme='dark'] .editor-surface :deep(h1) {
  color: #f4f7ff !important;
}

:root[data-theme='dark'] .editor-surface :deep(h2) {
  color: #c8d0e0 !important;
}

:root[data-theme='dark'] .editor-surface :deep(blockquote) {
  color: #c8d0e0 !important;
  background: rgba(94, 234, 212, 0.08) !important;
  border-left-color: rgba(94, 234, 212, 0.4) !important;
}

:root[data-theme='dark'] .editor-surface :deep(pre) {
  background: rgba(30, 41, 59, 0.6) !important;
}

:root[data-theme='dark'] .editor-surface :deep(code) {
  background: rgba(94, 234, 212, 0.12) !important;
  color: #f4f7ff !important;
}

:root[data-theme='dark'] .editor-surface :deep(a) {
  color: #5eead4 !important;
}

:root[data-theme='dark'] .editor-surface :deep(th),
:root[data-theme='dark'] .editor-surface :deep(td) {
  border-color: rgba(71, 85, 105, 0.5) !important;
  color: #f4f7ff !important;
}

:root[data-theme='dark'] .editor-surface :deep(th) {
  background: rgba(51, 65, 85, 0.5) !important;
}

:root[data-theme='dark'] .split-rich-text-editor.is-disabled {
  background: rgba(30, 41, 59, 0.5) !important;
}

:root[data-theme='dark'] .split-rich-text-editor.is-disabled .editor-surface {
  color: #6b7280 !important;
}
</style>
