<template>
  <el-dialog
    v-model="visible"
    title="关联内容管理"
    width="600px"
    align-center
    append-to-body
    destroy-on-close
  >
    <el-tabs v-model="relateContentTab">
      <el-tab-pane label="章节" name="chapters">
        <div class="relate-content-list">
          <div v-if="chapters.length === 0" class="relate-empty">
            <el-empty description="暂无章节" />
          </div>
          <el-checkbox-group v-else v-model="selectedRelateChapterIds">
            <div
              v-for="chapter in chapters"
              :key="chapter.id"
              class="relate-item"
            >
              <el-checkbox :label="chapter.id" :value="chapter.id">
                {{ chapter.title || '无标题' }}
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </el-tab-pane>
      <el-tab-pane label="备忘录" name="memos">
        <div class="relate-content-list">
          <div v-if="memos.length === 0" class="relate-empty">
            <el-empty description="暂无备忘录" />
          </div>
          <el-checkbox-group v-else v-model="selectedRelateMemoIds">
            <div
              v-for="memo in memos"
              :key="memo.id"
              class="relate-item"
            >
              <el-checkbox :label="memo.id" :value="memo.id">
                {{ memo.title || '无标题' }}
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </el-tab-pane>
    </el-tabs>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="handleConfirm">确认关联</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Chapter, Memo } from '@/types'

const props = defineProps<{
  chapters: Chapter[]
  memos: Memo[]
  initialSelectedChapterIds: number[]
  initialSelectedMemoIds: number[]
}>()

const emit = defineEmits<{
  confirm: [payload: { selectedChapterIds: number[]; selectedMemoIds: number[] }]
}>()

const visible = defineModel<boolean>('visible')

const relateContentTab = ref('chapters')
const selectedRelateChapterIds = ref<number[]>([])
const selectedRelateMemoIds = ref<number[]>([])

watch(visible, (val) => {
  if (val) {
    selectedRelateChapterIds.value = [...props.initialSelectedChapterIds]
    selectedRelateMemoIds.value = [...props.initialSelectedMemoIds]
    relateContentTab.value = 'chapters'
  }
})

const handleConfirm = () => {
  emit('confirm', {
    selectedChapterIds: [...selectedRelateChapterIds.value],
    selectedMemoIds: [...selectedRelateMemoIds.value]
  })
  visible.value = false
}
</script>

<style scoped>
.relate-content-list {
  max-height: 400px;
  overflow-y: auto;
}

.relate-empty {
  padding: 24px 0;
}

.relate-item {
  padding: 6px 0;
}
</style>
