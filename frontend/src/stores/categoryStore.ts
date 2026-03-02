import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listCategories } from '@/api/category'
import type { Category } from '@/types/category'

export const useCategoryStore = defineStore('category', () => {
  // ---- State ----
  const categories = ref<Category[]>([])

  // ---- Actions ----
  async function fetchCategories() {
    const { data } = await listCategories()
    categories.value = data.categories
  }

  return {
    categories,
    fetchCategories,
  }
})
