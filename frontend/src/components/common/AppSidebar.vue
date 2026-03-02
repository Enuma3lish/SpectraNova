<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { useCategoryStore } from '@/stores/categoryStore'
import TagSelector from '@/components/tag/TagSelector.vue'

const categoryStore = useCategoryStore()

const categories = computed(() => categoryStore.categories)

onMounted(() => {
  categoryStore.fetchCategories()
})
</script>

<template>
  <el-aside width="240px" class="bg-white h-full overflow-y-auto">
    <div class="p-4">
      <!-- Categories Section -->
      <div class="mb-6">
        <h3 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">
          Categories
        </h3>
        <el-menu class="border-r-0">
          <el-menu-item
            v-for="category in categories"
            :key="category.id"
            :index="`/category/${category.id}`"
          >
            <router-link
              :to="`/category/${category.id}`"
              class="no-underline text-gray-700 w-full block"
            >
              {{ category.name }}
            </router-link>
          </el-menu-item>
        </el-menu>
      </div>

      <!-- Tags Section -->
      <div>
        <h3 class="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">
          My Tags
        </h3>
        <TagSelector />
      </div>
    </div>
  </el-aside>
</template>
