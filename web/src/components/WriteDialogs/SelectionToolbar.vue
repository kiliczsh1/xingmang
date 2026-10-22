<template>
  <teleport to="body">
    <div
      v-if="visible"
      class="xm-sel-toolbar xm-pop"
      :style="{ left: x + 'px', top: y + 'px' }"
      @mousedown.prevent
    >
      <span class="xm-sel-toolbar__count">{{ count }} 字</span>
      <span class="xm-sel-toolbar__sep"></span>
      <button type="button" class="xm-sel-toolbar__btn" @click="emit('bold')">加粗</button>
      <button type="button" class="xm-sel-toolbar__btn xm-sel-toolbar__btn--glyph" title="斜体" @click="emit('format', 'italic')"><span class="xm-sel-toolbar__italic">I</span></button>
      <button type="button" class="xm-sel-toolbar__btn xm-sel-toolbar__btn--glyph" title="删除线" @click="emit('format', 'strikeThrough')"><span class="xm-sel-toolbar__strike">S</span></button>
      <span class="xm-sel-toolbar__sep"></span>
      <el-dropdown trigger="click" @command="(tag: string) => emit('format-block', tag)">
        <button type="button" class="xm-sel-toolbar__btn" title="标题">标题<el-icon class="xm-sel-toolbar__caret"><ArrowDown /></el-icon></button>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="p">正文</el-dropdown-item>
            <el-dropdown-item command="h1">标题 H1</el-dropdown-item>
            <el-dropdown-item command="h2">标题 H2</el-dropdown-item>
            <el-dropdown-item command="h3">标题 H3</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
      <button type="button" class="xm-sel-toolbar__btn" title="引用" @click="emit('format-block', 'blockquote')">引用</button>
      <button type="button" class="xm-sel-toolbar__btn" title="无序列表" @click="emit('format', 'insertUnorderedList')">无序</button>
      <button type="button" class="xm-sel-toolbar__btn" title="有序列表" @click="emit('format', 'insertOrderedList')">有序</button>
      <button type="button" class="xm-sel-toolbar__btn" title="行内代码" @click="emit('format', 'inlineCode')">代码</button>
      <span class="xm-sel-toolbar__sep"></span>
      <button type="button" class="xm-sel-toolbar__btn" @click="emit('highlight')">高亮</button>
      <button type="button" class="xm-sel-toolbar__btn" @click="emit('clear')">清除</button>
      <span class="xm-sel-toolbar__sep"></span>
      <button
        v-for="action in SELECTION_TOOL_ACTIONS"
        :key="action.id"
        type="button"
        class="xm-sel-toolbar__btn xm-sel-toolbar__btn--ai"
        @click="emit('tool', action.id)"
      >{{ action.label }}</button>
    </div>
  </teleport>
</template>

<script setup lang="ts">
import { SELECTION_TOOL_ACTIONS } from '@/constants/aiTools'
import { ArrowDown } from '@element-plus/icons-vue'

defineOptions({ name: 'SelectionToolbar' })

defineProps<{
  visible: boolean
  x: number
  y: number
  count: number
}>()

const emit = defineEmits<{
  (e: 'bold'): void
  (e: 'highlight'): void
  (e: 'clear'): void
  (e: 'format', kind: string): void
  (e: 'format-block', tag: string): void
  (e: 'tool', id: string): void
}>()
</script>

<style scoped>
.xm-sel-toolbar {
  position: fixed;
  z-index: 3000;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 4px 6px;
  background: var(--xm-surface, #ffffff);
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: var(--xm-radius-sm, 10px);
  box-shadow: var(--xm-shadow-pop, 0 4px 16px rgba(0, 0, 0, 0.12));
  white-space: nowrap;
}

.xm-sel-toolbar__count {
  font-size: 12px;
  color: var(--xm-accent, #5a8d85);
  padding: 0 6px;
  font-variant-numeric: tabular-nums;
}

.xm-sel-toolbar__sep {
  width: 1px;
  height: 16px;
  background: var(--xm-line, #e4e2dd);
  margin: 0 4px;
}

.xm-sel-toolbar__btn {
  height: 26px;
  padding: 0 8px;
  border: none;
  background: transparent;
  border-radius: 5px;
  font-size: 12.5px;
  color: var(--xm-ink-2, #5a6270);
  cursor: pointer;
  transition: all 0.2s var(--xm-ease, ease);
}

.xm-sel-toolbar__btn:hover {
  background: var(--xm-surface-3, #f3f5f7);
  color: var(--xm-ink, #1f2329);
}

.xm-sel-toolbar__btn--ai {
  color: #d97706;
  font-weight: 500;
}

.xm-sel-toolbar__btn--glyph {
  width: 26px;
  padding: 0;
  justify-content: center;
}

.xm-sel-toolbar__italic {
  font-style: italic;
  font-family: Georgia, serif;
}

.xm-sel-toolbar__strike {
  text-decoration: line-through;
}

.xm-sel-toolbar__caret {
  font-size: 11px;
  margin-left: 2px;
}

.xm-sel-toolbar__btn--ai:hover {
  background: #fff7ed;
  color: #b45309;
}

/* 壳/线/字的暗色翻转已由 --xm-* 令牌承担，手工暗色补丁删除；
   --ai 的橙色为业务色，暗色特调保留 */

:root[data-theme='dark'] .xm-sel-toolbar__btn--ai {
  color: #fbbf24;
}

:root[data-theme='dark'] .xm-sel-toolbar__btn--ai:hover {
  background: rgba(251, 191, 36, 0.12);
}
</style>
