import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getNotificationList, getUnreadCount } from '../api'
import type { Notification } from '../types/api'

export const useNotificationStore = defineStore('notification', () => {
  const list = ref<Notification[]>([])
  const unreadCount = ref(0)
  const loading = ref(false)
  const page = ref(1)
  const total = ref(0)

  async function fetchList() {
    loading.value = true
    page.value = 1
    try {
      const res: any = await getNotificationList(1, 20)
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
      const res: any = await getNotificationList(page.value, 20)
      list.value.push(...(res.list || []))
    } finally {
      loading.value = false
    }
  }

  async function fetchUnreadCount() {
    try {
      unreadCount.value = await getUnreadCount() as any
    } catch {
      unreadCount.value = 0
    }
  }

  return { list, unreadCount, loading, total, fetchList, loadMore, fetchUnreadCount }
})