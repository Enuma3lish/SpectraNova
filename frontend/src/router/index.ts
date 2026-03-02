import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/LoginView.vue'),
    meta: { guestOnly: true, layout: 'auth' },
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/HomeView.vue'),
  },
  {
    path: '/search',
    name: 'Search',
    component: () => import('@/views/SearchResultsView.vue'),
  },
  {
    path: '/category/:id',
    name: 'Category',
    component: () => import('@/views/CategoryView.vue'),
  },
  {
    path: '/channel/:id',
    name: 'Channel',
    component: () => import('@/views/ChannelView.vue'),
  },
  {
    path: '/video/:id',
    name: 'Video',
    component: () => import('@/views/VideoView.vue'),
  },
  {
    path: '/admin',
    name: 'Admin',
    component: () => import('@/views/admin/AdminView.vue'),
    meta: { requiresAuth: true, requiresAdmin: true, layout: 'admin' },
    children: [
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/AdminUsersView.vue'),
      },
      {
        path: 'tags',
        name: 'AdminTags',
        component: () => import('@/views/admin/AdminTagsView.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// ---- Navigation guards ----
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  // Guest-only routes (e.g. login): redirect to home if already logged in
  if (to.meta.guestOnly && authStore.isLoggedIn) {
    return next({ name: 'Home' })
  }

  // Auth-required routes: redirect to login if not authenticated
  if (to.meta.requiresAuth && !authStore.isLoggedIn) {
    return next({ name: 'Login', query: { redirect: to.fullPath } })
  }

  // Admin-required routes: redirect to home if not admin
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return next({ name: 'Home' })
  }

  next()
})

export default router
