<template>
  <section class="cap-panel cap-param-panel">
    <header class="cap-panel__head">
      <div class="cap-panel__title">
        <h4>参数组</h4>
        <p>图片 / 视频模型的可变生成参数，例如 duration / resolution / ratio</p>
      </div>
      <div class="cap-panel__actions">
        <button type="button" class="cfg-btn sm" @click="addGroup">
          <el-icon><Plus /></el-icon>参数组
        </button>
      </div>
    </header>

    <div v-if="groups.length === 0" class="cap-panel__empty">
      还没有参数组，点击右上角「参数组」添加，例如 duration / resolution / ratio。
    </div>

    <div v-for="(g, gi) in groups" :key="gi" class="cap-param-group">
      <div class="cap-param-group__head">
        <el-input
          v-model="g.name"
          placeholder="参数名，如 duration"
          class="cap-param-group__name"
        />
        <el-select v-model="g.type" class="cap-param-group__type" style="width: 88px">
          <el-option label="数值" value="number" />
          <el-option label="文本" value="text" />
        </el-select>
        <span class="cap-param-group__count">{{ g.options.length }} 个选项</span>
        <button type="button" class="cap-icon-btn" title="添加选项" @click="addOption(gi)">
          <el-icon><Plus /></el-icon>
        </button>
        <button
          type="button"
          class="cap-icon-btn cap-icon-btn--danger"
          title="删除参数组"
          @click="removeGroup(gi)"
        >
          <el-icon><Delete /></el-icon>
        </button>
      </div>

      <div class="cap-param-options">
        <div class="cap-param-option cap-param-option--head">
          <span>显示名称</span>
          <span>值</span>
          <span></span>
        </div>
        <div v-for="(o, oi) in g.options" :key="oi" class="cap-param-option">
          <el-input v-model="o.label" placeholder="如 15s" />
          <el-input
            :model-value="String(o.value)"
            :placeholder="g.type === 'number' ? '如 15' : '如 2K'"
            @update:model-value="onValueInput(g, o, $event)"
          />
          <button type="button" class="cap-icon-btn" title="删除选项" @click="removeOption(gi, oi)">
            <el-icon><Delete /></el-icon>
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Plus, Delete } from '@element-plus/icons-vue'
import type { ParamGroup } from '@/data/providerPresets'

defineOptions({ name: 'ParamGroupEditor' })

const props = defineProps<{
  /** 参数组列表（v-model） */
  modelValue: ParamGroup[]
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: ParamGroup[]): void
}>()

/**
 * 直接引用父层数组并原地修改，最后 emit 同一引用。
 * 这样输入框不会因为数组被整体替换而丢焦点。
 */
const groups = computed(() => props.modelValue)

const commit = () => emit('update:modelValue', props.modelValue)

const addGroup = () => {
  props.modelValue.push({ name: '', type: 'text', options: [] })
  commit()
}

const removeGroup = (gi: number) => {
  props.modelValue.splice(gi, 1)
  commit()
}

const addOption = (gi: number) => {
  props.modelValue[gi].options.push({ label: '', value: '' })
  commit()
}

const removeOption = (gi: number, oi: number) => {
  props.modelValue[gi].options.splice(oi, 1)
  commit()
}

/** 数值型参数组把输入转成数字，文本型保留原样 */
const onValueInput = (g: ParamGroup, o: { label: string; value: string | number }, raw: string) => {
  if (g.type === 'number') {
    const n = Number(raw)
    o.value = raw.trim() === '' || Number.isNaN(n) ? raw : n
  } else {
    o.value = raw
  }
  commit()
}
</script>

<style scoped>
/* 蓝本 CSS 里没有这两个类（只作钩子用），这里补上布局 */
.cap-param-group__type {
  flex: none;
}

.cap-param-panel .cap-panel__head {
  align-items: flex-start;
}
</style>
