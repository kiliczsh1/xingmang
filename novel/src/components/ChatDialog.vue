<template>
  <Teleport to="body">
    <Transition name="chat-dialog">
      <div v-if="visible" class="chat-dialog-overlay" @click.self="close">
        <div class="chat-dialog-container">
          <div class="chat-dialog-header">
            <div class="chat-dialog-title">
              <div class="title-dot"></div>
              <span>AI 对话</span>
            </div>
            <div class="chat-dialog-meta" v-if="currentConversation">
              <span class="conv-badge">{{ currentConversation.title }}</span>
            </div>
            <button class="chat-dialog-close" @click="close" title="关闭 (Esc)">
              <el-icon><Close /></el-icon>
            </button>
          </div>

          <div class="chat-dialog-body">
            <div class="chat-sidebar">
              <div class="sidebar-header">
                <span class="sidebar-label">对话</span>
                <button class="sidebar-new-btn" @click="createConversation" :disabled="!selectedBookId" title="新建对话">
                  <el-icon><Plus /></el-icon>
                </button>
              </div>
              <div class="book-selector" v-if="books.length > 0">
                <el-select v-model="selectedBookId" placeholder="选择作品" size="small" class="book-select" popper-class="linear-select-popper">
                  <el-option v-for="book in books" :key="book.id" :label="book.title" :value="book.id" />
                </el-select>
              </div>
              <div class="conversation-list">
                <div
                  v-for="conv in conversations"
                  :key="conv.id"
                  :class="['conversation-item', { active: currentConversation?.id === conv.id }]"
                  @click="selectConversation(conv)"
                >
                  <div class="conv-info">
                    <span class="conv-title">{{ conv.title }}</span>
                    <div class="conv-meta-row">
                      <span class="conv-meta">{{ conv.message_count || 0 }} 条</span>
                    </div>
                  </div>
                  <div class="conv-actions" @click.stop>
                    <button class="conv-action-btn" @click="renameConversation(conv)" title="重命名">
                      <el-icon><Edit /></el-icon>
                    </button>
                    <button class="conv-action-btn conv-action-delete" @click="deleteConversation(conv)" title="删除">
                      <el-icon><Delete /></el-icon>
                    </button>
                  </div>
                </div>
                <div v-if="conversations.length === 0" class="empty-list">
                  <p class="empty-list-text">暂无对话，新建一个开始</p>
                </div>
              </div>
            </div>

            <div class="chat-main">
              <div v-if="currentConversation" class="chat-content">
                <div class="chat-messages" ref="chatMessagesRef">
                  <div
                    v-for="(msg, index) in chatMessages"
                    :key="index"
                    :class="['message', msg.role]"
                  >
                    <div class="message-avatar">
                      <el-icon v-if="msg.role === 'user'"><User /></el-icon>
                      <el-icon v-else-if="msg.role === 'assistant'"><ChatDotRound /></el-icon>
                      <el-icon v-else><InfoFilled /></el-icon>
                    </div>
                    <div class="message-body">
                      <div class="message-role-label">
                        {{ msg.role === 'user' ? 'You' : msg.role === 'assistant' ? 'AI' : 'System' }}
                      </div>
                      <div class="message-content">
                        <MarkdownRenderer v-if="msg.role === 'assistant'" :content="msg.content" />
                        <div v-else class="plain-content">{{ msg.content }}</div>
                      </div>
                    </div>
                    <div class="message-actions">
                      <button class="msg-action-btn" @click="copyMessage(msg.content)" title="复制">
                        <el-icon><DocumentCopy /></el-icon>
                      </button>
                      <button class="msg-action-btn msg-action-delete" @click="deleteChatMessage(index)" title="删除">
                        <el-icon><Delete /></el-icon>
                      </button>
                    </div>
                  </div>
                  <div v-if="sending" class="message assistant">
                    <div class="message-avatar">
                      <el-icon><ChatDotRound /></el-icon>
                    </div>
                    <div class="message-body">
                      <div class="message-role-label">AI</div>
                      <div class="message-content typing">
                        <span class="typing-dot"></span>
                        <span class="typing-dot"></span>
                        <span class="typing-dot"></span>
                      </div>
                    </div>
                  </div>
                </div>

                <div class="chat-input-area">
                  <div class="config-row">
                    <el-select v-model="selectedConfigId" placeholder="选择模型" size="small" class="config-select" popper-class="linear-select-popper">
                      <el-option v-for="config in apiConfigs" :key="config.id" :label="config.name" :value="config.id" />
                    </el-select>
                  </div>
                  <div class="input-row">
                    <div class="input-wrapper">
                      <textarea
                        v-model="userInput"
                        class="chat-textarea"
                        :rows="2"
                        placeholder="输入消息… (Ctrl+Enter 发送)"
                        @keydown.ctrl.enter="sendMessage"
                        :disabled="sending"
                      ></textarea>
                      <button
                        v-if="!sending"
                        class="send-btn"
                        :class="{ disabled: !userInput.trim() }"
                        @click="sendMessage"
                        :disabled="!userInput.trim()"
                        title="发送"
                      >
                        <el-icon><Promotion /></el-icon>
                      </button>
                      <button
                        v-else
                        class="send-btn stop-btn"
                        @click="stopGeneration"
                        title="停止生成"
                      >
                        <el-icon><VideoPause /></el-icon>
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              <div v-else class="empty-chat">
                <div class="empty-state">
                  <div class="empty-icon">
                    <el-icon><ChatDotRound /></el-icon>
                  </div>
                  <p class="empty-title">开始对话</p>
                  <p class="empty-desc">选择或创建一个对话来开始与 AI 交流</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ChatDotRound,
  Close,
  Plus,
  Edit,
  Delete,
  User,
  InfoFilled,
  DocumentCopy,
  Promotion,
  VideoPause
} from '@element-plus/icons-vue'
import MarkdownRenderer from './MarkdownRenderer.vue'
import { conversationAPI, bookAPI, configAPI } from '@/api'
import type { ChatMessage, Book, ApiModel } from '@/types'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const close = () => {
  emit('update:visible', false)
}

const books = ref<Book[]>([])
const selectedBookId = ref<number | null>(null)
const conversations = ref<any[]>([])
const currentConversation = ref<any>(null)
const chatMessages = ref<ChatMessage[]>([])
const userInput = ref('')
const sending = ref(false)
const selectedConfigId = ref<number | null>(null)
const apiConfigs = ref<ApiModel[]>([])
const chatAbortController = ref<AbortController | null>(null)
const chatMessagesRef = ref<HTMLElement>()

const fetchBooks = async () => {
  try {
    const res = await bookAPI.getAll()
    if (res.success && res.data) {
      books.value = res.data
      if (res.data.length > 0 && !selectedBookId.value) {
        selectedBookId.value = res.data[0].id
      }
    }
  } catch (error) {
    console.error('加载作品列表失败:', error)
  }
}

const fetchConversations = async () => {
  try {
    const res = await conversationAPI.list()
    if (res.success && res.data) {
      conversations.value = res.data
    }
  } catch (error) {
    console.error('加载对话列表失败:', error)
  }
}

const fetchConfigs = async () => {
  try {
    const res = await configAPI.getAll()
    if (res.success && res.data) {
      apiConfigs.value = res.data.filter(m => m.enabled !== 0)
      if (res.data.length > 0 && !selectedConfigId.value) {
        selectedConfigId.value = res.data[0].id
      }
    }
  } catch (error) {
    console.error('加载API配置失败:', error)
  }
}

const createConversation = async () => {
  if (!selectedBookId.value) {
    ElMessage.warning('请先选择作品')
    return
  }
  try {
    const res = await conversationAPI.create({
      book_id: selectedBookId.value,
      title: `新对话 ${new Date().toLocaleTimeString()}`
    })
    if (res.success && res.data) {
      await fetchConversations()
      await selectConversation(res.data)
    }
  } catch (error) {
    console.error('创建对话失败:', error)
  }
}

const selectConversation = async (conv: any) => {
  currentConversation.value = conv
  try {
    const res = await conversationAPI.getMessages(conv.id)
    if (res.success && res.data) {
      chatMessages.value = res.data
    }
  } catch (error) {
    console.error('加载消息失败:', error)
  }
  await nextTick()
  scrollToBottom()
}

const renameConversation = async (conv: any) => {
  try {
    const { value } = await ElMessageBox.prompt('请输入新的对话名称', '重命名对话', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: conv.title || '新对话',
      inputErrorMessage: '对话名称不能为空'
    })
    if (value && value.trim()) {
      await conversationAPI.update(conv.id, { title: value.trim() })
      if (currentConversation.value?.id === conv.id) {
        currentConversation.value.title = value.trim()
      }
      await fetchConversations()
      ElMessage.success('重命名成功')
    }
  } catch {
    // 用户取消
  }
}

const deleteConversation = async (conv: any) => {
  try {
    await ElMessageBox.confirm('确定要删除此对话吗？此操作不可恢复。', '删除对话', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await conversationAPI.delete(conv.id)
    if (currentConversation.value?.id === conv.id) {
      currentConversation.value = null
      chatMessages.value = []
    }
    await fetchConversations()
    ElMessage.success('对话已删除')
  } catch {
    // 用户取消
  }
}

const deleteChatMessage = async (index: number) => {
  const msg = chatMessages.value[index]
  if (!msg || !msg.id) {
    chatMessages.value.splice(index, 1)
    return
  }
  try {
    await conversationAPI.deleteMessage(currentConversation.value.id, msg.id)
    chatMessages.value.splice(index, 1)
    ElMessage.success('消息已删除')
  } catch (error) {
    console.error('删除消息失败:', error)
  }
}

const copyMessage = (content: string) => {
  navigator.clipboard.writeText(content).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const scrollToBottom = () => {
  if (chatMessagesRef.value) {
    nextTick(() => {
      chatMessagesRef.value!.scrollTop = chatMessagesRef.value!.scrollHeight
    })
  }
}

const sendMessage = async () => {
  if (!currentConversation.value) {
    ElMessage.warning('请先选择或创建对话')
    return
  }
  if (!userInput.value.trim()) return

  const userContent = userInput.value.trim()
  const userMessage: ChatMessage = {
    role: 'user',
    content: userContent
  }
  chatMessages.value.push(userMessage)

  await conversationAPI.saveMessage(currentConversation.value.id, {
    role: 'user',
    content: userContent
  })

  userInput.value = ''

  const assistantMessage: ChatMessage = {
    role: 'assistant',
    content: ''
  }
  chatMessages.value.push(assistantMessage)
  const messageIndex = chatMessages.value.length - 1

  try {
    sending.value = true
    chatAbortController.value = new AbortController()
    scrollToBottom()

    const response = await fetch('/api/ai/chat', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        messages: chatMessages.value.slice(0, -1),
        configId: selectedConfigId.value
      }),
      signal: chatAbortController.value.signal
    })

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()
    if (!reader) throw new Error('无法获取响应流')

    let buffer = ''
    let rawContent = ''
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const data = line.slice(6)
          if (data === '[DONE]') break
          try {
            const parsed = JSON.parse(data)
            const currentMessage = chatMessages.value[messageIndex]
            if (parsed.content && currentMessage) {
              rawContent += parsed.content
              currentMessage.content = rawContent
              scrollToBottom()
            }
            if (parsed.error) {
              ElMessage.error(parsed.error)
            }
          } catch {
            // JSON parse error, skip
          }
        }
      }
    }

    const savedMessage = chatMessages.value[messageIndex]
    if (savedMessage?.content) {
      await conversationAPI.saveMessage(currentConversation.value.id, {
        role: 'assistant',
        content: savedMessage.content
      })
      await fetchConversations()
    } else {
      chatMessages.value.pop()
      ElMessage.error('AI未返回内容')
    }
  } catch (error: any) {
    if (error.name === 'AbortError') {
      const partialContent = chatMessages.value[messageIndex]?.content || ''
      if (partialContent) {
        if (chatMessages.value[messageIndex]) {
          chatMessages.value[messageIndex].content = partialContent
        }
        await conversationAPI.saveMessage(currentConversation.value.id, {
          role: 'assistant',
          content: partialContent
        })
        await fetchConversations()
        ElMessage.info('已停止生成，部分内容已保存')
      } else {
        chatMessages.value.pop()
      }
    } else {
      chatMessages.value.pop()
      ElMessage.error('发送失败: ' + (error.message || '未知错误'))
    }
  } finally {
    sending.value = false
    chatAbortController.value = null
  }
}

const stopGeneration = () => {
  if (chatAbortController.value) {
    chatAbortController.value.abort()
  }
}

watch(() => props.visible, (newVal) => {
  if (newVal) {
    fetchBooks()
    fetchConversations()
    fetchConfigs()
  }
})

onMounted(() => {
  fetchBooks()
  fetchConversations()
  fetchConfigs()
})
</script>

<style scoped>
/* ============================================
   Linear-inspired Design System
   ============================================ */

/* --- CSS Custom Properties --- */
.chat-dialog-overlay {
  --ds-bg-primary: #0d0d0d;
  --ds-bg-elevated: #141414;
  --ds-bg-hover: #1f1f1f;
  --ds-bg-active: #292929;
  --ds-bg-surface: #1a1a1a;
  --ds-accent: #5e6ad2;
  --ds-accent-hover: #7c82e0;
  --ds-accent-subtle: rgba(94, 106, 210, 0.12);
  --ds-accent-glow: rgba(94, 106, 210, 0.25);
  --ds-text-primary: rgba(255, 255, 255, 0.93);
  --ds-text-secondary: rgba(255, 255, 255, 0.62);
  --ds-text-tertiary: rgba(255, 255, 255, 0.40);
  --ds-text-inverse: rgba(0, 0, 0, 0.90);
  --ds-border-default: rgba(255, 255, 255, 0.07);
  --ds-border-hover: rgba(255, 255, 255, 0.12);
  --ds-border-active: rgba(94, 106, 210, 0.30);
  --ds-border-strong: rgba(255, 255, 255, 0.10);
  --ds-red: #ef4444;
  --ds-red-hover: #f87171;
  --ds-red-subtle: rgba(239, 68, 68, 0.12);
  --ds-green: #10b981;
  --ds-green-subtle: rgba(16, 185, 129, 0.12);
  --ds-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.3);
  --ds-shadow-md: 0 4px 12px rgba(0, 0, 0, 0.4);
  --ds-shadow-lg: 0 20px 60px rgba(0, 0, 0, 0.6);
  --ds-radius-sm: 6px;
  --ds-radius-md: 8px;
  --ds-radius-lg: 12px;
  --ds-radius-xl: 16px;
  --ds-radius-full: 9999px;
  --ds-ease-out: cubic-bezier(0.16, 1, 0.3, 1);
  --ds-ease-in-out: cubic-bezier(0.65, 0, 0.35, 1);
  --ds-duration-fast: 100ms;
  --ds-duration-normal: 150ms;
  --ds-duration-slow: 250ms;
  --ds-font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
  --ds-font-mono: 'JetBrains Mono', 'SF Mono', 'Monaco', 'Cascadia Code', monospace;
  --ds-text-xs: 11px;
  --ds-text-sm: 12px;
  --ds-text-base: 13px;
  --ds-text-lg: 14px;
}

/* 

/* Light theme overrides */
:root:not([data-theme='dark']) .chat-dialog-overlay {
  --ds-bg-primary: #ffffff;
  --ds-bg-elevated: #f7f8f8;
  --ds-bg-hover: #f0f1f2;
  --ds-bg-active: #e8e9ea;
  --ds-bg-surface: #fcfcfc;
  --ds-text-primary: rgba(0, 0, 0, 0.88);
  --ds-text-secondary: rgba(0, 0, 0, 0.55);
  --ds-text-tertiary: rgba(0, 0, 0, 0.36);
  --ds-border-default: rgba(0, 0, 0, 0.06);
  --ds-border-hover: rgba(0, 0, 0, 0.10);
  --ds-border-strong: rgba(0, 0, 0, 0.08);
  --ds-shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.04);
  --ds-shadow-md: 0 4px 12px rgba(0, 0, 0, 0.08);
  --ds-shadow-lg: 0 20px 60px rgba(0, 0, 0, 0.12);
}

/* ============================================
   Overlay & Container
   ============================================ */

.chat-dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  font-family: var(--ds-font-sans);
}

:root:not([data-theme='dark']) .chat-dialog-overlay {
  background: rgba(0, 0, 0, 0.30);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
}

.chat-dialog-container {
  width: 880px;
  max-width: 94vw;
  height: 620px;
  max-height: 84vh;
  border-radius: var(--ds-radius-xl);
  background: var(--ds-bg-primary);
  border: 1px solid var(--ds-border-default);
  box-shadow: var(--ds-shadow-lg), 0 0 0 1px var(--ds-border-strong) inset;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

@media (max-width: 768px) {
  .chat-dialog-container {
    width: 100vw;
    height: 100vh;
    max-width: 100vw;
    max-height: 100vh;
    border-radius: 0;
    border: none;
  }
}

/* ============================================
   Header
   ============================================ */

.chat-dialog-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  height: 48px;
  min-height: 48px;
  border-bottom: 1px solid var(--ds-border-default);
  background: var(--ds-bg-primary);
}

.chat-dialog-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: var(--ds-text-lg);
  font-weight: 600;
  color: var(--ds-text-primary);
  letter-spacing: -0.01em;
}

.title-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--ds-accent);
  box-shadow: 0 0 8px var(--ds-accent-glow);
}

.chat-dialog-meta {
  min-width: 0;
}

.conv-badge {
  font-size: var(--ds-text-sm);
  color: var(--ds-text-secondary);
  padding: 3px 10px;
  border-radius: var(--ds-radius-full);
  background: var(--ds-accent-subtle);
  border: 1px solid rgba(94, 106, 210, 0.12);
}

.chat-dialog-close {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: var(--ds-radius-sm);
  background: transparent;
  color: var(--ds-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--ds-duration-fast) var(--ds-ease-out);
}

.chat-dialog-close:hover {
  background: var(--ds-bg-hover);
  color: var(--ds-text-primary);
}

/* ============================================
   Body Layout
   ============================================ */

.chat-dialog-body {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}

/* ============================================
   Sidebar
   ============================================ */

.chat-sidebar {
  width: 224px;
  flex-shrink: 0;
  border-right: 1px solid var(--ds-border-default);
  display: flex;
  flex-direction: column;
  background: var(--ds-bg-elevated);
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 14px;
  border-bottom: 1px solid var(--ds-border-default);
}

.sidebar-label {
  font-size: var(--ds-text-sm);
  font-weight: 600;
  color: var(--ds-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.sidebar-new-btn {
  width: 28px;
  height: 28px;
  border: 1px solid var(--ds-border-default);
  border-radius: var(--ds-radius-sm);
  background: transparent;
  color: var(--ds-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--ds-duration-fast) var(--ds-ease-out);
}

.sidebar-new-btn:hover:not(:disabled) {
  background: var(--ds-accent);
  border-color: var(--ds-accent);
  color: #fff;
}

.sidebar-new-btn:disabled {
  opacity: 0.3;
  cursor: not-allowed;
}

.book-selector {
  padding: 10px 14px;
  border-bottom: 1px solid var(--ds-border-default);
}

.book-select {
  width: 100%;
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px;
}

.conversation-list::-webkit-scrollbar {
  width: 4px;
}

.conversation-list::-webkit-scrollbar-track {
  background: transparent;
}

.conversation-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.08);
  border-radius: 2px;
}

.conversation-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  border-radius: var(--ds-radius-md);
  cursor: pointer;
  transition: all var(--ds-duration-fast) var(--ds-ease-out);
  margin-bottom: 1px;
}

.conversation-item:hover {
  background: var(--ds-bg-hover);
}

.conversation-item.active {
  background: var(--ds-accent-subtle);
}

.conversation-item.active .conv-title {
  color: var(--ds-accent);
}

.conv-info {
  min-width: 0;
  flex: 1;
}

.conv-title {
  display: block;
  font-size: var(--ds-text-base);
  font-weight: 500;
  color: var(--ds-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.conv-meta-row {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 1px;
}

.conv-meta {
  font-size: var(--ds-text-xs);
  color: var(--ds-text-tertiary);
}

.conv-actions {
  display: flex;
  opacity: 0;
  transition: opacity var(--ds-duration-fast) var(--ds-ease-out);
  flex-shrink: 0;
  gap: 2px;
}

.conversation-item:hover .conv-actions {
  opacity: 1;
}

.conv-action-btn {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--ds-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--ds-duration-fast) var(--ds-ease-out);
  padding: 0;
}

.conv-action-btn:hover {
  background: var(--ds-bg-active);
  color: var(--ds-text-primary);
}

.conv-action-delete:hover {
  background: var(--ds-red-subtle);
  color: var(--ds-red);
}

.empty-list {
  padding: 24px 16px;
  text-align: center;
}

.empty-list-text {
  font-size: var(--ds-text-sm);
  color: var(--ds-text-tertiary);
  margin: 0;
}

/* ============================================
   Main Chat Area
   ============================================ */

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  background: var(--ds-bg-primary);
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* ============================================
   Messages
   ============================================ */

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.chat-messages::-webkit-scrollbar {
  width: 5px;
}

.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
}

.chat-messages::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.12);
}

.message {
  display: flex;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--ds-radius-md);
  transition: background var(--ds-duration-fast) var(--ds-ease-out);
}

.message:hover {
  background: var(--ds-bg-hover);
}

.message-avatar {
  width: 28px;
  height: 28px;
  border-radius: var(--ds-radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 13px;
}

.message.user .message-avatar {
  background: var(--ds-accent-subtle);
  color: var(--ds-accent);
}

.message.assistant .message-avatar {
  background: var(--ds-green-subtle);
  color: var(--ds-green);
}

.message.system .message-avatar {
  background: rgba(245, 158, 11, 0.12);
  color: #f59e0b;
}

.message-body {
  flex: 1;
  min-width: 0;
}

.message-role-label {
  font-size: var(--ds-text-xs);
  font-weight: 600;
  color: var(--ds-text-tertiary);
  margin-bottom: 2px;
  letter-spacing: 0.03em;
  text-transform: uppercase;
}

.message.user .message-role-label {
  color: var(--ds-accent);
}

.message.assistant .message-role-label {
  color: var(--ds-green);
}

.message-content {
  font-size: var(--ds-text-base);
  line-height: 1.65;
  color: var(--ds-text-primary);
}

.plain-content {
  white-space: pre-wrap;
  word-break: break-word;
  color: var(--ds-text-secondary);
}

.message-actions {
  display: flex;
  flex-direction: column;
  opacity: 0;
  transition: opacity var(--ds-duration-fast) var(--ds-ease-out);
  align-self: flex-start;
  padding-top: 18px;
  gap: 0;
}

.message:hover .message-actions {
  opacity: 1;
}

.msg-action-btn {
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--ds-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--ds-duration-fast) var(--ds-ease-out);
  padding: 0;
}

.msg-action-btn:hover {
  background: var(--ds-bg-active);
  color: var(--ds-text-primary);
}

.msg-action-delete:hover {
  background: var(--ds-red-subtle);
  color: var(--ds-red);
}

/* ============================================
   Typing Indicator
   ============================================ */

.typing {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 0;
}

.typing-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--ds-text-tertiary);
  animation: typingPulse 1.4s ease-in-out infinite;
}

.typing-dot:nth-child(2) {
  animation-delay: 0.2s;
}

.typing-dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes typingPulse {
  0%, 80%, 100% {
    opacity: 0.2;
  }
  40% {
    opacity: 0.8;
  }
}

/* ============================================
   Input Area
   ============================================ */

.chat-input-area {
  padding: 12px 20px 16px;
  border-top: 1px solid var(--ds-border-default);
  background: var(--ds-bg-primary);
}

.config-row {
  margin-bottom: 10px;
}

.config-select {
  width: 140px;
}

.input-row {
  display: flex;
  gap: 0;
}

.input-wrapper {
  flex: 1;
  display: flex;
  align-items: flex-end;
  border: 1px solid var(--ds-border-hover);
  border-radius: var(--ds-radius-lg);
  background: var(--ds-bg-elevated);
  transition: border-color var(--ds-duration-fast) var(--ds-ease-out);
  overflow: hidden;
}

.input-wrapper:focus-within {
  border-color: var(--ds-border-active);
  box-shadow: 0 0 0 1px var(--ds-accent-subtle);
}

.chat-textarea {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  color: var(--ds-text-primary);
  font-size: var(--ds-text-base);
  font-family: var(--ds-font-sans);
  line-height: 1.5;
  padding: 10px 14px;
  resize: none;
}

.chat-textarea::placeholder {
  color: var(--ds-text-tertiary);
}

.chat-textarea:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.send-btn {
  width: 36px;
  height: 36px;
  margin: 0 8px 8px 0;
  border: none;
  border-radius: var(--ds-radius-sm);
  background: var(--ds-accent);
  color: #fff;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: all var(--ds-duration-fast) var(--ds-ease-out);
}

.send-btn:hover:not(.disabled) {
  background: var(--ds-accent-hover);
  transform: scale(1.04);
}

.send-btn:active:not(.disabled) {
  transform: scale(0.96);
}

.send-btn.disabled {
  background: var(--ds-bg-active);
  color: var(--ds-text-tertiary);
  cursor: not-allowed;
}

.stop-btn {
  background: var(--ds-red-subtle);
  color: var(--ds-red);
}

.stop-btn:hover {
  background: var(--ds-red) !important;
  color: #fff !important;
}

/* ============================================
   Empty State
   ============================================ */

.empty-chat {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.empty-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--ds-radius-lg);
  background: var(--ds-accent-subtle);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
  color: var(--ds-accent);
  font-size: 20px;
}

.empty-title {
  font-size: var(--ds-text-lg);
  font-weight: 600;
  color: var(--ds-text-primary);
  margin: 0 0 6px;
}

.empty-desc {
  font-size: var(--ds-text-sm);
  color: var(--ds-text-tertiary);
  margin: 0;
}

/* ============================================
   Element Plus Overrides
   ============================================ */

:deep(.el-select) {
  --el-select-border-color-hover: var(--ds-border-hover);
  --el-select-input-focus-border-color: var(--ds-accent);
}

:deep(.el-select .el-input__wrapper) {
  background: var(--ds-bg-elevated);
  border-color: var(--ds-border-default);
  box-shadow: none;
  border-radius: var(--ds-radius-sm);
}

:deep(.el-select .el-input__wrapper:hover) {
  border-color: var(--ds-border-hover);
}

:deep(.el-select .el-input__wrapper.is-focus) {
  border-color: var(--ds-accent);
  box-shadow: 0 0 0 1px var(--ds-accent-subtle);
}

:deep(.el-select .el-input__inner) {
  color: var(--ds-text-primary);
  font-size: var(--ds-text-sm);
}

/* ============================================
   Transitions
   ============================================ */

.chat-dialog-enter-active {
  transition: opacity var(--ds-duration-slow) var(--ds-ease-out);
}

.chat-dialog-leave-active {
  transition: opacity 150ms var(--ds-ease-in-out);
}

.chat-dialog-enter-active .chat-dialog-container {
  transition: all var(--ds-duration-slow) var(--ds-ease-out);
}

.chat-dialog-leave-active .chat-dialog-container {
  transition: all 150ms var(--ds-ease-in-out);
}

.chat-dialog-enter-from {
  opacity: 0;
}

.chat-dialog-enter-from .chat-dialog-container {
  opacity: 0;
  transform: scale(0.97) translateY(4px);
}

.chat-dialog-leave-to {
  opacity: 0;
}

.chat-dialog-leave-to .chat-dialog-container {
  opacity: 0;
  transform: scale(0.98);
}

/* ============================================
   Markdown Content Overrides
   ============================================ */

.message-content :deep(pre) {
  background: var(--ds-bg-hover);
  border: 1px solid var(--ds-border-default);
  border-radius: var(--ds-radius-sm);
  padding: 12px 14px;
  font-size: var(--ds-text-sm);
  font-family: var(--ds-font-mono);
  overflow-x: auto;
}

.message-content :deep(code) {
  font-family: var(--ds-font-mono);
  font-size: var(--ds-text-sm);
  background: var(--ds-bg-hover);
  padding: 1px 5px;
  border-radius: 3px;
  color: var(--ds-accent);
}

.message-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: var(--ds-text-primary);
}

/* ============================================
   Responsive
   ============================================ */

@media (max-width: 768px) {
  .chat-sidebar {
    width: 180px;
  }

  .chat-messages {
    padding: 12px 14px;
  }

  .chat-input-area {
    padding: 10px 14px 14px;
  }
}
</style>
