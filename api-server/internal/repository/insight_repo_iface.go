package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate mockgen -destination=mock_insight_repo.go -package=repository wuxie-api/internal/repository InsightRepoInterface
//go:generate mockgen -destination=mock_insight_tag_repo.go -package=repository wuxie-api/internal/repository InsightTagRepoInterface
//go:generate mockgen -destination=mock_insight_like_repo.go -package=repository wuxie-api/internal/repository InsightLikeRepoInterface

// InsightRepoInterface 感悟笔记仓库接口
type InsightRepoInterface interface {
	Create(ctx context.Context, insight *model.Insight) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Insight, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	Delete(ctx context.Context, id, userID primitive.ObjectID) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	ListByUser(ctx context.Context, userID primitive.ObjectID, tag, mood string, page, pageSize int) ([]*model.Insight, int64, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error)
	ListPublic(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error)
	OnThisDay(ctx context.Context, userID primitive.ObjectID, month, day int) ([]*model.Insight, error)
	MoodStats(ctx context.Context, userID primitive.ObjectID, days int) (map[string]int, error)
	IncrLikeCount(ctx context.Context, id primitive.ObjectID, delta int) error
	IncrLikeCountWithSession(sessCtx mongo.SessionContext, id primitive.ObjectID, delta int) error
	StartSession() (mongo.Session, error)
}

// InsightTagRepoInterface 感悟标签仓库接口
type InsightTagRepoInterface interface {
	UpsertTags(ctx context.Context, userID primitive.ObjectID, tags []string) error
	DecrTags(ctx context.Context, userID primitive.ObjectID, tags []string) error
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.InsightTag, error)
}

// InsightLikeRepoInterface 感悟点赞仓库接口
type InsightLikeRepoInterface interface {
	Toggle(ctx context.Context, insightID, userID primitive.ObjectID) (bool, error)
	ToggleWithSession(sessCtx mongo.SessionContext, insightID, userID primitive.ObjectID) (bool, error)
}

var _ InsightRepoInterface = (*InsightRepo)(nil)
var _ InsightTagRepoInterface = (*InsightTagRepo)(nil)
var _ InsightLikeRepoInterface = (*InsightLikeRepo)(nil)
