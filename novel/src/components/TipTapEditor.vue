<template>
  <div class="tiptap-editor">
    <div class="editor-toolbar">
      <div class="toolbar-group">
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive('bold') }"
          title="加粗"
          @click="editor?.chain().focus().toggleBold().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M185.08 114.46A48 48 0 0 0 148 36H80a12 12 0 0 0-12 12v152a12 12 0 0 0 12 12h80a52 52 0 0 0 25.08-97.54M92 60h56a24 24 0 0 1 0 48H92Zm68 128H92v-56h68a28 28 0 0 1 0 56"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive('italic') }"
          title="斜体"
          @click="editor?.chain().focus().toggleItalic().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M200 56a8 8 0 0 1-8 8h-34.23L115.1 192H144a8 8 0 0 1 0 16H64a8 8 0 0 1 0-16h34.23L140.9 64H112a8 8 0 0 1 0-16h80a8 8 0 0 1 8 8"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive('underline') }"
          title="下划线"
          @click="editor?.chain().focus().toggleUnderline().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M200 224a8 8 0 0 1-8 8H64a8 8 0 0 1 0-16h128a8 8 0 0 1 8 8m-72-24a64.07 64.07 0 0 0 64-64V56a8 8 0 0 0-16 0v80a48 48 0 0 1-96 0V56a8 8 0 0 0-16 0v80a64.07 64.07 0 0 0 64 64"/>
          </svg>
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <div class="toolbar-group">
        <el-dropdown trigger="click" @command="handleHeading">
          <button type="button" class="toolbar-btn heading-btn" title="标题样式">
            <span>{{ headingLabel }}</span>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="12" height="12" style="margin-left: 4px;">
              <path fill="currentColor" d="m213.66 101.66l-80 80a8 8 0 0 1-11.32 0l-80-80a8 8 0 0 1 11.32-11.32L128 164.69l74.34-74.35a8 8 0 0 1 11.32 11.32"/>
            </svg>
          </button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="0">正文</el-dropdown-item>
              <el-dropdown-item command="1">标题 1</el-dropdown-item>
              <el-dropdown-item command="2">标题 2</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <div class="toolbar-divider"></div>

      <div class="toolbar-group">
        <el-color-picker
          :model-value="currentTextColor"
          size="small"
          @change="handleColorChange"
          @active-change="handleColorChange"
        >
          <template #default>
            <button type="button" class="toolbar-btn color-btn" title="字体颜色">
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
                <path fill="currentColor" d="M234.53 139.07a8 8 0 0 0 3.13-13.24L122.17 10.34a8 8 0 0 0-11.31 0L70.25 51l-24.6-24.66a8 8 0 0 0-11.31 11.32l24.6 24.6L15 106.17a24 24 0 0 0 0 33.94L99.89 225a24 24 0 0 0 33.94 0l78.49-78.49Zm-32.19-5.24l-79.83 79.83a8 8 0 0 1-11.31 0L26.34 128.8a8 8 0 0 1 0-11.31l43.91-43.92l29.12 29.12a28 28 0 1 0 11.31-11.32L81.57 62.26l35-34.95L217.19 128l-11.72 3.9a8.1 8.1 0 0 0-3.13 1.93m-86.83-26.31a13.26 13.26 0 1 1-.05.06s.05-.05.05-.06"/>
              </svg>
              <span class="color-indicator" :style="{ backgroundColor: currentTextColor || '#000' }"></span>
            </button>
          </template>
        </el-color-picker>
      </div>

      <div class="toolbar-divider"></div>

      <div class="toolbar-group">
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive({ textAlign: 'left' }) }"
          title="左对齐"
          @click="editor?.chain().focus().setTextAlign('left').run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M32 64a8 8 0 0 1 8-8h176a8 8 0 0 1 0 16H40a8 8 0 0 1-8-8m8 48h128a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16m0 48h176a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16m0 48h128a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive({ textAlign: 'center' }) }"
          title="居中对齐"
          @click="editor?.chain().focus().setTextAlign('center').run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M32 64a8 8 0 0 1 8-8h176a8 8 0 0 1 0 16H40a8 8 0 0 1-8-8m24 48h128a8 8 0 0 1 0 16H56a8 8 0 0 1 0-16m-24 48h176a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16m24 48h128a8 8 0 0 1 0 16H56a8 8 0 0 1 0-16"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive({ textAlign: 'right' }) }"
          title="右对齐"
          @click="editor?.chain().focus().setTextAlign('right').run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M32 64a8 8 0 0 1 8-8h176a8 8 0 0 1 0 16H40a8 8 0 0 1-8-8m8 48h176a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16m0 48h128a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16m0 48h176a8 8 0 0 1 0 16H40a8 8 0 0 1 0-16"/>
          </svg>
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <div class="toolbar-group">
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive('bulletList') }"
          title="无序列表"
          @click="editor?.chain().focus().toggleBulletList().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M80 64a8 8 0 0 1 8-8h128a8 8 0 0 1 0 16H88a8 8 0 0 1-8-8m0 64a8 8 0 0 1 8-8h128a8 8 0 0 1 0 16H88a8 8 0 0 1-8-8m0 64a8 8 0 0 1 8-8h128a8 8 0 0 1 0 16H88a8 8 0 0 1-8-8M40 68a12 12 0 1 1-12 12a12 12 0 0 1 12-12m0 64a12 12 0 1 1-12 12a12 12 0 0 1 12-12m0 64a12 12 0 1 1-12 12a12 12 0 0 1 12-12"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          :class="{ 'is-active': editor?.isActive('orderedList') }"
          title="有序列表"
          @click="editor?.chain().focus().toggleOrderedList().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M208 72H104a8 8 0 0 1 0-16h104a8 8 0 0 1 0 16m0 64H104a8 8 0 0 1 0-16h104a8 8 0 0 1 0 16m0 64H104a8 8 0 0 1 0-16h104a8 8 0 0 1 0 16M72 64v120a8 8 0 0 1-16 0V64a8 8 0 0 1 16 0m-8-24a8 8 0 1 0 8 8a8 8 0 0 0-8-8m-8 152v8a8 8 0 0 0 16 0v-8a8 8 0 0 0-16 0m8-48a24 24 0 0 0-8 46.63V200a8 8 0 0 0 16 0v-9.37A24 24 0 0 0 64 144"/>
          </svg>
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <div class="toolbar-group">
        <button
          type="button"
          class="toolbar-btn"
          title="插入链接"
          @click="handleLink"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M200 104a8 8 0 0 1-8 8h-64a8 8 0 0 1 0-16h64a8 8 0 0 1 8 8m-8 40h-64a8 8 0 0 0 0 16h64a8 8 0 0 0 0-16m80-32a80 80 0 0 1-80 80H88a80 80 0 0 1 0-160h104a80 80 0 0 1 80 80m-80-64H88a64 64 0 0 0 0 128h104a64 64 0 0 0 0-128"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          title="插入图片"
          @click="handleImage"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M216 40H40a16 16 0 0 0-16 16v144a16 16 0 0 0 16 16h176a16 16 0 0 0 16-16V56a16 16 0 0 0-16-16m0 160H40V56h176zm-32-80a32 32 0 1 1-32-32a32 32 0 0 1 32 32m-24 0a8 8 0 1 0-8 8a8 8 0 0 0 8-8m-56 48a8 8 0 0 0 0-16H64a8 8 0 0 0 0 16h40"/>
          </svg>
        </button>
      </div>

      <div class="toolbar-divider"></div>

      <div class="toolbar-group">
        <button
          type="button"
          class="toolbar-btn"
          :disabled="!editor?.can().undo()"
          title="撤销"
          @click="editor?.chain().focus().undo().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M232 144a64.07 64.07 0 0 1-64 64H80a8 8 0 0 1 0-16h88a48 48 0 0 0 0-96H51.31l34.35 34.34a8 8 0 0 1-11.32 11.32l-48-48a8 8 0 0 1 0-11.32l48-48a8 8 0 0 1 11.32 11.32L51.31 80H168a64.07 64.07 0 0 1 64 64"/>
          </svg>
        </button>
        <button
          type="button"
          class="toolbar-btn"
          :disabled="!editor?.can().redo()"
          title="重做"
          @click="editor?.chain().focus().redo().run()"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 256" width="16" height="16">
            <path fill="currentColor" d="M170.34 130.34L204.69 96H88a48 48 0 0 0 0 96h88a8 8 0 0 1 0 16H88a64 64 0 0 1 0-128h116.69l-34.35-34.34a8 8 0 0 1 11.32-11.32l48 48a8 8 0 0 1 0 11.32l-48 48a8 8 0 0 1-11.32-11.32"/>
          </svg>
        </button>
      </div>
    </div>

    <div class="editor-content">
      <EditorContent :editor="editor as any" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Editor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import TextAlign from '@tiptap/extension-text-align'
import Color from '@tiptap/extension-color'
import { TextStyle } from '@tiptap/extension-text-style'
import Link from '@tiptap/extension-link'
import Image from '@tiptap/extension-image'
import Placeholder from '@tiptap/extension-placeholder'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
}>(), {
  placeholder: '请输入正文内容...',
  disabled: false
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'blur'): void
}>()

const editor = ref<Editor | undefined>(undefined)

const headingLabel = computed(() => {
  if (editor.value?.isActive('heading', { level: 1 })) return '标题 1'
  if (editor.value?.isActive('heading', { level: 2 })) return '标题 2'
  return '正文'
})

const currentTextColor = computed(() => {
  return editor.value?.getAttributes('textStyle').color || ''
})

const initEditor = () => {
  editor.value = new Editor({
    content: props.modelValue,
    editable: !props.disabled,
    extensions: [
      StarterKit,
      Underline,
      TextAlign.configure({
        types: ['heading', 'paragraph'],
      }),
      Color,
      TextStyle,
      Link.configure({
        openOnClick: false,
      }),
      Image,
      Placeholder.configure({
        placeholder: props.placeholder,
      }),
    ],
    onUpdate: ({ editor }) => {
      const html = editor.getHTML()
      emit('update:modelValue', html)
    },
    onBlur: () => {
      emit('blur')
    },
  })
}

const handleHeading = (level: string) => {
  const lvl = parseInt(level)
  if (lvl === 0) {
    editor.value?.chain().focus().setParagraph().run()
  } else {
    editor.value?.chain().focus().toggleHeading({ level: lvl as 1 | 2 }).run()
  }
}

const handleColorChange = (color: string | null) => {
  if (color) {
    editor.value?.chain().focus().setColor(color).run()
  } else {
    editor.value?.chain().focus().unsetColor().run()
  }
}

const handleLink = () => {
  const url = window.prompt('请输入链接地址')?.trim()
  if (!url) return
  
  if (editor.value?.isActive('link')) {
    editor.value.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
  } else {
    editor.value?.chain().focus().setLink({ href: url }).run()
  }
}

const handleImage = () => {
  const url = window.prompt('请输入图片地址')?.trim()
  if (!url) return
  editor.value?.chain().focus().setImage({ src: url }).run()
}

watch(() => props.modelValue, (value) => {
  if (editor.value && editor.value.getHTML() !== value) {
    editor.value.commands.setContent(value, { emitUpdate: false })
  }
})

watch(() => props.disabled, (disabled) => {
  if (editor.value) {
    editor.value.setEditable(!disabled)
  }
})

onMounted(() => {
  nextTick(() => {
    initEditor()
  })
})

onBeforeUnmount(() => {
  if (editor.value) {
    editor.value.destroy()
    editor.value = undefined
  }
})
</script>

<style scoped>
.tiptap-editor {
  display: flex;
  flex-direction: column;
  border: 1px solid rgb(224, 224, 230);
  border-radius: 6px;
  background: #ffffff;
  overflow: hidden;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  border-bottom: 1px solid rgb(224, 224, 230);
  background: #fafafa;
  flex-wrap: wrap;
}

.toolbar-group {
  display: flex;
  align-items: center;
  gap: 2px;
}

.toolbar-divider {
  width: 1px;
  height: 20px;
  margin: 0 6px;
  background: rgb(224, 224, 230);
}

.toolbar-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 28px;
  height: 28px;
  padding: 0 6px;
  border: none;
  border-radius: 3px;
  background: transparent;
  color: #475569;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.toolbar-btn:hover:not(:disabled) {
  background: #ffffff;
  color: #18a058;
}

.toolbar-btn.is-active {
  background: rgba(24, 160, 88, 0.1);
  color: #18a058;
}

.toolbar-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.heading-btn {
  min-width: 60px;
  justify-content: space-between;
}

.color-btn {
  position: relative;
  min-width: 32px;
}

.color-indicator {
  position: absolute;
  bottom: 2px;
  left: 50%;
  transform: translateX(-50%);
  width: 14px;
  height: 3px;
  border-radius: 1px;
}

.editor-content {
  flex: 1;
  min-height: 300px;
  max-height: 500px;
  overflow-y: auto;
  padding: 16px;
}

.editor-content :deep(.tiptap) {
  outline: none;
  min-height: 100%;
}

.editor-content :deep(.tiptap p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: #94a3b8;
  pointer-events: none;
  height: 0;
}

.editor-content :deep(.tiptap p) {
  margin: 0 0 12px 0;
  line-height: 1.7;
}

.editor-content :deep(.tiptap h1) {
  font-size: 24px;
  font-weight: 600;
  margin: 16px 0 8px 0;
  line-height: 1.3;
}

.editor-content :deep(.tiptap h2) {
  font-size: 20px;
  font-weight: 600;
  margin: 14px 0 6px 0;
  line-height: 1.3;
}

.editor-content :deep(.tiptap ul),
.editor-content :deep(.tiptap ol) {
  padding-left: 24px;
  margin: 8px 0;
}

.editor-content :deep(.tiptap li) {
  margin: 4px 0;
}

.editor-content :deep(.tiptap a) {
  color: #18a058;
  text-decoration: underline;
}

.editor-content :deep(.tiptap img) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  margin: 8px 0;
}

.editor-content :deep(.tiptap blockquote) {
  border-left: 3px solid #18a058;
  padding-left: 16px;
  margin: 12px 0;
  color: #64748b;
}

.editor-content :deep(.tiptap code) {
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', monospace;
  font-size: 13px;
}

.editor-content :deep(.tiptap pre) {
  background: #1e293b;
  color: #e2e8f0;
  padding: 12px 16px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 12px 0;
}

.editor-content :deep(.tiptap pre code) {
  background: none;
  padding: 0;
  color: inherit;
}

@media (prefers-color-scheme: dark) {
  .tiptap-editor {
    border-color: rgba(255, 255, 255, 0.12);
    background: #1e1e2e;
  }

  .editor-toolbar {
    border-bottom-color: rgba(255, 255, 255, 0.12);
    background: #252536;
  }

  .toolbar-divider {
    background: rgba(255, 255, 255, 0.15);
  }

  .toolbar-btn {
    color: #cbd5e1;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.08);
    color: #4ade80;
  }

  .toolbar-btn.is-active {
    background: rgba(74, 222, 128, 0.15);
    color: #4ade80;
  }

  .heading-btn {
    color: #cbd5e1;
  }

  .color-btn {
    color: #cbd5e1;
  }

  .editor-content {
    background: #1e1e2e;
  }

  .editor-content :deep(.tiptap p.is-editor-empty:first-child::before) {
    color: #64748b;
  }

  .editor-content :deep(.tiptap p) {
    color: #e2e8f0;
  }

  .editor-content :deep(.tiptap h1),
  .editor-content :deep(.tiptap h2) {
    color: #f1f5f9;
  }

  .editor-content :deep(.tiptap a) {
    color: #4ade80;
  }

  .editor-content :deep(.tiptap img) {
    opacity: 0.95;
  }

  .editor-content :deep(.tiptap blockquote) {
    border-left-color: #4ade80;
    color: #94a3b8;
  }

  .editor-content :deep(.tiptap code) {
    background: rgba(255, 255, 255, 0.08);
    color: #fbbf24;
  }

  .editor-content :deep(.tiptap pre) {
    background: #0f172a;
    color: #e2e8f0;
  }
}
</style>
