import { ref, type Ref } from 'vue'

interface PaginationResult<T> {
  list: Ref<T[]>
  loading: Ref<boolean>
  hasMore: Ref<boolean>
  loadMore: () => Promise<void>
  refresh: () => Promise<void>
}

/**
 * 通用分页加载 composable
 * @param fetchFn 分页获取数据的函数，接收 (page, pageSize) 返回 { list, total }
 * @param pageSize 每页大小，默认 20
 */
export function usePagination<T>(
  fetchFn: (page: number, pageSize: number) => Promise<{ list: T[]; total: number }>,
  pageSize = 20,
): PaginationResult<T> {
  const list = ref<T[]>([]) as Ref<T[]>
  const loading = ref(false)
  const hasMore = ref(true)
  const page = ref(1)

  async function loadMore() {
    if (loading.value || !hasMore.value) return
    loading.value = true
    try {
      const res = await fetchFn(page.value, pageSize)
      list.value.push(...res.list)
      hasMore.value = res.list.length >= pageSize
      page.value++
    } catch (e) {
      console.error('[usePagination] loadMore failed:', e)
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    page.value = 1
    list.value = []
    hasMore.value = true
    loading.value = false
    await loadMore()
  }

  return { list, loading, hasMore, loadMore, refresh }
}