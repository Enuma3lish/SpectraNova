<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from '@/components/common/AppHeader.vue'
import AppSidebar from '@/components/common/AppSidebar.vue'

const sidebarCollapsed = ref(false)

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}
</script>

<template>
  <el-container class="min-h-screen">
    <el-header class="!p-0 h-auto">
      <AppHeader @toggle-sidebar="toggleSidebar" />
    </el-header>

    <el-container>
      <el-aside
        :width="sidebarCollapsed ? '0px' : '240px'"
        class="transition-all duration-300 overflow-hidden border-r border-gray-200"
      >
        <AppSidebar />
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
  scrollbar-width: thin;
}
</style>
