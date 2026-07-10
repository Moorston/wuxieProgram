package service

import (
	"context"
	"fmt"

	"wuxie-api/internal/config"
	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"
	"wuxie-api/pkg/jwt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type AdminService struct {
	userRepo    repository.UserRepoInterface
	checkinRepo repository.CheckinRepoInterface
	insightRepo repository.InsightRepoInterface
	jwtMgr      *jwt.JWTManager
	cfg         *config.Config
	logger      *zap.Logger
}

func NewAdminService(
	userRepo repository.UserRepoInterface,
	checkinRepo repository.CheckinRepoInterface,
	insightRepo repository.InsightRepoInterface,
	jwtMgr *jwt.JWTManager,
	cfg *config.Config,
	logger *zap.Logger,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		checkinRepo: checkinRepo,
		insightRepo: insightRepo,
		jwtMgr:      jwtMgr,
		cfg:         cfg,
		logger:      logger,
	}
}

func (s *AdminService) Login(username, password string) (string, error) {
	if s.cfg.Admin == nil {
		return "", fmt.Errorf("admin not configured")
	}
	if username != s.cfg.Admin.Username || password != s.cfg.Admin.Password {
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtMgr.GenerateWithRole("admin", model.UserRoleAdmin)
	if err != nil {
		return "", fmt.Errorf("generate token failed: %w", err)
	}
	return token, nil
}

func (s *AdminService) GetUsers(ctx context.Context, page, pageSize int, keyword string) ([]*model.User, int64, error) {
	return s.userRepo.FindByIDs(ctx, nil), 0, fmt.Errorf("not implemented")
}

func (s *AdminService) BanUser(ctx context.Context, userID primitive.ObjectID) error {
	return s.userRepo.Update(ctx, userID, bson.M{"status": model.UserStatusBanned})
}

func (s *AdminService) UnbanUser(ctx context.Context, userID primitive.ObjectID) error {
	return s.userRepo.Update(ctx, userID, bson.M{"status": model.UserStatusActive})
}

func (s *AdminService) DeleteCheckin(ctx context.Context, id primitive.ObjectID) error {
	return s.checkinRepo.DeleteByID(ctx, id)
}

func (s *AdminService) DeleteInsight(ctx context.Context, id primitive.ObjectID) error {
	return s.insightRepo.DeleteByID(ctx, id)
}