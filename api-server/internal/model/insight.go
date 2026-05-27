package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MoodType string

const (
	MoodBreakthrough MoodType = "breakthrough"
	MoodGood         MoodType = "good"
	MoodNormal       MoodType = "normal"
	MoodConfused     MoodType = "confused"
	MoodLow          MoodType = "low"
)

type Visibility string

const (
	VisibilityPrivate Visibility = "private"
	VisibilityPublic  Visibility = "public"
)

type Insight struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID   `bson:"user_id" json:"user_id"`
	Content   string               `bson:"content" json:"content"`
	Images    []string             `bson:"images" json:"images"`
	Mood      MoodType             `bson:"mood" json:"mood"`
	Tags      []string             `bson:"tags" json:"tags"`
	CheckinID primitive.ObjectID   `bson:"checkin_id,omitempty" json:"checkin_id,omitempty"`
	PlanID    primitive.ObjectID   `bson:"plan_id,omitempty" json:"plan_id,omitempty"`
	PlanDay   int                  `bson:"plan_day,omitempty" json:"plan_day,omitempty"`
	Visibility Visibility           `bson:"visibility" json:"visibility"`
	LikeCount int                  `bson:"like_count" json:"like_count"`
	CreatedAt time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time            `bson:"updated_at" json:"updated_at"`
	User      *User                `bson:"-" json:"user,omitempty"`
}

type InsightTag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Tag       string             `bson:"tag" json:"tag"`
	Count     int                `bson:"count" json:"count"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type InsightLike struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InsightID primitive.ObjectID `bson:"insight_id" json:"insight_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
