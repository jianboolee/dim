package repository

import (
	"context"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"d-im/internal/models"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{
		collection: db.Collection(models.CollectionUser),
	}
}

func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	now := time.Now()
	filter := bson.M{"id": user.ID}

	setFields := bson.M{
		"updated_at": now,
	}
	if user.Nickname != "" {
		setFields["nickname"] = user.Nickname
	}
	if user.Avatar != "" {
		setFields["avatar"] = user.Avatar
	}

	update := bson.M{
		"$set": setFields,
		"$setOnInsert": bson.M{
			"id":         user.ID,
			"bio":        user.Bio,
			"created_at": now,
		},
	}

	// 首次插入时若未传 nickname/avatar，写入空字符串占位，避免字段缺失
	if user.Nickname == "" {
		update["$setOnInsert"].(bson.M)["nickname"] = ""
	}
	if user.Avatar == "" {
		update["$setOnInsert"].(bson.M)["avatar"] = ""
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

func (r *UserRepository) UpsertMany(ctx context.Context, users []*models.User) error {
	for _, user := range users {
		if err := r.Upsert(ctx, user); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByIDs(ctx context.Context, ids []string) (map[string]*models.User, error) {
	if len(ids) == 0 {
		return map[string]*models.User{}, nil
	}

	cursor, err := r.collection.Find(ctx, bson.M{"id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := map[string]*models.User{}
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users[user.ID] = &user
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Search(ctx context.Context, keyword string, limit int64) ([]*models.User, error) {
	if keyword == "" {
		return []*models.User{}, nil
	}
	if limit <= 0 {
		limit = 50
	}

	pattern := regexp.QuoteMeta(keyword)
	filter := bson.M{
		"$or": []bson.M{
			{"id": bson.M{"$regex": pattern, "$options": "i"}},
			{"nickname": bson.M{"$regex": pattern, "$options": "i"}},
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	users := make([]*models.User, 0)
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
