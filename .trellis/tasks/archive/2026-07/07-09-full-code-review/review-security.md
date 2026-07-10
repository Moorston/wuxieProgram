# 武俱打卡项目安全审查报告

> 审查日期：2026-07-09
> 审查范围：api-server + media-server 核心代码
> 严重级别说明：CRITICAL（必须立即修复）/ HIGH（尽快修复）/ MEDIUM（计划修复）/ LOW（建议改进）

---

## 一、CRITICAL 级别问题

### 1.1 CORS 配置允许任意 Origin + Credentials
[CRITICAL] api-server/internal/middleware/cors.go:13-21 — CORS 中间件将请求头 `Origin` 直接回显到 `Access-Control-Allow-Origin`，同时设置了 `Access-Control-Allow-Credentials: true`。这意味着任意第三方网站都可以发起携带 Cookie/Authorization 的跨域请求，导致 CSRF 攻击风险。攻击者可以构造恶意页面，在用户登录状态下执行打卡、删除资源等操作。
**修复建议**：将 `Access-Control-Allow-Origin` 限制为具体的白名单域名列表（如小程序域名和官方前端域名），不要动态回显 Origin 头。或者移除 `Access-Control-Allow-Credentials`。

### 1.2 文件扩展名参数无任何验证，存在路径遍历风险
[CRITICAL] media-server/internal/handler/handler.go:35-36 — Presign 接口的 `ext` 查询参数完全由用户控制，未做任何验证就拼接到对象名中：`fmt.Sprintf("%s/%s.%s", ...uuid, ext)`。如果攻击者传入 `ext=../../secret`，对象名变为 `20260709/uuid.../secret`，可能导致 MinIO 对象存储的路径遍历。
**修复建议**：对 `ext` 参数做白名单验证（如只允许 `mp4`, `jpg`, `png` 等），或使用正则 `^[a-zA-Z0-9]+$` 限制格式。

### 1.3 JWT Secret / Media Secret / WX Secret 使用默认占位符
[CRITICAL] api-server/configs/config.yaml:15,19-26 — 生产配置中使用硬编码默认密钥：`wuxie-jwt-secret-change-in-production`（JWT）、`your-wx-app-secret`（微信）、`wuxie-media-secret-change-in-production`（内部API）。若实际部署时未修改，攻击者可伪造任意 JWT Token 或调用内部API。
**修复建议**：部署脚本中必须随机生成至少32字符的密钥，通过环境变量注入，禁止在配置文件中存放真实密钥。

### 1.4 MinIO 使用默认凭据
[CRITICAL] media-server/configs/config.yaml:5-8 — MinIO 的 `access_key: "minioadmin"` 和 `secret_key: "minioadmin"` 均为 MinIO 官方默认凭据。任何能访问 MinIO 服务的攻击者可直接管理所有存储桶。
**修复建议**：部署时修改为强密码，通过环境变量 `MINIO_ACCESS_KEY` 和 `MINIO_SECRET_KEY` 注入。

### 1.5 Redis 无密码认证
[CRITICAL] api-server/configs/config.yaml:11 — Redis 密码为空 (`password: ""`)。同时 media-server/configs/config.yaml:15 同样为空。攻击者若接入 Redis 端口，可完全控制缓存和 transcode 队列，甚至注入恶意转码任务。
**修复建议**：设置强 Redis 密码，通过 `REDIS_PASSWORD` 环境变量传入。

### 1.6 搜索功能 NoSQL 正则注入（ReDoS + 数据泄露）
[CRITICAL] api-server/internal/repository/resource_repo.go:84-88 — 资源列表搜索使用 `$regex` 处理用户输入的 keyword，虽然调用了 `sanitizeRegex`，但未调用 `validateSearchKeyword`，缺少对重复特殊字符和嵌套括号的限制。攻击者可构造恶意正则表达式（如 `(a+)+b`）导致 MongoDB 服务端拒绝服务。
**修复建议**：对 keyword 先调用 `validateSearchKeyword`（checkin_repo.go 已有该函数），再做 `sanitizeRegex`。并将 `$regex` 搜索限制在 `^` 前缀匹配模式或使用全文索引替代。

---

## 二、HIGH 级别问题

### 2.1 登录接口无有效速率限制
[HIGH] api-server/internal/middleware/rate_limit.go:96-109 — `LoginRateLimit` 函数虽然定义了一个登录速率限制器（5次/5分钟），但这个限制器 **从未在路由中注册**。且在实现上，每次调用 `LoginRateLimit()` 都会创建新的 `RateLimiter` 实例，导致限制完全无效。攻击者可对 `/api/auth/login` 进行无限暴力破解尝试。
**修复建议**：在路由注册时对 `/api/auth/login` 应用 `rate_limit.LoginRateLimit()` 中间件，并确保 `RateLimiter` 是单例而非每次请求创建。

### 2.2 训练计划报告无所有权检查
[HIGH] api-server/internal/handler/training_handler.go:283-296 — `GetReport` 接口只验证了 planID 是否有效，**完全没有检查当前用户是否为该计划的所有者**。任何登录用户只需遍历 planID 即可获取任意训练计划的完整报告。
**修复建议**：在 handler 层获取 plan 并校验 `plan.UserID == oid`，与 `GetPlan` 方法保持一致的权限检查。

### 2.3 Media-Server 错误信息直接暴露给客户端
[HIGH] media-server/internal/handler/handler.go:40,105 — `Presign` 和 `GetURL` 接口在 MinIO 操作失败时，直接返回 `err.Error()` 给客户端（`response.InternalError(c, err.Error())`）。这可能暴露 MinIO 内部端点、存储桶名称和文件路径等敏感信息。
**修复建议**：改为返回通用错误消息（如 `"internal server error"`），将详细错误写入服务端日志。

### 2.4 用户输入文本字段缺少长度限制
[HIGH] api-server/internal/handler/handler.go:112-114 — `PrepareReq.Description`、`CreateInsightReq.Content`、`CreateResourceReq.Title` 等文本字段均无长度验证。攻击者可提交超长文本导致 MongoDB 文档大小超限、内存耗尽或索引失败。多处 handler 均存在此问题。
**修复建议**：为所有用户文本输入字段增加 `binding:"max=2000"` 约束，在 service 层增加二次校验。

### 2.5 Insight 点赞操作缺少事务一致性
[HIGH] api-server/internal/repository/insight_repo.go:216-235 — `InsightLikeRepo.Toggle` 方法在执行点赞切换后直接返回，没有使用 MongoDB 事务同步更新 `Insight` 文档中的 `like_count`。这会导致点赞计数不一致（而 `LikeRepo` 使用了事务）。
**修复建议**：参考 `LikeRepo.ToggleWithSession` 的实现，在事务中同时执行点赞切换和 `IncrLikeCount`。

### 2.6 资源收藏接口错误地将自己资源作为必要条件
[HIGH] api-server/internal/repository/resource_repo.go:140 — `ToggleFavorite` 的查询条件为 `{"_id": id, "user_id": userID}`，这要求用户必须是资源的所有者才能收藏。正确的收藏功能应该允许用户收藏任何有权限查看的资源。
**修复建议**：移除 `user_id` 过滤条件，改为先检查资源是否可被当前用户访问。

### 2.7 FFmpeg binary 名称拼接错误
[HIGH] media-server/pkg/ffmpeg/ffmpeg.go:28 — `exec.CommandContext(ctx, binary+"probe", args...)` 将 binary 名称和 "probe" 拼接为 "ffmpegprobe"，但标准工具名应为 "ffprobe"。这导致每次调用 `Probe` 时都会执行失败并回退到 `probeWithFFmpeg`（解析文本输出，可靠性低）。虽然不影响主流程，但降低了视频信息获取的准确性和效率。
**修复建议**：改为 `"ffprobe"` 硬编码或通过配置项指定 probe 工具路径。

---

## 三、MEDIUM 级别问题

### 3.1 文件扩展名参数在 api-server 中也未验证
[MEDIUM] api-server/internal/handler/resource_handler.go:29 — `Presign` 接口同样接受用户控制的 `ext` 参数且无验证，传递给 `GenerateObjectName`（resource_service.go:178-183）直接拼接到对象名。
**修复建议**：同上，对 ext 做白名单验证。

### 3.2 createResourceReq.FileURL / CoverURL 用户可控无验证
[MEDIUM] api-server/internal/handler/resource_handler.go:83-87 — `CreateResourceReq.FileURL` 和 `CoverURL` 来自用户请求且 `binding:"required"`，但未验证 URL 格式和协议。用户可存储任意 URL（如 `javascript:alert(1)` 或 `file:///etc/passwd`），返回给其他用户时可能导致 XSS 或其他攻击。
**修复建议**：验证 URL 必须以 `https?://` 或合理协议开头，且符合预期域名模式。

### 3.3 帖子评论内容仅做了基本长度限制
[MEDIUM] api-server/internal/handler/handler.go:329-330 — `CommentReq.Content` 限定了 `min=1,max=500`，但没有做 XSS 相关的 HTML 转义。评论内容会存储在 MongoDB 并渲染给其他用户，存在存储型 XSS 风险。
**修复建议**：在接口返回评论时对 HTML 标签进行转义处理，或在前端渲染时做安全处理。

### 3.4 WX API Secret 通过 URL 参数传输
[MEDIUM] api-server/pkg/wx/wx.go:39 — 微信 AppSecret 以 URL 查询参数形式发送：`fmt.Sprintf("...&secret=%s", c.secret)`。URL 可能被 HTTP 代理日志、访问日志等记录，导致密钥泄露。
**修复建议**：使用 POST 方式调用微信接口（微信支持），或在发送前确认使用 HTTPS。目前已使用 HTTPS，但仍存在日志泄露风险。

### 3.5 JWT 无黑名单/刷新机制
[MEDIUM] api-server/pkg/jwt/jwt.go — JWT Token 签发后有效期为 72 小时，但没有任何刷新接口或黑名单机制。Token 泄露后无法撤销，攻击者可一直使用直到过期。
**修复建议**：增加 Token 刷新接口（使用 refresh_token），或维护 Token 黑名单（存 Redis，TTL 与 Token 过期时间一致）。

### 3.6 Media-Server 上传回调可能被重复消费
[MEDIUM] media-server/internal/worker/worker.go:62-83 — 转码任务使用 Redis List（BRPop）消费，没有幂等性保障。如果某个转码任务处理失败但文件已上传 MinIO，重试时会产生重复文件，同时回调 api-server 多次。
**修复建议**：在 api-server 的转码回调接口中做幂等性检查（根据 checkin_id 判断状态），或在 Redis 中使用处理中标记防重入。

### 3.7 内部API密钥使用非安全比较
[MEDIUM] api-server/internal/middleware/internal_auth.go:16 — 使用 `!=` 进行密钥比较，非恒定时间比较。虽然此场景下时序攻击可行性低，但应作为安全编码规范。
**修复建议**：使用 `crypto/subtle.ConstantTimeCompare` 进行密钥比较。

### 3.8 API-Server 日志可能记录敏感信息
[MEDIUM] api-server/internal/middleware/logger.go:18-25 — 日志中间件记录了完整请求路径和查询参数，如果Token或密钥出现在查询参数或路径中，会被记录到日志。
**修复建议**：对日志中的敏感查询参数值做脱敏处理，或只记录路径不记录查询参数。

### 3.9 转码任务失败信息泄露内部路径
[MEDIUM] media-server/internal/worker/worker.go:97-111 — 转码失败时，`err.Error()` 中可能包含临时文件路径（如 `/tmp/wuxie-transcode/xxx/`）等内容，被作为错误信息回调给 api-server 并可能记录到日志或返回给客户端。
**修复建议**：在通知 api-server 时将错误信息替换为通用错误码，详细日志只保留在服务端。

### 3.10 Transcode 回调中的 VideoURL/CoverURL 无校验
[MEDIUM] api-server/internal/handler/handler.go:460-487 — `TranscodeCallback` 接收来自 media-server 的 `VideoURL` 和 `CoverURL`，这些值直接存入数据库并返回给客户端。无格式验证可能导致存储恶意内容路径。
**修复建议**：验证 VideoURL/CoverURL 是否为预期格式（如 `/video/xxx.mp4`），避免存储异常值。

### 3.11 Login 接口 Code 参数缺少长度验证
[MEDIUM] api-server/internal/handler/handler.go:24 — `LoginReq.Code` 使用 `binding:"required"` 但未限制长度。虽然 WeChat code 有固定长度，但攻击者可传入极长字符串消耗服务端资源。
**修复建议**：增加 `max` 验证（如 `binding:"required,max=256"`）。

### 3.12 媒体服务器下载对象前未验证对象是否存在/类型
[MEDIUM] media-server/internal/worker/worker.go:91-100 — Worker 从 MinIO 下载文件时不验证文件类型和大小，直接传入 FFmpeg 转码。如果 `raw` 桶中存在非视频文件，可能导致 FFmpeg 崩溃。
**修复建议**：下载前调用 `StatObject` 检查文件类型和大小，拒绝非视频文件和非合理大小的文件。

---

## 四、LOW 级别问题

### 4.1 服务器运行模式为 debug
[LOW] api-server/configs/config.yaml:3 — `mode: "debug"` 模式下 Gin 框架可能输出更多调试信息（如栈追踪）。
**修复建议**：生产环境设置为 `release`。

### 4.2 Token 解析时缺少签名算法锁定
[LOW] api-server/pkg/jwt/jwt.go:40-42 — `jwt.ParseWithClaims` 默认接受多种签名算法。虽然本实现中使用了 HS256，但若攻击者修改 token 算法为 `none`，解析器可能接受。
**修复建议**：在 Parse 回调中显式验证签名算法：`if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { return nil, fmt.Errorf("unexpected signing method") }`。

### 4.3 In-memory 速率限制器无上限控制
[LOW] api-server/internal/middleware/rate_limit.go:36 — `RateLimiter.visitors` 使用无界 map，持续大量请求下内存会不断增长。虽然没有注册到路由，但如果后续使用，可能存在 OOM 风险。
**修复建议**：限制 map 大小，超过上限时淘汰最早的条目，或使用 Redis 实现分布式限流。

### 4.4 Insight Visibility/Mood 枚举值未做后端校验
[LOW] api-server/internal/handler/insight_handler.go:26,31 — `CreateInsightReq.Visibility` 和 `Mood` 直接从字符串转换，不检查是否是合法枚举值。用户可设置任意值。
**修复建议**：在 service 层增加枚举值校验，不符合预期的值使用默认值或返回错误。

### 4.5 Insight CheckinID/PlanID 解析错误被静默忽略
[LOW] api-server/internal/handler/insight_handler.go:55-63 — 当 `CreateInsightReq.CheckinID` 或 `PlanID` 是无效的 ObjectID 格式时，错误被静默忽略（`err == nil` 才赋值）。这可能导致数据关联不一致。
**修复建议**：即使关联字段是可选的，也应返回错误提示参数格式不正确，或至少记录警告日志。

### 4.6 CommentRepo.Create 使用的 SessionContext 类型强制转换
[LOW] api-server/internal/repository/social_repo.go — `CommentRepo.Create` 使用 `mongo.SessionContext` 而非 `context.Context`，导致非事务上下文无法直接调用。但这一设计是为了事务一致性，属于权衡设计，问题级别较低。

### 4.7 敏感信息硬编码注释提醒
[LOW] api-server/configs/config.yaml:15,19-26 — 配置文件中注释提醒用户修改密钥，但实际部署时容易遗忘。
**修复建议**：使用环境变量配置 `jwt.secret`、`wx.secret`、`media_secret`，配置文件只保留默认值，并在启动时检查是否已修改。

### 4.8 OpenID 未做格式校验直接作为查询条件
[LOW] api-server/internal/service/service.go:92 — 从微信 API 获取的 OpenID 直接用于数据库查询，无格式校验。虽然来源可信，但防御深度不足。
**修复建议**：对 OpenID 做格式校验（长度、字符集），前置防御。

### 4.9 Gin 默认 Recovery 中间件泄露栈信息
[LOW] api-server/internal/router/router.go:36 — 使用 `gin.Recovery()` 中间件，在 debug 模式下会输出完整调用栈到 HTTP 响应。虽然有 InternalError 返回通用消息，但 Recovery 中间件在所有 handler 之前捕获 panic。
**修复建议**：自定义 Recovery 中间件，只返回通用错误，将栈信息写入日志。

### 4.10 打卡查询未限制时间范围
[LOW] api-server/internal/repository/checkin_repo.go:91-122 — `List` 和 `Search` 方法未对查询结果做时间范围限制，随着时间推移数据量增大，可能导致全表扫描和性能问题。
**修复建议**：增加可选的 `start_time` / `end_time` 查询参数，建议默认限制最近 90 天。

---

## 五、问题统计汇总

| 严重级别 | 数量 | 关键影响 |
|---------|------|---------|
| CRITICAL | 6 | CSRF攻击、路径遍历、密钥泄露、Redis/MinIO未授权访问、NoSQL DoS |
| HIGH | 7 | 登录爆破、越权访问、信息泄露、数据一致性、收藏功能缺陷 |
| MEDIUM | 12 | XSS风险、日志泄露、无幂等性、枚举值未校验、时序攻击 |
| LOW | 10 | 配置不严、错误处理、性能风险、防御深度不足 |

**总计：35 个安全问题**

---

## 六、优先修复建议（Top 5）

1. **立即修改所有默认密钥和凭据**：JWT Secret、MinIO 凭据、Redis 密码、内部API Secret 必须改为强随机值，通过环境变量注入。
2. **收紧 CORS 策略**：限制 `Access-Control-Allow-Origin` 为具体白名单，移除或条件性地设置 `Access-Control-Allow-Credentials`。
3. **为 ext 参数增加白名单验证**：防止路径遍历攻击，限制文件扩展名只能为 `mp4`, `jpg`, `png` 等。
4. **修复训练计划报告越权访问**：在 `GetReport` 中添加用户所有权检查。
5. **启用登录速率限制并修复实现**：将 `LoginRateLimit` 注册到路由并确保单例模式。

---

*报告完*