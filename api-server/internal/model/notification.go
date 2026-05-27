package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationType string

const (
	NotifTypeLike          NotificationType = "like"
	NotifTypeComment       NotificationType = "comment"
	NotifTypeCommentReply  NotificationType = "comment_reply"
	NotifTypePlanRemind    NotificationType = "plan_remind"
	NotifTypePlanComplete  NotificationType = "plan_complete"
	NotifTypeGroupNotice   NotificationType = "group_notice"
	NotifTypeSystem        NotificationType = "system"
)

type Notification struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type       NotificationType   `bson:"type" json:"type"`
	Title      string             `bson:"title" json:"title"`
	Content    string             `bson:"content" json:"content"`
	SenderID   primitive.ObjectID `bson:"sender_id,omitempty" json:"sender_id,omitempty"`
	TargetType string             `bson:"target_type,omitempty" json:"target_type,omitempty"`
	TargetID   primitive.ObjectID `bson:"target_id,omitempty" json:"target_id,omitempty"`
	IsRead     bool               `bson:"is_read" json:"is_read"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	Sender     *User              `bson:"-" json:"sender,omitempty"`
}

type NotificationSettings struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	LikeNotify     bool               `bson:"like_notify" json:"like_notify"`
	CommentNotify  bool               `bson:"comment_notify" json:"comment_notify"`
	PlanRemind     bool               `bson:"plan_remind" json:"plan_remind"`
	PlanRemindTime string             `bson:"plan_remind_time" json:"plan_remind_time"`
	GroupNotify    bool               `bson:"group_notify" json:"group_notify"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
