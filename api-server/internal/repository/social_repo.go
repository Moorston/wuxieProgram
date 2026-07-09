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

type CommentRepo struct {
	coll *mongo.Collection
}

func NewCommentRepo(db *mongo.Database) *CommentRepo {
	return &CommentRepo{coll: db.Collection("comments")}
}

func (r *CommentRepo) Create(ctx context.Context, c *model.Comment) error {
	c.CreatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	c.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *CommentRepo) ListByCheckin(ctx context.Context, checkinID primitive.ObjectID, page, pageSize int) ([]*model.Comment, int64, error) {
	filter := bson.M{"checkin_id": checkinID}

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

	var comments []*model.Comment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

type LikeRepo struct {
	coll *mongo.Collection
}

func NewLikeRepo(db *mongo.Database) *LikeRepo {
	return &LikeRepo{coll: db.Collection("likes")}
}

func (r *LikeRepo) Toggle(ctx context.Context, checkinID, userID primitive.ObjectID) (liked bool, err error) {
	filter := bson.M{"checkin_id": checkinID, "user_id": userID}

	count, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	if count > 0 {
		_, err = r.coll.DeleteOne(ctx, filter)
		return false, err
	}

	like := &model.Like{
		CheckinID: checkinID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	_, err = r.coll.InsertOne(ctx, like)
	return true, err
}

// ToggleWithSession 在事务中切换点赞状态
func (r *LikeRepo) ToggleWithSession(sessCtx mongo.SessionContext, checkinID, userID primitive.ObjectID) (liked bool, err error) {
	filter := bson.M{"checkin_id": checkinID, "user_id": userID}

	count, err := r.coll.CountDocuments(sessCtx, filter)
	if err != nil {
		return false, err
	}

	if count > 0 {
		_, err = r.coll.DeleteOne(sessCtx, filter)
		return false, err
	}

	like := &model.Like{
		CheckinID: checkinID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	_, err = r.coll.InsertOne(sessCtx, like)
	return true, err
}

func (r *LikeRepo) IsLiked(ctx context.Context, checkinID, userID primitive.ObjectID) (bool, error) {
	count, err := r.coll.CountDocuments(ctx, bson.M{"checkin_id": checkinID, "user_id": userID})
	return count > 0, err
}

func (r *LikeRepo) BatchIsLiked(ctx context.Context, checkinIDs []primitive.ObjectID, userID primitive.ObjectID) (map[primitive.ObjectID]bool, error) {
	cursor, err := r.coll.Find(ctx, bson.M{
		"checkin_id": bson.M{"$in": checkinIDs},
		"user_id":    userID,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[primitive.ObjectID]bool)
	for cursor.Next(ctx) {
		var like model.Like
		if err := cursor.Decode(&like); err != nil {
			return nil, err
		}
		result[like.CheckinID] = true
	}
	return result, nil
}

func (r *LikeRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "checkin_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

// StartSession 启动MongoDB会话用于事务
func (r *LikeRepo) StartSession() (mongo.Session, error) {
	return r.coll.Database().Client().StartSession()
}

// StartSession 启动MongoDB会话用于事务
func (r *CommentRepo) StartSession() (mongo.Session, error) {
	return r.coll.Database().Client().StartSession()
}
