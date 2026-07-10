package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -destination=mock_user_repo.go -package=repository wuxie-api/internal/repository UserRepoInterface

// UserRepoInterface 用户仓库接口，用于依赖注入和测试 mock
type UserRepoInterface interface {
	Create(ctx context.Context, user *model.User) error
	FindByOpenID(ctx context.Context, openid string) (*model.User, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error)
	FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]*model.User, error)
	FindTopByScore(ctx context.Context, limit int) ([]*model.User, error)
	UpsertByOpenID(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	IncrScore(ctx context.Context, id primitive.ObjectID, score int) error
	IsBanned(ctx context.Context, id primitive.ObjectID) (bool, error)
	FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*model.User, int64, error)
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status int) (int64, error)
}

// 确保 *UserRepo 实现了 UserRepoInterface
var _ UserRepoInterface = (*UserRepo)(nil)
