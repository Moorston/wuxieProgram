package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResourceType string

const (
	ResourceTypeVideo    ResourceType = "video"
	ResourceTypeImage    ResourceType = "image"
	ResourceTypeDocument ResourceType = "document"
)

type ShareScope string

const (
	ShareScopePrivate ShareScope = "private"
	ShareScopeGroup   ShareScope = "group"
	ShareScopePublic  ShareScope = "public"
)

type Resource struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Title         string             `bson:"title" json:"title"`
	Description   string             `bson:"description" json:"description"`
	Type          ResourceType       `bson:"type" json:"type"`
	Category      string             `bson:"category" json:"category"`
	Tags          []string           `bson:"tags" json:"tags"`
	Difficulty    string             `bson:"difficulty" json:"difficulty"`
	FileURL       string             `bson:"file_url" json:"file_url"`
	FileSize      int64              `bson:"file_size" json:"file_size"`
	CoverURL      string             `bson:"cover_url" json:"cover_url"`
	Duration      float64            `bson:"duration" json:"duration"`
	ShareScope    ShareScope         `bson:"share_scope" json:"share_scope"`
	GroupID       primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	IsFavorite    bool               `bson:"is_favorite" json:"is_favorite"`
	ViewCount     int                `bson:"view_count" json:"view_count"`
	DownloadCount int                `bson:"download_count" json:"download_count"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	User          *User              `bson:"-" json:"user,omitempty"`
}

type ResourceTag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Tag       string             `bson:"tag" json:"tag"`
	Count     int                `bson:"count" json:"count"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type ResourceStats struct {
	TotalSize    int64            `json:"total_size"`
	TotalCount   int              `json:"total_count"`
	TypeStats    map[string]int64 `json:"type_stats"`
	TypeCounts   map[string]int   `json:"type_counts"`
	Quota        int64            `json:"quota"`
	UsagePercent float64          `json:"usage_percent"`
}
