import { apiClient } from './index'
import type { Video } from '@/types/video'
import type { SearchFilters } from '@/types/search'

export interface SearchResponse {
  videos: Video[]
  total: number
}

export function searchVideos(filters: SearchFilters) {
  return apiClient.get<SearchResponse>('/search', { params: filters })
}
