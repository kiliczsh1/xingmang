<template>
  <div class="font-size-selector">
    <el-dropdown trigger="click" @command="select">
      <button type="button" class="size-trigger">
        <span class="size-label">{{ modelValue }}px</span>
        <svg class="size-chevron" viewBox="0 0 10 6" width="10" height="6">
          <path d="M1 1 L5 5 L9 1" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item
            v-for="size in sizes"
            :key="size"
            :command="size"
            :class="{ 'is-active': modelValue === size }"
          >
            {{ size }}px
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  modelValue: number
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void
  (e: 'change', value: number): void
}>()

const sizes = [12, 14, 15, 16, 18, 20, 22, 24, 28, 32]

const select = (size: number) => {
  emit('update:modelValue', size)
  emit('change', size)
}
</script>

<style scoped>
.font-size-selector {
  flex-shrink: 0;
}

.size-trigger {
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

.size-trigger:hover {
  background: #e2e5e9;
}

.size-label {
  font-size: 13px;
  line-height: 1;
}

.size-chevron {
  opacity: 0.4;
  flex-shrink: 0;
}

:root[data-theme='dark'] .size-trigger {
  color: #bababa;
}

:root[data-theme='dark'] .size-trigger:hover {
  background: #2c313c;
}
</style>
