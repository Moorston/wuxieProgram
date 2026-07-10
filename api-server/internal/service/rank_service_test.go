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

// --- Mock RankRepo ---

type mockRankRepo struct {
	getRankListFn func(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error)
	getUserRankFn func(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error)
}

func (m *mockRankRepo) GetRankList(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error) {
	if m.getRankListFn != nil {
		return m.getRankListFn(ctx, period, page, pageSize)
	}
	return nil, nil
}

func (m *mockRankRepo) GetUserRank(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
	if m.getUserRankFn != nil {
		return m.getUserRankFn(ctx, userID, period)
	}
	return nil, nil
}

func (m *mockRankRepo) RefreshRank(ctx context.Context, period model.RankPeriod, entries []*model.RankEntry) error {
	return nil
}

// --- Tests ---

func TestRankService_GetRankList(t *testing.T) {
	userID := primitive.NewObjectID()

	mockRank := &mockRankRepo{
		getRankListFn: func(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error) {
			assert.Equal(t, model.RankPeriodWeek, period)
			assert.Equal(t, 1, page)
			assert.Equal(t, 10, pageSize)
			return []*model.RankEntry{
				{UserID: userID, Score: 100, Rank: 1, Period: model.RankPeriodWeek},
			}, nil
		},
	}
	svc := NewRankService(mockRank)

	entries, err := svc.GetRankList(context.Background(), model.RankPeriodWeek, 1, 10)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, 100, entries[0].Score)
	assert.Equal(t, 1, entries[0].Rank)
}

func TestRankService_GetRankList_Empty(t *testing.T) {
	mockRank := &mockRankRepo{
		getRankListFn: func(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error) {
			return []*model.RankEntry{}, nil
		},
	}
	svc := NewRankService(mockRank)

	entries, err := svc.GetRankList(context.Background(), model.RankPeriodDay, 1, 20)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestRankService_GetUserRank(t *testing.T) {
	userID := primitive.NewObjectID()

	mockRank := &mockRankRepo{
		getUserRankFn: func(ctx context.Context, uid primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, model.RankPeriodAll, period)
			return &model.RankEntry{
				UserID: userID,
				Score:  500,
				Rank:   3,
				Period: model.RankPeriodAll,
			}, nil
		},
	}
	svc := NewRankService(mockRank)

	entry, err := svc.GetUserRank(context.Background(), userID, model.RankPeriodAll)
	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, 500, entry.Score)
	assert.Equal(t, 3, entry.Rank)
}

func TestRankService_GetUserRank_NotFound(t *testing.T) {
	mockRank := &mockRankRepo{
		getUserRankFn: func(ctx context.Context, uid primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
			return nil, nil
		},
	}
	svc := NewRankService(mockRank)

	entry, err := svc.GetUserRank(context.Background(), primitive.NewObjectID(), model.RankPeriodDay)
	require.NoError(t, err)
	assert.Nil(t, entry)
}

func TestRankService_GetUserRank_Error(t *testing.T) {
	mockRank := &mockRankRepo{
		getUserRankFn: func(ctx context.Context, uid primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
			return nil, assert.AnError
		},
	}
	svc := NewRankService(mockRank)

	_, err := svc.GetUserRank(context.Background(), primitive.NewObjectID(), model.RankPeriodDay)
	assert.Error(t, err)
}

var _ repository.RankRepoInterface = (*mockRankRepo)(nil)