<template>
<!-- 预览弹窗 -->
<el-dialog
  v-model="previewDialogVisible"
  title="生成结果预览"
  width="50vw"
  destroy-on-close
  class="preview-dialog"
  append-to-body
>
  <div class="preview-dialog-content">
    <div class="preview-dialog-header">
      <el-button type="primary" link @click="copyPreviewContent">
        <el-icon><CopyDocument /></el-icon>
        复制
      </el-button>
    </div>
    <el-scrollbar class="preview-scrollbar">
      <MarkdownRenderer :content="previewContent" />
    </el-scrollbar>
  </div>
</el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

// 预览弹窗相关状态
const previewDialogVisible = ref(false)
const previewContent = ref('')

// 复制预览内容到剪贴板
const copyPreviewContent = async () => {
  try {
    await navigator.clipboard.writeText(previewContent.value)
    ElMessage.success('内容已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 打开预览弹窗
const open = (content: string) => {
  previewContent.value = content
  previewDialogVisible.value = true
}

defineExpose({ open })
</script>

<style scoped>
.preview-dialog-content {
  padding: 20px;
  max-height: 600px;
  display: flex;
  flex-direction: column;
}

.preview-dialog-header {
  margin-bottom: 15px;
  flex-shrink: 0;
}

.preview-scrollbar {
  flex-grow: 1;
  overflow-y: auto;
  max-height: 500px;
}
</style>
