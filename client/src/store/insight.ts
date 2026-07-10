import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listInsights, getUnreadCount } from '../api'
import type { Insight } from '../types/api'

export const useInsightStore = defineStore('insight', () => {
  const list = ref<Insight[]>([])
  const loading = ref(false)
  const page = ref(1)
  const total = ref(0)

  async function fetchList(tag?: string, mood?: string) {
    loading.value = true
    page.value = 1
    try {
      const res: any = await listInsights(1, 20, tag, mood)
      list.value = res.list || []
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  async function loadMore(tag?: string, mood?: string) {
    if (loading.value) return
    page.value++
    loading.value = true
    try {
      const res: any = await listInsights(page.value, 20, tag, mood)
      list.value.push(...(res.list || []))
    } finally {
      loading.value = false
    }
  }

  return { list, loading, total, fetchList, loadMore }
})