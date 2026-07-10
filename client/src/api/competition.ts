import { request } from '../utils/request'

export function getCompetitions() {
  return request({ url: '/api/competitions' })
}

export function getCompetitionDetail(id: string) {
  return request({ url: `/api/competitions/${id}` })
}

export function submitCompetitionEntry(competitionId: string, checkinId: string) {
  return request({
    url: `/api/competitions/${competitionId}/submit`,
    method: 'POST',
    data: { checkin_id: checkinId },
  })
}

export function getCompetitionEntries(id: string, page = 1, pageSize = 20) {
  return request({ url: `/api/competitions/${id}/entries?page=${page}&page_size=${pageSize}` })
}

export function getCompetitionRanking(id: string) {
  return request({ url: `/api/competitions/${id}/ranking` })
}

export function scoreEntry(competitionId: string, entryId: string, score: number) {
  return request({
    url: `/api/competitions/${competitionId}/entries/${entryId}/score`,
    method: 'POST',
    data: { score },
  })
}
