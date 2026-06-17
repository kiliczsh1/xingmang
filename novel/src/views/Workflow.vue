<template>
  <div class="workflow-page">
    <div class="page-header">
      <div>
        <h1>工作流</h1>
        <p class="subtitle">创建你的工作流卡片，后续可以继续扩展节点与执行逻辑。</p>
      </div>

      <div class="header-actions">
        <el-button-group class="import-export-group">
          <el-button @click="openImportDialog">
            <el-icon><Upload /></el-icon>
            导入
          </el-button>
          <el-button @click="openExportDialog" :disabled="workflows.length === 0">
            <el-icon><Download /></el-icon>
            导出
          </el-button>
        </el-button-group>
        <el-button type="primary" class="create-btn" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          创建工作流
        </el-button>
      </div>
    </div>

    <div v-if="workflows.length > 0" class="workflow-grid">
      <article
        v-for="workflow in workflows"
        :key="workflow.id"
        class="workflow-card"
      >
        <div class="card-sheen"></div>
        <el-button
          class="delete-btn"
          type="info"
          :icon="Delete"
          circle
          @click.stop="deleteWorkflow(workflow)"
        />
        <div class="card-header">
          <h3 class="workflow-name">{{ workflow.name }}</h3>
        </div>

        <div class="card-center">
          <div class="signal-orbit orbit-1"></div>
          <div class="signal-orbit orbit-2"></div>
          <div class="signal-core">
            <el-icon><Connection /></el-icon>
          </div>
        </div>

        <div class="card-actions">
          <el-button plain class="card-btn edit-btn" @click="openEditDialog(workflow)">
            编辑
          </el-button>
          <el-button type="primary" class="card-btn enter-btn" @click="openWorkbench(workflow)">
            进入
          </el-button>
        </div>

        <div class="hover-panel">
          <span class="hover-label">简介</span>
          <p class="hover-description">
            {{ workflow.description || '这个工作流暂时还没有填写简介。' }}
          </p>
          <span class="hover-time">更新于 {{ formatDate(workflow.updatedAt) }}</span>
        </div>
      </article>
    </div>

    <div v-else class="empty-state">
      <div class="empty-shell">
        <el-icon class="empty-icon"><Connection /></el-icon>
        <h3>还没有工作流</h3>
        <p>先创建一个工作流卡片，名字会显示在顶部，简介会在悬停时展示。</p>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          立即创建
        </el-button>
      </div>
    </div>

    <el-dialog
      v-model="exportDialogVisible"
      title="导出工作流"
      width="480px"
      destroy-on-close
    >
      <div class="export-dialog-body">
        <p class="export-hint">请选择要导出的工作流：</p>
        <el-checkbox-group v-model="selectedExportIds" class="export-checkbox-group">
          <el-checkbox
            v-for="workflow in workflows"
            :key="workflow.id"
            :label="workflow.id"
            class="export-checkbox-item"
          >
            <span class="workflow-label">{{ workflow.name }}</span>
            <span class="workflow-desc">{{ workflow.description || '暂无简介' }}</span>
          </el-checkbox>
        </el-checkbox-group>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="exportDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="exportWorkflows" :disabled="selectedExportIds.length === 0">
            导出 ({{ selectedExportIds.length }})
          </el-button>
        </div>
      </template>
    </el-dialog>

    <input
      ref="importInputRef"
      type="file"
      accept=".json"
      style="display: none"
      @change="handleImportFile"
    />

    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑工作流' : '创建工作流'"
      width="560px"
      destroy-on-close
    >
      <div class="dialog-body">
        <div class="form-block">
          <label class="field-label required">名字</label>
          <el-input
            v-model="formData.name"
            maxlength="24"
            show-word-limit
            placeholder="例如：角色设定工作流"
          />
        </div>

        <div class="form-block">
          <label class="field-label">简介</label>
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="5"
            maxlength="200"
            show-word-limit
            resize="none"
            placeholder="简单描述这个工作流的用途和目标"
          />
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitWorkflow">
            {{ isEdit ? '保存修改' : '创建工作流' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Delete, Connection, Plus, Upload, Download } from '@element-plus/icons-vue'
import {
  loadWorkflows,
  createWorkflow,
  deleteWorkflow as deleteWorkflowAPI,
  updateWorkflowItem,
  getWorkflowSceneById,
  saveWorkflowScene,
  type WorkflowItem,
  type WorkflowScene
} from '@/utils/workflow'

interface ExportedWorkflow {
  name: string
  description: string
  scene: WorkflowScene | null
  exportedAt: string
}

interface ExportData {
  version: string
  workflows: ExportedWorkflow[]
}

const router = useRouter()
const workflows = ref<WorkflowItem[]>([])

const loadWorkflowList = async () => {
  workflows.value = await loadWorkflows()
}

onMounted(() => {
  loadWorkflowList()
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const editingWorkflowId = ref<number | null>(null)
const formData = ref({
  name: '',
  description: ''
})

const exportDialogVisible = ref(false)
const selectedExportIds = ref<number[]>([])
const importInputRef = ref<HTMLInputElement | null>(null)

const resetForm = () => {
  formData.value = {
    name: '',
    description: ''
  }
}

const openCreateDialog = () => {
  isEdit.value = false
  editingWorkflowId.value = null
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (workflow: WorkflowItem) => {
  isEdit.value = true
  editingWorkflowId.value = workflow.id
  formData.value = {
    name: workflow.name,
    description: workflow.description
  }
  dialogVisible.value = true
}

const submitWorkflow = async () => {
  const name = formData.value.name.trim()
  const description = formData.value.description.trim()

  if (!name) {
    ElMessage.warning('请先填写工作流名字')
    return
  }

  if (isEdit.value && editingWorkflowId.value !== null) {
    const updated = await updateWorkflowItem(editingWorkflowId.value, {
      name,
      description
    })
    if (updated) {
      workflows.value = workflows.value.map(w =>
        w.id === editingWorkflowId.value ? updated : w
      )
      ElMessage.success('工作流已更新')
    } else {
      ElMessage.error('更新失败')
    }
  } else {
    const created = await createWorkflow(name, description)
    if (created) {
      workflows.value = [created, ...workflows.value]
      ElMessage.success('工作流已创建')
    } else {
      ElMessage.error('创建失败')
    }
  }

  dialogVisible.value = false
  resetForm()
}

const openWorkbench = (workflow: WorkflowItem) => {
  router.push(`/workflow/${workflow.id}`)
}

const deleteWorkflow = (workflow: WorkflowItem) => {
  ElMessageBox.confirm(
    `确定要删除工作流"${workflow.name}"吗？删除后无法恢复。`,
    '删除工作流',
    {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    const success = await deleteWorkflowAPI(workflow.id)
    if (success) {
      workflows.value = workflows.value.filter(w => w.id !== workflow.id)
      ElMessage.success('工作流已删除')
    } else {
      ElMessage.error('删除失败')
    }
  }).catch(() => {
  })
}

const formatDate = (value: string) => {
  const date = new Date(value)

  if (Number.isNaN(date.getTime())) {
    return '刚刚'
  }

  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

const openExportDialog = () => {
  selectedExportIds.value = []
  exportDialogVisible.value = true
}

const exportWorkflows = async () => {
  if (selectedExportIds.value.length === 0) {
    ElMessage.warning('请至少选择一个工作流')
    return
  }

  const exportData: ExportData = {
    version: '1.0',
    workflows: []
  }

  for (const id of selectedExportIds.value) {
    const workflow = workflows.value.find(w => w.id === id)
    if (!workflow) continue

    const scene = await getWorkflowSceneById(id)
    exportData.workflows.push({
      name: workflow.name,
      description: workflow.description,
      scene,
      exportedAt: new Date().toISOString()
    })
  }

  const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `workflows_${new Date().toISOString().slice(0, 10)}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)

  ElMessage.success(`成功导出 ${exportData.workflows.length} 个工作流`)
  exportDialogVisible.value = false
}

const openImportDialog = () => {
  importInputRef.value?.click()
}

const handleImportFile = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  try {
    const text = await file.text()
    const data = JSON.parse(text) as ExportData

    if (!data.version || !Array.isArray(data.workflows)) {
      ElMessage.error('无效的工作流文件格式')
      return
    }

    let importedCount = 0
    for (const workflow of data.workflows) {
      const created = await createWorkflow(workflow.name, workflow.description)
      if (created && workflow.scene) {
        await saveWorkflowScene(created.id, workflow.scene)
        importedCount++
      }
    }

    if (importedCount > 0) {
      await loadWorkflowList()
      ElMessage.success(`成功导入 ${importedCount} 个工作流`)
    } else {
      ElMessage.warning('没有工作流被导入')
    }
  } catch (error) {
    console.error('Import error:', error)
    ElMessage.error('导入失败，请检查文件格式')
  } finally {
    input.value = ''
  }
}
</script>

<style scoped>
.workflow-page {
  min-height: calc(100vh - 48px);
  display: flex;
  flex-direction: column;
  gap: 28px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 20px;
  padding: 28px 32px;
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, rgba(255, 255, 255, 0.8) 0%, transparent 32%),
    linear-gradient(135deg, rgba(8, 198, 190, 0.14) 0%, rgba(5, 150, 145, 0.18) 100%);
  border: 1px solid rgba(8, 198, 190, 0.14);
  box-shadow: 0 20px 48px rgba(8, 198, 190, 0.09);
}

.page-header h1 {
  margin: 0 0 8px;
  font-size: 32px;
  line-height: 1.1;
  color: #0d5551;
}

.subtitle {
  margin: 0;
  color: #5b8784;
  font-size: 14px;
}

.create-btn {
  flex-shrink: 0;
  height: 42px;
  padding: 0 18px;
  border-radius: 14px;
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border: none;
  box-shadow: 0 12px 24px rgba(8, 198, 190, 0.22);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

.import-export-group {
  height: 42px;
}

.import-export-group .el-button {
  height: 42px;
  padding: 0 16px;
  border-radius: 14px;
  border-color: rgba(8, 198, 190, 0.3);
  color: #14706c;
  background: rgba(255, 255, 255, 0.8);
}

.import-export-group .el-button:hover {
  background: rgba(8, 198, 190, 0.1);
  border-color: rgba(8, 198, 190, 0.5);
}

.import-export-group .el-button:disabled {
  opacity: 0.5;
}

.export-dialog-body {
  padding: 8px 0;
}

.export-hint {
  margin: 0 0 16px;
  color: #5b8784;
  font-size: 14px;
}

.export-checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.export-checkbox-item {
  display: flex;
  align-items: flex-start;
  padding: 12px 16px;
  border-radius: 12px;
  background: rgba(8, 198, 190, 0.06);
  border: 1px solid rgba(8, 198, 190, 0.12);
  transition: all 0.2s ease;
  margin: 0;
  height: auto;
}

.export-checkbox-item:hover {
  background: rgba(8, 198, 190, 0.1);
  border-color: rgba(8, 198, 190, 0.2);
}

.export-checkbox-item :deep(.el-checkbox__label) {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

.workflow-label {
  font-size: 15px;
  font-weight: 600;
  color: #114846;
}

.workflow-desc {
  font-size: 12px;
  color: #648c89;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workflow-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  gap: 22px;
}

.workflow-card {
  position: relative;
  aspect-ratio: 1;
  padding: 18px;
  border-radius: 24px;
  overflow: hidden;
  isolation: isolate;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.96) 0%, rgba(237, 250, 249, 0.98) 100%);
  border: 1px solid rgba(8, 198, 190, 0.15);
  box-shadow:
    0 18px 36px rgba(8, 198, 190, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  transition:
    transform 0.28s ease,
    box-shadow 0.28s ease,
    border-color 0.28s ease;
}

.workflow-card:hover {
  transform: translateY(-6px);
  border-color: rgba(8, 198, 190, 0.28);
  box-shadow:
    0 24px 44px rgba(8, 198, 190, 0.16),
    inset 0 1px 0 rgba(255, 255, 255, 0.95);
}

.card-sheen {
  position: absolute;
  inset: -40% auto auto -30%;
  width: 150px;
  height: 150px;
  background: radial-gradient(circle, rgba(8, 198, 190, 0.18) 0%, transparent 72%);
  z-index: -1;
}

.card-header {
  position: relative;
  z-index: 2;
}

.delete-btn {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 3;
  width: 28px;
  height: 28px;
  padding: 0;
  border: none;
  background: rgba(245, 108, 108, 0.12);
  color: #f56c6c;
  opacity: 0;
  transition: opacity 0.28s ease, background 0.28s ease;
}

.workflow-card:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background: rgba(245, 108, 108, 0.24);
}

.workflow-name {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.3;
  color: #114846;
  letter-spacing: 0.02em;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-center {
  position: relative;
  height: 88px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.signal-orbit {
  position: absolute;
  border-radius: 50%;
  border: 1px solid rgba(8, 198, 190, 0.18);
}

.orbit-1 {
  width: 76px;
  height: 76px;
}

.orbit-2 {
  width: 112px;
  height: 112px;
  border-style: dashed;
  opacity: 0.7;
}

.signal-core {
  width: 52px;
  height: 52px;
  border-radius: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #08a19a;
  background: linear-gradient(135deg, rgba(8, 198, 190, 0.16) 0%, rgba(5, 150, 145, 0.08) 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.card-actions {
  position: relative;
  z-index: 2;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.card-btn {
  margin: 0;
  border-radius: 12px;
}

.edit-btn {
  border-color: rgba(8, 198, 190, 0.18);
  color: #14706c;
}

.enter-btn {
  background: linear-gradient(135deg, #08c6be 0%, #059691 100%);
  border: none;
}

.hover-panel {
  position: absolute;
  inset: 52px 12px 70px 12px;
  min-height: 0;
  padding: 14px;
  border-radius: 18px;
  background: rgba(9, 48, 46, 0.92);
  color: #edfefd;
  transform: translateY(14px);
  opacity: 0;
  pointer-events: none;
  transition:
    opacity 0.24s ease,
    transform 0.24s ease;
  backdrop-filter: blur(10px);
  z-index: 3;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
}

.workflow-card:hover .hover-panel {
  opacity: 1;
  transform: translateY(0);
}

.hover-label {
  font-size: 12px;
  letter-spacing: 0.08em;
  color: rgba(210, 255, 252, 0.78);
}

.hover-description {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: rgba(241, 255, 254, 0.96);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.hover-time {
  font-size: 12px;
  color: rgba(210, 255, 252, 0.65);
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 480px;
}

.empty-shell {
  width: min(520px, 100%);
  padding: 42px 28px;
  border-radius: 28px;
  text-align: center;
  background:
    radial-gradient(circle at top, rgba(8, 198, 190, 0.14) 0%, transparent 42%),
    rgba(255, 255, 255, 0.82);
  border: 1px dashed rgba(8, 198, 190, 0.24);
  box-shadow: 0 20px 44px rgba(8, 198, 190, 0.08);
}

.empty-icon {
  font-size: 42px;
  color: #08a19a;
}

.empty-shell h3 {
  margin: 18px 0 10px;
  font-size: 24px;
  color: #114846;
}

.empty-shell p {
  margin: 0 0 22px;
  color: #648c89;
  line-height: 1.7;
}

.dialog-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.field-label {
  font-size: 14px;
  font-weight: 600;
  color: #1b5f5b;
}

.field-label.required::after {
  content: ' *';
  color: #f56c6c;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-actions {
    flex-direction: column;
    gap: 10px;
  }

  .import-export-group {
    width: 100%;
  }

  .import-export-group .el-button {
    flex: 1;
  }

  .create-btn {
    width: 100%;
  }

  .workflow-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  }

  .workflow-card {
    border-radius: 20px;
    padding: 16px;
  }

  .hover-panel {
    inset: 48px 10px 64px 10px;
  }
}

/* 暗色主题适配 */
:root[data-theme='dark'] .signal-orbit {
  border-color: rgba(94, 234, 212, 0.2);
}

:root[data-theme='dark'] .signal-core {
  color: #5eead4;
  background: linear-gradient(135deg, rgba(94, 234, 212, 0.2) 0%, rgba(45, 212, 191, 0.1) 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

:root[data-theme='dark'] .workflow-card {
  background: linear-gradient(180deg, rgba(30, 41, 59, 0.98) 0%, rgba(15, 23, 42, 0.95) 100%);
  border-color: rgba(71, 85, 105, 0.4);
  box-shadow: 0 18px 36px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .workflow-card:hover {
  border-color: rgba(94, 234, 212, 0.4);
  box-shadow: 0 24px 44px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .card-sheen {
  background: radial-gradient(circle, rgba(94, 234, 212, 0.15) 0%, transparent 72%);
}

:root[data-theme='dark'] .workflow-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .edit-btn {
  border-color: rgba(94, 234, 212, 0.3);
  color: #5eead4;
}

:root[data-theme='dark'] .enter-btn {
  background: linear-gradient(135deg, #0d9488 0%, #0f766e 100%);
}

:root[data-theme='dark'] .hover-panel {
  background: rgba(15, 23, 42, 0.95);
  color: #edfefd;
}

:root[data-theme='dark'] .page-header {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.9) 0%, rgba(15, 23, 42, 0.95) 100%);
  border-color: rgba(71, 85, 105, 0.4);
  box-shadow: 0 20px 48px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .page-header h1 {
  color: #f3f4f6;
}

:root[data-theme='dark'] .subtitle {
  color: #9ca3af;
}

:root[data-theme='dark'] .create-btn {
  background: linear-gradient(135deg, #0d9488 0%, #0f766e 100%);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .import-export-group .el-button {
  border-color: rgba(94, 234, 212, 0.3);
  color: #5eead4;
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .import-export-group .el-button:hover {
  background: rgba(94, 234, 212, 0.15);
  border-color: rgba(94, 234, 212, 0.5);
}

:root[data-theme='dark'] .empty-shell {
  background: rgba(30, 41, 59, 0.8);
  border-color: rgba(94, 234, 212, 0.3);
  box-shadow: 0 20px 44px rgba(0, 0, 0, 0.3);
}

:root[data-theme='dark'] .empty-icon {
  color: #5eead4;
}

:root[data-theme='dark'] .empty-shell h3 {
  color: #f3f4f6;
}

:root[data-theme='dark'] .empty-shell p {
  color: #9ca3af;
}

:root[data-theme='dark'] .field-label {
  color: #e5e7eb;
}
</style>
