<template>
  <!-- 马良同款：左下角毛玻璃图例卡片，彩色圆点 + 类型名 + 数量，点击筛选 -->
  <div class="graph-legend">
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
</template>

<script setup lang="ts">
import { computed } from 'vue'
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
  display: grid;
  grid-template-columns: repeat(3, auto);
  gap: 4px;
  align-items: center;
  max-width: calc(100% - 24px);
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid rgba(128, 138, 155, 0.18);
  background: rgba(255, 255, 255, 0.88);
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
  background: rgba(64, 158, 255, 0.14);
  color: #409eff;
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
  background: rgba(64, 158, 255, 0.12);
  color: #409eff;
  font-size: 11px;
  cursor: pointer;
}

.legend-clear:hover {
  background: rgba(64, 158, 255, 0.2);
}

:root[data-theme='dark'] .graph-legend {
  background: rgba(22, 27, 37, 0.88);
  border-color: rgba(75, 85, 105, 0.35);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
}

:root[data-theme='dark'] .legend-item {
  color: #d1d9e3;
}

:root[data-theme='dark'] .legend-item:hover {
  background: rgba(255, 255, 255, 0.08);
}

:root[data-theme='dark'] .legend-item.active {
  background: rgba(64, 158, 255, 0.22);
  color: #79aefc;
}
</style>
