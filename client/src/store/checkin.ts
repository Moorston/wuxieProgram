import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getCheckinList } from '../api'

export const useCheckinStore = defineStore('checkin', () => {
  const list = ref<any[]>([])
  const page = ref(1)
  const total = ref(0)
  const loading = ref(false)

  async function refresh() {
    page.value = 1
    loading.value = true
    try {
      const res: any = await getCheckinList(1)
      list.value = res.list || []
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  async function loadMore() {
    if (loading.value) return
    page.value++
    loading.value = true
    try {
      const res: any = await getCheckinList(page.value)
      list.value.push(...(res.list || []))
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  return { list, page, total, loading, refresh, loadMore }
})
