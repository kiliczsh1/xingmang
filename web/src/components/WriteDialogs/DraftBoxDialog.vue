<template>
  <el-dialog
    :model-value="visible"
    title="草稿箱"
    width="560px"
    append-to-body
    destroy-on-close
    class="xm-draft-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="xm-df-body">
      <p class="xm-df-hint">自动保存本章的历史快照（每 90 秒，最多保留 10 份），可随时恢复</p>
      <div v-if="!snapshots.length" class="xm-df-empty">暂无草稿快照</div>
      <div v-for="(s, i) in snapshots" :key="i" class="xm-df-row">
        <div class="xm-df-info">
          <span class="xm-df-time">{{ s.time }}</span>
          <span class="xm-df-count">{{ s.length }} 字</span>
        </div>
        <div class="xm-df-actions">
          <el-button size="small" text @click="previewIndex = previewIndex === i ? -1 : i">
            {{ previewIndex === i ? '收起预览' : '预览' }}
          </el-button>
          <el-button size="small" type="warning" @click="restore(i)">恢复</el-button>
          <el-button size="small" type="danger" text @click="remove(i)">删除</el-button>
        </div>
        <div v-if="previewIndex === i" class="xm-df-preview">{{ previewText(s.content) }}</div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

defineOptions({ name: 'DraftBoxDialog' })

export interface DraftSnapshot {
  time: string
  length: number
  content: string
}

const props = defineProps<{
  visible: boolean
  snapshots: DraftSnapshot[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
  (e: 'restore', content: string): void
  (e: 'remove', index: number): void
}>()

const previewIndex = ref(-1)

const previewText = (content: string) => content.slice(0, 400) + (content.length > 400 ? '…' : '')

const restore = (i: number) => {
  emit('restore', props.snapshots[i].content)
  emit('update:visible', false)
}

const remove = (i: number) => emit('remove', i)
</script>

<style scoped>
.xm-df-body {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 420px;
  overflow: auto;
}

.xm-df-hint {
  margin: 0 0 4px;
  font-size: 11.5px;
  color: #a3a8b3;
}

.xm-df-empty {
  font-size: 13px;
  color: #a3a8b3;
  text-align: center;
  padding: 40px 0;
}

.xm-df-row {
  border: 1px solid #eef0f2;
  border-radius: 8px;
  padding: 8px 12px;
}

.xm-df-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.xm-df-time {
  font-size: 12.5px;
  color: #374151;
  font-variant-numeric: tabular-nums;
}

.xm-df-count {
  font-size: 11px;
  color: #a3a8b3;
}

.xm-df-actions {
  display: flex;
  gap: 4px;
  margin-left: auto;
}

.xm-df-preview {
  margin-top: 8px;
  font-size: 12px;
  line-height: 1.7;
  color: #6b7280;
  background: #f7f8fa;
  border-radius: 6px;
  padding: 8px 10px;
  white-space: pre-wrap;
  word-break: break-word;
}

:root[data-theme='dark'] .xm-df-row {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-df-time {
  color: #e6edf3;
}

:root[data-theme='dark'] .xm-df-preview {
  background: rgba(28, 35, 51, 0.6);
  color: #8b98a8;
}

:root[data-theme='dark'] .xm-df-hint,
:root[data-theme='dark'] .xm-df-count,
:root[data-theme='dark'] .xm-df-empty {
  color: #6b7688;
}
</style>
