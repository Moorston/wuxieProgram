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

type InsightTagRepo struct {
	coll *mongo.Collection
}

func NewInsightTagRepo(db *mongo.Database) *InsightTagRepo {
	return &InsightTagRepo{coll: db.Collection("insight_tags")}
}

func (r *InsightTagRepo) UpsertTags(ctx context.Context, userID primitive.ObjectID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	now := time.Now()
	var models []mongo.WriteModel

	for _, tag := range tags {
		filter := bson.M{"user_id": userID, "tag": tag}
		update := bson.M{
			"$inc": bson.M{"count": 1},
			"$set": bson.M{"updated_at": now},
			"$setOnInsert": bson.M{
				"user_id":    userID,
				"tag":        tag,
				"created_at": now,
			},
		}
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
		models = append(models, model)
	}

	_, err := r.coll.BulkWrite(ctx, models)
	return err
}

func (r *InsightTagRepo) DecrTags(ctx context.Context, userID primitive.ObjectID, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	var models []mongo.WriteModel

	for _, tag := range tags {
		filter := bson.M{"user_id": userID, "tag": tag, "count": bson.M{"$gt": 0}}
		update := bson.M{
			"$inc": bson.M{"count": -1},
			"$set": bson.M{"updated_at": time.Now()},
		}
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update)
		models = append(models, model)
	}

	_, err := r.coll.BulkWrite(ctx, models)
	return err
}

func (r *InsightTagRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.InsightTag, error) {
	opts := options.Find().SetSort(bson.D{{Key: "count", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tags []*model.InsightTag
	if err := cursor.All(ctx, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (r *InsightTagRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "tag", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "count", Value: -1}}},
	})
	return err
}