package repository

import (
	"context"
	"fmt"
	"time"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepo struct {
	coll *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	return &UserRepo{coll: db.Collection("users")}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return fmt.Errorf("unexpected InsertedID type: %T", result.InsertedID)
	}
	user.ID = id
	return nil
}

func (r *UserRepo) FindByOpenID(ctx context.Context, openid string) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"openid": openid}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpsertByOpenID 原子化：查找或创建用户，同时更新昵称/头像/性别
// 返回 user 和 isCreated 标志
func (r *UserRepo) UpsertByOpenID(ctx context.Context, openid, nickname, avatar string, gender int) (*model.User, bool, error) {
	now := time.Now()
	filter := bson.M{"openid": openid}
	update := bson.M{
		"$setOnInsert": bson.M{
			"openid":     openid,
			"score":      0,
			"check_days": 0,
			"status":     0,
			"created_at": now,
		},
		"$set": bson.M{
			"nickname":   nickname,
			"avatar":     avatar,
			"gender":     gender,
			"updated_at": now,
		},
	}
	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var user model.User
	err := r.coll.FindOneAndUpdate(ctx, filter, update, opts).Decode(&user)
	if err != nil {
		return nil, false, err
	}

	// 判断是新建还是已有：如果 score=0 且 check_days=0 且 created_at 在 5 秒内，认为是新建
	// 使用宽松阈值避免时钟偏差问题
	isCreated := user.Score == 0 && user.CheckDays == 0 && now.Sub(user.CreatedAt) < 5*time.Second
	return &user, isCreated, nil
}

func (r *UserRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// IsBanned 检查用户是否被封禁（轻量查询，只返回状态）
func (r *UserRepo) IsBanned(ctx context.Context, id primitive.ObjectID) (bool, error) {
	var user model.User
	err := r.coll.FindOne(ctx, bson.M{"_id": id}, options.FindOne().SetProjection(bson.M{"status": 1})).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil // 用户不存在，不认为是封禁
		}
		return false, err
	}
	return user.Status == 1, nil
}

// FindAll 分页查询用户（支持关键词搜索）
func (r *UserRepo) FindAll(ctx context.Context, page, pageSize int, keyword string) ([]*model.User, int64, error) {
	filter := bson.M{}
	if keyword != "" {
		filter["nickname"] = bson.M{"$regex": keyword, "$options": "i"}
	}

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

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// Count 统计用户总数
func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{})
}

// CountByStatus 按状态统计用户数
func (r *UserRepo) CountByStatus(ctx context.Context, status int) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"status": status})
}

func (r *UserRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *UserRepo) IncrScore(ctx context.Context, id primitive.ObjectID, score int) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"score": score, "check_days": 1},
		"$set": bson.M{"updated_at": time.Now()},
	})
	return err
}

func (r *UserRepo) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*model.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) FindByGroupID(ctx context.Context, groupID primitive.ObjectID) ([]*model.User, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"group_id": groupID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) FindTopByScore(ctx context.Context, limit int) ([]*model.User, error) {
	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*model.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "openid", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}
