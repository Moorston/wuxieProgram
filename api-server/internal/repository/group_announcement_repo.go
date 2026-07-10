package repository

import (
	"context"
	"time"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GroupAnnouncementRepo struct {
	coll *mongo.Collection
}

func NewGroupAnnouncementRepo(db *mongo.Database) *GroupAnnouncementRepo {
	return &GroupAnnouncementRepo{coll: db.Collection("group_announcements")}
}

func (r *GroupAnnouncementRepo) Create(ctx context.Context, a *model.GroupAnnouncement) error {
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now
	result, err := r.coll.InsertOne(ctx, a)
	if err != nil {
		return err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		a.ID = id
	}
	return nil
}

func (r *GroupAnnouncementRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.GroupAnnouncement, error) {
	var a model.GroupAnnouncement
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *GroupAnnouncementRepo) ListByGroup(ctx context.Context, groupID primitive.ObjectID, page, pageSize int) ([]*model.GroupAnnouncement, int64, error) {
	filter := bson.M{"group_id": groupID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "is_pinned", Value: -1}, {Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var announcements []*model.GroupAnnouncement
	if err := cursor.All(ctx, &announcements); err != nil {
		return nil, 0, err
	}
	return announcements, total, nil
}

func (r *GroupAnnouncementRepo) Delete(ctx context.Context, id, authorID primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "author_id": authorID})
	return err
}

func (r *GroupAnnouncementRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "created_at", Value: -1}},
	})
	return err
}

// GroupAnnouncementRepoInterface 团组公告仓库接口
type GroupAnnouncementRepoInterface interface {
	Create(ctx context.Context, a *model.GroupAnnouncement) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.GroupAnnouncement, error)
	ListByGroup(ctx context.Context, groupID primitive.ObjectID, page, pageSize int) ([]*model.GroupAnnouncement, int64, error)
	Delete(ctx context.Context, id, authorID primitive.ObjectID) error
}

var _ GroupAnnouncementRepoInterface = (*GroupAnnouncementRepo)(nil)
