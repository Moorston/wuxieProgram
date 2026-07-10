package service

import (
	"context"
	"fmt"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SocialService struct {
	commentRepo  repository.CommentRepoInterface
	likeRepo     repository.LikeRepoInterface
	checkinRepo  repository.CheckinRepoInterface
	userRepo     repository.UserRepoInterface
	notifService *NotificationService
}

func NewSocialService(commentRepo repository.CommentRepoInterface, likeRepo repository.LikeRepoInterface, checkinRepo repository.CheckinRepoInterface, userRepo repository.UserRepoInterface, notifService *NotificationService) *SocialService {
	return &SocialService{commentRepo: commentRepo, likeRepo: likeRepo, checkinRepo: checkinRepo, userRepo: userRepo, notifService: notifService}
}

func (s *SocialService) ToggleLike(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
	session, err := s.likeRepo.StartSession()
	if err != nil {
		return false, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	var liked bool
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		var txErr error
		liked, txErr = s.likeRepo.ToggleWithSession(sessCtx, checkinID, userID)
		if txErr != nil {
			return nil, txErr
		}

		delta := -1
		if liked {
			delta = 1
		}
		if txErr = s.checkinRepo.IncrLikeCountWithSession(sessCtx, checkinID, delta); txErr != nil {
			return nil, txErr
		}

		return nil, nil
	})

	if err != nil {
		return false, fmt.Errorf("transaction failed: %w", err)
	}

	if liked && s.notifService != nil {
		checkin, err := s.checkinRepo.FindByID(ctx, checkinID)
		if err == nil && checkin.UserID != userID {
			sender, _ := s.userRepo.FindByID(ctx, userID)
			senderName := "有人"
			if sender != nil {
				senderName = sender.Nickname
			}
			s.notifService.Send(ctx, checkin.UserID, userID, model.NotifTypeLike,
				senderName+" 赞了你的打卡",
				"", "checkin", checkinID)
		}
	}

	return liked, nil
}

func (s *SocialService) AddComment(ctx context.Context, checkinID, userID primitive.ObjectID, content string, parentID ...primitive.ObjectID) (*model.Comment, error) {
	comment := &model.Comment{
		CheckinID: checkinID,
		UserID:    userID,
		Content:   content,
	}
	if len(parentID) > 0 {
		comment.ParentID = parentID[0]
	}

	session, err := s.commentRepo.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		if err := s.commentRepo.Create(sessCtx, comment); err != nil {
			return nil, err
		}
		if err := s.checkinRepo.IncrCommentCountWithSession(sessCtx, checkinID); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, fmt.Errorf("add comment transaction failed: %w", err)
	}

	user, _ := s.userRepo.FindByID(ctx, userID)
	comment.User = user

	if s.notifService != nil {
		checkin, err := s.checkinRepo.FindByID(ctx, checkinID)
		if err == nil && checkin.UserID != userID {
			senderName := "有人"
			if user != nil {
				senderName = user.Nickname
			}
			s.notifService.Send(ctx, checkin.UserID, userID, model.NotifTypeComment,
				senderName+" 评论了你的打卡",
				content, "checkin", checkinID)
		}
	}

	return comment, nil
}

func (s *SocialService) GetComments(ctx context.Context, checkinID primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error) {
	comments, total, err := s.commentRepo.ListByCheckin(ctx, checkinID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	userIDs := make([]primitive.ObjectID, 0, len(comments))
	for _, c := range comments {
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

	for _, c := range comments {
		c.User = userMap[c.UserID]
	}

	return comments, total, nil
}

func (s *SocialService) BatchIsLiked(ctx context.Context, checkinIDs []primitive.ObjectID, userID primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	return s.likeRepo.BatchIsLiked(ctx, checkinIDs, userID)
}