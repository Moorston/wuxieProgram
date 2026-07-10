package service

import (
	"context"
	"testing"

	"wuxie-api/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNotificationService_List(t *testing.T) {
	userID := primitive.NewObjectID()
	notifID := primitive.NewObjectID()
	senderID := primitive.NewObjectID()

	mockNotif := &mockNotificationRepo{
		listFn: func(ctx context.Context, uid primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, 1, page)
			assert.Equal(t, 20, pageSize)
			return []*model.Notification{
				{ID: notifID, UserID: userID, Title: "通知1", SenderID: senderID},
			}, 1, nil
		},
	}
	mockSettings := &mockNotifSettingsRepo{}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{{ID: senderID, Nickname: "Sender"}}, nil
		},
	}
	svc := NewNotificationService(mockNotif, mockSettings, mockUser)

	notifications, total, err := svc.List(context.Background(), userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, notifications, 1)
	assert.Equal(t, "通知1", notifications[0].Title)
	assert.NotNil(t, notifications[0].Sender)
	assert.Equal(t, "Sender", notifications[0].Sender.Nickname)
}

func TestNotificationService_List_NoSenders(t *testing.T) {
	userID := primitive.NewObjectID()

	mockNotif := &mockNotificationRepo{
		listFn: func(ctx context.Context, uid primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error) {
			return []*model.Notification{
				{ID: primitive.NewObjectID(), UserID: userID, Title: "系统通知"},
			}, 1, nil
		},
	}
	svc := NewNotificationService(mockNotif, &mockNotifSettingsRepo{}, &mockUserRepo{})

	notifications, total, err := svc.List(context.Background(), userID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
}

func TestNotificationService_UnreadCount(t *testing.T) {
	userID := primitive.NewObjectID()

	mockNotif := &mockNotificationRepo{
		unreadCountFn: func(ctx context.Context, uid primitive.ObjectID) (int64, error) {
			assert.Equal(t, userID, uid)
			return 5, nil
		},
	}
	svc := NewNotificationService(mockNotif, &mockNotifSettingsRepo{}, &mockUserRepo{})

	count, err := svc.UnreadCount(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestNotificationService_MarkRead(t *testing.T) {
	notifID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockNotif := &mockNotificationRepo{
		markReadFn: func(ctx context.Context, id, uid primitive.ObjectID) error {
			assert.Equal(t, notifID, id)
			assert.Equal(t, userID, uid)
			return nil
		},
	}
	svc := NewNotificationService(mockNotif, &mockNotifSettingsRepo{}, &mockUserRepo{})

	err := svc.MarkRead(context.Background(), notifID, userID)
	require.NoError(t, err)
}

func TestNotificationService_MarkAllRead(t *testing.T) {
	userID := primitive.NewObjectID()

	mockNotif := &mockNotificationRepo{
		markAllReadFn: func(ctx context.Context, uid primitive.ObjectID) error {
			assert.Equal(t, userID, uid)
			return nil
		},
	}
	svc := NewNotificationService(mockNotif, &mockNotifSettingsRepo{}, &mockUserRepo{})

	err := svc.MarkAllRead(context.Background(), userID)
	require.NoError(t, err)
}

func TestNotificationService_Delete(t *testing.T) {
	notifID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockNotif := &mockNotificationRepo{
		deleteFn: func(ctx context.Context, id, uid primitive.ObjectID) error {
			assert.Equal(t, notifID, id)
			assert.Equal(t, userID, uid)
			return nil
		},
	}
	svc := NewNotificationService(mockNotif, &mockNotifSettingsRepo{}, &mockUserRepo{})

	err := svc.Delete(context.Background(), notifID, userID)
	require.NoError(t, err)
}

func TestNotificationService_Send(t *testing.T) {
	userID := primitive.NewObjectID()
	senderID := primitive.NewObjectID()
	targetID := primitive.NewObjectID()

	mockSettings := &mockNotifSettingsRepo{
		getOrCreateFn: func(ctx context.Context, uid primitive.ObjectID) (*model.NotificationSettings, error) {
			return &model.NotificationSettings{
				UserID:      uid,
				LikeNotify:  true,
				CommentNotify: true,
			}, nil
		},
	}
	var createdNotif *model.Notification
	mockNotif := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *model.Notification) error {
			createdNotif = n
			return nil
		},
	}
	svc := NewNotificationService(mockNotif, mockSettings, &mockUserRepo{})

	err := svc.Send(context.Background(), userID, senderID, model.NotifTypeLike, "赞了你的打卡", "", "checkin", targetID)
	require.NoError(t, err)
	require.NotNil(t, createdNotif)
	assert.Equal(t, userID, createdNotif.UserID)
	assert.Equal(t, model.NotifTypeLike, createdNotif.Type)
	assert.Equal(t, "赞了你的打卡", createdNotif.Title)
}

func TestNotificationService_Send_Disabled(t *testing.T) {
	userID := primitive.NewObjectID()
	senderID := primitive.NewObjectID()

	mockSettings := &mockNotifSettingsRepo{
		getOrCreateFn: func(ctx context.Context, uid primitive.ObjectID) (*model.NotificationSettings, error) {
			return &model.NotificationSettings{
				LikeNotify: false, // 点赞通知已关闭
			}, nil
		},
	}
	var createCalled bool
	mockNotif := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *model.Notification) error {
			createCalled = true
			return nil
		},
	}
	svc := NewNotificationService(mockNotif, mockSettings, &mockUserRepo{})

	err := svc.Send(context.Background(), userID, senderID, model.NotifTypeLike, "test", "", "checkin", primitive.NewObjectID())
	require.NoError(t, err)
	assert.False(t, createCalled, "should not create notification when disabled")
}

func TestNotificationService_SendBatch(t *testing.T) {
	userID1 := primitive.NewObjectID()
	userID2 := primitive.NewObjectID()
	senderID := primitive.NewObjectID()

	var createdCount int
	mockSettings := &mockNotifSettingsRepo{
		getOrCreateFn: func(ctx context.Context, uid primitive.ObjectID) (*model.NotificationSettings, error) {
			return &model.NotificationSettings{LikeNotify: true}, nil
		},
	}
	mockNotif := &mockNotificationRepo{
		createFn: func(ctx context.Context, n *model.Notification) error {
			createdCount++
			return nil
		},
	}
	svc := NewNotificationService(mockNotif, mockSettings, &mockUserRepo{})

	err := svc.SendBatch(context.Background(), []primitive.ObjectID{userID1, senderID, userID2}, senderID, model.NotifTypeLike, "batch", "", "checkin", primitive.NewObjectID())
	require.NoError(t, err)
	// senderID should be skipped, so only 2 notifications
	assert.Equal(t, 2, createdCount)
}

func TestNotificationService_GetSettings(t *testing.T) {
	userID := primitive.NewObjectID()

	mockSettings := &mockNotifSettingsRepo{
		getOrCreateFn: func(ctx context.Context, uid primitive.ObjectID) (*model.NotificationSettings, error) {
			return &model.NotificationSettings{
				UserID:     uid,
				LikeNotify: true,
			}, nil
		},
	}
	svc := NewNotificationService(&mockNotificationRepo{}, mockSettings, &mockUserRepo{})

	settings, err := svc.GetSettings(context.Background(), userID)
	require.NoError(t, err)
	assert.True(t, settings.LikeNotify)
}

func TestNotificationService_UpdateSettings(t *testing.T) {
	userID := primitive.NewObjectID()
	likeNotify := true

	mockSettings := &mockNotifSettingsRepo{
		updateFn: func(ctx context.Context, uid primitive.ObjectID, update bson.M) error {
			assert.Equal(t, userID, uid)
			assert.Equal(t, true, update["like_notify"])
			return nil
		},
	}
	svc := NewNotificationService(&mockNotificationRepo{}, mockSettings, &mockUserRepo{})

	err := svc.UpdateSettings(context.Background(), userID, &likeNotify, nil, nil, nil, nil)
	require.NoError(t, err)
}

func TestNotificationService_UpdateSettings_NoChanges(t *testing.T) {
	mockSettings := &mockNotifSettingsRepo{
		updateFn: func(ctx context.Context, uid primitive.ObjectID, update bson.M) error {
			t.Error("should not be called")
			return nil
		},
	}
	svc := NewNotificationService(&mockNotificationRepo{}, mockSettings, &mockUserRepo{})

	err := svc.UpdateSettings(context.Background(), primitive.NewObjectID(), nil, nil, nil, nil, nil)
	require.NoError(t, err)
}