const BASE_URL = 'http://localhost:8080'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  skipAuth?: boolean // 跳过 token 注入（用于 refresh 等公开接口）
}

interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

let isRedirecting = false

export function request<T = any>(options: RequestOptions): Promise<T> {
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
      url: `${BASE_URL}${options.url}`,
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

async function handleUnauthorized(reject: (err: any) => void) {
  if (isRedirecting) {
    reject(new Error('unauthorized'))
    return
  }
  isRedirecting = true

  // 尝试用 refresh token 刷新
  const refreshToken = uni.getStorageSync('refresh_token')
  if (refreshToken) {
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
        isRedirecting = false
        // TODO: 自动重试原请求（需要保存原请求参数）
        // 当前方案：提示用户手动重试
        uni.showToast({ title: '登录已刷新，请重试操作', icon: 'none', duration: 2000 })
        reject(new Error('token_refreshed'))
        return
      }
    } catch {
      // refresh 失败，走登出流程
    }
  }

  // 清除所有 token 并跳转登录页
  uni.removeStorageSync('token')
  uni.removeStorageSync('refresh_token')
  isRedirecting = false
  uni.reLaunch({ url: '/pages/login/login' })
  reject(new Error('unauthorized'))
}

export const MEDIA_URL = 'http://localhost:8081'

export function mediaRequest<T = any>(options: RequestOptions): Promise<T> {
  return new Promise((resolve, reject) => {
    const token = uni.getStorageSync('token')
    const header: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.header,
    }

    if (token) {
      header['Authorization'] = `Bearer ${token}`
    }

    uni.request({
      url: `${MEDIA_URL}${options.url}`,
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
          reject(new Error(data.message))
        }
      },
      fail: reject,
    })
  })
}
