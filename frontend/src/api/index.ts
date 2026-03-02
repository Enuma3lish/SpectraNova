import axios from 'axios'

/**
 * Axios instance configured for the FenzVideo API.
 *
 * - baseURL defaults to '/api/v1' (handled by Vite proxy in dev),
 *   but can be overridden via the VITE_API_BASE_URL env variable.
 * - Automatically attaches the JWT Bearer token from localStorage.
 * - On 401 responses, clears stored credentials and redirects to /login.
 */
export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

// ---- Request interceptor: attach JWT token ----
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// ---- Response interceptor: handle 401 ----
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')

      // Only redirect if we are not already on the login page
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export default apiClient
