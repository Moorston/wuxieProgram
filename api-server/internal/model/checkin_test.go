package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCheckin_IsProcessed(t *testing.T) {
	tests := []struct {
		status CheckinStatus
		want   bool
	}{
		{CheckinStatusPending, false},
		{CheckinStatusProcessing, false},
		{CheckinStatusDone, true},
		{CheckinStatusFailed, false},
	}
	for _, tt := range tests {
		c := &Checkin{Status: tt.status}
		assert.Equal(t, tt.want, c.IsProcessed(), "status=%d", tt.status)
	}
}

func TestCheckin_IsFailed(t *testing.T) {
	tests := []struct {
		status CheckinStatus
		want   bool
	}{
		{CheckinStatusPending, false},
		{CheckinStatusProcessing, false},
		{CheckinStatusDone, false},
		{CheckinStatusFailed, true},
	}
	for _, tt := range tests {
		c := &Checkin{Status: tt.status}
		assert.Equal(t, tt.want, c.IsFailed(), "status=%d", tt.status)
	}
}

func TestCheckin_IsPending(t *testing.T) {
	tests := []struct {
		status CheckinStatus
		want   bool
	}{
		{CheckinStatusPending, true},
		{CheckinStatusProcessing, true},
		{CheckinStatusDone, false},
		{CheckinStatusFailed, false},
	}
	for _, tt := range tests {
		c := &Checkin{Status: tt.status}
		assert.Equal(t, tt.want, c.IsPending(), "status=%d", tt.status)
	}
}

func TestCheckin_BelongsTo_Match(t *testing.T) {
	userID := primitive.NewObjectID()
	c := &Checkin{UserID: userID}
	assert.True(t, c.BelongsTo(userID))
}

func TestCheckin_BelongsTo_NoMatch(t *testing.T) {
	c := &Checkin{UserID: primitive.NewObjectID()}
	assert.False(t, c.BelongsTo(primitive.NewObjectID()))
}

func TestCheckin_CanDelete_Author(t *testing.T) {
	userID := primitive.NewObjectID()
	c := &Checkin{UserID: userID}
	assert.True(t, c.CanDelete(userID))
}

func TestCheckin_CanDelete_NotAuthor(t *testing.T) {
	c := &Checkin{UserID: primitive.NewObjectID()}
	assert.False(t, c.CanDelete(primitive.NewObjectID()))
}
