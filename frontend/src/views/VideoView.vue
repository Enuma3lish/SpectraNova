<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useVideoStore } from '@/stores/videoStore'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { formatDate } from '@/utils/formatDate'
import { formatDuration } from '@/utils/formatDuration'
import { formatViews } from '@/utils/formatViews'
import { ElMessage } from 'element-plus'

const route = useRoute()
const videoStore = useVideoStore()

const loading = ref(false)

async function loadVideo(id: number) {
  loading.value = true
  try {
    await videoStore.fetchVideo(id)
  } catch (err: unknown) {
    const message =
      err instanceof Error ? err.message : 'Failed to load video'
    ElMessage.error(message)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const id = Number(route.params.id)
  if (id) loadVideo(id)
})

watch(
  () => route.params.id,
  (newId) => {
    if (newId) loadVideo(Number(newId))
  },
)
</script>

<template>
  <div class="video-view">
    <LoadingSpinner :loading="loading" />

    <template v-if="!loading && videoStore.currentVideo">
      <div class="max-w-5xl mx-auto">
        <!-- Video Player -->
        <div class="bg-black rounded-lg overflow-hidden aspect-video mb-6">
          <video
            :src="videoStore.currentVideo.videoUrl"
            controls
            class="w-full h-full"
            preload="metadata"
          >
            Your browser does not support the video tag.
          </video>
        </div>

        <!-- Video Info -->
        <div class="bg-white rounded-lg p-6 shadow-sm">
          <h1 class="text-2xl font-bold text-gray-900 mb-3">
            {{ videoStore.currentVideo.title }}
          </h1>

          <div class="flex flex-wrap items-center gap-4 text-sm text-gray-500 mb-4">
            <span>{{ formatViews(Number(videoStore.currentVideo.views)) }} views</span>
            <span>{{ formatDate(videoStore.currentVideo.createdAt) }}</span>
            <span>{{ formatDuration(videoStore.currentVideo.duration) }}</span>
          </div>

          <div class="flex flex-wrap items-center gap-2 mb-4">
            <router-link
              :to="`/category/${videoStore.currentVideo.categoryId}`"
              class="no-underline"
            >
              <el-tag type="primary" effect="dark" round>
                {{ videoStore.currentVideo.categoryName }}
              </el-tag>
            </router-link>

            <el-tag
              v-for="tag in videoStore.currentVideo.tags"
              :key="tag.id"
              type="info"
              effect="plain"
              round
            >
              {{ tag.name }}
            </el-tag>
          </div>

          <el-divider />

          <div class="flex items-center gap-3 mb-4">
            <el-avatar :size="40">
              {{ videoStore.currentVideo.username?.charAt(0)?.toUpperCase() }}
            </el-avatar>
            <router-link
              :to="`/channel/${videoStore.currentVideo.userId}`"
              class="text-base font-semibold text-gray-800 hover:text-blue-600 no-underline"
            >
              {{ videoStore.currentVideo.username }}
            </router-link>
          </div>

          <div
            v-if="videoStore.currentVideo.description"
            class="bg-gray-50 rounded-lg p-4 text-sm text-gray-700 leading-relaxed whitespace-pre-wrap"
          >
            {{ videoStore.currentVideo.description }}
          </div>
        </div>
      </div>
    </template>

    <div
      v-if="!loading && !videoStore.currentVideo"
      class="flex flex-col items-center justify-center py-20 text-gray-400"
    >
      <p class="text-lg">Video not found</p>
    </div>
  </div>
</template>
