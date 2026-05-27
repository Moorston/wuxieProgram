package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlanStatus int

const (
	PlanStatusDraft     PlanStatus = 0
	PlanStatusActive    PlanStatus = 1
	PlanStatusCompleted PlanStatus = 2
	PlanStatusTerminated PlanStatus = 3
)

type TaskStatus int

const (
	TaskStatusPending  TaskStatus = 0
	TaskStatusDone     TaskStatus = 1
	TaskStatusSkipped  TaskStatus = 2
)

type TaskType string

const (
	TaskTypeBasic  TaskType = "basic"
	TaskTypeTaolu  TaskType = "taolu"
	TaskTypeSanda  TaskType = "sanda"
	TaskTypeQigong TaskType = "qigong"
)

type TrainingTask struct {
	Title     string     `bson:"title" json:"title"`
	Type      TaskType   `bson:"type" json:"type"`
	Duration  int        `bson:"duration" json:"duration"`
	Reps      string     `bson:"reps" json:"reps"`
	Note      string     `bson:"note" json:"note"`
	CheckinID primitive.ObjectID `bson:"checkin_id,omitempty" json:"checkin_id,omitempty"`
	Status    TaskStatus `bson:"status" json:"status"`
}

type TrainingDay struct {
	Day   int            `bson:"day" json:"day"`
	Date  time.Time      `bson:"date" json:"date"`
	Tasks []TrainingTask `bson:"tasks" json:"tasks"`
}

type PlanStats struct {
	TotalTasks    int     `bson:"total_tasks" json:"total_tasks"`
	Completed     int     `bson:"completed" json:"completed"`
	CompletionRate float64 `bson:"completion_rate" json:"completion_rate"`
}

type TrainingPlan struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	GroupID     primitive.ObjectID `bson:"group_id,omitempty" json:"group_id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	StartDate   time.Time          `bson:"start_date" json:"start_date"`
	EndDate     time.Time          `bson:"end_date" json:"end_date"`
	Status      PlanStatus         `bson:"status" json:"status"`
	Days        []TrainingDay      `bson:"days" json:"days"`
	Stats       PlanStats          `bson:"stats" json:"stats"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type TrainingTemplate struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Category     string             `bson:"category" json:"category"`
	Style        string             `bson:"style" json:"style"`
	DurationDays int                `bson:"duration_days" json:"duration_days"`
	Description  string             `bson:"description" json:"description"`
	Days         []TrainingDay      `bson:"days" json:"days"`
	Author       string             `bson:"author" json:"author"`
	UsageCount   int                `bson:"usage_count" json:"usage_count"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}
