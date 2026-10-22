<template>
<!-- 提示词查重弹窗 -->
<el-dialog
  v-model="dedupDialogVisible"
  title="提示词查重"
  width="800px"
  top="6vh"
  :close-on-click-modal="false"
  destroy-on-close
  class="dedup-dialog"
>
  <div class="dedup-topbar">
    <div class="dedup-mode">
      <span class="dedup-mode-label">判重口径</span>
      <el-radio-group v-model="dedupMode" size="small" @change="rebuildDedupGroups">
        <el-radio-button value="titleCategory">标题 + 分类</el-radio-button>
        <el-radio-button value="title">仅标题</el-radio-button>
      </el-radio-group>
    </div>
    <div v-if="dedupGroups.length > 0" class="dedup-stats">
      共 <b>{{ dedupGroupCount }}</b> 组重复，涉及 <b>{{ dedupTotalCount }}</b> 张卡片
    </div>
  </div>

  <div v-if="dedupLoading" class="dedup-loading">正在扫描重复提示词…</div>
  <el-empty v-else-if="dedupGroups.length === 0" description="未发现重复提示词" :image-size="90" />
  <div v-else class="dedup-groups">
    <div v-for="group in dedupGroups" :key="group.key" class="dedup-group">
      <div class="dedup-group-head">
        <span class="dedup-group-title" :title="group.name">{{ group.name }}</span>
        <span v-if="group.category !== undefined" class="dedup-group-cat">{{ group.category }}</span>
        <span class="dedup-group-count">×{{ group.items.length }}</span>
        <span class="dedup-group-spacer"></span>
        <el-button size="small" text @click="setDedupKeep(group.key, 'oldest')">保留最早</el-button>
        <el-button size="small" text @click="setDedupKeep(group.key, 'newest')">保留最新</el-button>
      </div>
      <div class="dedup-group-items">
        <div
          v-for="item in group.items"
          :key="item.id"
          class="dedup-item"
          :class="{ selected: dedupSelected.has(item.id) }"
        >
          <el-checkbox
            :model-value="dedupSelected.has(item.id)"
            :disabled="item.card_type === 'encrypted'"
            @change="toggleDedupItem(item.id, $event)"
          />
          <div class="dedup-item-body">
            <div class="dedup-item-line1">
              <el-tag v-if="item.card_type === 'encrypted'" size="small" type="warning" class="dedup-lock-tag">
                <el-icon><Lock /></el-icon>
                需解锁
              </el-tag>
              <span class="dedup-item-name">{{ item.name }}</span>
              <span class="dedup-item-cat">{{ item.category }}</span>
              <span class="dedup-item-date">{{ formatDate(item.created_at) }}</span>
            </div>
            <div class="dedup-item-content">{{ getDedupContentPreview(item) }}</div>
          </div>
          <el-tag v-if="isDedupKeepItem(group, item)" size="small" type="success" effect="light" class="dedup-keep-tag">保留</el-tag>
        </div>
      </div>
    </div>
  </div>
  <template #footer>
    <span v-if="dedupSelectedCount > 0" class="dedup-footer-tip">
      将删除 {{ dedupSelectedCount }} 张重复卡片（加密卡片需先解锁，不参与一键删除）
    </span>
    <el-button @click="dedupDialogVisible = false">关闭</el-button>
    <el-button
      type="danger"
      :loading="dedupDeleting"
      :disabled="dedupSelectedCount === 0"
      @click="handleDedupDelete"
    >
      删除选中{{ dedupSelectedCount > 0 ? ` ${dedupSelectedCount} 张` : '' }}
    </el-button>
  </template>
</el-dialog>

<!-- 卡包合并弹窗 -->
<el-dialog
  v-model="mergeDialogVisible"
  title="卡包合并"
  width="760px"
  top="6vh"
  :close-on-click-modal="false"
  destroy-on-close
  class="merge-dialog"
>
  <div class="merge-mode-bar">
    <el-radio-group v-model="mergeMode" size="small">
      <el-radio-button value="new">合并为新卡包</el-radio-button>
      <el-radio-button value="absorb">主包吸收副包</el-radio-button>
    </el-radio-group>
    <span class="merge-mode-desc">
      {{ mergeMode === 'new' ? '多个自定义卡包合并为一个新卡包（需自定义命名）' : '选择主卡包与副卡包，副包卡片并入主包' }}
    </span>
  </div>

  <el-empty v-if="mergeablePacks.length === 0" description="暂无可合并的自定义卡包" :image-size="80" />

  <div v-else class="merge-body">
    <!-- 模式1：合并为新卡包 -->
    <template v-if="mergeMode === 'new'">
      <div class="merge-section">
        <div class="merge-section-title">新卡包名称</div>
        <el-input
          v-model="mergeNewName"
          placeholder="请输入新卡包名称"
          maxlength="40"
          clearable
        />
        <div class="merge-section-tip">合并后将创建该名称的新卡包，源卡包自动移除</div>
      </div>
      <div class="merge-section">
        <div class="merge-section-title">选择源卡包 <span class="merge-section-count">{{ mergeSelectedSources.length }} / {{ mergeablePacks.length }}</span></div>
        <div class="merge-pack-list">
          <el-checkbox-group v-model="mergeSelectedSources">
            <div v-for="pack in mergeablePacks" :key="pack.name" class="merge-pack-item">
              <el-checkbox :value="pack.name">
                <div class="merge-pack-info">
                  <el-icon class="merge-pack-ic"><FolderOpened /></el-icon>
                  <span class="merge-pack-name">{{ pack.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                  <el-tag v-if="pack.isBookAnalysis" size="small" type="info" effect="plain" class="merge-pack-tag">拆书</el-tag>
                  <span class="merge-pack-count">{{ pack.prompts.length }} 张</span>
                </div>
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </div>
      <div class="merge-preview">
        将合并 <b>{{ mergeSelectedSources.length }}</b> 个卡包，共 <b>{{ mergePreviewTotal }}</b> 张卡片
      </div>
    </template>

    <!-- 模式2：主包吸收副包 -->
    <template v-else>
      <div class="merge-section">
        <div class="merge-section-title">主卡包 <span class="merge-section-tip-inline">默认卡包也可作主包</span></div>
        <div class="merge-pack-list merge-pack-list--radio">
          <el-radio-group v-model="mergeMainName">
            <div v-for="pack in mergeMainOptions" :key="pack.name" class="merge-pack-item">
              <el-radio :value="pack.name">
                <div class="merge-pack-info">
                  <el-icon class="merge-pack-ic"><FolderOpened /></el-icon>
                  <span class="merge-pack-name">{{ pack.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                  <el-tag v-if="pack.isBookAnalysis" size="small" type="info" effect="plain" class="merge-pack-tag">拆书</el-tag>
                  <el-tag v-if="isDefaultPack(pack.name)" size="small" type="warning" effect="light" class="merge-pack-tag">默认</el-tag>
                  <span class="merge-pack-count">{{ pack.prompts.length }} 张</span>
                </div>
              </el-radio>
            </div>
          </el-radio-group>
        </div>
      </div>
      <div class="merge-section">
        <div class="merge-section-title">选择副卡包 <span class="merge-section-count">{{ mergeSelectedSubs.length }} / {{ mergeAvailableSubs.length }}</span></div>
        <div v-if="mergeAvailableSubs.length === 0" class="merge-empty-subs">请先选择主卡包，或无更多自定义卡包可选</div>
        <div v-else class="merge-pack-list">
          <el-checkbox-group v-model="mergeSelectedSubs">
            <div v-for="pack in mergeAvailableSubs" :key="pack.name" class="merge-pack-item">
              <el-checkbox :value="pack.name">
                <div class="merge-pack-info">
                  <el-icon class="merge-pack-ic"><FolderOpened /></el-icon>
                  <span class="merge-pack-name">{{ pack.name.replace(BOOK_ANALYSIS_PREFIX, '') }}</span>
                  <el-tag v-if="pack.isBookAnalysis" size="small" type="info" effect="plain" class="merge-pack-tag">拆书</el-tag>
                  <span class="merge-pack-count">{{ pack.prompts.length }} 张</span>
                </div>
              </el-checkbox>
            </div>
          </el-checkbox-group>
        </div>
      </div>
      <div class="merge-preview" v-if="mergeMainName">
        将 {{ mergeSelectedSubs.length }} 个副卡包并入「{{ mergeMainName.replace(BOOK_ANALYSIS_PREFIX, '') }}」，共移动 <b>{{ mergePreviewTotal }}</b> 张卡片
      </div>
      <div class="merge-preview" v-else>请选择主卡包</div>
    </template>

    <div class="merge-note">
      <el-icon><InfoFilled /></el-icon>
      <span>合并不去重，如产生同名重复卡片可使用「提示词查重」工具清理</span>
    </div>
  </div>

  <template #footer>
    <el-button @click="mergeDialogVisible = false">取消</el-button>
    <el-button
      type="primary"
      :loading="mergeLoading"
      :disabled="!canMerge"
      @click="handleMergeExecute"
    >
      开始合并
    </el-button>
  </template>
</el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { promptAPI } from '@/api'
import type { Prompt } from '@/types'
import { buildPromptKey, formatDate } from '@/utils/promptDedupMerge'
import type { DedupGroup, DedupMode, MergeMode } from '@/utils/promptDedupMerge'
import { BOOK_ANALYSIS_PREFIX, DEFAULT_CATEGORIES } from '@/constants/promptCategories'

const props = defineProps<{
  prompts: Prompt[]
  allExportablePacks: { name: string; prompts: Prompt[]; isBookAnalysis: boolean }[]
  customCategories: string[]
  refresh: () => Promise<void>
}>()

const emit = defineEmits<{
  (e: 'add-category', name: string): void
  (e: 'remove-categories', names: string[]): void
}>()

// ============ 提示词查重（快捷中心） ============
const dedupDialogVisible = ref(false)
const dedupLoading = ref(false)
const dedupDeleting = ref(false)
const dedupMode = ref<DedupMode>('titleCategory')
const dedupGroups = ref<DedupGroup[]>([])
const dedupSelected = ref<Set<number>>(new Set())

const dedupGroupCount = computed(() => dedupGroups.value.length)
const dedupTotalCount = computed(() => dedupGroups.value.reduce((sum, g) => sum + g.items.length, 0))
const dedupSelectedCount = computed(() => dedupSelected.value.size)

// 重建查重分组（组内按 id 升序，keepId 默认最早一条；默认勾选每组除保留项外的全部非加密卡片）
const rebuildDedupGroups = () => {
  const groupsMap = new Map<string, Prompt[]>()
  props.prompts.forEach(p => {
    const key = buildPromptKey(p.name, p.category, dedupMode.value)
    if (!key) return
    const list = groupsMap.get(key) || []
    list.push(p)
    groupsMap.set(key, list)
  })

  const groups: DedupGroup[] = []
  groupsMap.forEach((items, key) => {
    if (items.length < 2) return
    items.sort((a, b) => a.id - b.id)
    const [first] = items
    groups.push({
      key,
      name: first.name,
      category: dedupMode.value === 'title' ? undefined : first.category,
      items,
      keepId: first.id
    })
  })
  groups.sort((a, b) => b.items.length - a.items.length || a.name.localeCompare(b.name, 'zh-CN'))
  dedupGroups.value = groups

  const selected = new Set<number>()
  groups.forEach(g => {
    g.items.forEach(item => {
      if (item.id !== g.keepId && item.card_type !== 'encrypted') selected.add(item.id)
    })
  })
  dedupSelected.value = selected
}

const toggleDedupItem = (id: number, checked: boolean) => {
  const selected = new Set(dedupSelected.value)
  if (checked) selected.add(id)
  else selected.delete(id)
  dedupSelected.value = selected
}

// 每组保留最早/最新，其余非加密卡片勾选删除
const setDedupKeep = (groupKey: string, mode: 'oldest' | 'newest') => {
  const group = dedupGroups.value.find(g => g.key === groupKey)
  if (!group) return
  const sorted = [...group.items].sort((a, b) => (mode === 'oldest' ? a.id - b.id : b.id - a.id))
  group.keepId = sorted[0]?.id ?? group.keepId
  const selected = new Set(dedupSelected.value)
  group.items.forEach(item => {
    if (item.id === group.keepId || item.card_type === 'encrypted') selected.delete(item.id)
    else selected.add(item.id)
  })
  dedupSelected.value = selected
}

const isDedupKeepItem = (group: DedupGroup, item: Prompt) => item.id === group.keepId

const getDedupContentPreview = (item: Prompt) => {
  if (item.card_type === 'encrypted') return '内容已加密，需解锁后查看'
  const text = (item.content || '').replace(/\s+/g, ' ').trim()
  return text.length > 80 ? text.slice(0, 80) + '…' : text || '（无内容）'
}

const handleDedupDelete = async () => {
  if (dedupSelected.value.size === 0) return
  try {
    await ElMessageBox.confirm(
      `确定删除选中的 ${dedupSelected.value.size} 张重复提示词吗？此操作不可恢复！`,
      '删除重复提示词',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (error) {
    return // 取消删除
  }

  dedupDeleting.value = true
  let successCount = 0
  let failCount = 0
  try {
    const ids = Array.from(dedupSelected.value)
    const results = await Promise.all(ids.map(id => promptAPI.delete(id)))
    results.forEach(res => (res.success ? successCount++ : failCount++))
    if (failCount === 0) {
      ElMessage.success(`已删除 ${successCount} 张重复提示词`)
    } else {
      ElMessage.warning(`删除完成：成功 ${successCount} 张，失败 ${failCount} 张`)
    }
    await props.refresh()
    await nextTick() // 等父组件把新 prompts 传下来再重建分组
    rebuildDedupGroups()
    if (dedupGroups.value.length === 0) {
      dedupDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('删除失败，请重试')
  } finally {
    dedupDeleting.value = false
  }
}

// 打开查重弹窗（先刷新数据再扫描）
const openDedup = async () => {
  dedupDialogVisible.value = true
  dedupLoading.value = true
  try {
    // 重新拉取，保证数据最新
    await props.refresh()
    await nextTick() // 等新 prompts 传下来
    rebuildDedupGroups()
  } finally {
    dedupLoading.value = false
  }
}

// ============ 卡包合并（快捷中心 · 新建分组） ============
const mergeDialogVisible = ref(false)
const mergeLoading = ref(false)
const mergeMode = ref<MergeMode>('new')
const mergeNewName = ref('')
const mergeSelectedSources = ref<string[]>([]) // 模式1：选中的源卡包名
const mergeMainName = ref('') // 模式2：主卡包名
const mergeSelectedSubs = ref<string[]>([]) // 模式2：选中的副卡包名

// 判断是否默认卡包（DEFAULT_CATEGORIES 中的名称）
const isDefaultPack = (name: string): boolean => DEFAULT_CATEGORIES.includes(name)

const mergeablePacks = computed(() =>
  props.allExportablePacks
    .filter(pack => !isDefaultPack(pack.name) && pack.prompts.length > 0)
    .map(pack => ({ name: pack.name, prompts: pack.prompts, isBookAnalysis: pack.isBookAnalysis }))
)

// 可作主卡包的选项（默认卡包也可作主包，卡片数>0）
const mergeMainOptions = computed(() =>
  props.allExportablePacks
    .filter(pack => pack.prompts.length > 0)
    .map(pack => ({ name: pack.name, prompts: pack.prompts, isBookAnalysis: pack.isBookAnalysis }))
)

// 模式2可选副卡包：mergeablePacks 排除已选主包
const mergeAvailableSubs = computed(() =>
  mergeablePacks.value.filter(pack => pack.name !== mergeMainName.value)
)

// 预览将移动的卡片总数
const mergePreviewTotal = computed(() => {
  if (mergeMode.value === 'new') {
    return mergeablePacks.value
      .filter(pack => mergeSelectedSources.value.includes(pack.name))
      .reduce((sum, pack) => sum + pack.prompts.length, 0)
  }
  return mergeAvailableSubs.value
    .filter(pack => mergeSelectedSubs.value.includes(pack.name))
    .reduce((sum, pack) => sum + pack.prompts.length, 0)
})

const canMerge = computed(() => {
  if (mergeMode.value === 'new') {
    return mergeNewName.value.trim() !== '' && mergeSelectedSources.value.length >= 2
  }
  return mergeMainName.value !== '' && mergeSelectedSubs.value.length >= 1
})

// 打开合并弹窗（重置为「合并为新卡包」初始态）
const openMerge = () => {
  mergeMode.value = 'new'
  mergeNewName.value = ''
  mergeSelectedSources.value = []
  mergeMainName.value = ''
  mergeSelectedSubs.value = []
  mergeDialogVisible.value = true
}

// 切换模式时清空选择
watch(mergeMode, () => {
  mergeNewName.value = ''
  mergeSelectedSources.value = []
  mergeMainName.value = ''
  mergeSelectedSubs.value = []
})

const handleMergeExecute = async () => {
  if (!canMerge.value) return

  // 确定目标卡包名与受影响卡包名
  let targetName = ''
  let sourceNames: string[] = []
  let totalCards = 0
  if (mergeMode.value === 'new') {
    targetName = mergeNewName.value.trim()
    sourceNames = [...mergeSelectedSources.value]
    totalCards = mergePreviewTotal.value
    // 校验新名
    if (!targetName) {
      ElMessage.warning('请输入新卡包名称')
      return
    }
    if (isDefaultPack(targetName)) {
      ElMessage.warning('新卡包名称不能与默认分类同名')
      return
    }
    const allNames = props.allExportablePacks.map(p => p.name).concat(props.customCategories)
    if (allNames.includes(targetName)) {
      ElMessage.warning('该卡包名称已存在，请更换')
      return
    }
  } else {
    targetName = mergeMainName.value
    sourceNames = [...mergeSelectedSubs.value]
    totalCards = mergePreviewTotal.value
    if (!targetName) {
      ElMessage.warning('请选择主卡包')
      return
    }
  }

  const targetDisplay = targetName.replace(BOOK_ANALYSIS_PREFIX, '')
  try {
    await ElMessageBox.confirm(
      `确定将 ${sourceNames.length} 个卡包合并到「${targetDisplay}」吗？将移动 ${totalCards} 张卡片，源卡包合并后自动移除。`,
      '卡包合并',
      { confirmButtonText: '开始合并', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return // 取消
  }

  mergeLoading.value = true
  let successCount = 0
  let failCount = 0
  try {
    // 收集需要移动的卡片（源卡包 → 目标卡包）
    const packsToMove = mergeMode.value === 'new'
      ? mergeablePacks.value.filter(p => mergeSelectedSources.value.includes(p.name))
      : mergeAvailableSubs.value.filter(p => mergeSelectedSubs.value.includes(p.name))

    for (const pack of packsToMove) {
      for (const prompt of pack.prompts) {
        try {
          const res = await promptAPI.update(prompt.id, { ...prompt, category: targetName })
          if (res.success) successCount++
          else failCount++
        } catch {
          failCount++
        }
      }
    }

    // 模式1：新卡包加入 customCategories（若非默认分类）
    if (mergeMode.value === 'new' && !isDefaultPack(targetName) && !props.customCategories.includes(targetName)) {
      emit('add-category', targetName)
    }

    // 移除已合并的自定义源/副卡包（默认卡包不移除）
    const namesToRemove = new Set(sourceNames.filter(n => !isDefaultPack(n)))
    if (namesToRemove.size > 0) {
      emit('remove-categories', [...namesToRemove])
    }

    await props.refresh()

    if (failCount === 0) {
      ElMessage.success(`合并完成：已移动 ${successCount} 张卡片到「${targetDisplay}」`)
    } else {
      ElMessage.warning(`合并完成：成功 ${successCount} 张，失败 ${failCount} 张`)
    }
    mergeDialogVisible.value = false
  } catch (e) {
    ElMessage.error('合并失败，请重试')
  } finally {
    mergeLoading.value = false
  }
}

defineExpose({ openDedup, openMerge })
</script>

<style scoped>
/* ========== 提示词查重弹窗 ========== */
.dedup-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}
.dedup-mode {
  display: flex;
  align-items: center;
  gap: 8px;
}
.dedup-mode-label {
  font-size: 13px;
  color: #475569;
}
.dedup-stats {
  font-size: 12px;
  color: #64748b;
}
.dedup-stats b {
  color: #ef4444;
}
.dedup-loading {
  padding: 40px 0;
  text-align: center;
  color: #94a3b8;
  font-size: 13px;
}
.dedup-groups {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 56vh;
  overflow-y: auto;
}
.dedup-group {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  overflow: hidden;
  background: #f8fafc;
}
.dedup-group-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #ffffff;
  border-bottom: 1px solid #eef0f3;
}
.dedup-group-title {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 260px;
}
.dedup-group-cat {
  font-size: 11px;
  color: #4f46e5;
  background: #eef2ff;
  padding: 1px 8px;
  border-radius: 999px;
  white-space: nowrap;
}
.dedup-group-count {
  font-size: 11px;
  color: #fff;
  background: #ef4444;
  border-radius: 999px;
  padding: 1px 8px;
  white-space: nowrap;
}
.dedup-group-spacer {
  flex: 1;
}
.dedup-group-items {
  padding: 4px 0;
}
.dedup-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px dashed #eef0f3;
}
.dedup-item:last-child {
  border-bottom: none;
}
.dedup-item.selected {
  background: rgba(239, 68, 68, 0.04);
}
.dedup-item-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.dedup-item-line1 {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}
.dedup-lock-tag {
  margin: 0;
}
.dedup-item-name {
  font-size: 13px;
  font-weight: 500;
  color: #1f2937;
}
.dedup-item-cat {
  font-size: 11px;
  color: #94a3b8;
}
.dedup-item-date {
  margin-left: auto;
  font-size: 11px;
  color: #94a3b8;
}
.dedup-item-content {
  font-size: 12px;
  color: #64748b;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.dedup-keep-tag {
  margin: 0;
  flex-shrink: 0;
}
.dedup-footer-tip {
  font-size: 12px;
  color: #ef4444;
  margin-right: auto;
}

/* 深色主题适配 */
:root[data-theme='dark'] .dedup-group {
  background: #1e2434;
  border-color: #2e3650;
}
:root[data-theme='dark'] .dedup-group-head {
  background: #222838;
  border-color: #2e3650;
}
:root[data-theme='dark'] .dedup-group-title {
  color: #e5e7eb;
}
:root[data-theme='dark'] .dedup-item {
  border-color: #2e3650;
}
:root[data-theme='dark'] .dedup-item.selected {
  background: rgba(239, 68, 68, 0.1);
}
:root[data-theme='dark'] .dedup-item-name {
  color: #e5e7eb;
}
:root[data-theme='dark'] .dedup-mode-label {
  color: #9ca3af;
}

/* ========== 卡包合并弹窗 ========== */
.merge-mode-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 14px;
  margin-bottom: 16px;
  border-bottom: 1px solid #eef0f3;
  flex-wrap: wrap;
}
.merge-mode-desc {
  font-size: 12px;
  color: #94a3b8;
}
.merge-body {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.merge-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.merge-section-title {
  font-size: 13px;
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 8px;
}
.merge-section-tip {
  font-size: 11px;
  color: #94a3b8;
}
.merge-section-tip-inline {
  font-size: 11px;
  color: #94a3b8;
  font-weight: normal;
}
.merge-section-count {
  font-size: 11px;
  color: #94a3b8;
  font-weight: normal;
}
.merge-pack-list {
  max-height: 38vh;
  overflow-y: auto;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 4px;
  background: #f8fafc;
}
.merge-pack-list--radio {
  background: #ffffff;
}
.merge-pack-item {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background 0.15s ease;
}
.merge-pack-item:hover {
  background: #f1f5f9;
}
.merge-pack-item :deep(.el-checkbox) {
  width: 100%;
  height: auto;
  margin-right: 0;
}
.merge-pack-item :deep(.el-radio) {
  width: 100%;
  height: auto;
  margin-right: 0;
}
.merge-pack-item :deep(.el-checkbox__label),
.merge-pack-item :deep(.el-radio__label) {
  padding-left: 8px;
  width: 100%;
}
.merge-pack-info {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
}
.merge-pack-ic {
  color: #6366f1;
  flex-shrink: 0;
}
.merge-pack-name {
  font-size: 13px;
  font-weight: 500;
  color: #1f2937;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.merge-pack-tag {
  margin: 0;
  flex-shrink: 0;
}
.merge-pack-count {
  margin-left: auto;
  font-size: 11px;
  color: #94a3b8;
  white-space: nowrap;
}
.merge-preview {
  font-size: 13px;
  color: #475569;
  background: #eff6ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  padding: 10px 12px;
}
.merge-preview b {
  color: #3b82f6;
}
.merge-empty-subs {
  font-size: 12px;
  color: #94a3b8;
  text-align: center;
  padding: 16px 0;
  border: 1px dashed #e2e8f0;
  border-radius: 10px;
}
.merge-note {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #92400e;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 8px;
  padding: 8px 10px;
}
.merge-note .el-icon {
  color: #d97706;
  flex-shrink: 0;
}

/* 深色主题适配 */
:root[data-theme='dark'] .merge-mode-bar {
  border-color: #2e3650;
}
:root[data-theme='dark'] .merge-mode-desc {
  color: #6b7280;
}
:root[data-theme='dark'] .merge-section-title {
  color: #e5e7eb;
}
:root[data-theme='dark'] .merge-pack-list {
  background: #1a2030;
  border-color: #2e3650;
}
:root[data-theme='dark'] .merge-pack-list--radio {
  background: #1e2434;
}
:root[data-theme='dark'] .merge-pack-item:hover {
  background: #222838;
}
:root[data-theme='dark'] .merge-pack-ic {
  color: #818cf8;
}
:root[data-theme='dark'] .merge-pack-name {
  color: #e5e7eb;
}
:root[data-theme='dark'] .merge-preview {
  background: rgba(59, 130, 246, 0.12);
  border-color: #2e3650;
  color: #9ca3af;
}
:root[data-theme='dark'] .merge-preview b {
  color: #60a5fa;
}
:root[data-theme='dark'] .merge-empty-subs {
  color: #64748b;
  border-color: #2e3650;
}
:root[data-theme='dark'] .merge-note {
  background: rgba(217, 119, 6, 0.12);
  border-color: rgba(217, 119, 6, 0.35);
  color: #fcd34d;
}
:root[data-theme='dark'] .merge-note .el-icon {
  color: #f59e0b;
}
</style>
