package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GroupAnnouncement 团组公告
type GroupAnnouncement struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   primitive.ObjectID `bson:"group_id" json:"group_id"`
	AuthorID  primitive.ObjectID `bson:"author_id" json:"author_id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	IsPinned  bool               `bson:"is_pinned" json:"is_pinned"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`

	// 联查字段
	Author *User `bson:"-" json:"author,omitempty"`
}
