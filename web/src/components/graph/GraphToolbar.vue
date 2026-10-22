<template>
  <div class="graph-toolbar">
    <div class="toolbar-row">
      <div class="toolbar-left">
        <span class="graph-title">知识图谱</span>
        <span class="graph-stats" v-if="entityCount > 0">
          {{ entityCount }} 实体 · {{ relationCount }} 关系
        </span>
      </div>

      <div class="toolbar-right">
        <!--
          常驻可点元素按 DESIGN_GUIDELINES B1 收到 4 个：
          搜索框 · 显示▾ · 编辑▾ · AI 分析（唯一主操作）
          其余全部收进下拉，不再平铺。
        -->
        <el-input
          v-model="keyword"
          class="search-input"
          size="small"
          clearable
          placeholder="搜索实体名"
          @keyup.enter="emit('search-next')"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <span v-if="keyword" class="search-count">
          {{ matchCount > 0 ? `${matchCursor + 1}/${matchCount}` : '无匹配' }}
        </span>

        <!-- 显示：视图与可见性 -->
        <el-dropdown trigger="click" @command="onDisplayCommand">
          <el-button size="small" plain>
            <el-icon><View /></el-icon>
            显示
            <el-icon class="caret"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="fit-view">
                <el-icon><FullScreen /></el-icon>适配视图
              </el-dropdown-item>
              <el-dropdown-item command="reset-view">
                <el-icon><RefreshLeft /></el-icon>重置布局
              </el-dropdown-item>
              <el-dropdown-item command="zoom-in">
                <el-icon><ZoomIn /></el-icon>放大
              </el-dropdown-item>
              <el-dropdown-item command="zoom-out">
                <el-icon><ZoomOut /></el-icon>缩小
              </el-dropdown-item>
              <el-dropdown-item divided command="toggle-edge-labels">
                <el-icon><PriceTag /></el-icon>
                {{ showEdgeLabels ? '隐藏' : '显示' }}关系标签
              </el-dropdown-item>
              <el-dropdown-item command="toggle-arrows">
                <el-icon><ArrowRight /></el-icon>
                {{ showArrows ? '隐藏' : '显示' }}方向箭头
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <!-- 编辑：改图与参数 -->
        <el-dropdown trigger="click" @command="onEditCommand">
          <el-button size="small" plain>
            <el-icon><EditPen /></el-icon>
            编辑
            <el-icon class="caret"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="toggle-link-mode">
                <el-icon><Share /></el-icon>
                {{ linkMode ? '退出连线模式' : '人工连线' }}
              </el-dropdown-item>
              <el-dropdown-item command="open-force">
                <el-icon><Odometer /></el-icon>力度设置
              </el-dropdown-item>
              <el-dropdown-item divided command="refresh">
                <el-icon><Refresh /></el-icon>刷新数据
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>

        <!-- 唯一主操作 -->
        <el-button size="small" type="primary" plain @click="emit('analyze')">
          <el-icon><MagicStick /></el-icon>
          AI 分析
        </el-button>

        <!-- 连线模式提示（非可点元素，仅是状态提示） -->
        <span v-if="linkMode" class="link-badge">
          <el-icon><Share /></el-icon>
          连线中
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  ArrowDown,
  ArrowRight,
  EditPen,
  FullScreen,
  MagicStick,
  Odometer,
  PriceTag,
  Refresh,
  RefreshLeft,
  Search,
  Share,
  View,
  ZoomIn,
  ZoomOut
} from '@element-plus/icons-vue'

const props = withDefaults(
  defineProps<{
    entityCount: number
    relationCount: number
    showEdgeLabels: boolean
    loading: boolean
    searchKeyword: string
    matchCount: number
    matchCursor: number
    /** 人工连线模式是否开启 */
    linkMode?: boolean
    /** 是否显示方向箭头 */
    showArrows: boolean
  }>(),
  { matchCursor: 0, linkMode: false }
)

const emit = defineEmits<{
  (e: 'update:searchKeyword', keyword: string): void
  (e: 'toggle-edge-labels'): void
  (e: 'toggle-arrows'): void
  (e: 'open-force'): void
  (e: 'toggle-link-mode'): void
  (e: 'zoom-in'): void
  (e: 'zoom-out'): void
  (e: 'fit-view'): void
  (e: 'reset-view'): void
  (e: 'refresh'): void
  (e: 'search-next'): void
  (e: 'analyze'): void
}>()

const keyword = computed({
  get: () => props.searchKeyword,
  set: value => emit('update:searchKeyword', value)
})

/**
 * 下拉菜单的 command 名称与 emit 事件名一一对应，
 * 这样加菜单项时不用再写一遍 switch 分支。
 */
const DISPLAY_EVENTS = [
  'fit-view',
  'reset-view',
  'zoom-in',
  'zoom-out',
  'toggle-edge-labels',
  'toggle-arrows'
] as const

const EDIT_EVENTS = ['toggle-link-mode', 'open-force', 'refresh'] as const

function dispatch(command: string, allowed: readonly string[]) {
  if (!allowed.includes(command)) return
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  emit(command as any)
}

function onDisplayCommand(command: string) {
  dispatch(command, DISPLAY_EVENTS)
}

function onEditCommand(command: string) {
  dispatch(command, EDIT_EVENTS)
}
</script>

<style scoped>
.graph-toolbar {
  padding: 10px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.toolbar-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.toolbar-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
  flex-shrink: 0;
  min-width: 0;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  margin-left: auto;
}

.graph-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  white-space: nowrap;
}

.graph-stats {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.search-input {
  width: 160px;
}

.search-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
  min-width: 34px;
}

.caret {
  margin-left: 2px;
  font-size: 11px;
  opacity: 0.7;
}

/* 连线模式的状态提示：不可点，仅告知当前处于特殊模式 */
.link-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 12px;
  color: var(--el-color-warning);
  background: var(--el-color-warning-light-9);
  border: 1px solid var(--el-color-warning-light-7);
  white-space: nowrap;
}

:root[data-theme='dark'] .graph-toolbar {
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .graph-title {
  color: #e6edf3;
}

:root[data-theme='dark'] .graph-stats,
:root[data-theme='dark'] .search-count {
  color: #9ca3af;
}

:root[data-theme='dark'] .link-badge {
  background: rgba(245, 158, 11, 0.12);
  border-color: rgba(245, 158, 11, 0.35);
}
</style>
