# Hook 规范

> 武俱打卡项目前端的逻辑复用约定。

---

## 概览

本项目使用 Vue3 组合式 API，但**尚未创建自定义 composables/hooks**。当前的数据获取逻辑直接写在页面组件中。本节定义应遵循的 hooks 模式。

---

## 自定义 Composable 模式

### 基本结构

```typescript
// composables/usePagination.ts
import { ref, type Ref } from 'vue'

interface PaginationResult<T> {
  list: Ref<T[]>
  loading: Ref<boolean>
  hasMore: Ref<boolean>
  loadMore: () => Promise<void>
  refresh: () => Promise<void>
}

export function usePagination<T>(
  fetchFn: (page: number, pageSize: number) => Promise<{ list: T[]; total: number }>,
  pageSize = 20
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
      hasMore.value = res.list.length === pageSize
      page.value++
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    page.value = 1
    list.value = []
    hasMore.value = true
    await loadMore()
  }

  return { list, loading, hasMore, loadMore, refresh }
}
```

### 使用方法

```vue
<script setup lang="ts">
import { onMounted } from 'vue'
import { usePagination } from '@/composables/usePagination'
import { getCheckinList } from '@/api'

const { list, loading, hasMore, loadMore, refresh } = usePagination(
  (page, size) => getCheckinList(page, size, groupId),
  20
)

onMounted(() => loadMore())
</script>
```

---

## 数据获取模式

### 当前模式（直接写在页面中）

当前项目中分页逻辑在每个页面重复实现，应提取到 composable：

```vue
// ❌ 当前项目中每个页面重复的代码
const list = ref<any[]>([])
const page = ref(1)
const loading = ref(false)

async function loadMore() {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getXxxList(page.value, 20)
    list.value.push(...res.list)
    page.value++
  } finally {
    loading.value = false
  }
}
```

### 推荐模式

```vue
// ✅ 推荐：使用 composable 复用分页逻辑
const { list, loading, loadMore, refresh } = usePagination(getCheckinList)
```

---

## 命名约定

| 类型 | 约定 | 示例 |
|------|------|------|
| Composables 文件 | `use{功能}.ts` | `usePagination.ts`, `useUser.ts` |
| Composables 函数 | `use{功能}` | `usePagination()`, `useUserAuth()` |
| Composables 返回值 | 解构为语义变量 | `const { list, loading } = usePagination()` |

---

## 建议提取的 Composables

基于当前代码分析，建议提取以下 composables：

| Composable | 功能 | 使用页面 |
|-----------|------|---------|
| `usePagination` | 通用分页加载 | square, my-video, insight/list, notification/list 等 |
| `useUserAuth` | 用户认证检查 + 获取用户信息 | 所有需登录页面 |
| `useVideoPlayer` | 视频播放状态 + 进度管理 | video-detail, resource/detail |
| `useDebounce` | 输入防抖 | square 搜索 |
| `usePullRefresh` | 下拉刷新 + 上拉加载 | 所有列表页面 |

---

## 常见错误

1. **Composable 中直接修改组件状态**
   - **问题**：composable 中修改组件外部的 `ref()`
   - **修复**：composable 应返回响应式状态，由组件决定如何使用

2. **Composable 内调用 UI 方法**
   - **问题**：composable 中调用 `uni.showToast()` 等 UI 方法，降低复用性
   - **修复**：返回状态让组件处理 UI，或接受回调函数作为参数

3. **Composable 不处理清理逻辑**
   - **问题**：异步操作在组件销毁后继续执行
   - **修复**：使用 `onUnmounted` 清理定时器/取消请求