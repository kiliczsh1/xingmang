<template>
  <el-dialog
    v-model="dialogVisible"
    title="历史记录中心"
    width="1200px"
    top="5vh"
    destroy-on-close
    class="history-center-dialog"
    modal-class="history-center-dialog-modal"
  >
    <div class="history-center-content">
      <div class="history-header">
        <div class="header-left">
          <el-icon class="header-icon"><Clock /></el-icon>
          <span class="history-count">
            共 {{ filteredRecords.length }} 条记录
          </span>
        </div>
        <div class="header-filters">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索标题、预览或作品名"
            clearable
            size="small"
            class="search-input"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select
            v-model="activeTab"
            placeholder="选择分类"
            size="small"
            class="category-select"
          >
            <el-option label="全部分类" value="all"></el-option>
            <el-option label="创意工坊" value="creative"></el-option>
            <el-option label="拆书库" value="bookAnalysis"></el-option>
            <el-option label="正文 AI" value="textEditor"></el-option>
            <el-option label="工作流" value="workflow"></el-option>
          </el-select>
          <el-select
            v-model="selectedBookTitle"
            placeholder="选择作品"
            size="small"
            clearable
            class="book-select"
          >
            <el-option label="全部作品" value=""></el-option>
            <el-option
              v-for="book in availableBooks"
              :key="book"
              :label="book"
              :value="book"
            ></el-option>
          </el-select>
        </div>
        <div class="history-actions">
          <el-dropdown
            trigger="click"
            @command="handleMoreAction"
            class="more-actions-dropdown"
          >
            <el-button size="small" class="more-actions-btn" :disabled="totalRecords === 0">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  command="clearCurrent"
                  :disabled="filteredRecords.length === 0"
                >
                  <el-icon><Delete /></el-icon>
                  <span>清空当前分类</span>
                </el-dropdown-item>
                <el-dropdown-item
                  command="clearAll"
                  :disabled="totalRecords === 0"
                  divided
                >
                  <el-icon><Delete /></el-icon>
                  <span>清空全部</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>

      <div class="history-list">
        <div
          v-for="(record, index) in paginatedRecords"
          :key="record.id"
          class="history-item"
          @click="viewHistoryDetail(record)"
        >
          <div class="history-item-card">
            <div class="history-item-header">
              <div class="header-left-group">
                <div class="history-source-badge" :class="record.sourceType">
                  <el-icon><component :is="record.sourceIcon" /></el-icon>
                  <span>{{ record.sourceName }}</span>
                </div>
                <span class="history-time">{{ formatFullTimestamp(record.timestamp) }}</span>
              </div>
              <div class="header-right-group">
                <span class="add-to-label">添加到</span>
                <el-tooltip content="复制" placement="top">
                  <el-button
                    size="small"
                    link
                    @click.stop="copyRecordContent(record)"
                    class="icon-action-btn"
                  >
                    <el-icon><CopyDocument /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="继续对话" placement="top">
                  <el-button
                    size="small"
                    link
                    @click.stop="continueConversation(record)"
                    class="icon-action-btn"
                  >
                    <el-icon><ChatDotRound /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="查看详情" placement="top">
                  <el-button
                    size="small"
                    link
                    @click.stop="viewHistoryDetail(record)"
                    class="icon-action-btn"
                  >
                    <el-icon><Document /></el-icon>
                  </el-button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button
                    size="small"
                    link
                    @click.stop="deleteHistoryRecord(index, record)"
                    class="icon-action-btn"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </el-tooltip>
              </div>
            </div>

            <div class="history-item-body">
              <div v-if="record.metadata?.modelName" class="model-tag">
                <el-icon><Setting /></el-icon>
                <span>模型 {{ record.metadata.modelName }}</span>
              </div>

              <div v-if="record.metadata?.nodeTitle || record.title" class="history-item-title">
                <span class="title-text">{{ record.metadata?.nodeTitle || record.title }}</span>
              </div>

              <div class="history-item-preview">
                <pre class="preview-code"><code>{{ record.preview }}</code></pre>
              </div>
            </div>

            <div class="history-item-footer">
              <span class="log-id" :title="`#${record.id}`" @click.stop="copyToClipboard(`#${record.id}`, '日志ID已复制')">#{{ record.id }}</span>
              <div class="footer-stats">
                <span :class="['stat-consumed', { 'is-zero': (record.metadata?.inputTokens || 0) === 0 }]">
                  消耗{{ record.metadata?.inputTokens || 0 }}字
                </span>
                <span class="stat-generated">
                  生成{{ getContentLength(record.preview || '') }}字
                </span>
                <span class="stat-detail" @click.stop="viewHistoryDetail(record)">
                  详细信息 <el-icon class="detail-arrow"><ArrowRight /></el-icon>
                </span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="filteredRecords.length === 0" class="empty-history">
          <div class="empty-state-icon">
            <el-icon :size="64"><Document /></el-icon>
          </div>
          <p class="empty-title">
            {{ searchKeyword || selectedBookTitle ? '未找到匹配的记录' : '暂无历史记录' }}
          </p>
          <p class="empty-hint">
            <template v-if="searchKeyword || selectedBookTitle">
              请尝试更换搜索关键词或选择其他作品
            </template>
            <template v-else-if="activeTab === 'all'">
              开始使用后，历史记录会显示在这里
            </template>
            <template v-else>
              当前分类下暂无历史记录，请切换到其他分类或选择"全部分类"查看所有记录
            </template>
          </p>
        </div>

        <!-- 分页组件 -->
        <div v-if="filteredRecords.length > pageSize" class="history-pagination">
          <el-pagination
            v-model:current-page="currentPage"
            :page-size="pageSize"
            :total="filteredRecords.length"
            layout="prev, pager, next"
            :pager-count="5"
            small
          />
        </div>
      </div>
    </div>


  </el-dialog>

  <!-- 历史记录详情对话框 -->
  <el-dialog
    v-model="detailDialogVisible"
    :title="selectedRecord?.title || '历史记录详情'"
    width="900px"
    class="history-detail-dialog"
    modal-class="history-detail-dialog-modal"
    destroy-on-close
  >
    <div class="history-detail-content" v-if="selectedRecord">
      <div class="conversation-messages" v-if="selectedRecord.messages && selectedRecord.messages.length > 0">
        <template v-for="(msg, msgIndex) in selectedRecord.messages" :key="msgIndex">
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
            <div v-if="msg.role === 'user' && selectedRecord.originalRecord?.relatedContext?.length > 0 && msgIndex === selectedRecord.messages.findIndex(m => m.role === 'user')" class="message-related-context">
              <div class="related-context-inline">
                <el-icon class="related-icon"><Connection /></el-icon>
                <span class="related-label">关联内容</span>
              </div>
              <div class="related-context-list">
                <div v-for="(item, idx) in selectedRecord.originalRecord.relatedContext" :key="idx" class="related-item">
                  <div class="related-item-header">
                    <el-tag 
                      :type="item.type === 'book' ? 'primary' : item.type === 'chapter' ? 'success' : item.type === 'memo' ? 'warning' : 'info'"
                      size="small"
                    >
                      {{ item.type === 'book' ? '作品' : item.type === 'chapter' ? '章节' : item.type === 'memo' ? '备忘录' : item.type === 'character' ? '角色' : item.type }}
                    </el-tag>
                    <span class="related-item-title">{{ item.title }}</span>
                  </div>
                  <div v-if="item.content" class="related-item-content">
                    <MarkdownRenderer :content="item.content" />
                  </div>
                </div>
              </div>
            </div>
            <div class="message-content">
              <MarkdownRenderer :content="msg.content" />
            </div>
          </div>
        </template>
      </div>

      <div v-else class="detail-body">
        <div class="detail-section">
          <label class="detail-label">预览内容</label>
          <div class="detail-preview">
            <MarkdownRenderer :content="selectedRecord.preview" />
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="detailDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>

  <!-- 追问弹窗 -->
  <el-dialog
    v-model="continueDialogVisible"
    width="800px"
    class="continue-dialog"
    modal-class="continue-dialog-modal"
    destroy-on-close
  >
    <template #header>
      <div class="continue-dialog-header">
        <div class="header-title">
          <el-icon><ChatDotRound /></el-icon>
          <span>继续对话</span>
        </div>
        <el-select
          v-model="selectedModelId"
          placeholder="选择模型"
          class="model-select"
          size="small"
        >
          <el-option
            v-for="model in availableModels"
            :key="model.id"
            :label="model.name"
            :value="model.id"
          >
            <div class="model-option-content">
              <span>{{ model.name }}</span>
              <el-tag v-if="model.provider_name" size="small" type="info">{{ model.provider_name }}</el-tag>
            </div>
          </el-option>
        </el-select>
      </div>
    </template>
    <div class="continue-dialog-content">
      <div class="continue-messages" ref="continueMessagesContainer">
        <div v-if="continueMessages.length > 2 && !isHistoryExpanded" class="history-collapsed-hint">
          <el-button 
            text 
            type="primary" 
            @click="isHistoryExpanded = true"
            size="small"
            class="expand-btn"
          >
            <el-icon><ArrowDown /></el-icon>
            展开 {{ continueMessages.length - 2 }} 条历史消息
          </el-button>
        </div>
        
        <template v-if="isHistoryExpanded">
          <transition-group name="message-slide">
            <div
              v-for="(msg, msgIndex) in continueMessages"
              :key="msgIndex"
              :class="['continue-message-bubble', msg.role, { 'history-message': msgIndex < continueMessages.length - 2 }]"
            >
              <div class="continue-message-sender">
                <el-icon class="sender-icon">
                  <User v-if="msg.role === 'user'" />
                  <ChatDotRound v-if="msg.role === 'assistant'" />
                  <Setting v-if="msg.role === 'system'" />
                </el-icon>
                <span class="sender-name">
                  {{ msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '系统' }}
                </span>
              </div>
              <div v-if="msg.role === 'user' && currentContinueRecord?.originalRecord?.relatedContext?.length > 0 && msgIndex === continueMessages.findIndex(m => m.role === 'user')" class="message-related-context">
                <div class="related-context-inline">
                  <el-icon class="related-icon"><Connection /></el-icon>
                  <span class="related-label">关联内容</span>
                </div>
                <div class="related-context-list">
                  <div v-for="(item, idx) in currentContinueRecord?.originalRecord?.relatedContext || []" :key="idx" class="related-item">
                    <div class="related-item-header">
                      <el-tag 
                        :type="item.type === 'book' ? 'primary' : item.type === 'chapter' ? 'success' : item.type === 'memo' ? 'warning' : 'info'"
                        size="small"
                      >
                        {{ item.type === 'book' ? '作品' : item.type === 'chapter' ? '章节' : item.type === 'memo' ? '备忘录' : item.type === 'character' ? '角色' : item.type }}
                      </el-tag>
                      <span class="related-item-title">{{ item.title }}</span>
                    </div>
                    <div v-if="item.content" class="related-item-content">
                      <MarkdownRenderer :content="item.content" />
                    </div>
                  </div>
                </div>
              </div>
              <div class="continue-message-content">
                <MarkdownRenderer :content="msg.content" />
              </div>
            </div>
          </transition-group>
        </template>
        
        <template v-else>
          <transition-group name="message-slide">
            <div
              v-for="(msg, msgIndex) in continueMessages.slice(-2)"
              :key="continueMessages.length - 2 + msgIndex"
              :class="['continue-message-bubble', msg.role]"
            >
              <div class="continue-message-sender">
                <el-icon class="sender-icon">
                  <User v-if="msg.role === 'user'" />
                  <ChatDotRound v-if="msg.role === 'assistant'" />
                  <Setting v-if="msg.role === 'system'" />
                </el-icon>
                <span class="sender-name">
                  {{ msg.role === 'user' ? '用户' : msg.role === 'assistant' ? 'AI' : '系统' }}
                </span>
              </div>
              <div v-if="msg.role === 'user' && currentContinueRecord?.originalRecord?.relatedContext?.length > 0 && msgIndex === continueMessages.slice(-2).findIndex(m => m.role === 'user')" class="message-related-context">
                <div class="related-context-inline">
                  <el-icon class="related-icon"><Connection /></el-icon>
                  <span class="related-label">关联内容</span>
                  <el-button text type="primary" size="small" @click="isHistoryExpanded = true" class="inline-expand-btn">
                    <el-icon><ArrowDown /></el-icon>
                    展开完整内容
                  </el-button>
                </div>
                <div class="related-context-list">
                  <div v-for="(item, idx) in currentContinueRecord?.originalRecord?.relatedContext || []" :key="idx" class="related-item">
                    <div class="related-item-header">
                      <el-tag 
                        :type="item.type === 'book' ? 'primary' : item.type === 'chapter' ? 'success' : item.type === 'memo' ? 'warning' : 'info'"
                        size="small"
                      >
                        {{ item.type === 'book' ? '作品' : item.type === 'chapter' ? '章节' : item.type === 'memo' ? '备忘录' : item.type === 'character' ? '角色' : item.type }}
                      </el-tag>
                      <span class="related-item-title">{{ item.title }}</span>
                    </div>
                    <div v-if="item.content" class="related-item-content">
                      <MarkdownRenderer :content="item.content" />
                    </div>
                  </div>
                </div>
              </div>
              <div class="continue-message-content">
                <MarkdownRenderer :content="msg.content" />
              </div>
            </div>
          </transition-group>
        </template>
        
        <div v-if="isGenerating" class="continue-message-bubble assistant">
          <div class="continue-message-sender">
            <el-icon class="sender-icon"><ChatDotRound /></el-icon>
            <span class="sender-name">AI</span>
          </div>
          <div class="continue-message-content">
            <div class="typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
          </div>
        </div>
      </div>

      <div class="continue-input-area" :class="{ 'is-fullscreen': isFullscreen }">
        <el-input
          v-model="userInput"
          type="textarea"
          :rows="isFullscreen ? 12 : 2"
          placeholder="输入您的问题，按 Ctrl+Enter 发送..."
          @keydown.enter.ctrl="sendMessage"
          :disabled="isGenerating"
          resize="none"
          class="continue-input"
        />
        <div class="continue-input-actions">
          <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏编辑'" placement="top">
            <el-button
              @click="toggleFullscreen"
              :icon="FullScreen"
              circle
              class="fullscreen-btn"
            />
          </el-tooltip>
          <el-tooltip content="发送 (Ctrl+Enter)" placement="top">
            <el-button 
              type="primary" 
              @click="sendMessage"
              :loading="isGenerating"
              :disabled="!userInput.trim()"
              :icon="ChatDotRound"
              circle
              class="send-btn"
            />
          </el-tooltip>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { aiAPI, configAPI } from '@/api'
import {
  Clock,
  Delete,
  Document,
  Star,
  ChatDotSquare,
  Folder,
  User,
  ChatDotRound,
  Setting,
  Connection,
  ArrowDown,
  CopyDocument,
  ArrowRight,
  Search,
  MoreFilled,
  FullScreen
} from '@element-plus/icons-vue'
import MarkdownRenderer from './MarkdownRenderer.vue'

interface HistoryMessage {
  role: 'user' | 'assistant' | 'system' | 'prompt'
  content: string
  timestamp: number
}

interface UnifiedHistoryRecord {
  id: number | string
  sourceType: 'creative' | 'bookAnalysis' | 'textEditor' | 'workflow'
  sourceName: string
  sourceIcon: any
  itemIcon?: any
  title: string
  preview: string
  timestamp: number | string
  messages?: HistoryMessage[]
  metadata?: {
    generatorName?: string
    promptName?: string
    splitMode?: 'chapter' | 'merge'
    selectedChapters?: number[]
    promptCount?: number
    bookTitle?: string
    source?: 'chat' | 'continue' | 'creative2'
    sourceLabel?: string
    status?: string
    workflowName?: string
    nodeId?: string
    nodeTitle?: string
    level?: 'info' | 'success' | 'error' | 'warning'
    modelName?: string
    inputTokens?: number
    outputTokens?: number
  }
  originalRecord?: any
}

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'continue-conversation', record: UnifiedHistoryRecord): void
}>()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const activeTab = ref<'all' | 'creative' | 'bookAnalysis' | 'textEditor' | 'workflow'>('all')
const currentPage = ref(1)
const pageSize = 6
const searchKeyword = ref('')
const selectedBookTitle = ref<string>('')
const detailDialogVisible = ref(false)
const selectedRecord = ref<UnifiedHistoryRecord | null>(null)

const continueDialogVisible = ref(false)
const continueMessages = ref<HistoryMessage[]>([])
const userInput = ref('')
const isGenerating = ref(false)
const isFullscreen = ref(false)
const continueMessagesContainer = ref<HTMLElement | null>(null)
const currentContinueRecord = ref<UnifiedHistoryRecord | null>(null)
const selectedModelId = ref<number | null>(null)
const availableModels = ref<any[]>([])
const isHistoryExpanded = ref(false)

const creativeRecords = ref<UnifiedHistoryRecord[]>([])
const bookAnalysisRecords = ref<UnifiedHistoryRecord[]>([])
const textEditorRecords = ref<UnifiedHistoryRecord[]>([])
const workflowRecords = ref<UnifiedHistoryRecord[]>([])

const baseRecords = computed(() => {
  switch (activeTab.value) {
    case 'creative':
      return creativeRecords.value
    case 'bookAnalysis':
      return bookAnalysisRecords.value
    case 'textEditor':
      return textEditorRecords.value
    case 'workflow':
      return workflowRecords.value
    default:
      return [...creativeRecords.value, ...bookAnalysisRecords.value, ...textEditorRecords.value, ...workflowRecords.value]
        .sort((a, b) => {
          const timeA = typeof a.timestamp === 'string' ? new Date(a.timestamp).getTime() : a.timestamp
          const timeB = typeof b.timestamp === 'string' ? new Date(b.timestamp).getTime() : b.timestamp
          return timeB - timeA
        })
  }
})

// 可选作品列表（去重）
const availableBooks = computed(() => {
  const titles = new Set<string>()
  baseRecords.value.forEach(r => {
    const title = r.metadata?.bookTitle
    if (title && title.trim()) {
      titles.add(title)
    }
  })
  return Array.from(titles).sort()
})

const filteredRecords = computed(() => {
  let list = baseRecords.value
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (keyword) {
    list = list.filter(r => {
      const title = (r.title || '').toLowerCase()
      const preview = (r.preview || '').toLowerCase()
      const bookTitle = (r.metadata?.bookTitle || '').toLowerCase()
      return title.includes(keyword) || preview.includes(keyword) || bookTitle.includes(keyword)
    })
  }
  if (selectedBookTitle.value) {
    list = list.filter(r => r.metadata?.bookTitle === selectedBookTitle.value)
  }
  return list
})

const totalRecords = computed(() => {
  return creativeRecords.value.length + bookAnalysisRecords.value.length + textEditorRecords.value.length + workflowRecords.value.length
})

const paginatedRecords = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return filteredRecords.value.slice(start, end)
})

const totalPages = computed(() => {
  return Math.ceil(filteredRecords.value.length / pageSize)
})

// 切换 tab / 搜索 / 切换作品时重置到第一页
watch(activeTab, () => {
  currentPage.value = 1
})
watch(searchKeyword, () => {
  currentPage.value = 1
})
watch(selectedBookTitle, () => {
  currentPage.value = 1
})

const loadCreativeHistory = () => {
  const stored = localStorage.getItem('creative-history')
  if (stored) {
    try {
      const records = JSON.parse(stored)
      const now = Date.now()
      creativeRecords.value = records
        .filter((record: any) => record != null)
        .map((record: any, idx: number) => ({
          id: record.id || `creative-${record.timestamp || now}-${idx}`,
          sourceType: 'creative' as const,
          sourceName: '创意工坊',
          sourceIcon: Star,
          itemIcon: record.generatorIcon || 'Lightning',
          title: record.generatorName || record.name || '未知生成器',
          preview: record.previewContent?.substring(0, 150) || record.content?.substring(0, 150) || record.result?.substring(0, 150) || '无预览内容',
          timestamp: record.timestamp || record.createdAt || record.time || now - idx,
          messages: record.messages || [],
          metadata: {
            generatorName: record.generatorName,
            promptCount: record.selectedPrompts?.length || record.prompts?.length || 0,
            bookTitle: record.bookTitle,
            inputTokens: record.inputTokens,
            outputTokens: record.outputTokens
          },
          originalRecord: record
        }))
        .sort((a: any, b: any) => {
          const timeA = typeof a.timestamp === 'string' ? new Date(a.timestamp).getTime() : a.timestamp
          const timeB = typeof b.timestamp === 'string' ? new Date(b.timestamp).getTime() : b.timestamp
          return timeB - timeA
        })
    } catch (e) {
      console.error('Failed to load creative history:', e)
      creativeRecords.value = []
    }
  } else {
    creativeRecords.value = []
  }
}

const loadBookAnalysisHistory = () => {
  const stored = localStorage.getItem('book-analysis-history')
  if (stored) {
    try {
      const records = JSON.parse(stored)
      const now = Date.now()
      bookAnalysisRecords.value = records
        .filter((record: any) => record != null)
        .map((record: any, idx: number) => ({
          id: record.id || `bookAnalysis-${record.timestamp || now}-${idx}`,
          sourceType: 'bookAnalysis' as const,
          sourceName: '拆书库',
          sourceIcon: Folder,
          itemIcon: 'Document',
          title: record.promptName || record.name || '未知提示词',
          preview: record.previewContent?.substring(0, 150) || record.content?.substring(0, 150) || record.result?.substring(0, 150) || '无预览内容',
          timestamp: record.timestamp || record.createdAt || record.time || now - idx,
          messages: record.messages || [],
          metadata: {
            promptName: record.promptName,
            splitMode: record.splitMode,
            selectedChapters: record.selectedChapters,
            bookTitle: record.bookTitle,
            inputTokens: record.inputTokens,
            outputTokens: record.outputTokens
          },
          originalRecord: record
        }))
        .sort((a: any, b: any) => {
          const timeA = typeof a.timestamp === 'string' ? new Date(a.timestamp).getTime() : a.timestamp
          const timeB = typeof b.timestamp === 'string' ? new Date(b.timestamp).getTime() : b.timestamp
          return timeB - timeA
        })
    } catch (e) {
      console.error('Failed to load book analysis history:', e)
      bookAnalysisRecords.value = []
    }
  } else {
    bookAnalysisRecords.value = []
  }
}

const loadTextEditorHistory = () => {
  const stored = localStorage.getItem('write-ai-api-history')
  if (stored) {
    try {
      const records = JSON.parse(stored)
      const now = Date.now()
      textEditorRecords.value = records
        .filter((record: any) => record != null)
        .map((record: any, idx: number) => ({
          id: record.id || `textEditor-${record.timestamp || now}-${idx}`,
          sourceType: 'textEditor' as const,
          sourceName: '正文 AI',
          sourceIcon: ChatDotSquare,
          itemIcon: 'Document',
          title: record.promptName || record.name || 'AI 写作',
          preview: record.previewContent?.substring(0, 150) || record.content?.substring(0, 150) || record.result?.substring(0, 150) || record.output?.substring(0, 150) || '无预览内容',
          timestamp: record.timestamp || record.createdAt || record.time || now - idx,
          messages: record.messages || [],
          metadata: {
            promptName: record.promptName,
            promptCount: record.promptCount,
            source: record.source,
            sourceLabel: record.sourceLabel,
            status: record.status,
            bookTitle: record.bookTitle,
            inputTokens: record.inputTokens,
            outputTokens: record.outputTokens
          },
          originalRecord: record
        }))
        .sort((a: any, b: any) => {
          const timeA = typeof a.timestamp === 'string' ? new Date(a.timestamp).getTime() : a.timestamp
          const timeB = typeof b.timestamp === 'string' ? new Date(b.timestamp).getTime() : b.timestamp
          return timeB - timeA
        })
    } catch (e) {
      console.error('Failed to load text editor history:', e)
      textEditorRecords.value = []
    }
  } else {
    textEditorRecords.value = []
  }
}

const loadWorkflowHistory = () => {
  const stored = localStorage.getItem('workflow-logs')
  if (stored) {
    try {
      const records = JSON.parse(stored)
      const now = Date.now()
      workflowRecords.value = records
        .filter((record: any) => record != null)
        .map((record: any, idx: number) => ({
          id: record.id || `workflow-${record.timestamp || now}-${idx}`,
          sourceType: 'workflow' as const,
          sourceName: '工作流',
          sourceIcon: User,
          itemIcon: 'Connection',
          title: record.nodeTitle || record.title || record.name || '工作流节点',
          preview: record.outputContent?.substring(0, 150) || record.previewContent?.substring(0, 150) || record.content?.substring(0, 150) || record.message || '无预览内容',
          timestamp: record.timestamp || record.createdAt || record.time || now - idx,
          messages: record.messages || [],
          metadata: {
            workflowName: record.workflowName,
            nodeId: record.nodeId,
            nodeTitle: record.nodeTitle,
            level: record.level,
            modelName: record.modelName,
            inputTokens: record.inputTokens,
            outputTokens: record.outputTokens,
            bookTitle: record.bookTitle
          },
          originalRecord: record
        }))
        .sort((a: any, b: any) => {
          const timeA = typeof a.timestamp === 'string' ? new Date(a.timestamp).getTime() : a.timestamp
          const timeB = typeof b.timestamp === 'string' ? new Date(b.timestamp).getTime() : b.timestamp
          return timeB - timeA
        })
    } catch (e) {
      console.error('Failed to load workflow history:', e)
      workflowRecords.value = []
    }
  } else {
    workflowRecords.value = []
  }
}

const loadAllHistory = () => {
  loadCreativeHistory()
  loadBookAnalysisHistory()
  loadTextEditorHistory()
  loadWorkflowHistory()
}

const loadAvailableModels = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      availableModels.value = res.data.filter(m => m.enabled !== 0)
      const defaultModel = res.data.find((m: any) => m.is_default === 1)
      if (defaultModel) {
        selectedModelId.value = defaultModel.id
      } else if (res.data.length > 0) {
        selectedModelId.value = res.data[0].id
      }
    }
  } catch (error) {
    console.error('Failed to load models:', error)
  }
}

onMounted(() => {
  loadAvailableModels()
})

watch(dialogVisible, (newValue) => {
  if (newValue) {
    loadAllHistory()
  }
})

const formatTimestamp = (timestamp: number | string) => {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : new Date(timestamp)
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

const formatFullTimestamp = (timestamp: number | string) => {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : new Date(timestamp)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  const d = String(date.getDate()).padStart(2, '0')
  const h = String(date.getHours()).padStart(2, '0')
  const min = String(date.getMinutes()).padStart(2, '0')
  return `${y}/${m}/${d} ${h}:${min}`
}

const copyToClipboard = async (text: string, message = '已复制') => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    ElMessage.success(message)
  } catch (e) {
    ElMessage.error('复制失败')
  }
}

const copyRecordContent = (record: UnifiedHistoryRecord) => {
  const text = record.preview || record.title || `#${record.id}`
  copyToClipboard(text, '内容已复制')
}

const getContentLength = (content: string) => {
  if (!content) return 0
  const plainText = content
    .replace(/<[^>]*>/g, '')
    .replace(/```[\s\S]*?```/g, '')
    .replace(/`[^`]*`/g, '')
    .replace(/\s+/g, '')
  return [...plainText].length
}

const viewHistoryDetail = (record: UnifiedHistoryRecord) => {
  selectedRecord.value = record
  detailDialogVisible.value = true
}

const continueConversation = (record: UnifiedHistoryRecord) => {
  if (!record.messages || record.messages.length === 0) {
    ElMessage.warning('该历史记录没有对话内容，无法继续对话')
    return
  }
  currentContinueRecord.value = record
  continueMessages.value = [...record.messages]
  userInput.value = ''
  isGenerating.value = false
  isHistoryExpanded.value = false
  continueDialogVisible.value = true
  
  nextTick(() => {
    scrollToBottom()
  })
}

const scrollToBottom = () => {
  if (continueMessagesContainer.value) {
    continueMessagesContainer.value.scrollTop = continueMessagesContainer.value.scrollHeight
  }
}

const toggleFullscreen = () => {
  isFullscreen.value = !isFullscreen.value
}

const sendMessage = async () => {
  if (!userInput.value.trim() || isGenerating.value) return
  
  const userMessage: HistoryMessage = {
    role: 'user',
    content: userInput.value.trim(),
    timestamp: Date.now()
  }
  
  continueMessages.value.push(userMessage)
  const currentInput = userInput.value
  userInput.value = ''
  isGenerating.value = true
  
  nextTick(() => {
    scrollToBottom()
  })
  
  try {
    const validRoles = ['system', 'user', 'assistant', 'tool', 'latest_reminder']
    const messages = continueMessages.value
      .filter(msg => validRoles.includes(msg.role) || msg.role === 'prompt')
      .map(msg => ({
        role: msg.role === 'prompt' ? 'system' : msg.role,
        content: msg.content
      }))
    
    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        messages,
        configId: selectedModelId.value
      })
    })
    
    if (!response.ok) {
      throw new Error('请求失败')
    }
    
    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    let assistantMessage = ''
    
    const assistantMsg: HistoryMessage = {
      role: 'assistant',
      content: '',
      timestamp: Date.now()
    }
    continueMessages.value.push(assistantMsg)
    
    while (reader) {
      const { done, value } = await reader.read()
      if (done) break
      
      const chunk = decoder.decode(value)
      const lines = chunk.split('\n')
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') continue
          
          try {
            const parsed = JSON.parse(data)
            if (parsed.content) {
              assistantMessage += parsed.content
              assistantMsg.content = assistantMessage
              nextTick(() => {
                scrollToBottom()
              })
            } else if (parsed.error) {
              throw new Error(parsed.error)
            }
          } catch (e) {
            if (e instanceof SyntaxError) continue
            throw e
          }
        }
      }
    }
    
    saveContinueHistory(assistantMessage)
    
    ElMessage.success('对话完成')
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败')
    continueMessages.value.pop()
    continueMessages.value.pop()
  } finally {
    isGenerating.value = false
  }
}

const saveContinueHistory = (aiResponse: string) => {
  const relatedContext = currentContinueRecord.value?.originalRecord?.relatedContext || []
  
  const historyRecord = {
    id: Date.now(),
    generatorName: '追问对话',
    generatorIcon: 'ChatDotRound',
    fixedPrompt: '',
    selectedPrompts: [],
    inputParams: '',
    additionalInfo: '',
    fieldValues: {},
    relatedContext: relatedContext,
    messages: [...continueMessages.value],
    previewContent: aiResponse,
    timestamp: Date.now()
  }
  
  const allRecords = JSON.parse(localStorage.getItem('creative-history') || '[]')
  allRecords.push(historyRecord)
  localStorage.setItem('creative-history', JSON.stringify(allRecords))
  
  loadCreativeHistory()
}

const closeContinueDialog = () => {
  continueDialogVisible.value = false
  currentContinueRecord.value = null
}

const deleteHistoryRecord = (index: number, record: UnifiedHistoryRecord) => {
  ElMessageBox.confirm(
    `确定要删除这条历史记录吗？`,
    '删除确认',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    if (record.sourceType === 'creative') {
      const stored = localStorage.getItem('creative-history')
      if (stored) {
        const allRecords = JSON.parse(stored)
        const filteredRecords = allRecords.filter((r: any) => r.id !== record.id)
        localStorage.setItem('creative-history', JSON.stringify(filteredRecords))
        loadCreativeHistory()
      }
    } else if (record.sourceType === 'bookAnalysis') {
      const stored = localStorage.getItem('book-analysis-history')
      if (stored) {
        const allRecords = JSON.parse(stored)
        const filteredRecords = allRecords.filter((r: any) => r.id !== record.id)
        localStorage.setItem('book-analysis-history', JSON.stringify(filteredRecords))
        loadBookAnalysisHistory()
      }
    } else if (record.sourceType === 'textEditor') {
      const stored = localStorage.getItem('write-ai-api-history')
      if (stored) {
        const allRecords = JSON.parse(stored)
        const filteredRecords = allRecords.filter((r: any) => r.id !== record.id)
        localStorage.setItem('write-ai-api-history', JSON.stringify(filteredRecords))
        loadTextEditorHistory()
      }
    } else if (record.sourceType === 'workflow') {
      const stored = localStorage.getItem('workflow-logs')
      if (stored) {
        const allRecords = JSON.parse(stored)
        const filteredRecords = allRecords.filter((r: any) => r.id !== record.id)
        localStorage.setItem('workflow-logs', JSON.stringify(filteredRecords))
        loadWorkflowHistory()
      }
    }
    ElMessage.success('历史记录已删除')
  }).catch(() => {})
}

const handleMoreAction = (command: string) => {
  if (command === 'clearCurrent') {
    handleClearHistory()
  } else if (command === 'clearAll') {
    handleClearAllHistory()
  }
}

const handleClearHistory = () => {
  ElMessageBox.confirm(
    `确定要清空当前分类的所有历史记录吗？`,
    '清空确认',
    {
      confirmButtonText: '清空',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    if (activeTab.value === 'creative') {
      localStorage.removeItem('creative-history')
      creativeRecords.value = []
    } else if (activeTab.value === 'bookAnalysis') {
      localStorage.removeItem('book-analysis-history')
      bookAnalysisRecords.value = []
    } else if (activeTab.value === 'textEditor') {
      localStorage.removeItem('write-ai-api-history')
      textEditorRecords.value = []
    } else if (activeTab.value === 'workflow') {
      localStorage.removeItem('workflow-logs')
      workflowRecords.value = []
    }
    ElMessage.success('历史记录已清空')
  }).catch(() => {})
}

const handleClearAllHistory = () => {
  ElMessageBox.confirm(
    `确定要清空所有历史记录吗？此操作不可恢复！`,
    '清空全部确认',
    {
      confirmButtonText: '清空',
      cancelButtonText: '取消',
      type: 'error'
    }
  ).then(() => {
    localStorage.removeItem('creative-history')
    localStorage.removeItem('book-analysis-history')
    localStorage.removeItem('write-ai-api-history')
    localStorage.removeItem('workflow-logs')
    creativeRecords.value = []
    bookAnalysisRecords.value = []
    textEditorRecords.value = []
    workflowRecords.value = []
    ElMessage.success('所有历史记录已清空')
  }).catch(() => {})
}
</script>

<style scoped>
.history-center-dialog :deep(.el-dialog) {
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  margin: 0 auto !important;
  position: relative !important;
  top: auto !important;
  left: auto !important;
  right: auto !important;
  bottom: auto !important;
  max-height: 90vh;
  overflow: hidden;
}

.history-center-dialog :deep(.el-overlay-dialog) {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 20px;
  margin-top: 5vh;
}

.history-center-dialog :deep(.el-dialog__header) {
  padding: 14px 24px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color-page);
  border-radius: 12px 12px 0 0;
}

.history-center-dialog :deep(.el-dialog__title) {
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.history-center-dialog :deep(.el-dialog__body) {
  padding: 0;
  background: var(--el-bg-color);
}

.history-center-dialog :deep(.el-dialog__footer) {
  padding: 12px 24px;
  border-top: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
  border-radius: 0 0 12px 12px;
}

.history-center-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 24px;
}

.history-tabs-component {
  margin-bottom: 0;
}

.history-tabs-component :deep(.el-tabs__header) {
  margin-bottom: 0;
}

.tab-badge {
  display: none;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  border: 1px solid var(--el-border-color-light);
  gap: 16px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.header-tabs {
  flex: 1;
  min-width: 0;
}

.header-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}

.header-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none;
}

.header-tabs :deep(.el-tabs__item) {
  padding: 0 12px;
  font-size: 13px;
}

.header-filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.search-input {
  flex: 1 1 0;
  min-width: 0;
}

.search-input :deep(.el-input__wrapper) {
  border-radius: 6px;
  background: var(--el-bg-color);
  box-shadow: 0 0 0 1px var(--el-border-color-light) inset;
}

.search-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 0 0 1px var(--el-color-primary-light-5) inset;
}

.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--el-color-primary) inset;
}

.search-input :deep(.el-input__inner) {
  font-size: 13px;
  padding-top: 8px;
  padding-bottom: 8px;
}

.search-input :deep(.el-input__wrapper) {
  padding: 4px 11px;
}

.category-select,
.book-select {
  flex: 1 1 0;
  min-width: 0;
}

.category-select :deep(.el-select__wrapper),
.book-select :deep(.el-select__wrapper) {
  border-radius: 6px;
  background: var(--el-bg-color);
  box-shadow: 0 0 0 1px var(--el-border-color-light) inset;
  height: 32px;
}

.category-select :deep(.el-select__wrapper):hover,
.book-select :deep(.el-select__wrapper):hover {
  box-shadow: 0 0 0 1px var(--el-color-primary-light-5) inset;
}

.header-icon {
  color: var(--el-color-primary);
  font-size: 20px;
}

.history-count {
  color: var(--el-text-color-primary);
  font-size: 14px;
  font-weight: 500;
}

.history-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.more-actions-dropdown {
  display: inline-flex;
}

.more-actions-btn {
  padding: 6px 10px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-light);
  color: var(--el-text-color-regular);
  border-radius: 6px;
  transition: all 0.2s ease;
}

.more-actions-btn:hover:not(:disabled) {
  background: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary-light-5);
  color: var(--el-color-primary);
}

.more-actions-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.more-actions-dropdown :deep(.el-dropdown-menu__item) {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 140px;
}

.more-actions-dropdown :deep(.el-dropdown-menu__item .el-icon) {
  font-size: 14px;
}

.more-actions-dropdown :deep(.el-dropdown-menu__item.is-divided) {
  border-top: 1px solid var(--el-border-color-lighter);
  margin-top: 4px;
  padding-top: 8px;
}

.more-actions-dropdown :deep(.el-dropdown-menu__item.is-dividedbefore) {
  border-top: 1px solid var(--el-border-color-lighter);
}

.history-list {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  max-height: 65vh;
  overflow-y: auto;
  padding: 0 4px;
}

.history-list::-webkit-scrollbar {
  width: 8px;
}

.history-list::-webkit-scrollbar-track {
  background: var(--el-fill-color-blank);
  border-radius: 4px;
}

.history-list::-webkit-scrollbar-thumb {
  background: var(--el-border-color);
  border-radius: 4px;
}

.history-list::-webkit-scrollbar-thumb:hover {
  background: var(--el-text-color-disabled);
}

.history-item {
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.history-item-card {
  background: var(--el-bg-color);
  border-radius: 8px;
  padding: 0;
  border: 1px solid var(--el-border-color-light);
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.history-item-card:hover {
  border-color: var(--el-color-primary-light-5);
  box-shadow: 0 2px 8px var(--el-color-primary-light-9);
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  gap: 8px;
}

.header-left-group {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.header-right-group {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.add-to-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-right: 4px;
  user-select: none;
}

.icon-action-btn {
  padding: 4px;
  height: auto;
  color: var(--el-text-color-secondary);
  transition: color 0.2s ease, background 0.2s ease;
  border-radius: 4px;
}

.icon-action-btn:hover {
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.icon-action-btn .el-icon {
  font-size: 15px;
}

.history-source-badge {
  display: inline-flex;
  flex-direction: row;
  flex-wrap: nowrap;
  align-items: center;
  gap: 6px;
  padding: 2px 10px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid transparent;
  cursor: pointer;
  position: relative;
  white-space: nowrap;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.history-source-badge:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.history-source-badge .el-icon {
  transition: transform 0.25s ease;
  flex-shrink: 0;
}

.history-source-badge:hover .el-icon {
  transform: scale(1.4);
}

.history-source-badge.creative {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  border-color: var(--el-color-primary-light-7);
}

.history-source-badge.bookAnalysis {
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
  border-color: var(--el-color-warning-light-7);
}

.history-source-badge.textEditor {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
  border-color: var(--el-color-success-light-7);
}

.history-source-badge.workflow {
  background: var(--history-badge-workflow-bg, var(--el-color-danger-light-9));
  color: var(--history-badge-workflow-color, var(--el-color-danger));
  border-color: var(--history-badge-workflow-border, var(--el-color-danger-light-7));
}

.history-time {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: all;
  letter-spacing: 0.3px;
  transition: color 0.2s ease;
}

.history-time:hover {
  color: var(--el-color-primary);
}

.history-item-body {
  padding: 8px 12px 10px;
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 0;
}

.model-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  align-self: flex-start;
  padding: 1px 8px;
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  border-radius: 4px;
  font-size: 11px;
}

.model-tag .el-icon {
  font-size: 12px;
}

.history-item-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.title-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
}

.history-item-preview {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  border: 1px solid var(--el-border-color-extra-light);
}

.preview-code {
  margin: 0;
  padding: 8px 10px;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
  font-size: 11px;
  line-height: 1.55;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 130px;
  overflow-y: auto;
}

.preview-code code {
  font-family: inherit;
  color: var(--el-color-success-dark-2, var(--el-color-success));
  background: transparent;
}

.history-item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  border-top: 1px solid var(--el-border-color-extra-light);
  background: var(--el-fill-color-light);
  gap: 8px;
}

.log-id {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
  font-size: 10px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  user-select: all;
  letter-spacing: 0.3px;
  transition: color 0.2s ease;
  flex-shrink: 0;
}

.log-id:hover {
  color: var(--el-color-primary);
}

.footer-stats {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 11px;
  min-width: 0;
  flex: 1;
  justify-content: flex-end;
}

.stat-consumed {
  color: var(--el-color-danger);
  font-weight: 500;
  white-space: nowrap;
}

.stat-consumed.is-zero {
  color: var(--el-color-success);
}

.stat-generated {
  color: var(--el-color-success);
  font-weight: 500;
  white-space: nowrap;
}

.stat-detail {
  color: var(--el-color-primary);
  cursor: pointer;
  user-select: none;
  display: inline-flex;
  align-items: center;
  gap: 2px;
  transition: color 0.2s ease;
  white-space: nowrap;
}

.stat-detail:hover {
  color: var(--el-color-primary-light-3);
}

.detail-arrow {
  font-size: 11px;
  transition: transform 0.2s ease;
}

.stat-detail:hover .detail-arrow {
  transform: translateX(2px);
}

.empty-history {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 20px;
  color: var(--el-text-color-secondary);
}

.empty-state-icon {
  opacity: 0.3;
  color: var(--el-text-color-secondary);
}

.empty-title {
  margin-top: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-regular);
}

.empty-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-top: 4px;
}

.history-pagination {
  display: flex;
  justify-content: center;
  padding: 16px 0 8px;
  margin-top: 8px;
}

.history-pagination :deep(.el-pagination) {
  --el-pagination-button-bg-color: var(--el-fill-color-light);
  --el-pagination-hover-color: var(--el-color-primary);
}

.history-pagination :deep(.el-pager li) {
  background: var(--el-fill-color-light);
  border-radius: 4px;
  margin: 0 2px;
  min-width: 28px;
  height: 28px;
  line-height: 28px;
  font-size: 12px;
}

.history-pagination :deep(.el-pager li.is-active) {
  background: var(--el-color-primary);
  color: #fff;
}

.history-pagination :deep(.btn-prev),
.history-pagination :deep(.btn-next) {
  background: var(--el-fill-color-light);
  border-radius: 4px;
  min-width: 28px;
  height: 28px;
}

.history-detail-dialog :deep(.el-dialog) {
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}

.history-detail-dialog :deep(.el-dialog__header) {
  padding: 14px 24px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color-page);
}

.history-detail-dialog :deep(.el-dialog__title) {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.history-detail-dialog :deep(.el-dialog__body) {
  padding: 0;
  max-height: 70vh;
  overflow-y: auto;
}

.history-detail-dialog :deep(.el-dialog__footer) {
  padding: 12px 24px;
  border-top: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
}

.history-detail-content {
  background: var(--el-bg-color);
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 24px;
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-light);
}

.detail-source-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px;
  border-radius: 4px;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid transparent;
}

.detail-source-badge.creative {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  border-color: var(--el-color-primary-light-7);
}

.detail-source-badge.bookAnalysis {
  background: var(--el-color-warning-light-9);
  color: var(--el-color-warning);
  border-color: var(--el-color-warning-light-7);
}

.detail-source-badge.textEditor {
  background: var(--el-color-success-light-9);
  color: var(--el-color-success);
  border-color: var(--el-color-success-light-7);
}

.detail-source-badge.workflow {
  background: var(--el-color-info-light-9);
  color: var(--el-color-info);
  border-color: var(--el-color-info-light-7);
}

.detail-time {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.detail-body {
  padding: 20px 24px;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
}

.detail-preview {
  background: var(--el-fill-color-light);
  padding: 16px;
  border-radius: 6px;
  border: 1px solid var(--el-border-color-light);
}

.conversation-messages {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  background: var(--el-bg-color-page);
  min-height: 400px;
  max-height: 65vh;
  overflow-y: auto;
}

.message-related-context {
  margin-bottom: 8px;
}

.related-context-inline {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  margin-bottom: 6px;
}

.related-icon {
  color: var(--el-color-primary);
  font-size: 12px;
}

.related-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-weight: 500;
}

.related-context-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.related-item {
  padding: 6px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  border: 1px solid var(--el-border-color-extra-light);
}

.related-item-header {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 4px;
}

.related-item-title {
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.related-item-content {
  font-size: 12px;
  color: var(--el-text-color-regular);
  line-height: 1.5;
  max-height: 150px;
  overflow-y: auto;
}

.related-item-content :deep(p) {
  margin: 4px 0;
}

.related-item-content :deep(ul),
.related-item-content :deep(ol) {
  margin: 4px 0;
  padding-left: 20px;
}

.related-tag {
  font-size: 12px;
}

.message-bubble {
  border-radius: 8px;
  padding: 12px 16px;
  max-width: 85%;
  animation: messageSlideIn 0.3s ease;
  border: 1px solid var(--el-border-color-light);
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
  background: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary-light-7);
  margin-left: auto;
}

.message-bubble.assistant {
  background: var(--el-bg-color);
  border-color: var(--el-border-color-light);
}

.message-bubble.system {
  background: var(--el-color-warning-light-9);
  border-color: var(--el-color-warning-light-7);
  width: 100%;
  max-width: 100%;
}

.message-sender {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
}

.sender-icon {
  color: var(--el-color-primary);
}

.sender-name {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.message-time {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.message-content {
  color: var(--el-text-color-regular);
  line-height: 1.7;
}

.message-content :deep(p) {
  margin: 8px 0;
}

.message-content :deep(pre) {
  background: var(--el-color-info-light-9);
  color: var(--el-text-color-primary);
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 12px 0;
  border: 1px solid var(--el-border-color-lighter);
}

.message-content :deep(code) {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
}

.message-content :deep(blockquote) {
  border-left: 3px solid var(--el-color-primary-light-5);
  padding-left: 12px;
  margin: 12px 0;
  color: var(--el-text-color-regular);
  background: var(--el-fill-color-light);
  padding: 8px 12px;
  border-radius: 4px;
}

/* 追问弹窗样式 */
.continue-dialog :deep(.el-dialog) {
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}

.continue-dialog :deep(.el-dialog__header) {
  padding: 14px 24px;
  border-bottom: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color-page);
  border-radius: 16px 16px 0 0;
  margin-right: 0;
}

.continue-dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.header-title .el-icon {
  color: var(--el-color-primary);
  font-size: 18px;
}

.continue-dialog :deep(.el-dialog__body) {
  padding: 0;
  background: var(--el-bg-color);
}

.continue-dialog :deep(.el-dialog__footer) {
  padding: 12px 24px;
  border-top: 1px solid var(--el-border-color-light);
  background: var(--el-bg-color);
  border-radius: 0 0 16px 16px;
}

.continue-dialog-content {
  display: flex;
  flex-direction: column;
  height: 600px;
}

.continue-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  background: var(--el-bg-color-page);
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.history-collapsed-hint {
  text-align: center;
  padding: 12px 0;
  margin-bottom: 12px;
  position: relative;
}

.history-collapsed-hint::before,
.history-collapsed-hint::after {
  content: '';
  position: absolute;
  top: 50%;
  width: 30%;
  height: 1px;
  background: linear-gradient(to right, transparent, var(--el-border-color-light), transparent);
}

.history-collapsed-hint::before {
  left: 0;
}

.history-collapsed-hint::after {
  right: 0;
}

.expand-btn {
  font-size: 12px;
  color: var(--el-color-primary);
  padding: 4px 12px;
  border-radius: 4px;
  background: var(--el-color-primary-light-9);
  transition: background 0.2s ease, color 0.2s ease;
  position: relative;
  z-index: 1;
  border: 1px solid var(--el-color-primary-light-7);
}

.expand-btn:hover {
  background: var(--el-color-primary-light-7);
}

.inline-expand-btn {
  font-size: 11px;
  padding: 2px 8px;
  color: var(--el-color-primary);
  transition: color 0.2s ease;
}

.inline-expand-btn:hover {
  color: var(--el-color-primary-light-3);
}

.history-message {
  opacity: 0.75;
  transition: all 0.3s ease;
}

.history-message:hover {
  opacity: 1;
}

.message-slide-enter-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.message-slide-leave-active {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.message-slide-enter-from {
  opacity: 0;
  transform: translateY(20px);
}

.message-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

.message-slide-move {
  transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.continue-message-bubble {
  border-radius: 8px;
  padding: 12px 16px;
  max-width: 85%;
  animation: messageSlideIn 0.3s ease;
  border: 1px solid var(--el-border-color-light);
}

.continue-message-bubble.user {
  background: var(--el-color-primary-light-9);
  border-color: var(--el-color-primary-light-7);
  margin-left: auto;
}

.continue-message-bubble.assistant {
  background: var(--el-bg-color);
  border-color: var(--el-border-color-light);
}

.continue-message-bubble.system {
  background: var(--el-color-warning-light-9);
  border-color: var(--el-color-warning-light-7);
  width: 100%;
  max-width: 100%;
}

.continue-message-sender {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
}

.continue-message-content {
  color: var(--el-text-color-regular);
  line-height: 1.7;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 8px 0;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--el-color-primary);
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-indicator span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0.8);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.continue-input-area {
  padding: 12px 24px;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color-light);
  display: flex;
  align-items: flex-end;
  gap: 12px;
}

.continue-input {
  flex: 1;
}

.continue-input :deep(.el-textarea__inner) {
  padding: 10px 14px;
  border-radius: 8px;
  border: 1px solid var(--el-border-color-light);
  background: var(--el-fill-color-light);
  transition: border-color 0.2s ease, background 0.2s ease;
  font-size: 14px;
}

.continue-input :deep(.el-textarea__inner):focus {
  border-color: var(--el-color-primary);
  background: var(--el-bg-color);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-9);
}

.continue-input-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
  align-items: flex-end;
}

.fullscreen-btn {
  border-color: var(--el-border-color);
  color: var(--el-text-color-secondary);
}

.fullscreen-btn:hover {
  color: var(--el-color-primary);
  border-color: var(--el-color-primary-light-5);
}

.continue-input-area.is-fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 2000;
  padding: 40px 24px;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.continue-input-area.is-fullscreen .continue-input {
  width: 80%;
  max-width: 900px;
}

.model-select {
  width: 180px;
}

.model-select :deep(.el-select__wrapper) {
  border-radius: 8px;
  border: 1px solid var(--el-border-color-light);
  background: var(--el-fill-color-light);
  transition: border-color 0.2s ease, background 0.2s ease;
  height: 40px;
}

.model-select :deep(.el-select__wrapper):hover {
  border-color: var(--el-color-primary-light-5);
}

.model-select :deep(.el-select__wrapper).is-focused {
  border-color: var(--el-color-primary);
  background: var(--el-bg-color);
  box-shadow: 0 0 0 2px var(--el-color-primary-light-9);
}

.model-option-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.model-option-content span {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.send-btn {
  width: 40px;
  height: 40px;
  background: var(--el-color-primary);
  border: none;
  transition: background 0.2s ease;
}

.send-btn:hover:not(:disabled) {
  background: var(--el-color-primary-light-3);
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

<style>
/* 模型选择下拉菜单暗色主题 - 全局样式 */
:root[data-theme='dark'] .el-select-dropdown,
:root[data-theme='dark'] .el-select__popper,
:root[data-theme='dark'] .el-popper.el-select__popper {
  background: var(--el-bg-color-overlay) !important;
  border: 1px solid var(--el-border-color-light) !important;
  box-shadow: var(--el-box-shadow-light) !important;
}

:root[data-theme='dark'] .el-select-dropdown__item {
  color: var(--el-text-color-regular) !important;
  background: transparent !important;
}

:root[data-theme='dark'] .el-select-dropdown__item:hover {
  background: var(--el-fill-color-light) !important;
}

:root[data-theme='dark'] .el-select-dropdown__item.is-selected {
  color: var(--el-color-primary-light-3) !important;
  background: var(--el-color-primary-light-9) !important;
}

:root[data-theme='dark'] .el-select-dropdown__item.is-hovering {
  background: var(--el-fill-color-blank) !important;
}

:root[data-theme='dark'] .el-select-dropdown__wrap {
  background: var(--el-bg-color-overlay) !important;
}

:root[data-theme='dark'] .el-select-dropdown__list {
  background: var(--el-bg-color-overlay) !important;
  padding: 4px 0 !important;
}

:root[data-theme='dark'] .el-popper.is-light,
:root[data-theme='dark'] .el-popper.is-pure {
  background: var(--el-bg-color-overlay) !important;
  border: 1px solid var(--el-border-color-light) !important;
}

:root[data-theme='dark'] .el-popper.is-light .el-popper__arrow::before,
:root[data-theme='dark'] .el-popper.is-pure .el-popper__arrow::before {
  background: var(--el-bg-color-overlay) !important;
  border-color: var(--el-border-color-light) !important;
}

:root[data-theme='dark'] .el-scrollbar {
  background: var(--el-bg-color-overlay) !important;
}

:root[data-theme='dark'] .el-scrollbar__view {
  background: var(--el-bg-color-overlay) !important;
}

:root[data-theme='dark'] .el-scrollbar__bar {
  background: var(--el-border-color-light) !important;
}

:root[data-theme='dark'] .el-scrollbar__thumb {
  background: var(--el-color-primary-light-5) !important;
}

:root[data-theme='dark'] .el-select-dropdown__empty {
  color: var(--el-text-color-secondary) !important;
  background: var(--el-bg-color-overlay) !important;
}

:root[data-theme='dark'] .el-select-dropdown__item .el-tag,
:root[data-theme='dark'] .model-option-content .el-tag {
  background: var(--el-fill-color-light) !important;
  border-color: var(--el-color-primary-light-5) !important;
  color: var(--el-color-primary-light-3) !important;
}

/* 关联内容区域暗色模式 */
:root[data-theme='dark'] .message-related-context {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .related-context-inline {
  background: rgba(45, 212, 191, 0.06);
  border: 1px solid rgba(45, 212, 191, 0.12);
}

:root[data-theme='dark'] .related-icon {
  color: #2dd4bf;
}

:root[data-theme='dark'] .related-label {
  color: var(--el-text-color-secondary);
}

:root[data-theme='dark'] .related-item {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .related-item-header {
  color: var(--el-text-color-primary);
}

:root[data-theme='dark'] .related-item-title {
  color: var(--el-text-color-primary);
}

:root[data-theme='dark'] .related-item-content {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .related-item-content :deep(p),
:root[data-theme='dark'] .related-item-content :deep(ul),
:root[data-theme='dark'] .related-item-content :deep(ol) {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .inline-expand-btn {
  color: #2dd4bf;
}

:root[data-theme='dark'] .inline-expand-btn:hover {
  color: #5eead4;
}

:root[data-theme='dark'] .continue-message-content {
  color: var(--el-text-color-regular);
}

/* 历史记录列表项暗色模式 */
:root[data-theme='dark'] .model-tag {
  background: rgba(45, 212, 191, 0.1);
  color: #2dd4bf;
}

:root[data-theme='dark'] .title-text {
  color: #f1f5f9;
}

:root[data-theme='dark'] .history-item-preview {
  background: rgba(15, 29, 51, 0.5);
  border-color: rgba(45, 212, 191, 0.08);
}

:root[data-theme='dark'] .preview-code {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .preview-code code {
  color: #5eead4;
}

:root[data-theme='dark'] .history-item-footer {
  background: rgba(15, 29, 51, 0.4);
  border-color: rgba(45, 212, 191, 0.08);
}

:root[data-theme='dark'] .log-id {
  color: var(--el-text-color-placeholder);
}

:root[data-theme='dark'] .log-id:hover {
  color: #2dd4bf;
}

:root[data-theme='dark'] .stat-consumed {
  color: #f87171;
}

:root[data-theme='dark'] .stat-consumed.is-zero {
  color: #5eead4;
}

:root[data-theme='dark'] .stat-generated {
  color: #5eead4;
}

:root[data-theme='dark'] .stat-detail {
  color: #2dd4bf;
}

:root[data-theme='dark'] .stat-detail:hover {
  color: #5eead4;
}

/* 续写对话框暗色模式 */
:root[data-theme='dark'] .continue-messages {
  background: rgba(8, 18, 36, 0.4);
}

:root[data-theme='dark'] .history-collapsed-hint::before,
:root[data-theme='dark'] .history-collapsed-hint::after {
  background: linear-gradient(to right, transparent, rgba(45, 212, 191, 0.2), transparent);
}

:root[data-theme='dark'] .expand-btn {
  color: #2dd4bf;
  background: rgba(45, 212, 191, 0.1);
  border-color: rgba(45, 212, 191, 0.25);
}

:root[data-theme='dark'] .expand-btn:hover {
  background: rgba(45, 212, 191, 0.2);
  color: #5eead4;
}

/* 续写消息气泡暗色模式 */
:root[data-theme='dark'] .continue-message-bubble {
  border-color: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .continue-message-bubble.user {
  background: rgba(45, 212, 191, 0.08);
  border-color: rgba(45, 212, 191, 0.2);
}

:root[data-theme='dark'] .continue-message-bubble.assistant {
  background: rgba(15, 29, 51, 0.7);
  border-color: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .continue-message-bubble.system {
  background: rgba(245, 158, 11, 0.1);
  border-color: rgba(245, 158, 11, 0.2);
}

:root[data-theme='dark'] .continue-message-sender .sender-icon {
  color: #2dd4bf;
}

:root[data-theme='dark'] .continue-message-sender .sender-name {
  color: var(--el-text-color-primary);
}

/* 续写输入区与模型选择器暗色模式 */
:root[data-theme='dark'] .continue-input-area {
  background: rgba(8, 18, 36, 0.5);
  border-color: rgba(45, 212, 191, 0.12);
}

:root[data-theme='dark'] .continue-input-area.is-fullscreen {
  background: rgba(0, 0, 0, 0.6);
}

:root[data-theme='dark'] .continue-input :deep(.el-textarea__inner) {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.15);
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .continue-input :deep(.el-textarea__inner):focus {
  background: rgba(15, 29, 51, 0.8);
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 2px rgba(45, 212, 191, 0.15);
}

:root[data-theme='dark'] .model-select :deep(.el-select__wrapper) {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.15);
  box-shadow: 0 0 0 1px rgba(45, 212, 191, 0.05) inset;
}

:root[data-theme='dark'] .model-select :deep(.el-select__wrapper):hover {
  border-color: rgba(45, 212, 191, 0.3);
}

:root[data-theme='dark'] .model-select :deep(.el-select__wrapper).is-focused {
  background: rgba(15, 29, 51, 0.8);
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 2px rgba(45, 212, 191, 0.15);
}

:root[data-theme='dark'] .model-select :deep(.el-select__placeholder) {
  color: var(--el-text-color-placeholder);
}

:root[data-theme='dark'] .model-select :deep(.el-select__selected-item) {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .model-select :deep(.el-select__suffix) {
  color: var(--el-text-color-secondary);
}

:root[data-theme='dark'] .model-option-content {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .model-option-content span {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .send-button {
  background: linear-gradient(135deg, #2dd4bf, #14b8a6);
  color: #ffffff;
}

:root[data-theme='dark'] .send-button:hover {
  background: linear-gradient(135deg, #5eead4, #2dd4bf);
}

/* 历史记录卡片暗色模式 */
:root[data-theme='dark'] .history-item-card {
  background: rgba(15, 29, 51, 0.55);
  border-color: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .history-item-card:hover {
  border-color: rgba(45, 212, 191, 0.35);
  box-shadow: 0 2px 12px rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .add-to-label {
  color: var(--el-text-color-secondary);
}

:root[data-theme='dark'] .icon-action-btn {
  color: var(--el-text-color-secondary);
}

:root[data-theme='dark'] .icon-action-btn:hover {
  color: #2dd4bf;
  background: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .history-source-badge.creative {
  background: rgba(45, 212, 191, 0.12);
  color: #2dd4bf;
  border-color: rgba(45, 212, 191, 0.3);
}

:root[data-theme='dark'] .history-source-badge.bookAnalysis {
  background: rgba(245, 158, 11, 0.12);
  color: #fbbf24;
  border-color: rgba(245, 158, 11, 0.3);
}

:root[data-theme='dark'] .history-source-badge.textEditor {
  background: rgba(16, 185, 129, 0.12);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.3);
}

:root[data-theme='dark'] .history-source-badge.workflow {
  --history-badge-workflow-bg: rgba(239, 68, 68, 0.12);
  --history-badge-workflow-color: #f87171;
  --history-badge-workflow-border: rgba(239, 68, 68, 0.3);
}

:root[data-theme='dark'] .history-source-badge:hover {
  box-shadow: 0 2px 10px rgba(45, 212, 191, 0.15);
}

:root[data-theme='dark'] .history-time {
  color: var(--el-text-color-placeholder);
}

:root[data-theme='dark'] .history-time:hover {
  color: #2dd4bf;
}

/* 历史记录中心头部暗色模式 */
:root[data-theme='dark'] .history-header {
  background: rgba(15, 29, 51, 0.5);
  border-color: rgba(45, 212, 191, 0.12);
}

:root[data-theme='dark'] .header-icon {
  color: #2dd4bf;
}

:root[data-theme='dark'] .history-count {
  color: var(--el-text-color-primary);
}

:root[data-theme='dark'] .header-tabs :deep(.el-tabs__item) {
  color: var(--el-text-color-secondary);
}

:root[data-theme='dark'] .header-tabs :deep(.el-tabs__item.is-active) {
  color: #2dd4bf;
}

:root[data-theme='dark'] .header-tabs :deep(.el-tabs__active-bar) {
  background-color: #2dd4bf;
}

:root[data-theme='dark'] .header-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: rgba(45, 212, 191, 0.1);
}

:root[data-theme='dark'] .search-input :deep(.el-input__wrapper) {
  background: rgba(15, 29, 51, 0.6);
  box-shadow: 0 0 0 1px rgba(45, 212, 191, 0.15) inset;
}

:root[data-theme='dark'] .search-input :deep(.el-input__wrapper):hover {
  box-shadow: 0 0 0 1px rgba(45, 212, 191, 0.3) inset;
}

:root[data-theme='dark'] .search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #2dd4bf inset;
}

:root[data-theme='dark'] .search-input :deep(.el-input__inner) {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .search-input :deep(.el-input__inner::placeholder) {
  color: var(--el-text-color-placeholder);
}

:root[data-theme='dark'] .search-input :deep(.el-input__prefix) {
  color: var(--el-text-color-secondary);
}

:root[data-theme='dark'] .category-select :deep(.el-select__wrapper),
:root[data-theme='dark'] .book-select :deep(.el-select__wrapper) {
  background: rgba(15, 29, 51, 0.6);
  box-shadow: 0 0 0 1px rgba(45, 212, 191, 0.15) inset;
}

:root[data-theme='dark'] .category-select :deep(.el-select__wrapper):hover,
:root[data-theme='dark'] .book-select :deep(.el-select__wrapper):hover {
  box-shadow: 0 0 0 1px rgba(45, 212, 191, 0.3) inset;
}

:root[data-theme='dark'] .category-select :deep(.el-select__placeholder),
:root[data-theme='dark'] .book-select :deep(.el-select__placeholder) {
  color: var(--el-text-color-placeholder);
}

:root[data-theme='dark'] .category-select :deep(.el-select__selected-item),
:root[data-theme='dark'] .book-select :deep(.el-select__selected-item) {
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .more-actions-btn {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.15);
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .more-actions-btn:hover:not(:disabled) {
  background: rgba(45, 212, 191, 0.12);
  border-color: rgba(45, 212, 191, 0.35);
  color: #2dd4bf;
}

:root[data-theme='dark'] .more-actions-dropdown :deep(.el-dropdown-menu) {
  background: rgba(15, 29, 51, 0.95) !important;
  border: 1px solid rgba(45, 212, 191, 0.15) !important;
}

:root[data-theme='dark'] .more-actions-dropdown :deep(.el-dropdown-menu__item) {
  color: var(--el-text-color-regular) !important;
}

:root[data-theme='dark'] .more-actions-dropdown :deep(.el-dropdown-menu__item:hover) {
  background: rgba(45, 212, 191, 0.12) !important;
  color: #2dd4bf !important;
}

:root[data-theme='dark'] .more-actions-dropdown :deep(.el-dropdown-menu__item.is-divided) {
  border-top-color: rgba(45, 212, 191, 0.12) !important;
}

:root[data-theme='dark'] .more-actions-dropdown :deep(.el-dropdown-menu__item.is-disabled) {
  color: var(--el-text-color-disabled) !important;
}

/* 分页暗色模式 */
:root[data-theme='dark'] .history-pagination :deep(.el-pager li) {
  background: rgba(15, 29, 51, 0.6);
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .history-pagination :deep(.el-pager li:hover) {
  color: #2dd4bf;
}

:root[data-theme='dark'] .history-pagination :deep(.el-pager li.is-active) {
  background: #2dd4bf;
  color: #0a1628;
}

:root[data-theme='dark'] .history-pagination :deep(.btn-prev),
:root[data-theme='dark'] .history-pagination :deep(.btn-next) {
  background: rgba(15, 29, 51, 0.6);
  color: var(--el-text-color-regular);
}

:root[data-theme='dark'] .history-pagination :deep(.btn-prev:hover),
:root[data-theme='dark'] .history-pagination :deep(.btn-next:hover) {
  color: #2dd4bf;
}

:root[data-theme='dark'] .history-pagination :deep(.btn-prev:disabled),
:root[data-theme='dark'] .history-pagination :deep(.btn-next:disabled) {
  color: var(--el-text-color-disabled);
  background: rgba(15, 29, 51, 0.3);
}
</style>
