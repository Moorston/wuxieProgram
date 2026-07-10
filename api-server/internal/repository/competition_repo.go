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

type CompetitionRepo struct {
	coll *mongo.Collection
}

func NewCompetitionRepo(db *mongo.Database) *CompetitionRepo {
	return &CompetitionRepo{coll: db.Collection("competitions")}
}

func (r *CompetitionRepo) Create(ctx context.Context, c *model.Competition) error {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	result, err := r.coll.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil
	}
	c.ID = id
	return nil
}

func (r *CompetitionRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Competition, error) {
	var c model.Competition
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CompetitionRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *CompetitionRepo) List(ctx context.Context, page, pageSize int, status *model.CompetitionStatus) ([]*model.Competition, int64, error) {
	filter := bson.M{}
	if status != nil {
		filter["status"] = *status
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "start_date", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var competitions []*model.Competition
	if err := cursor.All(ctx, &competitions); err != nil {
		return nil, 0, err
	}
	return competitions, total, nil
}

func (r *CompetitionRepo) ListActive(ctx context.Context) ([]*model.Competition, error) {
	now := time.Now()
	filter := bson.M{
		"status":     model.CompetitionStatusActive,
		"start_date": bson.M{"$lte": now},
		"end_date":   bson.M{"$gte": now},
	}

	cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "start_date", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var competitions []*model.Competition
	if err := cursor.All(ctx, &competitions); err != nil {
		return nil, err
	}
	return competitions, nil
}

func (r *CompetitionRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "start_date", Value: -1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	return err
}

// CompetitionRepoInterface 赛事仓库接口
type CompetitionRepoInterface interface {
	Create(ctx context.Context, c *model.Competition) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Competition, error)
	Update(ctx context.Context, id primitive.ObjectID, update bson.M) error
	List(ctx context.Context, page, pageSize int, status *model.CompetitionStatus) ([]*model.Competition, int64, error)
	ListActive(ctx context.Context) ([]*model.Competition, error)
}

var _ CompetitionRepoInterface = (*CompetitionRepo)(nil)
