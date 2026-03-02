<script setup lang="ts">
import { computed, type Component } from 'vue'
import { useRoute } from 'vue-router'
import DefaultLayout from '@/layouts/DefaultLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'

const route = useRoute()

const layoutMap: Record<string, Component> = {
  auth: AuthLayout,
  admin: AdminLayout,
}

const layout = computed(() => {
  const name = route.meta?.layout
  if (name && name in layoutMap) {
    return layoutMap[name]
  }
  return DefaultLayout
})
</script>

<template>
  <component :is="layout" />
</template>
