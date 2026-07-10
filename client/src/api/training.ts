import { request } from '../utils/request'

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
