import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as tagApi from '@/api/tag'
import type { Tag } from '@/types/tag'

function getOrCreateSessionId(): string {
  let sessionId = localStorage.getItem('sessionId')
  if (!sessionId) {
    sessionId = crypto.randomUUID()
    localStorage.setItem('sessionId', sessionId)
  }
  return sessionId
}

export const useTagStore = defineStore('tag', () => {
  // ---- State ----
  const allTags = ref<Tag[]>([])
  const selectedTags = ref<Tag[]>([])
  const sessionId = ref<string>(getOrCreateSessionId())

  // ---- Actions ----
  async function fetchAllTags() {
    const { data } = await tagApi.listTags()
    allTags.value = data.tags
  }

  async function fetchMyTags() {
    const { data } = await tagApi.getMyTags(sessionId.value)
    selectedTags.value = data.tags
  }

  async function setMyTags(tagIds: number[]) {
    const { data } = await tagApi.setMyTags(tagIds, sessionId.value)
    selectedTags.value = data.tags
  }

  return {
    allTags,
    selectedTags,
    sessionId,
    fetchAllTags,
    fetchMyTags,
    setMyTags,
  }
})
