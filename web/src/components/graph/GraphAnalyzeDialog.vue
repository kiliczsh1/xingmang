<template>
  <el-dialog
    v-model="visible"
    title="AI 分析章节"
    width="640px"
    append-to-body
    class="xm-dialog graph-analyze-dialog"
    :close-on-click-modal="!running"
    :close-on-press-escape="!running"
    :show-close="!running"
  >
    <div class="ga-body">
      <el-alert
        v-if="models.length === 0 && !running"
        type="warning"
        :closable="false"
        show-icon
        title="没有可用的 AI 模型"
        description="请先到「配置」里添加并启用一个模型，再回来做 AI 分析。"
      />
      <el-alert
        v-else-if="chapters.length === 0 && !running"
        type="warning"
        :closable="false"
        show-icon
        title="这本书还没有章节"
        description="先导入章节内容后再做实体与关系抽取。"
      />

      <template v-else>
        <!-- 配置区 -->
        <div class="ga-field">
          <label class="ga-label">分析模型</label>
          <el-select v-model="modelId" size="small" :disabled="running" placeholder="选择模型" style="width: 100%">
            <el-option
              v-for="model in models"
              :key="model.id"
              :label="model.name || model.model"
              :value="model.id"
            />
          </el-select>
        </div>

        <div class="ga-field">
          <label class="ga-label">分析范围</label>
          <el-radio-group v-model="scope" size="small" :disabled="running">
            <el-radio-button value="single">单章</el-radio-button>
            <el-radio-button value="first5">前 5 章</el-radio-button>
            <el-radio-button value="first10">前 10 章</el-radio-button>
            <el-radio-button value="first15">前 15 章</el-radio-button>
            <el-radio-button value="custom">指定章节</el-radio-button>
          </el-radio-group>
        </div>

        <div class="ga-field">
          <label class="ga-label">拆分方式</label>
          <el-radio-group v-model="mode" size="small" :disabled="running">
            <el-radio-button value="merge">合并拆（一次跑完）</el-radio-button>
            <el-radio-button value="split">分章拆（逐章串行）</el-radio-button>
          </el-radio-group>
          <div class="ga-hint">
            合并拆把所选章节拼成一份正文整体抽取（跨章关系更完整，默认）；分章拆按章节逐条请求。
          </div>
        </div>

        <div v-if="scope === 'single'" class="ga-field">
          <label class="ga-label">选择章节</label>
          <el-select-v2
            v-model="singleChapterId"
            size="small"
            filterable
            :disabled="running"
            :options="chapterOptions"
            placeholder="搜索章节"
            style="width: 100%"
          />
        </div>

        <div v-else-if="scope === 'custom'" class="ga-field">
          <label class="ga-label">选择章节（可多选）</label>
          <el-select-v2
            v-model="customChapterIds"
            size="small"
            multiple
            filterable
            collapse-tags
            collapse-tags-tooltip
            :disabled="running"
            :options="chapterOptions"
            placeholder="搜索并多选章节"
            style="width: 100%"
          />
        </div>

        <div class="ga-summary">
          本次将分析 <strong>{{ targetChapters.length }}</strong> 章，
          约 <strong>{{ totalChars }}</strong> 字。
          <span v-if="targetChapters.length === 0" class="ga-warn">请先选择章节。</span>
          <span v-else-if="targetChapters.length > CHAPTER_WARN" class="ga-warn">
            章节较多时耗时较长，建议先用单章或前 5 章试跑。
          </span>
        </div>

        <!-- 进度 -->
        <div v-if="running || progress" class="ga-progress">
          <template v-if="running">
            <!-- 合并拆：不暴露内部分块，整体显示一条流动式加载动画 -->
            <div v-if="mode === 'merge'" class="ga-indeterminate-bar"></div>
            <!-- 分章拆：进度条 + 扫光动画 -->
            <div v-else class="ga-progress-wrap">
              <el-progress :percentage="percent" :stroke-width="10" />
              <div class="ga-progress-sheen"></div>
            </div>
            <div class="ga-progress-text">
              <span class="ga-running">
                <el-icon class="is-loading"><Loading /></el-icon>
                <template v-if="mode === 'merge'">
                  正在整体分析所选章节，一次跑完后汇总…
                </template>
                <template v-else>
                  正在分析第 {{ progress?.done ?? 0 }}/{{ progress?.total ?? targetChapters.length }} 章
                  <template v-if="progress?.currentTitle">· {{ progress.currentTitle }}</template>
                </template>
              </span>
              <span class="ga-progress-count">
                <el-icon class="is-loading ga-timer-ico"><Timer /></el-icon>
                {{ elapsedText }}
                · 已抽 {{ progress?.entityCount ?? 0 }} 实体 · {{ progress?.relationCount ?? 0 }} 关系
              </span>
            </div>
          </template>
          <template v-else>
            <el-progress :percentage="100" :status="'success'" :stroke-width="10" />
            <div class="ga-progress-text">
              <span>{{ result?.stats.cancelled ? '已取消' : '已完成' }}</span>
              <span class="ga-progress-count">
                <el-icon class="ga-timer-ico"><Timer /></el-icon>
                {{ elapsedText }}
                · 已抽 {{ progress?.entityCount ?? 0 }} 实体 · {{ progress?.relationCount ?? 0 }} 关系
              </span>
            </div>
          </template>
        </div>

        <!-- 结果预览 -->
        <div v-if="result" class="ga-result">
          <div v-if="result.stats.cancelled" class="ga-cancelled-tip">
            <el-icon><WarningFilled /></el-icon>
            <span>
              已取消分析。为防止算力浪费，下面保留的是取消前已抽取的
              <strong>部分结果</strong>，可存为新图谱，也可直接关闭丢弃。
            </span>
          </div>
          <div class="ga-result-row">
            <span>实体</span><strong>{{ result.stats.entityCount }}</strong>
          </div>
          <div class="ga-result-row">
            <span>关系</span><strong>{{ result.stats.relationCount }}</strong>
          </div>
          <div class="ga-result-row muted">
            <span>重复合并</span>
            <span>实体 {{ result.stats.duplicateEntities }} · 关系 {{ result.stats.duplicateRelations }}</span>
          </div>
          <div v-if="result.stats.reusedWithBase > 0" class="ga-result-row muted">
            <span>复用已有实体</span>
            <span>{{ result.stats.reusedWithBase }}（与当前图谱同名，不重复建点）</span>
          </div>
          <div v-if="mergePreview" class="ga-result-row">
            <span>合并后预览</span>
            <span>
              实体 {{ baseGraph.entities.length + mergePreview.newEntities }}
              （新增 {{ mergePreview.newEntities }} · 复用 {{ mergePreview.reusedEntities }}）
              · 关系 {{ baseGraph.relations.length + mergePreview.newRelations }}
              （新增 {{ mergePreview.newRelations }} · 去重 {{ mergePreview.duplicateRelations }}）
            </span>
          </div>
          <div v-if="result.stats.droppedRelations > 0" class="ga-result-row muted">
            <span>丢弃关系</span>
            <span>{{ result.stats.droppedRelations }}（端点未出现在实体列表）</span>
          </div>
          <div v-if="result.stats.failedChunks > 0" class="ga-result-row muted">
            <span>请求失败</span>
            <span>{{ result.stats.failedChunks }} 块（网络或上游报错）</span>
          </div>
          <div v-if="result.stats.parseFailures > 0" class="ga-result-row muted">
            <span>解析失败</span>
            <span>{{ result.stats.parseFailures }} 块（模型没按 JSON 输出，或输出被截断）</span>
          </div>
          <div v-if="result.stats.emptyChunks > 0" class="ga-result-row muted">
            <span>正文为空</span>
            <span>{{ result.stats.emptyChunks }} 块（输出额度被思考过程耗尽，未生成正文）</span>
          </div>
        </div>

        <div v-if="error" class="ga-error">{{ error }}</div>

        <div v-if="result && !running" class="ga-field">
          <label class="ga-label">保存内容</label>
          <el-radio-group v-model="saveMode" size="small">
            <el-radio-button value="merge">合并到当前图谱（推荐）</el-radio-button>
            <el-radio-button value="replace">只存本次结果</el-radio-button>
          </el-radio-group>
          <div class="ga-hint">
            <template v-if="saveMode === 'merge'">
              在当前图谱基础上<strong>累加合并</strong>本次抽取：同名实体复用旧节点，新实体自动与旧实体连边；
              存为一个<strong>新版本</strong>快照，旧版本不动、可回退。
            </template>
            <template v-else>
              本次结果单独存为一个<strong>新版本</strong>，不与当前图谱合并（旧行为）；可在左栏版本列表中切换或删除。
            </template>
          </div>
        </div>

        <div v-if="result && !running && saveMode === 'merge'" class="ga-field">
          <el-checkbox v-model="linkOrphans" size="small">
            为孤立节点推断补边{{ orphanCount > 0 ? `（当前 ${orphanCount} 个孤立节点）` : '' }}
          </el-checkbox>
          <div class="ga-hint">
            勾选后保存时会额外发起一次 AI 请求，为没有任何连边的实体推断关系；关系标签带<strong>「推断·」</strong>前缀便于识别，失败会自动跳过、不影响保存。
          </div>
        </div>

        <div v-if="result && !running" class="ga-field">
          <label class="ga-label">保存为版本</label>
          <el-input v-model="versionName" size="small" placeholder="版本名称" />
        </div>
      </template>
    </div>

    <template #footer>
      <div class="ga-footer">
        <el-button v-if="running" type="danger" plain @click="cancel">取消分析</el-button>
        <template v-else>
          <el-button @click="visible = false">关闭</el-button>
          <el-button
            v-if="!result"
            type="primary"
            :disabled="!canStart"
            @click="start"
          >
            开始分析
          </el-button>
          <el-button
            v-else
            type="primary"
            :disabled="!canSave"
            :loading="saving"
            @click="save"
          >
            存为新版本
          </el-button>
        </template>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Loading, Timer, WarningFilled } from '@element-plus/icons-vue'
import { configAPI } from '@/api'
import { analyzeChapters, linkOrphanEntities, type AnalyzeMode, type AnalyzeProgress, type AnalyzeResult } from '@/services/graphAnalyzeService'
import { mergeGraphData } from '@/services/graphMergeService'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'
import type { ApiModel, Chapter } from '@/types'

/** 超过这个章节数就提示耗时风险 */
const CHAPTER_WARN = 20

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved', versionName: string): void
}>()

const store = useKnowledgeGraphStore()

const visible = computed({
  get: () => props.visible,
  set: value => emit('update:visible', value)
})

const models = ref<ApiModel[]>([])
const modelId = ref<number>()
const scope = ref<'single' | 'first5' | 'first10' | 'first15' | 'custom'>('single')
/** 拆分方式：合并拆（默认，所选章节拼成一份正文一次跑完）/ 分章拆（逐章） */
const mode = ref<AnalyzeMode>('merge')
const singleChapterId = ref<number>()
const customChapterIds = ref<number[]>([])

const running = ref(false)
const progress = ref<AnalyzeProgress | null>(null)
const result = ref<AnalyzeResult | null>(null)
const error = ref<string | null>(null)
const versionName = ref('')
/** 保存内容：merge 累加合并到当前图谱（默认）；replace 只存本次结果（旧行为） */
const saveMode = ref<'merge' | 'replace'>('merge')
/** 运行耗时（毫秒），由计时器累加 */
const elapsedMs = ref(0)
/** 保存中（合并/补边/导入都有耗时），防止重复点击 */
const saving = ref(false)
/** 孤岛补边：保存合并时为孤立节点推断关系 */
const linkOrphans = ref(false)
let timer: number | null = null
let controller: AbortController | null = null

const chapters = computed<Chapter[]>(() => store.chapters)

const chapterOptions = computed(() =>
  chapters.value.map((chapter, index) => ({
    label: `${String(index + 1).padStart(4, '0')} ${chapter.title || `第${index + 1}章`}`,
    value: chapter.id
  }))
)

const targetChapters = computed(() => {
  const all = chapters.value
  switch (scope.value) {
    case 'single':
      return all.filter(c => c.id === singleChapterId.value)
    case 'first5':
      return all.slice(0, 5)
    case 'first10':
      return all.slice(0, 10)
    case 'first15':
      return all.slice(0, 15)
    case 'custom': {
      const picked = new Set(customChapterIds.value)
      return all.filter(c => picked.has(c.id))
    }
    default:
      return []
  }
})

const totalChars = computed(() =>
  targetChapters.value.reduce((sum, c) => sum + (c.content?.length || 0), 0)
)

const percent = computed(() => {
  const total = progress.value?.total || targetChapters.value.length
  if (total === 0) return 0
  return Math.round(((progress.value?.done || 0) / total) * 100)
})

/** 运行耗时文本，mm:ss，超过 1 小时显示 h:mm:ss */
const elapsedText = computed(() => {
  const s = Math.floor(elapsedMs.value / 1000)
  const hh = Math.floor(s / 3600)
  const mm = Math.floor((s % 3600) / 60)
  const ss = s % 60
  if (hh > 0) return `${hh}:${String(mm).padStart(2, '0')}:${String(ss).padStart(2, '0')}`
  return `${mm}:${String(ss).padStart(2, '0')}`
})

function startTimer() {
  elapsedMs.value = 0
  timer = window.setInterval(() => {
    if (!running.value) return
    elapsedMs.value += 100
  }, 100)
}

function stopTimer() {
  if (timer !== null) {
    window.clearInterval(timer)
    timer = null
  }
}

const canStart = computed(
  () => !!modelId.value && targetChapters.value.length > 0 && !running.value
)
const canSave = computed(
  () => !!result.value && (result.value.stats.entityCount > 0 || result.value.stats.relationCount > 0) && !!versionName.value.trim()
)

function defaultVersionName(): string {
  const picked = targetChapters.value
  const first = picked[0]?.title || ''
  const suffix = picked.length > 1 ? `等${picked.length}章` : first
  const stamp = new Date().toLocaleString('zh-CN', { hour12: false })
  // 增量语义：累加合并的版本名带上「+增量」后缀，方便在版本列表里辨认
  const tail = saveMode.value === 'merge' ? '增量' : '全书'
  return `AI分析_${suffix || tail}_${stamp}${saveMode.value === 'merge' ? '+增量' : ''}`
}

/** 当前版本数据快照：章节关联从 coverage 补齐，合并后保存才不会丢出处 */
const baseGraph = computed(() => ({
  entities: store.entities.map(e => ({
    ...e,
    chapter_ids: store.chapterCoverage.chapterIdsByEntity.get(e.id) || []
  })),
  relations: store.relations.map(r => ({
    ...r,
    chapter_ids: store.chapterCoverage.chapterIdsByRelation.get(r.id) || []
  }))
}))

/** 与当前图谱合并后的预览统计：保存前就知道图会变成多大 */
const mergePreview = computed(() => {
  if (!result.value || saveMode.value !== 'merge') return null
  return mergeGraphData(baseGraph.value, {
    entities: result.value.entities,
    relations: result.value.relations
  }).stats
})

/** 当前版本孤立节点数（度数 0），提示补边价值 */
const orphanCount = computed(() => {
  const degree = store.graphIndex.degree
  return store.entities.filter(e => (degree.get(e.id) || 0) === 0).length
})

async function loadModels() {
  if (models.value.length > 0) return
  try {
    const res: any = await configAPI.getAll()
    const list: ApiModel[] = res?.success && Array.isArray(res.data) ? res.data : []
    models.value = list.filter(m => m.enabled !== 0)
    const preferred = models.value.find(m => Number(m.is_default) === 1) || models.value[0]
    modelId.value = preferred?.id
  } catch (e) {
    console.error('加载模型配置失败:', e)
    models.value = []
  }
}

/** 章节多时先确认，避免一点就开始跑十几分钟 */
async function confirmIfLong(): Promise<boolean> {
  if (targetChapters.value.length <= CHAPTER_WARN) return true
  try {
    await ElMessageBox.confirm(
      `将对 ${targetChapters.value.length} 章发起 AI 分析，耗时可能较长。确定继续吗？`,
      '分析内容较多',
      { type: 'warning', confirmButtonText: '继续', cancelButtonText: '取消' }
    )
    return true
  } catch {
    return false
  }
}

async function start() {
  if (!modelId.value) return
  if (!(await confirmIfLong())) return

  running.value = true
  error.value = null
  result.value = null
  progress.value = { done: 0, total: targetChapters.value.length, currentTitle: '', entityCount: 0, relationCount: 0 }
  controller = new AbortController()
  startTimer()

  try {
    const res = await analyzeChapters({
      bookId: store.bookId ?? 0,
      chapters: targetChapters.value.map(c => ({ id: c.id, title: c.title, content: c.content || '' })),
      configId: modelId.value,
      mode: mode.value,
      // 累加合并：注入当前版本实体名，指向旧实体的关系才能解析成功
      baseEntities:
        store.entities.length > 0
          ? store.entities.map(e => ({ id: e.id, name: e.name, type: e.type }))
          : undefined,
      signal: controller.signal,
      onProgress: p => {
        progress.value = p
      }
    })
    result.value = res
    versionName.value = defaultVersionName()
    if (res.stats.entityCount === 0 && res.stats.relationCount === 0) {
      // 分清「模型没抽到」和「抽到了但没解析出来」：后者换模型也不一定好，
      // 往往是输出上限太小导致 JSON 被截断，得去配置里把那模型的额度调大。
      if (res.stats.emptyChunks > 0) {
        error.value = `全部 ${res.stats.emptyChunks} 个文本块都没生成正文：输出额度被思考过程吃光了。请到「配置」里把这个模型的 max_tokens 调大（或换成非思考模型）后重试。`
      } else if (res.stats.parseFailures > 0) {
        error.value = `${res.stats.parseFailures} 个文本块返回的内容解析不出 JSON 图谱（多半是输出被截断）。请调大该模型的 max_tokens，或减少每次分析的章节数后重试。`
      } else if (res.stats.failedChunks > 0) {
        error.value = `全部 ${res.stats.failedChunks} 个文本块请求失败。请检查模型配置或网络后重试。`
      } else {
        error.value = '本次没有抽到实体或关系。可以换一个模型，或换一章再试。'
      }
    }
  } catch (e: any) {
    error.value = e?.message || '分析失败'
  } finally {
    running.value = false
    stopTimer()
    controller = null
  }
}

function cancel() {
  controller?.abort()
}

async function save() {
  if (!result.value) return
  const name = versionName.value.trim()
  if (!name) {
    ElMessage.warning('请填写版本名称')
    return
  }

  saving.value = true
  try {
    let data = { entities: result.value.entities, relations: result.value.relations }
    if (saveMode.value === 'merge') {
      // 累加：新版本 = 当前图谱 ⊕ 本次抽取（旧版本不动，可回退）
      data = mergeGraphData(baseGraph.value, data).data

      if (linkOrphans.value && modelId.value) {
        try {
          const link = await linkOrphanEntities({
            bookId: store.bookId ?? 0,
            configId: modelId.value,
            base: data
          })
          data = link.data
          if (link.stats.proposed > 0) {
            ElMessage.info(`已为孤立节点推断补充 ${link.stats.proposed} 条关系（带「推断·」前缀）`)
          }
        } catch (e) {
          // 补边失败静默跳过，不阻塞保存主流程
          console.error('孤岛补边失败（已跳过）:', e)
        }
      }
    }

    const res = await store.importAsNewVersion(name, data)
    if (res.ok) {
      ElMessage.success(res.message)
      emit('saved', name)
      visible.value = false
    } else {
      ElMessage.error(res.message)
    }
  } finally {
    saving.value = false
  }
}

watch(visible, open => {
  if (!open) return
  result.value = null
  error.value = null
  progress.value = null
  elapsedMs.value = 0
  void loadModels()
})

onBeforeUnmount(() => stopTimer())
</script>

<style scoped>
.ga-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.ga-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.ga-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}

.ga-summary {
  font-size: 12px;
  color: var(--el-text-color-regular);
}

.ga-summary strong {
  color: var(--el-color-primary);
}

.ga-warn {
  color: var(--el-color-warning);
}

.ga-progress {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* 分章拆：进度条外层容器，叠加扫光动画 */
.ga-progress-wrap {
  position: relative;
}

/* 左往右扫过的高光，让百分比条看起来“活着” */
.ga-progress-sheen {
  position: absolute;
  right: 0;
  left: 0;
  bottom: 2px;
  height: 10px;
  border-radius: 6px;
  overflow: hidden;
  pointer-events: none;
}

.ga-progress-sheen::after {
  content: '';
  position: absolute;
  top: 0;
  left: -40%;
  width: 40%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, var(--el-color-primary) 25%, transparent),
    transparent
  );
  animation: ga-loading 1.6s ease-in-out infinite;
}

/* 合并拆的整体加载动画：一条横向流动的高光，代表“一口气跑完” */
.ga-indeterminate-bar {
  position: relative;
  height: 10px;
  border-radius: 6px;
  overflow: hidden;
  background: var(--el-fill-color);
}

.ga-indeterminate-bar::after {
  content: '';
  position: absolute;
  top: 0;
  left: -35%;
  width: 35%;
  height: 100%;
  border-radius: 6px;
  background: linear-gradient(90deg, transparent, var(--el-color-primary), transparent);
  animation: ga-loading 1.2s ease-in-out infinite;
}

@keyframes ga-loading {
  0% {
    left: -35%;
  }
  100% {
    left: 100%;
  }
}

.ga-running {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.ga-running .is-loading {
  animation: rotating 1.5s linear infinite;
  color: var(--el-color-primary);
}

@keyframes rotating {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.ga-progress-text {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.ga-timer-ico {
  vertical-align: -2px;
  margin-right: 2px;
  color: var(--el-text-color-regular);
}

.ga-progress-count {
  white-space: nowrap;
}

.ga-result {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 10px 12px;
  border-radius: 6px;
  background: var(--el-fill-color-light);
  font-size: 12px;
}

.ga-cancelled-tip {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-bottom: 4px;
  padding-bottom: 6px;
  border-bottom: 1px dashed var(--el-border-color);
  color: var(--el-color-warning);
  line-height: 1.5;
}

.ga-cancelled-tip strong {
  color: var(--el-color-warning);
}

.ga-result-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  color: var(--el-text-color-regular);
}

.ga-result-row strong {
  color: var(--el-color-primary);
  font-size: 14px;
}

.ga-result-row.muted {
  color: var(--el-text-color-secondary);
}

.ga-error {
  font-size: 12px;
  color: var(--el-color-danger);
}

.ga-hint {
  font-size: 11px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.ga-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
