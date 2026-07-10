package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type RankHistoryService struct {
	historyRepo repository.RankHistoryRepoInterface
	rankRepo    repository.RankRepoInterface
	logger      *zap.Logger
}

func NewRankHistoryService(
	historyRepo repository.RankHistoryRepoInterface,
	rankRepo repository.RankRepoInterface,
	logger *zap.Logger,
) *RankHistoryService {
	return &RankHistoryService{
		historyRepo: historyRepo,
		rankRepo:    rankRepo,
		logger:      logger,
	}
}

// GetUserRankTrend 获取用户排名趋势
func (s *RankHistoryService) GetUserRankTrend(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod, days int) ([]model.RankTrendPoint, error) {
	if days <= 0 || days > 90 {
		days = 30
	}
	return s.historyRepo.GetUserHistory(ctx, userID, period, days)
}

// SaveRankSnapshot 保存排名快照（定时任务调用）
func (s *RankHistoryService) SaveRankSnapshot(ctx context.Context, period model.RankPeriod) {
	entries, err := s.rankRepo.GetRankList(ctx, period, 1, 100)
	if err != nil {
		s.logger.Error("save rank snapshot: get rank list failed",
			zap.String("period", string(period)),
			zap.Error(err),
		)
		return
	}

	if err := s.historyRepo.SaveSnapshot(ctx, entries, period); err != nil {
		s.logger.Error("save rank snapshot: save failed",
			zap.String("period", string(period)),
			zap.Int("count", len(entries)),
			zap.Error(err),
		)
		return
	}

	s.logger.Info("rank snapshot saved",
		zap.String("period", string(period)),
		zap.Int("count", len(entries)),
	)
}
