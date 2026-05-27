package service

import (
	"context"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InsightService struct {
	insightRepo  *repository.InsightRepo
	tagRepo      *repository.InsightTagRepo
	likeRepo     *repository.InsightLikeRepo
	userRepo     *repository.UserRepo
}

func NewInsightService(insightRepo *repository.InsightRepo, tagRepo *repository.InsightTagRepo, likeRepo *repository.InsightLikeRepo, userRepo *repository.UserRepo) *InsightService {
	return &InsightService{insightRepo: insightRepo, tagRepo: tagRepo, likeRepo: likeRepo, userRepo: userRepo}
}

func (s *InsightService) Create(ctx context.Context, userID primitive.ObjectID, insight *model.Insight) error {
	insight.UserID = userID
	if insight.Images == nil {
		insight.Images = []string{}
	}
	if insight.Tags == nil {
		insight.Tags = []string{}
	}
	if insight.Visibility == "" {
		insight.Visibility = model.VisibilityPrivate
	}

	if err := s.insightRepo.Create(ctx, insight); err != nil {
		return err
	}

	if len(insight.Tags) > 0 {
		s.tagRepo.UpsertTags(ctx, userID, insight.Tags)
	}

	return nil
}

func (s *InsightService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
	return s.insightRepo.FindByID(ctx, id)
}

func (s *InsightService) Update(ctx context.Context, id, userID primitive.ObjectID, update map[string]interface{}) error {
	insight, err := s.insightRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if tags, ok := update["tags"].([]string); ok {
		s.tagRepo.DecrTags(ctx, userID, insight.Tags)
		s.tagRepo.UpsertTags(ctx, userID, tags)
	}

	return s.insightRepo.Update(ctx, id, update)
}

func (s *InsightService) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	insight, err := s.insightRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if len(insight.Tags) > 0 {
		s.tagRepo.DecrTags(ctx, userID, insight.Tags)
	}

	return s.insightRepo.Delete(ctx, id, userID)
}

func (s *InsightService) List(ctx context.Context, userID primitive.ObjectID, tag, mood string, page, pageSize int) ([]*model.Insight, int64, error) {
	insights, total, err := s.insightRepo.ListByUser(ctx, userID, tag, mood, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	s.fillUsers(ctx, insights)
	return insights, total, nil
}

func (s *InsightService) ListPublic(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error) {
	insights, total, err := s.insightRepo.ListPublic(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	s.fillUsers(ctx, insights)
	return insights, total, nil
}

func (s *InsightService) OnThisDay(ctx context.Context, userID primitive.ObjectID) ([]*model.Insight, error) {
	now := time.Now()
	return s.insightRepo.OnThisDay(ctx, userID, int(now.Month()), now.Day())
}

func (s *InsightService) MoodStats(ctx context.Context, userID primitive.ObjectID, days int) (map[string]int, error) {
	return s.insightRepo.MoodStats(ctx, userID, days)
}

func (s *InsightService) GetTags(ctx context.Context, userID primitive.ObjectID) ([]*model.InsightTag, error) {
	return s.tagRepo.ListByUser(ctx, userID)
}

func (s *InsightService) Like(ctx context.Context, id, userID primitive.ObjectID) (bool, error) {
	liked, err := s.likeRepo.Toggle(ctx, id, userID)
	if err != nil {
		return false, err
	}

	delta := -1
	if liked {
		delta = 1
	}
	s.insightRepo.IncrLikeCount(ctx, id, delta)

	return liked, nil
}

func (s *InsightService) fillUsers(ctx context.Context, insights []*model.Insight) {
	userIDs := make([]primitive.ObjectID, 0)
	for _, ins := range insights {
		if !ins.UserID.IsZero() {
			userIDs = append(userIDs, ins.UserID)
		}
	}
	if len(userIDs) == 0 {
		return
	}
	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return
	}
	userMap := make(map[primitive.ObjectID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}
	for _, ins := range insights {
		ins.User = userMap[ins.UserID]
	}
}
