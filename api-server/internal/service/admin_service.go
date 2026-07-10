package service

import (
	"context"
	"crypto/subtle"
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

type DashboardStats struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	BannedUsers   int64 `json:"banned_users"`
	TotalCheckins int64 `json:"total_checkins"`
}

func (s *AdminService) Login(username, password string) (string, error) {
	if s.cfg.Admin == nil {
		return "", fmt.Errorf("admin not configured")
	}
	// 常量时间比较防止时序攻击
	usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(s.cfg.Admin.Username))
	passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(s.cfg.Admin.Password))
	if usernameMatch&passwordMatch != 1 {
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtMgr.GenerateWithRole("admin", model.UserRoleAdmin)
	if err != nil {
		return "", fmt.Errorf("generate token failed: %w", err)
	}
	return token, nil
}

func (s *AdminService) GetUsers(ctx context.Context, page, pageSize int, keyword string) ([]*model.User, int64, error) {
	page, pageSize = validatePagination(page, pageSize)
	return s.userRepo.FindAll(ctx, page, pageSize, keyword)
}

func (s *AdminService) BanUser(ctx context.Context, userID primitive.ObjectID) error {
	return s.userRepo.Update(ctx, userID, bson.M{"status": model.UserStatusBanned})
}

func (s *AdminService) UnbanUser(ctx context.Context, userID primitive.ObjectID) error {
	return s.userRepo.Update(ctx, userID, bson.M{"status": model.UserStatusActive})
}

func (s *AdminService) GetStats(ctx context.Context) (*DashboardStats, error) {
	totalUsers, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	activeUsers, err := s.userRepo.CountByStatus(ctx, model.UserStatusActive)
	if err != nil {
		return nil, fmt.Errorf("count active users: %w", err)
	}
	bannedUsers, err := s.userRepo.CountByStatus(ctx, model.UserStatusBanned)
	if err != nil {
		return nil, fmt.Errorf("count banned users: %w", err)
	}
	totalCheckins, err := s.checkinRepo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("count checkins: %w", err)
	}

	return &DashboardStats{
		TotalUsers:    totalUsers,
		ActiveUsers:   activeUsers,
		BannedUsers:   bannedUsers,
		TotalCheckins: totalCheckins,
	}, nil
}

func (s *AdminService) GetCheckins(ctx context.Context, page, pageSize int) ([]*model.Checkin, int64, error) {
	page, pageSize = validatePagination(page, pageSize)
	return s.checkinRepo.ListAll(ctx, page, pageSize)
}

func (s *AdminService) GetInsights(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error) {
	page, pageSize = validatePagination(page, pageSize)
	return s.insightRepo.ListAll(ctx, page, pageSize)
}

func (s *AdminService) DeleteCheckin(ctx context.Context, id primitive.ObjectID) error {
	return s.checkinRepo.DeleteByID(ctx, id)
}

func (s *AdminService) DeleteInsight(ctx context.Context, id primitive.ObjectID) error {
	return s.insightRepo.DeleteByID(ctx, id)
}

// validatePagination 验证分页参数
func validatePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}
