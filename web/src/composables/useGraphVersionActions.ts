/**
 * 图谱版本的增删改查编排：弹窗提示、确认、toast、导出下载。
 *
 * WikiGraph 页面与 Write 里的知识图谱弹窗共用同一套侧栏，这两处的交互必须一致，
 * 所以把编排集中在这里，组件只负责渲染。
 */
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { GraphEntity, GraphRelation } from '@/types'
import { useKnowledgeGraphStore } from '@/stores/knowledgeGraph'

function formatTime(ts: number): string {
  const d = new Date(ts)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

export function useGraphVersionActions() {
  const store = useKnowledgeGraphStore()
  /** 正在切换的版本 id，用于侧栏显示 loading */
  const switchingId = ref<number | null>(null)

  async function switchVersion(versionId: number) {
    if (store.bookId === null) return
    if (versionId === store.currentVersionId) return
    switchingId.value = versionId
    try {
      const name = store.versions.find(v => v.id === versionId)?.name || ''
      if (await store.loadVersion(versionId)) {
        ElMessage.success(`已切换到：${name}`)
      } else {
        ElMessage.error('该版本数据读取失败')
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '切换失败')
    } finally {
      switchingId.value = null
    }
  }

  async function renameVersion(versionId: number) {
    const version = store.versions.find(v => v.id === versionId)
    if (!version) return
    try {
      const { value: name } = await ElMessageBox.prompt('重命名图谱版本', '重命名', {
        inputPlaceholder: '输入新名称',
        inputValue: version.name,
        confirmButtonText: '保存',
        cancelButtonText: '取消'
      })
      const trimmed = (name || '').trim()
      if (!trimmed || trimmed === version.name) return
      if (await store.renameVersion(versionId, trimmed)) {
        ElMessage.success('已重命名')
      } else {
        ElMessage.error('重命名失败')
      }
    } catch (e: any) {
      if (e === 'cancel' || e === 'close') return
      ElMessage.error(e?.message || '操作失败')
    }
  }

  async function deleteVersion(versionId: number) {
    if (store.bookId === null) return
    const version = store.versions.find(v => v.id === versionId)
    if (!version) return
    try {
      await ElMessageBox.confirm(
        `确认删除版本「${version.name}」？该操作不可恢复。`,
        '删除版本',
        { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
      )
      if (await store.removeVersion(versionId)) {
        ElMessage.success('已删除')
      } else {
        ElMessage.error('删除失败')
      }
    } catch (e: any) {
      if (e === 'cancel' || e === 'close') return
      ElMessage.error(e?.message || '操作失败')
    }
  }

  async function exportVersion(versionId: number) {
    const version = store.versions.find(v => v.id === versionId)
    if (!version) return
    try {
      const data = await store.fetchVersionData(versionId)
      if (!data) {
        ElMessage.error('无法读取版本数据')
        return
      }
      const payload = {
        name: version.name,
        description: version.description,
        scope: version.scope,
        bookId: store.bookId,
        exportedAt: Date.now(),
        entities: data.entities,
        relations: data.relations
      }
      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `知识图谱_${version.name}_${Date.now()}.json`
      a.click()
      URL.revokeObjectURL(url)
      ElMessage.success('已导出')
    } catch (e: any) {
      ElMessage.error(e?.message || '导出失败')
    }
  }

  /** 把当前图谱存为新版本（弹窗问名称） */
  async function saveAsNewVersion() {
    if (store.bookId === null) return
    if (store.entities.length === 0) {
      ElMessage.warning('当前图谱为空，无需保存为版本')
      return
    }
    try {
      const { value: name } = await ElMessageBox.prompt('为新版本命名', '保存为新版本', {
        inputPlaceholder: '例如：第一卷人物关系',
        inputValue: `图谱_${formatTime(Date.now())}`,
        confirmButtonText: '保存',
        cancelButtonText: '取消'
      })
      const trimmed = (name || '').trim()
      if (!trimmed) {
        ElMessage.warning('版本名不能为空')
        return
      }
      const result = await store.saveAsNewVersion(trimmed)
      if (result.ok) ElMessage.success(result.message)
      else ElMessage.error(result.message)
    } catch (e: any) {
      if (e === 'cancel' || e === 'close') return
      ElMessage.error(e?.message || '保存失败')
    }
  }

  /** 创建空白版本（弹窗命名），切换到空白画布从白纸构建 */
  async function createBlankVersion() {
    if (store.bookId === null) return
    try {
      const { value: name } = await ElMessageBox.prompt('为新图谱版本命名', '新建图谱', {
        inputPlaceholder: '例如：第二卷',
        inputValue: `新图谱_${formatTime(Date.now())}`,
        confirmButtonText: '创建',
        cancelButtonText: '取消'
      })
      const trimmed = (name || '').trim()
      if (!trimmed) {
        ElMessage.warning('版本名不能为空')
        return
      }
      const result = await store.createBlankVersion(trimmed)
      if (result.ok) ElMessage.success(result.message)
      else ElMessage.error(result.message)
    } catch (e: any) {
      if (e === 'cancel' || e === 'close') return
      ElMessage.error(e?.message || '创建失败')
    }
  }

  /** 把导入文件解析出的数据存为新版本 */
  async function importFile(payload: {
    entities: GraphEntity[]
    relations: GraphRelation[]
    name: string
  }) {
    if (store.bookId === null) return
    try {
      const result = await store.importAsNewVersion(payload.name, {
        entities: payload.entities,
        relations: payload.relations
      })
      if (result.ok) ElMessage.success(result.message)
      else ElMessage.error(result.message)
    } catch (e: any) {
      ElMessage.error(e?.message || '导入失败')
    }
  }

  return {
    switchingId,
    switchVersion,
    renameVersion,
    deleteVersion,
    exportVersion,
    saveAsNewVersion,
    createBlankVersion,
    importFile
  }
}
