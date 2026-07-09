package repository

import (
	"context"
	"regexp"
	"strings"
	"time"

	"wuxie-api/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var regexSpecialChars = regexp.MustCompile(`[[\]{}()*+?.\\^$|]`)

func sanitizeRegex(s string) string {
	return regexSpecialChars.ReplaceAllString(s, `\$&`)
}

// validateSearchKeyword 验证和清理搜索关键词，防止ReDoS攻击
func validateSearchKeyword(keyword string) string {
	// 限制关键词长度
	if len(keyword) > 50 {
		keyword = keyword[:50]
	}

	// 去除可能导致回溯攻击的嵌套括号
	keyword = strings.ReplaceAll(keyword, "((", "(")
	keyword = strings.ReplaceAll(keyword, "))", ")")

	// 去除重复的特殊字符
	for strings.Contains(keyword, "**") {
		keyword = strings.ReplaceAll(keyword, "**", "*")
	}
	for strings.Contains(keyword, "++") {
		keyword = strings.ReplaceAll(keyword, "++", "+")
	}

	return keyword
}

type CheckinRepo struct {
	coll *mongo.Collection
}

func NewCheckinRepo(db *mongo.Database) *CheckinRepo {
	return &CheckinRepo{coll: db.Collection("checkins")}
}

func (r *CheckinRepo) Create(ctx context.Context, c *model.Checkin) error {
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	c.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *CheckinRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Checkin, error) {
	var checkin model.Checkin
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&checkin)
	if err != nil {
		return nil, err
	}
	return &checkin, nil
}

func (r *CheckinRepo) UpdateStatus(ctx context.Context, id primitive.ObjectID, status model.CheckinStatus, videoURL, coverURL string, duration float64) error {
	update := bson.M{
		"status":    status,
		"updated_at": time.Now(),
	}
	if videoURL != "" {
		update["video_url"] = videoURL
	}
	if coverURL != "" {
		update["cover_url"] = coverURL
	}
	if duration > 0 {
		update["duration"] = duration
	}
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *CheckinRepo) List(ctx context.Context, userID primitive.ObjectID, groupUserIDs []primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	filter := bson.M{"status": model.CheckinStatusDone}

	if len(groupUserIDs) > 0 {
		filter["user_id"] = bson.M{"$in": groupUserIDs}
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var checkins []*model.Checkin
	if err := cursor.All(ctx, &checkins); err != nil {
		return nil, 0, err
	}

	return checkins, total, nil
}

func (r *CheckinRepo) ListByUser(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Checkin, int64, error) {
	filter := bson.M{"user_id": userID, "status": model.CheckinStatusDone}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var checkins []*model.Checkin
	if err := cursor.All(ctx, &checkins); err != nil {
		return nil, 0, err
	}

	return checkins, total, nil
}

func (r *CheckinRepo) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id, "user_id": userID})
	return err
}

func (r *CheckinRepo) IncrLikeCount(ctx context.Context, id primitive.ObjectID, delta int) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"like_count": delta},
	})
	return err
}

// IncrLikeCountWithSession 在事务中增加点赞计数
func (r *CheckinRepo) IncrLikeCountWithSession(sessCtx mongo.SessionContext, id primitive.ObjectID, delta int) error {
	_, err := r.coll.UpdateOne(sessCtx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"like_count": delta},
	})
	return err
}

func (r *CheckinRepo) IncrCommentCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"comment_count": 1},
	})
	return err
}

// IncrCommentCountWithSession 在事务中增加评论计数
func (r *CheckinRepo) IncrCommentCountWithSession(sessCtx mongo.SessionContext, id primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(sessCtx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"comment_count": 1},
	})
	return err
}

func (r *CheckinRepo) Search(ctx context.Context, keyword string, page, pageSize int) ([]*model.Checkin, int64, error) {
	// 使用验证函数清理和限制关键词
	keyword = validateSearchKeyword(keyword)

	safeKeyword := sanitizeRegex(keyword)
	filter := bson.M{
		"status": model.CheckinStatusDone,
		"description": bson.M{
			"$regex":   safeKeyword,
			"$options": "i",
		},
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var checkins []*model.Checkin
	if err := cursor.All(ctx, &checkins); err != nil {
		return nil, 0, err
	}

	return checkins, total, nil
}

func (r *CheckinRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
	})
	return err
}

func (r *CheckinRepo) Aggregate(ctx context.Context, pipeline []bson.M) (mongo.Cursor, error) {
	return r.coll.Aggregate(ctx, pipeline)
}
