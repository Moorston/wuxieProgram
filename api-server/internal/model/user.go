package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OpenID    string             `bson:"openid" json:"openid"`
	UnionID   string             `bson:"unionid,omitempty" json:"unionid,omitempty"`
	Nickname  string             `bson:"nickname" json:"nickname"`
	Avatar    string             `bson:"avatar" json:"avatar"`
	Gender    int                `bson:"gender" json:"gender"`
	Province  string             `bson:"province,omitempty" json:"province,omitempty"`
	City      string             `bson:"city,omitempty" json:"city,omitempty"`
	GroupID   primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	Score     int                `bson:"score" json:"score"`
	CheckDays int                `bson:"check_days" json:"check_days"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
