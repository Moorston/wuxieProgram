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

type CompetitionEntryRepo struct {
	coll *mongo.Collection
}

func NewCompetitionEntryRepo(db *mongo.Database) *CompetitionEntryRepo {
	return &CompetitionEntryRepo{coll: db.Collection("competition_entries")}
}

func (r *CompetitionEntryRepo) Create(ctx context.Context, entry *model.CompetitionEntry) error {
	entry.CreatedAt = time.Now()
	entry.Status = 0 // 待评分
	result, err := r.coll.InsertOne(ctx, entry)
	if err != nil {
		return err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		entry.ID = id
	}
	return nil
}

func (r *CompetitionEntryRepo) FindByUserAndCompetition(ctx context.Context, userID, competitionID primitive.ObjectID) (*model.CompetitionEntry, error) {
	var entry model.CompetitionEntry
	err := r.coll.FindOne(ctx, bson.M{
		"user_id":        userID,
		"competition_id": competitionID,
	}).Decode(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *CompetitionEntryRepo) Score(ctx context.Context, entryID, judgeID primitive.ObjectID, score float64) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": entryID}, bson.M{
		"$set": bson.M{
			"score":     score,
			"judge_id":  judgeID,
			"status":    1, // 已评分
		},
	})
	return err
}

func (r *CompetitionEntryRepo) ListByCompetition(ctx context.Context, competitionID primitive.ObjectID, page, pageSize int) ([]*model.CompetitionEntry, int64, error) {
	filter := bson.M{"competition_id": competitionID}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: -1}, {Key: "created_at", Value: 1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var entries []*model.CompetitionEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

// GetRanking 获取赛事排行榜
func (r *CompetitionEntryRepo) GetRanking(ctx context.Context, competitionID primitive.ObjectID, limit int) ([]*model.CompetitionEntry, error) {
	filter := bson.M{
		"competition_id": competitionID,
		"status":         1, // 已评分
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "score", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*model.CompetitionEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (r *CompetitionEntryRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "competition_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "competition_id", Value: 1}, {Key: "score", Value: -1}}},
	})
	return err
}

// CompetitionEntryRepoInterface 参赛作品仓库接口
type CompetitionEntryRepoInterface interface {
	Create(ctx context.Context, entry *model.CompetitionEntry) error
	FindByUserAndCompetition(ctx context.Context, userID, competitionID primitive.ObjectID) (*model.CompetitionEntry, error)
	Score(ctx context.Context, entryID, judgeID primitive.ObjectID, score float64) error
	ListByCompetition(ctx context.Context, competitionID primitive.ObjectID, page, pageSize int) ([]*model.CompetitionEntry, int64, error)
	GetRanking(ctx context.Context, competitionID primitive.ObjectID, limit int) ([]*model.CompetitionEntry, error)
}

var _ CompetitionEntryRepoInterface = (*CompetitionEntryRepo)(nil)
