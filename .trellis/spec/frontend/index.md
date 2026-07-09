# 前端开发规范

> 基于武俱打卡项目实际代码分析的前端开发指南。

---

## 概览

本目录包含前端开发的完整规范文档。所有文档基于代码库实际分析生成，引用具体文件路径和代码示例。

---

## 规范索引

| 指南 | 说明 | 状态 |
|------|------|------|
| [目录结构](./directory-structure.md) | 模块组织和文件布局 | ✅ 已填充 |
| [组件规范](./component-guidelines.md) | 组件模式、Props、样式 | ✅ 已填充 |
| [Hook 规范](./hook-guidelines.md) | Composable 模式、数据获取 | ✅ 已填充 |
| [状态管理](./state-management.md) | Pinia Store、全局 vs 本地状态 | ✅ 已填充 |
| [代码质量](./quality-guidelines.md) | 禁止/必需模式、审查清单 | ✅ 已填充 |
| [类型安全](./type-safety.md) | TypeScript 类型模式、API 类型 | ✅ 已填充 |

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 框架 | uni-app (Vue3 + TypeScript) |
| 状态管理 | Pinia |
| 构建工具 | Vite |
| 请求封装 | 自实现 request.ts |
| 目标平台 | 微信小程序 / Android / iOS |
| UI 风格 | rpx 响应式布局 |

---

## 核心原则

1. **类型安全**：全量 TypeScript，禁止 `any`
2. **逻辑复用**：提取 composables 替代页面内重复的分页/加载逻辑
3. **组件化**：提取可复用组件到 `components/`，避免重复代码
4. **安全处理**：用户输入在 URL 拼接前必须 `encodeURIComponent`
5. **分端兼容**：使用 uni-app 标签（`view`/`text`/`image`），禁用 HTML 原生标签