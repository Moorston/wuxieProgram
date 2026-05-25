import { request } from './request'

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

export function getCheckinList(page = 1, pageSize = 10) {
  return request({ url: `/api/checkin/list?page=${page}&page_size=${pageSize}` })
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
