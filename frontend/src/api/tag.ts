import { apiClient } from './index'
import type { Tag } from '@/types/tag'

export interface TagListResponse {
  tags: Tag[]
}

export function listTags() {
  return apiClient.get<TagListResponse>('/tags')
}

export function getMyTags(sessionId?: string) {
  return apiClient.get<TagListResponse>('/tags/my', {
    params: sessionId ? { session_id: sessionId } : undefined,
  })
}

export function setMyTags(tagIds: number[], sessionId?: string) {
  return apiClient.put<TagListResponse>('/tags/my', {
    tag_ids: tagIds,
    session_id: sessionId,
  })
}
