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

func TestUserService_GetProfile(t *testing.T) {
	userID := primitive.NewObjectID()

	mockUser := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			assert.Equal(t, userID, id)
			return &model.User{ID: userID, Nickname: "TestUser", Avatar: "http://avatar.url"}, nil
		},
	}
	svc := NewUserService(mockUser)

	user, err := svc.GetProfile(context.Background(), userID)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "TestUser", user.Nickname)
	assert.Equal(t, userID, user.ID)
}

func TestUserService_GetProfile_NotFound(t *testing.T) {
	mockUser := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			return nil, assert.AnError
		},
	}
	svc := NewUserService(mockUser)

	_, err := svc.GetProfile(context.Background(), primitive.NewObjectID())
	assert.Error(t, err)
}

func TestUserService_UpdateProfile(t *testing.T) {
	userID := primitive.NewObjectID()

	mockUser := &mockUserRepo{
		update: func(ctx context.Context, id primitive.ObjectID, update bson.M) error {
			assert.Equal(t, userID, id)
			assert.Equal(t, "新昵称", update["nickname"])
			assert.Equal(t, "http://new.avatar", update["avatar"])
			return nil
		},
	}
	svc := NewUserService(mockUser)

	err := svc.UpdateProfile(context.Background(), userID, "新昵称", "http://new.avatar")
	require.NoError(t, err)
}

func TestUserService_UpdateProfile_NicknameOnly(t *testing.T) {
	userID := primitive.NewObjectID()

	mockUser := &mockUserRepo{
		update: func(ctx context.Context, id primitive.ObjectID, update bson.M) error {
			assert.Equal(t, "onlynick", update["nickname"])
			_, hasAvatar := update["avatar"]
			assert.False(t, hasAvatar, "avatar should not be in update when empty")
			return nil
		},
	}
	svc := NewUserService(mockUser)

	err := svc.UpdateProfile(context.Background(), userID, "onlynick", "")
	require.NoError(t, err)
}

func TestUserService_UpdateProfile_Empty(t *testing.T) {
	userID := primitive.NewObjectID()

	var called bool
	mockUser := &mockUserRepo{
		update: func(ctx context.Context, id primitive.ObjectID, update bson.M) error {
			called = true
			assert.Empty(t, update["nickname"])
			assert.Empty(t, update["avatar"])
			return nil
		},
	}
	svc := NewUserService(mockUser)

	err := svc.UpdateProfile(context.Background(), userID, "", "")
	require.NoError(t, err)
	assert.True(t, called, "update should be called even with empty strings")
}