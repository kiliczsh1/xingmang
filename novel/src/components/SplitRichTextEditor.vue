<template>
  <div
    :id="editorId"
    class="split-rich-text-editor"
    :class="{ 'is-disabled': disabled }"
  ></div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Vditor from 'vditor'
import 'vditor/dist/index.css'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
  fontFamily?: string
  fontSize?: number
}>(), {
  placeholder: '请输入简介内容',
  disabled: false,
  fontFamily: 'Microsoft YaHei',
  fontSize: 16,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'blur'): void
}>()

const editorId = `vditor-${Math.random().toString(36).slice(2, 10)}`
const editorRef = ref<Vditor | null>(null)
let settingFromExternal = false
let initialized = false

const hasHtmlTag = (value: string) => /<\/?[a-z][\s\S]*>/i.test(value)

const isDark = () =>
  document.documentElement.getAttribute('data-theme') === 'dark'

const toolbar = [
  'emoji',
  'headings',
  'bold',
  'italic',
  'strike',
  '|',
  'list',
  'ordered-list',
  'check',
  '|',
  'outdent',
  'indent',
  '|',
  'quote',
  'link',
  'inline-code',
  'code-block',
  'table',
  '|',
  'undo',
  'redo',
  '|',
  'fullscreen',
  'both',
  'preview',
]

const fontStyle = computed(() => `
  .vditor-reset {
    font-family: ${props.fontFamily} !important;
    font-size: ${props.fontSize}px !important;
    line-height: 2 !important;
  }
  .vditor-reset p { line-height: 2 !important; }
`)

const toMd = (value: string): string => {
  if (!value) return ''
  if (!hasHtmlTag(value)) return value
  try {
    return editorRef.value ? editorRef.value.html2md(value) : value
  } catch {
    return value
  }
}

const syncContent = (value: string) => {
  if (!editorRef.value) return
  const md = toMd(value || '')
  if (editorRef.value.getValue() !== md) {
    settingFromExternal = true
    editorRef.value.setValue(md)
  }
}

onMounted(() => {
  const instance = new Vditor(editorId, {
    mode: 'wysiwyg',
    placeholder: props.placeholder,
    cache: { enable: false },
    theme: isDark() ? 'dark' : 'classic',
    toolbar,
    width: '100%',
    minHeight: 240,
    toolbarConfig: { pin: true },
    counter: {
      enable: false,
    },
    preview: {
      markdown: {
        autoSpace: true,
        paragraphBeginningSpace: false,
        fixTermTypo: false,
        toc: false,
        mark: true,
        gfmAutoLink: true,
      },
    },
    after: () => {
      initialized = true
      if (props.disabled) {
        try { instance.disabled() } catch { /* ignore */ }
      }
      nextTick(() => {
        if (editorRef.value && props.modelValue) {
          syncContent(props.modelValue)
        }
      })
    },
    input: (value: string) => {
      if (settingFromExternal) {
        settingFromExternal = false
        return
      }
      emit('update:modelValue', value)
    },
    blur: () => {
      emit('blur')
    },
  })

  editorRef.value = instance
})

watch(
  () => props.modelValue,
  (value) => {
    if (!editorRef.value || !initialized) return
    syncContent(value || '')
  }
)

watch(
  () => props.disabled,
  (val) => {
    if (!editorRef.value) return
    if (val) {
      editorRef.value.disabled()
    } else {
      editorRef.value.enable()
    }
  }
)

// 暴露给父组件的方法
const insertPlainText = (text: string) => {
  if (props.disabled || !editorRef.value) return
  editorRef.value.focus()
  editorRef.value.insertValue(text)
}

const wrapSelectionWithHtml = (html: string) => {
  if (props.disabled || !editorRef.value) return
  editorRef.value.focus()
  const md = toMd(html)
  editorRef.value.insertValue(md)
}

const focusEditor = () => {
  editorRef.value?.focus()
}

const getSelectedText = (): string => {
  return editorRef.value?.getSelection() || ''
}

defineExpose({
  wrapSelectionWithHtml,
  focusEditor,
  getSelectedText,
  insertContent: wrapSelectionWithHtml,
  insertPlainText,
})

onUnmounted(() => {
  if (editorRef.value) {
    try { editorRef.value.destroy() } catch { /* ignore */ }
    editorRef.value = null
  }
})
</script>

<style>
/* 容器样式 — 匹配原来的外观 */
.split-rich-text-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  border: 1px solid rgba(0, 201, 167, 0.12);
  border-radius: 16px;
  background: transparent;
  overflow: hidden;
  box-shadow:
    0 1px 3px rgba(0, 201, 167, 0.03),
    0 8px 24px rgba(0, 201, 167, 0.05);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}
.split-rich-text-editor:focus-within {
  border-color: rgba(0, 201, 167, 0.4);
  box-shadow:
    0 1px 3px rgba(0, 201, 167, 0.05),
    0 12px 32px rgba(0, 201, 167, 0.1);
}

/* Vditor 内部样式覆盖 */
.split-rich-text-editor .vditor {
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  height: 100% !important;
}
.split-rich-text-editor .vditor--dark {
  background: transparent !important;
}

.split-rich-text-editor .vditor-toolbar {
  padding: 8px 12px !important;
  border-bottom: 1px solid rgba(0, 201, 167, 0.1) !important;
  background: rgba(0, 201, 167, 0.04) !important;
  border-radius: 0 !important;
}
.split-rich-text-editor .vditor-toolbar--hide {
  display: none !important;
}

.split-rich-text-editor .vditor-content {
  height: calc(100% - 41px) !important;
}
.split-rich-text-editor .vditor-toolbar--hide ~ .vditor-content {
  height: 100% !important;
}

.split-rich-text-editor .vditor-wysiwyg {
  padding: 40px 48px 64px !important;
  background: transparent !important;
  color: #1f2937 !important;
  line-height: 2 !important;
}
.split-rich-text-editor .vditor-reset {
  color: #1f2937 !important;
}

/* 暗色主题 */
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar {
  background: rgba(30, 41, 59, 0.95) !important;
  border-bottom-color: rgba(71, 85, 105, 0.5) !important;
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-wysiwyg {
  color: #e5e7eb !important;
  background: transparent !important;
}

/* 禁用状态 */
.split-rich-text-editor.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
