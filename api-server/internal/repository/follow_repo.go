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

type FollowRepo struct {
	coll *mongo.Collection
}

func NewFollowRepo(db *mongo.Database) *FollowRepo {
	return &FollowRepo{coll: db.Collection("follows")}
}

// Follow 关注用户
func (r *FollowRepo) Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	follow := &model.Follow{
		FollowerID:  followerID,
		FollowingID: followingID,
		CreatedAt:   time.Now(),
	}
	_, err := r.coll.InsertOne(ctx, follow)
	if mongo.IsDuplicateKeyError(err) {
		return nil // 已关注，忽略
	}
	return err
}

// Unfollow 取消关注
func (r *FollowRepo) Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{
		"follower_id":  followerID,
		"following_id": followingID,
	})
	return err
}

// IsFollowing 检查是否已关注
func (r *FollowRepo) IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{
		"follower_id":  followerID,
		"following_id": followingID,
	})
	return count > 0, err
}

// GetFollowing 获取关注列表
func (r *FollowRepo) GetFollowing(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]primitive.ObjectID, int64, error) {
	filter := bson.M{"follower_id": userID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var f model.Follow
		if err := cursor.Decode(&f); err == nil {
			ids = append(ids, f.FollowingID)
		}
	}
	return ids, total, nil
}

// GetFollowers 获取粉丝列表
func (r *FollowRepo) GetFollowers(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]primitive.ObjectID, int64, error) {
	filter := bson.M{"following_id": userID}
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var f model.Follow
		if err := cursor.Decode(&f); err == nil {
			ids = append(ids, f.FollowerID)
		}
	}
	return ids, total, nil
}

// CountFollowing 关注数
func (r *FollowRepo) CountFollowing(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"follower_id": userID})
}

// CountFollowers 粉丝数
func (r *FollowRepo) CountFollowers(ctx context.Context, userID primitive.ObjectID) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"following_id": userID})
}

// EnsureIndexes 创建索引
func (r *FollowRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "follower_id", Value: 1}, {Key: "following_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "following_id", Value: 1}}},
	})
	return err
}

// FollowRepoInterface 关注仓库接口
type FollowRepoInterface interface {
	Follow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	Unfollow(ctx context.Context, followerID, followingID primitive.ObjectID) error
	IsFollowing(ctx context.Context, followerID, followingID primitive.ObjectID) (bool, error)
	GetFollowing(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]primitive.ObjectID, int64, error)
	GetFollowers(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]primitive.ObjectID, int64, error)
	CountFollowing(ctx context.Context, userID primitive.ObjectID) (int64, error)
	CountFollowers(ctx context.Context, userID primitive.ObjectID) (int64, error)
}

var _ FollowRepoInterface = (*FollowRepo)(nil)
