<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useTagStore } from '@/stores/tagStore'
import type { Tag } from '@/types/tag'

const MAX_TAGS = 5

const router = useRouter()
const tagStore = useTagStore()

const allTags = computed(() => tagStore.allTags)
const selectedTags = computed(() => tagStore.selectedTags)

const selectedIds = ref<number[]>([])

const hasSelection = computed(() => selectedIds.value.length > 0)

const isSelected = (tag: Tag): boolean => {
  return selectedIds.value.includes(tag.id)
}

const isDisabled = (tag: Tag): boolean => {
  return !isSelected(tag) && selectedIds.value.length >= MAX_TAGS
}

const toggleTag = async (tag: Tag) => {
  if (isSelected(tag)) {
    selectedIds.value = selectedIds.value.filter((id) => id !== tag.id)
  } else {
    if (selectedIds.value.length >= MAX_TAGS) return
    selectedIds.value.push(tag.id)
  }

  await tagStore.setMyTags(selectedIds.value)
}

function searchByTags() {
  if (!hasSelection.value) return
  router.push({
    name: 'Search',
    query: {
      tag_ids: selectedIds.value.join(','),
    },
  })
}

onMounted(async () => {
  await tagStore.fetchAllTags()
  await tagStore.fetchMyTags()
  selectedIds.value = selectedTags.value.map((t) => t.id)
})
</script>

<template>
  <div>
    <p class="text-xs text-gray-400 mb-2">
      Select up to {{ MAX_TAGS }} tags ({{ selectedIds.length }}/{{ MAX_TAGS }})
    </p>

    <div class="flex flex-wrap gap-2">
      <el-tag
        v-for="tag in allTags"
        :key="tag.id"
        :type="isSelected(tag) ? '' : 'info'"
        :effect="isSelected(tag) ? 'dark' : 'plain'"
        :class="[
          'cursor-pointer select-none transition-all',
          isDisabled(tag) ? 'opacity-40 cursor-not-allowed' : 'hover:opacity-80',
        ]"
        @click="toggleTag(tag)"
      >
        {{ tag.name }}
      </el-tag>
    </div>

    <p v-if="allTags.length === 0" class="text-xs text-gray-400 mt-2">
      No tags available
    </p>

    <!-- Search by selected tags -->
    <el-button
      v-if="hasSelection"
      type="primary"
      size="small"
      class="mt-3 w-full"
      @click="searchByTags"
    >
      <el-icon class="mr-1"><i class="el-icon-search" /></el-icon>
      Search by Tags ({{ selectedIds.length }})
    </el-button>
  </div>
</template>
