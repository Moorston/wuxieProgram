package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type InsightService struct {
	insightRepo  repository.InsightRepoInterface
	tagRepo      repository.InsightTagRepoInterface
	likeRepo     repository.InsightLikeRepoInterface
	userRepo     repository.UserRepoInterface
}

func NewInsightService(insightRepo repository.InsightRepoInterface, tagRepo repository.InsightTagRepoInterface, likeRepo repository.InsightLikeRepoInterface, userRepo repository.UserRepoInterface) *InsightService {
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
		if err := s.tagRepo.UpsertTags(ctx, userID, insight.Tags); err != nil {
			// 标签更新失败不影响主流程，仅记录日志
			log.Printf("[WARN] upsert insight tags failed: %v", err)
		}
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

	// 验证所有权
	if insight.UserID != userID {
		return fmt.Errorf("access denied: not insight owner")
	}

	// 处理JSON反序列化后的tags类型（可能是[]interface{}而非[]string）
	if tags := extractTags(update["tags"]); tags != nil {
		if err := s.tagRepo.DecrTags(ctx, userID, insight.Tags); err != nil {
			log.Printf("[WARN] decr insight tags failed: %v", err)
		}
		if err := s.tagRepo.UpsertTags(ctx, userID, tags); err != nil {
			log.Printf("[WARN] upsert insight tags failed: %v", err)
		}
	}

	return s.insightRepo.Update(ctx, id, update)
}

func (s *InsightService) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	insight, err := s.insightRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if len(insight.Tags) > 0 {
		if err := s.tagRepo.DecrTags(ctx, userID, insight.Tags); err != nil {
			log.Printf("[WARN] decr insight tags failed: %v", err)
		}
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
	// 使用MongoDB事务确保数据一致性
	session, err := s.insightRepo.StartSession()
	if err != nil {
		return false, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	var liked bool
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		var txErr error
		liked, txErr = s.likeRepo.ToggleWithSession(sessCtx, id, userID)
		if txErr != nil {
			return nil, txErr
		}

		delta := -1
		if liked {
			delta = 1
		}
		if txErr = s.insightRepo.IncrLikeCountWithSession(sessCtx, id, delta); txErr != nil {
			return nil, txErr
		}

		return nil, nil
	})

	if err != nil {
		return false, fmt.Errorf("transaction failed: %w", err)
	}

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
