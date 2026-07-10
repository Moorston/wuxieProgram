import { request } from '../utils/request'

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

export function generateGroupInviteCode(id: string) {
  return request({ url: `/api/group/${id}/invite`, method: 'POST' })
}

export function joinGroupByInviteCode(code: string) {
  return request({ url: '/api/group/join', method: 'POST', data: { code } })
}

export function createGroupAnnouncement(data: { group_id: string; title: string; content: string; is_pinned?: boolean }) {
  return request({ url: '/api/group/announcements', method: 'POST', data })
}

export function getGroupAnnouncements(groupId: string, page = 1, pageSize = 20) {
  return request({ url: `/api/group/${groupId}/announcements?page=${page}&page_size=${pageSize}` })
}

export function deleteGroupAnnouncement(id: string) {
  return request({ url: `/api/group/announcements/${id}`, method: 'DELETE' })
}

export function removeGroupMember(groupId: string, userId: string) {
  return request({ url: `/api/group/${groupId}/remove-member`, method: 'POST', data: { user_id: userId } })
}

export function leaveGroup(groupId: string) {
  return request({ url: `/api/group/${groupId}/leave`, method: 'POST' })
}

export function setGroupLeader(groupId: string, userId: string) {
  return request({ url: `/api/group/${groupId}/set-leader`, method: 'POST', data: { user_id: userId } })
}
