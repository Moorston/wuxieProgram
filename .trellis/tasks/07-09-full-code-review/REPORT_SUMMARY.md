# 武俱打卡项目全面审查综合报告

> 审查日期：2026-07-09
> 覆盖维度：代码结构 / 代码质量 / Bug / 架构 / 安全
> 审查文件：~80 个源文件（Go + TypeScript/Vue）

---

## 总体统计

| 审查维度 | CRITICAL | HIGH | MEDIUM | LOW | 合计 |
|---------|:-------:|:----:|:-----:|:---:|:---:|
| 🔒 安全审查 | 6 | 7 | 12 | 10 | **35** |
| 🎨 前端审查 | 6 | 9 | 11 | 5 | **31** |
| 🔧 后端质量 | 2 | 5 | 10 | 11 | **28** |
| 🏗️ 架构审查 | 0 | 1 | 4 | 10 | **15** |
| **合计** | **14** | **22** | **37** | **36** | **109** |

---

## 一、🔴 CRITICAL 问题（14 个，必须立即修复）

### 安全 - 前 3 紧急

| # | 问题 | 文件 | 风险 |
|---|------|------|------|
| S1 | CORS 配置回显 Origin + Credentials | `cors.go:13-21` | CSRF 攻击 |
| S2 | ext 参数无白名单验证 | `handler.go:35-36` | MinIO 路径遍历 |
| S3 | 默认密钥/凭据硬编码（JWT/WX/MinIO/Redis） | `config.yaml` 多处 | 完全伪造/越权 |
| S4 | 资源搜索 $regex 无防护 | `resource_repo.go:84-88` | ReDoS 拒绝服务 |

### 后端质量 - 运行时崩溃

| # | 问题 | 文件 | 风险 |
|---|------|------|------|
| B1 | nil 指针解引用 (`GetPlan` 失败后访问 `plan.UserID`) | `training_handler.go:160` | **运行时 panic** |
| B2 | 相同模式在 `UpdateTask` 中重复 | `training_handler.go:241` | **运行时 panic** |

### 前端 - 功能完全不可用

| # | 问题 | 文件 | 风险 |
|---|------|------|------|
| F1 | 头像上传发送本地临时路径给后端 | `mine.vue:126-128` | **功能完全不可用** |
| F2 | 硬编码 `localhost:8080` 无法部署生产 | `request.ts:1,54` | **无法部署** |
| F3 | 100% 使用 `any` 类型 | 全部 31 个文件 | 类型安全形同虚设 |
| F4 | URL 参数未 `encodeURIComponent` | `checkin.vue:65`, `api/index.ts:202` | XSS 风险 |
| F5 | 使用已弃用的 `uni.getUserProfile` | `login/login.vue:26-32` | 新版微信不可用 |
| F6 | 无 Token 刷新机制 | `store/user.ts:17-21` | 体验差 |

---

## 二、🟠 HIGH 问题（22 个，尽快修复）

### 安全（7 个）

| # | 问题 | 文件 |
|---|------|------|
| S5 | 登录接口无速率限制（`LoginRateLimit` 未注册） | `rate_limit.go:96-109` |
| S6 | 训练计划报告无所有权检查 | `training_handler.go:283-296` |
| S7 | Media Server 直接暴露 MinIO 错误给客户端 | `handler.go:40,105` |
| S8 | Insight 点赞缺少事务一致性 | `insight_repo.go:216-235` |
| S9 | 收藏功能错误限制为资源所有者自己 | `resource_repo.go:140` |
| S10 | FFmpeg probe 命令名拼写错误 `"ffmpegprobe"` | `ffmpeg.go:28` |
| S11 | 用户文本字段缺少长度限制 | 多处 handler |

### 后端质量（5 个）

| # | 问题 | 文件 |
|---|------|------|
| B3 | `sanitizeRegex` 转义函数失效（`\$&` 输出字面量） | `checkin_repo.go:20` |
| B4 | JWT 解析未限制签名算法 | `jwt.go:40` |
| B5 | `mongo.Disconnect(context.Background())` 无超时 | `main.go:210` |
| B6 | 配置加载无验证（缺失关键配置不报错） | `config.go:62` |
| B7 | 资源配额检查 TOCTOU 竞态条件 | `resource_service.go:37-56` |

### 前端（9 个）

| # | 问题 | 文件 |
|---|------|------|
| F7 | 分页逻辑在 10+ 页面完全重复 | 10 个页面 |
| F8 | 所有 `catch (e) {}` 吞噬异常 | 全部文件 |
| F9 | `formatTime` 在 4 个页面重复实现 | `my-video.vue`, `insight/list.vue` 等 |
| F10 | `formatSize` 在 5 个页面重复实现 | `resource/*.vue` 全部 |
| F11 | 类型字典常量重复定义（`typeMap`/`typeIcon`/`moodIcon`） | 10+ 处 |
| F12 | 未使用的 `onMounted` 导入 | 7 个文件 |
| F13 | 图片未使用 `lazy-load` | 全部列表页 |
| F14 | 资源上传只支持视频但声称支持图片/文档 | `resource/upload.vue:113-123` |
| F15 | 上传回调传空 `checkin_id` | `resource/upload.vue:163` |

### 架构（1 个）

| # | 问题 | 文件 |
|---|------|------|
| A1 | `handler.go`(487行) 和 `service.go`(463行) 文件过大 | 违反自身规范 |

---

## 三、🟡 MEDIUM 问题（37 个，计划修复）

### 安全 - 12 个（详见 security 报告）
关键包括：XSS（评论未转义）、WX Secret URL 参数泄露、JWT 无黑名单、内部密钥非恒定时间比较、转码缺乏幂等性等。

### 后端质量 - 10 个
关键包括：`BatchIsLiked` 错误静默忽略、微信 API 错误忽略、速率限制器 mutex 无 defer、goroutine 泄漏、训练计划服务层权限校验缺失等。

### 前端 - 11 个
关键包括：瀑布流按奇偶分割、`v-for` 使用 index 作 key、排行榜响应结构不确定性、评论不支持分页、`uni.request` 无超时等。

### 架构 - 4 个
关键包括：资料库上传绕过 media-server（职责交叉）、`NotificationService` 横向耦合、`main.go` 依赖注入膨胀、定时任务调度实现原始。

---

## 四、🟢 LOW 问题（36 个，建议改进）

包括魔法数字、配置 debug 模式、弃用 API、CSS 渐变重复、emoji 硬编码、注释缺失等。

---

## 五、优先修复建议（Top 10）

### 立即修复（P0）

| 优先级 | 问题 | 影响 | 时间 |
|--------|------|------|------|
| **P0-1** | `training_handler.go:160,241` nil 指针解引用 | 运行时 panic | 10 分钟 |
| **P0-2** | `cors.go:13-21` CORS 回显 Origin | CSRF 攻击 | 10 分钟 |
| **P0-3** | `mine.vue:126-128` 头像上传不可用 | 功能完全无法使用 | 20 分钟 |
| **P0-4** | `checkin_repo.go:20` `sanitizeRegex` 失效 | ReDoS 防护无效 | 5 分钟 |
| **P0-5** | 默认密钥/凭据硬编码 | 完全伪造/越权 | 30 分钟（+部署） |

### 尽快修复（P1）

| 优先级 | 问题 | 影响 | 时间 |
|--------|------|------|------|
| **P1-6** | `handler.go:160` 等 ext 参数无白名单 | MinIO 路径遍历 | 10 分钟 |
| **P1-7** | `request.ts` 硬编码 localhost | 无法部署生产 | 15 分钟 |
| **P1-8** | `training_handler.go:283` 计划报告越权 | 任意用户可查看 | 10 分钟 |
| **P1-9** | `handler.go`/`service.go` 文件过大拆分 | 违反自身规范 | 30 分钟 |
| **P1-10** | `LoginRateLimit` 注册到路由 | 登录爆破无防护 | 5 分钟 |

---

## 六、各维度详细报告

| 报告 | 路径 |
|------|------|
| 🔒 安全审查报告 | `.trellis/tasks/07-09-full-code-review/review-security.md` |
| 🔧 后端质量审查报告 | `.trellis/tasks/07-09-full-code-review/review-backend-quality.md` |
| 🎨 前端审查报告 | `.trellis/tasks/07-09-full-code-review/review-frontend.md` |
| 🏗️ 架构审查报告 | `.trellis/tasks/07-09-full-code-review/review-architecture.md` |

---

*报告生成时间：2026-07-09 09:21*