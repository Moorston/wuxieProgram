package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Challenge 打卡挑战
type Challenge struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Title       string               `bson:"title" json:"title"`
	Description string               `bson:"description" json:"description"`
	Duration    int                  `bson:"duration" json:"duration"` // 持续天数（如 7, 14, 30）
	CreatorID   primitive.ObjectID   `bson:"creator_id" json:"creator_id"`
	GroupID     primitive.ObjectID   `bson:"group_id,omitempty" json:"group_id,omitempty"` // 可选：限定团组
	StartDate   time.Time            `bson:"start_date" json:"start_date"`
	EndDate     time.Time            `bson:"end_date" json:"end_date"`
	ParticipantIDs []primitive.ObjectID `bson:"participant_ids" json:"participant_ids"`
	Status      ChallengeStatus      `bson:"status" json:"status"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`

	// 联查字段
	Creator      *User              `bson:"-" json:"creator,omitempty"`
	Participants []*ChallengeParticipant `bson:"-" json:"participants,omitempty"`
}

// ChallengeStatus 挑战状态
type ChallengeStatus int

const (
	ChallengeStatusActive  ChallengeStatus = 0 // 进行中
	ChallengeStatusEnded   ChallengeStatus = 1 // 已结束
)

// ChallengeParticipant 挑战参与者
type ChallengeParticipant struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ChallengeID primitive.ObjectID `bson:"challenge_id" json:"challenge_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CompletedDays int              `bson:"completed_days" json:"completed_days"` // 已完成天数
	Progress    float64            `bson:"progress" json:"progress"`             // 完成进度 (0-100)
	IsCompleted bool               `bson:"is_completed" json:"is_completed"`
	JoinedAt    time.Time          `bson:"joined_at" json:"joined_at"`

	// 联查字段
	User *User `bson:"-" json:"user,omitempty"`
}

// HasParticipant 检查用户是否已参与
func (c *Challenge) HasParticipant(userID primitive.ObjectID) bool {
	for _, id := range c.ParticipantIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// IsActive 检查挑战是否进行中
func (c *Challenge) IsActive() bool {
	now := time.Now()
	return c.Status == ChallengeStatusActive && now.Before(c.EndDate)
}

// DaysRemaining 剩余天数
func (c *Challenge) DaysRemaining() int {
	remaining := time.Until(c.EndDate).Hours() / 24
	if remaining < 0 {
		return 0
	}
	return int(remaining)
}
