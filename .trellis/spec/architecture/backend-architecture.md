# 架构约束文档 — 双服务架构

> 武俱打卡项目的核心架构约束：双服务分离设计。

---

## 架构概览

```
┌──────────┐     ┌─────────┐     ┌───────────────┐
│  client  │────▶│  nginx  │────▶│  api-server   │───▶ MongoDB
│ (uni-app)│     │ (:80)   │     │  (:8080)      │───▶ Redis
└──────────┘     └─────────┘     │  Go + Gin     │
       │              │          └───────┬───────┘
       │              │                  │ HTTP 回调 (X-Internal-Secret)
       │              │                  ▼
       │              └──────────▶┌───────────────┐
       │                          │ media-server  │───▶ MinIO
       │                          │  (:8081)      │───▶ Redis (转码队列)
       │                          │  Go + FFmpeg  │
       │                          └───────────────┘
       │                                  │
       └─────── 预签名 URL 直传 MinIO ────┘
```

---

## 核心约束

### C1: 双服务职责不可混淆

| 服务 | 职责 | 非职责 |
|------|------|--------|
| **api-server** | 业务逻辑（用户/打卡/社交/训练/通知/感悟/资料库） | ❌ 视频转码、❌ 视频上传直接处理 |
| **media-server** | 视频处理（上传预签名/转码/封面提取/播放URL） | ❌ 用户管理、❌ 业务数据操作 |

### C2: api-server 是唯一数据主控

- api-server 是唯一直接操作 MongoDB 和业务 Redis 的服务
- media-server **不得**连接 MongoDB 或操作业务数据
- media-server 仅消费 Redis 中的转码队列（`transcode:queue`）
- media-server 的持久化操作仅限于 MinIO 文件管理

### C3: 服务间通信需认证

- api-server 和 media-server 之间的 HTTP 回调必须携带 `X-Internal-Secret` 请求头
- 密钥在 `config.yaml` 中配置，两个服务共享同一密钥
- 内部接口路由使用 `InternalAuth` 中间件保护

```yaml
# api-server/configs/config.yaml
media_secret: "wuxie-media-secret-change-in-production"

# media-server/configs/config.yaml
api_secret: "wuxie-media-secret-change-in-production"  # 必须一致
```

```go
// router.go — 内部接口路由
internal := api.Group("/internal")
internal.Use(middleware.InternalAuth(cfg.MediaSecret))
{
    internal.POST("/transcode/done", checkinH.TranscodeCallback)
}
```

### C4: 客户端通过预签名 URL 直传 MinIO

- 客户端**不经过 api-server 或 media-server 上传文件**
- 客户端先请求预签名 PUT URL，然后直接上传到 MinIO
- 上传完成后通知 media-server（回调 URL）
- media-server 将转码任务推入 Redis 队列
- worker 消费队列完成转码

```
流程图：
客户请求预签名URL → media-server 返回 MinIO 预签名 PUT URL
                  ↓
客户直接上传到 MinIO
                  ↓
客户通知 media-server "上传完成"
                  ↓
media-server 推送转码任务到 Redis 队列
                  ↓
worker 消费队列 → FFmpeg 转码 → 上传转码后视频/封面到 MinIO
                  ↓
media-server 回调 api-server /api/internal/transcode/done
                  ↓
api-server 更新 checkin 状态为 "完成"
```

### C5: 视频播放 URL 有时效性

- video_url 和 cover_url 不持久存储完整 URL
- 每次播放时通过 media-server 获取预签名 GET URL（有效期 2 小时）
- Nginx 配置 `Range` 请求支持以支持视频拖拽播放

---

## 模块间依赖约束

### 三层依赖方向（严格单向）

```
Handler → Service → Repository → MongoDB
   │          │           │
   │          │           └── 操作数据库
   │          └── 业务逻辑编排
   └── HTTP 请求处理
```

- ❌ Handler 不能直接调用 Repository
- ❌ Handler 不能直接操作数据库
- ❌ Service 不能直接返回 HTTP 响应
- ❌ Repository 不能包含业务逻辑

### 依赖注入规则

- 所有依赖通过**构造函数注入**，不使用全局变量或 `init()`
- 注入方向：`main.go` → Handler → Service → Repository

```go
// ✅ 唯一合法的依赖注入方式
userRepo := repository.NewUserRepo(db)
userService := service.NewUserService(userRepo)
userH := handler.NewUserHandler(userService)
```

---

## 配置管理约束

### 配置加载优先级

当前：`config.yaml` → 环境变量（计划中）

### 敏感配置项

| 配置项 | 风险 | 生产要求 |
|--------|------|---------|
| `jwt.secret` | token 可伪造 | ≥32 字符随机字符串 |
| `wx.secret` | 微信接口可调用 | 环境变量覆盖 |
| `media_secret` | 内部接口可调用 | 与 media-server 一致 |

### 配置只读约束

- 配置在服务启动时加载一次，运行期内**不能修改**
- 配置变更需要重启服务

---

## API 路由约束

### 路由分类

| 类别 | 前缀 | 鉴权 | 示例 |
|------|------|------|------|
| 公开接口 | `/api/auth/*` | 无 | `/api/auth/login` |
| 业务接口 | `/api/*` | JWT Auth | `/api/user/profile` |
| 内部接口 | `/api/internal/*` | InternalAuth | `/api/internal/transcode/done` |
| 健康检查 | `/health` | 无 | `/health` |

### 路由注册约束

- 所有路由在 `router/router.go` 中集中注册
- 禁止在 handler 中注册路由
- 禁止在 middleware 中注册路由

---

## Nginx 代理约束

| 路径 | 代理目标 | 说明 |
|------|---------|------|
| `/api/*` | `api-server:8080` | 业务 API |
| `/media/*` | `media-server:8081` | 视频服务 |
| 其他可配置 | — | 扩展新服务时新增 location |

---

## 违反这些约束的后果

| 约束编号 | 典型违规 | 后果 |
|---------|---------|------|
| C1 | media-server 操作 MongoDB | 数据不一致、职责混乱 |
| C2 | api-server 直接处理视频上传 | 业务服务阻塞 |
| C3 | 内部接口无认证 | 数据可被伪造 |
| C4 | 文件经过服务端中转 | 带宽瓶颈、延迟增加 |
| C5 | 永久视频 URL | 盗链风险 |