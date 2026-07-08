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
	CreatedAt   time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updated_at" json:"updated_at"`

	Leader  *User  `bson:"-" json:"leader,omitempty"`
	Members []*User `bson:"-" json:"members,omitempty"`
}
