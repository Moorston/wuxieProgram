import { defineStore } from 'pinia'
import { ref } from 'vue'
import { listResources } from '../api'
import type { Resource } from '../types/api'

export const useResourceStore = defineStore('resource', () => {
  const list = ref<Resource[]>([])
  const loading = ref(false)
  const page = ref(1)
  const total = ref(0)

  async function fetchList(params?: Record<string, string>) {
    loading.value = true
    page.value = 1
    try {
      const res: any = await listResources({ page: 1, pageSize: 20, ...params })
      list.value = res.list || []
      total.value = res.total || 0
    } finally {
      loading.value = false
    }
  }

  async function loadMore(params?: Record<string, string>) {
    if (loading.value) return
    page.value++
    loading.value = true
    try {
      const res: any = await listResources({ page: page.value, pageSize: 20, ...params })
      list.value.push(...(res.list || []))
    } finally {
      loading.value = false
    }
  }

  return { list, loading, total, fetchList, loadMore }
})