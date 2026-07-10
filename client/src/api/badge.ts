import { request } from '../utils/request'

export function getAllBadges() {
  return request({ url: '/api/badges' })
}

export function getMyBadges() {
  return request({ url: '/api/badges/my' })
}
