package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RankHistory 排名历史快照
type RankHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Period    RankPeriod         `bson:"period" json:"period"` // day, week, all
	Rank      int                `bson:"rank" json:"rank"`
	Score     int                `bson:"score" json:"score"`
	SnapshotAt time.Time         `bson:"snapshot_at" json:"snapshot_at"`
}

// RankTrendPoint 排名趋势数据点
type RankTrendPoint struct {
	Date   string `json:"date"`
	Rank   int    `json:"rank"`
	Score  int    `json:"score"`
}
