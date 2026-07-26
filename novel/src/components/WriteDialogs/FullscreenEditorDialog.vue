<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="80%"
    top="5vh"
    align-center
    append-to-body
    destroy-on-close
  >
    <div class="fullscreen-editor-container">
      <el-input
        v-model="content"
        type="textarea"
        :rows="20"
        placeholder="请输入内容..."
        resize="vertical"
      />
      <div class="fullscreen-editor-footer">
        <span class="char-count">{{ content.length }} / 10000</span>
      </div>
    </div>
    <template #footer>
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" @click="handleSave">确认</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
const visible = defineModel<boolean>('visible', { required: true })
const content = defineModel<string>({ required: true })

defineProps<{
  title?: string
}>()

const emit = defineEmits<{
  save: [content: string]
  cancel: []
}>()

const handleSave = () => {
  emit('save', content.value)
  visible.value = false
}

const handleCancel = () => {
  emit('cancel')
  visible.value = false
}
</script>

<style scoped>
.fullscreen-editor-container {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.fullscreen-editor-container :deep(.el-textarea__inner) {
  font-size: 15px;
  line-height: 1.6;
  min-height: 400px;
}

.fullscreen-editor-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.fullscreen-editor-footer .char-count {
  font-size: 13px;
  color: #6b7280;
}
</style>
