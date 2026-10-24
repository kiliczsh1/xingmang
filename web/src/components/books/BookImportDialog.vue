<template>
  <el-dialog
    v-model="importDialogVisible"
    title="导入作品"
    :width="hasFiles ? '760px' : '500px'"
    :top="hasFiles ? '6vh' : '15vh'"
    class="book-import-dialog"
    :close-on-click-modal="false"
    append-to-body
  >
    <div
      class="import-dialog-body"
      @dragenter.prevent="handleDragEnter"
      @dragleave.prevent="handleDragLeave"
      @dragover.prevent
      @drop.prevent="handleFileDrop"
    >
      <!-- 没选文件时才展示大拖拽区；选完之后收起，只留文件卡片与预览 -->
      <div
        v-if="fileItems.length === 0"
        class="import-drop-zone"
        :class="{ 'is-dragover': isDragOver }"
        @click="triggerFileSelect"
      >
        <div class="import-drop-content">
          <el-icon class="import-icon"><Upload /></el-icon>
          <p class="import-text">拖拽文件到此处，或点击选择文件</p>
          <p class="import-hint">支持 TXT / JSON，可同时选择多个文件</p>
        </div>
      </div>

      <template v-else>
        <div class="files-header">
          <span class="files-summary">
            已选 <b>{{ fileItems.length }}</b> 个文件 · 共 <b>{{ totalChapterCount }}</b> 章
            <span v-if="parsingCount > 0" class="parsing-hint">（解析中 {{ parsingCount }}）</span>
          </span>
          <div class="files-header-ops">
            <el-button link type="primary" @click="triggerFileSelect">添加文件</el-button>
            <el-button link type="danger" @click="clearSelectedFiles">清空</el-button>
          </div>
        </div>

        <div class="file-list">
          <div v-for="item in fileItems" :key="item.uid" class="file-card">
            <div class="file-card-head">
              <el-icon class="file-type-icon" :class="{ 'is-json': item.kind === 'json' }">
                <Document />
              </el-icon>
              <span class="file-name" :title="item.name">{{ item.name }}</span>
              <span class="file-size">{{ formatFileSize(item.size) }}</span>
              <el-icon class="file-remove" @click="removeFile(item.uid)"><Close /></el-icon>
            </div>

            <div v-if="item.loading" class="file-state">正在解析章节…</div>
            <div v-else-if="item.error" class="file-state is-error">{{ item.error }}</div>
            <template v-else>
              <div class="file-card-meta">
                <span class="meta-chapters">{{ item.chapterCount }} 章</span>
                <span class="meta-dot">·</span>
                <span>{{ item.totalChars.toLocaleString() }} 字</span>
                <span class="meta-dot">·</span>
                <span class="meta-split">{{ item.splitModeLabel }}</span>
              </div>

              <div v-if="item.samples.length > 0" class="chapter-preview">
                <div class="preview-head">
                  预览前 {{ item.samples.length }} 章
                  <span v-if="item.chapterCount > item.samples.length" class="preview-rest">
                    （共 {{ item.chapterCount }} 章，其余导入后可见）
                  </span>
                </div>
                <div class="preview-list">
                  <div v-for="(sample, index) in item.samples" :key="index" class="preview-row">
                    <span class="preview-index">{{ index + 1 }}</span>
                    <div class="preview-main">
                      <div class="preview-name">{{ sample.title }}</div>
                      <div class="preview-excerpt">{{ sample.excerpt }}</div>
                    </div>
                    <span class="preview-chars">{{ sample.chars }} 字</span>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </template>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="importing"
          :disabled="fileItems.length === 0 || parsingCount > 0"
          @click="confirmImport"
        >
          确认导入{{ totalChapterCount > 0 ? `（${totalChapterCount} 章）` : '' }}
        </el-button>
      </div>
    </template>
  </el-dialog>

  <input
    ref="importFileInput"
    class="book-import-input"
    type="file"
    accept=".txt,text/plain,.json,application/json"
    multiple
    @change="handleImportFileSelect"
  />
</template>

<script setup lang="ts">
// 导入作品弹窗（拖拽 / 多选 / 编码探测 / 按章切分 / 导入前预览）
// 2026-09-21 第 3 期从 views/Books.vue 拆出；隐藏 input 一并搬入，开关由父组件传入
// 2026-10-24：选完文件即解析，隐藏大拖拽区，改为「文件卡片 + 前 10 章预览」
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload, Document, Close } from '@element-plus/icons-vue'
import { useBookStore } from '@/stores/book'
import { chapterAPI, bookAPI } from '@/api'
import { readTxtFile, splitTxtIntoChapters, formatFileSize, CHAPTER_HEADING_REGEX } from '@/utils/bookImport'
import type { ParsedImportChapter } from '@/utils/bookImport'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>()

const bookStore = useBookStore()

const importDialogVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v)
})

const importFileInput = ref<HTMLInputElement | null>(null)
const importingBookId = ref<number | null>(null)
const importing = ref(false)
const isDragOver = ref(false)

const PREVIEW_LIMIT = 10

interface ChapterSample {
  title: string
  excerpt: string
  chars: number
}

interface ImportFileItem {
  uid: number
  file: File
  name: string
  size: number
  kind: 'txt' | 'json'
  loading: boolean
  error: string
  /** 章节正文，导入时直接复用，避免二次读盘解析 */
  chapters: ParsedImportChapter[]
  /** JSON 原文本，仍交给后端按它自己的兼容规则建书 */
  rawText: string
  chapterCount: number
  totalChars: number
  samples: ChapterSample[]
  splitModeLabel: string
}

const fileItems = ref<ImportFileItem[]>([])
let uidSeed = 0

const totalChapterCount = computed(() =>
  fileItems.value.reduce((sum, item) => sum + (item.loading || item.error ? 0 : item.chapterCount), 0)
)
const parsingCount = computed(() => fileItems.value.filter((item) => item.loading).length)
// 空态保持原来的 500px 弹窗；有文件才撑宽放预览
const hasFiles = computed(() => fileItems.value.length > 0)

const countChars = (text: string) => (text || '').replace(/\s/g, '').length

const buildSamples = (chapters: ParsedImportChapter[]): ChapterSample[] =>
  chapters.slice(0, PREVIEW_LIMIT).map((chapter) => {
    const plain = (chapter.content || '').replace(/\s+/g, ' ').trim()
    return {
      title: chapter.title,
      excerpt: plain.length > 90 ? plain.slice(0, 90) + '…' : plain || '（本章正文为空）',
      chars: countChars(chapter.content)
    }
  })

const createItem = (file: File): ImportFileItem => ({
  uid: ++uidSeed,
  file,
  name: file.name,
  size: file.size,
  kind: /\.json$/i.test(file.name) ? 'json' : 'txt',
  loading: true,
  error: '',
  chapters: [],
  rawText: '',
  chapterCount: 0,
  totalChars: 0,
  samples: [],
  splitModeLabel: ''
})

const parseTxtItem = async (item: ImportFileItem) => {
  const text = await readTxtFile(item.file)
  if (!text) {
    item.error = '文件内容为空，已跳过'
    return
  }
  const baseTitle = item.name.replace(/\.txt$/i, '').trim()
  const chapters = splitTxtIntoChapters(text, baseTitle)
  if (chapters.length === 0) {
    item.error = '没有解析出任何章节'
    return
  }
  // 用独立正则实例，避免共享 lastIndex
  const headingHits = Array.from(text.matchAll(new RegExp(CHAPTER_HEADING_REGEX.source, 'gim')))
  item.chapters = chapters
  item.chapterCount = chapters.length
  item.totalChars = chapters.reduce((sum, chapter) => sum + countChars(chapter.content), 0)
  item.samples = buildSamples(chapters)
  item.splitModeLabel = headingHits.length > 0 ? '按章节标题识别' : '未识别到标题，按段落切分'
}

const parseJsonItem = async (item: ImportFileItem) => {
  const raw = await item.file.text()
  let data: any
  try {
    data = JSON.parse(raw)
  } catch {
    item.error = '不是合法的 JSON 文件'
    return
  }
  // 与后端 ImportBookJSON 的兼容口径保持一致
  const meta = data && typeof data === 'object' && !Array.isArray(data) && data.book && typeof data.book === 'object'
    ? data.book
    : data
  let list: any[] = []
  if (Array.isArray(data)) list = data
  else if (Array.isArray(data?.chapters)) list = data.chapters
  else if (Array.isArray(meta?.chapters)) list = meta.chapters
  const chapters: ParsedImportChapter[] = list
    .filter((chapter) => chapter && typeof chapter === 'object')
    .map((chapter) => ({
      title: String(chapter.title || chapter.name || chapter.chapter_title || '').trim(),
      content: String(chapter.content || chapter.text || chapter.body || chapter.chapter_content || '')
    }))
    .filter((chapter) => chapter.title || chapter.content)
  if (chapters.length === 0) {
    item.error = 'JSON 里没有识别到章节'
    return
  }
  item.rawText = raw
  item.chapters = chapters
  item.chapterCount = chapters.length
  item.totalChars = chapters.reduce((sum, chapter) => sum + countChars(chapter.content), 0)
  item.samples = buildSamples(chapters)
  item.splitModeLabel = 'JSON 章节'
}

const parseItem = async (item: ImportFileItem) => {
  try {
    if (item.kind === 'json') await parseJsonItem(item)
    else await parseTxtItem(item)
  } catch (error: any) {
    item.error = error?.message ? `解析失败：${error.message}` : '解析失败'
  } finally {
    item.loading = false
  }
}

const addFiles = (files: File[]) => {
  const accepted = files.filter((file) => /\.(txt|json)$/i.test(file.name))
  if (accepted.length === 0) {
    ElMessage.warning('请选择 TXT 或 JSON 格式的文件')
    return
  }
  const fresh = accepted.filter(
    (file) => !fileItems.value.some((item) => item.name === file.name && item.size === file.size)
  )
  if (fresh.length === 0) {
    ElMessage.warning('这些文件已经在列表里了')
    return
  }
  const created = fresh.map(createItem)
  const startIndex = fileItems.value.length
  fileItems.value.push(...created)
  // ⚠ 必须解析「数组里的响应式代理」而不是 created 里的原始对象：
  // 改原始对象不会触发依赖通知，界面会一直停在「解析中」
  for (let i = 0; i < created.length; i++) {
    void parseItem(fileItems.value[startIndex + i])
  }
}

const triggerFileSelect = () => {
  if (importFileInput.value) {
    importFileInput.value.value = ''
    importFileInput.value.click()
  }
}

// 拖拽高亮用进出计数：子元素间移动会连续触发 dragenter/dragleave，直接覆盖会闪
let dragDepth = 0
const handleDragEnter = () => {
  dragDepth++
  isDragOver.value = true
}
const handleDragLeave = () => {
  dragDepth = Math.max(0, dragDepth - 1)
  if (dragDepth === 0) isDragOver.value = false
}

const handleFileDrop = (event: DragEvent) => {
  dragDepth = 0
  isDragOver.value = false
  const files = event.dataTransfer?.files
  if (files && files.length > 0) addFiles(Array.from(files))
}

const handleImportFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) addFiles(Array.from(target.files))
}

const removeFile = (uid: number) => {
  const index = fileItems.value.findIndex((item) => item.uid === uid)
  if (index > -1) fileItems.value.splice(index, 1)
}

const clearSelectedFiles = () => {
  fileItems.value = []
}

const confirmImport = async () => {
  if (fileItems.value.length === 0 || importing.value) return
  const usable = fileItems.value.filter((item) => !item.error && !item.loading)
  if (usable.length === 0) {
    ElMessage.warning('没有可导入的文件')
    return
  }

  const chapterTotal = totalChapterCount.value
  // 预览阶段已经是用户确认过了：这里不再弹二次确认，点了就直接收起弹窗，
  // 导入在后台继续，避免「点了确认还要等卡片出现弹窗才消失」。
  importDialogVisible.value = false
  fileItems.value = []
  isDragOver.value = false

  const tip = ElMessage({
    message: `正在导入 ${usable.length} 个文件，共 ${chapterTotal} 章…`,
    type: 'info',
    duration: 0
  })
  try {
    await processImportFiles(usable)
  } finally {
    tip.close()
  }
}

const processImportFiles = async (items: ImportFileItem[]) => {
  let totalChapters = 0
  let successCount = 0
  importingBookId.value = -1 // 使用一个特殊值表示正在导入新书
  importing.value = true

  try {
    for (const item of items) {
      // JSON：旧版本导出的作品文件，交给后端按它自己的兼容规则落成章节
      if (item.kind === 'json') {
        try {
          const res: any = await bookAPI.importBookJson(item.rawText, item.name)
          if (res?.success && res.data) {
            successCount++
            totalChapters += res.data.chapters || 0
          } else {
            ElMessage.error(`导入 "${item.name}" 失败：${res?.message || '未知错误'}`)
          }
        } catch (e: any) {
          ElMessage.error(`导入 "${item.name}" 失败：${e?.response?.data?.message || e?.message || '未知错误'}`)
        }
        continue
      }

      const baseTitle = item.name.replace(/\.txt$/i, '').trim()
      const newBook = await bookStore.createBook({ title: baseTitle, description: '由文件导入' })

      if (!newBook) {
        ElMessage.error(`为文件 "${item.name}" 创建作品失败`)
        continue
      }

      // 章节用选文件时解析好的结果，不再重复读文件
      if (item.chapters.length > 0) {
        const batchImport = await tryBatchImportChapters(newBook.id, item.chapters)
        if (batchImport.fallback) {
          await createChaptersSequentially(newBook.id, item.chapters, 0)
        }
        totalChapters += item.chapters.length
      }

      successCount++
    }

    await bookStore.fetchBooks()
    ElMessage.success(`成功导入 ${successCount} 个文件，共 ${totalChapters} 章（已落成磁盘文件）`)
  } catch (error) {
    if (error instanceof Error && error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    importingBookId.value = null
    importing.value = false
  }
}

const createChaptersSequentially = async (
  bookId: number,
  chapters: ParsedImportChapter[],
  startOrder: number
) => {
  for (const [index, chapter] of chapters.entries()) {
    await chapterAPI.create({
      book_id: bookId,
      title: chapter.title,
      content: chapter.content,
      order_num: startOrder + index,
      type: 'chapter'
    })
  }
}

const tryBatchImportChapters = async (bookId: number, chapters: ParsedImportChapter[]) => {
  const response = await fetch('/api/chapters/import-book', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ bookId, chapters })
  })
  if (response.status === 404) {
    return { fallback: true as const }
  }
  const result = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(result?.message || `Request failed with status code ${response.status}`)
  }
  return { fallback: false as const, result }
}

// 打开时清空上一次的选择（原 triggerBookImport 里的复位）
watch(() => props.visible, (v) => {
  if (v) {
    fileItems.value = []
    isDragOver.value = false
  }
})
</script>

<style scoped>
.book-import-input {
  display: none;
}

.import-dialog-body {
  max-height: 62vh;
  overflow-y: auto;
  padding-right: 4px;
}

/* ---------- 空态：大拖拽区 ---------- */
.import-drop-zone {
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  padding: 40px 20px;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #fafafa;
}

.import-drop-zone:hover {
  border-color: #8b5cf6;
  background: #f5f3ff;
}

.import-drop-zone.is-dragover {
  border-color: #8b5cf6;
  background: #ede9fe;
}

.import-drop-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.import-icon {
  font-size: 48px;
  color: #8b5cf6;
}

.import-text {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  margin: 0;
}

.import-hint {
  font-size: 12px;
  color: #999;
  margin: 0;
}

/* ---------- 文件列表 ---------- */
.files-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
  font-size: 13px;
  color: #475569;
}

.files-summary b {
  color: #1f2937;
  font-weight: 700;
}

.parsing-hint {
  color: #8b5cf6;
}

.files-header-ops {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}

.file-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.file-card {
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  padding: 12px 14px;
  background: #fff;
}

.file-card-head {
  display: flex;
  align-items: center;
  gap: 8px;
}

.file-type-icon {
  color: #8b5cf6;
  flex-shrink: 0;
}

.file-type-icon.is-json {
  color: #f59e0b;
}

.file-name {
  flex: 1;
  font-size: 13px;
  font-weight: 600;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: #999;
  flex-shrink: 0;
}

.file-remove {
  cursor: pointer;
  color: #999 !important;
  transition: color 0.2s;
  flex-shrink: 0;
}

.file-remove:hover {
  color: #ef4444 !important;
}

.file-state {
  margin-top: 8px;
  font-size: 12px;
  color: #8b5cf6;
}

.file-state.is-error {
  color: #ef4444;
}

.file-card-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  font-size: 12px;
  color: #64748b;
}

.meta-chapters {
  font-weight: 700;
  color: #7c3aed;
}

.meta-dot {
  color: #cbd5e1;
}

.meta-split {
  color: #94a3b8;
}

/* ---------- 前 10 章预览 ---------- */
.chapter-preview {
  margin-top: 10px;
  border-top: 1px dashed #eef0f3;
  padding-top: 10px;
}

.preview-head {
  font-size: 12px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 8px;
}

.preview-rest {
  font-weight: 400;
  color: #94a3b8;
}

.preview-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.preview-row {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 7px 10px;
  border-radius: 8px;
  background: #f8fafc;
}

.preview-index {
  width: 18px;
  flex-shrink: 0;
  font-size: 11px;
  font-weight: 700;
  color: #94a3b8;
  text-align: center;
  padding-top: 2px;
}

.preview-main {
  flex: 1;
  min-width: 0;
}

.preview-name {
  font-size: 12.5px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview-excerpt {
  margin-top: 2px;
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.preview-chars {
  flex-shrink: 0;
  font-size: 11px;
  color: #94a3b8;
  padding-top: 2px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* ---------- 暗色 ---------- */
:root[data-theme='dark'] .import-drop-zone {
  background: rgba(22, 27, 34, 0.7);
  border-color: #303849;
}

:root[data-theme='dark'] .import-drop-zone:hover,
:root[data-theme='dark'] .import-drop-zone.is-dragover {
  border-color: #8b5cf6;
  background: rgba(139, 92, 246, 0.12);
}

:root[data-theme='dark'] .import-text {
  color: #e6edf3;
}

:root[data-theme='dark'] .import-hint,
:root[data-theme='dark'] .file-size,
:root[data-theme='dark'] .preview-rest {
  color: #6b7280;
}

:root[data-theme='dark'] .files-header {
  color: #9ca3af;
}

:root[data-theme='dark'] .files-summary b {
  color: #e6edf3;
}

:root[data-theme='dark'] .file-card {
  background: #1a2130;
  border-color: #2a3242;
}

:root[data-theme='dark'] .file-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .file-state {
  color: #a78bfa;
}

:root[data-theme='dark'] .file-state.is-error {
  color: #f87171;
}

:root[data-theme='dark'] .file-card-meta {
  color: #8b95a5;
}

:root[data-theme='dark'] .meta-chapters {
  color: #a78bfa;
}

:root[data-theme='dark'] .meta-dot {
  color: #3f4a5f;
}

:root[data-theme='dark'] .chapter-preview {
  border-top-color: #2a3242;
}

:root[data-theme='dark'] .preview-head {
  color: #cbd5e1;
}

:root[data-theme='dark'] .preview-row {
  background: rgba(42, 50, 66, 0.45);
}

:root[data-theme='dark'] .preview-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .preview-excerpt {
  color: #9ca3af;
}

:root[data-theme='dark'] .preview-index,
:root[data-theme='dark'] .preview-chars {
  color: #6b7280;
}
</style>

<style>
/* el-dialog 由 Element Plus 自己渲染，拿不到 scoped 的 data-v，只能写在非 scoped 块里。
   空态 500px → 有文件 760px 的切换做个过渡，避免跳一下。 */
.book-import-dialog {
  transition: width 0.22s ease;
}
</style>
