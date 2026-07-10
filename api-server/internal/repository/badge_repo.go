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

type BadgeRepo struct {
	coll *mongo.Collection
}

func NewBadgeRepo(db *mongo.Database) *BadgeRepo {
	return &BadgeRepo{coll: db.Collection("badges")}
}

func (r *BadgeRepo) FindAll(ctx context.Context) ([]*model.Badge, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort_order", Value: 1}})
	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var badges []*model.Badge
	if err := cursor.All(ctx, &badges); err != nil {
		return nil, err
	}
	return badges, nil
}

func (r *BadgeRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Badge, error) {
	var badge model.Badge
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&badge)
	if err != nil {
		return nil, err
	}
	return &badge, nil
}

func (r *BadgeRepo) SeedDefaults(ctx context.Context) error {
	for _, badge := range model.DefaultBadges {
		_, err := r.coll.UpdateOne(ctx,
			bson.M{"name": badge.Name, "type": badge.Type},
			bson.M{"$setOnInsert": badge},
			options.Update().SetUpsert(true),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *BadgeRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "sort_order", Value: 1}},
	})
	return err
}

// UserBadgeRepo 用户徽章仓库
type UserBadgeRepo struct {
	coll *mongo.Collection
}

func NewUserBadgeRepo(db *mongo.Database) *UserBadgeRepo {
	return &UserBadgeRepo{coll: db.Collection("user_badges")}
}

func (r *UserBadgeRepo) Grant(ctx context.Context, userID, badgeID primitive.ObjectID) error {
	doc := bson.M{
		"user_id":   userID,
		"badge_id":  badgeID,
		"earned_at": time.Now(),
	}
	_, err := r.coll.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return nil // 已拥有，忽略
	}
	return err
}

func (r *UserBadgeRepo) HasBadge(ctx context.Context, userID, badgeID primitive.ObjectID) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"user_id":  userID,
		"badge_id": badgeID,
	})
	return count > 0, err
}

func (r *UserBadgeRepo) ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.UserBadge, error) {
	opts := options.Find().SetSort(bson.D{{Key: "earned_at", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var badges []*model.UserBadge
	if err := cursor.All(ctx, &badges); err != nil {
		return nil, err
	}
	return badges, nil
}

func (r *UserBadgeRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "badge_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	})
	return err
}

// BadgeRepoInterface 徽章仓库接口
type BadgeRepoInterface interface {
	FindAll(ctx context.Context) ([]*model.Badge, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Badge, error)
	SeedDefaults(ctx context.Context) error
}

// UserBadgeRepoInterface 用户徽章仓库接口
type UserBadgeRepoInterface interface {
	Grant(ctx context.Context, userID, badgeID primitive.ObjectID) error
	HasBadge(ctx context.Context, userID, badgeID primitive.ObjectID) (bool, error)
	ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*model.UserBadge, error)
}

var _ BadgeRepoInterface = (*BadgeRepo)(nil)
var _ UserBadgeRepoInterface = (*UserBadgeRepo)(nil)
