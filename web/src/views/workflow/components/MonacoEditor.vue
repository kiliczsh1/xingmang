<template>
  <div ref="hostRef" class="monaco-editor-container"></div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
// 走项目自建的精简入口（monaco-lite.ts），而不是 monaco 官方的 editor.main.js —— 后者会把
// css/html/typescript 语言服务连同它们的 worker（合计 8.4 MB）一起打进产物。详见该文件头部注释。
import * as monaco from '../monaco-lite'
// ⚠ 只保留本项目**真正用到**的 worker。
//
// 工作流 IDE 打开的文件清一色是 *.workflow.json（WorkflowIde / WorkflowRun 均传 language="json"），
// 所以只需要 json（校验/补全）与 editor（兜底，纯文本编辑）。
// 注意：真正决定「哪些 worker 会被打进产物」的不是这里，而是 monaco-lite.ts ——
// monaco 内部用 `new Worker(new URL('x.worker.js', import.meta.url))` 静态引用 worker，
// 只要 register 被 import，worker 就一定会被构建出来，与运行时是否走到这里无关。
import editorWorker from 'monaco-editor/editor/editor.worker.js?worker'
import jsonWorker from 'monaco-editor/language/json/json.worker.js?worker'

// 源版同样以 worker 跑 Monaco（dist/monacoeditorwork）
self.MonacoEnvironment = {
  getWorker(_workerId: string, label: string) {
    if (label === 'json') return new jsonWorker()
    return new editorWorker()
  }
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    language?: string
    readOnly?: boolean
    fontSize?: number
    fontFamily?: string
    dark?: boolean
  }>(),
  { language: 'plaintext', readOnly: false, fontSize: 14, fontFamily: "Consolas, 'Courier New', monospace", dark: false }
)

const emit = defineEmits<{ (e: 'update:modelValue', v: string): void; (e: 'change', v: string): void }>()

const hostRef = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null
let suppress = false

onMounted(() => {
  editor = monaco.editor.create(hostRef.value!, {
    value: props.modelValue,
    language: props.language,
    theme: props.dark ? 'vs-dark' : 'vs',
    fontSize: props.fontSize,
    fontFamily: props.fontFamily,
    automaticLayout: true,
    minimap: { enabled: true },
    scrollBeyondLastLine: false,
    readOnly: props.readOnly,
    tabSize: 2,
    renderLineHighlight: 'all',
    smoothScrolling: true,
    cursorBlinking: 'smooth'
  })
  editor.onDidChangeModelContent(() => {
    if (suppress) return
    const v = editor!.getValue()
    emit('update:modelValue', v)
    emit('change', v)
  })
})

watch(
  () => props.modelValue,
  v => {
    if (!editor || v === editor.getValue()) return
    suppress = true
    editor.setValue(v ?? '')
    suppress = false
  }
)

watch(
  () => props.language,
  v => {
    const model = editor?.getModel()
    if (model && v) monaco.editor.setModelLanguage(model, v)
  }
)

watch(
  () => props.dark,
  v => monaco.editor.setTheme(v ? 'vs-dark' : 'vs')
)

watch(
  () => props.fontSize,
  v => editor?.updateOptions({ fontSize: v })
)
watch(
  () => props.fontFamily,
  v => editor?.updateOptions({ fontFamily: v })
)
watch(
  () => props.readOnly,
  v => editor?.updateOptions({ readOnly: v })
)

onBeforeUnmount(() => editor?.dispose())

defineExpose({ getEditor: () => editor })
</script>

<style scoped>
.monaco-editor-container {
  width: 100%;
  height: 100%;
  min-height: 0;
  font-size: v-bind('props.fontSize + "px"');
  font-family: v-bind('props.fontFamily');
}
</style>
