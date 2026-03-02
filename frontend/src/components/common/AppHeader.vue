<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import SearchBar from '@/components/common/SearchBar.vue'

const emit = defineEmits<{
  'toggle-sidebar': []
}>()

const router = useRouter()
const authStore = useAuthStore()

const isLoggedIn = computed(() => authStore.isLoggedIn)
const isAdmin = computed(() => authStore.isAdmin)
const displayName = computed(() => authStore.user?.displayName ?? '')

const handleLogin = () => {
  router.push({ name: 'Login' })
}

const handleLogout = () => {
  authStore.logout()
  router.push({ name: 'Home' })
}

const handleCommand = (command: string) => {
  switch (command) {
    case 'dashboard':
      router.push({ name: 'Dashboard' })
      break
    case 'admin':
      router.push({ name: 'Admin' })
      break
    case 'logout':
      handleLogout()
      break
  }
}
</script>

<template>
  <el-header class="flex items-center justify-between bg-white shadow-sm border-b border-gray-200 h-16 px-4">
    <!-- Left: Logo and sidebar toggle -->
    <div class="flex items-center gap-3">
      <el-button
        text
        class="!p-2"
        @click="emit('toggle-sidebar')"
      >
        <el-icon :size="20">
          <i class="el-icon-menu" />
        </el-icon>
      </el-button>

      <router-link
        to="/"
        class="text-xl font-bold text-blue-600 no-underline hover:text-blue-700 transition-colors"
      >
        FenzVideo
      </router-link>
    </div>

    <!-- Center: Search bar -->
    <div class="flex-1 max-w-xl mx-4">
      <SearchBar />
    </div>

    <!-- Right: Auth actions -->
    <div class="flex items-center">
      <template v-if="isLoggedIn">
        <el-dropdown trigger="click" @command="handleCommand">
          <span class="flex items-center cursor-pointer text-gray-700 hover:text-blue-600 transition-colors">
            <el-avatar :size="32" class="mr-2">
              {{ displayName.charAt(0).toUpperCase() }}
            </el-avatar>
            <span class="text-sm font-medium">{{ displayName }}</span>
            <el-icon class="ml-1"><i class="el-icon-arrow-down" /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="dashboard">Dashboard</el-dropdown-item>
              <el-dropdown-item v-if="isAdmin" command="admin">Admin</el-dropdown-item>
              <el-dropdown-item divided command="logout">Logout</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </template>

      <template v-else>
        <el-button type="primary" @click="handleLogin">Login</el-button>
      </template>
    </div>
  </el-header>
</template>
