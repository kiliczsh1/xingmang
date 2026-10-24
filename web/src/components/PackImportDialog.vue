<template>
  <el-dialog
    v-model="visible"
    :title="`导入到「${packName}」`"
    width="760px"
    top="8vh"
    :close-on-click-modal="false"
    class="pack-import-dialog"
    append-to-body
    destroy-on-close
    @closed="resetAll"
  >
    <div class="pack-import-body">
      <div class="pack-import-target">
        <el-icon><FolderOpened /></el-icon>
        <span>导入的卡片统一归入当前卡包「{{ packName }}」，不会进入其他卡包，也不会变成独立卡片。</span>
      </div>

      <el-radio-group v-model="mode" class="pack-import-modes" @change="handleModeChange">
        <el-radio-button value="pack">卡包文件</el-radio-button>
        <el-radio-button value="cards">卡片文件</el-radio-button>
        <el-radio-button value="convert">txt / md 转换</el-radio-button>
      </el-radio-group>

      <el-upload
        ref="uploadRef"
        drag
        :multiple="mode === 'convert'"
        :accept="accept"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleFilesChange"
        class="pack-import-upload"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
        <template #tip>
          <div class="el-upload__tip">{{ uploadTip }}</div>
        </template>
      </el-upload>

      <div v-if="preview.length > 0" class="pack-import-preview">
        <div class="pack-import-preview-head">
          <span>预览（{{ preview.length }} 张卡片）</span>
          <div class="pack-import-preview-actions">
            <el-button size="small" text @click="selectAllCards">全选</el-button>
            <el-button size="small" text @click="clearSelection">全不选</el-button>
          </div>
        </div>
        <div class="pack-import-list">
          <div
            v-for="(card, index) in preview"
            :key="index"
            class="pack-import-item"
            :class="{ selected: selected.includes(index) }"
            @click="toggleCard(index)"
          >
            <el-checkbox :model-value="selected.includes(index)" @click.stop @change="toggleCard(index)" />
            <div class="pack-import-item-body">
              <div class="pack-import-item-name">
                {{ card.name }}
                <el-tag v-if="card.encrypted" size="small" type="warning">加密</el-tag>
              </div>
              <div class="pack-import-item-preview">{{ card.preview }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :disabled="selected.length === 0" :loading="importing" @click="executeImport">
        导入 ({{ selected.length }})
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { FolderOpened, UploadFilled } from '@element-plus/icons-vue'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import { ENCRYPT_MARKER } from '@/utils/promptCrypto'
import { buildPromptKey } from '@/utils/promptDedupMerge'
import {
  decodeExportData,
  flushRestoredRatingNotes,
  registerImportedInsights,
  registerImportedRatingNotes,
  resetImportedRatingNotes,
  restoreInsightsFor,
  restoreRatingNotesFor,
} from '@/utils/promptTransfer'

// 卡包内导入：与快捷中心的导入同一套解析 / 防重 / 心得搬运机制，
// 唯一差别是目标分类被钉死为当前卡包，文件里原有的 category / 卡包名一律忽略。
type ImportMode = 'pack' | 'cards' | 'convert'

interface PreviewCard {
  name: string
  preview: string
  encrypted: boolean
  payload: Record<string, any>
}

const props = defineProps<{
  packName: string
  prompts: Prompt[]
}>()

const emit = defineEmits<{
  (e: 'imported'): void
}>()

const visible = ref(false)
const mode = ref<ImportMode>('pack')
const preview = ref<PreviewCard[]>([])
const selected = ref<number[]>([])
const importing = ref(false)
const uploadRef = ref<any>()

const accept = computed(() => (mode.value === 'convert' ? '.txt,.md' : '.json'))

const uploadTip = computed(() => {
  if (mode.value === 'convert') return '支持批量上传 txt / md 文件，文件名即卡片名称'
  if (mode.value === 'pack') return '支持卡包导出文件（普通 / 拆书卡包均可），卡片全部落到当前卡包'
  return '支持独立卡片 / 单张卡片导出文件，卡片全部落到当前卡包'
})

const resetPreview = () => {
  preview.value = []
  selected.value = []
  uploadRef.value?.clearFiles?.()
}

const resetAll = () => {
  resetPreview()
  mode.value = 'pack'
  importing.value = false
}

const open = () => {
  resetImportedRatingNotes()
  resetAll()
  visible.value = true
}

const handleModeChange = () => {
  resetPreview()
}

const readText = (file: File) =>
  new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => resolve(String(e.target?.result || ''))
    reader.onerror = () => reject(new Error('文件读取失败'))
    reader.readAsText(file)
  })

const isValidCard = (card: any) =>
  !!card && typeof card === 'object' && String(card.name || '').trim() && card.content !== undefined && card.content !== null && String(card.content).length > 0

const collectCards = (list: any) => (Array.isArray(list) ? list.filter(isValidCard) : [])

// 兼容快捷中心各入口能读的文件格式：卡包 / 拆书多卡包 / 独立卡片 / 单张卡片 / 裸数组
const extractCards = (data: any): any[] => {
  if (Array.isArray(data)) {
    const list = collectCards(data)
    if (list.length > 0) return list
  }
  if (data && typeof data === 'object') {
    const byPrompts = collectCards(data.prompts)
    if (byPrompts.length > 0) return byPrompts
    if (Array.isArray(data.packs)) {
      const flat = data.packs.flatMap((pack: any) => collectCards(pack?.prompts))
      if (flat.length > 0) return flat
    }
    if (data.prompt && isValidCard(data.prompt)) return [data.prompt]
  }
  return []
}

const toPreviewCard = (card: any): PreviewCard => {
  const content = String(card.content ?? '')
  const encrypted = card.card_type === 'encrypted' || content.startsWith(ENCRYPT_MARKER)
  return {
    name: String(card.name || '未命名'),
    preview: encrypted ? '已加密卡片，正文不可预览' : content.slice(0, 60) + (content.length > 60 ? '...' : ''),
    encrypted,
    payload: card,
  }
}

const applyPreview = (cards: PreviewCard[]) => {
  preview.value = cards
  selected.value = cards.map((_, index) => index)
}

const handleFilesChange = async (uploadFile: any, uploadFiles: any[]) => {
  if (mode.value === 'convert') {
    const files = (uploadFiles || []).map((item: any) => item?.raw).filter((f: File) => !!f) as File[]
    if (files.length === 0) return
    const cards: PreviewCard[] = []
    for (const file of files) {
      try {
        const content = await readText(file)
        const name = file.name.replace(/\.(txt|md)$/i, '')
        cards.push(toPreviewCard({ name, content }))
      } catch {
        // 单个文件失败不影响其余文件
      }
    }
    if (cards.length === 0) {
      ElMessage.error('文件读取失败')
      return
    }
    applyPreview(cards)
    ElMessage.success(`成功读取 ${cards.length} 个文件`)
    return
  }

  const file: File | undefined = uploadFile?.raw
  if (!file) return
  try {
    const text = await readText(file)
    const data = JSON.parse(decodeExportData(text))
    registerImportedInsights(data)
    registerImportedRatingNotes(data)
    const cards = extractCards(data)
    if (cards.length === 0) {
      ElMessage.warning('文件里没有识别到可导入的卡片')
      resetPreview()
      return
    }
    applyPreview(cards.map(toPreviewCard))
    ElMessage.success(`成功读取 ${cards.length} 张卡片`)
  } catch (e: any) {
    ElMessage.error(`文件解析失败：${e?.message || '请确认是提示词导出文件'}`)
    resetPreview()
  }
}

const selectAllCards = () => {
  selected.value = preview.value.map((_, index) => index)
}

const clearSelection = () => {
  selected.value = []
}

const toggleCard = (index: number) => {
  const at = selected.value.indexOf(index)
  if (at > -1) selected.value.splice(at, 1)
  else selected.value.push(index)
}

const executeImport = async () => {
  const target = (props.packName || '').trim()
  if (!target) {
    ElMessage.error('当前卡包名称缺失，无法导入')
    return
  }
  const cards = selected.value.map((index) => preview.value[index]).filter((c): c is PreviewCard => !!c)
  if (cards.length === 0) return

  // 防重只在目标卡包内比：已存在的（同名同卡包）+ 本次文件内重复
  const existingKeys = new Set<string>()
  props.prompts.forEach((p) => {
    if ((p.category || '默认') !== target) return
    const key = buildPromptKey(p.name, target)
    if (key) existingKeys.add(key)
  })
  const importedKeys = new Set<string>()

  importing.value = true
  let successCount = 0
  let failCount = 0
  let skipCount = 0

  try {
    for (const card of cards) {
      const key = buildPromptKey(card.name, target)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const raw = card.payload || {}
      try {
        const res = await promptAPI.create({
          name: card.name,
          description: raw.description || '',
          content: raw.content,
          category: target,
          order_num: raw.order_num || 0,
          created_at: raw.created_at,
          fields: raw.fields || [],
          subcategories: raw.subcategories || [],
          card_type: raw.card_type || 'normal',
          password: raw.password || null,
          metadata: raw.metadata || undefined,
          creator_name: raw.creator_name || undefined,
          version: raw.version || undefined,
        })
        if (res.success && res.data) {
          successCount++
          await restoreInsightsFor(card.name, target, res.data.id)
          restoreRatingNotesFor(card.name, target, res.data.id)
        } else {
          failCount++
        }
      } catch {
        failCount++
      }
    }
  } finally {
    importing.value = false
  }

  flushRestoredRatingNotes()

  if (failCount === 0) {
    ElMessage.success(
      skipCount > 0
        ? `已导入 ${successCount} 张卡片到「${target}」，跳过重复 ${skipCount} 张`
        : `已导入 ${successCount} 张卡片到「${target}」`
    )
  } else {
    ElMessage.warning(
      `导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`
    )
  }

  visible.value = false
  emit('imported')
}

defineExpose({ open })
</script>

<style scoped>
.pack-import-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.pack-import-target {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  color: #1d4ed8;
  font-size: 13px;
  line-height: 1.6;
}

.pack-import-target .el-icon {
  margin-top: 2px;
  flex-shrink: 0;
  font-size: 16px;
}

.pack-import-modes {
  display: flex;
  flex-wrap: wrap;
}

.pack-import-upload {
  width: 100%;
}

.pack-import-upload :deep(.el-upload-dragger) {
  padding: 22px 16px;
  border-radius: 10px;
}

.pack-import-preview {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.pack-import-preview-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 8px;
  border-bottom: 1px solid #eef0f3;
  font-size: 13px;
  font-weight: 600;
  color: #334155;
}

.pack-import-preview-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.pack-import-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  max-height: 300px;
  overflow-y: auto;
  padding: 2px;
}

.pack-import-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 10px 12px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #ffffff;
  cursor: pointer;
  transition: border-color 0.18s ease, background 0.18s ease;
}

.pack-import-item:hover {
  border-color: #c7d2fe;
}

.pack-import-item.selected {
  border-color: #3b82f6;
  background: #eff6ff;
}

.pack-import-item-body {
  flex: 1;
  min-width: 0;
}

.pack-import-item-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pack-import-item-preview {
  margin-top: 4px;
  font-size: 12px;
  color: #64748b;
  line-height: 1.45;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

/* ========== 深色主题 ========== */
:root[data-theme='dark'] .pack-import-target {
  background: rgba(59, 130, 246, 0.12);
  border-color: rgba(59, 130, 246, 0.35);
  color: #93c5fd;
}

:root[data-theme='dark'] .pack-import-preview-head {
  border-color: rgba(100, 116, 139, 0.25);
  color: #e5e7eb;
}

:root[data-theme='dark'] .pack-import-item {
  background: #1e2434;
  border-color: #2e3650;
}

:root[data-theme='dark'] .pack-import-item:hover {
  border-color: rgba(59, 130, 246, 0.5);
}

:root[data-theme='dark'] .pack-import-item.selected {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.16);
}

:root[data-theme='dark'] .pack-import-item-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .pack-import-item-preview {
  color: #9ca3af;
}

:root[data-theme='dark'] .pack-import-modes :deep(.el-radio-button__inner) {
  border-color: #3a4360;
  color: #9ca3af;
  background: transparent;
}

:root[data-theme='dark'] .pack-import-modes :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
}

/* el-upload 拖拽区自带白底，暗色下必须显式覆盖 */
:root[data-theme='dark'] .pack-import-upload :deep(.el-upload-dragger) {
  background: rgba(22, 27, 34, 0.8);
  border-color: #2e3650;
}

:root[data-theme='dark'] .pack-import-upload :deep(.el-upload-dragger:hover) {
  border-color: #3b82f6;
}

:root[data-theme='dark'] .pack-import-upload :deep(.el-icon--upload) {
  color: #4b5563;
}

:root[data-theme='dark'] .pack-import-upload :deep(.el-upload__text) {
  color: #9ca3af;
}

:root[data-theme='dark'] .pack-import-upload :deep(.el-upload__text em) {
  color: #60a5fa;
}

:root[data-theme='dark'] .pack-import-upload :deep(.el-upload__tip) {
  color: #6b7280;
}
</style>
