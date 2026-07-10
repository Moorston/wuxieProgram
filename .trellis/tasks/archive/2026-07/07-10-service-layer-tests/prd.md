# Service 层单元测试

## Goal

为剩余 9 个 Service 编写单元测试，利用已提取的 Repository 接口进行 mock 测试。

## Background

### 当前状态
- **已有测试**: `AuthService` — 9 个用例 ✅
- **待测试**: 9 个 Service，0 个测试用例 ❌

### 所有 Service 已接入 Repository 接口
- 10 个 Service 全部依赖接口而非具体类型
- 8 个接口文件已定义，覆盖所有 Repository 方法
- 可直接使用 gomock 或手写 mock 进行测试

### 优先级排序（按业务价值和复杂度）

| 优先级 | Service | 行数 | 原因 |
|:---:|---------|:---:|------|
| P0 | **CheckinService** | ~100 | 核心打卡业务，含搜索、权限判断 |
| P0 | **SocialService** | ~80 | 点赞+评论事务，含 MongoDB session |
| P0 | **InsightService** | ~120 | 感想笔记 CRUD，含标签+点赞事务 |
| P1 | **ResourceService** | ~100 | 资源库 CRUD，含标签+权限 |
| P1 | **TrainingService** | ~80 | 训练计划 CRUD |
| P1 | **NotificationService** | ~60 | 通知 CRUD |
| P2 | **UserService** | ~20 | 简单薄封装（2 方法） |
| P2 | **GroupService** | ~30 | 简单薄封装 |
| P2 | **RankService** | ~20 | 简单薄封装 |
| P3 | **CronService** | ~150 | 定时任务，依赖较多，可延迟 |

## Requirements

### R1: CheckinService 测试
- Prepare: 创建打卡成功后返回
- UpdateStatus: 转码回调更新状态
- List: 按条件分页查询
- ListByUser: 按用户查询
- Delete: 删除打卡
- Search: 关键词搜索
- GetByID: 详情查询

### R2: SocialService 测试
- ToggleLike: 点赞/取消点赞（事务）
- AddComment: 添加评论（事务）
- GetComments: 评论列表
- IsLiked: 批量检查是否已点赞

### R3: InsightService 测试
- Create: 创建感悟（含标签）
- List: 按条件分页
- Update: 更新感悟
- Delete: 删除感悟
- Like: 点赞/取消点赞（事务）
- MoodStats: 心情统计
- OnThisDay: 历史上的今天

### R4: 简单 Service 测试
- UserService: GetProfile, UpdateProfile
- RankService: GetRankList, GetUserRank
- GroupService: List, Detail

## Acceptance Criteria

- [ ] CheckinService 测试覆盖所有公开方法的主要分支
- [ ] SocialService 测试覆盖事务路径
- [ ] InsightService 测试覆盖 CRUD + 标签 + 点赞
- [ ] 全部测试 `go test -race` 通过
- [ ] 测试使用 testify/assert 编写

## Out of Scope

- CronService 测试（依赖多，需要 mock 多个 repo）
- 集成测试（需要真实 MongoDB）
- 端到端测试

## Open Questions

1. 测试文件组织方式？每个 Service 一个文件 vs 合并到一个文件？
2. Mock 方式？手写 mock struct 还是 gomock？