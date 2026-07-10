# 性能优化

## Goal

优化数据库查询性能、连接池配置和缓存策略。

## Performance Analysis (from codebase)

### 1. 索引缺失
- `checkin_repo.Search` 使用 `$regex` 查询 `description` 字段，无文本索引
- `user_repo.FindAll` 按 `nickname` 搜索，无索引
- `rank_repo` 查询排行榜无 `score` 降序索引
- `checkin_repo.ListAll` (admin) 全表扫描无优化索引

### 2. CountDocuments 性能
- 所有分页查询都执行 `CountDocuments` + `Find` 两次查询
- `CountDocuments(ctx, bson.M{})` 全表计数，数据量大时慢

### 3. 连接池未调优
- MongoDB 连接使用默认配置，未设置 MaxPoolSize/MinPoolSize

### 4. 无缓存层
- Dashboard 统计每次请求查询 4 个集合
- 排行榜数据每 10 分钟刷新但无应用层缓存

## Requirements

### R1: 索引优化
- checkin: 添加 `description` 文本索引
- user: 添加 `nickname` 索引
- rank: 添加 `score` 降序索引

### R2: MongoDB 连接池调优
- 设置 MaxPoolSize=100, MinPoolSize=10
- 设置 SocketTimeout, ServerSelectionTimeout

### R3: Dashboard 统计缓存
- 使用内存缓存（sync.Map）缓存统计结果，TTL 60 秒

### R4: 查询优化
- `ListAll` 等 admin 查询添加适当索引
- 考虑用 `estimatedDocumentCount` 替代 `CountDocuments`（无 filter 时）

## Acceptance Criteria

- [ ] 新增 3+ 个索引
- [ ] MongoDB 连接池参数已配置
- [ ] Dashboard 统计有缓存
- [ ] `go build ./...` 编译通过

## Out of Scope

- Redis 缓存层（需要新增依赖）
- 分页改为游标式（改动太大）
- 查询慢日志分析