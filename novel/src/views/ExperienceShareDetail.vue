<template>
  <div ref="pageRef" class="experience-detail-page" v-loading="loading" @scroll="handlePageScroll">
    <div class="detail-background"></div>

    <header v-if="experienceShare" class="page-header">
      <div class="header-left">
        <el-button class="back-btn" @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回
        </el-button>
      </div>

      <div class="header-title">
        <h1>{{ experienceShare.title }}</h1>
        <p v-if="experienceShare.summary" class="header-summary">{{ experienceShare.summary }}</p>
      </div>

      <div class="header-actions">
        <el-radio-group
          v-if="experienceShare.create_type === 'manual'"
          v-model="activeRenderMode"
          size="small"
          class="render-mode-switch"
        >
          <el-radio-button label="markdown">MD</el-radio-button>
          <el-radio-button label="html">HTML</el-radio-button>
        </el-radio-group>
        <el-button type="primary" @click="goToEdit">
          <el-icon><Edit /></el-icon>
          编辑
        </el-button>
        <el-button
          v-if="experienceShare.pdf_file_url"
          @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')"
        >
          <el-icon><Document /></el-icon>
          查看 PDF
        </el-button>
      </div>
    </header>

    <div v-if="experienceShare" class="page-layout">
      <aside class="page-sidebar">
        <div class="sidebar-card">
          <label class="field-label">基本信息</label>
          <div class="info-list">
            <div class="info-item">
              <span class="info-label">作者</span>
              <span class="info-value">{{ experienceShare.author_name || '星芒用户' }}</span>
            </div>
            <div v-if="experienceShare.version" class="info-item">
              <span class="info-label">版本号</span>
              <span class="info-value">{{ experienceShare.version }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ formatDateTime(experienceShare.created_at) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">创建方式</span>
              <span class="info-value">
                <el-tag :type="experienceShare.create_type === 'pdf_import' ? 'warning' : 'success'" size="small">
                  {{ experienceShare.create_type === 'pdf_import' ? 'PDF 导入' : '手动创建' }}
                </el-tag>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">状态</span>
              <span class="info-value">{{ experienceShare.status || 'published' }}</span>
            </div>
            <div
              v-if="experienceShare.updated_at && experienceShare.updated_at !== experienceShare.created_at"
              class="info-item"
            >
              <span class="info-label">更新时间</span>
              <span class="info-value">{{ formatDateTime(experienceShare.updated_at) }}</span>
            </div>
          </div>
        </div>

        <div v-if="experienceShare.pdf_file_url" class="sidebar-card attachment-card-wrap">
          <label class="field-label">附件信息</label>
          <div class="attachment-card">
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
                size="small"
                @click="openPdfPreview(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'PDF附件')"
              >
                预览
              </el-button>
              <el-button
                size="small"
                @click="downloadPdf(experienceShare.pdf_file_url, experienceShare.pdf_file_name || 'experience-share.pdf')"
              >
                下载
              </el-button>
            </div>
          </div>
        </div>
      </aside>

      <main class="page-main">
        <div class="detail-shell">
          <div class="main-content">
            <section v-if="experienceShare.cover_url" class="form-section cover-section">
              <div class="cover-image-wrap">
                <img :src="experienceShare.cover_url" :alt="experienceShare.title" class="cover-image" />
              </div>
            </section>

            <section class="form-section content-section">
              <label class="field-label">{{ shouldShowPdfBody ? 'PDF 页面预览' : '正文内容' }}</label>

              <div v-if="shouldShowPdfBody" class="content-body">
                <PdfPageGallery v-if="experienceShare.pdf_file_url" :src="experienceShare.pdf_file_url" :scale="1.25" />
              </div>

              <MarkdownRenderer
                v-else-if="hasContent"
                class="content-body"
                :content="experienceShare.content"
                :render-mode="activeRenderMode"
              />

              <div v-else class="empty-content">当前卡片暂无正文内容。</div>
            </section>

            <section v-if="shouldShowPdfBody && hasContent" class="form-section content-section">
              <label class="field-label">补充说明</label>
              <MarkdownRenderer
                class="content-body"
                :content="experienceShare.content"
                :render-mode="activeRenderMode"
              />
            </section>
          </div>
        </div>
      </main>

      <aside class="page-rightbar">
        <div v-if="recommendedShares.length > 0" class="sidebar-card recommended-card">
          <label class="field-label">推荐阅读</label>
          <div class="recommended-list">
            <article
              v-for="item in recommendedShares"
              :key="item.id"
              class="recommended-item"
              @click="goToRecommended(item.id)"
            >
              <h4 class="recommended-title">{{ item.title }}</h4>
              <div class="recommended-cover">
                <img v-if="item.cover_url" :src="item.cover_url" :alt="item.title" />
                <div v-else class="cover-placeholder"></div>
              </div>
            </article>
          </div>
        </div>
      </aside>
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

type ContentRenderMode = 'markdown' | 'html'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const experienceShare = ref<ExperienceShare | null>(null)
const recommendedShares = ref<ExperienceShare[]>([])
const pdfPreviewVisible = ref(false)
const pdfPreviewUrl = ref('')
const pdfPreviewFileName = ref('')
const pdfPreviewDownloadName = ref('')
const pageRef = ref<HTMLElement | null>(null)
const showBackToTop = ref(false)
const activeRenderMode = ref<ContentRenderMode>('markdown')

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
    activeRenderMode.value = experienceShare.value?.content_render_mode === 'html' ? 'html' : 'markdown'

    try {
      const listRes = await experienceShareAPI.getAll()
      if (listRes.success && listRes.data) {
        recommendedShares.value = listRes.data.filter((item: ExperienceShare) => item.id !== id).slice(0, 6)
      }
    } catch {
      recommendedShares.value = []
    }
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

const goToRecommended = (id: number) => {
  router.push(`/experience-shares/${id}`)
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
  overflow: hidden;
  padding: 8px 12px 8px;
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

.page-layout {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 12px;
  max-width: 100%;
  margin: 0;
  padding: 0 8px;
  height: calc(100vh - 52px);
}

.page-header {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.85);
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  position: relative;
  z-index: 10;
  backdrop-filter: blur(8px);
}

.page-sidebar {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: calc(100vh - 52px);
  overflow-y: auto;
}

.page-main {
  flex: 1;
  min-width: 0;
  max-width: 980px;
  display: flex;
  justify-content: center;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.15) transparent;
}

.page-main::-webkit-scrollbar {
  width: 6px;
}

.page-main::-webkit-scrollbar-track {
  background: transparent;
}

.page-main::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 3px;
}

.page-main::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.25);
}

.page-rightbar {
  width: 260px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  max-height: calc(100vh - 52px);
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.15) transparent;
}

.page-rightbar::-webkit-scrollbar {
  width: 6px;
}

.page-rightbar::-webkit-scrollbar-track {
  background: transparent;
}

.page-rightbar::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 3px;
}

.page-rightbar::-webkit-scrollbar-thumb:hover {
  background: rgba(0, 0, 0, 0.25);
}

.recommended-card {
  display: flex;
  flex-direction: column;
}

.recommended-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.recommended-item {
  display: flex;
  flex-direction: column;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.6);
  border: none;
  overflow: hidden;
  cursor: pointer;
  transition: background-color 0.15s ease;
}

.recommended-item:hover {
  background-color: rgba(255, 255, 255, 0.85);
}

.recommended-title {
  font-size: 13px;
  font-weight: 500;
  color: #1a1a1a;
  line-height: 1.4;
  padding: 8px 10px 6px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0;
}

.recommended-cover {
  width: 100%;
  height: 140px;
  overflow: hidden;
  background: linear-gradient(135deg, #f5f5f5 0%, #ebebeb 100%);
  flex-shrink: 0;
}

.recommended-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sidebar-card {
  background: transparent;
  border-radius: 4px;
  padding: 12px;
  box-shadow: none;
  border: none;
}

.attachment-card-wrap {
  margin-top: 0;
}

.detail-shell {
  width: 100%;
  height: 100%;
  background: #ffffff;
  border-radius: 0;
  box-shadow: none;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.header-left {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.back-btn {
  flex-shrink: 0;
}

.header-title {
  text-align: center;
}

.header-title h1 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  line-height: 1.3;
}

.header-summary {
  margin: 2px 0 0;
  font-size: 12px;
  color: #888;
  line-height: 1.4;
}

.header-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.main-content {
  padding: 14px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.field-label {
  display: block;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #333;
  letter-spacing: 0;
  padding-bottom: 6px;
  border-bottom: 1px solid #e0e0e0;
}

.cover-section {
  margin-bottom: 0;
}

.cover-image-wrap {
  overflow: hidden;
  border-radius: 4px;
  background: #f5f5f5;
  box-shadow: none;
  max-width: 100%;
}

.cover-image {
  display: block;
  width: 100%;
  height: auto;
  object-fit: contain;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 0;
  margin-top: 10px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 6px 0;
  border-bottom: 1px solid #eee;
}

.info-item:last-child {
  border-bottom: none;
}

.info-label {
  font-size: 11px;
  font-weight: 500;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.info-value {
  font-size: 13px;
  color: #1a1a1a;
  font-weight: 400;
}

.attachment-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 10px;
  border-radius: 4px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
}

.attachment-main {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.attachment-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(220, 38, 38, 0.1);
  color: #dc2626;
  font-size: 16px;
  flex-shrink: 0;
}

.attachment-info {
  min-width: 0;
}

.attachment-name {
  color: #1a1a1a;
  font-weight: 500;
  word-break: break-all;
  font-size: 13px;
}

.attachment-desc {
  margin-top: 2px;
  color: #888;
  font-size: 11px;
}

.attachment-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.content-section .content-body {
  padding: 14px;
  border-radius: 4px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  min-height: 160px;
}

.empty-content {
  padding: 14px;
  border-radius: 4px;
  color: #999;
  background: #fafafa;
  border: 1px dashed #ddd;
  text-align: center;
  font-size: 13px;
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

:deep(.content-body .markdown-body) {
  color: #332d23;
  font-size: 15px;
  line-height: 1.8;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

:deep(.content-body .markdown-body h1),
:deep(.content-body .markdown-body h2),
:deep(.content-body .markdown-body h3),
:deep(.content-body .markdown-body h4) {
  color: #111827;
  font-weight: 700;
}

:deep(.content-body .markdown-body p) {
  margin-bottom: 14px;
}

:deep(.content-body .markdown-body blockquote) {
  margin: 0 0 16px;
  padding: 10px 16px;
  border-left: 3px solid #08c6be;
  background: #f9fafb;
  color: #4b5563;
}

:deep(.content-body .markdown-body pre) {
  background: #f3f4f6;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
}

:deep(.content-body .markdown-body code) {
  background: #f3f4f6;
  color: #374151;
}

:deep(.content-body .markdown-body table th) {
  background: #f9fafb;
}

@media (max-width: 1200px) {
  .page-rightbar {
    display: none;
  }
}

@media (max-width: 900px) {
  .page-header {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 12px;
    padding: 12px 16px;
  }

  .header-left {
    justify-content: flex-start;
  }

  .header-title h1 {
    white-space: normal;
    font-size: 18px;
  }

  .header-actions {
    justify-content: flex-end;
  }

  .page-layout {
    flex-direction: column;
    padding: 8px;
    height: auto;
    max-height: calc(100vh - 100px);
    overflow-y: auto;
  }

  .page-sidebar {
    width: 100%;
    position: static;
    max-height: none;
  }

  .page-main {
    width: 100%;
    max-width: 100%;
    overflow-y: visible;
  }

  .page-rightbar {
    width: 100%;
    max-height: none;
    overflow-y: visible;
  }

  .detail-shell {
    max-width: 100%;
  }

  .attachment-card {
    flex-direction: column;
    align-items: stretch;
  }

  .attachment-actions {
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .experience-detail-page {
    padding: 6px 8px 6px;
  }

  .detail-shell {
    border-radius: 0;
  }

  .page-header {
    padding: 10px 12px;
  }

  .header-title h1 {
    font-size: 16px;
  }

  .page-layout {
    padding: 0 8px;
    max-height: calc(100vh - 80px);
  }
}

/* 暗色主题适配 */
:root[data-theme='dark'] .experience-detail-page {
  background:
    radial-gradient(circle at top left, rgba(8, 198, 190, 0.08), transparent 28%),
    radial-gradient(circle at bottom right, rgba(234, 179, 8, 0.06), transparent 24%),
    linear-gradient(180deg, #1a1a1a 0%, #1e1e1e 44%, #1a1f1e 100%);
}

:root[data-theme='dark'] .detail-background {
  background:
    linear-gradient(120deg, rgba(255, 255, 255, 0.04), transparent 28%),
    radial-gradient(circle at 20% 20%, rgba(255, 255, 255, 0.06), transparent 18%),
    repeating-linear-gradient(
      90deg,
      rgba(255, 255, 255, 0.02) 0,
      rgba(255, 255, 255, 0.02) 1px,
      transparent 1px,
      transparent 24px
    );
}

:root[data-theme='dark'] .detail-shell {
  background: #1f2937;
}

:root[data-theme='dark'] .page-header {
  background: rgba(17, 24, 39, 0.85);
  border-bottom-color: rgba(255, 255, 255, 0.08);
}

:root[data-theme='dark'] .header-title h1 {
  color: #f9fafb;
}

:root[data-theme='dark'] .header-summary {
  color: #9ca3af;
}

:root[data-theme='dark'] .field-label {
  color: #e5e7eb;
  border-color: #374151;
}

:root[data-theme='dark'] .cover-image-wrap {
  background: #374151;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .info-item,
:root[data-theme='dark'] .attachment-card,
:root[data-theme='dark'] .content-body,
:root[data-theme='dark'] .sidebar-card {
  background: transparent;
  border-color: transparent;
}

:root[data-theme='dark'] .recommended-item {
  background: rgba(50, 55, 63, 0.6);
  border-color: transparent;
}

:root[data-theme='dark'] .recommended-item:hover {
  background-color: rgba(50, 55, 63, 0.85);
}

:root[data-theme='dark'] .recommended-title {
  color: #e8eaed;
}

:root[data-theme='dark'] .recommended-cover {
  background: linear-gradient(135deg, #2a2e35 0%, #252930 100%);
}

:root[data-theme='dark'] .info-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .info-value {
  color: #f3f4f6;
}

:root[data-theme='dark'] .attachment-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .attachment-desc {
  color: #9ca3af;
}

:root[data-theme='dark'] .empty-content {
  background: #374151;
  border-color: rgba(255, 255, 255, 0.15);
  color: #9ca3af;
}

:root[data-theme='dark'] .pdf-preview-toolbar {
  background: #2a2a2a;
  border-color: rgba(255, 255, 255, 0.1);
}

:root[data-theme='dark'] .pdf-preview-file {
  color: #f3f4f6;
}

:root[data-theme='dark'] .backtop-float {
  background: rgba(8, 198, 190, 0.85);
  color: #fff;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .backtop-float:hover {
  background: rgba(8, 198, 190, 0.95);
  box-shadow: 0 16px 38px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .backtop-icon {
  background: rgba(255, 255, 255, 0.2);
}

:root[data-theme='dark'] :deep(.content-card .markdown-body) {
  color: #e5e7eb;
}

:root[data-theme='dark'] :deep(.content-card .markdown-body h1),
:root[data-theme='dark'] :deep(.content-card .markdown-body h2),
:root[data-theme='dark'] :deep(.content-card .markdown-body h3),
:root[data-theme='dark'] :deep(.content-card .markdown-body h4) {
  color: #f3f4f6;
}

:root[data-theme='dark'] :deep(.content-card .markdown-body blockquote) {
  background: rgba(8, 198, 190, 0.1);
  border-left-color: rgba(8, 198, 190, 0.5);
  color: #9ca3af;
}

:root[data-theme='dark'] :deep(.content-card .markdown-body pre) {
  background: #2a2a2a;
  border-color: rgba(255, 255, 255, 0.1);
}

:root[data-theme='dark'] :deep(.content-card .markdown-body code) {
  background: rgba(8, 198, 190, 0.15);
  color: #5eead4;
}

:root[data-theme='dark'] :deep(.content-card .markdown-body table th) {
  background: #2a2a2a;
}

:root[data-theme='dark'] :deep(.content-card .markdown-body table td),
:root[data-theme='dark'] :deep(.content-card .markdown-body table th) {
  border-color: rgba(255, 255, 255, 0.1);
}
</style>
