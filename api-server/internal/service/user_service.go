package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

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