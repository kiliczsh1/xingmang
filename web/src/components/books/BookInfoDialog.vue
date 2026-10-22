<template>
  <el-dialog
    v-model="infoDialogVisible"
    width="460px"
    :align-center="true"
    :close-on-click-modal="true"
    append-to-body
    class="book-info-dialog"
    destroy-on-close
  >
    <!-- 报头 Header -->
    <template #header>
      <div class="press-header">
        <div class="press-header-top">
          <span>EST. 2026 · 文学编辑部</span>
          <span class="press-issue">№ {{ String(infoBook?.id || 0).padStart(3, '0') }}</span>
        </div>
        <div class="press-header-title">作品信息<span class="press-accent"> · DAILY</span></div>
      </div>
    </template>

    <div v-if="infoBook" class="press-body">
      <!-- 顶部：封面 + 标题输入 -->
      <div class="press-top-row">
        <div
          class="press-cover"
          @click="previewCover"
        >
          <img v-if="infoBook.cover" :src="infoBook.cover" class="press-cover-img" alt="封面" />
          <span v-else class="press-cover-text">{{ (infoBook.title || '书')[0] }}</span>
        </div>
        <div class="press-title-block">
          <label class="press-field-label">HEADLINE · 作品标题</label>
          <input
            class="press-title-input"
            v-model="infoEditTitle"
            placeholder="作品名称"
          />
        </div>
      </div>

      <!-- Metadata 版权栏 -->
      <div class="press-meta-grid">
        <div class="press-meta-item">
          <span class="press-k">分类</span>
          <span class="press-v">{{ infoBook.category || '未分类' }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">状态</span>
          <span class="press-v">{{ getStatusText(infoBook.status) }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">可见性</span>
          <span class="press-v">{{ infoBook.is_public ? '公开' : '私密' }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">作者</span>
          <span class="press-v">{{ infoBook.author || '未填写' }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">ID</span>
          <span class="press-v">#{{ infoBook.id }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">章节</span>
          <span class="press-v">{{ infoBook.chapter_count || 0 }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">创建</span>
          <span class="press-v">{{ formatDateFull(infoBook.created_at) }}</span>
        </div>
        <div class="press-meta-item">
          <span class="press-k">更新</span>
          <span class="press-v">{{ formatDateFull(infoBook.updated_at) }}</span>
        </div>
      </div>

      <!-- 作品简介（报纸正文） -->
      <div class="press-section">
        <div class="press-section-label"><span>SYNOPSIS · 作品简介</span></div>
        <textarea
          class="press-textarea"
          v-model="infoEditDesc"
          placeholder="作品简介（可选）"
          rows="4"
        ></textarea>
      </div>

      <!-- 目标字数 + 进度条（油墨条） -->
      <div class="press-progress-section">
        <div class="press-progress-head">
          <span class="press-progress-label">PROGRESS · 字数进度</span>
          <span class="press-pct">{{ bookProgress(infoBook) }}%</span>
        </div>
        <div class="press-progress-track">
          <div
            class="press-progress-fill"
            :style="{ width: bookProgress(infoBook) + '%' }"
          ></div>
        </div>
        <div class="press-word-counts">
          <span><b>{{ formatWordCount(infoBook.word_count || 0) }}</b> 字已写 / {{ formatWordCount(targetWordsOf(infoBook)) }} 字目标</span>
          <span>{{ infoBook.chapter_count || 0 }} 章</span>
        </div>
        <div class="press-target-edit">
          <span>TARGET ─</span>
          <input
            class="press-target-input"
            v-model="targetWordsInput"
            type="number"
            placeholder="目标字数"
            @keyup.enter="saveTargetWords"
          />
          <span>字</span>
          <button class="press-target-save" @click="saveTargetWords">设定</button>
        </div>
      </div>

      <!-- 标签 -->
      <div class="press-section">
        <div class="press-section-label"><span>TAGS · 标签</span></div>
        <div class="press-tags-row">
          <span
            v-for="tag in infoBook.tags || []"
            :key="tag"
            class="press-tag"
          >{{ tag }}</span>
          <span v-if="!infoBook.tags || infoBook.tags.length === 0" class="press-empty">暂无标签</span>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="press-footer">
        <span class="press-footer-note">— Editorial Dept. —</span>
        <button class="press-save-btn" @click="saveInfoEdit">保存</button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
// 作品信息弹窗（报刊编辑风格）
// 2026-09-21 第 3 期从 views/Books.vue 拆出；编辑态自持，书与开关由父组件传入，封面点击向上冒泡给大图预览
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useBookStore } from '@/stores/book'
import type { Book } from '@/types'
import { getStatusText, bookProgress, formatWordCount, targetWordsOf } from '@/utils/bookDisplay'

const props = defineProps<{ visible: boolean; book: Book | null }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void; (e: 'preview', book: Book): void }>()

const bookStore = useBookStore()

const infoBook = computed(() => props.book)
const infoDialogVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v)
})

const infoEditTitle = ref('')

const infoEditDesc = ref('')

const targetWordsInput = ref('')

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

const saveInfoEdit = async () => {
  if (!infoBook.value) return
  const title = infoEditTitle.value.trim()
  if (!title) {
    ElMessage.warning('请输入作品名称')
    return
  }
  const updated = await bookStore.updateBook(infoBook.value.id, {
    title,
    description: infoEditDesc.value
  })
  if (updated) {
    ElMessage.success('保存成功')
    infoDialogVisible.value = false
  } else {
    ElMessage.error('保存失败，请重试')
  }
}

const saveTargetWords = async () => {
  if (!infoBook.value) return
  const raw = targetWordsInput.value.trim()
  let tw: number | undefined
  if (raw) {
    tw = Math.floor(Number(raw))
    if (!isFinite(tw) || tw <= 0) {
      ElMessage.warning('请输入大于 0 的目标字数')
      return
    }
  }
  const updated = await bookStore.updateBook(infoBook.value.id, { target_words: tw ?? 0 })
  if (updated) {
    targetWordsInput.value = updated.target_words ? String(updated.target_words) : ''
    ElMessage.success(
      tw
        ? `目标字数已设置为 ${formatWordCount(tw)}`
        : '已清除自定义目标，进度按 10 万字计算'
    )
  } else {
    ElMessage.error('保存失败，请重试')
  }
}
// 封面点击 → 交给父组件打开大图预览
const previewCover = () => {
  if (infoBook.value?.cover) emit('preview', infoBook.value)
}

// 打开时用传入的书初始化编辑态（原 handleShowInfo 的初始化段）
watch(() => props.visible, (v) => {
  if (!v || !props.book) return
  infoEditTitle.value = props.book.title || ''
  infoEditDesc.value = props.book.description || ''
  targetWordsInput.value = props.book.target_words ? String(props.book.target_words) : ''
})
</script>

<style scoped>
/* ========== 作品信息弹窗 · 报刊编辑风格 ========== */
.book-info-dialog :deep(.el-dialog) {
  width: 460px !important;
  max-width: calc(100vw - 32px);
  margin: 0 auto !important;
  border-radius: 2px;
  overflow: hidden;
  background-color: #f4f1ea;
  background-image:
    radial-gradient(circle at 25% 25%, rgba(0,0,0,0.015) 1px, transparent 1px),
    radial-gradient(circle at 75% 75%, rgba(0,0,0,0.015) 1px, transparent 1px);
  background-size: 4px 4px, 6px 6px;
  border: 1px solid #c8c4ba;
  box-shadow: 0 6px 24px rgba(0, 0, 0, 0.3);
  font-family: 'Lora', 'Noto Serif SC', Georgia, serif;
}

/* 报头 Header */
.book-info-dialog :deep(.el-dialog__header) {
  padding: 14px 22px 12px;
  border-bottom: 3px double #1a1a1a;
  margin-right: 0;
  position: relative;
}
.book-info-dialog :deep(.el-dialog__header)::after {
  content: '';
  position: absolute;
  bottom: -7px;
  left: 22px;
  right: 22px;
  height: 1px;
  background: #1a1a1a;
}
.book-info-dialog :deep(.el-dialog__headerbtn) {
  top: 16px;
  right: 18px;
  width: 28px;
  height: 28px;
  border: 1px solid #1a1a1a;
  border-radius: 2px;
  background: transparent;
  font-size: 16px;
  transition: all 0.15s;
}
.book-info-dialog :deep(.el-dialog__headerbtn:hover) {
  background: #c1272d;
  border-color: #c1272d;
}
.book-info-dialog :deep(.el-dialog__headerbtn .el-dialog__close) {
  color: #1a1a1a;
  font-family: 'Lora', serif;
}
.book-info-dialog :deep(.el-dialog__headerbtn:hover .el-dialog__close) {
  color: #f4f1ea;
}

/* Body */
.book-info-dialog :deep(.el-dialog__body) {
  padding: 16px 22px 18px;
  max-height: 60vh;
  overflow-y: auto;
}
.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar {
  width: 8px;
}
.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-track {
  background: #ece7da;
}
.book-info-dialog :deep(.el-dialog__body)::-webkit-scrollbar-thumb {
  background: #6b6b6b;
  border: 2px solid #ece7da;
}

/* Footer */
.book-info-dialog :deep(.el-dialog__footer) {
  padding: 14px 22px;
  border-top: 3px double #1a1a1a;
  background: #ece7da;
}

.book-info-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ---- 报头内容 ---- */
.press-header {
  width: 100%;
}
.press-header-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 10px;
  letter-spacing: 0.15em;
  color: #6b6b6b;
  text-transform: uppercase;
  margin-bottom: 6px;
  font-family: 'Lora', serif;
}
.press-issue {
  color: #c1272d;
  font-weight: 600;
}
.press-header-title {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 26px;
  font-weight: 900;
  letter-spacing: 0.04em;
  color: #1a1a1a;
  line-height: 1;
}
.press-accent {
  color: #c1272d;
}

/* ---- Body 内容 ---- */
.press-body {
  color: #1a1a1a;
}

/* 顶部：封面 + 标题 */
.press-top-row {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 16px;
  padding-bottom: 14px;
  border-bottom: 1px solid #c8c4ba;
}

/* 封面：原图彩色直出（旧版加了 grayscale 黑白处理，封面看起来是灰白的） */
.press-cover {
  width: 64px;
  height: 80px;
  background: #1a1a1a;
  border: 1px solid #1a1a1a;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}
.press-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.press-cover-text {
  color: #f4f1ea;
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 28px;
  font-weight: 900;
}

/* 标题输入框：头条标题 */
.press-title-block {
  flex: 1;
  min-width: 0;
}
.press-field-label {
  display: block;
  font-family: 'Lora', serif;
  font-size: 9px;
  font-weight: 600;
  letter-spacing: 0.2em;
  color: #6b6b6b;
  margin-bottom: 4px;
  text-transform: uppercase;
}
.press-title-input {
  width: 100%;
  background: transparent;
  border: none;
  border-bottom: 2px solid #1a1a1a;
  color: #1a1a1a;
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 22px;
  font-weight: 900;
  line-height: 1.2;
  padding: 4px 0 6px;
  outline: none;
  transition: border-color 0.15s;
}
.press-title-input:focus {
  border-bottom-color: #c1272d;
}
.press-title-input::placeholder {
  color: #c8c4ba;
  font-weight: 400;
}

/* ---- Metadata 版权栏 ---- */
.press-meta-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0;
  margin-bottom: 16px;
  padding: 10px 0;
  border-bottom: 1px solid #c8c4ba;
  font-family: 'Lora', serif;
}
.press-meta-item {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 11px;
  padding: 3px 0;
  border-bottom: 1px dotted #c8c4ba;
}
.press-meta-item:nth-last-child(-n+2) {
  border-bottom: none;
}
.press-meta-item:nth-child(odd) {
  padding-right: 10px;
}
.press-meta-item:nth-child(even) {
  padding-left: 10px;
  border-left: 1px dotted #c8c4ba;
}
.press-k {
  color: #6b6b6b;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-size: 9px;
  flex-shrink: 0;
}
.press-v {
  color: #1a1a1a;
  font-weight: 600;
  margin-left: auto;
  text-align: right;
}

/* ---- 章节标题 ---- */
.press-section {
  margin-top: 14px;
}
.press-section-label {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: #c1272d;
  text-transform: uppercase;
  padding-bottom: 4px;
  margin-bottom: 8px;
  border-bottom: 1px solid #1a1a1a;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.press-section-label::after {
  content: '◆';
  color: #1a1a1a;
  font-size: 8px;
}

/* ---- 简介正文 ---- */
.press-textarea {
  width: 100%;
  background: transparent;
  border: none;
  border-top: 1px solid transparent;
  color: #1a1a1a;
  font-family: 'Lora', 'Noto Serif SC', serif;
  font-size: 13px;
  line-height: 1.8;
  padding: 8px 0;
  outline: none;
  resize: vertical;
  min-height: 80px;
  text-align: justify;
  transition: border-color 0.15s;
}
.press-textarea:focus {
  border-top-color: #c1272d;
}
.press-textarea::placeholder {
  color: #c8c4ba;
  font-style: italic;
}

/* ---- 目标字数 + 进度条 ---- */
.press-progress-section {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #c8c4ba;
}
.press-progress-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 10px;
}
.press-progress-label {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.2em;
  color: #1a1a1a;
  text-transform: uppercase;
}
.press-pct {
  font-family: 'Playfair Display', serif;
  font-size: 22px;
  font-weight: 900;
  color: #c1272d;
  line-height: 1;
}

/* 进度条：油墨条 */
.press-progress-track {
  height: 10px;
  background: #e6e1d6;
  border: 1px solid #1a1a1a;
  margin-bottom: 8px;
  position: relative;
  overflow: hidden;
}
.press-progress-fill {
  height: 100%;
  background: repeating-linear-gradient(
    90deg,
    #1a1a1a 0, #1a1a1a 2px,
    #2a2a2a 2px, #2a2a2a 4px
  );
  position: relative;
  transition: width 0.4s ease;
}
.press-progress-fill::after {
  content: '';
  position: absolute;
  right: 0;
  top: 0;
  width: 2px;
  height: 100%;
  background: #c1272d;
}

.press-word-counts {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-family: 'Lora', serif;
  font-size: 11px;
  color: #6b6b6b;
  margin-bottom: 10px;
  font-style: italic;
}
.press-word-counts b {
  color: #1a1a1a;
  font-weight: 700;
  font-style: normal;
}

/* 目标字数编辑 */
.press-target-edit {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 8px;
  border-top: 1px dotted #c8c4ba;
  font-family: 'Lora', serif;
  font-size: 10px;
  letter-spacing: 0.1em;
  color: #6b6b6b;
  text-transform: uppercase;
}
.press-target-input {
  width: 100px;
  background: transparent;
  border: none;
  border-bottom: 1px solid #1a1a1a;
  color: #1a1a1a;
  font-family: 'Playfair Display', serif;
  font-size: 14px;
  font-weight: 700;
  padding: 2px 0;
  outline: none;
  text-align: right;
  transition: border-color 0.15s;
}
.press-target-input:focus {
  border-bottom-color: #c1272d;
}
.press-target-input::placeholder {
  color: #c8c4ba;
  font-weight: 400;
}
.press-target-save {
  background: transparent;
  border: 1px solid #1a1a1a;
  color: #1a1a1a;
  font-family: 'Lora', serif;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  padding: 2px 10px;
  cursor: pointer;
  border-radius: 2px;
  transition: all 0.15s;
}
.press-target-save:hover {
  background: #1a1a1a;
  color: #f4f1ea;
}

/* ---- 标签 ---- */
.press-tags-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
}
.press-tag {
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-size: 12px;
  font-weight: 700;
  color: #1a1a1a;
  letter-spacing: 0.04em;
}
.press-tag::before {
  content: '[';
  color: #c1272d;
  margin-right: 1px;
}
.press-tag::after {
  content: ']';
  color: #c1272d;
  margin-left: 1px;
}
.press-empty {
  color: #c8c4ba;
  font-size: 12px;
  font-style: italic;
  font-family: 'Lora', serif;
}

/* ---- Footer ---- */
.press-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.press-footer-note {
  font-family: 'Lora', serif;
  font-size: 10px;
  font-style: italic;
  color: #6b6b6b;
  letter-spacing: 0.05em;
}
.press-save-btn {
  background: #1a1a1a;
  color: #f4f1ea;
  border: 1px solid #1a1a1a;
  font-family: 'Playfair Display', 'Noto Serif SC', serif;
  font-weight: 700;
  font-size: 13px;
  letter-spacing: 0.25em;
  text-transform: uppercase;
  padding: 10px 28px;
  cursor: pointer;
  border-radius: 2px;
  transition: all 0.15s;
}
.press-save-btn::before {
  content: '▶ ';
  color: #c1272d;
  font-size: 9px;
}
.press-save-btn:hover {
  background: #c1272d;
  border-color: #c1272d;
}

/* ---- 响应式 ---- */
@media (max-width: 600px) {
  .book-info-dialog :deep(.el-dialog) {
    width: calc(100vw - 32px) !important;
  }
  .press-top-row {
    flex-direction: column;
    align-items: flex-start;
  }
  .press-meta-grid {
    grid-template-columns: 1fr;
  }
  .press-meta-item:nth-child(even) {
    padding-left: 0;
    border-left: none;
  }
}

/* ---- 作品信息弹窗（报刊编辑风）· 暗色适配 ----
   弹窗外壳底色由全局 styles/dark-theme.css 接管（.el-dialog → #161b22，带 !important），
   而报刊风的墨色是写死的字面量 → 暗色下正文会与底色撞成隐形。这里统一换成浅墨。 */
/* 正文 / 标题 / 数值 */
:root[data-theme='dark'] .press-body,
:root[data-theme='dark'] .press-header-title,
:root[data-theme='dark'] .press-v,
:root[data-theme='dark'] .press-progress-label,
:root[data-theme='dark'] .press-word-counts b,
:root[data-theme='dark'] .press-tag {
  color: #e6edf3;
}
/* 次级说明文字 */
:root[data-theme='dark'] .press-header-top,
:root[data-theme='dark'] .press-field-label,
:root[data-theme='dark'] .press-k,
:root[data-theme='dark'] .press-word-counts,
:root[data-theme='dark'] .press-target-edit,
:root[data-theme='dark'] .press-footer-note,
:root[data-theme='dark'] .press-empty {
  color: #8b98a8;
}
/* 报刊红：暗底上需提亮才有足够对比度 */
:root[data-theme='dark'] .press-issue,
:root[data-theme='dark'] .press-accent,
:root[data-theme='dark'] .press-section-label,
:root[data-theme='dark'] .press-pct,
:root[data-theme='dark'] .press-tag::before,
:root[data-theme='dark'] .press-tag::after {
  color: #ff7b72;
}
/* 分隔线 / 点线 */
:root[data-theme='dark'] .press-top-row,
:root[data-theme='dark'] .press-meta-grid,
:root[data-theme='dark'] .press-section-label {
  border-bottom-color: #30363d;
}
:root[data-theme='dark'] .press-meta-item {
  border-bottom-color: #30363d;
}
:root[data-theme='dark'] .press-meta-item:nth-child(even) {
  border-left-color: #30363d;
}
:root[data-theme='dark'] .press-progress-section,
:root[data-theme='dark'] .press-target-edit {
  border-top-color: #30363d;
}
:root[data-theme='dark'] .press-section-label::after {
  color: #30363d;
}
/* 封面：深底浅字 */
:root[data-theme='dark'] .press-cover {
  background: #21262d;
  border-color: #30363d;
}
:root[data-theme='dark'] .press-cover-text {
  color: #e6edf3;
}
/* 输入控件（标题 / 简介 / 目标字数 / 设定按钮） */
:root[data-theme='dark'] .press-title-input,
:root[data-theme='dark'] .press-target-input {
  color: #e6edf3;
  border-bottom-color: #30363d;
}
:root[data-theme='dark'] .press-textarea {
  color: #e6edf3;
}
:root[data-theme='dark'] .press-title-input::placeholder,
:root[data-theme='dark'] .press-textarea::placeholder,
:root[data-theme='dark'] .press-target-input::placeholder {
  color: #8b98a8;
}
:root[data-theme='dark'] .press-title-input:focus,
:root[data-theme='dark'] .press-target-input:focus {
  border-bottom-color: #ff7b72;
}
:root[data-theme='dark'] .press-textarea:focus {
  border-top-color: #ff7b72;
}
:root[data-theme='dark'] .press-target-save {
  color: #e6edf3;
  border-color: #30363d;
}
:root[data-theme='dark'] .press-target-save:hover {
  background: #30363d;
  color: #e6edf3;
}
/* 字数进度条：油墨条改浅墨条纹 */
:root[data-theme='dark'] .press-progress-track {
  background: #21262d;
  border-color: #30363d;
}
:root[data-theme='dark'] .press-progress-fill {
  background: repeating-linear-gradient(
    90deg,
    #ff7b72 0, #ff7b72 2px,
    #ffa39e 2px, #ffa39e 4px
  );
}
/* 页脚保存按钮 */
:root[data-theme='dark'] .press-save-btn {
  background: #21262d;
  border-color: #30363d;
}
:root[data-theme='dark'] .press-save-btn:hover {
  background: #c1272d;
  border-color: #c1272d;
}
</style>
