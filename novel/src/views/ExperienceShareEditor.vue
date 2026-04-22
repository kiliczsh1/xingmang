<template>
  <div class="experience-editor-page" v-loading="loading">
    <div class="editor-background"></div>

    <div v-if="loaded" class="editor-shell">
      <header class="editor-topbar">
        <el-button class="ghost-btn" @click="goBackToDetail">
          <el-icon><ArrowLeft /></el-icon>
          返回详情
        </el-button>

        <div class="topbar-actions">
          <el-tag :type="isPdfImportMode ? 'warning' : 'success'">
            {{ isPdfImportMode ? 'PDF 导入创建' : '手动创建' }}
          </el-tag>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            保存修改
          </el-button>
        </div>
      </header>

      <section class="editor-card hero-card">
        <div class="editor-caption">编辑经验卡片</div>
        <h1>{{ formData.title || '未命名经验卡片' }}</h1>
        <p>
          这里可以直接修改标题、简介、封面图和 PDF 附件。手动创建的卡片支持富文本正文，PDF 导入卡片则直接展示 PDF 页面内容。
        </p>
      </section>

      <section class="editor-card form-card">
        <el-form label-width="96px" class="experience-form">
          <el-form-item label="标题" required>
            <el-input
              v-model="formData.title"
              maxlength="120"
              show-word-limit
              placeholder="请输入经验卡片标题"
            />
          </el-form-item>

          <el-form-item label="简介">
            <el-input
              v-model="formData.summary"
              type="textarea"
              :rows="4"
              maxlength="200"
              show-word-limit
              placeholder="建议一句话概括核心经验，列表页会优先展示这里的内容"
            />
          </el-form-item>

          <el-form-item v-if="!isPdfImportMode" label="封面图">
            <div class="cover-field">
              <el-input
                v-model="formData.cover_url"
                placeholder="请输入封面图 URL，详情页顶部会展示这张图片"
                clearable
              />
              <div v-if="formData.cover_url" class="cover-preview-card">
                <img :src="formData.cover_url" alt="封面预览" class="cover-preview-image" />
              </div>
            </div>
          </el-form-item>

          <el-form-item v-if="isPdfImportMode" label="PDF 导入">
            <div class="import-panel">
              <div class="import-copy">
                <div class="import-title">重新导入 PDF</div>
                <div class="import-desc">
                  重新选择一个 PDF 后，会自动覆盖标题和简介；正文仍然按页展示原始 PDF 内容。
                </div>
              </div>
              <el-button type="primary" :loading="importingPdf" @click="openPdfPicker('import')">
                重新导入 PDF
              </el-button>
            </div>
          </el-form-item>

          <el-form-item v-if="!isPdfImportMode" label="正文" required>
            <ExperienceContentEditor
              v-model="formData.content"
              v-model:render-mode="formData.content_render_mode"
              placeholder="请输入正文内容，支持 Markdown 语法"
              class="content-editor"
            />
          </el-form-item>

          <el-form-item v-else label="PDF 页面">
            <div class="pdf-content-preview">
              <div class="pdf-content-preview__info">
                <strong>PDF 页面图片展示</strong>
                <span>PDF 有多少页，这里就展示多少页图片，正文不再作为必填项。</span>
              </div>
              <div v-if="currentPdfDisplay" class="pdf-content-preview__frame">
                <PdfPageGallery :src="currentPdfDisplay.url" :scale="1.1" compact />
              </div>
              <div v-else class="pdf-content-preview__empty">
                请先上传 PDF 文件，系统会按页展示图片内容。
              </div>
            </div>
          </el-form-item>

          <el-form-item label="PDF 附件">
            <div class="pdf-section">
              <div v-if="currentPdfDisplay" class="pdf-file-card">
                <div class="pdf-file-main">
                  <el-icon class="pdf-file-icon"><Document /></el-icon>
                  <div class="pdf-file-meta">
                    <div class="pdf-file-name">{{ currentPdfDisplay.fileName }}</div>
                    <div class="pdf-file-desc">
                      <span>{{ formatFileSize(currentPdfDisplay.fileSize) }}</span>
                      <span>{{ currentPdfDisplay.isLocal ? '待保存到系统' : '已保存附件' }}</span>
                    </div>
                  </div>
                </div>

                <div class="pdf-file-actions">
                  <el-button type="primary" plain @click="previewCurrentPdf">预览</el-button>
                  <el-button @click="openPdfPicker(isPdfImportMode ? 'import' : 'attachment')">
                    替换
                  </el-button>
                  <el-button type="danger" plain @click="removeCurrentPdf">移除</el-button>
                </div>
              </div>

              <div v-else class="pdf-empty">
                <div class="pdf-empty-text">
                  <strong>支持上传 1 个 PDF 附件</strong>
                  <span>仅支持 `.pdf`，单文件不超过 20MB。</span>
                </div>
                <el-button @click="openPdfPicker(isPdfImportMode ? 'import' : 'attachment')">
                  选择 PDF
                </el-button>
              </div>

              <el-alert
                v-if="formData.pdf_parse_result"
                :title="formData.pdf_parse_result"
                :type="formData.pdf_parse_status === 'empty' ? 'warning' : 'success'"
                :closable="false"
                show-icon
              />
            </div>
          </el-form-item>
        </el-form>
      </section>
    </div>

    <el-empty v-else-if="!loading" description="未找到这张经验卡片">
      <el-button type="primary" @click="router.push('/experience-shares')">返回列表</el-button>
    </el-empty>

    <el-dialog
      v-model="pdfPreviewVisible"
      title="PDF 在线预览"
      width="min(1180px, 96vw)"
      top="2vh"
      destroy-on-close
    >
      <div class="pdf-preview-shell">
        <div class="pdf-preview-toolbar">
          <div class="pdf-preview-file">
            <el-icon><Document /></el-icon>
            <span>{{ pdfPreviewFileName }}</span>
          </div>
          <div class="pdf-preview-actions">
            <el-button v-if="pdfPreviewUrl" @click="openPdfInNewTab">新窗口打开</el-button>
            <el-button v-if="pdfPreviewUrl" type="primary" @click="downloadCurrentPdf">下载</el-button>
          </div>
        </div>
        <div class="pdf-preview-body">
          <PdfPageGallery v-if="pdfPreviewUrl" :src="pdfPreviewUrl" :scale="1.45" />
        </div>
      </div>
    </el-dialog>

    <input
      ref="pdfFileInputRef"
      type="file"
      accept=".pdf,application/pdf"
      style="display: none"
      @change="handlePdfFileChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { experienceShareAPI } from '@/api'
import type { ExperienceShare } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'
import ExperienceContentEditor from '@/components/ExperienceContentEditor.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

type PdfPickerMode = 'attachment' | 'import'
type ContentRenderMode = 'markdown' | 'html'

type PdfUploadPayload = {
  name: string
  size: number
  data_base64: string
}

type LocalPdfAttachment = PdfUploadPayload & {
  objectUrl: string
}

type FormState = {
  id: number
  title: string
  summary: string
  cover_url: string
  content: string
  content_render_mode: ContentRenderMode
  create_type: 'manual' | 'pdf_import'
  author_name: string
  status: string
  pdf_parse_status: string
  pdf_parse_result: string
  source_file_name: string
  existing_pdf_file_url: string
  existing_pdf_file_name: string
  existing_pdf_file_size: number
  remove_pdf: boolean
}

const MAX_PDF_SIZE = 20 * 1024 * 1024

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const loaded = ref(false)
const submitting = ref(false)
const importingPdf = ref(false)
const pdfPreviewVisible = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pdfFileInputRef = ref<HTMLInputElement | null>(null)
const pickerMode = ref<PdfPickerMode>('attachment')
const localPdfAttachment = ref<LocalPdfAttachment | null>(null)

const createEmptyForm = (): FormState => ({
  id: 0,
  title: '',
  summary: '',
  cover_url: '',
  content: '',
  content_render_mode: 'markdown',
  create_type: 'manual',
  author_name: '星芒用户',
  status: 'published',
  pdf_parse_status: '',
  pdf_parse_result: '',
  source_file_name: '',
  existing_pdf_file_url: '',
  existing_pdf_file_name: '',
  existing_pdf_file_size: 0,
  remove_pdf: false
})

const formData = ref<FormState>(createEmptyForm())

const stripRichText = (value = '') =>
  value
    .replace(/<[^>]+>/g, ' ')
    .replace(/&nbsp;/gi, ' ')
    .replace(/&amp;/gi, '&')
    .replace(/&lt;/gi, '<')
    .replace(/&gt;/gi, '>')
    .replace(/&quot;/gi, '"')
    .replace(/&#39;/gi, "'")
    .replace(/\s+/g, ' ')
    .trim()

const hasMeaningfulContent = (value = '') => stripRichText(value).length > 0
const isPdfImportMode = computed(() => formData.value.create_type === 'pdf_import')

const currentPdfDisplay = computed(() => {
  if (localPdfAttachment.value) {
    return {
      fileName: localPdfAttachment.value.name,
      fileSize: localPdfAttachment.value.size,
      url: localPdfAttachment.value.objectUrl,
      downloadName: localPdfAttachment.value.name,
      isLocal: true
    }
  }

  if (!formData.value.remove_pdf && formData.value.existing_pdf_file_url) {
    return {
      fileName: formData.value.existing_pdf_file_name || 'PDF附件',
      fileSize: formData.value.existing_pdf_file_size,
      url: formData.value.existing_pdf_file_url,
      downloadName: formData.value.existing_pdf_file_name || 'experience-share.pdf',
      isLocal: false
    }
  }

  return null
})

const revokeLocalPdf = () => {
  if (localPdfAttachment.value?.objectUrl) {
    URL.revokeObjectURL(localPdfAttachment.value.objectUrl)
  }
  localPdfAttachment.value = null
}

const fillForm = (item: ExperienceShare) => {
  revokeLocalPdf()
  formData.value = {
    id: item.id,
    title: item.title || '',
    summary: item.summary || '',
    cover_url: item.cover_url || '',
    content: item.content || '',
    content_render_mode: item.content_render_mode === 'html' ? 'html' : 'markdown',
    create_type: item.create_type || 'manual',
    author_name: item.author_name || '星芒用户',
    status: item.status || 'published',
    pdf_parse_status: item.pdf_parse_status || '',
    pdf_parse_result: item.pdf_parse_result || '',
    source_file_name: item.source_file_name || '',
    existing_pdf_file_url: item.pdf_file_url || '',
    existing_pdf_file_name: item.pdf_file_name || '',
    existing_pdf_file_size: Number(item.pdf_file_size || 0),
    remove_pdf: false
  }
}

const formatFileSize = (size?: number) => {
  const value = Number(size || 0)
  if (value <= 0) return '未知大小'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / 1024 / 1024).toFixed(2)} MB`
}

const loadDetail = async () => {
  const id = Number(route.params.id)
  if (!Number.isFinite(id) || id <= 0) {
    loaded.value = false
    formData.value = createEmptyForm()
    return
  }

  loading.value = true
  try {
    const res = await experienceShareAPI.getOne(id)
    if (res.success && res.data) {
      fillForm(res.data)
      loaded.value = true
    } else {
      loaded.value = false
    }
  } finally {
    loading.value = false
  }
}

const goBackToDetail = () => {
  if (!formData.value.id) {
    router.push('/experience-shares')
    return
  }
  router.push(`/experience-shares/${formData.value.id}`)
}

const validatePdfFile = (file: File) => {
  if (!/\.pdf$/i.test(file.name) && file.type !== 'application/pdf') {
    throw new Error('仅支持 PDF 格式文件')
  }

  if (file.size > MAX_PDF_SIZE) {
    throw new Error(`PDF 文件大小不能超过 ${Math.floor(MAX_PDF_SIZE / 1024 / 1024)}MB`)
  }
}

const fileToBase64 = (file: File) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      resolve(result.replace(/^data:application\/pdf;base64,/i, ''))
    }
    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsDataURL(file)
  })

const createLocalPdfAttachment = async (file: File) => {
  validatePdfFile(file)
  revokeLocalPdf()

  localPdfAttachment.value = {
    name: file.name,
    size: file.size,
    data_base64: await fileToBase64(file),
    objectUrl: URL.createObjectURL(file)
  }
  formData.value.remove_pdf = false
}

const openPdfPicker = (mode: PdfPickerMode) => {
  pickerMode.value = mode
  pdfFileInputRef.value?.click()
}

const applyImportedDraft = (draft: Partial<ExperienceShare>) => {
  formData.value.title = draft.title || formData.value.title
  formData.value.summary = draft.summary || ''
  formData.value.content = ''
  formData.value.content_render_mode = 'markdown'
  formData.value.create_type = 'pdf_import'
  formData.value.pdf_parse_status = draft.pdf_parse_status || ''
  formData.value.pdf_parse_result = draft.pdf_parse_result || ''
  formData.value.source_file_name = draft.source_file_name || localPdfAttachment.value?.name || ''
}

const importPdfDraft = async (file: File) => {
  importingPdf.value = true
  try {
    await createLocalPdfAttachment(file)
    const res = await experienceShareAPI.importPdf({
      pdf_file: {
        name: localPdfAttachment.value!.name,
        size: localPdfAttachment.value!.size,
        data_base64: localPdfAttachment.value!.data_base64
      }
    })

    if (res.success && res.data) {
      applyImportedDraft(res.data)
      ElMessage.success('PDF 导入成功，已自动生成标题和简介')
    }
  } finally {
    importingPdf.value = false
  }
}

const handlePdfFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  try {
    if (!file) return

    if (pickerMode.value === 'import') {
      await importPdfDraft(file)
    } else {
      await createLocalPdfAttachment(file)
      ElMessage.success('PDF 附件已更新')
    }
  } catch (error: any) {
    ElMessage.error(error.message || 'PDF 处理失败')
  } finally {
    target.value = ''
  }
}

const openPdfPreview = (url: string, fileName: string, downloadName = fileName) => {
  pdfPreviewUrl.value = url
  pdfPreviewFileName.value = fileName
  pdfPreviewDownloadName.value = downloadName
  pdfPreviewVisible.value = true
}

const previewCurrentPdf = () => {
  if (!currentPdfDisplay.value) return
  openPdfPreview(currentPdfDisplay.value.url, currentPdfDisplay.value.fileName, currentPdfDisplay.value.downloadName)
}

const openPdfInNewTab = () => {
  if (!pdfPreviewUrl.value) return
  window.open(pdfPreviewUrl.value, '_blank', 'noopener')
}

const downloadPdf = (url: string, fileName: string) => {
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  link.target = '_blank'
  link.rel = 'noopener'
  link.click()
}

const downloadCurrentPdf = () => {
  if (!pdfPreviewUrl.value) return
  downloadPdf(pdfPreviewUrl.value, pdfPreviewDownloadName.value || 'experience-share.pdf')
}

const removeCurrentPdf = () => {
  if (localPdfAttachment.value) {
    revokeLocalPdf()
  } else if (formData.value.existing_pdf_file_url) {
    formData.value.remove_pdf = true
  }

  if (isPdfImportMode.value) {
    formData.value.pdf_parse_status = ''
    formData.value.pdf_parse_result = ''
  }
}

const handleSubmit = async () => {
  if (!formData.value.title.trim()) {
    ElMessage.warning('标题不能为空')
    return
  }

  if (!isPdfImportMode.value && !hasMeaningfulContent(formData.value.content)) {
    ElMessage.warning('正文不能为空')
    return
  }

  if (isPdfImportMode.value && !currentPdfDisplay.value) {
    ElMessage.warning('PDF 导入创建需要保留导入的 PDF 附件')
    return
  }

  submitting.value = true
  try {
    const payload: any = {
      title: formData.value.title.trim(),
      summary: formData.value.summary.trim(),
      cover_url: formData.value.cover_url.trim(),
      content: isPdfImportMode.value ? '' : formData.value.content,
      content_render_mode: formData.value.content_render_mode,
      create_type: formData.value.create_type,
      author_name: formData.value.author_name,
      status: formData.value.status,
      pdf_parse_status: formData.value.pdf_parse_status || null,
      pdf_parse_result: formData.value.pdf_parse_result || null,
      source_file_name: formData.value.source_file_name || null,
      remove_pdf: formData.value.remove_pdf
    }

    if (localPdfAttachment.value) {
      payload.pdf_file = {
        name: localPdfAttachment.value.name,
        size: localPdfAttachment.value.size,
        data_base64: localPdfAttachment.value.data_base64
      }
    }

    const res = await experienceShareAPI.update(formData.value.id, payload)
    if (res.success && res.data) {
      ElMessage.success('更新成功')
      fillForm(res.data)
      router.replace(`/experience-shares/${res.data.id}`)
    }
  } finally {
    submitting.value = false
  }
}

watch(() => route.params.id, loadDetail)

onMounted(loadDetail)

onBeforeUnmount(() => {
  revokeLocalPdf()
})
</script>

<style scoped>
.experience-editor-page {
  position: relative;
  min-height: 100vh;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 18px 20px 42px;
  background:
    radial-gradient(circle at top left, rgba(8, 198, 190, 0.1), transparent 28%),
    radial-gradient(circle at bottom right, rgba(234, 179, 8, 0.08), transparent 24%),
    linear-gradient(180deg, #f6efe1 0%, #f4f0e8 44%, #e7efe9 100%);
}

.editor-background {
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(120deg, rgba(255, 255, 255, 0.34), transparent 28%),
    radial-gradient(circle at 20% 20%, rgba(255, 255, 255, 0.24), transparent 18%),
    repeating-linear-gradient(
      90deg,
      rgba(15, 23, 42, 0.012) 0,
      rgba(15, 23, 42, 0.012) 1px,
      transparent 1px,
      transparent 24px
    );
}

.editor-shell {
  position: relative;
  z-index: 1;
  max-width: 1180px;
  margin: 0 auto;
}

.editor-topbar,
.topbar-actions,
.pdf-file-main,
.pdf-file-actions,
.pdf-preview-file,
.pdf-preview-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.editor-topbar {
  justify-content: space-between;
  margin-bottom: 14px;
}

.ghost-btn {
  border-color: rgba(255, 248, 235, 0.3);
  background: rgba(248, 255, 252, 0.1);
  color: #224847;
}

.editor-card {
  border-radius: 26px;
  background: rgba(255, 251, 244, 0.96);
  box-shadow: 0 18px 46px rgba(84, 73, 50, 0.1);
  border: 1px solid rgba(120, 94, 52, 0.08);
}

.hero-card {
  padding: 20px 24px;
  margin-bottom: 16px;
}

.editor-caption {
  color: #7c8a86;
  font-size: 13px;
  letter-spacing: 0.08em;
}

.hero-card h1 {
  margin: 8px 0 10px;
  color: #173b39;
  font-size: clamp(28px, 3vw, 40px);
  line-height: 1.24;
  font-family: 'STSong', 'Songti SC', 'Noto Serif SC', Georgia, serif;
}

.hero-card p {
  margin: 0;
  color: #4b5b58;
  line-height: 1.75;
}

.form-card {
  padding: 24px;
}

.experience-form {
  margin-top: 4px;
}

.content-editor {
  width: 100%;
  max-height: 820px;
}

.cover-field {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cover-preview-card {
  overflow: hidden;
  max-width: 420px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: #ffffff;
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.06);
}

.cover-preview-image {
  display: block;
  width: 100%;
  height: 220px;
  object-fit: cover;
}

.import-panel {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  border-radius: 18px;
  background: linear-gradient(135deg, rgba(255, 247, 230, 0.92), rgba(255, 252, 244, 0.95));
  border: 1px solid rgba(250, 173, 20, 0.2);
}

.import-title {
  font-size: 16px;
  font-weight: 700;
  color: #7c4700;
}

.import-desc {
  margin-top: 6px;
  color: #8b5e13;
  font-size: 13px;
  line-height: 1.7;
}

.pdf-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.pdf-file-card,
.pdf-empty {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  width: 100%;
  padding: 16px 18px;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.pdf-file-icon {
  font-size: 26px;
  color: #dc2626;
  flex-shrink: 0;
}

.pdf-file-meta {
  min-width: 0;
}

.pdf-file-name {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  word-break: break-all;
}

.pdf-file-desc {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 6px;
  font-size: 12px;
  color: #64748b;
}

.pdf-empty-text {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #475569;
  font-size: 13px;
}

.pdf-content-preview {
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.pdf-content-preview__info {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 16px;
  color: #4b5b58;
  background: linear-gradient(180deg, rgba(255, 251, 244, 0.96), rgba(247, 249, 247, 0.98));
  border: 1px solid rgba(120, 94, 52, 0.08);
}

.pdf-content-preview__frame {
  overflow: hidden;
  max-height: min(68vh, 760px);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  border-radius: 18px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: #ffffff;
}

.pdf-content-preview__empty {
  padding: 18px;
  border-radius: 16px;
  color: #64748b;
  background: #f8fafc;
  border: 1px dashed rgba(148, 163, 184, 0.4);
}

.pdf-preview-shell {
  display: flex;
  flex-direction: column;
  gap: 14px;
  max-height: 82vh;
}

.pdf-preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 14px;
  background: #f8fafc;
  border: 1px solid rgba(148, 163, 184, 0.18);
}

.pdf-preview-file {
  color: #0f172a;
  font-weight: 600;
}

.pdf-preview-body {
  overflow-y: auto;
  max-height: calc(82vh - 74px);
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  padding-right: 2px;
}

@media (max-width: 900px) {
  .editor-topbar,
  .topbar-actions,
  .import-panel,
  .pdf-file-card,
  .pdf-empty,
  .pdf-preview-toolbar {
    flex-direction: column;
    align-items: stretch;
  }
}

@media (max-width: 768px) {
  .experience-editor-page {
    padding: 14px 14px 28px;
  }

  .editor-card {
    border-radius: 20px;
  }

  .hero-card,
  .form-card {
    padding: 18px;
  }

  .hero-card h1 {
    font-size: 24px;
  }
}
</style>
