<template>
  <el-dialog
    :model-value="visible"
    :title="`全屏编辑 - ${fieldLabel}`"
    width="80%"
    top="5vh"
    align-center
    append-to-body
    destroy-on-close
    class="field-fullscreen-dialog"
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="fullscreen-editor-container">
      <el-input
        v-model="content"
        type="textarea"
        :rows="20"
        :placeholder="`请输入${fieldLabel}`"
        resize="vertical"
        maxlength="10000"
        show-word-limit
      />
      <div class="fullscreen-editor-footer">
        <span class="char-count">{{ content.length }} / 10000</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleSave">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  fieldLabel: string
  initialContent: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'save', content: string): void
}>()

const content = ref('')

// 每次打开时用父组件传入的初始值初始化
watch(
  () => props.visible,
  (v) => {
    if (v) content.value = props.initialContent
  }
)

const handleSave = () => {
  emit('save', content.value)
  emit('update:visible', false)
}
</script>
