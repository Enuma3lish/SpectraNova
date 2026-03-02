<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const query = ref('')

const handleSearch = () => {
  const trimmed = query.value.trim()
  if (!trimmed) return

  router.push({
    path: '/search',
    query: { q: trimmed },
  })
}
</script>

<template>
  <el-input
    v-model="query"
    placeholder="Search videos..."
    clearable
    size="large"
    class="w-full"
    @keyup.enter="handleSearch"
  >
    <template #suffix>
      <el-icon
        class="cursor-pointer hover:text-blue-500 transition-colors"
        @click="handleSearch"
      >
        <i class="el-icon-search" />
      </el-icon>
    </template>
  </el-input>
</template>
