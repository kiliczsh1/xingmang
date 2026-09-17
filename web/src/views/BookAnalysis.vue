<template>
  <div class="book-analysis-page">
    <!-- 极简顶栏：返回 + 标题 + 状态胶囊 + 历史 + 配置 -->
    <div class="topbar">
      <button class="back-btn" @click="goBack">← 返回</button>
      <div class="title">拆书库</div>
      <div class="status-capsule">
        <span class="dot"></span>
        已选 <b>{{ selectedChapters.length }}</b> 章 · <span>{{ statusPromptText }}</span>
      </div>
      <button class="hist-btn" @click="showHistoryDialog = true">🕘 历史 <b>{{ historyRecords.length }}</b></button>
      <button class="gear-btn" @click="toggleDrawer">⚙ 配置</button>
    </div>

    <!-- 主画布：结果为主角 -->
    <div class="canvas">
      <div v-if="previewContent" class="preview-tools">
        <button class="tool-btn" @click="copyPreview">⧉ 复制</button>
        <button class="tool-btn" @click="clearPreview">🗑 清空</button>
      </div>
      <div class="paper">
        <div v-if="!previewContent" class="empty-state">
          <div class="ic">📄</div>
          <h2>在这里开始一场拆书</h2>
          <p>点击右下角「开始拆书」，或先点右上角「配置」选择章节与提示词。</p>
        </div>
        <el-scrollbar v-else ref="scrollbarRef" class="paper-scroll">
          <div class="paper-inner">
            <MarkdownRenderer :content="previewContent" class="content-text" />
          </div>
        </el-scrollbar>
      </div>

      <button
        class="fab"
        :class="{ stop: analyzing }"
        :disabled="!canStart && !analyzing"
        @click="startOrStop"
      >
        {{ analyzing ? '◼ 停止拆书' : '▶ 开始拆书' }}
      </button>
    </div>

    <!-- 遮罩 + 左侧配置抽屉 -->
    <div class="mask" :class="{ show: maskVisible }" @click="closeDrawer"></div>
    <aside class="drawer" :class="{ show: drawerVisible }">
      <div class="drawer-head"><h3>拆书配置</h3><button class="x" @click="closeDrawer">✕</button></div>
      <div class="drawer-body">
        <!-- 1 选择章节 -->
        <div class="acc acc-chapters" :class="{ open: accOpen[1] }">
          <div class="acc-head" @click="accOpen[1] = !accOpen[1]">
            <span class="num">1</span> 选择章节
            <span class="badge">{{ selectedChapters.length }} 章</span>
            <span class="chev">▸</span>
          </div>
          <div class="acc-body">
            <div class="sel-row">
              <el-checkbox
                v-model="selectAll"
                :indeterminate="isIndeterminate"
                @change="handleSelectAllChange"
              >
                全选
              </el-checkbox>
            </div>
            <div class="chapter-list">
              <div
                v-for="chapter in chapters"
                :key="chapter.id"
                class="chapter-item"
                :class="{ on: selectedChapters.includes(chapter.id) }"
                @click="toggleChapter(chapter.id)"
              >
                <span class="ck">{{ selectedChapters.includes(chapter.id) ? '✓' : '' }}</span>
                <span class="ct">{{ chapter.title }}</span>
                <span class="w">{{ getWordCount(chapter.content).toLocaleString() }}字</span>
              </div>
              <div v-if="chapters.length === 0" class="empty-chapters">
                <p>暂无章节</p>
              </div>
            </div>
          </div>
        </div>

        <!-- 2 选择提示词（触发弹窗） -->
        <div class="acc" :class="{ open: accOpen[2] }">
          <div class="acc-head" @click="accOpen[2] = !accOpen[2]">
            <span class="num">2</span> 选择提示词
            <span class="badge">{{ selectedPromptInfo ? selectedPromptInfo.name : '未选' }}</span>
            <span class="chev">▸</span>
          </div>
          <div class="acc-body">
            <div class="trigger" @click="openPromptPicker">
              <template v-if="selectedPromptInfo">
                <div class="t-main">
                  <div class="t-name">{{ selectedPromptInfo.name }}<span class="p-cat">{{ categoryLabel(selectedPromptInfo.category) }}</span></div>
                  <div class="t-desc">{{ selectedPromptInfo.description || '点击重新选择拆书库提示词' }}</div>
                </div>
                <span class="t-open">更换 →</span>
              </template>
              <template v-else>
                <div class="t-main">
                  <div class="t-name">选择提示词</div>
                  <div class="t-desc">从提示词库中选择一种拆书方式</div>
                </div>
                <span class="t-open">选择 →</span>
              </template>
            </div>
            <button v-if="selectedPromptInfo" class="mini-link" @click="showPromptIntro">查看完整介绍 →</button>
          </div>
        </div>

        <!-- 3 模型 · 方式 · 补充信息 -->
        <div class="acc" :class="{ open: accOpen[3] }">
          <div class="acc-head" @click="accOpen[3] = !accOpen[3]">
            <span class="num">3</span> 模型 / 方式 / 补充
            <span class="chev">▸</span>
          </div>
          <div class="acc-body">
            <label class="lbl">AI 模型</label>
            <div class="trigger" @click="openModelDialog">
              <div class="t-main">
                <div class="t-name">{{ selectedModelLabel }}</div>
                <div class="t-desc">{{ selectedModelProvider }}</div>
              </div>
              <span class="t-open">选择 →</span>
            </div>
            <label class="lbl">拆分方式</label>
            <div class="seg">
              <button :class="{ on: splitMode === 'chapter' }" @click="splitMode = 'chapter'">分章拆</button>
              <button :class="{ on: splitMode === 'merge' }" @click="splitMode = 'merge'">合并拆</button>
            </div>
            <label class="lbl">补充信息（可选）</label>
            <el-input
              v-model="additionalInfo"
              type="textarea"
              :rows="2"
              placeholder="例如：重点分析第三章的论证逻辑…"
            />
          </div>
        </div>
      </div>
    </aside>

    <!-- 提示词介绍弹窗 -->
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

    <!-- 提示词选择弹窗（触发弹窗，分类卡片式） -->
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
            {{ category.replace('拆书-', '') }}
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
              <el-tag size="small" type="info">{{ prompt.category.replace('拆书-', '') }}</el-tag>
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

    <!-- 模型选择弹窗 -->
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

    <!-- 历史记录弹窗 -->
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

    <!-- 对话详情弹窗 -->
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
import { ref, computed, reactive, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useBookStore } from '@/stores/book'
import { chapterAPI, configAPI, promptAPI } from '@/api'
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

/* ---- 抽屉（Demo 03） ---- */
const drawerVisible = ref(true) // 默认打开配置抽屉
const maskVisible = ref(false) // 初始无遮罩，不遮挡画布与 FAB
const accOpen = reactive({ 1: true, 2: true, 3: true })
const scrollbarRef = ref()
const openDrawer = () => {
  drawerVisible.value = true
  maskVisible.value = true
}
const closeDrawer = () => {
  drawerVisible.value = false
  maskVisible.value = false
}
const toggleDrawer = () => {
  drawerVisible.value ? closeDrawer() : openDrawer()
}

const canStart = computed(() => {
  return selectedChapters.value.length > 0 && !!selectedPromptId.value && !!selectedModelId.value
})

const statusPromptText = computed(() => {
  return selectedPromptInfo.value ? '提示词：' + selectedPromptInfo.value.name : '未选提示词'
})

const selectedModelInfo = computed(() => models.value.find(m => m.id === selectedModelId.value))
const selectedModelLabel = computed(() => selectedModelInfo.value?.name || '未选择')
const selectedModelProvider = computed(() => selectedModelInfo.value?.provider_name || '点击选择 AI 模型')

const categoryLabel = (cat?: string) => (cat || '').replace('拆书-', '')

const toggleChapter = (id: number) => {
  const idx = selectedChapters.value.indexOf(id)
  if (idx >= 0) {
    selectedChapters.value.splice(idx, 1)
  } else {
    selectedChapters.value.push(id)
  }
}

watch(selectedChapters, list => {
  const total = chapters.value.length
  if (total === 0) {
    selectAll.value = false
    isIndeterminate.value = false
    return
  }
  const valid = list.filter(id => chapters.value.some(c => c.id === id))
  const n = valid.length
  selectAll.value = n === total
  isIndeterminate.value = n > 0 && n < total
})

/* 流式输出自动滚动到底部 */
watch(previewContent, async () => {
  await nextTick()
  const sb = scrollbarRef.value
  if (sb && sb.wrapRef) {
    sb.setScrollTop(sb.wrapRef.scrollHeight)
  }
})

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
    .filter(category => category && category !== '拆书-未分类')

  const categories = Array.from(new Set([...promptLibraryCategories.value, ...promptCategories]))
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

// 拆书库页面本身是暗色设计，需要避免全局暗色主题覆盖干扰
const savedTheme = ref<string | null>(null)

onMounted(async () => {
  const root = document.documentElement
  if (root.getAttribute('data-theme') === 'dark') {
    savedTheme.value = 'dark'
    root.removeAttribute('data-theme')
  }
  await bookStore.fetchBook(bookId)
  await fetchChapters()
  await fetchPrompts()
  await fetchModels()
  loadHistoryRecords()
})

onUnmounted(() => {
  if (savedTheme.value === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark')
  }
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
  try {
    const res = await promptAPI.getAll()
    if (res.success && res.data) {
      // 只加载拆书分类的提示词（category 以 "拆书-" 开头）
      prompts.value = res.data.filter((p: Prompt) => {
        const category = p.category || ''
        return category.startsWith('拆书-')
      }).map((p: Prompt, index: number) => ({
        ...p,
        id: Number(p.id) || index + 1,
        category: (typeof p.category === 'string' && p.category.trim()) ? p.category.trim() : '拆书-未分类',
        order_num: Number.isFinite(p.order_num) ? Number(p.order_num) : index,
      }))

      // 从提示词数据中提取分类（去掉"拆书-"前缀显示）
      const categories = Array.from(new Set(
        prompts.value
          .map((p: Prompt) => p.category)
          .filter((c: string) => c && c !== '拆书-未分类')
      ))
      promptLibraryCategories.value = categories
    }
  } catch (e) {
    console.error('Failed to fetch prompts from API:', e)
    prompts.value = []
    promptLibraryCategories.value = []
  }

  if (!promptPickerCategories.value.includes(activePromptCategory.value)) {
    activePromptCategory.value = promptPickerCategories.value[0] || '拆书-默认'
  }

  if (selectedPromptId.value && !prompts.value.some(prompt => prompt.id === selectedPromptId.value)) {
    selectedPromptId.value = undefined
  }
}

const fetchModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      models.value = res.data

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
  return content.replace(/\s/g, '').length
}

const goBack = () => {
  router.back()
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

  closeDrawer()
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

const startOrStop = () => {
  if (analyzing.value) {
    stopAnalysis()
  } else {
    startAnalysis()
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
  --accent: #4ade80;
  --accent-2: #22c55e;
  --accent-soft: rgba(74, 222, 128, 0.14);
  --danger: #f87171;
  --blue: #60a5fa;
  --border: rgba(255, 255, 255, 0.1);
  --panel: rgba(255, 255, 255, 0.05);
  --panel-2: rgba(255, 255, 255, 0.09);
  --text: #eef2f8;
  --text-dim: rgba(238, 242, 248, 0.55);
  --text-faint: rgba(238, 242, 248, 0.32);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: linear-gradient(135deg, #1a1a30 0%, #141428 50%, #0d0d1e 100%);
  display: flex;
  flex-direction: column;
  z-index: 2000;
  font-family: -apple-system, "PingFang SC", "Microsoft YaHei", "Segoe UI", sans-serif;
}

/* ============ 极简顶栏 ============ */
.topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--border);
  background: rgba(0, 0, 0, 0.22);
  flex-shrink: 0;
}

.back-btn {
  padding: 7px 13px;
  border-radius: 9px;
  border: 1px solid rgba(255, 255, 255, 0.25);
  background: transparent;
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  transition: background 0.2s;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.title {
  font-weight: 700;
  font-size: 17px;
  color: var(--text);
}

.status-capsule {
  margin-left: auto;
  display: flex;
  gap: 6px;
  align-items: center;
  padding: 7px 14px;
  border-radius: 99px;
  background: var(--panel);
  border: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-dim);
}

.status-capsule b {
  color: var(--accent);
  font-weight: 700;
}

.status-capsule .dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent);
  animation: pulse 1.6s infinite;
}

@keyframes pulse {
  50% { opacity: 0.3; }
}

.hist-btn,
.gear-btn {
  padding: 7px 13px;
  border-radius: 9px;
  border: 1px solid var(--border);
  background: var(--panel);
  color: var(--text);
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.2s;
}

.hist-btn:hover,
.gear-btn:hover {
  border-color: var(--accent);
}

.hist-btn b {
  color: var(--accent);
  font-variant-numeric: tabular-nums;
}

/* ============ 主画布 ============ */
.canvas {
  flex: 1;
  position: relative;
  min-height: 0;
}

.paper {
  max-width: 900px;
  height: 100%;
  margin: 0 auto;
  padding: 40px 48px;
  overflow: hidden;
}

.paper-scroll {
  height: 100%;
}

.paper-inner {
  padding-right: 8px;
}

.empty-state {
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  color: var(--text-faint);
  text-align: center;
}

.empty-state .ic {
  font-size: 56px;
  opacity: 0.45;
}

.empty-state h2 {
  color: var(--text-dim);
  font-size: 20px;
  font-weight: 600;
}

.empty-state p {
  font-size: 13px;
}

/* 画布右上角悬浮工具（复制/清空） */
.preview-tools {
  position: absolute;
  top: 16px;
  right: 24px;
  z-index: 6;
  display: flex;
  gap: 8px;
}

.tool-btn {
  padding: 6px 12px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: rgba(0, 0, 0, 0.35);
  color: var(--text-dim);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
  transition: all 0.15s;
}

.tool-btn:hover {
  color: var(--text);
  border-color: var(--accent);
}

/* Markdown 结果（暗色） */
.content-text {
  color: var(--text);
  font-size: 15px;
  line-height: 1.9;
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
  color: rgba(255, 255, 255, 0.88);
}

.content-text :deep(code) {
  background: rgba(255, 255, 255, 0.1);
  color: #e6e6e6;
}

.content-text :deep(pre) {
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}

.content-text :deep(pre code) {
  color: #e6e6e6;
}

.content-text :deep(blockquote) {
  color: rgba(255, 255, 255, 0.7);
  border-left-color: var(--accent);
}

.content-text :deep(table th) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.content-text :deep(table td) {
  border-color: rgba(255, 255, 255, 0.15);
}

.content-text :deep(a) {
  color: var(--accent);
}

/* 悬浮开始按钮 */
.fab {
  position: absolute;
  right: 32px;
  bottom: 32px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px 28px;
  border-radius: 99px;
  border: none;
  background: linear-gradient(135deg, var(--accent), var(--accent-2));
  color: #0b1220;
  font-size: 15px;
  font-weight: 700;
  font-family: inherit;
  box-shadow: 0 10px 30px rgba(74, 222, 128, 0.35);
  transition: all 0.25s;
  z-index: 5;
  cursor: pointer;
}

.fab:hover {
  transform: translateY(-2px) scale(1.02);
  box-shadow: 0 14px 36px rgba(74, 222, 128, 0.45);
}

.fab:disabled {
  opacity: 0.45;
  transform: none;
  cursor: not-allowed;
}

.fab.stop {
  background: rgba(248, 113, 113, 0.18);
  color: var(--danger);
  border: 1px solid rgba(248, 113, 113, 0.4);
  box-shadow: none;
}

/* ============ 抽屉遮罩 + 左抽屉 ============ */
.mask {
  position: fixed;
  inset: 0;
  background: rgba(5, 5, 12, 0.55);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.25s;
  z-index: 40;
}

.mask.show {
  opacity: 1;
  pointer-events: auto;
}

.drawer {
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  width: min(420px, 92vw);
  background: #191932;
  border-right: 1px solid var(--border);
  transform: translateX(-102%);
  transition: transform 0.28s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 41;
  display: flex;
  flex-direction: column;
}

.drawer.show {
  transform: none;
}

.drawer-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 20px;
  border-bottom: 1px solid var(--border);
}

.drawer-head h3 {
  font-size: 16px;
  flex: 1;
  color: var(--text);
}

.drawer-head .x {
  border: none;
  background: none;
  color: var(--text-dim);
  font-size: 18px;
  cursor: pointer;
}

.drawer-head .x:hover {
  color: var(--text);
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-gutter: stable;
  padding: 14px 18px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.drawer-body::-webkit-scrollbar,
.chapter-list::-webkit-scrollbar {
  width: 6px;
}

.drawer-body::-webkit-scrollbar-thumb,
.chapter-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.16);
  border-radius: 3px;
}

.drawer-body::-webkit-scrollbar-thumb:hover,
.chapter-list::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* 折叠分区 */
.acc {
  border: 1px solid var(--border);
  border-radius: 13px;
  overflow: hidden;
  background: var(--panel);
}

.acc-head {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  cursor: pointer;
  user-select: none;
  font-size: 14px;
  font-weight: 600;
  color: var(--text);
  transition: color 0.2s;
}

.acc-head .num {
  width: 22px;
  height: 22px;
  border-radius: 7px;
  background: var(--panel-2);
  display: grid;
  place-items: center;
  font-size: 12px;
  color: var(--text-dim);
}

.acc-head .badge {
  margin-left: auto;
  font-size: 11px;
  color: var(--accent);
  font-weight: 600;
  max-width: 140px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.acc-head .chev {
  color: var(--text-faint);
  transition: transform 0.2s;
}

.acc.open .acc-head .chev {
  transform: rotate(90deg);
}

.acc.open .acc-head {
  color: var(--accent);
}

.acc-body {
  display: none;
  padding: 4px 14px 16px;
}

.acc.open .acc-body {
  display: block;
  animation: accIn 0.2s ease;
}

@keyframes accIn {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: none; }
}

.sel-row {
  padding: 6px 0;
}

/* 章节（自适应抽屉剩余高度） */
.acc-chapters.open {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.acc-chapters.open .acc-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.chapter-list {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.chapter-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 10px;
  border-radius: 9px;
  cursor: pointer;
  font-size: 13px;
  transition: 0.14s;
}

.chapter-item:hover {
  background: var(--panel-2);
}

.chapter-item.on {
  background: var(--accent-soft);
  color: var(--accent);
}

.chapter-item .ck {
  width: 18px;
  height: 18px;
  border-radius: 5px;
  border: 1.5px solid rgba(255, 255, 255, 0.35);
  display: grid;
  place-items: center;
  font-size: 11px;
  flex-shrink: 0;
  color: #0b1220;
}

.chapter-item.on .ck {
  background: var(--accent-2);
  border-color: var(--accent-2);
}

.chapter-item .ct {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text);
}

.chapter-item .w {
  margin-left: auto;
  font-size: 11px;
  color: var(--text-faint);
  white-space: nowrap;
}

.empty-chapters {
  padding: 30px 0;
  text-align: center;
  color: var(--text-faint);
  font-size: 13px;
}

/* 触发卡 */
.trigger {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 15px;
  border-radius: 12px;
  background: var(--panel);
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all 0.16s;
}

.trigger:hover {
  border-color: var(--accent);
  background: var(--accent-soft);
}

.trigger .t-main {
  flex: 1;
  min-width: 0;
}

.trigger .t-name {
  font-size: 14px;
  font-weight: 700;
  color: var(--text);
  display: flex;
  align-items: center;
  gap: 8px;
}

.trigger .p-cat {
  font-size: 10px;
  color: var(--accent);
  border: 1px solid var(--accent);
  border-radius: 99px;
  padding: 1px 8px;
  white-space: nowrap;
}

.trigger .t-desc {
  font-size: 11.5px;
  color: var(--text-dim);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.trigger .t-open {
  font-size: 12px;
  color: var(--accent);
  font-weight: 700;
  white-space: nowrap;
}

.mini-link {
  margin-top: 10px;
  font-size: 11px;
  color: var(--blue);
  background: none;
  border: none;
  padding: 0;
  cursor: pointer;
  text-decoration: underline;
}

.mini-link:hover {
  color: var(--accent);
}

.lbl {
  font-size: 12px;
  color: var(--text-dim);
  margin: 10px 0 6px;
  display: block;
}

.seg {
  display: flex;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 3px;
  gap: 3px;
}

.seg button {
  flex: 1;
  padding: 8px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--text-dim);
  font-size: 12px;
  cursor: pointer;
  font-family: inherit;
}

.seg button.on {
  background: var(--accent-2);
  color: #0b1220;
  font-weight: 700;
}

/* 抽屉内 Element 组件暗色适配 */
.drawer :deep(.el-checkbox__label) {
  color: var(--text);
  font-size: 12.5px;
}

.drawer :deep(.el-checkbox.is-checked .el-checkbox__label) {
  color: var(--accent);
}

.drawer :deep(.el-textarea__inner) {
  background: var(--panel);
  border: 1px solid var(--border);
  color: var(--text);
  border-radius: 9px;
  font-family: inherit;
}

.drawer :deep(.el-textarea__inner:hover) {
  border-color: rgba(255, 255, 255, 0.3);
}

.drawer :deep(.el-textarea__inner:focus) {
  border-color: var(--accent);
}

.drawer :deep(.el-textarea__inner::placeholder) {
  color: var(--text-faint);
}

.paper :deep(.el-scrollbar__bar.is-vertical) {
  width: 6px;
}

/* ============ 提示词介绍弹窗 ============ */
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

/* ============ 提示词选择弹窗 ============ */
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
  line-clamp: 2;
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
  line-clamp: 5;
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

/* ============ 模型选择弹窗 ============ */
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

/* ============ 历史记录弹窗 ============ */
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

/* ============ 对话详情弹窗 ============ */
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

@media (max-width: 768px) {
  .status-capsule {
    display: none;
  }

  .paper {
    padding: 24px 20px;
  }

  .fab {
    right: 18px;
    bottom: 18px;
    padding: 13px 22px;
  }
}
</style>
