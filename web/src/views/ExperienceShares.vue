<template>
  <div class="experience-page">
    <div class="page-header">
      <div>
        <h2>经验分享</h2>
        <p>沉淀案例方法、实操心得和可复用的经验卡片。</p>
      </div>
    </div>

    <div class="d6-body">
      <!-- 侧栏：经验库分类 + 管理操作 -->
      <aside class="d6-side">
        <div class="d6-side-sec">经验库</div>
        <button class="d6-side-item" :class="{ active: activeFilter === 'all' }" @click="setFilter('all')">
          <span>全部卡片</span><span class="n">{{ countAll }}</span>
        </button>
        <button class="d6-side-item" :class="{ active: activeFilter === 'manual' }" @click="setFilter('manual')">
          <span>手动创建</span><span class="n">{{ countManual }}</span>
        </button>
        <button class="d6-side-item" :class="{ active: activeFilter === 'pdf' }" @click="setFilter('pdf')">
          <span>PDF 文档</span><span class="n">{{ countPdf }}</span>
        </button>
        <button class="d6-side-item" :class="{ active: activeFilter === 'attach' }" @click="setFilter('attach')">
          <span>带 PDF 附件</span><span class="n">{{ countAttach }}</span>
        </button>
        <button class="d6-side-item" :class="{ active: activeFilter === 'trash' }" @click="setFilter('trash')">
          <span>回收站</span><span class="n">0</span>
        </button>

        <div class="d6-side-sec">管理</div>
        <button class="d6-side-item" @click="handleImport">
          <span>导入 JSON</span>
        </button>
        <button class="d6-side-item" @click="handleExport">
          <span>导出全部</span>
        </button>
        <button class="d6-side-item" :class="{ active: selectionMode }" @click="toggleSelectionMode">
          <span>{{ selectionMode ? '退出选择' : '批量选择' }}</span>
          <span v-if="selectionMode" class="n">{{ selectedIds.length }}</span>
        </button>
      </aside>

      <!-- 主区域 -->
      <div class="d6-main">
        <div class="d6-toolbar">
          <el-input
            v-model="searchQuery"
            placeholder="搜索标题、摘要或作者…"
            clearable
            class="d6-search"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-dropdown trigger="click" @command="handleCreateCommand">
            <el-button type="primary">
              <el-icon><Plus /></el-icon>
              新建
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

        <div ref="gridRef" v-loading="loading" class="d6-grid">
          <template v-if="experienceShares.length > 0">
            <template v-if="paginatedExperienceShares.length > 0">
              <article
                v-for="item in paginatedExperienceShares"
                :key="item.id"
                class="d6-card"
                :class="{ 'is-selected': selectedIds.includes(item.id) }"
                @click="handleCardClick(item)"
              >
                <span v-if="selectionMode" class="d6-check">
                  <el-icon v-if="selectedIds.includes(item.id)"><Check /></el-icon>
                </span>
                <div v-else class="d6-hover-actions" @click.stop>
                  <el-button link size="small" @click.stop="openEdit(item)">编辑</el-button>
                  <el-button link type="danger" size="small" @click.stop="handleDelete(item)">删除</el-button>
                </div>

                <div v-if="item.cover_url" class="d6-cover">
                  <img :src="item.cover_url" :alt="item.title" />
                </div>
                <div v-else class="d6-cover d6-cover-empty" :style="{ background: coverGradient(item) }">
                  <span v-if="!selectionMode" class="d6-cover-type">
                    {{ item.create_type === 'pdf_import' ? 'PDF 文档' : '实操心得' }}
                  </span>
                  <span class="d6-cover-title">{{ item.title || '未命名' }}</span>
                </div>

                <div class="d6-body">
                  <h4 class="d6-title">{{ item.title }}</h4>
                  <div class="d6-meta">
                    <span class="d6-av">{{ (item.author_name || '星')[0] }}</span>
                    <span class="d6-author">{{ item.author_name || '星芒用户' }}</span>
                    <span class="d6-tag" :class="item.create_type === 'pdf_import' ? 't-amber' : 't-teal'">
                      {{ item.create_type === 'pdf_import' ? 'PDF 导入' : '手动创建' }}
                    </span>
                    <span v-if="item.pdf_file_url" class="d6-tag t-gray">含 PDF</span>
                  </div>
                </div>
              </article>
            </template>
            <div v-else class="d6-filtered-empty">
              <el-icon class="fe-ic"><Collection /></el-icon>
              <p class="fe-title">{{ activeFilter === 'trash' ? '回收站为空' : '没有匹配的经验卡片' }}</p>
              <p class="fe-sub">
                {{ activeFilter === 'trash' ? '当前版本删除即为永久移除，暂无回收功能' : '换个关键词或分类试试' }}
              </p>
              <el-button size="small" @click="resetFilter">清除筛选</el-button>
            </div>
          </template>

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

        <footer class="d6-stats">
          <div class="stats-left">
            <template v-if="selectionMode">
              <el-checkbox
                :model-value="selectAll"
                :indeterminate="isIndeterminate"
                @change="handleSelectAllChange"
              >
                全选
              </el-checkbox>
              <span class="selection-count">已选 {{ selectedIds.length }} 项</span>
            </template>
            <span v-else class="stat-title">{{ statTitle }}</span>
          </div>
          <div class="stats-right">
            <template v-if="selectionMode">
              <el-button size="small" type="primary" :disabled="selectedIds.length === 0" @click="handleExportSelected">
                <el-icon><Download /></el-icon>
                导出选中
              </el-button>
              <el-button size="small" @click="cancelSelection">取消选择</el-button>
            </template>
            <template v-else>
              <el-pagination
                v-if="totalPages > 1"
                small
                layout="prev, pager, next"
                :total="totalItems"
                :page-size="pageSize"
                v-model:current-page="currentPage"
                @current-change="handlePageChange"
              />
              <span class="stats-hint">支持批量导出 · 数据可随时备份</span>
            </template>
          </div>
        </footer>
      </div>
    </div>

    <!-- 创建 / 编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="min(1100px, 92vw)"
      top="2vh"
      :close-on-click-modal="false"
      append-to-body
      @closed="resetFormState"
    >
      <div class="dialog-body">
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
      </div>

      <template #footer>
        <div class="form-footer">
          <div class="form-footer-left">
            <el-button @click="openPreviewDialog">预览效果</el-button>
            <el-button v-if="!currentPdfDisplay" @click="openPdfPicker(isPdfImportMode ? 'import' : 'attachment')">
              选择 PDF
            </el-button>
          </div>
          <div class="form-footer-right">
            <el-button @click="dialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="submitting" @click="handleSubmit">
              {{ isEdit ? '保存修改' : '发布' }}
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 效果预览弹窗 -->
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

    <!-- PDF 在线预览弹窗 -->
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
import TipTapEditor from '@/components/TipTapEditor.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'

type CreateMode = 'manual' | 'pdf_import'
type PdfPickerMode = 'attachment' | 'import'
type FilterKey = 'all' | 'manual' | 'pdf' | 'attach' | 'trash'

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
  content_render_mode: 'markdown' | 'html'
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

const FILTER_LABELS: Record<FilterKey, string> = {
  all: '全部卡片',
  manual: '手动创建',
  pdf: 'PDF 文档',
  attach: '带 PDF 附件',
  trash: '回收站'
}

const COVER_GRADIENTS = [
  'linear-gradient(135deg,#667eea 0%,#764ba2 100%)',
  'linear-gradient(135deg,#08c6be 0%,#0ea5e9 100%)',
  'linear-gradient(135deg,#f59e0b 0%,#ef4444 100%)',
  'linear-gradient(135deg,#10b981 0%,#059669 100%)',
  'linear-gradient(135deg,#8b5cf6 0%,#ec4899 100%)',
  'linear-gradient(135deg,#0ea5e9 0%,#6366f1 100%)',
  'linear-gradient(135deg,#14b8a6 0%,#38bdf8 100%)',
  'linear-gradient(135deg,#f97316 0%,#f43f5e 100%)'
]

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const importingPdf = ref(false)

const experienceShares = ref<ExperienceShare[]>([])
const dialogVisible = ref(false)
const pdfPreviewVisible = ref(false)
const previewDialogVisible = ref(false)
const previewData = ref({
  title: '',
  summary: '',
  content: '',
  cover_url: ''
})

/* ---- 工作台：筛选 / 搜索 / 分页 ---- */
const activeFilter = ref<FilterKey>('all')
const searchQuery = ref('')
const currentPage = ref(1)
const pageSize = 8
const gridRef = ref<HTMLElement | null>(null)

const filteredExperienceShares = computed(() => {
  let list = experienceShares.value
  switch (activeFilter.value) {
    case 'manual':
      list = list.filter(i => i.create_type === 'manual')
      break
    case 'pdf':
      list = list.filter(i => i.create_type === 'pdf_import')
      break
    case 'attach':
      list = list.filter(i => !!i.pdf_file_url)
      break
    case 'trash':
      list = []
      break
  }
  const q = searchQuery.value.trim()
  if (q) {
    list = list.filter(i => {
      const hay = `${i.title || ''} ${i.summary || ''} ${i.author_name || ''}`.toLowerCase()
      return hay.includes(q.toLowerCase())
    })
  }
  return list
})

const paginatedExperienceShares = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  return filteredExperienceShares.value.slice(start, start + pageSize)
})

const totalItems = computed(() => filteredExperienceShares.value.length)
const totalPages = computed(() => Math.ceil(totalItems.value / pageSize))

const countAll = computed(() => experienceShares.value.length)
const countManual = computed(() => experienceShares.value.filter(i => i.create_type === 'manual').length)
const countPdf = computed(() => experienceShares.value.filter(i => i.create_type === 'pdf_import').length)
const countAttach = computed(() => experienceShares.value.filter(i => !!i.pdf_file_url).length)

const statTitle = computed(() => {
  const base = FILTER_LABELS[activeFilter.value]
  const q = searchQuery.value.trim()
  return `${base} · ${totalItems.value} 项` + (q ? ` · 搜索「${q}」` : '')
})

const coverGradient = (item: ExperienceShare) => {
  const idx = (item.id || 0) % COVER_GRADIENTS.length
  return COVER_GRADIENTS[idx]
}

const setFilter = (key: FilterKey) => {
  activeFilter.value = key
}

const resetFilter = () => {
  activeFilter.value = 'all'
  searchQuery.value = ''
}

watch([activeFilter, searchQuery], () => {
  currentPage.value = 1
})

/* ---- 选择模式 ---- */
const selectionMode = ref(false)
const selectedIds = ref<number[]>([])

const toggleSelectionMode = () => {
  if (selectionMode.value) {
    cancelSelection()
  } else {
    enterSelectionMode()
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

const selectAll = computed(() => {
  const list = filteredExperienceShares.value
  return list.length > 0 && list.every(i => selectedIds.value.includes(i.id))
})

const isIndeterminate = computed(() => {
  const list = filteredExperienceShares.value
  const total = list.length
  const selected = list.filter(i => selectedIds.value.includes(i.id)).length
  return selected > 0 && selected < total
})

const handleSelectAllChange = (val: boolean) => {
  const list = filteredExperienceShares.value
  if (val) {
    list.forEach(i => {
      if (!selectedIds.value.includes(i.id)) selectedIds.value.push(i.id)
    })
  } else {
    const ids = new Set(list.map(i => i.id))
    selectedIds.value = selectedIds.value.filter(id => !ids.has(id))
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

/* ---- 编辑状态 / 表单 ---- */
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
  content_render_mode: 'html',
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

const resetFormState = () => {
  revokeLocalPdf()
  revokeLocalCoverImage()
  formData.value = createEmptyForm()
  isEdit.value = false
  pickerMode.value = 'attachment'
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
  gridRef.value?.scrollTo({ top: 0, behavior: 'smooth' })
}

const openCreateDialog = (mode: CreateMode) => {
  resetFormState()
  formData.value = createEmptyForm(mode)
  dialogVisible.value = true
}

const handleCreateCommand = (command: CreateMode) => {
  openCreateDialog(command)
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
    content_render_mode: item.content_render_mode === 'markdown' ? 'markdown' : 'html',
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

/* ---- 导入 / 导出 ---- */
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
    const safeTitle = item.title.replace(/[^\w一-龥]/g, '_').slice(0, 50) || '经验卡片'
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
  --teal: #08c6be;
  --teal-soft: rgba(8, 198, 190, 0.12);
  --amber-soft: rgba(245, 158, 11, 0.15);
  --line: #e5e7eb;
  padding: 8px;
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
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

/* 精简页头：只保留标题与描述，操作全部下沉到侧栏与底部栏 */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 24px;
  padding: 8px 8px 14px;
  margin-bottom: 8px;
  flex-shrink: 0;
}

.page-header h2 {
  margin: 0 0 4px;
  font-size: 24px;
  font-weight: 800;
  color: #0f172a;
  letter-spacing: -0.03em;
}

.page-header p {
  margin: 0;
  color: #64748b;
  font-size: 13px;
  letter-spacing: -0.01em;
}

/* ============ 工作台主体 ============ */
.d6-body {
  flex: 1;
  min-height: 0;
  display: flex;
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid #e7ebf1;
  box-shadow: 0 6px 24px -14px rgba(15, 23, 42, 0.16);
}

/* 侧栏 */
.d6-side {
  width: 200px;
  flex-shrink: 0;
  background: #fff;
  border-right: 1px solid var(--line);
  padding: 14px 10px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.d6-side-sec {
  font-size: 11px;
  color: #9aa5b5;
  padding: 14px 12px 4px;
  font-weight: 700;
  letter-spacing: 0.06em;
}

.d6-side-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 9px 12px;
  border-radius: 9px;
  font-size: 13px;
  cursor: pointer;
  color: #475569;
  transition: all 0.12s ease;
  border: none;
  background: none;
  width: 100%;
  text-align: left;
  font-family: inherit;
  font-weight: 500;
}

.d6-side-item:hover {
  background: #f1f5f9;
}

.d6-side-item.active {
  background: var(--teal-soft);
  color: #059669;
  font-weight: 700;
}

.d6-side-item .n {
  font-size: 11px;
  color: #9aa5b5;
  font-variant-numeric: tabular-nums;
}

.d6-side-item.active .n {
  color: #059669;
}

/* 主区域 */
.d6-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.d6-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--line);
  background: #fff;
  flex-shrink: 0;
}

.d6-search {
  width: 340px;
  max-width: 100%;
}

.d6-search :deep(.el-input__wrapper) {
  border-radius: 10px;
  background: #f1f5f9;
  box-shadow: none;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.d6-search :deep(.el-input__wrapper:hover) {
  background: #f8fafc;
}

.d6-search :deep(.el-input__wrapper.is-focus) {
  background: #fff;
  border-color: var(--teal);
  box-shadow: 0 0 0 3px rgba(8, 198, 190, 0.12);
}

.d6-search :deep(.el-input__inner) {
  font-size: 13px;
}

.d6-toolbar :deep(.el-button--primary) {
  font-weight: 600;
  letter-spacing: -0.01em;
  border-radius: 10px;
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border: none;
  box-shadow: 0 2px 8px rgba(8, 198, 190, 0.25);
}

.d6-toolbar :deep(.el-button--primary:hover) {
  box-shadow: 0 4px 16px rgba(8, 198, 190, 0.35);
  transform: translateY(-1px);
}

/* 卡片网格 */
.d6-grid {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(208px, 1fr));
  gap: 14px;
  padding: 16px;
  align-content: start;
  min-height: 0;
  position: relative;
  background: #f6f8fb;
}

.d6-grid :deep(.el-empty) {
  grid-column: 1 / -1;
  align-self: center;
}

.d6-card {
  position: relative;
  background: #ffffff;
  border: 1px solid var(--line);
  border-radius: 12px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.18s ease;
}

.d6-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 10px 20px -12px rgba(15, 23, 42, 0.2);
  transform: translateY(-2px);
}

.d6-card.is-selected {
  border-color: var(--teal);
  box-shadow: 0 0 0 3px rgba(8, 198, 190, 0.18);
}

/* 选择态勾选 */
.d6-check {
  position: absolute;
  top: 8px;
  left: 8px;
  width: 20px;
  height: 20px;
  border-radius: 6px;
  background: rgba(15, 23, 42, 0.35);
  border: 1.5px solid #fff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  z-index: 2;
  backdrop-filter: blur(2px);
  transition: all 0.15s ease;
}

.d6-card.is-selected .d6-check {
  background: var(--teal);
  border-color: var(--teal);
}

/* 悬停操作（渐进披露） */
.d6-hover-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  gap: 2px;
  padding: 2px;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  opacity: 0;
  transition: opacity 0.15s ease;
  z-index: 2;
}

.d6-card:hover .d6-hover-actions {
  opacity: 1;
}

.d6-hover-actions :deep(.el-button) {
  font-size: 12px;
  font-weight: 600;
  margin: 0 2px;
}

/* 封面 */
.d6-cover {
  height: 92px;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.d6-cover img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.d6-cover-empty {
  display: flex;
  align-items: flex-end;
  padding: 10px;
  color: #fff;
}

.d6-cover-type {
  position: absolute;
  top: 8px;
  left: 8px;
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.05em;
  background: rgba(0, 0, 0, 0.24);
  padding: 2px 7px;
  border-radius: 99px;
  backdrop-filter: blur(2px);
}

.d6-cover-title {
  position: relative;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.22);
}

/* 卡片正文 */
.d6-body {
  padding: 10px 12px 12px;
}

.d6-title {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.45;
  color: #111827;
  letter-spacing: -0.01em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.d6-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  flex-wrap: wrap;
}

.d6-av {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  flex-shrink: 0;
}

.d6-author {
  font-size: 11px;
  color: #9aa5b5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 72px;
}

.d6-tag {
  display: inline-flex;
  align-items: center;
  height: 18px;
  padding: 0 6px;
  border-radius: 5px;
  font-size: 10px;
  font-weight: 700;
  flex-shrink: 0;
}

.d6-tag.t-teal {
  background: var(--teal-soft);
  color: #059669;
}

.d6-tag.t-amber {
  background: var(--amber-soft);
  color: #d97706;
}

.d6-tag.t-gray {
  background: rgba(107, 114, 128, 0.12);
  color: #6b7280;
}

/* 过滤后为空 */
.d6-filtered-empty {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 60px 0;
  color: #9aa5b5;
  text-align: center;
}

.d6-filtered-empty .fe-ic {
  font-size: 34px;
  opacity: 0.55;
}

.d6-filtered-empty .fe-title {
  font-size: 14px;
  font-weight: 600;
  color: #64748b;
}

.d6-filtered-empty .fe-sub {
  font-size: 12px;
  margin-bottom: 6px;
}

/* 底部状态栏：普通态 / 选择态 */
.d6-stats {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 9px 16px;
  border-top: 1px solid var(--line);
  background: #fff;
  font-size: 12px;
  color: #8b95a7;
  flex-shrink: 0;
  min-height: 46px;
}

.stats-left,
.stats-right {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.stat-title {
  color: #64748b;
  font-weight: 600;
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.selection-count {
  color: #64748b;
  font-size: 13px;
  font-weight: 600;
}

.stats-hint {
  color: #9aa5b5;
  font-size: 11px;
  white-space: nowrap;
}

.d6-stats :deep(.el-pagination) {
  --el-pagination-button-bg-color: #ffffff;
  --el-pagination-hover-color: #08c6be;
  --el-pagination-button-color: #64748b;
}

.d6-stats :deep(.el-pager li.is-active) {
  background: #08c6be;
  color: #ffffff;
  border-radius: 6px;
}

.d6-stats :deep(.el-pagination .btn-prev),
.d6-stats :deep(.el-pagination .btn-next) {
  background: #fff;
}

.d6-stats :deep(.el-checkbox__label) {
  font-size: 13px;
  font-weight: 500;
  color: #475569;
}

/* 空状态插图 */
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

/* ============ 弹窗（沿用原实现） ============ */
.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.section-alert {
  margin-top: 4px;
}

.import-panel {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 18px 20px;
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

.experience-form :deep(.el-form-item__label) {
  font-size: 14px;
  font-weight: 400;
  color: #1f2225;
  text-align: right;
  justify-content: flex-end;
}

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

.form-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.form-footer-left {
  display: flex;
  gap: 10px;
}

.form-footer-right {
  display: flex;
  gap: 10px;
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

/* ============ 响应式 ============ */
@media (max-width: 900px) {
  .d6-body {
    flex-direction: column;
  }

  .d6-side {
    width: 100%;
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
    border-right: none;
    border-bottom: 1px solid var(--line);
    padding: 8px;
    overflow: visible;
  }

  .d6-side-sec {
    display: none;
  }

  .d6-side-item {
    width: auto;
    padding: 7px 10px;
    font-size: 12px;
  }

  .d6-toolbar {
    flex-wrap: wrap;
  }

  .d6-search {
    flex: 1;
    min-width: 160px;
  }

  .d6-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }

  .d6-stats {
    flex-wrap: wrap;
  }

  .stats-hint {
    display: none;
  }
}

/* ============ 暗色主题适配 ============ */
:root[data-theme='dark'] .experience-page::before {
  background:
    radial-gradient(ellipse 80% 50% at 20% 0%, rgba(8, 198, 190, 0.04), transparent),
    radial-gradient(ellipse 60% 40% at 80% 100%, rgba(99, 102, 241, 0.02), transparent),
    radial-gradient(ellipse 50% 30% at 50% 50%, rgba(13, 17, 23, 0.5), transparent);
}

:root[data-theme='dark'] .page-header h2 {
  color: #eef5f0;
}

:root[data-theme='dark'] .page-header p {
  color: #5f6b7a;
}

:root[data-theme='dark'] .d6-body {
  background: #10141c;
  border-color: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .d6-side {
  background: #151a24;
  border-color: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .d6-side-item {
  color: #94a3b8;
}

:root[data-theme='dark'] .d6-side-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .d6-side-item.active {
  background: rgba(8, 198, 190, 0.16);
  color: #7ff0ea;
}

:root[data-theme='dark'] .d6-side-item .n {
  color: #4d5768;
}

:root[data-theme='dark'] .d6-side-item.active .n {
  color: #7ff0ea;
}

:root[data-theme='dark'] .d6-side-sec {
  color: #4d5768;
}

:root[data-theme='dark'] .d6-toolbar,
:root[data-theme='dark'] .d6-stats {
  background: rgba(13, 17, 23, 0.4);
  border-color: rgba(28, 35, 51, 0.4);
}

:root[data-theme='dark'] .d6-grid {
  background: #0d1117;
}

:root[data-theme='dark'] .d6-search :deep(.el-input__wrapper) {
  background: #1c2333;
}

:root[data-theme='dark'] .d6-search :deep(.el-input__wrapper:hover) {
  background: #232b3f;
}

:root[data-theme='dark'] .d6-search :deep(.el-input__wrapper.is-focus) {
  background: #1c2333;
  border-color: rgba(8, 198, 190, 0.4);
}

:root[data-theme='dark'] .d6-card {
  background: rgba(22, 27, 34, 0.8);
  border-color: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .d6-card:hover {
  border-color: rgba(8, 198, 190, 0.25);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.35);
}

:root[data-theme='dark'] .d6-title {
  color: #eef5f0;
}

:root[data-theme='dark'] .d6-author {
  color: #5f6b7a;
}

:root[data-theme='dark'] .d6-tag.t-teal {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}

:root[data-theme='dark'] .d6-tag.t-amber {
  background: rgba(245, 158, 11, 0.2);
  color: #fbbf24;
}

:root[data-theme='dark'] .d6-tag.t-gray {
  background: rgba(148, 163, 184, 0.15);
  color: #5f6b7a;
}

:root[data-theme='dark'] .d6-hover-actions {
  background: rgba(22, 27, 34, 0.95);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .stat-title {
  color: #8b95a7;
}

:root[data-theme='dark'] .d6-stats {
  color: #4d5768;
}

:root[data-theme='dark'] .stats-hint {
  color: #4d5768;
}

:root[data-theme='dark'] .selection-count {
  color: #8b95a7;
}

:root[data-theme='dark'] .d6-stats :deep(.el-checkbox__label) {
  color: #94a3b8;
}

:root[data-theme='dark'] .d6-filtered-empty .fe-title {
  color: #8b95a7;
}

:root[data-theme='dark'] .d6-filtered-empty .fe-sub {
  color: #4d5768;
}
</style>
