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

type InsightRepo struct {
	coll *mongo.Collection
}

func NewInsightRepo(db *mongo.Database) *InsightRepo {
	return &InsightRepo{coll: db.Collection("insights")}
}

func (r *InsightRepo) Create(ctx context.Context, insight *model.Insight) error {
	insight.CreatedAt = time.Now()
	insight.UpdatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, insight)
	if err != nil {
		return err
	}
	insight.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *InsightRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Insight, error) {
	var insight model.Insight
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&insight)
	if err != nil {
		return nil, err
	}
	return &insight, nil
}

func (r *InsightRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *InsightRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	return err
}

func (r *InsightRepo) ListByUser(ctx context.Context, userID primitive.ObjectID, tag, mood string, page, pageSize int) ([]*model.Insight, int64, error) {
	filter := bson.M{"user_id": userID}
	if tag != "" {
		filter["tags"] = tag
	}
	if mood != "" {
		filter["mood"] = mood
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

	var insights []*model.Insight
	if err := cursor.All(ctx, &insights); err != nil {
		return nil, 0, err
	}

	return insights, total, nil
}

func (r *InsightRepo) ListPublic(ctx context.Context, page, pageSize int) ([]*model.Insight, int64, error) {
	filter := bson.M{"visibility": model.VisibilityPublic}

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

	var insights []*model.Insight
	if err := cursor.All(ctx, &insights); err != nil {
		return nil, 0, err
	}

	return insights, total, nil
}

func (r *InsightRepo) OnThisDay(ctx context.Context, userID primitive.ObjectID, month, day int) ([]*model.Insight, error) {
	filter := bson.M{
		"user_id": userID,
		"$expr": bson.M{
			"$and": []bson.M{
				{"$eq": []interface{}{bson.M{"$month": "$created_at"}, month}},
				{"$eq": []interface{}{bson.M{"$dayOfMonth": "$created_at"}, day}},
			},
		},
	}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []*model.Insight
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *InsightRepo) MoodStats(ctx context.Context, userID primitive.ObjectID, days int) (map[string]int, error) {
	since := time.Now().AddDate(0, 0, -days)

	filter := bson.M{
		"user_id":    userID,
		"created_at": bson.M{"$gte": since},
	}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := map[string]int{
		"breakthrough": 0,
		"good":         0,
		"normal":       0,
		"confused":     0,
		"low":          0,
	}

	for cursor.Next(ctx) {
		var ins model.Insight
		if err := cursor.Decode(&ins); err != nil {
			continue
		}
		stats[string(ins.Mood)]++
	}

	return stats, nil
}

func (r *InsightRepo) IncrLikeCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"like_count": delta},
	})
	return err
}

// StartSession 启动MongoDB会话用于事务
func (r *InsightRepo) StartSession() (mongo.Session, error) {
	return r.coll.Database().Client().StartSession()
}

// IncrLikeCountWithSession 在事务中增加点赞计数
func (r *InsightRepo) IncrLikeCountWithSession(sessCtx mongo.SessionContext, id primitive.ObjectID, delta int) error {
	_, err := r.coll.UpdateOne(sessCtx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"like_count": delta},
	})
	return err
}

func (r *InsightRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "tags", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "mood", Value: 1}}},
		{Keys: bson.D{{Key: "visibility", Value: 1}, {Key: "created_at", Value: -1}}},
	})
	return err
}

type InsightLikeRepo struct {
	coll *mongo.Collection
}

func NewInsightLikeRepo(db *mongo.Database) *InsightLikeRepo {
	return &InsightLikeRepo{coll: db.Collection("insight_likes")}
}

func (r *InsightLikeRepo) Toggle(ctx context.Context, insightID, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{"insight_id": insightID, "user_id": userID}
	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	if count > 0 {
		_, err = r.coll.DeleteOne(ctx, filter)
		return false, err
	}

	like := &model.InsightLike{
		InsightID: insightID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	_, err = r.coll.InsertOne(ctx, like)
	return true, err
}

// ToggleWithSession 在事务中切换点赞状态
func (r *InsightLikeRepo) ToggleWithSession(sessCtx mongo.SessionContext, insightID, userID primitive.ObjectID) (bool, error) {
	filter := bson.M{"insight_id": insightID, "user_id": userID}

	count, err := r.coll.CountDocuments(sessCtx, filter)
	if err != nil {
		return false, err
	}

	if count > 0 {
		_, err = r.coll.DeleteOne(sessCtx, filter)
		return false, err
	}

	like := &model.InsightLike{
		InsightID: insightID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	_, err = r.coll.InsertOne(sessCtx, like)
	return true, err
}

func (r *InsightLikeRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "insight_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
