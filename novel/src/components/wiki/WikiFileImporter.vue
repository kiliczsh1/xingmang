<template>
  <div class="wiki-file-importer">
    <div
      class="drop-zone"
      :class="{ dragover: dragOver, disabled: !bookId }"
      @click="triggerFileInput"
      @dragover.prevent="dragOver = true"
      @dragleave.prevent="dragOver = false"
      @drop.prevent="onDrop"
    >
      <el-icon class="drop-icon" :size="36"><UploadFilled /></el-icon>
      <div class="drop-title">{{ bookId ? '点击或拖入 JSON 文件' : '请先选择书籍' }}</div>
      <div class="drop-desc">
        支持导出格式：<code>{ entities: [], relations: [] }</code>
      </div>
      <input
        ref="fileInputRef"
        type="file"
        accept=".json,application/json"
        style="display: none"
        @change="onFileChange"
      />
    </div>

    <div class="import-tips">
      <div class="tip-title">
        <el-icon><InfoFilled /></el-icon>
        导入说明
      </div>
      <ul class="tip-list">
        <li>导入后将以「新版本」形式保存到当前书籍</li>
        <li>支持星芒知识图谱的标准 JSON 导出格式</li>
        <li>导入不会覆盖现有版本，可在版本列表自由切换</li>
      </ul>
    </div>

    <!-- 预览弹窗 -->
    <el-dialog
      v-model="previewVisible"
      title="导入预览"
      width="480px"
      align-center
      append-to-body
      destroy-on-close
    >
      <div v-if="parsedData" class="preview-summary">
        <div class="preview-row">
          <span class="preview-label">文件来源</span>
          <span class="preview-value">{{ fileName }}</span>
        </div>
        <div class="preview-row">
          <span class="preview-label">默认名称</span>
          <span class="preview-value">{{ suggestedName }}</span>
        </div>
        <div class="preview-stats">
          <div class="preview-stat">
            <div class="preview-stat-num">{{ parsedData.entities?.length || 0 }}</div>
            <div class="preview-stat-label">实体</div>
          </div>
          <div class="preview-stat">
            <div class="preview-stat-num">{{ parsedData.relations?.length || 0 }}</div>
            <div class="preview-stat-label">关系</div>
          </div>
          <div class="preview-stat" v-if="entityTypeCount">
            <div class="preview-stat-num">{{ entityTypeCount }}</div>
            <div class="preview-stat-label">实体类型</div>
          </div>
        </div>
        <div class="preview-entity-types" v-if="Object.keys(entityTypeCountMap).length > 0">
          <el-tag
            v-for="(count, type) in entityTypeCountMap"
            :key="type"
            size="small"
            effect="plain"
            class="preview-type-tag"
          >
            {{ typeLabel(type) }} · {{ count }}
          </el-tag>
        </div>
        <el-alert
          v-if="!parsedData.entities || parsedData.entities.length === 0"
          type="error"
          :closable="false"
          title="未检测到任何实体，无法导入"
        />
      </div>
      <template #footer>
        <el-button @click="previewVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="importing"
          :disabled="!parsedData?.entities?.length"
          @click="confirmImport"
        >
          导入为新版本
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import type { GraphEntity, GraphRelation, GraphEntityType } from '@/types'
import { UploadFilled, InfoFilled } from '@element-plus/icons-vue'

const props = defineProps<{
  bookId: number | null
}>()

const emit = defineEmits<{
  (e: 'import', payload: {
    entities: GraphEntity[]
    relations: GraphRelation[]
    name: string
  }): void
}>()

const fileInputRef = ref<HTMLInputElement>()
const dragOver = ref(false)
const fileName = ref('')
const parsedData = ref<{ entities: GraphEntity[]; relations: GraphRelation[]; name?: string } | null>(null)
const suggestedName = ref('')
const previewVisible = ref(false)
const importing = ref(false)

const entityTypeCount = computed(() => Object.keys(entityTypeCountMap.value).length)

const entityTypeCountMap = computed(() => {
  const map: Record<string, number> = {}
  if (!parsedData.value?.entities) return map
  for (const e of parsedData.value.entities) {
    const t = e.type || 'unknown'
    map[t] = (map[t] || 0) + 1
  }
  return map
})

const TYPE_LABELS: Record<string, string> = {
  character: '角色',
  location: '地点',
  item: '物件',
  faction: '势力',
  event: '事件',
  skill: '功法',
  clue: '线索'
}

function typeLabel(t: string) {
  return TYPE_LABELS[t] || t
}

function triggerFileInput() {
  if (!props.bookId) {
    ElMessage.warning('请先选择书籍')
    return
  }
  fileInputRef.value?.click()
}

function onFileChange(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) handleFile(file)
  input.value = ''
}

function onDrop(e: DragEvent) {
  dragOver.value = false
  if (!props.bookId) {
    ElMessage.warning('请先选择书籍')
    return
  }
  const file = e.dataTransfer?.files?.[0]
  if (file) handleFile(file)
}

async function handleFile(file: File) {
  if (!file.name.toLowerCase().endsWith('.json') && file.type !== 'application/json') {
    ElMessage.warning('仅支持 .json 文件')
    return
  }
  try {
    const text = await file.text()
    const data = JSON.parse(text)
    const entities = data.entities || data.data?.entities
    const relations = data.relations || data.data?.relations
    if (!Array.isArray(entities)) {
      ElMessage.error('文件格式不正确，缺少 entities 数组')
      return
    }
    parsedData.value = {
      entities,
      relations: Array.isArray(relations) ? relations : [],
      name: data.name
    }
    fileName.value = file.name
    suggestedName.value = data.name || file.name.replace(/\.json$/i, '')
    previewVisible.value = true
  } catch (e) {
    console.error('解析失败:', e)
    ElMessage.error('JSON 解析失败，请检查文件内容')
  }
}

async function confirmImport() {
  if (!parsedData.value) return
  importing.value = true
  try {
    const name = suggestedName.value || `导入_${Date.now()}`
    emit('import', {
      entities: parsedData.value.entities,
      relations: parsedData.value.relations || [],
      name
    })
    previewVisible.value = false
    parsedData.value = null
  } finally {
    importing.value = false
  }
}
</script>

<style scoped>
.wiki-file-importer {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px 2px;
}

.drop-zone {
  border: 1.5px dashed rgba(45, 212, 191, 0.4);
  border-radius: 10px;
  padding: 28px 16px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  background: rgba(255, 255, 255, 0.4);
}

.drop-zone:hover:not(.disabled) {
  border-color: rgba(45, 212, 191, 0.7);
  background: rgba(45, 212, 191, 0.05);
}

.drop-zone.dragover {
  border-color: #2dd4bf;
  background: rgba(45, 212, 191, 0.12);
  transform: scale(1.01);
}

.drop-zone.disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.drop-icon {
  color: #2dd4bf;
  margin-bottom: 8px;
}

.drop-title {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  margin-bottom: 4px;
}

.drop-desc {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  line-height: 1.5;
}

.drop-desc code {
  background: rgba(45, 212, 191, 0.1);
  padding: 1px 4px;
  border-radius: 3px;
  font-size: 10px;
  color: #2dd4bf;
}

.import-tips {
  padding: 10px 12px;
  background: rgba(45, 212, 191, 0.05);
  border-radius: 8px;
  border-left: 3px solid #2dd4bf;
}

.tip-title {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  margin-bottom: 6px;
}

.tip-title .el-icon {
  color: #2dd4bf;
  font-size: 13px;
}

.tip-list {
  margin: 0;
  padding-left: 18px;
  font-size: 11px;
  color: var(--el-text-color-secondary);
  line-height: 1.7;
}

.preview-summary {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.preview-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}

.preview-label {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  width: 70px;
}

.preview-value {
  color: var(--el-text-color-primary);
  font-weight: 500;
  word-break: break-all;
}

.preview-stats {
  display: flex;
  gap: 10px;
  margin: 4px 0;
}

.preview-stat {
  flex: 1;
  padding: 10px;
  background: rgba(45, 212, 191, 0.06);
  border-radius: 8px;
  text-align: center;
}

.preview-stat-num {
  font-size: 20px;
  font-weight: 600;
  color: #2dd4bf;
  line-height: 1.2;
}

.preview-stat-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.preview-entity-types {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.preview-type-tag {
  font-size: 10px;
}
</style>
