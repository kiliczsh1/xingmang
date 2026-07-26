<template>
  <!-- 导入章节弹窗 -->
  <el-dialog
    v-model="visible"
    title="导入章节"
    width="580px"
    class="import-chapter-dialog"
    append-to-body
    :close-on-click-modal="false"
    @close="resetImportChapterDialog"
  >
    <div class="import-chapter-content">
      <!-- 上传区域 -->
      <div
        v-if="!importChapterPreview.length"
        class="import-chapter-upload"
        @click="importChapterFileInput?.click()"
      >
        <input
          ref="importChapterFileInput"
          type="file"
          accept=".txt,.docx"
          class="hidden-input"
          @change="handleImportChapterFileChange"
        />
        <div class="upload-icon-wrap">
          <el-icon class="upload-main-icon"><UploadFilled /></el-icon>
        </div>
        <p class="upload-text">点击或拖拽文件到此处上传</p>
        <p class="upload-hint">支持 txt、docx 格式，单文件不超过 20MB</p>
      </div>

      <!-- 章节预览区域 -->
      <template v-else>
        <div class="import-chapter-file-bar">
          <span class="file-name">
            <el-icon><Document /></el-icon>
            {{ importChapterFileName }}
          </span>
          <el-button size="small" text type="primary" @click="reselectChapterFile">
            重新选择
          </el-button>
        </div>
        <div class="import-chapter-stats">
          共识别出 <strong>{{ importChapterPreview.length }}</strong> 个章节
        </div>
        <div class="import-chapter-select-all">
          <el-checkbox v-model="importChapterSelectAll" @change="handleImportChapterSelectAll">
            全选 ({{ importChapterSelected.size }}/{{ importChapterPreview.length }})
          </el-checkbox>
        </div>
        <div class="import-chapter-list">
          <div
            v-for="(chapter, index) in importChapterPreview"
            :key="index"
            class="import-chapter-item"
            :class="{ selected: importChapterSelected.has(index) }"
            @click="toggleImportChapterItem(index)"
          >
            <el-checkbox
              :model-value="importChapterSelected.has(index)"
              @click.stop
              @change="toggleImportChapterItem(index)"
            />
            <div class="import-chapter-info">
              <div class="import-chapter-title">{{ chapter.title }}</div>
              <div class="import-chapter-preview">{{ chapter.content.slice(0, 70) }}{{ chapter.content.length > 70 ? '...' : '' }}</div>
            </div>
            <span class="import-chapter-len">{{ chapter.content.length }}字</span>
          </div>
        </div>
      </template>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        v-if="importChapterPreview.length"
        type="primary"
        :disabled="importChapterSelected.size === 0"
        @click="confirmImportChapters"
      >
        导入 ({{ importChapterSelected.size }})
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, UploadFilled } from '@element-plus/icons-vue'
import { chapterAPI } from '@/api'
import type { Chapter } from '@/types'

const props = defineProps<{
  bookId: number
}>()

const visible = defineModel<boolean>('visible')
const emit = defineEmits<{
  imported: []
}>()

const importChapterFileInput = ref<HTMLInputElement | null>(null)
const importChapterFileName = ref('')
const importChapterPreview = ref<Array<Pick<Chapter, 'title' | 'content'>>>([])
const importChapterSelected = ref<Set<number>>(new Set())
const importChapterSelectAll = ref(false)

/** 关闭弹窗时重置 */
const resetImportChapterDialog = () => {
  importChapterFileName.value = ''
  importChapterPreview.value = []
  importChapterSelected.value = new Set()
  importChapterSelectAll.value = false
}

/** 选择文件后解析 */
const handleImportChapterFileChange = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  const ext = file.name.split('.').pop()?.toLowerCase()
  if (ext !== 'txt' && ext !== 'docx') {
    ElMessage.warning('仅支持 txt 和 docx 文件')
    input.value = ''
    return
  }
  if (file.size > 20 * 1024 * 1024) {
    ElMessage.warning('文件大小不能超过 20MB')
    input.value = ''
    return
  }

  try {
    let textContent = ''
    if (ext === 'txt') {
      // 尝试多种编码读取：UTF-8 → GBK → GB18030 → Big5
      const arrayBuffer = await file.arrayBuffer()
      textContent = await decodeTextWithFallback(arrayBuffer)
    } else {
      // docx: 发送后端解析
      const reader = new FileReader()
      reader.onload = async () => {
        const data_base64 = (reader.result as string).split(',')[1]
        const res = await chapterAPI.importFile({
          bookId: props.bookId,
          file: { name: file.name, size: file.size, data_base64 }
        })
        if (res.success) {
          ElMessage.success(`成功导入 ${res.data?.insertedCount ?? 0} 个章节`)
          visible.value = false
          emit('imported')
        }
        input.value = ''
      }
      reader.readAsDataURL(file)
      return
    }

    // txt: 前端智能拆分章节
    const chapters = parseChaptersFromText(textContent, file.name)
    if (chapters.length === 0) {
      ElMessage.warning('文件内容为空或无法识别章节格式')
      input.value = ''
      return
    }
    importChapterFileName.value = file.name
    importChapterPreview.value = chapters
    importChapterSelected.value = new Set(chapters.map((_, i) => i))
    importChapterSelectAll.value = true
  } catch {
    ElMessage.error('读取文件失败')
  }
  input.value = ''
}

/**
 * 多编码解码：依次尝试 UTF-8 → GBK → GB18030 → Big5
 * 大部分中文小说 txt 是 GBK 编码
 */
const decodeTextWithFallback = async (buffer: ArrayBuffer): Promise<string> => {
  const encodings = ['utf-8', 'gbk', 'gb18030', 'big5']

  for (const enc of encodings) {
    try {
      const decoder = new TextDecoder(enc, { fatal: false })
      const text = decoder.decode(buffer)
      // 检查是否为有效中文内容（排除乱码）
      if (isLikelyValidChineseText(text)) {
        return text
      }
    } catch {
      // 该编码不支持，继续尝试下一个
    }
  }
  // 全部失败时回退到 UTF-8（不抛错模式）
  return new TextDecoder('utf-8', { fatal: false }).decode(buffer)
}

/** 粗略判断文本是否像正常中文内容 */
const isLikelyValidChineseText = (text: string): boolean => {
  if (!text || text.length < 10) return false
  // 统计中文字符占比
  const chineseChars = text.match(/[\u4e00-\u9fff\u3400-\u4dbf]/g)
  if (chineseChars && chineseChars.length > text.length * 0.15) return true
  // 如果是纯 ASCII 英文也接受
  if (/^[\x20-\x7e\r\n\t]+$/.test(text.slice(0, 500))) return true
  // 检查是否有大量连续乱码特征（如连续的 � 或不可见字符）
  const garbageMatch = text.match(/[�□■◆◇●○▲▼▽△▷◁♦♠♣♥]/g)
  if (garbageMatch && garbageMatch.length > text.length * 0.1) return false
  return chineseChars !== null && chineseChars.length >= 5
}

/** 重新选择文件 */
const reselectChapterFile = () => {
  importChapterFileInput.value?.click()
}

/**
 * 智能识别小说章节标题并拆分
 * 支持的格式：
 *   第N章 / 第一章 / 第1章 / Chapter 1 / Ch.1
 *   一、xxx / 1.xxx / 【第N章】xxx
 *   卷一 第一章（卷+章组合）
 */
const parseChaptersFromText = (text: string, fileName: string) => {
  // 辅助函数：从模式列表中收集匹配
  const collectMatches = (
    patternList: RegExp[],
    existing: Array<{ title: string; index: number }> = []
  ): Array<{ title: string; index: number }> => {
    const result = [...existing]
    for (const pattern of patternList) {
      for (const m of text.matchAll(pattern)) {
        const title = m[0].trim()
        if (title.length >= 2 && title.length <= 50 && m.index !== undefined) {
          const dup = result.find(
            existing => Math.abs(existing.index - m.index!) < 5
          )
          if (!dup) {
            result.push({ title, index: m.index! })
          }
        }
      }
    }
    return result
  }

  // 第一层：可靠模式 — 明确包含"章/节/回/卷"等关键词的模式
  const reliablePatterns: RegExp[] = [
    // 卷X 第X章 组合（优先级最高）
    /(?:卷[一二三四五六七八九十百千零\d]+[\s：:\-_]*第[一二三四五六七八九十百千零\d]+[章节回卷集部篇][^\n]*)/g,
    // 第N章/节/回 等 — 中文数字或阿拉伯数字
    /(?:第[一二三四五六七八九十百千零\d]+[章节回卷集部篇][^\n]*)/g,
    // 第1章 / 第2章 ... 阿拉伯数字
    /^第\d+[章节回卷集部篇][^\n]*/gm,
    // Chapter 1 / Ch.1 / CHAPTER ONE
    /^(?:Chapter|CHAPTER|Ch\.?)\s*\d+[^\n]*/gm,
    // 序言/前言/后记/楔子/引子 等特殊章节名
    /^(?:序言|前言|后记|楔子|引子|尾声|番外|外传|附录|终章)[^\n]{0,20}$/gm,
  ]

  const matches = collectMatches(reliablePatterns)

  // 第二层：模糊模式 — 仅当可靠模式未找到足够章节时才启用
  // 避免把正文中的"1、xxx"或"一、xxx"误识别为章节标题
  if (matches.length < 2) {
    const fuzzyPatterns: RegExp[] = [
      // 中文数字开头：一、xxx / 二、xxx
      /^[一二三四五六七八九十百千]+[、\.．\s:：\-_]*(.+)$/gm,
      // 数字编号：1. xxx / 1、xxx
      /^\d+[、\.．\s:：\-_]{1,2}.+/gm,
    ]
    const fuzzyMatches = collectMatches(fuzzyPatterns, matches)
    matches.length = 0
    matches.push(...fuzzyMatches)
  }

  // 启发式检测：短行+空行=标题，仅当所有模式都没找到足够章节时使用
  if (matches.length < 3) {
    const lines = text.split('\n')
    let pos = 0
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim()
      // 短行（2-35字）、非纯数字、非空、下一行为空或很短 → 可能是章节标题
      const isShortTitle =
        line.length >= 2 && line.length <= 35 &&
        !/^\d+$/.test(line) &&
        !line.startsWith('http') &&
        !line.startsWith('www') &&
        (i + 1 >= lines.length || lines[i + 1].trim().length <= 3)

      if (isShortTitle) {
        const dup = matches.find(
          existing => Math.abs(existing.index - pos) < line.length + 5
        )
        if (!dup) {
          matches.push({ title: line, index: pos })
        }
      }
      pos += lines[i].length + 1
    }
  }

  // 按位置排序，去重（合并过于接近的匹配）
  matches.sort((a, b) => a.index - b.index)
  const deduped: typeof matches = []
  for (const m of matches) {
    if (deduped.length === 0 || m.index - deduped[deduped.length - 1].index > 10) {
      deduped.push(m)
    }
  }

  // 拆分章节内容
  const chapters: Array<Pick<Chapter, 'title' | 'content'>> = []
  if (deduped.length > 0) {
    for (let i = 0; i < deduped.length; i++) {
      const startIdx = deduped[i].index
      const endIdx = i < deduped.length - 1 ? deduped[i + 1].index : text.length
      const content = text.slice(startIdx, endIdx).trim()
      if (content.length > 15) { // 至少15个字才算有效章节
        chapters.push({
          title: deduped[i].title,
          content
        })
      }
    }
  }

  // 如果识别到的章节太少且文本很长，补充按段落拆分
  if (chapters.length < 2 && text.trim().length > 1000) {
    chapters.length = 0 // 清空不够好的结果
    const segments = text.split(/\n\s*\n+/)
    let buffer = ''
    for (const seg of segments) {
      const trimmed = seg.trim()
      if (!trimmed) continue
      buffer += trimmed + '\n\n'
      if (buffer.length >= 1500) {
        chapters.push({
          title: `段落 ${chapters.length + 1}`,
          content: buffer.trim()
        })
        buffer = ''
      }
    }
    if (buffer.trim().length > 200) {
      chapters.push({ title: `段落 ${chapters.length + 1}`, content: buffer.trim() })
    }
  }

  // 最终兜底：整个文件作为一章
  if (chapters.length === 0) {
    chapters.push({ title: fileName.replace(/\.[^.]+$/, ''), content: text.trim() })
  }

  return chapters.filter(c => c.content && c.content.length > 5)
}

const handleImportChapterSelectAll = (val: boolean) => {
  if (val) {
    importChapterSelected.value = new Set(importChapterPreview.value.map((_, i) => i))
  } else {
    importChapterSelected.value.clear()
  }
}

const toggleImportChapterItem = (index: number) => {
  if (importChapterSelected.value.has(index)) {
    importChapterSelected.value.delete(index)
  } else {
    importChapterSelected.value.add(index)
  }
  importChapterSelected.value = new Set(importChapterSelected.value)
  importChapterSelectAll.value = importChapterSelected.value.size === importChapterPreview.value.length
}

const confirmImportChapters = async () => {
  if (importChapterSelected.value.size === 0) {
    ElMessage.warning('请至少选择一个章节')
    return
  }
  const selected = importChapterPreview.value.filter((_, i) => importChapterSelected.value.has(i))
  try {
    const res = await chapterAPI.importBook({ bookId: props.bookId, chapters: selected })
    if (res.success) {
      ElMessage.success(`成功导入 ${res.data!.insertedCount} 个章节`)
      visible.value = false
      emit('imported')
    }
  } catch {
    ElMessage.error('导入失败')
  }
}
</script>

<style scoped>
.hidden-input {
  display: none;
}

/* ========== 导入章节弹窗 ========== */
.import-chapter-content {
  min-height: 60px;
}

/* 上传区域 */
.import-chapter-upload {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 24px;
  border: 2px dashed var(--ds-border-default, #d9d9d9);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  background: rgba(16, 185, 129, 0.02);
}

.import-chapter-upload:hover {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.06);
}

.upload-icon-wrap {
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #10b981, #059669);
  border-radius: 14px;
  margin-bottom: 14px;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.25);
}

.upload-main-icon {
  font-size: 28px;
  color: #fff;
}

.upload-text {
  font-size: 15px;
  font-weight: 600;
  color: var(--ds-text-primary, #333);
  margin: 0 0 6px;
}

.upload-hint {
  font-size: 12.5px;
  color: var(--ds-text-tertiary, #999);
  margin: 0;
}

/* 文件信息栏 */
.import-chapter-file-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  background: rgba(16, 185, 129, 0.06);
  border-radius: 8px;
  margin-bottom: 12px;
}

.file-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 500;
  color: var(--ds-text-primary, #262626);
}

.import-chapter-stats {
  font-size: 13px;
  color: var(--ds-text-secondary, #666);
  padding: 4px 2px 10px;
}

.import-chapter-select-all {
  margin-bottom: 10px;
}

.import-chapter-list {
  max-height: 340px;
  overflow-y: auto;
  padding-right: 4px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.import-chapter-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 12px 14px;
  border: 1.5px solid var(--ds-border-default, #e8e8e8);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  background: #fff;
}

.import-chapter-item:hover {
  border-color: #10b981;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.1);
}

.import-chapter-item.selected {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.04);
}

.import-chapter-info {
  flex: 1;
  min-width: 0;
}

.import-chapter-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--ds-text-primary, #262626);
  margin-bottom: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.import-chapter-preview {
  font-size: 11.5px;
  color: var(--ds-text-tertiary, #999);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.import-chapter-len {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--ds-text-tertiary, #bbb);
  white-space: nowrap;
  margin-top: 2px;
}

/* 暗色主题 */
:root[data-theme='dark'] .import-chapter-upload {
  border-color: rgba(255, 255, 255, 0.12);
  background: rgba(52, 211, 153, 0.03);
}

:root[data-theme='dark'] .import-chapter-upload:hover {
  border-color: #34d399;
  background: rgba(52, 211, 153, 0.06);
}

:root[data-theme='dark'] .upload-text {
  color: rgba(255, 255, 255, 0.88);
}

:root[data-theme='dark'] .upload-hint {
  color: rgba(255, 255, 255, 0.38);
}

:root[data-theme='dark'] .import-chapter-file-bar {
  background: rgba(52, 211, 153, 0.06);
}

:root[data-theme='dark'] .file-name {
  color: rgba(255, 255, 255, 0.88);
}

:root[data-theme='dark'] .import-chapter-stats {
  color: rgba(255, 255, 255, 0.45);
}

:root[data-theme='dark'] .import-chapter-item {
  border-color: rgba(255, 255, 255, 0.08);
  background: transparent;
}

:root[data-theme='dark'] .import-chapter-item:hover {
  border-color: #34d399;
  box-shadow: 0 2px 8px rgba(52, 211, 153, 0.1);
}

:root[data-theme='dark'] .import-chapter-item.selected {
  border-color: #34d399;
  background: rgba(52, 211, 153, 0.06);
}

:root[data-theme='dark'] .import-chapter-title {
  color: rgba(255, 255, 255, 0.88);
}

:root[data-theme='dark'] .import-chapter-preview {
  color: rgba(255, 255, 255, 0.38);
}

:root[data-theme='dark'] .import-chapter-len {
  color: rgba(255, 255, 255, 0.25);
}
</style>
