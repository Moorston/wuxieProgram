package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate mockgen -destination=mock_comment_repo.go -package=repository wuxie-api/internal/repository CommentRepoInterface
//go:generate mockgen -destination=mock_like_repo.go -package=repository wuxie-api/internal/repository LikeRepoInterface

// CommentRepoInterface 评论仓库接口
type CommentRepoInterface interface {
	Create(ctx context.Context, c *model.Comment) error
	ListByCheckin(ctx context.Context, checkinID primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error)
	ListReplies(ctx context.Context, parentID primitive.ObjectID) ([]*model.Comment, error)
	StartSession() (mongo.Session, error)
}

// LikeRepoInterface 点赞仓库接口
type LikeRepoInterface interface {
	Toggle(ctx context.Context, checkinID, userID primitive.ObjectID) (liked bool, err error)
	ToggleWithSession(sessCtx mongo.SessionContext, checkinID, userID primitive.ObjectID) (liked bool, err error)
	IsLiked(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error)
	BatchIsLiked(ctx context.Context, checkinIDs []primitive.ObjectID, userID primitive.ObjectID) (map[primitive.ObjectID]bool, error)
	StartSession() (mongo.Session, error)
}

var _ CommentRepoInterface = (*CommentRepo)(nil)
var _ LikeRepoInterface = (*LikeRepo)(nil)
