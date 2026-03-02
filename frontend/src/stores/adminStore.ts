import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as adminApi from '@/api/admin'
import { listTags } from '@/api/tag'
import type { User } from '@/types/user'
import type { AdminVideo } from '@/api/admin'
import type { Tag } from '@/types/tag'

export const useAdminStore = defineStore('admin', () => {
  // ---- State ----
  const users = ref<User[]>([])
  const totalUsers = ref(0)

  const videos = ref<AdminVideo[]>([])
  const totalVideos = ref(0)

  const tags = ref<Tag[]>([])

  // ---- User actions ----
  async function fetchUsers(page = 1, pageSize = 20) {
    const { data } = await adminApi.listUsers({ page, page_size: pageSize })
    users.value = data.users
    totalUsers.value = data.total
  }

  async function deleteUser(id: number) {
    await adminApi.deleteUser(id)
    users.value = users.value.filter((u) => u.id !== id)
    totalUsers.value -= 1
  }

  // ---- Video actions ----
  async function fetchVideos(page = 1, pageSize = 20) {
    const { data } = await adminApi.listVideos({ page, page_size: pageSize })
    videos.value = data.videos
    totalVideos.value = data.total
  }

  async function deleteVideo(id: number) {
    await adminApi.deleteVideo(id)
    videos.value = videos.value.filter((v) => v.id !== id)
    totalVideos.value -= 1
  }

  // ---- Tag actions (reuses tag API for listing) ----
  async function fetchTags() {
    const { data } = await listTags()
    tags.value = data.tags
  }

  async function createTag(name: string, slug: string) {
    const { data } = await adminApi.createTag({ name, slug })
    tags.value.push(data.tag)
  }

  async function updateTag(id: number, name: string, slug: string) {
    const { data } = await adminApi.updateTag(id, { name, slug })
    const idx = tags.value.findIndex((t) => t.id === id)
    if (idx !== -1) {
      tags.value[idx] = data.tag
    }
  }

  async function deleteTag(id: number) {
    await adminApi.deleteTag(id)
    tags.value = tags.value.filter((t) => t.id !== id)
  }

  return {
    users,
    totalUsers,
    videos,
    totalVideos,
    tags,
    fetchUsers,
    deleteUser,
    fetchVideos,
    deleteVideo,
    fetchTags,
    createTag,
    updateTag,
    deleteTag,
  }
})
