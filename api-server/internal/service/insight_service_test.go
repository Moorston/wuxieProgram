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

// --- Mock InsightRepo ---

type mockInsightRepo struct {
	createFn                  func(ctx context.Context, insight *model.Insight) error
	findByIDFn                func(ctx context.Context, id primitive.ObjectID) (*model.Insight, error)
	updateFn                  func(ctx context.Context, id primitive.ObjectID, update bson.M) error
	deleteFn                  func(ctx context.Context, id, userID primitive.ObjectID) error
	listByUserFn              func(ctx context.Context, userID primitive.ObjectID, tag, mood string, page, pageSize int) ([]*model.Insight, int64, error)
	listPublicFn              func(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error)
	onThisDayFn               func(ctx context.Context, userID primitive.ObjectID, month, day int) ([]*model.Insight, error)
	moodStatsFn               func(ctx context.Context, userID primitive.ObjectID, days int) (map[string]int, error)
	incrLikeCountFn           func(ctx context.Context, id primitive.ObjectID, delta int) error
	incrLikeCountWithSessionFn func(sessCtx context.Context, id primitive.ObjectID, delta int) error
	startSessionFn            func() (mongo.Session, error)
}

func (m *mockInsightRepo) Create(ctx context.Context, insight *model.Insight) error {
	if m.createFn != nil {
		return m.createFn(ctx, insight)
	}
	return nil
}

func (m *mockInsightRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockInsightRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, update)
	}
	return nil
}

func (m *mockInsightRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}

func (m *mockInsightRepo) ListByUser(ctx context.Context, userID primitive.ObjectID, tag, mood string, page, pageSize int) ([]*model.Insight, int64, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID, tag, mood, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockInsightRepo) ListPublic(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error) {
	if m.listPublicFn != nil {
		return m.listPublicFn(ctx, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockInsightRepo) OnThisDay(ctx context.Context, userID primitive.ObjectID, month, day int) ([]*model.Insight, error) {
	if m.onThisDayFn != nil {
		return m.onThisDayFn(ctx, userID, month, day)
	}
	return nil, nil
}

func (m *mockInsightRepo) MoodStats(ctx context.Context, userID primitive.ObjectID, days int) (map[string]int, error) {
	if m.moodStatsFn != nil {
		return m.moodStatsFn(ctx, userID, days)
	}
	return nil, nil
}

func (m *mockInsightRepo) IncrLikeCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	if m.incrLikeCountFn != nil {
		return m.incrLikeCountFn(ctx, id, delta)
	}
	return nil
}

func (m *mockInsightRepo) IncrLikeCountWithSession(sessCtx context.Context, id primitive.ObjectID, delta int) error {
	if m.incrLikeCountWithSessionFn != nil {
		return m.incrLikeCountWithSessionFn(sessCtx, id, delta)
	}
	return nil
}

func (m *mockInsightRepo) StartSession() (mongo.Session, error) {
	if m.startSessionFn != nil {
		return m.startSessionFn()
	}
	return &mockSession{}, nil
}

// --- Mock InsightTagRepo ---

type mockInsightTagRepo struct {
	upsertTagsFn func(ctx context.Context, userID primitive.ObjectID, tags []string) error
	decrTagsFn   func(ctx context.Context, userID primitive.ObjectID, tags []string) error
	listByUserFn func(ctx context.Context, userID primitive.ObjectID) ([]*model.InsightTag, error)
}

func (m *mockInsightTagRepo) UpsertTags(ctx context.Context, userID primitive.ObjectID, tags []string) error {
	if m.upsertTagsFn != nil {
		return m.upsertTagsFn(ctx, userID, tags)
	}
	return nil
}

func (m *mockInsightTagRepo) DecrTags(ctx context.Context, userID primitive.ObjectID, tags []string) error {
	if m.decrTagsFn != nil {
		return m.decrTagsFn(ctx, userID, tags)
	}
	return nil
}

func (m *mockInsightTagRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.InsightTag, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID)
	}
	return nil, nil
}

// --- Mock InsightLikeRepo ---

type mockInsightLikeRepo struct {
	toggleFn            func(ctx context.Context, insightID, userID primitive.ObjectID) (bool, error)
	toggleWithSessionFn func(sessCtx context.Context, insightID, userID primitive.ObjectID) (bool, error)
}

func (m *mockInsightLikeRepo) Toggle(ctx context.Context, insightID, userID primitive.ObjectID) (bool, error) {
	if m.toggleFn != nil {
		return m.toggleFn(ctx, insightID, userID)
	}
	return false, nil
}

func (m *mockInsightLikeRepo) ToggleWithSession(sessCtx context.Context, insightID, userID primitive.ObjectID) (bool, error) {
	if m.toggleWithSessionFn != nil {
		return m.toggleWithSessionFn(sessCtx, insightID, userID)
	}
	return false, nil
}

// --- Tests ---

func TestInsightService_Create(t *testing.T) {
	userID := primitive.NewObjectID()
	insightID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		createFn: func(ctx context.Context, insight *model.Insight) error {
			insight.ID = insightID
			assert.Equal(t, userID, insight.UserID)
			assert.Equal(t, model.VisibilityPrivate, insight.Visibility)
			return nil
		},
	}
	mockTag := &mockInsightTagRepo{}
	mockLike := &mockInsightLikeRepo{}
	mockUser := &mockUserRepo{}
	svc := NewInsightService(mockInsight, mockTag, mockLike, mockUser)

	insight := &model.Insight{
		Content: "今日感悟",
		Mood:    model.MoodGood,
	}
	err := svc.Create(context.Background(), userID, insight)
	require.NoError(t, err)
	assert.NotEmpty(t, insight.ID)
	assert.Equal(t, userID, insight.UserID)
	assert.Equal(t, []string{}, insight.Images)
	assert.Equal(t, []string{}, insight.Tags)
}

func TestInsightService_Create_WithTags(t *testing.T) {
	userID := primitive.NewObjectID()

	var tagsCaptured []string
	mockInsight := &mockInsightRepo{
		createFn: func(ctx context.Context, insight *model.Insight) error {
			insight.ID = primitive.NewObjectID()
			return nil
		},
	}
	mockTag := &mockInsightTagRepo{
		upsertTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, userID, uid)
			tagsCaptured = tags
			return nil
		},
	}
	mockLike := &mockInsightLikeRepo{}
	mockUser := &mockUserRepo{}
	svc := NewInsightService(mockInsight, mockTag, mockLike, mockUser)

	insight := &model.Insight{
		Content: "带标签的感悟",
		Tags:    []string{"功夫", "太极"},
	}
	err := svc.Create(context.Background(), userID, insight)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"功夫", "太极"}, tagsCaptured)
}

func TestInsightService_Create_TagError(t *testing.T) {
	userID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		createFn: func(ctx context.Context, insight *model.Insight) error {
			insight.ID = primitive.NewObjectID()
			return nil
		},
	}
	mockTag := &mockInsightTagRepo{
		upsertTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			return fmt.Errorf("tag error")
		},
	}
	svc := NewInsightService(mockInsight, mockTag, &mockInsightLikeRepo{}, &mockUserRepo{})

	// 标签错误仅记录日志，不阻止主流程
	err := svc.Create(context.Background(), userID, &model.Insight{
		Content: "test",
		Tags:    []string{"tag1"},
	})
	require.NoError(t, err)
}

func TestInsightService_GetByID(t *testing.T) {
	insightID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
			return &model.Insight{
				ID:      insightID,
				Content: "test insight",
				UserID:  primitive.NewObjectID(),
			}, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, &mockInsightLikeRepo{}, &mockUserRepo{})

	insight, err := svc.GetByID(context.Background(), insightID)
	require.NoError(t, err)
	require.NotNil(t, insight)
	assert.Equal(t, insightID, insight.ID)
}

func TestInsightService_List(t *testing.T) {
	userID := primitive.NewObjectID()
	insightID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		listByUserFn: func(ctx context.Context, uid primitive.ObjectID, tag, mood string, page, pageSize int) ([]*model.Insight, int64, error) {
			assert.Equal(t, userID, uid)
			return []*model.Insight{
				{ID: insightID, UserID: userID, Content: "test"},
			}, 1, nil
		},
	}
	mockUser := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{{ID: userID, Nickname: "InsightUser"}}, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, &mockInsightLikeRepo{}, mockUser)

	insights, total, err := svc.List(context.Background(), userID, "", "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, insights, 1)
	assert.NotNil(t, insights[0].User)
}

func TestInsightService_Update(t *testing.T) {
	userID := primitive.NewObjectID()
	insightID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
			return &model.Insight{
				ID:      insightID,
				UserID:  userID,
				Content: "old content",
				Tags:    []string{"old"},
			}, nil
		},
		updateFn: func(ctx context.Context, id primitive.ObjectID, update bson.M) error {
			assert.Equal(t, insightID, id)
			return nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{
		decrTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, []string{"old"}, tags)
			return nil
		},
		upsertTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, []string{"new"}, tags)
			return nil
		},
	}, &mockInsightLikeRepo{}, &mockUserRepo{})

	err := svc.Update(context.Background(), insightID, userID, bson.M{"content": "new content", "tags": []string{"new"}})
	require.NoError(t, err)
}

func TestInsightService_Update_AccessDenied(t *testing.T) {
	userID := primitive.NewObjectID()
	otherUserID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
			return &model.Insight{
				ID:     id,
				UserID: otherUserID,
			}, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, &mockInsightLikeRepo{}, &mockUserRepo{})

	err := svc.Update(context.Background(), primitive.NewObjectID(), userID, bson.M{"content": "new"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestInsightService_Delete(t *testing.T) {
	userID := primitive.NewObjectID()
	insightID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
			return &model.Insight{ID: id, UserID: userID, Tags: []string{"tag1"}}, nil
		},
		deleteFn: func(ctx context.Context, id, uid primitive.ObjectID) error {
			assert.Equal(t, insightID, id)
			assert.Equal(t, userID, uid)
			return nil
		},
	}
	mockTag := &mockInsightTagRepo{
		decrTagsFn: func(ctx context.Context, uid primitive.ObjectID, tags []string) error {
			assert.Equal(t, []string{"tag1"}, tags)
			return nil
		},
	}
	svc := NewInsightService(mockInsight, mockTag, &mockInsightLikeRepo{}, &mockUserRepo{})

	err := svc.Delete(context.Background(), insightID, userID)
	require.NoError(t, err)
}

func TestInsightService_Like(t *testing.T) {
	insightID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		incrLikeCountWithSessionFn: func(sessCtx context.Context, id primitive.ObjectID, delta int) error {
			assert.Equal(t, 1, delta)
			return nil
		},
	}
	mockLike := &mockInsightLikeRepo{
		toggleWithSessionFn: func(sessCtx context.Context, iid, uid primitive.ObjectID) (bool, error) {
			return true, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, mockLike, &mockUserRepo{})

	liked, err := svc.Like(context.Background(), insightID, userID)
	require.NoError(t, err)
	assert.True(t, liked)
}

func TestInsightService_Like_Unlike(t *testing.T) {
	insightID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		incrLikeCountWithSessionFn: func(sessCtx context.Context, id primitive.ObjectID, delta int) error {
			assert.Equal(t, -1, delta)
			return nil
		},
	}
	mockLike := &mockInsightLikeRepo{
		toggleWithSessionFn: func(sessCtx context.Context, iid, uid primitive.ObjectID) (bool, error) {
			return false, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, mockLike, &mockUserRepo{})

	liked, err := svc.Like(context.Background(), insightID, userID)
	require.NoError(t, err)
	assert.False(t, liked)
}

func TestInsightService_MoodStats(t *testing.T) {
	userID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		moodStatsFn: func(ctx context.Context, uid primitive.ObjectID, days int) (map[string]int, error) {
			return map[string]int{"good": 3, "normal": 1}, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, &mockInsightLikeRepo{}, &mockUserRepo{})

	stats, err := svc.MoodStats(context.Background(), userID, 7)
	require.NoError(t, err)
	assert.Equal(t, 3, stats["good"])
	assert.Equal(t, 1, stats["normal"])
}

func TestInsightService_OnThisDay(t *testing.T) {
	userID := primitive.NewObjectID()

	mockInsight := &mockInsightRepo{
		onThisDayFn: func(ctx context.Context, uid primitive.ObjectID, month, day int) ([]*model.Insight, error) {
			return []*model.Insight{
				{Content: "去年的感悟"},
			}, nil
		},
	}
	svc := NewInsightService(mockInsight, &mockInsightTagRepo{}, &mockInsightLikeRepo{}, &mockUserRepo{})

	insights, err := svc.OnThisDay(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, insights, 1)
	assert.Equal(t, "去年的感悟", insights[0].Content)
}

var _ repository.InsightRepoInterface = (*mockInsightRepo)(nil)
var _ repository.InsightTagRepoInterface = (*mockInsightTagRepo)(nil)
var _ repository.InsightLikeRepoInterface = (*mockInsightLikeRepo)(nil)