<template>
  <div class="experience-content-editor">
    <div class="editor-toolbar">
      <div class="toolbar-left">
        <el-tooltip
          v-for="action in toolbarActions"
          :key="action.key"
          :content="action.title"
          placement="top"
          :show-after="150"
        >
          <button
            type="button"
            class="toolbar-btn"
            :disabled="disabled"
            @click="handleAction(action)"
          >
            <component :is="action.icon" v-if="action.icon" class="btn-icon" />
            <span v-else class="btn-text" :class="action.glyphClass">{{ action.label }}</span>
          </button>
        </el-tooltip>
      </div>
      <div class="toolbar-right">
        <el-radio-group v-model="localRenderMode" size="small" @change="handleRenderModeChange">
          <el-radio-button label="markdown">MD 渲染</el-radio-button>
          <el-radio-button label="html">HTML 渲染</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <div class="editor-body">
      <div class="editor-pane source-pane">
        <div class="pane-header">
          <span>源码</span>
          <span class="char-count">{{ charCount }} 字符</span>
        </div>
        <textarea
          ref="textareaRef"
          v-model="localContent"
          class="source-textarea"
          :placeholder="placeholder"
          :disabled="disabled"
          @input="handleInput"
          @focus="handleFocus"
          @blur="handleBlur"
        ></textarea>
      </div>

      <div class="editor-pane preview-pane">
        <div class="pane-header">
          <span>预览</span>
          <el-tag size="small" :type="localRenderMode === 'markdown' ? 'success' : 'info'">
            {{ localRenderMode === 'markdown' ? 'MD 渲染' : 'HTML 渲染' }}
          </el-tag>
        </div>
        <div class="preview-content">
          <MarkdownRenderer
            v-if="localContent"
            :content="localContent"
            :render-mode="localRenderMode"
          />
          <div v-else class="preview-empty">
            {{ placeholder }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, markRaw, ref, watch } from 'vue'
import {
  Brush,
  ChatLineSquare,
  Grid,
  Link,
  List,
  Picture,
  Tickets
} from '@element-plus/icons-vue'
import MarkdownRenderer from './MarkdownRenderer.vue'

type RenderMode = 'markdown' | 'html'

type ToolbarAction = {
  key: string
  label: string
  title: string
  icon?: object
  glyphClass?: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  renderMode?: RenderMode
  placeholder?: string
  disabled?: boolean
}>(), {
  renderMode: 'markdown',
  placeholder: '请输入内容',
  disabled: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'update:renderMode', value: RenderMode): void
  (e: 'blur'): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const localContent = ref(props.modelValue || '')
const localRenderMode = ref<RenderMode>(props.renderMode)

const charCount = computed(() => localContent.value.length)

const toolbarActions: ToolbarAction[] = [
  { key: 'h1', label: 'H1', title: '一级标题 (# )' },
  { key: 'h2', label: 'H2', title: '二级标题 (## )' },
  { key: 'bold', label: 'B', title: '粗体 (**text**)', glyphClass: 'is-bold' },
  { key: 'italic', label: 'I', title: '斜体 (*text*)', glyphClass: 'is-italic' },
  { key: 'strike', label: 'S', title: '删除线 (~~text~~)', glyphClass: 'is-strike' },
  { key: 'quote', label: '', title: '引用 (> )', icon: markRaw(ChatLineSquare) },
  { key: 'ul', label: '', title: '无序列表 (- )', icon: markRaw(List) },
  { key: 'ol', label: '', title: '有序列表 (1. )', icon: markRaw(Tickets) },
  { key: 'link', label: '', title: '插入链接', icon: markRaw(Link) },
  { key: 'image', label: '', title: '插入图片', icon: markRaw(Picture) },
  { key: 'code', label: '</>', title: '行内代码 (`code`)' },
  { key: 'codeBlock', label: '```', title: '代码块' },
  { key: 'table', label: '', title: '插入表格', icon: markRaw(Grid) },
  { key: 'hr', label: '—', title: '分割线 (---)' },
  { key: 'clear', label: '', title: '清空内容', icon: markRaw(Brush) }
]

const insertText = (text: string) => {
  if (!textareaRef.value || props.disabled) return
  
  const textarea = textareaRef.value
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const before = localContent.value.substring(0, start)
  const after = localContent.value.substring(end)
  
  localContent.value = before + text + after
  
  emit('update:modelValue', localContent.value)
  
  setTimeout(() => {
    textarea.focus()
    textarea.setSelectionRange(start + text.length, start + text.length)
  }, 0)
}

const wrapSelection = (before: string, after: string, placeholder = '') => {
  if (!textareaRef.value || props.disabled) return
  
  const textarea = textareaRef.value
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const selectedText = localContent.value.substring(start, end)
  
  const text = selectedText || placeholder
  const newText = before + text + after
  
  localContent.value = localContent.value.substring(0, start) + newText + localContent.value.substring(end)
  
  emit('update:modelValue', localContent.value)
  
  setTimeout(() => {
    textarea.focus()
    if (selectedText) {
      textarea.setSelectionRange(start + newText.length, start + newText.length)
    } else {
      textarea.setSelectionRange(start + before.length, start + before.length + placeholder.length)
    }
  }, 0)
}

const insertAtLineStart = (prefix: string) => {
  if (!textareaRef.value || props.disabled) return
  
  const textarea = textareaRef.value
  const start = textarea.selectionStart
  
  const beforeCursor = localContent.value.substring(0, start)
  const afterCursor = localContent.value.substring(start)
  
  const lastNewLine = beforeCursor.lastIndexOf('\n')
  const insertPos = lastNewLine + 1
  
  localContent.value = beforeCursor.substring(0, insertPos) + prefix + beforeCursor.substring(insertPos) + afterCursor
  
  emit('update:modelValue', localContent.value)
  
  setTimeout(() => {
    textarea.focus()
    textarea.setSelectionRange(start + prefix.length, start + prefix.length)
  }, 0)
}

const handleAction = (action: ToolbarAction) => {
  if (props.disabled) return
  
  switch (action.key) {
    case 'h1':
      insertAtLineStart('# ')
      break
    case 'h2':
      insertAtLineStart('## ')
      break
    case 'bold':
      wrapSelection('**', '**', '粗体文本')
      break
    case 'italic':
      wrapSelection('*', '*', '斜体文本')
      break
    case 'strike':
      wrapSelection('~~', '~~', '删除线文本')
      break
    case 'quote':
      insertAtLineStart('> ')
      break
    case 'ul':
      insertAtLineStart('- ')
      break
    case 'ol':
      insertAtLineStart('1. ')
      break
    case 'link': {
      const url = window.prompt('请输入链接地址')?.trim()
      if (!url) return
      const text = window.prompt('请输入链接文本（可选）')?.trim() || url
      insertText(`[${text}](${url})`)
      break
    }
    case 'image': {
      const imgUrl = window.prompt('请输入图片地址')?.trim()
      if (!imgUrl) return
      const alt = window.prompt('请输入图片说明（可选）')?.trim() || 'image'
      insertText(`![${alt}](${imgUrl})`)
      break
    }
    case 'code':
      wrapSelection('`', '`', 'code')
      break
    case 'codeBlock': {
      const lang = window.prompt('请输入代码语言（可选）', '')?.trim() || ''
      insertText(`\n\`\`\`${lang}\ncode here\n\`\`\`\n`)
      break
    }
    case 'table': {
      const rows = Number.parseInt(window.prompt('表格行数', '2') || '2', 10)
      const cols = Number.parseInt(window.prompt('表格列数', '3') || '3', 10)
      
      const safeRows = Number.isFinite(rows) && rows > 0 ? Math.min(rows, 10) : 2
      const safeCols = Number.isFinite(cols) && cols > 0 ? Math.min(cols, 8) : 3
      
      const header = '| ' + Array.from({ length: safeCols }, (_, i) => `标题${i + 1}`).join(' | ') + ' |'
      const separator = '| ' + Array.from({ length: safeCols }, () => '---').join(' | ') + ' |'
      const body = Array.from({ length: Math.max(safeRows - 1, 1) }, () =>
        '| ' + Array.from({ length: safeCols }, () => '内容').join(' | ') + ' |'
      ).join('\n')
      
      insertText(`\n${header}\n${separator}\n${body}\n`)
      break
    }
    case 'hr':
      insertText('\n---\n')
      break
    case 'clear':
      localContent.value = ''
      emit('update:modelValue', '')
      break
    default:
      break
  }
}

const handleInput = () => {
  emit('update:modelValue', localContent.value)
}

const handleRenderModeChange = (value: RenderMode) => {
  emit('update:renderMode', value)
}

const handleFocus = () => {
  textareaRef.value?.focus()
}

const handleBlur = () => {
  emit('blur')
}

watch(
  () => props.modelValue,
  (value) => {
    localContent.value = value || ''
  }
)

watch(
  () => props.renderMode,
  (value) => {
    localRenderMode.value = value
  }
)
</script>

<style scoped>
.experience-content-editor {
  display: flex;
  flex-direction: column;
  border: 1px solid rgba(0, 201, 167, 0.16);
  border-radius: 22px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(247, 255, 252, 0.98) 100%);
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 136, 110, 0.08);
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(0, 201, 167, 0.1);
  background: linear-gradient(180deg, rgba(247, 255, 252, 0.98) 0%, rgba(240, 253, 249, 0.96) 100%);
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toolbar-btn {
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

.toolbar-btn:hover:not(:disabled) {
  border-color: rgba(0, 201, 167, 0.28);
  background: rgba(255, 255, 255, 0.94);
  color: #00a187;
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(0, 168, 150, 0.12);
}

.toolbar-btn:disabled {
  opacity: 0.42;
  cursor: not-allowed;
}

.btn-icon {
  width: 13px;
  height: 13px;
}

.btn-text {
  font-size: 11px;
  font-weight: 700;
}

.btn-text.is-bold {
  font-weight: 800;
}

.btn-text.is-italic {
  font-style: italic;
}

.btn-text.is-strike {
  text-decoration: line-through;
}

.editor-body {
  display: flex;
  min-height: 400px;
  max-height: 600px;
}

.editor-pane {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.source-pane {
  border-right: 1px solid rgba(0, 201, 167, 0.1);
}

.pane-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 16px;
  border-bottom: 1px solid rgba(0, 201, 167, 0.08);
  background: rgba(247, 255, 252, 0.5);
  font-size: 13px;
  font-weight: 600;
  color: #436a63;
}

.char-count {
  font-size: 12px;
  color: #92b8b0;
  font-weight: 400;
}

.source-textarea {
  flex: 1;
  width: 100%;
  padding: 16px 20px;
  border: none;
  outline: none;
  resize: none;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.7;
  color: #29443f;
  background: transparent;
}

.source-textarea::placeholder {
  color: #92b8b0;
}

.source-textarea:disabled {
  color: #99b7b1;
  cursor: not-allowed;
  background: rgba(245, 252, 250, 0.5);
}

.preview-pane {
  background: rgba(255, 251, 244, 0.3);
}

.preview-content {
  flex: 1;
  padding: 16px 20px;
  overflow-y: auto;
}

.preview-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #92b8b0;
  font-size: 14px;
}

@media (max-width: 900px) {
  .editor-body {
    flex-direction: column;
    max-height: none;
  }

  .source-pane {
    border-right: none;
    border-bottom: 1px solid rgba(0, 201, 167, 0.1);
  }

  .editor-pane {
    min-height: 300px;
  }
}
</style>
