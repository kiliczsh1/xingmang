<template>
  <div
    class="context-node"
    :class="{
      selected,
      running: data.status === 'running',
      success: data.status === 'success',
      error: data.status === 'error'
    }"
  >
    <Handle type="target" :position="Position.Left" />
    <Handle type="source" :position="Position.Right" />
    <div class="node-header">
      <div class="header-left">
        <i class="fas fa-database" />
        <span>{{ data.label || '上下文节点' }}</span>
      </div>
      <div class="header-right">
        <el-tag v-if="data.status === 'running'" type="warning" size="small">
          <i class="fas fa-spinner fa-spin" /> 运行中
        </el-tag>
        <el-tag v-else-if="data.status === 'success'" type="success" size="small">
          <i class="fas fa-check" /> 成功
        </el-tag>
        <el-tag v-else-if="data.status === 'error'" type="danger" size="small">
          <i class="fas fa-times" /> 失败
        </el-tag>
      </div>
    </div>
    <div class="node-body">
      <div class="config-item">
        <i class="fas" :class="[actionIcon(data.config?.action)]" />
        <span>{{ actionLabel(data.config?.action) }}</span>
      </div>
      <div v-if="data.config?.variableName" class="config-item">
        <i class="fas fa-tag" />
        <span>{{ data.config.variableName }}</span>
      </div>
      <div v-if="data.config?.sourceVariable" class="config-item">
        <i class="fas fa-arrow-right" />
        <span>来源: {{ data.config.sourceVariable }}</span>
      </div>
      <div v-if="data.config?.action === 'append' && data.config?.maxHistory" class="config-item">
        <i class="fas fa-list-ol" />
        <span>最多 {{ data.config.maxHistory }} 条</span>
      </div>
      <div v-if="data.result" class="result-preview">
        <i class="fas fa-check-circle" />
        <span>操作完成</span>
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

const actionLabel = (a?: string) =>
  (({ save: '保存变量', load: '加载变量', clear: '清除变量', append: '追加历史' }) as Record<string, string>)[
    a || ''
  ] || '保存变量'

const actionIcon = (a?: string) =>
  (({
    save: 'fas fa-save',
    load: 'fas fa-download',
    clear: 'fas fa-eraser',
    append: 'fas fa-plus-circle'
  }) as Record<string, string>)[a || ''] || 'fas fa-save'
</script>

<style scoped>
.context-node {
  min-width: 180px;
  max-width: 260px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid #6366f1;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.context-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.context-node.selected {
  box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.2);
}
.context-node.running {
  border-color: var(--el-color-warning, #e6a23c);
  animation: wf-node-pulse 2s infinite;
}
.context-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.context-node.error {
  border-color: var(--el-color-danger, #f56c6c);
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(99, 102, 241, 0.05));
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
  color: #6366f1;
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
  color: #6366f1;
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
