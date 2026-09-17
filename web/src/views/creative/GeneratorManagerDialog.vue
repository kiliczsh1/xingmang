<template>
  <el-dialog
    :model-value="visible"
    title="生成器管理"
    width="960px"
    destroy-on-close
    class="generator-manager-dialog"
    modal-class="generator-manager-dialog-modal"
    append-to-body
    @update:model-value="emit('update:visible', $event)"
  >
    <div class="generator-manager-content">
      <!-- 左侧：生成器列表 -->
      <div class="generator-list-panel">
        <div class="panel-header">
          <span class="panel-title">生成器列表</span>
          <el-button type="primary" size="small" @click="createNewGenerator">
            <el-icon><Plus /></el-icon>
            新建
          </el-button>
        </div>
        <div class="generator-list">
          <div
            v-for="(gen, index) in generators"
            :key="gen.id"
            class="generator-item"
            :class="{ active: selectedGenerator?.id === gen.id }"
            @click="selectGenerator(gen)"
          >
            <div class="generator-item-icon">
              <el-icon :size="18">
                <component :is="getIconComponent(gen.icon)" />
              </el-icon>
            </div>
            <div class="generator-item-info">
              <div class="generator-item-name-row">
                <span class="generator-item-name">{{ gen.name }}</span>
                <el-tag v-if="gen.isDefault" size="small" type="info" class="generator-item-default">默认</el-tag>
              </div>
            </div>
            <div class="generator-item-actions" @click.stop>
              <el-button
                type="primary"
                link
                size="small"
                :disabled="index === 0"
                @click="emit('move', 'up', index)"
              >
                <el-icon><ArrowUp /></el-icon>
              </el-button>
              <el-button
                type="primary"
                link
                size="small"
                :disabled="index === generators.length - 1"
                @click="emit('move', 'down', index)"
              >
                <el-icon><ArrowDown /></el-icon>
              </el-button>
              <el-button type="danger" link size="small" @click="emit('delete', gen)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </div>
          <div v-if="generators.length === 0" class="empty-list">
            <el-icon :size="32"><Document /></el-icon>
            <p>暂无生成器</p>
          </div>
        </div>
      </div>

      <!-- 右侧：生成器详情（预览 / 编辑 双模式，管理即浏览） -->
      <div class="generator-detail-panel">
        <template v-if="selectedGenerator">
          <!-- 预览模式：安静地看一张卡 -->
          <div v-if="!managerEditing" class="gm-preview">
            <div class="gm-pv-top">
              <span class="gm-pv-ic">
                <el-icon :size="24"><component :is="getIconComponent(selectedGenerator.icon)" /></el-icon>
              </span>
              <div class="gm-pv-info">
                <div class="gm-pv-name">{{ selectedGenerator.name }}</div>
                <div class="gm-pv-cat">
                  {{ getGeneratorCategory(selectedGenerator) }}
                  <el-tag v-if="selectedGenerator.isDefault" size="small" type="info" class="gm-pv-default">默认</el-tag>
                </div>
              </div>
            </div>
            <div class="gm-pv-sample">
              <span class="gm-pv-label">输出示例 · 卡片上将展示此内容</span>
              <span class="gm-pv-sample-text">{{ getGeneratorExample(selectedGenerator) }}</span>
            </div>
            <p class="gm-pv-desc">{{ selectedGenerator.description || '暂无功能描述' }}</p>
            <div class="gm-pv-core">
              <span class="gm-pv-label">核心提示词</span>
              <div class="gm-pv-core-text">{{ selectedGenerator.core_prompt || '暂无' }}</div>
            </div>
            <div class="gm-pv-actions">
              <el-button size="small" type="danger" plain @click="emit('delete', selectedGenerator)">删除</el-button>
              <el-button type="primary" size="small" @click="managerEditing = true">✎ 编辑</el-button>
            </div>
          </div>
          <!-- 编辑模式：聚焦式最小表单 -->
          <div v-else class="detail-content">
            <div class="detail-section">
              <label class="detail-label">生成器名称</label>
              <el-input v-model="selectedGenerator.name" placeholder="请输入生成器名称" />
            </div>
            <div class="detail-section">
              <div class="detail-label">选择图标</div>
              <div class="icon-picker-strip">
                <div
                  v-for="icon in availableIcons"
                  :key="icon.name"
                  class="icon-picker-item"
                  :class="{ selected: selectedGenerator.icon === icon.name }"
                  @click="selectedGenerator.icon = icon.name"
                >
                  <el-icon :size="18"><component :is="icon.component" /></el-icon>
                </div>
              </div>
            </div>
            <div class="detail-section">
              <label class="detail-label">功能描述</label>
              <el-input
                v-model="selectedGenerator.description"
                type="textarea"
                :rows="2"
                placeholder="请输入功能描述"
              />
            </div>
            <div class="detail-section">
              <label class="detail-label">输出示例</label>
              <el-input
                v-model="selectedGenerator.example"
                type="textarea"
                :rows="2"
                placeholder="卡片与列表将直接展示这段示例文案…"
              />
              <div class="field-description">用于抽卡区卡片「识别优于回忆」展示，留空则使用默认示例</div>
            </div>
            <div class="detail-section">
              <label class="detail-label">核心提示词</label>
              <el-input
                v-model="selectedGenerator.core_prompt"
                type="textarea"
                :rows="5"
                placeholder="请输入核心提示词，作为AI的system层指令..."
              />
            </div>
            <div class="detail-section">
              <label class="detail-label">备注</label>
              <el-input
                v-model="selectedGenerator.remark"
                type="textarea"
                :rows="1"
                placeholder="备注信息（可选）"
              />
            </div>
            <div class="detail-actions">
              <el-button @click="cancelEditGenerator">取消</el-button>
              <el-button type="primary" @click="handleSave">保存</el-button>
            </div>
          </div>
        </template>
        <div v-else class="empty-detail">
          <el-icon :size="48"><Document /></el-icon>
          <p>选择或创建一个生成器</p>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Plus,
  Delete,
  ArrowUp,
  ArrowDown,
  Document,
  MagicStick,
  EditPen,
  Reading,
  Tickets,
  Lightning,
  Cpu,
  CollectionTag,
  User,
  OfficeBuilding,
  Star,
  Sunny,
  Moon,
  Picture,
  Timer,
  TrophyBase,
  Medal,
  Coin,
  DataLine,
  TrendCharts,
  Promotion,
  Clock,
  Brush,
  Setting,
  Grid
} from '@element-plus/icons-vue'
// 生成器类型（与父组件保持一致）
type GeneratorItem = {
  id: number
  name: string
  description: string
  icon: string
  core_prompt?: string
  remark?: string
  order_num?: number
  isDefault?: boolean
  isCustom?: boolean
  isNew?: boolean
  example?: string
  tutorial_url?: string
}

const props = defineProps<{
  visible: boolean
  generators: GeneratorItem[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'save', payload: { generator: Partial<GeneratorItem>; isNew: boolean }): void
  (e: 'delete', gen: Partial<GeneratorItem>): void
  (e: 'move', kind: 'up' | 'down', index: number): void
}>()

const selectedGenerator = ref<Partial<GeneratorItem> | null>(null)
const isNewGenerator = ref(false)
const managerEditing = ref(false)

// 图标映射表：将图标名称映射到 Element Plus 图标组件
const iconMap: Record<string, any> = {
  MagicStick,
  Document,
  EditPen,
  Reading,
  Tickets,
  Lightning,
  Cpu,
  CollectionTag,
  User,
  OfficeBuilding,
  Star,
  Sunny,
  Moon,
  Picture,
  Timer,
  TrophyBase,
  Medal,
  Coin,
  DataLine,
  TrendCharts,
  Promotion,
  Clock,
  Brush,
  Setting,
  Grid
}

const getIconComponent = (iconName?: string) => {
  return (iconName && iconMap[iconName]) || MagicStick
}

// 可选择的图标列表（用于图标选择器）
const availableIcons = computed(() => {
  return Object.keys(iconMap).map(name => ({
    name,
    component: iconMap[name]
  }))
})

// 默认生成器示例输出（卡片直接展示"生成出来长什么样"）
const DEFAULT_GENERATOR_EXAMPLES: Record<number, string> = {
  1: '「一个每天醒来都会重置回同一天的渔村，只有主角记得。」',
  2: '《开局被雷劈，我觉醒创世能力》',
  3: '「当所有人都以为她是废物，她默默翻开了第 99 卷禁书……」',
  4: '卷一·觉醒 → 卷二·势成 → 卷三·逆命',
  5: '第3章 冲突升级 → 第4章 反转 → 第5章 高潮',
  6: '开篇 800 字即遇生死危机，钩子拉满',
  7: '「万象神瞳」— 看破一切伪象',
  8: '轩辕 · 白芷 · 天机阁 · 苍梧剑宗',
  9: '苏瑾 · 孤傲天才 · 秘密：身世成谜',
  10: '九州大陆 · 灵气复苏 · 三界格局'
}

// 默认生成器分类（用于标签语义）
const DEFAULT_GENERATOR_CATEGORIES: Record<number, string> = {
  1: '灵感', 2: '命名', 3: '灵感', 4: '结构', 5: '结构',
  6: '灵感', 7: '设定', 8: '命名', 9: '设定', 10: '设定'
}

const getGeneratorCategory = (generator: Partial<GeneratorItem>) => {
  const cat = generator.id != null ? DEFAULT_GENERATOR_CATEGORIES[generator.id] : undefined
  return generator.isDefault ? (cat || '创作') : '自定义'
}

const getExampleOverrides = (): Record<string, string> => {
  try { return JSON.parse(localStorage.getItem('creative-generator-examples') || '{}') } catch { return {} }
}

const getGeneratorExample = (generator: Partial<GeneratorItem>) => {
  // 1) 用户覆盖的示例优先
  const override = generator.id != null ? getExampleOverrides()[String(generator.id)] : undefined
  if (override && override.trim()) return override
  // 2) 默认生成器的内置示例
  if (generator.isDefault && generator.id != null) {
    const example = DEFAULT_GENERATOR_EXAMPLES[generator.id]
    if (example) return example
  }
  // 3) 自定义生成器：退化为功能描述首句，保证卡片始终有可识别的内容
  return generator.description?.slice(0, 32) || '输入你的设定，开始创作'
}

// 打开时默认选中第一个生成器
watch(
  () => props.visible,
  (v) => {
    if (v && !selectedGenerator.value && props.generators.length > 0) {
      selectGenerator(props.generators[0])
    }
  }
)

// 保存/删除后同步选中状态：列表变化时校验选中项
const saving = ref(false)
watch(
  () => props.generators,
  (list) => {
    if (saving.value) {
      // 保存成功（列表已刷新）：退出编辑模式
      saving.value = false
      if (isNewGenerator.value) {
        selectedGenerator.value = null
      }
      managerEditing.value = false
      isNewGenerator.value = false
      return
    }
    const sel = selectedGenerator.value
    if (sel?.id != null && !list.some(g => g.id === sel.id)) {
      // 选中的生成器已被删除：清空选中
      selectedGenerator.value = null
      managerEditing.value = false
      isNewGenerator.value = false
    }
  },
  { deep: true }
)

const createNewGenerator = () => {
  selectedGenerator.value = {
    name: '',
    description: '',
    icon: 'Lightning',
    example: '',
    core_prompt: '',
    remark: '',
    order_num: 0,
    isDefault: false
  }
  isNewGenerator.value = true
  managerEditing.value = true
}

const selectGenerator = (gen: GeneratorItem) => {
  selectedGenerator.value = { ...gen }
  isNewGenerator.value = false
  managerEditing.value = false
}

const cancelEditGenerator = () => {
  // 新建中取消：回到空详情；编辑中取消：还原为原始值并回到预览
  if (isNewGenerator.value) {
    selectedGenerator.value = null
  } else if (selectedGenerator.value?.id != null) {
    const original = props.generators.find(g => g.id === selectedGenerator.value!.id)
    selectedGenerator.value = original ? { ...original } : null
  }
  managerEditing.value = false
  isNewGenerator.value = false
}

const handleSave = () => {
  if (!selectedGenerator.value?.name?.trim()) {
    ElMessage.warning('请输入生成器名称')
    return
  }
  saving.value = true
  emit('save', {
    generator: selectedGenerator.value,
    isNew: isNewGenerator.value
  })
}
</script>
