# 武俱打卡小程序

武俱打卡是专为武术爱好者打造的视频训练打卡平台，支持微信小程序、Android、iOS 三端使用。通过视频录制上传、智能转码、广场展示、点赞评论、排行榜和考核组等功能，帮助武馆教练和学员建立系统化的训练追踪体系，让武术训练更科学、更有趣、更有动力。

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
| 对象存储 | MinIO | 自建视频/图片存储 |
| 反向代理 | Nginx | 统一入口 + 视频流代理 + 静态资源缓存 |
| 部署 | Docker Compose | 容器化一键部署 |

---

## 项目结构

```
wuxieProgram/
├── api-server/                      # 业务API服务 (端口8080)
│   ├── cmd/main.go                  # 服务入口
│   ├── configs/config.yaml          # 配置文件
│   ├── Dockerfile
│   ├── go.mod
│   ├── internal/
│   │   ├── config/config.go         # 配置加载 (Server/Mongo/Redis/JWT/WX/MediaURL)
│   │   ├── handler/handler.go       # HTTP处理器
│   │   ├── middleware/
│   │   │   ├── auth.go              # JWT鉴权中间件
│   │   │   ├── cors.go              # 跨域中间件 + 请求ID
│   │   │   └── logger.go            # 请求日志中间件
│   │   ├── model/
│   │   │   ├── user.go              # 用户模型
│   │   │   ├── checkin.go           # 打卡/评论/点赞模型
│   │   │   ├── group.go             # 考核组模型
│   │   │   └── rank.go              # 排行榜模型
│   │   ├── repository/
│   │   │   ├── user_repo.go         # 用户数据访问
│   │   │   ├── checkin_repo.go      # 打卡数据访问
│   │   │   ├── social_repo.go       # 评论/点赞数据访问
│   │   │   └── rank_repo.go         # 排行榜/考核组数据访问
│   │   ├── router/router.go         # 路由注册
│   │   └── service/service.go       # 业务逻辑层
│   └── pkg/
│       ├── jwt/jwt.go               # JWT生成/解析
│       └── response/response.go     # 统一响应格式
│
├── media-server/                    # 视频服务 (端口8081)
│   ├── cmd/main.go                  # 服务入口
│   ├── configs/config.yaml          # 配置文件
│   ├── Dockerfile
│   ├── go.mod
│   ├── internal/
│   │   ├── config/config.go         # 配置加载 (Server/MinIO/Redis/FFmpeg/APIServer)
│   │   ├── handler/handler.go       # 上传预签名/回调/视频URL/健康检查
│   │   ├── router/router.go         # 路由注册
│   │   └── worker/worker.go         # FFmpeg转码Worker (消费Redis队列)
│   └── pkg/
│       ├── ffmpeg/ffmpeg.go         # FFmpeg调用封装 (Probe/Transcode)
│       ├── minio/minio.go           # MinIO客户端封装 (预签名/上传/下载/桶管理)
│       └── response/response.go     # 统一响应格式
│
├── client/                          # uni-app前端
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── src/
│       ├── api/index.ts             # 接口封装 (15个API方法)
│       ├── store/
│       │   ├── user.ts              # 用户状态 (token/登录态)
│       │   └── checkin.ts           # 打卡列表状态 (分页/加载)
│       ├── utils/request.ts         # HTTP请求封装 (鉴权/错误处理/双服务URL)
│       ├── pages.json               # 路由配置 (9个页面 + 5个TabBar)
│       ├── manifest.json            # 应用配置
│       ├── uni.scss                 # 全局样式变量
│       ├── App.vue                  # 应用入口
│       ├── main.ts                  # 创建App实例 + Pinia
│       └── pages/
│           ├── index/index.vue      # 首页
│           ├── square/square.vue    # 广场
│           ├── checkin/checkin.vue  # 打卡
│           ├── rank/rank.vue        # 排位
│           ├── mine/mine.vue        # 我的
│           ├── login/login.vue      # 登录
│           ├── video-detail/video-detail.vue  # 视频详情
│           ├── my-video/my-video.vue          # 我的视频
│           └── group/group.vue                # 考核组
│
├── deploy/                          # 部署配置
│   ├── docker-compose.yml           # 6个服务编排
│   ├── nginx.conf                   # Nginx配置
│   └── mongo-init.js                # MongoDB初始化脚本
│
└── README.md
```

---

## 已实现功能清单

### 一、用户系统

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 微信登录 | 小程序端调用 `wx.login` 获取 code，后端调用微信接口换取 openid，创建/查找用户，返回 JWT | `login.vue`, `handler.go:Login`, `service.go:WXLogin` |
| JWT鉴权 | 登录后生成 JWT Token，前端存储并附加到每次请求 Header，中间件校验 | `jwt.go`, `middleware/auth.go`, `utils/request.ts` |
| 用户信息获取 | 登录后获取个人资料（昵称/头像/积分/打卡天数） | `mine.vue`, `handler.go:GetProfile` |
| 用户信息更新 | 支持修改昵称和头像 | `handler.go:UpdateProfile` |
| 登录态管理 | 前端 Pinia Store 管理 token 持久化（`uni.setStorageSync`）和登录状态 | `store/user.ts` |
| 自动登录检测 | 应用启动时自动加载本地 token，跳过登录页 | `App.vue:loadToken` |
| 退出登录 | 清除本地 token 和用户信息，跳转登录页 | `mine.vue:onLogout` |

### 二、打卡系统

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 视频录制/选择 | 调用系统相机录制视频（最长60秒）或从相册选择，自动压缩 | `checkin.vue:chooseVideo` |
| 打卡准备 | 创建打卡记录（状态：待转码），返回 checkin_id | `handler.go:Prepare`, `service.go:Prepare` |
| 预签名上传 | 客户端获取 MinIO 预签名 PUT URL，视频直传 MinIO 不经过服务端 | `handler.go:Presign`, `checkin.vue:onSubmit` |
| 上传进度 | 实时显示上传百分比进度条 | `checkin.vue:uploadFile:onProgressUpdate` |
| 上传回调 | 上传完成后通知 Media Server，加入 Redis 转码队列 | `handler.go:UploadCallback` |
| 异步转码 | Worker 消费队列，FFmpeg 转码为 H.264 MP4（1280p, CRF28），提取封面帧 | `worker/worker.go:process` |
| 转码完成回调 | 转码完成后回调 API Server 更新记录状态、视频URL、封面URL、时长 | `worker.go:notifySuccess`, `handler.go:TranscodeCallback` |
| 打卡描述 | 提交打卡时可填写训练描述（最多200字） | `checkin.vue`, `model/checkin.go:Description` |
| 打卡积分 | 每次打卡默认获得10积分 | `service.go:Prepare:Score` |
| 打卡状态流转 | pending(待转码) → processing(转码中) → done(已完成) / failed(失败) | `model/checkin.go:CheckinStatus` |

### 三、广场系统

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 瀑布流展示 | 双列瀑布流布局，奇偶分列，展示封面图+描述+作者+点赞数 | `square.vue:waterfall` |
| 分页加载 | 支持分页查询，默认每页10条 | `handler.go:GetList`, `service.go:GetList` |
| 用户信息联查 | 列表接口自动填充打卡者的昵称和头像 | `service.go:GetList:userIDs` |
| 点赞状态 | 列表接口批量查询当前用户对每条记录的点赞状态 | `handler.go:GetList:BatchIsLiked` |
| 搜索栏 | 顶部搜索框 UI（搜索逻辑待接入） | `square.vue:search-bar` |
| Tab切换 | 广场/考核组 Tab 切换 UI（切换逻辑待接入） | `square.vue:tabs` |

### 四、社交系统

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 点赞/取消点赞 | Toggle 模式，重复点击取消点赞，同步更新 like_count | `handler.go:ToggleLike`, `service.go:ToggleLike` |
| 发表评论 | 对打卡记录发表文字评论 | `handler.go:AddComment`, `service.go:AddComment` |
| 评论列表 | 分页展示评论列表，包含评论者昵称和头像 | `handler.go:GetComments`, `service.go:GetComments` |
| 评论计数 | 发表评论自动递增 comment_count | `checkin_repo.go:IncrCommentCount` |
| 唯一点赞索引 | MongoDB 对 (checkin_id, user_id) 建立唯一索引，防止重复点赞 | `mongo-init.js` |

### 五、排行榜系统

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 排行榜查询 | 支持按时间段筛选：今日(day)/本周(week)/总榜(all) | `rank.vue`, `handler.go:GetRankList` |
| 我的排名 | 查询当前用户在指定时间段的排名和积分 | `handler.go:GetMyRank` |
| 排行榜缓存 | 使用 rank_cache 集合存储计算好的排名数据 | `repository/rank_repo.go` |
| 排名刷新 | 支持批量刷新排行榜（删除旧数据+插入新数据） | `rank_repo.go:RefreshRank` |

### 六、考核组系统

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 考核组列表 | 展示所有考核组，包含名称、描述、人数、成员头像 | `group.vue`, `handler.go:List` |
| 考核组详情 | 查看考核组详细信息和全部成员列表 | `handler.go:Detail` |
| 成员信息联查 | 自动查询组内成员的昵称和头像 | `service.go:List:FindByIDs` |

### 七、视频播放

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| 视频详情页 | 展示视频播放器 + 作者信息 + 描述 + 点赞数 + 评论数 + 评论列表 | `video-detail.vue` |
| 视频播放URL | 通过 Media Server 获取 MinIO 预签名 GET URL（有效期2小时） | `handler.go:GetURL` |
| 我的视频时间线 | 按时间线展示当前用户所有打卡视频，支持点击跳转详情 | `my-video.vue` |
| Nginx视频流代理 | 支持 Range 请求，实现视频拖拽播放 | `nginx.conf:/video/` |
| 封面图缓存 | Nginx 对封面图进行7天缓存 | `nginx.conf:/cover/` |

### 八、视频转码（Media Server）

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| MinIO桶自动创建 | 启动时自动创建 raw/video/cover 三个桶 | `minio.go:EnsureBuckets` |
| 视频信息探测 | FFmpeg 探测视频时长、分辨率等信息 | `ffmpeg.go:Probe` |
| H.264转码 | 将任意格式视频转码为 H.264 MP4，限制1280p，CRF28压缩 | `ffmpeg.go:Transcode` |
| 封面提取 | 从转码后视频第1秒提取帧作为封面图（JPEG） | `ffmpeg.go:Transcode:coverPath` |
| faststart优化 | 使用 `-movflags +faststart` 使MP4支持边下边播 | `ffmpeg.go:Transcode` |
| 多Worker并发 | 支持配置多个转码 Worker 并发消费队列 | `config.go:Workers`, `worker.go:Start` |
| Redis队列 | 使用 Redis List 作为转码任务队列，BRPop 阻塞消费 | `worker.go:processLoop` |
| 健康检查 | `/health` 端点用于服务健康监测 | `handler.go:Health` |

### 九、基础设施与部署

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| Docker Compose编排 | 一键部署 MongoDB + Redis + MinIO + API Server + Media Server + Nginx | `docker-compose.yml` |
| MongoDB索引优化 | openid唯一索引、checkin(user_id+created_at)复合索引、likes唯一索引 | `mongo-init.js` |
| Nginx统一入口 | 所有服务通过 Nginx 80端口统一代理 | `nginx.conf` |
| CORS跨域 | 后端全局 CORS 中间件，支持所有来源 | `middleware/cors.go` |
| 请求日志 | 结构化日志记录每次请求的 method/path/status/latency | `middleware/logger.go` |
| 请求ID | 每个请求生成唯一 X-Request-ID，便于链路追踪 | `middleware/cors.go:RequestID` |
| Panic恢复 | Gin Recovery 中间件防止服务崩溃 | `router.go:gin.Recovery()` |

### 十、前端工程化

| 功能 | 说明 | 涉及文件 |
|------|------|----------|
| TypeScript全量类型 | 所有页面和工具函数使用 TypeScript 编写 | 全部 `.ts`/`.vue` 文件 |
| Pinia状态管理 | 用户状态和打卡列表状态分离管理 | `store/user.ts`, `store/checkin.ts` |
| 请求统一封装 | 统一的 request 函数，自动附加 Token、处理 401 跳转、错误提示 | `utils/request.ts` |
| 双服务请求分离 | API 请求和 Media 请求分别指向不同服务地址 | `utils/request.ts:BASE_URL/MEDIA_URL` |
| TabBar五页签 | 首页/广场/打卡/排位/我的 五个底部导航 | `pages.json:tabBar` |
| 响应式布局 | 使用 rpx 单位适配不同屏幕尺寸 | 全部 `.vue` 文件 |

---

## API 接口清单

### 公开接口

| 接口 | 方法 | 说明 | 请求体/参数 |
|------|------|------|------------|
| `/api/auth/login` | POST | 微信登录 | `{ code, nickname, avatar, gender }` |

### 需鉴权接口（Bearer Token）

| 接口 | 方法 | 说明 | 请求体/参数 |
|------|------|------|------------|
| `/api/user/profile` | GET | 获取个人信息 | - |
| `/api/user/profile` | PUT | 更新个人信息 | `{ nickname?, avatar? }` |
| `/api/checkin/prepare` | POST | 准备打卡（创建记录） | `{ description? }` |
| `/api/checkin/list` | GET | 广场列表 | `?page=1&page_size=10` |
| `/api/checkin/mine` | GET | 我的打卡记录 | `?page=1&page_size=10` |
| `/api/checkin/:id` | DELETE | 删除打卡记录 | - |
| `/api/checkin/:id/like` | POST | 点赞/取消点赞 | - |
| `/api/checkin/:id/comment` | POST | 发表评论 | `{ content }` |
| `/api/checkin/:id/comments` | GET | 评论列表 | `?page=1&page_size=20` |
| `/api/rank` | GET | 排行榜 | `?period=all&page=1&page_size=20` |
| `/api/rank/me` | GET | 我的排名 | `?period=all` |
| `/api/group/list` | GET | 考核组列表 | - |
| `/api/group/:id` | GET | 考核组详情 | - |

### Media Server 接口

| 接口 | 方法 | 说明 | 请求体/参数 |
|------|------|------|------------|
| `/health` | GET | 健康检查 | - |
| `/media/upload/presign` | GET | 获取预签名上传URL | `?checkin_id=xxx&ext=mp4` |
| `/media/upload/callback` | POST | 上传完成回调（加入转码队列） | `{ checkin_id, object_name, bucket, file_size? }` |
| `/media/url` | GET | 获取视频播放URL | `?object=xxx&bucket=video` |

### 内部接口（服务间调用）

| 接口 | 方法 | 说明 | 请求体/参数 |
|------|------|------|------------|
| `/api/internal/transcode/done` | POST | 转码完成回调 | `{ checkin_id, video_url, cover_url, duration }` |

---

## 数据库设计

### users 集合

| 字段 | 类型 | 说明 | 索引 |
|------|------|------|------|
| _id | ObjectId | 主键 | - |
| openid | String | 微信openid | 唯一索引 |
| unionid | String | 微信unionid | - |
| nickname | String | 昵称 | - |
| avatar | String | 头像URL | - |
| gender | Int | 性别 | - |
| province | String | 省份 | - |
| city | String | 城市 | - |
| group_id | ObjectId | 所属考核组 | 普通索引 |
| score | Int | 总积分 | 普通索引(降序) |
| check_days | Int | 累计打卡天数 | - |
| created_at | DateTime | 创建时间 | - |
| updated_at | DateTime | 更新时间 | - |

### checkins 集合

| 字段 | 类型 | 说明 | 索引 |
|------|------|------|------|
| _id | ObjectId | 主键 | - |
| user_id | ObjectId | 打卡用户 | 复合索引(user_id+created_at) |
| video_url | String | 转码后视频路径 | - |
| cover_url | String | 封面图路径 | - |
| raw_url | String | 原始视频路径 | - |
| description | String | 打卡描述 | - |
| duration | Float | 视频时长(秒) | - |
| file_size | Int | 文件大小 | - |
| score | Int | 本次积分 | - |
| status | Int | 状态(0待转码/1转码中/2完成/3失败) | 普通索引 |
| like_count | Int | 点赞数 | - |
| comment_count | Int | 评论数 | - |
| created_at | DateTime | 创建时间 | 普通索引(降序) |
| updated_at | DateTime | 更新时间 | - |

### comments 集合

| 字段 | 类型 | 说明 | 索引 |
|------|------|------|------|
| _id | ObjectId | 主键 | - |
| checkin_id | ObjectId | 所属打卡记录 | 复合索引(checkin_id+created_at) |
| user_id | ObjectId | 评论用户 | 普通索引 |
| content | String | 评论内容 | - |
| created_at | DateTime | 创建时间 | - |

### likes 集合

| 字段 | 类型 | 说明 | 索引 |
|------|------|------|------|
| _id | ObjectId | 主键 | - |
| checkin_id | ObjectId | 所属打卡记录 | 复合唯一索引(checkin_id+user_id) |
| user_id | ObjectId | 点赞用户 | 普通索引 |
| created_at | DateTime | 创建时间 | - |

### groups 集合

| 字段 | 类型 | 说明 | 索引 |
|------|------|------|------|
| _id | ObjectId | 主键 | - |
| name | String | 组名 | - |
| description | String | 描述 | - |
| leader_id | ObjectId | 组长ID | 普通索引 |
| member_ids | Array[ObjectId] | 成员ID列表 | - |
| created_at | DateTime | 创建时间 | - |

### rank_cache 集合

| 字段 | 类型 | 说明 | 索引 |
|------|------|------|------|
| _id | ObjectId | 主键 | - |
| user_id | ObjectId | 用户ID | 复合索引(user_id+period) |
| score | Int | 积分 | - |
| rank | Int | 排名 | 复合索引(period+rank) |
| period | String | 时间段(day/week/all) | - |
| updated_at | DateTime | 更新时间 | - |

---

## 视频处理流程

```
┌─────────┐     ┌───────────┐     ┌─────────┐     ┌───────────┐     ┌─────────┐
│  客户端  │────→│ API Server│────→│  Redis  │────→│  Worker   │────→│  MinIO  │
│  录制    │     │ prepare   │     │  队列    │     │ FFmpeg    │     │ video桶 │
└─────────┘     └───────────┘     └─────────┘     └───────────┘     └─────────┘
     │                                                                    │
     │              ┌───────────┐                                         │
     └──presign──→  │ Media Srv │                                         │
                    │ presign   │                                         │
                    └───────────┘                                         │
     │                                                                    │
     └────直传MinIO(raw桶)────→ ┌───────────┐                             │
                               │ Media Srv │←──callback─── Worker ────────┘
                               │ callback  │
                               └───────────┘
                                      │
                                      ▼
                               ┌───────────┐
                               │ API Server│ ← 更新status/video_url/cover_url/duration
                               │ callback  │
                               └───────────┘
```

---

## 快速启动

### 方式一：本地开发

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

### 方式二：Docker 一键部署

```bash
cd deploy
docker-compose up -d
```

部署后服务访问：

| 服务 | 地址 | 说明 |
|------|------|------|
| Nginx | http://localhost:80 | 统一入口 |
| API Server | http://localhost:8080 | 业务接口 |
| Media Server | http://localhost:8081 | 视频服务 |
| MinIO Console | http://localhost:9001 | 存储管理 (minioadmin/minioadmin) |
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
  expires: 72             # Token有效期(小时)

wx:
  app_id: "your-wx-app-id"      # 微信小程序AppID
  secret: "your-wx-app-secret"  # 微信小程序Secret

media_url: "http://media-server:8081"  # Media Server地址
```

### media-server/configs/config.yaml

```yaml
server:
  port: "8081"

minio:
  endpoint: "minio:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  use_ssl: false
  raw_bucket: "raw"        # 原始视频桶
  video_bucket: "video"    # 转码后视频桶
  cover_bucket: "cover"    # 封面图桶

redis:
  addr: "redis:6379"
  password: ""
  db: 0

ffmpeg:
  binary: "ffmpeg"         # FFmpeg可执行文件路径
  workers: 2               # 转码Worker并发数

api_server: "http://api-server:8080"  # API Server地址(回调用)
```

---

## 后续规划功能

以下功能已纳入迭代计划，按优先级排列。

### 一、AI 视频评分系统

**目标**：接入视频理解大模型，通过 AI 自动分析武术动作质量并给出专业评分和改进建议，替代人工评分，提升评价客观性和即时性。

**核心能力**：

| 能力 | 说明 |
|------|------|
| 动作规范性评估 | 识别套路动作的标准度、连贯性、完整性 |
| 力度与劲力分析 | 评估发力方式、劲力传导、爆发力表现 |
| 身法与步法评价 | 分析重心转换、身法协调、步法稳定性 |
| 节奏与速度评分 | 评估动作节奏感、快慢相间、停顿控制 |
| 综合评分 | 加权计算总分（0-100），生成等级（S/A/B/C/D） |
| 改进建议 | 针对薄弱环节生成文字建议（如"马步重心偏高，建议加强下盘训练"） |

**技术方案**：

```
打卡视频上传
     │
     ▼
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│ Media Server │────→│  AI Server    │────→│  大模型推理   │
│ 转码完成      │     │  预处理/抽帧   │     │  视频理解模型  │
└──────────────┘     └───────────────┘     └──────────────┘
                                                │
                                                ▼
                                         ┌──────────────┐
                                         │  结构化输出   │
                                         │  评分+建议    │
                                         └──────────────┘
                                                │
                                                ▼
                                         ┌──────────────┐
                                         │  API Server  │
                                         │  写入数据库   │
                                         └──────────────┘
```

**模型选型方向**：

| 方案 | 说明 | 优劣 |
|------|------|------|
| 方案A：通用视频理解模型 + Prompt | 使用 Video-LLaVA / Qwen-VL 等开源多模态模型，通过 Prompt Engineering 引导评分 | 上手快，精度一般 |
| 方案B：微调视频模型 | 基于方案A，使用武术视频数据集进行 LoRA 微调 | 精度高，需要标注数据 |
| 方案C：关键点检测 + 评分模型 | 使用姿态估计（如 MMPose）提取骨骼关键点，训练评分回归模型 | 精度最高，开发量大 |

**推荐路径**：方案A快速验证 → 积累数据 → 方案B微调提升 → 方案C作为长期目标。

**新增数据模型**：

```
ai_scores 集合 {
  _id            ObjectId
  checkin_id     ObjectId     // 关联打卡记录
  user_id        ObjectId     // 用户
  total_score    Float        // 综合评分 (0-100)
  grade          String       // 等级 (S/A/B/C/D)
  detail_scores  {            // 分项评分
    technique    Float        // 动作规范 (0-20)
    power        Float        // 力度劲力 (0-20)
    movement     Float        // 身法步法 (0-20)
    rhythm       Float        // 节奏速度 (0-20)
    completeness Float        // 完整性   (0-20)
  }
  suggestions    [String]     // AI改进建议列表
  model_version  String       // 使用的模型版本
  raw_output     Object       // 模型原始输出(调试用)
  created_at     DateTime
}
```

**新增 API**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/ai/score/:checkin_id` | GET | 获取某条打卡的AI评分 |
| `/api/ai/history` | GET | 我的AI评分历史(分页) |
| `/api/ai/compare` | POST | 对比两次打卡的评分差异 |

---

### 二、训练计划模块

**目标**：为武术训练提供结构化的计划管理能力，支持教练为考核组制定集体训练计划，也支持个人自定义训练安排，帮助学员系统性提升。

**核心能力**：

| 能力 | 说明 |
|------|------|
| 计划模板 | 预设常见训练模板（如"初级长拳28天计划"、"太极入门4周计划"） |
| 自定义计划 | 用户可自由创建训练计划，设定周期、每日内容、目标 |
| 群主派发 | 群员可向组员派发统一训练计划 |
| 每日任务 | 按日期展示当天需要完成的训练项目，打卡关联 |
| 完成追踪 | 记录每日任务完成情况，统计完成率 |
| 周期报告 | 计划结束后生成训练报告（完成率/打卡频率/积分变化） |
| 提醒通知 | 通过微信订阅消息提醒当日未完成的训练任务 |

**数据模型**：

```
training_plans 集合 {
  _id            ObjectId
  user_id        ObjectId        // 创建者
  group_id       ObjectId        // 所属考核组(可选，0=个人计划)
  title          String          // 计划名称
  description    String          // 计划描述
  start_date     Date            // 开始日期
  end_date       Date            // 结束日期
  status         Int             // 状态(0草稿/1进行中/2已完成/3已终止)
  days           [{              // 每日任务列表
    day          Int             // 第几天
    date         Date            // 对应日期
    tasks        [{              // 训练项目
      title      String          // 项目名称(如"基本功:弓步冲拳")
      type       String          // 类型(basic/taolu/sanda/qigong)
      duration   Int             // 预计时长(分钟)
      reps       String          // 组数/次数(如"3组x10次")
      note       String          // 备注
      checkin_id ObjectId        // 关联的打卡记录(完成时填入)
      status     Int             // 状态(0未完成/1已完成/2跳过)
    }]
  }]
  stats          {               // 统计信息(定时更新)
    total_tasks  Int             // 总任务数
    completed    Int             // 已完成数
    completion_rate Float        // 完成率
  }
  created_at     DateTime
  updated_at     DateTime
}

training_templates 集合 {
  _id            ObjectId
  name           String          // 模板名称
  category       String          // 分类(初级/中级/高级)
  style          String          // 拳种(长拳/太极/南拳/散打...)
  duration_days  Int             // 周期天数
  description    String          // 描述
  days           [{              // 每日安排(同上)
    day          Int
    tasks        [{ title, type, duration, reps, note }]
  }]
  author         String          // 作者
  usage_count    Int             // 使用次数
  created_at     DateTime
}
```

**新增 API**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/training/plan` | POST | 创建训练计划 |
| `/api/training/plan/:id` | GET | 获取计划详情 |
| `/api/training/plans` | GET | 我的计划列表 |
| `/api/training/plan/:id` | PUT | 更新计划 |
| `/api/training/plan/:id` | DELETE | 删除计划 |
| `/api/training/today` | GET | 今日任务 |
| `/api/training/task/:plan_id/:day/:task_idx` | PUT | 更新任务状态 |
| `/api/training/plan/:id/report` | GET | 计划完成报告 |
| `/api/training/template/list` | GET | 训练模板列表 |
| `/api/training/template/:id/apply` | POST | 应用模板创建计划 |

---

### 三、个人资料库模块

**目标**：为武术学习者提供个人知识管理工具，支持上传和管理武术教学资料（视频、图片、文档），构建个人专属的武术资料库，方便随时查阅和复习。

**核心能力**：

| 能力 | 说明 |
|------|------|
| 资料上传 | 支持上传视频（教学录像/比赛录像）、图片（动作分解图）、文档（拳谱/理论） |
| 分类管理 | 按拳种/类型/难度等维度分类，支持自定义标签 |
| 资料浏览 | 瀑布流/列表视图切换，支持按标签/类型/时间筛选 |
| 资料搜索 | 全文搜索标题、描述、标签 |
| 收藏夹 | 收藏常用资料，快速访问 |
| 分享 | 将资料分享到考核组，组内成员可查看 |
| 存储配额 | 限制每用户的存储空间（如5GB），防止滥用 |

**数据模型**：

```
resources 集合 {
  _id            ObjectId
  user_id        ObjectId        // 上传者
  title          String          // 资料标题
  description    String          // 资料描述
  type           String          // 类型(video/image/document)
  category       String          // 分类(基本功/套路/散打/太极/理论...)
  tags           [String]        // 自定义标签
  difficulty     String          // 难度(初级/中级/高级)
  file_url       String          // 文件URL(MinIO)
  file_size      Int             // 文件大小(字节)
  cover_url      String          // 封面图URL
  duration       Float           // 视频时长(秒，仅视频)
  share_scope    String          // 分享范围(private/group/public)
  group_id       ObjectId        // 分享到的考核组(share_scope=group时)
  is_favorite    Boolean         // 是否收藏
  view_count     Int             // 查看次数
  download_count Int             // 下载次数
  created_at     DateTime
  updated_at     DateTime
}

resource_tags 集合 {              // 标签聚合(用于搜索提示)
  _id            ObjectId
  user_id        ObjectId
  tag            String
  count          Int             // 使用次数
}
```

**新增 API**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/resource/upload/presign` | GET | 获取资料上传预签名URL |
| `/api/resource/upload/callback` | POST | 上传完成回调 |
| `/api/resource` | POST | 创建资料记录(含标题/分类/标签) |
| `/api/resource/list` | GET | 资料列表(分页+筛选) |
| `/api/resource/:id` | GET | 资料详情 |
| `/api/resource/:id` | PUT | 更新资料信息 |
| `/api/resource/:id` | DELETE | 删除资料 |
| `/api/resource/:id/favorite` | POST | 收藏/取消收藏 |
| `/api/resource/search` | GET | 搜索资料 |
| `/api/resource/tags` | GET | 我的标签列表 |
| `/api/resource/stats` | GET | 存储用量统计 |

---

### 四、感悟笔记模块

**目标**：为武术练习者提供训练感悟和心得的记录工具，支持按时间线和标签分类追踪个人成长轨迹，将零散的训练反思结构化沉淀，形成可回顾的习武日志。

**核心能力**：

| 能力 | 说明 |
|------|------|
| 感悟记录 | 记录训练心得、体悟、疑问，支持富文本（文字+图片） |
| 关联打卡 | 可关联某次打卡视频，在感悟中引用视频片段 |
| 关联训练计划 | 可关联训练计划的某一天，记录当天训练感受 |
| 时间线展示 | 按时间倒序展示所有感悟，形成习武日志 |
| 标签分类 | 支持打标签（如"太极拳感悟"、"发力体会"、"比赛总结"） |
| 心情标记 | 标记训练心情（突破/满意/一般/困惑/低落），可视化情绪曲线 |
| 回顾提醒 | 定期推送历史感悟（如"一年前的今天你写道..."） |
| 隐私控制 | 感悟默认私密，可选择公开到广场 |

**数据模型**：

```
insights 集合 {
  _id            ObjectId
  user_id        ObjectId        // 作者
  content        String          // 感悟内容(支持Markdown)
  images         [String]        // 配图URL列表
  mood           String          // 心情(breakthrough/good/normal/confused/low)
  tags           [String]        // 标签
  checkin_id     ObjectId        // 关联打卡记录(可选)
  plan_id        ObjectId        // 关联训练计划(可选)
  plan_day       Int             // 计划第几天(与plan_id配合)
  visibility     String          // 可见性(private/public)
  like_count     Int             // 点赞数(公开时)
  created_at     DateTime
  updated_at     DateTime
}

insight_tags 集合 {               // 标签聚合
  _id            ObjectId
  user_id        ObjectId
  tag            String
  count          Int
}
```

**新增 API**：

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/insight` | POST | 创建感悟 |
| `/api/insight/list` | GET | 我的感悟时间线(分页) |
| `/api/insight/:id` | GET | 感悟详情 |
| `/api/insight/:id` | PUT | 编辑感悟 |
| `/api/insight/:id` | DELETE | 删除感悟 |
| `/api/insight/tags` | GET | 我的标签列表 |
| `/api/insight/mood-stats` | GET | 心情统计(近30天) |
| `/api/insight/on-this-day` | GET | 历史今日感悟 |
| `/api/insight/public` | GET | 公开感悟广场(分页) |

---

### 迭代优先级与排期

| 阶段 | 功能 | 预计周期 | 依赖 |
|------|------|---------|------|
| **P0** (当前) | 核心打卡 + 广场 + 排行榜 + 社交 | 已完成 | - |
| **P1** | 训练计划模块 | 2-3周 | 核心打卡 |
| **P1** | 感悟笔记模块 | 2周 | 核心打卡 |
| **P2** | 个人资料库模块 | 2-3周 | MinIO存储 |
| **P2** | AI视频评分（方案A：Prompt方式） | 2-3周 | 视频转码 |
| **P3** | AI视频评分（方案B：LoRA微调） | 4-6周 | 积累评分数据 |
| **P3** | 训练提醒订阅消息 | 1周 | 训练计划 |

### 后端服务扩展规划

随着功能增加，建议逐步拆分为以下微服务：

```
当前 (双服务)                    规划 (多服务)
┌──────────────┐              ┌──────────────┐
│  API Server  │              │  API Server  │ ← 用户/打卡/社交/排行榜
│  (所有业务)   │      →       ├──────────────┤
├──────────────┤              │ Media Server │ ← 视频上传/转码/播放
│ Media Server │              ├──────────────┤
│  (视频处理)   │              │  AI Server   │ ← 视频评分/模型推理
└──────────────┘              ├──────────────┤
                              │Training Svc  │ ← 训练计划/模板
                              ├──────────────┤
                              │Resource Svc  │ ← 资料库管理
                              └──────────────┘
```
