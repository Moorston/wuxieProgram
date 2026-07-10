import { request } from '../utils/request'

export function followUser(id: string) {
  return request({ url: `/api/follow/${id}`, method: 'POST' })
}

export function unfollowUser(id: string) {
  return request({ url: `/api/follow/${id}`, method: 'DELETE' })
}

export function getFollowing(page = 1, pageSize = 20) {
  return request({ url: `/api/follow/following?page=${page}&page_size=${pageSize}` })
}

export function getFollowers(page = 1, pageSize = 20) {
  return request({ url: `/api/follow/followers?page=${page}&page_size=${pageSize}` })
}

export function getFeed(page = 1, pageSize = 20) {
  return request({ url: `/api/feed?page=${page}&page_size=${pageSize}` })
}

export function getUserProfile(id: string) {
  return request({ url: `/api/user/${id}/profile` })
}
