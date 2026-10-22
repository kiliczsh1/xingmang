<template>
  <div class="start-node" :class="{ selected, running: data.status === 'running', success: data.status === 'success' }">
    <Handle type="source" :position="Position.Right" />
    <div class="node-content">
      <div class="node-icon">
        <i v-if="data.status === 'running'" class="fas fa-spinner fa-spin" />
        <i v-else-if="data.status === 'success'" class="fas fa-check" />
        <span v-else>🟢</span>
      </div>
      <div class="node-label">{{ data.label || '开始' }}</div>
    </div>
    <NodeMetaBar :ms="data.durationMs" :breakpoint="!!data.breakpoint" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'
import type { WNodeData } from '@/types/workflow'
import NodeMetaBar from './NodeMetaBar.vue'

defineProps<{ id?: string; data: WNodeData; selected?: boolean }>()
</script>

<style scoped>
.start-node {
  min-width: 120px;
  padding: 12px 16px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid var(--el-color-success, #67c23a);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.start-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.start-node.selected {
  box-shadow: 0 0 0 4px rgba(103, 194, 58, 0.2);
}
.start-node.running {
  border-color: var(--el-color-warning, #e6a23c);
  animation: wf-node-pulse 2s infinite;
}
.start-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.node-content {
  display: flex;
  align-items: center;
  gap: 8px;
}
.node-icon {
  font-size: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
}
.node-icon i {
  color: var(--el-color-success, #67c23a);
}
.node-label {
  font-weight: 600;
  color: var(--wf-el-text-primary, #303133);
  font-size: 14px;
}
</style>
