<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAdminStore } from '@/stores/adminStore'
import { useAuthStore } from '@/stores/authStore'
import { formatDate } from '@/utils/formatDate'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const adminStore = useAdminStore()
const authStore = useAuthStore()

const loading = ref(false)
const currentPage = ref(1)
const pageSize = 20

const confirmVisible = ref(false)
const pendingDeleteId = ref<number | null>(null)

async function loadUsers(page: number) {
  loading.value = true
  try {
    await adminStore.fetchUsers(page, pageSize)
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || 'Failed to load users')
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadUsers(page)
}

function requestDelete(userId: number) {
  if (authStore.user && authStore.user.id === userId) {
    ElMessage.error('Cannot delete your own admin account')
    return
  }
  pendingDeleteId.value = userId
  confirmVisible.value = true
}

async function confirmDelete() {
  if (pendingDeleteId.value === null) return

  try {
    await adminStore.deleteUser(pendingDeleteId.value)
    ElMessage.success('User deleted successfully')
    await loadUsers(currentPage.value)
  } catch (err: any) {
    const message = err?.response?.data?.error || 'Failed to delete user'
    ElMessage.error(message)
  } finally {
    pendingDeleteId.value = null
  }
}

function cancelDelete() {
  pendingDeleteId.value = null
}

onMounted(() => {
  loadUsers(currentPage.value)
})
</script>

<template>
  <div>
    <h1 class="text-2xl font-bold text-gray-800 mb-6">User Management</h1>

    <LoadingSpinner :loading="loading" />

    <template v-if="!loading">
      <el-table
        :data="adminStore.users"
        stripe
        class="w-full"
        empty-text="No users found"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="Username" min-width="140" />
        <el-table-column prop="displayName" label="Display Name" min-width="160" />
        <el-table-column prop="role" label="Role" width="100">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">
              {{ row.role }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Hidden" width="90" align="center">
          <template #default="{ row }">
            <el-tag :type="row.isHidden ? 'warning' : 'success'" size="small">
              {{ row.isHidden ? 'Yes' : 'No' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="Created At" width="160">
          <template #default="{ row }">
            {{ formatDate(row.createdAt) }}
          </template>
        </el-table-column>
        <el-table-column label="Actions" width="120" align="center">
          <template #default="{ row }">
            <el-button
              type="danger"
              size="small"
              :disabled="authStore.user?.id === row.id"
              @click="requestDelete(row.id)"
            >
              Delete
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <Pagination
        :total="adminStore.totalUsers"
        :page-size="pageSize"
        :current-page="currentPage"
        @update:current-page="handlePageChange"
      />
    </template>

    <ConfirmDialog
      :visible="confirmVisible"
      title="Delete User"
      message="Are you sure you want to delete this user? This action cannot be undone."
      @update:visible="confirmVisible = $event"
      @confirm="confirmDelete"
      @cancel="cancelDelete"
    />
  </div>
</template>
