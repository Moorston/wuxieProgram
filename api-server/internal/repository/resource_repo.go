package repository

import (
	"context"
	"regexp"
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

type ResourceRepo struct {
	coll *mongo.Collection
}

func NewResourceRepo(db *mongo.Database) *ResourceRepo {
	return &ResourceRepo{coll: db.Collection("resources")}
}

func (r *ResourceRepo) Create(ctx context.Context, res *model.Resource) error {
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()
	result, err := r.coll.InsertOne(ctx, res)
	if err != nil {
		return err
	}
	res.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *ResourceRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Resource, error) {
	var res model.Resource
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ResourceRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	update["updated_at"] = time.Now()
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *ResourceRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *ResourceRepo) List(ctx context.Context, userID primitive.ObjectID, resType, category, difficulty, tag, keyword, shareScope, sortBy string, groupID *primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
	filter := bson.M{}

	if shareScope == "visible" {
		or := []bson.M{
			{"user_id": userID},
			{"share_scope": model.ShareScopePublic},
		}
		if groupID != nil {
			or = append(or, bson.M{"share_scope": model.ShareScopeGroup, "group_id": *groupID})
		}
		filter["$or"] = or
	} else {
		filter["user_id"] = userID
	}

	if resType != "" {
		filter["type"] = resType
	}
	if category != "" {
		filter["category"] = category
	}
	if difficulty != "" {
		filter["difficulty"] = difficulty
	}
	if tag != "" {
		filter["tags"] = tag
	}
	if keyword != "" {
		safeKeyword := sanitizeRegex(keyword)
		keywordFilter := []bson.M{
			{"title": bson.M{"$regex": safeKeyword, "$options": "i"}},
			{"description": bson.M{"$regex": safeKeyword, "$options": "i"}},
			{"tags": bson.M{"$regex": safeKeyword, "$options": "i"}},
		}
		if existingOr, ok := filter["$or"].([]bson.M); ok {
			var mergedOr []bson.M
			for _, visFilter := range existingOr {
				for _, kwFilter := range keywordFilter {
					mergedOr = append(mergedOr, bson.M{"$and": []bson.M{visFilter, kwFilter}})
				}
			}
			delete(filter, "$or")
			filter["$and"] = []bson.M{{"$or": existingOr}, {"$or": keywordFilter}}
		} else {
			filter["$or"] = keywordFilter
		}
	}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	sort := bson.D{{Key: "created_at", Value: -1}}
	if sortBy == "name" {
		sort = bson.D{{Key: "title", Value: 1}}
	} else if sortBy == "size" {
		sort = bson.D{{Key: "file_size", Value: -1}}
	}

	opts := options.Find().
		SetSort(sort).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resources []*model.Resource
	if err := cursor.All(ctx, &resources); err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func (r *ResourceRepo) ToggleFavorite(ctx context.Context, id primitive.ObjectID) (bool, error) {
	var res model.Resource
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&res)
	if err != nil {
		return false, err
	}

	newFav := !res.IsFavorite
	_, err = r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"is_favorite": newFav, "updated_at": time.Now()},
	})
	return newFav, err
}

func (r *ResourceRepo) ListFavorites(ctx context.Context, userID primitive.ObjectID, page, pageSize int) ([]*model.Resource, int64, error) {
	filter := bson.M{"user_id": userID, "is_favorite": true}

	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	opts := options.Find().
		SetSort(bson.D{{Key: "updated_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resources []*model.Resource
	if err := cursor.All(ctx, &resources); err != nil {
		return nil, 0, err
	}

	return resources, total, nil
}

func (r *ResourceRepo) GetUserStats(ctx context.Context, userID primitive.ObjectID) (*model.ResourceStats, error) {
	filter := bson.M{"user_id": userID}

	pipeline := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id":        "$type",
			"total_size": bson.M{"$sum": "$file_size"},
			"count":      bson.M{"$sum": 1},
		}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := &model.ResourceStats{
		TypeStats:  make(map[string]int64),
		TypeCounts: make(map[string]int),
		Quota:      5 * 1024 * 1024 * 1024, // 5GB default
	}

	for cursor.Next(ctx) {
		var result struct {
			ID        string `bson:"_id"`
			TotalSize int64  `bson:"total_size"`
			Count     int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats.TypeStats[result.ID] = result.TotalSize
		stats.TypeCounts[result.ID] = result.Count
		stats.TotalSize += result.TotalSize
		stats.TotalCount += result.Count
	}

	if stats.Quota > 0 {
		stats.UsagePercent = float64(stats.TotalSize) / float64(stats.Quota) * 100
	}

	return stats, nil
}

func (r *ResourceRepo) IncrViewCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"view_count": 1},
	})
	return err
}

func (r *ResourceRepo) IncrDownloadCount(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$inc": bson.M{"download_count": 1},
	})
	return err
}

func (r *ResourceRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "type", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_favorite", Value: 1}}},
		{Keys: bson.D{{Key: "share_scope", Value: 1}, {Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "tags", Value: 1}}},
	})
	return err
}
