package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type BadgeService struct {
	badgeRepo    repository.BadgeRepoInterface
	userBadgeRepo repository.UserBadgeRepoInterface
	userRepo     repository.UserRepoInterface
	logger       *zap.Logger
}

func NewBadgeService(
	badgeRepo repository.BadgeRepoInterface,
	userBadgeRepo repository.UserBadgeRepoInterface,
	userRepo repository.UserRepoInterface,
	logger *zap.Logger,
) *BadgeService {
	return &BadgeService{
		badgeRepo:    badgeRepo,
		userBadgeRepo: userBadgeRepo,
		userRepo:     userRepo,
		logger:       logger,
	}
}

// GetAllBadges 获取所有徽章定义
func (s *BadgeService) GetAllBadges(ctx context.Context) ([]*model.Badge, error) {
	return s.badgeRepo.FindAll(ctx)
}

// GetUserBadges 获取用户已获得的徽章
func (s *BadgeService) GetUserBadges(ctx context.Context, userID primitive.ObjectID) ([]*model.UserBadge, error) {
	badges, err := s.userBadgeRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 填充徽章详情
	allBadges, _ := s.badgeRepo.FindAll(ctx)
	badgeMap := make(map[primitive.ObjectID]*model.Badge, len(allBadges))
	for _, b := range allBadges {
		badgeMap[b.ID] = b
	}
	for _, ub := range badges {
		ub.Badge = badgeMap[ub.BadgeID]
	}

	return badges, nil
}

// CheckAndGrantBadges 检查并授予符合条件的徽章
// 应在用户打卡、积分变动等场景调用
func (s *BadgeService) CheckAndGrantBadges(ctx context.Context, userID primitive.ObjectID, streakDays, totalCheckins, totalScore int) ([]*model.Badge, error) {
	allBadges, err := s.badgeRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var granted []*model.Badge

	for _, badge := range allBadges {
		// 检查是否已拥有
		has, _ := s.userBadgeRepo.HasBadge(ctx, userID, badge.ID)
		if has {
			continue
		}

		// 检查是否满足条件
		if s.checkCondition(badge, streakDays, totalCheckins, totalScore) {
			if err := s.userBadgeRepo.Grant(ctx, userID, badge.ID); err != nil {
				s.logger.Warn("grant badge failed",
					zap.String("user_id", userID.Hex()),
					zap.String("badge", badge.Name),
					zap.Error(err),
				)
				continue
			}
			granted = append(granted, badge)
			s.logger.Info("badge granted",
				zap.String("user_id", userID.Hex()),
				zap.String("badge", badge.Name),
				zap.String("icon", badge.Icon),
			)
		}
	}

	return granted, nil
}

// checkCondition 检查徽章条件是否满足
func (s *BadgeService) checkCondition(badge *model.Badge, streakDays, totalCheckins, totalScore int) bool {
	switch badge.Type {
	case model.BadgeTypeStreak:
		return streakDays >= badge.Condition
	case model.BadgeTypeTotal:
		return totalCheckins >= badge.Condition
	case model.BadgeTypeScore:
		return totalScore >= badge.Condition
	default:
		return false
	}
}

// SeedDefaults 初始化默认徽章数据
func (s *BadgeService) SeedDefaults(ctx context.Context) error {
	return s.badgeRepo.SeedDefaults(ctx)
}
