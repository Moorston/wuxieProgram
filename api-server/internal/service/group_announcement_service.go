package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type GroupAnnouncementService struct {
	annRepo    repository.GroupAnnouncementRepoInterface
	groupRepo  repository.GroupRepoInterface
	userRepo   repository.UserRepoInterface
	logger     *zap.Logger
}

func NewGroupAnnouncementService(
	annRepo repository.GroupAnnouncementRepoInterface,
	groupRepo repository.GroupRepoInterface,
	userRepo repository.UserRepoInterface,
	logger *zap.Logger,
) *GroupAnnouncementService {
	return &GroupAnnouncementService{
		annRepo:   annRepo,
		groupRepo: groupRepo,
		userRepo:  userRepo,
		logger:    logger,
	}
}

// Create 创建公告（仅组长可操作）
func (s *GroupAnnouncementService) Create(ctx context.Context, groupID, authorID primitive.ObjectID, title, content string, isPinned bool) (*model.GroupAnnouncement, error) {
	// 验证团组存在
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return nil, ErrGroupNotFound
	}

	// 验证是组长
	if !group.IsLeader(authorID) {
		return nil, ErrNotGroupLeader
	}

	announcement := &model.GroupAnnouncement{
		GroupID:  groupID,
		AuthorID: authorID,
		Title:    title,
		Content:  content,
		IsPinned: isPinned,
	}

	if err := s.annRepo.Create(ctx, announcement); err != nil {
		return nil, err
	}

	return announcement, nil
}

// List 获取团组公告列表
func (s *GroupAnnouncementService) List(ctx context.Context, groupID primitive.ObjectID, page, pageSize int) ([]*model.GroupAnnouncement, int64, error) {
	// 验证团组存在
	if _, err := s.groupRepo.FindByID(ctx, groupID); err != nil {
		return nil, 0, ErrGroupNotFound
	}

	announcements, total, err := s.annRepo.ListByGroup(ctx, groupID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 填充作者信息
	authorIDs := make([]primitive.ObjectID, 0, len(announcements))
	for _, a := range announcements {
		authorIDs = append(authorIDs, a.AuthorID)
	}
	users, err := s.userRepo.FindByIDs(ctx, authorIDs)
	if err != nil {
		s.logger.Warn("list announcements: load authors failed", zap.Error(err))
		users = []*model.User{}
	}
	userMap := make(map[primitive.ObjectID]*model.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}
	for _, a := range announcements {
		a.Author = userMap[a.AuthorID]
	}

	return announcements, total, nil
}

// Delete 删除公告（仅作者/组长可操作）
func (s *GroupAnnouncementService) Delete(ctx context.Context, annID, userID primitive.ObjectID) error {
	ann, err := s.annRepo.FindByID(ctx, annID)
	if err != nil {
		return ErrAnnouncementNotFound
	}

	// 验证是作者或组长
	group, err := s.groupRepo.FindByID(ctx, ann.GroupID)
	if err != nil {
		return ErrGroupNotFound
	}
	if ann.AuthorID != userID && !group.IsLeader(userID) {
		return ErrAnnouncementAccessDenied
	}

	return s.annRepo.Delete(ctx, annID, ann.AuthorID)
}

// 错误定义
var (
	ErrAnnouncementNotFound = &announcementError{"announcement not found"}
	ErrAnnouncementAccessDenied = &announcementError{"access denied: not author or group leader"}
)

type announcementError struct{ msg string }

func (e *announcementError) Error() string { return e.msg }
