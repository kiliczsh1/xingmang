<template>
  <el-dialog v-model="visible" title="移动到文件夹" width="420px" append-to-body>
    <div class="move-folder-content">
      <p class="move-folder-hint">选择要移动到的目标文件夹：</p>
      <div class="folder-list">
        <button
          type="button"
          class="folder-option"
          :class="{ active: targetFolderId === null }"
          @click="targetFolderId = null"
        >
          <el-icon><Folder /></el-icon>
          <span>根目录</span>
        </button>
        <button
          v-for="folder in folders"
          :key="folder.id"
          type="button"
          class="folder-option"
          :class="{ active: targetFolderId === folder.id }"
          @click="targetFolderId = folder.id"
        >
          <el-icon><Folder /></el-icon>
          <span>{{ folder.title }}</span>
        </button>
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleConfirm">移动</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { Folder } from '@element-plus/icons-vue'

const props = defineProps<{
  folders: any[]
  chapterToMove: any
}>()

const visible = defineModel<boolean>('visible')
const emit = defineEmits<{
  moved: [chapterId: number, targetFolderId: number | null]
}>()

const targetFolderId = ref<number | null>(null)

watch(visible, (val) => {
  if (val && props.chapterToMove) {
    targetFolderId.value = props.chapterToMove.volume_id || null
  }
})

const handleConfirm = () => {
  if (!props.chapterToMove) return
  emit('moved', props.chapterToMove.id, targetFolderId.value)
  visible.value = false
}
</script>

<style scoped>
.move-folder-content {
  padding: 10px 0;
}

.move-folder-hint {
  margin: 0 0 16px;
  color: #606266;
  font-size: 13px;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.folder-option {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 12px 16px;
  border: 1px solid #e4e7ed;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.folder-option:hover {
  border-color: #409eff;
  background: #ecf5ff;
  transform: translateX(4px);
}

.folder-option.active {
  border-color: #409eff;
  background: #ecf5ff;
}

.folder-option .el-icon {
  color: #e6a23c;
  font-size: 18px;
}
</style>
