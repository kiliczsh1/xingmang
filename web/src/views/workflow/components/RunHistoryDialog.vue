<template>
  <el-dialog
    :model-value="modelValue"
    title="执行历史"
    width="860px"
    append-to-body
    class="run-history-dialog"
    @update:model-value="emit('update:modelValue', $event)"
    @open="load"
  >
    <div class="rh-toolbar">
      <span class="rh-count">共 {{ runs.length }} 条记录（最多 50 条）</span>
      <el-button size="small" text :disabled="loading" @click="load">
        <i class="fas fa-rotate-right" /> 刷新
      </el-button>
    </div>

    <div v-if="!runs.length" class="rh-empty">
      <el-empty :description="loading ? '加载中...' : '还没有执行记录'" :image-size="60" />
    </div>

    <div v-else class="rh-body">
      <div class="rh-list">
        <div
          v-for="r in runs"
          :key="r.id"
          class="rh-item"
          :class="[`rh-${r.status}`, { active: active?.id === r.id }]"
          @click="select(r)"
        >
          <span class="rh-status">{{ statusLabel(r.status) }}</span>
          <span class="rh-time">{{ formatTime(r.started_at) }}</span>
          <span class="rh-duration">{{ durationOf(r) }}</span>
          <span class="rh-error" :title="r.error">{{ r.error || '—' }}</span>
          <el-button
            size="small"
            text
            class="rh-del"
            title="删除该记录"
            @click.stop="remove(r)"
          >
            <i class="fas fa-trash" />
          </el-button>
        </div>
      </div>

      <div v-if="active" class="rh-detail">
        <div class="rh-detail-head">
          <span>#{{ active.id }} 执行日志</span>
          <span class="rh-detail-sub">
            输出目录：{{ active.output_dir || '（未落盘）' }}
          </span>
        </div>
        <el-scrollbar class="rh-logs">
          <div v-if="!activeLogs.length" class="rh-logs-empty">该次执行没有日志</div>
          <div v-for="(l, i) in activeLogs" :key="i" class="rh-log-item" :class="`log-${l.type || l.level || 'info'}`">
            <span class="rh-log-time">{{ l.timestamp || l.time }}</span>
            <span class="rh-log-msg">{{ l.message }}</span>
          </div>
        </el-scrollbar>
      </div>
    </div>

    <template #footer>
      <el-button size="small" @click="emit('update:modelValue', false)">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import workflowAPI from '@/services/workflow/workflowAPI'
import type { LogEntry, WorkflowRun } from '@/types/workflow'

const props = defineProps<{ modelValue: boolean; workflowId?: number | null }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void }>()

const runs = ref<WorkflowRun[]>([])
const loading = ref(false)
const activeId = ref<number | null>(null)

const active = computed(() => runs.value.find(r => r.id === activeId.value) || null)

/** 后端把日志以 JSON 数组存进 runs.logs */
const activeLogs = computed<LogEntry[]>(() => {
  const raw = active.value?.logs
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
})

const load = async () => {
  loading.value = true
  try {
    const res: any = await workflowAPI.listRuns(props.workflowId ?? undefined, 50)
    runs.value = (res?.data || []) as WorkflowRun[]
    if (activeId.value && !runs.value.some(r => r.id === activeId.value)) activeId.value = null
  } catch {
    runs.value = []
  } finally {
    loading.value = false
  }
}

const select = (r: WorkflowRun) => {
  activeId.value = activeId.value === r.id ? null : r.id
}

const statusLabel = (s: string) =>
  (({ success: '✅ 成功', error: '❌ 失败', running: '⏳ 进行中', stopped: '⏹️ 已停止' }) as Record<string, string>)[
    s
  ] || s

const formatTime = (t?: string) => {
  if (!t) return '—'
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const durationOf = (r: WorkflowRun) => {
  if (!r.started_at || !r.finished_at) return '—'
  const ms = new Date(r.finished_at).getTime() - new Date(r.started_at).getTime()
  if (!Number.isFinite(ms) || ms < 0) return '—'
  if (ms < 1000) return `${ms}ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(1)}s`
  const m = Math.floor(s / 60)
  return `${m}m${Math.round(s % 60)}s`
}

const remove = async (r: WorkflowRun) => {
  try {
    await ElMessageBox.confirm(`确定删除执行记录 #${r.id} 吗？`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  try {
    await workflowAPI.removeRun(r.id)
    runs.value = runs.value.filter(x => x.id !== r.id)
    if (activeId.value === r.id) activeId.value = null
    ElMessage.success('已删除')
  } catch {
    /* 拦截器已提示 */
  }
}
</script>

<style scoped>
.rh-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}
.rh-count {
  font-size: 12px;
  color: var(--el-text-color-secondary, #909399);
}
.rh-empty {
  padding: 20px 0;
}
.rh-body {
  max-height: 60vh;
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.rh-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 220px;
  overflow-y: auto;
}
.rh-item {
  display: grid;
  grid-template-columns: 88px 170px 76px 1fr 32px;
  align-items: center;
  gap: 8px;
  padding: 7px 10px;
  border: 1px solid transparent;
  border-radius: 6px;
  background: var(--el-fill-color-lighter, #fafafa);
  font-size: 12px;
  cursor: pointer;
}
.rh-item:hover {
  background: var(--el-fill-color-light, #f5f7fa);
}
.rh-item.active {
  border-color: var(--el-color-primary, #409eff);
}
.rh-item.rh-error {
  border-left: 3px solid var(--el-color-danger, #f56c6c);
}
.rh-item.rh-success {
  border-left: 3px solid var(--el-color-success, #67c23a);
}
.rh-item.rh-stopped {
  border-left: 3px solid var(--el-color-warning, #e6a23c);
}
.rh-item.rh-running {
  border-left: 3px solid var(--el-color-primary, #409eff);
}
.rh-time {
  color: var(--el-text-color-regular, #606266);
  font-variant-numeric: tabular-nums;
}
.rh-duration {
  color: var(--el-text-color-secondary, #909399);
  font-variant-numeric: tabular-nums;
}
.rh-error {
  color: var(--el-color-danger, #f56c6c);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rh-del {
  color: var(--el-text-color-secondary, #909399);
}
.rh-detail {
  border-top: 1px solid var(--el-border-color-lighter, #ebeef5);
  padding-top: 8px;
}
.rh-detail-head {
  display: flex;
  align-items: baseline;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-primary, #303133);
  margin-bottom: 6px;
}
.rh-detail-sub {
  color: var(--el-text-color-secondary, #909399);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.rh-logs {
  max-height: 240px;
}
.rh-logs-empty {
  font-size: 12px;
  color: var(--el-text-color-placeholder, #c0c4cc);
  padding: 8px 2px;
}
.rh-log-item {
  display: flex;
  gap: 10px;
  padding: 3px 2px;
  font-size: 12px;
  border-left: 2px solid transparent;
  padding-left: 8px;
  margin-bottom: 2px;
}
.rh-log-item .rh-log-time {
  color: var(--el-text-color-secondary, #909399);
  font-family: Consolas, Monaco, monospace;
  font-size: 11px;
  flex-shrink: 0;
}
.rh-log-item .rh-log-msg {
  color: var(--el-text-color-primary, #303133);
  word-break: break-word;
  white-space: pre-wrap;
}
.rh-log-item.log-error {
  border-left-color: var(--el-color-danger, #f56c6c);
}
.rh-log-item.log-warning {
  border-left-color: var(--el-color-warning, #e6a23c);
}
.rh-log-item.log-success {
  border-left-color: var(--el-color-success, #67c23a);
}
.rh-log-item.log-info {
  border-left-color: var(--el-border-color, #dcdfe6);
}
</style>
