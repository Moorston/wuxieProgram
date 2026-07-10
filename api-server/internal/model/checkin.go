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
	Visibility  Visibility         `bson:"visibility" json:"visibility"` // 可见性
	LikeCount   int                `bson:"like_count" json:"like_count"`
	CommentCount int              `bson:"comment_count" json:"comment_count"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	// 联查字段，不存DB
	User        *User  `bson:"-" json:"user,omitempty"`
	IsLiked     bool   `bson:"-" json:"is_liked"`
}

// IsPublic 检查是否公开可见
func (c *Checkin) IsPublic() bool {
	return c.Visibility == VisibilityPublic
}

// IsVisibleTo 检查对指定用户是否可见
// groupMemberIDs: 查看者所属团组的成员 ID 列表（可选，用于团组可见性判断）
func (c *Checkin) IsVisibleTo(viewerID primitive.ObjectID, groupMemberIDs ...[]primitive.ObjectID) bool {
	// 作者始终可见
	if c.UserID == viewerID {
		return true
	}
	// 公开可见
	if c.Visibility == VisibilityPublic {
		return true
	}
	// 私有：仅作者可见
	if c.Visibility == VisibilityPrivate {
		return false
	}
	// 团组可见：检查查看者是否与作者在同一团组
	if c.Visibility == VisibilityGroup && len(groupMemberIDs) > 0 {
		for _, ids := range groupMemberIDs {
			for _, id := range ids {
				if id == c.UserID {
					return true
				}
			}
		}
	}
	return false
}

// IsProcessed 检查打卡视频是否已处理完成
func (c *Checkin) IsProcessed() bool {
	return c.Status == CheckinStatusDone
}

// IsFailed 检查打卡是否处理失败
func (c *Checkin) IsFailed() bool {
	return c.Status == CheckinStatusFailed
}

// IsPending 检查打卡是否待处理
func (c *Checkin) IsPending() bool {
	return c.Status == CheckinStatusPending || c.Status == CheckinStatusProcessing
}

// BelongsTo 检查打卡是否属于指定用户
func (c *Checkin) BelongsTo(userID primitive.ObjectID) bool {
	return c.UserID == userID
}

// CanDelete 检查打卡是否可以被删除（只有作者可以删除）
func (c *Checkin) CanDelete(userID primitive.ObjectID) bool {
	return c.BelongsTo(userID)
}

type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CheckinID primitive.ObjectID `bson:"checkin_id" json:"checkin_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	ParentID  primitive.ObjectID `bson:"parent_id,omitempty" json:"parent_id,omitempty"` // 回复的评论ID
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`

	User     *User     `bson:"-" json:"user,omitempty"`
	Replies  []*Comment `bson:"-" json:"replies,omitempty"` // 子回复（联查填充）
}

type Like struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CheckinID primitive.ObjectID `bson:"checkin_id" json:"checkin_id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
