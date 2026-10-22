<template>
  <el-drawer
    v-model="visible"
    title="设置"
    size="380px"
    append-to-body
    class="write-settings-drawer xm-drawer"
  >
    <div class="wsd-body">
      <!-- 字体 -->
      <div class="wsd-section">
        <div class="wsd-section-title">字体</div>
        <div class="wsd-row">
          <span class="wsd-label">字体</span>
          <FontSelector
            :model-value="fontFamily"
            @update:model-value="emit('font-family-change', $event)"
            @change="emit('font-family-change', $event)"
          />
        </div>
        <div class="wsd-row">
          <span class="wsd-label">字号</span>
          <FontSizeSelector
            :model-value="fontSize"
            @update:model-value="emit('font-size-change', $event)"
            @change="emit('font-size-change', $event)"
          />
        </div>
      </div>

      <!-- 段落 -->
      <div class="wsd-section">
        <div class="wsd-section-title">段落</div>
        <div class="wsd-row">
          <span class="wsd-label">段首缩进</span>
          <el-switch
            :model-value="textIndent"
            @update:model-value="emit('text-indent-change', $event)"
          />
        </div>
        <div class="wsd-row">
          <span class="wsd-label">段间距</span>
          <el-slider
            :model-value="paragraphSpacing"
            :min="0"
            :max="40"
            :step="2"
            show-input
            input-size="small"
            @update:model-value="emit('paragraph-spacing-change', $event)"
          />
        </div>
      </div>

      <!-- 文本横线 -->
      <div class="wsd-section">
        <div class="wsd-section-title">文本横线</div>
        <div class="wsd-row">
          <span class="wsd-label">段落下划线</span>
          <el-switch
            :model-value="paragraphLine"
            @update:model-value="emit('paragraph-line-change', $event)"
          />
        </div>
        <div class="wsd-tip">每个段落底部显示一条分隔横线，便于区分段落。</div>
      </div>

      <!-- 正文标记 -->
      <div class="wsd-section">
        <div class="wsd-section-title">正文标记</div>
        <div class="wsd-row">
          <span class="wsd-label">角色名高亮</span>
          <el-switch
            :model-value="roleMarkEnabled"
            @update:model-value="emit('role-mark-change', $event)"
          />
        </div>
        <div class="wsd-row">
          <span class="wsd-label">词条名高亮</span>
          <el-switch
            :model-value="termMarkEnabled"
            @update:model-value="emit('term-mark-change', $event)"
          />
        </div>
        <div class="wsd-tip">正文中已创建的角色名 / 设定词条会高亮显示，点击可查看资料卡。</div>
      </div>

      <!-- 外观 -->
      <div class="wsd-section">
        <div class="wsd-section-title">外观</div>
        <div class="wsd-row">
          <span class="wsd-label">主题</span>
          <el-radio-group
            :model-value="themeMode"
            @update:model-value="emit('theme-mode-change', $event)"
          >
            <el-radio-button value="light">浅色</el-radio-button>
            <el-radio-button value="dark">深色</el-radio-button>
            <el-radio-button value="system">跟随系统</el-radio-button>
          </el-radio-group>
        </div>
        <div class="wsd-row">
          <span class="wsd-label">编辑器预览</span>
          <div class="wsd-preview">
            <span class="wsd-preview-role">林晚</span>推开窗，
            <span class="wsd-preview-term">星轨</span>在夜空中缓缓转动。
          </div>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import FontSelector from '@/components/FontSelector.vue'
import FontSizeSelector from '@/components/FontSizeSelector.vue'

const props = defineProps<{
  visible: boolean
  fontFamily: string
  fontSize: number
  textIndent: boolean
  paragraphSpacing: number
  paragraphLine: boolean
  roleMarkEnabled: boolean
  termMarkEnabled: boolean
  themeMode: 'light' | 'dark' | 'system'
}>()

const emit = defineEmits<{
  (e: 'update:visible', v: boolean): void
  (e: 'font-family-change', v: string): void
  (e: 'font-size-change', v: number): void
  (e: 'text-indent-change', v: boolean): void
  (e: 'paragraph-spacing-change', v: number): void
  (e: 'paragraph-line-change', v: boolean): void
  (e: 'role-mark-change', v: boolean): void
  (e: 'term-mark-change', v: boolean): void
  (e: 'theme-mode-change', v: 'light' | 'dark' | 'system'): void
}>()

const visible = computed({
  get: () => props.visible,
  set: (v: boolean) => emit('update:visible', v),
})
</script>

<style scoped>
.wsd-body {
  display: flex;
  flex-direction: column;
  gap: 22px;
  padding: 4px 2px 24px;
}
.wsd-section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.wsd-section-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--xm-ink-3, #9d9c96);
  letter-spacing: .06em;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--xm-line, #f0efe9);
}
.wsd-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.wsd-label {
  font-size: 13px;
  color: var(--xm-ink-2, #555);
  flex-shrink: 0;
}
/* 业务语义：暗色正文标签保留柔和绿调特调，不走令牌翻转 */
:root[data-theme='dark'] .wsd-label {
  color: #c8d6cd;
}
.wsd-tip {
  font-size: 11.5px;
  color: var(--xm-ink-3, #9d9c96);
  line-height: 1.6;
}
.wsd-preview {
  font-size: 13px;
  line-height: 1.9;
  color: var(--xm-ink-2, #555);
  background: var(--xm-surface-2, #faf9f6);
  border: 1px solid var(--xm-line, #e4e2dd);
  border-radius: var(--xm-radius-sm, 10px);
  padding: 8px 12px;
  flex: 1;
}
/* 业务语义：暗色预览文字保留柔和绿调特调；底/边已由令牌翻转 */
:root[data-theme='dark'] .wsd-preview {
  color: #c8d6cd;
}
.wsd-preview-role {
  color: #b45309;
  background: rgba(245, 158, 11, .14);
  border-radius: 3px;
  padding: 0 2px;
}
.wsd-preview-term {
  color: #1d4ed8;
  background: rgba(59, 130, 246, .12);
  border-radius: 3px;
  padding: 0 2px;
}
:root[data-theme='dark'] .wsd-preview-role {
  color: #fbbf24;
  background: rgba(245, 158, 11, .18);
}
:root[data-theme='dark'] .wsd-preview-term {
  color: #93c5fd;
  background: rgba(59, 130, 246, .16);
}
</style>
