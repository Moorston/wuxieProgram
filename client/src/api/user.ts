import { request } from '../utils/request'

export function getProfile() {
  return request({ url: '/api/user/profile' })
}

export function updateProfile(data: { nickname?: string; avatar?: string }) {
  return request({ url: '/api/user/profile', method: 'PUT', data })
}
