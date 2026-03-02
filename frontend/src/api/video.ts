import { apiClient } from './index'
import type { Video } from '@/types/video'

export interface RecommendedParams {
  page?: number
  page_size?: number
  session_id?: string
}

export interface CreateVideoPayload {
  user_id: number
  category_id: number
  title: string
  description: string
  video_url: string
  thumbnail_url: string
  duration: number
  access_tier: string
  tag_ids: number[]
}

export interface RecommendedResponse {
  videos: Video[]
  total: number
}

export function getRecommended(params: RecommendedParams = {}) {
  return apiClient.get<RecommendedResponse>('/recommended', { params })
}

export function getVideo(id: number) {
  return apiClient.get<Video>(`/videos/${id}`)
}

export function createVideo(data: CreateVideoPayload) {
  return apiClient.post<Video>('/videos', data)
}

export function updateVideo(id: number, data: Partial<CreateVideoPayload>) {
  return apiClient.put<Video>(`/videos/${id}`, data)
}

export function deleteVideo(id: number) {
  return apiClient.delete(`/videos/${id}`)
}

export function togglePublish(id: number, published: boolean) {
  return apiClient.patch(`/videos/${id}/publish`, { published })
}
