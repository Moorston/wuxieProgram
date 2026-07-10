import { request } from '../utils/request'

export function wxLogin(code: string, userInfo?: any) {
  return request({
    url: '/api/auth/login',
    method: 'POST',
    data: {
      code,
      nickname: userInfo?.nickName || '',
      avatar: userInfo?.avatarUrl || '',
      gender: userInfo?.gender || 0,
    },
  })
}

export function wxLogout() {
  return request({ url: '/api/auth/logout', method: 'POST' })
}

export function refreshToken(refreshToken: string) {
  return request({
    url: '/api/auth/refresh',
    method: 'POST',
    data: { refresh_token: refreshToken },
    skipAuth: true,
  })
}
