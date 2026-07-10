import { request } from '../utils/request'

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
