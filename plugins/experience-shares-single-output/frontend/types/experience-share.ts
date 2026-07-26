export interface ExperienceShare {
  id: number
  title: string
  summary: string
  content: string
  cover_url?: string | null
  pdf_file_url?: string | null
  pdf_file_name?: string | null
  pdf_file_size?: number
  create_type: 'manual' | 'pdf_import'
  author_id?: number | null
  author_name?: string | null
  created_at: string
  updated_at: string
  status?: string | null
  pdf_parse_status?: 'success' | 'empty' | 'failed' | string | null
  pdf_parse_result?: string | null
  source_file_name?: string | null
}

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
}
