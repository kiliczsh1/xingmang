<template>
  <div ref="pageRef" class="experience-detail-page" v-loading="loading" @scroll="handlePageScroll">
    <div class="detail-background"></div>

    <div v-if="experienceShare" class="detail-shell">
      <header class="detail-topbar">
        <el-button class="ghost-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回经验分享
        </el-button>

        <div class="topbar-actions">
          <el-button class="ghost-btn" @click="goToEdit">
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
          <el-button
            v-if="experienceShare.pdf_file_url"
            type="primary"
            @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')"
          >
            <el-icon><Document /></el-icon>
            查看 PDF
          </el-button>
        </div>
      </header>

      <section v-if="experienceShare.cover_url" class="cover-stage">
        <div class="cover-image-wrap">
          <img :src="experienceShare.cover_url" :alt="experienceShare.title" class="cover-image" />
        </div>
      </section>

      <section class="info-card">
        <div class="info-title-wrap">
          <h1>{{ experienceShare.title }}</h1>
          <p v-if="experienceShare.summary" class="info-summary">{{ experienceShare.summary }}</p>
        </div>

        <div class="meta-line">
          <span class="meta-pill">{{ experienceShare.author_name || '星芒用户' }}</span>
          <span class="meta-pill">{{ formatDateTime(experienceShare.created_at) }}</span>
          <span class="meta-pill">{{ experienceShare.create_type === 'pdf_import' ? 'PDF导入' : '手动创建' }}</span>
          <span class="meta-pill">{{ experienceShare.status || 'published' }}</span>
          <span
            v-if="experienceShare.updated_at && experienceShare.updated_at !== experienceShare.created_at"
            class="meta-pill"
          >
            更新：{{ formatDateTime(experienceShare.updated_at) }}
          </span>
          <el-tag effect="dark" :type="experienceShare.create_type === 'pdf_import' ? 'warning' : 'success'">
            {{ experienceShare.create_type === 'pdf_import' ? 'PDF 导入创建' : '手动创建' }}
          </el-tag>
          <el-tag v-if="experienceShare.pdf_file_url" effect="plain" type="info">附带 PDF</el-tag>
        </div>

        <div v-if="experienceShare.pdf_file_url" class="attachment-card">
          <div class="attachment-main">
            <div class="attachment-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="attachment-info">
              <div class="attachment-name">{{ experienceShare.pdf_file_name || 'PDF附件' }}</div>
              <div class="attachment-desc">{{ formatFileSize(experienceShare.pdf_file_size) }}</div>
            </div>
          </div>
          <div class="attachment-actions">
            <el-button
              type="primary"
              @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')"
            >
              在线预览
            </el-button>
            <el-button
              @click="downloadPdf(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'experience-share.pdf')"
            >
              下载附件
            </el-button>
          </div>
        </div>
      </section>

      <section class="content-card">
        <div class="section-title">{{ shouldShowPdfBody ? 'PDF 页面' : '正文内容' }}</div>

        <div v-if="shouldShowPdfBody" class="inline-pdf-viewer">
          <div class="inline-pdf-viewer__note">
            <strong>页面图片展示</strong>
            <span></span>
          </div>
          <PdfPageGallery v-if="experienceShare.pdf_file_url" :src="experienceShare.pdf_file_url" :scale="1.25" />
        </div>

        <MarkdownRenderer v-else-if="hasContent" :content="experienceShare.content" />

        <div v-else class="empty-content">
          当前卡片暂无正文内容。
        </div>
      </section>

      <section v-if="shouldShowPdfBody && hasContent" class="content-card secondary-content-card">
        <div class="section-title">补充说明</div>
        <MarkdownRenderer :content="experienceShare.content" />
      </section>
    </div>

    <el-empty v-else-if="!loading" description="未找到这张经验卡片">
      <el-button type="primary" @click="goBack">返回列表</el-button>
    </el-empty>

    <button
      v-show="showBackToTop"
      type="button"
      class="backtop-float"
      aria-label="回到顶部"
      @click="scrollToTop"
    >
      <span class="backtop-icon">↑</span>
      <span class="backtop-text">回到顶部</span>
    </button>

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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { experienceShareAPI } from '@/api'
import type { ExperienceShare } from '@/types'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import PdfPageGallery from '@/components/PdfPageGallery.vue'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const experienceShare = ref<ExperienceShare | null>(null)
const pdfPreviewVisible = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pageRef = ref<HTMLElement | null>(null)
const showBackToTop = ref(false)

const hasContent = computed(() => Boolean(experienceShare.value?.content?.trim()))
const shouldShowPdfBody = computed(() => {
  if (!experienceShare.value) return false
  return experienceShare.value.create_type === 'pdf_import' && Boolean(experienceShare.value.pdf_file_url)
})

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

const loadDetail = async () => {
  const id = Number(route.params.id)
  if (!Number.isFinite(id) || id <= 0) {
    experienceShare.value = null
    return
  }

  loading.value = true
  try {
    const res = await experienceShareAPI.getOne(id)
    experienceShare.value = res.success && res.data ? res.data : null
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/experience-shares')
}

const goToEdit = () => {
  if (!experienceShare.value) return
  router.push(`/experience-shares/${experienceShare.value.id}/edit`)
}

const openPdfPreview = (url: string, fileName: string) => {
  pdfPreviewUrl.value = url
  pdfPreviewFileName.value = fileName
  pdfPreviewDownloadName.value = fileName
  pdfPreviewVisible.value = true
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

const handlePageScroll = () => {
  showBackToTop.value = (pageRef.value?.scrollTop || 0) > 320
}

const scrollToTop = () => {
  pageRef.value?.scrollTo({
    top: 0,
    behavior: 'smooth'
  })
}

watch(() => route.params.id, loadDetail)

onMounted(loadDetail)
</script>

<style scoped>
.experience-detail-page {
  position: relative;
  height: 100vh;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 18px 20px 42px;
  background:
    radial-gradient(circle at top left, rgba(8, 198, 190, 0.1), transparent 28%),
    radial-gradient(circle at bottom right, rgba(234, 179, 8, 0.08), transparent 24%),
    linear-gradient(180deg, #f6efe1 0%, #f4f0e8 44%, #e7efe9 100%);
}

.detail-background {
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

.detail-shell {
  position: relative;
  z-index: 1;
  max-width: 1080px;
  margin: 0 auto;
}

.detail-topbar,
.topbar-actions,
.attachment-main,
.attachment-actions,
.pdf-preview-file,
.pdf-preview-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.detail-topbar {
  justify-content: space-between;
  margin-bottom: 12px;
}

.ghost-btn {
  border-color: rgba(255, 248, 235, 0.3);
  background: rgba(248, 255, 252, 0.1);
  color: #224847;
}

.cover-stage {
  margin-bottom: 14px;
}

.cover-image-wrap {
  overflow: hidden;
  border-radius: 28px;
  background: #dfe8e3;
  box-shadow: 0 18px 60px rgba(50, 43, 31, 0.16);
}

.cover-image {
  display: block;
  width: 100%;
  height: 360px;
  object-fit: cover;
}

.info-card,
.content-card {
  border-radius: 26px;
  background: rgba(255, 251, 244, 0.96);
  box-shadow: 0 18px 46px rgba(84, 73, 50, 0.1);
}

.info-card {
  padding: 18px 24px 16px;
  margin-bottom: 16px;
  border: 1px solid rgba(120, 94, 52, 0.08);
}

.info-title-wrap h1 {
  margin: 0;
  color: #173b39;
  font-size: clamp(28px, 3vw, 40px);
  line-height: 1.24;
  font-family: 'STSong', 'Songti SC', 'Noto Serif SC', Georgia, serif;
}

.info-summary {
  margin: 8px 0 0;
  color: #4b5b58;
  line-height: 1.75;
  font-size: 15px;
}

.meta-line {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-top: 12px;
  color: #5f6f6b;
  font-size: 13px;
}

.meta-pill {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 12px;
  border-radius: 999px;
  background: rgba(245, 248, 245, 0.96);
  border: 1px solid rgba(15, 23, 42, 0.06);
  color: #415451;
}

.attachment-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  margin-top: 14px;
  padding: 12px 14px;
  border-radius: 18px;
  background: linear-gradient(180deg, rgba(255, 248, 238, 0.98), rgba(255, 252, 246, 0.96));
  border: 1px solid rgba(15, 23, 42, 0.06);
}

.attachment-icon {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(220, 38, 38, 0.12);
  color: #dc2626;
  font-size: 20px;
  flex-shrink: 0;
}

.attachment-name {
  color: #1f2937;
  font-weight: 700;
  word-break: break-all;
}

.attachment-desc {
  margin-top: 4px;
  color: #64748b;
  font-size: 13px;
}

.content-card {
  position: relative;
  padding: 28px 32px 32px;
  border: 1px solid rgba(120, 94, 52, 0.08);
}

.secondary-content-card {
  margin-top: 16px;
}

.section-title {
  margin-bottom: 16px;
  color: #16413f;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.inline-pdf-viewer {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.inline-pdf-viewer__note {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 16px;
  color: #4b5b58;
  background: linear-gradient(180deg, rgba(255, 251, 244, 0.96), rgba(247, 249, 247, 0.98));
  border: 1px solid rgba(120, 94, 52, 0.08);
}

.empty-content {
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

.backtop-float {
  position: fixed;
  left: 14px;
  top: 50%;
  z-index: 30;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  height: 48px;
  padding: 0 16px 0 12px;
  border: none;
  border-radius: 999px;
  background: rgba(22, 65, 63, 0.9);
  color: #fffaf0;
  box-shadow: 0 12px 32px rgba(15, 23, 42, 0.18);
  cursor: pointer;
  transform: translate(-6px, -50%);
  transition: transform 0.22s ease, background-color 0.22s ease, box-shadow 0.22s ease;
}

.backtop-float:hover {
  background: #16413f;
  transform: translate(0, -50%);
  box-shadow: 0 16px 38px rgba(15, 23, 42, 0.24);
}

.backtop-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: rgba(255, 250, 240, 0.14);
  font-size: 16px;
  font-weight: 700;
  line-height: 1;
  flex-shrink: 0;
}

.backtop-text {
  max-width: 0;
  overflow: hidden;
  white-space: nowrap;
  opacity: 0;
  font-size: 13px;
  transition: max-width 0.22s ease, opacity 0.18s ease;
}

.backtop-float:hover .backtop-text {
  max-width: 80px;
  opacity: 1;
}

:deep(.el-tag) {
  border-radius: 999px;
  padding-inline: 10px;
}

:deep(.content-card .markdown-body) {
  color: #332d23;
  font-size: 16px;
  line-height: 1.95;
  font-family: 'STSong', 'Songti SC', 'Noto Serif SC', Georgia, serif;
}

:deep(.content-card .markdown-body h1),
:deep(.content-card .markdown-body h2),
:deep(.content-card .markdown-body h3),
:deep(.content-card .markdown-body h4) {
  color: #173b39;
  font-family: 'STSong', 'Songti SC', 'Noto Serif SC', Georgia, serif;
  font-weight: 700;
  letter-spacing: 0.03em;
}

:deep(.content-card .markdown-body p) {
  margin-bottom: 18px;
}

:deep(.content-card .markdown-body blockquote) {
  margin: 0 0 20px;
  padding: 12px 18px;
  border-left: 4px solid rgba(5, 150, 145, 0.42);
  background: rgba(20, 96, 90, 0.06);
  color: #3f4d4b;
}

:deep(.content-card .markdown-body pre) {
  background: #f3ede1;
  border: 1px solid rgba(120, 94, 52, 0.08);
  border-radius: 14px;
}

:deep(.content-card .markdown-body code) {
  background: rgba(22, 65, 63, 0.08);
  color: #16413f;
}

:deep(.content-card .markdown-body table th) {
  background: #f0eadc;
}

@media (max-width: 900px) {
  .detail-topbar,
  .topbar-actions,
  .attachment-card,
  .attachment-actions,
  .pdf-preview-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .cover-image {
    height: 280px;
  }

  .backtop-float {
    left: 10px;
    top: auto;
    bottom: 24px;
    transform: translateX(-4px);
  }

  .backtop-float:hover {
    transform: translateX(0);
  }
}

@media (max-width: 768px) {
  .experience-detail-page {
    padding: 14px 14px 28px;
  }

  .cover-image-wrap,
  .info-card,
  .content-card {
    border-radius: 20px;
  }

  .info-card,
  .content-card {
    padding: 18px;
  }

  .info-title-wrap h1 {
    font-size: 24px;
  }
}
</style>
