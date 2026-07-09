# 武俱打卡项目规范索引

> 基于代码库分析构建的完整项目规范文档。

---

## 📂 目录结构

| 目录 | 说明 | 状态 |
|------|------|------|
| [architecture/](./architecture/backend-architecture.md) | 架构约束、数据流、安全规范 | ✅ 已填充 |
| [backend/](./backend/index.md) | 后端开发规范 | ✅ 已填充 |
| [frontend/](./frontend/index.md) | 前端开发规范 | ✅ 已填充 |
| [guides/](./guides/index.md) | 思维引导指南 | ✅ 已有内容 |

---

## 🏗️ 架构规范

| 文档 | 主要内容 |
|------|---------|
| [双服务架构](./architecture/backend-architecture.md) | api-server + media-server 职责分离、依赖方向、API 路由约束 |
| [数据流规范](./architecture/data-flow.md) | 层间数据流、服务间通信、数据格式、禁止数据流模式 |
| [安全约束](./architecture/security-constraints.md) | 认证/授权、输入验证、数据安全、防护措施 |

---

## 🔧 后端规范

| 文档 | 主要内容 |
|------|---------|
| [目录结构](./backend/directory-structure.md) | 三层架构布局、模块组织、命名约定、违规案例 |
| [数据库规范](./backend/database-guidelines.md) | 集合设计、索引策略、事务模式、Repository 方法约定 |
| [错误处理](./backend/error-handling.md) | 错误响应格式、层级错误处理、禁止模式、常见错误 |
| [代码质量](./backend/quality-guidelines.md) | 禁止/必需模式、代码审查清单、测试要求 |
| [日志规范](./backend/logging-guidelines.md) | 结构化日志、日志级别、敏感信息脱敏 |

---

## 🎨 前端规范

| 文档 | 主要内容 |
|------|---------|
| [目录结构](./frontend/directory-structure.md) | 页面组织、API/Store/Utils 分层、命名约定 |
| [组件规范](./frontend/component-guidelines.md) | 页面组件结构、Props/Events 约定、样式规范 |
| [Hook 规范](./frontend/hook-guidelines.md) | Composable 模式、建议提取的复用逻辑 |
| [状态管理](./frontend/state-management.md) | Pinia Store 结构、全局 vs 本地状态界限 |
| [代码质量](./frontend/quality-guidelines.md) | 禁止/必需模式、审查清单 |
| [类型安全](./frontend/type-safety.md) | 类型组织、API 类型定义、禁止 `any` |

---

## 🧠 思维指南

| 文档 | 用途 |
|------|------|
| [代码复用思考](./guides/code-reuse-thinking-guide.md) | 减少重复代码 |
| [跨层思考](./guides/cross-layer-thinking-guide.md) | 分析跨层数据流 |

---

## 生成时间

所有规范文档基于 `2026-07-08` 的代码库分析生成，反映了项目**实际的**代码模式、架构约定和已知问题。