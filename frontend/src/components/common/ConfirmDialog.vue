<script setup lang="ts">
const props = defineProps<{
  visible: boolean
  title?: string
  message?: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  confirm: []
  cancel: []
}>()

const handleClose = () => {
  emit('update:visible', false)
  emit('cancel')
}

const handleConfirm = () => {
  emit('update:visible', false)
  emit('confirm')
}

const handleCancel = () => {
  emit('update:visible', false)
  emit('cancel')
}
</script>

<template>
  <el-dialog
    :model-value="props.visible"
    :title="props.title ?? 'Confirm'"
    width="400px"
    align-center
    @close="handleClose"
  >
    <p class="text-gray-600">
      {{ props.message ?? 'Are you sure you want to proceed?' }}
    </p>

    <template #footer>
      <div class="flex justify-end gap-2">
        <el-button @click="handleCancel">Cancel</el-button>
        <el-button type="primary" @click="handleConfirm">Confirm</el-button>
      </div>
    </template>
  </el-dialog>
</template>
