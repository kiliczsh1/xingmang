<!-- ChatInputArea.vue —— 输入区（统一青绿主题色） -->
<template>
  <div class="ds-input-area">
    <!-- 引用标签 -->
    <div v-if="attachedReferences.length > 0" class="ds-ref-tags">
      <span
        v-for="ref in attachedReferences"
        :key="ref.id + ref.type"
        :class="['ds-ref-tag', `ds-ref-tag--${ref.type}`]"
      >
        {{ ref.label }}
        <button class="ds-ref-remove" @click="$emit('removeReference', ref)">&times;</button>
      </span>
    </div>

    <!-- 动态变量输入栏 -->
    <div v-if="parsedVariables.length > 0" class="ds-var-bar">
      <button class="ds-var-toggle" @click="varBarOpen = !varBarOpen">
        <svg :class="['ds-var-chevron', { open: varBarOpen }]" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"/></svg>
        动态变量
        <span class="ds-var-count">{{ filledCount }}/{{ parsedVariables.length }}</span>
      </button>
      <Transition name="ds-var-slide">
        <div v-if="varBarOpen" class="ds-var-fields">
          <div
            v-for="varName in parsedVariables"
            :key="varName"
            class="ds-var-row"
          >
            <span class="ds-var-label">{{ varName }}</span>
            <input
              class="ds-var-input"
              :value="variableValues[varName] || ''"
              @input="onVarInput(varName, ($event.target as HTMLInputElement).value)"
              :placeholder="`输入 ${varName}...`"
            />
          </div>
        </div>
      </Transition>
    </div>

    <div :class="['ds-input-box', { focused }]">
      <textarea
        ref="textareaRef"
        :value="modelValue"
        @input="$emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        @keydown.ctrl.enter="$emit('send')"
        @focus="focused = true"
        @blur="focused = false"
        :disabled="loading"
        placeholder="输入消息…  Ctrl+Enter 发送"
        rows="2"
        class="ds-textarea"
      />
      <div class="ds-input-toolbar">
        <!-- @ 引用 -->
        <button class="ds-toolbar-btn" @click="showAtMenu = !showAtMenu" title="@ 引用">
          <span class="ds-at-symbol">@</span>
        </button>
        <Transition name="ds-popover">
          <div v-if="showAtMenu" class="ds-at-popover">
            <button
              v-for="item in atMenuOptions"
              :key="item.type"
              class="ds-at-option"
              @click="$emit('atSelect', item.type); showAtMenu = false"
            >
              <span class="ds-at-dot" :style="{ background: item.color }" />
              <span class="ds-at-label">{{ item.label }}</span>
              <span class="ds-at-hint">{{ item.hint }}</span>
            </button>
          </div>
        </Transition>

        <div class="ds-toolbar-spacer" />

        <!-- 发送 / 停止 -->
        <button
          v-if="!loading"
          :class="['ds-send-btn', { disabled: !canSend }]"
          @click="$emit('send')"
          :disabled="!canSend"
          title="发送"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>
        </button>
        <button
          v-else
          class="ds-stop-btn"
          @click="$emit('stop')"
          title="停止生成"
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><rect x="6" y="6" width="12" height="12" rx="2"/></svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { RelatedContent } from '@/types'

const props = defineProps<{
  modelValue: string
  loading: boolean
  modelName: string
  attachedReferences: Array<RelatedContent & { tagType?: string; label?: string }>
  selectedPromptsCount: number
  relatedContentCount: number
  atMenuOptions: Array<{ type: string; label: string; hint: string; icon: any; color: string }>
  parsedVariables: string[]
  variableValues: Record<string, string>
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'send'): void
  (e: 'stop'): void
  (e: 'openModelSelect'): void
  (e: 'atSelect', type: string): void
  (e: 'removeReference', ref: any): void
  (e: 'updateVariable', name: string, value: string): void
}>()

const showAtMenu = ref(false)
const focused = ref(false)
const varBarOpen = ref(true)
const textareaRef = ref<HTMLTextAreaElement>()

const canSend = computed(() =>
  props.modelValue.trim() || props.selectedPromptsCount > 0 || props.relatedContentCount > 0
)

const filledCount = computed(() =>
  props.parsedVariables.filter(name => props.variableValues[name]?.trim()).length
)

const onVarInput = (name: string, value: string) => {
  emit('updateVariable', name, value)
}
</script>

<style scoped>
.ds-input-area {
  padding: 8px 16px 16px;
  background: rgba(30, 41, 59, 0.95);
  border-top: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
}

/* 引用标签 */
.ds-ref-tags {
  display: flex; flex-wrap: wrap; gap: 6px;
  padding: 0 4px 8px;
}
.ds-ref-tag {
  display: inline-flex; align-items: center; gap: 4px;
  padding: 2px 8px 2px 10px; border-radius: 6px;
  font-size: 11px; font-weight: 500;
}
.ds-ref-tag--chapter { background: rgba(94, 234, 212, 0.12); color: #5eead4; }
.ds-ref-tag--memo { background: rgba(96, 165, 250, 0.12); color: #60a5fa; }
.ds-ref-remove {
  width: 14px; height: 14px;
  display: flex; align-items: center; justify-content: center;
  border: none; background: transparent;
  color: inherit; opacity: 0.5; cursor: pointer;
  font-size: 14px; line-height: 1;
  border-radius: 50%; transition: all 120ms;
}
.ds-ref-remove:hover { opacity: 1; background: rgba(255,255,255,0.1); }

/* ---- 动态变量栏 ---- */
.ds-var-bar {
  padding: 0 4px 8px;
  border-bottom: 1px solid rgba(94, 234, 212, 0.06);
  margin-bottom: 8px;
}
.ds-var-toggle {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 8px;
  border: none;
  background: rgba(94, 234, 212, 0.06);
  color: var(--text-tertiary, #94a3b8);
  font-size: 12px;
  cursor: pointer;
  transition: all 120ms;
  margin-bottom: 4px;
}
.ds-var-toggle:hover {
  background: rgba(94, 234, 212, 0.1);
  color: #5eead4;
}
.ds-var-chevron {
  transition: transform 150ms;
}
.ds-var-chevron.open {
  transform: rotate(0deg);
}
.ds-var-chevron:not(.open) {
  transform: rotate(-90deg);
}
.ds-var-count {
  font-size: 11px;
  color: var(--text-muted, #6b7280);
}
.ds-var-fields {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0;
}
.ds-var-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.ds-var-label {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary, #c8d0e0);
  min-width: 60px;
  flex-shrink: 0;
}
.ds-var-input {
  flex: 1;
  padding: 6px 10px;
  border-radius: 8px;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  background: rgba(51, 65, 85, 0.5);
  color: var(--text-primary, #f4f7ff);
  font-size: 13px;
  outline: none;
  transition: border-color 150ms;
  font-family: inherit;
}
.ds-var-input:focus {
  border-color: rgba(94, 234, 212, 0.35);
  box-shadow: 0 0 0 2px rgba(94, 234, 212, 0.06);
}
.ds-var-input::placeholder {
  color: var(--text-muted, #6b7280);
  font-size: 12px;
}

/* 变量栏折叠动画 */
.ds-var-slide-enter-active,
.ds-var-slide-leave-active {
  transition: all 180ms cubic-bezier(0.4,0,0.2,1);
  overflow: hidden;
}
.ds-var-slide-enter-from,
.ds-var-slide-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}

/* 输入框 */
.ds-input-box {
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 14px;
  background: rgba(51, 65, 85, 0.5);
  transition: all 200ms;
  overflow: hidden;
  position: relative;
}
.ds-input-box.focused {
  border-color: rgba(94, 234, 212, 0.4);
  box-shadow: 0 0 0 2px rgba(94, 234, 212, 0.08), 0 4px 20px rgba(0,0,0,0.3);
}

.ds-textarea {
  display: block; width: 100%;
  padding: 12px 16px 4px;
  border: none; background: transparent;
  color: var(--text-primary, #f4f7ff);
  font-size: 14px; line-height: 1.6;
  resize: none; outline: none;
  font-family: inherit;
}
.ds-textarea::placeholder { color: var(--text-muted, #6b7280); }
.ds-textarea:disabled { opacity: 0.5; }

/* 底部工具栏 */
.ds-input-toolbar {
  display: flex; align-items: center; gap: 4px;
  padding: 4px 8px 8px;
}
.ds-toolbar-spacer { flex: 1; }

/* @ 按钮 */
.ds-toolbar-btn {
  width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; border: none; background: transparent;
  color: var(--text-muted, #6b7280);
  cursor: pointer; transition: all 120ms;
}
.ds-toolbar-btn:hover {
  background: rgba(94, 234, 212, 0.08);
  color: #5eead4;
}
.ds-at-symbol { font-weight: 800; font-size: 15px; }

/* @ 弹出菜单 */
.ds-at-popover {
  position: absolute;
  bottom: calc(100% + 4px); left: 8px;
  min-width: 200px;
  background: #1e293b;
  border: 1px solid rgba(94, 234, 212, 0.18);
  border-radius: 10px;
  padding: 4px;
  box-shadow: 0 8px 30px rgba(0,0,0,0.5);
  z-index: 30;
}
.ds-at-option {
  display: flex; align-items: center; gap: 8px;
  width: 100%; padding: 8px 10px;
  border-radius: 8px; border: none; background: transparent;
  color: var(--text-secondary, #c8d0e0);
  font-size: 13px; cursor: pointer;
  transition: background 120ms; text-align: left;
}
.ds-at-option:hover { background: rgba(94, 234, 212, 0.08); color: #5eead4; }
.ds-at-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.ds-at-label { flex: 1; }
.ds-at-hint { font-size: 11px; color: var(--text-muted, #6b7280); }

/* 发送按钮 */
.ds-send-btn {
  width: 32px; height: 32px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 10px; border: none;
  background: linear-gradient(135deg, #0f766e 0%, #115e59 100%);
  color: #fff; cursor: pointer; transition: all 150ms;
  box-shadow: 0 4px 12px rgba(15, 118, 110, 0.3);
}
.ds-send-btn:hover:not(.disabled) {
  background: linear-gradient(135deg, #115e59 0%, #134e4a 100%);
  transform: scale(1.05);
  box-shadow: 0 6px 18px rgba(15, 118, 110, 0.4);
}
.ds-send-btn.disabled {
  opacity: 0.3; cursor: not-allowed; box-shadow: none;
}

/* 停止按钮 */
.ds-stop-btn {
  width: 32px; height: 32px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 10px; border: none;
  background: rgba(248, 113, 113, 0.12);
  color: #f87171; cursor: pointer; transition: all 150ms;
}
.ds-stop-btn:hover { background: rgba(248, 113, 113, 0.2); }

/* 弹出菜单动画 */
.ds-popover-enter-active,
.ds-popover-leave-active {
  transition: all 150ms cubic-bezier(0.4,0,0.2,1);
}
.ds-popover-enter-from,
.ds-popover-leave-to {
  opacity: 0; transform: translateY(4px) scale(0.97);
}
</style>
