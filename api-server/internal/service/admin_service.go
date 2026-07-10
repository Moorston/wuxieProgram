package service

import (
	"context"
	"crypto/subtle"
	"fmt"
	"sync"
	"time"

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
	auditLog    repository.AuditLogInterface
	jwtMgr      *jwt.JWTManager
	cfg         *config.Config
	logger      *zap.Logger

	// 统计缓存
	statsCache     *DashboardStats
	statsCacheTime time.Time
	statsMu        sync.RWMutex
	statsTTL       time.Duration
}

func NewAdminService(
	userRepo repository.UserRepoInterface,
	checkinRepo repository.CheckinRepoInterface,
	insightRepo repository.InsightRepoInterface,
	auditLog repository.AuditLogInterface,
	jwtMgr *jwt.JWTManager,
	cfg *config.Config,
	logger *zap.Logger,
) *AdminService {
	return &AdminService{
		userRepo:    userRepo,
		checkinRepo: checkinRepo,
		insightRepo: insightRepo,
		auditLog:    auditLog,
		jwtMgr:      jwtMgr,
		cfg:         cfg,
		logger:      logger,
		statsTTL:    60 * time.Second, // 统计缓存 60 秒
	}
}

type DashboardStats struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	BannedUsers   int64 `json:"banned_users"`
	TotalCheckins int64 `json:"total_checkins"`
}

type UserDetail struct {
	User     *model.User      `json:"user"`
	Checkins []*model.Checkin  `json:"checkins"`
	Insights []*model.Insight  `json:"insights"`
}

func (s *AdminService) Login(username, password string) (string, error) {
	if s.cfg.Admin == nil {
		return "", fmt.Errorf("admin not configured")
	}
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

func (s *AdminService) GetUserDetail(ctx context.Context, userID primitive.ObjectID) (*UserDetail, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	checkins, _, err := s.checkinRepo.ListByUser(ctx, userID, 1, 20)
	if err != nil {
		s.logger.Warn("get user detail: load checkins failed", zap.String("user_id", userID.Hex()), zap.Error(err))
		checkins = []*model.Checkin{}
	}
	insights, _, err := s.insightRepo.ListByUser(ctx, userID, "", "", 1, 20)
	if err != nil {
		s.logger.Warn("get user detail: load insights failed", zap.String("user_id", userID.Hex()), zap.Error(err))
		insights = []*model.Insight{}
	}

	return &UserDetail{
		User:     user,
		Checkins: checkins,
		Insights: insights,
	}, nil
}

func (s *AdminService) BanUser(ctx context.Context, userID primitive.ObjectID, adminUser, ip string) error {
	err := s.userRepo.Update(ctx, userID, bson.M{"status": model.UserStatusBanned})
	if err == nil {
		s.logAction(ctx, adminUser, "ban_user", userID.Hex(), "user", "封禁用户", ip)
	}
	return err
}

func (s *AdminService) UnbanUser(ctx context.Context, userID primitive.ObjectID, adminUser, ip string) error {
	err := s.userRepo.Update(ctx, userID, bson.M{"status": model.UserStatusActive})
	if err == nil {
		s.logAction(ctx, adminUser, "unban_user", userID.Hex(), "user", "解封用户", ip)
	}
	return err
}

func (s *AdminService) GetStats(ctx context.Context) (*DashboardStats, error) {
	// 检查缓存
	s.statsMu.RLock()
	if s.statsCache != nil && time.Since(s.statsCacheTime) < s.statsTTL {
		cached := *s.statsCache // 返回副本
		s.statsMu.RUnlock()
		return &cached, nil
	}
	s.statsMu.RUnlock()

	// 缓存未命中，查询数据库
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

	stats := &DashboardStats{
		TotalUsers:    totalUsers,
		ActiveUsers:   activeUsers,
		BannedUsers:   bannedUsers,
		TotalCheckins: totalCheckins,
	}

	// 写入缓存
	s.statsMu.Lock()
	s.statsCache = stats
	s.statsCacheTime = time.Now()
	s.statsMu.Unlock()

	return stats, nil
}

func (s *AdminService) GetCheckins(ctx context.Context, page, pageSize int) ([]*model.Checkin, int64, error) {
	page, pageSize = validatePagination(page, pageSize)
	return s.checkinRepo.ListAll(ctx, page, pageSize)
}

func (s *AdminService) GetInsights(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error) {
	page, pageSize = validatePagination(page, pageSize)
	return s.insightRepo.ListAll(ctx, page, pageSize)
}

func (s *AdminService) DeleteCheckin(ctx context.Context, id primitive.ObjectID, adminUser, ip string) error {
	err := s.checkinRepo.DeleteByID(ctx, id)
	if err == nil {
		s.logAction(ctx, adminUser, "delete_checkin", id.Hex(), "checkin", "删除打卡", ip)
	}
	return err
}

func (s *AdminService) DeleteInsight(ctx context.Context, id primitive.ObjectID, adminUser, ip string) error {
	err := s.insightRepo.DeleteByID(ctx, id)
	if err == nil {
		s.logAction(ctx, adminUser, "delete_insight", id.Hex(), "insight", "删除感悟", ip)
	}
	return err
}

// BatchBanUsers 批量封禁用户
func (s *AdminService) BatchBanUsers(ctx context.Context, userIDs []primitive.ObjectID, adminUser, ip string) (int, error) {
	count := 0
	for _, id := range userIDs {
		if err := s.userRepo.Update(ctx, id, bson.M{"status": model.UserStatusBanned}); err != nil {
			s.logger.Warn("batch ban: failed for user", zap.String("user_id", id.Hex()), zap.Error(err))
		} else {
			count++
		}
	}
	s.logAction(ctx, adminUser, "batch_ban_users", fmt.Sprintf("%d users", count), "user", fmt.Sprintf("批量封禁 %d/%d 个用户", count, len(userIDs)), ip)
	return count, nil
}

// BatchDeleteCheckins 批量删除打卡
func (s *AdminService) BatchDeleteCheckins(ctx context.Context, ids []primitive.ObjectID, adminUser, ip string) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.checkinRepo.DeleteByID(ctx, id); err != nil {
			s.logger.Warn("batch delete: failed for checkin", zap.String("checkin_id", id.Hex()), zap.Error(err))
		} else {
			count++
		}
	}
	s.logAction(ctx, adminUser, "batch_delete_checkins", fmt.Sprintf("%d checkins", count), "checkin", fmt.Sprintf("批量删除 %d/%d 条打卡", count, len(ids)), ip)
	return count, nil
}

// BatchDeleteInsights 批量删除感悟
func (s *AdminService) BatchDeleteInsights(ctx context.Context, ids []primitive.ObjectID, adminUser, ip string) (int, error) {
	count := 0
	for _, id := range ids {
		if err := s.insightRepo.DeleteByID(ctx, id); err != nil {
			s.logger.Warn("batch delete: failed for insight", zap.String("insight_id", id.Hex()), zap.Error(err))
		} else {
			count++
		}
	}
	s.logAction(ctx, adminUser, "batch_delete_insights", fmt.Sprintf("%d insights", count), "insight", fmt.Sprintf("批量删除 %d/%d 条感悟", count, len(ids)), ip)
	return count, nil
}

// GetAuditLogs 获取操作日志
func (s *AdminService) GetAuditLogs(ctx context.Context, page, pageSize int) ([]*model.AuditLog, int64, error) {
	page, pageSize = validatePagination(page, pageSize)
	return s.auditLog.List(ctx, page, pageSize)
}

// GetSystemConfig 获取系统配置（脱敏）
func (s *AdminService) GetSystemConfig() map[string]interface{} {
	return map[string]interface{}{
		"server_port":    s.cfg.Server.Port,
		"server_mode":    s.cfg.Server.Mode,
		"jwt_expires_h":  s.cfg.JWT.Expires,
		"cors_origins":   s.cfg.CORS.AllowedOrigins,
		"wx_app_id":      s.cfg.WX.AppID,
		"media_url":      s.cfg.MediaURL,
	}
}

// logAction 记录操作日志
func (s *AdminService) logAction(ctx context.Context, adminUser, action, targetID, targetType, detail, ip string) {
	if s.auditLog == nil {
		return
	}
	if err := s.auditLog.Create(ctx, &model.AuditLog{
		AdminUser:  adminUser,
		Action:     action,
		TargetID:   targetID,
		TargetType: targetType,
		Detail:     detail,
		IP:         ip,
	}); err != nil {
		s.logger.Error("failed to write audit log", zap.String("action", action), zap.Error(err))
	}
}

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
