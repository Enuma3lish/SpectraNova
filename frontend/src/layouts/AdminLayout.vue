<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppHeader from '@/components/common/AppHeader.vue'

const route = useRoute()

const activeMenu = computed(() => {
  const path = route.path
  if (path.includes('/admin/tags')) return '/admin/tags'
  if (path.includes('/admin/users')) return '/admin/users'
  return '/admin/users'
})
</script>

<template>
  <el-container class="min-h-screen">
    <el-header class="!p-0 h-auto">
      <AppHeader />
    </el-header>

    <el-container>
      <el-aside width="220px" class="border-r border-gray-200 bg-white">
        <div class="p-4">
          <h2 class="text-lg font-semibold text-gray-700 mb-4">Admin Panel</h2>
        </div>
        <el-menu
          :default-active="activeMenu"
          router
          class="border-r-0"
        >
          <el-menu-item index="/admin/users">
            <el-icon><i class="el-icon-user" /></el-icon>
            <template #title>Users</template>
          </el-menu-item>
          <el-menu-item index="/admin/tags">
            <el-icon><i class="el-icon-price-tag" /></el-icon>
            <template #title>Tags</template>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-main class="bg-gray-50 p-6">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.el-aside {
  overflow-y: auto;
}
</style>
