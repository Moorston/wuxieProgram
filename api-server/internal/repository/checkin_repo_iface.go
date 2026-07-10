package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate mockgen -destination=mock_checkin_repo.go -package=repository wuxie-api/internal/repository CheckinRepoInterface

// CheckinRepoInterface 打卡仓库接口
type CheckinRepoInterface interface {
	Create(ctx context.Context, c *model.Checkin) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error)
	UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.CheckinStatus, videoURL, coverURL string, duration float64) error
	List(ctx context.Context, userID primitive.ObjectID, groupUserIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error)
	ListByUser(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error)
	ListByUserIDs(ctx context.Context, userIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*model.Checkin, int64, error)
	Delete(ctx context.Context, id, userID primitive.ObjectID) error
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	IncrLikeCount(ctx context.Context, id primitive.ObjectID, delta int) error
	IncrCommentCount(ctx context.Context, id primitive.ObjectID) error
	IncrLikeCountWithSession(sessCtx mongo.SessionContext, id primitive.ObjectID, delta int) error
	IncrCommentCountWithSession(sessCtx mongo.SessionContext, id primitive.ObjectID) error
	Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Checkin, int64, error)
	Aggregate(ctx context.Context, pipeline []bson.M) (mongo.Cursor, error)
	CountAll(ctx context.Context) (int64, error)
}

var _ CheckinRepoInterface = (*CheckinRepo)(nil)
