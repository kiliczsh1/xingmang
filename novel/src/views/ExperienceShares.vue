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
      <div v-if="experienceShares.length > 0">
        <div class="card-grid">
          <article
            v-for="item in paginatedExperienceShares"
            :key="item.id"
            class="experience-card"
            :class="{ 'is-selected': selectedIds.includes(item.id) }"
            @click="handleCardClick(item)"
          >
            <div v-if="item.cover_url" class="card-cover">
              <img :src="item.cover_url" :alt="item.title" class="card-cover-image" />
            </div>
            <div v-else class="card-cover card-cover-empty">
              <el-icon><Picture /></el-icon>
            </div>

            <div class="card-body">
              <div class="card-author">
                <div class="author-avatar">
                  {{ (item.author_name || '星')[0].toUpperCase() }}
                </div>
                <div class="author-info">
                  <span class="author-name">{{ item.author_name || '星芒用户' }}</span>
                  <span class="author-time">{{ formatDateTime(item.created_at) }}</span>
                </div>
              </div>

              <h3 class="card-title">{{ item.title }}</h3>

              <p class="card-summary" v-if="item.summary || item.content">
                {{ getSummaryPreview(item.summary || item.content) }}
              </p>
            </div>

            <div class="card-footer" @click.stop>
              <div class="footer-left">
                <el-tag size="small" :type="item.create_type === 'pdf_import' ? 'warning' : 'success'">
                  {{ item.create_type === 'pdf_import' ? 'PDF 导入' : '手动创建' }}
                </el-tag>
                <el-tag v-if="item.pdf_file_url" size="small" type="info">含 PDF</el-tag>
              </div>
              <div class="footer-right">
                <el-checkbox
                  v-if="selectionMode"
                  :model-value="selectedIds.includes(item.id)"
                  @change="(val: boolean) => handleSelectChange(item.id, val)"
                />
                <div class="footer-actions" v-else>
                  <el-button type="primary" link size="small" @click="openEdit(item)">编辑</el-button>
                  <el-button type="danger" link size="small" @click="handleDelete(item)">删除</el-button>
                </div>
              </div>
            </div>
          </article>
        </div>

        <div v-if="totalPages > 1" class="pagination-wrap">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="totalItems"
            layout="prev, pager, next"
            @current-change="handlePageChange"
          />
        </div>
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
      append-to-body
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
      append-to-body
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
      ref="coverImageInputRef"
      type="file"
      accept="image/jpeg,image/png,image/gif,image/webp"
      style="display: none"
      @change="handleCoverImageChange"
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
  remove_cover: boolean
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

const currentPage = ref(1)
const pageSize = 8

const paginatedExperienceShares = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return experienceShares.value.slice(start, end)
})

const totalItems = computed(() => experienceShares.value.length)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize))

const selectionMode = ref(false)
const selectedIds = ref<number[]>([])

const isEdit = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pdfFileInputRef = ref<HTMLInputElement | null>(null)
const coverImageInputRef = ref<HTMLInputElement | null>(null)
const pickerMode = ref<PdfPickerMode>('attachment')
const localPdfAttachment = ref<LocalPdfAttachment | null>(null)
const localCoverImage = ref<LocalCoverImage | null>(null)

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
  remove_pdf: false,
  remove_cover: false
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

const resetFormState = () => {
  revokeLocalPdf()
  revokeLocalCoverImage()
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
      currentPage.value = 1
    }
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  window.scrollTo({ top: 0, behavior: 'smooth' })
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
  revokeLocalCoverImage()
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
    remove_pdf: false,
    remove_cover: false
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
  revokeLocalCoverImage()
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
  position: relative;
}

.experience-page::before {
  content: '';
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: -1;
  background:
    radial-gradient(ellipse 80% 50% at 20% 0%, rgba(8, 198, 190, 0.05), transparent),
    radial-gradient(ellipse 60% 40% at 80% 100%, rgba(99, 102, 241, 0.03), transparent),
    radial-gradient(ellipse 50% 30% at 50% 50%, rgba(249, 250, 251, 1), transparent);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  margin-bottom: 28px;
}

.page-header h2 {
  margin: 0 0 6px;
  font-size: 28px;
  font-weight: 800;
  color: #0f172a;
  letter-spacing: -0.03em;
}

.page-header p {
  margin: 0;
  color: #64748b;
  font-size: 14px;
  letter-spacing: -0.01em;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.header-actions .el-button {
  font-weight: 500;
  letter-spacing: -0.01em;
  border-radius: 10px;
  transition: all 0.2s ease;
}

.header-actions .el-button--primary {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border: none;
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.25);
}

.header-actions .el-button--primary:hover {
  box-shadow: 0 4px 16px rgba(8, 198, 190, 0.35);
  transform: translateY(-1px);
}

.selection-count {
  color: #64748b;
  font-size: 13px;
  font-weight: 500;
}

.content-wrap {
  min-height: 320px;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  padding: 8px 0;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  margin-top: 32px;
  padding: 20px 0;
}

.pagination-wrap :deep(.el-pagination) {
  --el-pagination-button-bg-color: #ffffff;
  --el-pagination-hover-color: #08c6be;
}

.pagination-wrap :deep(.el-pager li.is-active) {
  background: #08c6be;
  color: #ffffff;
}

.experience-card {
  position: relative;
  display: flex;
  flex-direction: column;
  border-radius: 10px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.experience-card:hover {
  transform: translateY(-6px);
  box-shadow:
    0 12px 24px -8px rgba(0, 0, 0, 0.15),
    0 4px 8px -2px rgba(0, 0, 0, 0.1);
  border-color: #d1d5db;
}

.experience-card.is-selected {
  border-color: #08c6be;
  box-shadow: 0 0 0 3px rgba(8, 198, 190, 0.12), 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-cover {
  width: 100%;
  height: 160px;
  overflow: hidden;
  position: relative;
  background: linear-gradient(135deg, #f3f4f6 0%, #e5e7eb 100%);
  flex-shrink: 0;
}

.card-cover-empty {
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-cover-empty .el-icon {
  font-size: 48px;
  color: #9ca3af;
}

.card-cover-image {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.card-body {
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  flex: 1;
}

.card-author {
  display: flex;
  align-items: center;
  gap: 8px;
}

.author-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 700;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.25);
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.author-name {
  font-size: 14px;
  font-weight: 600;
  color: #374151;
  letter-spacing: -0.01em;
}

.author-time {
  font-size: 12px;
  color: #9ca3af;
}

.card-title {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
  line-height: 1.4;
  color: #111827;
  letter-spacing: -0.02em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-summary {
  margin: 0;
  color: #6b7280;
  line-height: 1.5;
  font-size: 12px;
  letter-spacing: -0.01em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-top: 1px solid #f3f4f6;
  background: #fafafa;
  min-height: 48px;
}

.footer-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  min-width: 0;
}

.footer-left .el-tag {
  height: 22px;
  padding: 0 8px;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.01em;
  border-radius: 6px;
  border: none;
  flex-shrink: 0;
}

.footer-left .el-tag--success {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.15) 0%, rgba(16, 185, 129, 0.08) 100%);
  color: #059669;
}

.footer-left .el-tag--warning {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.15) 0%, rgba(245, 158, 11, 0.08) 100%);
  color: #d97706;
}

.footer-left .el-tag--info {
  background: linear-gradient(135deg, rgba(107, 114, 128, 0.12) 0%, rgba(107, 114, 128, 0.06) 100%);
  color: #6b7280;
}

.footer-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.footer-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.25s ease;
}

.experience-card:hover .footer-actions {
  opacity: 1;
}

.footer-actions .el-button {
  font-size: 12px;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 6px;
  letter-spacing: -0.01em;
}

.footer-actions .el-button:hover {
  background: rgba(0, 0, 0, 0.04);
}

.footer-actions .el-button--danger:hover {
  background: rgba(239, 68, 68, 0.08);
  color: #dc2626;
}

.empty-illustration {
  width: 100px;
  height: 100px;
  border-radius: 24px;
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.12), rgba(99, 102, 241, 0.08));
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto;
}

.empty-illustration .el-icon {
  font-size: 40px;
  color: #08c6be;
  opacity: 0.7;
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

.cover-preview-actions {
  display: flex;
  justify-content: center;
  gap: 10px;
  padding: 12px;
  background: #f8fafc;
  border-top: 1px solid rgba(148, 163, 184, 0.12);
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
  border-radius: 18px;
  border: 2px dashed rgba(148, 163, 184, 0.35);
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
  cursor: pointer;
  transition: border-color 0.25s ease, background 0.25s ease;
}

.cover-upload-area:hover {
  border-color: #08c6be;
  background: linear-gradient(180deg, rgba(8, 198, 190, 0.04) 0%, rgba(244, 251, 250, 0.08) 100%);
}

.cover-upload-icon {
  font-size: 36px;
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
:root[data-theme='dark'] .experience-page::before {
  background:
    radial-gradient(ellipse 80% 50% at 20% 0%, rgba(8, 198, 190, 0.04), transparent),
    radial-gradient(ellipse 60% 40% at 80% 100%, rgba(99, 102, 241, 0.02), transparent),
    radial-gradient(ellipse 50% 30% at 50% 50%, rgba(15, 23, 42, 0.5), transparent);
}

:root[data-theme='dark'] .page-header h2 {
  color: #f1f5f9;
}

:root[data-theme='dark'] .page-header p {
  color: #94a3b8;
}

:root[data-theme='dark'] .experience-card {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(51, 65, 85, 0.6);
  box-shadow:
    0 1px 2px rgba(0, 0, 0, 0.2),
    0 4px 6px -1px rgba(0, 0, 0, 0.1);
}

:root[data-theme='dark'] .experience-card:hover {
  border-color: rgba(8, 198, 190, 0.25);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.3),
    0 8px 16px rgba(0, 0, 0, 0.2),
    0 0 0 1px rgba(8, 198, 190, 0.08);
}

:root[data-theme='dark'] .card-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .card-summary {
  color: #94a3b8;
}

:root[data-theme='dark'] .card-bottom {
  border-top-color: rgba(51, 65, 85, 0.4);
}

:root[data-theme='dark'] .card-bottom-meta {
  color: #64748b;
}

:root[data-theme='dark'] .card-bottom-actions .el-button:hover {
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .card-bottom-left .el-tag--info {
  background: rgba(148, 163, 184, 0.12);
  color: #94a3b8;
}
</style>
