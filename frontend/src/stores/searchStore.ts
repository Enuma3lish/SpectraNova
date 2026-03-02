import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'
import { searchVideos } from '@/api/search'
import type { Video } from '@/types/video'
import type { SearchFilters } from '@/types/search'

export const useSearchStore = defineStore('search', () => {
  // ---- State ----
  const query = ref('')
  const filters = reactive<SearchFilters>({})
  const results = ref<Video[]>([])
  const totalCount = ref(0)
  const page = ref(1)
  const pageSize = ref(20)

  // ---- Actions ----
  async function search() {
    const params: SearchFilters = {
      ...filters,
      query: query.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    }

    const { data } = await searchVideos(params)
    results.value = data.videos
    totalCount.value = data.total
  }

  function resetFilters() {
    query.value = ''
    Object.keys(filters).forEach((key) => {
      delete (filters as Record<string, unknown>)[key]
    })
    page.value = 1
    results.value = []
    totalCount.value = 0
  }

  return {
    query,
    filters,
    results,
    totalCount,
    page,
    pageSize,
    search,
    resetFilters,
  }
})
