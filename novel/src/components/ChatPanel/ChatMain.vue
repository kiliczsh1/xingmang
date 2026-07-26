<!-- ChatMain.vue —— 编排层（统一青绿主题色） -->
<template>
  <div class="ds-main">
    <ChatHeader
      :prompts-count="chat.selectedPrompts.value.length"
      :related-count="chat.relatedContent.value.length"
      :model-name="chat.currentModelName.value"
      :world-book-linked="chat.worldBookLinked.value"
      :world-book-enabled="worldBookStore.enabled"
      @toggle-drawer="$emit('toggleDrawer')"
      @new-conversation="chat.createConversation"
      @show-conversation-list="conversationListVisible = true"
      @show-prompt-select="promptSelectDialogVisible = true"
      @show-world-book="openWorldBookDialog"
      @show-relate-content="openRelateContentDialog"
      @show-model-select="openModelSelectDialog"
      @show-regex-filter="regexDialogVisible = true"
      @toggle-world-book-link="chat.worldBookLinked.value = $event"
      @show-full-prompt="showFullPromptDialog = true"
      @close="$emit('close')"
    />

    <div v-if="chat.currentConversation.value" class="ds-content">
      <ChatMessageList
        :messages="chat.chatMessages.value"
        :loading="chat.isLoading.value"
        :get-display-content="chat.getAssistantDisplayContent"
        :get-user-display-content="getUserDisplay"
        :get-copy-content="chat.getMessageCopyContent"
        :apply-to-cursor="chat.applyToCursor.value"
        @regenerate="chat.regenerateMessage"
        @delete-message="chat.deleteMessage"
      />

      <ChatInputArea
        :model-value="chat.userInput.value"
        @update:model-value="chat.userInput.value = $event"
        :loading="chat.isLoading.value"
        :model-name="chat.currentModelName.value"
        :attached-references="attachedReferences"
        :selected-prompts-count="chat.selectedPrompts.value.length"
        :related-content-count="chat.relatedContent.value.length"
        :at-menu-options="atMenuOptions"
        :parsed-variables="chat.parsedVariableNames.value"
        :variable-values="chat.promptVariables.value"
        @send="chat.sendMessage"
        @stop="chat.stopGeneration"
        @open-model-select="openModelSelectDialog"
        @at-select="handleAtSelect"
        @remove-reference="removeAttachedReference"
        @update-variable="onUpdateVariable"
      />
    </div>
    <div v-else class="ds-empty">
      <div class="ds-empty-icon">
        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2" style="opacity:0.2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        </svg>
      </div>
      <div class="ds-empty-text">开始新对话</div>
    </div>

    <!-- 模型选择弹窗 -->
    <Transition name="ds-modal">
      <div v-if="modelSelectDialogVisible" class="ds-modal-mask" @click.self="modelSelectDialogVisible = false">
        <div class="ds-modal">
          <div class="ds-modal-header">
            <span>选择模型</span>
            <button class="ds-icon-btn" @click="modelSelectDialogVisible = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="ds-modal-body">
            <div
              v-for="config in chat.apiConfigs.value"
              :key="config.id"
              :class="['ds-model-card', { active: chat.selectedConfigId.value === config.id }]"
              @click="chat.selectedConfigId.value = config.id; modelSelectDialogVisible = false"
            >
              <div class="ds-model-card-info">
                <span class="ds-model-card-name">{{ config.name }}</span>
                <span class="ds-model-card-desc">{{ config.description || config.model }}</span>
              </div>
              <div v-if="chat.selectedConfigId.value === config.id" class="ds-model-card-check">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#5eead4" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 提示词选择弹窗 -->
    <Transition name="ds-modal">
      <div v-if="promptSelectDialogVisible" class="ds-modal-mask" @click.self="promptSelectDialogVisible = false">
        <div class="ds-modal ds-modal--prompt-select">
          <div class="ds-modal-header">
            <span>选择提示词</span>
            <button class="ds-icon-btn" @click="promptSelectDialogVisible = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="ds-modal-body ds-modal-body--prompt">
            <!-- 固定顶部区域 -->
            <div class="ds-prompt-top">
              <!-- 搜索框 -->
              <div class="ds-prompt-search">
                <svg class="ds-prompt-search-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>
                </svg>
                <input
                  v-model="promptSearchKeyword"
                  class="ds-prompt-search-input"
                  placeholder="搜索提示词…"
                  autocomplete="off"
                />
              </div>

              <!-- 已选分类标签 -->
              <div v-if="selectedCategories.length > 0" class="ds-tag-filter-bar">
                <span
                  v-for="cat in selectedCategories"
                  :key="cat"
                  class="ds-tag-filter-chip"
                >
                  {{ cat }}
                  <button class="ds-tag-filter-remove" @click="toggleCategory(cat)">✕</button>
                </span>
                <button class="ds-tag-filter-clear" @click="selectedCategories = []">清除</button>
              </div>

              <!-- 标签云 -->
              <div class="ds-tag-cloud">
                <button
                  v-for="cat in availableCategories"
                  :key="cat"
                  :class="['ds-tag-cloud-item', { active: selectedCategories.includes(cat) }]"
                  @click="toggleCategory(cat)"
                >
                  {{ cat }}
                  <span class="ds-tag-cloud-count">{{ categoryCount(cat) }}</span>
                </button>
                <div v-if="availableCategories.length === 0" class="ds-modal-empty" style="padding:12px 0">暂无分类</div>
              </div>

              <div class="ds-prompt-divider" />
            </div>

            <!-- 可滚动提示词列表 -->
            <div class="ds-prompt-scroll">
              <template v-if="paginatedPrompts.length > 0">
                <label
                  v-for="prompt in paginatedPrompts"
                  :key="prompt.id"
                  :class="['ds-prompt-card', { active: chat.selectedPrompts.value.some(id => id == prompt.id) }]"
                >
                  <input
                    type="checkbox"
                    :value="prompt.id"
                    :checked="chat.selectedPrompts.value.some(id => id == prompt.id)"
                    @change="togglePrompt(prompt.id)"
                    class="ds-prompt-checkbox"
                  />
                  <div class="ds-prompt-card-info">
                    <div class="ds-prompt-card-head">
                      <span class="ds-prompt-card-name">{{ prompt.name }}</span>
                      <span class="ds-prompt-card-cat-tag">{{ prompt.category }}</span>
                    </div>
                    <span class="ds-prompt-card-desc">{{ prompt.description || prompt.content?.substring(0, 80) }}</span>
                  </div>
                </label>
              </template>
              <div v-else class="ds-modal-empty" style="grid-column:1/-1">
                {{ chat.prompts.value.length === 0 ? '暂无提示词' : '没有匹配的提示词' }}
              </div>
            </div>

            <!-- 分页 -->
            <div v-if="totalPages > 1" class="ds-prompt-pagination">
              <button
                class="ds-page-btn"
                :disabled="currentPage <= 1"
                @click="currentPage--"
              >‹</button>
              <template v-for="(item, idx) in paginationItems" :key="idx">
                <span v-if="item === '...'" class="ds-page-ellipsis">…</span>
                <button
                  v-else
                  :class="['ds-page-btn', { active: item === currentPage }]"
                  @click="currentPage = item"
                >{{ item }}</button>
              </template>
              <button
                class="ds-page-btn"
                :disabled="currentPage >= totalPages"
                @click="currentPage++"
              >›</button>
            </div>

            <!-- 固定底部 -->
            <div class="ds-prompt-footer">
              已选 {{ chat.selectedPrompts.value.length }} 个提示词
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 正则过滤弹窗 -->
    <Transition name="ds-modal">
      <div v-if="regexDialogVisible" class="ds-modal-mask" @click.self="regexDialogVisible = false">
        <div class="ds-modal ds-modal--wide">
          <div class="ds-modal-header">
            <span>正则过滤</span>
            <button class="ds-icon-btn" @click="regexDialogVisible = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="ds-modal-body">
            <div class="ds-regex-toggles">
              <label class="ds-toggle-row">
                <span>启用过滤</span>
                <button :class="['ds-wb-switch', { on: chat.regexFilter.displayRegexEnabled.value }]" @click="chat.regexFilter.displayRegexEnabled.value = !chat.regexFilter.displayRegexEnabled.value">
                  <span class="ds-wb-knob" />
                </button>
              </label>
              <label class="ds-toggle-row">
                <span>流式防闪</span>
                <button :class="['ds-wb-switch', { on: chat.regexFilter.streamGuardEnabled.value }]" @click="chat.regexFilter.streamGuardEnabled.value = !chat.regexFilter.streamGuardEnabled.value">
                  <span class="ds-wb-knob" />
                </button>
              </label>
              <label class="ds-toggle-row">
                <span>复制过滤后</span>
                <button :class="['ds-wb-switch', { on: chat.regexFilter.copyUsesFiltered.value }]" @click="chat.regexFilter.copyUsesFiltered.value = !chat.regexFilter.copyUsesFiltered.value">
                  <span class="ds-wb-knob" />
                </button>
              </label>
            </div>
            <div class="ds-regex-actions">
              <button class="ds-regex-action-btn" @click="chat.regexFilter.addRegexRule()">+ 新增规则</button>
              <button class="ds-regex-action-btn" @click="chat.regexFilter.resetRegexRules()">重置默认</button>
            </div>
            <div class="ds-regex-list">
              <div v-for="(rule, idx) in chat.regexFilter.regexRules.value" :key="rule.id" class="ds-regex-rule">
                <div class="ds-regex-rule-head">
                  <button :class="['ds-wb-switch', 'ds-wb-switch--sm', { on: rule.enabled }]" @click="rule.enabled = !rule.enabled">
                    <span class="ds-wb-knob" />
                  </button>
                  <input v-model="rule.name" class="ds-regex-input ds-regex-input--name" placeholder="规则名称" />
                  <input v-model="rule.flags" class="ds-regex-input ds-regex-input--flags" placeholder="flags" />
                  <button class="ds-regex-action-btn" :disabled="idx === 0" @click="chat.regexFilter.moveRule(idx, -1)">&#8593;</button>
                  <button class="ds-regex-action-btn" :disabled="idx === chat.regexFilter.regexRules.value.length - 1" @click="chat.regexFilter.moveRule(idx, 1)">&#8595;</button>
                  <button class="ds-regex-action-btn ds-regex-action-btn--danger" @click="chat.regexFilter.removeRule(rule.id)">删除</button>
                </div>
                <div class="ds-regex-rule-body">
                  <textarea v-model="rule.pattern" class="ds-regex-textarea" placeholder="pattern" rows="2" />
                  <textarea v-model="rule.replacement" class="ds-regex-textarea" placeholder="replacement" rows="2" />
                </div>
                <div v-if="rule.lastError" class="ds-regex-error">{{ rule.lastError }}</div>
              </div>
            </div>
            <div class="ds-regex-test">
              <div class="ds-regex-test-label">测试输入</div>
              <textarea v-model="chat.regexFilter.regexTestInput.value" class="ds-regex-textarea" placeholder="输入测试文本" rows="3" />
              <div class="ds-regex-test-label">过滤结果</div>
              <textarea :value="chat.regexFilter.regexTestOutput.value" class="ds-regex-textarea" readonly rows="3" />
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 完整提示词预览弹窗 -->
    <Transition name="ds-modal">
      <div v-if="showFullPromptDialog" class="ds-modal-mask" @click.self="showFullPromptDialog = false">
        <div class="ds-modal ds-modal--wide">
          <div class="ds-modal-header">
            <span>完整提示词预览</span>
            <div style="display:flex;gap:6px">
              <button class="ds-regex-action-btn" @click="copyFullPrompt">复制</button>
              <button class="ds-icon-btn" @click="showFullPromptDialog = false">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              </button>
            </div>
          </div>
          <div class="ds-modal-body">
            <pre class="ds-full-prompt">{{ fullPromptContent }}</pre>
          </div>
        </div>
      </div>
    </Transition>

    <!-- 世界书弹窗 -->
    <Transition name="ds-modal">
      <div v-if="worldBookDialogVisible" class="ds-modal-mask" @click.self="worldBookDialogVisible = false">
        <div class="ds-modal ds-modal--wide">
          <div class="ds-modal-header">
            <span>世界书条目</span>
            <button class="ds-icon-btn" @click="worldBookDialogVisible = false">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
            </button>
          </div>
          <div class="ds-modal-body">
            <div class="ds-wb-dialog-toolbar">
              <span class="ds-wb-dialog-info">
                共 {{ worldBookEntriesForDisplay.length }} 条
                <span v-if="chat.worldBookLinked.value" class="ds-wb-dialog-badge">已关联</span>
                <span v-else class="ds-wb-dialog-badge ds-wb-dialog-badge--off">未关联</span>
              </span>
              <button
                :class="['ds-regex-action-btn', { active: chat.worldBookLinked.value }]"
                @click="chat.worldBookLinked.value = !chat.worldBookLinked.value"
              >
                {{ chat.worldBookLinked.value ? '断开关联' : '关联世界书' }}
              </button>
            </div>
            <div v-if="worldBookEntriesForDisplay.length === 0" class="ds-modal-empty">
              暂无世界书条目，请在写作页面中添加
            </div>
            <div
              v-for="entry in worldBookEntriesForDisplay"
              :key="entry.uid"
              class="ds-wb-entry"
              :class="{ 'ds-wb-entry--disabled': entry.disable }"
            >
              <div class="ds-wb-entry-head">
                <span class="ds-wb-entry-name">{{ entry.name || entry.comment || '(未命名)' }}</span>
                <span v-if="entry.disable" class="ds-wb-entry-status">已禁用</span>
              </div>
              <div class="ds-wb-entry-content">{{ entry.content }}</div>
              <div v-if="entry.keyword" class="ds-wb-entry-keyword">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="3.5" cy="3.5" r="1.5"/><path d="M21 15l-5-5L5 21"/></svg>
                {{ entry.keywords || entry.keyword }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import ChatHeader from './ChatHeader.vue'
import ChatMessageList from './ChatMessageList.vue'
import ChatInputArea from './ChatInputArea.vue'
import { useChatPanel } from './composables/useChatPanel'
import { useWorldBookStore } from '@/stores/worldbook'
import { promptAPI } from '@/api'
import type { RelatedContent } from '@/types'
import { promptInjectPlugin } from './plugins/promptInjectPlugin'
import { worldBookInjectPlugin } from './plugins/worldBookInjectPlugin'
import { relatedContentPlugin } from './plugins/relatedContentPlugin'

const props = defineProps<{ bookId: number }>()
const emit = defineEmits<{
  (e: 'close'): void
  (e: 'toggleDrawer'): void
  (e: 'conversationsLoaded', list: any[]): void
  (e: 'conversationSelected', conv: any): void
}>()

const worldBookStore = useWorldBookStore()
const chat = useChatPanel({ bookId: props.bookId })

// ===== 注册插件链 =====
chat.pluginChain.register(promptInjectPlugin)
chat.pluginChain.register(worldBookInjectPlugin)
chat.pluginChain.register(relatedContentPlugin)

const worldBookDialogVisible = ref(false)
const modelSelectDialogVisible = ref(false)
const promptSelectDialogVisible = ref(false)
const regexDialogVisible = ref(false)
const showFullPromptDialog = ref(false)
const conversationListVisible = ref(false)

const atMenuOptions = [
  { type: 'chapter', label: '章节', hint: '引用章节内容', icon: 'Document', color: '#5eead4' },
  { type: 'memo', label: '备忘录', hint: '引用备忘录内容', icon: 'Notebook', color: '#60a5fa' },
]

const attachedReferences = computed(() =>
  chat.relatedContent.value.map(rc => ({
    ...rc,
    label: `${rc.type === 'chapter' ? '章节' : '备忘录'}：${rc.title}`,
  }))
)

const getUserDisplay = (msg: any) =>
  msg.displayContent || chat.extractUserDisplayContent(msg.content) || msg.content

const togglePrompt = (id: number) => {
  const idx = chat.selectedPrompts.value.findIndex(pid => pid == id)
  if (idx >= 0) chat.selectedPrompts.value.splice(idx, 1)
  else chat.selectedPrompts.value.push(id)
}

// ===== 标签云搜索过滤 =====
const promptSearchKeyword = ref('')
const selectedCategories = ref<string[]>([])

const availableCategories = computed(() => {
  const cats = new Set<string>()
  chat.prompts.value.forEach((p: any) => { if (p.category) cats.add(p.category) })
  return Array.from(cats).sort()
})

const categoryCount = (cat: string) => {
  return chat.prompts.value.filter((p: any) => p.category === cat).length
}

const filteredPrompts = computed(() => {
  let result = chat.prompts.value
  if (selectedCategories.value.length > 0) {
    result = result.filter((p: any) => selectedCategories.value.includes(p.category))
  }
  if (promptSearchKeyword.value.trim()) {
    const kw = promptSearchKeyword.value.toLowerCase().trim()
    result = result.filter((p: any) =>
      p.name.toLowerCase().includes(kw) ||
      (p.content && p.content.toLowerCase().includes(kw)) ||
      (p.description && p.description.toLowerCase().includes(kw))
    )
  }
  return result
})

const toggleCategory = (cat: string) => {
  const idx = selectedCategories.value.indexOf(cat)
  if (idx >= 0) selectedCategories.value.splice(idx, 1)
  else selectedCategories.value.push(cat)
}

const PAGE_SIZE = 6
const currentPage = ref(1)

const totalPages = computed(() => Math.max(1, Math.ceil(filteredPrompts.value.length / PAGE_SIZE)))

const paginatedPrompts = computed(() => {
  const start = (currentPage.value - 1) * PAGE_SIZE
  return filteredPrompts.value.slice(start, start + PAGE_SIZE)
})

const paginationItems = computed(() => {
  const total = totalPages.value
  const cur = currentPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const items: (number | '...')[] = [1]
  if (cur > 3) items.push('...')
  const start = Math.max(2, cur - 1)
  const end = Math.min(total - 1, cur + 1)
  for (let i = start; i <= end; i++) items.push(i)
  if (cur < total - 2) items.push('...')
  items.push(total)
  return items
})

watch([selectedCategories, promptSearchKeyword], () => { currentPage.value = 1 })

watch(promptSelectDialogVisible, (v) => {
  if (v) { promptSearchKeyword.value = ''; selectedCategories.value = [] }
})

const fullPromptContent = computed(() => {
  const selectedPromptRecords = chat.prompts.value.filter((p: any) => chat.selectedPrompts.value.some(id => id == p.id))
  const promptContents = selectedPromptRecords.map((p: any) => p.content).join('\n\n')
  const relatedContentSummary = chat.relatedContent.value.length > 0
    ? chat.relatedContent.value.map(item => `${item.type === 'chapter' ? '章节' : '备忘录'}：${item.title}\n${item.content}`).join('\n\n')
    : ''
  const parts: string[] = []
  if (promptContents) parts.push(`提示词：\n${promptContents}`)
  if (chat.worldBookLinked.value) {
    const entries = chat.worldBookEntries.value || []
    if (entries.length > 0) parts.push(`世界书：${entries.length} 条`)
  }
  if (relatedContentSummary) parts.push(relatedContentSummary)
  if (chat.userInput.value.trim()) parts.push(chat.userInput.value.trim())
  return parts.join('\n\n') || '（暂无内容）'
})

const copyFullPrompt = () => {
  navigator.clipboard.writeText(fullPromptContent.value)
  ElMessage.success('已复制')
}

// 世界书条目（用于弹窗显示、包含禁用条目）
const worldBookEntriesForDisplay = computed(() => {
  if (!props.bookId) return []
  const wb = worldBookStore.getBookWorldBook(String(props.bookId))
  return wb?.entries || []
})

const openModelSelectDialog = () => { modelSelectDialogVisible.value = true }
const openWorldBookDialog = () => {
  // 同步加载最新条目
  if (props.bookId) {
    const wb = worldBookStore.getBookWorldBook(String(props.bookId))
    chat.worldBookEntries.value = wb?.entries || []
  }
  worldBookDialogVisible.value = true
}
const openRelateContentDialog = () => { /* 由外部 Write.vue 处理 */ }
const handleAtSelect = async (_type: string) => { /* 由外部处理 */ }
const onUpdateVariable = (name: string, value: string) => {
  chat.promptVariables.value[name] = value
}
const removeAttachedReference = (ref: any) => {
  chat.relatedContent.value = chat.relatedContent.value.filter(rc => !(rc.type === ref.type && rc.id === ref.id))
}

const fetchPrompts = async () => {
  try {
    const res = await promptAPI.getAll()
    if (res.success && res.data) chat.prompts.value = res.data
  } catch (error) { console.error('加载提示词失败:', error) }
}

watch(() => chat.worldBookLinked.value, (linked) => {
  if (linked && props.bookId) {
    const wb = worldBookStore.getBookWorldBook(String(props.bookId))
    chat.worldBookEntries.value = wb?.entries || []
  } else {
    chat.worldBookEntries.value = []
  }
})

watch(() => chat.conversations.value, (list) => { emit('conversationsLoaded', list) }, { deep: true })
watch(() => chat.currentConversation.value, (conv) => { emit('conversationSelected', conv) })

onMounted(() => { chat.init(); fetchPrompts() })

// 暴露给父组件（index.vue 的抽屉需要调 createConversation 等）
defineExpose({
  chat,
  createConversation: chat.createConversation,
  selectConversation: chat.selectConversation,
  deleteConversation: chat.deleteConversation,
  renameConversation: chat.renameConversation,
})
</script>

<style scoped>
.ds-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans SC", sans-serif;
}
.ds-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}
.ds-empty {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}
.ds-empty-text {
  font-size: 14px;
  color: var(--text-muted, #6b7280);
}

/* ---- 弹窗通用 ---- */
.ds-modal-mask {
  position: fixed; inset: 0; z-index: 1000;
  background: rgba(0,0,0,0.6);
  display: flex; align-items: center; justify-content: center;
  backdrop-filter: blur(4px);
}
.ds-modal {
  width: 420px; max-height: 80vh;
  background: #0f1d33;
  border: 1px solid var(--border-secondary, rgba(45, 212, 191, 0.16));
  border-radius: 16px;
  display: flex; flex-direction: column;
  box-shadow: 0 20px 60px rgba(0,0,0,0.5);
  overflow: hidden;
}
.ds-modal--wide { width: 620px; }
.ds-modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  font-size: 15px; font-weight: 600;
  color: var(--text-primary, #f4f7ff);
}
.ds-modal-body {
  flex: 1; overflow-y: auto;
  padding: 16px 20px;
  display: flex; flex-direction: column; gap: 12px;
}
.ds-modal-empty {
  text-align: center; padding: 30px;
  color: var(--text-muted, #6b7280); font-size: 13px;
}

/* 通用图标按钮 */
.ds-icon-btn {
  width: 28px; height: 28px;
  display: flex; align-items: center; justify-content: center;
  border-radius: 8px; border: none; background: transparent;
  color: var(--text-tertiary, #94a3b8);
  cursor: pointer; transition: all 120ms;
}
.ds-icon-btn:hover { background: rgba(94, 234, 212, 0.08); color: #5eead4; }

/* ---- 模型卡片 ---- */
.ds-model-card {
  display: flex; align-items: center; justify-content: space-between;
  padding: 12px 14px; border-radius: 10px;
  cursor: pointer; transition: all 120ms;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
}
.ds-model-card:hover { background: rgba(94, 234, 212, 0.04); border-color: rgba(94, 234, 212, 0.2); }
.ds-model-card.active { border-color: rgba(94, 234, 212, 0.35); background: rgba(94, 234, 212, 0.06); }
.ds-model-card-info { display: flex; flex-direction: column; gap: 2px; }
.ds-model-card-name { font-size: 13px; font-weight: 500; color: var(--text-primary, #f4f7ff); }
.ds-model-card-desc { font-size: 11px; color: var(--text-muted, #6b7280); }
.ds-model-card-check { color: #5eead4; }

/* ---- 提示词卡片 ---- */
.ds-prompt-card {
  display: flex; align-items: flex-start; gap: 10px;
  padding: 10px 14px; border-radius: 10px;
  cursor: pointer; transition: all 120ms;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
}
.ds-prompt-card:hover { background: rgba(94, 234, 212, 0.04); }
.ds-prompt-card.active { border-color: rgba(94, 234, 212, 0.35); background: rgba(94, 234, 212, 0.06); }
.ds-prompt-checkbox { accent-color: #5eead4; margin-top: 3px; }
.ds-prompt-card-info { display: flex; flex-direction: column; gap: 2px; }
.ds-prompt-card-name { font-size: 13px; font-weight: 500; color: var(--text-primary, #f4f7ff); }
.ds-prompt-card-desc { font-size: 11px; color: var(--text-muted, #6b7280); }

/* ---- 提示词选择弹窗 - 标签云 + 搜索 ---- */
.ds-modal--prompt-select { width: 812px; overflow: hidden; }
.ds-modal-body--prompt {
  overflow: hidden;
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 16px 20px;
  gap: 0;
}
.ds-prompt-top {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.ds-prompt-scroll {
  flex: 1;
  overflow-y: auto;
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
  padding: 4px 2px 0 0;
  margin: 12px 0 0;
  min-height: 0;
  align-content: start;
}
.ds-prompt-search {
  position: relative;
  display: flex; align-items: center;
}
.ds-prompt-search-icon {
  position: absolute; left: 12px;
  color: var(--text-muted, #6b7280);
  pointer-events: none;
}
.ds-prompt-search-input {
  width: 100%;
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 10px;
  color: var(--text-primary, #f4f7ff);
  font-size: 13px;
  padding: 10px 12px 10px 34px;
  outline: none;
  font-family: inherit;
  transition: border-color 120ms;
}
.ds-prompt-search-input::placeholder { color: var(--text-muted, #6b7280); }
.ds-prompt-search-input:focus { border-color: rgba(94, 234, 212, 0.35); }

.ds-tag-filter-bar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  padding: 4px 0 2px;
}
.ds-tag-filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border-radius: 6px;
  font-size: 11px;
  background: rgba(94, 234, 212, 0.12);
  color: #5eead4;
  border: 1px solid rgba(94, 234, 212, 0.2);
}
.ds-tag-filter-remove {
  width: 14px; height: 14px;
  display: inline-flex; align-items: center; justify-content: center;
  border: none; background: transparent;
  color: #5eead4; cursor: pointer;
  font-size: 10px; line-height: 1;
  opacity: 0.6; transition: opacity 120ms;
  padding: 0;
}
.ds-tag-filter-remove:hover { opacity: 1; }
.ds-tag-filter-clear {
  font-size: 11px;
  color: var(--text-muted, #6b7280);
  border: none; background: transparent;
  cursor: pointer; transition: color 120ms;
  padding: 2px 4px;
}
.ds-tag-filter-clear:hover { color: #f87171; }

.ds-tag-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 2px 0;
}
.ds-tag-cloud-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 12px;
  border-radius: 8px;
  font-size: 12px;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  background: rgba(51, 65, 85, 0.3);
  color: var(--text-secondary, #c8d0e0);
  cursor: pointer;
  transition: all 120ms;
}
.ds-tag-cloud-item:hover {
  border-color: rgba(94, 234, 212, 0.25);
  background: rgba(94, 234, 212, 0.06);
  color: #e2e8f0;
}
.ds-tag-cloud-item.active {
  border-color: rgba(94, 234, 212, 0.35);
  background: rgba(94, 234, 212, 0.1);
  color: #5eead4;
}
.ds-tag-cloud-count {
  font-size: 10px;
  color: var(--text-muted, #6b7280);
  background: rgba(255,255,255,0.04);
  border-radius: 4px;
  padding: 0 5px;
  line-height: 1.5;
}
.ds-tag-cloud-item.active .ds-tag-cloud-count {
  color: rgba(94, 234, 212, 0.6);
}

.ds-prompt-divider {
  height: 1px;
  background: var(--border-primary, rgba(45, 212, 191, 0.08));
  margin: 0;
  flex-shrink: 0;
}

/* ---- 分页 ---- */
.ds-prompt-pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  flex-shrink: 0;
  padding: 8px 0 0;
}
.ds-page-btn {
  min-width: 28px; height: 28px;
  display: inline-flex; align-items: center; justify-content: center;
  border-radius: 6px;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  background: transparent;
  color: var(--text-secondary, #c8d0e0);
  font-size: 12px;
  cursor: pointer;
  transition: all 120ms;
  padding: 0 6px;
}
.ds-page-btn:hover:not(:disabled) {
  border-color: rgba(94, 234, 212, 0.25);
  background: rgba(94, 234, 212, 0.06);
  color: #e2e8f0;
}
.ds-page-btn.active {
  border-color: rgba(94, 234, 212, 0.35);
  background: rgba(94, 234, 212, 0.12);
  color: #5eead4;
}
.ds-page-btn:disabled {
  opacity: 0.25;
  cursor: not-allowed;
}
.ds-page-ellipsis {
  width: 20px;
  text-align: center;
  color: var(--text-muted, #6b7280);
  font-size: 12px;
  letter-spacing: 1px;
}

.ds-prompt-card-head {
  display: flex;
  align-items: center;
  gap: 6px;
}
.ds-prompt-card-cat-tag {
  font-size: 10px;
  padding: 0 6px;
  border-radius: 4px;
  background: rgba(94, 234, 212, 0.08);
  color: rgba(94, 234, 212, 0.55);
  line-height: 1.6;
}

.ds-prompt-footer {
  flex-shrink: 0;
  text-align: center;
  font-size: 11px;
  color: var(--text-muted, #6b7280);
  padding: 10px 0 2px;
  border-top: 1px solid var(--border-primary, rgba(45, 212, 191, 0.06));
  margin-top: 12px;
}

/* ---- 正则过滤 ---- */
.ds-regex-toggles { display: flex; gap: 16px; flex-wrap: wrap; }
.ds-toggle-row {
  display: flex; align-items: center; gap: 8px;
  font-size: 13px; color: var(--text-secondary, #c8d0e0);
}
.ds-wb-switch {
  width: 30px; height: 16px;
  border-radius: 8px; border: none;
  background: rgba(255,255,255,0.08);
  cursor: pointer; position: relative;
  transition: background 150ms; padding: 0;
}
.ds-wb-switch.on { background: rgba(16,185,129,0.4); }
.ds-wb-switch--sm { width: 24px; height: 14px; }
.ds-wb-knob {
  position: absolute; top: 2px; left: 2px;
  width: 12px; height: 12px; border-radius: 50%;
  background: var(--text-muted, #6b7280);
  transition: all 150ms;
}
.ds-wb-switch.on .ds-wb-knob { left: 16px; background: #5eead4; box-shadow: 0 0 6px rgba(94, 234, 212, 0.4); }
.ds-wb-switch--sm .ds-wb-knob { width: 10px; height: 10px; }
.ds-wb-switch--sm.on .ds-wb-knob { left: 12px; }

.ds-regex-actions { display: flex; gap: 8px; }
.ds-regex-action-btn {
  padding: 4px 12px; border-radius: 8px;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  background: rgba(94, 234, 212, 0.04);
  color: var(--text-secondary, #c8d0e0);
  font-size: 12px; cursor: pointer; transition: all 120ms;
}
.ds-regex-action-btn:hover { background: rgba(94, 234, 212, 0.1); color: #5eead4; }
.ds-regex-action-btn:disabled { opacity: 0.3; cursor: not-allowed; }
.ds-regex-action-btn--danger { border-color: rgba(248, 113, 113, 0.15); color: rgba(248, 113, 113, 0.7); }
.ds-regex-action-btn--danger:hover { background: rgba(248, 113, 113, 0.1); color: #f87171; }

.ds-regex-list { display: flex; flex-direction: column; gap: 10px; }
.ds-regex-rule {
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 10px; padding: 12px;
  background: rgba(51, 65, 85, 0.3);
}
.ds-regex-rule-head { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.ds-regex-input {
  background: rgba(51, 65, 85, 0.6);
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 6px; color: var(--text-primary, #f4f7ff);
  font-size: 12px; padding: 4px 8px; outline: none; font-family: inherit; transition: border-color 120ms;
}
.ds-regex-input:focus { border-color: rgba(94, 234, 212, 0.4); }
.ds-regex-input--name { width: 180px; }
.ds-regex-input--flags { width: 70px; }
.ds-regex-rule-body { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.ds-regex-textarea {
  background: rgba(30, 41, 59, 0.6);
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 8px; color: var(--text-primary, #f4f7ff);
  font-size: 12px; padding: 8px 10px; outline: none; resize: vertical;
  font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', monospace; transition: border-color 120ms;
}
.ds-regex-textarea:focus { border-color: rgba(94, 234, 212, 0.4); }
.ds-regex-error { margin-top: 6px; color: #f87171; font-size: 12px; }

.ds-regex-test { display: flex; flex-direction: column; gap: 6px; margin-top: 4px; }
.ds-regex-test-label { font-size: 11px; color: var(--text-muted, #6b7280); font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }

/* 完整提示词 */
.ds-full-prompt {
  white-space: pre-wrap; word-break: break-word;
  font-size: 12px; line-height: 1.6;
  font-family: 'SF Mono', 'Cascadia Code', monospace;
  color: var(--text-secondary, #c8d0e0);
  background: rgba(30, 41, 59, 0.5);
  border-radius: 10px; padding: 16px;
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
}

/* ---- 弹窗动画 ---- */
.ds-modal-enter-active,
.ds-modal-leave-active { transition: all 200ms cubic-bezier(0.4,0,0.2,1); }
.ds-modal-enter-from,
.ds-modal-leave-to { opacity: 0; }
.ds-modal-enter-from .ds-modal,
.ds-modal-leave-to .ds-modal { transform: scale(0.95) translateY(10px); }

/* ---- 世界书弹窗 ---- */
.ds-wb-dialog-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
}
.ds-wb-dialog-info {
  font-size: 13px;
  color: var(--text-secondary, #c8d0e0);
  display: flex;
  align-items: center;
  gap: 8px;
}
.ds-wb-dialog-badge {
  font-size: 11px;
  padding: 1px 8px;
  border-radius: 6px;
  background: rgba(16, 185, 129, 0.15);
  color: #34d399;
}
.ds-wb-dialog-badge--off {
  background: rgba(248, 113, 113, 0.1);
  color: #f87171;
}
.ds-wb-entry {
  border: 1px solid var(--border-primary, rgba(45, 212, 191, 0.10));
  border-radius: 10px;
  padding: 12px 14px;
  background: rgba(51, 65, 85, 0.3);
  transition: border-color 120ms;
}
.ds-wb-entry--disabled {
  opacity: 0.5;
}
.ds-wb-entry-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.ds-wb-entry-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary, #f4f7ff);
}
.ds-wb-entry-status {
  font-size: 11px;
  color: var(--text-muted, #6b7280);
}
.ds-wb-entry-content {
  font-size: 12.5px;
  line-height: 1.6;
  color: var(--text-secondary, #c8d0e0);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  line-clamp: 3;
  overflow: hidden;
}
.ds-wb-entry-keyword {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  font-size: 11px;
  color: var(--text-muted, #6b7280);
}
</style>