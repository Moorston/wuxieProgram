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
