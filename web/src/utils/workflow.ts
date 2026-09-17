import api from '@/api'

export interface WorkflowItem {
  id: number
  name: string
  description: string
  createdAt: string
  updatedAt: string
}

export interface WorkflowSceneAIConfig {
  modelId: number | null
  model: string
  temperature: number
  maxTokens: number
  promptCategory: string
  promptId: number | null
  promptText: string
  extraInput: string
  stepMode?: boolean
  promptEditorCollapsed?: boolean
}

export interface WorkflowScenePromptConfig {
  promptCategory: string
  promptId: number | null
  fieldValues?: Record<string, string>
}

export interface WorkflowSceneFieldConfig {
  name: string
  label: string
  type: 'text' | 'textarea' | 'select'
  options: string[]
  description: string
  required: boolean
}

export interface WorkflowSceneStartConfig {
  fields: WorkflowSceneFieldConfig[]
  fieldValues?: Record<string, string>
}

export interface WorkflowSceneTextInputImportedFile {
  name: string
  type: 'txt' | 'md' | 'json'
  content: string
}

export interface WorkflowSceneTextInputField {
  id: string
  name: string
  type: 'text' | 'variable'
  content: string
  placeholder: string
  importedFile?: WorkflowSceneTextInputImportedFile
}

export interface WorkflowSceneTextInputConfig {
  fields?: WorkflowSceneTextInputField[]
  text?: string
  placeholder?: string
  importedFile?: WorkflowSceneTextInputImportedFile
  outputField?: string
}

export interface WorkflowSceneNode {
  id: string
  type: string
  variant?: string
  title: string
  subtitle: string
  description: string
  outputs: string[]
  configPreview: string[]
  configDetails: Array<{
    label: string
    value: string
  }>
  position: {
    x: number
    y: number
  }
  width: number
  height: number
  startConfig?: WorkflowSceneStartConfig
  aiConfig?: WorkflowSceneAIConfig
  promptConfig?: WorkflowScenePromptConfig
  textInputConfig?: WorkflowSceneTextInputConfig
}

export interface WorkflowSceneEdge {
  id: string
  source: string
  target: string
  sourceHandle?: 'loop-body' | 'loop-done'
}

export interface WorkflowScene {
  nodes: WorkflowSceneNode[]
  edges: WorkflowSceneEdge[]
  viewport: {
    zoom: number
    camera: {
      x: number
      y: number
    }
  }
  updatedAt: string
}

export interface WorkflowNodeLibraryTemplate {
  id: string
  type: string
  variant?: string
  category: string
  label: string
  description: string
  outputs: string
  generated?: boolean
  template: {
    title: string
    subtitle: string
    description: string
    outputs: string[]
    configPreview: string[]
    configDetails: Array<{
      label: string
      value: string
    }>
    width: number
    height: number
    startConfig?: WorkflowSceneStartConfig
    aiConfig?: WorkflowSceneAIConfig
    promptConfig?: WorkflowScenePromptConfig
    textInputConfig?: WorkflowSceneTextInputConfig
  }
}

interface WorkflowDBItem {
  id: number
  name: string
  description: string
  scene: WorkflowScene | null
  createdAt: string
  updatedAt: string
}

interface ApiResponse<T> {
  success: boolean
  data: T
  message?: string
}

export const loadWorkflows = async (): Promise<WorkflowItem[]> => {
  try {
    const response = await api.get<any, ApiResponse<WorkflowDBItem[]>>('/workflows')
    if (response.success && response.data) {
      return response.data.map(item => ({
        id: item.id,
        name: item.name,
        description: item.description || '',
        createdAt: item.createdAt,
        updatedAt: item.updatedAt
      }))
    }
    return []
  } catch {
    return []
  }
}

export const getWorkflowById = async (id: number): Promise<WorkflowItem | null> => {
  try {
    const response = await api.get<any, ApiResponse<WorkflowDBItem>>(`/workflows/${id}`)
    if (response.success && response.data) {
      return {
        id: response.data.id,
        name: response.data.name,
        description: response.data.description || '',
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt
      }
    }
    return null
  } catch {
    return null
  }
}

export const createWorkflow = async (name: string, description: string = ''): Promise<WorkflowItem | null> => {
  try {
    const response = await api.post<any, ApiResponse<WorkflowDBItem>>('/workflows', {
      name,
      description,
      scene: { nodes: [], edges: [], viewport: { zoom: 1, camera: { x: 0, y: 0 } }, updatedAt: new Date().toISOString() }
    })
    if (response.success && response.data) {
      return {
        id: response.data.id,
        name: response.data.name,
        description: response.data.description || '',
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt
      }
    }
    return null
  } catch {
    return null
  }
}

export const updateWorkflowItem = async (id: number, patch: Partial<WorkflowItem>): Promise<WorkflowItem | null> => {
  try {
    const response = await api.put<any, ApiResponse<WorkflowDBItem>>(`/workflows/${id}`, patch)
    if (response.success && response.data) {
      return {
        id: response.data.id,
        name: response.data.name,
        description: response.data.description || '',
        createdAt: response.data.createdAt,
        updatedAt: response.data.updatedAt
      }
    }
    return null
  } catch {
    return null
  }
}

export const deleteWorkflow = async (id: number): Promise<boolean> => {
  try {
    const response = await api.delete<any, ApiResponse<void>>(`/workflows/${id}`)
    return response.success
  } catch {
    return false
  }
}

export const getWorkflowSceneById = async (id: number): Promise<WorkflowScene | null> => {
  try {
    const response = await api.get<any, ApiResponse<WorkflowDBItem>>(`/workflows/${id}`)
    if (response.success && response.data?.scene) {
      return response.data.scene
    }
    return null
  } catch {
    return null
  }
}

export const saveWorkflowScene = async (
  id: number,
  scene: Omit<WorkflowScene, 'updatedAt'> & {
    updatedAt?: string
  }
): Promise<WorkflowScene | null> => {
  try {
    const updatedAt = scene.updatedAt || new Date().toISOString()
    const response = await api.put<any, ApiResponse<WorkflowDBItem>>(`/workflows/${id}`, {
      scene: { ...scene, updatedAt }
    })
    if (response.success && response.data?.scene) {
      return response.data.scene
    }
    return null
  } catch {
    return null
  }
}

export const loadWorkflowNodeLibrary = async (): Promise<WorkflowNodeLibraryTemplate[]> => {
  try {
    const response = await api.get<any, ApiResponse<any[]>>('/workflows/node-library/all')
    if (response.success && response.data) {
      return response.data.map(item => ({
        id: item.id,
        type: item.type,
        variant: item.variant,
        category: item.category,
        label: item.label,
        description: item.description,
        outputs: item.outputs,
        generated: item.generated,
        template: item.template
      }))
    }
    return []
  } catch {
    return []
  }
}

export const saveWorkflowNodeLibrary = async (items: WorkflowNodeLibraryTemplate[]): Promise<boolean> => {
  try {
    for (const item of items) {
      await api.post<any, ApiResponse<void>>('/workflows/node-library', {
        id: item.id,
        type: item.type,
        variant: item.variant,
        category: item.category,
        label: item.label,
        description: item.description,
        outputs: item.outputs,
        generated: item.generated,
        template: item.template
      })
    }
    return true
  } catch {
    return false
  }
}

export const saveWorkflowNodeLibraryItem = async (item: WorkflowNodeLibraryTemplate): Promise<boolean> => {
  try {
    const response = await api.post<any, ApiResponse<void>>('/workflows/node-library', {
      id: item.id,
      type: item.type,
      variant: item.variant,
      category: item.category,
      label: item.label,
      description: item.description,
      outputs: item.outputs,
      generated: item.generated,
      template: item.template
    })
    return response.success
  } catch {
    return false
  }
}

export const deleteWorkflowNodeLibraryItem = async (id: string): Promise<boolean> => {
  try {
    const response = await api.delete<any, ApiResponse<void>>(`/workflows/node-library/${id}`)
    return response.success
  } catch {
    return false
  }
}
