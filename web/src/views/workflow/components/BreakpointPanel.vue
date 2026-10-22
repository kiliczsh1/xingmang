<template>
  <div class="bp-panel">
    <div class="bp-head">
      <i class="fas fa-circle-pause" />
      <span>{{ title }}</span>
      <span class="bp-node">{{ info?.label || info?.nodeId }}</span>
    </div>
    <div class="bp-body">
      <div class="bp-hint">当前上下文变量（截断显示）</div>
      <el-scrollbar class="bp-ctx">
        <div v-if="entries.length" class="bp-ctx-list">
          <div v-for="[k, v] in entries" :key="k" class="bp-ctx-item">
            <span class="bp-key">{{ k }}</span>
            <span class="bp-val">{{ v }}</span>
          </div>
        </div>
        <div v-else class="bp-empty">上下文为空</div>
      </el-scrollbar>
    </div>
    <div class="bp-foot">
      <el-button size="small" type="primary" @click="emit('continue')">
        <i class="fas fa-play" /> 继续执行
      </el-button>
      <el-button size="small" @click="emit('stop')">
        <i class="fas fa-stop" /> 停止
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  info: { nodeId?: string; label?: string; context?: Record<string, any>; reason?: string } | null
}>()
const emit = defineEmits<{ (e: 'continue'): void; (e: 'stop'): void }>()

/** 挂起原因：断点 / 分步按钮 / 节点生成被停止 */
const title = computed(() =>
  props.info?.reason === 'stopped' ? '节点生成已停止 · 已暂停' : '命中断点 · 已暂停'
)

const entries = computed<[string, string][]>(() =>
  Object.entries(props.info?.context || {}).map(([k, v]) => [
    k,
    typeof v === 'string' ? v : JSON.stringify(v)
  ])
)
</script>

<style scoped>
.bp-panel {
  position: absolute;
  /* 靠右下：左下角被缩放角标占着 */
  right: 16px;
  bottom: 16px;
  z-index: 25;
  width: 380px;
  max-height: 60%;
  display: flex;
  flex-direction: column;
  background: var(--wf-el-bg-overlay, #ffffff);
  border: 1px solid var(--el-color-warning, #e6a23c);
  border-radius: 8px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.25);
}
.bp-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  font-size: 13px;
  color: var(--el-color-warning, #e6a23c);
  border-bottom: 1px solid var(--wf-el-border-light, #ebeef5);
}
.bp-node {
  margin-left: auto;
  font-size: 12px;
  color: var(--wf-el-text-primary, #303133);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bp-body {
  flex: 1;
  min-height: 0;
  padding: 8px 12px;
}
.bp-hint {
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  margin-bottom: 6px;
}
.bp-ctx {
  max-height: 220px;
}
.bp-ctx-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.bp-ctx-item {
  display: flex;
  gap: 8px;
  font-size: 11px;
  font-family: Consolas, Monaco, monospace;
  padding: 4px 6px;
  background: var(--wf-el-fill-lighter, #fafafa);
  border-radius: 4px;
}
.bp-key {
  color: var(--el-color-primary, #409eff);
  flex-shrink: 0;
}
.bp-val {
  color: var(--wf-el-text-secondary, #909399);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.bp-empty {
  font-size: 11px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
  padding: 8px 0;
}
.bp-foot {
  display: flex;
  gap: 8px;
  padding: 10px 12px;
  border-top: 1px solid var(--wf-el-border-light, #ebeef5);
}
</style>
