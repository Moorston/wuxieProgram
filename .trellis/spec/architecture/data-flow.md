# 数据流约束

> 武俱打卡项目的跨层数据流和服务间通信规范。

---

## 数据流总览

```
┌─────────────────────────────────────────────────────────┐
│                    前端请求周期                           │
│                                                         │
│  客户端 → HTTP请求 → Nginx → API Server → 认证 → Handler│
│                                                     ↓   │
│  ← JSON响应 ← response包 ← Handler ← Service ← Repository│
│                                                     ↓   │
│                                                MongoDB   │
└─────────────────────────────────────────────────────────┘
```

---

## 层间数据流

### 1. Handler → Service

Handler 负责将 HTTP 请求参数**转换**为 Service 层的 Go 类型：

```go
// Handler 层：解析请求 → 调用 Service
func (h *CheckinHandler) GetByID(c *gin.Context) {
    id, err := primitive.ObjectIDFromHex(c.Param("id"))   // 参数类型转换
    if err != nil {
        response.BadRequest(c, "invalid checkin id")
        return
    }
    checkin, err := h.checkinService.GetByID(c.Request.Context(), id)  // 调用 Service
    if err != nil {
        response.NotFound(c, "checkin not found")
        return
    }
    response.Success(c, checkin)  // 返回响应
}
```

**数据流约束**：
- Handler 只做参数解析和响应发送，不包含业务逻辑
- Service 方法的上下文始终传递 `context.Context`（用于超时控制和链路追踪）
- Service 不能访问 `gin.Context`（不能获取 HTTP 请求/响应信息）

### 2. Service → Repository

Service 负责编排业务逻辑，Repository 负责数据访问：

```go
// Service 层：业务逻辑编排
func (s *CheckinService) GetList(ctx context.Context, userID primitive.ObjectID, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
    // 1. 调用 Repository 获取数据
    checkins, total, err := s.checkinRepo.List(ctx, userID, groupUserIDs, page, pageSize)
    if err != nil {
        return nil, 0, err
    }

    // 2. 跨 Repository 拼装数据（批量查询避免 N+1）
    users, err := s.userRepo.FindByIDs(ctx, userIDs)
    // ...
    for _, c := range checkins {
        c.User = userMap[c.UserID]
    }

    return checkins, total, nil
}
```

**数据流约束**：
- Service 编排多个 Repository 调用时使用**批量查询**避免 N+1
- Service 使用 **MongoDB 事务**确保跨集合写操作的原子性
- Service 负责将 Repository 的原始错误包装为业务错误

### 3. Repository → MongoDB

Repository 只做数据存取，不含业务逻辑：

```go
// Repository 层：纯数据访问
func (r *UserRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
    cursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
    if err != nil {
        return nil, err
    }
    var users []*model.User
    if err := cursor.All(ctx, &users); err != nil {
        return nil, err
    }
    return users, nil
}
```

---

## 服务间数据流

### API Server → Media Server

API Server **不主动调用** Media Server，仅通过配置定义 Media URL：

```go
// api-server/service/service.go:124
type CheckinService struct {
    checkinRepo *repository.CheckinRepo
    userRepo    *repository.UserRepo
    mediaURL    string  // 从配置读取 "http://media-server:8081"
}
```

### Media Server → API Server（回调）

Media Server 在转码完成后回调 API Server：

```
POST /api/internal/transcode/done
Headers:
  X-Internal-Secret: <shared-secret>
  Content-Type: application/json
Body:
  {
    "checkin_id": "...",
    "video_url": "...",
    "cover_url": "...",
    "duration": 30.5
  }
```

### 客户端 → MinIO（预签名 URL）

```
1. 客户端 GET /media/upload/presign → 获取预签名 PUT URL
2. 客户端 PUT <预签名URL> → 直接上传到 MinIO
3. 客户端 POST /media/upload/callback → 通知上传完成
```

---

## 数据格式约束

### 时间格式

- 所有时间使用 ISO 8601 格式（`time.RFC3339`）在 API 中传输
- MongoDB 中存储为 `time.Time`（BSON Date）
- 前端显示时根据时区本地化

```go
// 时间序列化示例
type Checkin struct {
    CreatedAt time.Time `json:"created_at" bson:"created_at"`
}
```

### ID 格式

- 所有资源 ID 使用 MongoDB ObjectID（`primitive.ObjectID`）
- API 中以十六进制字符串传输
- 前端传递 ID 时始终为字符串类型

```go
// Handler 中转换
id, err := primitive.ObjectIDFromHex(c.Param("id"))
oid, err := primitive.ObjectIDFromHex(userID)
```

### 分页格式

```json
// 请求
GET /api/checkin/list?page=1&page_size=10

// 响应
{
  "code": 0,
  "message": "success",
  "data": {
    "list": [...],
    "total": 100,
    "page": 1,
    "page_size": 10
  }
}
```

---

## 错误数据流

```
Repository: 返回原始错误或 mongo.ErrNoDocuments
     ↓
Service: fmt.Errorf("operation failed: %w", err) 包装
     ↓
Handler: log 记录详细错误 → response 返回通用信息
     ↓
客户端: 接收统一的 JSON 错误响应
```

---

## 禁止的数据流模式

### ❌ Service 直接操作 HTTP 响应

```go
// ❌ 禁止：Service 访问 gin.Context
func (s *UserService) GetProfile(c *gin.Context, id ...) {
    c.JSON(200, user)  // Service 不能处理 HTTP 响应
}

// ✅ 正确：Service 返回数据
func (s *UserService) GetProfile(ctx context.Context, id ...) (*model.User, error) {
    return s.userRepo.FindByID(ctx, id)
}
```

### ❌ Handler 直接操作数据库

```go
// ❌ 禁止：Handler 绕过 Service 直接查数据库
func (h *UserHandler) GetProfile(c *gin.Context) {
    var user model.User
    h.db.Collection("users").FindOne(...)  // Handler 不能操作数据库
}

// ✅ 正确：Handler 调用 Service
func (h *UserHandler) GetProfile(c *gin.Context) {
    user, err := h.userService.GetProfile(ctx, oid)
}
```

### ❌ 服务间同步调用链过长

- 避免 A → B → C → D 的深层调用链
- 服务间最多一级调用（api-server ↔ media-server）