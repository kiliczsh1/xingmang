<template>
  <div class="xm-ref-panel">
    <div class="xm-ref-head">
      <span class="xm-ref-title">资料栏</span>
      <button type="button" class="xm-ref-close" @click="emit('close')">
        <el-icon><Close /></el-icon>
      </button>
    </div>
    <div class="xm-ref-body">
      <el-tabs model-value="characters">
        <el-tab-pane label="角色卡" name="characters">
          <div v-if="!characters.length" class="xm-ref-empty">暂无角色，可在角色库中创建</div>
          <div v-for="c in characters" :key="c.id" class="xm-ref-card" @click="emit('insert', c.name)">
            <div class="xm-ref-card-head">
              <span class="xm-ref-avatar">{{ c.name?.slice(0, 1) || '角' }}</span>
              <span class="xm-ref-card-name">{{ c.name }}</span>
              <span class="xm-ref-card-folder">{{ c.folder || '未归档' }}</span>
            </div>
            <p v-if="c.personality" class="xm-ref-card-text">{{ c.personality }}</p>
            <p class="xm-ref-card-tip">点击将姓名插入正文光标处</p>
          </div>
        </el-tab-pane>
        <el-tab-pane label="词条卡" name="terms">
          <div v-if="!terms.length" class="xm-ref-empty">暂无词条，可在词条库中创建</div>
          <div v-for="(t, i) in terms" :key="i" class="xm-ref-card" @click="emit('insert', t.name)">
            <div class="xm-ref-card-head">
              <span class="xm-ref-card-name">{{ t.name }}</span>
              <span v-if="t.keys?.length" class="xm-ref-card-folder">{{ t.keys.join(' / ') }}</span>
            </div>
            <p v-if="t.content" class="xm-ref-card-text">{{ t.content.slice(0, 80) }}{{ t.content.length > 80 ? '…' : '' }}</p>
            <p class="xm-ref-card-tip">点击将词条名插入正文光标处</p>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Close } from '@element-plus/icons-vue'
import type { Character } from '@/types'

defineOptions({ name: 'ReferencePanel' })

const props = defineProps<{
  characters: Character[]
  /** 词条：名称 + 关键词 + 内容 */
  terms: Array<{ name: string; keys: string[]; content: string }>
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'insert', text: string): void
}>()
</script>

<style scoped>
.xm-ref-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #ffffff;
  border-left: 1px solid #e4e2dd;
}

.xm-ref-head {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-bottom: 1px solid #eeeeea;
  flex-shrink: 0;
}

.xm-ref-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: #1f2329;
}

.xm-ref-close {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: 5px;
  color: #8a8f99;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.xm-ref-close:hover {
  background: rgba(0, 0, 0, 0.06);
  color: #374151;
}

.xm-ref-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 8px 12px;
}

.xm-ref-empty {
  font-size: 12px;
  color: #a3a8b3;
  text-align: center;
  padding: 32px 0;
}

.xm-ref-card {
  border: 1px solid #eef0f2;
  border-radius: 8px;
  padding: 9px 11px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.12s ease;
}

.xm-ref-card:hover {
  border-color: rgba(79, 110, 247, 0.45);
  background: #f7f9ff;
}

.xm-ref-card-head {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.xm-ref-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1, #a855f7);
  color: #fff;
  font-size: 12px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.xm-ref-card-name {
  font-size: 13px;
  font-weight: 600;
  color: #1f2329;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-ref-card-folder {
  flex-shrink: 0;
  font-size: 10.5px;
  color: #a3a8b3;
  background: #f3f5f7;
  border-radius: 4px;
  padding: 1px 6px;
  max-width: 110px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.xm-ref-card-text {
  margin: 6px 0 0;
  font-size: 11.5px;
  line-height: 1.6;
  color: #6b7280;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.xm-ref-card-tip {
  margin: 5px 0 0;
  font-size: 10.5px;
  color: #b8c0cc;
}

:root[data-theme='dark'] .xm-ref-panel {
  background: rgba(22, 27, 34, 0.92);
  border-left-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-ref-title,
:root[data-theme='dark'] .xm-ref-card-name {
  color: #e6edf3;
}

:root[data-theme='dark'] .xm-ref-head {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-ref-card {
  border-color: #3f4a5c;
}

:root[data-theme='dark'] .xm-ref-card:hover {
  border-color: rgba(79, 110, 247, 0.6);
  background: rgba(59, 130, 246, 0.08);
}

:root[data-theme='dark'] .xm-ref-card-text {
  color: #8b98a8;
}

:root[data-theme='dark'] .xm-ref-card-folder,
:root[data-theme='dark'] .xm-ref-empty,
:root[data-theme='dark'] .xm-ref-card-tip {
  color: #6b7688;
  background: transparent;
}

:root[data-theme='dark'] .xm-ref-close:hover {
  background: rgba(255, 255, 255, 0.08);
}
</style>
