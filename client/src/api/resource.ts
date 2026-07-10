import { request } from '../utils/request'

export function getResourcePresign(ext: string) {
  return request({ url: `/api/resource/upload/presign?ext=${encodeURIComponent(ext)}` })
}

export function resourceUploadCallback(data: any) {
  return request({ url: '/api/resource/upload/callback', method: 'POST', data })
}

export function createResource(data: any) {
  return request({ url: '/api/resource', method: 'POST', data })
}

export function listResources(params: any) {
  let url = `/api/resource/list?page=${params.page || 1}&page_size=${params.pageSize || 20}`
  if (params.type) url += `&type=${params.type}`
  if (params.category) url += `&category=${params.category}`
  if (params.difficulty) url += `&difficulty=${params.difficulty}`
  if (params.tag) url += `&tag=${encodeURIComponent(params.tag)}`
  if (params.keyword) url += `&keyword=${encodeURIComponent(params.keyword)}`
  if (params.scope) url += `&scope=${params.scope}`
  if (params.sort) url += `&sort=${params.sort}`
  if (params.groupId) url += `&group_id=${params.groupId}`
  return request({ url })
}

export function getResource(id: string) {
  return request({ url: `/api/resource/${id}` })
}

export function updateResource(id: string, data: any) {
  return request({ url: `/api/resource/${id}`, method: 'PUT', data })
}

export function deleteResource(id: string) {
  return request({ url: `/api/resource/${id}`, method: 'DELETE' })
}

export function toggleResourceFavorite(id: string) {
  return request({ url: `/api/resource/${id}/favorite`, method: 'POST' })
}

export function listResourceFavorites(page = 1, pageSize = 20) {
  return request({ url: `/api/resource/favorites?page=${page}&page_size=${pageSize}` })
}

export function getResourceTags() {
  return request({ url: '/api/resource/tags' })
}

export function getResourceStats() {
  return request({ url: '/api/resource/stats' })
}
