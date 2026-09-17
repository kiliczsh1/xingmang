<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑文件夹' : '新建文件夹'" width="400px" append-to-body>
    <el-input
      v-model="formTitle"
      placeholder="请输入文件夹名称"
      maxlength="50"
      show-word-limit
    />
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { volumeAPI } from '@/api'
import type { Volume } from '@/types'

const props = defineProps<{
  bookId: number
  isEdit: boolean
  volumeId?: number
  initialTitle?: string
  parentId?: number | null
  volumes: Volume[]
  expandedFolderIds: (number | string)[]
}>()

const visible = defineModel<boolean>('visible')
const emit = defineEmits<{
  submitted: []
}>()

const formTitle = ref('')

watch(visible, (val) => {
  if (val) {
    formTitle.value = props.isEdit ? (props.initialTitle || '') : ''
  }
})

const handleSubmit = async () => {
  if (!formTitle.value.trim()) {
    ElMessage.warning('请输入文件夹名称')
    return
  }

  try {
    if (props.isEdit) {
      const res = await volumeAPI.update(props.volumeId!, {
        title: formTitle.value
      })
      if (res.success && res.data) {
        const index = props.volumes.findIndex(v => v.id === props.volumeId)
        if (index !== -1) {
          props.volumes[index] = res.data
        }
        ElMessage.success('更新成功')
      }
    } else {
      const res = await volumeAPI.create({
        book_id: props.bookId,
        parent_id: props.parentId || undefined,
        title: formTitle.value,
        order_num: props.volumes.length
      })
      if (res.success && res.data) {
        props.volumes.push(res.data)
        if (!props.expandedFolderIds.includes(res.data.id)) {
          props.expandedFolderIds.push(res.data.id)
        }
        if (props.parentId && !props.expandedFolderIds.includes(props.parentId)) {
          props.expandedFolderIds.push(props.parentId)
        }
        ElMessage.success('创建成功')
      }
    }
    visible.value = false
    emit('submitted')
  } catch (error) {
    ElMessage.error(props.isEdit ? '更新失败' : '创建失败')
  }
}
</script>
