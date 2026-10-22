/**
 * 图谱的视觉编码：类型 -> 标签/颜色，度数 -> 半径。
 * 色值沿用原 KnowledgeGraph.vue 的 typeColorMap，保持视觉连续。
 */
import type { GraphEntityType } from '@/types'

export interface GraphTypeMeta {
  value: GraphEntityType
  label: string
  color: string
}

/**
 * 人工连线添加的关系以此标记在 evidence 字段，
 * 图上用绿色区分 AI 抽取的边（GraphCanvas 与建关系表单共用）。
 */
export const MANUAL_EVIDENCE = '人工添加'

/**
 * 顺序即图例顺序
 *
 * 配色取自参照站（maliangwriter）公开库图谱的实际取值 —— Flat UI 色板：
 * 饱和度比 ECharts 默认色板高，节点更"实"、类型之间一眼分得开。
 */
export const GRAPH_TYPE_META: GraphTypeMeta[] = [
  { value: 'character', label: '角色', color: '#4A90D9' },
  { value: 'location', label: '地点', color: '#50C878' },
  { value: 'item', label: '物件', color: '#E8913A' },
  { value: 'faction', label: '势力', color: '#E67E22' },
  { value: 'event', label: '事件', color: '#E74C3C' },
  { value: 'skill', label: '功法', color: '#2ECC71' },
  { value: 'clue', label: '线索', color: '#9B59B6' }
]

/** 后端允许出现 7 类之外的 type（历史数据/导入数据），统一落到「其他」 */
export const FALLBACK_TYPE_LABEL = '其他'
export const FALLBACK_TYPE_COLOR = '#7F8C8D'

const META_BY_TYPE = new Map<string, GraphTypeMeta>(GRAPH_TYPE_META.map(m => [m.value, m]))

export function typeLabel(type: string): string {
  return META_BY_TYPE.get(type)?.label || FALLBACK_TYPE_LABEL
}

export function typeColor(type: string): string {
  return META_BY_TYPE.get(type)?.color || FALLBACK_TYPE_COLOR
}

/** 节点半径按度数：连接越多越大，6~20 之间 */
export function nodeRadius(degree: number): number {
  return Math.min(20, Math.max(6, 6 + Math.sqrt(Math.max(0, degree)) * 2))
}
