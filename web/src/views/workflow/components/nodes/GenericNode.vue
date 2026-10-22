<template>
  <div
    class="g-node"
    :class="[`g-${type}`, { selected, running: data.status === 'running', success: data.status === 'success', error: data.status === 'error' }]"
    :style="rootStyle"
  >
    <Handle v-if="meta?.hasTarget" type="target" :position="Position.Left" />
    <Handle v-if="meta && meta.sources.length > 0" type="source" :position="Position.Right" />

    <div v-if="type !== 'comment'" class="node-header">
      <div class="header-left">
        <i class="fas" :class="meta?.icon || 'fa-cube'" />
        <span>{{ data.label || meta?.label || type }}</span>
      </div>
      <div class="header-right">
        <el-tag v-if="data.status === 'running'" type="warning" size="small">
          <i class="fas fa-spinner fa-spin" /> 运行中
        </el-tag>
        <el-tag v-else-if="data.status === 'success'" type="success" size="small">
          <i class="fas fa-check" /> 成功
        </el-tag>
        <el-tag v-else-if="data.status === 'error'" type="danger" size="small">
          <i class="fas fa-times" /> 失败
        </el-tag>
      </div>
    </div>

    <!-- 注释：纯文本块 + 可拖拽调大小 -->
    <div v-if="type === 'comment'" class="comment-body">
      <i class="fas fa-note-sticky comment-icon" />
      <span class="comment-text" :style="{ color: cfg.textColor || undefined, fontSize: (cfg.fontSize || 14) + 'px' }">
        {{ cfg.text || data.label || '注释' }}
      </span>
      <span class="comment-resize" title="拖拽调整大小" @mousedown.stop.prevent="onResizeStart">
        <i class="fas fa-up-right-and-down-left-from-center" />
      </span>
    </div>

    <div v-else class="node-body">
      <!-- 星域·文本输入 / 提示词 -->
      <template v-if="type === 'textInput' || type === 'promptNode'">
        <div class="kv"><i class="fas fa-i-cursor" /><span>占位符 {{ placeholderLabel }}</span></div>
        <div class="kv" v-if="promptCount">
          <i class="fas fa-list-check" /><span>提示词列表 {{ promptCount }} 条 / 选中 {{ (cfg.selectedPrompts || []).length }}</span>
        </div>
        <div v-if="cfg.text" class="preview-block">
          <pre class="preview-text">{{ String(cfg.text).slice(0, 200) }}</pre>
        </div>
        <div v-else-if="!promptCount" class="warning-info"><i class="fas fa-exclamation-triangle" /><span>未填写文本</span></div>
      </template>

      <!-- 星域·AI 模型 -->
      <template v-else-if="type === 'aiModel'">
        <div v-if="!cfg.model && !cfg.configId" class="warning-info">
          <i class="fas fa-exclamation-triangle" /><span>未设置 AI 模型</span>
        </div>
        <div v-else class="kv"><i class="fas fa-microchip" /><span>{{ cfg.model || modelLabel }}</span></div>
        <div v-if="cfg.isEncrypted" class="kv"><i class="fas fa-lock" /><span>加密提示词</span></div>
        <div v-if="cfg.prompt" class="prompt-preview">
          <i class="fas fa-comment-dots" /><span>{{ String(cfg.prompt).slice(0, 44) }}</span>
        </div>
        <div v-else-if="!cfg.isEncrypted" class="warning-info">
          <i class="fas fa-exclamation-triangle" /><span>未设置提示词</span>
        </div>
        <div v-if="aiText" class="preview-block">
          <pre class="preview-text">{{ aiText.slice(-600) }}</pre>
          <span class="preview-more">{{ aiText.length }} 字{{ data.status === 'running' ? ' · 生成中…' : '' }}</span>
        </div>
      </template>

      <!-- 星域·文本处理 -->
      <template v-else-if="type === 'textProcessor'">
        <div class="kv"><i class="fas fa-wand-magic" /><span>{{ opLabel }}</span></div>
        <div class="kv"><i class="fas fa-arrow-down-long" /><span>取 {{ cfg.sourceVariable || 'lastOutput' }}</span></div>
        <div v-if="data.result" class="result-preview">
          <i class="fas fa-check-circle" /><span>{{ String(data.result).slice(0, 80) }}</span>
        </div>
      </template>

      <!-- 星域·中转 -->
      <template v-else-if="type === 'transfer'">
        <div class="kv"><i class="fas fa-right-left" /><span>占位符 {{ cfg.placeholder || '未设占位符' }}</span></div>
        <div class="kv"><i class="fas fa-code" /><span>{{ cfg.displayMode === 'json' ? 'JSON 模式' : '文本模式' }}</span></div>
        <div class="kv" v-if="cfg.autoForward !== false"><i class="fas fa-share" /><span>自动转发</span></div>
        <div v-if="cfg.receivedData" class="preview-block">
          <pre class="preview-text">{{ String(cfg.receivedData).slice(0, 200) }}</pre>
        </div>
      </template>

      <!-- 星域·文件输入 -->
      <template v-else-if="type === 'fileInput'">
        <div v-if="!cfg.path && !cfg.fileName" class="warning-info">
          <i class="fas fa-exclamation-triangle" /><span>未指定文件路径或导入文件</span>
        </div>
        <div v-else class="kv">
          <i class="fas fa-file" />
          <span class="path-text">{{ cfg.fileName || cfg.path }}</span>
        </div>
        <div class="kv"><i class="fas fa-right-long" /><span>{{ fileTypeLabel }} · {{ encodingLabel }} · {{ readModeLabel }}</span></div>
        <div class="kv"><i class="fas fa-tag" /><span>输出 {{ cfg.outputVariable || 'fileContent' }}</span></div>
        <div v-if="data.result" class="result-preview">
          <i class="fas fa-align-left" /><span>{{ String(data.result).slice(0, 90) }}…</span>
        </div>
      </template>

      <!-- 星域·文本输出 -->
      <template v-else-if="type === 'textOutput'">
        <div class="kv"><i class="fas fa-file-lines" /><span>{{ cfg.filename || 'output.txt' }} · {{ cfg.format || 'text' }}</span></div>
        <div class="kv"><i class="fas fa-arrow-down-long" /><span>取 {{ cfg.sourceVariable || 'lastOutput' }}</span></div>
        <div v-if="previewText" class="preview-block">
          <pre class="preview-text">{{ previewText }}</pre>
        </div>
        <div v-if="data.outputFile" class="node-actions">
          <button class="act-btn nodrag" title="把输出内容存进提示词库" @mousedown.stop @click.stop="saveToPromptLibrary">
            <i class="fas fa-bookmark" /> 存提示词库
          </button>
        </div>
      </template>

      <!-- 星域·文件输出 -->
      <template v-else-if="type === 'fileOutput'">
        <div class="kv">
          <i class="fas fa-file-export" />
          <span class="path-text">{{ (cfg.path || './exports/') + (cfg.filename || 'output.txt') }}</span>
        </div>
        <div class="kv"><i class="fas fa-arrow-down-long" /><span>取 {{ cfg.sourceVariable || 'lastOutput' }}</span></div>
        <div v-if="data.result" class="result-preview">
          <i class="fas fa-check-circle" /><span>{{ String(data.result).slice(0, 80) }}</span>
        </div>
      </template>

      <!-- 星域·文生图 -->
      <template v-else-if="type === 'imageGeneration'">
        <div v-if="!cfg.model && !cfg.configId" class="warning-info">
          <i class="fas fa-exclamation-triangle" /><span>未设置生图模型</span>
        </div>
        <div v-else class="kv"><i class="fas fa-palette" /><span>{{ sizeLabel }} · {{ cfg.num_images || 1 }} 张</span></div>
        <div v-if="cfg.prompt" class="prompt-preview">
          <i class="fas fa-comment-dots" />
          <span>{{ String(cfg.prompt).slice(0, 48) }}{{ String(cfg.prompt).length > 48 ? '…' : '' }}</span>
        </div>
        <div v-if="previewSrc" class="image-preview">
          <img :src="previewSrc" alt="生成结果" />
          <span class="image-count">{{ imageCount }} 张</span>
        </div>
      </template>

      <!-- 星域·图片分析 -->
      <template v-else-if="type === 'imageAnalyze'">
        <div v-if="!cfg.model && !cfg.configId" class="warning-info">
          <i class="fas fa-exclamation-triangle" /><span>未设置视觉模型</span>
        </div>
        <div class="kv">
          <i class="fas fa-image" />
          <span>{{ cfg.inputImageName ? '已导入 ' + cfg.inputImageName : '图片源 ' + (cfg.sourceVariable || 'imageB64') }}</span>
        </div>
        <div v-if="data.result" class="result-preview">
          <i class="fas fa-eye" /><span>{{ String(data.result).slice(0, 80) }}{{ String(data.result).length > 80 ? '…' : '' }}</span>
        </div>
      </template>

      <!-- 星域·文生视频 -->
      <template v-else-if="type === 'videoGeneration'">
        <div class="warning-info"><i class="fas fa-plug-circle-exclamation" /><span>视频服务商未接入</span></div>
        <div v-if="cfg.prompt" class="prompt-preview">
          <i class="fas fa-comment-dots" /><span>{{ String(cfg.prompt).slice(0, 44) }}</span>
        </div>
        <div class="kv">
          <i class="fas fa-clock" />
          <span>{{ cfg.size || '1920x1080' }} · {{ cfg.duration || 5 }} 秒 · {{ cfg.fps || 30 }} fps</span>
        </div>
      </template>

      <!-- 星域·图生视频 -->
      <template v-else-if="type === 'imageToVideo'">
        <div class="warning-info"><i class="fas fa-plug-circle-exclamation" /><span>视频服务商未接入</span></div>
        <div class="kv"><i class="fas fa-image" /><span>图片源 {{ cfg.sourceVariable || 'imageB64' }}</span></div>
        <div class="kv">
          <i class="fas fa-clock" />
          <span>{{ cfg.width || 1024 }}×{{ cfg.height || 576 }} · {{ cfg.duration || 5 }} 秒</span>
        </div>
      </template>

      <!-- 星域·图片输出 -->
      <template v-else-if="type === 'imageOutput'">
        <div class="kv">
          <i class="fas fa-image" />
          <span>{{ cfg.filename || 'output.png' }} · {{ String(cfg.format || 'png').toUpperCase() }}</span>
        </div>
        <div v-if="previewSrc" class="image-preview">
          <img :src="previewSrc" alt="输出图片" />
        </div>
        <div v-else class="kv">
          <i class="fas fa-hourglass-half" /><span>等待上游图片（取 {{ cfg.sourceVariable || 'images' }}）</span>
        </div>
      </template>

      <!-- 星域·视频输出 -->
      <template v-else-if="type === 'videoOutput'">
        <div class="kv">
          <i class="fas fa-video" />
          <span>{{ cfg.filename || 'output.mp4' }} · {{ String(cfg.format || 'mp4').toUpperCase() }}</span>
        </div>
        <div v-if="(data as any).videoUrl" class="video-preview">
          <video :src="(data as any).videoUrl" controls preload="metadata" />
        </div>
        <div v-else class="kv">
          <i class="fas fa-hourglass-half" /><span>等待上游视频（取 {{ cfg.sourceVariable || 'videoUrl' }}）</span>
        </div>
      </template>
    </div>

    <NodeMetaBar :ms="data.durationMs" :breakpoint="!!data.breakpoint" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Handle, Position } from '@vue-flow/core'
import NodeMetaBar from './NodeMetaBar.vue'
import { catalogOf } from '@/services/workflow/nodeCatalog'
import { promptAPI } from '@/api'
import type { WNodeData } from '@/types/workflow'

const props = defineProps<{ id?: string; type?: string; data: WNodeData; selected?: boolean }>()

const meta = computed(() => (props.type ? catalogOf(props.type) : undefined))
const cfg = computed<Record<string, any>>(() => (props.data?.config || {}) as Record<string, any>)

/** 注释块用配置里的底色/字色/尺寸（拖拽可调） */
const rootStyle = computed(() => {
  if (props.type !== 'comment') return undefined
  const style: Record<string, string> = {
    background: cfg.value.backgroundColor || '#fffacd',
    borderColor: cfg.value.borderColor || '#ffe082'
  }
  if (cfg.value.width) style.width = `${cfg.value.width}px`
  if (cfg.value.height) style.height = `${cfg.value.height}px`
  return style
})

/** 注释节点：拖右下角调大小（与星域 adjustCommentNodeSize 同口径） */
const resizing = ref(false)
const onResizeStart = (e: MouseEvent) => {
  resizing.value = true
  const startX = e.clientX
  const startY = e.clientY
  const w0 = Number(cfg.value.width || 220)
  const h0 = Number(cfg.value.height || 120)
  const onMove = (ev: MouseEvent) => {
    cfg.value.width = Math.max(140, Math.round(w0 + (ev.clientX - startX)))
    cfg.value.height = Math.max(60, Math.round(h0 + (ev.clientY - startY)))
  }
  const onUp = () => {
    resizing.value = false
    document.removeEventListener('mousemove', onMove)
    document.removeEventListener('mouseup', onUp)
  }
  document.addEventListener('mousemove', onMove)
  document.addEventListener('mouseup', onUp)
}

/** 文本输出的预览（执行期由执行器写进 nodeRuntime） */
const previewText = computed(() => {
  if (props.type !== 'textOutput') return ''
  const p = (props.data as any)?.preview
  return typeof p === 'string' ? p : ''
})

/** 把输出内容存进星芒提示词库，方便后续复用 */
const saveToPromptLibrary = async () => {
  const content = String((props.data as any)?.result || '')
  if (!content.trim()) {
    ElMessage.warning('还没有可保存的内容')
    return
  }
  const fileName = String(cfg.value.fileName || '工作流输出')
  try {
    await promptAPI.create({
      title: `${fileName}`.slice(0, 60),
      content,
      category: '工作流输出'
    } as any)
    ElMessage.success('已存入提示词库')
  } catch {
    /* 拦截器已提示 */
  }
}

const imageList = computed<any[]>(() => {
  const imgs = (props.data as any)?.images
  return Array.isArray(imgs) ? imgs : []
})
const imageCount = computed(() => imageList.value.length)
const previewSrc = computed(() => {
  const first = imageList.value[0]
  if (!first) return ''
  if (first.b64) return `data:image/png;base64,${first.b64}`
  return first.url || ''
})

/* ---------- 星域移植节点的展示辅助 ---------- */
/** AI 模型节点：内嵌输出（优先流式，其次结果） */
const aiText = computed(() => {
  const s = (props.data as any)?.streamText
  if (typeof s === 'string' && s) return s
  const r = (props.data as any)?.result
  return typeof r === 'string' ? r : ''
})
const modelLabel = computed(() => {
  const cfgv = cfg.value
  if (cfgv?.model) return String(cfgv.model)
  const name = cfgv?.modelConfig?.modelName || ''
  if (name) return name
  const mid = String(cfgv?.modelId || '')
  if (mid.includes('::')) return mid.split('::')[1]
  return mid || (cfgv?.configId ? `模型 #${cfgv.configId}` : '未设置')
})
/** 星域口径的操作名（UI 下拉里的 7 项，执行时只有格式化有内置逻辑） */
const opLabel = computed(() => {
  const map: Record<string, string> = {
    format: '格式化',
    clean: '清理文本',
    extract: '提取内容',
    replace: '替换文本',
    split: '分割文本',
    merge: '合并文本',
    transform: '转换格式',
    uppercase: '转大写',
    lowercase: '转小写'
  }
  return map[String(cfg.value.operation || 'format')] || '格式化'
})
const promptCount = computed(() => (Array.isArray(cfg.value.promptList) ? cfg.value.promptList.length : 0))
const placeholderLabel = computed(() => {
  const p = String(cfg.value.placeholder || '').trim()
  return !p || p === '未设占位符' ? 'textInput' : p
})
const fileTypeLabel = computed(() => {
  const map: Record<string, string> = { text: '文本', json: 'JSON', csv: 'CSV', xml: 'XML', markdown: 'Markdown' }
  return map[String(cfg.value.fileType || 'text')] || '文本'
})
const encodingLabel = computed(() => String(cfg.value.encoding || 'utf-8').toUpperCase())
const readModeLabel = computed(() => {
  const map: Record<string, string> = { full: '完整读取', lines: '按行读取', chunks: '分块读取' }
  return map[String(cfg.value.readMode || 'full')] || '完整读取'
})
const sizeLabel = computed(() => `${cfg.value.width || 1024}x${cfg.value.height || 1024}`)

</script>

<style scoped>
.g-node {
  position: relative;
  min-width: 200px;
  max-width: 320px;
  background: var(--wf-el-bg, #ffffff);
  border: 2px solid var(--wf-el-border, #dcdfe6);
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: all 0.3s;
}
.g-node:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}
.g-node.selected {
  box-shadow: 0 0 0 4px rgba(64, 158, 255, 0.2);
}
.g-node.running {
  border-color: var(--el-color-warning, #e6a23c);
  animation: wf-node-pulse 2s infinite;
}
.g-node.success {
  border-color: var(--el-color-success, #67c23a);
}
.g-node.error {
  border-color: var(--el-color-danger, #f56c6c);
}
/* 注释块：无表头、纯标注 */
.g-comment {
  min-width: 160px;
  border-style: dashed;
  box-shadow: none;
}
.comment-body {
  position: relative;
  height: 100%;
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 10px 12px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--wf-el-text-primary, #303133);
  white-space: pre-wrap;
  word-break: break-word;
}
.comment-icon {
  font-size: 11px;
  margin-top: 3px;
  opacity: 0.6;
}
.comment-text {
  flex: 1;
  min-width: 0;
}
/* 注释右下角拖拽调大小 */
.comment-resize {
  position: absolute;
  right: 2px;
  bottom: 2px;
  padding: 2px 3px;
  font-size: 9px;
  color: var(--wf-el-text-placeholder, #c0c4cc);
  cursor: nwse-resize;
  opacity: 0;
  transition: opacity 0.15s;
}
.g-comment:hover .comment-resize {
  opacity: 1;
}
/* 文本输出预览 */
.preview-block {
  position: relative;
  padding: 6px 8px;
  border-radius: 4px;
  background: var(--wf-el-fill-lighter, #fafafa);
  max-height: 160px;
  overflow: hidden;
}
.preview-text {
  margin: 0;
  font-size: 10px;
  line-height: 1.45;
  color: var(--wf-el-text-secondary, #909399);
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--font-mono, Consolas, Monaco, monospace);
}
.preview-more {
  display: block;
  margin-top: 4px;
  font-size: 10px;
  color: var(--el-color-warning, #e6a23c);
}
.path-text {
  word-break: break-all;
}
.node-actions {
  display: flex;
  gap: 6px;
}
.act-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 3px 8px;
  border: 1px solid var(--wf-el-border, #dcdfe6);
  border-radius: 4px;
  background: var(--wf-el-bg, #fff);
  color: var(--wf-el-text-regular, #606266);
  font-size: 11px;
  cursor: pointer;
}
.act-btn:hover {
  border-color: var(--el-color-primary, #409eff);
  color: var(--el-color-primary, #409eff);
}
.act-btn i {
  font-size: 10px;
}
.node-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  background: var(--wf-el-fill-lighter, #fafafa);
  border-bottom: 1px solid var(--wf-el-border, #dcdfe6);
}
.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}
.header-left i {
  font-size: 13px;
  flex-shrink: 0;
  color: var(--el-color-primary, #409eff);
}
.header-left span {
  font-weight: 600;
  font-size: 14px;
  color: var(--wf-el-text-primary, #303133);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.header-right {
  flex-shrink: 0;
}
.node-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
}
.kv,
.prompt-preview,
.result-preview {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  font-size: 11px;
  line-height: 1.4;
  color: var(--wf-el-text-secondary, #909399);
  word-break: break-word;
}
.kv i,
.prompt-preview i,
.result-preview i {
  margin-top: 2px;
  flex-shrink: 0;
  font-size: 10px;
  color: var(--el-color-primary, #409eff);
}
.result-preview {
  padding: 8px;
  border-radius: 4px;
  background: var(--wf-el-fill-lighter, #fafafa);
}
.result-preview i {
  color: var(--el-color-success, #67c23a);
}
.warning-info {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 4px;
  background: rgba(230, 162, 60, 0.1);
  color: var(--el-color-warning, #e6a23c);
  font-size: 11px;
}
.video-preview {
  border-radius: 6px;
  overflow: hidden;
  background: #000;
}
.video-preview video {
  display: block;
  width: 100%;
  max-height: 220px;
}
.image-preview {
  position: relative;
  border-radius: 6px;
  overflow: hidden;
  background: var(--wf-el-fill-lighter, #fafafa);
}
.image-preview img {
  display: block;
  width: 100%;
  max-height: 220px;
  object-fit: contain;
}
.image-count {
  position: absolute;
  right: 6px;
  bottom: 6px;
  padding: 1px 6px;
  border-radius: 8px;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  font-size: 10px;
}
</style>
