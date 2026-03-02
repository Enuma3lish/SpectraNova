import type { Tag } from './tag'

export interface Video {
  id: number
  userId: number
  username: string
  categoryId: number
  categoryName: string
  title: string
  description: string
  videoUrl: string
  thumbnailUrl: string
  duration: number
  views: number
  accessTier: number
  isPublished: boolean
  tags: Tag[]
  createdAt: string
}
