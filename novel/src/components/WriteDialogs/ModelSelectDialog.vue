<template>
  <el-dialog
    v-model="visible"
    title="选择 AI 模型"
    width="820px"
    destroy-on-close
    class="model-select-dialog"
    :close-on-click-modal="false"
  >
    <div class="model-dialog-body">
      <div class="model-dialog-list">
        <div class="model-list-title">推荐模型</div>
        <div
          v-for="model in enhancedApiConfigs"
          :key="model.id"
          class="model-item"
          :class="{ active: tempSelectedConfigId === model.id }"
          @click="tempSelectedConfigId = model.id"
        >
          <div class="model-item-left">
            <span class="model-item-icon">{{ model.icon }}</span>
            <span class="model-item-name">{{ model.name }}</span>
          </div>
          <div class="model-item-right">
            <el-tag v-if="model.badge" :type="model.badgeType" size="small" class="model-badge">{{ model.badge }}</el-tag>
            <el-tag type="success" size="small" class="model-flow-tag">流畅</el-tag>
          </div>
        </div>
        <div v-if="apiConfigs.length === 0" class="model-empty">暂无可用模型</div>
      </div>
      <div class="model-dialog-detail">
        <template v-if="selectedModelForDetail">
          <div class="model-detail-header">
            <span class="model-detail-icon">{{ selectedModelForDetail.icon }}</span>
            <span class="model-detail-name">{{ selectedModelForDetail.name }}</span>
          </div>
          <div class="model-detail-desc">{{ selectedModelForDetail.description }}</div>
          <div class="model-detail-section">
            <div class="model-detail-label">模型能力评级</div>
            <div class="model-rating-row">
              <div class="model-rating-item">
                <span class="model-rating-label">文采水平</span>
                <div class="model-rating-stars">
                  <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= selectedModelForDetail.ratings.creative }"></span>
                </div>
              </div>
              <div class="model-rating-item">
                <span class="model-rating-label">指令遵从</span>
                <div class="model-rating-stars">
                  <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= selectedModelForDetail.ratings.instruction }"></span>
                </div>
              </div>
              <div class="model-rating-item">
                <span class="model-rating-label">字数消耗</span>
                <div class="model-rating-stars">
                  <span v-for="i in 5" :key="i" class="star" :class="{ filled: i <= selectedModelForDetail.ratings.consumption }"></span>
                </div>
              </div>
            </div>
          </div>
          <div class="model-detail-section">
            <div class="model-detail-label">注意事项</div>
            <div class="model-detail-notice">{{ selectedModelForDetail.notice }}</div>
          </div>
        </template>
        <div v-else class="model-detail-empty">
          <el-icon :size="32"><InfoFilled /></el-icon>
          <span>请从左侧选择一个模型</span>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="model-dialog-footer">
        <div class="model-footer-left">
          <el-popover
            placement="top"
            :width="280"
            trigger="click"
            :teleported="true"
            :z-index="10010"
            popper-class="model-param-popover"
          >
            <template #reference>
              <el-button class="model-footer-btn">
                <el-icon><Connection /></el-icon>
                联想能力
              </el-button>
            </template>
            <div class="param-popover-content">
              <div class="param-popover-label">温度 (Temperature)</div>
              <div class="param-popover-value">{{ temperature.toFixed(2) }}</div>
              <el-slider
                v-model="temperature"
                :min="0"
                :max="2"
                :step="0.01"
                :show-tooltip="false"
              />
              <div class="param-popover-hint">值越高输出越随机，值越低输出越确定</div>
            </div>
          </el-popover>
          <el-popover
            placement="top"
            :width="280"
            trigger="click"
            :teleported="true"
            :z-index="10010"
            popper-class="model-param-popover"
          >
            <template #reference>
              <el-button class="model-footer-btn">
                <el-icon><Setting /></el-icon>
                思考预算
              </el-button>
            </template>
            <div class="param-popover-content">
              <div class="param-popover-label">Top P</div>
              <div class="param-popover-value">{{ topP.toFixed(2) }}</div>
              <el-slider
                v-model="topP"
                :min="0"
                :max="1"
                :step="0.01"
                :show-tooltip="false"
              />
              <div class="param-popover-hint">值越高采样范围越广，值越低输出越集中</div>
            </div>
          </el-popover>
        </div>
        <el-button type="primary" class="model-confirm-btn" @click="confirmModelSelect">使用此模型</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { InfoFilled, Connection, Setting } from '@element-plus/icons-vue'
import type { ApiConfig } from '@/types'

const visible = defineModel<boolean>('visible', { default: false })

const props = defineProps<{
  apiConfigs: ApiConfig[]
  selectedConfigId: number | undefined
  modelSelectMode: 'chat' | 'creative2'
}>()

const emit = defineEmits<{
  confirm: [payload: { configId: number; temperature: number; topP: number; mode: 'chat' | 'creative2' }]
}>()

const tempSelectedConfigId = ref<number>()
const temperature = defineModel<number>('temperature', { default: 0.7 })
const topP = defineModel<number>('topP', { default: 0.9 })

// 模型图标映射
const modelIconMap: Record<string, string> = {
  '细腻': '✦',
  '氛围': '✧',
  '智慧': '⚙',
  '豆包': '⌘',
  '思考': '💭',
  '灵光': '△',
  'deepseek': '◈',
  '奇想': '✪',
  'zhipu': '⊕',
  'glm': '⊕',
  'gemini': '✦',
  'gpt': '✦',
  'claude': '✦',
}

const defaultModelDescription = '通用AI模型，适用于大多数写作场景，具有良好的指令遵循能力和稳定的输出质量。'
const defaultModelNotice = '在关联内容较长或开启较高思考预算时，模型可能需要较长时间进行推理，极端情况下等待时间可能超过10分钟。如果不希望等待过久，可以将思考预算调至「极速」模式，以获得更快的响应。如果出现【无法生成该内容】的提示，通常是因为检测到敏感内容，建议适当替换相关词语，或临时切换其他模型继续生成。'

interface EnhancedConfig {
  id: number
  name: string
  icon: string
  description: string
  notice: string
  ratings: { creative: number; instruction: number; consumption: number }
  badge?: string
  badgeType?: 'danger' | 'warning' | 'info'
}

const enhancedApiConfigs = computed<EnhancedConfig[]>(() => {
  return props.apiConfigs.map(m => {
    const nameLower = m.name.toLowerCase()
    const iconKey = Object.keys(modelIconMap).find(k => nameLower.includes(k)) || ''
    const icon = modelIconMap[iconKey] || '✦'

    let ratings = { creative: 4, instruction: 4, consumption: 3 }
    if (nameLower.includes('细腻')) {
      ratings = { creative: 4, instruction: 5, consumption: 3 }
    } else if (nameLower.includes('氛围')) {
      ratings = { creative: 5, instruction: 3, consumption: 4 }
    } else if (nameLower.includes('智慧')) {
      ratings = { creative: 3, instruction: 5, consumption: 3 }
    } else if (nameLower.includes('思考')) {
      ratings = { creative: 4, instruction: 5, consumption: 4 }
    } else if (nameLower.includes('deepseek')) {
      ratings = { creative: 4, instruction: 4, consumption: 3 }
    } else if (nameLower.includes('gemini')) {
      ratings = { creative: 4, instruction: 4, consumption: 4 }
    }

    let badge = ''
    let badgeType: 'danger' | 'warning' | 'info' = 'info'
    if (nameLower.includes('hot') || nameLower.includes('热门')) {
      badge = 'hot'; badgeType = 'danger'
    } else if (nameLower.includes('free') || nameLower.includes('免费')) {
      badge = 'free'; badgeType = 'danger'
    } else if (nameLower.includes('new') || nameLower.includes('新')) {
      badge = 'new'; badgeType = 'danger'
    }

    return {
      id: m.id,
      name: m.name,
      icon,
      description: m.description || '',
      notice: m.description || '',
      ratings,
      badge,
      badgeType,
    }
  })
})

const selectedModelForDetail = computed(() => {
  const base = props.apiConfigs.find(m => m.id === tempSelectedConfigId.value)
  if (!base) return null
  const nameLower = base.name.toLowerCase()
  const iconKey = Object.keys(modelIconMap).find(k => nameLower.includes(k)) || ''
  const icon = modelIconMap[iconKey] || '✦'

  let ratings = { creative: 4, instruction: 4, consumption: 3 }
  if (nameLower.includes('细腻')) {
    ratings = { creative: 4, instruction: 5, consumption: 3 }
  } else if (nameLower.includes('氛围')) {
    ratings = { creative: 5, instruction: 3, consumption: 4 }
  } else if (nameLower.includes('智慧')) {
    ratings = { creative: 3, instruction: 5, consumption: 3 }
  } else if (nameLower.includes('思考')) {
    ratings = { creative: 4, instruction: 5, consumption: 4 }
  } else if (nameLower.includes('deepseek')) {
    ratings = { creative: 4, instruction: 4, consumption: 3 }
  } else if (nameLower.includes('gemini')) {
    ratings = { creative: 4, instruction: 4, consumption: 4 }
  }

  return {
    id: base.id,
    name: base.name,
    icon,
    description: base.description || '',
    notice: base.description || '',
    ratings,
  }
})

// 当弹窗打开时，根据 selectedConfigId 初始化临时选中
watch(visible, (val) => {
  if (val) {
    const currentId = props.selectedConfigId
    const defaultConfig = props.apiConfigs.find(c => c.is_default)
    const firstConfig = props.apiConfigs[0]

    let targetId: number | undefined
    if (currentId) {
      targetId = currentId
    } else if (defaultConfig) {
      targetId = defaultConfig.id
    } else if (firstConfig) {
      targetId = firstConfig.id
    }

    tempSelectedConfigId.value = targetId
  }
})

const confirmModelSelect = () => {
  if (tempSelectedConfigId.value) {
    emit('confirm', {
      configId: tempSelectedConfigId.value,
      temperature: temperature.value,
      topP: topP.value,
      mode: props.modelSelectMode,
    })
  }
  visible.value = false
}
</script>

<style scoped>
.model-select-dialog :deep(.el-dialog__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #e8e8e8;
}

.model-select-dialog :deep(.el-dialog__body) {
  padding: 0;
}

.model-select-dialog :deep(.el-dialog__footer) {
  padding: 12px 20px;
  border-top: 1px solid #e8e8e8;
}

.model-dialog-body {
  display: flex;
  min-height: 480px;
  max-height: 560px;
}

.model-dialog-list {
  width: 300px;
  min-width: 300px;
  border-right: 1px solid #e8e8e8;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.model-list-title {
  padding: 12px 16px;
  font-size: 13px;
  color: #999;
  font-weight: 500;
  border-bottom: 1px solid #e8e8e8;
}

.model-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-left: 3px solid transparent;
  gap: 8px;
}

.model-item:hover {
  background: #f5f7fa;
}

.model-item.active {
  background: #e8f4ff;
  border-left-color: #409eff;
}

.model-item-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.model-item-icon {
  font-size: 16px;
  width: 20px;
  text-align: center;
  flex-shrink: 0;
}

.model-item-name {
  font-size: 14px;
  color: #1f2937;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.model-item-right {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-shrink: 0;
}

.model-badge {
  font-size: 11px !important;
  padding: 0 6px !important;
  height: 20px !important;
  line-height: 20px !important;
}

.model-flow-tag {
  font-size: 11px !important;
  padding: 0 6px !important;
  height: 20px !important;
  line-height: 20px !important;
}

.model-empty {
  padding: 40px 16px;
  text-align: center;
  color: #999;
  font-size: 14px;
}

.model-dialog-detail {
  flex: 1;
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.model-detail-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
}

.model-detail-icon {
  font-size: 22px;
}

.model-detail-name {
  font-size: 22px;
  font-weight: 600;
  color: #1f2937;
}

.model-detail-desc {
  font-size: 14px;
  color: #666;
  line-height: 1.6;
  margin-bottom: 20px;
}

.model-detail-section {
  margin-bottom: 20px;
}

.model-detail-label {
  font-size: 12px;
  color: #999;
  margin-bottom: 10px;
}

.model-rating-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.model-rating-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.model-rating-label {
  font-size: 13px;
  color: #333;
  width: 60px;
  flex-shrink: 0;
}

.model-rating-stars {
  display: flex;
  gap: 4px;
}

.star {
  width: 16px;
  height: 16px;
  border-radius: 3px;
  background: #e0e0e0;
  transition: background 0.2s;
}

.star.filled {
  background: #f5c842;
}

.model-detail-notice {
  font-size: 13px;
  color: #666;
  line-height: 1.8;
  padding: 14px 16px;
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 8px;
}

.model-detail-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #999;
  font-size: 14px;
}

.model-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.model-footer-left {
  display: flex;
  gap: 10px;
}

.model-footer-btn {
  border-radius: 20px !important;
  padding: 8px 18px !important;
}

.model-confirm-btn {
  border-radius: 6px !important;
  padding: 10px 28px !important;
  font-size: 15px !important;
  background: #52c41a !important;
  border-color: #52c41a !important;
}

.model-confirm-btn:hover {
  background: #45a818 !important;
  border-color: #45a818 !important;
}

:deep(.model-param-popover) {
  padding: 16px !important;
  border-radius: 8px !important;
}

.param-popover-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.param-popover-label {
  font-size: 13px;
  font-weight: 500;
  color: #333;
}

.param-popover-value {
  font-size: 18px;
  font-weight: 600;
  color: #409eff;
  text-align: center;
}

.param-popover-hint {
  font-size: 11px;
  color: #999;
  text-align: center;
}

:deep(.model-param-popover .el-slider__runway) {
  margin: 8px 0;
}

/* Dark theme */
:root[data-theme='dark'] .model-select-dialog :deep(.el-dialog__header) {
  border-bottom-color: rgba(255, 255, 255, 0.14);
}

:root[data-theme='dark'] .model-select-dialog :deep(.el-dialog__footer) {
  border-top-color: rgba(255, 255, 255, 0.14);
}

:root[data-theme='dark'] .model-dialog-list {
  border-right-color: rgba(255, 255, 255, 0.14);
}

:root[data-theme='dark'] .model-list-title {
  border-bottom-color: rgba(255, 255, 255, 0.14);
  color: #6b7280;
}

:root[data-theme='dark'] .model-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

:root[data-theme='dark'] .model-item.active {
  background: rgba(64, 158, 255, 0.12);
  border-left-color: #409eff;
}

:root[data-theme='dark'] .model-item-name {
  color: #e5e7eb;
}

:root[data-theme='dark'] .model-detail-name {
  color: #e5e7eb;
}

:root[data-theme='dark'] .model-detail-desc {
  color: #9ca3af;
}

:root[data-theme='dark'] .model-detail-label {
  color: #6b7280;
}

:root[data-theme='dark'] .model-rating-label {
  color: #d1d5db;
}

:root[data-theme='dark'] .star {
  background: rgba(255, 255, 255, 0.12);
}

:root[data-theme='dark'] .star.filled {
  background: #f5c842;
}

:root[data-theme='dark'] .model-detail-notice {
  color: #9ca3af;
  background: rgba(255, 255, 255, 0.04);
  border-color: rgba(255, 255, 255, 0.14);
}

:root[data-theme='dark'] .model-detail-empty {
  color: #6b7280;
}

:root[data-theme='dark'] .model-empty {
  color: #6b7280;
}

:root[data-theme='dark'] .param-popover-label {
  color: #d1d5db;
}

:root[data-theme='dark'] .param-popover-hint {
  color: #6b7280;
}
</style>
