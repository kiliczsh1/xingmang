import type { NodeType, WNodeData } from '@/types/workflow'

/** 节点分组（节点库侧栏的折叠单元） */
export type NodeGroup = '输入' | '处理' | '流程控制' | '媒体' | '输出' | '辅助'

export const NODE_GROUP_ORDER: NodeGroup[] = ['输入', '处理', '流程控制', '媒体', '输出', '辅助']

export interface NodeCatalogItem {
  type: NodeType
  label: string
  /** font-awesome 类名（不含 fas） */
  icon: string
  /** 强调色，节点库图标与节点卡片边框共用 */
  color: string
  group: NodeGroup
  description: string
  /** 来源：star = 星域移植节点（与星芒原生节点并列） */
  origin?: 'star'
  /** 是否有输入口 */
  hasTarget: boolean
  /** 输出口 id 列表；空数组表示无输出（end）；loop 有两个出口 */
  sources: string[]
}

export const NODE_CATALOG: NodeCatalogItem[] = [
  {
    type: 'start',
    label: '开始',
    icon: 'fa-play-circle',
    color: '#67c23a',
    group: '输入',
    description: '工作流入口，可声明初始变量',
    hasTarget: false,
    sources: ['']
  },
  {
    type: 'text',
    label: '文本节点',
    icon: 'fa-file-text',
    color: '#909399',
    group: '输入',
    description: '直接文本或引用上游变量',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'aiProcess',
    label: 'AI 处理',
    icon: 'fa-robot',
    color: '#409eff',
    group: '处理',
    description: '调用模型生成，支持 {{变量}} 模板',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'merge',
    label: '合并节点',
    icon: 'fa-code-merge',
    color: '#e6a23c',
    group: '处理',
    description: '拼接 / 数组 / 对象三种合并方式',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'loop',
    label: '循环节点',
    icon: 'fa-rotate',
    color: '#9c27b0',
    group: '流程控制',
    description: '按数组或固定次数循环，两侧出口',
    hasTarget: true,
    sources: ['loop-body', 'loop-done']
  },
  {
    type: 'context',
    label: '上下文节点',
    icon: 'fa-database',
    color: '#00bcd4',
    group: '流程控制',
    description: '保存 / 追加 / 读取上下文变量',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'end',
    label: '结束',
    icon: 'fa-stop-circle',
    color: '#f56c6c',
    group: '输出',
    description: '汇总输出并落盘',
    hasTarget: true,
    sources: []
  },
  /* ---------- 星域移植节点（与星芒原生节点并列） ---------- */
  {
    type: 'textInput',
    label: '文本输入',
    icon: 'fa-i-cursor',
    color: '#7e57c2',
    group: '输入',
    description: '星域节点：手动文本 + 占位符名',
    origin: 'star',
    hasTarget: false,
    sources: ['']
  },
  {
    type: 'aiModel',
    label: 'AI 模型',
    icon: 'fa-microchip',
    color: '#3949ab',
    group: '处理',
    description: '星域节点：独立 AI 节点，支持 {{占位符}} 与流式预览',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'textProcessor',
    label: '文本处理',
    icon: 'fa-wand-magic',
    color: '#00897b',
    group: '处理',
    description: '星域节点：格式化 / 转大写 / 转小写',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'promptNode',
    label: '提示词',
    icon: 'fa-book-open',
    color: '#8e24aa',
    group: '处理',
    description: '星域节点：携带提示词库内容供下游引用',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'videoGeneration',
    label: '文生视频',
    icon: 'fa-film',
    color: '#d81b60',
    group: '媒体',
    description: '星域节点：需接入视频服务商',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'imageToVideo',
    label: '图生视频',
    icon: 'fa-clapperboard',
    color: '#c2185b',
    group: '媒体',
    description: '星域节点：需接入视频服务商',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'imageOutput',
    label: '图片输出',
    icon: 'fa-image',
    color: '#00838f',
    group: '输出',
    description: '星域节点：保存图片并内嵌预览',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'videoOutput',
    label: '视频输出',
    icon: 'fa-video',
    color: '#00695c',
    group: '输出',
    description: '星域节点：保存视频并内嵌预览',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'imageGeneration',
    label: '文生图',
    icon: 'fa-image',
    color: '#e91e63',
    group: '媒体',
    description: '按提示词生成图片（需配置生图模型）',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'imageAnalyze',
    label: '图片分析',
    icon: 'fa-eye',
    color: '#00acc1',
    group: '媒体',
    description: '多模态模型识图，流式输出描述',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'fileInput',
    label: '文件输入',
    icon: 'fa-file-import',
    color: '#5c6bc0',
    group: '输入',
    description: '读取本地文本文件（UTF-8/GBK 自动识别）',
    origin: 'star',
    hasTarget: false,
    sources: ['']
  },
  {
    type: 'textOutput',
    label: '文本输出',
    icon: 'fa-file-lines',
    color: '#26a69a',
    group: '输出',
    description: '按 md/txt/html 写出，节点内可预览',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'fileOutput',
    label: '文件输出',
    icon: 'fa-file-export',
    color: '#8d6e63',
    group: '输出',
    description: '把内容写入本次执行的输出目录',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'transfer',
    label: '中转节点',
    icon: 'fa-right-left',
    color: '#607d8b',
    group: '处理',
    description: '聚合/转发上游变量，便于多路复用',
    origin: 'star',
    hasTarget: true,
    sources: ['']
  },
  {
    type: 'comment',
    label: '注释',
    icon: 'fa-note-sticky',
    color: '#ffb300',
    group: '辅助',
    description: '画布标注，不参与执行',
    origin: 'star',
    hasTarget: false,
    sources: []
  }
]

export const catalogOf = (type: string): NodeCatalogItem | undefined =>
  NODE_CATALOG.find(i => i.type === type)

export const labelOf = (type: string): string => catalogOf(type)?.label || type

/** 各节点类型的新建默认配置（原 WorkflowIde.addNode 内联逻辑抽到这里，两处共用） */
export const createNodeConfig = (type: NodeType): Record<string, any> => {
  switch (type) {
    case 'start':
      return { initialVariables: {} }
    case 'aiProcess':
      return {
        modelId: '',
        userPrompt: '',
        contextVariables: [],
        contextMaxItems: 3,
        temperature: 0.7,
        maxTokens: 4000
      }
    case 'text':
      return { textSource: 'direct', text: '', outputVariable: 'textResult' }
    case 'merge':
      return { mergeType: 'concat' }
    case 'loop':
      return { loopType: 'fixed', fixedCount: 5, historyWindow: 3, keepContext: true }
    case 'context':
      return { action: 'save', variableName: '', inputVariables: [], maxHistory: 20 }
    case 'end':
      return {
        outputFormat: 'md',
        autoSave: false,
        includeOutputs: [],
        separator: '\n\n---\n\n'
      }
    /* ---------- 以下为星域移植节点：默认配置逐字段对齐星域 getDefaultNodeProperties ----------
       标注「星芒适配」的字段是星域没有、但星芒上下文变量模型必须的，不参与星域语义。 */
    case 'textInput':
      return { text: '', placeholder: '未设占位符', promptList: [], selectedPrompts: [] }
    case 'promptNode':
      return { text: '', placeholder: '未设占位符', promptList: [], selectedPrompts: [] }
    case 'fileInput':
      return {
        fileType: 'text',
        path: '',
        encoding: 'utf-8',
        readMode: 'full',
        outputVariable: 'fileContent' // 星芒适配
      }
    case 'aiModel':
      return {
        model: '',
        prompt: '',
        encryptedPrompt: '',
        isEncrypted: false,
        temperature: 0.7,
        maxTokens: null,
        outputVariable: 'aiOutput' // 星芒适配
      }
    case 'textProcessor':
      return {
        operation: 'format',
        parameters: {},
        outputFormat: 'text',
        preserveFormat: false,
        sourceVariable: 'lastOutput' // 星芒适配
      }
    case 'transfer':
      return {
        receivedData: '',
        placeholder: '未设占位符',
        autoForward: true,
        displayMode: 'text'
      }
    case 'textOutput':
      return { format: 'text', filename: 'output.txt', sourceVariable: 'lastOutput' } // sourceVariable 为星芒适配
    case 'fileOutput':
      return { path: './exports/', filename: 'output.txt', sourceVariable: 'lastOutput' } // sourceVariable 为星芒适配
    case 'imageGeneration':
      return {
        prompt: '',
        model: '',
        width: 1024,
        height: 1024,
        num_images: 1,
        style: '',
        apiProvider: 'custom'
      }
    case 'videoGeneration':
      return {
        prompt: '',
        model: '',
        size: '1920x1080',
        duration: 5,
        fps: 30,
        quality: 'quality',
        with_audio: true,
        style: '',
        apiProvider: 'custom'
      }
    case 'imageToVideo':
      return {
        prompt: '',
        model: '',
        width: 1024,
        height: 576,
        duration: 5,
        fps: 30,
        style: '',
        apiProvider: 'custom',
        sourceVariable: 'imageB64' // 星芒适配（星域从上游图片输出节点取图）
      }
    case 'imageOutput':
      return {
        filename: 'output.png',
        format: 'png',
        sourceVariable: 'images' // 星芒适配
      }
    case 'videoOutput':
      return {
        filename: 'output.mp4',
        format: 'mp4',
        sourceVariable: 'videoUrl' // 星芒适配
      }
    case 'comment':
      return {
        text: '在此输入注释内容...',
        fontSize: 14,
        textColor: '#666666',
        backgroundColor: '#fffacd'
      }
    case 'imageAnalyze':
      return {
        prompt: '',
        model: '',
        inputImage: null,
        inputImageName: '',
        apiProvider: 'custom',
        sourceVariable: 'imageB64' // 星芒适配（星域从上游图片输出节点取图）
      }
    default:
      return {}
  }
}

/** 新建节点的 data（运行时字段一并初始化，保持与既有模板一致的结构） */
export const createNodeData = (type: NodeType): WNodeData => ({
  label: labelOf(type),
  config: createNodeConfig(type),
  status: 'idle',
  result: null,
  error: null,
  executionHistory: []
})

/** 拖拽传输的 MIME 类型（节点库 → 画布） */
export const NODE_DND_MIME = 'application/x-wf-node'
