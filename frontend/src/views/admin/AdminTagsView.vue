<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { useAdminStore } from '@/stores/adminStore'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const adminStore = useAdminStore()

const loading = ref(false)

// ---- Dialog state ----
const dialogVisible = ref(false)
const dialogMode = ref<'create' | 'edit'>('create')
const editingTagId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const tagForm = reactive({
  name: '',
  slug: '',
})

const formRules = {
  name: [{ required: true, message: 'Tag name is required', trigger: 'blur' }],
  slug: [{ required: true, message: 'Slug is required', trigger: 'blur' }],
}

// ---- Confirm dialog state ----
const confirmVisible = ref(false)
const pendingDeleteId = ref<number | null>(null)

async function loadTags() {
  loading.value = true
  try {
    await adminStore.fetchTags()
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || 'Failed to load tags')
  } finally {
    loading.value = false
  }
}

// ---- Create / Edit ----
function openCreateDialog() {
  dialogMode.value = 'create'
  editingTagId.value = null
  tagForm.name = ''
  tagForm.slug = ''
  dialogVisible.value = true
}

function openEditDialog(id: number, name: string, slug: string) {
  dialogMode.value = 'edit'
  editingTagId.value = id
  tagForm.name = name
  tagForm.slug = slug
  dialogVisible.value = true
}

async function submitForm() {
  if (!formRef.value) return

  try {
    const valid = await formRef.value.validate()
    if (!valid) return
  } catch {
    return
  }

  try {
    if (dialogMode.value === 'create') {
      await adminStore.createTag(tagForm.name, tagForm.slug)
      ElMessage.success('Tag created successfully')
    } else if (editingTagId.value !== null) {
      await adminStore.updateTag(editingTagId.value, tagForm.name, tagForm.slug)
      ElMessage.success('Tag updated successfully')
    }
    dialogVisible.value = false
  } catch (err: any) {
    const message = err?.response?.data?.error || 'Failed to save tag'
    ElMessage.error(message)
  }
}

// ---- Delete ----
function requestDelete(id: number) {
  pendingDeleteId.value = id
  confirmVisible.value = true
}

async function confirmDelete() {
  if (pendingDeleteId.value === null) return

  try {
    await adminStore.deleteTag(pendingDeleteId.value)
    ElMessage.success('Tag deleted successfully')
  } catch (err: any) {
    const message = err?.response?.data?.error || 'Failed to delete tag'
    ElMessage.error(message)
  } finally {
    pendingDeleteId.value = null
  }
}

function cancelDelete() {
  pendingDeleteId.value = null
}

onMounted(() => {
  loadTags()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-6">
      <h1 class="text-2xl font-bold text-gray-800">Tag Management</h1>
      <el-button type="primary" @click="openCreateDialog">Create Tag</el-button>
    </div>

    <LoadingSpinner :loading="loading" />

    <template v-if="!loading">
      <el-table
        :data="adminStore.tags"
        stripe
        class="w-full"
        empty-text="No tags found"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="Name" min-width="200" />
        <el-table-column prop="slug" label="Slug" min-width="200" />
        <el-table-column label="Actions" width="180" align="center">
          <template #default="{ row }">
            <div class="flex items-center justify-center gap-2">
              <el-button
                type="primary"
                size="small"
                @click="openEditDialog(row.id, row.name, row.slug)"
              >
                Edit
              </el-button>
              <el-button
                type="danger"
                size="small"
                @click="requestDelete(row.id)"
              >
                Delete
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </template>

    <!-- Create / Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? 'Create Tag' : 'Edit Tag'"
      width="480px"
      align-center
      @closed="formRef?.resetFields()"
    >
      <el-form
        ref="formRef"
        :model="tagForm"
        :rules="formRules"
        label-width="80px"
        class="mt-2"
      >
        <el-form-item label="Name" prop="name">
          <el-input v-model="tagForm.name" placeholder="Enter tag name" />
        </el-form-item>
        <el-form-item label="Slug" prop="slug">
          <el-input v-model="tagForm.slug" placeholder="Enter slug (e.g. my-tag)" />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="flex justify-end gap-2">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="submitForm">
            {{ dialogMode === 'create' ? 'Create' : 'Save' }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Delete Confirm Dialog -->
    <ConfirmDialog
      :visible="confirmVisible"
      title="Delete Tag"
      message="Are you sure you want to delete this tag? This action cannot be undone."
      @update:visible="confirmVisible = $event"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>
