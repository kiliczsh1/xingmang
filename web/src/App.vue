<template>
  <div id="app">
    <router-view />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { loadLocalPlugins } from '@/components/ChatPanel/plugins'

// 应用启动即初始化本地插件（UI 类插件如主题切换按钮需全局挂载；
// 聊天链在 ChatMain 中追加注册，loadLocalPlugins 内部有 loaded 防重）
onMounted(() => {
  loadLocalPlugins()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

#app {
  width: 100%;
  height: 100%;
}
</style>

