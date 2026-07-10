package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name        string               `bson:"name" json:"name"`
	Description string               `bson:"description" json:"description"`
	LeaderID    primitive.ObjectID   `bson:"leader_id" json:"leader_id"`
	MemberIDs   []primitive.ObjectID `bson:"member_ids" json:"member_ids"`
	InviteCode  string               `bson:"invite_code,omitempty" json:"invite_code,omitempty"`
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`

	Leader  *User  `bson:"-" json:"leader,omitempty"`
	Members []*User `bson:"-" json:"members,omitempty"`
}

// HasMember 检查用户是否为团组成员
func (g *Group) HasMember(userID primitive.ObjectID) bool {
	for _, id := range g.MemberIDs {
		if id == userID {
			return true
		}
	}
	return false
}

// IsLeader 检查用户是否为团组组长
func (g *Group) IsLeader(userID primitive.ObjectID) bool {
	return g.LeaderID == userID
}
