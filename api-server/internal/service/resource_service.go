package service

import (
	"context"
	"fmt"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const defaultQuota = 5 * 1024 * 1024 * 1024 // 5GB

type ResourceService struct {
	resourceRepo *repository.ResourceRepo
	tagRepo      *repository.ResourceTagRepo
	userRepo     *repository.UserRepo
}

func NewResourceService(resourceRepo *repository.ResourceRepo, tagRepo *repository.ResourceTagRepo, userRepo *repository.UserRepo) *ResourceService {
	return &ResourceService{resourceRepo: resourceRepo, tagRepo: tagRepo, userRepo: userRepo}
}

func (s *ResourceService) Create(ctx context.Context, userID primitive.ObjectID, res *model.Resource) error {
	res.UserID = userID
	if res.Tags == nil {
		res.Tags = []string{}
	}
	if res.ShareScope == "" {
		res.ShareScope = model.ShareScopePrivate
	}

	stats, _ := s.resourceRepo.GetUserStats(ctx, userID)
	if stats != nil && stats.TotalSize+res.FileSize > stats.Quota {
		return fmt.Errorf("存储空间不足，已用 %.1f%%", stats.UsagePercent)
	}

	if err := s.resourceRepo.Create(ctx, res); err != nil {
		return err
	}

	if len(res.Tags) > 0 {
		s.tagRepo.UpsertTags(ctx, userID, res.Tags)
	}

	return nil
}

func (s *ResourceService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
	return s.resourceRepo.FindByID(ctx, id)
}

func (s *ResourceService) Update(ctx context.Context, id, userID primitive.ObjectID, update map[string]interface{}) error {
	res, err := s.resourceRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if tags, ok := update["tags"].([]string); ok {
		s.tagRepo.DecrTags(ctx, userID, res.Tags)
		s.tagRepo.UpsertTags(ctx, userID, tags)
	}

	return s.resourceRepo.Update(ctx, id, update)
}

func (s *ResourceService) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	res, err := s.resourceRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if len(res.Tags) > 0 {
		s.tagRepo.DecrTags(ctx, userID, res.Tags)
	}

	return s.resourceRepo.Delete(ctx, id)
}

func (s *ResourceService) List(ctx context.Context, userID primitive.ObjectID, resType, category, difficulty, tag, keyword, shareScope, sortBy string, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
	resources, total, err := s.resourceRepo.List(ctx, userID, resType, category, difficulty, tag, keyword, shareScope, sortBy, groupID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	s.fillUsers(ctx, resources)
	return resources, total, nil
}

func (s *ResourceService) ListFavorites(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
	resources, total, err := s.resourceRepo.ListFavorites(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	s.fillUsers(ctx, resources)
	return resources, total, nil
}

func (s *ResourceService) ToggleFavorite(ctx context.Context, id primitive.ObjectID) (bool, error) {
	return s.resourceRepo.ToggleFavorite(ctx, id)
}

func (s *ResourceService) GetStats(ctx context.Context, userID primitive.ObjectID) (*model.ResourceStats, error) {
	return s.resourceRepo.GetUserStats(ctx, userID)
}

func (s *ResourceService) GetTags(ctx context.Context, userID primitive.ObjectID) ([]*model.ResourceTag, error) {
	return s.tagRepo.ListByUser(ctx, userID)
}

func (s *ResourceService) IncrViewCount(ctx context.Context, id primitive.ObjectID) error {
	return s.resourceRepo.IncrViewCount(ctx, id)
}

func (s *ResourceService) IncrDownloadCount(ctx context.Context, id primitive.ObjectID) error {
	return s.resourceRepo.IncrDownloadCount(ctx, id)
}

func (s *ResourceService) fillUsers(ctx context.Context, resources []*model.Resource) {
	userIDs := make([]primitive.ObjectID, 0)
	for _, r := range resources {
		if !r.UserID.IsZero() {
			userIDs = append(userIDs, r.UserID)
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
	for _, r := range resources {
		r.User = userMap[r.UserID]
	}
}

func (s *ResourceService) GenerateObjectName(userID primitive.ObjectID, ext string) string {
	return fmt.Sprintf("resource/%s/%s/%s.%s",
		userID.Hex(),
		time.Now().Format("20060102"),
		primitive.NewObjectID().Hex(),
		ext)
}
