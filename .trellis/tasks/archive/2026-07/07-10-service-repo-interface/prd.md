# Service 层接入 Repository 接口

## Goal

将剩余 9 个 Service 的 Repository 依赖从具体类型改为接口，使其可 mock 测试。

## Background

### 当前状态
- **已迁移**: `AuthService` — 使用 `repository.UserRepoInterface` ✅
- **未迁移**: 9 个 Service 仍使用具体类型 `*repository.XxxRepo`

### 已有接口（无需新建）
8 个 `*_iface.go` 文件已定义全部所需接口，覆盖 14 个结构体、64 个方法。

### 需迁移的 Service

| Service | 依赖的 Repository | 涉及接口 |
|---------|-------------------|---------|
| UserService | UserRepo | UserRepoInterface |
| CheckinService | CheckinRepo, UserRepo | CheckinRepoInterface, UserRepoInterface |
| SocialService | CommentRepo, LikeRepo, CheckinRepo, UserRepo | CommentRepoInterface, LikeRepoInterface, CheckinRepoInterface, UserRepoInterface |
| InsightService | InsightRepo, InsightTagRepo, InsightLikeRepo, UserRepo | InsightRepoInterface, InsightTagRepoInterface, InsightLikeRepoInterface, UserRepoInterface |
| ResourceService | ResourceRepo, ResourceTagRepo, UserRepo | ResourceRepoInterface, ResourceTagRepoInterface, UserRepoInterface |
| TrainingService | TrainingRepo, TemplateRepo | TrainingRepoInterface, TemplateRepoInterface |
| NotificationService | NotificationRepo, NotificationSettingsRepo, UserRepo | NotificationRepoInterface, NotificationSettingsRepoInterface, UserRepoInterface |
| GroupService | GroupRepo, UserRepo | GroupRepoInterface, UserRepoInterface |
| RankService | RankRepo | RankRepoInterface |
| CronService | UserRepo, CheckinRepo, RankRepo, TrainingRepo, NotificationRepo | UserRepoInterface, CheckinRepoInterface, RankRepoInterface, TrainingRepoInterface, NotificationRepoInterface |

## Requirements

### R1: 逐个 Service 迁移
每个 Service 的变更模式相同：
1. struct 字段类型 `*repository.XxxRepo` → `repository.XxxRepoInterface`
2. 构造函数参数类型同步变更
3. 方法体不需要改（接口方法签名相同）

### R2: `main.go` / `app.go` 兼容性
- `*repository.XxxRepo` 已实现 `XxxRepoInterface`（`var _ Interface = (*Impl)(nil)` 保证）
- 构造函数调用不需要改（Go 接口赋值兼容）

### R3: 编译验证
- `go build ./...` 通过
- `go vet ./...` 无警告

## Acceptance Criteria

- [ ] 全部 10 个 Service 使用接口而非具体类型
- [ ] `go build ./...` 编译通过
- [ ] `go vet ./...` 无警告
- [ ] `cmd/main.go` 和 `internal/app/app.go` 调用不需要修改

## Out of Scope

- 为每个 Service 编写单元测试（下一步工作）
- 修改 Repository 实现代码
- 新建接口文件（已全部就绪）

## Open Questions

（无阻塞性问题 — 接口已定义，变更模式固定）
