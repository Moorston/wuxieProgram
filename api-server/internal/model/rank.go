package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RankPeriod string

const (
	RankPeriodDay RankPeriod = "day"
	RankPeriodWeek RankPeriod = "week"
	RankPeriodAll RankPeriod = "all"
)

type RankEntry struct {
	UserID   primitive.ObjectID `bson:"user_id" json:"user_id"`
	Score    int                `bson:"score" json:"score"`
	Rank     int                `bson:"rank" json:"rank"`
	Period   RankPeriod         `bson:"period" json:"period"`
	UpdateAt time.Time          `bson:"updated_at" json:"updated_at"`

	User *User `bson:"-" json:"user,omitempty"`
}
