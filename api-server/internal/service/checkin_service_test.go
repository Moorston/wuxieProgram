package service

import (
	"context"
	"fmt"
	"testing"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- Mock CheckinRepo ---

type mockCheckinRepo struct {
	createFn                    func(ctx context.Context, c *model.Checkin) error
	findByIDFn                  func(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error)
	updateStatusFn              func(ctx context.Context, id primitive.ObjectID, status model.CheckinStatus, videoURL, coverURL string, duration float64) error
	listFn                      func(ctx context.Context, userID primitive.ObjectID, groupUserIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error)
	listByUserFn                func(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error)
	deleteFn                    func(ctx context.Context, id, userID primitive.ObjectID) error
	incrLikeCountFn             func(ctx context.Context, id primitive.ObjectID, delta int) error
	incrCommentCountFn          func(ctx context.Context, id primitive.ObjectID) error
	incrLikeCountWithSessionFn  func(sessCtx context.Context, id primitive.ObjectID, delta int) error
	incrCommentCountWithSessionFn func(sessCtx context.Context, id primitive.ObjectID) error
	searchFn                    func(ctx context.Context, keyword string, page, pageSize int) ([]*model.Checkin, int64, error)
}

func (m *mockCheckinRepo) Create(ctx context.Context, c *model.Checkin) error {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	return nil
}

func (m *mockCheckinRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockCheckinRepo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.CheckinStatus, videoURL, coverURL string, duration float64) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status, videoURL, coverURL, duration)
	}
	return nil
}

func (m *mockCheckinRepo) List(ctx context.Context, userID primitive.ObjectID, groupUserIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, groupUserIDs, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockCheckinRepo) ListByUser(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockCheckinRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}

func (m *mockCheckinRepo) IncrLikeCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	if m.incrLikeCountFn != nil {
		return m.incrLikeCountFn(ctx, id, delta)
	}
	return nil
}

func (m *mockCheckinRepo) IncrCommentCount(ctx context.Context, id primitive.ObjectID) error {
	if m.incrCommentCountFn != nil {
		return m.incrCommentCountFn(ctx, id)
	}
	return nil
}

func (m *mockCheckinRepo) IncrLikeCountWithSession(sessCtx context.Context, id primitive.ObjectID, delta int) error {
	if m.incrLikeCountWithSessionFn != nil {
		return m.incrLikeCountWithSessionFn(sessCtx, id, delta)
	}
	return nil
}

func (m *mockCheckinRepo) IncrCommentCountWithSession(sessCtx context.Context, id primitive.ObjectID) error {
	if m.incrCommentCountWithSessionFn != nil {
		return m.incrCommentCountWithSessionFn(sessCtx, id)
	}
	return nil
}

func (m *mockCheckinRepo) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Checkin, int64, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, keyword, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockCheckinRepo) Aggregate(ctx context.Context, pipeline []bson.M) (mongo.Cursor, error) {
	return nil, nil
}

// ---- Tests ----

func TestCheckinService_Prepare(t *testing.T) {
	userID := primitive.NewObjectID()
	mockCheckin := &mockCheckinRepo{
		createFn: func(ctx context.Context, c *model.Checkin) error {
			c.ID = primitive.NewObjectID()
			assert.Equal(t, userID, c.UserID)
			assert.Equal(t, "今天练功了", c.Description)
			assert.Equal(t, model.CheckinStatusPending, c.Status)
			assert.Equal(t, 10, c.Score)
			return nil
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "http://media.test")

	checkin, err := svc.Prepare(context.Background(), userID, "今天练功了")
	require.NoError(t, err)
	require.NotNil(t, checkin)
	assert.Equal(t, userID, checkin.UserID)
	assert.Equal(t, "今天练功了", checkin.Description)
	assert.Equal(t, model.CheckinStatusPending, checkin.Status)
}

func TestCheckinService_Prepare_CreateError(t *testing.T) {
	mockCheckin := &mockCheckinRepo{
		createFn: func(ctx context.Context, c *model.Checkin) error {
			return fmt.Errorf("db error")
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	_, err := svc.Prepare(context.Background(), primitive.NewObjectID(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
}

func TestCheckinService_Callback(t *testing.T) {
	checkinID := primitive.NewObjectID()
	var capturedStatus model.CheckinStatus
	var capturedVideoURL, capturedCoverURL string
	var capturedDuration float64

	mockCheckin := &mockCheckinRepo{
		updateStatusFn: func(ctx context.Context, id primitive.ObjectID, status model.CheckinStatus, videoURL, coverURL string, duration float64) error {
			assert.Equal(t, checkinID, id)
			capturedStatus = status
			capturedVideoURL = videoURL
			capturedCoverURL = coverURL
			capturedDuration = duration
			return nil
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	err := svc.Callback(context.Background(), checkinID, "http://video.mp4", "http://cover.jpg", 30.5)
	require.NoError(t, err)
	assert.Equal(t, model.CheckinStatusDone, capturedStatus)
	assert.Equal(t, "http://video.mp4", capturedVideoURL)
	assert.Equal(t, "http://cover.jpg", capturedCoverURL)
	assert.Equal(t, 30.5, capturedDuration)
}

func TestCheckinService_GetByID(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	expectedCheckin := &model.Checkin{
		ID:          checkinID,
		UserID:      userID,
		Description: "test checkin",
		Status:      model.CheckinStatusDone,
	}

	mockCheckin := &mockCheckinRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
			assert.Equal(t, checkinID, id)
			return expectedCheckin, nil
		},
	}
	mockUser := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			assert.Equal(t, userID, id)
			return &model.User{ID: userID, Nickname: "TestUser"}, nil
		},
	}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	checkin, err := svc.GetByID(context.Background(), checkinID)
	require.NoError(t, err)
	require.NotNil(t, checkin)
	assert.Equal(t, checkinID, checkin.ID)
	assert.NotNil(t, checkin.User)
	assert.Equal(t, "TestUser", checkin.User.Nickname)
}

func TestCheckinService_GetByID_NotFound(t *testing.T) {
	mockCheckin := &mockCheckinRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	_, err := svc.GetByID(context.Background(), primitive.NewObjectID())
	assert.Error(t, err)
}

func TestCheckinService_GetList(t *testing.T) {
	userID := primitive.NewObjectID()
	groupID := primitive.NewObjectID()
	checkinID := primitive.NewObjectID()

	// Two users in the group
	userA := primitive.NewObjectID()
	userB := primitive.NewObjectID()

	mockCheckin := &mockCheckinRepo{
		listFn: func(ctx context.Context, uid primitive.ObjectID, groupUserIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
			assert.Equal(t, userID, uid)
			assert.ElementsMatch(t, []primitive.ObjectID{userA, userB}, groupUserIDs)
			return []*model.Checkin{
				{ID: checkinID, UserID: userA, Description: "checkin A"},
			}, 1, nil
		},
	}
	mockUser := &mockUserRepo{
		findByGroupID: func(ctx context.Context, gid primitive.ObjectID) ([]*model.User, error) {
			assert.Equal(t, groupID, gid)
			return []*model.User{
				{ID: userA, Nickname: "UserA"},
				{ID: userB, Nickname: "UserB"},
			}, nil
		},
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{
				{ID: userA, Nickname: "UserA"},
			}, nil
		},
	}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	checkins, total, err := svc.GetList(context.Background(), userID, &groupID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, checkins, 1)
	assert.NotNil(t, checkins[0].User)
	assert.Equal(t, "UserA", checkins[0].User.Nickname)
}

func TestCheckinService_GetList_NoGroup(t *testing.T) {
	userID := primitive.NewObjectID()
	checkinID := primitive.NewObjectID()

	mockCheckin := &mockCheckinRepo{
		listFn: func(ctx context.Context, uid primitive.ObjectID, groupUserIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
			assert.Empty(t, groupUserIDs)
			return []*model.Checkin{
				{ID: checkinID, UserID: userID, Description: "my checkin"},
			}, 1, nil
		},
	}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{{ID: userID, Nickname: "Me"}}, nil
		},
	}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	checkins, total, err := svc.GetList(context.Background(), userID, nil, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, checkins, 1)
}

func TestCheckinService_GetList_EmptyGroup(t *testing.T) {
	userID := primitive.NewObjectID()
	groupID := primitive.NewObjectID()

	mockUser := &mockUserRepo{
		findByGroupID: func(ctx context.Context, gid primitive.ObjectID) ([]*model.User, error) {
			return nil, nil
		},
	}
	mockCheckin := &mockCheckinRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	checkins, total, err := svc.GetList(context.Background(), userID, &groupID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, checkins)
}

func TestCheckinService_GetMine(t *testing.T) {
	userID := primitive.NewObjectID()
	checkinID := primitive.NewObjectID()

	mockCheckin := &mockCheckinRepo{
		listByUserFn: func(ctx context.Context, uid primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
			assert.Equal(t, userID, uid)
			return []*model.Checkin{
				{ID: checkinID, UserID: userID, Description: "my checkin"},
			}, 1, nil
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	checkins, total, err := svc.GetMine(context.Background(), userID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, checkins, 1)
}

func TestCheckinService_Delete(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockCheckin := &mockCheckinRepo{
		deleteFn: func(ctx context.Context, id, uid primitive.ObjectID) error {
			assert.Equal(t, checkinID, id)
			assert.Equal(t, userID, uid)
			return nil
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	err := svc.Delete(context.Background(), checkinID, userID)
	require.NoError(t, err)
}

func TestCheckinService_Delete_Error(t *testing.T) {
	mockCheckin := &mockCheckinRepo{
		deleteFn: func(ctx context.Context, id, uid primitive.ObjectID) error {
			return fmt.Errorf("delete failed")
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	err := svc.Delete(context.Background(), primitive.NewObjectID(), primitive.NewObjectID())
	assert.Error(t, err)
}

func TestCheckinService_Search(t *testing.T) {
	userID := primitive.NewObjectID()
	checkinID := primitive.NewObjectID()

	mockCheckin := &mockCheckinRepo{
		searchFn: func(ctx context.Context, keyword string, page, pageSize int) ([]*model.Checkin, int64, error) {
			assert.Equal(t, "keyword", keyword)
			return []*model.Checkin{
				{ID: checkinID, UserID: userID, Description: "found checkin"},
			}, 1, nil
		},
	}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{{ID: userID, Nickname: "FoundUser"}}, nil
		},
	}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	checkins, total, err := svc.Search(context.Background(), userID, "keyword", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, checkins, 1)
	assert.NotNil(t, checkins[0].User)
}

func TestCheckinService_Search_Error(t *testing.T) {
	mockCheckin := &mockCheckinRepo{
		searchFn: func(ctx context.Context, keyword string, page, pageSize int) ([]*model.Checkin, int64, error) {
			return nil, 0, fmt.Errorf("search error")
		},
	}
	mockUser := &mockUserRepo{}
	svc := NewCheckinService(mockCheckin, mockUser, "")

	_, _, err := svc.Search(context.Background(), primitive.NewObjectID(), "keyword", 1, 10)
	assert.Error(t, err)
}

var _ repository.CheckinRepoInterface = (*mockCheckinRepo)(nil)