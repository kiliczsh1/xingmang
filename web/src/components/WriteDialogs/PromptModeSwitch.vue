<template>
  <!--
    「快捷选项 / 自定义 / 更多」三段式切换（需求模式条）：
    点「更多」不切换当前模式，只触发 more 事件（打开更多提示词弹窗），
    所以选中态始终停在「快捷选项」，不会出现「切过去了但下面没内容」。
  -->
  <el-radio-group
    :model-value="modelValue"
    size="small"
    class="pm-mode-switch"
    @update:model-value="onChange"
  >
    <el-radio-button value="preset">快捷选项</el-radio-button>
    <el-radio-button v-if="!hideCustom" value="custom">自定义</el-radio-button>
    <el-radio-button value="more">{{ moreLabel }}</el-radio-button>
  </el-radio-group>
</template>

<script setup lang="ts">
defineOptions({ name: 'PromptModeSwitch' })

const props = withDefaults(
  defineProps<{
    modelValue: 'preset' | 'custom'
    /** 「更多」按钮文案 */
    moreLabel?: string
    /** 隐藏「自定义」（少数只允许选预设的场景） */
    hideCustom?: boolean
  }>(),
  { moreLabel: '更多', hideCustom: false }
)

const emit = defineEmits<{
  (e: 'update:modelValue', v: 'preset' | 'custom'): void
  (e: 'more'): void
}>()

function onChange(v: string | number | boolean | undefined) {
  if (v === 'more') {
    emit('more')
    // 不改变 modelValue：保持「快捷选项」选中，避免出现空的第三种模式
    return
  }
  if (v === 'preset' || v === 'custom') emit('update:modelValue', v)
}
</script>

<style scoped>
.pm-mode-switch :deep(.el-radio-button__inner) {
  padding: 6px 14px;
  font-size: 12px;
  border-color: var(--xm-line);
  background: var(--xm-surface-2);
  color: var(--xm-ink-2);
  box-shadow: none;
}
.pm-mode-switch :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: var(--xm-accent-soft);
  border-color: var(--xm-accent);
  color: var(--xm-accent);
  box-shadow: none;
}
.pm-mode-switch :deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-radius: 8px 0 0 8px;
}
.pm-mode-switch :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 0 8px 8px 0;
}
</style>
