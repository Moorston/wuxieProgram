# 构建全局代码规范与架构约束文档

## Goal

基于对武俱打卡项目的全面代码分析，填充 `.trellis/spec/` 下的全部规范文档，使其从模板占位符变为**可执行的开发指南**。文档需反映项目实际的编码模式、架构约定和约束，而非通用最佳实践。

## Confirmed Facts (from codebase analysis)

- **后端**: Go + Gin，双服务 (api-server:8080, media-server:8081)
- **前端**: uni-app (Vue3 + TypeScript)，编译到微信小程序/Android/iOS
- **数据库**: MongoDB (15+ 集合)，Redis (缓存 + 转码队列)
- **存储**: MinIO (raw/video/cover/resource 四个桶)
- **部署**: Docker Compose (6 个服务)，Nginx 统一入口
- **项目根目录**: `wuxieProgram/` 下三个主包（api-server, media-server, client）+ deploy/
- **代码模式**:
  - Handler → Service → Repository 三层架构
  - 依赖注入式初始化（main.go 中 assembler）
  - 事务支持已实现（点赞/评论使用 MongoDB session+transaction）
  - JWT 鉴权、内部 API 密钥鉴权、CORS 均已实现
  - 统一的 response 包（code/message/data 格式）
  - 结构化日志（zap Logger）
  - 错误处理风格不统一，魔法数字散落代码
  - 缺少单元测试，Repository 层无接口定义
  - 前端统一 request 封装，Pinia 状态管理

## Requirements

### R1: 填充全部 Backend Spec 文件
- [ ] backend/directory-structure.md - 目录结构 + 模块组织 + 命名约定
- [ ] backend/quality-guidelines.md - 代码质量标准 + 禁止/必需模式
- [ ] backend/database-guidelines.md - MongoDB 模式 + 索引 + 事务约定
- [ ] backend/error-handling.md - 错误处理策略 + response 规范
- [ ] backend/logging-guidelines.md - 结构化日志 + 日志级别规范

### R2: 填充全部 Frontend Spec 文件
- [ ] frontend/directory-structure.md - 目录结构 + 页面组织
- [ ] frontend/component-guidelines.md - 组件模式 + props + 复用策略
- [ ] frontend/hook-guidelines.md - 数据获取模式 + 状态管理
- [ ] frontend/state-management.md - 全局状态 + 本地状态界限
- [ ] frontend/quality-guidelines.md - 前端规范 + 禁止模式
- [ ] frontend/type-safety.md - TypeScript 类型安全 + API 类型约定

### R3: 创建架构约束文档
- [ ] spec/architecture/backend-architecture.md - 双服务架构约束
- [ ] spec/architecture/data-flow.md - 跨层数据流 + 服务间通信
- [ ] spec/architecture/security-constraints.md - 安全约束

### R4: 更新 thinking guides 中的架构约束部分

## Acceptance Criteria

1. 每个 spec 文件必须包含**项目实际的代码示例**，引用具体文件路径和行号
2. 每个 spec 文件必须列出**本项目中该规范被违反的具体实例**
3. 禁止模式必须附有**本项目中的真实违规案例**
4. 规范必须反映项目**实际**使用的技术栈和模式（非通用最佳实践）
5. 架构约束必须覆盖：双服务通信、数据流方向、依赖注入方式、鉴权层次
6. 所有文档以**中文**撰写（符合项目约定）

## Out of Scope

- 不修改任何源代码
- 不创建新的 spec 类别（仅填充现有 + 新增 architecture 分类）
- 不涉及 CI/CD 配置

## Open Questions

- 无（所有信息已从代码库中提取）