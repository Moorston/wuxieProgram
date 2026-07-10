package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	userRepo repository.UserRepoInterface
}

func NewUserService(userRepo repository.UserRepoInterface) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetProfile(ctx context.Context, userID primitive.ObjectID) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID primitive.ObjectID, nickname, avatar string) error {
	update := map[string]interface{}{}
	if nickname != "" {
		update["nickname"] = nickname
	}
	if avatar != "" {
		update["avatar"] = avatar
	}
	return s.userRepo.Update(ctx, userID, update)
}

// UpdateDefaultVisibility 更新默认打卡可见性
func (s *UserService) UpdateDefaultVisibility(ctx context.Context, userID primitive.ObjectID, visibility model.Visibility) error {
	if visibility < model.VisibilityPublic || visibility > model.VisibilityPrivate {
		return ErrInvalidVisibility
	}
	return s.userRepo.Update(ctx, userID, bson.M{"default_visibility": visibility})
}

// GetPrivacySettings 获取隐私设置
func (s *UserService) GetPrivacySettings(ctx context.Context, userID primitive.ObjectID) (model.Visibility, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return model.VisibilityPublic, err
	}
	return user.DefaultVisibility, nil
}

var ErrInvalidVisibility = &visibilityError{"invalid visibility value: must be 0 (public), 1 (group), or 2 (private)"}

type visibilityError struct{ msg string }

func (e *visibilityError) Error() string { return e.msg }