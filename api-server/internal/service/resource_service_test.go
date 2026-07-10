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
)

// --- Mock ResourceRepo ---

type mockResourceRepo struct {
	createFn           func(ctx context.Context, res *model.Resource) error
	findByIDFn         func(ctx context.Context, id primitive.ObjectID) (*model.Resource, error)
	updateFn           func(ctx context.Context, id primitive.ObjectID, update bson.M) error
	deleteFn           func(ctx context.Context, id primitive.ObjectID) error
	listFn             func(ctx context.Context, userID primitive.ObjectID, resType, category, difficulty, tag, keyword, shareScope, sortBy string, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error)
	toggleFavoriteFn   func(ctx context.Context, id, userID primitive.ObjectID) (bool, error)
	listFavoritesFn    func(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error)
	getUserStatsFn     func(ctx context.Context, userID primitive.ObjectID) (*model.ResourceStats, error)
	incrViewCountFn    func(ctx context.Context, id primitive.ObjectID) error
	incrDownloadCountFn func(ctx context.Context, id primitive.ObjectID) error
}

func (m *mockResourceRepo) Create(ctx context.Context, res *model.Resource) error {
	if m.createFn != nil {
		return m.createFn(ctx, res)
	}
	return nil
}

func (m *mockResourceRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockResourceRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, update)
	}
	return nil
}

func (m *mockResourceRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockResourceRepo) List(ctx context.Context, userID primitive.ObjectID, resType, category, difficulty, tag, keyword, shareScope, sortBy string, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, resType, category, difficulty, tag, keyword, shareScope, sortBy, groupID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockResourceRepo) ToggleFavorite(ctx context.Context, id, userID primitive.ObjectID) (bool, error) {
	if m.toggleFavoriteFn != nil {
		return m.toggleFavoriteFn(ctx, id, userID)
	}
	return false, nil
}

func (m *mockResourceRepo) ListFavorites(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
	if m.listFavoritesFn != nil {
		return m.listFavoritesFn(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockResourceRepo) GetUserStats(ctx context.Context, userID primitive.ObjectID) (*model.ResourceStats, error) {
	if m.getUserStatsFn != nil {
		return m.getUserStatsFn(ctx, userID)
	}
	return &model.ResourceStats{Quota: defaultQuota}, nil
}

func (m *mockResourceRepo) IncrViewCount(ctx context.Context, id primitive.ObjectID) error {
	if m.incrViewCountFn != nil {
		return m.incrViewCountFn(ctx, id)
	}
	return nil
}

func (m *mockResourceRepo) IncrDownloadCount(ctx context.Context, id primitive.ObjectID) error {
	if m.incrDownloadCountFn != nil {
		return m.incrDownloadCountFn(ctx, id)
	}
	return nil
}

// --- Mock ResourceTagRepo ---

type mockResourceTagRepo struct {
	upsertTagsFn func(ctx context.Context, userID primitive.ObjectID, tags []string) error
	decrTagsFn   func(ctx context.Context, userID primitive.ObjectID, tags []string) error
	listByUserFn func(ctx context.Context, userID primitive.ObjectID) ([]*model.ResourceTag, error)
}

func (m *mockResourceTagRepo) UpsertTags(ctx context.Context, userID primitive.ObjectID, tags []string) error {
	if m.upsertTagsFn != nil {
		return m.upsertTagsFn(ctx, userID, tags)
	}
	return nil
}

func (m *mockResourceTagRepo) DecrTags(ctx context.Context, userID primitive.ObjectID, tags []string) error {
	if m.decrTagsFn != nil {
		return m.decrTagsFn(ctx, userID, tags)
	}
	return nil
}

func (m *mockResourceTagRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.ResourceTag, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID)
	}
	return nil, nil
}

// --- Tests ---

func TestResourceService_Create(t *testing.T) {
	userID := primitive.NewObjectID()
	mockResource := &mockResourceRepo{
		getUserStatsFn: func(ctx context.Context, uid primitive.ObjectID) (*model.ResourceStats, error) {
			return &model.ResourceStats{TotalSize: 100, Quota: defaultQuota}, nil
		},
		createFn: func(ctx context.Context, res *model.Resource) error {
			res.ID = primitive.NewObjectID()
			assert.Equal(t, userID, res.UserID)
			assert.Equal(t, model.ShareScopePrivate, res.ShareScope)
			return nil
		},
	}
	mockTag := &mockResourceTagRepo{}
	mockUser := &mockUserRepo{}
	svc := NewResourceService(mockResource, mockTag, mockUser)

	res := &model.Resource{
		Title:   "功夫教学视频",
		FileSize: 1000,
	}
	err := svc.Create(context.Background(), userID, res)
	require.NoError(t, err)
	assert.NotEmpty(t, res.ID)
	assert.Equal(t, userID, res.UserID)
}

func TestResourceService_Create_QuotaExceeded(t *testing.T) {
	userID := primitive.NewObjectID()
	mockResource := &mockResourceRepo{
		getUserStatsFn: func(ctx context.Context, uid primitive.ObjectID) (*model.ResourceStats, error) {
			return &model.ResourceStats{TotalSize: defaultQuota, Quota: defaultQuota}, nil
		},
	}
	mockTag := &mockResourceTagRepo{}
	svc := NewResourceService(mockResource, mockTag, &mockUserRepo{})

	err := svc.Create(context.Background(), userID, &model.Resource{FileSize: 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "存储空间不足")
}

func TestResourceService_Create_WithTags(t *testing.T) {
	userID := primitive.NewObjectID()
	mockResource := &mockResourceRepo{
		getUserStatsFn: func(ctx context.Context, uid primitive.ObjectID) (*model.ResourceStats, error) {
			return &model.ResourceStats{TotalSize: 100, Quota: defaultQuota}, nil
		},
		createFn: func(ctx context.Context, res *model.Resource) error {
			res.ID = primitive.NewObjectID()
			return nil
		},
	}
	var upsertedTags []string
	mockTag := &mockResourceTagRepo{
		upsertTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			upsertedTags = tags
			return nil
		},
	}
	svc := NewResourceService(mockResource, mockTag, &mockUserRepo{})

	err := svc.Create(context.Background(), userID, &model.Resource{
		Title: "test",
		Tags:  []string{"功夫", "太极"},
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"功夫", "太极"}, upsertedTags)
}

func TestResourceService_GetByID(t *testing.T) {
	resourceID := primitive.NewObjectID()
	mockResource := &mockResourceRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
			return &model.Resource{ID: resourceID, Title: "test"}, nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, &mockUserRepo{})

	res, err := svc.GetByID(context.Background(), resourceID)
	require.NoError(t, err)
	assert.Equal(t, resourceID, res.ID)
}

func TestResourceService_Update(t *testing.T) {
	userID := primitive.NewObjectID()
	resourceID := primitive.NewObjectID()

	mockResource := &mockResourceRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
			return &model.Resource{ID: id, UserID: userID, Tags: []string{"old"}}, nil
		},
		updateFn: func(ctx context.Context, id primitive.ObjectID, update bson.M) error {
			assert.Equal(t, resourceID, id)
			return nil
		},
	}
	mockTag := &mockResourceTagRepo{
		decrTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, []string{"old"}, tags)
			return nil
		},
		upsertTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, []string{"new"}, tags)
			return nil
		},
	}
	svc := NewResourceService(mockResource, mockTag, &mockUserRepo{})

	err := svc.Update(context.Background(), resourceID, userID, bson.M{"title": "new title", "tags": []string{"new"}})
	require.NoError(t, err)
}

func TestResourceService_Update_AccessDenied(t *testing.T) {
	userID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()

	mockResource := &mockResourceRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
			return &model.Resource{ID: id, UserID: otherID}, nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, &mockUserRepo{})

	err := svc.Update(context.Background(), primitive.NewObjectID(), userID, bson.M{"title": "new"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestResourceService_Delete(t *testing.T) {
	userID := primitive.NewObjectID()
	resourceID := primitive.NewObjectID()

	mockResource := &mockResourceRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
			return &model.Resource{ID: id, UserID: userID, Tags: []string{"tag1"}}, nil
		},
		deleteFn: func(ctx context.Context, id primitive.ObjectID) error {
			assert.Equal(t, resourceID, id)
			return nil
		},
	}
	mockTag := &mockResourceTagRepo{
		decrTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, []string{"tag1"}, tags)
			return nil
		},
	}
	svc := NewResourceService(mockResource, mockTag, &mockUserRepo{})

	err := svc.Delete(context.Background(), resourceID, userID)
	require.NoError(t, err)
}

func TestResourceService_Delete_AccessDenied(t *testing.T) {
	mockResource := &mockResourceRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
			return &model.Resource{UserID: primitive.NewObjectID()}, nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, &mockUserRepo{})

	err := svc.Delete(context.Background(), primitive.NewObjectID(), primitive.NewObjectID())
	assert.Error(t, err)
}

func TestResourceService_ToggleFavorite(t *testing.T) {
	resourceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockResource := &mockResourceRepo{
		toggleFavoriteFn: func(ctx context.Context, id, uid primitive.ObjectID) (bool, error) {
			assert.Equal(t, resourceID, id)
			assert.Equal(t, userID, uid)
			return true, nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, &mockUserRepo{})

	liked, err := svc.ToggleFavorite(context.Background(), resourceID, userID)
	require.NoError(t, err)
	assert.True(t, liked)
}

func TestResourceService_GetStats(t *testing.T) {
	userID := primitive.NewObjectID()

	mockResource := &mockResourceRepo{
		getUserStatsFn: func(ctx context.Context, uid primitive.ObjectID) (*model.ResourceStats, error) {
			return &model.ResourceStats{TotalSize: 500, TotalCount: 3, Quota: defaultQuota}, nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, &mockUserRepo{})

	stats, err := svc.GetStats(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, int64(500), stats.TotalSize)
	assert.Equal(t, 3, stats.TotalCount)
}

func TestResourceService_List(t *testing.T) {
	userID := primitive.NewObjectID()
	resourceID := primitive.NewObjectID()

	mockResource := &mockResourceRepo{
		listFn: func(ctx context.Context, uid primitive.ObjectID, resType, category, difficulty, tag, keyword, shareScope, sortBy string, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
			return []*model.Resource{
				{ID: resourceID, UserID: userID, Title: "resource"},
			}, 1, nil
		},
	}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{{ID: userID, Nickname: "Owner"}}, nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, mockUser)

	resources, total, err := svc.List(context.Background(), userID, "", "", "", "", "", "", "", nil, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, resources, 1)
	assert.NotNil(t, resources[0].User)
}

func TestResourceService_IncrViewCount(t *testing.T) {
	resourceID := primitive.NewObjectID()
	called := false

	mockResource := &mockResourceRepo{
		incrViewCountFn: func(ctx context.Context, id primitive.ObjectID) error {
			assert.Equal(t, resourceID, id)
			called = true
			return nil
		},
	}
	svc := NewResourceService(mockResource, &mockResourceTagRepo{}, &mockUserRepo{})

	err := svc.IncrViewCount(context.Background(), resourceID)
	require.NoError(t, err)
	assert.True(t, called)
}

var _ repository.ResourceRepoInterface = (*mockResourceRepo)(nil)
var _ repository.ResourceTagRepoInterface = (*mockResourceTagRepo)(nil)