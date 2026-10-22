<template>
  <el-dialog
    v-model="importDialogVisible"
    title="导入作品"
    width="500px"
    :close-on-click-modal="false"
    append-to-body
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
        <p class="import-hint">支持 TXT / JSON，可同时选择多个文件</p>
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
    accept=".txt,text/plain,.json,application/json"
    multiple
    @change="handleImportFileSelect"
  />
</template>

<script setup lang="ts">
// 导入作品弹窗（拖拽 / 多选 / 编码探测 / 按章切分）
// 2026-09-21 第 3 期从 views/Books.vue 拆出；隐藏 input 一并搬入，开关由父组件传入
import { computed, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Document, Close } from '@element-plus/icons-vue'
import { useBookStore } from '@/stores/book'
import { chapterAPI, bookAPI } from '@/api'
import { readTxtFile, splitTxtIntoChapters, formatFileSize } from '@/utils/bookImport'
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

const selectedFiles = ref<File[]>([])

const isDragOver = ref(false)

const triggerFileSelect = () => {
  if (importFileInput.value) {
    importFileInput.value.value = ''
    importFileInput.value.click()
  }
}

const handleFileDrop = (event: DragEvent) => {
  isDragOver.value = false
  const files = event.dataTransfer?.files
  if (files) {
    const txtFiles = Array.from(files).filter(file => /\.(txt|json)$/i.test(file.name))
    if (txtFiles.length === 0) {
      ElMessage.warning('请选择 TXT 或 JSON 格式的文件')
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
    const txtFiles = Array.from(files).filter(file => /\.(txt|json)$/i.test(file.name))
    if (txtFiles.length === 0) {
      ElMessage.warning('请选择 TXT 或 JSON 格式的文件')
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
  if (selectedFiles.value.length === 0) return

  try {
    const fileNames = selectedFiles.value.map(f => f.name).join('、')
    await ElMessageBox.confirm(
      `确定要将 ${selectedFiles.value.length} 个文件（${fileNames}）作为新作品导入吗？`,
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
  let totalChapters = 0
  let successCount = 0
  importingBookId.value = -1 // 使用一个特殊值表示正在导入新书

  try {
    for (const file of selectedFiles.value) {
      // JSON：旧版本导出的作品文件，按章节落成真实文件
      if (/\.json$/i.test(file.name)) {
        try {
          const raw = await file.text()
          const res: any = await bookAPI.importBookJson(raw, file.name)
          if (res?.success && res.data) {
            successCount++
            totalChapters += res.data.chapters || 0
          } else {
            ElMessage.error(`导入 "${file.name}" 失败：${res?.message || '未知错误'}`)
          }
        } catch (e: any) {
          ElMessage.error(`导入 "${file.name}" 失败：${e?.response?.data?.message || e?.message || '未知错误'}`)
        }
        continue
      }

      const text = await readTxtFile(file)
      if (!text) {
        ElMessage.warning(`文件 "${file.name}" 内容为空，已跳过`)
        continue
      }
      
      const baseTitle = file.name.replace(/\.txt$/i, '').trim()
      const newBook = await bookStore.createBook({ title: baseTitle, description: '由文件导入' })
      
      if (!newBook) {
        ElMessage.error(`为文件 "${file.name}" 创建作品失败`)
        continue
      }

      const parsedChapters = splitTxtIntoChapters(text, baseTitle)
      if (parsedChapters.length > 0) {
        const batchImport = await tryBatchImportChapters(newBook.id, parsedChapters)
        if (batchImport.fallback) {
          await createChaptersSequentially(newBook.id, parsedChapters, 0)
        }
        totalChapters += parsedChapters.length
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
    selectedFiles.value = []
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
  if (v) selectedFiles.value = []
})
</script>

<style scoped>
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
