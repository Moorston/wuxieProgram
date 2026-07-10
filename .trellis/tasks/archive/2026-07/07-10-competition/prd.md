# 赛事活动系统

## Goal

支持管理员创建赛事活动，用户提交视频参与，评委打分排名。

## Feature Overview

### 核心流程
1. 管理员创建赛事（标题、描述、时间、规则）
2. 用户在赛事期间提交打卡视频参与
3. 评委（管理员/团组组长）对参赛视频打分
4. 自动统计排名，生成赛事结果
5. 展示赛事排行榜和优秀作品

## Requirements

### R1: 数据模型
- Competition 赛事（ID、标题、描述、时间、状态、评分规则）
- CompetitionEntry 参赛作品（赛事ID、用户ID、打卡ID、分数、评委ID）

### R2: 管理员 API
- POST /admin/competitions — 创建赛事
- GET /admin/competitions — 赛事列表
- PUT /admin/competitions/:id — 更新赛事

### R3: 用户 API
- GET /competitions — 赛事列表（公开）
- GET /competitions/:id — 赛事详情
- POST /competitions/:id/submit — 提交参赛作品
- GET /competitions/:id/entries — 参赛作品列表
- GET /competitions/:id/ranking — 赛事排行榜
- POST /competitions/:id/entries/:entryId/score — 评委打分

### R4: 客户端页面
- 赛事列表页
- 赛事详情页

## Acceptance Criteria

- [ ] 管理员可创建赛事
- [ ] 用户可查看赛事列表和详情
- [ ] 用户可提交参赛作品
- [ ] 评委可打分
- [ ] 赛事排行榜自动生成
- [ ] `go build ./...` 编译通过