<template>
<!-- 导出独立卡片弹窗 -->
<el-dialog
  v-model="exportStandaloneDialogVisible"
  title="导出独立卡片"
  width="800px"
  :close-on-click-modal="false"
  class="export-dialog"
  append-to-body
>
  <div class="export-standalone-content">
    <div class="export-hint">
      <el-alert
        title="请选择要导出的独立卡片"
        type="info"
        :closable="false"
        show-icon
      >
        勾选需要导出的卡片，然后点击导出按钮。导出的文件为JSON格式。
      </el-alert>
    </div>
    <div class="export-select-all">
      <el-checkbox
        v-model="selectAllStandalone"
        @change="handleSelectAllStandalone"
      >
        全选 ({{ selectedStandaloneCards.length }}/{{ combinedUncategorizedPrompts.length }})
      </el-checkbox>
    </div>
    <div v-if="combinedUncategorizedPrompts.length > 0" class="export-cards-grid">
      <div
        v-for="prompt in combinedUncategorizedPrompts"
        :key="prompt.id"
        class="export-card-item"
        :class="{ selected: selectedStandaloneCards.includes(prompt.id) }"
        @click="toggleStandaloneCard(prompt.id)"
      >
        <el-checkbox
          :model-value="selectedStandaloneCards.includes(prompt.id)"
          @click.stop
          @change="toggleStandaloneCard(prompt.id)"
        />
        <div class="export-card-content">
          <div class="export-card-name">
            {{ prompt.name }}
            <el-tag v-if="prompt.category === '拆书-未分类'" size="small" type="warning" style="margin-left: 6px;">拆书</el-tag>
          </div>
          <div class="export-card-preview">
            <template v-if="prompt.card_type === 'encrypted'">
              <el-icon><Lock /></el-icon> 已加密
            </template>
            <template v-else>
              {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
            </template>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="export-empty">
      <el-empty description="暂无独立卡片可导出" />
    </div>
  </div>
  <template #footer>
    <el-button @click="exportStandaloneDialogVisible = false">取消</el-button>
    <el-button
      type="primary"
      :disabled="selectedStandaloneCards.length === 0"
      @click="executeExportStandalone"
    >
      导出 ({{ selectedStandaloneCards.length }})
    </el-button>
  </template>
</el-dialog>

<!-- 导出卡包弹窗 -->
<el-dialog
  v-model="exportPackDialogVisible"
  title="导出卡包"
  width="600px"
  :close-on-click-modal="false"
  class="export-dialog"
  append-to-body
>
  <div class="export-pack-content">
    <div class="export-hint">
      <el-alert
        title="请选择要导出的卡包"
        type="info"
        :closable="false"
        show-icon
      >
        选择需要导出的卡包，导出的文件将包含卡包内的所有卡片。拆书卡包与普通卡包统一在此选择。
      </el-alert>
    </div>
    <div v-if="allExportablePacks.length > 0" class="export-pack-list">
      <div
        v-for="category in allExportablePacks"
        :key="category.name"
        class="export-pack-item"
        :class="{ selected: selectedPack === category.name }"
        @click="selectedPack = category.name"
      >
        <el-radio :model-value="selectedPack" :label="category.name">
          <div class="pack-item-info">
            <el-icon><FolderOpened /></el-icon>
            <span class="pack-item-name">{{ category.isBookAnalysis ? category.name.replace(BOOK_ANALYSIS_PREFIX, '') : category.name }}</span>
            <el-tag v-if="category.isBookAnalysis" size="small" type="warning">拆书</el-tag>
            <el-tag size="small" type="info">{{ category.prompts.length }} 张卡片</el-tag>
          </div>
        </el-radio>
      </div>
    </div>
    <div v-else class="export-empty">
      <el-empty description="暂无卡包可导出" />
    </div>
  </div>
  <template #footer>
    <el-button @click="exportPackDialogVisible = false">取消</el-button>
    <el-button
      type="primary"
      :disabled="!selectedPack"
      @click="executeExportPack"
    >
      导出卡包
    </el-button>
  </template>
</el-dialog>

<!-- 导入独立卡片弹窗 -->
<el-dialog
  v-model="importStandaloneDialogVisible"
  title="导入独立卡片"
  width="800px"
  :close-on-click-modal="false"
  class="import-dialog"
  append-to-body
>
  <div class="import-content">
    <div class="import-upload">
      <el-upload
        ref="standaloneUploadRef"
        drag
        accept=".json"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleStandaloneFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">只能上传 JSON 格式的提示词文件</div>
        </template>
      </el-upload>
    </div>
    <div v-if="importStandalonePreview.length > 0" class="import-preview">
      <div class="import-preview-header">
        <span>预览导入内容 ({{ importStandalonePreview.length }} 张卡片)</span>
        <el-checkbox
          v-model="selectAllImportStandalone"
          @change="handleSelectAllImportStandalone"
        >
          全选
        </el-checkbox>
      </div>
      <div class="import-cards-grid">
        <div
          v-for="(prompt, index) in importStandalonePreview"
          :key="index"
          class="import-card-item"
          :class="{ selected: selectedImportCards.includes(index) }"
          @click="toggleImportCard(index)"
        >
          <el-checkbox
            :model-value="selectedImportCards.includes(index)"
            @click.stop
            @change="toggleImportCard(index)"
          />
          <div class="import-card-content">
            <div class="import-card-name">{{ prompt.name }}</div>
            <div class="import-card-preview">
              <template v-if="prompt.card_type === 'encrypted'">
                <el-icon><Lock /></el-icon> 已加密
              </template>
              <template v-else>
                {{ prompt.content.slice(0, 30) }}{{ prompt.content.length > 30 ? '...' : '' }}
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <template #footer>
    <el-button @click="importStandaloneDialogVisible = false">取消</el-button>
    <el-button
      type="primary"
      :disabled="selectedImportCards.length === 0"
      @click="executeImportStandalone"
    >
      导入 ({{ selectedImportCards.length }})
    </el-button>
  </template>
</el-dialog>

<!-- 导入卡包弹窗 -->
<el-dialog
  v-model="importPackDialogVisible"
  title="导入卡包"
  width="600px"
  :close-on-click-modal="false"
  class="import-dialog"
  append-to-body
>
  <div class="import-content">
    <div class="import-upload">
      <el-upload
        ref="packUploadRef"
        drag
        accept=".json"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handlePackFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">只能上传 JSON 格式的卡包文件（普通 / 拆书自动识别）</div>
        </template>
      </el-upload>
    </div>
    <div v-if="importPackPreview" class="import-pack-preview">
      <div class="import-pack-info">
        <el-icon><FolderOpened /></el-icon>
        <span class="import-pack-name">{{ importPackPreview.packName }}</span>
        <el-tag v-if="importPackPreview.isBookAnalysis" size="small" type="warning">拆书</el-tag>
        <el-tag size="small" type="info">{{ importPackPreview.prompts.length }} 张卡片</el-tag>
      </div>
      <div class="import-pack-prompts">
        <div v-for="(prompt, index) in importPackPreview.prompts" :key="index" class="import-pack-prompt-item">
          <el-icon><Document /></el-icon>
          <span>{{ prompt.name }}</span>
        </div>
      </div>
    </div>
  </div>
  <template #footer>
    <el-button @click="importPackDialogVisible = false">取消</el-button>
    <el-button
      type="primary"
      :disabled="!importPackPreview"
      @click="executeImportPack"
    >
      导入卡包
    </el-button>
  </template>
</el-dialog>

<el-dialog
  v-model="importConvertDialogVisible"
  title="导入并转换 txt/md 文件"
  width="800px"
  :close-on-click-modal="false"
  class="import-dialog"
  append-to-body
>
  <div class="import-content">
    <div class="import-upload">
      <el-upload
        ref="convertUploadRef"
        drag
        multiple
        accept=".txt,.md"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleConvertFilesChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">支持批量上传 txt 或 md 格式文件</div>
        </template>
      </el-upload>
    </div>
    <div class="import-options">
      <el-switch v-model="mergeAsPack" active-text="合并为卡包" inactive-text="" />
      <el-input
        v-if="mergeAsPack"
        v-model="mergePackName"
        placeholder="请输入卡包名称"
        style="margin-top: 10px;"
      />
    </div>
    <div v-if="importConvertPreview.length > 0" class="import-preview">
      <div class="import-preview-header">
        <span>预览导入内容 ({{ importConvertPreview.length }} 个文件)</span>
        <el-checkbox
          v-model="selectAllImportConvert"
          @change="handleSelectAllImportConvert"
        >
          全选
        </el-checkbox>
      </div>
      <div class="import-cards-grid">
        <div
          v-for="(file, index) in importConvertPreview"
          :key="index"
          class="import-card-item"
          :class="{ selected: selectedImportConvertCards.includes(index) }"
          @click="toggleImportConvertCard(index)"
        >
          <el-checkbox
            :model-value="selectedImportConvertCards.includes(index)"
            @click.stop
            @change="toggleImportConvertCard(index)"
          />
          <div class="import-card-content">
            <div class="import-card-name">{{ file.name }}</div>
            <div class="import-card-preview">{{ file.content.slice(0, 50) }}{{ file.content.length > 50 ? '...' : '' }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
  <template #footer>
    <el-button @click="importConvertDialogVisible = false">取消</el-button>
    <el-button
      type="primary"
      :disabled="selectedImportConvertCards.length === 0"
      @click="executeImportConvert"
    >
      导入 ({{ selectedImportConvertCards.length }})
    </el-button>
  </template>
</el-dialog>

<!-- 整体备份导入对话框 -->
<el-dialog
  v-model="importAllDialogVisible"
  title="整体恢复（全部导入）"
  width="600px"
  :close-on-click-modal="false"
  class="import-all-dialog"
  append-to-body
>
  <div class="import-all-content">
    <div class="import-all-upload">
      <el-upload
        ref="importAllUploadRef"
        drag
        accept=".json"
        :auto-upload="false"
        :show-file-list="false"
        :on-change="handleImportAllFileChange"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将备份文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">只能上传整体备份的 JSON 格式文件</div>
        </template>
      </el-upload>
    </div>
    <div v-if="importAllPreview" class="import-all-preview">
      <el-alert
        title="警告"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 16px;"
      >
        导入后将覆盖所有现有数据，包括独立卡片和卡包。此操作不可恢复！
      </el-alert>
      <div class="import-all-info">
        <div class="import-all-item">
          <el-icon><Document /></el-icon>
          <span>独立卡片：<strong>{{ importAllPreview.standaloneCount }}</strong> 张</span>
        </div>
        <div class="import-all-item">
          <el-icon><FolderOpened /></el-icon>
          <span>卡包：<strong>{{ importAllPreview.packCount }}</strong> 个</span>
        </div>
        <div class="import-all-item">
          <el-icon><Collection /></el-icon>
          <span>总提示词数：<strong>{{ importAllPreview.totalPrompts }}</strong> 条</span>
        </div>
      </div>
    </div>
  </div>
  <template #footer>
    <el-button @click="importAllDialogVisible = false">取消</el-button>
    <el-button
      type="danger"
      :disabled="!importAllPreview"
      @click="executeImportAll"
    >
      导入并覆盖
    </el-button>
  </template>
</el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import { buildPromptKey } from '@/utils/promptDedupMerge'
import {
  encodeExportData,
  decodeExportData,
  importedInsights,
  registerImportedInsights,
  restoreInsightsFor,
  collectInsightsByKey,
  buildExistingKeySet,
  collectRatingNotesByKey,
  registerImportedRatingNotes,
  restoreRatingNotesFor,
  flushRestoredRatingNotes,
  resetImportedRatingNotes,
} from '@/utils/promptTransfer'
import { BOOK_ANALYSIS_PREFIX } from '@/constants/promptCategories'

const props = defineProps<{
  prompts: Prompt[]
  combinedUncategorizedPrompts: Prompt[]
  allExportablePacks: { name: string; prompts: Prompt[]; isBookAnalysis: boolean }[]
  categoryList: { name: string; prompts: Prompt[] }[]
  bookAnalysisCategoryList: { name: string; prompts: Prompt[] }[]
  customCategories: string[]
  refresh: () => Promise<void>
}>()

const emit = defineEmits<{
  (e: 'add-category', name: string): void
  (e: 'set-categories', names: string[]): void
}>()

// 导出独立卡片相关状态
const exportStandaloneDialogVisible = ref(false)
const selectedStandaloneCards = ref<number[]>([])
const selectAllStandalone = ref(false)

// 导出卡包相关状态
const exportPackDialogVisible = ref(false)
const selectedPack = ref<string | null>(null)

// 导入独立卡片相关状态
const importStandaloneDialogVisible = ref(false)
const importStandalonePreview = ref<Prompt[]>([])
const selectedImportCards = ref<number[]>([])
const selectAllImportStandalone = ref(false)

// 导入卡包相关状态
const importPackDialogVisible = ref(false)
const importPackPreview = ref<{ packName: string; prompts: Prompt[]; isBookAnalysis: boolean } | null>(null)

// 导入转换 txt/md 相关状态
const importConvertDialogVisible = ref(false)
const importConvertPreview = ref<{ name: string; content: string }[]>([])
const selectedImportConvertCards = ref<number[]>([])
const selectAllImportConvert = ref(false)
const convertUploadRef = ref<any>()
const mergeAsPack = ref(false)
const mergePackName = ref('')

// 整体备份导入相关状态
const importAllDialogVisible = ref(false)
const importAllPreview = ref<{
  standaloneCount: number
  packCount: number
  totalPrompts: number
  data: any
} | null>(null)
const importAllUploadRef = ref<any>()

// 打开整体导入对话框
const openImportAllDialog = () => {
  importedInsights.value = {}
  resetImportedRatingNotes()
  resetImportedRatingNotes()
  importAllPreview.value = null
  importAllDialogVisible.value = true
}

// 处理整体备份文件选择（统一入口：新版完整备份 + 旧版/拆书完整备份）
const handleImportAllFileChange = async (file: any) => {
  try {
    const text = await file.raw.text()
    const decodedText = decodeExportData(text)
    const backupData = JSON.parse(decodedText)
    registerImportedInsights(backupData)
    registerImportedRatingNotes(backupData)

    let standalone: any[] = []
    let packs: Record<string, any[]> = {}
    let isBookAnalysis = false

    // 新版完整备份（普通 + 拆书统一包含在内）
    if (backupData.type === 'full-backup' && backupData.data) {
      standalone = backupData.data.standalone || []
      packs = backupData.data.packs || {}
    } else if (backupData.type === 'book-analysis-full-backup' && backupData.data) {
      // 旧版拆书完整备份
      standalone = backupData.data.standalone || []
      packs = backupData.data.packs || {}
      isBookAnalysis = true
    } else if (Array.isArray(backupData)) {
      // 旧版格式1：直接 prompts 数组
      standalone = backupData.filter((p: any) => p && p.name && p.content)
      isBookAnalysis = true
    } else if (backupData.prompts && Array.isArray(backupData.prompts)) {
      // 旧版格式2：{ prompts, categories }
      standalone = backupData.prompts.filter((p: any) => p && p.name && p.content)
      ;(backupData.categories || []).forEach((cat: string) => {
        if (cat && cat !== '未分类') packs[cat] = []
      })
      standalone.forEach((p: any) => {
        const cat = p.category
        if (cat && cat !== '未分类') {
          if (!packs[cat]) packs[cat] = []
          packs[cat].push(p)
        }
      })
      isBookAnalysis = true
    } else {
      throw new Error('无效的备份文件格式')
    }

    const standaloneCount = standalone.length
    const packCount = Object.keys(packs).length
    let totalPrompts = standaloneCount
    Object.values(packs).forEach((pack: any) => {
      totalPrompts += Array.isArray(pack) ? pack.length : 0
    })

    importAllPreview.value = {
      standaloneCount,
      packCount,
      totalPrompts,
      data: { standalone, packs, isBookAnalysis }
    }
  } catch (error: any) {
    ElMessage.error(`文件解析失败：${error.message}`)
    importAllPreview.value = null
  }
}

// 执行整体导入
const executeImportAll = async () => {
  if (!importAllPreview.value) return

  try {
    await ElMessageBox.confirm(
      '确定要导入备份数据吗？这将覆盖所有现有数据，包括独立卡片和卡包。此操作不可恢复！',
      '整体恢复',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error'
      }
    )

    const { standalone, packs, isBookAnalysis } = importAllPreview.value.data

    // 先删除所有现有数据
    const deletePromises = props.prompts.map(p => promptAPI.delete(p.id))
    await Promise.all(deletePromises)

    // 防重：备份文件内同名（同分类）只保留第一条，避免恢复后仍有重复
    let skipCount = 0
    const importedKeys = new Set<string>()

    // 导入独立卡片
    if (standalone && standalone.length > 0) {
      for (const prompt of standalone) {
        let category = prompt.category || '未分类'
        if (isBookAnalysis) {
          if (category === '未分类') category = '拆书-未分类'
          else if (!category.startsWith(BOOK_ANALYSIS_PREFIX)) category = BOOK_ANALYSIS_PREFIX + category
        }
        const key = buildPromptKey(prompt.name, category)
        if (importedKeys.has(key)) {
          skipCount++
          continue
        }
        importedKeys.add(key)
        const res = await promptAPI.create({
          name: prompt.name,
          description: prompt.description || '',
          content: prompt.content,
          category,
          order_num: prompt.order_num || 0,
          created_at: prompt.created_at,
          fields: prompt.fields || [],
          subcategories: prompt.subcategories || [],
          card_type: prompt.card_type || 'normal',
          password: prompt.password || null,
          metadata: prompt.metadata || null,
          creator_name: prompt.creator_name || null,
          version: prompt.version || null,
        })
        if (res.success) {
          await restoreInsightsFor(prompt.name, category, res.data?.id)
          restoreRatingNotesFor(prompt.name, category, res.data?.id)
        }
      }
    }

    // 导入卡包
    if (packs) {
      for (const [packName, promptsInPack] of Object.entries(packs)) {
        let fullPackName = packName
        if (isBookAnalysis && !fullPackName.startsWith(BOOK_ANALYSIS_PREFIX)) {
          fullPackName = BOOK_ANALYSIS_PREFIX + fullPackName
        }
        if (Array.isArray(promptsInPack)) {
          for (const prompt of promptsInPack) {
            const key = buildPromptKey(prompt.name, fullPackName)
            if (importedKeys.has(key)) {
              skipCount++
              continue
            }
            importedKeys.add(key)
            const res = await promptAPI.create({
              name: prompt.name,
              description: prompt.description || '',
              content: prompt.content,
              category: fullPackName,
              order_num: prompt.order_num || 0,
              created_at: prompt.created_at,
              fields: prompt.fields || [],
              subcategories: prompt.subcategories || [],
              card_type: prompt.card_type || 'normal',
              password: prompt.password || null,
              metadata: prompt.metadata || null,
              creator_name: prompt.creator_name || null,
              version: prompt.version || null,
            })
            if (res.success) {
              await restoreInsightsFor(prompt.name, fullPackName, res.data?.id)
              restoreRatingNotesFor(prompt.name, fullPackName, res.data?.id)
            }
          }
        }
      }
    }

    // 更新自定义卡包列表
    if (packs) {
      emit('set-categories', Object.keys(packs).map(p =>
        isBookAnalysis && !p.startsWith(BOOK_ANALYSIS_PREFIX) ? BOOK_ANALYSIS_PREFIX + p : p
      ))
    }

    // 定级 / 心得笔记一次写回 localStorage（整体恢复会建很多卡片，逐张写太慢）
    flushRestoredRatingNotes()

    ElMessage.success(skipCount > 0 ? `整体恢复成功（跳过重复 ${skipCount} 张）` : '整体恢复成功')
    importAllDialogVisible.value = false
    importAllPreview.value = null
    await props.refresh()
  } catch (error) {
    // 取消导入
  }
}

// 打开导出独立卡片弹窗
const openExportStandaloneDialog = () => {
  selectedStandaloneCards.value = []
  selectAllStandalone.value = false
  exportStandaloneDialogVisible.value = true
}

// 全选/取消全选独立卡片
const handleSelectAllStandalone = (val: boolean) => {
  if (val) {
    selectedStandaloneCards.value = props.combinedUncategorizedPrompts.map(p => p.id)
  } else {
    selectedStandaloneCards.value = []
  }
}

// 切换单个独立卡片选择
const toggleStandaloneCard = (id: number) => {
  const index = selectedStandaloneCards.value.indexOf(id)
  if (index > -1) {
    selectedStandaloneCards.value.splice(index, 1)
  } else {
    selectedStandaloneCards.value.push(id)
  }
  selectAllStandalone.value = selectedStandaloneCards.value.length === props.combinedUncategorizedPrompts.length
}

// 执行导出独立卡片
const executeExportStandalone = async () => {
  const cardsToExport = props.combinedUncategorizedPrompts.filter(p => selectedStandaloneCards.value.includes(p.id))
  const insights = await collectInsightsByKey(props.prompts)
  const extras = collectRatingNotesByKey(cardsToExport)
  const exportData = {
    type: 'standalone-cards',
    version: '1.1',
    exportTime: new Date().toISOString(),
    insights,
    extras,
    prompts: cardsToExport.map(p => ({
      name: p.name,
      description: p.description || '',
      content: p.content,
      category: p.category,
      order_num: p.order_num,
      created_at: p.created_at,
      fields: p.fields,
      subcategories: p.subcategories,
      card_type: p.card_type || 'normal',
      password: p.password || null,
      metadata: p.metadata || null,
      creator_name: p.creator_name || null,
      version: p.version || null,
    }))
  }
  
  const blob = new Blob([encodeExportData(JSON.stringify(exportData))], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `独立卡片_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success(`成功导出 ${cardsToExport.length} 张独立卡片`)
  exportStandaloneDialogVisible.value = false
}

// 打开导出卡包弹窗
const openExportPackDialog = () => {
  selectedPack.value = null
  exportPackDialogVisible.value = true
}

// 执行导出卡包（统一入口：普通卡包与拆书卡包）
const executeExportPack = async () => {
  if (!selectedPack.value) return

  const insights = await collectInsightsByKey(props.prompts)

  // 优先匹配拆书卡包（拆书导出沿用"去前缀、可移植"格式）
  const bookCategory = props.bookAnalysisCategoryList.find(c => c.name === selectedPack.value)
  if (bookCategory) {
    const exportData = {
      type: 'book-analysis-packs',
      exportedAt: new Date().toISOString(),
      insights,
      extras: collectRatingNotesByKey(bookCategory.prompts),
      packs: [{
        packName: bookCategory.name.replace(BOOK_ANALYSIS_PREFIX, ''),
        prompts: bookCategory.prompts.map(p => ({
          name: p.name,
          description: p.description || '',
          content: p.content,
          category: bookCategory.name,
          order_num: p.order_num || 0,
          created_at: p.created_at,
          fields: p.fields || [],
          card_type: p.card_type || 'normal',
          password: p.password || null,
          metadata: p.metadata || null,
          creator_name: p.creator_name || null,
          version: p.version || null,
        }))
      }]
    }
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `拆书卡包导出_${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success(`成功导出拆书卡包"${bookCategory.name.replace(BOOK_ANALYSIS_PREFIX, '')}"，包含 ${bookCategory.prompts.length} 张卡片`)
    exportPackDialogVisible.value = false
    return
  }

  const category = props.categoryList.find(c => c.name === selectedPack.value)
  if (!category) return

  const exportData = {
    type: 'card-pack',
    version: '1.1',
    exportTime: new Date().toISOString(),
    insights,
    extras: collectRatingNotesByKey(category.prompts),
    packName: category.name,
    prompts: category.prompts.map(p => ({
      name: p.name,
      description: p.description || '',
      content: p.content,
      category: p.category,
      order_num: p.order_num,
      created_at: p.created_at,
      fields: p.fields,
      subcategories: p.subcategories,
      card_type: p.card_type || 'normal',
      password: p.password || null,
      metadata: p.metadata || null,
      creator_name: p.creator_name || null,
      version: p.version || null,
    }))
  }

  const blob = new Blob([encodeExportData(JSON.stringify(exportData))], { type: 'application/octet-stream' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `卡包_${category.name}_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.json`
  a.click()
  URL.revokeObjectURL(url)

  ElMessage.success(`成功导出卡包"${category.name}"，包含 ${category.prompts.length} 张卡片`)
  exportPackDialogVisible.value = false
}

// 打开导入独立卡片弹窗
const openImportStandaloneDialog = () => {
  importedInsights.value = {}
  resetImportedRatingNotes()
  importStandalonePreview.value = []
  selectedImportCards.value = []
  selectAllImportStandalone.value = false
  importStandaloneDialogVisible.value = true
}

// 处理独立卡片文件选择
const handleStandaloneFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const decodedContent = decodeExportData(content)
      const data = JSON.parse(decodedContent)
      registerImportedInsights(data)
      registerImportedRatingNotes(data)
      
      if (data.type === 'standalone-cards' && Array.isArray(data.prompts)) {
        importStandalonePreview.value = data.prompts
        selectedImportCards.value = data.prompts.map((_: any, index: number) => index)
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 ${data.prompts.length} 张卡片`)
      } else if (data.type === 'single-prompt' && data.prompt) {
        importStandalonePreview.value = [data.prompt]
        selectedImportCards.value = [0]
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 1 张卡片`)
      } else if (data.type === 'book-analysis-standalone' && Array.isArray(data.prompts)) {
        // 拆书独立卡片格式
        importStandalonePreview.value = data.prompts
        selectedImportCards.value = data.prompts.map((_: any, index: number) => index)
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 ${data.prompts.length} 张拆书独立卡片`)
      } else if (Array.isArray(data)) {
        // 兼容：直接数组
        const list = data.filter((p: any) => p && p.name && p.content)
        importStandalonePreview.value = list
        selectedImportCards.value = list.map((_: any, index: number) => index)
        selectAllImportStandalone.value = true
        ElMessage.success(`成功读取 ${list.length} 张卡片`)
      } else if (data.type === 'card-pack' || data.type === 'book-analysis-packs') {
        ElMessage.warning('这是卡包文件，请使用"导入卡包"功能')
      } else {
        ElMessage.error('文件格式不正确')
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的导出文件')
    }
  }
  reader.readAsText(file.raw)
}

// 全选/取消全选导入卡片
const handleSelectAllImportStandalone = (val: boolean) => {
  if (val) {
    selectedImportCards.value = importStandalonePreview.value.map((_, index) => index)
  } else {
    selectedImportCards.value = []
  }
}

// 切换单个导入卡片选择
const toggleImportCard = (index: number) => {
  const idx = selectedImportCards.value.indexOf(index)
  if (idx > -1) {
    selectedImportCards.value.splice(idx, 1)
  } else {
    selectedImportCards.value.push(index)
  }
  selectAllImportStandalone.value = selectedImportCards.value.length === importStandalonePreview.value.length
}

// 执行导入独立卡片
const executeImportStandalone = async () => {
  const cardsToImport = selectedImportCards.value
    .map(index => importStandalonePreview.value[index])
    .filter((card): card is Prompt => card !== undefined)
  
  let successCount = 0
  let failCount = 0
  let skipCount = 0

  // 防重：已存在的判重键（按目标分类规则计算）+ 文件内已导入的判重键
  const existingKeys = new Set<string>()
  props.prompts.forEach(p => {
    const cat = (p.category || '').startsWith(BOOK_ANALYSIS_PREFIX) ? p.category : '未分类'
    const key = buildPromptKey(p.name, cat)
    if (key) existingKeys.add(key)
  })
  const importedKeys = new Set<string>()

  for (const card of cardsToImport) {
    try {
      // 拆书卡片保留其拆书分类（拆书-未分类 等），普通卡片一律归入 未分类
      const category = (card.category || '').startsWith(BOOK_ANALYSIS_PREFIX)
        ? card.category
        : '未分类'
      const key = buildPromptKey(card.name, category)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const res = await promptAPI.create({
        name: card.name,
        description: card.description || '',
        content: card.content,
        category,
        order_num: card.order_num || 0,
        created_at: card.created_at,
        fields: card.fields,
        subcategories: card.subcategories,
        card_type: card.card_type || 'normal',
        password: card.password || null,
        metadata: card.metadata || undefined,
        creator_name: card.creator_name || undefined,
        version: card.version || undefined,
      })
      if (res.success) {
        successCount++
        await restoreInsightsFor(card.name, category, res.data?.id)
        restoreRatingNotesFor(card.name, category, res.data?.id)
      } else {
        failCount++
      }
    } catch (error) {
      failCount++
    }
  }
  
  await props.refresh()

  flushRestoredRatingNotes()
  
  if (failCount === 0) {
    ElMessage.success(
      skipCount > 0
        ? `成功导入 ${successCount} 张独立卡片，跳过重复 ${skipCount} 张`
        : `成功导入 ${successCount} 张独立卡片`
    )
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
  }
  
  importStandaloneDialogVisible.value = false
}

// 打开导入卡包弹窗
const openImportPackDialog = () => {
  importedInsights.value = {}
  resetImportedRatingNotes()
  importPackPreview.value = null
  importPackDialogVisible.value = true
}

// 处理卡包文件选择（统一入口：普通卡包 + 拆书卡包）
const handlePackFileChange = (file: any) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target?.result as string
      const decodedContent = decodeExportData(content)
      const data = JSON.parse(decodedContent)
      registerImportedInsights(data)
      registerImportedRatingNotes(data)

      // 普通卡包（编码格式）
      if (data.type === 'card-pack' && Array.isArray(data.prompts)) {
        importPackPreview.value = {
          packName: data.packName,
          prompts: data.prompts,
          isBookAnalysis: (data.packName || '').startsWith(BOOK_ANALYSIS_PREFIX)
        }
        ElMessage.success(`成功读取卡包"${data.packName}"，包含 ${data.prompts.length} 张卡片`)
      } else if (data.type === 'standalone-cards') {
        ElMessage.warning('这是独立卡片文件，请使用"导入独立卡片"功能')
      } else if (data.type === 'book-analysis-packs' && Array.isArray(data.packs) && data.packs.length > 0) {
        // 拆书多卡包格式：取第一个
        const first = data.packs[0]
        importPackPreview.value = {
          packName: first.packName,
          prompts: first.prompts || [],
          isBookAnalysis: true
        }
        ElMessage.success(`成功读取拆书卡包"${first.packName}"，包含 ${first.prompts?.length || 0} 张卡片`)
      } else if (data.packName && Array.isArray(data.prompts)) {
        // 拆书单卡包格式（无 type 标识）
        importPackPreview.value = {
          packName: data.packName,
          prompts: data.prompts,
          isBookAnalysis: true
        }
        ElMessage.success(`成功读取拆书卡包"${data.packName}"，包含 ${data.prompts.length} 张卡片`)
      } else {
        ElMessage.error('文件格式不正确')
      }
    } catch (error) {
      ElMessage.error('文件解析失败，请确保是有效的导出文件')
    }
  }
  reader.readAsText(file.raw)
}

// 执行导入卡包（统一入口：拆书卡包自动加前缀）
const executeImportPack = async () => {
  if (!importPackPreview.value) return

  const { packName, prompts: cardsToImport, isBookAnalysis } = importPackPreview.value

  // 拆书卡包自动补齐前缀，保证落入拆书 Tab
  const fullPackName = isBookAnalysis && !packName.startsWith(BOOK_ANALYSIS_PREFIX)
    ? BOOK_ANALYSIS_PREFIX + packName
    : packName

  // 检查卡包是否已存在，如果不存在则创建
  if (!props.customCategories.includes(fullPackName) && fullPackName !== '默认') {
    emit('add-category', fullPackName)
  }

  let successCount = 0
  let failCount = 0
  let skipCount = 0

  // 防重：已存在的判重键（目标包名）+ 文件内已导入的判重键
  const existingKeys = buildExistingKeySet(props.prompts)
  const importedKeys = new Set<string>()

  for (const card of cardsToImport) {
    try {
      const key = buildPromptKey(card.name, fullPackName)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const res = await promptAPI.create({
        name: card.name,
        description: card.description || '',
        content: card.content,
        category: fullPackName,
        order_num: card.order_num || 0,
        created_at: card.created_at,
        fields: card.fields,
        subcategories: card.subcategories,
        card_type: card.card_type || 'normal',
        password: card.password || null,
        metadata: card.metadata || undefined,
        creator_name: card.creator_name || undefined,
        version: card.version || undefined,
      })
      if (res.success) {
        successCount++
        await restoreInsightsFor(card.name, fullPackName, res.data?.id)
        restoreRatingNotesFor(card.name, fullPackName, res.data?.id)
      } else {
        failCount++
      }
    } catch (error) {
      failCount++
    }
  }

  await props.refresh()

  flushRestoredRatingNotes()

  if (failCount === 0) {
    ElMessage.success(
      skipCount > 0
        ? `成功导入卡包"${fullPackName}"，包含 ${successCount} 张卡片，跳过重复 ${skipCount} 张`
        : `成功导入卡包"${fullPackName}"，包含 ${successCount} 张卡片`
    )
  } else {
    ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
  }

  importPackDialogVisible.value = false
}

// 打开导入并转换txt/md弹窗
const openImportConvertDialog = () => {
  importedInsights.value = {}
  resetImportedRatingNotes()
  importConvertPreview.value = []
  selectedImportConvertCards.value = []
  selectAllImportConvert.value = false
  mergeAsPack.value = false
  mergePackName.value = ''
  convertUploadRef.value?.clearFiles()
  importConvertDialogVisible.value = true
}

// 读取文件内容
const readFileContent = (file: File): Promise<string> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = (e) => {
      resolve(e.target?.result as string)
    }
    reader.onerror = reject
    reader.readAsText(file)
  })
}

// 处理多个txt/md文件选择
const handleConvertFilesChange = (_uploadFile: any, uploadFiles: any[]) => {
  const files = uploadFiles?.map((item: any) => item.raw).filter((f: File) => f) || []
  
  if (files.length === 0) return
  
  const readPromises = files.map((f: File) => {
    const name = f.name.replace(/\.(txt|md)$/i, '')
    return readFileContent(f).then(content => ({ name, content }))
  })
  
  Promise.all(readPromises).then(items => {
    importConvertPreview.value = items
    selectedImportConvertCards.value = items.map((_, index) => index)
    selectAllImportConvert.value = true
    ElMessage.success(`成功读取 ${items.length} 个文件`)
  }).catch(() => {
    ElMessage.error('文件读取失败')
  })
}

// 全选/取消全选转换导入卡片
const handleSelectAllImportConvert = (val: boolean) => {
  if (val) {
    selectedImportConvertCards.value = importConvertPreview.value.map((_, index) => index)
  } else {
    selectedImportConvertCards.value = []
  }
}

// 切换单个转换导入卡片选择
const toggleImportConvertCard = (index: number) => {
  const idx = selectedImportConvertCards.value.indexOf(index)
  if (idx > -1) {
    selectedImportConvertCards.value.splice(idx, 1)
  } else {
    selectedImportConvertCards.value.push(index)
  }
  selectAllImportConvert.value = selectedImportConvertCards.value.length === importConvertPreview.value.length
}

// 执行导入转换txt/md
const executeImportConvert = async () => {
  const cardsToImport = selectedImportConvertCards.value
    .map(index => importConvertPreview.value[index])
    .filter((card): card is { name: string; content: string } => card !== undefined)
  
  let targetCategory = '未分类'
  if (mergeAsPack.value && mergePackName.value.trim()) {
    targetCategory = mergePackName.value.trim()
    if (!props.customCategories.includes(targetCategory)) {
      emit('add-category', targetCategory)
    }
  }
  
  let successCount = 0
  let failCount = 0
  let skipCount = 0

  // 防重：已存在的判重键（目标分类）+ 文件内已导入的判重键
  const existingKeys = buildExistingKeySet(props.prompts)
  const importedKeys = new Set<string>()

  for (const card of cardsToImport) {
    try {
      const key = buildPromptKey(card.name, targetCategory)
      if (existingKeys.has(key) || importedKeys.has(key)) {
        skipCount++
        continue
      }
      importedKeys.add(key)
      const res = await promptAPI.create({
        name: card.name,
        content: card.content,
        category: targetCategory,
        order_num: 0,
        fields: []
      })
      if (res.success) {
        successCount++
        await restoreInsightsFor(card.name, targetCategory, res.data?.id)
      } else {
        failCount++
      }
    } catch (error) {
      failCount++
    }
  }
  
  await props.refresh()
  
  if (mergeAsPack.value && mergePackName.value.trim()) {
    if (failCount === 0) {
      ElMessage.success(
        skipCount > 0
          ? `成功导入卡包"${targetCategory}"，包含 ${successCount} 张卡片，跳过重复 ${skipCount} 张`
          : `成功导入卡包"${targetCategory}"，包含 ${successCount} 张卡片`
      )
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
    }
  } else {
    if (failCount === 0) {
      ElMessage.success(
        skipCount > 0
          ? `成功导入 ${successCount} 张独立卡片，跳过重复 ${skipCount} 张`
          : `成功导入 ${successCount} 张独立卡片`
      )
    } else {
      ElMessage.warning(`导入完成：成功 ${successCount} 张，失败 ${failCount} 张${skipCount > 0 ? `，跳过重复 ${skipCount} 张` : ''}`)
    }
  }
  
  mergeAsPack.value = false
  mergePackName.value = ''
  convertUploadRef.value?.clearFiles()
  importConvertDialogVisible.value = false
}

defineExpose({
  openImportAllDialog,
  openExportStandaloneDialog,
  openExportPackDialog,
  openImportStandaloneDialog,
  openImportPackDialog,
  openImportConvertDialog,
})
</script>

<style scoped>
.export-dialog .export-standalone-content,
.export-dialog .export-pack-content {
  min-height: 300px;
}

.export-hint {
  margin-bottom: 16px;
}

.export-select-all {
  margin-bottom: 16px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 8px;
}

.export-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
  padding: 4px;
}

.export-card-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.export-card-item:hover {
  border-color: #fa8c16;
  box-shadow: 0 2px 8px rgba(250, 140, 22, 0.15);
}

.export-card-item.selected {
  border-color: #52c41a;
  background: #f6ffed;
}

.export-card-content {
  flex: 1;
  min-width: 0;
}

.export-card-name {
  font-weight: 600;
  font-size: 13px;
  color: #262626;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-card-preview {
  font-size: 11px;
  color: #8c8c8c;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.export-empty {
  padding: 40px;
}

.export-pack-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.export-pack-item {
  padding: 12px 16px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.export-pack-item:hover {
  border-color: #52c41a;
  box-shadow: 0 2px 8px rgba(82, 196, 26, 0.15);
}

.export-pack-item.selected {
  border-color: #52c41a;
  background: #f6ffed;
}

.pack-item-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pack-item-info .el-icon {
  color: #52c41a;
  font-size: 18px;
}

.pack-item-name {
  font-weight: 600;
  color: #262626;
}

/* 导入弹窗样式 */
.import-dialog .import-content {
  min-height: 300px;
}

/* 整体备份导入弹窗样式 */
.import-all-content {
  min-height: 300px;
}

.import-all-upload {
  margin-bottom: 20px;
}

.import-all-preview {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.import-all-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.import-all-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #fff;
  border-radius: 6px;
  font-size: 14px;
  color: #606266;
}

.import-all-item .el-icon {
  color: var(--el-color-primary);
  font-size: 18px;
}

.import-all-item strong {
  color: var(--el-color-primary);
  font-weight: 600;
}

.import-upload {
  margin-bottom: 20px;
}

.import-preview {
  margin-top: 20px;
}

.import-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e8e8e8;
  font-weight: 600;
  color: #262626;
}

.import-cards-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  max-height: 300px;
  overflow-y: auto;
  padding: 4px;
}

.import-card-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 12px;
  background: #fff;
  border: 2px solid #e8e8e8;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.import-card-item:hover {
  border-color: var(--el-color-primary);
  box-shadow: 0 2px 8px color-mix(in srgb, var(--el-color-primary) 15%, transparent);
}

.import-card-item.selected {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.import-card-content {
  flex: 1;
  min-width: 0;
}

.import-card-name {
  font-weight: 600;
  font-size: 13px;
  color: #262626;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.import-card-preview {
  font-size: 11px;
  color: #8c8c8c;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  overflow: hidden;
}

.import-pack-preview {
  margin-top: 20px;
  padding: 16px;
  background: #f6ffed;
  border-radius: 8px;
  border: 2px solid #b7eb8f;
}

.import-pack-info {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #b7eb8f;
}

.import-pack-info .el-icon {
  color: #52c41a;
  font-size: 24px;
}

.import-pack-name {
  font-weight: 600;
  font-size: 16px;
  color: #262626;
}

.import-pack-prompts {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.import-pack-prompt-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #fff;
  border-radius: 6px;
  font-size: 13px;
  color: #595959;
}

.import-pack-prompt-item .el-icon {
  color: #52c41a;
}

/* --- 暗色主题 - 创建按钮 --- */
:root[data-theme='dark'] .export-select-all {
  background: rgba(28, 35, 51, 0.6);
}

:root[data-theme='dark'] .export-card-item {
  background: rgba(22, 27, 34, 0.8);
  border-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .export-card-item:hover {
  border-color: rgba(249, 115, 22, 0.5);
}

:root[data-theme='dark'] .export-card-item.selected {
  border-color: #10b981;
  background: rgba(16, 185, 129, 0.15);
}

:root[data-theme='dark'] .export-card-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .export-card-preview {
  color: #9ca3af;
}
</style>
