package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CheckinService struct {
	checkinRepo repository.CheckinRepoInterface
	userRepo    repository.UserRepoInterface
	mediaURL    string
}

func NewCheckinService(checkinRepo repository.CheckinRepoInterface, userRepo repository.UserRepoInterface, mediaURL string) *CheckinService {
	return &CheckinService{checkinRepo: checkinRepo, userRepo: userRepo, mediaURL: mediaURL}
}

func (s *CheckinService) Prepare(ctx context.Context, userID primitive.ObjectID, description string) (*model.Checkin, error) {
	checkin := &model.Checkin{
		UserID:      userID,
		Description: description,
		Status:      model.CheckinStatusPending,
		Score:       10,
	}
	if err := s.checkinRepo.Create(ctx, checkin); err != nil {
		return nil, err
	}
	return checkin, nil
}

func (s *CheckinService) Callback(ctx context.Context, checkinID primitive.ObjectID, videoURL, coverURL string, duration float64) error {
	return s.checkinRepo.UpdateStatus(ctx, checkinID, model.CheckinStatusDone, videoURL, coverURL, duration)
}

func (s *CheckinService) GetList(ctx context.Context, userID primitive.ObjectID, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	var groupUserIDs []primitive.ObjectID
	if groupID != nil {
		users, err := s.userRepo.FindByGroupID(ctx, *groupID)
		if err != nil {
			return nil, 0, err
		}
		for _, u := range users {
			groupUserIDs = append(groupUserIDs, u.ID)
		}
		if len(groupUserIDs) == 0 {
			return nil, 0, nil
		}
	}

	checkins, total, err := s.checkinRepo.List(ctx, userID, groupUserIDs, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(checkins))
	for _, c := range checkins {
		userIDs = append(userIDs, c.UserID)
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	userMap := make(map[primitive.ObjectID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, c := range checkins {
		c.User = userMap[c.UserID]
	}

	return checkins, total, nil
}

func (s *CheckinService) GetMine(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	return s.checkinRepo.ListByUser(ctx, userID, page, pageSize)
}

func (s *CheckinService) Delete(ctx context.Context, checkinID, userID primitive.ObjectID) error {
	return s.checkinRepo.Delete(ctx, checkinID, userID)
}

func (s *CheckinService) GetByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
	checkin, err := s.checkinRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(ctx, checkin.UserID)
	if err == nil {
		checkin.User = user
	}

	return checkin, nil
}

func (s *CheckinService) Search(ctx context.Context, userID primitive.ObjectID, keyword string, page, pageSize int) ([]*model.Checkin, int64, error) {
	checkins, total, err := s.checkinRepo.Search(ctx, keyword, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(checkins))
	for _, c := range checkins {
		userIDs = append(userIDs, c.UserID)
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}

	userMap := make(map[primitive.ObjectID]*model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	for _, c := range checkins {
		c.User = userMap[c.UserID]
	}

	return checkins, total, nil
}