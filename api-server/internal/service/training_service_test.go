package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"wuxie-api/internal/model"
	"wuxie-api/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// --- Mock TrainingRepo ---

type mockTrainingRepo struct {
	createFn      func(ctx context.Context, plan *model.TrainingPlan) error
	findByIDFn    func(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error)
	updateFn      func(ctx context.Context, id primitive.ObjectID, update bson.M) error
	updateTasksFn func(ctx context.Context, id primitive.ObjectID, days []model.TrainingDay, stats model.PlanStats) error
	deleteFn      func(ctx context.Context, id, userID primitive.ObjectID) error
	listByUserFn  func(ctx context.Context, userID primitive.ObjectID, status *model.PlanStatus, page, pageSize int) ([]*model.TrainingPlan, int64, error)
	getTodayTasksFn func(ctx context.Context, userID primitive.ObjectID) ([]*model.TrainingPlan, error)
	listByGroupFn func(ctx context.Context, groupID primitive.ObjectID) ([]*model.TrainingPlan, error)
	findActiveFn  func(ctx context.Context, filter bson.M) (mongo.Cursor, error)
}

func (m *mockTrainingRepo) Create(ctx context.Context, plan *model.TrainingPlan) error {
	if m.createFn != nil {
		return m.createFn(ctx, plan)
	}
	return nil
}

func (m *mockTrainingRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockTrainingRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, update)
	}
	return nil
}

func (m *mockTrainingRepo) UpdateTasks(ctx context.Context, id primitive.ObjectID, days []model.TrainingDay, stats model.PlanStats) error {
	if m.updateTasksFn != nil {
		return m.updateTasksFn(ctx, id, days, stats)
	}
	return nil
}

func (m *mockTrainingRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}

func (m *mockTrainingRepo) ListByUser(ctx context.Context, userID primitive.ObjectID, status *model.PlanStatus, page, pageSize int) ([]*model.TrainingPlan, int64, error) {
	if m.listByUserFn != nil {
		return m.listByUserFn(ctx, userID, status, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockTrainingRepo) GetTodayTasks(ctx context.Context, userID primitive.ObjectID) ([]*model.TrainingPlan, error) {
	if m.getTodayTasksFn != nil {
		return m.getTodayTasksFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockTrainingRepo) ListByGroup(ctx context.Context, groupID primitive.ObjectID) ([]*model.TrainingPlan, error) {
	if m.listByGroupFn != nil {
		return m.listByGroupFn(ctx, groupID)
	}
	return nil, nil
}

func (m *mockTrainingRepo) FindActive(ctx context.Context, filter bson.M) (mongo.Cursor, error) {
	if m.findActiveFn != nil {
		return m.findActiveFn(ctx, filter)
	}
	return nil, nil
}

// --- Mock TemplateRepo ---

type mockTemplateRepo struct {
	createFn        func(ctx context.Context, t *model.TrainingTemplate) error
	findByIDFn      func(ctx context.Context, id primitive.ObjectID) (*model.TrainingTemplate, error)
	listFn          func(ctx context.Context, category, style string, page, pageSize int) ([]*model.TrainingTemplate, int64, error)
	incrUsageCountFn func(ctx context.Context, id primitive.ObjectID) error
}

func (m *mockTemplateRepo) Create(ctx context.Context, t *model.TrainingTemplate) error {
	if m.createFn != nil {
		return m.createFn(ctx, t)
	}
	return nil
}

func (m *mockTemplateRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.TrainingTemplate, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockTemplateRepo) List(ctx context.Context, category, style string, page, pageSize int) ([]*model.TrainingTemplate, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, category, style, page, pageSize)
	}
	return nil, 0, nil
}

func (m *mockTemplateRepo) IncrUsageCount(ctx context.Context, id primitive.ObjectID) error {
	if m.incrUsageCountFn != nil {
		return m.incrUsageCountFn(ctx, id)
	}
	return nil
}

// --- Tests ---

func TestTrainingService_CreatePlan(t *testing.T) {
	userID := primitive.NewObjectID()
	planID := primitive.NewObjectID()

	mockPlan := &mockTrainingRepo{
		createFn: func(ctx context.Context, plan *model.TrainingPlan) error {
			plan.ID = planID
			assert.Equal(t, userID, plan.UserID)
			assert.Equal(t, model.PlanStatusActive, plan.Status)
			assert.Equal(t, 2, plan.Stats.TotalTasks)
			return nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	plan := &model.TrainingPlan{
		Title: "我的训练计划",
		Days: []model.TrainingDay{
			{Day: 1, Tasks: []model.TrainingTask{{Title: "热身"}}},
			{Day: 2, Tasks: []model.TrainingTask{{Title: "套路练习"}, {Title: "力量训练"}}},
		},
	}
	err := svc.CreatePlan(context.Background(), userID, plan)
	require.NoError(t, err)
	assert.Equal(t, planID, plan.ID)
	assert.Equal(t, 2, plan.Stats.TotalTasks)
}

func TestTrainingService_GetPlan(t *testing.T) {
	planID := primitive.NewObjectID()
	mockPlan := &mockTrainingRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error) {
			return &model.TrainingPlan{ID: planID, Title: "test plan"}, nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	plan, err := svc.GetPlan(context.Background(), planID)
	require.NoError(t, err)
	assert.Equal(t, planID, plan.ID)
}

func TestTrainingService_ListPlans(t *testing.T) {
	userID := primitive.NewObjectID()
	status := model.PlanStatusActive

	mockPlan := &mockTrainingRepo{
		listByUserFn: func(ctx context.Context, uid primitive.ObjectID, s *model.PlanStatus, page, pageSize int) ([]*model.TrainingPlan, int64, error) {
			assert.Equal(t, userID, uid)
			assert.Equal(t, status, *s)
			return []*model.TrainingPlan{
				{ID: primitive.NewObjectID(), Title: "plan1"},
			}, 1, nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	plans, total, err := svc.ListPlans(context.Background(), userID, &status, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, plans, 1)
}

func TestTrainingService_UpdatePlan(t *testing.T) {
	planID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockPlan := &mockTrainingRepo{
		updateFn: func(ctx context.Context, id primitive.ObjectID, update bson.M) error {
			assert.Equal(t, planID, id)
			return nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	err := svc.UpdatePlan(context.Background(), planID, userID, bson.M{"title": "new title"})
	require.NoError(t, err)
}

func TestTrainingService_DeletePlan(t *testing.T) {
	planID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	mockPlan := &mockTrainingRepo{
		deleteFn: func(ctx context.Context, id, uid primitive.ObjectID) error {
			assert.Equal(t, planID, id)
			assert.Equal(t, userID, uid)
			return nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	err := svc.DeletePlan(context.Background(), planID, userID)
	require.NoError(t, err)
}

func TestTrainingService_GetTodayTasks(t *testing.T) {
	userID := primitive.NewObjectID()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	mockPlan := &mockTrainingRepo{
		getTodayTasksFn: func(ctx context.Context, uid primitive.ObjectID) ([]*model.TrainingPlan, error) {
			return []*model.TrainingPlan{
				{
					ID:        primitive.NewObjectID(),
					Title:     "今日计划",
					StartDate: today,
					Days: []model.TrainingDay{
						{
							Day:   0,
							Date:  today,
							Tasks: []model.TrainingTask{{Title: "晨练", Status: model.TaskStatusPending}},
						},
					},
				},
			}, nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	tasks, err := svc.GetTodayTasks(context.Background(), userID)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	day, ok := tasks[0]["day"].(model.TrainingDay)
	assert.True(t, ok)
	assert.Len(t, day.Tasks, 1)
}

func TestTrainingService_GetTodayTasks_NoPlan(t *testing.T) {
	userID := primitive.NewObjectID()

	mockPlan := &mockTrainingRepo{
		getTodayTasksFn: func(ctx context.Context, uid primitive.ObjectID) ([]*model.TrainingPlan, error) {
			return nil, nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	tasks, err := svc.GetTodayTasks(context.Background(), userID)
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestTrainingService_ApplyTemplate(t *testing.T) {
	userID := primitive.NewObjectID()
	templateID := primitive.NewObjectID()
	planID := primitive.NewObjectID()
	startDate := time.Now()

	mockTemplate := &mockTemplateRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.TrainingTemplate, error) {
			return &model.TrainingTemplate{
				ID:           templateID,
				Name:         "基础训练模板",
				Description:  "适合初学者",
				DurationDays: 3,
				Days: []model.TrainingDay{
					{Day: 1, Tasks: []model.TrainingTask{{Title: "热身", Type: model.TaskTypeBasic, Duration: 10}}},
				},
			}, nil
		},
		incrUsageCountFn: func(ctx context.Context, id primitive.ObjectID) error {
			assert.Equal(t, templateID, id)
			return nil
		},
	}
	mockPlan := &mockTrainingRepo{
		createFn: func(ctx context.Context, plan *model.TrainingPlan) error {
			plan.ID = planID
			assert.Equal(t, userID, plan.UserID)
			assert.Equal(t, "基础训练模板", plan.Title)
			assert.Equal(t, model.PlanStatusActive, plan.Status)
			assert.Equal(t, 1, plan.Stats.TotalTasks)
			return nil
		},
	}
	svc := NewTrainingService(mockPlan, mockTemplate, nil)

	plan, err := svc.ApplyTemplate(context.Background(), userID, templateID, startDate)
	require.NoError(t, err)
	require.NotNil(t, plan)
	assert.Equal(t, planID, plan.ID)
	assert.Equal(t, "基础训练模板", plan.Title)
}

func TestTrainingService_UpdateTaskStatus(t *testing.T) {
	planID := primitive.NewObjectID()
	userID := primitive.NewObjectID()
	checkinID := primitive.NewObjectID()

	mockPlan := &mockTrainingRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error) {
			return &model.TrainingPlan{
				ID:     planID,
				UserID: userID,
				Title:  "test plan",
				Days: []model.TrainingDay{
					{
						Day: 0,
						Tasks: []model.TrainingTask{
							{Title: "task1", Status: model.TaskStatusPending},
							{Title: "task2", Status: model.TaskStatusPending},
						},
					},
					{
						Day: 1,
						Tasks: []model.TrainingTask{
							{Title: "task3", Status: model.TaskStatusPending},
						},
					},
				},
			}, nil
		},
		updateTasksFn: func(ctx context.Context, id primitive.ObjectID, days []model.TrainingDay, stats model.PlanStats) error {
			assert.Equal(t, planID, id)
			assert.Equal(t, 3, stats.TotalTasks)
			assert.Equal(t, 1, stats.Completed)
			return nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	err := svc.UpdateTaskStatus(context.Background(), planID, 0, 0, model.TaskStatusDone, &checkinID)
	require.NoError(t, err)
}

func TestTrainingService_UpdateTaskStatus_OutOfRange(t *testing.T) {
	mockPlan := &mockTrainingRepo{
		findByIDFn: func(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error) {
			return &model.TrainingPlan{
				Days: []model.TrainingDay{
					{Day: 0, Tasks: []model.TrainingTask{{Title: "task1"}}},
				},
			}, nil
		},
	}
	svc := NewTrainingService(mockPlan, &mockTemplateRepo{}, nil)

	err := svc.UpdateTaskStatus(context.Background(), primitive.NewObjectID(), 5, 0, model.TaskStatusDone, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of range")
}

var _ repository.TrainingRepoInterface = (*mockTrainingRepo)(nil)
var _ repository.TemplateRepoInterface = (*mockTemplateRepo)(nil)