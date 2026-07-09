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

export function getProfile() {
  return request({ url: '/api/user/profile' })
}

export function updateProfile(data: { nickname?: string; avatar?: string }) {
  return request({ url: '/api/user/profile', method: 'PUT', data })
}

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

export function getRankList(period = 'all', page = 1, pageSize = 20) {
  return request({ url: `/api/rank?period=${period}&page=${page}&page_size=${pageSize}` })
}

export function getMyRank(period = 'all') {
  return request({ url: `/api/rank/me?period=${period}` })
}

export function getGroupList() {
  return request({ url: '/api/group/list' })
}

export function getGroupDetail(id: string) {
  return request({ url: `/api/group/${id}` })
}

export function createTrainingPlan(data: any) {
  return request({ url: '/api/training/plan', method: 'POST', data })
}

export function getTrainingPlan(id: string) {
  return request({ url: `/api/training/plan/${id}` })
}

export function listTrainingPlans(page = 1, pageSize = 10, status?: string) {
  let url = `/api/training/plans?page=${page}&page_size=${pageSize}`
  if (status !== undefined) url += `&status=${status}`
  return request({ url })
}

export function updateTrainingPlan(id: string, data: any) {
  return request({ url: `/api/training/plan/${id}`, method: 'PUT', data })
}

export function deleteTrainingPlan(id: string) {
  return request({ url: `/api/training/plan/${id}`, method: 'DELETE' })
}

export function getTodayTasks() {
  return request({ url: '/api/training/today' })
}

export function updateTaskStatus(planId: string, day: number, taskIdx: number, status: number, checkinId?: string) {
  return request({
    url: `/api/training/task/${planId}/${day}/${taskIdx}`,
    method: 'PUT',
    data: { status, checkin_id: checkinId },
  })
}

export function getTrainingReport(id: string) {
  return request({ url: `/api/training/plan/${id}/report` })
}

export function listTemplates(page = 1, pageSize = 20, category?: string, style?: string) {
  let url = `/api/training/template/list?page=${page}&page_size=${pageSize}`
  if (category) url += `&category=${category}`
  if (style) url += `&style=${style}`
  return request({ url })
}

export function getTemplate(id: string) {
  return request({ url: `/api/training/template/${id}` })
}

export function applyTemplate(id: string, startDate: string) {
  return request({ url: `/api/training/template/${id}/apply`, method: 'POST', data: { start_date: startDate } })
}

export function getNotificationList(page = 1, pageSize = 20) {
  return request({ url: `/api/notification/list?page=${page}&page_size=${pageSize}` })
}

export function getUnreadCount() {
  return request({ url: '/api/notification/unread' })
}

export function markNotificationRead(id: string) {
  return request({ url: `/api/notification/read/${id}`, method: 'PUT' })
}

export function markAllNotificationsRead() {
  return request({ url: '/api/notification/read-all', method: 'PUT' })
}

export function deleteNotification(id: string) {
  return request({ url: `/api/notification/${id}`, method: 'DELETE' })
}

export function getNotificationSettings() {
  return request({ url: '/api/notification/settings' })
}

export function updateNotificationSettings(data: any) {
  return request({ url: '/api/notification/settings', method: 'PUT', data })
}

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
