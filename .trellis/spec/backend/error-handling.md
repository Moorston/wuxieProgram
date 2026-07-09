# 日志规范

> 武俱打卡项目的结构化日志约定。

---

## 概览

项目使用 [zap](https://github.com/uber-go/zap) 作为日志库（`go.uber.org/zap`），提供高性能的结构化日志。日志格式为 JSON，便于日志聚合系统（如 ELK、Loki）解析。

---

## 日志初始化

```go
// api-server/cmd/main.go:40-43
logger, err := zap.NewProduction()
if err != nil {
    log.Fatalf("failed to init logger: %v", err)
}
defer logger.Sync()
```

**日志轮转**：当前使用 `zap.NewProduction()` 直接输出到 stdout（Docker 容器的标准做法）。生产环境建议添加 lumberjack 日志轮转：

```go
// 建议的日志轮转配置（可选）
hook := &lumberjack.Logger{
    Filename:   "logs/api-server.log",
    MaxSize:    100,      // MB
    MaxBackups: 7,
    MaxAge:     30,       // 天
    Compress:   true,
}
core := zapcore.NewCore(
    zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
    zapcore.AddSync(hook),
    zap.InfoLevel,
)
logger := zap.New(core)
```

---

## 日志级别

| 级别 | 使用场景 | 示例 |
|------|---------|------|
| `Debug` | 开发调试信息，生产环境不应启用 | SQL 语句、请求体 |
| `Info` | 正常业务流程的关键节点 | HTTP 请求记录、定时任务开始/完成 |
| `Warn` | 预期内的异常但不影响服务 | 索引创建失败、微信模板消息发送失败 |
| `Error` | 非预期错误需要人工关注 | 数据库连接失败、业务操作失败 |
| `Fatal` | 服务无法继续运行 | JWT 密钥长度不足（`log.Fatal` 在 main.go 中使用） |

---

## 结构化日志字段

### 请求日志（middleware/logger.go）

请求日志中间件应记录以下字段：

| 字段 | 说明 | 示例 |
|------|------|------|
| `method` | HTTP 方法 | `GET`, `POST` |
| `path` | 请求路径 | `/api/checkin/list` |
| `status` | HTTP 状态码 | `200`, `400` |
| `latency` | 请求耗时（毫秒） | `42` |
| `ip` | 客户端 IP | `192.168.1.1` |
| `request_id` | 请求唯一 ID | `req-abc123` |

### 业务日志

业务日志应包含以下字段：

| 字段 | 说明 | 必填 |
|------|------|------|
| `error` | 错误原因 | Error 级别必填 |
| `user_id` | 操作者用户 ID | 鉴权请求推荐 |
| `module` | 所属模块 | 推荐 |
| `request_id` | 关联的请求 ID | 推荐（用于链路追踪） |

### 示例

```go
// ✅ 正确的业务日志
logger.Error("operation failed",
    zap.String("module", "checkin"),
    zap.String("user_id", userID),
    zap.String("request_id", c.GetString("request_id")),
    zap.Error(err),
)

// ❌ 禁止：未使用结构化字段
logger.Error("operation failed: " + err.Error())
```

---

## 日志中间件（middleware/logger.go）

```go
// 请求日志中间件记录每次请求的方法、路径、状态码、耗时
r.Use(middleware.Logger(logger))
```

日志中间件应：

- **记录**：请求方法、路径、状态码、耗时、客户端 IP、请求 ID
- **过滤**：Authorization 头（`[REDACTED]` 替换原始 token）
- **不记录**：请求体中的密码、token、敏感个人信息

---

## 项目中的日志使用问题

### 一致性问题

项目中混合使用了多种日志方式，需要统一：

| 方式 | 位置 | 问题 |
|------|------|------|
| `zap.Logger` | `middleware/logger.go` | ✅ 正确方式 |
| `log.Printf("[ERROR] ...")` | `handler/handler.go` 多处 | ❌ 非结构化，无法日志聚合 |
| `log.Println(...)` | `cmd/main.go`, `service/training_service.go` | ❌ 非结构化 |
| `log.Printf("[WARN] ...")` | `service/training_service.go:203` | ❌ 非结构化 |

### 敏感信息

日志中不应记录：

- ❌ JWT Token（Authorization 头应脱敏为 `[REDACTED]`）
- ❌ 微信接口的 secret/app_id
- ❌ 数据库连接字符串中的密码
- ❌ 用户的完整手机号、身份证号等 PII

---

## 日志规则总结

| 规则 | 说明 |
|------|------|
| **使用结构化日志** | 所有日志使用 zap 的字段方法（`zap.String`, `zap.Error`），禁止字符串拼接 |
| **错误必须记录** | 所有 Error 级别的错误必须记录到日志 |
| **敏感信息脱敏** | Authorization header 必须过滤 |
| **禁止日志 fatal** | `log.Fatal` / `zap.Fatal` 只能在 `cmd/main.go` 中使用 |
| **日志级别得当** | 正常的业务操作使用 Info，预期异常使用 Warn，非预期错误使用 Error |

---

## 常见错误

1. **混合使用 `log.Printf` 和 zap**：新代码应统一使用通过依赖注入传入的 zap logger
2. **忘记调用 `logger.Sync()`**：`main.go` 中必须有 `defer logger.Sync()` 确保日志刷新
3. **Error 级别滥用**：参数验证失败应返回 400 给客户端，但日志级别应为 Warn 而非 Error
4. **未记录请求 ID**：业务日志中应携带 `request_id` 以支持多步骤操作的链路追踪