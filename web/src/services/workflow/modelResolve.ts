import { configAPI } from '@/api'
import type { ApiModel } from '@/types'
import type { ModelSnapshot, NodeConfig } from '@/types/workflow'

/**
 * 模型解析（本地化核心）。
 *
 * 源项目用 `providerId::modelName` + electronAPI.store.getProviders() 取 apiKey/baseURL；
 * 星芒的 AI 通道是 POST /api/ai/chat，按 **configId（api_models.id）** 选模型，key/baseURL
 * 都由后端从 provider 表读取。因此：
 * - 节点里同时保存 `configId`（星芒原生）、`modelId`（源格式，兼容旧工程）、`modelConfig` 快照（展示用）
 * - 执行时优先 configId；缺失则用 `providerId::modelName` 反查 api_models
 */

export const splitModelId = (modelId: string): [string, string] => {
  const [providerId, modelName] = String(modelId || '').split('::')
  return [providerId || '', modelName || '']
}

let modelCache: { at: number; list: ApiModel[] } | null = null
const CACHE_MS = 15_000

/** 拉取全部模型配置（带 15s 缓存，避免循环体内反复请求） */
export async function loadModels(force = false): Promise<ApiModel[]> {
  if (!force && modelCache && Date.now() - modelCache.at < CACHE_MS) return modelCache.list
  const res: any = await configAPI.getAll()
  const list: ApiModel[] = res?.data || []
  modelCache = { at: Date.now(), list }
  return list
}

export function clearModelCache() {
  modelCache = null
}

/** 把选中的 api_model 转成节点里保存的快照 */
export function buildModelSnapshot(model: ApiModel): ModelSnapshot {
  return {
    configId: model.id,
    providerId: model.provider_id,
    providerName: model.provider_name || '',
    baseURL: (model as any).api_url || '',
    apiFormat: (model as any).api_format || '',
    modelName: model.model,
    apiKey: (model as any).api_key || ''
  }
}

export const modelDisplayName = (m: ApiModel) =>
  `${m.provider_name ? m.provider_name + ' / ' : ''}${m.name || m.model}`

/**
 * 解析节点配置对应的 configId。
 * 顺序：configId → modelId(providerId::modelName) → modelConfig 快照 → 报错
 */
export async function resolveConfigId(config: NodeConfig): Promise<number> {
  const direct = Number(config?.configId || 0)
  if (direct > 0) return direct

  const list = await loadModels()
  const [pid, mname] = splitModelId(config?.modelId || '')
  if (pid && mname) {
    const hit = list.find(m => String(m.provider_id) === String(pid) && m.model === mname)
    if (hit) return hit.id
  }
  const snapName = config?.modelConfig?.modelName
  if (snapName) {
    const snapProvider = config?.modelConfig?.providerName
    const hit = list.find(m => m.model === snapName && (!snapProvider || m.provider_name === snapProvider))
    if (hit) return hit.id
  }

  throw new Error('未设置 AI 模型！请在节点属性中选择模型。')
}
