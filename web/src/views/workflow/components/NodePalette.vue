<template>
  <aside class="node-palette">
    <div class="palette-head">
      <i class="fas fa-shapes" />
      <span>节点库</span>
      <button class="head-btn" :title="allCollapsed ? '展开全部分组' : '折叠全部分组'" @click="toggleAll">
        <i class="fas" :class="allCollapsed ? 'fa-angle-double-down' : 'fa-angle-double-up'" />
      </button>
    </div>

    <div class="palette-search">
      <i class="fas fa-search" />
      <input v-model="keyword" type="text" placeholder="搜索节点" spellcheck="false" />
      <button v-if="keyword" class="head-btn" title="清空" @click="keyword = ''">
        <i class="fas fa-times" />
      </button>
    </div>

    <div class="palette-body">
      <template v-for="g in groups" :key="g.name">
        <div class="palette-group" @click="toggleGroup(g.name)">
          <i class="fas" :class="collapsed.has(g.name) ? 'fa-chevron-right' : 'fa-chevron-down'" />
          <span class="group-name">{{ g.name }}</span>
          <span class="group-count">{{ g.items.length }}</span>
        </div>
        <div v-show="!collapsed.has(g.name)" class="palette-items">
          <div
            v-for="it in g.items"
            :key="it.type"
            class="palette-item"
            draggable="true"
            :title="it.description"
            @dragstart="onDragStart($event, it.type)"
            @dragend="onDragEnd"
            @dblclick="emit('add', it.type)"
          >
            <i class="fas item-icon" :class="it.icon" :style="{ color: it.color }" />
            <div class="item-text">
              <div class="item-label">
                {{ it.label }}
                <span v-if="it.origin === 'star'" class="item-origin">星域</span>
              </div>
              <div class="item-desc">{{ it.description }}</div>
            </div>
          </div>
        </div>
      </template>
      <div v-if="!groups.length" class="palette-empty">
        <i class="fas fa-search-minus" />
        <span>没有匹配的节点</span>
      </div>
    </div>

    <div class="palette-foot">拖入画布或双击添加</div>
  </aside>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { NODE_CATALOG, NODE_DND_MIME, NODE_GROUP_ORDER, type NodeGroup } from '@/services/workflow/nodeCatalog'
import type { NodeType } from '@/types/workflow'

const emit = defineEmits<{ (e: 'add', type: NodeType): void }>()

const COLLAPSE_KEY = 'wf-palette-collapsed-v1'

const keyword = ref('')
const collapsed = ref<Set<NodeGroup>>(new Set())

onMounted(() => {
  try {
    const raw = localStorage.getItem(COLLAPSE_KEY)
    if (raw) collapsed.value = new Set(JSON.parse(raw) as NodeGroup[])
  } catch {
    /* 存储不可用则用默认展开 */
  }
})

const persistCollapsed = () => {
  try {
    localStorage.setItem(COLLAPSE_KEY, JSON.stringify(Array.from(collapsed.value)))
  } catch {
    /* 忽略 */
  }
}

const groups = computed(() => {
  const kw = keyword.value.trim().toLowerCase()
  const matched = kw
    ? NODE_CATALOG.filter(
        i =>
          i.label.toLowerCase().includes(kw) ||
          i.type.toLowerCase().includes(kw) ||
          i.description.toLowerCase().includes(kw)
      )
    : NODE_CATALOG
  return NODE_GROUP_ORDER.map(name => ({
    name,
    items: matched.filter(i => i.group === name)
  })).filter(g => g.items.length > 0)
})

const allCollapsed = computed(
  () => groups.value.length > 0 && groups.value.every(g => collapsed.value.has(g.name))
)

const toggleGroup = (name: NodeGroup) => {
  const next = new Set(collapsed.value)
  if (next.has(name)) next.delete(name)
  else next.add(name)
  collapsed.value = next
  persistCollapsed()
}

const collapseAll = () => {
  if (allCollapsed.value) return new Set<NodeGroup>()
  return new Set<NodeGroup>(groups.value.map(g => g.name) as NodeGroup[])
}

const toggleAll = () => {
  collapsed.value = collapseAll()
  persistCollapsed()
}

const onDragStart = (e: DragEvent, type: NodeType) => {
  if (!e.dataTransfer) return
  e.dataTransfer.setData(NODE_DND_MIME, type)
  e.dataTransfer.effectAllowed = 'copy'
}

const onDragEnd = () => {
  /* 拖拽结束由画布自行收起高亮，这里无需处理 */
}
</script>

<style scoped>
.node-palette {
  width: 208px;
  flex-shrink: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--wf-el-bg, #ffffff);
  border-right: 1px solid var(--wf-el-border, #dcdfe6);
  user-select: none;
}
.palette-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  font-size: 12px;
  font-weight: 500;
  color: var(--wf-el-text-primary, #303133);
  border-bottom: 1px solid var(--wf-el-border-light, #ebeef5);
}
.palette-head > i {
  color: var(--el-color-primary, #409eff);
  font-size: 12px;
}
.head-btn {
  margin-left: auto;
  border: none;
  background: transparent;
  color: var(--wf-el-text-secondary, #909399);
  cursor: pointer;
  padding: 2px 4px;
  border-radius: 4px;
  font-size: 11px;
}
.head-btn:hover {
  background: var(--wf-el-fill, #f0f2f5);
  color: var(--wf-el-text-primary, #303133);
}
.palette-search {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 8px 10px;
  padding: 4px 8px;
  border: 1px solid var(--wf-el-border, #dcdfe6);
  border-radius: 6px;
  background: var(--wf-el-fill-lighter, #fafafa);
}
.palette-search > i {
  font-size: 11px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
}
.palette-search input {
  flex: 1;
  min-width: 0;
  border: none;
  outline: none;
  background: transparent;
  font-size: 12px;
  color: var(--wf-el-text-primary, #303133);
}
.palette-search input::placeholder {
  color: var(--wf-el-text-placeholder, #c0c4cc);
}
.palette-body {
  flex: 1;
  overflow-y: auto;
  padding: 0 6px 8px;
}
.palette-group {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 6px;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  cursor: pointer;
  border-radius: 4px;
}
.palette-group:hover {
  background: var(--wf-el-fill-light, #f5f7fa);
}
.palette-group > i {
  font-size: 9px;
  width: 10px;
}
.group-name {
  flex: 1;
}
.group-count {
  font-variant-numeric: tabular-nums;
  opacity: 0.7;
}
.palette-items {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding-bottom: 4px;
}
.palette-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: grab;
  transition: background 0.15s, border-color 0.15s;
}
.palette-item:hover {
  background: var(--wf-el-fill-light, #f5f7fa);
  border-color: var(--wf-el-border-light, #ebeef5);
}
.palette-item:active {
  cursor: grabbing;
}
.item-icon {
  font-size: 13px;
  width: 16px;
  text-align: center;
  flex-shrink: 0;
}
.item-text {
  min-width: 0;
}
.item-origin {
  margin-left: 4px;
  padding: 0 3px;
  border-radius: 2px;
  font-size: 9px;
  line-height: 14px;
  vertical-align: 1px;
  color: #7e57c2;
  background: rgba(126, 87, 194, 0.14);
}
.item-label {
  font-size: 12px;
  color: var(--wf-el-text-primary, #303133);
  line-height: 1.3;
}
.item-desc {
  font-size: 10px;
  color: var(--wf-el-text-secondary, #909399);
  line-height: 1.3;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.palette-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 24px 0;
  font-size: 11px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
}
.palette-empty i {
  font-size: 16px;
}
.palette-foot {
  padding: 6px 10px;
  border-top: 1px solid var(--wf-el-border-light, #ebeef5);
  font-size: 10px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
  text-align: center;
}
</style>
