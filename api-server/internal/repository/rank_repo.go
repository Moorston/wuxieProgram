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

type GroupRepo struct {
	coll *mongo.Collection
}

func NewGroupRepo(db *mongo.Database) *GroupRepo {
	return &GroupRepo{coll: db.Collection("groups")}
}

func (r *GroupRepo) FindAll(ctx context.Context) ([]*model.Group, error) {
	cursor, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []*model.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *GroupRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Group, error) {
	var group model.Group
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// FindByInviteCode 根据邀请码查找团组
func (r *GroupRepo) FindByInviteCode(ctx context.Context, code string) (*model.Group, error) {
	var group model.Group
	err := r.coll.FindOne(ctx, bson.M{"invite_code": code}).Decode(&group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// AddMember 添加成员到团组
func (r *GroupRepo) AddMember(ctx context.Context, groupID, userID primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": groupID},
		bson.M{
			"$addToSet": bson.M{"member_ids": userID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// UpdateInviteCode 更新团组邀请码
func (r *GroupRepo) UpdateInviteCode(ctx context.Context, groupID primitive.ObjectID, code string) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": groupID},
		bson.M{
			"$set": bson.M{
				"invite_code": code,
				"updated_at":  time.Now(),
			},
		},
	)
	return err
}

// RemoveMember 从团组移除成员
func (r *GroupRepo) RemoveMember(ctx context.Context, groupID, userID primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": groupID},
		bson.M{
			"$pull": bson.M{"member_ids": userID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

// SetLeader 设置团组组长
func (r *GroupRepo) SetLeader(ctx context.Context, groupID, leaderID primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": groupID},
		bson.M{
			"$set": bson.M{
				"leader_id":  leaderID,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

type RankRepo struct {
	coll *mongo.Collection
}

func NewRankRepo(db *mongo.Database) *RankRepo {
	return &RankRepo{coll: db.Collection("rank_cache")}
}

func (r *RankRepo) GetRankList(ctx context.Context, period model.RankPeriod, page, pageSize int) ([]*model.RankEntry, error) {
	filter := bson.M{"period": period}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().
		SetSort(bson.D{{Key: "rank", Value: 1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*model.RankEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (r *RankRepo) GetUserRank(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod) (*model.RankEntry, error) {
	var entry model.RankEntry
	err := r.coll.FindOne(ctx, bson.M{"user_id": userID, "period": period}).Decode(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *RankRepo) RefreshRank(ctx context.Context, period model.RankPeriod, entries []*model.RankEntry) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if len(entries) == 0 {
		// 没有新数据时，清空该周期排名
		_, err := r.coll.DeleteMany(ctx, bson.M{"period": period})
		return err
	}

	// 两阶段无窗口更新：
	// 1. 先批量 upsert 新排名（替换旧数据，读者始终能看到数据）
	// 2. 再删除已不在新排名中的旧条目
	now := time.Now()
	for i, e := range entries {
		e.Period = period
		e.UpdateAt = now
	}

	// 收集新排名中的用户ID，用于清理旧数据
	newUserIDs := make([]primitive.ObjectID, len(entries))
	for i, e := range entries {
		newUserIDs[i] = e.UserID
	}

	// 第一阶段：批量 upsert 新排名
	var models []mongo.WriteModel
	for _, e := range entries {
		filter := bson.M{"user_id": e.UserID, "period": period}
		model := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(e).
			SetUpsert(true)
		models = append(models, model)
	}

	// 第二阶段：删除已不在新排名中的旧条目
	oldFilter := bson.M{
		"period":  period,
		"user_id": bson.M{"$nin": newUserIDs},
	}
	models = append(models, mongo.NewDeleteManyModel().SetFilter(oldFilter))

	_, err := r.coll.BulkWrite(ctx, models)
	return err
}
