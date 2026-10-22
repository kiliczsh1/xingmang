<template>
  <div
    class="ai-node"
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
        <i class="fas fa-robot" />
        <span>{{ data.label || 'AI 处理' }}</span>
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
      <div v-if="data.config?.modelId || data.config?.configId" class="model-info">
        <i class="fas fa-brain" />
        <span>{{ modelName(data.config?.modelId) }}</span>
      </div>
      <div v-else class="warning-info">
        <i class="fas fa-exclamation-triangle" />
        <span>未设置模型</span>
      </div>
      <div v-if="data.config?.systemPrompt" class="system-prompt-preview">
        <i class="fas fa-user-tie" />
        <span>系统: {{ data.config.systemPrompt.substring(0, 30) }}{{ data.config.systemPrompt.length > 30 ? '...' : '' }}</span>
      </div>
      <div v-if="data.config?.userPrompt || data.config?.prompt" class="prompt-preview">
        <i class="fas fa-comment-dots" />
        {{ (data.config.userPrompt || data.config.prompt).substring(0, 40) }}{{ (data.config.userPrompt || data.config.prompt).length > 40 ? '...' : '' }}
      </div>
      <div v-else class="warning-info">
        <i class="fas fa-exclamation-triangle" />
        <span>未设置提示词</span>
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

/** 源版：providerId::modelName 只展示模型名 */
const modelName = (modelId?: string) => {
  if (!modelId) return '未设置'
  const parts = modelId.split('::')
  return parts.length > 1 ? parts[1] : modelId
}
</script>

<style scoped>
.ai-node {
  min-width: 200px;
  max-width: 300px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid var(--el-color-primary, #409eff);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.ai-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.ai-node.selected {
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.2);
}
.ai-node.running {
  border-color: var(--el-color-warning, #e6a23c);
  animation: wf-node-pulse 2s infinite;
}
.ai-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.ai-node.error {
  border-color: var(--el-color-danger, #f56c6c);
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.1), rgba(64, 158, 255, 0.05));
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
  color: var(--el-color-primary, #409eff);
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
.model-info {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--wf-el-text-secondary, #909399);
}
.model-info i {
  color: var(--el-color-primary, #409eff);
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
.system-prompt-preview,
.prompt-preview {
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  line-height: 1.4;
  word-break: break-word;
  display: flex;
  align-items: flex-start;
  gap: 6px;
}
.system-prompt-preview i,
.prompt-preview i {
  margin-top: 2px;
  flex-shrink: 0;
  color: var(--el-color-primary, #409eff);
  font-size: 10px;
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
