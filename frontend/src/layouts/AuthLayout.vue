<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-gray-100 flex items-center justify-center px-4">
    <div class="w-full max-w-md">
      <el-card shadow="always" class="rounded-lg">
        <template #header>
          <div class="text-center">
            <h1 class="text-2xl font-bold text-gray-800">FenzVideo</h1>
          </div>
        </template>
        <router-view />
      </el-card>
    </div>
  </div>
</template>
