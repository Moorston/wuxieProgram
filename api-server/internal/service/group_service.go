package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
)

type GroupService struct {
	groupRepo repository.GroupRepoInterface
	userRepo  repository.UserRepoInterface
	logger    *zap.Logger
}

func NewGroupService(groupRepo repository.GroupRepoInterface, userRepo repository.UserRepoInterface, logger *zap.Logger) *GroupService {
	return &GroupService{groupRepo: groupRepo, userRepo: userRepo, logger: logger}
}

func (s *GroupService) List(ctx context.Context) ([]*model.Group, error) {
	groups, err := s.groupRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	allMemberIDs := make(map[primitive.ObjectID]struct{})
	for _, g := range groups {
		for _, mid := range g.MemberIDs {
			allMemberIDs[mid] = struct{}{}
		}
	}

	uniqueIDs := make([]primitive.ObjectID, 0, len(allMemberIDs))
	for id := range allMemberIDs {
		uniqueIDs = append(uniqueIDs, id)
	}

	userMap := make(map[primitive.ObjectID]*model.User)
	if len(uniqueIDs) > 0 {
		users, err := s.userRepo.FindByIDs(ctx, uniqueIDs)
		if err == nil {
			for _, u := range users {
				userMap[u.ID] = u
			}
		}
	}

	for _, g := range groups {
		g.Members = make([]*model.User, 0, len(g.MemberIDs))
		for _, mid := range g.MemberIDs {
			if u, ok := userMap[mid]; ok {
				g.Members = append(g.Members, u)
				if u.ID == g.LeaderID {
					g.Leader = u
				}
			}
		}
	}

	return groups, nil
}

func (s *GroupService) GetDetail(ctx context.Context, id primitive.ObjectID) (*model.Group, error) {
	group, err := s.groupRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	users, err := s.userRepo.FindByIDs(ctx, group.MemberIDs)
	if err != nil {
		s.logger.Warn("get group detail: load members failed", zap.String("group_id", id.Hex()), zap.Error(err))
		users = []*model.User{}
	}
	group.Members = users
	for _, u := range users {
		if u.ID == group.LeaderID {
			group.Leader = u
			break
		}
	}

	return group, nil
}

// GenerateInviteCode 生成邀请码（组长操作）
func (s *GroupService) GenerateInviteCode(ctx context.Context, groupID, userID primitive.ObjectID) (string, error) {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return "", ErrGroupNotFound
	}

	if !group.IsLeader(userID) {
		return "", ErrNotGroupLeader
	}

	// 生成 8 位随机码
	code, err := generateRandomCode(8)
	if err != nil {
		return "", fmt.Errorf("generate code failed: %w", err)
	}

	if err := s.groupRepo.UpdateInviteCode(ctx, groupID, code); err != nil {
		return "", fmt.Errorf("update invite code failed: %w", err)
	}

	return code, nil
}

// JoinByInviteCode 通过邀请码加入团组
func (s *GroupService) JoinByInviteCode(ctx context.Context, userID primitive.ObjectID, code string) (*model.Group, error) {
	if code == "" {
		return nil, ErrInvalidInviteCode
	}

	group, err := s.groupRepo.FindByInviteCode(ctx, code)
	if err != nil {
		return nil, ErrInvalidInviteCode
	}

	// 检查是否已是成员
	if group.HasMember(userID) {
		return nil, ErrAlreadyMember
	}

	if err := s.groupRepo.AddMember(ctx, group.ID, userID); err != nil {
		return nil, fmt.Errorf("add member failed: %w", err)
	}

	// 重新获取团组（包含新成员）
	updatedGroup, err := s.groupRepo.FindByID(ctx, group.ID)
	if err != nil {
		// 降级：返回旧数据但不报错
		group.MemberIDs = append(group.MemberIDs, userID)
		return group, nil
	}

	return updatedGroup, nil
}

// RemoveMember 移除成员（仅组长可操作）
func (s *GroupService) RemoveMember(ctx context.Context, groupID, operatorID, targetID primitive.ObjectID) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if !group.IsLeader(operatorID) {
		return ErrNotGroupLeader
	}

	if !group.HasMember(targetID) {
		return ErrNotGroupMember
	}

	// 不能移除自己（组长）
	if targetID == operatorID {
		return ErrCannotRemoveLeader
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, targetID); err != nil {
		return err
	}

	// 清除用户的团组关联
	_ = s.userRepo.Update(ctx, targetID, bson.M{"group_id": primitive.NilObjectID})
	return nil
}

// LeaveGroup 成员主动退出团组
func (s *GroupService) LeaveGroup(ctx context.Context, groupID, userID primitive.ObjectID) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if !group.HasMember(userID) {
		return ErrNotGroupMember
	}

	// 组长不能退出（需先转让组长）
	if group.IsLeader(userID) {
		return ErrLeaderCannotLeave
	}

	if err := s.groupRepo.RemoveMember(ctx, groupID, userID); err != nil {
		return err
	}

	// 清除用户的团组关联
	_ = s.userRepo.Update(ctx, userID, bson.M{"group_id": primitive.NilObjectID})
	return nil
}

// SetLeader 转让组长（仅当前组长可操作）
func (s *GroupService) SetLeader(ctx context.Context, groupID, operatorID, newLeaderID primitive.ObjectID) error {
	group, err := s.groupRepo.FindByID(ctx, groupID)
	if err != nil {
		return ErrGroupNotFound
	}

	if !group.IsLeader(operatorID) {
		return ErrNotGroupLeader
	}

	if !group.HasMember(newLeaderID) {
		return ErrNotGroupMember
	}

	// 新旧组长相同，无需操作
	if operatorID == newLeaderID {
		return nil
	}

	return s.groupRepo.SetLeader(ctx, groupID, newLeaderID)
}

// generateRandomCode 生成随机邀请码
func generateRandomCode(length int) (string, error) {
	bytes := make([]byte, length/2+1)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := hex.EncodeToString(bytes)
	return code[:length], nil
}

// 错误定义
var (
	ErrGroupNotFound      = &groupError{"group not found"}
	ErrNotGroupLeader     = &groupError{"not group leader"}
	ErrInvalidInviteCode  = &groupError{"invalid invite code"}
	ErrAlreadyMember      = &groupError{"already a member of this group"}
	ErrNotGroupMember     = &groupError{"not a member of this group"}
	ErrCannotRemoveLeader = &groupError{"cannot remove the group leader"}
	ErrLeaderCannotLeave  = &groupError{"leader cannot leave; transfer leadership first"}
)

type groupError struct {
	msg string
}

func (e *groupError) Error() string { return e.msg }