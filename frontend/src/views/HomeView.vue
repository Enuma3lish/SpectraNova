<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useVideoStore } from '@/stores/videoStore'
import { useTagStore } from '@/stores/tagStore'
import { useAuthStore } from '@/stores/authStore'
import VideoGrid from '@/components/common/VideoGrid.vue'
import Pagination from '@/components/common/Pagination.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { ElMessage } from 'element-plus'

const videoStore = useVideoStore()
const tagStore = useTagStore()
const authStore = useAuthStore()

const loading = ref(false)
const currentPage = ref(1)
const pageSize = 20

async function loadVideos(page: number) {
  loading.value = true
  try {
    const sessionId = authStore.isLoggedIn ? undefined : tagStore.sessionId
    await videoStore.fetchRecommended(page, pageSize, sessionId)
  } catch (err: unknown) {
    const message =
      err instanceof Error ? err.message : 'Failed to load videos'
    ElMessage.error(message)
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadVideos(page)
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

onMounted(() => {
  loadVideos(currentPage.value)
})
</script>

<template>
  <div class="home-view">
    <h1 class="text-2xl font-bold text-gray-800 mb-6">Recommended for You</h1>

    <LoadingSpinner :loading="loading" />

    <template v-if="!loading">
      <VideoGrid :videos="videoStore.recommendedVideos" />

      <Pagination
        :total="videoStore.totalRecommended"
        :page-size="pageSize"
        :current-page="currentPage"
        @update:current-page="handlePageChange"
      />
    </template>
  </div>
</template>
