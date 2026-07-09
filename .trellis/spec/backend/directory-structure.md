# 后端目录结构与模块组织规范

> 武俱打卡项目后端代码的组织方式和命名约定。

---

## 概览

后端采用 **Handler → Service → Repository** 三层架构，依赖注入采用**显式构造函数注入**模式。所有初始化在 `cmd/main.go` 中完成装配（assembler）。

---

## 目录布局

```
api-server/                         # 业务 API 服务 (端口 8080)
├── cmd/
│   └── main.go                     # 服务入口：配置加载 → 依赖注入 → 路由注册 → HTTP 启动
├── configs/
│   └── config.yaml                 # 所有可配置参数（含敏感信息占位）
├── internal/
│   ├── config/
│   │   └── config.go               # 配置加载（Viper 风格，从 yaml 读取）
│   ├── constants/
│   │   └── constants.go            # 全局常量（分页大小、积分、长度限制）
│   ├── handler/                     # HTTP 处理器层：请求解析 + 响应发送
│   │   ├── handler.go               # 核心 handler（Auth/User/Checkin/Social/Rank/Group）
│   │   ├── helper.go                # handler 辅助函数（getUserID 等）
│   │   ├── training_handler.go      # 训练计划 handler
│   │   ├── insight_handler.go       # 感悟笔记 handler
│   │   ├── notification_handler.go  # 消息通知 handler
│   │   └── resource_handler.go      # 个人资料库 handler
│   ├── middleware/                  # Gin 中间件层
│   │   ├── auth.go                  # JWT 鉴权中间件
│   │   ├── cors.go                  # CORS 跨域中间件
│   │   ├── internal_auth.go         # 内部 API 密钥认证中间件
│   │   ├── logger.go                # 请求日志中间件
│   │   └── rate_limit.go            # 速率限制中间件
│   ├── model/                       # 数据模型层
│   │   ├── user.go
│   │   ├── checkin.go
│   │   ├── group.go
│   │   ├── rank.go
│   │   ├── training.go
│   │   ├── insight.go
│   │   ├── notification.go
│   │   └── resource.go
│   ├── repository/                  # 数据访问层（MongoDB 操作）
│   │   ├── user_repo.go
│   │   ├── checkin_repo.go
│   │   ├── social_repo.go           # 评论/点赞（含 EnsureIndexes）
│   │   ├── rank_repo.go
│   │   ├── training_repo.go
│   │   ├── template_repo.go
│   │   ├── insight_repo.go
│   │   ├── insight_tag_repo.go
│   │   ├── notification_repo.go
│   │   ├── resource_repo.go
│   │   └── resource_tag_repo.go
│   ├── router/
│   │   └── router.go                # 路由注册（所有 60+ 端点在此集中定义）
│   └── service/                     # 业务逻辑层
│       ├── service.go               # 核心服务（Auth/User/Checkin/Social/Rank/Group）
│       ├── training_service.go      # 训练计划服务
│       ├── insight_service.go       # 感悟笔记服务
│       ├── notification_service.go  # 消息通知服务
│       ├── resource_service.go      # 个人资料库服务
│       ├── cron.go                  # 定时任务（排行榜刷新 + 训练提醒）
│       └── utils.go                 # 工具函数（正则清洗等）
└── pkg/                            # 可复用公共包
    ├── jwt/jwt.go                   # JWT 生成与验证
    ├── response/response.go         # 统一响应格式
    └── wx/wx.go                     # 微信客户端封装

media-server/                       # 视频服务 (端口 8081)
├── cmd/main.go
├── configs/config.yaml
├── internal/
│   ├── config/config.go
│   ├── handler/handler.go           # 上传预签名/回调/视频 URL
│   ├── middleware/auth.go           # 可选鉴权中间件
│   ├── router/router.go
│   └── worker/worker.go             # FFmpeg 转码 Worker（消费 Redis 队列）
└── pkg/
    ├── ffmpeg/ffmpeg.go             # FFmpeg 封装（转码 + 封面提取）
    ├── minio/minio.go               # MinIO 存储封装（CRUD + 预签名 URL）
    └── response/response.go         # 统一响应格式
```

---

## 模块组织原则

### 三层职责边界

| 层 | 职责 | 禁止 |
|----|------|------|
| **Handler** | 解析请求参数、调用 Service、发送 Response | 直接操作数据库、访问配置 |
| **Service** | 业务逻辑编排、事务管理、调用 Repository | 直接返回 HTTP 响应、访问 `gin.Context` |
| **Repository** | MongoDB CRUD、索引管理、聚合查询 | 包含业务逻辑 |

### 文件拆分规则

- **按业务模块拆分**：当单个文件超过 300 行或包含 3 个以上类型时，拆分到独立文件
  - ✅ `service.go` → `service.go`, `training_service.go`, `insight_service.go`（已拆分）
  - ❌ 反例：`handler.go` 414 行仍包含 Auth/User/Checkin/Social/Rank/Group 6 个类型 → 建议按模块拆分
- **Repository 按集合拆分**：每个 MongoDB Collection 对应一个 Repo 文件
  - ✅ `user_repo.go`, `checkin_repo.go`, `notification_repo.go` 等

### 定义顺序约定

在每个文件中，类型定义遵循以下顺序：
1. `type XxxService struct { ... }` — 结构体定义
2. `func NewXxxService(...) *XxxService { ... }` — 构造函数
3. `func (s *XxxService) MethodA(...)` — 公开方法
4. `func (s *XxxService) methodB(...)` — 私有方法

### 依赖注入模式

所有依赖通过构造函数参数显式注入（不使用全局变量或 `init()`）：

```go
// ✅ 正确：构造函数注入
func NewCheckinService(checkinRepo *repository.CheckinRepo, userRepo *repository.UserRepo, mediaURL string) *CheckinService {
    return &CheckinService{checkinRepo: checkinRepo, userRepo: userRepo, mediaURL: mediaURL}
}

// ❌ 禁止：使用全局变量
var globalCheckinRepo *repository.CheckinRepo
```

初始化在 `cmd/main.go` 中按顺序（Repository → Service → Handler → Router）完成装配。

---

## 命名约定

### 文件命名

- **Go 文件**：`snake_case.go`
  - 业务文件：`training_handler.go`, `insight_service.go`
  - Repository 文件：`user_repo.go`, `checkin_repo.go`
- **接口文件**：按集合名命名（`*_spec.go`、`*_test.go`）

### 类型命名

| 元素 | 约定 | 示例 |
|------|------|------|
| Handler | `{Module}Handler` | `AuthHandler`, `CheckinHandler` |
| Service | `{Module}Service` | `AuthService`, `SocialService` |
| Repository | `{Collection}Repo` | `UserRepo`, `InsightRepo` |
| 请求体 | `{Action}Req` | `LoginReq`, `UpdateProfileReq` |
| 模型 | `{CollectionName}` | `User`, `Checkin`, `TrainingPlan` |

### 变量命名

- **缩写保持小写**：`jwtMgr` ✅ | `jwtManager` ✅ | `JWTManager` ❌ (类型)
- **不缩写或仅用广泛接受的缩写**：`notificationService` ✅ | `notifService` ✅ | `ns` ❌
- **ID 始终大写**：`userID`, `checkinID`, `planID` ✅ | `userId`, `checkinId` ❌

### 路由命名

- **资源路由使用复数**：`/api/checkin/*`, `/api/notification/*`, `/api/resource/*`
  - 例外：`/api/rank`（不可数）
- **路径参数用小写 snake_case**：`/api/training/task/:plan_id/:day/:task_idx`

---

## 代码示例

### 三层协作模式（最佳实践）

```go
// handler/handler.go
func (h *CheckinHandler) GetByID(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))    // Handler: 解析请求
    if err != nil {
        response.BadRequest(c, "invalid checkin id")
        return
    }
    checkin, err := h.checkinService.GetByID(c.Request.Context(), id)  // Handler: 调用 Service
    if err != nil {
        response.NotFound(c, "checkin not found")
        return
    }
    response.Success(c, checkin)
}

// service/service.go
func (s *CheckinService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
    checkin, err := s.checkinRepo.FindByID(ctx, id)        // Service: 调用 Repository
    if err != nil {
        return nil, err
    }
    user, err := s.userRepo.FindByID(ctx, checkin.UserID)  // Service: 拼装数据
    if err == nil {
        checkin.User = user
    }
    return checkin, nil
}

// repository/checkin_repo.go
func (r *CheckinRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
    var checkin model.Checkin
    err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&checkin)  // Repository: 数据访问
    return &checkin, err
}
```

---

## 本项目中的违规例

- **单文件过大**：`handler.go` (414 行) 包含 6 个 Handler 类型，应拆分为独立文件
- **service.go 中的组函数**：`service.go` (464 行) 包含 AuthService/UserService/CheckinService/SocialService/RankService/GroupService，应拆分到独立文件
- **命名不一致**：部分变量用缩写 `notifService` 而非 `notificationService`
- **注释缺失**：公共函数缺少文档注释（如 `Search()`, `GetTodayTasks()` 等）