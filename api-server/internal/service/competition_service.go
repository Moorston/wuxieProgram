package service

import (
	"context"
	"fmt"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type CompetitionService struct {
	compRepo   repository.CompetitionRepoInterface
	entryRepo  repository.CompetitionEntryRepoInterface
	checkinRepo repository.CheckinRepoInterface
	userRepo   repository.UserRepoInterface
	logger     *zap.Logger
}

func NewCompetitionService(
	compRepo repository.CompetitionRepoInterface,
	entryRepo repository.CompetitionEntryRepoInterface,
	checkinRepo repository.CheckinRepoInterface,
	userRepo repository.UserRepoInterface,
	logger *zap.Logger,
) *CompetitionService {
	return &CompetitionService{
		compRepo:   compRepo,
		entryRepo:  entryRepo,
		checkinRepo: checkinRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// CreateCompetition 创建赛事
func (s *CompetitionService) CreateCompetition(ctx context.Context, comp *model.Competition) error {
	if comp.Status == 0 {
		comp.Status = model.CompetitionStatusDraft
	}
	return s.compRepo.Create(ctx, comp)
}

// UpdateCompetition 更新赛事
func (s *CompetitionService) UpdateCompetition(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	return s.compRepo.Update(ctx, id, update)
}

// ListCompetitions 赛事列表（管理端）
func (s *CompetitionService) ListCompetitions(ctx context.Context, page, pageSize int, status *model.CompetitionStatus) ([]*model.Competition, int64, error) {
	return s.compRepo.List(ctx, page, pageSize, status)
}

// ListActiveCompetitions 进行中的赛事（用户端）
func (s *CompetitionService) ListActiveCompetitions(ctx context.Context) ([]*model.Competition, error) {
	return s.compRepo.ListActive(ctx)
}

// GetCompetition 赛事详情
func (s *CompetitionService) GetCompetition(ctx context.Context, id primitive.ObjectID) (*model.Competition, error) {
	return s.compRepo.FindByID(ctx, id)
}

// SubmitEntry 提交参赛作品
func (s *CompetitionService) SubmitEntry(ctx context.Context, competitionID, userID, checkinID primitive.ObjectID) (*model.CompetitionEntry, error) {
	// 验证赛事存在且进行中
	comp, err := s.compRepo.FindByID(ctx, competitionID)
	if err != nil {
		return nil, ErrCompetitionNotFound
	}

	now := time.Now()
	if comp.Status != model.CompetitionStatusActive || now.Before(comp.StartDate) || now.After(comp.EndDate) {
		return nil, ErrCompetitionNotActive
	}

	// 验证打卡存在且属于用户
	checkin, err := s.checkinRepo.FindByID(ctx, checkinID)
	if err != nil {
		return nil, ErrCheckinNotFound
	}
	if checkin.UserID != userID {
		return nil, ErrNotCheckinOwner
	}

	// 检查是否已提交
	existing, _ := s.entryRepo.FindByUserAndCompetition(ctx, userID, competitionID)
	if existing != nil {
		return nil, ErrAlreadySubmitted
	}

	entry := &model.CompetitionEntry{
		CompetitionID: competitionID,
		UserID:        userID,
		CheckinID:     checkinID,
	}

	if err := s.entryRepo.Create(ctx, entry); err != nil {
		return nil, fmt.Errorf("create entry failed: %w", err)
	}

	return entry, nil
}

// ListEntries 参赛作品列表
func (s *CompetitionService) ListEntries(ctx context.Context, competitionID primitive.ObjectID, page, pageSize int) ([]*model.CompetitionEntry, int64, error) {
	return s.entryRepo.ListByCompetition(ctx, competitionID, page, pageSize)
}

// ScoreEntry 评委打分
func (s *CompetitionService) ScoreEntry(ctx context.Context, entryID, judgeID primitive.ObjectID, score float64) error {
	if score < 0 || score > 100 {
		return ErrInvalidScore
	}
	return s.entryRepo.Score(ctx, entryID, judgeID, score)
}

// GetRanking 赛事排行榜
type RankingEntry struct {
	Rank   int              `json:"rank"`
	User   *model.User      `json:"user"`
	Score  float64          `json:"score"`
	Entry  *model.CompetitionEntry `json:"entry"`
}

func (s *CompetitionService) GetRanking(ctx context.Context, competitionID primitive.ObjectID) ([]RankingEntry, error) {
	entries, err := s.entryRepo.GetRanking(ctx, competitionID, 100)
	if err != nil {
		return nil, err
	}

	ranking := make([]RankingEntry, 0, len(entries))
	for i, entry := range entries {
		user, _ := s.userRepo.FindByID(ctx, entry.UserID)
		ranking = append(ranking, RankingEntry{
			Rank:  i + 1,
			User:  user,
			Score: entry.Score,
			Entry: entry,
		})
	}

	return ranking, nil
}

// 错误定义
var (
	ErrCompetitionNotFound = &competitionError{"competition not found"}
	ErrCompetitionNotActive = &competitionError{"competition is not active"}
	ErrAlreadySubmitted    = &competitionError{"already submitted to this competition"}
	ErrInvalidScore        = &competitionError{"score must be between 0 and 100"}
	ErrCheckinNotFound     = &competitionError{"checkin not found"}
	ErrNotCheckinOwner     = &competitionError{"not checkin owner"}
)

type competitionError struct {
	msg string
}

func (e *competitionError) Error() string { return e.msg }
