<template>
  <el-dialog v-model="visible" title="重命名对话" width="400px" append-to-body>
    <el-input v-model="renameTitle" placeholder="请输入新名称" />
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="confirmRename">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { conversationAPI } from '@/api'

const props = defineProps<{
  conversation: any
  currentTitle: string
}>()

const visible = defineModel<boolean>()
const emit = defineEmits<{
  renamed: [newTitle: string]
}>()

const renameTitle = ref('')

watch(visible, (val) => {
  if (val) {
    renameTitle.value = props.currentTitle
  }
})

const confirmRename = async () => {
  if (!renameTitle.value.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  const res = await conversationAPI.update(props.conversation.id, {
    title: renameTitle.value
  })
  if (res.success) {
    visible.value = false
    emit('renamed', renameTitle.value)
    ElMessage.success('重命名成功')
  }
}
</script>
