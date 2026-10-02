const baseURL = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080/api/v1'
export const apiBaseURL = baseURL

export const resolveApiURL = (value: string) => {
  if (!value || /^https?:\/\//i.test(value) || value.startsWith('blob:') || value.startsWith('data:')) {
    return value
  }

  return new URL(value, baseURL).toString()
}

export class ApiError extends Error {
  status: number
  code: number

  constructor(message: string, status = 500, code = 50000) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

const tokenKey = 'lightdocs-access-token'

export const getAccessToken = () => sessionStorage.getItem(tokenKey)
export const setAccessToken = (token: string) => sessionStorage.setItem(tokenKey, token)
export const clearAccessToken = () => sessionStorage.removeItem(tokenKey)

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Accept', 'application/json')
  if (init.body && !(init.body instanceof FormData)) headers.set('Content-Type', 'application/json')
  const token = getAccessToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const response = await fetch(`${baseURL}${path}`, { ...init, headers })
  if (response.status === 204) return undefined as T
  const payload = await response.json().catch(() => null)
  if (!response.ok || payload?.code) {
    if (response.status === 401) {
      clearAccessToken()
      sessionStorage.removeItem('lightdocs-authenticated')
      if (location.pathname.startsWith('/admin')) {
        location.href = `/login?redirect=${encodeURIComponent(location.pathname + location.search)}`
      }
    }
    throw new ApiError(payload?.message || response.statusText, response.status, payload?.code)
  }
  return payload?.data as T
}

export const apiClient = { request, getAccessToken, setAccessToken, clearAccessToken }
