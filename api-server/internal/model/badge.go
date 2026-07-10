package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// BadgeType 徽章类型
type BadgeType string

const (
	BadgeTypeStreak    BadgeType = "streak"     // 连续打卡
	BadgeTypeTotal     BadgeType = "total"      // 累计打卡
	BadgeTypeScore     BadgeType = "score"      // 积分达标
	BadgeTypeCompetition BadgeType = "competition" // 赛事获奖
	BadgeTypeSocial    BadgeType = "social"     // 社交互动
)

// Badge 徽章定义
type Badge struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Icon        string             `bson:"icon" json:"icon"`           // emoji 或图标名
	Type        BadgeType          `bson:"type" json:"type"`
	Condition   int                `bson:"condition" json:"condition"` // 达成条件值（如 7 天、100 次）
	Level       int                `bson:"level" json:"level"`         // 等级：1=铜, 2=银, 3=金
	SortOrder   int                `bson:"sort_order" json:"sort_order"`
}

// UserBadge 用户获得的徽章
type UserBadge struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	BadgeID   primitive.ObjectID `bson:"badge_id" json:"badge_id"`
	EarnedAt  time.Time          `bson:"earned_at" json:"earned_at"`

	// 联查字段
	Badge *Badge `bson:"-" json:"badge,omitempty"`
}

// 预定义徽章
var DefaultBadges = []Badge{
	// 连续打卡徽章
	{Name: "初心者", Description: "连续打卡 3 天", Icon: "🌱", Type: BadgeTypeStreak, Condition: 3, Level: 1, SortOrder: 1},
	{Name: "坚持者", Description: "连续打卡 7 天", Icon: "🔥", Type: BadgeTypeStreak, Condition: 7, Level: 1, SortOrder: 2},
	{Name: "毅力者", Description: "连续打卡 30 天", Icon: "💪", Type: BadgeTypeStreak, Condition: 30, Level: 2, SortOrder: 3},
	{Name: "钢铁意志", Description: "连续打卡 100 天", Icon: "🏆", Type: BadgeTypeStreak, Condition: 100, Level: 3, SortOrder: 4},
	// 累计打卡徽章
	{Name: "初出茅庐", Description: "累计打卡 10 次", Icon: "🥋", Type: BadgeTypeTotal, Condition: 10, Level: 1, SortOrder: 5},
	{Name: "武者之路", Description: "累计打卡 50 次", Icon: "⚔️", Type: BadgeTypeTotal, Condition: 50, Level: 2, SortOrder: 6},
	{Name: "武林高手", Description: "累计打卡 100 次", Icon: "🐉", Type: BadgeTypeTotal, Condition: 100, Level: 2, SortOrder: 7},
	{Name: "一代宗师", Description: "累计打卡 365 次", Icon: "👑", Type: BadgeTypeTotal, Condition: 365, Level: 3, SortOrder: 8},
	// 积分徽章
	{Name: "积分新手", Description: "累计积分达到 100", Icon: "⭐", Type: BadgeTypeScore, Condition: 100, Level: 1, SortOrder: 9},
	{Name: "积分达人", Description: "累计积分达到 500", Icon: "🌟", Type: BadgeTypeScore, Condition: 500, Level: 2, SortOrder: 10},
	{Name: "积分大师", Description: "累计积分达到 1000", Icon: "💫", Type: BadgeTypeScore, Condition: 1000, Level: 3, SortOrder: 11},
}
