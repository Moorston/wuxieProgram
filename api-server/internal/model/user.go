package model

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OpenID    string             `bson:"openid" json:"-"`                    // 安全：不暴露给前端
	UnionID   string             `bson:"unionid,omitempty" json:"-"`         // 安全：不暴露给前端
	Nickname  string             `bson:"nickname" json:"nickname"`
	Avatar    string             `bson:"avatar" json:"avatar"`
	Gender    int                `bson:"gender" json:"gender"`
	Province  string             `bson:"province,omitempty" json:"province,omitempty"`
	City      string             `bson:"city,omitempty" json:"city,omitempty"`
	GroupID   primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	Score     int                `bson:"score" json:"score"`
	CheckDays int                `bson:"check_days" json:"check_days"`
	Status    int                `bson:"status" json:"-"`                    // 安全：不暴露给前端
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// 用户状态常量
const (
	UserStatusActive  = 0 // 正常
	UserStatusBanned  = 1 // 封禁
)

// IsBanned 判断用户是否被封禁
func (u *User) IsBanned() bool {
	return u.Status == UserStatusBanned
}

// CanCheckin 检查用户是否可以打卡
// 返回 nil 表示可以，返回 error 表示不可以及原因
func (u *User) CanCheckin() error {
	if u.IsBanned() {
		return fmt.Errorf("account has been suspended")
	}
	return nil
}

// DisplayName 返回用户显示名称，优先昵称，兜底为"匿名用户"
func (u *User) DisplayName() string {
	if u.Nickname != "" {
		return u.Nickname
	}
	return "匿名用户"
}
