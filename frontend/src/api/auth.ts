import { apiClient } from './index'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  displayName: string
}

export interface AuthResponse {
  id: number
  username: string
  displayName: string
  role?: string
  token: string
  refreshToken: string
}

export interface RefreshResponse {
  token: string
  refreshToken: string
}

export function login(data: LoginRequest) {
  return apiClient.post<AuthResponse>('/auth/login', data)
}

export function register(data: RegisterRequest) {
  return apiClient.post<AuthResponse>('/auth/register', data)
}

export function refreshToken(refreshToken: string) {
  return apiClient.post<RefreshResponse>('/auth/refresh', { refreshToken })
}
