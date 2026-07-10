package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -destination=mock_notification_repo.go -package=repository wuxie-api/internal/repository NotificationRepoInterface
//go:generate mockgen -destination=mock_notification_settings_repo.go -package=repository wuxie-api/internal/repository NotificationSettingsRepoInterface

// NotificationRepoInterface 通知仓库接口
type NotificationRepoInterface interface {
	Create(ctx context.Context, n *model.Notification) error
	List(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error)
	UnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error)
	MarkRead(ctx context.Context, id, userID primitive.ObjectID) error
	MarkAllRead(ctx context.Context, userID primitive.ObjectID) error
	Delete(ctx context.Context, id, userID primitive.ObjectID) error
}

// NotificationSettingsRepoInterface 通知设置仓库接口
type NotificationSettingsRepoInterface interface {
	GetOrCreate(ctx context.Context, userID primitive.ObjectID) (*model.NotificationSettings, error)
	Update(ctx context.Context, userID primitive.ObjectID, update bson.M) error
}

var _ NotificationRepoInterface = (*NotificationRepo)(nil)
var _ NotificationSettingsRepoInterface = (*NotificationSettingsRepo)(nil)
