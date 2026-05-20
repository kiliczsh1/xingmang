<template>
  <div class="book-analysis-page">
    <div class="page-header">
      <el-button class="back-btn" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h1 class="page-title">拆书库</h1>
      <div class="header-actions">
        <el-button @click="managePrompts">
          <el-icon><Document /></el-icon>
          拆书库提示词
        </el-button>
        <el-button @click="showHistoryDialog = true">
          <el-icon><Document /></el-icon>
          历史记录
        </el-button>
        <el-button type="primary" @click="startAnalysis" :loading="analyzing">
          <el-icon><Lightning /></el-icon>
          开始拆书
        </el-button>
        <el-button
          v-if="analyzing"
          type="danger"
          plain
          @click="stopAnalysis"
        >
          <el-icon><CloseBold /></el-icon>
          停止拆书
        </el-button>
      </div>
    </div>

    <div class="main-content">
      <div class="left-panel">
        <div class="chapter-select-area">
          <div class="section-header">
            <el-icon><FolderOpened /></el-icon>
            <span>选择章节</span>
            <el-checkbox
              v-model="selectAll"
              :indeterminate="isIndeterminate"
              @change="handleSelectAllChange"
              class="select-all-checkbox"
            >
              全选
            </el-checkbox>
          </div>
          <div class="chapter-list">
            <el-checkbox-group v-model="selectedChapters">
              <div
                v-for="chapter in chapters"
                :key="chapter.id"
                class="chapter-item"
              >
                <el-checkbox :value="chapter.id">
                  <span class="chapter-title">{{ chapter.title }}</span>
                  <span class="chapter-words">({{ getWordCount(chapter.content) }}字)</span>
                </el-checkbox>
              </div>
            </el-checkbox-group>
            <div v-if="chapters.length === 0" class="empty-chapters">
              <el-icon :size="32"><Document /></el-icon>
              <p>暂无章节</p>
            </div>
          </div>
        </div>

        <div class="config-area">
          <div class="config-item">
            <el-popover
              placement="right"
              :width="300"
              trigger="hover"
              :content="'选择章节和提示词，AI 将根据您的要求对内容进行分析拆解。'"
            >
              <template #reference>
                <div class="config-desc-trigger">
                  <el-icon><QuestionFilled /></el-icon>
                  <span>拆书要求</span>
                </div>
              </template>
            </el-popover>
          </div>

          <div class="config-item">
            <div class="config-label">
              <el-icon><Document /></el-icon>
              选择提示词
            </div>
            <div class="prompt-picker-trigger" @click="openPromptPicker">
              <template v-if="selectedPromptInfo">
                <div class="prompt-picker-trigger-main">
                  <el-tooltip :content="selectedPromptInfo.name" placement="top" :show-after="200">
                    <span class="prompt-picker-trigger-name">{{ selectedPromptInfo.name }}</span>
                  </el-tooltip>
                  <el-tag size="small" type="info">{{ selectedPromptInfo.category }}</el-tag>
                </div>
                <el-tooltip
                  :content="selectedPromptInfo.description || '点击重新选择拆书库提示词'"
                  placement="top"
                  :show-after="200"
                >
                  <div class="prompt-picker-trigger-desc">
                  {{ selectedPromptInfo.description || '' }}
                  </div>
                </el-tooltip>
              </template>
              <template v-else>
                <div class="prompt-picker-trigger-placeholder">选择提示词</div>
                <div class="prompt-picker-trigger-desc"></div>
              </template>
            </div>
            <el-button
              v-if="selectedPromptInfo"
              type="primary"
              link
              size="small"
              class="view-intro-btn"
              @click="showPromptIntro"
            >
              <el-icon><View /></el-icon>
              查看介绍
            </el-button>
          </div>

          <div class="config-item">
            <div class="config-label">
              <el-icon><EditPen /></el-icon>
              补充信息
            </div>
            <el-input
              v-model="additionalInfo"
              placeholder="输入额外的补充信息..."
              class="single-line-input"
            />
          </div>

          <div class="config-item">
            <div class="config-label">
              <el-icon><Cpu /></el-icon>
              AI模型选择
            </div>
            <el-select
              v-model="selectedModelId"
              placeholder="请选择AI模型"
              class="full-width"
              @click="openModelDialog"
            >
              <el-option
                v-for="model in models"
                :key="model.id"
                :label="model.name"
                :value="model.id"
              >
                <div class="model-option">
                  <span class="model-name">{{ model.name }}</span>
                  <el-tag v-if="model.provider_name" size="small" type="info">{{ model.provider_name }}</el-tag>
                </div>
              </el-option>
            </el-select>
          </div>

          <div class="config-item">
            <div class="config-label">
              <el-icon><Setting /></el-icon>
              拆分方式
            </div>
            <el-radio-group v-model="splitMode" class="split-mode-group">
              <el-radio value="chapter">分章拆</el-radio>
              <el-radio value="merge">合并拆</el-radio>
            </el-radio-group>
          </div>
        </div>
      </div>

      <div class="right-panel">
        <div class="preview-header">
          <el-icon><View /></el-icon>
          <span>预览区域</span>
          <div class="preview-actions" v-if="previewContent">
            <el-button type="primary" link size="small" @click="copyPreview">
              <el-icon><CopyDocument /></el-icon>
              复制
            </el-button>
            <el-button type="primary" link size="small" @click="clearPreview">
              <el-icon><Delete /></el-icon>
              清空
            </el-button>
          </div>
        </div>
        <div class="preview-content">
          <el-scrollbar v-if="previewContent">
            <MarkdownRenderer :content="previewContent" class="content-text" />
          </el-scrollbar>
          <div v-else class="empty-preview">
            <el-icon :size="64"><Document /></el-icon>
            <p>拆书结果将显示在这里</p>
            <p class="hint">请先选择章节和提示词，然后点击"开始拆书"</p>
          </div>
        </div>
      </div>
    </div>

    <el-dialog
      v-model="promptIntroDialogVisible"
      title="提示词介绍"
      width="600px"
      destroy-on-close
      class="prompt-intro-dialog"
    >
      <div class="prompt-intro-content">
        <h3>{{ selectedPromptInfo?.name }}</h3>
        <el-tag size="small" type="info" style="margin-bottom: 16px;">{{ selectedPromptInfo?.category }}</el-tag>
        <div class="intro-text">
          {{ selectedPromptInfo?.content }}
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="promptIntroDialogVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="promptPickerVisible"
      title="选择拆书提示词"
      width="50vw"
      top="18vh"
      append-to-body
      destroy-on-close
      class="prompt-picker-dialog"
    >
      <div class="prompt-picker-dialog-body">
        <div class="prompt-picker-categories">
          <button
            v-for="category in promptPickerCategories"
            :key="category"
            type="button"
            class="prompt-picker-category"
            :class="{ active: activePromptCategory === category }"
            @click="activePromptCategory = category"
          >
            {{ category }}
          </button>
        </div>

        <div class="prompt-picker-list">
          <div
            v-for="prompt in promptPickerPrompts"
            :key="prompt.id"
            class="prompt-picker-card"
            :class="{ active: selectedPromptId === prompt.id }"
            @click="selectPromptFromPicker(prompt)"
          >
            <div class="prompt-picker-card-header">
              <span class="prompt-picker-card-name">{{ prompt.name }}</span>
              <el-tag size="small" type="info">{{ prompt.category }}</el-tag>
            </div>
            <div class="prompt-picker-card-description">
              {{ prompt.description || '暂无介绍' }}
            </div>
            <div class="prompt-picker-card-preview">
              {{ prompt.content }}
            </div>
          </div>

          <div v-if="promptPickerPrompts.length === 0" class="prompt-picker-empty">
            当前卡包下暂无提示词
          </div>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="modelDialogVisible"
      title="选择 AI 模型"
      width="500px"
      destroy-on-close
    >
      <div class="model-dialog-content">
        <div
          v-for="model in models"
          :key="model.id"
          class="model-item"
          :class="{ active: selectedModelId === model.id }"
          @click="selectModel(model)"
        >
          <div class="model-info">
            <span class="model-name">{{ model.name }}</span>
            <el-tag v-if="model.provider_name" size="small" type="info">{{ model.provider_name }}</el-tag>
          </div>
          <el-icon v-if="selectedModelId === model.id" class="check-icon"><Check /></el-icon>
        </div>
      </div>
      <template #footer>
        <el-button @click="modelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmModel">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="showHistoryDialog"
      title="拆书历史记录"
      width="800px"
      destroy-on-close
    >
      <div class="history-dialog-content">
        <div class="history-header">
          <span class="history-count">共 {{ historyRecords.length }} 条记录</span>
          <div class="history-actions">
            <el-button 
              type="danger" 
              size="small" 
              @click="clearHistory"
              :disabled="historyRecords.length === 0"
            >
              <el-icon><Delete /></el-icon>
              清空历史
            </el-button>
          </div>
        </div>
        <div class="history-list">
          <div
            v-for="(record, index) in historyRecords"
            :key="record.id"
            class="history-item"
            @click="viewHistoryDetail(record)"
          >
            <div class="history-item-header">
              <span class="history-prompt">{{ record.promptName }}</span>
              <span class="history-time">{{ formatTimestamp(record.timestamp) }}</span>
            </div>
            <div class="history-item-body">
              <div class="history-meta">
                <el-tag size="small" type="info">{{ record.splitMode === 'chapter' ? '分章拆' : '合并拆' }}</el-tag>
                <span class="history-chapters">选择了 {{ record.selectedChapters.length }} 个章节</span>
              </div>
              <div class="history-preview">
                {{ record.previewContent?.substring(0, 100) || '无预览内容' }}...
              </div>
            </div>
            <div class="history-item-actions">
              <el-button 
                size="small" 
                type="primary" 
                @click.stop="viewHistoryDetail(record)"
              >
                <el-icon><View /></el-icon>
                查看对话
              </el-button>
              <el-button 
                size="small" 
                type="danger" 
                @click.stop="deleteHistoryRecord(index)"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
          <div v-if="historyRecords.length === 0" class="empty-history">
            <el-icon :size="48"><Document /></el-icon>
            <p>暂无历史记录</p>
            <p class="hint">开始第一次拆书后，历史记录会显示在这里</p>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button type="primary" @click="showHistoryDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="historyDialogVisible"
      :title="selectedHistoryRecord?.promptName || '对话详情'"
      width="900px"
      class="history-detail-dialog"
      destroy-on-close
    >
      <div class="history-detail-content" v-if="selectedHistoryRecord">
        <div class="conversation-messages">
          <template v-for="(msg, msgIndex) in selectedHistoryRecord.messages" :key="msgIndex">
            <div
              v-if="msg.role === 'user' || msg.role === 'assistant' || msg.role === 'system'"
              :class="['message-bubble', msg.role]"
            >
              <div class="message-sender">
                <el-icon class="sender-icon">
                  <User v-if="msg.role === 'user'" />
                  <ChatDotRound v-if="msg.role === 'assistant'" />
                  <Setting v-if="msg.role === 'system'" />
                </el-icon>
                <span class="sender-name">
                  {{ msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '提示词' }}
                </span>
                <span class="message-time">{{ formatTimestamp(msg.timestamp) }}</span>
              </div>
              <div class="message-content">
                <MarkdownRenderer :content="msg.content" />
              </div>
            </div>
          </template>
        </div>
      </div>
      <template #footer>
        <el-button @click="historyDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useBookStore } from '@/stores/book'
import { chapterAPI, configAPI } from '@/api'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import type { Chapter, Prompt } from '@/types'

const route = useRoute()
const router = useRouter()
const bookStore = useBookStore()

const bookId = parseInt(route.params.bookId as string)

const chapters = ref<Chapter[]>([])
const selectedChapters = ref<number[]>([])
const selectAll = ref(false)
const isIndeterminate = ref(false)

const prompts = ref<Prompt[]>([])
const selectedPromptId = ref<number>()
const promptPickerVisible = ref(false)
const promptLibraryCategories = ref<string[]>([])
const activePromptCategory = ref('默认')
const additionalInfo = ref('')
const splitMode = ref<'chapter' | 'merge'>('chapter')

const models = ref<Array<{ id: number; name: string; model: string; provider_name?: string; is_default?: number | boolean }>>([])
const selectedModelId = ref<number>()
const modelDialogVisible = ref(false)

const promptIntroDialogVisible = ref(false)
const showHistoryDialog = ref(false)
const historyDialogVisible = ref(false)

const analyzing = ref(false)
const previewContent = ref('')
const analysisAbortController = ref<AbortController | null>(null)
const analysisStopped = ref(false)

interface HistoryMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
  timestamp: number
}

interface HistoryRecord {
  id: number
  bookId: number
  bookTitle: string
  promptName: string
  splitMode: 'chapter' | 'merge'
  selectedChapters: number[]
  additionalInfo: string
  messages: HistoryMessage[]
  previewContent: string
  timestamp: number
}

const historyRecords = ref<HistoryRecord[]>([])
const selectedHistoryRecord = ref<HistoryRecord | null>(null)

const selectedPromptInfo = computed(() => {
  return prompts.value.find(p => p.id === selectedPromptId.value)
})

const promptPickerCategories = computed(() => {
  const promptCategories = prompts.value
    .map(prompt => prompt.category?.trim() || '')
    .filter(category => category && category !== '未分类')

  const categories = Array.from(new Set(['默认', ...promptLibraryCategories.value, ...promptCategories]))
  return categories
})

const promptPickerPrompts = computed(() => {
  return prompts.value
    .filter(prompt => prompt.category === activePromptCategory.value)
    .sort((a, b) => {
      if (a.order_num !== b.order_num) {
        return a.order_num - b.order_num
      }

      return a.id - b.id
    })
})

onMounted(async () => {
  await bookStore.fetchBook(bookId)
  await fetchChapters()
  await fetchPrompts()
  await fetchModels()
  loadHistoryRecords()
})

const loadHistoryRecords = () => {
  const stored = localStorage.getItem('book-analysis-history')
  if (stored) {
    try {
      const records = JSON.parse(stored) as HistoryRecord[]
      historyRecords.value = records
        .filter((r: HistoryRecord) => r.bookId === bookId)
        .sort((a, b) => b.timestamp - a.timestamp)
    } catch (e) {
      console.error('Failed to parse history records:', e)
      historyRecords.value = []
    }
  }
}

const saveHistoryRecord = (record: HistoryRecord) => {
  const allRecords = JSON.parse(localStorage.getItem('book-analysis-history') || '[]')
  allRecords.push(record)
  localStorage.setItem('book-analysis-history', JSON.stringify(allRecords))
  loadHistoryRecords()
}

const deleteHistoryRecord = (index: number) => {
  const recordToDelete = historyRecords.value[index]
  if (!recordToDelete) {
    return
  }
  historyRecords.value.splice(index, 1)
  const allRecords = JSON.parse(localStorage.getItem('book-analysis-history') || '[]')
  const filteredRecords = allRecords.filter((r: HistoryRecord) => 
    !(r.bookId === bookId && r.timestamp === recordToDelete.timestamp)
  )
  localStorage.setItem('book-analysis-history', JSON.stringify(filteredRecords))
}

const clearHistory = () => {
  const allRecords = JSON.parse(localStorage.getItem('book-analysis-history') || '[]')
  const filteredRecords = allRecords.filter((r: HistoryRecord) => r.bookId !== bookId)
  localStorage.setItem('book-analysis-history', JSON.stringify(filteredRecords))
  loadHistoryRecords()
  ElMessage.success('历史记录已清空')
}

const fetchChapters = async () => {
  const res = await chapterAPI.getByBook(bookId)
  if (res.success && res.data) {
    chapters.value = res.data.filter(c => c.type === 'chapter')
  }
}

const fetchPrompts = async () => {
  const stored = localStorage.getItem('book-analysis-prompts_prompts')
  const storedCategories = localStorage.getItem('book-analysis-prompts_categories')

  if (storedCategories) {
    try {
      const parsedCategories = JSON.parse(storedCategories)
      promptLibraryCategories.value = Array.isArray(parsedCategories)
        ? parsedCategories
            .map((item: unknown) => String(item).trim())
            .filter((item: string) => item && item !== '未分类')
        : []
    } catch (e) {
      console.error('Failed to parse stored categories:', e)
      promptLibraryCategories.value = []
    }
  } else {
    promptLibraryCategories.value = []
  }

  if (stored) {
    try {
      const allPrompts = JSON.parse(stored)
      prompts.value = Array.isArray(allPrompts)
        ? allPrompts
            .filter((p: any) => p && typeof p.name === 'string' && typeof p.content === 'string')
            .map((p: any, index: number) => ({
              ...p,
              id: Number(p.id) || index + 1,
              category: (typeof p.category === 'string' && p.category.trim()) ? p.category.trim() : '未分类',
              order_num: Number.isFinite(p.order_num) ? Number(p.order_num) : index,
            }))
        : []
    } catch (e) {
      console.error('Failed to parse stored prompts:', e)
      prompts.value = []
    }
  } else {
    prompts.value = []
  }

  if (!promptPickerCategories.value.includes(activePromptCategory.value)) {
    activePromptCategory.value = promptPickerCategories.value[0] || '默认'
  }

  if (selectedPromptId.value && !prompts.value.some(prompt => prompt.id === selectedPromptId.value)) {
    selectedPromptId.value = undefined
  }
}

const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data.filter(m => m.enabled !== 0)

      const defaultModel = models.value.find(model => model.is_default === 1 || model.is_default === true)
      const currentSelectedModel = models.value.find(model => model.id === selectedModelId.value)

      if (defaultModel) {
        selectedModelId.value = defaultModel.id
      } else if (!currentSelectedModel && models.value.length > 0) {
        selectedModelId.value = models.value[0]?.id
      }
    }
  } catch (error) {
    models.value = []
  }
}

const handleSelectAllChange = (val: boolean) => {
  selectedChapters.value = val ? chapters.value.map(c => c.id) : []
  isIndeterminate.value = false
}

const getWordCount = (content: string) => {
  if (!content) return 0
  const plainText = content.replace(/<[^>]*>/g, '')
  let length = 0
  for (const char of plainText) {
    if (/[\u4e00-\u9fa5]/.test(char)) {
      length += 1
    }
  }
  return length
}

const goBack = () => {
  router.back()
}

const managePrompts = () => {
  router.push('/book-analysis-prompts')
}

const showPromptIntro = () => {
  promptIntroDialogVisible.value = true
}

const openPromptPicker = () => {
  if (!promptPickerCategories.value.includes(activePromptCategory.value)) {
    activePromptCategory.value = promptPickerCategories.value[0] || '默认'
  }
  promptPickerVisible.value = true
}

const selectPromptFromPicker = (prompt: Prompt) => {
  selectedPromptId.value = prompt.id
  promptPickerVisible.value = false
}

const openModelDialog = () => {
  modelDialogVisible.value = true
}

const selectModel = (model: { id: number; name: string; model: string; provider_name?: string }) => {
  selectedModelId.value = model.id
}

const confirmModel = () => {
  modelDialogVisible.value = false
}

const buildAnalysisUserContent = (selectedChapterData: Chapter[]) => {
  if (splitMode.value === 'merge') {
    let userContent = `请分析以下内容：\n\n${selectedChapterData.map(c => `《${c.title}》\n${c.content}`).join('\n\n')}`
    if (additionalInfo.value) {
      userContent += `\n\n补充信息：${additionalInfo.value}`
    }
    return userContent
  }

  let userContent = '请依次分析以下章节：\n\n'
  selectedChapterData.forEach((chapter, index) => {
    userContent += `=== 章节${index + 1}：${chapter.title} ===\n${chapter.content}\n\n`
  })
  if (additionalInfo.value) {
    userContent += `补充信息：${additionalInfo.value}`
  }
  return userContent
}

const saveAnalysisHistory = (
  selectedPromptData: Prompt,
  selectedChapterData: Chapter[],
  stopped = false
) => {
  const assistantContent = stopped
    ? `${previewContent.value}${previewContent.value ? '\n\n' : ''}[本次拆书已手动停止]`
    : previewContent.value

  const historyRecord: HistoryRecord = {
    id: Date.now(),
    bookId: bookId,
    bookTitle: bookStore.currentBook?.title || '未知书籍',
    promptName: stopped ? `${selectedPromptData.name}（已停止）` : selectedPromptData.name,
    splitMode: splitMode.value,
    selectedChapters: [...selectedChapters.value],
    additionalInfo: additionalInfo.value,
    messages: [
      {
        role: 'system',
        content: selectedPromptData.content || '',
        timestamp: Date.now()
      },
      {
        role: 'user',
        content: buildAnalysisUserContent(selectedChapterData),
        timestamp: Date.now()
      },
      {
        role: 'assistant',
        content: assistantContent,
        timestamp: Date.now()
      }
    ],
    previewContent: assistantContent,
    timestamp: Date.now()
  }

  saveHistoryRecord(historyRecord)
}

const stopAnalysis = () => {
  if (!analyzing.value) {
    return
  }

  analysisStopped.value = true
  analysisAbortController.value?.abort()
}

const startAnalysis = async () => {
  if (selectedChapters.value.length === 0) {
    ElMessage.warning('请至少选择一个章节')
    return
  }
  
  if (!selectedPromptId.value) {
    ElMessage.warning('请选择提示词')
    return
  }
  
  if (!selectedModelId.value) {
    ElMessage.warning('请选择AI模型')
    return
  }
  
  analyzing.value = true
  analysisStopped.value = false
  analysisAbortController.value = new AbortController()
  previewContent.value = ''
  
  try {
    const selectedPromptData = prompts.value.find(p => p.id === selectedPromptId.value)
    const selectedModelData = models.value.find(m => m.id === selectedModelId.value)
    
    if (!selectedPromptData || !selectedModelData) {
      throw new Error('未找到提示词或模型')
    }
    
    const selectedChapterData = chapters.value.filter(c => selectedChapters.value.includes(c.id))
    
    let contentToAnalyze = ''
    if (splitMode.value === 'merge') {
      contentToAnalyze = selectedChapterData.map(c => `【${c.title}】\n${c.content}`).join('\n\n')
    }
    
    const messages: Array<{ role: string; content: string }> = [
      {
        role: 'system',
        content: selectedPromptData.content
      }
    ]
    
    if (splitMode.value === 'merge') {
      let userContent = `请分析以下内容：\n\n${contentToAnalyze}`
      if (additionalInfo.value) {
        userContent += `\n\n补充信息：${additionalInfo.value}`
      }
      messages.push({
        role: 'user',
        content: userContent
      })
    } else {
      let userContent = '请依次分析以下章节：\n\n'
      selectedChapterData.forEach((chapter, index) => {
        userContent += `=== 章节${index + 1}：${chapter.title} ===\n${chapter.content}\n\n`
      })
      if (additionalInfo.value) {
        userContent += `补充信息：${additionalInfo.value}`
      }
      messages.push({
        role: 'user',
        content: userContent
      })
    }
    
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      signal: analysisAbortController.value.signal,
      body: JSON.stringify({
        messages,
        configId: selectedModelData.id
      })
    })
    
    if (!response.ok) {
      throw new Error('请求失败')
    }
    
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    
    if (reader) {
      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        if (analysisStopped.value) break
        
        const chunk = decoder.decode(value)
        const lines = chunk.split('\n')
        
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6)
            if (data === '[DONE]') continue
            
            try {
              const json = JSON.parse(data)
              
              if (json.error) {
                throw new Error(json.error)
              }
              
              const content = json.content || ''
              if (content && !analysisStopped.value) {
                previewContent.value += content
              }
            } catch (e: any) {
              if (e.message) {
                throw e
              }
            }
          }
        }
      }
    }
    
    ElMessage.success('拆书完成')
    
    if (analysisStopped.value) {
      saveAnalysisHistory(selectedPromptData, selectedChapterData, true)
      ElMessage.info('已停止拆书')
      return
    }

    const historyRecord: HistoryRecord = {
      id: Date.now(),
      bookId: bookId,
      bookTitle: bookStore.currentBook?.title || '未知书籍',
      promptName: selectedPromptData?.name || '未知提示词',
      splitMode: splitMode.value,
      selectedChapters: [...selectedChapters.value],
      additionalInfo: additionalInfo.value,
      messages: [
        {
          role: 'system',
          content: selectedPromptData?.content || '',
          timestamp: Date.now()
        },
        {
          role: 'user',
          content: splitMode.value === 'merge' 
            ? `请分析以下内容：\n\n${selectedChapterData.map(c => `【${c.title}】\n${c.content}`).join('\n\n')}`
            : `请依次分析以下章节：\n\n${selectedChapterData.map((c, i) => `=== 章节${i + 1}：${c.title} ===\n${c.content}`).join('\n\n')}`,
          timestamp: Date.now()
        },
        {
          role: 'assistant',
          content: previewContent.value,
          timestamp: Date.now()
        }
      ],
      previewContent: previewContent.value,
      timestamp: Date.now()
    }
    
    saveHistoryRecord(historyRecord)
  } catch (error: any) {
    if (analysisStopped.value || error?.name === 'AbortError') {
      const selectedPromptData = prompts.value.find(p => p.id === selectedPromptId.value)
      const selectedChapterData = chapters.value.filter(c => selectedChapters.value.includes(c.id))
      if (selectedPromptData) {
        saveAnalysisHistory(selectedPromptData, selectedChapterData, true)
      }
      ElMessage.info('已停止拆书')
      return
    }
    ElMessage.error(error.message || '拆书失败')
  } finally {
    analyzing.value = false
    analysisAbortController.value = null
  }
}

const copyPreview = async () => {
  try {
    await navigator.clipboard.writeText(previewContent.value)
    ElMessage.success('已复制到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}

const clearPreview = () => {
  previewContent.value = ''
}

const formatTimestamp = (timestamp: number) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const viewHistoryDetail = (record: HistoryRecord) => {
  selectedHistoryRecord.value = record
  historyDialogVisible.value = true
}
</script>

<style scoped>
.book-analysis-page {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f0f23 100%);
  display: flex;
  flex-direction: column;
  z-index: 10000;
}

.page-header {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  background: rgba(0, 0, 0, 0.3);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.back-btn {
  background: transparent !important;
  border: 1px solid rgba(255, 255, 255, 0.3) !important;
  color: #fff !important;
  margin-right: 16px;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1) !important;
  border-color: rgba(255, 255, 255, 0.5) !important;
}

.page-title {
  flex: 1;
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.main-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.left-panel {
  width: 20%;
  display: flex;
  flex-direction: column;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
}

.chapter-select-area {
  flex: 1.2;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-top: 0;
  padding-left: 16px;
  padding-right: 16px;
  padding-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  padding-bottom: 8px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 8px;
}

.select-all-checkbox {
  margin-left: auto;
}

.chapter-list {
  flex: 1;
  overflow-y: auto;
}

.chapter-item {
  padding: 6px 12px;
  margin-bottom: 4px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  transition: all 0.3s ease;
}

.chapter-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.chapter-item :deep(.el-checkbox__label) {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #fff;
}

.chapter-title {
  font-size: 14px;
}

.chapter-words {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

.empty-chapters {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: rgba(255, 255, 255, 0.5);
}

.config-area {
  flex: 1;
  padding: 10px;
  background: rgba(0, 0, 0, 0.2);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.config-item {
  margin-bottom: 10px;
}

.config-item:last-child {
  margin-bottom: 0;
}

.config-label {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 13px;
  font-weight: 500;
  margin-bottom: 4px;
}

.config-desc-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
  color: rgba(255, 255, 255, 0.6);
  font-size: 11px;
  cursor: pointer;
  padding: 4px 8px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
  transition: all 0.3s ease;
  width: fit-content;
}

.config-desc-trigger:hover {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.9);
}

.config-desc {
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  line-height: 1.5;
  padding: 6px 10px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 4px;
  border-left: 2px solid #52c41a;
}

.full-width {
  width: 100%;
}

.single-line-input :deep(.el-input__wrapper) {
  min-height: 36px;
}

.prompt-picker-trigger {
  padding: 6px 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.14);
  cursor: pointer;
  transition: all 0.25s ease;
}

.prompt-picker-trigger:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.24);
}

.prompt-picker-trigger-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-width: 0;
}

.prompt-picker-trigger-name {
  flex: 1;
  min-width: 0;
  color: #fff;
  font-size: 14px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prompt-picker-trigger-placeholder {
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  font-weight: 500;
}

.prompt-picker-trigger-desc {
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: rgba(255, 255, 255, 0.6);
  font-size: 12px;
  line-height: 1.5;
}

.view-intro-btn {
  margin-top: 8px;
}

.prompt-picker-dialog :deep(.el-dialog) {
  position: fixed;
  margin: 0 auto;
  border-radius: 18px;
  overflow: hidden;
}

.prompt-picker-dialog :deep(.el-dialog__body) {
  padding-top: 12px;
}

.prompt-picker-dialog-body {
  height: 50vh;
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.prompt-picker-categories {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  padding-bottom: 12px;
  border-bottom: 1px solid #eef2f7;
}

.prompt-picker-category {
  padding: 8px 16px;
  border-radius: 999px;
  border: 1px solid #dbe5f0;
  background: #f8fafc;
  color: #475569;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.prompt-picker-category:hover,
.prompt-picker-category.active {
  background: #1677ff;
  border-color: #1677ff;
  color: #fff;
}

.prompt-picker-list {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 14px;
  align-content: start;
  padding-right: 4px;
}

.prompt-picker-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 180px;
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  cursor: pointer;
  transition: all 0.22s ease;
}

.prompt-picker-card:hover,
.prompt-picker-card.active {
  border-color: #93c5fd;
  box-shadow: 0 10px 24px rgba(59, 130, 246, 0.14);
  transform: translateY(-2px);
}

.prompt-picker-card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.prompt-picker-card-name {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.prompt-picker-card-description {
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.prompt-picker-card-preview {
  color: #334155;
  font-size: 12px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 5;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.prompt-picker-empty {
  grid-column: 1 / -1;
  min-height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 16px;
  border: 1px dashed #cbd5e1;
  color: #64748b;
  background: #f8fafc;
}

.split-mode-group {
  width: 100%;
  display: flex;
}

.split-mode-group :deep(.el-radio) {
  flex: 1;
  margin-right: 0;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  padding: 10px;
  border-radius: 8px;
  margin-right: 8px;
}

.split-mode-group :deep(.el-radio:last-child) {
  margin-right: 0;
}

.split-mode-group :deep(.el-radio.is-checked) {
  background: rgba(82, 196, 26, 0.2);
}

.split-mode-group :deep(.el-radio__label) {
  color: #fff;
}

.right-panel {
  width: 80%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.preview-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 24px;
  color: #fff;
  font-size: 16px;
  font-weight: 500;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.preview-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

.preview-content {
  flex: 1;
  overflow: hidden;
  padding: 16px;
}

.content-text {
  color: rgba(255, 255, 255, 0.9);
  font-size: 15px;
  line-height: 1.8;
  padding: 16px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  min-height: 100%;
}

.content-text :deep(h1),
.content-text :deep(h2),
.content-text :deep(h3),
.content-text :deep(h4),
.content-text :deep(h5),
.content-text :deep(h6) {
  color: #fff;
  border-bottom-color: rgba(255, 255, 255, 0.2);
}

.content-text :deep(p) {
  color: rgba(255, 255, 255, 0.85);
}

.content-text :deep(code) {
  background: rgba(255, 255, 255, 0.1);
  color: #e6e6e6;
}

.content-text :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.content-text :deep(pre code) {
  color: #e6e6e6;
}

.content-text :deep(blockquote) {
  color: rgba(255, 255, 255, 0.7);
  border-left-color: #52c41a;
}

.content-text :deep(table th) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.content-text :deep(table td) {
  border-color: rgba(255, 255, 255, 0.15);
}

.content-text :deep(a) {
  color: #52c41a;
}

.content-text :deep(ul),
.content-text :deep(ol) {
  color: rgba(255, 255, 255, 0.85);
}

.empty-preview {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.4);
}

.empty-preview p {
  margin-top: 16px;
  font-size: 16px;
}

.empty-preview .hint {
  font-size: 14px;
  margin-top: 8px;
  color: rgba(255, 255, 255, 0.3);
}

.prompt-intro-content h3 {
  color: #1f2937;
  margin-bottom: 8px;
}

.prompt-intro-dialog :deep(.el-dialog) {
  max-height: 72vh;
  display: flex;
  flex-direction: column;
}

.prompt-intro-dialog :deep(.el-dialog__body) {
  flex: 1;
  min-height: 0;
}

.prompt-intro-content {
  max-height: 52vh;
  overflow-y: auto;
  padding-right: 6px;
}

.intro-text {
  color: #666;
  font-size: 14px;
  line-height: 1.8;
  padding: 16px;
  background: #f5f5f5;
  border-radius: 8px;
  white-space: pre-wrap;
}

.model-dialog-content {
  max-height: 400px;
  overflow-y: auto;
}

.model-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.model-item:hover {
  border-color: #52c41a;
  background: rgba(82, 196, 26, 0.05);
}

.model-item.active {
  border-color: #52c41a;
  background: rgba(82, 196, 26, 0.1);
}

.model-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.model-name {
  font-size: 14px;
  color: #1f2937;
}

.check-icon {
  color: #52c41a;
  font-size: 18px;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: none;
}

:deep(.el-input__wrapper:hover) {
  border-color: rgba(255, 255, 255, 0.4);
}

:deep(.el-input__inner) {
  color: #fff;
}

:deep(.el-input__inner::placeholder) {
  color: rgba(255, 255, 255, 0.4);
}

:deep(.el-textarea__inner) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;
  resize: none;
}

:deep(.el-textarea__inner:hover) {
  border-color: rgba(255, 255, 255, 0.4);
}

:deep(.el-textarea__inner::placeholder) {
  color: rgba(255, 255, 255, 0.4);
}

:deep(.el-checkbox__label) {
  color: #fff;
}

:deep(.el-select-dropdown__item) {
  display: flex;
  align-items: center;
  gap: 8px;
}

.prompt-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.prompt-category {
  font-size: 12px;
  color: #999;
}

.model-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.history-dialog-content {
  max-height: 600px;
  display: flex;
  flex-direction: column;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 8px;
  margin-bottom: 16px;
}

.history-count {
  font-size: 14px;
  color: #666;
}

.history-actions {
  display: flex;
  gap: 8px;
}

.history-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.history-item {
  padding: 16px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ebef 100%);
  border-radius: 12px;
  border: 1px solid #e4e7ed;
  cursor: pointer;
  transition: all 0.3s ease;
}

.history-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  border-color: #409eff;
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.history-prompt {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.history-time {
  font-size: 13px;
  color: #909399;
}

.history-item-body {
  margin-bottom: 12px;
}

.history-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.history-chapters {
  font-size: 13px;
  color: #606266;
}

.history-preview {
  font-size: 13px;
  color: #909399;
  line-height: 1.6;
  background: rgba(255, 255, 255, 0.5);
  padding: 8px 12px;
  border-radius: 6px;
  border-left: 3px solid #409eff;
}

.history-item-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.empty-history {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  color: #909399;
}

.empty-history p {
  margin-top: 16px;
  font-size: 15px;
}

.empty-history .hint {
  font-size: 13px;
  color: #c0c4cc;
  margin-top: 8px;
}

.history-detail-dialog {
  :deep(.el-dialog__body) {
    padding: 0;
  }
}

.history-detail-content {
  max-height: 70vh;
  overflow-y: auto;
}

.conversation-messages {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  background: #f0f2f5;
  min-height: 400px;
  max-height: 65vh;
  overflow-y: auto;
}

.message-bubble {
  display: flex;
  flex-direction: column;
  max-width: 70%;
  animation: messageSlideIn 0.3s ease;
}

@keyframes messageSlideIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message-bubble.user {
  align-self: flex-end;
  margin-left: auto;
}

.message-bubble.assistant {
  align-self: flex-start;
  margin-right: auto;
}

.message-bubble.system {
  align-self: center;
  max-width: 85%;
  background: linear-gradient(135deg, #f6f8fa 0%, #eaecef 100%);
  border: 1px dashed #d0d7de;
  border-radius: 12px;
  padding: 12px 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

.message-bubble.system .sender-icon {
  color: #6a737d;
}

.message-bubble.system .sender-name {
  color: #24292e;
}

.message-bubble.system .message-content {
  background: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
  padding: 12px;
  font-size: 13px;
  line-height: 1.7;
  color: #57606a;
  max-height: 300px;
  overflow-y: auto;
}

.message-bubble.system .message-content :deep(p) {
  color: #57606a;
}

.message-bubble.system .message-content :deep(code) {
  background: rgba(209, 215, 220, 0.5);
  color: #24292e;
}

.message-bubble.system .message-content :deep(pre) {
  background: #f6f8fa;
  border: 1px solid #eaecef;
}

.message-bubble.system .message-content :deep(pre code) {
  color: #24292e;
}

.message-sender {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  padding: 0 4px;
}

.message-bubble.user .message-sender {
  justify-content: flex-end;
}

.message-bubble.assistant .message-sender {
  justify-content: flex-start;
}

.sender-icon {
  font-size: 16px;
}

.message-bubble.user .sender-icon {
  color: rgba(255, 255, 255, 0.9);
}

.message-bubble.assistant .sender-icon {
  color: #1890ff;
}

.sender-name {
  font-size: 13px;
  font-weight: 600;
}

.message-bubble.user .sender-name {
  color: rgba(255, 255, 255, 0.9);
}

.message-bubble.assistant .sender-name {
  color: #303133;
}

.message-time {
  font-size: 11px;
}

.message-bubble.user .message-time {
  color: rgba(255, 255, 255, 0.7);
}

.message-bubble.assistant .message-time {
  color: #909399;
}

.message-bubble .message-content {
  padding: 12px 16px;
  border-radius: 16px;
  font-size: 14px;
  line-height: 1.6;
  word-wrap: break-word;
}

.message-bubble.user .message-content {
  background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
  color: #fff;
  border-bottom-right-radius: 4px;
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.3);
}

.message-bubble.assistant .message-content {
  background: #fff;
  color: #303133;
  border-bottom-left-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.message-bubble.user .message-content :deep(p) {
  color: #fff;
}

.message-bubble.user .message-content :deep(h1),
.message-bubble.user .message-content :deep(h2),
.message-bubble.user .message-content :deep(h3),
.message-bubble.user .message-content :deep(h4),
.message-bubble.user .message-content :deep(h5),
.message-bubble.user .message-content :deep(h6) {
  color: #fff;
  border-bottom-color: rgba(255, 255, 255, 0.3);
}

.message-bubble.user .message-content :deep(code) {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.message-bubble.user .message-content :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.message-bubble.user .message-content :deep(pre code) {
  color: #fff;
}

.message-bubble.user .message-content :deep(blockquote) {
  border-left-color: rgba(255, 255, 255, 0.5);
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.9);
}

.message-bubble.assistant .message-content :deep(p) {
  color: #303133;
}

.message-bubble.assistant .message-content :deep(h1),
.message-bubble.assistant .message-content :deep(h2),
.message-bubble.assistant .message-content :deep(h3),
.message-bubble.assistant .message-content :deep(h4),
.message-bubble.assistant .message-content :deep(h5),
.message-bubble.assistant .message-content :deep(h6) {
  color: #1f2937;
  border-bottom-color: #e5e7eb;
}

.message-bubble.assistant .message-content :deep(code) {
  background: #f3f4f6;
  color: #1f2937;
}

.message-bubble.assistant .message-content :deep(pre) {
  background: #282c34;
  border: 1px solid #e5e7eb;
}

.message-bubble.assistant .message-content :deep(pre code) {
  color: #abb2bf;
}

.message-bubble.assistant .message-content :deep(blockquote) {
  border-left-color: #1890ff;
  background: rgba(24, 144, 255, 0.1);
  color: #606266;
}

.message-bubble .message-content :deep(p) {
  margin: 8px 0;
}

.message-bubble .message-content :deep(pre) {
  border-radius: 8px;
  padding: 12px;
  overflow-x: auto;
  margin: 12px 0;
}

.message-bubble .message-content :deep(code) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.message-bubble .message-content :deep(blockquote) {
  padding-left: 12px;
  margin: 12px 0;
  padding: 8px 12px;
  border-radius: 4px;
}
</style>
