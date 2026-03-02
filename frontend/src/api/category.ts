import { apiClient } from './index'
import type { Category } from '@/types/category'

export interface CategoryListResponse {
  categories: Category[]
}

export function listCategories() {
  return apiClient.get<CategoryListResponse>('/categories')
}
