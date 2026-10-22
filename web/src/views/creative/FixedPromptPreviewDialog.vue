<template>
  <el-dialog
    :model-value="visible"
    title="发送给AI的完整提示词"
    width="600px"
    destroy-on-close
    modal-class="fixed-prompt-dialog-modal"
    append-to-body
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="fixed-prompt-dialog-content">
      <el-input
        :model-value="content"
        type="textarea"
        :rows="12"
        readonly
        placeholder="在此编辑完整提示词..."
        class="full-width"
      />
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="emit('update:visible', false)">取消</el-button>
        <el-button type="primary" @click="copyFixedPrompt">复制</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'

const props = defineProps<{
  visible: boolean
  content: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const copyFixedPrompt = async () => {
  try {
    await navigator.clipboard.writeText(props.content)
    ElMessage.success('提示词已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败，请手动复制')
  }
}
</script>
