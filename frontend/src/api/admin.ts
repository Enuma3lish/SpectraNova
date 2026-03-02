import { apiClient } from './index'
import type { User } from '@/types/user'
import type { Tag } from '@/types/tag'

export interface PaginationParams {
  page?: number
  page_size?: number
}

export interface UsersResponse {
  users: User[]
  total: number
}

export interface AdminVideo {
  id: number
  title: string
  username: string
  userId: number
  categoryName: string
  accessTier: number
  isPublished: boolean
  isHidden: boolean
  viewsMember: number
  viewsNonMember: number
  createdAt: string
}

export interface VideosResponse {
  videos: AdminVideo[]
  total: number
}

export interface TagResponse {
  tag: Tag
}

// ---- User management ----

export function listUsers(params: PaginationParams = {}) {
  return apiClient.get<UsersResponse>('/admin/users', { params })
}

export function deleteUser(id: number) {
  return apiClient.delete(`/admin/users/${id}`)
}

// ---- Video management ----

export function listVideos(params: PaginationParams = {}) {
  return apiClient.get<VideosResponse>('/admin/videos', { params })
}

export function deleteVideo(id: number) {
  return apiClient.delete(`/admin/videos/${id}`)
}

// ---- Tag management ----

export interface TagPayload {
  name: string
  slug: string
}

export function createTag(data: TagPayload) {
  return apiClient.post<TagResponse>('/admin/tags', data)
}

export function updateTag(id: number, data: TagPayload) {
  return apiClient.put<TagResponse>(`/admin/tags/${id}`, data)
}

export function deleteTag(id: number) {
  return apiClient.delete(`/admin/tags/${id}`)
}
