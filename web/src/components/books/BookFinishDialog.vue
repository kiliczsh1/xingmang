<template>
  <el-dialog
    v-model="finishDialogVisible"
    width="400px"
    :align-center="true"
    :close-on-click-modal="true"
    append-to-body
    class="book-finish-dialog"
    destroy-on-close
  >
    <template #header>
      <div class="finish-head">
        <span class="finish-title">确认完结</span>
        <span class="finish-sub">FINAL STATUS</span>
      </div>
    </template>

    <div v-if="finishBook" class="finish-body">
      <div class="finish-book">
        <span class="finish-book-name">{{ finishBook.title }}</span>
        <span class="finish-book-now">当前 · {{ getStatusText(finishBook.status) }}</span>
      </div>

      <p class="finish-hint">选择该作品的最终状态，确认后立即生效。</p>

      <div class="finish-options">
        <button
          v-for="opt in FINISH_STATUS_OPTIONS"
          :key="opt.value"
          type="button"
          class="finish-option"
          :class="[opt.value, { active: finishStatus === opt.value }]"
          @click="finishStatus = opt.value"
        >
          <span class="finish-dot"></span>
          <span class="finish-option-main">
            <span class="finish-option-label">{{ opt.label }}</span>
            <span class="finish-option-desc">{{ opt.desc }}</span>
          </span>
          <el-icon class="finish-option-check"><Check /></el-icon>
        </button>
      </div>

      <p v-if="finishBook.status === 'completed'" class="finish-note">
        该作品已是完结状态，可在此改回连载或草稿。
      </p>
    </div>

    <template #footer>
      <div class="finish-footer">
        <el-button @click="finishDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="finishSaving" @click="confirmFinish">
          {{ finishStatus === 'completed' ? '确认完结' : '确认修改' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
// 确认完结 · 状态选择弹窗
// 2026-09-21 第 3 期从 views/Books.vue 拆出；弹窗自持「目标状态 / 保存中」，书与开关由父组件传入
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import { useBookStore } from '@/stores/book'
import type { Book } from '@/types'
import { getStatusText } from '@/utils/bookDisplay'

type BookStatus = 'draft' | 'published' | 'completed'

const props = defineProps<{ visible: boolean; book: Book | null }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>()

const bookStore = useBookStore()

const finishBook = computed(() => props.book)
const finishDialogVisible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v)
})

const FINISH_STATUS_OPTIONS: Array<{ value: BookStatus; label: string; desc: string }> = [
  { value: 'completed', label: '已完结', desc: '全书完结，归入「已完结」筛选' },
  { value: 'published', label: '已发布', desc: '连载中，已上架持续更新' },
  { value: 'draft', label: '创作中', desc: '草稿阶段，尚未完结' }
]

const finishStatus = ref<BookStatus>('completed')

const finishSaving = ref(false)

const confirmFinish = async () => {
  const book = finishBook.value
  if (!book) return
  const target = finishStatus.value
  if ((book.status || 'draft') === target) {
    ElMessage.info('作品状态没有变化')
    finishDialogVisible.value = false
    return
  }
  finishSaving.value = true
  try {
    const updated = await bookStore.updateBook(book.id, { status: target })
    if (updated) {
      const label = FINISH_STATUS_OPTIONS.find(o => o.value === target)?.label || ''
      ElMessage.success(
        target === 'completed'
          ? `《${updated.title}》已确认完结`
          : `《${updated.title}》状态已改为「${label}」`
      )
      finishDialogVisible.value = false
    } else {
      ElMessage.error('保存失败，请重试')
    }
  } catch (error) {
    ElMessage.error('保存失败，请重试')
  } finally {
    finishSaving.value = false
  }
}
// 打开时总是落在「已完结」——按钮本意就是确认完结；已是完结状态时同样回显，便于改回
watch(() => props.visible, (v) => {
  if (v) finishStatus.value = 'completed'
})
</script>

<style scoped>
/* ========== 确认完结 · 状态选择弹窗 ========== */
/* 注意：弹窗 append-to-body，拿不到 .books-container 上的 --c-* 变量，
   因此配色用字面量 + `:root[data-theme='dark']` 覆盖；强调色跟随 EP 主色（皮肤）。 */
.finish-head {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.finish-title {
  font-size: 16px;
  font-weight: 700;
  color: #23283a;
}

.finish-sub {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 1.5px;
  color: #6b7385;
}

.finish-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.finish-book {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #f4f6fb;
  border: 1px solid #e7eaf2;
}

.finish-book-name {
  font-size: 14px;
  font-weight: 600;
  color: #23283a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.finish-book-now {
  flex-shrink: 0;
  font-size: 12px;
  color: #6b7385;
}

.finish-hint {
  margin: 0;
  font-size: 13px;
  color: #6b7385;
}

.finish-options {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.finish-option {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 11px 14px;
  border-radius: 12px;
  border: 1px solid #e7eaf2;
  background: #fff;
  font-family: inherit;
  text-align: left;
  cursor: pointer;
  transition: all .16s ease;
}

.finish-option:hover {
  border-color: #b9c6ff;
  background: #f7f9ff;
}

.finish-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #c3c9d6;
}

.finish-option.completed .finish-dot { background: #f59e0b; }
.finish-option.published .finish-dot { background: #10b981; }
.finish-option.draft .finish-dot { background: #94a3b8; }

.finish-option-main {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
  min-width: 0;
}

.finish-option-label {
  font-size: 14px;
  font-weight: 600;
  color: #23283a;
}

.finish-option-desc {
  font-size: 12px;
  color: #6b7385;
}

.finish-option-check {
  flex-shrink: 0;
  font-size: 15px;
  color: transparent;
  transition: color .16s ease;
}

.finish-option.active {
  border-color: var(--el-color-primary, #6c8cff);
  background: color-mix(in srgb, var(--el-color-primary, #6c8cff) 8%, #fff);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--el-color-primary, #6c8cff) 16%, transparent);
}

.finish-option.active .finish-option-check {
  color: var(--el-color-primary, #6c8cff);
}

.finish-option.active .finish-option-label {
  color: var(--el-color-primary, #6c8cff);
}

.finish-note {
  margin: 0;
  font-size: 12px;
  color: #b45309;
}

.finish-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

/* 确认完结弹窗 · 暗色 */
:root[data-theme='dark'] .finish-title {
  color: #e6edf3;
}

:root[data-theme='dark'] .finish-sub {
  color: #6b7385;
}

:root[data-theme='dark'] .finish-book {
  background: #1b2230;
  border-color: #21262d;
}

:root[data-theme='dark'] .finish-book-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .finish-book-now,
:root[data-theme='dark'] .finish-option-desc,
:root[data-theme='dark'] .finish-hint {
  color: #8b98a8;
}

:root[data-theme='dark'] .finish-option {
  background: #161b22;
  border-color: #21262d;
}

:root[data-theme='dark'] .finish-option:hover {
  border-color: #3b4a6b;
  background: #1b2230;
}

:root[data-theme='dark'] .finish-option-label {
  color: #e6edf3;
}

:root[data-theme='dark'] .finish-option.active {
  background: color-mix(in srgb, var(--el-color-primary, #6c8cff) 18%, #161b22);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--el-color-primary, #6c8cff) 26%, transparent);
}

:root[data-theme='dark'] .finish-note {
  color: #fbbf24;
}
</style>
