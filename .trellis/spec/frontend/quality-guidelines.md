# 前端代码质量标准

> 武俱打卡项目前端的代码质量约束和禁止模式。

---

## 概览

前端使用 **Vue3 + TypeScript + uni-app**。质量标准基于项目实际代码分析和审查报告提炼。

---

## 禁止模式

### 🔴 P0 - 功能正确性

| # | 禁止模式 | 风险 | 真实案例 |
|---|---------|------|---------|
| 1 | **URL 模板字符串拼接错误** | 功能完全不可用 | `client/src/api/index.ts:106` 原 `url: \`/api/training/task/${planId}}${day}/${taskIdx}\`` 多了一个 `}` → 已修复 |
| 2 | **未 `encodeURIComponent` 用户输入** | XSS 攻击 | 搜索功能中未编码关键词可能注入恶意字符 |
| 3 | **localStorage 存储敏感信息** | XSS 导致 token 泄露 | `uni.getStorageSync('token')` 小程序中相对安全，但需注意 |

### 🟠 P1 - 代码质量

| # | 禁止模式 | 风险 | 真实案例 |
|---|---------|------|---------|
| 4 | **滥用 `any` 类型** | 类型安全失效 | 几乎所有页面的 `ref()` 都是 `ref<any[]>` |
| 5 | **分页逻辑在每个页面重复** | 代码膨胀，bug 难追踪 | `square.vue`, `my-video.vue`, `insight/list.vue` 等重复 |
| 6 | **魔法数字和字符串** | 难以维护 | 分页大小 `10`, `20` 硬编码 |
| 7 | **未使用 `uni-*` 标签** | 跨端兼容问题 | 使用 `<div>` 而非 `<view>` 等 |

### 🟡 P2 - 性能

| # | 禁止模式 | 风险 | 真实案例 |
|---|---------|------|---------|
| 8 | **列表无 `lazy-load`** | 流量浪费 | 封面图在 square.vue 中使用 `<image>` 未启用懒加载 |
| 9 | **无端组件销毁清理** | 内存泄漏 | 页面离开时未清理计时器等 |
| 10 | **请求结果未使用 `v-memo` 或计算属性** | 不必要的重渲染 | 列表渲染未优化 |

---

## 必需模式

### P0 - 必须遵守

1. **所有用户输入在 URL 拼接前必须使用 `encodeURIComponent`**
   ```typescript
   // ✅ 正确
   url: `/api/xxx?q=${encodeURIComponent(keyword)}`
   ```

2. **API 函数必须指定泛型返回类型**
   ```typescript
   // ✅ 正确
   export function getCheckinList(page = 1) {
     return request<CheckinListResponse>({ url: `/api/checkin/list?page=${page}` })
   }
   ```

3. **Token 必须通过请求封装自动附加，不在每个页面手动处理**
   ```typescript
   // ✅ 正确：request.ts 自动附加
   if (token) {
     header['Authorization'] = `Bearer ${token}`
   }
   ```

### P1 - 强烈建议

4. **组件状态必须类型化，禁止 `any`**
   ```typescript
   // ✅ 正确
   interface Checkin { ... }
   const list = ref<Checkin[]>([])
   ```

5. **提取复用逻辑到 composables**
   ```typescript
   // ✅ 正确
   const { list, loading, loadMore } = usePagination(getCheckinList)
   ```

6. **使用 constants 统一管理常量**
   ```typescript
   export const PAGE_SIZE = 20
   export const MAX_DESCRIPTION = 200
   ```

### P2 - 建议

7. **特定组件提取到 `components/` 目录**
   - 视频卡片组件
   - 分页加载容器组件
   - 瀑布流布局组件
   - 用户信息展示组件

---

## 代码审查清单

### 安全
- [ ] URL 拼接是否使用 `encodeURIComponent`？
- [ ] 是否存在 XSS 风险（用户输入直接渲染）？
- [ ] Token 存储是否使用安全的存储方式？

### 质量
- [ ] 是否使用了类型化的 `ref<>` 而非 `any`？
- [ ] 分页/加载逻辑是否提取为复用逻辑？
- [ ] 是否存在重复代码可抽取为组件或 composable？

### 性能
- [ ] 列表图片是否使用懒加载？
- [ ] 组件销毁时是否清理了副作用？
- [ ] 是否不必要的渲染？

---

## 测试要求

**当前状态**：项目中尚无前端测试文件。

**建议**：
- 优先为 `utils/request.ts` 添加单元测试（mock uni.request）
- 优先为 Pinia store 添加单元测试
- 使用 Vitest 作为测试框架（与 Vite 构建工具集成）