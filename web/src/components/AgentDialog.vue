<template>
  <el-dialog
    v-model="visible"
    width="var(--app-dialog-width-content, min(1080px, 92vw))"
    class="agent-dialog-shell"
    append-to-body
    destroy-on-close
    :show-close="false"
  >
    <template #header>
      <div class="ad-head">
        <span class="ad-title">
          <el-icon><MagicStick /></el-icon>
          星芒 Agent
        </span>
        <span class="ad-hint">对话式写作助手 · 与侧栏「星芒 Agent」是同一份数据</span>
        <button type="button" class="ad-close" title="关闭 (Esc)" @click="visible = false">
          <el-icon><Close /></el-icon>
        </button>
      </div>
    </template>

    <!--
      Agent.vue 根元素是 .agent-page { height: 100% }，
      全屏路由下高度来自 MainLayout 的内容区；放进弹窗必须由这里给一个确定高度，
      否则双栏会塌成一条。
    -->
    <div class="ad-body">
      <Agent />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { MagicStick, Close } from '@element-plus/icons-vue'
import Agent from '@/views/Agent.vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'update:visible', value: boolean): void }>()

const visible = computed({
  get: () => props.visible,
  set: value => emit('update:visible', value)
})
</script>

<style scoped>
.ad-head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.ad-title {
  display: inline-flex;
  align-items: center;
  gap: 7px;
  font-size: 15px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
}

.ad-title .el-icon {
  color: var(--accent, #2563eb);
}

.ad-hint {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  color: var(--ink2, #5a6a80);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ad-close {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  border: 0;
  border-radius: 7px;
  background: transparent;
  /* 纯图标控件常态色需 ≥3:1，--ink3 在亮底只有 2.5:1 */
  color: var(--ink2, #5a6a80);
  font-size: 15px;
  cursor: pointer;
  transition: background .15s, color .15s;
}

.ad-close:hover {
  background: var(--accent-soft, #eef4ff);
  color: var(--accent, #2563eb);
}

.ad-body {
  height: 100%;
  min-height: 0;
  /* Agent 内部有自己的面板配色与内边距，这里只负责给尺寸 */
  display: flex;
  /* ⚠ 必须显式撑满：Agent.vue 的根里还带一个 display:none 的 el-overlay，
     若用默认的 flex:0 1 auto，可用宽度会按「隐藏节点的固有宽度」算，
     实测弹窗 1080px 时 ad-body 只拿到 485px，右侧出现一大片空白。 */
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
}

.ad-body > * {
  flex: 1;
  min-width: 0;
}
</style>

<style>
/*
 * 弹窗外壳交给三层 flex：头部固定，Agent 主体吃掉剩余高度。
 * ⚠ .el-dialog / .el-dialog__body 由 Element Plus 渲染、不带组件的 data-v 属性，
 *   只能写在非 scoped 全局块里。
 * ⚠ body 必须 overflow:hidden —— Agent 内部各栏自己滚动，外层再滚会套两层滚动条。
 */
.agent-dialog-shell {
  display: flex;
  flex-direction: column;
  height: 86vh;
  /* 覆盖 EP 默认的 15vh 上边距：15vh + 86vh 会超出视口 */
  margin: 7vh auto;
}

.agent-dialog-shell .el-dialog__header {
  flex: 0 0 auto;
  margin-right: 0;
  padding-bottom: 12px;
}

.agent-dialog-shell .el-dialog__body {
  flex: 1 1 auto;
  min-height: 0;
  padding-top: 0;
  overflow: hidden;
  display: flex;
}

/* Agent 的双栏在弹窗里比全屏窄，两侧留白收到最小 */
.agent-dialog-shell .ad-body {
  padding: 0 4px 4px;
}

/* 暗色下 EP 只引入亮色变量，弹窗底色与分隔线需要显式给 */
:root[data-theme='dark'] .agent-dialog-shell {
  background: #0d1117;
}

:root[data-theme='dark'] .agent-dialog-shell .el-dialog__header {
  border-bottom-color: rgba(59, 130, 246, 0.12);
}
</style>
