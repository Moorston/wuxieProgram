# 社交功能

## Goal

增强用户社交互动：评论回复、关注系统、动态流、用户主页。

## Existing Social Features
- 点赞/取消点赞（ToggleLike）
- 添加评论（AddComment）
- 获取评论列表（GetComments）

## Requirements

### R1: 评论回复
- Comment 模型增加 `ParentID` 字段（回复的评论 ID）
- 添加评论时可选传入 `parent_id`
- 评论列表返回时附带回复列表

### R2: 关注/粉丝系统
- 新增 Follow 模型（follower_id, following_id）
- 关注/取消关注 API
- 获取关注列表/粉丝列表

### R3: 动态 Feed
- 查询关注用户的公开打卡
- 按时间倒序分页

### R4: 用户主页
- 查询指定用户的公开打卡和感悟
- 展示用户基本信息 + 统计

## Acceptance Criteria

- [ ] 评论支持回复
- [ ] 关注/取消关注 API 可用
- [ ] 动态 Feed 可查询
- [ ] 用户主页可访问
- [ ] `go build ./...` 编译通过

## Out of Scope

- 私信功能
- 评论通知推送
- 关注推荐算法