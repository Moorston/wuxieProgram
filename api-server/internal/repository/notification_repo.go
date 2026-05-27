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

type NotificationRepo struct {
	coll *mongo.Collection
}

func NewNotificationRepo(db *mongo.Database) *NotificationRepo {
	return &NotificationRepo{coll: db.Collection("notifications")}
}

func (r *NotificationRepo) Create(ctx context.Context, n *model.Notification) error {
	n.CreatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, n)
	if err != nil {
		return err
	}
	n.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *NotificationRepo) List(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Notification, int64, error) {
	filter := bson.M{"user_id": userID}

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

	var notifications []*model.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"user_id": userID, "is_read": false})
}

func (r *NotificationRepo) MarkRead(ctx context.Context, id, userID primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id, "user_id": userID}, bson.M{
		"$set": bson.M{"is_read": true},
	})
	return err
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.coll.UpdateMany(ctx, bson.M{"user_id": userID, "is_read": false}, bson.M{
		"$set": bson.M{"is_read": true},
	})
	return err
}

func (r *NotificationRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	return err
}

func (r *NotificationRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_read", Value: 1}}},
		{Keys: bson.D{{Key: "target_type", Value: 1}, {Key: "target_id", Value: 1}}},
	})
	return err
}

type NotificationSettingsRepo struct {
	coll *mongo.Collection
}

func NewNotificationSettingsRepo(db *mongo.Database) *NotificationSettingsRepo {
	return &NotificationSettingsRepo{coll: db.Collection("notification_settings")}
}

func (r *NotificationSettingsRepo) GetOrCreate(ctx context.Context, userID primitive.ObjectID) (*model.NotificationSettings, error) {
	var settings model.NotificationSettings
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID}).Decode(&settings)
	if err == nil {
		return &settings, nil
	}

	settings = model.NotificationSettings{
		UserID:         userID,
		LikeNotify:     true,
		CommentNotify:  true,
		PlanRemind:     true,
		PlanRemindTime: "20:00",
		GroupNotify:    true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	result, err := r.coll.InsertOne(ctx, settings)
	if err != nil {
		return nil, err
	}
	settings.ID = result.InsertedID.(primitive.ObjectID)
	return &settings, nil
}

func (r *NotificationSettingsRepo) Update(ctx context.Context, userID primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.coll.UpdateOne(ctx, bson.M{"user_id": userID}, bson.M{"$set": update})
	return err
}

func (r *NotificationSettingsRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
