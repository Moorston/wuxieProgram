package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type FollowService struct {
	followRepo  repository.FollowRepoInterface
	checkinRepo repository.CheckinRepoInterface
	insightRepo repository.InsightRepoInterface
	userRepo    repository.UserRepoInterface
	logger      *zap.Logger
}

func NewFollowService(
	followRepo repository.FollowRepoInterface,
	checkinRepo repository.CheckinRepoInterface,
	insightRepo repository.InsightRepoInterface,
	userRepo repository.UserRepoInterface,
	logger *zap.Logger,
) *FollowService {
	return &FollowService{
		followRepo:  followRepo,
		checkinRepo: checkinRepo,
		insightRepo: insightRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

// Follow 关注用户
func (s *FollowService) Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	if followerID == followingID {
		return ErrCannotFollowSelf
	}
	return s.followRepo.Follow(ctx, followerID, followingID)
}

// Unfollow 取消关注
func (s *FollowService) Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	return s.followRepo.Unfollow(ctx, followerID, followingID)
}

// IsFollowing 检查是否已关注
func (s *FollowService) IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error) {
	return s.followRepo.IsFollowing(ctx, followerID, followingID)
}

// GetFollowing 获取关注列表（带用户信息）
func (s *FollowService) GetFollowing(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.User, int64, error) {
	ids, total, err := s.followRepo.GetFollowing(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return []*model.User{}, 0, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, ids)
	return users, total, err
}

// GetFollowers 获取粉丝列表（带用户信息）
func (s *FollowService) GetFollowers(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.User, int64, error) {
	ids, total, err := s.followRepo.GetFollowers(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return []*model.User{}, 0, nil
	}
	users, err := s.userRepo.FindByIDs(ctx, ids)
	return users, total, err
}

// GetFollowStats 获取关注/粉丝数
func (s *FollowService) GetFollowStats(ctx context.Context, userID primitive.ObjectID) (following int64, followers int64, err error) {
	following, err = s.followRepo.CountFollowing(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	followers, err = s.followRepo.CountFollowers(ctx, userID)
	return
}

// FeedItem 动态条目
type FeedItem struct {
	Type      string      `json:"type"` // checkin, insight
	Checkin   *model.Checkin  `json:"checkin,omitempty"`
	Insight   *model.Insight  `json:"insight,omitempty"`
	CreatedAt interface{} `json:"created_at"`
}

// GetFeed 获取关注用户的动态流
func (s *FollowService) GetFeed(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]FeedItem, error) {
	// 获取关注列表
	followingIDs, _, err := s.followRepo.GetFollowing(ctx, userID, 1, 100)
	if err != nil {
		return nil, err
	}

	if len(followingIDs) == 0 {
		return []FeedItem{}, nil
	}

	// 查询关注用户的公开打卡
	var feed []FeedItem

	// 查询打卡（简化：查询所有关注用户的打卡，按时间排序）
	checkins, _, err := s.checkinRepo.ListAll(ctx, page, pageSize*2) // 多取一些再过滤
	if err == nil {
		followingSet := make(map[primitive.ObjectID]bool, len(followingIDs))
		for _, id := range followingIDs {
			followingSet[id] = true
		}
		for _, c := range checkins {
			if followingSet[c.UserID] {
				feed = append(feed, FeedItem{Type: "checkin", Checkin: c, CreatedAt: c.CreatedAt})
			}
		}
	}

	// 限制返回数量
	if len(feed) > pageSize {
		feed = feed[:pageSize]
	}

	return feed, nil
}

// UserProfile 用户主页数据
type UserProfile struct {
	User         *model.User `json:"user"`
	Following    int64       `json:"following"`
	Followers    int64       `json:"followers"`
	IsFollowing  bool        `json:"is_following"`
}

// GetUserProfile 获取用户主页
func (s *FollowService) GetUserProfile(ctx context.Context, targetID, viewerID primitive.ObjectID) (*UserProfile, error) {
	user, err := s.userRepo.FindByID(ctx, targetID)
	if err != nil {
		return nil, err
	}

	following, followers, _ := s.GetFollowStats(ctx, targetID)
	isFollowing, _ := s.IsFollowing(ctx, viewerID, targetID)

	return &UserProfile{
		User:        user,
		Following:   following,
		Followers:   followers,
		IsFollowing: isFollowing,
	}, nil
}

// 错误定义
var ErrCannotFollowSelf = &followError{"cannot follow yourself"}

type followError struct {
	msg string
}

func (e *followError) Error() string { return e.msg }
