<template>
  <el-dialog
    v-model="visible"
    width="780px"
    class="changelog-dialog"
    append-to-body
    destroy-on-close
  >
    <template #header>
      <div class="cl-header">
        <div class="cl-header-text">
          <span class="cl-title">更新日志</span>
          <span class="cl-sub">
            按时间倒序 · 依据仓库提交与打包产物整理
            <i class="cl-basis cl-basis-inline">补号</i>
            表示该版本号按时间线补入，改动内容取自实际提交
          </span>
        </div>
        <span v-if="currentVersion" class="cl-current">
          当前版本 <b>{{ currentVersion }}</b>
        </span>
      </div>
    </template>

    <!-- 展开 / 收起。收起即简略版，展开即详略版；单条也可以单独点开 -->
    <div class="cl-toolbar">
      <div class="cl-modes" role="group" aria-label="显示粒度">
        <button
          type="button"
          class="cl-mode"
          :class="{ active: !allExpanded }"
          @click="collapseAll"
        >
          简略版
        </button>
        <button
          type="button"
          class="cl-mode"
          :class="{ active: allExpanded }"
          @click="expandAll"
        >
          详略版
        </button>
      </div>
      <span class="cl-count">{{ releases.length }} 个版本</span>
    </div>

    <div class="cl-scroll">
      <div class="cl-list">
        <section v-for="rel in releases" :key="rel.version" class="cl-card">
          <button
            type="button"
            class="cl-card-head"
            :aria-expanded="isExpanded(rel.version)"
            @click="toggle(rel.version)"
          >
            <el-icon class="cl-chevron" :class="{ open: isExpanded(rel.version) }">
              <ArrowRight />
            </el-icon>
            <span class="cl-ver">{{ rel.version }}</span>
            <span class="cl-date">{{ rel.date }}</span>
            <span
              v-if="rel.basis === 'derived'"
              class="cl-basis"
              title="该版本号按时间线补入，改动内容取自实际提交"
            >补号</span>
            <span class="cl-summary">{{ rel.summary }}</span>
          </button>

          <!-- 收起态 = 简略版 -->
          <ul v-if="!isExpanded(rel.version)" class="cl-brief">
            <li v-for="(line, i) in rel.brief" :key="i">{{ line }}</li>
          </ul>

          <!-- 展开态 = 详略版 -->
          <template v-else>
            <div class="cl-commit-line">commit {{ rel.commit }}</div>
            <ul class="cl-detail">
              <li v-for="(item, i) in rel.details" :key="i">
                <span class="cl-tag" :class="tagClass(item.tag)">{{ item.tag }}</span>
                <span class="cl-text">{{ item.text }}</span>
              </li>
            </ul>
          </template>
        </section>
      </div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ArrowRight } from '@element-plus/icons-vue'
import { changelogReleases } from '@/data/changelog'

const props = withDefaults(
  defineProps<{
    visible: boolean
    currentVersion?: string
  }>(),
  { currentVersion: '' }
)

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const visible = computed({
  get: () => props.visible,
  set: value => emit('update:visible', value)
})

const releases = changelogReleases

/**
 * 已展开的版本号集合。默认全空 = 全部收起 = 简略版。
 * 用 Set 而不是在数据源上挂 isOpen 字段：更新日志是只读的静态数据，
 * 不该被视图状态污染。
 */
const expanded = ref<Set<string>>(new Set())

const allExpanded = computed(() => expanded.value.size === releases.length)

const isExpanded = (version: string) => expanded.value.has(version)

/** Set 是引用类型，替换实例才能触发依赖更新 */
const toggle = (version: string) => {
  const next = new Set(expanded.value)
  if (next.has(version)) {
    next.delete(version)
  } else {
    next.add(version)
  }
  expanded.value = next
}

const expandAll = () => {
  expanded.value = new Set(releases.map(r => r.version))
}

const collapseAll = () => {
  expanded.value = new Set()
}

/**
 * 标签配色。亮 / 暗两套值都写在 CSS 里，这里只负责给出类名，
 * 避免在 JS 里判主题（主题可能由系统偏好随时切换）。
 */
const tagClass = (tag: string) => {
  const known = ['功能', '界面', '重构', '修复', '工具链', '清理', '基线']
  return known.includes(tag) ? `tag-${known.indexOf(tag)}` : 'tag-other'
}
</script>

<style scoped>
.cl-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.cl-header-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.cl-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--app-color, #1a2230);
}

.cl-sub {
  font-size: 12px;
  color: var(--ink2, #5a6a80);
}

.cl-current {
  flex-shrink: 0;
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: var(--accent, #2563eb);
  background: var(--accent-soft, #eef4ff);
  white-space: nowrap;
}

.cl-current b {
  font-weight: 700;
}

/* ---------- 展开 / 收起 ---------- */
.cl-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.cl-modes {
  display: inline-flex;
  gap: 3px;
  padding: 3px;
  border-radius: 9px;
  background: var(--accent-soft, #eef4ff);
}

.cl-mode {
  border: 0;
  padding: 4px 14px;
  border-radius: 7px;
  background: transparent;
  color: var(--ink2, #5a6a80);
  font-size: 12.5px;
  line-height: 1.6;
  cursor: pointer;
  transition: background .15s, color .15s;
}

.cl-mode:hover {
  color: var(--accent, #2563eb);
}

.cl-mode.active {
  background: var(--card, #ffffff);
  color: var(--accent, #2563eb);
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(30, 36, 51, .08);
}

.cl-count {
  flex-shrink: 0;
  font-size: 12px;
  /* 用 --ink2 而非 --ink3：ink3 在亮底上只有 2.5:1，这段文字太小看不清 */
  color: var(--ink2, #5a6a80);
  font-variant-numeric: tabular-nums;
}

/* ---------- 列表 ---------- */
/* 撑满 body 的剩余高度（body 是 flex column），成为唯一滚动容器 */
.cl-scroll {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  overscroll-behavior: contain;
}

.cl-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 2px 2px 4px;
}

.cl-card {
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid var(--topbar-border, #e6eaf0);
  background: var(--card, #ffffff);
  transition: border-color .15s;
}

.cl-card:hover {
  border-color: var(--accent, #2563eb);
}

/* 整行可点：点标题行展开 / 收起该版本 */
.cl-card-head {
  display: flex;
  align-items: baseline;
  gap: 10px;
  flex-wrap: wrap;
  width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  font: inherit;
  color: inherit;
  text-align: left;
  cursor: pointer;
}

.cl-chevron {
  align-self: center;
  flex-shrink: 0;
  font-size: 13px;
  /* 纯图标控件常态色需 ≥3:1；--ink3 在亮底只有 2.5:1，故用 --ink2 */
  color: var(--ink2, #5a6a80);
  transition: transform .18s ease;
}

.cl-chevron.open {
  transform: rotate(90deg);
}

.cl-ver {
  font-size: 14px;
  font-weight: 700;
  color: var(--accent, #2563eb);
  font-variant-numeric: tabular-nums;
}

.cl-date {
  font-size: 12px;
  color: var(--ink2, #5a6a80);
  font-variant-numeric: tabular-nums;
}

/* 弹窗头部那句图例里的补号标记：跟着文字走，不要 flex-shrink */
.cl-basis-inline {
  flex-shrink: 1;
  margin: 0 2px;
  font-style: normal;
}

/* 补号标记：提示该版本号是按时间线补入的，改动内容仍取自实际提交 */
.cl-basis {
  flex-shrink: 0;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  line-height: 1.6;
  color: #92400e;
  background: #fef3c7;
  border: 1px solid #fcd34d;
}

:root[data-theme='dark'] .cl-basis {
  color: #f0b869;
  background: rgba(245, 158, 11, .14);
  border-color: rgba(245, 158, 11, .3);
}

.cl-summary {
  font-size: 12.5px;
  color: var(--ink2, #5a6a80);
}

.cl-commit-line {
  margin-top: 8px;
  font-family: ui-monospace, Consolas, monospace;
  font-size: 11.5px;
  color: var(--ink2, #5a6a80);
}

.cl-brief {
  margin: 10px 0 0;
  padding-left: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cl-brief li {
  font-size: 13px;
  line-height: 1.7;
  color: var(--app-color, #1a2230);
}

.cl-detail {
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 9px;
}

.cl-detail li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.cl-tag {
  flex-shrink: 0;
  margin-top: 2px;
  padding: 1px 7px;
  border-radius: 5px;
  font-size: 11.5px;
  line-height: 1.6;
  white-space: nowrap;
}

.cl-text {
  font-size: 13px;
  line-height: 1.72;
  color: var(--app-color, #1a2230);
}

/* 标签配色 —— 亮色 */
.cl-tag.tag-0 { color: #1d4ed8; background: #e6efff; }
.cl-tag.tag-1 { color: #6d28d9; background: #f1ebfe; }
.cl-tag.tag-2 { color: #0f766e; background: #e2f5f2; }
.cl-tag.tag-3 { color: #15803d; background: #e6f6ea; }
.cl-tag.tag-4 { color: #b45309; background: #fdf1e0; }
.cl-tag.tag-5 { color: #4b5563; background: #eef0f3; }
.cl-tag.tag-6 { color: #4b5563; background: #eef0f3; }
.cl-tag.tag-other { color: #4b5563; background: #eef0f3; }

/* 滚动条：亮色滚动条出现在深色弹窗里很扎眼，压暗一档 */
.cl-scroll::-webkit-scrollbar {
  width: 8px;
}

.cl-scroll::-webkit-scrollbar-thumb {
  border-radius: 4px;
  background: rgba(148, 163, 184, 0.35);
}

/* 展开 / 收起按钮 —— 暗色
 * ⚠ 项目只引入 EP 亮色变量，--card / --accent-soft 在暗色下的差值太小（#161b22 vs #12253f），
 *   按钮和轨道会糊在一起，所以这一组显式给值。 */
:root[data-theme='dark'] .cl-modes {
  background: rgba(148, 163, 184, .10);
}

:root[data-theme='dark'] .cl-mode {
  color: #8b98a8;
}

:root[data-theme='dark'] .cl-mode:hover {
  color: #c9d4e2;
}

:root[data-theme='dark'] .cl-mode.active {
  background: rgba(88, 166, 255, .18);
  color: #58a6ff;
  box-shadow: none;
}

:root[data-theme='dark'] .cl-scroll::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.25);
}

:root[data-theme='dark'] .cl-card {
  background: #1c2333;
  border-color: rgba(59, 130, 246, 0.14);
}

:root[data-theme='dark'] .cl-card:hover {
  border-color: rgba(88, 166, 255, 0.34);
}

/* 标签配色 —— 暗色（亮底浅色标签在深色弹窗里会刺眼，改为低透明底 + 提亮文字） */
:root[data-theme='dark'] .cl-tag.tag-0 { color: #7fb0ff; background: rgba(59, 130, 246, .16); }
:root[data-theme='dark'] .cl-tag.tag-1 { color: #b79bff; background: rgba(139, 92, 246, .16); }
:root[data-theme='dark'] .cl-tag.tag-2 { color: #5ed3c4; background: rgba(20, 184, 166, .14); }
:root[data-theme='dark'] .cl-tag.tag-3 { color: #6ee7a0; background: rgba(34, 197, 94, .14); }
:root[data-theme='dark'] .cl-tag.tag-4 { color: #f0b869; background: rgba(245, 158, 11, .14); }
:root[data-theme='dark'] .cl-tag.tag-5 { color: #9aa7b8; background: rgba(148, 163, 184, .14); }
:root[data-theme='dark'] .cl-tag.tag-6 { color: #9aa7b8; background: rgba(148, 163, 184, .14); }
:root[data-theme='dark'] .cl-tag.tag-other { color: #9aa7b8; background: rgba(148, 163, 184, .14); }
</style>

<style>
/*
 * 弹窗外壳的三层 flex：让标题与按钮固定，只有列表滚动。
 *   .el-dialog(flex column, 限高) → .el-dialog__body(flex:1, 不滚) → .cl-scroll(唯一滚动容器)
 * ⚠ body 必须 overflow:hidden，否则它会成为滚动容器，标题和按钮照样滚走。
 * ⚠ .el-dialog 等元素由 Element Plus 渲染、不带组件的 data-v 属性，
 *   这部分只能写在非 scoped 全局块里（scoped 会编译出 [data-v] 而整条失效）。
 *
 * .changelog-dialog 是 ElDialog 的 class 透传到 .el-dialog 上，指的就是弹窗自身。
 */
.changelog-dialog {
  display: flex;
  flex-direction: column;
  max-height: 86vh;
  /* 覆盖 EP 默认的 15vh 上边距：15vh + 86vh 会超出视口 */
  margin: 7vh auto;
}

.changelog-dialog .el-dialog__header {
  flex: 0 0 auto;
}

.changelog-dialog .el-dialog__body {
  flex: 1 1 auto;
  min-height: 0;
  padding-top: 6px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
