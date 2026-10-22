<template>
  <div class="file-explorer">
    <div class="explorer-header">
      <div class="header-title"><span>工作流</span></div>
      <div class="header-actions">
        <el-tooltip content="刷新" placement="bottom">
          <el-button text size="small" class="header-btn" @click="$emit('refresh')">
            <i class="fas fa-rotate" />
          </el-button>
        </el-tooltip>
        <el-dropdown trigger="click" @command="onCommand">
          <el-button text size="small" class="header-btn"><i class="fas fa-plus" /></el-button>
          <template #dropdown>
            <el-dropdown-menu class="file-type-dropdown">
              <el-dropdown-item command="workflow" class="file-type-item">
                <div class="file-type-row">
                  <div class="file-type-icon workflow-icon"><i class="fas fa-project-diagram" /></div>
                  <div class="file-type-info">
                    <span class="file-type-title">工作流</span>
                    <span class="file-type-ext">.workflow.json</span>
                  </div>
                </div>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </div>

    <div class="search-box">
      <el-input
        v-model="keyword"
        size="small"
        clearable
        placeholder="输入过滤文件名"
        @input="$emit('filter', keyword)"
      >
        <template #prefix><i class="fas fa-search" /></template>
      </el-input>
    </div>

    <div class="explorer-content">
      <div v-if="!files.length" class="empty-state">
        <i class="fas fa-folder-open empty-icon" />
        <p class="empty-text">暂无工作流</p>
        <p class="empty-hint">点击上方 <i class="fas fa-plus" /> 按钮新建工作流</p>
      </div>
      <div v-else class="vscode-file-tree">
        <div class="tree-root">
          <div
            v-for="f in filtered"
            :key="f.id"
            class="tree-node"
            :class="{ active: f.id === activeId }"
            :style="{ paddingLeft: '8px' }"
            @click="$emit('open', f)"
            @contextmenu.prevent.stop="onContext($event, f)"
          >
            <div class="expand-icon"><span class="icon-placeholder" /></div>
            <i class="fas fa-project-diagram node-icon workflow-icon" />
            <span class="node-label">{{ f.label }}</span>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="menu.visible"
      class="context-menu"
      :style="{ left: menu.x + 'px', top: menu.y + 'px' }"
      @click.stop
    >
      <div class="menu-item" @click="onRename"><i class="fas fa-edit" /><span>重命名</span></div>
      <div class="menu-item" @click="onDelete"><i class="fas fa-trash" /><span>删除</span></div>
      <div class="menu-divider" />
      <div class="menu-item" @click="onCopyPath">
        <i class="fas fa-copy" /><span>复制名称</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { WorkflowFile } from '@/types/workflow'

const props = defineProps<{
  files: WorkflowFile[]
  activeId: number | null
}>()

const emit = defineEmits<{
  (e: 'open', file: WorkflowFile): void
  (e: 'create'): void
  (e: 'refresh'): void
  (e: 'rename', file: WorkflowFile): void
  (e: 'delete', file: WorkflowFile): void
  (e: 'filter', keyword: string): void
}>()

const keyword = ref('')
const menu = reactive({ visible: false, x: 0, y: 0, node: null as WorkflowFile | null })

/** 源版节点 label = 文件名（带 .workflow.json 后缀） */
const decorated = computed(() =>
  props.files.map(f => ({ ...f, label: `${f.name}.workflow.json` }))
)

const filtered = computed(() => {
  const k = keyword.value.trim().toLowerCase()
  if (!k) return decorated.value
  return decorated.value.filter(f => f.label.toLowerCase().includes(k))
})

const onCommand = (cmd: string) => {
  if (cmd === 'workflow') emit('create')
}

const onContext = (e: MouseEvent, f: WorkflowFile) => {
  const menuH = 110
  const menuW = 200
  let x = e.clientX
  let y = e.clientY
  if (y + menuH > window.innerHeight) y = Math.max(10, window.innerHeight - menuH - 10)
  if (x + menuW > window.innerWidth) x = Math.max(10, window.innerWidth - menuW - 10)
  menu.visible = true
  menu.x = x
  menu.y = y
  menu.node = f
}

const closeMenu = () => {
  menu.visible = false
}
onMounted(() => document.addEventListener('click', closeMenu))
onBeforeUnmount(() => document.removeEventListener('click', closeMenu))

const onRename = () => {
  const f = menu.node
  closeMenu()
  if (f) emit('rename', f)
}
const onDelete = async () => {
  const f = menu.node
  closeMenu()
  if (!f) return
  try {
    await ElMessageBox.confirm(`删除「${f.name}」会同时删除它的执行记录，且不可恢复。`, '确认删除', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
  } catch {
    return
  }
  emit('delete', f)
}
const onCopyPath = async () => {
  const f = menu.node
  closeMenu()
  if (!f) return
  try {
    await navigator.clipboard.writeText(`${f.name}.workflow.json`)
    ElMessage.success('已复制名称到剪贴板')
  } catch {
    ElMessage.error('复制失败')
  }
}
</script>

<style scoped>
/* 源版 .file-explorer（scope e911a4e6）+ .tree-node（scope e72fb91d）*/
.file-explorer {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--wf-el-bg, #ffffff);
  border-right: 1px solid var(--wf-el-border, #dcdfe6);
}
.explorer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
  background: var(--wf-el-bg-page, #f2f3f5);
}
.header-title {
  display: flex;
  align-items: center;
  font-weight: 600;
  font-size: 13px;
  color: var(--wf-el-text-primary, #303133);
}
.header-actions {
  display: flex;
  gap: 2px;
}
.header-actions .header-btn {
  padding: 4px;
  transition: none !important;
}
.header-actions .header-btn:hover {
  background: transparent !important;
  box-shadow: none !important;
  transform: none !important;
}
.file-type-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  margin: 0 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
  cursor: pointer;
  width: calc(100% - 8px);
}
.file-type-row:hover {
  background: var(--wf-el-primary-light-9, #ecf5ff);
}
.file-type-icon {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  font-size: 14px;
  flex-shrink: 0;
}
.workflow-icon {
  color: #ff9800;
}
.file-type-title {
  font-size: 13px;
  color: var(--wf-el-text-primary, #303133);
  white-space: nowrap;
}
.file-type-ext {
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  font-family: monospace;
  opacity: 0.7;
}
.search-box {
  padding: 6px 10px;
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
}
.search-box :deep(.el-input__wrapper) {
  box-shadow: none;
  background: var(--wf-el-fill-light, #f5f7fa);
  padding: 0 8px;
}
.search-box :deep(.el-input__inner) {
  font-size: 12px;
}
.explorer-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}
.explorer-content::-webkit-scrollbar {
  width: 6px;
}
.explorer-content::-webkit-scrollbar-track {
  background: transparent;
}
.explorer-content::-webkit-scrollbar-thumb {
  background: var(--wf-el-border, #dcdfe6);
  border-radius: 3px;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: var(--wf-el-text-secondary, #909399);
}
.empty-icon {
  font-size: 40px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
  margin-bottom: 12px;
}
.empty-text {
  font-size: 14px;
  margin: 0 0 6px;
}
.empty-hint {
  font-size: 12px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
  margin: 0;
}
.vscode-file-tree {
  width: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  font-size: 13px;
  color: var(--wf-el-text-primary, #303133);
  user-select: none;
}
.tree-root {
  width: 100%;
}
.tree-node {
  display: flex;
  align-items: center;
  height: 22px;
  cursor: pointer;
  transition: background-color 0.1s;
  position: relative;
}
.tree-node:hover {
  background-color: var(--wf-el-fill-light, #f5f7fa);
}
.tree-node.active {
  background-color: var(--wf-el-primary-light-9, #ecf5ff);
  font-weight: 500;
}
.expand-icon {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 10px;
  color: var(--wf-el-text-secondary, #909399);
}
.icon-placeholder {
  width: 10px;
}
.node-icon {
  width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin: 0 6px;
  font-size: 14px;
}
.node-label {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}
.context-menu {
  position: fixed;
  z-index: 9999;
  min-width: 180px;
  background: var(--wf-el-bg-overlay, #ffffff);
  border: 1px solid var(--wf-el-border, #dcdfe6);
  border-radius: 6px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  padding: 4px 0;
}
.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}
.menu-item:hover {
  background-color: var(--wf-el-fill-light, #f5f7fa);
}
.menu-item i {
  width: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--wf-el-text-secondary, #909399);
}
.menu-divider {
  height: 1px;
  background-color: var(--wf-el-border-light, #ebeef5);
  margin: 4px 0;
}
</style>
