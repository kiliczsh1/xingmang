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
          <el-button @click="openPreviewDialog">预览效果</el-button>
          <el-button v-if="!currentPdfDisplay" @click="openPdfPicker(isPdfImportMode ? 'import' : 'attachment')">
            选择 PDF
          </el-button>
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

          <el-form-item label="作者">
            <el-input
              v-model="formData.author_name"
              maxlength="50"
              placeholder="请输入作者名称"
            />
          </el-form-item>

          <el-form-item label="版本号">
            <el-input
              v-model="formData.version"
              maxlength="30"
              placeholder="请输入版本号，如 v1.0.0"
            />
          </el-form-item>

          <el-form-item v-if="!isPdfImportMode" label="封面图">
            <div class="cover-field">
              <div v-if="currentCoverDisplay" class="cover-preview-card">
                <img :src="currentCoverDisplay.url" alt="封面预览" class="cover-preview-image" />
                <div class="cover-preview-actions">
                  <el-button type="primary" plain size="small" @click="openCoverImagePicker">
                    替换图片
                  </el-button>
                  <el-button type="danger" plain size="small" @click="removeCoverImage">
                    移除
                  </el-button>
                </div>
              </div>
              <div v-else class="cover-upload-area" @click="openCoverImagePicker">
                <el-icon class="cover-upload-icon"><Plus /></el-icon>
                <span class="cover-upload-text">点击上传封面图</span>
                <span class="cover-upload-hint">支持 jpg、png、gif 格式，建议尺寸 16:9</span>
              </div>
              <div class="cover-url-input">
                <span class="cover-url-label">或输入图片URL：</span>
                <el-input
                  v-model="formData.cover_url"
                  placeholder="https://example.com/cover.jpg"
                  clearable
                  @input="handleCoverUrlInput"
                />
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
            <TipTapEditor
              v-model="formData.content"
              placeholder="请输入正文内容"
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

    <input
      ref="coverImageInputRef"
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      style="display: none"
      @change="handleCoverImageChange"
    />

    <el-dialog
      v-model="previewDialogVisible"
      title="效果预览"
      width="800px"
      top="5vh"
      destroy-on-close
      append-to-body
      class="preview-dialog"
    >
      <div class="preview-content">
        <div class="preview-card">
          <div v-if="previewData.cover_url" class="preview-cover">
            <img :src="previewData.cover_url" alt="封面" />
          </div>
          <div class="preview-body">
            <h2 class="preview-title">{{ previewData.title || '未填写标题' }}</h2>
            <p v-if="previewData.summary" class="preview-summary">{{ previewData.summary }}</p>
            <div class="preview-main">
              <div v-if="previewData.content" class="preview-html" v-html="previewData.content"></div>
              <p v-else class="preview-empty">暂无正文内容</p>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useRoute, useRouter } from 'vue-router'
import { experienceShareAPI } from '@/api'
import type { ExperienceShare } from '@/types'
import TipTapEditor from '@/components/TipTapEditor.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'

type PdfPickerMode = 'attachment' | 'import'

type PdfUploadPayload = {
  name: string
  size: number
  data_base64: string
}

type LocalPdfAttachment = PdfUploadPayload & {
  objectUrl: string
}

type CoverImagePayload = {
  name: string
  size: number
  data_base64: string
}

type LocalCoverImage = CoverImagePayload & {
  objectUrl: string
}

type FormState = {
  id: number
  title: string
  summary: string
  cover_url: string
  content: string
  create_type: 'manual' | 'pdf_import'
  author_name: string
  version: string
  status: string
  pdf_parse_status: string
  pdf_parse_result: string
  source_file_name: string
  existing_pdf_file_url: string
  existing_pdf_file_name: string
  existing_pdf_file_size: number
  remove_pdf: boolean
  remove_cover: boolean
}

const MAX_PDF_SIZE = 20 * 1024 * 1024

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const loaded = ref(false)
const submitting = ref(false)
const previewDialogVisible = ref(false)
const previewData = ref({
  title: '',
  summary: '',
  content: '',
  cover_url: ''
})
const importingPdf = ref(false)
const pdfPreviewVisible = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pdfFileInputRef = ref<HTMLInputElement | null>(null)
const coverImageInputRef = ref<HTMLInputElement | null>(null)
const pickerMode = ref<PdfPickerMode>('attachment')
const localPdfAttachment = ref<LocalPdfAttachment | null>(null)
const localCoverImage = ref<LocalCoverImage | null>(null)

const createEmptyForm = (): FormState => ({
  id: 0,
  title: '',
  summary: '',
  cover_url: '',
  content: '',
  create_type: 'manual',
  author_name: '星芒用户',
  version: '',
  status: 'published',
  pdf_parse_status: '',
  pdf_parse_result: '',
  source_file_name: '',
  existing_pdf_file_url: '',
  existing_pdf_file_name: '',
  existing_pdf_file_size: 0,
  remove_pdf: false,
  remove_cover: false
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

const currentCoverDisplay = computed(() => {
  if (localCoverImage.value) {
    return {
      url: localCoverImage.value.objectUrl,
      name: localCoverImage.value.name,
      size: localCoverImage.value.size,
      isLocal: true
    }
  }

  if (formData.value.cover_url) {
    return {
      url: formData.value.cover_url,
      name: '',
      size: 0,
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

const revokeLocalCoverImage = () => {
  if (localCoverImage.value?.objectUrl) {
    URL.revokeObjectURL(localCoverImage.value.objectUrl)
  }
  localCoverImage.value = null
}

const MAX_COVER_IMAGE_SIZE = 5 * 1024 * 1024

const openCoverImagePicker = () => {
  coverImageInputRef.value?.click()
}

const handleCoverImageChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  try {
    if (!file) return

    if (!file.type.startsWith('image/')) {
      throw new Error('请选择图片文件')
    }

    if (file.size > MAX_COVER_IMAGE_SIZE) {
      throw new Error('图片大小不能超过 5MB')
    }

    revokeLocalCoverImage()

    const data_base64 = await new Promise<string>((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => {
        const result = reader.result as string
        const base64 = result.split(',')[1]
        resolve(base64)
      }
      reader.onerror = () => reject(new Error('读取图片失败'))
      reader.readAsDataURL(file)
    })

    const objectUrl = URL.createObjectURL(file)

    localCoverImage.value = {
      name: file.name,
      size: file.size,
      data_base64,
      objectUrl
    }
    formData.value.cover_url = ''
    formData.value.remove_cover = false
  } catch (error: any) {
    ElMessage.error(error.message || '上传图片失败')
  } finally {
    target.value = ''
  }
}

const removeCoverImage = () => {
  revokeLocalCoverImage()
  if (formData.value.cover_url) {
    formData.value.remove_cover = true
  }
  formData.value.cover_url = ''
}

const handleCoverUrlInput = () => {
  if (formData.value.cover_url && localCoverImage.value) {
    revokeLocalCoverImage()
  }
}

const fillForm = (item: ExperienceShare) => {
  revokeLocalPdf()
  revokeLocalCoverImage()
  formData.value = {
    id: item.id,
    title: item.title || '',
    summary: item.summary || '',
    cover_url: item.cover_url || '',
    content: item.content || '',
    create_type: item.create_type || 'manual',
    author_name: item.author_name || '星芒用户',
    version: item.version || '',
    status: item.status || 'published',
    pdf_parse_status: item.pdf_parse_status || '',
    pdf_parse_result: item.pdf_parse_result || '',
    source_file_name: item.source_file_name || '',
    existing_pdf_file_url: item.pdf_file_url || '',
    existing_pdf_file_name: item.pdf_file_name || '',
    existing_pdf_file_size: Number(item.pdf_file_size || 0),
    remove_pdf: false,
    remove_cover: false
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

const openPreviewDialog = () => {
  const coverUrl = localCoverImage.value?.objectUrl || formData.value.cover_url || ''
  previewData.value = {
    title: formData.value.title,
    summary: formData.value.summary,
    content: formData.value.content,
    cover_url: coverUrl
  }
  previewDialogVisible.value = true
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
      create_type: formData.value.create_type,
      author_name: formData.value.author_name.trim(),
      version: formData.value.version.trim(),
      status: formData.value.status,
      pdf_parse_status: formData.value.pdf_parse_status || null,
      pdf_parse_result: formData.value.pdf_parse_result || null,
      source_file_name: formData.value.source_file_name || null,
      remove_pdf: formData.value.remove_pdf,
      remove_cover: formData.value.remove_cover
    }

    if (localPdfAttachment.value) {
      payload.pdf_file = {
        name: localPdfAttachment.value.name,
        size: localPdfAttachment.value.size,
        data_base64: localPdfAttachment.value.data_base64
      }
    }

    if (localCoverImage.value) {
      payload.cover_image = {
        name: localCoverImage.value.name,
        size: localCoverImage.value.size,
        data_base64: localCoverImage.value.data_base64
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
  revokeLocalCoverImage()
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

/* 表单标签样式 - 参考项目风格 */
.experience-form :deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 400;
  color: #1f2225;
  text-align: right;
  justify-content: flex-end;
}

/* 输入框样式 - 参考项目风格 */
.experience-form :deep(.el-input__wrapper) {
  border-radius: 3px;
  box-shadow: 0 0 0 1px rgb(224, 224, 230) inset;
  background-color: rgba(255, 255, 255, 1);
  transition: box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.experience-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #36ad6a inset;
}

.experience-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.2), 0 0 0 1px #18a058 inset;
}

.experience-form :deep(.el-textarea__inner) {
  border-radius: 3px;
  box-shadow: 0 0 0 1px rgb(224, 224, 230) inset;
  transition: box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

.experience-form :deep(.el-textarea__inner:hover) {
  box-shadow: 0 0 0 1px #36ad6a inset;
}

.experience-form :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px rgba(24, 160, 88, 0.2), 0 0 0 1px #18a058 inset;
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
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
  background: #ffffff;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.08);
}

.cover-preview-image {
  display: block;
  width: 100%;
  height: 220px;
  object-fit: cover;
}

.cover-preview-actions {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
  background: #fafafa;
}

.cover-upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  width: 100%;
  max-width: 420px;
  height: 180px;
  border-radius: 6px;
  border: 2px dashed rgb(224, 224, 230);
  background: #fafafa;
  cursor: pointer;
  transition: border-color 0.2s cubic-bezier(0.4, 0, 0.2, 1), background 0.2s ease;
}

.cover-upload-area:hover {
  border-color: #36ad6a;
  background: rgba(24, 160, 88, 0.04);
}

.cover-upload-icon {
  font-size: 32px;
  color: #94a3b8;
}

.cover-upload-text {
  font-size: 14px;
  font-weight: 500;
  color: #475569;
}

.cover-upload-hint {
  font-size: 12px;
  color: #94a3b8;
}

.cover-url-input {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed rgba(148, 163, 184, 0.3);
}

.cover-url-label {
  font-size: 13px;
  color: #64748b;
}

.import-panel {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 18px;
  border-radius: 6px;
  background: #f8fafc;
  border: 1px solid rgb(224, 224, 230);
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
  border-radius: 6px;
  color: #4b5b58;
  background: #f8fafc;
  border: 1px solid rgb(224, 224, 230);
}

.pdf-content-preview__frame {
  overflow: hidden;
  max-height: min(68vh, 760px);
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
  background: #ffffff;
}

.pdf-content-preview__empty {
  padding: 18px;
  border-radius: 6px;
  color: #64748b;
  background: #f8fafc;
  border: 1px dashed rgb(224, 224, 230);
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

.preview-content {
  padding: 0;
}

.preview-card {
  border-radius: 6px;
  border: 1px solid rgb(224, 224, 230);
  background: #ffffff;
  overflow: hidden;
}

.preview-cover {
  width: 100%;
  max-height: 300px;
  overflow: hidden;
}

.preview-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-body {
  padding: 20px;
}

.preview-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2225;
  margin: 0 0 12px 0;
  line-height: 1.4;
}

.preview-summary {
  font-size: 14px;
  color: #64748b;
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.preview-main {
  font-size: 14px;
  color: #334155;
  line-height: 1.8;
}

.preview-html,
.preview-markdown {
  min-height: 100px;
}

.preview-empty {
  color: #94a3b8;
  text-align: center;
  padding: 40px 0;
}

@media (max-width: 900px) {
  .editor-topbar,
  .topbar-actions,
  .import-panel,
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
