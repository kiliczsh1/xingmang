<template>
  <div class="fullscreen-layout">
    <router-view />
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.fullscreen-layout {
  width: 100%;
  height: 100vh;
  overflow-x: hidden;
  overflow-y: auto;
}
</style>
