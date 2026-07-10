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

type ChallengeRepo struct {
	coll *mongo.Collection
}

func NewChallengeRepo(db *mongo.Database) *ChallengeRepo {
	return &ChallengeRepo{coll: db.Collection("challenges")}
}

func (r *ChallengeRepo) Create(ctx context.Context, c *model.Challenge) error {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	result, err := r.coll.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	id, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		c.ID = id
	}
	return nil
}

func (r *ChallengeRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Challenge, error) {
	var c model.Challenge
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ChallengeRepo) ListActive(ctx context.Context, page, pageSize int) ([]*model.Challenge, int64, error) {
	now := time.Now()
	filter := bson.M{
		"status":    model.ChallengeStatusActive,
		"end_date":  bson.M{"$gte": now},
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

	var challenges []*model.Challenge
	if err := cursor.All(ctx, &challenges); err != nil {
		return nil, 0, err
	}
	return challenges, total, nil
}

func (r *ChallengeRepo) AddParticipant(ctx context.Context, challengeID, userID primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": challengeID},
		bson.M{
			"$addToSet": bson.M{"participant_ids": userID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *ChallengeRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "status", Value: 1}, {Key: "start_date", Value: -1}}},
		{Keys: bson.D{{Key: "creator_id", Value: 1}}},
	})
	return err
}

// ChallengeParticipantRepo 挑战参与者仓库
type ChallengeParticipantRepo struct {
	coll *mongo.Collection
}

func NewChallengeParticipantRepo(db *mongo.Database) *ChallengeParticipantRepo {
	return &ChallengeParticipantRepo{coll: db.Collection("challenge_participants")}
}

func (r *ChallengeParticipantRepo) Upsert(ctx context.Context, p *model.ChallengeParticipant) error {
	p.JoinedAt = time.Now()
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"challenge_id": p.ChallengeID, "user_id": p.UserID},
		bson.M{"$setOnInsert": p},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *ChallengeParticipantRepo) FindByUserAndChallenge(ctx context.Context, userID, challengeID primitive.ObjectID) (*model.ChallengeParticipant, error) {
	var p model.ChallengeParticipant
	err := r.coll.FindOne(ctx, bson.M{
		"user_id":      userID,
		"challenge_id": challengeID,
	}).Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ChallengeParticipantRepo) IncrementCompletedDays(ctx context.Context, userID, challengeID primitive.ObjectID, duration int) error {
	res := r.coll.FindOneAndUpdate(ctx,
		bson.M{"user_id": userID, "challenge_id": challengeID},
		bson.M{
			"$inc": bson.M{"completed_days": 1},
		},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)
	var p model.ChallengeParticipant
	if err := res.Decode(&p); err != nil {
		return err
	}
	// 更新进度
	progress := float64(p.CompletedDays) / float64(duration) * 100
	isCompleted := p.CompletedDays >= duration
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"_id": p.ID},
		bson.M{
			"$set": bson.M{
				"progress":      progress,
				"is_completed":  isCompleted,
			},
		},
	)
	return err
}

func (r *ChallengeParticipantRepo) ListByChallenge(ctx context.Context, challengeID primitive.ObjectID) ([]*model.ChallengeParticipant, error) {
	opts := options.Find().SetSort(bson.D{{Key: "progress", Value: -1}})
	cursor, err := r.coll.Find(ctx, bson.M{"challenge_id": challengeID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var participants []*model.ChallengeParticipant
	if err := cursor.All(ctx, &participants); err != nil {
		return nil, err
	}
	return participants, nil
}

func (r *ChallengeParticipantRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "challenge_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
	})
	return err
}

// ChallengeRepoInterface 挑战仓库接口
type ChallengeRepoInterface interface {
	Create(ctx context.Context, c *model.Challenge) error
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Challenge, error)
	ListActive(ctx context.Context, page, pageSize int) ([]*model.Challenge, int64, error)
	AddParticipant(ctx context.Context, challengeID, userID primitive.ObjectID) error
}

// ChallengeParticipantRepoInterface 参与者仓库接口
type ChallengeParticipantRepoInterface interface {
	Upsert(ctx context.Context, p *model.ChallengeParticipant) error
	FindByUserAndChallenge(ctx context.Context, userID, challengeID primitive.ObjectID) (*model.ChallengeParticipant, error)
	IncrementCompletedDays(ctx context.Context, userID, challengeID primitive.ObjectID, duration int) error
	ListByChallenge(ctx context.Context, challengeID primitive.ObjectID) ([]*model.ChallengeParticipant, error)
}

var _ ChallengeRepoInterface = (*ChallengeRepo)(nil)
var _ ChallengeParticipantRepoInterface = (*ChallengeParticipantRepo)(nil)
