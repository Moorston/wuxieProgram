import { request } from '../utils/request'

export function prepareCheckin(description: string) {
  return request({ url: '/api/checkin/prepare', method: 'POST', data: { description } })
}

export function getCheckinList(page = 1, pageSize = 10, groupId?: string) {
  let url = `/api/checkin/list?page=${page}&page_size=${pageSize}`
  if (groupId) url += `&group_id=${groupId}`
  return request({ url })
}

export function getCheckinDetail(id: string) {
  return request({ url: `/api/checkin/${id}` })
}

export function searchCheckinList(keyword: string, page = 1, pageSize = 10) {
  return request({ url: `/api/checkin/search?q=${encodeURIComponent(keyword)}&page=${page}&page_size=${pageSize}` })
}

export function getMyCheckins(page = 1, pageSize = 10) {
  return request({ url: `/api/checkin/mine?page=${page}&page_size=${pageSize}` })
}

export function deleteCheckin(id: string) {
  return request({ url: `/api/checkin/${id}`, method: 'DELETE' })
}

export function toggleLike(id: string) {
  return request({ url: `/api/checkin/${id}/like`, method: 'POST' })
}

export function addComment(id: string, content: string) {
  return request({ url: `/api/checkin/${id}/comment`, method: 'POST', data: { content } })
}

export function getComments(id: string, page = 1, pageSize = 20) {
  return request({ url: `/api/checkin/${id}/comments?page=${page}&page_size=${pageSize}` })
}
