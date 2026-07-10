import { request } from '../utils/request'

export function createInsight(data: any) {
  return request({ url: '/api/insight', method: 'POST', data })
}

export function getInsight(id: string) {
  return request({ url: `/api/insight/${id}` })
}

export function listInsights(page = 1, pageSize = 20, tag?: string, mood?: string) {
  let url = `/api/insight/list?page=${page}&page_size=${pageSize}`
  if (tag) url += `&tag=${encodeURIComponent(tag)}`
  if (mood) url += `&mood=${mood}`
  return request({ url })
}

export function listPublicInsights(page = 1, pageSize = 20) {
  return request({ url: `/api/insight/public?page=${page}&page_size=${pageSize}` })
}

export function updateInsight(id: string, data: any) {
  return request({ url: `/api/insight/${id}`, method: 'PUT', data })
}

export function deleteInsight(id: string) {
  return request({ url: `/api/insight/${id}`, method: 'DELETE' })
}

export function getInsightTags() {
  return request({ url: '/api/insight/tags' })
}

export function getMoodStats(days = 30) {
  return request({ url: `/api/insight/mood-stats?days=${days}` })
}

export function getOnThisDay() {
  return request({ url: '/api/insight/on-this-day' })
}

export function likeInsight(id: string) {
  return request({ url: `/api/insight/${id}/like`, method: 'POST' })
}
