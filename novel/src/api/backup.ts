import axios from 'axios'
import type { ApiResponse } from '@/types'

// 直接使用 axios 走 /api，绕开统一的拦截器（下载需要原始 response）
const http = axios.create({
  baseURL: '/api',
  timeout: 300000
})

export interface BackupModuleSummary {
  key: string
  label: string
  count: number
  present: boolean
}

export interface BackupPreview {
  version: string
  appName: string
  exportedAt: string
  maskedApiKey: boolean
  approxSizeKB: number
  stats: Record<string, number>
  modules: BackupModuleSummary[]
}

export interface BackupImportSummary {
  module: string
  label: string
  inserted: number
  skipped: number
  failed: number
}

export interface BackupImportResult {
  mode: 'overwrite' | 'merge'
  summary: BackupImportSummary[]
  totals: { inserted: number; skipped: number; failed: number }
}

export const backupAPI = {
  /**
   * 触发浏览器下载备份文件
   * @param maskApiKey 是否对 API 密钥脱敏
   */
  async downloadBackup(maskApiKey: boolean = false): Promise<{ filename: string; size: number }> {
    const res = await http.get('/backup/export', {
      params: { mask: maskApiKey ? 1 : 0 },
      responseType: 'blob'
    })
    const blob = res.data as Blob
    // 从 Content-Disposition 中解析 filename
    const dispo = (res.headers['content-disposition'] || '') as string
    const m = dispo.match(/filename="?([^"]+)"?/)
    const filename = m ? m[1] : `xingnovel-backup-${new Date().toISOString().slice(0, 10)}.json`

    // 触发下载
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    URL.revokeObjectURL(url)
    return { filename, size: blob.size }
  },

  /**
   * 预览备份包内容
   */
  preview(payload: any): Promise<ApiResponse<BackupPreview>> {
    return http.post<any, ApiResponse<BackupPreview>>('/backup/preview', { payload })
  },

  /**
   * 执行导入
   */
  importBackup(data: {
    payload: any
    mode: 'overwrite' | 'merge'
    modules: string[]
  }): Promise<ApiResponse<BackupImportResult>> {
    return http.post<any, ApiResponse<BackupImportResult>>('/backup/import', data, { timeout: 600000 })
  }
}
