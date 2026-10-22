<template>
  <div class="books-container">
    <!-- ===== 仪表盘头部 ===== -->
    <header class="dash-header">
      <div class="header-actions">
        <el-button @click="triggerBookImport()" class="dash-btn">
          <el-icon><Upload /></el-icon>
          导入作品
        </el-button>
        <el-button @click="handleSwitchRoot" class="dash-btn">
          <el-icon><FolderOpened /></el-icon>
          切换目录
        </el-button>
        <el-tooltip content="把手工拷进作品库的目录登记进来（不复制文件）" placement="bottom">
          <el-button @click="handleRebuildIndex" class="dash-btn">
            <el-icon><Refresh /></el-icon>
            重建索引
          </el-button>
        </el-tooltip>
        <el-button type="primary" @click="handleCreate" class="dash-btn">
          <el-icon><Plus /></el-icon>
          创建作品
        </el-button>
      </div>
    </header>

    <!-- ===== 筛选与排序 ===== -->
    <div class="filters">
      <div class="seg">
        <span :class="{ active: filterKey === 'all' }" @click="setFilter('all')">全部</span>
        <span :class="{ active: filterKey === 'draft' }" @click="setFilter('draft')">创作中</span>
        <span :class="{ active: filterKey === 'published' }" @click="setFilter('published')">已发布</span>
        <span :class="{ active: filterKey === 'completed' }" @click="setFilter('completed')">已完结</span>
      </div>
      <span class="sort" @click="sortDesc = !sortDesc">排序：<b>{{ sortDesc ? '最近更新 ↓' : '最早更新 ↑' }}</b></span>
    </div>

    <!-- ===== 作品卡 ===== -->
    <div class="books-grid" v-if="filteredBooks.length">
      <div v-for="book in filteredBooks" :key="book.id" class="book-card" @click="handleWrite(book.id)">
        <div class="card-top">
          <span class="book-title">{{ book.title }}</span>
          <div class="chips">
            <span v-if="book.category" class="chip cat">{{ book.category }}</span>
            <span class="chip" :class="statusChipClass(book.status)">{{ getStatusText(book.status) }}</span>
          </div>
        </div>

        <div class="card-body">
          <div class="book-cover" @click.stop="handlePreviewBookCover(book)">
            <img v-if="coverSrcFor(book)" :src="coverSrcFor(book)" class="cover-image" alt="封面" @error="onCoverError(book)" />
            <div v-else class="cover-placeholder">
              <el-icon class="cover-icon"><Notebook /></el-icon>
            </div>
            <div class="cover-tag">书测</div>
          </div>
          <div class="info">
            <div class="words">
              <template v-if="book.word_count">
                已写 <b>{{ formatWordCount(book.word_count) }}</b> / 目标 {{ formatWordCount(targetWordsOf(book)) }} 字
                <template v-if="book.chapter_count"> · {{ book.chapter_count }} 章</template>
              </template>
              <template v-else-if="book.chapter_count">
                {{ book.chapter_count }} 章 · 目标 {{ formatWordCount(targetWordsOf(book)) }} 字
              </template>
              <template v-else>
                目标 {{ formatWordCount(targetWordsOf(book)) }} 字，还未开写
              </template>
            </div>
            <div class="prog">
              <i :style="{ width: bookProgress(book) + '%' }"></i>
            </div>
            <div class="upd">
              最近更新 {{ formatDateShort(book.updated_at || book.created_at) }} {{ formatTime(book.updated_at || book.created_at) }}
            </div>
            <div class="desc" v-if="book.description">{{ book.description }}</div>
          </div>
        </div>

        <div class="card-bottom" @click.stop>
          <span class="foot-tag">进度 {{ bookProgress(book) }}%</span>
          <div class="action-buttons">
            <el-tooltip content="封面" placement="top">
              <button class="action-btn cover" @click.stop="handleOpenCoverDialog(book)">
                <el-icon><Picture /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="记忆（对话/图谱/拆书/历史）" placement="top">
              <button class="action-btn" @click.stop="openMemoryDialog(book)">
                <el-icon><Collection /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="导出 ZIP" placement="top">
              <button class="action-btn" @click.stop="handleExportZip(book)">
                <el-icon><Files /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="打开文件夹" placement="top">
              <button class="action-btn" @click.stop="handleReveal(book)">
                <el-icon><FolderOpened /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip content="下载" placement="top">
              <button class="action-btn download" @click="handleDownload(book)">
                <el-icon><Download /></el-icon>
              </button>
            </el-tooltip>
            <el-tooltip :content="book.status === 'completed' ? '作品状态' : '确认完结'" placement="top">
              <button
                class="action-btn confirm"
                :class="{ 'is-done': book.status === 'completed' }"
                @click.stop="handleOpenFinishDialog(book)"
              >
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

    <!-- ===== 空状态 ===== -->
    <div class="empty-state" v-else>
      <el-icon><Notebook /></el-icon>
      <p v-if="bookStore.books.length">该分类下暂无作品，换个筛选试试</p>
      <p v-else>还没有作品，点击右上角「创建作品」开始你的第一部小说</p>
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
    <BookImportDialog v-model:visible="importDialogVisible" />

    <el-dialog
      v-model="coverDialogVisible"
      width="900px"
      :close-on-click-modal="false"
      append-to-body
      class="cover-dialog"
    >
      <template #header>
        <div class="cover-dialog-header">
          <span class="cover-dialog-title">设置封面 · {{ selectedBookForCover?.title || '' }}</span>
        </div>
      </template>
      <div class="cover-blend">
        <!-- ===== 左：实时预览 + 前后对比 ===== -->
        <div class="blend-left">
          <span class="blend-label">实 时 预 览</span>

          <div
            class="blend-stage"
            ref="cmpRef"
            :class="{ dragover: isCoverDragOver, comparing: compareMode && coverPreview && originalCoverUrl }"
            @dragenter.prevent="isCoverDragOver = true"
            @dragleave.prevent="isCoverDragOver = false"
            @dragover.prevent
            @drop.prevent="handleCoverDrop"
          >
            <div v-if="isGeneratingCover" class="blend-gen">
              <el-icon class="spin-icon"><Loading /></el-icon>
              <p>正在生成封面...</p>
              <el-button type="danger" size="small" @click="stopCoverGenerate">
                <el-icon><Close /></el-icon>停止
              </el-button>
            </div>
            <template v-else-if="coverPreview">
              <!-- 对比模式下：底层放原封面，上层放新封面并按中线裁切。
                   两者都在同一个 .blend-stage 里，所以切换对比不会多出一块、弹窗也不会变高。 -->
              <img
                v-if="showCompareLayers"
                class="cmp-back"
                :src="originalCoverUrl"
                alt="原封面"
              />
              <img
                class="blend-img"
                :src="coverPreview"
                alt="封面预览"
                :style="showCompareLayers ? { clipPath: 'inset(0 0 0 ' + cmpPos + '%)' } : undefined"
              />
              <template v-if="showCompareLayers">
                <div
                  class="cmp-handle"
                  :style="{ left: cmpPos + '%' }"
                  @mousedown.prevent="cmpStart"
                  @touchstart.prevent="cmpStart"
                ></div>
                <span class="cmp-lbl old">原封面</span>
                <span class="cmp-lbl new">新封面</span>
              </template>
              <button type="button" class="blend-zoom" title="放大预览" @click.stop="showCoverPreview">⤢</button>
            </template>
            <div v-else class="blend-ph" @click="triggerCoverSelect('cover')">
              <span class="ic">📖</span>
              <b>拖入图片 · 或 AI 生成</b>
              <small>JPG / PNG，建议 3:4</small>
            </div>
          </div>

          <div class="view-seg">
            <button type="button" :class="{ on: !compareMode }" @click="setCompareSafe(false)">预览</button>
            <button
              type="button"
              :class="{ on: compareMode }"
              :title="originalCoverUrl ? '拖动中线对比新旧封面' : '本书暂无原封面可对比'"
              @click="setCompareSafe(true)"
            >⇔ 对比原封面</button>
          </div>
        </div>

        <!-- ===== 右：配置面板 ===== -->
        <div class="blend-right">
          <div class="seg">
            <button type="button" :class="{ on: coverActiveTab === 'upload' }" @click="coverActiveTab = 'upload'">上传封面</button>
            <button type="button" :class="{ on: coverActiveTab === 'ai-generate' }" @click="coverActiveTab = 'ai-generate'">AI 生成</button>
          </div>

          <!-- 上传面板（含历史封面画廊） -->
          <div v-show="coverActiveTab === 'upload'" class="blend-form">
            <div class="field">
              <label>选择图片 <span class="opt-hint">比例 3:4 · ≤ 20MB · JPG / PNG</span></label>
              <button type="button" class="ctrl" @click="triggerCoverSelect('cover')">🖼 从电脑选择 JPG / PNG 图片</button>
              <div class="inline-tip">也可以直接把图片拖到左侧预览区，松手即替换封面</div>
            </div>
            <div class="field">
              <label>
                封面画廊 <span class="hist-count">{{ serverGallery.length }} 张</span>
                <button type="button" class="ctrl" style="margin-left: 8px; padding: 2px 8px; font-size: 12px" @click="triggerGallerySelect">
                  ＋ 从文件添加
                </button>
              </label>
              <div class="hist-grid">
                <button
                  v-for="g in serverGallery"
                  :key="g.name"
                  type="button"
                  class="hist-tile"
                  :class="{ sel: g.current }"
                  @click="useServerCover(g)"
                >
                  <img :src="galleryUrl(g.name)" alt="" />
                  <span class="del" title="移入回收站" @click.stop="removeServerCover(g)">✕</span>
                  <span class="ck">✓</span>
                </button>
              </div>
              <div v-if="serverGallery.length === 0" class="hist-empty">暂无画廊封面 — AI 生成的图和被顶掉的旧封面都会自动归档到 gallery/，随时点选复用</div>
            </div>
          </div>

          <!-- AI 面板（原版生成参数区块）
               结构：blend-pane（撑满右栏） = 可滚动的 blend-form + 固定的 blend-foot。
               这样提示词字段再多，也只是内部滚动，弹窗高度不变、按钮永远在视野里。 -->
          <div v-show="coverActiveTab === 'ai-generate'" class="blend-pane">
            <div class="blend-form">
              <!-- 只读信息压成一行：它不可编辑，没必要占一整块高度 -->
              <div class="readonly-row">
                <span class="ro-item">
                  <b>小说名称</b>{{ generationForm.name || '（未命名）' }}
                </span>
                <el-button
                  v-if="generationForm.selectedPromptId"
                  type="primary"
                  link
                  size="small"
                  @click="showFullPrompt"
                >
                  <el-icon><View /></el-icon>查看完整提示词
                </el-button>
              </div>
              <div class="field">
                <label>
                  作者 <span class="opt-hint">（可选，会写进提示词）</span>
                </label>
                <el-input v-model="generationForm.author" placeholder="作者名 / 笔名" clearable />
              </div>
              <div class="field">
                <label>生成模型 <span class="req">*</span></label>
                <el-select
                  v-model="generationForm.selectedModelId"
                  placeholder="选择图片生成模型"
                  :loading="modelsLoading"
                  class="ctrl-select"
                >
                  <el-option
                    v-for="m in coverModels"
                    :key="m.id"
                    :label="`${m.provider_name ? m.provider_name + ' - ' : ''}${m.name}`"
                    :value="m.id"
                  />
                </el-select>
              </div>
              <div class="field">
                <label>提示词 <span class="req">*</span></label>
                <el-select
                  v-model="generationForm.selectedPromptId"
                  placeholder="选择封面生成提示词"
                  clearable
                  :loading="promptsLoading"
                  @change="handlePromptSelect"
                  class="ctrl-select"
                >
                  <el-option
                    v-for="p in coverPrompts"
                    :key="p.id"
                    :label="p.name"
                    :value="p.id"
                  />
                </el-select>
              </div>
              <div class="field">
                <label>参考图 <span class="opt-hint">（可选，风格对齐用）</span></label>
                <div class="reference-image-actions">
                  <el-button @click="triggerReferenceImageSelect">
                    <el-icon><Upload /></el-icon>上传参考图
                  </el-button>
                  <el-button @click="openRankReferenceDialog">
                    <el-icon><Picture /></el-icon>排行榜选择
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
              <template v-if="coverPromptFields.length > 0">
                <!-- 提示词自己声明的字段一律列出来，一个都不删：
                     用户写了什么占位符，就应该能填什么 -->
                <div class="fields-head">
                  <span>提示词字段</span>
                  <span class="opt-hint">共 {{ coverPromptFields.length }} 项，可滚动查看</span>
                </div>
                <div v-for="field in coverPromptFields" :key="field.name" class="field">
                  <label>
                    <span class="field-label-text">{{ field.label }}</span>
                    <el-tag
                      :type="field.required !== false ? 'danger' : 'info'"
                      size="small"
                      class="field-required-tag"
                    >
                      {{ field.required !== false ? '必选' : '选填' }}
                    </el-tag>
                  </label>
                  <el-input
                    v-if="field.type === 'text'"
                    v-model="coverFieldValues[field.name]"
                    :placeholder="`请输入${field.label}`"
                  />
                  <el-input
                    v-else-if="field.type === 'textarea'"
                    v-model="coverFieldValues[field.name]"
                    type="textarea"
                    :rows="2"
                    :placeholder="`请输入${field.label}`"
                  />
                  <el-select
                    v-else-if="field.type === 'select'"
                    v-model="coverFieldValues[field.name]"
                    :placeholder="`请选择${field.label}`"
                    class="cover-field-select"
                    popper-class="cover-field-select-popper"
                  >
                    <el-option
                      v-for="(opt, optionIndex) in field.options"
                      :key="opt"
                      :label="field.optionLabels?.[optionIndex] || opt"
                      :value="opt"
                    />
                  </el-select>
                </div>
              </template>
            </div>
            <div class="blend-foot">
              <button
                type="button"
                class="gen-btn"
                :disabled="!generationForm.selectedModelId || isGeneratingCover"
                @click="handleGenerateCover"
              >
                <el-icon v-if="isGeneratingCover" class="spin"><Loading /></el-icon>
                {{ isGeneratingCover ? '正在生成...' : '✨ 开始生成' }}
              </button>
            </div>
          </div>
        </div>
      </div>
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

    <RankReferenceDialog
      v-model:visible="rankReferenceDialogVisible"
      @select="useRankReferenceBook"
    />

    <!-- 作品信息查看弹窗 · 报刊编辑风格 -->
    <BookInfoDialog
      v-model:visible="infoDialogVisible"
      :book="infoBook"
      @preview="handlePreviewBookCover"
    />

    <!-- 确认完结 · 作品状态选择弹窗 -->
    <BookFinishDialog v-model:visible="finishDialogVisible" :book="finishBook" />

    <!-- 封面放大预览 -->
    <BookCoverViewer v-model:visible="coverPreviewViewerVisible" :url="bookCoverPreviewUrl" />

    <!-- 「记忆」概览：只说数量与异常，不做内容列表 -->
    <el-dialog
      v-model="memoryDialogVisible"
      :title="'记忆 · ' + (memoryBook?.title || '')"
      width="440px"
      class="memory-dialog"
    >
      <div v-loading="memoryLoading" class="memory-body">
        <div class="memory-row" v-for="line in memoryLines" :key="line.label">
          <span class="m-label">{{ line.label }}</span>
          <span class="m-count">{{ line.n }}</span>
        </div>

        <div v-if="memorySummary" class="memory-meta">
          <div v-if="memorySummary.missing" class="m-warn">缺失 {{ memorySummary.missing }} 项</div>
          <div v-if="memorySummary.parse_errors" class="m-warn">解析失败 {{ memorySummary.parse_errors }} 项</div>
          <div v-if="memorySummary.pending_writes" class="m-pending">待写入 {{ memorySummary.pending_writes }} 项</div>
          <div v-if="!memorySummary.missing && !memorySummary.parse_errors && !memorySummary.pending_writes" class="m-ok">
            全部正常
          </div>
        </div>

        <div v-if="memoryFailures.length" class="memory-failures">
          <div class="f-title">这些记录没写成功（文件停在旧内容）</div>
          <div v-for="(f, i) in memoryFailures" :key="i" class="f-item">{{ f }}</div>
        </div>

        <p class="memory-hint">
          记忆文件在作品的 <code>memory/</code> 目录里，随作品一起备份。
          在外部编辑过某个文件后，程序不会覆盖它（会如实提示），改回原样或删掉即可让程序重建。
        </p>
      </div>
      <template #footer>
        <el-button @click="memoryDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="memoryBook && handleReveal(memoryBook)">打开文件夹</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Clock, View, EditPen, Delete, FolderOpened, Lock, Check, Upload, Loading, Notebook, Picture, Download, InfoFilled, Document, Close, ZoomIn, Files, Refresh, Collection } from '@element-plus/icons-vue'
import BookCoverViewer from '@/components/books/BookCoverViewer.vue'
import BookFinishDialog from '@/components/books/BookFinishDialog.vue'
import BookInfoDialog from '@/components/books/BookInfoDialog.vue'
import { targetWordsOf, bookProgress, formatWordCount, getStatusText } from '@/utils/bookDisplay'
import BookImportDialog from '@/components/books/BookImportDialog.vue'
import RankReferenceDialog from '@/components/books/RankReferenceDialog.vue'
import { useBookStore } from '@/stores/book'
import { promptAPI, aiAPI, chapterAPI, configAPI, bookAPI, settingsAPI, memoryAPI, rebuildBookIndex } from '@/api'
import type { LinkReport } from '@/api'
import { filterByTypes } from '@/services/modelType'
import type { Book, Prompt, ApiModel } from '@/types'
import { useCoverCompare } from '@/composables/useCoverCompare'
import type { ParsedImportChapter } from '@/utils/bookImport'
import { readTxtFile, splitTxtIntoChapters, formatFileSize } from '@/utils/bookImport'

const router = useRouter()
const bookStore = useBookStore()

// ===== 仪表盘：筛选 / 排序 =====
const filterKey = ref<'all' | 'draft' | 'published' | 'completed'>('all')
const sortDesc = ref(true)

const setFilter = (key: 'all' | 'draft' | 'published' | 'completed') => {
  filterKey.value = key
}

const filteredBooks = computed(() => {
  let list = bookStore.books
  if (filterKey.value !== 'all') {
    list = list.filter(b => (b.status || 'draft') === filterKey.value)
  }
  const sorted = [...list]
  sorted.sort((a, b) => {
    const ta = new Date(a.updated_at || a.created_at).getTime()
    const tb = new Date(b.updated_at || b.created_at).getTime()
    return sortDesc.value ? tb - ta : ta - tb
  })
  return sorted
})

const statusChipClass = (status?: string) => {
  if (status === 'published') return 'on'
  if (status === 'completed') return 'fin'
  return ''
}

const dialogVisible = ref(false)
const isEdit = ref(false)
const selectedPromptId = ref<number>()
const prompts = ref<Prompt[]>([])
const importDialogVisible = ref(false)
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

// AI Cover Generation
const coverActiveTab = ref('upload')
const isGeneratingCover = ref(false)
const promptsLoading = ref(false)
const modelsLoading = ref(false)
const coverPrompts = ref<Prompt[]>([])
const coverModels = ref<ApiModel[]>([])
const generationForm = ref({
  name: '',
  author: '',
  selectedPromptId: undefined as number | undefined,
  selectedModelId: undefined as number | undefined
})
const generatedCoverUrl = ref('')

const coverAbortController = ref<AbortController | null>(null)

// setCompare 自己会校验 books.cover（落盘后为空），所以这里只用它返回的状态，
// 开关走下面的 setCompareSafe。
const { compareMode, cmpPos, cmpDragging, cmpRef, setCmpPos, cmpStart, onCmpDocMove, onCmpDocUp, onCmpDocTouch } = useCoverCompare(coverPreview, selectedBookForCover)

/* 打开弹窗时抓一份"原封面"的地址，作为对比的底层。
 *
 * 为什么不在对比时现算：books.cover 落盘后是空字符串（正文只留文件），
 * 拿它判断"有没有原封面"会永远得到"没有"；而 coverSrcFor 读的是当前封面文件，
 * 保存之后它会变成新图，对比就失去意义。所以这里固定一份。
 */
const originalCoverUrl = ref('')

/** 对比层只有在"有新模式 + 有原封面"时才显示；否则退回普通预览。 */
const showCompareLayers = computed(() => compareMode.value && !!coverPreview.value && !!originalCoverUrl.value)

/** 对比开关：自己校验一次（不能依赖 books.cover，落盘后它是空的）。 */
const setCompareSafe = (v: boolean) => {
  if (v) {
    if (!coverPreview.value) {
      ElMessage.warning('请先上传或生成封面，再进行对比')
      return
    }
    if (!originalCoverUrl.value) {
      ElMessage.warning('本书暂无原封面可对比')
      return
    }
  }
  compareMode.value = v
}

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

const infoDialogVisible = ref(false)
const infoBook = ref<Book | null>(null)

onMounted(async () => {
  await bookStore.fetchBooks()
  const res = await promptAPI.getAll()
  if (res.success && res.data) {
    prompts.value = res.data
  }
  // 前后对比：拖动中线（鼠标 + 触摸）
  document.addEventListener('mousemove', onCmpDocMove)
  document.addEventListener('mouseup', onCmpDocUp)
  document.addEventListener('touchmove', onCmpDocTouch, { passive: false })
  document.addEventListener('touchend', onCmpDocUp)
})

onBeforeUnmount(() => {
  document.removeEventListener('mousemove', onCmpDocMove)
  document.removeEventListener('mouseup', onCmpDocUp)
  document.removeEventListener('touchmove', onCmpDocTouch)
  document.removeEventListener('touchend', onCmpDocUp)
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

// ===== 确认完结：卡片「确认」按钮 → 作品状态选择弹窗 =====

const finishDialogVisible = ref(false)
const finishBook = ref<Book | null>(null)

const handleOpenFinishDialog = (book: Book) => {
  finishBook.value = book
  finishDialogVisible.value = true
}

const handleWrite = (bookId: number) => {
  router.push(`/write/${bookId}`)
}

const triggerBookImport = () => {
  importDialogVisible.value = true
}

/* 提示词里这些名字由上面的基础表单负责（小说名称 / 作者）。
 *
 * ⚠ 注意：这里**只用来给占位符兜底**，不再用来过滤动态字段。
 * 提示词是用户自己写的，字段名中英文混用很常见（实测库里既有 `author`
 * 也有 `作者`）；而且同一个名字在不同提示词里含义可能完全不同
 * （`name` 可能指主角名而不是书名）。替他删字段 = 让他填不了，所以一个都不删。 */
const BASE_FIELD_ALIASES: Record<'name' | 'author', string[]> = {
  name: [
    'title', 'book_title', 'booktitle', 'book_name', 'bookname', 'name', 'novel_name', 'novelname',
    '书名', '小说名', '小说名称', '标题', '名称', '作品名', '作品名称'
  ],
  author: [
    'author', 'writer', 'author_name', 'authorname', 'authorName',
    '作者', '作者名', '作者名称', '笔名'
  ]
}

/** 某个动态字段是否其实对应"作品信息"（书名 / 作者）。
 *  同时看 `name` 与 `label`：有的提示词 name 是英文、label 是中文，只认 name 会漏。 */
const baseFieldKind = (field: { name: string; label?: string }): 'name' | 'author' | '' => {
  const keys = [field.name, field.label]
    .filter(Boolean)
    .map(k => String(k).trim().toLowerCase())
  if (keys.some(k => BASE_FIELD_ALIASES.author.includes(k))) return 'author'
  if (keys.some(k => BASE_FIELD_ALIASES.name.includes(k))) return 'name'
  return ''
}

/** 把 `${x}` / `{{x}}` 两种写法都替换掉；名字来自提示词，必须转义正则元字符。 */
const fillPlaceholder = (text: string, name: string, value: string) => {
  const safe = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return text
    .replace(new RegExp(`\\$\\{${safe}\\}`, 'g'), value)
    .replace(new RegExp(`\\{\\{${safe}\\}\\}`, 'g'), value)
}

/**
 * 拼出最终提示词：动态字段 + 基础字段兜底 + 作品信息头。
 *
 * 顺序很重要：**动态字段先填**，再由书名/作者去兜底剩下的占位符。
 * 两者同名时（提示词既写了 `${作者}`、顶部又填了作者），用户在字段里填的值优先 ——
 * 那是他这一次生成的真实意图。反过来先填基础字段的话，动态字段一旦留空
 * 就会把已经填好的值抹掉。
 *
 * 「查看完整提示词」和「开始生成」必须用同一份结果 —— 之前两处各写一遍，
 * 结果加了字段只改了其中一处，预览与实际发给模型的提示词不一致。
 */
const buildCoverPrompt = (): string => {
  const selected = coverPrompts.value.find(p => p.id === generationForm.value.selectedPromptId)
  if (!selected) return ''
  let text = selected.content || ''

  // 1) 动态字段（提示词自己声明的，以用户填的为准）
  for (const field of coverPromptFields.value) {
    text = fillPlaceholder(text, field.name, coverFieldValues.value[field.name] || '')
  }

  // 2) 书名 / 作者兜底：只影响没被动态字段吃掉的占位符
  const name = generationForm.value.name || ''
  const author = generationForm.value.author || ''
  for (const alias of BASE_FIELD_ALIASES.name) text = fillPlaceholder(text, alias, name)
  for (const alias of BASE_FIELD_ALIASES.author) text = fillPlaceholder(text, alias, author)

  // 3) 书名/作者作为"系统信息"放在最前，即使提示词里没有占位符也能带上
  const header: string[] = []
  if (name) header.push(`小说名称：${name}`)
  if (author) header.push(`作者：${author}`)
  return header.length ? `${header.join('\n')}\n\n${text}` : text
}

const showFullPrompt = () => {
  const selectedPrompt = coverPrompts.value.find(
    p => p.id === generationForm.value.selectedPromptId
  )
  if (!selectedPrompt) return

  const finalPrompt = buildCoverPrompt()
  const escaped = finalPrompt
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  ElMessageBox.alert(
    `<pre style="white-space: pre-wrap; word-break: break-word; max-height: 400px; overflow-y: auto; font-size: 13px; line-height: 1.6;">${escaped}</pre>`,
    selectedPrompt.name,
    {
      dangerouslyUseHTMLString: true,
      confirmButtonText: '关闭',
      customClass: 'prompt-preview-dialog'
    }
  )
}

const handleDelete = async (book: Book) => {
  // 三选一：归档保留 / 移入回收站 / 关闭窗口=取消。
  // 两个删除动作都不会真正删掉文件夹，只是"留在作品库"与"进回收站"的区别。
  let mode: 'archive' | 'trash' | '' = ''
  try {
    await ElMessageBox.confirm(
      `《${book.title}》的目录会被保留。
· 归档保留：移到 _归档/，之后还能找回
· 移入回收站：进系统回收站，可从回收站还原
（右上角关闭 = 取消）`,
      '删除作品',
      {
        confirmButtonText: '归档保留',
        cancelButtonText: '移入回收站',
        distinguishCancelAndClose: true,
        type: 'warning'
      }
    )
    mode = 'archive'
  } catch (action) {
    if (action === 'cancel') mode = 'trash'
  }
  if (!mode) return
  try {
    await bookStore.deleteBook(book.id, mode)
    ElMessage.success(mode === 'trash' ? '已移入回收站' : '已归档')
  } catch (e) {
  }
}

/* ---------------- 封面画廊（服务端 gallery/） ---------------- */
const serverGallery = ref<Array<{ name: string; mtime: number; current: boolean }>>([])

const galleryUrl = (name: string) => {
  const b = selectedBookForCover.value
  return b ? bookAPI.galleryUrl(b.id, name) : ''
}

const loadServerGallery = async () => {
  const b = selectedBookForCover.value
  if (!b) {
    serverGallery.value = []
    return
  }
  try {
    const res: any = await bookAPI.gallery(b.id)
    serverGallery.value = res?.data || []
  } catch (e) {
    serverGallery.value = []
  }
}

const useServerCover = async (g: { name: string }) => {
  const b = selectedBookForCover.value
  if (!b) return
  try {
    await bookAPI.activateCover(b.id, g.name)
    coverPreview.value = `/api/books/${b.id}/cover?v=${Date.now()}`
    bumpCover(b.id)
    await loadServerGallery()
    ElMessage.success('已设为封面')
  } catch (e) {
    ElMessage.error('设置封面失败')
  }
}

const removeServerCover = async (g: { name: string }) => {
  const b = selectedBookForCover.value
  if (!b) return
  try {
    await ElMessageBox.confirm('这张画廊封面会移入系统回收站，确定删除？', '删除画廊封面', { type: 'warning' })
    await apiDeleteGallery(b.id, g.name)
    await loadServerGallery()
    ElMessage.success('已移入回收站')
  } catch (e) {
  }
}

const apiDeleteGallery = async (id: number, name: string) => {
  await bookAPI.deleteGallery(id, name)
}

/* 切换目录 = 二选一：
 *   接管 —— 只登记目标目录里的作品，不搬移、不复制、不删除任何文件。
 *           这是"换机器 / 重装后指向原目录就能用"的落点。
 *   搬移 —— 现有行为：把库里的作品挪到新目录，旧目录清空。
 * 所以必须先看一眼目标目录里有什么，再让用户选。
 */
const handleSwitchRoot = async () => {
  let path = ''
  try {
    const picked: any = await settingsAPI.pickFolder()
    if (!picked?.success || picked.data?.cancelled) return
    path = String(picked.data?.path || '').trim()
    if (!path) return

    const current: any = await settingsAPI.getBooksRoot()
    const cur = String(current?.data?.path || '')
    if (cur && cur.toLowerCase() === path.toLowerCase()) {
      ElMessage.info('已经是当前作品库目录')
      return
    }

    // 1) 先预检：只读，不改任何东西
    const previewRes: any = await settingsAPI.previewRoot(path)
    if (!previewRes?.success) return
    const rep: LinkReport = previewRes.data || { root: path, books: 0 }
    const found = rep.books || 0
    const lines = [
      '目标目录：' + path,
      '',
      found > 0
        ? '发现 ' + found + ' 部作品：' + (rep.created || []).concat(rep.existing || []).slice(0, 6).join('、') +
          (found > 6 ? ' 等' : '')
        : '这个目录里还没有作品',
      '',
      '· 接管：登记上面的作品，文件留在原地，不搬移不复制',
      '· 搬移：把当前库的作品移到这个目录'
    ]
    if (found === 0) {
      lines.push('', '（目录里没有作品时，接管等于只把作品库指过去）')
    }
    if (cur) {
      lines.push('', '当前作品库：' + cur, '选择接管后，当前库的作品不会出现在书架里，但文件不会被删除。')
    }

    // 2) 三选一：确认=接管，取消按钮=搬移，关闭=什么都不做
    let mode: 'link' | 'move' | '' = ''
    try {
      await ElMessageBox.confirm(lines.join(String.fromCharCode(10)), '切换作品库目录', {
        confirmButtonText: '接管这个目录',
        cancelButtonText: '搬移过去',
        distinguishCancelAndClose: true,
        type: 'warning'
      })
      mode = 'link'
    } catch (action) {
      if (action === 'cancel') mode = 'move'
    }
    if (!mode) return

    const res: any = mode === 'link'
      ? await settingsAPI.linkRoot(path)
      : await settingsAPI.setBooksRoot(path)
    if (!res?.success) return

    await bookStore.fetchBooks()
    if (mode === 'link') {
      const d: any = res.data || {}
      const n = d.books || 0
      ElMessage.success('已接管 ' + n + ' 部作品' + (d.errors?.length ? '（' + d.errors.length + ' 项有问题）' : ''))
    } else {
      ElMessage.success('已切换作品库目录')
    }
  } catch (e) {
    // 取消或接口报错（接口错误已由拦截器提示）
  }
}

/* 手工拷进作品库的目录，用一次重建索引把它们收进来（不复制文件）。 */
const handleRebuildIndex = async () => {
  try {
    const res: any = await rebuildBookIndex()
    if (!res?.success) return
    const d: any = res.data || {}
    await bookStore.fetchBooks()
    const parts = Object.entries(d.details || {})
      .filter(([, n]) => Number(n) > 0)
      .map(([k, n]) => k + ' ' + n)
    ElMessage.success('已登记 ' + (d.books || 0) + ' 部作品' + (parts.length ? '（' + parts.join('、') + '）' : ''))
  } catch (e) {
    ElMessage.error('重建索引失败')
  }
}

/* ---------------- 「记忆」概览 ---------------- */
const memoryDialogVisible = ref(false)
const memoryBook = ref<Book | null>(null)
const memorySummary = ref<any>(null)
const memoryLoading = ref(false)

const openMemoryDialog = async (book: Book) => {
  memoryBook.value = book
  memoryDialogVisible.value = true
  memoryLoading.value = true
  memorySummary.value = null
  try {
    const res: any = await memoryAPI.summary(book.id)
    memorySummary.value = res?.success ? res.data : null
  } finally {
    memoryLoading.value = false
  }
}

/** 概览里的数字：把 counts 映射成给人看的一行。 */
const memoryLines = computed(() => {
  const c = memorySummary.value?.counts || {}
  return [
    { label: '对话', n: c.chats || 0 },
    { label: 'Agent 会话', n: c.agents || 0 },
    { label: '图谱版本', n: c.graph_versions || 0 },
    { label: '拆书', n: c.analysis || 0 },
    { label: '生成历史', n: c.generation || 0 }
  ]
})

/** 放弃写入的条目：这些文件停在旧内容上，必须让用户看见。 */
const memoryFailures = computed(() => {
  const f = memorySummary.value?.write_failures || {}
  return Object.entries(f).map(([k, v]) => k + '：' + v)
})

const handleExportZip = (book: Book) => {
  window.open(bookAPI.exportUrl(book.id), '_blank')
}

const handleReveal = async (book: Book) => {
  try {
    await bookAPI.reveal(book.id)
  } catch (e) {
    ElMessage.error('打开文件夹失败')
  }
}

/* 封面：新书的封面是磁盘文件（covers/），旧数据才内联 base64。
   文件不存在时（404）用 failedCoverIds 记住，回退到占位图，避免裂图。 */
const failedCoverIds = ref<Set<number>>(new Set())
// 保存封面后 updated_at 可能只精确到秒，URL 不变会命中旧缓存/旧的 404 结果。
// 用一个本地版本号强制换 URL，同时把失败标记清掉，卡片立刻重新拉图。
const coverVersion = ref<Record<number, number>>({})
const coverSrcFor = (b: Book) => {
  if (b.cover) return b.cover
  if (failedCoverIds.value.has(b.id)) return ''
  const v = coverVersion.value[b.id] || b.updated_at || ''
  return `/api/books/${b.id}/cover?v=${v}`
}
const bumpCover = (id: number) => {
  const next = new Set(failedCoverIds.value)
  next.delete(id)
  failedCoverIds.value = next
  coverVersion.value = { ...coverVersion.value, [id]: Date.now() }
}
const onCoverError = (b: Book) => {
  const next = new Set(failedCoverIds.value)
  next.add(b.id)
  failedCoverIds.value = next
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
  const current = coverSrcFor(book)
  coverPreview.value = current
  // 对比的底层固定成"打开时的封面"：保存之后当前封面会变，现算就没法对比了
  originalCoverUrl.value = current
  generatedCoverUrl.value = ''
  coverActiveTab.value = 'upload'
  compareMode.value = false
  cmpPos.value = 50
  generationForm.value = {
    name: book.title,
    author: book.author || '',
    selectedPromptId: undefined,
    selectedModelId: undefined
  }
  clearReferenceImage()
  // 画廊来自 gallery/ 目录（AI 生成图与被顶掉的封面都在这里），不再读 localStorage
  void loadServerGallery()
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
      // 优先按后端强类型 config_type='image' 过滤；老数据（未标类型）再用模型 ID 正则兜底
      const typed = filterByTypes(enabled, 'image')
      coverModels.value = typed.length > 0
        ? typed
        : enabled.filter(m => IMAGE_MODEL_PATTERN.test(m.model || m.name || ''))
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

  if (selected.fields && selected.fields.length > 0) {
    coverPromptFields.value = selected.fields as CoverPromptField[]
  } else {
    // 提示词没存 fields 配置时，从正文里把占位符抠出来。
    //
    // ⚠ 名字**不能限定 ASCII**：中文提示词里 `${书名}` `${作者}` `${风格}` 是主流写法，
    // 之前只认 [a-zA-Z0-9_] 会把它们全部漏掉，表现就是"提示词里的字段不见了"。
    // 这里跟 usePromptFields 的提取口径保持一致（任意非花括号字符）。
    const placeholderRegex = /\{\{([^{}]+)\}\}|\$\{([^{}]+)\}/g
    const placeholders = new Set<string>()
    let match
    while ((match = placeholderRegex.exec(selected.content || '')) !== null) {
      const name = (match[1] || match[2] || '').trim()
      if (name) placeholders.add(name)
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

  // 作者优先用作品里已有的；没有就留空让用户填（封面生成经常需要它）
  if (!generationForm.value.author && selectedBookForCover.value?.author) {
    generationForm.value.author = selectedBookForCover.value.author
  }

  // 提示词自己声明的字段**一个都不删**（用户写了什么占位符就该能填什么）。
  // 只把"作品信息"类字段的默认值带过来，省得同一个值手填两遍。
  for (const field of coverPromptFields.value) {
    if (coverFieldValues.value[field.name]) continue
    const kind = baseFieldKind(field)
    if (kind === 'name') coverFieldValues.value[field.name] = generationForm.value.name || ''
    else if (kind === 'author') coverFieldValues.value[field.name] = generationForm.value.author || ''
  }
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

  // 与「查看完整提示词」共用同一个拼装函数，保证看到的就是发出去的
  const finalPrompt = buildCoverPrompt()

  isGeneratingCover.value = true
  generatedCoverUrl.value = ''
  const abortController = new AbortController()
  coverAbortController.value = abortController

  try {
    const res = await aiAPI.generateImage({
      messages: [{ role: 'user', content: finalPrompt }],
      configId: generationForm.value.selectedModelId,
      // 带上作品 id：后端会把图直接落进 books/<书名>/gallery/，与作品绑在一起
      bookId: selectedBookForCover.value?.id,
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
      coverPreview.value = res.data.url
      // 新图已进 gallery/，立刻同步一次，用户不用等关窗再开
      void loadServerGallery()
      ElMessage.success('封面生成成功，已更新预览')
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

const openRankReferenceDialog = () => {
  rankReferenceDialogVisible.value = true
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

const downloadCurrentCover = async () => {
  if (!coverPreview.value) return
  await downloadCoverByUrl(coverPreview.value)
}

const coverUploadTarget = ref<'cover' | 'gallery'>('cover')

// target=gallery 时只把图加进画廊，不替换当前封面
const triggerCoverSelect = (target: 'cover' | 'gallery' = 'cover') => {
  coverUploadTarget.value = target
  if (coverFileInput.value) {
    coverFileInput.value.value = ''
    coverFileInput.value.click()
  }
}

const triggerGallerySelect = () => triggerCoverSelect('gallery')

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
  reader.onload = async (e) => {
    const src = e.target?.result as string
    if (coverUploadTarget.value === 'gallery') {
      const b = selectedBookForCover.value
      if (!b) return
      try {
        await bookAPI.addGallery(b.id, src)
        await loadServerGallery()
        ElMessage.success('已加入封面画廊')
      } catch (err) {
        ElMessage.error('加入画廊失败')
      }
      return
    }
    coverPreview.value = src
    // 不再往 localStorage 里另存一份"历史"：画廊就是 gallery/ 目录，
    // 旧封面由后端在保存时自动归档进去。两处各存一套只会让人以为"历史丢了"。
  }
  reader.readAsDataURL(file)
}

const handleSaveCover = async () => {
  if (!selectedBookForCover.value || !coverPreview.value) return

  const bookId = selectedBookForCover.value.id
  try {
    // coverPreview 可能是 data URL（本地上传）、作品 image/ 里的生成图、
    // 或历史遗留的缓存地址 —— 三种后端都会落到 covers/ 并把上一张归档进 gallery/。
    await bookStore.updateBook(bookId, {
      cover: coverPreview.value
    })
    // 先把画廊同步一遍再关窗：下次打开就是最新的，不用靠"重新进页面"
    await loadServerGallery()
    bumpCover(bookId)
    const archived = Math.max(0, serverGallery.value.length)
    ElMessage.success(archived > 0 ? `封面已更新，旧封面已存入画廊（共 ${archived} 张）` : '封面已更新')
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

const handlePreviewBookCover = (book: Book) => {
  const src = coverSrcFor(book)
  if (!src) return
  bookCoverPreviewUrl.value = src
  coverPreviewViewerVisible.value = true
}

const closeCoverZoom = () => {
  coverPreviewViewerVisible.value = false
}
</script>

<style scoped>
.books-container {
  --c-bg: #f4f6fb;
  --c-card: #ffffff;
  --c-border: #e7eaf2;
  --c-text: #23283a;
  --c-sub: #6b7385;
  --c-muted: #9aa1b3;
  --c-line: #eef0f6;
  --c-accent: #6c8cff;
  --c-accent2: #9c6cff;
  background: var(--c-bg);
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 26px 32px 0;
  transition: background .3s;
}

/* ---------- 仪表盘头部 ---------- */
.dash-header {
  flex-shrink: 0;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-bottom: 22px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.dash-btn.el-button {
  border-radius: 12px;
  font-weight: 600;
  border-color: var(--c-border);
  color: var(--c-sub);
  background: var(--c-card);
  box-shadow: 0 2px 8px rgba(30, 40, 60, .04);
  transition: all .18s;
}

.dash-btn.el-button:hover {
  border-color: var(--c-accent);
  color: var(--c-accent);
  box-shadow: 0 4px 14px rgba(108, 140, 255, .18);
}

.dash-btn.el-button--primary {
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  border: none;
  color: #fff;
}

.dash-btn.el-button--primary:hover {
  box-shadow: 0 8px 22px rgba(108, 140, 255, .32);
  color: #fff;
}

/* ---------- 筛选与排序 ---------- */
.filters {
  flex-shrink: 0;
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.seg {
  display: flex;
  gap: 4px;
  background: var(--c-line);
  padding: 4px;
  border-radius: 12px;
}

.seg span {
  padding: 7px 18px;
  border-radius: 9px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 600;
  color: var(--c-sub);
  transition: all .15s;
}

.seg span:hover { color: var(--c-text); }

.seg span.active {
  background: var(--c-card);
  color: var(--c-accent);
  box-shadow: 0 2px 8px rgba(60, 80, 130, .12);
}

.filters .sort {
  margin-left: auto;
  font-size: 12px;
  color: var(--c-muted);
  cursor: pointer;
  user-select: none;
}

.filters .sort b { color: var(--c-sub); font-weight: 600; }
.filters .sort:hover b { color: var(--c-accent); }

/* ---------- 空状态 ---------- */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 80px 0;
  color: var(--c-muted);
  background: var(--c-card);
  border: 1px dashed var(--c-border);
  border-radius: 16px;
}

.empty-state .el-icon { font-size: 48px; }
.empty-state p { font-size: 13px; }

/* ---------- 作品网格（唯一滚动区域） ---------- */
.books-grid {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
  gap: 16px;
  align-content: start;
  padding: 2px 6px 44px;
}

.book-card {
  background: var(--c-card);
  border: 1px solid var(--c-border);
  border-radius: 16px;
  padding: 16px;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 12px;
  transition: all .2s ease;
}

.book-card:hover {
  transform: translateY(-3px);
  border-color: #c9d4ff;
  box-shadow: 0 12px 30px rgba(60, 80, 130, .1);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.book-title {
  flex: 1;
  min-width: 0;
  font-size: 16px;
  font-weight: 800;
  color: var(--c-text);
  margin: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chips {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.chip {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 9px;
  border-radius: 999px;
  background: var(--c-line);
  color: var(--c-muted);
  white-space: nowrap;
}

.chip.cat { background: #eef1fb; color: #6c8cff; }
.chip.on { background: #e2f7ee; color: #16a365; }
.chip.fin { background: #eef4ff; color: #4f8ff0; }

.card-body {
  display: flex;
  gap: 14px;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.book-cover {
  position: relative;
  width: 100px;
  aspect-ratio: 3 / 4;
  flex-shrink: 0;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 6px 16px rgba(60, 80, 130, .16);
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

.info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 7px;
}

.info .words {
  font-size: 12px;
  color: var(--c-sub);
}

.info .words b {
  color: var(--c-text);
  font-weight: 700;
}

.info .prog {
  height: 5px;
  border-radius: 3px;
  background: var(--c-line);
  overflow: hidden;
}

.info .prog i {
  display: block;
  height: 100%;
  border-radius: 3px;
  background: linear-gradient(90deg, var(--c-accent), var(--c-accent2));
  transition: width .4s ease;
}

.info .upd {
  font-size: 11px;
  color: var(--c-muted);
}

.info .desc {
  font-size: 12px;
  color: var(--c-sub);
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-bottom {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid var(--c-line);
  padding-top: 10px;
  flex-shrink: 0;
}

.foot-tag {
  font-size: 11px;
  color: var(--c-muted);
}

.action-buttons {
  display: flex;
  gap: 6px;
  flex-shrink: 0;
}

.action-btn {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  border: 1px solid var(--c-border);
  background: var(--c-card);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .15s ease;
  font-size: 13px;
  flex-shrink: 0;
  color: var(--c-muted);
}

.action-btn:hover {
  transform: translateY(-1px);
}

.action-btn.create:hover {
  background: var(--c-accent);
  border-color: var(--c-accent);
  color: #fff;
  box-shadow: 0 4px 12px rgba(108, 140, 255, .3);
}

.action-btn.cover:hover {
  background: #22d3ee;
  border-color: #22d3ee;
  color: #fff;
  box-shadow: 0 4px 12px rgba(34, 211, 238, .3);
}

.action-btn.download:hover {
  background: #34d399;
  border-color: #34d399;
  color: #fff;
  box-shadow: 0 4px 12px rgba(52, 211, 153, .3);
}

.action-btn.confirm:hover {
  background: #fbbf24;
  border-color: #fbbf24;
  color: #fff;
  box-shadow: 0 4px 12px rgba(251, 191, 36, .3);
}

/* 已完结的作品：「确认」按钮转为常驻琥珀态，提示可调整状态 */
.action-btn.confirm.is-done {
  border-color: #fbbf24;
  color: #fbbf24;
  background: rgba(251, 191, 36, .12);
}

.action-btn.info:hover {
  background: #60a5fa;
  border-color: #60a5fa;
  color: #fff;
  box-shadow: 0 4px 12px rgba(96, 165, 250, .3);
}

.action-btn.delete:hover {
  background: #ef5350;
  border-color: #ef5350;
  color: #fff;
  box-shadow: 0 4px 12px rgba(239, 83, 80, .3);
}

.action-btn.loading,
.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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
  .books-container {
    padding: 20px 16px 0;
  }

  .dash-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 14px;
  }

  .books-grid {
    grid-template-columns: 1fr;
  }

  .card-body {
    flex-direction: row;
  }

  .book-cover {
    width: 92px;
  }

  .filters .sort {
    margin-left: 0;
  }

  .action-buttons {
    gap: 4px;
  }

  .action-btn {
    width: 26px;
    height: 26px;
    font-size: 12px;
  }
}

.cover-drop-zone.is-dragover {
  border-color: #3b82f6;
  background: #e0edff;
  transform: scale(1.02);
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
:root[data-theme='dark'] .books-container {
  --c-bg: #0d1117;
  --c-card: #161b22;
  --c-border: #21262d;
  --c-text: #e6edf3;
  --c-sub: #8b98a8;
  --c-muted: #4d5768;
  --c-line: #2a3042;
  --c-accent: #6c8cff;
  --c-accent2: #9c6cff;
}

:root[data-theme='dark'] .book-card:hover {
  border-color: #3b4a6b;
  box-shadow: 0 12px 30px rgba(0, 0, 0, .35);
}

:root[data-theme='dark'] .book-cover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, .4);
}

:root[data-theme='dark'] .stat.active {
  box-shadow: 0 0 0 2px rgba(108, 140, 255, .25), 0 10px 26px rgba(0, 0, 0, .25);
}

:root[data-theme='dark'] .seg span.active {
  box-shadow: 0 2px 8px rgba(0, 0, 0, .4);
}

:root[data-theme='dark'] .empty-state {
  background: #161b22;
}

:root[data-theme='dark'] .chip.cat {
  background: #1e2a45;
  color: #8fa7d6;
}

:root[data-theme='dark'] .chip.on {
  background: #143f31;
  color: #4ade80;
}

:root[data-theme='dark'] .chip.fin {
  background: #1e3050;
  color: #8fb6f0;
}

:root[data-theme='dark'] .dash-btn.el-button {
  box-shadow: 0 2px 8px rgba(0, 0, 0, .2);
}

:root[data-theme='dark'] .cover-placeholder {
  background: linear-gradient(135deg, #3b82f6 0%, #06b6d4 100%);
}

:root[data-theme='dark'] .cover-tag {
  background: rgba(0, 0, 0, 0.8);
}

:root[data-theme='dark'] .form-label {
  color: #9ca3af;
}

:root[data-theme='dark'] .cover-drop-zone.is-dragover {
  border-color: #3b82f6;
  background: #1e3a5f;
}

:root[data-theme='dark'] .simple-book-dialog .el-dialog__header {
  border-bottom-color: #21262d;
}

:root[data-theme='dark'] .simple-book-dialog .el-dialog__footer {
  border-top-color: #21262d;
}

:root[data-theme='dark'] .simple-form .form-label {
  color: #9ca3af;
}

/* ========== 参考图 / 排行榜参考（上面「确认完结 · 暗色」到此为止） ========== */
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
  border: 1px solid var(--c-border);
  border-radius: 10px;
  background: var(--c-bg);
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

.cover-dialog :deep(.el-dialog) {
  width: 900px !important;
  max-width: 96vw;
  margin-top: 6vh !important;
}

.cover-dialog :deep(.el-dialog__body) {
  padding: 0;
  /* 高度交给 .cover-blend 自己定：body 再写死一个高度，
     会和内容里的 600px 打架（底部留一条空白，或反过来被内容顶开）。 */
  overflow: hidden;
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

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* ================= 封面弹窗 · 综合重构（01 分屏 + 09 对比 + 03 历史画廊） ================= */
.cover-blend {
  /* 弹窗被 append-to-body 到 <body>，拿不到 .books-container 上的 --c-*。
     这里自带一份，明暗各一套，色彩才不会"飘"。 */
  --c-bg: #f4f6fb;
  --c-card: #ffffff;
  --c-border: #e7eaf2;
  --c-text: #23283a;
  --c-sub: #6b7385;
  --c-muted: #9aa1b3;
  --c-line: #eef0f6;
  --c-accent: #6c8cff;
  --c-accent2: #9c6cff;

  display: grid;
  grid-template-columns: 320px minmax(0, 1fr);
  grid-template-rows: minmax(0, 1fr);
  /* 高度定在这里而不是只靠 el-dialog__body：
     动态字段再多，也只有右栏滚动，弹窗不会跟着长。 */
  height: min(600px, 76vh);
  max-height: min(600px, 76vh);
  overflow: hidden;
}

:root[data-theme='dark'] .cover-blend {
  --c-bg: #0d1117;
  --c-card: #161b22;
  --c-border: #21262d;
  --c-text: #e6edf3;
  --c-sub: #8b98a8;
  --c-muted: #5f6b7a;
  --c-line: #21262d;
  --c-accent: #3b82f6;
  --c-accent2: #8b5cf6;

  /* Element Plus 的暗色变量不会自己跟进来（弹窗被 teleport 到 body，
     实测 --el-bg-color 与 --el-fill-color-blank 仍是 #ffffff）：
     "小说名称"这类 disabled 输入框、输入框描边因此全是浅色，在深底上很刺眼。
     挂在 .cover-blend 上最省事 —— 变量会往下继承给所有内部控件。 */
  --el-bg-color: #161b22;
  --el-bg-color-overlay: #161b22;
  --el-bg-color-page: #0d1117;
  --el-fill-color-blank: #1c2333;
  --el-fill-color: #1c2333;
  --el-fill-color-light: #1c2333;
  --el-fill-color-lighter: #192031;
  --el-border-color: rgba(59, 130, 246, .18);
  --el-border-color-light: rgba(59, 130, 246, .14);
  --el-border-color-lighter: rgba(59, 130, 246, .10);
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #e6edf3;
  --el-text-color-secondary: #8b98a8;
  --el-text-color-placeholder: #4d5768;
  --el-disabled-bg-color: #161b22;
  --el-disabled-text-color: #5f6b7a;
  --el-disabled-border-color: #21262d;
}

/* 输入框/下拉框的框线是 inset box-shadow（不是 border），
   全局暗色只改了 background 与 border-color，所以框线仍然是浅的。 */
:root[data-theme='dark'] .cover-blend :deep(.el-input__wrapper),
:root[data-theme='dark'] .cover-blend :deep(.el-select__wrapper),
:root[data-theme='dark'] .cover-blend :deep(.el-textarea__inner) {
  box-shadow: 0 0 0 1px var(--el-border-color) inset !important;
}

:root[data-theme='dark'] .cover-blend :deep(.el-input.is-disabled .el-input__wrapper) {
  box-shadow: 0 0 0 1px #21262d inset !important;
}

.blend-left {
  background: linear-gradient(180deg, #eef1fb, #f8f6ff);
  border-right: 1px solid var(--c-line);
  padding: 20px 22px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 14px;
  overflow-y: auto;
  min-height: 0;
}

.blend-label {
  font-size: 11px;
  color: var(--c-muted);
  font-weight: 600;
  letter-spacing: 1px;
  align-self: flex-start;
}

.blend-stage {
  width: 230px;
  aspect-ratio: 3 / 4;
  border-radius: 14px;
  overflow: hidden;
  position: relative;
  box-shadow: 0 12px 28px rgba(60, 80, 130, .18);
  background: var(--c-card);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: box-shadow .2s;
  flex-shrink: 0;
}

.blend-stage.dragover {
  outline: 3px dashed var(--c-accent);
  outline-offset: 4px;
  box-shadow: 0 0 0 8px rgba(108, 140, 255, .12);
}

.blend-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.blend-ph {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: var(--c-muted);
  text-align: center;
  padding: 0 16px;
  cursor: pointer;
}

.blend-ph .ic {
  width: 52px;
  height: 52px;
  border-radius: 16px;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
  box-shadow: 0 8px 18px rgba(108, 140, 255, .35);
}

.blend-ph b {
  font-size: 12.5px;
  color: var(--c-sub);
  font-weight: 600;
}

.blend-ph small {
  font-size: 11px;
}

.blend-zoom {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  border: none;
  background: rgba(0, 0, 0, .45);
  color: #fff;
  cursor: pointer;
  font-size: 13px;
  opacity: 0;
  transition: .2s;
}

.blend-stage:hover .blend-zoom {
  opacity: 1;
}

.blend-gen {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  color: var(--c-muted);
  font-size: 13px;
  padding: 0 16px;
  text-align: center;
}

.blend-gen .spin-icon {
  font-size: 42px;
  animation: spin 1s linear infinite;
}

/* 前后对比（原地叠加，不额外占一块）
 *
 * 两张图都叠在 .blend-stage 里：底层是原封面，上层是新封面按中线裁切。
 * 之前对比是另起一个 .blend-cmp 元素接在下面，于是每点一次就多出一块、
 * 弹窗高度直接翻倍 —— 现在与原图共用同一个框，高度恒定。 */
.blend-stage .cmp-back {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  z-index: 1;
  pointer-events: none;
}

.blend-stage.comparing .blend-img {
  position: relative;
  z-index: 2;
}

.blend-stage .cmp-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 3px;
  margin-left: -1.5px;
  background: #fff;
  z-index: 4;
  cursor: ew-resize;
  box-shadow: 0 0 0 1px rgba(0, 0, 0, .15), 0 0 10px rgba(0, 0, 0, .25);
  touch-action: none;
}

.blend-stage .cmp-handle::after {
  content: "";
  position: absolute;
  top: 50%;
  left: 50%;
  width: 28px;
  height: 28px;
  margin: -14px 0 0 -14px;
  border-radius: 50%;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, .3);
}

.blend-stage .cmp-handle::before {
  content: "⇔";
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  z-index: 1;
  font-size: 12px;
  color: #5b6478;
  line-height: 1;
}

.blend-stage .cmp-lbl {
  position: absolute;
  top: 10px;
  z-index: 5;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 10.5px;
  font-weight: 700;
  color: #fff;
  background: rgba(0, 0, 0, .5);
  backdrop-filter: blur(4px);
  pointer-events: none;
}

.blend-stage .cmp-lbl.old { left: 10px; }

.blend-stage .cmp-lbl.new {
  right: 10px;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
}

.view-seg {
  display: flex;
  gap: 4px;
  background: rgba(233, 235, 246, .85);
  padding: 3px;
  border-radius: 10px;
  width: 230px;
  flex-shrink: 0;
}

.view-seg button {
  flex: 1;
  border: none;
  background: transparent;
  padding: 6px 0;
  border-radius: 8px;
  font-size: 11.5px;
  font-weight: 700;
  color: var(--c-sub);
  cursor: pointer;
  transition: .15s;
  font-family: inherit;
}

.view-seg button.on {
  background: var(--c-card);
  color: var(--c-accent);
  box-shadow: 0 2px 6px rgba(60, 80, 130, .14);
}

/* 右侧面板 */
.blend-right {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.blend-right .seg {
  display: flex;
  gap: 4px;
  background: var(--c-line);
  padding: 4px;
  border-radius: 12px;
  margin: 16px 20px 4px;
  flex-shrink: 0;
}

.blend-right .seg button {
  flex: 1;
  border: none;
  background: transparent;
  padding: 8px 0;
  border-radius: 9px;
  font-size: 12.5px;
  font-weight: 700;
  color: var(--c-sub);
  cursor: pointer;
  transition: .15s;
  font-family: inherit;
}

.blend-right .seg button.on {
  background: var(--c-card);
  color: var(--c-accent);
  box-shadow: 0 2px 8px rgba(60, 80, 130, .12);
}

/* AI 面板：撑满右栏，内容区滚动、按钮固定。
   提示词字段是用户自己写的，数量不可控 —— 高度只能由容器管住，
   绝不能让它把弹窗顶高（弹窗高度写死在 .cover-blend 上）。 */
.blend-pane {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.blend-foot {
  flex-shrink: 0;
  padding: 10px 20px 16px;
  border-top: 1px solid var(--c-line);
  background: var(--c-card);
}

.blend-form {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 12px 20px 18px;
  display: flex;
  flex-direction: column;
  gap: 13px;
}

/* 只读信息压成一行：不可编辑的东西不该占一整块高度 */
.readonly-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 7px 11px;
  border-radius: 9px;
  background: var(--c-bg);
  border: 1px solid var(--c-border);
}

.readonly-row .ro-item {
  font-size: 12.5px;
  color: var(--c-text);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.readonly-row .ro-item b {
  font-size: 11px;
  font-weight: 600;
  color: var(--c-muted);
  margin-right: 8px;
}

.fields-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  font-weight: 700;
  color: var(--c-sub);
  padding-top: 4px;
  border-top: 1px dashed var(--c-line);
}

/* 行内提示（替代原来占一整块的"格式建议""快捷方式"） */
.inline-tip {
  font-size: 11.5px;
  color: var(--c-muted);
  line-height: 1.6;
}

.blend-form::-webkit-scrollbar {
  width: 5px;
}

.blend-form::-webkit-scrollbar-thumb {
  background: #d5d9e8;
  border-radius: 3px;
}

.blend-form .field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.blend-form .field > label {
  font-size: 12px;
  font-weight: 600;
  color: var(--c-sub);
  display: flex;
  align-items: center;
  gap: 6px;
  line-height: 1.6;
}

.blend-form .field .req { color: #ef5350; }
.blend-form .field .opt-hint { font-size: 10px; color: var(--c-muted); font-weight: 400; }
.blend-form .field .ctrl-select { width: 100%; }

.blend-form .ctrl {
  padding: 9px 12px;
  border: 1px solid var(--c-border);
  border-radius: 10px;
  font-size: 13px;
  font-family: inherit;
  color: var(--c-text);
  background: var(--c-bg);
  outline: none;
  transition: .15s;
  text-align: left;
  cursor: pointer;
}

.blend-form .ctrl:hover { border-color: var(--c-accent); }

.hist-count {
  font-size: 10.5px;
  color: var(--c-muted);
  font-weight: 400;
  background: var(--c-line);
  padding: 1px 8px;
  border-radius: 999px;
}

.hist-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}

.hist-tile {
  position: relative;
  border: 2px solid var(--c-border);
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  padding: 0;
  background: var(--c-card);
  transition: .15s;
}

.hist-tile img {
  width: 100%;
  aspect-ratio: 3 / 4;
  object-fit: cover;
  display: block;
}

.hist-tile:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 14px rgba(60, 80, 130, .15);
}

.hist-tile.sel {
  border-color: var(--c-accent);
  box-shadow: 0 0 0 2px rgba(108, 140, 255, .22);
}

.hist-tile .del {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 18px;
  height: 18px;
  border-radius: 6px;
  border: none;
  background: rgba(0, 0, 0, .5);
  color: #fff;
  font-size: 9px;
  line-height: 18px;
  text-align: center;
  cursor: pointer;
  opacity: 0;
  transition: .15s;
}

.hist-tile:hover .del { opacity: 1; }
.hist-tile .del:hover { background: rgba(239, 83, 80, .9); }

.hist-tile .ck {
  position: absolute;
  left: 4px;
  bottom: 4px;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  display: none;
  align-items: center;
  justify-content: center;
  font-size: 9px;
}

.hist-tile.sel .ck { display: flex; }

.hist-empty {
  font-size: 11.5px;
  color: var(--c-muted);
  line-height: 1.7;
  border: 1px dashed var(--c-border);
  border-radius: 10px;
  padding: 12px;
  text-align: center;
  background: var(--c-bg);
}

.gen-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  width: 100%;
  padding: 12px 0;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, var(--c-accent), var(--c-accent2));
  color: #fff;
  font-size: 13.5px;
  font-weight: 800;
  cursor: pointer;
  font-family: inherit;
  box-shadow: 0 8px 20px rgba(108, 140, 255, .35);
  transition: .18s;
}

.gen-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 12px 26px rgba(108, 140, 255, .45);
}

.gen-btn:disabled {
  opacity: .55;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.gen-btn .spin {
  animation: spin 1s linear infinite;
}

.field-required-tag {
  margin-left: 4px;
  transform: scale(.85);
}

/* 暗色适配
 *
 * 调色板已经在 .cover-blend 上换成暗色一套，这里只处理"没用变量"的地方：
 * 渐变、阴影、滚动条、hover 底色。 */
:root[data-theme='dark'] .blend-left {
  background: linear-gradient(180deg, #141a22, #10161d);
  border-right-color: #21262d;
}

:root[data-theme='dark'] .blend-stage { box-shadow: 0 12px 28px rgba(0, 0, 0, .45); }
:root[data-theme='dark'] .blend-stage.dragover {
  box-shadow: 0 0 0 8px rgba(59, 130, 246, .16);
}

:root[data-theme='dark'] .view-seg { background: rgba(255, 255, 255, .06); }
:root[data-theme='dark'] .view-seg button.on {
  background: #21262d;
  box-shadow: 0 2px 6px rgba(0, 0, 0, .4);
}

:root[data-theme='dark'] .blend-right .seg { background: rgba(255, 255, 255, .05); }
:root[data-theme='dark'] .blend-right .seg button.on {
  background: #21262d;
  box-shadow: 0 2px 8px rgba(0, 0, 0, .4);
}

:root[data-theme='dark'] .blend-form .ctrl:hover { border-color: var(--c-accent); background: #1c2333; }
:root[data-theme='dark'] .blend-form::-webkit-scrollbar-thumb { background: #2b3a4f; }

:root[data-theme='dark'] .hist-tile { border-color: #21262d; }
:root[data-theme='dark'] .hist-empty,
:root[data-theme='dark'] .readonly-row { background: var(--c-bg); border-color: var(--c-border); }

:root[data-theme='dark'] .reference-image-preview { background: var(--c-bg); border-color: var(--c-border); }

/* 弹窗自身的标题/页脚：全局暗色只改了 el-dialog，标题用的是 el-text-color-primary，
   这里显式对齐一次，避免残留深色字压深色底。 */
:root[data-theme='dark'] .cover-dialog .cover-dialog-title { color: #e6edf3; }
:root[data-theme='dark'] .cover-dialog {
  background: #161b22;
  border: 1px solid #21262d;

  /* 弹窗被 teleport 到 body，Element Plus 的暗色变量不会自己跟过来 ——
     实测里面 --el-bg-color 还是 #ffffff、--el-fill-color-blank 也是 #ffffff，
     于是"小说名称"这类 disabled 输入框、输入框描边全是浅色，在深底上非常刺眼。
     与其一个个组件去盖，不如把这一套变量就地换掉。 */
  --el-bg-color: #161b22;
  --el-bg-color-overlay: #161b22;
  --el-bg-color-page: #0d1117;
  --el-fill-color-blank: #1c2333;
  --el-fill-color: #1c2333;
  --el-fill-color-light: #1c2333;
  --el-fill-color-lighter: #192031;
  --el-border-color: rgba(59, 130, 246, .18);
  --el-border-color-light: rgba(59, 130, 246, .14);
  --el-border-color-lighter: rgba(59, 130, 246, .10);
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #e6edf3;
  --el-text-color-secondary: #8b98a8;
  --el-text-color-placeholder: #4d5768;
  --el-disabled-bg-color: #161b22;
  --el-disabled-text-color: #5f6b7a;
  --el-disabled-border-color: #21262d;
}

/* 输入框/下拉框的边框是 inset box-shadow（不是 border），
   全局暗色只改了 background 与 border-color，框线因此还是浅的。 */
:root[data-theme='dark'] .cover-dialog .el-input__wrapper,
:root[data-theme='dark'] .cover-dialog .el-select__wrapper,
:root[data-theme='dark'] .cover-dialog .el-textarea__inner {
  box-shadow: 0 0 0 1px var(--el-border-color) inset !important;
}

:root[data-theme='dark'] .cover-dialog .el-input.is-disabled .el-input__wrapper {
  box-shadow: 0 0 0 1px #21262d inset !important;
}
</style>

<style scoped>
.memory-body {
  min-height: 120px;
}
.memory-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 2px;
  border-bottom: 1px dashed var(--app-border, #e8ecf2);
  font-size: 13px;
}
.memory-row .m-count {
  font-weight: 700;
  color: var(--app-color, #1a2230);
}
.memory-meta {
  margin-top: 12px;
  font-size: 12px;
}
.memory-meta .m-warn {
  color: #e6a23c;
}
.memory-meta .m-pending {
  color: #909399;
}
.memory-meta .m-ok {
  color: #67c23a;
}
.memory-failures {
  margin-top: 12px;
  padding: 8px 10px;
  border-radius: 8px;
  background: rgba(230, 162, 60, 0.12);
  font-size: 12px;
}
.memory-failures .f-title {
  font-weight: 600;
  margin-bottom: 4px;
}
.memory-hint {
  margin: 14px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--app-muted, #7a8699);
}
.memory-hint code {
  padding: 1px 4px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.06);
}
</style>
