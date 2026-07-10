package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RankHistoryRepoInterface 排名历史仓库接口
type RankHistoryRepoInterface interface {
	SaveSnapshot(ctx context.Context, entries []*model.RankEntry, period model.RankPeriod) error
	GetUserHistory(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod, days int) ([]model.RankTrendPoint, error)
}

var _ RankHistoryRepoInterface = (*RankHistoryRepo)(nil)
