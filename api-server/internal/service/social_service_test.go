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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- mockSession implements mongo.Session ---

type mockSession struct {
	withTxFn func(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error)) (interface{}, error)
}

func (s *mockSession) Client() *mongo.Client                                        { return nil }
func (s *mockSession) ClusterTime() bson.Raw                                         { return nil }
func (s *mockSession) AdvanceClusterTime(bson.Raw) error                               { return nil }
func (s *mockSession) OperationTime() *primitive.Timestamp                            { return nil }
func (s *mockSession) AdvanceOperationTime(*primitive.Timestamp) error                 { return nil }
func (s *mockSession) KillCursor(context.Context) error                               { return nil }
func (s *mockSession) EndSession(context.Context)                                    {}
func (s *mockSession) StartTransaction(...*options.TransactionOptions) error          { return nil }
func (s *mockSession) AbortTransaction(context.Context) error                         { return nil }
func (s *mockSession) CommitTransaction(context.Context) error                        { return nil }
func (s *mockSession) ID() bson.Raw                                                 { return nil }

func (s *mockSession) WithTransaction(ctx context.Context, fn func(mongo.SessionContext) (interface{}, error), _ ...*options.TransactionOptions) (interface{}, error) {
	if s.withTxFn != nil {
		return s.withTxFn(ctx, fn)
	}
	sessCtx := &mockSessionCtx{Context: ctx, session: s}
	return fn(sessCtx)
}

// mockSessionCtx implements mongo.SessionContext
type mockSessionCtx struct {
	context.Context
	session mongo.Session
}

func (msc *mockSessionCtx) Session() mongo.Session {
	return msc.session
}

// --- Mock CommentRepo ---

type mockCommentRepo struct {
	createFn        func(ctx context.Context, c *model.Comment) error
	listByCheckinFn func(ctx context.Context, checkinID primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error)
	startSessionFn  func() (mongo.Session, error)
}

func (m *mockCommentRepo) Create(ctx context.Context, c *model.Comment) error {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	return nil
}

func (m *mockCommentRepo) ListByCheckin(ctx context.Context, checkinID primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error) {
	if m.listByCheckinFn != nil {
		return m.listByCheckinFn(ctx, checkinID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockCommentRepo) StartSession() (mongo.Session, error) {
	if m.startSessionFn != nil {
		return m.startSessionFn()
	}
	return &mockSession{}, nil
}

// --- Mock LikeRepo ---

type mockLikeRepo struct {
	toggleFn            func(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error)
	toggleWithSessionFn func(sessCtx context.Context, checkinID, userID primitive.ObjectID) (bool, error)
	isLikedFn           func(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error)
	batchIsLikedFn      func(ctx context.Context, checkinIDs []primitive.ObjectID, userID primitive.ObjectID) (map[primitive.ObjectID]bool, error)
	startSessionFn      func() (mongo.Session, error)
}

func (m *mockLikeRepo) Toggle(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
	if m.toggleFn != nil {
		return m.toggleFn(ctx, checkinID, userID)
	}
	return false, nil
}

func (m *mockLikeRepo) ToggleWithSession(sessCtx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
	if m.toggleWithSessionFn != nil {
		return m.toggleWithSessionFn(sessCtx, checkinID, userID)
	}
	return false, nil
}

func (m *mockLikeRepo) IsLiked(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
	if m.isLikedFn != nil {
		return m.isLikedFn(ctx, checkinID, userID)
	}
	return false, nil
}

func (m *mockLikeRepo) BatchIsLiked(ctx context.Context, checkinIDs []primitive.ObjectID, userID primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	if m.batchIsLikedFn != nil {
		return m.batchIsLikedFn(ctx, checkinIDs, userID)
	}
	return nil, nil
}

func (m *mockLikeRepo) StartSession() (mongo.Session, error) {
	if m.startSessionFn != nil {
		return m.startSessionFn()
	}
	return &mockSession{}, nil
}

// --- Mock NotificationRepo (shared with notification_service_test.go) ---

type mockNotificationRepo struct {
	createFn       func(ctx context.Context, n *model.Notification) error
	listFn         func(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error)
	unreadCountFn  func(ctx context.Context, userID primitive.ObjectID) (int64, error)
	markReadFn     func(ctx context.Context, id, userID primitive.ObjectID) error
	markAllReadFn  func(ctx context.Context, userID primitive.ObjectID) error
	deleteFn       func(ctx context.Context, id, userID primitive.ObjectID) error
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *model.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, n)
	}
	return nil
}

func (m *mockNotificationRepo) List(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockNotificationRepo) UnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	if m.unreadCountFn != nil {
		return m.unreadCountFn(ctx, userID)
	}
	return 0, nil
}

func (m *mockNotificationRepo) MarkRead(ctx context.Context, id, userID primitive.ObjectID) error {
	if m.markReadFn != nil {
		return m.markReadFn(ctx, id, userID)
	}
	return nil
}

func (m *mockNotificationRepo) MarkAllRead(ctx context.Context, userID primitive.ObjectID) error {
	if m.markAllReadFn != nil {
		return m.markAllReadFn(ctx, userID)
	}
	return nil
}

func (m *mockNotificationRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}

// --- Mock NotificationSettingsRepo (shared with notification_service_test.go) ---

type mockNotifSettingsRepo struct {
	getOrCreateFn func(ctx context.Context, userID primitive.ObjectID) (*model.NotificationSettings, error)
	updateFn      func(ctx context.Context, userID primitive.ObjectID, update bson.M) error
}

func (m *mockNotifSettingsRepo) GetOrCreate(ctx context.Context, userID primitive.ObjectID) (*model.NotificationSettings, error) {
	if m.getOrCreateFn != nil {
		return m.getOrCreateFn(ctx, userID)
	}
	return &model.NotificationSettings{}, nil
}

func (m *mockNotifSettingsRepo) Update(ctx context.Context, userID primitive.ObjectID, update bson.M) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, userID, update)
	}
	return nil
}

// --- Tests ---

func TestSocialService_ToggleLike_Liked(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	ownerID := primitive.NewObjectID()

	mockL := &mockLikeRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		toggleWithSessionFn: func(sessCtx context.Context, cid, uid primitive.ObjectID) (bool, error) {
			return true, nil
		},
	}
	mockC := &mockCheckinRepo{
		incrLikeCountWithSessionFn: func(sessCtx context.Context, id primitive.ObjectID, delta int) error {
			assert.Equal(t, 1, delta)
			return nil
		},
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
			return &model.Checkin{ID: checkinID, UserID: ownerID}, nil
		},
	}
	mockU := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			return &model.User{ID: userID, Nickname: "Liker"}, nil
		},
	}
	mockNotif := &mockNotificationRepo{}
	mockNotifSettings := &mockNotifSettingsRepo{
		getOrCreateFn: func(ctx context.Context, uid primitive.ObjectID) (*model.NotificationSettings, error) {
			return &model.NotificationSettings{LikeNotify: true}, nil
		},
	}
	notifSvc := NewNotificationService(mockNotif, mockNotifSettings, mockU)
	svc := NewSocialService(mockComment, mockL, mockC, mockU, notifSvc)

	liked, err := svc.ToggleLike(context.Background(), checkinID, userID)
	require.NoError(t, err)
	assert.True(t, liked)
}

func TestSocialService_ToggleLike_Unliked(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockL := &mockLikeRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		toggleWithSessionFn: func(sessCtx context.Context, cid, uid primitive.ObjectID) (bool, error) {
			return false, nil
		},
	}
	mockC := &mockCheckinRepo{
		incrLikeCountWithSessionFn: func(sessCtx context.Context, id primitive.ObjectID, delta int) error {
			assert.Equal(t, -1, delta)
			return nil
		},
	}
	svc := NewSocialService(mockComment, mockL, mockC, &mockUserRepo{}, nil)

	liked, err := svc.ToggleLike(context.Background(), checkinID, userID)
	require.NoError(t, err)
	assert.False(t, liked)
}

func TestSocialService_ToggleLike_LikeSelf(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockL := &mockLikeRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		toggleWithSessionFn: func(sessCtx context.Context, cid, uid primitive.ObjectID) (bool, error) {
			return true, nil
		},
	}
	mockC := &mockCheckinRepo{
		incrLikeCountWithSessionFn: func(sessCtx context.Context, id primitive.ObjectID, delta int) error {
			return nil
		},
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
			return &model.Checkin{ID: checkinID, UserID: userID}, nil
		},
	}
	svc := NewSocialService(mockComment, mockL, mockC, &mockUserRepo{}, &NotificationService{})

	liked, err := svc.ToggleLike(context.Background(), checkinID, userID)
	require.NoError(t, err)
	assert.True(t, liked)
}

func TestSocialService_ToggleLike_StartSessionError(t *testing.T) {
	mockL := &mockLikeRepo{
		startSessionFn: func() (mongo.Session, error) {
			return nil, fmt.Errorf("session error")
		},
	}
	svc := NewSocialService(mockComment, mockL, &mockCheckinRepo{}, &mockUserRepo{}, nil)

	_, err := svc.ToggleLike(context.Background(), primitive.NewObjectID(), primitive.NewObjectID())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start session")
}

func TestSocialService_AddComment(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	ownerID := primitive.NewObjectID()

	mockComm := &mockCommentRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		createFn: func(ctx context.Context, c *model.Comment) error {
			c.ID = primitive.NewObjectID()
			assert.Equal(t, checkinID, c.CheckinID)
			assert.Equal(t, userID, c.UserID)
			assert.Equal(t, "好功夫！", c.Content)
			return nil
		},
	}
	mockC := &mockCheckinRepo{
		incrCommentCountWithSessionFn: func(sessCtx context.Context, id primitive.ObjectID) error {
			assert.Equal(t, checkinID, id)
			return nil
		},
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
			return &model.Checkin{ID: checkinID, UserID: ownerID}, nil
		},
	}
	mockU := &mockUserRepo{
		findByID: func(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
			return &model.User{ID: userID, Nickname: "Commenter"}, nil
		},
	}
	mockNotifSettings := &mockNotifSettingsRepo{
		getOrCreateFn: func(ctx context.Context, uid primitive.ObjectID) (*model.NotificationSettings, error) {
			return &model.NotificationSettings{CommentNotify: true}, nil
		},
	}
	mockNotifRepo := &mockNotificationRepo{}
	notifSvc := NewNotificationService(mockNotifRepo, mockNotifSettings, mockU)
	svc := NewSocialService(mockComm, &mockLikeRepo{}, mockC, mockU, notifSvc)

	comment, err := svc.AddComment(context.Background(), checkinID, userID, "好功夫！")
	require.NoError(t, err)
	require.NotNil(t, comment)
	assert.Equal(t, checkinID, comment.CheckinID)
	assert.Equal(t, "好功夫！", comment.Content)
	assert.NotNil(t, comment.User)
}

func TestSocialService_AddComment_TransactionError(t *testing.T) {
	mockComm := &mockCommentRepo{
		startSessionFn: func() (mongo.Session, error) {
			return &mockSession{}, nil
		},
		createFn: func(ctx context.Context, c *model.Comment) error {
			return fmt.Errorf("create error")
		},
	}
	svc := NewSocialService(mockComm, &mockLikeRepo{}, &mockCheckinRepo{}, &mockUserRepo{}, nil)

	_, err := svc.AddComment(context.Background(), primitive.NewObjectID(), primitive.NewObjectID(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction failed")
}

func TestSocialService_GetComments(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	commentID := primitive.NewObjectID()

	mockComm := &mockCommentRepo{
		listByCheckinFn: func(ctx context.Context, cid primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error) {
			assert.Equal(t, checkinID, cid)
			return []*model.Comment{
				{ID: commentID, CheckinID: checkinID, UserID: userID, Content: "nice"},
			}, 1, nil
		},
	}
	mockU := &mockUserRepo{
		findByIDs: func(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
			return []*model.User{{ID: userID, Nickname: "User"}}, nil
		},
	}
	svc := NewSocialService(mockComm, &mockLikeRepo{}, &mockCheckinRepo{}, mockU, nil)

	comments, total, err := svc.GetComments(context.Background(), checkinID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, comments, 1)
	assert.NotNil(t, comments[0].User)
	assert.Equal(t, "User", comments[0].User.Nickname)
}

func TestSocialService_BatchIsLiked(t *testing.T) {
	checkinID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockL := &mockLikeRepo{
		batchIsLikedFn: func(ctx context.Context, checkinIDs []primitive.ObjectID, uid primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
			return map[primitive.ObjectID]bool{checkinID: true}, nil
		},
	}
	svc := NewSocialService(&mockCommentRepo{}, mockL, &mockCheckinRepo{}, &mockUserRepo{}, nil)

	result, err := svc.BatchIsLiked(context.Background(), []primitive.ObjectID{checkinID}, userID)
	require.NoError(t, err)
	assert.True(t, result[checkinID])
}

// mockComment var to share across tests
var mockComment = &mockCommentRepo{}

var _ repository.CommentRepoInterface = (*mockCommentRepo)(nil)
var _ repository.LikeRepoInterface = (*mockLikeRepo)(nil)