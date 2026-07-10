# 鉴权模块单元测试

## Goal

为鉴权模块（auth module）的 15 项改进编写单元测试，确保改动不引入回归，并为后续重构提供安全网。

## Background

### 刚完成的改进涉及以下文件
- `api-server/internal/middleware/auth.go` — Auth 中间件、Bearer 校验、UserStatusCheck
- `api-server/internal/middleware/token_blacklist.go` — Token 黑名单（内存实现）
- `api-server/internal/middleware/cors.go` — CORS 配置化
- `api-server/pkg/jwt/jwt.go` — JWT Generate/Parse + Refresh Token
- `api-server/internal/repository/user_repo.go` — UpsertByOpenID、IsBanned、Create
- `api-server/internal/service/auth_service.go` — WXLogin、RefreshToken、getOpenID 重试
- `api-server/internal/handler/auth_handler.go` — Login、Logout、Refresh 端点

### 当前测试状态
- 项目零测试文件（无 `*_test.go`、无 `*.test.ts`）
- 无第三方测试框架依赖（仅 Go stdlib `testing`）
- Go 1.21，使用 Gin、mongo-driver、jwt/v5

### 外部依赖
- MongoDB（需 mock）
- 微信 API（需 mock HTTP Transport）
- Redis（当前未引入，blacklist 用内存实现）

## Design Decisions

### D1: Repository 测试策略 — 接口抽象 + Mock
- 为 `UserRepo` 定义接口，Service 层依赖接口而非具体类型
- 测试时注入 mock 实现，验证 Service 逻辑
- Repository 实现本身的正确性由 MongoDB 唯一索引 + 集成环境保证

### D2: Mock 实现方式 — gomock
- 使用 `go.uber.org/mock` 生成 mock 代码
- 通过 `go generate` 在接口定义处标注 `//go:generate mockgen`
- 自动生成的 mock 随接口变更自动同步

### D3: 测试覆盖率目标 — 全面覆盖
- 所有鉴权相关包 ≥ 80%，核心安全包（JWT、Auth 中间件、TokenBlacklist）≥ 90%
- 预计工作量 3-4 天

### D4: 断言库 — testify/assert + require
- 引入 `github.com/stretchr/testify`（仅 `assert` 和 `require` 子包）
- `require` 用于前置条件（失败立即 `t.FailNow`），`assert` 用于非关键断言

### D5: 微信 API Mock — 注入 HTTP Client
- `NewAuthService` 增加可选 `*http.Client` 参数，默认使用 `http.DefaultClient`
- 测试时注入自定义 `http.Transport`，精确控制返回值、延迟和错误
- Transport 层拦截所有出站请求，不需要改业务逻辑 URL

## Requirements

### R1: JWT 包测试 (`pkg/jwt/jwt_test.go`)
- Generate 生成有效 token，Parse 能正确解析
- Parse 拒绝过期 token
- Parse 拒绝无效签名（错误密钥）
- Parse 拒绝错误 Issuer
- GenerateRefreshToken + ParseRefreshToken 正确工作
- Refresh token 使用不同密钥，不能被 Parse 解析

### R2: TokenBlacklist 测试 (`middleware/token_blacklist_test.go`)
- Revoke 后 IsRevoked 返回 true
- 未 Revoke 的 token 返回 false
- 过期条目被正确清理
- 并发安全性（`go test -race` 通过）

### R3: Auth 中间件测试 (`middleware/auth_test.go`)
- 无 Authorization header → 401
- 非 Bearer 格式 → 401
- 有效 token → 200 + user_id 在 context 中
- 已撤销 token → 401
- 过期 token → 401
- UserStatusCheck 正常用户 → 放行
- UserStatusCheck 封禁用户 → 403

### R4: CORS 中间件测试 (`middleware/cors_test.go`)
- 白名单 origin → 正确设置 Allow-Origin + Credentials
- 非白名单 origin → 不设置 Allow-Origin
- 无 origin（空字符串）→ 设置 `*`
- OPTIONS 请求 → 204 + 不调用 Next

### R5: Service 层测试 (`service/auth_service_test.go`)
- WXLogin 新用户 → 创建 + 返回双 token
- WXLogin 老用户 → 更新资料 + 返回双 token
- RefreshToken 有效 → 新双 token
- RefreshToken 过期 → 错误
- getOpenID 网络错误 → 重试 3 次
- getOpenID 微信业务错误 → 不重试，直接返回错误
- getOpenID 超时 → 重试

### R6: Repository 接口定义 (`repository/user_repo_iface.go`)
- 提取 `UserRepoInterface` 接口
- 包含：`UpsertByOpenID`、`FindByOpenID`、`FindByID`、`IsBanned`、`Create`
- `//go:generate mockgen` 标注

## Acceptance Criteria

- [ ] `go test ./...` 全部通过
- [ ] `go test -race ./...` 无竞态报告
- [ ] JWT 包测试覆盖率 ≥ 90%
- [ ] TokenBlacklist 测试覆盖率 ≥ 90%
- [ ] Auth 中间件核心路径 100% 覆盖
- [ ] Service 层 mock 测试覆盖所有分支
- [ ] gomock 生成的 mock 代码能正常编译

## Out of Scope

- 客户端测试（UniApp 需要独立测试环境）
- 集成测试（需要完整 MongoDB + 微信沙箱）
- 性能测试 / 基准测试
- E2E 测试
- Repository 实现层测试（由集成环境覆盖）

## Open Questions

（无阻塞性问题，所有决策已确认）
