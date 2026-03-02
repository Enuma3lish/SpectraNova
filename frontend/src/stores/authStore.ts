import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import type { User } from '@/types/user'

export const useAuthStore = defineStore('auth', () => {
  // ---- State ----
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refreshToken'))

  // ---- Computed ----
  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  // ---- Actions ----
  async function login(username: string, password: string) {
    const { data } = await authApi.login({ username, password })
    token.value = data.token
    refreshToken.value = data.refreshToken
    localStorage.setItem('token', data.token)
    localStorage.setItem('refreshToken', data.refreshToken)
    user.value = {
      id: data.id,
      username: data.username,
      displayName: data.displayName,
      role: data.role ?? 'user',
      isHidden: false,
      createdAt: '',
    }
  }

  async function register(username: string, password: string, displayName: string) {
    const { data } = await authApi.register({
      username,
      password,
      displayName,
    })
    token.value = data.token
    refreshToken.value = data.refreshToken
    localStorage.setItem('token', data.token)
    localStorage.setItem('refreshToken', data.refreshToken)
    user.value = {
      id: data.id,
      username: data.username,
      displayName: data.displayName,
      role: 'user',
      isHidden: false,
      createdAt: '',
    }
  }

  function logout() {
    user.value = null
    token.value = null
    refreshToken.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('refreshToken')
  }

  async function refreshAuthToken() {
    if (!refreshToken.value) {
      logout()
      return
    }
    try {
      const { data } = await authApi.refreshToken(refreshToken.value)
      token.value = data.token
      refreshToken.value = data.refreshToken
      localStorage.setItem('token', data.token)
      localStorage.setItem('refreshToken', data.refreshToken)
    } catch {
      logout()
    }
  }

  return {
    user,
    token,
    refreshToken,
    isLoggedIn,
    isAdmin,
    login,
    register,
    logout,
    refreshAuthToken,
  }
})
