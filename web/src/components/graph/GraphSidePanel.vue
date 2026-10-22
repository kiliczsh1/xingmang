<template>
  <div class="graph-side-panel-wrap" :class="{ collapsed }">
    <aside class="graph-side-panel" :class="{ collapsed }">
      <button v-if="!collapsed" class="panel-handle" title="收起侧栏" @click="collapsed = true">
        <el-icon><ArrowLeft /></el-icon>
      </button>

      <template v-if="!collapsed">
        <div class="panel-tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            class="panel-tab"
            :class="{ active: activeTab === tab.key }"
            @click="activeTab = tab.key"
          >
            <el-icon><component :is="tab.icon" /></el-icon>
            <span>{{ tab.label }}</span>
            <span v-if="tab.badge" class="panel-tab-badge">{{ tab.badge }}</span>
          </button>
        </div>

        <div class="panel-content">
          <GraphVersionList
            v-show="activeTab === 'versions'"
            :versions="store.versions"
            :current-version-id="store.currentVersionId"
            :book-id="store.bookId"
            :switching-id="switchingId"
            @switch-version="switchVersion"
            @rename-version="renameVersion"
            @delete-version="deleteVersion"
            @export-version="exportVersion"
            @create-version="createBlankVersion"
            @save-as-version="saveAsNewVersion"
          />
          <WikiFileImporter
            v-show="activeTab === 'import'"
            :book-id="store.bookId"
            @import="importFile"
          />
        </div>
      </template>
    </aside>

    <button v-if="collapsed" class="panel-expand" title="展开侧栏" @click="collapsed = false">
      <el-icon><ArrowRight /></el-icon>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowLeft, ArrowRight, UploadFilled, Share } from '@element-plus/icons-vue'
import GraphVersionList from '@/components/wiki/GraphVersionList.vue'
import WikiFileImporter from '@/components/wiki/WikiFileImporter.vue'
import { useGraphVersionActions } from '@/composables/useGraphVersionActions'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'

const store = useKnowledgeGraphStore()
const {
  switchingId,
  switchVersion,
  renameVersion,
  deleteVersion,
  exportVersion,
  saveAsNewVersion,
  createBlankVersion,
  importFile
} = useGraphVersionActions()

const props = withDefaults(defineProps<{ initialTab?: 'versions' | 'import' }>(), {
  initialTab: 'versions'
})

const activeTab = ref<'versions' | 'import'>(props.initialTab)
const collapsed = ref(true)

const tabs = computed(() => [
  { key: 'versions' as const, label: '版本', icon: Share, badge: store.versions.length || '' },
  { key: 'import' as const, label: '导入', icon: UploadFilled, badge: '' }
])
</script>

<style scoped>
.graph-side-panel-wrap {
  position: relative;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-shrink: 0;
}

.graph-side-panel-wrap.collapsed {
  width: 0;
}

.graph-side-panel {
  position: relative;
  width: 260px;
  min-width: 260px;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  border-right: 1px solid var(--el-border-color-lighter);
  transition: width 0.2s ease, min-width 0.2s ease;
  overflow: hidden;
}

.graph-side-panel.collapsed {
  width: 0;
  min-width: 0;
  border-right: none;
}

.panel-handle {
  position: absolute;
  top: 50%;
  right: -10px;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--el-border-color);
  border-radius: 50%;
  background: var(--el-bg-color);
  color: var(--el-text-color-secondary);
  cursor: pointer;
  z-index: 4;
  transition: all 0.2s ease;
}

.panel-handle:hover {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}

.panel-handle .el-icon {
  font-size: 12px;
}

/* 完全收起后，画布左侧露出一小块展开按钮，便于重新打开 */
.panel-expand {
  position: absolute;
  top: 50%;
  left: 6px;
  transform: translateY(-50%);
  width: 28px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  background: var(--el-bg-color);
  color: var(--el-text-color-secondary);
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  z-index: 5;
  transition: all 0.2s ease;
}

.panel-expand:hover {
  border-color: var(--el-color-primary);
  color: var(--el-color-primary);
}

.panel-expand .el-icon {
  font-size: 14px;
}

.panel-tabs {
  display: flex;
  flex-shrink: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.panel-tab {
  flex: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 9px 6px;
  border: none;
  border-bottom: 2px solid transparent;
  background: transparent;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  transition: all 0.18s ease;
}

.panel-tab:hover {
  color: var(--el-color-primary);
  background: var(--el-fill-color-light);
}

.panel-tab.active {
  color: var(--el-color-primary);
  border-bottom-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  font-weight: 500;
}

.panel-tab .el-icon {
  font-size: 13px;
}

.panel-tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 15px;
  padding: 0 5px;
  border-radius: 8px;
  background: var(--el-color-primary-light-8);
  color: var(--el-color-primary);
  font-size: 10px;
  line-height: 1;
}

.panel-content {
  flex: 1;
  min-height: 0;
  padding: 10px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-content > * {
  flex: 1;
  min-height: 0;
}

:root[data-theme='dark'] .graph-side-panel {
  border-right-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .panel-tabs {
  border-bottom-color: rgba(42, 50, 66, 0.4);
}

:root[data-theme='dark'] .panel-handle,
:root[data-theme='dark'] .panel-expand {
  background: #161b22;
  border-color: rgba(42, 50, 66, 0.6);
  color: #9ca3af;
}
</style>
