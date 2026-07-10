import { request } from '../utils/request'

export function createChallenge(data: { title: string; description?: string; duration: number }) {
  return request({ url: '/api/challenges', method: 'POST', data })
}

export function getChallenges(page = 1, pageSize = 20) {
  return request({ url: `/api/challenges?page=${page}&page_size=${pageSize}` })
}

export function getChallengeDetail(id: string) {
  return request({ url: `/api/challenges/${id}` })
}

export function joinChallenge(id: string) {
  return request({ url: `/api/challenges/${id}/join`, method: 'POST' })
}

export function getChallengeRanking(id: string) {
  return request({ url: `/api/challenges/${id}/ranking` })
}
