# 鉴权模块测试 — 实现计划

## 实现顺序

按依赖关系从底层到上层，每步完成后运行 `go test ./...` 验证。

### Step 1: 安装依赖 + 初始化
```bash
cd api-server
go get go.uber.org/mock/mockgen@latest
go get github.com/stretchr/testify@latest
```

**产出**: `go.mod` / `go.sum` 更新

---

### Step 2: 定义 Repository 接口

**文件**: `api-server/internal/repository/user_repo_iface.go`（新建）

内容：
- `UserRepoInterface` 接口，包含 5 个方法
- `//go:generate mockgen` 标注

**文件**: `api-server/internal/repository/user_repo.go`（不改）

`*UserRepo` 已实现所有接口方法，无需修改。

**验证**:
```bash
go build ./internal/repository/...
```

---

### Step 3: 生成 Mock + 修改 Service 依赖

**操作**:
```bash
cd api-server/internal/repository
go generate ./...
```

**产出**: `api-server/internal/repository/mock_user_repo.go`（自动生成）

**文件**: `api-server/internal/service/auth_service.go`（修改）

变更：
1. `userRepo *repository.UserRepo` → `userRepo repository.UserRepoInterface`
2. 增加 `httpClient *http.Client` 字段
3. `NewAuthService` 签名增加 variadic `httpClient ...*http.Client`
4. `getOpenID` 使用 `s.httpClient` 替代新建 `http.Client`

**文件**: `api-server/cmd/main.go`（无需修改，variadic 兼容）

**验证**:
```bash
go build ./...
```

---

### Step 4: JWT 包测试

**文件**: `api-server/pkg/jwt/jwt_test.go`（新建）

测试用例：
| 测试名 | 验证点 |
|--------|--------|
| TestGenerate_And_Parse | 生成 + 解析 round-trip |
| TestParse_ExpiredToken | 过期 token → error |
| TestParse_InvalidSignature | 错误密钥签名 → error |
| TestParse_WrongIssuer | Issuer 不匹配 → error |
| TestGenerateRefreshToken_And_Parse | refresh token round-trip |
| TestParse_CannotParseRefreshToken | access Parse 无法解析 refresh token |
| TestParseRefreshToken_WrongIssuer | refresh Issuer 不匹配 → error |

**验证**:
```bash
go test ./pkg/jwt/ -v -cover
```

---

### Step 5: TokenBlacklist 测试

**文件**: `api-server/internal/middleware/token_blacklist_test.go`（新建）

测试用例：
| 测试名 | 验证点 |
|--------|--------|
| TestIsRevoked_NotRevoked | 未撤销 → false |
| TestIsRevoked_Revoked | 撤销后 → true |
| TestIsRevoked_Expired | 过期条目 → false + 自动清理 |
| TestIsRevoked_Concurrent | 100 goroutine 并发读写无 race |

**验证**:
```bash
go test ./internal/middleware/ -v -race -run TestIsRevoked
```

---

### Step 6: Auth 中间件测试

**文件**: `api-server/internal/middleware/auth_test.go`（新建）

测试用例：
| 测试名 | 验证点 |
|--------|--------|
| TestAuth_MissingHeader | 无 Authorization → 401 |
| TestAuth_InvalidFormat | 非 Bearer 格式 → 401 |
| TestAuth_ValidToken | 有效 token → 200, user_id in context |
| TestAuth_RevokedToken | 已撤销 token → 401 |
| TestAuth_ExpiredToken | 过期 token → 401 |
| TestUserStatusCheck_Active | 正常用户 → 放行 |
| TestUserStatusCheck_Banned | 封禁用户 → 403 |
| TestUserStatusCheck_DBError | DB 错误 → 放行（降级策略） |

使用 `httptest` 创建 Gin 测试引擎。

**验证**:
```bash
go test ./internal/middleware/ -v -cover -run TestAuth
```

---

### Step 7: CORS 中间件测试

**文件**: `api-server/internal/middleware/cors_test.go`（新建）

测试用例：
| 测试名 | 验证点 |
|--------|--------|
| TestCORS_AllowedOrigin | 白名单 origin → Allow-Origin = origin |
| TestCORS_DisallowedOrigin | 非白名单 → 无 Allow-Origin |
| TestCORS_NoOrigin | 空 origin → Allow-Origin = * |
| TestCORS_OptionsPreflight | OPTIONS → 204 |
| TestCORS_RequestID | 无 X-Request-ID → 自动生成 |

**验证**:
```bash
go test ./internal/middleware/ -v -cover -run TestCORS
```

---

### Step 8: Service 层测试

**文件**: `api-server/internal/service/auth_service_test.go`（新建）

使用 gomock + 自定义 HTTP Transport。

测试用例：
| 测试名 | 验证点 |
|--------|--------|
| TestWXLogin_NewUser | 新用户 → UpsertByOpenID(isCreated=true) + 双 token |
| TestWXLogin_ExistingUser | 老用户 → UpsertByOpenID(isCreated=false) + 双 token |
| TestWXLogin_WXAPIError | 微信业务错误 → 不重试，直接返回 |
| TestWXLogin_NetworkRetry | 网络错误 → 重试 3 次 |
| TestWXLogin_AllRetriesFail | 3 次全失败 → 返回错误 |
| TestRefreshToken_Valid | 有效 refresh token → 新双 token |
| TestRefreshToken_Expired | 过期 refresh token → 错误 |
| TestRefreshToken_UserBanned | 用户已封禁 → 错误 |

mock HTTP Transport 结构：
```go
type mockTransport struct {
    responses []responseEntry
    callCount int
    mu        sync.Mutex
}
type responseEntry struct {
    resp *http.Response
    err  error
}
```

**验证**:
```bash
go test ./internal/service/ -v -cover -run TestWXLogin
go test ./internal/service/ -v -cover -run TestRefreshToken
```

---

### Step 9: 全量验证

```bash
cd api-server
go test ./... -v -race -cover
```

**期望输出**:
- 所有测试 PASS
- 无 race condition
- JWT 包 ≥ 90% coverage
- Middleware 包 ≥ 90% coverage
- Service 包 ≥ 80% coverage

---

## 风险文件

| 文件 | 风险 | 缓解 |
|------|------|------|
| `service/auth_service.go` | 改构造函数影响调用方 | variadic 参数兼容 |
| `repository/mock_user_repo.go` | 生成代码可能与接口不同步 | `go generate` 在 CI 中运行 |
| `cmd/main.go` | 理论上不需要改，但需验证编译 | Step 3 后 `go build` |

## 回滚方案

如果测试引入的问题多于解决的问题：
1. 删除所有 `*_test.go` 文件
2. 删除 `mock_user_repo.go` 和 `user_repo_iface.go`
3. 还原 `auth_service.go` 中的接口/HTTP Client 变更
4. `go mod tidy` 清理依赖
