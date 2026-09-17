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
        <el-button
          size="small"
          :disabled="matchCount === 0"
          @click="emit('search-next')"
        >
          下一个
        </el-button>

        <el-divider direction="vertical" />

        <el-tooltip content="放大" placement="top">
          <el-button size="small" @click="emit('zoom-in')">
            <el-icon><ZoomIn /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip content="缩小" placement="top">
          <el-button size="small" @click="emit('zoom-out')">
            <el-icon><ZoomOut /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip content="适配视图" placement="top">
          <el-button size="small" @click="emit('fit-view')">
            <el-icon><FullScreen /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip content="重置布局与视图" placement="top">
          <el-button size="small" @click="emit('reset-view')">
            <el-icon><RefreshLeft /></el-icon>
          </el-button>
        </el-tooltip>

        <el-tooltip content="显示/隐藏关系标签" placement="top">
          <el-button
            size="small"
            :type="showEdgeLabels ? 'primary' : 'default'"
            plain
            @click="emit('toggle-edge-labels')"
          >
            <el-icon><PriceTag /></el-icon>
          </el-button>
        </el-tooltip>

        <el-tooltip
          content="方向箭头：单向关系平时看不出指向哪边，打开后所有关系线都带箭头"
          placement="top"
        >
          <el-button
            size="small"
            :type="showArrows ? 'primary' : 'default'"
            plain
            @click="emit('toggle-arrows')"
          >
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </el-tooltip>

        <el-tooltip
          content="力度：图谱向心力 / 节点间的排斥力 / 相连节点间的吸引力 / 连线长度"
          placement="top"
        >
          <el-button size="small" plain @click="emit('open-force')">
            <el-icon><Odometer /></el-icon>
            力度
          </el-button>
        </el-tooltip>

        <el-tooltip content="人工连线：在两个节点之间拖出关系" placement="top">
          <el-button
            size="small"
            :type="linkMode ? 'warning' : 'default'"
            plain
            @click="emit('toggle-link-mode')"
          >
            <el-icon><Share /></el-icon>
            连线
          </el-button>
        </el-tooltip>

        <el-divider direction="vertical" />

        <el-tooltip content="用 AI 从章节中抽取实体与关系" placement="top">
          <el-button size="small" type="primary" plain @click="emit('analyze')">
            <el-icon><MagicStick /></el-icon>
            AI 分析
          </el-button>
        </el-tooltip>

        <el-button size="small" :loading="loading" @click="emit('refresh')">
          <el-icon v-if="!loading"><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  ArrowRight,
  FullScreen,
  MagicStick,
  Odometer,
  PriceTag,
  Refresh,
  RefreshLeft,
  Search,
  Share,
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
</style>
