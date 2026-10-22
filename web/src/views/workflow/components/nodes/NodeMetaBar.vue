<template>
  <div v-if="breakpoint || ms" class="node-meta">
    <span v-if="breakpoint" class="node-bp" title="已设为断点">
      <i class="fas fa-circle-pause" />
      断点
    </span>
    <span v-if="ms" class="node-duration" :title="`本节点耗时 ${text}`">
      <i class="fas fa-clock" />
      {{ text }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ ms?: number | null; breakpoint?: boolean }>()

const text = computed(() => {
  const ms = Number(props.ms ?? 0)
  if (!ms || ms < 0) return ''
  if (ms < 1000) return `${Math.round(ms)}ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(1)}s`
  const m = Math.floor(s / 60)
  const rest = Math.round(s % 60)
  return `${m}m${rest}s`
})
</script>

<style scoped>
.node-meta {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  padding: 0 12px 8px;
  font-size: 10px;
  font-variant-numeric: tabular-nums;
  user-select: none;
}
.node-duration {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--wf-el-text-secondary, #909399);
}
.node-bp {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 1px 5px;
  border-radius: 3px;
  color: var(--el-color-danger, #f56c6c);
  background: rgba(245, 108, 108, 0.12);
}
.node-meta i {
  font-size: 9px;
  opacity: 0.85;
}
</style>
