# 安全约束

> 武俱打卡项目的安全规范和各层安全约束。

---

## 安全架构总览

```
┌───────────────────────────────────────────────┐
│ 安全层次                                         │
│                                                  │
│ 1. 传输层: HTTPS（生产环境）、Nginx TLS 终止        │
│ 2. 认证层: JWT（用户认证）+ InternalAuth（内部服务） │
│ 3. 授权层: 资源所有权验证                           │
│ 4. 输入层: 参数验证 + 输入清洗 + 长度限制           │
│ 5. 存储层: MinIO 预签名 URL（有时效性）             │
│ 6. 审计层: 结构化日志（脱敏敏感信息）                │
└───────────────────────────────────────────────┘
```

---

## 认证约束

### A1: 用户认证 — JWT

- 所有业务 API 必须经过 `middleware.Auth(jwtMgr)` 中间件
- JWT 密钥长度必须 ≥ 32 字符
- Token 有效期 72 小时（配置在 `config.yaml`）
- Token 在 `Authorization: Bearer <token>` 头中传递

```go
// middleware/auth.go — 认证流程
func Auth(jwtMgr *jwt.JWTManager) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            response.Unauthorized(c, "missing authorization header")
            c.Abort()
            return
        }
        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := jwtMgr.Parse(tokenStr)
        if err != nil {
            response.Unauthorized(c, "invalid token")
            c.Abort()
            return
        }
        // 验证 user_id 是合法的 ObjectID
        if _, err := primitive.ObjectIDFromHex(claims.UserID); err != nil {
            response.Unauthorized(c, "invalid user identity in token")
            c.Abort()
            return
        }
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

**例外**：仅 `/api/auth/login` 和 `/health` 不需要认证

### A2: 内部服务认证 — InternalAuth

- api-server 和 media-server 之间的回调必须携带 `X-Internal-Secret` 请求头
- 共享密钥在两个服务的配置文件中配置，必须一致
- 密钥默认值为 `"wuxie-media-secret-change-in-production"`，生产环境必须修改

```go
// middleware/internal_auth.go — 内部 API 认证
func InternalAuth(apiSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        secret := c.GetHeader("X-Internal-Secret")
        if secret != apiSecret {
            c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
            return
        }
        c.Next()
    }
}
```

---

## 授权约束

### Z1: 资源所有权验证

以下操作必须验证当前用户是资源的所有者：

| 操作 | 端点 | 当前状态 |
|------|------|---------|
| 删除打卡 | `DELETE /api/checkin/:id` | ✅ 已实现 |
| 删除训练计划 | `DELETE /api/training/plan/:id` | ✅ 已实现 |
| 删除感悟 | `DELETE /api/insight/:id` | ⚠️ 需确认 |
| 删除资料库资源 | `DELETE /api/resource/:id` | ⚠️ 需确认 |
| 更新资料库资源 | `PUT /api/resource/:id` | ⚠️ 需确认 |
| 更新感悟 | `PUT /api/insight/:id` | ⚠️ 需确认 |

**实现模式**：

```go
// ✅ 正确的所有权验证
func (h *ResourceHandler) Delete(c *gin.Context) {
    userID := c.GetString("user_id")
    oid, _ := primitive.ObjectIDFromHex(userID)

    resource, err := h.resourceService.GetByID(c.Request.Context(), resourceID)
    if resource.UserID != oid {
        response.Forbidden(c, "permission denied")
        return
    }
    // ...
}
```

### Z2: 可见性控制

以下资源有可见性范围，读取时需要验证：

| 资源 | 可见性控制 |
|------|-----------|
| 感悟笔记 | `visibility` 字段：私密/公开，公开的才在广场展示 |
| 资料库 | `share_scope` 字段：仅自己/考核组/公开 |

---

## 输入验证约束

### V1: 参数长度限制

| 字段 | 最大长度 | 验证位置 |
|------|---------|---------|
| 打卡描述 | 200 字 | Handler binding tag |
| 评论内容 | 500 字 | `binding:"required,min=1,max=500"` |
| 搜索关键词 | 50 字符 | `checkin_repo.go` 中截断 |
| 感悟内容 | 2000 字 | Model 约束 |

### V2: 文件类型白名单

上传预签名时验证文件扩展名（待完全实现）：

| 模块 | 允许类型 | 验证方式 |
|------|---------|---------|
| 资源库上传 | .mp4, .jpg, .jpeg, .png, .pdf, .doc, .docx | ext 参数白名单校验 |
| 打卡视频 | 视频格式 | → 类型验证 |

### V3: 搜索关键词安全

```go
// ✅ 搜索关键词安全处理
// 1. 限制长度
if len(keyword) > constants.MaxSearchKeyword {
    keyword = keyword[:constants.MaxSearchKeyword]
}
// 2. 正则特殊字符转义（防止 ReDoS）
var regexSpecialChars = regexp.MustCompile(`[[\]{}()*+?.\\^$|]`)
keyword = regexSpecialChars.ReplaceAllString(keyword, `\$&`)
```

---

## 数据安全约束

### D1: 错误信息

- **日志**：记录完整错误详情（包含 stack trace）
- **API 响应**：返回通用信息，不暴露内部细节

```go
// ✅ 正确
logger.Error("database query failed", zap.Error(err))
response.InternalError(c, "操作失败，请稍后重试")

// ❌ 禁止
response.InternalError(c, err.Error())
response.InternalError(c, "mongo: no documents in result set")  // 暴露数据库结构
```

### D2: 预签名 URL 时效

| 资源 | 有效期 | 风险 |
|------|--------|------|
| 视频上传预签名 URL | 5 分钟 | 过期无法上传 |
| 视频播放预签名 URL | 2 小时 | 过期需重新获取 |

### D3: 敏感配置

- 禁止将配置文件中的默认密钥用于生产环境
- JWT 密钥必须 ≥ 32 字符（`cmd/main.go` 中强制检查）
- 微信 secret 和 media_secret 在生产环境应通过环境变量注入

---

## 防护措施约束

### R1: 速率限制（待完全实现）

| 接口 | 限制 | 状态 |
|------|------|------|
| `/api/auth/login` | 5 次/分钟/IP | ⚠️ 中间件已创建，尚未集成到路由 |
| 全局 | 100 次/分钟/IP | ⚠️ 中间件已创建，尚未集成到路由 |
| 打卡准备 | 待定 | ❌ 未实现 |

### R2: CORS 配置

```go
// ✅ 当前 CORS 配置
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        // 对于客户端（小程序/App）CORS 不适用，仅为开发环境 H5 配置
        // 生产环境应限制白名单
        c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

### R3: Panic 恢复

- 所有路由使用 `gin.Recovery()` 中间件防止服务因 panic 崩溃
- Media Server 同样使用 Recovery 中间件

---

## 安全审查清单

每个 PR 必须检查：

- [ ] 新 API 端点是否添加了正确的认证中间件？
- [ ] 删除/更新操作是否验证了资源所有权？
- [ ] 用户输入是否有长度限制和内容验证？
- [ ] 预签名 URL 是否有合理的过期时间？
- [ ] 错误信息是否返回通用信息而非内部细节？
- [ ] CORS 配置是否合理（开发环境 vs 生产环境）？
- [ ] 敏感配置是否已从代码中剥离？