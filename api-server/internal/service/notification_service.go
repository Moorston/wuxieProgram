package service

import (
	"context"
	"fmt"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	notifRepo     repository.NotificationRepoInterface
	settingsRepo  repository.NotificationSettingsRepoInterface
	userRepo      repository.UserRepoInterface
}

func NewNotificationService(
	notifRepo repository.NotificationRepoInterface,
	settingsRepo repository.NotificationSettingsRepoInterface,
	userRepo repository.UserRepoInterface,
) *NotificationService {
	return &NotificationService{
		notifRepo:    notifRepo,
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
	}
}

func (s *NotificationService) Send(ctx context.Context, userID, senderID primitive.ObjectID, notifType model.NotificationType, title, content, targetType string, targetID primitive.ObjectID) error {
	settings, _ := s.settingsRepo.GetOrCreate(ctx, userID)

	switch notifType {
	case model.NotifTypeLike:
		if !settings.LikeNotify {
			return nil
		}
	case model.NotifTypeComment, model.NotifTypeCommentReply:
		if !settings.CommentNotify {
			return nil
		}
	case model.NotifTypePlanRemind:
		if !settings.PlanRemind {
			return nil
		}
	case model.NotifTypeGroupNotice:
		if !settings.GroupNotify {
			return nil
		}
	}

	notif := &model.Notification{
		UserID:     userID,
		Type:       notifType,
		Title:      title,
		Content:    content,
		SenderID:   senderID,
		TargetType: targetType,
		TargetID:   targetID,
		IsRead:     false,
	}

	return s.notifRepo.Create(ctx, notif)
}

func (s *NotificationService) SendBatch(ctx context.Context, userIDs []primitive.ObjectID, senderID primitive.ObjectID, notifType model.NotificationType, title, content, targetType string, targetID primitive.ObjectID) error {
	var errs []error
	for _, uid := range userIDs {
		if uid == senderID {
			continue
		}
		if err := s.Send(ctx, uid, senderID, notifType, title, content, targetType, targetID); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("send batch: %d errors, first: %w", len(errs), errs[0])
	}
	return nil
}

func (s *NotificationService) List(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error) {
	notifications, total, err := s.notifRepo.List(ctx, userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	senderIDs := make([]primitive.ObjectID, 0)
	for _, n := range notifications {
		if !n.SenderID.IsZero() {
			senderIDs = append(senderIDs, n.SenderID)
		}
	}

	if len(senderIDs) > 0 {
		users, err := s.userRepo.FindByIDs(ctx, senderIDs)
		if err == nil {
			userMap := make(map[primitive.ObjectID]*model.User)
			for _, u := range users {
				userMap[u.ID] = u
			}
			for _, n := range notifications {
				n.Sender = userMap[n.SenderID]
			}
		}
	}

	return notifications, total, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return s.notifRepo.UnreadCount(ctx, userID)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID primitive.ObjectID) error {
	return s.notifRepo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID primitive.ObjectID) error {
	return s.notifRepo.MarkAllRead(ctx, userID)
}

func (s *NotificationService) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	return s.notifRepo.Delete(ctx, id, userID)
}

func (s *NotificationService) GetSettings(ctx context.Context, userID primitive.ObjectID) (*model.NotificationSettings, error) {
	return s.settingsRepo.GetOrCreate(ctx, userID)
}

func (s *NotificationService) UpdateSettings(ctx context.Context, userID primitive.ObjectID, likeNotify, commentNotify, planRemind, groupNotify *bool, planRemindTime *string) error {
	update := bson.M{}
	if likeNotify != nil {
		update["like_notify"] = *likeNotify
	}
	if commentNotify != nil {
		update["comment_notify"] = *commentNotify
	}
	if planRemind != nil {
		update["plan_remind"] = *planRemind
	}
	if groupNotify != nil {
		update["group_notify"] = *groupNotify
	}
	if planRemindTime != nil {
		update["plan_remind_time"] = *planRemindTime
	}

	if len(update) == 0 {
		return nil
	}

	return s.settingsRepo.Update(ctx, userID, update)
}
