<template>
  <div class="experience-page">
    <div class="page-header">
      <div>
        <h2>经验分享</h2>
        <p>沉淀案例方法、实操心得和可复用的经验卡片。</p>
      </div>

      <div class="header-actions">
        <template v-if="selectionMode">
          <el-checkbox
            v-model="selectAll"
            :indeterminate="isIndeterminate"
            @change="handleSelectAllChange"
          >
            全选
          </el-checkbox>
          <span class="selection-count">已选 {{ selectedIds.length }} 项</span>
          <el-button size="large" type="primary" @click="handleExportSelected">
            <el-icon><Download /></el-icon>
            导出选中
          </el-button>
          <el-button size="large" @click="cancelSelection">
            取消选择
          </el-button>
        </template>
        <template v-else>
          <el-button size="large" @click="enterSelectionMode">
            <el-icon><Select /></el-icon>
            选择
          </el-button>
          <el-button size="large" @click="handleImport">
            <el-icon><Upload /></el-icon>
            导入
          </el-button>
          <el-button size="large" @click="handleExport">
            <el-icon><Download /></el-icon>
            导出全部
          </el-button>
        </template>
        <el-dropdown trigger="click" @command="handleCreateCommand">
          <el-button type="primary" size="large">
            <el-icon><Plus /></el-icon>
            创建
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="manual">
                <el-icon><EditPen /></el-icon>
                手动创建
              </el-dropdown-item>
              <el-dropdown-item command="pdf_import">
                <el-icon><UploadFilled /></el-icon>
                PDF 导入创建
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div v-loading="loading" class="content-wrap">
      <div v-if="experienceShares.length > 0" class="card-grid">
        <article
          v-for="item in experienceShares"
          :key="item.id"
          class="experience-card"
          :class="{ 'is-selected': selectedIds.includes(item.id) }"
          @click="handleCardClick(item)"
        >
          <div class="card-top">
            <div class="card-tags">
              <el-checkbox
                v-if="selectionMode"
                :model-value="selectedIds.includes(item.id)"
                @change="(val: boolean) => handleSelectChange(item.id, val)"
                @click.stop
              />
              <el-tag size="small" :type="item.create_type === 'pdf_import' ? 'warning' : 'success'">
                {{ item.create_type === 'pdf_import' ? 'PDF 导入' : '手动创建' }}
              </el-tag>
              <el-tag v-if="item.pdf_file_url" size="small" type="info">含 PDF 附件</el-tag>
            </div>

            <div class="card-actions" @click.stop>
              <el-button type="primary" link size="small" @click="openEdit(item)">
                编辑
              </el-button>
              <el-button type="info" link size="small" @click="handleExportSingle(item)">
                导出
              </el-button>
              <el-button type="danger" link size="small" @click="handleDelete(item)">
                删除
              </el-button>
            </div>
          </div>

          <h3 class="card-title">{{ item.title }}</h3>
          <p class="card-summary">
            {{ getSummaryPreview(item.summary || item.content) || '暂无摘要，点击查看详情。' }}
          </p>

          <div class="card-meta">
            <span>{{ item.author_name || '星芒用户' }}</span>
            <span>{{ formatDateTime(item.created_at) }}</span>
          </div>

          <div class="card-footer" @click.stop>
            <el-button
              v-if="item.pdf_file_url"
              type="primary"
              size="small"
              plain
              @click="openPdfPreview(item.pdf_file_url, item.pdf_file_name || 'PDF附件')"
            >
              查看 PDF
            </el-button>
            <el-button type="default" size="small" @click="openDetail(item)">
              查看详情
            </el-button>
          </div>
        </article>
      </div>

      <el-empty v-else description="还没有经验卡片">
        <template #image>
          <div class="empty-illustration">
            <el-icon><Collection /></el-icon>
          </div>
        </template>
        <el-button type="primary" @click="openCreateDialog('manual')">手动创建第一张卡片</el-button>
        <el-button style="margin-left: 12px" @click="openCreateDialog('pdf_import')">通过 PDF 导入</el-button>
      </el-empty>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="min(1100px, 92vw)"
      top="4vh"
      :close-on-click-modal="false"
      destroy-on-close
      @closed="resetFormState"
    >
      <div class="dialog-body">
        <div class="dialog-mode-row">
          <el-radio-group
            v-if="!isEdit"
            :model-value="formData.create_type"
            @update:model-value="switchCreateMode"
          >
            <el-radio-button label="manual">手动创建</el-radio-button>
            <el-radio-button label="pdf_import">PDF 导入创建</el-radio-button>
          </el-radio-group>
          <el-tag v-else :type="isPdfImportMode ? 'warning' : 'success'">
            {{ isPdfImportMode ? 'PDF 导入创建' : '手动创建' }}
          </el-tag>
        </div>

        <el-alert
          v-if="isPdfImportMode"
          title="导入后系统只会自动生成标题和简介，正文不再强行解析，而是直接保留 PDF 原始预览效果。"
          type="info"
          :closable="false"
          show-icon
          class="section-alert"
        />

        <div v-if="isPdfImportMode" class="import-panel">
          <div class="import-info">
            <div class="import-title">PDF 导入</div>
            <div class="import-desc">
              上传一个 PDF 后，系统会自动回填标题和简介，正文区域则直接展示原始 PDF 预览，避免版式被错误解析。
            </div>
          </div>
          <div class="import-actions">
            <el-button type="primary" :loading="importingPdf" @click="openPdfPicker('import')">
              {{ currentPdfDisplay ? '重新导入 PDF' : '选择 PDF 并生成初稿' }}
            </el-button>
            <span v-if="currentPdfDisplay" class="import-file-name">
              {{ currentPdfDisplay.fileName }}
            </span>
          </div>
        </div>

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
                      <span>{{ currentPdfDisplay.isLocal ? '待发布后保存到系统' : '已保存附件' }}</span>
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
                class="parse-alert"
              />
            </div>
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          {{ isEdit ? '保存修改' : '发布' }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="pdfPreviewVisible"
      title="PDF 在线预览"
      width="min(1100px, 95vw)"
      top="3vh"
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
            <el-button v-if="pdfPreviewUrl" type="primary" @click="downloadPdf">下载</el-button>
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
      ref="importFileInputRef"
      type="file"
      accept=".json,application/json"
      style="display: none"
      @change="handleImportFileChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { experienceShareAPI } from '@/api'
import type { ExperienceShare } from '@/types'
import SplitRichTextEditor from '@/components/SplitRichTextEditor.vue'
import ExperienceContentEditor from '@/components/ExperienceContentEditor.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

type CreateMode = 'manual' | 'pdf_import'
type ContentRenderMode = 'markdown' | 'html'
type PdfPickerMode = 'attachment' | 'import'

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
  create_type: CreateMode
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
const submitting = ref(false)
const importingPdf = ref(false)

const experienceShares = ref<ExperienceShare[]>([])
const dialogVisible = ref(false)
const pdfPreviewVisible = ref(false)

const selectionMode = ref(false)
const selectedIds = ref<number[]>([])

const isEdit = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pdfFileInputRef = ref<HTMLInputElement | null>(null)
const pickerMode = ref<PdfPickerMode>('attachment')
const localPdfAttachment = ref<LocalPdfAttachment | null>(null)

const createEmptyForm = (mode: CreateMode = 'manual'): FormState => ({
  id: 0,
  title: '',
  summary: '',
  cover_url: '',
  content: '',
  content_render_mode: 'markdown',
  create_type: mode,
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

const dialogTitle = computed(() => {
  if (isEdit.value) return '编辑经验卡片'
  return formData.value.create_type === 'pdf_import' ? 'PDF 导入创建' : '手动创建经验卡片'
})

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
      downloadName: formData.value.existing_pdf_file_name || '经验分享附件.pdf',
      isLocal: false
    }
  }

  return null
})

const selectAll = computed({
  get: () => {
    if (experienceShares.value.length === 0) return false
    return selectedIds.value.length === experienceShares.value.length
  },
  set: () => {}
})

const isIndeterminate = computed(() => {
  const total = experienceShares.value.length
  const selected = selectedIds.value.length
  return selected > 0 && selected < total
})

const revokeLocalPdf = () => {
  if (localPdfAttachment.value?.objectUrl) {
    URL.revokeObjectURL(localPdfAttachment.value.objectUrl)
  }
  localPdfAttachment.value = null
}

const resetFormState = () => {
  revokeLocalPdf()
  formData.value = createEmptyForm()
  isEdit.value = false
  pickerMode.value = 'attachment'
}

const formatDateTime = (value?: string | null) => {
  if (!value) return '未知时间'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatFileSize = (size?: number) => {
  const value = Number(size || 0)
  if (value <= 0) return '未知大小'
  if (value < 1024) return `${value} B`
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`
  return `${(value / 1024 / 1024).toFixed(2)} MB`
}

const getSummaryPreview = (value = '') => {
  const text = stripRichText(value)
  if (!text) return ''
  return text.length > 120 ? `${text.slice(0, 120).trim()}...` : text
}

const fetchExperienceShares = async () => {
  loading.value = true
  try {
    const res = await experienceShareAPI.getAll()
    if (res.success && res.data) {
      experienceShares.value = res.data
    }
  } finally {
    loading.value = false
  }
}

const openCreateDialog = (mode: CreateMode) => {
  resetFormState()
  formData.value = createEmptyForm(mode)
  dialogVisible.value = true
}

const handleCreateCommand = (command: CreateMode) => {
  openCreateDialog(command)
}

const switchCreateMode = (mode: string | number | boolean) => {
  const nextMode = mode === 'pdf_import' ? 'pdf_import' : 'manual'
  formData.value.create_type = nextMode
  formData.value.pdf_parse_status = ''
  formData.value.pdf_parse_result = ''
  formData.value.content = nextMode === 'pdf_import' ? '' : formData.value.content
  formData.value.content_render_mode = nextMode === 'pdf_import' ? 'markdown' : formData.value.content_render_mode
  if (nextMode === 'manual') {
    formData.value.source_file_name = ''
  }
}

const openDetail = (item: ExperienceShare) => {
  router.push(`/experience-shares/${item.id}`)
}

const fillEditForm = (item: ExperienceShare) => {
  revokeLocalPdf()
  isEdit.value = true
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
  dialogVisible.value = true
}

const openEdit = (item: ExperienceShare) => {
  fillEditForm(item)
}

const syncEditFromQuery = async () => {
  const editId = Number(route.query.edit)
  if (!Number.isFinite(editId) || editId <= 0) return

  const existing = experienceShares.value.find(item => item.id === editId)

  if (existing) {
    fillEditForm(existing)
  } else {
    const res = await experienceShareAPI.getOne(editId)
    if (res.success && res.data) {
      fillEditForm(res.data)
    }
  }

  router.replace({ path: '/experience-shares' })
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
      ElMessage.success('PDF 附件已添加')
    }
  } catch (error: any) {
    ElMessage.error(error.message || 'PDF 处理失败')
  } finally {
    target.value = ''
  }
}

const previewCurrentPdf = () => {
  if (!currentPdfDisplay.value) return
  openPdfPreview(currentPdfDisplay.value.url, currentPdfDisplay.value.fileName, currentPdfDisplay.value.downloadName)
}

const openPdfPreview = (url: string, fileName: string, downloadName = fileName) => {
  pdfPreviewUrl.value = url
  pdfPreviewFileName.value = fileName
  pdfPreviewDownloadName.value = downloadName
  pdfPreviewVisible.value = true
}

const openPdfInNewTab = () => {
  if (!pdfPreviewUrl.value) return
  window.open(pdfPreviewUrl.value, '_blank', 'noopener')
}

const downloadPdf = () => {
  if (!pdfPreviewUrl.value) return
  const link = document.createElement('a')
  link.href = pdfPreviewUrl.value
  link.download = pdfPreviewDownloadName.value || 'experience-share.pdf'
  link.target = '_blank'
  link.rel = 'noopener'
  link.click()
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

const handleDelete = async (item: ExperienceShare) => {
  try {
    await ElMessageBox.confirm(`确定删除经验卡片“${item.title}”吗？`, '删除经验卡片', {
      type: 'warning'
    })

    const res = await experienceShareAPI.delete(item.id)
    if (res.success) {
      ElMessage.success('删除成功')
      await fetchExperienceShares()
    }
  } catch {
    // 用户取消时不处理
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

    if (isEdit.value) {
      const res = await experienceShareAPI.update(formData.value.id, payload)
      if (res.success) {
        ElMessage.success('更新成功')
      }
    } else {
      const res = await experienceShareAPI.create(payload)
      if (res.success) {
        ElMessage.success('发布成功')
      }
    }

    dialogVisible.value = false
    await fetchExperienceShares()
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  await fetchExperienceShares()
  await syncEditFromQuery()
})

onBeforeUnmount(() => {
  revokeLocalPdf()
})

watch(
  () => route.query.edit,
  () => {
    syncEditFromQuery()
  }
)

const handleExport = async () => {
  if (experienceShares.value.length === 0) {
    ElMessage.warning('没有可导出的经验卡片')
    return
  }

  try {
    const exportData = experienceShares.value.map(item => ({
      title: item.title,
      summary: item.summary,
      content: item.content,
      content_render_mode: item.content_render_mode,
      author_name: item.author_name,
      create_type: item.create_type
    }))

    const jsonStr = JSON.stringify(exportData, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `经验卡片导出_${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success(`成功导出 ${exportData.length} 张经验卡片`)
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}

const handleExportSingle = (item: ExperienceShare) => {
  try {
    const exportData = [{
      title: item.title,
      summary: item.summary,
      content: item.content,
      content_render_mode: item.content_render_mode,
      author_name: item.author_name,
      create_type: item.create_type
    }]

    const jsonStr = JSON.stringify(exportData, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    const safeTitle = item.title.replace(/[^\w\u4e00-\u9fa5]/g, '_').slice(0, 50) || '经验卡片'
    link.download = `${safeTitle}_${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}

const handleExportSelected = () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请先选择要导出的经验卡片')
    return
  }

  try {
    const selectedItems = experienceShares.value.filter(item => selectedIds.value.includes(item.id))
    const exportData = selectedItems.map(item => ({
      title: item.title,
      summary: item.summary,
      content: item.content,
      content_render_mode: item.content_render_mode,
      author_name: item.author_name,
      create_type: item.create_type
    }))

    const jsonStr = JSON.stringify(exportData, null, 2)
    const blob = new Blob([jsonStr], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `经验卡片导出_${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success(`成功导出 ${exportData.length} 张经验卡片`)
    cancelSelection()
  } catch (error: any) {
    ElMessage.error(error.message || '导出失败')
  }
}

const enterSelectionMode = () => {
  selectionMode.value = true
  selectedIds.value = []
}

const cancelSelection = () => {
  selectionMode.value = false
  selectedIds.value = []
}

const handleSelectAllChange = (val: boolean) => {
  if (val) {
    selectedIds.value = experienceShares.value.map(item => item.id)
  } else {
    selectedIds.value = []
  }
}

const handleSelectChange = (id: number, val: boolean) => {
  if (val) {
    if (!selectedIds.value.includes(id)) {
      selectedIds.value.push(id)
    }
  } else {
    selectedIds.value = selectedIds.value.filter(i => i !== id)
  }
}

const handleCardClick = (item: ExperienceShare) => {
  if (selectionMode.value) {
    const isSelected = selectedIds.value.includes(item.id)
    handleSelectChange(item.id, !isSelected)
  } else {
    openDetail(item)
  }
}

const importFileInputRef = ref<HTMLInputElement | null>(null)

const handleImport = () => {
  importFileInputRef.value?.click()
}

const handleImportFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]

  try {
    if (!file) return

    if (!file.name.endsWith('.json')) {
      throw new Error('仅支持 JSON 格式文件')
    }

    const text = await file.text()
    const data = JSON.parse(text)

    if (!Array.isArray(data)) {
      throw new Error('文件格式错误，应为经验卡片数组')
    }

    const res = await experienceShareAPI.importAll({ cards: data })
    if (res.success) {
      ElMessage.success(`成功导入 ${data.length} 张经验卡片`)
      await fetchExperienceShares()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '导入失败')
  } finally {
    target.value = ''
  }
}

</script>

<style scoped>
.experience-page {
  padding: 8px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0 0 8px;
  font-size: 28px;
  color: #0f172a;
}

.page-header p {
  margin: 0;
  color: #64748b;
  font-size: 14px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.selection-count {
  color: #64748b;
  font-size: 14px;
}

.content-wrap {
  min-height: 320px;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 18px;
}

.experience-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 20px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(244, 251, 250, 0.98) 100%);
  background: var(--card-bg, linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(244, 251, 250, 0.98) 100%));
  border: 1px solid var(--border-primary, rgba(8, 198, 190, 0.12));
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.05);
  cursor: pointer;
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
}

.experience-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 18px 40px rgba(8, 198, 190, 0.12);
  border-color: rgba(8, 198, 190, 0.28);
}

.experience-card.is-selected {
  border-color: #08c6be;
  background: linear-gradient(180deg, rgba(8, 198, 190, 0.08) 0%, rgba(244, 251, 250, 0.98) 100%);
  box-shadow: 0 18px 40px rgba(8, 198, 190, 0.18);
}

.card-top,
.card-footer,
.card-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.card-tags,
.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.card-title {
  margin: 0;
  font-size: 18px;
  line-height: 1.45;
  color: #0f172a;
}

.card-summary {
  margin: 0;
  min-height: 66px;
  color: #475569;
  line-height: 1.7;
  font-size: 14px;
}

.card-meta {
  margin-top: auto;
  font-size: 12px;
  color: #64748b;
}

.card-footer {
  justify-content: flex-end;
}

.empty-illustration {
  width: 120px;
  height: 120px;
  border-radius: 28px;
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.18), rgba(5, 150, 145, 0.08));
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.empty-illustration .el-icon {
  font-size: 44px;
  color: #059691;
}

.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.dialog-mode-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.section-alert,
.parse-alert {
  margin-top: 4px;
}

.import-panel {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 18px 20px;
  border-radius: 16px;
  background: linear-gradient(135deg, rgba(255, 247, 230, 0.9), rgba(255, 252, 244, 0.95));
  border: 1px solid rgba(250, 173, 20, 0.22);
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

.import-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.import-file-name {
  max-width: 320px;
  color: #7c4700;
  font-size: 13px;
  word-break: break-all;
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
  border-radius: 16px;
  border: 1px solid rgba(148, 163, 184, 0.2);
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.pdf-file-main {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
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

.pdf-file-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
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

.pdf-preview-file,
.pdf-preview-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
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
  .page-header,
  .import-panel,
  .pdf-file-card,
  .pdf-empty,
  .pdf-preview-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    justify-content: flex-start;
  }

  .card-grid {
    grid-template-columns: 1fr;
  }

  .card-footer {
    justify-content: flex-start;
  }
}

/* 暗色主题适配 */
:root[data-theme='dark'] .card-title {
  color: #f3f4f6;
}
</style>
