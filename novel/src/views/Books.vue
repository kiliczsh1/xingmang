<template>
  <div class="books-container">
    <div class="header">
      <h2>作品管理</h2>
      <div class="header-actions">
        <el-button @click="triggerBookImport()">
          <el-icon><Upload /></el-icon>
          导入作品
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          创建作品
        </el-button>
      </div>
    </div>

    <div class="books-grid">
      <div v-for="book in bookStore.books" :key="book.id" class="book-card" @click="handleEdit(book)">
        <div class="card-title-section">
          <h3 class="book-title">{{ book.title }}</h3>
        </div>

        <div class="card-body-section">
          <div class="book-cover" @click.stop="handlePreviewBookCover(book)">
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
              <button class="action-btn info" @click.stop="handleShowInfo(book)">
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
      append-to-body
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
      width="600px"
      :close-on-click-modal="false"
      append-to-body
      class="cover-dialog"
    >
      <template #header>
        <div class="cover-dialog-header">
          <span class="cover-dialog-title">设置封面</span>
          <el-button
            v-if="coverActiveTab === 'ai-generate' && generationForm.selectedPromptId"
            type="primary"
            link
            size="small"
            @click="showFullPrompt"
          >
            <el-icon><View /></el-icon>
            查看完整提示词
          </el-button>
        </div>
      </template>
      <el-tabs v-model="coverActiveTab" class="cover-tabs">
        <el-tab-pane label="上传封面" name="upload">
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
              <el-button
                v-if="coverPreview"
                class="cover-preview-zoom-btn"
                type="primary"
                link
                @click.stop="showCoverPreview"
              >
                <el-icon><ZoomIn /></el-icon>
              </el-button>
              <div v-else class="cover-drop-placeholder">
                <el-icon class="cover-drop-icon"><Picture /></el-icon>
                <p class="cover-drop-text">拖拽图片到此处，或点击选择图片</p>
                <p class="cover-drop-hint">支持 JPG、PNG 格式</p>
              </div>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="AI 生成封面" name="ai-generate">
          <div class="ai-cover-form">
            <el-form :model="generationForm" label-position="top">
              <el-form-item label="小说名称">
                <el-input v-model="generationForm.name" disabled />
              </el-form-item>
              <el-form-item label="生成模型">
                <el-select
                  v-model="generationForm.selectedModelId"
                  placeholder="选择图片生成模型"
                  :loading="modelsLoading"
                >
                  <el-option
                    v-for="m in coverModels"
                    :key="m.id"
                    :label="`${m.provider_name ? m.provider_name + ' - ' : ''}${m.name}`"
                    :value="m.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="提示词">
                <el-select
                  v-model="generationForm.selectedPromptId"
                  placeholder="选择封面生成提示词"
                  clearable
                  :loading="promptsLoading"
                  @change="handlePromptSelect"
                >
                  <el-option
                    v-for="p in coverPrompts"
                    :key="p.id"
                    :label="p.name"
                    :value="p.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="参考图">
                <div class="reference-image-panel">
                  <div class="reference-image-actions">
                    <el-button @click="triggerReferenceImageSelect">
                      <el-icon><Upload /></el-icon>
                      上传参考图
                    </el-button>
                    <el-button @click="openRankReferenceDialog">
                      <el-icon><Picture /></el-icon>
                      排行榜选择
                    </el-button>
                  </div>
                  <span class="reference-image-hint">支持 JPG、PNG、WEBP、GIF，单张最大 20MB</span>
                  <div v-if="referenceImage" class="reference-image-preview">
                    <img :src="referenceImage.previewUrl" alt="参考图预览" class="reference-image-thumb" />
                    <div class="reference-image-meta">
                      <span class="reference-image-name">{{ referenceImage.name }}</span>
                      <span class="reference-image-size">{{ formatFileSize(referenceImage.size) }}</span>
                    </div>
                    <el-button type="danger" link @click="clearReferenceImage">移除</el-button>
                  </div>
                </div>
              </el-form-item>
              <template v-if="coverPromptFields.length > 0">
                <el-form-item
                  v-for="field in coverPromptFields"
                  :key="field.name"
                  :required="field.required !== false"
                >
                  <template #label>
                    <span class="field-label-text">{{ field.label }}</span>
                    <el-tag
                      :type="field.required !== false ? 'danger' : 'info'"
                      size="small"
                      class="field-required-tag"
                    >
                      {{ field.required !== false ? '必选' : '选填' }}
                    </el-tag>
                  </template>
                  <el-input
                    v-if="field.type === 'text'"
                    v-model="coverFieldValues[field.name]"
                    :placeholder="`请输入${field.label}`"
                  />
                  <el-input
                    v-else-if="field.type === 'textarea'"
                    v-model="coverFieldValues[field.name]"
                    type="textarea"
                    :rows="3"
                    :placeholder="`请输入${field.label}`"
                  />
                  <el-select
                    v-else-if="field.type === 'select'"
                    v-model="coverFieldValues[field.name]"
                    :placeholder="`请选择${field.label}`"
                    class="cover-field-select"
                    popper-class="cover-field-select-popper"
                    :teleported="false"
                  >
                    <el-option
                      v-for="(opt, optionIndex) in field.options"
                      :key="opt"
                      :label="field.optionLabels?.[optionIndex] || opt"
                      :value="opt"
                    />
                  </el-select>
                </el-form-item>
              </template>
            </el-form>
             <div class="generation-actions">
              <el-button
                type="primary"
                :loading="isGeneratingCover"
                :disabled="!generationForm.selectedModelId"
                @click="handleGenerateCover"
                class="generate-btn"
              >
                <el-icon><MagicStick /></el-icon>
                {{ isGeneratingCover ? '正在生成...' : '开始生成' }}
              </el-button>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <div class="dialog-footer">
          <el-button
            v-if="coverPreview"
            @click="downloadCurrentCover"
          >
            下载封面
          </el-button>
          <el-button
            v-if="coverPreview"
            @click="showCoverPreview"
          >
            <el-icon><ZoomIn /></el-icon>
            放大预览
          </el-button>
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

    <input
      ref="referenceImageInput"
      class="cover-file-input"
      type="file"
      accept="image/jpeg,image/png,image/jpg,image/webp,image/gif"
      @change="handleReferenceImageSelect"
    />

    <el-dialog
      v-model="rankReferenceDialogVisible"
      title="排行榜参考图"
      width="880px"
      append-to-body
      class="rank-reference-dialog"
    >
      <div class="rank-reference-layout">
        <div class="rank-reference-groups">
          <button
            v-for="group in rankReferenceGroups"
            :key="group.key"
            type="button"
            class="rank-reference-group-btn"
            :class="{ active: activeRankGroupKey === group.key }"
            @click="selectRankGroup(group.key)"
          >
            {{ group.label }}
          </button>
        </div>
        <div class="rank-reference-content">
          <div v-if="activeRankGroup" class="rank-reference-categories">
            <button
              v-for="category in activeRankGroup.categories"
              :key="category.id"
              type="button"
              class="rank-reference-category-btn"
              :class="{ active: activeRankCategoryPath === category.path }"
              @click="selectRankCategory(category.path)"
            >
              {{ category.label }}
            </button>
          </div>
          <div v-if="rankReferenceLoading" class="rank-reference-state">正在加载排行榜封面...</div>
          <div v-else-if="rankReferenceBooks.length === 0" class="rank-reference-state">暂无可用封面</div>
          <div v-else class="rank-reference-grid">
            <button
              v-for="book in rankReferenceBooks"
              :key="`${activeRankCategoryPath}_${book.rank}_${book.coverUrl}`"
              type="button"
              class="rank-reference-card"
              @click="useRankReferenceBook(book)"
            >
              <img :src="book.coverUrl" :alt="book.title || '排行榜封面'" class="rank-reference-cover" />
              <div class="rank-reference-book-meta">
                <span class="rank-reference-rank">#{{ String(book.rank).padStart(2, '0') }}</span>
                <span class="rank-reference-title">{{ book.title || '未命名作品' }}</span>
                <span class="rank-reference-author">{{ book.author || '未知作者' }}</span>
              </div>
            </button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 作品信息查看弹窗 -->
    <el-dialog
      v-model="infoDialogVisible"
      title="作品信息"
      width="520px"
      :align-center="true"
      :close-on-click-modal="false"
      append-to-body
      class="book-info-dialog"
      destroy-on-close
    >
      <div v-if="infoBook" class="book-info-content">
        <div class="book-info-header">
          <div class="book-info-cover" @click="infoBook.cover && handlePreviewBookCover(infoBook)">
            <img v-if="infoBook.cover" :src="infoBook.cover" class="book-info-cover-img" alt="封面" />
            <div v-else class="book-info-cover-placeholder">
              <el-icon><Notebook /></el-icon>
            </div>
          </div>
          <div class="book-info-summary">
            <h3 class="book-info-title">{{ infoBook.title || '未命名作品' }}</h3>
            <div class="book-info-tags">
              <el-tag v-if="infoBook.category" type="primary" size="small" effect="light">
                {{ infoBook.category }}
              </el-tag>
              <el-tag :type="statusTagType(infoBook.status)" size="small" effect="light">
                {{ getStatusText(infoBook.status) }}
              </el-tag>
              <el-tag :type="infoBook.is_public ? 'success' : 'info'" size="small" effect="light">
                {{ infoBook.is_public ? '公开' : '私密' }}
              </el-tag>
            </div>
            <div class="book-info-author" v-if="infoBook.author">
              <el-icon><User /></el-icon>
              <span>{{ infoBook.author }}</span>
            </div>
          </div>
        </div>

        <el-divider class="book-info-divider" />

        <div class="book-info-fields">
          <div class="book-info-field">
            <span class="book-info-label">作品 ID</span>
            <span class="book-info-value">#{{ infoBook.id }}</span>
          </div>
          <div class="book-info-field">
            <span class="book-info-label">作者</span>
            <span class="book-info-value">{{ infoBook.author || '未填写' }}</span>
          </div>
          <div class="book-info-field">
            <span class="book-info-label">分类</span>
            <span class="book-info-value">{{ infoBook.category || '未分类' }}</span>
          </div>
          <div class="book-info-field">
            <span class="book-info-label">状态</span>
            <span class="book-info-value">{{ getStatusText(infoBook.status) }}</span>
          </div>
          <div class="book-info-field">
            <span class="book-info-label">可见性</span>
            <span class="book-info-value">{{ infoBook.is_public ? '公开' : '私密' }}</span>
          </div>
          <div class="book-info-field">
            <span class="book-info-label">创建时间</span>
            <span class="book-info-value">{{ formatDateFull(infoBook.created_at) }}</span>
          </div>
          <div class="book-info-field">
            <span class="book-info-label">更新时间</span>
            <span class="book-info-value">{{ formatDateFull(infoBook.updated_at) }}</span>
          </div>
          <div class="book-info-field book-info-field-full">
            <span class="book-info-label">标签</span>
            <div class="book-info-value book-info-tags-list">
              <el-tag
                v-for="tag in infoBook.tags || []"
                :key="tag"
                size="small"
                effect="plain"
                class="book-info-tag"
              >
                {{ tag }}
              </el-tag>
              <span v-if="!infoBook.tags || infoBook.tags.length === 0" class="book-info-empty">暂无标签</span>
            </div>
          </div>
          <div class="book-info-field book-info-field-full">
            <span class="book-info-label">作品简介</span>
            <div class="book-info-value book-info-description">
              <p v-if="infoBook.description">{{ infoBook.description }}</p>
              <p v-else class="book-info-empty">暂无简介</p>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="book-info-footer">
          <el-button @click="infoDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="handleEditFromInfo">
            <el-icon><Edit /></el-icon>
            编辑作品
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 封面预览弹窗 -->
    <el-dialog
      v-model="coverPreviewDialogVisible"
      title="封面生成结果"
      width="600px"
      destroy-on-close
      class="cover-preview-dialog"
      append-to-body
    >
      <div class="cover-preview-content">
        <div class="cover-preview-header">
          <el-button
            v-if="isGeneratingCover"
            type="danger"
            link
            @click="stopCoverGenerate"
          >
            <el-icon><Close /></el-icon>
            停止
          </el-button>
          <el-button
            v-if="!isGeneratingCover && coverPreviewUrl"
            type="primary"
            link
            @click="downloadCoverImage"
          >
            <el-icon><Download /></el-icon>
            下载
          </el-button>
        </div>
        <div class="cover-preview-image-container">
          <div v-if="isGeneratingCover" class="cover-loading">
            <el-icon class="loading-icon"><Loading /></el-icon>
            <p>正在生成封面...</p>
          </div>
          <img
            v-else-if="coverPreviewUrl"
            :src="coverPreviewUrl"
            class="cover-preview-image"
            alt="生成的封面"
          />
          <div v-else class="cover-empty">
            <el-icon><Picture /></el-icon>
            <p>暂无生成结果</p>
          </div>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="coverPreviewDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :disabled="!coverPreviewUrl || isGeneratingCover"
            @click="confirmCoverPreview"
          >
            使用此封面
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 封面放大预览 -->
    <Teleport to="body">
      <div
        v-if="coverPreviewViewerVisible"
        class="cover-zoom-mask"
        @click="closeCoverZoom"
      >
        <div class="cover-zoom-stage" @click.stop>
          <img
            :src="bookCoverPreviewUrl"
            class="cover-zoom-image"
            alt="封面放大"
          />
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Clock, View, EditPen, Edit, Delete, User, FolderOpened, Lock, Check, Upload, Loading, Notebook, Picture, Download, InfoFilled, Document, Close, MagicStick, ZoomIn } from '@element-plus/icons-vue'
import { useBookStore } from '@/stores/book'
import { promptAPI, aiAPI, chapterAPI, configAPI } from '@/api'
import type { Book, Prompt, Chapter, ApiModel } from '@/types'
import { nextTick } from 'vue'

const router = useRouter()
const bookStore = useBookStore()

const dialogVisible = ref(false)
const isEdit = ref(false)
const generating = ref(false)
const selectedPromptId = ref<number>()
const prompts = ref<Prompt[]>([])
const importFileInput = ref<HTMLInputElement | null>(null)
const importingBookId = ref<number | null>(null)
const importDialogVisible = ref(false)
const selectedFiles = ref<File[]>([])
const isDragOver = ref(false)
const coverDialogVisible = ref(false)
const selectedBookForCover = ref<Book | null>(null)
const coverPreview = ref<string>('')
const coverFileInput = ref<HTMLInputElement | null>(null)
const isCoverDragOver = ref(false)
const referenceImageInput = ref<HTMLInputElement | null>(null)
const referenceImage = ref<{
  name: string
  size: number
  mimeType: string
  dataBase64: string
  previewUrl: string
} | null>(null)
const rankReferenceDialogVisible = ref(false)
const rankReferenceLoading = ref(false)
const rankReferenceGroups = ref<Array<{
  key: string
  label: string
  categories: Array<{
    id: string
    label: string
    path: string
    fullUrl: string
  }>
}>>([])
const activeRankGroupKey = ref('')
const activeRankCategoryPath = ref('')
const rankReferenceBooks = ref<Array<{
  rank: number
  title: string
  author: string
  coverUrl: string
  bookPath: string
}>>([])

// AI Cover Generation
const coverActiveTab = ref('upload')
const isGeneratingCover = ref(false)
const promptsLoading = ref(false)
const modelsLoading = ref(false)
const coverPrompts = ref<Prompt[]>([])
const coverModels = ref<ApiModel[]>([])
const generationForm = ref({
  name: '',
  selectedPromptId: undefined as number | undefined,
  selectedModelId: undefined as number | undefined
})
const generatedCoverUrl = ref('')

const coverPreviewDialogVisible = ref(false)
const coverPreviewUrl = ref('')
const coverAbortController = ref<AbortController | null>(null)

const coverFieldValues = ref<Record<string, string>>({})
const coverPreviewViewerVisible = ref(false)
const bookCoverPreviewUrl = ref('')

type CoverPromptField = { name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }

const coverPromptFields = ref<CoverPromptField[]>([])

const IMAGE_MODEL_PATTERN = /(image|imagen|dall[-_ ]?e|gpt-image|codex-gpt-image|flux|stable[-_ ]?diffusion|seedream|jimeng|midjourney|ideogram|recraft|hidream|playground|pixverse|gemini)/i
const MAX_REFERENCE_IMAGE_SIZE = 20 * 1024 * 1024
const SUPPORTED_REFERENCE_IMAGE_TYPES = /^(image\/(jpeg|png|jpg|webp|gif))$/i

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

const categoryOptions = ['玄幻', '奇幻', '武侠', '仙侠', '都市', '历史', '军事', '游戏', '竞技', '科幻', '灵异', '同人', '轻小说', '其他']

const infoDialogVisible = ref(false)
const infoBook = ref<Book | null>(null)

type ParsedImportChapter = Pick<Chapter, 'title' | 'content'>

const CHAPTER_HEADING_REGEX = /^(?:\s| )*(第\s*[0-9零一二三四五六七八九十百千万两〇○O0-9①-⑳]+?\s*[章节卷回部篇集幕折][^\r\n]{0,30}|(?:序章|序言|前言|楔子|引子|终章|尾声|后记|番外(?:篇|合集)?|大结局)[^\r\n]{0,30}|chapter\s*[0-9ivxlcdm]+[^\r\n]{0,30})\s*$/gim

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

const triggerBookImport = () => {
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
    ElMessage.success(`成功导入 ${successCount} 个文件，共创建 ${successCount} 个作品`)
  } catch (error) {
    if (error instanceof Error && error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    importingBookId.value = null
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

const showFullPrompt = () => {
  const selectedPrompt = coverPrompts.value.find(
    p => p.id === generationForm.value.selectedPromptId
  )
  if (!selectedPrompt) return

  // Build final prompt with variable replacement
  let finalPrompt = selectedPrompt.content || ''
  for (const field of coverPromptFields.value) {
    const value = coverFieldValues.value[field.name] || ''
    const dollarSyntax = new RegExp(`\\$\\{${field.name}\\}`, 'g')
    const curlySyntax = new RegExp(`\\{\\{${field.name}\\}\\}`, 'g')
    finalPrompt = finalPrompt.replace(dollarSyntax, value).replace(curlySyntax, value)
  }

  // Add system prompt with book name
  if (generationForm.value.name) {
    finalPrompt = `小说名称：${generationForm.value.name}\n\n${finalPrompt}`
  }
  
  ElMessageBox.alert(
    `<pre style="white-space: pre-wrap; word-break: break-word; max-height: 400px; overflow-y: auto; font-size: 13px; line-height: 1.6;">${finalPrompt}</pre>`,
    selectedPrompt.name,
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '关闭',
      customClass: 'prompt-preview-dialog'
    }
  )
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

const handleOpenCoverDialog = async (book: Book) => {
  selectedBookForCover.value = book
  coverPreview.value = book.cover || ''
  generatedCoverUrl.value = ''
  coverActiveTab.value = 'upload'
  generationForm.value = {
    name: book.title,
    selectedPromptId: undefined,
    selectedModelId: undefined
  }
  clearReferenceImage()
  coverFieldValues.value = {}
  coverPromptFields.value = []

  promptsLoading.value = true
  modelsLoading.value = true
  try {
    const [promptRes, modelRes] = await Promise.all([
      promptAPI.getAll(),
      configAPI.getAll()
    ])

    if (promptRes.success && promptRes.data) {
      coverPrompts.value = promptRes.data.filter(
        p => p.category === '封面生成'
      )
    }

    if (modelRes.success && modelRes.data) {
      const enabled = modelRes.data.filter(m => m.enabled !== 0)
      coverModels.value = enabled.filter(m =>
        IMAGE_MODEL_PATTERN.test(m.model || m.name || '')
      )
      if (coverModels.value.length === 0) {
        coverModels.value = enabled
      }
      if (coverModels.value.length > 0) {
        const defaultModel = coverModels.value.find(m => m.is_default === 1)
        generationForm.value.selectedModelId = defaultModel
          ? defaultModel.id
          : coverModels.value[0].id
      }
    }
  } catch (error) {
    // fail silently
  } finally {
    promptsLoading.value = false
    modelsLoading.value = false
  }

  coverDialogVisible.value = true
}

const handlePromptSelect = (promptId: number | undefined) => {
  coverFieldValues.value = {}
  coverPromptFields.value = []
  if (!promptId) return

  const selected = coverPrompts.value.find(p => p.id === promptId)
  if (!selected) return

  const book = selectedBookForCover.value
  
  if (selected.fields && selected.fields.length > 0) {
    coverPromptFields.value = selected.fields as CoverPromptField[]
  } else {
    // Fallback to parsing placeholders from content
    const placeholderRegex = /\{\{([a-zA-Z0-9_]+)\}\}|\$\{([a-zA-Z0-9_]+)\}/g
    const placeholders = new Set<string>()
    let match
    while ((match = placeholderRegex.exec(selected.content || '')) !== null) {
      placeholders.add(match[1] || match[2])
    }
    
    coverPromptFields.value = Array.from(placeholders).map(name => ({
      name,
      label: name,
      type: 'text',
      options: [],
      description: '',
      required: true
    }))
  }

  // Pre-fill values
  coverPromptFields.value.forEach(field => {
    if (field.name === 'title' && generationForm.value.name) {
      coverFieldValues.value.title = generationForm.value.name
    }
    if (field.name === 'author' && book?.author) {
      coverFieldValues.value.author = book.author
    }
  })
}

const handleGenerateCover = async () => {
  if (!generationForm.value.selectedModelId) {
    ElMessage.warning('请选择生成模型')
    return
  }
  if (!generationForm.value.selectedPromptId) {
    ElMessage.warning('请选择提示词')
    return
  }
  if (isGeneratingCover.value) return

  const selectedPrompt = coverPrompts.value.find(
    p => p.id === generationForm.value.selectedPromptId
  )
  if (!selectedPrompt || !selectedPrompt.content) {
    ElMessage.warning('所选提示词内容为空')
    return
  }

  let finalPrompt = selectedPrompt.content
  for (const field of coverPromptFields.value) {
    const value = coverFieldValues.value[field.name] || ''
    const dollarSyntax = new RegExp(`\\$\\{${field.name}\\}`, 'g')
    const curlySyntax = new RegExp(`\\{\\{${field.name}\\}\\}`, 'g')
    finalPrompt = finalPrompt.replace(dollarSyntax, value).replace(curlySyntax, value)
  }

  // Add system prompt with book name
  if (generationForm.value.name) {
    finalPrompt = `小说名称：${generationForm.value.name}\n\n${finalPrompt}`
  }

  isGeneratingCover.value = true
  generatedCoverUrl.value = ''
  coverPreviewUrl.value = ''
  coverPreviewDialogVisible.value = true
  const abortController = new AbortController()
  coverAbortController.value = abortController
  
  try {
    const res = await aiAPI.generateImage({
      messages: [{ role: 'user', content: finalPrompt }],
      configId: generationForm.value.selectedModelId,
      size: '600x800',
      reference_images: referenceImage.value
        ? [{
            name: referenceImage.value.name,
            size: referenceImage.value.size,
            mime_type: referenceImage.value.mimeType,
            data_base64: referenceImage.value.dataBase64
          }]
        : undefined
    }, {
      signal: abortController.signal
    })
    if (res.success && res.data?.url) {
      generatedCoverUrl.value = res.data.url
      coverPreviewUrl.value = res.data.url
      ElMessage.success('封面生成成功')
    } else {
      ElMessage.error(res.message || '生成失败')
    }
  } catch (error: any) {
    if (abortController.signal.aborted || error.name === 'AbortError' || error.code === 'ERR_CANCELED' || error.message === 'canceled') {
      ElMessage.info('已停止生成')
      return
    }
    if (error.name !== 'AbortError' && error.message !== 'canceled') {
      ElMessage.error('生成请求失败')
    }
  } finally {
    if (coverAbortController.value === abortController) {
      coverAbortController.value = null
    }
    isGeneratingCover.value = false
  }
}

const stopCoverGenerate = () => {
  if (coverAbortController.value) {
    coverAbortController.value.abort()
    isGeneratingCover.value = false
  }
}

const triggerReferenceImageSelect = () => {
  if (referenceImageInput.value) {
    referenceImageInput.value.value = ''
    referenceImageInput.value.click()
  }
}

const clearReferenceImage = () => {
  if (referenceImage.value?.previewUrl) {
    URL.revokeObjectURL(referenceImage.value.previewUrl)
  }
  referenceImage.value = null
}

const setReferenceImageFromRemote = async (imageUrl: string, name: string) => {
  const response = await fetch(imageUrl)
  const blob = await response.blob()
  if (blob.size > MAX_REFERENCE_IMAGE_SIZE) {
    throw new Error('排行榜封面大小超过 20MB，无法作为参考图')
  }
  const mimeType = blob.type || 'image/jpeg'
  if (!SUPPORTED_REFERENCE_IMAGE_TYPES.test(mimeType)) {
    throw new Error('排行榜封面格式不受支持')
  }
  const dataBase64 = await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      resolve(result.includes(',') ? result.split(',')[1] : result)
    }
    reader.onerror = () => reject(new Error('排行榜封面读取失败'))
    reader.readAsDataURL(blob)
  })
  clearReferenceImage()
  referenceImage.value = {
    name,
    size: blob.size,
    mimeType,
    dataBase64,
    previewUrl: imageUrl
  }
}

const loadRankReferenceGroups = async () => {
  const res = await aiAPI.getRankReferenceCategories()
  if (!res.success || !res.data) {
    throw new Error(res.message || '加载排行榜分类失败')
  }
  rankReferenceGroups.value = res.data
  if (!activeRankGroupKey.value && res.data.length > 0) {
    activeRankGroupKey.value = res.data[0].key
  }
}

const activeRankGroup = computed(() =>
  rankReferenceGroups.value.find(group => group.key === activeRankGroupKey.value) || null
)

const loadRankReferenceBooks = async (path: string) => {
  rankReferenceLoading.value = true
  try {
    const res = await aiAPI.getRankReferenceBooks(path)
    if (!res.success || !res.data) {
      throw new Error(res.message || '加载排行榜封面失败')
    }
    rankReferenceBooks.value = res.data
  } finally {
    rankReferenceLoading.value = false
  }
}

const selectRankGroup = async (groupKey: string) => {
  activeRankGroupKey.value = groupKey
  const group = rankReferenceGroups.value.find(item => item.key === groupKey)
  if (!group || group.categories.length === 0) {
    activeRankCategoryPath.value = ''
    rankReferenceBooks.value = []
    return
  }
  await selectRankCategory(group.categories[0].path)
}

const selectRankCategory = async (path: string) => {
  activeRankCategoryPath.value = path
  await loadRankReferenceBooks(path)
}

const openRankReferenceDialog = async () => {
  rankReferenceDialogVisible.value = true
  if (rankReferenceGroups.value.length === 0) {
    try {
      rankReferenceLoading.value = true
      await loadRankReferenceGroups()
      if (activeRankGroup.value?.categories?.[0]?.path) {
        await selectRankCategory(activeRankGroup.value.categories[0].path)
      }
    } catch (error: any) {
      ElMessage.error(error.message || '加载排行榜参考图失败')
    } finally {
      rankReferenceLoading.value = false
    }
    return
  }

  if (!activeRankCategoryPath.value && activeRankGroup.value?.categories?.[0]?.path) {
    try {
      await selectRankCategory(activeRankGroup.value.categories[0].path)
    } catch (error: any) {
      ElMessage.error(error.message || '加载排行榜参考图失败')
    }
  }
}

const useRankReferenceBook = async (book: { rank: number; title: string; author: string; coverUrl: string }) => {
  try {
    await setReferenceImageFromRemote(book.coverUrl, `${book.title || 'rank-cover'}_${book.rank}.jpg`)
    rankReferenceDialogVisible.value = false
    ElMessage.success('已选中排行榜封面作为参考图')
  } catch (error: any) {
    ElMessage.error(error.message || '设置排行榜参考图失败')
  }
}

const handleReferenceImageSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return
  if (!SUPPORTED_REFERENCE_IMAGE_TYPES.test(file.type)) {
    ElMessage.error('请上传 JPG、PNG、WEBP 或 GIF 格式的图片')
    target.value = ''
    return
  }
  if (file.size > MAX_REFERENCE_IMAGE_SIZE) {
    ElMessage.error('参考图大小不能超过 20MB')
    target.value = ''
    return
  }

  const reader = new FileReader()
  reader.onload = () => {
    const result = typeof reader.result === 'string' ? reader.result : ''
    const base64 = result.includes(',') ? result.split(',')[1] : result
    clearReferenceImage()
    referenceImage.value = {
      name: file.name,
      size: file.size,
      mimeType: file.type || 'image/png',
      dataBase64: base64,
      previewUrl: URL.createObjectURL(file)
    }
  }
  reader.onerror = () => {
    ElMessage.error('参考图读取失败')
  }
  reader.readAsDataURL(file)
}

const getCoverDownloadName = () => {
  const title = selectedBookForCover.value?.title?.trim() || 'cover'
  return `${title}_${Date.now()}.png`
}

const downloadCoverByUrl = async (imageUrl: string) => {
  try {
    const response = await fetch(imageUrl)
    const blob = await response.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = getCoverDownloadName()
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    ElMessage.success('下载成功')
  } catch (error) {
    ElMessage.error('下载失败')
  }
}

const downloadCoverImage = async () => {
  if (!coverPreviewUrl.value) return
  await downloadCoverByUrl(coverPreviewUrl.value)
}

const downloadCurrentCover = async () => {
  if (!coverPreview.value) return
  await downloadCoverByUrl(coverPreview.value)
}

const confirmCoverPreview = () => {
  if (coverPreviewUrl.value) {
    coverPreview.value = coverPreviewUrl.value
    coverActiveTab.value = 'upload'
    coverPreviewDialogVisible.value = false
    coverDialogVisible.value = false
    handleSaveCover()
  }
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

const showCoverPreview = () => {
  if (!coverPreview.value) return
  bookCoverPreviewUrl.value = coverPreview.value
  coverPreviewViewerVisible.value = true
}

const handleShowInfo = (book: Book) => {
  infoBook.value = book
  infoDialogVisible.value = true
}

const handleEditFromInfo = () => {
  if (!infoBook.value) return
  const book = infoBook.value
  infoDialogVisible.value = false
  handleEdit(book)
}

const statusTagType = (status?: string): 'success' | 'warning' | 'info' => {
  if (status === 'completed') return 'success'
  if (status === 'published') return 'warning'
  return 'info'
}

const formatDateFull = (dateStr?: string) => {
  if (!dateStr) return '未知'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const handlePreviewBookCover = (book: Book) => {
  if (!book.cover) return
  bookCoverPreviewUrl.value = book.cover
  coverPreviewViewerVisible.value = true
}

const closeCoverZoom = () => {
  coverPreviewViewerVisible.value = false
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

.cover-zoom-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  cursor: zoom-out;
}

.cover-zoom-stage {
  cursor: default;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-zoom-image {
  display: block;
  max-width: 80vw;
  max-height: 80vh;
  object-fit: contain;
  background: #1f2937;
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35);
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
  padding: 8px 0;
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
  position: relative;
}

.cover-preview-image {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.cover-preview-zoom-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border-radius: 50%;
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s;
}

.cover-preview-zoom-btn:hover {
  background: rgba(0, 0, 0, 0.7);
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

/* ========== 作品信息弹窗 ========== */
.book-info-dialog :deep(.el-dialog) {
  width: 520px !important;
  max-width: calc(100vw - 32px);
  margin: 0 auto !important;
  border-radius: 12px;
  overflow: hidden;
}

.book-info-dialog :deep(.el-dialog__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  margin-right: 0;
}

.book-info-dialog :deep(.el-dialog__body) {
  padding: 20px 24px;
  max-height: 60vh;
  overflow-y: auto;
}

.book-info-dialog :deep(.el-dialog__footer) {
  padding: 12px 20px;
  border-top: 1px solid #f0f0f0;
}

.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar {
  width: 6px;
}

.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.book-info-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: center;
  justify-content: center;
}

.book-info-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.book-info-header {
  display: flex;
  gap: 16px;
  align-items: center;
}

.book-info-cover {
  width: 88px;
  height: 110px;
  border-radius: 8px;
  overflow: hidden;
  flex-shrink: 0;
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.2s ease;
}

.book-info-cover:hover {
  transform: scale(1.03);
}

.book-info-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.book-info-cover-placeholder {
  color: rgba(255, 255, 255, 0.85);
  font-size: 32px;
}

.book-info-summary {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.book-info-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.book-info-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.book-info-author {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #6b7280;
}

.book-info-divider {
  margin: 16px 0 12px;
}

.book-info-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px 20px;
}

.book-info-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.book-info-field-full {
  grid-column: 1 / -1;
}

.book-info-label {
  font-size: 12px;
  color: #9ca3af;
  font-weight: 500;
}

.book-info-value {
  font-size: 14px;
  color: #1f2937;
  word-break: break-word;
}

.book-info-tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.book-info-tag {
  border-radius: 6px;
}

.book-info-description {
  background: #f9fafb;
  border-radius: 8px;
  padding: 10px 12px;
  font-size: 13px;
  color: #4b5563;
  line-height: 1.7;
  max-height: 140px;
  overflow-y: auto;
}

.book-info-description p {
  margin: 0;
  white-space: pre-wrap;
}

.book-info-empty {
  color: #c0c4cc;
  font-size: 13px;
  font-style: italic;
}

.book-info-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 600px) {
  .book-info-dialog :deep(.el-dialog) {
    width: calc(100vw - 32px) !important;
  }

  .book-info-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .book-info-cover {
    width: 72px;
    height: 90px;
  }

  .book-info-fields {
    grid-template-columns: 1fr;
  }
}

/* 暗色主题 */
:root[data-theme='dark'] .book-info-dialog :deep(.el-dialog__header) {
  border-bottom-color: #334155;
}

:root[data-theme='dark'] .book-info-dialog :deep(.el-dialog__footer) {
  border-top-color: #334155;
}

:root[data-theme='dark'] .book-info-title {
  color: #e5e7eb;
}

:root[data-theme='dark'] .book-info-label {
  color: #6b7280;
}

:root[data-theme='dark'] .book-info-value {
  color: #e5e7eb;
}

:root[data-theme='dark'] .book-info-author {
  color: #9ca3af;
}

:root[data-theme='dark'] .book-info-divider {
  border-top-color: #334155;
}

:root[data-theme='dark'] .book-info-description {
  background: #1e293b;
  color: #cbd5e1;
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

.ai-cover-form {
  padding: 8px 12px;
}

.ai-cover-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.ai-cover-form :deep(.el-form-item__label) {
  line-height: 1.5;
  padding-bottom: 4px;
}

.reference-image-panel {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
}

.reference-image-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.reference-image-hint {
  font-size: 12px;
  color: #909399;
}

.reference-image-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ebeef5;
  border-radius: 10px;
  background: #fafafa;
}

.reference-image-thumb {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: 8px;
  flex-shrink: 0;
}

.reference-image-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
}

.reference-image-name {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reference-image-size {
  font-size: 12px;
  color: #909399;
}

.rank-reference-layout {
  display: grid;
  grid-template-columns: 116px 1fr;
  gap: 10px;
  height: 100%;
  min-height: 0;
}

.rank-reference-groups {
  display: flex;
  flex-direction: column;
  gap: 6px;
  border-right: 1px solid #ebeef5;
  padding-right: 8px;
  min-height: 0;
  overflow-y: auto;
}

.rank-reference-group-btn,
.rank-reference-category-btn,
.rank-reference-card {
  border: none;
  background: transparent;
  cursor: pointer;
}

.rank-reference-group-btn {
  text-align: left;
  padding: 8px 10px;
  border-radius: 8px;
  color: #606266;
  font-size: 12px;
}

.rank-reference-group-btn.active {
  background: #ecf5ff;
  color: #409eff;
  font-weight: 600;
}

.rank-reference-content {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.rank-reference-categories {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  flex-shrink: 0;
  max-height: 72px;
  overflow-y: auto;
  align-content: flex-start;
}

.rank-reference-category-btn {
  padding: 4px 10px;
  border-radius: 999px;
  background: #f5f7fa;
  color: #606266;
  font-size: 12px;
}

.rank-reference-category-btn.active {
  background: #409eff;
  color: #fff;
}

.rank-reference-state {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.rank-reference-grid-wrap {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 6px;
}

.rank-reference-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(110px, 1fr));
  gap: 10px;
  align-content: start;
}

.rank-reference-card {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  border: 1px solid #ebeef5;
  border-radius: 12px;
  text-align: left;
  transition: border-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
  min-height: 214px;
  height: 214px;
  overflow: hidden;
}

.rank-reference-card:hover {
  border-color: #409eff;
  transform: translateY(-2px);
  box-shadow: 0 8px 20px rgba(64, 158, 255, 0.12);
}

.rank-reference-cover {
  width: 100%;
  height: 124px;
  object-fit: cover;
  border-radius: 8px;
  background: #f5f7fa;
  flex-shrink: 0;
}

.rank-reference-book-meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  flex: 1;
  overflow: hidden;
}

.rank-reference-rank {
  font-size: 11px;
  color: #f56c6c;
  font-weight: 600;
}

.rank-reference-title {
  font-size: 12px;
  color: #303133;
  line-height: 1.45;
  min-height: 34px;
  max-height: 34px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.rank-reference-author {
  font-size: 11px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.rank-reference-dialog :deep(.el-dialog) {
  width: min(880px, calc(100vw - 32px)) !important;
  max-width: calc(100vw - 32px);
  height: min(360px, calc(100vh - 32px));
  max-height: calc(100vh - 32px);
  margin: 0 auto !important;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.rank-reference-dialog :deep(.el-dialog__header) {
  flex-shrink: 0;
}

.rank-reference-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
  padding: 12px;
  overflow: hidden;
}

.rank-reference-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  padding: 16px;
  box-sizing: border-box;
}

.field-label-text {
  margin-right: 8px;
}

.field-required-tag {
  margin-left: 4px;
  transform: scale(0.85);
}

.cover-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.cover-dialog-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.generation-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}

.generate-btn {
  width: 100%;
}

.generate-btn .el-icon {
  margin-right: 6px;
}

.cover-tabs {
  margin: -10px 0;
}

.cover-tabs :deep(.el-tabs__header) {
  margin-bottom: 10px;
}

.cover-tabs :deep(.el-tabs__item) {
  height: 36px;
  line-height: 36px;
}

.cover-dialog :deep(.el-dialog) {
  width: 600px !important;
  max-width: 90vw;
  margin-top: 8vh !important;
}

.cover-dialog :deep(.el-dialog__body) {
  padding: 15px 20px;
  height: 500px;
  max-height: 60vh;
  overflow-y: auto;
  overflow-x: hidden;
}

.cover-dialog :deep(.el-dialog__body)::-webkit-scrollbar {
  width: 6px;
}

.cover-dialog :deep(.el-dialog__body)::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.cover-dialog :deep(.el-dialog__body)::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.cover-dialog :deep(.el-dialog__body)::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.cover-field-select {
  width: 100%;
}

:deep(.cover-field-select-popper .el-select-dropdown__wrap),
:deep(.cover-field-select-popper .el-scrollbar__wrap) {
  max-height: 320px;
}

:deep(.cover-field-select-popper .el-scrollbar__bar.is-vertical) {
  opacity: 1;
}

.cover-preview-dialog .cover-preview-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.cover-preview-dialog .cover-preview-header {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.cover-preview-dialog .cover-preview-image-container {
  width: 100%;
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
  border-radius: 8px;
  overflow: hidden;
}

.cover-preview-dialog .cover-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  color: #909399;
}

.cover-preview-dialog .cover-loading .loading-icon {
  font-size: 48px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.cover-preview-dialog .cover-preview-image {
  max-width: 100%;
  max-height: 500px;
  object-fit: contain;
}

.cover-preview-dialog .cover-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #c0c4cc;
}

.cover-preview-dialog .cover-empty .el-icon {
  font-size: 64px;
}
</style>
