<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getChannel, subscribe, unsubscribe } from '@/api/channel'
import { useAuthStore } from '@/stores/authStore'
import type { Channel } from '@/types/channel'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const route = useRoute()
const authStore = useAuthStore()

const channel = ref<Channel | null>(null)
const loading = ref(false)
const subscribing = ref(false)

const channelId = computed(() => Number(route.params.id))

const isOwnChannel = computed(() => {
  return authStore.isLoggedIn && authStore.user?.id === channelId.value
})

async function loadChannel() {
  loading.value = true
  try {
    const { data } = await getChannel(channelId.value)
    channel.value = data
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || 'Failed to load channel')
  } finally {
    loading.value = false
  }
}

async function handleSubscribe() {
  if (!channel.value) return
  subscribing.value = true
  try {
    await subscribe(channelId.value)
    channel.value.subscriberCount += 1
    ElMessage.success('Subscribed successfully')
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || 'Failed to subscribe')
  } finally {
    subscribing.value = false
  }
}

async function handleUnsubscribe() {
  if (!channel.value) return
  subscribing.value = true
  try {
    await unsubscribe(channelId.value)
    channel.value.subscriberCount = Math.max(0, channel.value.subscriberCount - 1)
    ElMessage.success('Unsubscribed successfully')
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.error || 'Failed to unsubscribe')
  } finally {
    subscribing.value = false
  }
}

onMounted(() => {
  loadChannel()
})
</script>

<template>
  <div class="max-w-5xl mx-auto">
    <LoadingSpinner :loading="loading" />

    <template v-if="!loading && channel">
      <!-- Channel Header -->
      <div class="bg-white rounded-lg shadow p-6 mb-6">
        <div class="flex items-center gap-4">
          <el-avatar
            :size="80"
            :src="channel.avatarUrl || undefined"
            class="flex-shrink-0"
          >
            {{ channel.displayName?.charAt(0)?.toUpperCase() || '?' }}
          </el-avatar>

          <div class="flex-1">
            <h1 class="text-2xl font-bold text-gray-800">
              {{ channel.displayName }}
            </h1>
            <p class="text-gray-500 mt-1">
              {{ channel.subscriberCount }} subscriber{{ channel.subscriberCount !== 1 ? 's' : '' }}
            </p>
            <p v-if="channel.monthlyFee > 0" class="text-gray-500 mt-1">
              Monthly fee: ${{ channel.monthlyFee.toFixed(2) }}
            </p>
          </div>

          <!-- Subscribe / Unsubscribe button -->
          <div v-if="authStore.isLoggedIn && !isOwnChannel" class="flex-shrink-0">
            <el-button
              type="primary"
              :loading="subscribing"
              @click="handleSubscribe"
            >
              Subscribe
            </el-button>
            <el-button
              :loading="subscribing"
              @click="handleUnsubscribe"
            >
              Unsubscribe
            </el-button>
          </div>
        </div>
      </div>
    </template>

    <div v-if="!loading && !channel" class="text-center text-gray-400 py-20">
      Channel not found.
    </div>
  </div>
</template>
