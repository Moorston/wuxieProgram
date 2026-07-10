# 架构优化测试

## Goal

为刚实施的 5 项架构优化编写测试，确保重构不引入回归。

## Background

### 刚完成的架构优化
1. **Composition Root** — `internal/app/app.go` 提取，`main.go` 缩减到 45 行
2. **Repository 接口** — 8 个接口文件，覆盖 14 个结构体、64 个方法
3. **客户端 API 拆分** — `api/index.ts` 拆分为 7 个领域模块 + barrel re-export
4. **领域模型方法** — User (3 方法 + 2 常量) + Checkin (5 方法)
5. **重试工具** — `pkg/retry/retry.go` 泛型重试，支持 injectable IsRetryable

### 当前测试状态
- JWT 包：10 个测试用例 ✅
- TokenBlacklist：8 个测试用例 ✅
- Auth 中间件：11 个测试用例 ✅
- CORS 中间件：8 个测试用例 ✅
- Auth Service：9 个测试用例 ✅
- 新增架构代码：**0 个测试用例** ❌

## Requirements

### R1: 领域模型测试 (`model/user_test.go`, `model/checkin_test.go`)
- User.IsBanned() — 正常用户返回 false，封禁用户返回 true
- User.CanCheckin() — 正常用户返回 nil，封禁用户返回 error
- User.DisplayName() — 有昵称返回昵称，无昵称返回"匿名用户"
- Checkin.IsProcessed() — 各状态正确判断
- Checkin.IsFailed() — 各状态正确判断
- Checkin.IsPending() — 各状态正确判断
- Checkin.BelongsTo() — 匹配/不匹配
- Checkin.CanDelete() — 作者/非作者

### R2: 重试工具测试 (`pkg/retry/retry_test.go`)
- 成功一次不重试
- 网络错误重试 3 次后成功
- 非可重试错误立即失败
- 全部重试失败返回最后一次错误
- DoVoid 正确传播错误

### R3: Repository 接口编译验证
- `go build ./internal/repository/...` 通过
- `go vet ./...` 通过

### R4: app.go 编译验证
- `go build ./internal/app/...` 通过
- `go build ./cmd/...` 通过

## Acceptance Criteria

- [ ] `go test ./internal/model/ -v` 全部通过
- [ ] `go test ./pkg/retry/ -v` 全部通过
- [ ] `go build ./...` 编译通过
- [ ] `go vet ./...` 无警告
- [ ] 领域模型测试覆盖所有方法的正向 + 反向路径

## Out of Scope

- `internal/app/app.go` 的单元测试（组装代码，依赖真实 MongoDB）
- 客户端 API 拆分的运行时测试（barrel re-export 由 TypeScript 编译器保证）
- Repository 接口的 mock 测试（已在 auth 模块测试中验证模式）

## Open Questions

（无阻塞性问题）
