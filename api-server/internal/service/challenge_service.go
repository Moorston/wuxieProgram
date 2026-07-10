package service

import (
	"context"
	"fmt"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type ChallengeService struct {
	challengeRepo repository.ChallengeRepoInterface
	participantRepo repository.ChallengeParticipantRepoInterface
	userRepo      repository.UserRepoInterface
	logger        *zap.Logger
}

func NewChallengeService(
	challengeRepo repository.ChallengeRepoInterface,
	participantRepo repository.ChallengeParticipantRepoInterface,
	userRepo repository.UserRepoInterface,
	logger *zap.Logger,
) *ChallengeService {
	return &ChallengeService{
		challengeRepo:  challengeRepo,
		participantRepo: participantRepo,
		userRepo:       userRepo,
		logger:         logger,
	}
}

// CreateChallenge 创建挑战
func (s *ChallengeService) CreateChallenge(ctx context.Context, creatorID primitive.ObjectID, title, description string, duration int) (*model.Challenge, error) {
	if duration < 1 || duration > 365 {
		return nil, ErrInvalidDuration
	}

	now := time.Now()
	challenge := &model.Challenge{
		Title:       title,
		Description: description,
		Duration:    duration,
		CreatorID:   creatorID,
		StartDate:   now,
		EndDate:     now.AddDate(0, 0, duration),
		Status:      model.ChallengeStatusActive,
	}

	if err := s.challengeRepo.Create(ctx, challenge); err != nil {
		return nil, fmt.Errorf("create challenge failed: %w", err)
	}

	// 创建者自动参与
	if err := s.challengeRepo.AddParticipant(ctx, challenge.ID, creatorID); err != nil {
		s.logger.Warn("add creator to challenge failed", zap.Error(err))
	}
	_ = s.participantRepo.Upsert(ctx, &model.ChallengeParticipant{
		ChallengeID: challenge.ID,
		UserID:      creatorID,
	})

	return challenge, nil
}

// ListChallenges 获取进行中的挑战列表
func (s *ChallengeService) ListChallenges(ctx context.Context, page, pageSize int) ([]*model.Challenge, int64, error) {
	return s.challengeRepo.ListActive(ctx, page, pageSize)
}

// GetChallenge 获取挑战详情
func (s *ChallengeService) GetChallenge(ctx context.Context, id primitive.ObjectID) (*model.Challenge, error) {
	challenge, err := s.challengeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, ErrChallengeNotFound
	}

	// 填充参与者信息
	participants, err := s.participantRepo.ListByChallenge(ctx, id)
	if err != nil {
		s.logger.Warn("get challenge: load participants failed", zap.String("challenge_id", id.Hex()), zap.Error(err))
		participants = []*model.ChallengeParticipant{}
	}
	userIDs := make([]primitive.ObjectID, 0, len(participants))
	for _, p := range participants {
		userIDs = append(userIDs, p.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		s.logger.Warn("get challenge: load users failed", zap.Error(err))
		users = []*model.User{}
	}
	userMap := make(map[primitive.ObjectID]*model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}
	for _, p := range participants {
		p.User = userMap[p.UserID]
	}
	challenge.Participants = participants

	// 填充创建者信息
	creator, err := s.userRepo.FindByID(ctx, challenge.CreatorID)
	if err != nil {
		s.logger.Warn("get challenge: load creator failed", zap.String("challenge_id", id.Hex()), zap.Error(err))
	}
	challenge.Creator = creator

	return challenge, nil
}

// JoinChallenge 参加挑战
func (s *ChallengeService) JoinChallenge(ctx context.Context, challengeID, userID primitive.ObjectID) (*model.ChallengeParticipant, error) {
	challenge, err := s.challengeRepo.FindByID(ctx, challengeID)
	if err != nil {
		return nil, ErrChallengeNotFound
	}

	if !challenge.IsActive() {
		return nil, ErrChallengeNotActive
	}

	if challenge.HasParticipant(userID) {
		return nil, ErrAlreadyJoined
	}

	if err := s.challengeRepo.AddParticipant(ctx, challengeID, userID); err != nil {
		return nil, fmt.Errorf("add participant failed: %w", err)
	}

	participant := &model.ChallengeParticipant{
		ChallengeID: challengeID,
		UserID:      userID,
	}
	if err := s.participantRepo.Upsert(ctx, participant); err != nil {
		return nil, fmt.Errorf("create participant record failed: %w", err)
	}

	return participant, nil
}

// RecordCheckin 记录打卡（在挑战期间打卡时调用）
// 应由 CheckinService 在打卡成功后调用
func (s *ChallengeService) RecordCheckin(ctx context.Context, userID primitive.ObjectID) {
	// 查询用户参与的所有进行中挑战
	participants, err := s.participantRepo.ListByUser(ctx, userID)
	if err != nil {
		s.logger.Warn("record checkin: list user challenges failed", zap.String("user_id", userID.Hex()), zap.Error(err))
		return
	}

	now := time.Now()
	for _, p := range participants {
		challenge, err := s.challengeRepo.FindByID(ctx, p.ChallengeID)
		if err != nil || !challenge.IsActive() {
			continue
		}
		// 检查今天是否已打卡（防止重复计数）
		if err := s.participantRepo.IncrementCompletedDays(ctx, userID, p.ChallengeID, challenge.Duration); err != nil {
			s.logger.Warn("record checkin: increment failed",
				zap.String("user_id", userID.Hex()),
				zap.String("challenge_id", p.ChallengeID.Hex()),
				zap.Error(err),
			)
		} else {
			s.logger.Info("challenge checkin recorded",
				zap.String("user_id", userID.Hex()),
				zap.String("challenge_id", p.ChallengeID.Hex()),
				zap.Time("time", now),
			)
		}
	}
}

// GetChallengeRanking 获取挑战排行榜
type ChallengeRankingEntry struct {
	Rank           int                `json:"rank"`
	User           *model.User        `json:"user"`
	CompletedDays  int                `json:"completed_days"`
	Progress       float64            `json:"progress"`
	IsCompleted    bool               `json:"is_completed"`
}

func (s *ChallengeService) GetChallengeRanking(ctx context.Context, challengeID primitive.ObjectID) ([]ChallengeRankingEntry, error) {
	participants, err := s.participantRepo.ListByChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(participants))
	for _, p := range participants {
		userIDs = append(userIDs, p.UserID)
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		s.logger.Warn("get challenge ranking: load users failed", zap.Error(err))
		users = []*model.User{}
	}
	userMap := make(map[primitive.ObjectID]*model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	ranking := make([]ChallengeRankingEntry, 0, len(participants))
	for i, p := range participants {
		ranking = append(ranking, ChallengeRankingEntry{
			Rank:          i + 1,
			User:          userMap[p.UserID],
			CompletedDays: p.CompletedDays,
			Progress:      p.Progress,
			IsCompleted:   p.IsCompleted,
		})
	}

	return ranking, nil
}

// 错误定义
var (
	ErrChallengeNotFound = &challengeError{"challenge not found"}
	ErrChallengeNotActive = &challengeError{"challenge is not active"}
	ErrAlreadyJoined     = &challengeError{"already joined this challenge"}
	ErrInvalidDuration   = &challengeError{"invalid duration: must be between 1 and 365 days"}
)

type challengeError struct{ msg string }

func (e *challengeError) Error() string { return e.msg }
