# 数据库规范

> 武俱打卡项目的 MongoDB 使用约定、索引策略和事务规范。

---

## 概览

项目使用 MongoDB 作为主数据库，`go.mongodb.org/mongo-driver` 官方驱动。数据库名 `wuxie`，包含 15+ 个集合。

Redis 用于缓存（排行榜缓存）和队列（视频转码任务队列）。

---

## MongoDB 连接配置

所有数据库配置在 `configs/config.yaml` 集中管理：

```yaml
mongo:
  uri: "mongodb://mongo:27017"
  database: "wuxie"
```

初始化方式（`api-server/cmd/main.go:47-52`）：

```go
mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI))
db := mongoClient.Database(cfg.Mongo.Database)
```

---

## 集合设计规范

### 命名约定

- **集合名**：全小写 snake_case、复数
  - ✅ `users`, `checkins`, `training_plans`, `notification_settings`
  - ❌ `user`, `Checkin`, `trainingPlan`
- **字段名**：全小写 snake_case
  - ✅ `user_id`, `created_at`, `like_count`
  - ❌ `userId`, `createdAt`, `likeCount`

### 集合总览

| 集合 | 说明 | 索引数 | EnsureIndexes 位置 |
|------|------|--------|-------------------|
| users | 用户 | 3 | `user_repo.go` |
| checkins | 打卡记录 | 4 | `checkin_repo.go` |
| comments | 评论 | 3 | `social_repo.go` |
| likes | 点赞（含唯一索引） | 2 | `social_repo.go` |
| groups | 考核组 | 1 | 未实现 EnsureIndexes |
| rank_cache | 排行榜缓存 | 2 | `rank_repo.go` |
| training_plans | 训练计划 | 4 | `training_repo.go` |
| training_templates | 训练模板 | 3 | `template_repo.go` |
| notifications | 通知 | 3 | `notification_repo.go` |
| notification_settings | 通知设置（唯一索引） | 1 | `notification_repo.go` |
| insights | 感悟笔记 | 4 | `insight_repo.go` |
| insight_tags | 感悟标签（唯一索引） | 2 | `insight_tag_repo.go` |
| insight_likes | 感悟点赞（唯一索引） | 1 | `insight_like_repo.go` |
| resources | 个人资料 | 5 | `resource_repo.go` |
| resource_tags | 资料标签（唯一索引） | 2 | `resource_tag_repo.go` |

---

## 索引策略

### 索引定义模式

每个 Repository 必须实现 `EnsureIndexes(ctx context.Context) error` 方法，在 `cmd/main.go` 启动时统一调用：

```go
// ✅ 索引定义模式（api-server/internal/repository/checkin_repo.go）
func (r *CheckinRepo) EnsureIndexes(ctx context.Context) error {
    _, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
        {Keys: bson.D{{Key: "status", Value: 1}}},
        {Keys: bson.D{{Key: "description", Value: "text"}}},
        {Keys: bson.D{{Key: "like_count", Value: -1}}},
    })
    return err
}
```

### 索引分类

| 类型 | 用途 | 示例 |
|------|------|------|
| **单字段索引** | 精确过滤 | `{status: 1}` |
| **复合索引** | 多条件查询排序 | `{user_id: 1, created_at: -1}` |
| **唯一索引** | 防重复 | `{openid: 1}` (unique), `{user_id: 1, checkin_id: 1}` (unique in likes) |
| **文本索引** | 全文搜索 | `{description: "text"}` |
| **排序索引** | ORDER BY 优化 | `{like_count: -1}`, `{created_at: -1}` |

### 关键查询路径与所需索引

| 查询场景 | 集合 | 关键索引 |
|----------|------|---------|
| 按 openid 查用户登录 | users | `{openid: 1}` (unique) |
| 广场列表（按状态+时间） | checkins | `{status: 1, created_at: -1}` |
| 我的打卡（按用户+时间） | checkins | `{user_id: 1, created_at: -1}` |
| 打卡搜索（按关键词） | checkins | `{description: "text"}` |
| 防重复点赞 | likes | `{user_id: 1, checkin_id: 1}` (unique) |
| 按积分排行榜 | rank_cache | `{period: 1, score: -1}` |
| 按用户查通知 | notifications | `{user_id: 1, created_at: -1}` |
| 按用户+状态查计划 | training_plans | `{user_id: 1, status: 1}` |
| 多维筛选资料 | resources | `{user_id: 1, type: 1, share_scope: 1}` |

### 本项目已知的索引问题

- **缺失索引**: `checkins.group_id`（考核组筛选）、`resources.created_at`（资料排序）、部分集合缺少 TTL 索引
- **EnsureIndexes 缺失**: `groups` 集合尚未实现

---

## 事务规范

### 使用场景

事务**仅用于跨集合的写一致性**。当前使用事务的两个场景：

1. **点赞切换** (`service.go:250-276`): likes 集合 + checkins 集合 (like_count)
2. **添加评论** (`service.go:307-322`): comments 集合 + checkins 集合 (comment_count)

### 事务模式

```go
// ✅ 正确的事务模式
func (s *SocialService) ToggleLike(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
    session, err := s.likeRepo.StartSession()
    if err != nil {
        return false, fmt.Errorf("failed to start session: %w", err)
    }
    defer session.EndSession(ctx)

    var liked bool
    _, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
        liked, err := s.likeRepo.ToggleWithSession(sessCtx, checkinID, userID)
        if err != nil { return nil, err }

        delta := -1
        if liked { delta = 1 }
        if err = s.checkinRepo.IncrLikeCountWithSession(sessCtx, checkinID, delta); err != nil {
            return nil, err
        }
        return nil, nil
    })
    if err != nil { return false, fmt.Errorf("transaction failed: %w", err) }

    // 非关键操作在事务外执行
    if liked && s.notifService != nil {
        // ... 发送通知 ...
    }
    return liked, nil
}
```

### 事务规则

| 规则 | 说明 |
|------|------|
| **仅写操作需要事务** | 读操作不需要事务 |
| **事务内不进行 HTTP 调用** | 避免分布式事务问题 |
| **非关键操作放在事务外** | 如通知发送、日志记录 |
| **事务超时** | 默认 30 秒（mongo driver 默认值） |

---

## Repository 方法签名约定

### 方法命名

- `FindBy*` — 查询单条记录
- `List*` / `ListBy*` — 查询多条记录
- `Create` — 创建记录
- `Update*` — 更新记录
- `Delete` — 删除记录
- `EnsureIndexes` — 创建索引
- `Incr*` — 原子计数器操作

### 事务方法后缀

需要事务支持的 repo 方法加 `WithSession` 后缀：

```go
// 非事务版本
func (r *LikeRepo) Toggle(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error)

// 事务版本
func (r *LikeRepo) ToggleWithSession(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error)
```

### 错误处理

- Repository 层返回原始错误或 `mongo.ErrNoDocuments`
- Service 层用 `fmt.Errorf("...: %w", err)` 包装
- Handler 层转换为 HTTP 响应

---

## 查询模式

### 分页

```go
// 统一使用 skip + limit（或 FindOptions）
filter := bson.M{"user_id": userID}
opts := options.Find().
    SetSort(bson.D{{Key: "created_at", Value: -1}}).
    SetSkip(int64((page - 1) * pageSize)).
    SetLimit(int64(pageSize))
```

### 批量查询

使用 `$in` 操作符一次性查询，避免 N+1：

```go
// ✅ 正确的批量查询
users, err := s.userRepo.FindByIDs(ctx, userIDs)

// FindByIDs 内部实现
func (r *UserRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
    cursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
}
```

---

## Redis 规范

### 用途限定

| 用途 | Key 模式 | 说明 |
|------|---------|------|
| 转码任务队列 | `transcode:queue` | Redis List, LPUSH + BRPop 阻塞消费 |

### 禁止用途

- ❌ 用户会话存储（使用 JWT）
- ❌ 业务数据主存储（使用 MongoDB）
- ❌ 跨服务 RPC 通信（使用 HTTP 回调）

---

## 数据归档策略

**当前状态**: 项目尚未实现数据归档

**建议**:
- 打卡记录超过 1 年的迁移到 `checkins_archive` 集合
- 通知记录超过 3 个月的删除或归档
- 为 `rank_cache` 添加 TTL 索引（3 小时过期）

---

## 常见错误

1. **未在 EnsureIndexes 中添加新查询的索引**
   - 症状：集合扫描导致查询慢
   - 修复：添加新查询路径后，同步在对应 repo 的 EnsureIndexes 中添加索引

2. **忘记使用 WithSession 方法**
   - 症状：事务内调用非事务方法
   - 修复：确保事务内部调用 `*WithSession` 版本

3. **高 skip 值**
   - 症状：大页码查询越来越慢
   - 修复：使用 `_id` 游标分页或限制最大页码

4. **使用 `$regex` 代替 `$text`**
   - 症状：搜索关键词可能触发 ReDoS
   - 修复：优先使用 MongoDB 文本索引 (`$text`)，回退时严格限制 keyword 长度