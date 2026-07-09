# 状态管理规范

> 武俱打卡项目前端状态管理的约定。

---

## 概览

本项目使用 **Pinia** 作为全局状态管理方案。全局状态按模块拆分到独立的 store 文件，服务于跨页面共享的数据。

---

## 状态分类

| 类别 | 存储方式 | 例子 |
|------|---------|------|
| **全局用户状态** | Pinia store | Token、用户信息、登录状态 |
| **全局列表状态** | Pinia store | 打卡列表的缓存数据 |
| **本地组件状态** | `ref()` / `reactive()` | 分页参数、加载状态、表单数据 |
| **持久化状态** | `uni.getStorageSync()` | Token（持久化到本地） |
| **页面参数** | URL query | `id`, `page` |

---

## Store 结构

### store/user.ts — 用户状态

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const userInfo = ref<any>(null)

  // 持久化到本地存储
  function loadFromStorage() {
    token.value = uni.getStorageSync('token') || ''
    const info = uni.getStorageSync('userInfo')
    if (info) userInfo.value = JSON.parse(info)
  }

  // 保存到本地存储
  function save(tokenStr: string, info: any) {
    token.value = tokenStr
    userInfo.value = info
    uni.setStorageSync('token', tokenStr)
    uni.setStorageSync('userInfo', JSON.stringify(info))
  }

  // 退出登录
  function logout() {
    token.value = ''
    userInfo.value = null
    uni.removeStorageSync('token')
    uni.removeStorageSync('userInfo')
    uni.reLaunch({ url: '/pages/login/login' })
  }

  // 是否已登录
  const isLoggedIn = computed(() => !!token.value)

  return { token, userInfo, isLoggedIn, loadFromStorage, save, logout }
})
```

### store/checkin.ts — 打卡状态

```typescript
export const useCheckinStore = defineStore('checkin', () => {
  const list = ref<any[]>([])
  const page = ref(1)
  const hasMore = ref(true)

  function append(items: any[]) {
    list.value.push(...items)
    page.value++
    hasMore.value = items.length > 0
  }

  function reset() {
    list.value = []
    page.value = 1
    hasMore.value = true
  }

  return { list, page, hasMore, append, reset }
})
```

---

## 何时使用全局状态 vs 本地状态

### 使用 Pinia 全局状态

- ✅ 用户登录状态（跨页面共享）
- ✅ 用户信息（多页面需要）
- ✅ 需要持久化的数据（Token）
- ✅ 跨页面的列表缓存（打卡列表）

### 使用本地 `ref()` 状态

- ✅ 页面内的加载状态（`loading`, `refreshing`）
- ✅ 分页参数（`page`, `pageSize`）
- ✅ 表单数据
- ✅ 临时 UI 状态（弹窗显示、Tab 切换）

---

## 数据流

```
API Response → request() 封装 → 解析响应(code 检查)
                            ↓
                    Pinia Store (全局)
                            ↓
                    页面组件 (ref)
                            ↓
                    Template 渲染
```

### 认证流程

1. `request.ts` 拦截器从 store 获取 token
2. 添加到 `Authorization: Bearer <token>` 请求头
3. 收到 401 → 清除本地 token → reLaunch 到登录页

---

## 常见错误

1. **将不需要共享的状态放入 Pinia**
   - **问题**：局部加载状态、临时表单数据放入全局 store，导致不必要的耦合
   - **修复**：仅当状态需要在**多个不相关的组件/页面间共享**时才使用 Pinia

2. **状态持久化策略不一致**
   - **问题**：部分状态手动 `setStorageSync`，部分自动
   - **修复**：明确持久化边界 — 仅 `token` 和 `userInfo` 需要持久化

3. **直接在组件中修改 store 状态**
   - **问题**：组件中 `store.userInfo = {...}` 绕过 store action
   - **修复**：通过 store 定义的 action 修改状态

4. **未处理 store 重置**
   - **问题**：退出登录后 store 状态未清空
   - **修复**：`logout()` action 应重置所有相关 store