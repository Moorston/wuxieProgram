package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Follow 关注关系
type Follow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FollowerID  primitive.ObjectID `bson:"follower_id" json:"follower_id"`   // 关注者
	FollowingID primitive.ObjectID `bson:"following_id" json:"following_id"` // 被关注者
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}