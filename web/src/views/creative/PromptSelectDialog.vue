<template>
  <el-dialog
    :model-value="visible"
    width="800px"
    destroy-on-close
    class="prompt-select-dialog"
    modal-class="prompt-select-dialog-modal"
    top="5vh"
    append-to-body
    :show-close="false"
    @update:model-value="emit('update:visible', $event)"
  >
    <!-- 自定义标题栏 -->
    <template #header>
      <div class="ps-dialog-header">
        <div class="ps-dialog-title">
          <el-icon class="ps-title-icon"><CollectionTag /></el-icon>
          <span>{{ promptSelectDialogTitle }}</span>
        </div>
        <el-button
          type="primary"
          link
          size="small"
          class="ps-close-btn"
          @click="emit('update:visible', false)"
        >
          <el-icon :size="18"><Close /></el-icon>
        </el-button>
      </div>
    </template>
    <div class="prompt-select-content">
      <!-- 第一步：顶部标签导航 -->
      <div class="tab-navigation">
        <div
          v-for="tab in promptTabs"
          :key="tab.key"
          class="tab-item"
          :class="{ active: activePromptTab === tab.key }"
          @click="activePromptTab = tab.key"
        >
          {{ tab.label }}
        </div>
      </div>

      <!-- 搜索、筛选、排序区域 - 仅在"全部"和"已固定"标签页显示 -->
      <template v-if="activePromptTab === 'all' || activePromptTab === 'pinned'">
        <!-- 第二步：重新设计的搜索区域 -->
        <div class="search-section">
          <div class="search-box">
            <el-input
              v-model="promptSearchKeyword"
              placeholder="搜索提示词..."
              clearable
              class="search-input-field"
              @keyup.enter="handleSearchPrompts"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <button class="search-btn" @click="handleSearchPrompts">
              <el-icon><Search /></el-icon>
              搜索
            </button>
          </div>
        </div>

        <!-- 第三步：优化分类筛选 - 下拉选择器 + 标签云 -->
        <div class="filter-section">
          <div class="filter-header" @click="toggleCategoryDropdown">
            <el-icon><Filter /></el-icon>
            <span>分类筛选</span>
            <el-icon class="filter-arrow" :class="{ expanded: showCategoryDropdown }"><ArrowDown /></el-icon>
          </div>

          <div v-if="showCategoryDropdown" class="category-dropdown">
            <div
              v-for="category in promptCategories"
              :key="category"
              class="dropdown-item"
              :class="{
                selected: selectedPromptCategory === category || (category === '全部' && selectedPromptCategory === 'all')
              }"
              @click="selectCategory(category)"
            >
              {{ category }}
            </div>
          </div>

          <div v-if="selectedPromptCategory !== 'all' && selectedPromptCategory !== '全部'" class="selected-tags">
            <span class="tag-label">当前选择：</span>
            <el-tag
              closable
              type="success"
              size="small"
              @close="clearCategorySelection"
            >
              {{ selectedPromptCategory === 'all' ? '全部' : selectedPromptCategory }}
            </el-tag>
          </div>
        </div>

        <!-- 第四步：排序选项栏 -->
        <div class="sort-section">
          <div class="sort-label">
            <el-icon><Sort /></el-icon>
            <span>排序方式</span>
          </div>
          <div class="sort-options">
            <button
              v-for="option in sortOptions"
              :key="option.key"
              class="sort-btn"
              :class="{ active: activeSortOption === option.key }"
              @click="activeSortOption = option.key"
            >
              <el-icon v-if="option.icon"><component :is="option.icon" /></el-icon>
              {{ option.label }}
            </button>
          </div>
        </div>
      </template>

      <!-- 第五步+第六步：根据标签显示不同内容 -->

      <!-- 全部标签页：显示提示词列表 -->
      <template v-if="activePromptTab === 'all'">
        <div class="prompt-select-list">
          <div
            v-for="prompt in sortedAndFilteredPrompts"
            :key="prompt.id"
            class="prompt-card-item"
            :class="{ selected: isPromptPinned(prompt.id) }"
          >
            <div class="card-main" @click="emit('toggle-pin', prompt)">
              <div class="card-top-row">
                <div class="author-avatar" :style="{ background: getAvatarColor(prompt.author || prompt.name) }">
                  {{ (prompt.author || '匿').charAt(0) }}
                </div>
                <div class="card-info">
                  <div class="card-title-row">
                    <span class="card-title">{{ prompt.name }}</span>
                  </div>
                  <div class="card-meta-row">
                    <span class="author-name">{{ prompt.author || '匿名用户' }}</span>
                    <el-tag
                      v-if="prompt.is_featured"
                      size="small"
                      class="card-featured-tag"
                    >
                      精选
                    </el-tag>
                    <span class="author-stats">
                      <span class="stat-item">
                        <el-icon class="fire-icon"><Sunny /></el-icon>
                        {{ formatUsageCount(prompt.usage_count) }}
                      </span>
                    </span>
                    <span class="card-date">{{ formatPromptTime(prompt.created_at) }}</span>
                  </div>
                  <div class="card-tags">
                    <el-tag size="small" class="card-category-tag">{{ prompt.category }}</el-tag>
                    <el-tag size="small" class="card-version-tag">{{ prompt.version || 'v1.0' }}</el-tag>
                  </div>
                </div>
              </div>

              <div class="card-content">
                <p class="card-desc">{{ prompt.description || '1' }}</p>
              </div>
            </div>

            <div class="card-actions" @click.stop>
              <button
                v-if="isPromptPinned(prompt.id)"
                class="action-btn action-btn--remove"
                @click="emit('remove-pin', prompt.id)"
              >
                <el-icon><Close /></el-icon>
                删除
              </button>
              <button
                v-else
                class="action-btn action-btn--add"
                @click="emit('toggle-pin', prompt)"
              >
                <el-icon><Plus /></el-icon>
                添加
              </button>
            </div>
          </div>

          <div v-if="sortedAndFilteredPrompts.length === 0" class="empty-prompts">
            <el-icon :size="48"><Document /></el-icon>
            <p>暂无匹配的提示词</p>
            <p class="empty-hint">尝试调整筛选条件或搜索关键词</p>
          </div>
        </div>
      </template>

      <!-- 已固定标签页：显示已固定的提示词 -->
      <template v-else-if="activePromptTab === 'pinned'">
        <div class="prompt-select-list">
          <div
            v-for="prompt in pinnedPrompts"
            :key="prompt.id"
            class="prompt-card-item selected"
          >
            <div class="card-main">
              <div class="card-top-row">
                <div class="author-avatar" :style="{ background: getAvatarColor(prompt.author || prompt.name) }">
                  {{ (prompt.author || '匿').charAt(0) }}
                </div>
                <div class="card-info">
                  <div class="card-title-row">
                    <span class="card-title">{{ prompt.name }}</span>
                  </div>
                  <div class="card-meta-row">
                    <span class="author-name">{{ prompt.author || '匿名用户' }}</span>
                    <el-tag
                      v-if="prompt.is_featured"
                      size="small"
                      class="card-featured-tag"
                    >
                      精选
                    </el-tag>
                    <span class="author-stats">
                      <span class="stat-item">
                        <el-icon class="fire-icon"><Sunny /></el-icon>
                        {{ formatUsageCount(prompt.usage_count) }}
                      </span>
                    </span>
                    <span class="card-date">{{ formatPromptTime(prompt.created_at) }}</span>
                  </div>
                  <div class="card-tags">
                    <el-tag size="small" class="card-category-tag">{{ prompt.category }}</el-tag>
                    <el-tag size="small" class="card-version-tag">{{ prompt.version || 'v1.0' }}</el-tag>
                  </div>
                </div>
              </div>

              <div class="card-content">
                <p class="card-desc">{{ prompt.description || '1' }}</p>
              </div>
            </div>

            <div class="card-actions">
              <el-switch
                :model-value="prompt.enabled"
                size="small"
                active-text="启用"
                inactive-text="禁用"
                @change="emit('toggle-enabled', prompt.id)"
              />
              <button
                class="action-btn action-btn--remove"
                @click="emit('remove-pin', prompt.id)"
                style="margin-left: 8px;"
              >
                <el-icon><Close /></el-icon>
                移除
              </button>
            </div>
          </div>

          <div v-if="pinnedPrompts.length === 0" class="empty-prompts">
            <el-icon :size="48"><CollectionTag /></el-icon>
            <p>暂无固定的提示词</p>
            <p class="empty-hint">在"全部"标签页中添加提示词</p>
          </div>
        </div>
      </template>

      <!-- 已创建标签页：创建新提示词 -->
      <template v-else-if="activePromptTab === 'created'">
        <div class="create-prompt-section">
          <div class="create-prompt-form">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">提示词名称</label>
                <el-input
                  v-model="newPromptForm.name"
                  placeholder="输入提示词名称..."
                  clearable
                />
              </div>

              <div class="form-group">
                <label class="form-label">分类</label>
                <el-select
                  v-model="newPromptForm.category"
                  placeholder="选择分类"
                  style="width: 100%"
                  allow-create
                  filterable
                >
                  <el-option
                    v-for="cat in promptCategories.filter(c => c !== '全部')"
                    :key="cat"
                    :label="cat"
                    :value="cat"
                  />
                </el-select>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label class="form-label">作者</label>
                <el-input
                  v-model="newPromptForm.author"
                  placeholder="输入作者名称..."
                  clearable
                />
              </div>

              <div class="form-group">
                <label class="form-label">版本号</label>
                <el-input
                  v-model="newPromptForm.version"
                  placeholder="如: v1.0"
                  clearable
                />
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">提示词内容</label>
              <el-input
                v-model="newPromptForm.content"
                type="textarea"
                :rows="8"
                placeholder="输入提示词内容，可使用 {{变量名}} 格式定义变量..."
              />
            </div>

            <div class="form-group">
              <label class="form-label">变量字段（可选）</label>
              <div class="fields-config">
                <div
                  v-for="(field, index) in newPromptForm.fields"
                  :key="index"
                  class="field-item"
                >
                  <el-input v-model="field.name" placeholder="变量名" style="width: 100px" />
                  <el-input v-model="field.label" placeholder="显示名称" style="flex: 1" />
                  <el-select v-model="field.type" style="width: 100px">
                    <el-option label="文本" value="text" />
                    <el-option label="多行" value="textarea" />
                    <el-option label="选择" value="select" />
                  </el-select>
                  <el-button type="danger" link @click="newPromptForm.fields.splice(index, 1)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
                <el-button type="primary" link @click="addNewPromptField">
                  <el-icon><Plus /></el-icon>
                  添加字段
                </el-button>
              </div>
            </div>

            <div class="form-actions">
              <el-button @click="resetNewPromptForm">重置</el-button>
              <el-button type="primary" @click="saveNewPrompt" :disabled="!canSaveNewPrompt">
                <el-icon><Check /></el-icon>
                保存提示词
              </el-button>
            </div>
          </div>
        </div>
      </template>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  CollectionTag,
  Close,
  Search,
  Filter,
  ArrowDown,
  Sort,
  Sunny,
  Document,
  Plus,
  Check,
  Delete
} from '@element-plus/icons-vue'

type PromptItem = {
  id: number
  name: string
  category: string
  content: string
  description?: string
  author?: string
  version?: string
  fields?: Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }>
  usage_count?: number
  created_at?: string
  is_featured?: boolean
}

type PinnedPrompt = PromptItem & { enabled: boolean }

const props = defineProps<{
  visible: boolean
  prompts: PromptItem[]
  pinnedPrompts: PinnedPrompt[]
  generatorName: string
  cardPacks: string[]
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'toggle-pin', prompt: PromptItem): void
  (e: 'remove-pin', promptId: number): void
  (e: 'toggle-enabled', promptId: number): void
  (e: 'create', prompt: PromptItem): void
}>()

// 切换生成器时重置搜索和分类筛选状态
watch(
  () => props.generatorName,
  () => {
    promptSearchKeyword.value = ''
    selectedPromptCategory.value = 'all'
  }
)

// 卡包变化时重置分类筛选
watch(
  () => props.cardPacks,
  () => {
    selectedPromptCategory.value = 'all'
  }
)

// 标签页
const activePromptTab = ref('all') // all: 全部, pinned: 已固定, created: 已创建
const promptTabs = [
  { key: 'all', label: '全部' },
  { key: 'pinned', label: '已固定' },
  { key: 'created', label: '已创建' }
]

// 搜索与分类筛选
const promptSearchKeyword = ref('')
const selectedPromptCategory = ref('all')
const showCategoryDropdown = ref(false)

const handleSearchPrompts = () => {
  // 搜索逻辑已在 computed 中实现
}

const toggleCategoryDropdown = () => {
  showCategoryDropdown.value = !showCategoryDropdown.value
}

const selectCategory = (category: string) => {
  selectedPromptCategory.value = category === '全部' ? 'all' : category
  showCategoryDropdown.value = false
}

const clearCategorySelection = () => {
  selectedPromptCategory.value = 'all'
}

// 排序
const activeSortOption = ref('latest') // latest: 最新, hot: 最热, monthly: 月榜, featured: 精选, search: 搜索
const sortOptions = [
  { key: 'latest', label: '最新' },
  { key: 'hot', label: '最热', icon: 'TrendCharts' },
  { key: 'monthly', label: '月榜' },
  { key: 'featured', label: '精选' },
  { key: 'search', label: '相关' }
]

const promptCategories = computed(() => {
  const categories = new Set<string>()

  // 如果当前打开了生成器，使用用户选择的卡包（支持1-2个）
  if (props.generatorName && props.cardPacks.length > 0) {
    const selectedPacks = props.cardPacks.filter(pack =>
      props.prompts.some(p => p.category === pack)
    )
    if (selectedPacks.length > 0) {
      return selectedPacks
    }
  }

  // 否则显示所有卡包
  props.prompts.forEach(p => {
    if (p.category) {
      categories.add(p.category)
    }
  })
  return ['全部', ...Array.from(categories)]
})

const promptSelectDialogTitle = computed(() => {
  const base = props.generatorName || '选择提示词'
  if (props.generatorName && props.cardPacks.length > 0) {
    const packNames = props.cardPacks.join('+')
    return `${base} · ${packNames}卡包`
  }
  return base
})

const filteredPrompts = computed(() => {
  let result = props.prompts

  // 如果当前打开了生成器且选择了卡包，只显示对应卡包的提示词
  if (props.generatorName && props.cardPacks.length > 0) {
    result = result.filter(p => props.cardPacks.includes(p.category))
  } else if (selectedPromptCategory.value !== 'all' && selectedPromptCategory.value !== '全部') {
    result = result.filter(p => p.category === selectedPromptCategory.value)
  }

  if (promptSearchKeyword.value) {
    const keyword = promptSearchKeyword.value.toLowerCase()
    result = result.filter(p =>
      p.name.toLowerCase().includes(keyword) ||
      p.category.toLowerCase().includes(keyword) ||
      p.content?.toLowerCase().includes(keyword)
    )
  }

  return result
})

const formatPromptTime = (timeStr?: string) => {
  const date = timeStr ? new Date(timeStr) : new Date()
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatUsageCount = (count?: number) => {
  if (count === undefined || count === null) return '0'
  if (count < 1000) return count.toString()
  if (count < 1_000_000) return `${(count / 1000).toFixed(1)}K`
  return `${(count / 1_000_000).toFixed(1)}M`
}

const getAvatarColor = (name: string) => {
  const colors = [
    'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
    'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    'linear-gradient(135deg, #a8edea 0%, #fed6e3 100%)',
    'linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%)',
    'linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%)',
    'linear-gradient(135deg, #a1c4fd 0%, #c2e9fb 100%)',
    'linear-gradient(135deg, #d299c2 0%, #fef9d7 100%)',
    'linear-gradient(135deg, #89f7fe 0%, #66a6ff 100%)',
    'linear-gradient(135deg, #cd9cf2 0%, #f6f3ff 100%)',
  ]
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return colors[Math.abs(hash) % colors.length]
}

// 排序和过滤后的提示词列表
const sortedAndFilteredPrompts = computed(() => {
  let result = [...filteredPrompts.value]

  // 根据排序选项排序
  switch (activeSortOption.value) {
    case 'hot':
      // 按使用次数降序（假设有 usage_count 字段）
      result.sort((a, b) => (b.usage_count || 0) - (a.usage_count || 0))
      break
    case 'monthly':
      // 本月热门（简化实现：按时间倒序）
      const thisMonth = new Date()
      const monthAgo = new Date(thisMonth.getTime() - 30 * 24 * 60 * 60 * 1000)
      result.sort((a, b) => {
        const aInMonth = new Date(a.created_at || '') > monthAgo ? 1 : 0
        const bInMonth = new Date(b.created_at || '') > monthAgo ? 1 : 0
        return bInMonth - aInMonth || (b.usage_count || 0) - (a.usage_count || 0)
      })
      break
    case 'featured':
      // 精选（标记为 is_featured 的优先）
      result.sort((a, b) => (b.is_featured ? 1 : 0) - (a.is_featured ? 1 : 0))
      break
    case 'search':
      // 搜索相关度排序（如果有搜索关键词，按匹配度排序）
      if (promptSearchKeyword.value.trim()) {
        const keyword = promptSearchKeyword.value.toLowerCase()
        result.sort((a, b) => {
          const aScore = calculateRelevanceScore(a, keyword)
          const bScore = calculateRelevanceScore(b, keyword)
          return bScore - aScore
        })
      }
      break
    case 'latest':
    default:
      // 默认按时间倒序
      result.sort((a, b) => (b.created_at || '').localeCompare(a.created_at || ''))
      break
  }

  return result
})

// 计算搜索相关度得分
const calculateRelevanceScore = (prompt: PromptItem, keyword: string): number => {
  let score = 0

  // 名称完全匹配
  if (prompt.name.toLowerCase().includes(keyword)) score += 10

  // 分类匹配
  if (prompt.category && prompt.category.toLowerCase().includes(keyword)) score += 5

  // 内容匹配
  if (prompt.content && prompt.content.toLowerCase().includes(keyword)) score += 3

  // 名称开头匹配加分更多
  if (prompt.name.toLowerCase().startsWith(keyword)) score += 5

  return score
}

const isPromptPinned = (promptId: number) => {
  return props.pinnedPrompts.some(p => p.id === promptId)
}

// 新建提示词表单
const newPromptForm = ref({
  name: '',
  category: '',
  author: '',
  version: '',
  content: '',
  fields: [] as Array<{ name: string; label: string; type: 'text' | 'textarea' | 'select'; options: string[]; optionLabels?: string[]; description: string; required: boolean }>
})

const canSaveNewPrompt = computed(() => {
  return newPromptForm.value.name.trim() !== '' &&
         newPromptForm.value.category !== '' &&
         newPromptForm.value.content.trim() !== ''
})

const addNewPromptField = () => {
  newPromptForm.value.fields.push({
    name: '',
    label: '',
    type: 'text',
    options: [],
    description: '',
    required: false
  })
}

const resetNewPromptForm = () => {
  newPromptForm.value = {
    name: '',
    category: '',
    author: '',
    version: '',
    content: '',
    fields: []
  }
}

const saveNewPrompt = () => {
  if (!canSaveNewPrompt.value) return

  const newPrompt: PromptItem = {
    id: Date.now(),
    name: newPromptForm.value.name,
    category: newPromptForm.value.category,
    description: newPromptForm.value.content.slice(0, 50),
    content: newPromptForm.value.content,
    author: newPromptForm.value.author || '匿名',
    version: newPromptForm.value.version || 'v1.0',
    fields: newPromptForm.value.fields.length > 0 ? newPromptForm.value.fields : undefined,
    created_at: new Date().toISOString(),
    usage_count: 0
  }

  emit('create', newPrompt)
  resetNewPromptForm()
  ElMessage.success('提示词创建成功！')
}
</script>
