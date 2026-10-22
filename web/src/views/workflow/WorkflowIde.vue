<template>
  <div class="wf-ide" :class="{ 'wf-dark': isDark }">
    <div class="ide-layout">
      <!-- 左侧：文件资源管理器 -->
      <div v-show="ide.leftSidebarVisible" class="sidebar left-sidebar" :style="{ width: ide.leftSidebarWidth + 'px' }">
        <FileExplorer
          :files="workflows"
          :active-id="ide.activeFileId"
          @open="openWorkflow"
          @create="createWorkflow"
          @refresh="loadList"
          @rename="startRename"
          @delete="deleteWorkflow"
        />
        <div class="resize-handle left" @mousedown="startResize('left', $event)" />
      </div>

      <!-- 中间：标签页 + 编辑器 -->
      <div class="editor-area">
        <div class="file-tabs">
          <div v-if="ide.openFiles.length" ref="tabsListRef" class="tabs-list" @wheel="onTabsWheel">
            <div
              v-for="f in ide.openFiles"
              :key="f.id"
              class="file-tab"
              :class="{ active: f.id === ide.activeFileId }"
              @click.stop="ide.activeFileId = f.id"
            >
              <i class="fas fa-project-diagram workflow-icon" />
              <span class="tab-label">{{ f.label }}</span>
              <button class="close-btn" @click.stop="ide.closeFile(f.id)"><i class="fas fa-times" /></button>
            </div>
          </div>
          <div v-else class="tabs-list-empty"><span class="empty-text">未打开文件</span></div>
          <div class="tabs-toolbar">
            <el-tooltip content="保存 (Ctrl+S)" placement="bottom">
              <el-button text class="toolbar-btn" :disabled="!activeFile" @click="saveActive">
                <i class="fas fa-save" />
              </el-button>
            </el-tooltip>
            <el-tooltip content="编辑器设置" placement="bottom">
              <el-button text class="toolbar-btn" @click="settingsVisible = true">
                <i class="fas fa-cog" />
              </el-button>
            </el-tooltip>
          </div>
        </div>

        <div class="editor-container">
          <!-- 工作流工具栏（源版：仅工作流文件显示） -->
          <div v-if="activeFile" class="workflow-toolbar">
            <div class="workflow-toolbar-left">
              <el-button-group>
                <el-button
                  size="small"
                  :type="activeFile.viewMode === 'editor' ? 'primary' : ''"
                  @click="switchViewMode('editor')"
                >
                  <i class="fas fa-code" /> 代码
                </el-button>
                <el-button
                  size="small"
                  :type="activeFile.viewMode === 'visual' ? 'primary' : ''"
                  @click="switchViewMode('visual')"
                >
                  <i class="fas fa-project-diagram" /> 可视化
                </el-button>
              </el-button-group>
            </div>
            <div v-if="activeFile.viewMode === 'editor'" class="workflow-toolbar-right">
              <el-button size="small" type="primary" @click="generatorVisible = true">
                <i class="fas fa-magic" /> AI生成工作流
              </el-button>
            </div>
            <div v-else class="workflow-toolbar-right">
              <el-button
                size="small"
                :type="paletteVisible ? 'primary' : ''"
                @click="paletteVisible = !paletteVisible"
              >
                <i class="fas fa-shapes" /> 节点库
              </el-button>
              <el-dropdown trigger="click" @command="addNode">
                <el-button size="small"><i class="fas fa-plus" /> 添加节点</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="start" divided><i class="fas fa-play-circle" /> 开始节点</el-dropdown-item>
                    <el-dropdown-item command="aiProcess"><i class="fas fa-robot" /> AI 处理节点</el-dropdown-item>
                    <el-dropdown-item command="text"><i class="fas fa-file-text" /> 文本节点</el-dropdown-item>
                    <el-dropdown-item command="merge"><i class="fas fa-code-merge" /> 合并节点</el-dropdown-item>
                    <el-dropdown-item command="loop"><i class="fas fa-rotate" /> 循环节点</el-dropdown-item>
                    <el-dropdown-item command="context"><i class="fas fa-database" /> 上下文节点</el-dropdown-item>
                    <el-dropdown-item command="end" divided><i class="fas fa-stop-circle" /> 结束节点</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-button size="small" :disabled="!selectedNode" @click="deleteSelectedNode">
                <i class="fas fa-trash" /> 删除
              </el-button>
              <el-divider direction="vertical" style="margin: 0 8px" />
              <el-button-group>
                <el-tooltip content="顺序引擎：从开始节点沿连线递归执行（默认）" placement="bottom">
                  <el-button
                    size="small"
                    :type="runMode === 'sequential' ? 'primary' : ''"
                    @click="runMode = 'sequential'"
                  >
                    <i class="fas fa-arrow-right" /> 顺序
                  </el-button>
                </el-tooltip>
                <el-tooltip content="并发引擎：拓扑就绪队列 + 限流，可并行执行互不依赖的节点" placement="bottom">
                  <el-button
                    size="small"
                    :type="runMode === 'concurrent' ? 'primary' : ''"
                    @click="runMode = 'concurrent'"
                  >
                    <i class="fas fa-bolt" /> 并发
                  </el-button>
                </el-tooltip>
              </el-button-group>
              <el-tooltip content="并发参数" placement="bottom">
                <el-button size="small" @click="limitsVisible = true">
                  <i class="fas fa-sliders" />
                </el-button>
              </el-tooltip>
              <el-divider direction="vertical" style="margin: 0 8px" />
              <el-button v-if="!store.isExecuting && !store.isPaused" size="small" type="primary" @click="runWorkflow">
                <i class="fas fa-play" /> 运行
              </el-button>
              <el-button v-else size="small" @click="stopRun"><i class="fas fa-stop" /> 停止</el-button>
              <el-divider direction="vertical" style="margin: 0 8px" />
              <el-button size="small" @click="ide.logsPanelVisible = !ide.logsPanelVisible">
                <i class="fas fa-list" /> 日志
              </el-button>
              <el-button size="small" @click="historyVisible = true">
                <i class="fas fa-clock-rotate-left" /> 历史
              </el-button>
            </div>
          </div>

          <!-- 编辑器主体 -->
          <div v-if="activeFile" class="editor-wrapper">
            <!-- 代码模式 -->
            <MonacoEditor
              v-if="activeFile.viewMode === 'editor'"
              :key="`monaco-${activeFile.id}`"
              v-model="activeFile.content"
              language="json"
              :dark="isDark"
              :font-size="ide.editorFontSize"
              :font-family="ide.editorFontFamily"
              @change="onMonacoChange"
            />
            <!-- 可视化模式 -->
            <div
              v-else
              :key="`visual-${activeFile.id}`"
              class="workflow-visual-container"
            >
              <NodePalette v-if="paletteVisible" @add="addNode" />
              <div class="workflow-visual-editor">
                <div v-if="!visualNodes || !visualEdges" class="workflow-loading">
                  <i class="fas fa-spinner" />
                  <span>正在加载工作流...</span>
                </div>
                <WorkflowCanvas
                  v-else
                  ref="canvasRef"
                  :initial-nodes="visualNodes"
                  :initial-edges="visualEdges"
                  @node-click="onNodeClick"
                  @node-added="onNodeAdded"
                  @test-run-node="onTestRunNode"
                  @rerun-node="onRerunNode"
                  @nodes-update="onNodesUpdate"
                  @edges-update="onEdgesUpdate"
                />
                <BreakpointPanel
                  v-if="store.isPaused"
                  :info="store.checkpointInfo"
                  @continue="onBreakpointContinue"
                  @stop="onBreakpointStop"
                />
              </div>
              <transition name="el-fade-in-linear">
                <LogsPanel
                  v-if="ide.logsPanelVisible"
                  :logs="store.executionLogs"
                  @clear="store.clearExecutionLogs()"
                />
              </transition>
            </div>
          </div>

          <!-- 欢迎页 -->
          <div v-else class="welcome-screen">
            <div class="welcome-content">
              <div v-if="ide.recentFiles.length" class="recent-files-section">
                <div class="section-header"><i class="fas fa-history" /><span>最近打开</span></div>
                <div class="recent-files-list">
                  <div
                    v-for="r in ide.recentFiles.slice(0, 3)"
                    :key="r.id"
                    class="recent-file-item"
                    @click="reopenRecent(r.id)"
                  >
                    <i class="fas fa-project-diagram file-icon" style="color: #ff9800" />
                    <div class="file-info">
                      <div class="file-name">{{ r.name }}</div>
                      <div class="file-path">{{ r.fullPath }}</div>
                    </div>
                    <div class="file-time">{{ formatRecentTime(r.lastOpened) }}</div>
                  </div>
                </div>
              </div>
              <div class="quick-actions-section">
                <div class="section-header"><i class="fas fa-bolt" /><span>快速开始</span></div>
                <div class="quick-actions">
                  <div class="action-card" @click="createWorkflow">
                    <i class="fas fa-feather" />
                    <div class="action-title">新建工作流</div>
                    <div class="action-desc">创建一张空白画布，从节点库拖入节点开始</div>
                  </div>
                  <div class="action-card" @click="createNovelWorkflow">
                    <i class="fas fa-graduation-cap" />
                    <div class="action-title">章节流水线模板</div>
                    <div class="action-desc">开始 → 循环逐章 AI 处理 → 攒历史 → 汇总</div>
                  </div>
                  <div class="action-card" @click="loadList">
                    <i class="fas fa-folder-open" />
                    <div class="action-title">刷新列表</div>
                    <div class="action-desc">重新拉取工作流列表</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 底部统计 -->
          <div v-if="activeFile" class="editor-footer">
            <div class="footer-stats">
              <div class="stat-item">
                <i class="fas fa-file-alt" />
                <span class="stat-value">{{ footerStats.wordCount }}</span>
              </div>
              <div class="stat-item">
                <i class="fas fa-paragraph" />
                <span class="stat-value">{{ footerStats.paragraphs }}</span>
              </div>
              <div class="stat-item">
                <i class="fas fa-clock" />
                <span class="stat-value">{{ footerStats.readTime }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 节点属性抽屉 -->
    <NodePropertyDrawer
      :visible="!!selectedNode && drawerVisible"
      :node="selectedNode"
      :known-variables="knownVariables"
      :nodes="store.currentWorkflow?.nodes || []"
      :edges="store.currentWorkflow?.edges || []"
      @close="drawerVisible = false"
      @change="onNodeConfigChange"
    />

    <!-- 编辑器设置 -->
    <el-dialog v-model="settingsVisible" title="编辑器设置" width="600px" append-to-body class="editor-settings-dialog">
      <div class="settings-content">
        <div class="setting-row">
          <div class="setting-label"><i class="fas fa-font" /> 字体大小</div>
          <div class="size-controls">
            <el-input-number v-model="ide.editorFontSize" :min="10" :max="32" />
          </div>
        </div>
        <div class="setting-row">
          <div class="setting-label"><i class="fas fa-font" /> 字体</div>
          <el-input v-model="ide.editorFontFamily" />
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button size="small" @click="resetEditorSettings"><i class="fas fa-undo" /> 恢复默认</el-button>
          <el-button size="small" type="primary" @click="settingsVisible = false">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- AI 生成工作流 -->
    <AiGeneratorDialog v-model="generatorVisible" :existing="activeFile?.content || ''" @apply="applyGenerated" />

    <!-- 并发参数 -->
    <el-dialog v-model="limitsVisible" title="并发参数" width="440px" append-to-body>
      <el-form label-width="140px" @submit.prevent>
        <el-form-item label="同时执行节点数上限">
          <el-input-number v-model="store.concurrencyLimits.total" :min="1" :max="16" />
        </el-form-item>
        <el-form-item label="其中 AI 节点上限">
          <el-input-number v-model="store.concurrencyLimits.ai" :min="1" :max="16" />
        </el-form-item>
      </el-form>
      <div class="limits-hint">
        AI 节点会真实调用模型接口，并发过高容易触发上游限流，建议保持 AI ≤ 2。循环体内部始终串行执行，不受这里影响。
      </div>
      <template #footer>
        <el-button size="small" @click="limitsVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 执行历史 -->
    <RunHistoryDialog v-model="historyVisible" :workflow-id="activeFile?.id ?? null" />

    <!-- 重命名 -->
    <el-dialog v-model="renameVisible" title="重命名工作流" width="440px" append-to-body>
      <el-form label-width="70px" @submit.prevent>
        <el-form-item label="名称">
          <el-input v-model="renameName" maxlength="60" show-word-limit @keyup.enter="submitRename" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renameVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRename">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import FileExplorer from './components/FileExplorer.vue'
import WorkflowCanvas from './components/WorkflowCanvas.vue'
import LogsPanel from './components/LogsPanel.vue'
import MonacoEditor from './components/MonacoEditor.vue'
import NodePropertyDrawer from './components/NodePropertyDrawer.vue'
import AiGeneratorDialog from './components/AiGeneratorDialog.vue'
import NodePalette from './components/NodePalette.vue'
import BreakpointPanel from './components/BreakpointPanel.vue'
import RunHistoryDialog from './components/RunHistoryDialog.vue'
import workflowAPI, { toWorkflowFile } from '@/services/workflow/workflowAPI'
import { createNodeData } from '@/services/workflow/nodeCatalog'
import { WorkflowExecutor } from '@/services/workflow/executor'
import { useWorkflowStore } from '@/stores/workflow'
import { useWorkflowIdeStore, type IdeFile } from '@/stores/workflowIde'
import type { NodeType, WEdge, WNode, WorkflowFile } from '@/types/workflow'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const store = useWorkflowStore()
const ide = useWorkflowIdeStore()

const workflows = ref<WorkflowFile[]>([])
const canvasRef = ref<InstanceType<typeof WorkflowCanvas> | null>(null)
const tabsListRef = ref<HTMLElement | null>(null)
const drawerVisible = ref(false)
/** 节点库侧栏（可视化模式左侧，可折叠） */
const paletteVisible = ref(true)
/** 执行引擎：顺序（沿边递归，默认） / 并发（拓扑就绪队列） */
const runMode = ref<'sequential' | 'concurrent'>('sequential')
const limitsVisible = ref(false)
const historyVisible = ref(false)
const settingsVisible = ref(false)
const generatorVisible = ref(false)
const renameVisible = ref(false)
const renameTarget = ref<WorkflowFile | null>(null)
const renameName = ref('')
const isDark = ref(document.documentElement.getAttribute('data-theme') === 'dark')

// ---------------- 主题跟随（MainLayout 挂 <html data-theme>） ----------------
const themeObserver = new MutationObserver(() => {
  isDark.value = document.documentElement.getAttribute('data-theme') === 'dark'
})

// ---------------- 列表 / 打开 ----------------
const loadList = async () => {
  try {
    const res: any = await workflowAPI.getAll()
    workflows.value = ((res?.data || []) as any[]).map(toWorkflowFile)
  } catch {
    workflows.value = []
  }
  return workflows.value
}

const toIdeFile = (wf: WorkflowFile): IdeFile => ({
  id: wf.id!,
  label: `${wf.name}.workflow.json`,
  fullPath: `工作流/${wf.name}.workflow.json`,
  language: 'json',
  isWorkflowFile: true,
  content: serializeWorkflow(wf),
  viewMode: 'visual',
  dirty: false,
  loadedAt: Date.now()
})

const openWorkflow = async (wf: WorkflowFile) => {
  let full = workflows.value.find(w => w.id === wf.id)
  if (!full) {
    const res: any = await workflowAPI.getOne(wf.id!)
    full = toWorkflowFile(res.data)
    workflows.value.push(full)
  }
  store.setCurrentWorkflow(full)
  store.selectNode(null)
  ide.openFile(toIdeFile(full))
}

const reopenRecent = async (id: number) => {
  const wf = workflows.value.find(w => w.id === id)
  if (wf) await openWorkflow(wf)
  else await loadList().then(() => reopenRecent(id))
}

// ---------------- 序列化（清理运行期字段，对齐源版 bi()） ----------------
/**
 * 运行期字段白名单剔除。
 * 星域把生成结果也放在 properties 上（会随工作流一起存），这里不跟——
 * analysisResult / outputContent 是无上限的生成文本，一次长分析就能让工作流文件涨几十 KB。
 * 保留的是用户配置（inputImage / inputImageName 是用户导入的图，必须持久化）。
 */
const RUNTIME_ONLY_CONFIG_KEYS = [
  'analysisResult',
  'analysisStatus',
  'analysisComplete',
  'outputContent',
  'receivedData',
  'savedImages',
  'savedVideos',
  'generatedStoppedAt'
]
const stripRuntimeKeys = (cfg: any) => {
  if (!cfg || typeof cfg !== 'object') return {}
  const out: Record<string, any> = {}
  for (const k of Object.keys(cfg)) {
    if (!RUNTIME_ONLY_CONFIG_KEYS.includes(k)) out[k] = cfg[k]
  }
  return out
}
const cleanNode = (n: any) => ({
  id: n.id,
  type: n.type,
  position: { x: Math.round(n.position?.x ?? 0), y: Math.round(n.position?.y ?? 0) },
  data: { label: n.data?.label || '', config: stripRuntimeKeys(n.data?.config) }
})

const serializeWorkflow = (wf: WorkflowFile) =>
  JSON.stringify(
    {
      name: wf.name,
      description: wf.description || '',
      config: wf.config && Object.keys(wf.config).length ? wf.config : { globalVariables: {} },
      nodes: (wf.nodes || []).map(cleanNode),
      edges: (wf.edges || []).map((e: any) => ({
        id: e.id,
        source: e.source,
        target: e.target,
        sourceHandle: e.sourceHandle ?? null,
        targetHandle: e.targetHandle ?? null,
        type: e.type || 'default'
      }))
    },
    null,
    2
  )

const parseWorkflowContent = (content: string): Partial<WorkflowFile> => {
  try {
    const raw = JSON.parse(content)
    return {
      name: raw.name,
      description: raw.description,
      config: raw.config,
      nodes: Array.isArray(raw.nodes) ? raw.nodes : [],
      edges: Array.isArray(raw.edges) ? raw.edges : []
    }
  } catch {
    ElMessage.error('工作流 JSON 解析失败，请检查语法')
    return {}
  }
}

// ---------------- 当前文件的画布数据 ----------------
const activeFile = computed<IdeFile | null>(() => ide.activeFile)

const visualNodes = ref<WNode[] | null>(null)
const visualEdges = ref<WEdge[] | null>(null)

const syncVisualFromStore = () => {
  const wf = store.currentWorkflow
  if (!wf) {
    visualNodes.value = null
    visualEdges.value = null
    return
  }
  // 直接引用 store 的数组（不复制）：画布、store、执行器共用同一批节点对象
  visualNodes.value = wf.nodes || []
  visualEdges.value = wf.edges || []
}

watch(
  activeFile,
  async f => {
    if (!f) {
      visualNodes.value = null
      visualEdges.value = null
      return
    }
    if (store.currentWorkflow?.id !== f.id) {
      const res: any = await workflowAPI.getOne(f.id)
      const wf = toWorkflowFile(res.data)
      store.setCurrentWorkflow(wf)
      if (!workflows.value.find(w => w.id === wf.id)) workflows.value.push(wf)
    }
    syncVisualFromStore()
    store.selectNode(null)
    drawerVisible.value = false
    await nextTick()
    canvasRef.value?.fitView?.({ padding: 0.25 })
  },
  { immediate: true }
)

// ---------------- 保存 ----------------
const saveActive = async () => {
  const f = activeFile.value
  if (!f) {
    ElMessage.warning('没有打开的文件')
    return
  }
  try {
    if (f.viewMode === 'visual' && store.currentWorkflow) {
      // 可视化模式：画布数据同步回 JSON
      f.content = serializeWorkflow(store.currentWorkflow)
    }
    const parsed = parseWorkflowContent(f.content)
    if (!parsed.nodes) return
    await workflowAPI.update(f.id, {
      name: parsed.name,
      description: parsed.description,
      config: JSON.stringify(parsed.config || {}),
      graph: JSON.stringify({ nodes: parsed.nodes, edges: parsed.edges })
    })
    if (store.currentWorkflow?.id === f.id) {
      store.currentWorkflow.name = parsed.name || store.currentWorkflow.name
      store.currentWorkflow.description = parsed.description || ''
      store.currentWorkflow.config = parsed.config || {}
      // ⚠ 可视化模式下画布节点就是 store 里的同一批对象，这里绝不能用解析出的副本替换数组，
      // 否则执行器 updateNodeStatus 改的是副本，画布上的状态色不再变化。
      if (f.viewMode !== 'visual') {
        store.currentWorkflow.nodes = parsed.nodes as WNode[]
        store.currentWorkflow.edges = parsed.edges as WEdge[]
      }
    }
    const inList = workflows.value.find(w => w.id === f.id)
    if (inList) {
      inList.name = parsed.name || inList.name
      inList.nodes = parsed.nodes as WNode[]
      inList.edges = parsed.edges as WEdge[]
      inList.config = parsed.config || {}
    }
    f.label = `${parsed.name || '未命名'}.workflow.json`
    ide.markSaved(f.id)
    ElMessage.success({ message: `已保存: ${f.label}`, duration: 1200, showClose: false })
  } catch {
    /* 拦截器已提示 */
  }
}

/** 源码 ⇄ 可视化切换：切换前先把另一侧的数据同步过来，避免丢改动 */
const switchViewMode = async (mode: 'editor' | 'visual') => {
  const f = activeFile.value
  if (!f || f.viewMode === mode) return
  if (mode === 'visual') {
    const parsed = parseWorkflowContent(f.content)
    if (parsed.nodes && store.currentWorkflow) {
      store.currentWorkflow.nodes = parsed.nodes as WNode[]
      store.currentWorkflow.edges = (parsed.edges || []) as WEdge[]
      store.currentWorkflow.config = parsed.config || {}
      if (parsed.name) store.currentWorkflow.name = parsed.name
    }
    syncVisualFromStore()
    ide.setViewMode(f.id, 'visual')
    await nextTick()
    canvasRef.value?.fitView?.({ padding: 0.25 })
  } else {
    if (store.currentWorkflow) f.content = serializeWorkflow(store.currentWorkflow)
    ide.setViewMode(f.id, 'editor')
  }
}

const onMonacoChange = () => {
  if (activeFile.value) ide.touchFile(activeFile.value.id)
}

// ---------------- 可视化画布联动 ----------------
let updateTimer: number | null = null
const scheduleContentSync = () => {
  if (updateTimer) window.clearTimeout(updateTimer)
  updateTimer = window.setTimeout(() => {
    updateTimer = null
    const f = activeFile.value
    if (f && store.currentWorkflow) f.content = serializeWorkflow(store.currentWorkflow)
  }, 300)
}

const onNodesUpdate = (nodes: WNode[]) => {
  if (store.currentWorkflow) store.currentWorkflow.nodes = nodes
  scheduleContentSync()
}
const onEdgesUpdate = (edges: WEdge[]) => {
  if (store.currentWorkflow) store.currentWorkflow.edges = edges
  scheduleContentSync()
}

const selectedNode = computed<WNode | null>(() => {
  if (!store.currentWorkflow) return null
  const id = store.selectedNodeId
  if (!id) return null
  return store.currentWorkflow.nodes.find(n => n.id === id) || null
})

const onNodeClick = (node: WNode) => {
  store.selectNode(node.id)
  drawerVisible.value = true
}

const onNodeConfigChange = () => {
  scheduleContentSync()
}

const deleteSelectedNode = () => {
  const node = selectedNode.value
  if (!node || !canvasRef.value) return
  canvasRef.value.removeNode(node.id)
  store.selectNode(null)
  drawerVisible.value = false
}

const addNode = (type: NodeType) => {
  const canvas = canvasRef.value
  if (!canvas) return
  const seq = canvas.nodes.length + 1
  const id = `${type}-${Date.now()}`
  canvas.addNode({
    id,
    type,
    position: { x: 120 + seq * 30, y: 120 + (seq % 5) * 40 },
    data: createNodeData(type)
  })
  store.selectNode(id)
  drawerVisible.value = true
}

/** 节点库拖入：画布已按落点建好节点，这里只负责选中并打开属性面板 */
const onNodeAdded = (id: string) => {
  store.selectNode(id)
  drawerVisible.value = true
}

// ---------------- 变量候选 ----------------
const knownVariables = computed(() => {
  const set = new Set<string>(['lastOutput', 'lastAIOutput', 'aiResult', 'merged', 'text', 'textResult', 'currentIteration', 'currentItem', 'loopResults', 'chapterHistory', 'previousChapters'])
  const wf = store.currentWorkflow
  if (wf) {
    wf.nodes.forEach(n => {
      const cfg = n.data?.config || {}
      if (n.type === 'start' && cfg.initialVariables) Object.keys(cfg.initialVariables).forEach(k => set.add(k))
      if (n.type === 'context' && cfg.variableName) set.add(cfg.variableName)
      if (n.type === 'aiProcess' && cfg.outputVariable) set.add(cfg.outputVariable)
      if (n.type === 'text' && cfg.outputVariable) set.add(cfg.outputVariable)
    })
    if (wf.config?.globalVariables) Object.keys(wf.config.globalVariables).forEach(k => set.add(k))
  }
  const ctx = (store.executionResult as any)?.context
  if (ctx && typeof ctx === 'object') Object.keys(ctx).forEach(k => set.add(k))
  return Array.from(set).filter(Boolean)
})

// ---------------- 运行 / 停止 ----------------
const runWorkflow = async () => {
  const f = activeFile.value
  if (!f) return
  // 断点挂起期间 isExecuting 为 false，这里必须一并拦住，否则会开出第二个执行实例
  if (store.isExecuting || store.isPaused) return
  await saveActive()

  store.startExecution()
  let runId: number | null = null
  let outputDir: string | null = null
  try {
    const res: any = await workflowAPI.createRun(f.id!, store.currentWorkflow?.config?.outputDir)
    runId = res?.data?.id ?? null
    outputDir = res?.data?.output_dir ?? null
  } catch {
    runId = null
  }
  store.currentRunId = runId

  const wf = store.currentWorkflow
  if (!wf) {
    ElMessage.warning('请先打开一个工作流')
    return
  }
  const exec = new WorkflowExecutor(
    { ...wf, nodes: wf.nodes, edges: wf.edges },
    store,
    {
      workflowId: f.id,
      runId,
      outputDir,
      mode: runMode.value,
      limits: { ...store.concurrencyLimits }
    }
  )
  executor = exec
  try {
    const result = await exec.execute()
    // 保留实例：上下文变量池供「从指定节点重跑」复用
    lastExecutor = exec
    lastExecutorWorkflowId = f.id
    store.completeExecution(result)
    if (runId) {
      await workflowAPI.updateRun(runId, {
        status: result.stopped ? 'stopped' : 'success',
        logs: JSON.stringify(store.executionLogs),
        context: JSON.stringify(result.context || {})
      })
    }
    // 画布节点就是 store 里的同一批对象，无需回灌（回灌会抹掉成功/失败配色）
  } catch (e: any) {
    store.handleExecutionError(e)
    if (runId) {
      try {
        await workflowAPI.updateRun(runId, {
          status: 'error',
          logs: JSON.stringify(store.executionLogs),
          context: '{}',
          error: e?.message || String(e)
        })
      } catch {
        /* 忽略 */
      }
    }
  } finally {
    executor = null
  }
}

const stopRun = () => executor?.stop()
let executor: WorkflowExecutor | null = null
/** 上一次执行实例（保留下来给断点继续 / 重跑 / 单节点测试复用上下文） */
let lastExecutor: WorkflowExecutor | null = null
let lastExecutorWorkflowId: number | null = null

// ---------------- 节点级操作：单节点测试 / 从指定节点重跑 ----------------
/**
 * 这两类操作都需要一个存活的执行器（上下文变量池就是缓存）。
 * 优先复用上一次执行的实例；没有就新建一个不落盘的实例。
 * 注意：一律用顺序引擎 —— 「重跑」的语义就是沿边递归往下走，
 * 并发调度器会把下游交给队列而不是立即执行。
 */
const ensureExecutor = (): WorkflowExecutor | null => {
  const wf = store.currentWorkflow
  const f = activeFile.value
  if (!wf || !f) {
    ElMessage.warning('请先打开一个工作流')
    return null
  }
  if (lastExecutor && lastExecutorWorkflowId === f.id) return lastExecutor
  const exec = new WorkflowExecutor({ ...wf, nodes: wf.nodes, edges: wf.edges }, store, {
    workflowId: f.id,
    runId: null,
    outputDir: null,
    mode: 'sequential'
  })
  lastExecutor = exec
  lastExecutorWorkflowId = f.id
  return exec
}

const onTestRunNode = async (id: string) => {
  if (store.isExecuting) return
  const exec = ensureExecutor()
  if (!exec) return
  try {
    await exec.runSingleNode(id)
  } catch (e: any) {
    store.handleExecutionError(e)
  }
}

const onRerunNode = async (id: string) => {
  if (store.isExecuting) return
  const exec = ensureExecutor()
  if (!exec) return
  try {
    await exec.rerunFromNode(id)
  } catch (e: any) {
    store.handleExecutionError(e)
  }
}

/** 断点挂起面板：继续 */
const onBreakpointContinue = () => {
  const exec = executor || lastExecutor
  if (!exec) {
    store.resumeExecution()
    return
  }
  exec.resume()
}

/** 断点挂起面板：停止 */
const onBreakpointStop = () => {
  ;(executor || lastExecutor)?.stop()
}

// ---------------- 新建 / 重命名 / 删除 ----------------
/** 新建即空白画布：不预置节点，从节点库拖入（与列表页口径一致） */
const buildNewWorkflowContent = (name: string) => {
  const wf: WorkflowFile = {
    name,
    description: '',
    config: { globalVariables: {} },
    nodes: [],
    edges: []
  }
  return wf
}

const createWorkflow = async () => {
  const res: any = await workflowAPI.create({
    name: '未命名工作流',
    description: '',
    graph: JSON.stringify({
      nodes: buildNewWorkflowContent('').nodes,
      edges: buildNewWorkflowContent('').edges
    }),
    config: JSON.stringify({ globalVariables: {} })
  })
  const wf = toWorkflowFile(res.data)
  workflows.value.unshift(wf)
  await openWorkflow(wf)
  ElMessage.success(`工作流 ${wf.name}.workflow.json 创建成功`)
}

const createNovelWorkflow = async () => {
  const graph = {
    nodes: [
      {
        id: 'start-1',
        type: 'start',
        position: { x: 60, y: 180 },
        data: { label: '开始', status: 'idle', result: null, error: null, config: { initialVariables: { chapterPlans: [] } }, executionHistory: [] }
      },
      {
        id: 'loop-1',
        type: 'loop',
        position: { x: 380, y: 180 },
        data: { label: '逐章循环', status: 'idle', result: null, error: null, config: { loopType: 'array', arraySource: { variableName: 'chapterPlans' }, historyWindow: 3, keepContext: true }, executionHistory: [] }
      },
      {
        id: 'ai-1',
        type: 'aiProcess',
        position: { x: 720, y: 60 },
        data: { label: '写本章', status: 'idle', result: null, error: null, config: { modelId: '', userPrompt: '写第 {{currentIteration}} 章，本章计划：{{currentItem}}\n\n前文：\n{{previousChapters}}', contextVariables: [{ variableName: 'previousChapters' }], contextMaxItems: 3, maxTokens: 4000 }, executionHistory: [] }
      },
      {
        id: 'ctx-1',
        type: 'context',
        position: { x: 720, y: 330 },
        data: { label: '攒章节历史', status: 'idle', result: null, error: null, config: { action: 'append', variableName: 'chapterHistory', inputVariables: [{ variableName: 'aiResult' }], includeTitle: true, maxHistory: 20 }, executionHistory: [] }
      },
      {
        id: 'end-1',
        type: 'end',
        position: { x: 380, y: 420 },
        data: { label: '结束', status: 'idle', result: null, error: null, config: { includeOutputs: ['chapterHistory'], outputFormat: 'md', separator: '\n\n---\n\n' }, executionHistory: [] }
      }
    ],
    edges: [
      { id: 'e1', source: 'start-1', target: 'loop-1', type: 'default' },
      { id: 'e2', source: 'loop-1', target: 'ai-1', sourceHandle: 'loop-body', type: 'default' },
      { id: 'e3', source: 'ai-1', target: 'ctx-1', type: 'default' },
      { id: 'e4', source: 'ctx-1', target: 'loop-1', type: 'default' },
      { id: 'e5', source: 'loop-1', target: 'end-1', sourceHandle: 'loop-done', type: 'default' }
    ]
  }
  const res: any = await workflowAPI.create({
    name: '章节流水线',
    description: '开始 → 循环逐章 AI 处理 → 攒历史 → 汇总',
    graph: JSON.stringify(graph),
    config: JSON.stringify({ globalVariables: {} })
  })
  const wf = toWorkflowFile(res.data)
  workflows.value.unshift(wf)
  await openWorkflow(wf)
  ElMessage.success(`工作流 ${wf.name}.workflow.json 创建成功`)
}

const startRename = (wf: WorkflowFile) => {
  renameTarget.value = wf
  renameName.value = wf.name
  renameVisible.value = true
}

const submitRename = async () => {
  const name = renameName.value.trim()
  const target = renameTarget.value
  if (!name || !target) return
  await workflowAPI.update(target.id!, { name })
  const inList = workflows.value.find(w => w.id === target.id)
  if (inList) inList.name = name
  const file = ide.openFiles.find(f => f.id === target.id)
  if (file) {
    file.label = `${name}.workflow.json`
    file.fullPath = `工作流/${name}.workflow.json`
  }
  renameVisible.value = false
  ElMessage.success('已重命名')
}

const deleteWorkflow = async (wf: WorkflowFile) => {
  await workflowAPI.remove(wf.id!)
  workflows.value = workflows.value.filter(w => w.id !== wf.id)
  ide.closeFile(wf.id!)
  ide.removeRecent(wf.id!)
  if (store.currentWorkflow?.id === wf.id) store.setCurrentWorkflow(null)
  ElMessage.success('已删除')
}

// ---------------- 标签页横滚 / 侧栏拖宽 ----------------
const onTabsWheel = (e: WheelEvent) => {
  if (!tabsListRef.value) return
  tabsListRef.value.scrollLeft += e.deltaY || e.deltaX
}

let resizing: 'left' | null = null
const startResize = (side: 'left', e: MouseEvent) => {
  resizing = side
  e.preventDefault()
}
const onMove = (e: MouseEvent) => {
  if (resizing !== 'left') return
  const w = Math.min(Math.max(e.clientX, 180), 480)
  ide.leftSidebarWidth = w
}
const onUp = () => {
  resizing = null
}

// ---------------- 底部统计（源版口径：中日韩字符 + 字母数字，300 字/分钟） ----------------
const countChars = (text?: string) => {
  if (!text) return 0
  let n = 0
  const cjk = text.match(/[\u4e00-\u9fa5]/g)
  if (cjk) n += cjk.length
  const alnum = text.match(/[a-zA-Z0-9]/g)
  if (alnum) n += alnum.length
  return n
}

const footerStats = computed(() => {
  const f = activeFile.value
  if (!f) return { wordCount: '0 字数', paragraphs: '0 段落', readTime: '约0分钟' }
  const content = f.content || ''
  const chars = countChars(content)
  const paras = content.split(/\n\s*\n/).filter(s => s.trim()).length
  const minutes = Math.ceil(chars / 300)
  return {
    wordCount: `${chars.toLocaleString()} 字数`,
    paragraphs: `${paras} 段落`,
    readTime: `约${minutes}分钟`
  }
})

const formatRecentTime = (ts: number) => {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const resetEditorSettings = () => {
  ide.editorFontSize = 14
  ide.editorFontFamily = "Consolas, 'Courier New', monospace"
}

const applyGenerated = async (content: string) => {
  const f = activeFile.value
  if (!f) {
    ElMessage.warning('请先打开一个工作流文件')
    return
  }
  f.content = content
  ide.setViewMode(f.id, 'editor')
  await saveActive()
}

// ---------------- 生命周期 ----------------
const onKeydown = (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault()
    saveActive()
  }
}

onMounted(async () => {
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('mousemove', onMove)
  window.addEventListener('mouseup', onUp)
  await loadList()
  // 默认打开最近一个
  const first = workflows.value[0]
  if (first && !ide.activeFileId) await openWorkflow(first)
})

onBeforeUnmount(() => {
  themeObserver.disconnect()
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('mousemove', onMove)
  window.removeEventListener('mouseup', onUp)
  executor?.stop()
})
</script>

<style scoped>
@import './styles/ide.css';

/* 源版直接用 EP 明暗变量；星芒只引了 EP 亮色变量，这里把表面/文字色映射到本站主题 */
.wf-ide {
  --wf-el-bg: #ffffff;
  --wf-el-bg-page: #f2f3f5;
  --wf-el-bg-overlay: #ffffff;
  --wf-el-border: #dcdfe6;
  --wf-el-border-light: #ebeef5;
  --wf-el-text-primary: #303133;
  --wf-el-text-regular: #606266;
  --wf-el-text-secondary: #909399;
  --wf-el-text-placeholder: #c0c4cc;
  --wf-el-fill-lighter: #fafafa;
  --wf-el-fill-light: #f5f7fa;
  --wf-el-fill: #f0f2f5;
  --wf-el-fill-dark: #e0e2e6;
  --wf-el-fill-blank: #ffffff;
  --wf-el-primary-light-9: #ecf5ff;
  --wf-el-primary-light-8: #d9ecff;
  --wf-el-primary-light-5: #a0cfff;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
.wf-ide.wf-dark {
  --wf-el-bg: #2a2a2a;
  --wf-el-bg-page: #1a1a1a;
  --wf-el-bg-overlay: #1d1e1f;
  --wf-el-border: #414243;
  --wf-el-border-light: #363637;
  --wf-el-text-primary: #e5eaf3;
  --wf-el-text-regular: #cfd3dc;
  --wf-el-text-secondary: #a3a6ad;
  --wf-el-text-placeholder: #6c6e72;
  --wf-el-fill-lighter: #262727;
  --wf-el-fill-light: #262727;
  --wf-el-fill: #303030;
  --wf-el-fill-dark: #3a3a3c;
  --wf-el-fill-blank: #1a1a1a;
  --wf-el-primary-light-9: #1f2d3d;
  --wf-el-primary-light-8: #24344d;
  --wf-el-primary-light-5: #3a5a80;
}
/* 并发参数弹窗提示 */
.limits-hint {
  margin-top: 4px;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--wf-el-fill-lighter, #fafafa);
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  font-size: 12px;
  line-height: 1.6;
  color: var(--wf-el-text-secondary, #909399);
}
</style>
