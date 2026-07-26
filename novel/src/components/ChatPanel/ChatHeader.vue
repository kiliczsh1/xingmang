<!-- ChatHeader.vue —— 顶栏（统一青绿主题色） -->
<template>
  <div class="ds-header">
    <div class="ds-header-left">
      <button class="ds-icon-btn" @click="$emit('toggleDrawer')" title="对话列表">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
          <rect x="3" y="3" width="18" height="18" rx="3"/>
          <line x1="9" y1="8" x2="15" y2="8"/>
          <line x1="9" y1="12" x2="15" y2="12"/>
          <line x1="9" y1="16" x2="13" y2="16"/>
        </svg>
      </button>
      <span class="ds-header-model" @click="$emit('showModelSelect')">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="opacity:0.5"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
        {{ modelName || '选择模型' }}
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="opacity:0.3"><polyline points="6 9 12 15 18 9"/></svg>
      </span>
    </div>

    <div class="ds-header-center">
      <button
        v-for="btn in centerButtons"
        :key="btn.key"
        :class="['ds-header-btn', { active: btn.active }]"
        @click="$emit(btn.emit as any)"
        :title="btn.title"
      >
        <span v-if="btn.badge" class="ds-btn-badge">{{ btn.badge }}</span>
        <component :is="btn.icon" />
      </button>
    </div>

    <div class="ds-header-right">
      <div class="ds-worldbook-toggle" :title="worldBookLinked ? '世界书已关联' : '世界书未关联'">
        <span class="ds-wb-label">{{ worldBookLinked ? '联' : '断' }}</span>
        <button
          :class="['ds-wb-switch', { on: worldBookLinked }]"
          @click="$emit('toggleWorldBookLink', !worldBookLinked)"
        >
          <span class="ds-wb-knob" />
        </button>
      </div>
      <button class="ds-icon-btn" @click="$emit('showFullPrompt')" title="查看完整提示词">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
      </button>
      <button class="ds-icon-btn ds-close-btn" @click="$emit('close')" title="收起面板">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Document, Reading, Notebook, Setting } from '@element-plus/icons-vue'

const props = defineProps<{
  promptsCount: number
  relatedCount: number
  modelName: string
  worldBookLinked: boolean
  worldBookEnabled: boolean
}>()

defineEmits<{
  (e: 'toggleDrawer'): void
  (e: 'newConversation'): void
  (e: 'showConversationList'): void
  (e: 'showPromptSelect'): void
  (e: 'showWorldBook'): void
  (e: 'showRelateContent'): void
  (e: 'showModelSelect'): void
  (e: 'showRegexFilter'): void
  (e: 'toggleWorldBookLink', value: boolean): void
  (e: 'showFullPrompt'): void
  (e: 'close'): void
}>()

const centerButtons = computed(() => {
  const btns = [
    { key: 'prompt', title: '提示词', icon: Document, emit: 'showPromptSelect', badge: props.promptsCount || null, active: props.promptsCount > 0 },
    { key: 'worldbook', title: '世界书', icon: Reading, emit: 'showWorldBook', badge: props.relatedCount || null, active: false, show: props.worldBookEnabled },
    { key: 'relate', title: '关联内容', icon: Notebook, emit: 'showRelateContent', badge: null, active: false },
    { key: 'regex', title: '正则过滤', icon: Setting, emit: 'showRegexFilter', badge: null, active: false },
  ]
  return btns.filter(b => b.show !== false)
})
</script>

<style scoped>
.ds-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  gap: 8px;
  min-height: 44px;
  background: linear-gradient(135deg, rgba(15, 23, 42, 0.6) 0%, rgba(30, 41, 59, 0.4) 100%);
}
.ds-header-left,
.ds-header-right {
  display: flex;
  align-items: center;
  gap: 6px;
}
.ds-header-center {
  display: flex;
  align-items: center;
  gap: 2px;
}

/* 模型名 */
.ds-header-model {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 10px;
  border-radius: 8px;
  font-size: 12px;
  color: var(--text-tertiary, #94a3b8);
  cursor: pointer;
  transition: all 120ms;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
}
.ds-header-model:hover {
  color: #5eead4;
  border-color: rgba(94, 234, 212, 0.25);
  background: rgba(94, 234, 212, 0.05);
}

/* 通用图标按钮 */
.ds-icon-btn {
  width: 30px; height: 30px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; border: none; background: transparent;
  color: var(--text-tertiary, #94a3b8);
  cursor: pointer; transition: all 120ms;
}
.ds-icon-btn:hover {
  background: rgba(94, 234, 212, 0.08);
  color: #5eead4;
}
.ds-close-btn:hover {
  background: rgba(248, 113, 113, 0.1);
  color: #f87171;
}

/* 中间功能按钮 */
.ds-header-btn {
  position: relative;
  width: 30px; height: 30px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; border: none; background: transparent;
  color: var(--text-muted, #6b7280);
  cursor: pointer; transition: all 120ms; font-size: 15px;
}
.ds-header-btn:hover {
  background: rgba(94, 234, 212, 0.06);
  color: #5eead4;
}
.ds-header-btn.active {
  color: #5eead4;
  background: rgba(94, 234, 212, 0.08);
}
.ds-btn-badge {
  position: absolute;
  top: 2px; right: 2px;
  min-width: 14px; height: 14px;
  font-size: 9px; font-weight: 700;
  line-height: 14px; text-align: center;
  border-radius: 7px;
  background: linear-gradient(135deg, #0f766e 0%, #115e59 100%);
  color: #fff;
  padding: 0 3px;
}

/* 世界书开关 */
.ds-worldbook-toggle {
  display: flex;
  align-items: center;
  gap: 5px;
}
.ds-wb-label {
  font-size: 10px;
  color: var(--text-muted, #6b7280);
  font-weight: 600;
}
.ds-wb-switch {
  width: 30px; height: 16px;
  border-radius: 8px; border: none;
  background: rgba(255,255,255,0.08);
  cursor: pointer; position: relative;
  transition: background 150ms; padding: 0;
}
.ds-wb-switch.on { background: rgba(16,185,129,0.4); }
.ds-wb-knob {
  position: absolute; top: 2px; left: 2px;
  width: 12px; height: 12px; border-radius: 50%;
  background: var(--text-muted, #6b7280);
  transition: all 150ms;
}
.ds-wb-switch.on .ds-wb-knob {
  left: 16px;
  background: #5eead4;
  box-shadow: 0 0 6px rgba(94, 234, 212, 0.4);
}
</style>