<template>
  <div class="wfr-page" :class="{ 'wf-dark': isDark }">
    <!-- 顶栏 -->
    <div class="wfr-topbar">
      <div class="wfr-topbar-left">
        <el-button text class="wfr-back" @click="goBack">
          <i class="fas fa-arrow-left" /> <span>返回列表</span>
        </el-button>
        <el-divider direction="vertical" />
        <i class="fas fa-project-diagram wfr-wf-icon" />
        <span class="wfr-name" :title="store.currentWorkflow?.name">{{ store.currentWorkflow?.name || '加载中...' }}</span>
        <span v-if="saveState" class="wfr-save-state">{{ saveState }}</span>
      </div>
      <div class="wfr-topbar-right">
        <el-button-group class="wfr-mode-group">
          <el-button size="small" :type="viewMode === 'editor' ? 'primary' : ''" @click="switchViewMode('editor')">
            <i class="fas fa-code" /> 代码
          </el-button>
          <el-button size="small" :type="viewMode === 'visual' ? 'primary' : ''" @click="switchViewMode('visual')">
            <i class="fas fa-project-diagram" /> 可视化
          </el-button>
        </el-button-group>
        <template v-if="viewMode === 'editor'">
          <el-button size="small" type="primary" @click="generatorVisible = true">
            <i class="fas fa-magic" /> AI生成工作流
          </el-button>
        </template>
        <template v-else>
          <el-button size="small" :type="paletteVisible ? 'primary' : ''" @click="paletteVisible = !paletteVisible">
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
          <el-button-group>
            <el-tooltip content="顺序引擎：从开始节点沿连线递归执行（默认）" placement="bottom">
              <el-button size="small" :type="runMode === 'sequential' ? 'primary' : ''" @click="runMode = 'sequential'">
                <i class="fas fa-arrow-right" /> 顺序
              </el-button>
            </el-tooltip>
            <el-tooltip content="并发引擎：拓扑就绪队列 + 限流，互不依赖的节点并行执行" placement="bottom">
              <el-button size="small" :type="runMode === 'concurrent' ? 'primary' : ''" @click="runMode = 'concurrent'">
                <i class="fas fa-bolt" /> 并发
              </el-button>
            </el-tooltip>
          </el-button-group>
          <el-tooltip content="并发参数" placement="bottom">
            <el-button size="small" @click="limitsVisible = true"><i class="fas fa-sliders" /></el-button>
          </el-tooltip>
        </template>
        <el-divider direction="vertical" />
        <el-button
          v-if="!store.isExecuting && !store.isPaused"
          size="small"
          type="primary"
          :disabled="!store.currentWorkflow"
          @click="runWorkflow"
        >
          <i class="fas fa-play" /> 运行
        </el-button>
        <el-button v-else size="small" type="danger" plain @click="stopRun"><i class="fas fa-stop" /> 停止</el-button>
        <el-divider direction="vertical" />
        <el-button size="small" :type="logsVisible ? 'primary' : ''" @click="logsVisible = !logsVisible">
          <i class="fas fa-list" /> 日志
        </el-button>
        <el-button size="small" @click="historyVisible = true">
          <i class="fas fa-clock-rotate-left" /> 历史
        </el-button>
        <el-tooltip content="保存 (Ctrl+S)" placement="bottom">
          <el-button size="small" :disabled="!store.currentWorkflow" @click="saveWorkflow">
            <i class="fas fa-save" />
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <!-- 主体 -->
    <div class="wfr-body">
      <div v-if="loading" class="wfr-loading">
        <i class="fas fa-spinner" />
        <span>正在加载工作流...</span>
      </div>

      <template v-else-if="store.currentWorkflow">
        <!-- 代码模式 -->
        <MonacoEditor
          v-if="viewMode === 'editor'"
          :key="`monaco-${workflowId}`"
          v-model="content"
          language="json"
          :dark="isDark"
          :font-size="14"
          font-family="Consolas, 'Courier New', monospace"
          @change="onMonacoChange"
        />
        <!-- 可视化模式 -->
        <div v-else class="wfr-visual">
          <NodePalette v-if="paletteVisible" @add="addNode" />
          <div class="wfr-canvas-wrap">
            <WorkflowCanvas
              ref="canvasRef"
              :initial-nodes="visualNodes || []"
              :initial-edges="visualEdges || []"
              @node-click="onNodeClick"
              @node-added="onNodeAdded"
              @test-run-node="onTestRunNode"
              @rerun-node="onRerunNode"
              @nodes-update="onNodesUpdate"
              @edges-update="onEdgesUpdate"
              @zoom-change="onZoomChange"
            />
            <div class="wfr-zoom-badge">{{ zoomPercent }}%</div>
            <BreakpointPanel
              v-if="store.isPaused"
              :info="store.checkpointInfo"
              @continue="onBreakpointContinue"
              @stop="onBreakpointStop"
            />
          </div>
          <transition name="el-fade-in-linear">
            <LogsPanel v-if="logsVisible" :logs="store.executionLogs" @clear="store.clearExecutionLogs()" />
          </transition>
        </div>
      </template>

      <div v-else class="wfr-loading">
        <i class="fas fa-triangle-exclamation" />
        <span>未找到该工作流</span>
        <el-button size="small" @click="goBack">返回列表</el-button>
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
      @continue-node="onContinueNode"
      @repeat-node="onRepeatNode"
      @stop-node="onStopNode"
      @pause-toggle="onPauseToggle"
    />

    <!-- AI 生成工作流 -->
    <AiGeneratorDialog v-model="generatorVisible" :existing="content" @apply="applyGenerated" />

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
      <div class="wfr-limits-hint">
        AI 节点会真实调用模型接口，并发过高容易触发上游限流，建议保持 AI ≤ 2。循环体内部始终串行执行，不受这里影响。
      </div>
      <template #footer>
        <el-button size="small" @click="limitsVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 执行历史 -->
    <RunHistoryDialog v-model="historyVisible" :workflow-id="workflowId || null" />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
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
import type { NodeType, WEdge, WNode } from '@/types/workflow'
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import '@vue-flow/minimap/dist/style.css'

const route = useRoute()
const router = useRouter()
const store = useWorkflowStore()

const workflowId = computed(() => Number(route.params.id))
const loading = ref(true)
const viewMode = ref<'editor' | 'visual'>('visual')
const content = ref('')
const logsVisible = ref(false) // 需求：进入画布日志默认关闭
const drawerVisible = ref(false)
const generatorVisible = ref(false)
/** 节点库侧栏（可视化模式左侧，可折叠） */
const paletteVisible = ref(true)
/** 执行引擎：顺序（沿边递归，默认） / 并发（拓扑就绪队列） */
const runMode = ref<'sequential' | 'concurrent'>('sequential')
const limitsVisible = ref(false)
const historyVisible = ref(false)
const saveState = ref('')
const isDark = ref(document.documentElement.getAttribute('data-theme') === 'dark')
const zoomPercent = ref(50)

const canvasRef = ref<InstanceType<typeof WorkflowCanvas> | null>(null)

/** ⚠ 本页的暗色是 JS 类（.wf-dark），不是纯 CSS 选择器：
   setup 阶段 <html data-theme> 可能还没挂（独立路由），靠 observer 又只能收到「之后」的变化，
   所以补挂主题后必须手动同步一次，否则首屏恒亮。 */
const syncDarkFromDom = () => {
  isDark.value = document.documentElement.getAttribute('data-theme') === 'dark'
}

const themeObserver = new MutationObserver(syncDarkFromDom)

/* ---------- FullScreenLayout 直达时的主题兜底（同 ExperienceShareDetail 口径） ---------- */
let themeSetByPage = false
const applyStandaloneTheme = () => {
  const root = document.documentElement
  if (root.getAttribute('data-theme')) return
  const saved = localStorage.getItem('main-theme-mode')
  const legacy = localStorage.getItem('main-theme')
  const mode = saved === 'light' || saved === 'dark' || saved === 'system' ? saved : legacy === 'dark' ? 'dark' : 'light'
  const dark = mode === 'system' ? window.matchMedia('(prefers-color-scheme: dark)').matches : mode === 'dark'
  if (dark) {
    root.setAttribute('data-theme', 'dark')
    themeSetByPage = true
  }
}

// ---------------- 加载 ----------------
/** 需求：进入画布初始缩放 50% */
const INITIAL_ZOOM = 0.5

/** ⚠ 必须在 loading=false（画布真正挂载）之后调用，否则 canvasRef 还是 null，fitView 静默失效 */
const applyInitialZoom = async () => {
  await nextTick()
  canvasRef.value?.applyInitialZoom?.(INITIAL_ZOOM)
}

const loadWorkflow = async () => {
  const id = workflowId.value
  if (!Number.isFinite(id) || id <= 0) {
    loading.value = false
    return
  }
  loading.value = true
  let ok = false
  try {
    const res: any = await workflowAPI.getOne(id)
    const wf = toWorkflowFile(res.data)
    store.setCurrentWorkflow(wf)
    store.selectNode(null)
    drawerVisible.value = false
    content.value = serializeWorkflow(wf)
    viewMode.value = 'visual'
    ok = true
  } catch {
    store.setCurrentWorkflow(null)
    ElMessage.error('工作流加载失败')
  } finally {
    loading.value = false
  }
  if (ok) await applyInitialZoom()
}

watch(workflowId, () => {
  loadWorkflow()
})

// ---------------- 序列化（对齐原 IDE） ----------------
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

const serializeWorkflow = (wf: any) =>
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

const parseWorkflowContent = (text: string): Partial<any> => {
  try {
    const raw = JSON.parse(text)
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

// ---------------- 画布数据（与 store 共享同一批对象） ----------------
const visualNodes = ref<WNode[] | null>(null)
const visualEdges = ref<WEdge[] | null>(null)

const syncVisualFromStore = () => {
  const wf = store.currentWorkflow
  visualNodes.value = wf?.nodes || []
  visualEdges.value = wf?.edges || []
}

watch(
  () => store.currentWorkflow?.id,
  () => syncVisualFromStore(),
  { immediate: true }
)

// ---------------- 保存 ----------------
let saveTimer: number | null = null
const flashSaveState = (text: string) => {
  saveState.value = text
  if (saveTimer) window.clearTimeout(saveTimer)
  saveTimer = window.setTimeout(() => {
    saveState.value = ''
  }, 1600)
}

const saveWorkflow = async () => {
  const wf = store.currentWorkflow
  if (!wf) {
    ElMessage.warning('没有打开的工作流')
    return
  }
  try {
    if (viewMode.value === 'visual') content.value = serializeWorkflow(wf)
    const parsed = parseWorkflowContent(content.value)
    if (!parsed.nodes) return
    await workflowAPI.update(wf.id!, {
      name: parsed.name,
      description: parsed.description,
      config: JSON.stringify(parsed.config || {}),
      graph: JSON.stringify({ nodes: parsed.nodes, edges: parsed.edges })
    })
    wf.name = parsed.name || wf.name
    wf.description = parsed.description || ''
    wf.config = parsed.config || {}
    // ⚠ 可视化模式下画布节点就是 store 里的同一批对象，绝不能用解析出的副本替换数组，
    // 否则执行器 updateNodeStatus 改的是副本，画布状态色不再变化。
    if (viewMode.value !== 'visual') {
      wf.nodes = parsed.nodes as WNode[]
      wf.edges = parsed.edges as WEdge[]
      syncVisualFromStore()
    }
    flashSaveState('已保存')
  } catch {
    /* 拦截器已提示 */
  }
}

/** 源码 ⇄ 可视化切换：切换前先同步另一侧数据，避免丢改动 */
const switchViewMode = async (mode: 'editor' | 'visual') => {
  const wf = store.currentWorkflow
  if (!wf || viewMode.value === mode) return
  if (mode === 'visual') {
    const parsed = parseWorkflowContent(content.value)
    if (parsed.nodes) {
      wf.nodes = parsed.nodes as WNode[]
      wf.edges = (parsed.edges || []) as WEdge[]
      wf.config = parsed.config || {}
      if (parsed.name) wf.name = parsed.name
    }
    syncVisualFromStore()
    viewMode.value = 'visual'
    await applyInitialZoom()
  } else {
    content.value = serializeWorkflow(wf)
    viewMode.value = 'editor'
  }
}

const onMonacoChange = () => flashSaveState('未保存')

// ---------------- 画布联动 ----------------
let updateTimer: number | null = null
const scheduleContentSync = () => {
  if (updateTimer) window.clearTimeout(updateTimer)
  updateTimer = window.setTimeout(() => {
    updateTimer = null
    if (store.currentWorkflow) content.value = serializeWorkflow(store.currentWorkflow)
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

const onZoomChange = (zoom: number) => {
  zoomPercent.value = Math.round(zoom * 100)
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
  flashSaveState('未保存')
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
  flashSaveState('未保存')
}

// ---------------- 变量候选 ----------------
const knownVariables = computed(() => {
  const set = new Set<string>([
    'lastOutput', 'lastAIOutput', 'aiResult', 'merged', 'text', 'textResult',
    'currentIteration', 'currentItem', 'loopResults', 'chapterHistory', 'previousChapters'
  ])
  const wf = store.currentWorkflow
  if (wf) {
    wf.nodes.forEach(n => {
      const cfg = n.data?.config || {}
      if (n.type === 'start' && cfg.initialVariables) Object.keys(cfg.initialVariables).forEach(k => set.add(k))
      if (n.type === 'context' && cfg.variableName) set.add(cfg.variableName)
      if (n.type === 'aiProcess' && cfg.outputVariable) set.add(cfg.outputVariable)
      if (n.type === 'text' && cfg.outputVariable) set.add(cfg.outputVariable)
      // 星域节点：textInput / promptNode 执行时以 placeholder 名为输出键（'未设占位符' 除外），
      // AI 模型节点提示词里的 {{占位符名}} 靠它替换 —— 不收集的话占位符插入就无从谈起
      if ((n.type === 'textInput' || n.type === 'promptNode') && cfg.placeholder) {
        const ph = String(cfg.placeholder).trim()
        if (ph && ph !== '未设占位符') set.add(ph)
      }
    })
    if (wf.config?.globalVariables) Object.keys(wf.config.globalVariables).forEach(k => set.add(k))
  }
  const ctx = (store.executionResult as any)?.context
  if (ctx && typeof ctx === 'object') Object.keys(ctx).forEach(k => set.add(k))
  return Array.from(set).filter(Boolean)
})

// ---------------- 运行 / 停止 ----------------
const runWorkflow = async () => {
  const wf = store.currentWorkflow
  // 断点挂起期间 isExecuting 为 false，必须一并拦住，否则会开出第二个执行实例
  if (!wf || store.isExecuting || store.isPaused) return
  await saveWorkflow()

  store.startExecution()
  let runId: number | null = null
  let outputDir: string | null = null
  try {
    const res: any = await workflowAPI.createRun(wf.id!, wf.config?.outputDir)
    runId = res?.data?.id ?? null
    outputDir = res?.data?.output_dir ?? null
  } catch {
    runId = null
  }
  store.currentRunId = runId

  const exec = new WorkflowExecutor({ ...wf, nodes: wf.nodes, edges: wf.edges }, store, {
    workflowId: wf.id,
    runId,
    outputDir,
    mode: runMode.value,
    limits: { ...store.concurrencyLimits }
  })
  executor = exec
  try {
    const result = await exec.execute()
    // 保留实例：上下文变量池供「从指定节点重跑」复用
    lastExecutor = exec
    lastExecutorWorkflowId = wf.id ?? null
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
 * 一律用顺序引擎 —— 「重跑」语义是沿边递归往下走，并发调度器会把下游交给队列而非立即执行。
 */
const ensureExecutor = (): WorkflowExecutor | null => {
  const wf = store.currentWorkflow
  if (!wf) {
    ElMessage.warning('请先打开一个工作流')
    return null
  }
  if (lastExecutor && lastExecutorWorkflowId === wf.id) return lastExecutor
  const exec = new WorkflowExecutor({ ...wf, nodes: wf.nodes, edges: wf.edges }, store, {
    workflowId: wf.id ?? null,
    runId: null,
    outputDir: null,
    mode: 'sequential'
  })
  lastExecutor = exec
  lastExecutorWorkflowId = wf.id ?? null
  return exec
}

const onTestRunNode = async (id: string) => {
  if (store.isExecuting || store.isPaused) return
  const exec = ensureExecutor()
  if (!exec) return
  try {
    await exec.runSingleNode(id)
  } catch (e: any) {
    store.handleExecutionError(e)
  }
}

const onRerunNode = async (id: string) => {
  if (store.isExecuting || store.isPaused) return
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

/* ---------------- 星域 AI 节点「执行控制」按钮 ---------------- */

/** 继续：放行挂起（与断点面板的继续等价） */
const onContinueNode = () => onBreakpointContinue()

/**
 * 重复：执行中 → 从该节点重跑；未执行 → 单独跑该节点。
 * 与星域 handleAIControlRepeat 的分支判断一致。
 */
const onRepeatNode = async (id: string) => {
  if (store.isExecuting) {
    await onRerunNode(id)
    return
  }
  await onTestRunNode(id)
}

/** 停止：只中止该节点本次生成，工作流转入分步暂停（不是整体停止） */
const onStopNode = (id: string) => {
  const exec = executor || lastExecutor
  if (!exec) {
    ElMessage.warning('当前没有正在执行的流程')
    return
  }
  exec.stopNode(id)
}

/** 分步开关：同步 store 断点，卡片红点与挂起面板据此显示 */
const onPauseToggle = (id: string, paused: boolean) => {
  if (!id) return
  store.setBreakpoint(id, paused)
}

/** 断点挂起面板：停止 */
const onBreakpointStop = () => {
  ;(executor || lastExecutor)?.stop()
}

// ---------------- AI 生成 ----------------
const applyGenerated = async (text: string) => {
  content.value = text
  viewMode.value = 'editor'
  await saveWorkflow()
  viewMode.value = 'visual'
  syncVisualFromStore()
  await applyInitialZoom()
}

// ---------------- 返回 / 离开守卫 ----------------
const goBack = () => router.push('/workflow')

onBeforeRouteLeave(async () => {
  // 断点挂起时 isExecuting 为 false，但执行器仍活着并持有挂起的 Promise，必须一起拦
  if (!store.isExecuting && !store.isPaused) return true
  try {
    await ElMessageBox.confirm('工作流正在执行，离开页面将停止运行。确定离开吗？', '提示', {
      confirmButtonText: '停止并离开',
      cancelButtonText: '留在本页',
      type: 'warning'
    })
    executor?.stop()
    lastExecutor?.stop()
    return true
  } catch {
    return false
  }
})

// ---------------- 生命周期 ----------------
const onKeydown = (e: KeyboardEvent) => {
  if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 's') {
    e.preventDefault()
    saveWorkflow()
  }
}

onMounted(() => {
  applyStandaloneTheme()
  syncDarkFromDom()
  themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  window.addEventListener('keydown', onKeydown)
  loadWorkflow()
})

onBeforeUnmount(() => {
  themeObserver.disconnect()
  window.removeEventListener('keydown', onKeydown)
  if (updateTimer) window.clearTimeout(updateTimer)
  if (saveTimer) window.clearTimeout(saveTimer)
  executor?.stop()
  lastExecutor?.stop()
  if (themeSetByPage) {
    document.documentElement.removeAttribute('data-theme')
    themeSetByPage = false
  }
})
</script>

<style scoped>
/* EP 语义色映射到本站主题（同 WorkflowIde 的 --wf-el-* 方案） */
.wfr-page {
  --wf-el-bg: #ffffff;
  --wf-el-bg-page: #f2f3f5;
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
  --wf-el-bg-overlay: #ffffff;
  height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--wf-el-bg-page, #f2f3f5);
}
.wfr-page.wf-dark {
  --wf-el-bg: #2a2a2a;
  --wf-el-bg-page: #1a1a1a;
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
  --wf-el-bg-overlay: #1d1e1f;
}

/* ---------- 顶栏 ---------- */
.wfr-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 0 16px;
  height: 48px;
  flex-shrink: 0;
  background: var(--wf-el-bg, #ffffff);
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
}
.wfr-topbar-left,
.wfr-topbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.wfr-back {
  padding: 0 6px;
}
.wfr-back span {
  font-size: 13px;
}
.wfr-wf-icon {
  color: #ff9800;
  font-size: 15px;
}
.wfr-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--wf-el-text-primary, #303133);
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.wfr-save-state {
  font-size: 12px;
  color: var(--wf-el-text-secondary, #909399);
}
.wfr-mode-group {
  margin-right: 4px;
}

/* ---------- 主体 ---------- */
.wfr-body {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}
.wfr-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--wf-el-text-secondary, #909399);
  font-size: 14px;
}
.wfr-loading i {
  font-size: 32px;
  animation: wfr-rotating 1.5s linear infinite;
}
@keyframes wfr-rotating {
  to {
    transform: rotate(360deg);
  }
}
.wfr-visual {
  flex: 1;
  display: flex;
  min-height: 0;
  overflow: hidden;
}
.wfr-canvas-wrap {
  position: relative;
  flex: 1;
  min-width: 0;
  height: 100%;
  overflow: hidden;
}
/* 当前缩放角标（进入时 50%） */
.wfr-zoom-badge {
  position: absolute;
  left: 12px;
  bottom: 12px;
  padding: 2px 10px;
  border-radius: 10px;
  font-size: 12px;
  font-family: Consolas, Monaco, monospace;
  color: var(--wf-el-text-secondary, #909399);
  background: var(--wf-el-bg, #ffffff);
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  pointer-events: none;
}
/* 并发参数弹窗提示 */
.wfr-limits-hint {
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
