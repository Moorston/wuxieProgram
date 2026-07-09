# 后端代码质量标准

> 武俱打卡项目后端的代码质量约束、禁止模式和必需模式。

---

## 概览

本项目使用 Go 1.21+，遵循 [Go 官方编码规范](https://go.dev/doc/effective_go)。以下规范基于代码审查报告（`CODE_REVIEW_REPORT.md`）中发现的真实问题提炼而成。

---

## 禁止模式

### 🔴 P0 - 安全类

| # | 禁止模式 | 风险 | 真实案例 |
|---|---------|------|---------|
| 1 | **HTTP 请求无超时** | 请求挂起耗尽 goroutine | `service/service.go:66` 原代码使用 `http.Get()` 无超时 → 已修复为 `&http.Client{Timeout: 10 * time.Second}` |
| 2 | **内部接口无认证** | 未授权调用可伪造数据 | `router.go:117-120` 原 `/api/internal/transcode/done` 无认证 → 已修复添加 `InternalAuth` 中间件 |
| 3 | **JWT 密钥长度不足** | 可暴力破解伪造 token | `config.yaml:18` 默认密钥仅 30 字符 → 修复后强制 ≥32 字符 |
| 4 | **用户输入直接构造正则** | ReDoS 攻击 / MongoDB 性能耗尽 | `checkin_repo.go:151-159` 使用 `sanitizeRegex()` 但未限制 keyword 长度 |
| 5 | **错误信息返回给客户端** | 信息泄露辅助攻击 | `handler.go` 多处使用 `response.InternalError(c, err.Error())` 暴露内部错误 |
| 6 | **日志记录敏感信息** | token/密钥泄露 | `middleware/logger.go` 应过滤 `Authorization` 头 |
| 7 | **文件上传无类型白名单** | 可上传任意文件 | `resource_handler.go` 预签名接口未验证文件扩展名 |

### 🟠 P1 - 数据一致性类

| # | 禁止模式 | 风险 | 真实案例 |
|---|---------|------|---------|
| 8 | **跨集合写操作无事务** | 数据不一致（如点赞计数错误） | `service.go:243-272` 点赞和计数更新原无事务 → 已修复 |
| 9 | **删除资源前无所有权检查** | 越权删除 | `resource_handler.go` 的 `Delete()`、`DELETE /api/insight/:id` |
| 10 | **评论/描述无长度限制** | 数据库性能问题 | `handler.go:305-332` `CommentReq` 原无 `max` 标签 → 已修复 |
| 11 | **敏感配置明文存储** | 密钥泄露 | `config.yaml` 含 JWT secret, wx secret, media_secret |

### 🟡 P2 - 代码质量类

| # | 禁止模式 | 风险 | 真实案例 |
|---|---------|------|---------|
| 12 | **魔法数字直接出现在代码中** | 难以维护和修改 | `handler.go:175` `pageSize > 50`、`service.go:133` `Score: 10` → 已部分修复为 constants 包 |
| 13 | **Repository 层无接口定义** | 无法 mock 进行单元测试 | 所有 repo 都无接口（如 `UserRepo` 直接使用结构体） |
| 14 | **单文件包含多个不相关类型** | 代码组织混乱 | `handler.go` 414 行含 6 个 handler，`service.go` 464 行含 6 个 service |
| 15 | **error 信息用 `fmt.Errorf` 拼接而非 `%w`** | 无法用 `errors.Is()` 判断错误链 | 部分 repo 方法返回 "not found" 字符串而非哨兵错误 |
| 16 | **`log.Printf` 替代结构化日志** | 无法进行日志聚合分析 | `handler.go` 多处使用 `log.Printf("[ERROR] ...")` 而非 zap |

---

## 必需模式

### P0 - 必须遵守

1. **所有 HTTP 客户端调用必须设置超时**
   ```go
   // ✅ 正确
   client := &http.Client{Timeout: 10 * time.Second}
   resp, err := client.Do(req)
   ```
2. **所有内部 API 端点必须鉴权**
   ```go
   // ✅ 正确：使用 InternalAuth 中间件
   internal.Use(middleware.InternalAuth(cfg.MediaSecret))
   ```
3. **删除操作必须验证所有权**
   ```go
   // ✅ 正确：查询+验证后再删除
   resource, err := h.resourceService.GetByID(ctx, resourceID)
   if resource.UserID != oid { ... }
   ```
4. **用户输入必须验证长度并转义**
   ```go
   // ✅ 正确：Gin binding tags
   type CommentReq struct {
       Content string `json:"content" binding:"required,min=1,max=500"`
   }
   ```
5. **Search 关键词必须限制长度**
   ```go
   // ✅ 正确
   if len(keyword) > constants.MaxSearchKeyword {
       keyword = keyword[:constants.MaxSearchKeyword]
   }
   ```
6. **错误信息返回通用文案，详细错误仅记录日志**
   ```go
   // ✅ 正确
   logger.Error("operation failed", zap.Error(err))
   response.InternalError(c, "操作失败，请稍后重试")
   ```

### P1 - 强烈建议

7. **跨集合写操作使用 MongoDB 事务**
   ```go
   // ✅ 正确：如 ToggleLike 使用 WithTransaction
   session, err := s.likeRepo.StartSession()
   _, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
       liked, err := s.likeRepo.ToggleWithSession(sessCtx, ...)
       // ...
   })
   ```
8. **分页参数复用全局常量**
   ```go
   // ✅ 正确
   if pageSize < 1 || pageSize > constants.MaxPageSize {
       pageSize = constants.DefaultPageSize
   }
   ```
9. **第 0 页应按第 1 页处理**
   ```go
   // ✅ 正确
   if page < 1 { page = 1 }
   ```

### P2 - 建议

10. **Repository 层定义接口**
   ```go
   type UserRepository interface {
       FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
       FindByOpenID(ctx context.Context, openid string) (*model.User, error)
       Create(ctx context.Context, user *model.User) error
       Update(ctx context.Context, id primitive.ObjectID, update map[string]interface{}) error
   }
   ```
11. **使用 `%w` 包装错误以支持错误链**
   ```go
   // ✅ 正确
   return nil, fmt.Errorf("find user failed: %w", err)
   // ❌ 错误：丢失错误链
   return nil, fmt.Errorf("find user failed: %v", err)
   ```
12. **所有公共函数必须有文档注释**

---

## 禁止依赖

| 依赖 | 原因 | 替代方案 |
|------|------|---------|
| `database/sql` | 项目使用 MongoDB，非关系型数据库 | `go.mongodb.org/mongo-driver/mongo` |
| `encoding/xml` | 项目 API 仅使用 JSON | `encoding/json` |
| `gorm` / ORM | MongoDB 不适合 ORM | 直接使用 mongo driver |

---

## 代码格式要求

- 所有 `.go` 文件必须通过 `go fmt` 格式化
  ```bash
  cd api-server && go fmt ./...
  ```
- 不允许有未使用的导入（CI 中运行 `go vet` 检查）
- 不允许有未处理的错误（使用 `_` 忽略错误必须加注释说明原因）

---

## 测试要求

- **单元测试覆盖率目标**：核心业务逻辑 ≥ 60%（Service 层）
- **优先测试模块**:
  1. `pkg/jwt/jwt.go` — JWT 生成和验证
  2. `internal/service/service.go` — 核心业务（Auth/Checkin/Social）
  3. `internal/middleware/auth.go` — 认证中间件
- **目前状态**: 项目中尚无 `*_test.go` 文件，测试覆盖率 ≈ 0%

### 测试模式约定

```go
// service_test.go — 使用 mock repository 测试业务逻辑
func TestSocialService_ToggleLike(t *testing.T) {
    // 使用 testify/mock 或实现内存版 repository
    mockLikeRepo := new(MockLikeRepo)
    mockCheckinRepo := new(MockCheckinRepo)
    svc := NewSocialService(mockLikeRepo, mockCheckinRepo, ...)
    
    liked, err := svc.ToggleLike(ctx, checkinID, userID)
    assert.NoError(t, err)
    assert.True(t, liked)
}
```

---

## 代码审查清单

审查每个 PR 时检查以下项目：

### 安全
- [ ] 新的 HTTP 客户端调用是否设置了超时？
- [ ] 新的 API 端点是否添加了正确的鉴权中间件？
- [ ] 用户输入是否有长度/类型/内容验证？
- [ ] 删除/修改操作是否验证了资源所有权？
- [ ] 错误是否返回了通用信息而非内部细节？

### 质量
- [ ] 是否有魔法数字应抽取为常量？
- [ ] 跨集合操作是否使用了事务？
- [ ] 函数是否有文档注释？
- [ ] 错误是否正确使用 `%w` 包装？
- [ ] 文件是否通过 `go fmt` 格式化？
- [ ] 是否有未使用的导入或变量？

### 性能
- [ ] 是否有 N+1 查询模式？
- [ ] 新查询是否添加了必要的数据库索引？
- [ ] 分页查询是否限制了最大 page_size？