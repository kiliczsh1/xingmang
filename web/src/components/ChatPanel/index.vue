<!-- ChatPanel/index.vue —— 主入口（统一青绿主题色） -->
<template>
  <div class="ds-panel" :style="{ width: panelWidth + 'px' }">
    <!-- 左侧对话列表抽屉 -->
    <Transition name="slide-drawer">
      <div v-if="drawerOpen" class="ds-drawer">
        <div class="ds-drawer-header">
          <span class="ds-drawer-title">对话</span>
          <button class="ds-icon-btn" @click="handleNewConversation" title="新建对话">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>
          </button>
          <button class="ds-icon-btn" @click="drawerOpen = false" title="关闭">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
          </button>
        </div>
        <div class="ds-drawer-list">
          <div
            v-for="conv in conversations"
            :key="conv.id"
            :class="['ds-drawer-item', { active: currentConversationId === conv.id }]"
            @click="selectConversation(conv)"
          >
            <span class="ds-drawer-item-title">{{ conv.title }}</span>
            <div class="ds-drawer-item-actions">
              <span class="ds-drawer-item-meta">{{ conv.message_count || 0 }}</span>
              <button class="ds-drawer-item-btn" @click.stop="renameConversation(conv)" title="重命名">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M17 3a2.85 2.85 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z"/></svg>
              </button>
              <button class="ds-drawer-item-btn ds-drawer-item-btn--danger" @click.stop="deleteConversation(conv)" title="删除">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
              </button>
            </div>
          </div>
          <div v-if="conversations.length === 0" class="ds-drawer-empty">暂无对话</div>
        </div>
      </div>
    </Transition>

    <!-- 主体 -->
    <ChatMain
      ref="chatMainRef"
      :book-id="bookId"
      @close="$emit('close')"
      @toggle-drawer="drawerOpen = !drawerOpen"
      @conversations-loaded="onConversationsLoaded"
      @conversation-selected="onConversationSelected"
    />

  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import ChatMain from './ChatMain.vue'

const props = withDefaults(defineProps<{ bookId: number; panelWidth?: number }>(), {
  panelWidth: 480,
})
defineEmits<{ (e: 'close'): void }>()

const drawerOpen = ref(false)
const conversations = ref<any[]>([])
const currentConversationId = ref<number | null>(null)

// ChatMain 实例引用（用于调 createConversation / selectConversation 等）
const chatMainRef = ref<InstanceType<typeof ChatMain>>()

const onConversationsLoaded = (list: any[]) => { conversations.value = list }
const onConversationSelected = (conv: any) => { currentConversationId.value = conv?.id ?? null }

// 新建对话
const handleNewConversation = async () => {
  await chatMainRef.value?.createConversation()
}

// 选择对话（点击列表项）
const selectConversation = async (conv: any) => {
  currentConversationId.value = conv.id
  await chatMainRef.value?.selectConversation(conv)
  drawerOpen.value = false
}

// 删除对话
const deleteConversation = async (conv: any) => {
  await chatMainRef.value?.deleteConversation(conv)
}

// 重命名对话
const renameConversation = async (conv: any) => {
  await chatMainRef.value?.renameConversation(conv)
}
</script>

<style scoped>
.ds-panel {
  flex-shrink: 0;
  display: flex;
  position: relative;
  height: 100%;
  border-radius: 12px;
  overflow: hidden;
  background: linear-gradient(180deg, rgba(22, 27, 34, 0.98) 0%, rgba(28, 35, 51, 0.92) 100%);
  border-left: 1px solid var(--border-primary, rgba(59, 130, 246, 0.10));
  box-shadow: -4px 0 24px rgba(0, 0, 0, 0.3);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans SC", sans-serif;
}

/* ---- 抽屉 ---- */
.ds-drawer {
  position: absolute;
  inset: 0;
  z-index: 20;
  background: linear-gradient(180deg, rgba(13, 17, 23, 0.98) 0%, rgba(22, 27, 34, 0.95) 100%);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--border-primary, rgba(59, 130, 246, 0.10));
}
.ds-drawer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border-primary, rgba(59, 130, 246, 0.10));
}
.ds-drawer-title {
  flex: 1;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary, #e6edf3);
}
.ds-drawer-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}
.ds-drawer-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  border-radius: 10px;
  cursor: pointer;
  transition: background 120ms;
  gap: 8px;
}
.ds-drawer-item:hover { background: rgba(59, 130, 246, 0.06); }
.ds-drawer-item.active {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.18) 0%, rgba(30, 58, 138, 0.12) 100%);
  border: 1px solid rgba(59, 130, 246, 0.2);
}
.ds-drawer-item-title {
  flex: 1;
  font-size: 13px;
  color: var(--text-secondary, #8b98a8);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
.ds-drawer-item-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.ds-drawer-item-meta {
  font-size: 11px;
  color: var(--text-muted, #4d5768);
}
.ds-drawer-item-btn {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text-muted, #4d5768);
  cursor: pointer;
  opacity: 0;
  transition: all 120ms;
}
.ds-drawer-item:hover .ds-drawer-item-btn {
  opacity: 0.6;
}
.ds-drawer-item-btn:hover {
  opacity: 1 !important;
  background: rgba(59, 130, 246, 0.1);
  color: #93c5fd;
}
.ds-drawer-item-btn--danger:hover {
  background: rgba(248, 113, 113, 0.1);
  color: #f87171 !important;
}
.ds-drawer-empty {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-muted, #4d5768);
  font-size: 13px;
}

/* ---- 通用按钮 ---- */
.ds-icon-btn {
  width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; border: none; background: transparent;
  color: var(--text-tertiary, #5f6b7a);
  cursor: pointer; transition: all 120ms;
}
.ds-icon-btn:hover {
  background: rgba(59, 130, 246, 0.08);
  color: #93c5fd;
}

/* ---- 动画 ---- */
.slide-drawer-enter-active,
.slide-drawer-leave-active {
  transition: transform 200ms cubic-bezier(0.4,0,0.2,1), opacity 200ms;
}
.slide-drawer-enter-from,
.slide-drawer-leave-to {
  transform: translateX(-100%);
  opacity: 0;
}

/* ---- 浅色主题（跟随全局 data-theme） ---- */
:root:not([data-theme='dark']) .ds-panel {
  --text-primary: #1e1e1c;
  --text-secondary: #666560;
  --text-muted: #9d9c96;
  --text-tertiary: #9d9c96;
  --border-primary: rgba(0, 0, 0, 0.08);
  --border-secondary: rgba(0, 0, 0, 0.12);
  background: #ffffff;
  border-left-color: #e4e2dd;
  box-shadow: none;
}
:root:not([data-theme='dark']) .ds-drawer {
  background: #faf9f7;
  border-right-color: #e4e2dd;
}
:root:not([data-theme='dark']) .ds-drawer-item:hover { background: rgba(0, 0, 0, 0.04); }
:root:not([data-theme='dark']) .ds-drawer-item.active {
  background: linear-gradient(135deg, rgba(42, 120, 214, 0.1) 0%, rgba(42, 120, 214, 0.05) 100%);
  border-color: rgba(42, 120, 214, 0.2);
}
:root:not([data-theme='dark']) .ds-drawer-item-btn:hover { background: rgba(42, 120, 214, 0.08); color: #1c5cab; }
:root:not([data-theme='dark']) .ds-drawer-item-btn--danger:hover { background: rgba(248, 113, 113, 0.1); color: #dc2626 !important; }
:root:not([data-theme='dark']) .ds-icon-btn:hover { background: rgba(42, 120, 214, 0.08); color: #1c5cab; }
</style>
