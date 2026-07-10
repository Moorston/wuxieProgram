package repository

import (
	"context"
	"time"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogRepo struct {
	coll *mongo.Collection
}

func NewAuditLogRepo(db *mongo.Database) *AuditLogRepo {
	return &AuditLogRepo{coll: db.Collection("audit_logs")}
}

func (r *AuditLogRepo) Create(ctx context.Context, log *model.AuditLog) error {
	log.CreatedAt = time.Now()
	_, err := r.coll.InsertOne(ctx, log)
	return err
}

func (r *AuditLogRepo) List(ctx context.Context, page, pageSize int) ([]*model.AuditLog, int64, error) {
	total, err := r.coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []*model.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *AuditLogRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "admin_user", Value: 1}}},
	})
	return err
}

// AuditLogInterface 审计日志仓库接口
type AuditLogInterface interface {
	Create(ctx context.Context, log *model.AuditLog) error
	List(ctx context.Context, page, pageSize int) ([]*model.AuditLog, int64, error)
}

var _ AuditLogInterface = (*AuditLogRepo)(nil)
