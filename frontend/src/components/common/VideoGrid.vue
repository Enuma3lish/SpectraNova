<script setup lang="ts">
import type { Video } from '@/types/video'
import VideoCard from '@/components/common/VideoCard.vue'

defineProps<{
  videos: Video[]
}>()
</script>

<template>
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
    <VideoCard
      v-for="video in videos"
      :key="video.id"
      :video="video"
    />
  </div>

  <div
    v-if="videos.length === 0"
    class="flex flex-col items-center justify-center py-16 text-gray-400"
  >
    <el-icon :size="48"><i class="el-icon-video-camera" /></el-icon>
    <p class="mt-4 text-sm">No videos found</p>
  </div>
</template>
