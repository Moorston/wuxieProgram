package service

import (
	"context"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AnalyticsService struct {
	checkinRepo repository.CheckinRepoInterface
	userRepo    repository.UserRepoInterface
	logger      *zap.Logger
}

func NewAnalyticsService(checkinRepo repository.CheckinRepoInterface, userRepo repository.UserRepoInterface, logger *zap.Logger) *AnalyticsService {
	return &AnalyticsService{checkinRepo: checkinRepo, userRepo: userRepo, logger: logger}
}

// CheckinHeatmap 打卡热力图数据
type CheckinHeatmap map[string]int

func (s *AnalyticsService) GetCheckinHeatmap(ctx context.Context, userID primitive.ObjectID, months int) (CheckinHeatmap, error) {
	if months <= 0 || months > 12 {
		months = 6
	}

	startDate := time.Now().AddDate(0, -months, 0)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"status":     model.CheckinStatusDone,
			"created_at": bson.M{"$gte": startDate},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"},
			},
			"count": bson.M{"$sum": 1},
		}},
	}

	cursor, err := s.checkinRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	heatmap := make(CheckinHeatmap)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		heatmap[result.ID] = result.Count
	}

	if err := cursor.Err(); err != nil {
		s.logger.Warn("heatmap cursor error", zap.Error(err))
	}

	return heatmap, nil
}

// CheckinTrend 打卡趋势数据
type TrendPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
	Score int    `json:"score"`
}

func (s *AnalyticsService) GetCheckinTrend(ctx context.Context, userID primitive.ObjectID, days int) ([]TrendPoint, error) {
	if days <= 0 || days > 365 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"status":     model.CheckinStatusDone,
			"created_at": bson.M{"$gte": startDate},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"},
			},
			"count": bson.M{"$sum": 1},
			"score": bson.M{"$sum": "$score"},
		}},
		{"$sort": bson.M{"_id": 1}},
	}

	cursor, err := s.checkinRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var trend []TrendPoint
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
			Score int    `bson:"score"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		trend = append(trend, TrendPoint{
			Date:  result.ID,
			Count: result.Count,
			Score: result.Score,
		})
	}

	if err := cursor.Err(); err != nil {
		s.logger.Warn("trend cursor error", zap.Error(err))
	}

	return trend, nil
}

// Overview 个人数据概览
type Overview struct {
	TotalCheckDays  int `json:"total_check_days"`
	StreakDays      int `json:"streak_days"`
	WeekCheckins    int `json:"week_checkins"`
	MonthCheckins   int `json:"month_checkins"`
	TotalScore      int `json:"total_score"`
}

func (s *AnalyticsService) GetOverview(ctx context.Context, userID primitive.ObjectID) (*Overview, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	// 本周打卡数
	weekPipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"status":     model.CheckinStatusDone,
			"created_at": bson.M{"$gte": weekStart},
		}},
		{"$count": "count"},
	}
	weekCount := s.aggregateCount(ctx, weekPipeline)

	// 本月打卡数
	monthPipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"status":     model.CheckinStatusDone,
			"created_at": bson.M{"$gte": monthStart},
		}},
		{"$count": "count"},
	}
	monthCount := s.aggregateCount(ctx, monthPipeline)

	// 连续打卡天数
	streakDays := s.calculateStreak(ctx, userID)

	return &Overview{
		TotalCheckDays: user.CheckDays,
		StreakDays:     streakDays,
		WeekCheckins:   weekCount,
		MonthCheckins:  monthCount,
		TotalScore:     user.Score,
	}, nil
}

func (s *AnalyticsService) aggregateCount(ctx context.Context, pipeline []bson.M) int {
	cursor, err := s.checkinRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			Count int `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			return result.Count
		}
	}
	return 0
}

func (s *AnalyticsService) calculateStreak(ctx context.Context, userID primitive.ObjectID) int {
	// 查询最近 60 天的打卡日期（足够计算连续天数）
	startDate := time.Now().AddDate(0, 0, -60)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	pipeline := []bson.M{
		{"$match": bson.M{
			"user_id":    userID,
			"status":     model.CheckinStatusDone,
			"created_at": bson.M{"$gte": startDate},
		}},
		{"$group": bson.M{
			"_id": bson.M{
				"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$created_at"},
			},
		}},
		{"$sort": bson.M{"_id": -1}},
	}

	cursor, err := s.checkinRepo.Aggregate(ctx, pipeline)
	if err != nil {
		return 0
	}
	defer cursor.Close(ctx)

	var dates []string
	for cursor.Next(ctx) {
		var result struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&result); err == nil {
			dates = append(dates, result.ID)
		}
	}

	if len(dates) == 0 {
		return 0
	}

	// 从最近的打卡日开始倒推计算连续天数
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	// 确定起始基准日
	var baseDate time.Time
	if dates[0] == today {
		baseDate = time.Now()
	} else if dates[0] == yesterday {
		baseDate = time.Now().AddDate(0, 0, -1)
	} else {
		return 0 // 最近打卡不是今天也不是昨天，连续中断
	}

	streak := 0
	for i := 0; i < len(dates); i++ {
		expected := baseDate.AddDate(0, 0, -i).Format("2006-01-02")
		if dates[i] == expected {
			streak++
		} else {
			break
		}
	}

	return streak
}
