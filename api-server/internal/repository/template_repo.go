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

type TemplateRepo struct {
	coll *mongo.Collection
}

func NewTemplateRepo(db *mongo.Database) *TemplateRepo {
	return &TemplateRepo{coll: db.Collection("training_templates")}
}

func (r *TemplateRepo) Create(ctx context.Context, t *model.TrainingTemplate) error {
	t.CreatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, t)
	if err != nil {
		return err
	}
	t.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TemplateRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.TrainingTemplate, error) {
	var t model.TrainingTemplate
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) List(ctx context.Context, category, style string, page, pageSize int) ([]*model.TrainingTemplate, int64, error) {
	filter := bson.M{}
	if category != "" {
		filter["category"] = category
	}
	if style != "" {
		filter["style"] = style
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().
		SetSort(bson.D{{Key: "usage_count", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var templates []*model.TrainingTemplate
	if err := cursor.All(ctx, &templates); err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

func (r *TemplateRepo) IncrUsageCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"usage_count": 1},
	})
	return err
}

func (r *TemplateRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "style", Value: 1}}},
		{Keys: bson.D{{Key: "category", Value: 1}}},
		{Keys: bson.D{{Key: "usage_count", Value: -1}}},
	})
	return err
}
