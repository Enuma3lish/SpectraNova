<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Video } from '@/types/video'

const props = defineProps<{
  video: Video
}>()

const router = useRouter()

const formattedDuration = computed(() => {
  const totalSeconds = props.video.duration
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  }
  return `${minutes}:${String(seconds).padStart(2, '0')}`
})

const formattedViews = computed(() => {
  const count = Number(props.video.views) || 0
  if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M views`
  if (count >= 1_000) return `${(count / 1_000).toFixed(1)}K views`
  return `${count} views`
})

const navigateToVideo = () => {
  router.push({ name: 'Video', params: { id: props.video.id } })
}

const navigateToChannel = (event: Event) => {
  event.stopPropagation()
  router.push({ name: 'Channel', params: { id: props.video.userId } })
}
</script>

<template>
  <el-card
    shadow="hover"
    class="cursor-pointer overflow-hidden transition-transform hover:-translate-y-1"
    :body-style="{ padding: '0' }"
    @click="navigateToVideo"
  >
    <!-- Thumbnail -->
    <div class="relative w-full aspect-video bg-gray-200 overflow-hidden">
      <img
        v-if="video.thumbnailUrl"
        :src="video.thumbnailUrl"
        :alt="video.title"
        class="w-full h-full object-cover"
      />
      <div
        v-else
        class="w-full h-full flex items-center justify-center bg-gray-300 text-gray-500"
      >
        <el-icon :size="48"><i class="el-icon-video-camera" /></el-icon>
      </div>

      <!-- Duration badge -->
      <span
        class="absolute bottom-2 right-2 bg-black/80 text-white text-xs px-1.5 py-0.5 rounded"
      >
        {{ formattedDuration }}
      </span>

      <!-- Member-only badge -->
      <el-tag
        v-if="video.accessTier > 0"
        type="warning"
        size="small"
        class="absolute top-2 left-2"
      >
        Members
      </el-tag>
    </div>

    <!-- Info -->
    <div class="p-3">
      <h3 class="text-sm font-medium text-gray-900 line-clamp-2 mb-1" :title="video.title">
        {{ video.title }}
      </h3>

      <p
        class="text-xs text-blue-600 hover:underline cursor-pointer mb-1"
        @click="navigateToChannel"
      >
        {{ video.username }}
      </p>

      <p class="text-xs text-gray-500">
        {{ formattedViews }}
      </p>
    </div>
  </el-card>
</template>
