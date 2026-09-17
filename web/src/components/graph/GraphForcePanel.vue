<template>
  <section class="graph-force">
    <p class="force-intro">参数相互耦合：按下面的顺序调，一次只动一个。</p>

    <div v-for="item in FORCE_ITEMS" :key="item.key" class="force-row">
      <div class="force-head">
        <span class="force-label">{{ item.label }}</span>
        <span class="force-value">{{ formatParam(item.key) }}</span>
      </div>
      <el-slider
        :model-value="params[item.key]"
        :min="FORCE_PARAM_RANGES[item.key].min"
        :max="FORCE_PARAM_RANGES[item.key].max"
        :step="FORCE_PARAM_RANGES[item.key].step"
        size="small"
        :show-tooltip="false"
        @input="handleInput(item.key, $event)"
      />
      <p class="force-hint">{{ item.hint }}</p>
    </div>

    <el-button size="small" link class="force-reset" @click="emit('reset')">
      恢复默认
    </el-button>
  </section>
</template>

<script setup lang="ts">
import { FORCE_PARAM_RANGES, type ForceParams } from '@/composables/useForceLayout'

const props = defineProps<{
  /** 当前力度参数（受控：状态由页面层持有，画布侧只负责应用） */
  params: ForceParams
}>()

const emit = defineEmits<{
  (e: 'update', patch: Partial<ForceParams>): void
  (e: 'reset'): void
}>()

/**
 * 顺序刻意按「调参顺序」而非参数本身的重要性排列 —— 这四个是耦合的，
 * 先定留白（连线长度）再定其它，比从上往下随手拉更容易调好看。
 * 文案里的坑来自标定与实测，详见 scripts/calibrate_force_params.mjs。
 */
const FORCE_ITEMS: { key: keyof ForceParams; label: string; hint: string }[] = [
  {
    key: 'linkDistance',
    label: '连线长度',
    hint: '留白的主因，先调它。太大时边会横穿整张图'
  },
  {
    key: 'repelForce',
    label: '节点间的排斥力',
    hint: '调高节点更散；过高会把簇结构抹平'
  },
  {
    key: 'linkForce',
    label: '相连节点间的吸引力',
    hint: '默认值 1 已是滑块上限，调低会松开紧邻的节点'
  },
  {
    key: 'centerForce',
    label: '图谱向心力',
    hint: '只用于兜底，防止孤立节点飞走；过高会压成圆饼'
  }
]

/** 整数显示整数、小数保留两位，读数习惯与 Obsidian 一致 */
function formatParam(key: keyof ForceParams): string {
  const value = props.params[key]
  return Number.isInteger(value) ? String(value) : value.toFixed(2)
}

/** emit 增量而不是整份对象，父组件合并后转给画布热更新 */
function handleInput(key: keyof ForceParams, value: number | number[]) {
  const next = Array.isArray(value) ? value[0] : value
  if (typeof next !== 'number') return
  emit('update', { [key]: next })
}
</script>

<style scoped>
.graph-force {
  display: flex;
  flex-direction: column;
}

.force-intro {
  margin: 0 0 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.force-row + .force-row {
  margin-top: 10px;
}

.force-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
}

.force-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.force-value {
  font-size: 12px;
  font-variant-numeric: tabular-nums;
  color: var(--el-text-color-secondary);
}

.force-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-placeholder);
}

.force-reset {
  margin-top: 12px;
  align-self: flex-start;
}
</style>
