package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -destination=mock_rank_repo.go -package=repository wuxie-api/internal/repository RankRepoInterface
//go:generate mockgen -destination=mock_group_repo.go -package=repository wuxie-api/internal/repository GroupRepoInterface

// RankRepoInterface 排行榜仓库接口
type RankRepoInterface interface {
	GetRankList(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error)
	GetUserRank(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error)
	RefreshRank(ctx context.Context, period model.RankPeriod, entries []*model.RankEntry) error
}

// GroupRepoInterface 团组仓库接口
type GroupRepoInterface interface {
	FindAll(ctx context.Context) ([]*model.Group, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Group, error)
	FindByInviteCode(ctx context.Context, code string) (*model.Group, error)
	AddMember(ctx context.Context, groupID, userID primitive.ObjectID) error
	UpdateInviteCode(ctx context.Context, groupID primitive.ObjectID, code string) error
}

var _ RankRepoInterface = (*RankRepo)(nil)
var _ GroupRepoInterface = (*GroupRepo)(nil)
