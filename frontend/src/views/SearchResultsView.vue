<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
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

// Local filter state bound to UI controls
const selectedCategory = ref<number | undefined>(undefined)
const minDuration = ref<number | undefined>(undefined)
const maxDuration = ref<number | undefined>(undefined)
const accessType = ref<string | undefined>(undefined)
const sortBy = ref<string | undefined>(undefined)

const sortOptions = [
  { label: 'Most Views', value: 'views_desc' },
  { label: 'Least Views', value: 'views_asc' },
  { label: 'Newest', value: 'date_desc' },
  { label: 'Oldest', value: 'date_asc' },
]

const accessOptions = [
  { label: 'All', value: undefined },
  { label: 'Public', value: 'public' },
  { label: 'Members Only', value: 'member' },
]

async function performSearch() {
  loading.value = true
  try {
    // Sync local filter state to the store
    searchStore.filters.category_id = selectedCategory.value
    searchStore.filters.min_duration = minDuration.value
    searchStore.filters.max_duration = maxDuration.value
    searchStore.filters.access_type = accessType.value
    searchStore.filters.sort_by = sortBy.value

    await searchStore.search()
  } catch (err: unknown) {
    const message =
      err instanceof Error ? err.message : 'Search failed'
    ElMessage.error(message)
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  searchStore.page = page
  performSearch()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

function handleFilterChange() {
  searchStore.page = 1
  performSearch()
}

function clearFilters() {
  selectedCategory.value = undefined
  minDuration.value = undefined
  maxDuration.value = undefined
  accessType.value = undefined
  sortBy.value = undefined
  handleFilterChange()
}

// Initialize from route query
function syncFromRoute() {
  const q = (route.query.q as string) || ''
  searchStore.query = q
  searchStore.page = 1
}

onMounted(async () => {
  syncFromRoute()
  // Ensure categories are loaded for the filter dropdown
  if (categoryStore.categories.length === 0) {
    try {
      await categoryStore.fetchCategories()
    } catch {
      // Categories are non-critical; continue without them
    }
  }
  await performSearch()
})

// Re-search when query param changes (e.g. user searches from header)
watch(
  () => route.query.q,
  () => {
    syncFromRoute()
    performSearch()
  },
)
</script>

<template>
  <div class="search-results-view">
    <h1 class="text-2xl font-bold text-gray-800 mb-6">
      <template v-if="searchStore.query">
        Results for "{{ searchStore.query }}"
      </template>
      <template v-else>
        Browse Videos
      </template>
    </h1>

    <div class="flex gap-6">
      <!-- Filter Sidebar -->
      <aside class="w-64 flex-shrink-0">
        <el-card shadow="never" class="sticky top-4">
          <template #header>
            <div class="flex items-center justify-between">
              <span class="font-semibold text-gray-700">Filters</span>
              <el-button text type="primary" size="small" @click="clearFilters">
                Clear
              </el-button>
            </div>
          </template>

          <div class="space-y-5">
            <!-- Category -->
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Category</label>
              <el-select
                v-model="selectedCategory"
                placeholder="All Categories"
                clearable
                class="w-full"
                @change="handleFilterChange"
              >
                <el-option
                  v-for="cat in categoryStore.categories"
                  :key="cat.id"
                  :label="cat.name"
                  :value="cat.id"
                />
              </el-select>
            </div>

            <!-- Duration Range -->
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Duration (seconds)</label>
              <div class="flex items-center gap-2">
                <el-input-number
                  v-model="minDuration"
                  :min="0"
                  placeholder="Min"
                  controls-position="right"
                  size="small"
                  class="flex-1"
                  @change="handleFilterChange"
                />
                <span class="text-gray-400">-</span>
                <el-input-number
                  v-model="maxDuration"
                  :min="0"
                  placeholder="Max"
                  controls-position="right"
                  size="small"
                  class="flex-1"
                  @change="handleFilterChange"
                />
              </div>
            </div>

            <!-- Access Type -->
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Access Type</label>
              <el-select
                v-model="accessType"
                placeholder="All"
                clearable
                class="w-full"
                @change="handleFilterChange"
              >
                <el-option
                  v-for="opt in accessOptions"
                  :key="opt.value ?? 'all'"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </div>

            <!-- Sort By -->
            <div>
              <label class="block text-sm font-medium text-gray-600 mb-1">Sort By</label>
              <el-select
                v-model="sortBy"
                placeholder="Default"
                clearable
                class="w-full"
                @change="handleFilterChange"
              >
                <el-option
                  v-for="opt in sortOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </div>
          </div>
        </el-card>
      </aside>

      <!-- Results Area -->
      <div class="flex-1 min-w-0">
        <LoadingSpinner :loading="loading" />

        <template v-if="!loading">
          <!-- Result count -->
          <p class="text-sm text-gray-500 mb-4">
            {{ searchStore.totalCount }} result{{ searchStore.totalCount !== 1 ? 's' : '' }} found
          </p>

          <VideoGrid :videos="searchStore.results" />

          <Pagination
            :total="searchStore.totalCount"
            :page-size="searchStore.pageSize"
            :current-page="searchStore.page"
            @update:current-page="handlePageChange"
          />
        </template>
      </div>
    </div>
  </div>
</template>
