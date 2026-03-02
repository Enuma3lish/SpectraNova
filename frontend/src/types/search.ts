export interface SearchFilters {
  query?: string
  category_id?: number
  min_duration?: number
  max_duration?: number
  start_date?: string
  end_date?: string
  sort_by?: string
  access_type?: string
  page?: number
  page_size?: number
}
