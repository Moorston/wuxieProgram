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
	if !visibility.IsValid() {
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

// UserLevelInfo 用户等级信息
type UserLevelInfo struct {
	Level       model.UserLevel `json:"level"`
	NextLevel   *model.UserLevel `json:"next_level,omitempty"`
	Progress    float64          `json:"progress"` // 到下一级的进度 (0-100)
}

// GetUserLevel 获取用户等级信息
func (s *UserService) GetUserLevel(ctx context.Context, userID primitive.ObjectID) (*UserLevelInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	level := model.GetUserLevel(user.CheckDays)
	nextLevel := model.GetNextLevel(user.CheckDays)

	var progress float64
	if nextLevel != nil {
		// 计算到下一级的进度
		currentMin := level.MinDays
		nextMin := nextLevel.MinDays
		progress = float64(user.CheckDays-currentMin) / float64(nextMin-currentMin) * 100
		if progress > 100 {
			progress = 100
		}
	} else {
		progress = 100 // 已满级
	}

	return &UserLevelInfo{
		Level:     level,
		NextLevel: nextLevel,
		Progress:  progress,
	}, nil
}

var ErrInvalidVisibility = &visibilityError{"invalid visibility value: must be 0 (public), 1 (group), or 2 (private)"}

type visibilityError struct{ msg string }

func (e *visibilityError) Error() string { return e.msg }