<template>
  <div class="knowledge-graph-panel">
    <div class="kg-layout">
      <!-- 知识图谱主区域 -->
      <div class="kg-content">
        <div class="graph-toolbar">
          <div class="toolbar-row">
            <div class="toolbar-left">
              <span class="graph-title">知识图谱</span>
              <div class="layout-switcher" v-if="graphData.entities.length > 0">
                <el-tooltip content="力导向动态布局" placement="top">
                  <button
                    class="layout-btn"
                    :class="{ active: layoutMode === 'force' }"
                    @click="setLayoutMode('force')"
                  >
                    <el-icon><Aim /></el-icon>
                    <span>动态</span>
                  </button>
                </el-tooltip>
                <el-tooltip content="按类型环形布局" placement="top">
                  <button
                    class="layout-btn"
                    :class="{ active: layoutMode === 'circular' }"
                    @click="setLayoutMode('circular')"
                  >
                    <el-icon><Refresh /></el-icon>
                    <span>环形</span>
                  </button>
                </el-tooltip>
                <el-tooltip content="网格布局" placement="top">
                  <button
                    class="layout-btn"
                    :class="{ active: layoutMode === 'grid' }"
                    @click="setLayoutMode('grid')"
                  >
                    <el-icon><Grid /></el-icon>
                    <span>网格</span>
                  </button>
                </el-tooltip>
                <el-tooltip content="同心圆布局" placement="top">
                  <button
                    class="layout-btn"
                    :class="{ active: layoutMode === 'concentric' }"
                    @click="setLayoutMode('concentric')"
                  >
                    <el-icon><MostlyCloudy /></el-icon>
                    <span>同心圆</span>
                  </button>
                </el-tooltip>
                <el-tooltip content="分层分类布局" placement="top">
                  <button
                    class="layout-btn"
                    :class="{ active: layoutMode === 'static' }"
                    @click="setLayoutMode('static')"
                  >
                    <el-icon><Files /></el-icon>
                    <span>分类</span>
                  </button>
                </el-tooltip>
              </div>
            </div>
            <div class="toolbar-center">
              <span class="scope-label">分析范围</span>
              <div class="scope-chips">
                <button
                  v-for="opt in scopeOptions"
                  :key="opt.value"
                  class="scope-chip"
                  :class="{ active: analyzeScope === opt.value }"
                  @click="handleScopeSelect(opt.value)"
                >
                  {{ opt.label }}
                </button>
              </div>
              <el-select
                v-model="selectedChapterIds"
                multiple
                collapse-tags
                collapse-tags-tooltip
                filterable
                placeholder="选择章节"
                size="small"
                style="width: 160px;"
                :disabled="analyzeScope !== 'chapters'"
              >
                <el-option
                  v-for="chapter in chapters"
                  :key="chapter.id"
                  :label="chapter.title"
                  :value="chapter.id"
                />
              </el-select>
              <span class="chapter-count" v-if="analyzeScope === 'chapters'">已选{{ selectedChapterIds.length }}章</span>
            </div>
            <div class="toolbar-right">
              <el-select
                v-model="analyzeConfigId"
                placeholder="AI 模型"
                size="small"
                style="width: 110px;"
              >
                <el-option
                  v-for="config in apiConfigs"
                  :key="config.id"
                  :label="config.name"
                  :value="config.id"
                />
              </el-select>
              <el-button
                type="primary"
                size="small"
                :loading="analyzing"
                @click="handleAnalyze"
                :disabled="!analyzeConfigId"
              >
                <el-icon><MagicStick /></el-icon>
                分析
              </el-button>
              <el-button size="small" @click="handleClearGraph" :disabled="graphData.entities.length === 0">
                <el-icon><Delete /></el-icon>
              </el-button>
              <el-button size="small" @click="handleRefreshGraph" :disabled="graphData.entities.length === 0">
                <el-icon><Refresh /></el-icon>
              </el-button>
              <el-button size="small" @click="openHistoryDialog">
                <el-icon><Clock /></el-icon>
              </el-button>
            </div>
          </div>
        </div>

        <div class="graph-filter" v-if="graphData.entities.length > 0">
          <div class="filter-chips">
            <span
              class="filter-chip"
              :class="{ active: filterType === '' }"
              @click="filterType = ''"
            >全部</span>
            <span
              v-for="t in entityTypes"
              :key="t.value"
              class="filter-chip"
              :class="{ active: filterType === t.value }"
              :style="filterType === t.value ? { background: typeColorMap[t.value], borderColor: typeColorMap[t.value], color: '#fff' } : {}"
              @click="filterType = t.value"
            >{{ t.label }}</span>
          </div>
        </div>

        <div class="graph-chart-container" ref="chartContainerRef">
          <div v-if="graphData.entities.length === 0 && !analyzing" class="graph-empty">
            <el-icon class="empty-icon"><Connection /></el-icon>
            <p>暂无图谱数据</p>
            <p class="empty-hint">点击「分析」按钮，AI 将从章节内容中提取实体和关系</p>
          </div>
        </div>

        <div v-if="analyzing" class="graph-loading-overlay">
          <div class="graph-loading">
            <el-icon class="loading-icon is-loading"><Loading /></el-icon>
            <p>AI 正在分析文本...</p>
            <p class="loading-hint">提取实体和关系中，请稍候</p>
          </div>
        </div>
      </div>

      <!-- 右侧信息面板 -->
      <div
        v-if="selectedEntity"
        class="graph-detail-panel"
        :class="{ collapsed: detailPanelCollapsed }"
      >
        <div class="detail-collapse-handle" @click="detailPanelCollapsed = !detailPanelCollapsed">
          <el-icon :class="{ 'rotate-180': detailPanelCollapsed }"><ArrowRight /></el-icon>
        </div>
        <div class="detail-content">
          <div class="detail-panel-header">
            <h3>{{ selectedEntity.name }}</h3>
            <el-tag size="small" :type="getEntityTypeTagType(selectedEntity.type)">
              {{ getEntityTypeLabel(selectedEntity.type) }}
            </el-tag>
          </div>

          <div class="detail-section" v-if="selectedEntity.description">
            <h4>描述</h4>
            <p class="detail-desc">{{ selectedEntity.description }}</p>
          </div>

          <div class="detail-section">
            <div class="section-header" @click="incomingCollapsed = !incomingCollapsed">
              <h4>⬇️ 被引用 ({{ incomingRelations.length }})</h4>
              <el-icon class="collapse-icon" :class="{ rotated: incomingCollapsed }"><ArrowDown /></el-icon>
            </div>
            <div class="relation-list" v-show="!incomingCollapsed">
              <div
                v-for="rel in incomingRelations"
                :key="rel.id"
                class="relation-item"
                @click="selectEntityByRelation(rel, 'source')"
              >
                <span class="relation-source">{{ getSourceName(rel) }}</span>
                <el-tag size="small" type="info">{{ rel.relation_type }}</el-tag>
              </div>
              <p class="empty-hint" v-if="incomingRelations.length === 0">暂无入链关系</p>
            </div>
          </div>

          <div class="detail-section">
            <div class="section-header" @click="outgoingCollapsed = !outgoingCollapsed">
              <h4>➡️ 引用他人 ({{ outgoingRelations.length }})</h4>
              <el-icon class="collapse-icon" :class="{ rotated: outgoingCollapsed }"><ArrowDown /></el-icon>
            </div>
            <div class="relation-list" v-show="!outgoingCollapsed">
              <div
                v-for="rel in outgoingRelations"
                :key="rel.id"
                class="relation-item"
                @click="selectEntityByRelation(rel, 'target')"
              >
                <span class="relation-target">{{ getTargetName(rel) }}</span>
                <el-tag size="small" type="info">{{ rel.relation_type }}</el-tag>
              </div>
              <p class="empty-hint" v-if="outgoingRelations.length === 0">暂无出链关系</p>
            </div>
          </div>

          <div class="detail-section">
            <el-button type="primary" plain size="small" style="width: 100%;" @click="openAISupplement">
              <el-icon><MagicStick /></el-icon> AI 补充描述
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 实体编辑弹窗 -->
    <el-dialog
      v-model="entityEditVisible"
      title="编辑实体"
      width="520px"
      align-center
      append-to-body
      destroy-on-close
    >
      <div class="entity-edit-form">
        <div class="edit-field">
          <label class="edit-label">名称</label>
          <el-input v-model="entityEditForm.name" placeholder="实体名称" />
        </div>
        <div class="edit-field">
          <label class="edit-label">类型</label>
          <el-select v-model="entityEditForm.type" style="width: 100%;">
            <el-option
              v-for="t in entityTypes"
              :key="t.value"
              :label="t.label"
              :value="t.value"
            />
          </el-select>
        </div>
        <div class="edit-field">
          <div class="edit-label-row">
            <label class="edit-label">描述</label>
            <el-button size="small" text @click="entityEditMDPreview = !entityEditMDPreview">
              {{ entityEditMDPreview ? '编辑' : '预览' }}
            </el-button>
          </div>
          <el-input
            v-if="!entityEditMDPreview"
            v-model="entityEditForm.description"
            type="textarea"
            :rows="5"
            placeholder="支持 Markdown 格式"
          />
          <div v-else class="md-preview-box">
            <MarkdownRenderer :content="entityEditForm.description || '（空）'" />
          </div>
        </div>
        <div v-if="entityEditForm.relations.length > 0" class="edit-field">
          <label class="edit-label">关联关系</label>
          <div class="edit-relations-list">
            <div
              v-for="(rel, rIdx) in entityEditForm.relations"
              :key="rel.id"
              class="edit-relation-item"
            >
              <span class="rel-direction">{{ rel.direction }}</span>
              <span class="rel-target-name">{{ rel.targetName }}</span>
              <el-input
                v-model="rel.relation_type"
                size="small"
                placeholder="关系类型"
                style="width: 120px;"
              />
              <el-input
                v-model="rel.description"
                size="small"
                placeholder="关系描述"
                style="flex: 1;"
              />
            </div>
          </div>
        </div>
        <div v-else class="edit-no-relations">
          该实体暂无关联系关
        </div>
      </div>
      <template #footer>
        <el-button @click="entityEditVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEntityEdit">保存</el-button>
      </template>
    </el-dialog>

    <!-- AI 补充描述弹窗 -->
    <el-dialog
      v-model="aiSupplementVisible"
      title="AI 补充描述"
      width="560px"
      align-center
      append-to-body
      destroy-on-close
    >
      <div class="ai-supplement-form">
        <div class="edit-field">
          <label class="edit-label">目标实体</label>
          <div class="ai-supplement-entity">{{ selectedEntity?.name || '' }}</div>
        </div>
        <div class="edit-field">
          <label class="edit-label">AI 模型</label>
          <el-select
            v-model="supplementConfigId"
            placeholder="选择 AI 模型"
            size="small"
            style="width: 100%;"
          >
            <el-option
              v-for="config in apiConfigs"
              :key="config.id"
              :label="config.name"
              :value="config.id"
            />
          </el-select>
        </div>
        <div class="edit-field">
          <label class="edit-label">关联章节</label>
          <el-select
            v-model="supplementChapterIds"
            multiple
            collapse-tags
            collapse-tags-tooltip
            filterable
            placeholder="选择关联章节（可选）"
            size="small"
            style="width: 100%;"
          >
            <el-option
              v-for="chapter in chapters"
              :key="chapter.id"
              :label="chapter.title"
              :value="chapter.id"
            />
          </el-select>
        </div>
        <div class="edit-field">
          <label class="edit-label">补充要点（可选）</label>
          <el-input
            v-model="supplementHint"
            type="textarea"
            :rows="3"
            placeholder="例如：补充该角色的背景故事、能力来源、与其他人物的恩怨等"
          />
        </div>
        <div class="ai-supplement-actions">
          <el-button type="primary" :loading="supplementLoading" @click="doAISupplement" :disabled="!selectedEntity || !supplementConfigId">
            <el-icon><MagicStick /></el-icon> 开始生成
          </el-button>
        </div>
        <div v-if="supplementResult" class="ai-supplement-result">
          <label class="edit-label">AI 生成结果</label>
          <div class="supplement-content-box">{{ supplementResult }}</div>
          <div class="supplement-result-actions">
            <el-button size="small" @click="replaceDescription">替换描述</el-button>
            <el-button size="small" @click="appendDescription">追加描述</el-button>
            <el-button size="small" @click="supplementResult = ''">清空</el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="closeAISupplement">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 选择世界书弹窗 -->
    <el-dialog
      v-model="worldBookSelectVisible"
      title="选择世界书"
      width="400px"
      align-center
      append-to-body
      destroy-on-close
    >
      <div class="worldbook-select-list">
        <div
          v-for="wb in worldBookList"
          :key="wb"
          class="worldbook-select-item"
          @click="addToSelectedWorldBook(wb)"
        >
          <el-icon><Collection /></el-icon>
          <span>{{ wb }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="worldBookSelectVisible = false">取消</el-button>
      </template>
    </el-dialog>

    <!-- 图谱版本管理弹窗 -->
    <el-dialog
      v-model="historyVisible"
      title="图谱版本管理"
      width="560px"
      align-center
      append-to-body
      destroy-on-close
    >
      <div class="history-list" v-if="versions.length > 0 || graphData.entities.length > 0">
        <div
          v-if="graphData.entities.length > 0"
          class="history-item current-snapshot"
        >
          <div class="history-item-info">
            <span class="history-item-time">
              当前编辑：{{ versions.find(v => v.id === currentVersionId)?.name || '未保存' }}
            </span>
            <span class="history-item-stats">{{ graphData.entities.length }} 实体 · {{ graphData.relations.length }} 关系</span>
          </div>
          <div class="history-item-actions">
            <el-button size="small" type="primary" link @click="saveAsNewVersion">保存为新版本</el-button>
          </div>
        </div>
        <div
          v-for="ver in versions"
          :key="ver.id"
          class="history-item"
          :class="{ active: ver.id === currentVersionId }"
        >
          <div class="history-item-info">
            <span class="history-item-time">
              {{ ver.name }}
              <el-tag v-if="ver.id === currentVersionId" size="small" type="success" effect="plain" style="margin-left: 6px;">当前</el-tag>
            </span>
            <span class="history-item-stats">
              {{ ver.entity_count }} 实体 · {{ ver.relation_count }} 关系 · {{ formatTime(new Date(ver.created_at).getTime()) }}
            </span>
          </div>
          <div class="history-item-actions">
            <el-button
              v-if="ver.id !== currentVersionId"
              size="small"
              type="primary"
              link
              @click="switchToVersion(ver.id)"
            >切换</el-button>
            <el-button size="small" link @click="renameVersionItem(ver.id)">重命名</el-button>
            <el-button size="small" link @click="exportVersionItem(ver.id)">导出</el-button>
            <el-button size="small" type="danger" link @click="deleteVersionItem(ver.id)">删除</el-button>
          </div>
        </div>
      </div>
      <div v-else class="history-empty">暂无版本记录</div>
      <template #footer>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <div style="display: flex; gap: 8px;">
            <el-button size="small" text @click="importVersionFromFile">
              <el-icon><Upload /></el-icon>
              从 JSON 导入
            </el-button>
            <el-button
              size="small"
              type="danger"
              text
              @click="clearAllVersions"
              :disabled="versions.length === 0"
            >清空全部</el-button>
          </div>
          <el-button @click="historyVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 右键菜单 -->
    <Teleport to="body">
      <div
        v-if="contextMenuVisible"
        class="kg-context-menu"
        :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
        @click.stop
      >
        <div class="context-menu-item" @click="editContextMenuEntity">
          <el-icon><Edit /></el-icon>
          <span>编辑</span>
        </div>
        <div class="context-menu-item" @click="addEntityToWorldBook">
          <el-icon><Plus /></el-icon>
          <span>添加到世界书</span>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, nextTick, computed } from 'vue'
import * as echarts from 'echarts/core'
import { GraphChart } from 'echarts/charts'
import {
  TooltipComponent,
  LegendComponent
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  GraphChart,
  TooltipComponent,
  LegendComponent,
  CanvasRenderer
])
import { knowledgeGraphAPI } from '@/api'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'
import type { GraphEntity, GraphRelation, KnowledgeGraphData, GraphEntityType, ApiConfig, Chapter, GraphVersion } from '@/types'
import {
  MagicStick, Delete, Link, User, Close, Connection, Loading, Edit, Pointer, Refresh, ArrowDown, Collection, Clock, Upload, ArrowRight, Aim, Grid, MostlyCloudy, Files
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useWorldBookStore } from '@/stores/worldbook'
import { createDefaultWorldBookEntry } from '@/types/worldbook'

const props = defineProps<{
  bookId: number
  apiConfigs: ApiConfig[]
  chapters: Chapter[]
}>()

const emit = defineEmits<{
  (e: 'analyze-start'): void
  (e: 'analyze-end', data: KnowledgeGraphData): void
}>()

const chartContainerRef = ref<HTMLElement>()
let chartInstance: echarts.ECharts | null = null

const graphData = ref<KnowledgeGraphData>({ entities: [], relations: [] })
const analyzing = ref(false)
let restoringRef = false
const analyzeConfigId = ref<number>()
const analyzeScope = ref<'chapters' | 'first5' | 'first10' | 'first15'>('chapters')
const selectedChapterIds = ref<number[]>([])
const filterType = ref<string>('')
const selectedEntity = ref<GraphEntity | null>(null)
const incomingCollapsed = ref(true)
const outgoingCollapsed = ref(true)
const detailPanelCollapsed = ref(false)

// 布局模式：force 动态 | circular 环形 | grid 网格 | concentric 同心圆 | static 分类分层
type LayoutMode = 'force' | 'circular' | 'grid' | 'concentric' | 'static'
const layoutMode = ref<LayoutMode>('static')

// 暗色主题检测（与全局 data-theme='dark' 同步）
const isDarkTheme = ref(false)
function detectTheme() {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  const attrDark = root.getAttribute('data-theme') === 'dark'
  const classDark = root.classList.contains('dark')
  // 同时检查 body 和 el-app 容器，提高检测准确度
  const bodyDark = document.body?.classList.contains('dark')
  // 最终兜底：通过实际背景色判断，避免属性检测遗漏
  let bgDark = false
  try {
    const bgStyle = getComputedStyle(document.body).backgroundColor
    // 解析 rgb 值，暗色背景通常亮度较低
    const match = bgStyle.match(/rgb\((\d+),\s*(\d+),\s*(\d+)\)/)
    if (match) {
      const r = parseInt(match[1]), g = parseInt(match[2]), b = parseInt(match[3])
      // 相对亮度公式，< 128 视为暗色背景
      const luminance = (0.299 * r + 0.587 * g + 0.114 * b)
      bgDark = luminance < 140
    }
  } catch (_) { /* ignore */ }
  isDarkTheme.value = attrDark || classDark || bodyDark || bgDark
}
let themeObserver: MutationObserver | null = null

// AI 补充描述
const aiSupplementVisible = ref(false)
const supplementConfigId = ref<number>()
const supplementChapterIds = ref<number[]>([])
const supplementHint = ref('')
const supplementLoading = ref(false)
const supplementResult = ref('')

// 右键菜单
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuEntity = ref<GraphEntity | null>(null)

// 世界书选择
const worldBookSelectVisible = ref(false)
const worldBookList = ref<string[]>([])

// 图谱历史记录
const historyVisible = ref(false)
// 图谱版本（后端 API 管理，替代旧版 sessionStorage 历史快照）
const versions = ref<GraphVersion[]>([])
const currentVersionId = ref<number | null>(null)

async function loadVersions() {
  try {
    const res: any = await knowledgeGraphAPI.getVersions(props.bookId)
    if (res && res.success && Array.isArray(res.data)) {
      versions.value = res.data
      // 默认激活第一个版本（最近一个）
      if (!currentVersionId.value && versions.value.length > 0) {
        currentVersionId.value = versions.value[0].id
      }
    } else {
      versions.value = []
    }
  } catch (e) {
    console.error('加载图谱版本失败:', e)
    versions.value = []
  }
}

async function saveAsNewVersion() {
  if (graphData.value.entities.length === 0) {
    ElMessage.warning('当前图谱为空，无需保存为版本')
    return
  }
  try {
    const { value: name } = await ElMessageBox.prompt(
      '为新版本命名',
      '保存为新版本',
      {
        inputPlaceholder: '例如：第一卷人物关系',
        inputValue: `图谱_${formatTime(Date.now())}`,
        confirmButtonText: '保存',
        cancelButtonText: '取消'
      }
    )
    const trimmed = (name || '').trim()
    if (!trimmed) {
      ElMessage.warning('版本名不能为空')
      return
    }
    // 1) 创建版本记录
    const createRes: any = await knowledgeGraphAPI.createVersion(props.bookId, trimmed)
    if (!createRes || !createRes.success || !createRes.data) {
      ElMessage.error('创建版本失败')
      return
    }
    const newVersion = createRes.data
    // 2) 把当前图谱数据写入该版本
    const importRes: any = await knowledgeGraphAPI.importGraph(props.bookId, {
      ...graphData.value,
      versionId: newVersion.id,
      name: newVersion.name
    } as any)
    if (importRes && importRes.success) {
      ElMessage.success(`已保存为新版本：${newVersion.name}`)
      currentVersionId.value = newVersion.id
      await loadVersions()
    } else {
      ElMessage.error('写入图谱数据失败')
    }
  } catch (e: any) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e?.message || '保存失败')
  }
}

async function switchToVersion(versionId: number) {
  if (versionId === currentVersionId.value) return
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  try {
    historyVisible.value = false
    await nextTick()
    restoringRef = true
    disposeChart()
    const res: any = await knowledgeGraphAPI.getVersionData(versionId)
    if (res && res.success && res.data) {
      graphData.value = res.data
      currentVersionId.value = versionId
      await nextTick()
      renderChart()
      ElMessage.success(`已切换到：${ver.name}`)
    } else {
      ElMessage.error('该版本数据读取失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '切换失败')
  } finally {
    restoringRef = false
  }
}

// 供外部（WikiGraph.vue）调用的版本切换：不做 dialog 关闭/恢复标记副作用
async function loadVersionData(versionId: number): Promise<boolean> {
  if (versionId === currentVersionId.value) return true
  try {
    const res: any = await knowledgeGraphAPI.getVersionData(versionId)
    if (res && res.success && res.data) {
      restoringRef = true
      graphData.value = res.data
      currentVersionId.value = versionId
      await nextTick()
      renderChart()
      restoringRef = false
      return true
    }
    return false
  } catch (e) {
    console.error('加载版本数据失败:', e)
    return false
  } finally {
    restoringRef = false
  }
}

async function renameVersionItem(versionId: number) {
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  try {
    const { value: name } = await ElMessageBox.prompt(
      '重命名图谱版本',
      '重命名',
      {
        inputPlaceholder: '输入新名称',
        inputValue: ver.name,
        confirmButtonText: '保存',
        cancelButtonText: '取消'
      }
    )
    const trimmed = (name || '').trim()
    if (!trimmed || trimmed === ver.name) return
    const res: any = await knowledgeGraphAPI.renameVersion(versionId, trimmed)
    if (res && res.success) {
      ElMessage.success('已重命名')
      await loadVersions()
    } else {
      ElMessage.error('重命名失败')
    }
  } catch (e: any) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e?.message || '操作失败')
  }
}

async function deleteVersionItem(versionId: number) {
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  try {
    await ElMessageBox.confirm(
      `确认删除版本「${ver.name}」？该操作不可恢复。`,
      '删除版本',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    )
    const res: any = await knowledgeGraphAPI.deleteVersion(props.bookId, versionId)
    if (res && res.success) {
      ElMessage.success('已删除')
      if (currentVersionId.value === versionId) {
        currentVersionId.value = null
      }
      await loadVersions()
    } else {
      ElMessage.error('删除失败')
    }
  } catch (e: any) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e?.message || '操作失败')
  }
}

function exportVersionItem(versionId: number) {
  const ver = versions.value.find(v => v.id === versionId)
  if (!ver) return
  const payload = {
    name: ver.name,
    description: ver.description,
    scope: ver.scope,
    bookId: props.bookId,
    exportedAt: Date.now(),
    entities: graphData.value.entities,
    relations: graphData.value.relations
  }
  const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `知识图谱_${ver.name}_${formatTime(Date.now()).replace(/[: ]/g, '_')}.json`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('已导出')
}

async function importVersionFromFile() {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json,application/json'
  input.onchange = async () => {
    const file = input.files?.[0]
    if (!file) return
    try {
      const text = await file.text()
      const data = JSON.parse(text)
      const entities = data.entities || data.data?.entities
      const relations = data.relations || data.data?.relations
      if (!Array.isArray(entities) || entities.length === 0) {
        ElMessage.warning('文件格式不正确，缺少有效的 entities 数组')
        return
      }
      const defaultName = data.name || file.name.replace(/\.json$/i, '')
      const { value: name } = await ElMessageBox.prompt(
        `即将导入 ${entities.length} 个实体、${(relations || []).length} 条关系，请为新版本命名`,
        '导入为新版本',
        {
          inputPlaceholder: '版本名称',
          inputValue: defaultName,
          confirmButtonText: '导入',
          cancelButtonText: '取消'
        }
      )
      const trimmed = (name || '').trim()
      if (!trimmed) {
        ElMessage.warning('版本名不能为空')
        return
      }
      const createRes: any = await knowledgeGraphAPI.createVersion(props.bookId, trimmed)
      if (!createRes || !createRes.success || !createRes.data) {
        ElMessage.error('创建版本失败')
        return
      }
      const newVersion = createRes.data
      const importRes: any = await knowledgeGraphAPI.importGraph(props.bookId, {
        entities,
        relations: relations || [],
        versionId: newVersion.id,
        name: newVersion.name
      } as any)
      if (importRes && importRes.success) {
        ElMessage.success(`已导入：${newVersion.name}`)
        currentVersionId.value = newVersion.id
        await loadVersions()
        // 立即切换到新版本
        await switchToVersion(newVersion.id)
      } else {
        ElMessage.error('写入数据失败')
      }
    } catch (e: any) {
      if (e === 'cancel' || e === 'close') return
      console.error('导入失败:', e)
      ElMessage.error(e?.message || '文件解析失败')
    }
  }
  input.click()
}

function openHistoryDialog() {
  loadVersions()
  historyVisible.value = true
}

async function clearAllVersions() {
  if (versions.value.length === 0) return
  try {
    await ElMessageBox.confirm(
      `确认清空全部 ${versions.value.length} 个版本？该操作不可恢复。`,
      '清空全部版本',
      { type: 'warning', confirmButtonText: '清空', cancelButtonText: '取消' }
    )
    let successCount = 0
    for (const ver of [...versions.value]) {
      try {
        const res: any = await knowledgeGraphAPI.deleteVersion(props.bookId, ver.id)
        if (res && res.success) successCount++
      } catch (e) {
        // 单个失败不影响其他
      }
    }
    currentVersionId.value = null
    await loadVersions()
    ElMessage.success(`已清空 ${successCount} 个版本`)
  } catch (e: any) {
    if (e === 'cancel' || e === 'close') return
    ElMessage.error(e?.message || '清空失败')
  }
}

function formatTime(ts: number): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 实体编辑弹窗
const entityEditVisible = ref(false)
const entityEditTarget = ref<GraphEntity | null>(null)
const entityEditMDPreview = ref(false)
interface EditableRelation {
  id: number
  direction: '→' | '←'
  targetName: string
  relation_type: string
  description: string
}
const entityEditForm = ref<{
  name: string
  type: GraphEntityType
  description: string
  relations: EditableRelation[]
}>({
  name: '',
  type: 'character',
  description: '',
  relations: []
})

const scopeOptions: { value: 'chapters' | 'first5' | 'first10' | 'first15', label: string }[] = [
  { value: 'chapters', label: '指定章节' },
  { value: 'first5', label: '前 5 章' },
  { value: 'first10', label: '前 10 章' },
  { value: 'first15', label: '前 15 章' }
]

const entityTypes = [
  { value: 'character', label: '角色' },
  { value: 'location', label: '地点' },
  { value: 'item', label: '物件' },
  { value: 'faction', label: '势力' },
  { value: 'event', label: '事件' },
  { value: 'skill', label: '功法' },
  { value: 'clue', label: '线索' }
]

const typeColorMap: Record<string, string> = {
  character: '#5470c6',
  location: '#91cc75',
  item: '#fac858',
  faction: '#ee6666',
  event: '#73c0de',
  skill: '#3ba272',
  clue: '#fc8452'
}

const entityRelations = computed(() => {
  if (!selectedEntity.value) return []
  return graphData.value.relations.filter(
    r => r.source_id === selectedEntity.value!.id || r.target_id === selectedEntity.value!.id
  )
})

const incomingRelations = computed(() => {
  if (!selectedEntity.value) return []
  return graphData.value.relations.filter(
    r => r.target_id === selectedEntity.value!.id
  )
})

const outgoingRelations = computed(() => {
  if (!selectedEntity.value) return []
  return graphData.value.relations.filter(
    r => r.source_id === selectedEntity.value!.id
  )
})

const filteredGraphData = computed(() => {
  if (!filterType.value) return graphData.value
  const filteredEntities = graphData.value.entities.filter(e => e.type === filterType.value)
  const entityIds = new Set(filteredEntities.map(e => e.id))
  const filteredRelations = graphData.value.relations.filter(
    r => entityIds.has(r.source_id) && entityIds.has(r.target_id)
  )
  return { entities: filteredEntities, relations: filteredRelations }
})

function getEntityTypeLabel(type: string): string {
  const found = entityTypes.find(t => t.value === type)
  return found ? found.label : type
}

function getEntityTypeTagType(type: string): '' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    character: '',
    location: 'success',
    item: 'warning',
    faction: 'danger',
    event: 'info',
    skill: 'success',
    clue: 'warning'
  }
  return map[type] || 'info'
}

function getRelationTargetName(rel: GraphRelation): string {
  if (!selectedEntity.value) return ''
  if (rel.source_id === selectedEntity.value.id) {
    const target = graphData.value.entities.find(e => e.id === rel.target_id)
    return target ? `→ ${target.name}` : ''
  }
  const source = graphData.value.entities.find(e => e.id === rel.source_id)
  return source ? `← ${source.name}` : ''
}

function getSourceName(relation: GraphRelation): string {
  const entity = graphData.value.entities.find(e => e.id === relation.source_id)
  return entity?.name || '未知'
}

function getTargetName(relation: GraphRelation): string {
  const entity = graphData.value.entities.find(e => e.id === relation.target_id)
  return entity?.name || '未知'
}

function selectEntityByRelation(relation: GraphRelation, direction: 'source' | 'target') {
  const entityId = direction === 'source' ? relation.source_id : relation.target_id
  const entity = graphData.value.entities.find(e => e.id === entityId)
  if (entity) {
    selectedEntity.value = entity
  }
}

function closeContextMenu() {
  contextMenuVisible.value = false
  contextMenuEntity.value = null
}

const worldBookStore = useWorldBookStore()

function addEntityToWorldBook() {
  const entity = contextMenuEntity.value
  if (!entity) return
  
  worldBookList.value = worldBookStore.getBookWorldBookList(String(props.bookId))
  closeContextMenu()

  if (worldBookList.value.length === 0) {
    doAddEntityToWorldBook(entity)
    return
  }
  contextMenuEntity.value = entity
  worldBookSelectVisible.value = true
}

function addToSelectedWorldBook(wbName: string) {
  const entity = contextMenuEntity.value
  if (!entity) { worldBookSelectVisible.value = false; return }
  worldBookSelectVisible.value = false
  contextMenuEntity.value = null
  doAddEntityToWorldBook(entity, wbName)
}

function doAddEntityToWorldBook(entity: GraphEntity, wbName?: string) {
  const entry = createDefaultWorldBookEntry()
  entry.comment = entity.name

  let content = entity.description || ''

  const relations = graphData.value.relations.filter(
    r => r.source_id === entity.id || r.target_id === entity.id
  )
  if (relations.length > 0) {
    if (content) content += '\n\n'
    content += '【关联关系】\n'
    relations.forEach(r => {
      const isSource = r.source_id === entity.id
      const otherId = isSource ? r.target_id : r.source_id
      const otherEntity = graphData.value.entities.find(e => e.id === otherId)
      const otherName = otherEntity?.name || '未知'
      const arrow = isSource ? '→' : '←'
      const relationDesc = r.description ? `（${r.description}）` : ''
      content += `${arrow} ${otherName}：${r.relation_type}${relationDesc}\n`
    })
  }

  entry.content = content
  entry.constant = true

  const typeLabel = getEntityTypeLabel(entity.type)
  if (typeLabel) {
    entry.key = [entity.name, typeLabel]
  } else {
    entry.key = [entity.name]
  }

  worldBookStore.addEntry(entry, String(props.bookId), wbName)
  ElMessage.success(`已将"${entity.name}"添加到${wbName || '世界书'}`)
}

function editContextMenuEntity() {
  const entity = contextMenuEntity.value
  if (!entity) return
  entityEditTarget.value = entity
  
  // 获取该实体的所有关联关系
  const rels = graphData.value.relations.filter(
    r => r.source_id === entity.id || r.target_id === entity.id
  )
  const editableRelations: EditableRelation[] = rels.map(r => {
    const isSource = r.source_id === entity.id
    const otherId = isSource ? r.target_id : r.source_id
    const otherEntity = graphData.value.entities.find(e => e.id === otherId)
    return {
      id: r.id,
      direction: isSource ? '→' as const : '←' as const,
      targetName: otherEntity?.name || '未知',
      relation_type: r.relation_type,
      description: r.description || ''
    }
  })
  
  entityEditForm.value = {
    name: entity.name,
    type: entity.type,
    description: entity.description || '',
    relations: editableRelations
  }
  entityEditMDPreview.value = false
  entityEditVisible.value = true
  closeContextMenu()
}

async function saveEntityEdit() {
  const entity = entityEditTarget.value
  if (!entity) return
  if (!entityEditForm.value.name.trim()) {
    ElMessage.warning('请输入实体名称')
    return
  }
  try {
    // 更新实体
    await knowledgeGraphAPI.updateEntity(entity.id, {
      name: entityEditForm.value.name.trim(),
      type: entityEditForm.value.type,
      description: entityEditForm.value.description.trim()
    })
    // 更新本地实体数据
    const idx = graphData.value.entities.findIndex(e => e.id === entity.id)
    if (idx >= 0) {
      graphData.value.entities[idx] = {
        ...graphData.value.entities[idx],
        name: entityEditForm.value.name.trim(),
        type: entityEditForm.value.type,
        description: entityEditForm.value.description.trim()
      }
    }
    if (selectedEntity.value?.id === entity.id) {
      selectedEntity.value = {
        ...selectedEntity.value,
        name: entityEditForm.value.name.trim(),
        type: entityEditForm.value.type,
        description: entityEditForm.value.description.trim()
      }
    }
    
    // 更新关联关系
    for (const rel of entityEditForm.value.relations) {
      await knowledgeGraphAPI.updateRelation(rel.id, {
        relation_type: rel.relation_type,
        description: rel.description
      })
      // 同步本地关系数据
      const ridx = graphData.value.relations.findIndex(r => r.id === rel.id)
      if (ridx >= 0) {
        graphData.value.relations[ridx] = {
          ...graphData.value.relations[ridx],
          relation_type: rel.relation_type,
          description: rel.description
        }
      }
    }
    
    entityEditVisible.value = false
    ElMessage.success('实体已更新')
    await nextTick()
    renderChart()
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  }
}

function openAISupplement() {
  if (!selectedEntity.value) return
  supplementConfigId.value = analyzeConfigId.value
  supplementChapterIds.value = []
  supplementHint.value = ''
  supplementResult.value = ''
  aiSupplementVisible.value = true
}

function closeAISupplement() {
  aiSupplementVisible.value = false
  supplementResult.value = ''
}

async function doAISupplement() {
  if (!selectedEntity.value || !supplementConfigId.value) return
  supplementLoading.value = true
  supplementResult.value = ''

  try {
    const res: any = await knowledgeGraphAPI.aiSupplement({
      bookId: props.bookId,
      configId: supplementConfigId.value,
      entityName: selectedEntity.value.name,
      entityType: selectedEntity.value.type,
      entityDescription: selectedEntity.value.description || '',
      hint: supplementHint.value,
      chapterIds: supplementChapterIds.value.length > 0 ? supplementChapterIds.value : undefined
    })

    if (res && res.data && res.data.content) {
      supplementResult.value = res.data.content
    } else if (res && res.content) {
      supplementResult.value = res.content
    } else {
      ElMessage.warning('AI 未返回有效内容')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || 'AI 补充失败')
  } finally {
    supplementLoading.value = false
  }
}

async function replaceDescription() {
  if (!selectedEntity.value || !supplementResult.value) return
  try {
    await knowledgeGraphAPI.updateEntity(selectedEntity.value.id, {
      description: supplementResult.value
    })
    const idx = graphData.value.entities.findIndex(e => e.id === selectedEntity.value!.id)
    if (idx >= 0) {
      graphData.value.entities[idx] = { ...graphData.value.entities[idx], description: supplementResult.value }
    }
    selectedEntity.value = { ...selectedEntity.value, description: supplementResult.value }
    supplementResult.value = ''
    ElMessage.success('描述已替换')
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  }
}

async function appendDescription() {
  if (!selectedEntity.value || !supplementResult.value) return
  const newDesc = (selectedEntity.value.description || '') + '\n\n' + supplementResult.value
  try {
    await knowledgeGraphAPI.updateEntity(selectedEntity.value.id, {
      description: newDesc
    })
    const idx = graphData.value.entities.findIndex(e => e.id === selectedEntity.value!.id)
    if (idx >= 0) {
      graphData.value.entities[idx] = { ...graphData.value.entities[idx], description: newDesc }
    }
    selectedEntity.value = { ...selectedEntity.value, description: newDesc }
    supplementResult.value = ''
    ElMessage.success('描述已追加')
  } catch (e: any) {
    ElMessage.error(e?.message || '更新失败')
  }
}

function calculateNodePositions(entities: GraphEntity[], relations: GraphRelation[]) {
  const positions: { id: number; x: number; y: number }[] = []

  if (entities.length === 0) return positions

  const centerX = 0
  const centerY = 0

  const connectionCount = new Map<number, number>()
  entities.forEach(e => connectionCount.set(e.id, 0))
  relations.forEach(rel => {
    if (connectionCount.has(rel.source_id)) connectionCount.set(rel.source_id, (connectionCount.get(rel.source_id) || 0) + 1)
    if (connectionCount.has(rel.target_id)) connectionCount.set(rel.target_id, (connectionCount.get(rel.target_id) || 0) + 1)
  })

  const characters = entities.filter(e => e.type === 'character').sort((a, b) => (connectionCount.get(b.id) || 0) - (connectionCount.get(a.id) || 0))
  const nonCharacters = entities.filter(e => e.type !== 'character')

  const characterRadius = Math.max(100, characters.length * 25)
  characters.forEach((entity, index) => {
    const angle = (2 * Math.PI / characters.length) * index - Math.PI / 2
    positions.push({
      id: entity.id,
      x: centerX + characterRadius * Math.cos(angle),
      y: centerY + characterRadius * Math.sin(angle)
    })
  })

  if (nonCharacters.length > 0) {
    const typeGroups = new Map<string, GraphEntity[]>()
    nonCharacters.forEach(entity => {
      const type = entity.type || 'other'
      if (!typeGroups.has(type)) typeGroups.set(type, [])
      typeGroups.get(type)!.push(entity)
    })

    const typePriority = ['location', 'faction', 'item', 'event', 'skill', 'clue']
    const sortedTypes = Array.from(typeGroups.keys()).sort((a, b) => {
      return (typePriority.indexOf(a) || 99) - (typePriority.indexOf(b) || 99)
    })

    const baseRadius = characterRadius + 120
    const radiusStep = 110

    let globalAngleOffset = 0
    sortedTypes.forEach((type, typeIndex) => {
      const group = typeGroups.get(type)!
      const radius = baseRadius + (typeIndex * radiusStep)

      const groupAngleSpan = (group.length / nonCharacters.length) * 2 * Math.PI
      const startAngle = globalAngleOffset - Math.PI / 2

      group.sort((a, b) => (connectionCount.get(b.id) || 0) - (connectionCount.get(a.id) || 0))

      group.forEach((entity, index) => {
        const angle = startAngle + (groupAngleSpan / group.length) * (index + 0.5)
        positions.push({
          id: entity.id,
          x: centerX + radius * Math.cos(angle),
          y: centerY + radius * Math.sin(angle)
        })
      })

      globalAngleOffset += groupAngleSpan
    })
  }

  return positions
}

function calculateCircularPositions(entities: GraphEntity[]) {
  const positions: { id: number; x: number; y: number }[] = []
  if (entities.length === 0) return positions
  const r = 280
  const sorted = [...entities].sort((a, b) => a.type.localeCompare(b.type) || a.name.localeCompare(b.name))
  sorted.forEach((entity, i) => {
    const angle = (2 * Math.PI * i) / sorted.length - Math.PI / 2
    positions.push({
      id: entity.id,
      x: r * Math.cos(angle),
      y: r * Math.sin(angle)
    })
  })
  return positions
}

function calculateGridPositions(entities: GraphEntity[]) {
  const positions: { id: number; x: number; y: number }[] = []
  if (entities.length === 0) return positions
  const sorted = [...entities].sort((a, b) => a.type.localeCompare(b.type) || a.name.localeCompare(b.name))
  const cols = Math.ceil(Math.sqrt(sorted.length))
  const cellW = 110
  const cellH = 90
  const totalW = cols * cellW
  const totalH = Math.ceil(sorted.length / cols) * cellH
  sorted.forEach((entity, i) => {
    const col = i % cols
    const row = Math.floor(i / cols)
    positions.push({
      id: entity.id,
      x: col * cellW - totalW / 2 + cellW / 2,
      y: row * cellH - totalH / 2 + cellH / 2
    })
  })
  return positions
}

function calculateConcentricPositions(entities: GraphEntity[], relations: GraphRelation[]) {
  const positions: { id: number; x: number; y: number }[] = []
  if (entities.length === 0) return positions
  const degree = new Map<number, number>()
  entities.forEach(e => degree.set(e.id, 0))
  relations.forEach(rel => {
    if (degree.has(rel.source_id)) degree.set(rel.source_id, (degree.get(rel.source_id) || 0) + 1)
    if (degree.has(rel.target_id)) degree.set(rel.target_id, (degree.get(rel.target_id) || 0) + 1)
  })
  const sorted = [...entities].sort((a, b) => (degree.get(b.id) || 0) - (degree.get(a.id) || 0))
  const rings: GraphEntity[][] = [[]]
  let curDeg = sorted[0] ? (degree.get(sorted[0].id) || 0) : 0
  for (const n of sorted) {
    const d = degree.get(n.id) || 0
    if (d < curDeg && rings[rings.length - 1].length > 0) {
      rings.push([])
      curDeg = d
    }
    rings[rings.length - 1].push(n)
  }
  if (rings.length === 1 && rings[0].length > 1) {
    const all = rings[0]
    rings[0] = [all[0]]
    rings.push(all.slice(1))
  }
  const maxR = 300
  const ringGap = rings.length > 1 ? maxR / rings.length : maxR
  for (let ri = 0; ri < rings.length; ri++) {
    const ring = rings[ri]
    const r = ri === 0 && ring.length === 1 ? 0 : ringGap * (ri + 0.5)
    for (let i = 0; i < ring.length; i++) {
      const n = ring[i]
      const angle = (2 * Math.PI * i) / Math.max(ring.length, 1) - Math.PI / 2
      positions.push({
        id: n.id,
        x: r * Math.cos(angle),
        y: r * Math.sin(angle)
      })
    }
  }
  return positions
}

function setLayoutMode(mode: LayoutMode) {
  if (layoutMode.value === mode) return
  layoutMode.value = mode
  renderChart()
}

function buildChartOption(data: KnowledgeGraphData) {
  const categories = entityTypes.map((t, i) => ({
    name: t.label,
    itemStyle: { color: typeColorMap[t.value] }
  }))

  // 根据当前布局模式选择位置计算
  let nodePositions: { id: number; x: number; y: number }[] = []
  if (layoutMode.value === 'force') {
    // 动态模式不预设位置，由 ECharts 力导向布局引擎计算
    nodePositions = []
  } else if (layoutMode.value === 'circular') {
    nodePositions = calculateCircularPositions(data.entities)
  } else if (layoutMode.value === 'grid') {
    nodePositions = calculateGridPositions(data.entities)
  } else if (layoutMode.value === 'concentric') {
    nodePositions = calculateConcentricPositions(data.entities, data.relations)
  } else {
    nodePositions = calculateNodePositions(data.entities, data.relations)
  }
  const positionMap = new Map(nodePositions.map(p => [p.id, p]))

  const isForce = layoutMode.value === 'force'

  const nodes = data.entities.map(entity => {
    const pos = positionMap.get(entity.id)
    return {
      id: String(entity.id),
      name: entity.name,
      category: entityTypes.findIndex(t => t.value === entity.type),
      symbolSize: 20,
      // 动态模式不传 x/y，让 ECharts 力引擎自动计算
      ...(pos && !isForce ? { x: pos.x, y: pos.y } : {}),
      itemStyle: {
        color: typeColorMap[entity.type] || '#999'
      },
      label: {
        show: true,
        fontSize: 12,
        color: '#1a1a1a',
        position: 'bottom'
      },
      value: entity.description || ''
    }
  })

  // 计算节点对之间的连线数量，用于设置曲率避免重叠
  const edgePairCount = new Map<string, number>()
  data.relations.forEach(rel => {
    const key = `${Math.min(rel.source_id, rel.target_id)}_${Math.max(rel.source_id, rel.target_id)}`
    edgePairCount.set(key, (edgePairCount.get(key) || 0) + 1)
  })
  const edgePairIndex = new Map<string, number>()

  const links = data.relations.map(rel => {
    const pairKey = `${Math.min(rel.source_id, rel.target_id)}_${Math.max(rel.source_id, rel.target_id)}`
    const totalPairs = edgePairCount.get(pairKey) || 1
    const currentIdx = edgePairIndex.get(pairKey) || 0
    edgePairIndex.set(pairKey, currentIdx + 1)

    // 多条边时使用不同曲率避免重叠
    let curveness = 0
    if (totalPairs > 1) {
      curveness = 0.2 + (currentIdx * 0.25)
    } else {
      curveness = 0.1
    }

    return {
      source: String(rel.source_id),
      target: String(rel.target_id),
      value: 1,
      relationLabel: rel.relation_type,
      lineStyle: {
        width: 1.4,
        color: '#000000',
        curveness
      },
      label: {
        show: true,
        formatter: rel.relation_type,
        fontSize: 10,
        color: '#1a1a1a',
        backgroundColor: 'transparent'
      }
    }
  })

  return {
    tooltip: {
      trigger: 'item',
      confine: true,
      extraCssText: 'max-width: 280px; word-break: break-word; white-space: normal;',
      formatter: (params: any) => {
        if (params.dataType === 'node') {
          const nodeData = params.data
          if (nodeData) {
            const typeLabel = entityTypes[nodeData.category]?.label || '未知'
            let html = `<div style="font-weight:bold;font-size:13px;margin-bottom:4px;word-break:break-word;">${nodeData.name}</div>`
            html += `<div style="color:#666;font-size:11px;">类型: ${typeLabel}</div>`
            if (nodeData.value) {
              html += `<div style="color:#888;font-size:11px;margin-top:4px;word-break:break-word;white-space:pre-wrap;">${nodeData.value}</div>`
            }
            return html
          }
        } else if (params.dataType === 'edge') {
          const edgeData = params.data
          if (edgeData) {
            const sourceNode = nodes.find(n => n.id === edgeData.source)
            const targetNode = nodes.find(n => n.id === edgeData.target)
            if (sourceNode && targetNode) {
              let html = `<div style="font-size:12px;word-break:break-word;">${sourceNode.name} → ${targetNode.name}</div>`
              html += `<div style="color:#666;font-size:11px;">关系: ${edgeData.relationLabel || ''}</div>`
              return html
            }
          }
        }
        return ''
      },
      backgroundColor: 'rgba(255,255,255,0.95)',
      borderColor: '#e0e0e0',
      borderWidth: 1,
      padding: [8, 12],
      textStyle: {
        color: '#333'
      }
    },
    legend: {
      data: categories.map(c => c.name),
      bottom: 0,
      textStyle: {
        fontSize: 12,
        color: '#1a1a1a'
      },
      itemWidth: 12,
      itemHeight: 12,
      itemGap: 14
    },
    series: [{
      type: 'graph',
      layout: isForce ? 'force' : 'none',
      data: nodes,
      links: links,
      categories: categories,
      roam: true,
      draggable: true,
      zoom: 1.33,
      // 动态模式：ECharts 力导向配置
      ...(isForce ? {
        force: {
          repulsion: 220,
          edgeLength: [60, 140],
          gravity: 0.08,
          friction: 0.35,
          initLayout: 'circular'
        },
        animation: true,
        animationDurationUpdate: 800,
        animationEasingUpdate: 'cubicInOut'
      } : {
        animation: true,
        animationDuration: 600,
        animationEasing: 'cubicInOut'
      }),
      emphasis: {
        focus: 'adjacency',
        itemStyle: {
          borderWidth: 3,
          borderColor: '#409EFF',
          shadowBlur: 12,
          shadowColor: 'rgba(64,158,255,0.3)'
        },
        lineStyle: {
          width: 3,
          color: '#409EFF'
        },
        label: {
          fontSize: 13,
          fontWeight: 'bold'
        }
      },
      edgeSymbol: ['none', 'arrow'],
      edgeSymbolSize: [0, 8],
      label: {
        show: true,
        position: 'bottom',
        fontSize: 12,
        color: '#1a1a1a'
      }
    }]
  }
}

function disposeChart() {
  if (chartInstance) {
    try {
      chartInstance.dispose()
    } catch (e) {
      console.error('Dispose error:', e)
    }
    chartInstance = null
  }
}

function ensureChart() {
  const container = chartContainerRef.value
  if (!container || !document.body.contains(container)) {
    return false
  }

  if (!chartInstance) {
    try {
      chartInstance = echarts.init(container)
      chartInstance.on('click', (params: any) => {
        if (params.dataType === 'node') {
          const entityId = Number(params.data?.id || params.id)
          const entity = graphData.value.entities.find(e => e.id === entityId)
          if (entity) {
            selectedEntity.value = entity
          }
        }
      })
      chartInstance.on('contextmenu', (params: any) => {
        if (params.dataType === 'node') {
          const evt = params.event?.event ?? params.event
          if (evt && evt.preventDefault) evt.preventDefault()
          const entityId = Number(params.data?.id || params.id)
          const entity = graphData.value.entities.find(e => e.id === entityId)
          if (entity) {
            contextMenuEntity.value = entity
            contextMenuX.value = (evt.clientX ?? 0)
            contextMenuY.value = (evt.clientY ?? 0)
            contextMenuVisible.value = true
          }
        }
      })
      // 非节点区域右键/点击关闭菜单
      chartInstance.getZr().on('contextmenu', (params: any) => {
        if (params.event?.preventDefault) params.event.preventDefault()
        closeContextMenu()
      })
      document.addEventListener('click', closeContextMenu)
    } catch (e) {
      console.error('Init chart error:', e)
      return false
    }
  }
  return true
}

function renderChart() {
  if (analyzing.value) return
  if (!ensureChart()) return

  const container = chartContainerRef.value
  if (!container) return

  if (filteredGraphData.value.entities.length === 0) {
    try {
      chartInstance?.clear()
    } catch (e) {
      // 清理失败忽略
    }
    return
  }

  try {
    const option = buildChartOption(filteredGraphData.value)
    chartInstance?.setOption(option, true)

    nextTick(() => {
      if (chartInstance && container && document.body.contains(container)) {
        try {
          chartInstance.resize()
        } catch (e) {
          // resize 失败忽略
        }
      }
    })
  } catch (err) {
    console.error('ECharts render error:', err)
  }
}

async function fetchGraphData() {
  try {
    const res: any = await knowledgeGraphAPI.getGraph(props.bookId)
    
    // 健壮的数据提取
    let responseData: KnowledgeGraphData | null = null
    if (res && (res.success === true || res.success === 'true') && res.data) {
      responseData = res.data
    } else if (res && Array.isArray(res.entities)) {
      responseData = res as KnowledgeGraphData
    }
    
    if (responseData) {
      graphData.value = responseData
      await nextTick()
      renderChart()
    }
  } catch (error) {
    console.error('Failed to fetch graph data:', error)
  }
}

function handleScopeSelect(value: 'chapters' | 'first5' | 'first10' | 'first15') {
  analyzeScope.value = value
  if (value === 'first5') {
    selectedChapterIds.value = props.chapters.slice(0, 5).map(c => c.id)
  } else if (value === 'first10') {
    selectedChapterIds.value = props.chapters.slice(0, 10).map(c => c.id)
  } else if (value === 'first15') {
    selectedChapterIds.value = props.chapters.slice(0, 15).map(c => c.id)
  }
}

async function handleAnalyze() {
  if (!analyzeConfigId.value) return

  let chapterIds: number[] | undefined
  let scope: 'all' | 'chapters' = 'chapters'

  if (analyzeScope.value === 'chapters') {
    if (selectedChapterIds.value.length === 0) return
    chapterIds = selectedChapterIds.value
  } else if (analyzeScope.value === 'first5') {
    chapterIds = props.chapters.slice(0, 5).map(c => c.id)
  } else if (analyzeScope.value === 'first10') {
    chapterIds = props.chapters.slice(0, 10).map(c => c.id)
  } else if (analyzeScope.value === 'first15') {
    chapterIds = props.chapters.slice(0, 15).map(c => c.id)
  }

  if (!chapterIds || chapterIds.length === 0) return

  analyzing.value = true
  selectedEntity.value = null
  emit('analyze-start')

  try {
    let versionId = 0
    try {
      const versionRes = await knowledgeGraphAPI.getVersions(props.bookId)
      if (versionRes.success && versionRes.data && versionRes.data.length > 0) {
        versionId = versionRes.data[0].id
      } else {
        const createRes = await knowledgeGraphAPI.createVersion(props.bookId, '默认图谱')
        if (createRes.success && createRes.data) {
          versionId = createRes.data.id
        }
      }
    } catch (e) {
      // 版本获取失败不阻塞流程
    }

    const res: any = await knowledgeGraphAPI.analyze({
      bookId: props.bookId,
      configId: analyzeConfigId.value,
      versionId,
      scope,
      chapterIds
    })

    let responseEntities: GraphEntity[] = []
    let responseRelations: GraphRelation[] = []

    if (res && res.data && res.data.entities) {
      responseEntities = res.data.entities
      responseRelations = res.data.relations || []
    } else if (res && res.entities) {
      responseEntities = res.entities
      responseRelations = res.relations || []
    } else {
      analyzeConfigId.value = 0
      ElMessage.error('AI 返回数据格式异常，请重试')
      return
    }

    const newCount = responseEntities.length

    const entityMap = new Map<string, GraphEntity>()
    graphData.value.entities.forEach(e => {
      entityMap.set(`${e.name}_${e.type}`, e)
    })
    responseEntities.forEach(e => {
      const key = `${e.name}_${e.type}`
      if (!entityMap.has(key)) entityMap.set(key, e)
    })

    const relationMap = new Map<string, GraphRelation>()
    graphData.value.relations.forEach(r => {
      relationMap.set(`${r.source_id}_${r.target_id}_${r.relation_type}`, r)
    })
    responseRelations.forEach(r => {
      const key = `${r.source_id}_${r.target_id}_${r.relation_type}`
      if (!relationMap.has(key)) relationMap.set(key, r)
    })

    graphData.value = {
      entities: Array.from(entityMap.values()),
      relations: Array.from(relationMap.values())
    }

    emit('analyze-end', graphData.value)
    ElMessage.success(`分析完成！新增 ${newCount} 个实体，${responseRelations.length} 条关系`)
    setTimeout(() => loadVersions(), 300)
  } catch (error: any) {
    console.error('Analyze failed:', error)
    ElMessage.error(error.response?.data?.message || error.message || '分析失败')
  } finally {
    analyzing.value = false
    nextTick(() => renderChart())
  }
}

async function handleClearGraph() {
  try {
    await knowledgeGraphAPI.clearGraph(props.bookId)
    graphData.value = { entities: [], relations: [] }
    selectedEntity.value = null
    if (chartInstance) {
      chartInstance.clear()
    }
    ElMessage.success('已清空')
  } catch (error) {
    console.error('Clear graph failed:', error)
  }
}

async function handleRefreshGraph() {
  try {
    const data = await knowledgeGraphAPI.getGraph(props.bookId)
    if (data && data.data && data.data.entities && data.data.entities.length > 0) {
      graphData.value = data.data
      selectedEntity.value = null
      renderChart()
      ElMessage.success('刷新成功')
    } else {
      ElMessage.warning('暂无图谱数据')
    }
  } catch (error) {
    console.error('Refresh graph failed:', error)
    ElMessage.error('刷新失败')
  }
}

function handleResize() {
  if (chartInstance) {
    try {
      chartInstance.resize()
    } catch (e) {
      console.error('Resize error:', e)
    }
  }
}

watch(filterType, () => {
  renderChart()
})

watch(() => graphData.value.entities.length, (newLen, oldLen) => {
  if (newLen === 0) return
  if (restoringRef) return
  setTimeout(() => {
    if (!analyzing.value) {
      nextTick(() => renderChart())
    }
  }, 300)
})

onMounted(async () => {
  await nextTick()

  loadVersions()

  window.addEventListener('resize', handleResize)

  // 主题检测 + 监听变化
  detectTheme()
  if (typeof MutationObserver !== 'undefined' && document.documentElement) {
    themeObserver = new MutationObserver(() => {
      const wasDark = isDarkTheme.value
      detectTheme()
      if (wasDark !== isDarkTheme.value) {
        renderChart()
      }
    })
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['data-theme', 'class']
    })
  }

  if (props.apiConfigs.length > 0) {
    const defaultConfig = props.apiConfigs.find(c => c.is_default)
    analyzeConfigId.value = defaultConfig?.id || props.apiConfigs[0].id
  }

  // 加载已有的图谱数据
  await fetchGraphData()
})

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
  window.removeEventListener('resize', handleResize)
  themeObserver?.disconnect()
  themeObserver = null
  disposeChart()
})

defineExpose({
  fetchGraphData,
  handleAnalyze,
  saveAsNewVersion,
  loadVersions,
  switchToVersion,
  loadVersionData
})
</script>

<style scoped>
.knowledge-graph-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.kg-layout {
  display: flex;
  height: 100%;
  flex: 1;
  overflow: hidden;
}

.kg-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
  position: relative;
}

.graph-detail-panel {
  width: 320px;
  min-width: 320px;
  max-width: 320px;
  border-left: 1px solid var(--el-border-color-lighter);
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0;
  position: relative;
  transition: width 0.25s ease, min-width 0.25s ease, max-width 0.25s ease;
}

.graph-detail-panel.collapsed {
  width: 48px !important;
  min-width: 48px !important;
  max-width: 48px !important;
}

.graph-detail-panel.collapsed .detail-content {
  opacity: 0;
  pointer-events: none;
}

.detail-collapse-handle {
  position: absolute;
  top: 50%;
  left: -10px;
  transform: translateY(-50%);
  background: #fff;
  border: 1px solid rgba(45, 212, 191, 0.3);
  border-radius: 50%;
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  transition: all 0.2s ease;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
}

.detail-collapse-handle:hover {
  background: #2dd4bf;
  border-color: #2dd4bf;
  color: #fff;
}

.detail-collapse-handle .el-icon {
  font-size: 12px;
  transition: transform 0.25s ease;
}

.detail-collapse-handle .el-icon.rotate-180 {
  transform: rotate(180deg);
}

.detail-panel-header {
  padding: 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}

.detail-panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.detail-content {
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
  transition: opacity 0.2s ease;
}

.detail-section {
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.detail-section h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s;
}

.section-header:hover {
  background-color: var(--el-fill-color-light);
}

.collapse-icon {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  transition: transform 0.3s ease;
}

.collapse-icon.rotated {
  transform: rotate(-90deg);
}

.detail-desc {
  font-size: 13px;
  color: var(--el-text-color-regular);
  line-height: 1.6;
  margin: 0;
}

.relation-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.relation-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.relation-item:hover {
  background: var(--el-color-primary-light-9);
}

.relation-source,
.relation-target {
  font-size: 13px;
  color: var(--el-text-color-primary);
  font-weight: 500;
}

.empty-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  text-align: center;
  padding: 8px 0;
  margin: 0;
}

.graph-toolbar {
  padding: 10px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.toolbar-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-shrink: 0;
}

.toolbar-center {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.graph-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  white-space: nowrap;
}

.layout-switcher {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-left: 12px;
  padding: 2px;
  background: var(--el-fill-color-light, #f5f7fa);
  border: 1px solid var(--el-border-color-lighter, #ebeef5);
  border-radius: 8px;
}

.layout-btn {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 4px 8px;
  background: transparent;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  color: var(--el-text-color-regular, #606266);
  transition: all 0.18s ease;
  white-space: nowrap;
}

.layout-btn:hover {
  color: #2dd4bf;
  background: rgba(45, 212, 191, 0.08);
}

.layout-btn.active {
  background: #2dd4bf;
  color: #fff;
  font-weight: 500;
  box-shadow: 0 2px 4px rgba(45, 212, 191, 0.25);
}

.layout-btn .el-icon {
  font-size: 13px;
}

.scope-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  font-weight: 500;
  white-space: nowrap;
}

.scope-chips {
  display: flex;
  gap: 6px;
}

.scope-chip {
  padding: 4px 10px;
  border-radius: 12px;
  border: 1px solid var(--el-border-color);
  background: var(--el-bg-color);
  color: var(--el-text-color-regular);
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  white-space: nowrap;
}

.scope-chip:hover {
  border-color: var(--el-color-primary-light-3);
  color: var(--el-color-primary);
}

.scope-chip.active {
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: #fff;
}

.chapter-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.graph-stats {
  display: flex;
  gap: 16px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.graph-filter {
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.filter-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  font-size: 12px;
  border-radius: 16px;
  border: 1px solid var(--el-border-color);
  background: var(--el-fill-color-blank);
  color: var(--el-text-color-regular);
  cursor: pointer;
  transition: all 0.2s ease;
  user-select: none;
}

.filter-chip:hover {
  border-color: var(--el-color-primary-light-3);
  color: var(--el-color-primary);
}

.filter-chip.active {
  background: var(--el-color-primary);
  border-color: var(--el-color-primary);
  color: #fff;
}

.graph-chart-container {
  flex: 1;
  min-height: 680px;
  position: relative;
  width: 100%;
}

.graph-empty,
.graph-loading-overlay {
  position: absolute;
  inset: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.85);
}

.graph-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--el-text-color-secondary);
}

.empty-icon,
.loading-icon {
  font-size: 48px;
  margin-bottom: 12px;
  color: var(--el-color-info-light-5);
}

.empty-hint,
.loading-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
  margin-top: 4px;
}

.graph-detail {
  border-top: 1px solid var(--el-border-color-lighter);
  padding: 12px 16px;
  max-height: 200px;
  overflow-y: auto;
  flex-shrink: 0;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.detail-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.detail-desc {
  font-size: 12px;
  color: var(--el-text-color-regular);
  line-height: 1.6;
  margin-bottom: 8px;
}

.detail-section-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 6px;
}

.detail-relation-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  font-size: 12px;
}

.rel-target {
  color: var(--el-text-color-primary);
  font-weight: 500;
}

.rel-desc {
  color: var(--el-text-color-secondary);
}

.entity-edit-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.edit-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.edit-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.edit-label-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.md-preview-box {
  min-height: 100px;
  padding: 12px;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  background: var(--el-fill-color-lighter);
  overflow-y: auto;
  max-height: 260px;
}

.edit-relations-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 200px;
  overflow-y: auto;
}

.edit-relation-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 0;
}

.rel-direction {
  font-weight: bold;
  color: var(--el-color-primary);
  flex-shrink: 0;
}

.rel-target-name {
  font-weight: 500;
  white-space: nowrap;
  flex-shrink: 0;
  min-width: 60px;
}

.edit-no-relations {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  padding: 12px 0;
}

.ai-supplement-entity {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  padding: 8px 12px;
  background: var(--el-fill-color-lighter);
  border-radius: 6px;
}

.ai-supplement-actions {
  margin-top: 16px;
  display: flex;
  justify-content: center;
}

.ai-supplement-result {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.supplement-content-box {
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  padding: 14px;
  margin-top: 8px;
  font-size: 13px;
  line-height: 1.7;
  color: var(--el-text-color-regular);
  max-height: 260px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
}

.supplement-result-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.worldbook-select-list {
  max-height: 300px;
  overflow-y: auto;
}

.worldbook-select-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--el-text-color-primary);
  transition: background-color 0.2s;
}

.worldbook-select-item:hover {
  background-color: var(--el-fill-color-light);
}

.history-list {
  max-height: 360px;
  overflow-y: auto;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.history-item:last-child {
  border-bottom: none;
}

.current-snapshot {
  background: var(--el-color-success-light-9);
  border-radius: 6px;
  margin-bottom: 8px;
  border-bottom: 2px solid var(--el-color-success);
}

.history-item-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.history-item-time {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.history-item-stats {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.history-item-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.history-empty {
  text-align: center;
  padding: 40px 0;
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

/* 暗色主题适配 */
:root[data-theme='dark'] .graph-toolbar {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .kg-content {
  background: rgba(30, 41, 59, 0.8);
}

:root[data-theme='dark'] .graph-detail-panel {
  background: rgba(30, 41, 59, 0.9);
  border-left-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .detail-collapse-handle {
  background: rgba(30, 41, 59, 0.9);
  border-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .detail-collapse-handle:hover {
  background: #2dd4bf;
  border-color: #2dd4bf;
}

:root[data-theme='dark'] .detail-panel-header {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .detail-section {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .relation-item {
  background: rgba(51, 65, 85, 0.5);
}

:root[data-theme='dark'] .relation-item:hover {
  background: rgba(94, 234, 212, 0.1);
}

:root[data-theme='dark'] .graph-chart-container {
  background: transparent;
}

:root[data-theme='dark'] .scope-chip {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  color: #d1d5db;
}

:root[data-theme='dark'] .scope-chip:hover {
  border-color: rgba(94, 234, 212, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .scope-chip.active {
  background: #00c9a7;
  border-color: #00c9a7;
  color: #fff;
}

:root[data-theme='dark'] .graph-stats {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .graph-filter {
  border-bottom-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .filter-chip {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
  color: #d1d5db;
}

:root[data-theme='dark'] .filter-chip:hover {
  border-color: rgba(94, 234, 212, 0.4);
  color: #5eead4;
}

:root[data-theme='dark'] .filter-chip.active {
  background: #00c9a7;
  border-color: #00c9a7;
  color: #fff;
}

:root[data-theme='dark'] .graph-detail {
  border-top-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .graph-empty,
:root[data-theme='dark'] .graph-loading-overlay {
  background: rgba(15, 23, 42, 0.85);
}

:root[data-theme='dark'] .graph-loading {
  color: #9ca3af;
}

:root[data-theme='dark'] .empty-icon,
:root[data-theme='dark'] .loading-icon {
  color: #6b7280;
}

:root[data-theme='dark'] .empty-hint,
:root[data-theme='dark'] .loading-hint {
  color: #6b7280;
}

:root[data-theme='dark'] .graph-title {
  color: #f3f4f6;
}

:root[data-theme='dark'] .layout-switcher {
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(71, 85, 105, 0.4);
}

:root[data-theme='dark'] .layout-btn {
  color: #d1d5db;
}

:root[data-theme='dark'] .layout-btn:hover {
  color: #5eead4;
  background: rgba(94, 234, 212, 0.12);
}

:root[data-theme='dark'] .layout-btn.active {
  background: #00c9a7;
  color: #fff;
}

:root[data-theme='dark'] .scope-label {
  color: #e5e7eb;
}

:root[data-theme='dark'] .chapter-count {
  color: #9ca3af;
}

:root[data-theme='dark'] .stat-item {
  color: #9ca3af;
}

:root[data-theme='dark'] .detail-name {
  color: #f3f4f6;
}

:root[data-theme='dark'] .detail-desc {
  color: #d1d5db;
}

:root[data-theme='dark'] .detail-section-title {
  color: #9ca3af;
}

:root[data-theme='dark'] .rel-target {
  color: #f3f4f6;
}

:root[data-theme='dark'] .rel-desc {
  color: #9ca3af;
}

:root[data-theme='dark'] .ai-supplement-entity {
  background: rgba(51, 65, 85, 0.6);
  color: #e2e8f0;
}

:root[data-theme='dark'] .supplement-content-box {
  background: rgba(30, 41, 59, 0.8);
  color: #cbd5e1;
}
</style>

<style>
.kg-context-menu {
  position: fixed;
  z-index: 9999;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  padding: 4px 0;
  min-width: 160px;
}

.kg-context-menu .context-menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px;
  color: #333;
  transition: background-color 0.15s;
}

.kg-context-menu .context-menu-item:hover {
  background-color: #f0f2f5;
}

:root[data-theme='dark'] .kg-context-menu {
  background: #1f2937;
  border-color: #374151;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.4);
}

:root[data-theme='dark'] .kg-context-menu .context-menu-item {
  color: #e5e7eb;
}

:root[data-theme='dark'] .kg-context-menu .context-menu-item:hover {
  background-color: #374151;
}
</style>
