<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSearchStore } from '@/stores/searchStore'
import { useCategoryStore } from '@/stores/categoryStore'
import VideoGrid from '@/components/common/VideoGrid.vue'
import Pagination from '@/components/common/Pagination.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { ElMessage } from 'element-plus'

const route = useRoute()
const searchStore = useSearchStore()
const categoryStore = useCategoryStore()

const loading = ref(false)
const currentPage = ref(1)
const pageSize = 20

const categoryId = computed(() => Number(route.params.id))

const categoryName = computed(() => {
  const cat = categoryStore.categories.find((c) => c.id === categoryId.value)
  return cat?.name ?? 'Category'
})

async function loadVideos(page: number) {
  loading.value = true
  try {
    searchStore.query = ''
    searchStore.filters.category_id = categoryId.value
    searchStore.page = page
    searchStore.pageSize = pageSize
    await searchStore.search()
  } catch (err: unknown) {
    const message =
      err instanceof Error ? err.message : 'Failed to load category videos'
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

onMounted(async () => {
  // Ensure categories are loaded so we can display the name
  if (categoryStore.categories.length === 0) {
    try {
      await categoryStore.fetchCategories()
    } catch {
      // Non-critical; the title will fall back to "Category"
    }
  }
  await loadVideos(currentPage.value)
})

// Reload when navigating to a different category
watch(
  () => route.params.id,
  () => {
    currentPage.value = 1
    loadVideos(1)
  },
)
</script>

<template>
  <div class="category-view">
    <h1 class="text-2xl font-bold text-gray-800 mb-6">{{ categoryName }}</h1>

    <LoadingSpinner :loading="loading" />

    <template v-if="!loading">
      <VideoGrid :videos="searchStore.results" />

      <Pagination
        :total="searchStore.totalCount"
        :page-size="pageSize"
        :current-page="currentPage"
        @update:current-page="handlePageChange"
      />
    </template>
  </div>
</template>
