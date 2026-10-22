<template>
  <div class="workflow-logs-panel">
    <div class="logs-header">
      <span class="logs-title"><i class="fas fa-list" /> 执行日志</span>
      <el-button size="small" text :disabled="!logs.length" @click="$emit('clear')">
        <i class="fas fa-eraser" /> 清空
      </el-button>
    </div>
    <el-scrollbar ref="scrollRef" class="logs-content">
      <div v-if="!logs.length" class="logs-empty">
        <el-empty description="暂无日志" :image-size="60" />
      </div>
      <div v-else class="logs-list">
        <div v-for="(l, i) in logs" :key="l.id ?? i" class="log-item" :class="`log-${l.level || l.type}`">
          <span class="log-time">{{ l.timestamp || l.time }}</span>
          <span class="log-message">{{ l.message }}<span v-if="l.isStreaming" class="cursor">▍</span></span>
        </div>
      </div>
    </el-scrollbar>
  </div>
</template>

<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import type { LogEntry } from '@/types/workflow'

const props = defineProps<{ logs: LogEntry[] }>()
defineEmits<{ (e: 'clear'): void }>()

// el-scrollbar 的滚动容器在组件内部，直接用它 expose 的 setScrollTop
const scrollRef = ref<any>(null)

const scrollToEnd = async () => {
  await nextTick()
  try {
    scrollRef.value?.setScrollTop?.(1e7)
  } catch {
    /* 面板尚未挂载 */
  }
}

watch(
  () => props.logs.length,
  () => scrollToEnd()
)
watch(
  () => props.logs[props.logs.length - 1]?.message,
  () => scrollToEnd()
)
</script>

<style scoped>
/* 源版：画布右侧独立面板，350px，带左侧色条的日志条目 */
.workflow-logs-panel {
  width: 350px;
  background: var(--wf-el-bg, #ffffff);
  border-left: 1px solid var(--wf-el-border, #dcdfe6);
  display: flex;
  flex-direction: column;
}
.logs-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
  background: var(--wf-el-bg-page, #f2f3f5);
}
.logs-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--wf-el-text-primary, #303133);
}
.logs-title i {
  color: var(--el-color-primary, #409eff);
}
.logs-content {
  flex: 1;
  height: 0;
}
.logs-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 40px 20px;
}
.logs-list {
  padding: 8px;
}
.log-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 12px;
  margin-bottom: 8px;
  border-radius: 6px;
  border-left: 3px solid transparent;
  background: var(--wf-el-fill-lighter, #fafafa);
  font-size: 13px;
  transition: all 0.2s;
}
.log-item:hover {
  background: var(--wf-el-fill-light, #f5f7fa);
}
.log-time {
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  font-family: Consolas, Monaco, monospace;
}
.log-message {
  color: var(--wf-el-text-primary, #303133);
  word-break: break-word;
  white-space: pre-wrap;
}
.log-item.log-info {
  border-left-color: var(--el-color-info, #909399);
}
.log-item.log-success {
  border-left-color: var(--el-color-success, #67c23a);
  background: rgba(103, 194, 58, 0.08);
}
.log-item.log-warning {
  border-left-color: var(--el-color-warning, #e6a23c);
  background: rgba(230, 162, 60, 0.08);
}
.log-item.log-error {
  border-left-color: var(--el-color-danger, #f56c6c);
  background: rgba(245, 108, 108, 0.08);
}
.cursor {
  animation: wf-blink 1s steps(2, start) infinite;
}
@keyframes wf-blink {
  to {
    visibility: hidden;
  }
}
</style>
