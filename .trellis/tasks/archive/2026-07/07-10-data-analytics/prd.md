# 数据分析功能

## Goal

为用户提供个人数据分析看板，增强用户粘性和训练动力。

## Background

### 已有分析能力
- `/insight/mood-stats` — 心情统计（按心情分组计数）
- `/resource/stats` — 资源存储统计（总量、分类）
- `/rank` — 排行榜（日/周/总）
- Admin Dashboard — 全局统计（用户数、打卡数）
- Cron: 每 10 分钟刷新排行榜

### 缺失的分析维度
- 打卡热力图（哪些天打卡了）
- 连续打卡天数趋势
- 个人训练完成率
- 每周/月打卡频率统计

## Requirements

### R1: 打卡热力图 API
- `GET /analytics/checkin-heatmap?months=6`
- 返回每天的打卡数量 `{ "2026-01-15": 2, "2026-01-16": 1, ... }`
- 用 MongoDB Aggregate 按日期分组

### R2: 打卡趋势 API
- `GET /analytics/checkin-trend?days=30`
- 返回每天的打卡数和积分 `{ "date": "2026-01-15", "count": 2, "score": 20 }`
- 用于折线图展示

### R3: 个人数据概览 API
- `GET /analytics/overview`
- 返回：总打卡天数、连续打卡天数、本周打卡数、本月打卡数、总积分、当前排名

### R4: 客户端数据看板页面
- 新增 `pages/analytics/overview.vue`
- 展示热力图、趋势图、个人概览

## Acceptance Criteria

- [ ] 3 个分析 API 端点可调用
- [ ] 热力图数据按日期正确聚合
- [ ] 连续打卡天数计算正确
- [ ] 客户端页面可展示数据
- [ ] `go build ./...` 编译通过

## Out of Scope

- 实时数据流（WebSocket）
- 复杂图表库（ECharts 等）
- 管理端数据分析（已有 Dashboard）