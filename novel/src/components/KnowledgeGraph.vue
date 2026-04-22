<template>
  <div class="knowledge-graph-panel">
    <div class="kg-layout">
      <!-- 左侧版本列表 -->
      <div class="kg-versions">
        <div class="versions-header">
          <span class="versions-title">图谱文件</span>
          <div class="versions-actions">
            <el-button size="small" @click="handleCreateVersion" :disabled="analyzing">
              <el-icon><Plus /></el-icon>
            </el-button>
            <el-button size="small" @click="handleRefreshVersions" :disabled="analyzing">
              <el-icon><Refresh /></el-icon>
            </el-button>
          </div>
        </div>
        <div class="versions-list">
          <div
            v-for="version in versions"
            :key="version.id"
            class="version-item"
            :class="{ active: currentVersionId === version.id }"
            @click="handleSelectVersion(version)"
          >
            <div class="version-main">
              <div class="version-name">{{ version.name }}</div>
              <div class="version-time">{{ formatDate(version.created_at) }}</div>
            </div>
            <div class="version-stats">
              <span class="version-stat">📊 {{ version.entity_count }}实体</span>
              <span class="version-stat">🔗 {{ version.relation_count }}关系</span>
            </div>
            <div class="version-desc" v-if="version.description">{{ version.description }}</div>
            <div class="version-actions">
              <el-button
                size="small"
                type="danger"
                text
                @click.stop="handleDeleteVersion(version)"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
          <div v-if="versions.length === 0 && !analyzing" class="versions-empty">
            <el-icon><Document /></el-icon>
            <p>暂无图谱文件</p>
            <p class="empty-hint">点击上方 + 按钮创建新文件</p>
          </div>
        </div>
      </div>

      <!-- 右侧图谱区域 -->
      <div class="kg-content">
        <div class="graph-header">
          <span class="graph-title">知识图谱</span>
          <div class="graph-header-actions">
            <el-select
              v-model="analyzeConfigId"
              placeholder="AI 模型"
              size="small"
              style="width: 130px;"
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
              :disabled="!analyzeConfigId || !currentVersionId"
            >
              <el-icon><MagicStick /></el-icon>
              分析
            </el-button>
            <el-button size="small" @click="handleClearGraph" :disabled="graphData.entities.length === 0 || !currentVersionId">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
        </div>

        <div class="graph-scope-section">
          <div class="scope-row">
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
          </div>
          <div class="chapter-select-row">
            <el-select
              v-model="selectedChapterIds"
              multiple
              collapse-tags
              collapse-tags-tooltip
              filterable
              placeholder="选择要分析的章节"
              size="small"
              style="flex: 1;"
              :disabled="analyzeScope !== 'chapters'"
            >
              <el-option
                v-for="chapter in chapters"
                :key="chapter.id"
                :label="chapter.title"
                :value="chapter.id"
              />
            </el-select>
            <span class="chapter-count">已选 {{ selectedChapterIds.length }} 章</span>
          </div>
        </div>

        <div class="graph-stats" v-if="graphData.entities.length > 0">
          <span class="stat-item">
            <el-icon><User /></el-icon>
            {{ graphData.entities.length }} 实体
          </span>
          <span class="stat-item">
            <el-icon><Link /></el-icon>
            {{ graphData.relations.length }} 关系
          </span>
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
          <div v-else-if="analyzing" class="graph-loading">
            <el-icon class="loading-icon is-loading"><Loading /></el-icon>
            <p>AI 正在分析文本...</p>
            <p class="loading-hint">提取实体和关系中，请稍候</p>
          </div>
        </div>
      </div>
    </div>

    <!-- 实体详情面板 -->
    <div v-if="selectedEntity" class="entity-detail-panel">
      <div class="entity-detail-header">
        <h3 class="entity-detail-title">{{ selectedEntity.name }}</h3>
        <el-button size="small" text @click="selectedEntity = null">
          <el-icon><Close /></el-icon>
        </el-button>
      </div>
      <div class="entity-detail-content">
        <div class="detail-section">
          <div class="detail-section-title">类型</div>
          <el-tag :type="getEntityTypeTagType(selectedEntity.type)" size="small">
            {{ getEntityTypeLabel(selectedEntity.type) }}
          </el-tag>
        </div>
        <div class="detail-section" v-if="selectedEntity.description">
          <div class="detail-section-title">描述</div>
          <p class="detail-desc">{{ selectedEntity.description }}</p>
        </div>
        <div class="detail-section">
          <div class="detail-section-title">关联关系</div>
          <div class="detail-relations">
            <div
              v-for="rel in entityRelations"
              :key="rel.id"
              class="detail-relation-item"
            >
              <span class="rel-target">{{ getRelationTargetName(rel) }}</span>
              <el-tag size="small" type="info">{{ rel.relation_type }}</el-tag>
              <span class="rel-desc" v-if="rel.description">{{ rel.description }}</span>
            </div>
            <div v-if="entityRelations.length === 0" class="no-relations">
              暂无关联关系
            </div>
          </div>
        </div>
      </div>
    </div>
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
import type { GraphEntity, GraphRelation, KnowledgeGraphData, GraphEntityType, ApiConfig, Chapter, GraphVersion } from '@/types'
import {
  MagicStick, Delete, Link, User, Close, Connection, Loading, Refresh, Document, Plus
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

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
const analyzeConfigId = ref<number>()
const analyzeScope = ref<'chapters' | 'first5' | 'first10' | 'first15'>('chapters')
const selectedChapterIds = ref<number[]>([])
const filterType = ref<string>('')
const selectedEntity = ref<GraphEntity | null>(null)
const versions = ref<GraphVersion[]>([])
const currentVersionId = ref<number>()

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

function buildChartOption(data: KnowledgeGraphData) {
  const categories = entityTypes.map((t, i) => ({
    name: t.label,
    itemStyle: { color: typeColorMap[t.value] }
  }))

  const entityMap = new Map(data.entities.map(e => [e.id, e]))

  const nodes = data.entities.map(entity => ({
    id: String(entity.id),
    name: entity.name,
    category: entityTypes.findIndex(t => t.value === entity.type),
    symbolSize: Math.max(
      30,
      Math.min(60, 30 + data.relations.filter(r => r.source_id === entity.id || r.target_id === entity.id).length * 5)
    ),
    itemStyle: {
      color: typeColorMap[entity.type] || '#999'
    },
    label: {
      show: true,
      fontSize: 11,
      color: '#333'
    },
    value: entity.description || ''
  }))

  const links = data.relations.map(rel => ({
    source: String(rel.source_id),
    target: String(rel.target_id),
    value: 1,
    relationLabel: rel.relation_type,
    lineStyle: {
      width: 1.5,
      color: '#aaa',
      curveness: 0.2
    },
    label: {
      show: true,
      formatter: rel.relation_type,
      fontSize: 9,
      color: '#888'
    }
  }))

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
      textStyle: { fontSize: 10 }
    },
    series: [{
      type: 'graph',
      layout: 'force',
      data: nodes,
      links: links,
      categories: categories,
      roam: true,
      draggable: true,
      force: {
        repulsion: 300,
        gravity: 0.1,
        edgeLength: [80, 200],
        layoutAnimation: true
      },
      emphasis: {
        focus: 'adjacency',
        lineStyle: { width: 3 }
      },
      edgeSymbol: ['none', 'arrow'],
      edgeSymbolSize: [0, 8],
      label: {
        show: true,
        position: 'bottom',
        fontSize: 11
      }
    }]
  }
}

let renderRetryCount = 0
const MAX_RENDER_RETRY = 10
let renderTimer: any = null

function disposeChart() {
  if (chartInstance) {
    try {
      chartInstance.dispose()
    } catch (e) {
      console.error('Dispose error:', e)
    }
    chartInstance = null
  }
  if (renderTimer) {
    clearTimeout(renderTimer)
    renderTimer = null
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
          const entityId = Number(params.id)
          const entity = graphData.value.entities.find(e => e.id === entityId)
          if (entity) {
            selectedEntity.value = entity
          }
        }
      })
    } catch (e) {
      console.error('Init chart error:', e)
      return false
    }
  }
  return true
}

function renderChart() {
  // 清理旧的定时器
  if (renderTimer) {
    clearTimeout(renderTimer)
    renderTimer = null
  }

  // 确保图表实例存在
  if (!ensureChart()) {
    return
  }

  const container = chartContainerRef.value
  if (!container) return

  // 确保容器有正确的高度
  const containerHeight = container.offsetHeight
  if (containerHeight < 100) {
    if (renderRetryCount < MAX_RENDER_RETRY) {
      renderRetryCount++
      renderTimer = setTimeout(() => {
        renderChart()
      }, 100)
    }
    return
  }

  renderRetryCount = 0

  if (filteredGraphData.value.entities.length === 0) {
    try {
      chartInstance?.clear()
    } catch (e) {
      console.error('Clear chart error:', e)
    }
    return
  }

  try {
    const option = buildChartOption(filteredGraphData.value)
    chartInstance?.setOption(option, true)

    // 渲染完成后 resize
    renderTimer = setTimeout(() => {
      if (chartInstance && document.body.contains(container)) {
        try {
          chartInstance.resize()
        } catch (e) {
          console.error('Resize error:', e)
        }
      }
    }, 50)
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
  if (!analyzeConfigId.value || !currentVersionId.value) return
  
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
    const res: any = await knowledgeGraphAPI.analyze({
      bookId: props.bookId,
      configId: analyzeConfigId.value,
      scope,
      chapterIds,
      versionId: currentVersionId.value
    })
    
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
      emit('analyze-end', responseData)
      
      // 刷新版本列表以获取最新统计
      await fetchVersions()
    }
  } catch (error: any) {
    console.error('Analyze failed:', error)
    ElMessage.error(error.response?.data?.message || '分析失败')
  } finally {
    analyzing.value = false
  }
}

async function handleClearGraph() {
  if (!currentVersionId.value) return
  try {
    // 删除当前版本的所有数据
    await knowledgeGraphAPI.deleteVersion(props.bookId, currentVersionId.value)
    graphData.value = { entities: [], relations: [] }
    selectedEntity.value = null
    if (chartInstance) {
      chartInstance.clear()
    }
    // 刷新版本列表
    await fetchVersions()
    ElMessage.success('已清空')
  } catch (error) {
    console.error('Clear graph failed:', error)
  }
}

async function fetchVersions() {
  try {
    const res = await knowledgeGraphAPI.getVersions(props.bookId)
    if (res.success && res.data) {
      versions.value = res.data
    }
  } catch (error) {
    console.error('Failed to fetch versions:', error)
  }
}

function handleRefreshVersions() {
  fetchVersions()
}

async function handleCreateVersion() {
  try {
    const res = await knowledgeGraphAPI.createVersion(props.bookId)
    if (res.success && res.data) {
      versions.value.unshift(res.data)
      currentVersionId.value = res.data.id
      graphData.value = { entities: [], relations: [] }
      if (chartInstance) {
        chartInstance.clear()
      }
      ElMessage.success('已创建新图谱文件')
    }
  } catch (error) {
    console.error('Create version failed:', error)
    ElMessage.error('创建失败')
  }
}

async function handleSelectVersion(version: GraphVersion) {
  if (currentVersionId.value === version.id) return
  
  try {
    const res: any = await knowledgeGraphAPI.getVersionData(version.id)
    
    // 提取数据 - 简化逻辑，成功就尝试显示
    let responseData: KnowledgeGraphData | null = null
    
    if (res && (res.success === true || res.success === 'true')) {
      responseData = res.data
    } else if (res && Array.isArray(res.entities) && Array.isArray(res.relations)) {
      responseData = res
    } else if (res && res.data && Array.isArray(res.data.entities)) {
      responseData = res.data
    }
    
    if (responseData) {
      currentVersionId.value = version.id
      graphData.value = responseData
      await nextTick()
      try {
        renderChart()
      } catch (e) {
        // 图表渲染错误不影响数据加载
        console.error('图表渲染出错:', e)
      }
    } else {
      console.error('无法解析响应:', res)
      ElMessage.error('数据格式错误')
    }
  } catch (error: any) {
    console.error('handleSelectVersion error:', error)
    const errorMsg = error?.response?.data?.message || error?.message || '网络请求异常'
    ElMessage.error(`加载失败: ${errorMsg}`)
  }
}

async function handleDeleteVersion(version: GraphVersion) {
  try {
    await ElMessageBox.confirm(`确定要删除图谱版本 "${version.name}" 吗？`, '确认删除', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await knowledgeGraphAPI.deleteVersion(props.bookId, version.id)
    if (res.success) {
      ElMessage.success('删除成功')
      await fetchVersions()
      if (currentVersionId.value === version.id) {
        graphData.value = { entities: [], relations: [] }
        if (chartInstance) {
          chartInstance.clear()
        }
      }
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('Delete version failed:', error)
    }
  }
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 7) return `${days}天前`
  
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
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

onMounted(async () => {
  await nextTick()
  
  // 先获取版本列表
  await fetchVersions()
  
  window.addEventListener('resize', handleResize)

  if (props.apiConfigs.length > 0) {
    const defaultConfig = props.apiConfigs.find(c => c.is_default)
    analyzeConfigId.value = defaultConfig?.id || props.apiConfigs[0].id
  }

  // 如果有版本，自动选择第一个
  if (versions.value.length > 0 && !currentVersionId.value) {
    await handleSelectVersion(versions.value[0])
  }

  // 额外的延迟渲染，确保容器完全展开
  setTimeout(() => {
    if (currentVersionId.value) {
      renderChart()
    }
  }, 500)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  disposeChart()
})

defineExpose({
  fetchGraphData,
  handleAnalyze
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

.kg-versions {
  width: 260px;
  min-width: 260px;
  border-right: 1px solid var(--el-border-color-lighter);
  display: flex;
  flex-direction: column;
  background: var(--el-fill-color-lighter);
}

.versions-header {
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.versions-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.versions-actions {
  display: flex;
  gap: 4px;
}

.versions-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.version-item {
  padding: 10px;
  margin-bottom: 8px;
  background: var(--el-bg-color);
  border-radius: 6px;
  border: 1px solid var(--el-border-color-lighter);
  cursor: pointer;
  transition: all 0.2s ease;
}

.version-item:hover {
  border-color: var(--el-color-primary-light-3);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.version-item.active {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}

.version-main {
  margin-bottom: 6px;
}

.version-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  margin-bottom: 3px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.version-time {
  font-size: 10px;
  color: var(--el-text-color-secondary);
}

.version-stats {
  display: flex;
  gap: 6px;
  margin-bottom: 5px;
}

.version-stat {
  font-size: 10px;
  color: var(--el-text-color-regular);
}

.version-desc {
  font-size: 10px;
  color: var(--el-text-color-secondary);
  line-height: 1.3;
  margin-bottom: 5px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.version-actions {
  display: flex;
  justify-content: flex-end;
}

.versions-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 16px;
  color: var(--el-text-color-secondary);
  text-align: center;
}

.versions-empty .el-icon {
  font-size: 48px;
  margin-bottom: 12px;
  opacity: 0.5;
}

.versions-empty p {
  margin: 4px 0;
}

.empty-hint {
  font-size: 12px;
  color: var(--el-text-color-placeholder);
}

.kg-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.graph-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  gap: 8px;
  flex-shrink: 0;
}

.graph-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  white-space: nowrap;
}

.graph-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

.graph-scope-section {
  padding: 10px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
}

.scope-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.scope-label {
  font-size: 13px;
  color: var(--el-text-color-regular);
  font-weight: 500;
}

.scope-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.scope-chip {
  padding: 6px 14px;
  border-radius: 16px;
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

.chapter-select-row {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.chapter-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.graph-stats {
  display: flex;
  gap: 16px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  flex-shrink: 0;
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
  min-height: 480px;
  position: relative;
  width: 100%;
}

.graph-empty,
.graph-loading {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
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
</style>
