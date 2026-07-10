package repository

import (
	"context"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate mockgen -destination=mock_training_repo.go -package=repository wuxie-api/internal/repository TrainingRepoInterface
//go:generate mockgen -destination=mock_template_repo.go -package=repository wuxie-api/internal/repository TemplateRepoInterface

// TrainingRepoInterface 训练计划仓库接口
type TrainingRepoInterface interface {
	Create(ctx context.Context, plan *model.TrainingPlan) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	UpdateTasks(ctx context.Context, id primitive.ObjectID, days []model.TrainingDay, stats model.PlanStats) error
	Delete(ctx context.Context, id, userID primitive.ObjectID) error
	ListByUser(ctx context.Context, userID primitive.ObjectID, status *model.PlanStatus, page, pageSize int) ([]*model.TrainingPlan, int64, error)
	GetTodayTasks(ctx context.Context, userID primitive.ObjectID) ([]*model.TrainingPlan, error)
	ListByGroup(ctx context.Context, groupID primitive.ObjectID) ([]*model.TrainingPlan, error)
	FindActive(ctx context.Context, filter bson.M) (mongo.Cursor, error)
}

// TemplateRepoInterface 训练模板仓库接口
type TemplateRepoInterface interface {
	Create(ctx context.Context, t *model.TrainingTemplate) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.TrainingTemplate, error)
	List(ctx context.Context, category, style string, page, pageSize int) ([]*model.TrainingTemplate, int64, error)
	IncrUsageCount(ctx context.Context, id primitive.ObjectID) error
}

var _ TrainingRepoInterface = (*TrainingRepo)(nil)
var _ TemplateRepoInterface = (*TemplateRepo)(nil)
