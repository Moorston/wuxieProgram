# api-server Go 代码质量审查报告

审查时间：2026-07-09
审查范围：api-server 全部 Go 源文件（40+ 文件）
审查类型：错误处理、并发安全、资源泄漏、逻辑错误、魔法数字、代码组织

---

## 严重级别说明

| 级别 | 定义 |
|------|------|
| CRITICAL | 必然导致运行时 panic 或数据损坏 |
| HIGH | 大概率导致功能异常、安全漏洞或严重性能问题 |
| MEDIUM | 一定条件下引发问题，或影响可维护性 |
| LOW | 代码风格、小优化、最佳实践建议 |

---

## 一、CRITICAL 级问题

### 1.1 nil 指针解引用 — UpdatePlan 中 GetPlan 失败后访问 plan.UserID

**文件**: `api-server/internal/handler/training_handler.go:160`
**问题**: `UpdatePlan` 先调用 `GetPlan` 获取 plan，但 `err != nil` 和 `plan.UserID` 合并在同一个 if 条件中。当 `GetPlan` 返回错误时，`plan` 为 nil，访问 `plan.UserID` 会触发 nil pointer dereference panic。
**代码**:
```go
plan, err := h.trainingService.GetPlan(c.Request.Context(), id)
if err != nil || plan.UserID != oid {  // plan 为 nil 时崩溃
```
**修复**: 拆分为两个 if 判断：
```go
plan, err := h.trainingService.GetPlan(c.Request.Context(), id)
if err != nil {
    response.NotFound(c, "plan not found")
    return
}
if plan.UserID != oid {
    response.Forbidden(c, "no access")
    return
}
```

### 1.2 nil 指针解引用 — UpdateTask 中相同模式

**文件**: `api-server/internal/handler/training_handler.go:241`
**问题**: 与 1.1 完全相同的模式，`GetPlan` 返回错误时 `plan` 为 nil。
**代码**:
```go
plan, err := h.trainingService.GetPlan(c.Request.Context(), planID)
if err != nil || plan.UserID != oid {
```
**修复**: 同上，拆分为两个 if 判断。

---

## 二、HIGH 级问题

### 2.1 正则转义函数失效 — 搜索关键词绕过 ReDoS 防护

**文件**: `api-server/internal/repository/checkin_repo.go:20`
**问题**: `sanitizeRegex` 中使用 `\$&` 作为替换字符串，但 Go 的 `regexp.ReplaceAllString` 中 `\$` 被解释为字面量 `$`，`&` 被解释为字面量 `&`。实际替换结果是字面量 `$&`，而不是在特殊字符前加反斜杠。这导致 MongoDB $regex 查询中的关键词未被正确转义，存在 ReDoS 攻击风险。
**代码**:
```go
func sanitizeRegex(s string) string {
    return regexSpecialChars.ReplaceAllString(s, `\$&`)  // 错误：替换为字面量 "$&"
}
```
**修复**: 应使用 `\\$&`（在原始字符串中为 `` `\\$&` ``），其中 `\\` 生成字面量反斜杠，`$&` 为匹配文本：
```go
func sanitizeRegex(s string) string {
    return regexSpecialChars.ReplaceAllString(s, `\$0`)  // 方式1: \$0 美元符+匹配组0
    // 或
    return regexSpecialChars.ReplaceAllString(s, `${0}`) // 方式2: 显式组引用
}
```
**注意**: 实际上由于 Go 替换规则，正确转义应使用 `\\$&` 或 `$0`。推荐使用 `$0` 避免转义混淆。

### 2.2 JWT 算法未显式限制 — 可能接受 alg: none 攻击

**文件**: `api-server/pkg/jwt/jwt.go:40`
**问题**: `ParseWithClaims` 未使用 `WithValidMethods` 显式限制签名算法。虽然 `golang-jwt/jwt/v5` 默认行为有所改进，但未显式指定算法白名单仍存在风险。
**代码**:
```go
token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
    return j.secret, nil
})
```
**修复**: 添加 `jwt.WithValidMethods` 选项：
```go
token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
    return j.secret, nil
}, jwt.WithValidMethods([]string{"HS256"}))
```

### 2.3 MongoDB 断开连接无超时 — 优雅关闭可能永久阻塞

**文件**: `api-server/cmd/main.go:210`
**问题**: `mongoClient.Disconnect(context.Background())` 使用不可取消的 Background 上下文。如果 MongoDB 连接异常，Disconnect 可能永久阻塞，导致进程无法退出。
**修复**: 为 Disconnect 添加超时上下文：
```go
disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 10*time.Second)
defer disconnectCancel()
if err := mongoClient.Disconnect(disconnectCtx); err != nil {
    log.Printf("mongo disconnect error: %v", err)
}
```

### 2.4 配置缺少验证 — 缺失关键配置项时无提示

**文件**: `api-server/internal/config/config.go:62`
**问题**: `Load()` 函数读取配置文件后直接 Unmarshal，未验证 MongoDB URI、JWT Secret 等关键配置项是否为空。缺失配置会导致下游出现难以诊断的运行时错误。
**修复**: 在 Unmarshal 后添加配置验证逻辑：
```go
if cfg.Mongo.URI == "" {
    log.Fatal("mongo.uri is required")
}
if cfg.JWT.Secret == "" {
    log.Fatal("jwt.secret is required")
}
```

### 2.5 资源创建配额检查有 TOCTOU 竞态条件

**文件**: `api-server/internal/service/resource_service.go:37-56`
**问题**: 配额检查（检查/创建/再检查）存在 TOCTOU 竞态。在并发写入场景下，多个请求可能同时通过配额检查，导致最终存储超出配额。虽然代码有后置检查并尝试回滚，但回滚删除失败时不会阻断请求，导致超配额数据残留。
**修复**: 建议使用 MongoDB 事务或原子操作实现配额检查，或将配额逻辑移到数据库层使用 `$inc` 和条件判断。

---

## 三、MEDIUM 级问题

### 3.1 BatchIsLiked 错误被静默忽略

**文件**: `api-server/internal/handler/handler.go:158,198,286`
**问题**: `BatchIsLiked` 的返回值错误被 `_` 忽略（共 3 处）。当数据库查询失败时，函数返回 nil map，导致所有打卡条目显示为"未点赞"，且无任何日志记录。
**修复**: 记录错误日志，或至少记录一次警告：
```go
likedMap, err := h.socialService.BatchIsLiked(c.Request.Context(), checkinIDs, oid)
if err != nil {
    log.Printf("[WARN] batch is_liked failed: %v", err)
}
```

### 3.2 打卡列表查看计数错误被静默忽略

**文件**: `api-server/internal/handler/resource_handler.go:156`
**问题**: `IncrViewCount` 的错误被完全忽略，如果视图计数更新失败，用户无感知且无日志。
**修复**: 至少记录日志：
```go
if err := h.resourceService.IncrViewCount(c.Request.Context(), id); err != nil {
    log.Printf("[WARN] incr view count failed: %v", err)
}
```

### 3.3 微信 API 请求中 io.ReadAll 错误被忽略

**文件**: `api-server/pkg/wx/wx.go:46,103`
**问题**: `getAccessToken` 和 `SendSubscribeMessage` 中 `io.ReadAll(resp.Body)` 的返回值错误被 `_` 忽略。如果读取响应体失败，后续的 json.Unmarshal 会失败并返回错误，但此时错误信息不准确（"parse response failed" 而非 "read response failed"）。
**修复**: 检查并返回错误：
```go
body, err := io.ReadAll(resp.Body)
if err != nil {
    return "", fmt.Errorf("read response failed: %w", err)
}
```

### 3.4 微信 API 请求中 json.Marshal 错误被忽略

**文件**: `api-server/pkg/wx/wx.go:94`
**问题**: `json.Marshal(req)` 的返回错误被忽略。如果序列化失败，将发送空请求体。
**修复**: 检查并返回错误。

### 3.5 速率限制器 Mutex 非 defer 解锁 — panic 风险

**文件**: `api-server/internal/middleware/rate_limit.go:66-91`
**问题**: `Limit()` 函数中 `mu.Lock()` 和 `mu.Unlock()` 分布在多个条件分支中，没有使用 `defer`。如果中间发生 panic（例如 `c.AbortWithStatusJSON` 中），mutex 将永远锁定，导致后续请求死锁。
**修复**: 使用 `defer r.mu.Unlock()` 替代手动解锁，并将整个临界区逻辑提取到独立的函数中。

### 3.6 清理协程无法停止 — goroutine 泄漏

**文件**: `api-server/internal/middleware/rate_limit.go:47-58`
**问题**: `cleanupLoop` 在 `NewRateLimiter` 中启动后永远运行，没有停止机制。如果 `RateLimiter` 被 GC 回收，goroutine 泄漏。
**修复**: 添加 `ctx` 参数或 `stop` channel 以支持优雅停止。

### 3.7 训练计划 UpdatePlan 缺少服务层权限校验

**文件**: `api-server/internal/service/training_service.go:46-48`
**问题**: `UpdatePlan` 在服务层没有检查 `userID` 是否匹配计划的 `UserID`。权限校验仅在 handler 层完成（且 handler 层有 nil pointer bug）。如果未来新增路由跳过 handler 直接调用，权限检查被绕过。
**修复**: 在服务层添加所有权校验：
```go
func (s *TrainingService) UpdatePlan(ctx context.Context, id, userID primitive.ObjectID, update map[string]interface{}) error {
    plan, err := s.planRepo.FindByID(ctx, id)
    if err != nil {
        return err
    }
    if plan.UserID != userID {
        return fmt.Errorf("access denied")
    }
    return s.planRepo.Update(ctx, id, update)
}
```

### 3.8 通知设置 GetOrCreate 存在代码重复

**文件**: `api-server/internal/repository/notification_repo.go:110-132`
**问题**: 默认通知设置的结构体字面量（第 110-118 行）与 `$setOnInsert` 中的 map（第 122-131 行）定义了两遍相同的默认值。修改其中一个时容易忘记同步另一个，导致数据不一致。
**修复**: 提取公共默认值函数或常量。

### 3.9 游标遍历后未检查 cursor.Err()

**文件**: `api-server/internal/repository/social_repo.go:129-136`
**问题**: `BatchIsLiked` 在 `for cursor.Next(ctx)` 循环后没有检查 `cursor.Err()`。如果游标在遍历过程中遇到网络错误，错误会被静默忽略。
**修复**: 在循环后添加：
```go
if err := cursor.Err(); err != nil {
    return nil, err
}
```

### 3.10 配置加载时未找到配置文件仅打印日志

**文件**: `api-server/internal/config/config.go:57-58`
**问题**: `viper.ReadInConfig()` 失败时仅打印日志，不终止程序。如果配置文件确实需要存在（例如生产环境），程序将在缺失配置的情况下继续运行，可能导致诡异行为。
**修复**: 对于生产环境，考虑使用 `log.Fatalf` 或通过环境变量强制要求配置项。

---

## 四、LOW 级问题

### 4.1 魔法数字 10 和 20 未使用常量

**文件**: 多处 handler 文件
**问题**: 默认分页大小多处硬编码为 `10` 或 `20`，应使用 `constants.DefaultPageSize` 和 `constants.MaxPageSize`。
**涉及文件**:
- `api-server/internal/handler/handler.go:171,175,217,219,267,271,369,395`
- `api-server/internal/handler/training_handler.go:111,115,303`
- `api-server/internal/handler/insight_handler.go:109,115,133,138,249,251`
- `api-server/internal/handler/notification_handler.go:29,33`
- `api-server/internal/handler/resource_handler.go:175,179,301`

### 4.2 ObjectID 解析错误静默忽略

**文件**: 多处 handler 文件
**问题**: `ObjectIDFromHex` 的错误被静默忽略，无效的 ID 被当作空值处理。用户发送错误 ID 格式时得不到任何反馈。
**涉及文件**:
- `api-server/internal/handler/training_handler.go:65-67` (GroupID)
- `api-server/internal/handler/insight_handler.go:56-64` (CheckinID, PlanID)
- `api-server/internal/handler/resource_handler.go:118-121` (GroupID)
- `api-server/internal/handler/handler.go:180-184` (GroupID)

### 4.3 路径参数整数转换错误被忽略

**文件**: `api-server/internal/handler/training_handler.go:246-247`
**问题**: `dayIndex` 和 `taskIndex` 从字符串转换，错误被忽略。非数字参数会被转换为 0。虽然后续的边界检查可能会捕获到问题，但更合理的做法是返回错误信息给用户。
**修复**: 检查转换错误并返回 400。

### 4.4 time.Now().Sub 已弃用

**文件**: `api-server/internal/service/training_service.go:177,181`
**问题**: `time.Now().Sub(plan.StartDate)` 和 `plan.EndDate.Sub(plan.StartDate)` 应使用 `time.Since(plan.StartDate)` 和 `plan.EndDate.Sub(plan.StartDate)`（后者正确，但前者应改用 `time.Since`）。
**修复**: `time.Now().Sub(plan.StartDate)` → `time.Since(plan.StartDate)`

### 4.5 创建资源时 type assertion 可能 panic

**文件**: 多个 `*_repo.go` 文件中的 `Create` 方法
**问题**: `result.InsertedID.(primitive.ObjectID)` 类型断言未使用 `ok` 模式。如果 MongoDB driver 行为异常，会 panic。虽然实际中极少发生，但防御性编程更安全。
**涉及文件**: `checkin_repo.go:60`, `insight_repo.go:30`, `training_repo.go:30`, `user_repo.go:30`, `resource_repo.go:31`, `template_repo.go:29`, `notification_repo.go:29`

### 4.6 排行榜刷新内部超时覆盖外部上下文

**文件**: `api-server/internal/repository/rank_repo.go:89`
**问题**: `RefreshRank` 内部创建 `context.WithTimeout(ctx, 30*time.Second)` 覆盖了外部传入的上下文。如果外部上下文已有超时，内部超时可能更长或更短，导致行为不一致。
**修复**: 使用 `context.WithTimeoutCause` 或添加注释说明设计意图。

### 4.7 短信验证码未实现

**文件**: `api-server/internal/service/notification_service.go`
**问题**: 虽然 `SendBatch` 方法存在，但整个项目中的错误处理在某些地方使用 `fmt.Errorf` 包装错误，在另一些地方仅记录日志，风格不一致。

### 4.8 社交服务中的 session 启动源不一致

**文件**: `api-server/internal/service/service.go:308`
**问题**: `AddComment` 中使用 `s.likeRepo.StartSession()` 启动 MongoDB 会话，逻辑上应使用 `s.commentRepo.StartSession()`。虽然功能上相同（因为来自同一个数据库），但语义不当。

### 4.9 日志同步错误未处理

**文件**: `api-server/cmd/main.go:44`
**问题**: `defer logger.Sync()` 的返回值未检查。zap 的 Sync 在写入文件时可能失败。虽然这在 Go 中是一个常见模式，但忽略错误可能导致日志丢失而不被察觉。
**修复**: 考虑在程序退出时检查错误：
```go
if err := logger.Sync(); err != nil {
    log.Printf("failed to sync logger: %v", err)
}
```

### 4.10 随机字符串生成错误被忽略

**文件**: `api-server/internal/middleware/cors.go:52`
**问题**: `rand.Int(rand.Reader, ...)` 的错误被忽略。虽然 `rand.Reader` 几乎不会失败，但理论上有可能。

### 4.11 训练提醒发现 active plans 的 cursor 缺少 Close

**文件**: `api-server/internal/repository/training_repo.go:153-155`
**问题**: `FindActive` 返回 `mongo.Cursor` 给调用方，但调用方 `cron.go:90` 调用了 `defer cursor.Close(ctx)`，所以 OK。但 `FindActive` 方法本身不关闭 cursor，而是交给调用方，这是一种约定，需要在文档中注明。

---

## 五、代码组织与架构建议

### 5.1 分页逻辑重复

**问题**: 每个 handler 中分页参数解析代码高度重复（`page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))` 等）。
**建议**: 抽取公共分页参数解析函数：
```go
func parsePagination(c *gin.Context, defaultSize int) (page, pageSize int) {
    page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(defaultSize)))
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > constants.MaxPageSize { pageSize = defaultSize }
    return
}
```

### 5.2 游标遍历模式重复

**问题**: 多个 repository 中 `Find` → `cursor.Close` → `cursor.All` 模式高度重复。
**建议**: 抽取通用辅助函数：
```go
func findAndDecode[T any](ctx context.Context, coll *mongo.Collection, filter bson.M, opts ...*options.FindOptions) ([]*T, error) {
    cursor, err := coll.Find(ctx, filter, opts...)
    if err != nil { return nil, err }
    defer cursor.Close(ctx)
    var result []*T
    if err := cursor.All(ctx, &result); err != nil { return nil, err }
    return result, nil
}
```

### 5.3 Service 层权限校验不统一

**问题**: 部分 service 方法校验了所有权（如 `InsightService.Update`、`ResourceService.Update`），部分未校验（如 `TrainingService.UpdatePlan`），还有部分依赖 handler 层传入 `userID` 到 repo 的 filter 中（如 `Delete` 方法）。风格不统一。
**建议**: 统一约定：所有修改操作在 service 层校验所有权，handler 层只做参数解析和响应。

---

## 六、总结

| 级别 | 数量 | 关键问题 |
|------|------|----------|
| CRITICAL | 2 | nil 指针解引用（training_handler.go 第 160 行和第 241 行） |
| HIGH | 5 | 正则转义失效、JWT 算法未限制、MongoDB 断开阻塞、配置无验证、配额竞态 |
| MEDIUM | 10 | 错误静默忽略、mutex 使用风险、goroutine 泄漏、权限校验缺失等 |
| LOW | 11 | 魔法数字、类型断言、弃用 API、风格不一致等 |

**最紧急的修复建议**：
1. 立即修复 `training_handler.go:160` 和 `training_handler.go:241` 的 nil 指针解引用（CRITICAL）
2. 修复 `checkin_repo.go:20` 的 `sanitizeRegex` 函数（HIGH，安全漏洞）
3. 为 `pkg/jwt/jwt.go:40` 添加 `WithValidMethods`（HIGH，安全漏洞）