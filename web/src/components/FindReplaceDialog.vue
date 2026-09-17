<template>
  <Teleport to="body">
    <Transition name="fr-slide">
      <div v-if="visible" class="fr-bar xm-pop">
        <!-- 查找输入 -->
        <div class="fr-input-wrap">
          <svg class="fr-ico" viewBox="0 0 24 24"><circle cx="10.5" cy="10.5" r="6"/><path d="M15 15l5.5 5.5"/></svg>
          <input
            ref="findInputRef"
            :value="findText"
            class="fr-input"
            placeholder="查找"
            @input="onFindInput(($event.target as HTMLInputElement).value)"
            @compositionstart="onCompositionStart"
            @compositionend="onCompositionEnd"
            @keydown.enter.prevent="onFindEnter($event)"
          />
          <span class="fr-count" :class="{ zero: findText && matchCount === 0 }">
            {{ findText ? `${matchIndex + 1}/${matchCount}` : '' }}
          </span>
        </div>

        <!-- 替换输入 -->
        <div v-if="mode === 'replace'" class="fr-input-wrap sub">
          <svg class="fr-ico" viewBox="0 0 24 24"><path d="M4 17l7-10M8 17l4-5.5M15 8h5M15 8l2-2M15 8l2 2M15 14h5M15 14l2-2M15 14l2 2"/></svg>
          <input
            :value="replaceText"
            class="fr-input"
            placeholder="替换为"
            @input="emit('update:replace-text', ($event.target as HTMLInputElement).value)"
            @keydown.enter.prevent="onReplaceEnter($event)"
          />
        </div>

        <!-- 操作 -->
        <div class="fr-acts">
          <button class="fr-btn" :disabled="!findText" title="上一个 (Shift+Enter)" @click="emit('find-prev')">
            <svg viewBox="0 0 24 24"><path d="M12 19V5M6 11l6-6 6 6"/></svg>
          </button>
          <button class="fr-btn" :disabled="!findText" title="下一个 (Enter)" @click="emit('find-next')">
            <svg viewBox="0 0 24 24"><path d="M12 5v14M6 13l6 6 6-6"/></svg>
          </button>
          <template v-if="mode === 'replace'">
            <span class="fr-sep"></span>
            <button class="fr-btn txt" :disabled="!findText" @click="emit('replace-one')">替换</button>
            <button class="fr-btn txt" :disabled="!findText" @click="emit('replace-all')">全部</button>
          </template>
          <span class="fr-sep"></span>
          <label class="fr-cb" title="区分大小写">
            <input type="checkbox" :checked="caseSensitive" @change="emit('update:case-sensitive', ($event.target as HTMLInputElement).checked)" />
            <span>Aa</span>
          </label>
        </div>

        <button class="fr-x" title="关闭 (Esc)" @click="emit('update:visible', false)">
          <svg viewBox="0 0 24 24"><path d="M6 6l12 12M18 6L6 18"/></svg>
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  mode: 'find' | 'replace'
  findText: string
  replaceText: string
  caseSensitive: boolean
  matchCount: number
  matchIndex: number
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'update:find-text', v: string): void
  (e: 'update:replace-text', v: string): void
  (e: 'update:case-sensitive', v: boolean): void
  (e: 'find-next'): void
  (e: 'find-prev'): void
  (e: 'replace-one'): void
  (e: 'replace-all'): void
}>()

const findInputRef = ref<HTMLInputElement>()

/* ---- 输入法（IME）组合期间不触发查找 ----
   拼音输入时每个字母都会触发 input 事件，若此时去查找并让编辑器抢焦点，
   组合会被打断（表现为"输入一个拼音字母就中断"）。
   组合结束（compositionend）后才用最终文本触发查找。 */
const composing = ref(false)
let lastEmittedFind = ''

const emitFindText = (v: string) => {
  if (v === lastEmittedFind) return // 部分浏览器组合结束后还会补发一次 input，去重
  lastEmittedFind = v
  emit('update:find-text', v)
}

const onCompositionStart = () => {
  composing.value = true
}

const onCompositionEnd = () => {
  composing.value = false
  emitFindText(findInputRef.value?.value ?? '')
}

const onFindInput = (v: string) => {
  if (composing.value) return // 组合中：只接受 input，不更新模型、不查找
  emitFindText(v)
}

const onFindEnter = (e: KeyboardEvent) => {
  if (e.isComposing) return // 输入法回车确认候选词，不触发查找
  if (e.shiftKey) emit('find-prev')
  else emit('find-next')
}

const onReplaceEnter = (e: KeyboardEvent) => {
  if (e.isComposing) return
  emit('replace-one')
}

watch(
  () => props.visible,
  (v) => {
    if (v) nextTick(() => findInputRef.value?.focus())
  }
)

// 切换"区分大小写"后立即按新规则重新查找（记事本行为）
watch(
  () => props.caseSensitive,
  () => {
    if (props.visible && props.findText) emit('find-next')
  }
)
</script>

<style scoped>
/* 极简查找替换条 — 居中顶端、工具栏之下 */
.fr-bar {
  position: fixed;
  top: 52px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 3000;
  display: flex;
  align-items: center;
  gap: 0;
  background: var(--xm-surface, #ffffff);
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: var(--xm-radius-sm, 10px);
  box-shadow: var(--xm-shadow-pop, 0 6px 24px rgba(0, 0, 0, 0.1));
  overflow: hidden;
  padding: 4px;
}

/* 查找/替换输入区 */
.fr-input-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 8px;
  height: 30px;
  min-width: 180px;
  border-radius: 6px;
  transition: background 0.2s var(--xm-ease, ease);
}
.fr-input-wrap.sub {
  border-top: 1px solid var(--xm-line, #f0efe9);
  border-radius: 0 0 6px 6px;
}
.fr-input-wrap:focus-within {
  background: var(--xm-surface-2, #f5f4f1);
}
.fr-ico {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  fill: none;
  stroke: var(--xm-ink-3, #9d9c96);
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.fr-input {
  flex: 1;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  color: var(--xm-ink, #1e1e1c);
  font-family: inherit;
  min-width: 0;
}
.fr-input::placeholder {
  color: var(--xm-ink-3, #b8b6b0);
}
.fr-count {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--xm-ink-3, #9d9c96);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.fr-count.zero {
  color: #e5484d;
}

/* 操作按钮 */
.fr-acts {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 0 4px 0 6px;
  border-left: 1px solid var(--xm-line, #f0efe9);
}
.fr-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--xm-ink-2, #666560);
  cursor: pointer;
  transition: background 0.2s var(--xm-ease, ease), color 0.2s var(--xm-ease, ease);
}
.fr-btn svg {
  width: 14px;
  height: 14px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}
.fr-btn.txt {
  width: auto;
  padding: 0 10px;
  font-size: 12px;
  font-family: inherit;
}
.fr-btn:hover:not(:disabled) {
  background: var(--xm-surface-3, #eceae6);
  color: var(--xm-ink, #1e1e1c);
}
.fr-btn:active:not(:disabled) {
  transform: scale(.92);
}
.fr-btn:disabled {
  opacity: .35;
  cursor: default;
}
.fr-sep {
  width: 1px;
  height: 16px;
  background: var(--xm-line, #f0efe9);
  margin: 0 3px;
  flex-shrink: 0;
}

/* 区分大小写 Aa 切换 */
.fr-cb {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  user-select: none;
  padding: 0 6px;
  height: 28px;
  border-radius: 6px;
  transition: background 0.2s var(--xm-ease, ease);
}
.fr-cb:hover {
  background: var(--xm-surface-3, #eceae6);
}
.fr-cb input {
  display: none;
}
.fr-cb span {
  font-size: 11px;
  font-weight: 600;
  color: var(--xm-ink-3, #b8b6b0);
  letter-spacing: .02em;
}
.fr-cb:has(input:checked) span {
  color: var(--xm-accent, #1c5cab);
}

/* 关闭按钮 */
.fr-x {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--xm-ink-3, #9d9c96);
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.2s var(--xm-ease, ease), color 0.2s var(--xm-ease, ease);
}
.fr-x svg {
  width: 14px;
  height: 14px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
}
.fr-x:hover {
  background: var(--xm-surface-3, #eceae6);
  color: var(--xm-ink, #1e1e1c);
}

/* 入场/退场动画 */
.fr-slide-enter-active {
  transition: opacity 0.2s var(--xm-ease, ease), transform 0.2s var(--xm-ease, ease);
}
.fr-slide-leave-active {
  transition: opacity 0.2s var(--xm-ease, ease), transform 0.2s var(--xm-ease, ease);
}
.fr-slide-enter-from {
  opacity: 0;
  transform: translateX(-50%) translateY(-8px);
}
.fr-slide-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-4px);
}

/* ============ 暗色 ============
   壳/线/字/强调已由 --xm-* 双态令牌自动翻转，手工暗色补丁整体删除；
   .fr-count.zero 的红为业务语义色，亮暗同值保留。 */
</style>