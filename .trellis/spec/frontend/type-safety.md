# 类型安全规范

> 武俱打卡项目前端的 TypeScript 类型使用约定。

---

## 概览

项目全量使用 TypeScript 编写。前端 API 请求有统一的泛型 `ApiResponse<T>` 封装，但**类型定义不够严格**，多处使用 `any`。

---

## API 类型系统

### 基础响应类型

```typescript
// utils/request.ts
interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}
```

### 建议的 API 类型定义

应定义每个 API 响应的类型，替换 `any`：

```typescript
// ✅ 推荐：为每个模块定义类型
// types/checkin.ts
export interface Checkin {
  _id: string
  user_id: string
  video_url: string
  cover_url: string
  description: string
  status: number
  like_count: number
  comment_count: number
  score: number
  duration: number
  created_at: string
  user?: User
  is_liked?: boolean
}

export interface CheckinListResponse {
  list: Checkin[]
  total: number
  page: number
  page_size: number
}

// ❌ 当前项目中大量使用 any
const list = ref<any[]>([])
```

---

## 类型组织

当前项目中类型定义分散，建议集中管理：

```
client/
└── src/
    └── types/              # 建议新建的类型目录
        ├── api.ts          # API 通用类型（ApiResponse, PaginationParams）
        ├── user.ts         # 用户相关类型
        ├── checkin.ts      # 打卡相关类型
        ├── training.ts     # 训练计划相关类型
        ├── insight.ts      # 感悟笔记相关类型
        ├── notification.ts # 通知相关类型
        └── resource.ts     # 资料库相关类型
```

---

## 禁止模式

### ❌ 滥用 `any`

```typescript
// ❌ 禁止
const list = ref<any[]>([])
const userInfo = ref<any>(null)
const data = await request({ url: '/api/xxx' })

// ✅ 正确
interface Checkin { ... }
const list = ref<Checkin[]>([])
const userInfo = ref<User | null>(null)
const data = await request<CheckinListResponse>({ url: '/api/xxx' })
```

### ❌ `as` 类型断言绕过类型检查

```typescript
// ❌ 禁止
const res = await request({ url: '/api/xxx' })
const data = res as Checkin[]  // 绕过类型检查

// ✅ 正确
const res = await request<Checkin[]>({ url: '/api/xxx' })
```

### ❌ 未类型化的对象

```typescript
// ❌ 禁止
const plan = { id: '123', name: 'test' }  // 类型推断为 { id: string; name: string }

// ✅ 正确
interface TrainingPlan { id: string; name: string }
const plan: TrainingPlan = { id: '123', name: 'test' }
```

---

## 必需模式

### ✅ 为 API 封装提供完整类型

```typescript
// api/index.ts — 使用泛型指定返回类型
export function getCheckinList(page = 1, pageSize = 10) {
  return request<CheckinListResponse>({
    url: `/api/checkin/list?page=${page}&page_size=${pageSize}`
  })
}
```

### ✅ 使用 `as const` 联合类型

```typescript
// ✅ 正确：联合类型替代魔法字符串
export type CheckinStatus = 0 | 1 | 2 | 3
export type PlanStatus = 'active' | 'completed' | 'draft' | 'cancelled'
export type TaskType = 'basic' | 'taolu' | 'sanda' | 'qigong'
```

---

## 项目现状

- ✅ 全部使用 TypeScript 编写
- ✅ 有基础的泛型 `ApiResponse<T>` 封装
- ❌ **几乎所有页面的 `ref()` 都使用 `any`**（需要类型化）
- ❌ 没有独立的类型定义文件
- ❌ API 调用未指定泛型参数
- ❌ 后端模型和前端类型不同步（手动维护）