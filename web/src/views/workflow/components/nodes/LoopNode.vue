<template>
  <div
    class="loop-node"
    :class="{
      selected,
      running: data.status === 'running',
      success: data.status === 'success',
      error: data.status === 'error'
    }"
  >
    <Handle type="target" :position="Position.Left" />
    <div class="node-header">
      <div class="header-left">
        <i class="fas fa-rotate" />
        <span>{{ data.label || '循环节点' }}</span>
      </div>
      <div class="header-right">
        <el-tag v-if="data.status === 'running'" type="warning" size="small">
          <i class="fas fa-spinner fa-spin" /> {{ currentIteration }}/{{ totalIterations }}
        </el-tag>
        <el-tag v-else-if="data.status === 'success'" type="success" size="small">
          <i class="fas fa-check" /> 完成
        </el-tag>
        <el-tag v-else-if="data.status === 'error'" type="danger" size="small">
          <i class="fas fa-times" /> 失败
        </el-tag>
      </div>
    </div>
    <div class="node-body">
      <div class="config-item">
        <i class="fas fa-repeat" />
        <span>循环 {{ totalIterations }} 次</span>
      </div>
      <div v-if="data.config?.loopVariable" class="config-item">
        <i class="fas fa-code" />
        <span>变量: {{ data.config.loopVariable }}</span>
      </div>
      <div v-if="data.config?.keepContext" class="config-item">
        <i class="fas fa-link" />
        <span>保持上下文连续</span>
      </div>
      <div v-if="data.totalChapters" class="config-item">
        <i class="fas fa-list-ol" />
        <span>计划章节：{{ data.totalChapters }}</span>
      </div>
      <div v-if="data.status === 'running' && currentIteration > 0" class="config-item">
        <i class="fas fa-book" />
        <span>当前：第 {{ currentIteration }} 章{{ data.currentChapterTitle ? `（${data.currentChapterTitle}）` : '' }}</span>
      </div>
      <div v-if="data.result" class="result-preview">
        <i class="fas fa-check-circle" />
        <span>已执行 {{ totalIterations }} 次</span>
      </div>
    </div>
    <Handle type="source" :position="Position.Right" id="loop-body" :style="{ top: '50%' }" />
    <Handle type="source" :position="Position.Right" id="loop-done" :style="{ top: '85%' }" />
    <NodeMetaBar :ms="data.durationMs" :breakpoint="!!data.breakpoint" />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import type { WNodeData } from '@/types/workflow'
import NodeMetaBar from './NodeMetaBar.vue'

const props = defineProps<{ id?: string; data: WNodeData; selected?: boolean }>()

const currentIteration = computed(() => props.data.currentIteration || 0)
const totalIterations = computed(() => props.data.iterations || props.data.config?.iterations || 1)
</script>

<style scoped>
.loop-node {
  min-width: 180px;
  max-width: 260px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid var(--el-color-warning, #e6a23c);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.loop-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.loop-node.selected {
  box-shadow: 0 0 0 4px rgba(230, 162, 60, 0.2);
}
.loop-node.running {
  border-color: var(--el-color-warning, #e6a23c);
  animation: wf-node-pulse 2s infinite;
}
.loop-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.loop-node.error {
  border-color: var(--el-color-danger, #f56c6c);
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: linear-gradient(135deg, rgba(230, 162, 60, 0.1), rgba(230, 162, 60, 0.05));
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
  gap: 8px;
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}
.header-left i {
  color: var(--el-color-warning, #e6a23c);
  font-size: 14px;
  flex-shrink: 0;
}
.header-left span {
  font-weight: 600;
  color: var(--wf-el-text-primary, #303133);
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.header-right {
  flex-shrink: 0;
}
.node-body {
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.config-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
}
.config-item i {
  color: var(--el-color-warning, #e6a23c);
  font-size: 10px;
  width: 14px;
  flex-shrink: 0;
}
.config-item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.result-preview {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  background: var(--wf-el-fill-lighter, #fafafa);
  border-radius: 4px;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
}
.result-preview i {
  color: var(--el-color-success, #67c23a);
  flex-shrink: 0;
}
</style>
