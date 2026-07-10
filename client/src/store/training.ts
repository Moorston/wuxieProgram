import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listTrainingPlans, getTodayTasks } from '../api'
import type { TrainingPlan } from '../types/api'

export const useTrainingStore = defineStore('training', () => {
  const plans = ref<TrainingPlan[]>([])
  const todayTasks = ref<TrainingPlan[]>([])
  const loading = ref(false)
  const page = ref(1)
  const total = ref(0)

  async function fetchPlans(status?: string) {
    loading.value = true
    page.value = 1
    try {
      const res: any = await listTrainingPlans(1, 10, status)
      plans.value = res.list || []
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  async function loadMorePlans(status?: string) {
    if (loading.value) return
    page.value++
    loading.value = true
    try {
      const res: any = await listTrainingPlans(page.value, 10, status)
      plans.value.push(...(res.list || []))
    } finally {
      loading.value = false
    }
  }

  async function fetchTodayTasks() {
    try {
      todayTasks.value = await getTodayTasks() as any
    } catch {
      todayTasks.value = []
    }
  }

  return { plans, todayTasks, loading, total, fetchPlans, loadMorePlans, fetchTodayTasks }
})