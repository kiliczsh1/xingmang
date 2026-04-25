<template>
  <div class="books-container">
    <div class="header">
      <h2>作品管理</h2>
      <div class="header-actions">
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建作品
        </el-button>
      </div>
    </div>

    <div class="books-grid">
      <div v-for="book in bookStore.books" :key="book.id" class="book-card" @click="handleEdit(book)">
        <div class="card-title-section">
          <el-tooltip :content="importingBookId === book.id ? '导入中...' : '导入作品'" placement="top">
            <div
              class="type-badge"
              :class="{ loading: importingBookId === book.id }"
              @click.stop="triggerBookImport(book)"
              style="cursor: pointer;"
            >
              <span>{{ importingBookId === book.id ? '导入中...' : '导入作品' }}</span>
            </div>
          </el-tooltip>
          <h3 class="book-title">{{ book.title }}</h3>
        </div>

        <div class="card-body-section">
          <div class="book-cover">
            <img v-if="book.cover" :src="book.cover" class="cover-image" alt="封面" />
            <div v-else class="cover-placeholder">
              <el-icon class="cover-icon"><Notebook /></el-icon>
            </div>
            <div class="cover-tag">书测</div>
          </div>
          <div class="book-description">
            <p>{{ book.description || '暂无简介' }}</p>
          </div>
        </div>

        <div class="card-bottom-section" @click.stop>
          <div class="timestamp">
            <div class="date">{{ formatDateShort(book.created_at) }}</div>
            <div class="time">{{ formatTime(book.created_at) }}</div>
          </div>
          <div class="action-buttons">
            <el-tooltip content="创作" placement="top">
              <button class="action-btn create" @click="handleWrite(book.id)">
                <el-icon><EditPen /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="封面" placement="top">
              <button class="action-btn cover" @click.stop="handleOpenCoverDialog(book)">
                <el-icon><Picture /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="下载" placement="top">
              <button class="action-btn download" @click="handleDownload(book)">
                <el-icon><Download /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="确认" placement="top">
              <button class="action-btn confirm" @click="handleEdit(book)">
                <el-icon><Check /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="信息" placement="top">
              <button class="action-btn info">
                <el-icon><InfoFilled /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <button class="action-btn delete" @click="handleDelete(book)">
                <el-icon><Delete /></el-icon>
              </button>
            </el-tooltip>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑作品' : '创建新作品'"
      width="420px"
      class="simple-book-dialog"
    >
      <div class="simple-form">
        <div class="form-group">
          <label class="form-label required">作品名称</label>
          <el-input v-model="formData.title" placeholder="输入作品名称" />
        </div>
        <div class="form-group">
          <label class="form-label">作品简介</label>
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="4"
            placeholder="作品简介（可选）"
            resize="none"
          />
        </div>
      </div>
      <template #footer>
        <div class="simple-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 导入文件弹窗 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入作品"
      width="500px"
      :close-on-click-modal="false"
    >
      <div
        class="import-drop-zone"
        :class="{ 'is-dragover': isDragOver }"
        @dragenter.prevent="isDragOver = true"
        @dragleave.prevent="isDragOver = false"
        @dragover.prevent
        @drop.prevent="handleFileDrop"
        @click="triggerFileSelect"
      >
        <div class="import-drop-content">
          <el-icon class="import-icon"><Upload /></el-icon>
          <p class="import-text">拖拽文件到此处，或点击选择文件</p>
          <p class="import-hint">支持 TXT 格式，可同时选择多个文件</p>
        </div>
      </div>

      <div v-if="selectedFiles.length > 0" class="selected-files">
        <div class="files-header">
          <span>已选择 {{ selectedFiles.length }} 个文件</span>
          <el-button type="danger" link @click="clearSelectedFiles">清空</el-button>
        </div>
        <div class="files-list">
          <div v-for="(file, index) in selectedFiles" :key="index" class="file-item">
            <el-icon><Document /></el-icon>
            <span class="file-name">{{ file.name }}</span>
            <span class="file-size">{{ formatFileSize(file.size) }}</span>
            <el-icon class="file-remove" @click="removeFile(index)"><Close /></el-icon>
          </div>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="importDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :disabled="selectedFiles.length === 0"
            @click="confirmImport"
          >
            确认导入
          </el-button>
        </div>
      </template>
    </el-dialog>

    <input
      ref="importFileInput"
      class="book-import-input"
      type="file"
      accept=".txt,text/plain"
      multiple
      @change="handleImportFileSelect"
    />

    <el-dialog
      v-model="coverDialogVisible"
      title="设置封面"
      width="500px"
      :close-on-click-modal="false"
    >
      <div class="cover-dialog-content">
        <div
          class="cover-drop-zone"
          :class="{ 'is-dragover': isCoverDragOver, 'has-image': coverPreview }"
          @dragenter.prevent="isCoverDragOver = true"
          @dragleave.prevent="isCoverDragOver = false"
          @dragover.prevent
          @drop.prevent="handleCoverDrop"
          @click="triggerCoverSelect"
        >
          <img v-if="coverPreview" :src="coverPreview" class="cover-preview-image" alt="封面预览" />
          <div v-else class="cover-drop-placeholder">
            <el-icon class="cover-drop-icon"><Picture /></el-icon>
            <p class="cover-drop-text">拖拽图片到此处，或点击选择图片</p>
            <p class="cover-drop-hint">支持 JPG、PNG 格式</p>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="coverDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :disabled="!coverPreview"
            @click="handleSaveCover"
          >
            保存封面
          </el-button>
        </div>
      </template>
    </el-dialog>

    <input
      ref="coverFileInput"
      class="cover-file-input"
      type="file"
      accept="image/jpeg,image/png,image/jpg"
      @change="handleCoverFileSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Clock, View, EditPen, Edit, Delete, User, FolderOpened, Lock, Check, Upload, Loading, Notebook, Picture, Download, InfoFilled, Document, Close } from '@element-plus/icons-vue'
import { useBookStore } from '@/stores/book'
import { promptAPI, aiAPI, chapterAPI } from '@/api'
import type { Book, Prompt, Chapter } from '@/types'

const router = useRouter()
const bookStore = useBookStore()

const dialogVisible = ref(false)
const isEdit = ref(false)
const generating = ref(false)
const selectedPromptId = ref<number>()
const prompts = ref<Prompt[]>([])
const importFileInput = ref<HTMLInputElement | null>(null)
const pendingImportBook = ref<Book | null>(null)
const importingBookId = ref<number | null>(null)
const importDialogVisible = ref(false)
const selectedFiles = ref<File[]>([])
const isDragOver = ref(false)
const coverDialogVisible = ref(false)
const selectedBookForCover = ref<Book | null>(null)
const coverPreview = ref<string>('')
const coverFileInput = ref<HTMLInputElement | null>(null)
const isCoverDragOver = ref(false)

const formData = ref({
  id: 0,
  title: '',
  description: '',
  author: '',
  category: '',
  tags: '',
  status: 'draft' as 'draft' | 'published' | 'completed',
  is_public: false
})

type ParsedImportChapter = Pick<Chapter, 'title' | 'content'>

const CHAPTER_HEADING_REGEX = /^(?:\s| )*(第\s*[0-9零一二三四五六七八九十百千万两〇○O0-9①-⑳]+?\s*[章节卷回部篇集幕折][^\r\n]{0,30}|(?:序章|序言|前言|楔子|引子|终章|尾声|后记|番外(?:篇|合集)?|大结局)[^\r\n]{0,30}|chapter\s*[0-9ivxlcdm]+[^\r\n]{0,30})\s*$/gim

const IMPORT_BUTTON_TEXT = '导入整书'
const IMPORT_BUTTON_LOADING_TEXT = '导入中...'

onMounted(async () => {
  await bookStore.fetchBooks()
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
})

const handleCreate = () => {
  isEdit.value = false
  formData.value = {
    id: 0,
    title: '',
    description: '',
    author: '',
    category: '',
    tags: '',
    status: 'draft',
    is_public: false
  }
  dialogVisible.value = true
}

const handleEdit = (book: Book) => {
  isEdit.value = true
  formData.value = {
    id: book.id,
    title: book.title,
    description: book.description,
    author: book.author || '',
    category: book.category || '',
    tags: book.tags ? book.tags.join(',') : '',
    status: book.status || 'draft',
    is_public: book.is_public || false
  }
  dialogVisible.value = true
}

const handleWrite = (bookId: number) => {
  router.push(`/write/${bookId}`)
}

const triggerBookImport = (book: Book) => {
  pendingImportBook.value = book
  selectedFiles.value = []
  importDialogVisible.value = true
}

const triggerFileSelect = () => {
  if (importFileInput.value) {
    importFileInput.value.value = ''
    importFileInput.value.click()
  }
}

const formatFileSize = (size: number) => {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  return (size / (1024 * 1024)).toFixed(1) + ' MB'
}

const handleFileDrop = (event: DragEvent) => {
  isDragOver.value = false
  const files = event.dataTransfer?.files
  if (files) {
    const txtFiles = Array.from(files).filter(file => /\.txt$/i.test(file.name))
    if (txtFiles.length === 0) {
      ElMessage.warning('请选择 TXT 格式的文件')
      return
    }
    selectedFiles.value.push(...txtFiles)
    ElMessage.success(`已添加 ${txtFiles.length} 个文件`)
  }
}

const handleImportFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    const txtFiles = Array.from(files).filter(file => /\.txt$/i.test(file.name))
    if (txtFiles.length === 0) {
      ElMessage.warning('请选择 TXT 格式的文件')
      return
    }
    selectedFiles.value.push(...txtFiles)
    ElMessage.success(`已添加 ${txtFiles.length} 个文件`)
  }
}

const removeFile = (index: number) => {
  selectedFiles.value.splice(index, 1)
}

const clearSelectedFiles = () => {
  selectedFiles.value = []
}

const confirmImport = async () => {
  if (selectedFiles.value.length === 0 || !pendingImportBook.value) return

  try {
    await ElMessageBox.confirm(
      `确定要导入 ${selectedFiles.value.length} 个文件到《${pendingImportBook.value.title}》吗？`,
      '确认导入',
      {
        confirmButtonText: '开始导入',
        cancelButtonText: '取消',
        type: 'info'
      }
    )

    importDialogVisible.value = false
    await processImportFiles()
  } catch (error) {
    // 用户取消
  }
}

const processImportFiles = async () => {
  const book = pendingImportBook.value
  if (!book || selectedFiles.value.length === 0) return

  importingBookId.value = book.id
  let totalChapters = 0
  let successCount = 0

  try {
    for (const file of selectedFiles.value) {
      const text = await readTxtFile(file)
      if (!text) {
        ElMessage.warning(`文件 "${file.name}" 内容为空，已跳过`)
        continue
      }

      const baseTitle = file.name.replace(/\.txt$/i, '').trim() || book.title
      const parsedChapters = splitTxtIntoChapters(text, baseTitle)

      if (parsedChapters.length === 0) {
        ElMessage.warning(`文件 "${file.name}" 未识别到可导入的内容`)
        continue
      }

      const existingChaptersRes = await chapterAPI.getByBook(book.id)
      const existingChapterCount =
        existingChaptersRes.success && existingChaptersRes.data
          ? existingChaptersRes.data.filter(chapter => chapter.type === 'chapter').length
          : 0

      const batchImport = await tryBatchImportChapters(book.id, parsedChapters)
      if (batchImport.fallback) {
        await createChaptersSequentially(book.id, parsedChapters, existingChapterCount)
      }

      totalChapters += parsedChapters.length
      successCount++
    }

    await bookStore.fetchBooks()
    ElMessage.success(`成功导入 ${successCount} 个文件，共 ${totalChapters} 个章节`)
  } catch (error) {
    if (error instanceof Error && error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    importingBookId.value = null
    pendingImportBook.value = null
    selectedFiles.value = []
  }
}

const detectTextEncoding = (bytes: Uint8Array) => {
  if (bytes.length >= 3 && bytes[0] === 0xef && bytes[1] === 0xbb && bytes[2] === 0xbf) {
    return 'utf-8'
  }
  if (bytes.length >= 2 && bytes[0] === 0xff && bytes[1] === 0xfe) {
    return 'utf-16le'
  }
  if (bytes.length >= 2 && bytes[0] === 0xfe && bytes[1] === 0xff) {
    return 'utf-16be'
  }
  return ''
}

const readTxtFile = async (file: File) => {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  const preferredEncoding = detectTextEncoding(bytes)
  const encodings = [preferredEncoding, 'utf-8', 'gb18030', 'big5'].filter(Boolean)
  let bestText = ''
  let bestScore = Number.POSITIVE_INFINITY
  for (const encoding of encodings) {
    try {
      const text = new TextDecoder(encoding).decode(bytes)
      const invalidCharCount = (text.match(/\uFFFD/g) || []).length
      if (invalidCharCount < bestScore) {
        bestScore = invalidCharCount
        bestText = text
      }
      if (invalidCharCount === 0) {
        break
      }
    } catch (error) {
    }
  }
  return bestText
    .replace(/\r\n?/g, '\n')
    .replace(/\u0000/g, '')
    .replace(/^\uFEFF/, '')
    .trim()
}

const cleanImportTitle = (title: string, fallbackIndex: number) => {
  const normalizedTitle = title
    .replace(/^[\s ]+|[\s ]+$/g, '')
    .replace(/[：:]\s*$/, '')
    .replace(/\s+/g, ' ')
    .trim()
  return normalizedTitle || `第${fallbackIndex + 1}章`
}

const chunkPlainText = (text: string, baseTitle: string) => {
  const paragraphs = text
    .split(/\n{2,}/)
    .map(item => item.trim())
    .filter(Boolean)
  const chunks: ParsedImportChapter[] = []
  let currentChunk = ''
  for (const paragraph of paragraphs) {
    const nextChunk = currentChunk ? `${currentChunk}\n\n${paragraph}` : paragraph
    if (nextChunk.length > 6000 && currentChunk.length >= 1500) {
      chunks.push({ title: `第${chunks.length + 1}章`, content: currentChunk.trim() })
      currentChunk = paragraph
    } else {
      currentChunk = nextChunk
    }
  }
  if (currentChunk.trim()) {
    chunks.push({
      title: chunks.length === 0 ? `${baseTitle} 正文` : `第${chunks.length + 1}章`,
      content: currentChunk.trim()
    })
  }
  if (chunks.length > 1) {
    return chunks.map((chapter, index) => ({
      ...chapter,
      title: `第${index + 1}章`
    }))
  }
  return chunks
}

const splitTxtIntoChapters = (text: string, baseTitle: string) => {
  const normalizedText = text.replace(/\n{3,}/g, '\n\n').trim()
  if (!normalizedText) {
    return []
  }
  const matches = Array.from(normalizedText.matchAll(CHAPTER_HEADING_REGEX))
  if (matches.length === 0) {
    return chunkPlainText(normalizedText, baseTitle)
  }
  const chapters: ParsedImportChapter[] = []
  matches.forEach((match, index) => {
    const title = cleanImportTitle(match[0], index)
    const start = (match.index || 0) + match[0].length
    const end = index + 1 < matches.length ? (matches[index + 1].index || normalizedText.length) : normalizedText.length
    const content = normalizedText.slice(start, end).trim()
    if (content) {
      chapters.push({ title, content })
    }
  })
  if (chapters.length > 0) {
    return chapters
  }
  return chunkPlainText(normalizedText, baseTitle)
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

const handleImportFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  const book = pendingImportBook.value
  target.value = ''
  if (!file || !book) {
    pendingImportBook.value = null
    return
  }
  if (!/\.txt$/i.test(file.name)) {
    ElMessage.warning('目前仅支持导入 TXT 文本文件')
    pendingImportBook.value = null
    return
  }
  importingBookId.value = book.id
  try {
    const text = await readTxtFile(file)
    if (!text) {
      ElMessage.warning('TXT 文件内容为空，无法导入')
      return
    }
    const baseTitle = file.name.replace(/\.txt$/i, '').trim() || book.title
    const parsedChapters = splitTxtIntoChapters(text, baseTitle)
    if (parsedChapters.length === 0) {
      ElMessage.warning('没有识别到可导入的内容')
      return
    }
    const existingChaptersRes = await chapterAPI.getByBook(book.id)
    const existingChapterCount =
      existingChaptersRes.success && existingChaptersRes.data
        ? existingChaptersRes.data.filter(chapter => chapter.type === 'chapter').length
        : 0
    const previewTitles = parsedChapters
      .slice(0, 3)
      .map(chapter => chapter.title)
      .join('、')
    await ElMessageBox.confirm(
      `检测到 ${parsedChapters.length} 个章节，将追加到《${book.title}》当前内容后。` +
        `${existingChapterCount > 0 ? ` 现有 ${existingChapterCount} 章。` : ''}` +
        `${previewTitles ? ` 前几章：${previewTitles}` : ''}`,
      '导入整书',
      {
        type: 'info',
        confirmButtonText: '开始导入',
        cancelButtonText: '取消'
      }
    )
    const batchImport = await tryBatchImportChapters(book.id, parsedChapters)
    if (batchImport.fallback) {
      await createChaptersSequentially(book.id, parsedChapters, existingChapterCount)
    }
    await bookStore.fetchBooks()
    ElMessage.success(`成功导入 ${parsedChapters.length} 个章节`)
  } catch (error) {
    if (error instanceof Error && error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    importingBookId.value = null
    pendingImportBook.value = null
  }
}

const handleDelete = async (book: Book) => {
  try {
    await ElMessageBox.confirm(`确定删除书本"${book.title}"吗？`, '提示', { type: 'warning' })
    await bookStore.deleteBook(book.id)
    ElMessage.success('删除成功')
  } catch (error) {
  }
}

const handleGenerateDescription = async () => {
  if (!formData.value.title) {
    ElMessage.warning('请先输入书名')
    return
  }
  try {
    generating.value = true
    const res = await aiAPI.generateDescription({
      title: formData.value.title,
      promptId: selectedPromptId.value
    })
    if (res.success && res.data) {
      formData.value.description = res.data
      ElMessage.success('生成成功')
    }
  } catch (error) {
  } finally {
    generating.value = false
  }
}

const handleSubmit = async () => {
  if (!formData.value.title) {
    ElMessage.warning('请输入书名')
    return
  }
  try {
    const submitData = {
      ...formData.value,
      tags: formData.value.tags
        ? formData.value.tags
            .split(',')
            .map(tag => tag.trim())
            .filter(Boolean)
        : []
    }
    if (isEdit.value) {
      await bookStore.updateBook(formData.value.id, submitData)
      ElMessage.success('更新成功')
    } else {
      await bookStore.createBook(submitData)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
  } catch (error) {
  }
}

const getStatusText = (status?: string) => {
  const statusMap: Record<string, string> = {
    draft: '草稿',
    published: '已发布',
    completed: '已完结'
  }
  return statusMap[status || 'draft'] || '草稿'
}

const formatDate = (dateStr?: string) => {
  if (!dateStr) return '未知'
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  if (days === 0) return '今天'
  if (days === 1) return '昨天'
  if (days < 7) return `${days}天前`
  if (days < 30) return `${Math.floor(days / 7)}周前`
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

const formatDateShort = (dateStr?: string) => {
  if (!dateStr) return '--'
  const date = new Date(dateStr)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const formatTime = (dateStr?: string) => {
  if (!dateStr) return '--'
  const date = new Date(dateStr)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

const handleDownload = async (book: Book) => {
  try {
    const res = await chapterAPI.getByBook(book.id)
    if (!res.success || !res.data) {
      ElMessage.warning('暂无内容可下载')
      return
    }

    const chapters = res.data.filter(ch => ch.type === 'chapter')
    if (chapters.length === 0) {
      ElMessage.warning('暂无章节可下载')
      return
    }

    let content = `${book.title}\n\n`
    chapters.forEach(chapter => {
      content += `${chapter.title}\n\n${chapter.content}\n\n`
    })

    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${book.title}.txt`
    link.click()
    URL.revokeObjectURL(url)

    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const handleOpenCoverDialog = (book: Book) => {
  selectedBookForCover.value = book
  coverPreview.value = book.cover || ''
  coverDialogVisible.value = true
}

const triggerCoverSelect = () => {
  if (coverFileInput.value) {
    coverFileInput.value.value = ''
    coverFileInput.value.click()
  }
}

const handleCoverDrop = (event: DragEvent) => {
  isCoverDragOver.value = false
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    handleCoverFile(files[0])
  }
}

const handleCoverFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    handleCoverFile(target.files[0])
  }
}

const handleCoverFile = (file: File) => {
  if (!file.type.match(/^image\/(jpeg|png|jpg)$/)) {
    ElMessage.error('请上传 JPG 或 PNG 格式的图片')
    return
  }
  const reader = new FileReader()
  reader.onload = (e) => {
    coverPreview.value = e.target?.result as string
  }
  reader.readAsDataURL(file)
}

const handleSaveCover = async () => {
  if (!selectedBookForCover.value || !coverPreview.value) return

  try {
    await bookStore.updateBook(selectedBookForCover.value.id, {
      cover: coverPreview.value
    })
    ElMessage.success('封面设置成功')
    coverDialogVisible.value = false
  } catch (error) {
    ElMessage.error('封面设置失败')
  }
}
</script>

<style scoped>
.books-container {
  padding: 24px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h2 {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
  color: #1f2937;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.books-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.book-card {
  background: #fff;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  height: 220px;
  display: flex;
  flex-direction: column;
  padding: 16px;
  border: 1px solid #e8e8e8;
  overflow: hidden;
}

.book-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
}

.card-title-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
  flex-shrink: 0;
}

.type-badge {
  background: linear-gradient(135deg, #8b5cf6 0%, #ec4899 100%);
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 11px;
  color: #ffffff;
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.type-badge:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.4);
}

.type-badge.loading {
  opacity: 0.7;
  cursor: wait;
}

.book-title {
  flex: 1;
  font-size: 16px;
  font-weight: 700;
  color: #1f2937;
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.card-body-section {
  display: flex;
  gap: 12px;
  flex: 1;
  min-height: 0;
  margin-bottom: 12px;
  overflow: hidden;
}

.book-cover {
  position: relative;
  width: 80px;
  aspect-ratio: 4 / 3;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
}

.cover-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-icon {
  font-size: 28px;
  color: rgba(255, 255, 255, 0.8);
}

.cover-tag {
  position: absolute;
  bottom: 3px;
  right: 3px;
  background: rgba(0, 0, 0, 0.7);
  color: #ffffff;
  font-size: 9px;
  padding: 1px 4px;
  border-radius: 3px;
}

.book-description {
  flex: 1;
  display: flex;
  align-items: flex-start;
  min-width: 0;
  overflow: hidden;
}

.book-description p {
  margin: 0;
  font-size: 13px;
  color: #666;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-bottom-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid #f0f0f0;
  flex-shrink: 0;
}

.timestamp {
  display: flex;
  flex-direction: column;
  gap: 1px;
  flex-shrink: 0;
}

.date {
  font-size: 11px;
  color: #999;
  font-weight: 500;
}

.time {
  font-size: 10px;
  color: #bbb;
}

.action-buttons {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.action-btn {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 12px;
  flex-shrink: 0;
}

.action-btn:hover {
  transform: scale(1.1);
}

.action-btn.create {
  background: linear-gradient(135deg, #8b5cf6 0%, #a855f7 100%);
  color: #ffffff;
}

.action-btn.create:hover {
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.5);
}

.action-btn.cover {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: #ffffff;
}

.action-btn.cover:hover {
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.5);
}

.action-btn.download {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: #ffffff;
}

.action-btn.download:hover {
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.5);
}

.action-btn.confirm {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  color: #ffffff;
}

.action-btn.confirm:hover {
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.5);
}

.action-btn.info {
  background: linear-gradient(135deg, #06b6d4 0%, #0891b2 100%);
  color: #ffffff;
}

.action-btn.info:hover {
  box-shadow: 0 4px 12px rgba(6, 182, 212, 0.5);
}

.action-btn.delete {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: #ffffff;
}

.action-btn.delete:hover {
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.5);
}

.action-btn.loading,
.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.book-import-input {
  display: none;
}

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

.selected-files {
  margin-top: 20px;
}

.files-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.files-list {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.file-item:last-child {
  border-bottom: none;
}

.file-item .el-icon {
  color: #8b5cf6;
}

.file-name {
  flex: 1;
  font-size: 13px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: #999;
}

.file-remove {
  cursor: pointer;
  color: #999 !important;
  transition: color 0.2s;
}

.file-remove:hover {
  color: #ef4444 !important;
}

.dialog-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-section {
  background: #fafafa;
  border-radius: 8px;
  padding: 16px;
}

.section-header {
  margin-bottom: 16px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  color: #1f2937;
}

.form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 13px;
  font-weight: 500;
  color: #666;
}

.form-label.required::after {
  content: '*';
  color: #ff4d4f;
  margin-left: 4px;
}

.description-editor {
  position: relative;
}

.char-count {
  position: absolute;
  bottom: 10px;
  right: 12px;
  font-size: 11px;
  color: #999;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 1200px) {
  .books-grid {
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  }
}

@media (max-width: 768px) {
  .books-grid {
    grid-template-columns: 1fr;
  }

  .book-card {
    height: auto;
    min-height: 200px;
  }

  .card-body-section {
    flex-direction: column;
  }

  .book-cover {
    width: 100%;
    aspect-ratio: 4 / 3;
  }

  .action-buttons {
    flex-wrap: wrap;
    gap: 4px;
  }

  .action-btn {
    width: 26px;
    height: 26px;
    font-size: 11px;
  }
}

.cover-dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.cover-drop-zone {
  width: 100%;
  aspect-ratio: 4 / 3;
  max-width: 400px;
  border: 2px dashed #dcdfe6;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
  overflow: hidden;
  background: #fafafa;
}

.cover-drop-zone:hover {
  border-color: #3b82f6;
  background: #f0f7ff;
}

.cover-drop-zone.is-dragover {
  border-color: #3b82f6;
  background: #e0edff;
  transform: scale(1.02);
}

.cover-drop-zone.has-image {
  border-style: solid;
  border-color: #dcdfe6;
}

.cover-preview-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.cover-drop-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.cover-drop-icon {
  font-size: 48px;
  color: #c0c4cc;
}

.cover-drop-text {
  margin: 0;
  font-size: 14px;
  color: #606266;
}

.cover-drop-hint {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.cover-file-input {
  display: none;
}

/* 简约作品弹窗样式 */
.simple-book-dialog .el-dialog__body {
  padding: 20px 24px;
}

.simple-book-dialog .el-dialog__header {
  padding: 16px 24px;
  border-bottom: 1px solid #f0f0f0;
}

.simple-book-dialog .el-dialog__footer {
  padding: 12px 24px;
  border-top: 1px solid #f0f0f0;
}

.simple-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.simple-form .form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.simple-form .form-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.simple-form .form-label.required::after {
  content: '*';
  color: #f56c6c;
  margin-left: 4px;
}

.simple-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* ========== 暗色主题适配 ========== */
:root[data-theme='dark'] .book-card {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .book-card:hover {
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .book-title {
  color: #e5e7eb;
}

:root[data-theme='dark'] .book-description p {
  color: #9ca3af;
}

:root[data-theme='dark'] .card-bottom-section {
  border-top-color: #334155;
}

:root[data-theme='dark'] .date {
  color: #6b7280;
}

:root[data-theme='dark'] .time {
  color: #4b5563;
}

:root[data-theme='dark'] .cover-placeholder {
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
}

:root[data-theme='dark'] .cover-tag {
  background: rgba(0, 0, 0, 0.8);
}

:root[data-theme='dark'] .form-section {
  background: #1e293b;
}

:root[data-theme='dark'] .section-title {
  color: #e5e7eb;
}

:root[data-theme='dark'] .form-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .dialog-content {
  background: transparent;
}

:root[data-theme='dark'] .cover-drop-zone {
  border-color: #475569;
  background: #1e293b;
}

:root[data-theme='dark'] .cover-drop-zone:hover {
  border-color: #3b82f6;
  background: #1e3a5f;
}

:root[data-theme='dark'] .cover-drop-zone.is-dragover {
  border-color: #3b82f6;
  background: #1e3a5f;
}

:root[data-theme='dark'] .cover-drop-zone.has-image {
  border-color: #475569;
}

:root[data-theme='dark'] .cover-drop-placeholder {
  color: #6b7280;
}

:root[data-theme='dark'] .cover-drop-icon {
  color: #4b5563;
}

:root[data-theme='dark'] .cover-drop-text {
  color: #9ca3af;
}

:root[data-theme='dark'] .cover-drop-hint {
  color: #6b7280;
}

:root[data-theme='dark'] .simple-book-dialog .el-dialog__header {
  border-bottom-color: #334155;
}

:root[data-theme='dark'] .simple-book-dialog .el-dialog__footer {
  border-top-color: #334155;
}

:root[data-theme='dark'] .simple-form .form-label {
  color: #9ca3af;
}
</style>
