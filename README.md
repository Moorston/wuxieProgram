# 武俱打卡

武俱打卡是专为武术爱好者（下面简称武者）打造的视频训练打卡平台，支持微信小程序、Android、iOS 三端使用。通过视频录制上传、智能转码、广场展示、点赞评论、排行榜、考核组、训练计划、感悟笔记、个人资料库等功能，帮助武者建立系统化的训练追踪体系，让武术训练更科学、更有趣、更有动力。

## 架构

前后端分离架构，支持 Android / iOS / 微信小程序三端。采用双服务分离设计，API Server 处理业务逻辑，Media Server 处理视频上传/转码/播放，互不干扰。

---

## 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| 前端 | uni-app (Vue3 + TypeScript) | 一套代码编译到 Android/iOS/小程序 |
| 后端API | Go + Gin | 业务逻辑服务 |
| 视频服务 | Go + Gin + FFmpeg | 视频上传/转码/播放服务 |
| 数据库 | MongoDB | 文档型数据库 |
| 缓存/队列 | Redis | 缓存 + 转码任务队列 |
| 对象存储 | MinIO | 自建视频/图片/资料存储 |
| 反向代理 | Nginx | 统一入口 + 视频流代理 + 静态资源缓存 |
| 部署 | Docker Compose | 容器化一键部署 + deploy.sh 脚本 |

---

## 项目结构

```
wuxieProgram/
├── api-server/                          # 业务API服务 (端口8080)
│   ├── cmd/main.go                      # 服务入口 + 定时任务启动
│   ├── configs/config.yaml              # 配置文件
│   ├── Dockerfile
│   ├── go.mod
│   ├── internal/
│   │   ├── config/config.go             # 配置加载
│   │   ├── handler/
│   │   │   ├── handler.go               # 打卡/用户/社交处理器
│   │   │   ├── training_handler.go      # 训练计划处理器
│   │   │   ├── insight_handler.go       # 感悟笔记处理器
│   │   │   ├── notification_handler.go  # 消息通知处理器
│   │   │   └── resource_handler.go      # 个人资料库处理器
│   │   ├── middleware/
│   │   │   ├── auth.go                  # JWT鉴权中间件
│   │   │   ├── cors.go                  # CORS + 请求ID
│   │   │   └── logger.go               # 请求日志
│   │   ├── model/
│   │   │   ├── user.go                  # 用户模型
│   │   │   ├── checkin.go               # 打卡/评论/点赞模型
│   │   │   ├── group.go                 # 考核组模型
│   │   │   ├── rank.go                  # 排行榜模型
│   │   │   ├── training.go              # 训练计划/模板模型
│   │   │   ├── insight.go               # 感悟笔记/标签模型
│   │   │   ├── notification.go          # 通知/通知设置模型
│   │   │   └── resource.go              # 资料库/标签模型
│   │   ├── repository/
│   │   │   ├── user_repo.go
│   │   │   ├── checkin_repo.go
│   │   │   ├── social_repo.go           # 评论/点赞(含EnsureIndexes)
│   │   │   ├── rank_repo.go
│   │   │   ├── training_repo.go
│   │   │   ├── template_repo.go
│   │   │   ├── insight_repo.go          # 感悟+点赞(含InsightLikeRepo)
│   │   │   ├── insight_tag_repo.go
│   │   │   ├── notification_repo.go     # 通知+设置(含FindOneAndUpdate)
│   │   │   ├── resource_repo.go
│   │   │   └── resource_tag_repo.go
│   │   ├── router/router.go             # 路由注册 (60+ 端点)
│   │   └── service/
│   │       ├── service.go               # 核心业务逻辑
│   │       ├── training_service.go      # 训练计划逻辑
│   │       ├── insight_service.go       # 感悟笔记逻辑
│   │       ├── notification_service.go  # 通知逻辑
│   │       ├── resource_service.go      # 资料库逻辑
│   │       ├── cron.go                  # 定时任务(排行榜刷新+训练提醒)
│   │       └── utils.go                 # 工具函数(正则清洗)
│   └── pkg/
│       ├── jwt/jwt.go
│       ├── response/response.go
│       └── wx/wx.go                     # 微信订阅消息客户端
│
├── media-server/                        # 视频服务 (端口8081)
│   ├── cmd/main.go                      # 服务入口
│   ├── configs/config.yaml
│   ├── Dockerfile
│   ├── go.mod
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── handler/handler.go           # 上传预签名/回调/视频URL
│   │   ├── router/router.go
│   │   └── worker/worker.go             # FFmpeg转码Worker
│   └── pkg/
│       ├── ffmpeg/ffmpeg.go             # FFmpeg封装
│       ├── minio/minio.go               # MinIO封装(含RemoveObject/ListObjects)
│       └── response/response.go
│
├── client/                              # uni-app前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── src/
│       ├── api/index.ts                 # 接口封装 (80+ API方法)
│       ├── store/
│       │   ├── user.ts
│       │   └── checkin.ts
│       ├── utils/request.ts             # HTTP请求封装(含mediaRequest鉴权)
│       ├── pages.json                   # 路由配置 (30+页面 + 5个TabBar)
│       ├── static/                      # 静态资源(tab图标、头像、Logo)
│       ├── App.vue
│       ├── main.ts
│       └── pages/
│           ├── index/index.vue          # 首页(含通知铃铛)
│           ├── square/square.vue        # 广场(搜索+Tab筛选+分页)
│           ├── checkin/checkin.vue      # 打卡
│           ├── rank/rank.vue            # 排位(分页)
│           ├── mine/mine.vue            # 我的(编辑资料)
│           ├── login/login.vue          # 登录
│           ├── video-detail/video-detail.vue  # 视频详情
│           ├── my-video/my-video.vue          # 我的视频(分页)
│           ├── group/
│           │   ├── group.vue            # 考核组列表
│           │   └── detail.vue           # 考核组详情
│           ├── training/
│           │   ├── list.vue             # 训练计划列表
│           │   ├── create.vue           # 创建计划
│           │   ├── detail.vue           # 计划详情
│           │   ├── template.vue         # 模板列表
│           │   ├── template-detail.vue  # 模板详情
│           │   ├── today.vue            # 今日任务
│           │   └── report.vue           # 训练报告
│           ├── insight/
│           │   ├── list.vue             # 感悟时间线
│           │   ├── create.vue           # 写感悟
│           │   ├── detail.vue           # 感悟详情
│           │   ├── tags.vue             # 标签管理
│           │   ├── mood.vue             # 心情统计
│           │   ├── on-this-day.vue      # 历史今日
│           │   └── public.vue           # 公开感悟广场
│           ├── notification/
│           │   ├── list.vue             # 通知列表
│           │   └── settings.vue         # 通知设置
│           └── resource/
│               ├── list.vue             # 资料库(瀑布流/列表)
│               ├── upload.vue           # 上传资料
│               ├── detail.vue           # 资料详情
│               ├── favorites.vue        # 收藏列表
│               ├── share.vue            # 分享设置
│               ├── tags.vue             # 资料标签
│               └── stats.vue            # 存储统计
│
├── deploy/                              # 部署配置
│   ├── docker-compose.yml               # 6个服务编排(含health checks)
│   ├── nginx.conf                       # Nginx配置(含resource代理)
│   ├── mongo-init.js                    # MongoDB初始化(15+集合)
│   ├── deploy.sh                        # 一键部署脚本
│   └── .env.example                     # 环境变量模板
│
└── README.md
```

---

## 已实现功能清单

### 一、用户系统 ✅

| 功能 | 说明 |
|------|------|
| 微信登录 | 小程序端调用 `wx.login` 获取 code，后端换取 openid，返回 JWT |
| JWT鉴权 | 登录后生成 JWT Token，前端存储并附加到每次请求 Header |
| 用户信息获取 | 获取个人资料（昵称/头像/积分/打卡天数） |
| 用户信息更新 | 支持修改昵称和头像（mine页弹窗编辑） |
| 登录态管理 | Pinia Store 管理 token 持久化和登录状态 |
| 自动登录检测 | 应用启动时自动加载本地 token |
| 退出登录 | 清除本地 token，reLaunch 跳转登录页 |

### 二、打卡系统 ✅

| 功能 | 说明 |
|------|------|
| 视频录制/选择 | 调用系统相机录制（最长60秒）或从相册选择，自动压缩 |
| 打卡准备 | 创建打卡记录（状态：待转码），返回 checkin_id |
| 预签名上传 | 客户端获取 MinIO 预签名 PUT URL，视频直传 MinIO |
| 上传进度 | 实时显示上传百分比进度条 |
| 上传回调 | 上传完成后通知 Media Server，加入 Redis 转码队列 |
| 异步转码 | Worker 消费队列，FFmpeg 转码为 H.264 MP4（1280p, CRF28） |
| 转码完成回调 | 转码完成后回调 API Server 更新记录状态 |
| 打卡描述 | 提交打卡时可填写训练描述（最多200字） |
| 打卡积分 | 每次打卡默认获得10积分 |
| 打卡详情查询 | 支持单条打卡记录详情查询（含点赞状态） |

### 三、广场系统 ✅

| 功能 | 说明 |
|------|------|
| 瀑布流展示 | 双列瀑布流布局，展示封面图+描述+作者+点赞数 |
| 搜索功能 | 顶部搜索框，支持关键词搜索打卡描述（正则清洗防注入） |
| 考核组Tab筛选 | 切换"考核组"Tab按 group_id 过滤组内打卡 |
| 分页加载 | 支持分页查询 + 下拉刷新 + 上拉加载更多 |
| 用户信息联查 | 列表接口自动填充打卡者的昵称和头像 |
| 点赞状态 | 列表接口批量查询当前用户对每条记录的点赞状态 |

### 四、社交系统 ✅

| 功能 | 说明 |
|------|------|
| 点赞/取消点赞 | Toggle 模式，唯一索引防重复，同步更新 like_count |
| 发表评论 | 对打卡记录发表文字评论 |
| 评论列表 | 分页展示评论列表，包含评论者昵称和头像 |
| 评论计数 | 发表评论自动递增 comment_count |

### 五、排行榜系统 ✅

| 功能 | 说明 |
|------|------|
| 排行榜查询 | 支持按时间段筛选：今日(day)/本周(week)/总榜(all) |
| 我的排名 | 查询当前用户在指定时间段的排名和积分 |
| 排行榜缓存 | 使用 rank_cache 集合存储计算好的排名数据 |
| 自动刷新 | 后端 cron 每10分钟自动重算日榜/周榜/总榜 |
| 分页加载 | 排行榜支持分页查询 |

### 六、考核组系统 ✅

| 功能 | 说明 |
|------|------|
| 考核组列表 | 展示所有考核组，包含名称、描述、人数、成员头像 |
| 考核组详情页 | 展示组信息、组长、成员列表（含积分） |
| 成员信息联查 | 自动查询组内成员的昵称和头像 |

### 七、视频播放 ✅

| 功能 | 说明 |
|------|------|
| 视频详情页 | 视频播放器 + 作者信息 + 描述 + 点赞 + 评论列表 |
| 视频播放URL | 通过 Media Server 获取 MinIO 预签名 GET URL（有效期2小时） |
| 我的视频时间线 | 按时间线展示打卡记录，支持分页 |
| Nginx视频流代理 | 支持 Range 请求，实现视频拖拽播放 |
| 封面图缓存 | Nginx 对封面图进行7天缓存 |

### 八、视频转码（Media Server）✅

| 功能 | 说明 |
|------|------|
| MinIO桶自动创建 | 启动时自动创建 raw/video/cover/resource 四个桶 |
| 视频信息探测 | FFmpeg 探测视频时长、分辨率等信息 |
| H.264转码 | 将任意格式视频转码为 H.264 MP4，限制1280p，CRF28压缩 |
| 封面提取 | 从转码后视频第1秒提取帧作为封面图（JPEG） |
| faststart优化 | 使用 `-movflags +faststart` 使MP4支持边下边播 |
| 多Worker并发 | 支持配置多个转码 Worker 并发消费队列 |
| Redis队列 | 使用 Redis List 作为转码任务队列，BRPop 阻塞消费 |
| 健康检查 | `/health` 端点用于服务健康监测 |

### 九、训练计划模块 ✅

| 功能 | 说明 |
|------|------|
| 创建训练计划 | 自定义名称/描述/周期，按天添加训练项目（基本功/套路/散打/气功） |
| 计划列表 | 展示所有计划，支持按状态筛选（全部/进行中/已完成/草稿） |
| 计划详情 | 可折叠展开每日任务列表，支持完成/跳过操作 |
| 训练模板 | 预设模板浏览，支持按拳种/难度筛选 |
| 应用模板 | 一键基于模板创建计划，自动计算日期和任务 |
| 今日任务 | 自动计算当前应完成的训练项，支持完成/跳过 |
| 训练报告 | 完成率统计、训练类型分布、进度可视化 |
| 微信订阅消息提醒 | 定时任务检查未完成任务，推送微信订阅消息 |
| Ownership检查 | 计划的查看/编辑/任务更新均验证用户归属 |

### 十、消息通知系统 ✅

| 功能 | 说明 |
|------|------|
| 点赞通知 | 有人点赞了你的打卡时创建通知 |
| 评论通知 | 有人评论了你的打卡时创建通知 |
| 训练提醒 | 每日定时检查未完成任务，创建提醒通知 |
| 计划完成通知 | 训练计划全部完成时自动通知 |
| 通知列表 | 按时间分组展示（今天/昨天/更早），支持分页 |
| 未读计数 | 首页铃铛图标显示未读数 |
| 标记已读 | 支持单条已读和全部已读 |
| 通知设置 | 各类型通知开关 + 提醒时间设置 |
| 点击跳转 | 通知点击自动跳转到相关页面 |
| 隐私控制 | 可独立控制各类型通知的开关 |

### 十一、感悟笔记模块 ✅

| 功能 | 说明 |
|------|------|
| 创建感悟 | 支持文字记录（最多2000字）、图片配图（最多9张）、心情标记 |
| 5种心情 | 突破🔥 / 满意😊 / 一般😐 / 困惑🤔 / 低落😔 |
| 标签管理 | 自定义标签 + 标签聚合统计 + 按标签筛选 |
| 隐私控制 | 默认私密，可选择公开到广场 |
| 感悟时间线 | 按时间分组展示，支持心情/标签筛选 |
| 感悟详情 | 完整内容 + 图片预览 + 关联打卡/计划 |
| 心情统计 | 近30天心情分布柱状图 + 总结 |
| 历史今日 | 查看"一年前的今天"的感悟 |
| 公开感悟广场 | 展示所有公开感悟 |
| per-user点赞 | 感悟点赞支持每人一次 + 唯一索引 |

### 十二、个人资料库模块 ✅

| 功能 | 说明 |
|------|------|
| 资料上传 | 支持视频/图片/文档上传，预签名URL直传MinIO |
| 资料浏览 | 瀑布流/列表视图切换 |
| 多维筛选 | 按类型/分类/难度/标签/关键词筛选 |
| 排序 | 按时间/名称/大小排序 |
| 资料详情 | 视频播放/图片预览/文档展示 |
| 收藏功能 | 一键收藏/取消收藏，收藏列表 |
| 分享功能 | 分享范围设置（仅自己/考核组/公开） |
| 标签管理 | 自定义标签 + 标签聚合 |
| 存储统计 | 已用/总配额(5GB)可视化 + 类型分布 |
| Ownership检查 | 资料查看/更新/删除均验证可见性 |
| 资料库桶 | 新增 resource 桶，MinIO 自动创建 |

### 十三、基础设施与部署 ✅

| 功能 | 说明 |
|------|------|
| Docker Compose编排 | 一键部署 6 个服务（含 health checks + 网络隔离） |
| deploy.sh 部署脚本 | 支持 up/down/restart/logs/status/rebuild |
| MongoDB索引优化 | 15+ 集合，30+ 索引（含唯一索引、复合索引） |
| Nginx统一入口 | 所有服务通过 Nginx 80 端口统一代理 |
| CORS跨域 | 反射 Origin + credentials 支持 |
| 请求日志 | 结构化日志记录每次请求 |
| 请求ID | 每个请求生成唯一 X-Request-ID |
| Panic恢复 | Gin Recovery 中间件防止服务崩溃 |
| 排行榜定时刷新 | cron 每10分钟自动重算排名 |
| 训练提醒定时任务 | 每天固定时间检查并推送提醒 |

### 十四、前端工程化 ✅

| 功能 | 说明 |
|------|------|
| TypeScript全量类型 | 所有页面和工具函数使用 TypeScript 编写 |
| Pinia状态管理 | 用户状态和打卡列表状态分离管理 |
| 请求统一封装 | 统一的 request 函数，自动附加 Token、401 reLaunch |
| 双服务请求分离 | API 请求和 Media 请求分别指向不同服务地址 |
| TabBar五页签 | 首页/广场/打卡/排位/我的 五个底部导航 |
| 30+ 页面 | 覆盖所有功能模块 |
| 80+ API方法 | 覆盖所有后端接口 |
| 静态资源 | Tab图标、默认头像、Logo |
| 下拉刷新/分页加载 | 所有列表页面支持 |
| rpx响应式布局 | 适配不同屏幕尺寸 |

---

## API 接口清单

### 公开接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/auth/login` | POST | 微信登录 |

### 用户接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/user/profile` | GET | 获取个人信息 |
| `/api/user/profile` | PUT | 更新个人信息 |

### 打卡接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/checkin/prepare` | POST | 准备打卡 |
| `/api/checkin/list` | GET | 广场列表（支持 group_id 筛选） |
| `/api/checkin/mine` | GET | 我的打卡记录 |
| `/api/checkin/search` | GET | 搜索打卡（关键词 + 正则清洗） |
| `/api/checkin/:id` | GET | 打卡详情 |
| `/api/checkin/:id` | DELETE | 删除打卡 |
| `/api/checkin/:id/like` | POST | 点赞/取消 |
| `/api/checkin/:id/comment` | POST | 发表评论 |
| `/api/checkin/:id/comments` | GET | 评论列表 |

### 排行榜接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/rank` | GET | 排行榜（自动刷新） |
| `/api/rank/me` | GET | 我的排名 |

### 考核组接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/group/list` | GET | 考核组列表 |
| `/api/group/:id` | GET | 考核组详情 |

### 训练计划接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/training/plan` | POST | 创建计划 |
| `/api/training/plans` | GET | 计划列表（按状态筛选） |
| `/api/training/plan/:id` | GET | 计划详情（ownership检查） |
| `/api/training/plan/:id` | PUT | 更新计划 |
| `/api/training/plan/:id` | DELETE | 删除计划 |
| `/api/training/today` | GET | 今日任务 |
| `/api/training/task/:plan_id/:day/:task_idx` | PUT | 更新任务状态 |
| `/api/training/plan/:id/report` | GET | 训练报告 |
| `/api/training/template/list` | GET | 模板列表 |
| `/api/training/template/:id` | GET | 模板详情 |
| `/api/training/template/:id/apply` | POST | 应用模板 |

### 通知接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/notification/list` | GET | 通知列表 |
| `/api/notification/unread` | GET | 未读数量 |
| `/api/notification/read/:id` | PUT | 标记已读 |
| `/api/notification/read-all` | PUT | 全部已读 |
| `/api/notification/:id` | DELETE | 删除通知 |
| `/api/notification/settings` | GET | 通知设置 |
| `/api/notification/settings` | PUT | 更新设置 |

### 感悟笔记接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/insight` | POST | 创建感悟 |
| `/api/insight/list` | GET | 感悟列表（按心情/标签筛选） |
| `/api/insight/public` | GET | 公开感悟广场 |
| `/api/insight/tags` | GET | 标签列表 |
| `/api/insight/mood-stats` | GET | 心情统计 |
| `/api/insight/on-this-day` | GET | 历史今日 |
| `/api/insight/:id` | GET | 感悟详情（visibility检查） |
| `/api/insight/:id` | PUT | 更新感悟 |
| `/api/insight/:id` | DELETE | 删除感悟 |
| `/api/insight/:id/like` | POST | 点赞（per-user） |

### 个人资料库接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/resource/upload/presign` | GET | 获取上传预签名URL |
| `/api/resource/upload/callback` | POST | 上传完成回调 |
| `/api/resource` | POST | 创建资料记录 |
| `/api/resource/list` | GET | 资料列表（多维筛选+搜索） |
| `/api/resource/tags` | GET | 标签列表 |
| `/api/resource/stats` | GET | 存储统计 |
| `/api/resource/favorites` | GET | 收藏列表 |
| `/api/resource/:id` | GET | 资料详情（shareScope检查） |
| `/api/resource/:id` | PUT | 更新资料 |
| `/api/resource/:id` | DELETE | 删除资料 |
| `/api/resource/:id/favorite` | POST | 收藏/取消 |

### Media Server 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/media/upload/presign` | GET | 获取预签名上传URL |
| `/media/upload/callback` | POST | 上传完成回调 |
| `/media/url` | GET | 获取视频播放URL |

---

## 数据库设计

### 集合总览

| 集合 | 说明 | 索引数 |
|------|------|--------|
| users | 用户 | 3 |
| checkins | 打卡记录 | 3 |
| comments | 评论 | 2 |
| likes | 点赞 | 2 (含唯一索引) |
| groups | 考核组 | 1 |
| rank_cache | 排行榜缓存 | 2 |
| training_plans | 训练计划 | 4 |
| training_templates | 训练模板 | 3 |
| notifications | 通知 | 3 |
| notification_settings | 通知设置 | 1 (唯一) |
| insights | 感悟笔记 | 4 |
| insight_tags | 感悟标签 | 2 (含唯一索引) |
| insight_likes | 感悟点赞 | 1 (唯一) |
| resources | 个人资料 | 5 |
| resource_tags | 资料标签 | 2 (含唯一索引) |

### 关键集合字段

<details>
<summary>点击展开详细字段</summary>

**users**: openid(唯一), nickname, avatar, gender, group_id, score, check_days

**checkins**: user_id, video_url, cover_url, description, duration, score, status(0-3), like_count, comment_count

**training_plans**: user_id, group_id, title, start_date, end_date, status(0-3), days[{day, date, tasks[{title, type, duration, reps, status}]}], stats

**insights**: user_id, content, images[], mood, tags[], checkin_id, plan_id, visibility, like_count

**resources**: user_id, title, type(video/image/document), category, tags[], difficulty, file_url, file_size, share_scope, group_id, is_favorite

</details>

---

## 快速启动

### 方式一：Docker 一键部署（推荐）

```bash
cd deploy
./deploy.sh up
```

### 方式二：本地开发

```bash
# 1. 启动基础设施
cd deploy
docker-compose up -d mongo redis minio

# 2. 启动 API Server
cd api-server
go run cmd/main.go

# 3. 启动 Media Server（需要本地安装 FFmpeg）
cd media-server
go run cmd/main.go

# 4. 启动前端
cd client
npm install
npm run dev:mp-weixin    # 微信小程序
npm run dev:h5           # H5
npm run dev:app          # App
```

### 服务访问

| 服务 | 地址 | 说明 |
|------|------|------|
| Nginx | http://localhost | 统一入口 |
| API Server | http://localhost:8080 | 业务接口 |
| Media Server | http://localhost:8081 | 视频服务 |
| MinIO Console | http://localhost:9001 | 存储管理 |
| MongoDB | localhost:27017 | 数据库 |
| Redis | localhost:6379 | 缓存/队列 |

---

## 配置说明

### api-server/configs/config.yaml

```yaml
server:
  port: "8080"
  mode: "debug"          # debug/release

mongo:
  uri: "mongodb://mongo:27017"
  database: "wuxie"

redis:
  addr: "redis:6379"
  password: ""
  db: 0

jwt:
  secret: "wuxie-jwt-secret-change-in-production"
  expires: 72

wx:
  app_id: "your-wx-app-id"
  secret: "your-wx-app-secret"
  template_id: "your-subscribe-template-id"
  remind_hour: 20

media_url: "http://media-server:8081"
```

### 部署脚本

```bash
./deploy.sh up       # 启动
./deploy.sh down     # 停止
./deploy.sh restart  # 重启
./deploy.sh logs     # 查看日志
./deploy.sh status   # 查看状态
./deploy.sh rebuild  # 强制重建
```

---

## 后续规划功能

### AI 视频评分系统（P2/P3）

| 能力 | 说明 | 状态 |
|------|------|------|
| 动作规范性评估 | 识别套路动作的标准度、连贯性、完整性 | ❌ 待实现 |
| 力度与劲力分析 | 评估发力方式、劲力传导、爆发力表现 | ❌ 待实现 |
| 身法与步法评价 | 分析重心转换、身法协调、步法稳定性 | ❌ 待实现 |
| 节奏与速度评分 | 评估动作节奏感、快慢相间、停顿控制 | ❌ 待实现 |
| 综合评分 | 加权计算总分（0-100），生成等级（S/A/B/C/D） | ❌ 待实现 |
| 改进建议 | 针对薄弱环节生成文字建议 | ❌ 待实现 |

**推荐路径**：方案A（通用视频理解模型 + Prompt）快速验证 → 方案B（LoRA微调）提升 → 方案C（关键点检测）长期目标。

---

## 功能完成度

| 优先级 | 模块 | 状态 |
|--------|------|------|
| **P0** | 核心打卡 + 广场 + 排行榜 + 社交 | ✅ 已完成 |
| **P1** | 训练计划模块 | ✅ 已完成 |
| **P1** | 感悟笔记模块 | ✅ 已完成 |
| **P1** | 消息通知系统 | ✅ 已完成 |
| **P2** | 个人资料库模块 | ✅ 已完成 |
| **P2** | AI 视频评分（方案A） | ❌ 待实现 |
| **P3** | AI 视频评分（方案B） | ❌ 待实现 |
