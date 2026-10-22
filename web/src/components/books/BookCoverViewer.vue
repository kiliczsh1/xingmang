<template>
  <Teleport to="body">
    <div
      v-if="visible"
      class="cover-zoom-mask"
      @click="close"
    >
      <div class="cover-zoom-stage" @click.stop>
        <img
          :src="url"
          class="cover-zoom-image"
          alt="封面放大"
        />
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
// 封面放大预览（Teleport 到 body 的纯展示层）
// 2026-09-21 第 3 期从 views/Books.vue 拆出；开/关状态由父组件持有
defineProps<{ visible: boolean; url: string }>()
const emit = defineEmits<{ (e: 'update:visible', v: boolean): void }>()

// 点遮罩关闭（原父组件的 closeCoverZoom）
const close = () => emit('update:visible', false)
</script>

<style scoped>
.cover-zoom-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  cursor: zoom-out;
}

.cover-zoom-stage {
  cursor: default;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-zoom-image {
  display: block;
  max-width: 80vw;
  max-height: 80vh;
  object-fit: contain;
  background: #1f2937;
  border-radius: 8px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.35);
}
</style>
