package service

import (
	"context"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroupService struct {
	groupRepo repository.GroupRepoInterface
	userRepo  repository.UserRepoInterface
}

func NewGroupService(groupRepo repository.GroupRepoInterface, userRepo repository.UserRepoInterface) *GroupService {
	return &GroupService{groupRepo: groupRepo, userRepo: userRepo}
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

	users, _ := s.userRepo.FindByIDs(ctx, group.MemberIDs)
	group.Members = users
	for _, u := range users {
		if u.ID == group.LeaderID {
			group.Leader = u
			break
		}
	}

	return group, nil
}