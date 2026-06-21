<template>
  <div class="font-selector">
    <button
      type="button"
      class="font-selector-trigger"
      @click="visible = true"
    >
      <span class="trigger-label" :style="{ fontFamily: modelValue }">{{ displayLabel }}</span>
      <svg class="trigger-chevron" viewBox="0 0 10 6" width="10" height="6">
        <path d="M1 1 L5 5 L9 1" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
      </svg>
    </button>

    <Teleport to="body">
      <div v-if="visible" class="font-modal-overlay" @click="visible = false" />
      <div v-if="visible" class="font-modal">
        <div class="font-modal-header">
          <span class="font-modal-title">选择字体</span>
          <div class="font-modal-search">
            <input
              ref="searchRef"
              v-model="search"
              type="text"
              placeholder="搜索字体..."
              class="search-input"
            />
          </div>
          <button type="button" class="font-modal-close" @click="visible = false">
            <svg viewBox="0 0 24 24" width="18" height="18"><path d="M18 6L6 18M6 6l12 12" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
          </button>
        </div>

        <div class="font-modal-body">
          <template v-for="group in filteredGroups" :key="group.name">
            <div class="font-section-label">{{ group.name }}</div>
            <div class="font-card-grid">
              <button
                v-for="font in group.fonts"
                :key="font.value"
                type="button"
                class="font-card"
                :class="{ selected: modelValue === font.value }"
                :style="{ fontFamily: font.value }"
                @click="select(font.value)"
              >
                <span class="font-card-preview">{{ font.preview || 'Aa' }}</span>
                <span class="font-card-name">{{ font.label }}</span>
              </button>
            </div>
          </template>
          <div v-if="totalVisible === 0" class="font-modal-empty">
            没有找到匹配的字体
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

interface FontOption {
  value: string
  label: string
  preview?: string
}

interface FontGroup {
  name: string
  fonts: FontOption[]
}

const visible = ref(false)
const search = ref('')
const searchRef = ref<HTMLInputElement>()

const fontGroups: FontGroup[] = [
  {
    name: '推荐',
    fonts: [
      { value: 'Microsoft YaHei', label: '微软雅黑' },
      { value: 'PingFang SC', label: '苹方' },
      { value: 'Source Han Sans SC', label: '思源黑体' },
      { value: 'Source Han Serif SC', label: '思源宋体' },
    ]
  },
  {
    name: '中文无衬线',
    fonts: [
      { value: 'SimHei', label: '黑体' },
      { value: 'STHeiti', label: '华文黑体' },
      { value: 'Arial', label: 'Arial' },
      { value: 'Helvetica Neue', label: 'Helvetica Neue' },
    ]
  },
  {
    name: '中文衬线',
    fonts: [
      { value: 'SimSun', label: '宋体' },
      { value: 'FangSong', label: '仿宋' },
      { value: 'KaiTi', label: '楷体' },
      { value: 'STKaiti', label: '华文楷体' },
      { value: 'STSong', label: '华文宋体' },
      { value: 'Georgia', label: 'Georgia' },
    ]
  },
  {
    name: '等宽',
    fonts: [
      { value: 'Cascadia Code', label: 'Cascadia Code', preview: '1l' },
      { value: 'Fira Code', label: 'Fira Code', preview: '1l' },
      { value: 'JetBrains Mono', label: 'JetBrains Mono', preview: '1l' },
      { value: 'Consolas', label: 'Consolas', preview: '1l' },
      { value: 'Courier New', label: 'Courier New', preview: '1l' },
    ]
  },
]

const allOptions = computed(() => fontGroups.flatMap(g => g.fonts))
const displayLabel = computed(() => {
  const found = allOptions.value.find(f => f.value === props.modelValue)
  return found ? found.label : props.modelValue
})

const filteredGroups = computed(() => {
  const q = search.value.trim().toLowerCase()
  if (!q) return fontGroups
  return fontGroups
    .map(g => ({
      ...g,
      fonts: g.fonts.filter(f =>
        f.label.toLowerCase().includes(q) ||
        f.value.toLowerCase().includes(q)
      )
    }))
    .filter(g => g.fonts.length > 0)
})

const totalVisible = computed(() =>
  filteredGroups.value.reduce((sum, g) => sum + g.fonts.length, 0)
)

const select = (value: string) => {
  emit('update:modelValue', value)
  emit('change', value)
  nextTick(() => { visible.value = false })
}

onMounted(() => {
  const handler = (e: KeyboardEvent) => {
    if (e.key === 'Escape' && visible.value) {
      visible.value = false
    }
  }
  window.addEventListener('keydown', handler)
})
</script>

<style scoped>
.font-selector {
  flex-shrink: 0;
}

.font-selector-trigger {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 30px;
  padding: 0 10px;
  border: none;
  border-radius: 5px;
  background: transparent;
  cursor: pointer;
  color: #555;
  font-size: 13px;
  transition: background 0.1s ease;
  white-space: nowrap;
}

.font-selector-trigger:hover {
  background: #e2e5e9;
}

.trigger-label {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 13px;
  line-height: 1;
}

.trigger-chevron {
  opacity: 0.4;
  flex-shrink: 0;
}

:root[data-theme='dark'] .font-selector-trigger {
  color: #bababa;
}

:root[data-theme='dark'] .font-selector-trigger:hover {
  background: #2c313c;
}
</style>

<style>
.font-modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 3000;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(4px);
}

.font-modal {
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 680px;
  max-width: 92vw;
  max-height: 80vh;
  z-index: 3001;
  background: #fff;
  border-radius: 12px;
  box-shadow:
    0 0 0 1px rgba(0, 0, 0, 0.04),
    0 8px 40px rgba(0, 0, 0, 0.18);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.font-modal-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid #ebedf0;
  flex-shrink: 0;
}

.font-modal-title {
  font-size: 15px;
  font-weight: 600;
  color: #111;
  flex-shrink: 0;
}

.font-modal-search {
  flex: 1;
}

.font-modal .search-input {
  width: 100%;
  height: 32px;
  padding: 0 12px;
  border: none;
  border-radius: 6px;
  background: #f6f7f8;
  font-size: 13px;
  color: #333;
  outline: none;
  box-sizing: border-box;
}

.font-modal .search-input::placeholder {
  color: #999;
}

.font-modal-close {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #999;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.1s;
}

.font-modal-close:hover {
  background: #e2e5e9;
  color: #555;
}

.font-modal-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}

.font-section-label {
  font-size: 11px;
  font-weight: 600;
  color: #999;
  margin-bottom: 10px;
}

.font-section-label + .font-section-label {
  margin-top: 24px;
}

.font-card-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 8px;
  margin-bottom: 8px;
}

.font-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  aspect-ratio: 1;
  padding: 10px 6px;
  border: 1.5px solid #ebedf0;
  border-radius: 8px;
  background: #fff;
  cursor: pointer;
  transition: all 0.12s ease;
  min-width: 0;
  font-family: inherit;
}

.font-card:hover {
  border-color: #c4c4c4;
  background: #f6f7f8;
}

.font-card.selected {
  border-color: #086ddd;
  background: rgba(8, 109, 221, 0.04);
}

.font-card-preview {
  font-size: 28px;
  line-height: 1;
  color: #111;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
}

.font-card-name {
  font-size: 10.5px;
  color: #999;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.font-card.selected .font-card-name {
  color: #086ddd;
  font-weight: 500;
}

.font-modal-body::-webkit-scrollbar {
  width: 5px;
}

.font-modal-body::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.15);
  border-radius: 3px;
}

.font-modal-empty {
  text-align: center;
  padding: 40px;
  color: #999;
  font-size: 13px;
}

/* Dark mode */
:root[data-theme='dark'] .font-modal-overlay {
  background: rgba(0, 0, 0, 0.6);
}

:root[data-theme='dark'] .font-modal {
  background: #1c2127;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.06),
    0 8px 40px rgba(0, 0, 0, 0.5);
}

:root[data-theme='dark'] .font-modal-header {
  border-bottom-color: #35393e;
}

:root[data-theme='dark'] .font-modal-title {
  color: #dadada;
}

:root[data-theme='dark'] .font-modal .search-input {
  background: #282c34;
  color: #dadada;
}

:root[data-theme='dark'] .font-modal-close {
  color: #666;
}

:root[data-theme='dark'] .font-modal-close:hover {
  background: #2c313c;
  color: #bababa;
}

:root[data-theme='dark'] .font-card {
  background: #1c2127;
  border-color: #35393e;
}

:root[data-theme='dark'] .font-card:hover {
  border-color: #555;
  background: #282c34;
}

:root[data-theme='dark'] .font-card.selected {
  border-color: #027aff;
  background: rgba(2, 122, 255, 0.06);
}

:root[data-theme='dark'] .font-card-preview {
  color: #dadada;
}

:root[data-theme='dark'] .font-card-name {
  color: #666;
}

:root[data-theme='dark'] .font-card.selected .font-card-name {
  color: #027aff;
}

:root[data-theme='dark'] .font-modal-body::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.15);
}

:root[data-theme='dark'] .font-section-label {
  color: #666;
}
</style>
