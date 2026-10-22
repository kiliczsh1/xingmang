<template>
  <div
    class="end-node"
    :class="{ selected, success: data.status === 'success' }"
  >
    <Handle type="target" :position="Position.Left" />
    <div class="node-header">
      <div class="node-icon">
        <i v-if="data.status === 'success'" class="fas fa-check" />
        <span v-else>🔴</span>
      </div>
      <div class="node-label">{{ data.label || '结束' }}</div>
    </div>
    <div v-if="data.config" class="node-body">
      <div v-if="data.config.outputFileName" class="config-item">
        <i class="fas fa-file" />
        <span>{{ data.config.outputFileName }}</span>
      </div>
      <div class="config-item">
        <i class="fas fa-file-code" />
        <span>{{ formatLabel(data.config.outputFormat) }}</span>
      </div>
      <div class="config-item">
        <i class="fas" :class="[data.config.autoSave ? 'fa-check-circle' : 'fa-save']" />
        <span>{{ data.config.autoSave ? '自动保存' : '保存对话框' }}</span>
      </div>
    </div>
    <NodeMetaBar :ms="data.durationMs" :breakpoint="!!data.breakpoint" />
  </div>
</template>

<script setup lang="ts">
import { Handle, Position } from '@vue-flow/core'
import type { WNodeData } from '@/types/workflow'
import NodeMetaBar from './NodeMetaBar.vue'

defineProps<{ id?: string; data: WNodeData; selected?: boolean }>()

const formatLabel = (t?: string) =>
  (({ md: 'Markdown', txt: '纯文本', json: 'JSON' }) as Record<string, string>)[t || ''] || 'Markdown'
</script>

<style scoped>
.end-node {
  min-width: 180px;
  max-width: 260px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid var(--el-color-danger, #f56c6c);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.end-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.end-node.selected {
  box-shadow: 0 0 0 4px rgba(245, 108, 108, 0.2);
}
.end-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: linear-gradient(135deg, rgba(245, 108, 108, 0.1), rgba(245, 108, 108, 0.05));
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
}
.node-icon {
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  flex-shrink: 0;
}
.node-icon i {
  color: var(--el-color-success, #67c23a);
}
.node-label {
  font-weight: 600;
  color: var(--wf-el-text-primary, #303133);
  font-size: 14px;
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
  color: var(--el-color-danger, #f56c6c);
  font-size: 10px;
  width: 14px;
  flex-shrink: 0;
}
.config-item span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
