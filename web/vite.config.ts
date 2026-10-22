import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:19825',
        changeOrigin: true
      }
    },
    watch: {
      // ⚠ 构建产物目录必须排除，否则会引发「整页白屏重载」：
      // Vite 默认把项目根（web/）整棵递归纳入 watch，而 web/ 下存在
      // dist / dist-tmp-* / dist.bak* 等打包与备份目录。这些目录一旦被写入
      // （npm run build、build.bat、打包工具，甚至另一个会话跑构建），
      // watcher 会判定为「页面级变更」并触发 full-reload，浏览器整页刷新，
      // 首屏白屏重来一遍。实测：写入 web/dist-tmp-wp9/xxx.html →
      //   [vite] (client) page reload dist-tmp-wp9/xxx.html
      // 这些目录不是源码、也不需要 HMR，排除后无副作用。
      ignored: [
        '**/dist/**',
        '**/dist-*/**',
        '**/dist.*/**',
        '**/node_modules/**',
        '**/.git/**',
      ],
    },
  },
})
