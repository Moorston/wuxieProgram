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

	// 删除旧排名
	_, err := r.coll.DeleteMany(ctx, bson.M{"period": period})
	if err != nil {
		return err
	}

	// 批量插入新排名
	if len(entries) == 0 {
		return nil
	}

	docs := make([]interface{}, len(entries))
	for i, e := range entries {
		e.Period = period
		e.UpdateAt = time.Now()
		docs[i] = e
	}

	_, err = r.coll.InsertMany(ctx, docs)
	return err
}
