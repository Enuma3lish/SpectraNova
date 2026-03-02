<script setup lang="ts">
const props = defineProps<{
  total: number
  pageSize: number
  currentPage: number
}>()

const emit = defineEmits<{
  'update:currentPage': [page: number]
}>()

const handleCurrentChange = (page: number) => {
  emit('update:currentPage', page)
}
</script>

<template>
  <div v-if="props.total > props.pageSize" class="flex justify-center mt-6">
    <el-pagination
      :current-page="props.currentPage"
      :page-size="props.pageSize"
      :total="props.total"
      layout="prev, pager, next"
      background
      @current-change="handleCurrentChange"
    />
  </div>
</template>
