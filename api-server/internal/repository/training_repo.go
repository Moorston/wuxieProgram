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

type TrainingRepo struct {
	coll *mongo.Collection
}

func NewTrainingRepo(db *mongo.Database) *TrainingRepo {
	return &TrainingRepo{coll: db.Collection("training_plans")}
}

func (r *TrainingRepo) Create(ctx context.Context, plan *model.TrainingPlan) error {
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, plan)
	if err != nil {
		return err
	}
	plan.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TrainingRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.TrainingPlan, error) {
	var plan model.TrainingPlan
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&plan)
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *TrainingRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *TrainingRepo) UpdateTasks(ctx context.Context, id primitive.ObjectID, days []model.TrainingDay, stats model.PlanStats) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"days":       days,
			"stats":      stats,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *TrainingRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	return err
}

func (r *TrainingRepo) ListByUser(ctx context.Context, userID primitive.ObjectID, status *model.PlanStatus, page, pageSize int) ([]*model.TrainingPlan, int64, error) {
	filter := bson.M{"user_id": userID}
	if status != nil {
		filter["status"] = *status
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var plans []*model.TrainingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, 0, err
	}

	return plans, total, nil
}

func (r *TrainingRepo) GetTodayTasks(ctx context.Context, userID primitive.ObjectID) ([]*model.TrainingPlan, error) {
	today := time.Now()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	todayEnd := todayStart.AddDate(0, 0, 1)

	filter := bson.M{
		"user_id":    userID,
		"status":     model.PlanStatusActive,
		"start_date": bson.M{"$lte": todayEnd},
		"end_date":   bson.M{"$gte": todayStart},
	}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []*model.TrainingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *TrainingRepo) ListByGroup(ctx context.Context, groupID primitive.ObjectID) ([]*model.TrainingPlan, error) {
	filter := bson.M{"group_id": groupID}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []*model.TrainingPlan
	if err := cursor.All(ctx, &plans); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *TrainingRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "group_id", Value: 1}}},
		{Keys: bson.D{{Key: "start_date", Value: 1}, {Key: "end_date", Value: 1}}},
	})
	return err
}

func (r *TrainingRepo) FindActive(ctx context.Context, filter bson.M) (mongo.Cursor, error) {
	return r.coll.Find(ctx, filter)
}
