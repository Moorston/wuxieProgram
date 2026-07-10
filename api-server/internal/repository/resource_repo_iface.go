package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -destination=mock_resource_repo.go -package=repository wuxie-api/internal/repository ResourceRepoInterface
//go:generate mockgen -destination=mock_resource_tag_repo.go -package=repository wuxie-api/internal/repository ResourceTagRepoInterface

// ResourceRepoInterface 资源仓库接口
type ResourceRepoInterface interface {
	Create(ctx context.Context, res *model.Resource) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Resource, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, userID primitive.ObjectID, resType, category, difficulty, tag, keyword, shareScope, sortBy string, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error)
	ToggleFavorite(ctx context.Context, id primitive.ObjectID) (bool, error)
	ListFavorites(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error)
	GetUserStats(ctx context.Context, userID primitive.ObjectID) (*model.ResourceStats, error)
	IncrViewCount(ctx context.Context, id primitive.ObjectID) error
	IncrDownloadCount(ctx context.Context, id primitive.ObjectID) error
}

// ResourceTagRepoInterface 资源标签仓库接口
type ResourceTagRepoInterface interface {
	UpsertTags(ctx context.Context, userID primitive.ObjectID, tags []string) error
	DecrTags(ctx context.Context, userID primitive.ObjectID, tags []string) error
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.ResourceTag, error)
}

var _ ResourceRepoInterface = (*ResourceRepo)(nil)
var _ ResourceTagRepoInterface = (*ResourceTagRepo)(nil)
