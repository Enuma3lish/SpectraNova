import { apiClient } from './index'
import type { Video } from '@/types/video'
import type { SearchFilters } from '@/types/search'

export interface SearchResponse {
  videos: Video[]
  total: number
}

export function searchVideos(filters: SearchFilters) {
  // Build params manually so tag_ids becomes repeated query params: tag_ids=1&tag_ids=2
  const params = new URLSearchParams()
  if (filters.query) params.append('query', filters.query)
  if (filters.category_id != null) params.append('category_id', String(filters.category_id))
  if (filters.min_duration != null) params.append('min_duration', String(filters.min_duration))
  if (filters.max_duration != null) params.append('max_duration', String(filters.max_duration))
  if (filters.start_date) params.append('start_date', filters.start_date)
  if (filters.end_date) params.append('end_date', filters.end_date)
  if (filters.sort_by) params.append('sort_by', filters.sort_by)
  if (filters.access_type) params.append('access_type', filters.access_type)
  if (filters.page != null) params.append('page', String(filters.page))
  if (filters.page_size != null) params.append('page_size', String(filters.page_size))
  if (filters.tag_ids && filters.tag_ids.length > 0) {
    for (const id of filters.tag_ids) {
      params.append('tag_ids', String(id))
    }
  }

  return apiClient.get<SearchResponse>('/search', {
    params,
  })
}
