<template>
  <el-dialog
    v-model="visible"
    title="历史对话"
    width="400px"
    align-center
    append-to-body
    class="conversation-list-dialog"
  >
    <div class="conversation-list">
      <div
        v-for="conv in conversations"
        :key="conv.id"
        :class="['conversation-item', { active: currentConversationId === conv.id }]"
        @click="handleSelect(conv)"
      >
        <div class="conversation-info">
          <div class="conversation-title">{{ conv.title }}</div>
          <div class="conversation-time">{{ formatTime(conv.updated_at || conv.created_at) }}</div>
        </div>
        <div class="conversation-actions" @click.stop>
          <el-button size="small" text @click="emit('rename', conv)">
            <el-icon><Edit /></el-icon>
          </el-button>
          <el-button size="small" text type="danger" @click="emit('delete', conv)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>
      <div v-if="conversations.length === 0" class="empty-conversations">
        暂无对话记录
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { Edit, Delete } from '@element-plus/icons-vue'

const visible = defineModel<boolean>('visible', { default: false })

const props = defineProps<{
  conversations: any[]
  currentConversationId: number | null
}>()

const emit = defineEmits<{
  select: [conv: any]
  rename: [conv: any]
  delete: [conv: any]
}>()

const handleSelect = (conv: any) => {
  emit('select', conv)
  visible.value = false
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
</script>

<style scoped>
.conversation-list-dialog :deep(.el-dialog__body) {
  padding: 0;
  max-height: 60vh;
  overflow-y: auto;
}

.conversation-list {
  padding: 8px;
}

.conversation-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  margin-bottom: 4px;
  border: 1px solid transparent;
}

.conversation-item:hover {
  background: rgba(16, 185, 129, 0.08);
  border-color: rgba(16, 185, 129, 0.15);
}

.conversation-item.active {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.12) 0%, rgba(6, 78, 59, 0.06) 100%);
  border-color: rgba(16, 185, 129, 0.25);
}

.conversation-info {
  flex: 1;
  min-width: 0;
}

.conversation-title {
  font-size: 14px;
  font-weight: 500;
  color: #065f46;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.conversation-time {
  font-size: 12px;
  color: #6b7280;
  margin-top: 4px;
}

.conversation-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.conversation-item:hover .conversation-actions {
  opacity: 1;
}

.conversation-actions .el-button {
  padding: 4px;
  color: #059669 !important;
}

.conversation-actions .el-button:hover {
  background: rgba(16, 185, 129, 0.1) !important;
}

.conversation-actions .el-button--danger {
  color: #ef4444 !important;
}

.empty-conversations {
  text-align: center;
  padding: 40px 20px;
  color: #9ca3af;
  font-size: 14px;
}
</style>
