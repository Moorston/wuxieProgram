# 管理后台功能规划

## Goal

为武俱打卡应用添加独立 Web 管理端：用户管理、内容审核、数据统计。

## Design Decisions

### D1: 管理端形态 — 独立 Web 应用
- 独立于 UniApp 小程序的 Web 管理端
- 可使用 Vue 3 或纯 HTML/CSS/JS 构建
- 部署在 `admin.example.com` 或 `/admin` 路径

### D2: 认证方式 — 账号密码登录
- 新增 `POST /api/admin/login` 端点
- 使用用户名 + 密码登录，返回 JWT
- 使用现有 `pkg/jwt` 签发 token，增加 `role` claim
- 密码使用 bcrypt 存储

### D3: 管理员角色（最小改动）
- User 模型增加 `Role` 字段（0=普通用户, 1=管理员）
- `Role` 不暴露给前端，只通过 admin API 识别
- 管理员通过后台 API 设置（开发阶段手动 DB 操作）

## Architecture

```
┌─────────────────────┐     ┌──────────────────────────┐
│  Web Admin (Vue 3)  │────▶│  API Server (Go + Gin)   │
│  /admin/*           │     │                          │
│                     │     │  /api/admin/login        │
│  - 用户管理         │     │  /api/admin/users        │
│  - 内容审核         │     │  /api/admin/checkins     │
│  - 数据统计         │     │  /api/admin/stats        │
│  - 仪表盘           │     └──────────┬───────────────┘
└─────────────────────┘                │
                                       ▼
                                ┌──────────────┐
                                │   MongoDB    │
                                └──────────────┘
```

## Requirements

### R1: 后端 API
- `POST /api/admin/login` — 管理员登录，返回 JWT
- `GET /api/admin/users` — 用户列表（分页）
- `PUT /api/admin/users/:id/ban` — 封禁用户
- `PUT /api/admin/users/:id/unban` — 解封用户
- `GET /api/admin/checkins` — 打卡列表（全部用户）
- `DELETE /api/admin/checkins/:id` — 删除打卡
- `GET /api/admin/insights` — 感悟列表（全部用户）
- `DELETE /api/admin/insights/:id` — 删除感悟
- `GET /api/admin/stats` — 仪表盘统计数据

### R2: 管理员中间件
- 复用现有 Auth 中间件，在其后增加 AdminOnly 检查
- 验证 JWT 中 `role` claim == 1

### R3: Web 管理端（Vue 3）
- 登录页
- 仪表盘（总用户数、今日打卡数、活跃用户数）
- 用户管理页（列表、搜索、封禁/解封）
- 内容管理页（打卡列表、感悟列表、删除操作）

## Acceptance Criteria

- [ ] API 新增 9 个 admin 端点
- [ ] AdminOnly 中间件生效
- [ ] 管理员登录返回 JWT
- [ ] 非管理员调用 admin API 返回 403
- [ ] Web 管理端可编译运行

## Out of Scope

- 权限细分（超管、运营等）
- 微信消息推送
- 系统配置管理
- 操作日志记录

## Open Questions

（无阻塞性问题 — 架构已确定）