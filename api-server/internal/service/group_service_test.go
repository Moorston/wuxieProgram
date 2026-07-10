package service

import (
	"context"
	"testing"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Mock GroupRepo ---

type mockGroupRepo struct {
	findAllFn func(ctx context.Context) ([]*model.Group, error)
	findByIDFn func(ctx context.Context, id primitive.ObjectID) (*model.Group, error)
}

func (m *mockGroupRepo) FindAll(ctx context.Context) ([]*model.Group, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return nil, nil
}

func (m *mockGroupRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Group, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

// --- Tests ---

func TestGroupService_List(t *testing.T) {
	leaderID := primitive.NewObjectID()
	memberID := primitive.NewObjectID()

	mockGroup := &mockGroupRepo{
		findAllFn: func(ctx context.Context) ([]*model.Group, error) {
			return []*model.Group{
				{
					ID:        primitive.NewObjectID(),
					Name:      "少林组",
					LeaderID:  leaderID,
					MemberIDs: []primitive.ObjectID{leaderID, memberID},
				},
			}, nil
		},
	}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{
				{ID: leaderID, Nickname: "Leader"},
				{ID: memberID, Nickname: "Member"},
			}, nil
		},
	}
	svc := NewGroupService(mockGroup, mockUser)

	groups, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, "少林组", groups[0].Name)
	assert.Len(t, groups[0].Members, 2)
	assert.NotNil(t, groups[0].Leader)
	assert.Equal(t, "Leader", groups[0].Leader.Nickname)
}

func TestGroupService_List_NoUsers(t *testing.T) {
	mockGroup := &mockGroupRepo{
		findAllFn: func(ctx context.Context) ([]*model.Group, error) {
			return []*model.Group{
				{ID: primitive.NewObjectID(), Name: "空组"},
			}, nil
		},
	}
	svc := NewGroupService(mockGroup, &mockUserRepo{})

	groups, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Empty(t, groups[0].Members)
}

func TestGroupService_GetDetail(t *testing.T) {
	groupID := primitive.NewObjectID()
	leaderID := primitive.NewObjectID()
	memberID := primitive.NewObjectID()

	mockGroup := &mockGroupRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Group, error) {
			assert.Equal(t, groupID, id)
			return &model.Group{
				ID:        groupID,
				Name:      "武当组",
				LeaderID:  leaderID,
				MemberIDs: []primitive.ObjectID{leaderID, memberID},
			}, nil
		},
	}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{
				{ID: leaderID, Nickname: "Master"},
				{ID: memberID, Nickname: "Student"},
			}, nil
		},
	}
	svc := NewGroupService(mockGroup, mockUser)

	group, err := svc.GetDetail(context.Background(), groupID)
	require.NoError(t, err)
	require.NotNil(t, group)
	assert.Equal(t, "武当组", group.Name)
	assert.Len(t, group.Members, 2)
	assert.NotNil(t, group.Leader)
	assert.Equal(t, "Master", group.Leader.Nickname)
}

func TestGroupService_GetDetail_NotFound(t *testing.T) {
	mockGroup := &mockGroupRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Group, error) {
			return nil, assert.AnError
		},
	}
	svc := NewGroupService(mockGroup, &mockUserRepo{})

	_, err := svc.GetDetail(context.Background(), primitive.NewObjectID())
	assert.Error(t, err)
}

var _ repository.GroupRepoInterface = (*mockGroupRepo)(nil)