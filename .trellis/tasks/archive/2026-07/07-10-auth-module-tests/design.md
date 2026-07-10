# 鉴权模块测试 — 技术设计

## Architecture

### 重构：提取 Repository 接口

当前 `AuthService` 直接依赖 `*repository.UserRepo`（具体类型）。需要提取接口使 Service 层可测试。

```
Before:
  AuthService ──depends──▶ *repository.UserRepo (concrete)

After:
  AuthService ──depends──▶ repository.UserRepoInterface (interface)
                                ▲
                    ┌───────────┴───────────┐
                    │                       │
            *repository.UserRepo    *mock_repository.MockUserRepoInterface
            (production)            (test)
```

### 接口定义

```go
// repository/user_repo_iface.go
type UserRepoInterface interface {
    Create(ctx context.Context, user *model.User) error
    FindByOpenID(ctx context.Context, openid string) (*model.User, error)
    FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
    UpsertByOpenID(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error)
    IsBanned(ctx context.Context, id primitive.ObjectID) (bool, error)
}
```

### Mock 生成

```go
//go:generate mockgen -destination=mock_user_repo.go -package=repository wuxie-api/internal/repository UserRepoInterface
```

### AuthService 依赖注入变更

```go
// Before
func NewAuthService(userRepo *repository.UserRepo, jwtMgr *jwt.JWTManager, cfg *config.Config, logger *zap.Logger) *AuthService

// After
func NewAuthService(userRepo repository.UserRepoInterface, jwtMgr *jwt.JWTManager, cfg *config.Config, logger *zap.Logger, httpClient ...*http.Client) *AuthService
```

- `httpClient` 为可变参数，不传时使用 `http.DefaultClient`
- Service 内部保存 `httpClient`，在 `getOpenID` 中使用

## Data Flow

### 测试时的依赖注入

```
Test_WXLogin_NewUser:
  mockUserRepo.EXPECT().UpsertByOpenID(...) → returns (user, true, nil)
  mockHTTPTransport → returns wx API response with openid
  jwtMgr → real JWT manager (for Generate/Parse)

  → AuthService.WXLogin(ctx, "code", "nick", "avatar", 1)
  → assert: token != "", refreshToken != "", user != nil, isCreated
```

### HTTP Client Mock 策略

```
Test_getOpenID_NetworkRetry:
  transport = &mockTransport{
    responses: [
      error(net.Error),       // attempt 1: network error → retry
      error(net.Error),       // attempt 2: network error → retry
      success(`{"openid":"x"}`), // attempt 3: success
    ]
  }
  client = &http.Client{Transport: transport, Timeout: 1*time.Second}

  → getOpenID("code")
  → assert: openid == "x", transport.callCount == 3
```

## Compatibility

### 需要修改的现有代码

| 文件 | 变更 | 影响 |
|------|------|------|
| `repository/user_repo.go` | `UserRepo` 已实现所有接口方法，无需改代码 | 无 |
| `repository/user_repo_iface.go` | **新增**接口定义 | 无 |
| `repository/mock_user_repo.go` | **新增** gomock 生成 | 无 |
| `service/auth_service.go` | `UserRepo` → `UserRepoInterface`，增加 `httpClient` 参数 | 调用方需更新 |
| `cmd/main.go` | `NewAuthService` 调用不变（variadic 参数兼容） | **无影响** |

### 向后兼容性

- `*UserRepo` 已实现 `UserRepoInterface` 的所有方法，无需改实现代码
- `httpClient` 使用 variadic 参数，现有调用 `NewAuthService(repo, jwt, cfg, logger)` 不需要改
- `main.go` 中的调用无需变更

## Trade-offs

### 选择 gomock 而非手写 Mock
- **优势**: mock 随接口变更自动同步，支持参数匹配器（`gomock.Eq()`、`gomock.Any()`）
- **代价**: 需要安装 `mockgen` 工具，`go generate` 步骤

### 选择注入 HTTP Client 而非抽象 WeChatClient
- **优势**: 最小改动，符合 Go 惯例
- **代价**: 测试需要理解 `http.Transport` 接口

### 不测 Repository 实现层
- **优势**: 不依赖 Docker/MongoDB，测试快
- **代价**: UpsertByOpenID 的 MongoDB 行为需靠集成环境验证
