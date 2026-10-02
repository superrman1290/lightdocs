import { ref } from 'vue'
import { apiClient, request } from '../services/apiClient'

export const authenticated = ref(Boolean(apiClient.getAccessToken() || sessionStorage.getItem('lightdocs-authenticated') === 'true'))

export const login = async (username: string, password: string, rememberMe = false) => {
  apiClient.clearAccessToken()
  sessionStorage.removeItem('lightdocs-authenticated')
  const result = await request<{ accessToken: string }>('/auth/session', { method: 'POST', body: JSON.stringify({ username, password, rememberMe }) })
  if (!result?.accessToken) {
    throw new Error('登录响应缺少 access token')
  }
  apiClient.setAccessToken(result.accessToken)
  sessionStorage.setItem('lightdocs-authenticated', 'true')
  authenticated.value = true
}

export const logout = async () => {
  try { await request('/auth/session', { method: 'DELETE' }) } finally {
    apiClient.clearAccessToken()
    sessionStorage.removeItem('lightdocs-authenticated')
    authenticated.value = false
  }
}
