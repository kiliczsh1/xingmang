<template>
  <el-dialog
    v-model="listVisible"
    title="抽卡区历史记录"
    width="900px"
    destroy-on-close
    class="history-dialog"
    modal-class="history-dialog-modal"
    append-to-body
  >
    <div class="history-dialog-content">
      <div class="history-header">
        <div class="header-left">
          <el-icon class="header-icon"><Document /></el-icon>
          <span class="history-count">共 {{ records.length }} 条记录</span>
        </div>
        <div class="history-actions">
          <el-button
            type="danger"
            size="small"
            @click="emit('clear')"
            :disabled="records.length === 0"
            class="clear-all-btn"
          >
            <el-icon><Delete /></el-icon>
            清空历史
          </el-button>
        </div>
      </div>
      <div class="history-list">
        <div
          v-for="(record, index) in records"
          :key="record.id"
          class="history-item"
          @click="viewDetail(record)"
        >
          <div class="history-item-card">
            <div class="history-item-header">
              <div class="generator-info">
                <div class="generator-icon-wrapper">
                  <el-icon class="generator-icon" :size="28">
                    <component :is="record.generatorIcon || 'Lightning'" />
                  </el-icon>
                </div>
                <div class="generator-details">
                  <span class="generator-name">{{ record.generatorName }}</span>
                  <span class="generator-time">{{ formatTimestamp(record.timestamp) }}</span>
                </div>
              </div>
              <div class="history-item-tags">
                <el-tag size="small" effect="plain" class="prompt-count-tag">
                  <el-icon><Document /></el-icon>
                  {{ record.selectedPrompts.length }} 个提示词
                </el-tag>
              </div>
            </div>
            <div class="history-item-actions">
              <el-button
                size="small"
                type="primary"
                @click.stop="viewDetail(record)"
                class="view-detail-btn"
              >
                <el-icon><View /></el-icon>
                查看详情
              </el-button>
              <el-button
                size="small"
                type="danger"
                @click.stop="emit('delete', index)"
                class="delete-btn"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </div>
          </div>
        </div>
        <div v-if="records.length === 0" class="empty-history">
          <div class="empty-state-icon">
            <el-icon :size="64"><Document /></el-icon>
          </div>
          <p class="empty-title">暂无历史记录</p>
          <p class="empty-hint">开始第一次生成后，历史记录会显示在这里</p>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="listVisible = false" class="cancel-btn">取消</el-button>
        <el-button type="primary" @click="listVisible = false" class="confirm-btn">关闭</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
    v-model="detailVisible"
    :title="selectedRecord?.generatorName || '对话详情'"
    width="900px"
    class="history-detail-dialog"
    modal-class="history-detail-dialog-modal"
    destroy-on-close
    align-center
    append-to-body
  >
    <div class="history-detail-content" v-if="selectedRecord">
      <div class="conversation-messages">
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
            <div class="message-content">
              <MarkdownRenderer :content="msg.content" />
            </div>
          </div>
        </template>
      </div>
    </div>
    <template #footer>
      <el-button @click="detailVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Document, Delete, View, User, ChatDotRound, Setting } from '@element-plus/icons-vue'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

interface HistoryMessage {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
}

interface HistoryRecord {
  id: number
  generatorName: string
  generatorIcon?: string
  fixedPrompt: string
  selectedPrompts: Array<{ name: string; content: string }>
  inputParams: string
  additionalInfo: string
  fieldValues: Record<string, string>
  relatedContext: Array<{ type: string; title: string }>
  messages: HistoryMessage[]
  previewContent: string
  timestamp: number
}

const props = defineProps<{
  visible: boolean
  records: HistoryRecord[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'clear'): void
  (e: 'delete', index: number): void
}>()

const listVisible = computed({
  get: () => props.visible,
  set: (v) => emit('update:visible', v)
})

// 详情对话框状态（组件内部管理）
const detailVisible = ref(false)
const selectedRecord = ref<HistoryRecord | null>(null)

const viewDetail = (record: HistoryRecord) => {
  selectedRecord.value = record
  detailVisible.value = true
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
</script>
