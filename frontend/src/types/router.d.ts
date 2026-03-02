import 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    layout?: string
    requiresAuth?: boolean
    requiresAdmin?: boolean
    guestOnly?: boolean
  }
}
