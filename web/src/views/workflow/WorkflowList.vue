<template>
  <div class="wf-list-page">
    <div class="wf-list-inner">
      <div class="wf-list-header">
        <div class="wf-list-title">
          <i class="fas fa-project-diagram" />
          <h2>工作流</h2>
          <span class="wf-count">{{ filtered.length }} 个</span>
        </div>
        <div class="wf-list-actions">
          <el-input v-model="keyword" placeholder="搜索工作流" clearable class="wf-search">
            <template #prefix><i class="fas fa-search" /></template>
          </el-input>
          <el-select v-model="categoryFilter" class="wf-cat-filter" placeholder="全部分类">
            <el-option label="全部分类" value="" />
            <el-option label="未分类" value="__none__" />
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
          <el-button @click="loadList"><i class="fas fa-rotate" /> 刷新</el-button>
          <el-button @click="createNovelWorkflow"><i class="fas fa-graduation-cap" /> 章节流水线模板</el-button>
          <!-- ⚠ 必须显式调用：直接写 @click="createWorkflow" 会把 MouseEvent 当成第一个参数传进去，
               后端按 name 字段反序列化 string 就会报「请求体格式错误」 -->
          <el-button type="primary" @click="openCreateDialog"><i class="fas fa-plus" /> 新建工作流</el-button>
        </div>
      </div>

      <div v-if="loading" class="wf-hint">
        <i class="fas fa-spinner fa-spin" />
        <span>正在加载工作流...</span>
      </div>
      <div v-else-if="!filtered.length" class="wf-hint">
        <i class="fas fa-folder-open" />
        <span>{{ keyword ? '没有匹配的工作流' : '暂无工作流，点击右上角新建' }}</span>
      </div>

      <div v-else class="wf-grid">
        <div v-for="wf in filtered" :key="wf.id" class="wf-card" @click="openWorkflow(wf)">
          <div class="wf-card-top">
            <div class="wf-card-icon"><i class="fas fa-project-diagram" /></div>
            <div class="wf-card-name" :title="wf.name">{{ wf.name }}</div>
            <el-dropdown trigger="click" @command="(cmd: string) => onCardCommand(cmd, wf)">
              <button class="wf-card-more" title="更多操作" @click.stop>
                <i class="fas fa-ellipsis-v" />
              </button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="rename"><i class="fas fa-edit" /> 重命名</el-dropdown-item>
                  <el-dropdown-item command="category"><i class="fas fa-folder-tree" /> 改分类</el-dropdown-item>
                  <el-dropdown-item command="delete" divided>
                    <span class="wf-danger-item"><i class="fas fa-trash" /> 删除</span>
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div class="wf-card-desc" :title="wf.description">{{ wf.description || '暂无描述' }}</div>
          <div class="wf-card-meta">
            <span v-if="wf.category" class="wf-card-cat"><i class="fas fa-folder" /> {{ wf.category }}</span>
            <span><i class="fas fa-circle-nodes" /> {{ wf.nodes.length }} 节点</span>
            <span><i class="fas fa-arrow-right-arrow-left" /> {{ wf.edges.length }} 连线</span>
            <span v-if="wf.run_count != null" class="wf-meta-runs"><i class="fas fa-play" /> {{ wf.run_count }} 次运行</span>
            <span class="wf-card-time">{{ formatTime(wf.updated_at) }}</span>
          </div>
        </div>
      </div>
    </div>

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

    <!-- 新建：先填名称/简介再落库（画布为空，节点从节点库拖入） -->
    <el-dialog v-model="createVisible" title="新建工作流" width="480px" append-to-body @open="resetCreateForm">
      <el-form label-width="70px" @submit.prevent="submitCreate">
        <el-form-item label="名称" required>
          <el-input
            ref="createNameRef"
            v-model="createForm.name"
            placeholder="给工作流起个名字"
            maxlength="60"
            show-word-limit
            @keyup.enter="submitCreate"
          />
        </el-form-item>
        <el-form-item label="简介">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="一句话描述它是干什么的（可选）"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="分类">
          <el-select
            v-model="createForm.category"
            class="wf-cat-input"
            placeholder="选择或输入新分类（可选）"
            filterable
            allow-create
            default-first-option
            clearable
          >
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </el-form>
      <div class="wf-create-hint">创建后进入空白画布，从左侧「节点库」拖入开始 / AI 处理 / 结束等节点搭建流程。</div>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="submitCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 改分类 -->
    <el-dialog v-model="catVisible" title="修改分类" width="420px" append-to-body>
      <el-form label-width="70px" @submit.prevent="submitCategory">
        <el-form-item label="分类">
          <el-select
            v-model="catForm.category"
            class="wf-cat-input"
            placeholder="选择或输入新分类"
            filterable
            allow-create
            default-first-option
            clearable
          >
            <el-option v-for="c in categories" :key="c" :label="c" :value="c" />
          </el-select>
        </el-form-item>
      </el-form>
      <div class="wf-create-hint">留空表示「未分类」。同一分类名的工作流会在列表里归为一组过滤。</div>
      <template #footer>
        <el-button @click="catVisible = false">取消</el-button>
        <el-button type="primary" @click="submitCategory">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import workflowAPI, { toWorkflowFile } from '@/services/workflow/workflowAPI'
import type { WorkflowFile } from '@/types/workflow'

const router = useRouter()

const workflows = ref<WorkflowFile[]>([])
const loading = ref(true)
const keyword = ref('')
/** 分类过滤：'' 全部；'__none__' 未分类；其它为分类名 */
const categoryFilter = ref('')

/** 现有分类（去重、忽略空串） */
const categories = computed(() =>
  Array.from(new Set(workflows.value.map(w => (w.category || '').trim()).filter(Boolean))).sort()
)

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  const cat = categoryFilter.value
  return workflows.value.filter(w => {
    if (cat === '__none__') {
      if ((w.category || '').trim()) return false
    } else if (cat && (w.category || '').trim() !== cat) {
      return false
    }
    if (!kw) return true
    return w.name.toLowerCase().includes(kw) || (w.description || '').toLowerCase().includes(kw)
  })
})

const loadList = async () => {
  loading.value = true
  try {
    const res: any = await workflowAPI.getAll()
    workflows.value = ((res?.data || []) as any[]).map(toWorkflowFile)
  } catch {
    workflows.value = []
  } finally {
    loading.value = false
  }
  return workflows.value
}

const openWorkflow = (wf: WorkflowFile) => {
  router.push(`/workflow/run/${wf.id}`)
}

// ---------------- 新建（先弹窗填名称/简介，画布为空） ----------------
/** 新建工作流不再预置节点：空画布，节点从节点库拖入 */
const blankGraph = {
  nodes: [],
  edges: []
}

const novelGraph = {
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

const createWorkflow = async (name = '未命名工作流', description = '', graph: any = blankGraph, category = '') => {
  // 兜底：本函数若被直接绑到 @click 上，第一个参数会是 MouseEvent，
  // 序列化后 name 变成对象，后端报「cannot unmarshal object into Go struct field Workflow.name」
  if (typeof name !== 'string') name = '未命名工作流'
  if (typeof description !== 'string') description = ''
  if (!graph || typeof graph !== 'object') graph = blankGraph
  if (typeof category !== 'string') category = ''

  creating.value = true
  try {
    const res: any = await workflowAPI.create({
      name,
      description,
      category: category.trim(),
      graph: JSON.stringify(graph),
      config: JSON.stringify({ globalVariables: {} })
    })
    const wf = toWorkflowFile(res.data)
    workflows.value.unshift(wf)
    createVisible.value = false
    ElMessage.success(`工作流「${wf.name}」已创建`)
    await router.push(`/workflow/run/${wf.id}`)
  } finally {
    creating.value = false
  }
}

// ---------------- 新建弹窗 ----------------
const createVisible = ref(false)
const creating = ref(false)
const createNameRef = ref<{ focus: () => void } | null>(null)
const createForm = ref({ name: '', description: '', category: '' })

const resetCreateForm = () => {
  createForm.value = { name: '', description: '', category: '' }
  nextTick(() => createNameRef.value?.focus())
}

const openCreateDialog = () => {
  createVisible.value = true
}

const submitCreate = () => {
  const name = createForm.value.name.trim() || '未命名工作流'
  void createWorkflow(name, createForm.value.description.trim(), blankGraph, createForm.value.category || '')
}

const createNovelWorkflow = () =>
  createWorkflow('章节流水线', '开始 → 循环逐章 AI 处理 → 攒历史 → 汇总', novelGraph)

// ---------------- 重命名 / 删除 ----------------
const renameVisible = ref(false)
const renameTarget = ref<WorkflowFile | null>(null)
const renameName = ref('')

const onCardCommand = (cmd: string, wf: WorkflowFile) => {
  if (cmd === 'rename') {
    renameTarget.value = wf
    renameName.value = wf.name
    renameVisible.value = true
  } else if (cmd === 'category') {
    catTarget.value = wf
    catForm.value = { category: wf.category || '' }
    catVisible.value = true
  } else if (cmd === 'delete') {
    deleteWorkflow(wf)
  }
}

// ---------------- 改分类 ----------------
const catVisible = ref(false)
const catTarget = ref<WorkflowFile | null>(null)
const catForm = ref({ category: '' })

const submitCategory = async () => {
  const target = catTarget.value
  if (!target) return
  const category = (catForm.value.category || '').trim()
  await workflowAPI.update(target.id!, { category })
  const inList = workflows.value.find(w => w.id === target.id)
  if (inList) inList.category = category
  catVisible.value = false
  ElMessage.success(category ? `已归入分类「${category}」` : '已设为未分类')
}

const submitRename = async () => {
  const name = renameName.value.trim()
  const target = renameTarget.value
  if (!name || !target) return
  await workflowAPI.update(target.id!, { name })
  const inList = workflows.value.find(w => w.id === target.id)
  if (inList) inList.name = name
  renameVisible.value = false
  ElMessage.success('已重命名')
}

const deleteWorkflow = async (wf: WorkflowFile) => {
  try {
    await ElMessageBox.confirm(`确定删除工作流「${wf.name}」吗？删除后不可恢复。`, '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
  } catch {
    return
  }
  await workflowAPI.remove(wf.id!)
  workflows.value = workflows.value.filter(w => w.id !== wf.id)
  ElMessage.success('已删除')
}

// ---------------- 展示辅助 ----------------
const formatTime = (ts?: string | number) => {
  if (!ts) return ''
  if (typeof ts === 'string') return ts.replace('T', ' ').slice(0, 16)
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

onMounted(loadList)
</script>

<style scoped>
.wf-list-page {
  /* 与 WorkflowIde/WorkflowRun 同一套 EP 语义色映射（否则 var() 取不到值会回落亮色深字） */
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
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  background: var(--wf-el-bg-page, #f2f3f5);
}
html[data-theme='dark'] .wf-list-page {
  --wf-el-bg: #1d2026;
  --wf-el-bg-page: #16181d;
  --wf-el-border: #3a4048;
  --wf-el-border-light: #2c313a;
  --wf-el-text-primary: #e5eaf3;
  --wf-el-text-regular: #cfd3dc;
  --wf-el-text-secondary: #a3a6ad;
  --wf-el-text-placeholder: #7d828b;
  --wf-el-fill-lighter: #23272e;
  --wf-el-fill-light: #262b33;
  --wf-el-fill: #2c313a;
  --wf-el-fill-dark: #3a4048;
}
.wf-list-inner {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px 24px 40px;
}
.wf-list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
  margin-bottom: 20px;
}
.wf-list-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.wf-list-title i {
  font-size: 20px;
  color: var(--accent, #2563eb);
}
.wf-list-title h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--wf-el-text-primary, #303133);
}
.wf-count {
  font-size: 12px;
  color: var(--wf-el-text-secondary, #909399);
}
.wf-list-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.wf-search {
  width: 220px;
}
/* 分类过滤 / 分类输入 */
.wf-cat-filter {
  width: 150px;
}
.wf-cat-input {
  width: 100%;
}
.wf-card-cat {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 1px 6px;
  border-radius: 3px;
  background: rgba(64, 158, 255, 0.12);
  color: var(--el-color-primary, #409eff);
}

.wf-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 100px 0;
  color: var(--wf-el-text-secondary, #909399);
  font-size: 14px;
}
.wf-hint i {
  font-size: 36px;
  opacity: 0.6;
}
.wf-create-hint {
  margin-top: 4px;
  padding: 8px 10px;
  border-radius: 6px;
  background: var(--wf-el-fill-lighter, #fafafa);
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  font-size: 12px;
  line-height: 1.6;
  color: var(--wf-el-text-secondary, #909399);
}

.wf-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}
.wf-card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px;
  background: var(--wf-el-bg, #ffffff);
  border: 1px solid var(--wf-el-border-light, #ebeef5);
  border-radius: 10px;
  cursor: pointer;
  transition: box-shadow 0.2s, border-color 0.2s, transform 0.2s;
}
.wf-card:hover {
  border-color: var(--accent, #2563eb);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
  transform: translateY(-2px);
}
.wf-card-top {
  display: flex;
  align-items: center;
  gap: 10px;
}
.wf-card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: var(--accent-soft, #eff6ff);
  color: var(--accent, #2563eb);
  font-size: 16px;
  flex-shrink: 0;
}
.wf-card-name {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 500;
  color: var(--wf-el-text-primary, #303133);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.wf-card-more {
  all: unset;
  cursor: pointer;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  color: var(--wf-el-text-secondary, #909399);
  transition: background 0.15s, color 0.15s;
}
.wf-card-more:hover {
  background: var(--wf-el-fill, #f0f2f5);
  color: var(--wf-el-text-primary, #303133);
}
.wf-danger-item {
  color: var(--el-color-danger, #f56c6c);
}
.wf-card-desc {
  font-size: 13px;
  color: var(--wf-el-text-secondary, #909399);
  line-height: 1.5;
  min-height: 20px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.wf-card-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
}
.wf-card-meta i {
  margin-right: 3px;
}
.wf-meta-runs {
  color: var(--wf-el-text-secondary, #909399);
}
.wf-card-time {
  margin-left: auto;
}

/* 暗色下卡片投影加深（颜色已由上面的变量块接管） */
html[data-theme='dark'] .wf-card:hover {
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.45);
}
</style>
