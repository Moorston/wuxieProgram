package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CheckinStatus int

const (
	CheckinStatusPending    CheckinStatus = 0 // 待转码
	CheckinStatusProcessing CheckinStatus = 1 // 转码中
	CheckinStatusDone       CheckinStatus = 2 // 已完成
	CheckinStatusFailed     CheckinStatus = 3 // 失败
)

type Checkin struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	VideoURL    string             `bson:"video_url" json:"video_url"`
	CoverURL    string             `bson:"cover_url" json:"cover_url"`
	RawURL      string             `bson:"raw_url" json:"raw_url,omitempty"`
	Description string             `bson:"description" json:"description"`
	Duration    float64            `bson:"duration" json:"duration"`
	FileSize    int64              `bson:"file_size" json:"file_size"`
	Score       int                `bson:"score" json:"score"`
	Status      CheckinStatus      `bson:"status" json:"status"`
	LikeCount   int                `bson:"like_count" json:"like_count"`
	CommentCount int              `bson:"comment_count" json:"comment_count"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	// 联查字段，不存DB
	User        *User  `bson:"-" json:"user,omitempty"`
	IsLiked     bool   `bson:"-" json:"is_liked"`
}

type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CheckinID primitive.ObjectID `bson:"checkin_id" json:"checkin_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`

	User *User `bson:"-" json:"user,omitempty"`
}

type Like struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CheckinID primitive.ObjectID `bson:"checkin_id" json:"checkin_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
