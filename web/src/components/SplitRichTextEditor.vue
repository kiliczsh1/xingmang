<template>
  <div
    :id="editorId"
    class="split-rich-text-editor"
    :class="{ 'is-disabled': disabled }"
  ></div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Vditor from 'vditor'
import 'vditor/dist/index.css'

const props = withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
  disabled?: boolean
  fontFamily?: string
  fontSize?: number
  textIndent?: boolean
  paragraphSpacing?: number
  paragraphLine?: boolean
  roleMarks?: string[]
  termMarks?: string[]
}>(), {
  placeholder: '请输入简介内容',
  disabled: false,
  fontFamily: 'Microsoft YaHei',
  fontSize: 16,
  textIndent: false,
  paragraphSpacing: 0,
  paragraphLine: false,
  roleMarks: () => [],
  termMarks: () => [],
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'blur'): void
  (e: 'mark-click', payload: { type: 'role' | 'term'; name: string }): void
  (e: 'open-find'): void
  (e: 'open-replace'): void
}>()

const editorId = `vditor-${Math.random().toString(36).slice(2, 10)}`
const editorRef = ref<Vditor | null>(null)
let settingFromExternal = false
// 记录最后一次外部同步写入的值，用于识别「Vditor 回显」事件并忽略它，
// 避免把回显误当成用户输入，也避免标志残留吞掉用户输入。
let lastExternalValue: string | null = null
let externalResetTimer: ReturnType<typeof setTimeout> | null = null
let initialized = false
let markTimer: ReturnType<typeof setTimeout> | null = null
// 上次 input 回调收到的完整值，用于计算本次输入增量
let lastInputValue = ''
// 计算本次输入相对上次的新增部分，检测是否包含汉字（CJK）。
// 拼音输入法组合期间输入的是拉丁字母（如 "nihao"），汉字上屏时才出现汉字，
// 因此只在增量含汉字时重建标记，拼音字母期间不触发 refreshMarks，避免打断输入法。
const hasNewHan = (cur: string): boolean => {
  const prev = lastInputValue
  lastInputValue = cur
  if (!cur || cur === prev) return false
  // 找公共前后缀，中间的差异即本次插入的字符
  let i = 0
  while (i < prev.length && i < cur.length && prev[i] === cur[i]) i++
  let j = 0
  while (j < prev.length - i && j < cur.length - i && prev[prev.length - 1 - j] === cur[cur.length - 1 - j]) j++
  const inserted = cur.slice(i, cur.length - j)
  return /[\u4e00-\u9fff]/.test(inserted)
}

const hasHtmlTag = (value: string) => /<\/?[a-z][\s\S]*>/i.test(value)

const isDark = () =>
  document.documentElement.getAttribute('data-theme') === 'dark'

// 本地内置中文语言包：传 i18n 后 Vditor 跳过 CDN 动态加载（不依赖外网），
// 工具栏按钮的中文提示文案即来源于此
const ZH_I18N: Record<string, string> = {
  bold: '粗体',
  headings: '标题',
  italic: '斜体',
  strike: '删除线',
  list: '无序列表',
  'ordered-list': '有序列表',
  quote: '引用',
  link: '链接',
  'inline-code': '行内代码',
  code: '代码块',
  undo: '撤销',
  redo: '重做',
  fullscreen: '全屏切换',
  preview: '预览',
  both: '编辑 & 预览',
  check: '任务列表',
  table: '表格',
  indent: '列表缩进',
  outdent: '列表反向缩进',
  emoji: '表情',
  more: '更多',
  heading1: '一级标题',
  heading2: '二级标题',
  heading3: '三级标题',
  heading4: '四级标题',
  heading5: '五级标题',
  heading6: '六级标题',
  edit: '编辑',
  'edit-mode': '切换编辑模式',
  export: '导出',
  outline: '大纲',
  line: '分隔线',
  imageURL: '图片地址',
  tip: '提示',
  help: '帮助',
  info: '关于',
  wysiwyg: '所见即所得',
  close: '关闭',
  copy: '复制',
  down: '下',
  up: '上',
  remove: '删除',
  update: '更新',
}

// 精简后的工具栏：去除低频的 emoji / 缩进调整 / 任务列表 / 表格 / 分屏，
// 保留写作核心能力，并加入查找 / 替换（记事本式）
const toolbar = [
  'headings',
  'bold',
  'italic',
  'strike',
  '|',
  'list',
  'ordered-list',
  '|',
  'quote',
  'link',
  'inline-code',
  'code',
  '|',
  'undo',
  'redo',
  '|',
  {
    name: 'find',
    icon: '<svg viewBox="0 0 24 24"><circle cx="10.5" cy="10.5" r="6"/><path d="M15 15l5.5 5.5"/></svg>',
    tip: '查找 (Ctrl+F)',
    click: () => emit('open-find'),
  },
  {
    name: 'replace',
    icon: '<svg viewBox="0 0 24 24"><path d="M4 17l7-10M8 17l4-5.5M15 8h5M15 8l2-2M15 8l2 2M15 14h5M15 14l2-2M15 14l2 2"/></svg>',
    tip: '替换 (Ctrl+H)',
    click: () => emit('open-replace'),
  },
  '|',
  'fullscreen',
  'preview',
]

/* 重新设计的工具栏图标：直观线框风格，替换 Vditor 默认抽象图标。
   viewBox 统一 0 0 24 24，stroke:currentColor 风格，与 CSS 中
   .vditor-tooltipped svg { fill:none; stroke:currentColor; stroke-width:1.8 } 配合。 */
const ICON_MAP: Record<string, string> = {
  // 标题：H 字母 + 上下箭头表示可分级
  headings: '<svg viewBox="0 0 24 24"><path d="M4 6h16M12 6v12M9 18h6"/></svg>',
  // 粗体：实心 B 字形
  bold: '<svg viewBox="0 0 24 24"><path d="M7 5h6a3.5 3.5 0 0 1 0 7H7zM7 12h7a3.5 3.5 0 0 1 0 7H7z"/></svg>',
  // 斜体：倾斜的 I
  italic: '<svg viewBox="0 0 24 24"><path d="M10 5h9M6 19h9M14 5l-4 14"/></svg>',
  // 删除线：横线穿过 S
  strike: '<svg viewBox="0 0 24 24"><path d="M5 12h14M8 7c.5-1.2 2.2-2 4-2 2.5 0 4 1.2 4 3 0 3.5-7 2.5-7 6 0 1.8 1.5 3 4 3 1.8 0 3.5-.8 4-2"/></svg>',
  // 无序列表：三个圆点 + 横线
  list: '<svg viewBox="0 0 24 24"><circle cx="5" cy="6.5" r="1.4"/><circle cx="5" cy="12" r="1.4"/><circle cx="5" cy="17.5" r="1.4"/><path d="M9.5 6.5h10M9.5 12h10M9.5 17.5h10"/></svg>',
  // 有序列表：1.2.3. + 横线
  'ordered-list': '<svg viewBox="0 0 24 24"><path d="M4.5 5.5h1v3M5 5.5l-1 .6M5.2 13.2c-.6-.3-1-.9-1-1.5 0-.9.8-1.5 1.6-1.5.7 0 1.3.4 1.5 1l-2.4 2.6H7M4.5 18h2.4c0 .8-.6 1.5-1.4 1.5-.6 0-1.1-.3-1.4-.8M10 6.5h10M10 12h10M10 17.5h10"/></svg>',
  // 引用：引号块
  quote: '<svg viewBox="0 0 24 24"><path d="M4 5h5v7H6c0 3 1.5 5 4 6M14 5h5v7h-3c0 3 1.5 5 4 6"/></svg>',
  // 链接：锁链
  link: '<svg viewBox="0 0 24 24"><path d="M10 14a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.5 1.5M14 10a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7l1.5-1.5"/></svg>',
  // 行内代码：< >
  'inline-code': '<svg viewBox="0 0 24 24"><path d="M8 7l-5 5 5 5M16 7l5 5-5 5M13.5 5l-3 14"/></svg>',
  // 代码块：带边框的 </>
  code: '<svg viewBox="0 0 24 24"><rect x="4" y="5" width="16" height="14" rx="2"/><path d="M9 9.5L7 12l2 2.5M15 9.5l2 2.5-2 2.5M12 8.5l-1.5 7"/></svg>',
  // 撤销：左弯箭头
  undo: '<svg viewBox="0 0 24 24"><path d="M9 7L5 11l4 4M5 11h9a5 5 0 0 1 0 10h-2"/></svg>',
  // 重做：右弯箭头
  redo: '<svg viewBox="0 0 24 24"><path d="M15 7l4 4-4 4M19 11h-9a5 5 0 0 0 0 10h2"/></svg>',
  // 全屏：四角扩张
  fullscreen: '<svg viewBox="0 0 24 24"><path d="M4 9V4h5M15 4h5v5M20 15v5h-5M9 20H4v-5"/></svg>',
  // 预览：眼睛
  preview: '<svg viewBox="0 0 24 24"><path d="M2.5 12S6 5.5 12 5.5 21.5 12 21.5 12 18 18.5 12 18.5 2.5 12 2.5 12z"/><circle cx="12" cy="12" r="3"/></svg>',
}

/** 在 after 回调中替换 Vditor 默认图标为重新设计的直观图标 */
const applyCustomIcons = () => {
  document.querySelectorAll(`#${editorId} .vditor-toolbar .vditor-tooltipped`).forEach(el => {
    const btn = el as HTMLElement
    const type = btn.getAttribute('data-type')
    if (type && ICON_MAP[type]) {
      btn.innerHTML = ICON_MAP[type]
    }
  })
}

// 按实例 id 作用域，避免同页多个编辑器（写作页 + 弹窗简介）样式互相污染。
// 注意：Vditor 初始化时 initUI 会执行 vditor.element.innerHTML = "" 清空挂载元素，
// 因此样式不能放在模板的 <style> 标签里（会被清掉），改为注入 document.head 由 JS 管理。
// 自带工具栏已由父组件的合并单行替代，这里整体隐藏（快捷键不受影响）。
const fontStyle = computed(() => `
  #${editorId} .vditor-toolbar {
    display: none !important;
  }
  #${editorId} .vditor-reset,
  #${editorId} .vditor-wysiwyg,
  #${editorId} .vditor-reset p {
    font-family: ${props.fontFamily} !important;
    font-size: ${props.fontSize}px !important;
    line-height: 2 !important;
  }
  #${editorId} .vditor-wysiwyg p {
    ${props.textIndent ? 'text-indent: 2em !important;' : ''}
    ${props.paragraphSpacing > 0 ? `margin-bottom: ${props.paragraphSpacing}px !important;` : ''}
  }
  #${editorId} .vditor-wysiwyg p.xm-paragraph-line {
    border-bottom: 1px solid rgba(120, 120, 120, .25);
    padding-bottom: 10px;
  }
  #${editorId} .vditor-wysiwyg p.xm-paragraph-line:last-child {
    border-bottom: none;
  }
  #${editorId} .vditor-wysiwyg .xm-mark {
    cursor: pointer;
    border-radius: 3px;
    padding: 0 2px;
    transition: background .15s;
  }
  #${editorId} .vditor-wysiwyg .xm-mark-role {
    color: #b45309;
    background: rgba(245, 158, 11, .14);
  }
  #${editorId} .vditor-wysiwyg .xm-mark-role:hover {
    background: rgba(245, 158, 11, .3);
  }
  #${editorId} .vditor-wysiwyg .xm-mark-term {
    color: #1d4ed8;
    background: rgba(59, 130, 246, .12);
  }
  #${editorId} .vditor-wysiwyg .xm-mark-term:hover {
    background: rgba(59, 130, 246, .28);
  }
  :root[data-theme='dark'] #${editorId} .vditor-wysiwyg p.xm-paragraph-line {
    border-bottom-color: rgba(148, 163, 184, .3);
  }
  :root[data-theme='dark'] #${editorId} .vditor-wysiwyg .xm-mark-role {
    color: #fbbf24;
    background: rgba(245, 158, 11, .18);
  }
  :root[data-theme='dark'] #${editorId} .vditor-wysiwyg .xm-mark-role:hover {
    background: rgba(245, 158, 11, .32);
  }
  :root[data-theme='dark'] #${editorId} .vditor-wysiwyg .xm-mark-term {
    color: #93c5fd;
    background: rgba(59, 130, 246, .16);
  }
  :root[data-theme='dark'] #${editorId} .vditor-wysiwyg .xm-mark-term:hover {
    background: rgba(59, 130, 246, .3);
  }
`)

// 注入到 document.head 的 style 元素（Vditor 清空挂载元素不影响这里）
const styleEl = document.createElement('style')
styleEl.dataset.editorStyle = editorId
document.head.appendChild(styleEl)

watch(
  fontStyle,
  (css) => {
    styleEl.textContent = css
  },
  { immediate: true }
)

const toMd = (value: string): string => {
  if (!value) return ''
  if (!hasHtmlTag(value)) return value
  try {
    return editorRef.value ? editorRef.value.html2md(value) : value
  } catch {
    return value
  }
}

const syncContent = (value: string) => {
  if (!editorRef.value) return
  const md = toMd(value || '')
  if (editorRef.value.getValue() !== md) {
    settingFromExternal = true
    lastExternalValue = md
    editorRef.value.setValue(md)
    // 兜底复位：若 Vditor 的回显 input 事件没有单独触发（新建章节时它常与
    // 用户输入合并成一次事件），标志绝不能残留，否则会吞掉用户输入导致内容无法保存。
    if (externalResetTimer) clearTimeout(externalResetTimer)
    externalResetTimer = setTimeout(() => {
      settingFromExternal = false
      lastExternalValue = null
    }, 1200)
  }
}

/* ===== 正文标记（角色名 / 词条名 / 横线） =====
   在 Vditor 的 contentEditable 上做纯展示层标记：
   - 不修改 markdown 源文本（保存内容不受影响）
   - 标记随输入/切换章节自动重建 */
const escapeReg = (s: string) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const applyMarks = () => {
  const vd = editorRef.value
  // Vditor 实例的 wysiwyg DOM 在 vd.vditor.wysiwyg.element（内部 IVditor）
  const root = vd?.vditor?.wysiwyg?.element
  if (!root) return
  // 先清除旧标记（保留用户手动加的删除线样式类）
  root.querySelectorAll('.xm-mark-role, .xm-mark-term').forEach(el => {
    const span = el as HTMLElement
    const parent = span.parentNode
    if (parent) {
      parent.replaceChild(document.createTextNode(span.textContent || ''), span)
      parent.normalize()
    }
  })
  const roleNames = props.roleMarks || []
  const termNames = props.termMarks || []
  if (!roleNames.length && !termNames.length) return
  const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT)
  const textNodes: Text[] = []
  while (walker.nextNode()) textNodes.push(walker.currentNode as Text)
  for (const node of textNodes) {
    const text = node.nodeValue || ''
    if (!text.trim()) continue
    // 跳过代码块 / 行内代码，避免污染代码内容
    if (node.parentElement?.closest('pre,code')) continue
    const matches: { start: number; end: number; type: 'role' | 'term'; name: string }[] = []
    for (const name of roleNames) {
      if (!name) continue
      const re = new RegExp(escapeReg(name), 'g')
      let m: RegExpExecArray | null
      while ((m = re.exec(text))) matches.push({ start: m.index, end: m.index + name.length, type: 'role', name })
    }
    for (const name of termNames) {
      if (!name) continue
      const re = new RegExp(escapeReg(name), 'g')
      let m: RegExpExecArray | null
      while ((m = re.exec(text))) matches.push({ start: m.index, end: m.index + name.length, type: 'term', name })
    }
    if (!matches.length) continue
    matches.sort((a, b) => a.start - b.start || b.end - a.end)
    const frag = document.createDocumentFragment()
    let pos = 0
    for (const mt of matches) {
      if (mt.start < pos) continue
      if (mt.start > pos) frag.appendChild(document.createTextNode(text.slice(pos, mt.start)))
      const span = document.createElement('span')
      span.className = `xm-mark xm-mark-${mt.type}`
      span.dataset.markType = mt.type
      span.dataset.markName = mt.name
      span.textContent = text.slice(mt.start, mt.end)
      span.addEventListener('click', (e) => {
        e.preventDefault()
        e.stopPropagation()
        emit('mark-click', { type: mt.type, name: mt.name })
      })
      frag.appendChild(span)
      pos = mt.end
    }
    if (pos < text.length) frag.appendChild(document.createTextNode(text.slice(pos)))
    node.parentNode?.replaceChild(frag, node)
  }
}

/* ===== 段落下划线 =====
   给每个段落 p 加/去 xm-paragraph-line 类，由 CSS 渲染底部横线 */
const applyParagraphLine = () => {
  const vd = editorRef.value
  const root = vd?.vditor?.wysiwyg?.element
  if (!root) return
  const paras = Array.from(root.querySelectorAll('p')) as HTMLElement[]
  for (const p of paras) {
    if (props.paragraphLine) p.classList.add('xm-paragraph-line')
    else p.classList.remove('xm-paragraph-line')
  }
}

const refreshMarks = () => {
  if (!editorRef.value || !initialized) return
  applyMarks()
  applyParagraphLine()
}

watch(
  () => [props.roleMarks, props.termMarks, props.paragraphLine, props.textIndent, props.paragraphSpacing],
  () => { refreshMarks() },
  { deep: true }
)

onMounted(() => {
  const instance = new Vditor(editorId, {
    mode: 'wysiwyg',
    placeholder: props.placeholder,
    cache: { enable: false },
    theme: isDark() ? 'dark' : 'classic',
    toolbar,
    // vditor 运行时会直接用该对象作整份语言包（window.VditorI18n = options.i18n），
    // 局部覆盖即可；ITips 声明比实际 zh_CN 语言包更严格（缺 heading1-6 / tip 等键），故收窄为 ITips
    i18n: ZH_I18N as ITips,
    width: '100%',
    minHeight: 240,
    toolbarConfig: { pin: true },
    counter: {
      enable: false,
    },
    preview: {
      markdown: {
        autoSpace: true,
        paragraphBeginningSpace: false,
        fixTermTypo: false,
        toc: false,
        mark: true,
        gfmAutoLink: true,
      },
    },
    after: () => {
      initialized = true
      if (props.disabled) {
        try { instance.disabled() } catch { /* ignore */ }
      }
      nextTick(() => {
        if (editorRef.value && props.modelValue) {
          syncContent(props.modelValue)
        }
        // 替换 Vditor 默认抽象图标为重新设计的直观图标
        applyCustomIcons()
        // 工具栏悬停提示：Vditor 用 aria-label + CSS 伪元素渲染，会被
        // .vditor-toolbar 的 overflow-y: hidden 裁剪；同步到原生 title，
        // 由浏览器渲染在顶层，hover 即可显示中文提示。
        document.querySelectorAll(`#${editorId} .vditor-toolbar .vditor-tooltipped`).forEach(el => {
          const label = el.getAttribute('aria-label')
          if (label) el.setAttribute('title', label)
        })
        refreshMarks()
      })
    },
    input: (value: string) => {
      if (settingFromExternal) {
        // 只吞掉「与外部同步值完全一致」的那一次回显事件。
        // 若值与外部值不同（用户已在回显前输入了新内容），必须放行，
        // 否则新建章节后输入的文字会被吞掉、内容无法保存。
        settingFromExternal = false
        if (lastExternalValue !== null && value === lastExternalValue) {
          lastExternalValue = null
          return
        }
        lastExternalValue = null
      }
      // v-model 同步保持每次（中英文输入都能保存）
      emit('update:modelValue', value)
      // 只有本次新增内容含汉字才重建标记；拼音字母期间不刷新，避免打断输入法
      if (hasNewHan(value)) {
        if (markTimer) clearTimeout(markTimer)
        markTimer = setTimeout(() => refreshMarks(), 200)
      }
    },
    blur: () => {
      emit('blur')
    },
  })

  editorRef.value = instance
})

watch(
  () => props.modelValue,
  (value) => {
    if (!editorRef.value || !initialized) return
    syncContent(value || '')
  }
)

watch(
  () => props.disabled,
  (val) => {
    if (!editorRef.value) return
    if (val) {
      editorRef.value.disabled()
    } else {
      editorRef.value.enable()
    }
  }
)

// 暴露给父组件的方法
const insertPlainText = (text: string) => {
  if (props.disabled || !editorRef.value) return
  editorRef.value.focus()
  editorRef.value.insertValue(text)
}

const wrapSelectionWithHtml = (html: string) => {
  if (props.disabled || !editorRef.value) return
  editorRef.value.focus()
  const md = toMd(html)
  editorRef.value.insertValue(md)
}

const focusEditor = () => {
  editorRef.value?.focus()
}

const getSelectedText = (): string => {
  return editorRef.value?.getSelection() || ''
}

// ===== 补充暴露的常用方法（父组件可直接调用，无需经 vditor() 拿原始实例） =====
const getValue = (): string => {
  return editorRef.value?.getValue() || ''
}

const getHtml = (): string => {
  return editorRef.value?.getHTML() || ''
}

const setValue = (value: string) => {
  if (!editorRef.value || !initialized) return
  syncContent(value || '')
}

const insertValue = (text: string) => {
  if (props.disabled || !editorRef.value) return
  editorRef.value.focus()
  editorRef.value.insertValue(text)
}

const clear = () => {
  if (!editorRef.value) return
  editorRef.value.setValue('')
}

const getCursorPosition = () => {
  return editorRef.value?.getCursorPosition()
}

const setDisabled = (val: boolean) => {
  if (!editorRef.value) return
  if (val) editorRef.value.disabled()
  else editorRef.value.enable()
}

defineExpose({
  wrapSelectionWithHtml,
  focusEditor,
  getSelectedText,
  insertContent: wrapSelectionWithHtml,
  insertPlainText,
  // 补充：内容读写 / 光标 / 禁用
  getValue,
  getHtml,
  setValue,
  insertValue,
  clear,
  getSelection: getSelectedText,
  focus: focusEditor,
  getCursorPosition,
  disabled: setDisabled,
  enable: () => setDisabled(false),
  // 场景大纲联动：暴露 Vditor 实例，供父组件定位段落做滚动跳转
  vditor: () => editorRef.value,
  // 正文标记：外部数据变化（角色/词条列表）时手动刷新
  refreshMarks,
})

onUnmounted(() => {
  if (editorRef.value) {
    try { editorRef.value.destroy() } catch { /* ignore */ }
    editorRef.value = null
  }
  // 清理定时器和注入到 head 的实例样式
  if (markTimer) clearTimeout(markTimer)
  styleEl.remove()
})
</script>

<style>
/* 容器样式 — 匹配写作页 sheets 风格 */
.split-rich-text-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  border: 1px solid rgba(0, 201, 167, 0.12);
  border-radius: 16px;
  background: transparent;
  overflow: hidden;
  box-shadow:
    0 1px 3px rgba(0, 201, 167, 0.03),
    0 8px 24px rgba(0, 201, 167, 0.05);
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}
.split-rich-text-editor:focus-within {
  border-color: rgba(0, 201, 167, 0.4);
  box-shadow:
    0 1px 3px rgba(0, 201, 167, 0.05),
    0 12px 32px rgba(0, 201, 167, 0.1);
}

/* Vditor 主容器：交给 flex 布局撑满，不再写死高度 */
.split-rich-text-editor .vditor {
  flex: 1 !important;
  min-height: 0 !important;
  height: auto !important;
  border: none !important;
  border-radius: 0 !important;
  background: transparent !important;
  /* 亮色·薄荷气泡：半透明磨砂白底 + 薄荷绿图标 */
  --toolbar-height: 44px;
  --toolbar-background-color: rgba(255, 255, 255, .72);
  --toolbar-icon-color: #5b6b72;
  --toolbar-icon-hover-color: #00a896;
  --border-color: rgba(0, 201, 167, .28);
  --second-color: rgba(0, 201, 167, .5);
  --toolbar-divider-margin-top: 13px;
}
.split-rich-text-editor .vditor--dark {
  background: transparent !important;
  /* 暗色·玻璃全息：玻璃拟态底 + 全息三色图标 */
  --toolbar-height: 46px;
  --toolbar-background-color: rgba(255, 255, 255, .06);
  --toolbar-icon-color: #94a3b8;
  --toolbar-icon-hover-color: #e2e8f0;
  --border-color: rgba(255, 255, 255, .14);
  --second-color: rgba(129, 140, 248, .6);
  --toolbar-divider-margin-top: 14px;
}

/* ============ 亮色 · 薄荷气泡工具栏 ============ */
.split-rich-text-editor .vditor-toolbar {
  flex-shrink: 0;
  overflow: visible;
  scrollbar-width: none;
  padding: 0 14px;
  border-bottom: 1px solid var(--border-color) !important;
  background: var(--toolbar-background-color) !important;
  backdrop-filter: blur(16px) saturate(1.4);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, .5) inset,
    0 8px 28px rgba(0, 168, 150, .14);
}

/* 极窄屏降级：恢复横向滚动（六边形描边会被裁剪，可接受） */
@media (max-width: 600px) {
  .split-rich-text-editor .vditor-toolbar {
    overflow-x: auto;
    overflow-y: hidden;
  }
}

/* 悬停提示：隐藏 Vditor 伪元素 tooltip（after 回调已同步 aria-label → title）。
   注意：Vditor 的 tooltip 伪元素是 .vditor-tooltipped 的 ::after，
   而新皮肤涟漪/描边层是 ::before，两者不冲突。 */
.split-rich-text-editor .vditor-toolbar .vditor-tooltipped::after {
  display: none !important;
}
.split-rich-text-editor .vditor-toolbar--hide {
  display: none !important;
}

/* 薄荷气泡按钮：正圆 + 涟漪 hover */
.split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped {
  position: relative;
  width: 30px !important;
  height: 30px !important;
  margin: 7px 2px;
  padding: 0 !important;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--toolbar-icon-color);
  display: inline-flex !important;
  align-items: center;
  justify-content: center;
  transition: transform .15s, color .15s, box-shadow .2s;
}
/* 涟漪层 */
.split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(0, 201, 167, .12), transparent 70%);
  opacity: 0;
  transform: scale(.4);
  transition: opacity .25s, transform .25s;
  pointer-events: none;
}
.split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped:hover {
  color: var(--toolbar-icon-hover-color);
}
.split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped:hover::before {
  opacity: 1;
  transform: scale(1);
}
.split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped:active {
  transform: translateY(1px) scale(.9);
}
/* SVG 图标：线框风格（覆盖 Vditor 默认 fill:currentColor; stroke-width:0） */
.split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped svg {
  fill: none !important;
  stroke: currentColor;
  stroke-width: 1.8;
  stroke-linecap: round;
  stroke-linejoin: round;
  width: 15px;
  height: 15px;
}

/* 薄荷气泡分隔符：渐变竖线 */
.split-rich-text-editor .vditor-toolbar__divider {
  border-left: none !important;
  width: 1px;
  height: 18px;
  margin: 13px 5px !important;
  background: linear-gradient(180deg, transparent, rgba(0, 201, 167, .5), transparent);
}

/* 内容区：由 Vditor 原生 flex:1 自动填满剩余空间 */
.split-rich-text-editor .vditor-content {
  flex: 1 !important;
  min-height: 0 !important;
}

/* 正文区：纸张式留白 */
.split-rich-text-editor .vditor-wysiwyg {
  /* 默认留白：作为通用默认值；页面可通过更高特异性规则覆盖，无需 !important */
  padding: 36px 44px 56px;
  background: transparent !important;
  color: #1f2937 !important;
  line-height: 2 !important;
}
.split-rich-text-editor .vditor-reset {
  color: #1f2937 !important;
}

/* 暗色主题：容器与内容主区 */
:root[data-theme='dark'] .split-rich-text-editor .vditor,
:root[data-theme='dark'] .split-rich-text-editor .vditor-split,
:root[data-theme='dark'] .split-rich-text-editor .vditor-content,
:root[data-theme='dark'] .split-rich-text-editor .vditor-ir,
:root[data-theme='dark'] .split-rich-text-editor .vditor-sv,
:root[data-theme='dark'] .split-rich-text-editor .vditor-preview,
:root[data-theme='dark'] .split-rich-text-editor .vditor-resize {
  background-color: #1e1e1e !important;
  border-color: #363636 !important;
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-wysiwyg {
  color: #e6edf3 !important;
  background-color: #1e1e1e !important;
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset {
  color: #e6edf3 !important;
  background-color: #1e1e1e !important;
}

/* ============ 暗色 · 玻璃全息工具栏 ============ */
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar {
  background-color: rgba(255, 255, 255, .06) !important;
  border-bottom-color: rgba(255, 255, 255, .14) !important;
  backdrop-filter: blur(20px) saturate(1.6);
  padding: 0 16px;
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, .04) inset,
    0 0 40px rgba(34, 211, 238, .06),
    0 12px 36px rgba(0, 0, 0, .4);
}

/* 玻璃全息按钮：六边形 clip-path + 彩虹描边 hover + 发光脉冲 active */
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped {
  width: 32px !important;
  height: 32px !important;
  margin: 7px 2px;
  border-radius: 0;
  background: rgba(255, 255, 255, .05);
  clip-path: polygon(25% 4%, 75% 4%, 98% 50%, 75% 96%, 25% 96%, 2% 50%);
  color: #94a3b8;
}
/* 彩虹描边层（六边形外扩 1.5px） */
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped::before {
  inset: -1.5px;
  border-radius: 0;
  background: conic-gradient(from 0deg, #22d3ee, #818cf8, #e879f9, #22d3ee);
  clip-path: polygon(25% 4%, 75% 4%, 98% 50%, 75% 96%, 25% 96%, 2% 50%);
  z-index: -1;
  opacity: 0;
  transform: none;
  transition: opacity .25s;
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped:hover {
  color: #e2e8f0;
  background: rgba(255, 255, 255, .09);
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar__item .vditor-tooltipped:hover::before {
  opacity: 1;
  animation: holoSpin 2.4s linear infinite;
}
/* 彩虹色相旋转 */
@keyframes holoSpin {
  to { filter: hue-rotate(360deg); }
}
/* 发光脉冲（active 状态） */
@keyframes holoPulse {
  0%, 100% { box-shadow: 0 0 8px rgba(34, 211, 238, .5); }
  50%      { box-shadow: 0 0 22px rgba(232, 121, 249, .65); }
}

/* 全息分隔符：渐变竖线 + 紫红光点 */
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar__divider {
  position: relative;
  width: 1px;
  height: 20px;
  margin: 13px 6px !important;
  background: linear-gradient(180deg, transparent, rgba(129, 140, 248, .6), transparent);
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-toolbar__divider::after {
  content: '';
  position: absolute;
  left: -1px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 3px;
  border-radius: 50%;
  background: #e879f9;
  box-shadow: 0 0 8px #e879f9;
}

/* 暗色主题：分栏拖拽手柄 */
:root[data-theme='dark'] .split-rich-text-editor .vditor-split-resize,
:root[data-theme='dark'] .split-rich-text-editor .vditor-split-resize--left,
:root[data-theme='dark'] .split-rich-text-editor .vditor-split-resize--right {
  background-color: #363636 !important;
  border-color: #363636 !important;
}

/* 暗色主题：代码块 / 引用块 / 表格单元格（卡片式次级背景） */
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset pre,
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset blockquote,
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset table,
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset th,
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset td {
  background-color: #161b22 !important;
  border-color: #363636 !important;
}

/* 暗色主题：细节线条与行内 code 背景（细小缝隙不能是亮色） */
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset hr {
  border-top-color: #363636 !important;
  background-color: #363636 !important;
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset blockquote {
  border-left-color: #5a6a80 !important;
}
:root[data-theme='dark'] .split-rich-text-editor .vditor-reset code {
  background-color: rgba(110, 118, 129, 0.4) !important;
  border-color: #363636 !important;
}

/* 禁用状态 */
.split-rich-text-editor.is-disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
