# 后端开发规范

> 基于武俱打卡项目实际代码分析的后端开发指南。

---

## 概览

本目录包含后端开发的完整规范文档。所有文档基于代码库实际分析生成，引用具体文件路径和代码示例。

---

## 规范索引

| 指南 | 说明 | 状态 |
|------|------|------|
| [目录结构](./directory-structure.md) | 模块组织和文件布局 | ✅ 已填充 |
| [数据库规范](./database-guidelines.md) | MongoDB 模式、索引、事务 | ✅ 已填充 |
| [错误处理](./error-handling.md) | 错误类型、处理策略、响应格式 | ✅ 已填充 |
| [代码质量](./quality-guidelines.md) | 禁止/必需模式、测试要求 | ✅ 已填充 |
| [日志规范](./logging-guidelines.md) | 结构化日志、日志级别 | ✅ 已填充 |

---

## 技术栈

| 组件 | 技术 |
|------|------|
| 语言 | Go 1.21+ |
| HTTP 框架 | Gin |
| 数据库 | MongoDB (go.mongodb.org/mongo-driver) |
| 缓存/队列 | Redis (go-redis) |
| 日志 | zap (go.uber.org/zap) |
| JWT | 自实现 (pkg/jwt) |

---

## 核心原则

1. **三层严格分离**：Handler → Service → Repository，单向依赖
2. **依赖注入**：构造函数注入，无全局变量、无 `init()`
3. **事务保护**：跨集合写操作使用 MongoDB 事务
4. **安全第一**：所有内部接口需认证、错误不暴露内部细节、用户输入验证
5. **结构化日志**：使用 zap 结构化日志，禁止 log.Printf