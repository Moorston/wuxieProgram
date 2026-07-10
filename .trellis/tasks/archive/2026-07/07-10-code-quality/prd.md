# 代码质量提升

## Goal

统一日志标准化、错误处理规范化、消除魔法字符串三大代码质量问题。

## Background

### 发现的代码质量问题

#### Q1: 日志混用（严重）
- **Handlers**（9 个文件, ~40 处）: 使用 `log.Printf("[ERROR]...")` 
- **Cron**（1 个文件, ~15 处）: 使用 `log.Println/Printf` 加 `[cron]` 前缀
- **Service**（3 个文件, ~10 处）: 使用 `log.Printf` 加 `[WARN]` 前缀
- **AuthService**（1 个文件）: 使用 `zap.Logger` ✅
- **app.go**（1 个文件, 2 处）: 使用 `log.Fatal/Printf`

**问题**: 项目已经引入了 `zap.Logger` 并传递到各个 Handler，但大多数 Handler 仍使用标准库 `log`。日志格式不统一，无法进行结构化查询。

#### Q2: 错误处理未规范化（中）
- 所有错误使用 `fmt.Errorf` 构建，无自定义错误类型
- 错误消息散落在各个文件中，无统一常量
- 常见错误类型："not found"、"access denied"、"account suspended"、"failed to..." 等

#### Q3: 魔法字符串（低）
- 如 `user.Status == 1`（已改为 `user.IsBanned()` ✅）
- 但许多地方仍有魔法数字（如分页大小 `50`、超时 `30s` 等）

## Requirements

### R1: Handler 日志标准化
将 9 个 Handler 的 `log.Printf` 替换为 `zap.Logger`
- checkin_handler.go（~10 处）
- insight_handler.go（~10 处）
- resource_handler.go（~10 处）
- training_handler.go（~10 处）
- notification_handler.go（~7 处）
- social_handler.go（~3 处）
- user_handler.go（~1 处）
- rank_handler.go（~1 处）
- group_handler.go（~1 处）

### R2: Service 日志标准化
将 Service 中的 `log.Printf` 替换为 `zap.Logger`
- training_service.go（~1 处）
- resource_service.go（~5 处）
- insight_service.go（~4 处）

### R3: Cron 日志标准化
- 将 `cron.go` 中的 `log.Println/Printf`（~15 处）替换为 `zap.Logger`
- 移除 `[cron]` 前缀，改用结构化字段 `zap.String("component", "cron")`

### R4: 错误类型定义
- 定义常见错误类型：`ErrNotFound`、`ErrAccessDenied`、`ErrAccountSuspended`
- 在 Service 层使用标准错误类型，便于调用方用 `errors.Is` 判断

### R5: 常量替换魔法数字
- `constants/constants.go` 已存在，但在代码中未完全使用
- 替换零散魔法数字：分页大小、超时时间、状态值等

## Acceptance Criteria

- [ ] 9 个 Handler 不再使用 `log.Printf`
- [ ] `cron.go` 使用 `zap.Logger` 结构化日志
- [ ] 日志包含 `component`、`error`、`user_id` 等结构化字段
- [ ] 自定义错误类型定义在 `pkg/errors` 包
- [ ] `go build ./...` 编译通过
- [ ] `constants` 包在代码中实际使用

## Out of Scope

- 客户端日志标准化
- 添加新的日志级别（debug、trace）
- 日志轮转配置
- 错误码体系（如 HTTP 错误码映射）

## Open Questions

1. Handler 的 zap.Logger 从哪里注入？目前 Handler 构造函数已接受 logger，但部分 Handler 未使用