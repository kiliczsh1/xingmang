<template>
  <div class="tree-level">
    <div
      v-for="folder in currentFolders"
      :key="`folder_${folder.id}`"
      class="tree-node"
    >
      <div
        class="tree-row tree-row-folder"
        :class="{ 'is-open': isExpanded(folder.id), 'is-drop-target': dragOverKey === String(folder.id) }"
        @click="toggle(folder.id)"
        @dragover.prevent="emit('drag-over-folder', String(folder.id))"
        @dragleave="emit('drag-leave-folder', String(folder.id))"
        @drop.prevent="emit('drop-folder', String(folder.id))"
      >
        <svg class="tree-chevron" viewBox="0 0 10 16" width="10" height="16">
          <path d="M2 2 L8 8 L2 14" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <svg class="tree-icon" viewBox="0 0 24 24" width="16" height="16">
          <path d="M2 6a2 2 0 012-2h5l2 2h9a2 2 0 012 2v10a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" />
        </svg>
        <span class="tree-label">{{ folder.title }}</span>
        <span class="tree-actions" @click.stop>
          <button type="button" class="tree-btn" title="新建文件夹" @click="emit('create-folder', folder.id)">
            <svg viewBox="0 0 24 24" width="14" height="14"><path d="M12 5v14M5 12h14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
          </button>
          <button type="button" class="tree-btn" title="新建文件" @click="emit('create-file', folder.id)">
            <svg viewBox="0 0 24 24" width="14" height="14"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /><path d="M14 2v6h6" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /></svg>
          </button>
          <button type="button" class="tree-btn" title="重命名" @click="emit('edit-folder', folder)">
            <svg viewBox="0 0 24 24" width="14" height="14"><path d="M17 3a2.828 2.828 0 114 4L7.5 20.5 2 22l1.5-5.5L17 3z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" /></svg>
          </button>
          <button type="button" class="tree-btn tree-btn-danger" title="删除" @click="emit('delete-folder', folder)">
            <svg viewBox="0 0 24 24" width="14" height="14"><path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" /></svg>
          </button>
        </span>
      </div>

      <div v-show="isExpanded(folder.id)" class="tree-children">
        <WriteCatalogTree
          v-if="getChildFolders(folder.id).length > 0 || getChildFiles(folder.id).length > 0"
          :folders="folders"
          :files="files"
          :parent-id="folder.id"
          :expanded-ids="expandedIds"
          :current-file-id="currentFileId"
          :drag-over-key="dragOverKey"
          @toggle-folder="emit('toggle-folder', $event)"
          @select-file="emit('select-file', $event)"
          @create-folder="emit('create-folder', $event)"
          @create-file="emit('create-file', $event)"
          @edit-folder="emit('edit-folder', $event)"
          @delete-folder="emit('delete-folder', $event)"
          @move-file="emit('move-file', $event)"
          @delete-file="emit('delete-file', $event)"
          @drag-file-start="emit('drag-file-start', $event)"
          @drag-file-end="emit('drag-file-end')"
          @drag-over-folder="emit('drag-over-folder', $event)"
          @drag-leave-folder="emit('drag-leave-folder', $event)"
          @drop-folder="emit('drop-folder', $event)"
        />
        <div v-else class="tree-empty">空文件夹</div>
      </div>
    </div>

    <div
      v-for="file in currentFiles"
      :key="`file_${file.id}`"
      class="tree-node"
    >
      <!-- 未选中：简洁行 -->
      <div
        v-if="currentFileId !== file.id"
        class="chapter-row"
        draggable="true"
        @dragstart="emit('drag-file-start', file.id)"
        @dragend="emit('drag-file-end')"
        @click="emit('select-file', file)"
      >
        <span class="chapter-row-title">{{ file.title }}</span>
        <span class="chapter-row-word-count">{{ computeWordCount(file) }}字</span>
      </div>

      <!-- 选中：复用同一皮肤，仅加重指示 -->
      <div
        v-else
        class="chapter-row is-active"
        draggable="true"
        @dragstart="emit('drag-file-start', file.id)"
        @dragend="emit('drag-file-end')"
        @click="emit('select-file', file)"
      >
        <div class="chapter-row-indicator"></div>
        <div class="chapter-row-main">
          <div class="chapter-row-line">
            <span class="chapter-row-title">{{ file.title }}</span>
            <div class="chapter-row-meta">
              <span class="chapter-row-word-count">{{ computeWordCount(file) }}字</span>
              <svg class="chapter-row-bookmark" viewBox="0 0 24 24" width="13" height="13">
                <path d="M5 3v18l7-4 7 4V3a2 2 0 00-2-2H7a2 2 0 00-2 2z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/>
              </svg>
            </div>
          </div>
          <div class="chapter-row-sub">
            创建于{{ formatTime(file.created_at || file.updated_at) }}
          </div>

          <!-- 操作按钮行（与原皮肤一致：仅删除） -->
          <div class="chapter-row-actions" @click.stop>
            <button type="button" class="chapter-action-btn chapter-action-delete" title="删除" @click="emit('delete-file', file)">删除</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'WriteCatalogTree' })

import { computed } from 'vue'
import type { Chapter, Volume } from '@/types'

const props = defineProps<{
  folders: Volume[]
  files: Chapter[]
  parentId?: number | null
  expandedIds: Array<number | string>
  currentFileId?: number | null
  dragOverKey?: string | null
}>()

const emit = defineEmits<{
  (e: 'toggle-folder', id: number | string): void
  (e: 'select-file', file: Chapter): void
  (e: 'create-folder', parentId: number | null): void
  (e: 'create-file', parentId: number | null): void
  (e: 'edit-folder', folder: Volume): void
  (e: 'delete-folder', folder: Volume): void
  (e: 'move-file', file: Chapter): void
  (e: 'delete-file', file: Chapter): void
  (e: 'drag-file-start', fileId: number): void
  (e: 'drag-file-end'): void
  (e: 'drag-over-folder', key: string): void
  (e: 'drag-leave-folder', key: string): void
  (e: 'drop-folder', key: string): void
}>()

const isExpanded = (id: number | string) => props.expandedIds.includes(id)
const toggle = (id: number | string) => emit('toggle-folder', id)

const formatTime = (val: string | undefined) => {
  if (!val) return ''
  try {
    const d = new Date(val)
    const pad = (n: number) => String(n).padStart(2, '0')
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
  } catch {
    return ''
  }
}

// 与 Write.vue 中 getContentLength 保持一致：去标签/去代码块/去空白后按字符数计
const computeWordCount = (file: Chapter): number => {
  const content = (file as any).content
  if (typeof content === 'string' && content.length > 0) {
    const plain = content
      .replace(/<[^>]*>/g, '')
      .replace(/```[\s\S]*?```/g, '')
      .replace(/`[^`]*`/g, '')
      .replace(/\s+/g, '')
    return [...plain].length
  }
  return file.word_count || 0
}

const currentFolders = computed(() => getChildFolders(props.parentId ?? null))
const currentFiles = computed(() => getChildFiles(props.parentId ?? null))

const getChildFolders = (parentId: number | null) => {
  return props.folders
    .filter(folder => Number(folder.parent_id || 0) === Number(parentId || 0))
    .sort((a, b) => a.order_num - b.order_num || a.id - b.id)
}

const getChildFiles = (parentId: number | null) => {
  return props.files
    .filter(file => Number(file.volume_id || 0) === Number(parentId || 0))
    .sort((a, b) => a.order_num - b.order_num || a.id - b.id)
}
</script>

<style scoped>
/* ── Things Theme · Light ── */
.tree-level {
  display: flex;
  flex-direction: column;
}

.tree-children {
  padding-left: 16px;
}

.tree-row {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 30px;
  padding: 0 8px;
  margin: 0;
  border-radius: 5px;
  cursor: pointer;
  user-select: none;
  color: var(--t-fg, #555);
  transition: background 0.08s ease, color 0.08s ease;
}

.tree-row:hover {
  background: var(--t-hover, #e2e5e9);
  color: var(--t-fg-hover, #111);
}

.tree-row-folder.is-drop-target {
  background: var(--t-drop, rgba(8, 109, 221, 0.08));
  box-shadow: inset 0 0 0 1.5px var(--t-drop-border, rgba(8, 109, 221, 0.25));
}

/* ── 章节行（统一皮肤：未选中 / 选中复用同套样式）── */
.chapter-row {
  position: relative;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 11px 12px;
  margin: 2px 0;
  border-radius: 6px;
  cursor: pointer;
  user-select: none;
  background: transparent;
  color: #1f2329;
  transition: background 0.12s ease;
}

.chapter-row:hover {
  background: #f3f5f7;
}

.chapter-row-title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  font-weight: 500;
  color: #1f2329;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  letter-spacing: 0.1px;
}

.chapter-row-word-count {
  flex-shrink: 0;
  font-size: 11.5px;
  color: #a3a8b3;
  font-variant-numeric: tabular-nums;
}

/* 选中态：左侧渐变指示条 + 浅色面 + 时间副行 + 收藏图标 */
.chapter-row.is-active {
  flex-direction: row;
  align-items: stretch;
  padding: 0;
  background: linear-gradient(180deg, #eef4ff 0%, #e6efff 100%);
  box-shadow: inset 0 0 0 1px rgba(99, 124, 255, 0.18);
  border-radius: 7px;
  overflow: hidden;
}

.chapter-row.is-active:hover {
  background: linear-gradient(180deg, #e6efff 0%, #dfe9ff 100%);
}

.chapter-row-indicator {
  width: 3.5px;
  flex-shrink: 0;
  background: linear-gradient(180deg, #6366f1 0%, #8b5cf6 50%, #a855f7 100%);
}

.chapter-row-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 11px 12px 11px 14px;
  min-width: 0;
}

.chapter-row-line {
  display: flex;
  align-items: center;
  gap: 8px;
}

.chapter-row-line .chapter-row-title {
  font-weight: 600;
  color: #1a1f36;
}

.chapter-row-meta {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
}

.chapter-row-bookmark {
  color: #b8c0cc;
  transition: color 0.15s ease;
}

.chapter-row-bookmark:hover {
  color: #6366f1;
}

.chapter-row-sub {
  font-size: 11px;
  color: #94a0b3;
  letter-spacing: 0.1px;
}

/* 操作按钮行 */
.chapter-row-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
}

.chapter-action-btn {
  font-size: 11.5px;
  padding: 3px 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  transition: all 0.12s ease;
  font-weight: 500;
}

.chapter-action-delete {
  background: #fef2f2;
  color: #dc2626;
}

.chapter-action-delete:hover {
  background: #fee2e2;
}

.tree-chevron {
  flex-shrink: 0;
  width: 10px;
  opacity: 0.4;
  transition: transform 0.15s ease;
  margin-left: -2px;
  color: var(--t-muted, #999);
}

.tree-row-folder.is-open > .tree-chevron {
  transform: rotate(90deg);
}

.tree-icon {
  flex-shrink: 0;
  opacity: 0.4;
  color: var(--t-muted, #999);
}

.tree-row-folder.is-open > .tree-icon {
  opacity: 0.55;
}

.tree-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  line-height: 1;
}

.tree-actions {
  display: flex;
  gap: 0;
  opacity: 0;
  flex-shrink: 0;
  transition: opacity 0.1s ease;
}

.tree-row:hover .tree-actions,
.tree-row-file.active .tree-actions {
  opacity: 1;
}

.tree-btn {
  width: 24px;
  height: 24px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--t-muted, #999);
  opacity: 0.65;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: opacity 0.1s, background 0.1s, color 0.1s;
}

.tree-btn:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.06);
  color: var(--t-btn-hover-fg, #333);
}

.tree-btn-danger:hover {
  background: rgba(232, 62, 62, 0.1);
  color: #e83e3e;
}

.tree-empty {
  padding: 4px 10px;
  font-size: 12px;
  color: var(--t-muted, #999);
  font-style: italic;
}
</style>

<style>
/* ── 暗色模式（非 scoped，确保 ::root 选择器正常生效） ── */

/* 章节行 */
:root[data-theme='dark'] .chapter-row {
  color: #e2e8f0 !important;
  background: rgba(30, 41, 59, 0.4) !important;
}

:root[data-theme='dark'] .chapter-row:hover {
  background: rgba(51, 65, 85, 0.6) !important;
}

:root[data-theme='dark'] .chapter-row-title {
  color: #f1f5f9 !important;
}

:root[data-theme='dark'] .chapter-row-word-count {
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .chapter-row.is-active {
  background: linear-gradient(180deg, rgba(15, 118, 110, 0.25) 0%, rgba(20, 100, 90, 0.18) 100%) !important;
  box-shadow: inset 0 0 0 1px rgba(94, 234, 212, 0.28) !important;
}

:root[data-theme='dark'] .chapter-row.is-active:hover {
  background: linear-gradient(180deg, rgba(15, 118, 110, 0.32) 0%, rgba(20, 100, 90, 0.24) 100%) !important;
}

:root[data-theme='dark'] .chapter-row-indicator {
  background: linear-gradient(180deg, #5eead4 0%, #2dd4bf 50%, #14b8a6 100%) !important;
}

:root[data-theme='dark'] .chapter-row-line .chapter-row-title {
  color: #f8fafc !important;
}

:root[data-theme='dark'] .chapter-row-sub {
  color: #94a3b8 !important;
}

:root[data-theme='dark'] .chapter-row-bookmark {
  color: #64748b !important;
}

:root[data-theme='dark'] .chapter-row-bookmark:hover {
  color: #5eead4 !important;
}

/* 删除按钮 */
:root[data-theme='dark'] .chapter-action-delete {
  background: rgba(220, 38, 38, 0.12) !important;
  color: #f87171 !important;
}

:root[data-theme='dark'] .chapter-action-delete:hover {
  background: rgba(220, 38, 38, 0.2) !important;
}

/* 文件夹/树行 */
:root[data-theme='dark'] .tree-row {
  color: #cbd5e1 !important;
}

:root[data-theme='dark'] .tree-row:hover {
  background: rgba(51, 65, 85, 0.5) !important;
  color: #f1f5f9 !important;
}

:root[data-theme='dark'] .tree-label {
  color: #e2e8f0 !important;
}

:root[data-theme='dark'] .tree-chevron,
:root[data-theme='dark'] .tree-icon {
  color: #64748b !important;
}

:root[data-theme='dark'] .tree-btn:hover {
  background: rgba(94, 234, 212, 0.12) !important;
  color: #5eead4 !important;
}

:root[data-theme='dark'] .tree-btn-danger:hover {
  background: rgba(248, 113, 113, 0.15) !important;
  color: #f87171 !important;
}

:root[data-theme='dark'] .tree-empty {
  color: #475569 !important;
}
</style>
