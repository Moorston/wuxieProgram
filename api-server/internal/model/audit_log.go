package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLog 管理员操作日志
type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AdminUser string             `bson:"admin_user" json:"admin_user"`
	Action    string             `bson:"action" json:"action"`       // ban_user, unban_user, delete_checkin, delete_insight, update_config
	TargetID  string             `bson:"target_id" json:"target_id"` // 操作目标ID
	TargetType string            `bson:"target_type" json:"target_type"` // user, checkin, insight, config
	Detail    string             `bson:"detail" json:"detail"`       // 操作详情
	IP        string             `bson:"ip" json:"ip"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
