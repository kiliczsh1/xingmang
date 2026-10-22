<template>
  <div
    v-if="visible"
    ref="rootRef"
    class="node-menu"
    :style="{ left: pos.x + 'px', top: pos.y + 'px' }"
    @mousedown.stop
    @contextmenu.prevent
  >
    <div class="menu-title" :title="nodeLabel">{{ nodeLabel || '节点' }}</div>
    <button class="menu-item" @click="emit('toggle-breakpoint')">
      <i class="fas" :class="isBreakpoint ? 'fa-play-circle' : 'fa-circle-pause'" />
      {{ isBreakpoint ? '取消断点' : '设为断点' }}
    </button>
    <button class="menu-item" @click="emit('test-run')">
      <i class="fas fa-flask" />
      单节点测试运行
    </button>
    <button class="menu-item" @click="emit('rerun')">
      <i class="fas fa-rotate-right" />
      从此节点重跑
    </button>
    <div class="menu-sep" />
    <button class="menu-item danger" @click="emit('delete')">
      <i class="fas fa-trash" />
      删除节点
    </button>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  x: number
  y: number
  nodeLabel?: string
  isBreakpoint?: boolean
}>()
const emit = defineEmits<{
  (e: 'toggle-breakpoint'): void
  (e: 'test-run'): void
  (e: 'rerun'): void
  (e: 'delete'): void
  (e: 'close'): void
}>()

const rootRef = ref<HTMLElement | null>(null)
const pos = ref({ x: 0, y: 0 })

const MENU_W = 176
const MENU_H = 176

const place = async () => {
  await nextTick()
  const parent = rootRef.value?.offsetParent as HTMLElement | null
  const pw = parent?.clientWidth ?? 0
  const ph = parent?.clientHeight ?? 0
  pos.value = {
    x: pw ? Math.max(4, Math.min(props.x, pw - MENU_W - 4)) : props.x,
    y: ph ? Math.max(4, Math.min(props.y, ph - MENU_H - 4)) : props.y
  }
}

const onDocDown = (e: MouseEvent) => {
  const el = rootRef.value
  if (el && e.target instanceof Node && el.contains(e.target)) return
  emit('close')
}
const onDocKey = (e: KeyboardEvent) => {
  if (e.key === 'Escape') emit('close')
}
const onResize = () => {
  if (props.visible) void place()
}

watch(
  () => props.visible,
  v => {
    if (v) {
      void place()
      document.addEventListener('mousedown', onDocDown, true)
      document.addEventListener('keydown', onDocKey, true)
      window.addEventListener('resize', onResize)
    } else {
      document.removeEventListener('mousedown', onDocDown, true)
      document.removeEventListener('keydown', onDocKey, true)
      window.removeEventListener('resize', onResize)
    }
  }
)

onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocDown, true)
  document.removeEventListener('keydown', onDocKey, true)
  window.removeEventListener('resize', onResize)
})
</script>

<style scoped>
.node-menu {
  position: absolute;
  z-index: 30;
  width: 176px;
  padding: 4px;
  background: var(--wf-el-bg-overlay, #ffffff);
  border: 1px solid var(--wf-el-border, #dcdfe6);
  border-radius: 6px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
  user-select: none;
}
.menu-title {
  padding: 5px 8px 6px;
  font-size: 11px;
  color: var(--wf-el-text-secondary, #909399);
  border-bottom: 1px solid var(--wf-el-border-light, #ebeef5);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.menu-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 7px 8px;
  border: none;
  border-radius: 4px;
  background: transparent;
  font-size: 13px;
  color: var(--wf-el-text-regular, #606266);
  cursor: pointer;
  text-align: left;
}
.menu-item:hover {
  background: var(--wf-el-fill, #f0f2f5);
}
.menu-item.danger {
  color: var(--el-color-danger, #f56c6c);
}
.menu-item.danger:hover {
  background: rgba(245, 108, 108, 0.12);
}
.menu-item i {
  font-size: 11px;
  width: 12px;
}
.menu-sep {
  height: 1px;
  margin: 4px 6px;
  background: var(--wf-el-border-light, #ebeef5);
}
</style>
