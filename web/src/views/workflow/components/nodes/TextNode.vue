<template>
  <div
    class="text-node"
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
        <i class="fas fa-file-text" />
        <span>{{ data.label || '文本节点' }}</span>
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
      <div v-if="data.config?.text" class="text-preview">
        <i class="fas fa-align-left" />
        <span>{{ String(data.config.text).substring(0, 50) }}{{ String(data.config.text).length > 50 ? '...' : '' }}</span>
      </div>
      <div v-else class="warning-info">
        <i class="fas fa-exclamation-triangle" />
        <span>未设置文本内容</span>
      </div>
      <div v-if="hasVariableRef(data.config?.text)" class="variable-info">
        <i class="fas fa-code" />
        <span>包含变量引用</span>
      </div>
      <div v-if="data.result" class="result-preview">
        <i class="fas fa-check-circle" />
        <span>{{ String(data.result).substring(0, 40) }}...</span>
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

const hasVariableRef = (text?: string) => (text ? /\{\{.+?\}\}/.test(text) : false)
</script>

<style scoped>
.text-node {
  min-width: 200px;
  max-width: 300px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid #909399;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.text-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.text-node.selected {
  box-shadow: 0 0 0 4px rgba(144, 147, 153, 0.2);
}
.text-node.running {
  border-color: var(--el-color-warning, #e6a23c);
  animation: wf-node-pulse 2s infinite;
}
.text-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.text-node.error {
  border-color: var(--el-color-danger, #f56c6c);
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: linear-gradient(135deg, rgba(144, 147, 153, 0.1), rgba(144, 147, 153, 0.05));
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
  color: #909399;
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
  gap: 8px;
}
.text-preview {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  line-height: 1.4;
  word-break: break-word;
}
.text-preview i {
  margin-top: 2px;
  flex-shrink: 0;
  color: #909399;
  font-size: 10px;
}
.warning-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--el-color-warning, #e6a23c);
  padding: 4px 8px;
  background: rgba(230, 162, 60, 0.1);
  border-radius: 4px;
}
.warning-info i {
  color: var(--el-color-warning, #e6a23c);
}
.variable-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--el-color-info, #909399);
  padding: 4px 8px;
  background: rgba(144, 147, 153, 0.1);
  border-radius: 4px;
}
.variable-info i {
  color: #909399;
}
.result-preview {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 8px;
  background: var(--wf-el-fill-lighter, #fafafa);
  border-radius: 4px;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  line-height: 1.4;
  word-break: break-word;
}
.result-preview i {
  color: var(--el-color-success, #67c23a);
  margin-top: 2px;
  flex-shrink: 0;
}
</style>
