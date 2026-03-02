<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const activeTab = ref('login')
const loginLoading = ref(false)
const registerLoading = ref(false)

// ---- Login form ----
const loginFormRef = ref<FormInstance>()
const loginForm = reactive({
  username: '',
  password: '',
})

const loginRules = reactive<FormRules>({
  username: [
    { required: true, message: 'Please enter your username', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Please enter your password', trigger: 'blur' },
  ],
})

async function handleLogin() {
  const valid = await loginFormRef.value?.validate().catch(() => false)
  if (!valid) return

  loginLoading.value = true
  try {
    await authStore.login(loginForm.username, loginForm.password)
    ElMessage.success('Login successful')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (err: unknown) {
    const message =
      err instanceof Error ? err.message : 'Login failed. Please check your credentials.'
    ElMessage.error(message)
  } finally {
    loginLoading.value = false
  }
}

// ---- Register form ----
const registerFormRef = ref<FormInstance>()
const registerForm = reactive({
  username: '',
  display_name: '',
  password: '',
  confirm_password: '',
})

const validateConfirmPassword = (
  _rule: unknown,
  value: string,
  callback: (error?: Error) => void,
) => {
  if (value !== registerForm.password) {
    callback(new Error('Passwords do not match'))
  } else {
    callback()
  }
}

const registerRules = reactive<FormRules>({
  username: [
    { required: true, message: 'Please enter a username', trigger: 'blur' },
    { min: 3, max: 30, message: 'Username must be 3-30 characters', trigger: 'blur' },
  ],
  display_name: [
    { required: true, message: 'Please enter a display name', trigger: 'blur' },
  ],
  password: [
    { required: true, message: 'Please enter a password', trigger: 'blur' },
    { min: 6, message: 'Password must be at least 6 characters', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: 'Please confirm your password', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
})

async function handleRegister() {
  const valid = await registerFormRef.value?.validate().catch(() => false)
  if (!valid) return

  registerLoading.value = true
  try {
    await authStore.register(
      registerForm.username,
      registerForm.password,
      registerForm.display_name,
    )
    ElMessage.success('Registration successful')
    const redirect = (route.query.redirect as string) || '/'
    router.push(redirect)
  } catch (err: unknown) {
    const message =
      err instanceof Error ? err.message : 'Registration failed. Please try again.'
    ElMessage.error(message)
  } finally {
    registerLoading.value = false
  }
}
</script>

<template>
  <el-tabs v-model="activeTab" class="login-tabs" stretch>
    <!-- Login Tab -->
    <el-tab-pane label="Login" name="login">
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        label-position="top"
        class="mt-4"
        @submit.prevent="handleLogin"
      >
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="Enter your username"
            :prefix-icon="'User'"
          />
        </el-form-item>

        <el-form-item label="Password" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="Enter your password"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            class="w-full"
            :loading="loginLoading"
            @click="handleLogin"
          >
            Sign In
          </el-button>
        </el-form-item>
      </el-form>
    </el-tab-pane>

    <!-- Register Tab -->
    <el-tab-pane label="Register" name="register">
      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="registerRules"
        label-position="top"
        class="mt-4"
        @submit.prevent="handleRegister"
      >
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="Choose a username"
          />
        </el-form-item>

        <el-form-item label="Display Name" prop="display_name">
          <el-input
            v-model="registerForm.display_name"
            placeholder="Your display name"
          />
        </el-form-item>

        <el-form-item label="Password" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            placeholder="Create a password"
            show-password
          />
        </el-form-item>

        <el-form-item label="Confirm Password" prop="confirm_password">
          <el-input
            v-model="registerForm.confirm_password"
            type="password"
            placeholder="Confirm your password"
            show-password
            @keyup.enter="handleRegister"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            class="w-full"
            :loading="registerLoading"
            @click="handleRegister"
          >
            Create Account
          </el-button>
        </el-form-item>
      </el-form>
    </el-tab-pane>
  </el-tabs>
</template>

<style scoped>
.login-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
}
</style>
