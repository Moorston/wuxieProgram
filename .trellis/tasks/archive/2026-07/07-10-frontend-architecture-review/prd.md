# 前端架构审查

## Goal

审查 UniApp 前端架构，找出摩擦点并提出改进方案。

## Frontend Overview

### 技术栈
- UniApp (Vue 3 + Pinia + TypeScript)
- SSR (createSSRApp)
- 5 个 Tab 页：首页、广场、打卡、排位、我的
- 34 个页面，~10 个领域模块

### 现有架构

```
client/src/
├── api/           # 按领域拆分的 API 层 (8 文件) ✅
├── store/         # Pinia 状态管理 (2 文件)
├── composables/   # 组合式函数 (1 文件)
├── constants/     # 常量映射 (1 文件)
├── utils/         # 工具函数 (1 文件)
├── pages/         # 34 个页面，按领域组织
├── App.vue
├── main.ts
├── pages.json     # 路由配置
└── manifest.json
```

## Confirmed Facts

1. **API 层已按领域拆分** — 8 个文件 (auth, user, checkin, rank, training, notification, insight, resource) ✅
2. **request.ts 支持 token 自动注入 + refresh 机制** ✅
3. **constants/index.ts** 存在常量映射表 ✅
4. **Pinia store 只有 2 个** (`user.ts`, `checkin.ts`) ❌ — 34 页面对应 2 个 store
5. **大量 `any` 类型** — API 返回值、组件 props 无类型定义 ❌
6. **无路由守卫** — 401 时通过 `uni.reLaunch` 跳转，非导航守卫 ❌
7. **只有 1 个 composable** (`usePagination.ts`) ❌
8. **无共享组件** — 每个页面独立实现 UI 模式 ❌

## Architecture Friction Points

### F1: Store 覆盖不足（High）
- 只有 `user.ts` 和 `checkin.ts` 两个 store
- 训练、感悟、资源、通知等模块使用本地状态（页面内 `ref`）
- 跨页面无法共享状态（如修改通知设置后回到列表页不刷新）

### F2: TypeScript 类型缺失（High）
- API 响应用 `any` 类型（`request<T = any>`）
- 页面组件内数据用 `ref<any>` 或 `const res: any`
- 没有 API 响应类型定义
- 无法获得 IDE 自动补全和编译时类型检查

### F3: 无共享组件库（Medium）
- 打卡列表、训练列表、感悟列表等 UI 模式重复
- 下拉刷新、分页加载、空状态等通用模式未抽象
- 缺少统一的加载状态 / 错误状态组件

### F4: 缺少 composable（Low）
- 只有 `usePagination`
- 缺少 `usePullRefresh`、`useToast`、`useAuth` 等常见 composable
- 每个页面重复 try/catch 错误处理逻辑

## Recommendations

### R1: 扩展 Store 覆盖
- 新增 `training.ts`, `insight.ts`, `resource.ts`, `notification.ts` store
- 每个 store 管理列表 + 分页 + loading 状态

### R2: 定义 API 类型
- 在 `types/` 目录定义 API 响应类型
- `request<T>` 使用具体类型替代 `any`
- 为模型添加接口：`User`, `Checkin`, `TrainingPlan` 等

### R3: 共享组件
- 基础组件：`PageStatus`（加载/空/错误）、`PullRefresh`、`Pagination`
- 领域组件：`CheckinCard`, `InsightCard`, `ResourceCard`

### R4: 路由守卫
- 导航守卫检查登录状态（非 401 时提前拦截）
- `isLoggedIn` 状态从 Pinia store 读取

## Out of Scope

- 更换 UI 框架（如 uView、uni-ui）
- 重构页面 UI 样式
- 性能优化（分包加载、懒加载）

## Open Questions

1. 哪些改进点优先实施？