import api from '../shared-api-instance'
import type { ApiResponse, ExperienceShare } from '../types/experience-share'

export const experienceShareAPI = {
  getAll: () => api.get<any, ApiResponse<ExperienceShare[]>>('/experience-shares'),
  getOne: (id: number) => api.get<any, ApiResponse<ExperienceShare>>(`/experience-shares/${id}`),
  create: (data: any) => api.post<any, ApiResponse<ExperienceShare>>('/experience-shares', data),
  update: (id: number, data: any) => api.put<any, ApiResponse<ExperienceShare>>(`/experience-shares/${id}`, data),
  delete: (id: number) => api.delete<any, ApiResponse>(`/experience-shares/${id}`),
  importPdf: (data: {
    pdf_file: {
      name: string
      size: number
      data_base64: string
    }
  }) => api.post<any, ApiResponse<Partial<ExperienceShare>>>('/experience-shares/import-pdf', data)
}

export default experienceShareAPI
