import { request } from '../utils/request'

export function getProfile() {
  return request({ url: '/api/user/profile' })
}

export function updateProfile(data: { nickname?: string; avatar?: string }) {
  return request({ url: '/api/user/profile', method: 'PUT', data })
}

export function getPrivacySettings() {
  return request({ url: '/api/user/privacy' })
}

export function updatePrivacySettings(visibility: number) {
  return request({ url: '/api/user/privacy', method: 'PUT', data: { visibility } })
}

export function getUserLevel() {
  return request({ url: '/api/user/level' })
}

export function exportMyCheckins() {
  return request({ url: '/api/checkin/export' })
}
