# 武俱打卡项目 — 整体架构审查报告

> 审查日期: 2026-07-09
> 审查范围: api-server / media-server / deploy 全栈架构
> 严重级别: CRITICAL / HIGH / MEDIUM / LOW

---

## 一、双服务职责边界

### [MEDIUM] 资料库上传绕过 media-server，造成职责交叉
- api-server 的 resource_handler.go 中实现了资料库文件的 Presign/UploadCallback 逻辑，直接处理文件上传预签名和上传完成回调，绕过了 media-server。根据架构约束 C1，media-server 应处理所有媒体文件上传，而 resource 模块自行处理了文件上传流程。
- **文件**: `api-server/internal/handler/resource_handler.go` (Presign / UploadCallback 方法)
- **建议**: 将资料库的文件上传流程统一由 media-server 处理，api-server 仅负责业务记录创建。或明确在架构文档中声明 resource 模块的上传是例外情况。

### [LOW] api-server 的 CheckinService 持有 mediaURL 配置
- `CheckinService` 直接持有 `mediaURL` 字符串配置（`service.go:121`），这是配置信息泄漏到业务层的表现。虽然当前用途是占位，但原则上配置应仅在 config 层和 handler 层使用。
- **文件**: `api-server/internal/service/service.go:121`
- **建议**: 删除无用的 mediaURL 字段，或仅在 handler 层持有。Service 层不应感知服务地址。

### [LOW] api-server 与 media-server 的认证方案不一致
- api-server 的内部接口使用 `X-Internal-Secret` 请求头认证（`internal_auth.go`）。
- media-server 的受保护接口同时支持 `Authorization: Bearer` 和 `X-API-Key` 两种方式（`media-server/internal/middleware/auth.go:13-21`）。
- 虽然当前两个服务使用同一密钥，但认证方式的不一致在未来可能造成混淆。
- **建议**: 统一认证方式，两服务均使用 `X-Internal-Secret` 或均使用 `Authorization: Bearer`。

---

## 二、层间依赖方向

### [LOW] handler.go 中的 CheckinHandler 同时持有 checkinService 和 socialService
- `CheckinHandler` 同时持有两个 Service，使其成为跨域 handler。虽然功能上是为了在 GetByID/GetList 中注入点赞状态，但职责边界模糊。理论上 SocialHandler 应该处理所有社交相关操作。
- **文件**: `api-server/internal/handler/handler.go:103-109`
- **建议**: 将 GetByID/GetList 中的点赞状态填充逻辑移到 Service 层，Handler 只持有单一 Service。

### [LOW] Service 层间接依赖 config.Config
- `AuthService` 的构造函数接收 `*config.Config` 完整配置对象（`service.go:27`），仅使用了 `cfg.WX` 字段。这违反了 Service 不应感知配置的原则。
- **文件**: `api-server/internal/service/service.go:27-28`
- **建议**: 只传入 Service 需要的具体值（如 AppID, Secret, TemplateID），而非完整 Config 对象。

### [OK] 依赖方向总体正确
- Handler → Service → Repository 的依赖方向在整体上严格保持，没有发现反向依赖。
- Handler 没有直接操作数据库 ✅
- Service 不使用 `gin.Context` ✅（全部使用 `context.Context`）
- Repository 不包含业务逻辑 ✅

---

## 三、数据流

### [OK] 请求/响应链路合理
- 客户端 → Nginx → API Server (业务) / Media Server (媒体) 的链路清晰。
- 客户端 → 预签名 URL → MinIO 直传模式避免了服务端带宽瓶颈。
- Media Server → API Server 的回调模式（转码完成回调）设计合理。

### [LOW] 日志记录分散，未使用结构化日志
- handler.go 中多处使用 `log.Printf` 而非注入的 `zap.Logger`（如 `handler.go:39`）。虽然 router.go 接收了 logger 参数，但 Handler 层并未使用。
- **文件**: `api-server/internal/handler/handler.go` (多处 `log.Printf`)
- **建议**: 将 `*zap.Logger` 注入到 Handler 结构体中，替换全局 `log.Printf`。

### [OK] 服务间通信
- 服务间最多一级调用（media-server → api-server 回调），没有深层调用链。
- 使用共享密钥认证，符合架构约束 C3。

---

## 四、模块耦合

### [MEDIUM] Service 层横向耦合（NotificationService 被多处注入）
- `NotificationService` 同时被 `SocialService`（`service.go:243`）和 `TrainingService`（`training_service.go:18`）依赖，创建了 Service 层的循环依赖风险域。当通知逻辑变更时，可能影响社交和训练模块。
- **文件**: `api-server/internal/service/service.go:243`, `api-server/internal/service/training_service.go:18`
- **建议**: 引入事件/观察者模式解耦，社交和训练模块只需发出事件，通知模块自行订阅。或保持当前模式，但要确保通知接口稳定。

### [LOW] CronService 持有 6 个依赖，职责范围过大
- `CronService` 持有 `userRepo`, `checkinRepo`, `rankRepo`, `planRepo`, `notifRepo`, `wxClient`, `cfg` 共 7 个依赖，几乎覆盖了所有领域。虽然定时任务需要跨域操作，但这种"大管家"模式在模块增加时会持续膨胀。
- **文件**: `api-server/internal/service/cron.go:20-28`
- **建议**: 将 cron 任务拆分为多个独立的 Job 结构体（如 RankRefreshJob, TrainingRemindJob），各自持有所需的最小依赖。

### [LOW] 定时任务调度实现过于原始
- 排行榜刷新使用 `time.NewTicker` + goroutine，训练提醒使用 `time.After` + goroutine。没有重试机制、没有错误恢复、没有监控指标。
- **文件**: `api-server/cmd/main.go:121-156`
- **建议**: 引入 cronexpr 库或 robfig/cron 库管理定时任务，增加重试和告警机制。

---

## 五、扩展性

### [MEDIUM] 依赖注入完全硬编码在 main.go
- 所有 Repository → Service → Handler 的装配在 main.go 中逐一手动完成。当前 14 个 Repo + 11 个 Service + 10 个 Handler 的规模已经使 main.go 达到 216 行且仍在增长。未来新增模块会导致 main.go 持续膨胀。
- **文件**: `api-server/cmd/main.go:57-169`
- **建议**: 
  - 引入依赖注入容器（如 `google/wire` 或 `uber-go/fx`）减少样板代码
  - 或将初始化逻辑按模块拆分为独立函数

### [LOW] 无插件/扩展机制
- 当前架构不支持模块热插拔或条件编译。所有功能模块（打卡、训练、感悟、资料库）都是编译时绑定。
- **建议**: 引入模块注册机制，通过配置控制启用/禁用特定模块。

### [LOW] 无事件驱动机制
- 社交模块的通知发送同步调用 `notifService.Send()`（如点赞、评论后的通知），若通知服务不可用会阻塞主流程。
- **文件**: `api-server/internal/service/service.go:283-295`
- **建议**: 引入内存事件总线或消息队列解耦，通知发送改为异步模式。

### [OK] 模块扩展模式清晰
- 新增模块遵循 model → repository → service → handler → router 的固定模式，可预测性强。
- 目录结构规范，易于定位。

---

## 六、代码组织

### [HIGH] handler.go 和 service.go 文件过大，违反项目自身规范
- `handler.go` 487 行，包含 AuthHandler / UserHandler / CheckinHandler / SocialHandler / RankHandler / GroupHandler 共 6 个类型。
- `service.go` 463 行，包含 AuthService / UserService / CheckinService / SocialService / RankService / GroupService 共 6 个类型。
- 项目的目录结构规范文件（`directory-structure.md:106-107`）已明确指出这是违规例，但尚未修复。
- **建议**: 立即拆分：
  - `handler.go` → `auth_handler.go`, `user_handler.go`, `checkin_handler.go`, `social_handler.go`, `rank_handler.go`, `group_handler.go`
  - `service.go` → `auth_service.go`, `user_service.go`, `checkin_service.go`, `social_service.go`, `rank_service.go`, `group_service.go`

### [LOW] rate_limit.go 存在但未使用
- `LoginRateLimit()` 中间件已实现（5分钟5次登录限制），但在 `router.go` 中未注册。安全架构文档（`security-constraints.md:203`）已标注为"待完全实现"状态。
- **文件**: `api-server/internal/middleware/rate_limit.go:97-109`
- **建议**: 在 router.go 中添加 `api.POST("/auth/login", middleware.LoginRateLimit(), authH.Login)`。

### [LOW] media-server 使用 gin.Default() 而非 gin.New()
- `media-server/internal/router/router.go:11` 使用 `gin.Default()`（自动包含 Logger 和 Recovery），但 api-server 使用 `gin.New()` + 自定义中间件。这导致 media-server 缺少 RequestID 中间件。
- **文件**: `media-server/internal/router/router.go:11`
- **建议**: 统一为 `gin.New()` + 显式中间件注册，确保 media-server 也有 RequestID。

### [LOW] utils.go 内容过少
- `service/utils.go` 仅包含 `extractTags` 函数，文件存在感弱。
- **文件**: `api-server/internal/service/utils.go`
- **建议**: 将工具函数归类到合适的包（如 `internal/pkg/utils`），或扩充该文件以容纳更多共享工具函数。

### [LOW] constants/ 目录存在但未使用
- `api-server/internal/constants/` 目录已创建但未在任何文件中引用（根据 git status 显示为新增）。
- **建议**: 将散落在各处的魔数（如 `CheckinStatusPending`, `defaultQuota` 等）统一抽取到 constants 包。

---

## 总结

| 维度 | 评分 | 关键问题数 |
|------|------|-----------|
| 双服务职责边界 | B (良好) | 2 MEDIUM |
| 层间依赖方向 | A (优秀) | 2 LOW |
| 数据流 | B+ (良好) | 1 LOW |
| 模块耦合 | B (良好) | 3 MEDIUM/LOW |
| 扩展性 | B (良好) | 2 MEDIUM/LOW |
| 代码组织 | C (需改进) | 1 HIGH + 4 LOW |

**总体评价**: 架构设计合理，双服务分离、三层架构、依赖注入模式都执行到位。主要问题是 **handler.go/service.go 的单文件过大的遗留问题**和 **LoginRateLimit 未注册的安全漏洞**。建议优先解决 HIGH 和 MEDIUM 级别问题，其余 LOW 问题可在日常迭代中逐步优化。