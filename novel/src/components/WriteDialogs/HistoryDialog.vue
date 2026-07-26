<template>
  <!-- 历史记录列表弹窗 -->
  <el-dialog v-model="visible" title="正文 AI 历史记录" width="800px" class="history-dialog-modal" append-to-body>
    <div class="history-dialog-content">
      <div v-if="loadingHistory" class="loading-state">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载历史记录中...</span>
      </div>
      <template v-else>
        <div class="history-header">
          <span class="history-count">共 {{ historyRecords.length }} 条记录</span>
          <div class="history-actions">
            <el-button
              type="danger"
              size="small"
              @click="clearHistoryRecords"
              :disabled="historyRecords.length === 0"
            >
              <el-icon><Delete /></el-icon>
              清空历史
            </el-button>
          </div>
        </div>

        <div v-if="historyRecords.length === 0" class="empty-state">
          <el-icon :size="48"><Document /></el-icon>
          <p>暂无历史记录</p>
          <p class="hint">每次 AI API 调用都会在这里留下一条连续对话记录</p>
        </div>

        <div v-else class="history-list">
          <div
            v-for="record in historyRecords"
            :key="record.id"
            class="history-item"
            @click="viewHistoryDetail(record)"
          >
            <div class="history-item-header">
              <div class="history-main">
                <div class="history-icon-wrapper">
                  <el-icon class="history-icon"><ChatDotSquare /></el-icon>
                </div>
                <div class="history-info">
                  <div class="history-title-row">
                    <span class="history-title">{{ record.title }}</span>
                    <el-tag size="small" :type="getHistoryStatusType(record.status)" class="status-tag">
                      {{ getHistoryStatusLabel(record.status) }}
                    </el-tag>
                  </div>
                  <div class="history-meta">
                    <span class="meta-item">{{ record.sourceLabel }}</span>
                    <span class="meta-divider">·</span>
                    <span class="meta-item">{{ record.promptCount }} 条提示词</span>
                    <span class="meta-divider">·</span>
                    <span class="meta-item time">{{ formatTimestamp(record.timestamp) }}</span>
                  </div>
                </div>
              </div>
              <div class="history-actions-mini">
                <el-button
                  size="small"
                  text
                  @click.stop="viewHistoryDetail(record)"
                  title="查看详情"
                >
                  <el-icon><View /></el-icon>
                </el-button>
                <el-button
                  size="small"
                  text
                  type="danger"
                  @click.stop="deleteHistoryRecord(record.id)"
                  title="删除"
                >
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
            <div class="history-preview-line">{{ record.previewContent || '无预览内容' }}</div>
          </div>
        </div>
      </template>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">关闭</el-button>
        <el-button type="primary" @click="loadAllHistory">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 历史记录详情弹窗 -->
  <el-dialog
    v-model="historyDialogVisible"
    :title="selectedHistoryConversation?.title || '历史记录详情'"
    width="900px"
    class="history-detail-dialog"
    destroy-on-close
    append-to-body
  >
    <div v-if="selectedHistoryConversation" class="history-detail-content">
      <div class="history-detail-meta">
        <el-tag size="small" type="info">{{ selectedHistoryConversation.sourceLabel }}</el-tag>
        <el-tag size="small" :type="getHistoryStatusType(selectedHistoryConversation.status)">
          {{ getHistoryStatusLabel(selectedHistoryConversation.status) }}
        </el-tag>
        <span class="history-detail-time">{{ formatTimestamp(selectedHistoryConversation.timestamp) }}</span>
      </div>

      <div class="conversation-messages">
        <div class="messages-container">
          <div
            v-for="(msg, msgIndex) in selectedHistoryConversation.messages"
            :key="msgIndex"
            class="message-bubble"
            :class="getHistoryMessageClass(msg.role)"
          >
            <div class="bubble-avatar">
              <el-icon v-if="msg.role === 'user'"><User /></el-icon>
              <el-icon v-else-if="msg.role === 'assistant'"><ChatDotRound /></el-icon>
              <el-icon v-else-if="msg.role === 'prompt'"><Star /></el-icon>
              <el-icon v-else><Monitor /></el-icon>
            </div>
            <div class="bubble-content">
              <div class="bubble-header">
                <span class="bubble-role">{{ getHistoryRoleLabel(msg.role) }}</span>
                <span class="bubble-time">{{ formatTimestamp(msg.timestamp) }}</span>
              </div>
              <div class="bubble-text">
                <MarkdownRenderer :content="msg.content" />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <el-button @click="historyDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Loading, Delete, Document, ChatDotSquare, View, Refresh, User, ChatDotRound, Star, Monitor } from '@element-plus/icons-vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

type WriteHistoryRole = 'system' | 'user' | 'prompt' | 'assistant'
type WriteHistoryStatus = 'completed' | 'cancelled' | 'failed'

interface WriteHistoryMessage {
  role: WriteHistoryRole
  content: string
  timestamp: number
}

interface WriteHistoryRecord {
  id: string
  bookId: number
  title: string
  source: 'chat' | 'continue' | 'creative2'
  sourceLabel: string
  promptName: string
  promptCount: number
  status: WriteHistoryStatus
  previewContent: string
  timestamp: number
  messages: WriteHistoryMessage[]
}

const WRITE_HISTORY_STORAGE_KEY = 'write-ai-api-history'

const props = defineProps<{
  bookId: number
}>()

const visible = defineModel<boolean>('visible')

const historyRecords = ref<WriteHistoryRecord[]>([])
const loadingHistory = ref(false)
const selectedHistoryConversation = ref<WriteHistoryRecord | null>(null)
const historyDialogVisible = ref(false)

const readHistoryStorage = () => {
  try {
    const raw = localStorage.getItem(WRITE_HISTORY_STORAGE_KEY)
    const parsed = raw ? JSON.parse(raw) : []
    return Array.isArray(parsed) ? parsed as WriteHistoryRecord[] : []
  } catch (error) {
    console.error('读取正文历史记录失败:', error)
    return []
  }
}

const writeHistoryStorage = (records: WriteHistoryRecord[]) => {
  localStorage.setItem(WRITE_HISTORY_STORAGE_KEY, JSON.stringify(records))
}

// 加载所有历史记录
const loadAllHistory = async () => {
  loadingHistory.value = true
  try {
    historyRecords.value = readHistoryStorage()
      .filter(item => item.bookId === props.bookId)
      .sort((a, b) => b.timestamp - a.timestamp)
  } catch (error) {
    console.error('加载历史记录失败:', error)
    ElMessage.error('加载历史记录失败')
  } finally {
    loadingHistory.value = false
  }
}

const viewHistoryDetail = (record: WriteHistoryRecord) => {
  selectedHistoryConversation.value = record
  historyDialogVisible.value = true
}

const deleteHistoryRecord = (recordId: string) => {
  const records = readHistoryStorage().filter(item => item.id !== recordId)
  writeHistoryStorage(records)
  historyRecords.value = records
    .filter(item => item.bookId === props.bookId)
    .sort((a, b) => b.timestamp - a.timestamp)

  if (selectedHistoryConversation.value?.id === recordId) {
    selectedHistoryConversation.value = null
    historyDialogVisible.value = false
  }
}

const clearHistoryRecords = () => {
  const records = readHistoryStorage().filter(item => item.bookId !== props.bookId)
  writeHistoryStorage(records)
  historyRecords.value = []
  selectedHistoryConversation.value = null
  historyDialogVisible.value = false
  ElMessage.success('历史记录已清空')
}

const getHistoryStatusLabel = (status: WriteHistoryStatus) => {
  if (status === 'completed') return '已完成'
  if (status === 'cancelled') return '已取消'
  return '已失败'
}

const getHistoryStatusType = (status: WriteHistoryStatus) => {
  if (status === 'completed') return 'success'
  if (status === 'cancelled') return 'warning'
  return 'danger'
}

const getHistoryRoleLabel = (role: WriteHistoryRole) => {
  if (role === 'user') return '用户'
  if (role === 'prompt') return '提示词'
  if (role === 'assistant') return 'AI'
  return '系统'
}

const getHistoryMessageClass = (role: WriteHistoryRole) => {
  if (role === 'user') return 'user'
  if (role === 'prompt') return 'prompt'
  if (role === 'assistant') return 'assistant'
  return 'system'
}

const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const date = new Date(timeStr)
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

const formatTimestamp = (timestamp: number) => {
  return formatTime(new Date(timestamp).toISOString())
}

// 监听历史记录弹窗打开
watch(visible, async (newVal) => {
  if (newVal) {
    await loadAllHistory()
  }
})
</script>

<style scoped>
.history-dialog-modal :deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
}

.history-dialog-content {
  max-height: 70vh;
  overflow-y: auto;
}

.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 50px 20px;
  color: #999;
}

.loading-state .el-icon {
  font-size: 28px;
  margin-bottom: 12px;
  color: #00c9a7;
}

.empty-state .el-icon {
  margin-bottom: 12px;
  color: #ccc;
}

.empty-state p {
  margin-top: 10px;
  font-size: 13px;
}

.empty-state .hint {
  color: #999;
  font-size: 12px;
}

.history-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eaeaea;
}

.history-count {
  font-size: 13px;
  font-weight: 600;
  color: #3a4a48;
}

.history-actions {
  display: flex;
  gap: 8px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-item {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 14px 16px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.history-item:hover {
  border-color: #10b981;
  background: linear-gradient(135deg, #f0fdf9 0%, #ecfdf5 100%);
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.12);
}

.history-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.history-main {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.history-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #d1fae5 0%, #a7f3d0 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.2);
}

.history-icon {
  color: #059669;
  font-size: 18px;
}

.history-info {
  flex: 1;
  min-width: 0;
}

.history-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 4px;
}

.history-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.status-tag {
  flex-shrink: 0;
  border-radius: 999px !important;
  font-size: 11px !important;
  padding: 2px 10px !important;
}

.history-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #9ca3af;
}

.meta-item {
  white-space: nowrap;
}

.meta-item.time {
  color: #d1d5db;
}

.meta-divider {
  color: #ddd;
  margin: 0 2px;
}

.history-actions-mini {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
}

.history-actions-mini .el-button {
  width: 28px;
  height: 28px;
  padding: 0;
  border-radius: 50%;
  font-size: 14px;
  background: transparent !important;
  border: none !important;
  color: #9ca3af !important;
  transition: all 0.15s ease;
}

.history-actions-mini .el-button:hover {
  background: #f3f4f6 !important;
  color: #374151 !important;
}

.history-actions-mini .el-button--danger:hover {
  background: #fef2f2 !important;
  color: #dc2626 !important;
}

/* 暗色模式 */
:root[data-theme='dark'] .history-item {
  background: #1e293b;
  border-color: #334155;
}

:root[data-theme='dark'] .history-item:hover {
  border-color: #10b981;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.12) 0%, rgba(16, 185, 129, 0.08) 100%);
  box-shadow: 0 2px 12px rgba(16, 185, 129, 0.15);
}

:root[data-theme='dark'] .history-icon-wrapper {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.3) 0%, rgba(16, 185, 129, 0.2) 100%);
  box-shadow: 0 2px 8px rgba(16, 185, 129, 0.25);
}

:root[data-theme='dark'] .history-icon {
  color: #34d399;
}

:root[data-theme='dark'] .history-title {
  color: #f1f5f9;
}

:root[data-theme='dark'] .history-meta {
  color: #64748b;
}

:root[data-theme='dark'] .meta-item.time {
  color: #475569;
}

:root[data-theme='dark'] .history-actions-mini .el-button {
  color: #64748b !important;
}

:root[data-theme='dark'] .history-actions-mini .el-button:hover {
  background: #334155 !important;
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .history-actions-mini .el-button--danger:hover {
  background: rgba(220, 38, 38, 0.15) !important;
  color: #f87171 !important;
}

.history-preview-line {
  font-size: 12px;
  color: #666;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding-left: 42px;
}

.history-detail-content {
  max-height: 70vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.history-detail-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #eaeaea;
}

.history-detail-time {
  font-size: 12px;
  color: #999;
}

.conversation-messages {
  padding: 0;
  background: #fafafa;
  max-height: 600px;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
  min-height: 0;
}

.messages-container {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

.message-bubble {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  max-width: 80%;
  min-width: 0;
}

.message-bubble.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message-bubble.assistant,
.message-bubble.prompt,
.message-bubble.system {
  align-self: flex-start;
}

.bubble-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: linear-gradient(135deg, #00c9a7 0%, #00a896 100%);
  color: #fff;
  font-size: 16px;
}

.message-bubble.user .bubble-avatar {
  background: linear-gradient(135deg, #4a90a4 0%, #3a7a94 100%);
}

.message-bubble.prompt .bubble-avatar {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
}

.message-bubble.system .bubble-avatar {
  background: linear-gradient(135deg, #7c8aa5 0%, #5f6b85 100%);
}

.bubble-content {
  flex: 1;
  min-width: 0;
  max-width: 100%;
  overflow: hidden;
}

.bubble-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 11px;
}

.bubble-role {
  font-weight: 600;
  color: #00c9a7;
}

.message-bubble.user .bubble-role {
  color: #4a90a4;
}

.message-bubble.prompt .bubble-role {
  color: #d97706;
}

.message-bubble.system .bubble-role {
  color: #5f6b85;
}

.bubble-time {
  color: #bbb;
  font-size: 10px;
}

.bubble-text {
  padding: 10px 12px;
  border-radius: 8px;
  font-size: 13px;
  line-height: 1.5;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  border: 1px solid #eaeaea;
  word-break: break-word;
  overflow-wrap: break-word;
  max-width: 100%;
  overflow-x: auto;
}

.message-bubble.user .bubble-text {
  background: #f0fdf9;
  border-color: #d4f4e8;
  border-bottom-right-radius: 4px;
}

.message-bubble.assistant .bubble-text,
.message-bubble.prompt .bubble-text,
.message-bubble.system .bubble-text {
  border-bottom-left-radius: 4px;
}

.message-bubble.prompt .bubble-text {
  background: linear-gradient(135deg, #fff7e6 0%, #fffbeb 100%);
  border-color: rgba(245, 158, 11, 0.2);
}

.message-bubble.system .bubble-text {
  background: linear-gradient(135deg, #f5f7fb 0%, #eef2f7 100%);
  border-color: rgba(95, 107, 133, 0.18);
}

.bubble-text :deep(.markdown-body) {
  font-size: 14px;
  line-height: 1.6;
}

.bubble-text :deep(.markdown-body p) {
  margin-bottom: 8px;
}

.bubble-text :deep(.markdown-body p:last-child) {
  margin-bottom: 0;
}

.bubble-text :deep(.markdown-body pre) {
  background: #f6f8fa;
  border-radius: 6px;
  padding: 12px;
  overflow-x: auto;
  margin: 8px 0;
  border: 1px solid rgba(0, 0, 0, 0.06);
}

.bubble-text :deep(.markdown-body code) {
  background: rgba(0, 201, 167, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
}

.message-bubble.user .bubble-text :deep(.markdown-body code) {
  background: rgba(0, 201, 167, 0.15);
}

.bubble-text :deep(.markdown-body pre code) {
  background: transparent;
  padding: 0;
}

.bubble-text :deep(.markdown-body ul),
.bubble-text :deep(.markdown-body ol) {
  padding-left: 20px;
  margin: 8px 0;
}

.bubble-text :deep(.markdown-body strong) {
  font-weight: 600;
  color: #333;
}

.message-bubble.user .bubble-text :deep(.markdown-body strong) {
  color: #1a4a45;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
