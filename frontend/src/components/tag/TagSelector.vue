<script setup lang="ts">
import { onMounted, computed, ref } from 'vue'
import { useTagStore } from '@/stores/tagStore'
import type { Tag } from '@/types/tag'

const MAX_TAGS = 5

const tagStore = useTagStore()

const allTags = computed(() => tagStore.allTags)
const selectedTags = computed(() => tagStore.selectedTags)

const selectedIds = ref<number[]>([])

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
  </div>
</template>
