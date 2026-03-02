<script setup lang="ts">
defineProps<{
  loading: boolean
}>()
</script>

<template>
  <div
    v-if="loading"
    v-loading="true"
    element-loading-text="Loading..."
    element-loading-background="rgba(255, 255, 255, 0.8)"
    class="flex items-center justify-center min-h-[200px] w-full"
  />
</template>
