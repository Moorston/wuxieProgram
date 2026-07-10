package repository

import (
	"context"
	"sort"
	"time"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RankHistoryRepo struct {
	coll *mongo.Collection
}

func NewRankHistoryRepo(db *mongo.Database) *RankHistoryRepo {
	return &RankHistoryRepo{coll: db.Collection("rank_history")}
}

// SaveSnapshot 保存排名快照（定时任务调用）
func (r *RankHistoryRepo) SaveSnapshot(ctx context.Context, entries []*model.RankEntry, period model.RankPeriod) error {
	if len(entries) == 0 {
		return nil
	}

	now := time.Now()
	docs := make([]interface{}, 0, len(entries))
	for _, entry := range entries {
		docs = append(docs, &model.RankHistory{
			UserID:     entry.UserID,
			Period:     period,
			Rank:       entry.Rank,
			Score:      entry.Score,
			SnapshotAt: now,
		})
	}

	_, err := r.coll.InsertMany(ctx, docs)
	return err
}

// GetUserHistory 获取用户排名历史
func (r *RankHistoryRepo) GetUserHistory(ctx context.Context, userID primitive.ObjectID, period model.RankPeriod, days int) ([]model.RankTrendPoint, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	filter := bson.M{
		"user_id":     userID,
		"period":      period,
		"snapshot_at": bson.M{"$gte": startDate},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "snapshot_at", Value: 1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []model.RankHistory
	if err := cursor.All(ctx, &history); err != nil {
		return nil, err
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// 聚合为每天一个数据点（取当天最新快照）
	dayMap := make(map[string]*model.RankHistory)
	for i := range history {
		day := history[i].SnapshotAt.Format("2006-01-02")
		if existing, ok := dayMap[day]; !ok || history[i].SnapshotAt.After(existing.SnapshotAt) {
			dayMap[day] = &history[i]
		}
	}

	// 按日期排序
	points := make([]model.RankTrendPoint, 0, len(dayMap))
	for _, h := range dayMap {
		points = append(points, model.RankTrendPoint{
			Date:  h.SnapshotAt.Format("2006-01-02"),
			Rank:  h.Rank,
			Score: h.Score,
		})
	}

	// 按日期升序排序
	sort.Slice(points, func(i, j int) bool {
		return points[i].Date < points[j].Date
	})

	return points, nil
}

func (r *RankHistoryRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "period", Value: 1}, {Key: "snapshot_at", Value: -1}}},
		{Keys: bson.D{{Key: "snapshot_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(90 * 24 * 3600)}, // 90天TTL
	})
	return err
}
