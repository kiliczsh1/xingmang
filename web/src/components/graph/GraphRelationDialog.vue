<template>
  <el-dialog
    v-model="visible"
    title="新建关系"
    width="460px"
    append-to-body
    class="xm-dialog"
    @closed="onClosed"
  >
    <div v-if="source && target" class="gr-body">
      <div class="gr-endpoints">
        <div class="gr-endpoint">
          <span class="gr-dot" :style="{ background: typeColor(source.type) }"></span>
          <span class="gr-name">{{ source.name }}</span>
          <span class="gr-type">{{ typeLabel(source.type) }}</span>
        </div>
        <div class="gr-arrow">
          <el-button link size="small" :type="reversed ? 'warning' : 'default'" @click="reversed = !reversed">
            <el-icon><Sort /></el-icon>
            {{ reversed ? '已反转' : '反转方向' }}
          </el-button>
        </div>
        <div class="gr-endpoint">
          <span class="gr-dot" :style="{ background: typeColor(target.type) }"></span>
          <span class="gr-name">{{ target.name }}</span>
          <span class="gr-type">{{ typeLabel(target.type) }}</span>
        </div>
      </div>

      <div v-if="duplicate" class="gr-duplicate">
        <el-icon><WarningFilled /></el-icon>
        <span>两实体之间已存在「{{ duplicate.relation_type }}」关系，再建会产生平行边。</span>
      </div>

      <div class="gr-field">
        <label class="gr-label">关系类型</label>
        <el-input ref="typeInputRef" v-model="relationType" size="small" placeholder="如：师徒、位于、敌对" />
        <div v-if="existingTypes.length > 0" class="gr-chips">
          <el-tag
            v-for="t in existingTypes"
            :key="t"
            size="small"
            effect="plain"
            class="gr-chip"
            @click="relationType = t"
          >
            {{ t }}
          </el-tag>
        </div>
      </div>

      <div class="gr-field">
        <label class="gr-label">描述（选填）</label>
        <el-input
          v-model="description"
          size="small"
          type="textarea"
          :rows="2"
          maxlength="200"
          placeholder="补充说明这条关系的背景"
        />
      </div>

      <div class="gr-hint">保存后写入当前版本，图上以绿色边标识人工连线。</div>
    </div>

    <template #footer>
      <div class="gr-footer">
        <el-button size="small" @click="visible = false">取消</el-button>
        <el-button size="small" type="primary" :disabled="!canSubmit" :loading="saving" @click="submit">
          添加关系
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { Sort, WarningFilled } from '@element-plus/icons-vue'
import type { GraphEntity, GraphRelation } from '@/types'
import { typeColor, typeLabel } from '@/utils/graphColor'

const props = defineProps<{
  visible: boolean
  source: GraphEntity | null
  target: GraphEntity | null
  /** 当前图谱已有的关系类型，作为快捷 chips */
  existingTypes: string[]
  /** 两实体之间已存在的关系（查重提示用），无则为 null */
  duplicate: GraphRelation | null
  saving: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'submit', payload: { sourceId: number; targetId: number; relationType: string; description: string }): void
}>()

const visible = computed({
  get: () => props.visible,
  set: value => emit('update:visible', value)
})

const relationType = ref('')
const description = ref('')
/** 方向反转标记：true 表示以 target 为源点提交 */
const reversed = ref(false)
const typeInputRef = ref<{ focus?: () => void }>()

const canSubmit = computed(
  () => !!props.source && !!props.target && relationType.value.trim().length > 0
)

watch(
  () => props.visible,
  open => {
    if (!open) return
    relationType.value = ''
    description.value = ''
    reversed.value = false
    void nextTick(() => typeInputRef.value?.focus?.())
  }
)

function onClosed() {
  relationType.value = ''
  description.value = ''
  reversed.value = false
}

function submit() {
  if (!canSubmit.value || !props.source || !props.target) return
  const from = reversed.value ? props.target : props.source
  const to = reversed.value ? props.source : props.target
  emit('submit', {
    sourceId: from.id,
    targetId: to.id,
    relationType: relationType.value.trim(),
    description: description.value.trim()
  })
}
</script>

<style scoped>
.gr-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.gr-endpoints {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
}

.gr-endpoint {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.gr-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.gr-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gr-type {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.gr-arrow {
  flex-shrink: 0;
}

.gr-duplicate {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-color-warning);
}

.gr-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.gr-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.gr-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.gr-chip {
  cursor: pointer;
}

.gr-hint {
  font-size: 11px;
  color: var(--el-text-color-secondary);
}

.gr-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
