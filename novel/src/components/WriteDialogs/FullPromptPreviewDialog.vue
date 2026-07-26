<template>
  <el-dialog
    v-model="visible"
    title="完整提示词预览"
    width="800px"
    top="3vh"
    append-to-body
    destroy-on-close
  >
    <div class="full-prompt-content">
      <pre>{{ content }}</pre>
    </div>
    <template #footer>
      <el-button @click="visible = false">关闭</el-button>
      <el-button type="primary" @click="copyFullPrompt">复制全部</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'

const props = defineProps<{
  content: string
}>()

const visible = defineModel<boolean>('visible')

const copyFullPrompt = () => {
  navigator.clipboard.writeText(props.content)
  ElMessage.success('已复制完整提示词')
}
</script>

<style scoped>
.full-prompt-content {
  max-height: 60vh;
  overflow-y: auto;
}

.full-prompt-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.6;
  color: var(--el-text-color-primary);
}
</style>
