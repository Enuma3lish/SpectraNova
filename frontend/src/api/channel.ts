import { apiClient } from './index'
import type { Channel } from '@/types/channel'

export function getChannel(id: number) {
  return apiClient.get<Channel>(`/channels/${id}`)
}

export function subscribe(id: number) {
  return apiClient.post(`/channels/${id}/subscribe`)
}

export function unsubscribe(id: number) {
  return apiClient.delete(`/channels/${id}/subscribe`)
}
