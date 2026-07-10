# 错误处理统一化

## Goal

统一 Service → Handler 的错误处理模式，使用标准错误类型替代字符串字面量。

## Problem Analysis

### 当前状态
- `pkg/errors/errors.go` 定义了 14 个错误常量（已存在但**未使用**）
- Service 层返回 `fmt.Errorf("...")` 字符串
- Handler 层用字符串字面量匹配错误：`"internal server error"`、`"invalid params"` 等
- 无法用 `errors.Is()` 做类型安全的错误判断

### 示例（当前问题）
```go
// service: 返回字符串
return nil, fmt.Errorf("account suspended")

// handler: 无法区分错误类型，统一返回 500
if err != nil {
    response.InternalError(c, "internal server error")  // 丢失了原始错误信息
}
```

### 目标状态
```go
// service: 返回标准错误
return nil, errors.ErrAccountSuspended

// handler: 类型安全的错误映射
if err != nil {
    respondWithError(c, err)  // ErrAccountSuspended → 403
}
```

## Requirements

### R1: Service 层使用标准错误
- 所有 Service 方法返回 `pkg/errors` 中定义的错误
- 需要包装时使用 `fmt.Errorf("%w", errors.ErrXxx)`

### R2: Handler 错误映射函数
- 创建 `respondWithError(c, err)` 辅助函数
- 使用 `errors.Is()` 匹配错误类型
- 映射规则：
  - `ErrNotFound/ErrUserNotFound/ErrCheckinNotFound/...` → 404
  - `ErrAccessDenied/ErrNotCheckinOwner/...` → 403
  - `ErrAccountSuspended` → 403
  - `ErrInvalidParams` → 400
  - `ErrInvalidToken/...` → 401
  - 其他 → 500 (internal server error)

### R3: 清理 Handler 中的字符串字面量
- 用 `respondWithError(c, err)` 替代手动匹配

## Acceptance Criteria

- [ ] Service 层 80%+ 使用 `pkg/errors` 错误类型
- [ ] `respondWithError` 函数覆盖所有错误类型
- [ ] Handler 中不再有 `"internal server error"` 字符串字面量
- [ ] `go build ./...` 编译通过

## Out of Scope

- 错误码体系（如 10001、10002）
- 国际化错误消息
- 错误日志格式统一