<template>
  <!--
    左下角图例。默认折叠成一个小入口按钮（B1 预算只占 1 个），
    点开展开浮层显示各类型 chip 与筛选。

    为什么折叠：8 个 chip 常驻在密集图上是很显眼的视觉噪声，
    而"按类型筛选"不是每次看图都做的事。折叠后画布立刻干净，
    这本身就是视觉升级的一部分。
  -->
  <div class="graph-legend" :class="{ 'is-open': open }">
    <button
      class="legend-toggle"
      :class="{ 'has-filter': activeTypes.length > 0 }"
      :title="open ? '收起图例' : '展开图例与类型筛选'"
      @click="toggle"
    >
      <el-icon><Grid /></el-icon>
      <span>图例</span>
      <span v-if="activeTypes.length > 0" class="legend-toggle-count">
        {{ activeTypes.length }}/{{ visibleItems.length }}
      </span>
    </button>

    <div v-if="open" class="legend-panel">
      <button
        v-for="item in visibleItems"
        :key="item.value"
        class="legend-item"
        :class="{ active: activeTypes.includes(item.value), muted: activeTypes.length > 0 && !activeTypes.includes(item.value) }"
        :title="`${item.label} ${item.count} 个，点击筛选`"
        @click="emit('toggle-type', item.value)"
      >
        <span class="legend-dot" :style="{ background: item.color }"></span>
        <span class="legend-text">{{ item.label }}</span>
        <span class="legend-count">{{ item.count }}</span>
      </button>

      <button
        v-if="unknownCount > 0"
        class="legend-item is-static"
        title="不在预置类型中的实体，无法按类型筛选"
      >
        <span class="legend-dot" :style="{ background: FALLBACK_TYPE_COLOR }"></span>
        <span class="legend-text">{{ FALLBACK_TYPE_LABEL }}</span>
        <span class="legend-count">{{ unknownCount }}</span>
      </button>

      <button
        v-if="activeTypes.length > 0"
        class="legend-clear"
        title="清除类型筛选"
        @click="emit('clear-types')"
      >
        清除筛选
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Grid } from '@element-plus/icons-vue'
import type { GraphEntity, GraphEntityType } from '@/types'
import {
  FALLBACK_TYPE_COLOR,
  FALLBACK_TYPE_LABEL,
  GRAPH_TYPE_META
} from '@/utils/graphColor'

const props = defineProps<{
  /** 各类型实体数；只在全量数据上统计，不受筛选影响 */
  typeCounts: Map<string, number>
  activeTypes: GraphEntityType[]
  entities: GraphEntity[]
}>()

const emit = defineEmits<{
  (e: 'toggle-type', type: GraphEntityType): void
  (e: 'clear-types'): void
}>()

/** 折叠态 / 展开态。默认折叠，避免 chip 常驻占满 B1 预算。 */
const open = ref(false)

function toggle() {
  open.value = !open.value
}

const visibleItems = computed(() =>
  GRAPH_TYPE_META.map(meta => ({
    ...meta,
    count: props.typeCounts.get(meta.value) || 0
  })).filter(item => item.count > 0)
)

/** 后端允许出现预置 7 类之外的 type，单独汇总展示 */
const unknownCount = computed(() => {
  const known = new Set(GRAPH_TYPE_META.map(m => m.value as string))
  let count = 0
  for (const [type, n] of props.typeCounts) {
    if (!known.has(type)) count += n
  }
  return count
})
</script>

<style scoped>
.graph-legend {
  position: absolute;
  bottom: 12px;
  left: 12px;
  z-index: 4;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  max-width: calc(100% - 24px);
}

/* 折叠态的入口按钮：毛玻璃小胶囊 */
.legend-toggle {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 10px;
  border-radius: 999px;
  border: 1px solid rgba(128, 138, 155, 0.18);
  background: rgba(255, 255, 255, 0.88);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow: 0 4px 14px rgba(15, 23, 42, 0.08);
  font-size: 12px;
  line-height: 1.4;
  color: #1f2329;
  cursor: pointer;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.legend-toggle:hover {
  background: rgba(255, 255, 255, 0.98);
  border-color: rgba(128, 138, 155, 0.38);
}

.legend-toggle.has-filter {
  border-color: color-mix(in srgb, var(--el-color-primary) 45%, transparent);
  color: var(--el-color-primary);
}

.legend-toggle-count {
  font-size: 11px;
  opacity: 0.75;
}

/* 展开态面板：沿用原来的毛玻璃卡片 */
.legend-panel {
  display: grid;
  grid-template-columns: repeat(3, auto);
  gap: 4px;
  align-items: center;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid rgba(128, 138, 155, 0.18);
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.08);
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  border: none;
  border-radius: 6px;
  background: transparent;
  font-size: 12px;
  line-height: 1.4;
  color: #1f2329;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s ease, opacity 0.15s ease;
}

.legend-item:hover {
  background: rgba(15, 23, 42, 0.06);
}

.legend-item.active {
  background: color-mix(in srgb, var(--el-color-primary) 14%, transparent);
  color: var(--el-color-primary);
}

.legend-item.muted {
  opacity: 0.4;
}

.legend-item.is-static {
  cursor: default;
}

.legend-item.is-static:hover {
  background: transparent;
}

.legend-dot {
  width: 9px;
  height: 9px;
  border-radius: 50%;
  flex-shrink: 0;
}

.legend-count {
  font-size: 11px;
  color: #8a94a6;
}

.legend-clear {
  grid-column: 1 / -1;
  justify-self: center;
  margin-top: 2px;
  padding: 2px 10px;
  border: none;
  border-radius: 6px;
  background: color-mix(in srgb, var(--el-color-primary) 12%, transparent);
  color: var(--el-color-primary);
  font-size: 11px;
  cursor: pointer;
}

.legend-clear:hover {
  background: color-mix(in srgb, var(--el-color-primary) 20%, transparent);
}

:root[data-theme='dark'] .legend-toggle,
:root[data-theme='dark'] .legend-panel {
  background: rgba(22, 27, 37, 0.9);
  border-color: rgba(75, 85, 105, 0.35);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
}

:root[data-theme='dark'] .legend-toggle {
  color: #d1d9e3;
}

:root[data-theme='dark'] .legend-toggle:hover {
  background: rgba(30, 36, 48, 0.98);
  border-color: rgba(110, 122, 145, 0.5);
}

:root[data-theme='dark'] .legend-toggle.has-filter {
  color: var(--el-color-primary-light-3);
}

:root[data-theme='dark'] .legend-item {
  color: #d1d9e3;
}

:root[data-theme='dark'] .legend-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

:root[data-theme='dark'] .legend-item.active {
  background: color-mix(in srgb, var(--el-color-primary) 22%, transparent);
  color: var(--el-color-primary-light-3);
}
</style>
