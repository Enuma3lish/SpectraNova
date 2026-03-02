import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as videoApi from '@/api/video'
import type { Video } from '@/types/video'

export const useVideoStore = defineStore('video', () => {
  // ---- State ----
  const recommendedVideos = ref<Video[]>([])
  const currentVideo = ref<Video | null>(null)
  const totalRecommended = ref(0)

  // ---- Actions ----
  async function fetchRecommended(
    page = 1,
    pageSize = 20,
    sessionId?: string,
  ) {
    const { data } = await videoApi.getRecommended({
      page,
      page_size: pageSize,
      session_id: sessionId,
    })
    recommendedVideos.value = data.videos
    totalRecommended.value = data.total
  }

  async function fetchVideo(id: number) {
    const { data } = await videoApi.getVideo(id)
    currentVideo.value = data
  }

  return {
    recommendedVideos,
    currentVideo,
    totalRecommended,
    fetchRecommended,
    fetchVideo,
  }
})
