const BASE_URL = process.env.VUE_APP_API_BASE_URL || 'http://localhost:8080'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: unknown
  header?: Record<string, string>
  skipAuth?: boolean // 跳过 token 注入（用于 refresh 等公开接口）
}

interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// Promise-based refresh queue to handle concurrent 401s
let refreshPromise: Promise<boolean> | null = null

function doRefresh(): Promise<boolean> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const refreshToken = uni.getStorageSync('refresh_token')
      if (!refreshToken) return false
      try {
        const res: any = await new Promise((resolve, rejectInner) => {
          uni.request({
            url: `${BASE_URL}/api/auth/refresh`,
            method: 'POST',
            data: { refresh_token: refreshToken },
            header: { 'Content-Type': 'application/json' },
            timeout: 10000,
            success: resolve,
            fail: rejectInner,
          })
        })
        const data = res.data as ApiResponse<{ token: string; refresh_token: string }>
        if (data.code === 0 && data.data?.token) {
          uni.setStorageSync('token', data.data.token)
          uni.setStorageSync('refresh_token', data.data.refresh_token)
          return true
        }
      } catch (error) {
        console.error('[request] token refresh failed:', error)
      }
      return false
    })()
      .finally(() => {
        refreshPromise = null
      })
  }
  return refreshPromise
}

async function handleUnauthorized(reject: (err: any) => void) {
  const refreshed = await doRefresh()
  if (refreshed) {
    uni.showToast({ title: '登录已刷新，请重试操作', icon: 'none', duration: 2000 })
    reject(new Error('token_refreshed'))
    return
  }

  // 清除所有 token 并跳转登录页
  uni.removeStorageSync('token')
  uni.removeStorageSync('refresh_token')
  uni.reLaunch({ url: '/pages/login/login' })
  reject(new Error('unauthorized'))
}

function doRequest<T>(
  url: string,
  options: RequestOptions,
): Promise<T> {
  return new Promise((resolve, reject) => {
    const header: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.header,
    }

    if (!options.skipAuth) {
      const token = uni.getStorageSync('token')
      if (token) {
        header['Authorization'] = `Bearer ${token}`
      }
    }

    uni.request({
      url,
      method: options.method || 'GET',
      data: options.data,
      header,
      timeout: 15000,
      success: (res: any) => {
        const data = res.data as ApiResponse<T>
        if (data.code === 0) {
          resolve(data.data)
        } else if (data.code === 401) {
          handleUnauthorized(reject)
        } else {
          uni.showToast({ title: data.message, icon: 'none' })
          reject(new Error(data.message))
        }
      },
      fail: (err: any) => {
        uni.showToast({ title: '网络错误', icon: 'none' })
        reject(err)
      },
    })
  })
}

export function request<T = any>(options: RequestOptions): Promise<T> {
  return doRequest<T>(`${BASE_URL}${options.url}`, options)
}

export const MEDIA_URL = process.env.VUE_APP_MEDIA_BASE_URL || 'http://localhost:8081'

export function mediaRequest<T = any>(options: RequestOptions): Promise<T> {
  return doRequest<T>(`${MEDIA_URL}${options.url}`, options)
}