<!-- ChatMessageList.vue —— 消息列表（统一青绿主题色） -->
<template>
  <div class="ds-messages" ref="messagesRef">
    <template v-for="(msg, index) in messages" :key="index">
      <!-- 用户消息 -->
      <div v-if="msg.role === 'user'" class="ds-msg ds-msg--user">
        <div class="ds-msg-avatar ds-msg-avatar--user">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
        </div>
        <div class="ds-msg-body">
          <div class="ds-msg-role ds-msg-role--user">你</div>
          <div class="ds-msg-text ds-msg-text--user">{{ getUserDisplayContent(msg) }}</div>
          <div class="ds-msg-actions">
            <button class="ds-action-btn" @click="copyMessage(msg)" title="复制">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            </button>
            <button v-if="applyToCursor" class="ds-action-btn" @click="applyToCursor(getCopyContent(msg))" title="应用到编辑器">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12l7 7 7-7"/></svg>
            </button>
            <button class="ds-action-btn ds-action-btn--danger" @click="$emit('deleteMessage', index)" title="删除">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </div>
      </div>

      <!-- AI 消息 -->
      <div v-else-if="msg.role === 'assistant'" class="ds-msg ds-msg--ai">
        <div class="ds-msg-avatar ds-msg-avatar--ai">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2a10 10 0 1 0 10 10H12V2z"/>
            <path d="M20 12a8 8 0 1 0-8 8" opacity="0.4"/>
          </svg>
        </div>
        <div class="ds-msg-body ds-msg-body--ai">
          <div class="ds-msg-role ds-msg-role--ai">AI</div>

          <!-- 流式中的消息（最后一条 assistant + loading） -->
          <template v-if="isStreaming(index)">
            <!-- 已有内容 → 显示文本 + 闪烁光标 -->
            <div v-if="getDisplayContent(msg)" class="ds-msg-text ds-msg-text--ai ds-msg-text--streaming">
              <MarkdownRenderer :content="getDisplayContent(msg)" />
              <span class="ds-stream-cursor" />
            </div>
            <!-- 尚无内容 → 思考动画 -->
            <div v-else class="ds-thinking">
              <div class="ds-thinking-dots">
                <span /><span /><span />
              </div>
              <span class="ds-thinking-text">思考中...</span>
            </div>
          </template>

          <!-- 已完成的消息 -->
          <template v-else>
            <div class="ds-msg-text ds-msg-text--ai">
              <MarkdownRenderer :content="getDisplayContent(msg)" />
            </div>
          </template>

          <!-- 操作按钮（仅已完成的消息显示） -->
          <div v-if="!isStreaming(index)" class="ds-msg-actions">
            <button class="ds-action-btn" @click="copyMessage(msg)" title="复制">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
            </button>
            <button v-if="applyToCursor" class="ds-action-btn" @click="applyToCursor(getCopyContent(msg))" title="应用到编辑器">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12l7 7 7-7"/></svg>
            </button>
            <button class="ds-action-btn" @click="$emit('regenerate', index)" title="重新生成">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            </button>
            <button class="ds-action-btn ds-action-btn--danger" @click="$emit('deleteMessage', index)" title="删除">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
            </button>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import type { ChatMessage } from '@/types'

const props = defineProps<{
  messages: ChatMessage[]
  loading: boolean
  getDisplayContent: (msg: ChatMessage) => string
  getUserDisplayContent: (msg: ChatMessage) => string
  getCopyContent: (msg: ChatMessage) => string
  applyToCursor?: ((content: string) => void) | null
}>()

defineEmits<{
  (e: 'regenerate', index: number): void
  (e: 'deleteMessage', index: number): void
}>()

/** 判断指定索引的消息是否正在流式输出中 */
const isStreaming = (index: number): boolean => {
  return props.loading
    && index === props.messages.length - 1
    && props.messages[index]?.role === 'assistant'
}

const messagesRef = ref<HTMLElement>()

const copyMessage = (msg: ChatMessage) => {
  navigator.clipboard.writeText(props.getCopyContent(msg)).then(() => {
    ElMessage.success('已复制')
  }).catch(() => ElMessage.error('复制失败'))
}

watch(() => props.messages.length, () => {
  nextTick(() => {
    if (messagesRef.value) {
      messagesRef.value.scrollTop = messagesRef.value.scrollHeight
    }
  })
})

defineExpose({ messagesRef })
</script>

<style scoped>
.ds-messages {
  flex: 1;
  overflow-y: auto;
  padding: 20px 16px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0.5) 0%, rgba(30, 41, 59, 0.3) 100%);
}

.ds-messages::-webkit-scrollbar { width: 4px; }
.ds-messages::-webkit-scrollbar-track { background: transparent; }
.ds-messages::-webkit-scrollbar-thumb { background: rgba(94, 234, 212, 0.08); border-radius: 2px; }

/* ---- 消息行 ---- */
.ds-msg {
  display: flex;
  gap: 12px;
  padding: 12px 0;
}

/* 头像 */
.ds-msg-avatar {
  width: 28px; height: 28px;
  border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
  margin-top: 2px;
}
.ds-msg-avatar--user {
  background: linear-gradient(135deg, #0f766e 0%, #115e59 100%);
  color: #5eead4;
  box-shadow: 0 4px 16px rgba(15, 118, 110, 0.4);
}
.ds-msg-avatar--ai {
  background: linear-gradient(135deg, #0f766e 0%, #134e4a 100%);
  color: #5eead4;
  box-shadow: 0 4px 16px rgba(15, 118, 110, 0.35);
}

/* 消息体 */
.ds-msg-body { flex: 1; min-width: 0; }

.ds-msg-body--ai {
  background: rgba(51, 65, 85, 0.6);
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 2px 14px 14px 14px;
  padding: 12px 16px;
  position: relative;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
}
/* AI 气泡顶部青绿光带 */
.ds-msg-body--ai::before {
  content: '';
  position: absolute;
  top: 0; left: 10%; width: 80%; height: 1px;
  background: linear-gradient(90deg, transparent, rgba(94, 234, 212, 0.3), transparent);
  pointer-events: none;
}

.ds-msg-role {
  font-size: 11px; font-weight: 600;
  margin-bottom: 6px;
  text-transform: uppercase; letter-spacing: 0.5px;
}
.ds-msg-role--user { color: #5eead4; }
.ds-msg-role--ai { color: #5eead4; }

/* 文本 */
.ds-msg-text {
  font-size: 14px; line-height: 1.75;
  word-break: break-word;
}
.ds-msg-text--user {
  white-space: pre-wrap;
  color: var(--text-primary, #f4f7ff);
}
.ds-msg-text--ai {
  color: var(--text-primary, #f4f7ff);
}

/* 流式输出追加闪烁光标 */
.ds-msg-text--streaming :deep(p:last-of-type)::after {
  content: '▍';
  display: inline;
  animation: ds-blink 0.9s step-end infinite;
  color: #5eead4;
  margin-left: 1px;
}
@keyframes ds-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}

/* Markdown 内容样式 */
.ds-msg-text--ai :deep(p) { margin: 0 0 8px; }
.ds-msg-text--ai :deep(p:last-child) { margin-bottom: 0; }
.ds-msg-text--ai :deep(pre) {
  background: rgba(0,0,0,0.25);
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 8px;
  padding: 12px 14px;
  font-size: 12.5px;
  overflow-x: auto;
  margin: 8px 0;
}
.ds-msg-text--ai :deep(code) {
  font-size: 12.5px;
  background: rgba(94, 234, 212, 0.1);
  padding: 1px 5px;
  border-radius: 4px;
}
.ds-msg-text--ai :deep(pre code) { background: transparent; padding: 0; }
.ds-msg-text--ai :deep(ul),
.ds-msg-text--ai :deep(ol) { padding-left: 20px; margin: 4px 0; }
.ds-msg-text--ai :deep(blockquote) {
  border-left: 3px solid rgba(94, 234, 212, 0.3);
  padding-left: 12px; margin: 8px 0;
  color: var(--text-secondary, #c8d0e0);
}
.ds-msg-text--ai :deep(a) { color: #5eead4; text-decoration: none; }
.ds-msg-text--ai :deep(a:hover) { text-decoration: underline; }
.ds-msg-text--ai :deep(table) { border-collapse: collapse; margin: 8px 0; font-size: 13px; }
.ds-msg-text--ai :deep(th),
.ds-msg-text--ai :deep(td) { border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10)); padding: 6px 10px; }
.ds-msg-text--ai :deep(th) { background: rgba(94, 234, 212, 0.06); }

/* 操作按钮（仅已完成的消息 hover 显示） */
.ds-msg-actions {
  display: flex; gap: 2px;
  opacity: 0; transition: opacity 150ms;
  margin-top: 8px;
}
.ds-msg:hover .ds-msg-actions { opacity: 1; }

.ds-action-btn {
  width: 26px; height: 26px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 6px; border: none; background: transparent;
  color: var(--text-muted, #6b7280);
  cursor: pointer; transition: all 120ms;
}
.ds-action-btn:hover {
  background: rgba(94, 234, 212, 0.08);
  color: #5eead4;
}
.ds-action-btn--danger:hover {
  background: rgba(248, 113, 113, 0.12);
  color: #f87171;
}

/* ---- 思考中动画（流式初始、尚无内容时） ---- */
.ds-thinking {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 0;
}
.ds-thinking-dots {
  display: flex;
  align-items: center;
  gap: 4px;
}
.ds-thinking-dots span {
  width: 6px; height: 6px; border-radius: 50%;
  background: rgba(94, 234, 212, 0.5);
  animation: ds-think-pulse 1.4s ease-in-out infinite;
}
.ds-thinking-dots span:nth-child(2) { animation-delay: 0.2s; }
.ds-thinking-dots span:nth-child(3) { animation-delay: 0.4s; }
@keyframes ds-think-pulse {
  0%, 80%, 100% { transform: scale(0.6); opacity: 0.3; }
  40% { transform: scale(1); opacity: 1; }
}
.ds-thinking-text {
  font-size: 13px;
  color: var(--text-muted, #6b7280);
}
</style>
