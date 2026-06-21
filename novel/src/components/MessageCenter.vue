<template>
  <el-drawer
    v-model="drawerVisible"
    direction="rtl"
    size="520px"
    :close-on-click-modal="true"
    :destroy-on-close="false"
    class="message-center-drawer"
    modal-class="message-center-drawer-modal"
  >
    <template #header>
      <div class="drawer-header">
        <div class="header-title">
          <span class="title-dot"></span>
          <span>消息中心</span>
        </div>
      </div>
    </template>

    <div class="message-center-content">
      <el-tabs v-model="activeTab" class="message-tabs">
        <el-tab-pane label="系统消息" name="system">
          <div class="message-list">
            <div v-if="systemMessages.length === 0" class="empty-messages">
              <div class="empty-icon">
                <el-icon :size="48"><Bell /></el-icon>
              </div>
              <p class="empty-title">暂无消息</p>
              <p class="empty-desc">系统消息和版本更新会在这里显示</p>
            </div>

            <div
              v-for="msg in systemMessages"
              :key="msg.id"
              :class="['message-card', { unread: !msg.read }]"
              @click="markAsRead(msg)"
            >
              <div class="message-dot" v-if="!msg.read"></div>
              <div class="message-icon-wrapper" :class="msg.type">
                <el-icon :size="18">
                  <Promotion v-if="msg.type === 'announcement'" />
                  <Present v-else-if="msg.type === 'feature'" />
                  <WarningFilled v-else-if="msg.type === 'maintenance'" />
                  <InfoFilled v-else />
                </el-icon>
              </div>
              <div class="message-body">
                <div class="message-header-row">
                  <span class="message-title">{{ msg.title }}</span>
                  <span class="message-time">{{ formatTime(msg.time) }}</span>
                </div>
                <p class="message-content-text">{{ msg.content }}</p>
                <div v-if="msg.link" class="message-link-row">
                  <a :href="msg.link" target="_blank" rel="noopener noreferrer" class="message-link">
                    查看详情 <el-icon><ArrowRight /></el-icon>
                  </a>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="意见反馈" name="feedback">
          <div class="feedback-section">
            <div class="feedback-intro">
              <div class="intro-icon">
                <el-icon :size="28"><ChatLineRound /></el-icon>
              </div>
              <div class="intro-text">
                <p class="intro-title">你的声音，我们都听见</p>
                <p class="intro-desc">欢迎反馈使用中遇到的问题、功能建议或任何想法。每一条反馈我们都会认真对待。</p>
              </div>
            </div>

            <div class="feedback-action-card" @click="openFeedbackForm">
              <div class="action-card-icon">
                <el-icon :size="32"><EditPen /></el-icon>
              </div>
              <div class="action-card-body">
                <span class="action-card-title">填写反馈表单</span>
                <span class="action-card-desc">腾讯文档 · 扫码登录后可持久保持登录态</span>
              </div>
              <div class="action-card-arrow">
                <el-icon :size="20"><TopRight /></el-icon>
              </div>
            </div>

            <p class="feedback-note">点击后将打开反馈表单，登录后可保持自动登录状态。</p>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  Bell,
  Promotion,
  Present,
  WarningFilled,
  InfoFilled,
  ArrowRight,
  ChatLineRound,
  EditPen,
  TopRight
} from '@element-plus/icons-vue'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const drawerVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

const activeTab = ref('system')

const feedbackFormUrl = 'https://docs.qq.com/form/page/sequence/DTnpNb3pWVUh2Z2tP'

const openFeedbackForm = () => {
  window.open(feedbackFormUrl, '_blank', 'noopener,noreferrer')
}

const STORAGE_KEY = 'message-center-messages'

interface SystemMessage {
  id: string
  type: 'announcement' | 'feature' | 'maintenance' | 'info'
  title: string
  content: string
  time: number
  read: boolean
  link?: string
}

const defaultMessages: SystemMessage[] = [
  {
    id: 'welcome',
    type: 'announcement',
    title: '🎉 欢迎使用星芒 AI 写作',
    content: '感谢你选择星芒 AI 写作！本工具致力于为创作者提供智能、高效的写作辅助体验。如果遇到任何问题，欢迎通过「意见反馈」标签提交反馈。',
    time: Date.now(),
    read: false
  },
  {
    id: 'forum',
    type: 'feature',
    title: '🌐 星芒社区论坛已上线',
    content: '欢迎加入星芒创作者社区！在这里你可以与其他创作者交流写作心得、分享使用技巧、提出功能建议。点击下方链接前往论坛。',
    time: Date.now(),
    read: false,
    link: 'https://kook.vip/sLySCy'
  },
  {
    id: 'feedback',
    type: 'feature',
    title: '💬 意见反馈功能已上线',
    content: '点击左侧「意见反馈」标签即可提交你的使用反馈和建议。你的每一条反馈都会帮助我们让星芒变得更好！',
    time: Date.now(),
    read: false
  }
]

const loadMessages = (): SystemMessage[] => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const parsed = JSON.parse(stored)
      return parsed
    }
  } catch {
    // ignore
  }
  return []
}

const saveMessages = (messages: SystemMessage[]) => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(messages))
}

const storedMessages = loadMessages()

if (storedMessages.length === 0) {
  saveMessages(defaultMessages)
}

const mergeNewDefaults = (existing: SystemMessage[]) => {
  const existingIds = new Set(existing.map(m => m.id))
  const missing = defaultMessages.filter(m => !existingIds.has(m.id))
  if (missing.length > 0) {
    return [...existing, ...missing]
  }
  return existing
}

const systemMessages = ref<SystemMessage[]>(
  mergeNewDefaults(storedMessages.length > 0 ? storedMessages : defaultMessages)
)

watch(
  systemMessages,
  (newVal) => {
    saveMessages(newVal)
  },
  { deep: true }
)

const formatTime = (timestamp: number) => {
  const now = Date.now()
  const diff = now - timestamp
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 7) return `${days} 天前`

  const date = new Date(timestamp)
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const markAsRead = (msg: SystemMessage) => {
  msg.read = true
}

</script>

<style scoped>
/* ============================================
   Message Center Drawer Styles
   ============================================ */

.drawer-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--app-color, #1e293b);
}

.title-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #2dd4bf;
  box-shadow: 0 0 8px rgba(45, 212, 191, 0.4);
}

.message-center-content {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.message-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.message-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow-y: auto;
}

.message-tabs :deep(.el-tabs__header) {
  margin-bottom: 16px;
}

.message-tabs :deep(.el-tabs__item) {
  font-size: 13px;
  color: var(--ds-text-secondary, rgba(0, 0, 0, 0.55));
}

.message-tabs :deep(.el-tabs__item.is-active) {
  color: #2dd4bf;
  font-weight: 600;
}

.message-tabs :deep(.el-tabs__active-bar) {
  background-color: #2dd4bf;
}

/* --- Message List --- */

.message-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.message-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 12px;
  background: var(--ds-bg-elevated, #f7f8f8);
  border: 1px solid var(--ds-border-default, rgba(0, 0, 0, 0.06));
  cursor: pointer;
  transition: all 0.15s ease;
  position: relative;
}

.message-card:hover {
  background: var(--ds-bg-hover, #f0f1f2);
  border-color: var(--ds-border-hover, rgba(0, 0, 0, 0.10));
}

.message-card.unread {
  border-left: 2px solid #2dd4bf;
  background: rgba(45, 212, 191, 0.04);
}

.message-dot {
  position: absolute;
  top: 14px;
  left: 6px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #2dd4bf;
  box-shadow: 0 0 6px rgba(45, 212, 191, 0.5);
}

.message-icon-wrapper {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 1px;
}

.message-icon-wrapper.announcement {
  background: rgba(45, 212, 191, 0.1);
  color: #2dd4bf;
}

.message-icon-wrapper.feature {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
}

.message-icon-wrapper.maintenance {
  background: rgba(245, 158, 11, 0.1);
  color: #f59e0b;
}

.message-icon-wrapper.info {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.message-body {
  flex: 1;
  min-width: 0;
}

.message-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 4px;
}

.message-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--app-color, #1e293b);
  line-height: 1.4;
}

.message-time {
  font-size: 11px;
  color: var(--ds-text-tertiary, rgba(0, 0, 0, 0.36));
  white-space: nowrap;
  flex-shrink: 0;
}

.message-content-text {
  font-size: 12px;
  color: var(--ds-text-secondary, rgba(0, 0, 0, 0.55));
  line-height: 1.6;
  margin: 0;
}

.message-link-row {
  margin-top: 8px;
}

.message-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #2dd4bf;
  text-decoration: none;
  transition: color 0.15s ease;
}

.message-link:hover {
  color: #5ee8d4;
}

/* --- Empty State --- */

.empty-messages {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px 20px;
  text-align: center;
}

.empty-icon {
  color: var(--ds-text-tertiary, rgba(0, 0, 0, 0.36));
  margin-bottom: 16px;
  opacity: 0.5;
}

.empty-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ds-text-secondary, rgba(0, 0, 0, 0.55));
  margin: 0 0 6px 0;
}

.empty-desc {
  font-size: 12px;
  color: var(--ds-text-tertiary, rgba(0, 0, 0, 0.36));
  margin: 0;
}

/* --- Feedback Tab --- */

.feedback-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.feedback-intro {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 16px;
  border-radius: 12px;
  background: rgba(45, 212, 191, 0.06);
  border: 1px solid rgba(45, 212, 191, 0.12);
}

.intro-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: rgba(45, 212, 191, 0.12);
  color: #2dd4bf;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.intro-text {
  flex: 1;
}

.intro-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--app-color, #1e293b);
  margin: 0 0 4px 0;
}

.intro-desc {
  font-size: 12px;
  color: var(--ds-text-secondary, rgba(0, 0, 0, 0.55));
  line-height: 1.6;
  margin: 0;
}

.feedback-action-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.08), rgba(34, 211, 238, 0.04));
  border: 1px solid rgba(45, 212, 191, 0.16);
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.feedback-action-card:hover {
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.14), rgba(34, 211, 238, 0.08));
  border-color: rgba(45, 212, 191, 0.32);
  transform: translateY(-1px);
  box-shadow: 0 8px 24px rgba(45, 212, 191, 0.12);
}

.feedback-action-card:active {
  transform: translateY(0);
}

.action-card-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.18), rgba(34, 211, 238, 0.08));
  color: #2dd4bf;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.25s ease;
}

.feedback-action-card:hover .action-card-icon {
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.28), rgba(34, 211, 238, 0.14));
}

.action-card-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.action-card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--app-color, #1e293b);
}

.action-card-desc {
  font-size: 12px;
  color: var(--ds-text-secondary, rgba(0, 0, 0, 0.55));
}

.action-card-arrow {
  color: #2dd4bf;
  flex-shrink: 0;
  transition: transform 0.25s ease;
}

.feedback-action-card:hover .action-card-arrow {
  transform: translate(3px, -3px);
}

.feedback-note {
  font-size: 12px;
  color: var(--ds-text-tertiary, rgba(0, 0, 0, 0.36));
  text-align: center;
  margin: 0;
}

/* --- Dark Theme Overrides --- */

:root[data-theme='dark'] .message-card {
  background: rgba(15, 29, 51, 0.6);
  border-color: rgba(45, 212, 191, 0.08);
}

:root[data-theme='dark'] .message-card:hover {
  background: rgba(15, 29, 51, 0.8);
  border-color: rgba(45, 212, 191, 0.16);
}

:root[data-theme='dark'] .message-card.unread {
  background: rgba(45, 212, 191, 0.06);
}

:root[data-theme='dark'] .feedback-intro {
  background: rgba(45, 212, 191, 0.08);
  border-color: rgba(45, 212, 191, 0.14);
}

:root[data-theme='dark'] .feedback-action-card {
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.06), rgba(34, 211, 238, 0.02));
  border-color: rgba(45, 212, 191, 0.10);
}

:root[data-theme='dark'] .feedback-action-card:hover {
  background: linear-gradient(135deg, rgba(45, 212, 191, 0.12), rgba(34, 211, 238, 0.06));
  border-color: rgba(45, 212, 191, 0.24);
}
</style>
