package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CompetitionStatus 赛事状态
type CompetitionStatus int

const (
	CompetitionStatusDraft  CompetitionStatus = 0 // 草稿
	CompetitionStatusActive CompetitionStatus = 1 // 进行中
	CompetitionStatusEnded  CompetitionStatus = 2 // 已结束
)

// Competition 赛事活动
type Competition struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Title       string              `bson:"title" json:"title"`
	Description string              `bson:"description" json:"description"`
	StartDate   time.Time           `bson:"start_date" json:"start_date"`
	EndDate     time.Time           `bson:"end_date" json:"end_date"`
	Status      CompetitionStatus   `bson:"status" json:"status"`
	Rules       string              `bson:"rules,omitempty" json:"rules,omitempty"`
	GroupID     primitive.ObjectID  `bson:"group_id,omitempty" json:"group_id,omitempty"` // 限定团组
	CreatedBy   primitive.ObjectID  `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at" json:"updated_at"`
}

// CompetitionEntry 参赛作品
type CompetitionEntry struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CompetitionID primitive.ObjectID `bson:"competition_id" json:"competition_id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	CheckinID     primitive.ObjectID `bson:"checkin_id" json:"checkin_id"`
	Score         float64            `bson:"score" json:"score"`
	JudgeID       primitive.ObjectID `bson:"judge_id,omitempty" json:"judge_id,omitempty"`
	Status        int                `bson:"status" json:"status"` // 0=待评分, 1=已评分
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`

	// 联查字段
	User     *User     `bson:"-" json:"user,omitempty"`
	Checkin  *Checkin  `bson:"-" json:"checkin,omitempty"`
}
