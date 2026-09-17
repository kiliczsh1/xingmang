<template>
  <div class="pdf-gallery">
    <div v-if="loading" class="pdf-gallery-state">
      <el-skeleton :rows="4" animated />
    </div>

    <div v-else-if="error" class="pdf-gallery-state error">
      <span>{{ error }}</span>
    </div>

    <div v-else-if="pages.length" class="pdf-gallery-list" :class="{ compact }">
      <figure v-for="page in pages" :key="page.pageNumber" class="pdf-page-card">
        <img :src="page.src" :alt="`PDF 第 ${page.pageNumber} 页`" class="pdf-page-image" />
        <figcaption class="pdf-page-caption">第 {{ page.pageNumber }} 页</figcaption>
      </figure>
    </div>

    <div v-else class="pdf-gallery-state">
      <span>暂无可展示的 PDF 页面</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import * as pdfjsLib from 'pdfjs-dist'
import workerSrc from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

type PdfPagePreview = {
  pageNumber: number
  src: string
}

const props = withDefaults(defineProps<{
  src?: string | null
  scale?: number
  compact?: boolean
}>(), {
  src: '',
  scale: 1.4,
  compact: false
})

pdfjsLib.GlobalWorkerOptions.workerSrc = workerSrc

const loading = ref(false)
const error = ref('')
const pages = ref<PdfPagePreview[]>([])

let renderToken = 0
let currentTask: ReturnType<typeof pdfjsLib.getDocument> | null = null

const resetState = () => {
  error.value = ''
  pages.value = []
}

const loadPages = async () => {
  renderToken += 1
  const currentToken = renderToken

  currentTask?.destroy()
  currentTask = null
  resetState()

  if (!props.src) {
    return
  }

  loading.value = true

  try {
    currentTask = pdfjsLib.getDocument(props.src)
    const pdf = await currentTask.promise

    try {
      const nextPages: PdfPagePreview[] = []

      for (let pageNumber = 1; pageNumber <= pdf.numPages; pageNumber += 1) {
        if (currentToken !== renderToken) return

        const page = await pdf.getPage(pageNumber)
        const viewport = page.getViewport({ scale: props.scale })
        const canvas = document.createElement('canvas')
        const context = canvas.getContext('2d')

        if (!context) {
          continue
        }

        canvas.width = Math.ceil(viewport.width)
        canvas.height = Math.ceil(viewport.height)

        await page.render({
          canvas,
          canvasContext: context,
          viewport
        }).promise

        nextPages.push({
          pageNumber,
          src: canvas.toDataURL('image/png')
        })

        if (typeof page.cleanup === 'function') {
          page.cleanup()
        }
      }

      if (currentToken === renderToken) {
        pages.value = nextPages
      }
    } finally {
      await pdf.destroy()
    }
  } catch (err: any) {
    if (currentToken === renderToken) {
      error.value = err?.message || 'PDF 页面渲染失败'
    }
  } finally {
    if (currentToken === renderToken) {
      loading.value = false
    }
    currentTask = null
  }
}

watch(
  () => [props.src, props.scale],
  () => {
    loadPages()
  },
  { immediate: true }
)

onBeforeUnmount(() => {
  renderToken += 1
  currentTask?.destroy()
})
</script>

<style scoped>
.pdf-gallery {
  width: 100%;
}

.pdf-gallery-state {
  padding: 18px;
  border-radius: 16px;
  color: #64748b;
  background: #f8fafc;
  border: 1px dashed rgba(148, 163, 184, 0.4);
}

.pdf-gallery-state.error {
  color: #b42318;
  background: #fef3f2;
  border-color: rgba(217, 45, 32, 0.25);
}

.pdf-gallery-list {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.pdf-gallery-list.compact {
  gap: 12px;
}

.pdf-page-card {
  margin: 0;
  overflow: hidden;
  border-radius: 18px;
  background: #ffffff;
  border: 1px solid rgba(148, 163, 184, 0.18);
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.06);
}

.pdf-page-image {
  display: block;
  width: 100%;
  height: auto;
  background: #ffffff;
}

.pdf-page-caption {
  padding: 10px 14px 12px;
  color: #64748b;
  font-size: 12px;
  text-align: center;
}
</style>
