package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RankService struct {
	rankRepo *repository.RankRepo
}

func NewRankService(rankRepo *repository.RankRepo) *RankService {
	return &RankService{rankRepo: rankRepo}
}

func (s *RankService) GetRankList(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error) {
	return s.rankRepo.GetRankList(ctx, period, page, pageSize)
}

func (s *RankService) GetUserRank(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
	return s.rankRepo.GetUserRank(ctx, userID, period)
}