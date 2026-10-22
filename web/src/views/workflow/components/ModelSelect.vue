<template>
  <div class="mp">
    <el-select
      :model-value="modelValue"
      filterable
      clearable
      placeholder="选择模型"
      :loading="loading"
      class="mp-select"
      @update:model-value="onSelect"
    >
      <el-option-group v-for="g in groups" :key="g.provider" :label="g.provider">
        <el-option v-for="m in g.items" :key="m.id" :label="m.label" :value="m.id" />
      </el-option-group>
    </el-select>
    <el-button class="mp-refresh" :loading="loading" @click="load(true)">
      <i class="fas fa-rotate" />
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { buildModelSnapshot, loadModels, modelDisplayName } from '@/services/workflow/modelResolve'
import { filterByTypes } from '@/services/modelType'
import type { ApiModel, ApiConfigType } from '@/types'
import type { ModelSnapshot } from '@/types/workflow'

const props = defineProps<{
  modelValue?: number
  /** 只列出指定类型的模型（对话 / 图片 / 视频 / 视觉）；不传则全部列出 */
  configType?: ApiConfigType | ApiConfigType[]
}>()
const emit = defineEmits<{
  (e: 'update:modelValue', value: number | undefined): void
  (e: 'picked', payload: { configId: number; modelId: string; modelConfig: ModelSnapshot }): void
}>()

const loading = ref(false)
const models = ref<ApiModel[]>([])

const groups = computed(() => {
  const map = new Map<string, Array<{ id: number; label: string }>>()
  filterByTypes(models.value.filter(m => m.enabled !== 0), props.configType)
    .forEach(m => {
      const provider = m.provider_name || '未分组'
      if (!map.has(provider)) map.set(provider, [])
      map.get(provider)!.push({ id: m.id, label: modelDisplayName(m) })
    })
  return Array.from(map.entries()).map(([provider, items]) => ({ provider, items }))
})

const load = async (force = false) => {
  loading.value = true
  try {
    models.value = await loadModels(force)
  } catch {
    models.value = []
  } finally {
    loading.value = false
  }
}

const onSelect = (value: number | undefined) => {
  emit('update:modelValue', value ?? undefined)
  if (!value) return
  const hit = models.value.find(m => m.id === value)
  if (!hit) return
  emit('picked', {
    configId: hit.id,
    modelId: `${hit.provider_id}::${hit.model}`,
    modelConfig: buildModelSnapshot(hit)
  })
  ElMessage.success(`已选择 ${modelDisplayName(hit)}`)
}

onMounted(() => load(false))
</script>

<style scoped>
.mp {
  display: flex;
  gap: 6px;
  width: 100%;
}
.mp-select {
  flex: 1;
  min-width: 0;
}
.mp-refresh {
  flex: none;
  padding: 0 8px;
}
</style>
