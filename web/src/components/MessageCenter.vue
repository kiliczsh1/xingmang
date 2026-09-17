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
          <div class="qq-group-card">
            <div class="qq-icon-wrapper">
              <el-icon :size="40"><ChatLineRound /></el-icon>
            </div>
            <div class="qq-info">
              <p class="qq-title">星芒创作者 QQ 群</p>
              <p class="qq-desc">加入 QQ 群与开发者和其他创作者交流、反馈问题、获取最新动态。</p>
              <div class="qq-number-row">
                <span class="qq-number">821979308</span>
                <button class="qq-copy-btn" type="button" @click="copyQQGroup">
                  <el-icon><CopyDocument /></el-icon>
                  <span>{{ copied ? '已复制' : '复制群号' }}</span>
                </button>
              </div>
              <p class="qq-hint">复制后在 QQ 中搜索群号即可加入</p>
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
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { ChatLineRound, CopyDocument, EditPen, TopRight } from '@element-plus/icons-vue'

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

// ===== QQ 群引导 =====
const QQ_GROUP = '821979308'
const copied = ref(false)

const copyQQGroup = async () => {
  try {
    await navigator.clipboard.writeText(QQ_GROUP)
  } catch {
    // 兜底：创建临时输入框复制
    const input = document.createElement('input')
    input.value = QQ_GROUP
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
  }
  copied.value = true
  ElMessage.success('群号已复制：' + QQ_GROUP)
  setTimeout(() => { copied.value = false }, 2000)
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
  background: var(--accent, #3b82f6);
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.4);
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
  color: var(--ink2, rgba(0, 0, 0, 0.55));
}

.message-tabs :deep(.el-tabs__item.is-active) {
  color: var(--accent, #3b82f6);
  font-weight: 600;
}

.message-tabs :deep(.el-tabs__active-bar) {
  background-color: var(--accent, #3b82f6);
}

/* --- QQ Group Card --- */

.qq-group-card {
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 24px;
  border-radius: 16px;
  background: var(--accent-soft, rgba(59, 130, 246, 0.08));
  border: 1px solid var(--topbar-border, rgba(59, 130, 246, 0.16));
}

.qq-icon-wrapper {
  width: 72px;
  height: 72px;
  border-radius: 18px;
  background: var(--accent-soft, rgba(59, 130, 246, 0.20));
  color: var(--accent, #3b82f6);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.qq-info {
  flex: 1;
  min-width: 0;
}

.qq-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--app-color, #1e293b);
  margin: 0 0 4px 0;
}

.qq-desc {
  font-size: 12px;
  color: var(--ink2, rgba(0, 0, 0, 0.55));
  line-height: 1.6;
  margin: 0 0 12px 0;
}

.qq-number-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.qq-number {
  font-size: 24px;
  font-weight: 800;
  color: var(--accent, #3b82f6);
  font-family: 'SF Mono', 'JetBrains Mono', Consolas, monospace;
  letter-spacing: .02em;
}

.qq-copy-btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 6px 12px;
  border: 0;
  border-radius: 8px;
  background: var(--accent-soft, rgba(59, 130, 246, 0.12));
  color: var(--accent, #3b82f6);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s ease;
}

.qq-copy-btn:hover {
  background: var(--accent-soft, rgba(59, 130, 246, 0.22));
  filter: brightness(1.08);
}

.qq-copy-btn .el-icon {
  font-size: 13px;
}

.qq-hint {
  font-size: 11px;
  color: var(--ink3, rgba(0, 0, 0, 0.36));
  margin: 10px 0 0 0;
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
  background: var(--accent-soft, rgba(59, 130, 246, 0.06));
  border: 1px solid var(--topbar-border, rgba(59, 130, 246, 0.12));
}

.intro-icon {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--accent-soft, rgba(59, 130, 246, 0.12));
  color: var(--accent, #3b82f6);
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
  color: var(--ink2, rgba(0, 0, 0, 0.55));
  line-height: 1.6;
  margin: 0;
}

.feedback-action-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: 14px;
  background: var(--accent-soft, rgba(59, 130, 246, 0.08));
  border: 1px solid var(--topbar-border, rgba(59, 130, 246, 0.16));
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.feedback-action-card:hover {
  background: var(--accent-soft, rgba(59, 130, 246, 0.14));
  border-color: var(--accent, rgba(59, 130, 246, 0.32));
  transform: translateY(-1px);
  box-shadow: 0 8px 24px rgba(59, 130, 246, 0.12);
  filter: brightness(1.06);
}

.feedback-action-card:active {
  transform: translateY(0);
}

.action-card-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  background: var(--accent-soft, rgba(59, 130, 246, 0.18));
  color: var(--accent, #3b82f6);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: background 0.25s ease;
}

.feedback-action-card:hover .action-card-icon {
  background: var(--accent-soft, rgba(59, 130, 246, 0.28));
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
  color: var(--ink2, rgba(0, 0, 0, 0.55));
}

.action-card-arrow {
  color: var(--accent, #3b82f6);
  flex-shrink: 0;
  transition: transform 0.25s ease;
}

.feedback-action-card:hover .action-card-arrow {
  transform: translate(3px, -3px);
}

.feedback-note {
  font-size: 12px;
  color: var(--ink3, rgba(0, 0, 0, 0.36));
  text-align: center;
  margin: 0;
}
</style>

<!-- 非作用域样式：el-drawer 以 append-to-body 挂载，需用 :root[data-theme] 全局适配暗色 -->
<style>
.message-center-drawer-modal .el-overlay {
  background-color: rgba(0, 0, 0, 0.5);
}

.message-center-drawer .el-drawer {
  background: var(--card, #ffffff);
  color: var(--app-color, #1e293b);
}

.message-center-drawer .el-drawer__header {
  color: var(--app-color, #1e293b);
  border-bottom: 1px solid var(--topbar-border, #e6eaf0);
}

.message-center-drawer .el-drawer__body {
  background: var(--card, #ffffff);
}

.message-center-drawer .el-drawer__close-btn {
  color: var(--ink2, #5a6a80);
}

.message-center-drawer .el-drawer__close-btn:hover {
  color: var(--accent, #3b82f6);
}

/* 暗色模式：抽屉面板/头部/正文/滚动条跟随深色主题 */
:root[data-theme='dark'] .message-center-drawer .el-drawer,
:root[data-theme='dark'] .message-center-drawer .el-drawer__body {
  background: var(--card, #161b22);
  color: var(--app-color, #e6edf3);
}

:root[data-theme='dark'] .message-center-drawer .el-drawer__header {
  color: var(--app-color, #e6edf3);
  border-bottom: 1px solid var(--topbar-border, #21262d);
}

:root[data-theme='dark'] .message-center-drawer .el-drawer__close-btn {
  color: var(--ink2, #8b98a8);
}

:root[data-theme='dark'] .message-center-drawer .el-drawer__close-btn:hover {
  color: var(--accent, #58a6ff);
}

:root[data-theme='dark'] .message-center-drawer-modal .el-overlay {
  background-color: rgba(0, 0, 0, 0.6);
}

/* 抽屉内滚动条：暗色下用浅色半透明 */
:root[data-theme='dark'] .message-center-drawer .el-drawer__body::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.18);
}

:root[data-theme='dark'] .message-center-drawer .el-drawer__body::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.28);
}
</style>
