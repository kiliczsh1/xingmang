<template>
  <el-dialog
    :model-value="visible"
    :title="card?.type === 'role' ? '角色资料卡' : '设定说明'"
    width="420px"
    align-center
    append-to-body
    class="xm-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div v-if="card" class="mark-card">
      <template v-if="card.type === 'role'">
        <div class="mark-card-name">{{ card.data?.name || card.name }}</div>
        <div class="mark-card-tags">
          <el-tag size="small" type="info">{{ genderText }}</el-tag>
          <el-tag v-if="card.data?.folder" size="small">{{ card.data.folder }}</el-tag>
        </div>
        <div class="mark-card-block">
          <div class="mark-card-label">性格</div>
          <div class="mark-card-text">{{ card.data?.personality || '—' }}</div>
        </div>
        <div class="mark-card-block">
          <div class="mark-card-label">信息</div>
          <div class="mark-card-text">{{ card.data?.info || '—' }}</div>
        </div>
      </template>
      <template v-else>
        <div class="mark-card-name">{{ card.data?.comment || card.name }}</div>
        <div class="mark-card-tags">
          <el-tag v-if="card.data?.group" size="small">{{ card.data.group }}</el-tag>
          <el-tag v-if="card.data?.constant" size="small" type="success">常驻</el-tag>
        </div>
        <div class="mark-card-block">
          <div class="mark-card-label">触发词</div>
          <div class="mark-card-text">{{ triggerWords || '—' }}</div>
        </div>
        <div class="mark-card-block">
          <div class="mark-card-label">设定说明</div>
          <div class="mark-card-text mark-card-content">{{ card.data?.content || '—' }}</div>
        </div>
      </template>
    </div>
    <div v-else class="mark-card-empty">未找到对应资料</div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Character } from '@/types'
import type { WorldBookEntry } from '@/types/worldbook'

export interface RoleMarkCardData {
  type: 'role'
  name: string
  data?: Character | null
}

export interface TermMarkCardData {
  type: 'term'
  name: string
  data?: WorldBookEntry | null
}

// 可辨识联合：按 type 收窄 data，模板与计算属性中能拿到各自的字段
export type MarkCardData = RoleMarkCardData | TermMarkCardData

const props = defineProps<{
  visible: boolean
  card: MarkCardData | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
}>()

const genderText = computed(() => {
  const g = props.card?.type === 'role' ? props.card.data?.gender : undefined
  if (g === 'male') return '男'
  if (g === 'female') return '女'
  if (g === 'none') return '无'
  return '未知'
})

const triggerWords = computed(() => {
  const card = props.card
  if (card?.type !== 'term') return ''
  const d = card.data
  if (!d) return ''
  const parts = [d.keyword, d.keywords, ...d.key, ...d.keysecondary]
    .filter(Boolean)
  return Array.from(new Set(parts)).join('、')
})
</script>

<style scoped>
.mark-card {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.mark-card-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--xm-ink, #1e1e1c);
}
.mark-card-tags {
  display: flex;
  gap: 6px;
}
.mark-card-block {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.mark-card-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--xm-ink-3, #9d9c96);
  letter-spacing: .05em;
}
.mark-card-text {
  font-size: 13px;
  color: var(--xm-ink-2, #555);
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}
/* 业务语义：暗色正文保留柔和绿调特调，不走令牌翻转 */
:root[data-theme='dark'] .mark-card-text {
  color: #c8d6cd;
}
.mark-card-content {
  max-height: 220px;
  overflow-y: auto;
  background: var(--xm-surface-2, #faf9f6);
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: var(--xm-radius-sm, 10px);
  padding: 10px 12px;
}
.mark-card-empty {
  text-align: center;
  color: var(--xm-ink-3, #9d9c96);
  padding: 30px 0;
  font-size: 13px;
}
</style>
