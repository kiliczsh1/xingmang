# Write.vue 拆分解耦方案

## 现状分析

Write.vue 当前 **14130 行**，三个部分严重膨胀：

| 区域 | 行数 | 占比 | 内容 |
|------|------|------|------|
| `<template>` | ~2424 | 17% | 主布局 + 19 个 el-dialog + 3 个面板 + 角色库弹窗 |
| `<script>` | ~5076 | 36% | 所有逻辑混在一起，无模块划分 |
| `<style scoped>` | ~6628 | 47% | 所有 CSS 在一个 scoped 块 |

其他文件对比：

| 文件 | 行数 | 可读性 |
|------|------|--------|
| CharacterLibrary.vue | ~1643 | ✅ 正常 |
| `src/views/` 下其他页面 | 300-1500 | ✅ 正常 |
| **Write.vue** | **14130** | ❌ 屎山临界 |

---

## 拆解策略

**原则：** 每次只拆一个独立模块，不改内部逻辑，不改现有行为。从大到小、从"视觉独立"到"逻辑独立"逐步推进。

分为三阶段，**阶段一做完就已经解决 70% 的问题**。

---

## 阶段一：弹窗/面板组件化（优先，减重 ~4000 行）

这是收益最高的拆分。Write.vue 模板中有 **19 个 el-dialog** 和 **3 个独立面板**，每个都是视觉独立的模块，直接拆为子组件。

### 1.1 弹窗组件清单

| # | 当前位置（行） | 组件名 | 估行数（template+script+style） |
|---|--------------|--------|-------------------------------|
| ① | ~528-545 | `ModelConfigDialog.vue` | ~150 |
| ② | ~548-635 | `ModelSelectDialog.vue` | ~300 |
| ③ | ~638-817 | `PromptSelectDialog.vue` | ~300 |
| ④ | ~820-837 | `MoveChapterDialog.vue` | ~100 |
| ⑤ | ~840-857 | `EditFolderDialog.vue` | ~80 |
| ⑥ | ~860-950 | `BatchCreateChaptersDialog.vue` | ~200 |
| ⑦ | ~968-995 | `ExportDialog.vue` | ~150 |
| ⑧ | ~997-1340 | `HistoryDialog.vue` | ~500 |
| ⑨ | ~1343-1349 | `RenameDialog.vue` | ~60 |
| ⑩ | ~1352-1434 | `ImportChapterDialog.vue` | ~250 |
| ⑪ | ~1437-1489 | 其他设置弹窗 ×2 | ~100 |
| ⑫ | ~1492-1544 | `GlobalMemoDialog.vue` | ~300 |
| ⑬ | ~1547-1579 | 其他弹窗 | ~80 |
| ⑭ | ~1582-1614 | 其他弹窗 | ~80 |
| ⑮ | ~1617-1628 | `VolumeDialog.vue` | ~80 |

**这些合计约 2700 行 template + 1300 行 style，拆出去后 Write.vue 直接减重 ~4000 行。**

### 1.2 面板组件清单

| # | 面板 | 当前处理方式 | 建议 |
|---|------|-------------|------|
| ① | AI 对话面板（ChatPanel） | ✅ 已拆为 `@/components/ChatPanel/index.vue` | 不动 |
| ② | AI 写作面板 | 内嵌在 Write.vue 中 | 拆为 `AiWritingPanel.vue` |
| ③ | 知识图谱面板 | 已用 `<KnowledgeGraph>` 组件 | 不动 |

### 1.3 操作方法

以 `ImportChapterDialog.vue` 为例：

```vue
<!-- 新建 src/components/WriteDialogs/ImportChapterDialog.vue -->
<template>
  <el-dialog v-model="visible" title="导入章节" width="700px" append-to-body>
    ...原模板内容...
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { chapterAPI } from '@/api'

const props = defineProps<{
  bookId: number
}>()

const visible = defineModel<boolean>('visible')
// ...原 ImportChapterDialog 的所有逻辑（从 Write.vue 搬过来）...
</script>

<style scoped>
/* 原 ImportChapterDialog 的样式（从 Write.vue 搬过来）*/
</style>
```

在 Write.vue 中原位置替换为：

```vue
<ImportChapterDialog
  v-model:visible="importChapterDialogVisible"
  :book-id="bookId"
  @imported="fetchChapters"
/>
```

---

## 阶段二：逻辑 Composable 化（进一步减重，拆 ~2000 行）

把 script 中按功能域拆到独立的 composable 文件，Write.vue 只剩 import + 组合调用。

### 2.1 Composable 拆分方案

```
src/composables/
  useChapterManager.ts     # 章节 CRUD、排序、拖拽、文件夹管理（~800 行）
  useMemoManager.ts        # 备忘录 CRUD、批量操作（~300 行）
  useEditor.ts             # 编辑器状态、字体、字号、选中文本、续写（~400 行）
  useAutoSave.ts           # 自动保存逻辑（~100 行）
  useContinueWrite.ts      # 续写功能（~400 行）
  useAIWriting.ts          # AI 写作面板逻辑（~300 行）
  useAIChat.ts             # AI 对话面板逻辑（~300 行）
  useKeybindings.ts        # 快捷键绑定（~200 行）
  useModalState.ts         # 所有弹窗/面板的可见性状态（~100 行）
  useModelConfig.ts        # 模型选择/配置逻辑（~400 行）
  usePromptPicker.ts       # 提示词选择逻辑（~200 行）
```

### 2.2 示例：useChapterManager.ts

```ts
// src/composables/useChapterManager.ts
import { ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { chapterAPI, volumeAPI } from '@/api'
import type { Chapter, Volume } from '@/types'

export function useChapterManager(bookId: computed<number>) {
  // 状态
  const chaptersList = ref<Chapter[]>([])
  const currentChapter = ref<Chapter | null>(null)
  const catalogType = ref('chapters')
  const volumes = ref<Volume[]>([])
  const isDescending = ref(false)
  const expandedFolderIds = ref<Set<number>>(new Set())
  const draggingChapterId = ref<number | null>(null)

  // 计算属性
  const sortedChapters = computed(() => { ... })
  const rootFiles = computed(() => ...)
  const rootFolders = computed(() => ...)

  // 方法
  const fetchChapters = async () => { ... }
  const selectChapter = (chapter: Chapter) => { ... }
  const createChapter = async (parentId?: number | null) => { ... }
  const saveChapter = async () => { ... }
  const deleteChapter = async (chapter: Chapter) => { ... }
  const toggleFolder = (id: number) => { ... }
  const toggleChapterOrder = () => { ... }
  const handleChapterDragStart = (id: number) => { ... }
  const handleChapterDrop = (targetId: number) => { ... }
  // ...所有章节相关方法

  return {
    chaptersList, currentChapter, catalogType,
    volumes, isDescending, expandedFolderIds, draggingChapterId,
    sortedChapters, rootFiles, rootFolders,
    fetchChapters, selectChapter, createChapter, saveChapter,
    deleteChapter, toggleFolder, toggleChapterOrder,
    handleChapterDragStart, handleChapterDrop,
    // ...其他
  }
}
```

### 2.3 改造后 Write.vue script 减至 ~300 行

```ts
<script setup lang="ts">
import { useRouter, useRoute } from 'vue-router'
import { useBookStore } from '@/stores/book'
import { useChapterManager } from '@/composables/useChapterManager'
import { useMemoManager } from '@/composables/useMemoManager'
import { useEditor } from '@/composables/useEditor'
import { useAutoSave } from '@/composables/useAutoSave'
import { useModalState } from '@/composables/useModalState'
// ...其他 import

const route = useRoute()
const router = useRouter()
const bookId = computed(() => parseInt(route.params.bookId as string))

// 各功能模块组合（像搭积木一样）
const chapters = useChapterManager(bookId)
const memos = useMemoManager(bookId)
const editor = useEditor()
const autoSave = useAutoSave(bookId, chapters.currentChapter, editor.content)
const modals = useModalState()

// 需要跨模块交互的连接代码（约 30 行）
const handleCreateChapter = () => { chapters.createChapter(); modals.showVolumeDialog = false }
const handleImportCompleted = () => { chapters.fetchChapters(); modals.importChapterDialogVisible = false }
</script>
```

---

## 阶段三：样式拆分（最后做）

Style 的 6628 行是最终极的拆分。

### 3.1 方案

#### 方式 A：样式跟随组件（推荐）

阶段一拆出子组件时，把对应的样式**一起带走**。拆完所有弹窗/面板后，Write.vue 的 style 至少减半。

#### 方式 B：css 文件分模块

剩余的通用样式按功能拆分 CSS 文件，在 Write.vue 中 import：

```vue
<!-- 在 Write.vue 的 <style scoped> 中 @import -->
<style scoped>
@import './write/toolbar.css';
@import './write/layout.css';
@import './write/catalog.css';
@import './write/editor.css';
@import './write/chat.css';
@import './write/dialogs.css';
@import './write/dark-theme.css';
</style>
```

或使用 Vite 的 glob import（不通过 scoped，全局 CSS）：

```ts
// main.ts 或 Write.vue 的 <style> 中
import './styles/write/toolbar.css'
import './styles/write/layout.css'
```

---

## 优先级建议

| 优先级 | 任务 | 减行数 | 难度 | 风险 |
|--------|------|--------|------|------|
| 🥇 | **拆弹窗组件 × 15 个** | **~4000** | ⭐⭐ | 低 — 视觉独立，纯搬运 |
| 🥇 | **拆 AI 写作面板** | **~500** | ⭐⭐ | 低 — 已有 ChatPanel 可参考 |
| 🥈 | **拆 composable × 10 个** | **~2000** | ⭐⭐⭐ | 中 — 需梳理跨模块依赖 |
| 🥉 | **样式拆分** | **~3000** | ⭐⭐ | 低 — 随组件走，剩余按文件拆 |

**阶段一做完 Write.vue 就从 14130 行降到 ~10000 行**，再做阶段二降到 ~5000 行（含模板和样式），已经是一个健康的中型组件。

---

## 不会影响功能的关键措施

1. **先新建文件，再逐步替换** — 组件建在新目录 `src/components/WriteDialogs/`，不影响原来
2. **defineModel 双向绑定** — `v-model:visible` 接管弹窗显隐，和原来 `v-model="showXxx"` 等效
3. **事件冒泡** — 子组件 `emit('xxx')`，Write.vue 监听，原有逻辑链不变
4. **composable 按函数提取** — 不改变函数签名，只搬位置
5. **每个拆完后验证** — 弹窗打开/关闭/功能正常再拆下一个

建议**先拆一个小的（如 RenameDialog 或 VolumeDialog）**，理顺流程后再批量拆剩下的 14 个弹窗。
