<template>
  <div class="graph-version-list">
    <div class="version-toolbar">
      <el-input
        v-model="searchText"
        placeholder="搜索版本名"
        size="small"
        clearable
        class="version-search"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-button
        type="primary"
        size="small"
        :disabled="!bookId"
        @click="$emit('create-version')"
      >
        <el-icon><Plus /></el-icon>
        新建
      </el-button>
    </div>

    <div class="version-meta" v-if="filteredVersions.length > 0">
      共 {{ versions.length }} 个版本
    </div>
    <div class="version-meta" v-else-if="bookId">暂无图谱版本</div>
    <div class="version-meta" v-else>请先选择书籍</div>

    <div class="version-scroll">
      <div
        v-for="ver in filteredVersions"
        :key="ver.id"
        class="version-card"
        :class="{ active: ver.id === currentVersionId, switching: switchingId === ver.id }"
        @click="$emit('switch-version', ver.id)"
      >
        <div class="version-card-main">
          <div class="version-card-title">
            <el-icon class="version-icon"><Share /></el-icon>
            <span class="version-name" :title="ver.name">{{ ver.name }}</span>
            <el-tag v-if="ver.id === currentVersionId" size="small" type="success" effect="plain" class="active-tag">当前</el-tag>
          </div>
          <div class="version-card-stats">
            <span class="stat">
              <el-icon><User /></el-icon>
              {{ ver.entity_count }}
            </span>
            <span class="stat-sep">·</span>
            <span class="stat">
              <el-icon><Link /></el-icon>
              {{ ver.relation_count }}
            </span>
            <span class="stat-sep">·</span>
            <span class="stat-time">{{ formatRelative(ver.created_at) }}</span>
          </div>
        </div>
        <div class="version-card-actions" @click.stop>
          <el-tooltip content="重命名" placement="top">
            <el-button size="small" link @click="$emit('rename-version', ver.id)">
              <el-icon><Edit /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="导出" placement="top">
            <el-button size="small" link @click="$emit('export-version', ver.id)">
              <el-icon><Download /></el-icon>
            </el-button>
          </el-tooltip>
          <el-tooltip content="删除" placement="top">
            <el-button size="small" link type="danger" @click="$emit('delete-version', ver.id)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-tooltip>
        </div>
        <div v-if="switchingId === ver.id" class="version-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import type { GraphVersion } from '@/types'
import {
  Search, Plus, Share, User, Link, Edit, Download, Delete, Loading
} from '@element-plus/icons-vue'

const props = defineProps<{
  versions: GraphVersion[]
  currentVersionId: number | null
  bookId: number | null
  switchingId?: number | null
}>()

defineEmits<{
  (e: 'switch-version', id: number): void
  (e: 'rename-version', id: number): void
  (e: 'delete-version', id: number): void
  (e: 'export-version', id: number): void
  (e: 'create-version'): void
}>()

const searchText = ref('')

const filteredVersions = computed(() => {
  const kw = searchText.value.trim().toLowerCase()
  if (!kw) return props.versions
  return props.versions.filter(v => v.name.toLowerCase().includes(kw))
})

function formatRelative(iso: string): string {
  if (!iso) return ''
  const t = new Date(iso).getTime()
  if (isNaN(t)) return iso
  const diff = Date.now() - t
  const min = 60 * 1000
  const hour = 60 * min
  const day = 24 * hour
  if (diff < min) return '刚刚'
  if (diff < hour) return `${Math.floor(diff / min)} 分钟前`
  if (diff < day) return `${Math.floor(diff / hour)} 小时前`
  if (diff < 7 * day) return `${Math.floor(diff / day)} 天前`
  const d = new Date(t)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}
</script>

<style scoped>
.graph-version-list {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.version-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 4px 12px;
}

.version-search {
  flex: 1;
}

.version-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  padding: 0 4px 8px;
}

.version-scroll {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding-right: 2px;
}

.version-card {
  position: relative;
  padding: 10px 12px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.55);
  border: 1px solid rgba(45, 212, 191, 0.15);
  cursor: pointer;
  transition: all 0.18s ease;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.version-card:hover {
  background: rgba(45, 212, 191, 0.08);
  border-color: rgba(45, 212, 191, 0.35);
  transform: translateY(-1px);
}

.version-card.active {
  background: rgba(45, 212, 191, 0.12);
  border-color: rgba(45, 212, 191, 0.55);
  box-shadow: 0 2px 8px rgba(45, 212, 191, 0.18);
}

.version-card.switching {
  opacity: 0.7;
  pointer-events: none;
}

.version-card-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.version-card-title {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.version-icon {
  color: #2dd4bf;
  flex-shrink: 0;
  font-size: 14px;
}

.version-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}

.active-tag {
  flex-shrink: 0;
  height: 18px;
  line-height: 17px;
  padding: 0 6px;
  font-size: 10px;
}

.version-card-stats {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  padding-left: 20px;
}

.version-card-stats .stat {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.version-card-stats .stat .el-icon {
  font-size: 11px;
}

.stat-sep {
  opacity: 0.5;
}

.stat-time {
  margin-left: 2px;
}

.version-card-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 2px;
  opacity: 0;
  transition: opacity 0.15s;
}

.version-card:hover .version-card-actions,
.version-card.active .version-card-actions {
  opacity: 1;
}

.version-card-actions .el-button {
  padding: 2px 4px;
  font-size: 12px;
}

.version-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 8px;
}

.version-loading .el-icon {
  font-size: 18px;
  color: #2dd4bf;
}
</style>
