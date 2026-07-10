import { request } from '../utils/request'

export function getCheckinHeatmap(months = 6) {
  return request({ url: `/api/analytics/checkin-heatmap?months=${months}` })
}

export function getCheckinTrend(days = 30) {
  return request({ url: `/api/analytics/checkin-trend?days=${days}` })
}

export function getAnalyticsOverview() {
  return request({ url: '/api/analytics/overview' })
}
